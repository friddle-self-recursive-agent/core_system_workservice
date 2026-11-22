package permission

import (
	"strings"
	"testing"
)

const roleDevtools = `apiVersion: core/v1
kind: Role
metadata: {name: devtools-admin, scope: global}
spec:
  on: [agent/ci-repair-agent]
  permissions:
  - {scope: "systems/developer-system/*", actions: [read, write, approve]}
  - {scope: "ops/onedev-ci", actions: [call]}
`

const roleReadonly = `apiVersion: core/v1
kind: Role
metadata: {name: readonly, scope: global}
spec:
  permissions:
  - {scope: "global/*", actions: [read]}
`

const groupDevtools = `apiVersion: core/v1
kind: Group
metadata: {name: devtools-team, scope: global}
spec:
  roles: [role/devtools-admin]
`

func loadFixtures(t *testing.T) *System {
	t.Helper()
	s := &System{}
	err := s.LoadObjects([]ObjectDoc{
		{Kind: "Role", Name: "devtools-admin", Scope: "global", Content: []byte(roleDevtools)},
		{Kind: "Role", Name: "readonly", Scope: "global", Content: []byte(roleReadonly)},
		{Kind: "Group", Name: "devtools-team", Scope: "global", Content: []byte(groupDevtools)},
	})
	if err != nil {
		t.Fatalf("LoadObjects: %v", err)
	}
	return s
}

func TestLoadObjectsAndCheck(t *testing.T) {
	s := loadFixtures(t)
	if len(s.Roles) != 2 || len(s.Groups) != 1 {
		t.Fatalf("装载计数不符: roles=%d groups=%d", len(s.Roles), len(s.Groups))
	}

	// role 权限表查表：断言 role 名（带 kind 前缀惯例）命中对象授权。
	// 对象域=对象 scope 值（系统层 "<n>"），模式 systems/developer-system/* 经路径形匹配。
	sub := Subject{User: "agent:ci", Roles: []string{"role/devtools-admin"}}
	if err := s.Check(sub, "developer-system", "write", ByRole); err != nil {
		t.Fatalf("持 role 主体应放行（系统层域）: %v", err)
	}
	if err := s.Check(sub, "developer-system/inst-1", "write", ByRole); err != nil {
		t.Fatalf("实例层域应被子域通配覆盖: %v", err)
	}
	// scope 域外拒绝（骨架 1：对象落在哪个 scope 受哪层权限管）。
	if err := s.Check(sub, "global", "write", ByRole); err == nil {
		t.Fatal("scope 外应拒绝")
	}
	if err := s.Check(sub, "kube-inner", "write", ByRole); err == nil {
		t.Fatal("跨系统域应拒绝")
	}
	// 未知 role 名=断言不产生权限（授权只来自 Role 对象）。
	if err := s.Check(Subject{User: "x", Roles: []string{"role/nope"}}, "developer-system", "write", ByRole); err == nil {
		t.Fatal("未注册 role 应拒绝")
	}
}

func TestCheckViaGroupBinding(t *testing.T) {
	s := loadFixtures(t)
	// user→group→role 链：group 绑定带前缀引用（role/devtools-admin），装载时已规整。
	sub := Subject{User: "u-zhang", Groups: []string{"group/devtools-team"}}
	if err := s.Check(sub, "systems/developer-system/x", "write", ByRole); err != nil {
		t.Fatalf("经 group 绑定应放行: %v", err)
	}
	if err := s.Check(sub, "ops/onedev-ci", "call", ByRole); err != nil {
		t.Fatalf("group 链第二条权限应放行: %v", err)
	}
	// group 不直接绑权限：无 role 落点的 group 断言无授权。
	if err := s.Check(Subject{Groups: []string{"group/empty"}}, "global", "write", ByRole); err == nil {
		t.Fatal("无 role 绑定的 group 应拒绝")
	}
}

func TestDefaultAdminAndHardRules(t *testing.T) {
	s := loadFixtures(t)
	sub := Subject{User: "cored:local", Roles: []string{DefaultAdmin}}
	if err := s.Check(sub, "任意/未注册域", "write", ByRole); err != nil {
		t.Fatalf("内置 admin 应全权限: %v", err)
	}
	if err := HardRules("agent/builtinRoot", "delete"); err == nil {
		t.Fatal("硬规则应拒删 builtinRoot")
	}
}

