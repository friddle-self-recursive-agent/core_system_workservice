// run_test.go —— 对账 run 投影面验收（引擎主体第三段 v0.15——真实拉起第一子段；
// v0.23 相位推进器首段——迁移判定=声明序列单调前进；v0.25 显式迁移表首段——判定器=声明表
// 查表）：锚解析纯映射（对账锚→core-workflow、直接引用→builtin 注册表/平台对象可见）+
// run 环容量/丢最旧/dropped 可见 + 受理即启动口径（builtin=running/第二相位、object=triggered
// 不推进）+ 声明表物化锚（相邻行+审批边/非相邻驳回边）+ 判定器查表（尾相位/越界/空表零迁移）。
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
		wf, source, phases, trans := resolveRunAnchor(c.route)
		if wf != c.wf || source != c.source || len(phases) != c.nPha {
			t.Errorf("resolveRunAnchor(%q)={%s,%s,%d} want {%s,%s,%d}",
				c.route, wf, source, len(phases), c.wf, c.source, c.nPha)
		}
		// 声明表随锚投影（v0.25）：builtin 命中=物化表非空；object/空锚=不投影。
		if (len(trans) > 0) != (c.source == SpecBuiltin) {
			t.Errorf("resolveRunAnchor(%q) 声明表投影与来源不符：source=%s trans=%d",
				c.route, source, len(trans))
		}
	}
	// 对账锚的声明起点=core-workflow 首相位 Idle（起点投影可判）。
	_, _, phases, _ := resolveRunAnchor("owner-workflow")
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
	// 声明表随 run 投影（v0.25 显式迁移表首段）：物化表非空、首行=受理查表的推进行
	//（Idle,advance,Analyzing）、审批边在表（新载荷键=代际判别器的代码同源面）。
	if len(r.Transitions) == 0 {
		t.Fatalf("builtin run 应携物化声明表: %+v", r.Transitions)
	}
	if r.Transitions[0].From != string(Idle) || r.Transitions[0].On != EdgeAdvance || r.Transitions[0].To != string(Analyzing) {
		t.Fatalf("声明表首行应 {Idle,advance,Analyzing}: %+v", r.Transitions[0])
	}
	var hasReject bool
	for _, row := range r.Transitions {
		if row.On == SignalReject && row.From == string(AwaitingApproval) && row.To == string(Failed) {
			hasReject = true
		}
	}
	if !hasReject {
		t.Fatalf("声明表应携驳回边（非相邻表行先例）: %+v", r.Transitions)
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
	// 声明表同口径：object 源不投影=空数组（非 nil——spec 随文档面不在本进程）。
	if r.Transitions == nil || len(r.Transitions) != 0 {
		t.Fatalf("object 源 run 声明表应为空数组（非 nil）: %#v", r.Transitions)
	}
	if r.SpecSource != SpecObject || r.Phase != "" {
		t.Fatalf("object 源 run 口径 {object, 空相位起点}: %+v", r)
	}
	// 相位不投影=无从迁移：object 源 run 不被推进，status 恒受理词（v0.23 口径）。
	if r.Status != RunTriggered {
		t.Fatalf("object 源 run 状态应恒 triggered（相位不投影无从迁移）: %+v", r)
	}
}

