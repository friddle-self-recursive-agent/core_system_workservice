// validate_test.go —— spec 结构准入校验器测试（loop-10 起点①）：
// 声明解析（标注机读语义）→ 校验规则（必填/枚举/嵌套/未知放行/散文放行/解析失败移交存储层）
// → live 面兼容回归（真实声明全文 + docsRoot 三对象合规——闸门不误伤现存对象，探针⑦幂等前提）
// → 真仓对账（checker 直读 core_design 权威声明与部署 docsRoot，缺仓/未注入 skip——部署路径不入公开仓）。
package scheme

import (
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"gopkg.in/yaml.v3"
)

func mustRealScheme(t *testing.T) *Scheme {
	t.Helper()
	s, err := Load([]byte(fixtureSchemeYAML))
	if err != nil {
		t.Fatalf("载入真实形态 scheme: %v", err)
	}
	return s
}

func kindOf(t *testing.T, s *Scheme, name string) Kind {
	t.Helper()
	k, err := s.Lookup(name)
	if err != nil {
		t.Fatalf("kind %s 未注册: %v", name, err)
	}
	return k
}

func validateKindDoc(t *testing.T, s *Scheme, kindName, doc string) error {
	t.Helper()
	return ValidateDoc(kindOf(t, s, kindName), []byte(doc))
}

// ───────────────────────── 声明解析（标注机读语义）─────────────────────────

func TestParseDeclAnnotationSemantics(t *testing.T) {
	cases := []struct {
		decl     string
		required bool
		typ      string
		enum     []string
		wild     bool
	}{
		{"!string", true, "string", nil, false},
		{"?string", false, "string", nil, false},
		{"string (=true)", false, "string", nil, false}, // 括注剥离
		{"?string (=dsh 外部运行时)", false, "string", nil, false},
		{"normal|strict", false, "any", []string{"normal", "strict"}, false},
		{"?ask|auto|all", false, "any", []string{"ask", "auto", "all"}, false},
		{"llm|im|email|cloud|network|generic", false, "any",
			[]string{"llm", "im", "email", "cloud", "network", "generic"}, false},
		{"developer|storage|tool|business|…", false, "any",
			[]string{"developer", "storage", "tool", "business"}, false}, // 开放式枚举剥离 …
		{"'global|systems/<n>/*'!", true, "any", []string{"global", "systems/<n>/*"}, true}, // 子字段必填=!后缀 + 占位通配段
		{"ref!", true, "ref", nil, false},
		{"?map", false, "map", nil, false},
		{"[]string", false, "list", nil, false},
		{"?[]ref", false, "list", nil, false},
		{"{kinds: []string, scopes: []string}!", true, "object", nil, false},
		{"[read|write|approve|call|*]!", true, "list", nil, false},
		// 散文（空白/=/冒号段）→ 不可机读，透明放行
		{"无前缀=默认org | 带前缀=github.com/<org>/<name>", false, "any", nil, false},
		{"'vault:|env:|keychain:'", false, "any", nil, false},
	}
	for _, c := range cases {
		d := parseDecl(c.decl)
		if d.required != c.required || d.typ != c.typ || d.hasWildcard != c.wild {
			t.Errorf("parseDecl(%q) = required=%v typ=%s wild=%v, 要 required=%v typ=%s wild=%v",
				c.decl, d.required, d.typ, d.hasWildcard, c.required, c.typ, c.wild)
		}
		if len(d.enum) != len(c.enum) {
			t.Errorf("parseDecl(%q) enum = %v, 要 %v", c.decl, d.enum, c.enum)
			continue
		}
		for i := range c.enum {
			if d.enum[i] != c.enum[i] {
				t.Errorf("parseDecl(%q) enum = %v, 要 %v", c.decl, d.enum, c.enum)
				break
			}
		}
	}
}

// ───────────────────────── 校验规则 ─────────────────────────

