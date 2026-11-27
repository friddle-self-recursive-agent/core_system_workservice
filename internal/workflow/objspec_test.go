// objspec_test.go —— 对象 spec 解析面机械锚（v0.27 表投影首段+v0.28 相位投影与迁移判定
// 接线，workflows 规则 16）。
package workflow

import (
	"strings"
	"testing"
	"time"
)

const objSpecDoc = `apiVersion: core/v1
kind: Workflow
metadata: {name: obj-probe-wf, scope: global}
spec:
  transitions:
  - {from: Idle, on: advance, to: Analyzing}
  - {from: Idle, on: ApprovalRejected, to: Failed}
  - {from: Analyzing, on: advance, to: Deploying}
`

// objWiredDoc 相位+声明表齐备的对象（v0.28 判定接线面 fixture）：相位词汇=Agent 八相位
// +Deploying 基座内取词（闸锚 AwaitingApproval/终点锚 Failed 在词汇内——防御面自然生效）。
const objWiredDoc = `apiVersion: core/v1
kind: Workflow
metadata: {name: obj-wired-wf, scope: global}
spec:
  phases: [Idle, Analyzing, AwaitingApproval, Executing, Failed]
  transitions:
  - {from: Idle, on: advance, to: Analyzing}
  - {from: Analyzing, on: advance, to: AwaitingApproval}
  - {from: AwaitingApproval, on: ApprovalGranted, to: Executing}
  - {from: AwaitingApproval, on: ApprovalRejected, to: Failed}
`

// TestDocsSpecsParse 声明→投影纯映射：行集=spec 声明且顺序保留；非相邻行原样保留
// （投影面不解释语义）；未命中/损坏/未声明=零投影。
func TestDocsSpecsParse(t *testing.T) {
	docs := map[string]string{
		"obj-probe-wf": objSpecDoc,
		"no-table":     "apiVersion: core/v1\nkind: Workflow\nmetadata: {name: no-table}\nspec: {}\n",
		"broken":       "spec: [\n",
	}
	ds := DocsSpecs{Lookup: func(name string) ([]byte, bool) {
		raw, ok := docs[name]
		return []byte(raw), ok
	}}
	view := ds.ResolveSpec("obj-probe-wf")
	if len(view.Transitions) != 3 {
		t.Fatalf("应投影 3 行声明表: %+v", view)
	}
	want := []TransitionRow{
		{From: "Idle", On: "advance", To: "Analyzing"},
		{From: "Idle", On: "ApprovalRejected", To: "Failed"},
		{From: "Analyzing", On: "advance", To: "Deploying"},
	}
	for i, w := range want {
		if view.Transitions[i] != w {
			t.Fatalf("行 %d 应 %v 实得 %v（顺序保留∧行集=声明）", i, w, view.Transitions[i])
		}
	}
	if len(view.Phases) != 0 {
		t.Fatalf("未声明 phases=零相位投影（不冒充空声明）: %v", view.Phases)
	}
	if v := ds.ResolveSpec("no-table"); v.Phases != nil || v.Transitions != nil {
		t.Fatalf("未声明=零投影（不冒充空表声明）: %+v", v)
	}
	if v := ds.ResolveSpec("broken"); v.Phases != nil || v.Transitions != nil {
		t.Fatalf("损坏文档=零投影（防御面）: %+v", v)
	}
	if v := ds.ResolveSpec("absent"); v.Phases != nil || v.Transitions != nil {
		t.Fatalf("未命中=零投影: %+v", v)
	}
	if v := ds.ResolveSpec(""); v.Phases != nil || v.Transitions != nil {
		t.Fatalf("空锚=零投影（防御面）: %+v", v)
	}
	if v := (DocsSpecs{}).ResolveSpec("obj-probe-wf"); v.Phases != nil || v.Transitions != nil {
		t.Fatalf("lookup 未注入=零投影（防御面）: %+v", v)
	}
	// SpecFunc 适配器同契约。
	if len(SpecFunc(func(string) SpecView { return view }).ResolveSpec("x").Transitions) != 3 {
		t.Fatalf("SpecFunc 适配器应透传投影")
	}
}

