// Package repository 文档库存储：v1 git 后端（串行写队列模拟事务）。
// 设计依据：core_system/concepts/core.md 可行性、instance 仓 runtime/ops-log.md 双写。
package repository

import "context"

// Object 文档面对象的最小表示（完整 meta 见 core_system/architecture.md §5）。
type Object struct {
	Kind            string
	Name            string
	Scope           string
	ResourceVersion int64  // 乐观锁（core_system/concepts/revision.md）
	Content         []byte // yaml/md 原文
}

// Repository 存储接口（未来换 DB 只换实现，布局降级为导入导出格式）。
// 对象键 = (kind, scope, name)：scope 四层物化为目录前缀（layout.md §2「路径=scope」），
// 同名对象可并存于不同 scope（如 Workflow 的 global 与系统两栖，scheme.yaml storage 交替模板），
// 故 scope 必须入键；scope="" 仅用于 List 的跨 scope 枚举。
type Repository interface {
	Get(ctx context.Context, kind, scope, name string) (*Object, error)
	List(ctx context.Context, kind, scope string) ([]*Object, error)
	// Write 串行队列内执行：乐观锁校验 + git commit + ops DB 双写。
	Write(ctx context.Context, o *Object, fieldManager, reasonRef string) (int64, error)
	// Watch 变更通知（v1：轮询 diff 模拟，instance 仓 loop/loop.md §4 兜底）。
	Watch(ctx context.Context) (<-chan string, error)
}