// TestBuiltinTransitionTable 显式迁移表首段（v0.25，workflow.md 规则 14）——builtin 声明表
// 物化机械锚：表=声明序列物化相邻行（from=seq[i] on=advance to=seq[i+1]，行数=len-1）
// +审批边表行先例（含闸锚序列：批准边+驳回边；无闸锚序列=无审批边）；驳回边
// （AwaitingApproval→Failed）=表的首处**非相邻**迁移行（机械判别器）。
func TestBuiltinTransitionTable(t *testing.T) {
	// core-workflow：8 相位 → 7 相邻行 + 批准边 + 驳回边 = 9 行。
	cw := Builtin("core-workflow")
	if got := len(cw.Transitions); got != 9 {
		t.Fatalf("core-workflow 物化表应 9 行（7 相邻+批准+驳回），得 %d: %+v", got, cw.Transitions)
	}
	for i := 0; i+1 < len(cw.Phases); i++ {
		row := cw.Transitions[i]
		if row.From != cw.Phases[i] || row.On != EdgeAdvance || row.To != cw.Phases[i+1] {
			t.Errorf("相邻行 %d 应 {%s,advance,%s}: %+v", i, cw.Phases[i], cw.Phases[i+1], row)
		}
	}
	// 审批边=表行先例：批准边目标=闸锚后继（与相邻行同目标异事件名）；驳回边=非相邻终点锚。
	if to, ok := cw.Can(AwaitingApproval, SignalResume); !ok || to != Executing {
		t.Fatalf("批准边应查表命中 AwaitingApproval→Executing: %s %v", to, ok)
	}
	if to, ok := cw.Can(AwaitingApproval, SignalReject); !ok || to != Failed {
		t.Fatalf("驳回边应查表命中 AwaitingApproval→Failed（非相邻迁移按表命中——机械判别器）: %s %v", to, ok)
	}
	// Can 实接（v0.25 前 builtin 表恒空=永不命中）：相邻推进行与未知 on 的判定面。
	if to, ok := cw.Can(Idle, EdgeAdvance); !ok || to != Analyzing {
		t.Fatalf("推进行应查表命中 Idle→Analyzing: %s %v", to, ok)
	}
	if _, ok := cw.Can(Idle, SignalReject); ok {
		t.Fatalf("非闸锚源无驳回边，不应命中")
	}
	if _, ok := cw.Can(AwaitingApproval, "DocApplied"); ok {
		t.Fatalf("未知 on 不应命中（查表未命中=零迁移）")
	}
	// 非相邻迁移=表驱动而非序列推导的构造反证：手工 spec 声明非相邻推进行（Idle→Reporting），
	// 查表命中即迁移——判定器只认表行，不隐式走序列。
	synthetic := &Spec{
		Name: "synthetic", Phases: []Phase{Idle, Analyzing, Reporting, Failed},
		Transitions: []Transition{{From: Idle, On: EdgeAdvance, To: Reporting}},
	}
	if to, ok := synthetic.Can(Idle, EdgeAdvance); !ok || to != Reporting {
		t.Fatalf("非相邻推进行应按表命中 Idle→Reporting（不隐式走序列）: %s %v", to, ok)
	}
	// 无闸锚序列（simple-shot/watch-only）=无审批边（零动作面在表物化即收口）；
	// 无终点锚序列亦无驳回边。
	for _, name := range []string{"simple-shot", "watch-only"} {
		s := Builtin(name)
		wantRows := len(s.Phases) - 1
		if got := len(s.Transitions); got != wantRows {
			t.Fatalf("%s 无闸锚序列应恰 %d 相邻行（无审批边），得 %d", name, wantRows, got)
		}
		for _, row := range s.Transitions {
			if row.On != EdgeAdvance {
				t.Errorf("%s 不应出现审批边行: %+v", name, row)
			}
		}
	}
	// 六项 builtin 全量：有闸锚序列的审批边成对出现（批准+驳回）。
	for _, name := range []string{"core-workflow", "default-task", "dev-delivery", "apply-workflow"} {
		s := Builtin(name)
		var grant, reject bool
		for _, row := range s.Transitions {
			if row.From == AwaitingApproval && row.On == SignalResume && row.To == s.Phases[indexGate(s.Phases)+1] {
				grant = true
			}
			if row.From == AwaitingApproval && row.On == SignalReject && row.To == Failed {
				reject = true
			}
		}
		if !grant || !reject {
			t.Errorf("%s 审批边应成对物化（批准+驳回）: grant=%v reject=%v", name, grant, reject)
		}
	}
}

// indexGate 闸锚序位（辅助断言用）。
func indexGate(seq []Phase) int {
	for i, p := range seq {
		if p == AwaitingApproval {
			return i
		}
	}
	return -1
}

