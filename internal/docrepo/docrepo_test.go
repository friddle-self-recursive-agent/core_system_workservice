// docrepo_test.go —— 设计文档受限文件面单测（能力 A）：
// 白名单闸/HEAD 读面/R-PUB-1 lint 收编/乐观锁 409/幂等 noop/工作树脏拒绝/git commit 落地+
// DocApplied 同源+审计行/权限准入/dry-run 零副作用。判据锚：任务书机读判据的 Go 测试核。
package docrepo

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
)

// ── 测试桩 ───────────────────────────────────────────────────────────────────

type fakeBus struct{ events []eventbus.Event }

func (f *fakeBus) Publish(e *eventbus.Event) error { f.events = append(f.events, *e); return nil }

type fakeOps struct{ rows []opslog.ApplyLog }

func (f *fakeOps) AppendApply(l opslog.ApplyLog) error { f.rows = append(f.rows, l); return nil }

type fakePerm struct{ allow bool }

func (f fakePerm) Check(sub permission.Subject, scope, action string, by permission.JudgedBy) error {
	if f.allow {
		return nil
	}
	return errors.New("permission: denied write on core/design-docs")
}

// fixture 建临时 git 仓（本地提交身份注入——不依赖宿主全局配置）。
func fixture(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	git := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		env := append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		cmd.Env = env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v（%s）", args, err, out)
		}
	}
	git("init", "-q")
	git("config", "user.name", "t")
	git("config", "user.email", "t@t")
	for p, c := range files {
		full := filepath.Join(root, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
		git("add", "--", p)
	}
	git("commit", "-q", "-m", "seed")
	return root
}

func engine(allow bool, bus *fakeBus, ops *fakeOps) *Engine {
	return &Engine{Bus: bus, Ops: ops, Perm: fakePerm{allow: allow}, Domain: "core/design-docs"}
}

// ── 白名单 ───────────────────────────────────────────────────────────────────

func TestValidatePath(t *testing.T) {
	ok := []string{"concepts/proposal.md", "a.md", "runtime/console-ui.md"}
	for _, p := range ok {
		if err := validatePath("core_design", p); err != nil {
			t.Errorf("白名单内 %q 误拒: %v", p, err)
		}
	}
	bad := map[string]string{
		"绝对路径":   "/etc/passwd.md",
		"路径穿越":   "../x.md",
		"穿越中段":   "a/../b.md",
		"非规范化":   "./a.md",
		"非 md":   "core/scheme.yaml",
		"git 目录": ".git/config.md",
		"非法字符":   "a\x00b.md",
		"反斜杠":    `a\b.md`,
	}
	for why, p := range bad {
		err := validatePath("core_design", p)
		if !errors.Is(err, ErrForbidden) {
			t.Errorf("%s（%q）未按 403 拒: %v", why, p, err)
		}
	}
	if err := validatePath("core_design", ""); !errors.Is(err, ErrInvalid) {
		t.Errorf("空路径未按 400 拒: %v", err)
	}
}

// ── lint（R-PUB-1 收编） ──────────────────────────────────────────────────────

func TestLint(t *testing.T) {
	if vs := Lint([]byte("# 标题\n\n127.0.0.1 本机回环与 192.0.2.1 文档段是合规示例。\n")); len(vs) != 0 {
		t.Errorf("合规正文误报: %+v", vs)
	}
	cases := []struct {
		content string
		rule    string
	}{
		{"连接 10.1.2.3 可达\n", "ip"},
		{"配置在 /home/friddle/core-docs\n", "path"},
		{"key=DEEPSEEK_API_KEY 形态\n", "ident"},
	}
	for _, c := range cases {
		vs := Lint([]byte(c.content))
		if len(vs) != 1 || vs[0].Rule != c.rule {
			t.Errorf("未按规则 %s 检出: %q → %+v", c.rule, c.content, vs)
		}
		if vs[0].Line != 1 {
			t.Errorf("行号投影错: %+v", vs[0])
		}
	}
}