// TestDocsSpecsPhases 相位序列投影（v0.28）：spec.phases 声明序列原样顺序保留；相位与表
// 同源同拍（单次 lookup 两面齐投）；只声明其一=另一面零投影。
func TestDocsSpecsPhases(t *testing.T) {
	lookupCount := 0
	ds := DocsSpecs{Lookup: func(name string) ([]byte, bool) {
		lookupCount++
		if name == "obj-wired-wf" {
			return []byte(objWiredDoc), true
		}
		if name == "phases-only" {
			return []byte("apiVersion: core/v1\nkind: Workflow\nmetadata: {name: phases-only}\nspec:\n  phases: [Idle, Analyzing]\n"), true
		}
		return nil, false
	}}
	view := ds.ResolveSpec("obj-wired-wf")
	wantPhases := "Idle|Analyzing|AwaitingApproval|Executing|Failed"
	if got := strings.Join(view.Phases, "|"); got != wantPhases {
		t.Fatalf("相位序列应=声明顺序保留: %s", got)
	}
	if len(view.Transitions) != 4 {
		t.Fatalf("相位与表应同源齐投（4 行）: %+v", view)
	}
	if lookupCount != 1 {
		t.Fatalf("单次 lookup 两面齐投（零缓存不双读）: %d", lookupCount)
	}
	only := ds.ResolveSpec("phases-only")
	if strings.Join(only.Phases, "|") != "Idle|Analyzing" || only.Transitions != nil {
		t.Fatalf("只声明 phases=相位投影∧表零投影: %+v", only)
	}
}

// TestObjectRunTransitionsProjected 投影首段判别器（v0.27 基线，obj-probe-wf 未声明
// phases）：object 源 run 声明表=文档库 spec 声明行投影 ∧ 相位零投影 ∧ status 恒
// triggered（零投影口径不稀释——判定闸门=声明面）。
func TestObjectRunTransitionsProjected(t *testing.T) {
	ds := DocsSpecs{Lookup: func(name string) ([]byte, bool) {
		if name == "obj-probe-wf" {
			return []byte(objSpecDoc), true
		}
		return nil, false
	}}
	runs := NewRunLog(&RunOptions{Specs: ds})
	runs.StartRun(TriggerRecord{
		Seq: 7, EventID: "ev-obj", Type: "DocApplied", DedupKey: "apply:x:7",
		ReceivedAt: time.Now(), Route: "workflows/obj-probe-wf", Owner: "probe",
	}, ExecutorBinding{Executor: ExecWorkflow, Target: "obj-probe-wf"})
	r := runs.State().Runs[0]
	if r.SpecSource != SpecObject {
		t.Fatalf("object 源: %+v", r)
	}
	if len(r.Transitions) != 3 {
		t.Fatalf("声明表应投影 3 行: %+v", r.Transitions)
	}
	if r.Transitions[1].On != "ApprovalRejected" || r.Transitions[1].To != "Failed" {
		t.Fatalf("非相邻驳回行应原样投影（不解释语义）: %+v", r.Transitions[1])
	}
	if r.Phases == nil || len(r.Phases) != 0 {
		t.Fatalf("未声明 phases=零相位投影（空数组）: %#v", r.Phases)
	}
	if r.Phase != "" || r.Status != RunTriggered {
		t.Fatalf("零投影=零判定：status 恒 triggered 相位恒空: %+v", r)
	}
}

// TestObjectRunTransitionsMiss 解析面未命中/未注入=既有行为面（空数组零迁移）。
func TestObjectRunTransitionsMiss(t *testing.T) {
	ds := DocsSpecs{Lookup: func(string) ([]byte, bool) { return nil, false }}
	for _, tc := range []struct {
		name  string
		opts  *RunOptions
		descr string
	}{
		{"miss", &RunOptions{Specs: ds}, "解析面未命中"},
		{"nil-specs", nil, "解析面未注入（缺省行为面）"},
	} {
		runs := NewRunLog(tc.opts)
		runs.StartRun(TriggerRecord{
			Seq: 1, EventID: "ev", Type: "DocApplied", DedupKey: "k",
			ReceivedAt: time.Now(), Route: "workflows/nowhere", Owner: "probe",
		}, ExecutorBinding{Executor: ExecWorkflow, Target: "nowhere"})
		r := runs.State().Runs[0]
		if r.Transitions == nil || len(r.Transitions) != 0 {
			t.Fatalf("%s: object 源声明表应空数组: %+v", tc.descr, r.Transitions)
		}
		if r.Status != RunTriggered || r.Phase != "" {
			t.Fatalf("%s: 零迁移口径不变: %+v", tc.descr, r)
		}
	}
}