// TestStepByTable 判定器=纯查表（v0.25）：from+on 命中→to；未命中=零迁移（受理首步/
// 批准/驳回三路径共用同一落点）；空表与 nil 表=零迁移（object 源口径）。
func TestStepByTable(t *testing.T) {
	table := transitionRows(Builtin("core-workflow").Transitions)
	if p, ok := stepByTable(table, string(Idle), EdgeAdvance); !ok || p != string(Analyzing) {
		t.Fatalf("Idle 推进行应命中 Analyzing: %s %v", p, ok)
	}
	if p, ok := stepByTable(table, string(Verifying), EdgeAdvance); !ok || p != string(Reporting) {
		t.Fatalf("Verifying 推进行应命中 Reporting: %s %v", p, ok)
	}
	// 非相邻迁移按表命中（驳回边）。
	if p, ok := stepByTable(table, string(AwaitingApproval), SignalReject); !ok || p != string(Failed) {
		t.Fatalf("驳回边应命中 AwaitingApproval→Failed: %s %v", p, ok)
	}
	// 序列尾（终点相位）无推进行：零迁移。
	if p, ok := stepByTable(table, string(Failed), EdgeAdvance); ok || p != string(Failed) {
		t.Fatalf("序列尾 Failed 零迁移: %s %v", p, ok)
	}
	// 当前相位不在表内（含空串）/未知 on/空表/nil 表：零迁移。
	if _, ok := stepByTable(table, "Nowhere", EdgeAdvance); ok {
		t.Fatalf("不在表内的源相位零迁移")
	}
	if _, ok := stepByTable(table, "", EdgeAdvance); ok {
		t.Fatalf("空源相位零迁移")
	}
	if _, ok := stepByTable(nil, string(Idle), EdgeAdvance); ok {
		t.Fatalf("nil 表零迁移（object 源口径）")
	}
	if _, ok := stepByTable([]TransitionRow{}, string(Idle), EdgeAdvance); ok {
		t.Fatalf("空表零迁移")
	}
}

// TestApprovalSignalLifecycle 挂起-恢复首段全周期（v0.24，workflows 规则 13）：
// 受理即启动（v0.23 口径不变）→ 审批提升落闸锚挂起 → 批准沿声明序列续行（恢复）
// → 再挂起 → 驳回终态。全程**不另起新 run**（审批信号是相位机迁移事件，非对账触发）。
func TestApprovalSignalLifecycle(t *testing.T) {
	runs := NewRunLog(nil)
	runs.StartRun(TriggerRecord{
		Seq: 1, EventID: "ev1", Type: "DocApplied", DedupKey: "apply:x:1",
		ReceivedAt: time.Now(), Route: "workflows/dev-delivery", Owner: "global",
	}, ExecutorBinding{Executor: ExecWorkflow, Target: "dev-delivery"})
	sig := func(typ string) {
		runs.ResumeSignal(TriggerRecord{Seq: 100, Type: typ, Involved: "Workflow/builtin/dev-delivery", ReceivedAt: time.Now()})
	}
	want := func(phase, status string, when string) {
		t.Helper()
		r := runs.State().Runs[0]
		if r.Phase != phase || r.Status != status {
			t.Fatalf("%s 后应 {phase=%s, status=%s}: {phase=%s, status=%s}", when, phase, status, r.Phase, r.Status)
		}
	}
	want(string(Analyzing), RunRunning, "受理") // v0.23 基线：受理即启动
	sig(SignalSuspend)
	want(string(AwaitingApproval), RunSuspended, "审批提升") // 闸锚挂起
	sig(SignalResume)
	want(string(Planning), RunRunning, "批准恢复") // dev-delivery 闸锚序位 2 → 续行 Planning
	sig(SignalSuspend)
	want(string(AwaitingApproval), RunSuspended, "再挂起")
	sig(SignalReject)
	want(string(Failed), RunFailed, "驳回终态") // 终点锚
	// 全程不另起新 run：run 环条数不变（审批信号是迁移事件非对账触发）。
	if n := len(runs.State().Runs); n != 1 {
		t.Fatalf("审批信号不应另起新 run，环内应恒 1 条: %d", n)
	}
}

