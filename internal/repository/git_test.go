// git_test.go —— git 后端验收：并发写串行化、乐观锁冲突、幂等重入（loop-2 起点三必测）
// + storage 模板解析、watch、注释保留、写守卫、ops 双写钩子。
package repository

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// testScheme 结构对齐 core_design/core/scheme.yaml：覆盖全部模板形态
// （固定路径、<s>/<name>、<s>/<i>/<name>、双分支交替、多分支交替）。
const testScheme = `apiVersion: core/v1
kind: Scheme
metadata: {name: test, scope: builtin}
spec:
  version: core/v1
  kinds:
    - {kind: Scheme, builtin: true, storage: core/scheme.yaml}
    - {kind: Agent, builtin: true, storage: global/agents/<name>/agent.yaml}
    - {kind: ServiceDesc, builtin: true, storage: systems/<name>/system.yaml}
    - {kind: ServiceInstance, builtin: true, storage: systems/<s>/instances/<name>/instance.yaml}
    - {kind: Workflow, builtin: true, storage: "global/workflows/<name>.yaml | systems/<s>/workflows/<name>.yaml"}
    - {kind: ConnectorOps, builtin: true, storage: "global/connector-ops/<name>.yaml | systems/<s>/connector-ops/<name>.yaml | systems/<s>/instances/<i>/connector-ops/<name>.yaml"}
    - {kind: Ref, builtin: true, storage: "global/refs/<name>.yaml | systems/<s>/refs/<name>.yaml"}
    - {kind: Skill, builtin: true, storage: systems/<s>/instances/<i>/skills/<name>.yaml}
    - {kind: Provider, builtin: true, storage: global/providers/<name>.yaml}
    - {kind: WorkService, builtin: true, storage: global/workservices/<name>/workservice.yaml}
`

func mustScheme(t *testing.T) *scheme.Scheme {
	t.Helper()
	s, err := scheme.Load([]byte(testScheme))
	if err != nil {
		t.Fatalf("载入测试 scheme: %v", err)
	}
	return s
}

func newRepo(t *testing.T) *GitRepo {
	t.Helper()
	r, err := NewGit(t.TempDir(), mustScheme(t), &Options{WatchInterval: 10 * time.Millisecond})
	if err != nil {
		t.Fatalf("NewGit: %v", err)
	}
	t.Cleanup(func() { _ = r.Close() })
	return r
}

func (r *GitRepo) commitCount(t *testing.T) int {
	t.Helper()
	out, err := r.git(context.Background(), "rev-list", "--count", "HEAD")
	if err != nil {
		t.Fatalf("rev-list: %v", err)
	}
	var n int
	fmt.Sscanf(out, "%d", &n)
	return n
}

func agentYAML(name, extra string) []byte {
	return []byte(fmt.Sprintf(`apiVersion: core/v1
kind: Agent
metadata:
  name: %s # 注释保留探针
  scope: global
spec:
  autoLevel: auto%s
`, name, extra))
}

// ───────────────────────── storage 模板解析 ─────────────────────────

