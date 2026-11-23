// trigger.go —— 触发骨架（dispatch 接线：第一段 v0.11 订阅留痕 → 第二段 v0.12 订阅面动态化
// → 第三段 v0.13 引擎主体第一段：owner 解析激活 + 派发决策可查询
// → 第四段 v0.14 引擎主体第二段：派发执行投影——route/cc 目标 → 执行面绑定可查询
// → 第五段 v0.15 引擎主体第三段（真实拉起第一子段）：workflow 触发面接线——
// route 目标执行面=workflow 且装配层已登记（WireExecutor+BindRuns）时，留痕推进为一次
// 对账 run 投影（run.go；查询面 /v1/workflows/runs）
// → 第六段 v0.16 引擎主体第四段（真实拉起第二子段）：notification 送达面接线——
// 执行面=notification 的绑定（route 主路与 cc 旁路，origin 维度留痕）且装配层已登记
// （WireExecutor+BindDeliveries）时，留痕推进为一条送达投影（deliver.go；查询面
// /v1/notifications/deliveries）→ 第四段收尾 v0.17：受理后真实通道出站挂接（送达状态机
// accepted→sent/failed，deliver.go）
// → 第七段 v0.18 引擎主体第五段（真实拉起第三子段）：agent 唤醒面接线——
// 执行面=agent 的绑定（route 主路与 cc 旁路，origin 维度留痕）且装配层已登记
// （WireExecutor+BindAgents）时，留痕推进为一条唤醒投影（wake.go；查询面 /v1/agents/wakes）
// → 第八段 v0.24 挂起-恢复首段：审批信号族 run 域迁移——审批信号按可选能力接口 ResumeSink
// 转发给 runs sink（不是新对账触发，不另起 run；run.go ResumeSignal 三层迁移判定）：
// 订阅总线事件流，收到即落**进程内可查状态**（收到留痕：事件 id/dedupKey/type/at/receivedAt
// + 命中行投影 route/for/owner/cc + 执行绑定投影 executor/target/wired）。
// 订阅集合由 scheme.dispatch 表推导（on: 值表序去重、通配 * 坍缩全类型一条）+ for 谓词口径
// （any/owner:*/scope:<root>/*/**owner:<字面量>**）：owner 解析模型=involved scope 根
// `systems/<s>`→`<s>`、global/builtin 同名（event.md 规则 6）。
// 死信兜底路由可见（全表未命中 → 留痕 route=dead-letter，event.md rule 4）。
// **派发决策面可查询**（Resolve / GET /v1/triggers/resolve）：对当前表快照回答「这一事件
// 若到达会路由到哪」——纯决策（零事件注入、零留痕、零派发），命中行投影 route/cc/for/owner/row
// + execution 投影（route/cc 目标的执行面绑定：映射表 execFor 家族收编，未知目标=unbound
// 缺口可见不静默吞并；wired=执行面接线态，装配层登记前恒 false——诚实投影）。
// 触发引擎主体剩余面（相位态机→刹车→冷却→minLevel、通知重试/降级/静默窗口、agent 真实
// 拉起）declared-deferred（cored-bridge.md §2.3）——决策面回答「路由到哪 + 谁来执行」，
// workflow 面把触发推进为对账 run 起点（v0.15），notification 面把绑定推进为送达受理投影
// （v0.16，route+cc 双路），agent 面把绑定推进为唤醒受理投影（v0.18，route+cc 双路），
// 真实拉起/相位推进仍 deferred。
package workflow

