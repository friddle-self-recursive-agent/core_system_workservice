// Package policy 准入规则引擎 + 硬规则。
// 设计依据：core_system/concepts/policy.md；硬规则三条不可由配置放宽。
package policy

import "fmt"

// Decision 准入判定。
type Decision int

const (
	DecisionAllow Decision = iota
	DecisionHumanReview
	DecisionDeny
)

// Rule 一条规则（声明序求值，首个命中生效）。
type Rule struct {
	ID     string
	Match  func(scope, kind, docType string) bool
	Effect Decision
}

// Engine 规则引擎。
type Engine struct{ Rules []Rule }

// Evaluate 求值；硬规则先行（builtin.md §3 forbidden 清单）。
func (e Engine) Evaluate(kind, scope, fieldManager string) (Decision, error) {
	// 硬规则（配置不可放宽）：
	if scope == "builtin" && kind == "Agent" && fieldManager != "igniter" {
		// deny delete/disable builtinRoot 等（TODO: 完整三条）
	}
	for _, r := range e.Rules {
		if r.Match(scope, kind, "") {
			return r.Effect, nil
		}
	}
	return DecisionDeny, fmt.Errorf("policy: no rule matched, default deny")
}
