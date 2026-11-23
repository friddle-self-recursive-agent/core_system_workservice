// run.go —— 对账 run 投影面（引擎主体第三段 v0.15——真实拉起第一子段：workflow 触发面接线）。
// 事件语义锚（concepts/event.md）：「Event 只触发，不修改：收到事件的反应永远是触发一次
// Workflow 对账」——本文件把「触发一次对账」落为**进程内可查的 run 留痕**：route 目标执行面
// =workflow 且装配层已登记接线（TriggerHub.WireExecutor(ExecWorkflow)）时，触发骨架把该次
// 决策推进为一条 RunRecord（一次对账的起点投影），查询面见 GET /v1/workflows/runs。
//
// 相位推进器首段（v0.23——迁移判定=声明序列单调前进，workflows 规则 12）：run 受理即启动
// 信号——引擎在受理路径上执行**单步迁移判定**：builtin run 从声明首相位沿声明相位序列单调
// 前进一步（Idle→Analyzing），status 词形翻转 triggered→running（触发已受理→首步进行中）。
// 迁移判定为纯函数 advancePhase（序列即声明：builtin 相位序列为线性声明，concepts/workflow.md
// 「相位机=标准 FSM，序即声明」——显式 Transition 表/挂起-恢复/异步推进属推进器后续段，
// 换实现不动本查询面词汇）；object 源 run 相位不投影=无从迁移，status 恒 triggered；
// 已居序列尾（Failed 等终点）不再前进。受理与首步推进同拍完成（同步路径——真实异步派发
// deferred，投影面只见推进后形态）。
//
// 诚实投影边界不变：run 出现 ≠ 对账执行完毕——status=running 只证明引擎已把 run 推进到
// 声明序列第二相位（Analyzing 进行中），后续相位推进/刹车/冷却/minLevel 仍属引擎主体余量
//（cored-bridge.md §2.3）。
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

// run 状态词汇（v0.15 起点 triggered；v0.23 相位推进器首段扩枚举 running）。
const (
	RunTriggered = "triggered" // 触发已受理（object 源 run 的终点态——相位不投影无从迁移）
	RunRunning   = "running"   // 引擎已把 run 从声明首相位沿序列推进一步（首步进行中；非执行完毕）
)

// spec 来源词汇（run 锚解析的声明来源投影）。
const (
	SpecBuiltin = "builtin" // 内置注册表命中（相位序列随代码声明，可投影）
	SpecObject  = "object"  // 平台对象引用（Workflow 对象，spec 随文档面；run 只留锚）
)

// ReconcileWorkflow 对账锚的循环本体锚（owner 对账 workflow → core-workflow）。
// 依据：concepts/workflow.md——core-workflow=系统级循环（触发→整理→处理→回归），
// 即「收到事件的反应=触发一次 Workflow 对账」的那个对账循环。
const ReconcileWorkflow = "core-workflow"

// RunRecord 一条对账 run 投影（触发面→workflow 域的执行留痕；观测面非账本，
// 账本=event_log 归档，ops-log.md §2）。
type RunRecord struct {
	Run        int64     // run 序号（本 sink 进程内单调递增，1 起）
	Workflow   string    // workflow 锚（对账锚解析=core-workflow；直接引用=按名）
	SpecSource string    // 锚的声明来源（builtin=内置注册表命中；object=平台对象引用；空=未解析）
	Phases     []string  // 声明相位序列（builtin 命中才投影；object 引用 spec 不在本进程=空）
	Phase      string    // 当前相位投影（v0.23 迁移判定首段：builtin run 受理即沿声明序列前进一步=第二相位；object 源恒空）
	Status     string    // run 状态（builtin=RunRunning 引擎已推进首步；object=RunTriggered 受理态）
	TriggerSeq int64     // 来源触发留痕 seq（可回指 GET /v1/triggers records）
	EventID    string    // 与发布方回执同源（apply 响应 event.id / event_log.event_id）
	Type       string    // 事件类型
	DedupKey   string    // 与发布方回执同源
	Route      string    // 命中行 route 原文（owner-workflow / workflows/<name> 等锚的出处）
	Owner      string    // owner 解析值（对账锚的 owner 维度在此承载）
	Target     string    // 执行面锚点投影（=ExecutorBinding.Target）
	StartedAt  time.Time // run 起始时刻（收到留痕同刻）
}

// RunState 查询快照（RunLog.State() 返回值；apiserver GET /v1/workflows/runs 载荷来源）。
type RunState struct {
	Started int64       // 累计 start（含已被环挤出的）
	Dropped int64       // 环溢出丢弃计数（可观测）
	Runs    []RunRecord // 环内 run 投影（旧→新）
}

