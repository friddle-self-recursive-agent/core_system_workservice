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
	// 派发=出锁 goroutine（非阻塞契约）——轮询等待受理落账。
	deadline := time.Now().Add(2 * time.Second)
	for len(runner.seen) == 0 {
		if time.Now().After(deadline) {
			t.Fatalf("执行面未派发（2s 超时）")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if runner.seen[0].node.Connector != "connectops-docs" || runner.seen[0].node.Op != "release-publish" {
		t.Fatalf("执行面未按节点 Connector/Op 派发：%+v", runner.seen)
	}

	runs.OpsSignal(last.Run, last.Workflow, "docs committed abc1234", true, "")
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

	runs.OpsSignal(last.Run, last.Workflow, "op 执行失败: 连接器 connectops-docs 未注册", false, "")
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

// service 首步落点（2026-11-28 sybran-release-update 用户自有 workflow——首步直落 unit=service
// 且 ServiceSpawner 挂接 → 挂起+派发；缺口与 ops 钩子同源同修）。
func TestServiceNodeFirstStepSuspend(t *testing.T) {
	spec := SpecView{
		Phases: []string{"Idle", "Reporting"},
		Transitions: []TransitionRow{
			{From: "Idle", On: "advance", To: "Reporting"},
		},
		Nodes: []NodeView{
			{Name: "dispatch-update", Phase: "Reporting", Unit: "service", Service: "sybran-test-agent", Op: "update-on-release"},
		},
	}
	spawner := &fakeSpawner{}
	runs := NewRunLog(&RunOptions{Specs: SpecFunc(func(string) SpecView { return spec }), ServiceSpawns: spawner})
	runs.StartRun(serviceRunRec(), ExecutorBinding{Executor: ExecWorkflow, Target: "workflows/dev-loop"})
	st := runs.State()
	last := st.Runs[len(st.Runs)-1]
	if last.Phase != "Reporting" || last.Status != RunSuspended || last.SuspendWhy != SuspendService {
		t.Fatalf("service 首步落点相位/状态/因由=%s/%s/%s，欲 Reporting/suspended/service", last.Phase, last.Status, last.SuspendWhy)
	}
	deadline := time.Now().Add(2 * time.Second)
	for len(spawner.seen) == 0 {
		if time.Now().After(deadline) {
			t.Fatalf("拉起面未派发（2s 超时）")
		}
		time.Sleep(5 * time.Millisecond)
	}
	if spawner.seen[0].node.Service != "sybran-test-agent" {
		t.Fatalf("拉起未按 service 声明派发：%+v", spawner.seen[0].node)
	}
}

// 判断分叉（2026-11-28 用户裁决「workflow做分叉」）：service 完成回执携带 decision 词 →
// 查表命中判断边立即走（history on=release）→ 落点钩子/机械推进照常。
func TestServiceDecisionBranch(t *testing.T) {
	spec := SpecView{
		Phases: []string{"Idle", "Verifying", "Released", "Reporting"},
		Transitions: []TransitionRow{
			{From: "Idle", On: "advance", To: "Verifying"},
			{From: "Verifying", On: "advance", To: "Reporting"}, // 默认：跳过发版
			{From: "Verifying", On: "release", To: "Released"},  // 判断边
			{From: "Released", On: "advance", To: "Reporting"},
		},
		Nodes: []NodeView{
			{Name: "verify", Phase: "Verifying", Unit: "service", Service: "developer-system", Op: "acceptance"},
		},
	}
	spawner := &fakeSpawner{}
	runs := NewRunLog(&RunOptions{Specs: SpecFunc(func(string) SpecView { return spec }), ServiceSpawns: spawner})
	runs.StartRun(serviceRunRec(), ExecutorBinding{Executor: ExecWorkflow, Target: "workflows/dev-loop"})
	st := runs.State()
	last := st.Runs[len(st.Runs)-1]

	runs.ServiceSignal(TriggerRecord{
		Type: SignalServiceDone, Involved: "Workflow/global/dev-loop", TargetRun: last.Run,
		Decision: "release", At: time.Now(), ReceivedAt: time.Now(),
	}, "sess-1")
	st = runs.State()
	last = st.Runs[len(st.Runs)-1]
	if last.Phase != "Released" || last.Status != RunRunning {
		t.Fatalf("decision=release 应沿判断边走 Released/running，得 %s/%s", last.Phase, last.Status)
	}
	var branchRow *HistoryRow
	for i := range last.History {
		if last.History[i].On == "release" {
			branchRow = &last.History[i]
		}
	}
	if branchRow == nil || branchRow.From != "Verifying" || branchRow.To != "Released" {
		t.Fatalf("判断边迁移行缺席或端点错： %+v", last.History)
	}
	runs.Advance() // Released→Reporting
	runs.Advance() // 无出边 → converged
	st = runs.State()
	last = st.Runs[len(st.Runs)-1]
	if last.Status != RunConverged {
		t.Fatalf("分叉后应收束，得 %s", last.Status)
	}
}

// 未携带 decision=机械 advance 兜底（默认路径不动——advance 边=默认跳过发版词形）。
func TestServiceNoDecisionDefaultAdvance(t *testing.T) {
	spec := SpecView{
		Phases: []string{"Idle", "Verifying", "Released", "Reporting"},
		Transitions: []TransitionRow{
			{From: "Idle", On: "advance", To: "Verifying"},
			{From: "Verifying", On: "advance", To: "Reporting"},
			{From: "Verifying", On: "release", To: "Released"},
			{From: "Released", On: "advance", To: "Reporting"},
		},
		Nodes: []NodeView{
			{Name: "verify", Phase: "Verifying", Unit: "service", Service: "developer-system", Op: "acceptance"},
		},
	}
	spawner := &fakeSpawner{}
	runs := NewRunLog(&RunOptions{Specs: SpecFunc(func(string) SpecView { return spec }), ServiceSpawns: spawner})
	runs.StartRun(serviceRunRec(), ExecutorBinding{Executor: ExecWorkflow, Target: "workflows/dev-loop"})
	st := runs.State()
	last := st.Runs[len(st.Runs)-1]

	runs.ServiceSignal(TriggerRecord{
		Type: SignalServiceDone, Involved: "Workflow/global/dev-loop", TargetRun: last.Run,
		At: time.Now(), ReceivedAt: time.Now(),
	}, "sess-1")
	runs.Advance()
	st = runs.State()
	last = st.Runs[len(st.Runs)-1]
	if last.Phase != "Reporting" {
		t.Fatalf("无 decision 应机械 advance 落 Reporting（默认跳过发版），得 %s", last.Phase)
	}
}
