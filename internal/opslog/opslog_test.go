// opslog_test.go —— ops DB SQLite 实现验收：建表落行、持久化、库内对账不变量。
package opslog

import (
	"path/filepath"
	"testing"
	"time"
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