func TestStoragePathResolution(t *testing.T) {
	r := newRepo(t)
	s := mustScheme(t)
	cases := []struct {
		kind, scope, name string
		want              string
		wantErr           error
	}{
		{"Scheme", "builtin", "default", "core/scheme.yaml", nil},
		{"Agent", "global", "core-headless", "global/agents/core-headless/agent.yaml", nil},
		{"ServiceDesc", "developer-system", "developer-system", "systems/developer-system/system.yaml", nil},
		{"ServiceInstance", "developer-system", "ci", "systems/developer-system/instances/ci/instance.yaml", nil},
		{"Workflow", "global", "default-task", "global/workflows/default-task.yaml", nil},
		{"Workflow", "developer-system", "review", "systems/developer-system/workflows/review.yaml", nil},
		{"ConnectorOps", "global", "onedev", "global/connector-ops/onedev.yaml", nil},
		{"ConnectorOps", "developer-system", "git", "systems/developer-system/connector-ops/git.yaml", nil},
		{"ConnectorOps", "developer-system/ci", "git", "systems/developer-system/instances/ci/connector-ops/git.yaml", nil},
		{"Skill", "developer-system/ci", "pr-open", "systems/developer-system/instances/ci/skills/pr-open.yaml", nil},
		{"Ref", "global", "main-nas", "global/refs/main-nas.yaml", nil},
		// WorkService global 桶（v0.44 收口：作用域=Global 物化——旧根级无 scope 段模板不可寻址）
		{"WorkService", "global", "notify", "global/workservices/notify/workservice.yaml", nil},
		// 错误分支
		{"WorkService", "notify", "x", "", ErrScope}, // 模板仅 global 分支——系统 scope 无可寻址分支
		{"Workflow", "builtin", "w", "", ErrScope},      // 模板无 builtin 分支
		{"ServiceInstance", "", "ci", "", ErrScope},     // 具体寻址必须带 scope
		{"Agent", "global", "../escape", "", ErrScope},  // 名越界
		{"Agent", "../../etc", "x", "", ErrScope},       // scope 越界
		{"Agent", "a/b/c", "x", "", ErrScope},           // scope 超四层
		{"Agent", "global", "Upper_Kase", "", ErrScope}, // 命名规则
		{"NotRegistered", "global", "x", "", ErrNotRegistered},
	}
	for _, c := range cases {
		k, lookupErr := s.Lookup(c.kind)
		if c.wantErr == ErrNotRegistered {
			if lookupErr == nil {
				t.Errorf("%s 未注册却 Lookup 成功", c.kind)
			}
			continue
		}
		if lookupErr != nil {
			t.Fatalf("Lookup(%s): %v", c.kind, lookupErr)
		}
		pats, err := storagePatterns(k, c.scope, c.name, true)
		if c.wantErr != nil {
			// 名字/scope 级校验在 concretePath（全管线），此处补跑它
			_, cerr := r.concretePath(k, c.scope, c.name)
			if !errors.Is(err, c.wantErr) && !errors.Is(cerr, c.wantErr) {
				t.Errorf("%s/%s/%s 错误 = %v/%v, 要 %v", c.kind, c.scope, c.name, err, cerr, c.wantErr)
			}
			continue
		}
		if err != nil {
			t.Errorf("%s/%s/%s 意外失败: %v", c.kind, c.scope, c.name, err)
			continue
		}
		if len(pats) != 1 || pats[0] != c.want {
			t.Errorf("%s/%s/%s = %v, 要 [%s]", c.kind, c.scope, c.name, pats, c.want)
		}
	}
}

func TestStorageWildcardPatterns(t *testing.T) {
	s := mustScheme(t)
	cases := []struct {
		kind, scope string
		want        []string
	}{
		{"Scheme", "", []string{"core/scheme.yaml"}},
		{"Agent", "", []string{"global/agents/*/agent.yaml"}},
		{"Workflow", "", []string{"global/workflows/*.yaml", "systems/*/workflows/*.yaml"}},
		{"ServiceInstance", "developer-system", []string{"systems/developer-system/instances/*/instance.yaml"}},
		{"Skill", "developer-system", []string{"systems/developer-system/instances/*/skills/*.yaml"}},
		{"ConnectorOps", "developer-system", []string{
			"systems/developer-system/connector-ops/*.yaml",
			"systems/developer-system/instances/*/connector-ops/*.yaml",
		}},
	}
	for _, c := range cases {
		k, _ := s.Lookup(c.kind)
		got, err := storagePatterns(k, c.scope, "", false)
		if err != nil {
			t.Fatalf("%s/%s: %v", c.kind, c.scope, err)
		}
		if strings.Join(got, "|") != strings.Join(c.want, "|") {
			t.Errorf("%s/%s = %v, 要 %v", c.kind, c.scope, got, c.want)
		}
	}
}

// ───────────────────────── 写路径三必测 ─────────────────────────

