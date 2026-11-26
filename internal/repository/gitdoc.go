// gitdoc.go —— git 后端的路径/文档编解码助手（自 git.go 按域拆分，R-ENG-7 ≤800）：
// scope 四层物化（scopeDir/scopePrefix/scopeIsEmpty）、storage 模板解析（storagePatterns/
// validPathTemplate/expandTemplate）、md frontmatter/YAML 文档编解码（splitDoc/metaOf/mapValue/
// setResourceVersion/encodeDoc/canonicalWithoutRV/removePair）、globObjects 读侧枚举。
// 设计依据同 git.go 头注：concepts/core.md §7 / revision.md / scope.md / layout.md §3。
package repository

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"gopkg.in/yaml.v3"
)

func scopeDir(scope string) (string, error) {
	switch scope {
	case "":
		return "", nil
	case "builtin":
		return "core", nil
	case "global":
		return "global", nil
	}
	segs := strings.Split(scope, "/")
	if len(segs) > 2 {
		return "", fmt.Errorf("%w: %q（scope 至多 系统/实例 两段）", ErrScope, scope)
	}
	for _, s := range segs {
		if !nameSeg.MatchString(s) {
			return "", fmt.Errorf("%w: 段 %q 不合命名规则（小写中划线）", ErrScope, s)
		}
	}
	dir := "systems/" + segs[0]
	if len(segs) == 2 {
		dir += "/instances/" + segs[1]
	}
	return dir, nil
}

// scopePrefix scope 对应的目录前缀（路径一致性闸门：模板展开结果必须落在本 scope 的目录之下）。
// scope="" 返回空串（跨 scope 通配，前缀不设限）。
func scopePrefix(scope string) (string, error) {
	if scope == "" {
		return "", nil
	}
	dir, err := scopeDir(scope)
	if err != nil {
		return "", err
	}
	return dir + "/", nil
}

// storagePatterns 解析 kind 的 storage 模板为路径模式列表。
//   - concrete（Get/Write）：scope/name 必须给全，返回恰一个具体相对路径；
//   - 通配（List）：未给定的段落展开为 `*`，返回全部可寻址模式。
//
// 模板分支有效性（scheme.yaml 个别条目混有范围词/散文，运行层按此闸门跳过）：
//   - 以 .yaml 结尾，且含 `<name>` 或完全不含占位符（固定单例路径，如 core/scheme.yaml）；
//   - 展开后必须以 scope 前缀开头（global 的对象不落 systems/，反之亦然）；
//   - 具体寻址下展开结果不得残留 `*` 或未识别占位符。
func storagePatterns(k scheme.Kind, scope, name string, concrete bool) ([]string, error) {
	segs := strings.Split(scope, "/")
	prefix, err := scopePrefix(scope)
	if err != nil {
		return nil, err
	}
	var patterns []string
	skipped := 0
	for alt := range strings.SplitSeq(k.Storage, "|") {
		alt = strings.TrimSpace(alt)
		if !validPathTemplate(alt) {
			skipped++
			continue
		}
		p, ok := expandTemplate(alt, segs, name, concrete)
		if !ok || !strings.HasPrefix(p, prefix) || strings.Contains(p, "*") && concrete {
			skipped++
			continue
		}
		patterns = append(patterns, p)
	}
	if len(patterns) == 0 {
		return nil, fmt.Errorf("%w: kind %q storage=%q 对 scope=%q name=%q 无可寻址分支（跳过 %d 个）",
			ErrScope, k.Name, k.Storage, scope, name, skipped)
	}
	sort.Strings(patterns)
	return patterns, nil
}

// validPathTemplate 分支是否为可寻址的路径模板。
func validPathTemplate(alt string) bool {
	if alt == "" || !strings.HasSuffix(alt, ".yaml") {
		return false
	}
	return strings.Contains(alt, "<name>") || !strings.Contains(alt, "<")
}

