// context_face_test.go —— 上下文读面四路由 HTTP 验收（v0.65 context P0 数据面 + trace 1:N，
// BUG-002/005/011——B2/B5/B11 判据；测试锚 L81-01 系随包内单测共续）：
// 200 载荷词形（count/docs、count/hits、a/b/score/breakdown、implements/implemented_by/
// verified_by）、参数缺省/非法 400、对象缺席 404、face 缺席 503 防御面、旧构建 405 代际
// 判别器、文档无 trace 段=空聚合 200 不 404。
package apiserver

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/contextidx"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
)

// ctxFakeSrc 测试对象源（contextidx.ObjectSource 与 apiserver.ObjectSource 同形）。
type ctxFakeSrc struct{ objs []*repository.Object }

func (f *ctxFakeSrc) List(_ context.Context, kind, scope string) ([]*repository.Object, error) {
	var out []*repository.Object
	for _, o := range f.objs {
		if o.Kind != kind {
			continue
		}
		if scope != "" && o.Scope != scope {
			continue
		}
		out = append(out, o)
	}
	return out, nil
}

func (f *ctxFakeSrc) Get(_ context.Context, kind, scope, name string) (*repository.Object, error) {
	for _, o := range f.objs {
		if o.Kind == kind && o.Scope == scope && o.Name == name {
			return o, nil
		}
	}
	return nil, repository.ErrNotFound
}

func ctxMD(kind, scope, name, fm string, rv ...int64) *repository.Object {
	ver := int64(1)
	if len(rv) > 0 {
		ver = rv[0]
	}
	return &repository.Object{Kind: kind, Scope: scope, Name: name, ResourceVersion: ver,
		Content: []byte("---\n" + fm + "\n---\n# " + name + "\n正文\n")}
}

// ctxFixture 1 设计（rv=3）+ 2 实现（一篇对齐 designRev=3、一篇过期 designRev=2）+ 1 验证
// + 1 无 trace 文档 + 1 无关文档——B11 全词形覆盖。
func ctxFixture() *ctxFakeSrc {
	return &ctxFakeSrc{objs: []*repository.Object{
		ctxMD("Rule", "core_system", "design-doc", "name: design-doc\nimportance: 90\ntrace:\n  requires: [impl-a]", 3),
		ctxMD("Rule", "core_system", "impl-a", "name: impl-a\nimportance: 70\ntrace:\n  implements:\n    - {doc: Rule/core_system/design-doc, sections: [\"§3\"], designRev: \"3\"}"),
		ctxMD("Rule", "core_system", "impl-b", "name: impl-b\nimportance: 60\ntrace:\n  implements:\n    - {doc: Rule/core_system/design-doc, designRev: \"2\"}"),
		ctxMD("Rule", "core_system", "verify-a", "name: verify-a\ntrace:\n  verifies:\n    - {doc: Rule/core_system/design-doc, targetRev: \"3\"}"),
		ctxMD("Rule", "core_system", "plain-doc", "name: plain-doc\nimportance: 50"),
		ctxMD("Rule", "deploy", "other-doc", "name: other-doc\ntags: [{tag: notify, tag_type: domain}]"),
	}}
}

func ctxServer(t *testing.T, objs []*repository.Object) *httptest.Server {
	t.Helper()
	ix := contextidx.New(&ctxFakeSrc{objs: objs}, func() []string { return []string{"Rule"} })
	if err := ix.Build(context.Background()); err != nil {
		t.Fatalf("索引构建失败: %v", err)
	}
	ts := httptest.NewServer(New(testScheme(t), nil, nil, nil, nil, nil, nil, nil, nil, WithContextIdx(ix)))
	t.Cleanup(ts.Close)
	return ts
}

func ctxGet(t *testing.T, url string) (int, map[string]any) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("载荷不可解析: %v", err)
	}
	return resp.StatusCode, out
}

// ───────────────── 索引投影面 ─────────────────

