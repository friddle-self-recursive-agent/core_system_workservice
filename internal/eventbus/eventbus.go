// Package eventbus 事件队列：dedupKey 聚合、dispatch 路由、死信。
// 设计依据：core_system/concepts/event.md（只触发不修改；P3 level-triggered）。
package eventbus

import "time"

// Event 事件（短生命周期数据，非文档对象）。
type Event struct {
	ID       string
	Type     string
	Source   string
	Involved string // involvedRef：路由按 类型+involved/owner+scope（v0.11 A47，无 selector）
	At       time.Time
	DedupKey string
	Level    string // debug/info/warn/error/fatal（A42）
}

// Bus 事件总线（Core 内）。
type Bus struct {
	deadLetters []Event
}

// Publish 去重窗口内同 DedupKey 聚合（防风暴，event.md 规则 1）。
func (b *Bus) Publish(e Event) error { return nil } // TODO: 窗口聚合+路由

// DeadLetters 路由不到的事件（必须可见，观测面健康指标）。
func (b *Bus) DeadLetters() []Event { return b.deadLetters }