// 幂等重入：同内容重复 apply（含携带新版本/零版本重放）→ no-op，不产生新提交。
func TestWriteIdempotentReentry(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	body := agentYAML("core-headless", "")

	rv, _, err := r.Write(ctx, &Object{Kind: "Agent", Name: "core-headless", Scope: "global", Content: body}, "human:friddle", "plan:loop-2")
	if err != nil || rv != 1 {
		t.Fatalf("首次创建: rv=%d err=%v", rv, err)
	}
	if n := r.commitCount(t); n != 1 {
		t.Fatalf("首次创建后提交数 = %d, 要 1", n)
	}
	// 落盘内容带 rv 注入
	obj, err := r.Get(ctx, "Agent", "global", "core-headless")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if obj.ResourceVersion != 1 || !strings.Contains(string(obj.Content), "resourceVersion: 1") {
		t.Fatalf("落盘对象缺 rv 注入: rv=%d content=%s", obj.ResourceVersion, obj.Content)
	}
	// 幂等重入（原稿内容重放，expected=1）
	rv2, _, err := r.Write(ctx, &Object{Kind: "Agent", Name: "core-headless", Scope: "global", ResourceVersion: 1, Content: body}, "human:friddle", "")
	if err != nil || rv2 != 1 {
		t.Fatalf("幂等重写: rv=%d err=%v", rv2, err)
	}
	// 幂等重放（expected=0 的 create 语义撞上同内容 → 仍 no-op）
	rv3, act, err := r.Write(ctx, &Object{Kind: "Agent", Name: "core-headless", Scope: "global", Content: body}, "human:friddle", "")
	if err != nil || rv3 != 1 || act != "noop" {
		t.Fatalf("零版本重放: rv=%d action=%s err=%v（要 noop）", rv3, act, err)
	}
	if n := r.commitCount(t); n != 1 {
		t.Fatalf("幂等重入不应产生提交：提交数 = %d, 要 1", n)
	}
}

// 乐观锁冲突：陈旧版本写被拒、现值不损、错误携带版本对。
func TestWriteOptimisticLockConflict(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	v1 := agentYAML("a", "")
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Name: "a", Scope: "global", Content: v1}, "fm", ""); err != nil {
		t.Fatal(err)
	}
	// 正常推进：expected=1 → rv 2
	if rv, _, err := r.Write(ctx, &Object{Kind: "Agent", Name: "a", Scope: "global", ResourceVersion: 1,
		Content: agentYAML("a", "\n  retry:\n    maxAttempts: 3")}, "agent:x", "plan:p1"); err != nil || rv != 2 {
		t.Fatalf("正常更新: rv=%d err=%v", rv, err)
	}
	if n := r.commitCount(t); n != 2 {
		t.Fatalf("提交数 = %d, 要 2", n)
	}
	// 陈旧写：expected=1（现库 2）→ Conflict，且 2<3（防活锁预算由上层管）
	_, _, err := r.Write(ctx, &Object{Kind: "Agent", Name: "a", Scope: "global", ResourceVersion: 1,
		Content: agentYAML("a", "\n  stale: true")}, "agent:y", "")
	var ce *ConflictError
	if !errors.As(err, &ce) {
		t.Fatalf("陈旧写应得 ConflictError, 得 %v", err)
	}
	if ce.Expected != 1 || ce.Current != 2 {
		t.Fatalf("冲突版本对不符: expected=%d current=%d", ce.Expected, ce.Current)
	}
	// 现值不损
	obj, _ := r.Get(ctx, "Agent", "global", "a")
	if obj.ResourceVersion != 2 || strings.Contains(string(obj.Content), "stale") {
		t.Fatalf("冲突写污染了现值: %+v", obj)
	}
	// 盲写（expected=0 撞不同内容）→ AlreadyExists
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Name: "a", Scope: "global",
		Content: agentYAML("a", "\n  blind: true")}, "agent:z", ""); !errors.Is(err, ErrAlreadyExists) {
		t.Fatalf("盲写应得 AlreadyExists, 得 %v", err)
	}
	// 幽灵更新（expected>0 但对象不在）→ Conflict(Current=0)
	_, _, err = r.Write(ctx, &Object{Kind: "Agent", Name: "ghost", Scope: "global", ResourceVersion: 7,
		Content: agentYAML("ghost", "")}, "fm", "")
	if !errors.As(err, &ce) || ce.Current != 0 {
		t.Fatalf("幽灵更新应得 Conflict(Current=0), 得 %v", err)
	}
	if n := r.commitCount(t); n != 2 {
		t.Fatalf("被拒写不应产生提交：提交数 = %d, 要 2", n)
	}
}

