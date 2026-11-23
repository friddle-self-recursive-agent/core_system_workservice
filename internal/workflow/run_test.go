// run_test.go —— 对账 run 投影面验收（引擎主体第三段 v0.15——真实拉起第一子段；
// v0.23 相位推进器首段——迁移判定=声明序列单调前进）：
// 锚解析纯映射（对账锚→core-workflow、直接引用→builtin 注册表/平台对象可见）+
// run 环容量/丢最旧/dropped 可见 + 受理即启动口径（builtin=running/第二相位、object=triggered 不推进）
// + 迁移判定纯函数（尾相位/越界/空序列不前进）。
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
		{"workflows/dev-delivery", "dev-delivery", SpecBuiltin, 9},                    // v0.19 注册表六项对齐：交付态机 9 相位（含 Deploying）
		{"workflows/apply-workflow", "apply-workflow", SpecBuiltin, 8},                // 落地 agent 任务态机=default-task 词汇全量复用
		{"workflows/deploy-system-workflow", "deploy-system-workflow", SpecObject, 0}, // 未命中=平台对象引用（注册表外不收编，锚保留）
		{"workflows/builtin-root", "builtin-root", SpecObject, 0},                     // 同上（权威表词汇）
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

// TestIsBuiltin 注册表判定：scheme builtinSet 六项全量对齐（v0.19——声明面相位词汇映射
// 先行落设计，注册表随后收编，声明↔注册表同源）；注册表外不冒充 builtin。
func TestIsBuiltin(t *testing.T) {
	for _, name := range []string{"core-workflow", "default-task", "simple-shot", "watch-only", "dev-delivery", "apply-workflow"} {
		if !IsBuiltin(name) {
			t.Errorf("IsBuiltin(%q) 应 true（注册表与 scheme builtinSet 六项同源）", name)
		}
	}
	for _, name := range []string{"deploy-system-workflow", "builtin-root", ""} {
		if IsBuiltin(name) {
			t.Errorf("IsBuiltin(%q) 应 false（注册表外不收编）", name)
		}
	}
}

