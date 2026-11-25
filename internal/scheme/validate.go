// validate.go —— spec 结构准入校验器（R-TYPE：声明即规则，有规则即有 checker）。
// 规则来源：core_design/core/scheme.yaml 各 kind 注册的 spec 声明（Kind.Spec，map 形态的声明原文）。
// 标注语义（v0.13 声明的诚实机读；scheme.md「校验按 Scheme 注册的 schema 做」）：
//   - `!` = 必填（前置或收尾均可，如 "!string"、"{…}!"）；`?` = 可选；
//   - 无标注 = 不强制必填（声明未显式必填不臆测——现存对象面兼容优先），
//     但**出现即校验形状**（类型/枚举/嵌套结构，enum 亦同：出现即须命中）；
//   - `a|b|c` = 枚举：各段须为无空白 token；含 `<占位>` 的段=通配段（任意值命中，如 systems/<n>/*）；
//     `…`/`...` 段=开放式枚举标记（剥离后剩余段照枚举判）；含空白/=/冒号的段=散文 → 整体不判（透明放行）；
//   - `[]X` = 数组（元素递归按 X）；`{k: d, k2}` = 对象（字段递归；裸键=仅存在性；
//     含 `|` 或空白并集词的键名跳过——如 probes 的 httpGet|tcp|version 不可机读）；
//   - 未知 spec 字段放行（scheme.md「之后只能追加与扩容」，CRD 演进语义）；
//   - 文档解析失败不判（返回 nil）：不可解析文档由存储层 ErrContent 管辖（repository.splitDoc 权威），
//     准入只对可解析文档的 spec 结构负责。
//
// 应用点：apply.Engine.admit 第三步（kind 注册→权限准入→spec 结构），真执行与 dry-run 同闸门
// （cored-bridge.md §2.1）；准入先于幂等比对——坏内容即使与库内现存一致也拒绝。
package scheme

import (
	"fmt"
	"regexp"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

// ───────────────────────── 声明解析（scheme.yaml 标注 → 机读规则）─────────────────────────

// fieldDecl 一条字段的机读规则。
type fieldDecl struct {
	required    bool
	typ         string // any|string|int|bool|ref|duration|map|timestamp|list|object
	enum        []string
	hasWildcard bool                 // 枚举含 <占位> 段：任意值命中
	elem        *fieldDecl           // typ=list
	fields      map[string]fieldDecl // typ=object
}

var enumTokenRe = regexp.MustCompile(`^[A-Za-z0-9_./<>*-]+$`)
var fieldNameRe = regexp.MustCompile(`^[A-Za-z][A-Za-z0-9_-]*$`)

// parseDeclValue 解析声明原文（any 形态）：scheme.yaml 的 spec 声明混合两种形态——
// 带引号字符串（"!string"、"{…}!"）与真实嵌套映射（WorkService 的 connector:/data:/observation: 等）。
func parseDeclValue(raw any) fieldDecl {
	switch v := raw.(type) {
	case string:
		return parseDecl(v)
	case map[string]any:
		d := fieldDecl{typ: "object", fields: make(map[string]fieldDecl, len(v))}
		for key, fv := range v {
			d.fields[key] = parseDeclValue(fv)
		}
		return d
	case []any:
		d := fieldDecl{typ: "list"}
		if len(v) > 0 {
			e := parseDeclValue(v[0])
			d.elem = &e
		}
		return d
	default:
		return fieldDecl{typ: "any"}
	}
}

// parseDecl 解析一条声明原文为机读规则。
func parseDecl(decl string) fieldDecl {
	d := fieldDecl{typ: "any"}
	s := unquote(strings.TrimSpace(decl)) // 先去外层引号：yaml 值常整体带引号（"!string" 的 ! 在引号内）
	for {
		switch {
		case strings.HasPrefix(s, "!"):
			d.required = true
			s = strings.TrimSpace(s[1:])
			continue
		case strings.HasSuffix(s, "!"):
			d.required = true
			s = strings.TrimSpace(s[:len(s)-1])
			continue
		case strings.HasPrefix(s, "?"):
			s = strings.TrimSpace(s[1:])
			continue
		}
		break
	}
	s = unquote(s)
	if s == "" {
		return d
	}
	switch {
	case strings.HasPrefix(s, "[]"):
		d.typ = "list"
		e := parseDecl(s[2:])
		d.elem = &e
		return d
	case strings.HasPrefix(s, "[") && strings.HasSuffix(s, "]"):
		d.typ = "list"
		e := parseDecl(s[1 : len(s)-1])
		d.elem = &e
		return d
	case strings.HasPrefix(s, "{") && strings.HasSuffix(s, "}"):
		d.typ = "object"
		d.fields = parseFields(s[1 : len(s)-1])
		return d
	}
	s = stripParens(s)
	if parts := enumParts(s); parts != nil {
		d.enum = parts
		for _, p := range parts {
			if strings.Contains(p, "<") {
				d.hasWildcard = true
			}
		}
		return d
	}
	switch s {
	case "string", "text", "int", "bool", "ref", "duration", "map", "timestamp", "number":
		d.typ = s
	default:
		d.typ = "any" // 不可机读 → 仅存在性
	}
	return d
}

// unquote 去成对引号（声明内嵌枚举常带引号，如 'vault:|env:|keychain:'）。
func unquote(s string) string {
	s = strings.TrimSpace(s)
	if len(s) >= 2 && (s[0] == '\'' || s[0] == '"') && s[len(s)-1] == s[0] {
		return strings.TrimSpace(s[1 : len(s)-1])
	}
	return s
}

// stripParens 去括注（"(=true)"、"（默认 pause）"等散文注解）。
func stripParens(s string) string {
	for {
		l, r := strings.IndexAny(s, "(（"), strings.IndexAny(s, ")）")
		if l < 0 || r < l {
			return strings.TrimSpace(s)
		}
		s = strings.TrimSpace(s[:l] + s[r+1:])
	}
}

// enumParts 尝试把标量声明判为枚举；不可机读（段含空白/=/冒号等）返回 nil。
func enumParts(s string) []string {
	parts := strings.Split(s, "|")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		p = unquote(p)
		if p == "" || p == "…" || p == "..." {
			continue
		}
		if !enumTokenRe.MatchString(p) {
			return nil
		}
		out = append(out, p)
	}
	if len(out) < 2 {
		return nil
	}
	return out
}

