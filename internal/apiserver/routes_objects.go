// routes_objects.go —— 对象域路由组（R-ENG-7 server.go 拆分首件：纯搬家零语义变化——
// objects 读面 / apply 写面 / secrets 秘密面 / drafts 草稿面与各自请求体、错误映射 helper
// 自 server.go 原位迁入，handler 体逐字节保持不动；装配与公共 helper 仍在 server.go）。
package apiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/apply"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/draft"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventpub"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/flowbind"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
)

// registerObjectRoutes 对象域路由注册（objects / apply / secrets / drafts——纯搬家）。
func registerObjectRoutes(d *deps, mux *http.ServeMux) {
	sch := d.sch
	objects := d.objects
	applier := d.applier
	drafts := d.drafts
	o := d.o
	secrets := o.secrets
	approval := o.approval
	applyLog := o.applyLog
	inbox := o.inbox
	logApproval := d.logApproval
	mux.HandleFunc("GET /v1/objects/{kind}", func(w http.ResponseWriter, r *http.Request) {
		if objects == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "对象存储未配置（cored 缺 --docs）"})
			return
		}
		kind, scope := r.PathValue("kind"), r.URL.Query().Get("scope")
		if _, err := sch.Lookup(kind); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
			return
		}
		objs, err := listMetaOrList(objects, r.Context(), kind, scope) // G-15：清单投影只需键/rv——元数据直读优先
		if err != nil {
			writeObjErr(w, err)
			return
		}
		items := make([]map[string]any, 0, len(objs))
		for _, o := range objs {
			items = append(items, objRef(o))
		}
		writeJSON(w, http.StatusOK, map[string]any{"count": len(items), "kind": kind, "scope": scope, "objects": items})
	})
	mux.HandleFunc("GET /v1/objects/{kind}/{name}", func(w http.ResponseWriter, r *http.Request) {
		if objects == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "对象存储未配置（cored 缺 --docs）"})
			return
		}
		kind, name := r.PathValue("kind"), r.PathValue("name")
		scope := r.URL.Query().Get("scope")
		if _, err := sch.Lookup(kind); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
			return
		}
		o, err := objects.Get(r.Context(), kind, scope, name)
		if err != nil {
			writeObjErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"kind": o.Kind, "name": o.Name, "scope": o.Scope,
			"resourceVersion": o.ResourceVersion, "content": string(o.Content),
		})
	})
	// 删除面（2026-11-26 假数据清理批）：DELETE /v1/objects/{kind}/{name}?scope=s——
	// 写面同门（--allow-apply 关=405）；能力面（后端未实现 ObjectDeleter=501 显式缺席，
	// git 后端现况）；主体=X-Core-Subject 头（同 apply），权限=write×对象 scope（与 apply
	// 同域判定）；expectRV 查询参数=乐观锁期望（0=不设防）。删除落三账：doc_history 终态行
	// +ops apply_log "deleted"+DocDeleted 事件（类型未注册=best-effort 记日志不回滚——
	// 删除已成事实，账面缺口可见但不阻塞）。
	mux.HandleFunc("DELETE /v1/objects/{kind}/{name}", func(w http.ResponseWriter, r *http.Request) {
		if applier == nil {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "删除写路径未开启（cored 缺 --allow-apply；缺省只读）"})
			return
		}
		if objects == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "对象存储未配置（cored 缺 --docs）"})
			return
		}
		del, ok := objects.(repository.ObjectDeleter)
		if !ok {
			writeJSON(w, http.StatusNotImplemented, map[string]any{"error": "存储后端未实现删除能力（git 后端缺席——sqlite docs DB 后端在册）"})
			return
		}
		kind, name := r.PathValue("kind"), r.PathValue("name")
		if kind == "Workflow" && sch != nil && sch.IsUndeletableWorkflow(name) {
			writeJSON(w, http.StatusForbidden, map[string]any{"error": "内置核心工作流禁止删除 (undeletable: " + name + ")"})
			return
		}
		scope := r.URL.Query().Get("scope")
		if scope == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "repository: scope 与 storage 模板不相容: 写入/读取需显式 scope"})
			return
		}
		if _, err := sch.Lookup(kind); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
			return
		}
		sub := subjectFrom(r, nil)
		if o.perm != nil {
			if err := o.perm.Check(*sub, scope, "write", permission.ByRole); err != nil {
				writeJSON(w, http.StatusForbidden, map[string]any{"error": "删除被权限判定拒绝: " + err.Error()})
				return
			}
		}
		var expectRV int64
		if v := r.URL.Query().Get("expectRV"); v != "" {
			if n, perr := strconv.ParseInt(v, 10, 64); perr == nil && n > 0 {
				expectRV = n
			}
		}
		fm := r.URL.Query().Get("fieldManager")
		if fm == "" {
			fm = "console:delete"
		}
		rv, err := del.Delete(kind, scope, name, expectRV, fm, r.URL.Query().Get("reasonRef"))
		if err != nil {
			writeObjErr(w, err)
			return
		}
		// DocDeleted 事件（best-effort）：involved=被删对象锚；publisher 缺席/类型未注册只记
		// 日志——删除已落账（doc_history+apply_log），事件缺口经日志可见。
		if d.publisher != nil {
			if _, perr := d.publisher.Publish(eventpub.Request{
				Type: "DocDeleted", Involved: fmt.Sprintf("%s/%s/%s", kind, scope, name),
				DedupKey: fmt.Sprintf("doc-deleted:%s/%s/%s", kind, scope, name),
			}, sub); perr != nil {
				log.Printf("[apiserver] DocDeleted 事件发布失败（%s/%s/%s）: %v", kind, scope, name, perr)
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"deleted": true, "kind": kind, "scope": scope, "name": name, "resourceVersion": rv})
	})
	// 文档导出独立配置面（2026-11-27 用户裁决：文档导出作为独立配置项存储在数据库 docs_export_config，
	// 初值由 scheme.yaml docTypes[].exportDefault 派生，绝不在导出器代码中硬编码过滤某类型）：
	// GET 恒开（存储未实现 ExportConfigStore 时返回 scheme 派生默认配置）；PUT 随 --allow-apply 门控。
	mux.HandleFunc("GET /v1/docs-export/config", func(w http.ResponseWriter, r *http.Request) {
		if store, ok := objects.(repository.ExportConfigStore); ok {
			cfg, err := store.GetExportConfig(r.Context())
			if err != nil {
				writeObjErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, cfg)
			return
		}
		writeJSON(w, http.StatusOK, repository.DefaultExportConfig(sch))
	})
	mux.HandleFunc("PUT /v1/docs-export/config", func(w http.ResponseWriter, r *http.Request) {
		if applier == nil {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "导出配置写路径未开启（cored 缺 --allow-apply；缺省只读）"})
			return
		}
		store, ok := objects.(repository.ExportConfigStore)
		if !ok {
			writeJSON(w, http.StatusNotImplemented, map[string]any{"error": "当前存储后端不支持持久化导出配置（需 sqlite docs DB）"})
			return
		}
		var req repository.ExportConfig
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("请求体不可解析: %v", err)})
			return
		}
		saved, err := store.SetExportConfig(r.Context(), req)
		if err != nil {
			writeObjErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, saved)
	})
	// 写入口（缺省关闭）：POST /v1/apply——收敛到 apply 管线三步闸门，不旁路存储。
	// 未开启（applier=nil）时恒 405，缺省只读面语义不变。
	// 主体（v0.9 请求级）：请求体 subject 字段（优先）或 X-Core-Subject 头（同形 JSON）；
	// 两者皆缺=匿名主体照判（无 role 落点 → 403，进程级本地信任不兜底 HTTP 匿名请求）。
	// fromDraft=true：稿→面落库（cored-bridge.md §2.2）——差异以草稿缓存为准，不携带 content。
	mux.HandleFunc("POST /v1/apply", func(w http.ResponseWriter, r *http.Request) {
		if applier == nil {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "apply 写路径未开启（cored 缺 --allow-apply；缺省只读）"})
			return
		}
		var req applyRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<20)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("请求体不可解析: %v", err)})
			return
		}
		opt := apply.Options{
			DryRun: req.DryRun, Force: req.Force,
			FieldManager: req.FieldManager, ReasonRef: req.ReasonRef,
			Subject: resolveSubject(r, &req),
		}
		var res *apply.Result
		if req.FromDraft {
			if req.Content != "" {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "fromDraft=true 不携带 content（差异以草稿缓存为准）"})
				return
			}
			if req.ResourceVersion != 0 {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "fromDraft=true 不携带 resourceVersion（乐观锁期望值以草稿 baseResourceVersion 为准）"})
				return
			}
			if req.Kind == "" || req.Scope == "" || req.Name == "" {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "对象键不全（需 kind、scope、name）"})
				return
			}
			var err error
			res, err = applier.ApplyDraft(r.Context(), req.Kind, req.Scope, req.Name, opt)
			if err != nil {
				syncDeniedJudgmentRow(req, opt, err, inbox) // v0.57：denied 判定行随行同步（观测投影零旁路——账本照落）
				writeApplyErr(w, err)
				return
			}
		} else {
			obj := &repository.Object{
				Kind: req.Kind, Name: req.Name, Scope: req.Scope,
				ResourceVersion: req.ResourceVersion, Content: []byte(req.Content),
			}
			var err error
			res, err = applier.Apply(r.Context(), obj, opt)
			if err != nil {
				syncDeniedJudgmentRow(req, opt, err, inbox) // v0.57：denied 判定行随行同步（观测投影零旁路——账本照落）
				writeApplyErr(w, err)
				return
			}
		}
		resp := map[string]any{
			"resource": fmt.Sprintf("%s/%s/%s", req.Kind, req.Scope, req.Name),
			"rv":       res.RV, "action": res.Action, "dryRun": req.DryRun,
		}
		// 审批判定落账（v0.46 审批独立账，ops-log.md §2；v0.51 对称扩展）：console:approve 批准 /
		// console:reject 驳回判定——fromDraft ∧ 非 dry-run ∧ 判定成功即落对应行（dry-run=预检非判定；
		// 非 fromDraft=直接落库非审批屏动作；noop 幂等重入照落，rv=面上现值——判定发生了，零变更可落）。
		decision := ""
		switch req.FieldManager {
		case ApproveFieldManager, flowbind.ApprovalManager: // console:approve | workflow:approve（v0.56 派单批准动作面——§11.5）
			decision = opslog.DecisionGranted
		case RejectFieldManager, flowbind.ApprovalRejecter: // console:reject | workflow:reject
			decision = opslog.DecisionRejected
		}
		// 派单 ask 批准动作面（v0.56 R-10——workflow-runtime.md §11.5 批准投影翻转判定源）：
		// workflow:approve|workflow:reject 溯源串行=派单批准判定（Apply ok/noop 成功事实即翻转
		// 依据——flowbind.SyncApprovals.Sync 消费 apply_log 行集）。判定权威=apply_log 审计行；
		// 本翻转面=观测投影同步（失败不翻转写路径结果，下一拍派单查询面重扫自愈）。
		// v0.58 同拍判定行在账（workflow-runtime.md §11.5 动作面 console 动作化——判定行是
		// 账本事实，动作返回后立即读得到：活体判定窗「批准动作 → inbox 下一次装载反映」等价
		// 缩到「动作返回拍」——历拍重扫自愈路径保持不变作兜底）。判据语义锚不变（记录新增行）。
		if inbox != nil && decision != "" && !req.DryRun && res != nil && applyLog != nil {
			if rows, err := applyLog.QueryApplies(fmt.Sprintf("%s/%s/%s", req.Kind, req.Scope, req.Name), 5); err == nil {
				for _, row := range rows {
					if row.FieldManager == req.FieldManager {
						inbox.Sync([]opslog.ApplyLog{row}) // v0.58 判定行在账（含 denied 形账本行零旁路——匹配判定不发明）
						break
					}
				}
			}
		}
		if inbox != nil && decision != "" && !req.DryRun && res != nil {
			row := opslog.ApplyLog{
				Ts: time.Now(), FieldManager: req.FieldManager, Subject: func() string {
					if opt.Subject != nil {
						return opt.Subject.ID()
					}
					return "anonymous"
				}(), Obj: fmt.Sprintf("%s/%s/%s", req.Kind, req.Scope, req.Name),
				ResourceVersion: res.RV, Result: func() string {
					if res.Action == "noop" {
						return "noop"
					}
					return "ok"
				}(),
			}
			inbox.Sync([]opslog.ApplyLog{row}) // 同步翻转（观测投影——失败仅丢本轮，下一拍重扫自愈）
		}
		if approval != nil && req.FromDraft && !req.DryRun && decision != "" {
			judged := "anonymous"
			if opt.Subject != nil {
				judged = opt.Subject.ID()
			}
			objKey := fmt.Sprintf("%s/%s/%s", req.Kind, req.Scope, req.Name)
			logApproval(opslog.ApprovalLog{
				Ts: time.Now(), ApprovalID: "draft:" + objKey,
				Obj: objKey, Decision: decision,
				JudgedBy: judged, FieldManager: req.FieldManager,
				ResourceVersion: res.RV, ReasonRef: req.ReasonRef,
			})
			// 审批信号族事件随行（v0.51，event.md 规则 11/15——判定是人经 console 做出的，
			// 经发布面同管线 source=human；involved=判定对象——非 Workflow 引用不命中 run 域迁移，
			// dispatch 派发面对审批族无目标=dead-letter 行，零副作用；发布失败不翻转 apply 结果——
			// 事件是通知，判定权威=approval_log 行）。
			if d.publisher != nil && req.Kind == "Proposal" && opt.Subject != nil {
				typ := "ApprovalGranted"
				if decision == opslog.DecisionRejected {
					typ = "ApprovalRejected"
				}
				if et, err := sch.LookupEventType(typ); err == nil { // 注册表闸门前置
					if _, err := d.publisher.Publish(eventpub.Request{
						Type: typ, Involved: objKey, DedupKey: "approval:" + objKey, Level: et.DefaultLevel,
					}, opt.Subject); err != nil {
						log.Printf("apiserver: %s 事件发布未遂（非权威，判定不受影响）: %v", typ, err)
					}
				}
			}
		}
		if res.Event != nil {
			resp["event"] = map[string]any{"id": res.Event.ID, "type": res.Event.Type, "dedupKey": res.Event.DedupKey}
		}
		writeJSON(w, http.StatusOK, resp)
	})

	// 秘密存储面（v0.34 secret 存储落地首段，provider.md 规则 2/6 运行时面；cored
	// --secrets-db 装配）：GET /v1/secrets 列表只出名称与更新时刻（**值永不回显**——脱敏
	// 红线）；POST /v1/secrets 落值（受理即吞值，回执零值）；DELETE /v1/secrets/{name} 幂等
	// 删除。写操作与 apply/events 写面同门控（--allow-apply；未开启恒 405=代际判别器）；
	// face 未注入（--secrets-db 未配置）时全部 503（面注册但能力缺席，关断口径可见）。
	mux.HandleFunc("GET /v1/secrets", func(w http.ResponseWriter, _ *http.Request) {
		if secrets == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "秘密存储未配置（cored 缺 --secrets-db）"})
			return
		}
		rows, err := secrets.List()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		type row struct {
			Name      string    `json:"name"`
			UpdatedAt time.Time `json:"updatedAt"`
		}
		out := make([]row, 0, len(rows))
		for _, r0 := range rows {
			out = append(out, row{Name: r0.Name, UpdatedAt: r0.UpdatedAt})
		}
		writeJSON(w, http.StatusOK, map[string]any{"count": len(out), "secrets": out})
	})
	mux.HandleFunc("POST /v1/secrets", func(w http.ResponseWriter, r *http.Request) {
		if applier == nil {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "写路径未开启（cored 缺 --allow-apply；缺省只读）"})
			return
		}
		if secrets == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "秘密存储未配置（cored 缺 --secrets-db）"})
			return
		}
		var req struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("请求体不可解析: %v", err)})
			return
		}
		if req.Name == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "name 必填"})
			return
		}
		if err := secrets.Set(req.Name, req.Value); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		// 回执零值（受理即吞值——请求体里的 value 不出现在任何响应/留痕）。
		writeJSON(w, http.StatusOK, map[string]any{"name": req.Name, "set": true})
	})
	mux.HandleFunc("DELETE /v1/secrets/{name}", func(w http.ResponseWriter, r *http.Request) {
		if applier == nil {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "写路径未开启（cored 缺 --allow-apply；缺省只读）"})
			return
		}
		if secrets == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "秘密存储未配置（cored 缺 --secrets-db）"})
			return
		}
		name := r.PathValue("name")
		if name == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "name 必填"})
			return
		}
		if err := secrets.Delete(name); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"name": name, "deleted": true})
	})

	// 草稿缓存面（cored-bridge.md §2.2；apply.md §1 编辑与发布分离）：GET 属查询面恒开；
	// 写操作（落稿/弃稿）与 fromDraft apply 随写面开关（缺省 405）。
	mux.HandleFunc("GET /v1/drafts", func(w http.ResponseWriter, _ *http.Request) {
		if drafts == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "草稿缓存未配置"})
			return
		}
		items := make([]map[string]any, 0)
		for _, d := range drafts.List() {
			items = append(items, draftRef(d))
		}
		writeJSON(w, http.StatusOK, map[string]any{"count": len(items), "drafts": items})
	})
	mux.HandleFunc("GET /v1/drafts/{kind}/{name}", func(w http.ResponseWriter, r *http.Request) {
		if drafts == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "草稿缓存未配置"})
			return
		}
		kind, name := r.PathValue("kind"), r.PathValue("name")
		scope := r.URL.Query().Get("scope")
		d, err := drafts.Get(kind, scope, name)
		if err != nil {
			writeDraftErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"kind": d.Kind, "name": d.Name, "scope": d.Scope,
			"fieldManager": d.FieldManager, "reasonRef": d.ReasonRef,
			"baseResourceVersion": d.BaseRV,
			"createdAt":           d.CreatedAt.UTC().Format(time.RFC3339Nano),
			"updatedAt":           d.UpdatedAt.UTC().Format(time.RFC3339Nano),
			"content":             string(d.Content),
		})
	})
	mux.HandleFunc("POST /v1/drafts", func(w http.ResponseWriter, r *http.Request) {
		if applier == nil {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "草稿写面未开启（cored 缺 --allow-apply；草稿读面恒开）"})
			return
		}
		if drafts == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "草稿缓存未配置"})
			return
		}
		var req draftRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<20)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("请求体不可解析: %v", err)})
			return
		}
		d := &draft.Draft{
			Kind: req.Kind, Scope: req.Scope, Name: req.Name,
			Content: []byte(req.Content), FieldManager: req.FieldManager,
			ReasonRef: req.ReasonRef, BaseRV: req.BaseResourceVersion,
		}
		if err := drafts.Put(d); err != nil {
			writeDraftErr(w, err) // ErrInvalid → 400（键不全/无主/空稿）
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"resource": fmt.Sprintf("%s/%s/%s", req.Kind, req.Scope, req.Name),
			"action":   "staged", "baseResourceVersion": req.BaseResourceVersion,
		})
	})
	mux.HandleFunc("DELETE /v1/drafts/{kind}/{name}", func(w http.ResponseWriter, r *http.Request) {
		if applier == nil {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "草稿写面未开启（cored 缺 --allow-apply；草稿读面恒开）"})
			return
		}
		if drafts == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "草稿缓存未配置"})
			return
		}
		kind, name := r.PathValue("kind"), r.PathValue("name")
		scope := r.URL.Query().Get("scope")
		if err := drafts.Delete(kind, scope, name); err != nil {
			writeDraftErr(w, err)
			return
		}
		// 审批判定落账（v0.46 审批独立账）：驳回=弃稿（零文档面动作）——判定发生了，账面落行
		// 才可审计（弃稿动作此前在 ops DB 零留痕）。judged_by=请求主体头（DELETE 无 body）；
		// 幂等 404（本就无稿）不落账——无判定发生。
		if approval != nil {
			logApproval(opslog.ApprovalLog{
				Ts: time.Now(), ApprovalID: "draft:" + fmt.Sprintf("%s/%s/%s", kind, scope, name),
				Obj: fmt.Sprintf("%s/%s/%s", kind, scope, name), Decision: opslog.DecisionRejected,
				JudgedBy: subjectFrom(r, nil).ID(), FieldManager: "draft:delete",
				ReasonRef: "draft-delete",
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"resource": fmt.Sprintf("%s/%s/%s", kind, scope, name), "action": "discarded",
		})
	})
}

