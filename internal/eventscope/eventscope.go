// Package eventscope 事件 scope 词汇与匹配（BUG-010 事件分层与 scope 隔离——event.md
// 规则 19 的机械落点）。叶子包：零内部依赖，eventbus/eventpub/flowbind/apiserver/cored
// 共用同一份推导与匹配口径（判定面无双源漂移）。
//
// 词汇（layout.md「路径=scope」同源）：四层 Scope= builtin | global | systems/<s> |
// systems/<s>/instances/<i>。每条事件入总线时服务侧强制回填 metadata.scope（显式声明
// 优先，缺席按 involved 派生——派生不出合法四层词=兜底 global，禁止无 scope 的野事件）；
// Agent 侧 spec.eventScopes 为 scopePattern 清单，支持精确路径与子树通配（systems/<s>/*
// ——尾随 /*；单字符 * 恒全收，rule 19 仅 builtinRoot 可持——词汇层不判身份，收口在闸门）。
package eventscope

import "strings"

// 四层 Scope 词汇根（event.md 规则 19；layout.md §2 路径=scope 准绳）。
const (
	ScopeBuiltin = "builtin"
	ScopeGlobal  = "global"
	// Wildcard 单字符通配（scopePattern 专用；event scope 本体不使用——事件必带精确
	// 四层词，通配只出现在 Agent spec.eventScopes 声明面）。
	Wildcard = "*"
)

// Valid 判 scope 是否合法四层词（事件发布面显式声明校验用）：
//
//	builtin | global
//	systems/<s>
//	systems/<s>/instances/<i>
//
// 空=不合法（事件必带 scope——缺席走派生兜底，不静默放空词）。
func Valid(scope string) bool {
	switch scope {
	case ScopeBuiltin, ScopeGlobal:
		return true
	}
	segs := strings.Split(scope, "/")
	// systems/<s>（2 段）或 systems/<s>/instances/<i>（4 段，段 2 恒 instances）。
	if len(segs) == 2 && segs[0] == "systems" && segs[1] != "" {
		return true
	}
	return len(segs) == 4 && segs[0] == "systems" && segs[1] != "" &&
		segs[2] == "instances" && segs[3] != ""
}

// FromInvolved 按 involved 投影派生 scope（服务侧强制回填的缺省来源——event.md 规则 19①
// 「由产生源或 involvedRef 所属 scope 服务侧强制回填」）。
//
// involved 同构=kind/scope/name（apply objKey 同形；scope 自身可含 systems 层级多段），
// 取首段之后、末段之前的部分；派生值不是合法四层词（畸形 involved/非对象锚如
// Session/cored/<id>）=兜底 global（保守不炸——旧数据兼容口径）。
// involved 缺段（无 scope 段）同口径兜底。
func FromInvolved(involved string) string {
	segs := strings.Split(involved, "/")
	if len(segs) < 3 {
		return ScopeGlobal // 无 scope 段（kind/name 两段或裸值）——保守兜底
	}
	scope := strings.Join(segs[1:len(segs)-1], "/")
	if !Valid(scope) {
		return ScopeGlobal // 畸形（如 Session/cored/<id> 的 cored）——保守兜底不发明词
	}
	return scope
}

// Match scopePattern 匹配（agent.spec.eventScopes ∈ 校验的判定面；event.md 规则 19②③）：
//
//	*            → 恒命中（全收；声明纪律「仅 builtinRoot 可持」由闸门调用方把关）
//	<p>/*        → 前缀式子树命中（scope == <p> 或 <p>/ 前缀——路径=scope，子树含根）
//	其余         → 精确等值
//
// patterns 空=零命中（调用方决定保守缺省——本包不做缺省语义，缺省词形属闸门声明面）。
func Match(patterns []string, scope string) bool {
	for _, p := range patterns {
		if p == Wildcard {
			return true
		}
		if root, hasGlob := strings.CutSuffix(p, "/*"); hasGlob {
			if scope == root || strings.HasPrefix(scope, root+"/") {
				return true
			}
			continue
		}
		if p == scope {
			return true
		}
	}
	return false
}
