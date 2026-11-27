// approval_target_test.go —— 审批信号定向 run（2026-11-27 用户实测「批准 #11 无动作」根因修复）：
// 同锚多 run 挂起时，console 批闸行携带 run 序号——ResumeSignal 只作用该 run；缺席=既有
// 「同锚最新挂起 run」口径（不破旧调用面）。
package workflow

import "testing"

func TestResumeSignalTargetRun(t *testing.T) {
	runs := NewRunLog(nil)
	seedAdvance(t, runs, "workflows/core-workflow", "core-workflow", AwaitingApproval, RunSuspended)
	seedAdvance(t, runs, "workflows/core-workflow", "core-workflow", AwaitingApproval, RunSuspended)
	// 定向 #1：只恢复 #1，#2 仍在闸锚挂起。
	runs.ResumeSignal(TriggerRecord{Type: SignalResume, Involved: "Workflow/global/core-workflow", TargetRun: 1})
	if r := runAt(runs, 0); r.Phase != string(Executing) || r.Status != RunRunning {
		t.Fatalf("定向 #1 应续行到 {Executing, running}: {phase=%s, status=%s}", r.Phase, r.Status)
	}
	if r := runAt(runs, 1); r.Phase != string(AwaitingApproval) || r.Status != RunSuspended {
		t.Fatalf("#2 不应被 #1 的批准带走: {phase=%s, status=%s}", r.Phase, r.Status)
	}
	// 缺席 target=既有口径（同锚最新挂起 run=#2）。
	runs.ResumeSignal(TriggerRecord{Type: SignalResume, Involved: "Workflow/global/core-workflow"})
	if r := runAt(runs, 1); r.Phase != string(Executing) || r.Status != RunRunning {
		t.Fatalf("无 target 应恢复最新挂起 run #2: {phase=%s, status=%s}", r.Phase, r.Status)
	}
	// 目标不存在=零动作（不回落最新——定向词形的失败面必须显式无动作）。
	runs.ResumeSignal(TriggerRecord{Type: SignalResume, Involved: "Workflow/global/core-workflow", TargetRun: 99})
	if r := runAt(runs, 1); r.Phase != string(Executing) {
		t.Fatalf("不存在 target 应零动作: {phase=%s}", r.Phase)
	}
}

func TestRejectSignalTargetRun(t *testing.T) {
	runs := NewRunLog(nil)
	seedAdvance(t, runs, "workflows/core-workflow", "core-workflow", AwaitingApproval, RunSuspended)
	seedAdvance(t, runs, "workflows/core-workflow", "core-workflow", AwaitingApproval, RunSuspended)
	runs.ResumeSignal(TriggerRecord{Type: SignalReject, Involved: "Workflow/global/core-workflow", TargetRun: 1})
	if r := runAt(runs, 0); r.Phase != string(Failed) || r.Status != RunFailed {
		t.Fatalf("定向 #1 驳回应终态 {Failed, failed}: {phase=%s, status=%s}", r.Phase, r.Status)
	}
	if r := runAt(runs, 1); r.Status != RunSuspended {
		t.Fatalf("#2 不应被 #1 的驳回带走: status=%s", r.Status)
	}
}

func TestHistoryRowCarriesSessionID(t *testing.T) {
	runs := NewRunLog(nil)
	seedAdvance(t, runs, "workflows/core-workflow", "core-workflow", Analyzing, RunRunning)
	runs.mu.Lock()
	runs.ring[len(runs.ring)-1].SessionID = "session-real-1"
	runs.mu.Unlock()
	runs.mu.Lock()
	histAppend(&runs.ring[len(runs.ring)-1], "Analyzing", "Planning", "advance")
	runs.mu.Unlock()
	rows := runs.State().Runs[0].History
	if len(rows) < 1 || rows[len(rows)-1].SessionID != "session-real-1" {
		t.Fatalf("history row should carry session anchor: %+v", rows)
	}
}
