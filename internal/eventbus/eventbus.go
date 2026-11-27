// Package eventbus 事件队列：dedupKey 时间窗聚合、按类型订阅分发、死信可见。
// 设计依据：core_system/concepts/event.md——事件只触发不修改、可丢可重可乱序（P3 level-triggered）；
// v1 进程内最小实现（内存队列），dsh 侧持久队列/滚动归档随 connector 接入演进。
package eventbus

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventscope"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
)

// Event 事件（短生命周期数据，非文档对象；event.md §3 结构的进程内投影）。
type Event struct {
	ID       string
	Type     string
	Source   string
	Involved string // involvedRef：路由按 类型+involved/owner+scope（v0.11 A47，无 selector）
	At       time.Time
	DedupKey string
	Level    string // debug/info/warn/error/fatal（A42，rule 1b 强制）
	Scope    string // metadata.scope 四层词（BUG-010，event.md 规则 19——总线侧强制回填，空=按 involved 派生兜底）
	Run      int64  // 审批信号目标 run 序号（可选：0=按锚最新挂起 run；>0=定向 run——console 批闸行携带）
}

// Options 总线可调项（零值可用；测试注入小窗口/假 ops store）。
type Options struct {
	// DedupWindow 同 DedupKey 聚合窗口（缺省 5s；event.md rule 1 防风暴。
	// 正式数值随 Scheme.dispatch/Workflow 冷却期接入校准）。
	DedupWindow time.Duration
	// SubscriberBuffer 单订阅方频道缓冲（缺省 16；满=慢消费者，事件转死信——可丢语义）。
	SubscriberBuffer int
	// DeadLetterCap 死信环形容量（缺省 128；event.md rule 4 死信必须可见）。
	DeadLetterCap int
	// Ops ops DB 归档钩子（event_log）。nil=跳过。debug 级不入库（event.md rule 1b）。
	Ops opslog.Store
}

// Bus 事件总线（Core 内）。Publish/Subscribe 并发安全。
type Bus struct {
	opts Options

	mu      sync.Mutex
	lastDup map[string]time.Time // DedupKey → 窗口内最近发布时刻
	subs    []*subscription
	dead    []Event
}

type subscription struct {
	typ string // "" = 全类型
	ch  chan Event
}

// NewBus 构建总线。
func NewBus(opts *Options) *Bus {
	b := &Bus{lastDup: map[string]time.Time{}}
	if opts != nil {
		b.opts = *opts
	}
	if b.opts.DedupWindow <= 0 {
		b.opts.DedupWindow = 5 * time.Second
	}
	if b.opts.SubscriberBuffer <= 0 {
		b.opts.SubscriberBuffer = 16
	}
	if b.opts.DeadLetterCap <= 0 {
		b.opts.DeadLetterCap = 128
	}
	return b
}

// Subscribe 订阅事件流（typ 为空订阅全类型）。返回只读频道与退订函数（退订幂等，频道关闭）。
func (b *Bus) Subscribe(typ string) (<-chan Event, func()) {
	sub := &subscription{typ: typ, ch: make(chan Event, b.opts.SubscriberBuffer)}
	b.mu.Lock()
	b.subs = append(b.subs, sub)
	b.mu.Unlock()
	cancel := func() {
		b.mu.Lock()
		defer b.mu.Unlock()
		for i, s := range b.subs {
			if s == sub {
				b.subs = append(b.subs[:i], b.subs[i+1:]...)
				close(sub.ch)
				return
			}
		}
	}
	return sub.ch, cancel
}

// ErrRejected 事件被总线拒绝（缺关键字段；区别于可丢语义的聚合/死信）。
var ErrRejected = errors.New("eventbus: 事件缺 type")

