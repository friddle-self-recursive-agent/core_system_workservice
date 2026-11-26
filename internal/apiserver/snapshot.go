// snapshot.go —— console 首屏聚合快照（v0.66：GET /v1/snapshot + /v1/stream 推送共用）。
// 设计依据：cored-bridge.md §2 读面最小集扩展（R-ENG 任务书 §2）——一次返回 console
// 首屏聚合载荷 {schemeKinds, objects:{kind:[refs]}, inbox, events, runs摘要, tasks, sessions,
// deliveries, wakes, dispatches, proposals, drafts, metrics, generatedAt}。组装纪律：
// 内部直接复用 apiserver 现有各 face 的查询函数与各路由 handler 的既有载荷构造器
// （triggersPayload/runsPayload/deliveriesPayload/wakesPayload/dispatchesPayload/inboxPayload/
// eventRowOf——routes_*.go 与本件共用同一构造器零双源），不发明新查询；face 缺席的字段=null
// （不冒充——与各路由 503 防御面同口径词形）。本面只读零副作用。
package apiserver

import (
	"context"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
)

// snapshotKeyGeneratedAt generatedAt 载荷键（diff 不参与——重算时刻抖动不产生 patch 噪音，
// 消费方本地该键随初装快照定格；数据键全部参与 diff）。
const snapshotKeyGeneratedAt = "generatedAt"

// snapshotPayload 组装首屏聚合快照（纯读面现读；face 缺席=null）。
func snapshotPayload(d *deps) map[string]any {
	o := d.o
	out := map[string]any{
		snapshotKeyGeneratedAt: time.Now().UTC().Format(time.RFC3339Nano),
	}
	// schemeKinds：注册表 kind 名清单（scheme 恒在——装配自检闸门前置，无缺席词形）。
	out["schemeKinds"] = schemeKindNames(d.sch)
	// objects：{kind:[refs]} 跨 kind 现读（objects face 缺席=null；单 kind 读失败=该 kind
	// 跳过不冒充——读面降级不放大为整面缺席）。
	out["objects"] = objectsByKind(d.objects, d.sch)
	// inbox：通知中心收件箱投影（face 缺席=null）。
	out["inbox"] = faceOrNil(o.inbox != nil, func() any {
		return inboxPayload(d.objects, d.drafts, o.inbox, o.eventLog)
	})
	// events：事件归档最近窗投影（G-02 聚合完整性——console 事件流页快照聚合源，前端 store
	// events() 直读；face 缺席=null。GET /v1/events 同一行投影构造器零双源：默认视窗
	// tier≠infra + 有界窗——快照聚合面不冒充全量清单，全量走 GET /v1/events 翻页面）。
	out["events"] = faceOrNil(o.eventLog != nil, func() any {
		return eventsRowsSnapshot(o.eventLog, d.sch)
	})
	// runs 摘要（face 缺席=null；State() 同一查询——摘要=行集有界不冒充全量清单，
	// 全量走 GET /v1/workflows/runs）。
	out["runs"] = faceOrNil(d.runs != nil, func() any { return runsSummary(d.runs.State()) })
	// tasks：文档库 Task 简表（objects 缺席=null）。
	out["tasks"] = faceOrNil(d.objects != nil, func() any { return tasksRowsSnapshot(d.objects, d.sch) })
	// sessions：会话活性心跳行（face 缺席=null；limit 与 GET /v1/sessions 缺省同口径；
	// 读失败=null 不冒充空清单——诚实投影）。
	out["sessions"] = faceOrNil(o.sessionLiveness != nil, func() any {
		rows, err := o.sessionLiveness.Heartbeats(200)
		if err != nil {
			return nil
		}
		return rows
	})
	// deliveries/wakes/dispatches：各投影面载荷（face 缺席=null）。
	out["deliveries"] = faceOrNil(d.deliveries != nil, func() any {
		return deliveriesPayload(d.deliveries.State())
	})
	out["wakes"] = faceOrNil(d.wakes != nil, func() any { return wakesPayload(d.wakes.State()) })
	out["dispatches"] = faceOrNil(o.dispatch != nil, func() any {
		return dispatchesPayload(o.dispatch.State(), o.inbox)
	})
	// proposals：文档库 Proposal 简表（objects 缺席=null）。
	out["proposals"] = faceOrNil(d.objects != nil, func() any { return proposalsRowsSnapshot(d.objects, d.sch) })
	// drafts：草稿瘦身清单（缓存缺席=null）。
	out["drafts"] = faceOrNil(d.drafts != nil, func() any {
		items := make([]map[string]any, 0)
		for _, dd := range d.drafts.List() {
			items = append(items, draftRef(dd))
		}
		return items
	})
	// metrics：数据面指标清单（face 缺席=null——与 /v1/metrics 恒注册空投影词形分立：
	// 快照面缺席=缺席，不冒充空面）。
	out["metrics"] = faceOrNil(o.metrics != nil, func() any { return metricsItems(o.metrics) })
	return out
}

