// credential.go —— 写路径凭据来源闸门（用户 2026-11-23 指示：凭据脱敏放在 Provider 里面，
// 默认走 Secret，运行时态不导出）。机械落点：Provider 对象 auth.secret.source 只允许引用式
// 取值（provider.md 规则 2——secret:/env:/vault:/keychain: 前缀），出现非引用形态=拒绝落库。
// 两个后端（git/sqlite）同一闸门——脱敏语义属存储写路径，不属后端实现细节。
//
// 边界：本闸门只挡「声明面凭据脱敏」的形态（明文直写 auth.secret.source）；真实凭据值永远
// 不进文档面——它们住部署侧 secretstore（独立 sqlite，--secrets-db），任何查询/导出面不携带。

package repository

import (
	"errors"
	"fmt"

	"gopkg.in/yaml.v3"
)

// ErrCredentialSource Provider 凭据来源不合引用式形态（provider.md 规则 2 脱敏闸门）。
var ErrCredentialSource = errors.New("repository: 凭据来源须为引用式取值（secret:/env:/vault:/keychain:）")

// credentialSourcePrefixes 引用式取值合法前缀（provider.md 规则 2；secret:=部署侧 secretstore
// 运行时面首段，与 vault:/env:/keychain: 并列）。
var credentialSourcePrefixes = []string{"secret:", "env:", "vault:", "keychain:"}

// checkCredentialSource kind=Provider 的写路径闸门：spec.auth.secret.source 若声明则必须为
// 引用式前缀（非引用形态=明文凭据嫌疑，拒绝落库——宁可拒也不静默入库）。
// 其余 kind 零额外校验（闸门只挂在凭据声明所在对象——脱敏放在 Provider 里面）。
func checkCredentialSource(kind string, doc *yaml.Node) error {
	if kind != "Provider" {
		return nil
	}
	spec := mapValue(doc, "spec")
	if spec == nil {
		return nil // 结构校验归 scheme/apply 管线；此处只做凭据来源形态闸门
	}
	auth := mapValue(spec, "auth")
	if auth == nil {
		return nil
	}
	secret := mapValue(auth, "secret")
	if secret == nil || secret.Kind != yaml.MappingNode {
		return nil // secret 缺省=无凭据声明（provider.md：四要素可空）
	}
	src := mapValue(secret, "source")
	if src == nil || src.Kind != yaml.ScalarNode || src.Value == "" {
		return nil
	}
	for _, p := range credentialSourcePrefixes {
		if hasPrefix(src.Value, p) {
			return nil
		}
	}
	return fmt.Errorf("%w: auth.secret.source=%q", ErrCredentialSource, src.Value)
}

func hasPrefix(s, p string) bool { return len(s) >= len(p) && s[:len(p)] == p }