// applyRequest POST /v1/apply 请求体（cored-bridge.md 契约；content=文档原文）。
type applyRequest struct {
	Kind            string          `json:"kind"`
	Scope           string          `json:"scope"`
	Name            string          `json:"name"`
	ResourceVersion int64           `json:"resourceVersion"`
	Content         string          `json:"content"`
	FieldManager    string          `json:"fieldManager"`
	ReasonRef       string          `json:"reasonRef"`
	DryRun          bool            `json:"dryRun"`
	Force           bool            `json:"force"`
	FromDraft       bool            `json:"fromDraft"` // true=稿→面（§2.2）；content/resourceVersion 须缺省
	Subject         *subjectRequest `json:"subject"`   // 请求级主体（v0.9）；缺席回退 X-Core-Subject 头
}

// subjectRequest 请求级主体（v0.9，permission.md §3 的 HTTP 投影）：
// user=主体标识（user-system 目录侧未接前由调用方断言，审计留痕）；
// roles/groups=主体断言的持有关系——权限只由文档面 Role/Group 对象授予，断言名须能对上装载的对象。
type subjectRequest struct {
	User   string   `json:"user"`
	Roles  []string `json:"roles"`
	Groups []string `json:"groups"`
}

// SubjectHeader 主体头名（与请求体 subject 字段同形 JSON；桥接方不便改体时用头）。
const SubjectHeader = "X-Core-Subject"

