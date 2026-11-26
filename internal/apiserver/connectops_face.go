// connectops_face.go —— POST /v1/connectops/{name}/{op} ops 调试面。
//
// 分发通路现状（R-ENG-3 / R-TYPE-3）：正式通路=workflow Verify 接线（dev-delivery Verifying
// 相位硬 gate 经 ConnectorOps 调 connectops-code-audit 的 audit/findings），待 workflow 侧
// 接入；本路由=通路建成前的最小调试暴露面（JSON body，subject 记录审计行）。只读语义由
// 实现侧契约保证（connectops-code-audit 只读不写；其他实现接入时随其自身契约），路由层
// 不做即兴写动作。
//
// 权限门（G-03 fail-closed，权限门契约 subject 三段必须 check）：subject 解析后先过权限
// 判定再 Invoke——资源域=ops/<name> × 动作=call（permission.md §1 资源域词形 {scope:
// "ops/onedev-ci", actions: [call]}；Role 对象按 ops/<name> 授权 call）。装配未注入权限
// 系统=零值系统（空授权面——仅 DefaultAdmin 字面量兜底；进程级本地信任不兜底 HTTP 匿名
// 请求，与 apply 写面同口径）；无效/无权 subject → 403，Invoke 不达（先拒后调）。
//
// 错误码：connector 未注册=404；op 不在声明集=400；params 不可解析=400；subject 无效/
// 无权=403（fail-closed）；op 执行失败=502（ErrNoAudit=404——回查未命中与执行失败分立）；
// 成功=200 透传实现侧 JSON 结果。
package apiserver

import (
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"

	"github.com/friddle-self-recursive-agent/core_system_workservice/connectors"
	"github.com/friddle-self-recursive-agent/core_system_workservice/connectors/codeaudit"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
)

// connectopsBodyMax 调试面请求体上限（1MiB——audit params 体量：target+checks/ruleChecks 投影）。
const connectopsBodyMax = 1 << 20

// connectops 资源域词形（permission.md §1 资源域：{scope: "ops/<name>", actions: [call]}
// ——Role.Holds 资源域形态精确比对，与 events publish 域分立同范式：ops call 权限不产生
// 对象 write，对象 write 不产生 ops call）。
const (
	connectOpsDomainPrefix = "ops/"
	connectOpsActionCall   = "call"
)

// WithPerm 注入权限系统（G-03 connectops 权限门判定权威；装配层与 eventpub.Publisher/apply
// 写面共用同一 *permission.System——同一代权威快照，判定依据可追。nil=零值系统 fail-closed：
// 仅 subject 断言 role/global/admin 字面量可过，Role 对象授权面缺席）。
func WithPerm(s *permission.System) Option {
	return func(o *options) { o.perm = s }
}

// codeAuditFindingsLister 只读 findings 投影口面（*codeaudit.Connector 满足）。
type codeAuditFindingsLister interface {
	ListRecent() []codeaudit.Result
}

// codeAuditSnapshotPayload 代码审核只读投影载荷（V-18：GET /v1/code-audit/findings 与
// /v1/snapshot 共用零双源；驱动未注册=nil 诚实缺席，不冒充空）。
func codeAuditSnapshotPayload() map[string]any {
	c, ok := connectors.LookupOpsConnector(codeaudit.TypeName)
	if !ok {
		return nil
	}
	lister, ok := c.(codeAuditFindingsLister)
	if !ok {
		return nil
	}
	audits := lister.ListRecent()
	if audits == nil {
		audits = []codeaudit.Result{}
	}
	return map[string]any{
		"count":  len(audits),
		"audits": audits,
		"source": codeaudit.TypeName,
	}
}

// registerConnectOpsRoutes ops 调试面与代码审核只读投影路由注册（server.go New 尾段装配；
// POST 调试面 subject 权限门 fail-closed；GET /v1/code-audit/findings 为只读投影零写面）。
func registerConnectOpsRoutes(mux *http.ServeMux, perm *permission.System) {
	mux.HandleFunc("GET /v1/code-audit/findings", func(w http.ResponseWriter, _ *http.Request) {
		p := codeAuditSnapshotPayload()
		if p == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "connectops-code-audit 只读投影未装配（驱动未注册）"})
			return
		}
		writeJSON(w, http.StatusOK, p)
	})
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
		// 权限门（G-03 fail-closed——先拒后调）：subject 三段必须 check 才可 Invoke。判定方
		// ByRole=请求级主体 role 表查表（permission.md §4 b 路，events 发布面同款）；域=ops/<name>
		// 资源域精确比对 × 动作=call。装配未注入权限系统=零值系统（空授权面——匿名/无权一律
		// 403，DefaultAdmin 字面量是最低保障；宁拒不降级）。
		sys := perm
		if sys == nil {
			sys = &permission.System{}
		}
		if err := sys.Check(*subject, connectOpsDomainPrefix+name, connectOpsActionCall, permission.ByRole); err != nil {
			writeJSON(w, http.StatusForbidden, map[string]any{"error": "connectops 权限拒绝（fail-closed）: " + err.Error()})
			return
		}
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
