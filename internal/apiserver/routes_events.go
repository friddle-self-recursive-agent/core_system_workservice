// routes_events.go —— 事件域路由组（R-ENG-7 server.go 拆分：纯搬家零语义变化——
// 事件注入/事件归档查询/关注报告/收件箱/会话活性面与各自请求体、错误映射 helper
// 自 server.go 原位迁入，handler 体逐字节保持不动；装配与公共 helper 仍在 server.go）。
package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventpub"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/flowbind"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/proposal"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/sessionliveness"
)

// registerEventRoutes 事件域路由注册（events / attention / inbox / sessions——纯搬家）。
func registerEventRoutes(d *deps, mux *http.ServeMux) {
	sch := d.sch
	objects := d.objects
	applier := d.applier
	drafts := d.drafts
	runs := d.runs
	publisher := d.publisher
	o := d.o
	eventLog := o.eventLog
	agentScopes := o.agentScopes
	// 事件注入面（v0.20 发布面首段，concepts/event.md 规则 8；写面随 --allow-apply 门控，
	// 未开启恒 405 与 apply 写面同口径）：POST /v1/events——类型注册校验（Scheme.eventTypes，
	// 事件只校验类型注册）→ 主体发布权判定（publish × events 资源域，与 apply 写权限域分立）
	// → level 缺省=注册表 defaultLevel 投影 → 总线发布（与引擎内部事件同一条主干：窗口去重/
	// 订阅派发/event_log 归档零旁路）。注入通道 source 由服务侧判定恒 human，载荷声明的 source
	// 不采纳（来源标注是审计事实不是请求参数）；回执 {id,type,dedupKey,level,source,involved}
	// 与留痕 eventId 同源。旧构建（face 缺席）经兜底 handler 返回 405=代际判别器。
	mux.HandleFunc("POST /v1/events", func(w http.ResponseWriter, r *http.Request) {
		if publisher == nil {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "事件注入面未开启（cored 缺 --allow-apply；缺省只读）"})
			return
		}
		var req eventsRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("请求体不可解析: %v", err)})
			return
		}
		if req.Source != "" && req.Source != eventpub.SourceHuman {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("载荷声明的 source=%q 不采纳：来源标注由服务侧判定（注入通道恒 %s）", req.Source, eventpub.SourceHuman)})
			return
		}
		ev, err := publisher.Publish(eventpub.Request{
			Type: req.Type, Involved: req.Involved, Level: req.Level, DedupKey: req.DedupKey,
			Scope: req.Scope,
		}, subjectFrom(r, req.Subject))
		if err != nil {
			writePublishErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"id": ev.ID, "type": ev.Type, "dedupKey": ev.DedupKey,
			"level": ev.Level, "source": ev.Source, "involved": ev.Involved,
			"scope": ev.Scope, // metadata.scope 回执（BUG-010——显式声明或 involved 推导回填后的生效词）
		})
	})
	// 关注报告（R-UX-6 / concepts/attention.md；2026-11-26）：读侧派生，确定性候选+facts+
	// 证据链；v1 规则版（generatedBy=rules，LLM 叙述/缓存/权限指纹后续段）。数据源=objects+
	// runs+eventLog+inbox；face 缺席=degraded 报告（不 405，不空白）。只读 GET。
	mux.HandleFunc("GET /v1/attention", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, attentionReport(r.Context(), objects, runs, o, sch.HiddenEventTypes()))
	})
	// 通知中心收件箱投影面（v0.56 R-10——console-adoption.md §1.25「待你处理收件箱」页面兜底
	// 首段 + notification.md 规则 6「未配置通知通道=仅页面展示」的机读落点）。四类判定分类
	// （机械判定=读面现读自分类——console 不自建持久状态红线 2）：NeedsHuman（归档 NeedsHuman
	// ∧Escalate 行——acceptance.md §9 接管链）；ApprovalNeeded（归档行——审批提升人注意终态族）
	// ∧批准投影挂起行（v0.56 派单 ask 批准动作面——flowbind.SyncApprovals 同一投影）；Proposal
	// 待批（status=submitted 提案——capability B 状态机词汇）；DocReview（草稿缓存行——代改代审
	// 的家，apply.md §6）。载荷 {count, classes, items, watchdog}；行含 class/kind/name/scope/
	// digest/eventId/at/link（link=console 深链路由——规则 6 页面兜底入口恒达操作面）。
	// face 缺席=503（防御面）；旧构建 405=代际判别器。watchdog 投影={approvals, pending,
	// approveIdle}——真实词形（approveIdle=true=批准投影无挂起待决，非停摆）。
	mux.HandleFunc("GET /v1/inbox", func(w http.ResponseWriter, _ *http.Request) {
		if o.inbox == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "通知中心收件箱面未接线（装配缺 flowbind 批准投影 sink）"})
			return
		}
		writeJSON(w, http.StatusOK, inboxPayload(objects, drafts, o.inbox, o.eventLog))
	})
	// 会话活性面（KI-003，event.md 规则 14；agent-session.md §7）：
	// POST /v1/sessions/heartbeat=会话侧心跳注册（--allow-apply 门控——写面动作与注入面同开关）；
	// GET /v1/sessions=查询面恒开（判定态=state 词 suspected-stall 判定词 SuspectedStall 的 DB 词形）。
	mux.HandleFunc("POST /v1/sessions/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		if o.sessionLiveness == nil {
			writeJSON(w, http.StatusNotImplemented, map[string]any{"error": "sessionliveness: 会话活性面未装配（--session-watch-interval）"})
			return
		}
		var req sessionliveness.Heartbeat
		if err := json.NewDecoder(io.LimitReader(r.Body, 64<<10)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "sessionliveness: 心跳载荷解析失败: " + err.Error()})
			return
		}
		// 心跳防伪（v0.56——agent-session.md §7 心跳面唯一口径）：watchdog 自身动作的续租
		// （source=watchdog——注入「继续」后的活性观测心跳）不可经本面自报（来源伪冒=自续租
		// 死循环）；watchdog 侧经 BeatSource 落行，HTTP 面一律拒收（形态校验前置于写面门控）。
		// 判定/注入动作不续租语义（v0.54 既有：状态迁移不续租）随行不变。
		if req.Source == "watchdog" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "sessionliveness: 心跳来源=watchdog 不可经注册面上报（判定面自续租禁入——agent-session.md §7 心跳面唯一口径）"})
			return
		}
		if applier == nil {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "写面关断（--allow-apply 未开启）——心跳注册面随写面门控"})
			return
		}
		if err := o.sessionLiveness.Beat(req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "sessionId": req.SessionID, "at": time.Now().UTC().Format(time.RFC3339Nano)})
	})
	mux.HandleFunc("GET /v1/sessions", func(w http.ResponseWriter, r *http.Request) {
		if o.sessionLiveness == nil {
			writeJSON(w, http.StatusNotImplemented, map[string]any{"error": "sessionliveness: 会话活性面未装配（--session-watch-interval）"})
			return
		}
		limit := atoiDefault(r.URL.Query().Get("limit"), 200)
		rows, err := o.sessionLiveness.Heartbeats(limit)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"count": len(rows), "sessions": rows,
			"watchdog": map[string]any{
				"enabled":        o.sessionLiveness.Enabled(),
				"interval":       o.sessionLiveness.Interval().String(),
				"stallThreshold": o.sessionLiveness.Threshold().String(),
				"healRetryLimit": o.sessionLiveness.HealLimit(),
			},
		})
	})
	// 事件归档查询面（v0.36 console 读面首段，cored-bridge.md §2 `GET /v1/events`）：
	// event_log 归档行的最新优先查询投影（limit/type/level/involved/since/until 六参最小集；
	// limit≤0=缺省 50，>500 收敛 500——读面防拖库，全量走重建/导出面）。since/until=半开窗
	// [since, until)（until 严格不含端点——E7 向后分页游标：取上一页最旧行 ts 翻页，零重叠）。
	// 只读投影恒开（face 随 ops DB 装配恒在）；旧构建（face 缺席路由未注册）经兜底 handler
	// 返回 405=代际判别器。数据来源=归档（event.md：审计素材）， Welcome 最近重要事件与
	// 事件流页（E6）共用本面——渲染数据无双源红线（console-adoption.md §5 规则 2）。
	// visibility 过滤（BUG-010，event.md 规则 18）：默认视窗只出 tier≠infra（生效词形
	// default-hidden）的类型——type→tier/visibility 映射以 scheme 注册表为唯一权威逐请求
	// 现取（热刷新跟随；未注册类型不在隐藏集=按 domain/default-visible 处理不误伤）；
	// `?showInfra=true` 显式展开（调试视图勾选）。`?agent=<锚>` 消费闸（BUG-010 规则 19③）：
	// 逐行过 flowbind.AgentScopeGate 同一判定（与派发闸共用 Resolver）——Agent 会话取事件
	// 越界行服务侧过滤（不可见即不可消费），行级拦截不逐条告警（请求级注记一次）。
	mux.HandleFunc("GET /v1/events", func(w http.ResponseWriter, r *http.Request) {
		if eventLog == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "事件归档查询面未接线（装配缺 ops DB 口面）"})
			return
		}
		hidden := sch.HiddenEventTypes() // 注册表现取（type→tier/visibility 唯一权威；未注册不误伤）
		q := opslog.EventQuery{
			Type:        r.URL.Query().Get("type"),
			Level:       r.URL.Query().Get("level"),
			Involved:    r.URL.Query().Get("involved"),
			Limit:       atoiDefault(r.URL.Query().Get("limit"), 0),
			HiddenTypes: hidden,
			ShowInfra:   r.URL.Query().Get("showInfra") == "true",
		}
		if raw := r.URL.Query().Get("since"); raw != "" {
			t, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("since 不可解析（需 RFC3339）: %v", err)})
				return
			}
			q.Since = t
		}
		if raw := r.URL.Query().Get("until"); raw != "" {
			t, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("until 不可解析（需 RFC3339）: %v", err)})
				return
			}
			q.Until = t
		}
		// 消费闸（BUG-010 规则 19③）：agent 参数给出=Agent 取事件视窗，行级过闸过滤。
		var consumer string
		if raw := r.URL.Query().Get("agent"); raw != "" {
			if agentScopes == nil {
				writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "消费闸未接线（装配缺 Agent scope 解析面——cored 缺文档库）"})
				return
			}
			consumer = raw
		}
		rows, err := eventLog.QueryEvents(q)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		gate := &flowbind.AgentScopeGate{Resolver: agentScopes}
		blocked := 0
		items := make([]map[string]any, 0, len(rows))
		for _, e := range rows {
			if consumer != "" {
				ok, blk := gate.JudgeConsume(e.Scope, consumer)
				if !ok {
					blocked++ // 越界行不投毒给消费面：服务侧过滤（不可见即不可消费——规则 19③）
					log.Printf("apiserver: %s\n", blk.BlockedNote())
					continue
				}
			}
			tier, visibility := eventTierOf(sch, e.Type)
			items = append(items, map[string]any{
				"ts": e.Ts.UTC().Format(time.RFC3339Nano), "eventId": e.EventID,
				"type": e.Type, "source": e.Source, "involved": e.Involved,
				"dedupKey": e.DedupKey, "level": e.Severity, "routedTo": e.RoutedTo,
				"scope": e.Scope,                        // metadata.scope 投影（BUG-010——归档行强制携带，旧行读侧推导兜底）
				"tier":  tier, "visibility": visibility, // 分层/展示词形随行（规则 18——机读判据）
			})
		}
		resp := map[string]any{"count": len(items), "events": items}
		if consumer != "" {
			resp["agent"] = consumer
			resp["scopeBlocked"] = blocked // 请求级拦截注记（可见不静默；行级不逐条刷屏）
		}
		writeJSON(w, http.StatusOK, resp)
	})
	// 单事件详情读面（v0.66 R-ENG §3——`GET /v1/events/{eventId}`）：归档行直投
	// （EventQuery.EventID 精确直取，404=不在归档——不冒充空对象）+ 关联链：① 同 eventId
	// 的 run（runs face 环内按 EventID 过滤——face 缺席=null）；② 同 dedupKey 后续事件
	// （EventQuery.DedupKey 查询，剔除本行，limit 20 有界——dedupKey 空=无聚合键无后续）；
	// ③ involved 对象键解析（kind/scope/name 三段——畸形/非对象引用=null 不编造）。
	// 详情面 ShowInfra=true（显式单事件查阅——visibility 默认窗不适用于直取详情，与
	// `?showInfra=true` 调试视图同口径）。face 缺席=503（防御面同 GET /v1/events）。
	mux.HandleFunc("GET /v1/events/{eventId}", func(w http.ResponseWriter, r *http.Request) {
		if eventLog == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "事件归档查询面未接线（装配缺 ops DB 口面）"})
			return
		}
		id := r.PathValue("eventId")
		if id == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "eventId 必填"})
			return
		}
		rows, err := eventLog.QueryEvents(opslog.EventQuery{EventID: id, Limit: 1, ShowInfra: true})
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		if len(rows) == 0 {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "事件不在归档（归档行不删不改—— eventId 与发布回执同源）", "eventId": id})
			return
		}
		e := rows[0]
		tier, visibility := eventTierOf(sch, e.Type)
		out := map[string]any{
			"event": map[string]any{
				"ts": e.Ts.UTC().Format(time.RFC3339Nano), "eventId": e.EventID,
				"type": e.Type, "source": e.Source, "involved": e.Involved,
				"dedupKey": e.DedupKey, "level": e.Severity, "routedTo": e.RoutedTo,
				"scope": e.Scope, "tier": tier, "visibility": visibility,
			},
			"involved": involvedRefOf(e.Involved),
		}
		// 关联链①：同 eventId 的 run（runs face 环内过滤；face 缺席=null）。
		out["runs"] = faceOrNil(runs != nil, func() any {
			matched := []map[string]any{}
			for _, rr := range runs.State().Runs {
				if rr.EventID == e.EventID {
					matched = append(matched, map[string]any{
						"run": rr.Run, "workflow": rr.Workflow, "phase": rr.Phase, "status": rr.Status,
						"startedAt": rr.StartedAt.UTC().Format(time.RFC3339Nano),
					})
				}
			}
			return matched
		})
		// 关联链②：同 dedupKey 后续事件（本行剔除；空 dedupKey=无聚合键=null）。
		out["followUps"] = faceOrNil(e.DedupKey != "", func() any {
			dups, err := eventLog.QueryEvents(opslog.EventQuery{DedupKey: e.DedupKey, Limit: 20, ShowInfra: true})
			if err != nil {
				return nil // 读失败=null 不冒充空链（诚实投影）
			}
			follows := []map[string]any{}
			for _, d := range dups {
				if d.EventID == e.EventID {
					continue
				}
				follows = append(follows, map[string]any{
					"ts": d.Ts.UTC().Format(time.RFC3339Nano), "eventId": d.EventID,
					"type": d.Type, "level": d.Severity, "dedupKey": d.DedupKey,
				})
			}
			return follows
		})
		writeJSON(w, http.StatusOK, out)
	})
}

