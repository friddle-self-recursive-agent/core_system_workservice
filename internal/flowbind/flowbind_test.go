// flowbind_test.go —— workflow 执行模型绑定面验收（R-7，workflow-runtime.md §11.1）：
// 绑定解析（spec.connector 覆写/缺省同名/前缀归一化）、骨架确定性（origin 标注+宿主声明+节点投影）、
// 物化器幂等闭环（创建→noop→spec 变更 rv+1）、所有面（非 workflow-bind 对象零盲写）、
// 决策面 execution 投影（bound/unbound/skeleton+scope 优先级）。
package flowbind

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/apply"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// 测试 scheme：Workflow/ConnectorOps/WorkService 三 kind 最小注册（flow 声明与权威 scheme 同词汇）。
const testScheme = `apiVersion: core/v1
kind: Scheme
metadata: {name: test, scope: builtin}
spec:
  version: core/v1
  kinds:
    - {kind: Scheme, builtin: true, storage: core/scheme.yaml}
    - kind: Workflow
      builtin: true
      storage: "global/workflows/<name>.yaml | systems/<s>/workflows/<name>.yaml"
      spec:
        enabled: "?bool (=true)"
        services: "{involved: []ref, mode: observe|act}"
        connector: "?ref (=<同名 ConnectorOps>)"
        phases: "[]Phase"
    - kind: ConnectorOps
      builtin: true
      storage: "global/connector-ops/<name>.yaml | systems/<s>/connector-ops/<name>.yaml"
      spec:
        workservice: "!ref"
        workflow: "?string (=default-task)"
        config: "?map"
        auth: "?{provider: ref}"
        ops: "[]string"
        flow: "?{nodes: []{name: string, title: ?string, unit: session|agent, op: ?string, connector: ?string, provider: ?string}}"
    - kind: WorkService
      builtin: true
      storage: "global/workservices/<name>/workservice.yaml"
      spec:
        category: "developer|storage|tool|business|…"
        connector:
          types: "[]string"
          impl: "{kind: tool|image|source, ref: ref}"
          download: "!string"
          scheme: "{ops: []string, params: map}"
          usage: "?string"
          llmPrompt: "!string"
        tools: "[]{op, params, idempotent, undo: full|partial|none, cleanup: ?op}"
  eventTypes:
    - {type: DocApplied, source: core, defaultLevel: info, route: owner-workflow}
`

type world struct {
	repo   *repository.GitRepo
	ops    *opslog.SQLiteStore
	bus    *eventbus.Bus
	engine *apply.Engine
}

