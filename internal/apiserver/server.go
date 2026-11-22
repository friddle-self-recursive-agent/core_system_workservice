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

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/apply"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// ObjectSource 文档对象读取源（Repository 的读半边；接口化便于测试注入假仓库）。
type ObjectSource interface {
	Get(ctx context.Context, kind, scope, name string) (*repository.Object, error)
	List(ctx context.Context, kind, scope string) ([]*repository.Object, error)
}

// Applier 写路径（apply.Engine 的口面；接口化便于测试注入假引擎）。
type Applier interface {
	Apply(ctx context.Context, o *repository.Object, opt apply.Options) (*apply.Result, error)
}

// compile 期校验：git 后端满足读半边，apply 引擎满足写口面。
var (
	_ ObjectSource = (*repository.GitRepo)(nil)
	_ Applier      = (*apply.Engine)(nil)
)

// New 构建查询面 HTTP 处理器。端点：
//
//	GET /healthz                        存活探针
//	GET /v1/scheme                      scheme 概览（加载校验结果的投影）
//	GET /v1/scheme/kinds                对象类型清单（list）
//	GET /v1/scheme/kinds/{kind}         单个对象类型详情（get）
//	GET /v1/objects/{kind}?scope=s      文档对象清单（list；scope 缺省=跨 scope 枚举）
//	GET /v1/objects/{kind}/{name}?scope=s 单个文档对象（get；对象键=(kind,scope,name)）
//	POST /v1/apply                      文档写入口（--allow-apply 显式开启；未开启=405）
//
// objects 为 nil（cored 未配置 --docs）时 /v1/objects/* 返回 503；
// applier 为 nil（未开启 --allow-apply）时 POST /v1/apply 返回 405（缺省只读面保持不变）。
func New(sch *scheme.Scheme, objects ObjectSource, applier Applier) http.Handler {
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
		o := &repository.Object{
			Kind: req.Kind, Name: req.Name, Scope: req.Scope,
			ResourceVersion: req.ResourceVersion, Content: []byte(req.Content),
		}
		res, err := applier.Apply(r.Context(), o, apply.Options{
			DryRun: req.DryRun, Force: req.Force,
			FieldManager: req.FieldManager, ReasonRef: req.ReasonRef,
		})
		if err != nil {
			writeApplyErr(w, err)
			return
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
	Kind            string `json:"kind"`
	Scope           string `json:"scope"`
	Name            string `json:"name"`
	ResourceVersion int64  `json:"resourceVersion"`
	Content         string `json:"content"`
	FieldManager    string `json:"fieldManager"`
	ReasonRef       string `json:"reasonRef"`
	DryRun          bool   `json:"dryRun"`
	Force           bool   `json:"force"`
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
		errors.Is(err, repository.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
	case errors.Is(err, apply.ErrDenied):
		writeJSON(w, http.StatusForbidden, map[string]any{"error": err.Error()})
	case errors.Is(err, apply.ErrNoFieldManager), errors.Is(err, apply.ErrObject),
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

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(payload)
}
