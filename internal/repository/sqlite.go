// Package repository —— 文档库 sqlite 后端（--docs-backend sqlite；用户 2026-11-23 指示：
// 文档 storage 走 sqlite，Go sqlite 驱动）。设计依据：core_design/concepts/core.md
// （Repository 接口可换后端）、core_design/core/scheme.yaml runtime.storage.docs 口径
// （git 布局=导入导出格式）、instance 仓 runtime/ops-log.md（双轨分工收口段）。
//
// 取舍（本后端与 git 后端的关系，声明面见 scheme.yaml runtime.storage.docs）：
//   - 并存、单权威：部署选择其一为权威（--docs-backend git|sqlite，缺省 sqlite=root 默认；
//     git=兼容后端，非 root 部署可显式选择）；两后端同实现 Repository 接口（读/写/watch 语义一致）。
//   - 权威=内容面：sqlite 后端下 doc 表即文档内容权威（resourceVersion=乐观锁，
//     kind/scope/name=对象键）；git 布局降级为「导入导出格式」。
//   - 迁移路径：`--docs-backend sqlite --docs-db <path> --docs <旧git布局根>` 首次打开
//     空库时一次性种子导入（SeedFromLayout，保留各对象 resourceVersion；非空库不导入=幂等）。
//   - 写语义零迁移：串行写队列 + 乐观锁 + 幂等重入 noop + apply_log 双写钩子，与 git 后端
//     同一套校验（kind 注册闸门/scope 四层/命名规则/身份一致/canonical 比对/凭据来源闸门）。
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"sync"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"

	_ "modernc.org/sqlite" // 纯 Go sqlite 驱动（与 ops DB 同驱动；cored 单写者串行）
)

// SQLiteRepo 文档库 sqlite 后端。写路径全部经单 goroutine 串行队列（与 GitRepo 同范式）；
// 连接池收敛为 1（单写者，SQLITE_BUSY 面与 ops DB 同口径消灭）。
type SQLiteRepo struct {
	db   *sql.DB
	sch  *scheme.Scheme
	opts Options

	reqs chan *writeReq
	stop chan struct{}
	once sync.Once
	done chan struct{}

	mu   sync.Mutex
	subs map[chan string]struct{} // Watch 订阅扇出（进程内写直通知——无外部进程可改库）
}

// OpenSQLite 打开（必要时建表）文档库 sqlite 后端。dbPath 为空 → 进程内存库（仅探针/测试；
// 持久化部署显式传文件路径——docs DB=权威内容面，不可丢弃）。seedRoot 非空且库为空 →
// 一次性种子导入（git 布局 → doc 表，保留 resourceVersion）。
func OpenSQLite(dbPath string, sch *scheme.Scheme, opts *Options, seedRoot string) (*SQLiteRepo, error) {
	if sch == nil {
		return nil, errors.New("repository: 需 scheme 注册表（野对象闸门）")
	}
	path := dbPath
	if path == "" {
		path = ":memory:"
	} else if abs, err := filepath.Abs(path); err == nil {
		path = abs
	}
	dsn := path
	if path != ":memory:" {
		if dir := filepath.Dir(path); dir != "" {
			if err := os.MkdirAll(dir, 0o755); err != nil {
				return nil, fmt.Errorf("repository: 建 docs DB 目录 %s: %w", dir, err)
			}
		}
		dsn = "file:" + path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	}
	db, err := sql.Open(sqliteDriver, dsn)
	if err != nil {
		return nil, fmt.Errorf("repository: 打开 docs DB %s: %w", path, err)
	}
	db.SetMaxOpenConns(1) // 单写者（concepts/core.md §7 串行队列；与 ops DB 同口径）
	if path != ":memory:" {
		if _, err := db.Exec(`PRAGMA journal_mode=WAL; PRAGMA busy_timeout=5000;`); err != nil {
			_ = db.Close()
			return nil, fmt.Errorf("repository: docs DB pragma: %w", err)
		}
	}
	r := &SQLiteRepo{
		db:   db,
		sch:  sch,
		reqs: make(chan *writeReq),
		stop: make(chan struct{}),
		done: make(chan struct{}),
		subs: map[chan string]struct{}{},
	}
	if opts != nil {
		r.opts = *opts
	}
	if err := r.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	if seedRoot != "" {
		n, err := r.SeedFromLayout(seedRoot)
		if err != nil {
			_ = db.Close()
			return nil, err
		}
		if n > 0 {
			fmt.Fprintf(os.Stderr, "repository: docs DB 种子导入 %d 对象（来自 %s；非空库不导入=幂等）\n", n, seedRoot)
		}
	}
	go r.processWrites()
	return r, nil
}

