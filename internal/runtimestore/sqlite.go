// sqlite.go —— 运行态投影存储 SQLite WAL 实现（R-ENG 任务书 §1 sqlite 件）。
// 同仓口径（opslog.OpenSQLite 对齐）：WAL 日志 + busy_timeout 忙等待兜底；连接池=1
// （Core 单写者，concepts/core.md §7；内存库多连接会各自建库）；表前缀 rt_（runtime 投影
// 面，与 ops 账本表 event_log/apply_log 分域）；幂等建表迁移=R-ENG-2（CREATE TABLE IF
// NOT EXISTS，可重建可无损打开既有库）。
// 追加幂等：五类流水行=INSERT OR IGNORE（同 (kind,id) 重放 noop——事件可重语义的存储侧
// 投影）；会话=INSERT OR REPLACE（演化态 upsert，恢复取最新词形）。
package runtimestore

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	_ "modernc.org/sqlite" // DriverName=sqlite（opslog 同款驱动）
)

// DriverName sqlite 驱动名（诊断与测试可复用）。
const DriverName = "sqlite"

// SQLiteStore SQLite 运行态投影存储（单写者连接池）。
type SQLiteStore struct {
	db *sql.DB
}

// openSQLite 打开（必要时建表）runtime store。path 为空 → 进程内存库；文件路径父目录
// 自动创建（opslog 同款口径）。
func openSQLite(path string) (Store, error) {
	if path == "" {
		path = ":memory:"
	} else {
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
		if dir := filepath.Dir(path); dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("runtimestore: 建 DB 目录 %s: %w", dir, err)
			}
		}
	}
	dsn := path
	if path != ":memory:" {
		// WAL（ops-log.md §5 同款）+ 忙等待兜底（外部诊断并发读不阻塞写入方）。
		dsn = "file:" + path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	}
	db, err := sql.Open(DriverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("runtimestore: 打开 %s: %w", path, err)
	}
	db.SetMaxOpenConns(1) // Core 单写者（concepts/core.md §7）；内存库多连接会各自建库
	s := &SQLiteStore{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

// migrate 幂等建表（R-ENG-2——IF NOT EXISTS 可重入；既有库无损打开，六表六前缀 rt_）。
func (s *SQLiteStore) migrate() error {
	const ddl = `
CREATE TABLE IF NOT EXISTS rt_trigger_records (
	seq           INTEGER PRIMARY KEY,
	event_id      TEXT NOT NULL DEFAULT '',
	type          TEXT NOT NULL DEFAULT '',
	dedup_key     TEXT NOT NULL DEFAULT '',
	source        TEXT NOT NULL DEFAULT '',
	involved      TEXT NOT NULL DEFAULT '',
	level         TEXT NOT NULL DEFAULT '',
	at            TEXT NOT NULL DEFAULT '',
	received_at   TEXT NOT NULL DEFAULT '',
	route         TEXT NOT NULL DEFAULT '',
	cc            TEXT NOT NULL DEFAULT '',
	for_pred      TEXT NOT NULL DEFAULT '',
	owner         TEXT NOT NULL DEFAULT '',
	exec_json     TEXT NOT NULL DEFAULT '',
	cc_exec_json  TEXT NOT NULL DEFAULT '',
	scope         TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS rt_run_records (
	run              INTEGER PRIMARY KEY,
	workflow         TEXT NOT NULL DEFAULT '',
	spec_source      TEXT NOT NULL DEFAULT '',
	phases_json      TEXT NOT NULL DEFAULT '[]',
	transitions_json TEXT NOT NULL DEFAULT '[]',
	phase            TEXT NOT NULL DEFAULT '',
	status           TEXT NOT NULL DEFAULT '',
	brakes_json      TEXT NOT NULL DEFAULT '{}',
	trigger_seq      INTEGER NOT NULL DEFAULT 0,
	event_id         TEXT NOT NULL DEFAULT '',
	type             TEXT NOT NULL DEFAULT '',
	dedup_key        TEXT NOT NULL DEFAULT '',
	route            TEXT NOT NULL DEFAULT '',
	owner            TEXT NOT NULL DEFAULT '',
	target           TEXT NOT NULL DEFAULT '',
	involved         TEXT NOT NULL DEFAULT '',
	started_at       TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS rt_delivery_records (
	delivery     INTEGER PRIMARY KEY,
	channel      TEXT NOT NULL DEFAULT '',
	origin       TEXT NOT NULL DEFAULT '',
	status       TEXT NOT NULL DEFAULT '',
	trigger_seq  INTEGER NOT NULL DEFAULT 0,
	event_id     TEXT NOT NULL DEFAULT '',
	type         TEXT NOT NULL DEFAULT '',
	dedup_key    TEXT NOT NULL DEFAULT '',
	level        TEXT NOT NULL DEFAULT '',
	route        TEXT NOT NULL DEFAULT '',
	owner        TEXT NOT NULL DEFAULT '',
	target       TEXT NOT NULL DEFAULT '',
	at           TEXT NOT NULL DEFAULT '',
	attempts     INTEGER NOT NULL DEFAULT 0,
	completed_at TEXT NOT NULL DEFAULT '',
	error        TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS rt_wake_records (
	wake        INTEGER PRIMARY KEY,
	agent       TEXT NOT NULL DEFAULT '',
	origin      TEXT NOT NULL DEFAULT '',
	status      TEXT NOT NULL DEFAULT '',
	level       TEXT NOT NULL DEFAULT '',
	trigger_seq INTEGER NOT NULL DEFAULT 0,
	event_id    TEXT NOT NULL DEFAULT '',
	type        TEXT NOT NULL DEFAULT '',
	dedup_key   TEXT NOT NULL DEFAULT '',
	route       TEXT NOT NULL DEFAULT '',
	owner       TEXT NOT NULL DEFAULT '',
	at          TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS rt_dispatch_records (
	seq             INTEGER PRIMARY KEY,
	agent           TEXT NOT NULL DEFAULT '',
	status          TEXT NOT NULL DEFAULT '',
	auto_level      TEXT NOT NULL DEFAULT '',
	action          TEXT NOT NULL DEFAULT '',
	kind            TEXT NOT NULL DEFAULT '',
	workflow        TEXT NOT NULL DEFAULT '',
	scope           TEXT NOT NULL DEFAULT '',
	connector       TEXT NOT NULL DEFAULT '',
	proposal        TEXT NOT NULL DEFAULT '',
	title           TEXT NOT NULL DEFAULT '',
	summary         TEXT NOT NULL DEFAULT '',
	acceptance_json TEXT NOT NULL DEFAULT '[]',
	priority        TEXT NOT NULL DEFAULT '',
	proposer        TEXT NOT NULL DEFAULT '',
	rv              INTEGER NOT NULL DEFAULT 0,
	event_id        TEXT NOT NULL DEFAULT '',
	nodes_json      TEXT NOT NULL DEFAULT '[]',
	at              TEXT NOT NULL DEFAULT ''
);
CREATE TABLE IF NOT EXISTS rt_intake_sessions (
	session_id      TEXT PRIMARY KEY,
	kind            TEXT NOT NULL DEFAULT '',
	scope           TEXT NOT NULL DEFAULT '',
	name            TEXT NOT NULL DEFAULT '',
	proposer        TEXT NOT NULL DEFAULT '',
	turns           INTEGER NOT NULL DEFAULT 0,
	gen_version     INTEGER NOT NULL DEFAULT 0,
	asking_field    TEXT NOT NULL DEFAULT '',
	ready           INTEGER NOT NULL DEFAULT 0,
	last_dry_run_ok INTEGER NOT NULL DEFAULT 0,
	field_error_cnt INTEGER NOT NULL DEFAULT 0,
	reason_ref      TEXT NOT NULL DEFAULT '',
	fields_json     TEXT NOT NULL DEFAULT '{}',
	sources_json    TEXT NOT NULL DEFAULT '{}',
	created_at      TEXT NOT NULL DEFAULT '',
	updated_at      TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS rt_intake_sessions_updated ON rt_intake_sessions(updated_at);
`
	if _, err := s.db.Exec(ddl); err != nil {
		return fmt.Errorf("runtimestore: 建表: %w", err)
	}
	// 既有 DB 无损升列（opslog 同款口径——幂等：duplicate column 容忍；历史行缺省空=未回填/未审，
	// 读侧零值词形诚实投影不回填写面）：
	//   rt_run_records.session_id   —— G-08 loop 会话关联（唤醒桥受理回执回填）；
	//   rt_run_records.verify_json  —— G-09/G-12 Verify 硬 gate 判定投影（卡 Verify 证据随行）；
	//   rt_wake_records.session_id  —— G-08 loop 会话关联（/v1/loop/current 读面）。
	for _, col := range []struct{ table, name, ddl string }{
		{"rt_run_records", "session_id", `ALTER TABLE rt_run_records ADD COLUMN session_id TEXT NOT NULL DEFAULT ''`},
		{"rt_run_records", "verify_json", `ALTER TABLE rt_run_records ADD COLUMN verify_json TEXT NOT NULL DEFAULT ''`},
		{"rt_run_records", "spec_json", `ALTER TABLE rt_run_records ADD COLUMN spec_json TEXT NOT NULL DEFAULT ''`},
		{"rt_wake_records", "session_id", `ALTER TABLE rt_wake_records ADD COLUMN session_id TEXT NOT NULL DEFAULT ''`},
	} {
		if _, err := s.db.Exec(col.ddl); err != nil &&
			!strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
			return fmt.Errorf("runtimestore: %s 升列 %s: %w", col.table, col.name, err)
		}
	}
	return nil
}

