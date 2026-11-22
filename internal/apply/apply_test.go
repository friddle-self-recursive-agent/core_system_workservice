// apply_test.go —— apply 管线最小闭环验收（loop-3 起点①，concepts/apply.md）：
// 野对象/无主写入拒绝、DocApplied 落 event log、git 提交数 vs apply_log 行数对账一致、
// Conflict 后 recover 路径、dry-run 零副作用。
package apply

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

const testScheme = `apiVersion: core/v1
kind: Scheme
metadata: {name: test, scope: builtin}
spec:
  version: core/v1
  kinds:
    - {kind: Scheme, builtin: true, storage: core/scheme.yaml}
    - {kind: Agent, builtin: true, storage: global/agents/<name>/agent.yaml}
  eventTypes:
    - {type: DocApplied, source: core, defaultLevel: info, route: owner-workflow}
`

func agentYAML(name, extra string) []byte {
	return []byte(`apiVersion: core/v1
kind: Agent
metadata: {name: ` + name + `, scope: global}
spec: {autoLevel: all` + extra + `}
`)
}

// newWorld 组装真实依赖链：git 后端 + sqlite ops DB + 事件总线 + 写引擎（本地 admin 主体）。
type world struct {
	repo   *repository.GitRepo
	ops    *opslog.SQLiteStore
	bus    *eventbus.Bus
	engine *Engine
}

