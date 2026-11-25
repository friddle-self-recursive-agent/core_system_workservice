// approval.go —— 派单 ask 批准动作面（v0.56，R-10——workflow-runtime.md §11.4 第一段）。
// 设计依据：instance 仓 runtime/workflow-runtime.md §11.2/§11.3 余量行「ask 批准动作面
// （ApprovalNeeded→人批→spawn 投影）随 agent-session 域落地轮扩」+ agent-session.md §2
// （AUTOLEVEL：ask=Planning 完成后挂起→ApprovalNeeded→上级批/驳回）+ agent-session.md §7
// 自愈链 ApprovalNeeded（引擎内部源：权限分诊路——v0.54 既有注册类型，不新增）。
//
// 触发接线：派单受理（v0.50 既有 sink.Dispatch）同拍记一条 pending 批准投影（审批同步器=
// autoLevel=ask 的派单默认挂批准——受红线 a 约束：不 enable workflow、不自动执行代码，
// 本面只做「批准动作面投影」，真实拉起/执行仍 observe）。人批/驳回=Apply 判定（红线 1）：
// fieldManager=workflow:approve|workflow:reject 的 ok 行按 obj 匹配 → 同拍翻转为 approved|rejected。
//
// 诚实投影（受理≠拉起同构）：approved 派单投影 ≠ spawn 已拉起——spawn 接线随 agent-session
// 域真实拉起段落地（declared-deferred）。观测面非账本：进程内投影环，持久态=apply_log
// 审计（workflow:approve 行）+骨架对象本体+dispatches 受理行。
//
// 解耦口径：零新写路径零旁路（apply_log 行集=既有账面；判定=内存匹配）；sink 装配层注入
// （nil=静默零投影——headless 载体/旧装配零影响）。
package flowbind

import (
	"sort"
	"sync"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
)

// 审批词汇（agent-session.md §2 ask 分诊词汇同源；apply_log fieldManager 溯源串）。
const (
	ApprovalPending  = "pending"          // 批准投影初态（ask 派单挂起待批）
	ApprovalApproved = "approved"         // 人批=Apply ok（fieldManager=workflow:approve）
	ApprovalRejected = "rejected"         // 驳回=Apply ok（fieldManager=workflow:reject）
	ApprovalManager  = "workflow:approve" // 批准溯源串（approval_log 同款约定串——apply.md rule 2）
	ApprovalRejecter = "workflow:reject"  // 驳回溯源串（对称扩展——proposals_face.go 同款）
	ApprovalSource   = "dispatch"         // 来源词（批准投影行 source 键——派单 ask 面）
)

// DecisionApprove/DecisionReject 派单批准/驳回动作（v0.58——§11.5 动作面 console 动作化）：
// Apply 判定溯源串常量的动作对（常量本身=v0.56 既有判定词汇，零新词形零旁路——动作面只是
// 把「构造同串 Apply」的调用形态单点化，字面量散用=漂移温床）。
const (
	DecisionApprove = "approve" // 批准动作→ Apply fieldManager=workflow:approve
	DecisionReject  = "reject"  // 驳回动作→ Apply fieldManager=workflow:reject
)

// v0.58a 判定主体分族（console-adoption.md §1.33 缺口收口——活体实证 GAP-Q60-2）：
// kind=proposal 派单行的动作=提案判定通路（console:approve|console:reject——proposalFlip 既有
// fromDraft 判定串，字段归属对象=Proposal），kind=workflow 派单行=骨架对象族（workflow:* 判定串
// ——骨架 ConnectorOps 对象）。两族同一判定词汇（approve|reject），判定源族收口见
// DecisionJudgmentFamilies。字面量散用=漂移温床（DecisionApprove/DecisionReject 同款原则）。
const (
	ProposalApprovalManager  = "console:approve" // 提案批准判定串（E6 审批屏同款——apiserver ApproveFieldManager 同值）
	ProposalApprovalRejecter = "console:reject"  // 提案驳回判定串（对称扩展——apiserver RejectFieldManager 同值）
)

