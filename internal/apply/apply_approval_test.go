// apply_approval_test.go —— ApprovalNeeded 自产面（v0.21 引擎内部真实事件源首段，
// event.md 规则 9 + permission.md §4 人审档机读判定）：写路径权限准入拒绝 ∧ 人审档对象
// （Role|Group=授权面装载类）→ 引擎自产 ApprovalNeeded 进总线（source=permission、
// level=注册表 defaultLevel 投影、dedupKey=permission:<obj> 窗口聚合）；非人审档拒绝/
// dry-run/类型未注册/Bus 缺席 → 不产事件且拒绝结论不变。
package apply

import (
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"gopkg.in/yaml.v3"
)

// approvalScheme：Role/Group/Agent 三 kind + ApprovalNeeded 注册（defaultLevel=warn）。
const approvalScheme = `apiVersion: core/v1
kind: Scheme
metadata: {name: approval, scope: builtin}
spec:
  version: core/v1
  kinds:
    - {kind: Scheme, builtin: true, storage: core/scheme.yaml}
    - {kind: Role, builtin: true, storage: global/roles/<name>.yaml}
    - {kind: Group, builtin: true, storage: global/groups/<name>.yaml}
    - {kind: Agent, builtin: true, storage: global/agents/<name>/agent.yaml}
  eventTypes:
    - {type: DocApplied, source: core, defaultLevel: info, route: owner-workflow}
    - {type: ApprovalNeeded, source: permission, defaultLevel: warn, dedup: "<rule>:<obj>", route: notification/*}
`

// 无 ApprovalNeeded 注册的对照 scheme（注册表闸门=引擎内部源唯一开关）。
const noApprovalScheme = `apiVersion: core/v1
kind: Scheme
metadata: {name: noapproval, scope: builtin}
spec:
  version: core/v1
  kinds:
    - {kind: Scheme, builtin: true, storage: core/scheme.yaml}
    - {kind: Role, builtin: true, storage: global/roles/<name>.yaml}
    - {kind: Agent, builtin: true, storage: global/agents/<name>/agent.yaml}
  eventTypes:
    - {type: DocApplied, source: core, defaultLevel: info, route: owner-workflow}
`

func roleYAML(name string) []byte {
	return []byte(`apiVersion: core/v1
kind: Role
metadata: {name: ` + name + `, scope: global}
spec:
  permissions:
  - {scope: "global/*", actions: [read]}
`)
}

// newApprovalWorld 组装：受限主体（只持 read）+ 注册 ApprovalNeeded 的 scheme + 短去重窗口。
type approvalWorld struct {
	ops *opslog.SQLiteStore
	bus *eventbus.Bus
	eng *Engine
	sub *permission.Subject
}

func newApprovalWorld(t *testing.T, schemeYAML string) *approvalWorld {
	t.Helper()
	sch, err := scheme.Load([]byte(schemeYAML))
	if err != nil {
		t.Fatalf("载入 scheme: %v", err)
	}
	dir := t.TempDir()
	ops, err := opslog.OpenSQLite(filepath.Join(dir, "ops.db"))
	if err != nil {
		t.Fatalf("ops DB: %v", err)
	}
	t.Cleanup(func() { _ = ops.Close() })
	bus := eventbus.NewBus(&eventbus.Options{Ops: ops, DedupWindow: 80 * time.Millisecond})
	w := &approvalWorld{ops: ops, bus: bus}
	w.eng = &Engine{
		Sch: sch, Bus: bus, Ops: ops, Repo: nil,
		Perm:    &permission.System{},
		Subject: permission.Subject{User: "cored:local", Roles: []string{permission.DefaultAdmin}},
	}
	// 受限请求级主体：持 role/reader（read only）——对 Role/Agent 的写全拒。
	w.sub = &permission.Subject{User: "u1", Roles: []string{"role/reader"}}
	return w
}

// mustObject 构造写对象（kind/scope/name 从文档投影；本组测试只走拒绝路径不落库）。
func mustObject(t *testing.T, content []byte) *repository.Object {
	t.Helper()
	var doc struct {
		Kind     string `yaml:"kind"`
		Metadata struct {
			Name  string `yaml:"name"`
			Scope string `yaml:"scope"`
		} `yaml:"metadata"`
	}
	if err := yaml.Unmarshal(content, &doc); err != nil {
		t.Fatalf("解析对象: %v", err)
	}
	return &repository.Object{Kind: doc.Kind, Scope: doc.Metadata.Scope, Name: doc.Metadata.Name, Content: content}
}

// drain 非阻塞收取通道内全部事件（Publish 在 Apply 返回前同步投递，无竞态等待）。
func drain(ch <-chan eventbus.Event) []eventbus.Event {
	var out []eventbus.Event
	for {
		select {
		case e := <-ch:
			out = append(out, e)
		default:
			return out
		}
	}
}

func TestHumanReviewKindPredicate(t *testing.T) {
	if !permission.HumanReviewKind("Role") || !permission.HumanReviewKind("Group") {
		t.Fatalf("授权面装载类 Role/Group 应命中人审档")
	}
	for _, k := range []string{"Agent", "Workflow", "role", "", "Provider"} {
		if permission.HumanReviewKind(k) {
			t.Fatalf("kind %q 不应命中人审档（首段只收 Role|Group）", k)
		}
	}
}