// expandTemplate 单分支占位符展开；不可满足返回 ok=false（如模板需系统段而 scope=global）。
// `…` 通配 = scope 目录相对 systems/<s>/ 的余部（空则整段收敛），兼容旧式模板。
func expandTemplate(alt string, segs []string, name string, concrete bool) (string, bool) {
	sys, inst := "", ""
	if len(segs) > 0 {
		sys = segs[0]
	}
	if len(segs) > 1 {
		inst = segs[1]
	}
	validSeg := func(v string) bool { return v == "*" || nameSeg.MatchString(v) }
	out := alt

	// <s>：系统段。builtin/global 作用域没有系统段——需要 <s> 的分支对该 scope 不可用。
	if strings.Contains(out, "<s>") {
		switch {
		case scopeIsEmpty(segs):
			out = strings.ReplaceAll(out, "<s>", "*") // List 跨 scope 通配
		case sys == "builtin" || sys == "global":
			return "", false
		case !validSeg(sys):
			return "", false
		default:
			out = strings.ReplaceAll(out, "<s>", sys)
		}
	}

	// <i>：实例段（scope 第二段；缺省时 List 通配、具体寻址不可用）。
	if strings.Contains(out, "<i>") {
		switch {
		case inst == "":
			if concrete {
				return "", false
			}
			out = strings.ReplaceAll(out, "<i>", "*")
		case !validSeg(inst):
			return "", false
		default:
			out = strings.ReplaceAll(out, "<i>", inst)
		}
	}

	// `…`：scope 目录余部（systems/<s>/ 之后的 instances/<i>/ 或空）。
	if strings.Contains(out, "…") {
		rest := ""
		if inst != "" {
			rest = "instances/" + inst + "/"
		}
		out = strings.ReplaceAll(out, "…/", rest)
		out = strings.ReplaceAll(out, "…", rest)
	}

	// <name>：对象名。空名仅在通配枚举下合法。
	if strings.Contains(out, "<name>") {
		if name == "" {
			if concrete {
				return "", false
			}
			name = "*"
		}
		out = strings.ReplaceAll(out, "<name>", name)
	}

	if strings.Contains(out, "<") || strings.Contains(out, ">") {
		return "", false // 仍有未识别占位符
	}
	return filepath.ToSlash(filepath.Clean(out)), true
}

func scopeIsEmpty(segs []string) bool {
	return len(segs) == 0 || (len(segs) == 1 && segs[0] == "")
}

// splitDoc 拆出文档与元数据容器：前导 `---` 围栏=md frontmatter（扁平键，layout.md §3：
// name/docType/scope 直接住顶层），否则整文=yaml（身份字段住 `metadata:` 区，scheme.yaml conventions）。
// 返回文档映射节点（保注释/顺序）、元数据容器节点、正文与形态标记。
func splitDoc(raw []byte) (doc, container *yaml.Node, body []byte, isMD bool, err error) {
	text := string(raw)
	if strings.HasPrefix(text, "---") {
		lines := strings.Split(text, "\n")
		if strings.TrimSpace(lines[0]) != "---" {
			return nil, nil, nil, false, fmt.Errorf("%w: md 首行须为 --- 围栏", ErrContent)
		}
		closeIdx := -1
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				closeIdx = i
				break
			}
		}
		if closeIdx < 0 {
			return nil, nil, nil, false, fmt.Errorf("%w: md frontmatter 未闭合", ErrContent)
		}
		front := strings.Join(lines[1:closeIdx], "\n")
		body = []byte(strings.Join(lines[closeIdx+1:], "\n"))
		var node yaml.Node
		if err := yaml.Unmarshal([]byte(front), &node); err != nil {
			return nil, nil, nil, false, fmt.Errorf("%w: frontmatter 解析失败: %v", ErrContent, err)
		}
		if len(node.Content) == 0 || node.Content[0].Kind != yaml.MappingNode {
			return nil, nil, nil, false, fmt.Errorf("%w: frontmatter 须为映射", ErrContent)
		}
		doc = node.Content[0]
		return doc, doc, body, true, nil // md：容器=frontmatter 根（扁平）
	}
	var node yaml.Node
	if err := yaml.Unmarshal(raw, &node); err != nil {
		return nil, nil, nil, false, fmt.Errorf("%w: yaml 解析失败: %v", ErrContent, err)
	}
	if len(node.Content) == 0 || node.Content[0].Kind != yaml.MappingNode {
		return nil, nil, nil, false, fmt.Errorf("%w: 文档根须为映射", ErrContent)
	}
	doc = node.Content[0]
	container = mapValue(doc, "metadata")
	if container == nil || container.Kind != yaml.MappingNode {
		return nil, nil, nil, false, fmt.Errorf("%w: 缺 metadata 区", ErrContent)
	}
	return doc, container, raw, false, nil
}

