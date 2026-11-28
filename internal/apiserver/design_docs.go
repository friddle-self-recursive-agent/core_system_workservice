// design_docs.go —— 设计文档受限文件面（能力 A——console 屏幕 6 直接编辑设计文档 md →
// Apply → git commit；任务书裁决路线②）。读面=docrepo（git HEAD 权威）；写面=三检查
// （白名单 403 / R-PUB-1 lint 400 / 乐观锁 409）+ 权限准入（permission.md 机读）+ 审计
// （apply_log 行 + DocApplied 事件 involved=DesignDoc/<repo>/<path>）。草稿=与对象草稿同一
// 缓存同一持久化面（kind=DesignDoc, scope=<repo>, name=<path>——baseSha 在 apply 请求随行，
// 草稿缓存不承载 sha）。设计依据：core_system_instance/runtime/cored-bridge.md §2.4（v0.52）。
package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/docrepo"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/draft"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
)

// DesignDocs 设计文档面（docrepo.Set 满足；接口化便于测试注入假面）。
type DesignDocs interface {
	List() ([]docrepo.File, error)
	Read(repo, path string) (*docrepo.Doc, error)
	Apply(ctx context.Context, req docrepo.ApplyRequest) (*docrepo.Result, error)
	// WriteEnabled 写面就绪态（引擎装配=--design-repo ∧ --allow-apply；未就绪=写路由 405）。
	WriteEnabled() bool
}

// draftDocKey 设计文档草稿的缓存键（draft.Store=(kind,scope,name)；kind=docrepo.DraftKind）。
func draftDocKey(repo, path string) (kind, scope, name string) {
	return docrepo.DraftKind, repo, path
}

// designDocRepos 设计文档清单投影：{count, repos:[{name, files:[{path,sha}]}], docs:[…]}。
// docs=扁平行集（树渲染直吃的第二种投影，与 repos 同源同拍——零二次拉取）。
func designDocRepos(files []docrepo.File) ([]map[string]any, []map[string]any) {
	byRepo := map[string][]map[string]any{}
	var order []string
	docs := make([]map[string]any, 0, len(files))
	for _, f := range files {
		if _, ok := byRepo[f.Repo]; !ok {
			byRepo[f.Repo] = []map[string]any{}
			order = append(order, f.Repo)
		}
		row := map[string]any{"path": f.Path, "sha": f.Sha}
		byRepo[f.Repo] = append(byRepo[f.Repo], row)
		docs = append(docs, map[string]any{"repo": f.Repo, "path": f.Path, "sha": f.Sha})
	}
	repos := make([]map[string]any, 0, len(order))
	for _, name := range order {
		repos = append(repos, map[string]any{"name": name, "count": len(byRepo[name]), "files": byRepo[name]})
	}
	return repos, docs
}

