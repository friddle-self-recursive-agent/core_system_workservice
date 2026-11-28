// doc_history_test.go —— 逐次正文历史（doc_history）与 sqlite rev-content 读面验收
// （R-DOC-7 规则 4：逐次账非周期快照——issue #23 起每版一行：create/update（含回滚/fromDraft
// 特殊语义）都把本次产生的新版本正文落一行，rev=该版本 rv、fieldManager=本次写者；noop/conflict
// 无新版本不落行；清单不含正文；快照缺席=ErrRevNotFound 与对象缺席 ErrNotFound 分立，不冒充）。
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
	// 每版一行：create 落 rev=1，改两次内容各落 rev=2/3（行 rev=该版本自带 rv）。
	v2 := agentYAML("h1", "\n  extra: 1")
	if rv, action, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "h1", ResourceVersion: 1, Content: v2}, "fm-2", ""); err != nil || rv != 2 || action != "update" {
		t.Fatalf("第一改: rv=%d action=%s err=%v", rv, action, err)
	}
	v3 := agentYAML("h1", "\n  extra: 2")
	if rv, action, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "h1", ResourceVersion: 2, Content: v3}, "fm-3", ""); err != nil || rv != 3 || action != "update" {
		t.Fatalf("第二改: rv=%d action=%s err=%v", rv, action, err)
	}
	// noop（幂等重入）与 conflict（乐观锁冲突）都无新版本 → 不落历史行。
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
	if len(hs) != 3 {
		t.Fatalf("历史行数=%d（期望 3——create/两改各一行；noop/conflict 不落）", len(hs))
	}
	if hs[0].Rev != 3 || hs[1].Rev != 2 || hs[2].Rev != 1 {
		t.Fatalf("rev 序应降序 [3 2 1]，得 [%d %d %d]", hs[0].Rev, hs[1].Rev, hs[2].Rev)
	}
	if hs[0].FieldManager != "fm-3" || hs[1].FieldManager != "fm-2" || hs[2].FieldManager != "fm-1" {
		t.Fatalf("fieldManager 应为该版写者（r3←fm-3 r2←fm-2 r1←fm-1），得 %q/%q/%q", hs[0].FieldManager, hs[1].FieldManager, hs[2].FieldManager)
	}
	if hs[0].AppliedAt.IsZero() || hs[1].AppliedAt.IsZero() || hs[2].AppliedAt.IsZero() {
		t.Fatal("appliedAt 应在册（落库时刻）")
	}
	// 清单不含正文：Bytes=该版正文字节长度（与 rev-content 同 rev 实取正文交叉对账）。
	for i, rv := range []int64{3, 2, 1} {
		c, err := r.RevContent(ctx, "Agent", "global", "h1", rv)
		if err != nil {
			t.Fatalf("rev-content r%d: %v", rv, err)
		}
		if hs[i].Bytes != len(c) {
			t.Fatalf("r%d bytes 应与该版正文恒等（%d vs %d）", rv, hs[i].Bytes, len(c))
		}
	}
	// 对象缺席=ErrNotFound（与快照缺席分立）。
	if _, err := r.DocHistory("Agent", "global", "absent"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("缺席对象应 ErrNotFound，得 %v", err)
	}
	// 仅创建对象也有一行账（每版一行含创建版——issue #23 逐次账无缺口）。
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "h2", Content: agentYAML("h2", "")}, "fm", ""); err != nil {
		t.Fatal(err)
	}
	hs2, err := r.DocHistory("Agent", "global", "h2")
	if err != nil || len(hs2) != 1 || hs2[0].Rev != 1 {
		t.Fatalf("仅创建应恰好 rev=1 一行，得 %d 行 err=%v", len(hs2), err)
	}
	// 键闸门：kind 未注册 / scope 缺失。
	if _, err := r.DocHistory("Nope", "global", "h1"); !errors.Is(err, ErrNotRegistered) {
		t.Fatalf("野对象应 ErrNotRegistered，得 %v", err)
	}
	if _, err := r.DocHistory("Agent", "", "h1"); !errors.Is(err, ErrScope) {
		t.Fatalf("空 scope 应 ErrScope，得 %v", err)
	}
}

// TestSQLiteRollbackWriteLandsHistoryRow issue #23 回归锚：回滚写（把正文改回历史版本）
// 必须落 doc_history 对应 rev 行——对象现值 rv=4 时账面须含 rev=4，不得只到 rev=3。
func TestSQLiteRollbackWriteLandsHistoryRow(t *testing.T) {
	r, _ := newSQLiteRepo(t)
	ctx := context.Background()
	steps := []struct {
		rv      int64
		content []byte
		fm      string
	}{
		{0, agentYAML("rb", ""), "corectl"},
		{1, agentYAML("rb", "\n  extra: 1"), "console:doc-chat"},
		{2, agentYAML("rb", "\n  extra: 2"), "probe-cleanup"},
		// 回滚写：正文退回 v2 词形，rv=3→4（issue #23 实证的那次回滚）。
		{3, agentYAML("rb", "\n  extra: 1"), "agent:rollback"},
	}
	for i, s := range steps {
		rv, action, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "rb",
			ResourceVersion: s.rv, Content: s.content}, s.fm, "issue-23")
		want := "update"
		if i == 0 {
			want = "create"
		}
		if err != nil || action != want || rv != int64(i+1) {
			t.Fatalf("第 %d 写: rv=%d action=%s err=%v（期望 rv=%d/%s）", i+1, rv, action, err, i+1, want)
		}
	}
	cur, err := r.Get(ctx, "Agent", "global", "rb")
	gotRV := int64(-1)
	if cur != nil {
		gotRV = cur.ResourceVersion
	}
	if err != nil || gotRV != 4 {
		t.Fatalf("现值应为 rv=4，得 %d err=%v", gotRV, err)
	}
	hs, err := r.DocHistory("Agent", "global", "rb")
	if err != nil {
		t.Fatalf("history: %v", err)
	}
	if len(hs) != 4 {
		t.Fatalf("逐次账应 4 行（rev1-4），得 %d 行——回滚写 rv4 漏账", len(hs))
	}
	if hs[0].Rev != 4 || hs[0].FieldManager != "agent:rollback" {
		t.Fatalf("rev4 行应在册且写者为回滚者，得 rev=%d fm=%q", hs[0].Rev, hs[0].FieldManager)
	}
	seen := map[int64]bool{}
	for _, h := range hs {
		if seen[h.Rev] {
			t.Fatalf("逐次账 rev 重复：%d", h.Rev)
		}
		seen[h.Rev] = true
	}
	for rv := int64(1); rv <= 4; rv++ {
		if !seen[rv] {
			t.Fatalf("逐次账缺 rev=%d 行", rv)
		}
	}
	// 回滚版本正文可读且与现行逐字节一致（账内 rev4 行非占位）。
	c4, err := r.RevContent(ctx, "Agent", "global", "rb", 4)
	if err != nil || c4 != string(cur.Content) {
		t.Fatalf("rev4 正文应与现行一致，err=%v", err)
	}
	if !strings.Contains(c4, "extra: 1") || strings.Contains(c4, "extra: 2") {
		t.Fatalf("rev4 正文应是回滚后的 v2 词形，得 %.80q", c4)
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