// TestObjectRunAdvanceStillZero 未声明 phases 的 object 源 run 异步推进零动作（零投影
// 口径不稀释——判定闸门=声明面的推进侧锚）。
func TestObjectRunAdvanceStillZero(t *testing.T) {
	ds := DocsSpecs{Lookup: func(name string) ([]byte, bool) {
		if name == "obj-probe-wf" {
			return []byte(objSpecDoc), true
		}
		return nil, false
	}}
	runs := NewRunLog(&RunOptions{Specs: ds})
	runs.StartRun(TriggerRecord{
		Seq: 1, EventID: "ev", Type: "DocApplied", DedupKey: "k",
		ReceivedAt: time.Now(), Route: "workflows/obj-probe-wf", Owner: "probe",
	}, ExecutorBinding{Executor: ExecWorkflow, Target: "obj-probe-wf"})
	n := runs.Advance()
	if n != 0 {
		t.Fatalf("未声明 phases 的 object 源 run 不被异步推进（零投影=零动作）: advanced=%d", n)
	}
	r := runs.State().Runs[0]
	if r.Status != RunTriggered || r.Phase != "" {
		t.Fatalf("推进零动作后口径不变: %+v", r)
	}
	if !strings.Contains(r.Workflow, "obj-probe-wf") {
		t.Fatalf("锚保留: %+v", r)
	}
}

// TestObjectRunJudgmentWired 迁移判定接线（v0.28，workflows 规则 16 下一段）：声明了
// phases 的 object 源 run 参与同一查表判定——受理首步（on=advance）推进∧status 翻
// running；异步推进随拍沿表前进∧落闸锚挂起即停；批准恢复续行；序列尾零动作停；驳回
// 终态。判定器零改动（stepByTable 落点通用，投影即点亮）。
func TestObjectRunJudgmentWired(t *testing.T) {
	ds := DocsSpecs{Lookup: func(name string) ([]byte, bool) {
		if name == "obj-wired-wf" {
			return []byte(objWiredDoc), true
		}
		return nil, false
	}}
	runs := NewRunLog(&RunOptions{Specs: ds})
	runs.StartRun(TriggerRecord{
		Seq: 1, EventID: "ev-wired", Type: "DocApplied", DedupKey: "apply:x:1",
		ReceivedAt: time.Now(), Route: "workflows/obj-wired-wf", Owner: "probe",
	}, ExecutorBinding{Executor: ExecWorkflow, Target: "obj-wired-wf"})
	snap := func() RunRecord { return runs.State().Runs[0] }
	r := snap()
	if r.Phases == nil || strings.Join(r.Phases, "|") != "Idle|Analyzing|AwaitingApproval|Executing|Failed" {
		t.Fatalf("object 源相位序列应投影（顺序保留）: %#v", r.Phases)
	}
	if r.Phase != string(Analyzing) || r.Status != RunRunning {
		t.Fatalf("受理首步应查表推进（Idle→Analyzing running）: %+v", r)
	}
	// 异步推进：Analyzing→AwaitingApproval=落闸锚挂起即停。
	if n := runs.Advance(); n != 1 {
		t.Fatalf("异步推进应 1 步: %d", n)
	}
	r = snap()
	if r.Phase != string(AwaitingApproval) || r.Status != RunSuspended {
		t.Fatalf("落闸锚应挂起: %+v", r)
	}
	if n := runs.Advance(); n != 0 {
		t.Fatalf("挂起面不前进: %d", n)
	}
	// 审批提升幂等（已挂起）。
	runs.ResumeSignal(TriggerRecord{Type: SignalSuspend, Involved: "Workflow/global/obj-wired-wf", ReceivedAt: time.Now()})
	r = snap()
	if r.Phase != string(AwaitingApproval) || r.Status != RunSuspended {
		t.Fatalf("重复提升应幂等零动作: %+v", r)
	}
	// 批准恢复：查表命中批准边（闸锚 on=ApprovalGranted → Executing）。
	runs.ResumeSignal(TriggerRecord{Type: SignalResume, Involved: "Workflow/global/obj-wired-wf", ReceivedAt: time.Now()})
	r = snap()
	if r.Phase != string(Executing) || r.Status != RunRunning {
		t.Fatalf("批准应沿声明表续行到 Executing: %+v", r)
	}
	// 序列尾（Executing 无 advance 行命中）=落 converged（F-04 正向末步收束——
	// object 源声明表与 builtin 同一判定器：advance 边止于末步即收束，不冒充仍在执行）。
	if n := runs.Advance(); n != 0 {
		t.Fatalf("序列尾收束拍应零迁移计数: %d", n)
	}
	r = snap()
	if r.Phase != string(Executing) || r.Status != RunConverged {
		t.Fatalf("序列尾应落 {Executing, converged}（F-04）: %+v", r)
	}
	if runs.State().Advances != 1 {
		t.Fatalf("累计推进计数应=1: %d", runs.State().Advances)
	}

	// 驳回终态：第二条 run 审批提升（running 态落闸锚）后 ApprovalRejected → 终点锚 Failed。
	runs.StartRun(TriggerRecord{
		Seq: 2, EventID: "ev-wired2", Type: "DocApplied", DedupKey: "apply:x:2",
		ReceivedAt: time.Now(), Route: "workflows/obj-wired-wf", Owner: "probe",
	}, ExecutorBinding{Executor: ExecWorkflow, Target: "obj-wired-wf"})
	runs.ResumeSignal(TriggerRecord{Type: SignalSuspend, Involved: "Workflow/global/obj-wired-wf", ReceivedAt: time.Now()})
	runs2 := runs.State().Runs
	r2 := runs2[len(runs2)-1]
	if r2.Phase != string(AwaitingApproval) || r2.Status != RunSuspended {
		t.Fatalf("审批提升应对 object run 生效（相位落闸锚挂起）: %+v", r2)
	}
	runs.ResumeSignal(TriggerRecord{Type: SignalReject, Involved: "Workflow/global/obj-wired-wf", ReceivedAt: time.Now()})
	runs2 = runs.State().Runs
	r2 = runs2[len(runs2)-1]
	if r2.Phase != string(Failed) || r2.Status != RunFailed {
		t.Fatalf("驳回应落终点锚终态: %+v", r2)
	}
	if n := runs.Advance(); n != 0 {
		t.Fatalf("终态不迁移: %d", n)
	}
}

