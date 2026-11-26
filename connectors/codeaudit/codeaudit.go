// Package codeaudit connectops-code-audit —— R-TYPE-3 规则验证 code 路径执行体
// （rule/rule_type.md R-TYPE-3 三段链：Rule 条目（mode:code + checker）→ scheme
// runtime.ruleChecks（ruleId→checker 机读注册表）→ 本 ConnectorOps 实现）。
//
// 纪律（concepts/rule-verification.md 设计规则，逐条落码）：
//
//	只读——不写对象、不改目标仓文件、不提交、不 Apply（findings 只报告，修复走提案/Apply）；
//	未绑定显式失败——rule 无已注册检查器=unbound（显式列出，绝不静默通过/跳过）；
//	空审核=失败——checks/ruleChecks 均缺席或零 code 行，按失败处理不冒充通过；
//	agent 类规则不走这里——mode:agent 由调用侧丢给 agent 验证（本实现列 skipped 供对账）；
//	失败显式——checker 超时/崩溃/非零退出=error 级 finding，不降级为 warn 不静默；
//	命令执行安全——固定模板+白名单占位符（仅 {target}），直 exec 走 argv 不经 shell，
//	禁止把任意字符串拼进 shell（backend.go buildArgv 闸门）。
//
// 分发通路现状（R-ENG-3）：本包 init 自注册进 connectors.OpsConnector 注册表
// （RegisterOpsConnector——同 notify 族契约层先例）；正式通路=workflow Verify 接线
// （dev-delivery Verifying 相位硬 gate 经 ConnectorOps 调 audit/findings，待 workflow
// 侧接入）；当前最小暴露面=apiserver 调试路由 POST /v1/connectops/{name}/{op}。
package codeaudit

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/connectors"
)

// 类型/版本/op 声明集（WorkService.tools 描述面的实现侧对应——connector.md 规则 2）。
const (
	// TypeName ops 连接器类型标注（=connectops-code-audit；注册表 key）。
	TypeName = "connectops-code-audit"
	// Version 实现版本（协议/行为演进标记；只加不改）。
	Version = "v0.1.0"

	// OpAudit 逐条执行已注册检查器产出 findings（只读）。
	OpAudit = "audit"
	// OpFindings 取最近一次审核结果（auditId 回查；内存 LRU，v1 不落库）。
	OpFindings = "findings"

	// 验证模式词汇（scheme runtime.ruleChecks.mode 同源）。
	modeCode  = "code"
	modeAgent = "agent"
	modeHuman = "human"
	sevError  = "error"
	sevWarn   = "warn"
)

// ErrNoAudit findings 回查未命中（HTTP 404 投影；与对象缺席分立）。
var ErrNoAudit = errors.New("connectops-code-audit: 无此 auditId 的审核结果（LRU 容量外或从未运行）")

// Finding 一条审核发现（audit 契约的 findings 行；checkRef=检查器引用，三段链回锚）。
type Finding struct {
	Rule     string `json:"rule"`
	Severity string `json:"severity"` // error|warn（未知值归一 error——保守显式不降级）
	File     string `json:"file"`
	Line     int    `json:"line"`
	Message  string `json:"message"`
	CheckRef string `json:"checkRef"`
}

// Unbound 一条未绑定（code 规则无已注册检查器=未生效——不是"跳过"，Verify 必须显式报）。
type Unbound struct {
	Rule     string `json:"rule"`
	CheckRef string `json:"checkRef,omitempty"`
	Reason   string `json:"reason"`
}

// Skipped 一条不归本实现管的规则行（mode:agent/human——调用侧路由，列出让对账可见）。
type Skipped struct {
	Rule   string `json:"rule"`
	Mode   string `json:"mode"`
	Reason string `json:"reason"`
}

// ToolInfo 工具版本信息（rule-verification.md 规则 9 可复现：版本随结果记录）。
type ToolInfo struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

