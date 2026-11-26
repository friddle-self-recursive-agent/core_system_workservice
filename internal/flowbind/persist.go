// persist.go —— 派单投影写穿透/启动恢复的行转换与装载件（R-ENG §1 消费侧——flowbind
// DispatchRecord ↔ runtimestore.DispatchRow）。与 workflow/persist.go 同款口径：写穿透=
// 内存照旧工作（State() 语义零变化，读仍走内存），受理同步 Append；启动恢复=New 时装载
// 历史行续接序号（rt_dispatch_records 权威续接，memory 空启动=零行）。行字段直投自内存
// 结构：时间=RFC3339Nano 文本列，复杂字段（验收清单/节点清单）=JSON 文本列。
package flowbind

import (
	"encoding/json"
	"log"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/runtimestore"
)

// restorePage 恢复面分页步长（与 runtimestore listLimit 缺省同值——分页循环消费）。
const restorePage = 500

// stampNano RFC3339Nano 文本列词形（UTC 归一——排序字面序=时序）。
func stampNano(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.UTC().Format(time.RFC3339Nano)
}

// parseNano 文本列 → 时刻（空/不可解析=零值——恢复面诚实降级不炸装载）。
func parseNano(s string) time.Time {
	if s == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339Nano, s)
	if err != nil {
		return time.Time{}
	}
	return t
}

// mustJSON 结构体 → JSON 文本列（内存投影结构恒可 marshal——失败兜底零值列不炸写穿透）。
func mustJSON(v any) string {
	b, err := json.Marshal(v)
	if err != nil {
		log.Printf("flowbind: 投影行 JSON 列编码失败（零值列兜底）: %v", err)
		return "[]"
	}
	return string(b)
}

// persistAppend 写穿透落库（失败只注记 stderr——Store 非权威，投影行为不因落库失败翻转）。
func persistAppend(what string, err error) {
	if err != nil {
		log.Printf("flowbind: %s 写穿透落库失败（非权威，投影行为不受影响）: %v", what, err)
	}
}

// dispatchToRow 派单投影 → 行（受理即终态——同 seq 重放 noop）。
func dispatchToRow(d DispatchRecord) runtimestore.DispatchRow {
	return runtimestore.DispatchRow{
		Seq: d.Seq, Agent: d.Agent, Status: d.Status, AutoLevel: d.AutoLevel,
		Action: d.Action, Kind: d.Kind, Workflow: d.Workflow, Scope: d.Scope,
		Connector: d.Connector, Proposal: d.Proposal, Title: d.Title, Summary: d.Summary,
		AcceptanceJSON: mustJSON(d.Acceptance), Priority: d.Priority, Proposer: d.Proposer,
		RV: d.RV, EventID: d.EventID, NodesJSON: mustJSON(d.Nodes), At: stampNano(d.At),
	}
}

// dispatchFromRow 行 → 派单投影（恢复装载；JSON 列不可解析=零值清单——诚实降级）。
func dispatchFromRow(r runtimestore.DispatchRow) DispatchRecord {
	d := DispatchRecord{
		Seq: r.Seq, Agent: r.Agent, Status: r.Status, AutoLevel: r.AutoLevel,
		Action: r.Action, Kind: r.Kind, Workflow: r.Workflow, Scope: r.Scope,
		Connector: r.Connector, Proposal: r.Proposal, Title: r.Title, Summary: r.Summary,
		Priority: r.Priority, Proposer: r.Proposer, RV: r.RV, EventID: r.EventID,
		At: parseNano(r.At),
	}
	_ = json.Unmarshal([]byte(r.AcceptanceJSON), &d.Acceptance)
	_ = json.Unmarshal([]byte(r.NodesJSON), &d.Nodes)
	return d
}

// loadDispatchRows 全量装载派单行（游标分页；读失败=恢复降级不炸启动——启动续接缺口可见）。
func loadDispatchRows(s runtimestore.Store) []runtimestore.DispatchRow {
	var out []runtimestore.DispatchRow
	var since int64
	for {
		rows, err := s.ListDispatches(restorePage, since)
		if err != nil {
			log.Printf("flowbind: 派单投影恢复装载失败（恢复降级——启动续接缺口可见）: %v", err)
			return out
		}
		out = append(out, rows...)
		if len(rows) < restorePage {
			return out
		}
		since = rows[len(rows)-1].Cursor
	}
}
