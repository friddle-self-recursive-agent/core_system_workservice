// wake.go —— agent 唤醒投影面（引擎主体第五段 v0.18——真实拉起第三子段：agent 触发面接线）。
// 事件语义锚（concepts/event.md 规则 6）：`<p>-agent` 族目标（root-agent/parent-agent——
// root=builtinRoot 根 Agent，parent=involved 对象 ownerRefs 的 ParentAgent）在执行面=agent
// 且装配层已登记接线（TriggerHub.WireExecutor(ExecAgent)）时，触发骨架把该绑定推进为一条
// WakeRecord（一次唤醒的受理投影），查询面见 GET /v1/agents/wakes。
//
// 解耦边界（诚实投影，与 run.go/deliver.go 同构）：唤醒记录=唤醒已受理 + agent 锚投影——
// **唤醒投影出现 ≠ agent 会话已被拉起**（真实拉起=agent-session 域/session spawn，属
// declared-deferred；本面只证明触发面把事件接进了 agent 域）。无出站机器（与 deliver.go
// 的差异点：送达面 v0.17 有出站注入，唤醒面暂无注入点——真实拉起落地轮在此扩接口）。
//
// route 与 cc 双路（与送达面同口径的声明）：run 投影只消费 route 主路（v0.15 口径），
// 唤醒投影同时消费 route 主路与 cc 旁路——绑定落在哪一路由 Origin 字段留痕，不为旁路
// 单开 sink 面（单 sink 双路收编；现行 scheme.dispatch 表 agent 族目标均在 route 主路：
// Escalate→root-agent、HeartbeatLost→parent-agent；cc-agent 为通用性预留，单测锚定）。
// 非 agent 执行面（workflow/notification/none）不产生唤醒投影。
package workflow

import (
	"sync"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/runtimestore"
)

// wake 状态词汇（v0.18 只有受理；真实拉起/session 语义落地轮在此扩枚举）。
const (
	WakeAccepted = "accepted" // 唤醒已受理（agent 锚投影在位；真实拉起 deferred）
)

// WakeRecord 一条唤醒投影（agent 族绑定→agent 域的受理留痕；观测面非账本，
// 账本=event_log 归档，ops-log.md §2）。
type WakeRecord struct {
	Wake       int64     // 唤醒序号（本 sink 进程内单调递增，1 起）
	Agent      string    // agent 锚投影（=ExecutorBinding.Target：root/parent 等 <p>-agent 前缀段）
	Origin     string    // 绑定来源（OriginRoute/OriginCC——route 主路/cc 旁路留痕）
	Status     string    // 唤醒状态（v0.18 恒 WakeAccepted）
	Level      string    // 事件等级（Escalate/HeartbeatLost 均 error——唤醒分诊观测点）
	TriggerSeq int64     // 来源触发留痕 seq（可回指 GET /v1/triggers records）
	EventID    string    // 与发布方回执同源（apply 响应 event.id / event_log.event_id）
	Type       string    // 事件类型
	DedupKey   string    // 与发布方回执同源
	Route      string    // 命中行 route 原文（root-agent / parent-agent 等锚的出处）
	Owner      string    // owner 解析值（parent 锚的 owner 维度在此承载）
	At         time.Time // 唤醒受理时刻（收到留痕同刻）
}

// WakeState 查询快照（WakeLog.State() 返回值；apiserver GET /v1/agents/wakes 载荷来源）。
type WakeState struct {
	Accepted int64        // 累计受理（含已被环挤出的）
	Dropped  int64        // 环溢出丢弃计数（可观测）
	Wakes    []WakeRecord // 环内唤醒投影（旧→新）
}

// WakeSink agent 触发面的唤醒投影落点（引擎主体第五段 v0.18）。
// 解耦：hub 只把「执行面=agent ∧ 已接线」的绑定推进给 sink，不感知 agent 域内部
// （真实拉起/session spawn/agent 池——agent-pool-machine.md 属余量）；扩展性：真实拉起
// 落地换实现不动 hub。实现须**非阻塞**（hub 在订阅路径上锁内同步调用）且**不得回调 hub**
// （进程内投影面）；origin 区分主/旁路。
type WakeSink interface {
	Wake(rec TriggerRecord, b ExecutorBinding, origin string)
}

