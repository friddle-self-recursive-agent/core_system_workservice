// sqlite_crossproc_test.go —— L88-01/L88-02 断言锚（loop-88 A 象机械根因随行验收）：
// 多进程共享同一 docs SQLite 库时（authority/执行体各自 --allow-apply 的真实部署形态），
// applyWrite 的现行行读取必须与翻页写入同事务（BEGIN IMMEDIATE 起事务即持写锁——读改写
// 原子）。SELECT 事务外时跨进程 read-modify-write TOCTOU：同基 rv 双写双行（rv 不前进）/
// rv 回退/已删行复活（loop-88 rev10 双写同基 rv 双行活体实证）。防御对称锚：Git 后端以
// HEAD 为基准天然防跨进程竞；SQLite 后端修复后同判。
// 本文件自带 DSN/建库辅助（不依赖被测文件符号以外的改动——须能在修复前 HEAD 上编译，
// 作红/绿对照锚；修复前：竞败者持陈旧基线照样提交成功=同基双行；修复后：败者必得
// conflict/busy=409 重读收敛语义，绝无同基双行）。
package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"
)

// crossProcDSN 第二进程句柄 DSN（与被测库同形：WAL+busy_timeout+immediate）。
func crossProcDSN(path string) string {
	return "file:" + path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_txlock=immediate"
}

// newCrossProcRepoAt 在定点目录建首个 repo 实例（第二实例用 openCrossProcRepo 同路径开启
// ——两实例=两进程各持串行队列共享同一库的等价模型）。
func newCrossProcRepoAt(t *testing.T, dir string) (*SQLiteRepo, *memOps) {
	t.Helper()
	ops := &memOps{}
	r, err := OpenSQLite(filepath.Join(dir, "docs.db"), mustScheme(t), &Options{Ops: ops}, "")
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = r.Close() })
	return r, ops
}

// openCrossProcRepo 第二 repo 实例（同库第二写面——独立串行队列，跨进程竞速真实形态）。
func openCrossProcRepo(t *testing.T, dir string) (*SQLiteRepo, *memOps) {
	return newCrossProcRepoAt(t, dir)
}

func isConflictErr(err error) bool {
	var ce *ConflictError
	return errors.As(err, &ce)
}

// isBusyErr SQLITE_BUSY/SQLITE_LOCKED（写锁被他进程事务持有；busy_timeout 窗口内等锁
// 超时=拒绝——调用方重试语义，零半程写入）。
func isBusyErr(err error) bool {
	s := err.Error()
	return strings.Contains(s, "SQLITE_BUSY") || strings.Contains(s, "SQLITE_LOCKED") ||
		strings.Contains(s, "database is locked") || strings.Contains(s, "database table is locked")
}

// TestSQLiteCrossProcImmediateLockShape L88-01 机械判据（确定性、无竞速依赖）：repo 自身
// 连接上开启写事务后，第二进程的 BEGIN IMMEDIATE 必须被拒（写锁自事务开启即被持有——
// 基线读取在锁窗之内）。修复前（BEGIN 默认 deferred、SELECT 事务外）：事务开启不持写锁，
// 第二进程 BEGIN IMMEDIATE 直取成功=锁窗不覆盖基线读（TOCTOU 开放）。
func TestSQLiteCrossProcImmediateLockShape(t *testing.T) {
	dir := t.TempDir()
	r, _ := newCrossProcRepoAt(t, dir)
	db2, err := sql.Open(sqliteDriver, "file:"+filepath.Join(dir, "docs.db")+
		"?_pragma=journal_mode(WAL)&_pragma=busy_timeout(300)&_txlock=immediate")
	if err != nil {
		t.Fatal(err)
	}
	defer db2.Close()

	txA, err := r.db.Begin() // 被测写面的事务开启
	if err != nil {
		t.Fatalf("repo 事务开启: %v", err)
	}
	defer txA.Rollback()

	tx2, err := db2.Begin() // 第二进程立即事务：修复后=写锁已被持有→busy 拒绝
	if err == nil {
		_ = tx2.Rollback()
		t.Fatal("L88-01 破防：repo 写事务开启后第二进程 BEGIN IMMEDIATE 直取成功" +
			"（写锁未随事务开启即持有——基线读取在锁窗外，跨进程 TOCTOU 开放）")
	}
	if !isBusyErr(err) {
		t.Fatalf("第二进程 BEGIN IMMEDIATE 应 busy 拒绝，得 %v", err)
	}
}

