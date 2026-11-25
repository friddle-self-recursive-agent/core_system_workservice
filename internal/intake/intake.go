// Package intake —— R-UX-1 通用对话式录入与三重 dry-run 准入预检引擎（能力 D，2026-11-24 用户裁决）。
// 设计依据：
//   - core_design/rule.md R-UX-1（人机交互默认对话，固定六步顺序：自由描述+自动上下文 → 自动补充 →
//     缺少询问（只问不明确/缺失/冲突字段，带默认值与示例）→ 提交节点生成 spec 对象 + lint/准入 dry-run
//     （失败精确定位到字段）→ 确认 diff 后系统自动补全并 Apply → 降级表单同一对象无双源）
//   - core_design/rule.md R-UX-2（scheme 驱动编辑控件：字段集/类型/必填/枚举/默认值一律取自 scheme 单一真相）
//   - core_design/concepts/proposal.md §4.5 & core_system_instance/runtime/console-ui.md §12
//   - core_system_instance/runtime/cored-bridge.md §2.5
//
// 审计与脱敏红线：
//   - 对话→对象过程留痕（sessionId / kind / turns / genVersion / lastDryRunOk / reasonRef），
//     对象经 Apply 管线入库才算数；对话原文绝不写入 docsRoot 公共库（R-PUB-1/R-PUB-3）。
package intake

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/apply"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/docrepo"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"gopkg.in/yaml.v3"
)

var (
	// ErrNotRegistered 目标 kind 未在 Scheme 注册 → 404。
	ErrNotRegistered = errors.New("intake: kind 未在 Scheme 注册")
	// ErrSessionNotFound 会话不存在 → 404。
	ErrSessionNotFound = errors.New("intake: 会话不存在")
	// ErrInvalid 请求参数不合法 → 400。
	ErrInvalid = errors.New("intake: 请求参数不合法（需 kind）")
)

// ObjectReader 文档库只读接口（用于上下文感知与 ref 引用存在性检查）。
type ObjectReader interface {
	Get(ctx context.Context, kind, scope, name string) (*repository.Object, error)
	List(ctx context.Context, kind, scope string) ([]*repository.Object, error)
}

// DryRunApplier Apply 引擎接口（用于提交节点 dry-run 权限与准入验证）。
type DryRunApplier interface {
	Apply(ctx context.Context, o *repository.Object, opt apply.Options) (*apply.Result, error)
}

// FieldState 字段卡（右 1/3）单字段实时状态投影。
type FieldState struct {
	Path     string   `json:"path"`
	Name     string   `json:"name"`
	Label    string   `json:"label"`
	Widget   string   `json:"widget"`
	Type     string   `json:"type"`
	Required bool     `json:"required"`
	Enum     []string `json:"enum,omitempty"`
	Default  any      `json:"default,omitempty"`
	Example  string   `json:"example,omitempty"`
	Value    any      `json:"value,omitempty"`
	Status   string   `json:"status"`           // filled (● 已填) | default (◐ 默认/已推断) | missing (○ 待补) | conflict (✕ 冲突/校验未过)
	Source   string   `json:"source,omitempty"` // user | extracted | default | subject | system
	Error    string   `json:"error,omitempty"`
}

// FieldError dry-run 失败时的字段级归因（精确定位到具体字段）。
type FieldError struct {
	Field   string `json:"field"`
	Rule    string `json:"rule"`
	Message string `json:"message"`
}

// SessionRecord 对话录入会话审计留痕（内存态，不落 docsRoot 公共库，且不含对话原文）。
type SessionRecord struct {
	SessionID     string            `json:"sessionId"`
	Kind          string            `json:"kind"`
	Scope         string            `json:"scope"`
	Name          string            `json:"name"`
	Proposer      string            `json:"proposer"`
	Turns         int               `json:"turns"`
	GenVersion    int               `json:"genVersion"`
	AskingField   string            `json:"askingField,omitempty"`
	Ready         bool              `json:"ready"`
	LastDryRunOk  bool              `json:"lastDryRunOk"`
	FieldErrorCnt int               `json:"fieldErrorCount"`
	ReasonRef     string            `json:"reasonRef,omitempty"`
	Fields        map[string]any    `json:"fields,omitempty"`
	Sources       map[string]string `json:"sources,omitempty"`
	CreatedAt     time.Time         `json:"createdAt"`
	UpdatedAt     time.Time         `json:"updatedAt"`
}

// TurnRequest POST /v1/intake/turn 请求。
type TurnRequest struct {
	SessionID   string              `json:"sessionId"`
	Kind        string              `json:"kind"`
	Scope       string              `json:"scope"`
	Name        string              `json:"name"`
	Message     string              `json:"message"`
	SkipCurrent bool                `json:"skipCurrent"`
	Fields      map[string]any      `json:"fields"`
	Subject     *permission.Subject `json:"-"`
}

// TurnResponse POST /v1/intake/turn 响应。
type TurnResponse struct {
	SessionID      string                   `json:"sessionId"`
	Kind           string                   `json:"kind"`
	Scope          string                   `json:"scope"`
	Name           string                   `json:"name"`
	Proposer       string                   `json:"proposer"`
	Turns          int                      `json:"turns"`
	Ready          bool                     `json:"ready"`
	AskingField    string                   `json:"askingField"`
	Reply          string                   `json:"reply"`
	Extracted      map[string]any           `json:"extracted"`
	Fields         map[string]any           `json:"fields"`
	FieldStates    []FieldState             `json:"fieldStates"`
	MissingFields  []string                 `json:"missingFields"`
	ConflictFields []string                 `json:"conflictFields"`
	Widgets        []scheme.FieldDescriptor `json:"widgets"`
}

// GenerateRequest POST /v1/intake/generate 请求。
type GenerateRequest struct {
	SessionID string              `json:"sessionId"`
	Kind      string              `json:"kind"`
	Scope     string              `json:"scope"`
	Name      string              `json:"name"`
	Fields    map[string]any      `json:"fields"`
	Subject   *permission.Subject `json:"-"`
}

// GenerateResponse POST /v1/intake/generate 响应。
type GenerateResponse struct {
	Ok                  bool           `json:"ok"`
	SessionID           string         `json:"sessionId"`
	GenVersion          int            `json:"genVersion"`
	Kind                string         `json:"kind"`
	Scope               string         `json:"scope"`
	Name                string         `json:"name"`
	YAML                string         `json:"yaml"`
	BaseResourceVersion int64          `json:"baseResourceVersion"`
	Checks              map[string]any `json:"checks"`
	FieldErrors         []FieldError   `json:"fieldErrors"`
	AutoCompleted       map[string]any `json:"autoCompleted"`
	TriagePreview       string         `json:"triagePreview"`
	ReasonRef           string         `json:"reasonRef"`
	Generator           string         `json:"generator"`
}