import (
	"strings"
	"sync"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// Wildcard 订阅通配令牌：Attach 时映射为总线全类型订阅（eventbus typ=""），
// 留痕 types 集合中以 "*" 呈现（表驱动判别锚：types=[*] ≠ 固定枚举）。
const Wildcard = "*"

// DeadLetterRoute 全 dispatch 表未命中的留痕路由标注（死信必须可见，event.md rule 4；
// 与 eventbus event_log routed_to="dead-letter" 同一词汇表）。
const DeadLetterRoute = "dead-letter"

// NoRow 派发决策未命中任何表行时的行号标注（决策不可归因到表行：全表未命中=合成死信、
// 未挂表=无决策；row≥0 恒可回指 scheme.yaml 具体行）。
const NoRow = -1

// 执行面种类（ExecutorBinding.Executor 词汇；映射表 execFor 家族收编）。
const (
	ExecWorkflow     = "workflow"     // workflow 触发面（owner 对账 workflow / 直接 workflows/<n> 引用）
	ExecNotification = "notification" // notification 送达面（通道投影，*=全部已绑通道）
	ExecAgent        = "agent"        // agent 唤醒面（parent、root 等前缀）
	ExecNone         = "none"         // 无执行（死信=观测兜底出口，event.md rule 4）
	ExecUnbound      = "unbound"      // 未知目标族（映射表未收编——缺口可见，不静默吞并）
)

// ExecutorBinding route/cc 目标的执行面绑定投影（引擎主体第二段 v0.14——纯映射不执行）。
// 决策面回答「路由到哪」之后的下一问「谁来执行」：绑定=执行面种类+锚点+接线态，
// 真实拉起（workflow 触发/notification 送达/agent 唤醒）仍属引擎主体余量。
type ExecutorBinding struct {
	Executor string `json:"executor"` // 执行面种类（Exec* 之一；空串=无目标，如空 cc）
	Target   string `json:"target"`   // 执行面锚点投影（通道名/workflow 名/agent 前缀；none/unbound 按映射行语义）
	Wired    bool   `json:"wired"`    // 执行面是否已接线（装配层登记前恒 false；落地轮翻转=机械判别器）
}

// DispatchDecision 一次派发决策（表序命中行解析结果）。Resolve 查询面与 record 留痕
// 共用同一解析口径（resolveDecisionLocked）——决策=数据，消费面（留痕/查询/未来派发执行）
// 各取投影，这是引擎主体可分步落地的解耦点。
type DispatchDecision struct {
	Matched bool   // 表内命中（含显式 dead-letter 兜底行）；false=全表未命中（合成死信）或未挂表
	Route   string // 命中行 dispatch 投影（全表未命中=DeadLetterRoute 合成值；未挂表=空）
	CC      string // 命中行 cc 投影（旁路目标；执行随派发 deferred）
	For     string // 命中行 for 谓词原文
	Owner   string // owner 解析值（谓词匹配依据；与具体行无关恒可解析）
	Row     int    // 命中表行号（0 起，可回指 scheme.yaml）；未命中=NoRow

	Exec   ExecutorBinding // route 目标执行绑定投影（execFor 映射；合成死信=none）
	CCExec ExecutorBinding // cc 目标执行绑定投影（空 cc=零值绑定 executor=""）
}

// TriggerRecord 一条收到留痕（订阅方视角；事件本体语义见 concepts/event.md——
// 事件只触发不修改、可丢可重可乱序，留痕同样只是观测面）。
type TriggerRecord struct {
	Seq        int64     // 收到序号（本 hub 进程内单调递增，1 起）
	EventID    string    // 与发布方回执同源（apply 响应 event.id / event_log.event_id）
	Type       string    // 事件类型（Scheme.eventTypes 之一）
	DedupKey   string    // 与发布方回执同源（apply:kind/scope/name:rv）
	Source     string    // 来源（apply/connector/metrics/timer/manual）
	Involved   string    // involvedRef 投影（kind/scope/name）
	Level      string    // debug/info/warn/error/fatal（A42）
	At         time.Time // 事件时间（发布方时钟）
	ReceivedAt time.Time // 收到时刻（订阅方时钟；与 At 差值=总线+调度延迟观测点）
	Route      string    // 命中 dispatch 行的 route 投影（留痕=过去的决策投影，不执行派发）；
	//             全表未命中=dead-letter；未挂 dispatch 表（裸 Attach）=空
	CC    string // 命中行的 cc 投影（旁路目标，与 Route 同一行；执行随派发 deferred）
	For   string // 命中行的 for 谓词原文（与 Route 同一行；便于机械对账表行↔留痕）
	Owner string // owner 解析值（命中谓词的匹配依据：involved scope 根 systems/<s>→<s>，
	//         global/builtin 同名=scope 根本身；决策依据可查——v0.13）

	Exec   ExecutorBinding // route 目标执行绑定投影（收到时刻的执行计划快照——v0.14）
	CCExec ExecutorBinding // cc 目标执行绑定投影（空 cc=零值绑定）
}

// TriggerOptions 触发骨架可调项（零值可用）。
type TriggerOptions struct {
	// RingSize 留痕环形容量（缺省 256）：溢出丢最旧、dropped 计数可见。
	// 留痕是观测面不是账本（账本=event_log 归档，ops-log.md §2）。
	RingSize int
}

// TriggerState 查询快照（State() 返回值；apiserver GET /v1/triggers 的载荷来源）。
type TriggerState struct {
	Types    []string        // 已订阅事件类型（Attach 集合；"*"=通配全类型，dispatch 表推导）
	Received int64           // 累计收到（含已被环挤出的）
	Dropped  int64           // 环溢出丢弃计数（可观测）
	Records  []TriggerRecord // 环内留痕（旧→新）
}

// TriggerHub 触发骨架（**单订阅者**）：每事件类型一条总线订阅，收到即落留痕。
// 并发安全；退订后订阅 goroutine 排干频道存量自然退场（不丢已入队事件）。
type TriggerHub struct {
	opts TriggerOptions

	mu       sync.Mutex
	seq      int64
	ring     []TriggerRecord
	received int64
	dropped  int64
	types    []string
	cancels  []func()
	rules    []scheme.DispatchRule // dispatch 表快照（AttachDispatch 挂接；裸 Attach 为空=留痕不标 route）
	wired    map[string]bool       // 执行面接线态（WireExecutor 登记；缺省全 false=真实拉起 deferred）
	runs     RunSink               // workflow 触发面 run 投影落点（BindRuns 挂接；nil=不推进 run）
	deliver  DeliverySink          // notification 送达面送达投影落点（BindDeliveries 挂接；nil=不推进送达）
	wakes    WakeSink              // agent 唤醒面唤醒投影落点（BindAgents 挂接；nil=不推进唤醒）
	advance  EventAdvanceSink      // 事件驱动推进面落点（v0.29 BindEventAdvance 挂接；nil=事件拍缺席——推进总闸关断）
}

// SourceEngineInternal 引擎内部源事件的 source 审计值（event.md 规则 10 服务侧判定口径——
// 引擎自身产生的心跳/巡检类来源标注；与 patrol.EventSource 同字面，声明面=event.md 规则 10）。
const SourceEngineInternal = "core"

// SourceConnectorWatch connector watch 桥接源事件的 source 审计值（event.md 规则 12 服务
// 侧判定口径——外部变化来源标注是审计事实；与 watchsrc.EventSource 同字面，声明面=
// event.md 规则 12）。
const SourceConnectorWatch = "connector"

// beatSources 事件拍驱动源集合（引擎装配的内部源——workflow.md 规则 15：v0.29 core=
// 定时巡检 ReconcileTick，v0.30 connector=watch 桥接源 ExternalChanged；metrics 阈值等
// 随各自声明面扩展顺延一行即入）。注入通道（source 服务侧判定恒 human）与权限自产
// （permission）不在拍面——来源维度分立（v0.29 判别器措辞收窄：注入通道本就恒 human）。
var beatSources = map[string]bool{
	SourceEngineInternal: true,
	SourceConnectorWatch: true,
}

// NewTriggerHub 构建触发骨架。
func NewTriggerHub(opts *TriggerOptions) *TriggerHub {
	h := &TriggerHub{}
	if opts != nil {
		h.opts = *opts
	}
	if h.opts.RingSize <= 0 {
		h.opts.RingSize = 256
	}
	return h
}

// Attach 订阅事件类型（每类型一条总线订阅；同类型重复 Attach 去重；通配令牌 "*" 映射
// 总线全类型订阅一条）。返回总退订函数（幂等）：退订全部订阅，各订阅 goroutine 排干存量后退场。
// bus 为 nil 时返回空操作（装配层无总线部署不受影响）。
func (h *TriggerHub) Attach(bus *eventbus.Bus, types ...string) func() {
	if bus == nil {
		return func() {}
	}
	h.mu.Lock()
	for _, typ := range types {
		if h.hasTypeLocked(typ) {
			continue
		}
		busTyp := typ
		if typ == Wildcard {
			busTyp = "" // 总线口径："" = 全类型
		}
		ch, cancel := bus.Subscribe(busTyp)
		h.types = append(h.types, typ)
		h.cancels = append(h.cancels, cancel)
		go h.drain(ch)
	}
	cancels := append([]func(){}, h.cancels...)
	h.mu.Unlock()
	return func() { // 总退订（总线 cancel 幂等，重复调用安全）
		for _, c := range cancels {
			c()
		}
	}
}

// AttachDispatch 按 scheme.dispatch 表推导订阅集合并挂接（dispatch 第二段 v0.12：订阅面动态化）。
// 推导：on 值按表序去重；含通配 "*" → 坍缩为单条通配订阅（全类型订阅已覆盖枚举，并存会重复留痕）；
// 空表=不订阅（事件全部落总线死信，路由不到可见，event.md rule 4）。
// 表快照随挂接存入 hub：此后每条留痕按表序解析命中行（route/for 投影，只留痕不派发）。
// 返回总退订函数（同 Attach，幂等）。bus 为 nil 时返回空操作。
func (h *TriggerHub) AttachDispatch(bus *eventbus.Bus, rules []scheme.DispatchRule) func() {
	if bus == nil {
		return func() {}
	}
	snapshot := make([]scheme.DispatchRule, len(rules))
	copy(snapshot, rules)
	h.mu.Lock()
	h.rules = snapshot
	h.mu.Unlock()

	seen := map[string]bool{}
	var types []string
	wildcard := false
	for _, r := range snapshot {
		switch {
		case r.On == "":
			continue // 装配层防御：Load 校验已拒空 on（scheme.validate）
		case r.On == Wildcard:
			wildcard = true
		default:
			if !seen[r.On] {
				seen[r.On] = true
				types = append(types, r.On)
			}
		}
	}
	if wildcard {
		return h.Attach(bus, Wildcard)
	}
	return h.Attach(bus, types...)
}

// drain 订阅 goroutine：总线退订关频道后 range 排干存量退出（已入队事件不丢）。
func (h *TriggerHub) drain(ch <-chan eventbus.Event) {
	for ev := range ch {
		h.record(ev)
	}
}

// record 收到留痕（环满丢最旧，dropped 计数可见——观测面非账本）。
// 挂有 dispatch 表时解析命中行：route/cc/for/owner + 执行绑定投影留痕（决策投影，不执行派发）；
// 全表未命中 → route=dead-letter（死信兜底可见）；未挂表（裸 Attach）→ 路由字段留空。
func (h *TriggerHub) record(e eventbus.Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.seq++
	h.received++
	if excess := len(h.ring) + 1 - h.opts.RingSize; excess > 0 {
		h.ring = h.ring[excess:]
		h.dropped += int64(excess)
	}
	d := h.resolveDecisionLocked(e.Type, e.Involved)
	rec := TriggerRecord{
		Seq: h.seq, EventID: e.ID, Type: e.Type, DedupKey: e.DedupKey,
		Source: e.Source, Involved: e.Involved, Level: e.Level,
		At: e.At, ReceivedAt: time.Now(),
		Route: d.Route, CC: d.CC, For: d.For, Owner: d.Owner,
		Exec: d.Exec, CCExec: d.CCExec,
	}
	h.ring = append(h.ring, rec)
	// 真实拉起第一子段（v0.15）：workflow 触发面已接线（WireExecutor 登记 + BindRuns 挂接）
	// 时，route 目标执行面=workflow 的留痕推进为一次对账 run 投影（v0.23 相位推进器首段：
	// 受理即启动——引擎单步迁移判定随受理同拍完成；v0.25 显式迁移表首段：判定器=声明表
	// 查表，run.go stepByTable；后续推进/挂起-恢复已落 v0.24，异步推进 deferred）。
	// 同步调用：sink 契约=非阻塞进程内投影（run.go RunSink），锁内调用保 run 与留痕同序。
	landed := false
	if h.runs != nil && h.wired[ExecWorkflow] && d.Exec.Executor == ExecWorkflow {
		h.runs.StartRun(rec, d.Exec)
		landed = true
	}
	// 事件驱动推进源首段（v0.29——巡检拍与推进拍合一，workflow.md 规则 15 v0.29 段）：引擎
	// 内部源事件落 run 的同拍对该拍现势 run 全量执行一次单步推进（与定时拍同一查表落点——
	// 判定器零改动，只加驱动源；本拍新落 run 参与本拍——全量口径不含特判，受理首步与事件拍
	// 是两条独立口径的拍步）。拍的三重闸：①落了 run（未落 run 的事件不拍）；②source ∈ 拍面
	// 驱动源集合 beatSources（引擎装配的内部源审计事实——v0.29=core 巡检、v0.30+connector
	// watch 桥；注入通道恒 human 与权限自产 permission 不拍，source 维度分立）；③装配
	// 层已挂接（BindEventAdvance 只在推进总闸开启时调用——事件驱动源不旁路总闸，关断语义
	// 组合声明面先定）。锁内同步调用：与留痕/run 同序（确定性——新落 run 同拍即得第二拍步）。
	if landed && beatSources[rec.Source] && h.advance != nil {
		h.advance.Advance()
	}
	// 真实拉起第二子段（v0.16）：notification 送达面已接线（WireExecutor 登记 +
	// BindDeliveries 挂接）时，执行面=notification 的绑定推进为一条送达投影——route 主路与
	// cc 旁路双路收编（origin 维度留痕；通知本身即旁路语义，单 sink 双路，deliver.go 口径）。
	// 同一锁内推进：送达与留痕/run 同序；非 notification 执行面与未接线态不推进（单登记不生效）。
	if h.deliver != nil && h.wired[ExecNotification] {
		if d.Exec.Executor == ExecNotification {
			h.deliver.Deliver(rec, d.Exec, OriginRoute)
		}
		if d.CCExec.Executor == ExecNotification {
			h.deliver.Deliver(rec, d.CCExec, OriginCC)
		}
	}
	// 真实拉起第三子段（v0.18）：agent 唤醒面已接线（WireExecutor 登记 + BindAgents 挂接）
	// 时，执行面=agent 的绑定推进为一条唤醒投影——route 主路与 cc 旁路双路收编（origin
	// 维度留痕，与送达面同口径；wake.go）。同一锁内推进：唤醒与留痕/run/送达同序；
	// 非 agent 执行面与未接线态不推进（单登记不生效）。
	if h.wakes != nil && h.wired[ExecAgent] {
		if d.Exec.Executor == ExecAgent {
			h.wakes.Wake(rec, d.Exec, OriginRoute)
		}
		if d.CCExec.Executor == ExecAgent {
			h.wakes.Wake(rec, d.CCExec, OriginCC)
		}
	}
	// 挂起-恢复首段（v0.24）：审批信号族（ApprovalNeeded/ApprovalGranted/ApprovalRejected，
	// event.md 规则 11/workflows 规则 13）——run 域迁移事件，**不是新对账触发**（不另起
	// run）：hub 按可选能力接口 ResumeSink 转发给 runs sink（类型断言失败=未声明能力不转发，
	// 能力增量不破既有 sink 契约），involved=Workflow 引用寻径与三层迁移判定在 sink 侧
	//（run.go ResumeSignal）。与 dispatch 派发分立（审批信号族 dispatch 行=dead-letter 标注
	// 「派发面无目标」），同一订阅主干零旁路（同留痕/同去重/同归档）；锁内转发与留痕/run 同序。
	if h.runs != nil {
		if rs, ok := h.runs.(ResumeSink); ok && isApprovalSignalType(rec.Type) {
			rs.ResumeSignal(rec)
		}
	}
}

// WireExecutor 登记执行面接线态（引擎主体第二段 v0.14 登记 → 第三段 v0.15 起装配层对
// 已落地执行面实际登记）。登记前投影 wired 恒 false——诚实投影：执行计划可查 ≠ 执行面
// 已在位。幂等；登记即生效（下次决策投影照读）。v0.15 装配层登记 ExecWorkflow
// （run 投影面在位，run.go）；v0.16 登记 ExecNotification（送达投影面在位，deliver.go）；
// v0.18 登记 ExecAgent（唤醒投影面在位，wake.go）。
func (h *TriggerHub) WireExecutor(kinds ...string) {
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.wired == nil {
		h.wired = map[string]bool{}
	}
	for _, k := range kinds {
		h.wired[k] = true
	}
}

// BindRuns 挂接 workflow 触发面的 run 投影落点（引擎主体第三段 v0.15）。挂接本身不翻转
// wired——真实拉起的接线判定=装配层**同时**登记 WireExecutor(ExecWorkflow)（机械判别器：
// 决策投影 wired 翻转 ∧ run 面出现）。幂等；nil=解除挂接。须在 AttachDispatch 前调用
// （事件开始流入后无接线窗口）。sink 契约见 RunSink（非阻塞、不回调 hub）。
func (h *TriggerHub) BindRuns(sink RunSink) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.runs = sink
}