// TestApprovalSignalGuards 挂起-恢复零动作防御面（v0.24）：object 源无从迁移、非 Workflow
// involved 不命中、无闸锚序列无挂起面、非挂起态无可恢复/驳回、未知锚零作用、非审批信号族
// 零作用、重复提升幂等。
func TestApprovalSignalGuards(t *testing.T) {
	start := func(runs *RunLog, route, target string) {
		runs.StartRun(TriggerRecord{
			Seq: 1, EventID: "ev1", Type: "DocApplied", DedupKey: "apply:x:1",
			ReceivedAt: time.Now(), Route: route, Owner: "global",
		}, ExecutorBinding{Executor: ExecWorkflow, Target: target})
	}
	sig := func(runs *RunLog, typ, involved string) {
		runs.ResumeSignal(TriggerRecord{Type: typ, Involved: involved, ReceivedAt: time.Now()})
	}

	// object 源（解析面未注入=零投影）：相位不投影无从迁移——审批提升零动作（status 恒
	// triggered；v0.28 起声明了 phases 且解析面注入的对象 run 参与判定——TestObjectRunJudgmentWired）。
	objLog := NewRunLog(nil)
	start(objLog, "workflows/deploy-system-workflow", "deploy-system-workflow")
	sig(objLog, SignalSuspend, "Workflow/builtin/deploy-system-workflow")
	if r := objLog.State().Runs[0]; r.Status != RunTriggered || r.Phase != "" {
		t.Fatalf("object 源 run 审批提升应零动作: %+v", r)
	}

	// 非 Workflow involved（既有 ApprovalNeeded 权限拒绝产源 involved=Role/Group）：不命中 run 域。
	runningLog := NewRunLog(nil)
	start(runningLog, "workflows/dev-delivery", "dev-delivery")
	sig(runningLog, SignalSuspend, "Role/global/reviewer")
	if r := runningLog.State().Runs[0]; r.Status != RunRunning {
		t.Fatalf("非 Workflow involved 审批提升应零动作: %+v", r)
	}

	// 无闸锚序列（watch-only）：声明无闸锚=无挂起面（零动作）。
	noGateLog := NewRunLog(nil)
	start(noGateLog, "workflows/watch-only", "watch-only")
	sig(noGateLog, SignalSuspend, "Workflow/builtin/watch-only")
	if r := noGateLog.State().Runs[0]; r.Status != RunRunning || r.Phase != string(Analyzing) {
		t.Fatalf("无闸锚序列审批提升应零动作: %+v", r)
	}

	// 非挂起态：批准/驳回零动作；未知锚/非审批信号族零作用；重复提升幂等。
	gateLog := NewRunLog(nil)
	start(gateLog, "workflows/dev-delivery", "dev-delivery")
	sig(gateLog, SignalResume, "Workflow/builtin/dev-delivery")
	sig(gateLog, SignalReject, "Workflow/builtin/dev-delivery")
	if r := gateLog.State().Runs[0]; r.Status != RunRunning || r.Phase != string(Analyzing) {
		t.Fatalf("非挂起态批准/驳回应零动作: %+v", r)
	}
	sig(gateLog, SignalSuspend, "Workflow/builtin/dev-delivery")
	sig(gateLog, SignalSuspend, "Workflow/builtin/dev-delivery") // 幂等：已挂起再提升=零动作
	if r := gateLog.State().Runs[0]; r.Status != RunSuspended || r.Phase != string(AwaitingApproval) {
		t.Fatalf("重复审批提升应幂等零动作: %+v", r)
	}
	sig(gateLog, "DocApplied", "Workflow/builtin/dev-delivery") // 非审批信号族零作用
	sig(gateLog, SignalResume, "Workflow/builtin/nowhere")      // 未知锚零作用
	if r := gateLog.State().Runs[0]; r.Status != RunSuspended || r.Phase != string(AwaitingApproval) {
		t.Fatalf("非审批信号族/未知锚应零作用: %+v", r)
	}
}

