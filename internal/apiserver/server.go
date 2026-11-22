// Package apiserver cored 最小只读查询面：scheme 加载校验结果 + 对象类型清单 get/list。
// 设计依据：core_system_instance/runtime/cored-bridge.md（plugin↔cored 桥契约）；
// 只读——写路径唯一入口是 Apply 管线（core_design/concepts/apply.md），本面不提供写端点。
package apiserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// New 构建只读 HTTP 处理器。端点：
//
//	GET /healthz                存活探针
//	GET /v1/scheme              scheme 概览（加载校验结果的投影）
//	GET /v1/scheme/kinds        对象类型清单（list）
//	GET /v1/scheme/kinds/{kind} 单个对象类型详情（get）
func New(sch *scheme.Scheme) http.Handler {
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

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(payload)
}
