// connectops_face.go —— POST /v1/connectops/{name}/{op} ops 调试面。
//
// 分发通路现状（R-ENG-3 / R-TYPE-3）：正式通路=workflow Verify 接线（dev-delivery Verifying
// 相位硬 gate 经 ConnectorOps 调 connectops-code-audit 的 audit/findings），待 workflow 侧
// 接入；本路由=通路建成前的最小调试暴露面（JSON body，subject 记录审计行）。只读语义由
// 实现侧契约保证（connectops-code-audit 只读不写；其他实现接入时随其自身契约），路由层
// 不做即兴写动作。
//
// 错误码：connector 未注册=404；op 不在声明集=400；params 不可解析=400；op 执行失败=502
// （ErrNoAudit=404——回查未命中与执行失败分立）；成功=200 透传实现侧 JSON 结果。
package apiserver

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/friddle-self-recursive-agent/core_system_workservice/connectors"
	"github.com/friddle-self-recursive-agent/core_system_workservice/connectors/codeaudit"
)

// connectopsBodyMax 调试面请求体上限（1MiB——audit params 体量：target+checks/ruleChecks 投影）。
const connectopsBodyMax = 1 << 20

// registerConnectOpsRoutes ops 调试面路由注册（server.go New 尾段装配；面恒开——调用即
// 只读审核，非写面，不随 --allow-apply 门控）。
func registerConnectOpsRoutes(mux *http.ServeMux) {
	mux.HandleFunc("POST /v1/connectops/{name}/{op}", func(w http.ResponseWriter, r *http.Request) {
		name, op := r.PathValue("name"), r.PathValue("op")
		c, ok := connectors.LookupOpsConnector(name)
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "connectops " + name + " 未注册（驱动包未导入/装配缺席）"})
			return
		}
		if !connectors.HasOp(c, op) {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "op " + op + " 不在 " + name + " 声明集 " + fmtOps(c.Ops())})
			return
		}
		body, err := io.ReadAll(io.LimitReader(r.Body, connectopsBodyMax+1))
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求体读取失败: " + err.Error()})
			return
		}
		if len(body) > connectopsBodyMax {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求体超上限（1MiB）"})
			return
		}
		// 信封拆封：subject 记录审计（v0.9 请求级主体同款投影——subject 字段优先，
		// 缺席回退 X-Core-Subject 头）；其余字段=op params 原样透传（契约归实现侧）。
		var envelope map[string]json.RawMessage
		if len(body) > 0 {
			if err := json.Unmarshal(body, &envelope); err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "请求体不可解析: " + err.Error()})
				return
			}
		}
		var subj *subjectRequest
		if raw, has := envelope["subject"]; has {
			var s subjectRequest
			if err := json.Unmarshal(raw, &s); err == nil {
				subj = &s
			}
			delete(envelope, "subject")
		}
		subject := subjectFrom(r, subj)
		params, err := json.Marshal(envelope)
		if err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "params 编码失败: " + err.Error()})
			return
		}
		result, err := c.Invoke(r.Context(), op, params)
		if err != nil {
			if errors.Is(err, codeaudit.ErrNoAudit) {
				writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
				return
			}
			writeJSON(w, http.StatusBadGateway, map[string]any{"error": "op 执行失败: " + err.Error()})
			return
		}
		// 审计留痕：subject+op 结构化行（调试面 stderr；正式审计=workflow Verify 接线走
		// apply_log/VerifyDoc/Event 链——rule-verification.md 生命周期第 4 步口径）。
		log.Printf("apiserver: connectops %s/%s subject.user=%q bytes=%d", name, op, subject.User, len(result))
		writeJSON(w, http.StatusOK, json.RawMessage(result))
	})
}

// fmtOps op 集展示串（错误提示用；字典序保持实现侧声明顺序）。
func fmtOps(ops []string) string {
	out := "["
	for i, o := range ops {
		if i > 0 {
			out += " "
		}
		out += o
	}
	return out + "]"
}
