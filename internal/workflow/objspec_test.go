// objspec_test.go —— 对象 spec 声明表解析首段（v0.27，workflows 规则 16）机械锚。
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

// TestDocsSpecsParse 声明→投影纯映射：行集=spec 声明且顺序保留；非相邻行原样保留
//（判定未接线——投影面不解释语义）；未命中/损坏/未声明=零投影。
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
	rows := ds.ResolveSpec("obj-probe-wf")
	if len(rows) != 3 {
		t.Fatalf("应投影 3 行声明表: %+v", rows)
	}
	want := []TransitionRow{
		{From: "Idle", On: "advance", To: "Analyzing"},
		{From: "Idle", On: "ApprovalRejected", To: "Failed"},
		{From: "Analyzing", On: "advance", To: "Deploying"},
	}
	for i, w := range want {
		if rows[i] != w {
			t.Fatalf("行 %d 应 %v 实得 %v（顺序保留∧行集=声明）", i, w, rows[i])
		}
	}
	if ds.ResolveSpec("no-table") != nil {
		t.Fatalf("未声明表=零投影（不冒充空表声明）")
	}
	if ds.ResolveSpec("broken") != nil {
		t.Fatalf("损坏文档=零投影（防御面）")
	}
	if ds.ResolveSpec("absent") != nil {
		t.Fatalf("未命中=零投影")
	}
	if ds.ResolveSpec("") != nil {
		t.Fatalf("空锚=零投影（防御面）")
	}
	if (DocsSpecs{}).ResolveSpec("obj-probe-wf") != nil {
		t.Fatalf("lookup 未注入=零投影（防御面）")
	}
	// SpecFunc 适配器同契约。
	if len(SpecFunc(func(string) []TransitionRow { return want }).ResolveSpec("x")) != 3 {
		t.Fatalf("SpecFunc 适配器应透传投影行")
	}
}

// TestObjectRunTransitionsProjected 投影首段判别器：object 源 run 声明表=文档库 spec
// 声明行投影（transitions 从恒空数组变声明行集）∧ 相位仍不投影 ∧ status 恒 triggered
//（投影≠判定——判定接线顺延）。
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
		t.Fatalf("相位仍不投影（空数组）: %#v", r.Phases)
	}
	if r.Phase != "" || r.Status != RunTriggered {
		t.Fatalf("投影≠判定：status 恒 triggered 相位恒空: %+v", r)
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

// TestObjectRunAdvanceStillZero 异步推进对 object 源投影 run 仍零动作（相位不投影防御面
// 原样生效——投影≠判定的推进侧锚）。
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
		t.Fatalf("object 源投影 run 不被异步推进（相位不投影=零动作）: advanced=%d", n)
	}
	r := runs.State().Runs[0]
	if r.Status != RunTriggered || r.Phase != "" {
		t.Fatalf("推进零动作后口径不变: %+v", r)
	}
	if !strings.Contains(r.Workflow, "obj-probe-wf") {
		t.Fatalf("锚保留: %+v", r)
	}
}