// ── 读面（权威=HEAD） ─────────────────────────────────────────────────────────

func TestReadAndList(t *testing.T) {
	root := fixture(t, map[string]string{
		"concepts/apply.md":  "# apply\n正文 A\n",
		"concepts/README.md": "# readme\n",
		"core/scheme.yaml":   "apiVersion: core/v1\n",
	})
	s, err := NewSet("core_design=" + root)
	if err != nil {
		t.Fatal(err)
	}
	files, err := s.List()
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 {
		t.Fatalf("清单应只含 .md（实得 %d）: %+v", len(files), files)
	}
	if files[0].Path != "concepts/README.md" || files[0].Repo != "core_design" {
		t.Errorf("清单排序/归属错: %+v", files[0])
	}
	d, err := s.Read("core_design", "concepts/apply.md")
	if err != nil {
		t.Fatal(err)
	}
	if d.Sha == "" || len(d.Content) == 0 || d.ContentStr != "# apply\n正文 A\n" {
		t.Errorf("读面投影错: %+v", d)
	}
	if d.LastCommit == nil || d.LastCommit.Subject != "seed" || d.LastCommit.At.IsZero() {
		t.Errorf("最近提交投影缺席: %+v", d.LastCommit)
	}
	if _, err := s.Read("core_design", "core/scheme.yaml"); !errors.Is(err, ErrForbidden) {
		t.Errorf("非白名单读未拒 403 语义: %v", err)
	}
	if _, err := s.Read("other", "a.md"); !errors.Is(err, ErrNoRepo) {
		t.Errorf("未注册仓读未拒: %v", err)
	}
	if _, err := s.Read("core_design", "nope.md"); !errors.Is(err, ErrNotFound) {
		t.Errorf("HEAD 无此文件未拒 404 语义: %v", err)
	}
}

// ── Apply 全周期 ─────────────────────────────────────────────────────────────

