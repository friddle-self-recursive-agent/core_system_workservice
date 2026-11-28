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

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/runtimestore"
)

// run 状态词汇（v0.15 起点 triggered；v0.23 相位推进器首段扩枚举 running；v0.24 挂起-恢复
// 首段扩枚举 suspended/failed——闸锚挂起与终态迁移的 status 词形；2026-11-26 F-04 扩枚举
// converged——正向序列末步收束态，workflow.md §1 单次 run 状态闭合集第六值）。
const (
	RunTriggered = "triggered" // 触发已受理（相位未投影 run 的稳态——零投影口径不稀释）
	RunRunning   = "running"   // 引擎已把 run 从声明首相位沿序列推进一步（首步进行中；非执行完毕）
	RunSuspended = "suspended" // 闸锚挂起（相位机等待外部审批信号——非终态非执行完毕）
	RunFailed    = "failed"    // 驳回终态（相位=侧轨终点锚 Failed；仅经 ApprovalRejected 边可达）
	RunConverged = "converged" // 正常收束（相位停驻正向序列末步——Reporting 完成即收束；非引擎失败语义）
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

// InvolvedWorkflowRef involved 投影 → workflow 锚名（导出面：装配层 service 桥寻径同一判定
// ——workflowRefOf 单点真相导出，无双源；serviceexec.go 完成信号面消费）。
func InvolvedWorkflowRef(involved string) (string, bool) {
	return workflowRefOf(involved)
}

// spec 来源词汇（run 锚解析的声明来源投影）。
const (
	SpecBuiltin = "builtin" // 内置注册表命中（相位序列随代码声明，可投影）
	SpecObject  = "object"  // 平台对象引用（Workflow 对象，spec 随文档面；run 只留锚）
)

// SubRunRef 子 run 投影（fanout 扇出——状态随子 run 终态回填；子 run 被环挤出后本行仍是
// join 判据的唯一权威，不依赖回查子 run 投影）。
type SubRunRef struct {
	Run     int64  `json:"run"`               // 子 run 序号
	ItemKey string `json:"itemKey,omitempty"` // 执行计划项键（幂等 dedup 锚）
	Status  string `json:"status"`            // spawned|converged|failed
	Session string `json:"session,omitempty"` // 子 run 会话锚（若回填）
}

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

// HistoryRow 运行轨迹一行（F-04——真实迁移留痕：静态表 transitions 的运行期对应行；
// UI 按本行集标绿已走过节点与时间线，静态表不再当时间线展示）。
type HistoryRow struct {
	From      string `json:"from"`                // 迁移源相位
	To        string `json:"to"`                  // 迁移目标相位
	On        string `json:"on"`                  // 迁移事件（advance/ApprovalGranted/ApprovalRejected/converged）
	At        string `json:"at"`                  // 迁移时刻（RFC3339Nano）
	SessionID string `json:"sessionId,omitempty"` // 迁移时刻 run 会话锚（空=未回填——诚实缺席，不编造节点会话）
	Note      string `json:"note,omitempty"`      // 迁移注记（2026-11-28 connectops 执行面——op 结果摘要/显式错误；新载荷键=代际判别器）
}

// ReconcileWorkflow 对账锚的循环本体锚（owner 对账 workflow → core-workflow）。
// 依据：concepts/workflow.md——core-workflow=系统级循环（触发→整理→处理→回归），
// 即「收到事件的反应=触发一次 Workflow 对账」的那个对账循环。
const ReconcileWorkflow = "core-workflow"

// RunRecord 一条对账 run 投影（触发面→workflow 域的执行留痕；观测面非账本，
// 账本=event_log 归档，ops-log.md §2）。
type RunRecord struct {
	Run            int64           // run 序号（本 sink 进程内单调递增，1 起）
	Workflow       string          // workflow 锚（对账锚解析=core-workflow；直接引用=按名）
	SpecSource     string          // 锚的声明来源（builtin=内置注册表命中；object=平台对象引用；空=未解析）
	Phases         []string        // 声明相位序列（builtin 命中或 object 源声明 spec.phases 时投影；未声明=空）
	Transitions    []TransitionRow // 声明迁移表投影（v0.25 显式迁移表首段：builtin=物化相邻行+审批边；v0.27/v0.28 对象自声明 spec 解析：object 源=文档库 spec.transitions 投影）
	ApprovalMode   string          // 审批策略投影（2026-11-26「Agent 节点自动级别」："auto"=闸锚自动通过；空=human 缺省=闸锚挂起等人。新载荷键=代际判别器）
	ApprovalPolicy ApprovalPolicy  // 审批策略全投影（autoCriteria/maxRiskLevel/approvalDiff——2026-11-28 自审批真判消费；零值=未声明）
	Phase          string          // 当前相位投影（v0.23 迁移判定首段：受理即沿声明序列前进一步=第二相位；v0.24 审批提升落闸锚/批准续行/驳回落终点锚；v0.25 判定器改查表；v0.28 object 源随声明 phases 参与同一查表判定）
	Status         string          // run 状态（v0.24 词汇四值：triggered 受理/running 首步进行中/suspended 闸锚挂起/failed 驳回终态；相位未投影 run 恒 triggered）
	Brakes         BrakeView       // 生效刹车值投影（v0.32 刹车首段：声明→留痕面——spec.brakes 钳定后的每锚生效值；新载荷键=代际判别器）
	TriggerSeq     int64           // 来源触发留痕 seq（可回指 GET /v1/triggers records）
	EventID        string          // 与发布方回执同源（apply 响应 event.id / event_log.event_id）
	Type           string          // 事件类型
	DedupKey       string          // 与发布方回执同源
	Route          string          // 命中行 route 原文（owner-workflow / workflows/<name> 等锚的出处）
	Owner          string          // owner 解析值（对账锚的 owner 维度在此承载）
	Target         string          // 执行面锚点投影（=ExecutorBinding.Target）
	Involved       string          // 来源留痕 involved 投影（kind/scope/name——run 涉及对象链的第一跳；
	//                             v0.37 console 读面第二段新增载荷键=代际判别器：运行抽屉「涉及文档」
	//                             直接取本字段落对象链接，无前端拼凑）
	SessionID string // 会话锚（G-08：loop 会话关联——唤醒桥受理回执经 AttachSessionID 回填，
	//                         export.go LoopSession 读面消费；空=未回填不冒充已拉起）
	Verify *VerifyVerdict // Verify 相位硬 gate 判定投影（G-09/G-12：离开 Verifying 相位先过 gate——
	//                         判定随 run 留痕写穿透，恢复取最新词形；nil=未审。新载荷键=代际判别器）
	// History 运行期实际轨迹（2026-11-26 F-04——workflow.md 规则 14：静态迁移规则表
	// transitions 与运行轨迹 history 分离；UI 节点图按 history 标绿已走过的真实节点，
	// 不按数组下标盲标）。行=每次真实迁移追加（受理首步/推进/批准/驳回/收束同口径）。
	FanoutPlan string       `json:"fanoutPlan,omitempty"` // 执行计划文档锚（batch-judge 产物；观测面）
	ParentRun  int64        `json:"parentRun,omitempty"`  // 子 run 的父 run 序号（0=非子 run；2026-11-28 #6）
	SubRuns    []SubRunRef  `json:"subRuns,omitempty"`    // 父 run 扇出的子 run 行集（join 判据）
	History    []HistoryRow `json:"history,omitempty"`
	StartedAt  time.Time    `json:"startedAt"` // run 起始时刻（收到留痕同刻）
	// Nodes 节点声明投影快照（2026-11-27 service session 裁决——新载荷键=代际判别器）：
	// 受理时随对象 spec 一次解析随行（运行期零文件 IO——Advance 落点判定纯内存查投影；
	// 恢复面 NewRunLog 重解析回填）。builtin 源恒空（内置注册表无节点细节——service 面
	// 只对 object 源声明了富节点的 run 生效）。不落行（RunRow 无列——恢复重解析，声明面
	// 是权威）。
	Nodes []NodeView `json:"nodes,omitempty"`
	// SuspendWhy 挂起因由（2026-11-27 service session 裁决——新载荷键=代际判别器）：
	// approval=闸锚挂起（缺省词形=既有口径不标注）；service=service 执行 session 进行中
	// （ServiceSpawner 挂接时落点即挂起——serviceexec.go）。空=非挂起/未细分。
	SuspendWhy string `json:"suspendWhy,omitempty"` // service|inner-agent（2026-11-28 内置轻量 agent——低智商润滑剂）
}

// RunState 查询快照（RunLog.State() 返回值；apiserver GET /v1/workflows/runs 载荷来源）。
type RunState struct {
	Started      int64       // 累计 start（含已被环挤出的）
	Dropped      int64       // 环溢出丢弃计数（可观测）
	Advances     int64       // 累计异步推进的 run·步计数（v0.26 异步推进首段——新载荷键=代际判别器）
	Braked       int64       // 累计冷却门拦阻触发数（v0.32 刹车首段——新载荷键=代际判别器）
	Capped       int64       // 累计推进拍上限拦阻 run·步数（v0.32 刹车首段——新载荷键=代际判别器）
	Disabled     int64       // 累计 enabled=false 锚拦阻触发数（v0.58b enabled honoring——新载荷键=代际判别器）
	Gated        int64       // 累计 Verify 硬 gate 卡 Verify 拦阻 run·步数（G-09/G-12——新载荷键=代际判别器）
	SelfApproved int64       // 累计自审批判决消费数（2026-11-28 #4——新载荷键=代际判别器）
	Runs         []RunRecord // 环内 run 投影（旧→新）
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
	// Brakes 刹车闸挂接项（v0.32 刹车首段，workflows 规则 17——可选能力，缺省零值=
	// 零刹车既有行为面零翻转：触发全落 run、推进不限量）。注入开启后冷却门（受理路径）
	// 与推进拍变更上限（推进路径）按生效刹车值生效（解析面未注入=全锚强制默认）。
	Brakes BrakeOptions
	// EnabledRules 规则开关判定面（v0.58b 触发路径 enabled honoring——E3 余量段收口，
	// 可选能力，缺省 nil=零闸既有行为面零翻转）。注入后 enabled=false 的锚触发留痕仍在、
	// run 不落（disabled 计数可见）；判定缺席（未命中/解析失败）=放行（缺口可见不静默拦）。
	EnabledRules EnabledResolver
	// Store 运行态投影存储挂接（R-ENG §1 可选能力——nil=纯内存既有行为面零翻转）。挂接后
	// 写穿透（run 落环/推进迁移/审批信号迁移/会话回填/Verify 判定同步 upsert——run 是可演化
	// 投影，恢复取最新词形）+ 启动恢复（New 时装载历史行续接 seq/started 与环内最近 run；
	// 冷却窗键面 lastAccepted 不恢复=重启重置冷却窗——窗口语义可接受偏差，留痕仍在）。
	// 读面零变化。
	Store runtimestore.Store
	// VerifyGate Verify 相位硬 gate（G-09/G-12，R-TYPE-3——可选能力，nil=既有行为面零翻转）。
	// 挂接后 run 试图离开 Verifying 相位先过 gate（connectops-code-audit 审核；判定锁外执行
	// 不阻塞订阅主干）：OK=false 卡 Verify 不放行（gated 计数可见）+ OnVerifyBlocked 上报；
	// OK=true 放行且判定随 run 留痕。装配层（cored）恒挂接（修复非选项——验收纪律收口）。
	VerifyGate VerifyGate
	// OnVerifyBlocked 卡 Verify 上报钩子（G-09/G-12——nil=只卡不上报）。gate 判定 blocked 时
	// 在锁外调用（可发布 Escalate 事件——cored 走既有注册类型 root 通道）；同 run 每拍重审
	// 重复 blocked 时重复回调（总线 dedup 窗口承担风暴抑制；checker 幂等只读）。
	OnVerifyBlocked func(run RunRecord, v VerifyVerdict)
	// ServiceSpawns service 节点 session 拉起面（2026-11-27 service session 裁决——可选能力，
	// nil=既有行为面零翻转：service 节点照旧纯查表推进不挂起不拉起）。挂接后 run 推进落
	// unit=service 节点即挂起（suspendWhy=service）+ 拉起投影（锁内收集锁外派发——非阻塞
	// 契约）；完成信号经 ServiceSignal 续行。
	ServiceSpawns ServiceSpawner
	// DiffSummarizer 审批判决文档生成面（2026-11-28 用户裁定「审批裁定的时候走一个小内置
	// agent[调用 llm 即可]整理下 DocumentDiff。生成 Diff 文档。告诉我改了什么」——可选能力，
	// nil=既有行为面零翻转：人审挂起不生成判决文档）。挂接后 run 落 AwaitingApproval 且
	// 实际挂起等人（auto/agent 模式直续不触发）时锁外异步调用实现侧（读稿+现值→内置小
	// agent 调 LLM 整理→DocumentDiff 文档落库——ops_core_inner_agent 词形，实现归装配层）。
	DiffSummarizer DiffSummarizer
	// InnerAgent 内置轻量 agent 执行面（2026-11-28 用户裁定「doc-apply 的审批以及部分服务
	// 运行起来可以使用内置的 agent 来做润滑剂。即低智商要求的可以使用内置 agent」——可选
	// 能力，nil=既有行为面零翻转）。挂接后 run 落 executor=inner 的 agent/service 节点即挂起
	// （suspendWhy=inner-agent）+ 锁外派发 RunInner（引擎内直调 Provider 轻量 LLM——不拉
	// session）；实现完成调 InnerDone 续行（on=inner-resume 轨迹行）。
	InnerAgent InnerAgentRunner
	// ApprovalJudge 自审批真判面（2026-11-28 用户裁定「自审批未接线」修复 / workservice #4
	// ——可选能力，nil=旧词形零翻转：mode=agent 仍盲放行）。挂接后 mode=agent 落闸锚即
	// 挂起（suspendWhy=self-approve）+ 锁外派发 ops_core_inner_agent 真判决；判决回执
	// ApprovalVerdictDone 沿声明表续行/终态/回退人审。
	ApprovalJudge ApprovalJudge
	// OnApprovalVerdict 判决留痕上报钩子（nil=只落 run history；装配层可发布
	// ApprovalGranted/ApprovalRejected 事件进 event_log 审计账）。锁外调用。
	OnApprovalVerdict func(run RunRecord, v ApprovalVerdict)
	// FanoutLauncher 第四词形 unit=workflow 的扇出执行面（2026-11-28 #6——可选能力，
	// nil=零翻转：unit=workflow 节点照旧纯查表推进）。挂接后落扇出节点挂起+锁外读执行计划
	// 建子 run，父 run join 等全部子 run 终态后 fanout-complete 迁移到汇聚相位。
	FanoutLauncher FanoutLauncher
	// OnFanoutLaunched 扇出回执钩子（nil=只落 run history；装配层可发通知/审计）。锁外调用。
	OnFanoutLaunched func(run RunRecord, refs []SubRunRef)
	// ConnectOpsRunner connectops 节点 op 执行面（2026-11-28 用户裁定「做成workflow的一个节点。
	// 即gh-release包括git push/release[scheme调用]」——可选能力，nil=零翻转：connectops 节点
	// 照旧纯查表推进）。挂接后 run 推进落 unit=connectops 节点即挂起（suspendWhy=connectops）
	// + 锁外派发 RunOp（实现侧按 node.Connector 查 ops 注册表 Invoke）；完成回执经 OpsSignal
	// 续行（on=connectops-resume；失败 on=connectops-error 留挂起缺口可见）。
	ConnectOpsRunner ConnectOpsRunner
}

// SetVerifyGate 装配后挂接 Verify 硬 gate（cored 构造序：RunLog 先于 verify 配置就绪——
// 旧写法改 runOpts 不生效（opts 已拷入），本 setter 为唯一权威挂接点；nil=关断）。
func (r *RunLog) SetVerifyGate(g VerifyGate, onBlocked func(run RunRecord, v VerifyVerdict)) {
	r.mu.Lock()
	r.opts.VerifyGate = g
	r.opts.OnVerifyBlocked = onBlocked
	r.mu.Unlock()
}

// InnerAgentRunner 内置轻量 agent 契约（ops_core_inner_agent 词形——润滑剂：低智商环节
// 不拉 session 引擎内直调 LLM）。实现侧自管产物落库与幂等；err 非 nil=留挂起（续行靠重试）。
type InnerAgentRunner interface {
	RunInner(run RunRecord, node NodeView) error
}

// DiffSummarizer 审批判决文档生成契约（实现侧异步执行——引擎只通知不等待不重试；
// 幂等/防重归实现侧按 target rv 自管）。objKey=审批对象键 kind/scope/name。
type DiffSummarizer interface {
	OnApprovalSuspended(run RunRecord, objKey string)
}

// RunLog run 投影的进程内落点（RunSink 缺省实现；并发安全）。
type RunLog struct {
	opts RunOptions

	mu                  sync.Mutex
	seq                 int64
	ring                []RunRecord
	started             int64
	dropped             int64
	advances            int64                // 累计异步推进 run·步数（v0.26 异步推进入口 Advance 累加；advance.go）
	braked              int64                // 累计冷却门拦阻触发数（v0.32 刹车首段——新载荷键=代际判别器）
	capped              int64                // 累计推进拍上限拦阻 run·步数（v0.32 刹车首段——新载荷键=代际判别器）
	disabled            int64                // 累计 enabled=false 锚拦阻触发数（v0.58b enabled honoring——新载荷键=代际判别器）
	gated               int64                // 累计 Verify 硬 gate 卡 Verify 拦阻 run·步数（G-09/G-12——新载荷键=代际判别器）
	selfApproved        int64                // 累计自审批判决消费数（2026-11-28 #4——新载荷键=代际判别器）
	selfApproveInFlight map[int64]bool       // 自审批判决在途防重（同 run 重复挂起/重启不重复调 LLM）
	lastAccepted        map[string]time.Time // 冷却门键→最新受理时刻（刹车闸挂接时维护；v0.32）
}

// NewRunLog 构建 run 投影落点。挂有 Store 时先做启动恢复（R-ENG §1——装载历史行填环内
// 最近 RingSize 条、seq/started 续接为行数口径；冷却窗键面不恢复=重启重置，见 RunOptions.Store 注）。
func NewRunLog(opts *RunOptions) *RunLog {
	r := &RunLog{}
	if opts != nil {
		r.opts = *opts
	}
	if r.opts.RingSize <= 0 {
		r.opts.RingSize = 256
	}
	r.lastAccepted = map[string]time.Time{}
	r.selfApproveInFlight = map[int64]bool{}
	if r.opts.Store != nil {
		rows := loadRunRows(r.opts.Store)
		if len(rows) > 0 {
			runs := make([]RunRecord, len(rows))
			maxSeq := int64(0)
			for i, row := range rows {
				runs[i] = runFromRow(row)
				if row.Run > maxSeq {
					maxSeq = row.Run
				}
			}
			r.seq = maxSeq
			r.started = int64(len(rows))
			if excess := len(runs) - r.opts.RingSize; excess > 0 {
				runs = runs[excess:]
			}
			// 节点投影恢复回填（2026-11-27 service session 裁决）：object 源 run 的 Nodes 不落行
			// （声明面是权威）——恢复时重解析回填，挂起 run 的 service 落点判定照常纯内存。
			// 启动期文件 IO（非热路径）；解析缺席=零投影（诚实降级，缺口可见不编造）。
			if r.opts.Specs != nil {
				for i := range runs {
					if len(runs[i].Nodes) == 0 { // docs 命中即回填（builtin 遮蔽同口径——StartRun 节点投影注）
						runs[i].Nodes = r.opts.Specs.ResolveSpec(runs[i].Workflow).Nodes
					}
				}
			}
			r.ring = runs
		}
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
	approvalMode := "" // 空=human 缺省（审批策略声明走对象 spec）
	var approvalPolicy ApprovalPolicy
	var nodes []NodeView
	if r.opts.Specs != nil {
		view := r.opts.Specs.ResolveSpec(wf)
		if source == SpecObject {
			if len(view.Transitions) > 0 {
				trans = view.Transitions
			}
			if len(view.Phases) > 0 {
				phases = view.Phases
			}
		}
		// 节点投影随 run 快照（2026-11-27 service session 裁决——service 落点判定纯内存，
		// 运行期零文件 IO）。口径与 approvalMode 同款：docs 对象命中即投影（builtin 遮蔽下
		// docs 权威种子的富节点 unit:service 声明仍供拉起面——纯 builtin 无文档对象=零投影
		// 零翻转；文档词形与 builtin 相位词汇分歧时 serviceNodeOf 查不中=天然零动作防御）。
		nodes = view.Nodes
		approvalPolicy = view.Approval
		if view.ApprovalMode != "" {
			approvalMode = view.ApprovalMode // spec.approval.mode=auto|agent 投影（Agent 节点自动判断/自动级别）
		}
	}
	// 刹车首段（v0.32，workflows 规则 17）：每锚生效刹车值（表驱动——spec.brakes → 强制
	// 默认钳定 → 生效值；声明→留痕面投影不依赖挂接态）。冷却门只在刹车闸挂接时生效（未挂
	// 接=零刹车既有行为面零翻转——基线保护）：同锚同 involved 最新 run 受理时刻起 cooldown
	// 窗口内的后续触发不落 run（触发留痕仍在——braked 计数可见）；窗口外照落；跨 involved
	// 不互踩（防震荡语义粒度=同一对象）。cooldown 锚首段=run 受理时刻（rec.ReceivedAt）。
	// 判定在锁外算生效值（解析面文件 IO 不持 r.mu），闸判在锁内（braked/lastAccepted 并发安全）。
	// 规则开关闸（v0.58b enabled honoring——E3 余量段收口）：enabled=false 的锚=该规则
	// 不再触发（workflow.md §1 语义零新词形）——触发留痕仍在、run 不落、disabled 计数可见；
	// 判定在锁外（解析面文件 IO 不持 r.mu——与刹车生效值同一拍面），缺席（ok=false）=放行
	// 既有行为面（判定面不放大解析形变为整面拦阻——缺口可见不静默）。在途 run 的迁移判定
	// 不受开关作用（ResumeSignal/Advance 零改动——已完成轮次与回归记录保留）。
	brk := effectiveBrakes(r.opts.Brakes.Resolver, wf)
	if r.opts.EnabledRules != nil {
		if on, ok := r.opts.EnabledRules.Enabled(wf); ok && !on {
			r.mu.Lock()
			r.disabled++
			r.mu.Unlock()
			return
		}
	}
	r.mu.Lock()
	var judgeQ []spawnReq // 自审批判决派发请求（锁内收集、出锁派发——LLM 不持 run 锁）
	defer func() {
		r.mu.Unlock()
		for _, sp := range judgeQ {
			r.dispatchSpawnReq(sp)
		}
	}()
	if r.opts.Brakes.Enabled && brk.Cooldown > 0 {
		key := brakeKey(wf, rec.Involved)
		if last, ok := r.lastAccepted[key]; ok && rec.ReceivedAt.Sub(last) < brk.Cooldown {
			r.braked++ // 拦阻=触发留痕仍在、run 不落（防震荡——冷却窗口语义）
			return
		}
		r.lastAccepted[key] = rec.ReceivedAt
	}
	r.seq++
	r.started++
	if excess := len(r.ring) + 1 - r.opts.RingSize; excess > 0 {
		r.ring = r.ring[excess:]
		r.dropped += int64(excess)
	}
	phase := ""
	status := RunTriggered
	var hist []HistoryRow
	if len(phases) > 0 {
		phase = string(phases[0]) // 受理投影=声明首相位（起点词先落，迁移判定在其上判）
		// 多节点入口切入（2026-11-27 解释器多入口：如用户对话修改文档或临时 BugDoc 触发 DocApplied，
		// 声明表含 on="entry:DocApplied" → 直接切入 Planning / doc-apply 节点，再顺沿 advance 进入 Executing）。
		if rec.Type != "" {
			if entryTo, ok := stepByTable(trans, phase, "entry:"+rec.Type); ok {
				hist = append(hist, HistoryRow{From: phase, To: entryTo, On: "entry:" + rec.Type, At: stampNano(rec.ReceivedAt)})
				phase, status = entryTo, RunRunning
			}
		}
		// 相位推进器首段（v0.23）+显式迁移表首段（v0.25）：受理即启动信号——引擎单步迁移
		// 判定，按声明表查推进行（from=当前入口相位 on=advance）并把状态词翻转为 running
		//（stepByTable 纯查表，判定口径见包注释——序列前进=表的特例）。
		if next, ok := stepByTable(trans, phase, EdgeAdvance); ok {
			hist = append(hist, HistoryRow{From: phase, To: next, On: EdgeAdvance, At: stampNano(rec.ReceivedAt)})
			phase, status = next, RunRunning
		}
		// 挂起-恢复首段（v0.24）：迁移判定落点=闸锚 → 挂起（相位机等待外部审批信号，
		// workflows 规则 13 第一层判定；auto 盲放行；agent 挂接 judge=真判决挂起，
		// judge 缺席=旧词形零翻转盲放行）。
		if status == RunRunning && Phase(phase) == AwaitingApproval && !r.gateAutoPass(approvalMode) {
			status = RunSuspended
		}
	}
	if phases == nil {
		phases = []string{} // 相位未投影（未声明 phases 的 object 源）：载荷形状归一为空数组（null=缺席语义，[]=声明为空——消费方可 len）
	}
	if trans == nil {
		trans = []TransitionRow{} // object 源不投影声明表：载荷形状归一同口径（新键可 len）
	}
	nr := RunRecord{
		Run: r.seq, Workflow: wf, SpecSource: source, Phases: phases, Transitions: trans,
		ApprovalMode: approvalMode, ApprovalPolicy: approvalPolicy, Nodes: nodes,
		Phase: phase, Status: status, Brakes: brk, History: hist,
		TriggerSeq: rec.Seq, EventID: rec.EventID, Type: rec.Type, DedupKey: rec.DedupKey,
		Route: rec.Route, Owner: rec.Owner, Target: b.Target, Involved: rec.Involved, StartedAt: rec.ReceivedAt,
		ParentRun: rec.ParentRun,
	}
	if nr.Status == RunSuspended && Phase(nr.Phase) == AwaitingApproval && r.selfApproveMode(nr.ApprovalMode) {
		if sp := r.beginSelfApproveLocked(&nr); sp != nil {
			judgeQ = append(judgeQ, *sp)
		}
	}
	// connectops 落点钩子（2026-11-28 opsrun.go）：受理首步直落 connectops 执行单元且
	// ConnectOpsRunner 已挂接 → 与 Advance 同一挂起-派发落面（首步不走 landAfterAdvanceLocked
	// 的历史缺口——入口节点=connectops 的声明词形如 gh-issue-fetch 原先机械越过不执行）。
	// 窄口=仅 ops（service/fanout 首步既有行为面不动——landAfter 复用会连带翻转）。
	if nr.Status == RunRunning && r.opts.ConnectOpsRunner != nil && Phase(nr.Phase) != AwaitingApproval {
		if node, ok := opsNodeOf(nr.Nodes, nr.Phase); ok {
			nr.History = append(nr.History, HistoryRow{From: nr.Phase, To: nr.Phase, On: EdgeOpsRun, At: stampNano(rec.ReceivedAt)})
			nr.Status = RunSuspended
			nr.SuspendWhy = SuspendOps
			judgeQ = append(judgeQ, spawnReq{kind: spawnKindOps, run: nr, node: node})
		}
	}
	// service 落点钩子（2026-11-28 sybran-release-update 用户自有 workflow 首条真派发——
	// 与 ops 钩子同缺口同修法：受理首步直落 unit=service 且 ServiceSpawner 已挂接 → 挂起+
	// 拉起派发；既有 spec 的 service 节点均在相位链中后段（经 Advance 落点）不受影响）。
	if nr.Status == RunRunning && r.opts.ServiceSpawns != nil && Phase(nr.Phase) != AwaitingApproval {
		if node, ok := serviceNodeOf(nr.Nodes, nr.Phase); ok {
			nr.History = append(nr.History, HistoryRow{From: nr.Phase, To: nr.Phase, On: EdgeServiceSpawn, At: stampNano(rec.ReceivedAt)})
			nr.Status = RunSuspended
			nr.SuspendWhy = SuspendService
			judgeQ = append(judgeQ, spawnReq{kind: spawnKindService, run: nr, node: node})
		}
	}
	r.ring = append(r.ring, nr)
	// 写穿透（R-ENG §1）：run 落环同步 upsert 至挂接的 Store（run 是可演化投影——同 run
	// id 覆写最新词形；锁内同步落库与投影同序，失败只注记不翻转行为）。
	if r.opts.Store != nil {
		persistAppend("run 投影", r.opts.Store.AppendRun(runToRow(r.ring[len(r.ring)-1])))
	}
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
	var judgeQ []spawnReq // 审批提升触发的真判决派发（出锁后 dispatch）
	defer func() {
		r.mu.Unlock()
		for _, sp := range judgeQ {
			r.dispatchSpawnReq(sp)
		}
	}()
	// 定向 run（TargetRun>0，console 批闸行携带）：只作用该序号 run——per-run 批准/驳回
	// 不再落到「同锚最新挂起 run」（多 run 同锚挂起时 UI 行与事件目标必须一致）。
	target := rec.TargetRun
	for i := len(r.ring) - 1; i >= 0; i-- {
		run := &r.ring[i]
		if target > 0 && run.Run != target {
			continue
		}
		if run.Workflow != name || len(run.Phases) == 0 {
			continue // 锚不符 / 相位未投影（未声明 phases 的 object 源）无从迁移
		}
		switch action {
		case "suspend": // 审批提升：相位落闸锚挂起（终态/已挂起=零动作；声明无闸锚=零动作；auto 盲放行不挂起；agent 挂 judge=真判决）
			if r.gateAutoPass(run.ApprovalMode) {
				return
			}
			if run.Status == RunFailed || run.Status == RunSuspended ||
				!containsPhase(run.Phases, string(AwaitingApproval)) {
				return
			}
			histAppend(run, run.Phase, string(AwaitingApproval), SignalSuspend)
			run.Phase, run.Status = string(AwaitingApproval), RunSuspended
			if r.selfApproveMode(run.ApprovalMode) {
				if sp := r.beginSelfApproveLocked(run); sp != nil {
					judgeQ = append(judgeQ, *sp)
				}
			}
		case "resume": // 批准：查表命中审批边（闸锚 on=ApprovalGranted）沿表续行（非挂起态=继续扫描更旧 run——同锚最新 run 已终态时不挡旧挂起 run 的恢复；表无批准边=零动作防御）
			if run.Status != RunSuspended || Phase(run.Phase) != AwaitingApproval {
				continue
			}
			next, ok := stepByTable(run.Transitions, run.Phase, SignalResume)
			if !ok {
				return
			}
			histAppend(run, run.Phase, next, SignalResume)
			run.Phase = next
			run.Status = RunRunning // 闸锚续行落点非闸锚（序列无连续闸锚防御：若落闸锚再挂起）
			if Phase(next) == AwaitingApproval {
				run.Status = RunSuspended
			}
		case "reject": // 驳回：查表命中驳回边（闸锚 on=ApprovalRejected → 侧轨终点锚 Failed——表的首处非相邻行）终态（非挂起态=继续扫描更旧 run——同 resume 口径；表无驳回边=零动作）
			if run.Status != RunSuspended {
				continue
			}
			to, ok := stepByTable(run.Transitions, run.Phase, SignalReject)
			if !ok {
				return
			}
			histAppend(run, run.Phase, to, SignalReject)
			run.Phase, run.Status = to, RunFailed
			r.notifyParentLocked(run) // 子 run 驳回终态→父 join（#6）
		}
		// 写穿透（R-ENG §1）：审批信号迁移同步 upsert（可演化投影——恢复取迁移后词形）。
		if r.opts.Store != nil {
			persistAppend("run 迁移", r.opts.Store.AppendRun(runToRow(*run)))
		}
		return // 只作用现势 run（最新命中即返回）
	}
}

// ServiceSignal service 完成信号的 run 域续行判定（2026-11-27 service session 裁决——
// serviceexec.go；hub 按可选能力接口 ServiceSignalSink 转发，cored 装配的桥在翻完拉起投影
// 终态后转交本入口并携带会话锚）。匹配口径与 ResumeSignal 同款：service 信号族事件按
// involved=Workflow 对象引用寻径，作用于该锚**最新挂起 run**（TargetRun>0=定向覆盖）中
// SuspendWhy=service 的现势 run；命中=沿轨迹落续行行（on=service-resume，SessionID=会话锚
// 回执——真实会话锚在续行行留痕）、status 翻回 running（SuspendWhy 清空），下拍 Advance
// 沿表续行（落点若又是 service 节点=下拍再挂起再拉起——developer-system→verify 链式通路）。
// 零动作防御面：非本族事件/involved 非 Workflow 引用/锚无 run/非 service 挂起态（含审批
// 挂起——approval 归 ResumeSignal 管辖）/已续行（重复完成信号幂等 noop）一律原样保留。
// session 参数=拉起投影回执的会话锚（空=诚实缺席，续行行不带会话锚不编造）。
func (r *RunLog) ServiceSignal(rec TriggerRecord, session string) {
	if !isServiceSignalType(rec.Type) {
		return
	}
	name, ok := workflowRefOf(rec.Involved)
	if !ok {
		return
	}
	// 判断分叉派发队列（锁内收集出锁派发——decision 边落点=service/ops 节点时与 Advance
	// 同一挂起-派发落面；2026-11-28 用户裁决「跳过这一步可以在workflow做分叉」）。
	var forkQ []spawnReq
	defer func() {
		r.mu.Unlock()
		for _, sp := range forkQ {
			r.dispatchSpawnReq(sp)
		}
	}()
	r.mu.Lock()
	target := rec.TargetRun
	for i := len(r.ring) - 1; i >= 0; i-- {
		run := &r.ring[i]
		if target > 0 && run.Run != target {
			continue
		}
		if run.Workflow != name || len(run.Phases) == 0 {
			continue
		}
		if run.Status != RunSuspended || run.SuspendWhy != SuspendService {
			continue // 非 service 挂起态（审批挂起/终态/在途）不作用——继续扫描更旧 run
		}
		histAppend(run, run.Phase, run.Phase, EdgeServiceSpawnResume)
		if session != "" {
			run.History[len(run.History)-1].SessionID = session
		}
		run.Status = RunRunning
		run.SuspendWhy = ""
		// 判断分叉（2026-11-28 用户裁决）：回执携带 decision 词且当前相位有该判断边 →
		// 立即沿判断边走（迁移行 on=<判断词>——与机械 advance 行分词可对账）；落点=执行
		// 单元节点则同一挂起-派发落面（forkQ 出锁派发）。未命中/未携带=纯续行（机械
		// advance 兜底——默认路径不动）。
		if d := rec.Decision; d != "" {
			if next, ok := stepByTable(run.Transitions, run.Phase, d); ok && next != run.Phase {
				run.History = append(run.History, HistoryRow{From: run.Phase, To: next, On: d, At: stampNano(time.Now())})
				run.Phase = next
				if sp := r.landAfterAdvanceLocked(run, next); sp != nil {
					forkQ = append(forkQ, *sp)
				}
			}
		}
		// 写穿透（R-ENG §1）：续行同步 upsert（可演化投影——恢复取续行后词形）。
		if r.opts.Store != nil {
			persistAppend("run service 续行", r.opts.Store.AppendRun(runToRow(*run)))
		}
		return
	}
}

// OpsSignal connectops op 完成回执续行（2026-11-28 opsrun.go——进程内直接调用，不走事件面：
// 完成源与引擎同进程，service 完成源在宿主进程故走事件注入，两路径各自的诚实边界）。
// ok=true → 续行（history 行 on=connectops-resume，Note=结果摘要）；ok=false → 保持挂起
// （history 行 on=connectops-error，Note=显式错误——缺口可见不静默，service 失败同口径）。
// decision 判断词（2026-11-28 分叉裁决——op 结果 JSON decision 字段；语义同 ServiceSignal）。
func (r *RunLog) OpsSignal(run int64, workflowName, note string, ok bool, decision string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := len(r.ring) - 1; i >= 0; i-- {
		rg := &r.ring[i]
		if run > 0 && rg.Run != run {
			continue
		}
		if rg.Workflow != workflowName || len(rg.Phases) == 0 {
			continue
		}
		if rg.Status != RunSuspended || rg.SuspendWhy != SuspendOps {
			continue // 非 connectops 挂起态（审批挂起/终态/在途）不作用——继续扫描更旧 run
		}
		edge := EdgeOpsResume
		if !ok {
			edge = EdgeOpsError
		}
		histAppend(rg, rg.Phase, rg.Phase, edge)
		rg.History[len(rg.History)-1].Note = note
		if ok {
			rg.Status = RunRunning
			rg.SuspendWhy = ""
			// 判断分叉（同 ServiceSignal——op 结果 decision 字段；命中判断边立即走+落点钩子）。
			if d := decision; d != "" {
				if next, ok2 := stepByTable(rg.Transitions, rg.Phase, d); ok2 && next != rg.Phase {
					rg.History = append(rg.History, HistoryRow{From: rg.Phase, To: next, On: d, At: stampNano(time.Now())})
					rg.Phase = next
					if sp := r.landAfterAdvanceLocked(rg, next); sp != nil {
						r.dispatchSpawnReq(*sp) // OpsSignal 调用方=桥 goroutine（非锁路径）——同步派发安全
					}
				}
			}
		}
		// 写穿透（R-ENG §1）：回执同步 upsert（可演化投影——恢复取回执后词形；失败态同样留账）。
		if r.opts.Store != nil {
			persistAppend("run connectops 回执", r.opts.Store.AppendRun(runToRow(*rg)))
		}
		return
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

// histAppend 运行轨迹追加（F-04——真实迁移留痕；Phase 改写前调用，from=改写前词形。
// 上限 64 行防长 run 无界增长——轨迹属投影面非账本，环挤出即随之消失）。
func histAppend(run *RunRecord, from, to, on string) {
	const histMax = 64
	if len(run.History) >= histMax {
		run.History = run.History[len(run.History)-histMax+1:]
	}
	run.History = append(run.History, HistoryRow{From: from, To: to, On: on, At: stampNano(time.Now()), SessionID: run.SessionID})
}

// State 查询快照（拷贝返回，调用方可安全持有；runs 旧→新）。
func (r *RunLog) State() RunState {
	r.mu.Lock()
	defer r.mu.Unlock()
	runs := make([]RunRecord, len(r.ring))
	copy(runs, r.ring)
	return RunState{Started: r.started, Dropped: r.dropped, Advances: r.advances, Braked: r.braked, Capped: r.capped, Disabled: r.disabled, Gated: r.gated, SelfApproved: r.selfApproved, Runs: runs}
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

// RunAnchor 派发 route 原文 → run 锚名与声明来源（v0.37 console 读面第二段导出面：
// `GET /v1/workflows` 聚合投影按 dispatch 决议归档计数时复用同一锚解析——聚合与 run 投影
// 同一解析函数，无双源；空锚=""）。source 词汇：builtin=内置注册表命中；object=平台对象引用。
func RunAnchor(routeTarget string) (wf, source string) {
	wf, source, _, _ = resolveRunAnchor(routeTarget)
	return wf, source
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