// BindDeliveries 挂接 notification 送达面的送达投影落点（引擎主体第四段 v0.16）。挂接本身
// 不翻转 wired——真实拉起的接线判定=装配层**同时**登记 WireExecutor(ExecNotification)
// （机械判别器：决策投影 wired 翻转 ∧ 送达面出现）。幂等；nil=解除挂接。须在 AttachDispatch
// 前调用（事件开始流入后无接线窗口）。sink 契约见 DeliverySink（非阻塞、不回调 hub）。
func (h *TriggerHub) BindDeliveries(sink DeliverySink) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.deliver = sink
}

// BindAgents 挂接 agent 触发面的唤醒投影落点（引擎主体第五段 v0.18）。挂接本身不翻转
// wired——真实拉起的接线判定=装配层**同时**登记 WireExecutor(ExecAgent)
// （机械判别器：决策投影 wired 翻转 ∧ 唤醒面出现，与 run/送达面同口径）。幂等；nil=解除
// 挂接。须在 AttachDispatch 前调用（事件开始流入后无接线窗口）。sink 契约见 WakeSink
// （非阻塞、不回调 hub）。
func (h *TriggerHub) BindAgents(sink WakeSink) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.wakes = sink
}

// BindEventAdvance 挂接事件驱动推进面（v0.29 事件驱动推进源首段，workflow.md 规则 15
// v0.29 段）。**挂接时机=装配层推进总闸开启时**（--advance-interval>0）——总闸关断=
// 不挂接=事件拍缺席（事件驱动源不旁路总闸：推进=一个行为面一个总闸，关断语义单一可预期；
// 巡检-only 部署行为面零翻转）。事件拍的三重闸另两者在 hub.record 内（落了 run ∧
// source ∈ beatSources 拍面驱动源集合——v0.29 core 巡检、v0.30 connector watch 桥）。幂等；
// nil=解除挂接。须在 AttachDispatch 前调用（同 Bind* 口径——事件
// 开始流入后无接线窗口）。sink 契约见 EventAdvanceSink（非阻塞、不回调 hub）。
func (h *TriggerHub) BindEventAdvance(sink EventAdvanceSink) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.advance = sink
}