// TestValidateSpecRequiredMissing 必填段缺失拦截（声明 ! 前缀与子字段 ! 后缀同法）。
func TestValidateSpecRequiredMissing(t *testing.T) {
	s := mustRealScheme(t)
	bad := "apiVersion: core/v1\nkind: User\nmetadata: {name: u, scope: global}\nspec: {roles: [], groups: []}\n"
	err := validateKindDoc(t, s, "User", bad)
	var se *SpecError
	if err == nil || !errors.As(err, &se) || !strings.Contains(err.Error(), "display") {
		t.Fatalf("缺 display（!string）应拦: %v", err)
	}
	// spec 区整体缺席且声明含必填 → 拦。
	err = validateKindDoc(t, s, "User", "apiVersion: core/v1\nkind: User\nmetadata: {name: u, scope: global}\n")
	if err == nil || !strings.Contains(err.Error(), "display") {
		t.Fatalf("无 spec 区应拦: %v", err)
	}
	// 好对象放行。
	good := "apiVersion: core/v1\nkind: User\nmetadata: {name: u, scope: global}\nspec: {display: 用户, contact: {}, roles: [], groups: []}\n"
	if err := validateKindDoc(t, s, "User", good); err != nil {
		t.Fatalf("好 User 应放行: %v", err)
	}
}

// TestValidateSpecRolePermissionEntries ⑬E 形状在闸门收口：权限条目缺 scope/actions → 拦
// （声明子字段必填=!后缀，与 permission.LoadObjects 同判）；actions 枚举外值 → 拦；
// 占位通配段（systems/<n>/*）任意 scope 值放行；permissions 区整体缺席放行（声明未标必填——
// 无权限 role 不产生授权，授权面同判接受）。
func TestValidateSpecRolePermissionEntries(t *testing.T) {
	s := mustRealScheme(t)
	base := "apiVersion: core/v1\nkind: Role\nmetadata: {name: r, scope: global}\nspec:\n  permissions:\n"
	cases := []struct {
		name    string
		entry   string
		wantErr string // 空=放行
	}{
		{"好条目", "  - {scope: global, actions: [read, write]}\n", ""},
		{"通配 scope", "  - {scope: 'systems/dev/*', actions: [read]}\n", ""},
		{"actions 星号字面值", "  - {scope: global, actions: [read, *]}\n", ""},
		{"缺 scope（⑬E 形状）", "  - {actions: [read]}\n", "permissions[0].scope: 必填段缺失"},
		{"缺 actions（⑬E 形状）", "  - {scope: global}\n", "permissions[0].actions: 必填段缺失"},
		{"actions 枚举外", "  - {scope: global, actions: [fly]}\n", "不在枚举"},
		{"actions publish（事件发布域——event.md 规则 8 v0.20 枚举扩展）", "  - {scope: events, actions: [publish]}\n", ""},
	}
	for _, c := range cases {
		err := validateKindDoc(t, s, "Role", base+c.entry)
		if c.wantErr == "" {
			if err != nil {
				t.Errorf("%s: 应放行, 得 %v", c.name, err)
			}
			continue
		}
		if err == nil || !strings.Contains(err.Error(), c.wantErr) {
			t.Errorf("%s: 应含 %q, 得 %v", c.name, c.wantErr, err)
		}
	}
	if err := validateKindDoc(t, s, "Role", "apiVersion: core/v1\nkind: Role\nmetadata: {name: r, scope: global}\nspec: {}\n"); err != nil {
		t.Errorf("无 permissions 区应放行（授权面同判）: %v", err)
	}
}

