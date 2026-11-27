// persist.go —— 运行态投影写穿透/启动恢复的行转换与装载件（R-ENG 任务书 §1 消费侧——
// workflow 四类投影 TriggerRecord/RunRecord/DeliveryRecord/WakeRecord ↔ runtimestore 行）。
// 设计依据：写穿透=内存照旧工作（State()/查询语义零变化——读仍走内存），每次变更同步
// Append（sync 落库；落库失败只注记 stderr 不翻转投影行为——Store 是持久化面非权威，
// 内存投影面仍是权威）；启动恢复=New 时装载历史行续接序号/游标（sqlite 为权威续接，
// memory 空启动=零行）。行字段直投自内存结构：时间=RFC3339Nano 文本列，复杂字段
// （执行绑定/相位序列/迁移表/刹车值）=JSON 文本列。
package workflow

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
		log.Printf("workflow: 投影行 JSON 列编码失败（零值列兜底）: %v", err)
		return "{}"
	}
	return string(b)
}

// ───────────────────────── 行转换（字段直投，零语义改造） ─────────────────────────

// triggerToRow 触发留痕 → 行。
func triggerToRow(rec TriggerRecord) runtimestore.TriggerRow {
	return runtimestore.TriggerRow{
		Seq: rec.Seq, EventID: rec.EventID, Type: rec.Type, DedupKey: rec.DedupKey,
		Source: rec.Source, Involved: rec.Involved, Level: rec.Level,
		At: stampNano(rec.At), ReceivedAt: stampNano(rec.ReceivedAt),
		Route: rec.Route, CC: rec.CC, ForPred: rec.For, Owner: rec.Owner,
		ExecJSON: mustJSON(rec.Exec), CCExecJSON: mustJSON(rec.CCExec), Scope: rec.Scope,
	}
}

// triggerFromRow 行 → 触发留痕（恢复装载；JSON 列不可解析=零值绑定——诚实降级）。
func triggerFromRow(r runtimestore.TriggerRow) TriggerRecord {
	rec := TriggerRecord{
		Seq: r.Seq, EventID: r.EventID, Type: r.Type, DedupKey: r.DedupKey,
		Source: r.Source, Involved: r.Involved, Level: r.Level,
		At: parseNano(r.At), ReceivedAt: parseNano(r.ReceivedAt),
		Route: r.Route, CC: r.CC, For: r.ForPred, Owner: r.Owner, Scope: r.Scope,
	}
	_ = json.Unmarshal([]byte(r.ExecJSON), &rec.Exec)
	_ = json.Unmarshal([]byte(r.CCExecJSON), &rec.CCExec)
	return rec
}

// runToRow 对账 run → 行（可演化投影——upsert 语义，恢复取最新词形；会话锚/Verify 判定
// 随演化词形同行覆写——G-08/G-09 写穿透面）。spec 侧车列携 approvalMode+suspendWhy
// （2026-11-27 service session 裁决——挂起因由随恢复词形保留）。
func runToRow(run RunRecord) runtimestore.RunRow {
	return runtimestore.RunRow{
		Run: run.Run, Workflow: run.Workflow, SpecSource: run.SpecSource,
		PhasesJSON: mustJSON(run.Phases), TransitionsJSON: mustJSON(run.Transitions),
		Phase: run.Phase, Status: run.Status, BrakesJSON: mustJSON(run.Brakes),
		SpecJSON:    mustJSON(map[string]string{"approvalMode": run.ApprovalMode, "suspendWhy": run.SuspendWhy}),
		HistoryJSON: mustJSON(run.History), // F-04 运行轨迹写穿透（空轨迹=null 词形）
		TriggerSeq:  run.TriggerSeq, EventID: run.EventID, Type: run.Type,
		DedupKey: run.DedupKey, Route: run.Route, Owner: run.Owner, Target: run.Target,
		Involved: run.Involved, StartedAt: stampNano(run.StartedAt),
		SessionID: run.SessionID, VerifyJSON: verifyToJSON(run.Verify),
	}
}

