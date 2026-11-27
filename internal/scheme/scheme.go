// Package scheme 载入并校验类型注册表（core/scheme.yaml）。
// 设计依据：core_design/concepts/scheme.md；对象模型权威 = core_design/core/scheme.yaml。
// 启动自检：结构必填项、kind 唯一性、storage 模板必填——不合规即拒绝启动（野对象在
// Repository 层由 Lookup 闸门拒绝，见 concepts/core.md）。
package scheme

import (
	"errors"
	"fmt"
	"os"
	"sync"

	"gopkg.in/yaml.v3"
)

// Kind 注册表中的一个对象类型。
// Storage 是存储路径模板（布局见 core_system_instance/layout.md）；
// Spec/Fields/Status/Relations 保留声明原文（map 形态），供查询面透出；
// Widgets 由 Spec/Fields 实时机读派生（R-UX-2 scheme→widget 单一真相）；
// 未识别的顶层字段收敛进 Extra，避免声明演进时载入端先行失配。
type Kind struct {
	Name       string            `yaml:"kind" json:"name"`
	Builtin    bool              `yaml:"builtin" json:"builtin,omitempty"`
	Storage    string            `yaml:"storage" json:"storage"`
	Spec       map[string]any    `yaml:"spec,omitempty" json:"spec,omitempty"`
	Fields     map[string]any    `yaml:"fields,omitempty" json:"fields,omitempty"`
	Status     map[string]any    `yaml:"status,omitempty" json:"status,omitempty"`
	Relations  map[string]any    `yaml:"relations,omitempty" json:"relations,omitempty"`
	BuiltinSet []string          `yaml:"builtinSet,omitempty" json:"builtinSet,omitempty"`
	Rules      []string          `yaml:"rules,omitempty" json:"rules,omitempty"`
	Widgets    []FieldDescriptor `yaml:"-" json:"widgets,omitempty"`
	Extra      map[string]any    `yaml:",inline" json:"extra,omitempty"`
}

// DocType 文档类型注册项（本系统的 CRD 机制，concepts/doctype.md）。
// Category=design|runtime（缺省 design）；ExportDefault=默认是否纳入导出（缺省 true；
// 运行期文档如 BugDoc/DriftReport/PlanDoc/SessionLog 在 scheme.yaml 中声明 exportDefault: false，
// 作为数据库导出配置 docs_export_config 的种子初值，绝不在导出器代码中硬编码过滤）。
type DocType struct {
	Name          string   `yaml:"name" json:"name"`
	Mandatory     string   `yaml:"mandatory" json:"mandatory"`
	Category      string   `yaml:"category,omitempty" json:"category,omitempty"`
	ExportDefault *bool    `yaml:"exportDefault,omitempty" json:"exportDefault,omitempty"`
	Sections      []string `yaml:"sections,omitempty" json:"sections,omitempty"`
	Writers       []string `yaml:"writers,omitempty" json:"writers,omitempty"`
	AppliesTo     string   `yaml:"appliesTo" json:"appliesTo"`
	Admission     string   `yaml:"admission,omitempty" json:"admission,omitempty"`
}

// IsRuntimeDoc 是否为运行期文档（scheme 声明 category=runtime 或 exportDefault=false）。
func (d DocType) IsRuntimeDoc() bool {
	if d.Category == "runtime" {
		return true
	}
	if d.ExportDefault != nil && !*d.ExportDefault {
		return true
	}
	return false
}

// EventType 事件类型注册项（v0.5x 事件分诊 + BUG-010 分层/展示词形：tier/visibility
// 行可选字段——concepts/event.md 规则 18；core/scheme.yaml eventTypes 行同构）。
type EventType struct {
	Type         string `yaml:"type" json:"type"`
	Source       string `yaml:"source" json:"source"`
	DefaultLevel string `yaml:"defaultLevel,omitempty" json:"defaultLevel,omitempty"`
	Dedup        string `yaml:"dedup,omitempty" json:"dedup,omitempty"`
	Route        string `yaml:"route" json:"route"`
	Tier         string `yaml:"tier,omitempty" json:"tier,omitempty"`             // infra|domain|attention（缺省 domain）
	Visibility   string `yaml:"visibility,omitempty" json:"visibility,omitempty"` // default-visible|default-hidden（tier=infra 缺省 default-hidden）
}