// TestValidateSpecEnumRange 出现即判枚举（含可选字段）：autoLevel/mode 枚举外值拦截、命中放行。
func TestValidateSpecEnumRange(t *testing.T) {
	s := mustRealScheme(t)
	if err := validateKindDoc(t, s, "Agent", "apiVersion: core/v1\nkind: Agent\nmetadata: {name: a, scope: global}\nspec: {autoLevel: sometimes}\n"); err == nil || !strings.Contains(err.Error(), "不在枚举") {
		t.Fatalf("autoLevel 枚举外应拦: %v", err)
	}
	if err := validateKindDoc(t, s, "Agent", "apiVersion: core/v1\nkind: Agent\nmetadata: {name: a, scope: global}\nspec: {autoLevel: all}\n"); err != nil {
		t.Fatalf("autoLevel=all 应放行: %v", err)
	}
	if err := validateKindDoc(t, s, "Workflow", "apiVersion: core/v1\nkind: Workflow\nmetadata: {name: w, scope: global}\nspec: {services: {involved: [system://s], mode: wild}}\n"); err == nil || !strings.Contains(err.Error(), "不在枚举") {
		t.Fatalf("services.mode 枚举外应拦: %v", err)
	}
	if err := validateKindDoc(t, s, "Workflow", "apiVersion: core/v1\nkind: Workflow\nmetadata: {name: w, scope: global}\nspec: {services: {involved: [system://s], mode: observe}}\n"); err != nil {
		t.Fatalf("services.mode=observe 应放行: %v", err)
	}
}

// TestValidateSpecUnknownAndProsePass 未知字段放行（scheme.md「只增不破坏」）；
// 不可机读散文声明透明放行；bool/int 类型错形拦截。
func TestValidateSpecUnknownAndProsePass(t *testing.T) {
	s := mustRealScheme(t)
	// Agent spec 附加未知字段 → 放行。
	doc := "apiVersion: core/v1\nkind: Agent\nmetadata: {name: a, scope: global}\nspec: {autoLevel: all, futureField: {x: 1}}\n"
	if err := validateKindDoc(t, s, "Agent", doc); err != nil {
		t.Fatalf("未知字段应放行: %v", err)
	}
	// WorkService：散文声明字段（nameResolution）放行；connector.download/llmPrompt（!string）齐备。
	ws := "apiVersion: core/v1\nkind: WorkService\nmetadata: {name: w, scope: global}\nspec:\n  category: tool\n  connector:\n    types: [dsh]\n    impl: {kind: tool, ref: \"ref://dsh\"}\n    download: \"tool://local\"\n    scheme: {ops: [x], params: {}}\n    nameResolution: 无前缀=默认org\n    llmPrompt: 调用 dsh\n"
	if err := validateKindDoc(t, s, "WorkService", ws); err != nil {
		t.Fatalf("WorkService 好对象应放行: %v", err)
	}
	// connector.download（!string，真实嵌套映射声明的必填子字段）缺失 → 拦。
	wsBad := "apiVersion: core/v1\nkind: WorkService\nmetadata: {name: w, scope: global}\nspec:\n  category: tool\n  connector:\n    types: [dsh]\n    impl: {kind: tool, ref: \"ref://dsh\"}\n    scheme: {ops: [x], params: {}}\n    nameResolution: 无前缀=默认org\n    llmPrompt: 调用 dsh\n"
	if err := validateKindDoc(t, s, "WorkService", wsBad); err == nil || !strings.Contains(err.Error(), "connector.download") {
		t.Fatalf("connector.download 缺失应拦: %v", err)
	}
	// 类型错形：Workflow enabled 期待布尔、给了字符串 → 拦。
	wfBad := "apiVersion: core/v1\nkind: Workflow\nmetadata: {name: w, scope: global}\nspec: {enabled: \"yes-not-bool\", services: {involved: [system://s], mode: act}}\n"
	if err := validateKindDoc(t, s, "Workflow", wfBad); err == nil || !strings.Contains(err.Error(), "enabled") {
		t.Fatalf("enabled 类型错形应拦: %v", err)
	}
}