// Close 关闭底层连接。
func (s *SQLiteStore) Close() error { return s.db.Close() }

// AppendTrigger 追加触发留痕（INSERT OR IGNORE——同 seq 重放 noop）。
func (s *SQLiteStore) AppendTrigger(r TriggerRow) error {
	const q = `INSERT OR IGNORE INTO rt_trigger_records
(seq, event_id, type, dedup_key, source, involved, level, at, received_at, route, cc, for_pred, owner, exec_json, cc_exec_json, scope)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	_, err := s.db.Exec(q, r.Seq, r.EventID, r.Type, r.DedupKey, r.Source, r.Involved, r.Level,
		r.At, r.ReceivedAt, r.Route, r.CC, r.ForPred, r.Owner, r.ExecJSON, r.CCExecJSON, r.Scope)
	if err != nil {
		return fmt.Errorf("runtimestore: rt_trigger_records 追加: %w", err)
	}
	return nil
}

// AppendRun 追加/更新对账 run（INSERT OR REPLACE——run 是**可演化投影**：受理落环后随
// 推进迁移/审批信号迁移/会话回填/Verify 判定就地翻转相位与状态，写穿透同 run id 覆写最新
// 词形，恢复取终态）。
func (s *SQLiteStore) AppendRun(r RunRow) error {
	const q = `INSERT OR REPLACE INTO rt_run_records
(run, workflow, spec_source, phases_json, transitions_json, phase, status, brakes_json, trigger_seq, event_id, type, dedup_key, route, owner, target, involved, started_at, session_id, verify_json, spec_json)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	_, err := s.db.Exec(q, r.Run, r.Workflow, r.SpecSource, r.PhasesJSON, r.TransitionsJSON,
		r.Phase, r.Status, r.BrakesJSON, r.TriggerSeq, r.EventID, r.Type, r.DedupKey,
		r.Route, r.Owner, r.Target, r.Involved, r.StartedAt, r.SessionID, r.VerifyJSON, r.SpecJSON)
	if err != nil {
		return fmt.Errorf("runtimestore: rt_run_records 追加: %w", err)
	}
	return nil
}