// inboxPayload GET /v1/inbox 载荷构造器（原 handler 内联体原样上提——snapshot 面与路由
// 共用同一构造器零双源，R-ENG-7 拆分注记；face 缺席由调用方先行判 nil）。
func inboxPayload(objects ObjectSource, drafts DraftStaging, face InboxFace, eventLog EventLogFace) map[string]any {
	var evs []opslog.EventLog
	if eventLog != nil {
		// 三类人注意事件各拉一窗（type 精确过滤——归档窗 DocApplied 刷屏不挤占人注意行；
		// 半开窗语义=QueryEvents 既有 limit 口径）。
		for _, typ := range []string{"NeedsHuman", "Escalate", "ApprovalNeeded"} {
			if rows, err := eventLog.QueryEvents(opslog.EventQuery{Type: typ, Limit: 100}); err == nil {
				evs = append(evs, rows...)
			}
		}
	}
	var draftRows []flowbind.InboxDraftRow
	if drafts != nil {
		for _, d := range drafts.List() {
			draftRows = append(draftRows, flowbind.InboxDraftRow{Kind: d.Kind, Scope: d.Scope,
				Name: d.Name, FieldManager: d.FieldManager, UpdatedAt: d.UpdatedAt.UTC().Format(time.RFC3339Nano)})
		}
	}
	var propRows []flowbind.InboxProposalRow
	if objects != nil {
		if props, err := objects.List(context.Background(), proposalKind, ""); err == nil {
			for _, po := range props {
				if spec, ok := proposal.Parse(po.Content); ok {
					propRows = append(propRows, flowbind.InboxProposalRow{Name: po.Name, Scope: po.Scope,
						Title: spec.Title, Status: spec.Status, UpdatedAt: ""})
				}
			}
		}
	}
	st := face.State()
	pending := 0
	var pendRows []flowbind.InboxRow
	for _, r := range st.Records {
		if r.State != flowbind.ApprovalPending {
			continue // v0.58：终态行（已判定）零挂起行——回链闭环拍即生效（动作面 console 动作化）
		}
		pending++
		// v0.57 挂起行进行集（console-adoption.md §1.32 L59-UI-02——幽灵计数缺口收口）：
		// ask 批准投影挂起行=ApprovalNeeded 类条目行（行集与计数同源不翻倍——加法投影
		// 摘除，计数=行数）。落点=派单投影所在 wf 详情配置 tab（#/wf/<锚>——批准动作=
		// Apply workflow:approve 判定链路）。
		name := r.Connector
		anchor := r.Workflow
		if r.Kind == flowbind.DispatchKindProposal || r.Proposal != "" {
			name = r.Proposal
			anchor = r.Proposal
		}
		pendRows = append(pendRows, flowbind.InboxRow{
			Class: flowbind.InboxClassApprovalNeeded, Kind: "dispatch", Name: name, Scope: r.Scope,
			Digest:  "派单待批（" + anchor + " · autoLevel=" + r.AutoLevel + "）",
			EventID: r.EventID, At: r.At.UTC().Format(time.RFC3339Nano),
			Link: "#/wf/" + anchor,
		})
	}
	box := flowbind.BuildInbox(evs, draftRows, propRows, nil)
	if len(pendRows) > 0 {
		box.Items = append(box.Items, pendRows...) // 人注意优先类序由 BuildInbox 排序——挂起行同类合并后重排
		flowbind.SortInboxRows(box.Items)
		box.Count = len(box.Items)
		if box.Classes == nil {
			box.Classes = map[string]int{}
		}
		box.Classes[flowbind.InboxClassApprovalNeeded] += len(pendRows)
	}
	return map[string]any{
		"count": box.Count, "classes": box.Classes, "items": box.Items,
		"approvals": map[string]any{"count": len(st.Records), "pending": pending,
			"records": approvalsPayload(st.Records)},
		"watchdog": map[string]any{"approvals": len(st.Records), "pending": pending,
			"approveIdle": face.ApproveIdle()},
	}
}