// DecisionJudgmentFamilies 判定溯源串族（v0.58a 判定源族收口——批准投影 Sync 匹配词形的
// 单点声明；返回批准态词形：approved|rejected，非判定串=零值不判定）。两族=两类判定主体
// 的动作词汇（wf 派单动作=骨架对象族 workflow:*；proposal 派单动作=提案判定通路 console:*），
// 判定词汇同一（approve|reject）零新词形——族=溯源串前缀域，非新增判定面。
func DecisionJudgmentFamilies() map[string]string {
	return map[string]string{
		ApprovalManager:          ApprovalApproved, // workflow:approve → approved
		ApprovalRejecter:         ApprovalRejected, // workflow:reject → rejected
		ProposalApprovalManager:  ApprovalApproved, // console:approve → approved（v0.58a 收口）
		ProposalApprovalRejecter: ApprovalRejected, // console:reject → rejected（v0.58a 收口）
	}
}

// DecisionAction 派单批准判定动作的 Apply 输入面（红线 1 机械落点——批准/驳回=Apply 判定，
// 本函数只拼既有写路径的载荷字段：对象键=matchingKeyOf（kind=workflow 派单→骨架 ConnectorOps
// 对象、kind=proposal 派单→Proposal 对象；与 apply_log 判定行 Obj 同键=Sync 精确匹配零模糊），
// fieldManager=workflow:approve|workflow:reject 判定溯源串，resourceVersion=批准投影行 rv
// （派单落库后版本=乐观锁基线——同一派单行的同拍 diff 基线），content 零变更提交（批准判定
// 只需成功事实，键已含对象身份——v0.57 loop-59 探针实证零 content noop 行照落照判定）。
// 本函数零 HTTP 零引擎依赖——调用方（console 写路由）拿载荷走既有 /v1/apply 通路。
func DecisionAction(r ApprovalRecord, decision string) (ApplyActionInput, bool) {
	if decision != DecisionApprove && decision != DecisionReject {
		return ApplyActionInput{}, false // 非判定动作=零输入（动作面不发明写词形）
	}
	fm := ApprovalManager
	if decision == DecisionReject {
		fm = ApprovalRejecter
	}
	kind, name := "ConnectorOps", r.Connector // kind=workflow 派单→骨架 ConnectorOps 对象（matchingKeyOf 同键）
	if r.Kind == DispatchKindProposal || r.Proposal != "" {
		// v0.58a 判定主体分族：kind=proposal 派单行动作=提案判定通路（console:approve|reject
		// ——fromDraft 翻转稿判定串；判定主体=console，判定词汇同一 approve|reject）。
		kind, name = "Proposal", r.Proposal // kind=proposal 派单→Proposal 对象
		if decision == DecisionApprove {
			fm = ProposalApprovalManager
		} else {
			fm = ProposalApprovalRejecter
		}
	}
	return ApplyActionInput{
		Kind: kind, Scope: r.Scope, Name: name,
		FieldManager: fm, ReasonRef: fm + ":" + r.ApprovalID,
		ResourceVersion: r.RV,
	}, true
}

// ApplyActionInput 派单批准判定动作的 Apply 载荷（宽松键零宿主依赖——console 写路由转译为
// /v1/apply JSON；零值字段=不适用）。
type ApplyActionInput struct {
	Kind            string `json:"kind"`
	Scope           string `json:"scope,omitempty"`
	Name            string `json:"name"`
	FieldManager    string `json:"fieldManager"`
	ReasonRef       string `json:"reasonRef,omitempty"`
	ResourceVersion int64  `json:"resourceVersion,omitempty"`
}

// dispatchApprovalKey 批准投影与派单受理行的同拍关联键（matchingKeyOf 同源）。
func dispatchApprovalKey(rec DispatchRecord) string {
	return matchingKeyOf(rec)
}

