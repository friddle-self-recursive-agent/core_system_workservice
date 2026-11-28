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
// 文件库连接池放开为读侧并行（WAL 读者不阻塞写者；写者互斥由 busy_timeout 兜底——单写者
// 语义由串行队列保证，池只在读路径并行）；内存库恒串行（:memory: 每连接=独立库，池>1 即分裂）。
type SQLiteRepo struct {
	db       *sql.DB
	maxConns int // 读连接池上限（1=内存库串行；>1=文件库 WAL 读侧并行——G-15 机械判据锚）
	sch      *scheme.Scheme
	opts     Options

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
		dsn = sqliteDSN(path)
	}
	db, err := sql.Open(sqliteDriver, dsn)
	if err != nil {
		return nil, fmt.Errorf("repository: 打开 docs DB %s: %w", path, err)
	}
	r := &SQLiteRepo{
		db:   db,
		sch:  sch,
		reqs: make(chan *writeReq),
		stop: make(chan struct{}),
		done: make(chan struct{}),
		subs: map[chan string]struct{}{},
	}
	// 连接池（G-15 读侧高性能）：文件库=WAL+busy_timeout（DSN pragma 每连接生效）+
	// 池放开——读多写少读面并行、写面仍单 goroutine 串行队列（单写者语义不变）；
	// 内存库恒 1（:memory: 多连接=多个独立库，池>1 即表分裂）。
	if path == ":memory:" {
		db.SetMaxOpenConns(1)
		r.maxConns = 1
	} else {
		db.SetMaxOpenConns(sqliteReadPool)
		db.SetMaxIdleConns(sqliteReadPool)
		r.maxConns = sqliteReadPool
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

// sqliteDSN 文件库 DSN：WAL + busy_timeout + _txlock=immediate。_txlock 使 sql.Begin
// 以 BEGIN IMMEDIATE 起事务——事务开启即持写锁，事务内首条 SELECT 读到的现行行与随后的
// 翻页写入原子（loop-88 A 象修复：跨进程 read-modify-write TOCTOU——SELECT 事务外时，
// 他进程可在读与写之间翻页，同基 rv 双写双行/rv 回退/已删行复活；Git 后端以 HEAD 为基准
// 天然防竞，本参数补齐 SQLite 后端防御对称）。
func sqliteDSN(path string) string {
	return "file:" + path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_txlock=immediate"
}

// sqliteReadPool 文件库读连接池上限（G-15 读侧高性能——WAL 模式读者不阻塞写者；写面单
// goroutine 串行队列=单写者，池只放大读路径并行度；busy_timeout=5000 兜底残余写互斥窗口）。
const sqliteReadPool = 4

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
CREATE INDEX IF NOT EXISTS doc_kind_scope ON doc(kind, scope);
CREATE TABLE IF NOT EXISTS doc_history (
	id            INTEGER PRIMARY KEY AUTOINCREMENT,
	kind          TEXT NOT NULL,
	scope         TEXT NOT NULL,
	name          TEXT NOT NULL,
	rev           INTEGER NOT NULL,
	content       TEXT NOT NULL,
	applied_at    TEXT NOT NULL,
	field_manager TEXT NOT NULL
);
CREATE INDEX IF NOT EXISTS doc_history_key ON doc_history(kind, scope, name, rev);`
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
	var rv int64
	var action string
	var err error
	if req.del != nil {
		rv, err = r.applyDelete(req)
		action = "delete"
	} else {
		rv, action, err = r.applyWrite(req)
	}
	select {
	case req.resp <- writeResp{rv: rv, action: action, err: err}:
	case <-req.ctx.Done():
	}
}

// Delete 实现 ObjectDeleter（串行队列同通道；expectRV>0=乐观锁期望）。
// 删除=现行行移除 + doc_history 终态兜底行（rev=被删 rv、正文=库内原文；仅当该 rv 尚无账
// 行时补——每版一行的新口径下被删版本在写入时已落账，不重复插） +
// ops apply_log "deleted" 行 + watch 广播——删除也是账面动作（R-DOC-7 规则 4；
// workflow.md §9a bug 文档修复后删除即走此通路）。
func (r *SQLiteRepo) Delete(kind, scope, name string, expectRV int64, fieldManager, reasonRef string) (int64, error) {
	if fieldManager == "" {
		return 0, ErrNoFieldManager
	}
	req := &writeReq{
		del:    &deleteReq{kind: kind, scope: scope, name: name, expectRV: expectRV},
		fm:     fieldManager,
		reason: reasonRef,
		ctx:    context.Background(),
		resp:   make(chan writeResp, 1),
	}
	select {
	case r.reqs <- req:
	case <-r.stop:
		return 0, ErrClosed
	}
	resp := <-req.resp
	return resp.rv, resp.err
}

// applyDelete 删除事务（串行队列内）：键校验→现行行读取（缺席=ErrNotFound；期望 rv 不符=
// ConflictError）→doc_history 终态兜底行（该 rv 无账行才补）+现行行删除同事务→ops 双写+广播。
func (r *SQLiteRepo) applyDelete(req *writeReq) (int64, error) {
	d := req.del
	if _, err := objectKeyValidation(r.sch, d.kind, d.scope, d.name); err != nil {
		return 0, err
	}
	var storedRV int64
	var storedCanon string
	// L88-02（loop-88 A 象修复同拍）：现行行读取移入立即事务——删除判定的读与删原子；
	// 跨进程写面在本事务持锁期间不得插入翻页/复活行。
	tx, terr := r.db.Begin()
	if terr != nil {
		return 0, fmt.Errorf("repository: docs DB 事务: %w", terr)
	}
	err := tx.QueryRow(`SELECT content, resource_version FROM doc WHERE kind=? AND scope=? AND name=?`,
		d.kind, d.scope, d.name).Scan(&storedCanon, &storedRV)
	if errors.Is(err, sql.ErrNoRows) {
		_ = tx.Rollback()
		r.logApply(&Object{Kind: d.kind, Scope: d.scope, Name: d.name}, 0, "delete-missing", req)
		return 0, fmt.Errorf("%w: %s/%s/%s", ErrNotFound, d.kind, d.scope, d.name)
	}
	if err != nil {
		_ = tx.Rollback()
		return 0, fmt.Errorf("repository: docs DB 查询: %w", err)
	}
	if d.expectRV > 0 && d.expectRV != storedRV {
		_ = tx.Rollback()
		r.logApply(&Object{Kind: d.kind, Scope: d.scope, Name: d.name}, storedRV, "conflict", req)
		return 0, &ConflictError{Kind: d.kind, Scope: d.scope, Name: d.name,
			Expected: d.expectRV, Current: storedRV}
	}
	// issue #23 每版一行口径：被删版本在写入时已落 doc_history 同 rev 行——删除不重复插，
	// 避免同 rev 双行破坏逐次账唯一性。仅当该 rv 无账行（历史库/种子导入前的旧数据）才补
	// 终态行兜底，保证「删掉的版本正文仍可考古」不破。条件插入与删除同事务，任一步失败同滚。
	if _, err := tx.Exec(`INSERT INTO doc_history
	(kind, scope, name, rev, content, applied_at, field_manager)
SELECT ?, ?, ?, ?, ?, ?, ? WHERE NOT EXISTS
	(SELECT 1 FROM doc_history WHERE kind=? AND scope=? AND name=? AND rev=?)`,
		d.kind, d.scope, d.name, storedRV, storedCanon, stampNow(), "console:delete:"+req.fm,
		d.kind, d.scope, d.name, storedRV); err != nil {
		_ = tx.Rollback()
		return 0, fmt.Errorf("repository: docs DB 删除留账: %w", err)
	}
	if _, err := tx.Exec(`DELETE FROM doc WHERE kind=? AND scope=? AND name=?`,
		d.kind, d.scope, d.name); err != nil {
		_ = tx.Rollback()
		return 0, fmt.Errorf("repository: docs DB 删除: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, fmt.Errorf("repository: docs DB 提交: %w", err)
	}
	obj := &Object{Kind: d.kind, Scope: d.scope, Name: d.name}
	r.logApply(obj, storedRV, "deleted", req)
	r.broadcast(fmt.Sprintf("%d", storedRV))
	return storedRV, nil
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

	// L88-02/L88-01（loop-88 A 象修复）：SELECT 移入立即事务——BEGIN IMMEDIATE 即持写锁，
	// 读到的现行行与翻页写入原子；跨进程第二写面（共享同一 docs 库的 authority/执行体）
	// 在本事务持锁期间被拒（busy_timeout 等锁/超时重试），杜绝「读到基线 rv→他进程翻页→
	// 陈旧基线覆盖提交」的 TOCTOU 窗口（防御与 Git 后端 HEAD 基准对称）。
	var storedRV int64
	var storedCanon string
	tx, terr := r.db.Begin()
	if terr != nil {
		return 0, "", fmt.Errorf("repository: docs DB 事务: %w", terr)
	}
	err = tx.QueryRow(`SELECT content, resource_version FROM doc WHERE kind=? AND scope=? AND name=?`,
		o.Kind, o.Scope, o.Name).Scan(&storedCanon, &storedRV)
	switch {
	case err == nil:
		// 库内存的是落库后的完整原文（含 resourceVersion）；幂等比对用去 rv 规范文本。
		cur, cerr := canonicalWithoutRV([]byte(storedCanon))
		if cerr != nil {
			_ = tx.Rollback()
			return 0, "", fmt.Errorf("repository: docs DB 现行内容解析: %w", cerr)
		}
		if inCanon == cur {
			_ = tx.Rollback()
			r.logApply(o, storedRV, "noop", req) // 幂等重入：内容一致 → no-op
			return storedRV, "noop", nil
		}
		if o.ResourceVersion == 0 {
			_ = tx.Rollback()
			return 0, "", fmt.Errorf("%w: %s/%s/%s 现版本 %d", ErrAlreadyExists, o.Kind, o.Scope, o.Name, storedRV)
		}
		if o.ResourceVersion != storedRV {
			_ = tx.Rollback()
			r.logApply(o, storedRV, "conflict", req)
			return 0, "", &ConflictError{Kind: o.Kind, Scope: o.Scope, Name: o.Name,
				Expected: o.ResourceVersion, Current: storedRV}
		}
	case errors.Is(err, sql.ErrNoRows):
		if o.ResourceVersion != 0 {
			// 写者读到过它、如今已不在：并发删除语义，按冲突处理（与 git 后端同口径）。
			_ = tx.Rollback()
			r.logApply(o, 0, "conflict", req)
			return 0, "", &ConflictError{Kind: o.Kind, Scope: o.Scope, Name: o.Name,
				Expected: o.ResourceVersion, Current: 0}
		}
	default:
		_ = tx.Rollback()
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
	// 权威落库 + 逐次正文历史捕获同事务（事务上方已开——SELECT 与写入原子，loop-88 A 象
	// 修复）：每次「产生新版本」的写（create/update，含回滚/fromDraft 等特殊语义的 apply）
	// 都把本次产生的新版本正文落 doc_history 一行——
	//   rev=本次产生的新 resourceVersion、正文=落库原文逐字节、field_manager=本次写者、
	//   applied_at=落库时刻（issue #23：旧口径只记「被替换版本」导致当前版本 rv 在逐次账
	//   缺席——回滚写 rv=4 查无对应行；改为每版一行后「每次改动都有一行账」，回滚写同样
	//   落 rev=4 行，账面不再断环）。
	// 再 upsert 现行行；任一步失败同滚——不产生「现行已翻页、历史缺席该 rv 行」的缺口。
	// noop/conflict 不落（无新版本产生，本就不该有行）。现行行仍在 doc 表（rv 反写行在册，
	// RevContent 现行命中优先读 doc；历史 rv 读 doc_history 同 rev 行——两源正文逐字节同源）。
	if _, err := tx.Exec(`INSERT INTO doc_history
	(kind, scope, name, rev, content, applied_at, field_manager)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
		o.Kind, o.Scope, o.Name, newRV, string(out), stampNow(), req.fm); err != nil {
		_ = tx.Rollback()
		return 0, "", fmt.Errorf("repository: docs DB 历史落行: %w", err)
	}
	if _, err := tx.Exec(`INSERT INTO doc (kind, scope, name, content, resource_version, updated_at)
VALUES (?, ?, ?, ?, ?, ?)
ON CONFLICT (kind, scope, name) DO UPDATE SET
	content=excluded.content, resource_version=excluded.resource_version, updated_at=excluded.updated_at`,
		o.Kind, o.Scope, o.Name, string(out), newRV, stampNow()); err != nil {
		_ = tx.Rollback()
		return 0, "", fmt.Errorf("repository: docs DB 落库: %w", err)
	}
	if err := tx.Commit(); err != nil {
		return 0, "", fmt.Errorf("repository: docs DB 提交: %w", err)
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
// 实现注记（G-15 读侧高性能）：元数据直读 doc 表列（scope/name/resource_version）构造
// Object，不再逐行 splitDoc+YAML 解析——列=写路径身份闸门的权威投影（写入时 metadata.name/
// scope 与对象键强一致校验、resourceVersion 同拍反写正文；种子导入同闸），列直读与既有
// docObject 投影逐字段等价。Content 仍随行（Repository 接口契约：调度器/索引/导出等消费方
// 现读正文；列表瘦身面走 ListMeta——见下）。
func (r *SQLiteRepo) List(_ context.Context, kind, scope string) ([]*Object, error) {
	if _, err := r.sch.Lookup(kind); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotRegistered, err)
	}
	var rows *sql.Rows
	var err error
	if scope == "" {
		rows, err = r.db.Query(`SELECT scope, name, resource_version, content FROM doc WHERE kind=? ORDER BY scope, name`, kind)
	} else {
		rows, err = r.db.Query(`SELECT scope, name, resource_version, content FROM doc WHERE kind=? AND scope=? ORDER BY name`, kind, scope)
	}
	if err != nil {
		return nil, fmt.Errorf("repository: docs DB 枚举: %w", err)
	}
	defer rows.Close()
	var out []*Object
	for rows.Next() {
		var o Object
		var raw string
		if err := rows.Scan(&o.Scope, &o.Name, &o.ResourceVersion, &raw); err != nil {
			return nil, fmt.Errorf("repository: docs DB 读行: %w", err)
		}
		o.Kind = kind
		o.Content = []byte(raw)
		out = append(out, &o)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Scope != out[j].Scope {
			return out[i].Scope < out[j].Scope
		}
		return out[i].Name < out[j].Name
	})
	return out, rows.Err()
}