// TestBuiltinSequences 六项声明相位序列机械锚（concepts/workflow.md §1 映射表的代码同源面）：
// 序列、首尾相位、dev-delivery 的 Deploying 锚位、apply-workflow 与 default-task 同族全等。
func TestBuiltinSequences(t *testing.T) {
	want := map[string][]Phase{
		"core-workflow":  {Idle, Analyzing, Planning, AwaitingApproval, Executing, Verifying, Reporting, Failed},
		"default-task":   {Idle, Analyzing, Planning, AwaitingApproval, Executing, Verifying, Reporting, Failed},
		"simple-shot":    {Idle, Executing, Reporting},
		"watch-only":     {Idle, Analyzing, Reporting},
		"dev-delivery":   {Idle, Analyzing, AwaitingApproval, Planning, Executing, Verifying, Deploying, Reporting, Failed},
		"apply-workflow": {Idle, Analyzing, Planning, AwaitingApproval, Executing, Verifying, Reporting, Failed},
	}
	for name, seq := range want {
		got := Builtin(name)
		if got.Name != name || len(got.Phases) != len(seq) {
			t.Fatalf("Builtin(%q) 序列长度不符: %+v", name, got)
		}
		for i, p := range seq {
			if got.Phases[i] != p {
				t.Errorf("Builtin(%q).Phases[%d]=%s want %s", name, i, got.Phases[i], p)
			}
		}
	}
	// 声明面锚点：dev-delivery 的 Deploying=交付相位（§1.1 Deploy 步），序列尾恒 Failed。
	if Builtin("dev-delivery").Phases[6] != Deploying {
		t.Fatalf("dev-delivery 相位锚位 6 应为 Deploying: %v", Builtin("dev-delivery").Phases)
	}
	// apply-workflow 任务态机=default-task 词汇全量复用（不另立词汇——解耦口径）。
	a, d := Builtin("apply-workflow").Phases, Builtin("default-task").Phases
	if len(a) != len(d) {
		t.Fatalf("apply-workflow 与 default-task 序列应同族全等: %v vs %v", a, d)
	}
	for i := range a {
		if a[i] != d[i] {
			t.Errorf("apply-workflow[%d]=%s ≠ default-task[%d]=%s", i, a[i], i, d[i])
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

// TestRunStartProjection 受理投影的诚实口径（v0.23 相位推进器首段）：受理即启动——引擎
// 单步迁移判定把 builtin run 从声明首相位沿序列推进一步（Idle→Analyzing）、状态词翻转
// triggered→running；相位序列仍携声明全序列（尾=Failed）。「run 出现 ≠ 对账执行完毕」
// 口径不变：running=首步进行中，非执行完毕。
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
	if r.Status != RunRunning {
		t.Fatalf("builtin run 受理即启动，状态应为 running: %+v", r)
	}
	if r.Workflow != ReconcileWorkflow || r.SpecSource != SpecBuiltin || r.Phase != string(Analyzing) {
		t.Fatalf("对账锚 run 应 {core-workflow,builtin,Analyzing}（声明首相位 Idle 沿序列前进一步）: %+v", r)
	}
	if r.TriggerSeq != 7 || r.Owner != "global" || r.Target != "owner" || r.Route != "owner-workflow" {
		t.Fatalf("run 溯源字段不符: %+v", r)
	}
	if len(r.Phases) == 0 || r.Phases[len(r.Phases)-1] != string(Failed) {
		t.Fatalf("相位序列应携声明全序列（尾=Failed）: %v", r.Phases)
	}
}

// TestObjectRunPhasesEmptyArray object 源 run 的相位载荷形状：锚保留、相位不投影，
// 载荷=空数组（非 null——消费方可 len，null=缺席语义两分）。
func TestObjectRunPhasesEmptyArray(t *testing.T) {
	runs := NewRunLog(nil)
	runs.StartRun(TriggerRecord{
		Seq: 3, EventID: "ev9", Type: "DocApplied", DedupKey: "apply:x:9",
		ReceivedAt: time.Now(), Route: "workflows/deploy-system-workflow", Owner: "builtin",
	}, ExecutorBinding{Executor: ExecWorkflow, Target: "deploy-system-workflow"})
	r := runs.State().Runs[0]
	if r.Phases == nil || len(r.Phases) != 0 {
		t.Fatalf("object 源 run 相位应为空数组（非 nil）: %#v", r.Phases)
	}
	if r.SpecSource != SpecObject || r.Phase != "" {
		t.Fatalf("object 源 run 口径 {object, 空相位起点}: %+v", r)
	}
	// 相位不投影=无从迁移：object 源 run 不被推进，status 恒受理词（v0.23 口径）。
	if r.Status != RunTriggered {
		t.Fatalf("object 源 run 状态应恒 triggered（相位不投影无从迁移）: %+v", r)
	}
}

// TestAdvancePhase 迁移判定纯函数（v0.23 相位推进器首段）：沿声明序列单调前进一步；
// 序列尾/不在序列内/空序列不前进（受理态保留）。
func TestAdvancePhase(t *testing.T) {
	seq := phaseNames(Builtin("core-workflow").Phases) // Idle→…→Failed
	if p, ok := advancePhase(seq, "Idle"); !ok || p != string(Analyzing) {
		t.Fatalf("Idle 应前进一步到 Analyzing: %s %v", p, ok)
	}
	if p, ok := advancePhase(seq, string(Verifying)); !ok || p != string(Reporting) {
		t.Fatalf("Verifying 应前进一步到 Reporting: %s %v", p, ok)
	}
	// 序列尾（终点相位）不前进：ok=false 原相位保留。
	if p, ok := advancePhase(seq, string(Failed)); ok || p != string(Failed) {
		t.Fatalf("序列尾 Failed 不应前进: %s %v", p, ok)
	}
	// 当前相位不在序列内（含空串）与空序列：不前进。
	if _, ok := advancePhase(seq, "Nowhere"); ok {
		t.Fatalf("不在序列内的相位不应前进")
	}
	if _, ok := advancePhase(nil, "Idle"); ok {
		t.Fatalf("空序列不应前进")
	}
}
