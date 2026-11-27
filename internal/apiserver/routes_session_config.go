// routes_session_config.go —— 三段式 Session 控制面路由（session.md；2026-11-27 用户裁决
// 「有没有 scheme 来承载我这个三段式的描述。比如我想控制一些描述？页面上怎么承载」的承载面）。
// GET /v1/session-config          恒开——存储缺席=scheme 缺省（DefaultSessionConfig）不冒充持久化；
// PUT /v1/session-config          随 --allow-apply 门控（与 docs-export 同开关）——三段描述/背景
//                                 预算/逐块溯源/属性回整开关与域/Provider/模型 全量覆盖；
// GET /v1/session-config/writebacks  第三段回整账最近 N 条（只读恒开——「自动回整」来源标记数据源）。
package apiserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
)

func registerSessionConfigRoutes(mux *http.ServeMux, objects ObjectSource, applier Applier) {
	mux.HandleFunc("GET /v1/session-config", func(w http.ResponseWriter, r *http.Request) {
		store, ok := objects.(repository.SessionConfigStore)
		if !ok {
			writeJSON(w, http.StatusOK, map[string]any{
				"config": repository.DefaultSessionConfig(),
				"persistent": false,
				"note": "当前存储后端不支持持久化三段控制面（需 sqlite docs DB）——返回 scheme 缺省，不冒充已保存",
			})
			return
		}
		cfg, err := store.GetSessionConfig(r.Context())
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"config": cfg, "persistent": false,
				"note": "存储读失败返回缺省（" + err.Error() + "）",
			})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"config": cfg, "persistent": true})
	})
	mux.HandleFunc("PUT /v1/session-config", func(w http.ResponseWriter, r *http.Request) {
		if applier == nil {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "三段控制面写路径未开启（cored 缺 --allow-apply；缺省只读）"})
			return
		}
		store, ok := objects.(repository.SessionConfigStore)
		if !ok {
			writeJSON(w, http.StatusNotImplemented, map[string]any{"error": "当前存储后端不支持持久化三段控制面（需 sqlite docs DB）"})
			return
		}
		var req repository.SessionConfig
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("请求体不可解析: %v", err)})
			return
		}
		saved, err := store.SetSessionConfig(r.Context(), req)
		if err != nil {
			writeObjErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, saved)
	})
	mux.HandleFunc("GET /v1/session-config/writebacks", func(w http.ResponseWriter, r *http.Request) {
		n, _ := strconv.Atoi(r.URL.Query().Get("n"))
		store, ok := objects.(repository.SessionConfigStore)
		if !ok {
			writeJSON(w, http.StatusOK, map[string]any{"items": []repository.SessionWriteback{}, "persistent": false})
			return
		}
		items, err := store.LastSessionWritebacks(r.Context(), n)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]any{"items": []repository.SessionWriteback{}, "persistent": false, "note": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items, "persistent": true})
	})
}
