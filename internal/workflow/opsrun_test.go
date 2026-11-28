// opsrun_test.go —— connectops 节点真执行面验收（2026-11-28 用户裁定「做成workflow的一个节点。
// 即gh-release包括git push/release[scheme调用]」——opsrun.go）：落点挂起（suspendWhy=connectops
// + on=connectops-run）→ 回执续行（on=connectops-resume + Note 摘要）/ 失败留挂起
// （on=connectops-error）；零翻转口径=Runner 缺席照旧纯查表推进。
package workflow

import (
	"testing"
	"time"
)

// opsSpec 测试 spec：Idle→Deploying(connectops)→Reporting，节点带 Connector/Params 投影。
var opsSpec = SpecView{
	Phases: []string{"Idle", "Deploying", "Reporting"},
	Transitions: []TransitionRow{
		{From: "Idle", On: "advance", To: "Deploying"},
		{From: "Deploying", On: "advance", To: "Reporting"},
	},
	Nodes: []NodeView{
		{Name: "gh-release", Phase: "Deploying", Unit: "connectops", Connector: "connectops-docs",
			Op: "release-publish", Params: map[string]any{"repo": "core_system_workservice"}},
	},
}

func opsRunRec() TriggerRecord {
	return TriggerRecord{
		Seq: 1, EventID: "evt-ops-1", Type: "ReconcileTick", DedupKey: "d-ops-1",
		At: time.Now(), ReceivedAt: time.Now(), Route: "workflows/dev-loop",
		Involved: "Task/global/t-ops", Exec: ExecutorBinding{Executor: ExecWorkflow, Target: "workflows/dev-loop"},
	}
}

// 落点→挂起（connectops-run 行）→成功回执→续行（connectops-resume 行+Note）。
func TestOpsNodeSuspendResume(t *testing.T) {
	runner := &fakeOpsRunner{}
	runs := NewRunLog(&RunOptions{Specs: SpecFunc(func(string) SpecView { return opsSpec }), ConnectOpsRunner: runner})
	runs.StartRun(opsRunRec(), ExecutorBinding{Executor: ExecWorkflow, Target: "workflows/dev-loop"})
	runs.Advance() // Idle→Deploying（connectops 落点挂起）

	st := runs.State()
	last := st.Runs[len(st.Runs)-1]
	if last.Phase != "Deploying" || last.Status != RunSuspended || last.SuspendWhy != SuspendOps {
		t.Fatalf("connectops 落点相位/状态/因由=%s/%s/%s，欲 Deploying/suspended/connectops", last.Phase, last.Status, last.SuspendWhy)
	}
	h := last.History[len(last.History)-1]
	if h.On != EdgeOpsRun {
		t.Fatalf("落点迁移行 on=%s，欲 %s", h.On, EdgeOpsRun)
	}
	if len(runner.seen) != 1 || runner.seen[0].node.Connector != "connectops-docs" || runner.seen[0].node.Op != "release-publish" {
		t.Fatalf("执行面未按节点 Connector/Op 派发：%+v", runner.seen)
	}

	runs.OpsSignal(last.Run, last.Workflow, "docs committed abc1234", true)
	st = runs.State()
	last = st.Runs[len(st.Runs)-1]
	if last.Status != RunRunning || last.SuspendWhy != "" {
		t.Fatalf("成功回执后续行状态=%s/%s，欲 running/（空）", last.Status, last.SuspendWhy)
	}
	h = last.History[len(last.History)-1]
	if h.On != EdgeOpsResume || h.Note != "docs committed abc1234" {
		t.Fatalf("续行行 on/note=%s/%q，欲 %s/docs committed abc1234", h.On, h.Note, EdgeOpsResume)
	}

	// 续行后机械推进：Deploying→Reporting→无出边收束（F-04 末步收束）。
	runs.Advance()
	runs.Advance()
	st = runs.State()
	last = st.Runs[len(st.Runs)-1]
	if last.Status != RunConverged {
		t.Fatalf("末步应收束 converged，得 %s", last.Status)
	}
}

// 失败回执→保持挂起（connectops-error 行+显式错误）——缺口可见不静默。
func TestOpsNodeFailureStaysSuspended(t *testing.T) {
	runner := &fakeOpsRunner{}
	runs := NewRunLog(&RunOptions{Specs: SpecFunc(func(string) SpecView { return opsSpec }), ConnectOpsRunner: runner})
	runs.StartRun(opsRunRec(), ExecutorBinding{Executor: ExecWorkflow, Target: "workflows/dev-loop"})
	runs.Advance()
	st := runs.State()
	last := st.Runs[len(st.Runs)-1]

	runs.OpsSignal(last.Run, last.Workflow, "op 执行失败: 连接器 connectops-docs 未注册", false)
	st = runs.State()
	last = st.Runs[len(st.Runs)-1]
	if last.Status != RunSuspended || last.SuspendWhy != SuspendOps {
		t.Fatalf("失败回执后应保持挂起，得 %s/%s", last.Status, last.SuspendWhy)
	}
	h := last.History[len(last.History)-1]
	if h.On != EdgeOpsError || h.Note == "" {
		t.Fatalf("失败行 on/note=%s/%q，欲 %s/非空显式错误", h.On, h.Note, EdgeOpsError)
	}
}

// 零翻转口径：Runner 缺席（nil）=connectops 节点照旧纯查表推进不挂起。
func TestOpsNodeAbsentRunnerZeroFlip(t *testing.T) {
	runs := NewRunLog(&RunOptions{Specs: SpecFunc(func(string) SpecView { return opsSpec })})
	runs.StartRun(opsRunRec(), ExecutorBinding{Executor: ExecWorkflow, Target: "workflows/dev-loop"})
	runs.Advance()
	runs.Advance()
	st := runs.State()
	last := st.Runs[len(st.Runs)-1]
	if last.Status == RunSuspended {
		t.Fatalf("Runner 缺席=零翻转（不挂起），得 %s/%s", last.Status, last.SuspendWhy)
	}
}

// fakeOpsRunner op 执行面测试替身（记录受理节点投影）。
type fakeOpsRunner struct{ seen []spawnReq }

func (f *fakeOpsRunner) RunOp(run RunRecord, node NodeView) {
	f.seen = append(f.seen, spawnReq{run: run, node: node})
}