// Result audit op 整体结果。ok=无 error 级 finding 且 unbound 空——warn 不翻转 ok
// （rule-verification.md：error 阻断，warn 进 Report）；unbound 非空恒 ok=false。
type Result struct {
	OK       bool      `json:"ok"`
	AuditID  string    `json:"auditId"`
	Findings []Finding `json:"findings"`
	Unbound  []Unbound `json:"unbound"`
	Skipped  []Skipped `json:"skipped,omitempty"`
	RanAt    string    `json:"ranAt"`
	Tool     ToolInfo  `json:"tool"`
}

// auditCheck checks 数组元素（调用侧 curated 的审核项；rule-verification.md §3.1 条目投影）。
type auditCheck struct {
	Rule     string `json:"rule"`               // ruleId（Rule/Scheme/实现三段同锚）
	Ref      string `json:"ref,omitempty"`      // 规则原文引用（如 rule/rule.md#R-PUB-1）
	Mode     string `json:"mode,omitempty"`     // 缺省 code；显式非 code=skipped（agent 规则丢给 agent）
	Checker  string `json:"checker,omitempty"`  // 检查器引用（builtin 名或 config 命令模板名）
	Severity string `json:"severity,omitempty"` // error|warn，缺省 error
}

// ruleCheckRow scheme runtime.ruleChecks 投影行（checks 缺省时的注册表输入；
// checks/ruleChecks 皆缺=空审核按失败——绝不静默通过）。
type ruleCheckRow struct {
	RuleID   string   `json:"ruleId"`
	Mode     string   `json:"mode"`
	Checker  string   `json:"checker,omitempty"`
	Severity string   `json:"severity,omitempty"`
	Targets  []string `json:"targets,omitempty"`
}

// auditRequest audit op 入参（JSON）。
type auditRequest struct {
	Target     string         `json:"target"`     // 被审目录路径（绝对路径，只读遍历）
	Checks     []auditCheck   `json:"checks"`     // 缺省读 ruleChecks（scheme 投影）
	RuleChecks []ruleCheckRow `json:"ruleChecks"` // scheme runtime.ruleChecks 投影
	SchemePath string         `json:"schemePath"` // cored-lint 后端所需注册表（可选；缺省用装配侧）
}

// findingsRequest findings op 入参（JSON）。
type findingsRequest struct {
	AuditID string `json:"auditId"`
}

// checkEntry 归一后的单条待审项（resolveChecks 输出；执行单元）。
type checkEntry struct {
	rule, ref, checker, severity string
}

// Connector connectops-code-audit 实现（connectors.OpsConnector 契约五方法）。
// 可安全零配置构建（builtin 检查器在册；模板/超时=装配侧 Option 注入）。
type Connector struct {
	mu         sync.Mutex
	builtin    map[string]CheckFunc
	templates  map[string]string // config 声明的命令模板（checker 引用→模板；部署侧固定）
	schemePath string            // cored-lint 后端缺省 scheme 路径（装配侧）
	timeout    time.Duration     // 单次审核整体超时（缺省 10m——ruleChecks scheme.params 同源）
	recent     *resultLRU
}

// Option 装配选项（部署侧 config 投影；零值缺省安全）。
type Option func(*Connector)

// WithTemplate 声明命令模板后端（config 声明命令模板，占位符白名单见 buildArgv）。
func WithTemplate(checkerRef, template string) Option {
	return func(c *Connector) { c.templates[checkerRef] = template }
}

// WithSchemePath 注入 cored-lint 后端缺省 scheme 路径（请求 params.schemePath 优先）。
func WithSchemePath(path string) Option {
	return func(c *Connector) { c.schemePath = path }
}

// WithTimeout 注入单次审核整体超时（context 截止；缺省 10m）。
func WithTimeout(d time.Duration) Option {
	return func(c *Connector) { c.timeout = d }
}

// New 构建实现（builtin 检查器已注册；可无凭证/无配置安全构建——契约层同口径）。
func New(opts ...Option) *Connector {
	c := &Connector{
		builtin:   map[string]CheckFunc{CheckerCoreLint: coreLintCheck},
		templates: map[string]string{},
		timeout:   10 * time.Minute,
		recent:    newResultLRU(64),
	}
	for _, f := range opts {
		f(c)
	}
	return c
}