// AppendDelivery 追加/更新送达投影（INSERT OR REPLACE——送达是**可演化投影**：受理后随
// 出站起跑/终态翻转就地更新 attempts/status/completed_at，写穿透同 delivery id 覆写最新词形）。
func (s *SQLiteStore) AppendDelivery(r DeliveryRow) error {
	const q = `INSERT OR REPLACE INTO rt_delivery_records
(delivery, channel, origin, status, trigger_seq, event_id, type, dedup_key, level, route, owner, target, at, attempts, completed_at, error)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	_, err := s.db.Exec(q, r.Delivery, r.Channel, r.Origin, r.Status, r.TriggerSeq, r.EventID,
		r.Type, r.DedupKey, r.Level, r.Route, r.Owner, r.Target, r.At, r.Attempts,
		r.CompletedAt, r.Error)
	if err != nil {
		return fmt.Errorf("runtimestore: rt_delivery_records 追加: %w", err)
	}
	return nil
}

// AppendWake 追加唤醒投影（同 wake 重放 noop；会话锚回填走 upsert 语义——
// session_id 非空的覆写词形经 AppendWakeSession，受理词形仍 IGNORE 幂等）。
func (s *SQLiteStore) AppendWake(r WakeRow) error {
	const q = `INSERT OR IGNORE INTO rt_wake_records
(wake, agent, origin, status, level, trigger_seq, event_id, type, dedup_key, route, owner, at, session_id)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?)`
	_, err := s.db.Exec(q, r.Wake, r.Agent, r.Origin, r.Status, r.Level, r.TriggerSeq,
		r.EventID, r.Type, r.DedupKey, r.Route, r.Owner, r.At, r.SessionID)
	if err != nil {
		return fmt.Errorf("runtimestore: rt_wake_records 追加: %w", err)
	}
	return nil
}

// AppendWakeSession 回填唤醒会话锚（G-08 loop 会话关联——同 wake 行就地覆写 session_id；
// 受理词形已落、仅补会话锚的演化面。行缺席=幂等 noop（受理未落库时回填无锚可挂））。
func (s *SQLiteStore) AppendWakeSession(wake int64, sessionID string) error {
	const q = `UPDATE rt_wake_records SET session_id = ? WHERE wake = ?`
	if _, err := s.db.Exec(q, sessionID, wake); err != nil {
		return fmt.Errorf("runtimestore: rt_wake_records 回填 session_id: %w", err)
	}
	return nil
}

// AppendDispatch 追加派单投影（INSERT OR IGNORE——同 seq 重放 noop）。
func (s *SQLiteStore) AppendDispatch(r DispatchRow) error {
	const q = `INSERT OR IGNORE INTO rt_dispatch_records
