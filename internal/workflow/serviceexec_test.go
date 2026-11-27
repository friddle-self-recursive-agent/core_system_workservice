package workflow

import (
	"testing"
	"time"
)

// serviceSpec 测试 spec：Idle→Analyzing→Executing(service)→Verifying(service)→Reporting，
// 富节点投影含 unit/service/op/permissionRole/contextPassing。
var serviceSpec = SpecView{
	Phases: []string{"Idle", "Analyzing", "Executing", "Verifying", "Reporting"},
	Transitions: []TransitionRow{
		{From: "Idle", On: "advance", To: "Analyzing"},
		{From: "Analyzing", On: "advance", To: "Executing"},
		{From: "Executing", On: "advance", To: "Verifying"},
		{From: "Verifying", On: "advance", To: "Reporting"},
	},
	Nodes: []NodeView{
		{Name: "judge", Phase: "Analyzing", Unit: "agent", Op: "judge-issue"},
		{Name: "developer-system", Phase: "Executing", Unit: "service", Service: "developer-system", Op: "develop", Role: "role/dev-ops", ContextFields: []string{"judge", "plan"}},
		{Name: "verify", Phase: "Verifying", Unit: "service", Service: "developer-system", Op: "acceptance"},
		{Name: "report", Phase: "Reporting", Unit: "agent", Op: "archive-report"},
	},
}

// fakeSpawner 拉起面测试替身（记录受理序）。
type fakeSpawner struct{ seen []spawnReq }

func (f *fakeSpawner) SpawnFromRun(run RunRecord, node NodeView) {
	f.seen = append(f.seen, spawnReq{run: run, node: node})
}

func serviceRunRec() TriggerRecord {
	return TriggerRecord{
		Seq: 1, EventID: "evt-1", Type: "ReconcileTick", DedupKey: "d-1",
		At: time.Now(), ReceivedAt: time.Now(), Route: "workflows/dev-loop",
		Involved: "Task/global/t-1", Exec: ExecutorBinding{Executor: ExecWorkflow, Target: "workflows/dev-loop"},
	}
}

// 落点→挂起→拉起→完成续行→下一 service 节点再挂起（链式 developer-system→verify 通路）。
func TestServiceNodeSuspendSpawnResume(t *testing.T) {
	spawner := &fakeSpawner{}
	runs := NewRunLog(&RunOptions{Specs: SpecFunc(func(string) SpecView { return serviceSpec }), ServiceSpawns: spawner})
	runs.StartRun(serviceRunRec(), ExecutorBinding{Executor: ExecWorkflow, Target: "workflows/dev-loop"})

	// 受理首步：Idle→Analyzing（agent 节点——纯续行不挂起）。
	st := runs.State()
	last := st.Runs[len(st.Runs)-1]
	if last.Phase != "Analyzing" || last.Status != RunRunning {
		t.Fatalf("受理首步相位/状态=%s/%s，欲 Analyzing/running", last.Phase, last.Status)
	}
	// 第一拍：Analyzing→Executing（service 节点）——挂起+拉起。
	runs.Advance()
	st = runs.State()
	last = st.Runs[len(st.Runs)-1]
	if last.Phase != "Executing" || last.Status != RunSuspended || last.SuspendWhy != SuspendService {
		t.Fatalf("service 落点相位/状态/因由=%s/%s/%s，欲 Executing/suspended/service", last.Phase, last.Status, last.SuspendWhy)
	}
	h := last.History[len(last.History)-1]
	if h.On != EdgeServiceSpawn || h.To != "Executing" {
		t.Fatalf("落点迁移行 on/to=%s/%s，欲 service-spawn/Executing", h.On, h.To)
	}
	if len(spawner.seen) != 1 || spawner.seen[0].node.Service != "developer-system" ||
		spawner.seen[0].node.Op != "develop" || spawner.seen[0].node.Role != "role/dev-ops" ||
		len(spawner.seen[0].node.ContextFields) != 2 {
		t.Fatalf("拉起请求投影不符: %+v", spawner.seen)
	}
	// 挂起态不被推进拍重复触发（幂等——单飞）。
	runs.Advance()
	if len(spawner.seen) != 1 {
		t.Fatalf("挂起态被重复拉起: %d 次", len(spawner.seen))
	}
	// 完成信号续行（携带会话锚）。
	runs.ServiceSignal(TriggerRecord{Type: SignalServiceDone, Involved: "Workflow/global/dev-loop"}, "sess-123")
	st = runs.State()
	last = st.Runs[len(st.Runs)-1]
	if last.Status != RunRunning || last.SuspendWhy != "" {
		t.Fatalf("续行后状态/因由=%s/%s，欲 running/空", last.Status, last.SuspendWhy)
	}
	h = last.History[len(last.History)-1]
	if h.On != EdgeServiceSpawnResume || h.SessionID != "sess-123" {
		t.Fatalf("续行行 on/session=%s/%s，欲 service-resume/sess-123", h.On, h.SessionID)
	}
	// 第二拍：Executing→Verifying（又是 service 节点）——链式再挂起再拉起。
	runs.Advance()
	st = runs.State()
	last = st.Runs[len(st.Runs)-1]
	if last.Phase != "Verifying" || last.Status != RunSuspended || last.SuspendWhy != SuspendService {
		t.Fatalf("verify 落点相位/状态/因由=%s/%s/%s，欲 Verifying/suspended/service", last.Phase, last.Status, last.SuspendWhy)
	}
	if len(spawner.seen) != 2 || spawner.seen[1].node.Op != "acceptance" {
		t.Fatalf("第二落点拉起投影不符: %+v", spawner.seen)
	}
	// 重复完成信号防护面在装配层桥（MarkCompleted 幂等——TestServiceSpawnLogLifecycle 锚定）；
	// 本面是裸迁移原语，断言「异锚完成信号=零作用」：involved 指向别的 workflow 不误续本锚。
	histLen := len(last.History)
	runs.ServiceSignal(TriggerRecord{Type: SignalServiceDone, Involved: "Workflow/other/wf"}, "sess-x")
	st = runs.State()
	last = st.Runs[len(st.Runs)-1]
	if last.Status != RunSuspended || last.Phase != "Verifying" || len(last.History) != histLen {
		t.Fatalf("异锚完成信号误作用: %s/%s/%d vs %d", last.Phase, last.Status, len(last.History), histLen)
	}
}

