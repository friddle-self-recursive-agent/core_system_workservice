// Package patrol 定时巡检事件源（v0.22 引擎内部源第二段）：cored 定时器到点自产
// ReconcileTick 事件进总线——event.md §7 风险表缓解项「定时巡检兜底（level-triggered
// 的保险丝）：最迟 N 分钟全量对账一次」的机读落点，event.md 规则 10。
// 与 ApprovalNeeded（v0.21 首段）同范式：类型注册闸门前置（Scheme.eventTypes 是引擎
// 内部源的唯一开关，未注册即不启动）+ source 服务侧判定恒 core + level=注册表
// defaultLevel 投影 + 与注入面同一条主干零旁路（窗口去重/订阅派发/event_log 归档）。
// 差异：生产者是装配层定时器（零新 HTTP face，第三类真实来源），与写面开关分立
// （巡检是控制面保险丝，只读装配亦可开启）；dedupKey=reconcile:<at> 拍点即因由——
// 同拍重放窗口聚合、跨拍不聚合，最迟 N 分钟语义不被总线去重窗口吞没。
// 缺省关断（interval=0 零事件）：保险丝按部署显式开启，缺省不改既有行为面。
package patrol

import (
	"context"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// 巡检事件声明常量（与 scheme.yaml eventTypes/dispatch 行同源；event.md 规则 10）。
const (
	// EventType 巡检到点事件类型（注册表闸门键）。
	EventType = "ReconcileTick"
	// EventSource 服务侧判定恒 core（引擎自身产生的巡检来源标注是审计事实）。
	EventSource = "core"
	// EventInvolved 巡检对象=对账循环本体（owner 维度由留痕承载）。
	EventInvolved = "Workflow/builtin/core-workflow"
)

// Ticker 定时巡检源（cored 装配层持有；零值不可用，经 New 构建）。
type Ticker struct {
	bus      *eventbus.Bus
	sch      *scheme.Scheme
	interval time.Duration
}

// New 构建巡检源。interval<=0=关断（Enabled=false，Run 即返）。
func New(bus *eventbus.Bus, sch *scheme.Scheme, interval time.Duration) *Ticker {
	return &Ticker{bus: bus, sch: sch, interval: interval}
}

// Enabled 巡检源是否开启：interval>0 ∧ 类型已注册（注册表闸门前置——未注册即不启动，
// 声明先行的机械落点）。调用方据 stderr 报告开启/未开启因由。
func (t *Ticker) Enabled() bool {
	if t == nil || t.interval <= 0 || t.bus == nil || t.sch == nil {
		return false
	}
	_, err := t.sch.LookupEventType(EventType)
	return err == nil
}

// Interval 巡检间隔（stderr 装配锚报告用；未开启时返回零值）。
func (t *Ticker) Interval() time.Duration {
	if t == nil {
		return 0
	}
	return t.interval
}

// Run 巡检主循环：每 interval 到点自产一条 ReconcileTick（ctx 取消即退场，幂等可重启）。
// 未开启（Enabled=false）即返——关断=零事件（机械判别器的关断面）。
func (t *Ticker) Run(ctx context.Context) {
	if !t.Enabled() {
		return
	}
	tk := time.NewTicker(t.interval)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-tk.C:
			t.tick(now) // 发布失败不中断巡检（可丢语义，event.md §1——下一拍照常到点）
		}
	}
}

// tick 到点自产一条巡检事件（拍点时刻显式注入，测试可重放同拍验证聚合口径）。
func (t *Ticker) tick(now time.Time) error {
	et, err := t.sch.LookupEventType(EventType)
	if err != nil {
		return err // 闸门前置：未注册不发布（Enabled 已挡，防御性再断）
	}
	evt := eventbus.Event{
		Type:     EventType,
		Source:   EventSource,
		Involved: EventInvolved,
		At:       now,
		// 拍点即因由：同拍重放窗口聚合、跨拍不聚合（event.md 规则 10）。
		DedupKey: "reconcile:" + now.UTC().Format(time.RFC3339Nano),
		Level:    et.DefaultLevel, // 注册表缺省投影（ReconcileTick→info）；空值由总线归一 info
	}
	return t.bus.Publish(&evt)
}
