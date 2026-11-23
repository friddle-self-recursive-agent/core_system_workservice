// run.go —— 对账 run 投影面（引擎主体第三段 v0.15——真实拉起第一子段：workflow 触发面接线）。
// 事件语义锚（concepts/event.md）：「Event 只触发，不修改：收到事件的反应永远是触发一次
// Workflow 对账」——本文件把「触发一次对账」落为**进程内可查的 run 留痕**：route 目标执行面
// =workflow 且装配层已登记接线（TriggerHub.WireExecutor(ExecWorkflow)）时，触发骨架把该次
// 决策推进为一条 RunRecord（一次对账的起点投影），查询面见 GET /v1/workflows/runs。
//
// 解耦边界（诚实投影）：run 记录=触发已受理 + 声明起点（builtin 锚的首相位），**不含相位态机
// 推进**（相位态机/刹车/冷却/minLevel 仍属引擎主体余量，cored-bridge.md §2.3）——run 出现
// ≠ 对账已执行完毕，只等于「触发面已把事件接进 workflow 域」。
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

// run 状态词汇（v0.15 只有起点；相位推进落地轮在此扩枚举）。
const (
	RunTriggered = "triggered" // 触发已受理（run 投影终点=相位态机 deferred）
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
	Phase      string    // 当前相位投影（run 起点=声明首相位；相位推进器余量——声明面已完备 v0.19，推进由引擎段承接）
	Status     string    // run 状态（v0.15 恒 RunTriggered）
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
	if len(phases) > 0 {
		phase = string(phases[0]) // 声明起点=首相位（相位推进器余量，起点后不推进——诚实投影口径不变）
	}
	r.ring = append(r.ring, RunRecord{
		Run: r.seq, Workflow: wf, SpecSource: source, Phases: phases,
		Phase: phase, Status: RunTriggered,
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
