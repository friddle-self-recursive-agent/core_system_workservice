// Package apiserver cored 查询面：scheme 加载校验结果 + 对象类型清单 get/list
// + 文档对象读面（/v1/objects，读库内真实文档对象而非类型注册表）。
// 设计依据：core_system_instance/runtime/cored-bridge.md（plugin↔cored 桥契约）。
// 缺省只读；写路径唯一入口是 Apply 管线（core_design/concepts/apply.md）——仅当部署方显式开启
// --allow-apply 时暴露 POST /v1/apply，内部收敛到 apply.Engine 三步闸门，不旁路存储。
package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/apply"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/draft"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
)

// ObjectSource 文档对象读取源（Repository 的读半边；接口化便于测试注入假仓库）。
type ObjectSource interface {
	Get(ctx context.Context, kind, scope, name string) (*repository.Object, error)
	List(ctx context.Context, kind, scope string) ([]*repository.Object, error)
}

// Applier 写路径（apply.Engine 的口面；接口化便于测试注入假引擎）。
type Applier interface {
	Apply(ctx context.Context, o *repository.Object, opt apply.Options) (*apply.Result, error)
	// ApplyDraft 稿→面落库（cored-bridge.md §2.2：POST /v1/apply + fromDraft=true 的引擎侧）。
	ApplyDraft(ctx context.Context, kind, scope, name string, opt apply.Options) (*apply.Result, error)
}

// DraftStaging 草稿缓存口面（draft.Store 满足；接口化便于测试注入假缓存）。
type DraftStaging interface {
	Get(kind, scope, name string) (*draft.Draft, error)
	List() []*draft.Draft
	Put(d *draft.Draft) error
	Delete(kind, scope, name string) error
}

// TriggerFace 触发骨架查询口面（workflow.TriggerHub 满足；接口化便于测试注入假面）。
// dispatch 接线第一段（cored-bridge.md §2.3）：apply 落面→DocApplied 事件→订阅收到留痕；
// v0.13 增 Resolve 派发决策查询（引擎主体第一段：命中行→owner 解析→派发目标投影）；
// v0.14 决策增执行绑定投影（引擎主体第二段：route/cc 目标→执行面 executor/target/wired）。
type TriggerFace interface {
	State() workflow.TriggerState
	Resolve(typ, involved string) workflow.DispatchDecision
}

// RunFace 对账 run 投影查询口面（workflow.RunLog 满足；引擎主体第三段 v0.15——真实拉起
// 第一子段：workflow 触发面接线，run=触发面→workflow 域的执行留痕起点投影）。
type RunFace interface {
	State() workflow.RunState
}

// compile 期校验：git 后端满足读半边，apply 引擎满足写口面，draft.Store 满足草稿口面，
// workflow.TriggerHub 满足触发骨架口面。
var (
	_ ObjectSource = (*repository.GitRepo)(nil)
	_ Applier      = (*apply.Engine)(nil)
	_ DraftStaging = (*draft.Store)(nil)
	_ TriggerFace  = (*workflow.TriggerHub)(nil)
	_ RunFace      = (*workflow.RunLog)(nil)
)