// 事件分层与展示词汇（event.md 规则 18：凡 tier=infra 缺省 default-hidden——默认在
// 事件流/关注报告/第一段侦察中隐藏，仅调试显式开启 showInfra 展开；attention 异常族
// （HeartbeatLost/SessionStalled/Escalate）恒默认上屏）。
const (
	TierInfra     = "infra"
	TierDomain    = "domain"
	TierAttention = "attention"

	VisibilityDefaultVisible = "default-visible"
	VisibilityDefaultHidden  = "default-hidden"
)

// EffectiveTier 分层生效词（注册行缺省词形=domain，与 scheme.yaml 行注释同源）。
func (e EventType) EffectiveTier() string {
	if e.Tier == "" {
		return TierDomain
	}
	return e.Tier
}

// EffectiveVisibility 展示生效词：显式声明优先；缺席按 tier 投影（tier=infra→
// default-hidden，其余→default-visible——event.md 规则 18② 的缺省词形）。
func (e EventType) EffectiveVisibility() string {
	if e.Visibility != "" {
		return e.Visibility
	}
	if e.EffectiveTier() == TierInfra {
		return VisibilityDefaultHidden
	}
	return VisibilityDefaultVisible
}

// DispatchRule 事件路由规则（穷举缺省 + 死信兜底，concepts/event.md）。
type DispatchRule struct {
	On       string `yaml:"on" json:"on"`
	For      string `yaml:"for" json:"for"`
	Dispatch string `yaml:"dispatch" json:"dispatch"`
	CC       string `yaml:"cc,omitempty" json:"cc,omitempty"`
}

// Metadata scheme 自身的元数据。
type Metadata struct {
	Name  string `yaml:"name" json:"name"`
	Scope string `yaml:"scope,omitempty" json:"scope,omitempty"`
}

// SpecBody spec 区（声明文档里嵌在 spec: 之下，字段与文档同构）。
type SpecBody struct {
	Version     string         `yaml:"version" json:"version"`
	Kinds       []Kind         `yaml:"kinds" json:"kinds,omitempty"`
	DocTypes    []DocType      `yaml:"docTypes" json:"docTypes,omitempty"`
	EventTypes  []EventType    `yaml:"eventTypes" json:"eventTypes,omitempty"`
	Dispatch    []DispatchRule `yaml:"dispatch" json:"dispatch,omitempty"`
	Runtime     map[string]any `yaml:"runtime,omitempty" json:"runtime,omitempty"`
	Conventions map[string]any `yaml:"conventions,omitempty" json:"conventions,omitempty"`
}

// Scheme 类型注册表（core/scheme.yaml 的内存态；支持 Replace 原子热刷新）。
type Scheme struct {
	mu         sync.RWMutex `yaml:"-" json:"-"`
	APIVersion string       `yaml:"apiVersion" json:"apiVersion"`
	KindName   string       `yaml:"kind" json:"kind"`
	Metadata   Metadata     `yaml:"metadata" json:"metadata"`
	Spec       SpecBody     `yaml:"spec" json:"spec"`
	// Source 载入来源路径（仅内存标注，不序列化）。
	Source string `yaml:"-" json:"-"`
}

// Load 从字节流载入并校验 scheme 声明。
func Load(data []byte) (*Scheme, error) {
	var s Scheme
	if err := yaml.Unmarshal(data, &s); err != nil {
		return nil, fmt.Errorf("scheme: 解析失败: %w", err)
	}
	if err := s.validate(); err != nil {
		return nil, err
	}
	return &s, nil
}

// LoadFile 从文件载入（cored 启动自检入口）。
func LoadFile(path string) (*Scheme, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("scheme: 读取 %q: %w", path, err)
	}
	s, err := Load(data)
	if err != nil {
		return nil, err
	}
	s.Source = path
	return s, nil
}

