// Package opslog ops 历史操作日志（SQLite WAL 单文件；表见 instance 仓 runtime/ops-log.md §2）。
// git=文档内容版本（权威）；DB=操作流水（可丢弃可重建）。DB 非权威——这是它敢用 SQLite 的原因。
// v1 最小实现：apply_log / event_log / draft 先行；v0.46 approval_log 审批判定账面落地
// （instance 仓 runtime/ops-log.md §2 v0.46 注记）；control/escalate 随各自概念落地后建表。
package opslog

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
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

// 审批判定闭合集（ops-log.md §2 v0.46：decision 扩展走迁移行不重写语义）。
const (
	DecisionGranted  = "granted"
	DecisionRejected = "rejected"
)

// ApprovalLog approval_log 表（ops-log.md §2 v0.46）：一次审批判定的独立账行。
// 批准/驳回判定与 apply 落库流水（apply_log）双账并存不互斥：apply_log 回答"落了什么版本"，
// approval_log 回答"谁在何时对哪条提案做了什么判定"——驳回（弃稿零文档面动作）只有这里有行。
type ApprovalLog struct {
	Ts              time.Time
	ApprovalID      string // 判定流水键：现口径 "draft:<kind>/<scope>/<name>"（同键覆盖=最新判定）
	Obj             string // "kind/scope/name"（与 apply_log.obj 同构）
	Decision        string // granted | rejected（闭合集）
	JudgedBy        string // 判定主体（permission.md §3 请求级主体 ID；全空=anonymous）
	FieldManager    string // 判定来源面（console:approve / draft:delete；A44 落地后承判定方）
	ResourceVersion int64  // granted=落库后 rv（noop 幂等重入=面上现值）；rejected=0（零文档面动作）
	ReasonRef       string // 判定理由引用（console:approve / draft-delete）
}

// ApprovalQuery approval_log 查询参数（QueryApprovals 用；语义同 ApplyLogFace 读面口径）。
type ApprovalQuery struct {
	Decision string // 精确匹配（空=不过滤）
	Obj      string // "kind/scope/name" 精确匹配（空=不过滤）
	Limit    int    // ≤0=缺省 DefaultApplyLimit；>MaxApplyLimit 收敛（读面防拖库同口径）
}

