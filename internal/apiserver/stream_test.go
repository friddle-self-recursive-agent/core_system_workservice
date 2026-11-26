// stream_test.go —— stream 面验收（v0.66 R-ENG §2）：snapshot 载荷形状（face 缺席=null）、
// WS 连接即推 snapshot、真总线事件→含 patch 的帧（coder/websocket 客户端 + 进程内存归档
// + 真 bus + run 投影面接线）、diff 器单测见 diff_test.go。
package apiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
)

// TestSnapshotPayloadShape /v1/snapshot 载荷形状：13 键恒在；face 缺席=null 不冒充；
// schemeKinds=注册表 kind 名清单；generatedAt 可解析。
func TestSnapshotPayloadShape(t *testing.T) {
	sch := testScheme(t)
	ts := httptest.NewServer(New(sch, nil, nil, nil, nil, nil, nil, nil, nil))
	defer ts.Close()
	resp, err := ts.Client().Get(ts.URL + "/v1/snapshot")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("snapshot 应 200（聚合投影恒注册），得 %d", resp.StatusCode)
	}
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("snapshot 非 JSON: %v", err)
	}
	for _, key := range []string{"schemeKinds", "objects", "inbox", "runs", "tasks",
		"sessions", "deliveries", "wakes", "dispatches", "proposals", "drafts", "metrics", "generatedAt"} {
		if _, ok := body[key]; !ok {
			t.Fatalf("快照缺键 %q：%v", key, body)
		}
	}
	// face 缺席=字段 null（不冒充空清单——诚实投影词形）。
	for _, key := range []string{"objects", "inbox", "runs", "tasks", "sessions",
		"deliveries", "wakes", "dispatches", "proposals", "drafts", "metrics"} {
		if body[key] != nil {
			t.Fatalf("face 缺席时 %q 应为 null，得 %v", key, body[key])
		}
	}
	kinds, ok := body["schemeKinds"].([]any)
	if !ok || len(kinds) == 0 {
		t.Fatalf("schemeKinds 应为非空清单，得 %v", body["schemeKinds"])
	}
	if kinds[0] != "Scheme" {
		t.Fatalf("schemeKinds 首项应为注册表首 kind=Scheme，得 %v", kinds)
	}
	if _, err := time.Parse(time.RFC3339Nano, body["generatedAt"].(string)); err != nil {
		t.Fatalf("generatedAt 应 RFC3339Nano 可解析: %v", err)
	}
}

// TestStreamSnapshotThenPatch WS 协议验收：连接即推 snapshot；真总线事件（DocApplied×
// dispatch 表→run 投影变化）后收到含非空 patch 的帧（seq 单调大于快照帧）。
func TestStreamSnapshotThenPatch(t *testing.T) {
	sch := testScheme(t)
	// 真 bus + 进程内存归档（apply 落面→DocApplied→订阅派发→run 投影的同一条主干）。
	opsDB, err := opslog.OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer opsDB.Close()
	bus := eventbus.NewBus(&eventbus.Options{Ops: opsDB})
	hub := workflow.NewTriggerHub(nil)
	runs := workflow.NewRunLog(nil)
	hub.BindRuns(runs)
	hub.WireExecutor(workflow.ExecWorkflow)
	stop := hub.AttachDispatch(bus, []scheme.DispatchRule{{On: "DocApplied", For: "any", Dispatch: "workflows/probe"}})
	defer stop()

	ts := httptest.NewServer(New(sch, nil, nil, nil, hub, runs, nil, nil, nil, WithEventBus(bus)))
	defer ts.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	ws, _, err := websocket.Dial(ctx, "ws://"+strings.TrimPrefix(ts.URL, "http://")+"/v1/stream", nil)
	if err != nil {
		t.Fatalf("ws 连接失败: %v", err)
	}
	defer ws.Close(websocket.StatusNormalClosure, "test")

	// 第一帧：snapshot（连接即推）。
	var first struct {
		Type string         `json:"type"`
		Seq  int64          `json:"seq"`
		Data map[string]any `json:"data"`
	}
	if err := readJSONFrame(ctx, ws, &first, 5*time.Second); err != nil {
		t.Fatalf("首帧读取失败: %v", err)
	}
	if first.Type != "snapshot" {
		t.Fatalf("首帧应为 snapshot，得 %q", first.Type)
	}
	// runs face 已接线：基线快照 runs=摘要投影（started=0 零 run 起点）。
	rs, ok := first.Data["runs"].(map[string]any)
	if !ok || rs["started"].(float64) != 0 {
		t.Fatalf("基线快照 runs 应为零 run 摘要，得 %v", first.Data["runs"])
	}

	// 触发一次对象 apply 同源事件（DocApplied 经真总线发布——留痕→run 投影推进）。
	e := eventbus.Event{
		Type: "DocApplied", Source: "apply", Involved: "Agent/global/probe",
		DedupKey: "apply:Agent/global/probe:1", Level: "info",
	}
	if err := bus.Publish(&e); err != nil {
		t.Fatal(err)
	}

	// 后续帧：patch（非空数组；seq>快照 seq；路径命中 /runs 域）。
	deadline := time.After(5 * time.Second)
	for {
		var frame struct {
			Type  string    `json:"type"`
			Seq   int64     `json:"seq"`
			Patch []PatchOp `json:"patch"`
		}
		err := readJSONFrame(ctx, ws, &frame, 500*time.Millisecond)
		if err != nil {
			select {
			case <-deadline:
				t.Fatalf("事件后未收到 patch 帧（变更感知链失败）: %v", err)
			case <-time.After(10 * time.Millisecond):
				continue
			}
		}
		if frame.Type != "patch" {
			continue // 心跳/未知词形帧跳过
		}
		if len(frame.Patch) == 0 {
			t.Fatalf("patch 帧应非空数组，得 %v", frame)
		}
		if frame.Seq <= first.Seq {
			t.Fatalf("patch seq 应单调大于快照 seq：%d vs %d", frame.Seq, first.Seq)
		}
		hit := false
		for _, op := range frame.Patch {
			if strings.HasPrefix(op.Path, "/runs") {
				hit = true
			}
		}
		if !hit {
			t.Fatalf("patch 应命中 /runs 域（run 投影变化），得 %v", frame.Patch)
		}
		break
	}
}

// readJSONFrame 读一文本帧并解码（超时=err；控制帧由库在 Read 内处理不占帧）。
func readJSONFrame(ctx context.Context, ws *websocket.Conn, v any, timeout time.Duration) error {
	rctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	_, data, err := ws.Read(rctx)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}
