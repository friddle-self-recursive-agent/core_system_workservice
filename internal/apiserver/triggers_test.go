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
	ts := httptest.NewServer(New(testScheme(t), nil, nil, nil, nil))
	defer ts.Close()
	resp, err := http.Get(ts.URL + "/v1/triggers")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusServiceUnavailable {
		t.Fatalf("未接线应 503，得 %d", resp.StatusCode)
	}
}

func TestTriggersFaceRecordsTrace(t *testing.T) {
	sch := testScheme(t)
	bus := eventbus.NewBus(nil)
	hub := workflow.NewTriggerHub(nil)
	stop := hub.Attach(bus, "DocApplied")
	defer stop()

	ts := httptest.NewServer(New(sch, nil, nil, nil, hub))
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
		{On: "*", For: "any", Dispatch: "deadletter"},
	})()

	ts := httptest.NewServer(New(sch, nil, nil, nil, hub))
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
}