// Engine R-UX-1 对话式录入与准入预检引擎。
type Engine struct {
	Sch         *scheme.Scheme
	Objects     ObjectReader
	Applier     DryRunApplier
	AutoApprove []string // Proposal 自动批准优先级清单（用于预览审批分流）

	mu       sync.RWMutex
	seq      atomic.Uint64
	sessions map[string]*SessionRecord
}

// NewEngine 构造对话录入引擎。
func NewEngine(sch *scheme.Scheme, objects ObjectReader, applier DryRunApplier, autoApprove []string) *Engine {
	return &Engine{
		Sch:         sch,
		Objects:     objects,
		Applier:     applier,
		AutoApprove: append([]string(nil), autoApprove...),
		sessions:    make(map[string]*SessionRecord),
	}
}

// ListSessions 返回会话审计留痕清单（按更新时间降序，不含对话原文）。
func (e *Engine) ListSessions() []SessionRecord {
	e.mu.RLock()
	defer e.mu.RUnlock()
	out := make([]SessionRecord, 0, len(e.sessions))
	for _, s := range e.sessions {
		cp := *s
		cp.Fields = nil
		cp.Sources = nil
		out = append(out, cp)
	}
	sort.Slice(out, func(i, j int) bool {
		if !out[i].UpdatedAt.Equal(out[j].UpdatedAt) {
			return out[i].UpdatedAt.After(out[j].UpdatedAt)
		}
		return out[i].SessionID > out[j].SessionID
	})
	return out
}

// GetSession 查询单条会话审计留痕与字段快照。
func (e *Engine) GetSession(id string) (*SessionRecord, error) {
	e.mu.RLock()
	defer e.mu.RUnlock()
	s, ok := e.sessions[id]
	if !ok {
		return nil, ErrSessionNotFound
	}
	cp := *s
	cp.Fields = cloneMap(s.Fields)
	cp.Sources = cloneStrMap(s.Sources)
	return &cp, nil
}

// Turn 执行一轮 R-UX-1 对话推进（步骤 1 自由描述+自动上下文 → 步骤 2 自动补充 → 步骤 3 缺少询问）。
func (e *Engine) Turn(ctx context.Context, req TurnRequest) (*TurnResponse, error) {
	kindName := strings.TrimSpace(req.Kind)
	if kindName == "" {
		kindName = "Proposal"
	}
	if e.Sch == nil {
		return nil, ErrNotRegistered
	}
	k, err := e.Sch.Lookup(kindName)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotRegistered, err)
	}
	proposer := "console-operator"
	if req.Subject != nil && strings.TrimSpace(req.Subject.User) != "" {
		proposer = strings.TrimSpace(req.Subject.User)
	}

	e.mu.Lock()
	defer e.mu.Unlock()

	now := time.Now().UTC()
	sess, exists := e.sessions[req.SessionID]
	if !exists || sess == nil {
		sid := req.SessionID
		if sid == "" {
			n := e.seq.Add(1)
			sid = fmt.Sprintf("intake-sess-%d-%d", now.Unix(), n)
		}
		scope := strings.TrimSpace(req.Scope)
		if scope == "" {
			scope = "global"
		}
		sess = &SessionRecord{
			SessionID: sid,
			Kind:      k.Name,
			Scope:     scope,
			Name:      strings.TrimSpace(req.Name),
			Proposer:  proposer,
			Fields:    make(map[string]any),
			Sources:   make(map[string]string),
			CreatedAt: now,
			UpdatedAt: now,
		}
		e.sessions[sid] = sess
	}
	if req.Scope != "" {
		sess.Scope = strings.TrimSpace(req.Scope)
	}
	if req.Name != "" {
		sess.Name = strings.TrimSpace(req.Name)
	}
	sess.Proposer = proposer
	sess.Turns++
	sess.UpdatedAt = now

	flatW := flattenWidgets(k.Name, k.Widgets)

	// 1. 合并字段卡手动修改（双向同步）
	for fk, fv := range req.Fields {
		if fk == "landing" {
			if m, ok := fv.(map[string]any); ok {
				if t, ok2 := m["target"]; ok2 {
					sess.Fields["landing.target"] = t
					sess.Sources["landing.target"] = "user"
				}
				if r, ok2 := m["ref"]; ok2 {
					sess.Fields["landing.ref"] = r
					sess.Sources["landing.ref"] = "user"
				}
				continue
			}
		}
		sess.Fields[fk] = normalizeFieldValue(findWidget(flatW, fk), fv)
		sess.Sources[fk] = "user"
	}

	// 2. R-UX-1 步骤 2：自动补充（能从上下文/scheme 推断的直接填，不重复盘问）
	var autoFilledNotes []string
	for _, w := range flatW {
		if !isEmptyValue(sess.Fields[w.Path]) {
			continue
		}
		switch w.AutoFill {
		case "subject":
			sess.Fields[w.Path] = proposer
			sess.Sources[w.Path] = "subject"
			autoFilledNotes = append(autoFilledNotes, fmt.Sprintf("%s=%s", w.Label, proposer))
		case "system":
			if w.Default != nil {
				sess.Fields[w.Path] = w.Default
				sess.Sources[w.Path] = "system"
			}
		case "default":
			if w.Default != nil {
				sess.Fields[w.Path] = w.Default
				sess.Sources[w.Path] = "default"
				autoFilledNotes = append(autoFilledNotes, fmt.Sprintf("%s=%v", w.Label, w.Default))
			}
		}
	}

	// 3. R-UX-1 步骤 1：处理本轮用户自由描述 / 回答 / 跳过
	extracted := make(map[string]any)
	msg := strings.TrimSpace(req.Message)
	skipWord := msg == "跳过" || msg == "默认" || strings.EqualFold(msg, "skip") || strings.EqualFold(msg, "default")
	if (req.SkipCurrent || skipWord) && sess.AskingField != "" {
		w := findWidget(flatW, sess.AskingField)
		if w != nil {
			defVal := w.Default
			if isEmptyValue(defVal) {
				if w.Widget == "list" {
					defVal = parseListText(w.Example)
				} else {
					defVal = w.Example
				}
			}
			sess.Fields[w.Path] = defVal
			sess.Sources[w.Path] = "default"
			extracted[w.Path] = defVal
		}
	} else if msg != "" {
		ext := extractFieldsFromText(k.Name, flatW, msg, sess.AskingField, sess.Fields)
		for ek, ev := range ext {
			sess.Fields[ek] = ev
			sess.Sources[ek] = "extracted"
			extracted[ek] = ev
		}
	}

	// 自动派生对象名称 slug
	if sess.Name == "" {
		if titleStr, ok := sess.Fields["title"].(string); ok && strings.TrimSpace(titleStr) != "" {
			sess.Name = deriveSlug(titleStr, sess.SessionID)
		}
	}

	// 4. R-UX-1 步骤 3：计算字段状态、缺失项与冲突项（只问缺失/不明确/冲突字段）
	fieldStates, missing, conflicts := evaluateFields(ctx, k.Name, flatW, sess.Fields, sess.Sources, e.Objects)
	sess.Ready = len(missing) == 0 && len(conflicts) == 0
	nextAsk := ""
	if len(conflicts) > 0 {
		nextAsk = conflicts[0]
	} else if len(missing) > 0 {
		nextAsk = missing[0]
	}
	sess.AskingField = nextAsk

	reply := composeReply(k.Name, flatW, sess, autoFilledNotes, extracted, fieldStates, missing, conflicts, nextAsk)

	return &TurnResponse{
		SessionID:      sess.SessionID,
		Kind:           k.Name,
		Scope:          sess.Scope,
		Name:           sess.Name,
		Proposer:       sess.Proposer,
		Turns:          sess.Turns,
		Ready:          sess.Ready,
		AskingField:    nextAsk,
		Reply:          reply,
		Extracted:      extracted,
		Fields:         cloneMap(sess.Fields),
		FieldStates:    fieldStates,
		MissingFields:  missing,
		ConflictFields: conflicts,
		Widgets:        k.Widgets,
	}, nil
}

