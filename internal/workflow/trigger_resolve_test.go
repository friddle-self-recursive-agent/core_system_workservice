// trigger_resolve_test.go —— 触发骨架决策面/执行投影/审批信号转发验收（R-ENG-7 行数收口
// 纯搬家：Resolve 决策快照/execFor 映射表/WireExecutor 接线翻转/workflow face 接线/审批
// 信号 ResumeSink 转发——零语义变化，trigger_test.go 拆分伴生件）。
package workflow

import (
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

func TestResolveDecisionFace(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	stop := h.AttachDispatch(b, []scheme.DispatchRule{
		{On: "DocApplied", For: "owner:*", Dispatch: "owner-workflow"},
		{On: "HeartbeatLost", For: "owner:agent", Dispatch: "parent-agent", CC: "notification/*"},
		{On: "*", For: "scope:builtin/*", Dispatch: "workflows/builtin-root"},
		{On: "*", For: "any", Dispatch: "dead-letter"},
	})
	defer stop()

	d := h.Resolve("HeartbeatLost", "ServiceInstance/systems/agent/ci")
	if !d.Matched || d.Route != "parent-agent" || d.CC != "notification/*" ||
		d.For != "owner:agent" || d.Owner != "agent" || d.Row != 1 {
		t.Fatalf("owner 命中决策不符: %+v", d)
	}
	d = h.Resolve("MetricThreshold", "ServiceInstance/systems/dev/ci") // 无具体行 → 通配兜底行
	if !d.Matched || d.Route != "dead-letter" || d.Row != 3 || d.Owner != "dev" {
		t.Fatalf("兜底行决策不符: %+v", d)
	}
	d = h.Resolve("UnknownType", "Doc/builtin/core") // 未枚举类型 → scope:builtin/* 通配行
	if !d.Matched || d.Route != "workflows/builtin-root" || d.Row != 2 || d.Owner != "builtin" {
		t.Fatalf("builtin 通配决策不符: %+v", d)
	}
	d = h.Resolve("DocApplied", "Agent/global/probe")
	if !d.Matched || d.Route != "owner-workflow" || d.Row != 0 || d.Owner != "global" {
		t.Fatalf("owner:* 决策不符: %+v", d)
	}
	// 决策面零副作用：多次 resolve 后 received/records 不动。
	if st := h.State(); st.Received != 0 || len(st.Records) != 0 {
		t.Fatalf("决策面不得产生留痕：received=%d records=%d", st.Received, len(st.Records))
	}
	// 全表未命中（畸形谓词行只行）→ 合成死信、不可归因表行。
	h2 := NewTriggerHub(nil)
	defer h2.AttachDispatch(b, []scheme.DispatchRule{{On: "X", For: "bad-pred", Dispatch: "nowhere"}})()
	if d := h2.Resolve("X", "Agent/global/a"); d.Matched || d.Route != DeadLetterRoute || d.Row != NoRow {
		t.Fatalf("全表未命中应合成死信 row=NoRow: %+v", d)
	}
	// 未挂表 → 无决策（route 空；owner 解析照常给出）。
	h3 := NewTriggerHub(nil)
	if d := h3.Resolve("DocApplied", "Agent/global/a"); d.Matched || d.Route != "" || d.Row != NoRow || d.Owner != "global" {
		t.Fatalf("未挂表应无决策: %+v", d)
	}
}

// ═══════════ 引擎主体第二段（v0.14：派发执行投影——route/cc 目标→执行面绑定可查询） ═══════════

// TestExecForMapping 执行面映射表单测（家族收编 + 未知目标 unbound 缺口可见 + 退化输入不收编）。
func TestExecForMapping(t *testing.T) {
	cases := []struct {
		target string
		exec   string
		tgt    string
	}{
		{"", "", ""},                                      // 空 cc=零值绑定
		{"dead-letter", ExecNone, ""},                     // 死信=none（观测兜底出口）
		{"notification/*", ExecNotification, "*"},         // 通配通道
		{"notification/email", ExecNotification, "email"}, // 具名通道
		{"workflows/builtin-root", ExecWorkflow, "builtin-root"},
		{"workflows/deploy-system-workflow", ExecWorkflow, "deploy-system-workflow"},
		{"parent-agent", ExecAgent, "parent"},
		{"root-agent", ExecAgent, "root"},
		{"owner-agent", ExecAgent, "owner"},                         // 新 agent 前缀同族自动收编（扩展性）
		{"owner-workflow", ExecWorkflow, "owner"},                   // owner 对账 workflow
		{"owner-instance-workflow", ExecWorkflow, "owner-instance"}, // scheme.md 同族词汇
		{"observation-bus", ExecUnbound, "observation-bus"},         // 未设计执行面的目标族=缺口可见
		{"instance-docs", ExecUnbound, "instance-docs"},
		{"mystery-face", ExecUnbound, "mystery-face"},   // 未知目标原样投影（不静默吞并）
		{"notification/", ExecUnbound, "notification/"}, // 空通道名不收编
		{"workflows/", ExecUnbound, "workflows/"},       // 空 workflow 名不收编
		{"-agent", ExecUnbound, "-agent"},               // 空前缀不收编
		{"-workflow", ExecUnbound, "-workflow"},
	}
	for _, c := range cases {
		b := execFor(c.target)
		if b.Executor != c.exec || b.Target != c.tgt {
			t.Errorf("execFor(%q)={%s,%s} want {%s,%s}", c.target, b.Executor, b.Target, c.exec, c.tgt)
		}
		if b.Wired {
			t.Errorf("execFor 纯映射不得自带接线态（wired 由 hub 覆写）: %q", c.target)
		}
	}
}

// TestWireExecutorFlip 接线态投影：缺省恒 false（诚实投影——执行计划可查 ≠ 执行面已在位）；
// WireExecutor 登记后仅该执行面翻 true（机械判别器：真实拉起落地轮的翻转锚点）；留痕同投影。
func TestWireExecutorFlip(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	defer h.AttachDispatch(b, []scheme.DispatchRule{
		{On: "DocApplied", For: "owner:*", Dispatch: "owner-workflow"},
		{On: "Escalate", For: "any", Dispatch: "root-agent", CC: "notification/*"},
	})()

	d := h.Resolve("Escalate", "Agent/global/a")
	if d.Exec.Executor != ExecAgent || d.Exec.Target != "root" || d.Exec.Wired ||
		d.CCExec.Executor != ExecNotification || d.CCExec.Target != "*" || d.CCExec.Wired {
		t.Fatalf("未登记执行面应投影 wired=false: %+v", d)
	}
	h.WireExecutor(ExecAgent)
	h.WireExecutor(ExecAgent, ExecAgent) // 幂等重复登记
	d = h.Resolve("Escalate", "Agent/global/a")
	if !d.Exec.Wired || d.CCExec.Wired {
		t.Fatalf("登记 agent 后应仅 agent 面翻 wired=true: %+v", d)
	}
	// 留痕同投影：收到事件携当时的执行计划快照。
	pub(t, b, "Escalate", "Agent/global/a")
	st := waitReceived(t, h, 1)
	r := st.Records[0]
	if r.Route != "root-agent" || r.Exec.Executor != ExecAgent || r.Exec.Target != "root" || !r.Exec.Wired ||
		r.CCExec.Executor != ExecNotification || r.CCExec.Target != "*" {
		t.Fatalf("留痕应携执行绑定投影: %+v", r)
	}
	// 合成死信（全表未命中）= none 绑定；登记 none 前 wired=false。
	h2 := NewTriggerHub(nil)
	defer h2.AttachDispatch(b, []scheme.DispatchRule{{On: "X", For: "bad-pred", Dispatch: "nowhere"}})()
	if d := h2.Resolve("X", "Agent/global/a"); d.Exec.Executor != ExecNone || d.CCExec.Executor != "" {
		t.Fatalf("合成死信应 none 绑定+零值 cc 绑定: %+v", d)
	}
	h2.WireExecutor(ExecNone)
	if d := h2.Resolve("X", "Agent/global/a"); !d.Exec.Wired {
		t.Fatalf("登记 none 后死信投影应 wired=true（观测兜底面在位口径）: %+v", d)
	}
}

// TestResolveDecisionExecProjection 决策面执行投影（权威表同构 fixture）：
// owner-workflow→workflow/owner、notification/*→通道、agent 前缀→agent、
// dead-letter 兜底→none、未挂表=零值绑定（无决策即无执行计划）。
func TestResolveDecisionExecProjection(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	defer h.AttachDispatch(b, []scheme.DispatchRule{
		{On: "DocApplied", For: "owner:*", Dispatch: "owner-workflow"},
		{On: "ApprovalNeeded", For: "any", Dispatch: "notification/*"},
		{On: "HeartbeatLost", For: "owner:agent", Dispatch: "parent-agent", CC: "notification/*"},
		{On: "*", For: "any", Dispatch: "dead-letter"},
	})()

	d := h.Resolve("HeartbeatLost", "ServiceInstance/systems/agent/ci")
	if d.Exec.Executor != ExecAgent || d.Exec.Target != "parent" || d.CCExec.Executor != ExecNotification || d.CCExec.Target != "*" {
		t.Fatalf("parent-agent+cc 投影不符: %+v", d)
	}
	d = h.Resolve("DocApplied", "Agent/global/x")
	if d.Exec.Executor != ExecWorkflow || d.Exec.Target != "owner" || d.CCExec.Executor != "" {
		t.Fatalf("owner-workflow 投影不符（无 cc=零值绑定）: %+v", d)
	}
	d = h.Resolve("ApprovalNeeded", "Role/global/r")
	if d.Exec.Executor != ExecNotification || d.Exec.Target != "*" {
		t.Fatalf("notification 通道投影不符: %+v", d)
	}
	d = h.Resolve("UnknownType", "Agent/global/a")
	if d.Route != "dead-letter" || d.Exec.Executor != ExecNone {
		t.Fatalf("显式 dead-letter 行应 none 绑定: %+v", d)
	}
	h3 := NewTriggerHub(nil)
	defer h3.AttachDispatch(b, nil)()
	if d := h3.Resolve("DocApplied", "Agent/global/a"); d.Exec.Executor != "" || d.CCExec.Executor != "" {
		t.Fatalf("未挂表应无执行计划（零值绑定）: %+v", d)
	}
}

// ═══════════ 引擎主体第三段（v0.15：真实拉起第一子段——workflow 触发面接线+对账 run 投影） ═══════════

// TestWorkflowFaceWiring workflow 触发面接线（真实拉起第一子段）：
// 装配层双登记（BindRuns + WireExecutor(ExecWorkflow)）后，route 目标执行面=workflow 的
// 留痕推进为一次对账 run 投影（对账锚→core-workflow；直接引用→builtin/平台对象可见）；
// 其余执行面（notification/agent/none）不产生 run；单登记（只 BindRuns 未 WireExecutor）
// 不推进——接线判定=两登记同时在场（机械判别器：wired 翻转 ∧ run 面出现）。
func TestWorkflowFaceWiring(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	runs := NewRunLog(nil)
	h.BindRuns(runs)
	h.WireExecutor(ExecWorkflow)
	defer h.AttachDispatch(b, []scheme.DispatchRule{
		{On: "DocApplied", For: "owner:*", Dispatch: "owner-workflow"},
		{On: "InstallRequested", For: "any", Dispatch: "workflows/watch-only"},
		{On: "TemplateFound", For: "any", Dispatch: "workflows/deploy-system-workflow"}, // 非builtin=平台对象引用
		{On: "Escalate", For: "any", Dispatch: "root-agent"},
		{On: "ApprovalNeeded", For: "any", Dispatch: "notification/*"},
		{On: "*", For: "any", Dispatch: "dead-letter"},
	})()

	// 对账锚：owner-workflow → run {core-workflow, builtin, Idle 起点, triggered}。
	pub(t, b, "DocApplied", "Agent/global/probe")
	st := waitState(t, h, 1)
	rs := runs.State()
	if len(rs.Runs) != 1 {
		t.Fatalf("对账锚事件应产生 1 条 run: %+v", rs)
	}
	r := rs.Runs[0]
	rec := st.Records[0]
	if r.Workflow != ReconcileWorkflow || r.SpecSource != SpecBuiltin || r.Status != RunRunning ||
		r.Phase != string(Analyzing) || r.TriggerSeq != rec.Seq || r.Route != "owner-workflow" ||
		r.Owner != rec.Owner || r.Target != "owner" || r.EventID != rec.EventID {
		t.Fatalf("对账锚 run 投影不符（v0.23 迁移判定首段：受理即启动 running/Analyzing）: %+v vs 留痕 %+v", r, rec)
	}
	if len(r.Phases) != 7 || r.Phases[0] != string(Idle) || r.Phases[6] != string(Reporting) {
		t.Fatalf("run 应携正向相位全序列（F-04 尾=Reporting）: %v", r.Phases)
	}

	// 直接引用：builtin 命中=按名；未命中=平台对象引用（锚保留、相位空——缺口/对象面可见）。
	pub(t, b, "InstallRequested", "Agent/global/probe")
	pub(t, b, "TemplateFound", "Agent/global/probe")
	waitState(t, h, 3)
	rs = runs.State()
	if len(rs.Runs) != 3 {
		t.Fatalf("两条直接引用应各产生一条 run（共 3）: %+v", rs)
	}
	if rs.Runs[1].Workflow != "watch-only" || rs.Runs[1].SpecSource != SpecBuiltin || rs.Runs[1].Phase != string(Analyzing) {
		t.Fatalf("builtin 直接引用 run 不符（v0.23 受理即启动=第二相位）: %+v", rs.Runs[1])
	}
	if rs.Runs[2].Workflow != "deploy-system-workflow" || rs.Runs[2].SpecSource != SpecObject ||
		len(rs.Runs[2].Phases) != 0 || rs.Runs[2].Phase != "" {
		t.Fatalf("非 builtin 直接引用应投影平台对象引用（锚保留相位空）: %+v", rs.Runs[2])
	}

	// 其余执行面不产生 run：agent 送达面、notification 通道、死信兜底。
	n := len(runs.State().Runs)
	pub(t, b, "Escalate", "Agent/global/probe")
	pub(t, b, "ApprovalNeeded", "Role/global/r")
	pub(t, b, "MysteryType", "Agent/global/probe")  // 全表未命中→合成死信
	pub(t, b, "UnknownType2", "Agent/global/probe") // 显式 dead-letter 行
	waitState(t, h, 7)
	if got := len(runs.State().Runs); got != n {
		t.Fatalf("非 workflow 执行面不得产生 run：before=%d after=%d", n, got)
	}

	// 单登记不推进（只 BindRuns 未 WireExecutor）：接线判定=两登记同时在场。
	h2 := NewTriggerHub(nil)
	h2.BindRuns(NewRunLog(nil))
	defer h2.AttachDispatch(b, []scheme.DispatchRule{{On: "DocApplied", For: "owner:*", Dispatch: "owner-workflow"}})()
	pub(t, b, "DocApplied", "Agent/global/probe")
	waitState(t, h2, 1)
	if rs := h2.runsSink().(*RunLog).State(); len(rs.Runs) != 0 {
		t.Fatalf("未登记 wired 不应推进 run（诚实投影）: %+v", rs)
	}

	// 只 WireExecutor 未 BindRuns：不推进也不 panic（sink 缺席=投影面未挂）。
	h3 := NewTriggerHub(nil)
	h3.WireExecutor(ExecWorkflow)
	defer h3.AttachDispatch(b, []scheme.DispatchRule{{On: "DocApplied", For: "owner:*", Dispatch: "owner-workflow"}})()
	pub(t, b, "DocApplied", "Agent/global/probe")
	waitState(t, h3, 1)
}

// runsSink 测试取面（hub.runs 私有；测试经同包访问）。
func (h *TriggerHub) runsSink() RunSink { return h.runs }

// agentsSink 测试取面（hub.wakes 私有；测试经同包访问）。
func (h *TriggerHub) agentsSink() WakeSink { return h.wakes }

// noResumeSink 只实现 RunSink 的假 sink（能力断言反证面：未声明 ResumeSink 能力=审批
// 信号不转发，能力增量不破既有 sink 契约）。
type noResumeSink struct{ started int }

func (s *noResumeSink) StartRun(rec TriggerRecord, b ExecutorBinding) { s.started++ }

// TestHubApprovalSignalForwarding 挂起-恢复首段 hub 转发面（v0.24）：审批信号族事件经
// 同一订阅主干到达 → 按可选能力接口转发 runs sink → run 域迁移判定生效；**不另起 run**
// （审批信号是相位机迁移事件，dispatch 行=dead-letter 标注派发面无目标）；无能力 sink 不转发。
func TestHubApprovalSignalForwarding(t *testing.T) {
	b := eventbus.NewBus(nil)
	runs := NewRunLog(nil)
	h := NewTriggerHub(nil)
	h.BindRuns(runs)
	h.WireExecutor(ExecWorkflow)
	defer h.AttachDispatch(b, []scheme.DispatchRule{
		{On: "DocApplied", For: "owner:*", Dispatch: "workflows/dev-delivery"},
		{On: "ApprovalNeeded", For: "any", Dispatch: "notification/*"}, // 既有行：主路送达不变
		{On: "ApprovalGranted", For: "any", Dispatch: "dead-letter"},   // 审批信号族：dispatch 面无目标
		{On: "ApprovalRejected", For: "any", Dispatch: "dead-letter"},
	})()

	applyEvent(t, b, "Agent/global/probe", 1)
	waitState(t, h, 1)
	r0 := runs.State().Runs[0]
	if r0.Workflow != "dev-delivery" || r0.Status != RunRunning || r0.Phase != string(Analyzing) {
		t.Fatalf("对账 run 受理基线（v0.23 口径）应 {dev-delivery,running,Analyzing}: %+v", r0)
	}

	// 审批提升：ApprovalNeeded involved=Workflow 引用 → 挂起（不另起 run；dispatch 面=notification 主路不变）。
	if err := b.Publish(&eventbus.Event{Type: "ApprovalNeeded", Source: "permission",
		Involved: "Workflow/builtin/dev-delivery", DedupKey: "permission:Workflow/builtin/dev-delivery", Level: "warn"}); err != nil {
		t.Fatalf("发布 ApprovalNeeded: %v", err)
	}
	waitState(t, h, 2)
	r0 = runs.State().Runs[0]
	if r0.Phase != string(AwaitingApproval) || r0.Status != RunSuspended {
		t.Fatalf("审批提升应落闸锚挂起: %+v", r0)
	}
	if n := len(runs.State().Runs); n != 1 {
		t.Fatalf("审批信号不应另起 run: %d", n)
	}
	// 既有行为面不变：ApprovalNeeded（involved=Role/Group 产源形态）留痕照落、通知主路照走。
	last := h.State().Records[len(h.State().Records)-1]
	if last.Route != "notification/*" {
		t.Fatalf("ApprovalNeeded 留痕 route 应沿既有 dispatch 行: %+v", last)
	}

	// 批准恢复：闸锚沿声明序列续行。
	if err := b.Publish(&eventbus.Event{Type: "ApprovalGranted", Source: "human",
		Involved: "Workflow/builtin/dev-delivery", DedupKey: "approval:Workflow/builtin/dev-delivery", Level: "info"}); err != nil {
		t.Fatalf("发布 ApprovalGranted: %v", err)
	}
	waitState(t, h, 3)
	r0 = runs.State().Runs[0]
	if r0.Phase != string(Planning) || r0.Status != RunRunning {
		t.Fatalf("批准应沿声明序列续行到 Planning: %+v", r0)
	}
	last = h.State().Records[len(h.State().Records)-1]
	if last.Route != DeadLetterRoute {
		t.Fatalf("ApprovalGranted dispatch 面应无目标（dead-letter 标注）: %+v", last)
	}

	// 无能力 sink：类型断言失败不转发（契约不破）。
	fake := &noResumeSink{}
	h2 := NewTriggerHub(nil)
	h2.BindRuns(fake)
	h2.WireExecutor(ExecWorkflow)
	defer h2.AttachDispatch(b, []scheme.DispatchRule{
		{On: "DocApplied", For: "owner:*", Dispatch: "owner-workflow"},
		{On: "ApprovalGranted", For: "any", Dispatch: "dead-letter"}, // 审批信号到达 h2：验证无能力 sink 不转发
	})()
	applyEvent(t, b, "Agent/global/probe2", 2)
	waitState(t, h2, 1)
	if fake.started != 1 {
		t.Fatalf("无能力 sink 应只收 StartRun: %d", fake.started)
	}
	if err := b.Publish(&eventbus.Event{Type: "ApprovalGranted", Source: "human",
		Involved: "Workflow/builtin/dev-delivery", DedupKey: "approval: Workflow/builtin/dev-delivery", Level: "info"}); err != nil {
		t.Fatalf("发布 ApprovalGranted: %v", err)
	}
	waitState(t, h2, 2)
	if fake.started != 1 {
		t.Fatalf("无能力 sink 审批信号不应转发/不应起 run: %d", fake.started)
	}
}
