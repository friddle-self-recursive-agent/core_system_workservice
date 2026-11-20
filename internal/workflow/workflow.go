// Package workflow 执行态机（Workflow 概念，A26）。
// 管相位（什么时候允许干什么）；TaskScheme 管 Executing 相位内的步骤图。
// builtin：default-task（Agent 任务态机）、simple-shot、watch-only。
package workflow

import "time"

// Phase 相位。
type Phase string

const (
	Idle             Phase = "Idle"
	Analyzing        Phase = "Analyzing"
	Planning         Phase = "Planning"
	AwaitingApproval Phase = "AwaitingApproval"
	Executing        Phase = "Executing"
	Verifying        Phase = "Verifying"
	Reporting        Phase = "Reporting"
	Failed           Phase = "Failed"
)

// Transition 相位迁移（纯声明：from + on 事件 → to）。
type Transition struct {
	From Phase
	On   string // 事件名（Scheme 注册）
	To   Phase
}

// OnFailed 失败出路：backoff → 超限 Escalate（Bind 的人）→ Idle。
type OnFailed struct {
	Backoff     time.Duration
	MaxAttempts int
	EscalateTo  string // bind 引用
}

// Spec Workflow 声明。
type Spec struct {
	Name        string
	Phases      []Phase
	Transitions []Transition
	OnFailed    OnFailed
	PhaseBudget map[Phase]time.Duration
}

// Builtin 默认态机（core_system/concepts/workflow.md §3）。
func Builtin(name string) *Spec {
	switch name {
	case "simple-shot":
		return &Spec{Name: name, Phases: []Phase{Idle, Executing, Reporting}}
	default: // default-task
		return &Spec{Name: "default-task", Phases: []Phase{
			Idle, Analyzing, Planning, AwaitingApproval, Executing, Verifying, Reporting, Failed,
		}}
	}
}

// Can 迁移合法性判定。
func (s *Spec) Can(from Phase, on string) (Phase, bool) {
	for _, t := range s.Transitions {
		if t.From == from && t.On == on {
			return t.To, true
		}
	}
	return from, false
}
