// connectors —— 幂等驱动契约层。
// v0.13（A53）：ConnectorProvider 并入 WorkService 六面体的 connector 面（见 internal/workservice）；
// 本包保留五方法幂等实现接口（驱动代码实现它）与 Ops 挂载结构（concepts/connector.md）。
// 仓库归属（rule.md R-ENG-3）：本目录只放 builtin 驱动（gitfs/dsh）；用法型驱动
// （onedev/nas/oss 等，换部署即换）在部署侧独立仓（sybran_usage_connector），依赖方向恒为用法仓 → 本仓。
package connectors

import (
	"context"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workservice"
)

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
// v0.13（A53）：此结构迁往 internal/workservice.ConnectorFacet，保留别名兼容。
type ProviderSpec = workservice.ConnectorFacet

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
// Agent 默认挂 ConnectorOps；使用要权限（P7）。v0.13：Workservice 指向六面体（A53）。
type OpsSpec struct {
	Name        string            // ops 名（Agent/TaskScheme 引用它）
	Workservice string            // v0.13 A53：指向 WorkService（原 Provider 字段）
	Workflow    string            // 绑定的 Workflow（缺省 default-task，A26）
	Config     map[string]string // endpoint 等（值可为 ref: 引用）
	AuthProvider string          // v0.8：鉴权走 Provider.auth.secret（Credential 已删，A40）
	Ops        []string          // 提供的功能（op 子集）
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
