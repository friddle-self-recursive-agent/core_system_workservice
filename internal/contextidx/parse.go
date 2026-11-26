// parse.go —— 对象 frontmatter → 索引条目（Doc chunk）的单对象切分器。
//
// 词形（与 repository.splitDoc 同构）：前导 `---` 围栏=md（元数据容器=frontmatter 根，
// 扁平键）；否则整文=yaml（容器=`metadata:` 区，scheme.yaml conventions；无 metadata 区=
// 根映射兜底——声明面对象 tags 直住根的导出词形兼容）。
// 采集字段（索引字段口径=gen-metadatainfo.mjs / R-DOC-6 / R-DOC-9）：
//   - tags: [{tag, tag_type}] 对清单（tagType 别名 tag_type/tagType 双收）；
//   - importance: 0–100 整数（缺省 0——不冒充缺省重要度）；
//   - title: frontmatter title，缺席取正文首个 `# ` 标题，再缺席=name；
//   - digest/summary: 摘要词形（S_kw 关键词第二来源）；
//   - trace: metadata.trace（yaml 嵌套）或 frontmatter 扁平 trace（md 导出词形）——R-DOC-9。
//
// 解析失败=返回 nil（增量路径=摘除自身 chunk；坏对象在写路径已被存储层 ErrContent 拦，
// 读面只对可解析文档负责）。
package contextidx

import (
	"strings"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"gopkg.in/yaml.v3"
)

// ParseDoc 对象 → 索引条目（单对象 chunk）。不可解析=nil。
func ParseDoc(o *repository.Object) *Doc {
	if o == nil || len(o.Content) == 0 {
		return nil
	}
	root, container, body, ok := splitFront(o.Content)
	if !ok {
		return nil
	}
	d := &Doc{Kind: o.Kind, Scope: o.Scope, Name: o.Name, RV: o.ResourceVersion}
	d.Title = scalarIn(container, root, "title")
	if d.Title == "" {
		d.Title = firstHeading(body)
	}
	if v, ok := intIn(container, root, "importance"); ok {
		d.Importance = v
	}
	d.Tags = tagsIn(container, root)
	d.Digest = scalarIn(container, root, "digest")
	if d.Digest == "" {
		d.Digest = scalarIn(container, root, "summary")
	}
	d.Trace = traceIn(container, root)
	// 关键词=标题+摘要+name（索引期一次切好——find 零正文扫描红线）。
	kwText := d.Title + " " + d.Digest + " " + d.Name + " " + strings.Join(tagNames(d.Tags), " ")
	d.Keywords = Tokenize(kwText)
	if d.Title == "" {
		d.Title = o.Name
	}
	return d
}

// splitFront 拆 frontmatter：返回根映射、元数据容器、正文。
func splitFront(raw []byte) (root, container *yaml.Node, body []byte, ok bool) {
	text := string(raw)
	if strings.HasPrefix(text, "---") {
		lines := strings.Split(text, "\n")
		if strings.TrimSpace(lines[0]) != "---" {
			return nil, nil, nil, false
		}
		closeIdx := -1
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				closeIdx = i
				break
			}
		}
		if closeIdx < 0 {
			return nil, nil, nil, false
		}
		front := strings.Join(lines[1:closeIdx], "\n")
		var node yaml.Node
		if err := yaml.Unmarshal([]byte(front), &node); err != nil {
			return nil, nil, nil, false
		}
		if len(node.Content) == 0 || node.Content[0].Kind != yaml.MappingNode {
			return nil, nil, nil, false
		}
		root = node.Content[0]
		return root, root, []byte(strings.Join(lines[closeIdx+1:], "\n")), true
	}
	var node yaml.Node
	if err := yaml.Unmarshal(raw, &node); err != nil {
		return nil, nil, nil, false
	}
	if len(node.Content) == 0 || node.Content[0].Kind != yaml.MappingNode {
		return nil, nil, nil, false
	}
	root = node.Content[0]
	container = mapGet(root, "metadata")
	if container == nil || container.Kind != yaml.MappingNode {
		container = root // 无 metadata 区=根兜底（tags 直住根的词形）
	}
	return root, container, raw, true
}

