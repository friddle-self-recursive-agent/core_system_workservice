// widgets.go —— R-UX-2 scheme 驱动编辑控件描述符派生（core_design/rule.md R-UX-2 & console-ui.md §12）。
// 单一真相：一切录入/编辑面（对话字段卡、降级表单、对象编辑器）的字段集、类型、必填、枚举、默认值、
// 控件类型（selector/input/textarea/list/ref/switch/map/number/object）全部从 Kind.Spec/Fields 实时机读派生，
// UI 零硬编码；当 Scheme 经 Apply 修改后，DescribeKindWidgets 即时输出更新后的控件树。
package scheme

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// FieldDescriptor 单个字段的机读控件描述符（R-UX-2 scheme → widget 映射 + R-UX-1 询问元数据）。
type FieldDescriptor struct {
	Name        string            `json:"name"`
	Path        string            `json:"path"`
	Label       string            `json:"label"`
	RawDecl     string            `json:"rawDecl,omitempty"`
	Type        string            `json:"type"`                  // enum|string|text|list|ref|bool|map|number|object|duration|timestamp
	Widget      string            `json:"widget"`                // selector|input|textarea|list|ref|switch|map|number|object
	Required    bool              `json:"required"`
	Enum        []string          `json:"enum,omitempty"`        // widget=selector 时的合法闭合集
	HasWildcard bool              `json:"hasWildcard,omitempty"` // 枚举含 <占位> 通配段
	Default     any               `json:"default,omitempty"`     // scheme (=...) 或上下文约定缺省值
	Example     string            `json:"example,omitempty"`     // R-UX-1 缺少询问示例与占位符
	AutoFill    string            `json:"autoFill,omitempty"`    // subject|default|system|""（R-UX-1 步骤 2 自动补充标记）
	Min         *float64          `json:"min,omitempty"`         // widget=number 时的范围下限
	Max         *float64          `json:"max,omitempty"`         // widget=number 时的范围上限
	RefKind     string            `json:"refKind,omitempty"`     // widget=ref 时的目标对象 Kind 提示
	Elem        *FieldDescriptor  `json:"elem,omitempty"`        // widget=list 时的元素控件描述符
	Fields      []FieldDescriptor `json:"fields,omitempty"`      // widget=object 时的子字段控件描述符
}

var (
	parenDefaultRe = regexp.MustCompile(`[(\x{ff08}]\s*=\s*([^)\x{ff09}]+)[)\x{ff09}]`)
	parenRangeRe   = regexp.MustCompile(`[(\x{ff08}]\s*(-?\d+(?:\.\d+)?)\s*(?:-|\.\.)\s*(-?\d+(?:\.\d+)?)\s*[)\x{ff09}]`)
)

// proposalFieldOrder Proposal 字段的首选展示/询问顺序（新増动态字段按字典序追加在后，零硬编码丢字段）。
var proposalFieldOrder = []string{
	"title", "summary", "acceptance", "impact", "priority", "landing", "proposer", "status",
}

// fieldLabels 常见字段的中文可读标签（未知/新增字段自动回退为其字段名）。
var fieldLabels = map[string]string{
	"title":              "标题",
	"summary":            "目标",
	"acceptance":         "验收标准",
	"impact":             "影响面",
	"priority":           "优先级",
	"proposer":           "提出人",
	"status":             "状态",
	"landing":            "产物落点",
	"landing.target":     "产物落点",
	"landing.ref":        "产物引用",
	"target":             "落点类型",
	"ref":                "引用目标",
	"enabled":            "启用状态",
	"workservice":        "所属 WorkService",
	"workflow":           "绑定 Workflow",
	"config":             "配置映射 (map)",
	"ops":                "操作列表 (ops)",
	"goalRef":            "目标文档引用 (goalRef)",
	"descRef":            "描述文档引用 (descRef)",
	"docs":               "关联文档引用",
	"roles":              "绑定角色 (roles)",
	"groups":             "所属分组 (groups)",
	"providers":          "账号包引用 (providers)",
	"refs":               "逻辑引用 (refs)",
	"runtime":            "运行时",
	"autoLevel":          "自治级别 (autoLevel)",
	"kind":               "类型 (kind)",
	"connector":          "连接器引用",
	"display":            "显示名称",
	"purpose":            "用途说明",
	"ttl":                "缓存时长 (ttl)",
	"prompt":             "提示词 (prompt)",
	"llmPrompt":          "LLM 提示词",
	"download":           "下载源",
	"usage":              "用法说明",
	"category":           "分类",
	"sourceOps":          "来源 ConnectorOps",
	"requiredSections":   "必填章节",
	"writers":            "允许写主体",
	"admission":          "准入级别",
	"mandatory":          "强制级别",
	"observedGeneration": "观测代数",
	"phase":              "相位 (phase)",
	"limit":              "阈值 (limit)",
	"op":                 "比较算子 (op)",
}

