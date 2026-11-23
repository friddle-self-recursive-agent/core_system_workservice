package apiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/apply"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/draft"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
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
	return httptest.NewServer(New(sch, nil, nil, nil))
}

// fixtureAdminRole/fixtureDevRole 授权面 fixture（请求级主体测试的 Role 对象来源——
// 权限只由 Role 对象授予，主体断言名须对上装载对象）。
const fixtureAdminRole = `apiVersion: core/v1
kind: Role
metadata: {name: fixture-admin, scope: global}
spec:
  permissions:
  - {scope: "global/*", actions: [read, write]}
`

const fixtureDevRole = `apiVersion: core/v1
kind: Role
metadata: {name: fixture-dev, scope: global}
spec:
  permissions:
  - {scope: "systems/dev/*", actions: [read, write]}
`

// fixturePerm 装载 fixture 授权面（v0.9 对象准入）。
func fixturePerm(t *testing.T) *permission.System {
	t.Helper()
	perm := &permission.System{}
	if err := perm.LoadObjects([]permission.ObjectDoc{
		{Kind: "Role", Name: "fixture-admin", Scope: "global", Content: []byte(fixtureAdminRole)},
		{Kind: "Role", Name: "fixture-dev", Scope: "global", Content: []byte(fixtureDevRole)},
	}); err != nil {
		t.Fatalf("装载 fixture 授权面: %v", err)
	}
	return perm
}

// authHeader 请求级主体头（X-Core-Subject，与请求体 subject 同形 JSON）。
func authHeader(user string, roles ...string) string {
	b, _ := json.Marshal(map[string]any{"user": user, "roles": roles})
	return string(b)
}

// postAuth 以 fixture-admin 主体（X-Core-Subject 头）POST——写面闭环测试的默认主体。
func postAuth(t *testing.T, ts *httptest.Server, path string, payload any) (int, map[string]any) {
	t.Helper()
	return postHeader(t, ts, path, payload, authHeader("agent:test", "role/fixture-admin"))
}

