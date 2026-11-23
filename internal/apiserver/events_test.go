// events_test.go —— 事件注入面 HTTP 验收（v0.20 发布面首段，concepts/event.md 规则 8）：
// 写面门控 405（face 缺席代际判别器）、回执形状与 source 强制、类型/level 校验闸门、
// 发布权限域与 apply 写权限域分立、同主干六面投影（留痕/run/送达/唤醒 live）。
package apiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventpub"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
)

// evFixture 注入面 fixture：注册表三类（含 defaultLevel 分诊）+ 权威表同款三行
// （Escalate→root-agent cc notification/*、HeartbeatLost owner:agent→parent-agent、
// DocApplied owner:*→owner-workflow）——六面 live 的最小表驱动。
const evFixture = `apiVersion: core/v1
kind: Scheme
metadata: {name: default, scope: builtin}
spec:
  version: core/v1
  kinds:
    - {kind: Scheme, builtin: true, storage: core/scheme.yaml}
    - {kind: Agent, builtin: true, storage: global/agents/<name>/agent.yaml}
  eventTypes:
    - {type: DocApplied, source: core, defaultLevel: info, route: owner-workflow}
    - {type: Escalate, source: engine, defaultLevel: error, route: root+notification}
    - {type: HeartbeatLost, source: core, defaultLevel: error, route: parent-agent}
  dispatch:
    - {on: Escalate, for: any, dispatch: root-agent, cc: notification/*}
    - {on: HeartbeatLost, for: "owner:agent", dispatch: parent-agent, cc: notification/*}
    - {on: DocApplied, for: "owner:*", dispatch: owner-workflow}
    - {on: "*", for: any, dispatch: dead-letter}
`

const evPublisherRole = `apiVersion: core/v1
kind: Role
metadata: {name: ev-publisher, scope: global}
spec:
  permissions:
  - {scope: events, actions: [publish]}
`

// evServer 组装注入面实例：真实总线+触发骨架+三投影 sink（六面 live 的进程内合证）。
func evServer(t *testing.T, perm *permission.System) (*httptest.Server, *workflow.TriggerHub) {
	t.Helper()
	sch, err := scheme.Load([]byte(evFixture))
	if err != nil {
		t.Fatalf("载入 fixture 失败: %v", err)
	}
	bus := eventbus.NewBus(nil)
	hub := workflow.NewTriggerHub(nil)
	runs := workflow.NewRunLog(nil)
	deliveries := workflow.NewDeliveryLog(&workflow.DeliveryOptions{})
	wakes := workflow.NewWakeLog(nil)
	defer deliveries.Close()
	hub.BindRuns(runs)
	hub.BindDeliveries(deliveries)
	hub.BindAgents(wakes)
	hub.WireExecutor(workflow.ExecWorkflow, workflow.ExecNotification, workflow.ExecAgent)
	stop := hub.AttachDispatch(bus, sch.Spec.Dispatch)
	t.Cleanup(stop)
	pub := &eventpub.Publisher{Sch: sch, Bus: bus, Perm: perm}
	ts := httptest.NewServer(New(sch, nil, nil, nil, hub, runs, deliveries, wakes, pub))
	t.Cleanup(ts.Close)
	return ts, hub
}

func postEvents(t *testing.T, ts *httptest.Server, subject string, body string) (int, map[string]any) {
	t.Helper()
	req, err := http.NewRequest(http.MethodPost, ts.URL+"/v1/events", strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(SubjectHeader, subject)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("响应体不可解析: %v", err)
	}
	return resp.StatusCode, out
}

func TestEventsFaceAbsentReturns405(t *testing.T) {
	// face 缺席=405（兜底 handler 口径——旧构建代际判别器；与 apply 写面同门控同语义）。
	ts := httptest.NewServer(New(testScheme(t), nil, nil, nil, nil, nil, nil, nil, nil))
	defer ts.Close()
	code, _ := postEvents(t, ts, authHeader("probe", "global/admin"), `{"type":"Escalate"}`)
	if code != http.StatusMethodNotAllowed {
		t.Fatalf("未开启注入面应 405，得 %d", code)
	}
}