// validate 启动自检：必填项 + 唯一性。失败即拒绝启动（concepts/scheme.md 设计规则）。
func (s *Scheme) validate() error {
	if s.APIVersion == "" {
		return errors.New("scheme: apiVersion 必填")
	}
	if s.KindName != "" && s.KindName != "Scheme" {
		return fmt.Errorf("scheme: kind 必须为 Scheme，得到 %q", s.KindName)
	}
	if s.Metadata.Name == "" {
		return errors.New("scheme: metadata.name 必填")
	}
	if s.Spec.Version == "" {
		return errors.New("scheme: spec.version 必填")
	}
	if len(s.Spec.Kinds) == 0 {
		return errors.New("scheme: spec.kinds 不能为空（至少自注册 Scheme）")
	}
	seenKind := map[string]bool{}
	for i := range s.Spec.Kinds {
		k := &s.Spec.Kinds[i]
		if k.Name == "" {
			return fmt.Errorf("scheme: kinds[%d] 缺 kind 名", i)
		}
		if seenKind[k.Name] {
			return fmt.Errorf("scheme: kind %q 重复注册", k.Name)
		}
		seenKind[k.Name] = true
		if k.Storage == "" {
			return fmt.Errorf("scheme: kind %q 缺 storage 模板", k.Name)
		}
		k.Widgets = DescribeKindWidgets(*k)
	}
	seenDoc := map[string]bool{}
	for _, d := range s.Spec.DocTypes {
		if d.Name == "" {
			return errors.New("scheme: docTypes 存在缺 name 的条目")
		}
		if seenDoc[d.Name] {
			return fmt.Errorf("scheme: docType %q 重复注册", d.Name)
		}
		seenDoc[d.Name] = true
	}
	seenEvent := map[string]bool{}
	for _, e := range s.Spec.EventTypes {
		if e.Type == "" {
			return errors.New("scheme: eventTypes 存在缺 type 的条目")
		}
		if seenEvent[e.Type] {
			return fmt.Errorf("scheme: eventType %q 重复注册", e.Type)
		}
		seenEvent[e.Type] = true
		// tier/visibility 行可选词形校验（BUG-010，event.md 规则 18——出现即校验枚举，
		// 未知词拒绝启动；缺席=缺省词形投影见 EffectiveTier/EffectiveVisibility）。
		switch e.Tier {
		case "", TierInfra, TierDomain, TierAttention:
		default:
			return fmt.Errorf("scheme: eventType %q tier %q 不在词汇表（infra|domain|attention）", e.Type, e.Tier)
		}
		switch e.Visibility {
		case "", VisibilityDefaultVisible, VisibilityDefaultHidden:
		default:
			return fmt.Errorf("scheme: eventType %q visibility %q 不在词汇表（default-visible|default-hidden）", e.Type, e.Visibility)
		}
	}
	for i, r := range s.Spec.Dispatch {
		if r.On == "" || r.Dispatch == "" {
			return fmt.Errorf("scheme: dispatch[%d] 缺 on 或 dispatch", i)
		}
	}
	return nil
}

// Replace 原子热刷新内存中的 Scheme（当 Scheme/builtin/default 或 core/scheme.yaml 经准入落库后调用，
// R-UX-2「改 scheme 后 UI 自动反映」）。
func (s *Scheme) Replace(next *Scheme) {
	if s == nil || next == nil {
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	s.APIVersion = next.APIVersion
	s.KindName = next.KindName
	s.Metadata = next.Metadata
	s.Spec = next.Spec
	for i := range s.Spec.Kinds {
		s.Spec.Kinds[i].Widgets = DescribeKindWidgets(s.Spec.Kinds[i])
	}
}

// Lookup 查找对象类型；未注册返回错误（Scheme 闸门：野对象拒绝）。
func (s *Scheme) Lookup(kind string) (Kind, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for _, k := range s.Spec.Kinds {
		if k.Name == kind {
			if len(k.Widgets) == 0 {
				k.Widgets = DescribeKindWidgets(k)
			}
			return k, nil
		}
	}
	return Kind{}, fmt.Errorf("scheme: kind %q not registered", kind)
}

// LookupEventType 事件类型注册查询（事件发布面 v0.20 入口：事件只校验类型注册不校验
// 内容，concepts/event.md §2 边界+规则 8——注入面以本表为类型合法性唯一权威）。
// 未注册返回错误；命中返回注册项指针（DefaultLevel 等由发布面按表投影）。
func (s *Scheme) LookupEventType(typ string) (*EventType, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	for i := range s.Spec.EventTypes {
		if s.Spec.EventTypes[i].Type == typ {
			ev := s.Spec.EventTypes[i]
			return &ev, nil
		}
	}
	return nil, fmt.Errorf("scheme: 事件类型 %q 未注册", typ)
}

// HiddenEventTypes 默认隐藏事件类型集（BUG-010 visibility 过滤投影——event.md 规则 18）：
// 注册表内生效词形为 default-hidden 的类型集合（tier=infra 缺省词形随行）。读面过滤用：
// 事件行只存 type 字符串，type→tier/visibility 映射以本表为唯一权威；**未注册类型不在
// 集内=按 domain/default-visible 处理（不误伤）**。scheme 热刷新（Replace）后下次调用
// 反映新表——调用方逐请求现取，零缓存。
func (s *Scheme) HiddenEventTypes() map[string]bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := map[string]bool{}
	for _, e := range s.Spec.EventTypes {
		if e.EffectiveVisibility() == VisibilityDefaultHidden {
			out[e.Type] = true
		}
	}
	return out
}

