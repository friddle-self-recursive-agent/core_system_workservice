// export_test.go —— sqlite→git 导出机制验收（R-DOC-7 规则 2：git=导出/公开 output/public——
// 2026-11-28 用户裁决后=release 节点 op 执行机制，非常驻定时器）：真 git 仓（t.TempDir）上
// 验证 commit 产生与内容字节恒等、消息带对象数+rv、同水位幂等零提交、陈旧文件删除、sidecar
// 随行保留、过滤=scheme docTypes 声明派生。不测 push（远端属部署侧——remote 未配置=本地
// commit 止步，no force-push 语义在部署侧凭据面）。
package repository

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

func newExporter(t *testing.T, r *SQLiteRepo, dir string) *Exporter {
	t.Helper()
	e, err := NewExporter(r, mustScheme(t), dir, "")
	if err != nil {
		t.Fatalf("NewExporter: %v", err)
	}
	return e
}

// gitIn 在导出工作树执行 git 子命令（测试侧手工摆件用——身份注入与导出器同范式）。
func gitIn(t *testing.T, dir string, args ...string) string {
	t.Helper()
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=tester", "GIT_AUTHOR_EMAIL=tester@localhost",
		"GIT_COMMITTER_NAME=tester", "GIT_COMMITTER_EMAIL=tester@localhost",
	)
	var out, errOut strings.Builder
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		t.Fatalf("git %s: %v（%s）", strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}
	return strings.TrimSpace(out.String())
}

func commitCountIn(t *testing.T, dir string) int {
	t.Helper()
	out, err := gitInOut(dir, "rev-list", "--count", "HEAD")
	if err != nil {
		t.Fatalf("rev-list: %v（空仓尚未有提交属测试时序错误）", err)
	}
	var n int
	_, _ = fmt.Sscanf(out, "%d", &n)
	return n
}

// gitInOut 同 gitIn 但容错返回（空仓 rev-parse 场景）。
func gitInOut(dir string, args ...string) (string, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=tester", "GIT_AUTHOR_EMAIL=tester@localhost",
		"GIT_COMMITTER_NAME=tester", "GIT_COMMITTER_EMAIL=tester@localhost",
	)
	var out, errOut strings.Builder
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("git %s: %w（%s）", strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

func TestExporterCommitContentAndIdempotent(t *testing.T) {
	r, _ := newSQLiteRepo(t)
	ctx := context.Background()
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "exp-a", Content: agentYAML("exp-a", "")}, "fm", ""); err != nil {
		t.Fatal(err)
	}
	w1 := []byte("apiVersion: core/v1\nkind: Workflow\nmetadata:\n  name: exp-w\n  scope: global\nspec:\n  steps: []\n")
	if _, _, err := r.Write(ctx, &Object{Kind: "Workflow", Scope: "global", Name: "exp-w", Content: w1}, "fm", ""); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	e := newExporter(t, r, dir)

	// 首拍：全量快照落树 + 一个 commit（消息带对象数+rv）。
	res, err := e.Export(ctx)
	if err != nil {
		t.Fatalf("export: %v", err)
	}
	if !res.Committed || res.Objects != 2 || res.HighWater != 1 || res.Pushed {
		t.Fatalf("首拍投影: %+v", res)
	}
	// 文件字节=DB content 字节（逐字节，不重排）。
	o, err := r.Get(ctx, "Agent", "global", "exp-a")
	if err != nil {
		t.Fatal(err)
	}
	got, err := os.ReadFile(filepath.Join(dir, "global", "agents", "exp-a", "agent.yaml"))
	if err != nil || string(got) != string(o.Content) {
		t.Fatalf("导出文件应与 doc.content 逐字节恒等，err=%v", err)
	}
	ow, err := r.Get(ctx, "Workflow", "global", "exp-w")
	if err != nil {
		t.Fatal(err)
	}
	gotw, err := os.ReadFile(filepath.Join(dir, "global", "workflows", "exp-w.yaml"))
	if err != nil || string(gotw) != string(ow.Content) {
		t.Fatalf("嵌套布局导出应恒等，err=%v", err)
	}
	msg := gitIn(t, dir, "log", "-1", "--format=%s")
	if !strings.HasPrefix(msg, "export: ") || !strings.Contains(msg, "rv=1") || !strings.Contains(msg, "objects=2") {
		t.Fatalf("提交消息应 export: <ts> rv=1 objects=2，得 %q", msg)
	}
	// 幂等：同水位重复导出=零提交（判据 docs-export.md §6）。
	res2, err := e.Export(ctx)
	if err != nil || res2.Committed {
		t.Fatalf("同水位应零提交，得 %+v err=%v", res2, err)
	}
	if n := commitCountIn(t, dir); n != 1 {
		t.Fatalf("幂等后 commit 数应 1，得 %d", n)
	}
	// 内容翻页 → 新 commit，消息 rv=2；工作树文件随之更新。
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "exp-a",
		ResourceVersion: 1, Content: agentYAML("exp-a", "\n  extra: 1")}, "fm", ""); err != nil {
		t.Fatal(err)
	}
	res3, err := e.Export(ctx)
	if err != nil || !res3.Committed || res3.HighWater != 2 {
		t.Fatalf("翻页拍应提交 rv=2，得 %+v err=%v", res3, err)
	}
	msg3 := gitIn(t, dir, "log", "-1", "--format=%s")
	if !strings.Contains(msg3, "rv=2") {
		t.Fatalf("翻页提交消息应 rv=2，得 %q", msg3)
	}
	o2, _ := r.Get(ctx, "Agent", "global", "exp-a")
	got2, _ := os.ReadFile(filepath.Join(dir, "global", "agents", "exp-a", "agent.yaml"))
	if string(got2) != string(o2.Content) {
		t.Fatal("翻页后导出文件应随行更新")
	}
}

