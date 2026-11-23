// Package metricsrc metrics 阈值事件桥（v0.31 引擎内部源第四段）：装配层把数据面指标
// 样本经阈值判定桥接进总线——event.md §5「MetricsInfo 阈值判断后发 MetricThreshold
// Event——数据面到信号面的桥」的机读落点（event.md 规则 13）。与 patrol.Ticker/
// watchsrc.Bridge 同范式：类型注册闸门前置（Scheme.eventTypes 唯一开关，未注册即不启动）
// + source 服务侧判定恒 metrics（来源标注是审计事实——数据面样本不携带来源）+ level=
// 注册表 defaultLevel 投影（MetricThreshold→warn）+ 与注入面同一条主干零旁路（窗口去重/
// 订阅派发/event_log 归档全在 eventbus.Publish 内）。
// 契约先行（声明面在位、数据面 declared-deferred）：MetricsInfo 数据面尚不存在，本包只落
// 「阈值判定桥契约」——数据源通道（Sample）+ 装配层阈值声明（Threshold）+ 服务侧判定；
// 真实样本生产者（MetricsInfo store/queries 投影）随数据面落地顺延接入，契约不变。
// dedupKey=<metric>:<window>（scheme.yaml MetricThreshold 行 dedup 声明的机械投影——
// 同指标同窗口的重复越限样本窗口聚合防风暴，规则 1；窗口标签由样本侧声明，服务侧透传
// 不改造）。阈值判定是服务侧动作：数据面只产样本不带判定（判定谓词属装配声明面，不属
// 数据载荷——与 source 恒服务侧判定同口径）。
// 缺省关断（装配旗标关断=零样本源零事件）：缺省不改既有行为面（规则 10 保险丝同款）。
// 事件拍组合：metrics 事件落 run 同拍即拍（workflow.md 规则 15——拍面驱动源集合扩展，
// 零新开关；v0.31 段 core+connector+metrics）。
package metricsrc

import (
	"context"
	"fmt"
	"sync"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// metrics 阈值事件声明常量（与 scheme.yaml eventTypes MetricThreshold 行同源；event.md 规则 13）。
const (
	// EventType 阈值越限事件类型（注册表闸门键）。
	EventType = "MetricThreshold"
	// EventSource 服务侧判定恒 metrics（来源标注是审计事实——数据面样本不携带来源）。
	EventSource = "metrics"
)

// Sample 数据面样本（数据侧经装配层通道声明——同一窗口聚合后的指标样本逐条投递；
// 真实生产者=MetricsInfo 数据面投影，declared-deferred，契约先行）。
type Sample struct {
	Metric string  // 指标名（阈值命中键；dedupKey 因由成分一）
	Value  float64 // 样本值（服务侧与阈值判定）
	Window string  // 聚合窗口标签（dedupKey 因由成分二——scheme.yaml dedup "<metric>:<window>" 投影）
}

// Threshold 阈值声明（装配层绑定锚——判定谓词在服务侧声明面，数据面只产样本不带判定）。
type Threshold struct {
	Metric string  // 指标名（与样本同键命中；声明序首个同键阈值参与判定——表序首个同 dispatch 口径）
	Op     string  // 比较谓词：">" / ">=" / "<" / "<="（最小集；声明面外谓词不判定不发布不即兴语义）
	Limit  float64 // 阈值
}

// Source 一路阈值桥接源（装配层绑定声明：数据侧通道+该源阈值清单）。
type Source struct {
	Name       string        // 源名（观测溯源锚；v0.31 首段=数据面投影名）
	Involved   string        // involvedRef 投影（路由 owner 解析锚；v0.31 首段=装配绑定锚 WorkService/metrics——逐对象 involved 声明面随 MetricsInfo 数据面落地顺延）
	Samples    <-chan Sample // 数据侧样本通道
	Thresholds []Threshold   // 阈值声明（服务侧判定依据）
}

// Bridge metrics 阈值事件桥（cored 装配层持有；零值不可用，经 New 构建）。
type Bridge struct {
	bus  *eventbus.Bus
	sch  *scheme.Scheme
	srcs []Source
}

// New 构建 metrics 阈值桥。srcs 为装配层绑定声明（数据侧样本通道已在装配侧取得；
// 真实数据面落地前装配层可零源绑定——no-source 缺口可见）。
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

// Run 桥主循环：每源一条 goroutine 排干数据侧通道（ctx 取消即退场，幂等可重启；数据侧
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

// drain 单源排干：数据侧样本逐条服务侧阈值判定（发布失败不中断排干——可丢语义，
// event.md §1：事件可丢、可重、可乱序，系统正确性不依赖它；下一拍同因 dedupKey 聚合兜底）。
func (b *Bridge) drain(ctx context.Context, s Source) {
	for {
		select {
		case <-ctx.Done():
			return
		case smp, ok := <-s.Samples:
			if !ok {
				return
			}
			_ = b.publish(s, smp)
		}
	}
}

// publish 服务侧阈值判定桥接：类型闸门 → 谓词判定（声明序首个同指标阈值参与判定——
// 表序首个同 dispatch 口径；声明面外谓词不判定不发布不即兴语义）→ 事件构造（source/
// level 服务侧判定——审计事实不是数据参数；dedupKey=<metric>:<window>——scheme.yaml
// dedup 声明的机械投影，同指标同窗口聚合防风暴，规则 1）→ 总线发布（同主干：窗口去重/
// 订阅派发/event_log 归档）。
func (b *Bridge) publish(s Source, smp Sample) error {
	et, err := b.sch.LookupEventType(EventType)
	if err != nil {
		return err // 闸门前置：未注册不发布（Enabled 已挡，防御性再断）
	}
	for _, th := range s.Thresholds {
		if th.Metric != smp.Metric {
			continue
		}
		if !judge(th, smp.Value) {
			return nil // 首个同键阈值未越限=该样本静默（数据面照常产样本零事件）
		}
		e := eventbus.Event{
			Type:     EventType,
			Source:   EventSource,
			Involved: s.Involved,
			DedupKey: fmt.Sprintf("%s:%s", smp.Metric, smp.Window), // <metric>:<window> 因由口径
			Level:    et.DefaultLevel,                              // 注册表缺省投影（MetricThreshold→warn）
		}
		return b.bus.Publish(&e)
	}
	return nil // 无同键阈值=该指标未装配判定面，样本静默
}

// judge 谓词判定（最小集；未知谓词=false——声明面外谓词不判定不发布不即兴语义）。
func judge(th Threshold, v float64) bool {
	switch th.Op {
	case ">":
		return v > th.Limit
	case ">=":
		return v >= th.Limit
	case "<":
		return v < th.Limit
	case "<=":
		return v <= th.Limit
	default:
		return false
	}
}
