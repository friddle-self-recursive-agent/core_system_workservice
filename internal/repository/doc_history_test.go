// doc_history_test.go —— 逐次正文历史（doc_history）与 sqlite rev-content 读面验收
// （R-DOC-7 规则 4：逐次账非周期快照——apply 成功且内容变化先落旧内容一行；rev=该次时点
// 正文逐字节；noop/conflict 不落行；create 无历史起点；清单不含正文；快照缺席=ErrRevNotFound
// 与对象缺席 ErrNotFound 分立，不冒充）。
package repository

import (
	"context"
	"errors"
	"strings"
	"testing"
)

func TestSQLiteHistoryCaptureOnContentChange(t *testing.T) {
	r, _ := newSQLiteRepo(t)
	ctx := context.Background()
	v1 := agentYAML("h1", "")
	if _, action, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "h1", Content: v1}, "fm-1", ""); err != nil || action != "create" {
		t.Fatalf("create: action=%s err=%v", action, err)
	}
	// 改两次内容 → history 两条，rev=1/2 正确（行 rev=被翻页那次的 rv）。
	v2 := agentYAML("h1", "\n  extra: 1")
	if rv, action, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "h1", ResourceVersion: 1, Content: v2}, "fm-2", ""); err != nil || rv != 2 || action != "update" {
		t.Fatalf("第一改: rv=%d action=%s err=%v", rv, action, err)
	}
	v3 := agentYAML("h1", "\n  extra: 2")
	if rv, action, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "h1", ResourceVersion: 2, Content: v3}, "fm-3", ""); err != nil || rv != 3 || action != "update" {
		t.Fatalf("第二改: rv=%d action=%s err=%v", rv, action, err)
	}
	// noop（幂等重入）与 conflict（乐观锁冲突）都无翻页 → 不落历史行。
	if _, action, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "h1", ResourceVersion: 3, Content: v3}, "fm-4", ""); err != nil || action != "noop" {
		t.Fatalf("noop: action=%s err=%v", action, err)
	}
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "h1", ResourceVersion: 1, Content: agentYAML("h1", "\n  extra: 9")}, "fm-5", ""); err == nil {
		t.Fatal("冲突写应失败")
	}
	hs, err := r.DocHistory("Agent", "global", "h1")
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(hs) != 2 {
		t.Fatalf("历史行数=%d（期望 2——仅两次内容翻页；noop/conflict 不落）", len(hs))
	}
	if hs[0].Rev != 2 || hs[1].Rev != 1 {
		t.Fatalf("rev 序应降序 [2 1]，得 [%d %d]", hs[0].Rev, hs[1].Rev)
	}
	if hs[0].FieldManager != "fm-3" || hs[1].FieldManager != "fm-2" {
		t.Fatalf("fieldManager 应为翻页写者（rev2←fm-3 rev1←fm-2），得 %q/%q", hs[0].FieldManager, hs[1].FieldManager)
	}
	if hs[0].AppliedAt.IsZero() || hs[1].AppliedAt.IsZero() {
		t.Fatal("appliedAt 应在册（翻页时刻）")
	}
	// 清单不含正文：Bytes=该次正文字节长度（与 rev-content 实取正文交叉对账）。
	c2, err := r.RevContent(ctx, "Agent", "global", "h1", 2)
	if err != nil {
		t.Fatalf("rev-content r2: %v", err)
	}
	if hs[0].Bytes != len(c2) {
		t.Fatalf("bytes 应与该次正文字节恒等（%d vs %d）", hs[0].Bytes, len(c2))
	}
	// 对象缺席=ErrNotFound（与快照缺席分立）。
	if _, err := r.DocHistory("Agent", "global", "absent"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("缺席对象应 ErrNotFound，得 %v", err)
	}
	// 在册无翻页史（仅创建）=空清单合法词形（不冒充）。
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "h2", Content: agentYAML("h2", "")}, "fm", ""); err != nil {
		t.Fatal(err)
	}
	hs2, err := r.DocHistory("Agent", "global", "h2")
	if err != nil || len(hs2) != 0 {
		t.Fatalf("零历史应为空清单，得 %d 行 err=%v", len(hs2), err)
	}
	// 键闸门：kind 未注册 / scope 缺失。
	if _, err := r.DocHistory("Nope", "global", "h1"); !errors.Is(err, ErrNotRegistered) {
		t.Fatalf("野对象应 ErrNotRegistered，得 %v", err)
	}
	if _, err := r.DocHistory("Agent", "", "h1"); !errors.Is(err, ErrScope) {
		t.Fatalf("空 scope 应 ErrScope，得 %v", err)
	}
}

