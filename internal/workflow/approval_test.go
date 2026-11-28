package workflow

import (
	"testing"
	"time"
)

// stubJudge 自审批判决替身（机器词形由测试给定；调次数在途防重判据）。
type stubJudge struct {
	v     ApprovalVerdict
	err   error
	calls chan int64
}

func (s *stubJudge) JudgeApproval(run RunRecord, policy ApprovalPolicy, objKey string) (ApprovalVerdict, error) {
	if s.calls != nil {
		s.calls <- run.Run
	}
	return s.v, s.err
}

const agentGateDoc = `apiVersion: core/v1
kind: Workflow
metadata: {name: self-approve-test, scope: global}
spec:
  approval: {mode: agent, autoCriteria: "测试准则", maxRiskLevel: low}
  phases:
    - {name: idle, phase: Idle, unit: connectops}
    - {name: judge, phase: Analyzing, unit: agent}
    - {name: gate, phase: AwaitingApproval, unit: gate}
    - {name: plan, phase: Planning, unit: agent}
    - {name: done, phase: Reporting, unit: agent}
  transitions:
    - {from: Idle, on: advance, to: Analyzing}
    - {from: Analyzing, on: advance, to: AwaitingApproval}
    - {from: AwaitingApproval, on: advance, to: Planning}
    - {from: AwaitingApproval, on: ApprovalGranted, to: Planning}
    - {from: AwaitingApproval, on: ApprovalRejected, to: Failed}
    - {from: Planning, on: advance, to: Reporting}
`

func agentGateLog(t *testing.T, judge ApprovalJudge) *RunLog {
	t.Helper()
	ds := SpecFunc(func(name string) SpecView {
		return DocsSpecs{Lookup: func(string) ([]byte, bool) { return []byte(agentGateDoc), true }}.ResolveSpec(name)
	})
	return NewRunLog(&RunOptions{Specs: ds, ApprovalJudge: judge})
}

func startAgentGate(t *testing.T, r *RunLog) {
	t.Helper()
	r.StartRun(TriggerRecord{Seq: 1, Type: "GhIssueEvent", DedupKey: "gh:t:1", ReceivedAt: time.Now(),
		Route: "workflows/self-approve-test", Involved: "Workflow/global/self-approve-test"},
		ExecutorBinding{Executor: ExecWorkflow, Target: "self-approve-test"})
}

