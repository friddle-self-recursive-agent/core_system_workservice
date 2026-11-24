// lint_test.go —— 净入 lint 检出力单测（adoption §1.19 L48-02：checker 自身可证伪——
// 「无 checker 的 lint 规则=未生效」收口的检测面）。五闸逐类坏对象 fixture 检出 + 合规对象零误报。
package repository

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// lintScheme 净入 lint 专用 scheme：模板形态对齐 core_design/core/scheme.yaml，并给
// WorkService/Provider 挂真实 spec 声明（spec 准入闸/凭据闸的声明面镜像）。
const lintScheme = `apiVersion: core/v1
kind: Scheme
metadata: {name: lint, scope: builtin}
spec:
  version: core/v1
  kinds:
    - {kind: Scheme, builtin: true, storage: core/scheme.yaml}
    - {kind: Agent, builtin: true, storage: global/agents/<name>/agent.yaml}
    - {kind: ServiceDesc, builtin: true, storage: systems/<name>/system.yaml}
    - {kind: Workflow, builtin: true, storage: "global/workflows/<name>.yaml | systems/<s>/workflows/<name>.yaml"}
    - {kind: Provider, builtin: true, storage: global/providers/<name>.yaml}
    - {kind: Notification, builtin: true, storage: global/notifications/<name>.yaml}
    - kind: WorkService
      builtin: true
      storage: global/workservices/<name>/workservice.yaml
      spec:
        category: "developer|storage|tool|business|…"
        connector:
          types: "[]string"
          impl: "{kind: tool|image|source, ref: ref}"
          download: "!string"
          scheme: "{ops: []string, params: map}"
          usage: "?string"
          llmPrompt: "!string"
        observation:
          metrics: "[]{name, query, thresholds: ?[]{op: gt|gte|lt|lte, limit: number}}"
`

func lintObjs(t *testing.T, files map[string]string) ([]Finding, int) {
	t.Helper()
	root := t.TempDir()
	for p, c := range files {
		full := filepath.Join(root, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatalf("建目录 %q: %v", p, err)
		}
		if err := os.WriteFile(full, []byte(c), 0o644); err != nil {
			t.Fatalf("写 %q: %v", p, err)
		}
	}
	sch, err := scheme.Load([]byte(lintScheme))
	if err != nil {
		t.Fatalf("载入 lint scheme: %v", err)
	}
	f, n, err := LintDocs(root, sch)
	if err != nil {
		t.Fatalf("LintDocs: %v", err)
	}
	return f, n
}

func has(fs []Finding, code, pathPart string) bool {
	for _, f := range fs {
		if f.Code == code && strings.Contains(f.Path, pathPart) {
			return true
		}
	}
	return false
}

// 有效 WorkService（与生产 notify 对象同构的最小形态）。
const goodWS = `apiVersion: core/v1
kind: WorkService
metadata:
  name: good
  scope: global
  resourceVersion: 1
spec:
  category: tool
  connector:
    types: [notify]
    impl: {kind: source, ref: common_connector/notify}
    download: pkg://common_connector/notify
    scheme: {ops: [notify], params: {subject: string}}
    usage: "Invoke(ctx, op, params)"
    llmPrompt: "发送通知"
`

func TestLintDocsValidObjectZeroFindings(t *testing.T) {
	f, n := lintObjs(t, map[string]string{
		"global/workservices/good/workservice.yaml": goodWS,
	})
	if len(f) != 0 {
		t.Fatalf("合规对象应零发现，得 %v", f)
	}
	if n != 1 {
		t.Fatalf("文件计数应 1，得 %d", n)
	}
}

func TestLintDocsUnregisteredPath(t *testing.T) {
	f, _ := lintObjs(t, map[string]string{
		"global/unknown-things/foo.yaml": goodWS, // 无模板覆盖的根级野路径
	})
	if !has(f, "unregistered-path", "foo.yaml") {
		t.Fatalf("野路径应检出 unregistered-path，得 %v", f)
	}
}

func TestLintDocsUnparseable(t *testing.T) {
	f, _ := lintObjs(t, map[string]string{
		"global/workservices/bad/workservice.yaml": "kind: [unclosed\n  metadata: {",
	})
	if !has(f, "unparseable", "bad/workservice.yaml") {
		t.Fatalf("不可解析应检出 unparseable（读面静默丢弃面补检出），得 %v", f)
	}
}

func TestLintDocsIdentityScopeDrift(t *testing.T) {
	// L44 迁移缺口同型：metadata.scope 与物化路径不一致。
	drift := strings.Replace(goodWS, "scope: global", "scope: notify", 1)
	f, _ := lintObjs(t, map[string]string{
		"global/workservices/good/workservice.yaml": drift,
	})
	if !has(f, "identity", "good/workservice.yaml") {
		t.Fatalf("scope 漂移应检出 identity，得 %v", f)
	}
}

func TestLintDocsIdentityNameMismatch(t *testing.T) {
	// metadata.name 与 <name> 段不符（applyWrite ErrIdentity 同闸）。
	bad := strings.Replace(goodWS, "name: good", "name: other", 1)
	f, _ := lintObjs(t, map[string]string{
		"global/workservices/good/workservice.yaml": bad,
	})
	if !has(f, "identity", "good/workservice.yaml") {
		t.Fatalf("name 不符应检出 identity，得 %v", f)
	}
}

