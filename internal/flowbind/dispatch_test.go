// dispatch_test.go —— apply-workflow agent 派单面验收（v0.50，R-8——workflow-runtime.md §11.2
// 拉起形态首段）：骨架物化动作（created|updated）即派单受理投影；noop/skipped 不派单；
// 投影字段（agent/status/autoLevel/action/workflow/scope/connector/rv/eventId）与
// apply 回执同源；环有界 dropped 可观测；受理≠拉起（无会话语义——诚实投影）。
package flowbind

import (
	"context"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/apply"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
)

func TestDispatchOnMaterialized(t *testing.T) {
	w := newWorld(t)
	dl := NewDispatchLog(nil)
	m := &Materializer{Repo: w.repo, Apply: w.engine, Dispatch: dl}
	ctx := context.Background()

	// 创建物化 → 派单受理（created，rv=1，eventId=DocApplied 回执）。
	act, err := m.Ensure(ctx, "global", "flow-probe", wfYAML("flow-probe", "global", "", twoPhases))
	if err != nil || act != "created" {
		t.Fatalf("首物化: act=%s err=%v", act, err)
	}
	st := dl.State()
	if st.Accepted != 1 || len(st.Records) != 1 {
		t.Fatalf("受理计数: %+v", st)
	}
	r := st.Records[0]
	if r.Seq != 1 || r.Agent != DispatchAgent || r.Status != DispatchAccepted || r.AutoLevel != DispatchAutoLevel {
		t.Fatalf("投影锚/词汇: %+v", r)
	}
	if r.Action != "created" || r.Workflow != "flow-probe" || r.Scope != "global" || r.Connector != "flow-probe" || r.RV != 1 {
		t.Fatalf("投影字段: %+v", r)
	}
	if r.EventID == "" {
		t.Fatalf("eventId 缺席（DocApplied 回执应同源）: %+v", r)
	}
	if r.At.IsZero() || time.Since(r.At) > time.Minute {
		t.Fatalf("受理时刻异常: %+v", r)
	}

	// spec 变更再物化 → 派单（updated，rv=2）。
	act, err = m.Ensure(ctx, "global", "flow-probe", wfYAML("flow-probe", "global", "", twoPhases+"    - {name: report, title: Report}\n"))
	if err != nil || act != "updated" {
		t.Fatalf("更新物化: act=%s err=%v", act, err)
	}
	st = dl.State()
	if st.Accepted != 2 || st.Records[1].Action != "updated" || st.Records[1].RV != 2 {
		t.Fatalf("更新派单: accepted=%d rec1=%+v", st.Accepted, st.Records[1])
	}

	// 幂等重入（noop）→ 不派单（无新实现可做）。
	if _, err := m.Ensure(ctx, "global", "flow-probe", wfYAML("flow-probe", "global", "", twoPhases+"    - {name: report, title: Report}\n")); err != nil {
		t.Fatalf("幂等物化: %v", err)
	}
	if st2 := dl.State(); st2.Accepted != 2 {
		t.Fatalf("noop 不派单: accepted=%d", st2.Accepted)
	}

	// 零相位（不可解析）→ skipped 不派单。
	if _, err := m.Ensure(ctx, "global", "wf-empty", []byte("metadata: {name: wf-empty, scope: global}\nspec: {phases: []}\n")); err != nil {
		t.Fatalf("零相位: %v", err)
	}
	if st2 := dl.State(); st2.Accepted != 2 {
		t.Fatalf("skipped 不派单: accepted=%d", st2.Accepted)
	}
}

func TestDispatchSkipsOwnedFace(t *testing.T) {
	// 所有面：人写 ConnectorOps（origin≠workflow-bind）→ 物化器 skipped → 零派单（零盲写语义随行）。
	w := newWorld(t)
	dl := NewDispatchLog(nil)
	m := &Materializer{Repo: w.repo, Apply: w.engine, Dispatch: dl}
	ctx := context.Background()
	owned := []byte("apiVersion: core/v1\nkind: ConnectorOps\nmetadata: {name: owned-flow, scope: global}\nspec:\n  workservice: workflow-engine\n  workflow: owned-wf\n  ops: [flow]\n")
	if _, err := w.engine.Apply(ctx, &repository.Object{Kind: "ConnectorOps", Name: "owned-flow", Scope: "global", ResourceVersion: 0, Content: owned},
		apply.Options{FieldManager: "test:owned"}); err != nil {
		t.Fatalf("人写对象: %v", err)
	}
	act, err := m.Ensure(ctx, "global", "owned-wf", wfYAML("owned-wf", "global", "owned-flow", twoPhases))
	if err != nil || act != "skipped" {
		t.Fatalf("所有面 skip: act=%s err=%v", act, err)
	}
	if st := dl.State(); st.Accepted != 0 {
		t.Fatalf("skipped 零派单: accepted=%d", st.Accepted)
	}
}

func TestDispatchNilSinkSilent(t *testing.T) {
	// sink 缺席（headless/旧装配）→ 静默零派单零 panic（物化照常）。
	w := newWorld(t)
	m := &Materializer{Repo: w.repo, Apply: w.engine}
	act, err := m.Ensure(context.Background(), "global", "flow-probe", wfYAML("flow-probe", "global", "", twoPhases))
	if err != nil || act != "created" {
		t.Fatalf("静默物化: act=%s err=%v", act, err)
	}
}

func TestDispatchRingBounded(t *testing.T) {
	// 环有界：溢出丢最旧 + dropped 计数可见（WakeLog 同款语义）。
	dl := NewDispatchLog(&DispatchOptions{RingSize: 3})
	for i := 0; i < 5; i++ {
		dl.Dispatch(DispatchRecord{Agent: DispatchAgent, Status: DispatchAccepted, AutoLevel: DispatchAutoLevel, Action: "created", Workflow: "wf", At: time.Now()})
	}
	st := dl.State()
	if st.Accepted != 5 || st.Dropped != 2 || len(st.Records) != 3 {
		t.Fatalf("环界: accepted=%d dropped=%d n=%d", st.Accepted, st.Dropped, len(st.Records))
	}
	if st.Records[0].Seq != 3 || st.Records[2].Seq != 5 {
		t.Fatalf("环序: first=%d last=%d", st.Records[0].Seq, st.Records[2].Seq)
	}
}

func TestDispatchStateSnapshotIsolated(t *testing.T) {
	// State 拷贝隔离：快照后追加不影响既有快照（查询面安全持有）。
	dl := NewDispatchLog(nil)
	dl.Dispatch(DispatchRecord{Agent: DispatchAgent, At: time.Now()})
	s1 := dl.State()
	dl.Dispatch(DispatchRecord{Agent: DispatchAgent, At: time.Now()})
	if len(s1.Records) != 1 || dl.State().Accepted != 2 {
		t.Fatalf("快照隔离: s1=%d live=%d", len(s1.Records), dl.State().Accepted)
	}
	if DispatchAgent != "apply-workflow" || DispatchAutoLevel != "ask" || DispatchAccepted != "accepted" {
		t.Fatalf("锚/词汇常量漂移: %s/%s/%s", DispatchAgent, DispatchAutoLevel, DispatchAccepted)
	}
}
