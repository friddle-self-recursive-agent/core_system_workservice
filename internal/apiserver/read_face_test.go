// read_face_test.go —— v0.66 读面验收（R-ENG §3）：单事件详情（200 关联链/404 不在归档/
// 503 face 缺席）与 loop/current（有 wake=running 直投/无 wake=unknown 全 null）。
package apiserver

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
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
	inv, _ := body["involved"].(map[string]any)
	if inv == nil || inv["kind"] != "Agent" || inv["scope"] != "global" || inv["name"] != "probe" {
		t.Fatalf("involved 键解析三段应齐：%v", inv)
	}
	runsRows, _ := body["runs"].([]any)
	if len(runsRows) != 1 {
		t.Fatalf("同 eventId 的 run 关联链应恰 1 行，得 %v", body["runs"])
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
// 直投（seq/startedAt/eventId/wake 行同源）。
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
	for _, key := range []string{"seq", "startedAt", "eventId", "wake", "session"} {
		if body[key] != nil {
			t.Fatalf("无数据时 %q 应 null 不编造，得 %v", key, body[key])
		}
	}

	// 两态之二：LoopRequested 经 dispatch 表 root-agent 行推进唤醒投影 → running 直投。
	opsDB, err := opslog.OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer opsDB.Close()
	bus := eventbus.NewBus(&eventbus.Options{Ops: opsDB, DedupWindow: time.Millisecond})
	hub := workflow.NewTriggerHub(nil)
	wakes := workflow.NewWakeLog(nil)
	hub.BindAgents(wakes)
	hub.WireExecutor(workflow.ExecAgent)
	stop := hub.AttachDispatch(bus, []scheme.DispatchRule{{On: "LoopRequested", For: "any", Dispatch: "root-agent"}})
	defer stop()
	ts2 := httptest.NewServer(New(sch, nil, nil, nil, hub, nil, nil, wakes, nil))
	defer ts2.Close()
	e := eventbus.Event{Type: "LoopRequested", Source: "human", Involved: "Agent/global/root",
		DedupKey: "cli-loop-test", Level: "info"}
	if err := bus.Publish(&e); err != nil {
		t.Fatal(err)
	}
	deadline := time.After(2 * time.Second)
	for len(wakes.State().Wakes) == 0 {
		select {
		case <-deadline:
			t.Fatal("唤醒投影未在时限内落环")
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
	if body["session"] != nil {
		t.Fatalf("dsh 会话 id 恒 null（真实拉起 deferred 不冒充），得 %v", body["session"])
	}
}
