// approval_test.go —— approval_log 审批判定独立账验收（v0.46，ops-log.md §2 v0.46 注记）：
// 落行/闭合集校验、读面过滤与排序（rv 降序+limit 收敛）、账面对账零耦合（Reconcile 不变量
// 不受 approval_log 影响）。
package opslog

import (
	"testing"
	"time"
)

func TestApprovalAppendAndQuery(t *testing.T) {
	s, err := OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	base := time.Now().UTC().Add(-time.Hour)
	rows := []ApprovalLog{
		{Ts: base, ApprovalID: "draft:Agent/global/a", Obj: "Agent/global/a", Decision: DecisionGranted,
			JudgedBy: "console-operator", FieldManager: "console:approve", ResourceVersion: 1, ReasonRef: "console:approve"},
		{Ts: base.Add(time.Minute), ApprovalID: "draft:Agent/global/b", Obj: "Agent/global/b", Decision: DecisionRejected,
			JudgedBy: "console-operator", FieldManager: "draft:delete", ReasonRef: "draft-delete"},
		{Ts: base.Add(2 * time.Minute), ApprovalID: "draft:Agent/global/a", Obj: "Agent/global/a", Decision: DecisionGranted,
			JudgedBy: "anonymous", FieldManager: "console:approve", ResourceVersion: 2, ReasonRef: "console:approve"},
	}
	for _, r := range rows {
		if err := s.AppendApproval(r); err != nil {
			t.Fatal(err)
		}
	}
	// 全量：rv 降序（2,1,rejected 0 沉底但时刻可比——rejected rv=0 在 granted rv=1 之前）。
	all, err := s.QueryApprovals(ApprovalQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 3 {
		t.Fatalf("行数=%d, 要 3", len(all))
	}
	if all[0].ResourceVersion != 2 || all[1].ResourceVersion != 1 {
		t.Fatalf("granted 行应 rv 降序: %d,%d", all[0].ResourceVersion, all[1].ResourceVersion)
	}
	if all[2].Decision != DecisionRejected || all[2].ResourceVersion != 0 {
		t.Fatalf("rejected 行应保留 rv=0: %+v", all[2])
	}
	// decision 过滤。
	granted, err := s.QueryApprovals(ApprovalQuery{Decision: DecisionGranted})
	if err != nil || len(granted) != 2 {
		t.Fatalf("granted 行=%d err=%v, 要 2", len(granted), err)
	}
	rejected, err := s.QueryApprovals(ApprovalQuery{Decision: DecisionRejected})
	if err != nil || len(rejected) != 1 {
		t.Fatalf("rejected 行=%d err=%v, 要 1", len(rejected), err)
	}
	// obj 精确过滤。
	one, err := s.QueryApprovals(ApprovalQuery{Obj: "Agent/global/b"})
	if err != nil || len(one) != 1 || one[0].ApprovalID != "draft:Agent/global/b" {
		t.Fatalf("obj 过滤=%+v err=%v", one, err)
	}
	// limit 收敛。
	if lim, err := s.QueryApprovals(ApprovalQuery{Limit: 1}); err != nil || len(lim) != 1 {
		t.Fatalf("limit=1 行=%d err=%v", len(lim), err)
	}
	// 计数（探针「查询代替日志」取证面）。
	if n, err := s.CountApprovals(DecisionGranted); err != nil || n != 2 {
		t.Fatalf("CountApprovals(granted)=%d err=%v, 要 2", n, err)
	}
	if n, err := s.CountApprovals(""); err != nil || n != 3 {
		t.Fatalf("CountApprovals(all)=%d err=%v, 要 3", n, err)
	}
	// 闭合集校验：decision 集外拒、approval_id/obj 必填。
	if err := s.AppendApproval(ApprovalLog{Obj: "Agent/global/c", Decision: "maybe"}); err == nil {
		t.Fatal("decision 闭合集外应被拒")
	}
	if err := s.AppendApproval(ApprovalLog{Obj: "Agent/global/c", Decision: DecisionGranted}); err == nil {
		t.Fatal("缺 approval_id 应被拒")
	}
	if err := s.AppendApproval(ApprovalLog{ApprovalID: "draft:x", Decision: DecisionRejected}); err == nil {
		t.Fatal("缺 obj 应被拒")
	}
	if _, err := s.QueryApprovals(ApprovalQuery{Decision: "maybe"}); err == nil {
		t.Fatal("查询 decision 闭合集外应被拒")
	}
}

func TestApprovalLedgerUncoupledFromReconcile(t *testing.T) {
	s, err := OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	// 既有不变量：apply_log(ok) ≙ event_log(DocApplied)。
	if err := s.AppendApply(ApplyLog{Obj: "Agent/global/a", Result: "ok", ResourceVersion: 1}); err != nil {
		t.Fatal(err)
	}
	if err := s.AppendEvent(EventLog{EventID: "e1", Type: "DocApplied"}); err != nil {
		t.Fatal(err)
	}
	if diff, err := s.Reconcile(); err != nil || diff != 0 {
		t.Fatalf("基线对账 diff=%d err=%v, 要 0", diff, err)
	}
	// 审批账落行不进对账不变量（双账并存零耦合——L46-07 判据）。
	if err := s.AppendApproval(ApprovalLog{
		ApprovalID: "draft:Agent/global/a", Obj: "Agent/global/a", Decision: DecisionGranted, ResourceVersion: 1,
	}); err != nil {
		t.Fatal(err)
	}
	if err := s.AppendApproval(ApprovalLog{
		ApprovalID: "draft:Agent/global/b", Obj: "Agent/global/b", Decision: DecisionRejected,
	}); err != nil {
		t.Fatal(err)
	}
	if diff, err := s.Reconcile(); err != nil || diff != 0 {
		t.Fatalf("审批账落行后对账 diff=%d err=%v, 要 0（approval_log 不进 Reconcile 不变量）", diff, err)
	}
}