// TestApprovalResumeFromGateSequence 恢复续行的序列同源锚（v0.24）：core-workflow 闸锚
// 序位 3，批准续行落 Executing（闸锚的声明序列后继）；序列尾防御：闸锚居尾（构造序列）
// 批准零动作。
func TestApprovalResumeFromGateSequence(t *testing.T) {
	runs := NewRunLog(nil)
	runs.StartRun(TriggerRecord{Seq: 1, Type: "DocApplied", ReceivedAt: time.Now(),
		Route: "owner-workflow"}, ExecutorBinding{Executor: ExecWorkflow, Target: "owner"})
	sig := func(typ string) {
		runs.ResumeSignal(TriggerRecord{Type: typ, Involved: "Workflow/builtin/core-workflow", ReceivedAt: time.Now()})
	}
	sig(SignalSuspend)
	if r := runs.State().Runs[0]; r.Phase != string(AwaitingApproval) || r.Status != RunSuspended {
		t.Fatalf("core-workflow 审批提升应落闸锚: %+v", runs.State().Runs[0])
	}
	sig(SignalResume)
	if r := runs.State().Runs[0]; r.Phase != string(Executing) || r.Status != RunRunning {
		t.Fatalf("core-workflow 批准应沿序列续行到 Executing: %+v", runs.State().Runs[0])
	}
}

// L61-02 触发路径 enabled honoring（v0.58b——E3 余量段收口，workflow.md §1 语义零新词形）：
// enabled=false 锚触发留痕仍在、run 不落（disabled 计数可见）；enabled=true/未声明/判定缺席
// （ok=false）=放行既有行为面；precedence=系统层 override 优先（字典序首个）→ global 本体 →
// 零对象=缺省启用；在途 run 迁移判定不受开关作用（ResumeSignal 零改动）。
func TestStartRunEnabledGate(t *testing.T) {
	newRuns := func(res EnabledResolver) *RunLog {
		return NewRunLog(&RunOptions{EnabledRules: res})
	}
	rec := func() TriggerRecord {
		return TriggerRecord{Seq: 1, Type: "DocApplied", ReceivedAt: time.Now(), Route: "owner-workflow"}
	}
	bind := ExecutorBinding{Executor: ExecWorkflow, Target: "owner"}
	obj := func(scope string, enabled string) EnabledObject {
		c := "apiVersion: core/v1\nkind: Workflow\nmetadata:\n  name: core-workflow\n  scope: " + scope + "\nspec:\n"
		if enabled != "" {
			c += "  enabled: " + enabled + "\n"
		}
		return EnabledObject{Scope: scope, Content: []byte(c)}
	}
	// ① enabled=false（系统 override）→ 拦：run 零落 + disabled=1；判定缺席面（nil resolver）
	//    对照 = 全落 + disabled=0（既有行为面零翻转）。
	gated := newRuns(EnabledFunc(func(name string) (bool, bool) { return false, true }))
	gated.StartRun(rec(), bind)
	if st := gated.State(); len(st.Runs) != 0 || st.Disabled != 1 || st.Started != 0 {
		t.Fatalf("enabled=false 应拦（run 零落∧disabled 计数可见）: %+v", st)
	}
	plain := newRuns(nil)
	plain.StartRun(rec(), bind)
	if st := plain.State(); len(st.Runs) != 1 || st.Disabled != 0 {
		t.Fatalf("判定面缺席应放行（既有行为面零翻转）: %+v", st)
	}
	// ② ok=false（判定缺席——未命中/零对象）→ 放行；enabled=true → 落；未声明 → 缺省启用落。
	miss := newRuns(EnabledFunc(func(name string) (bool, bool) { return false, false }))
	miss.StartRun(rec(), bind)
	if st := miss.State(); len(st.Runs) != 1 || st.Disabled != 0 {
		t.Fatalf("判定缺席应放行: %+v", st)
	}
	on := newRuns(EnabledFunc(func(name string) (bool, bool) { return true, true }))
	on.StartRun(rec(), bind)
	if st := on.State(); len(st.Runs) != 1 || st.Disabled != 0 {
		t.Fatalf("enabled=true 应照落: %+v", st)
	}
	undeclared := newRuns(EnabledFunc(func(name string) (bool, bool) { return parseEnabled([]byte("spec: {}\n")), true }))
	undeclared.StartRun(rec(), bind)
	if st := undeclared.State(); len(st.Runs) != 1 {
		t.Fatalf("enabled 未声明应缺省启用（scheme.yaml (=true)）: %+v", st)
	}
	// ③ precedence：系统层 override false 压 global 本体 true（拦）；global false 拦；
	//    两系统并存=字典序首个生效（a-sys false 拦，即便 b-sys true）。
	override := newRuns(EnabledFunc(func(name string) (bool, bool) {
		return EnabledOf([]EnabledObject{obj("global", "true"), obj("systems/b-sys", "true"), obj("systems/a-sys", "false")})
	}))
	override.StartRun(rec(), bind)
	if st := override.State(); len(st.Runs) != 0 || st.Disabled != 1 {
		t.Fatalf("系统 override enabled=false 应压 global 本体与后序系统（字典序首个）: %+v", st)
	}
	globOnly := newRuns(EnabledFunc(func(name string) (bool, bool) {
		return EnabledOf([]EnabledObject{obj("global", "false")})
	}))
	globOnly.StartRun(rec(), bind)
	if st := globOnly.State(); len(st.Runs) != 0 || st.Disabled != 1 {
		t.Fatalf("global 本体 enabled=false 应拦: %+v", st)
	}
	noObj := newRuns(EnabledFunc(func(name string) (bool, bool) { return EnabledOf(nil) }))
	noObj.StartRun(rec(), bind)
	if st := noObj.State(); len(st.Runs) != 1 || st.Disabled != 0 {
		t.Fatalf("零对象应缺省启用（builtin 注册表缺省——缺省不改既有行为面）: %+v", st)
	}
	// ④ 在途 run 迁移判定不受开关作用：落 run 后翻转 enabled=false，审批提升照迁移（闸锚挂起）。
	live := newRuns(EnabledFunc(func(name string) (bool, bool) { return true, true }))
	live.StartRun(rec(), bind)
	if len(live.State().Runs) != 1 {
		t.Fatalf("前置 run 应在环: %+v", live.State())
	}
	live.ResumeSignal(TriggerRecord{Type: SignalSuspend, Involved: "Workflow/builtin/core-workflow", ReceivedAt: time.Now()})
	if r := live.State().Runs[0]; r.Phase != string(AwaitingApproval) || r.Status != RunSuspended {
		t.Fatalf("开关生效后在途 run 迁移判定应零改动（已完成轮次保留）: %+v", r)
	}
}

