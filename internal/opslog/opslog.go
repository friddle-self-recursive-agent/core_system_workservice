// Package opslog ops 历史操作日志（SQLite WAL 单文件；表见 instance 仓 runtime/ops-log.md §2）。
// git=文档内容版本（权威）；DB=操作流水（可丢弃可重建）。DB 非权威——这是它敢用 SQLite 的原因。
// v1 最小实现：apply_log / event_log 两表先行；control/approval/escalate 随各自概念落地后建表。
package opslog

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	_ "modernc.org/sqlite" // 纯 Go sqlite 驱动（无 cgo；Core 单写者串行，吞吐足够）
)

// ApplyLog apply_log 表（ops-log.md §2）：一次 Apply 判定的流水。
type ApplyLog struct {
	Ts           time.Time
	FieldManager string
	Subject      string // 判定主体标识（v0.9 请求级主体）：denied 行必带（user 优先，断言次之，匿名=anonymous）；
	//   ok/noop/conflict 行由 repository 双写落（Write 契约冻结不携主体），归属看 fieldManager
	Obj             string // "kind/scope/name"
	ResourceVersion int64
	Result          string // ok | noop | denied | conflict | recover（§3 恢复对账补记）
	PolicyDecisions string // 权限系统判定摘要（v0.9：判定方+域+动作，permission.md 规则 3 留痕）
	ReasonRef       string // 指向 DriftReport/Event/工单
}

// EventLog event_log 表：一次事件收发的归档行。
type EventLog struct {
	Ts       time.Time
	EventID  string
	Type     string
	Source   string
	Involved string
	DedupKey string
	Severity string // event.md：log 式等级（debug/info/warn/error/fatal）
	RoutedTo string // 命中的订阅方（空=死信）
}

// Store ops DB 接口。三个方法签名 v1 冻结（instance 仓 runtime/ops-log.md；扩展走新方法，不改签名）。
type Store interface {
	AppendApply(l ApplyLog) error
	AppendEvent(l EventLog) error
	// Reconcile 对账：返回差异计数。v1 = apply_log(ok) 与 event_log(DocApplied) 的库内不变量差
	// （每次成功 apply 恰好伴随一条 DocApplied 归档；noop 不发事件不落 ok 行）。
	// git 侧对账（apply_log.count ≈ git log count，ops-log.md §3）需要文档库句柄，
	// 由每日 verify 探针在外部执行——签名保持无参固定，不在本接口引入存储反向依赖。
	Reconcile() (diff int, err error)
}

// SQLiteStore ops DB 的 SQLite 实现（单写者：连接池收敛为 1，消灭 SQLITE_BUSY 面）。
type SQLiteStore struct {
	db *sql.DB
}

// 驱动名（测试与诊断可复用）。
const DriverName = "sqlite"

// OpenSQLite 打开（必要时建表）ops DB。path 为空 → 进程内存库（DB 可重建，进程生命周期即账本；
// 持久化部署显式传文件路径）。文件路径的父目录自动创建。
func OpenSQLite(path string) (*SQLiteStore, error) {
	if path == "" {
		path = ":memory:"
	} else {
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
		if dir := filepath.Dir(path); dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("opslog: 建 DB 目录 %s: %w", dir, err)
			}
		}
	}
	dsn := path
	if path != ":memory:" {
		// WAL（ops-log.md §5）+ 忙等待兜底（外部诊断工具并发读时不阻塞写入方）。
		dsn = "file:" + path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	}
	db, err := sql.Open(DriverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("opslog: 打开 %s: %w", path, err)
	}
	db.SetMaxOpenConns(1) // Core 单写者（concepts/core.md §7 串行队列）；内存库多连接会各自建库
	s := &SQLiteStore{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *SQLiteStore) migrate() error {
	const ddl = `
CREATE TABLE IF NOT EXISTS apply_log (
	ts               TEXT NOT NULL,
	field_manager    TEXT NOT NULL DEFAULT '',
	subject          TEXT NOT NULL DEFAULT '',
	obj              TEXT NOT NULL,
	resource_version INTEGER NOT NULL DEFAULT 0,
	result           TEXT NOT NULL,
	policy_decisions TEXT NOT NULL DEFAULT '',
	reason_ref       TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS apply_log_obj    ON apply_log(obj);
CREATE INDEX IF NOT EXISTS apply_log_result ON apply_log(result);
CREATE TABLE IF NOT EXISTS event_log (
	ts       TEXT NOT NULL,
	event_id TEXT NOT NULL,
	type     TEXT NOT NULL,
	source   TEXT NOT NULL DEFAULT '',
	involved TEXT NOT NULL DEFAULT '',
	dedup_key TEXT NOT NULL DEFAULT '',
	severity TEXT NOT NULL DEFAULT '',
	routed_to TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS event_log_type ON event_log(type);
`
	if _, err := s.db.Exec(ddl); err != nil {
		return fmt.Errorf("opslog: 建表: %w", err)
	}
	// 既有 DB 升级：apply_log 补 subject 列（DB 非权威可重建，但原库无损升列更省一次重建）。
	if _, err := s.db.Exec(`ALTER TABLE apply_log ADD COLUMN subject TEXT NOT NULL DEFAULT ''`); err != nil &&
		!strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return fmt.Errorf("opslog: apply_log 升列 subject: %w", err)
	}
	return nil
}

