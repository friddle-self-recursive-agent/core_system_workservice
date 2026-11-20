// Package plugin dsh 插件宿主接口（五方法契约）。
// 设计依据：core_system/concepts/plugin.md；基础 agent=dsh（R1.2.1）。
package plugin

import "context"

// Plugin 五方法契约（所有插件同一接口）。
type Plugin interface {
	Init(config map[string]string) error
	OnEvent(ctx context.Context, evt Event) ([]Action, error)
	Dispatch(ctx context.Context, task Task) (Result, error) // root 直控通道调用点
	Report() Status
	Shutdown() error
}

// Event 插件视角的事件（与 eventbus.Event 对齐）。
type Event struct{ Type, Source, DedupKey string }

// Task 派发任务（含 TaskScheme 引用与参数）。
type Task struct {
	ID       string
	Scheme   string // TaskScheme 名（instance 仓 runtime/execution-model.md）
	Params   map[string]string
	BudgetMs int64
}

// Action 插件响应动作（上报/派发/中断）。
type Action struct{ Kind, Target string }

// Result 任务结果。
type Result struct{ TaskID, Outcome string }

// Status 插件状态上报。
type Status struct {
	Healthy bool
	Detail  string
}

// Host dsh 宿主：装载/看护插件（崩溃=HeartbeatLost→root 重启）。
type Host interface {
	Load(name string, p Plugin) error
	Plugins() []string
}