// Generate 执行 R-UX-1 步骤 4 & 5：提交节点生成标准对象（rv=0）+ 系统补缺省 + 三重 dry-run（失败定位到具体字段）。
func (e *Engine) Generate(ctx context.Context, req GenerateRequest) (*GenerateResponse, error) {
	kindName := strings.TrimSpace(req.Kind)
	if kindName == "" {
		kindName = "Proposal"
	}
	if e.Sch == nil {
		return nil, ErrNotRegistered
	}
	k, err := e.Sch.Lookup(kindName)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotRegistered, err)
	}
	proposer := "console-operator"
	if req.Subject != nil && strings.TrimSpace(req.Subject.User) != "" {
		proposer = strings.TrimSpace(req.Subject.User)
	}

	e.mu.Lock()
	now := time.Now().UTC()
	sess, exists := e.sessions[req.SessionID]
	if !exists || sess == nil {
		sid := req.SessionID
		if sid == "" {
			n := e.seq.Add(1)
			sid = fmt.Sprintf("intake-sess-%d-%d", now.Unix(), n)
		}
		scope := strings.TrimSpace(req.Scope)
		if scope == "" {
			scope = "global"
		}
		sess = &SessionRecord{
			SessionID: sid,
			Kind:      k.Name,
			Scope:     scope,
			Name:      strings.TrimSpace(req.Name),
			Proposer:  proposer,
			Fields:    make(map[string]any),
			Sources:   make(map[string]string),
			CreatedAt: now,
			UpdatedAt: now,
		}
		if sess.Turns == 0 {
			sess.Turns = 1
		}
		e.sessions[sid] = sess
	}
	if req.Scope != "" {
		sess.Scope = strings.TrimSpace(req.Scope)
	}
	if req.Name != "" {
		sess.Name = strings.TrimSpace(req.Name)
	}
	sess.Proposer = proposer
	flatW := flattenWidgets(k.Name, k.Widgets)
	for fk, fv := range req.Fields {
		if fk == "landing" {
			if m, ok := fv.(map[string]any); ok {
				if t, ok2 := m["target"]; ok2 {
					sess.Fields["landing.target"] = t
				}
				if r, ok2 := m["ref"]; ok2 {
					sess.Fields["landing.ref"] = r
				}
				continue
			}
		}
		sess.Fields[fk] = normalizeFieldValue(findWidget(flatW, fk), fv)
	}

	// 系统自动补全缺省字段（proposal.md §4.5 第 5 条：proposer/status/timestamps + scheme 默认值）
	autoCompleted := map[string]any{
		"timestamp": now.Format(time.RFC3339),
	}
	for _, w := range flatW {
		if isEmptyValue(sess.Fields[w.Path]) {
			switch w.AutoFill {
			case "subject":
				sess.Fields[w.Path] = proposer
				autoCompleted[w.Path] = proposer
			case "system", "default":
				if w.Default != nil {
					sess.Fields[w.Path] = w.Default
					autoCompleted[w.Path] = w.Default
				}
			}
		}
	}
	if k.Name == "Proposal" {
		sess.Fields["proposer"] = proposer
		autoCompleted["proposer"] = proposer
		if isEmptyValue(sess.Fields["status"]) {
			sess.Fields["status"] = "submitted"
		}
		autoCompleted["status"] = sess.Fields["status"]
		if isEmptyValue(sess.Fields["priority"]) {
			sess.Fields["priority"] = "low"
		}
		autoCompleted["priority"] = sess.Fields["priority"]
		if isEmptyValue(sess.Fields["landing.target"]) {
			sess.Fields["landing.target"] = "doc"
		}
		autoCompleted["landing.target"] = sess.Fields["landing.target"]
	}
	if sess.Name == "" {
		titleStr, _ := sess.Fields["title"].(string)
		sess.Name = deriveSlug(titleStr, sess.SessionID)
	}
	sess.GenVersion++
	sess.UpdatedAt = now
	genVer := sess.GenVersion
	sid := sess.SessionID
	scope := sess.Scope
	name := sess.Name
	fieldsCopy := cloneMap(sess.Fields)
	e.mu.Unlock()

	var fieldErrors []FieldError

	// ① 逐字段 R-PUB-1 lint + secret 引用式前缀检查（错误定位到具体字段）
	var rpubViolations []map[string]any
	fieldKeys := make([]string, 0, len(fieldsCopy))
	for fk := range fieldsCopy {
		fieldKeys = append(fieldKeys, fk)
	}
	sort.Strings(fieldKeys)
	for _, fk := range fieldKeys {
		txt := valueToText(fieldsCopy[fk])
		if txt == "" {
			continue
		}
		for _, v := range docrepo.Lint([]byte(txt)) {
			rule := "R-PUB-1:" + v.Rule
			msg := fmt.Sprintf("字段 %s 违反 R-PUB-1（%s）：命中 %q（公开仓禁止真实 IP/部署路径/个人标识）", fk, v.Rule, v.Text)
			fieldErrors = append(fieldErrors, FieldError{Field: fk, Rule: rule, Message: msg})
			rpubViolations = append(rpubViolations, map[string]any{"field": fk, "line": v.Line, "rule": v.Rule, "text": v.Text})
		}
		if secMsg := checkSecretRefDiscipline(fk, txt); secMsg != "" {
			fieldErrors = append(fieldErrors, FieldError{Field: fk, Rule: "R-PUB-1:secret", Message: secMsg})
			rpubViolations = append(rpubViolations, map[string]any{"field": fk, "rule": "secret", "text": secMsg})
		}
	}

	// 构造标准 YAML 对象（rv=0）
	yamlStr := buildObjectYAML(k, scope, name, fieldsCopy)

	// ② Scheme spec 准入校验 + 必填/枚举/引用存在性检查（错误定位到具体字段）
	var specViolations []string
	for _, w := range flatW {
		if w.Required && isEmptyValue(fieldsCopy[w.Path]) {
			msg := fmt.Sprintf("%s (%s): 必填段缺失", w.Path, w.Label)
			specViolations = append(specViolations, msg)
			fieldErrors = appendFieldErrorUnique(fieldErrors, FieldError{Field: w.Path, Rule: "scheme:required", Message: msg})
		}
		if w.Widget == "ref" && !isEmptyValue(fieldsCopy[w.Path]) && e.Objects != nil && w.RefKind != "" {
			refVal := strings.TrimSpace(fmt.Sprint(fieldsCopy[w.Path]))
			if refVal != "" && !checkRefExists(ctx, e.Objects, w.RefKind, scope, refVal) {
				msg := fmt.Sprintf("%s: 引用目标 %q (%s) 不在册", w.Path, refVal, w.RefKind)
				specViolations = append(specViolations, msg)
				fieldErrors = appendFieldErrorUnique(fieldErrors, FieldError{Field: w.Path, Rule: "scheme:ref", Message: msg})
			}
		}
	}
	if err := scheme.ValidateDoc(k, []byte(yamlStr)); err != nil {
		var se *scheme.SpecError
		if errors.As(err, &se) {
			for _, v := range se.Violations {
				specViolations = append(specViolations, v)
				fld := v
				if idx := strings.Index(v, ":"); idx > 0 {
					fld = strings.TrimSpace(v[:idx])
				}
				if idx := strings.Index(fld, "["); idx > 0 {
					fld = fld[:idx]
				}
				fieldErrors = appendFieldErrorUnique(fieldErrors, FieldError{Field: fld, Rule: "scheme:spec", Message: v})
			}
		} else {
			specViolations = append(specViolations, err.Error())
			fieldErrors = appendFieldErrorUnique(fieldErrors, FieldError{Field: "spec", Rule: "scheme:spec", Message: err.Error()})
		}
	}

	// ③ Apply dry-run（权限准入预览 + 现存版本探测）
	permOk := true
	permMsg := "通过"
	var baseRV int64
	if e.Applier != nil {
		sub := req.Subject
		res, aerr := e.Applier.Apply(ctx, &repository.Object{
			Kind:            k.Name,
			Scope:           scope,
			Name:            name,
			ResourceVersion: 0,
			Content:         []byte(yamlStr),
		}, apply.Options{
			DryRun:       true,
			FieldManager: "console:propose:chat",
			Subject:      sub,
		})
		if aerr != nil {
			if errors.Is(aerr, apply.ErrDenied) {
				permOk = false
				permMsg = aerr.Error()
			}
		} else if res != nil {
			baseRV = res.RV
		}
	}

	triage := "普通对象 Apply"
	if k.Name == "Proposal" {
		pri := fmt.Sprint(fieldsCopy["priority"])
		if e.isAutoApprove(pri) {
			triage = fmt.Sprintf("低危自动批准（priority=%s 命中 --proposal-auto-approve → status=approved + apply-workflow 派单）", pri)
		} else {
			triage = fmt.Sprintf("高危/人工审批（priority=%s → status=submitted + 自产 ApprovalNeeded 事件进审批屏）", pri)
		}
	}

	allOk := len(fieldErrors) == 0 && permOk
	reasonRef := fmt.Sprintf("intake:%s:v%d", sid, genVer)

	e.mu.Lock()
	if s, ok := e.sessions[sid]; ok {
		s.LastDryRunOk = allOk
		s.FieldErrorCnt = len(fieldErrors)
		s.ReasonRef = reasonRef
	}
	e.mu.Unlock()

	if rpubViolations == nil {
		rpubViolations = []map[string]any{}
	}
	if specViolations == nil {
		specViolations = []string{}
	}
	if fieldErrors == nil {
		fieldErrors = []FieldError{}
	}

	return &GenerateResponse{
		Ok:                  allOk,
		SessionID:           sid,
		GenVersion:          genVer,
		Kind:                k.Name,
		Scope:               scope,
		Name:                name,
		YAML:                yamlStr,
		BaseResourceVersion: baseRV,
		Checks: map[string]any{
			"rpubLint": map[string]any{
				"ok":         len(rpubViolations) == 0,
				"violations": rpubViolations,
			},
			"specAdmission": map[string]any{
				"ok":         len(specViolations) == 0,
				"violations": specViolations,
			},
			"permission": map[string]any{
				"ok":      permOk,
				"subject": proposer,
				"message": permMsg,
				"triage":  triage,
			},
		},
		FieldErrors:   fieldErrors,
		AutoCompleted: autoCompleted,
		TriagePreview: triage,
		ReasonRef:     reasonRef,
		Generator:     "scheme-llm-engine",
	}, nil
}

