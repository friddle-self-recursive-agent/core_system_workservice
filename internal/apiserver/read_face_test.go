// read_face_test.go —— v0.66 读面验收（R-ENG §3）：单事件详情（200 关联链/404 不在归档/
// 503 face 缺席）与 loop/current（有 wake=running 直投/无 wake=unknown 全 null）。
package apiserver

import (
	"context"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
)

// TestEventDetailFace 单事件详情：归档直投 + 同 dedupKey 后续事件 + involved 键解析 +
// 同 eventId 的 run 关联链；404=不在归档。
func TestEventDetailFace(t *testing.T) {
	sch := testScheme(t)
	opsDB, err := opslog.OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer opsDB.Close()
	// 窗口毫秒级：同 dedupKey 两事件都可归档（真窗口内会聚合——关联链测试需要两行）。
	bus := eventbus.NewBus(&eventbus.Options{Ops: opsDB, DedupWindow: time.Millisecond})
	hub := workflow.NewTriggerHub(nil)
	runs := workflow.NewRunLog(nil)
	hub.BindRuns(runs)
	hub.WireExecutor(workflow.ExecWorkflow)
	stop := hub.AttachDispatch(bus, []scheme.DispatchRule{{On: "DocApplied", For: "any", Dispatch: "workflows/probe"}})
	defer stop()
	hub.Attach(bus, "DocApplied")
	ts := httptest.NewServer(New(sch, nil, nil, nil, hub, runs, nil, nil, nil, WithEventLog(opsDB)))
	defer ts.Close()

	e1 := eventbus.Event{Type: "DocApplied", Source: "apply", Involved: "Agent/global/probe",
		DedupKey: "apply:Agent/global/probe:1", Level: "info"}
	if err := bus.Publish(&e1); err != nil {
		t.Fatal(err)
	}
	time.Sleep(5 * time.Millisecond) // 出窗——同键第二行可归档
	e2 := eventbus.Event{Type: "DocApplied", Source: "apply", Involved: "Agent/global/probe",
		DedupKey: "apply:Agent/global/probe:1", Level: "info"}
	if err := bus.Publish(&e2); err != nil {
		t.Fatal(err)
	}
	e3 := eventbus.Event{Type: "DocApplied", Source: "apply", Involved: "Agent/global/other",
		DedupKey: "apply:Agent/global/other:1", Level: "info"}
	if err := bus.Publish(&e3); err != nil {
		t.Fatal(err)
	}
	// 等 run 投影落环（hub 订阅 goroutine 异步）。
	deadline := time.After(2 * time.Second)
	for len(runs.State().Runs) == 0 {
		select {
		case <-deadline:
			t.Fatal("run 投影未在时限内落环")
		case <-time.After(5 * time.Millisecond):
		}
	}

	code, body := get(t, ts, "/v1/events/"+e1.ID)
	if code != http.StatusOK {
		t.Fatalf("详情应 200，得 %d: %v", code, body)
	}
	ev, _ := body["event"].(map[string]any)
	if ev == nil || ev["eventId"] != e1.ID || ev["dedupKey"] != e1.DedupKey || ev["type"] != "DocApplied" {
		t.Fatalf("event 直投应与发布回执同源：%v", ev)
	}
	// G-08 顶层词形平铺：归档行键同源直投（console 详情屏读顶层 type/level/involved/…）。
	if body["type"] != "DocApplied" || body["involved"] != "Agent/global/probe" || body["dedupKey"] != e1.DedupKey {
		t.Fatalf("顶层词形平铺应与归档行同源：%v", body)
	}
	inv, _ := body["involvedRef"].(map[string]any)
	if inv == nil || inv["kind"] != "Agent" || inv["scope"] != "global" || inv["name"] != "probe" {
		t.Fatalf("involvedRef 键解析三段应齐：%v", inv)
	}
	runsRows, _ := body["runs"].([]any)
	if len(runsRows) != 1 {
		t.Fatalf("同 eventId 的 run 关联链应恰 1 行，得 %v", body["runs"])
	}
	// G-08 顶层便利词：run/runPhase/runStatus（最近一条命中 run 直投）。本 fixture=object 源
	// 零投影（workflows/probe 未声明 phases）——status=triggered、phase=""（词位恒在，值随
	// run 声明面；builtin 锚受理即 running/第二相位——runsPayload 同口径）。
	if body["run"].(float64) != 1 || body["runStatus"] != "triggered" {
		t.Fatalf("run 便利词应直投最近命中 run：%v/%v", body["run"], body["runStatus"])
	}
	if _, ok := body["runPhase"].(string); !ok {
		t.Fatalf("runPhase 应为字符串词位（object 源零投影=空串），得 %v", body["runPhase"])
	}
	if body["wake"] != nil || body["session"] != nil || body["agentSession"] != nil {
		t.Fatalf("未接线面/无链词位应 null 不编造：wake=%v session=%v agentSession=%v",
			body["wake"], body["session"], body["agentSession"])
	}
	follows, _ := body["followUps"].([]any)
	if len(follows) != 1 || follows[0].(map[string]any)["eventId"] != e2.ID {
		t.Fatalf("同 dedupKey 后续事件应恰 e2（本行剔除/异键不入），得 %v", body["followUps"])
	}

	// 404=不在归档。
	code, body = get(t, ts, "/v1/events/evt-missing")
	if code != http.StatusNotFound {
		t.Fatalf("不在归档应 404，得 %d: %v", code, body)
	}
}