// execBindingLocked 目标 → 执行面绑定（纯映射 execFor + hub 接线态覆写）。
func (h *TriggerHub) execBindingLocked(target string) ExecutorBinding {
	b := execFor(target)
	if b.Executor != "" && h.wired[b.Executor] {
		b.Wired = true
	}
	return b
}

// Resolve 派发决策查询（引擎主体第一段 v0.13 命中行解析 → 第二段 v0.14 增执行绑定投影）。
// 纯决策面：不注入事件、不落留痕、不执行派发——对当前表快照回答「这一事件若到达会
// 路由到哪、由谁执行」。与留痕同一解析口径（resolveDecisionLocked），HTTP 投影见
// /v1/triggers/resolve（execution 区）。
func (h *TriggerHub) Resolve(typ, involved string) DispatchDecision {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.resolveDecisionLocked(typ, involved)
}

// resolveDecisionLocked 按 dispatch 表快照解析命中行（表序首行胜出；typeMatch ∧ forMatch）。
// owner 解析恒先行（与行无关的决策依据）；owner:<字面量> 谓词随解析激活（v0.13）；
// route/cc 目标随行投影执行绑定（v0.14）。
func (h *TriggerHub) resolveDecisionLocked(typ, involved string) DispatchDecision {
	owner := ownerOf(involved)
	if len(h.rules) == 0 {
		return DispatchDecision{Matched: false, Owner: owner, Row: NoRow} // 未挂表=无决策
	}
	for i, r := range h.rules {
		if typeMatch(r.On, typ) && forMatch(r.For, involved) {
			return DispatchDecision{Matched: true, Route: r.Dispatch, CC: r.CC, For: r.For, Owner: owner, Row: i,
				Exec: h.execBindingLocked(r.Dispatch), CCExec: h.execBindingLocked(r.CC)}
		}
	}
	return DispatchDecision{Matched: false, Route: DeadLetterRoute, Owner: owner, Row: NoRow,
		Exec: h.execBindingLocked(DeadLetterRoute)} // 合成死信=none 绑定；无 cc=零值绑定
}