func TestEventsInjectReceiptShape(t *testing.T) {
	ts, _ := evServer(t, fixturePerm(t))
	code, out := postEvents(t, ts, authHeader("probe", "global/admin"),
		`{"type":"Escalate","involved":"Agent/global/esc-1"}`)
	if code != http.StatusOK {
		t.Fatalf("注入应 200，得 %d: %v", code, out)
	}
	if out["source"] != eventpub.SourceHuman {
		t.Errorf("source=%v, 要服务侧判定恒 human", out["source"])
	}
	if out["level"] != "error" {
		t.Errorf("level=%v, 要注册表 defaultLevel 投影（Escalate→error）", out["level"])
	}
	if out["type"] != "Escalate" || out["involved"] != "Agent/global/esc-1" {
		t.Errorf("type/involved 透传不符: %v", out)
	}
	if id, _ := out["id"].(string); id == "" {
		t.Errorf("回执缺 id")
	}
	if dk, _ := out["dedupKey"].(string); dk == "" {
		t.Errorf("回执缺 dedupKey（空入参=ID 兜底，仍应有值）")
	}
}

func TestEventsInjectValidationGates(t *testing.T) {
	ts, _ := evServer(t, fixturePerm(t))
	admin := authHeader("probe", "global/admin")
	if code, _ := postEvents(t, ts, admin, `{}`); code != http.StatusBadRequest {
		t.Errorf("缺 type 应 400，得 %d", code)
	}
	code, out := postEvents(t, ts, admin, `{"type":"Mystery"}`)
	if code != http.StatusNotFound {
		t.Errorf("未注册类型应 404，得 %d", code)
	}
	if !strings.Contains(out["error"].(string), "未注册") {
		t.Errorf("错误文本缺类型注册闸门语义: %v", out)
	}
	if code, _ := postEvents(t, ts, admin, `{"type":"Escalate","level":"loud"}`); code != http.StatusBadRequest {
		t.Errorf("越词 level 应 400，得 %d", code)
	}
	code, out = postEvents(t, ts, admin, `{"type":"Escalate","source":"connector"}`)
	if code != http.StatusBadRequest {
		t.Errorf("载荷声明 source=connector 应 400，得 %d", code)
	}
	if !strings.Contains(out["error"].(string), "服务侧判定") {
		t.Errorf("错误文本缺来源强制语义: %v", out)
	}
	if code, _ := postEvents(t, ts, admin, `{"type":"Escalate","source":"human"}`); code != http.StatusOK {
		t.Errorf("source=human 透传应放行，得 %d", code)
	}
}

func TestEventsInjectPermissionDomainSeparate(t *testing.T) {
	perm := permission.System{}
	if err := perm.LoadObjects([]permission.ObjectDoc{
		{Kind: "Role", Name: "ev-publisher", Scope: "global", Content: []byte(evPublisherRole)},
		{Kind: "Role", Name: "fixture-dev", Scope: "global", Content: []byte(fixtureDevRole)},
	}); err != nil {
		t.Fatalf("装载授权面: %v", err)
	}
	ts, _ := evServer(t, &perm)
	// 匿名（无 subject 字段且无头）→ 403：请求级判定不兜底。
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/v1/events", strings.NewReader(`{"type":"Escalate"}`))
	req.Header.Set("Content-Type", "application/json")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("匿名注入应 403，得 %d", resp.StatusCode)
	}
	// 仅持对象 write 的主体 → 403（write×scope 不产生 publish×events——两域分立面）。
	code, out := postEvents(t, ts, authHeader("dev", "fixture-dev"), `{"type":"Escalate"}`)
	if code != http.StatusForbidden {
		t.Errorf("对象 write 主体注入应 403，得 %d", code)
	}
	if !strings.Contains(out["error"].(string), "publish") {
		t.Errorf("拒绝文本缺动作×域摘要: %v", out)
	}
	// 持 publish×events 的主体 → 200（发布域独立授权面）。
	if code, _ := postEvents(t, ts, authHeader("human-1", "ev-publisher"), `{"type":"Escalate"}`); code != http.StatusOK {
		t.Errorf("publish×events 主体注入应 200")
	}
}