// sqliteDriver 驱动名（modernc.org/sqlite 注册名；与 opslog 包同驱动同进程注册一次）。
const sqliteDriver = "sqlite"

func (r *SQLiteRepo) migrate() error {
	const ddl = `
CREATE TABLE IF NOT EXISTS doc (
	kind             TEXT NOT NULL,
	scope            TEXT NOT NULL,
	name             TEXT NOT NULL,
	content          TEXT NOT NULL,
	resource_version INTEGER NOT NULL DEFAULT 0,
	updated_at       TEXT NOT NULL,
	PRIMARY KEY (kind, scope, name)
);
CREATE INDEX IF NOT EXISTS doc_kind_scope ON doc(kind, scope);`
	if _, err := r.db.Exec(ddl); err != nil {
		return fmt.Errorf("repository: docs DB 建表: %w", err)
	}
	return nil
}

// Close 关停写队列（幂等：排空在途请求后处理器退出）并关闭连接。
func (r *SQLiteRepo) Close() error {
	r.once.Do(func() { close(r.stop) })
	<-r.done
	return r.db.Close()
}

func (r *SQLiteRepo) processWrites() {
	defer close(r.done)
	drain := func() (empty bool) {
		select {
		case req := <-r.reqs:
			r.serve(req)
			return false
		default:
			return true
		}
	}
	for {
		select {
		case req := <-r.reqs:
			r.serve(req)
		case <-r.stop:
			for !drain() {
			}
			return
		}
	}
}

func (r *SQLiteRepo) serve(req *writeReq) {
	rv, action, err := r.applyWrite(req)
	select {
	case req.resp <- writeResp{rv: rv, action: action, err: err}:
	case <-req.ctx.Done():
	}
}

// Write 提交写入（并发安全：单 goroutine 串行队列；语义与 GitRepo.Write 一致，
// concepts/revision.md——fm 为空拒绝、乐观锁、幂等重入 noop、盲写拒绝）。
func (r *SQLiteRepo) Write(ctx context.Context, o *Object, fieldManager, reasonRef string) (int64, string, error) {
	if fieldManager == "" {
		return 0, "", ErrNoFieldManager
	}
	if o == nil || o.Content == nil {
		return 0, "", fmt.Errorf("%w: 空内容", ErrContent)
	}
	req := &writeReq{
		o: &Object{Kind: o.Kind, Name: o.Name, Scope: o.Scope, ResourceVersion: o.ResourceVersion,
			Content: append([]byte(nil), o.Content...)},
		fm:     fieldManager,
		reason: reasonRef,
		ctx:    ctx,
		resp:   make(chan writeResp, 1),
	}
	select {
	case r.reqs <- req:
	case <-r.stop:
		return 0, "", ErrClosed
	case <-ctx.Done():
		return 0, "", ctx.Err()
	}
	select {
	case resp := <-req.resp:
		return resp.rv, resp.action, resp.err
	case <-ctx.Done():
		return 0, "", ctx.Err()
	}
}