// execFor route/cc 目标 → 执行面绑定（映射表，引擎主体第二段 v0.14；纯函数——接线态由 hub 覆写）。
// 词汇表按**家族**收编（通用性：新目标落已知家族即自动映射；扩展性：新家族在此登记一行；
// 解耦：映射只投影不执行，真实拉起=引擎主体余量）：
//
//	dead-letter             → none         死信无执行，观测兜底可见（event.md rule 4）
//	notification/<channel>  → notification 通道投影（*=全部已绑通道）
//	workflows/<name>        → workflow     直接 workflow 引用（deploy-system-workflow 等）
//	<p>-agent（p≠空）        → agent        agent 唤醒面（parent、root 等；v0.18 接线）
//	<p>-workflow（p≠空）     → workflow     owner 对账 workflow（p=owner 锚语义，如 owner-instance）
//	其余（含 observation-bus/instance-docs 等未设计执行面的目标族）→ unbound 原样投影（缺口可见）
func execFor(target string) ExecutorBinding {
	switch {
	case target == "":
		return ExecutorBinding{} // 无目标（空 cc）=零值绑定
	case target == DeadLetterRoute:
		return ExecutorBinding{Executor: ExecNone}
	}
	if ch, ok := strings.CutPrefix(target, "notification/"); ok && ch != "" {
		return ExecutorBinding{Executor: ExecNotification, Target: ch}
	}
	if name, ok := strings.CutPrefix(target, "workflows/"); ok && name != "" {
		return ExecutorBinding{Executor: ExecWorkflow, Target: name}
	}
	if p, ok := strings.CutSuffix(target, "-agent"); ok && p != "" {
		return ExecutorBinding{Executor: ExecAgent, Target: p}
	}
	if p, ok := strings.CutSuffix(target, "-workflow"); ok && p != "" {
		return ExecutorBinding{Executor: ExecWorkflow, Target: p}
	}
	return ExecutorBinding{Executor: ExecUnbound, Target: target}
}

