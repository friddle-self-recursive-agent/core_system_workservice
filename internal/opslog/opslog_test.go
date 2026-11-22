// opslog_test.go —— ops DB SQLite 实现验收：建表落行、持久化、库内对账不变量。
package opslog

import (
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func TestOpenMemoryAndAppend(t *testing.T) {
	s, err := OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.AppendApply(ApplyLog{
		Ts: time.Now(), FieldManager: "fm", Obj: "Agent/global/a", ResourceVersion: 1,
		Result: "ok", ReasonRef: "plan:p",
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.AppendEvent(EventLog{
		Ts: time.Now(), EventID: "evt-1", Type: "DocApplied", Severity: "info", RoutedTo: "1 subscriber(s)",
	}); err != nil {
		t.Fatal(err)
	}
	if ok, err := s.CountApply("ok"); err != nil || ok != 1 {
		t.Fatalf("ok 行 = %d err=%v", ok, err)
	}
	if n, err := s.CountEvent("DocApplied"); err != nil || n != 1 {
		t.Fatalf("DocApplied 行 = %d err=%v", n, err)
	}
	if diff, err := s.Reconcile(); err != nil || diff != 0 {
		t.Fatalf("对账 diff=%d err=%v, 要 0", diff, err)
	}
	// 破坏不变量：多一行 ok（无对应事件）→ diff=1。
	if err := s.AppendApply(ApplyLog{Obj: "Agent/global/b", Result: "ok"}); err != nil {
		t.Fatal(err)
	}
	if diff, err := s.Reconcile(); err != nil || diff != 1 {
		t.Fatalf("对账 diff=%d err=%v, 要 1", diff, err)
	}
	// 必填校验。
	if err := s.AppendApply(ApplyLog{Obj: "Agent/global/c"}); err == nil {
		t.Fatal("缺 result 应被拒")
	}
	if err := s.AppendEvent(EventLog{EventID: "evt-2"}); err == nil {
		t.Fatal("缺 type 应被拒")
	}
}

func TestFilePersistence(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "nested", "ops.db") // 父目录自动创建
	s, err := OpenSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AppendApply(ApplyLog{Obj: "Agent/global/a", Result: "ok", ResourceVersion: 1}); err != nil {
		t.Fatal(err)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	// 重开后账本仍在（WAL 检查点归并）。
	s2, err := OpenSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()
	if ok, err := s2.CountApply("ok"); err != nil || ok != 1 {
		t.Fatalf("重开 ok 行 = %d err=%v, 要持久", ok, err)
	}
}

// TestApplyLogSubjectColumn 主体列（v0.9 请求级主体）：落行/查询往返；
// denied 行主体归属断言（验收探针「查询代替日志」的机械取证面）。
func TestApplyLogSubjectColumn(t *testing.T) {
	s, err := OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.AppendApply(ApplyLog{
		Obj: "Agent/global/a", Result: "denied", FieldManager: "agent:x", Subject: "agent:none",
		PolicyDecisions: "permission: denied write on global (subject=agent:none judged by role)",
	}); err != nil {
		t.Fatal(err)
	}
	rows, err := s.QueryApply("denied")
	if err != nil || len(rows) != 1 {
		t.Fatalf("QueryApply(denied) = %d 行 err=%v", len(rows), err)
	}
	if rows[0].Subject != "agent:none" || rows[0].Obj != "Agent/global/a" {
		t.Fatalf("denied 行主体/对象不符: %+v", rows[0])
	}
	if rows[0].Ts.IsZero() {
		t.Fatal("时间戳应往返")
	}
	// repository 双写行（不携主体）→ subject 列空串。
	if err := s.AppendApply(ApplyLog{Obj: "Agent/global/a", Result: "ok", FieldManager: "agent:x"}); err != nil {
		t.Fatal(err)
	}
	okRows, err := s.QueryApply("ok")
	if err != nil || len(okRows) != 1 || okRows[0].Subject != "" {
		t.Fatalf("ok 行主体应为空（归属看 fieldManager）: %+v err=%v", okRows, err)
	}
}

// TestMigrateAddsSubjectColumn 既有 v1 库无损升列：老 schema（无 subject）打开 → ALTER 补列 → 可写可查。
func TestMigrateAddsSubjectColumn(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "legacy.db")
	// 手工建 v1 老库（无 subject 列）。
	legacy, err := sql.Open(DriverName, path)
	if err != nil {
		t.Fatal(err)
	}
	const oldDDL = `CREATE TABLE apply_log (
		ts TEXT NOT NULL, field_manager TEXT NOT NULL DEFAULT '', obj TEXT NOT NULL,
		resource_version INTEGER NOT NULL DEFAULT 0, result TEXT NOT NULL,
		policy_decisions TEXT NOT NULL DEFAULT '', reason_ref TEXT NOT NULL DEFAULT '');`
	if _, err := legacy.Exec(oldDDL); err != nil {
		t.Fatal(err)
	}
	if _, err := legacy.Exec(`INSERT INTO apply_log (ts, obj, result) VALUES ('2026-01-01T00:00:00Z', 'Agent/global/old', 'ok')`); err != nil {
		t.Fatal(err)
	}
	if err := legacy.Close(); err != nil {
		t.Fatal(err)
	}
	// OpenSQLite 迁移打开：升列后旧行 subject='' 可读，新行可携主体。
	s, err := OpenSQLite(path)
	if err != nil {
		t.Fatalf("迁移打开: %v", err)
	}
	defer s.Close()
	rows, err := s.QueryApply("ok")
	if err != nil || len(rows) != 1 || rows[0].Obj != "Agent/global/old" || rows[0].Subject != "" {
		t.Fatalf("旧行往返 = %+v err=%v", rows, err)
	}
	if err := s.AppendApply(ApplyLog{Obj: "Agent/global/new", Result: "denied", Subject: "agent:mig"}); err != nil {
		t.Fatalf("升列后落行: %v", err)
	}
	if n, _ := s.CountApply("denied"); n != 1 {
		t.Fatalf("denied 行 = %d, 要 1", n)
	}
}
