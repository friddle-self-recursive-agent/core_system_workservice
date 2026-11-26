// rev_content_test.go —— 历史 rv 正文快照读面验收（v0.63 rev-content diff 面，console-adoption.md
// §1.40——历史 tab「点行 diff 上一版」余量收口读半边）：rv→commit 机械对账（提交消息键段+blob
// resourceVersion 反写行双段）、点行 diff 上一版判据（旧 rv 快照可读且与新 rv 分立）、noop 零提交、
// 手编 docsRoot 通道（无 rv 反写行提交）诚实缺席、对象键过滤。
package repository

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
)

func mustRevContent(t *testing.T, r *GitRepo, kind, scope, name string, rv int64) string {
	t.Helper()
	out, err := r.RevContent(context.Background(), kind, scope, name, rv)
	if err != nil {
		t.Fatalf("RevContent(%s/%s/%s r%d): %v", kind, scope, name, rv, err)
	}
	return out
}

func TestRevContentSnapshotAcrossRevisions(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	const n = "hist-probe"
	// rv1=create
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Name: n, Scope: "global", ResourceVersion: 0,
		Content: agentYAML(n, "")}, "agent:probe", "bootstrap"); err != nil {
		t.Fatalf("rv1 create: %v", err)
	}
	got1, err := r.Get(ctx, "Agent", "global", n)
	if err != nil {
		t.Fatal(err)
	}
	want1 := string(got1.Content)
	// rv2=update（语义差异：autoLevel 词形变化——canonical 比对必判 update）
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Name: n, Scope: "global", ResourceVersion: 1,
		Content: agentYAML(n, "-v2")}, "console:editor", "edit"); err != nil {
		t.Fatalf("rv2 update: %v", err)
	}
	got2, err := r.Get(ctx, "Agent", "global", n)
	if err != nil {
		t.Fatal(err)
	}
	want2 := string(got2.Content)

	// L40-① rv1 快照=create 时点（点行 diff 上一版判据——旧 rv 正文可读）
	rev1 := mustRevContent(t, r, "Agent", "global", n, 1)
	if rev1 != want1 {
		t.Errorf("rv1 快照应与 create 时点面一致（含 rv 反写行）， got %.80q want %.80q", rev1, want1)
	}
	// L40-② rv2 快照=update 时点且与 rv1 分立（diff 面可查）
	rev2 := mustRevContent(t, r, "Agent", "global", n, 2)
	if rev2 != want2 {
		t.Errorf("rv2 快照应与现面一致，得 %.80q want %.80q", rev2, want2)
	}
	if rev1 == rev2 {
		t.Errorf("rv1/rv2 快照应分立（diff 面可查）")
	}
	if !strings.Contains(rev2, "auto-v2") || strings.Contains(rev1, "auto-v2") {
		t.Errorf("rv 词形应随提交时点分立：rv1 无 v2 词形、rv2 有——rev1=%q rev2=%q", rev1, rev2)
	}
	// L40-③ noop 零提交：同内容重写 → rv 恒=2，rv=3 诚实缺席
	if rv, action, err := r.Write(ctx, &Object{Kind: "Agent", Name: n, Scope: "global", ResourceVersion: 2,
		Content: agentYAML(n, "-v2")}, "console:editor", "replay"); err != nil || action != "noop" || rv != 2 {
		t.Fatalf("noop 复跑应 rv=2 noop，得 rv=%d action=%s err=%v", rv, action, err)
	}
	if _, err := r.RevContent(ctx, "Agent", "global", n, 3); err == nil || !strings.Contains(err.Error(), "快照缺席") {
		t.Errorf("rv=3（无落库提交）应 ErrRevNotFound 词形，得 %v", err)
	}
	// L40-④ 手编 docsRoot 通道（无 rv 反写行前进的提交）：rv=3 恒缺席（rv 反写行判据——
	// 外来提交不改 blob rv，任何 rv 都不误配）
	k, err := mustScheme(t).Lookup("Agent")
	if err != nil {
		t.Fatal(err)
	}
	path, err := r.concretePath(k, "global", n)
	if err != nil {
		t.Fatal(err)
	}
	manual := strings.Replace(want2, "auto-v2", "auto-manual", 1) // rv 反写行保持 2
	if err := os.WriteFile(path, []byte(manual), 0o644); err != nil {
		t.Fatal(err)
	}
	gitEnv := append(os.Environ(), "GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t", "GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
	cmd := exec.Command("git", "-C", r.Root(), "add", "-A")
	cmd.Env = gitEnv
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git add: %v (%s)", err, out)
	}
	cmd = exec.Command("git", "-C", r.Root(), "-c", "commit.gpgsign=false", "commit", "-m", "manual edit outside cored")
	cmd.Env = gitEnv
	if out, err := cmd.CombinedOutput(); err != nil {
		t.Fatalf("git commit: %v (%s)", err, out)
	}
	// 外来提交不改 blob rv → rv=2 快照回退到最新携带 rv=2 的提交（=手编正文，诚实词形）
	if got := mustRevContent(t, r, "Agent", "global", n, 2); !strings.Contains(got, "auto-manual") {
		t.Errorf("手编提交后 rv=2 快照应=最新携带 rv2 的提交（手编正文），得 %.80q", got)
	}
	if _, err := r.RevContent(ctx, "Agent", "global", n, 3); err == nil {
		t.Errorf("外来提交零 rv 前进：rv=3 应仍缺席，得快照（判据失守）")
	}
	// L40-⑤ 后续 ok apply → rv=3 可读（链条恢复）
	if rv, action, err := r.Write(ctx, &Object{Kind: "Agent", Name: n, Scope: "global", ResourceVersion: 2,
		Content: agentYAML(n, "-v3")}, "console:editor", "edit3"); err != nil || action != "update" || rv != 3 {
		t.Fatalf("手编后 ok apply 应 update rv=3，得 rv=%d action=%s err=%v", rv, action, err)
	}
	rev3 := mustRevContent(t, r, "Agent", "global", n, 3)
	if !strings.Contains(rev3, "auto-v3") {
		t.Errorf("rv3 快照应含 v3 词形，得 %.80q", rev3)
	}
	rev2b := mustRevContent(t, r, "Agent", "global", n, 2)
	if !strings.Contains(rev2b, "auto-manual") {
		t.Errorf("rv2 快照应仍=手编时点（历史不可变），得 %.80q", rev2b)
	}
}

