// run.go —— 对账 run 投影面（引擎主体第三段 v0.15——真实拉起第一子段：workflow 触发面接线）。
// 事件语义锚（concepts/event.md）：「Event 只触发，不修改：收到事件的反应永远是触发一次
// Workflow 对账」——本文件把「触发一次对账」落为**进程内可查的 run 留痕**：route 目标执行面
// =workflow 且装配层已登记接线（TriggerHub.WireExecutor(ExecWorkflow)）时，触发骨架把该次
// 决策推进为一条 RunRecord（一次对账的起点投影），查询面见 GET /v1/workflows/runs。
//
// 相位推进器首段（v0.23——迁移判定=声明序列单调前进，workflows 规则 12）：run 受理即启动
// 信号——引擎在受理路径上执行**单步迁移判定**：builtin run 从声明首相位沿声明相位序列单调
// 前进一步（Idle→Analyzing），status 词形翻转 triggered→running（触发已受理→首步进行中）。
// v0.25 显式迁移表首段：迁移判定从「序列前进」泛化为**声明表驱动**（workflows 规则 14）——
// builtin 表=声明序列物化相邻行（from=seq[i] on=advance to=seq[i+1]，序列前进=表的特例）
// +挂起-恢复审批边=表行先例；判定器=纯查表（from+on 命中→to，未命中=零迁移），受理首步/
// 批准恢复/驳回终态三路径共用同一查表落点（行为面不变，序列语义零翻转）；object 源 run
// 自 v0.28 起随对象自声明 spec（spec.phases 相位序列+spec.transitions 迁移表，装配层注入
// 解析面）参与同一查表判定——相位投影即点亮判定，判定器零改动；未声明 phases 的对象=零
// 投影，status 恒 triggered（零投影口径不稀释）；已居序列尾（Failed 等终点）不再前进。
// 受理与首步推进同拍完成（同步路径——真实异步派发 deferred，投影面只见推进后形态）。
// 查询面声明表可见键：run 载荷新增 transitions（物化表行投影 from/on/to——新载荷键出现
// =代际判别器）。
//
// 挂起-恢复首段（v0.24——闸锚挂起判定+审批信号迁移，workflows 规则 13/event.md 规则 11）：
// 闸锚=声明相位序列中的 AwaitingApproval 锚（人批闸；闸锚必须在声明序列内，无闸锚的序列
// 无挂起面）。三层判定：(1) 闸锚挂起——run 迁移判定落点=闸锚 → status 翻 suspended（相位机
// 等待外部审批信号，非终态非执行完毕）；(2) 审批提升——ApprovalNeeded（involved=Workflow
// 对象引用）到达 → 该锚最新 run 相位落闸锚挂起（「AwaitingApproval 相位插入」的机读落点）；
// (3) 批准恢复/驳回终态——ApprovalGranted → 闸锚沿声明序列续行（status 翻回 running——
// 恢复=序列续行，原相位不回跳；细粒度断点续跑靠 SessionLog，载体未落地挂账）；
// ApprovalRejected → 终态（相位=声明终点锚 Failed，status=failed）。审批信号按 involved
// 的 workflow 锚寻径、作用于该锚最新 run，**是相位机的迁移事件不是新对账触发**——不另起
// run（hub 侧经可选 ResumeSink 能力接口转发，非 dispatch 派发路径；同一条总线订阅主干零
// 旁路）。零动作防御面：object 源无从迁移、终态不迁移、非挂起态无可恢复、锚无 run/无闸锚
// 无可作用、既有 ApprovalNeeded 权限拒绝产源（involved=Role/Group）不命中 run 域。
//
// 诚实投影边界不变：run 出现 ≠ 对账执行完毕——status=running 只证明引擎已把 run 推进到
// 声明序列第二相位（Analyzing 进行中），后续相位推进/刹车/冷却/minLevel 仍属引擎主体余量
// （cored-bridge.md §2.3）。
//
// 锚解析（resolveRunAnchor，纯映射，与 execFor 家族收编同构）：
//   - 以 workflows 为前缀的直接引用 → 按名解析：builtin 注册表命中=内置 spec（相位序列可投影）；
//     未命中=平台对象引用（Workflow 是平台对象，spec 随文档面——run 只留锚，不吞并不缺口化）
//   - 对账锚（p 加 -workflow 后缀，如 owner-workflow）→ 对账循环锚 core-workflow
//     （对账=系统级循环本体，concepts/workflow.md「Controller/Reconciler（触发+对账+刹车）」
//     对照；owner 维度由留痕 Owner/Target 字段承载，不丢失）
//   - 其余 → 空锚（防御性：sink 只被 workflow 执行面调用；万一到达原样可见）
package workflow