// RunSink workflow 触发面的 run 投影落点（引擎主体第三段 v0.15）。
// 解耦：hub 只把「route 目标执行面=workflow ∧ 已接线」的留痕推进给 sink，不感知 run 内部
// （相位态机/刹车/冷却）；扩展性：真实相位态机落地换实现不动 hub。实现须**非阻塞**（hub 在
// 订阅路径上同步调用）且**不得回调 hub**（进程内投影面；真实异步派发落地时由实现自行解耦）。
type RunSink interface {
	StartRun(rec TriggerRecord, b ExecutorBinding)
}

// RunOptions run 投影环可调项（零值可用）。
type RunOptions struct {
	// RingSize run 环容量（缺省 256）：溢出丢最旧、dropped 计数可见。
	RingSize int
}

// RunLog run 投影的进程内落点（RunSink 缺省实现；并发安全）。
type RunLog struct {
	opts RunOptions

	mu      sync.Mutex
	seq     int64
	ring    []RunRecord
	started int64
	dropped int64
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
	wf, source, phases := resolveRunAnchor(rec.Route)
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
		// 相位推进器首段（v0.23）：受理即启动信号——引擎单步迁移判定，沿声明序列
		// 前进一步并把状态词翻转为 running（advancePhase 纯函数，判定口径见包注释）。
		if next, ok := advancePhase(phases, phase); ok {
			phase, status = next, RunRunning
		}
	}
	if phases == nil {
		phases = []string{} // object 源不投影相位：载荷形状归一为空数组（null=缺席语义，[]=声明为空——消费方可 len）
	}
	r.ring = append(r.ring, RunRecord{
		Run: r.seq, Workflow: wf, SpecSource: source, Phases: phases,
		Phase: phase, Status: status,
		TriggerSeq: rec.Seq, EventID: rec.EventID, Type: rec.Type, DedupKey: rec.DedupKey,
		Route: rec.Route, Owner: rec.Owner, Target: b.Target, StartedAt: rec.ReceivedAt,
	})
}

// State 查询快照（拷贝返回，调用方可安全持有；runs 旧→新）。
func (r *RunLog) State() RunState {
	r.mu.Lock()
	defer r.mu.Unlock()
	runs := make([]RunRecord, len(r.ring))
	copy(runs, r.ring)
	return RunState{Started: r.started, Dropped: r.dropped, Runs: runs}
}

// advancePhase 相位迁移判定首段（v0.23 纯函数）：沿声明相位序列从当前相位单调前进一步。
// 迁移规则=序列前进（builtin 相位序列为线性声明，concepts/workflow.md「相位机=标准 FSM，
// 序即声明」——触发受理即启动信号；显式 Transition 表/挂起-恢复属推进器后续段，落此扩展）。
// 已居序列尾（终点相位）、当前相位不在序列内、序列为空 → 不前进（ok=false，受理态保留）。
func advancePhase(phases []string, cur string) (string, bool) {
	for i, p := range phases {
		if p == cur && i+1 < len(phases) {
			return phases[i+1], true
		}
	}
	return cur, false
}

// resolveRunAnchor 命中行 route 原文 → run 投影的 workflow 锚（纯映射；引擎主体第三段 v0.15）。
// 口径见包注释；直接引用按 builtin 注册表解析，未命中=平台对象引用（SpecObject，锚保留、
// 相位不投影——对象 spec 随文档面，不在本进程）。
func resolveRunAnchor(routeTarget string) (wf, source string, phases []string) {
	if name, ok := strings.CutPrefix(routeTarget, "workflows/"); ok && name != "" {
		if IsBuiltin(name) {
			spec := Builtin(name)
			return spec.Name, SpecBuiltin, phaseNames(spec.Phases)
		}
		return name, SpecObject, nil // 平台对象引用：锚可见，spec 随文档面
	}
	if _, ok := strings.CutSuffix(routeTarget, "-workflow"); ok && routeTarget != "-workflow" {
		spec := Builtin(ReconcileWorkflow) // 对账锚 → 对账循环锚 core-workflow（owner 维度在留痕）
		return spec.Name, SpecBuiltin, phaseNames(spec.Phases)
	}
	return "", "", nil // 防御性：非 workflow 家族原文原样可见（空锚）
}

// phaseNames 相位序列投影为字符串切片（声明面 → 留痕面；顺序保留）。
func phaseNames(ps []Phase) []string {
	out := make([]string, len(ps))
	for i, p := range ps {
		out[i] = string(p)
	}
	return out
}

// 编译期断言：RunLog 满足 RunSink。
var _ RunSink = (*RunLog)(nil)
