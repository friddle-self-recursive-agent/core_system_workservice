package apiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
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
	return httptest.NewServer(New(sch, nil))
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
	resp2, err := http.Post(ts.URL+"/v1/objects/Agent", "application/json", strings.NewReader("{}"))
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusMethodNotAllowed {
		t.Errorf("objects POST 状态码 = %d, 要 405（只读面，无写端点）", resp2.StatusCode)
	}
}

// ───────────────────────── /v1/objects（文档对象读面）─────────────────────────

func objServer(t *testing.T) (*httptest.Server, *repository.GitRepo) {
	t.Helper()
	sch, err := scheme.Load([]byte(fixture))
	if err != nil {
		t.Fatal(err)
	}
	repo, err := repository.NewGit(t.TempDir(), sch, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	return httptest.NewServer(New(sch, repo)), repo
}

func TestObjectsStoreNotConfigured(t *testing.T) {
	ts := testServer(t) // objects=nil
	defer ts.Close()
	code, body := get(t, ts, "/v1/objects/Agent")
	if code != http.StatusServiceUnavailable {
		t.Errorf("未配置 docs 时状态码 = %d, 要 503", code)
	}
	if !strings.Contains(body["error"].(string), "--docs") {
		t.Errorf("503 载荷缺指引: %v", body)
	}
}

func TestObjectsListAndGet(t *testing.T) {
	ts, repo := objServer(t)
	defer ts.Close()
	ctx := context.Background()
	seed := []repository.Object{
		{Kind: "Agent", Name: "core-headless", Scope: "global", Content: []byte("apiVersion: core/v1\nkind: Agent\nmetadata: {name: core-headless, scope: global}\nspec: {autoLevel: all}\n")},
		{Kind: "Agent", Name: "observer", Scope: "global", Content: []byte("apiVersion: core/v1\nkind: Agent\nmetadata: {name: observer, scope: global}\nspec: {autoLevel: ask}\n")},
	}
	for i := range seed {
		o := seed[i]
		if _, err := repo.Write(ctx, &o, "seed", ""); err != nil {
			t.Fatalf("seed %s: %v", o.Name, err)
		}
	}
	// list
	code, body := get(t, ts, "/v1/objects/Agent")
	if code != http.StatusOK || body["count"] != float64(2) {
		t.Fatalf("list = %d %v", code, body)
	}
	items := body["objects"].([]any)
	first := items[0].(map[string]any)
	if first["name"] != "core-headless" || first["scope"] != "global" {
		t.Errorf("清单项不符: %v", first)
	}
	if _, hasContent := first["content"]; hasContent {
		t.Error("清单不应携带正文（瘦身载荷）")
	}
	// get：读文档对象（含注入的 resourceVersion），非类型注册表
	code, body = get(t, ts, "/v1/objects/Agent/core-headless?scope=global")
	if code != http.StatusOK {
		t.Fatalf("get = %d %v", code, body)
	}
	if body["resourceVersion"] != float64(1) || body["kind"] != "Agent" {
		t.Errorf("get 载荷不符: %v", body)
	}
	content := body["content"].(string)
	if !strings.Contains(content, "autoLevel: all") || !strings.Contains(content, "resourceVersion: 1") {
		t.Errorf("get 正文不符: %s", content)
	}
}

func TestObjectsErrors(t *testing.T) {
	ts, _ := objServer(t)
	defer ts.Close()
	// 未注册 kind（Scheme 闸门）
	code, body := get(t, ts, "/v1/objects/Nope")
	if code != http.StatusNotFound {
		t.Errorf("未注册 kind = %d, 要 404: %v", code, body)
	}
	// 对象缺席
	code, body = get(t, ts, "/v1/objects/Agent/absent?scope=global")
	if code != http.StatusNotFound || !strings.Contains(body["error"].(string), "对象不存在") {
		t.Errorf("缺席对象 = %d %v", code, body)
	}
	// scope 缺失（具体寻址必带）
	code, _ = get(t, ts, "/v1/objects/Agent/absent")
	if code != http.StatusBadRequest {
		t.Errorf("缺 scope = %d, 要 400", code)
	}
}