// TestEventDetailFaceAbsent 事件归档 face 缺席=503（防御面同 GET /v1/events）。
func TestEventDetailFaceAbsent(t *testing.T) {
	ts := httptest.NewServer(New(testScheme(t), nil, nil, nil, nil, nil, nil, nil, nil))
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/v1/events/evt-x")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("face 缺席应 503，得 %d", resp.StatusCode)
	}
}

// TestLoopCurrentFace 无 wake 数据=unknown 全 null；有 LoopRequested 唤醒投影=running
// 直投（seq/startedAt/eventId/wake/agent 行同源）；run 相位=runs face 环内回指（EventID
// 同源优先）；agentSession=run involved→Task status.agentSession 回填链（G-08）。
func TestLoopCurrentFace(t *testing.T) {
	sch := testScheme(t)
	// 两态之一：无任何 wake 数据（wakes face 接线但环空）→ unknown + 字段 null。
	emptyWakes := workflow.NewWakeLog(nil)
	ts := httptest.NewServer(New(sch, nil, nil, nil, nil, nil, nil, emptyWakes, nil))
	code, body := get(t, ts, "/v1/loop/current")
	if code != http.StatusOK {
		t.Fatalf("loop/current 应 200，得 %d", code)
	}
	if body["status"] != "unknown" {
		t.Fatalf("无 wake 数据应 unknown，得 %v", body["status"])
	}
	for _, key := range []string{"seq", "startedAt", "eventId", "wake", "session", "agentSession", "run", "phase", "agent"} {
		if body[key] != nil {
			t.Fatalf("无数据时 %q 应 null 不编造，得 %v", key, body[key])
		}
	}

	// 两态之二：LoopRequested 经 dispatch 表双行（root-agent 唤醒 + 对账 workflow run）
	// 推进 → running 直投；run 相位=EventID 同源回指；involved=Task 对象带 agentSession
	// → agentSession 回填链命中。
	opsDB, err := opslog.OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer opsDB.Close()
	// 文档库：Task 对象（status.agentSession=回填链目标——真实回填事实读面直投）。
	taskSch, err := scheme.Load([]byte(`apiVersion: core/v1
kind: Scheme
metadata: {name: loop-test, scope: builtin}
spec:
  version: core/v1
  kinds:
    - {kind: Scheme, builtin: true, storage: core/scheme.yaml}
    - {kind: Agent, builtin: true, storage: global/agents/<name>/agent.yaml}
    - {kind: Task, builtin: true, storage: global/tasks/<name>/task.yaml}
`))
	if err != nil {
		t.Fatal(err)
	}
	repo, err := repository.OpenSQLite(filepath.Join(t.TempDir(), "docs.db"), taskSch, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	if _, _, err := repo.Write(context.Background(), &repository.Object{Kind: "Task", Scope: "global", Name: "loop-task",
		Content: []byte("apiVersion: core/v1\nkind: Task\nmetadata:\n  name: loop-task\n  scope: global\nspec:\n  title: loop probe\nstatus:\n  phase: running\n  agentSession: dsh-loop-7\n")}, "fm", ""); err != nil {
		t.Fatal(err)
	}

	bus := eventbus.NewBus(&eventbus.Options{Ops: opsDB, DedupWindow: time.Millisecond})
	hub := workflow.NewTriggerHub(nil)
	wakes := workflow.NewWakeLog(nil)
	runs := workflow.NewRunLog(nil)
	hub.BindAgents(wakes)
	hub.BindRuns(runs)
	hub.WireExecutor(workflow.ExecAgent, workflow.ExecWorkflow) // 双面接线（唤醒+run 同一事件两行绑定）
	// 同一事件 route 主路（对账 workflow run——run 只认 route 执行面）+ cc 旁路
	//（root-agent 唤醒——agent 面双路收编）双面推进——wake 与 run 同 EventID
	//（loop/current EventID 同源回指的真实派发形态）。
	stop := hub.AttachDispatch(bus, []scheme.DispatchRule{
		{On: "LoopRequested", For: "any", Dispatch: "workflows/" + workflow.ReconcileWorkflow, CC: "root-agent"},
	})
	defer stop()
	ts2 := httptest.NewServer(New(taskSch, repo, nil, nil, hub, runs, nil, wakes, nil))
	defer ts2.Close()
	e := eventbus.Event{Type: "LoopRequested", Source: "human", Involved: "Task/global/loop-task",
		DedupKey: "cli-loop-test", Level: "info"}
	if err := bus.Publish(&e); err != nil {
		t.Fatal(err)
	}
	deadline := time.After(2 * time.Second)
	for len(wakes.State().Wakes) == 0 || len(runs.State().Runs) == 0 {
		select {
		case <-deadline:
			t.Fatal("唤醒/run 投影未在时限内落环")
		case <-time.After(5 * time.Millisecond):
		}
	}
	code, body = get(t, ts2, "/v1/loop/current")
	if code != http.StatusOK {
		t.Fatalf("loop/current 应 200，得 %d: %v", code, body)
	}
	if body["status"] != "running" {
		t.Fatalf("有 LoopRequested 唤醒投影应 running，得 %v", body["status"])
	}
	if body["eventId"] != e.ID {
		t.Fatalf("eventId 应与发布回执同源：%v vs %s", body["eventId"], e.ID)
	}
	if body["seq"].(float64) != 1 {
		t.Fatalf("seq 应=唤醒序号 1，得 %v", body["seq"])
	}
	wk, _ := body["wake"].(map[string]any)
	if wk == nil || wk["agent"] != "root" || wk["type"] != "LoopRequested" {
		t.Fatalf("wake 行直投应含 agent 锚与类型：%v", wk)
	}
	if body["agent"] != "root" {
		t.Fatalf("agent 顶层词应=wake agent 锚直投，得 %v", body["agent"])
	}
	// run 相位（G-08）：EventID 同源回指——run 行/顶层 phase 词非 null。
	runRow, _ := body["run"].(map[string]any)
	if runRow == nil || runRow["workflow"] != workflow.ReconcileWorkflow {
		t.Fatalf("run 应回指 EventID 同源 run（对账锚），得 %v", body["run"])
	}
	if body["phase"] == nil || body["phase"] == "" {
		t.Fatalf("phase 顶层词应为 run 现相位，得 %v", body["phase"])
	}
	// agentSession 回填链（G-08）：run involved=Task 对象 → status.agentSession 直投。
	if body["agentSession"] != "dsh-loop-7" {
		t.Fatalf("agentSession 应自 involved Task 对象回填，得 %v", body["agentSession"])
	}
	// session=dsh 会话 id 恒 null（真实拉起 deferred 不冒充）。
	if body["session"] != nil {
		t.Fatalf("dsh 会话 id 恒 null（真实拉起 deferred 不冒充），得 %v", body["session"])
	}
}