// 零翻转口径：ServiceSpawner 缺席（nil）——service 节点照旧纯查表续行不挂起不拉起。
func TestServiceNodeZeroFlipWithoutSpawner(t *testing.T) {
	runs := NewRunLog(&RunOptions{Specs: SpecFunc(func(string) SpecView { return serviceSpec })})
	runs.StartRun(serviceRunRec(), ExecutorBinding{Executor: ExecWorkflow, Target: "workflows/dev-loop"})
	runs.Advance() // → Executing（service）
	runs.Advance() // → Verifying（service）
	st := runs.State()
	last := st.Runs[len(st.Runs)-1]
	if last.Phase != "Verifying" || last.Status != RunRunning || last.SuspendWhy != "" {
		t.Fatalf("零翻转口径被翻转: %s/%s/%s", last.Phase, last.Status, last.SuspendWhy)
	}
}

// ServiceSpawnLog 单飞幂等 + 会话锚回填 + 终态翻转。
func TestServiceSpawnLogLifecycle(t *testing.T) {
	log := NewServiceSpawnLog(nil)
	rec := log.Dispatch(ServiceSpawnRecord{Run: 7, Workflow: "dev-loop", Node: "developer-system", Phase: "Executing", Service: "developer-system", Op: "develop", Status: SpawnSpawned, At: time.Now()})
	if rec.Seq != 1 {
		t.Fatalf("受理序号=%d，欲 1", rec.Seq)
	}
	if dup := log.Dispatch(ServiceSpawnRecord{Run: 7, Workflow: "dev-loop", Node: "developer-system", Status: SpawnSpawned}); dup.Seq != 0 {
		t.Fatalf("单飞未命中（dup.Seq=%d）", dup.Seq)
	}
	if !log.AttachSession(1, "sess-9") {
		t.Fatal("会话锚回填未命中")
	}
	done, sess, ok := log.MarkCompleted(7, "developer-system", SpawnCompleted)
	if !ok || done.Status != SpawnCompleted || sess != "sess-9" {
		t.Fatalf("终态翻转不符: ok=%v status=%s sess=%s", ok, done.Status, sess)
	}
	if _, _, ok := log.MarkCompleted(7, "developer-system", SpawnCompleted); ok {
		t.Fatal("重复完成信号未幂等")
	}
}

// DocsSpecs 富节点解析：unit/service/op/permissionRole/contextPassing.fields 进 NodeView 投影。
func TestDocsSpecsNodeViewParsing(t *testing.T) {
	raw := []byte(`apiVersion: core/v1
kind: Workflow
metadata: {name: dev-loop, scope: global}
spec:
  phases:
    - {name: judge, phase: Analyzing, unit: agent, op: judge-issue}
    - {name: developer-system, phase: Executing, unit: service, service: developer-system, op: develop, permissionRole: role/dev-ops, contextPassing: {previousRound: true, fields: [judge, plan]}}
  transitions:
    - {from: Analyzing, on: advance, to: Executing}
`)
	docs := DocsSpecs{Lookup: func(string) ([]byte, bool) { return raw, true }}
	view := docs.ResolveSpec("dev-loop")
	if len(view.Nodes) != 2 {
		t.Fatalf("节点投影数=%d，欲 2", len(view.Nodes))
	}
	nv := view.Nodes[1]
	if nv.Unit != "service" || nv.Service != "developer-system" || nv.Op != "develop" ||
		nv.Role != "role/dev-ops" || len(nv.ContextFields) != 2 || nv.ContextFields[0] != "judge" {
		t.Fatalf("富节点投影不符: %+v", nv)
	}
	if view.Nodes[0].Unit != "agent" || view.Nodes[0].ContextFields != nil {
		t.Fatalf("agent 节点投影不符: %+v", view.Nodes[0])
	}
}
