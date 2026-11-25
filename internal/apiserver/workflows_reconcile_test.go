// workflows_reconcile_test.go —— 构建生效链判别面 HTTP 载荷验收（v0.55 R-9——console-adoption.md
// §1.29）：聚合面行集 `reconcile` 键（骨架 rv ↔ 绑定 apply 行 rv 对账 chain 三词）+
// `executionFlow`（骨架节点接线投影）+ `executionDigest`（产物 digest 单点真相）+
// `executionHistory`（apply_log 对象链深链参数）；对账三件联动：execution.skeleton /
// reconcile.lastBindRv / executionHistory 同源可查。派单行 nodes 载荷随行断言（拉起投影面）。
package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/flowbind"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
)

// skeletonYaml 骨架对象原文（物化器 Skeleton 同构最小形）。
func skeletonYaml(name, scope string) string {
	return fmt.Sprintf(`apiVersion: core/v1
kind: ConnectorOps
metadata:
  name: %s
  scope: %s
  labels: {core/origin: workflow-bind}
spec:
  workservice: workflow-engine
  workflow: %s
  ops: [flow]
  flow:
    nodes:
      - {name: analyze, unit: session}
      - {name: report, unit: agent}
`, name, scope, name)
}

// fakeApplyLog 固定应用史口面（测试注入：QueryApplies 全量返回行集）。
type fakeApplyLog struct{ rows []opslog.ApplyLog }

func (f fakeApplyLog) QueryEvents(opslog.EventQuery) ([]opslog.EventLog, error) { return nil, nil }
func (f fakeApplyLog) GroupEvents(opslog.EventQuery) ([]opslog.EventGroup, int64, bool, error) {
	return nil, 0, false, nil
}
func (f fakeApplyLog) AppendApply(opslog.ApplyLog) error          { return nil }
func (f fakeApplyLog) CountApply(string) (int64, error)           { return 0, nil }
func (f fakeApplyLog) QueryApply(string) ([]opslog.ApplyLog, error) { return nil, nil }
func (f fakeApplyLog) QueryApplies(obj string, limit int) ([]opslog.ApplyLog, error) {
	return f.rows, nil
}

