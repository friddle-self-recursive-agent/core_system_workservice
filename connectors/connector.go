// connectors —— 幂等驱动（v0.3 拆分：ConnectorProvider 类型层 / ConnectorOps 使用层）。
// 设计依据：core_system/concepts/connector.md（A25）。
// 每个驱动的代码实现一个 Provider；Ops 是运行时挂载（环境+子集+Workflow），见 ConnectorOps。
package connectors

import "context"

// ProviderImpl 类型层：支持什么类型、对应什么代码（一个驱动包实现一份）。
type ProviderImpl interface {
	Type() string            // 类型标注（如 onedev-driver）
	Types() []string         // 支持的能力类型（如 scm, ci）
	Version() string
	Contract() Contract
	// 五方法幂等实现（P4）：
	Get(ctx context.Context, spec []byte) (state []byte, err error)
	Plan(ctx context.Context, spec, state []byte) (diff []byte, err error)
	Apply(ctx context.Context, spec []byte, dryRun bool) (actions []byte, err error)
	Verify(ctx context.Context, spec, state []byte) (bool, error)
	Watch(ctx context.Context) (<-chan Event, error) // 可降级 poll(30s)
	Undo(ctx context.Context, actionLog []byte) error
}

// Contract 契约支持度声明（undo none ⇒ 使用处强制人审）。
type Contract struct {
	Get, Plan, Apply, Verify, Watch bool
	Undo                            string // full | partial | none
}

// OpsSpec 使用层：挂载在 global 或 service/instance 的"一份可调用操作"。
// Agent 默认挂 ConnectorOps；使用要权限（P7）。
type OpsSpec struct {
	Name        string            // ops 名（Agent/TaskScheme 引用它）
	Provider    string            // 类型来源（ProviderImpl.Type()）
	Workflow    string            // 绑定的 Workflow（缺省 default-task，A26）
	Config      map[string]string // endpoint 等（值可为 ref:/credential: 引用）
	Credentials []string          // credential 引用
	Ops         []string          // 暴露的 op 子集
}

// Event 外部变化（转 eventbus.Event）。
type Event struct{ DedupKey, Detail string }

// providerRegistry 类型注册表（Scheme 闸门：未注册的 provider，Ops 不予创建）。
var providerRegistry = map[string]ProviderImpl{}

// RegisterProvider 注册类型层实现（驱动包 init 时调用）。
func RegisterProvider(p ProviderImpl) { providerRegistry[p.Type()] = p }

// LookupProvider 查类型层实现。
func LookupProvider(typeName string) (ProviderImpl, bool) {
	p, ok := providerRegistry[typeName]
	return p, ok
}