func TestLintDocsIdentityNameSeg(t *testing.T) {
	// 对象名不合命名规则（大写/下划线——concretePath 同闸）。
	f, _ := lintObjs(t, map[string]string{
		"global/workservices/Bad_Name/workservice.yaml": goodWS,
	})
	if !has(f, "identity", "Bad_Name") {
		t.Fatalf("命名规则违规应检出 identity，得 %v", f)
	}
}

func TestLintDocsIdentityKindField(t *testing.T) {
	// 文档 kind 与路径所属 kind 不符（写面盲区：对象 kind 来自路由）。
	bad := strings.Replace(goodWS, "kind: WorkService", "kind: Provider", 1)
	f, _ := lintObjs(t, map[string]string{
		"global/workservices/good/workservice.yaml": bad,
	})
	if !has(f, "identity", "good/workservice.yaml") {
		t.Fatalf("kind 字段漂移应检出 identity，得 %v", f)
	}
}

func TestLintDocsSpecShapeViolation(t *testing.T) {
	// loop-47 deviations ③ 存量形态：connector.scheme.params 字符串形态（读面放行/写面拒绝）。
	bad := strings.Replace(goodWS,
		"scheme: {ops: [notify], params: {subject: string}}",
		`scheme: {ops: [notify], params: "subject=string"}`, 1)
	f, _ := lintObjs(t, map[string]string{
		"global/workservices/good/workservice.yaml": bad,
	})
	if !has(f, "spec", "good/workservice.yaml") {
		t.Fatalf("params 字符串形态应检出 spec 违规，得 %v", f)
	}
}

func TestLintDocsCredentialEmbedded(t *testing.T) {
	// Provider 凭据明文直写（provider.md 规则 2——引用式前缀同闸）。
	bad := `apiVersion: core/v1
kind: Provider
metadata: {name: leak, scope: global, resourceVersion: 1}
spec:
  kind: email
  auth: {scheme: basic, secret: {source: "user:password", rotation: none}}
`
	f, _ := lintObjs(t, map[string]string{
		"global/providers/leak.yaml": bad,
	})
	if !has(f, "credential", "leak.yaml") {
		t.Fatalf("凭据明文应检出 credential，得 %v", f)
	}
}

func TestLintDocsSameKindMultiBranchNotAmbiguous(t *testing.T) {
	// 同 kind 多分支命中（Workflow global 与 systems 两栖）不算歧义。
	f, _ := lintObjs(t, map[string]string{
		"global/workflows/dev.yaml":     "apiVersion: core/v1\nkind: Workflow\nmetadata: {name: dev, scope: global, resourceVersion: 1}\n",
		"systems/sx/workflows/dev.yaml": "apiVersion: core/v1\nkind: Workflow\nmetadata: {name: dev, scope: sx, resourceVersion: 1}\n",
	})
	if len(f) != 0 {
		t.Fatalf("同 kind 双分支命中应零发现，得 %v", f)
	}
}

func TestLintDocsNonYamlIgnored(t *testing.T) {
	f, n := lintObjs(t, map[string]string{
		"global/workservices/good/workservice.yaml": goodWS,
		"README.md":  "# docsRoot 说明",
		".gitignore": "acceptance/\n",
	})
	if len(f) != 0 || n != 1 {
		t.Fatalf("非 yaml 不在准入域（findings=%v files=%d）", f, n)
	}
}

func TestLintSchemeMirrorDrift(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "core"), 0o755); err != nil {
		t.Fatalf("建目录: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "core", "scheme.yaml"), []byte("apiVersion: core/v1\n"), 0o644); err != nil {
		t.Fatalf("写镜像: %v", err)
	}
	present, drift, err := LintSchemeMirror(root, []byte("apiVersion: core/v1\nkind: Scheme\n"))
	if err != nil || !present || drift == nil {
		t.Fatalf("镜像漂移应检出（present=%v drift=%v err=%v）", present, drift, err)
	}
	present, drift, err = LintSchemeMirror(root, []byte("apiVersion: core/v1\n"))
	if err != nil || !present || drift != nil {
		t.Fatalf("镜像一致应绿（present=%v drift=%v err=%v）", present, drift, err)
	}
	present, drift, err = LintSchemeMirror(t.TempDir(), []byte("x"))
	if err != nil || present || drift != nil {
		t.Fatalf("镜像缺席应不判（present=%v drift=%v err=%v）", present, drift, err)
	}
}

func TestDerivedScope(t *testing.T) {
	cases := map[string]string{
		"core/scheme.yaml":                           "builtin",
		"global/agents/a/agent.yaml":                 "global",
		"systems/s/system.yaml":                      "s",
		"systems/s/instances/i/instance.yaml":        "s/i",
		"doctypes/x.yaml":                            "",
		"global/workflows/dev.yaml":                  "global",
		"systems/s/instances/i/skills/k.yaml":        "s/i",
		"systems/s/instances/i/connector-ops/c.yaml": "s/i",
	}
	for rel, want := range cases {
		if got := derivedScope(rel); got != want {
			t.Errorf("derivedScope(%q)=%q 期望 %q", rel, got, want)
		}
	}
}
