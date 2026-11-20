// Package apply 文档面唯一写路径：权限系统准入 → 文档级比对(no-op) → 乐观锁 → 双写。
// 设计依据：core_system/concepts/apply.md；instance 仓 runtime/ops-log.md。
package apply

import (
	"context"
	"fmt"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
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
	Perm  permission.System
}

// Apply 执行写路径。幂等：同内容 no-op 且不产生事件（防自我唤醒）。
func (e *Engine) Apply(ctx context.Context, o *repository.Object, opt Options) (rv int64, err error) {
	if opt.FieldManager == "" {
		return 0, fmt.Errorf("apply: fieldManager required")
	}
	// 1. 权限系统准入（v0.9 A44：原 Policy 并入权限系统；人审档命中=ApprovalNeeded）
	if err := e.Perm.Check(permission.Subject{User: opt.FieldManager}, o.Scope, "write", permission.ByRole); err != nil {
		return 0, fmt.Errorf("apply: %w", err)
	}
	if err := permission.HardRules(o.Kind+"/"+o.Name, "write"); err != nil {
		return 0, err
	}
	if opt.DryRun {
		return o.ResourceVersion, nil // TODO: plan 输出
	}
	// 2~4. 文档级比对/乐观锁/双写（TODO: skeleton；语义 diff=LLM 提案制，A33，不做字段级合并）
	return e.Repo.Write(ctx, o, opt.FieldManager, opt.ReasonRef)
}