// parseFields 解析内联对象字段（"kinds: []string, scopes: []string"）。
// 深度感知切分（嵌套 {}/[] 内的逗号不切）；裸键=仅存在性；并集词键名跳过。
func parseFields(s string) map[string]fieldDecl {
	out := map[string]fieldDecl{}
	for _, part := range splitTop(s, ',', '{', '[', '}', ']') {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		key, decl := part, ""
		if i := indexTop(part, ':', '{', '[', '}', ']'); i >= 0 {
			key, decl = strings.TrimSpace(part[:i]), strings.TrimSpace(part[i+1:])
		}
		key = unquote(key)
		if !fieldNameRe.MatchString(key) {
			continue // httpGet|tcp|version 等并集词/散文键：不可机读，跳过
		}
		out[key] = parseDecl(decl)
	}
	return out
}

// splitTop 按 sep 切分，深度感知（openers/closers 配对内的 sep 不切）。
func splitTop(s string, sep byte, pairs ...byte) []string {
	var out []string
	depth := 0
	start := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case containsByte(pairs, c, true):
			depth++
		case containsByte(pairs, c, false):
			if depth > 0 {
				depth--
			}
		case c == sep && depth == 0:
			out = append(out, s[start:i])
			start = i + 1
		}
	}
	out = append(out, s[start:])
	return out
}

// indexTop 首个深度 0 处的 sep 下标；无则 -1。
func indexTop(s string, sep byte, pairs ...byte) int {
	depth := 0
	for i := 0; i < len(s); i++ {
		switch c := s[i]; {
		case containsByte(pairs, c, true):
			depth++
		case containsByte(pairs, c, false):
			if depth > 0 {
				depth--
			}
		case c == sep && depth == 0:
			return i
		}
	}
	return -1
}

func containsByte(pairs []byte, c byte, open bool) bool {
	for i := 0; i < len(pairs); i += 2 {
		if open && pairs[i] == c {
			return true
		}
		if !open && i+1 < len(pairs) && pairs[i+1] == c {
			return true
		}
	}
	return false
}

// ───────────────────────── 校验器 ─────────────────────────

// SpecError spec 结构未过准入（违规清单按路径排序，机械可验）。
type SpecError struct {
	Kind       string
	Violations []string
}

func (e *SpecError) Error() string {
	return fmt.Sprintf("scheme: %s spec 未过准入（%d 项）：%s", e.Kind, len(e.Violations),
		strings.Join(e.Violations, "；"))
}

// ValidateDoc 解析对象文档原文并校验 spec 区（apply 引擎准入用；真执行与 dry-run 同闸门）。
// 解析失败返回 nil：不可解析文档由存储层 ErrContent 管辖，准入只对可解析文档的 spec 负责。
// md+frontmatter 文档只取 frontmatter（正文散文不判）；无 spec 区 = 空节点（按声明必填性判）。
func ValidateDoc(k Kind, raw []byte) error {
	if len(k.Spec) == 0 {
		return nil // 无声明=无规则
	}
	spec, err := docSpecNode(raw)
	if err != nil {
		return nil // 解析不可达=存储层管辖
	}
	return ValidateSpec(k, spec)
}