func newWorld(t *testing.T) *world {
	t.Helper()
	sch, err := scheme.Load([]byte(testScheme))
	if err != nil {
		t.Fatalf("载入测试 scheme: %v", err)
	}
	dir := t.TempDir()
	ops, err := opslog.OpenSQLite(filepath.Join(dir, "ops.db"))
	if err != nil {
		t.Fatalf("ops DB: %v", err)
	}
	t.Cleanup(func() { _ = ops.Close() })
	repo, err := repository.NewGit(dir, sch, &repository.Options{Ops: ops, WatchInterval: 10 * time.Millisecond})
	if err != nil {
		t.Fatalf("git 后端: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	bus := eventbus.NewBus(&eventbus.Options{Ops: ops, DedupWindow: 50 * time.Millisecond})
	w := &world{
		repo: repo, ops: ops, bus: bus,
		engine: &Engine{
			Sch: sch, Repo: repo, Bus: bus, Ops: ops,
			Perm:    &permission.System{},
			Subject: permission.Subject{User: "agent:test", Roles: []string{permission.DefaultAdmin}},
		},
	}
	return w
}

func (w *world) commits(t *testing.T) int {
	t.Helper()
	cmd := exec.Command("git", "rev-list", "--count", "HEAD")
	cmd.Dir = w.repo.Root()
	out, err := cmd.Output()
	if err != nil {
		// 空库（unborn HEAD）=0 提交；其余错误照实报。
		if ee, ok := err.(*exec.ExitError); ok && ee.ExitCode() == 128 {
			return 0
		}
		t.Fatalf("rev-list: %v", err)
	}
	var n int
	if _, err := fmt.Sscanf(strings.TrimSpace(string(out)), "%d", &n); err != nil {
		t.Fatalf("提交数解析 %q: %v", out, err)
	}
	return n
}

func (w *world) okRows(t *testing.T) int {
	t.Helper()
	n, err := w.ops.CountApply("ok")
	if err != nil {
		t.Fatalf("统计 ok 行: %v", err)
	}
	return n
}

// TestApplyRejectsWildKind 野对象拒绝：kind 未注册 → ErrNotRegistered + denied 审计行 + 零提交。
func TestApplyRejectsWildKind(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	_, err := w.engine.Apply(ctx, &repository.Object{
		Kind: "Unicorn", Name: "rainbow", Scope: "global", Content: []byte("kind: Unicorn\n"),
	}, Options{FieldManager: "agent:x"})
	if !errors.Is(err, ErrNotRegistered) {
		t.Fatalf("野对象应被拒（ErrNotRegistered），得 %v", err)
	}
	if n, _ := w.ops.CountApply("denied"); n != 1 {
		t.Fatalf("denied 审计行 = %d, 要 1", n)
	}
	if w.commits(t) != 0 {
		t.Fatal("被拒写入不得产生提交")
	}
	// dry-run 同样过闸门（准入结论与真执行一致）。
	dr, err := w.engine.Apply(ctx, &repository.Object{
		Kind: "Unicorn", Name: "rainbow", Scope: "global", Content: []byte("kind: Unicorn\n"),
	}, Options{FieldManager: "agent:x", DryRun: true})
	if !errors.Is(err, ErrNotRegistered) {
		t.Fatalf("dry-run 野对象也应被拒，得 %v", err)
	}
	if dr != nil {
		t.Fatal("拒绝路径不得返回 Result")
	}
}

// TestApplyRejectsMasterless 无主写入拒绝：fieldManager 缺失（apply.md rule 2）；
// 权限默认拒绝：主体无任何 role → ErrDenied（denied 审计行）。
func TestApplyRejectsMasterless(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	_, err := w.engine.Apply(ctx, &repository.Object{
		Kind: "Agent", Name: "a", Scope: "global", Content: agentYAML("a", ""),
	}, Options{})
	if !errors.Is(err, ErrNoFieldManager) {
		t.Fatalf("无主写入应被拒，得 %v", err)
	}
	// 键不全。
	if _, err := w.engine.Apply(ctx, &repository.Object{Kind: "Agent", Name: "a"}, Options{FieldManager: "fm"}); !errors.Is(err, ErrObject) {
		t.Fatalf("键不全应被拒（ErrObject），得 %v", err)
	}
	// 权限系统默认拒绝（主体无 role）。
	denied := &Engine{Sch: w.engine.Sch, Repo: w.repo, Bus: w.bus, Ops: w.ops, Perm: &permission.System{}}
	_, err = denied.Apply(ctx, &repository.Object{
		Kind: "Agent", Name: "a", Scope: "global", Content: agentYAML("a", ""),
	}, Options{FieldManager: "agent:stranger"})
	if !errors.Is(err, ErrDenied) {
		t.Fatalf("无 role 主体应被拒（ErrDenied），得 %v", err)
	}
	if n, _ := w.ops.CountApply("denied"); n != 1 {
		t.Fatalf("denied 审计行 = %d, 要 1（权限拒绝落行；无主/键不全在可审计身份前拒绝，不落行）", n)
	}
	if w.commits(t) != 0 {
		t.Fatal("被拒写入不得产生提交")
	}
}

// TestApplyDocAppliedEventLog DocApplied 必发（订阅方收到 + event_log 归档行）；
// 幂等重入 noop 不发事件（防自我唤醒）；库内对账不变量 diff=0。
func TestApplyDocAppliedEventLog(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	events, cancel := w.bus.Subscribe("DocApplied")
	defer cancel()

	res, err := w.engine.Apply(ctx, &repository.Object{
		Kind: "Agent", Name: "core-headless", Scope: "global", Content: agentYAML("core-headless", ""),
	}, Options{FieldManager: "agent:bootstrap", ReasonRef: "plan:loop-3"})
	if err != nil {
		t.Fatal(err)
	}
	if res.Action != "create" || res.RV != 1 || res.Event == nil {
		t.Fatalf("创建结果不符: %+v", res)
	}

	select {
	case evt := <-events:
		if evt.Type != "DocApplied" || evt.Involved != "Agent/global/core-headless" || evt.Level != "info" {
			t.Fatalf("DocApplied 载荷不符: %+v", evt)
		}
		if evt.ID == "" || evt.ID != res.Event.ID {
			t.Fatalf("回执 ID 应与投递事件一致且非空: %+v vs %+v", evt, res.Event)
		}
	case <-time.After(time.Second):
		t.Fatal("DocApplied 未送达订阅方")
	}
	if n, err := w.ops.CountEvent("DocApplied"); err != nil || n != 1 {
		t.Fatalf("event_log DocApplied 行 = %d err=%v, 要 1", n, err)
	}

	// 幂等重入：同内容 → noop，无事件、无新提交。
	res2, err := w.engine.Apply(ctx, &repository.Object{
		Kind: "Agent", Name: "core-headless", Scope: "global", ResourceVersion: 1,
		Content: agentYAML("core-headless", ""),
	}, Options{FieldManager: "agent:bootstrap"})
	if err != nil || res2.Action != "noop" || res2.Event != nil {
		t.Fatalf("幂等重入应 noop 无事件: %+v err=%v", res2, err)
	}
	select {
	case evt := <-events:
		t.Fatalf("noop 不得发事件，收到 %+v", evt)
	case <-time.After(120 * time.Millisecond):
	}
	if n, _ := w.ops.CountEvent("DocApplied"); n != 1 {
		t.Fatalf("noop 后 event_log 行 = %d, 要仍 1", n)
	}

	// 库内不变量：ok 行 ↔ DocApplied 归档 一一对应。
	diff, err := w.ops.Reconcile()
	if err != nil || diff != 0 {
		t.Fatalf("Reconcile diff=%d err=%v, 要 0", diff, err)
	}
}

// TestApplyReconcileGitVsOps 对账一致性（ops-log.md §3 的单测级）：
// git 提交数 == apply_log(ok) 行数（denied/conflict/noop 行不对应提交）。
func TestApplyReconcileGitVsOps(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	fm := Options{FieldManager: "agent:bootstrap"}

	if _, err := w.engine.Apply(ctx, &repository.Object{Kind: "Agent", Name: "a", Scope: "global",
		Content: agentYAML("a", "")}, fm); err != nil {
		t.Fatal(err)
	}
	if _, err := w.engine.Apply(ctx, &repository.Object{Kind: "Agent", Name: "a", Scope: "global",
		ResourceVersion: 1, Content: agentYAML("a", ",\n    tier: edge")}, fm); err != nil {
		t.Fatal(err)
	}
	if _, err := w.engine.Apply(ctx, &repository.Object{Kind: "Agent", Name: "a", Scope: "global",
		ResourceVersion: 2, Content: agentYAML("a", ",\n    tier: edge")}, fm); err != nil {
		t.Fatal(err)
	} // noop
	_, _ = w.engine.Apply(ctx, &repository.Object{Kind: "Ghost", Name: "g", Scope: "global",
		Content: []byte("x")}, fm) // denied
	if _, err := w.engine.Apply(ctx, &repository.Object{Kind: "Agent", Name: "b", Scope: "global",
		Content: agentYAML("b", "")}, fm); err != nil {
		t.Fatal(err)
	}

	commits := w.commits(t)
	ok := w.okRows(t)
	if commits != 3 || ok != 3 {
		t.Fatalf("对账不一致: git 提交=%d apply_log(ok)=%d, 要相等（3/3）", commits, ok)
	}
	if noop, _ := w.ops.CountApply("noop"); noop != 1 {
		t.Fatalf("noop 行 = %d, 要 1", noop)
	}
	if denied, _ := w.ops.CountApply("denied"); denied != 1 {
		t.Fatalf("denied 行 = %d, 要 1", denied)
	}
	if diff, err := w.ops.Reconcile(); err != nil || diff != 0 {
		t.Fatalf("Reconcile diff=%d err=%v, 要 0", diff, err)
	}
}

// TestApplyConflictRecover Conflict 后 recover 路径（apply.md §3：后写者重读再提交）：
// 冲突落 conflict 审计行（含现库版本）→ 重读 → 携新版本重提 → ok；现值全程不损。
func TestApplyConflictRecover(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	fm := Options{FieldManager: "agent:bootstrap"}

	if _, err := w.engine.Apply(ctx, &repository.Object{Kind: "Agent", Name: "a", Scope: "global",
		Content: agentYAML("a", "")}, fm); err != nil {
		t.Fatal(err)
	}
	v2 := agentYAML("a", ",\n    gen: 2")
	if _, err := w.engine.Apply(ctx, &repository.Object{Kind: "Agent", Name: "a", Scope: "global",
		ResourceVersion: 1, Content: v2}, fm); err != nil {
		t.Fatal(err)
	}
	// 陈旧写：expected=1（现库 2）→ Conflict。
	stale := agentYAML("a", ",\n    stale: true")
	_, err := w.engine.Apply(ctx, &repository.Object{Kind: "Agent", Name: "a", Scope: "global",
		ResourceVersion: 1, Content: stale}, Options{FieldManager: "agent:late"})
	var ce *repository.ConflictError
	if !errors.As(err, &ce) || ce.Current != 2 {
		t.Fatalf("陈旧写应 Conflict(Current=2)，得 %v", err)
	}
	if n, _ := w.ops.CountApply("conflict"); n != 1 {
		t.Fatalf("conflict 审计行 = %d, 要 1", n)
	}
	// 现值不损。
	cur, err := w.repo.Get(ctx, "Agent", "global", "a")
	if err != nil || cur.ResourceVersion != 2 || strings.Contains(string(cur.Content), "stale") {
		t.Fatalf("冲突后现值不损: rv=%d err=%v content=%s", cur.ResourceVersion, err, cur.Content)
	}
	// recover：重读后携现版本重提（写者自己的意图内容，与现库不同）→ 成功，rv 2→3。
	late := agentYAML("a", ",\n    late: true")
	res, err := w.engine.Apply(ctx, &repository.Object{Kind: "Agent", Name: "a", Scope: "global",
		ResourceVersion: cur.ResourceVersion, Content: late}, Options{FieldManager: "agent:late"})
	if err != nil || res.Action != "update" || res.RV != 3 || res.Event == nil {
		t.Fatalf("recover 提交不符: %+v err=%v", res, err)
	}
	// 对账仍一致（ok=3 ↔ 提交=3；conflict/noop 行不产生提交）。
	if c := w.commits(t); c != w.okRows(t) {
		t.Fatalf("recover 后对账不一致: 提交=%d ok=%d", c, w.okRows(t))
	}
}

// TestApplyDryRunZeroSideEffect dry-run：准入结论与真执行一致，零副作用（无提交/无审计行/无事件）。
func TestApplyDryRunZeroSideEffect(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	fm := Options{FieldManager: "agent:bootstrap"}

	// 不存在 → 将创建（RV=0）。
	res, err := w.engine.Apply(ctx, &repository.Object{Kind: "Agent", Name: "a", Scope: "global",
		Content: agentYAML("a", "")}, Options{FieldManager: fm.FieldManager, DryRun: true})
	if err != nil || res.Action != "dry-run" || res.RV != 0 || res.Event != nil {
		t.Fatalf("dry-run（缺席）不符: %+v err=%v", res, err)
	}
	// 落库后 → dry-run 返回现版本。
	if _, err := w.engine.Apply(ctx, &repository.Object{Kind: "Agent", Name: "a", Scope: "global",
		Content: agentYAML("a", "")}, fm); err != nil {
		t.Fatal(err)
	}
	res, err = w.engine.Apply(ctx, &repository.Object{Kind: "Agent", Name: "a", Scope: "global",
		ResourceVersion: 1, Content: agentYAML("a", ",\n    next: true")},
		Options{FieldManager: fm.FieldManager, DryRun: true})
	if err != nil || res.Action != "dry-run" || res.RV != 1 || res.Event != nil {
		t.Fatalf("dry-run（在场）不符: %+v err=%v", res, err)
	}
	// 零副作用核对。
	if w.commits(t) != 1 {
		t.Fatalf("dry-run 不得产生提交：=%d", w.commits(t))
	}
	if n, _ := w.ops.CountApply("ok"); n != 1 {
		t.Fatalf("dry-run 不得落 ok 行：=%d", n)
	}
	if n, _ := w.ops.CountApply("dry-run"); n != 0 {
		t.Fatalf("dry-run 是零副作用，不得落审计行：=%d", n)
	}
}

// TestApplyRequestLevelSubject 请求级主体（v0.9，permission.md §3）：
// Options.Subject 非 nil=按请求判定（引擎 fallback 不兜底空主体）；授权面=Role/Group 对象装载；
// denied 审计行携带主体标识；Options.Subject=nil → 引擎缺省 fallback（进程级本地信任）。
func TestApplyRequestLevelSubject(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	perm := &permission.System{}
	if err := perm.LoadObjects([]permission.ObjectDoc{
		{Kind: "Role", Name: "writer", Scope: "global", Content: []byte(
			`kind: Role
metadata: {name: writer}
spec:
  permissions:
  - {scope: "global/*", actions: [write]}
`)},
		{Kind: "Group", Name: "devs", Scope: "global", Content: []byte(
			`kind: Group
metadata: {name: devs}
spec:
  roles: [role/writer]
`)},
	}); err != nil {
		t.Fatal(err)
	}
	w.engine.Perm = perm
	obj := func() *repository.Object {
		return &repository.Object{Kind: "Agent", Name: "doc", Scope: "global", Content: agentYAML("doc", "")}
	}

	// ① 请求级空主体 → 拒绝（本地信任 admin 不兜底）；denied 行携带主体标识。
	_, err := w.engine.Apply(ctx, obj(), Options{
		FieldManager: "agent:a", Subject: &permission.Subject{User: "agent:none"},
	})
	if !errors.Is(err, ErrDenied) {
		t.Fatalf("请求级空主体应拒绝，得 %v", err)
	}
	rows, err := w.ops.QueryApply("denied")
	if err != nil || len(rows) != 1 || rows[0].Subject != "agent:none" {
		t.Fatalf("denied 行应携带主体 agent:none: %+v err=%v", rows, err)
	}
	if !strings.Contains(rows[0].PolicyDecisions, "subject=agent:none") {
		t.Fatalf("denied 行判定摘要应留痕主体: %q", rows[0].PolicyDecisions)
	}

	// ② 持 role 主体（断言对上装载的 Role 对象）→ 放行 create。
	res, err := w.engine.Apply(ctx, obj(), Options{
		FieldManager: "agent:b", Subject: &permission.Subject{User: "agent:dev", Roles: []string{"role/writer"}},
	})
	if err != nil || res.Action != "create" || res.RV != 1 {
		t.Fatalf("持 role 主体应放行: %+v err=%v", res, err)
	}

	// ③ group→role 链主体放行（user→group→role，同内容幂等 noop；装载为一次性快照投影）。
	res, err = w.engine.Apply(ctx, obj(), Options{
		FieldManager: "agent:c", Subject: &permission.Subject{User: "agent:g", Groups: []string{"group/devs"}},
	})
	if err != nil || res.Action != "noop" {
		t.Fatalf("group 链主体应放行: %+v err=%v", res, err)
	}

	// ④ Options.Subject=nil → 引擎缺省 fallback（cored 本地信任=DefaultAdmin）。
	res, err = w.engine.Apply(ctx, obj(), Options{FieldManager: "agent:d"})
	if err != nil || res.Action != "noop" {
		t.Fatalf("fallback 主体应放行: %+v err=%v", res, err)
	}
}