func TestApprovalNeededOnHumanReviewDeny(t *testing.T) {
	w := newApprovalWorld(t, approvalScheme)
	ch, cancel := w.bus.Subscribe("")
	defer cancel()
	o := mustObject(t, roleYAML("hijack-role"))
	_, err := w.eng.Apply(t.Context(), o, Options{FieldManager: "fm-1", Subject: w.sub})
	if !errors.Is(err, ErrDenied) {
		t.Fatalf("权限拒绝应返回 ErrDenied: %v", err)
	}
	evts := drain(ch)
	if len(evts) != 1 {
		t.Fatalf("人审档拒绝应自产恰一条事件: got %d", len(evts))
	}
	e := evts[0]
	if e.Type != "ApprovalNeeded" || e.Source != "permission" {
		t.Fatalf("类型/来源错: %+v", e)
	}
	if e.Involved != "Role/global/hijack-role" {
		t.Fatalf("involved 应为对象键: %q", e.Involved)
	}
	if e.DedupKey != "permission:Role/global/hijack-role" {
		t.Fatalf("dedupKey 应为 permission:<obj>: %q", e.DedupKey)
	}
	if e.Level != "warn" {
		t.Fatalf("level 应为注册表 defaultLevel=warn: %q", e.Level)
	}
	if e.ID == "" {
		t.Fatalf("事件 ID 未归一化: %+v", e)
	}
	// 拒绝路径审计行照落（权威留痕）。
	if n, err := w.ops.CountApply("denied"); err != nil || n != 1 {
		t.Fatalf("denied 审计行应恰 1: n=%d err=%v", n, err)
	}
	// event_log 归档（同主干零旁路；source=permission 不入库校验在 source 恒随归档行）。
	if n, err := w.ops.CountEvent("ApprovalNeeded"); err != nil || n != 1 {
		t.Fatalf("event_log 应恰 1 条 ApprovalNeeded: n=%d err=%v", n, err)
	}
}

func TestApprovalNeededNotForNonHumanReviewKind(t *testing.T) {
	w := newApprovalWorld(t, approvalScheme)
	ch, cancel := w.bus.Subscribe("")
	defer cancel()
	o := mustObject(t, agentYAML("sneaky-agent", ""))
	_, err := w.eng.Apply(t.Context(), o, Options{FieldManager: "fm-1", Subject: w.sub})
	if !errors.Is(err, ErrDenied) {
		t.Fatalf("非人审档写仍应拒绝: %v", err)
	}
	if evts := drain(ch); len(evts) != 0 {
		t.Fatalf("非人审档拒绝不产事件: %+v", evts)
	}
	if n, err := w.ops.CountEvent("ApprovalNeeded"); err != nil || n != 0 {
		t.Fatalf("event_log 不应出现 ApprovalNeeded: n=%d err=%v", n, err)
	}
}

func TestApprovalNeededDryRunNoEvent(t *testing.T) {
	w := newApprovalWorld(t, approvalScheme)
	ch, cancel := w.bus.Subscribe("")
	defer cancel()
	o := mustObject(t, roleYAML("probe-role"))
	_, err := w.eng.Apply(t.Context(), o, Options{FieldManager: "fm-1", Subject: w.sub, DryRun: true})
	if !errors.Is(err, ErrDenied) {
		t.Fatalf("dry-run 拒绝结论应一致: %v", err)
	}
	if evts := drain(ch); len(evts) != 0 {
		t.Fatalf("dry-run 零副作用不产事件: %+v", evts)
	}
}

func TestApprovalNeededNotRegisteredNoEvent(t *testing.T) {
	w := newApprovalWorld(t, noApprovalScheme)
	ch, cancel := w.bus.Subscribe("")
	defer cancel()
	o := mustObject(t, roleYAML("any-role"))
	_, err := w.eng.Apply(t.Context(), o, Options{FieldManager: "fm-1", Subject: w.sub})
	if !errors.Is(err, ErrDenied) {
		t.Fatalf("拒绝结论不受注册表影响: %v", err)
	}
	if evts := drain(ch); len(evts) != 0 {
		t.Fatalf("类型未注册不发布（注册表闸门）: %+v", evts)
	}
}

func TestApprovalNeededDedupWindowAggregates(t *testing.T) {
	w := newApprovalWorld(t, approvalScheme)
	ch, cancel := w.bus.Subscribe("")
	defer cancel()
	o := mustObject(t, roleYAML("dup-role"))
	opt := Options{FieldManager: "fm-1", Subject: w.sub}
	for i := 0; i < 3; i++ {
		if _, err := w.eng.Apply(t.Context(), o, opt); !errors.Is(err, ErrDenied) {
			t.Fatalf("第 %d 次拒绝: %v", i+1, err)
		}
	}
	if evts := drain(ch); len(evts) != 1 {
		t.Fatalf("窗口内同 obj 三次拒绝应聚合为 1 条: got %d", len(evts))
	}
	time.Sleep(120 * time.Millisecond) // 出窗
	if _, err := w.eng.Apply(t.Context(), o, opt); !errors.Is(err, ErrDenied) {
		t.Fatalf("出窗后拒绝: %v", err)
	}
	if evts := drain(ch); len(evts) != 1 {
		t.Fatalf("出窗后应新产一条: got %d", len(evts))
	}
}

func TestApprovalNeededPublishFailureDenyStands(t *testing.T) {
	// Bus=nil：发布面缺席（如只读装配直接构造引擎），拒绝结论不变。
	sch, err := scheme.Load([]byte(approvalScheme))
	if err != nil {
		t.Fatalf("载入 scheme: %v", err)
	}
	eng := &Engine{
		Sch: sch, Bus: nil, Perm: &permission.System{},
		Subject: permission.Subject{User: "cored:local", Roles: []string{permission.DefaultAdmin}},
	}
	o := mustObject(t, roleYAML("x-role"))
	if _, err := eng.Apply(t.Context(), o, Options{FieldManager: "fm-1",
		Subject: &permission.Subject{User: "u1", Roles: []string{"role/reader"}}}); !errors.Is(err, ErrDenied) {
		t.Fatalf("Bus 缺席时拒绝结论不变: %v", err)
	}
}
