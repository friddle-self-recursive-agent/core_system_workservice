// triggers_test.go —— 触发骨架查询面验收（cored-bridge.md §2.3）：
// wired 面 200 载荷形状 + 收到留痕可查；nil 面 503；查询面恒开语义（写面无关）。
package apiserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
)

func testScheme(t *testing.T) *scheme.Scheme {
	t.Helper()
	sch, err := scheme.Load([]byte(fixture))
	if err != nil {
		t.Fatalf("载入 fixture 失败: %v", err)
	}
	return sch
}

func TestTriggersFaceNilReturnsUnavailable(t *testing.T) {
	ts := httptest.NewServer(New(testScheme(t), nil, nil, nil, nil, nil))
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/v1/triggers")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("未接线应 503，得 %d", resp.StatusCode)
	}
	// 决策查询面同口径：未接线 → 503。
	resp2, err := http.Get(ts.URL + "/v1/triggers/resolve?type=DocApplied&involved=Agent/global/a")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("resolve 未接线应 503，得 %d", resp2.StatusCode)
	}
}

func TestTriggersFaceRecordsTrace(t *testing.T) {
	sch := testScheme(t)
	bus := eventbus.NewBus(nil)
	hub := workflow.NewTriggerHub(nil)
	stop := hub.Attach(bus, "DocApplied")
	defer stop()

	ts := httptest.NewServer(New(sch, nil, nil, nil, hub, nil))
	defer ts.Close()

	// 未收到前：200 空面（查询面恒开，与写面开关无关）。
	resp, err := http.Get(ts.URL + "/v1/triggers")
	if err != nil {
		t.Fatal(err)
	}
	var empty struct {
		Count    int              `json:"count"`
		Received int64            `json:"received"`
		Dropped  int64            `json:"dropped"`
		Types    []string         `json:"types"`
		Records  []map[string]any `json:"records"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&empty); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK || empty.Count != 0 || empty.Received != 0 {
		t.Fatalf("空面应 200/{count:0,received:0}，得 %d %v", resp.StatusCode, empty)
	}
	if len(empty.Types) != 1 || empty.Types[0] != "DocApplied" {
		t.Fatalf("types 应=[DocApplied]，得 %v", empty.Types)
	}

	// 模拟 apply 落面→DocApplied 发布（同源字段：id/dedupKey 由总线归一化）。
	e := eventbus.Event{
		Type: "DocApplied", Source: "apply", Involved: "Agent/global/probe",
		DedupKey: "apply:Agent/global/probe:1", Level: "info",
	}
	if err := bus.Publish(&e); err != nil {
		t.Fatal(err)
	}
	deadline := time.After(time.Second)
	var got struct {
		Count    int64            `json:"count"`
		Received int64            `json:"received"`
		Records  []map[string]any `json:"records"`
	}
	for {
		resp, err := http.Get(ts.URL + "/v1/triggers")
		if err != nil {
			t.Fatal(err)
		}
		err = json.NewDecoder(resp.Body).Decode(&got)
		resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if got.Received >= 1 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("留痕未在时限内可查（主干第一段判据失败）")
		case <-time.After(10 * time.Millisecond):
		}
	}
	if got.Count != 1 || got.Received != 1 {
		t.Fatalf("应恰 1 条留痕：count=%d received=%d", got.Count, got.Received)
	}
	r := got.Records[0]
	if r["eventId"] != e.ID || r["dedupKey"] != e.DedupKey || r["type"] != "DocApplied" {
		t.Fatalf("留痕与发布回执不同源：%v vs {id:%s dedup:%s}", r, e.ID, e.DedupKey)
	}
	if r["seq"].(float64) != 1 || r["source"] != "apply" || r["involved"] != "Agent/global/probe" {
		t.Fatalf("留痕 seq/source/involved 不符: %v", r)
	}
	if fmt.Sprintf("%v", r["receivedAt"]) == "" || fmt.Sprintf("%v", r["at"]) == "" {
		t.Fatalf("at/receivedAt 应为 RFC3339 字符串: %v", r)
	}
	// 裸 Attach（未挂 dispatch 表）→ route/for 留空（路由标注只在表驱动挂接下出现）。
	if r["route"] != "" || r["for"] != "" {
		t.Fatalf("裸 Attach 留痕不应带 route/for: %v", r)
	}
}

// TestTriggersFaceTableDrivenWiring 表驱动挂接的端点投影：types=表推导集合（"*"=通配坍缩）、
// records 携命中行 route/for、全表未命中=dead-letter（cored-bridge.md §2.3 v0.12）。
func TestTriggersFaceTableDrivenWiring(t *testing.T) {
	sch := testScheme(t)
	bus := eventbus.NewBus(nil)
	hub := workflow.NewTriggerHub(nil)
	defer hub.AttachDispatch(bus, []scheme.DispatchRule{
		{On: "DocApplied", For: "owner:*", Dispatch: "owner-workflow"},
		{On: "*", For: "scope:builtin/*", Dispatch: "workflows/builtin-root"},
		{On: "*", For: "any", Dispatch: "dead-letter"},
	})()

	ts := httptest.NewServer(New(sch, nil, nil, nil, hub, nil))
	defer ts.Close()

	e := eventbus.Event{
		Type: "DocApplied", Source: "apply", Involved: "Agent/global/probe",
		DedupKey: "apply:Agent/global/probe:1", Level: "info",
	}
	if err := bus.Publish(&e); err != nil {
		t.Fatal(err)
	}
	deadline := time.After(time.Second)
	var got struct {
		Types   []string         `json:"types"`
		Records []map[string]any `json:"records"`
	}
	for {
		resp, err := http.Get(ts.URL + "/v1/triggers")
		if err != nil {
			t.Fatal(err)
		}
		err = json.NewDecoder(resp.Body).Decode(&got)
		resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if len(got.Records) >= 1 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("表驱动留痕未在时限内可查")
		case <-time.After(10 * time.Millisecond):
		}
	}
	if len(got.Types) != 1 || got.Types[0] != "*" {
		t.Fatalf("types 应坍缩为 [*]（判别器：≠固定 [DocApplied]），得 %v", got.Types)
	}
	r := got.Records[0]
	if r["route"] != "owner-workflow" || r["for"] != "owner:*" {
		t.Fatalf("留痕应携命中行投影 route=owner-workflow for=owner:*，得 %v", r)
	}
	// v0.13 决策依据可查：owner 解析值随留痕投影（involved scope 根 global → owner=global）；
	// cc 空串=命中行无旁路目标。
	if r["owner"] != "global" || r["cc"] != "" {
		t.Fatalf("留痕应携 owner 解析值与 cc 投影，得 %v", r)
	}
}

// TestTriggersResolveFace 派发决策查询面（cored-bridge.md §2.3 v0.13 引擎主体第一段）：
// 200 决策投影 {type,involved,matched,route,cc,for,owner,row}；缺 type → 400；
// 纯决策零副作用（resolve 后 /v1/triggers received 不动）。
func TestTriggersResolveFace(t *testing.T) {
	sch := testScheme(t)
	bus := eventbus.NewBus(nil)
	hub := workflow.NewTriggerHub(nil)
	defer hub.AttachDispatch(bus, []scheme.DispatchRule{
		{On: "DocApplied", For: "owner:*", Dispatch: "owner-workflow"},
		{On: "HeartbeatLost", For: "owner:agent", Dispatch: "parent-agent", CC: "notification/*"},
		{On: "*", For: "any", Dispatch: "dead-letter"},
	})()
	ts := httptest.NewServer(New(sch, nil, nil, nil, hub, nil))
	defer ts.Close()

	var d struct {
		Type, Involved string
		Matched        bool
		Route, CC      string
		For, Owner     string
		Row            int
	}
	get := func(url string) int {
		t.Helper()
		resp, err := http.Get(ts.URL + url)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		if resp.StatusCode == http.StatusOK {
			if err := json.NewDecoder(resp.Body).Decode(&d); err != nil {
				t.Fatal(err)
			}
		}
		return resp.StatusCode
	}

	if c := get("/v1/triggers/resolve?type=HeartbeatLost&involved=ServiceInstance/systems/agent/ci"); c != 200 {
		t.Fatalf("resolve 应 200，得 %d", c)
	}
	if !d.Matched || d.Route != "parent-agent" || d.CC != "notification/*" ||
		d.For != "owner:agent" || d.Owner != "agent" || d.Row != 1 || d.Type != "HeartbeatLost" {
		t.Fatalf("owner 命中决策投影不符: %+v", d)
	}
	if c := get("/v1/triggers/resolve?type=HeartbeatLost&involved=Agent/global/a1"); c != 200 {
		t.Fatalf("resolve 应 200，得 %d", c)
	}
	if !d.Matched || d.Route != "dead-letter" || d.Row != 2 || d.Owner != "global" {
		t.Fatalf("owner 不命中应顺延兜底行: %+v", d)
	}
	if c := get("/v1/triggers/resolve?type=DocApplied"); c != 200 {
		t.Fatalf("involved 缺省应照常决策，得 %d", c)
	}
	if !d.Matched || d.Route != "owner-workflow" || d.Owner != "" || d.Row != 0 {
		t.Fatalf("空 involved 决策不符（owner 解析为空）: %+v", d)
	}
	if c := get("/v1/triggers/resolve"); c != http.StatusBadRequest {
		t.Fatalf("缺 type 应 400，得 %d", c)
	}
	// 纯决策面零副作用：上述 resolve 不产生任何留痕。
	resp, err := http.Get(ts.URL + "/v1/triggers")
	if err != nil {
		t.Fatal(err)
	}
	var st struct {
		Received int64            `json:"received"`
		Records  []map[string]any `json:"records"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&st); err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if st.Received != 0 || len(st.Records) != 0 {
		t.Fatalf("决策面不得产生留痕：received=%d records=%d", st.Received, len(st.Records))
	}
}