func init() { connectors.RegisterOpsConnector(New()) }

// Type ops 连接器类型标注（注册表 key）。
func (c *Connector) Type() string { return TypeName }

// Version 实现版本。
func (c *Connector) Version() string { return Version }

// Ops 提供的 op 集（⊆ WorkService.tools 声明——两 op 均只读，undo:none 不适用）。
func (c *Connector) Ops() []string { return []string{OpAudit, OpFindings} }

// Probe 通道可达性探测（本地进程内实现恒可达；不执行任何检查器）。
func (c *Connector) Probe(_ context.Context) error { return nil }

// Invoke 执行一次 op（调用即效果；本实现的效果=只读审核与结果回查，零写面）。
func (c *Connector) Invoke(ctx context.Context, op string, params []byte) ([]byte, error) {
	switch op {
	case OpAudit:
		return c.invokeAudit(ctx, params)
	case OpFindings:
		return c.invokeFindings(params)
	default:
		return nil, fmt.Errorf("%s: op %q 不在声明集 [audit findings]", TypeName, op)
	}
}

// invokeAudit audit op 主体：解析待审集 → 逐条执行已注册检查器 → 汇总 findings/unbound
// → 落 LRU 供 findings 回查。任何一步都不产生"默认通过"：无检查器=unbound、执行失败=
// error finding、空待审集=unbound（合成行说明）。
func (c *Connector) invokeAudit(ctx context.Context, params []byte) ([]byte, error) {
	var req auditRequest
	if len(params) > 0 {
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("%s: params 不可解析: %w", TypeName, err)
		}
	}
	if req.Target == "" {
		return nil, fmt.Errorf("%s: target 必填（被审目录路径）", TypeName)
	}
	if !strings.HasPrefix(req.Target, "/") {
		return nil, fmt.Errorf("%s: target 须为绝对路径，得到 %q", TypeName, req.Target)
	}
	// 只读红线的前置防御：目标必须存在且为目录（不存在=显式失败，不冒充零发现）。
	if fi, err := os.Stat(req.Target); err != nil || !fi.IsDir() {
		return nil, fmt.Errorf("%s: target %q 不可达或非目录（审核面零发现≠目录缺席）", TypeName, req.Target)
	}

	entries, unbound, skipped, err := c.resolveChecks(req)
	if err != nil {
		return nil, err
	}

	// 整体超时（context 截止；checker 侧经 CommandContext/切片继承）。
	ctx, cancel := context.WithTimeout(ctx, c.timeout)
	defer cancel()

	findings := c.runChecks(ctx, req, entries)

	// 排序确定性（机械对账锚点）：rule→file→line→message。
	sort.SliceStable(findings, func(i, j int) bool {
		a, b := findings[i], findings[j]
		if a.Rule != b.Rule {
			return a.Rule < b.Rule
		}
		if a.File != b.File {
			return a.File < b.File
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.Message < b.Message
	})

	ok := len(unbound) == 0
	for _, f := range findings {
		if f.Severity == sevError {
			ok = false
			break
		}
	}
	res := Result{
		OK:       ok,
		AuditID:  auditID(req.Target, entries),
		Findings: findings,
		Unbound:  unbound,
		Skipped:  skipped,
		RanAt:    time.Now().UTC().Format(time.RFC3339),
		Tool:     ToolInfo{Name: TypeName, Version: Version},
	}
	c.recent.put(res)
	out, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("%s: 结果编码失败: %w", TypeName, err)
	}
	return out, nil
}

