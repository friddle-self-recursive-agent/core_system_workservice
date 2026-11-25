// approval_test.go —— v0.56 R-10 派单 ask 批准动作面 + 通知中心收件箱投影单测。
// （workflow-runtime.md §11.4：ApprovalNeeded→人批→spawn 投影；console-adoption.md §1.25
// 四类判定分类汇总——机械判定=读面现读自分类。）
package flowbind

import (
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
)

// L58-05 派单受理同拍挂批准投影：Dispatch→pending 行（ApprovalID=ConnectorOps/<scope>/<conn>
// 同键；autoLevel=ask 数据面）。
func TestDispatchAttachesApproval(t *testing.T) {
	log := NewApprovalLog(nil)
	syncA := &SyncApprovals{Log: log}
	d := NewDispatchLog(nil)
	d.Approvals = syncA
	d.Dispatch(DispatchRecord{Agent: DispatchAgent, Status: DispatchAccepted, AutoLevel: DispatchAutoLevel,
		Action: "created", Kind: DispatchKindWorkflow, Workflow: "wf-a", Scope: "global",
		Connector: "wf-a", RV: 1, At: time.Now()})
	st := log.State()
	if len(st.Records) != 1 {
		t.Fatalf("批准投影行数=%d（期望 1——受理同拍挂 pending）", len(st.Records))
	}
	r := st.Records[0]
	if r.State != ApprovalPending || r.AutoLevel != DispatchAutoLevel {
		t.Fatalf("投影词形=%s/%s（期望 pending/ask）", r.State, r.AutoLevel)
	}
	if r.ApprovalID != "ConnectorOps/global/wf-a" {
		t.Fatalf("批准对象键=%s（期望 ConnectorOps/global/wf-a——apply_log Obj 同键精确对上）", r.ApprovalID)
	}
	if !syncA.ApproveIdle() == false {
		t.Fatal("挂起待决时 ApproveIdle 应为 false（watchdog 真实词形）")
	}
}

// L58-06 人批=Apply 判定翻转（红线 1 机械落点）：workflow:approve ok 行→approved
// （judgedBy/ts 随行）；驳回=workflow:reject 行→rejected；无关 fieldManager 零翻转；
// denied 行不判定。
func TestSyncApprovalsFlipByApplyRows(t *testing.T) {
	log := NewApprovalLog(nil)
	syncA := &SyncApprovals{Log: log}
	d := NewDispatchLog(nil)
	d.Approvals = syncA
	d.Dispatch(DispatchRecord{Kind: DispatchKindWorkflow, Workflow: "wf-a", Scope: "global",
		Connector: "wf-a", RV: 1, At: time.Now()})
	d.Dispatch(DispatchRecord{Kind: DispatchKindWorkflow, Workflow: "wf-b", Scope: "global",
		Connector: "wf-b", RV: 1, At: time.Now()})
	ts := time.Now()
	flips := syncA.Sync([]opslog.ApplyLog{
		{Ts: ts, FieldManager: ApprovalManager, Subject: "console-operator", Obj: "ConnectorOps/global/wf-a", ResourceVersion: 2, Result: "ok"},
		{Ts: ts, FieldManager: ApprovalRejecter, Subject: "console-operator", Obj: "ConnectorOps/global/wf-b", ResourceVersion: 2, Result: "ok"},
		{Ts: ts, FieldManager: "console:approve", Obj: "ConnectorOps/global/wf-a", Result: "ok"},                        // 非 ask 判定串=零翻转
		{Ts: ts, FieldManager: ApprovalManager, Obj: "ConnectorOps/global/wf-x", Result: "ok"},                          // 无挂起投影=零发明
		{Ts: ts, FieldManager: ApprovalManager, Obj: "ConnectorOps/global/ghost", ResourceVersion: 1, Result: "denied"}, // denied 不判定
	})
	if flips != 2 {
		t.Fatalf("翻转行数=%d（期望 2——approve/reject 各一，无关行零翻转）", flips)
	}
	st := log.State()
	got := map[string]string{}
	var judgedBy, rejected string
	for _, r := range st.Records {
		got[r.Connector] = r.State
		if r.State == ApprovalApproved {
			judgedBy = r.JudgedBy
		}
		if r.State == ApprovalRejected {
			rejected = r.Connector
		}
	}
	if got["wf-a"] != ApprovalApproved || got["wf-b"] != ApprovalRejected {
		t.Fatalf("判定词形=%v（期望 wf-a=approved wf-b=rejected）", got)
	}
	if judgedBy != "console-operator" {
		t.Fatalf("判定主体=%q（apply_log subject 投影缺席）", judgedBy)
	}
	if rejected != "wf-b" {
		t.Fatalf("驳回词形错位: %s", rejected)
	}
	if !syncA.ApproveIdle() {
		t.Fatal("全终态后 ApproveIdle 应为 true")
	}
}

