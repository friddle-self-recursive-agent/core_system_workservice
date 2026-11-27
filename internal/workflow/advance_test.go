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
// 序列尾（watch-only Reporting 无 advance 出行）首拍落 converged（F-04 正常收束——
// workflow.md §1：到达 Reporting 完成即正常收束），其后零动作停（终态不迁移）。
func TestAdvanceSingleStep(t *testing.T) {
	runs := NewRunLog(nil)
	seedAdvance(t, runs, "workflows/watch-only", "watch-only", Analyzing, RunRunning)
	if n := runs.Advance(); n != 1 {
		t.Fatalf("单步推进应推进 1 条 run，得 %d", n)
	}
	if r := runAt(runs, 0); r.Phase != string(Reporting) || r.Status != RunRunning {
		t.Fatalf("推进后应 {Reporting, running}: {phase=%s, status=%s}", r.Phase, r.Status)
	}
	// 序列尾：Reporting 无 advance 出行——本拍即落 converged（正向末步收束，F-04）。
	if n := runs.Advance(); n != 0 {
		t.Fatalf("序列尾收束拍应零迁移计数，得 %d", n)
	}
	if r := runAt(runs, 0); r.Phase != string(Reporting) || r.Status != RunConverged {
		t.Fatalf("序列尾应落 {Reporting, converged}（F-04 正常收束）: {phase=%s, status=%s}", r.Phase, r.Status)
	}
	// 收束轨迹随行：{Reporting,converged,on=converged} 一行（history 与静态表分离）。
	hist := runAt(runs, 0).History
	if len(hist) == 0 || hist[len(hist)-1].On != "converged" || hist[len(hist)-1].To != string(Reporting) {
		t.Fatalf("收束轨迹应落 converged 行: %+v", hist)
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

// TestAdvanceConvergedLanding 正向末步收束（F-04——Reporting→converged，非 Failed）：
// dev-delivery Reporting 处 advance 零动作停（侧轨 Failed 已摘除线性尾，advance 边止于
// Reporting）→ 正向末步落 converged（终态不迁移）。
func TestAdvanceConvergedLanding(t *testing.T) {
	runs := NewRunLog(nil)
	seedAdvance(t, runs, "workflows/dev-delivery", "dev-delivery", Reporting, RunRunning)
	// Reporting 已是正向末步：无 advance 出行——零迁移计数，status 落 converged。
	if n := runs.Advance(); n != 0 {
		t.Fatalf("正向末步收束拍应零迁移计数，得 %d", n)
	}
	if r := runAt(runs, 0); r.Phase != string(Reporting) || r.Status != RunConverged {
		t.Fatalf("正向末步应 {Reporting, converged}（F-04 收束——绝不 Reporting→Failed）: {phase=%s, status=%s}", r.Phase, r.Status)
	}
	if n := runs.Advance(); n != 0 {
		t.Fatalf("终态推进应零动作，得 %d", n)
	}
}

// TestAdvanceGuards 零动作防御面：object 源（相位不投影）/suspended（挂起面不前进）/failed
// （终态不迁移）一律原样保留；空锚（无 phases）零作用。
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
	seedAdvance(t, runs, "workflows/watch-only", "watch-only", Analyzing, RunRunning)       // →Reporting
	seedAdvance(t, runs, "workflows/core-workflow", "core-workflow", Analyzing, RunRunning) // →Planning
	seedAdvance(t, runs, "workflows/dev-delivery", "dev-delivery", Analyzing, RunRunning)   // →AwaitingApproval（落闸锚）
	if n := runs.Advance(); n != 3 {
		t.Fatalf("一拍应推进 3 条 run，得 %d", n)
	}
	rs := runs.State().Runs
	if rs[0].Phase != string(Reporting) || rs[0].Status != RunRunning {
		t.Fatalf("watch-only 应 {Reporting, running}（收束=下一拍序列尾判定——一步一迁移）: %+v", rs[0])
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
	// watch-only 只一步可走+并发拍序列尾收束（F-04 converged）：相位/状态终值确定。
	if r := runAt(runs, 0); r.Phase != string(Reporting) || r.Status != RunConverged {
		t.Fatalf("并发推进后终值应 {Reporting, converged}: {phase=%s, status=%s}", r.Phase, r.Status)
	}
	if st := runs.State(); st.Advances != 1 {
		t.Fatalf("并发推进累计应=1（迁移计数；收束零迁移计数），得 %d", st.Advances)
	}
}

// TestAdvanceFullPhasePath G-13 相位表全路径锚：core-workflow run 自受理首步起随拍走完
// 声明序列（Analyzing→Planning→AwaitingApproval 挂起→批准恢复→Executing→Verifying→gate
// 放行→Reporting→converged 正常收束——F-04 状态机纠偏：正向末步收束绝不落 Failed），
// advances 随 tick 单调增长、落闸锚挂起面不前进、恢复续行、收束即停。
func TestAdvanceFullPhasePath(t *testing.T) {
	runs := NewRunLog(&RunOptions{VerifyGate: VerifyGateFunc(func(run RunRecord) VerifyVerdict {
		return VerifyVerdict{OK: true, AuditID: "audit-full", Note: "audit ok"}
	})})
	runs.StartRun(TriggerRecord{
		Seq: 1, EventID: "ev-full", Type: "DocApplied", DedupKey: "apply:full:1",
		ReceivedAt: time.Now(), Route: "workflows/core-workflow", Owner: "global",
	}, ExecutorBinding{Executor: ExecWorkflow, Target: "core-workflow"})
	// 受理首步（同拍同步）：Idle→Analyzing，status=running。
	if r := runAt(runs, 0); r.Phase != string(Analyzing) || r.Status != RunRunning {
		t.Fatalf("受理首步应 {Analyzing, running}: {phase=%s, status=%s}", r.Phase, r.Status)
	}
	tick := func(wantPhase Phase, wantStatus string) {
		t.Helper()
		if n := runs.Advance(); n != 1 {
			t.Fatalf("tick 应推进 1 步，得 %d", n)
		}
		r := runAt(runs, 0)
		if r.Phase != string(wantPhase) || r.Status != wantStatus {
			t.Fatalf("tick 后应 {%s, %s}: {phase=%s, status=%s}", wantPhase, wantStatus, r.Phase, r.Status)
		}
	}
	tick(Planning, RunRunning)
	tick(AwaitingApproval, RunSuspended) // 落闸锚=挂起
	// 挂起面不前进：闸锚拍推进零动作（审批信号之前相位机等待）。
	if n := runs.Advance(); n != 0 {
		t.Fatalf("挂起面应零动作，得 %d", n)
	}
	// 批准恢复（ApprovalGranted——闸锚沿声明序列续行，status 翻回 running）。
	runs.ResumeSignal(TriggerRecord{
		Seq: 2, EventID: "ev-grant", Type: SignalResume, Involved: "Workflow/global/core-workflow",
		ReceivedAt: time.Now(),
	})
	if r := runAt(runs, 0); r.Phase != string(Executing) || r.Status != RunRunning {
		t.Fatalf("批准恢复应续行到 {Executing, running}: {phase=%s, status=%s}", r.Phase, r.Status)
	}
	tick(Verifying, RunRunning)
	tick(Reporting, RunRunning) // gate 放行拍（Verifying→Reporting 过审）
	// F-04：Reporting=正向末步（侧轨 Failed 已摘除线性尾）——下一拍序列尾收束（零迁移计数）。
	if n := runs.Advance(); n != 0 {
		t.Fatalf("收束拍应零迁移计数，得 %d", n)
	}
	// 终态不迁移：converged 后续拍零动作（终态制）。
	if n := runs.Advance(); n != 0 {
		t.Fatalf("终态后续拍应零动作，得 %d", n)
	}
	st := runs.State()
	if st.Advances != 4 { // Planning/AwaitingApproval/Verifying/Reporting(gate 放行) 四拍步（挂起拍零动作、恢复走审批信号、收束零迁移计数——F-04）
		t.Fatalf("advances 累计应=4，得 %d", st.Advances)
	}
	r := runAt(runs, 0)
	if r.Status != RunConverged || r.Phase != string(Reporting) {
		t.Fatalf("终态词形应 {Reporting, converged}（F-04 正向末步收束）: %+v", r)
	}
	if r.Verify == nil || !r.Verify.OK {
		t.Fatalf("Verify 放行判定应随 run 留痕：%+v", r.Verify)
	}
}

// F04-HIST-01 运行轨迹留痕（2026-11-26 F-04——静态 transitions 与运行轨迹 history 分离）：
// 走完 Analyzing→Reporting 全路径的 run 携带逐拍真实迁移行（from/to/on/at 全量词形），
// 收束行 on=converged；UI 按 history 标绿已走节点，静态表不再当事件时间线。
func TestRunHistoryTrajectory(t *testing.T) {
	runs := NewRunLog(&RunOptions{VerifyGate: VerifyGateFunc(func(run RunRecord) VerifyVerdict {
		return VerifyVerdict{OK: true, AuditID: "audit-hist"}
	})})
	runs.StartRun(TriggerRecord{Seq: 1, Type: "DocApplied", DedupKey: "apply:hist:1",
		ReceivedAt: time.Now(), Route: "workflows/watch-only", Owner: "global"},
		ExecutorBinding{Executor: ExecWorkflow, Target: "watch-only"})
	if n := runs.Advance(); n != 1 { // Analyzing→Reporting
		t.Fatalf("单步推进: %d", n)
	}
	if n := runs.Advance(); n != 0 { // 序列尾收束
		t.Fatalf("收束拍零迁移: %d", n)
	}
	r := runAt(runs, 0)
	if len(r.History) != 3 {
		t.Fatalf("轨迹应 3 行（受理首步+推进+收束）: %+v", r.History)
	}
	want := [][3]string{{string(Idle), string(Analyzing), EdgeAdvance}, {string(Analyzing), string(Reporting), EdgeAdvance}, {string(Reporting), string(Reporting), "converged"}}
	for i, w := range want {
		h := r.History[i]
		if h.From != w[0] || h.To != w[1] || h.On != w[2] || h.At == "" {
			t.Fatalf("轨迹行 %d 应 %+v: %+v", i, w, h)
		}
	}
}

// F04-CONV-01 驳回仍达侧轨 Failed（纠偏后的负向通路锚）：声明序列不带 Failed，
// ApprovalRejected 边仍物化（MaterializeTransitions 无条件物化）——驳回查表命中
// AwaitingApproval→Failed 终态迁移（侧轨语义与正向收敛并存）。
func TestRejectSideTrackUnchanged(t *testing.T) {
	runs := NewRunLog(nil)
	runs.StartRun(TriggerRecord{Seq: 1, Type: "DocApplied", DedupKey: "apply:rj:1",
		ReceivedAt: time.Now(), Route: "workflows/dev-delivery", Owner: "global"},
		ExecutorBinding{Executor: ExecWorkflow, Target: "dev-delivery"})
	sig := func(typ string) {
		runs.ResumeSignal(TriggerRecord{Seq: 100, Type: typ, Involved: "Workflow/builtin/dev-delivery", ReceivedAt: time.Now()})
	}
	sig(SignalSuspend)
	sig(SignalReject)
	r := runAt(runs, 0)
	if r.Phase != string(Failed) || r.Status != RunFailed {
		t.Fatalf("驳回应 {Failed, failed}（侧轨终点锚）: {phase=%s, status=%s}", r.Phase, r.Status)
	}
	// 轨迹含驳回行（on=ApprovalRejected）——侧轨迁移与推进迁移同口径留痕。
	last := r.History[len(r.History)-1]
	if last.On != SignalReject || last.To != string(Failed) {
		t.Fatalf("轨迹尾应驳回行: %+v", last)
	}
}