func TestWildcardAction(t *testing.T) {
	s := &System{}
	if err := s.LoadObjects([]ObjectDoc{{Kind: "Role", Name: "ops", Scope: "global", Content: []byte(
		`kind: Role
metadata: {name: ops}
spec:
  permissions:
  - {scope: "global/*", actions: ["*"]}
`)}}); err != nil {
		t.Fatalf("LoadObjects: %v", err)
	}
	if err := s.Check(Subject{Roles: []string{"role/ops"}}, "global", "approve", ByRole); err != nil {
		t.Fatalf("通配 action+基域应放行: %v", err)
	}
}

func TestResourceDomainExactMatch(t *testing.T) {
	s := &System{}
	if err := s.LoadObjects([]ObjectDoc{{Kind: "Role", Name: "ci-caller", Scope: "global", Content: []byte(
		`kind: Role
metadata: {name: ci-caller}
spec:
  permissions:
  - {scope: "ops/onedev-ci", actions: [call]}
`)}}); err != nil {
		t.Fatalf("LoadObjects: %v", err)
	}
	sub := Subject{User: "agent:ci", Roles: []string{"role/ci-caller"}}
	if err := s.Check(sub, "ops/onedev-ci", "call", ByRole); err != nil {
		t.Fatalf("资源域应精确放行: %v", err)
	}
	if err := s.Check(sub, "ops/onedev-ci", "write", ByRole); err == nil {
		t.Fatal("资源域外动作应拒绝")
	}
}

func TestLoadObjectsRejectsBadDocs(t *testing.T) {
	cases := []struct {
		name string
		docs []ObjectDoc
	}{
		{"坏 yaml", []ObjectDoc{{Kind: "Role", Name: "r", Content: []byte("spec: [")}}},
		{"未知 kind", []ObjectDoc{{Kind: "Agent", Name: "a", Content: []byte("kind: Agent")}}},
		{"缺名", []ObjectDoc{{Kind: "Role", Content: []byte("kind: Role\nspec: {}\n")}}},
		{"权限缺 scope", []ObjectDoc{{Kind: "Role", Name: "r", Content: []byte(
			"kind: Role\nmetadata: {name: r}\nspec:\n  permissions:\n  - {actions: [read]}\n")}}},
		{"权限缺 actions", []ObjectDoc{{Kind: "Role", Name: "r", Content: []byte(
			"kind: Role\nmetadata: {name: r}\nspec:\n  permissions:\n  - {scope: global}\n")}}},
		{"role 重名", []ObjectDoc{
			{Kind: "Role", Name: "r", Content: []byte("kind: Role\nmetadata: {name: r}\nspec: {}\n")},
			{Kind: "Role", Name: "r2", Content: []byte("kind: Role\nmetadata: {name: r}\nspec: {}\n")},
		}},
	}
	for _, c := range cases {
		s := &System{}
		if err := s.LoadObjects(c.docs); err == nil {
			t.Fatalf("%s: 应拒绝", c.name)
		}
	}
	// 整体拒绝=不半装载：坏文档批次后表保持空（快照替换语义）。
	s := &System{}
	bad := append([]ObjectDoc{
		{Kind: "Role", Name: "good", Scope: "global", Content: []byte(roleReadonly)},
	}, cases[0].docs...)
	if err := s.LoadObjects(bad); err == nil {
		t.Fatal("含坏对象批次应整体拒绝")
	}
	if len(s.Roles) != 0 {
		t.Fatalf("拒绝批次不得半装载: roles=%d", len(s.Roles))
	}
}

func TestSubjectID(t *testing.T) {
	if got := (Subject{User: "u-zhang"}).ID(); got != "u-zhang" {
		t.Fatalf("user 优先: %q", got)
	}
	if got := (Subject{Roles: []string{"role/a", "role/b"}}).ID(); got != "role/a,role/b" {
		t.Fatalf("无 user 记断言: %q", got)
	}
	if got := (Subject{}).ID(); got != "anonymous" {
		t.Fatalf("全空=anonymous: %q", got)
	}
}

func TestDeniedErrorCarriesSubjectAndJudge(t *testing.T) {
	s := loadFixtures(t)
	err := s.Check(Subject{User: "agent:x", Roles: []string{"role/readonly"}}, "global/y", "write", ByRole)
	if err == nil {
		t.Fatal("应拒绝")
	}
	msg := err.Error()
	if !strings.Contains(msg, "subject=agent:x") || !strings.Contains(msg, "judged by role") {
		t.Fatalf("拒绝错误应携带主体与判定方（规则 3 留痕）: %v", msg)
	}
}