func newWorld(t *testing.T) *world {
	t.Helper()
	sch, err := scheme.Load([]byte(testScheme))
	if err != nil {
		t.Fatalf("载入测试 scheme: %v", err)
	}
	dir := t.TempDir()
	ops, err := opslog.OpenSQLite(filepath.Join(dir, "ops.db"))
	if err != nil {
		t.Fatalf("ops DB: %v", err)
	}
	t.Cleanup(func() { _ = ops.Close() })
	repo, err := repository.NewGit(dir, sch, &repository.Options{Ops: ops, WatchInterval: 10 * time.Millisecond})
	if err != nil {
		t.Fatalf("git 后端: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	bus := eventbus.NewBus(&eventbus.Options{Ops: ops, DedupWindow: 50 * time.Millisecond})
	w := &world{
		repo: repo, ops: ops, bus: bus,
		engine: &apply.Engine{
			Sch: sch, Repo: repo, Bus: bus, Ops: ops,
			Perm:    &permission.System{},
			Subject: permission.Subject{User: "cored:local", Roles: []string{permission.DefaultAdmin}},
		},
	}
	return w
}

func wfYAML(name, scope, connector, phases string) []byte {
	var c string
	if connector != "" {
		c = "  connector: " + connector + "\n"
	}
	return []byte(`apiVersion: core/v1
kind: Workflow
metadata: {name: ` + name + `, scope: ` + scope + `}
spec:
  enabled: true
` + c + `  phases:
` + phases)
}

var twoPhases = "    - {name: analyze, title: Analyze}\n    - {name: verify, title: Verify}\n"

func mustApply(t *testing.T, w *world, kind, scope, name string, rv int64, content []byte) *apply.Result {
	t.Helper()
	res, err := w.engine.Apply(context.Background(), &repository.Object{
		Kind: kind, Name: name, Scope: scope, ResourceVersion: rv, Content: content,
	}, apply.Options{FieldManager: "test:" + name})
	if err != nil {
		t.Fatalf("apply %s/%s/%s: %v", kind, scope, name, err)
	}
	return res
}

func TestResolveBinding(t *testing.T) {
	// 缺省同名 + 字符串相位行投影（unit 缺省 session）。
	b, ok := Resolve("global", "flow-probe", []byte(`metadata: {name: flow-probe, scope: global}
spec:
  phases:
    - analyze
    - {name: verify, title: Verify}
`))
	if !ok || b.Connecto != "flow-probe" || b.Scope != "global" || b.Workflow != "flow-probe" {
		t.Fatalf("缺省同名解析: ok=%v b=%+v", ok, b)
	}
	if len(b.Nodes) != 2 || b.Nodes[0].Name != "analyze" || b.Nodes[0].Unit != "session" ||
		b.Nodes[1].Title != "Verify" {
		t.Fatalf("节点投影: %+v", b.Nodes)
	}
	// spec.connector 覆写 + kind 前缀归一化 + 富行接线字段 + unit 声明。
	b2, ok := Resolve("s1", "wf-x", []byte(`metadata: {name: wf-x, scope: s1}
spec:
  connector: connector-ops/wf-x-flow
  phases:
    - {name: analyze, title: Analyze, op: notify/send-email, connector: notify-send-email, provider: provider/email-notify}
    - {name: dispatch, unit: agent}
`))
	if !ok || b2.Connecto != "wf-x-flow" {
		t.Fatalf("覆写归一化: ok=%v b=%+v", ok, b2)
	}
	if b2.Nodes[0].Op != "notify/send-email" || b2.Nodes[0].Connector != "notify-send-email" ||
		b2.Nodes[0].Provider != "provider/email-notify" {
		t.Fatalf("富行接线投影: %+v", b2.Nodes[0])
	}
	if b2.Nodes[1].Unit != "agent" {
		t.Fatalf("unit 声明投影: %+v", b2.Nodes[1])
	}
	// 零相位=不可绑定（骨架无节点可接线）。
	if _, ok := Resolve("global", "empty", []byte(`metadata: {name: empty, scope: global}
spec: {}
`)); ok {
		t.Fatal("零相位应不可绑定")
	}
}

func TestSkeletonShape(t *testing.T) {
	b := Binding{Workflow: "wf-x", Scope: "s1", Connecto: "wf-x-flow", Nodes: []Node{
		{Name: "analyze", Title: "Analyze", Unit: "session"},
		{Name: "dispatch", Unit: "agent"},
	}}
	s := string(Skeleton(b))
	for _, want := range []string{
		"kind: ConnectorOps", "name: wf-x-flow", "scope: s1",
		"core/origin: " + OriginLabel,
		"workservice: " + FlowWorkservice, "workflow: wf-x",
		"ops:\n    - " + FlowOp,
		"unit: session", "unit: agent",
	} {
		if !strings.Contains(s, want) {
			t.Fatalf("骨架缺 %q:\n%s", want, s)
		}
	}
	again := string(Skeleton(b))
	if again != s {
		t.Fatal("骨架非确定性（两次生成不一致）")
	}
}

// TestEnsureLifecycle 物化闭环：创建 → 幂等 noop → spec 变更 rv+1 → 所有面（非本物化器对象零盲写）。
func TestEnsureLifecycle(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	m := &Materializer{Repo: w.repo, Apply: w.engine}
	res := mustApply(t, w, "Workflow", "global", "flow-probe", 0, wfYAML("flow-probe", "global", "", twoPhases))

	// ① 创建。
	act, err := m.Ensure(ctx, "global", "flow-probe", wfYAML("flow-probe", "global", "", twoPhases))
	if err != nil || act != "created" {
		t.Fatalf("首物化: act=%s err=%v", act, err)
	}
	obj, err := w.repo.Get(ctx, "ConnectorOps", "global", "flow-probe")
	if err != nil {
		t.Fatalf("骨架对象缺席: %v", err)
	}
	if obj.ResourceVersion != res.RV*0+1 && obj.ResourceVersion != 1 {
		// 骨架 rv 与 workflow 无关——自身首版=1
		t.Fatalf("骨架 rv=%d（期望 1）", obj.ResourceVersion)
	}
	if originOf(obj.Content) != OriginLabel {
		t.Fatal("骨架对象缺 origin=workflow-bind 标注")
	}

	// ② 幂等重入：同内容 → noop（rv 不动）。
	act, err = m.Ensure(ctx, "global", "flow-probe", wfYAML("flow-probe", "global", "", twoPhases))
	if err != nil || act != "noop" {
		t.Fatalf("幂等重入: act=%s err=%v", act, err)
	}
	obj2, _ := w.repo.Get(ctx, "ConnectorOps", "global", "flow-probe")
	if obj2.ResourceVersion != 1 {
		t.Fatalf("noop 后 rv 前进=%d", obj2.ResourceVersion)
	}

	// ③ spec 变更（+1 相位）再物化 → updated（rv+1，diff 可查=apply_log ok 行）。
	three := twoPhases + "    - {name: report, title: Report}\n"
	mustApply(t, w, "Workflow", "global", "flow-probe", 1, wfYAML("flow-probe", "global", "", three))
	act, err = m.Ensure(ctx, "global", "flow-probe", wfYAML("flow-probe", "global", "", three))
	if err != nil || act != "updated" {
		t.Fatalf("变更物化: act=%s err=%v", act, err)
	}
	obj3, _ := w.repo.Get(ctx, "ConnectorOps", "global", "flow-probe")
	if obj3.ResourceVersion != 2 || !strings.Contains(string(obj3.Content), "name: report") {
		t.Fatalf("变更后 rv=%d 内容缺新节点", obj3.ResourceVersion)
	}
	nOK, _ := w.ops.CountApply("ok")
	if nOK < 3 { // workflow create + 骨架 create + 骨架 update（+workflow update）
		t.Fatalf("apply_log ok 行=%d（期望 ≥3——绑定走 apply 闸门可 diff）", nOK)
	}

	// ④ 所有面：origin≠workflow-bind 的对象（人写的实现）→ 零盲写。
	mustApply(t, w, "ConnectorOps", "global", "human-flow", 0, []byte(`apiVersion: core/v1
kind: ConnectorOps
metadata:
  name: human-flow
  scope: global
spec:
  workservice: workflow-engine
  workflow: human-wf
  ops: [flow]
`))
	act, err = m.Ensure(ctx, "global", "human-wf", wfYAML("human-wf", "global", "human-flow", twoPhases))
	if err != nil || act != "skipped" {
		t.Fatalf("人写对象应跳过: act=%s err=%v", act, err)
	}
	obj4, _ := w.repo.Get(ctx, "ConnectorOps", "global", "human-flow")
	if obj4.ResourceVersion != 1 || strings.Contains(string(obj4.Content), "nodes:") {
		t.Fatal("人写对象被盲写（所有面破缺）")
	}

	// ⑤ 零相位 workflow → skipped（无节点可接线）。
	mustApply(t, w, "Workflow", "global", "bare-wf", 0, []byte(`apiVersion: core/v1
kind: Workflow
metadata: {name: bare-wf, scope: global}
spec: {enabled: true}
`))
	if act, _ := m.Ensure(ctx, "global", "bare-wf", []byte("spec: {enabled: true}\n")); act != "skipped" {
		t.Fatalf("零相位应 skipped: act=%s", act)
	}
}

// TestEnsureAllAndHotPath 存量枚举初值 + DocApplied×Workflow 热路径（过滤+积压合并）。
func TestEnsureAllAndHotPath(t *testing.T) {
	w := newWorld(t)
	ctx := context.Background()
	mustApply(t, w, "Workflow", "global", "wf-a", 0, wfYAML("wf-a", "global", "", twoPhases))
	mustApply(t, w, "Workflow", "global", "wf-b", 0, wfYAML("wf-b", "global", "", twoPhases))
	m := &Materializer{Repo: w.repo, Apply: w.engine}

	n, fails := m.EnsureAll(ctx)
	if len(fails) != 0 {
		t.Fatalf("存量枚举失败: %+v", fails)
	}
	if n != 2 {
		t.Fatalf("存量枚举动作数=%d（期望 2）", n)
	}
	// 再跑=幂等（全 noop）零动作。
	if n2, _ := m.EnsureAll(ctx); n2 != 0 {
		t.Fatalf("二次枚举动作数=%d（期望 0——幂等 noop）", n2)
	}

	// 热路径：apply 第三个 workflow → DocApplied×Workflow 命中 → 骨架出现。
	applied := make(chan int, 4)
	errCh := make(chan []Fail, 4)
	runCtx, cancel := context.WithCancel(ctx)
	defer cancel()
	go m.Start(runCtx, w.bus, func(n int) { applied <- n }, func(f []Fail) { errCh <- f })
	mustApply(t, w, "Workflow", "global", "wf-c", 0, wfYAML("wf-c", "global", "", twoPhases))
	select {
	case n := <-applied:
		if n != 1 {
			t.Fatalf("热路径动作数=%d（期望 1）", n)
		}
	case f := <-errCh:
		t.Fatalf("热路径失败: %+v", f)
	case <-time.After(3 * time.Second):
		t.Fatal("热路径超时（DocApplied 未触发物化）")
	}
	if _, err := w.repo.Get(ctx, "ConnectorOps", "global", "wf-c"); err != nil {
		t.Fatalf("热路径骨架缺席: %v", err)
	}

	// 过滤：非 Workflow apply 不触发（applied 通道零到达）。
	mustApply(t, w, "WorkService", "global", "ws-x", 0, []byte(`apiVersion: core/v1
kind: WorkService
metadata: {name: ws-x, scope: global}
spec:
  category: tool
  connector: {types: [t], impl: {kind: source, ref: r/x}, download: pkg://r/x, llmPrompt: p}
  tools: [{op: flow, params: {node: string}, idempotent: true, undo: none}]
`))
	select {
	case n := <-applied:
		t.Fatalf("非 Workflow apply 触发了物化（过滤破缺）n=%d", n)
	case <-time.After(300 * time.Millisecond):
	}
}

func TestExecutionProjection(t *testing.T) {
	w := newWorld(t)
	mustApply(t, w, "Workflow", "global", "wf-g", 0, wfYAML("wf-g", "global", "", twoPhases))
	mustApply(t, w, "Workflow", "s1", "wf-o", 0, wfYAML("wf-o", "s1", "", twoPhases))
	// bound 对象三 scope 形态：global / s1（workflow 本 scope）/ s2（跨 scope 兜底）。
	skel := func(name, scope string) []byte {
		return Skeleton(Binding{Workflow: name, Scope: scope, Connecto: name,
			Nodes: []Node{{Name: "analyze", Unit: "session"}}})
	}
	mustApply(t, w, "ConnectorOps", "global", "wf-g", 0, skel("wf-g", "global"))
	mustApply(t, w, "ConnectorOps", "s1", "wf-o", 0, skel("wf-o", "s1"))

	rows := IndexRows(mustList(t, w))
	// bound（global 对象，骨架标注 → skeleton=true）。
	e := ExecutionOf("global", "wf-g", mustGet(t, w, "Workflow", "global", "wf-g"), rows)
	if e.State != ExecBound || e.Connector != "wf-g" || e.ConnectorScope != "global" || !e.Skeleton {
		t.Fatalf("wf-g execution: %+v", e)
	}
	// bound（本 scope 对象优先）。
	e = ExecutionOf("s1", "wf-o", mustGet(t, w, "Workflow", "s1", "wf-o"), rows)
	if e.State != ExecBound || e.ConnectorScope != "s1" {
		t.Fatalf("wf-o execution: %+v", e)
	}
	// unbound（绑定名解析成功而对象缺席——缺口可见）。
	e = ExecutionOf("global", "absent-wf", wfYAML("absent-wf", "global", "", twoPhases), rows)
	if e.State != ExecUnbound || e.Connector != "absent-wf" || e.ConnectorScope != "" {
		t.Fatalf("absent-wf execution: %+v", e)
	}
	// spec.connector 覆写投影进 execution。
	e = ExecutionOf("global", "ovr", []byte(`metadata: {name: ovr, scope: global}
spec:
  connector: connector-ops/wf-g
  phases:
    - analyze
`), rows)
	if e.State != ExecBound || e.Connector != "wf-g" {
		t.Fatalf("覆写 execution: %+v", e)
	}
}

func mustList(t *testing.T, w *world) []*repository.Object {
	t.Helper()
	objs, err := w.repo.List(context.Background(), "ConnectorOps", "")
	if err != nil {
		t.Fatalf("枚举 ConnectorOps: %v", err)
	}
	return objs
}

func mustGet(t *testing.T, w *world, kind, scope, name string) []byte {
	t.Helper()
	o, err := w.repo.Get(context.Background(), kind, scope, name)
	if err != nil {
		t.Fatalf("get %s/%s/%s: %v", kind, scope, name, err)
	}
	return o.Content
}
