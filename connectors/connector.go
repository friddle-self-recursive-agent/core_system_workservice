// connectors —— 五方法幂等驱动（core_system/concepts/connector.md）。
// 每个驱动一个包：onedev / gitfs / nas / oss / dsh。
package connectors

import "context"

// Connector 统一契约：get/plan/apply/verify/watch/undo（全部幂等，P4）。
type Connector interface {
	Type() string
	Version() string
	Get(ctx context.Context, spec []byte) (state []byte, err error)
	Plan(ctx context.Context, spec, state []byte) (diff []byte, err error)
	Apply(ctx context.Context, spec []byte, dryRun bool) (actions []byte, err error)
	Verify(ctx context.Context, spec, state []byte) (bool, error)
	Watch(ctx context.Context) (<-chan Event, error) // watch 可降级 poll(30s)
	// Undo 回滚；不支持则注册时声明 undoUnsupported → Policy 强制人审。
	Undo(ctx context.Context, actionLog []byte) error
}

// Event 外部变化（转 eventbus.Event）。
type Event struct{ DedupKey, Detail string }

// Registry 驱动注册表（Scheme 闸门：未注册不存储）。
var Registry = map[string]Connector{}

// Register 注册驱动（init 时调用）。
func Register(c Connector) { Registry[c.Type()] = c }