// Publish 发布事件：窗口内同 DedupKey 聚合（返回 nil，不投递不归档）；命中订阅方投递；
// 无订阅方或频道满 → 死信（event.md rule 4）。返回 err 仅代表事件本身不可发布。
// e 传指针：缺省字段（ID/At/Level/DedupKey）就地归一化，调用方持有回执即可复读真实 ID。
func (b *Bus) Publish(e *Event) error {
	if e.Type == "" {
		return fmt.Errorf("%w: %+v", ErrRejected, e)
	}
	if e.At.IsZero() {
		e.At = time.Now()
	}
	if e.Level == "" {
		e.Level = "info"
	}
	if e.ID == "" {
		e.ID = fmt.Sprintf("evt-%s-%d", e.Type, e.At.UnixNano())
	}
	if e.DedupKey == "" {
		e.DedupKey = e.ID // 无显式 dedup 键=永不聚合
	}
	if e.Scope == "" {
		// metadata.scope 强制回填（BUG-010，event.md 规则 19①）：显式声明优先，缺席按
		// involvedRef 所属 scope 派生（服务侧回填，禁止无 scope 的野事件）；派生不出合法
		// 四层词（畸形 involved/空 involved）=兜底 global——保守词形不炸旧发布链。
		e.Scope = eventscope.FromInvolved(e.Involved)
	}

	b.mu.Lock()
	now := time.Now() // 窗口以墙钟计；e.At 只是事件时间戳（重放事件不重置其窗口判定）
	if last, ok := b.lastDup[e.DedupKey]; ok && now.Sub(last) < b.opts.DedupWindow {
		b.mu.Unlock()
		return nil // 窗口内同因事件合并（不投递不归档；防风暴）
	}
	b.lastDup[e.DedupKey] = now
	b.pruneDupLocked(now)

	routed := make([]*subscription, 0, len(b.subs))
	for _, s := range b.subs {
		if s.typ == "" || s.typ == e.Type {
			routed = append(routed, s)
		}
	}
	slow := 0
	if len(routed) == 0 {
		b.pushDeadLocked(*e)
	}
	// 投递与退订同锁互斥：退订关频道在锁内进行，投递（非阻塞 select）也在锁内——
	// 排除「路由快照已含订阅方、退订抢先关频道」的 send-on-closed 竞态。
	// 非阻塞投递锁内驻留有界（满即转死信，不等待消费者）。
	for _, s := range routed {
		select {
		case s.ch <- *e:
		default: // 慢消费者：转死信（可丢语义，event.md §1）
			slow++
		}
	}
	for range slow {
		b.pushDeadLocked(*e)
	}
	b.mu.Unlock()

	if b.opts.Ops != nil && e.Level != "debug" { // rule 1b：debug 只进内存队列
		routedTo := ""
		if len(routed) == 0 {
			routedTo = "dead-letter"
		} else {
			routedTo = fmt.Sprintf("%d subscriber(s)", len(routed))
		}
		_ = b.opts.Ops.AppendEvent(opslog.EventLog{
			Ts: e.At, EventID: e.ID, Type: e.Type, Source: e.Source,
			Involved: e.Involved, DedupKey: e.DedupKey, Severity: e.Level, RoutedTo: routedTo,
			Scope: e.Scope,
		})
	}
	return nil
}

// DeadLetters 路由不到/投递失败的事件（观测面健康指标；最多保留 DeadLetterCap 条新到事件）。
func (b *Bus) DeadLetters() []Event {
	b.mu.Lock()
	defer b.mu.Unlock()
	out := make([]Event, len(b.dead))
	copy(out, b.dead)
	return out
}

func (b *Bus) pushDeadLocked(e Event) {
	b.dead = append(b.dead, e)
	if excess := len(b.dead) - b.opts.DeadLetterCap; excess > 0 {
		b.dead = b.dead[excess:] // 丢旧保新：死信是健康信号不是账本（账本在 event_log 归档）
	}
}

// pruneDupLocked 清理窗口外的去重键（防长进程 map 膨胀）。
func (b *Bus) pruneDupLocked(now time.Time) {
	for k, t := range b.lastDup {
		if now.Sub(t) >= b.opts.DedupWindow {
			delete(b.lastDup, k)
		}
	}
}
