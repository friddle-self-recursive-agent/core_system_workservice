// sqlite_test.go —— 文档库 sqlite 后端验收（loop-33 文档 storage 走 sqlite 首段）：
// 乐观锁/幂等重入 noop/串行写/ops 双写钩子（与 git 后端同语义）/种子导入/watch 直通知/
// 凭据来源闸门（Provider 脱敏，双后端同闸）。
package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
)

func newSQLiteRepo(t *testing.T) (*SQLiteRepo, *memOps) {
	t.Helper()
	ops := &memOps{}
	r, err := OpenSQLite(t.TempDir()+"/docs.db", mustScheme(t), &Options{Ops: ops}, "")
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = r.Close() })
	return r, ops
}

// newSQLiteRepoAt 定点建库（跨进程 TOCTOU 锚测试用——第二进程句柄需打开同一文件）。
func newSQLiteRepoAt(t *testing.T, dir string) (*SQLiteRepo, *memOps) {
	t.Helper()
	ops := &memOps{}
	r, err := OpenSQLite(filepath.Join(dir, "docs.db"), mustScheme(t), &Options{Ops: ops}, "")
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = r.Close() })
	return r, ops
}

// memOps 最小 ops 双写钩子（记录 result 序列，供乐观锁/幂等断言）。
type memOps struct {
	mu    sync.Mutex
	kinds []string
}

func (m *memOps) AppendApply(l opslog.ApplyLog) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.kinds = append(m.kinds, l.Result)
	return nil
}

func (m *memOps) AppendEvent(opslog.EventLog) error { return nil }
func (m *memOps) Reconcile() (int, error)           { return 0, nil }

func (m *memOps) results() []string {
	m.mu.Lock()
	defer m.mu.Unlock()
	return append([]string{}, m.kinds...)
}

func TestSQLiteWriteCreateUpdateConflictNoop(t *testing.T) {
	r, ops := newSQLiteRepo(t)
	ctx := context.Background()

	rv, action, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "a1", Content: agentYAML("a1", "")}, "tester", "")
	if err != nil || action != "create" || rv != 1 {
		t.Fatalf("create: rv=%d action=%s err=%v", rv, action, err)
	}
	// 乐观锁更新：携带 rv=1 → update，rv=2。
	rv, action, err = r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "a1", ResourceVersion: 1, Content: agentYAML("a1", "\n  extra: 1")}, "tester", "")
	if err != nil || action != "update" || rv != 2 {
		t.Fatalf("update: rv=%d action=%s err=%v", rv, action, err)
	}
	// 乐观锁冲突：旧 rv=1 → ConflictError，现值不损。
	_, _, err = r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "a1", ResourceVersion: 1, Content: agentYAML("a1", "\n  extra: 2")}, "tester", "")
	var ce *ConflictError
	if !errors.As(err, &ce) || ce.Current != 2 {
		t.Fatalf("conflict: err=%v", err)
	}
	// 幂等重入：现内容原样携带现 rv → noop。
	_, action, err = r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "a1", ResourceVersion: 2, Content: agentYAML("a1", "\n  extra: 1")}, "tester", "")
	if err != nil || action != "noop" {
		t.Fatalf("noop: action=%s err=%v", action, err)
	}
	// 盲写已存在 → ErrAlreadyExists。
	_, _, err = r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "a1", Content: agentYAML("a1", "")}, "tester", "")
	if !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("blind write: err=%v", err)
	}
	got := ops.results()
	if len(got) != 4 || got[0] != "ok" || got[1] != "ok" || got[2] != "conflict" || got[3] != "noop" {
		t.Fatalf("apply_log 双写序列: %v", got)
	}
}