// mapGet 映射节点取键值（yaml.v3 Content 交替 key/value）。
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

// scalarIn 标量字段查找（metadata 区优先、root 兜底——词形双收；缺席/非标量=空）。
func scalarIn(c1, c2 *yaml.Node, key string) string {
	if v := scalarOf(c1, key); v != "" {
		return v
	}
	return scalarOf(c2, key)
}

// scalarOf 容器内标量字段（缺席/非标量=空）。
func scalarOf(container *yaml.Node, key string) string {
	n := mapGet(container, key)
	if n == nil || n.Kind != yaml.ScalarNode {
		return ""
	}
	return strings.TrimSpace(n.Value)
}

// intOf 容器内整数字段（0–100 夹取——importance 标度红线）。
func intOf(container *yaml.Node, key string) (int, bool) {
	n := mapGet(container, key)
	if n == nil || n.Kind != yaml.ScalarNode {
		return 0, false
	}
	var v int
	if err := yaml.Unmarshal([]byte(n.Value), &v); err != nil {
		return 0, false
	}
	if v < 0 {
		v = 0
	}
	if v > 100 {
		v = 100
	}
	return v, true
}

// intIn 整数字段（metadata 区优先 root 兜底）。
func intIn(c1, c2 *yaml.Node, key string) (int, bool) {
	if v, ok := intOf(c1, key); ok {
		return v, true
	}
	return intOf(c2, key)
}

// tagsOf tags 对清单（[{tag, tag_type}]；tagType/tag_type 双收；缺 tag 名的项跳过）。
func tagsOf(container *yaml.Node) []Tag {
	n := mapGet(container, "tags")
	if n == nil || n.Kind != yaml.SequenceNode {
		return nil
	}
	var out []Tag
	for _, item := range n.Content {
		if item.Kind != yaml.MappingNode {
			// 字符串词形 "domain:console" 兼容收（导出/手写便捷词形）
			if item.Kind == yaml.ScalarNode && item.Value != "" {
				if typ, tag, found := strings.Cut(item.Value, ":"); found {
					out = append(out, Tag{Tag: tag, TagType: typ})
				}
			}
			continue
		}
		name := scalarOf(item, "tag")
		if name == "" {
			continue
		}
		typ := scalarOf(item, "tag_type")
		if typ == "" {
			typ = scalarOf(item, "tagType")
		}
		out = append(out, Tag{Tag: name, TagType: typ})
	}
	return out
}

// tagsIn tags 清单（metadata 区优先 root 兜底）。
func tagsIn(c1, c2 *yaml.Node) []Tag {
	if t := tagsOf(c1); len(t) > 0 {
		return t
	}
	return tagsOf(c2)
}

// traceIn metadata.trace 段（metadata 区优先 root 兜底——md 扁平与 yaml 嵌套双词形）。
func traceIn(c1, c2 *yaml.Node) Trace {
	var t Trace
	node := mapGet(c1, "trace")
	if node == nil {
		node = mapGet(c2, "trace")
	}
	if node == nil {
		return t
	}
	raw, _ := yaml.Marshal(node)
	_ = yaml.Unmarshal(raw, &t) // 结构不符=零值（trace 段畸形不阻断索引——诚实零投影）
	return t
}

// firstHeading 正文首个 `# ` 一级标题（title 缺席词形）。
func firstHeading(body []byte) string {
	for _, line := range strings.Split(string(body), "\n") {
		t := strings.TrimSpace(line)
		if strings.HasPrefix(t, "# ") {
			return strings.TrimSpace(strings.TrimPrefix(t, "# "))
		}
	}
	return ""
}

// tagNames tag 名清单（关键词拼入——tag 词进 S_kw 空间）。
func tagNames(tags []Tag) []string {
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		out = append(out, t.Tag)
	}
	return out
}
