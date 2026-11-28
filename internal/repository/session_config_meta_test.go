package repository

import "testing"

// ParseSidecarMeta 侧车只读投影（2026-11-28 metadatainfo 上屏批）三词形+缺席诚实。
func TestParseSidecarMeta(t *testing.T) {
	content := `# 元数据侧车（R-DOC-6）：正文以同名 .md 为准；导出/GitHub 镜像保留
tags:
  - {tag: console, tag_type: surface}
  - {tag: style, tag_type: domain}
  - domain:ux
  - bare
importance: 93
`
	tags, imp, ok := ParseSidecarMeta(content)
	if !ok {
		t.Fatal("有侧车内容应 ok")
	}
	if imp != 93 {
		t.Fatalf("importance=93, got %d", imp)
	}
	if len(tags) != 4 {
		t.Fatalf("4 tags, got %d: %+v", len(tags), tags)
	}
	if tags[0].Tag != "console" || tags[0].TagType != "surface" {
		t.Fatalf("map 词形解析错: %+v", tags[0])
	}
	if tags[2].Tag != "ux" || tags[2].TagType != "domain" {
		t.Fatalf("domain:b 前缀词形解析错: %+v", tags[2])
	}
	if tags[3].Tag != "bare" || tags[3].TagType != "" {
		t.Fatalf("bare 词形解析错: %+v", tags[3])
	}
	if _, _, ok := ParseSidecarMeta("# 只有注释\n"); ok {
		t.Fatal("两缺应 ok=false（诚实不冒充）")
	}
	if _, _, ok := ParseSidecarMeta(""); ok {
		t.Fatal("空侧车应 ok=false")
	}
	if _, imp, ok := ParseSidecarMeta("importance: 77\n"); !ok || imp != 77 {
		t.Fatalf("仅 importance 应 ok 且 77: %v %d", ok, imp)
	}
}
