package contextidx

import (
	"reflect"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
)

// ParseDocFromYAMLForTest 裸 YAML 体 → 索引条目（测试辅助——对象形包装 ParseDoc）。
func ParseDocFromYAMLForTest(body []byte) *Doc {
	return ParseDoc(&repository.Object{Kind: "Task", Scope: "global", Name: "t1", ResourceVersion: 1, Content: body})
}

// 2026-11-28 tags 批：metadata.tags 字符串清单三词形（前缀/裸词/映射对）全部入索引——
// 存量对象（Task/Template/Proposal 裸词形 tags）零迁移可检索。
func TestTagsOfStringWordForms(t *testing.T) {
	body := []byte(`kind: Task
metadata:
  name: t1
  scope: global
  tags: ["domain:workflow", "console", {tag: audit, tag_type: mech}]
spec:
  title: x
`)
	d := ParseDocFromYAMLForTest(body)
	if d == nil {
		t.Fatalf("parse nil")
	}
	want := []Tag{{Tag: "workflow", TagType: "domain"}, {Tag: "console"}, {Tag: "audit", TagType: "mech"}}
	if !reflect.DeepEqual(d.Tags, want) {
		t.Fatalf("tags=%+v want %+v", d.Tags, want)
	}
	// 残词形："domain:"（前缀无 tag 名）=跳过不冒充
	body2 := []byte(`kind: Task
metadata:
  tags: ["domain:", "  ", "ux"]
`)
	d2 := ParseDocFromYAMLForTest(body2)
	if d2 == nil {
		t.Fatalf("parse2 nil")
	}
	if !reflect.DeepEqual(d2.Tags, []Tag{{Tag: "ux"}}) {
		t.Fatalf("tags2=%+v want [{ux}]", d2.Tags)
	}
}
