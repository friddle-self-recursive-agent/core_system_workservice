// routes_workflow.go —— workflow/触发域路由组（R-ENG-7 server.go 拆分：纯搬家零语义变化——
// 触发骨架/派发决策/run/送达/唤醒/派单投影/workflow 聚合面自 server.go 原位迁入，handler 体
// 逐字节保持不动；装配与公共 helper 仍在 server.go）。
package apiserver

import (
	"context"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/flowbind"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheduler"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
)

// registerWorkflowRoutes workflow/触发域路由注册（纯搬家）。
func registerWorkflowRoutes(d *deps, mux *http.ServeMux) {
	triggers := d.triggers
	runs := d.runs
	deliveries := d.deliveries
	wakes := d.wakes
	objects := d.objects
	o := d.o
	dispatch := o.dispatch
	serviceSpawns := o.serviceSpawns
	eventLog := o.eventLog
	applyLog := o.applyLog
	inbox := o.inbox
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
		writeJSON(w, http.StatusOK, triggersPayload(st))
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
	// 对账 run 投影面（§2.3 v0.15 引擎主体第三段——真实拉起第一子段：workflow 触发面接线；
	// v0.23 相位推进器首段：受理即启动——引擎单步迁移判定随受理同拍完成，builtin run
	// phase=声明第二相位∧status=running；v0.25 显式迁移表首段：载荷新增 transitions=物化
	// 声明表行投影（新载荷键=代际判别器）；v0.26 异步推进首段：载荷新增 advances=累计异步
	// 推进计数（新载荷键=代际判别器）；v0.28 对象源相位投影与判定接线：object 源 run 相位/
	// 状态随对象自声明 spec 参与同一表驱动判定——未声明 phases=零投影口径不变）。run 出现 ≠
	// 对账执行完毕口径不变。旧构建（face 缺席）经兜底 handler 返回 405=代际判别器
	//（与 loop-13 固化口径一致）。
	mux.HandleFunc("GET /v1/workflows/runs", func(w http.ResponseWriter, _ *http.Request) {
		if runs == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "run 投影面未接线"})
			return
		}
		st := runs.State()
		writeJSON(w, http.StatusOK, runsPayload(st))
	})
	// service 拉起投影面（2026-11-27 service session 裁决「builtin 的 service 的 agent 可以用
	// session 来解决。突破下限制」——workflow.md §1.3）：run 落 unit=service 节点挂起
	// （suspendWhy=service）+ 为 service 的 agent 开执行 session 的拉起投影逐条可查
	// （run/节点/服务/op/agent/上下文字段/状态/会话锚）。拉起投影出现 ≠ session 已执行完毕
	// （诚实投影与 run/唤醒面同口径）。face 缺席=503（防御面）；旧构建 405 兜底=代际判别器。
	mux.HandleFunc("GET /v1/workflows/service-spawns", func(w http.ResponseWriter, _ *http.Request) {
		if serviceSpawns == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "service 拉起投影面未接线（ServiceSpawner 未挂接——零翻转口径）"})
			return
		}
		st := serviceSpawns.State()
		writeJSON(w, http.StatusOK, map[string]any{
			"spawned": st.Spawned, "dropped": st.Dropped, "records": st.Records,
		})
	})
	// 送达投影面（§2.3 v0.16 引擎主体第四段——真实拉起第二子段：notification 送达面接线；
	// v0.17 收尾：受理后的真实通道出站挂接——status 翻转 accepted→sent/failed + 出站留痕）：
	// 执行面=notification 且装配层已接线的绑定（route 主路与 cc 旁路，origin 维度留痕）推进为
	// 一条送达投影（送达已受理+通道投影；出站注入后异步翻转终态，未注入恒 accepted）。
	// 旧构建（face 缺席）经兜底 handler 返回 405=代际判别器（与 runs 面同口径）；
	// 新代新增载荷键 sent/failed/attempts/completedAt（恒在）与 error（有错才在）——
	// 键的出现/缺席=代际判别器（loop-16 口径）。
	mux.HandleFunc("GET /v1/notifications/deliveries", func(w http.ResponseWriter, _ *http.Request) {
		if deliveries == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "送达投影面未接线"})
			return
		}
		st := deliveries.State()
		writeJSON(w, http.StatusOK, deliveriesPayload(st))
	})
	// 唤醒投影面（§2.3 v0.18 引擎主体第五段——真实拉起第三子段：agent 触发面接线）：
	// 执行面=agent（`<p>-agent` 族：root-agent/parent-agent）且装配层已接线的绑定（route
	// 主路与 cc 旁路，origin 维度留痕）推进为一条唤醒投影（唤醒已受理+agent 锚投影；
	// 唤醒投影出现 ≠ agent 会话已拉起——诚实投影与 run/送达面同口径）。Escalate→root-agent
	// 行的机械观察点。旧构建（face 缺席）经兜底 handler 返回 405=代际判别器（runs/送达面
	// 同口径）。
	mux.HandleFunc("GET /v1/agents/wakes", func(w http.ResponseWriter, _ *http.Request) {
		if wakes == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "唤醒投影面未接线"})
			return
		}
		st := wakes.State()
		writeJSON(w, http.StatusOK, wakesPayload(st))
	})
	// 派单受理投影面（v0.50 R-8——workflow-runtime.md §11.2 拉起形态首段；v0.51 扩展=第二触发源
	// Proposal 批准派单）：骨架物化动作（created|updated——「该版本实现未落」的机械签名，kind=workflow）
	// 与 Proposal 批准落库（action=approved，kind=proposal——concepts/proposal.md 规则 2）→
	// apply-workflow 锚派单受理投影；noop/skipped 不派单（无新实现可做）。载荷 {count, accepted,
	// dropped, dispatches}，行含 kind/agent/status/autoLevel/action/锚/rv/eventId（kind=workflow 行含
	// workflow/connector；kind=proposal 行含 proposal/title/summary/acceptance/priority/proposer——
	// 键恒在投影，空值=不适用）。**派单受理 ≠ agent 会话已拉起**（诚实投影与 wake 面同口径：真实拉起=
	// dsh 会话挂载/agent-session 域，declared-deferred）；进程内环（观测面非账本——持久态=骨架对象
	// execution.skeleton=true+apply_log 审计+Proposal 状态机，重启归零=幂等重入不重复派单）。
	// face 缺席=503（防御面，装配层 --allow-apply 随 flowbind/proposal 注入）；旧构建（路由未注册）
	// 经兜底 handler 405=代际判别器。
	mux.HandleFunc("GET /v1/agents/dispatches", func(w http.ResponseWriter, _ *http.Request) {
		if dispatch == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "派单受理投影面未接线（装配缺 flowbind 派单 sink）"})
			return
		}
		st := dispatch.State()
		writeJSON(w, http.StatusOK, dispatchesPayload(st, inbox))
	})
	// 当前/最近一轮自循环投影面（v0.66 R-ENG §3——`GET /v1/loop/current`）：**只读聚合
	// 不发明状态机**。数据锚=wakes face 最近一条 LoopRequested 唤醒投影（wake.go——自循环
	// 拉起=LoopRequested 事件经 dispatch 表 root-agent 行推进的唤醒受理，corectl loop /
	// wake 桥同源）；seq/startedAt/eventId/wake 行=该锚直投（无 wake 数据=字段 null 不编造）；
	// G-08 读面补强（event 详情↔最近 loop↔session 链）：
	//   - run/phase：runs face 环内回指该 wake 的真实 RunRecord（EventID 同源优先；兜底=
	//     对账循环锚 core-workflow 最近 run——「最近一次 loop 的 run 相位」投影，不发明新
	//     run）；phase=run 相位顶层词形（console 当前循环卡读词）。runs face 缺席/无命中=null。
	//   - agent：wake 行 agent 锚直投（root/parent 等 <p>-agent 前缀段）。
	//   - session：dsh 会话 id——真实拉起=agent-session 域 declared-deferred（本构建无循环
	//     会话数据源：intake 会话是录入会话——恒 null 不冒充）。
	//   - agentSession：run involved 解析为 Task 对象时回填 status.agentSession（G-08
	//     「Task 无真实 sessionId 回填」的读面半边——真实回填链；非 Task/对象缺席/未回填=null）。
	// status 机械判定词：有 LoopRequested 唤醒投影=running（唤醒已受理——真实循环态机
	// deferred 不发明 idle 判据，idle=词形预留）、无任何 wake 数据=unknown。face 缺席=
	// status unknown+字段 null（不冒充空轮次）。只读 GET 恒开；旧构建（路由未注册）经
	// 兜底 handler 405=代际判别器。
	mux.HandleFunc("GET /v1/loop/current", func(w http.ResponseWriter, r *http.Request) {
		var latest *workflow.WakeRecord
		if wakes != nil {
			for i := len(wakes.State().Wakes) - 1; i >= 0; i-- { // 旧→新序——尾向扫描取最近一条
				wk := wakes.State().Wakes[i]
				if wk.Type == "LoopRequested" {
					latest = &wk
					break
				}
			}
		}
		out := map[string]any{
			"seq":          nil,
			"startedAt":    nil,
			"eventId":      nil,
			"wake":         nil,
			"session":      nil, // dsh 会话 id：真实拉起 deferred——无数据源恒 null 不编造（见路由注）
			"agentSession": nil, // run involved→Task status.agentSession 回填链（G-08）——无链=null
			"run":          nil, // 最近真实 RunRecord 投影（EventID 同源优先/core-workflow 兜底）——无命中=null
			"phase":        nil, // run 相位顶层词形（console 当前循环卡读词）
			"agent":        nil, // wake 行 agent 锚直投
			"status":       "unknown",
		}
		if latest != nil {
			out["seq"] = latest.Wake
			out["startedAt"] = latest.At.UTC().Format(time.RFC3339Nano)
			out["eventId"] = latest.EventID
			out["agent"] = latest.Agent
			out["wake"] = map[string]any{
				"wake": latest.Wake, "agent": latest.Agent, "origin": latest.Origin,
				"status": latest.Status, "level": latest.Level, "triggerSeq": latest.TriggerSeq,
				"eventId": latest.EventID, "type": latest.Type, "dedupKey": latest.DedupKey,
				"route": latest.Route, "owner": latest.Owner, "at": latest.At.UTC().Format(time.RFC3339Nano),
			}
			out["status"] = "running" // 唤醒受理态（机械词——非执行完毕非循环态机推断）
			if runs != nil {
				if rr := loopRunOf(runs.State(), latest.EventID); rr != nil {
					out["run"] = map[string]any{
						"run": rr.Run, "workflow": rr.Workflow, "phase": rr.Phase, "status": rr.Status,
						"startedAt": rr.StartedAt.UTC().Format(time.RFC3339Nano),
					}
					out["phase"] = rr.Phase
					out["agentSession"] = loopSessionOf(r.Context(), objects, rr.Involved)
				}
			}
		}
		writeJSON(w, http.StatusOK, out)
	})
	// workflow 运行管理聚合面（v0.37 console 读面第二段，E3——cored-bridge.md §2
	// `GET /v1/workflows`）：三源同投（① 声明=builtin 注册表+文档库 Workflow 对象；
	// ② 现势=run 环每锚最新 run 的相位/状态；③ 计数=event_log 归档分组 × dispatch 决议，
	// workflow.RunAnchor 与 run 投影同一锚解析）。只读 GET；旧构建（路由未注册）经兜底
	// handler 405=代际判别器。投影/构造细节见 workflows_face.go。
		mux.HandleFunc("GET /v1/workflows", func(w http.ResponseWriter, r *http.Request) {
			rows, scanned, capped := collectWorkflowRowsWithScheme(r.Context(), d.sch, objects, runs, triggers, eventLog, applyLog)
			items := make([]map[string]any, 0, len(rows))
			for _, row := range rows {
				specMap, rawYaml, primaryScope, primaryRV := primaryWfSpecAndRaw(row.Name, row.Objects)
				items = append(items, map[string]any{
					"name": row.Name, "specSource": row.SpecSource, "phases": row.Phases,
					"objects": workflowObjectsJSON(row.Objects), "effective": row.Effective,
					"latest": row.Latest, "runsTotal": row.RunsTotal, "runs24h": row.Runs24h,
					"execution": row.Execution, // v0.49 R-7 执行模型绑定面（新载荷键=代际判别器）
					// v0.55 R-9 构建生效链判别面（console-adoption.md §1.29——骨架 rv ↔ 绑定 apply 行
					// rv 对账+骨架节点接线投影；新载荷键=代际判别器）。
					"reconcile":        row.Reconcile,
					"executionFlow":    flowNodesJSON(row.Reconcile),
					"executionDigest":  executionDigest(row.Reconcile),
					"executionHistory": workflowHistoryRef(row),
					// 2026-11-27 Scheme 解释器声明投影（供前端全节点可视化配置器直读直编）
					"undeletable":     d.sch.IsUndeletableWorkflow(row.Name),
					"immutableNodes":  row.Name == "core-system-developer-loop",
					"spec":            specMap,
					"rawYaml":         rawYaml,
					"primaryScope":    primaryScope,
					"primaryRv":       primaryRV,
				})
			}
			writeJSON(w, http.StatusOK, map[string]any{
				"count": len(items), "window": "24h",
				"scanned": scanned, "scanCapped": capped, // 计数=有界窗如实标注（不冒充全量承诺）
				"aliases":  d.sch.WorkflowAliases(),
				"adapters": d.sch.WorkflowAdapters(),
				"llm":      d.sch.LLMRuntime(),
				// 2026-11-28 #21：connectops 引擎挂接态现读投影（与 cmd/cored 装配同一
				// CORE_SYSTEM_CONNECTOPS_RUN=on 判据）——决策面执行状态三值（bound/partial/
				// unbound）的引擎门输入。缺席/on 之外的取值一律 false（不冒充已挂接）。
				"connectopsRun": os.Getenv("CORE_SYSTEM_CONNECTOPS_RUN") == "on",
				"workflows":     items,
			})
		})
	}