// TestValidateWorkflowTransitions 对象声明表形态锚（对象声明表解析首段 v0.27，workflows
// 规则 16）：spec.transitions 声明终态形态（scheme.yaml `?[]{from: Phase, on: string,
// to: Phase}`）的校验闸门在位——好对象放行（apply 落库通道）、非数组拦、行非映射拦。
func TestValidateWorkflowTransitions(t *testing.T) {
	// 权威声明直读（fixture=冻结快照滞后于 v0.25 新增字段；同 TestValidateDocAuthoritativeScheme 口径）。
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "core_design", "core", "scheme.yaml"))
	if errors.Is(err, fs.ErrNotExist) {
		t.Skipf("权威 scheme.yaml 不可达（standalone 检出）: %v", err)
	}
	if err != nil {
		t.Fatalf("读权威声明: %v", err)
	}
	s, err := Load(raw)
	if err != nil {
		t.Fatalf("载入权威声明: %v", err)
	}
	good := "apiVersion: core/v1\nkind: Workflow\nmetadata: {name: w, scope: global}\nspec:\n  transitions:\n  - {from: Idle, on: advance, to: Analyzing}\n  - {from: Idle, on: ApprovalRejected, to: Failed}\n"
	if err := validateKindDoc(t, s, "Workflow", good); err != nil {
		t.Fatalf("声明表好对象应放行（落库通道在位）: %v", err)
	}
	badList := "apiVersion: core/v1\nkind: Workflow\nmetadata: {name: w, scope: global}\nspec: {transitions: not-a-list}\n"
	if err := validateKindDoc(t, s, "Workflow", badList); err == nil || !strings.Contains(err.Error(), "须为数组") {
		t.Fatalf("transitions 非数组应拦: %v", err)
	}
	badRow := "apiVersion: core/v1\nkind: Workflow\nmetadata: {name: w, scope: global}\nspec:\n  transitions:\n  - 3\n"
	if err := validateKindDoc(t, s, "Workflow", badRow); err == nil || !strings.Contains(err.Error(), "须为映射") {
		t.Fatalf("transitions 行须为映射: %v", err)
	}
}

// TestValidateDocParseFailureSkipped 解析不可达移交存储层 ErrContent 管辖（准入不重复判）。
func TestValidateDocParseFailureSkipped(t *testing.T) {
	s := mustRealScheme(t)
	if err := ValidateDoc(kindOf(t, s, "User"), []byte("\t\t:not yaml: [")); err != nil {
		t.Fatalf("不可解析文档应返回 nil（存储层管辖）: %v", err)
	}
	// md+frontmatter：frontmatter 无 spec 区 → 声明含必填（User.display）→ 拦（正文散文不判）。
	md := "---\nname: u\nscope: global\ndocType: ImplDoc\n---\n正文散文，无 spec 结构。\n"
	if err := ValidateDoc(kindOf(t, s, "User"), []byte(md)); err == nil || !strings.Contains(err.Error(), "display") {
		t.Fatalf("md 无 spec 区应拦: %v", err)
	}
}

// ───────────────────────── live 面兼容回归（内嵌快照）─────────────────────────

// fixtureAgentDoc/workflowDoc/descDoc —— docsRoot 现存三对象逐字快照（loop-10 时点）。
// 闸门上线不得误伤现存对象（探针⑦幂等重入前提）；docsRoot 演进后由 CORE_DOCS_ROOT live 对账覆盖。
const fixtureAgentDoc = `apiVersion: core/v1
kind: Agent
metadata:
  name: core-headless
  scope: global
  resourceVersion: 2
  labels: {tier: control-plane, bootstrap-loop: "3"}
spec:
  runtime: dsh
  autoLevel: all
  docsPolicy:
    read: ["core/*", "global/*", "systems/*"]
    onDemand: true
  retry:
    maxAttempts: 3
    backoff: 10s
    escalateTo: root
`

const fixtureWorkflowDoc = `apiVersion: core/v1
kind: Workflow
metadata:
  name: dev-delivery
  scope: global
  resourceVersion: 1
spec:
  enabled: true
  services:
    involved: ["system://self-recursive-core"]
    mode: act
  triggers:
    events: {types: ["DocApplied", "ExternalChanged"], minLevel: info}
    timer: {resync: 30m}
  onEscalate: {action: pause}
  onFailed: {backoff: 5m, maxAttempts: 3, escalate: {to: root}}
  brakes: {cooldown: 5m, maxMutationsPerTick: 3, avoidRepeatFix: true}
  regression: {statusWriteback: true, reports: [DriftReport, SessionLog]}
`