// L61-02b EnabledOf 纯映射验收（precedence 三态+缺省词形+损坏文档防御）。
func TestEnabledOfPrecedence(t *testing.T) {
	obj := func(scope string, enabled string) EnabledObject {
		c := "apiVersion: core/v1\nkind: Workflow\nmetadata:\n  name: wf\n  scope: " + scope + "\nspec:\n"
		if enabled != "" {
			c += "  enabled: " + enabled + "\n"
		}
		return EnabledObject{Scope: scope, Content: []byte(c)}
	}
	if v, ok := EnabledOf([]EnabledObject{obj("systems/z-sys", "false"), obj("systems/a-sys", "true")}); !ok || v != true {
		t.Fatalf("多系统并存应字典序首个生效（a-sys=true——z-sys 不参与）: v=%v ok=%v", v, ok)
	}
	if v, ok := EnabledOf([]EnabledObject{obj("global", "")}); !ok || v != true {
		t.Fatalf("global 未声明应缺省启用且判定在册: v=%v ok=%v", v, ok)
	}
	if v, ok := EnabledOf([]EnabledObject{{Scope: "systems/s", Content: []byte(":::broken")}}); !ok || v != true {
		t.Fatalf("损坏文档应不拦阻（写路径校验=权威闸门）: v=%v ok=%v", v, ok)
	}
	if v, ok := EnabledOf(nil); ok || v != true {
		t.Fatalf("零对象=缺省启用且判定缺席: v=%v ok=%v", v, ok)
	}
}

