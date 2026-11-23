// Package apply 文档面唯一写路径（core_system/concepts/apply.md）：
// scheme 准入（kind 注册 + spec 结构校验）→ 权限系统准入 → repository.Write（乐观锁/幂等，事务边界在存储串行队列）
// → DocApplied 事件必发（noop/无差异不发，防自我唤醒）。
// 设计依据：concepts/apply.md §1/§3/§6、concepts/revision.md、instance 仓 runtime/ops-log.md。
package apply

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/draft"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// 拒绝哨兵（读面/桥按此映射状态码；存储层哨兵原样透传）。
var (
	// ErrNoFieldManager 无主写入拒绝（apply.md rule 2：任何文档写入必须携带 fieldManager）。
	ErrNoFieldManager = errors.New("apply: 写入必须携带 fieldManager（无主写入拒绝）")
	// ErrNotRegistered 野对象拒绝：kind 未在 Scheme 注册（concepts/core.md 闸门）。
	ErrNotRegistered = errors.New("apply: kind 未注册（野对象拒绝）")
	// ErrDenied 权限系统准入拒绝（Permission Check / HardRules）。
	ErrDenied = errors.New("apply: 权限准入拒绝")
	// ErrObject 对象键不全（kind/scope/name 缺项）。
	ErrObject = errors.New("apply: 对象键不全（需 kind、scope、name）")
	// ErrSpecInvalid spec 结构准入拒绝：对象文档 spec 区未过该 kind 注册声明
	// （scheme.yaml 机读子集；R-TYPE 有规则即有 checker，checker 在 scheme 包）。
	ErrSpecInvalid = errors.New("apply: spec 结构未过准入（kind 注册声明校验拒绝）")
)

// Options Apply 选项（concepts/apply.md §3）。
type Options struct {
	DryRun       bool   // 零副作用 + 与真执行一致的准入判定（v1 计划=准入+现版本，语义 diff 归 LLM，A33）
	Force        bool   // 预留：跳过待审区直达（跨系统提案场景；v1 未接线）
	FieldManager string // 必填（rule 2）
	ReasonRef    string // 指向 DriftReport/Event/工单（审计回链）
	// Subject 请求级主体（v0.9，permission.md §3）：HTTP 写面按请求解析携带（subject 字段/X-Core-Subject），
	// 指针非 nil 即按请求判定——空主体照判（无 role 落点 → 拒绝，匿名请求不享受 fallback）；
	// nil=引擎缺省 fallback（e.Subject 进程级本地信任：程序化调用如 bootstrap/工作流自写）。
	Subject *permission.Subject
}

// Result 一次 Apply 的结果。
type Result struct {
	RV     int64           // 落库后 resourceVersion（noop=现版本；dry-run=现版本，缺席=0）
	Action string          // create | update | noop；dry-run 恒为 "dry-run"
	Event  *eventbus.Event // DocApplied 回执（noop/dry-run 无差异=nil，apply.md §3）
}

// Engine 写路径引擎。文档面写入的唯一编排点——任何写端点都收敛到这里。
type Engine struct {
	Sch  *scheme.Scheme
	Repo repository.Repository
	Perm *permission.System
	Bus  *eventbus.Bus
	// Ops ops DB（拒绝路径审计：denied 行；ok/noop/conflict 行由 repository 双写钩子落）。
	// nil=拒绝不落审计（不推荐：审计可追溯依赖此表，ops-log.md §2）。
	Ops opslog.Store
	// Subject 缺省写入主体（权限判定 fallback 输入）。v0.9 请求级主体落地后：
	// HTTP 写面一律按请求解析（Options.Subject 非 nil 即请求级，不再读本字段）；
	// 本字段退化为程序化调用的缺省 fallback（cored 本地信任面=显式开启时持 DefaultAdmin）。
	Subject permission.Subject
	// Drafts 草稿缓存（apply.md §1 编辑与发布分离）。nil=无草稿面（只读部署不受影响）；
	// 配置后 ApplyDraft 从缓存取稿走同一三步闸门，成功落库即清稿。
	Drafts *draft.Store
}

// Apply 执行写路径三步闸门。幂等：同内容重入 no-op，不产生提交、不发事件。
func (e *Engine) Apply(ctx context.Context, o *repository.Object, opt Options) (*Result, error) {
	if opt.FieldManager == "" {
		return nil, ErrNoFieldManager
	}
	if o == nil || o.Kind == "" || o.Name == "" || o.Scope == "" {
		return nil, ErrObject
	}
	objKey := fmt.Sprintf("%s/%s/%s", o.Kind, o.Scope, o.Name)

	if err := e.admit(objKey, o, opt); err != nil {
		return nil, err
	}

	// ③ dry-run：零副作用（不发事件、不落审计），返回与真执行一致的准入结论 + 现版本。
	// 语义 diff 计划（哪里变了/是否冲突/合并提案）= LLM 语义活（A33），v1 不做字段级合并。
	if opt.DryRun {
		cur, err := e.Repo.Get(ctx, o.Kind, o.Scope, o.Name)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return &Result{RV: 0, Action: "dry-run"}, nil // 将创建
			}
			return nil, err
		}
		return &Result{RV: cur.ResourceVersion, Action: "dry-run"}, nil
	}

	// ④ 乐观锁落库（串行队列=事务边界；幂等去 rv 规范比对在存储层）。
	rv, action, err := e.Repo.Write(ctx, o, opt.FieldManager, opt.ReasonRef)
	if err != nil {
		// conflict 审计行由 repository 落（它持有现库版本）；其余错误透传。
		return nil, err
	}
	return e.finish(objKey, rv, action)
}