// ValidateSpec 依据 kind 注册声明校验 spec 区（nil=文档无 spec 区：声明含必填字段→拒绝）。
func ValidateSpec(k Kind, spec *yaml.Node) error {
	if len(k.Spec) == 0 {
		return nil
	}
	names := make([]string, 0, len(k.Spec))
	decls := make(map[string]fieldDecl, len(k.Spec))
	for name, raw := range k.Spec {
		names = append(names, name)
		decls[name] = parseDeclValue(raw)
	}
	sort.Strings(names) // 违规输出确定性（机械对账锚点）
	var violations []string
	for _, name := range names {
		d := decls[name]
		node := spec
		if spec != nil {
			node = mapGet(spec, name)
		}
		if node == nil || isNull(node) {
			if d.required {
				violations = append(violations, name+": 必填段缺失")
			}
			continue
		}
		violations = append(violations, checkNode(name, node, d)...)
	}
	if len(violations) == 0 {
		return nil
	}
	return &SpecError{Kind: k.Name, Violations: violations}
}

// checkNode 按字段规则递归校验节点；返回违规清单（空=通过）。
func checkNode(path string, node *yaml.Node, d fieldDecl) []string {
	switch d.typ {
	case "object":
		if node.Kind != yaml.MappingNode {
			return []string{path + ": 须为映射"}
		}
		keys := make([]string, 0, len(d.fields))
		for key := range d.fields {
			keys = append(keys, key)
		}
		sort.Strings(keys)
		var out []string
		for _, key := range keys {
			fd := d.fields[key]
			child := mapGet(node, key)
			if child == nil || isNull(child) {
				if fd.required {
					out = append(out, path+"."+key+": 必填段缺失")
				}
				continue
			}
			out = append(out, checkNode(path+"."+key, child, fd)...)
		}
		return out
	case "list":
		if node.Kind != yaml.SequenceNode {
			return []string{path + ": 须为数组"}
		}
		if d.elem == nil {
			return nil
		}
		var out []string
		for i, item := range node.Content {
			if isNull(item) {
				continue // 数组中的 null 元素不深究（声明语义未定义）
			}
			out = append(out, checkNode(fmt.Sprintf("%s[%d]", path, i), item, *d.elem)...)
		}
		return out
	}
	if len(d.enum) > 0 {
		if node.Kind != yaml.ScalarNode {
			return []string{path + ": 枚举须为标量"}
		}
		if d.hasWildcard {
			return nil // 含 <占位> 段=通配枚举（模板段任意值命中）
		}
		for _, v := range d.enum {
			if node.Value == v {
				return nil
			}
		}
		return []string{fmt.Sprintf("%s: 值 %q 不在枚举 [%s]", path, node.Value, strings.Join(d.enum, "|"))}
	}
	switch d.typ {
	case "map":
		if node.Kind != yaml.MappingNode {
			return []string{path + ": 须为映射"}
		}
	case "string", "text", "ref":
		if node.Kind != yaml.ScalarNode || node.Tag != "!!str" {
			return []string{path + ": 须为字符串"}
		}
	case "int", "number":
		if node.Kind != yaml.ScalarNode || (node.Tag != "!!int" && node.Tag != "!!float") {
			return []string{path + ": 须为数值"}
		}
	case "bool":
		if node.Kind != yaml.ScalarNode || node.Tag != "!!bool" {
			return []string{path + ": 须为布尔"}
		}
	case "duration":
		if node.Kind != yaml.ScalarNode || (node.Tag != "!!str" && node.Tag != "!!int") {
			return []string{path + ": 须为时长（如 5m）"}
		}
	case "timestamp":
		if node.Kind != yaml.ScalarNode {
			return []string{path + ": 须为时间戳标量"}
		}
	}
	return nil
}

// docSpecNode 从文档原文提取 spec 区节点（yaml 全文或 md frontmatter）。
func docSpecNode(raw []byte) (*yaml.Node, error) {
	text := string(raw)
	if strings.HasPrefix(text, "---") {
		lines := strings.Split(text, "\n")
		if strings.TrimSpace(lines[0]) != "---" {
			return nil, fmt.Errorf("md 首行须为 --- 围栏")
		}
		closeIdx := -1
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				closeIdx = i
				break
			}
		}
		if closeIdx < 0 {
			return nil, fmt.Errorf("md frontmatter 未闭合")
		}
		text = strings.Join(lines[1:closeIdx], "\n")
	}
	var node yaml.Node
	if err := yaml.Unmarshal([]byte(text), &node); err != nil {
		return nil, err
	}
	if len(node.Content) == 0 || node.Content[0].Kind != yaml.MappingNode {
		return nil, fmt.Errorf("文档根须为映射")
	}
	return mapGet(node.Content[0], "spec"), nil
}

// mapGet 在映射节点里取键的值节点（yaml.v3 映射：Content 交替存放 key/value）。
func mapGet(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

// isNull 判 yaml 空值节点（`key:` 留空 = 视为缺席）。
func isNull(n *yaml.Node) bool {
	return n.Kind == yaml.ScalarNode && n.Tag == "!!null"
}