// DescribeKindWidgets 从 Kind 注册项派生机读字段控件描述符列表（R-UX-2 单一真相）。
func DescribeKindWidgets(k Kind) []FieldDescriptor {
	src := k.Spec
	if len(src) == 0 && len(k.Fields) > 0 {
		src = k.Fields
	}
	if len(src) == 0 {
		return nil
	}
	names := orderedKeys(k.Name, src)
	out := make([]FieldDescriptor, 0, len(names))
	for _, name := range names {
		out = append(out, describeField(k.Name, name, name, src[name]))
	}
	return out
}

func orderedKeys(kindName string, m map[string]any) []string {
	keys := make([]string, 0, len(m))
	seen := make(map[string]bool, len(m))
	if kindName == "Proposal" {
		for _, pref := range proposalFieldOrder {
			if _, ok := m[pref]; ok {
				keys = append(keys, pref)
				seen[pref] = true
			}
		}
	}
	var rest []string
	for k := range m {
		if !seen[k] {
			rest = append(rest, k)
		}
	}
	sort.Strings(rest)
	return append(keys, rest...)
}

func describeField(kindName, name, path string, raw any) FieldDescriptor {
	d := parseDeclValue(raw)
	rawStr := ""
	if s, ok := raw.(string); ok {
		rawStr = s
	}
	fd := FieldDescriptor{
		Name:     name,
		Path:     path,
		Label:    labelFor(path, name),
		RawDecl:  rawStr,
		Required: d.required,
	}
	if def, ok := parseParenDefault(rawStr); ok {
		fd.Default = def
		fd.AutoFill = "default"
	}
	if minV, maxV, ok := parseParenRange(rawStr); ok {
		fd.Min = &minV
		fd.Max = &maxV
	}

	switch {
	case d.typ == "object":
		fd.Type = "object"
		fd.Widget = "object"
		subKeys := make([]string, 0, len(d.fields))
		if path == "landing" {
			if _, ok := d.fields["target"]; ok {
				subKeys = append(subKeys, "target")
			}
		}
		var rest []string
		for sk := range d.fields {
			if path == "landing" && sk == "target" {
				continue
			}
			rest = append(rest, sk)
		}
		sort.Strings(rest)
		subKeys = append(subKeys, rest...)
		var rawMap map[string]any
		if m, ok := raw.(map[string]any); ok {
			rawMap = m
		}
		for _, sk := range subKeys {
			subPath := path + "." + sk
			if rawMap != nil {
				fd.Fields = append(fd.Fields, describeField(kindName, sk, subPath, rawMap[sk]))
			} else {
				fd.Fields = append(fd.Fields, describeFromDecl(kindName, sk, subPath, d.fields[sk]))
			}
		}
	case d.typ == "list":
		fd.Type = "list"
		fd.Widget = "list"
		if d.elem != nil {
			elemDesc := describeFromDecl(kindName, name+"Item", path+"[]", *d.elem)
			fd.Elem = &elemDesc
		} else {
			elemDesc := FieldDescriptor{Name: name + "Item", Path: path + "[]", Label: fd.Label, Type: "string", Widget: "input"}
			fd.Elem = &elemDesc
		}
	case len(d.enum) > 0:
		fd.Type = "enum"
		fd.Widget = "selector"
		fd.Enum = append([]string(nil), d.enum...)
		fd.HasWildcard = d.hasWildcard
	case d.typ == "ref":
		fd.Type = "ref"
		fd.Widget = "ref"
		fd.RefKind = inferRefKind(kindName, name, path)
	case d.typ == "bool":
		fd.Type = "bool"
		fd.Widget = "switch"
	case d.typ == "map":
		fd.Type = "map"
		fd.Widget = "map"
	case d.typ == "int" || d.typ == "number":
		fd.Type = "number"
		fd.Widget = "number"
	default:
		if d.typ == "text" || isLongTextField(kindName, name, path) {
			fd.Type = "text"
			fd.Widget = "textarea"
		} else {
			if d.typ == "" || d.typ == "any" {
				fd.Type = "string"
			} else {
				fd.Type = d.typ
			}
			fd.Widget = "input"
		}
	}

	applyKindConventions(kindName, &fd)
	return fd
}