// TestApprovalModeAuto —— 审批策略投影（2026-11-26「Workflow 可设置 Agent 节点是否自动级别」）：
// spec.approval.mode=auto 的对象源 run——(1) ApprovalMode 投影随行；(2) 受理首步落闸锚不挂起
// （auto=闸锚自动通过）；(3) ApprovalNeeded 提升对 auto run 零动作；(4) 词形白名单——非 auto
// 词形一律 human 缺省（挂起语义保持）。
func TestApprovalModeAuto(t *testing.T) {
	const seq = `apiVersion: core/v1
kind: Workflow
metadata: {name: auto-wf, scope: global}
spec:
  approval: {mode: auto}
  phases: [Idle, AwaitingApproval, Executing, Reporting]
  transitions:
    - {from: Idle, on: advance, to: AwaitingApproval}
    - {from: AwaitingApproval, on: advance, to: Executing}
    - {from: AwaitingApproval, on: ApprovalGranted, to: Executing}
    - {from: AwaitingApproval, on: ApprovalRejected, to: Reporting}
    - {from: Executing, on: advance, to: Reporting}
`
	const seqHuman = `apiVersion: core/v1
kind: Workflow
metadata: {name: human-wf, scope: global}
spec:
  approval: {mode: manual}
  phases: [Idle, AwaitingApproval, Executing, Reporting]
  transitions:
    - {from: Idle, on: advance, to: AwaitingApproval}
    - {from: AwaitingApproval, on: advance, to: Executing}
    - {from: AwaitingApproval, on: ApprovalGranted, to: Executing}
    - {from: AwaitingApproval, on: ApprovalRejected, to: Reporting}
    - {from: Executing, on: advance, to: Reporting}
`
	lookup := func(name string) ([]byte, bool) {
		switch name {
		case "auto-wf":
			return []byte(seq), true
		case "human-wf":
			return []byte(seqHuman), true
		}
		return nil, false
	}
	mk := func(docs DocsSpecs) *RunLog {
		return NewRunLog(&RunOptions{RingSize: 8, Specs: docs})
	}
	// auto：受理首步 Idle→AwaitingApproval——不挂起（running 直续）。
	auto := mk(DocsSpecs{Lookup: lookup})
	auto.StartRun(TriggerRecord{Seq: 1, EventID: "e1", Type: "DocApplied", ReceivedAt: time.Now(),
		Route: "workflows/auto-wf"}, ExecutorBinding{Executor: ExecWorkflow, Target: "auto-wf"})
	r := auto.State().Runs[0]
	if r.ApprovalMode != "auto" {
		t.Fatalf("ApprovalMode 应投影 auto: %+v", r)
	}
	if r.Phase != string(AwaitingApproval) || r.Status != RunRunning {
		t.Fatalf("auto run 落闸锚应 running 续行（不挂起）: phase=%s status=%s", r.Phase, r.Status)
	}
	// 审批提升对 auto run=零动作（闸锚自动通过——ApprovalNeeded 不挂起）。
	auto.ResumeSignal(TriggerRecord{Seq: 2, EventID: "e2", Type: "ApprovalNeeded", Involved: "Workflow/global/auto-wf", ReceivedAt: time.Now()})
	if got := auto.State().Runs[0]; got.Status != RunRunning {
		t.Fatalf("auto run 提升后应保持 running: %+v", got)
	}
	// 词形白名单：mode=manual（非 auto 词形）→ human 缺省——落闸锚挂起。
	human := mk(DocsSpecs{Lookup: lookup})
	human.StartRun(TriggerRecord{Seq: 3, EventID: "e3", Type: "DocApplied", ReceivedAt: time.Now(),
		Route: "workflows/human-wf"}, ExecutorBinding{Executor: ExecWorkflow, Target: "human-wf"})
	h := human.State().Runs[0]
	if h.ApprovalMode != "" {
		t.Fatalf("非 auto 词形应缺省空（human）: %+v", h.ApprovalMode)
	}
	if h.Phase != string(AwaitingApproval) || h.Status != RunSuspended {
		t.Fatalf("human run 落闸锚应挂起: phase=%s status=%s", h.Phase, h.Status)
	}
	// 挂起后批准恢复（表行审批边）仍可达——auto 不削减人工兜底通路。
	human.ResumeSignal(TriggerRecord{Seq: 4, EventID: "e4", Type: "ApprovalGranted", Involved: "Workflow/global/human-wf", ReceivedAt: time.Now()})
	if got := human.State().Runs[0]; got.Status != RunRunning || got.Phase != string(Executing) {
		t.Fatalf("human run 批准后续行 Executing: phase=%s status=%s", got.Phase, got.Status)
	}
}
