package apiserver

// 会话活性面（KI-003，event.md 规则 14）HTTP 契约测试：
// - 心跳注册面 POST /v1/sessions/heartbeat：写面门控（未开启=405 与 apply 同口径）+缺 session_id=400；
// - 查询面 GET /v1/sessions：恒开+watchdog 投影（enabled/interval/stallThreshold/healRetryLimit）；
// - 面缺席=501（装配缺 --session-watch-interval 的防御面）。

import (
	"context"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/apply"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/sessionliveness"
)

var errNoSessionStub = errors.New("sessionId 必填")

// stubLive 最小 SessionLivenessFace（内存账——HTTP 契约面测试不依赖 DB/watchdog 全链）。
type stubLive struct {
	beats []sessionliveness.Heartbeat
}

func (s *stubLive) Beat(h sessionliveness.Heartbeat) error {
	if h.SessionID == "" {
		return errNoSessionStub
	}
	s.beats = append(s.beats, h)
	return nil
}
func (s *stubLive) Heartbeats(limit int) ([]sessionliveness.SessionView, error) {
	return []sessionliveness.SessionView{{
		SessionID: "s1", State: "suspected-stall", Phase: "inject-continue",
		Heartbeats: 3, HealAttempts: 1, Involved: "Session/cored/s1",
		LastBeat: time.Date(2026, 11, 26, 10, 0, 0, 0, time.UTC),
	}}, nil
}
func (s *stubLive) Enabled() bool            { return true }
func (s *stubLive) Interval() time.Duration  { return time.Minute }
func (s *stubLive) Threshold() time.Duration { return 10 * time.Minute }
func (s *stubLive) HealLimit() int           { return 2 }

func TestSessionsHeartbeatWriteGate(t *testing.T) {
	sch, err := scheme.Load([]byte(fixture))
	if err != nil {
		t.Fatalf("载入 fixture 失败: %v", err)
	}
	live := &stubLive{}
	// 写面关断（applier=nil）：405。
	ts := httptest.NewServer(New(sch, nil, nil, nil, nil, nil, nil, nil, nil, WithSessionLiveness(live)))
	defer ts.Close()
	res, err := http.Post(ts.URL+"/v1/sessions/heartbeat", "application/json",
		strings.NewReader(`{"sessionId":"s1","phase":"tool-call"}`))
	if err != nil {
		t.Fatalf("POST: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("写面关断=%d（期望 405）", res.StatusCode)
	}
	if len(live.beats) != 0 {
		t.Fatal("关断面不应受理心跳")
	}
	// 写面开启：202 受理。
	ts2 := httptest.NewServer(New(sch, nil, &applierFaceStub{}, nil, nil, nil, nil, nil, nil, WithSessionLiveness(live)))
	defer ts2.Close()
	res2, err := http.Post(ts2.URL+"/v1/sessions/heartbeat", "application/json",
		strings.NewReader(`{"sessionId":"s1","phase":"tool-call","lastAction":"apply scheme.yaml"}`))
	if err != nil {
		t.Fatalf("POST2: %v", err)
	}
	defer res2.Body.Close()
	if res2.StatusCode != http.StatusOK {
		t.Fatalf("写面开启=%d（期望 200）", res2.StatusCode)
	}
	if len(live.beats) != 1 || live.beats[0].SessionID != "s1" {
		t.Fatalf("心跳未入账=%+v", live.beats)
	}
	// 缺 sessionId=400。
	res3, err3 := http.Post(ts2.URL+"/v1/sessions/heartbeat", "application/json", strings.NewReader(`{"phase":"x"}`))
	if err3 != nil {
		t.Fatal(err3)
	}
	defer res3.Body.Close()
	if res3.StatusCode != http.StatusBadRequest {
		t.Fatalf("缺 sessionId=%d（期望 400）", res3.StatusCode)
	}
}

func TestSessionsFaceAbsent(t *testing.T) {
	sch, err := scheme.Load([]byte(fixture))
	if err != nil {
		t.Fatalf("载入 fixture 失败: %v", err)
	}
	ts := httptest.NewServer(New(sch, nil, nil, nil, nil, nil, nil, nil, nil))
	defer ts.Close()
	res, err := http.Get(ts.URL + "/v1/sessions")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusNotImplemented {
		t.Fatalf("面缺席=%d（期望 501）", res.StatusCode)
	}
}

func TestSessionsQueryShape(t *testing.T) {
	sch, err := scheme.Load([]byte(fixture))
	if err != nil {
		t.Fatalf("载入 fixture 失败: %v", err)
	}
	ts := httptest.NewServer(New(sch, nil, nil, nil, nil, nil, nil, nil, nil, WithSessionLiveness(&stubLive{})))
	defer ts.Close()
	res, err := http.Get(ts.URL + "/v1/sessions")
	if err != nil {
		t.Fatalf("GET: %v", err)
	}
	defer res.Body.Close()
	if res.StatusCode != http.StatusOK {
		t.Fatalf("查询面=%d（期望 200）", res.StatusCode)
	}
	buf, _ := io.ReadAll(res.Body)
	body := string(buf)
	for _, key := range []string{`"count"`, `"sessions"`, `"watchdog"`, `"suspected-stall"`, `"stallThreshold"`, `"healRetryLimit"`} {
		if !strings.Contains(body, key) {
			t.Fatalf("载荷缺键 %s：%s", key, body[:minInt(len(body), 400)])
		}
	}
}

// applierFaceStub Applier 口面最小实现（写面开启态；本测试不触发 apply 语义）。
type applierFaceStub struct{}

func (applierFaceStub) Apply(_ context.Context, _ *repository.Object, _ apply.Options) (*apply.Result, error) {
	return &apply.Result{}, nil
}
func (applierFaceStub) ApplyDraft(_ context.Context, _, _, _ string, _ apply.Options) (*apply.Result, error) {
	return &apply.Result{}, nil
}

func minInt(a, b int) int {
	if a < b {
		return a
	}
	return b
}