func TestContextIndexWordShape(t *testing.T) {
	ts := ctxServer(t, ctxFixture().objs)
	code, out := ctxGet(t, ts.URL+"/v1/context/index")
	if code != http.StatusOK {
		t.Fatalf("应 200，得 %d：%v", code, out)
	}
	if n, ok := out["count"].(float64); !ok || int(n) != 6 {
		t.Fatalf("count 应 6，得 %v", out["count"])
	}
	docs, _ := out["docs"].([]any)
	if len(docs) != 6 {
		t.Fatalf("docs 应 6 行，得 %d", len(docs))
	}
	first, _ := docs[0].(map[string]any)
	for _, k := range []string{"kind", "scope", "name", "title", "importance", "tags"} {
		if _, ok := first[k]; !ok {
			t.Fatalf("投影行缺键 %s：%v", k, first)
		}
	}
	// scope 前缀过滤 + tagType 过滤。
	code, out = ctxGet(t, ts.URL+"/v1/context/index?scope=core_system")
	if code != http.StatusOK {
		t.Fatalf("scope 过滤应 200，得 %d", code)
	}
	if n, _ := out["count"].(float64); int(n) != 5 {
		t.Fatalf("scope=core_system 应 5 行（前缀同域），得 %v", out["count"])
	}
	code, out = ctxGet(t, ts.URL+"/v1/context/index?tagType=domain")
	if code != http.StatusOK {
		t.Fatalf("tagType 过滤应 200，得 %d", code)
	}
	if n, _ := out["count"].(float64); int(n) != 1 {
		t.Fatalf("tagType=domain 应 1 行，得 %v", out["count"])
	}
}

// ───────────────── find 检索面 ─────────────────

func TestContextFindWordShapeAndClamp(t *testing.T) {
	ts := ctxServer(t, ctxFixture().objs)
	code, out := ctxGet(t, ts.URL+"/v1/context/find?q=design&tags=doc:rule&k=3")
	_ = code
	hits, _ := out["hits"].([]any)
	if len(hits) == 0 {
		t.Fatalf("查询词+tag 双输入应有命中：%v", out)
	}
	hit, _ := hits[0].(map[string]any)
	for _, k := range []string{"kind", "scope", "name", "score", "relevance", "why"} {
		if _, ok := hit[k]; !ok {
			t.Fatalf("命中行缺键 %s：%v", k, hit)
		}
	}
	rel, _ := hit["relevance"].(map[string]any)
	for _, k := range []string{"tag", "aff", "kw", "scope"} {
		if _, ok := rel[k]; !ok {
			t.Fatalf("relevance 缺分项 %s：%v", k, rel)
		}
	}
	if s, _ := hit["score"].(float64); s <= 0 || s > 100 {
		t.Fatalf("score 越界（0–100）: %v", hit["score"])
	}
	// k 上限收敛 50（防拖库同口径——>50 入参归一）。
	_, out = ctxGet(t, ts.URL+"/v1/context/find?q=doc&k=999")
	if n, _ := out["count"].(float64); int(n) > 50 {
		t.Fatalf("k 应收敛 ≤50，得 %v", out["count"])
	}
	// 零词输入=零命中不炸（200 空榜）。
	code, out = ctxGet(t, ts.URL+"/v1/context/find")
	if code != http.StatusOK {
		t.Fatalf("空查询应 200，得 %d：%v", code, out)
	}
}

// ───────────────── relevance 调试面（B2 落点）─────────────────

func TestContextRelevanceRoute(t *testing.T) {
	ts := ctxServer(t, ctxFixture().objs)
	code, out := ctxGet(t, ts.URL+"/v1/context/relevance?a=Rule/core_system/design-doc&b=Rule/core_system/impl-a")
	if code != http.StatusOK {
		t.Fatalf("应 200，得 %d：%v", code, out)
	}
	for _, k := range []string{"a", "b", "score", "breakdown"} {
		if _, ok := out[k]; !ok {
			t.Fatalf("载荷缺键 %s：%v", k, out)
		}
	}
	bd, _ := out["breakdown"].(map[string]any)
	for _, k := range []string{"tag", "aff", "kw", "scope"} {
		if _, ok := bd[k]; !ok {
			t.Fatalf("breakdown 缺分项 %s：%v", k, bd)
		}
	}
	if s, _ := out["score"].(float64); s < 0 || s > 100 {
		t.Fatalf("score 越界（0–100 标度——B2）: %v", out["score"])
	}
	// 参数不全=400；对象缺席=404（不冒充零分）。
	code, _ = ctxGet(t, ts.URL+"/v1/context/relevance?a=Rule/core_system/design-doc")
	if code != http.StatusBadRequest {
		t.Fatalf("缺 b 应 400，得 %d", code)
	}
	code, _ = ctxGet(t, ts.URL+"/v1/context/relevance?a=Rule/nowhere/ghost&b=Rule/core_system/impl-a")
	if code != http.StatusNotFound {
		t.Fatalf("缺席对象应 404，得 %d", code)
	}
}