// 并发写串行化：多 goroutine 争写同一对象 + 各写各对象，零丢更新、提交数精确。
func TestConcurrentWriteSerialization(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	// 共享对象先建：rv=1
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Name: "shared", Scope: "global",
		Content: agentYAML("shared", "")}, "bootstrap", ""); err != nil {
		t.Fatal(err)
	}
	const writers = 8
	var wg sync.WaitGroup
	errs := make([]error, writers*2)
	// A：争写同一对象（读-改-写+冲突重试，最终一个不丢）
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			for attempt := 0; attempt < 200; attempt++ {
				cur, err := r.Get(ctx, "Agent", "global", "shared")
				if err != nil {
					errs[i] = err
					return
				}
				content := agentYAML("shared", fmt.Sprintf("\n  writer: w%d\n  gen: %d", i, attempt))
				rv, _, err := r.Write(ctx, &Object{Kind: "Agent", Name: "shared", Scope: "global",
					ResourceVersion: cur.ResourceVersion, Content: content}, fmt.Sprintf("agent:w%d", i), "")
				if err == nil {
					if rv != cur.ResourceVersion+1 {
						errs[i] = fmt.Errorf("rv 跳变: %d -> %d", cur.ResourceVersion, rv)
					}
					return
				}
				if !errors.As(err, new(*ConflictError)) {
					errs[i] = err
					return
				}
			}
			errs[i] = errors.New("重试预算耗尽")
		}(i)
	}
	// B：各写各对象（并发创建互不串扰）
	for i := 0; i < writers; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			name := fmt.Sprintf("own-%02d", i)
			_, _, err := r.Write(ctx, &Object{Kind: "Agent", Name: name, Scope: "global",
				Content: agentYAML(name, "")}, fmt.Sprintf("agent:w%d", i), "")
			errs[writers+i] = err
		}(i)
	}
	wg.Wait()
	for i, err := range errs {
		if err != nil {
			t.Fatalf("写者 %d: %v", i, err)
		}
	}
	// 串行化 ⇒ 无丢更新：共享对象 rv = 1 + writers，提交数 = 1 + writers + writers
	obj, err := r.Get(ctx, "Agent", "global", "shared")
	if err != nil {
		t.Fatal(err)
	}
	if obj.ResourceVersion != int64(writers+1) {
		t.Fatalf("共享对象 rv = %d, 要 %d（有丢更新）", obj.ResourceVersion, writers+1)
	}
	if n := r.commitCount(t); n != 1+writers*2 {
		t.Fatalf("提交数 = %d, 要 %d", n, 1+writers*2)
	}
	own, err := r.List(ctx, "Agent", "global")
	if err != nil {
		t.Fatal(err)
	}
	if len(own) != writers+1 {
		t.Fatalf("并发创建对象数 = %d, 要 %d", len(own), writers+1)
	}
}

// ───────────────────────── Watch ─────────────────────────

func TestWatchEmitsHeadChanges(t *testing.T) {
	r := newRepo(t)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	ch, err := r.Watch(ctx)
	if err != nil {
		t.Fatal(err)
	}
	// 空库无 HEAD：先写一个提交
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Name: "a", Scope: "global",
		Content: agentYAML("a", "")}, "fm", ""); err != nil {
		t.Fatal(err)
	}
	var head1 string
	select {
	case head1 = <-ch:
	case <-time.After(2 * time.Second):
		t.Fatal("未收到初始 HEAD")
	}
	if len(head1) != 40 {
		t.Fatalf("HEAD 非 sha: %q", head1)
	}
	// 新提交 → 新事件
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Name: "a", Scope: "global", ResourceVersion: 1,
		Content: agentYAML("a", "\n  v: 2")}, "fm", ""); err != nil {
		t.Fatal(err)
	}
	select {
	case head2 := <-ch:
		if head2 == head1 {
			t.Fatal("HEAD 未前移")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("变更未通知")
	}
	// 取消 → 频道关闭
	cancel()
	select {
	case _, ok := <-ch:
		if ok {
			t.Fatal("取消后不应再有事件")
		}
	case <-time.After(2 * time.Second):
		t.Fatal("取消后频道未关闭")
	}
}

// ───────────────────────── 文档保真 ─────────────────────────