import (
	"strings"
	"sync"
	"time"
)

// run 状态词汇（v0.15 起点 triggered；v0.23 相位推进器首段扩枚举 running；v0.24 挂起-恢复
// 首段扩枚举 suspended/failed——闸锚挂起与终态迁移的 status 词形）。
const (
	RunTriggered = "triggered" // 触发已受理（相位未投影 run 的稳态——零投影口径不稀释）
	RunRunning   = "running"   // 引擎已把 run 从声明首相位沿序列推进一步（首步进行中；非执行完毕）
	RunSuspended = "suspended" // 闸锚挂起（相位机等待外部审批信号——非终态非执行完毕）
	RunFailed    = "failed"    // 驳回终态（相位=声明终点锚；迁移终点，非引擎失败语义）
)

// 审批信号族（v0.24 挂起-恢复首段，event.md 规则 11/workflows 规则 13）：类型注册表
// 闸门前置（未注册类型进不了总线主干——发布面 404 语义），此处词汇=注册表已声明三型的
// run 域作用语义。信号按 involved=Workflow 对象引用寻径（kind/scope/name 首段 kind=Workflow）。
const (
	SignalSuspend = "ApprovalNeeded"   // 审批提升：相位落闸锚挂起（既有产源 involved=Role/Group 不命中 run 域）
	SignalResume  = "ApprovalGranted"  // 批准：闸锚沿声明序列续行（恢复）
	SignalReject  = "ApprovalRejected" // 驳回：闸锚→声明终点锚（终态迁移）
)

// approvalSignal 事件类型 → run 域信号语义（非审批信号族=零作用）。
func approvalSignal(typ string) (string, bool) {
	switch typ {
	case SignalSuspend:
		return "suspend", true
	case SignalResume:
		return "resume", true
	case SignalReject:
		return "reject", true
	}
	return "", false
}

// isApprovalSignalType 事件类型是否审批信号族（hub 转发闸——廉价前置判定，信号语义
// 细分与 involved 寻径在 ResumeSignal）。
func isApprovalSignalType(typ string) bool {
	_, ok := approvalSignal(typ)
	return ok
}

// workflowRefOf involved 投影 → workflow 锚名（involved=Workflow 对象引用即锚：
// kind/scope/name 三段首段 kind=Workflow 且 name 非空；其余=非锚，零作用）。
func workflowRefOf(involved string) (string, bool) {
	segs := strings.Split(involved, "/")
	if len(segs) >= 3 && segs[0] == "Workflow" && segs[2] != "" {
		return segs[2], true
	}
	return "", false
}

// spec 来源词汇（run 锚解析的声明来源投影）。
const (
	SpecBuiltin = "builtin" // 内置注册表命中（相位序列随代码声明，可投影）
	SpecObject  = "object"  // 平台对象引用（Workflow 对象，spec 随文档面；run 只留锚）
)

// TransitionRow 迁移表行投影（v0.25 显式迁移表首段——查询面载荷词汇：from/on/to 小写键，
// 表行可回指声明面；物化口径见 MaterializeTransitions）。
type TransitionRow struct {
	From string `json:"from"` // 迁移源相位
	On   string `json:"on"`   // 迁移事件名（advance=序列推进；ApprovalGranted/ApprovalRejected=审批边）
	To   string `json:"to"`   // 迁移目标相位
}