func TestApplyLifecycle(t *testing.T) {
	root := fixture(t, map[string]string{"concepts/proposal.md": "# 提案\n"})
	s, _ := NewSet("core_design=" + root)
	bus, ops := &fakeBus{}, &fakeOps{}
	s.BindEngine(engine(true, bus, ops))
	ctx := context.Background()
	d, _ := s.Read("core_design", "concepts/proposal.md")

	// ① create（新文件）：提交在账+事件同源+审计行。
	res, err := s.Apply(ctx, ApplyRequest{Repo: "core_design", Path: "concepts/probe.md",
		Content: []byte("# probe\n"), BaseSha: "", FieldManager: "console:editor", ReasonRef: "console:apply",
		Subject: &permission.Subject{User: "probe"}})
	if err != nil {
		t.Fatalf("create: %v", err)
	}
	if res.Action != "create" || res.Commit == nil || res.Sha == "" {
		t.Fatalf("create 结果投影错: %+v", res)
	}
	if len(bus.events) != 1 || bus.events[0].Type != "DocApplied" ||
		bus.events[0].Involved != "DesignDoc/core_design/concepts/probe.md" ||
		!strings.HasPrefix(bus.events[0].DedupKey, "apply:DesignDoc/core_design/concepts/probe.md:") {
		t.Fatalf("DocApplied 同源判据不达: %+v", bus.events)
	}
	if len(ops.rows) != 1 || ops.rows[0].Result != "ok" || ops.rows[0].Obj != "DesignDoc/core_design/concepts/probe.md" {
		t.Fatalf("ok 审计行不达: %+v", ops.rows)
	}
	d2, _ := s.Read("core_design", "concepts/probe.md")
	if d2.ContentStr != "# probe\n" {
		t.Fatalf("读回正文不达: %q", d2.ContentStr)
	}

	// ② update（基线 sha 一致）：第二提交。
	res2, err := s.Apply(ctx, ApplyRequest{Repo: "core_design", Path: "concepts/probe.md",
		Content: []byte("# probe\n改\n"), BaseSha: d2.Sha, FieldManager: "console:editor",
		Subject: &permission.Subject{User: "probe"}})
	if err != nil {
		t.Fatalf("update: %v", err)
	}
	if res2.Action != "update" || res2.Commit == nil || res2.Commit.Sha == res.Commit.Sha {
		t.Fatalf("update 结果投影错: %+v", res2)
	}
	if len(bus.events) != 2 || len(ops.rows) != 2 {
		t.Fatalf("update 审计/事件计数不达: %+v %+v", bus.events, ops.rows)
	}

	// ③ noop 幂等：同内容重入零提交零事件（noop 行照落）。
	before := commitCount(t, root)
	res3, err := s.Apply(ctx, ApplyRequest{Repo: "core_design", Path: "concepts/probe.md",
		Content: []byte("# probe\n改\n"), BaseSha: res2.Sha, FieldManager: "console:editor",
		Subject: &permission.Subject{User: "probe"}})
	if err != nil {
		t.Fatalf("noop: %v", err)
	}
	if res3.Action != "noop" || res3.Commit != nil || res3.Event != nil {
		t.Fatalf("noop 结果投影错: %+v", res3)
	}
	if commitCount(t, root) != before {
		t.Fatalf("noop 产生了提交")
	}
	if len(bus.events) != 2 {
		t.Fatalf("noop 发了事件")
	}

	// ④ 乐观锁：陈旧基线 → ConflictError（409 形载荷 expected/current）。
	_, err = s.Apply(ctx, ApplyRequest{Repo: "core_design", Path: "concepts/probe.md",
		Content: []byte("# probe\n陈旧\n"), BaseSha: d.Sha, FieldManager: "console:editor",
		Subject: &permission.Subject{User: "probe"}})
	var ce *ConflictError
	if !errors.As(err, &ce) || ce.Current != res2.Sha || ce.Expected != d.Sha {
		t.Fatalf("乐观锁 409 形不达: %v", err)
	}

	// ⑤ lint 未过 → 400 形携违规（不落库不提交）。
	before = commitCount(t, root)
	_, err = s.Apply(ctx, ApplyRequest{Repo: "core_design", Path: "concepts/probe.md",
		Content: []byte("# probe\n去 10.9.8.7\n"), BaseSha: res2.Sha, FieldManager: "console:editor",
		Subject: &permission.Subject{User: "probe"}})
	var le *LintError
	if !errors.As(err, &le) || len(le.Violations) != 1 || le.Violations[0].Rule != "ip" {
		t.Fatalf("lint 400 形不达: %v", err)
	}
	if commitCount(t, root) != before {
		t.Fatalf("lint 未过仍落了提交")
	}

	// ⑥ 越权路径 → 403 语义；未注册仓 → 404 语义。
	_, err = s.Apply(ctx, ApplyRequest{Repo: "core_design", Path: "core/scheme.yaml",
		Content: []byte("x\n"), BaseSha: "", FieldManager: "f", Subject: &permission.Subject{User: "probe"}})
	if !errors.Is(err, ErrForbidden) {
		t.Fatalf("非白名单路径未拒 403: %v", err)
	}
	_, err = s.Apply(ctx, ApplyRequest{Repo: "nope", Path: "a.md",
		Content: []byte("x\n"), BaseSha: "", FieldManager: "f", Subject: &permission.Subject{User: "probe"}})
	if !errors.Is(err, ErrNoRepo) {
		t.Fatalf("未注册仓未拒: %v", err)
	}

	// ⑦ 工作树未提交改动 → ErrDirty（不吞他人未提交编辑）。
	if err := os.WriteFile(filepath.Join(root, "concepts", "probe.md"), []byte("# probe\n在途\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	_, err = s.Apply(ctx, ApplyRequest{Repo: "core_design", Path: "concepts/probe.md",
		Content: []byte("# probe\n覆盖\n"), BaseSha: res2.Sha, FieldManager: "console:editor",
		Subject: &permission.Subject{User: "probe"}})
	if !errors.Is(err, ErrDirty) {
		t.Fatalf("脏工作树未拒: %v", err)
	}
}