// matchingKeyOfRecord 派单行的批准判定键（/v1/agents/dispatches approval.approvalId——
// flowbind.DispatchRecord → matchingKeyOf 同源；空 connector/proposal 行=零值键不匹配任何投影）。
func matchingKeyOfRecord(d flowbind.DispatchRecord) string {
	return flowbind.ApprovalKeyOf(d)
}

// loopRunOf 最近一次 loop run 解析（G-08——loop/current run 相位源；环内**真实投影**零发明）：
// EventID 与 loop wake 同源的 run 优先（该次唤醒推进的对账 run）；兜底=对账循环锚
// core-system-developer-loop / core-workflow 最近 run。无命中=nil（face 环空/无 run——不编造）。
func loopRunOf(st workflow.RunState, eventId string) *workflow.RunRecord {
	for i := len(st.Runs) - 1; i >= 0; i-- {
		if st.Runs[i].EventID == eventId {
			return &st.Runs[i]
		}
	}
	for i := len(st.Runs) - 1; i >= 0; i-- {
		if st.Runs[i].Workflow == "core-system-developer-loop" || st.Runs[i].Workflow == workflow.ReconcileWorkflow {
			return &st.Runs[i]
		}
	}
	return nil
}

// loopSessionOf run involved → Task 对象 agentSession 回填链（G-08 读面半边——session 深链
// 的真实数据源）：involved 三段解析且 kind=Task → 现读 Task 对象 → scheduler.Parse 的
// status.agentSession。非 Task 引用/对象缺席/未回填=nil 不编造（诚实投影——写面回填属
// scheduler 执行域，本面只投已回填事实）。
func loopSessionOf(ctx context.Context, objects ObjectSource, involved string) any {
	segs := strings.Split(involved, "/")
	if len(segs) < 3 || segs[0] != taskKind || objects == nil {
		return nil
	}
	o, err := objects.Get(ctx, taskKind, segs[1], segs[2])
	if err != nil {
		return nil
	}
	_, st, ok := scheduler.Parse(o.Content)
	if !ok || st.AgentSession == "" {
		return nil
	}
	return st.AgentSession
}