// transitionRows 声明表 → 投影行（声明面 → 留痕面；顺序保留）。
func transitionRows(ts []Transition) []TransitionRow {
	out := make([]TransitionRow, len(ts))
	for i, t := range ts {
		out[i] = TransitionRow{From: string(t.From), On: t.On, To: string(t.To)}
	}
	return out
}

// ReconcileWorkflow 对账锚的循环本体锚（owner 对账 workflow → core-workflow）。
// 依据：concepts/workflow.md——core-workflow=系统级循环（触发→整理→处理→回归），
// 即「收到事件的反应=触发一次 Workflow 对账」的那个对账循环。
const ReconcileWorkflow = "core-workflow"

// RunRecord 一条对账 run 投影（触发面→workflow 域的执行留痕；观测面非账本，
// 账本=event_log 归档，ops-log.md §2）。
type RunRecord struct {
	Run         int64           // run 序号（本 sink 进程内单调递增，1 起）
	Workflow    string          // workflow 锚（对账锚解析=core-workflow；直接引用=按名）
	SpecSource  string          // 锚的声明来源（builtin=内置注册表命中；object=平台对象引用；空=未解析）
	Phases      []string        // 声明相位序列（builtin 命中或 object 源声明 spec.phases 时投影；未声明=空）
	Transitions []TransitionRow // 声明迁移表投影（v0.25 显式迁移表首段：builtin=物化相邻行+审批边；v0.27/v0.28 对象自声明 spec 解析：object 源=文档库 spec.transitions 投影）
	Phase       string          // 当前相位投影（v0.23 迁移判定首段：受理即沿声明序列前进一步=第二相位；v0.24 审批提升落闸锚/批准续行/驳回落终点锚；v0.25 判定器改查表；v0.28 object 源随声明 phases 参与同一查表判定）
	Status      string          // run 状态（v0.24 词汇四值：triggered 受理/running 首步进行中/suspended 闸锚挂起/failed 驳回终态；相位未投影 run 恒 triggered）
	TriggerSeq  int64           // 来源触发留痕 seq（可回指 GET /v1/triggers records）
	EventID     string          // 与发布方回执同源（apply 响应 event.id / event_log.event_id）
	Type        string          // 事件类型
	DedupKey    string          // 与发布方回执同源
	Route       string          // 命中行 route 原文（owner-workflow / workflows/<name> 等锚的出处）
	Owner       string          // owner 解析值（对账锚的 owner 维度在此承载）
	Target      string          // 执行面锚点投影（=ExecutorBinding.Target）
	StartedAt   time.Time       // run 起始时刻（收到留痕同刻）
}

// RunState 查询快照（RunLog.State() 返回值；apiserver GET /v1/workflows/runs 载荷来源）。
type RunState struct {
	Started  int64       // 累计 start（含已被环挤出的）
	Dropped  int64       // 环溢出丢弃计数（可观测）
	Advances int64       // 累计异步推进的 run·步计数（v0.26 异步推进首段——新载荷键=代际判别器）
	Runs     []RunRecord // 环内 run 投影（旧→新）
}

// RunSink workflow 触发面的 run 投影落点（引擎主体第三段 v0.15）。
// 解耦：hub 只把「route 目标执行面=workflow ∧ 已接线」的留痕推进给 sink，不感知 run 内部
// （相位态机/刹车/冷却）；扩展性：真实相位态机落地换实现不动 hub。实现须**非阻塞**（hub 在
// 订阅路径上同步调用）且**不得回调 hub**（进程内投影面；真实异步派发落地时由实现自行解耦）。
type RunSink interface {
	StartRun(rec TriggerRecord, b ExecutorBinding)
}

