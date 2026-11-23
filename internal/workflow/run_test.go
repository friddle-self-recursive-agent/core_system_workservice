// run_test.go —— 对账 run 投影面验收（引擎主体第三段 v0.15——真实拉起第一子段）：
// 锚解析纯映射（对账锚→core-workflow、直接引用→builtin 注册表/平台对象可见）+
// run 环容量/丢最旧/dropped 可见 + 状态词与起点投影（相位态机 deferred 的诚实口径）。
package workflow

import (
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// TestResolveRunAnchor 锚解析（与 execFor 家族收编同构的纯映射验收）：
// 对账锚解析到对账循环锚 core-workflow（owner 维度不在此，由留痕承载）；
// 直接引用按 builtin 注册表解析，未命中=平台对象引用（锚保留、相位不投影）；
// 非 workflow 家族=空锚（防御性可见）。
func TestResolveRunAnchor(t *testing.T) {
	cases := []struct {
		route  string
		wf     string
		source string
		nPha   int
	}{
		{"owner-workflow", "core-workflow", SpecBuiltin, 8},                           // 对账锚→对账循环锚（8 相位）
		{"owner-instance-workflow", "core-workflow", SpecBuiltin, 8},                  // 同族词汇同锚
		{"workflows/watch-only", "watch-only", SpecBuiltin, 3},                        // 直接引用 builtin 命中
		{"workflows/core-workflow", "core-workflow", SpecBuiltin, 8},                  // 直接引用对账循环本体
		{"workflows/deploy-system-workflow", "deploy-system-workflow", SpecObject, 0}, // 未命中=平台对象引用
		{"workflows/builtin-root", "builtin-root", SpecObject, 0},                     // 同上（权威表词汇）
		{"workflows/dev-delivery", "dev-delivery", SpecObject, 0},                     // scheme builtinSet 六项 vs 注册表四项差额=缺口可见（不冒充 builtin）
		{"root-agent", "", "", 0},                                                     // 非 workflow 家族=空锚（不到达；防御性）
		{"dead-letter", "", "", 0},                                                    // 死信无 run（none 绑定面，event.md rule 4）
		{"notification/*", "", "", 0},                                                 // 通道送达面非本 sink 职责
		{"", "", "", 0},                                                               // 空 route=空锚
	}
	for _, c := range cases {
		wf, source, phases := resolveRunAnchor(c.route)
		if wf != c.wf || source != c.source || len(phases) != c.nPha {
			t.Errorf("resolveRunAnchor(%q)={%s,%s,%d} want {%s,%s,%d}",
				c.route, wf, source, len(phases), c.wf, c.source, c.nPha)
		}
	}
	// 对账锚的声明起点=core-workflow 首相位 Idle（起点投影可判）。
	_, _, phases := resolveRunAnchor("owner-workflow")
	if phases[0] != string(Idle) {
		t.Fatalf("对账 run 起点应为首相位 Idle，得 %s", phases[0])
	}
}

// TestIsBuiltin 注册表判定：已声明相位的四项收编；未落相位声明的（dev-delivery、
// apply-workflow，scheme builtinSet 六项差额）不冒充 builtin——缺口由 SpecObject 可见。
func TestIsBuiltin(t *testing.T) {
	for _, name := range []string{"core-workflow", "default-task", "simple-shot", "watch-only"} {
		if !IsBuiltin(name) {
			t.Errorf("IsBuiltin(%q) 应 true（注册表同源）", name)
		}
	}
	for _, name := range []string{"deploy-system-workflow", "builtin-root", "dev-delivery", "apply-workflow", ""} {
		if IsBuiltin(name) {
			t.Errorf("IsBuiltin(%q) 应 false（相位声明未落设计，不即兴发明）", name)
		}
	}
}

// TestRunLogRing 环容量：溢出丢最旧、dropped 计数可见、run 序号进程内单调。
func TestRunLogRing(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	runs := NewRunLog(&RunOptions{RingSize: 2})
	h.BindRuns(runs)
	h.WireExecutor(ExecWorkflow)
	defer h.AttachDispatch(b, []scheme.DispatchRule{{On: "DocApplied", For: "owner:*", Dispatch: "owner-workflow"}})()

	for i := 1; i <= 4; i++ {
		applyEvent(t, b, "Agent/global/probe", int64(i))
	}
	waitState(t, h, 4)
	st := runs.State()
	if st.Started != 4 || st.Dropped != 2 || len(st.Runs) != 2 {
		t.Fatalf("环口径应 started=4 dropped=2 len=2: %+v", st)
	}
	if st.Runs[0].Run != 3 || st.Runs[1].Run != 4 {
		t.Fatalf("丢最旧后应剩 run 3、4: %+v", st.Runs)
	}
	// run↔留痕可回指：TriggerSeq 与留痕 seq 同源。
	if st.Runs[1].TriggerSeq != 4 {
		t.Fatalf("run 应携来源留痕 seq=4: %+v", st.Runs[1])
	}
}

// TestRunStartProjection 起点 projection 的诚实口径：status=triggered、phase=声明首相位、
// 状态词不含任何「已完成/执行中」语义（相位态机 deferred——run 出现 ≠ 对账执行完毕）。
func TestRunStartProjection(t *testing.T) {
	runs := NewRunLog(nil)
	runs.StartRun(TriggerRecord{
		Seq: 7, EventID: "ev1", Type: "DocApplied", DedupKey: "apply:x:1",
		ReceivedAt: time.Now(), Route: "owner-workflow", Owner: "global",
	}, ExecutorBinding{Executor: ExecWorkflow, Target: "owner"})
	st := runs.State()
	if len(st.Runs) != 1 {
		t.Fatalf("应 1 条 run: %+v", st)
	}
	r := st.Runs[0]
	if r.Status != RunTriggered {
		t.Fatalf("run 状态应为起点词 triggered: %+v", r)
	}
	if r.Workflow != ReconcileWorkflow || r.SpecSource != SpecBuiltin || r.Phase != string(Idle) {
		t.Fatalf("对账锚 run 应 {core-workflow,builtin,Idle}: %+v", r)
	}
	if r.TriggerSeq != 7 || r.Owner != "global" || r.Target != "owner" || r.Route != "owner-workflow" {
		t.Fatalf("run 溯源字段不符: %+v", r)
	}
	if len(r.Phases) == 0 || r.Phases[len(r.Phases)-1] != string(Failed) {
		t.Fatalf("相位序列应携声明全序列（尾=Failed）: %v", r.Phases)
	}
}
