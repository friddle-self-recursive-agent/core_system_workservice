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
	s := NewStore(nil)
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
	s := NewStore(nil)
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
	s := NewStore(nil)
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
	s := NewStore(nil)
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
	s := NewStore(nil)
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

// --- 持久化面（ops-log.md §2 draft 表；--ops-db 持久化部署：写穿 + 启动恢复） ---

// fakePersistence 测试替身：记录写穿调用序列与失败注入点。
type fakePersistence struct {
	puts    []string // 每次 PutDraft 的对象键（调用序）
	dels    []string // 每次 DeleteDraft 的对象键（调用序）
	rows    []*Draft // ListDrafts 返回的现态
	failPut bool     // 注入 PutDraft 失败
}

func (f *fakePersistence) PutDraft(d *Draft) error {
	if f.failPut {
		return errors.New("注入：存储不可写")
	}
	f.puts = append(f.puts, objKey(d.Kind, d.Scope, d.Name))
	f.rows = append(f.rows, d.clone())
	return nil
}

func (f *fakePersistence) DeleteDraft(kind, scope, name string) error {
	f.dels = append(f.dels, objKey(kind, scope, name))
	return nil
}

func (f *fakePersistence) ListDrafts() ([]*Draft, error) { return f.rows, nil }

func TestPutDeleteWriteThrough(t *testing.T) {
	f := &fakePersistence{}
	s := NewStore(&Options{Persistence: f})
	if err := s.Put(draftOf("core-headless", 2)); err != nil {
		t.Fatalf("落稿: %v", err)
	}
	if len(f.puts) != 1 || f.puts[0] != "Agent/global/core-headless" {
		t.Fatalf("落稿应写穿一次: %+v", f.puts)
	}
	time.Sleep(2 * time.Millisecond)
	if err := s.Put(draftOf("core-headless", 5)); err != nil { // 覆稿=写穿 upsert
		t.Fatalf("覆稿: %v", err)
	}
	if len(f.puts) != 2 {
		t.Fatalf("覆稿应再次写穿: %+v", f.puts)
	}
	if err := s.Delete("Agent", "global", "core-headless"); err != nil {
		t.Fatalf("弃稿: %v", err)
	}
	if len(f.dels) != 1 || f.dels[0] != "Agent/global/core-headless" {
		t.Fatalf("弃稿应写穿删除: %+v", f.dels)
	}
	// 写穿失败=落稿失败（可见即已持久：内存不动，重启恢复不丢稿也不多稿）。
	f.failPut = true
	if err := s.Put(draftOf("blocked", 1)); err == nil {
		t.Fatal("写穿失败应上抛")
	}
	if s.Len() != 0 {
		t.Fatalf("写穿失败内存不应入稿: Len=%d", s.Len())
	}
	if _, err := s.Get("Agent", "global", "blocked"); !errors.Is(err, ErrNoDraft) {
		t.Fatalf("写穿失败稿不应可见: %v", err)
	}
}

func TestRestoreFromPersistence(t *testing.T) {
	created := time.Date(2026, 11, 22, 8, 0, 0, 0, time.UTC)
	updated := created.Add(time.Hour)
	good1 := draftOf("alpha", 3)
	good1.CreatedAt, good1.UpdatedAt = created, updated
	good2 := draftOf("beta", 7)
	good2.Scope = "developer-system"
	good2.CreatedAt, good2.UpdatedAt = created, updated
	bad := &Draft{Kind: "Agent", Scope: "global", Name: "no-owner", Content: []byte("x")} // 无主稿
	f := &fakePersistence{rows: []*Draft{good1, bad, good2}}
	s := NewStore(&Options{Persistence: f})
	restored, skipped, err := s.Restore()
	if err != nil {
		t.Fatalf("恢复: %v", err)
	}
	if restored != 2 || skipped != 1 {
		t.Fatalf("恢复 restored=%d skipped=%d，要 2/1", restored, skipped)
	}
	got, err := s.Get("Agent", "developer-system", "beta")
	if err != nil || got.BaseRV != 7 {
		t.Fatalf("恢复稿回读: %v %+v", err, got)
	}
	if !got.CreatedAt.Equal(created) || !got.UpdatedAt.Equal(updated) {
		t.Fatalf("恢复应保留落稿时间戳: %+v", got)
	}
	// 整表替换语义：恢复前在内存的稿不留（恢复后内存态=持久化面现态，确定性）。
	if err := s.Put(draftOf("pre-restore", 1)); err != nil {
		t.Fatal(err)
	}
	f2 := &fakePersistence{rows: []*Draft{good1}}
	s2 := NewStore(&Options{Persistence: f2})
	if _, _, err := s2.Restore(); err != nil {
		t.Fatal(err)
	}
	if _, err := s2.Get("Agent", "global", "pre-restore"); !errors.Is(err, ErrNoDraft) {
		t.Fatal("Restore 应整表替换内存（确定性恢复）")
	}
	// 无持久化面 → 无恢复语义。
	if _, _, err := NewStore(nil).Restore(); !errors.Is(err, ErrNoStore) {
		t.Fatalf("纯内存恢复期望 ErrNoStore，得 %v", err)
	}
}
