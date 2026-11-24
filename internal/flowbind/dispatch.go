// dispatch.go —— apply-workflow agent 派单面（v0.50，R-8——workflow-runtime.md §8+§11.2）。
// 设计依据：instance 仓 runtime/workflow-runtime.md §8（workflow 落地=专用 agent
// apply-workflow，AUTOLEVEL=ask 默认）+ §11.1 落地形态二分（骨架+节点接线=cored 确定性
// 机械面；实现代码=apply-workflow agent 面）+ §11.2 拉起形态首段（骨架物化即派单受理投影）
// + agent-session.md §2（AUTOLEVEL：ask=「问不问」的数据面在册——批准/拉起动作面随真实
// 拉起段落地）。
//
// 触发接线：物化器一次真实物化动作（created|updated——骨架新建/更新，即「该版本实现未落」
// 的机械签名）→ 记一条对 apply-workflow 锚的派单受理投影；noop（幂等重入零差异）/skipped
// （非本物化器所有面）不派单——无新实现可做，不制造噪音。
//
// 诚实投影（wake.go 同构）：**派单受理 ≠ agent 会话已拉起**（真实拉起=dsh 会话挂载/
// agent-session 域，declared-deferred）——本面只证明「骨架→agent 派单接口面已接线」。
// 观测面非账本：持久态=骨架对象本体（聚合面 execution.skeleton=true）+ apply_log 审计
// （workflow:bind 行）；本面=进程内受理投影（环有界，重启归零=幂等重入不重复派单的
// 语义锚——重启后未落实现仍由 skeleton=true 持久可见，不因投影丢失而缺口静默）。
//
// 解耦口径：本件只依赖 apply.Result 投影，零新写路径零旁路；sink 由装配层注入
// （nil=静默不派单——headless 载体/旧装配零影响）。
package flowbind

import (
	"sync"
	"time"
)

// 派单锚与词汇（workflow-runtime.md §8：builtin 派单锚+AUTOLEVEL=ask 默认；扩展=Agent
// 文档面 spec.autoLevel 读面落地时改读声明——当前 builtin 常量，声明面外不即兴）。
const (
	// DispatchAgent 派单目标锚（builtin 专用 agent——concepts/workflow.md §1.2）。
	DispatchAgent = "apply-workflow"
	// DispatchAccepted 受理状态词（与 workflow.WakeAccepted 同语义族：受理投影，非拉起）。
	DispatchAccepted = "accepted"
	// DispatchAutoLevel 派单 autoLevel 数据（§8 默认 ask——批准/拉起动作面随真实拉起落地）。
	DispatchAutoLevel = "ask"
)

// DispatchRecord 一条派单受理投影（apply-workflow 锚——骨架物化→agent 派单接口面的受理留痕；
// 观测面非账本，账本=apply_log 归档+骨架对象本体）。
type DispatchRecord struct {
	Seq       int64     // 派单序号（本 sink 进程内单调递增，1 起）
	Agent     string    // 派单目标锚（apply-workflow）
	Status    string    // 受理状态（v0.50 恒 DispatchAccepted）
	AutoLevel string    // 派单 autoLevel（ask 默认——agent-session.md §2 级别数据面）
	Action    string    // 触发动作（created|updated——物化器 Ensure 动作；noop/skipped 不派单）
	Workflow  string    // workflow 锚名（实现对象）
	Scope     string    // workflow 对象 scope
	Connector string    // 骨架对象名（绑定 ConnectorOps——实现落点）
	RV        int64     // 骨架对象落库后 resourceVersion（实现侧 diff 基线）
	EventID   string    // 关联 DocApplied 回执 id（workflow apply 来源链——可回指 event_log）
	At        time.Time // 受理时刻
}

// DispatchSink 派单投影落点（装配层注入；实现须**非阻塞**且不得回调物化器——进程内投影面，
// WakeSink 同款约束）。nil=静默零派单。
type DispatchSink interface {
	Dispatch(rec DispatchRecord)
}

// DispatchOptions 派单投影环可调项（零值可用）。
type DispatchOptions struct {
	// RingSize 派单环容量（缺省 256）：溢出丢最旧、dropped 计数可见。
	RingSize int
}

// DispatchState 查询快照（DispatchLog.State() 返回值）。
type DispatchState struct {
	Accepted int64            // 累计受理（含已被环挤出的）
	Dropped  int64            // 环溢出丢弃计数（可观测）
	Records  []DispatchRecord // 环内派单投影（旧→新）
}

// DispatchLog 派单受理投影的进程内落点（DispatchSink 缺省实现；并发安全；与 workflow.WakeLog
// 同构——无出站机器，受理即终态投影）。
type DispatchLog struct {
	opts DispatchOptions

	mu       sync.Mutex
	seq      int64
	ring     []DispatchRecord
	accepted int64
	dropped  int64
}

// NewDispatchLog 构建派单投影落点。
func NewDispatchLog(opts *DispatchOptions) *DispatchLog {
	d := &DispatchLog{}
	if opts != nil {
		d.opts = *opts
	}
	if d.opts.RingSize <= 0 {
		d.opts.RingSize = 256
	}
	return d
}

// Dispatch 记一条派单受理投影（环满丢最旧，dropped 计数可见）。
func (d *DispatchLog) Dispatch(rec DispatchRecord) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.seq++
	rec.Seq = d.seq
	d.accepted++
	if excess := len(d.ring) + 1 - d.opts.RingSize; excess > 0 {
		d.ring = d.ring[excess:]
		d.dropped += int64(excess)
	}
	d.ring = append(d.ring, rec)
}

// State 查询快照（拷贝返回，调用方可安全持有；records 旧→新）。
func (d *DispatchLog) State() DispatchState {
	d.mu.Lock()
	defer d.mu.Unlock()
	rs := make([]DispatchRecord, len(d.ring))
	copy(rs, d.ring)
	return DispatchState{Accepted: d.accepted, Dropped: d.dropped, Records: rs}
}

// 编译期断言：DispatchLog 满足 DispatchSink。
var _ DispatchSink = (*DispatchLog)(nil)