// postHeader POST 并携带自定义 X-Core-Subject 头（请求级主体解析路径用）。
func postHeader(t *testing.T, ts *httptest.Server, path string, payload any, subject string) (int, map[string]any) {
	t.Helper()
	buf, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, ts.URL+path, bytes.NewReader(buf))
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(SubjectHeader, subject)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	defer resp.Body.Close()
	var body map[string]any
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&body); err != nil {
		t.Fatalf("POST %s 响应非 JSON: %v", path, err)
	}
	return resp.StatusCode, body
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
	return httptest.NewServer(New(sch, repo, nil, nil)), repo
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
		if _, _, err := repo.Write(ctx, &o, "seed", ""); err != nil {
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

// post JSON 到服务器并解析响应（apply 端点测试用）。
func post(t *testing.T, ts *httptest.Server, path string, payload any) (int, map[string]any) {
	t.Helper()
	buf, err := json.Marshal(payload)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	resp, err := http.Post(ts.URL+path, "application/json", bytes.NewReader(buf))
	if err != nil {
		t.Fatalf("POST %s: %v", path, err)
	}
	defer resp.Body.Close()
	var body map[string]any
	dec := json.NewDecoder(resp.Body)
	if err := dec.Decode(&body); err != nil {
		t.Fatalf("POST %s 响应非 JSON: %v", path, err)
	}
	return resp.StatusCode, body
}

func agentBody(name string, resourceVersion int64) map[string]any {
	return map[string]any{
		"kind": "Agent", "scope": "global", "name": name,
		"resourceVersion": resourceVersion,
		"content":         "apiVersion: core/v1\nkind: Agent\nmetadata: {name: " + name + ", scope: global}\nspec: {autoLevel: all}\n",
	}
}

// TestApplyEndpointDisabled 未开启 --allow-apply：POST /v1/apply 恒 405（缺省只读面不变）。
func TestApplyEndpointDisabled(t *testing.T) {
	ts := testServer(t)
	defer ts.Close()
	code, body := post(t, ts, "/v1/apply", agentBody("a", 0))
	if code != http.StatusMethodNotAllowed || !strings.Contains(body["error"].(string), "未开启") {
		t.Fatalf("未开启写面 = %d %v, 要 405+未开启", code, body)
	}
	// GET /v1/apply 也 405（仅 POST 有意义）。
	code, _ = get(t, ts, "/v1/apply")
	if code != http.StatusMethodNotAllowed {
		t.Fatalf("GET /v1/apply = %d, 要 405", code)
	}
}

// TestApplyEndpointClosedLoop 写面闭环（经真实 git 后端 + apply 引擎）：
// 创建/更新/DryRun 零副作用/错误映射 400(无主)/404(野对象)/409(冲突)。
func TestApplyEndpointClosedLoop(t *testing.T) {
	sch, err := scheme.Load([]byte(fixture))
	if err != nil {
		t.Fatal(err)
	}
	ops, err := opslog.OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer ops.Close()
	repo, err := repository.NewGit(t.TempDir(), sch, &repository.Options{Ops: ops})
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	engine := &apply.Engine{
		Sch: sch, Repo: repo, Bus: eventbus.NewBus(nil), Ops: ops,
		Perm:    fixturePerm(t),                                                             // 授权面=Role 对象（v0.9 对象准入）
		Subject: permission.Subject{User: "test", Roles: []string{permission.DefaultAdmin}}, // 程序化 fallback
	}
	ts := httptest.NewServer(New(sch, repo, engine, draft.NewStore(nil)))
	defer ts.Close()

	// 创建 → 200 create rv=1 + DocApplied 回执（请求级主体经 X-Core-Subject 头）。
	code, body := postAuth(t, ts, "/v1/apply", map[string]any{
		"kind": "Agent", "scope": "global", "name": "core-headless",
		"content":      "apiVersion: core/v1\nkind: Agent\nmetadata: {name: core-headless, scope: global}\nspec: {autoLevel: all}\n",
		"fieldManager": "agent:bootstrap",
	})
	if code != http.StatusOK || body["action"] != "create" || body["rv"] != float64(1) {
		t.Fatalf("create = %d %v", code, body)
	}
	if evt, ok := body["event"].(map[string]any); !ok || evt["type"] != "DocApplied" {
		t.Fatalf("create 应带 DocApplied 回执: %v", body)
	}
	// 读面复测：GET 读出 rv=1。
	code, body = get(t, ts, "/v1/objects/Agent/core-headless?scope=global")
	if code != http.StatusOK || body["resourceVersion"] != float64(1) {
		t.Fatalf("读回 = %d %v", code, body)
	}
	// 无主写入 → 400。
	code, body = postAuth(t, ts, "/v1/apply", agentBody("core-headless", 1))
	if code != http.StatusBadRequest {
		t.Fatalf("无主写入 = %d %v, 要 400", code, body)
	}
	// 野对象 → 404（kind 未注册；闸门先于存储身份校验）。
	wild := agentBody("x", 0)
	wild["fieldManager"] = "agent:bootstrap"
	wild["kind"] = "Unicorn"
	code, _ = postAuth(t, ts, "/v1/apply", wild)
	if code != http.StatusNotFound {
		t.Fatalf("野对象 = %d, 要 404", code)
	}
	// 冲突 → 409（expected/current 版本对）。
	code, body = postAuth(t, ts, "/v1/apply", map[string]any{
		"kind": "Agent", "scope": "global", "name": "core-headless", "resourceVersion": 9,
		"content":      "apiVersion: core/v1\nkind: Agent\nmetadata: {name: core-headless, scope: global}\nspec: {autoLevel: ask}\n",
		"fieldManager": "agent:late",
	})
	if code != http.StatusConflict || body["expected"] != float64(9) || body["current"] != float64(1) {
		t.Fatalf("冲突 = %d %v, 要 409(9→1)", code, body)
	}
	// dryRun → 200 零副作用。
	code, body = postAuth(t, ts, "/v1/apply", map[string]any{
		"kind": "Agent", "scope": "global", "name": "core-headless", "resourceVersion": 1,
		"content":      "apiVersion: core/v1\nkind: Agent\nmetadata: {name: core-headless, scope: global}\nspec: {autoLevel: ask}\n",
		"fieldManager": "agent:bootstrap", "dryRun": true,
	})
	if code != http.StatusOK || body["action"] != "dry-run" || body["rv"] != float64(1) {
		t.Fatalf("dryRun = %d %v", code, body)
	}
	code, body = get(t, ts, "/v1/objects/Agent/core-headless?scope=global")
	if body["resourceVersion"] != float64(1) {
		t.Fatalf("dryRun 后版本演进 = %v, 要仍 1", body["resourceVersion"])
	}
	// 正常更新 → rv=2，读面复测新版本。
	code, body = postAuth(t, ts, "/v1/apply", map[string]any{
		"kind": "Agent", "scope": "global", "name": "core-headless", "resourceVersion": 1,
		"content":      "apiVersion: core/v1\nkind: Agent\nmetadata: {name: core-headless, scope: global}\nspec: {autoLevel: ask}\n",
		"fieldManager": "agent:bootstrap",
	})
	if code != http.StatusOK || body["action"] != "update" || body["rv"] != float64(2) {
		t.Fatalf("update = %d %v", code, body)
	}
	code, body = get(t, ts, "/v1/objects/Agent/core-headless?scope=global")
	if body["resourceVersion"] != float64(2) || !strings.Contains(body["content"].(string), "autoLevel: ask") {
		t.Fatalf("读回新版本 = %v", body)
	}
}

// TestApplyRequestSubject 请求级主体契约（v0.9）：
// 匿名拒绝（403）+ denied 审计行携带主体、域外 role 拒绝、请求体 subject 字段放行、
// 不可解析主体头=匿名照判、草稿落稿不受主体门控（编辑面与存储闸门解耦）。
func TestApplyRequestSubject(t *testing.T) {
	sch, err := scheme.Load([]byte(fixture))
	if err != nil {
		t.Fatal(err)
	}
	ops, err := opslog.OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer ops.Close()
	repo, err := repository.NewGit(t.TempDir(), sch, &repository.Options{Ops: ops})
	if err != nil {
		t.Fatal(err)
	}
	defer repo.Close()
	engine := &apply.Engine{
		Sch: sch, Repo: repo, Bus: eventbus.NewBus(nil), Ops: ops,
		Perm:    fixturePerm(t),
		Subject: permission.Subject{User: "test", Roles: []string{permission.DefaultAdmin}}, // fallback：不兜底 HTTP 匿名
	}
	ts := httptest.NewServer(New(sch, repo, engine, draft.NewStore(nil)))
	defer ts.Close()
	createBody := func() map[string]any {
		return map[string]any{
			"kind": "Agent", "scope": "global", "name": "core-headless",
			"content":      "apiVersion: core/v1\nkind: Agent\nmetadata: {name: core-headless, scope: global}\nspec: {autoLevel: all}\n",
			"fieldManager": "agent:bootstrap",
		}
	}

	// ① 无主体（体/头皆缺）→ 匿名照判拒绝 403；denied 审计行携带主体=anonymous。
	code, body := post(t, ts, "/v1/apply", createBody())
	if code != http.StatusForbidden || !strings.Contains(body["error"].(string), "准入拒绝") {
		t.Fatalf("匿名 apply = %d %v, 要 403", code, body)
	}
	denied, err := ops.QueryApply("denied")
	if err != nil || len(denied) != 1 {
		t.Fatalf("denied 审计行查询 = %d 行, err=%v", len(denied), err)
	}
	if denied[0].Subject != "anonymous" || denied[0].Obj != "Agent/global/core-headless" {
		t.Fatalf("denied 行主体不符: %+v", denied[0])
	}
	if !strings.Contains(denied[0].PolicyDecisions, "judged by role") {
		t.Fatalf("denied 行应留痕判定方: %q", denied[0].PolicyDecisions)
	}

	// ② 持 role 主体但域外（fixture-dev 只授 systems/dev/*）→ 403；denied 行携带该主体。
	code, body = postHeader(t, ts, "/v1/apply", createBody(), authHeader("agent:limited", "role/fixture-dev"))
	if code != http.StatusForbidden {
		t.Fatalf("域外 role apply = %d %v, 要 403", code, body)
	}
	denied, _ = ops.QueryApply("denied")
	if len(denied) != 2 || denied[1].Subject != "agent:limited" {
		t.Fatalf("denied 行应携带请求主体: %+v", denied)
	}

	// ③ 请求体 subject 字段 + 命名 role 对象 → 放行 create（体字段优先路径）。
	b := createBody()
	b["subject"] = map[string]any{"user": "agent:ci", "roles": []string{"role/fixture-admin"}}
	code, body = post(t, ts, "/v1/apply", b)
	if code != http.StatusOK || body["action"] != "create" {
		t.Fatalf("持 role 主体 apply = %d %v, 要 200 create", code, body)
	}

	// ④ 不可解析主体头=视为未携带 → 匿名照判 403（不静默降级成本地信任）。
	code, _ = postHeader(t, ts, "/v1/apply", createBody(), "not-json")
	if code != http.StatusForbidden {
		t.Fatalf("坏主体头 apply = %d, 要 403", code)
	}

	// ⑤ 草稿落稿不受主体门控（编辑面无闸门；准入在 apply 写路径）。
	code, _ = post(t, ts, "/v1/drafts", map[string]any{
		"kind": "Agent", "scope": "global", "name": "staged", "content": "x", "fieldManager": "agent/dev",
	})
	if code != http.StatusOK {
		t.Fatalf("落稿 = %d, 要 200（草稿面不接权限闸门）", code)
	}
}