// ResumeSink 审批信号面（挂起-恢复首段 v0.24）——RunSink 的**可选能力接口**：实现侧
// 声明自己承接审批信号的 run 域迁移判定（workflows 规则 13），hub 按能力断言转发（类型
// 断言失败=不转发，未声明能力的 sink 不受审批信号作用——解耦：能力增量不破既有 sink 契约）。
// 转发时机=hub.record 锁内（与留痕/run 同序）；转发前提=事件类型∈审批信号族（类型注册
// 闸门在总线/发布面前置——未注册类型进不了主干），involved 寻径由实现侧判定。
type ResumeSink interface {
	ResumeSignal(rec TriggerRecord)
}

// RunOptions run 投影环可调项（零值可用）。
type RunOptions struct {
	// RingSize run 环容量（缺省 256）：溢出丢最旧、dropped 计数可见。
	RingSize int
	// Specs 对象 spec 解析面（对象声明表解析 v0.27+v0.28 相位投影接线，workflows 规则 16
	// ——可选能力，缺省 nil=既有行为面不变：object 源零投影零迁移）。注入后 object 源 run
	// 的相位序列与声明表按对象 spec 投影（builtin 物化口径之外的第二声明来源——同一载荷
	// 键）；相位投影即点亮既有表驱动判定（受理首步/审批信号/异步推进三路径自然生效，判定器
	// 零改动）；未声明 phases 的对象=零投影既有口径不稀释。
	Specs SpecResolver
}

// RunLog run 投影的进程内落点（RunSink 缺省实现；并发安全）。
type RunLog struct {
	opts RunOptions

	mu       sync.Mutex
	seq      int64
	ring     []RunRecord
	started  int64
	dropped  int64
	advances int64 // 累计异步推进 run·步数（v0.26 异步推进入口 Advance 累加；advance.go）
}

// NewRunLog 构建 run 投影落点。
func NewRunLog(opts *RunOptions) *RunLog {
	r := &RunLog{}
	if opts != nil {
		r.opts = *opts
	}
	if r.opts.RingSize <= 0 {
		r.opts.RingSize = 256
	}
	return r
}