// runFromRow 行 → 对账 run（恢复装载；Verify 列空=未审 nil——诚实降级不编造判定）。
func runFromRow(r runtimestore.RunRow) RunRecord {
	spec := specFromJSON(r.SpecJSON)
	run := RunRecord{
		Run: r.Run, Workflow: r.Workflow, SpecSource: r.SpecSource,
		Phase: r.Phase, Status: r.Status,
		ApprovalMode: spec["approvalMode"],
		SuspendWhy:   spec["suspendWhy"],
		TriggerSeq:   r.TriggerSeq, EventID: r.EventID, Type: r.Type,
		DedupKey: r.DedupKey, Route: r.Route, Owner: r.Owner, Target: r.Target,
		Involved: r.Involved, StartedAt: parseNano(r.StartedAt),
		SessionID: r.SessionID, Verify: verifyFromJSON(r.VerifyJSON),
	}
	_ = json.Unmarshal([]byte(r.PhasesJSON), &run.Phases)
	_ = json.Unmarshal([]byte(r.TransitionsJSON), &run.Transitions)
	_ = json.Unmarshal([]byte(r.BrakesJSON), &run.Brakes)
	_ = json.Unmarshal([]byte(r.HistoryJSON), &run.History) // F-04 轨迹恢复（空/损坏列=零轨迹——诚实降级）
	return run
}

// specFromJSON spec 侧车 JSON 列 → map（空列/损坏=空 map——诚实降级不冒充声明）。
func specFromJSON(s string) map[string]string {
	var m map[string]string
	if s == "" || json.Unmarshal([]byte(s), &m) != nil {
		return map[string]string{}
	}
	return m
}

// verifyToJSON Verify 判定 → JSON 文本列（nil=空列——未审词形）。
func verifyToJSON(v *VerifyVerdict) string {
	if v == nil {
		return ""
	}
	return mustJSON(*v)
}

// verifyFromJSON JSON 文本列 → Verify 判定（空/不可解析=nil——未审词形，诚实降级不编造）。
func verifyFromJSON(s string) *VerifyVerdict {
	if s == "" {
		return nil
	}
	var v VerifyVerdict
	if err := json.Unmarshal([]byte(s), &v); err != nil {
		return nil // 列不可解析=未审词形（诚实降级，不编造判定）
	}
	return &v
}

// deliveryToRow 送达投影 → 行（可演化投影——出站终态就地翻转，upsert 语义）。
func deliveryToRow(d DeliveryRecord) runtimestore.DeliveryRow {
	return runtimestore.DeliveryRow{
		Delivery: d.Delivery, Channel: d.Channel, Origin: d.Origin, Status: d.Status,
		TriggerSeq: d.TriggerSeq, EventID: d.EventID, Type: d.Type, DedupKey: d.DedupKey,
		Level: d.Level, Route: d.Route, Owner: d.Owner, Target: d.Target,
		At: stampNano(d.At), Attempts: d.Attempts, CompletedAt: stampNano(d.CompletedAt),
		Error: d.Error,
	}
}

// deliveryFromRow 行 → 送达投影（恢复装载）。
func deliveryFromRow(r runtimestore.DeliveryRow) DeliveryRecord {
	return DeliveryRecord{
		Delivery: r.Delivery, Channel: r.Channel, Origin: r.Origin, Status: r.Status,
		TriggerSeq: r.TriggerSeq, EventID: r.EventID, Type: r.Type, DedupKey: r.DedupKey,
		Level: r.Level, Route: r.Route, Owner: r.Owner, Target: r.Target,
		At: parseNano(r.At), Attempts: r.Attempts, CompletedAt: parseNano(r.CompletedAt),
		Error: r.Error,
	}
}

// wakeToRow 唤醒投影 → 行（受理即终态——INSERT OR IGNORE 幂等重放 noop；会话锚空=受理词形）。
func wakeToRow(w WakeRecord) runtimestore.WakeRow {
	return runtimestore.WakeRow{
		Wake: w.Wake, Agent: w.Agent, Origin: w.Origin, Status: w.Status, Level: w.Level,
		TriggerSeq: w.TriggerSeq, EventID: w.EventID, Type: w.Type, DedupKey: w.DedupKey,
		Route: w.Route, Owner: w.Owner, At: stampNano(w.At), SessionID: w.SessionID,
	}
}

// wakeFromRow 行 → 唤醒投影（恢复装载；session_id 列随行——G-08 loop 会话关联恢复面）。
func wakeFromRow(r runtimestore.WakeRow) WakeRecord {
	return WakeRecord{
		Wake: r.Wake, Agent: r.Agent, Origin: r.Origin, Status: r.Status, Level: r.Level,
		TriggerSeq: r.TriggerSeq, EventID: r.EventID, Type: r.Type, DedupKey: r.DedupKey,
		Route: r.Route, Owner: r.Owner, At: parseNano(r.At), SessionID: r.SessionID,
	}
}