// ListKinds 返回全部已注册对象类型（只读清单，含 R-UX-2 widgets 投影）。
func (s *Scheme) ListKinds() []Kind {
	s.mu.RLock()
	defer s.mu.RUnlock()
	out := make([]Kind, len(s.Spec.Kinds))
	copy(out, s.Spec.Kinds)
	for i := range out {
		if len(out[i].Widgets) == 0 {
			out[i].Widgets = DescribeKindWidgets(out[i])
		}
	}
	return out
}

// Summary scheme 概览（查询面 /v1/scheme 的载荷）。
type Summary struct {
	APIVersion     string `json:"apiVersion"`
	Name           string `json:"name"`
	Scope          string `json:"scope,omitempty"`
	Version        string `json:"version"`
	KindCount      int    `json:"kindCount"`
	DocTypeCount   int    `json:"docTypeCount"`
	EventTypeCount int    `json:"eventTypeCount"`
	DispatchCount  int    `json:"dispatchCount"`
}

// Summary 计算 scheme 概览。
func (s *Scheme) Summary() Summary {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return Summary{
		APIVersion:     s.APIVersion,
		Name:           s.Metadata.Name,
		Scope:          s.Metadata.Scope,
		Version:        s.Spec.Version,
		KindCount:      len(s.Spec.Kinds),
		DocTypeCount:   len(s.Spec.DocTypes),
		EventTypeCount: len(s.Spec.EventTypes),
		DispatchCount:  len(s.Spec.Dispatch),
	}
}

// BuiltinWorkflows 从 spec.runtime.workflows.builtin 读取 scheme 声明的内置 workflow 清单。
// 未声明时返回 nil（调用方可回退代码默认注册表，保持测试 fixture 兼容）。
func (s *Scheme) BuiltinWorkflows() []string {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	wfMap, ok := s.Spec.Runtime["workflows"].(map[string]any)
	if !ok {
		return nil
	}
	raw, ok := wfMap["builtin"].([]any)
	if !ok || len(raw) == 0 {
		return nil
	}
	out := make([]string, 0, len(raw))
	for _, it := range raw {
		if str, ok := it.(string); ok && str != "" {
			out = append(out, str)
		}
	}
	return out
}

// WorkflowAliases 从 spec.runtime.workflows.aliases 读取历史 workflow 别名到标准名的映射。
func (s *Scheme) WorkflowAliases() map[string]string {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	wfMap, ok := s.Spec.Runtime["workflows"].(map[string]any)
	if !ok {
		return nil
	}
	raw, ok := wfMap["aliases"].(map[string]any)
	if !ok || len(raw) == 0 {
		return nil
	}
	out := make(map[string]string, len(raw))
	for k, v := range raw {
		if str, ok := v.(string); ok && k != "" && str != "" {
			out[k] = str
		}
	}
	return out
}

// WorkflowAdapters 从 spec.runtime.workflows.adapters 读取 Scheme 层节点 Adapter 契约声明。
func (s *Scheme) WorkflowAdapters() map[string]any {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	wfMap, ok := s.Spec.Runtime["workflows"].(map[string]any)
	if !ok {
		return nil
	}
	raw, _ := wfMap["adapters"].(map[string]any)
	return raw
}

// LLMRuntime 从 spec.runtime.llm 读取默认 Provider 与模型分级路由声明。
func (s *Scheme) LLMRuntime() map[string]any {
	if s == nil {
		return nil
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	raw, _ := s.Spec.Runtime["llm"].(map[string]any)
	return raw
}

// IsUndeletableWorkflow 判定给定 workflow 名称是否受内置不可删除保护（core-system-developer-loop / default-task）。
func (s *Scheme) IsUndeletableWorkflow(name string) bool {
	if name == "core-system-developer-loop" {
		return true
	}
	for _, b := range s.BuiltinWorkflows() {
		if b == name {
			return true
		}
	}
	return false
}