// metaOf 从元数据容器提取 name/scope/resourceVersion。
func metaOf(container *yaml.Node) (docMeta, error) {
	var m docMeta
	if container == nil || container.Kind != yaml.MappingNode {
		return m, fmt.Errorf("%w: 缺 metadata 区", ErrContent)
	}
	if n := mapValue(container, "name"); n != nil && n.Kind == yaml.ScalarNode {
		m.Name = n.Value
	}
	if s := mapValue(container, "scope"); s != nil && s.Kind == yaml.ScalarNode {
		m.Scope = s.Value
	}
	if rv := mapValue(container, "resourceVersion"); rv != nil && rv.Kind == yaml.ScalarNode {
		var v int64
		if _, err := fmt.Sscanf(rv.Value, "%d", &v); err == nil {
			m.RV = v
		}
	}
	return m, nil
}

// mapValue 在映射节点里取键的值节点（yaml.v3 映射：Content 交替存放 key/value）。
func mapValue(mapping *yaml.Node, key string) *yaml.Node {
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

// setResourceVersion 就地注入/改写元数据容器中的 resourceVersion（保留既有键序与注释）。
func setResourceVersion(container *yaml.Node, rv int64) error {
	if container == nil || container.Kind != yaml.MappingNode {
		return fmt.Errorf("%w: 缺 metadata 区", ErrContent)
	}
	if v := mapValue(container, "resourceVersion"); v != nil {
		v.Value = fmt.Sprintf("%d", rv)
		v.Tag = "!!int"
		v.Style = 0
		return nil
	}
	k := &yaml.Node{Kind: yaml.ScalarNode, Value: "resourceVersion"}
	v := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: fmt.Sprintf("%d", rv)}
	container.Content = append(container.Content, k, v)
	return nil
}

// encodeDoc 重编码文档：yaml 走节点树（保注释，格式规整为 2 空格缩进）；
// md 重排 frontmatter 围栏 + 原文正文。
func encodeDoc(doc *yaml.Node, body []byte, isMD bool) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return nil, fmt.Errorf("repository: 文档编码失败: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	out := buf.Bytes()
	if isMD {
		var merged bytes.Buffer
		merged.WriteString("---\n")
		merged.Write(out)
		merged.WriteString("---\n")
		merged.Write(body)
		out = merged.Bytes()
	}
	return out, nil
}

// canonicalWithoutRV 把原文规整为"去掉 resourceVersion"的规范文本（同一编码器产出），
// 幂等判定 = 两份规范文本相等（concepts/revision.md：机器只做整文档文本比对）。
func canonicalWithoutRV(raw []byte) (string, error) {
	doc, container, body, isMD, err := splitDoc(raw)
	if err != nil {
		return "", err
	}
	_ = doc
	removePair(container, "resourceVersion")
	out, err := encodeDoc(doc, body, isMD)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func removePair(mapping *yaml.Node, key string) {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			mapping.Content = append(mapping.Content[:i], mapping.Content[i+2:]...)
			return
		}
	}
}

// globObjects git 布局的读半边：按 storage 模板通配枚举 (kind, scope) 下全部对象文件
// （GitRepo.List 与 sqlite 后端种子导入共用——git 布局=导入导出格式口径的机械落点）。
func globObjects(root string, sch *scheme.Scheme, kind, scope string) ([]*Object, error) {
	k, err := sch.Lookup(kind)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotRegistered, err)
	}
	pats, err := storagePatterns(k, scope, "", false)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []*Object
	for _, pat := range pats {
		matches, _ := filepath.Glob(filepath.Join(root, filepath.FromSlash(pat)))
		sort.Strings(matches)
		for _, path := range matches {
			if seen[path] {
				continue
			}
			seen[path] = true
			raw, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			doc, container, _, _, err := splitDoc(raw)
			if err != nil {
				continue
			}
			_ = doc
			m, err := metaOf(container)
			if err != nil {
				continue
			}
			objName := m.Name
			if objName == "" {
				objName = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			}
			objScope := m.Scope
			if objScope == "" {
				objScope = scope
			}
			out = append(out, &Object{Kind: kind, Name: objName, Scope: objScope, ResourceVersion: m.RV, Content: raw})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Scope != out[j].Scope {
			return out[i].Scope < out[j].Scope
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// 写路径：串行队列 + 乐观锁 + 幂等 + git commit
// ─────────────────────────────────────────────────────────────────────────────

type writeReq struct {
	o      *Object
	fm     string
	reason string
	ctx    context.Context
	resp   chan writeResp
	del    *deleteReq // 非空=删除请求（ObjectDeleter——串行队列同通道复用；sqlite 后端消费）
}

// deleteReq 删除请求载荷（键+乐观锁期望——经同一串行队列与写事务边界）。
type deleteReq struct {
	kind, scope, name string
	expectRV          int64
}

type writeResp struct {
	rv     int64
	action string
	err    error
}