func waitVerdict(t *testing.T, r *RunLog) {
	t.Helper()
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) {
		st := r.State()
		if st.SelfApproved > 0 {
			return
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("判决未在时限内消费: %+v", r.State())
}

// TestSelfApproveGrant agent 模式落闸锚挂起→真判通过→ApprovalGranted 沿表续行。
func TestSelfApproveGrant(t *testing.T) {
	calls := make(chan int64, 4)
	j := &stubJudge{v: ApprovalVerdict{Granted: true, Risk: "low", Reason: "文档修改测试全绿"}, calls: calls}
	r := agentGateLog(t, j)
	startAgentGate(t, r)
	if n := r.Advance(); n != 1 {
		t.Fatalf("推进应 1 步: %d", n)
	}
	run := r.State().Runs[0]
	if run.Phase != string(AwaitingApproval) || run.Status != RunSuspended || run.SuspendWhy != SuspendSelfApprove {
		t.Fatalf("agent 落闸锚应真判挂起: %+v", run)
	}
	waitVerdict(t, r)
	run = r.State().Runs[0]
	if run.Phase != string(Planning) || run.Status != RunRunning {
		t.Fatalf("批准后应沿 ApprovalGranted 续行到 Planning: %+v", run)
	}
	if got := <-calls; got != run.Run {
		t.Fatalf("判决 run 不符: %d", got)
	}
	found := false
	for _, h := range run.History {
		if h.On == EdgeApprovalJudge || h.On == SignalResume {
			found = true
		}
	}
	if !found {
		t.Fatalf("缺判决/批准轨迹行: %+v", run.History)
	}
}

// TestSelfApproveReject low 风险驳回→ApprovalRejected 终态。
func TestSelfApproveReject(t *testing.T) {
	j := &stubJudge{v: ApprovalVerdict{Granted: false, Risk: "low", Reason: "不满足准则"}}
	r := agentGateLog(t, j)
	startAgentGate(t, r)
	r.Advance()
	waitVerdict(t, r)
	run := r.State().Runs[0]
	if run.Phase != string(Failed) || run.Status != RunFailed {
		t.Fatalf("驳回应终态 Failed/failed: %+v", run)
	}
}

// TestSelfApproveRiskFallback 超险（high>max low）→保持挂起并回退人审（SuspendWhy 清空）。
func TestSelfApproveRiskFallback(t *testing.T) {
	j := &stubJudge{v: ApprovalVerdict{Granted: true, Risk: "high", Reason: "生产配置变更"}}
	r := agentGateLog(t, j)
	startAgentGate(t, r)
	r.Advance()
	waitVerdict(t, r)
	run := r.State().Runs[0]
	if run.Phase != string(AwaitingApproval) || run.Status != RunSuspended || run.SuspendWhy != "" {
		t.Fatalf("超险应保持人审挂起: %+v", run)
	}
	escalated := false
	for _, h := range run.History {
		if h.On == EdgeApprovalEscalate {
			escalated = true
		}
	}
	if !escalated {
		t.Fatalf("缺回退人审轨迹行: %+v", run.History)
	}
	if n := r.Advance(); n != 0 {
		t.Fatalf("人审挂起不应被推进: %d", n)
	}
}

// TestSelfApproveZeroFlipWithoutJudge judge 未挂接=旧词形零翻转（agent 盲放行）。
func TestSelfApproveZeroFlipWithoutJudge(t *testing.T) {
	r := agentGateLog(t, nil)
	startAgentGate(t, r)
	r.Advance()
	run := r.State().Runs[0]
	if run.Status != RunRunning || run.Phase != string(AwaitingApproval) {
		t.Fatalf("judge 缺席应保留旧词形盲放行: %+v", run)
	}
}

// TestRiskExceeds 风险序口径（未知词=high 失败收敛）。
func TestRiskExceeds(t *testing.T) {
	cases := []struct {
		risk, max string
		want      bool
	}{
		{"low", "low", false}, {"medium", "low", true}, {"high", "medium", true},
		{"", "", true}, {"???", "high", false},
	}
	for _, c := range cases {
		if got := RiskExceeds(c.risk, c.max); got != c.want {
			t.Errorf("RiskExceeds(%q,%q)=%v want %v", c.risk, c.max, got, c.want)
		}
	}
}

// TestJudgeSuspendedApprovalsBatch 存量挂起 run 批量清账入口（in-flight 防重复派发）。
func TestJudgeSuspendedApprovalsBatch(t *testing.T) {
	calls := make(chan int64, 8)
	j := &stubJudge{v: ApprovalVerdict{Granted: true, Risk: "low"}, calls: calls}
	r := agentGateLog(t, j)
	for i := int64(1); i <= 3; i++ {
		r.StartRun(TriggerRecord{Seq: i, Type: "GhIssueEvent", DedupKey: "gh:batch:" + string(rune('0'+i)),
			ReceivedAt: time.Now(), Route: "workflows/self-approve-test", Involved: "Workflow/global/self-approve-test"},
			ExecutorBinding{Executor: ExecWorkflow, Target: "self-approve-test"})
	}
	r.Advance() // 全部落闸锚（StartRun 首步只到 Analyzing）
	// 入口路径已派发首轮判决（3 条），在途防重使 sweep 不再重复消费。
	d, s := r.JudgeSuspendedApprovals(0)
	if d != 0 {
		t.Fatalf("在途 run 不应重复派发: dispatched=%d skipped=%d", d, s)
	}
	waitVerdict(t, r)
	deadline := time.Now().Add(3 * time.Second)
	for time.Now().Before(deadline) && len(calls) < 3 {
		time.Sleep(5 * time.Millisecond)
	}
	if len(calls) != 3 {
		t.Fatalf("三条挂起 run 各判一次，实际=%d", len(calls))
	}
}
