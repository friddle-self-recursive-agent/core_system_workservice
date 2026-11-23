// trigger.go —— 触发骨架（dispatch 接线：第一段 v0.11 → 第二段 v0.12 订阅面动态化）：
// 订阅总线事件流，收到即落**进程内可查状态**（收到留痕：事件 id/dedupKey/type/at/receivedAt）。
// 「apply→事件→订阅」是 core 主干第一段；第二段把订阅集合从固定 [DocApplied] 泛化为
// scheme.dispatch 表驱动（on: 类型集合/通配 *，Attach 去重/总退订幂等）+ for 谓词首段过滤
// （any/owner:*/scope:<root>/*；owner 字面量解析随引擎主体）+ 死信兜底路由可见
// （全表未命中 → 留痕 route=dead-letter，event.md rule 4）。
// 触发引擎主体（owner 解析→相位态机→派发→刹车→冷却）declared-deferred（cored-bridge.md §2.3）——
// route 字段只做命中行投影留痕，不执行派发。
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
	Route      string    // 命中 dispatch 行的 route 投影（只留痕不派发，引擎主体 deferred）；
	//             全表未命中=dead-letter；未挂 dispatch 表（裸 Attach）=空
	For string // 命中行的 for 谓词原文（与 Route 同一行；便于机械对账表行↔留痕）
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
// 挂有 dispatch 表时按表序解析命中行：route/for 投影留痕（不派发）；
// 全表未命中 → route=dead-letter（死信兜底可见）；未挂表（裸 Attach）→ 留空。
func (h *TriggerHub) record(e eventbus.Event) {
	h.mu.Lock()
	defer h.mu.Unlock()
	h.seq++
	h.received++
	if excess := len(h.ring) + 1 - h.opts.RingSize; excess > 0 {
		h.ring = h.ring[excess:]
		h.dropped += int64(excess)
	}
	route, forSel := h.resolveLocked(e.Type, e.Involved)
	h.ring = append(h.ring, TriggerRecord{
		Seq: h.seq, EventID: e.ID, Type: e.Type, DedupKey: e.DedupKey,
		Source: e.Source, Involved: e.Involved, Level: e.Level,
		At: e.At, ReceivedAt: time.Now(),
		Route: route, For: forSel,
	})
}

// resolveLocked 按 dispatch 表快照解析命中行（表序首行胜出；typeMatch ∧ forMatch）。
func (h *TriggerHub) resolveLocked(typ, involved string) (route, forSel string) {
	if len(h.rules) == 0 {
		return "", ""
	}
	for _, r := range h.rules {
		if typeMatch(r.On, typ) && forMatch(r.For, involved) {
			return r.Dispatch, r.For
		}
	}
	return DeadLetterRoute, ""
}

// typeMatch 行 on 与事件类型：通配恒命中，其余等值。
func typeMatch(on, typ string) bool {
	return on == Wildcard || on == typ
}

// forMatch for 谓词首段口径（v0.12，event.md 设计规则；谓词=any|<selector>:<pattern>）：
//
//	any            → 恒命中；
//	owner:*        → 恒命中（任意 owner）；owner:<字面量> → 首段不含 owner 字面量解析
//	                 （随引擎主体 owner 解析接入）→ 不命中，事件顺延后行（死信兜底行可见）；
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
		return false // owner 字面量解析=引擎主体（declared-deferred），首段只放行 owner:*
	case "scope":
		scope := involvedScope(involved)
		if root, hasGlob := strings.CutSuffix(pat, "/*"); hasGlob {
			return scope == root || strings.HasPrefix(scope, root+"/")
		}
		return scope == pat
	}
	return false
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