(seq, agent, status, auto_level, action, kind, workflow, scope, connector, proposal, title, summary, acceptance_json, priority, proposer, rv, event_id, nodes_json, at)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	_, err := s.db.Exec(q, r.Seq, r.Agent, r.Status, r.AutoLevel, r.Action, r.Kind,
		r.Workflow, r.Scope, r.Connector, r.Proposal, r.Title, r.Summary, r.AcceptanceJSON,
		r.Priority, r.Proposer, r.RV, r.EventID, r.NodesJSON, r.At)
	if err != nil {
		return fmt.Errorf("runtimestore: rt_dispatch_records 追加: %w", err)
	}
	return nil
}

// AppendIntakeSession 追加/更新录入会话（INSERT OR REPLACE——演化态 upsert，同 id 覆写
// 最新投影；恢复取最新词形）。
func (s *SQLiteStore) AppendIntakeSession(r IntakeSessionRow) error {
	const q = `INSERT OR REPLACE INTO rt_intake_sessions
(session_id, kind, scope, name, proposer, turns, gen_version, asking_field, ready, last_dry_run_ok, field_error_cnt, reason_ref, fields_json, sources_json, created_at, updated_at)
VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`
	_, err := s.db.Exec(q, r.SessionID, r.Kind, r.Scope, r.Name, r.Proposer, r.Turns,
		r.GenVersion, r.AskingField, boolInt(r.Ready), boolInt(r.LastDryRunOk),
		r.FieldErrorCnt, r.ReasonRef, r.FieldsJSON, r.SourcesJSON, r.CreatedAt, r.UpdatedAt)
	if err != nil {
		return fmt.Errorf("runtimestore: rt_intake_sessions 追加: %w", err)
	}
	return nil
}

