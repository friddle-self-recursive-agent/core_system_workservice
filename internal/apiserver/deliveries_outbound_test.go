// deliveries_outbound_test.go —— 送达投影面 v0.17 出站状态机的 HTTP 载荷验收：
// 新代载荷键 sent/failed/attempts/completedAt（恒在=代际判别器）+ status 经出站完成
// 就地翻转 accepted→sent/failed（查询面直接可验）。
package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
)

// deliverStub 出站注入实现（记录调用；可编程失败）。
type deliverStub struct {
	fail error
	got  chan struct{}
}

func (s *deliverStub) Send(_ context.Context, _ workflow.DeliveryRecord) error {
	if s.got != nil {
		s.got <- struct{}{}
	}
	return s.fail
}

// 编译期保险丝：签名必须匹配 workflow.Outbound（契约漂移此处即炸）。
var _ workflow.Outbound = (*deliverStub)(nil)

// TestDeliveriesOutboundKeysAndFlip 新代载荷键 + 出站翻转经 HTTP 全流程可见：
// ① 空态：sent/failed 键恒在（旧构建缺席=代际判别器）；② 出站成功：deliveries 最新条
// status=sent ∧ attempts=1 ∧ completedAt 非空 ∧ 无 error 键；③ 出站失败：status=failed
// ∧ error 键携原因。
func TestDeliveriesOutboundKeysAndFlip(t *testing.T) {
	sch := testScheme(t)

	// ① 空态键位。
	bus0 := eventbus.NewBus(nil)
	hub0 := workflow.NewTriggerHub(nil)
	dl0 := workflow.NewDeliveryLog(nil)
	hub0.BindDeliveries(dl0)
	hub0.WireExecutor(workflow.ExecNotification)
	defer hub0.AttachDispatch(bus0, []scheme.DispatchRule{
		{On: "DocApplied", For: "any", Dispatch: "notification/mail-chan"},
	})()
	ts0 := httptest.NewServer(New(sch, nil, nil, nil, hub0, nil, dl0, nil, nil))
	defer ts0.Close()
	code, empty := get(t, ts0, "/v1/notifications/deliveries")
	if code != 200 {
		t.Fatalf("空态送达面应 200，得 %d", code)
	}
	for _, k := range []string{"sent", "failed", "accepted", "dropped", "deliveries"} {
		if _, ok := empty[k]; !ok {
			t.Fatalf("新代空态载荷缺键 %q（代际判别器）: %v", k, empty)
		}
	}
	if v, _ := empty["sent"].(float64); v != 0 {
		t.Fatalf("空态 sent 应 0: %v", empty)
	}

	// ②③ 出站翻转经 HTTP 可见。
	bus := eventbus.NewBus(nil)
	hub := workflow.NewTriggerHub(nil)
	stub := &deliverStub{got: make(chan struct{}, 4)}
	dl := workflow.NewDeliveryLog(&workflow.DeliveryOptions{Outbound: stub})
	defer dl.Close()
	hub.BindDeliveries(dl)
	hub.WireExecutor(workflow.ExecNotification)
	defer hub.AttachDispatch(bus, []scheme.DispatchRule{
		{On: "DocApplied", For: "any", Dispatch: "notification/mail-chan"},
	})()
	ts := httptest.NewServer(New(sch, nil, nil, nil, hub, nil, dl, nil, nil))
	defer ts.Close()

	if err := bus.Publish(&eventbus.Event{Type: "DocApplied", Source: "apply",
		Involved: "Agent/global/probe", DedupKey: "apply:probe:1", Level: "info"}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-stub.got:
	case <-time.After(2 * time.Second):
		t.Fatal("出站实现未被调用")
	}
	d := pollDelivery(t, ts)
	if d["status"] != "sent" || d["attempts"].(float64) != 1 || d["completedAt"] == "" {
		t.Fatalf("成功翻转应 {status=sent, attempts=1, completedAt≠空}: %v", d)
	}
	if _, has := d["error"]; has {
		t.Fatalf("成功投影不应携 error 键: %v", d)
	}

	// ③ 失败翻转：新事件 → 出站失败 → error 键出现。
	stub.fail = errors.New("中继拒绝")
	if err := bus.Publish(&eventbus.Event{Type: "DocApplied", Source: "apply",
		Involved: "Agent/global/probe", DedupKey: "apply:probe:2", Level: "info"}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-stub.got:
	case <-time.After(2 * time.Second):
		t.Fatal("出站实现未被调用（失败例）")
	}
	deadline := time.After(2 * time.Second)
	for {
		d = pollDelivery(t, ts)
		if d["status"] == "failed" {
			break
		}
		select {
		case <-deadline:
			t.Fatalf("失败翻转未在时限内可查: %v", d)
		case <-time.After(10 * time.Millisecond):
		}
	}
	if d["attempts"].(float64) != 1 || d["error"] != "中继拒绝" {
		t.Fatalf("失败投影应 {attempts=1, error=中继拒绝}: %v", d)
	}
}

// pollDelivery 取送达面最新一条投影（HTTP 查询；空面=阻塞重试至有）。
func pollDelivery(t *testing.T, ts *httptest.Server) map[string]any {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		code, body := get(t, ts, "/v1/notifications/deliveries")
		if code != 200 {
			t.Fatalf("送达面应 200，得 %d", code)
		}
		if raw, err := json.Marshal(body["deliveries"]); err == nil {
			var items []map[string]any
			if json.Unmarshal(raw, &items) == nil && len(items) > 0 {
				return items[len(items)-1]
			}
		}
		select {
		case <-deadline:
			t.Fatalf("送达投影未在时限内出现: %v", body)
		case <-time.After(10 * time.Millisecond):
		}
	}
}