// matchingKeyOf 审批判定行匹配键（apply_log Obj="kind/scope/name" 三段同构——批准动作的
// Apply 落点对象）：kind=workflow 派单→骨架 ConnectorOps 对象；kind=proposal 派单→Proposal
// 对象。投影行 ApprovalID 与判定行 Obj 同键=精确对上（零模糊匹配）。
func matchingKeyOf(rec DispatchRecord) string {
	if rec.Kind == DispatchKindProposal || rec.Proposal != "" {
		return "Proposal/" + rec.Scope + "/" + rec.Proposal
	}
	return "ConnectorOps/" + rec.Scope + "/" + rec.Connector
}

// ApprovalKeyOf matchingKeyOf 的导出形（/v1/agents/dispatches approval 键消费——同一键源
// 零硬编码；空行零值键照返，调用方按匹配失败处理）。
func ApprovalKeyOf(rec DispatchRecord) string { return matchingKeyOf(rec) }

// ApprovalRecord 一条 ask 批准投影（workflow-runtime.md §11.4：ApprovalNeeded→人批→spawn 投影
// 的动作面数据；审批判定=Apply——apply_log 审计账为准，本投影=观测面）。
type ApprovalRecord struct {
	Seq        int64     // 投影序号（本 sink 进程内单调递增，1 起）
	Kind       string    // 派单源种类（workflow|proposal——DispatchKindWorkflow 同词汇）
	Workflow   string    // workflow 锚名（kind=proposal 空）
	Scope      string    // 对象 scope
	Connector  string    // 骨架对象名（kind=workflow）
	Proposal   string    // proposal 锚名（kind=proposal）
	State      string    // pending|approved|rejected（翻转=Apply 判定同拍——apply_log 行集匹配）
	AutoLevel  string    // 派单 autoLevel（ask——agent-session.md §2 级别数据面）
	RV         int64     // 派单落库后版本（实现侧 diff 基线——dispatches 行 rv 同源）
	EventID    string    // DocApplied 回执（来源链可回指）
	ApprovalID string    // 审批对象键（approval_log approval_id 同词汇：draft:<kind>/<scope>/<name> 形——对象族=kind=ConnectorOps）
	JudgedBy   string    // 判定主体（apply_log 行 subject；pending=空）
	JudgedAt   time.Time // 判定时刻（apply_log 行 ts；pending=零值）
	At         time.Time // 投影记录时刻（派单受理同拍）
}

// ApprovalSink 批准投影落点（装配层注入；实现须非阻塞。nil=静默零投影）。
type ApprovalSink interface {
	Approval(rec ApprovalRecord)
}

// ApprovalOptions 批准投影环可调项（零值可用）。
type ApprovalOptions struct {
	// RingSize 投影环容量（缺省 128）：溢出丢最旧、dropped 计数可见。
	RingSize int
}

// ApprovalState 查询快照（ApprovalLog.State() 返回值）。
type ApprovalState struct {
	Accepted int64            // 累计投影（含已被环挤出的）
	Dropped  int64            // 环溢出丢弃计数（可观测）
	Records  []ApprovalRecord // 环内投影（旧→新）
}

// ApprovalLog ask 批准投影的进程内落点（ApprovalSink 缺省实现；并发安全；DispatchLog 同构
// ——无出站机器，投影即终态面；判定翻转在投影内同步完成——apply_log 账本事实的实时投影）。
type ApprovalLog struct {
	opts ApprovalOptions

	mu       sync.Mutex
	seq      int64
	ring     []ApprovalRecord
	accepted int64
	dropped  int64
}

// NewApprovalLog 构建批准投影落点。
func NewApprovalLog(opts *ApprovalOptions) *ApprovalLog {
	a := &ApprovalLog{}
	if opts != nil {
		a.opts = *opts
	}
	if a.opts.RingSize <= 0 {
		a.opts.RingSize = 128
	}
	return a
}

