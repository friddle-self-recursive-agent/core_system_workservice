// docsdocs_test.go —— connectops-docs 验收（2026-11-28 用户裁决「做成workflow的一个节点。
// 即gh-release包括git push/release[scheme调用]」）：ApplyConfig 声明注入、dir 缺席=显式缺口、
// git-push 真拍（真 sqlite docs 库+真 git 树——水位无变化幂等零提交）、release-publish 委托
// connectops-gh 缺席=显式失败不静默。
package docsdocs

import (
	"context"
	"strings"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

const testScheme = `apiVersion: core/v1
kind: Scheme
metadata: {name: test, scope: builtin}
spec:
  version: core/v1
  kinds:
    - {kind: Agent, builtin: true, storage: global/agents/<name>/agent.yaml}
    - {kind: Scheme, builtin: true, storage: core/scheme.yaml}
`

func mustSch(t *testing.T) *scheme.Scheme {
	t.Helper()
	s, err := scheme.Load([]byte(testScheme))
	if err != nil {
		t.Fatalf("载入测试 scheme: %v", err)
	}
	return s
}

func newRepo(t *testing.T, sch *scheme.Scheme) repository.Repository {
	t.Helper()
	r, err := repository.OpenSQLite(t.TempDir()+"/docs.db", sch, &repository.Options{}, "")
	if err != nil {
		t.Fatalf("OpenSQLite: %v", err)
	}
	t.Cleanup(func() { _ = r.Close() })
	return r
}

func TestApplyConfigAndGitPush(t *testing.T) {
	ctx := context.Background()
	sch := mustSch(t)
	repo := newRepo(t, sch)
	if _, _, err := repo.Write(ctx, &repository.Object{Kind: "Agent", Scope: "global", Name: "a1", Content: []byte("apiVersion: core/v1\nkind: Agent\nmetadata:\n  name: a1\n  scope: global\nspec:\n  mode: chat\n")}, "fm", ""); err != nil {
		t.Fatal(err)
	}
	c := New(repo, sch)
	// dir 缺席=显式缺口（Probe/Invoke 照实报错不猜测目标树）。
	if err := c.Probe(ctx); err == nil {
		t.Fatal("dir 缺席 Probe 应显式报错")
	}
	if err := c.ApplyConfig([]byte(`{"dir": "` + t.TempDir() + `"}`)); err != nil {
		t.Fatal(err)
	}
	if err := c.Probe(ctx); err != nil {
		t.Fatalf("dir 就绪后 Probe 应通过: %v", err)
	}
	res, err := c.Invoke(ctx, "git-push", nil)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(res), `"Committed":true`) {
		t.Fatalf("首拍应产生提交: %s", res)
	}
	// 幂等：水位无变化=零提交。
	res2, err := c.Invoke(ctx, "git-push", nil)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(res2), `"Committed":true`) {
		t.Fatalf("同水位第二拍应零提交: %s", res2)
	}
	// op 不在声明集=显式错误。
	if _, err := c.Invoke(ctx, "nope", nil); err == nil {
		t.Fatal("未声明 op 应显式拒绝")
	}
}

func TestReleasePublishGhAbsentFailsExplicit(t *testing.T) {
	ctx := context.Background()
	sch := mustSch(t)
	c := New(newRepo(t, sch), sch)
	if err := c.ApplyConfig([]byte(`{"dir": "` + t.TempDir() + `"}`)); err != nil {
		t.Fatal(err)
	}
	// connectops-gh 未注册（测试进程未导入驱动包）=显式失败不静默。
	if _, err := c.Invoke(ctx, "release-publish", []byte(`{"owner":"o","repo":"r","tag":"v1"}`)); err == nil {
		t.Fatal("gh 委托缺席应显式失败")
	}
}

func TestReleasePublishTagAbsentSkips(t *testing.T) {
	ctx := context.Background()
	sch := mustSch(t)
	c := New(newRepo(t, sch), sch)
	if err := c.ApplyConfig([]byte(`{"dir": "` + t.TempDir() + `"}`)); err != nil {
		t.Fatal(err)
	}
	// tag 未声明=release 跳过（docs push 照走、显式注记不冒充——不调用 gh 委托）。
	res, err := c.Invoke(ctx, "release-publish", []byte(`{"repo":"core_system_workservice"}`))
	if err != nil {
		t.Fatalf("tag 缺席应跳过而非失败: %v", err)
	}
	if !strings.Contains(string(res), `"skipped"`) {
		t.Fatalf("skip 注记缺席: %s", res)
	}
}
