// Package workflow 执行态机 + 触发规则（v0.8 吸收 Loop，A41）。
// Workflow = 触发规则→事件整理→事件处理（相位态机）→文档回归；
// TaskScheme 管 Executing 相位内的步骤图。
// builtin：core-loop（系统级循环）、default-task（任务态机）、simple-shot、watch-only。
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

// Triggers 触发规则（原 Loop.watch，A41）。
type Triggers struct {
	Events   map[string]string // selector
	MinLevel string            // 事件 Level 过滤（A42：debug/info/warn/error/fatal）
	Resync   time.Duration     // 巡检兜底（level-triggered 保险丝）
}

// Brakes 刹车三件套（R7 确认保留；可收紧不可关闭）。
type Brakes struct {
	Cooldown         time.Duration // 成功后最小间隔（防震荡）
	MaxMutationsTick int           // 单轮最大变更数
	AvoidRepeatFix   bool          // DriftReport 记录"改过什么"，避免重复同类修复
}

// Cost 一次 loop 的成本（每轮必记 budget_log）。
type Cost struct {
	BudgetMinutes int
	BudgetTokens  int64
	BudgetMutations int
}

// FullSpec 完整 Workflow（触发+态机+刹车+成本+回归）。
type FullSpec struct {
	Spec
	Triggers Triggers
	Brakes   Brakes
	Cost     Cost
}

// Builtin 默认态机（core_system/concepts/workflow.md §3）。
func Builtin(name string) *Spec {
	switch name {
	case "simple-shot":
		return &Spec{Name: name, Phases: []Phase{Idle, Executing, Reporting}}
	case "watch-only":
		return &Spec{Name: name, Phases: []Phase{Idle, Analyzing, Reporting}}
	case "core-loop":
		return &Spec{Name: name, Phases: []Phase{ // 系统级循环（A41）
			Idle, Analyzing, Planning, AwaitingApproval, Executing, Verifying, Reporting, Failed,
		}}
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
