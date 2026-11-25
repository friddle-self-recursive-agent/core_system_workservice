// execution_test.go —— 构建生效链判别投影单测（v0.55，R-9——console-adoption.md §1.29）：
// chain 三词闭合集（consistent/divergent/absent）、绑定行过滤（fieldManager=workflow:bind ∧
// reasonRef 前缀=workflow-bind:）、rv 对账（ok=落库后版本/noop=面上现值）、NodesOf 宽松解析、
// digest 单点真相。纯函数面——零 IO 零时钟依赖。
package flowbind

import (
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
)

func bindApply(result string, rv int64, ts time.Time) opslog.ApplyLog {
	return opslog.ApplyLog{Ts: ts, FieldManager: fieldManager,
		Obj: "ConnectorOps/global/flow-probe", ResourceVersion: rv, Result: result,
		ReasonRef: BindReasonPrefix + "flow-probe"}
}

func TestReconcileConsistentOnOkRow(t *testing.T) {
	conn := &repository.Object{Kind: "ConnectorOps", Name: "flow-probe", Scope: "global",
		ResourceVersion: 3, Content: []byte("apiVersion: core/v1\nkind: ConnectorOps\nmetadata: {name: flow-probe, scope: global, labels: {core/origin: workflow-bind}}\nspec:\n  workservice: workflow-engine\n  workflow: flow-probe\n  ops: [flow]\n  flow:\n    nodes:\n      - {name: analyze, unit: session}\n      - {name: report, unit: agent, provider: cc}\n")}
	rc := ReconcileOf(&Execution{State: ExecBound, Connector: "flow-probe", ConnectorScope: "global", Skeleton: true},
		conn, []opslog.ApplyLog{bindApply("ok", 3, time.Now())})
	if rc.Chain != ChainConsistent || rc.SkeletonRV != 3 || !rc.Bound || !rc.Skeleton {
		t.Fatalf("chain 应 consistent: %+v", rc)
	}
	if rc.LastBindApply != "ok" || rc.LastBindRV != 3 || rc.LastBindAt == "" {
		t.Fatalf("绑定行投影不符: %+v", rc)
	}
	if len(rc.Nodes) != 2 || rc.Nodes[0].Name != "analyze" || rc.Nodes[1].Unit != "agent" || rc.Nodes[1].Provider != "cc" {
		t.Fatalf("骨架节点投影不符: %+v", rc.Nodes)
	}
}

func TestReconcileConsistentOnNoopRow(t *testing.T) {
	conn := &repository.Object{Kind: "ConnectorOps", Name: "flow-probe", Scope: "global",
		ResourceVersion: 2, Content: []byte("spec: {flow: {nodes: [{name: n1}]}}\n")}
	// noop 幂等重入照落面上现值 rv=2 → 一致（同窗）。
	rc := ReconcileOf(&Execution{State: ExecBound, Skeleton: true}, conn,
		[]opslog.ApplyLog{bindApply("noop", 2, time.Now())})
	if rc.Chain != ChainConsistent || rc.LastBindApply != "noop" || rc.LastBindRV != 2 {
		t.Fatalf("noop 行应同窗一致: %+v", rc)
	}
}

func TestReconcileDivergentOnForeignWrite(t *testing.T) {
	conn := &repository.Object{Kind: "ConnectorOps", Name: "flow-probe", Scope: "global",
		ResourceVersion: 5, Content: []byte("spec: {}\n")}
	// 骨架翻转为实现后：面上 rv=5、最近绑定行 rv=2 → divergent（链分裂可见不静默）。
	rc := ReconcileOf(&Execution{State: ExecBound, Skeleton: false}, conn,
		[]opslog.ApplyLog{bindApply("ok", 2, time.Now())})
	if rc.Chain != ChainDivergent || rc.SkeletonRV != 5 || rc.LastBindRV != 2 {
		t.Fatalf("人写演进应 divergent: %+v", rc)
	}
	if rc.Skeleton {
		t.Fatalf("execution.skeleton=false 应复读为 !Skeleton: %+v", rc)
	}
}

func TestReconcileAbsentOnMissingObject(t *testing.T) {
	rc := ReconcileOf(&Execution{State: ExecUnbound, Connector: "core-workflow"}, nil, nil)
	if rc.Chain != ChainAbsent || rc.SkeletonRV != 0 || rc.Bound {
		t.Fatalf("缺席应 absent: %+v", rc)
	}
}

func TestReconcileIgnoresForeignFieldManager(t *testing.T) {
	conn := &repository.Object{Kind: "ConnectorOps", Name: "flow-probe", Scope: "global",
		ResourceVersion: 4, Content: []byte("spec: {}\n")}
	rows := []opslog.ApplyLog{
		{FieldManager: "console:editor", Obj: "ConnectorOps/global/flow-probe", ResourceVersion: 4, Result: "ok"},
		{FieldManager: fieldManager, Obj: "ConnectorOps/global/flow-probe", ResourceVersion: 3, Result: "ok",
			ReasonRef: "other:" + "x"}, // 绑定 manager 但 reasonRef 非本物化器——过滤
		bindApply("ok", 3, time.Now()),
	}
	rc := ReconcileOf(&Execution{State: ExecBound, Skeleton: true}, conn, rows)
	if rc.LastBindRV != 3 || rc.Chain != ChainDivergent {
		t.Fatalf("外来行应被过滤（人写编辑行/他锚 reasonRef 行——最近绑定行 rv=3，面上 rv=4=divergent）: %+v", rc)
	}
}

func TestReconcileDivergentWhenBindRowMissing(t *testing.T) {
	conn := &repository.Object{Kind: "ConnectorOps", Name: "flow-probe", Scope: "global",
		ResourceVersion: 4, Content: []byte("spec: {}\n")}
	rc := ReconcileOf(&Execution{State: ExecBound, Skeleton: false}, conn,
		[]opslog.ApplyLog{{FieldManager: "console:editor", ResourceVersion: 4, Result: "ok"}})
	if rc.Chain != ChainDivergent || rc.LastBindApply != "" {
		t.Fatalf("绑定行缺席应如实分裂: %+v", rc)
	}
}

func TestNodesOfLenient(t *testing.T) {
	if got := NodesOf([]byte("not: [valid: yaml\n")); got != nil {
		t.Fatalf("不可解析应零节点: %v", got)
	}
	got := NodesOf([]byte("apiVersion: core/v1\nkind: ConnectorOps\nspec:\n  flow:\n    nodes:\n      - name: a\n        unit: session\n      - {name: b, unit: agent, connector: other, op: flow}\n"))
	if len(got) != 2 || got[1].Connector != "other" || got[1].Op != "flow" {
		t.Fatalf("节点投影不符: %+v", got)
	}
}

func TestDigestSingleTruth(t *testing.T) {
	// digest 单点真相：同内容两次判别 rv 恒等（指纹稳定——消费方可比对）。
	conn := &repository.Object{Kind: "ConnectorOps", Name: "x", Scope: "global",
		ResourceVersion: 7, Content: []byte("spec: {}\n")}
	r1 := ReconcileOf(&Execution{State: ExecBound, Skeleton: true}, conn, []opslog.ApplyLog{bindApply("ok", 7, time.Now())})
	r2 := ReconcileOf(&Execution{State: ExecBound, Skeleton: true}, conn, []opslog.ApplyLog{bindApply("ok", 7, time.Now())})
	if r1.SkeletonRV != r2.SkeletonRV || r1.Chain != r2.Chain {
		t.Fatalf("digest 应单点真相: %+v vs %+v", r1, r2)
	}
}