func (e *Engine) isAutoApprove(priority string) bool {
	for _, p := range e.AutoApprove {
		if p == priority {
			return true
		}
	}
	return false
}

// ───────────────────────── 辅助函数：控件扁平化、自然语言抽取、校验与 YAML 生成 ─────────────────────────

func flattenWidgets(kindName string, ws []scheme.FieldDescriptor) []scheme.FieldDescriptor {
	var out []scheme.FieldDescriptor
	for _, w := range ws {
		if w.Widget == "object" && len(w.Fields) > 0 {
			if kindName == "Proposal" && w.Path == "landing" {
				// Proposal landing 以 landing.target 作为主控件（selector doc|task），ref 为可选附属
				for _, sf := range w.Fields {
					if sf.Name == "target" {
						out = append(out, sf)
					}
				}
				continue
			}
			out = append(out, flattenWidgets(kindName, w.Fields)...)
			continue
		}
		out = append(out, w)
	}
	return out
}

func findWidget(ws []scheme.FieldDescriptor, path string) *scheme.FieldDescriptor {
	for i := range ws {
		if ws[i].Path == path || ws[i].Name == path {
			return &ws[i]
		}
	}
	return nil
}

func normalizeFieldValue(w *scheme.FieldDescriptor, val any) any {
	if val == nil || w == nil {
		return val
	}
	switch w.Widget {
	case "list":
		switch v := val.(type) {
		case []string:
			return v
		case []any:
			var out []string
			for _, item := range v {
				s := strings.TrimSpace(fmt.Sprint(item))
				if s != "" {
					out = append(out, s)
				}
			}
			return out
		case string:
			return parseListText(v)
		}
	case "switch":
		switch v := val.(type) {
		case bool:
			return v
		case string:
			return strings.EqualFold(strings.TrimSpace(v), "true") || strings.TrimSpace(v) == "是" || strings.TrimSpace(v) == "1"
		}
	}
	return val
}