// TestObjectRunPhasesNoTable 只声明 phases 未声明表：相位投影（声明首相位起步）但无从
// advance（表零投影）→ status 恒 triggered（诚实投影：无迁移行=零迁移，不冒充推进）。
func TestObjectRunPhasesNoTable(t *testing.T) {
	ds := DocsSpecs{Lookup: func(name string) ([]byte, bool) {
		if name == "phases-only" {
			return []byte("apiVersion: core/v1\nkind: Workflow\nmetadata: {name: phases-only}\nspec:\n  phases: [Idle, Analyzing]\n"), true
		}
		return nil, false
	}}
	runs := NewRunLog(&RunOptions{Specs: ds})
	runs.StartRun(TriggerRecord{
		Seq: 1, EventID: "ev", Type: "DocApplied", DedupKey: "k",
		ReceivedAt: time.Now(), Route: "workflows/phases-only", Owner: "probe",
	}, ExecutorBinding{Executor: ExecWorkflow, Target: "phases-only"})
	r := runs.State().Runs[0]
	if strings.Join(r.Phases, "|") != "Idle|Analyzing" {
		t.Fatalf("相位序列应投影: %#v", r.Phases)
	}
	if r.Phase != string(Idle) || r.Status != RunTriggered {
		t.Fatalf("表零投影=受理投影止于声明首相位（查表未命中零迁移）: %+v", r)
	}
	if n := runs.Advance(); n != 0 {
		t.Fatalf("无 advance 行=零动作: %d", n)
	}
}

