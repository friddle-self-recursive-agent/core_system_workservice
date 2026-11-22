// draft_test.go —— 草稿缓存语义验收（concepts/apply.md §1/§6 rule 6）：
// 键全/无主/空稿拒绝、upsert 覆盖（CreatedAt 定格 UpdatedAt 更新）、读面返回副本、
// 可查/可弃、弃不存在的稿=ErrNoDraft、List 键序确定性。
package draft

import (
	"bytes"
	"errors"
	"testing"
	"time"
)

func draftOf(name string, rv int64) *Draft {
	return &Draft{
		Kind: "Agent", Scope: "global", Name: name,
		Content:      []byte("apiVersion: core/v1\nkind: Agent\nmetadata: {name: " + name + ", scope: global}\n"),
		FieldManager: "agent/dev-agent",
		ReasonRef:    "drift-report/42",
		BaseRV:       rv,
	}
}

func TestPutRejectsInvalid(t *testing.T) {
	s := NewStore()
	cases := map[string]*Draft{
		"nil 稿":        nil,
		"缺 kind":       {Scope: "global", Name: "a", FieldManager: "m", Content: []byte("x")},
		"缺 scope":      {Kind: "Agent", Name: "a", FieldManager: "m", Content: []byte("x")},
		"缺 name":       {Kind: "Agent", Scope: "global", FieldManager: "m", Content: []byte("x")},
		"无主（rule2 同源）": {Kind: "Agent", Scope: "global", Name: "a", Content: []byte("x")},
		"空稿":           {Kind: "Agent", Scope: "global", Name: "a", FieldManager: "m"},
	}
	for label, d := range cases {
		if err := s.Put(d); !errors.Is(err, ErrInvalid) {
			t.Fatalf("%s: 期望 ErrInvalid，得 %v", label, err)
		}
	}
	if s.Len() != 0 {
		t.Fatalf("非法稿不应入缓存：Len=%d", s.Len())
	}
}

func TestPutGetRoundTripAndCopyIsolation(t *testing.T) {
	s := NewStore()
	d := draftOf("core-headless", 2)
	if err := s.Put(d); err != nil {
		t.Fatalf("落稿: %v", err)
	}
	got, err := s.Get("Agent", "global", "core-headless")
	if err != nil {
		t.Fatalf("读稿: %v", err)
	}
	if got.BaseRV != 2 || !bytes.Equal(got.Content, d.Content) || got.FieldManager != "agent/dev-agent" {
		t.Fatalf("字段回读不符: %+v", got)
	}
	// 读面副本隔离：改动回读稿不影响缓存内正文。
	got.Content[0] = '#'
	again, _ := s.Get("Agent", "global", "core-headless")
	if bytes.Equal(again.Content, got.Content) {
		t.Fatal("Get 应返回副本（缓存正文被读面污染）")
	}
	// 落稿侧同源隔离：传入对象后续改动不回写缓存。
	d.Content[0] = '#'
	third, _ := s.Get("Agent", "global", "core-headless")
	if third.Content[0] == '#' {
		t.Fatal("Put 应克隆入缓存（调用方对象回写污染）")
	}
}

func TestUpsertOverwriteKeepsCreatedAt(t *testing.T) {
	s := NewStore()
	first := draftOf("ci", 1)
	if err := s.Put(first); err != nil {
		t.Fatalf("首稿: %v", err)
	}
	created := first.CreatedAt
	time.Sleep(2 * time.Millisecond)
	second := draftOf("ci", 3) // 迭代后的新稿：覆盖=最新代改稿
	second.FieldManager = "agent/reviewer"
	if err := s.Put(second); err != nil {
		t.Fatalf("覆稿: %v", err)
	}
	got, _ := s.Get("Agent", "global", "ci")
	if !got.CreatedAt.Equal(created) {
		t.Fatal("覆盖不应动 CreatedAt（首稿落点保留）")
	}
	if !got.UpdatedAt.After(created) {
		t.Fatal("覆盖应更新 UpdatedAt")
	}
	if got.BaseRV != 3 || got.FieldManager != "agent/reviewer" {
		t.Fatalf("覆盖后应为最新稿: %+v", got)
	}
	if s.Len() != 1 {
		t.Fatalf("同键覆盖不应增行: Len=%d", s.Len())
	}
}

func TestListSortedAndDiscard(t *testing.T) {
	s := NewStore()
	for _, n := range []string{"zed", "alpha", "mid"} {
		if err := s.Put(draftOf(n, 1)); err != nil {
			t.Fatalf("落稿 %s: %v", n, err)
		}
	}
	list := s.List()
	if len(list) != 3 {
		t.Fatalf("List 数量: %d", len(list))
	}
	for i := 1; i < len(list); i++ {
		if list[i-1].Name > list[i].Name {
			t.Fatalf("List 非键序: %s > %s", list[i-1].Name, list[i].Name)
		}
	}
	// 可弃：弃后可查面归零；再弃=ErrNoDraft。
	if err := s.Delete("Agent", "global", "mid"); err != nil {
		t.Fatalf("弃稿: %v", err)
	}
	if s.Len() != 2 {
		t.Fatalf("弃后 Len=%d", s.Len())
	}
	if err := s.Delete("Agent", "global", "mid"); !errors.Is(err, ErrNoDraft) {
		t.Fatalf("再弃期望 ErrNoDraft，得 %v", err)
	}
	if _, err := s.Get("Agent", "global", "mid"); !errors.Is(err, ErrNoDraft) {
		t.Fatalf("弃后读期望 ErrNoDraft，得 %v", err)
	}
}

func TestScopeEntersKey(t *testing.T) {
	s := NewStore()
	g := draftOf("deploy", 1)
	if err := s.Put(g); err != nil {
		t.Fatalf("global 稿: %v", err)
	}
	sys := draftOf("deploy", 5)
	sys.Scope = "developer-system"
	if err := s.Put(sys); err != nil {
		t.Fatalf("系统 scope 稿: %v", err)
	}
	if s.Len() != 2 {
		t.Fatalf("同名不同 scope 应两稿并存: Len=%d", s.Len())
	}
	sysGot, err := s.Get("Agent", "developer-system", "deploy")
	if err != nil || sysGot.BaseRV != 5 {
		t.Fatalf("系统 scope 稿回读: %v %+v", err, sysGot)
	}
}