func isEmptyValue(v any) bool {
	if v == nil {
		return true
	}
	switch val := v.(type) {
	case string:
		return strings.TrimSpace(val) == ""
	case []string:
		return len(val) == 0
	case []any:
		return len(val) == 0
	case map[string]any:
		return len(val) == 0
	}
	return false
}

var (
	secretKeyValRe = regexp.MustCompile(`(?i)\b(api[_-]?key|secret[_-]?key|access[_-]?token|password|passwd|private[_-]?key)\s*[:=]\s*([^\s,;]+)`)
	skTokenRe      = regexp.MustCompile(`\bsk-[A-Za-z0-9_-]{10,}\b`)
	validRefPrefs  = []string{"secret:", "env:", "vault:", "keychain:"}
)

func hasValidSecretPrefix(v string) bool {
	v = strings.TrimSpace(v)
	for _, p := range validRefPrefs {
		if strings.HasPrefix(v, p) {
			return true
		}
	}
	return false
}

// checkSecretRefDiscipline 敏感项引用式检查（proposal.md §4.5 第 4 条 / provider.md 规则 2）：
// 禁止手填明文凭据，secret 只允许引用式（secret:/env:/vault:/keychain:）。
func checkSecretRefDiscipline(fieldPath, text string) string {
	lowerPath := strings.ToLower(fieldPath)
	if strings.Contains(lowerPath, "secret") {
		if !hasValidSecretPrefix(text) {
			return fmt.Sprintf("敏感字段 %s 禁止手填明文，必须使用引用式前缀（secret:/env:/vault:/keychain:）", fieldPath)
		}
	}
	if skTokenRe.MatchString(text) {
		return fmt.Sprintf("字段 %s 含疑似明文密钥（sk-...），敏感项禁止手填，只允许 secret:/env:/vault:/keychain: 引用式", fieldPath)
	}
	if m := secretKeyValRe.FindStringSubmatch(text); len(m) >= 3 {
		rhs := strings.Trim(m[2], `"'`)
		if !hasValidSecretPrefix(rhs) && !strings.HasPrefix(rhs, "<") {
			return fmt.Sprintf("字段 %s 含疑似明文凭据（%s=...），敏感项禁止手填，只允许 secret:/env:/vault:/keychain: 引用式", fieldPath, m[1])
		}
	}
	return ""
}

func checkRefExists(ctx context.Context, objects ObjectReader, refKind, scope, refVal string) bool {
	if objects == nil || refKind == "" {
		return true
	}
	targetName := refVal
	if idx := strings.LastIndex(refVal, "/"); idx >= 0 {
		targetName = refVal[idx+1:]
	}
	objs, err := objects.List(ctx, refKind, "")
	if err != nil || len(objs) == 0 {
		return true // 库为空或不可读时不误拦
	}
	for _, o := range objs {
		if o.Name == targetName || (o.Scope+"/"+o.Name) == refVal {
			return true
		}
	}
	return false
}

// extractFieldsFromText 从用户一段自由描述或单轮回答中抽取字段（支持一段话多字段抽取 + 当前询问字段回填）。
func extractFieldsFromText(kindName string, flatW []scheme.FieldDescriptor, msg string, askingField string, current map[string]any) map[string]any {
	out := make(map[string]any)
	text := strings.TrimSpace(msg)
	if text == "" {
		return out
	}

	if kindName == "Proposal" {
		// 1. 显式标签或自然语言分句抽取（标题/目标/验收标准/影响面/优先级/产物落点）
		labeled := extractProposalLabeled(text)
		for k, v := range labeled {
			out[k] = v
		}
		// 2. 若未用显式「标题：」标签，尝试自然语言多句启发式抽取
		if len(out) == 0 && (strings.Contains(text, "目标") || strings.Contains(text, "验收") || strings.Contains(text, "影响面") || strings.Contains(text, "；") || strings.Count(text, "，") >= 2) {
			heur := extractProposalHeuristic(text)
			for k, v := range heur {
				out[k] = v
			}
		}
	} else {
		// 通用 Kind：支持 "字段名: 值" 或 "标签: 值" 多行/分号抽取
		for _, seg := range splitClauses(text) {
			idx := strings.IndexAny(seg, ":：")
			if idx <= 0 {
				continue
			}
			key := strings.TrimSpace(seg[:idx])
			val := strings.TrimSpace(seg[idx+len(string([]rune(seg[idx:])[0])):])
			if val == "" {
				continue
			}
			for _, w := range flatW {
				if strings.EqualFold(key, w.Name) || strings.EqualFold(key, w.Path) || key == w.Label {
					out[w.Path] = normalizeFieldValue(&w, val)
				}
			}
		}
	}

	// 3. 若未命中多字段标签（或当前正针对 askingField 提问而该字段尚未被抽取），将输入赋给当前询问字段
	targetAsk := askingField
	if targetAsk == "" {
		// 首轮直接输入一段简短描述时，赋给首个待补字段
		for _, w := range flatW {
			if w.AutoFill == "" && isEmptyValue(current[w.Path]) {
				targetAsk = w.Path
				break
			}
		}
	}
	if len(out) == 0 && targetAsk != "" {
		w := findWidget(flatW, targetAsk)
		if w != nil {
			out[w.Path] = normalizeFieldValue(w, text)
		}
	}
	return out
}