// TestCoreSystemDeveloperLoopMultiEntryAndInterpreter 验证「Workflow+Scheme 为声明层、代码为通用解释器」：
//  1. core-system-developer-loop 富节点声明、多节点触发入口（entryPoints / entry:DocApplied）、
//     agent 自动判断审批（approval.mode=agent）、收尾 postActions（bugDocPolicy=invalidate|delete|keep）投影；
//  2. 入口 1（BugReport）从 Idle→Analyzing 启动，经 agent 自动审批不挂起直通 Planning；
//  3. 入口 2（DocApplied）直接从 Planning（doc-apply 节点）切入并推进至 Executing（developer-system）。
func TestCoreSystemDeveloperLoopMultiEntryAndInterpreter(t *testing.T) {
	if !IsBuiltin("core-system-developer-loop") {
		t.Fatal("core-system-developer-loop 应注册为 builtin")
	}
	const richLoopDoc = `apiVersion: core/v1
kind: Workflow
metadata: {name: core-system-developer-loop, scope: global}
spec:
  enabled: true
  approval: {mode: agent, autoCriteria: "单测全绿且无高危变更自动放行"}
  entryPoints:
    - {node: doc-apply, phase: Planning, enabled: true, events: [DocApplied]}
  phases:
    - {name: log-event,        phase: Idle,             title: "1. 监听日志与 Event", unit: connectops}
    - {name: agent-judge,      phase: Analyzing,        title: "2. Agent 判断",      unit: agent}
    - {name: approval-gate,    phase: AwaitingApproval, title: "3. 审批/自动审批",    unit: gate}
    - {name: doc-apply,        phase: Planning,         title: "4. Doc Apply",       unit: session, entryEvents: [DocApplied]}
    - {name: developer-system, phase: Executing,        title: "5. developer-system", unit: agent}
    - {name: verify,           phase: Verifying,        title: "6. 验收",            unit: connectops}
    - {name: deploy-system,    phase: Deploying,        title: "7. deploy-system",   unit: connectops}
    - {name: report-archive,   phase: Reporting,        title: "8. 总结归档",         unit: session}
  regression:
    statusWriteback: true
    bugDocPolicy: invalidate
    postActions:
      - {id: bugdoc-lifecycle, kind: doc-lifecycle, targetDocType: BugDoc, action: invalidate, enabled: true, title: "临时 BugDoc 标记失效"}
`
	ds := DocsSpecs{Lookup: func(name string) ([]byte, bool) {
		if name == "core-system-developer-loop" {
			return []byte(richLoopDoc), true
		}
		return nil, false
	}}
	view := ds.ResolveSpec("core-system-developer-loop")
	if view.ApprovalMode != "agent" || view.BugDocPolicy != "invalidate" || len(view.PostActions) != 1 {
		t.Fatalf("富声明解释器投影不符: %+v", view)
	}
	if len(view.Phases) != 8 || view.Phases[3] != "Planning" {
		t.Fatalf("富节点相位提取不符: %v", view.Phases)
	}

	runs := NewRunLog(&RunOptions{Specs: ds})
	// 入口 1：BugReport → Idle → Analyzing → (Advance: AwaitingApproval 自动通过不挂起) → Planning
	runs.StartRun(TriggerRecord{
		Seq: 1, EventID: "ev-bug-1", Type: "BugReport", DedupKey: "bug:1",
		ReceivedAt: time.Now(), Route: "workflows/core-system-developer-loop", Owner: "core",
	}, ExecutorBinding{Executor: ExecWorkflow, Target: "core-system-developer-loop"})
	r1 := runs.State().Runs[0]
	if r1.Phase != string(Analyzing) || r1.Status != RunRunning {
		t.Fatalf("入口 1 首步应落 Analyzing running: %+v", r1)
	}
	if n := runs.Advance(); n != 1 {
		t.Fatalf("推进应 1 步: %d", n)
	}
	r1 = runs.State().Runs[0]
	if r1.Phase != string(AwaitingApproval) || r1.Status != RunRunning {
		t.Fatalf("approval.mode=agent 落闸锚应自动通过保持 running: %+v", r1)
	}

	// 入口 2：DocApplied → 直接切入 Planning (节点 4 doc-apply) 并推进至 Executing (节点 5 developer-system)
	runs.StartRun(TriggerRecord{
		Seq: 2, EventID: "ev-doc-2", Type: "DocApplied", DedupKey: "apply:doc:2",
		ReceivedAt: time.Now(), Route: "workflows/core-system-developer-loop", Owner: "core",
	}, ExecutorBinding{Executor: ExecWorkflow, Target: "core-system-developer-loop"})
	r2 := runs.State().Runs[1]
	if r2.Phase != string(Executing) || r2.Status != RunRunning {
		t.Fatalf("入口 2 (DocApplied) 应从 Planning 切入并推进至 Executing: %+v", r2)
	}
	if len(r2.History) < 2 || r2.History[0].On != "entry:DocApplied" || r2.History[0].To != string(Planning) {
		t.Fatalf("入口 2 轨迹应记录 entry:DocApplied 切入 Planning: %+v", r2.History)
	}
}