func TestSQLiteWriteGuards(t *testing.T) {
	r, _ := newSQLiteRepo(t)
	ctx := context.Background()
	cases := []struct {
		name   string
		obj    *Object
		fm     string
		wantIs error
	}{
		{"无主写入", &Object{Kind: "Agent", Scope: "global", Name: "a", Content: agentYAML("a", "")}, "", ErrNoFieldManager},
		{"野对象", &Object{Kind: "Nope", Scope: "global", Name: "a", Content: agentYAML("a", "")}, "fm", ErrNotRegistered},
		{"空 scope", &Object{Kind: "Agent", Name: "a", Content: agentYAML("a", "")}, "fm", ErrScope},
		{"坏命名", &Object{Kind: "Agent", Scope: "global", Name: "Bad_Name", Content: agentYAML("a", "")}, "fm", ErrScope},
		{"模板不相容", &Object{Kind: "Agent", Scope: "systems/dev", Name: "a", Content: agentYAML("a", "")}, "fm", ErrScope},
		{"缺 metadata", &Object{Kind: "Agent", Scope: "global", Name: "a", Content: []byte("spec: {}")}, "fm", ErrContent},
		{"身份不符", &Object{Kind: "Agent", Scope: "global", Name: "a", Content: agentYAML("b", "")}, "fm", ErrIdentity},
	}
	for _, c := range cases {
		if _, _, err := r.Write(ctx, c.obj, c.fm, ""); !errors.Is(err, c.wantIs) {
			t.Errorf("%s: 期望 %v，实得 %v", c.name, c.wantIs, err)
		}
	}
}

func TestSQLiteGetList(t *testing.T) {
	r, _ := newSQLiteRepo(t)
	ctx := context.Background()
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "a1", Content: agentYAML("a1", "")}, "fm", ""); err != nil {
		t.Fatal(err)
	}
	if _, _, err := r.Write(ctx, &Object{Kind: "Workflow", Scope: "global", Name: "w1",
		Content: []byte("apiVersion: core/v1\nkind: Workflow\nmetadata:\n  name: w1\n  scope: global\nspec:\n  steps: []\n")}, "fm", ""); err != nil {
		t.Fatal(err)
	}
	o, err := r.Get(ctx, "Agent", "global", "a1")
	if err != nil || o.Name != "a1" || o.ResourceVersion != 1 || o.Scope != "global" {
		t.Fatalf("get: %+v err=%v", o, err)
	}
	if _, err := r.Get(ctx, "Agent", "global", "absent"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("absent: err=%v", err)
	}
	// 跨 scope 通配：global 下两种 kind 互不串。
	ags, err := r.List(ctx, "Agent", "")
	if err != nil || len(ags) != 1 || ags[0].Name != "a1" {
		t.Fatalf("list agent: %v err=%v", ags, err)
	}
	wfs, err := r.List(ctx, "Workflow", "global")
	if err != nil || len(wfs) != 1 || wfs[0].Name != "w1" {
		t.Fatalf("list workflow: %v err=%v", wfs, err)
	}
	if _, err := r.List(ctx, "Nope", ""); !errors.Is(err, ErrNotRegistered) {
		t.Fatalf("list unregistered: err=%v", err)
	}
}

