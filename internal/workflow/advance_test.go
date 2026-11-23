// advance_test.go —— 相位推进器异步推进首段验收（v0.26——装配层定时驱动，workflow.md 规则 15）：
// 单步查表推进（stepByTable 复用——判定器不动）/ 落闸锚挂起 / 落终点锚终态 / 序列尾零动作停
// （不冒充执行完毕）/ 零动作防御面（triggered/suspended/failed）/ 多 run 独立推进 /
// Advancer 装配范式（Enabled/Interval/关断零推进/定时驱动）/ advances 计数可查 / 并发安全。
package workflow

import (
	"context"
	"sync"
	"testing"
	"time"
)

// seedAdvance 种一条 run 到指定相位/状态（环内直写——白盒构造任意迁移判定前置态）。
func seedAdvance(t *testing.T, runs *RunLog, route, wf string, phase Phase, status string) {
	t.Helper()
	runs.StartRun(TriggerRecord{
		Seq: 1, EventID: "ev-seed", Type: "DocApplied", DedupKey: "apply:seed:1",
		ReceivedAt: time.Now(), Route: route, Owner: "global",
	}, ExecutorBinding{Executor: ExecWorkflow, Target: wf})
	runs.mu.Lock()
	// 受理首步已走一拍（Idle→Analyzing）；测试前置态覆盖其上。
	runs.ring[len(runs.ring)-1].Phase = string(phase)
	runs.ring[len(runs.ring)-1].Status = status
	runs.mu.Unlock()
}

// runAt 取第 n 条 run 的投影（断言辅助）。
func runAt(runs *RunLog, n int) RunRecord { return runs.State().Runs[n] }

// TestAdvanceSingleStep 序列尾前单步推进：查表命中 advance 行→前进一步（status 保持 running）；
// 序列尾（watch-only Reporting 无 advance 出行）零动作停——status 保持 running
//（诚实投影不冒充执行完毕，「run 出现≠对账执行完毕」口径）。
func TestAdvanceSingleStep(t *testing.T) {
	runs := NewRunLog(nil)
	seedAdvance(t, runs, "workflows/watch-only", "watch-only", Analyzing, RunRunning)
	if n := runs.Advance(); n != 1 {
		t.Fatalf("单步推进应推进 1 条 run，得 %d", n)
	}
	if r := runAt(runs, 0); r.Phase != string(Reporting) || r.Status != RunRunning {
		t.Fatalf("推进后应 {Reporting, running}: {phase=%s, status=%s}", r.Phase, r.Status)
	}
	// 序列尾：Reporting 无 advance 出行——零动作停（不冒充执行完毕，无 done 词形）。
	if n := runs.Advance(); n != 0 {
		t.Fatalf("序列尾推进应零动作，得 %d", n)
	}
	if r := runAt(runs, 0); r.Phase != string(Reporting) || r.Status != RunRunning {
		t.Fatalf("序列尾应原样保留 {Reporting, running}: {phase=%s, status=%s}", r.Phase, r.Status)
	}
	if st := runs.State(); st.Advances != 1 {
		t.Fatalf("advances 累计应=1，得 %d", st.Advances)
	}
}

// TestAdvanceGateLanding 落闸锚=挂起（迁移判定落点即判定——规则 13 第一层通用化）：
// core-workflow Planning→（advance）→AwaitingApproval → status=suspended 且停（挂起面不前进）。
func TestAdvanceGateLanding(t *testing.T) {
	runs := NewRunLog(nil)
	seedAdvance(t, runs, "workflows/core-workflow", "core-workflow", Planning, RunRunning)
	if n := runs.Advance(); n != 1 {
		t.Fatalf("落闸锚推进应计入，得 %d", n)
	}
	if r := runAt(runs, 0); r.Phase != string(AwaitingApproval) || r.Status != RunSuspended {
		t.Fatalf("落闸锚应 {AwaitingApproval, suspended}: {phase=%s, status=%s}", r.Phase, r.Status)
	}
	if n := runs.Advance(); n != 0 {
		t.Fatalf("挂起面推进应零动作，得 %d", n)
	}
	if r := runAt(runs, 0); r.Phase != string(AwaitingApproval) || r.Status != RunSuspended {
		t.Fatalf("挂起面应原样保留: {phase=%s, status=%s}", r.Phase, r.Status)
	}
}

// TestAdvanceTerminalLanding 落终点锚=终态（RunFailed=迁移终点语义，非引擎失败语义）：
// dev-delivery Reporting→（advance 相邻行）→Failed → status=failed 且停（终态不迁移）。
func TestAdvanceTerminalLanding(t *testing.T) {
	runs := NewRunLog(nil)
	seedAdvance(t, runs, "workflows/dev-delivery", "dev-delivery", Reporting, RunRunning)
	if n := runs.Advance(); n != 1 {
		t.Fatalf("落终点锚推进应计入，得 %d", n)
	}
	if r := runAt(runs, 0); r.Phase != string(Failed) || r.Status != RunFailed {
		t.Fatalf("落终点锚应 {Failed, failed}: {phase=%s, status=%s}", r.Phase, r.Status)
	}
	if n := runs.Advance(); n != 0 {
		t.Fatalf("终态推进应零动作，得 %d", n)
	}
}