// triggersPayload GET /v1/triggers 载荷构造器（原 handler 内联体原样上提——snapshot 面
// 与路由共用同一构造器零双源，R-ENG-7 拆分注记）。
func triggersPayload(st workflow.TriggerState) map[string]any {
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
	return map[string]any{
		"count": len(records), "received": st.Received, "dropped": st.Dropped,
		"types": st.Types, "records": records,
	}
}

// runsPayload GET /v1/workflows/runs 载荷构造器（原 handler 内联体原样上提——注释同源随行）。
func runsPayload(st workflow.RunState) map[string]any {
	items := make([]map[string]any, 0, len(st.Runs))
	for _, r := range st.Runs {
		items = append(items, map[string]any{
			"run": r.Run, "workflow": r.Workflow, "specSource": r.SpecSource,
			"phases": r.Phases, "transitions": r.Transitions, "phase": r.Phase, "status": r.Status,
			"triggerSeq": r.TriggerSeq, "eventId": r.EventID, "type": r.Type,
			"dedupKey": r.DedupKey, "route": r.Route, "owner": r.Owner,
			"target": r.Target, "startedAt": r.StartedAt.UTC().Format(time.RFC3339Nano),
			// v0.37 console 读面第二段：来源留痕 involved 投影（run 涉及对象链第一跳——
			// 运行抽屉「涉及文档」取本字段落对象链接，无前端拼凑；新载荷键=代际判别器）。
			"involved": r.Involved,
			// #16 执行 agent 锚（与 involved 被涉对象正交——Agent 对象页按 agent 反查相关运行）。
			"agent": r.Agent,
			// 会话锚（G-08 loop 会话关联；空=未回填——节点/任务会话链接的真实来源，不编造）。
			"sessionId": r.SessionID,
			// v0.32 刹车首段：生效刹车值投影（声明→留痕面——spec.brakes 钳定后的每锚生效值；
			// 新载荷键=代际判别器）。
			"brakes": map[string]any{
				"cooldown":         r.Brakes.Cooldown.String(),
				"maxMutationsTick": r.Brakes.MaxMutationsTick,
				"source":           r.Brakes.Source,
			},
			// 审批策略投影（2026-11-26「Workflow 可设置 Agent 节点是否自动级别」）：
			// "auto"=闸锚自动通过；空=human 缺省（闸锚挂起等人）。新载荷键=代际判别器。
			"approvalMode": r.ApprovalMode,
			"approvalPolicy": map[string]any{
				"autoCriteria": r.ApprovalPolicy.AutoCriteria, "maxRiskLevel": r.ApprovalPolicy.MaxRiskLevel,
			},
			"suspendWhy":  r.SuspendWhy,               // 挂起细分（self-approve|service|inner-agent|fanout|空=人审）
			// F-04 运行轨迹投影（2026-11-26：静态迁移规则表 transitions 与运行轨迹 history
			// 分离——UI 节点图按 history 标绿已走过的真实节点，静态表不再当事件时间线展示。
			// 零轨迹=nil 词形缺席语义）。新载荷键=代际判别器。
			"history": historyRowsPayload(r.History),
			// Verify 判定随行（G-09/G-12 已在留痕——nil=未审不投影键）。
		})
	}
	return map[string]any{
		"count": len(items), "started": st.Started, "dropped": st.Dropped,
		"advances": st.Advances, // v0.26 异步推进首段：累计推进 run·步计数（新载荷键=代际判别器）
		"braked":   st.Braked,   // v0.32 刹车首段：累计冷却门拦阻触发数（新载荷键=代际判别器）
		"capped":   st.Capped,   // v0.32 刹车首段：累计推进拍上限拦阻 run·步数（新载荷键=代际判别器）
		"disabled": st.Disabled, // v0.58b enabled honoring：累计 enabled=false 锚拦阻触发数（新载荷键=代际判别器）
		"selfApproved": st.SelfApproved, // 2026-11-28 #4：累计自审批真判消费数
		"runs":     items,
	}
}