func TestSQLiteRevContentFromDocHistory(t *testing.T) {
	r, _ := newSQLiteRepo(t)
	ctx := context.Background()
	_, _, _ = r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "h1", Content: agentYAML("h1", "")}, "fm-1", "")
	_, _, _ = r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "h1", ResourceVersion: 1, Content: agentYAML("h1", "\n  extra: 1")}, "fm-2", "")
	_, _, _ = r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "h1", ResourceVersion: 2, Content: agentYAML("h1", "\n  extra: 2")}, "fm-3", "")

	// rv=1/2=历史行正文逐字节（该次时点原文，rv 反写行在册；与现面分立=点行 diff 判据）。
	c1, err := r.RevContent(ctx, "Agent", "global", "h1", 1)
	if err != nil || !strings.Contains(c1, "resourceVersion: 1") || strings.Contains(c1, "extra: 1") {
		t.Fatalf("r1 应为第一版快照（无 v2 词形），err=%v 得 %.80q", err, c1)
	}
	c2, err := r.RevContent(ctx, "Agent", "global", "h1", 2)
	if err != nil || !strings.Contains(c2, "resourceVersion: 2") || !strings.Contains(c2, "extra: 1") {
		t.Fatalf("r2 应为第二版快照，err=%v 得 %.80q", err, c2)
	}
	// rv=现行 → doc 表现行正文（与 Get 逐字节恒等）。
	cur, err := r.Get(ctx, "Agent", "global", "h1")
	if err != nil {
		t.Fatal(err)
	}
	c3, err := r.RevContent(ctx, "Agent", "global", "h1", 3)
	if err != nil || c3 != string(cur.Content) {
		t.Fatalf("r3 应与现行正文逐字节恒等，err=%v", err)
	}
	// 快照缺席=ErrRevNotFound 分立词形（对象在册但无该翻页行——不冒充现行）。
	if _, err := r.RevContent(ctx, "Agent", "global", "h1", 9); !errors.Is(err, ErrRevNotFound) {
		t.Fatalf("r9 应 ErrRevNotFound，得 %v", err)
	}
	// 对象缺席=ErrNotFound（与快照缺席分立）。
	if _, err := r.RevContent(ctx, "Agent", "global", "absent", 1); !errors.Is(err, ErrNotFound) {
		t.Fatalf("缺席对象应 ErrNotFound，得 %v", err)
	}
	// 键闸门同口径。
	if _, err := r.RevContent(ctx, "Nope", "global", "h1", 1); !errors.Is(err, ErrNotRegistered) {
		t.Fatalf("野对象应 ErrNotRegistered，得 %v", err)
	}
	// 数据源词形标注（不冒充——与 git 后端 "git-backend" 分立）。
	if got := r.Source(); got != "sqlite-doc-history" {
		t.Fatalf("Source 应 sqlite-doc-history，得 %q", got)
	}
}

// TestSQLiteDelete —— ObjectDeleter 删除通路（2026-11-26 假数据清理批）：
// 删除=现行行移除+doc_history 终态行（rev=被删 rv）+乐观锁期望闸+缺席=ErrNotFound+重建从 rv1。
func TestSQLiteDelete(t *testing.T) {
	r, _ := newSQLiteRepo(t)
	ctx := context.Background()
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "d1", Content: agentYAML("d1", "")}, "fm-1", ""); err != nil {
		t.Fatalf("create: %v", err)
	}
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "d1", ResourceVersion: 1, Content: agentYAML("d1", "\n  extra: 1")}, "fm-2", ""); err != nil {
		t.Fatalf("update: %v", err)
	}
	// 乐观锁期望不符 → ConflictError，对象仍在册。
	if _, err := r.Delete("Agent", "global", "d1", 1, "fm-del", ""); err == nil {
		t.Fatal("expectRV=1（现 rv=2）应冲突")
	}
	if _, err := r.Get(ctx, "Agent", "global", "d1"); err != nil {
		t.Fatalf("冲突删后对象应在册: %v", err)
	}
	// 正删：rv=2 返回；Get=ErrNotFound；history 终态行 rev=2（删除前现行留账）。
	rv, err := r.Delete("Agent", "global", "d1", 2, "fm-del", "cleanup")
	if err != nil || rv != 2 {
		t.Fatalf("delete: rv=%d err=%v", rv, err)
	}
	if _, err := r.Get(ctx, "Agent", "global", "d1"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("删除后 Get 应 ErrNotFound，得 %v", err)
	}
	hs, err := r.DocHistory("Agent", "global", "d1")
	if err != nil || len(hs) == 0 {
		t.Fatalf("删除应留终态历史行: n=%d err=%v", len(hs), err)
	}
	if hs[0].Rev != 2 {
		t.Fatalf("终态行 rev 应=2（被删版本），得 %d", hs[0].Rev)
	}
	// 再删=缺席显式；重建从 rv=1（全新生命周期）。
	if _, err := r.Delete("Agent", "global", "d1", 0, "fm-del", ""); !errors.Is(err, ErrNotFound) {
		t.Fatalf("缺席删应 ErrNotFound，得 %v", err)
	}
	if rv2, action, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "d1", Content: agentYAML("d1", "")}, "fm-3", ""); err != nil || action != "create" || rv2 != 1 {
		t.Fatalf("重建应 create rv=1: rv=%d action=%s err=%v", rv2, action, err)
	}
}
