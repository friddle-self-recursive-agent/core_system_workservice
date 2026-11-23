// draft_test.go —— 草稿面 HTTP 契约验收（cored-bridge.md §2.2）：
// 读面恒开（空=0）、写面缺省 405、落稿 upsert、读单含正文、弃稿、
// fromDraft 闭环（create→清稿→404；冲突 409 保留稿；fromDraft 约束 400）。
package apiserver

import (
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

// draftServer 带写面与草稿缓存的测试宿主（真实依赖链：git 后端 + 引擎 + 内存草稿缓存）。
func draftServer(t *testing.T) *httptest.Server {
	t.Helper()
	sch, err := scheme.Load([]byte(fixture))
	if err != nil {
		t.Fatalf("载入 fixture: %v", err)
	}
	ops, err := opslog.OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ops.Close() })
	repo, err := repository.NewGit(t.TempDir(), sch, &repository.Options{Ops: ops})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	drafts := draft.NewStore(nil)
	engine := &apply.Engine{
		Sch: sch, Repo: repo, Bus: eventbus.NewBus(nil), Ops: ops, Drafts: drafts,
		Perm:    fixturePerm(t), // 授权面=Role 对象（v0.9 对象准入）；fromDraft 同闸门
		Subject: permission.Subject{User: "test", Roles: []string{permission.DefaultAdmin}},
	}
	ts := httptest.NewServer(New(sch, repo, engine, drafts, nil, nil))
	t.Cleanup(ts.Close)
	return ts
}

func del(t *testing.T, ts *httptest.Server, path string) (int, map[string]any) {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, ts.URL+path, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE %s: %v", path, err)
	}
	defer resp.Body.Close()
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("DELETE %s 响应非 JSON: %v", path, err)
	}
	return resp.StatusCode, body
}

func draftBody(name string, baseRV int64) map[string]any {
	return map[string]any{
		"kind": "Agent", "scope": "global", "name": name,
		"content":             "apiVersion: core/v1\nkind: Agent\nmetadata: {name: " + name + ", scope: global}\nspec: {autoLevel: all}\n",
		"fieldManager":        "agent/dev-agent",
		"reasonRef":           "drift/7",
		"baseResourceVersion": baseRV,
	}
}

// TestDraftReadFaceAlwaysOn 写面关闭（缺省只读）下草稿读面恒开、写操作 405。
func TestDraftReadFaceAlwaysOn(t *testing.T) {
	ts := testServer(t) // applier=nil, drafts=nil → 读面 503；本用例专测 nil 缓存
	defer ts.Close()
	code, body := get(t, ts, "/v1/drafts")
	if code != http.StatusServiceUnavailable {
		t.Fatalf("nil 草稿缓存 = %d, 要 503", code)
	}
	if !strings.Contains(body["error"].(string), "草稿缓存未配置") {
		t.Fatalf("503 载荷缺指引: %v", body)
	}

	// 写面关闭、缓存在线：GET 恒开返回空清单，写操作 405。
	sch, _ := scheme.Load([]byte(fixture))
	ts2 := httptest.NewServer(New(sch, nil, nil, draft.NewStore(nil), nil, nil))
	defer ts2.Close()
	code, body = get(t, ts2, "/v1/drafts")
	if code != http.StatusOK || body["count"] != float64(0) {
		t.Fatalf("空草稿清单 = %d %v", code, body)
	}
	code, body = post(t, ts2, "/v1/drafts", draftBody("a", 0))
	if code != http.StatusMethodNotAllowed {
		t.Fatalf("写面关：落稿 = %d %v, 要 405", code, body)
	}
	code, _ = del(t, ts2, "/v1/drafts/Agent/a?scope=global")
	if code != http.StatusMethodNotAllowed {
		t.Fatalf("写面关：弃稿 = %d, 要 405", code)
	}
	code, _ = post(t, ts2, "/v1/apply", map[string]any{"kind": "Agent", "scope": "global", "name": "a", "fromDraft": true, "fieldManager": "fm"})
	if code != http.StatusMethodNotAllowed {
		t.Fatalf("写面关：fromDraft apply = %d, 要 405", code)
	}
}

