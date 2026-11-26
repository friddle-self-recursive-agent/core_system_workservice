// stream_test.go —— stream 面验收（v0.66 R-ENG §2）：snapshot 载荷形状（face 缺席=null）、
// WS 连接即推 snapshot、真总线事件→含 patch 的帧（coder/websocket 客户端 + 进程内存归档
// + 真 bus + run 投影面接线）、diff 器单测见 diff_test.go。
package apiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/coder/websocket"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
)

// TestSnapshotPayloadShape /v1/snapshot 载荷形状：14 键恒在（G-02 聚合完整性——events 键
// v0.67 补齐）；face 缺席=null 不冒充；schemeKinds=注册表 kind 名清单；generatedAt 可解析。
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
	for _, key := range []string{"schemeKinds", "objects", "inbox", "events", "runs", "tasks",
		"sessions", "deliveries", "wakes", "dispatches", "proposals", "drafts", "metrics", "generatedAt"} {
		if _, ok := body[key]; !ok {
			t.Fatalf("快照缺键 %q：%v", key, body)
		}
	}
	// face 缺席=字段 null（不冒充空清单——诚实投影词形）。
	for _, key := range []string{"objects", "inbox", "events", "runs", "tasks", "sessions",
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
	firstPatch := readPatchFrame(t, ctx, ws, deadlineAt(ctx, 5*time.Second), first.Seq)
	if firstPatch.Seq <= first.Seq {
		t.Fatalf("patch seq 应单调大于快照 seq：%d vs %d", firstPatch.Seq, first.Seq)
	}
	hit := false
	for _, op := range firstPatch.Patch {
		if strings.HasPrefix(op.Path, "/runs") {
			hit = true
		}
	}
	if !hit {
		t.Fatalf("patch 应命中 /runs 域（run 投影变化），得 %v", firstPatch.Patch)
	}

	// G-02 seq 单调锚：第二事件→第二 patch 帧，seq 严格递增（断点续传判据——消费方
	// 丢弃 seq≤已应用帧）。
	e2 := eventbus.Event{
		Type: "DocApplied", Source: "apply", Involved: "Agent/global/probe2",
		DedupKey: "apply:Agent/global/probe2:1", Level: "info",
	}
	if err := bus.Publish(&e2); err != nil {
		t.Fatal(err)
	}
	secondPatch := readPatchFrame(t, ctx, ws, deadlineAt(ctx, 5*time.Second), firstPatch.Seq)
	if secondPatch.Seq <= firstPatch.Seq {
		t.Fatalf("第二 patch seq 应严格大于第一 patch seq：%d vs %d", secondPatch.Seq, firstPatch.Seq)
	}
}

// patchFrame stream 推送帧（patch 词形）。
type patchFrame struct {
	Type  string    `json:"type"`
	Seq   int64     `json:"seq"`
	Patch []PatchOp `json:"patch"`
}

