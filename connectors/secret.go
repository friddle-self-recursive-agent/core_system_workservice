// secret.go —— secret 引用式取值的运行时面（R-ENG-3 契约层；用户 2026-11-23 指示：
// 凭据脱敏默认走 Secret，运行时态不导出）。
//
// 声明面（provider.md 规则 2）：secret 只允许引用式取值（secret:/env:/vault:/keychain:）。
// 契约层只定义「部署侧实现注入什么」：SecretStore 是 secret:<name> 来源的查询口面
//（真实值住部署侧 secretstore sqlite，workservice 导出面包）——实现侧凭据解析时经
// ResolveSecret 统一处理前缀集合，值永不进信封/参数/日志（信封零凭据口径不变）。
//
// 依赖方向：连接器仓 → 本包（本包不 import 连接器仓）。

package connectors

import (
	"fmt"
	"os"
	"strings"
)

// SecretStore secret:<name> 引用式取值的部署侧查询口面（workservice secretstore.Store
// 满足；连接器实现只认本口面，不反向依赖 workservice 内部）。
type SecretStore interface {
	// Secret 取一条秘密值；不存在=错误（实现侧自行映射缺口可见）。
	Secret(name string) (string, error)
}

// ResolveSecret 解析一次引用式取值（provider.md 规则 2 前缀集合的运行时面）：
//   - "secret:<name>" → store（nil store=缺口可见报错——部署侧未挂秘密库）
//   - "env:<VAR>"     → 进程环境变量
//   - "vault:…" / "keychain:…" → declared-deferred：显式报错不猜
//
// 其余形态（含明文）=错误（引用式闸门——与 repository 写路径同前缀集合）。
// getenv 缺省 os.Getenv（测试可注入）。
func ResolveSecret(source string, store SecretStore, getenv func(string) string) (string, error) {
	if getenv == nil {
		getenv = os.Getenv
	}
	if name, ok := strings.CutPrefix(source, "secret:"); ok {
		if name == "" {
			return "", fmt.Errorf("connectors: secret: 来源缺名称（%q）", source)
		}
		if store == nil {
			return "", fmt.Errorf("connectors: secret: 来源 %q 无部署侧秘密库（SecretStore 未注入——缺口可见）", source)
		}
		return store.Secret(name)
	}
	if v, ok := strings.CutPrefix(source, "env:"); ok {
		if v == "" {
			return "", fmt.Errorf("connectors: env: 来源缺变量名（%q）", source)
		}
		if got := getenv(v); got != "" {
			return got, nil
		}
		return "", fmt.Errorf("connectors: 环境变量未设或为空（%s）", source)
	}
	return "", fmt.Errorf("connectors: 来源不支持（只允许 secret:/env: 引用；vault:/keychain: declared-deferred；实得 %q）", source)
}