// L58-07 proposal 派单批准键：Proposal/<scope>/<name>（对象族待批键——批准动作=Apply Proposal）。
func TestProposalApprovalKey(t *testing.T) {
	rec := DispatchRecord{Kind: DispatchKindProposal, Scope: "global", Proposal: "p1"}
	if matchingKeyOf(rec) != "Proposal/global/p1" {
		t.Fatalf("proposal 批准键=%s", matchingKeyOf(rec))
	}
}

// L58-08 收件箱四类判定分类（§1.25）：NeedsHuman/Escalate→NeedsHuman 类；ApprovalNeeded→
// ApprovalNeeded；submitted 提案→Proposal；草稿→DocReview；空输入=零行空态（不冒充）。
func TestBuildInboxClasses(t *testing.T) {
	now := time.Date(2026, 11, 27, 9, 0, 0, 0, time.UTC)
	events := []opslog.EventLog{
		{Ts: now, EventID: "e1", Type: "NeedsHuman", Involved: "Proposal/global/p1"},
		{Ts: now, EventID: "e2", Type: "Escalate", Involved: "Session/cored/s1"},
		{Ts: now, EventID: "e3", Type: "ApprovalNeeded", Involved: "ConnectorOps/global/wf-a"},
		{Ts: now, EventID: "e4", Type: "DocApplied", Involved: "Workflow/global/wf-a"}, // 非人注意类=零行
	}
	drafts := []InboxDraftRow{{Kind: "Workflow", Scope: "global", Name: "wf-a", FieldManager: "agent:probe", UpdatedAt: now.UTC().Format(time.RFC3339Nano)}}
	proposals := []InboxProposalRow{
		{Name: "p2", Scope: "global", Title: "对话式录入", Status: "submitted", UpdatedAt: now.UTC().Format(time.RFC3339Nano)},
		{Name: "p3", Scope: "global", Title: "已批", Status: "approved", UpdatedAt: now.UTC().Format(time.RFC3339Nano)},
	}
	box := BuildInbox(events, drafts, proposals, nil)
	classes := map[string]int{}
	for _, it := range box.Items {
		classes[it.Class]++
	}
	if classes[inboxNeedsHuman] != 2 || classes[inboxApprovalNeeded] != 1 ||
		classes[inboxProposal] != 1 || classes[inboxDocReview] != 1 {
		t.Fatalf("四类计数=%v（期望 NeedsHuman=2 ApprovalNeeded=1 Proposal=1 DocReview=1）", classes)
	}
	if box.Count != 5 {
		t.Fatalf("总计数=%d（期望 5）", box.Count)
	}
	// 类序：NeedsHuman 最前（人注意优先）；行深链在册（规则 6 页面兜底入口）。
	if box.Items[0].Class != inboxNeedsHuman {
		t.Fatalf("首行类=%s（期望 NeedsHuman 人注意优先）", box.Items[0].Class)
	}
	for _, it := range box.Items {
		if it.Link == "" || it.Link[0] != '#' {
			t.Fatalf("行 %s 深链缺席/非法: %q", it.Class, it.Link)
		}
	}
	empty := BuildInbox(nil, nil, nil, nil)
	if empty.Count != 0 || len(empty.Items) != 0 || empty.Classes[inboxNeedsHuman] != 0 {
		t.Fatalf("空输入投影=%+v（期望零行空态显式）", empty)
	}
}