// WakeOptions 唤醒投影环可调项（零值可用）。
type WakeOptions struct {
	// RingSize 唤醒环容量（缺省 256）：溢出丢最旧、dropped 计数可见。
	RingSize int
	// Store 运行态投影存储挂接（R-ENG §1 可选能力——nil=纯内存既有行为面零翻转）。挂接后
	// 写穿透（每次受理同步 Append——唤醒受理即终态，同 wake id 重放 noop）+ 启动恢复
	// （New 时装载历史行续接 seq/accepted 与环内最近投影）。读面零变化。
	Store runtimestore.Store
}

// WakeLog 唤醒投影的进程内落点（WakeSink 缺省实现；并发安全；与 RunLog 同构——
// 无出站机器，受理即终态投影）。
type WakeLog struct {
	opts WakeOptions

	mu       sync.Mutex
	seq      int64
	ring     []WakeRecord
	accepted int64
	dropped  int64
}

// NewWakeLog 构建唤醒投影落点。挂有 Store 时先做启动恢复（R-ENG §1——装载历史行填环内
// 最近 RingSize 条、seq/accepted 续接为行数口径）。
func NewWakeLog(opts *WakeOptions) *WakeLog {
	w := &WakeLog{}
	if opts != nil {
		w.opts = *opts
	}
	if w.opts.RingSize <= 0 {
		w.opts.RingSize = 256
	}
	if w.opts.Store != nil {
		rows := loadWakeRows(w.opts.Store)
		if len(rows) > 0 {
			wakes := make([]WakeRecord, len(rows))
			maxSeq := int64(0)
			for i, row := range rows {
				wakes[i] = wakeFromRow(row)
				if row.Wake > maxSeq {
					maxSeq = row.Wake
				}
			}
			w.seq = maxSeq
			w.accepted = int64(len(rows))
			if excess := len(wakes) - w.opts.RingSize; excess > 0 {
				wakes = wakes[excess:]
			}
			w.ring = wakes
		}
	}
	return w
}

// Wake 记一条唤醒投影（环满丢最旧，dropped 计数可见）。agent 锚投影取绑定锚点
// （root-agent→root、parent-agent→parent——<p>-agent 前缀段，execFor 家族映射产物）。
func (w *WakeLog) Wake(rec TriggerRecord, b ExecutorBinding, origin string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.seq++
	w.accepted++
	if excess := len(w.ring) + 1 - w.opts.RingSize; excess > 0 {
		w.ring = w.ring[excess:]
		w.dropped += int64(excess)
	}
	w.ring = append(w.ring, WakeRecord{
		Wake: w.seq, Agent: b.Target, Origin: origin, Status: WakeAccepted,
		Level: rec.Level, TriggerSeq: rec.Seq, EventID: rec.EventID,
		Type: rec.Type, DedupKey: rec.DedupKey, Route: rec.Route, Owner: rec.Owner,
		At: rec.ReceivedAt,
	})
	// 写穿透（R-ENG §1）：受理词形同步 Append（受理即终态——同 wake id 重放 noop）。
	if w.opts.Store != nil {
		persistAppend("唤醒受理", w.opts.Store.AppendWake(wakeToRow(w.ring[len(w.ring)-1])))
	}
}

// State 查询快照（拷贝返回，调用方可安全持有；wakes 旧→新）。
func (w *WakeLog) State() WakeState {
	w.mu.Lock()
	defer w.mu.Unlock()
	ws := make([]WakeRecord, len(w.ring))
	copy(ws, w.ring)
	return WakeState{Accepted: w.accepted, Dropped: w.dropped, Wakes: ws}
}

// 编译期断言：WakeLog 满足 WakeSink。
var _ WakeSink = (*WakeLog)(nil)