func TestSQLiteSeedFromLayout(t *testing.T) {
	r, _ := newSQLiteRepo(t)
	root := t.TempDir()
	// 手工摆一个 git 布局（<name> 与双分支模板各一；一个越界文件=不导入）。
	mk := func(rel, content string) {
		p := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(p, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	mk("global/agents/seed-a/agent.yaml", `apiVersion: core/v1
kind: Agent
metadata:
  name: seed-a
  scope: global
  resourceVersion: 5 # 布局文件自带 rv=种子导入保留（迁移不重排代）
spec:
  autoLevel: auto
`)
	mk("global/workflows/seed-w.yaml", "apiVersion: core/v1\nkind: Workflow\nmetadata:\n  name: seed-w\n  scope: global\nspec:\n  steps: []\n")
	mk("global/agents/rogue/agent.yaml", "spec: {}") // 损坏文件（缺 metadata）=跳过不导入
	n, err := r.SeedFromLayout(root)
	if err != nil {
		t.Fatalf("seed: %v", err)
	}
	if n != 2 {
		t.Fatalf("seed 导入行数=%d（期望 2）", n)
	}
	o, err := r.Get(context.Background(), "Agent", "global", "seed-a")
	if err != nil || o.ResourceVersion != 5 {
		t.Fatalf("seed get rv: %+v err=%v", o, err)
	}
	// issue #23 特殊写入路径记账：种子导入绕过 applyWrite 也必须落导入版本一行账——
	// 账自导入 rv 起算，fieldManager 诚实标注 seed:import（不伪造导入前的历史正文）。
	hs, err := r.DocHistory("Agent", "global", "seed-a")
	if err != nil || len(hs) != 1 || hs[0].Rev != 5 || hs[0].FieldManager != "seed:import" {
		t.Fatalf("seed 逐次账应恰 rev=5/seed:import 一行，得 %+v err=%v", hs, err)
	}
	// 携现 rv 更新=update rv=6（种子导入不重排代——乐观锁续链）。
	rv, action, err := r.Write(context.Background(), &Object{Kind: "Agent", Scope: "global", Name: "seed-a",
		ResourceVersion: 5, Content: agentYAML("seed-a", "\n  extra: 1")}, "fm", "")
	if err != nil || action != "update" || rv != 6 {
		t.Fatalf("seed 续链: rv=%d action=%s err=%v", rv, action, err)
	}
	hs, err = r.DocHistory("Agent", "global", "seed-a")
	if err != nil || len(hs) != 2 || hs[0].Rev != 6 || hs[1].Rev != 5 {
		t.Fatalf("续链后逐次账应 rev=[6 5]，得 %+v err=%v", hs, err)
	}
	// 非空库再导入=幂等零行。
	n2, err := r.SeedFromLayout(root)
	if err != nil || n2 != 0 {
		t.Fatalf("幂等种子导入: n=%d err=%v", n2, err)
	}
}

func TestSQLiteWatch(t *testing.T) {
	r, _ := newSQLiteRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, err := r.Watch(ctx)
	if err != nil {
		t.Fatal(err)
	}
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("watch 首信号超时")
	}
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "a1", Content: agentYAML("a1", "")}, "fm", ""); err != nil {
		t.Fatal(err)
	}
	select {
	case <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("watch 写后信号超时")
	}
}

func TestSQLiteCredentialGate(t *testing.T) {
	r, _ := newSQLiteRepo(t)
	ctx := context.Background()
	bad := []byte(`apiVersion: core/v1
kind: Provider
metadata:
  name: p1
  scope: global
spec:
  kind: email
  auth:
    scheme: plain
    secret:
      source: "s3cret-password" # 明文=闸门拒绝（provider.md 规则 2）
`)
	if _, _, err := r.Write(ctx, &Object{Kind: "Provider", Scope: "global", Name: "p1", Content: bad}, "fm", ""); !errors.Is(err, ErrCredentialSource) {
		t.Fatalf("明文凭据未拒绝: err=%v", err)
	}
	good := []byte(`apiVersion: core/v1
kind: Provider
metadata:
  name: p1
  scope: global
spec:
  kind: email
  auth:
    scheme: plain
    secret:
      source: "secret:smtp-auth" # 引用式=放行（运行时值只住部署侧 secretstore）
`)
	rv, _, err := r.Write(ctx, &Object{Kind: "Provider", Scope: "global", Name: "p1", Content: good}, "fm", "")
	if err != nil || rv != 1 {
		t.Fatalf("引用式放行: rv=%d err=%v", rv, err)
	}
}

func TestSQLiteConcurrentWritesSerial(t *testing.T) {
	r, _ := newSQLiteRepo(t)
	ctx := context.Background()
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_, _, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global",
				Name: fmt.Sprintf("c%d", i), Content: agentYAML(fmt.Sprintf("c%d", i), "")}, "fm", "")
			if err != nil {
				t.Errorf("并发写 %d: %v", i, err)
			}
		}(i)
	}
	wg.Wait()
	ags, err := r.List(ctx, "Agent", "global")
	if err != nil || len(ags) != 8 {
		t.Fatalf("并发写落库数=%d err=%v", len(ags), err)
	}
}