// TestSQLiteCrossProcWriteLostUpdate L88-01 竞速锚（双实例活体竞速）：两 repo 实例并发
// 同基写入同一对象——不变量=rev=1 历史行至多一行、rv 不回退、败者必 conflict/busy
// （409 重读收敛），绝不出现「败者持陈旧基线照样提交成功」（loop-88 rev10 双写同基双行）。
func TestSQLiteCrossProcWriteLostUpdate(t *testing.T) {
	dir := t.TempDir()
	repoA, _ := newCrossProcRepoAt(t, dir)
	repoB, _ := openCrossProcRepo(t, dir)
	ctx := context.Background()
	rv1, _, err := repoA.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "x1",
		Content: agentYAML("x1", "")}, "proc-a", "")
	if err != nil || rv1 != 1 {
		t.Fatalf("首写: rv=%d err=%v", rv1, err)
	}

	type outcome struct {
		rv     int64
		action string
		err    error
	}
	var wg sync.WaitGroup
	out := make([]outcome, 2)
	start := time.Now()
	wg.Add(2)
	for i, rr := range []*SQLiteRepo{repoA, repoB} {
		go func(i int, rr *SQLiteRepo) {
			defer wg.Done()
			rv, action, werr := rr.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "x1",
				ResourceVersion: rv1, Content: agentYAML("x1", fmt.Sprintf("-%d", i))},
				fmt.Sprintf("proc-%d", i), "l88-anchor")
			out[i] = outcome{rv, action, werr}
			_ = start
		}(i, rr)
	}
	wg.Wait()

	winners := 0
	var finalRV int64
	for i, o := range out {
		switch {
		case o.err == nil:
			winners++
			if o.rv != 2 || o.action != "update" {
				t.Fatalf("胜者应 rv=2/update，实例 %d 得 rv=%d/%s", i, o.rv, o.action)
			}
		case isConflictErr(o.err) || isBusyErr(o.err):
			// 败者=conflict（现库 rv 已前进，409 重读收敛）或 busy（写锁拒斥重试）——合法。
		default:
			t.Fatalf("实例 %d 非法败因（应 conflict/busy）: %v", i, o.err)
		}
	}
	if winners != 1 {
		t.Fatalf("L88-01 破防：同基并发写应恰一胜者，得 %d（%+v）", winners, out)
	}
	if err := repoA.db.QueryRow(`SELECT resource_version FROM doc WHERE kind='Agent' AND scope='global' AND name='x1'`).Scan(&finalRV); err != nil {
		t.Fatal(err)
	}
	if finalRV != 2 {
		t.Fatalf("L88-01 破防：rv 应恰前进一位（2），得 %d", finalRV)
	}
	var hist int
	if err := repoA.db.QueryRow(`SELECT COUNT(*) FROM doc_history WHERE kind='Agent' AND scope='global' AND name='x1' AND rev=1`).Scan(&hist); err != nil {
		t.Fatal(err)
	}
	if hist != 1 {
		t.Fatalf("L88-01 破防：同基 rv 双写双行（rev=1 应恰一行）hist=%d", hist)
	}
}

// TestSQLiteCrossProcWriteVsDelete L88-02 语义保全锚：缺席+expectRV>0=conflict（绝不
// upsert 复活——D 象通路封闭）；apply_log 落 conflict 行（写者以 409 重读收敛的账面凭据）。
func TestSQLiteCrossProcWriteVsDelete(t *testing.T) {
	dir := t.TempDir()
	repoA, ops := newCrossProcRepoAt(t, dir)
	repoB, _ := openCrossProcRepo(t, dir)
	ctx := context.Background()
	wf := func(rv int64) *Object {
		return &Object{Kind: "Workflow", Scope: "global", Name: "w1", ResourceVersion: rv,
			Content: []byte("apiVersion: core/v1\nkind: Workflow\nmetadata:\n  name: w1\n  scope: global\nspec:\n  steps: []\n")}
	}
	if _, _, err := repoA.Write(ctx, wf(0), "proc-a", ""); err != nil {
		t.Fatal(err)
	}
	// 第二实例（跨进程）删除现行行（rv=1）。
	if _, err := repoB.Delete("Workflow", "global", "w1", 1, "proc-b", "l88-anchor"); err != nil {
		t.Fatalf("第二实例删除: %v", err)
	}
	// 首实例持陈旧期望（rv=1、对象已缺席）提交：必须 conflict，绝不复活。
	_, _, err := repoA.Write(ctx, wf(1), "proc-a", "l88-anchor-stale")
	if !isConflictErr(err) {
		t.Fatalf("L88-02 破防：缺席+expectRV>0 应 conflict，得 %v", err)
	}
	var alive int
	if err := repoA.db.QueryRow(`SELECT COUNT(*) FROM doc WHERE kind='Workflow' AND scope='global' AND name='w1'`).Scan(&alive); err != nil {
		t.Fatal(err)
	}
	if alive != 0 {
		t.Fatalf("L88-02 破防：已删对象复活（D 象通路）alive=%d", alive)
	}
	found := false
	for _, res := range ops.results() {
		if res == "conflict" {
			found = true
		}
	}
	if !found {
		t.Fatalf("L88-02 破防：apply_log 应含 conflict 行，得 %v", ops.results())
	}
}