// serviceSpawnToRow service 拉起投影 → 行（可演化投影——受理/回填/终态 upsert 同 seq 覆写）。
func serviceSpawnToRow(s ServiceSpawnRecord) runtimestore.ServiceSpawnRow {
	return runtimestore.ServiceSpawnRow{
		Seq: s.Seq, Run: s.Run, Workflow: s.Workflow, Node: s.Node, Phase: s.Phase,
		Service: s.Service, Op: s.Op, Agent: s.Agent, Role: s.Role,
		Context: mustJSON(s.Context), Status: s.Status, Session: s.Session, At: stampNano(s.At),
	}
}

// serviceSpawnFromRow 行 → service 拉起投影（恢复装载；context 列空/损坏=空字段——诚实降级）。
func serviceSpawnFromRow(r runtimestore.ServiceSpawnRow) ServiceSpawnRecord {
	rec := ServiceSpawnRecord{
		Seq: r.Seq, Run: r.Run, Workflow: r.Workflow, Node: r.Node, Phase: r.Phase,
		Service: r.Service, Op: r.Op, Agent: r.Agent, Role: r.Role,
		Status: r.Status, Session: r.Session, At: parseNano(r.At),
	}
	if r.Context != "" {
		_ = json.Unmarshal([]byte(r.Context), &rec.Context)
	}
	return rec
}

// ───────────────────────── 启动恢复装载（分页循环；读失败=恢复降级不炸启动） ─────────────────────────

// loadTriggerRows 全量装载触发行（游标分页）。
func loadTriggerRows(s runtimestore.Store) []runtimestore.TriggerRow {
	var out []runtimestore.TriggerRow
	var since int64
	for {
		rows, err := s.ListTriggers(restorePage, since)
		if err != nil {
			log.Printf("workflow: 触发留痕恢复装载失败（恢复降级——启动续接缺口可见）: %v", err)
			return out
		}
		out = append(out, rows...)
		if len(rows) < restorePage {
			return out
		}
		since = rows[len(rows)-1].Cursor
	}
}

// loadRunRows 全量装载 run 行。
func loadRunRows(s runtimestore.Store) []runtimestore.RunRow {
	var out []runtimestore.RunRow
	var since int64
	for {
		rows, err := s.ListRuns(restorePage, since)
		if err != nil {
			log.Printf("workflow: run 投影恢复装载失败（恢复降级——启动续接缺口可见）: %v", err)
			return out
		}
		out = append(out, rows...)
		if len(rows) < restorePage {
			return out
		}
		since = rows[len(rows)-1].Cursor
	}
}

// loadDeliveryRows 全量装载送达行。
func loadDeliveryRows(s runtimestore.Store) []runtimestore.DeliveryRow {
	var out []runtimestore.DeliveryRow
	var since int64
	for {
		rows, err := s.ListDeliveries(restorePage, since)
		if err != nil {
			log.Printf("workflow: 送达投影恢复装载失败（恢复降级——启动续接缺口可见）: %v", err)
			return out
		}
		out = append(out, rows...)
		if len(rows) < restorePage {
			return out
		}
		since = rows[len(rows)-1].Cursor
	}
}

// loadWakeRows 全量装载唤醒行。
func loadWakeRows(s runtimestore.Store) []runtimestore.WakeRow {
	var out []runtimestore.WakeRow
	var since int64
	for {
		rows, err := s.ListWakes(restorePage, since)
		if err != nil {
			log.Printf("workflow: 唤醒投影恢复装载失败（恢复降级——启动续接缺口可见）: %v", err)
			return out
		}
		out = append(out, rows...)
		if len(rows) < restorePage {
			return out
		}
		since = rows[len(rows)-1].Cursor
	}
}

// loadServiceSpawnRows 全量装载 service 拉起行（2026-11-27 service session 裁决恢复面）。
func loadServiceSpawnRows(s runtimestore.Store) []runtimestore.ServiceSpawnRow {
	var out []runtimestore.ServiceSpawnRow
	var since int64
	for {
		rows, err := s.ListServiceSpawns(restorePage, since)
		if err != nil {
			log.Printf("workflow: service 拉起投影恢复装载失败（恢复降级——启动续接缺口可见）: %v", err)
			return out
		}
		out = append(out, rows...)
		if len(rows) < restorePage {
			return out
		}
		since = rows[len(rows)-1].Cursor
	}
}

// persistAppend 写穿透落库（失败只注记 stderr——Store 非权威，投影行为不因落库失败翻转）。
func persistAppend(what string, err error) {
	if err != nil {
		log.Printf("workflow: %s 写穿透落库失败（非权威，投影行为不受影响）: %v", what, err)
	}
}
