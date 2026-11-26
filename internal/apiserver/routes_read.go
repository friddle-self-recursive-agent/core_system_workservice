// routes_read.go —— 读面路由组（R-ENG-7 server.go 拆分：纯搬家零语义变化——metrics/
// history/rev-content/doc-history/approvals/tasks/proposals 与 design_docs、intake、
// connectops、context 各 face 注册自 server.go 原位迁入，handler 体逐字节保持不动；
// 装配与公共 helper 仍在 server.go）。
package apiserver

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/metricsinfo"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
)

// registerReadRoutes 读面路由注册（纯搬家；子 face 的整段注册器在此统一收口）。
func registerReadRoutes(d *deps, mux *http.ServeMux) {
	sch := d.sch
	objects := d.objects
	o := d.o
	metrics := o.metrics
	applyLog := o.applyLog
	revContent := o.revContent
	docHistory := o.docHistory
	approval := o.approval
	// 数据面查询面（v0.33 MetricsInfo 数据面首段——store/queries 投影读面）：指标清单
	// （样本保留数，排序确定性投影）与窗口聚合查询（queries 最小集投影）。数据面未装配
	// =空投影（关断口径可见）；查询未命中/声明面外聚合=404（不冒充零值）。数据面零信号
	// 面依赖同口径——本面只读样本库，零总线零判定。
	mux.HandleFunc("GET /v1/metrics", func(w http.ResponseWriter, _ *http.Request) {
		items := []map[string]any{}
		if metrics != nil {
			for _, name := range metrics.Metrics() {
				items = append(items, map[string]any{"name": name, "samples": metrics.Count(name)})
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"metrics": items, "total": len(items)})
	})
	mux.HandleFunc("GET /v1/metrics/query", func(w http.ResponseWriter, r *http.Request) {
		qm := r.URL.Query().Get("metric")
		qw := r.URL.Query().Get("window")
		qa := r.URL.Query().Get("agg")
		if qm == "" || qa == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "metrics 查询需 metric 与 agg 参数"})
			return
		}
		if metrics == nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "数据面未装配（关断口径）"})
			return
		}
		v, ok := metrics.Query(metricsinfo.Query{Metric: qm, Window: qw, Agg: qa})
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "无投影结果（未知聚合或无同键样本）", "metric": qm, "window": qw, "agg": qa})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"metric": qm, "window": qw, "agg": qa, "value": v})
	})
	// 对象应用史投影面（v0.38 console 读面第三段，E4——cored-bridge.md §2 `GET /v1/history/{kind}/{name}`）：
	// 数据源=ops DB apply_log（ops-log.md §2 一次 Apply 判定的流水；repository 双写 ok/noop/conflict
	// 行+引擎 denied 行）——rev 流水的权威=审计账（git 权威在正文，不在此考古）。obj=kind/scope/name
	// 精确过滤+rv 降序+limit 收敛（≤500 防拖库同口径）。只读 GET；face 缺席=503（防御面，装配层
	// 随 ops DB 恒注入）；旧构建（路由未注册）经兜底 handler 405=代际判别器。
	mux.HandleFunc("GET /v1/history/{kind}/{name}", func(w http.ResponseWriter, r *http.Request) {
		if applyLog == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "应用史查询面未接线（装配缺 ops DB 口面）"})
			return
		}
		kind, name := r.PathValue("kind"), r.PathValue("name")
		scope := r.URL.Query().Get("scope")
		obj := fmt.Sprintf("%s/%s/%s", kind, scope, name)
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
	// 历史 rv 正文投影面（v0.63 rev-content diff 面，§1.40——历史 tab「点行 diff 上一版」余量
	// 收口的读半边）：`GET /v1/rev-content/{kind}/{name}?scope=s&rv=N` 按对象键+目标 rv 取该
	// 次时点权威正文快照。双后端分立词形（source 字段据实标注，不冒充同源）：git 后端=裸库
	// 提交考古（concepts/revision.md §3 git=内容版本权威）；sqlite 后端=doc_history 逐次账
	// （R-DOC-7 规则 4——周期 git 导出快照不得冒充逐次历史）。只读 GET 零副作用；
	// face 缺席=503 防御面（旧装配——诚实降级非 200 假面）；旧构建 405=代际判别器；
	// 对象键不可解析/缺席 → 400/404 分立；rv 缺席或不可解析 → 400；该 rv 无落库快照 → 404
	// ErrRevNotFound 分立词形（与对象缺席分立——审计不冒充）。
	mux.HandleFunc("GET /v1/rev-content/{kind}/{name}", func(w http.ResponseWriter, r *http.Request) {
		if revContent == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "历史 rv 正文读面未接线（face 缺席——旧装配诚实降级）"})
			return
		}
		kind, name := r.PathValue("kind"), r.PathValue("name")
		scope := r.URL.Query().Get("scope")
		rawRV := r.URL.Query().Get("rv")
		rv, err := strconv.ParseInt(rawRV, 10, 64)
		if err != nil || rv <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("rv 必须为正整数（乐观锁词形），得 %q", rawRV)})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		content, err := revContent.RevContent(ctx, kind, scope, name, rv)
		switch {
		case err == nil:
			writeJSON(w, http.StatusOK, map[string]any{
				"kind": kind, "scope": scope, "name": name, "rv": rv, "content": content,
				"source": revContent.Source(),
			})
		case errors.Is(err, repository.ErrRevNotFound):
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error(), "obj": fmt.Sprintf("%s/%s/%s", kind, scope, name), "rv": rv})
		case errors.Is(err, repository.ErrNotFound):
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
		case errors.Is(err, repository.ErrScope):
			writeJSON(w, http.StatusBadRequest, map[string]error{"error": err})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		}
	})
	// 逐次正文历史清单投影面（R-DOC-7 规则 4 doc_history 落地面——`GET /v1/doc-history/
	// {kind}/{name}?scope=s`）：逐次翻页行集（rev 降序：rev/翻页时刻/写者/正文字节数——
	// 不含正文，正文经 GET /v1/rev-content?rv=N 点取）。与 GET /v1/history（apply_log 审计
	// 流水——谁/何时/rv/结果，ops-log.md §2 双轨分工）词形分立：审计面管操作流水，本面管
	// 逐次正文翻页账。只读 GET 零副作用；face 缺席=503（git 后端不装配：git 历史=周期导出
	// 快照非逐次账——不冒充）；kind 未注册=404；scope 不合模板=400；对象缺席=404（与在册
	// 无翻页史=空清单 200 分立）；旧构建（路由未注册）经兜底 handler 405=代际判别器。
	mux.HandleFunc("GET /v1/doc-history/{kind}/{name}", func(w http.ResponseWriter, r *http.Request) {
		if docHistory == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "逐次正文历史清单面未接线（doc_history 仅 sqlite 后端装配；git 后端历史=周期导出快照非逐次账——不冒充）"})
			return
		}
		kind, name := r.PathValue("kind"), r.PathValue("name")
		scope := r.URL.Query().Get("scope")
		if _, err := sch.Lookup(kind); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
			return
		}
		rows, err := docHistory.History(kind, scope, name)
		if err != nil {
			writeObjErr(w, err)
			return
		}
		items := make([]map[string]any, 0, len(rows))
		for _, h := range rows {
			items = append(items, map[string]any{
				"rev": h.Rev, "at": h.AppliedAt.UTC().Format(time.RFC3339Nano),
				"fieldManager": h.FieldManager, "bytes": h.Bytes,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"kind": kind, "scope": scope, "name": name,
			"count": len(items), "history": items,
		})
	})
	// 审批判定账面查询面（v0.46 审批独立账，ops-log.md §2 v0.46 注记——`GET /v1/approvals`）：
	// 数据源=ops DB approval_log（审批判定独立账：批准 granted/驳回 rejected——apply_log 承
	// 落库流水，本账承"谁在何时对哪条提案做了什么判定"；驳回=弃稿零文档面动作也落行）。
	// decision/obj 精确过滤+rv 降序+limit 收敛（≤500 防拖库同口径）。只读 GET；face 缺席=503
	// （防御面，装配层随 ops DB 恒注入）；旧构建（路由未注册）经兜底 handler 405=代际判别器。
	mux.HandleFunc("GET /v1/approvals", func(w http.ResponseWriter, r *http.Request) {
		if approval == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "审批判定账面查询面未接线（装配缺 ops DB 口面）"})
			return
		}
		qs := opslog.ApprovalQuery{
			Decision: r.URL.Query().Get("decision"),
			Obj:      r.URL.Query().Get("obj"),
			Limit:    atoiDefault(r.URL.Query().Get("limit"), 0),
		}
		rows, err := approval.QueryApprovals(qs)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		items := make([]map[string]any, 0, len(rows))
		for _, a := range rows {
			items = append(items, map[string]any{
				"ts": a.Ts.UTC().Format(time.RFC3339Nano), "approvalId": a.ApprovalID,
				"obj": a.Obj, "decision": a.Decision, "judgedBy": a.JudgedBy,
				"fieldManager": a.FieldManager, "rv": a.ResourceVersion, "reasonRef": a.ReasonRef,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"count": len(items), "approvals": items})
	})
	// 任务队列读面（concepts/task-scheduler.md §7——cored-bridge.md §2 `GET /v1/tasks`）：
	// 数据源=文档库 Task 对象（零缓存现读）；列表=status 投影简表（?phase= 闭合集过滤+
	// ?scope= 枚举收窄，现势→priority→scope/name 确定性投影）；详情=spec+status 投影+content
	// 原文。只读 GET 恒开（objects 缺席=503；kind 未注册=404——scheme 闸门语义与 proposals 面
	// 同口径）；status 写面不在此（P2 只由调度器/执行体写），任务提交/取消走 apply/events 既有写链。
	mux.HandleFunc("GET /v1/tasks", func(w http.ResponseWriter, r *http.Request) {
		if objects == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "对象存储未配置（cored 缺 --docs）"})
			return
		}
		if _, err := sch.Lookup(taskKind); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
			return
		}
		phaseFilter, scopeQ := r.URL.Query().Get("phase"), r.URL.Query().Get("scope")
		objs, err := objects.List(r.Context(), taskKind, scopeQ)
		if err != nil {
			writeObjErr(w, err)
			return
		}
		rows := make([]taskRow, 0, len(objs))
		for _, o := range objs {
			row := taskRowOf(o.Name, o.Scope, o.ResourceVersion, o.Content)
			if phaseFilter != "" && row.Phase != phaseFilter {
				continue // phase 过滤（闭合集词形——词外值=恒空结果不冒充全量）
			}
			rows = append(rows, row)
		}
		sortTasks(rows)
		writeJSON(w, http.StatusOK, map[string]any{"count": len(rows), "tasks": rows})
	})
	mux.HandleFunc("GET /v1/tasks/{scope}/{name}", func(w http.ResponseWriter, r *http.Request) {
		if objects == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "对象存储未配置（cored 缺 --docs）"})
			return
		}
		scope, name := r.PathValue("scope"), r.PathValue("name")
		if _, err := sch.Lookup(taskKind); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
			return
		}
		o, err := objects.Get(r.Context(), taskKind, scope, name) // 键全给定=直取（404 语义与 objects 面同口径）
		if err != nil {
			writeObjErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, taskDetail(o))
	})
	// 需求提案读面（v0.51 能力 B——cored-bridge.md §2 `GET /v1/proposals`；concepts/proposal.md）：
	// 数据源=文档库 Proposal 对象（零缓存现读）；列表=瘦身投影（?status= 闭合集过滤+?scope=
	// 枚举收窄，rv 降序确定性投影）；详情=全 spec 投影+content 原文（scope 缺省=global 优先→其余
	// 首个）。只读 GET 恒开（objects 缺席=503；kind 未注册=404——scheme 闸门语义）；旧构建（路由
	// 未注册）经兜底 handler 405=代际判别器。写路径零新端点（Proposal 走 apply 管线——proposal.md 规则 1）。
	mux.HandleFunc("GET /v1/proposals", func(w http.ResponseWriter, r *http.Request) {
		if objects == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "对象存储未配置（cored 缺 --docs）"})
			return
		}
		if _, err := sch.Lookup(proposalKind); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
			return
		}
		statusFilter, scopeQ := r.URL.Query().Get("status"), r.URL.Query().Get("scope")
		objs, err := objects.List(r.Context(), proposalKind, scopeQ)
		if err != nil {
			writeObjErr(w, err)
			return
		}
		rows := make([]proposalRow, 0, len(objs))
		for _, o := range objs {
			row := proposalRowOf(o.Name, o.Scope, o.ResourceVersion, o.Content)
			if statusFilter != "" && row.Status != statusFilter {
				continue // 状态过滤（闭合集词形——词外值=恒空结果不冒充全量）
			}
			rows = append(rows, row)
		}
		sortProposals(rows)
		writeJSON(w, http.StatusOK, map[string]any{"count": len(rows), "proposals": rows})
	})
	mux.HandleFunc("GET /v1/proposals/{name}", func(w http.ResponseWriter, r *http.Request) {
		if objects == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "对象存储未配置（cored 缺 --docs）"})
			return
		}
		name, scopeQ := r.PathValue("name"), r.URL.Query().Get("scope")
		if _, err := sch.Lookup(proposalKind); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
			return
		}
		if scopeQ != "" { // 键全给定=直取（404 语义与 objects 面同口径）
			o, err := objects.Get(r.Context(), proposalKind, scopeQ, name)
			if err != nil {
				writeObjErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, proposalDetail(o))
			return
		}
		objs, err := objects.List(r.Context(), proposalKind, "")
		if err != nil {
			writeObjErr(w, err)
			return
		}
		var pick *repository.Object
		for _, o := range objs {
			if o.Name != name {
				continue
			}
			if o.Scope == "global" {
				pick = o
				break // global 同名优先（确定性投影）
			}
			if pick == nil {
				pick = o
			}
		}
		if pick == nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "proposal 未在册（对象键=(kind,scope,name)）"})
			return
		}
		writeJSON(w, http.StatusOK, proposalDetail(pick))
	})
	// 设计文档面（v0.52 能力 A——console 屏幕 6 直接编辑设计文档 md → Apply → git commit）：
	// 读/史/草稿/apply 路由注册（handler 体见 design_docs.go）。面未装配=503；写面随
	// --allow-apply；白名单外 403 / lint 未过 400 携违规 / 基线 sha 不符 409——任务书判据三件。
	registerDesignDocRoutes(mux, o.designDocs, d.drafts, applyLog)
	// 对话式录入与三重 dry-run 预检面（v0.53 能力 D + R-UX-1/R-UX-2——cored-bridge.md §2.5）：
	// GET /v1/intake/sessions* 恒开（旧构建 405=代际判别器）；POST /v1/intake/turn 与 /generate 随 --allow-apply 门控。
	registerIntakeRoutes(mux, sch, objects, d.applier)
	// ops 调试面（R-TYPE-3 rule-verification.md——POST /v1/connectops/{name}/{op}）：
	// 正式通路=workflow Verify 接线（dev-delivery Verifying 相位硬 gate 经 ConnectorOps 调
	// connectops-code-audit audit/findings），待 workflow 侧接入；本路由=通路建成前的最小
	// 调试暴露面（JSON body，subject 记录审计行；只读语义由实现侧契约保证，handler 体见
	// connectops_face.go）。
	registerConnectOpsRoutes(mux)
	// 上下文读面（v0.65 context P0 数据面 + trace 1:N 追踪——context_face.go）：四路由全 GET
	// 恒开（/v1/context/index|find|relevance + /v1/trace/{kind}/{scope}/{name}）；face 缺席=503
	// 防御面；旧构建（路由未注册）经兜底 handler 405=代际判别器。
	registerContextRoutes(mux, o.contextIdx)
}