func TestDocRoundTripPreservesCommentsAndMD(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	// yaml：注释与行内风格保留
	body := agentYAML("commented", "")
	rv, _, err := r.Write(ctx, &Object{Kind: "Agent", Name: "commented", Scope: "global", Content: body}, "fm", "")
	if err != nil {
		t.Fatal(err)
	}
	obj, _ := r.Get(ctx, "Agent", "global", "commented")
	if !strings.Contains(string(obj.Content), "# 注释保留探针") {
		t.Errorf("yaml 注释丢失: %s", obj.Content)
	}
	if !strings.Contains(string(obj.Content), "autoLevel: auto") {
		t.Errorf("spec 字段丢失: %s", obj.Content)
	}
	// md frontmatter：正文原样保留、frontmatter 注入 rv 与注释保留
	//（对象寻址由 kind 的 storage 模板决定，与内容形态 yaml/md 无关）。
	md := []byte("---\nname: impl\ndocType: ImplDoc # 类型注释\nscope: global\n---\n# 实现文档\n\n正文段落。\n")
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Name: "impl", Scope: "global", Content: md}, "fm", ""); err != nil {
		t.Fatalf("md 对象写入: %v", err)
	}
	mdObj, err := r.Get(ctx, "Agent", "global", "impl")
	if err != nil {
		t.Fatal(err)
	}
	if mdObj.ResourceVersion != 1 || !strings.Contains(string(mdObj.Content), "# 实现文档") ||
		!strings.Contains(string(mdObj.Content), "# 类型注释") || !strings.Contains(string(mdObj.Content), "resourceVersion: 1") {
		t.Errorf("md 保真不符: rv=%d content=\n%s", mdObj.ResourceVersion, mdObj.Content)
	}
	_ = rv
}

// ───────────────────────── List ─────────────────────────

func TestListAcrossScopes(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	seed := []Object{
		{Kind: "Workflow", Name: "w1", Scope: "global", Content: []byte("apiVersion: core/v1\nkind: Workflow\nmetadata: {name: w1, scope: global}\nspec: {}\n")},
		{Kind: "Workflow", Name: "w2", Scope: "developer-system", Content: []byte("apiVersion: core/v1\nkind: Workflow\nmetadata: {name: w2, scope: developer-system}\nspec: {}\n")},
		{Kind: "Agent", Name: "a1", Scope: "global", Content: agentYAML("a1", "")},
	}
	for i := range seed {
		o := seed[i]
		if _, _, err := r.Write(ctx, &o, "seed", ""); err != nil {
			t.Fatalf("seed %s: %v", o.Name, err)
		}
	}
	// scope="" 跨 scope 枚举 Workflow 两分支（排序：scope 字典序 developer-system < global）
	all, err := r.List(ctx, "Workflow", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 2 || all[0].Name != "w2" || all[1].Name != "w1" {
		t.Fatalf("跨 scope 清单不符: %+v %+v", all[0], all[1])
	}
	// scope 限定
	sys, err := r.List(ctx, "Workflow", "developer-system")
	if err != nil || len(sys) != 1 || sys[0].Name != "w2" {
		t.Fatalf("scope 限定清单不符: %+v err=%v", sys, err)
	}
	// 排序确定
	got, err := r.List(ctx, "Agent", "global")
	if err != nil || len(got) != 1 {
		t.Fatalf("Agent 清单: %+v err=%v", got, err)
	}
}

// TestWorkServiceGlobalBucket WorkService 单对象读面收口（v0.44 对账选点，adoption §1.15
// L44-01 的 Go 侧锚）：global 桶模板写入→Get 可寻址→枚举行 scope=global；系统 scope 无分支
// （旧根级无 scope 段模板的缺口形态=写读双向 400 不回归）。
func TestWorkServiceGlobalBucket(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	ws := &Object{Kind: "WorkService", Name: "notify", Scope: "global", Content: []byte(`apiVersion: core/v1
kind: WorkService
metadata:
  name: notify
  scope: global
spec:
  category: tool
  connector:
    types: [notify-email]
    impl: {kind: tool, ref: "ref://notify"}
    download: "tool://local"
    scheme: {ops: [notify], params: {}}
    llmPrompt: 发送通知邮件
`)}
	if _, _, err := r.Write(ctx, ws, "probe", ""); err != nil {
		t.Fatalf("Write WorkService: %v", err)
	}
	o, err := r.Get(ctx, "WorkService", "global", "notify")
	if err != nil {
		t.Fatalf("单对象 Get（global 桶）: %v", err)
	}
	if o.Name != "notify" || o.Scope != "global" {
		t.Fatalf("身份不符: %+v", o)
	}
	rows, err := r.List(ctx, "WorkService", "")
	if err != nil || len(rows) != 1 || rows[0].Scope != "global" {
		t.Fatalf("跨 scope 枚举: %+v err=%v", rows, err)
	}
	if _, err := r.Get(ctx, "WorkService", "notify", "notify"); !errors.Is(err, ErrScope) {
		t.Fatalf("系统 scope 应无可寻址分支, 得 %v", err)
	}
}

// ───────────────────────── 写守卫 ─────────────────────────

func TestWriteGuards(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	cases := []struct {
		name string
		o    *Object
		fm   string
		want error
	}{
		{"无主写入", &Object{Kind: "Agent", Name: "x", Scope: "global", Content: agentYAML("x", "")}, "", ErrNoFieldManager},
		{"未注册kind", &Object{Kind: "Nope", Name: "x", Scope: "global", Content: []byte("metadata: {name: x}\n")}, "fm", ErrNotRegistered},
		{"缺metadata", &Object{Kind: "Agent", Name: "x", Scope: "global", Content: []byte("spec: {}\n")}, "fm", ErrContent},
		{"身份不符", &Object{Kind: "Agent", Name: "x", Scope: "global", Content: agentYAML("y", "")}, "fm", ErrIdentity},
		{"scope不符", &Object{Kind: "Agent", Name: "x", Scope: "global", Content: []byte("metadata: {name: x, scope: systems}\n")}, "fm", ErrIdentity},
		{"空内容", &Object{Kind: "Agent", Name: "x", Scope: "global"}, "fm", ErrContent},
	}
	for _, c := range cases {
		if _, _, err := r.Write(ctx, c.o, c.fm, ""); !errors.Is(err, c.want) {
			t.Errorf("%s: 得 %v, 要 %v", c.name, err, c.want)
		}
	}
	// Get 未注册 / 不存在
	if _, err := r.Get(ctx, "Nope", "global", "x"); !errors.Is(err, ErrNotRegistered) {
		t.Errorf("Get 未注册: %v", err)
	}
	if _, err := r.Get(ctx, "Agent", "global", "absent"); !errors.Is(err, ErrNotFound) {
		t.Errorf("Get 缺席: %v", err)
	}
}

// ───────────────────────── ops 双写钩子 ─────────────────────────

type fakeOps struct {
	mu   sync.Mutex
	rows []opslog.ApplyLog
}

func (f *fakeOps) AppendApply(l opslog.ApplyLog) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rows = append(f.rows, l)
	return nil
}