func TestApplyDeniedAndDryRun(t *testing.T) {
	root := fixture(t, map[string]string{"a.md": "# a\n"})
	s, _ := NewSet("r=" + root)
	bus, ops := &fakeBus{}, &fakeOps{}
	s.BindEngine(engine(false, bus, ops)) // 权限面拒绝
	_, err := s.Apply(context.Background(), ApplyRequest{Repo: "r", Path: "b.md",
		Content: []byte("# b\n"), BaseSha: "", FieldManager: "f",
		Subject: &permission.Subject{User: "probe"}})
	if !errors.Is(err, ErrDenied) {
		t.Fatalf("权限拒绝未达: %v", err)
	}
	if len(ops.rows) != 1 || ops.rows[0].Result != "denied" || ops.rows[0].Subject != "probe" {
		t.Fatalf("denied 审计行不达（携主体标识）: %+v", ops.rows)
	}
	if len(bus.events) != 0 {
		t.Fatalf("拒绝路径发了事件")
	}
	// dry-run：零副作用（无提交无事件无 ok 行）——准入结论一致。
	ops.rows = nil
	s.BindEngine(engine(true, bus, ops))
	res, err := s.Apply(context.Background(), ApplyRequest{Repo: "r", Path: "b.md",
		Content: []byte("# b\n"), BaseSha: "", FieldManager: "f", DryRun: true,
		Subject: &permission.Subject{User: "probe"}})
	if err != nil {
		t.Fatalf("dry-run: %v", err)
	}
	if res.Action != "dry-run" || res.Commit != nil || res.Event != nil || len(ops.rows) != 0 {
		t.Fatalf("dry-run 副作用面不净: %+v %+v", res, ops.rows)
	}
	// 无主写入拒绝。
	if _, err := s.Apply(context.Background(), ApplyRequest{Repo: "r", Path: "b.md",
		Content: []byte("# b\n"), BaseSha: ""}); !errors.Is(err, ErrNoFieldManager) {
		t.Fatalf("无主写入未拒: %v", err)
	}
	// 引擎未装配 → 405 语义。
	s2, _ := NewSet("r=" + root)
	if _, err := s2.Apply(context.Background(), ApplyRequest{Repo: "r", Path: "b.md",
		Content: []byte("# b\n"), BaseSha: "", FieldManager: "f", Subject: &permission.Subject{}}); !errors.Is(err, ErrNoEngine) {
		t.Fatalf("引擎缺席未拒: %v", err)
	}
}

func TestNewSetValidation(t *testing.T) {
	root := fixture(t, map[string]string{"a.md": "x\n"})
	if _, err := NewSet("bad"); err == nil {
		t.Fatal("缺 path 规格未拒")
	}
	if _, err := NewSet("draft=" + root); err == nil {
		t.Fatal("保留仓名未拒")
	}
	nogit := t.TempDir()
	if _, err := NewSet("x=" + nogit); err == nil {
		t.Fatal("非 git 仓根未拒")
	}
	s, err := NewSet("core_design="+root, "  ")
	if err != nil || len(s.Names()) != 1 || s.Names()[0] != "core_design" {
		t.Fatalf("正常装配不达: %v %+v", err, s.Names())
	}
}

func commitCount(t *testing.T, root string) int {
	t.Helper()
	out, err := r0git(root, "rev-list", "--count", "HEAD")
	if err != nil {
		t.Fatal(err)
	}
	n := 0
	for _, c := range out {
		if c < '0' || c > '9' {
			break
		}
		n = n*10 + int(c-'0')
	}
	return n
}

var _ = time.Now