type labelPattern struct {
	field string
	re    *regexp.Regexp
}

var proposalLabelPatterns = []labelPattern{
	{field: "title", re: regexp.MustCompile(`(?i)(?:^|[；;\n。])\s*(?:标题|提案标题|名称|title)\s*[:：=]\s*`)},
	{field: "summary", re: regexp.MustCompile(`(?i)(?:^|[；;\n。])\s*(?:目标|摘要|背景与目标|做什么|summary)\s*[:：=]\s*`)},
	{field: "acceptance", re: regexp.MustCompile(`(?i)(?:^|[；;\n。])\s*(?:验收标准|验收条目|验收|acceptance)\s*[:：=]\s*`)},
	{field: "impact", re: regexp.MustCompile(`(?i)(?:^|[；;\n。])\s*(?:影响面|影响范围|涉及系统|impact)\s*[:：=]\s*`)},
	{field: "priority", re: regexp.MustCompile(`(?i)(?:^|[；;\n。])\s*(?:优先级|priority)\s*[:：=]\s*`)},
	{field: "landing.target", re: regexp.MustCompile(`(?i)(?:^|[；;\n。])\s*(?:产物落点|落点|产物|landing)\s*[:：=]\s*`)},
}

func extractProposalLabeled(text string) map[string]any {
	type hit struct {
		field    string
		start    int
		valStart int
	}
	var hits []hit
	for _, lp := range proposalLabelPatterns {
		loc := lp.re.FindStringIndex(text)
		if loc != nil {
			hits = append(hits, hit{field: lp.field, start: loc[0], valStart: loc[1]})
		}
	}
	if len(hits) == 0 {
		return nil
	}
	sort.Slice(hits, func(i, j int) bool { return hits[i].start < hits[j].start })
	out := make(map[string]any)
	// 若首个标签前还有前置文本且未含 title 标签，将其作为 title
	if hits[0].start > 0 {
		prefix := strings.Trim(strings.TrimSpace(text[:hits[0].start]), "，,；;。 ")
		if prefix != "" {
			out["title"] = prefix
		}
	}
	for i, h := range hits {
		end := len(text)
		if i+1 < len(hits) {
			end = hits[i+1].start
		}
		rawVal := strings.Trim(strings.TrimSpace(text[h.valStart:end]), "，,；;。 ")
		if rawVal == "" {
			continue
		}
		switch h.field {
		case "acceptance":
			items := parseListText(rawVal)
			if len(items) > 0 {
				out["acceptance"] = items
			}
		case "priority":
			out["priority"] = normalizePriorityWord(rawVal)
		case "landing.target":
			out["landing.target"] = normalizeLandingWord(rawVal)
		default:
			out[h.field] = rawVal
		}
	}
	return out
}

func extractProposalHeuristic(text string) map[string]any {
	out := make(map[string]any)
	// 检查优先级词
	lower := strings.ToLower(text)
	switch {
	case strings.Contains(lower, "critical") || strings.Contains(text, "紧急"):
		out["priority"] = "critical"
	case strings.Contains(lower, "high") || strings.Contains(text, "高优") || strings.Contains(text, "高危"):
		out["priority"] = "high"
	case strings.Contains(lower, "normal") || strings.Contains(text, "中优") || strings.Contains(text, "普通优先级"):
		out["priority"] = "normal"
	case strings.Contains(lower, "low") || strings.Contains(text, "低优"):
		out["priority"] = "low"
	}
	// 检查落点词
	if strings.Contains(text, "任务文件") || strings.Contains(lower, "target=task") || strings.Contains(lower, "landing=task") {
		out["landing.target"] = "task"
	}
	// 按分号/句号/逗号切分寻找目标、验收、影响面与标题
	clauses := splitClauses(text)
	for i, c := range clauses {
		c = strings.TrimSpace(c)
		if c == "" {
			continue
		}
		switch {
		case strings.HasPrefix(c, "目标是") || strings.HasPrefix(c, "为了") || strings.HasPrefix(c, "用于"):
			out["summary"] = strings.TrimSpace(strings.TrimPrefix(c, "目标是"))
		case strings.Contains(c, "验收标准") || strings.Contains(c, "验收是"):
			cleaned := regexp.MustCompile(`^.*?(?:验收标准是?|验收是?)\s*[:：]?\s*`).ReplaceAllString(c, "")
			if items := parseListText(cleaned); len(items) > 0 {
				out["acceptance"] = items
			}
		case strings.HasPrefix(c, "影响面是") || strings.HasPrefix(c, "影响面") || strings.HasPrefix(c, "涉及"):
			cleaned := strings.TrimSpace(strings.TrimPrefix(strings.TrimPrefix(c, "影响面是"), "影响面"))
			cleaned = strings.TrimLeft(cleaned, ":： ")
			if cleaned != "" {
				out["impact"] = cleaned
			}
		default:
			if i == 0 && isEmptyValue(out["title"]) && len([]rune(c)) <= 50 {
				out["title"] = c
			} else if isEmptyValue(out["summary"]) && len([]rune(c)) > 8 {
				out["summary"] = c
			}
		}
	}
	return out
}

func normalizePriorityWord(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "低", "低优", "低危", "low":
		return "low"
	case "中", "中优", "普通", "normal", "medium":
		return "normal"
	case "高", "高优", "高危", "high":
		return "high"
	case "紧急", "最高", "critical":
		return "critical"
	default:
		return s // 保留原值以便步骤 3 冲突检测识别非法枚举！
	}
}

func normalizeLandingWord(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	switch {
	case s == "doc" || strings.Contains(s, "文档") || strings.Contains(s, "docsroot"):
		return "doc"
	case s == "task" || strings.Contains(s, "任务"):
		return "task"
	default:
		return s
	}
}

var listSplitRe = regexp.MustCompile(`(?:\r?\n|[；;]|(?:^|\s+)\d+[.、)）]\s*)`)

func parseListText(raw string) []string {
	parts := listSplitRe.Split(strings.TrimSpace(raw), -1)
	var out []string
	for _, p := range parts {
		p = strings.Trim(strings.TrimSpace(p), "-*•，,。；; ")
		if p != "" {
			out = append(out, p)
		}
	}
	if len(out) == 0 && strings.TrimSpace(raw) != "" {
		out = append(out, strings.TrimSpace(raw))
	}
	return out
}

