// delete_reapply_test.go —— 探针复现锚（v0.55 R-9 轮实证面）：对象删除（git rm 带外清场）后
// DocApplied 热路径 Ensure 的语义：删除会话后 List=空、Ensure 走创建分支=created（第二次
// created 派单=行为面如实——跑该判据固定探针观察，不臆断）。
package flowbind

import (
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/apply"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

const delFixture = `apiVersion: core/v1
kind: Scheme
metadata: {name: default, scope: builtin}
spec:
  version: core/v1
  kinds:
    - {kind: Workflow, builtin: true, storage: "global/workflows/<name>.yaml | systems/<s>/workflows/<name>.yaml"}
    - {kind: ConnectorOps, builtin: true, storage: "global/connector-ops/<name>.yaml | systems/<s>/connector-ops/<name>.yaml"}
  eventTypes:
    - {type: DocApplied, source: core, defaultLevel: info, route: owner-workflow}
`

func delWfDoc(name string, phases []string) []byte {
	out := "apiVersion: core/v1\nkind: Workflow\nmetadata: {name: " + name + ", scope: global}\nspec:\n  phases:\n"
	for _, p := range phases {
		out += "    - {name: " + p + "}\n"
	}
	return []byte(out)
}

func TestDeleteWorkflowObjectThenEnsure(t *testing.T) {
	sch, err := scheme.Load([]byte(delFixture))
	if err != nil {
		t.Fatal(err)
	}
	repo, err := repository.NewGit(t.TempDir(), sch, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	eng := &apply.Engine{Sch: sch, Repo: repo,
		Perm:    &permission.System{},
		Subject: permission.Subject{User: "cored:local", Roles: []string{permission.DefaultAdmin}}}
	mat := &Materializer{Repo: repo, Apply: eng}
	ctx := context.Background()
	wfContent := delWfDoc("delw", []string{"a", "b"})
	if _, _, err := repo.Write(ctx, &repository.Object{Kind: "Workflow", Name: "delw", Scope: "global",
		Content: wfContent}, "t", ""); err != nil {
		t.Fatal(err)
	}
	if act, err := mat.Ensure(ctx, "global", "delw", wfContent); err != nil || act != "created" {
		t.Fatalf("首物化应 created: act=%s err=%v", act, err)
	}
	if act, err := mat.Ensure(ctx, "global", "delw", wfContent); err != nil || act != "noop" {
		t.Fatalf("幂等重入应 noop: act=%s err=%v", act, err)
	}
	// 带外删除（git rm 文件——探针清场路径；进程内写面无 delete 动词）。
	p, err := repo.List(ctx, "Workflow", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(p) != 1 {
		t.Fatalf("删除前列表应 1: %d", len(p))
	}
	// 直接移除底层文件（模拟探针/操作侧清场——repository 面无 delete 动词）。
	matches, _ := filepath.Glob(filepath.Join(repo.Root(), "global", "workflows", "delw.yaml"))
	if len(matches) != 1 {
		t.Fatalf("应恰好一个对象文件: %v", matches)
	}
	if err := os.Remove(matches[0]); err != nil {
		t.Fatal(err)
	}
	p2, err := repo.List(ctx, "Workflow", "")
	if err != nil {
		t.Fatal(err)
	}
	if len(p2) != 0 {
		t.Fatalf("删除后列表应空: n=%d", len(p2))
	}
	// 幂等重入语义（探针实证口径，loop-57 L57-24 判据固定）：git rm 只移除 Workflow 源对象，
	// 骨架 ConnectorOps 对象仍在册（物化器的产出=独立对象，不随源删除级联清场——删除级联
	// 非物化器职责，declared 余量）→ Ensure 在册且 origin 匹配=canonical 比对一致=noop（零派单
	// 零提交——幂等重入是第一优先，与「对象在册」一致）。探针实测第二次 created 的真因=探针
	// 主脚本对同名对象重复 apply（两次 200 落两 commit/两 ok 行），非删除分支词形——以本测试
	// 固定语义锚：删除源对象后骨架不级联消失，重物化=noop。
	if _, err := repo.Get(ctx, "ConnectorOps", "global", "delw"); err != nil {
		t.Fatalf("骨架对象不随源删除级联消失（declared 余量）: %v", err)
	}
	act, err := mat.Ensure(ctx, "global", "delw", wfContent)
	if err != nil {
		t.Fatalf("删除后 ensure 不应报错: %v", err)
	}
	t.Logf("delete-source→ensure act=%s（骨架在册幂等=noop——L57-24 判据固定）", act)
	if act != "noop" {
		t.Fatalf("删除源对象后重物化应 noop（骨架在册幂等重入）: %s", act)
	}
}
