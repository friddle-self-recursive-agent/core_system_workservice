// Package apiserver cored 最小只读查询面：scheme 加载校验结果 + 对象类型清单 get/list
// + 文档对象读面（/v1/objects，读库内真实文档对象而非类型注册表）。
// 设计依据：core_system_instance/runtime/cored-bridge.md（plugin↔cored 桥契约）；
// 只读——写路径唯一入口是 Apply 管线（core_design/concepts/apply.md），本面不提供写端点。
package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// ObjectSource 文档对象读取源（Repository 的读半边；接口化便于测试注入假仓库）。
type ObjectSource interface {
	Get(ctx context.Context, kind, scope, name string) (*repository.Object, error)
	List(ctx context.Context, kind, scope string) ([]*repository.Object, error)
}

// compile 期校验：git 后端满足读半边。
var _ ObjectSource = (*repository.GitRepo)(nil)

// New 构建只读 HTTP 处理器。端点：
//
//	GET /healthz                        存活探针
//	GET /v1/scheme                      scheme 概览（加载校验结果的投影）
//	GET /v1/scheme/kinds                对象类型清单（list）
//	GET /v1/scheme/kinds/{kind}         单个对象类型详情（get）
//	GET /v1/objects/{kind}?scope=s      文档对象清单（list；scope 缺省=跨 scope 枚举）
//	GET /v1/objects/{kind}/{name}?scope=s 单个文档对象（get；对象键=(kind,scope,name)）
//
// objects 为 nil（cored 未配置 --docs）时 /v1/objects/* 返回 503。
func New(sch *scheme.Scheme, objects ObjectSource) http.Handler {
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
	// 其余路径与方法一律 404/405，保持只读面最小。
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

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(payload)
}
