// proposal_test.go —— 需求提案分流/派单面验收（v0.51 能力 B——concepts/proposal.md §4/§5 +
// console-adoption.md §1.24 L52 判据的单测锚）：解析/状态翻转、批准派单（kind=proposal 行——
// eventId 同源）、驳回不派单、高危分流 ApprovalNeeded 自产（source=core）、低危自动批准
// （引擎自写 proposal:auto + granted 行 + 单环两拍无自激）、draft/landed 零动作。
package proposal

import (
	"context"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/apply"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/flowbind"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// 测试 scheme：Proposal kind 最小注册（枚举闭合集与权威声明同词汇）+ 审批族事件类型。
const testScheme = `apiVersion: core/v1
kind: Scheme
metadata: {name: test, scope: builtin}
spec:
  version: core/v1
  kinds:
    - {kind: Scheme, builtin: true, storage: core/scheme.yaml}
    - kind: Proposal
      builtin: true
      storage: "global/proposals/<name>.yaml | systems/<s>/proposals/<name>.yaml"
      spec:
        title: "!string"
        summary: "!string"
        acceptance: "[]string"
        impact: "?string"
        priority: "low|normal|high|critical"
        proposer: "!string"
        status: "!draft|submitted|approved|rejected|landed"
        landing: "?{target: doc|task, ref: ?string}"
  eventTypes:
    - {type: DocApplied, source: core, defaultLevel: info, route: owner-workflow}
    - {type: ApprovalNeeded, source: "permission|core", defaultLevel: warn, dedup: "<rule>:<obj>", route: notification/*}
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
	return &world{
		repo: repo, ops: ops, bus: bus,
		engine: &apply.Engine{
			Sch: sch, Repo: repo, Bus: bus, Ops: ops,
			Perm:    &permission.System{},
			Subject: permission.Subject{User: "cored:local", Roles: []string{permission.DefaultAdmin}},
		},
	}
}

func proposalYAML(name, status, priority string) []byte {
	return []byte(`apiVersion: core/v1
kind: Proposal
metadata: {name: ` + name + `, scope: global}
spec:
  title: 探针提案
  summary: 验证需求提交流闭环
  acceptance:
    - 提案对象在库
    - 批准后派单受理在册
  impact: probe-only（零真实影响面）
  priority: ` + priority + `
  proposer: console-operator
  status: ` + status + `
  landing: {target: doc}
`)
}

// waitFor 轮询断言（watcher 异步——2s 窗内条件成立即过；超时判 fail）。
func waitFor(t *testing.T, what string, cond func() bool) {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if cond() {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("超时等待 %s", what)
}

func mustApplyProp(t *testing.T, w *world, name, status, priority string, rv int64) *apply.Result {
	t.Helper()
	res, err := w.engine.Apply(context.Background(), &repository.Object{
		Kind: "Proposal", Name: name, Scope: "global", ResourceVersion: rv, Content: proposalYAML(name, status, priority),
	}, apply.Options{FieldManager: "test:propose"})
	if err != nil {
		t.Fatalf("apply Proposal/%s: %v", status, err)
	}
	return res
}

func TestParseAndFlipStatus(t *testing.T) {
	spec, ok := Parse(proposalYAML("probe-prop", "submitted", "high"))
	if !ok || spec.Title != "探针提案" || spec.Status != "submitted" || spec.Priority != "high" ||
		spec.Proposer != "console-operator" || spec.LandingTarget != "doc" ||
		len(spec.Acceptance) != 2 || spec.Acceptance[1] != "批准后派单受理在册" {
		t.Fatalf("spec 投影: ok=%v %+v", ok, spec)
	}
	// 无状态=非 Proposal 形。
	if _, ok := Parse([]byte("apiVersion: core/v1\nkind: Proposal\nmetadata: {name: x}\nspec: {title: t}\n")); ok {
		t.Fatal("无 status 应不可解析")
	}
	// 状态翻转：命中翻转 / 零命中如实 false / 同值拒翻。
	out, ok := FlipStatus(proposalYAML("probe-prop", "submitted", "low"), StatusSubmitted, StatusApproved)
	if !ok {
		t.Fatal("submitted→approved 应命中")
	}
	if spec2, _ := Parse(out); spec2.Status != "approved" || spec2.Title != "探针提案" {
		t.Fatalf("翻转后形态保持: %+v", spec2)
	}
	if _, ok := FlipStatus(proposalYAML("probe-prop", "submitted", "low"), StatusRejected, StatusApproved); ok {
		t.Fatal("from=rejected 未命中应 false")
	}
	if _, ok := FlipStatus(proposalYAML("probe-prop", "submitted", "low"), StatusSubmitted, StatusSubmitted); ok {
		t.Fatal("同值翻转应 false")
	}
}

// recSink 派单落点桩（记录投影行——断言 face 载荷同源；并发安全=watcher goroutine 写/测试读）。
type recSink struct {
	mu   sync.Mutex
	recs []flowbind.DispatchRecord
}

func (r *recSink) Dispatch(rec flowbind.DispatchRecord) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.recs = append(r.recs, rec)
}

func (r *recSink) count() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.recs)
}

func (r *recSink) last() flowbind.DispatchRecord {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.recs[len(r.recs)-1]
}

func startWatcher(t *testing.T, w *world, auto []string) (*recSink, func()) {
	t.Helper()
	sink := &recSink{}
	wat := &Watcher{
		Repo: w.repo, Apply: w.engine, Bus: w.bus, Sch: w.engine.Sch,
		Dispatch: sink, Approvals: w.ops, AutoApprove: auto,
	}
	ctx, cancel := context.WithCancel(context.Background())
	go wat.Start(ctx, w.bus, nil, func(err error) { t.Logf("watcher: %v", err) })
	return sink, cancel
}

func TestDispatchOnApproved(t *testing.T) {
	w := newWorld(t)
	sink, cancel := startWatcher(t, w, nil)
	defer cancel()
	res := mustApplyProp(t, w, "probe-prop", StatusApproved, "critical", 0)
	waitFor(t, "派单受理行", func() bool { return sink.count() > 0 })
	rec := sink.last()
	if rec.Agent != flowbind.DispatchAgent || rec.Status != flowbind.DispatchAccepted ||
		rec.AutoLevel != flowbind.DispatchAutoLevel || rec.Action != "approved" ||
		rec.Kind != flowbind.DispatchKindProposal || rec.Proposal != "probe-prop" ||
		rec.Scope != "global" || rec.RV != res.RV || rec.EventID != res.Event.ID ||
		rec.Title != "探针提案" || rec.Priority != "critical" || rec.Proposer != "console-operator" ||
		len(rec.Acceptance) != 2 {
		t.Fatalf("派单受理投影同源性: %+v", rec)
	}
}

func TestNoDispatchOnRejectedAndDraftAndLanded(t *testing.T) {
	w := newWorld(t)
	sink, cancel := startWatcher(t, w, nil)
	defer cancel()
	mustApplyProp(t, w, "probe-rej", StatusRejected, "high", 0)
	mustApplyProp(t, w, "probe-land", StatusLanded, "high", 0)
	time.Sleep(200 * time.Millisecond)
	if n := sink.count(); n != 0 {
		t.Fatalf("rejected/landed 不派单，实得 %d 行", n)
	}
}

func TestHumanTriageEmitsApprovalNeeded(t *testing.T) {
	w := newWorld(t)
	sink, cancel := startWatcher(t, w, nil) // 空清单=全人审
	defer cancel()
	ch, stop := w.bus.Subscribe("ApprovalNeeded")
	defer stop()
	mustApplyProp(t, w, "probe-prop", StatusSubmitted, "high", 0)
	select {
	case ev := <-ch:
		if ev.Type != "ApprovalNeeded" || ev.Source != "core" ||
			ev.Involved != "Proposal/global/probe-prop" || ev.DedupKey != "proposal:Proposal/global/probe-prop" ||
			ev.Level != "warn" {
			t.Fatalf("ApprovalNeeded 自产投影: %+v", ev)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("高危提交应自产 ApprovalNeeded（event.md 规则 15）")
	}
	time.Sleep(150 * time.Millisecond)
	if n := sink.count(); n != 0 {
		t.Fatalf("submitted（待人审）不派单，实得 %d 行", n)
	}
}

func TestAutoApproveLowRisk(t *testing.T) {
	w := newWorld(t)
	sink, cancel := startWatcher(t, w, []string{"low"})
	defer cancel()
	res := mustApplyProp(t, w, "probe-prop", StatusSubmitted, "low", 0)
	if res.Action != "create" || res.RV != 1 {
		t.Fatalf("提交落库: %+v", res)
	}
	// 单环两拍：提交落库（rv=1）→ 引擎自写批准（rv=2）→ 派单受理（rv=2，eventId=第二拍 DocApplied）。
	waitFor(t, "自动批准派单", func() bool {
		obj, err := w.repo.Get(context.Background(), "Proposal", "global", "probe-prop")
		if err != nil {
			return false
		}
		spec, ok := Parse(obj.Content)
		return ok && spec.Status == StatusApproved && obj.ResourceVersion == 2 && sink.count() == 1
	})
	rec := sink.last()
	if rec.Action != "approved" || rec.RV != 2 || rec.EventID == "" ||
		rec.EventID == res.Event.ID || rec.Priority != "low" {
		t.Fatalf("自动批准派单（eventId=第二拍 DocApplied 同源）: %+v", rec)
	}
	// 判定账面：granted 行 judgedBy=引擎主体 fieldManager=proposal:auto。
	rows, err := w.ops.QueryApprovals(opslog.ApprovalQuery{})
	if err != nil || len(rows) != 1 || rows[0].Decision != opslog.DecisionGranted ||
		rows[0].FieldManager != FieldManagerAuto || rows[0].JudgedBy != "cored:local" || rows[0].ResourceVersion != 2 {
		t.Fatalf("auto-approve granted 行: err=%v rows=%+v", err, rows)
	}
	// 无自激：环稳态（单行）——approved 不再写面。
	time.Sleep(150 * time.Millisecond)
	if n := sink.count(); n != 1 {
		t.Fatalf("自激检出（派单 %d 行）", n)
	}
}
