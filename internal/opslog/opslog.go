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

// EventCountsSince 窗口内 event_log 按类型行数（v0.33 MetricsInfo 数据面首段真实生产者的
// 查询投影落点——事件流水窗口计数即样本；不属于冻结 Store 接口，扩展走新方法）。
// 时间过滤在 Go 侧逐行解析（RFC3339Nano 去尾零致字面序≠时序，SQL 字符串比较有边界伪差
// ——量级与可重建口径下全量扫描过滤足够；since 零值=全量）。查询失败=调用方本拍零样本。
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
		if !since.IsZero() && t.Before(since) {
			continue
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
	Type     string // 精确匹配（空=不过滤）
	Level    string // 精确匹配（空=不过滤）
	Involved string // 精确匹配（空=不过滤）
	Since    time.Time // 时间下界含端点（零值=不限）
	Limit    int       // ≤0=缺省 50；>MaxEventLimit 收敛（读面防拖库——归档全量走重建/导出面）
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