// eventsRowsSnapshot 事件归档最近窗投影（G-02 聚合完整性——GET /v1/events 同一行投影
// 构造器 eventRowOf 零双源；默认视窗隐藏 infra 分层、50 行有界窗；读失败=空清单不冒充）。
func eventsRowsSnapshot(eventLog EventLogFace, sch *scheme.Scheme) []map[string]any {
	rows, err := eventLog.QueryEvents(opslog.EventQuery{Limit: 50, HiddenTypes: sch.HiddenEventTypes()})
	if err != nil {
		return []map[string]any{}
	}
	items := make([]map[string]any, 0, len(rows))
	for _, e := range rows {
		items = append(items, eventRowOf(sch, e))
	}
	return items
}

// faceOrNil face 缺席→nil（JSON null）的兜底器（face 在=构造器求值）。
func faceOrNil(present bool, build func() any) any {
	if !present {
		return nil
	}
	return build()
}

// schemeKindNames 注册表 kind 名清单（ListKinds 顺序投影）。
func schemeKindNames(sch *scheme.Scheme) []string {
	kinds := sch.ListKinds()
	names := make([]string, 0, len(kinds))
	for _, k := range kinds {
		names = append(names, k.Name)
	}
	return names
}

// objectsByKind {kind:[refs]} 跨 kind 现读（List 跨 scope 通配——GET /v1/objects/{kind}
// 同一查询语义；G-15：refs 投影只需键/rv——元数据直读优先，读失败=该 kind 跳过；空清单
// kind 不占键——载荷瘦身）。
func objectsByKind(objects ObjectSource, sch *scheme.Scheme) map[string][]map[string]any {
	if objects == nil {
		return nil
	}
	out := map[string][]map[string]any{}
	for _, name := range schemeKindNames(sch) {
		objs, err := listMetaOrList(objects, context.Background(), name, "")
		if err != nil {
			continue // 读面降级：单 kind 缺席不冒充空清单
		}
		if len(objs) == 0 {
			continue
		}
		refs := make([]map[string]any, 0, len(objs))
		for _, ob := range objs {
			refs = append(refs, objRef(ob))
		}
		out[name] = refs
	}
	return out
}

// runsSummary run 投影摘要（State() 同一查询——计数全量投影 + 最近 10 条 run 行瘦身）。
func runsSummary(st workflow.RunState) map[string]any {
	recent := st.Runs
	if len(recent) > 10 {
		recent = recent[len(recent)-10:]
	}
	rows := make([]map[string]any, 0, len(recent))
	for _, r := range recent {
		rows = append(rows, map[string]any{
			"run": r.Run, "workflow": r.Workflow, "phase": r.Phase, "status": r.Status,
			"eventId": r.EventID, "involved": r.Involved,
			"startedAt": r.StartedAt.UTC().Format(time.RFC3339Nano),
		})
	}
	return map[string]any{
		"started": st.Started, "dropped": st.Dropped, "advances": st.Advances,
		"braked": st.Braked, "capped": st.Capped, "disabled": st.Disabled,
		"count": len(st.Runs), "recent": rows,
	}
}

// tasksRowsSnapshot Task 简表（GET /v1/tasks 同一查询语义：List+taskRowOf+sortTasks，
// 无过滤参数——首屏全量口径）。
func tasksRowsSnapshot(objects ObjectSource, sch *scheme.Scheme) []taskRow {
	if _, err := sch.Lookup(taskKind); err != nil {
		return []taskRow{} // kind 未注册=空简表（scheme 闸门语义与路由面同口径）
	}
	objs, err := objects.List(context.Background(), taskKind, "")
	if err != nil {
		return []taskRow{} // 读面降级不冒充
	}
	rows := make([]taskRow, 0, len(objs))
	for _, o := range objs {
		rows = append(rows, taskRowOf(o.Name, o.Scope, o.ResourceVersion, o.Content))
	}
	sortTasks(rows)
	return rows
}

// proposalsRowsSnapshot Proposal 简表（GET /v1/proposals 同一查询语义）。
func proposalsRowsSnapshot(objects ObjectSource, sch *scheme.Scheme) []proposalRow {
	if _, err := sch.Lookup(proposalKind); err != nil {
		return []proposalRow{}
	}
	objs, err := objects.List(context.Background(), proposalKind, "")
	if err != nil {
		return []proposalRow{}
	}
	rows := make([]proposalRow, 0, len(objs))
	for _, o := range objs {
		rows = append(rows, proposalRowOf(o.Name, o.Scope, o.ResourceVersion, o.Content))
	}
	sortProposals(rows)
	return rows
}

// metricsItems 数据面指标清单投影（GET /v1/metrics 同一构造）。
func metricsItems(m MetricsFace) []map[string]any {
	items := []map[string]any{}
	for _, name := range m.Metrics() {
		items = append(items, map[string]any{"name": name, "samples": m.Count(name)})
	}
	return items
}

// stripGeneratedAt 拷贝去掉 generatedAt 键（diff 基线——重算时刻不产生 patch 噪音）。
func stripGeneratedAt(snap map[string]any) map[string]any {
	out := make(map[string]any, len(snap))
	for k, v := range snap {
		if k == snapshotKeyGeneratedAt {
			continue
		}
		out[k] = v
	}
	return out
}
