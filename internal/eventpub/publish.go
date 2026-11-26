// Package eventpub 事件发布面（v0.20 多类型事件注入面首段）：类型注册校验+来源强制
// +level 缺省投影+总线发布。设计依据：core_design/concepts/event.md 规则 8——
// 发布源三分类（connector watch / 引擎内部 / 人），发布权限域与 apply 写权限域分立
// （publish × events 资源域 ≠ write × 对象 scope），注入事件与引擎内部事件同一条
// 总线主干（去重/订阅派发/event_log 归档零旁路）；注入通道 source 由服务侧判定恒 human
// （来源标注是审计事实不是请求参数）；level 缺省=注册表 defaultLevel 投影；debug 级
// 不入库只进内存队列（规则 1b，eventbus 既有口径）。
package eventpub

import (
	"errors"
	"fmt"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventscope"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// 发布面哨兵错误（HTTP 投影：ErrNoType/ErrLevel/ErrSpec→400、ErrNotRegistered→404、
// ErrDenied→403；eventbus.ErrRejected 透传→400）。
var (
	// ErrNoType 缺事件类型（决策必须有事件类型——与 resolve 查询同口径）。
	ErrNoType = errors.New("eventpub: 缺 type（事件必有类型）")
	// ErrNotRegistered 类型未注册（事件只校验类型注册，event.md §2 边界；野类型拒绝）。
	ErrNotRegistered = errors.New("eventpub: 事件类型未注册（Scheme.eventTypes 闸门）")
	// ErrLevel level 不在 log 式等级词汇表（debug/info/warn/error/fatal，rule 1b）。
	ErrLevel = errors.New("eventpub: level 不在词汇表（debug/info/warn/error/fatal）")
	// ErrDenied 发布权限域拒绝（publish × events；与 apply 写权限域分立——
	// 对象 write 不产生发布权，发布权不产生对象 write）。
	ErrDenied = errors.New("eventpub: 发布权限拒绝（动作=publish × 域=events）")
	// ErrScope 显式声明的 scope 越出四层词汇（BUG-010，event.md 规则 19——事件必带
	// 精确四层词 builtin|global|systems/<s>|systems/<s>/instances/<i>，野 scope 拒绝；
	// 缺席不是错误——走 involved 派生回填）。
	ErrScope = errors.New("eventpub: scope 不在四层词汇（builtin|global|systems/<s>|systems/<s>/instances/<i>）")
)

// SourceHuman 注入通道的来源标注（服务侧判定，载荷声明不采纳——event.md 规则 8）。
const SourceHuman = "human"

// LevelVocabulary log 式等级词汇表（event.md rule 1b；与 eventbus 归一化口径一致——
// 发布面在入总线前显式校验，拒绝越词值而不是静默归一）。
var LevelVocabulary = map[string]bool{"debug": true, "info": true, "warn": true, "error": true, "fatal": true}

// Request 一次注入请求（HTTP 载荷投影；Source 不在此——来源是服务侧判定）。
type Request struct {
	Type     string // 必填：Scheme.eventTypes 之一
	Involved string // involvedRef 投影 kind/scope/name（可缺省：路由按类型照走，owner 解析为空）
	Level    string // 可缺省：空=注册表 defaultLevel 投影；越词=拒绝
	DedupKey string // 可缺省：空=ID（永不聚合，eventbus 既有口径）
	Scope    string // 可缺省：metadata.scope 四层词显式声明（BUG-010，event.md 规则 19）；
	// 空=按 involved 所属 scope 服务侧推导回填（FromInvolved——派生不出合法词=兜底 global）；
	// 显式给出但越出四层词汇=拒绝（ErrScope，400）
}

// Publisher 事件发布面引擎（进程内单例，cored 装配；测试零值注入假总线）。
// Perm 缺省=零值权限系统（空授权面 fail-closed，仅 DefaultAdmin 字面量兜底）——
// 装配层在 --allow-apply 权限对象装载后注入同一 *permission.System（与 apply 写面
// 同一代权威快照，判定依据可追）。
type Publisher struct {
	Sch  *scheme.Scheme
	Bus  *eventbus.Bus
	Perm *permission.System
}

// Publish 注入一条事件：类型注册校验 → 主体发布权判定 → level 投影/校验 → 总线发布
// （同主干：窗口去重、订阅派发、event_log 归档全在 eventbus.Publish 内）。
// 返回规范化后的事件（ID/At/Level/DedupKey 就地归一）——回执与留痕 eventId 同源。
// 窗口内同 DedupKey 聚合返回原事件但不再投递不归档（eventbus 聚合语义，调用方可比对此前回执）。
func (p *Publisher) Publish(req Request, sub *permission.Subject) (*eventbus.Event, error) {
	if req.Type == "" {
		return nil, ErrNoType
	}
	et, err := p.Sch.LookupEventType(req.Type)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotRegistered, err)
	}
	// 发布权限域判定（ByRole：请求级主体 role 表查表——permission.md §4 b 路；
	// 判定留痕=event_log source=human + 权限审计随调用方 ops 面）。
	sys := p.Perm
	if sys == nil {
		sys = &permission.System{} // 空授权面 fail-closed（宁拒不降级）
	}
	if err := sys.Check(*sub, PublishDomain, PublishAction, permission.ByRole); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrDenied, err)
	}
	level := req.Level
	if level == "" {
		level = et.DefaultLevel // 注册表缺省投影（Escalate→error 等权威分诊口径）
	}
	if !LevelVocabulary[level] {
		return nil, fmt.Errorf("%w: %q", ErrLevel, level)
	}
	// metadata.scope（BUG-010，event.md 规则 19①）：显式声明先校验四层词汇（越词=拒绝
	// 不静默归一）；缺席=服务侧按 involved 所属 scope 推导回填（总线 Publish 内同口径兜底，
	// 这里先定形使回执 scope 与归档行一致——回执可回指）。
	if req.Scope != "" && !eventscope.Valid(req.Scope) {
		return nil, fmt.Errorf("%w: %q", ErrScope, req.Scope)
	}
	e := &eventbus.Event{
		Type: req.Type, Source: SourceHuman, Involved: req.Involved,
		Level: level, DedupKey: req.DedupKey,
		Scope: req.Scope,
	}
	if e.Scope == "" {
		e.Scope = eventscope.FromInvolved(e.Involved)
	}
	if err := p.Bus.Publish(e); err != nil {
		return nil, err // ErrRejected（缺 type 已前置拦截，防御性透传）
	}
	return e, nil
}

// 发布权限域词汇（event.md 规则 8：动作=publish × 域=events 资源域——
// permission.Role.Holds 资源域形态精确比对）。
const (
	PublishDomain = "events"
	PublishAction = "publish"
)