func TestExporterPruneStaleAndKeepSidecar(t *testing.T) {
	r, _ := newSQLiteRepo(t)
	ctx := context.Background()
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "side-a", Content: agentYAML("side-a", "")}, "fm", ""); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	e := newExporter(t, r, dir)
	if _, err := e.Export(ctx); err != nil {
		t.Fatal(err)
	}
	// 侧车随行：对象旁手工落 `<文件>.metadatainfo`（生成逻辑在 DocIndex/打标面，不在本仓——
	// 导出侧只随行保留不发明；R-DOC-6/7 同目录同提交）并独立提交。
	sidecar := filepath.Join(dir, "global", "agents", "side-a", "agent.yaml.metadatainfo")
	if err := os.WriteFile(sidecar, []byte("doc: global/agents/side-a/agent.yaml\ntags: [{tag: doc, tag_type: doc}]\nimportance: 50\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, dir, "add", "-A")
	gitIn(t, dir, "-c", "commit.gpgsign=false", "commit", "-m", "test: sidecar 手工落件")
	// DB 外陈旧文件（上一版残留——先随手工提交入库，导出拍删除才产生 git 变更；
	// 未追踪野文件删除不产生 commit=git 语义，不是导出器缺口）。
	stale := filepath.Join(dir, "stale.yaml")
	if err := os.WriteFile(stale, []byte("stale: true\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	gitIn(t, dir, "add", "-A")
	gitIn(t, dir, "-c", "commit.gpgsign=false", "commit", "-m", "test: 上一版残留入库")
	res, err := e.Export(ctx)
	if err != nil || !res.Committed {
		t.Fatalf("清理拍应提交，得 %+v err=%v", res, err)
	}
	if _, err := os.Stat(stale); !os.IsNotExist(err) {
		t.Fatalf("陈旧文件应删除，err=%v", err)
	}
	if _, err := os.Stat(sidecar); err != nil {
		t.Fatalf("在册对象 sidecar 应随行保留，err=%v", err)
	}
	// 清理后水位稳定=零提交。
	res2, err := e.Export(ctx)
	if err != nil || res2.Committed {
		t.Fatalf("清理后应幂等零提交，得 %+v err=%v", res2, err)
	}
}

func TestNewExporterGuards(t *testing.T) {
	r, _ := newSQLiteRepo(t)
	if _, err := NewExporter(nil, mustScheme(t), t.TempDir(), ""); err == nil {
		t.Fatal("缺文档库应拒绝")
	}
	if _, err := NewExporter(r, nil, t.TempDir(), ""); err == nil {
		t.Fatal("缺 scheme 应拒绝")
	}
	if _, err := NewExporter(r, mustScheme(t), "  ", ""); err == nil {
		t.Fatal("缺目标目录应拒绝")
	}
}

func TestExportExcludesRuntimeDocTypesFromScheme(t *testing.T) {
	// 2026-11-28 用户裁决：docs_export_config 独立配置面已删——导出过滤=scheme docTypes 声明
	// 唯一真相源（category=runtime 缺省排除，零运行期可写口）。
	ctx := context.Background()
	sch := mustSchemeWithRuntimeDocType(t, "BugDoc")
	ops := &memOps{}
	r, err := OpenSQLite(t.TempDir()+"/docs.db", sch, &Options{Ops: ops}, "")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = r.Close() })
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "design-a", Content: agentYAML("design-a", "")}, "fm", ""); err != nil {
		t.Fatal(err)
	}
	bugContent := []byte("apiVersion: core/v1\nkind: Agent\nmetadata:\n  name: bug-temp\n  scope: global\n  docType: BugDoc\nspec:\n  role: r\n")
	if _, _, err := r.Write(ctx, &Object{Kind: "Agent", Scope: "global", Name: "bug-temp", Content: bugContent}, "fm", ""); err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	exp, err := NewExporter(r, sch, dir, "")
	if err != nil {
		t.Fatal(err)
	}
	res, err := exp.Export(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if res.Objects != 1 {
		t.Fatalf("category=runtime docType 应按 scheme 声明排除（只导出 1 个对象），得 %+v", res)
	}
}

// mustSchemeWithRuntimeDocType 测试 scheme（testScheme 基础+docTypes 声明一条 category=runtime
// 类型——scheme 派生导出过滤的判据面）。
func mustSchemeWithRuntimeDocType(t *testing.T, name string) *scheme.Scheme {
	t.Helper()
	base := strings.Replace(testScheme, "spec:\n", "spec:\n  docTypes:\n    - {name: "+name+", category: runtime}\n", 1)
	s, err := scheme.Load([]byte(base))
	if err != nil {
		t.Fatalf("载入测试 scheme: %v", err)
	}
	return s
}
