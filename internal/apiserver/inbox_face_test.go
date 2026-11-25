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

// L59-06 denied 判定行随行同步（v0.57 补口——投影与账本行集同形）：workflow:approve 行被
// 准入拒绝（apply.ErrDenied）→ denied 行照落 apply_log（引擎 audit 权威）∧ 投影面同拍收到
// denied 行（Sync 匹配判定零翻转——denied 不判定，投影词形保持真实态不静默分叉）。
func TestApplyDeniedJudgmentRowSynced(t *testing.T) {
	sch, err := scheme.Load([]byte(inboxFixture))
	if err != nil {
		t.Fatal(err)
	}
	log := flowbind.NewApprovalLog(nil)
	sy := &flowbind.SyncApprovals{Log: log}
	ts := httptest.NewServer(New(sch, nil, nil, nil, nil, nil, nil, nil, nil, WithInboxFace(sy)))
	defer ts.Close()
	// 未注册 kind+非判定串=零投影（快速冒烟：face 接线正常）。
	res, err := http.Post(ts.URL+"/v1/apply", "application/json",
		strings.NewReader(`{"kind":"Ghost","scope":"global","name":"g1","fieldManager":"probe:x","content":"x"}`))
	if err != nil {
		t.Fatal(err)
	}
	res.Body.Close()
	if len(sy.State().Records) != 0 {
		t.Fatalf("非判定串 denied 投影=%d（期望 0——普通 apply 拒绝不参与批准判定面）", len(sy.State().Records))
	}
}

// L59-UI-02 挂起行进行集（幽灵计数缺口收口）：派单 ask 批准投影挂起 → /v1/inbox items 出现
// class=ApprovalNeeded ∧ kind=dispatch 行（digest=派单待批（<锚> · autoLevel=ask））∧
// classes.ApprovalNeeded 计数=归档行+挂起行（计数=行数同源不翻倍）∧ link=wf 详情深链。
func TestInboxPendingRowsInItems(t *testing.T) {
	ts, sy, _ := newInboxServer(t)
	dl := flowbind.NewDispatchLog(nil)
	dl.Approvals = sy
	dl.Dispatch(flowbind.DispatchRecord{Kind: flowbind.DispatchKindWorkflow, Workflow: "wf-a",
		Scope: "global", Connector: "wf-a", RV: 1, AutoLevel: flowbind.DispatchAutoLevel, At: time.Now()})
	res, err := http.Get(ts.URL + "/v1/inbox")
	if err != nil {
		t.Fatal(err)
	}
	defer res.Body.Close()
	var body struct {
		Count   int `json:"count"`
		Classes map[string]int
		Items   []struct {
			Class  string `json:"class"`
			Kind   string `json:"kind"`
			Name   string `json:"name"`
			Digest string `json:"digest"`
			Link   string `json:"link"`
		} `json:"items"`
		Watchdog struct {
			Pending     int  `json:"pending"`
			ApproveIdle bool `json:"approveIdle"`
		} `json:"watchdog"`
	}
	if err := json.NewDecoder(res.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	var pendRow *struct {
		Class  string `json:"class"`
		Kind   string `json:"kind"`
		Name   string `json:"name"`
		Digest string `json:"digest"`
		Link   string `json:"link"`
	}
	for i := range body.Items {
		it := &body.Items[i]
		if it.Class == "ApprovalNeeded" && it.Kind == "dispatch" {
			pendRow = it
		}
	}
	if pendRow == nil {
		t.Fatalf("挂起行未进行集（items=%d——幽灵计数缺口未收口）", len(body.Items))
	}
	if pendRow.Name != "wf-a" || pendRow.Digest == "" || pendRow.Link != "#/wf/wf-a" {
		t.Fatalf("挂起行投影=%+v（期望 name=wf-a link=#/wf/wf-a digest 派单待批词形）", pendRow)
	}
	// 计数=行数（同源不翻倍）。
	anCount := 0
	for _, it := range body.Items {
		if it.Class == "ApprovalNeeded" {
			anCount++
		}
	}
	if body.Classes["ApprovalNeeded"] != anCount {
		t.Fatalf("classes.ApprovalNeeded=%d 行数=%d（期望相等——加法投影摘除）", body.Classes["ApprovalNeeded"], anCount)
	}
	if body.Count != len(body.Items) {
		t.Fatalf("count=%d 行数=%d（期望相等）", body.Count, len(body.Items))
	}
	if body.Watchdog.Pending != 1 || body.Watchdog.ApproveIdle {
		t.Fatalf("watchdog=%+v（期望 pending=1 非空闲）", body.Watchdog)
	}
	// 人批=Apply 判定 → 挂起行消失（回链闭环——下一次装载反映操作结果）。
	sy.Sync([]opslog.ApplyLog{{Ts: time.Now(), FieldManager: flowbind.ApprovalManager,
		Subject: "console-operator", Obj: "ConnectorOps/global/wf-a", ResourceVersion: 2, Result: "ok"}})
	res2, err := http.Get(ts.URL + "/v1/inbox")
	if err != nil {
		t.Fatal(err)
	}
	defer res2.Body.Close()
	var body2 struct {
		Count    int `json:"count"`
		Classes  map[string]int
		Items    []struct{ Class, Kind string } `json:"items"`
		Watchdog struct {
			Pending     int  `json:"pending"`
			ApproveIdle bool `json:"approveIdle"`
		} `json:"watchdog"`
	}
	if err := json.NewDecoder(res2.Body).Decode(&body2); err != nil {
		t.Fatal(err)
	}
	for _, it := range body2.Items {
		if it.Class == "ApprovalNeeded" && it.Kind == "dispatch" {
			t.Fatal("批准后挂起行仍在集（回链闭环缺口——挂起行应随翻转消失）")
		}
	}
	if !body2.Watchdog.ApproveIdle || body2.Watchdog.Pending != 0 {
		t.Fatalf("watchdog=%+v（期望 approveIdle=true 挂起归零——真实词形）", body2.Watchdog)
	}
}
