// opslog_test.go —— ops DB SQLite 实现验收：建表落行、持久化、库内对账不变量。
package opslog

import (
	"database/sql"
	"fmt"
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

// --- draft 表（ops-log.md §2 扩行；草稿持久化部署决策：--ops-db 持久化面） ---

func TestDraftUpsertListDelete(t *testing.T) {
	s, err := OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	created := time.Date(2026, 11, 22, 9, 0, 0, 0, time.UTC)
	updated := created.Add(30 * time.Minute)
	row := DraftRow{
		Kind: "Agent", Scope: "global", Name: "ci-runner",
		Content: []byte("apiVersion: core/v1\nkind: Agent\n"), FieldManager: "agent/dev",
		ReasonRef: "drift-report/7", BaseRV: 4, CreatedAt: created, UpdatedAt: updated,
	}
	if err := s.UpsertDraft(row); err != nil {
		t.Fatalf("落 draft 行: %v", err)
	}
	rows, err := s.ListDrafts()
	if err != nil || len(rows) != 1 {
		t.Fatalf("ListDrafts = %d 行 err=%v", len(rows), err)
	}
	got := rows[0]
	if got.Kind != "Agent" || got.Scope != "global" || got.Name != "ci-runner" ||
		string(got.Content) != string(row.Content) || got.FieldManager != "agent/dev" ||
		got.ReasonRef != "drift-report/7" || got.BaseRV != 4 ||
		!got.CreatedAt.Equal(created) || !got.UpdatedAt.Equal(updated) {
		t.Fatalf("字段回读不符: %+v", got)
	}
	// 同键覆盖=最新代改稿：created_at 定格、updated_at/内容更新、仍只一行。
	if err := s.UpsertDraft(DraftRow{
		Kind: "Agent", Scope: "global", Name: "ci-runner", Content: []byte("v2"),
		FieldManager: "agent/reviewer", BaseRV: 6, CreatedAt: created, UpdatedAt: updated.Add(time.Hour),
	}); err != nil {
		t.Fatalf("覆稿: %v", err)
	}
	rows, err = s.ListDrafts()
	if err != nil || len(rows) != 1 {
		t.Fatalf("覆稿后行数 = %d err=%v", len(rows), err)
	}
	if !rows[0].CreatedAt.Equal(created) || !rows[0].UpdatedAt.Equal(updated.Add(time.Hour)) ||
		string(rows[0].Content) != "v2" || rows[0].FieldManager != "agent/reviewer" || rows[0].BaseRV != 6 {
		t.Fatalf("覆稿后应为最新稿且 created_at 定格: %+v", rows[0])
	}
	// 弃稿 + 幂等（键不存在的删除=成功 0 行）。
	if err := s.DeleteDraft("Agent", "global", "ci-runner"); err != nil {
		t.Fatalf("删稿: %v", err)
	}
	if err := s.DeleteDraft("Agent", "global", "ci-runner"); err != nil {
		t.Fatalf("幂等删稿: %v", err)
	}
	if n, err := s.CountDraft(); err != nil || n != 0 {
		t.Fatalf("删后 draft 行 = %d err=%v", n, err)
	}
	// 必填校验：键不全 / 无主 / 空稿。
	if err := s.UpsertDraft(DraftRow{Scope: "global", Name: "x", FieldManager: "m", Content: []byte("c")}); err == nil {
		t.Fatal("缺 kind 应被拒")
	}
	if err := s.UpsertDraft(DraftRow{Kind: "Agent", Scope: "global", Name: "x", Content: []byte("c")}); err == nil {
		t.Fatal("缺 field_manager 应被拒")
	}
	if err := s.UpsertDraft(DraftRow{Kind: "Agent", Scope: "global", Name: "x", FieldManager: "m"}); err == nil {
		t.Fatal("空稿应被拒")
	}
}

func TestDraftPersistenceAcrossReopen(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "ops.db")
	s, err := OpenSQLite(path)
	if err != nil {
		t.Fatal(err)
	}
	row := DraftRow{Kind: "Agent", Scope: "global", Name: "persist-me", Content: []byte("正文"),
		FieldManager: "agent/dev", CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
	if err := s.UpsertDraft(row); err != nil {
		t.Fatal(err)
	}
	// draft 表不参与 §3 对账不变量（对账对象=apply_log(ok) vs DocApplied；草稿是缓存非账）。
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s2, err := OpenSQLite(path) // 重开=重启恢复入口（cored --ops-db 启动恢复读的就是它）
	if err != nil {
		t.Fatal(err)
	}
	defer s2.Close()
	rows, err := s2.ListDrafts()
	if err != nil || len(rows) != 1 || rows[0].Name != "persist-me" || string(rows[0].Content) != "正文" {
		t.Fatalf("重开后恢复稿: %d 行 err=%v %+v", len(rows), err, rows)
	}
	if diff, err := s2.Reconcile(); err != nil || diff != 0 {
		t.Fatalf("draft 表不应影响对账: diff=%d err=%v", diff, err)
	}
}

// TestQueryEventsOrderLimitClamp 事件归档读查询验收（v0.36 console 读面首段底层）：
// 最新优先排序、空 Limit=缺省 50、超上限收敛 500（读面防拖库）。
func TestQueryEventsOrderLimitClamp(t *testing.T) {
	s, err := OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	base := time.Now().UTC().Add(-time.Hour)
	// 缺省 Limit（0）下落 3 行 + type 过滤命中 1：全量小库排序断言即可。
	for i, typ := range []string{"Escalate", "DocApplied", "DocApplied"} {
		if err := s.AppendEvent(EventLog{
			Ts: base.Add(time.Duration(i) * time.Minute), EventID: fmt.Sprintf("evt-%d", i),
			Type: typ, Severity: "info",
		}); err != nil {
			t.Fatal(err)
		}
	}
	rows, err := s.QueryEvents(EventQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 3 || rows[0].EventID != "evt-2" || rows[2].EventID != "evt-0" {
		t.Fatalf("要最新优先（evt-2,evt-1,evt-0），得 %+v", rows)
	}
	rows, err = s.QueryEvents(EventQuery{Type: "Escalate"})
	if err != nil || len(rows) != 1 || rows[0].EventID != "evt-0" {
		t.Fatalf("type 过滤要 1 行 evt-0，得 %+v err=%v", rows, err)
	}
	// 超上限收敛：落 MaxEventLimit+3 行，Limit>上限 → 只出 500。
	for i := 0; i < MaxEventLimit+3; i++ {
		if err := s.AppendEvent(EventLog{
			Ts: base.Add(time.Duration(i) * time.Second), EventID: fmt.Sprintf("bulk-%d", i),
			Type: "DocApplied",
		}); err != nil {
			t.Fatal(err)
		}
	}
	rows, err = s.QueryEvents(EventQuery{Limit: MaxEventLimit + 100})
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != MaxEventLimit {
		t.Fatalf("超上限应收敛 %d，得 %d", MaxEventLimit, len(rows))
	}
	if rows[0].EventID[:4] != "bulk" {
		t.Fatalf("收敛后最新行=%s, 要 bulk 族（最新优先）", rows[0].EventID)
	}
}
