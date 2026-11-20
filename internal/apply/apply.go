// Package apply 文档面唯一写路径：Policy → 文档级比对(no-op) → 乐观锁 → 双写。
// 设计依据：core_system/concepts/apply.md；instance 仓 runtime/ops-log.md。
package apply

import (
	"context"
	"fmt"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/policy"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
)

// Options Apply 选项（core_system/concepts/apply.md §3）。
type Options struct {
	DryRun        bool
	Force         bool
	FieldManager  string // 必填，无主写入拒绝（revision.md 规则 1）
	ReasonRef     string // 指向 DriftReport/Event/工单
}

// Engine 写路径引擎。
type Engine struct {
	Repo  repository.Repository
	Pol   policy.Engine
}

// Apply 执行写路径。幂等：同内容 no-op 且不产生事件（防自我唤醒）。
func (e *Engine) Apply(ctx context.Context, o *repository.Object, opt Options) (rv int64, err error) {
	if opt.FieldManager == "" {
		return 0, fmt.Errorf("apply: fieldManager required")
	}
	// 1. Policy 准入（deny/humanReview/allow）
	dec, err := e.Pol.Evaluate(o.Kind, o.Scope, opt.FieldManager)
	if err != nil {
		return 0, err
	}
	if dec == policy.DecisionHumanReview {
		return 0, fmt.Errorf("apply: needs human review (ApprovalNeeded)")
	}
	if dec == policy.DecisionDeny {
		return 0, fmt.Errorf("apply: denied by policy")
	}
	if opt.DryRun {
		return o.ResourceVersion, nil // TODO: plan 输出
	}
	// 2~4. 文档级比对/乐观锁/双写（TODO: skeleton；语义 diff=LLM 提案制，A33，不做字段级合并）
	return e.Repo.Write(ctx, o, opt.FieldManager, opt.ReasonRef)
}