// historyRowsPayload 运行轨迹行投影（F-04——nil=零轨迹不投影空数组占位；行字段直投）。
func historyRowsPayload(rows []workflow.HistoryRow) []map[string]any {
	if len(rows) == 0 {
		return nil
	}
	out := make([]map[string]any, 0, len(rows))
	for _, h := range rows {
		row := map[string]any{"from": h.From, "to": h.To, "on": h.On, "at": h.At}
		if h.SessionID != "" {
			row["sessionId"] = h.SessionID
		}
		out = append(out, row)
	}
	return out
}

// deliveriesPayload GET /v1/notifications/deliveries 载荷构造器（原 handler 内联体原样上提）。
func deliveriesPayload(st workflow.DeliveryState) map[string]any {
	items := make([]map[string]any, 0, len(st.Deliveries))
	for _, d := range st.Deliveries {
		completed := ""
		if !d.CompletedAt.IsZero() {
			completed = d.CompletedAt.UTC().Format(time.RFC3339Nano)
		}
		item := map[string]any{
			"delivery": d.Delivery, "channel": d.Channel, "origin": d.Origin,
			"status": d.Status, "triggerSeq": d.TriggerSeq, "eventId": d.EventID,
			"type": d.Type, "dedupKey": d.DedupKey, "level": d.Level,
			"route": d.Route, "owner": d.Owner, "at": d.At.UTC().Format(time.RFC3339Nano),
			"attempts": d.Attempts, "completedAt": completed,
		}
		if d.Error != "" {
			item["error"] = d.Error
		}
		items = append(items, item)
	}
	return map[string]any{
		"count": len(items), "accepted": st.Accepted, "sent": st.Sent, "failed": st.Failed,
		"dropped": st.Dropped,
		// v0.35 通知门链三计数（KI-001——新载荷键=代际判别器，键恒在口径）：
		// gated=阶段门拦阻（非相位终态事实）；filtered=minLevel 等级过滤；deduped=窗口去重聚合。
		"gated": st.Gated, "filtered": st.Filtered, "deduped": st.Deduped,
		"deliveries": items,
	}
}