// Approval 记一条批准投影（环满丢最旧，dropped 计数可见）。
func (a *ApprovalLog) Approval(rec ApprovalRecord) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.seq++
	rec.Seq = a.seq
	a.accepted++
	if excess := len(a.ring) + 1 - a.opts.RingSize; excess > 0 {
		a.ring = a.ring[excess:]
		a.dropped += int64(excess)
	}
	a.ring = append(a.ring, rec)
}

// State 查询快照（拷贝返回；records 旧→新）。
func (a *ApprovalLog) State() ApprovalState {
	a.mu.Lock()
	defer a.mu.Unlock()
	rs := make([]ApprovalRecord, 0, len(a.ring))
	rs = append(rs, a.ring...)
	return ApprovalState{Accepted: a.accepted, Dropped: a.dropped, Records: rs}
}

// SyncApprovals ask 批准投影同步器（派单面装配的审批同步器——workflow-runtime.md §11.4）。
// Attach（监听 Dispatch sink 同拍记 pending 投影）+ Sync（apply_log 行集 → 判定翻转投影——
// 红线 1 机械落点：批准/驳回=Apply 判定，零新写路径）。判定规则：fieldManager=workflow:approve
// → approved / workflow:reject → rejected；同一对象多行取 ts 最大=最新判定；subject=判定主体、
// ts=判定时刻投影随行。
type SyncApprovals struct {
	Log *ApprovalLog
	mu  sync.Mutex
}

// Attach 派单受理同拍记 pending 批准投影（DispatchLog.Dispatch 的同步伴生；autoLevel=ask
// 的派单默认挂批准——approval 记录=dispatch 记录的投影面分身）。nil 接收器=静默。
func (s *SyncApprovals) Attach(rec DispatchRecord) {
	if s == nil || s.Log == nil {
		return
	}
	s.Log.Approval(ApprovalRecord{
		Kind: rec.Kind, Workflow: rec.Workflow, Scope: rec.Scope, Connector: rec.Connector,
		Proposal: rec.Proposal, State: ApprovalPending, AutoLevel: rec.AutoLevel,
		RV: rec.RV, EventID: rec.EventID,
		ApprovalID: approvalIDOf(rec), At: time.Now(),
	})
}

// Sync apply_log 行集 → 判定翻转投影（装配层定时/派单查询面触发；幂等——重复行集零重复
// 翻转：已终态投影不回 pending，后行覆盖前行按 ts）。
func (s *SyncApprovals) Sync(rows []opslog.ApplyLog) int {
	if s == nil || s.Log == nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	st := s.Log.State()
	index := make(map[string]int, len(st.Records))
	for i, r := range st.Records {
		if r.State == ApprovalPending {
			index[matchingKeyOf(r2d(r))] = i
		}
	}
	n := 0
	judgments := DecisionJudgmentFamilies() // v0.58a 判定源族收口（workflow:*+console:* 两族单点表——字面量散用摘除）
	for _, row := range rows {
		state, isJudgment := judgments[row.FieldManager]
		if !isJudgment {
			continue // 非判定族行零翻转（观测面不发明判定——普通 apply 不参与批准判定面）
		}
		if row.Result != "ok" && row.Result != "noop" {
			continue // denied/conflict 行不判定（红线 1：批准=Apply 成功事实）
		}
		i, ok := index[row.Obj]
		if !ok {
			continue // 无挂起投影的判定行（对象键未匹配）=零翻转（观测面不发明投影）
		}
		r := st.Records[i]
		if r.State != ApprovalPending && !r.JudgedAt.Before(row.Ts) {
			continue // 已终态且判定行不新于现投影判定=幂等（重复行集零重复翻转）
		}
		// 原地翻转（观测面=apply_log 账本事实的实时投影——判定行替换 pending 行，不双行并存：
		// pending 与 approved 并存=同一待办双计数的假词形）。rv/判定词随判定行（v0.57 spawn
		// 声明投影的 rv 基线=批准 Apply 落库后版本——接线声明与账本行同源零硬编码）。
		r.State = state
		r.JudgedBy = row.Subject
		r.JudgedAt = row.Ts
		r.RV = row.ResourceVersion
		s.Log.ring[i] = r
		n++
	}
	return n
}

