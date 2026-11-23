// opsconnector.go —— Ops/工具调用域契约（R-ENG-3 契约层，v0.16.1）。
// connector.md：Agent/Workflow 经 ConnectorOps 调外部工具；描述性内容（op 集/参数/
// 幂等性）全在 WorkService 六面体 tools/connector 面——本契约只约束**实现侧**：一个
// OpsConnector 就是一份可执行的 ops 实现（notify/send-email、IM 通知、未来 webhook 等）。
//
// 仓库归属（rule.md R-ENG-3）：本包只放 builtin 契约+注册表；通用类实现（notify/email）
// 在部署侧独立仓 common_connector，用法类实现（onedev/nas/oss）在 sybran_usage_connector
// ——依赖方向恒为 **connector 仓 → 本包**（本仓不 import 连接器仓，装配在部署侧）。
//
// 与资源域 ProviderImpl 的分野（同一包两契约，两物）：ProviderImpl=资源域幂等驱动
// （Get/Plan/Apply/Verify/Watch/Undo 六方法，声明面对账）；OpsConnector=工具调用域
// （调用即效果，无 spec/state 对账语义）——通知类 OpsConnector 是它的第一个实现族。
//
// secret 口径（provider.md 规则 2/6）：实现侧凭据一律引用式取值（env:/vault:/keychain:），
// 声明面永不落明文；本契约的 params/result 均为 JSON 字节（UTF-8），op 名小写中划线。
package connectors

import (
	"context"
	"sort"
	"sync"
)

// OpsConnector 工具调用域契约（五方法）。实现包在 init 时 RegisterOpsConnector 自注册；
// 调用侧经 LookupOpsConnector 取用。实现须可在无凭证配置下安全构建（真实取值在 Invoke/
// Probe 时才解析——声明/运行时分离，provider.md 规则 6）。
type OpsConnector interface {
	// Type ops 连接器类型标注（如 notify-email；同一 WorkService.connector.types 词汇）。
	Type() string
	// Version 实现版本（协议/行为演进标记；只加不改）。
	Version() string
	// Ops 提供的 op 集（⊆ 所属 WorkService.tools 声明，connector.md 规则 2 的实现侧对应）。
	Ops() []string
	// Invoke 执行一次 op（调用即效果）。params/result=JSON 字节；op 不在 Ops() 内 → 错误；
	// 幂等性与 undo 语义由 WorkService.tools 声明（如 send-email：不可撤回，undo:none）。
	Invoke(ctx context.Context, op string, params []byte) (result []byte, err error)
	// Probe 通道可达性探测（Notification 生命周期「登记走 Provider 校验：channel 可达性
	// 探测」的机读落点，notification.md §4；不发送任何用户内容）。
	Probe(ctx context.Context) error
}

var (
	opsMu       sync.RWMutex
	opsRegistry = map[string]OpsConnector{} // key=Type()；同型重注册=后到覆盖（部署侧重配语义）
)

// RegisterOpsConnector 注册 ops 实现（驱动包 init 调用；同型覆盖，幂等安全）。
func RegisterOpsConnector(c OpsConnector) {
	opsMu.Lock()
	defer opsMu.Unlock()
	opsRegistry[c.Type()] = c
}

// LookupOpsConnector 查 ops 实现。
func LookupOpsConnector(typeName string) (OpsConnector, bool) {
	opsMu.RLock()
	defer opsMu.RUnlock()
	c, ok := opsRegistry[typeName]
	return c, ok
}

// OpsConnectorTypes 已注册 ops 实现类型清单（观测面；字典序）。
func OpsConnectorTypes() []string {
	opsMu.RLock()
	defer opsMu.RUnlock()
	ts := make([]string, 0, len(opsRegistry))
	for t := range opsRegistry {
		ts = append(ts, t)
	}
	sort.Strings(ts)
	return ts
}

// HasOp op 是否在该连接器声明的 op 集内（调用侧防御：contract Ops() ⊆ tools Lint 的运行时对应）。
func HasOp(c OpsConnector, op string) bool {
	for _, o := range c.Ops() {
		if o == op {
			return true
		}
	}
	return false
}
