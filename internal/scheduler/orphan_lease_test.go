// orphan_lease_test.go —— 孤儿租约认领回归锚（L84-01）：
// 场景=loop-82/83 活体实证的缺口：cored 重启后进程内 leases 清零，仓库里 running 任务成
// 孤儿（attempts 恒 0、lastAction 恒「拉起中（spawn）」、永不回收也永不重拉）。锚=重启后的
// 新调度器实例首拍把归档里的 running 判定认领为租约起点，TTL 到点走既有 ready 回收路。
package scheduler

import (
	"context"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

func TestAdoptOrphanLeases_RecoverRunningAfterRestart(t *testing.T) {
	coreSchemePathOnce := func(t *testing.T) *scheme.Scheme {
		t.Helper()
		sch, err := scheme.LoadFile(coreSchemePath())
		if err != nil {
			t.Fatalf("装载 scheme: %v", err)
		}
		return sch
	}
	sch := coreSchemePathOnce(t)
	repo, err := repository.OpenSQLite(t.TempDir()+"/docs.db", sch, nil, "")
	if err != nil {
		t.Fatalf("打开 sqlite: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	ops, err := opslog.OpenSQLite(t.TempDir() + "/ops.db")
	if err != nil {
		t.Fatalf("打开 ops DB: %v", err)
	}
	t.Cleanup(func() { _ = ops.Close() })
	ctx := context.Background()

	// 重启前的旧进程：拉起任务 → status=running 写面 + running 事件归档（可回查的审计事实）。
	old := New(repo, eventbus.NewBus(&eventbus.Options{Ops: ops}), sch, ops, ops, Options{Sessions: ops})
	writeTaskObject(t, ctx, repo, "self-recursive-core", "orphan-probe", "queued", "")
	obj, err := repo.Get(ctx, "Task", "self-recursive-core", "orphan-probe")
	if err != nil {
		t.Fatalf("读任务: %v", err)
	}
	if err := old.writeStatus(ctx, obj, statusPatch{Phase: PhaseRunning}, "test:spawn"); err != nil {
		t.Fatalf("running 写面: %v", err)
	}
	old.publishState("self-recursive-core/orphan-probe", PhaseRunning)
	time.Sleep(20 * time.Millisecond) // 归档落库（异步 append）

	// 重启后的新进程：进程内租约清零（leases 空）——首拍应从归档认领。
	fresh := New(repo, eventbus.NewBus(&eventbus.Options{Ops: ops}), sch, ops, ops, Options{Sessions: ops})
	if len(fresh.leases) != 0 {
		t.Fatalf("新进程租约应清零（重启口径），实得 %d", len(fresh.leases))
	}
	fresh.Tick(ctx)
	fresh.mu.Lock()
	start, leased := fresh.leases["self-recursive-core/orphan-probe"]
	adopted := fresh.adopted
	fresh.mu.Unlock()
	if !adopted {
		t.Fatalf("首拍应扫档认领（adopted=true）")
	}
	if !leased {
		t.Fatalf("孤儿 running 未被认领为租约（缺口复现）")
	}
	if start.IsZero() {
		t.Fatalf("租约起点应取归档 running 判定时刻")
	}
	// TTL 到点（now-认领起点 > leaseTtl）→ recoverLeases 走既有 ready 回收路（attempts+1）。
	cfg := ConfigOf(sch)
	ttl := cfg.LeaseTTL
	late := New(repo, eventbus.NewBus(&eventbus.Options{Ops: ops}), sch, ops, ops,
		Options{Sessions: ops, Now: func() time.Time { return start.Add(2 * ttl) }})
	late.Tick(ctx)
	_, st, ok := Parse(mustGet(t, ctx, repo, "self-recursive-core", "orphan-probe"))
	if !ok {
		t.Fatalf("读回不可解析")
	}
	if st.Phase != PhaseReady {
		t.Fatalf("孤儿 running 应被回收为 ready（缺口修复），实得 %s", st.Phase)
	}
	if st.Attempts != 1 {
		t.Fatalf("回收应 attempts+1，实得 %d", st.Attempts)
	}
	if _, leased := late.leases["self-recursive-core/orphan-probe"]; leased {
		t.Fatalf("回收后租约应删除")
	}
}

// obj2 读回任务对象（测试内小帮手——避免与既有帮手名冲突）。
func obj2(t *testing.T, ctx context.Context, repo *repository.SQLiteRepo, scope, name string) *repository.Object {
	t.Helper()
	o, err := repo.Get(ctx, "Task", scope, name)
	if err != nil {
		t.Fatalf("读回 %s/%s: %v", scope, name, err)
	}
	return o
}

func mustGet(t *testing.T, ctx context.Context, repo *repository.SQLiteRepo, scope, name string) []byte {
	t.Helper()
	o := obj2(t, ctx, repo, scope, name)
	return o.Content
}

// L87-01 + 规则12（2026-11-28）：跨域 running（本库归档无 running 判定=另一实例判定面）
// 由 watchLeases 的 rv 真相源认领——租约起点=首见 now（全 TTL 观察窗，不冒充历史起点；
// loop-87 的误判 failed 根因=用陈旧归档 ts 当租约起点，now 锚已消除）。语义升级锚：
// authority 对他 core 拉起的 running 不再「保守悬挂」，rv 停更超 TTL 照常回收（issue #1）。
func TestAdoptOrphanLeases_CrossDomainRunningNotAdopted(t *testing.T) {
	sch := func(t *testing.T) *scheme.Scheme {
		t.Helper()
		sch, err := scheme.LoadFile(coreSchemePath())
		if err != nil {
			t.Fatalf("装载 scheme: %v", err)
		}
		return sch
	}(t)
	repo, err := repository.OpenSQLite(t.TempDir()+"/docs.db", sch, nil, "")
	if err != nil {
		t.Fatalf("打开 sqlite: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	// 认领方 ops 库（模拟 authority）：空档——无该任务的 running 判定行。
	ops, err := opslog.OpenSQLite(t.TempDir() + "/ops-authority.db")
	if err != nil {
		t.Fatalf("打开 ops DB: %v", err)
	}
	t.Cleanup(func() { _ = ops.Close() })
	ctx := context.Background()
	writeTaskObject(t, ctx, repo, "self-recursive-core", "cross-domain-probe", "running", "")
	fresh := New(repo, eventbus.NewBus(&eventbus.Options{Ops: ops}), sch, ops, ops, Options{Sessions: ops})
	fresh.Tick(ctx)
	fresh.mu.Lock()
	at, leased := fresh.leases["self-recursive-core/cross-domain-probe"]
	adopted := fresh.adopted
	fresh.mu.Unlock()
	if !adopted {
		t.Fatalf("首拍应扫档（adopted=true）")
	}
	if !leased {
		t.Fatalf("跨域 running 应由 watchLeases 认领（规则12 账目真相源——不再保守悬挂，issue #1）")
	}
	// now 锚+TTL 半窗内=不回收（loop-87 误判 failed 不复现：年轻任务不受陈旧起点拖累）。
	views := []TaskView{{ID: "self-recursive-core/cross-domain-probe", Phase: PhaseRunning}}
	half := ConfigDefaults().LeaseTTL / 2
	d := Recover(views[0], at, ConfigDefaults(), time.Now().Add(half))
	if d.Phase != "" {
		t.Fatalf("TTL 半窗内不应回收（loop-87 误判复现）：got %+v", d)
	}
}

// 规则12 活性锚（2026-11-28）：rv 推进=租约续期（长跑任务不被误回收）；
// rv 停更超 leaseTtl=悬挂回收 ready（attempts+1 重拉——issue #1 悬挂病灶收口）。
func TestWatchLeases_RVProgressRefreshAndStaleRecover(t *testing.T) {
	sch, err := scheme.LoadFile(coreSchemePath())
	if err != nil {
		t.Fatalf("装载 scheme: %v", err)
	}
	repo, err := repository.OpenSQLite(t.TempDir()+"/docs.db", sch, nil, "")
	if err != nil {
		t.Fatalf("打开 sqlite: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	ops, err := opslog.OpenSQLite(t.TempDir() + "/ops.db")
	if err != nil {
		t.Fatalf("打开 ops DB: %v", err)
	}
	t.Cleanup(func() { _ = ops.Close() })
	ctx := context.Background()
	base := time.Unix(1795860000, 0)
	clock := base
	fresh := New(repo, eventbus.NewBus(&eventbus.Options{Ops: ops}), sch, ops, ops,
		Options{Sessions: ops, Now: func() time.Time { return clock }})
	writeTaskObject(t, ctx, repo, "self-recursive-core", "rv-probe", "running", "")
	fresh.Tick(ctx) // 首见认领：lease=base
	// rv 推进（模拟 runner 心跳写面）+ 时钟过 2/3 TTL → 续期后不回收
	writeTaskObject(t, ctx, repo, "self-recursive-core", "rv-probe", "running", "")
	clock = base.Add(2 * ConfigDefaults().LeaseTTL / 3)
	fresh.Tick(ctx)
	if ph, _ := phaseOf(t, ctx, repo, "self-recursive-core", "rv-probe"); ph != PhaseRunning {
		t.Fatalf("rv 推进+TTL 2/3 应续期保 running，got %s", ph)
	}
	// rv 停更（最后一次写面在 2/3 TTL 处）+ 时钟过「上次活性+TTL」→ 回收 ready
	clock = base.Add(2*ConfigDefaults().LeaseTTL/3 + ConfigDefaults().LeaseTTL + time.Minute)
	fresh.Tick(ctx)
	if ph, _ := phaseOf(t, ctx, repo, "self-recursive-core", "rv-probe"); ph != PhaseReady {
		t.Fatalf("rv 停更超 TTL 应回收 ready（规则12 重拉），got %s", ph)
	}
}