// resolveChecks 待审集解析：checks 优先；缺省投影 ruleChecks。mode 非 code 行不归本实现
// （agent=丢给 agent、human=人审）——列 skipped 供对账，绝不静默。空集=unbound 合成行
// （空审核按失败：Verify 硬 gate 不允许"什么都没审"冒充通过）。
func (c *Connector) resolveChecks(req auditRequest) (entries []checkEntry, unbound []Unbound, skipped []Skipped, err error) {
	unbound = []Unbound{}
	skipped = []Skipped{}
	addEntry := func(rule, ref, checker, severity, mode string) {
		if mode != "" && mode != modeCode {
			skipped = append(skipped, Skipped{Rule: rule, Mode: mode,
				Reason: "mode=" + mode + " 规则不经 connectops-code-audit（agent=丢给 agent 验证、human=人审）——调用侧路由"})
			return
		}
		if severity == "" {
			severity = sevError
		}
		if severity != sevError && severity != sevWarn {
			severity = sevError // 未知等级归一 error（保守显式，不降级）
		}
		entries = append(entries, checkEntry{rule: rule, ref: ref, checker: checker, severity: severity})
	}
	switch {
	case len(req.Checks) > 0:
		for _, ck := range req.Checks {
			if ck.Rule == "" {
				return nil, nil, nil, fmt.Errorf("%s: checks 项缺 rule（ruleId 三段链锚）", TypeName)
			}
			addEntry(ck.Rule, ck.Ref, ck.Checker, ck.Severity, ck.Mode)
		}
	case len(req.RuleChecks) > 0:
		for _, row := range req.RuleChecks {
			if row.RuleID == "" {
				return nil, nil, nil, fmt.Errorf("%s: ruleChecks 行缺 ruleId", TypeName)
			}
			addEntry(row.RuleID, "", row.Checker, row.Severity, row.Mode)
		}
	default:
		unbound = append(unbound, Unbound{Rule: "*",
			Reason: "checks/ruleChecks 均未提供或无 code 行——空审核按失败处理（绝不静默通过）"})
		return entries, unbound, skipped, nil
	}

	// 后端解析：builtin 在册→命令模板在册→unbound（显式列出，不静默跳过）。
	resolved := entries[:0]
	for _, e := range entries {
		if _, ok := c.builtin[e.checker]; ok {
			resolved = append(resolved, e)
			continue
		}
		if _, ok := c.templates[e.checker]; ok {
			resolved = append(resolved, e)
			continue
		}
		unbound = append(unbound, Unbound{Rule: e.rule, CheckRef: e.checker,
			Reason: "检查器未注册（builtin/config 命令模板均未命中）——mode:code 规则无 checker=未生效，不是跳过"})
	}
	return resolved, unbound, skipped, nil
}

// runChecks 并发执行已解析检查器（并发上限 4——ruleChecks scheme.params.concurrency 同源）。
// 后端失败（超时/崩溃/非零退出/输出不合契约）→ 合成 error 级 finding（失败显式，不静默放行）。
func (c *Connector) runChecks(ctx context.Context, req auditRequest, entries []checkEntry) []Finding {
	findings := []Finding{}
	if len(entries) == 0 {
		return findings
	}
	var (
		mu  sync.Mutex
		wg  sync.WaitGroup
		sem = make(chan struct{}, 4)
	)
	for _, e := range entries {
		wg.Add(1)
		go func(e checkEntry) {
			defer wg.Done()
			sem <- struct{}{}
			defer func() { <-sem }()
			fs, err := c.invokeCheck(ctx, req, e)
			mu.Lock()
			defer mu.Unlock()
			if err != nil {
				findings = append(findings, Finding{
					Rule: e.rule, Severity: sevError, File: req.Target, Line: 0,
					Message:  "checker " + e.checker + " 执行失败: " + err.Error(),
					CheckRef: e.checker,
				})
				return
			}
			findings = append(findings, fs...)
		}(e)
	}
	wg.Wait()
	return findings
}