// binding 投影读取辅助：execution 区 {route:{executor,target,wired}, cc:{…}}。
func binding(m map[string]any, k, field string) any {
	b, _ := m["execution"].(map[string]any)[k].(map[string]any)
	return b[field]
}

// TestTriggersExecutionProjection 派发执行投影面（cored-bridge.md §2.3 v0.14 引擎主体第二段）：
// resolve 增 execution={route:{executor,target,wired}, cc:{…}}——owner-workflow→workflow/owner、
// parent-agent+cc→agent/parent+notification/*、兜底行 dead-letter→none；装配层未登记执行面
// → wired=false（诚实投影），WireExecutor 登记后端点翻转（真实拉起落地轮的机械判别器）；
// 未知目标=unbound 缺口可见；留痕携执行计划快照。
func TestTriggersExecutionProjection(t *testing.T) {
	sch := testScheme(t)
	bus := eventbus.NewBus(nil)
	hub := workflow.NewTriggerHub(nil)
	defer hub.AttachDispatch(bus, []scheme.DispatchRule{
		{On: "DocApplied", For: "owner:*", Dispatch: "owner-workflow"},
		{On: "HeartbeatLost", For: "owner:agent", Dispatch: "parent-agent", CC: "notification/*"},
		{On: "Escalate", For: "any", Dispatch: "mystery-face"}, // 未知目标族=unbound 可见
		{On: "*", For: "any", Dispatch: "dead-letter"},
	})()
	ts := httptest.NewServer(New(sch, nil, nil, nil, hub, nil))
	defer ts.Close()

	get := func(url string) (int, map[string]any) {
		t.Helper()
		resp, err := http.Get(ts.URL + url)
		if err != nil {
			t.Fatal(err)
		}
		defer resp.Body.Close()
		var m map[string]any
		if resp.StatusCode == http.StatusOK {
			if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
				t.Fatal(err)
			}
		}
		return resp.StatusCode, m
	}

	// owner-workflow 行：workflow/owner 绑定，无 cc=零值绑定；装配层未登记 → wired=false。
	c, m := get("/v1/triggers/resolve?type=DocApplied&involved=Agent/global/a")
	if c != 200 {
		t.Fatalf("resolve 应 200，得 %d", c)
	}
	if binding(m, "route", "executor") != "workflow" || binding(m, "route", "target") != "owner" {
		t.Fatalf("execution.route 应=workflow/owner: %v", m["execution"])
	}
	if binding(m, "route", "wired") != false {
		t.Fatalf("装配层未登记执行面应 wired=false: %v", m["execution"])
	}
	if binding(m, "cc", "executor") != "" {
		t.Fatalf("无 cc 应零值绑定: %v", m["execution"])
	}

	// parent-agent 行 + cc：agent/parent + notification/* 双绑定。
	if c, m = get("/v1/triggers/resolve?type=HeartbeatLost&involved=ServiceInstance/systems/agent/ci"); c != 200 {
		t.Fatalf("resolve 应 200，得 %d", c)
	}
	if binding(m, "route", "executor") != "agent" || binding(m, "route", "target") != "parent" ||
		binding(m, "cc", "executor") != "notification" || binding(m, "cc", "target") != "*" {
		t.Fatalf("双绑定投影不符: %v", m["execution"])
	}

	// 未知目标族：unbound 原样投影（缺口可见，不静默吞并）。
	if c, m = get("/v1/triggers/resolve?type=Escalate&involved=Agent/global/a"); c != 200 {
		t.Fatalf("resolve 应 200，得 %d", c)
	}
	if binding(m, "route", "executor") != "unbound" || binding(m, "route", "target") != "mystery-face" {
		t.Fatalf("未知目标应 unbound 原样投影: %v", m["execution"])
	}

	// 兜底行 dead-letter → none 绑定。
	if c, m = get("/v1/triggers/resolve?type=UnknownType&involved=Agent/global/a"); c != 200 {
		t.Fatalf("resolve 应 200，得 %d", c)
	}
	if binding(m, "route", "executor") != "none" {
		t.Fatalf("dead-letter 行应 none 绑定: %v", m["execution"])
	}

	// 接线态翻转经端点可见（真实拉起落地轮的机械判别器）：登记 workflow 面后同查询 wired=true。
	hub.WireExecutor(workflow.ExecWorkflow)
	if c, m = get("/v1/triggers/resolve?type=DocApplied&involved=Agent/global/a"); c != 200 {
		t.Fatalf("resolve 应 200，得 %d", c)
	}
	if binding(m, "route", "wired") != true {
		t.Fatalf("WireExecutor 登记后端点应投影 wired=true: %v", m["execution"])
	}

	// 留痕携执行计划快照（收到时刻的绑定投影随 records 可查）。
	e := eventbus.Event{
		Type: "DocApplied", Source: "apply", Involved: "Agent/global/probe",
		DedupKey: "apply:Agent/global/probe:1", Level: "info",
	}
	if err := bus.Publish(&e); err != nil {
		t.Fatal(err)
	}
	deadline := time.After(time.Second)
	var recs []map[string]any
	for {
		resp, err := http.Get(ts.URL + "/v1/triggers")
		if err != nil {
			t.Fatal(err)
		}
		var st struct {
			Received int64            `json:"received"`
			Records  []map[string]any `json:"records"`
		}
		err = json.NewDecoder(resp.Body).Decode(&st)
		resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if st.Received >= 1 {
			recs = st.Records
			break
		}
		select {
		case <-deadline:
			t.Fatal("留痕未在时限内可查")
		case <-time.After(10 * time.Millisecond):
		}
	}
	ex, _ := recs[0]["execution"].(map[string]any)
	route, _ := ex["route"].(map[string]any)
	if route["executor"] != "workflow" || route["target"] != "owner" || route["wired"] != true {
		t.Fatalf("留痕应携执行计划快照 workflow/owner/wired=true: %v", recs[0]["execution"])
	}
}