// SyncOnce 一次同步（装配层派单查询面/watchdog 拍触发；生产由 main 装配注入 applyLog 行源）。
// rows=apply_log 行集（obj 精确查询按挂起键逐行拉取由调用方聚合——本方法只做匹配判定）。
func (s *SyncApprovals) SyncOnce(rows []opslog.ApplyLog) int {
	return s.Sync(rows)
}

// ApproveIdle 批准投影是否无挂起待决（真实词形——零挂起≠停摆不冒充；/v1/inbox watchdog
// 投影用）。nil 接收器=空闲（无投影面即无挂起）。
func (s *SyncApprovals) ApproveIdle() bool {
	if s == nil || s.Log == nil {
		return true
	}
	st := s.State()
	for _, r := range st.Records {
		if r.State == ApprovalPending {
			return false
		}
	}
	return true
}

// State 查询快照透出（/v1/inbox approvals.records 载荷源——apiserver.InboxFace 端面）。
func (s *SyncApprovals) State() ApprovalState {
	if s == nil || s.Log == nil {
		return ApprovalState{}
	}
	return s.Log.State()
}

// ApprovalSpawnDeclaration spawn 接线声明投影（v0.57，workflow-runtime.md §11.7——拉起投影
// 面末段：ask 批准动作面（§11.5）→ spawn 接线声明 → 真实拉起（agent-session 域 deferred）。
// 单一真相=approved 批准投影行（人批=Apply workflow:approve ok 行——apply_log 账本事实）；
// 本纯函数零新零旁路零写面，approved 行缺席=声明缺席（诚实词形，不发明）。
//
// **红线 a 机械锚**：wired 恒 false——真实拉起=dsh 会话挂载（agent-session.md §3），本声明
// 面零拉起执行器零自动触发；接线缺席=缺口可见（wired=true 而无执行面=静默缺口，本面禁此词形）。
func ApprovalSpawnDeclaration(r ApprovalRecord) SpawnDeclaration {
	if r.State != ApprovalApproved {
		return SpawnDeclaration{} // 未批准（pending/rejected）=零声明（接线面未进入）
	}
	return SpawnDeclaration{
		Declared:   true,
		Wired:      false, // 真实拉起接线=agent-session 域 declared-deferred（诚实词形）
		Target:     DispatchAgent,
		ApprovalID: r.ApprovalID,
		RV:         r.RV,
		EventID:    r.EventID,
	}
}

// SpawnDeclaration spawn 接线声明（approved 派单行的 spawn 键——接线声明词形，非拉起事实）。
type SpawnDeclaration struct {
	Declared   bool   `json:"declared"`   // approved 批准投影=接线声明在册
	Wired      bool   `json:"wired"`      // 恒 false=真实拉起接线 deferred（红线 a 机械锚）
	Target     string `json:"target"`     // 拉起目标锚（apply-workflow——§8 专用 agent）
	ApprovalID string `json:"approvalId"` // 批准投影键（判据回链 approvals.records）
	RV         int64  `json:"rv"`         // 骨架落库后版本（实现侧 diff 基线随行）
	EventID    string `json:"eventId"`    // DocApplied 回执（来源链可回指）
}

// SpawnOf 批准投影 → spawn 声明（导出形——/apiserver 派单行集消费；ApprovalSpawnDeclaration
// 同一投影零硬编码）。
func SpawnOf(r ApprovalRecord) SpawnDeclaration { return ApprovalSpawnDeclaration(r) }

