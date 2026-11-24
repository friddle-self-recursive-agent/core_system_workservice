// proposals_face.go —— 需求提案读面（v0.51 能力 B——cored-bridge.md §2 `GET /v1/proposals`）：
// 数据源=文档库 Proposal 对象（跨 scope 枚举/单取，零缓存现读——git 后端权威）；列表=瘦身投影
// （rv 降序确定性）；详情=全 spec 投影+原文。派单投影不在本面（详情屏组合 /v1/agents/dispatches
// 的 proposal 过滤行——聚合零 duplication）。只读 GET 恒开（objects 缺席=503 与 objects 面同口径）；
// kind 未注册（旧 scheme）=404 与 scheme 闸门语义一致；非 GET 由方法路由 405/兜底 405=代际判别器。
package apiserver

import (
	"sort"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/proposal"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
)

// proposalRow 列表瘦身行（详情另取——与 objects 面瘦身同口径）。
type proposalRow struct {
	Name            string         `json:"name"`
	Scope           string         `json:"scope"`
	ResourceVersion int64          `json:"resourceVersion"`
	Title           string         `json:"title"`
	Summary         string         `json:"summary"`
	Status          string         `json:"status"`
	Priority        string         `json:"priority"`
	Proposer        string         `json:"proposer"`
	AcceptanceCount int            `json:"acceptanceCount"`
	Landing         map[string]any `json:"landing"`
}

// proposalKind 读面锚（scheme.yaml kind: Proposal——未注册=旧声明面，404）。
const proposalKind = "Proposal"

func landingRow(t, ref string) map[string]any {
	m := map[string]any{"target": t}
	if ref != "" {
		m["ref"] = ref
	}
	return m
}

// sortProposals rv 降序（同 rv 名升序——确定性投影；跨 scope 枚举顺序不稳定面收敛）。
func sortProposals(rows []proposalRow) {
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].ResourceVersion != rows[j].ResourceVersion {
			return rows[i].ResourceVersion > rows[j].ResourceVersion
		}
		return rows[i].Name < rows[j].Name
	})
}

// proposalRowOf 对象 → 瘦身行（解析失败=零值行照出——读面不放大声明形变；详情原文兜底可查）。
func proposalRowOf(name, scope string, rv int64, content []byte) proposalRow {
	r := proposalRow{Name: name, Scope: scope, ResourceVersion: rv, Landing: map[string]any{}}
	if spec, ok := proposal.Parse(content); ok {
		r.Title, r.Summary, r.Status = spec.Title, spec.Summary, spec.Status
		r.Priority, r.Proposer = spec.Priority, spec.Proposer
		r.AcceptanceCount = len(spec.Acceptance)
		if spec.LandingTarget != "" || spec.LandingRef != "" {
			r.Landing = landingRow(spec.LandingTarget, spec.LandingRef)
		}
	}
	return r
}

// pickProposal 详情单取（scope 缺省=global 优先→其余首个——与 execution 投影同确定性口径）。
func pickProposal(rows []proposalRow) *proposalRow {
	for i := range rows {
		if rows[i].Scope == "global" {
			return &rows[i]
		}
	}
	if len(rows) > 0 {
		return &rows[0]
	}
	return nil
}

// proposalDetail 单条详情投影（spec 全量+原文——声明演进不破坏查询面；解析失败=空值诚实，
// content 原文兜底可查）。
func proposalDetail(o *repository.Object) map[string]any {
	spec, ok := proposal.Parse(o.Content)
	acceptance := spec.Acceptance
	if acceptance == nil {
		acceptance = []string{}
	}
	landing := map[string]any{}
	if ok && (spec.LandingTarget != "" || spec.LandingRef != "") {
		landing = landingRow(spec.LandingTarget, spec.LandingRef)
	}
	return map[string]any{
		"name": o.Name, "scope": o.Scope, "resourceVersion": o.ResourceVersion,
		"title": spec.Title, "summary": spec.Summary, "status": spec.Status,
		"priority": spec.Priority, "proposer": spec.Proposer, "landing": landing,
		"spec": map[string]any{
			"title": spec.Title, "summary": spec.Summary, "acceptance": acceptance,
			"impact": spec.Impact, "priority": spec.Priority, "proposer": spec.Proposer,
			"status": spec.Status, "landing": landing,
		},
		"content": string(o.Content),
	}
}
