package apiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

const fixture = `apiVersion: core/v1
kind: Scheme
metadata: {name: default, scope: builtin}
spec:
  version: core/v1
  kinds:
    - {kind: Scheme, builtin: true, storage: core/scheme.yaml}
    - {kind: Agent, builtin: true, storage: global/agents/<name>/agent.yaml}
  eventTypes:
    - {type: DocApplied, source: core, defaultLevel: info, route: owner-workflow}
`

func testServer(t *testing.T) *httptest.Server {
	t.Helper()
	sch, err := scheme.Load([]byte(fixture))
	if err != nil {
		t.Fatalf("载入 fixture 失败: %v", err)
	}
	return httptest.NewServer(New(sch))
}

func get(t *testing.T, ts *httptest.Server, path string) (int, map[string]any) {
	t.Helper()
	resp, err := http.Get(ts.URL + path)
	if err != nil {
		t.Fatalf("GET %s: %v", path, err)
	}
	defer resp.Body.Close()
	var body map[string]any
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&body); err != nil {
		t.Fatalf("GET %s 响应非 JSON: %v", path, err)
	}
	return resp.StatusCode, body
}

func TestHealthz(t *testing.T) {
	ts := testServer(t)
	defer ts.Close()
	code, body := get(t, ts, "/healthz")
	if code != http.StatusOK || body["ok"] != true {
		t.Errorf("healthz = %d %v", code, body)
	}
}

func TestSchemeSummary(t *testing.T) {
	ts := testServer(t)
	defer ts.Close()
	code, body := get(t, ts, "/v1/scheme")
	if code != http.StatusOK {
		t.Fatalf("scheme 概览状态码 = %d", code)
	}
	if body["apiVersion"] != "core/v1" || body["version"] != "core/v1" {
		t.Errorf("概览字段不符: %v", body)
	}
	if body["kindCount"] != float64(2) || body["eventTypeCount"] != float64(1) {
		t.Errorf("概览计数不符: %v", body)
	}
}

func TestListKinds(t *testing.T) {
	ts := testServer(t)
	defer ts.Close()
	code, body := get(t, ts, "/v1/scheme/kinds")
	if code != http.StatusOK {
		t.Fatalf("kinds 清单状态码 = %d", code)
	}
	if body["count"] != float64(2) {
		t.Errorf("count = %v, 要 2", body["count"])
	}
	kinds, ok := body["kinds"].([]any)
	if !ok || len(kinds) != 2 {
		t.Fatalf("kinds 载荷不符: %v", body["kinds"])
	}
	first := kinds[0].(map[string]any)
	if first["name"] != "Scheme" || first["storage"] != "core/scheme.yaml" {
		t.Errorf("首个 kind 不符: %v", first)
	}
}

func TestGetKind(t *testing.T) {
	ts := testServer(t)
	defer ts.Close()
	code, body := get(t, ts, "/v1/scheme/kinds/Agent")
	if code != http.StatusOK {
		t.Fatalf("get Agent 状态码 = %d", code)
	}
	if body["name"] != "Agent" || body["builtin"] != true {
		t.Errorf("Agent 详情不符: %v", body)
	}
}

func TestGetUnknownKind(t *testing.T) {
	ts := testServer(t)
	defer ts.Close()
	code, body := get(t, ts, "/v1/scheme/kinds/Nope")
	if code != http.StatusNotFound {
		t.Errorf("未注册 kind 状态码 = %d, 要 404", code)
	}
	if !strings.Contains(body["error"].(string), "not registered") {
		t.Errorf("错误载荷不符: %v", body)
	}
}

func TestMethodGuard(t *testing.T) {
	ts := testServer(t)
	defer ts.Close()
	resp, err := http.Post(ts.URL+"/v1/scheme/kinds", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("POST 状态码 = %d, 要 405（只读面）", resp.StatusCode)
	}
}