// New 构建查询面 HTTP 处理器。端点：
//
//	GET /healthz                        存活探针
//	GET /v1/scheme                      scheme 概览（加载校验结果的投影）
//	GET /v1/scheme/kinds                对象类型清单（list）
//	GET /v1/scheme/kinds/{kind}         单个对象类型详情（get）
//	GET /v1/objects/{kind}?scope=s      文档对象清单（list；scope 缺省=跨 scope 枚举）
//	GET /v1/objects/{kind}/{name}?scope=s 单个文档对象（get；对象键=(kind,scope,name)）
//	GET /v1/drafts                      草稿清单（瘦身；查询面恒开）
//	GET /v1/drafts/{kind}/{name}?scope=s 单条草稿（含正文；无稿 → 404）
//	POST /v1/drafts                     落稿（upsert；写面开关门控）
//	DELETE /v1/drafts/{kind}/{name}?scope=s 弃稿（写面开关门控）
//	GET /v1/triggers                    触发骨架查询面（dispatch 接线：表驱动订阅+留痕；恒开，§2.3）
//	GET /v1/triggers/resolve            派发决策查询面（v0.13 决策投影+v0.14 执行绑定投影；恒开，§2.3）
//	GET /v1/workflows/runs              对账 run 投影面（v0.15 真实拉起第一子段：workflow 触发面接线；恒开，§2.3）
//	POST /v1/apply                      文档写入口（--allow-apply 显式开启；未开启=405；fromDraft=true 走草稿）
//
// objects 为 nil（cored 未配置 --docs）时 /v1/objects/* 返回 503；
// applier 为 nil（未开启 --allow-apply）时 POST /v1/apply 与草稿写操作返回 405（缺省只读面不变；
// 草稿 GET 属查询面恒开）。drafts 为 nil（缓存未接线）时 /v1/drafts* 返回 503。
// triggers 为 nil（触发骨架未接线）时 /v1/triggers 与 /v1/triggers/resolve 返回 503。
// runs 为 nil（run 投影面未接线）时 /v1/workflows/runs 返回 503。
func New(sch *scheme.Scheme, objects ObjectSource, applier Applier, drafts DraftStaging, triggers TriggerFace, runs RunFace) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	mux.HandleFunc("GET /v1/scheme", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, sch.Summary())
	})
	mux.HandleFunc("GET /v1/scheme/kinds", func(w http.ResponseWriter, _ *http.Request) {
		kinds := sch.ListKinds()
		writeJSON(w, http.StatusOK, map[string]any{"count": len(kinds), "kinds": kinds})
	})
	mux.HandleFunc("GET /v1/scheme/kinds/{kind}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("kind")
		k, err := sch.Lookup(name)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, k)
	})
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
		objs, err := objects.List(r.Context(), kind, scope)
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
				writeApplyErr(w, err)
				return
			}
		} else {
			o := &repository.Object{
				Kind: req.Kind, Name: req.Name, Scope: req.Scope,
				ResourceVersion: req.ResourceVersion, Content: []byte(req.Content),
			}
			var err error
			res, err = applier.Apply(r.Context(), o, opt)
			if err != nil {
				writeApplyErr(w, err)
				return
			}
		}
		resp := map[string]any{
			"resource": fmt.Sprintf("%s/%s/%s", req.Kind, req.Scope, req.Name),
			"rv":       res.RV, "action": res.Action, "dryRun": req.DryRun,
		}
		if res.Event != nil {
			resp["event"] = map[string]any{"id": res.Event.ID, "type": res.Event.Type, "dedupKey": res.Event.DedupKey}
		}
		writeJSON(w, http.StatusOK, resp)
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
		writeJSON(w, http.StatusOK, map[string]any{
			"resource": fmt.Sprintf("%s/%s/%s", kind, scope, name), "action": "discarded",
		})
	})
	// 触发骨架面（cored-bridge.md §2.3；dispatch 接线：第二段 v0.12 订阅面动态化 →
	// 第三段 v0.13 引擎主体第一段 owner 解析激活）：GET 恒开——
	// 「apply 落面→事件→订阅收到留痕」主干第一段的机械可验查询面；
	// types=dispatch 表推导订阅集合（"*"=通配），records 携命中行投影 route/cc/for/owner
	// （全表未命中=dead-letter，死信兜底可见）。留痕是进程内观测面（账本=event_log 归档）；
	// 写面关断不影响本端点（received=0 照常 200）。
	mux.HandleFunc("GET /v1/triggers", func(w http.ResponseWriter, _ *http.Request) {
		if triggers == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "触发骨架未接线"})
			return
		}
		st := triggers.State()
		records := make([]map[string]any, 0, len(st.Records))
		for _, r := range st.Records {
			records = append(records, map[string]any{
				"seq": r.Seq, "eventId": r.EventID, "type": r.Type,
				"dedupKey": r.DedupKey, "source": r.Source, "involved": r.Involved,
				"level": r.Level, "route": r.Route, "cc": r.CC, "for": r.For, "owner": r.Owner,
				"execution":  map[string]any{"route": r.Exec, "cc": r.CCExec},
				"at":         r.At.UTC().Format(time.RFC3339Nano),
				"receivedAt": r.ReceivedAt.UTC().Format(time.RFC3339Nano),
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"count": len(records), "received": st.Received, "dropped": st.Dropped,
			"types": st.Types, "records": records,
		})
	})
	// 派发决策查询面（§2.3 v0.13 引擎主体第一段 → v0.14 第二段增执行投影）：对当前 dispatch
	// 表快照回答「这一事件若到达会路由到哪、由谁执行」——纯决策（零事件注入、零留痕、零派发），
	// 投影 {type, involved, matched, route, cc, for, owner, row, execution}；row=命中表行号（0 起），
	// 全表未命中 row=-1 且 route=dead-letter（合成死信，execution.route.executor=none）。
	// owner:<字面量> 谓词的 live 机械探针落点（无事件注入面也能验证 owner 解析命中）；
	// execution={route:{executor,target,wired}, cc:{…}}=route/cc 目标执行面绑定投影
	// （映射表家族收编，未知目标=unbound 缺口可见；wired=执行面接线态，装配层登记前恒 false）。
	mux.HandleFunc("GET /v1/triggers/resolve", func(w http.ResponseWriter, r *http.Request) {
		if triggers == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "触发骨架未接线"})
			return
		}
		typ := r.URL.Query().Get("type")
		if typ == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "缺 type 查询参数（决策必须有事件类型）"})
			return
		}
		involved := r.URL.Query().Get("involved") // 可缺省（解析为空 owner，决策照常给出）
		d := triggers.Resolve(typ, involved)
		writeJSON(w, http.StatusOK, map[string]any{
			"type": typ, "involved": involved, "matched": d.Matched,
			"route": d.Route, "cc": d.CC, "for": d.For, "owner": d.Owner, "row": d.Row,
			"execution": map[string]any{"route": d.Exec, "cc": d.CCExec},
		})
	})
	// 对账 run 投影面（§2.3 v0.15 引擎主体第三段——真实拉起第一子段：workflow 触发面接线）：
	// route 目标执行面=workflow 且装配层已接线的留痕推进为一次 run 投影（触发已受理+声明
	// 起点；相位态机 deferred——run 出现 ≠ 对账执行完毕）。旧构建（face 缺席）经兜底
	// handler 返回 405=代际判别器（与 loop-13 固化口径一致）。
	mux.HandleFunc("GET /v1/workflows/runs", func(w http.ResponseWriter, _ *http.Request) {
		if runs == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "run 投影面未接线"})
			return
		}
		st := runs.State()
		items := make([]map[string]any, 0, len(st.Runs))
		for _, r := range st.Runs {
			items = append(items, map[string]any{
				"run": r.Run, "workflow": r.Workflow, "specSource": r.SpecSource,
				"phases": r.Phases, "phase": r.Phase, "status": r.Status,
				"triggerSeq": r.TriggerSeq, "eventId": r.EventID, "type": r.Type,
				"dedupKey": r.DedupKey, "route": r.Route, "owner": r.Owner,
				"target": r.Target, "startedAt": r.StartedAt.UTC().Format(time.RFC3339Nano),
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"count": len(items), "started": st.Started, "dropped": st.Dropped, "runs": items,
		})
	})
	// 其余路径与方法一律 404/405，保持查询面最小。
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v1/") && r.URL.Path != "/healthz" {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": fmt.Sprintf("method %s not allowed", r.Method)})
	})
	return mux
}