// ListTriggers 触发留痕行集（seq 升序；since=游标）。
func (s *SQLiteStore) ListTriggers(limit int, since int64) ([]TriggerRow, error) {
	rows, err := s.db.Query(`SELECT seq, event_id, type, dedup_key, source, involved, level, at, received_at, route, cc, for_pred, owner, exec_json, cc_exec_json, scope FROM rt_trigger_records WHERE seq > ? ORDER BY seq ASC LIMIT ?`, since, listLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("runtimestore: rt_trigger_records 列举: %w", err)
	}
	defer rows.Close()
	out := []TriggerRow{}
	for rows.Next() {
		var r TriggerRow
		if err := rows.Scan(&r.Seq, &r.EventID, &r.Type, &r.DedupKey, &r.Source, &r.Involved,
			&r.Level, &r.At, &r.ReceivedAt, &r.Route, &r.CC, &r.ForPred, &r.Owner,
			&r.ExecJSON, &r.CCExecJSON, &r.Scope); err != nil {
			return nil, fmt.Errorf("runtimestore: 读 rt_trigger_records 行: %w", err)
		}
		r.Cursor = r.Seq
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListRuns run 行集（run 升序；since=游标）。
func (s *SQLiteStore) ListRuns(limit int, since int64) ([]RunRow, error) {
	rows, err := s.db.Query(`SELECT run, workflow, spec_source, phases_json, transitions_json, phase, status, brakes_json, trigger_seq, event_id, type, dedup_key, route, owner, target, involved, started_at, session_id, verify_json, spec_json FROM rt_run_records WHERE run > ? ORDER BY run ASC LIMIT ?`, since, listLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("runtimestore: rt_run_records 列举: %w", err)
	}
	defer rows.Close()
	out := []RunRow{}
	for rows.Next() {
		var r RunRow
		if err := rows.Scan(&r.Run, &r.Workflow, &r.SpecSource, &r.PhasesJSON, &r.TransitionsJSON,
			&r.Phase, &r.Status, &r.BrakesJSON, &r.TriggerSeq, &r.EventID, &r.Type,
			&r.DedupKey, &r.Route, &r.Owner, &r.Target, &r.Involved, &r.StartedAt,
			&r.SessionID, &r.VerifyJSON, &r.SpecJSON); err != nil {
			return nil, fmt.Errorf("runtimestore: 读 rt_run_records 行: %w", err)
		}
		r.Cursor = r.Run
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListDeliveries 送达行集（delivery 升序；since=游标）。
func (s *SQLiteStore) ListDeliveries(limit int, since int64) ([]DeliveryRow, error) {
	rows, err := s.db.Query(`SELECT delivery, channel, origin, status, trigger_seq, event_id, type, dedup_key, level, route, owner, target, at, attempts, completed_at, error FROM rt_delivery_records WHERE delivery > ? ORDER BY delivery ASC LIMIT ?`, since, listLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("runtimestore: rt_delivery_records 列举: %w", err)
	}
	defer rows.Close()
	out := []DeliveryRow{}
	for rows.Next() {
		var r DeliveryRow
		if err := rows.Scan(&r.Delivery, &r.Channel, &r.Origin, &r.Status, &r.TriggerSeq,
			&r.EventID, &r.Type, &r.DedupKey, &r.Level, &r.Route, &r.Owner, &r.Target,
			&r.At, &r.Attempts, &r.CompletedAt, &r.Error); err != nil {
			return nil, fmt.Errorf("runtimestore: 读 rt_delivery_records 行: %w", err)
		}
		r.Cursor = r.Delivery
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListWakes 唤醒行集（wake 升序；since=游标）。
func (s *SQLiteStore) ListWakes(limit int, since int64) ([]WakeRow, error) {
	rows, err := s.db.Query(`SELECT wake, agent, origin, status, level, trigger_seq, event_id, type, dedup_key, route, owner, at, session_id FROM rt_wake_records WHERE wake > ? ORDER BY wake ASC LIMIT ?`, since, listLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("runtimestore: rt_wake_records 列举: %w", err)
	}
	defer rows.Close()
	out := []WakeRow{}
	for rows.Next() {
		var r WakeRow
		if err := rows.Scan(&r.Wake, &r.Agent, &r.Origin, &r.Status, &r.Level, &r.TriggerSeq,
			&r.EventID, &r.Type, &r.DedupKey, &r.Route, &r.Owner, &r.At, &r.SessionID); err != nil {
			return nil, fmt.Errorf("runtimestore: 读 rt_wake_records 行: %w", err)
		}
		r.Cursor = r.Wake
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListDispatches 派单行集（seq 升序；since=游标）。
func (s *SQLiteStore) ListDispatches(limit int, since int64) ([]DispatchRow, error) {
	rows, err := s.db.Query(`SELECT seq, agent, status, auto_level, action, kind, workflow, scope, connector, proposal, title, summary, acceptance_json, priority, proposer, rv, event_id, nodes_json, at FROM rt_dispatch_records WHERE seq > ? ORDER BY seq ASC LIMIT ?`, since, listLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("runtimestore: rt_dispatch_records 列举: %w", err)
	}
	defer rows.Close()
	out := []DispatchRow{}
	for rows.Next() {
		var r DispatchRow
		if err := rows.Scan(&r.Seq, &r.Agent, &r.Status, &r.AutoLevel, &r.Action, &r.Kind,
			&r.Workflow, &r.Scope, &r.Connector, &r.Proposal, &r.Title, &r.Summary,
			&r.AcceptanceJSON, &r.Priority, &r.Proposer, &r.RV, &r.EventID, &r.NodesJSON, &r.At); err != nil {
			return nil, fmt.Errorf("runtimestore: 读 rt_dispatch_records 行: %w", err)
		}
		r.Cursor = r.Seq
		out = append(out, r)
	}
	return out, rows.Err()
}

// ListIntakeSessions 会话行集（rowid 升序游标——行序不透明，消费方透传；updated_at 索引
// 供诊断）。upsert 同 id 覆写后 rowid 不变（REPLACE 语义=删旧插新，rowid 重排——游标
// 分页以本轮 List 内单调为准，恢复面用全量拉取不依赖跨轮 rowid 稳定性）。
func (s *SQLiteStore) ListIntakeSessions(limit int, since int64) ([]IntakeSessionRow, error) {
	rows, err := s.db.Query(`SELECT rowid, session_id, kind, scope, name, proposer, turns, gen_version, asking_field, ready, last_dry_run_ok, field_error_cnt, reason_ref, fields_json, sources_json, created_at, updated_at FROM rt_intake_sessions WHERE rowid > ? ORDER BY rowid ASC LIMIT ?`, since, listLimit(limit))
	if err != nil {
		return nil, fmt.Errorf("runtimestore: rt_intake_sessions 列举: %w", err)
	}
	defer rows.Close()
	out := []IntakeSessionRow{}
	for rows.Next() {
		var r IntakeSessionRow
		var ready, dryOk int
		if err := rows.Scan(&r.Cursor, &r.SessionID, &r.Kind, &r.Scope, &r.Name, &r.Proposer,
			&r.Turns, &r.GenVersion, &r.AskingField, &ready, &dryOk, &r.FieldErrorCnt,
			&r.ReasonRef, &r.FieldsJSON, &r.SourcesJSON, &r.CreatedAt, &r.UpdatedAt); err != nil {
			return nil, fmt.Errorf("runtimestore: 读 rt_intake_sessions 行: %w", err)
		}
		r.Ready, r.LastDryRunOk = ready == 1, dryOk == 1
		out = append(out, r)
	}
	return out, rows.Err()
}

// boolInt 布尔 → SQLite 整形（0/1）。
func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