func TestRevContentKeyFilterAndErrors(t *testing.T) {
	r := newRepo(t)
	ctx := context.Background()
	// 两个对象各自提交——键过滤互不串
	for _, n := range []string{"hist-a", "hist-b"} {
		if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Name: n, Scope: "global", ResourceVersion: 0,
			Content: agentYAML(n, "")}, "agent:probe", "bootstrap"); err != nil {
			t.Fatal(err)
		}
	}
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Name: "hist-b", Scope: "global", ResourceVersion: 1,
		Content: agentYAML("hist-b", "-x")}, "console:editor", "edit"); err != nil {
		t.Fatal(err)
	}
	revA := mustRevContent(t, r, "Agent", "global", "hist-a", 1)
	if !strings.Contains(revA, "name: hist-a") {
		t.Errorf("hist-a 快照不应被 hist-b 提交串染，得 %.80q", revA)
	}
	// 未注册 kind / scope 缺席 / 坏对象名 → 既有 Err 词形原样透传（HTTP 面映射 404/400）
	if _, err := r.RevContent(ctx, "NoKind", "global", "x", 1); err == nil || !strings.Contains(err.Error(), "未注册") {
		t.Errorf("未注册 kind 应 ErrNotRegistered 词形，得 %v", err)
	}
	if _, err := r.RevContent(ctx, "Agent", "", "x", 1); err == nil || !strings.Contains(err.Error(), "显式 scope") {
		t.Errorf("scope 缺席应 ErrScope 词形，得 %v", err)
	}
	if _, err := r.RevContent(ctx, "Agent", "global", "X_Bad", 1); err == nil || !strings.Contains(err.Error(), "命名规则") {
		t.Errorf("坏对象名应 ErrScope 词形，得 %v", err)
	}
	// 从未落库的对象 → ErrRevNotFound（非对象缺席 ErrNotFound——快照域分立词形）
	_, err := r.RevContent(ctx, "Agent", "global", "never-was", 1)
	if err == nil || !strings.Contains(err.Error(), "快照缺席") {
		t.Errorf("从未落库对象应 ErrRevNotFound 词形，得 %v", err)
	}
}
