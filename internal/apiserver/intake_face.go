// intake_face.go —— R-UX-1 对话式录入与三重 dry-run 预检路由（v0.53 能力 D——cored-bridge.md §2.5）：
//   - GET /v1/intake/sessions      会话留痕清单（恒开，不含对话原文；旧构建 405=代际判别器）
//   - GET /v1/intake/sessions/{id} 单条会话留痕与字段快照（未命中 404）
//   - POST /v1/intake/turn         推进一轮对话/字段卡双向同步（写面门控 --allow-apply，未开启 405）
//   - POST /v1/intake/generate     提交节点生成标准 YAML（rv=0）+ 系统补缺省 + 三重 dry-run（R-PUB lint/spec 准入/权限预览，失败定位到具体字段）
package apiserver

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/intake"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

type intakeTurnBody struct {
	SessionID   string          `json:"sessionId"`
	Kind        string          `json:"kind"`
	Scope       string          `json:"scope"`
	Name        string          `json:"name"`
	Message     string          `json:"message"`
	SkipCurrent bool            `json:"skipCurrent"`
	Fields      map[string]any  `json:"fields"`
	Subject     *subjectRequest `json:"subject"`
}

type intakeGenerateBody struct {
	SessionID string          `json:"sessionId"`
	Kind      string          `json:"kind"`
	Scope     string          `json:"scope"`
	Name      string          `json:"name"`
	Fields    map[string]any  `json:"fields"`
	Subject   *subjectRequest `json:"subject"`
}

func registerIntakeRoutes(mux *http.ServeMux, sch *scheme.Scheme, objects ObjectSource, applier Applier) {
	eng := intake.NewEngine(sch, objects, applier, nil)

	mux.HandleFunc("GET /v1/intake/sessions", func(w http.ResponseWriter, _ *http.Request) {
		rows := eng.ListSessions()
		writeJSON(w, http.StatusOK, map[string]any{"count": len(rows), "sessions": rows})
	})

	mux.HandleFunc("GET /v1/intake/sessions/{id}", func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		rec, err := eng.GetSession(id)
		if err != nil {
			if errors.Is(err, intake.ErrSessionNotFound) {
				writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, rec)
	})

	mux.HandleFunc("POST /v1/intake/turn", func(w http.ResponseWriter, r *http.Request) {
		if applier == nil {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "对话录入写面未开启（cored 缺 --allow-apply）"})
			return
		}
		var body intakeTurnBody
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("请求体不可解析: %v", err)})
			return
		}
		res, err := eng.Turn(r.Context(), intake.TurnRequest{
			SessionID:   body.SessionID,
			Kind:        body.Kind,
			Scope:       body.Scope,
			Name:        body.Name,
			Message:     body.Message,
			SkipCurrent: body.SkipCurrent,
			Fields:      body.Fields,
			Subject:     subjectFrom(r, body.Subject),
		})
		if err != nil {
			if errors.Is(err, intake.ErrNotRegistered) {
				writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, res)
	})

	mux.HandleFunc("POST /v1/intake/generate", func(w http.ResponseWriter, r *http.Request) {
		if applier == nil {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "对话录入生成面未开启（cored 缺 --allow-apply）"})
			return
		}
		var body intakeGenerateBody
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 1<<20)).Decode(&body); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("请求体不可解析: %v", err)})
			return
		}
		res, err := eng.Generate(r.Context(), intake.GenerateRequest{
			SessionID: body.SessionID,
			Kind:      body.Kind,
			Scope:     body.Scope,
			Name:      body.Name,
			Fields:    body.Fields,
			Subject:   subjectFrom(r, body.Subject),
		})
		if err != nil {
			if errors.Is(err, intake.ErrNotRegistered) {
				writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, res)
	})
}