// reconcileServer 组装带判别面的聚合面实例：flow-probe 骨架在册（workflow-bind 标注）+
// 绑定器 ok 行 rv=2（同窗一致）+ 人写编辑行（外来行——过滤面反证）。
func reconcileServer(t *testing.T, connContent string, connRV int64, rows []opslog.ApplyLog) *httptest.Server {
	t.Helper()
	sch := testSchemeRaw(t, wfFaceFixture)
	repo, err := repository.NewGit(t.TempDir(), sch, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	ctx := context.Background()
	seed := []repository.Object{
		{Kind: "Workflow", Name: "flow-probe", Scope: "global",
			Content: []byte(wfObjectYaml("flow-probe", "global", nil, []string{"analyze", "report"}, ""))},
		{Kind: "ConnectorOps", Name: "flow-probe", Scope: "global", Content: []byte(connContent)},
	}
	for i := range seed {
		o := seed[i]
		if _, _, err := repo.Write(ctx, &o, "seed", ""); err != nil {
			t.Fatalf("seed %s: %v", o.Name, err)
		}
	}
	return httptest.NewServer(New(sch, repo, nil, nil, fakeTriggers{route: "workflows/flow-probe",
		executor: workflow.ExecWorkflow, target: "flow-probe"}, runLogWithRun(t), nil, nil, nil,
		WithApplyLog(fakeApplyLog{rows: rows})))
}

func reconcileRow(t *testing.T, ts *httptest.Server) map[string]any {
	t.Helper()
	code, body := wfGet(t, ts.URL+"/v1/workflows")
	if code != http.StatusOK {
		t.Fatalf("应 200，得 %d", code)
	}
	row := wfRowNamed(t, body, "flow-probe")
	if _, ok := row["reconcile"]; !ok {
		t.Fatalf("行缺 reconcile 键（v0.55 代际判别器）: %v", row)
	}
	return row
}

func TestWorkflowsReconcileConsistent(t *testing.T) {
	now := time.Now()
	rows := []opslog.ApplyLog{
		{FieldManager: "console:editor", Obj: "ConnectorOps/global/flow-probe", ResourceVersion: 1, Result: "ok", Ts: now},
		{FieldManager: "workflow:bind", Obj: "ConnectorOps/global/flow-probe", ResourceVersion: 1,
			Result: "ok", ReasonRef: "workflow-bind:flow-probe", Ts: now},
	}
	ts := reconcileServer(t, skeletonYaml("flow-probe", "global"), 1, rows)
	defer ts.Close()
	row := reconcileRow(t, ts)
	rc := row["reconcile"].(map[string]any)
	if rc["chain"] != flowbind.ChainConsistent || rc["skeletonRv"] != float64(1) || rc["lastBindApply"] != "ok" || rc["lastBindRv"] != float64(1) {
		t.Fatalf("reconcile 应 consistent: %v", rc)
	}
	if rc["lastBindAt"] == "" || rc["bound"] != true || rc["skeleton"] != true {
		t.Fatalf("reconcile 投影不全: %v", rc)
	}
	// executionFlow=骨架节点投影（同源单点真相）；executionDigest 三键；executionHistory 深链。
	flow := row["executionFlow"].([]any)
	if len(flow) != 2 || flow[0].(map[string]any)["name"] != "analyze" || flow[1].(map[string]any)["unit"] != "agent" {
		t.Fatalf("executionFlow 应=骨架节点投影: %v", flow)
	}
	dg := row["executionDigest"].(map[string]any)
	if dg["skeletonRv"] != float64(1) || dg["chain"] != flowbind.ChainConsistent || dg["skeleton"] != true {
		t.Fatalf("executionDigest 不符: %v", dg)
	}
	if row["executionHistory"] != "kind=ConnectorOps&scope=global&name=flow-probe" {
		t.Fatalf("executionHistory 深链不符: %v", row["executionHistory"])
	}
	// 对账三件联动：execution.skeleton=true ∧ reconcile.lastBindRv=2 ∧ history 参数指向同一对象。
	ex := row["execution"].(map[string]any)
	if ex["skeleton"] != true || ex["state"] != "bound" {
		t.Fatalf("execution 投影不符: %v", ex)
	}
}

func TestWorkflowsReconcileDivergent(t *testing.T) {
	// 骨架翻转为实现：对象无 workflow-bind 标注（人写）+ 最近绑定行 rv=2 ≠ 面上 rv=1 → divergent
	// （对象在绑定器之外演进/重建过——链照实分裂可见不静默）。
	rows := []opslog.ApplyLog{{FieldManager: "workflow:bind", Obj: "ConnectorOps/global/flow-probe",
		ResourceVersion: 2, Result: "ok", ReasonRef: "workflow-bind:flow-probe", Ts: time.Now()}}
	ts := reconcileServer(t, `apiVersion: core/v1
kind: ConnectorOps
metadata: {name: flow-probe, scope: global}
spec: {workservice: workflow-engine, workflow: flow-probe, ops: [flow]}
`, 1, rows)
	defer ts.Close()
	row := reconcileRow(t, ts)
	rc := row["reconcile"].(map[string]any)
	if rc["chain"] != flowbind.ChainDivergent || rc["lastBindRv"] != float64(2) || rc["skeletonRv"] != float64(1) {
		t.Fatalf("人写演进应 divergent: %v", rc)
	}
	if rc["skeleton"] != false {
		t.Fatalf("skeleton 应随 execution 翻转=false: %v", rc)
	}
}

func TestWorkflowsReconcileAbsent(t *testing.T) {
	// 对象缺席：reconcile 面缺席=execution 照实 unbound（缺口词形在 execution 侧，
	// reconcile=null 不冒充）。
	sch := testSchemeRaw(t, wfFaceFixture)
	repo, err := repository.NewGit(t.TempDir(), sch, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	ctx := context.Background()
	wf := repository.Object{Kind: "Workflow", Name: "self-only", Scope: "global",
		Content: []byte(wfObjectYaml("self-only", "global", nil, []string{"analyze"}, ""))}
	if _, _, err := repo.Write(ctx, &wf, "seed", ""); err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(New(sch, repo, nil, nil, fakeTriggers{}, nil, nil, nil, nil,
		WithApplyLog(fakeApplyLog{})))
	defer ts.Close()
	_, body := wfGet(t, ts.URL+"/v1/workflows")
	row := wfRowNamed(t, body, "self-only")
	if row["execution"].(map[string]any)["state"] != "unbound" {
		t.Fatalf("缺席应 unbound: %v", row["execution"])
	}
	if row["reconcile"] != nil {
		t.Fatalf("unbound 行 reconcile 应缺席（null）: %v", row["reconcile"])
	}
}

func TestWorkflowsReconcileFaceMissingServesRows(t *testing.T) {
	// applyLog 缺席=判别面缺席（reconcile=null），行集照出不吞行（不冒充）。
	sch := testSchemeRaw(t, wfFaceFixture)
	repo, err := repository.NewGit(t.TempDir(), sch, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	ctx := context.Background()
	wf := repository.Object{Kind: "Workflow", Name: "flow-probe", Scope: "global",
		Content: []byte(wfObjectYaml("flow-probe", "global", nil, []string{"analyze"}, ""))}
	if _, _, err := repo.Write(ctx, &wf, "seed", ""); err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(New(sch, repo, nil, nil, fakeTriggers{}, nil, nil, nil, nil))
	defer ts.Close()
	code, body := wfGet(t, ts.URL+"/v1/workflows")
	if code != http.StatusOK {
		t.Fatalf("应 200，得 %d", code)
	}
	row := wfRowNamed(t, body, "flow-probe")
	if row["reconcile"] != nil {
		t.Fatalf("applyLog 缺席 reconcile 应 null: %v", row["reconcile"])
	}
}

func TestDispatchesPayloadNodesKey(t *testing.T) {
	// 派单行 nodes 键恒在（拉起投影面——workflow-runtime.md §11.2 第三段）。
	dl := flowbind.NewDispatchLog(nil)
	dl.Dispatch(flowbind.DispatchRecord{Agent: flowbind.DispatchAgent, Status: flowbind.DispatchAccepted,
		AutoLevel: flowbind.DispatchAutoLevel, Action: "created", Kind: flowbind.DispatchKindWorkflow,
		Workflow: "flow-probe", Scope: "global", Connector: "flow-probe", RV: 1, EventID: "evt-n1",
		Nodes:    []flowbind.Node{{Name: "analyze", Unit: "session"}, {Name: "report", Unit: "agent"}},
		At:       time.Now()})
	ts := httptest.NewServer(New(testScheme(t), nil, nil, nil, nil, nil, nil, nil, nil, WithDispatchFace(dl)))
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/v1/agents/dispatches")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var got struct {
		Dispatches []map[string]any `json:"dispatches"`
	}
	_ = json.NewDecoder(resp.Body).Decode(&got)
	if len(got.Dispatches) != 1 {
		t.Fatalf("应 1 行: %v", got)
	}
	nodes, ok := got.Dispatches[0]["nodes"].([]any)
	if !ok || len(nodes) != 2 || nodes[0].(map[string]any)["name"] != "analyze" {
		t.Fatalf("派单行缺 nodes 键或投影不符（v0.55 代际判别器）: %v", got.Dispatches[0])
	}
}