// ApplyDraft 稿→面落库（apply.md §1「人确定后调用 apply 命令」）：取草稿差异走同一三步闸门，
// 以草稿 BaseRV 为乐观锁期望值——文档面前移则 Conflict（草稿保留，后写者重读后重整草稿）；
// 成功（create|update|noop）即清稿（rule 6「apply 后清空」）；无草稿 → draft.ErrNoDraft。
func (e *Engine) ApplyDraft(ctx context.Context, kind, scope, name string, opt Options) (*Result, error) {
	if opt.FieldManager == "" {
		return nil, ErrNoFieldManager
	}
	if kind == "" || scope == "" || name == "" {
		return nil, ErrObject
	}
	if e.Drafts == nil {
		return nil, draft.ErrNoStore
	}
	d, err := e.Drafts.Get(kind, scope, name)
	if err != nil {
		return nil, err // ErrNoDraft 原样透传（读面映射 404）
	}
	objKey := fmt.Sprintf("%s/%s/%s", kind, scope, name)
	// 稿即写者输入：内容=草稿正文，期望版本=落稿所基于的文档面版本。
	o := &repository.Object{Kind: kind, Scope: scope, Name: name, ResourceVersion: d.BaseRV, Content: d.Content}

	if err := e.admit(objKey, o, opt); err != nil {
		return nil, err // 拒绝路径不清稿：稿仍在，改正后再落
	}
	if opt.DryRun {
		// 零副作用：不清稿、不落库，返回与真执行一致的准入结论 + 现版本。
		cur, err := e.Repo.Get(ctx, kind, scope, name)
		if err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return &Result{RV: 0, Action: "dry-run"}, nil // 将创建
			}
			return nil, err
		}
		return &Result{RV: cur.ResourceVersion, Action: "dry-run"}, nil
	}

	rv, action, err := e.Repo.Write(ctx, o, opt.FieldManager, opt.ReasonRef)
	if err != nil {
		return nil, err // Conflict：草稿保留（乐观锁语义，apply.md §3「后写者重读」）
	}
	// rule 6「apply 后清空」：create/update/noop 都算 apply 已确认（noop=稿与面一致，
	// 空差异无操作不落库不发事件，但稿已被确认消费）。清稿失败=落库成功但缓存脏，如实上抛。
	if err := e.Drafts.Delete(kind, scope, name); err != nil {
		return nil, fmt.Errorf("apply: 落库成功（rv=%d action=%s）但清稿失败: %w", rv, action, err)
	}
	return e.finish(objKey, rv, action)
}

// admit 写路径前置闸门（① scheme 准入 ② 权限准入）；拒绝路径落 denied 审计行。
// Apply 与 ApplyDraft 共用同一条闸门——草稿不旁路 scheme/权限（apply.md：准入是写路径的一环）。
// 主体解析：Options.Subject 非 nil=请求级主体（空主体照判，匿名不享受 fallback）；
// nil=引擎缺省 fallback（e.Subject 进程级本地信任，程序化调用）。
func (e *Engine) admit(objKey string, o *repository.Object, opt Options) error {
	sub := e.Subject
	if opt.Subject != nil {
		sub = *opt.Subject // 请求级主体：非 nil 即按请求判定（空主体照判，匿名不享受 fallback）
	}
	// ① Scheme 准入：野对象拒绝（concepts/core.md：未注册 kind 不落库）。
	if e.Sch == nil {
		return fmt.Errorf("%w: scheme 注册表未配置", ErrNotRegistered)
	}
	k, err := e.Sch.Lookup(o.Kind)
	if err != nil {
		e.audit(opt, objKey, 0, "denied", err.Error(), sub)
		return fmt.Errorf("%w: %v", ErrNotRegistered, err)
	}

	// ② 权限系统准入（apply.md：准入是写路径的一环；A44 硬规则代码级不可放宽）。
	// 权限只来自文档面 Role/Group 对象（permission.System 装载面）；主体断言本身不授权。
	if e.Perm == nil {
		return fmt.Errorf("%w: 权限系统未配置", ErrDenied)
	}
	if err := e.Perm.Check(sub, o.Scope, "write", permission.ByRole); err != nil {
		e.audit(opt, objKey, 0, "denied", err.Error(), sub)
		e.emitApprovalNeeded(o.Kind, objKey, opt) // 人审档拒绝：自产 ApprovalNeeded（拒绝结论不变，事件是通知）
		return fmt.Errorf("%w: %v", ErrDenied, err)
	}
	if err := permission.HardRules(objKey, "write"); err != nil {
		e.audit(opt, objKey, 0, "denied", err.Error(), sub)
		return fmt.Errorf("%w: %v", ErrDenied, err)
	}

	// ③ spec 结构准入（v0.14）：对象文档 spec 区须过该 kind 注册声明（scheme.yaml 机读子集；
	// R-TYPE：声明即规则，checker 在 scheme 包 ValidateDoc）。dry-run 与真执行同闸门（准入结论一致）；
	// 准入先于幂等比对——坏内容即使与库内现存一致也拒绝（现存对象由声明校验器验证全合规）。
	// 解析失败不在此判（存储层 ErrContent 管辖）。⑬E 收口：坏对象从授权面兜底升为写入口拦截。
	if err := scheme.ValidateDoc(k, o.Content); err != nil {
		e.audit(opt, objKey, 0, "denied", err.Error(), sub)
		return fmt.Errorf("%w: %v", ErrSpecInvalid, err)
	}
	return nil
}