// typeMatch 行 on 与事件类型：通配恒命中，其余等值。
func typeMatch(on, typ string) bool {
	return on == Wildcard || on == typ
}

// forMatch for 谓词口径（event.md 设计规则 6；谓词=any|<selector>:<pattern>）：
//
//	any            → 恒命中；
//	owner:*        → 恒命中（任意 owner）；owner:<字面量> → owner 解析激活（v0.13）：
//	                 解析值=字面量即命中（解析模型=involved scope 根 systems/<s>→<s>，
//	                 global/builtin 同名=scope 根本身）；解析不等 → 不命中，事件顺延后行；
//	scope:<root>/* → involved scope 为 <root> 或其子路径（builtin/* 命中 builtin）；
//	                 scope:<root>（无 /*）→ 等值；
//	其他/畸形      → 不命中（顺延）。
//
// involved 投影 = kind/scope/name（apply objKey 同构）；scope 是路径（global、builtin、
// systems/<s>——layout.md「路径=scope」），含 systems/<s> 层级时按已知根并段重组。
func forMatch(pred, involved string) bool {
	switch {
	case pred == "any":
		return true
	case pred == "owner:"+Wildcard:
		return true
	}
	sel, pat, ok := strings.Cut(pred, ":")
	if !ok {
		return false // 畸形谓词不命中（顺延后行）
	}
	switch sel {
	case "owner":
		return ownerOf(involved) == pat // owner 字面量解析（引擎主体第一段 v0.13）
	case "scope":
		scope := involvedScope(involved)
		if root, hasGlob := strings.CutSuffix(pat, "/*"); hasGlob {
			return scope == root || strings.HasPrefix(scope, root+"/")
		}
		return scope == pat
	}
	return false
}