// StartRun 记一条 run 投影（环满丢最旧，dropped 计数可见）。锚解析见 resolveRunAnchor。
func (r *RunLog) StartRun(rec TriggerRecord, b ExecutorBinding) {
	wf, source, phases, trans := resolveRunAnchor(rec.Route)
	// 对象 spec 解析接线（v0.27 表投影首段+v0.28 相位投影，workflows 规则 16）：object
	// 源且装配层注入了解析面 → 对象自声明 spec 一次解析两面投影（相位序列+声明表，声明→
	// 留痕面，顺序保留）。解析在锁外（文档库读面文件 IO 不持 r.mu——hub 同步调用路径不因
	// 投影阻塞）。相位投影即点亮判定：phases 非空后，下方受理首步判定分支对 object 源自然
	// 生效（判定器零改动——stepByTable 纯查表落点通用，推进器已完全表驱动，接表即得全量
	// 推进语义）；未声明 phases 的对象=零投影（phases 恒空），三条判定路径零动作防御面
	// 原样生效，既有行为面零翻转。
	if source == SpecObject && r.opts.Specs != nil {
		view := r.opts.Specs.ResolveSpec(wf)
		if len(view.Transitions) > 0 {
			trans = view.Transitions
		}
		if len(view.Phases) > 0 {
			phases = view.Phases
		}
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	r.seq++
	r.started++
	if excess := len(r.ring) + 1 - r.opts.RingSize; excess > 0 {
		r.ring = r.ring[excess:]
		r.dropped += int64(excess)
	}
	phase := ""
	status := RunTriggered
	if len(phases) > 0 {
		phase = string(phases[0]) // 受理投影=声明首相位（起点词先落，迁移判定在其上判）
		// 相位推进器首段（v0.23）+显式迁移表首段（v0.25）：受理即启动信号——引擎单步迁移
		// 判定，按声明表查推进行（from=首相位 on=advance）并把状态词翻转为 running
		//（stepByTable 纯查表，判定口径见包注释——序列前进=表的特例）。
		if next, ok := stepByTable(trans, phase, EdgeAdvance); ok {
			phase, status = next, RunRunning
		}
		// 挂起-恢复首段（v0.24）：迁移判定落点=闸锚 → 挂起（相位机等待外部审批信号，
		// workflows 规则 13 第一层判定；builtin 序列受理首步不落闸锚=防御性通用规则，
		// 单测以闸锚邻位序列覆盖）。
		if status == RunRunning && Phase(phase) == AwaitingApproval {
			status = RunSuspended
		}
	}
	if phases == nil {
		phases = []string{} // 相位未投影（未声明 phases 的 object 源）：载荷形状归一为空数组（null=缺席语义，[]=声明为空——消费方可 len）
	}
	if trans == nil {
		trans = []TransitionRow{} // object 源不投影声明表：载荷形状归一同口径（新键可 len）
	}
	r.ring = append(r.ring, RunRecord{
		Run: r.seq, Workflow: wf, SpecSource: source, Phases: phases, Transitions: trans,
		Phase: phase, Status: status,
		TriggerSeq: rec.Seq, EventID: rec.EventID, Type: rec.Type, DedupKey: rec.DedupKey,
		Route: rec.Route, Owner: rec.Owner, Target: b.Target, StartedAt: rec.ReceivedAt,
	})
}

// ResumeSignal 审批信号的 run 域迁移判定（挂起-恢复首段 v0.24，workflows 规则 13）——
// v0.25 显式迁移表首段起批准/驳回迁移判定走声明表查表（stepByTable，规则 14：批准/驳回
// 边=表行先例，物化随 spec 携带在 run 投影上）：
// 审批信号族事件按 involved=Workflow 对象引用寻径，作用于该锚**最新 run**（环内该锚
// 最大 Run 序号=现势 run）。判定三层：审批提升=相位落闸锚挂起（声明无闸锚=零动作）；
// 批准=查表命中审批边（闸锚 on=ApprovalGranted）续行（恢复，status 翻回 running）；
// 驳回=查表命中驳回边（闸锚 on=ApprovalRejected → 终点锚 Failed，非相邻行）终态。
// 零动作防御面：非审批信号族/involved 非 Workflow 引用/相位未投影（未声明 phases 的
// object 源无从迁移——v0.28 起声明了 phases 的对象 run 参与审批信号判定）/终态/
// 非挂起态无可恢复/锚无 run——一律原样保留（事件仍是留痕，迁移判定不命中不冒充动作）。
// 幂等同拍：审批提升对已挂起 run 重复到达=零动作（已是闸锚挂起态）。
func (r *RunLog) ResumeSignal(rec TriggerRecord) {
	action, ok := approvalSignal(rec.Type)
	if !ok {
		return
	}
	name, ok := workflowRefOf(rec.Involved)
	if !ok {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := len(r.ring) - 1; i >= 0; i-- {
		run := &r.ring[i]
		if run.Workflow != name || len(run.Phases) == 0 {
			continue // 锚不符 / 相位未投影（未声明 phases 的 object 源）无从迁移
		}
		switch action {
		case "suspend": // 审批提升：相位落闸锚挂起（终态/已挂起=零动作；声明无闸锚=零动作）
			if run.Status == RunFailed || run.Status == RunSuspended ||
				!containsPhase(run.Phases, string(AwaitingApproval)) {
				return
			}
			run.Phase, run.Status = string(AwaitingApproval), RunSuspended
		case "resume": // 批准：查表命中审批边（闸锚 on=ApprovalGranted）沿表续行（非闸锚挂起态=零动作；表无批准边=零动作防御）
			if run.Status != RunSuspended || Phase(run.Phase) != AwaitingApproval {
				return
			}
			next, ok := stepByTable(run.Transitions, run.Phase, SignalResume)
			if !ok {
				return
			}
			run.Phase = next
			run.Status = RunRunning // 闸锚续行落点非闸锚（序列无连续闸锚防御：若落闸锚再挂起）
			if Phase(next) == AwaitingApproval {
				run.Status = RunSuspended
			}
		case "reject": // 驳回：查表命中驳回边（闸锚 on=ApprovalRejected → 终点锚 Failed，表的首处非相邻行）终态（非挂起态/表无驳回边=零动作）
			if run.Status != RunSuspended {
				return
			}
			to, ok := stepByTable(run.Transitions, run.Phase, SignalReject)
			if !ok {
				return
			}
			run.Phase, run.Status = to, RunFailed
		}
		return // 只作用现势 run（最新命中即返回）
	}
}

// containsPhase 序列内判定（纯函数）。
func containsPhase(phases []string, p string) bool {
	for _, x := range phases {
		if x == p {
			return true
		}
	}
	return false
}

// State 查询快照（拷贝返回，调用方可安全持有；runs 旧→新）。
func (r *RunLog) State() RunState {
	r.mu.Lock()
	defer r.mu.Unlock()
	runs := make([]RunRecord, len(r.ring))
	copy(runs, r.ring)
	return RunState{Started: r.started, Dropped: r.dropped, Advances: r.advances, Runs: runs}
}

// stepByTable 迁移判定（v0.25 显式迁移表首段，workflows 规则 14——纯查表）：声明表内命中
// from+on 行 → to；未命中=零迁移（ok=false，原相位保留）。判定器从「序列前进」泛化为
// 「声明表驱动」：受理首步（on=advance）与审批边（on=ApprovalGranted/ApprovalRejected）
// 共用同一查表落点——序列前进=表的特例（表=序列物化相邻行，MaterializeTransitions），
// 非相邻迁移按表命中（驳回边=表的首处非相邻行，机械判别器）。行为面不变：既有序列语义
// 零翻转（物化相邻行⟷序列前进等价）。
func stepByTable(table []TransitionRow, from, on string) (string, bool) {
	for _, t := range table {
		if t.From == from && t.On == on {
			return t.To, true
		}
	}
	return from, false
}

// resolveRunAnchor 命中行 route 原文 → run 投影的 workflow 锚（纯映射；引擎主体第三段 v0.15）。
// 口径见包注释；直接引用按 builtin 注册表解析，未命中=平台对象引用（SpecObject，锚保留、
// 相位/表皆零投影——对象 spec 随文档面，投影由装配层注入的解析面承接，v0.27 表投影+v0.28
// 相位投影）。
func resolveRunAnchor(routeTarget string) (wf, source string, phases []string, trans []TransitionRow) {
	if name, ok := strings.CutPrefix(routeTarget, "workflows/"); ok && name != "" {
		if IsBuiltin(name) {
			spec := Builtin(name)
			return spec.Name, SpecBuiltin, phaseNames(spec.Phases), transitionRows(spec.Transitions)
		}
		return name, SpecObject, nil, nil // 平台对象引用：锚可见，spec 随文档面
	}
	if _, ok := strings.CutSuffix(routeTarget, "-workflow"); ok && routeTarget != "-workflow" {
		spec := Builtin(ReconcileWorkflow) // 对账锚 → 对账循环锚 core-workflow（owner 维度在留痕）
		return spec.Name, SpecBuiltin, phaseNames(spec.Phases), transitionRows(spec.Transitions)
	}
	return "", "", nil, nil // 防御性：非 workflow 家族原文原样可见（空锚）
}

// phaseNames 相位序列投影为字符串切片（声明面 → 留痕面；顺序保留）。
func phaseNames(ps []Phase) []string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = string(p)
	}
	return out
}

// 编译期断言：RunLog 满足 RunSink 与 ResumeSink（审批信号能力，v0.24）。
var (
	_ RunSink    = (*RunLog)(nil)
	_ ResumeSink = (*RunLog)(nil)
)