// finish 落库后的收尾：DocApplied 必发（apply.md rule 1；noop 无差异不发——防自我唤醒循环）。
// fieldManager 白名单豁免（Workflow 自写回执抑制）随 workflow 引擎接入（declared-deferred）。
func (e *Engine) finish(objKey string, rv int64, action string) (*Result, error) {
	res := &Result{RV: rv, Action: action}
	if action == "noop" {
		return res, nil
	}
	evt := eventbus.Event{
		Type:     "DocApplied",
		Source:   "apply",
		Involved: objKey,
		DedupKey: fmt.Sprintf("apply:%s:%d", objKey, rv), // 同对象同版本=同因，窗口内聚合
		Level:    "info",
	}
	if e.Bus != nil {
		if err := e.Bus.Publish(&evt); err != nil {
			return nil, fmt.Errorf("apply: DocApplied 发布失败（落库已成功 rv=%d）: %w", rv, err)
		}
	}
	res.Event = &evt
	return res, nil
}

// emitApprovalNeeded 人审档拒绝的自产事件（v0.21 引擎内部真实事件源首段）：
// 写路径权限准入拒绝 ∧ 被拒对象属人审档（permission.HumanReviewKind=授权面装载类
// Role|Group——permission.md §4 机读判定 + event.md 规则 9）时自产 ApprovalNeeded 进总线，
// 走与注入/引擎内部事件同一条主干（窗口去重/订阅派发/event_log 归档零旁路），主路
// notification/* 送达（scheme.yaml 派发行）。与注入面同主干差异三处（event.md 规则 9）：
// source 服务侧判定恒 "permission"；类型注册闸门前置——ApprovalNeeded 未注册即不发布
// （注册表是引擎内部源的唯一开关）；level 缺省=注册表 defaultLevel 投影。
// dedupKey="permission:<obj>"（注册表 dedup "<rule>:<obj>" 的机读实例：同对象待审变更
// 窗口聚合防风暴）。dry-run 零副作用不产事件；发布失败不改变拒绝结论（拒绝与审计留痕
// 是权威，事件是通知——audit 已先行落行）。hard rules 拒绝在独立分支，不经此路径。
func (e *Engine) emitApprovalNeeded(kind, objKey string, opt Options) {
	if opt.DryRun || e.Bus == nil || e.Sch == nil {
		return
	}
	if !permission.HumanReviewKind(kind) {
		return
	}
	et, err := e.Sch.LookupEventType("ApprovalNeeded")
	if err != nil {
		return // 未注册不发布（注册表闸门，声明先行的机械落点）
	}
	evt := eventbus.Event{
		Type:     "ApprovalNeeded",
		Source:   "permission",
		Involved: objKey,
		DedupKey: "permission:" + objKey,
		Level:    et.DefaultLevel, // 注册表缺省投影（ApprovalNeeded→warn）；空值由总线归一 info
	}
	_ = e.Bus.Publish(&evt) // 发布失败不改变拒绝结论（可丢语义，event.md §1）
}

// audit 拒绝路径审计（denied 行，携主体标识——permission.md 规则 3 判定留痕）。
// 审计失败不改变拒绝结论（流水非权威）。
func (e *Engine) audit(opt Options, objKey string, rv int64, result, decision string, sub permission.Subject) {
	if e.Ops == nil {
		return
	}
	_ = e.Ops.AppendApply(opslog.ApplyLog{
		Ts: time.Now(), FieldManager: opt.FieldManager, Subject: sub.ID(), Obj: objKey,
		ResourceVersion: rv, Result: result, PolicyDecisions: decision, ReasonRef: opt.ReasonRef,
	})
}
