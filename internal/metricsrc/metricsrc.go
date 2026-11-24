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
// 热重绑（v0.47，event.md 规则 13 注记；console-adoption.md §1.18 R-4）：**绑定表与样本流
// 解耦**——样本流是桥级单通道（装配层一条订阅经泵适配直投，进程生命周期常驻），绑定表是
// 可整体替换的纯声明（Source 无通道字段）。per-source 通道形态下重绑=通道级联重建（订阅
// 泄漏/泵 goroutine 阻塞弃收），单流+绑定表=Rebind 只换声明不碰流、O(1) 零通道抖动；事件
// 产出语义面零变化（同判定谓词/同 dedupKey/同 involved/同主干——多绑定同键越限由总线窗口
// 聚合，与原多源各自发布同态）。**零绑定=合法等待态**：桥挂流即运行，装配期零声明不再是
// 不启动的终态（no-source 退场为 on(0)），apply 落库（DocApplied×WorkService）驱动
// StartRebind 全量重枚举热激活——阈值声明面 apply 落库即生效免重启。
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

// Source 一路阈值绑定声明（v0.47 热重绑后的纯声明形态：无数据通道字段——样本流是桥级
// 单通道，绑定表可整体替换。装配层绑定声明：该绑定的观测溯源锚+involvedRef 投影+阈值清单）。
type Source struct {
	Name       string      // 绑定名（观测溯源锚；现装配=WorkService 对象名）
	Involved   string      // involvedRef 投影（路由 owner 解析锚；逐对象=WorkService/<scope>/<name> 三段同构）
	Thresholds []Threshold // 阈值声明（服务侧判定依据）
}

// Bridge metrics 阈值事件桥（cored 装配层持有；零值不可用，经 New 构建）。绑定表与样本流
// 解耦（v0.47）：样本流单通道常驻，绑定表经 Rebind 整体替换（并发安全——判定与换表锁互斥）。
type Bridge struct {
	bus     *eventbus.Bus
	sch     *scheme.Scheme
	samples <-chan Sample // 桥级单通道（装配层泵适配直投；关频道=桥退场——可重启语义由装配层决定）

	mu   sync.Mutex
	srcs []Source
}

// New 构建 metrics 阈值桥。samples 为装配层接线的桥级样本通道（装配层持数据面与信号面经
// 此适配——两包零互依）；srcs 为初始绑定表（可为空——v0.47 起零绑定=合法等待态，热重绑激活）。
func New(bus *eventbus.Bus, sch *scheme.Scheme, samples <-chan Sample, srcs []Source) *Bridge {
	return &Bridge{bus: bus, sch: sch, samples: samples, srcs: srcs}
}

// Enabled 桥是否开启：样本流已挂 ∧ 类型已注册（注册表闸门前置——未注册即不启动，声明先行的
// 机械落点）。绑定表可空（等待态不算关断——热重绑可激活，v0.47）。调用方据 stderr 报告开启/未开启因由。
func (b *Bridge) Enabled() bool {
	if b == nil || b.bus == nil || b.sch == nil || b.samples == nil {
		return false
	}
	_, err := b.sch.LookupEventType(EventType)
	return err == nil
}

// Sources 当前绑定表深度（stderr 装配锚报告用；Rebind 随行变化——装配报告是初值快照）。
func (b *Bridge) Sources() int {
	if b == nil {
		return 0
	}
	b.mu.Lock()
	defer b.mu.Unlock()
	return len(b.srcs)
}

// Rebind 热重绑：绑定表整体替换（v0.47——权限热装载同范式：声明面变更→快照整体替换，
// 不半装载）。样本流不动（数据面通道与判定面解耦的机械落点）；零绑定=合法（等待后续
// apply 热激活）。并发安全：与 Run 的判定路径锁互斥。
func (b *Bridge) Rebind(srcs []Source) {
	b.mu.Lock()
	b.srcs = srcs
	b.mu.Unlock()
}

// Run 桥主循环：桥级单通道排干，逐样本对**当前绑定表**全量判定（每绑定独立判定独立发布
// ——同键越限多绑定由总线窗口聚合，与多源各自发布同态；ctx 取消即退场，幂等可重启；数据侧
// 关频道=桥退场，可重启语义由装配层决定）。未开启（Enabled=false）即返——关断=零事件
// （机械判别器的关断面）。
func (b *Bridge) Run(ctx context.Context) {
	if !b.Enabled() {
		return
	}
	for {
		select {
		case <-ctx.Done():
			return
		case smp, ok := <-b.samples:
			if !ok {
				return
			}
			b.judgeAll(smp)
		}
	}
}

// judgeAll 单样本对当前绑定表全量判定（锁内快照——Rebind 换表与判定互斥不撕读；发布失败
// 不中断判定——可丢语义，event.md §1：事件可丢、可重、可乱序，系统正确性不依赖它；下一拍
// 同因 dedupKey 聚合兜底）。
func (b *Bridge) judgeAll(smp Sample) {
	b.mu.Lock()
	srcs := b.srcs
	b.mu.Unlock()
	for _, s := range srcs {
		_ = b.publish(s, smp)
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