func describeFromDecl(kindName, name, path string, d fieldDecl) FieldDescriptor {
	fd := FieldDescriptor{
		Name:     name,
		Path:     path,
		Label:    labelFor(path, name),
		Required: d.required,
	}
	switch {
	case d.typ == "object":
		fd.Type = "object"
		fd.Widget = "object"
		subKeys := make([]string, 0, len(d.fields))
		if path == "landing" {
			if _, ok := d.fields["target"]; ok {
				subKeys = append(subKeys, "target")
			}
		}
		var rest []string
		for sk := range d.fields {
			if path == "landing" && sk == "target" {
				continue
			}
			rest = append(rest, sk)
		}
		sort.Strings(rest)
		subKeys = append(subKeys, rest...)
		for _, sk := range subKeys {
			subPath := path + "." + sk
			fd.Fields = append(fd.Fields, describeFromDecl(kindName, sk, subPath, d.fields[sk]))
		}
	case d.typ == "list":
		fd.Type = "list"
		fd.Widget = "list"
		if d.elem != nil {
			elemDesc := describeFromDecl(kindName, name+"Item", path+"[]", *d.elem)
			fd.Elem = &elemDesc
		} else {
			elemDesc := FieldDescriptor{Name: name + "Item", Path: path + "[]", Label: fd.Label, Type: "string", Widget: "input"}
			fd.Elem = &elemDesc
		}
	case len(d.enum) > 0:
		fd.Type = "enum"
		fd.Widget = "selector"
		fd.Enum = append([]string(nil), d.enum...)
		fd.HasWildcard = d.hasWildcard
	case d.typ == "ref":
		fd.Type = "ref"
		fd.Widget = "ref"
		fd.RefKind = inferRefKind(kindName, name, path)
	case d.typ == "bool":
		fd.Type = "bool"
		fd.Widget = "switch"
	case d.typ == "map":
		fd.Type = "map"
		fd.Widget = "map"
	case d.typ == "int" || d.typ == "number":
		fd.Type = "number"
		fd.Widget = "number"
	default:
		if d.typ == "text" || isLongTextField(kindName, name, path) {
			fd.Type = "text"
			fd.Widget = "textarea"
		} else {
			if d.typ == "" || d.typ == "any" {
				fd.Type = "string"
			} else {
				fd.Type = d.typ
			}
			fd.Widget = "input"
		}
	}
	applyKindConventions(kindName, &fd)
	return fd
}

func labelFor(path, name string) string {
	if l, ok := fieldLabels[path]; ok {
		return l
	}
	if l, ok := fieldLabels[name]; ok {
		return l
	}
	return name
}

func isLongTextField(kindName, name, path string) bool {
	switch name {
	case "summary", "impact", "prompt", "llmPrompt", "description", "usage", "purpose":
		return true
	}
	return false
}

