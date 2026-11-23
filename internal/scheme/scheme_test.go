package scheme

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// fixture 结构上对齐 core_design/core/scheme.yaml 的真实形态：
// 块映射 + 行内流映射 + 注释 + inline 未知字段（cascade）。
const fixture = `# test scheme
apiVersion: core/v1
kind: Scheme
metadata:
  name: default
  scope: builtin
spec:
  version: core/v1
  kinds:
    - kind: Scheme                    # 自注册
      builtin: true
      storage: core/scheme.yaml
      spec: {version: "!string", kinds: "[]KindReg"}

    - kind: ServiceInstance
      builtin: true
      storage: systems/<s>/instances/<name>/instance.yaml
      spec: {}
      status:
        phase: "InSync|Drifting|Remediating|Failed|Escalated"
      cascade: 删 systems/<s>/ 目录即级联（目录即对象）

    - kind: Agent
      builtin: true
      storage: global/agents/<name>/agent.yaml
      spec:
        autoLevel: "?ask|auto|all"
      builtinSet: []
  docTypes:
    - {name: Goal, mandatory: required, sections: [intent, acceptance], writers: [human], appliesTo: ServiceDesc, admission: normal}
    - {name: ImplDoc, mandatory: required, sections: [how, mapping], writers: [agent], appliesTo: ServiceInstance, admission: normal}
  eventTypes:
    - {type: DocApplied, source: core, defaultLevel: info, dedup: "apply 落库必发；engine 自写豁免", route: owner-workflow}
    - {type: HeartbeatLost, source: core, defaultLevel: error, dedup: "<agent>", route: parent-agent}
  dispatch:
    - {on: DocApplied, for: "owner:*", dispatch: owner-workflow}
    - {on: "*", for: "any", dispatch: dead-letter}    # 死信必须可见
`

func mustLoad(t *testing.T, data string) *Scheme {
	t.Helper()
	s, err := Load([]byte(data))
	if err != nil {
		t.Fatalf("Load() 意外失败: %v", err)
	}
	return s
}

func TestLoadValidFixture(t *testing.T) {
	s := mustLoad(t, fixture)
	if s.APIVersion != "core/v1" {
		t.Errorf("apiVersion = %q, 要 core/v1", s.APIVersion)
	}
	if s.Spec.Version != "core/v1" {
		t.Errorf("spec.version = %q, 要 core/v1", s.Spec.Version)
	}
	if s.Metadata.Name != "default" || s.Metadata.Scope != "builtin" {
		t.Errorf("metadata = %+v, 要 {default builtin}", s.Metadata)
	}
	sum := s.Summary()
	if sum.KindCount != 3 || sum.DocTypeCount != 2 || sum.EventTypeCount != 2 || sum.DispatchCount != 2 {
		t.Errorf("Summary 计数不符: %+v", sum)
	}
}

func TestLoadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "scheme.yaml")
	if err := os.WriteFile(path, []byte(fixture), 0o644); err != nil {
		t.Fatal(err)
	}
	s, err := LoadFile(path)
	if err != nil {
		t.Fatalf("LoadFile() 意外失败: %v", err)
	}
	if s.Source != path {
		t.Errorf("Source = %q, 要 %q", s.Source, path)
	}
}

func TestLookupGate(t *testing.T) {
	s := mustLoad(t, fixture)
	k, err := s.Lookup("Agent")
	if err != nil {
		t.Fatalf("Lookup(Agent) 失败: %v", err)
	}
	if !k.Builtin || k.Storage != "global/agents/<name>/agent.yaml" {
		t.Errorf("Agent 注册项不符: %+v", k)
	}
	if k.Spec["autoLevel"] != "?ask|auto|all" {
		t.Errorf("Agent.spec.autoLevel = %v, 要 ?ask|auto|all", k.Spec["autoLevel"])
	}
	if _, err := s.Lookup("NotRegistered"); err == nil {
		t.Error("未注册 kind 应返回错误（野对象闸门）")
	} else if !strings.Contains(err.Error(), "not registered") {
		t.Errorf("错误信息不符: %v", err)
	}
}

func TestInlineExtraCaptured(t *testing.T) {
	s := mustLoad(t, fixture)
	k, err := s.Lookup("ServiceInstance")
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := k.Extra["cascade"]; !ok {
		t.Errorf("未知顶层字段 cascade 应进 Extra: %+v", k.Extra)
	}
}

func TestListKinds(t *testing.T) {
	s := mustLoad(t, fixture)
	kinds := s.ListKinds()
	if len(kinds) != 3 {
		t.Fatalf("ListKinds 长度 = %d, 要 3", len(kinds))
	}
	// 返回副本：外部修改不污染注册表。
	kinds[0].Name = "mutated"
	again, _ := s.Lookup("Scheme")
	if again.Name != "Scheme" {
		t.Errorf("ListKinds 返回了内部切片（被外部改写）: %q", again.Name)
	}
}

func TestValidationErrors(t *testing.T) {
	cases := []struct {
		name string
		yaml string
		want string
	}{
		{"缺apiVersion", strings.Replace(fixture, "apiVersion: core/v1", "", 1), "apiVersion 必填"},
		{"缺version", strings.Replace(fixture, "  version: core/v1\n", "", 1), "spec.version 必填"},
		{"缺kinds", strings.Replace(fixture, "  kinds:", "  xkinds:", 1), "kinds 不能为空"},
		{"kind重复", strings.Replace(fixture, "- kind: Agent", "- kind: Scheme", 1), `kind "Scheme" 重复注册`},
		{"缺storage", strings.Replace(fixture, "      storage: global/agents/<name>/agent.yaml", "", 1), `kind "Agent" 缺 storage`},
		{"docType重复", strings.Replace(fixture, "{name: ImplDoc,", "{name: Goal,", 1), `docType "Goal" 重复注册`},
		{"dispatch缺on", strings.Replace(fixture, "{on: DocApplied,", "{onx: DocApplied,", 1), "dispatch[0] 缺 on 或 dispatch"},
		{"kind字段非法", strings.Replace(fixture, "kind: Scheme", "kind: NotScheme", 1), "kind 必须为 Scheme"},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := Load([]byte(c.yaml))
			if err == nil {
				t.Fatalf("应校验失败（要含 %q）", c.want)
			}
			if !strings.Contains(err.Error(), c.want) {
				t.Errorf("错误 %q 不含 %q", err.Error(), c.want)
			}
		})
	}
}

func TestLoadBadYAML(t *testing.T) {
	if _, err := Load([]byte("apiVersion: [unclosed")); err == nil {
		t.Fatal("坏 YAML 应解析失败")
	}
}