// ───────────────── trace 聚合面（B11）─────────────────

func TestTraceRouteWordShape(t *testing.T) {
	ts := ctxServer(t, ctxFixture().objs)
	code, out := ctxGet(t, ts.URL+"/v1/trace/Rule/core_system/design-doc")
	if code != http.StatusOK {
		t.Fatalf("应 200，得 %d：%v", code, out)
	}
	for _, k := range []string{"kind", "scope", "name", "implements", "implemented_by", "verified_by"} {
		if _, ok := out[k]; !ok {
			t.Fatalf("载荷缺键 %s：%v", k, out)
		}
	}
	// 1 对 N 反向聚合：implemented_by=2、verified_by=1。
	implBy, _ := out["implemented_by"].([]any)
	if len(implBy) != 2 {
		t.Fatalf("implemented_by 应 2 篇（1 对 N），得 %d：%v", len(implBy), out)
	}
	verBy, _ := out["verified_by"].([]any)
	if len(verBy) != 1 {
		t.Fatalf("verified_by 应 1 篇，得 %d", len(verBy))
	}
	// stale 派生：designRev=2 < 现 rv=3 → stale=true；designRev=3 → synced。
	staleSeen := false
	syncedSeen := false
	for _, e := range implBy {
		m, _ := e.(map[string]any)
		doc, _ := m["doc"].(string)
		stale, _ := m["stale"].(bool)
		if doc == "Rule/core_system/impl-b" && stale {
			staleSeen = true
		}
		if doc == "Rule/core_system/impl-a" && !stale {
			syncedSeen = true
		}
	}
	if !staleSeen || !syncedSeen {
		t.Fatalf("stale 派生词形异常（应 stale 与 synced 并存）：%v", implBy)
	}
	// N 端正向：impl-a 查 implements。
	code, out = ctxGet(t, ts.URL+"/v1/trace/Rule/core_system/impl-a")
	if code != http.StatusOK {
		t.Fatalf("impl-a 应 200，得 %d", code)
	}
	impl, _ := out["implements"].([]any)
	if len(impl) != 1 {
		t.Fatalf("implements 应 1 条，得 %d", len(impl))
	}
	e0, _ := impl[0].(map[string]any)
	if e0["doc"] != "Rule/core_system/design-doc" {
		t.Fatalf("implements.doc 词形异常：%v", e0)
	}
	if e0["designRev"] != "3" {
		t.Fatalf("designRev 应透传：%v", e0)
	}
	// 无 trace 段=空聚合 200 不 404（对象本身缺席才 404）。
	code, out = ctxGet(t, ts.URL+"/v1/trace/Rule/core_system/plain-doc")
	if code != http.StatusOK {
		t.Fatalf("无 trace 段应 200 空聚合，得 %d", code)
	}
	if n, _ := out["implements"].([]any); len(n) != 0 {
		t.Fatalf("空聚合词形异常：%v", out)
	}
	code, _ = ctxGet(t, ts.URL+"/v1/trace/Rule/nowhere/ghost")
	if code != http.StatusNotFound {
		t.Fatalf("对象缺席应 404，得 %d", code)
	}
	// 键不全=400。
	code, _ = ctxGet(t, ts.URL+"/v1/trace/Rule")
	if code != http.StatusBadRequest {
		t.Fatalf("trace 键不全应 400，得 %d", code)
	}
}

// ───────────────── 防御面与代际判别器 ─────────────────

func TestContextFaceAbsent503And405(t *testing.T) {
	ts := httptest.NewServer(New(testScheme(t), nil, nil, nil, nil, nil, nil, nil, nil))
	t.Cleanup(ts.Close)
	for _, path := range []string{
		"/v1/context/index", "/v1/context/find", "/v1/context/relevance", "/v1/trace/Rule/core_system/design-doc",
	} {
		resp, err := http.Get(ts.URL + path)
		if err != nil {
			t.Fatal(err)
		}
		resp.Body.Close()
		if resp.StatusCode != http.StatusServiceUnavailable {
			t.Fatalf("face 缺席 %s 应 503，得 %d", path, resp.StatusCode)
		}
	}
	// 旧构建（路由未注册）经兜底 handler 405=代际判别器（非 GET 方法）。
	resp, err := http.Post(ts.URL+"/v1/context/index", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("非 GET 应 405（代际判别器），得 %d", resp.StatusCode)
	}
}