// ApproveFieldManager 批准判定的 fieldManager 约定串（E6 审批屏批准=Apply 全流程复用；
// v0.46 审批独立账判定识别键——console 客户端与 cored 判定面共用此约定，双账并存：
// apply_log 承落库流水，approval_log 承审批判定）。
const ApproveFieldManager = "console:approve"

// RejectFieldManager 驳回判定的 fieldManager 约定串（v0.51 能力 B 对称扩展——Proposal 屏
// 驳回=对象态写 rejected 的判定动作（rv=落库后版本，approval_log rejected 行+ApprovalRejected
// 事件）；与弃稿 DELETE（rv=0 零文档面动作）两分——两种「不采纳」的判定都在账）。
const RejectFieldManager = "console:reject"

// subjectFrom 请求级主体解析（apply/events 写面共用）：subject 字段优先，缺席回退
// X-Core-Subject 头；两者皆缺 → 匿名主体（非 nil 指针=强制请求级判定，进程级本地信任
// 不兜底 HTTP 匿名请求）。
func subjectFrom(r *http.Request, sr *subjectRequest) *permission.Subject {
	if sr != nil {
		return &permission.Subject{User: sr.User, Roles: sr.Roles, Groups: sr.Groups}
	}
	if h := r.Header.Get(SubjectHeader); h != "" {
		var s subjectRequest
		if err := json.Unmarshal([]byte(h), &s); err == nil {
			return &permission.Subject{User: s.User, Roles: s.Roles, Groups: s.Groups}
		}
		// 头不可解析=视为未携带（匿名照判拒绝；不静默降级成本地信任）。
	}
	return &permission.Subject{}
}

