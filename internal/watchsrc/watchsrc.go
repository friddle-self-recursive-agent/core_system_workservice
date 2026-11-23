// Package watchsrc connector watch 事件桥（v0.30 引擎内部源第三段）：装配层把驱动侧
// connectors.ProviderImpl.Watch 契约通道桥接进总线——event.md §1 五源之首「Connector
// watch 外部变化」的机读落点，event.md 规则 12。与 patrol.Ticker/ApprovalNeeded 同范式：
// 类型注册闸门前置（Scheme.eventTypes 唯一开关，未注册即不启动）+ source 服务侧判定恒
// connector（来源标注是审计事实——驱动载荷不携带来源）+ level=注册表 defaultLevel 投影 +
// 与注入面同一条主干零旁路（窗口去重/订阅派发/event_log 归档全在 eventbus.Publish 内）。
// 依赖方向（R-ENG-3）：驱动仓（用法型 connector）→ connectors 契约层 → 本桥只消费契约
// 通道——驱动仓零总线依赖，总线面全部收敛在服务侧（cored 不 import 连接器仓的既有口径
// 不被破坏：注册表枚举的是本仓已 import 的驱动实现）。
// 缺省关断（装配旗标关断=零驱动零事件）：缺省不改既有行为面（规则 10 保险丝同款）。
// 事件拍组合：watch 事件落 run 同拍即拍（workflow.md 规则 15 v0.30 段——拍面驱动源
// 集合扩展，零新开关）。
package watchsrc

import (
	"context"
	"sync"

	"github.com/friddle-self-recursive-agent/core_system_workservice/connectors"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// watch 事件声明常量（与 scheme.yaml eventTypes ExternalChanged 行同源；event.md 规则 12）。
const (
	// EventType 外部变化事件类型（注册表闸门键）。
	EventType = "ExternalChanged"
	// EventSource 服务侧判定恒 connector（外部变化来源标注是审计事实——驱动载荷不携带来源）。
	EventSource = "connector"
)

// Source 一路桥接源（装配层绑定声明：驱动侧通道+挂载锚）。
type Source struct {
	Name     string                  // 源名（观测溯源锚；v0.30 首段=驱动类型名）
	Involved string                  // involvedRef 投影（路由 owner 解析锚；v0.30 首段=WorkService/<type>）
	Events   <-chan connectors.Event // 驱动侧外部变化通道（ProviderImpl.Watch 契约）
}

// Bridge watch 事件桥（cored 装配层持有；零值不可用，经 New 构建）。
type Bridge struct {
	bus  *eventbus.Bus
	sch  *scheme.Scheme
	srcs []Source
}

// New 构建 watch 桥。srcs 为装配层绑定声明（驱动通道已在装配侧经 ProviderImpl.Watch 取得）。
func New(bus *eventbus.Bus, sch *scheme.Scheme, srcs []Source) *Bridge {
	return &Bridge{bus: bus, sch: sch, srcs: srcs}
}

// Enabled 桥是否开启：有源 ∧ 类型已注册（注册表闸门前置——未注册即不启动，声明先行的
// 机械落点）。调用方据 stderr 报告开启/未开启因由。
func (b *Bridge) Enabled() bool {
	if b == nil || b.bus == nil || b.sch == nil || len(b.srcs) == 0 {
		return false
	}
	_, err := b.sch.LookupEventType(EventType)
	return err == nil
}

// Sources 桥接源数（stderr 装配锚报告用）。
func (b *Bridge) Sources() int {
	if b == nil {
		return 0
	}
	return len(b.srcs)
}

// Run 桥主循环：每源一条 goroutine 排干驱动通道（ctx 取消即退场，幂等可重启；驱动侧
// 关频道=该源退场，其余源不受影响）。未开启（Enabled=false）即返——关断=零事件
// （机械判别器的关断面）。
func (b *Bridge) Run(ctx context.Context) {
	if !b.Enabled() {
		return
	}
	var wg sync.WaitGroup
	for _, s := range b.srcs {
		wg.Add(1)
		go func(s Source) {
			defer wg.Done()
			b.drain(ctx, s)
		}(s)
	}
	wg.Wait()
}

// drain 单源排干：驱动事件逐条桥接发布（发布失败不中断排干——可丢语义，event.md §1：
// 事件可丢、可重、可乱序，系统正确性不依赖它；下一拍驱动重报同因由 dedupKey 聚合兜底）。
func (b *Bridge) drain(ctx context.Context, s Source) {
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-s.Events:
			if !ok {
				return
			}
			_ = b.publish(s, ev)
		}
	}
}

// publish 服务侧桥接：类型闸门 → 事件构造（source/level 服务侧判定——审计事实不是驱动
// 参数；dedupKey=驱动侧因由声明透传——同因聚合防外部抖动风暴，规则 1/§3；detail 不进
// 总线——规则 2 事件只带指针，消费者被叫醒后自己看现实）→ 总线发布（同主干：窗口去重/
// 订阅派发/event_log 归档）。
func (b *Bridge) publish(s Source, ev connectors.Event) error {
	et, err := b.sch.LookupEventType(EventType)
	if err != nil {
		return err // 闸门前置：未注册不发布（Enabled 已挡，防御性再断）
	}
	e := eventbus.Event{
		Type:     EventType,
		Source:   EventSource,
		Involved: s.Involved,
		DedupKey: ev.DedupKey,     // 驱动侧因由声明（空=永不聚合，eventbus 既有口径）
		Level:    et.DefaultLevel, // 注册表缺省投影（ExternalChanged→info）；空值由总线归一 info
	}
	return b.bus.Publish(&e)
}
