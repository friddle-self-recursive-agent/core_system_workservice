// docsdocs_test.go —— connectops-docs 验收（2026-11-28 用户裁决「做成workflow的一个节点。
// 即gh-release包括git push/release[scheme调用]」）：ApplyConfig 声明注入、dir 缺席=显式缺口、
// git-push 真拍（真 sqlite docs 库+真 git 树——水位无变化幂等零提交）、release-publish 委托
// connectops-gh 缺席=显式失败不静默。
package docsdocs

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/connectors"

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

// fakeGH 日 tag 策略测试替身（connectops-gh 词形——release-list 在册/release-create 记账）。
type fakeGH struct {
	existing []string
	created  []string
}

func (f *fakeGH) Type() string                    { return "connectops-gh" }
func (f *fakeGH) Version() string                 { return "test" }
func (f *fakeGH) Ops() []string                   { return []string{"release-list", "release-create"} }
func (f *fakeGH) Probe(ctx context.Context) error { return nil }
func (f *fakeGH) Invoke(ctx context.Context, op string, params []byte) ([]byte, error) {
	var q struct {
		Tag string `json:"tag"`
	}
	_ = json.Unmarshal(params, &q)
	switch op {
	case "release-list":
		out := make([]map[string]any, 0, len(f.existing))
		for _, t := range f.existing {
			out = append(out, map[string]any{"tag_name": t})
		}
		return json.Marshal(out)
	case "release-create":
		f.created = append(f.created, q.Tag)
		return json.Marshal(map[string]any{"tag_name": q.Tag})
	}
	return nil, fmt.Errorf("op %s 未路由", op)
}

// 日 tag 策略（2026-11-28 用户裁决「默认1天做多1个」）：tag 缺席派生 v<UTC 日期>——
// 当日已在册=skipped 不冒充新发版；不在册=release-create 真发（派生 tag 回填）。
func TestReleasePublishDailyTagPolicy(t *testing.T) {
	ctx := context.Background()
	sch := mustSch(t)
	today := "v" + time.Now().UTC().Format("2006.01.02")
	// ①当日已在册 → skipped。
	f1 := &fakeGH{existing: []string{today, "v0.1.0"}}
	connectors.RegisterOpsConnector(f1)
	c1 := New(newRepo(t, sch), sch)
	if err := c1.ApplyConfig([]byte(`{"dir": "` + t.TempDir() + `"}`)); err != nil {
		t.Fatal(err)
	}
	res, err := c1.Invoke(ctx, "release-publish", []byte(`{"owner":"o","repo":"r"}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(res), `"skipped"`) || !strings.Contains(string(res), today) {
		t.Fatalf("当日已发应 skipped 注记含当日 tag: %s", res)
	}
	if len(f1.created) != 0 {
		t.Fatalf("当日已发不应 create: %v", f1.created)
	}
	// ②当日未发 → 派生 tag 真发。
	f2 := &fakeGH{existing: []string{"v0.1.0"}}
	connectors.RegisterOpsConnector(f2)
	c2 := New(newRepo(t, sch), sch)
	if err := c2.ApplyConfig([]byte(`{"dir": "` + t.TempDir() + `"}`)); err != nil {
		t.Fatal(err)
	}
	res2, err := c2.Invoke(ctx, "release-publish", []byte(`{"owner":"o","repo":"r"}`))
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(string(res2), today) || len(f2.created) != 1 || f2.created[0] != today {
		t.Fatalf("当日未发应派生 tag 真发: res=%s created=%v", res2, f2.created)
	}
}

// gh 委托缺席 + tag 缺席 → 显式失败（日检查与发版都不可执行——不静默不编造）。
func TestReleasePublishTagAbsentWithoutGhFailsExplicit(t *testing.T) {
	ctx := context.Background()
	sch := mustSch(t)
	c := New(newRepo(t, sch), sch)
	if err := c.ApplyConfig([]byte(`{"dir": "` + t.TempDir() + `"}`)); err != nil {
		t.Fatal(err)
	}
	if _, err := c.Invoke(ctx, "release-publish", []byte(`{"repo":"core_system_workservice"}`)); err == nil {
		t.Fatal("gh 委托缺席应显式失败")
	}
}