// TestDraftStagingLifecycle 落稿→查→覆→弃 全生命周期（机械可验载荷形状）。
func TestDraftStagingLifecycle(t *testing.T) {
	ts := draftServer(t)

	// 落稿 → staged。
	code, body := post(t, ts, "/v1/drafts", draftBody("ci", 3))
	if code != http.StatusOK || body["action"] != "staged" || body["baseResourceVersion"] != float64(3) {
		t.Fatalf("落稿 = %d %v", code, body)
	}
	// 清单瘦身：无 content，含 baseResourceVersion/fieldManager。
	code, body = get(t, ts, "/v1/drafts")
	if code != http.StatusOK || body["count"] != float64(1) {
		t.Fatalf("清单 = %d %v", code, body)
	}
	item := body["drafts"].([]any)[0].(map[string]any)
	if item["baseResourceVersion"] != float64(3) || item["fieldManager"] != "agent/dev-agent" {
		t.Fatalf("清单引用不符: %v", item)
	}
	if _, has := item["content"]; has {
		t.Fatalf("清单应为瘦身载荷（无 content）: %v", item)
	}
	// 读单含正文 + reasonRef。
	code, body = get(t, ts, "/v1/drafts/Agent/ci?scope=global")
	if code != http.StatusOK || body["reasonRef"] != "drift/7" {
		t.Fatalf("读单 = %d %v", code, body)
	}
	if c, _ := body["content"].(string); len(c) == 0 {
		t.Fatalf("读单应含正文: %v", body)
	}
	// upsert 覆盖：同键第二稿（迭代后的新 BaseRV）。
	second := draftBody("ci", 5)
	second["fieldManager"] = "agent/reviewer"
	if code, _ := post(t, ts, "/v1/drafts", second); code != http.StatusOK {
		t.Fatalf("覆稿 = %d", code)
	}
	code, body = get(t, ts, "/v1/drafts")
	if body["count"] != float64(1) {
		t.Fatalf("同键覆盖应仍 1 稿: %v", body)
	}
	item = body["drafts"].([]any)[0].(map[string]any)
	if item["baseResourceVersion"] != float64(5) || item["fieldManager"] != "agent/reviewer" {
		t.Fatalf("覆盖后应为最新稿: %v", item)
	}
	// 弃稿 → discarded；读单 404；清单归零。
	code, body = del(t, ts, "/v1/drafts/Agent/ci?scope=global")
	if code != http.StatusOK || body["action"] != "discarded" {
		t.Fatalf("弃稿 = %d %v", code, body)
	}
	code, _ = get(t, ts, "/v1/drafts/Agent/ci?scope=global")
	if code != http.StatusNotFound {
		t.Fatalf("弃后读单 = %d, 要 404", code)
	}
	code, _ = del(t, ts, "/v1/drafts/Agent/ci?scope=global")
	if code != http.StatusNotFound {
		t.Fatalf("再弃 = %d, 要 404", code)
	}
	code, body = get(t, ts, "/v1/drafts")
	if body["count"] != float64(0) {
		t.Fatalf("弃后清单应归零: %v", body)
	}
}

// TestDraftValidationErrors 落稿校验：无主/缺键/空稿 → 400（ErrInvalid）。
func TestDraftValidationErrors(t *testing.T) {
	ts := draftServer(t)
	cases := []struct {
		label string
		body  map[string]any
	}{
		{"无主（rule2 同源拒绝）", func() map[string]any {
			b := draftBody("a", 0)
			delete(b, "fieldManager")
			return b
		}()},
		{"缺 name", func() map[string]any {
			b := draftBody("a", 0)
			delete(b, "name")
			return b
		}()},
		{"缺 scope", func() map[string]any {
			b := draftBody("a", 0)
			delete(b, "scope")
			return b
		}()},
		{"空稿", func() map[string]any {
			b := draftBody("a", 0)
			b["content"] = ""
			return b
		}()},
	}
	for _, c := range cases {
		code, body := post(t, ts, "/v1/drafts", c.body)
		if code != http.StatusBadRequest {
			t.Fatalf("%s = %d %v, 要 400", c.label, code, body)
		}
	}
	// 落稿面不做 scheme 校验（编辑面与存储闸门解耦）：wild kind 草稿可落，
	// apply-draft 时被闸门拒绝 → 404。
	wild := draftBody("rainbow", 0)
	wild["kind"] = "Unicorn"
	code, _ := post(t, ts, "/v1/drafts", wild)
	if code != http.StatusOK {
		t.Fatalf("wild 草稿应可落（闸门在写路径）: %d", code)
	}
	code, _ = postAuth(t, ts, "/v1/apply", map[string]any{
		"kind": "Unicorn", "scope": "global", "name": "rainbow", "fromDraft": true, "fieldManager": "agent:confirm",
	})
	if code != http.StatusNotFound {
		t.Fatalf("wild 稿 apply-draft = %d, 要 404", code)
	}
	// 被拒稿保留（不消费）。
	code, _ = get(t, ts, "/v1/drafts/Unicorn/rainbow?scope=global")
	if code != http.StatusOK {
		t.Fatalf("被拒稿应保留: %d", code)
	}
}