const fixtureServiceDescDoc = `apiVersion: core/v1
kind: ServiceDesc
metadata:
  name: self-recursive-core
  scope: self-recursive-core   # layout.md 物化：builtin 标记系统落位系统 scope（声明侧 scope: builtin 是打包标记）
  resourceVersion: 1
  labels: {tier: control-plane, builtin: "true"}
spec:
  goalRef: system://self-recursive-core/goal
  descRef: system://self-recursive-core/desc
  docs:
    - system://self-recursive-core/rules
  externalRefs:
    - {capabilities: [dsh-host, llm-gateway]}
  instances:
    - name: headless
      agentRefs: ["agent://core-headless"]
  builtinServices: [dsh, gitfs, opslog]
status: {}
`

func TestValidateDocLiveDocsConform(t *testing.T) {
	s := mustRealScheme(t)
	for kind, doc := range map[string]string{
		"Agent":       fixtureAgentDoc,
		"Workflow":    fixtureWorkflowDoc,
		"ServiceDesc": fixtureServiceDescDoc,
	} {
		if err := validateKindDoc(t, s, kind, doc); err != nil {
			t.Errorf("live %s 对象应过闸门（现存对象不误伤——探针⑦幂等前提）: %v", kind, err)
		}
	}
	// scheme 自身 vs Scheme kind 声明（自注册自校验）。
	if err := validateKindDoc(t, s, "Scheme", fixtureSchemeYAML); err != nil {
		t.Errorf("scheme.yaml 应过 Scheme kind 声明（自注册）: %v", err)
	}
}

// ───────────────────────── 真仓对账（声明↔checker↔部署面）─────────────────────────

// TestValidateDocAuthoritativeScheme checker 直读 core_design 权威声明（相对路径；
// standalone 检出缺仓 skip）。权威声明引入 checker 不可机读形态时此处先红（R-TYPE 对账锚点）。
func TestValidateDocAuthoritativeScheme(t *testing.T) {
	path := filepath.Join("..", "..", "..", "core_design", "core", "scheme.yaml")
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			t.Skipf("权威 scheme.yaml 不可达（standalone 检出）: %v", err)
		}
		t.Fatalf("读权威声明: %v", err)
	}
	s, err := Load(raw)
	if err != nil {
		t.Fatalf("权威 scheme.yaml 载入失败: %v", err)
	}
	if err := ValidateDoc(kindOf(t, s, "Scheme"), raw); err != nil {
		t.Fatalf("权威 scheme.yaml 自校验失败: %v", err)
	}
	// 部署 docsRoot 全量对账（CORE_DOCS_ROOT 注入；缺省 skip——部署路径不入公开仓，R-PUB-1）。
	root := os.Getenv("CORE_DOCS_ROOT")
	if root == "" {
		t.Skip("CORE_DOCS_ROOT 未注入（部署路径不入仓，R-PUB-1）")
	}
	var paths []string
	_ = filepath.WalkDir(root, func(p string, d fs.DirEntry, err error) error {
		if err == nil && !d.IsDir() && strings.HasSuffix(p, ".yaml") {
			paths = append(paths, p)
		}
		return nil // walk 错误不打断（个别不可读文件不阻塞对账）
	})
	if len(paths) == 0 {
		t.Fatalf("docsRoot 无 yaml 对象（%s）", root)
	}
	var head struct {
		Kind     string `yaml:"kind"`
		Metadata struct {
			Name  string `yaml:"name"`
			Scope string `yaml:"scope"`
		} `yaml:"metadata"`
	}
	for _, p := range paths {
		rawDoc, err := os.ReadFile(p)
		if err != nil {
			t.Errorf("读 %s: %v", p, err)
			continue
		}
		if err := yaml.Unmarshal(rawDoc, &head); err != nil {
			t.Errorf("解析 %s 头部: %v", p, err)
			continue
		}
		if head.Kind == "" {
			continue
		}
		k, err := s.Lookup(head.Kind)
		if err != nil {
			t.Errorf("docsRoot %s: kind %q 未注册（野对象落面）", p, head.Kind)
			continue
		}
		if err := ValidateDoc(k, rawDoc); err != nil {
			t.Errorf("docsRoot %s 未过闸门: %v", p, err)
		}
	}
}