// objectKeyValidation (kind, scope, name) 合法性闸门（与 git 后端 concretePath 同一套
// 语义：kind 注册 + scope 四层 + 命名规则 + storage 模板相容——模板只作语义闸门，
// sqlite 后端以 (kind, scope, name) 为键，不物化为路径）。
func objectKeyValidation(sch *scheme.Scheme, kind, scope, name string) (scheme.Kind, error) {
	k, err := sch.Lookup(kind)
	if err != nil {
		return k, fmt.Errorf("%w: %v", ErrNotRegistered, err)
	}
	if scope == "" {
		return k, fmt.Errorf("%w: 写入/读取 %s 需显式 scope", ErrScope, k.Name)
	}
	if _, err := scopeDir(scope); err != nil {
		return k, err
	}
	if !nameSeg.MatchString(name) {
		return k, fmt.Errorf("%w: 对象名 %q 不合命名规则（小写中划线）", ErrScope, name)
	}
	if _, err := storagePatterns(k, scope, name, true); err != nil {
		return k, err // 模板相容性闸门（与 git 后端同判——scope/name 组合必须可寻址）
	}
	return k, nil
}

func (r *SQLiteRepo) applyWrite(req *writeReq) (int64, string, error) {
	// 出队即事务：SQL 读写脱离调用方 ctx（提交中不可中断，与 git 落盘段同口径）。
	o := req.o
	if err := req.ctx.Err(); err != nil {
		return 0, "", err
	}
	if _, err := objectKeyValidation(r.sch, o.Kind, o.Scope, o.Name); err != nil {
		return 0, "", err
	}
	// 入向文档校验：可解析 + 身份一致 + 凭据来源闸门（与 git 后端同一套）。
	inDoc, inContainer, inBody, inMD, err := splitDoc(o.Content)
	if err != nil {
		return 0, "", err
	}
	inMeta, err := metaOf(inContainer)
	if err != nil {
		return 0, "", err
	}
	if inMeta.Name != "" && inMeta.Name != o.Name {
		return 0, "", fmt.Errorf("%w: metadata.name=%q 与对象名 %q 不符", ErrIdentity, inMeta.Name, o.Name)
	}
	if inMeta.Scope != "" && inMeta.Scope != o.Scope {
		return 0, "", fmt.Errorf("%w: metadata.scope=%q 与 scope %q 不符", ErrIdentity, inMeta.Scope, o.Scope)
	}
	inCanon, err := canonicalWithoutRV(o.Content)
	if err != nil {
		return 0, "", err
	}
	if err := checkCredentialSource(o.Kind, inDoc); err != nil {
		return 0, "", err
	}

	var storedRV int64
	var storedCanon string
	err = r.db.QueryRow(`SELECT content, resource_version FROM doc WHERE kind=? AND scope=? AND name=?`,
		o.Kind, o.Scope, o.Name).Scan(&storedCanon, &storedRV)
	switch {
	case err == nil:
		// 库内存的是落库后的完整原文（含 resourceVersion）；幂等比对用去 rv 规范文本。
		cur, cerr := canonicalWithoutRV([]byte(storedCanon))
		if cerr != nil {
			return 0, "", fmt.Errorf("repository: docs DB 现行内容解析: %w", cerr)
		}
		if inCanon == cur {
			r.logApply(o, storedRV, "noop", req) // 幂等重入：内容一致 → no-op
			return storedRV, "noop", nil
		}
		if o.ResourceVersion == 0 {
			return 0, "", fmt.Errorf("%w: %s/%s/%s 现版本 %d", ErrAlreadyExists, o.Kind, o.Scope, o.Name, storedRV)
		}
		if o.ResourceVersion != storedRV {
			r.logApply(o, storedRV, "conflict", req)
			return 0, "", &ConflictError{Kind: o.Kind, Scope: o.Scope, Name: o.Name,
				Expected: o.ResourceVersion, Current: storedRV}
		}
	case errors.Is(err, sql.ErrNoRows):
		if o.ResourceVersion != 0 {
			// 写者读到过它、如今已不在：并发删除语义，按冲突处理（与 git 后端同口径）。
			r.logApply(o, 0, "conflict", req)
			return 0, "", &ConflictError{Kind: o.Kind, Scope: o.Scope, Name: o.Name,
				Expected: o.ResourceVersion, Current: 0}
		}
	default:
		return 0, "", fmt.Errorf("repository: docs DB 查询: %w", err)
	}

	action := "update"
	newRV := storedRV + 1
	if err == sql.ErrNoRows {
		action = "create"
		newRV = 1
	}
	if err := setResourceVersion(inContainer, newRV); err != nil {
		return 0, "", err
	}
	out, err := encodeDoc(inDoc, inBody, inMD)
	if err != nil {
		return 0, "", err
	}
	if _, err := r.db.Exec(`INSERT INTO doc (kind, scope, name, content, resource_version, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT (kind, scope, name) DO UPDATE SET
	content=excluded.content, resource_version=excluded.resource_version, updated_at=excluded.updated_at`,
		o.Kind, o.Scope, o.Name, string(out), newRV, stampNow()); err != nil {
		return 0, "", fmt.Errorf("repository: docs DB 落库: %w", err)
	}
	r.logApply(o, newRV, "ok", req)
	r.broadcast(fmt.Sprintf("%d", newRV))
	return newRV, action, nil
}