func TestEventsInjectSixFacesLive(t *testing.T) {
	ts, hub := evServer(t, fixturePerm(t))
	admin := authHeader("probe", "global/admin")

	// ① Escalate 注入 → 唤醒面真实投影（route 主路 root）∧ 送达旁路（cc notification/*）。
	code, out := postEvents(t, ts, admin, `{"type":"Escalate","involved":"Agent/global/esc-1","dedupKey":"wf-x:quota"}`)
	if code != http.StatusOK {
		t.Fatalf("Escalate 注入应 200: %v", out)
	}
	eid, _ := out["id"].(string)
	hst := hub.State()
	if len(hst.Records) != 1 || hst.Records[0].EventID != eid || hst.Records[0].Route != "root-agent" {
		t.Fatalf("留痕面: records=%+v 要 eventId=%s route=root-agent", hst.Records, eid)
	}
	wst := wakeState(t, ts)
	if wst["count"].(float64) != 1 {
		t.Fatalf("唤醒面 count=%v, 要 1", wst["count"])
	}
	w0 := wst["wakes"].([]any)[0].(map[string]any)
	if w0["agent"] != "root" || w0["origin"] != "route" || w0["eventId"] != eid {
		t.Errorf("唤醒投影不符: %v", w0)
	}
	dst := deliveryState(t, ts)
	if dst["count"].(float64) != 1 {
		t.Fatalf("送达面 count=%v, 要 1（cc 旁路）", dst["count"])
	}
	d0 := dst["deliveries"].([]any)[0].(map[string]any)
	if d0["origin"] != "cc" || d0["channel"] != "*" || d0["status"] != workflow.WakeAccepted {
		t.Errorf("旁路送达投影不符: %v", d0)
	}
	rst := runState(t, ts)
	if rst["count"].(float64) != 0 {
		t.Errorf("agent 族事件不产生 run（执行面分野），得 %v", rst["count"])
	}

	// ② HeartbeatLost @ owner=agent → parent-agent 行命中（owner:agent 谓词 live）。
	code, _ = postEvents(t, ts, admin, `{"type":"HeartbeatLost","involved":"ServiceInstance/systems/agent/ci"}`)
	if code != http.StatusOK {
		t.Fatalf("HeartbeatLost 注入应 200")
	}
	wst = wakeState(t, ts)
	if wst["count"].(float64) != 2 {
		t.Fatalf("第二次唤醒应累计 2，得 %v", wst["count"])
	}
	w1 := wst["wakes"].([]any)[1].(map[string]any)
	if w1["agent"] != "parent" || w1["owner"] != "agent" {
		t.Errorf("parent 锚/owner 解析不符: %v", w1)
	}

	// ③ DocApplied 注入 → run 投影 live（owner:* 行→owner-workflow 对账锚）。
	code, out2 := postEvents(t, ts, admin, `{"type":"DocApplied","involved":"Workflow/systems/probe-ev/wf-1"}`)
	if code != http.StatusOK {
		t.Fatalf("DocApplied 注入应 200: %v", out2)
	}
	rst = runState(t, ts)
	if rst["count"].(float64) != 1 {
		t.Fatalf("注入事件应推进一条对账 run，得 %v", rst["count"])
	}
	r0 := rst["runs"].([]any)[0].(map[string]any)
	if r0["route"] != "owner-workflow" || r0["owner"] != "probe-ev" {
		t.Errorf("run 溯源不符: %v", r0)
	}
	if r0["eventId"] != out2["id"] {
		t.Errorf("run eventId=%v 与注入回执 id=%v 不同源", r0["eventId"], out2["id"])
	}
}

func TestEventsInjectDedupSameBackbone(t *testing.T) {
	ts, hub := evServer(t, fixturePerm(t))
	admin := authHeader("probe", "global/admin")
	// 窗口内同 DedupKey 聚合：两次注入回执照常 200，主干只投递一次（同管线去重语义）。
	if code, _ := postEvents(t, ts, admin, `{"type":"Escalate","dedupKey":"same:1"}`); code != http.StatusOK {
		t.Fatalf("首条注入应 200")
	}
	if code, _ := postEvents(t, ts, admin, `{"type":"Escalate","dedupKey":"same:1"}`); code != http.StatusOK {
		t.Fatalf("窗口内重入应 200（聚合非拒绝）")
	}
	if got := hub.State().Received; got != 1 {
		t.Errorf("主干 received=%d, 要 1（同因事件窗口内合并）", got)
	}
}

// runState/deliveryState/wakeState HTTP 载荷读取 helper（查询面取证）。
func runState(t *testing.T, ts *httptest.Server) map[string]any {
	return getJSON(t, ts.URL+"/v1/workflows/runs")
}

func deliveryState(t *testing.T, ts *httptest.Server) map[string]any {
	return getJSON(t, ts.URL+"/v1/notifications/deliveries")
}

func wakeState(t *testing.T, ts *httptest.Server) map[string]any {
	return getJSON(t, ts.URL+"/v1/agents/wakes")
}

func getJSON(t *testing.T, url string) map[string]any {
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
	return out
}