// ListMeta 元数据直读清单（G-15——列表投影面瘦身：SELECT 不取 content 列、零逐行 YAML
// 解析；正文仅单读路径 Get/RevContent 拉）。行集键/rv/排序与 List 同一（doc 表列权威投影，
// 见 List 实现注记）；Object.Content=nil——消费方（console 对象索引/关注报告/snapshot
// objects 桶等只需键与 rv 的读面）按需 Get 拉正文，不冒充全量行。
func (r *SQLiteRepo) ListMeta(_ context.Context, kind, scope string) ([]*Object, error) {
	if _, err := r.sch.Lookup(kind); err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotRegistered, err)
	}
	var rows *sql.Rows
	var err error
	if scope == "" {
		rows, err = r.db.Query(`SELECT scope, name, resource_version FROM doc WHERE kind=? ORDER BY scope, name`, kind)
	} else {
		rows, err = r.db.Query(`SELECT scope, name, resource_version FROM doc WHERE kind=? AND scope=? ORDER BY name`, kind, scope)
	}
	if err != nil {
		return nil, fmt.Errorf("repository: docs DB 元数据枚举: %w", err)
	}
	defer rows.Close()
	var out []*Object
	for rows.Next() {
		var o Object
		if err := rows.Scan(&o.Scope, &o.Name, &o.ResourceVersion); err != nil {
			return nil, fmt.Errorf("repository: docs DB 元数据读行: %w", err)
		}
		o.Kind = kind
		out = append(out, &o)
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
			// issue #23 特殊写入路径记账：种子导入直接 INSERT doc（绕过 applyWrite），
			// 旧口径下逐次账零行=「导入对象无账」破口。补齐导入版本本身的一行账（rev=导入
			// rv、正文=布局原文逐字节、field_manager=seed:import 诚实标注导入来源）；导入前
			// 的历史版本正文布局未提供，不伪造——口径=导入对象逐次账自导入版本起算。
			if o.ResourceVersion > 0 {
				if _, err := r.db.Exec(`INSERT INTO doc_history
	(kind, scope, name, rev, content, applied_at, field_manager)
VALUES (?, ?, ?, ?, ?, ?, ?)`,
					o.Kind, o.Scope, o.Name, o.ResourceVersion, string(o.Content), stampNow(), "seed:import"); err != nil {
					return inserted, fmt.Errorf("repository: 种子历史落账 %s/%s/%s: %w", o.Kind, o.Scope, o.Name, err)
				}
			}
			inserted++
		}
	}
	return inserted, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// 逐次正文历史读面（R-DOC-7 规则 4：doc_history=逐次账——周期 git 导出快照不得冒充；
