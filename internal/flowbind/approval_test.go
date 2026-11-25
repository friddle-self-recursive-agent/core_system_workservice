// approval_test.go —— v0.56 R-10 派单 ask 批准动作面 + 通知中心收件箱投影单测。
// （workflow-runtime.md §11.4：ApprovalNeeded→人批→spawn 投影；console-adoption.md §1.25
// 四类判定分类汇总——机械判定=读面现读自分类。）
package flowbind

import (
	"strings"
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
		{Ts: ts, FieldManager: ApprovalManager, Obj: "ConnectorOps/global/wf-x", Result: "ok"},                          // 无挂起投影=零发明
		{Ts: ts, FieldManager: ApprovalManager, Obj: "ConnectorOps/global/ghost", ResourceVersion: 1, Result: "denied"}, // denied 不判定
		{Ts: ts, FieldManager: "unknown:approve", Obj: "ConnectorOps/global/wf-a", Result: "ok"},                        // 非判定族串=零翻转（v0.58a 族外）
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

// L61-01 needsHuman 类 Escalate 行 digest 尾随空串收口（loop-61 队列 1——§1.25 收件箱
// 展示面余量）：involved 缺席（空/纯空白）→ 降级词形 `<Type> · （involved 缺席）`——无尾随
// 分隔符悬空；involved 在册 → 既有词形 `<Type> · <Involved>` 不变；EventID 深链照旧
// （机读判据不变——纯展示面零判定语义）。ApprovalNeeded 行同款拼装随行（单点 helper）。
func TestBuildInboxDigestWithoutInvolved(t *testing.T) {
	now := time.Date(2026, 11, 27, 10, 0, 0, 0, time.UTC)
	events := []opslog.EventLog{
		{Ts: now, EventID: "e-empty", Type: "Escalate", Involved: ""},
		{Ts: now, EventID: "e-blank", Type: "NeedsHuman", Involved: "   "},
		{Ts: now, EventID: "e-ok", Type: "Escalate", Involved: "Session/cored/s1"},
		{Ts: now, EventID: "e-ap", Type: "ApprovalNeeded", Involved: ""},
	}
	box := BuildInbox(events, nil, nil, nil)
	got := map[string]string{}
	for _, it := range box.Items {
		got[it.EventID] = it.Digest
	}
	if d := got["e-empty"]; d != "Escalate · （involved 缺席）" {
		t.Fatalf("Escalate involved 缺席 digest=%q（期望降级词形——尾随分隔符不得悬空）", d)
	}
	if d := got["e-blank"]; d != "NeedsHuman · （involved 缺席）" {
		t.Fatalf("NeedsHuman involved 纯空白 digest=%q（期望降级词形）", d)
	}
	if d := got["e-ok"]; d != "Escalate · Session/cored/s1" {
		t.Fatalf("Escalate involved 在册 digest=%q（期望既有词形不变）", d)
	}
	if d := got["e-ap"]; d != "ApprovalNeeded · （involved 缺席）" {
		t.Fatalf("ApprovalNeeded involved 缺席 digest=%q（期望同款降级词形）", d)
	}
	// 机读判据不变：EventID 深链照旧（事件详情兜底路由）。
	for _, it := range box.Items {
		if it.Link != "#/event/"+it.EventID {
			t.Fatalf("行 %s 深链=%s（期望 #/event/<eventId> 不变）", it.EventID, it.Link)
		}
		if strings.HasSuffix(it.Digest, " · ") {
			t.Fatalf("行 %s digest 尾随分隔符悬空: %q", it.EventID, it.Digest)
		}
	}
	// 类归属不变：缺席 involved 不改 NeedsHuman/ApprovalNeeded 分类。
	classes := map[string]int{}
	for _, it := range box.Items {
		classes[it.Class]++
	}
	if classes[inboxNeedsHuman] != 3 || classes[inboxApprovalNeeded] != 1 {
		t.Fatalf("分类计数=%v（期望 NeedsHuman=3 ApprovalNeeded=1——纯展示面零判定语义）", classes)
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

// ───────────────── v0.58 派单批准动作面 console 动作化（§11.5 动作面——console-adoption §1.33） ─────────────────

// L60-01 动作判定串单点真相：DecisionAction 拼装=既有写路径判定溯源串（workflow:approve|
// workflow:reject——红线 1 机械落点零新词形）；对象键=matchingKeyOf 同键（kind=workflow 派单→
// 骨架 ConnectorOps 对象 / kind=proposal 派单→Proposal 对象）；rv=批准投影行 rv（乐观锁基线=
// 派单落库后版本）；未知动作=零输入（动作面不发明写词形）。
func TestDecisionActionInputs(t *testing.T) {
	wf := ApprovalRecord{Kind: DispatchKindWorkflow, Workflow: "wf-a", Scope: "global",
		Connector: "wf-a", RV: 7, ApprovalID: "ConnectorOps/global/wf-a"}
	in, ok := DecisionAction(wf, DecisionApprove)
	if !ok {
		t.Fatal("approve 动作应产输入")
	}
	if in.Kind != "ConnectorOps" || in.Scope != "global" || in.Name != "wf-a" {
		t.Fatalf("approve 对象键=%+v（期望 ConnectorOps/global/wf-a——matchingKeyOf 同键）", in)
	}
	if in.FieldManager != ApprovalManager || in.ReasonRef != "workflow:approve:"+wf.ApprovalID {
		t.Fatalf("approve 判定串=%s/%s（期望 workflow:approve + reasonRef 单点拼装）", in.FieldManager, in.ReasonRef)
	}
	if in.ResourceVersion != 7 {
		t.Fatalf("rv=%d（期望 7——批准投影行 rv=乐观锁基线）", in.ResourceVersion)
	}
	pr := ApprovalRecord{Kind: DispatchKindProposal, Workflow: "", Scope: "global",
		Proposal: "p1", RV: 3, ApprovalID: "Proposal/global/p1"}
	inR, ok := DecisionAction(pr, DecisionReject)
	if !ok {
		t.Fatal("reject 动作应产输入")
	}
	// v0.58a 判定主体分族：proposal 派单行动作=提案判定通路（console:reject——fromDraft 翻转稿
	// 判定串；活体实证 GAP-Q60-2 收口——判定串与 console 动作面同串=判定行可被族表命中）。
	if inR.Kind != "Proposal" || inR.Name != "p1" || inR.FieldManager != ProposalApprovalRejecter {
		t.Fatalf("proposal 驳回对象=%+v（期望 Proposal/global/p1 × console:reject——提案判定通路同串）", inR)
	}
	inP, ok := DecisionAction(pr, DecisionApprove)
	if !ok || inP.FieldManager != ProposalApprovalManager {
		t.Fatalf("proposal 批准判定串=%+v（期望 console:approve——提案判定通路同串）", inP)
	}
	if _, ok := DecisionAction(wf, "spawn"); ok {
		t.Fatal("非判定动作（spawn）不得产输入——动作面零发明写词形（零拉起执行器红线 a）")
	}
}

// L60-02 动作 → 判定 → 声明全链投影一致性（动作面=既有 Apply 判定通路的输入拼装——本测以
// DecisionAction 输出喂 Sync 判定行：同键 ok 行翻转 approved ∧ spawn 声明随行——输入串与
// 判定串漂移即本测红）。
func TestDecisionActionFeedsSync(t *testing.T) {
	log := NewApprovalLog(nil)
	syncA := &SyncApprovals{Log: log}
	d := NewDispatchLog(nil)
	d.Approvals = syncA
	d.Dispatch(DispatchRecord{Kind: DispatchKindWorkflow, Workflow: "wf-a", Scope: "global",
		Connector: "wf-a", RV: 1, At: time.Now()})
	in, ok := DecisionAction(log.State().Records[0], DecisionApprove)
	if !ok {
		t.Fatal("approve 动作应产输入")
	}
	flips := syncA.Sync([]opslog.ApplyLog{{Ts: time.Now(), FieldManager: in.FieldManager,
		Subject: "console-operator", Obj: in.Kind + "/" + in.Scope + "/" + in.Name,
		ResourceVersion: 2, Result: "ok"}})
	if flips != 1 {
		t.Fatalf("翻转行数=%d（期望 1——动作输入串与判定串同源）", flips)
	}
	r := log.State().Records[0]
	if r.State != ApprovalApproved || !ApprovalSpawnDeclaration(r).Declared {
		t.Fatalf("动作后投影态=%s（期望 approved+声明在册）", r.State)
	}
}

// L60-06（v0.58a 判定源族收口——GAP-Q60-2 回归锚）：proposal 派单行挂起 × console:approve
// 判定行（提案判定通路，fromDraft 翻转稿同串）→ 投影 approved ∧ spawn 声明随行（declared=true
// ∧ wired 恒 false）。收口前该词形=恒 pending 恒零值（活体实证）——本测漂移即红。
func TestProposalDispatchConsoleFamilySync(t *testing.T) {
	log := NewApprovalLog(nil)
	syncA := &SyncApprovals{Log: log}
	d := NewDispatchLog(nil)
	d.Approvals = syncA
	d.Dispatch(DispatchRecord{Kind: DispatchKindProposal, Scope: "global",
		Proposal: "p-fam", RV: 2, AutoLevel: DispatchAutoLevel, At: time.Now()})
	in, ok := DecisionAction(syncA.Log.State().Records[0], DecisionApprove)
	if !ok || in.Kind != "Proposal" || in.FieldManager != ProposalApprovalManager {
		t.Fatalf("DecisionAction proposal 分族=%+v ok=%v（期望 Proposal × console:approve）", in, ok)
	}
	n := syncA.Sync([]opslog.ApplyLog{{
		Ts: time.Now(), FieldManager: in.FieldManager, Obj: "Proposal/global/p-fam",
		ResourceVersion: in.ResourceVersion, Result: "ok",
	}})
	if n != 1 {
		t.Fatalf("族表命中翻转行数=%d（期望 1——console:approve ∈ 判定源族）", n)
	}
	r := syncA.Log.State().Records[0]
	if r.State != ApprovalApproved {
		t.Fatalf("proposal 派单行判定态=%s（期望 approved——判定串分叉收口）", r.State)
	}
	sp := ApprovalSpawnDeclaration(r)
	if !sp.Declared || sp.Wired || sp.Target != DispatchAgent {
		t.Fatalf("spawn 声明=%+v（期望 declared=true wired=false target=apply-workflow——红线 a 机械锚）", sp)
	}
}