// invokeCheck 单条检查执行：选后端（builtin→命令模板）→ 归一 findings（rule/checkRef 回锚、
// severity 归一继承条目等级）。
func (c *Connector) invokeCheck(ctx context.Context, req auditRequest, e checkEntry) ([]Finding, error) {
	var fn CheckFunc
	if f, ok := c.builtin[e.checker]; ok {
		fn = f
	} else {
		fn = c.templateCheck(c.templates[e.checker])
	}
	schemePath := req.SchemePath
	if schemePath == "" {
		schemePath = c.schemePath
	}
	fs, err := fn(ctx, CheckRequest{
		Rule: e.rule, CheckRef: e.checker, Severity: e.severity,
		Target: req.Target, SchemePath: schemePath,
	})
	if err != nil {
		return nil, err
	}
	out := make([]Finding, 0, len(fs))
	for _, f := range fs {
		if f.Rule == "" {
			f.Rule = e.rule
		}
		if f.CheckRef == "" {
			f.CheckRef = e.checker
		}
		if f.Severity != sevError && f.Severity != sevWarn {
			f.Severity = e.severity // 后端未标/标了未知等级→继承条目等级（缺省 error，保守）
		}
		out = append(out, f)
	}
	return out, nil
}

// invokeFindings findings op 主体：auditId 回查最近结果（内存 LRU，v1 不落库——正式
// 归档=VerifyDoc/apply_log 链，随 workflow Verify 接线落）。
func (c *Connector) invokeFindings(params []byte) ([]byte, error) {
	var req findingsRequest
	if len(params) > 0 {
		if err := json.Unmarshal(params, &req); err != nil {
			return nil, fmt.Errorf("%s: params 不可解析: %w", TypeName, err)
		}
	}
	if req.AuditID == "" {
		return nil, fmt.Errorf("%s: auditId 必填", TypeName)
	}
	res, ok := c.recent.get(req.AuditID)
	if !ok {
		return nil, ErrNoAudit
	}
	out, err := json.Marshal(res)
	if err != nil {
		return nil, fmt.Errorf("%s: 结果编码失败: %w", TypeName, err)
	}
	return out, nil
}

// auditID 结果句柄：时间桶 + 输入摘要前 8 位（同输入同窗同 id——可复现口径的轻量锚）。
func auditID(target string, entries []checkEntry) string {
	h := sha256.New()
	h.Write([]byte(target))
	for _, e := range entries {
		h.Write([]byte{0})
		h.Write([]byte(e.rule + "|" + e.checker))
	}
	sum := hex.EncodeToString(h.Sum(nil))[:8]
	return "audit-" + time.Now().UTC().Format("20060102T150405Z") + "-" + sum
}

// ListRecent 返回最近审核结果副本（最新在前——只读投影 GET /v1/code-audit/findings 与快照共用）。
func (c *Connector) ListRecent() []Result {
	if c == nil || c.recent == nil {
		return []Result{}
	}
	return c.recent.list()
}

// resultLRU 最近结果环形账（内存容量 64；溢出淘汰最旧——findings 回查容量外语义=ErrNoAudit）。
type resultLRU struct {
	mu  sync.Mutex
	cap int
	ids []string
	m   map[string]Result
}

func newResultLRU(cap int) *resultLRU {
	return &resultLRU{cap: cap, m: map[string]Result{}}
}

func (l *resultLRU) put(r Result) {
	l.mu.Lock()
	defer l.mu.Unlock()
	// 同 id 重放（同秒同输入）：移除旧位再入队，不重复占位。
	for i, id := range l.ids {
		if id == r.AuditID {
			l.ids = append(l.ids[:i], l.ids[i+1:]...)
			break
		}
	}
	l.ids = append(l.ids, r.AuditID)
	l.m[r.AuditID] = r
	for len(l.ids) > l.cap {
		delete(l.m, l.ids[0])
		l.ids = l.ids[1:]
	}
}

func (l *resultLRU) get(id string) (Result, bool) {
	l.mu.Lock()
	defer l.mu.Unlock()
	r, ok := l.m[id]
	return r, ok
}

func (l *resultLRU) list() []Result {
	l.mu.Lock()
	defer l.mu.Unlock()
	out := make([]Result, 0, len(l.ids))
	for i := len(l.ids) - 1; i >= 0; i-- {
		if r, ok := l.m[l.ids[i]]; ok {
			out = append(out, r)
		}
	}
	return out
}