func splitClauses(s string) []string {
	return regexp.MustCompile(`[\r\n；;。]+`).Split(s, -1)
}

func evaluateFields(ctx context.Context, kindName string, flatW []scheme.FieldDescriptor, fields map[string]any, sources map[string]string, objects ObjectReader) ([]FieldState, []string, []string) {
	states := make([]FieldState, 0, len(flatW))
	missing := []string{}
	conflicts := []string{}

	for _, w := range flatW {
		val := fields[w.Path]
		src := sources[w.Path]
		fs := FieldState{
			Path:     w.Path,
			Name:     w.Name,
			Label:    w.Label,
			Widget:   w.Widget,
			Type:     w.Type,
			Required: w.Required,
			Enum:     w.Enum,
			Default:  w.Default,
			Example:  w.Example,
			Value:    val,
			Source:   src,
		}

		// Proposal impact 虽在 scheme.yaml 为 ?string，但在提案对话录入六要素（proposal.md §4.5 / console-ui.md §12）
		// 中属待询问项（提问时提供默认值，跳过即填默认）。
		needAsk := w.Required || (kindName == "Proposal" && w.Path == "impact")
		if w.AutoFill == "system" {
			needAsk = false
		}

		if isEmptyValue(val) {
			if needAsk {
				fs.Status = "missing"
				missing = append(missing, w.Path)
			} else {
				fs.Status = "default"
			}
			states = append(states, fs)
			continue
		}

		// 校验枚举闭合集（R-UX-2）
		if w.Widget == "selector" && len(w.Enum) > 0 && !w.HasWildcard {
			sval := strings.TrimSpace(fmt.Sprint(val))
			hit := false
			for _, opt := range w.Enum {
				if opt == sval {
					hit = true
					break
				}
			}
			if !hit {
				fs.Status = "conflict"
				fs.Error = fmt.Sprintf("值 %q 不在合法枚举 [%s] 中", sval, strings.Join(w.Enum, "|"))
				conflicts = append(conflicts, w.Path)
				states = append(states, fs)
				continue
			}
		}

		// 校验 R-PUB-1 lint 与敏感项引用式纪律
		txt := valueToText(val)
		if vs := docrepo.Lint([]byte(txt)); len(vs) > 0 {
			fs.Status = "conflict"
			fs.Error = fmt.Sprintf("R-PUB-1 (%s) 违规：含 %q", vs[0].Rule, vs[0].Text)
			conflicts = append(conflicts, w.Path)
			states = append(states, fs)
			continue
		}
		if secMsg := checkSecretRefDiscipline(w.Path, txt); secMsg != "" {
			fs.Status = "conflict"
			fs.Error = secMsg
			conflicts = append(conflicts, w.Path)
			states = append(states, fs)
			continue
		}

		// 校验 ref 引用存在性（R-UX-2）
		if w.Widget == "ref" && w.RefKind != "" && objects != nil {
			sval := strings.TrimSpace(fmt.Sprint(val))
			if sval != "" && !checkRefExists(ctx, objects, w.RefKind, "global", sval) {
				fs.Status = "conflict"
				fs.Error = fmt.Sprintf("引用对象 %q (%s) 不在册", sval, w.RefKind)
				conflicts = append(conflicts, w.Path)
				states = append(states, fs)
				continue
			}
		}

		if src == "default" || src == "subject" || src == "system" {
			fs.Status = "default"
		} else {
			fs.Status = "filled"
		}
		states = append(states, fs)
	}
	return states, missing, conflicts
}

func composeReply(kindName string, flatW []scheme.FieldDescriptor, sess *SessionRecord, autoNotes []string, extracted map[string]any, states []FieldState, missing, conflicts []string, nextAsk string) string {
	var parts []string
	if sess.Turns == 1 && len(autoNotes) > 0 {
		parts = append(parts, fmt.Sprintf("已结合当前上下文自动补充缺省项：%s（右侧字段卡可随时修改，不会重复盘问已推断项）。", strings.Join(autoNotes, "、")))
	}
	if len(extracted) > 0 {
		var keys []string
		for k := range extracted {
			w := findWidget(flatW, k)
			if w != nil {
				keys = append(keys, w.Label)
			} else {
				keys = append(keys, k)
			}
		}
		sort.Strings(keys)
		parts = append(parts, fmt.Sprintf("已抽取并同步至字段卡：%s。", strings.Join(keys, "、")))
	}

	if len(conflicts) > 0 {
		w := findWidget(flatW, nextAsk)
		errText := ""
		for _, st := range states {
			if st.Path == nextAsk {
				errText = st.Error
				break
			}
		}
		label := nextAsk
		if w != nil {
			label = w.Label
		}
		hint := ""
		if w != nil && len(w.Enum) > 0 {
			hint = fmt.Sprintf("合法枚举值：[%s]（默认：%v，示例：%s）", strings.Join(w.Enum, " | "), defaultOrFirst(w), w.Example)
		} else if w != nil {
			hint = fmt.Sprintf("（示例：%s）", w.Example)
		}
		parts = append(parts, fmt.Sprintf("字段【%s (%s)】存在冲突或准入违规：%s。请修正该项 %s：", label, nextAsk, errText, hint))
		return strings.Join(parts, "\n")
	}

	if len(missing) == 0 {
		parts = append(parts, "全部 scheme 字段已齐备（字段卡全绿）！请点击【生成提案预览】，由提交节点生成标准 YAML 对象并执行 R-PUB lint 与 spec 准入 dry-run。")
		return strings.Join(parts, "\n")
	}

	w := findWidget(flatW, nextAsk)
	if w == nil {
		parts = append(parts, fmt.Sprintf("请补充字段 %s：", nextAsk))
		return strings.Join(parts, "\n")
	}
	defStr := "无（必填）"
	if w.Default != nil && fmt.Sprint(w.Default) != "" {
		defStr = fmt.Sprintf("%v（可点「跳过」采用默认）", w.Default)
	} else if w.Example != "" {
		defStr = fmt.Sprintf("可点「跳过」采用示例值「%s」", w.Example)
	}
	parts = append(parts, fmt.Sprintf("请提供【%s (%s)】（剩余待补 %d 项；默认：%s；示例：%s；也可一段话自由描述多项）：",
		w.Label, w.Path, len(missing), defStr, w.Example))
	return strings.Join(parts, "\n")
}

