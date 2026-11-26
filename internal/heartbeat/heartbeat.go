// Package heartbeat 底层心跳事件源（BUG-010 第一段——event.md 规则 18：Heartbeat 一等
// 建模为周期正常心跳事件，tier=infra / visibility=default-hidden）。cored 装配层定时器
// 到点自产 Heartbeat 进总线（patrol.Ticker 同款装配范式——零新 HTTP face 的定时器面）。
// 与注入面同一条总线主干零旁路（窗口去重/订阅派发/event_log 归档），差异四处：
// source 服务侧判定恒 core（引擎自身的心跳来源标注是审计事实——巡检/阈值桥同口径）；
// 类型注册闸门前置（Heartbeat 未注册即不启动——注册表是引擎内部源唯一开关）；
// level=注册表缺省投影（Heartbeat→debug——rule 1b：debug 只进内存队列不入归档，
// 默认事件流/关注报告天然不可见=visibility 过滤的双保险）；缺省关断（interval=0
// 零事件，缺省不改既有行为面——env CORE_SYSTEM_HEARTBEAT_INTERVAL 显式开启）。
// route=dead-letter 零分发（scheme.dispatch 表行 on:Heartbeat→dead-letter——观测兜底
// 出口，零工作流零唤醒零通知副作用）。
package heartbeat

import (
	"context"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// 心跳事件声明常量（与 scheme.yaml eventTypes/dispatch 行同源；event.md 规则 18）。
const (
	// EventType 心跳事件类型（注册表闸门键）。
	EventType = "Heartbeat"
	// EventSource 服务侧判定恒 core（引擎自身产生的底层心跳是审计事实）。
	EventSource = "core"
	// Subject 心跳主体（本包心跳=cored 常驻进程自身——involved=自身，dedup 模板
	// heartbeat:<subject>:<window> 的 <subject> 段；多主体心跳随 agent-session 域扩展顺延）。
	Subject = "cored"
	// EventInvolved involved=自身（kind 投影不做注册闸——involved 是路由投影不是对象准入；
	// scope 段=global，总线侧 FromInvolved 派生回填 global）。
	EventInvolved = "Core/global/" + Subject
)

// Ticker 心跳事件源（cored 装配层持有；零值不可用，经 New 构建）。
type Ticker struct {
	bus      *eventbus.Bus
	sch      *scheme.Scheme
	interval time.Duration
}

// New 构建心跳源。interval<=0=关断（Enabled=false，Run 即返）。
func New(bus *eventbus.Bus, sch *scheme.Scheme, interval time.Duration) *Ticker {
	return &Ticker{bus: bus, sch: sch, interval: interval}
}

// Enabled 心跳源是否开启：interval>0 ∧ 类型已注册（注册表闸门前置——未注册即不启动，
// 声明先行的机械落点）。调用方据 stderr 报告开启/未开启因由。
func (t *Ticker) Enabled() bool {
	if t == nil || t.interval <= 0 || t.bus == nil || t.sch == nil {
		return false
	}
	_, err := t.sch.LookupEventType(EventType)
	return err == nil
}

// Interval 心跳间隔（stderr 装配锚报告用；未开启时返回零值）。
func (t *Ticker) Interval() time.Duration {
	if t == nil {
		return 0
	}
	return t.interval
}

// Run 心跳主循环：每 interval 到点自产一条 Heartbeat（ctx 取消即退场，幂等可重启）。
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
			_ = t.tick(now) // 发布失败不中断心跳（可丢语义，event.md §1——下一拍照常到点）
		}
	}
}

// tick 到点自产一条心跳事件（拍点时刻显式注入，测试可重放同拍验证聚合口径）。
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
		// 拍点即因由（patrol 同款口径）：同拍重放窗口聚合、跨拍不聚合——周期心跳语义
		// 不被总线去重窗口吞没（dedup 模板 heartbeat:<subject>:<window> 的 <window>=拍点）。
		DedupKey: "heartbeat:" + Subject + ":" + now.UTC().Format(time.RFC3339Nano),
		Level:    et.DefaultLevel, // 注册表缺省投影（Heartbeat→debug；rule 1b 不入归档）
		// metadata.scope 强制回填（BUG-010 规则 19）：involved scope 段=global，
		// 总线 Publish 缺席推导兜底同词——此处显式定形使拍面判据可直读。
		Scope: "global",
	}
	return t.bus.Publish(&evt)
}