// resolveSubject POST /v1/apply 请求级主体解析（subjectFrom 的 apply 侧入口）。
func resolveSubject(r *http.Request, req *applyRequest) *permission.Subject {
	return subjectFrom(r, req.Subject)
}

// draftRequest POST /v1/drafts 请求体（cored-bridge.md §2.2）。
type draftRequest struct {
	Kind                string `json:"kind"`
	Scope               string `json:"scope"`
	Name                string `json:"name"`
	Content             string `json:"content"`
	FieldManager        string `json:"fieldManager"`
	ReasonRef           string `json:"reasonRef"`
	BaseResourceVersion int64  `json:"baseResourceVersion"` // 落稿所基于的文档面 rv（0=预期创建）
}

// writeApplyErr Apply 管线错误 → HTTP 状态码映射（cored-bridge.md §2.1 错误表）。
func writeApplyErr(w http.ResponseWriter, err error) {
	var conflict *repository.ConflictError
	switch {
	case errors.As(err, &conflict):
		writeJSON(w, http.StatusConflict, map[string]any{
			"error": err.Error(), "expected": conflict.Expected, "current": conflict.Current,
		})
	case errors.Is(err, apply.ErrNotRegistered), errors.Is(err, repository.ErrNotRegistered),
		errors.Is(err, repository.ErrNotFound), errors.Is(err, draft.ErrNoDraft):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
	case errors.Is(err, apply.ErrDenied):
		writeJSON(w, http.StatusForbidden, map[string]any{"error": err.Error()})
	case errors.Is(err, apply.ErrNoFieldManager), errors.Is(err, apply.ErrObject),
		errors.Is(err, apply.ErrSpecInvalid),
		errors.Is(err, repository.ErrScope), errors.Is(err, repository.ErrIdentity),
		errors.Is(err, repository.ErrContent), errors.Is(err, repository.ErrAlreadyExists),
		errors.Is(err, repository.ErrNoFieldManager), errors.Is(err, repository.ErrCredentialSource):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
	case errors.Is(err, repository.ErrClosed):
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
}

// writeDraftErr 草稿面错误 → HTTP 状态码映射（cored-bridge.md §2.2）。
func writeDraftErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, draft.ErrNoDraft):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
	case errors.Is(err, draft.ErrInvalid):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
	case errors.Is(err, draft.ErrNoStore):
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
}