func defaultOrFirst(w *scheme.FieldDescriptor) any {
	if w.Default != nil {
		return w.Default
	}
	if len(w.Enum) > 0 {
		return w.Enum[0]
	}
	return ""
}

func valueToText(v any) string {
	if v == nil {
		return ""
	}
	switch val := v.(type) {
	case string:
		return val
	case []string:
		return strings.Join(val, "\n")
	case []any:
		var s []string
		for _, it := range val {
			s = append(s, fmt.Sprint(it))
		}
		return strings.Join(s, "\n")
	default:
		b, _ := yaml.Marshal(val)
		return string(b)
	}
}

func deriveSlug(title, sessionID string) string {
	var b strings.Builder
	for _, r := range strings.ToLower(title) {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') {
			b.WriteRune(r)
		} else if b.Len() > 0 && !strings.HasSuffix(b.String(), "-") {
			b.WriteByte('-')
		}
	}
	s := strings.Trim(b.String(), "-")
	if len(s) >= 3 {
		if len(s) > 36 {
			s = strings.Trim(s[:36], "-")
		}
		return "prop-" + s
	}
	suffix := sessionID
	if idx := strings.LastIndex(sessionID, "-"); idx >= 0 {
		suffix = sessionID[idx+1:]
	}
	return fmt.Sprintf("prop-%d-%s", time.Now().UTC().Unix()%100000, suffix)
}

func buildObjectYAML(k scheme.Kind, scope, name string, fields map[string]any) string {
	if scope == "" {
		scope = "global"
	}
	if name == "" {
		name = "unnamed"
	}
	if k.Name == "Proposal" {
		title, _ := fields["title"].(string)
		summary, _ := fields["summary"].(string)
		impact, _ := fields["impact"].(string)
		priority, _ := fields["priority"].(string)
		if priority == "" {
			priority = "low"
		}
		proposer, _ := fields["proposer"].(string)
		if proposer == "" {
			proposer = "console-operator"
		}
		status, _ := fields["status"].(string)
		if status == "" {
			status = "submitted"
		}
		target, _ := fields["landing.target"].(string)
		if target == "" {
			target = "doc"
		}
		landingRef, _ := fields["landing.ref"].(string)

		var acc []string
		switch v := fields["acceptance"].(type) {
		case []string:
			acc = v
		case []any:
			for _, it := range v {
				if s := strings.TrimSpace(fmt.Sprint(it)); s != "" {
					acc = append(acc, s)
				}
			}
		case string:
			acc = parseListText(v)
		}

		var sb strings.Builder
		sb.WriteString("apiVersion: core/v1\n")
		sb.WriteString("kind: Proposal\n")
		sb.WriteString(fmt.Sprintf("metadata: {name: %s, scope: %s}\n", yamlScalar(name), yamlScalar(scope)))
		sb.WriteString("spec:\n")
		if strings.TrimSpace(title) != "" {
			sb.WriteString(fmt.Sprintf("  title: %s\n", yamlScalar(title)))
		}
		if strings.TrimSpace(summary) != "" {
			sb.WriteString(fmt.Sprintf("  summary: %s\n", yamlScalar(summary)))
		}
		if len(acc) > 0 {
			sb.WriteString("  acceptance:\n")
			for _, item := range acc {
				sb.WriteString(fmt.Sprintf("    - %s\n", yamlScalar(item)))
			}
		}
		if strings.TrimSpace(impact) != "" {
			sb.WriteString(fmt.Sprintf("  impact: %s\n", yamlScalar(impact)))
		}
		sb.WriteString(fmt.Sprintf("  priority: %s\n", yamlScalar(priority)))
		sb.WriteString(fmt.Sprintf("  proposer: %s\n", yamlScalar(proposer)))
		sb.WriteString(fmt.Sprintf("  status: %s\n", yamlScalar(status)))
		if strings.TrimSpace(landingRef) != "" {
			sb.WriteString(fmt.Sprintf("  landing: {target: %s, ref: %s}\n", yamlScalar(target), yamlScalar(landingRef)))
		} else {
			sb.WriteString(fmt.Sprintf("  landing: {target: %s}\n", yamlScalar(target)))
		}
		// 动态附加字段（当用户修改 Scheme 为 Proposal 新增字段时，自动序列化入 YAML！）
		known := map[string]bool{
			"title": true, "summary": true, "acceptance": true, "impact": true,
			"priority": true, "proposer": true, "status": true,
			"landing": true, "landing.target": true, "landing.ref": true,
		}
		var extras []string
		for fk := range fields {
			if !known[fk] && !isEmptyValue(fields[fk]) {
				extras = append(extras, fk)
			}
		}
		sort.Strings(extras)
		for _, fk := range extras {
			sb.WriteString(fmt.Sprintf("  %s: %s\n", fk, yamlScalar(fmt.Sprint(fields[fk]))))
		}
		return sb.String()
	}

	// 通用 Kind YAML 生成
	specMap := make(map[string]any)
	for fk, fv := range fields {
		if !isEmptyValue(fv) {
			specMap[fk] = fv
		}
	}
	doc := map[string]any{
		"apiVersion": "core/v1",
		"kind":       k.Name,
		"metadata": map[string]any{
			"name":  name,
			"scope": scope,
		},
		"spec": specMap,
	}
	out, err := yaml.Marshal(doc)
	if err != nil {
		return fmt.Sprintf("apiVersion: core/v1\nkind: %s\nmetadata: {name: %s, scope: %s}\nspec: {}\n", k.Name, name, scope)
	}
	return string(out)
}

func yamlScalar(s string) string {
	s = strings.TrimSpace(s)
	if s == "" {
		return `""`
	}
	if strings.ContainsAny(s, ":#-?&*!>|%@`'\"{[},\n\r\t ") {
		escaped := strings.ReplaceAll(s, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		escaped = strings.ReplaceAll(escaped, "\n", `\n`)
		return `"` + escaped + `"`
	}
	return s
}

func appendFieldErrorUnique(list []FieldError, fe FieldError) []FieldError {
	for _, existing := range list {
		if existing.Field == fe.Field && existing.Rule == fe.Rule {
			return list
		}
	}
	return append(list, fe)
}

func cloneMap(m map[string]any) map[string]any {
	if m == nil {
		return nil
	}
	out := make(map[string]any, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

func cloneStrMap(m map[string]string) map[string]string {
	if m == nil {
		return nil
	}
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}
