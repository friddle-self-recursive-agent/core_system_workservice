// task_test.go —— TaskScheduler 纯逻辑机读锚（R-ENG-5 / runtime/task-scheduler.md §3）：
// 相位解析归一/依赖 DAG/写集冲突/四级上限/迁移合法性。判据面零 IO，全表驱动。
package scheduler

import (
	"testing"
	"time"
)

const taskDocEmptyStatus = `kind: Task
scope: self-recursive-core
name: t1
spec:
  title: 冒烟
  source: manual
  priority: high
  agent: agent://core-headless
  workspace: {mode: inplace}
  touches: {repos: [], paths: [], objects: []}
`

// TestParseEmptyStatusIsQueued 2026-11-26 活体缺口锚：新建 Task status 为空（P2：用户不写
// status）必须归一 queued 参选——曾因「空相位=非 Task 形」直接跳过导致新任务永不调度。
func TestParseEmptyStatusIsQueued(t *testing.T) {
	spec, st, ok := Parse([]byte(taskDocEmptyStatus))
	if !ok {
		t.Fatal("空 status 的 Task 应可解析（归一 queued），得 ok=false")
	}
	if st.Phase != PhaseQueued {
		t.Fatalf("空相位应归一 %q，得 %q", PhaseQueued, st.Phase)
	}
	if spec.Priority != "high" || spec.Agent != "agent://core-headless" {
		t.Fatalf("spec 投影错：%+v", spec)
	}
}

// TestParseExplicitPhaseWins 显式相位不被归一覆盖。
func TestParseExplicitPhaseWins(t *testing.T) {
	doc := taskDocEmptyStatus + `status:
  phase: running
  attempts: 2
`
	_, st, ok := Parse([]byte(doc))
	if !ok || st.Phase != PhaseRunning || st.Attempts != 2 {
		t.Fatalf("显式相位应保留，得 ok=%v %+v", ok, st)
	}
}

// TestPlanIndependentTasksRunConcurrently 无依赖+写集不相交 → 同拍双双拉起（并发上限内）。
func TestPlanIndependentTasksRunConcurrently(t *testing.T) {
	now := time.Now()
	cfg := ConfigDefaults()
	cfg.Enabled = true
	cfg.MaxConcurrent = 4
	cfg.PerService = 2
	cfg.PerRepo = 1
	tasks := []TaskView{
		{ID: "s/a", Scope: "s", Name: "a", Phase: PhaseReady, Touches: Touches{Repos: []string{"r1"}}},
		{ID: "s/b", Scope: "s", Name: "b", Phase: PhaseReady, Touches: Touches{Repos: []string{"r2"}}},
	}
	d := Plan(tasks, cfg, now)
	if !d["s/a"].Start || !d["s/b"].Start {
		t.Fatalf("无关任务应并行拉起，得 %+v", d)
	}
}

// TestPlanSameRepoSerializes perRepo=1：同仓两任务只放行一个，另一个保持 ready 等写集。
func TestPlanSameRepoSerializes(t *testing.T) {
	now := time.Now()
	cfg := ConfigDefaults()
	cfg.Enabled = true
	cfg.MaxConcurrent = 4
	cfg.PerService = 4
	cfg.PerRepo = 1
	tasks := []TaskView{
		{ID: "s/a", Scope: "s", Name: "a", Phase: PhaseReady, Touches: Touches{Repos: []string{"r1"}}},
		{ID: "s/b", Scope: "s", Name: "b", Phase: PhaseReady, Touches: Touches{Repos: []string{"r1"}}},
	}
	d := Plan(tasks, cfg, now)
	starts := 0
	for _, dec := range d {
		if dec.Start {
			starts++
		}
	}
	if starts != 1 {
		t.Fatalf("同仓 perRepo=1 应只拉起 1 个，得 %d（%+v）", starts, d)
	}
}

// TestPlanDependsOnBlocks 未完成依赖 → blocked（原因可见）。
func TestPlanDependsOnBlocks(t *testing.T) {
	now := time.Now()
	cfg := ConfigDefaults()
	cfg.Enabled = true
	tasks := []TaskView{
		{ID: "s/a", Scope: "s", Name: "a", Phase: PhaseRunning},
		{ID: "s/b", Scope: "s", Name: "b", Phase: PhaseQueued, DependsOn: []string{"s/a"}},
	}
	d := Plan(tasks, cfg, now)
	if dec, ok := d["s/b"]; !ok || dec.Phase != PhaseBlocked {
		t.Fatalf("依赖未完成应 blocked，得 %+v", d["s/b"])
	}
}

// TestPlanDependsOnReadyUnblocks 依赖 done → ready 且可拉起。
func TestPlanDependsOnReadyUnblocks(t *testing.T) {
	now := time.Now()
	cfg := ConfigDefaults()
	cfg.Enabled = true
	tasks := []TaskView{
		{ID: "s/a", Scope: "s", Name: "a", Phase: PhaseDone},
		{ID: "s/b", Scope: "s", Name: "b", Phase: PhaseBlocked, DependsOn: []string{"s/a"}},
	}
	d := Plan(tasks, cfg, now)
	dec := d["s/b"]
	if dec.Phase != PhaseReady || !dec.Start {
		t.Fatalf("依赖 done 应 ready+拉起，得 %+v", dec)
	}
}

// TestPlanExclusiveMutualExclusion touches.exclusive 同值互斥。
func TestPlanExclusiveMutualExclusion(t *testing.T) {
	now := time.Now()
	cfg := ConfigDefaults()
	cfg.Enabled = true
	cfg.PerRepo = 4
	cfg.PerService = 4
	tasks := []TaskView{
		{ID: "s/a", Scope: "s", Name: "a", Phase: PhaseReady, Touches: Touches{Exclusive: []string{"deploy/prod"}}},
		{ID: "s/b", Scope: "s", Name: "b", Phase: PhaseReady, Touches: Touches{Exclusive: []string{"deploy/prod"}}},
	}
	d := Plan(tasks, cfg, now)
	starts := 0
	for _, dec := range d {
		if dec.Start {
			starts++
		}
	}
	if starts != 1 {
		t.Fatalf("exclusive 同值应互斥只放 1 个，得 %d", starts)
	}
}

// TestCanMigrateClosedSet 相位迁移闭合集（野词/终态出边一律拒）。
func TestCanMigrateClosedSet(t *testing.T) {
	cases := []struct {
		from, to string
		want     bool
	}{
		{PhaseQueued, PhaseReady, true},
		{PhaseQueued, PhaseBlocked, true},
		{PhaseReady, PhaseRunning, true},
		{PhaseRunning, PhaseDone, true},
		{PhaseRunning, PhaseVerifying, true},
		{PhaseDone, PhaseReady, false},   // 终态零出边
		{PhaseCanceled, PhaseReady, false},
		{"bogus", PhaseReady, false},     // 闭合集外
		{PhaseReady, PhaseQueued, false}, // 回退非法
	}
	for _, c := range cases {
		if got := CanMigrate(c.from, c.to); got != c.want {
			t.Errorf("CanMigrate(%s→%s)=%v 期望 %v", c.from, c.to, got, c.want)
		}
	}
}