// 读走连接池直连（G-15：文件库 WAL 读侧并行、写面串行队列单写者——读写互不阻塞））
// ─────────────────────────────────────────────────────────────────────────────

// DocHistoryRow 逐次历史行投影（不含正文——列表面瘦身；正文经 RevContent 读面）。
type DocHistoryRow struct {
	Rev          int64
	AppliedAt    time.Time
	FieldManager string
	Bytes        int // 该次正文字节长度（列表面容量观察，不携正文）
}

// History 满足 apiserver.DocHistoryFace（GET /v1/doc-history 清单读面）——语义同 DocHistory，
// 双名词形只为不改读面装配层契约（face 方法名 History，实现主名 DocHistory 保读面同源）。
func (r *SQLiteRepo) History(kind, scope, name string) ([]DocHistoryRow, error) {
	return r.DocHistory(kind, scope, name)
}

// DocHistory 对象逐次正文历史行集（rev 降序=最新翻页在前）。对象不在册（doc 与 history
// 双缺）→ ErrNotFound；在册尚无翻页史（仅创建）→ 空行集（零历史=合法词形，不冒充）。
func (r *SQLiteRepo) DocHistory(kind, scope, name string) ([]DocHistoryRow, error) {
	if _, err := objectKeyValidation(r.sch, kind, scope, name); err != nil {
		return nil, err
	}
	rows, err := r.db.Query(`SELECT rev, applied_at, field_manager, LENGTH(CAST(content AS BLOB)) FROM doc_history
WHERE kind=? AND scope=? AND name=? ORDER BY rev DESC`, kind, scope, name)
	if err != nil {
		return nil, fmt.Errorf("repository: docs DB 历史查询: %w", err)
	}
	defer rows.Close()
	out := []DocHistoryRow{}
	for rows.Next() {
		var h DocHistoryRow
		var applied string
		if err := rows.Scan(&h.Rev, &applied, &h.FieldManager, &h.Bytes); err != nil {
			return nil, fmt.Errorf("repository: docs DB 历史读行: %w", err)
		}
		if t, perr := time.Parse(time.RFC3339Nano, applied); perr == nil {
			h.AppliedAt = t
		}
		out = append(out, h)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("repository: docs DB 历史枚举: %w", err)
	}
	if len(out) == 0 {
		// 零历史行：对象在册=合法零词形；对象缺席=ErrNotFound（与快照缺席分立，不冒充）。
		var n int
		if err := r.db.QueryRow(`SELECT COUNT(*) FROM doc WHERE kind=? AND scope=? AND name=?`,
			kind, scope, name).Scan(&n); err != nil {
			return nil, fmt.Errorf("repository: docs DB 计数: %w", err)
		}
		if n == 0 {
			return nil, fmt.Errorf("%w: %s/%s/%s", ErrNotFound, kind, scope, name)
		}
	}
	return out, nil
}

