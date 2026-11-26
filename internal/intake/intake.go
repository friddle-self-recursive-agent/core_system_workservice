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
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/apply"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/docrepo"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/runtimestore"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
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
	// store 运行态投影存储挂接（G-01——nil=纯内存既有行为面零翻转；挂接后会话每次变更
	// 同步 upsert rt_intake_sessions，挂接时启动恢复装载历史行——persist.go）。
	store runtimestore.Store
}

// NewEngine 构造对话录入引擎。包级装配锚已设定（SetDefaultStore）时自动挂接并启动恢复
// （G-01——face 侧构造零迁移的注入面；显式通道=AttachStore）。
func NewEngine(sch *scheme.Scheme, objects ObjectReader, applier DryRunApplier, autoApprove []string) *Engine {
	storeMu.Lock()
	pick := defaultStore
	storeMu.Unlock()
	e := &Engine{
		Sch:         sch,
		Objects:     objects,
		Applier:     applier,
		AutoApprove: append([]string(nil), autoApprove...),
		sessions:    make(map[string]*SessionRecord),
		store:       pick,
	}
	if pick != nil {
		e.restoreLocked(pick)
	}
	return e
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

	// 写穿透（G-01）：本轮会话投影同步 upsert（锁内同步落库与投影同序——persist.go）。
	e.persistSessionLocked(sess)

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
	// 写穿透（G-01）：受理/字段合并词形同步 upsert（锁内同步落库与投影同序——persist.go）。
	e.persistSessionLocked(sess)
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
		e.persistSessionLocked(s) // 写穿透（G-01）：dry-run 判定词形同步 upsert
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
