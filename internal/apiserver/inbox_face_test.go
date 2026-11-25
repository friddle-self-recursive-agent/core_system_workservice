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
		Count   int            `json:"count"`
		Classes map[string]int `json:"classes"`
		Items   []struct {
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

// L59-03 spawn 接线声明投影面（v0.57——workflow-runtime.md §11.7 拉起投影面末段）：派单行
// spawn 键恒在（代际判别器）；零值=`{declared:false,wired:false}` 诚实缺席词形；approved
// 批准投影 → `{declared:true, wired:false, target:apply-workflow}`（接线声明≠拉起——红线 a）。
func TestDispatchFaceSpawnDeclaration(t *testing.T) {
	// 派单面须随 --allow-apply 装配（WithDispatchFace）——双面装配 server（newSpawnServer）。
	ts2, sy2, dl2 := newSpawnServer(t)
	dl2.Approvals = sy2
	dl2.Dispatch(flowbind.DispatchRecord{Kind: flowbind.DispatchKindWorkflow, Workflow: "wf-a",
		Scope: "global", Connector: "wf-a", RV: 1, At: time.Now()})
	// 批准前：spawn 键在册且零值（诚实缺席——pending 行不冒充接线）。
	res, err := http.Get(ts2.URL + "/v1/agents/dispatches")
	if err != nil {
		t.Fatal(err)
	}
	body := readDispatchBody(t, res)
	res.Body.Close()
	if body.Count != 1 {
		t.Fatalf("派单行数=%d", body.Count)
	}
	if body.Dispatches[0].Spawn.Declared || body.Dispatches[0].Spawn.Wired {
		t.Fatalf("pending 行 spawn=%+v（期望 declared=false wired=false——接线缺席词形）", body.Dispatches[0].Spawn)
	}
	if body.Dispatches[0].Approval.State != flowbind.ApprovalPending {
		t.Fatalf("approval=%+v（期望 pending）", body.Dispatches[0].Approval)
	}
	// 人批=Apply 判定（workflow:approve ok 行）→ 投影翻转 → spawn 声明在册且 wired=false。
	sy2.Sync([]opslog.ApplyLog{{Ts: time.Now(), FieldManager: flowbind.ApprovalManager,
		Subject: "console-operator", Obj: "ConnectorOps/global/wf-a", ResourceVersion: 2, Result: "ok"}})
	res2, err := http.Get(ts2.URL + "/v1/agents/dispatches")
	if err != nil {
		t.Fatal(err)
	}
	body2 := readDispatchBody(t, res2)
	res2.Body.Close()
	sp := body2.Dispatches[0].Spawn
	if !sp.Declared || sp.Wired || sp.Target != flowbind.DispatchAgent || sp.RV != 2 ||
		sp.ApprovalID != "ConnectorOps/global/wf-a" {
		t.Fatalf("approved 行 spawn=%+v（期望 declared=true wired=false target=apply-workflow rv=2 approvalId 同键）", sp)
	}
	if sp.EventID != "" {
		t.Fatalf("spawn 声明 eventId=%q（期望空——判定行不复制派单载荷，接线声明=零拉起事实）", sp.EventID)
	}
	if body2.Dispatches[0].Approval.State != flowbind.ApprovalApproved {
		t.Fatalf("approval=%+v（期望 approved）", body2.Dispatches[0].Approval)
	}
}

// readDispatchBody 派单面载荷解析（spawn 键形状断言用）。
func readDispatchBody(t *testing.T, res *http.Response) struct {
	Count      int `json:"count"`
	Dispatches []struct {
		Workflow string `json:"workflow"`
		Approval struct {
			State      string `json:"state"`
			ApprovalID string `json:"approvalId"`
		} `json:"approval"`
		Spawn struct {
			Declared   bool   `json:"declared"`
			Wired      bool   `json:"wired"`
			Target     string `json:"target"`
			ApprovalID string `json:"approvalId"`
			RV         int64  `json:"rv"`
			EventID    string `json:"eventId"`
		} `json:"spawn"`
	} `json:"dispatches"`
} {
	t.Helper()
	var body struct {
		Count      int `json:"count"`
		Dispatches []struct {
			Workflow string `json:"workflow"`
			Approval struct {
				State      string `json:"state"`
				ApprovalID string `json:"approvalId"`
			} `json:"approval"`
			Spawn struct {
				Declared   bool   `json:"declared"`
				Wired      bool   `json:"wired"`
				Target     string `json:"target"`
				ApprovalID string `json:"approvalId"`
				RV         int64  `json:"rv"`
				EventID    string `json:"eventId"`
			} `json:"spawn"`
		} `json:"dispatches"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	return body
}

// newSpawnServer 派单面+收件箱面双装配 server（spawn 声明投影面测用——WithDispatchFace 随
// --allow-apply 装配同款；返回注入的 DispatchLog=同环投影，零旁路）。
func newSpawnServer(t *testing.T) (*httptest.Server, *flowbind.SyncApprovals, *flowbind.DispatchLog) {
	t.Helper()
	sch, err := scheme.Load([]byte(inboxFixture))
	if err != nil {
		t.Fatal(err)
	}
	log := flowbind.NewApprovalLog(nil)
	sy := &flowbind.SyncApprovals{Log: log}
	dl := flowbind.NewDispatchLog(nil)
	ts := httptest.NewServer(New(sch, nil, nil, nil, nil, nil, nil, nil, nil,
		WithInboxFace(sy), WithDispatchFace(dl)))
	t.Cleanup(ts.Close)
	return ts, sy, dl
}