// Close 关闭底层连接。
func (s *SQLiteStore) Close() error { return s.db.Close() }

func stamp(t time.Time) string {
	if t.IsZero() {
		t = time.Now()
	}
	return t.UTC().Format(time.RFC3339Nano)
}

// AppendApply 落一条 apply 流水（repository 双写钩子与 apply 引擎拒绝路径共用）。
func (s *SQLiteStore) AppendApply(l ApplyLog) error {
	const q = `INSERT INTO apply_log (ts, field_manager, subject, obj, resource_version, result, policy_decisions, reason_ref)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	if l.Result == "" {
		return errors.New("opslog: apply_log.result 必填（ok|noop|denied|conflict|recover）")
	}
	if l.Obj == "" {
		return errors.New("opslog: apply_log.obj 必填（kind/scope/name）")
	}
	_, err := s.db.Exec(q, stamp(l.Ts), l.FieldManager, l.Subject, l.Obj, l.ResourceVersion, l.Result, l.PolicyDecisions, l.ReasonRef)
	if err != nil {
		return fmt.Errorf("opslog: 落 apply_log: %w", err)
	}
	return nil
}

// AppendEvent 归档一条事件（event.md：归档=审计素材；debug 默认不入库只进内存队列，见 Bus）。
func (s *SQLiteStore) AppendEvent(l EventLog) error {
	const q = `INSERT INTO event_log (ts, event_id, type, source, involved, dedup_key, severity, routed_to)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	if l.EventID == "" || l.Type == "" {
		return errors.New("opslog: event_log 需要 event_id 与 type")
	}
	_, err := s.db.Exec(q, stamp(l.Ts), l.EventID, l.Type, l.Source, l.Involved, l.DedupKey, l.Severity, l.RoutedTo)
	if err != nil {
		return fmt.Errorf("opslog: 落 event_log: %w", err)
	}
	return nil
}

// Reconcile 库内不变量对账：|apply_log(ok) − event_log(DocApplied)|。
// 0 = 每笔成功 apply 恰有一条 DocApplied 归档（apply.md：apply 必发 DocApplied）。
// git 侧对账（每日 verify 探针）在调用方比对 git log 后另行落 DriftReport。
func (s *SQLiteStore) Reconcile() (diff int, err error) {
	const q = `SELECT
  (SELECT COUNT(*) FROM apply_log WHERE result = 'ok') -
  (SELECT COUNT(*) FROM event_log WHERE type = 'DocApplied')`
	if err := s.db.QueryRow(q).Scan(&diff); err != nil {
		return 0, fmt.Errorf("opslog: 对账查询: %w", err)
	}
	if diff < 0 {
		diff = -diff
	}
	return diff, nil
}

// CountApply 按 result 统计 apply_log 行数（诊断与对账探针用；不属于冻结接口，具体实现方法）。
func (s *SQLiteStore) CountApply(result string) (int, error) {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM apply_log WHERE result = ?`, result).Scan(&n); err != nil {
		return 0, fmt.Errorf("opslog: 统计 apply_log(%s): %w", result, err)
	}
	return n, nil
}

// QueryApply 按 result 取 apply_log 行（按落库序；验收探针「查询代替日志」机械取证用，
// 如 denied 行主体归属断言；不属于冻结接口，具体实现方法）。
func (s *SQLiteStore) QueryApply(result string) ([]ApplyLog, error) {
	const q = `SELECT ts, field_manager, subject, obj, resource_version, result, policy_decisions, reason_ref
FROM apply_log WHERE result = ? ORDER BY rowid`
	rows, err := s.db.Query(q, result)
	if err != nil {
		return nil, fmt.Errorf("opslog: 查询 apply_log(%s): %w", result, err)
	}
	defer rows.Close()
	var out []ApplyLog
	for rows.Next() {
		var l ApplyLog
		var ts string
		if err := rows.Scan(&ts, &l.FieldManager, &l.Subject, &l.Obj, &l.ResourceVersion,
			&l.Result, &l.PolicyDecisions, &l.ReasonRef); err != nil {
			return nil, fmt.Errorf("opslog: 读 apply_log 行: %w", err)
		}
		if l.Ts, err = time.Parse(time.RFC3339Nano, ts); err != nil {
			return nil, fmt.Errorf("opslog: apply_log.ts 解析 %q: %w", ts, err)
		}
		out = append(out, l)
	}
	return out, rows.Err()
}

// CountEvent 按类型统计 event_log 行数（诊断用）。
func (s *SQLiteStore) CountEvent(eventType string) (int, error) {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM event_log WHERE type = ?`, eventType).Scan(&n); err != nil {
		return 0, fmt.Errorf("opslog: 统计 event_log(%s): %w", eventType, err)
	}
	return n, nil
}