// TestSQLiteListMetaNoContent G-15 元数据直读面：ListMeta 行集键/rv/排序与 List 同一，
// Content=nil（列表投影不带正文——正文仅单读路径 Get 拉）；List 契约保持 Content 随行
// （行为兼容锚——调度器/索引/导出等全量消费方零迁移）。
func TestSQLiteListMetaNoContent(t *testing.T) {
	r, _ := newSQLiteRepo(t)
	ctx := context.Background()
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "a1", Content: agentYAML("a1", "")}, "fm", ""); err != nil {
		t.Fatal(err)
	}
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "a2", Content: agentYAML("a2", "")}, "fm", ""); err != nil {
		t.Fatal(err)
	}
	if _, _, err := r.Write(ctx, &Object{Kind: "Workflow", Scope: "global", Name: "w1",
		Content: []byte("apiVersion: core/v1\nkind: Workflow\nmetadata:\n  name: w1\n  scope: global\nspec:\n  steps: []\n")}, "fm", ""); err != nil {
		t.Fatal(err)
	}

	metas, err := r.ListMeta(ctx, "Agent", "")
	if err != nil || len(metas) != 2 {
		t.Fatalf("ListMeta 跨 scope 枚举: %v err=%v", metas, err)
	}
	for _, m := range metas {
		if len(m.Content) != 0 {
			t.Fatalf("ListMeta 行不得携带 content: %s/%s", m.Scope, m.Name)
		}
		if m.Kind != "Agent" || m.Scope != "global" || m.Name == "" || m.ResourceVersion == 0 {
			t.Fatalf("ListMeta 元数据列直读应齐（kind/scope/name/rv）: %+v", m)
		}
	}
	if metas[0].Name != "a1" || metas[1].Name != "a2" {
		t.Fatalf("ListMeta 排序应与 List 同一（scope,name 字典序）: %s, %s", metas[0].Name, metas[1].Name)
	}
	// scope 限定枚举同口径。
	metas, err = r.ListMeta(ctx, "Workflow", "global")
	if err != nil || len(metas) != 1 || metas[0].Name != "w1" {
		t.Fatalf("ListMeta scope 限定: %v err=%v", metas, err)
	}
	if _, err := r.ListMeta(ctx, "Nope", ""); !errors.Is(err, ErrNotRegistered) {
		t.Fatalf("ListMeta 未注册 kind 应 ErrNotRegistered: err=%v", err)
	}

	// 行为兼容锚：List 仍携 content（接口契约——键/rv 与 ListMeta 同源）。
	full, err := r.List(ctx, "Agent", "")
	if err != nil || len(full) != 2 {
		t.Fatalf("List 全量枚举: %v err=%v", full, err)
	}
	for _, o := range full {
		if len(o.Content) == 0 {
			t.Fatalf("List 行应携带 content（契约不变）: %s/%s", o.Scope, o.Name)
		}
	}
}

// TestSQLiteWALAndReadPool G-15 机械判据：文件库=WAL 模式+读连接池>1；内存库恒串行
// （:memory: 多连接=多独立库，池>1 即表分裂——串行保一致）。
func TestSQLiteWALAndReadPool(t *testing.T) {
	// 文件库：journal_mode=wal + maxConns>1。
	r, err := OpenSQLite(filepath.Join(t.TempDir(), "docs.db"), mustScheme(t), nil, "")
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	if r.maxConns <= 1 {
		t.Fatalf("文件库读连接池应 >1（G-15），得 %d", r.maxConns)
	}
	var mode string
	if err := r.db.QueryRow(`PRAGMA journal_mode`).Scan(&mode); err != nil {
		t.Fatal(err)
	}
	if mode != "wal" {
		t.Fatalf("文件库应 WAL 模式，得 %q", mode)
	}

	// 内存库：恒串行（池=1）。
	m, err := OpenSQLite("", mustScheme(t), nil, "")
	if err != nil {
		t.Fatal(err)
	}
	defer m.Close()
	if m.maxConns != 1 {
		t.Fatalf("内存库应恒串行（池=1），得 %d", m.maxConns)
	}
}