// r2d 投影行 → 派单行形（matchingKeyOf 键源复用——pending 投影键=派单行键同构）。
func r2d(r ApprovalRecord) DispatchRecord {
	return DispatchRecord{Kind: r.Kind, Workflow: r.Workflow, Scope: r.Scope,
		Connector: r.Connector, Proposal: r.Proposal}
}

// approvalIDOf 审批对象键（approval_log approval_id 同词汇：draft:<kind>/<scope>/<name>
// ——对象族待批键；proposal 派单=kind/proposal 键）。与 matchingKeyOf 同源（同键=投影行与
// apply_log 判定行 obj 精确对上）。
func approvalIDOf(rec DispatchRecord) string {
	return matchingKeyOf(rec)
}

// ───────────────── console 通知中心收件箱（§1.31——console-adoption.md 规则 6 页面兜底） ─────────────────

// InboxRow 通知中心收件箱一行（机读判定分类自 cored 读面现读——console 不自建持久状态，
// 渲染数据无双源红线 2；四类词=NeedsHuman|ApprovalNeeded|Proposal|DocReview）。
type InboxRow struct {
	Class   string `json:"class"`   // inboxNeedsHuman|inboxApprovalNeeded|inboxProposal|inboxDocReview（inboxClass* 词）
	Kind    string `json:"kind"`    // 深链对象 kind（events 类=空——详情路由按 eventId）
	Name    string `json:"name"`    // 对象名
	Scope   string `json:"scope"`   // 对象 scope
	Digest  string `json:"digest"`  // 摘要（标题/事件类型+involved——页面展示词形）
	EventID string `json:"eventId"` // 事件锚（events 类）
	At      string `json:"at"`      // 归档/更新时刻（RFC3339，现读面原值）
	Link    string `json:"link"`    // console 深链路由（#/…——规则 6 页面兜底入口；操作面=编辑器/审批/提案详情）
}

// Inbox 通知中心收件箱投影（四类汇总+计数+深链；未配置通知通道=仅页面展示，规则 6）。
type Inbox struct {
	Count    int            `json:"count"`
	Classes  map[string]int `json:"classes"`
	Items    []InboxRow     `json:"items"`
	Watchdog map[string]any `json:"watchdog,omitempty"`
}

// inbox 词汇常量（四类判定分类——console 判别词同源；导出形=装配层挂起行投影消费）。
const (
	inboxNeedsHuman     = "NeedsHuman"
	inboxApprovalNeeded = "ApprovalNeeded"
	inboxProposal       = "Proposal"
	inboxDocReview      = "DocReview"
)

// InboxClassApprovalNeeded ApprovalNeeded 类导出常量（/v1/inbox 装配层挂起行投影消费——
// 同一判定词零硬编码）。
const InboxClassApprovalNeeded = inboxApprovalNeeded

