// Package opslog ops 历史操作日志（SQLite WAL；六表见 instance 仓 runtime/ops-log.md §2）。
// git=文档内容版本；DB=操作流水。DB 非权威、可重建。
package opslog

import "time"

// ApplyLog apply_log 表。
type ApplyLog struct {
	Ts              time.Time
	FieldManager    string
	Obj             string
	ResourceVersion int64
	Result          string // ok | denied | conflict | recover
	ReasonRef       string
}

// EventLog event_log 表。
type EventLog struct {
	Ts     time.Time
	EventID string
	Type   string
	RoutedTo string
}

// Store ops DB 接口（v1 sqlite）。
type Store interface {
	AppendApply(l ApplyLog) error
	AppendEvent(l EventLog) error
	// Reconcile 每日对账：git 提交数 vs apply_log（self-recursive-core Goal.A3）。
	Reconcile() (diff int, err error)
}
