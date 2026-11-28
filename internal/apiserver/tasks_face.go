// tasks_face.go —— 任务队列读面（concepts/task-scheduler.md §7——cored-bridge.md §2
// `GET /v1/tasks` 列表 + `GET /v1/tasks/{scope}/{name}` 详情）：数据源=文档库 Task 对象
// （跨 scope 枚举/单取，零缓存现读——git/sqlite 后端权威）；列表=status 投影简表
// （phase/attempts/agentSession/lastAction/lastError/dependsOn——任务中心与 corectl 同源）；
// 详情=spec+status 投影+原文。只读 GET 恒开（objects 缺席=503 与 objects 面同口径）；
// kind 未注册（旧 scheme）=404 与 scheme 闸门语义一致。status 写面不在此（P2：只由
// 调度器/执行体写——本面零写路径，任务提交/取消走 apply/events 既有写链）。
package apiserver

import (
	"sort"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheduler"
)

// taskKind 读面锚（scheme.yaml kind: Task——未注册=旧声明面，404）。
const taskKind = "Task"

// taskRowSpec #16 关联面回投：spec.agent=最强关联字段（哪个 agent 做的），
// title/summary/source/sourceRef=对象页文本匹配 hay——此前 taskRow 只投 name/title，
// objects.js 的 spec.* 全 undefined，匹配退化成两项。ownerRefs=归属对象引用反查。
type taskRowSpec struct {
	Title     string   `json:"title,omitempty"`
	Summary   string   `json:"summary,omitempty"`
	Source    string   `json:"source,omitempty"`
	SourceRef string   `json:"sourceRef,omitempty"`
	Agent     string   `json:"agent,omitempty"`
	OwnerRefs []string `json:"ownerRefs,omitempty"`
}

// taskRow 列表瘦身行（status 投影简表——排队原因/在做什么/谁在做一屏可见）。
type taskRow struct {
	Name            string   `json:"name"`
	Scope           string   `json:"scope"`
	ResourceVersion int64    `json:"resourceVersion"`
	Title           string   `json:"title"`
	Priority        string   `json:"priority"`
	Phase           string   `json:"phase"`
	Attempts        int      `json:"attempts"`
	AgentSession    string   `json:"agentSession,omitempty"`
	DependsOn       []string `json:"dependsOn"`
	LastAction      string   `json:"lastAction,omitempty"`
	LastError       string   `json:"lastError,omitempty"`
	// #16 关联面顶层镜像（扁平快照行兼容——前端两种词形都读得到）：
	Agent     string      `json:"agent,omitempty"`
	OwnerRefs []string    `json:"ownerRefs,omitempty"`
	Spec      taskRowSpec `json:"spec"`
}

// sortTasks 确定性投影：phase 活动态优先（running/verifying 在前）→ priority 权重 →
// scope/name 字典序（跨 scope 枚举顺序不稳定面收敛，与 proposals 面 rv 降序同口径——
// 任务面按现势排序更贴任务中心语义）。
func sortTasks(rows []taskRow) {
	rank := map[string]int{
		scheduler.PhaseRunning: 0, scheduler.PhaseVerifying: 0,
		scheduler.PhaseReady: 1, scheduler.PhaseQueued: 1, scheduler.PhaseBlocked: 1,
		scheduler.PhaseDone: 2, scheduler.PhaseFailed: 2, scheduler.PhaseCanceled: 2,
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rank[rows[i].Phase] != rank[rows[j].Phase] {
			return rank[rows[i].Phase] < rank[rows[j].Phase]
		}
		pi, pj := rows[i].Priority, rows[j].Priority
		order := map[string]int{"critical": 0, "high": 1, "normal": 2, "low": 3}
		if order[pi] != order[pj] {
			return order[pi] < order[pj]
		}
		if rows[i].Scope != rows[j].Scope {
			return rows[i].Scope < rows[j].Scope
		}
		return rows[i].Name < rows[j].Name
	})
}

// taskRowOf 对象 → 瘦身行（解析失败=零值行照出——读面不放大声明形变；详情原文兜底可查）。
func taskRowOf(name, scope string, rv int64, content []byte) taskRow {
	r := taskRow{Name: name, Scope: scope, ResourceVersion: rv, DependsOn: []string{}, OwnerRefs: []string{},
		Spec: taskRowSpec{OwnerRefs: []string{}}}
	spec, st, ok := scheduler.Parse(content)
	if !ok {
		return r
	}
	r.Title, r.Priority, r.Phase = spec.Title, spec.Priority, st.Phase
	r.Attempts, r.AgentSession = st.Attempts, st.AgentSession
	r.LastAction, r.LastError = st.LastAction, st.LastError
	// #16 关联面：最强字段 spec.agent + ownerRefs 回投（顶层镜像与嵌套 spec 同值，
	// 快照扁平行与嵌套行两词形前端均可直读）。
	r.Agent, r.OwnerRefs = spec.Agent, spec.OwnerRefs
	r.Spec = taskRowSpec{
		Title: spec.Title, Summary: spec.Summary, Source: spec.Source, SourceRef: spec.SourceRef,
		Agent: spec.Agent, OwnerRefs: spec.OwnerRefs,
	}
	if len(st.DependsOn) > 0 {
		r.DependsOn = st.DependsOn
	}
	return r
}

// taskDetail 单条详情投影（spec+status 全量+原文——声明演进不破坏查询面；解析失败=空值诚实）。
func taskDetail(o *repository.Object) map[string]any {
	spec, st, ok := scheduler.Parse(o.Content)
	dependsOn := []string{}
	if ok && len(spec.DependsOn) > 0 {
		dependsOn = spec.DependsOn
	}
	statusDeps := []string{}
	if ok && len(st.DependsOn) > 0 {
		statusDeps = st.DependsOn
	}
	return map[string]any{
		"name": o.Name, "scope": o.Scope, "resourceVersion": o.ResourceVersion,
		"spec": map[string]any{
			"title": spec.Title, "summary": spec.Summary, "source": spec.Source,
			"sourceRef": spec.SourceRef, "scopeRef": spec.ScopeRef,
			"dependsOn": dependsOn, "priority": spec.Priority, "agent": spec.Agent,
			"ownerRefs": spec.OwnerRefs,
			"touches": map[string]any{
				"repos": spec.Touches.Repos, "paths": spec.Touches.Paths,
				"objects": spec.Touches.Objects, "exclusive": spec.Touches.Exclusive,
			},
			"workspace": map[string]any{
				"mode": spec.Workspace.Mode, "base": spec.Workspace.Base, "branch": spec.Workspace.Branch,
			},
			"limits": map[string]any{"maxConcurrent": spec.MaxConcurrent, "timeout": int64(spec.Timeout / 1e6)},
		},
		"status": map[string]any{
			"phase": st.Phase, "attempts": st.Attempts, "agentSession": st.AgentSession,
			"dependsOn": statusDeps, "lastAction": st.LastAction, "lastError": st.LastError,
		},
		"content": string(o.Content),
	}
}