func objRef(o *repository.Object) map[string]any {
	return map[string]any{"kind": o.Kind, "name": o.Name, "scope": o.Scope, "resourceVersion": o.ResourceVersion}
}

// draftRef 草稿瘦身引用（清单载荷：无正文；bytes 供容量观察）。
func draftRef(d *draft.Draft) map[string]any {
	return map[string]any{
		"kind": d.Kind, "name": d.Name, "scope": d.Scope,
		"fieldManager":        d.FieldManager,
		"baseResourceVersion": d.BaseRV,
		"updatedAt":           d.UpdatedAt.UTC().Format(time.RFC3339Nano),
		"bytes":               len(d.Content),
	}
}

// writeObjErr 仓储哨兵错误 → HTTP 状态码映射。
func writeObjErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrNotRegistered), errors.Is(err, repository.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
	case errors.Is(err, repository.ErrScope):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
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

// resolveSubject 请求级主体解析：请求体 subject 字段优先，缺席回退 X-Core-Subject 头；
// 两者皆缺 → 匿名主体（非 nil 指针=强制请求级判定，进程级本地信任不兜底 HTTP 匿名请求）。
func resolveSubject(r *http.Request, req *applyRequest) *permission.Subject {
	if req.Subject != nil {
		return &permission.Subject{User: req.Subject.User, Roles: req.Subject.Roles, Groups: req.Subject.Groups}
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
		errors.Is(err, repository.ErrNoFieldManager):
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

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(payload)
}