func inferRefKind(kindName, name, path string) string {
	lower := strings.ToLower(name)
	switch {
	case strings.Contains(lower, "workservice"):
		return "WorkService"
	case strings.Contains(lower, "connector") || strings.Contains(lower, "ops"):
		return "ConnectorOps"
	case strings.Contains(lower, "provider"):
		return "Provider"
	case strings.Contains(lower, "role"):
		return "Role"
	case strings.Contains(lower, "group"):
		return "Group"
	case strings.Contains(lower, "agent"):
		return "Agent"
	case strings.Contains(lower, "system") || strings.Contains(lower, "service"):
		return "ServiceDesc"
	case strings.Contains(lower, "workflow"):
		return "Workflow"
	default:
		return ""
	}
}

func parseParenDefault(raw string) (any, bool) {
	m := parenDefaultRe.FindStringSubmatch(raw)
	if len(m) < 2 {
		return nil, false
	}
	val := strings.TrimSpace(m[1])
	if val == "" || strings.Contains(val, "<") {
		return nil, false
	}
	if val == "true" {
		return true, true
	}
	if val == "false" {
		return false, true
	}
	if n, err := strconv.Atoi(val); err == nil {
		return n, true
	}
	return val, true
}

func parseParenRange(raw string) (float64, float64, bool) {
	m := parenRangeRe.FindStringSubmatch(raw)
	if len(m) < 3 {
		return 0, 0, false
	}
	minV, err1 := strconv.ParseFloat(m[1], 64)
	maxV, err2 := strconv.ParseFloat(m[2], 64)
	if err1 != nil || err2 != nil {
		return 0, 0, false
	}
	return minV, maxV, true
}

// applyKindConventions 补充 R-UX-1 询问示例与上下文自动推断约定（proposal.md §4.5 / R-UX-1）。
func applyKindConventions(kindName string, fd *FieldDescriptor) {
	if kindName == "Proposal" {
		switch fd.Path {
		case "title":
			fd.Example = "编辑器深色模式"
		case "summary":
			fd.Example = "对象编辑器提供深色主题，降低夜间巡检时的视觉疲劳"
		case "acceptance":
			fd.Required = true // Proposal 六要素之一（proposal.md §1/§3）
			fd.Example = "1. 主题切换即时生效且刷新后保持；2. 对比度满足 WCAG AA"
		case "impact":
			fd.Default = "console 前端组件（无写路径变更）"
			fd.Example = "core_system_dsh_plugin console 编辑器屏（纯前端改动，无写路径变更）"
		case "priority":
			fd.Default = "low"
			fd.AutoFill = "default"
			fd.Example = "low"
		case "proposer":
			fd.Default = "console-operator"
			fd.AutoFill = "subject"
			fd.Example = "console-operator"
		case "status":
			fd.Default = "submitted"
			fd.AutoFill = "system"
			fd.Example = "submitted"
		case "landing":
			fd.AutoFill = "default"
			fd.Default = map[string]any{"target": "doc"}
			fd.Example = "doc"
		case "landing.target":
			fd.Default = "doc"
			fd.AutoFill = "default"
			fd.Example = "doc"
		case "landing.ref":
			fd.Example = "docs/proposal-draft"
		}
		return
	}
	if fd.Example == "" {
		switch fd.Widget {
		case "selector":
			if len(fd.Enum) > 0 {
				fd.Example = fd.Enum[0]
				if fd.Default == nil && !fd.Required && !fd.HasWildcard {
					fd.Default = fd.Enum[0]
					fd.AutoFill = "default"
				}
			}
		case "switch":
			if fd.Default == nil {
				fd.Default = false
			}
			fd.Example = fmt.Sprintf("%v", fd.Default)
		case "number":
			fd.Example = "1"
		case "list":
			fd.Example = "条目 1, 条目 2"
		case "ref":
			if fd.RefKind != "" {
				fd.Example = strings.ToLower(fd.RefKind) + "/default"
			} else {
				fd.Example = "ref/default"
			}
		case "map":
			fd.Example = "key=value"
		default:
			fd.Example = fmt.Sprintf("示例 %s", fd.Label)
		}
	}
}