// DraftRow draft 表一行（ops-log.md §2）：草稿缓存的持久化面（--ops-db 持久化部署）。
// 草稿不是文档面（无 Revision，git 权威不受碰）也不是操作流水——它是「待发布稿」缓存，
// 落 ops DB 是部署决策（可丢弃可重建：丢草稿=丢待确认编辑，可重放；不碰文档面权威性）。
// 本包不 import draft 包（两叶解耦）：draft 包定义同名语义的 Persistence 消费方接口，
// 字段映射由装配方（cmd/cored）完成。
type DraftRow struct {
	Kind         string
	Scope        string
	Name         string
	Content      []byte
	FieldManager string
	ReasonRef    string
	BaseRV       int64
	CreatedAt    time.Time
	UpdatedAt    time.Time
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
CREATE TABLE IF NOT EXISTS approval_log (
	ts               TEXT NOT NULL,
	approval_id      TEXT NOT NULL,
	obj              TEXT NOT NULL,
	decision         TEXT NOT NULL,
	judged_by        TEXT NOT NULL DEFAULT '',
	field_manager    TEXT NOT NULL DEFAULT '',
	resource_version INTEGER NOT NULL DEFAULT 0,
	reason_ref       TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS approval_log_decision ON approval_log(decision);
CREATE INDEX IF NOT EXISTS approval_log_obj      ON approval_log(obj);
CREATE TABLE IF NOT EXISTS draft (
	kind          TEXT NOT NULL,
	scope         TEXT NOT NULL,
	name          TEXT NOT NULL,
	content       TEXT NOT NULL,
	field_manager TEXT NOT NULL,
	reason_ref    TEXT NOT NULL DEFAULT '',
	base_rv       INTEGER NOT NULL DEFAULT 0,
	created_at    TEXT NOT NULL,
	updated_at    TEXT NOT NULL,
	PRIMARY KEY (kind, scope, name)
);
CREATE TABLE IF NOT EXISTS session_heartbeat (
	session_id    TEXT NOT NULL,
	ts            TEXT NOT NULL,
	phase         TEXT NOT NULL DEFAULT '',
	last_action   TEXT NOT NULL DEFAULT '',
	task_name     TEXT NOT NULL DEFAULT '',
	long_task     INTEGER NOT NULL DEFAULT 0,
	heal_attempts INTEGER NOT NULL DEFAULT 0,
	healed_at     TEXT NOT NULL DEFAULT '',
	updated_at    TEXT NOT NULL,
	PRIMARY KEY (session_id)
);
CREATE INDEX IF NOT EXISTS session_heartbeat_updated ON session_heartbeat(updated_at);
`
	if _, err := s.db.Exec(ddl); err != nil {
		return fmt.Errorf("opslog: 建表: %w", err)
	}
	// 既有 DB 升级：apply_log 补 subject 列（DB 非权威可重建，但原库无损升列更省一次重建）。
	if _, err := s.db.Exec(`ALTER TABLE apply_log ADD COLUMN subject TEXT NOT NULL DEFAULT ''`); err != nil &&
		!strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return fmt.Errorf("opslog: apply_log 升列 subject: %w", err)
	}
	// 既有 DB 升级：session_heartbeat 补 state 列（v0.54 KI-003——v0.54-rc 探针库已建无 state 版本表，
	// 升列后历史行缺省 live=合法词；DB 非权威，无损升列同款口径）。
	if _, err := s.db.Exec(`ALTER TABLE session_heartbeat ADD COLUMN state TEXT NOT NULL DEFAULT ''`); err != nil &&
		!strings.Contains(strings.ToLower(err.Error()), "duplicate column") {
		return fmt.Errorf("opslog: session_heartbeat 升列 state: %w", err)
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

// EventCountsSince 窗口内 event_log 按类型行数（v0.33 MetricsInfo 数据面首段真实生产者的
// 查询投影落点——事件流水窗口计数即样本；不属于冻结 Store 接口，扩展走新方法）。
// 时间过滤在 Go 侧逐行解析（RFC3339Nano 去尾零致字面序≠时序，SQL 字符串比较有边界伪差
// ——量级与可重建口径下全量扫描过滤足够；since 零值=全量）。窗口=半开 (since, now]
// （v0.45 对齐 E7 半开窗纪律：端点瞬刻归上一窗——相邻窗零重叠，首拍全量口径不变）。
// 查询失败=调用方本拍零样本。
func (s *SQLiteStore) EventCountsSince(since time.Time) (map[string]int, error) {
	rows, err := s.db.Query(`SELECT ts, type FROM event_log`)
	if err != nil {
		return nil, fmt.Errorf("opslog: 查询 event_log 窗口计数: %w", err)
	}
	defer rows.Close()
	out := map[string]int{}
	for rows.Next() {
		var ts, typ string
		if err := rows.Scan(&ts, &typ); err != nil {
			return nil, fmt.Errorf("opslog: 读 event_log 行: %w", err)
		}
		t, err := time.Parse(time.RFC3339Nano, ts)
		if err != nil {
			return nil, fmt.Errorf("opslog: event_log.ts 解析 %q: %w", ts, err)
		}
		if !since.IsZero() && !t.After(since) {
			continue // 半开窗 (since, now]：端点瞬刻归上一窗——相邻窗零重叠（E7 纪律同口径）
		}
		out[typ]++
	}
	return out, rows.Err()
}

// EventGroup event_log 归档行的 (type, involved) 分组计数（v0.37 console 读面第二段——
// workflow 运行管理聚合投影 `GET /v1/workflows` 的底层读取：每组的触发量与派发决议由
// 上层按组解析一次（dispatch 表快照纯函数），归档行不重复判定）。
type EventGroup struct {
	Type     string
	Involved string
	Count    int
}

// MaxGroupScanRows 分组查询的归档扫描行数上限（防拖库同口径；超出=结果带 capped 标记，
// 调用方按「扫描有界」如实投影——聚合计数不冒充全量）。
const MaxGroupScanRows = 20000

// GroupEvents event_log 分组计数（v0.37 console 读面第二段；不属于冻结 Store 接口，扩展走
// 新方法）。过滤语义同 QueryEvents（Type/Level/Involved 精确匹配、Since 时间下界）；
// 分组键=(type, involved)——计数按计数排序降序、同数按 (type, involved) 字典序（机械可复现）。
// scanned=实际扫描行数（过滤前）；capped=扫描触及上限提前收口（计数=有界窗，非全量承诺）。
func (s *SQLiteStore) GroupEvents(q EventQuery) (groups []EventGroup, scanned int64, capped bool, err error) {
	rows, err := s.db.Query(`SELECT ts, type, involved, severity FROM event_log`)
	if err != nil {
		return nil, 0, false, fmt.Errorf("opslog: 查询 event_log 分组: %w", err)
	}
	defer rows.Close()
	key := func(t, inv string) string { return t + "\x00" + inv }
	agg := map[string]*EventGroup{}
	for rows.Next() {
		if scanned >= MaxGroupScanRows {
			capped = true
			break
		}
		var ts, typ, inv, sev string
		if err := rows.Scan(&ts, &typ, &inv, &sev); err != nil {
			return nil, 0, false, fmt.Errorf("opslog: 读 event_log 行: %w", err)
		}
		scanned++
		t, err := time.Parse(time.RFC3339Nano, ts)
		if err != nil {
			return nil, 0, false, fmt.Errorf("opslog: event_log.ts 解析 %q: %w", ts, err)
		}
		if q.Type != "" && typ != q.Type {
			continue
		}
		if q.Level != "" && sev != q.Level {
			continue
		}
		if q.Involved != "" && inv != q.Involved {
			continue
		}
		if !q.Since.IsZero() && t.Before(q.Since) {
			continue
		}
		k := key(typ, inv)
		g, ok := agg[k]
		if !ok {
			g = &EventGroup{Type: typ, Involved: inv}
			agg[k] = g
		}
		g.Count++
	}
	if err := rows.Err(); err != nil {
		return nil, 0, false, fmt.Errorf("opslog: 遍历 event_log: %w", err)
	}
	groups = make([]EventGroup, 0, len(agg))
	for _, g := range agg {
		groups = append(groups, *g)
	}
	sort.Slice(groups, func(i, j int) bool {
		if groups[i].Count != groups[j].Count {
			return groups[i].Count > groups[j].Count
		}
		if groups[i].Type != groups[j].Type {
			return groups[i].Type < groups[j].Type
		}
		return groups[i].Involved < groups[j].Involved
	})
	return groups, scanned, capped, nil
}

// EventQuery event_log 读查询（v0.36 console 读面首段——event_log 归档行的查询投影，
// cored-bridge.md §2 `GET /v1/events` 的底层读取）。过滤字段全部精确匹配、空=不过滤；
// 时间过滤在 Go 侧逐行解析（同 EventCountsSince：RFC3339Nano 去尾零致字面序≠时序）。
type EventQuery struct {
	Type     string    // 精确匹配（空=不过滤）
	Level    string    // 精确匹配（空=不过滤）
	Involved string    // 精确匹配（空=不过滤）
	Since    time.Time // 时间下界含端点（零值=不限）
	Until    time.Time // 时间上界严格不含（零值=不限）——v0.39 E7：向后分页游标（上一页最旧行 ts，
	//  取 strictly-earlier 行实现翻页；与 Since 组合=半开窗 [since, until)）
	Limit int // ≤0=缺省 50；>MaxEventLimit 收敛（读面防拖库——归档全量走重建/导出面）
}

// MaxEventLimit 读面单次返回行数上限（防拖库；归档全量不属于查询面职责）。
const MaxEventLimit = 500

// DefaultEventLimit Limit 非正值时的缺省行数。
const DefaultEventLimit = 50

// apply_log 查询行数边界（v0.38）：与事件归档同值同口径（防拖库一表一策，不另设数）。
const (
	MaxApplyLimit     = MaxEventLimit
	DefaultApplyLimit = DefaultEventLimit
)

// QueryEvents event_log 归档行查询（最新优先：ts 降序，同刻按 rowid 降序——机械对账可复现）。
// 不属于冻结 Store 接口，扩展走新方法（ops-log.md §2：具体实现方法不占冻结签名）。
func (s *SQLiteStore) QueryEvents(q EventQuery) ([]EventLog, error) {
	const q0 = `SELECT ts, event_id, type, source, involved, dedup_key, severity, routed_to FROM event_log`
	rows, err := s.db.Query(q0)
	if err != nil {
		return nil, fmt.Errorf("opslog: 查询 event_log: %w", err)
	}
	defer rows.Close()
	limit := q.Limit
	if limit <= 0 {
		limit = DefaultEventLimit
	}
	if limit > MaxEventLimit {
		limit = MaxEventLimit
	}
	var all []EventLog
	for rows.Next() {
		var l EventLog
		var ts string
		if err := rows.Scan(&ts, &l.EventID, &l.Type, &l.Source, &l.Involved, &l.DedupKey, &l.Severity, &l.RoutedTo); err != nil {
			return nil, fmt.Errorf("opslog: 读 event_log 行: %w", err)
		}
		if l.Ts, err = time.Parse(time.RFC3339Nano, ts); err != nil {
			return nil, fmt.Errorf("opslog: event_log.ts 解析 %q: %w", ts, err)
		}
		if q.Type != "" && l.Type != q.Type {
			continue
		}
		if q.Level != "" && l.Severity != q.Level {
			continue
		}
		if q.Involved != "" && l.Involved != q.Involved {
			continue
		}
		if !q.Since.IsZero() && l.Ts.Before(q.Since) {
			continue
		}
		if !q.Until.IsZero() && !l.Ts.Before(q.Until) {
			continue
		}
		all = append(all, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("opslog: 遍历 event_log: %w", err)
	}
	// 最新优先（时间降序；同刻按归档序 rowid 降序——all 保持扫描序即 rowid 升序，
	// 稳定排序后末尾先出）。
	sort.SliceStable(all, func(i, j int) bool { return all[i].Ts.After(all[j].Ts) })
	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

// QueryApplies apply_log 行查询（v0.38 console 读面第三段：文档编辑器历史 tab 的权威数据源
// ——rev 流水=审计账不是 git 考古，ops-log.md §2 一次 Apply 判定的流水；不属于冻结 Store
// 接口，扩展走新方法）。obj="kind/scope/name" 精确匹配（空=全表——调用方不传即全量，
// 查询面恒带键，此为防御面）；排序=rv 降序、同 rv 按时刻降序（落库版本最新优先，机械
// 对账可复现：ok 行 rv=落库后版本）；limit≤0=缺省 50，>MaxApplyLimit 收敛=读面防拖库同口径。
func (s *SQLiteStore) QueryApplies(obj string, limit int) ([]ApplyLog, error) {
	const q0 = `SELECT ts, field_manager, subject, obj, resource_version, result, policy_decisions, reason_ref FROM apply_log`
	rows, err := s.db.Query(q0)
	if err != nil {
		return nil, fmt.Errorf("opslog: 查询 apply_log: %w", err)
	}
	defer rows.Close()
	if limit <= 0 {
		limit = DefaultApplyLimit
	}
	if limit > MaxApplyLimit {
		limit = MaxApplyLimit
	}
	var all []ApplyLog
	for rows.Next() {
		var l ApplyLog
		var ts string
		if err := rows.Scan(&ts, &l.FieldManager, &l.Subject, &l.Obj, &l.ResourceVersion, &l.Result, &l.PolicyDecisions, &l.ReasonRef); err != nil {
			return nil, fmt.Errorf("opslog: 读 apply_log 行: %w", err)
		}
		if l.Ts, err = time.Parse(time.RFC3339Nano, ts); err != nil {
			return nil, fmt.Errorf("opslog: apply_log.ts 解析 %q: %w", ts, err)
		}
		if obj != "" && l.Obj != obj {
			continue
		}
		all = append(all, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("opslog: 遍历 apply_log: %w", err)
	}
	sort.SliceStable(all, func(i, j int) bool {
		if all[i].ResourceVersion != all[j].ResourceVersion {
			return all[i].ResourceVersion > all[j].ResourceVersion
		}
		return all[i].Ts.After(all[j].Ts)
	})
	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

// UpsertDraft 写穿一条草稿（draft 表；同键覆盖=最新代改稿，CreatedAt 由调用方定格）。
// 不属于冻结 Store 接口——草稿持久化是部署决策（--ops-db 持久化面），具体实现方法。
func (s *SQLiteStore) UpsertDraft(r DraftRow) error {
	if r.Kind == "" || r.Scope == "" || r.Name == "" {
		return errors.New("opslog: draft 行需 kind、scope、name")
	}
	if r.FieldManager == "" || len(r.Content) == 0 {
		return errors.New("opslog: draft 行需 field_manager 与非空 content")
	}
	const q = `INSERT INTO draft (kind, scope, name, content, field_manager, reason_ref, base_rv, created_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT (kind, scope, name) DO UPDATE SET
	content=excluded.content, field_manager=excluded.field_manager, reason_ref=excluded.reason_ref,
	base_rv=excluded.base_rv, updated_at=excluded.updated_at` // created_at 不动——首次落稿定格
	_, err := s.db.Exec(q, r.Kind, r.Scope, r.Name, string(r.Content), r.FieldManager, r.ReasonRef,
		r.BaseRV, stamp(r.CreatedAt), stamp(r.UpdatedAt))
	if err != nil {
		return fmt.Errorf("opslog: 落 draft(%s/%s/%s): %w", r.Kind, r.Scope, r.Name, err)
	}
	return nil
}

// DeleteDraft 删一条草稿；键不存在=幂等成功（0 行受影响不算错——缓存/存储短暂不一致时写穿不放大）。
func (s *SQLiteStore) DeleteDraft(kind, scope, name string) error {
	if _, err := s.db.Exec(`DELETE FROM draft WHERE kind = ? AND scope = ? AND name = ?`, kind, scope, name); err != nil {
		return fmt.Errorf("opslog: 删 draft(%s/%s/%s): %w", kind, scope, name, err)
	}
	return nil
}

// ListDrafts 全量草稿行（按对象键排序——确定性恢复，机械对账可复现）。
func (s *SQLiteStore) ListDrafts() ([]DraftRow, error) {
	const q = `SELECT kind, scope, name, content, field_manager, reason_ref, base_rv, created_at, updated_at
FROM draft ORDER BY kind, scope, name`
	rows, err := s.db.Query(q)
	if err != nil {
		return nil, fmt.Errorf("opslog: 查询 draft: %w", err)
	}
	defer rows.Close()
	var out []DraftRow
	for rows.Next() {
		var r DraftRow
		var content, created, updated string
		if err := rows.Scan(&r.Kind, &r.Scope, &r.Name, &content, &r.FieldManager,
			&r.ReasonRef, &r.BaseRV, &created, &updated); err != nil {
			return nil, fmt.Errorf("opslog: 读 draft 行: %w", err)
		}
		r.Content = []byte(content)
		if r.CreatedAt, err = time.Parse(time.RFC3339Nano, created); err != nil {
			return nil, fmt.Errorf("opslog: draft.created_at 解析 %q: %w", created, err)
		}
		if r.UpdatedAt, err = time.Parse(time.RFC3339Nano, updated); err != nil {
			return nil, fmt.Errorf("opslog: draft.updated_at 解析 %q: %w", updated, err)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// CountDraft 草稿行数（诊断与恢复探针用）。
func (s *SQLiteStore) CountDraft() (int, error) {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM draft`).Scan(&n); err != nil {
		return 0, fmt.Errorf("opslog: 统计 draft: %w", err)
	}
	return n, nil
}

// AppendApproval 落一条审批判定（v0.46 审批独立账：apiserver 批准/驳回写钩子共用）。
// decision 闭合集校验（granted|rejected）；approval_id/obj 必填。账面非权威——调用方
// （apiserver 写钩子）落账失败不翻转写路径结果，stderr 注记即可（DB 可重建）。
func (s *SQLiteStore) AppendApproval(l ApprovalLog) error {
	const q = `INSERT INTO approval_log (ts, approval_id, obj, decision, judged_by, field_manager, resource_version, reason_ref)
VALUES (?, ?, ?, ?, ?, ?, ?, ?)`
	if l.Decision != DecisionGranted && l.Decision != DecisionRejected {
		return fmt.Errorf("opslog: approval_log.decision 闭合集外 %q（granted|rejected）", l.Decision)
	}
	if l.ApprovalID == "" {
		return errors.New("opslog: approval_log.approval_id 必填（判定流水键）")
	}
	if l.Obj == "" {
		return errors.New("opslog: approval_log.obj 必填（kind/scope/name）")
	}
	_, err := s.db.Exec(q, stamp(l.Ts), l.ApprovalID, l.Obj, l.Decision, l.JudgedBy, l.FieldManager, l.ResourceVersion, l.ReasonRef)
	if err != nil {
		return fmt.Errorf("opslog: 落 approval_log: %w", err)
	}
	return nil
}

// QueryApprovals approval_log 行查询（v0.46 读面 GET /v1/approvals 底层；不属于冻结 Store
// 接口，扩展走新方法）。decision/obj 精确匹配（空=不过滤）；排序=rv 降序、同 rv 按时刻降序
// （granted 行 rv=落库后版本最新优先，机械对账可复现；rejected 行 rv=0 沉底但时刻仍可比）；
// limit 同口径收敛（≤0=缺省，>MaxApplyLimit 收敛=读面防拖库）。
func (s *SQLiteStore) QueryApprovals(q ApprovalQuery) ([]ApprovalLog, error) {
	if q.Decision != "" && q.Decision != DecisionGranted && q.Decision != DecisionRejected {
		return nil, fmt.Errorf("opslog: approval_log.decision 闭合集外 %q（granted|rejected）", q.Decision)
	}
	const q0 = `SELECT ts, approval_id, obj, decision, judged_by, field_manager, resource_version, reason_ref FROM approval_log`
	rows, err := s.db.Query(q0)
	if err != nil {
		return nil, fmt.Errorf("opslog: 查询 approval_log: %w", err)
	}
	defer rows.Close()
	limit := q.Limit
	if limit <= 0 {
		limit = DefaultApplyLimit
	}
	if limit > MaxApplyLimit {
		limit = MaxApplyLimit
	}
	var all []ApprovalLog
	for rows.Next() {
		var l ApprovalLog
		var ts string
		if err := rows.Scan(&ts, &l.ApprovalID, &l.Obj, &l.Decision, &l.JudgedBy, &l.FieldManager, &l.ResourceVersion, &l.ReasonRef); err != nil {
			return nil, fmt.Errorf("opslog: 读 approval_log 行: %w", err)
		}
		if l.Ts, err = time.Parse(time.RFC3339Nano, ts); err != nil {
			return nil, fmt.Errorf("opslog: approval_log.ts 解析 %q: %w", ts, err)
		}
		if q.Decision != "" && l.Decision != q.Decision {
			continue
		}
		if q.Obj != "" && l.Obj != q.Obj {
			continue
		}
		all = append(all, l)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("opslog: 遍历 approval_log: %w", err)
	}
	sort.SliceStable(all, func(i, j int) bool {
		if all[i].ResourceVersion != all[j].ResourceVersion {
			return all[i].ResourceVersion > all[j].ResourceVersion
		}
		return all[i].Ts.After(all[j].Ts)
	})
	if len(all) > limit {
		all = all[:limit]
	}
	return all, nil
}

// CountApprovals 按判定统计 approval_log 行数（诊断与验收探针「查询代替日志」机械取证用）。
func (s *SQLiteStore) CountApprovals(decision string) (int, error) {
	if decision != "" && decision != DecisionGranted && decision != DecisionRejected {
		return 0, fmt.Errorf("opslog: approval_log.decision 闭合集外 %q（granted|rejected）", decision)
	}
	var n int
	q := `SELECT COUNT(*) FROM approval_log`
	var arg any
	if decision != "" {
		q += ` WHERE decision = ?`
		arg = decision
	}
	if err := s.db.QueryRow(q, arg).Scan(&n); err != nil {
		return 0, fmt.Errorf("opslog: 统计 approval_log(%s): %w", decision, err)
	}
	return n, nil
}

// SessionHeartbeat session_heartbeat 表一行（ops-log.md §2 KI-003 心跳账面）：会话活性租约
// 的 ops 侧权威投影——心跳即 upsert（进程内计数在 sessionliveness 包，DB 行=可查询面）。
// 状态词闭合集：state=live（现势）/suspected-stall（判定态词=SuspectedStall，event.md 规则 14b
// ——判定词≠事件词）/long-task（长任务显式化）/healed（自愈恢复留痕）/escalated（自愈无效升级）。
type SessionHeartbeat struct {
	SessionID    string    // 会话锚（involved 投影 Session/<id> 的 id 段）
	Ts           time.Time // 本次推进时刻（=updated_at 同值）
	Phase        string    // 相位/推进点词汇（phase-advance|tool-call|artifact-write|inject 等）
	LastAction   string    // 最后动作摘要（有界；不携带凭据/内网信息——R-PUB-1 事件载荷同红线）
	TaskName     string    // 长任务名（LongTaskRunning 判据字段；空=非长任务）
	LongTask     bool      // 声明为长任务（true=watchdog 不判停滞只显式汇报）
	HealAttempts int       // 累计自愈注入次数（连续无效达阈值=升级）
	HealedAt     time.Time // 最近一次 SessionHealed 时刻（零值=未恢复过）
	State        string    // live|suspected-stall|long-task|healed|escalated
	UpdatedAt    time.Time // 行级更新时刻（watchdog 扫描键=MAX(updated_at)）
}

// 心跳状态词闭合集（sessionliveness/watchdog 与读面共用；词汇=agent-session.md §7）。
const (
	SessionStateLive           = "live"
	SessionStateSuspectedStall = "suspected-stall" // 判定态词=SuspectedStall（下划线形=DB 词形）
	SessionStateLongTask       = "long-task"
	SessionStateHealed         = "healed"
	SessionStateEscalated      = "escalated"
)

// UpsertSessionHeartbeat 落/刷一行心跳（心跳即写=活性租约续租；SQLite 单写者串行）。
func (s *SQLiteStore) UpsertSessionHeartbeat(h SessionHeartbeat) error {
	if h.SessionID == "" {
		return errors.New("opslog: session_heartbeat.session_id 必填（会话锚）")
	}
	if h.State == "" {
		h.State = SessionStateLive
	}
	if h.Ts.IsZero() {
		h.Ts = time.Now()
	}
	if h.UpdatedAt.IsZero() {
		h.UpdatedAt = h.Ts
	}
	const q = `INSERT INTO session_heartbeat
	(session_id, ts, phase, last_action, task_name, long_task, heal_attempts, healed_at, state, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(session_id) DO UPDATE SET
	ts=excluded.ts, phase=excluded.phase, last_action=excluded.last_action,
	task_name=excluded.task_name, long_task=excluded.long_task,
	heal_attempts=excluded.heal_attempts, healed_at=excluded.healed_at,
	state=excluded.state, updated_at=excluded.updated_at`
	_, err := s.db.Exec(q, h.SessionID, stamp(h.Ts), h.Phase, h.LastAction, h.TaskName,
		boolInt(h.LongTask), h.HealAttempts, stamp(h.HealedAt), h.State, stamp(h.UpdatedAt))
	if err != nil {
		return fmt.Errorf("opslog: 落 session_heartbeat(%s): %w", h.SessionID, err)
	}
	return nil
}

// StaleSessionHeartbeats updated_at 早于 before 的心跳行（watchdog 扫描投影；按 updated_at
// 升序=最旧停滞优先；limit 收敛防拖库）。
func (s *SQLiteStore) StaleSessionHeartbeats(before time.Time, limit int) ([]SessionHeartbeat, error) {
	if limit <= 0 {
		limit = DefaultApplyLimit
	}
	if limit > MaxApplyLimit {
		limit = MaxApplyLimit
	}
	rows, err := s.db.Query(`SELECT session_id, ts, phase, last_action, task_name, long_task,
heal_attempts, healed_at, state, updated_at FROM session_heartbeat
WHERE updated_at < ? ORDER BY updated_at ASC LIMIT ?`, stamp(before), limit)
	if err != nil {
		return nil, fmt.Errorf("opslog: 查询停滞心跳行: %w", err)
	}
	defer rows.Close()
	return scanSessionHeartbeats(rows)
}

// AllSessionHeartbeats 心跳行全量投影（GET /v1/sessions 读面底层；updated_at 降序最新优先）。
func (s *SQLiteStore) AllSessionHeartbeats(limit int) ([]SessionHeartbeat, error) {
	if limit <= 0 {
		limit = DefaultApplyLimit
	}
	if limit > MaxApplyLimit {
		limit = MaxApplyLimit
	}
	rows, err := s.db.Query(`SELECT session_id, ts, phase, last_action, task_name, long_task,
heal_attempts, healed_at, state, updated_at FROM session_heartbeat
ORDER BY updated_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("opslog: 查询心跳行: %w", err)
	}
	defer rows.Close()
	return scanSessionHeartbeats(rows)
}

func scanSessionHeartbeats(rows *sql.Rows) ([]SessionHeartbeat, error) {
	var out []SessionHeartbeat
	for rows.Next() {
		var h SessionHeartbeat
		var ts, healedAt, updatedAt string
		var longTask int
		if err := rows.Scan(&h.SessionID, &ts, &h.Phase, &h.LastAction, &h.TaskName, &longTask,
			&h.HealAttempts, &healedAt, &h.State, &updatedAt); err != nil {
			return nil, fmt.Errorf("opslog: 读 session_heartbeat 行: %w", err)
		}
		var err error
		if h.Ts, err = time.Parse(time.RFC3339Nano, ts); err != nil {
			return nil, fmt.Errorf("opslog: session_heartbeat.ts 解析 %q: %w", ts, err)
		}
		if h.HealedAt, err = time.Parse(time.RFC3339Nano, healedAt); err != nil && healedAt != "" {
			return nil, fmt.Errorf("opslog: session_heartbeat.healed_at 解析 %q: %w", healedAt, err)
		}
		if h.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt); err != nil {
			return nil, fmt.Errorf("opslog: session_heartbeat.updated_at 解析 %q: %w", updatedAt, err)
		}
		h.LongTask = longTask != 0
		out = append(out, h)
	}
	return out, rows.Err()
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
