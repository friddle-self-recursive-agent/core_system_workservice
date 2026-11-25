// inbox_face_test.go —— v0.56 R-10 通知中心收件箱面单测（console-adoption.md §1.25：
// 四类待办计数可查∧条目深链在册∧face 缺席=503∧心跳来源防伪=判定面自续租禁入）。
package apiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/flowbind"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// inboxFixture 最小 scheme（BuildInbox 组装只读行集——kind 注册面复用既有 fixture 惯例）。
const inboxFixture = `apiVersion: core/v1
kind: Scheme
metadata: {name: inbox-test, scope: builtin}
spec:
  version: core/v1
  kinds:
    - {kind: Workflow, builtin: true, storage: global/workflows/<name>.yaml}
  eventTypes:
    - {type: DocApplied, source: core, defaultLevel: info, route: owner-workflow}
`

func newInboxServer(t *testing.T) (*httptest.Server, *flowbind.SyncApprovals, *opslog.SQLiteStore) {
	t.Helper()
	sch, err := scheme.Load([]byte(inboxFixture))
	if err != nil {
		t.Fatal(err)
	}
	log := flowbind.NewApprovalLog(nil)
	sy := &flowbind.SyncApprovals{Log: log}
	ev, err := opslog.OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	if err := ev.AppendEvent(opslog.EventLog{Ts: time.Now(), EventID: "ev-1", Type: "NeedsHuman", Involved: "Proposal/global/p1", Severity: "error"}); err != nil {
		t.Fatal(err)
	}
	if err := ev.AppendEvent(opslog.EventLog{Ts: time.Now(), EventID: "ev-2", Type: "ApprovalNeeded", Involved: "ConnectorOps/global/wf-a", Severity: "warn"}); err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(New(sch, nil, nil, nil, nil, nil, nil, nil, nil,
		WithEventLog(ev), WithInboxFace(sy)))
	t.Cleanup(ts.Close)
	return ts, sy, ev
}

func TestInboxFaceProjection(t *testing.T) {
	ts, sy, _ := newInboxServer(t)
	// 派单受理→挂批准投影（autoLevel=ask）。
	dl := flowbind.NewDispatchLog(nil)
	dl.Approvals = sy
	dl.Dispatch(flowbind.DispatchRecord{Kind: flowbind.DispatchKindWorkflow, Workflow: "wf-a",
		Scope: "global", Connector: "wf-a", RV: 1, At: time.Now()})
	res, err := http.Get(ts.URL + "/v1/inbox")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("GET /v1/inbox=%d（期望 200）", res.StatusCode)
	}
	var body struct {
		Count    int            `json:"count"`
		Classes  map[string]int `json:"classes"`
		Items    []struct {
			Class   string `json:"class"`
			EventID string `json:"eventId"`
			Link    string `json:"link"`
		} `json:"items"`
		Approvals struct {
			Pending int `json:"pending"`
		} `json:"approvals"`
		Watchdog struct {
			Pending     int  `json:"pending"`
			ApproveIdle bool `json:"approveIdle"`
		} `json:"watchdog"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	if body.Watchdog.Pending != 1 || body.Watchdog.ApproveIdle {
		t.Fatalf("watchdog 投影=%+v（期望 pending=1 非空闲）", body.Watchdog)
	}
	if body.Approvals.Pending != 1 {
		t.Fatalf("批准挂起=%d", body.Approvals.Pending)
	}
	// 四类词形至少 NeedsHuman/ApprovalNeeded 在册（机械判定=读面现读自分类）。
	if body.Classes["NeedsHuman"] < 1 || body.Classes["ApprovalNeeded"] < 1 {
		t.Fatalf("分类计数=%v", body.Classes)
	}
	for _, it := range body.Items {
		if !strings.HasPrefix(it.Link, "#/") {
			t.Fatalf("行深链非法: %q", it.Link)
		}
	}
}

func TestInboxFaceAbsent(t *testing.T) {
	sch, err := scheme.Load([]byte(inboxFixture))
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(New(sch, nil, nil, nil, nil, nil, nil, nil, nil))
	defer ts.Close()
	res, _ := http.Get(ts.URL + "/v1/inbox")
	defer res.Body.Close()
	if res.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("face 缺席=%d（期望 503）", res.StatusCode)
	}
}

// 心跳防伪：source=watchdog 不可经注册面（判定面自续租禁入——v0.56）。
func TestHeartbeatWatchdogSourceRejected(t *testing.T) {
	sch, err := scheme.Load([]byte(inboxFixture))
	if err != nil {
		t.Fatal(err)
	}
	live := &stubLive{}
	ts := httptest.NewServer(New(sch, nil, nil, nil, nil, nil, nil, nil, nil, WithSessionLiveness(live)))
	defer ts.Close()
	res, _ := http.Post(ts.URL+"/v1/sessions/heartbeat", "application/json",
		strings.NewReader(`{"sessionId":"s1","source":"watchdog"}`))
	defer res.Body.Close()
	if res.StatusCode != http.StatusBadRequest {
		t.Fatalf("watchdog 来源心跳=%d（期望 400 禁入）", res.StatusCode)
	}
	if len(live.beats) != 0 {
		t.Fatalf("watchdog 来源心跳入账=%d（期望 0）", len(live.beats))
	}
}
