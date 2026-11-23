// trigger.go —— 触发骨架（dispatch 接线第一段，v0.11 起点①）：
// 订阅总线事件流，收到即落**进程内可查状态**（收到留痕：事件 id/dedupKey/type/at/receivedAt）。
// 「apply→事件→订阅」是 core 主干第一段；触发引擎主体（dispatch 表路由→owner 解析→
// 相位态机→派发→刹车→冷却）declared-deferred（cored-bridge.md §2.3）——
// 本骨架只保证主干第一段机械可验：事件被订阅方收到、留痕可查询、对账锚点（eventId/dedupKey）
// 与发布方回执同源。
package workflow

import (
	"sync"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
)

// TriggerRecord 一条收到留痕（订阅方视角；事件本体语义见 concepts/event.md——
// 事件只触发不修改、可丢可重可乱序，留痕同样只是观测面）。
type TriggerRecord struct {
	Seq        int64     // 收到序号（本 hub 进程内单调递增，1 起）
	EventID    string    // 与发布方回执同源（apply 响应 event.id / event_log.event_id）
	Type       string    // 事件类型（Scheme.eventTypes 之一；暂为固定订阅集合）
	DedupKey   string    // 与发布方回执同源（apply:kind/scope/name:rv）
	Source     string    // 来源（apply/connector/metrics/timer/manual）
	Involved   string    // involvedRef 投影（kind/scope/name）
	Level      string    // debug/info/warn/error/fatal（A42）
	At         time.Time // 事件时间（发布方时钟）
	ReceivedAt time.Time // 收到时刻（订阅方时钟；与 At 差值=总线+调度延迟观测点）
}

// TriggerOptions 触发骨架可调项（零值可用）。
type TriggerOptions struct {
	// RingSize 留痕环形容量（缺省 256）：溢出丢最旧、dropped 计数可见。
	// 留痕是观测面不是账本（账本=event_log 归档，ops-log.md §2）。
	RingSize int
}

// TriggerState 查询快照（State() 返回值；apiserver GET /v1/triggers 的载荷来源）。
type TriggerState struct {
	Types    []string        // 已订阅事件类型（Attach 集合；暂固定 [DocApplied]）
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

// Attach 订阅事件类型（每类型一条总线订阅；同类型重复 Attach 去重）。
// 返回总退订函数（幂等）：退订全部订阅，各订阅 goroutine 排干存量后退场。
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
		ch, cancel := bus.Subscribe(typ)
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

// drain 订阅 goroutine：总线退订关频道后 range 排干存量退出（已入队事件不丢）。
func (h *TriggerHub) drain(ch <-chan eventbus.Event) {
	for ev := range ch {
		h.record(ev)
	}
}

// record 收到留痕（环满丢最旧，dropped 计数可见——观测面非账本）。
func (h *TriggerHub) record(e eventbus.Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.seq++
	h.received++
	if excess := len(h.ring) + 1 - h.opts.RingSize; excess > 0 {
		h.ring = h.ring[excess:]
		h.dropped += int64(excess)
	}
	h.ring = append(h.ring, TriggerRecord{
		Seq: h.seq, EventID: e.ID, Type: e.Type, DedupKey: e.DedupKey,
		Source: e.Source, Involved: e.Involved, Level: e.Level,
		At: e.At, ReceivedAt: time.Now(),
	})
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