// syncDeniedJudgmentRow denied 判定行随行同步（v0.57——批准投影翻转判定的 denied 词形补口：
// apply_log 审计=判定权威（引擎 denied 行照落——apply.go audit），观测投影此前只喂 ok/noop
// 行=denied 判定缺席的投影缺口（判定行重扫词形与账本行集分叉）。本函数只投影账本既有事实，
// 零新写路径零旁路（Sync 匹配判定零翻转——denied 不判定，TestSyncApprovalsFlipByApplyRows
// 锚定）；applyLog face 缺席/非判定串=静默零投影）。
func syncDeniedJudgmentRow(req applyRequest, opt apply.Options, err error, face InboxFace) {
	if face == nil {
		return
	}
	switch req.FieldManager {
	case ApproveFieldManager, flowbind.ApprovalManager, RejectFieldManager, flowbind.ApprovalRejecter:
	default:
		return // 非判定溯源串=零投影（普通 apply 拒绝不参与批准判定面）
	}
	if !errors.Is(err, apply.ErrDenied) && !errors.Is(err, apply.ErrNotRegistered) {
		return // 非 deny 族错误（conflict/传输层）不投——判定行词形=denied 族
	}
	sub := "anonymous"
	if opt.Subject != nil {
		sub = opt.Subject.ID()
	}
	face.Sync([]opslog.ApplyLog{{
		Ts: time.Now(), FieldManager: req.FieldManager, Subject: sub,
		Obj:    fmt.Sprintf("%s/%s/%s", req.Kind, req.Scope, req.Name),
		Result: "denied", PolicyDecisions: err.Error(),
	}})
}