// ownerOf owner 解析（event.md 规则 6）：involved scope 根 `systems/<s>` → `<s>`；
// global/builtin 同名（owner=scope 根本身）。involved 投影=kind/scope/name，
// 段数不足（无 scope 段）时解析为空串（不命中任何 owner 字面量谓词）。
func ownerOf(involved string) string {
	scope := involvedScope(involved)
	if s, ok := strings.CutPrefix(scope, "systems/"); ok {
		return s
	}
	return scope
}

// involvedScope 从 involved 投影重组 scope（kind 后首段；systems/<s> 层级并段）。
// 段数不足时尽力而为（返回已得部分或空串）。
func involvedScope(involved string) string {
	segs := strings.Split(involved, "/")
	if len(segs) < 2 {
		return ""
	}
	if segs[1] == "systems" && len(segs) >= 3 {
		return segs[1] + "/" + segs[2] // systems/<s> 层级 scope
	}
	return segs[1]
}

// State 查询快照（拷贝返回，调用方可安全持有；records 旧→新）。
func (h *TriggerHub) State() TriggerState {
	h.mu.Lock()
	defer h.mu.Unlock()
	records := make([]TriggerRecord, len(h.ring))
	copy(records, h.ring)
	types := append([]string{}, h.types...)
	return TriggerState{Types: types, Received: h.received, Dropped: h.dropped, Records: records}
}

func (h *TriggerHub) hasTypeLocked(typ string) bool {
	for _, t := range h.types {
		if t == typ {
			return true
		}
	}
	return false
}
