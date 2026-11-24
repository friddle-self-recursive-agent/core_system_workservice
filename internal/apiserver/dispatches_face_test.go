// dispatches_face_test.go —— 派单受理投影面 HTTP 载荷验收（v0.50 R-8——workflow-runtime.md
// §11.2 拉起形态首段）：nil 面 503；接线（flowbind.Materializer 派单 sink）后受理可查——
// 载荷形状 {count, accepted, dropped, dispatches:[{seq, agent, status, autoLevel, action,
// workflow, scope, connector, rv, eventId, at}]}；旧代（face 缺席）405 由兜底 handler 固有
// 提供（代际判别器与 wakes/送达面同口径）。
package apiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/flowbind"
)

func TestAgentsDispatchesFace(t *testing.T) {
	// 未接线 → 503（防御面与 wakes 同口径）。
	ts := httptest.NewServer(New(testScheme(t), nil, nil, nil, nil, nil, nil, nil, nil))
	resp, err := http.Get(ts.URL + "/v1/agents/dispatches")
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	ts.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("派单面未接线应 503，得 %d", resp.StatusCode)
	}

	// 接线后：受理投影可查且字段同源（agent/status/autoLevel/action/锚/rv）。
	sch := testScheme(t)
	dl := flowbind.NewDispatchLog(nil)
	dl.Dispatch(flowbind.DispatchRecord{
		Agent: flowbind.DispatchAgent, Status: flowbind.DispatchAccepted, AutoLevel: flowbind.DispatchAutoLevel,
		Action: "created", Workflow: "flow-probe", Scope: "global", Connector: "flow-probe",
		RV: 1, EventID: "evt-1", At: time.Now(),
	})
	ts2 := httptest.NewServer(New(sch, nil, nil, nil, nil, nil, nil, nil, nil, WithDispatchFace(dl)))
	defer ts2.Close()
	resp2, err := http.Get(ts2.URL + "/v1/agents/dispatches")
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("接线后应 200，得 %d", resp2.StatusCode)
	}
	var got struct {
		Count      int              `json:"count"`
		Accepted   int64            `json:"accepted"`
		Dropped    int64            `json:"dropped"`
		Dispatches []map[string]any `json:"dispatches"`
	}
	if err := json.NewDecoder(resp2.Body).Decode(&got); err != nil {
		t.Fatal(err)
	}
	if got.Count != 1 || got.Accepted != 1 || got.Dropped != 0 || len(got.Dispatches) != 1 {
		t.Fatalf("派单面口径应 count=1 accepted=1: %+v", got)
	}
	d := got.Dispatches[0]
	for k, want := range map[string]any{
		"agent": flowbind.DispatchAgent, "status": flowbind.DispatchAccepted,
		"autoLevel": flowbind.DispatchAutoLevel, "action": "created",
		"workflow": "flow-probe", "scope": "global", "connector": "flow-probe",
		"rv": float64(1), "eventId": "evt-1",
	} {
		if d[k] != want {
			t.Fatalf("派单字段 %s: 期望 %v 实际 %v", k, want, d[k])
		}
	}
	if _, ok := d["at"].(string); !ok {
		t.Fatalf("at 应为 RFC3339 串: %v", d["at"])
	}
}