// L59-01 spawn 接线声明投影（workflow-runtime.md §11.7——拉起投影面末段）：approved 批准
// 投影行 → {declared:true, wired:false, target:apply-workflow}；pending/rejected 行=零声明
// （零值键恒在）；wired 恒 false=真实拉起 deferred 的机械锚（红线 a——本面零拉起执行器）。
func TestApprovalSpawnDeclaration(t *testing.T) {
	approved := ApprovalRecord{Kind: DispatchKindWorkflow, Workflow: "wf-a", Scope: "global",
		Connector: "wf-a", State: ApprovalApproved, AutoLevel: DispatchAutoLevel, RV: 2,
		EventID: "evt-DocApplied-1", ApprovalID: "ConnectorOps/global/wf-a"}
	d := ApprovalSpawnDeclaration(approved)
	if !d.Declared || d.Wired || d.Target != DispatchAgent || d.RV != 2 || d.EventID != "evt-DocApplied-1" {
		t.Fatalf("approved 行声明=%+v（期望 declared=true wired=false target=apply-workflow rv=2）", d)
	}
	pending := approved
	pending.State = ApprovalPending
	if pd := ApprovalSpawnDeclaration(pending); pd.Declared || pd.Wired {
		t.Fatalf("pending 行=%+v（期望零声明——接线面未进入）", pd)
	}
	rejected := approved
	rejected.State = ApprovalRejected
	if rd := ApprovalSpawnDeclaration(rejected); rd.Declared {
		t.Fatalf("rejected 行=%+v（期望零声明——驳回≠接线）", rd)
	}
}

// L59-02 判定权威=apply_log 行：SyncApprovals 仅凭 workflow:approve ok 行翻转 approved →
// spawn 声明在册（进程内投影环重启归零后，判定行重放=声明再现——账本事实投影单点真相）。
func TestSpawnDeclarationFromApplyLogOnly(t *testing.T) {
	log := NewApprovalLog(nil)
	syncA := &SyncApprovals{Log: log}
	d := NewDispatchLog(nil)
	d.Approvals = syncA
	d.Dispatch(DispatchRecord{Kind: DispatchKindWorkflow, Workflow: "wf-a", Scope: "global",
		Connector: "wf-a", RV: 1, At: time.Now()})
	flips := syncA.Sync([]opslog.ApplyLog{{Ts: time.Now(), FieldManager: ApprovalManager,
		Subject: "console-operator", Obj: "ConnectorOps/global/wf-a", ResourceVersion: 2, Result: "ok"}})
	if flips != 1 {
		t.Fatalf("翻转行数=%d（期望 1）", flips)
	}
	st := log.State()
	if len(st.Records) != 1 || st.Records[0].State != ApprovalApproved {
		t.Fatalf("投影态=%+v（期望 approved 单行）", st.Records)
	}
	decl := ApprovalSpawnDeclaration(st.Records[0])
	if !decl.Declared || decl.Wired || decl.Target != DispatchAgent {
		t.Fatalf("apply_log 判定行 spawn 声明=%+v（期望 declared=true wired=false target=apply-workflow——判定权威=apply_log 行）", decl)
	}
	if decl.RV != 2 || st.Records[0].EventID != "" {
		t.Fatalf("声明载荷=%+v（rv=%d 应=判定行 rv；eventId 应空——Sync 不复制派单载荷，声明面零拉起事实）", decl, decl.RV)
	}
}