// RevContent 历史版本正文读面（RevContentFace 契约——GET /v1/rev-content 读半边；rv 由
// face 层验为正整数）。数据源=doc_history 逐次账（R-DOC-7 规则 4，issue #23 每版一行口径）：
// rv=现行 → doc 表现行正文（同 rv 的 history 行亦在册，现行命中优先读 doc 省一跳）；
// rv<现行 → doc_history 同 rev 行正文（该版落库原文逐字节）；其余 → ErrRevNotFound
// （快照缺席，与对象缺席 ErrNotFound 分立——不冒充现行）。读面零副作用零写路径。
func (r *SQLiteRepo) RevContent(_ context.Context, kind, scope, name string, rv int64) (string, error) {
	if _, err := objectKeyValidation(r.sch, kind, scope, name); err != nil {
		return "", err
	}
	// 现行命中优先：现行正文即该 rv 时点正文（doc 表 resource_version 即该 rv）。
	var curRaw string
	var curRV int64
	err := r.db.QueryRow(`SELECT content, resource_version FROM doc WHERE kind=? AND scope=? AND name=?`,
		kind, scope, name).Scan(&curRaw, &curRV)
	switch {
	case err == nil && curRV == rv:
		return curRaw, nil
	case errors.Is(err, sql.ErrNoRows):
		return "", fmt.Errorf("%w: %s/%s/%s", ErrNotFound, kind, scope, name)
	case err != nil:
		return "", fmt.Errorf("repository: docs DB 读取: %w", err)
	}
	var raw string
	var got int64
	err = r.db.QueryRow(`SELECT content, rev FROM doc_history WHERE kind=? AND scope=? AND name=? AND rev=?`,
		kind, scope, name, rv).Scan(&raw, &got)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w: %s/%s/%s r%d（doc_history 无该翻页行——逐次账缺席，不冒充现行）",
			ErrRevNotFound, kind, scope, name, rv)
	}
	if err != nil {
		return "", fmt.Errorf("repository: docs DB 历史正文查询: %w", err)
	}
	return raw, nil
}

// Source face 词形标注（不冒充——rev-content 响应 source 据实标注数据源，与 git 后端
// "git-backend" 分立：本面=doc_history 逐次账）。
func (r *SQLiteRepo) Source() string { return "sqlite-doc-history" }

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