// logApply ops DB 双写钩子（与 GitRepo.logApply 同契约：ok/noop/conflict 行，ops-log.md §2）。
func (r *SQLiteRepo) logApply(o *Object, rv int64, result string, req *writeReq) {
	if r.opts.Ops == nil {
		return
	}
	_ = r.opts.Ops.AppendApply(opslog.ApplyLog{
		Ts: time.Now(), FieldManager: req.fm, Obj: fmt.Sprintf("%s/%s/%s", o.Kind, o.Scope, o.Name),
		ResourceVersion: rv, Result: result, ReasonRef: req.reason,
	})
}

func stampNow() string { return time.Now().UTC().Format(time.RFC3339Nano) }

// ─────────────────────────────────────────────────────────────────────────────
// 读路径：Get / List（语义与 git 后端一致：身份以文档内 metadata 为权威；只读面宽松）
// ─────────────────────────────────────────────────────────────────────────────

// Get 读取单个文档对象。
func (r *SQLiteRepo) Get(_ context.Context, kind, scope, name string) (*Object, error) {
	if _, err := objectKeyValidation(r.sch, kind, scope, name); err != nil {
		return nil, err
	}
	var raw string
	err := r.db.QueryRow(`SELECT content FROM doc WHERE kind=? AND scope=? AND name=?`,
		kind, scope, name).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, fmt.Errorf("%w: %s/%s/%s", ErrNotFound, kind, scope, name)
	}
	if err != nil {
		return nil, fmt.Errorf("repository: docs DB 读取: %w", err)
	}
	return docObject(kind, []byte(raw), scope, name)
}

// List 枚举对象：scope 给定则限定，scope="" 跨 scope 通配（该 kind 全部对象）。
func (r *SQLiteRepo) List(_ context.Context, kind, scope string) ([]*Object, error) {
	if _, err := r.sch.Lookup(kind); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotRegistered, err)
	}
	var rows *sql.Rows
	var err error
	if scope == "" {
		rows, err = r.db.Query(`SELECT content FROM doc WHERE kind=? ORDER BY scope, name`, kind)
	} else {
		rows, err = r.db.Query(`SELECT content FROM doc WHERE kind=? AND scope=? ORDER BY name`, kind, scope)
	}
	if err != nil {
		return nil, fmt.Errorf("repository: docs DB 枚举: %w", err)
	}
	defer rows.Close()
	var out []*Object
	for rows.Next() {
		var raw string
		if err := rows.Scan(&raw); err != nil {
			return nil, fmt.Errorf("repository: docs DB 读行: %w", err)
		}
		o, err := docObject(kind, []byte(raw), scope, "")
		if err != nil {
			continue // 坏行跳过不致命（只读面宽松，与 git 后端 List 同口径）
		}
		out = append(out, o)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Scope != out[j].Scope {
			return out[i].Scope < out[j].Scope
		}
		return out[i].Name < out[j].Name
	})
	return out, rows.Err()
}