func (f *fakeOps) AppendEvent(opslog.EventLog) error { return nil }
func (f *fakeOps) Reconcile() (int, error)           { return 0, nil }

func TestOpsDoubleWrite(t *testing.T) {
	dir := t.TempDir()
	store := &fakeOps{}
	r, err := NewGit(dir, mustScheme(t), &Options{Ops: store})
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	ctx := context.Background()
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Name: "a", Scope: "global",
		Content: agentYAML("a", "")}, "human:f", "plan:p"); err != nil {
		t.Fatal(err)
	}
	// 幂等 no-op 也留痕（Result=noop）
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Name: "a", Scope: "global", ResourceVersion: 1,
		Content: agentYAML("a", "")}, "human:f", ""); err != nil {
		t.Fatal(err)
	}
	store.mu.Lock()
	defer store.mu.Unlock()
	if len(store.rows) != 2 {
		t.Fatalf("apply_log 行数 = %d, 要 2: %+v", len(store.rows), store.rows)
	}
	if store.rows[0].Result != "ok" || store.rows[0].ResourceVersion != 1 || store.rows[0].ReasonRef != "plan:p" {
		t.Errorf("首行不符: %+v", store.rows[0])
	}
	if store.rows[1].Result != "noop" {
		t.Errorf("次行应 noop: %+v", store.rows[1])
	}
}

// Close 后拒收新写；Close 幂等。
func TestCloseSemantics(t *testing.T) {
	dir := t.TempDir()
	r, err := NewGit(dir, mustScheme(t), nil)
	if err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal(err)
	}
	if err := r.Close(); err != nil {
		t.Fatal("Close 应幂等")
	}
	_, _, err = r.Write(context.Background(), &Object{Kind: "Agent", Name: "a", Scope: "global",
		Content: agentYAML("a", "")}, "fm", "")
	if err == nil {
		t.Fatal("Close 后写入应失败")
	}
}

// git init 自动化：无 .git 的目录可直接打开。
func TestAutoGitInit(t *testing.T) {
	dir := t.TempDir()
	if _, err := os.Stat(filepath.Join(dir, ".git")); err == nil {
		t.Fatal("前置条件失败：目录不应已有 .git")
	}
	r, err := NewGit(dir, mustScheme(t), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer r.Close()
	if _, err := exec.Command("git", "-C", dir, "rev-parse", "--git-dir").Output(); err != nil {
		t.Fatalf("NewGit 未完成 git init: %v", err)
	}
}
