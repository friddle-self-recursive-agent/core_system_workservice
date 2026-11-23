// workflows_face_test.go —— workflow 运行管理聚合面 HTTP 验收（v0.37 console 读面第二段，
// E3——cored-bridge.md §2 `GET /v1/workflows`）：三源同投（声明∪现势∪计数）、生效 override
// 投影、run involved 载荷键（代际判别器）、方法分立（非 GET → 405）、计数面失败不吞行。
package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
	"gopkg.in/yaml.v3"
)

// wfFaceFixture 聚合面 fixture scheme（Workflow kind 注册 + 双 scope storage 模板——真实
// scheme.yaml 同款形态）。
const wfFaceFixture = `apiVersion: core/v1
kind: Scheme
metadata: {name: default, scope: builtin}
spec:
  version: core/v1
  kinds:
    - {kind: Scheme, builtin: true, storage: core/scheme.yaml}
    - {kind: Workflow, builtin: true, storage: "global/workflows/<name>.yaml | systems/<s>/workflows/<name>.yaml"}
  eventTypes:
    - {type: DocApplied, source: core, defaultLevel: info, route: owner-workflow}
`

// fakeTriggers 固定决议触发面（测试注入：DocApplied→workflows/dev-delivery，其余 dead-letter）。
type fakeTriggers struct{ route, executor, target string }

func (f fakeTriggers) State() workflow.TriggerState { return workflow.TriggerState{} }
func (f fakeTriggers) Resolve(typ, involved string) workflow.DispatchDecision {
	if typ == "DocApplied" && f.route != "" {
		return workflow.DispatchDecision{Matched: true, Route: f.route, For: "any",
			Exec: workflow.ExecutorBinding{Executor: f.executor, Target: f.target}}
	}
	return workflow.DispatchDecision{Route: "dead-letter",
		Exec: workflow.ExecutorBinding{Executor: workflow.ExecNone}}
}

// fakeEventLog 固定分组事件归档面（测试注入：since 零值=全量组；since 非零=窗口组）。
type fakeEventLog struct {
	all    []opslog.EventGroup
	recent []opslog.EventGroup
}

func (f fakeEventLog) QueryEvents(q opslog.EventQuery) ([]opslog.EventLog, error) { return nil, nil }
func (f fakeEventLog) GroupEvents(q opslog.EventQuery) ([]opslog.EventGroup, int64, bool, error) {
	if q.Since.IsZero() {
		return f.all, 10, false, nil
	}
	return f.recent, 4, false, nil
}

// wfObjectYaml 组一个 Workflow 对象声明原文（enabled 可空=未声明）。
func wfObjectYaml(name, scope string, enabled *bool, phases []string, origin string) string {
	doc := map[string]any{
		"apiVersion": "core/v1", "kind": "Workflow",
		"metadata": map[string]any{"name": name, "scope": scope},
		"spec": map[string]any{
			"services": map[string]any{"involved": []string{"system://s1"}, "mode": "act"},
			"dataflow": map[string]any{
				"inputs":  []map[string]any{{"events": []string{"DocApplied"}, "docs": "systems/s1/*"}},
				"outputs": []map[string]any{{"status": "self", "reports": []string{"DriftReport"}}},
			},
		},
	}
	if enabled != nil {
		doc["spec"].(map[string]any)["enabled"] = *enabled
	}
	if len(phases) > 0 {
		doc["spec"].(map[string]any)["phases"] = phases
	}
	md := doc["metadata"].(map[string]any)
	if origin != "" {
		md["labels"] = map[string]any{"core/origin": origin}
	}
	raw, _ := yaml.Marshal(doc)
	return string(raw)
}

