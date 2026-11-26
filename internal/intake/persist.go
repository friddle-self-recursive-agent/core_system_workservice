// persist.go —— 录入会话投影写穿透/启动恢复（G-01 收口——R-ENG §1 消费侧第六类投影：
// intake SessionRecord ↔ runtimestore.IntakeSessionRow）。会话是**演化态**（同 id 多轮推进）：
// Append=upsert（同 id 覆写最新投影），恢复取最新词形。写穿透=内存照旧工作（ListSessions/
// GetSession 语义零变化，读仍走内存），每次会话变更同步 Append（锁内同步落库与投影同序，
// 落库失败只注记 stderr 不翻转投影行为——Store 是持久化面非权威）。启动恢复=挂接时装载
// 历史行（rt_intake_sessions 权威续接，memory 空启动=零行）。
//
// 装配注记：Engine 由 apiserver intake face 构造（路由域另一 agent 所有）——装配侧注入走
// 包级 SetDefaultStore（cored 于 apiserver.New 之前设定，NewEngine 拾取）；显式通道=
// (*Engine).AttachStore（face 侧/测试直挂并即时恢复装载）。
package intake

import (
	"encoding/json"
	"log"
	"sync"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/runtimestore"
)

// defaultStore 包级装配锚（cored 装配层于 face 构造前设定；nil=纯内存既有行为面零翻转）。
// mu 保护装配竞态（装配层单次设定，读侧 NewEngine 锁内快照）。
var (
	storeMu      sync.Mutex
	defaultStore runtimestore.Store
)

// SetDefaultStore 装配侧注入运行态投影存储（G-01——包级单点：apiserver intake face 在
// apiserver.New 内构造 Engine，cored 无法经构造参数注入，故以包级装配锚承接；须在
// apiserver.New 之前调用）。nil=关断（纯内存既有行为面零翻转）。
func SetDefaultStore(s runtimestore.Store) {
	storeMu.Lock()
	defer storeMu.Unlock()
	defaultStore = s
}

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

// sessionToRow 会话投影 → 行（演化态 upsert——同 id 覆写最新词形）。
func sessionToRow(s *SessionRecord) runtimestore.IntakeSessionRow {
	return runtimestore.IntakeSessionRow{
		SessionID: s.SessionID, Kind: s.Kind, Scope: s.Scope, Name: s.Name,
		Proposer: s.Proposer, Turns: s.Turns, GenVersion: s.GenVersion,
		AskingField: s.AskingField, Ready: s.Ready, LastDryRunOk: s.LastDryRunOk,
		FieldErrorCnt: s.FieldErrorCnt, ReasonRef: s.ReasonRef,
		FieldsJSON: mustJSONMap(s.Fields), SourcesJSON: mustJSONMapStr(s.Sources),
		CreatedAt: stampNano(s.CreatedAt), UpdatedAt: stampNano(s.UpdatedAt),
	}
}

// sessionFromRow 行 → 会话投影（恢复装载；JSON 列不可解析=零值快照——诚实降级）。
func sessionFromRow(r runtimestore.IntakeSessionRow) *SessionRecord {
	s := &SessionRecord{
		SessionID: r.SessionID, Kind: r.Kind, Scope: r.Scope, Name: r.Name,
		Proposer: r.Proposer, Turns: r.Turns, GenVersion: r.GenVersion,
		AskingField: r.AskingField, Ready: r.Ready, LastDryRunOk: r.LastDryRunOk,
		FieldErrorCnt: r.FieldErrorCnt, ReasonRef: r.ReasonRef,
		CreatedAt: parseNano(r.CreatedAt), UpdatedAt: parseNano(r.UpdatedAt),
		Fields:  map[string]any{},
		Sources: map[string]string{},
	}
	_ = json.Unmarshal([]byte(r.FieldsJSON), &s.Fields)
	_ = json.Unmarshal([]byte(r.SourcesJSON), &s.Sources)
	if s.Fields == nil {
		s.Fields = map[string]any{}
	}
	if s.Sources == nil {
		s.Sources = map[string]string{}
	}
	return s
}

// mustJSONMap 字段快照 → JSON 列（失败兜底零值列不炸写穿透）。
func mustJSONMap(m map[string]any) string {
	b, err := json.Marshal(m)
	if err != nil {
		log.Printf("intake: 字段快照 JSON 列编码失败（零值列兜底）: %v", err)
		return "{}"
	}
	return string(b)
}

// mustJSONMapStr 字段来源 → JSON 列（同上口径）。
func mustJSONMapStr(m map[string]string) string {
	b, err := json.Marshal(m)
	if err != nil {
		log.Printf("intake: 字段来源 JSON 列编码失败（零值列兜底）: %v", err)
		return "{}"
	}
	return string(b)
}

// persistSessionLocked 写穿透落库（调用方持 e.mu——与投影同序；失败只注记不翻转行为）。
func (e *Engine) persistSessionLocked(s *SessionRecord) {
	if e.store == nil || s == nil {
		return
	}
	if err := e.store.AppendIntakeSession(sessionToRow(s)); err != nil {
		log.Printf("intake: 会话 %s 写穿透落库失败（非权威，投影行为不受影响）: %v", s.SessionID, err)
	}
}

// AttachStore 显式挂接运行态投影存储并做启动恢复（G-01——face 侧/测试直挂通道）：
// 装载 rt_intake_sessions 历史行入内存映射（权威续接；挂接后写穿透随每次会话变更生效）。
func (e *Engine) AttachStore(s runtimestore.Store) {
	if s == nil {
		return
	}
	e.mu.Lock()
	defer e.mu.Unlock()
	e.store = s
	e.restoreLocked(s)
}

// restoreLocked 启动恢复装载（调用方持 e.mu；分页循环全量装载——读失败=恢复降级不炸启动，
// 启动续接缺口 stderr 可见）。
func (e *Engine) restoreLocked(s runtimestore.Store) {
	const page = 500
	var since int64
	n := 0
	for {
		rows, err := s.ListIntakeSessions(page, since)
		if err != nil {
			log.Printf("intake: 会话投影恢复装载失败（恢复降级——启动续接缺口可见）: %v", err)
			return
		}
		for _, r := range rows {
			e.sessions[r.SessionID] = sessionFromRow(r)
			n++
		}
		if len(rows) < page {
			break
		}
		since = rows[len(rows)-1].Cursor
	}
	if n > 0 {
		log.Printf("intake: 会话投影恢复装载 %d 条（rt_intake_sessions——重启续接）", n)
	}
}