// involvedRefOf involved 投影 → 对象键解析（kind/scope/name 三段——单事件详情关联链③；
// 畸形/非对象引用（段数不足或键空）=nil 不编造——与 run.go workflowRefOf 同构保守口径）。
func involvedRefOf(involved string) map[string]any {
	segs := strings.Split(involved, "/")
	if len(segs) < 3 || segs[0] == "" || segs[2] == "" {
		return nil
	}
	return map[string]any{"kind": segs[0], "scope": segs[1], "name": segs[2]}
}

// eventsRequest POST /v1/events 请求体（v0.20 发布面；source 仅接受 human 值或缺省——
// 服务侧强制判定，其他值=400）。
type eventsRequest struct {
	Type     string          `json:"type"`
	Involved string          `json:"involved"`
	Level    string          `json:"level"`
	DedupKey string          `json:"dedupKey"`
	Scope    string          `json:"scope"`  // 可选：metadata.scope 四层词显式声明（BUG-010；越词 400，缺席按 involved 推导）
	Source   string          `json:"source"` // 载荷声明不采纳：非 human 值直接拒绝（来源=服务侧判定）
	Subject  *subjectRequest `json:"subject"`
}

// writePublishErr 发布面错误 → HTTP 状态码映射（cored-bridge.md §2.3 v0.20）。
func writePublishErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, eventpub.ErrNotRegistered):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
	case errors.Is(err, eventpub.ErrDenied):
		writeJSON(w, http.StatusForbidden, map[string]any{"error": err.Error()})
	case errors.Is(err, eventpub.ErrNoType), errors.Is(err, eventpub.ErrLevel),
		errors.Is(err, eventpub.ErrScope), errors.Is(err, eventbus.ErrRejected):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
}