// wfFaceServer 组装带聚合面的实例（git 仓库种子=global 本体+系统层 override）。
func wfFaceServer(t *testing.T) (*httptest.Server, *repository.GitRepo) {
	t.Helper()
	sch := testSchemeRaw(t, wfFaceFixture)
	repo, err := repository.NewGit(t.TempDir(), sch, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	on := true
	off := false
	seed := []repository.Object{
		{Kind: "Workflow", Name: "dev-delivery", Scope: "global",
			Content: []byte(wfObjectYaml("dev-delivery", "global", &on, nil, ""))},
		{Kind: "Workflow", Name: "dev-delivery", Scope: "s1",
			Content: []byte(wfObjectYaml("dev-delivery", "s1", &off,
				[]string{"Idle", "Analyzing", "Reporting"}, "override-of-builtin"))},
		{Kind: "Workflow", Name: "self-only", Scope: "s1",
			Content: []byte(wfObjectYaml("self-only", "s1", nil, nil, ""))},
	}
	ctx := context.Background()
	for i := range seed {
		o := seed[i]
		if _, _, err := repo.Write(ctx, &o, "seed", ""); err != nil {
			t.Fatalf("seed %s: %v", o.Name, err)
		}
	}
	ts := httptest.NewServer(New(sch, repo, nil, nil, fakeTriggers{route: "workflows/dev-delivery",
		executor: workflow.ExecWorkflow, target: "dev-delivery"}, runLogWithRun(t), nil, nil, nil,
		WithEventLog(fakeEventLog{
			all:    []opslog.EventGroup{{Type: "DocApplied", Involved: "Rule/s1/impl-1", Count: 7}},
			recent: []opslog.EventGroup{{Type: "DocApplied", Involved: "Rule/s1/impl-1", Count: 3}},
		})))
	return ts, repo
}


// runLogWithRun 现势源：一条 dev-delivery run（受理首步=第二相位 running）。
func runLogWithRun(t *testing.T) *workflow.RunLog {
	t.Helper()
	rl := workflow.NewRunLog(nil)
	rl.StartRun(workflow.TriggerRecord{
		Seq: 7, EventID: "evt-wf-1", Type: "DocApplied", Involved: "Rule/s1/impl-1",
		DedupKey: "apply:Rule/systems/s1/impl-1:1", ReceivedAt: time.Now().UTC(),
		Route: "workflows/dev-delivery",
	}, workflow.ExecutorBinding{Executor: workflow.ExecWorkflow, Target: "dev-delivery"})
	return rl
}

func wfGet(t *testing.T, url string) (int, map[string]any) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("载荷不可解析: %v", err)
	}
	return resp.StatusCode, out
}

func wfRowNamed(t *testing.T, payload map[string]any, name string) map[string]any {
	t.Helper()
	for _, it := range payload["workflows"].([]any) {
		if m, ok := it.(map[string]any); ok && m["name"] == name {
			return m
		}
	}
	t.Fatalf("聚合面缺行 %s: %v", name, payload["workflows"])
	return nil
}

func TestWorkflowsAggregateThreeSources(t *testing.T) {
	ts, _ := wfFaceServer(t)
	defer ts.Close()
	code, body := wfGet(t, ts.URL+"/v1/workflows")
	if code != http.StatusOK {
		t.Fatalf("应 200，得 %d %v", code, body)
	}
	row := wfRowNamed(t, body, "dev-delivery")
	// ① 声明源：本体+override 双对象行（global 在前）；生效 override=systems/s1 enabled=false。
	objs := row["objects"].([]any)
	if len(objs) != 2 {
		t.Fatalf("dev-delivery 对象行数=%d，要 2（global 本体+系统层 override）", len(objs))
	}
	if objs[0].(map[string]any)["scope"] != "global" || objs[0].(map[string]any)["origin"] != "builtin" {
		t.Errorf("首行应 global 本体 origin=builtin: %v", objs[0])
	}
	ov := objs[1].(map[string]any)
	if ov["scope"] != "s1" || ov["origin"] != "override-of-builtin" {
		t.Errorf("override 行不符: %v", ov)
	}
	if en, ok := ov["enabled"].(bool); !ok || en {
		t.Errorf("override enabled 应 false（spec.enabled 投影）: %v", ov["enabled"])
	}
	eff := row["effective"].(map[string]any)
	if eff["scope"] != "s1" {
		t.Errorf("effective.scope=%v，要 s1", eff["scope"])
	}
	// 相位序列=override 声明优先（对象 spec.phases 投影，非 builtin 注册表序列）。
	if got := fmt.Sprint(row["phases"]); got != "[Idle Analyzing Reporting]" {
		t.Errorf("phases=%s，要 override 声明序列", got)
	}
	// ③ 计数源：全量=7、24h=3（fakeEventLog 分组 ×dispatch 决议）。
	if row["runsTotal"] != float64(7) || row["runs24h"] != float64(3) {
		t.Errorf("计数不符: total=%v 24h=%v", row["runsTotal"], row["runs24h"])
	}
	// ② 现势源：环内最新 run（受理首步=第二相位 running）。
	latest := row["latest"].(map[string]any)
	if latest["phase"] != "Analyzing" || latest["status"] != "running" {
		t.Errorf("latest 不符: %v", latest)
	}
	if latest["eventId"] != "evt-wf-1" {
		t.Errorf("latest.eventId 应与留痕同源: %v", latest)
	}
	// 系统层自有对象行：无 override 语义时 effective=null（builtin 本体生效语义对 non-builtin=无覆盖面）。
	selfRow := wfRowNamed(t, body, "self-only")
	if selfRow["effective"] != nil {
		t.Errorf("self-only 无 override 行 effective 应 null: %v", selfRow["effective"])
	}
	if selfRow["specSource"] != workflow.SpecObject {
		t.Errorf("self-only specSource 应 object: %v", selfRow["specSource"])
	}
}