// docObject 从存储原文投影 Object（身份以文档内 metadata 为权威；keyName 非空=读取请求键兜底）。
func docObject(kind string, raw []byte, scope, keyName string) (*Object, error) {
	doc, container, _, _, err := splitDoc(raw)
	if err != nil {
		return nil, err
	}
	_ = doc
	m, err := metaOf(container)
	if err != nil {
		return nil, err
	}
	objName := m.Name
	if objName == "" {
		objName = keyName
	}
	objScope := m.Scope
	if objScope == "" {
		objScope = scope
	}
	return &Object{Kind: kind, Name: objName, Scope: objScope, ResourceVersion: m.RV, Content: raw}, nil
}

// SeedFromLayout git 布局 → doc 表一次性种子导入（迁移路径：--docs 旧根 --docs-backend sqlite
// --docs-db 新库；空库才导入=幂等）。保留各对象 resourceVersion；布局里模板不可寻址/损坏的
// 文件跳过（与只读面宽松口径一致），返回导入行数。
func (r *SQLiteRepo) SeedFromLayout(root string) (int, error) {
	var n int
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM doc`).Scan(&n); err != nil {
		return 0, fmt.Errorf("repository: docs DB 计数: %w", err)
	}
	if n > 0 {
		return 0, nil // 非空库不导入（幂等——重复启动不重复灌）
	}
	inserted := 0
	for _, k := range r.sch.ListKinds() {
		objs, err := globObjects(root, r.sch, k.Name, "")
		if err != nil {
			continue // 该 kind 在布局下无可寻址分支（如 docType 单例文件）——跳过不致命
		}
		for _, o := range objs {
			if _, err := objectKeyValidation(r.sch, o.Kind, o.Scope, o.Name); err != nil {
				continue // 野对象/越界路径不导入（闸门口径与写路径一致）
			}
			if _, err := docObject(o.Kind, o.Content, o.Scope, o.Name); err != nil {
				continue // 损坏文件跳过
			}
			if _, err := r.db.Exec(`INSERT INTO doc (kind, scope, name, content, resource_version, updated_at)
VALUES (?, ?, ?, ?, ?, ?) ON CONFLICT (kind, scope, name) DO NOTHING`,
				o.Kind, o.Scope, o.Name, string(o.Content), o.ResourceVersion, stampNow()); err != nil {
				return inserted, fmt.Errorf("repository: 种子导入 %s/%s/%s: %w", o.Kind, o.Scope, o.Name, err)
			}
			inserted++
		}
	}
	return inserted, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Watch：进程内写直通知（单权威在库——无外部进程可改库，无需轮询兜底）
// ─────────────────────────────────────────────────────────────────────────────

// Watch 订阅库变更：先发当前代（对象数计数值），其后每次成功写发新代值。
// ctx 取消即关停（频道关闭）。语义与 GitRepo.Watch 对齐（首发现值 + 后续变更信号，
// 慢消费者丢旧保新——watch 是通知不是事件总线）。
func (r *SQLiteRepo) Watch(ctx context.Context) (<-chan string, error) {
	ch := make(chan string, 8)
	var init string
	if err := r.db.QueryRow(`SELECT COUNT(*) FROM doc`).Scan(&init); err != nil {
		init = "0"
	}
	r.mu.Lock()
	r.subs[ch] = struct{}{}
	r.mu.Unlock()
	go func() {
		defer func() {
			r.mu.Lock()
			delete(r.subs, ch)
			r.mu.Unlock()
			close(ch)
		}()
		sendNonBlock(ch, init)
		<-ctx.Done()
	}()
	return ch, nil
}

// broadcast 向全部订阅扇出一次变更信号（非阻塞丢旧保新——与 sendChange 同口径，无 ctx 阻塞）。
func (r *SQLiteRepo) broadcast(v string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for ch := range r.subs {
		sendNonBlock(ch, v)
	}
}

func sendNonBlock(ch chan string, v string) {
	select {
	case ch <- v:
	default:
		select {
		case <-ch:
		default:
		}
		select {
		case ch <- v:
		default:
		}
	}
}

// 编译期断言：SQLiteRepo 满足存储接口（Repository）。
var _ Repository = (*SQLiteRepo)(nil)
