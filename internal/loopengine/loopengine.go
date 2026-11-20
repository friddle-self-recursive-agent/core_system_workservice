// Package loopengine Loop 调度：触发合并、backoff、预算、Escalate。
// 设计依据：core_system/concepts/loop.md + instance 仓 loop/loop.md 状态机。
package loopengine

import "time"

// Spec Loop 对象的 spec（core_system/concepts/loop.md §3）。
type Spec struct {
	Resync       time.Duration
	Cooldown     time.Duration
	MaxAttempts  int
	MaxMutations int // 单轮刹车
}

// Phase 实例级状态机（loop/loop.md §4）。
type Phase string

const (
	Idle        Phase = "Idle"
	Reconciling Phase = "Reconciling"
	CoolingDown Phase = "CoolingDown"
	Backoff     Phase = "Backoff"
	Escalated   Phase = "Escalated" // 一等终态，非异常
)

// Reconcile 单轮对账：幂等、level-triggered（同现实两遍结论一致）。
// 四层执行模型的 Agent/Scheme/代码层在此挂接（instance 仓 runtime/execution-model.md）。
func Reconcile(_ Spec) (Phase, error) { return Idle, nil } // TODO: skeleton