// eventTierOf type→(tier, visibility) 投影（BUG-010，event.md 规则 18——注册表唯一权威；
// 未注册类型=domain/default-visible 缺省词形，不误伤野类型/旧类型行）。
func eventTierOf(sch *scheme.Scheme, typ string) (string, string) {
	if et, err := sch.LookupEventType(typ); err == nil {
		return et.EffectiveTier(), et.EffectiveVisibility()
	}
	return scheme.TierDomain, scheme.VisibilityDefaultVisible
}

// approvalsPayload 收件箱批准投影行（/v1/inbox approvals.records 载荷——pending/approved/
// rejected 三态词形随行；judgedAt 零值=未判定不投影）。
func approvalsPayload(rs []flowbind.ApprovalRecord) []map[string]any {
	out := make([]map[string]any, 0, len(rs))
	for _, r := range rs {
		out = append(out, map[string]any{
			"seq": r.Seq, "kind": r.Kind, "workflow": r.Workflow, "scope": r.Scope,
			"connector": r.Connector, "proposal": r.Proposal, "state": r.State,
			"autoLevel": r.AutoLevel, "rv": r.RV, "eventId": r.EventID,
			"approvalId": r.ApprovalID, "judgedBy": r.JudgedBy,
			"judgedAt": nilIfZero(r.JudgedAt), "at": r.At.UTC().Format(time.RFC3339Nano),
		})
	}
	return out
}

// nilIfZero 零值时刻不投影（nil=未判定——与 SessionView HealedAt 同口径）。
func nilIfZero(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.UTC().Format(time.RFC3339Nano)
}
