// context_face.go —— 上下文读面四路由（v0.65 context P0 数据面 + trace 1:N 追踪）：
//
//	GET /v1/context/index?scope=&tagType=   索引投影（DocIndex 读侧派生）
//	GET /v1/context/find?q=&tags=&k=        TopK 相关性检索（query-doc 由 q/tags 虚构）
//	GET /v1/context/relevance?a=&b=         全局相关性调试面（B2 判据落点）
//	GET /v1/trace/{kind}/{scope}/{name}     trace 双向聚合（R-DOC-9 / B11）
//
// 设计依据：concepts/context.md（0–100 全局相关性/按 Tag 分片）、rule_doc.md R-DOC-9、
// acceptance/bugs/bug.md B2/B5/B11。数据面=contextidx.Index 内存索引（读侧派生，
// 零 DocType、零落库、零写路径）；face 缺席=503 防御面；旧构建（路由未注册）经兜底
// handler 405=代际判别器（/v1/attention 行文风格）。
package apiserver

import (
	"net/http"
	"strings"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/contextidx"
)

// contextIdxFace 上下文索引读面口面（contextidx.Index 满足；接口化便于测试注入假索引）。
type contextIdxFace interface {
	Project(scope, tagType string) []contextidx.DocView
	Find(q contextidx.FindQuery) []contextidx.Hit
	RelevanceByKey(a, b string) (contextidx.RelevanceReport, error)
	Trace(kind, scope, name string) (contextidx.TraceReport, bool)
}

var _ contextIdxFace = (*contextidx.Index)(nil)

// registerContextRoutes 上下文读面路由注册（face 可为 nil=缺席，handler 内恒 503）。
func registerContextRoutes(mux *http.ServeMux, ctxIdx contextIdxFace) {
	// 索引投影面：GET /v1/context/index?scope=&tagType= → {count, docs:[{kind,scope,name,
	// title,importance,tags[],stale?}]}。scope=前缀同域过滤（段边界）；tagType=该类型有打标的
	// 条目；stale=trace 派生过期（下游设计 rv 推进——B11 读侧词形，零写路径）。
	mux.HandleFunc("GET /v1/context/index", func(w http.ResponseWriter, r *http.Request) {
		if ctxIdx == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "上下文索引面未接线（contextidx 未装配——读侧派生面缺席）"})
			return
		}
		docs := ctxIdx.Project(r.URL.Query().Get("scope"), r.URL.Query().Get("tagType"))
		writeJSON(w, http.StatusOK, map[string]any{"count": len(docs), "docs": docs})
	})
	// TopK 检索面：GET /v1/context/find?q=&tags=&k= → {count, hits:[{kind,scope,name,score,
	// relevance:{tag,aff,kw,scope},why}]}。query-doc 由 q/tags 输入虚构（与在册对象同公式）；
	// tags 词形=逗号分隔 "tagType:tag" 或裸 "tag"；k 缺省 10、>50 收敛 50（防拖库同口径）；
	// score 一律 0–100 一位小数（B2 标度词形）。
	mux.HandleFunc("GET /v1/context/find", func(w http.ResponseWriter, r *http.Request) {
		if ctxIdx == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "上下文索引面未接线（contextidx 未装配——读侧派生面缺席）"})
			return
		}
		q := r.URL.Query()
		var tags []string
		for _, t := range strings.Split(q.Get("tags"), ",") {
			if t = strings.TrimSpace(t); t != "" {
				tags = append(tags, t)
			}
		}
		k := atoiDefault(q.Get("k"), 10)
		if k > 50 {
			k = 50
		}
		hits := ctxIdx.Find(contextidx.FindQuery{Q: q.Get("q"), Tags: tags, K: k})
		// 命中行扁平投影（kind/scope/name/score/relevance/why——载荷词形与对象面同构）。
		rows := make([]map[string]any, 0, len(hits))
		for _, h := range hits {
			rows = append(rows, map[string]any{
				"kind": h.Doc.Kind, "scope": h.Doc.Scope, "name": h.Doc.Name,
				"score": h.Score, "relevance": h.Relevance, "why": h.Why,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"count": len(rows), "hits": rows})
	})
	// 相关性调试面：GET /v1/context/relevance?a=&b=（kind/scope/name 词形）→ {a,b,score,
	// breakdown:{tag,aff,kw,scope}}。参数不全=400；任一对象不在索引=404（与 objects 面缺席
	// 分立同口径——不冒充零分）。
	mux.HandleFunc("GET /v1/context/relevance", func(w http.ResponseWriter, r *http.Request) {
		if ctxIdx == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "上下文索引面未接线（contextidx 未装配——读侧派生面缺席）"})
			return
		}
		a, b := r.URL.Query().Get("a"), r.URL.Query().Get("b")
		if a == "" || b == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "缺 a/b 查询参数（词形 kind/scope/name）"})
			return
		}
		rep, err := ctxIdx.RelevanceByKey(a, b)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, rep)
	})
	// trace 双向聚合面：GET /v1/trace/{kind}/{scope}/{name} → {kind,scope,name,implements:[],
	// implemented_by:[], verified_by:[]}（R-DOC-9 1:N 双向追踪；stale=designRev<上游现 rv 的
	// 读侧派生——设计更新自动标下游过期，零写路径）。路径手工解析（scope 允许多段）；
	// 文档无 trace 段=空聚合 200 不 404；对象不在索引=404。
	mux.HandleFunc("GET /v1/trace/", func(w http.ResponseWriter, r *http.Request) {
		if ctxIdx == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "上下文索引面未接线（contextidx 未装配——读侧派生面缺席）"})
			return
		}
		segs := strings.Split(strings.TrimPrefix(r.URL.Path, "/v1/trace/"), "/")
		if len(segs) < 3 || segs[0] == "" || segs[len(segs)-1] == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "trace 键不全（词形 /v1/trace/{kind}/{scope}/{name}）"})
			return
		}
		kind := segs[0]
		name := segs[len(segs)-1]
		scope := strings.Join(segs[1:len(segs)-1], "/")
		rep, ok := ctxIdx.Trace(kind, scope, name)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "对象不在索引（trace 只对在册对象聚合）",
				"obj": kind + "/" + scope + "/" + name})
			return
		}
		writeJSON(w, http.StatusOK, rep)
	})
}
