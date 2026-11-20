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

// ProviderSpec 类型层声明（v0.6 简化：就是绑定，A38）。
// yaml 字段：types/scheme/usage/llmPrompt/download；impl 三形态 tool/image/source。
type ProviderSpec struct {
	Types    []string        // 类型标注（scm, ci…）
	Impl     Impl            // 绑定的实现
	Download string          // 工具下载地方（image:// git:// pkg://）
	Scheme   map[string]any  // 使用方式 schema（ops/params，机器读）
	Usage    string          // 使用方式（人读）
	LLMPrompt string         // LLM 注入提示：Agent 会话使用本工具时注入
}

// Impl 绑定的实现三形态。
type Impl struct {
	Kind string // tool（外部直接工具）| image（镜像）| source（源代码）
	Ref  string // source 形态：Ref 指代码地址（自己仓统一管理）；image 形态：镜像地址
}

// ResolveProviderName 命名解析（A38）：无前缀=默认 org；带前缀=github.com/<org>/<name>。
func ResolveProviderName(name, defaultOrg string) string {
	for _, c := range name {
		if c == '/' {
			return "github.com/" + name // 带前缀 ⇒ GitHub 完整配置
		}
	}
	return defaultOrg + "/" + name // 无前缀 ⇒ 默认 org
}

// OpsSpec 使用层：工具调用面/提供的功能（v0.6：无探活/drift/finalizer 概念级机制）。
// Agent 默认挂 ConnectorOps；使用要权限（P7）。
type OpsSpec struct {
	Name        string            // ops 名（Agent/TaskScheme 引用它）
	Provider    string            // 类型来源（经命名解析）
	Workflow    string            // 绑定的 Workflow（缺省 default-task，A26）
	Config      map[string]string // endpoint 等（值可为 ref:/credential: 引用）
	Credentials []string          // credential 引用
	Ops         []string          // 提供的功能（op 子集）
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