// TestWorkflowRunsFace 对账 run 投影面（cored-bridge.md §2.3 v0.15 引擎主体第三段——
// 真实拉起第一子段）：nil 面 503；接线（BindRuns+WireExecutor）后 run 可查——载荷形状
// {count, started, dropped, runs:[{run, workflow, specSource, phases, phase, status,
// triggerSeq, eventId, route, owner, target, startedAt}]}。
func TestWorkflowRunsFace(t *testing.T) {
	// 未接线 → 503（查询面语义与 triggers 同口径）。
	ts := httptest.NewServer(New(testScheme(t), nil, nil, nil, nil, nil))
	resp, err := http.Get(ts.URL + "/v1/workflows/runs")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	ts.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("run 面未接线应 503，得 %d", resp.StatusCode)
	}

	// 接线后：DocApplied（owner-workflow）→ run 可查且溯源字段同源。
	sch := testScheme(t)
	bus := eventbus.NewBus(nil)
	hub := workflow.NewTriggerHub(nil)
	runs := workflow.NewRunLog(nil)
	hub.BindRuns(runs)
	hub.WireExecutor(workflow.ExecWorkflow)
	defer hub.AttachDispatch(bus, []scheme.DispatchRule{
		{On: "DocApplied", For: "owner:*", Dispatch: "owner-workflow"},
	})()
	ts2 := httptest.NewServer(New(sch, nil, nil, nil, hub, runs))
	defer ts2.Close()
	e := eventbus.Event{
		Type: "DocApplied", Source: "apply", Involved: "Agent/global/probe",
		DedupKey: "apply:Agent/global/probe:9", Level: "info",
	}
	if err := bus.Publish(&e); err != nil {
		t.Fatal(err)
	}
	deadline := time.After(time.Second)
	var got struct {
		Count   int              `json:"count"`
		Started int64            `json:"started"`
		Dropped int64            `json:"dropped"`
		Runs    []map[string]any `json:"runs"`
	}
	for {
		resp, err := http.Get(ts2.URL + "/v1/workflows/runs")
		if err != nil {
			t.Fatal(err)
		}
		err = json.NewDecoder(resp.Body).Decode(&got)
		resp.Body.Close()
		if err != nil {
			t.Fatal(err)
		}
		if got.Count >= 1 {
			break
		}
		select {
		case <-deadline:
			t.Fatal("run 未在时限内可查（真实拉起第一子段判据失败）")
		case <-time.After(10 * time.Millisecond):
		}
	}
	if got.Started != 1 || got.Dropped != 0 || len(got.Runs) != 1 {
		t.Fatalf("run 面口径应 started=1 dropped=0 count=1: %+v", got)
	}
	r := got.Runs[0]
	if r["workflow"] != "core-workflow" || r["specSource"] != "builtin" || r["status"] != "triggered" ||
		r["phase"] != "Idle" || r["route"] != "owner-workflow" || r["owner"] != "global" ||
		r["target"] != "owner" || r["eventId"] != e.ID {
		t.Fatalf("run 载荷不符: %v", r)
	}
	if r["triggerSeq"].(float64) != 1 {
		t.Fatalf("run 应携来源留痕 seq=1: %v", r)
	}
	phases, ok := r["phases"].([]any)
	if !ok || len(phases) != 8 || phases[0] != "Idle" {
		t.Fatalf("run 应携声明相位序列: %v", r["phases"])
	}
}