func TestWorkflowsAggregateEmptyFacesStillServesRegistry(t *testing.T) {
	// 全 face 缺席（objects/runs/triggers/eventLog=nil）：builtin 注册表行照出（注册表=代码声明源），
	// 计数/现势=诚实零值/缺席——不 503（本面无新契约面，缺的是源数据不是能力）。
	sch := testSchemeRaw(t, wfFaceFixture)
	ts := httptest.NewServer(New(sch, nil, nil, nil, nil, nil, nil, nil, nil))
	defer ts.Close()
	code, body := wfGet(t, ts.URL+"/v1/workflows")
	if code != http.StatusOK {
		t.Fatalf("应 200，得 %d", code)
	}
	if body["count"].(float64) == 0 {
		t.Fatalf("builtin 注册表六项应入册（声明源①注册表半边）: %v", body)
	}
	row := wfRowNamed(t, body, "dev-delivery")
	if row["specSource"] != workflow.SpecBuiltin {
		t.Errorf("builtin 行 specSource 应 builtin: %v", row["specSource"])
	}
	if n := len(row["phases"].([]any)); n == 0 {
		t.Errorf("builtin 行相位序列应=注册表声明（v0.19 六项对齐）: %v", row["phases"])
	}
	if row["latest"] != nil || row["objects"] == nil {
		t.Errorf("空面行 latest 应缺席∧objects 应空数组: %v", row)
	}
	if row["runsTotal"] != float64(0) || row["runs24h"] != float64(0) {
		t.Errorf("空面计数应 0: %v", row)
	}
}

func TestWorkflowsRunsPayloadInvolvedKey(t *testing.T) {
	// run 载荷新增 involved 键（v0.37 代际判别器）：与留痕 involved 同源（涉及对象链第一跳）。
	sch := testSchemeRaw(t, wfFaceFixture)
	ts := httptest.NewServer(New(sch, nil, nil, nil, nil, runLogWithRun(t), nil, nil, nil))
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/v1/workflows/runs")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	run := body["runs"].([]any)[0].(map[string]any)
	if run["involved"] != "Rule/s1/impl-1" {
		t.Errorf("run.involved 应=留痕 involved: %v", run["involved"])
	}
}

func TestWorkflowsMethodDisallowed(t *testing.T) {
	sch := testSchemeRaw(t, wfFaceFixture)
	ts := httptest.NewServer(New(sch, nil, nil, nil, nil, nil, nil, nil, nil))
	defer ts.Close()
	resp, err := http.Post(ts.URL+"/v1/workflows", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("POST /v1/workflows 应 405（查询面只读），得 %d", resp.StatusCode)
	}
}

// testSchemeRaw 从原文装载 scheme（聚合面 fixture：Workflow kind 双 scope storage 模板）。
func testSchemeRaw(t *testing.T, raw string) *scheme.Scheme {
	t.Helper()
	sch, err := scheme.Load([]byte(raw))
	if err != nil {
		t.Fatalf("载入 fixture 失败: %v", err)
	}
	return sch
}

func TestParseWfObjectRichPhasesLenient(t *testing.T) {
	// 富相位行（{name,title,op,…}——v0.8 吸收 Loop 的对象声明形态）+ 局部字段缺失=宽松投影
	//（phases 照投影，enabled/origin 不丢）。
	raw := []byte(`apiVersion: core/v1
kind: Workflow
metadata:
  name: dev-delivery
  scope: s1
  labels: {core/origin: override-of-builtin}
spec:
  enabled: true
  phases:
    - {name: analyze, title: Analyze}
    - name: notify-report
      title: Notify Report
      op: notify/send-email
`)
	row := parseWfObject(raw, 3, "s1")
	if row.Origin != "override-of-builtin" || row.Enabled == nil || !*row.Enabled {
		t.Fatalf("origin/enabled 投影不符: %v %v", row.Origin, row.Enabled)
	}
	if fmt.Sprint(row.Phases) != "[analyze notify-report]" {
		t.Fatalf("富相位行展示名投影不符: %v", row.Phases)
	}
	// 坏文档=零投影不致命（行仍入册）。
	bad := parseWfObject([]byte("\t\t不可解析 \x00{{"), 1, "s1")
	if bad.Scope != "s1" || bad.Phases == nil || bad.Enabled != nil {
		t.Fatalf("坏文档投影口径不符: %+v", bad)
	}
}