// TestFromDraftApplyClosedLoop 稿→面闭环：fromDraft create→清稿→404；
// 冲突 409 草稿保留；fromDraft 约束（携 content/resourceVersion → 400）。
func TestFromDraftApplyClosedLoop(t *testing.T) {
	ts := draftServer(t)

	// ① 落稿（BaseRV=0=将创建）→ fromDraft apply → create rv=1 + DocApplied 回执。
	if code, _ := post(t, ts, "/v1/drafts", draftBody("core-headless", 0)); code != http.StatusOK {
		t.Fatal("落稿失败")
	}
	code, body := postAuth(t, ts, "/v1/apply", map[string]any{
		"kind": "Agent", "scope": "global", "name": "core-headless", "fromDraft": true, "fieldManager": "agent:confirm",
	})
	if code != http.StatusOK || body["action"] != "create" || body["rv"] != float64(1) {
		t.Fatalf("fromDraft create = %d %v", code, body)
	}
	if evt, ok := body["event"].(map[string]any); !ok || evt["type"] != "DocApplied" {
		t.Fatalf("fromDraft create 应带 DocApplied 回执: %v", body)
	}
	// ② apply 后清稿：读单 404；再 fromDraft → 404（区别于 noop）。
	code, _ = get(t, ts, "/v1/drafts/Agent/core-headless?scope=global")
	if code != http.StatusNotFound {
		t.Fatalf("apply 后应清稿: %d", code)
	}
	code, body = postAuth(t, ts, "/v1/apply", map[string]any{
		"kind": "Agent", "scope": "global", "name": "core-headless", "fromDraft": true, "fieldManager": "agent:confirm",
	})
	if code != http.StatusNotFound {
		t.Fatalf("空草稿 fromDraft = %d %v, 要 404", code, body)
	}
	// ③ 无主 fromDraft → 400（写路径 rule 2）。
	if code, _ := post(t, ts, "/v1/drafts", draftBody("ci", 0)); code != http.StatusOK {
		t.Fatal("落稿失败")
	}
	code, _ = post(t, ts, "/v1/apply", map[string]any{
		"kind": "Agent", "scope": "global", "name": "ci", "fromDraft": true,
	})
	if code != http.StatusBadRequest {
		t.Fatalf("无主 fromDraft = %d, 要 400", code)
	}
	// ④ fromDraft 约束：携 content / resourceVersion → 400。
	code, _ = post(t, ts, "/v1/apply", map[string]any{
		"kind": "Agent", "scope": "global", "name": "ci", "fromDraft": true,
		"fieldManager": "agent:confirm", "content": "x",
	})
	if code != http.StatusBadRequest {
		t.Fatalf("fromDraft+content = %d, 要 400", code)
	}
	code, _ = post(t, ts, "/v1/apply", map[string]any{
		"kind": "Agent", "scope": "global", "name": "ci", "fromDraft": true,
		"fieldManager": "agent:confirm", "resourceVersion": 1,
	})
	if code != http.StatusBadRequest {
		t.Fatalf("fromDraft+resourceVersion = %d, 要 400", code)
	}
	// ⑤ 正常落库后稿清空（noop 路径由引擎级测试覆盖）。
	code, body = postAuth(t, ts, "/v1/apply", map[string]any{
		"kind": "Agent", "scope": "global", "name": "ci", "fromDraft": true, "fieldManager": "agent:confirm",
	})
	if code != http.StatusOK || body["action"] != "create" {
		t.Fatalf("fromDraft ci = %d %v", code, body)
	}
	code, _ = get(t, ts, "/v1/drafts/Agent/ci?scope=global")
	if code != http.StatusNotFound {
		t.Fatalf("fromDraft 后应清稿: %d", code)
	}
}