// readPatchFrame 读到下一个 patch 帧（跳过心跳/未知词形帧；超时= Fatal；patch 空= Fatal
// ——deadline 由调用方给）。
func readPatchFrame(t *testing.T, ctx context.Context, ws *websocket.Conn, deadline <-chan time.Time, afterSeq int64) patchFrame {
	t.Helper()
	for {
		var frame patchFrame
		err := readJSONFrame(ctx, ws, &frame, 500*time.Millisecond)
		if err != nil {
			select {
			case <-deadline:
				t.Fatalf("事件后未收到 patch 帧（变更感知链失败；基线 seq=%d）: %v", afterSeq, err)
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
		return frame
	}
}

// deadlineAt ctx 截止前的有界 deadline 通道（readPatchFrame 超时判据用）。
func deadlineAt(ctx context.Context, d time.Duration) <-chan time.Time {
	c := make(chan time.Time, 1)
	go func() {
		timer := time.NewTimer(d)
		defer timer.Stop()
		select {
		case <-timer.C:
			c <- time.Now()
		case <-ctx.Done():
		}
	}()
	return c
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

// TestSnapshotAggregateIntegrity G-02 聚合完整性集成锚：faces 真接线（sqlite 文档库+内存
// 归档+真 bus+run 投影面）时快照 objects/events/tasks/runs 关键键在位且**有货**——
// objects 桶携真实对象 ref、events 携归档行、tasks 携 Task 行、runs 摘要计数推进。
// （face 缺席=null 的形状锚另见 TestSnapshotPayloadShape——两锚合起来覆盖聚合面两态）。
func TestSnapshotAggregateIntegrity(t *testing.T) {
	schRaw := `apiVersion: core/v1
kind: Scheme
metadata: {name: snap-agg, scope: builtin}
spec:
  version: core/v1
  kinds:
    - {kind: Scheme, builtin: true, storage: core/scheme.yaml}
    - {kind: Agent, builtin: true, storage: global/agents/<name>/agent.yaml}
    - {kind: Task, builtin: true, storage: global/tasks/<name>/task.yaml}
  eventTypes:
    - {type: DocApplied, source: core, defaultLevel: info, route: owner-workflow}
`
	sch, err := scheme.Load([]byte(schRaw))
	if err != nil {
		t.Fatal(err)
	}
	// sqlite 文档库（ListMeta 元数据直读路径照走）+ Task/Agent 对象各一。
	repo, err := repository.OpenSQLite(filepath.Join(t.TempDir(), "docs.db"), sch, nil, "")
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	ctx := context.Background()
	if _, _, err := repo.Write(ctx, &repository.Object{Kind: "Agent", Scope: "global", Name: "agg-a",
		Content: []byte("apiVersion: core/v1\nkind: Agent\nmetadata:\n  name: agg-a\n  scope: global\n")}, "fm", ""); err != nil {
		t.Fatal(err)
	}
	if _, _, err := repo.Write(ctx, &repository.Object{Kind: "Task", Scope: "global", Name: "agg-t",
		Content: []byte("apiVersion: core/v1\nkind: Task\nmetadata:\n  name: agg-t\n  scope: global\nspec:\n  title: agg probe\nstatus:\n  phase: queued\n")}, "fm", ""); err != nil {
		t.Fatal(err)
	}
	// 真 bus+run 面：归档一行 DocApplied + run 投影落环（events/runs 键数据源）。
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
	stop := hub.AttachDispatch(bus, []scheme.DispatchRule{{On: "DocApplied", For: "any", Dispatch: "workflows/" + workflow.ReconcileWorkflow}})
	defer stop()
	e := eventbus.Event{Type: "DocApplied", Source: "apply", Involved: "Agent/global/agg-a",
		DedupKey: "agg:1", Level: "info"}
	if err := bus.Publish(&e); err != nil {
		t.Fatal(err)
	}
	deadline := time.After(2 * time.Second)
	for len(runs.State().Runs) == 0 {
		select {
		case <-deadline:
			t.Fatal("run 投影未在时限内落环")
		case <-time.After(5 * time.Millisecond):
		}
	}

	ts := httptest.NewServer(New(sch, repo, nil, nil, hub, runs, nil, nil, nil, WithEventLog(opsDB)))
	defer ts.Close()
	resp, err := ts.Client().Get(ts.URL + "/v1/snapshot")
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatal(err)
	}
	// objects：Agent 桶在位携真实 ref。
	objs, _ := body["objects"].(map[string]any)
	agents, _ := objs["Agent"].([]any)
	if len(agents) == 0 {
		t.Fatalf("objects.Agent 桶应有真实 ref，得 %v", body["objects"])
	}
	ref := agents[0].(map[string]any)
	if ref["name"] != "agg-a" || ref["scope"] != "global" {
		t.Fatalf("objects ref 应与落库对象同源：%v", ref)
	}
	// events：归档行在位（eventId 同源）。
	evs, _ := body["events"].([]any)
	if len(evs) == 0 {
		t.Fatalf("events 聚合应携归档行，得 %v", body["events"])
	}
	if evs[0].(map[string]any)["eventId"] != e.ID {
		t.Fatalf("events 行应与发布回执同源：%v vs %s", evs[0], e.ID)
	}
	// tasks：Task 行在位（phase 投影）。
	tasks, _ := body["tasks"].([]any)
	if len(tasks) != 1 || tasks[0].(map[string]any)["phase"] != "queued" {
		t.Fatalf("tasks 简表应携 Task 行（phase 投影），得 %v", body["tasks"])
	}
	// runs：摘要计数推进 + recent 行在位。
	rs, _ := body["runs"].(map[string]any)
	if rs == nil || rs["count"].(float64) != 1 {
		t.Fatalf("runs 摘要应计数 1，得 %v", body["runs"])
	}
	recent, _ := rs["recent"].([]any)
	if len(recent) != 1 || recent[0].(map[string]any)["run"].(float64) != 1 {
		t.Fatalf("runs.recent 应携 run 行，得 %v", rs)
	}
}