// TestAdvanceGuards 零动作防御面：object 源（相位不投影）/suspended（挂起面不前进）/failed
//（终态不迁移）一律原样保留；空锚（无 phases）零作用。
func TestAdvanceGuards(t *testing.T) {
	runs := NewRunLog(nil)
	// object 源：受理即 triggered 且相位不投影——推进零动作。
	runs.StartRun(TriggerRecord{Seq: 1, Type: "DocApplied", ReceivedAt: time.Now(),
		Route: "workflows/deploy-system-workflow"}, ExecutorBinding{Executor: ExecWorkflow, Target: "deploy-system-workflow"})
	// 挂起态与终态 builtin run（白盒前置）。
	seedAdvance(t, runs, "workflows/core-workflow", "core-workflow", AwaitingApproval, RunSuspended)
	seedAdvance(t, runs, "workflows/dev-delivery", "dev-delivery", Failed, RunFailed)
	if n := runs.Advance(); n != 0 {
		t.Fatalf("防御面全零动作，得 %d", n)
	}
	rs := runs.State().Runs
	if rs[0].Status != RunTriggered || rs[0].Phase != "" {
		t.Fatalf("object 源 run 应原样保留 triggered/空相位: %+v", rs[0])
	}
	if rs[1].Status != RunSuspended || rs[1].Phase != string(AwaitingApproval) {
		t.Fatalf("挂起 run 应原样保留: %+v", rs[1])
	}
	if rs[2].Status != RunFailed || rs[2].Phase != string(Failed) {
		t.Fatalf("终态 run 应原样保留: %+v", rs[2])
	}
}

// TestAdvanceMultiRun 多 run 独立推进：一拍内各 run 沿各自声明表前进一步（逐段承接不跳步）。
func TestAdvanceMultiRun(t *testing.T) {
	runs := NewRunLog(nil)
	seedAdvance(t, runs, "workflows/watch-only", "watch-only", Analyzing, RunRunning)   // →Reporting
	seedAdvance(t, runs, "workflows/core-workflow", "core-workflow", Analyzing, RunRunning) // →Planning
	seedAdvance(t, runs, "workflows/dev-delivery", "dev-delivery", Analyzing, RunRunning)   // →AwaitingApproval（落闸锚）
	if n := runs.Advance(); n != 3 {
		t.Fatalf("一拍应推进 3 条 run，得 %d", n)
	}
	rs := runs.State().Runs
	if rs[0].Phase != string(Reporting) || rs[0].Status != RunRunning {
		t.Fatalf("watch-only 应 {Reporting, running}: %+v", rs[0])
	}
	if rs[1].Phase != string(Planning) || rs[1].Status != RunRunning {
		t.Fatalf("core-workflow 应 {Planning, running}: %+v", rs[1])
	}
	if rs[2].Phase != string(AwaitingApproval) || rs[2].Status != RunSuspended {
		t.Fatalf("dev-delivery 落闸锚应 {AwaitingApproval, suspended}: %+v", rs[2])
	}
}

// TestAdvancerTicker 装配层定时驱动器：interval>0 定时驱动（多拍累计推进）；interval=0
// 关断（Enabled=false，Run 即返——关断=零推进机械判别器）。
func TestAdvancerTicker(t *testing.T) {
	runs := NewRunLog(nil)
	seedAdvance(t, runs, "workflows/watch-only", "watch-only", Analyzing, RunRunning)
	if NewAdvancer(runs, 0).Enabled() || NewAdvancer(nil, time.Millisecond).Enabled() || NewAdvancer(runs, 0).Interval() != 0 {
		t.Fatalf("关断面判据不符（interval=0 或 runs 缺席=不开启）")
	}
	adv := NewAdvancer(runs, 15*time.Millisecond)
	if !adv.Enabled() || adv.Interval() != 15*time.Millisecond {
		t.Fatalf("开启态判据不符: %v %v", adv.Enabled(), adv.Interval())
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go adv.Run(ctx)
	deadline := time.Now().Add(2 * time.Second)
	for runs.State().Advances == 0 && time.Now().Before(deadline) {
		time.Sleep(5 * time.Millisecond)
	}
	if runs.State().Advances == 0 {
		t.Fatalf("定时驱动未推进（advances 恒 0）")
	}
	// 关断器 Run 即返（不可取消语境下也即刻退出——不挂起不泄漏）。
	done := make(chan struct{})
	go func() { NewAdvancer(runs, 0).Run(context.Background()); close(done) }()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatalf("关断推进器 Run 应即刻返回")
	}
}

// TestAdvanceConcurrent 并发安全：Advance 与 State/ResumeSignal 并发（-race 锚）。
func TestAdvanceConcurrent(t *testing.T) {
	runs := NewRunLog(nil)
	seedAdvance(t, runs, "workflows/watch-only", "watch-only", Analyzing, RunRunning)
	var wg sync.WaitGroup
	for i := 0; i < 8; i++ {
		wg.Add(2)
		go func() { defer wg.Done(); runs.Advance() }()
		go func() { defer wg.Done(); _ = runs.State() }()
	}
	wg.Wait()
	// watch-only 序列尾前只有一步可走：其余拍全零动作；相位/状态终值确定。
	if r := runAt(runs, 0); r.Phase != string(Reporting) || r.Status != RunRunning {
		t.Fatalf("并发推进后终值应 {Reporting, running}: {phase=%s, status=%s}", r.Phase, r.Status)
	}
	if st := runs.State(); st.Advances != 1 {
		t.Fatalf("并发推进累计应=1（序列尾零动作不计），得 %d", st.Advances)
	}
}