// BuildInbox 收件箱投影组装（纯函数——装配层喂行集+watchdog 投影；读面缺席类=零行，
// 缺口=空态显式不冒充）。行序：NeedsHuman→ApprovalNeeded→Proposal→DocReview（人注意优先），
// 同类内按 At 降序。
func BuildInbox(events []opslog.EventLog, drafts []InboxDraftRow, proposals []InboxProposalRow, watchdog map[string]any) *Inbox {
	rows := make([]InboxRow, 0, len(events)+len(drafts)+len(proposals))
	for _, e := range events {
		switch e.Type {
		case "NeedsHuman", "Escalate":
			rows = append(rows, InboxRow{Class: inboxNeedsHuman, Digest: e.Type + " · " + e.Involved,
				EventID: e.EventID, At: e.Ts.UTC().Format(time.RFC3339Nano),
				Link: "#/event/" + e.EventID})
		case "ApprovalNeeded":
			rows = append(rows, InboxRow{Class: inboxApprovalNeeded, Kind: involvedKind(e.Involved),
				Name: involvedName(e.Involved), Scope: involvedScope(e.Involved),
				Digest: e.Type + " · " + e.Involved, EventID: e.EventID,
				At: e.Ts.UTC().Format(time.RFC3339Nano), Link: inboxLinkOf(e.Involved, e.EventID)})
		}
	}
	for _, d := range drafts {
		rows = append(rows, InboxRow{Class: inboxDocReview, Kind: d.Kind, Name: d.Name, Scope: d.Scope,
			Digest: d.FieldManager + " 稿待审（" + d.Kind + "/" + d.Name + "）", At: d.UpdatedAt,
			Link: inboxDocLinkOf(d.Kind, d.Scope, d.Name)})
	}
	for _, p := range proposals {
		if p.Status != "submitted" {
			continue // 待批提案=status=submitted（approved/rejected/landed 非待办——状态机词汇闭合）
		}
		rows = append(rows, InboxRow{Class: inboxProposal, Kind: "Proposal", Name: p.Name, Scope: p.Scope,
			Digest: "提案待批（" + p.Title + "）", At: p.UpdatedAt, Link: "#/proposal/" + p.Scope + "/" + p.Name})
	}
	// 稳定排序：类序→At 降序（同刻按 Digest 字典序收尾——投影确定性）。
	SortInboxRows(rows)
	classes := map[string]int{}
	for _, r := range rows {
		classes[r.Class]++
	}
	return &Inbox{Count: len(rows), Classes: classes, Items: rows, Watchdog: watchdog}
}

// involvedKind/involvedScope/involvedName involvedRef（kind/scope/name 三段）宽松拆解。
func involvedKind(involved string) string {
	k, _, _, ok := involvedRef(involved)
	if !ok {
		return ""
	}
	return k
}
func involvedScope(involved string) string {
	_, s, _, ok := involvedRef(involved)
	if !ok {
		return ""
	}
	return s
}
func involvedName(involved string) string {
	_, _, n, ok := involvedRef(involved)
	if !ok {
		return ""
	}
	return n
}

// inboxLinkOf ApprovalNeeded involved → 操作面深链（对象族=编辑器屏；不可解析=事件详情兜底）。
func inboxLinkOf(involved, eventID string) string {
	k, s, n, ok := involvedRef(involved)
	if !ok || k == "" || n == "" {
		return "#/event/" + eventID
	}
	return "#/doc/" + k + "/" + s + "/" + n
}

// inboxDocLinkOf 稿对象 → 操作面深链（DesignDoc=设计文档工作台；对象族=编辑器屏）。
func inboxDocLinkOf(kind, scope, name string) string {
	if kind == "DesignDoc" {
		return "#/designdoc/" + scope + "/" + name
	}
	return "#/doc/" + kind + "/" + scope + "/" + name
}

// SortInboxRows 收件箱行稳定排序（导出形——BuildInbox 内序与装配层挂起行合并后复排同一
// 投影：类序→At 降序（同刻按 Digest 字典序收尾）——行集与计数同源零硬编码）。
func SortInboxRows(rows []InboxRow) {
	classRank := map[string]int{inboxNeedsHuman: 0, inboxApprovalNeeded: 1, inboxProposal: 2, inboxDocReview: 3}
	sort.SliceStable(rows, func(i, j int) bool {
		if classRank[rows[i].Class] != classRank[rows[j].Class] {
			return classRank[rows[i].Class] < classRank[rows[j].Class]
		}
		if rows[i].At != rows[j].At {
			return rows[i].At > rows[j].At
		}
		return rows[i].Digest < rows[j].Digest
	})
}

// InboxDraftRow / InboxProposalRow 收件箱组装输入行（装配层从读面行集投影——宽松键，零宿主依赖）。
type InboxDraftRow struct {
	Kind, Scope, Name, FieldManager, UpdatedAt string
}

type InboxProposalRow struct {
	Name, Scope, Title, Status, UpdatedAt string
}