// wakesPayload GET /v1/agents/wakes 载荷构造器（原 handler 内联体原样上提）。
func wakesPayload(st workflow.WakeState) map[string]any {
	items := make([]map[string]any, 0, len(st.Wakes))
	for _, wk := range st.Wakes {
		items = append(items, map[string]any{
			"wake": wk.Wake, "agent": wk.Agent, "origin": wk.Origin,
			"status": wk.Status, "level": wk.Level, "triggerSeq": wk.TriggerSeq,
			"eventId": wk.EventID, "type": wk.Type, "dedupKey": wk.DedupKey,
			"route": wk.Route, "owner": wk.Owner, "at": wk.At.UTC().Format(time.RFC3339Nano),
		})
	}
	return map[string]any{
		"count": len(items), "accepted": st.Accepted, "dropped": st.Dropped, "wakes": items,
	}
}

// dispatchesPayload GET /v1/agents/dispatches 载荷构造器（原 handler 内联体原样上提；
// inbox=批准投影面——行随行 approval/spawn 键，face 缺席=零值不冒充）。
func dispatchesPayload(st flowbind.DispatchState, inbox InboxFace) map[string]any {
	items := make([]map[string]any, 0, len(st.Records))
	// v0.56 ask 批准动作面（workflow-runtime.md §11.5——拉起投影面第三段）：派单行随行
	// approval 键（state 词=批准投影翻转词——人批/驳回=Apply 判定；审批投影面缺席=零值不冒充）。
	approvalIdx := map[string]string{}
	spawnIdx := map[string]flowbind.SpawnDeclaration{}
	if inbox != nil {
		for _, r := range inbox.State().Records {
			approvalIdx[r.ApprovalID] = r.State
			if d := flowbind.SpawnOf(r); d.Declared {
				spawnIdx[r.ApprovalID] = d // v0.57 spawn 接线声明投影（approved 行——wired 恒 false）
			}
		}
	}
	for _, d := range st.Records {
		kind := d.Kind
		if kind == "" {
			kind = flowbind.DispatchKindWorkflow // v0.50 形记录投影补 kind（旧 sink 兼容）
		}
		items = append(items, map[string]any{
			"seq": d.Seq, "agent": d.Agent, "status": d.Status, "autoLevel": d.AutoLevel,
			"action": d.Action, "kind": kind,
			"workflow": d.Workflow, "scope": d.Scope, "connector": d.Connector,
			// v0.51 proposal 派单载荷（键恒在=代际判别器；kind=workflow 行空值=不适用）。
			"proposal": d.Proposal, "title": d.Title, "summary": d.Summary,
			"acceptance": d.Acceptance, "priority": d.Priority, "proposer": d.Proposer,
			"rv": d.RV, "eventId": d.EventID,
			// v0.55 节点接线载荷（拉起投影面首件，workflow-runtime.md §11.2 第三段——
			// apply-workflow 拉起输入=骨架节点清单；kind=proposal 行空值=不适用）。
			"nodes": d.Nodes,
			// v0.56 ask 批准动作面（拉起投影面第三段，workflow-runtime.md §11.5）：
			// state=pending|approved|rejected（approved≠spawn 已拉起——真实执行仍 observe）。
			"approval": map[string]any{"state": approvalIdx[matchingKeyOfRecord(d)], "approvalId": matchingKeyOfRecord(d)},
			// v0.57 spawn 接线声明投影（拉起投影面末段，workflow-runtime.md §11.7——零值键
			// =代际判别器）：approved 批准投影行 → {declared:true, wired:false, target, …}
			// 接线声明词形（approved≠拉起——真实拉起=agent-session 域 deferred，红线 a）。
			"spawn": spawnPayload(spawnIdx[matchingKeyOfRecord(d)]),
			"at":    d.At.UTC().Format(time.RFC3339Nano),
		})
	}
	return map[string]any{
		"count": len(items), "accepted": st.Accepted, "dropped": st.Dropped, "dispatches": items,
	}
}

// spawnPayload spawn 接线声明 → 载荷行（/v1/agents/dispatches spawn 键——键恒在=代际判别器；
// 零值声明=接线缺席 `{declared:false, wired:false}` 诚实词形——pending/rejected 行或批准投影
// 面缺席照落零值，不冒充已接线）。
func spawnPayload(d flowbind.SpawnDeclaration) map[string]any {
	return map[string]any{
		"declared": d.Declared, "wired": d.Wired, "target": d.Target,
		"approvalId": d.ApprovalID, "rv": d.RV, "eventId": d.EventID,
	}
}
