// scope.go —— Agent 事件 scope 闸（BUG-010，event.md 规则 19③——派发与消费双侧
// `matchScope(event.metadata.scope, agent.spec.eventScopes)` 强制校验）。
//
// 两个面共用一份判定：
//   - 派发闸：装配层把 AgentScopeGate 挂到触发骨架（workflow.TriggerHub.BindScopeGate），
//     执行面=agent 的绑定在推进唤醒投影前过闸——越界事件**不投递**（唤醒投影缺席），
//     OnBlocked 告警回调可见不静默（装配层注 log warning），ScopeBlocked 计数可查询。
//   - 消费闸：apiserver `GET /v1/events?agent=<锚>` 逐行过同一判定（Allow 静态词）——
//     Agent 会话取事件面越界拉取被服务侧过滤（不可见即不可消费）。
//
// 保守缺省（规则 19②）：Agent 对象在册但未声明 spec.eventScopes=只收 global/自身 scope；
// Agent 对象不在册=只收 global（锚名解析不命中的保守面——不发明 scope）。声明面支持
// 精确路径与子树通配（systems/<s>/*），单字符 * 恒全收（纪律「仅 builtinRoot 可持」由
// 声明面把关，词汇层不判身份）。positional parent（parent-agent 的 owner 定位）属
// agent-session 域 declared-deferred——本闸按锚名解析，未命中走保守缺省不炸。
//
// 解耦口径：本件只依赖 eventscope 词汇包与 workflow.ScopeGate 接口形（结构化满足，
// 零 import——判定=数据，装配层在 cored 侧做编译期断言）；Resolver 由装配层注入
// （nil=装配不完整，闸恒拒绝并告警——宁拒不降级）。
package flowbind

import (
	"fmt"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventscope"
)

// AgentEventScopes Agent 事件 scope 声明解析口面（装配层注入——文档库 Agent 对象
// spec.eventScopes 的现读投影；apiserver 消费闸与 cored 派发闸共用同一实例零双源）。
type AgentEventScopes interface {
	// EventScopes 按 agent 锚名（`<p>-agent` 前缀段：root/parent/<p>）返回该 Agent 的
	// spec.eventScopes 声明（patterns）与对象自身 scope（own——保守缺省的第二收窄项）。
	// ok=false=Agent 对象不在册（闸门走保守缺省：只收 global）。
	EventScopes(agent string) (patterns []string, own string, ok bool)
}

// AgentEventScopesFunc 函数适配器（装配层匿名实现用——免建 struct）。
type AgentEventScopesFunc func(agent string) (patterns []string, own string, ok bool)

// EventScopes 满足 AgentEventScopes。
func (f AgentEventScopesFunc) EventScopes(agent string) ([]string, string, bool) {
	return f(agent)
}

// ScopeBlockedBlock 一条拦截告警的静态词（OnBlocked 回调载荷——装配层转 log warning，
// 可见不静默：拦截必有痕，不留「事件消失」的静默面）。
type ScopeBlockedBlock struct {
	Agent      string // 被拦投递目标（agent 锚名）
	EventScope string // 事件 metadata.scope
	EventID    string // 事件 id（可回指 event_log）
	Type       string // 事件类型
	Face       string // 拦截面：dispatch（派发闸）/consume（消费闸）
	Reason     string // 未放行因由（no-agent=锚不在册 / undeclared=未声明保守缺省 / out-of-declaration=声明外）
}

// BlockedNote 拦截告警的人读词形（装配层 log.Printf 用；单点拼装零字面量散用）。
func (b ScopeBlockedBlock) BlockedNote() string {
	return fmt.Sprintf("scope 闸拦截 face=%s agent=%s scope=%s type=%s event=%s 因由=%s",
		b.Face, b.Agent, b.EventScope, b.Type, b.EventID, b.Reason)
}

// 拦截因由词汇（闭合集——闸内判定产物，测试锚定）。
const (
	BlockReasonNoAgent          = "no-agent"           // Agent 锚不在册（解析未命中）
	BlockReasonUndeclared       = "undeclared"         // 在册但未声明 eventScopes——保守缺省外
	BlockReasonOutOfDeclaration = "out-of-declaration" // 已声明但事件 scope 在声明外
	// FaceDispatch 派发闸（触发骨架唤醒投影投递前）。
	FaceDispatch = "dispatch"
	// FaceConsume 消费闸（Agent 会话取事件面查询侧）。
	FaceConsume = "consume"
)

// AgentScopeGate scope 闸（BUG-010 派发/消费双侧判定面；零值不可用——Resolver 必配）。
// 实现 workflow.ScopeGate（Allow(eventScope, agent)——装配层编译期断言）。
type AgentScopeGate struct {
	// Resolver Agent 声明解析源（nil=装配不完整——恒拒绝并告警，宁拒不降级）。
	Resolver AgentEventScopes
	// OnBlocked 拦截告警回调（nil=静默计数不推荐——规则「可见不静默」的机械落点；
	// 回调须非阻塞——闸在派发锁内路径上被调）。
	OnBlocked func(b ScopeBlockedBlock)
}

// Allow 判事件 scope 是否可投递/可消费（eventScope ∈ agent 声明集）。判定口径：
//
//	① 锚不在册（ok=false）       → 只收 global（保守缺省，因由 no-agent）
//	② 在册未声明 eventScopes     → 只收 global/自身 scope（规则 19② 保守缺省，因由 undeclared）
//	③ 已声明                     → eventscope.Match(patterns, scope)（因由 out-of-declaration）
//
// 拦截（返回 false）必经 OnBlocked 告警（nil Resolver 同口径——装配缺口也可见不静默）。
func (g *AgentScopeGate) Allow(eventScope, agent string) bool {
	allowed, reason := g.Judge(eventScope, agent)
	if allowed {
		return true
	}
	if g.OnBlocked != nil {
		g.OnBlocked(ScopeBlockedBlock{
			Agent: agent, EventScope: eventScope, Face: FaceDispatch, Reason: reason,
		})
	}
	return false
}

// Judge 判定（不带告警副作用——消费闸逐行过滤用：行级拦截告警会刷屏，消费闸只在
// 请求级注记一次）。返回 (放行, 拦截因由)；放行时因由空。
func (g *AgentScopeGate) Judge(eventScope, agent string) (bool, string) {
	if g.Resolver == nil {
		return false, BlockReasonNoAgent // 装配不完整=宁拒不降级（告警由调用方承担）
	}
	patterns, own, ok := g.Resolver.EventScopes(agent)
	if !ok {
		if eventScope == eventscope.ScopeGlobal {
			return true, ""
		}
		return false, BlockReasonNoAgent
	}
	if len(patterns) == 0 {
		// 规则 19② 保守缺省：未声明=只收 global/自身 scope。
		def := []string{eventscope.ScopeGlobal}
		if own != "" {
			def = append(def, own)
		}
		if eventscope.Match(def, eventScope) {
			return true, ""
		}
		return false, BlockReasonUndeclared
	}
	if eventscope.Match(patterns, eventScope) {
		return true, ""
	}
	return false, BlockReasonOutOfDeclaration
}

// JudgeConsume 消费闸行级判定（Face=consume 的告警变体；apiserver 逐行过滤用 Judge，
// 请求级注记用本词形拼装）。
func (g *AgentScopeGate) JudgeConsume(eventScope, agent string) (bool, ScopeBlockedBlock) {
	allowed, reason := g.Judge(eventScope, agent)
	if allowed {
		return true, ScopeBlockedBlock{}
	}
	return false, ScopeBlockedBlock{Agent: agent, EventScope: eventScope, Face: FaceConsume, Reason: reason}
}
