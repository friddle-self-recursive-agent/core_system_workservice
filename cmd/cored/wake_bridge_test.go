package main

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
)

type fakeWakeSink struct{ n int }

func (f *fakeWakeSink) Wake(workflow.TriggerRecord, workflow.ExecutorBinding, string) { f.n++ }

func TestWakeNotifierForwardsOnlyWhitelistedRoot(t *testing.T) {
	got := make(chan map[string]any, 4)
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/wake/spawn" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatal(err)
		}
		got <- body
		w.WriteHeader(http.StatusAccepted)
	}))
	defer ts.Close()

	inner := &fakeWakeSink{}
	n := newWakeNotifier(inner, ts.URL)
	rec := workflow.TriggerRecord{Type: "LoopRequested", EventID: "evt-1", DedupKey: "d1", Level: "info"}
	n.Wake(rec, workflow.ExecutorBinding{Executor: workflow.ExecAgent, Target: "root-agent"}, "route")

	select {
	case body := <-got:
		if body["type"] != "LoopRequested" || body["eventId"] != "evt-1" || body["agent"] != "root-agent" {
			t.Fatalf("payload=%+v", body)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("wake bridge 未转发")
	}

	// 非白名单类型 / 非 root 锚：只落投影，不出站。
	n.Wake(workflow.TriggerRecord{Type: "ReconcileTick", EventID: "evt-2"}, workflow.ExecutorBinding{Target: "root-agent"}, "route")
	n.Wake(workflow.TriggerRecord{Type: "LoopRequested", EventID: "evt-3"}, workflow.ExecutorBinding{Target: "parent-agent"}, "route")
	select {
	case body := <-got:
		t.Fatalf("不应转发：%+v", body)
	case <-time.After(150 * time.Millisecond):
	}
	if inner.n != 3 {
		t.Fatalf("投影落点=%d（期望 3）", inner.n)
	}
}
