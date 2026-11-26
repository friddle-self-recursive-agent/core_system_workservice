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

// L87-01 认领归档域收窄锚：running 在册但本库归档无 running 判定（跨域=另一实例判定面，
// loop-87 18:00 活体实证：authority core-system.db 无 17:35 running 行、executor
// core-headless.db 有）→ 不认领租约起点、不进 TTL 回收路（不误判 failed）。
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
	_, leased := fresh.leases["self-recursive-core/cross-domain-probe"]
	adopted := fresh.adopted
	fresh.mu.Unlock()
	if !adopted {
		t.Fatalf("首拍应扫档（adopted=true）")
	}
	if leased {
		t.Fatalf("跨域 running（本库归档无 running 判定）不应认领租约——认领即 TTL 到点误判 failed（loop-87 活体缺口复现）")
	}
}