// writeDesignDocErr 设计文档面错误 → HTTP 状态码映射（cored-bridge.md §2.4 错误表）。
func writeDesignDocErr(w http.ResponseWriter, err error) {
	var conflict *docrepo.ConflictError
	var lintErr *docrepo.LintError
	switch {
	case errors.As(err, &conflict):
		writeJSON(w, http.StatusConflict, map[string]any{
			"error": err.Error(), "expected": conflict.Expected, "current": conflict.Current,
		})
	case errors.As(err, &lintErr):
		writeJSON(w, http.StatusBadRequest, map[string]any{
			"error":      err.Error(),
			"violations": lintErr.Violations, // 违规路径/行号在册——任务书判据「400 含违规路径」
		})
	case errors.Is(err, docrepo.ErrNoRepo), errors.Is(err, docrepo.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
	case errors.Is(err, docrepo.ErrForbidden), errors.Is(err, docrepo.ErrDenied):
		writeJSON(w, http.StatusForbidden, map[string]any{"error": err.Error()})
	case errors.Is(err, docrepo.ErrNoFieldManager), errors.Is(err, docrepo.ErrInvalid):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
	case errors.Is(err, docrepo.ErrNoEngine):
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": err.Error()})
	case errors.Is(err, docrepo.ErrDirty):
		// 工作树有未提交改动=乐观锁族（重整后再落）——409 与冲突同族（原因字段区分）。
		writeJSON(w, http.StatusConflict, map[string]any{"error": err.Error(), "reason": "dirty-worktree"})
	case errors.Is(err, draft.ErrNoDraft):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
}

// designApplyRequest POST /v1/design-docs/apply 请求体（cored-bridge.md §2.4）。
type designApplyRequest struct {
	Repo         string          `json:"repo"`
	Path         string          `json:"path"`
	BaseSha      string          `json:"baseSha"` // 乐观锁基准=打开时读到的 HEAD blob sha（""=预期创建）
	Content      string          `json:"content"` // fromDraft=true 时须缺省（差异以草稿缓存为准）
	FieldManager string          `json:"fieldManager"`
	ReasonRef    string          `json:"reasonRef"`
	DryRun       bool            `json:"dryRun"`
	FromDraft    bool            `json:"fromDraft"`
	Subject      *subjectRequest `json:"subject"`
	SessionID    string          `json:"sessionId,omitempty"` // 第三段回整记账锚（可选）
	MetaTags     []string        `json:"metaTags,omitempty"`  // 第三段并入侧车的会话 tag（可选；缺=仅 importance 回整）
}

// sessionMetaWriteback 第三段·属性回整（session.md §5 默认自动——规则版）：
// 侧车 importance +5 封顶 100 + 会话 tag 并入（tag_type: session）+ 账行落
// session_writeback。LLM 判断回整=declared-deferred（Mode=rule-default 诚实标注）；
// 回整失败只记 Note 不影响 apply 主结果（回整是附带动作不是闸门）。
func sessionMetaWriteback(ctx context.Context, dd DesignDocs, store repository.SessionConfigStore, repo, path, subject, sessionID string, tags []string) *repository.SessionWriteback {
	wb := &repository.SessionWriteback{
		At: time.Now().UTC().Format(time.RFC3339), Resource: repo + "/" + path,
		Subject: subject, SessionID: sessionID, Mode: "rule-default",
	}
	sidePath := path + ".metadatainfo"
	d, err := dd.Read(repo, sidePath)
	if err != nil {
		wb.Note = "侧车缺席（" + sidePath + "）——仅记账不动侧车"
		_ = store.AppendSessionWriteback(ctx, *wb)
		return wb
	}
	content := d.ContentStr
	if newC, delta, ok := repository.BumpSidecarImportance(content, 5); ok {
		content = newC
		wb.ImportanceDelta = delta
	}
	if len(tags) > 0 {
		if newC, added := repository.MergeSidecarTags(content, tags); len(added) > 0 {
			content = newC
			wb.TagsAdded = added
		}
	}
	if _, err := dd.Apply(ctx, docrepo.ApplyRequest{
		Repo: repo, Path: sidePath, BaseSha: d.Sha, Content: []byte(content),
		FieldManager: "session:meta-writeback", ReasonRef: "session:writeback:" + repo + "/" + path,
	}); err != nil {
		wb.Note = "侧车回写失败：" + err.Error()
	}
	_ = store.AppendSessionWriteback(ctx, *wb)
	return wb
}

// designApplyResult 响应投影（commit 仅真实落库在册；event 仅真实落库在册）。
func designApplyResult(res *docrepo.Result) map[string]any {
	out := map[string]any{"action": res.Action, "sha": res.Sha}
	if res.Commit != nil {
		out["commit"] = map[string]any{"sha": res.Commit.Sha, "subject": res.Commit.Subject,
			"at": res.Commit.At.UTC().Format(time.RFC3339Nano)}
	}
	if res.Event != nil {
		out["event"] = map[string]any{"id": res.Event.ID, "type": res.Event.Type, "dedupKey": res.Event.DedupKey}
	}
	return out
}

// registerDesignDocRoutes 设计文档面路由注册（New() 内调用——路由表统一收口）。
// objects 供第三段回整账存储（SessionConfigStore——sqlite docs DB 后端才有；缺席=不回整只 Apply）。
func registerDesignDocRoutes(mux *http.ServeMux, designDocs DesignDocs, drafts DraftStaging, applyLog ApplyLogFace, objects ObjectSource) {
	// 读面清单：{count, repos, docs}——面未装配=503；只读 GET 恒开（装配即可读）。
	mux.HandleFunc("GET /v1/design-docs", func(w http.ResponseWriter, _ *http.Request) {
		if designDocs == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "设计文档面未装配（cored 缺 --design-repo）"})
			return
		}
		files, err := designDocs.List()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		repos, docs := designDocRepos(files)
		writeJSON(w, http.StatusOK, map[string]any{"count": len(docs), "repos": repos, "docs": docs})
	})
	// 读面单文档：HEAD 正文+blob sha+最近提交；白名单外=403；未注册仓/HEAD 缺席=404。
	mux.HandleFunc("GET /v1/design-docs/{repo}/{path...}", func(w http.ResponseWriter, r *http.Request) {
		if designDocs == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "设计文档面未装配（cored 缺 --design-repo）"})
			return
		}
		repo, path := r.PathValue("repo"), r.PathValue("path")
		d, err := designDocs.Read(repo, path)
		if err != nil {
			writeDesignDocErr(w, err)
			return
		}
		out := map[string]any{
			"repo": d.Repo, "path": d.Path, "sha": d.Sha, "bytes": d.Bytes, "content": d.ContentStr,
		}
		if d.LastCommit != nil {
			out["lastCommit"] = map[string]any{"sha": d.LastCommit.Sha, "subject": d.LastCommit.Subject,
				"at": d.LastCommit.At.UTC().Format(time.RFC3339Nano)}
		}
		// 侧车元数据投影（2026-11-28 用户裁决：metadatainfo 上屏）——`<path>.metadatainfo`
		// HEAD 现读（tags/importance）；缺席=不带 meta 键零编造。
		if side, serr := designDocs.Read(repo, path+".metadatainfo"); serr == nil {
			if tags, imp, ok := repository.ParseSidecarMeta(side.ContentStr); ok {
				metaMap := map[string]any{"importance": imp}
				if tags != nil {
					metaMap["tags"] = tags
				}
				out["meta"] = metaMap
			}
		}
		writeJSON(w, http.StatusOK, out)
	})
	// 应用史投影面：apply_log 行（obj=DesignDoc/<repo>/<path> 精确过滤——rv 恒 0，按时刻降序）。
	mux.HandleFunc("GET /v1/design-docs/history/{repo}/{path...}", func(w http.ResponseWriter, r *http.Request) {
		if designDocs == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "设计文档面未装配（cored 缺 --design-repo）"})
			return
		}
		if applyLog == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "应用史查询面未接线（装配缺 ops DB 口面）"})
			return
		}
		repo, path := r.PathValue("repo"), r.PathValue("path")
		obj := docrepo.ObjectKey(repo, path)
		limit := atoiDefault(r.URL.Query().Get("limit"), 0)
		rows, err := applyLog.QueryApplies(obj, limit)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		items := make([]map[string]any, 0, len(rows))
		for _, a := range rows {
			items = append(items, map[string]any{
				"ts": a.Ts.UTC().Format(time.RFC3339Nano), "rv": a.ResourceVersion,
				"result": a.Result, "fieldManager": a.FieldManager, "subject": a.Subject,
				"reasonRef": a.ReasonRef,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"count": len(items), "obj": obj, "entries": items})
	})
	// 草稿读面（恒开）：kind=DesignDoc, scope=<repo>, name=<path>。
	mux.HandleFunc("GET /v1/design-docs/draft/{repo}/{path...}", func(w http.ResponseWriter, r *http.Request) {
		if designDocs == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "设计文档面未装配（cored 缺 --design-repo）"})
			return
		}
		if drafts == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "草稿缓存未配置"})
			return
		}
		repo, path := r.PathValue("repo"), r.PathValue("path")
		kind, scope, name := draftDocKey(repo, path)
		d, err := drafts.Get(kind, scope, name)
		if err != nil {
			writeDraftErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"kind": d.Kind, "repo": d.Scope, "path": d.Name,
			"fieldManager": d.FieldManager, "reasonRef": d.ReasonRef,
			"createdAt": d.CreatedAt.UTC().Format(time.RFC3339Nano),
			"updatedAt": d.UpdatedAt.UTC().Format(time.RFC3339Nano),
			"content":   string(d.Content),
		})
	})
	// 落稿（upsert，写面门控=引擎装配）：{repo, path, content, fieldManager, reasonRef}。
	mux.HandleFunc("POST /v1/design-docs/draft", func(w http.ResponseWriter, r *http.Request) {
		if designDocs == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "设计文档面未装配（cored 缺 --design-repo）"})
			return
		}
		if !designDocs.WriteEnabled() {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "设计文档写面未开启（引擎未装配——--design-repo 需随 --allow-apply）"})
			return
		}
		if drafts == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "草稿缓存未配置"})
			return
		}
		var req struct {
			Repo         string `json:"repo"`
			Path         string `json:"path"`
			Content      string `json:"content"`
			FieldManager string `json:"fieldManager"`
			ReasonRef    string `json:"reasonRef"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<20)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("请求体不可解析: %v", err)})
			return
		}
		if req.Repo == "" || req.Path == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "repo 与 path 必填"})
			return
		}
		// 白名单闸前置（草稿也只收白名单内 .md——越权路径 403 与写面同口径）。
		if _, err := designDocs.Read(req.Repo, req.Path); err != nil && !errors.Is(err, docrepo.ErrNotFound) {
			// HEAD 缺席=新文件草稿合法；其余（未注册仓/越权路径）原样映射。
			writeDesignDocErr(w, err)
			return
		}
		kind, scope, name := draftDocKey(req.Repo, req.Path)
		d := &draft.Draft{
			Kind: kind, Scope: scope, Name: name, Content: []byte(req.Content),
			FieldManager: req.FieldManager, ReasonRef: req.ReasonRef, BaseRV: 0, // baseSha 在 apply 请求随行（草稿缓存不承载 sha）
		}
		if err := drafts.Put(d); err != nil {
			writeDraftErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"resource": docrepo.ObjectKey(req.Repo, req.Path), "action": "staged",
		})
	})
	// 弃稿（写面门控；幂等 404）。
	mux.HandleFunc("DELETE /v1/design-docs/draft/{repo}/{path...}", func(w http.ResponseWriter, r *http.Request) {
		if designDocs == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "设计文档面未装配（cored 缺 --design-repo）"})
			return
		}
		if !designDocs.WriteEnabled() {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "设计文档写面未开启（引擎未装配——--design-repo 需随 --allow-apply）"})
			return
		}
		if drafts == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "草稿缓存未配置"})
			return
		}
		repo, path := r.PathValue("repo"), r.PathValue("path")
		kind, scope, name := draftDocKey(repo, path)
		if err := drafts.Delete(kind, scope, name); err != nil {
			writeDraftErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"resource": docrepo.ObjectKey(repo, path), "action": "discarded",
		})
	})
	// Apply（写面门控；dry-run 三检查零副作用；fromDraft 取稿/清稿）。
	mux.HandleFunc("POST /v1/design-docs/apply", func(w http.ResponseWriter, r *http.Request) {
		if designDocs == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "设计文档面未装配（cored 缺 --design-repo）"})
			return
		}
		if !designDocs.WriteEnabled() {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "设计文档写面未开启（引擎未装配——--design-repo 需随 --allow-apply）"})
			return
		}
		var req designApplyRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<20)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("请求体不可解析: %v", err)})
			return
		}
		if req.Repo == "" || req.Path == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "repo 与 path 必填"})
			return
		}
		if req.FromDraft && req.Content != "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "fromDraft=true 不携带 content（差异以草稿缓存为准）"})
			return
		}
		content := []byte(req.Content)
		if req.FromDraft {
			if drafts == nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "草稿缓存未配置"})
				return
			}
			kind, scope, name := draftDocKey(req.Repo, req.Path)
			d, err := drafts.Get(kind, scope, name)
			if err != nil {
				writeDesignDocErr(w, err) // 无稿=404（与 noop 可区分）
				return
			}
			content = d.Content
		}
		res, err := designDocs.Apply(r.Context(), docrepo.ApplyRequest{
			Repo: req.Repo, Path: req.Path, BaseSha: req.BaseSha, Content: content,
			FieldManager: req.FieldManager, ReasonRef: req.ReasonRef, DryRun: req.DryRun,
			Subject: subjectFrom(r, req.Subject),
		})
		if err != nil {
			writeDesignDocErr(w, err)
			return
		}
		// rule 6「apply 后清空」：create/update/noop 都算确认消费（dry-run 不清稿）。
		if req.FromDraft && !req.DryRun {
			kind, scope, name := draftDocKey(req.Repo, req.Path)
			if drafts != nil {
				_ = drafts.Delete(kind, scope, name)
			}
		}
		out := designApplyResult(res)
		out["resource"] = docrepo.ObjectKey(req.Repo, req.Path)
		out["dryRun"] = req.DryRun
		// 第三段·属性回整（session.md §5：apply 触发默认自动；dry-run 零副作用不回整）。
		if !req.DryRun {
			if store, ok := objects.(repository.SessionConfigStore); ok {
				if cfg, err := store.GetSessionConfig(r.Context()); err == nil && cfg.MetaAutoUpdate {
					subj := subjectFrom(r, req.Subject)
					wb := sessionMetaWriteback(r.Context(), designDocs, store, req.Repo, req.Path,
						subj.ID(), req.SessionID, req.MetaTags)
					out["writeback"] = wb
				}
			}
		}
		writeJSON(w, http.StatusOK, out)
	})
}
