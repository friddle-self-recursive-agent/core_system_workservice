// wake_test.go —— agent 唤醒投影面验收（引擎主体第五段 v0.18——真实拉起第三子段）：
// 权威表两 agent 行机械验证（Escalate→root-agent 主路、HeartbeatLost→parent-agent
// owner 谓词主路）+ 双登记接线判定（单登记不推进）+ 双路收编（route/cc，origin 留痕）
// + 环容量/丢最旧/dropped 可见 + 受理投影锚定（status=accepted，无后续态机——诚实口径）。
package workflow

import (
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// TestWakeProjection 直投落点：锚投影=root/parent（<p>-agent 前缀段）、origin 留痕、
// 状态=受理词 accepted、溯源字段与留痕同源（wake 可回指 /v1/triggers records）。
func TestWakeProjection(t *testing.T) {
	wakes := NewWakeLog(nil)
	wakes.Wake(TriggerRecord{
		Seq: 3, EventID: "ev1", Type: "Escalate", DedupKey: "core-workflow:retries",
		Level: "error", ReceivedAt: time.Now(), Route: "root-agent", Owner: "global",
	}, ExecutorBinding{Executor: ExecAgent, Target: "root"}, OriginRoute)
	st := wakes.State()
	if st.Accepted != 1 || len(st.Wakes) != 1 {
		t.Fatalf("应 1 条唤醒投影: %+v", st)
	}
	w := st.Wakes[0]
	if w.Status != WakeAccepted {
		t.Fatalf("唤醒状态应为受理词 accepted（无后续态机）: %+v", w)
	}
	if w.Agent != "root" || w.Origin != OriginRoute || w.Level != "error" ||
		w.TriggerSeq != 3 || w.EventID != "ev1" || w.DedupKey != "core-workflow:retries" ||
		w.Route != "root-agent" || w.Owner != "global" {
		t.Fatalf("唤醒投影溯源字段不符: %+v", w)
	}
	wakes.Wake(TriggerRecord{Seq: 4, Level: "error", ReceivedAt: time.Now(), Route: "parent-agent"},
		ExecutorBinding{Executor: ExecAgent, Target: "parent"}, OriginCC)
	st = wakes.State()
	if st.Accepted != 2 || st.Wakes[1].Agent != "parent" || st.Wakes[1].Origin != OriginCC {
		t.Fatalf("cc 旁路唤醒应 origin=cc agent=parent: %+v", st.Wakes[1])
	}
	if st.Wakes[0].Wake >= st.Wakes[1].Wake {
		t.Fatalf("唤醒序号应单调: %+v", st.Wakes)
	}
}

// TestWakeRing 环容量：溢出丢最旧、dropped 计数可见、accepted=dropped+环内恒等式
// （与 run 环/送达环同口径）；唤醒序号跨挤出仍单调。
func TestWakeRing(t *testing.T) {
	wakes := NewWakeLog(&WakeOptions{RingSize: 2})
	for i := 0; i < 4; i++ {
		wakes.Wake(TriggerRecord{Seq: int64(i + 1), ReceivedAt: time.Now()},
			ExecutorBinding{Executor: ExecAgent, Target: "root"}, OriginRoute)
	}
	st := wakes.State()
	if st.Accepted != 4 || st.Dropped != 2 || len(st.Wakes) != 2 {
		t.Fatalf("环口径应 accepted=4 dropped=2 len=2: %+v", st)
	}
	if st.Accepted != st.Dropped+int64(len(st.Wakes)) {
		t.Fatalf("accepted=dropped+环内恒等式不成立: %+v", st)
	}
	if st.Wakes[0].Wake != 3 || st.Wakes[1].Wake != 4 {
		t.Fatalf("丢最旧后应剩唤醒 3、4: %+v", st.Wakes)
	}
}

// TestAgentFaceWiring agent 触发面接线（真实拉起第三子段）——权威表两行机械验证：
// ① Escalate（for any）→ route=root-agent：唤醒 1 条 {agent=root, origin=route}，
//
//	cc=notification/* 旁路归送达面（不串面）；② HeartbeatLost（for owner:agent）→
//	route=parent-agent：owner 谓词按解析命中（systems/agent→agent），唤醒
//	{agent=parent}；非 agent 事件（workflow/notification 执行面）零唤醒；
//	决策投影 wired 翻转与唤醒面出现同证（两面合证同口径）。
func TestAgentFaceWiring(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	runs := NewRunLog(nil)
	deliver := NewDeliveryLog(nil)
	wakes := NewWakeLog(nil)
	h.BindRuns(runs)
	h.BindDeliveries(deliver)
	h.BindAgents(wakes)
	h.WireExecutor(ExecWorkflow, ExecNotification, ExecAgent)
	defer h.AttachDispatch(b, []scheme.DispatchRule{
		{On: "DocApplied", For: "owner:*", Dispatch: "owner-workflow"},
		{On: "ApprovalNeeded", For: "any", Dispatch: "notification/*"},
		{On: "Escalate", For: "any", Dispatch: "root-agent", CC: "notification/*"},                // 权威表 Escalate 行
		{On: "HeartbeatLost", For: "owner:agent", Dispatch: "parent-agent", CC: "notification/*"}, // 权威表 HeartbeatLost 行
		{On: "*", For: "any", Dispatch: "dead-letter"},                                            // 死信兜底可见（未知类型可达=none 面不串唤醒）
	})()

	// 决策面 wired 翻转（判别器之一：接线前恒 false，登记后 true）。
	d := h.Resolve("Escalate", "Agent/global/probe")
	if d.Exec.Executor != ExecAgent || d.Exec.Target != "root" || !d.Exec.Wired {
		t.Fatalf("Escalate 决策投影应 agent/root/wired=true: %+v", d)
	}
	d = h.Resolve("HeartbeatLost", "ServiceInstance/systems/agent/ci")
	if d.Exec.Executor != ExecAgent || d.Exec.Target != "parent" || !d.Exec.Wired {
		t.Fatalf("HeartbeatLost 决策投影应 agent/parent/wired=true: %+v", d)
	}

	// ① Escalate：唤醒 1 条（root/route）；旁路通知归送达面；owner-workflow 主路……
	// 注意权威表 Escalate 行 dispatch=root-agent（无 owner-workflow 路），不产生 run。
	pub(t, b, "Escalate", "Agent/global/probe")
	waitState(t, h, 1)
	ws := wakes.State()
	if ws.Accepted != 1 || len(ws.Wakes) != 1 {
		t.Fatalf("Escalate 应恰 1 条唤醒: %+v", ws)
	}
	if ws.Wakes[0].Agent != "root" || ws.Wakes[0].Origin != OriginRoute ||
		ws.Wakes[0].Type != "Escalate" || ws.Wakes[0].Level != "info" {
		t.Fatalf("Escalate 唤醒投影不符: %+v", ws.Wakes[0])
	}
	if rs := runs.State(); rs.Started != 0 {
		t.Fatalf("agent 主路不产生对账 run（root-agent≠workflow 族）: %+v", rs)
	}
	if dl := deliver.State(); dl.Accepted != 1 || dl.Deliveries[0].Origin != OriginCC {
		t.Fatalf("Escalate cc=notification/* 应旁路送达恰 1 条: %+v", dl)
	}

	// ② HeartbeatLost：owner:agent 谓词按解析命中（systems/agent→agent）→ parent 唤醒。
	e := eventbus.Event{
		Type: "HeartbeatLost", Source: "core", Involved: "ServiceInstance/systems/agent/ci",
		DedupKey: "hb:ci", Level: "error",
	}
	if err := b.Publish(&e); err != nil {
		t.Fatalf("发布 HeartbeatLost: %v", err)
	}
	waitState(t, h, 2)
	ws = wakes.State()
	if ws.Accepted != 2 || ws.Wakes[1].Agent != "parent" || ws.Wakes[1].Origin != OriginRoute ||
		ws.Wakes[1].Level != "error" {
		t.Fatalf("HeartbeatLost 应唤醒 parent（origin=route level=error）: %+v", ws.Wakes[1])
	}

	// 非 agent 事件零唤醒（workflow/notification/none 面不串）。
	n := len(wakes.State().Wakes)
	pub(t, b, "DocApplied", "Agent/global/probe")
	pub(t, b, "ApprovalNeeded", "Role/global/r")
	pub(t, b, "MysteryType", "Agent/global/probe") // 全表未命中→合成死信 none
	waitState(t, h, 5)
	if got := len(wakes.State().Wakes); got != n {
		t.Fatalf("非 agent 执行面不得产生唤醒：before=%d after=%d", n, got)
	}
}

// TestAgentFaceUnwired 单登记不推进（接线判定判据）：BindAgents 挂接而执行面未登记时
// 零唤醒——与 run/送达面同口径的机械判别器（单登记不生效，双登记才接线）；
// 只 WireExecutor 未 BindAgents：不推进也不 panic（sink 缺席=投影面未挂）。
func TestAgentFaceUnwired(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	wakes := NewWakeLog(nil)
	h.BindAgents(wakes) // 只挂 sink，不 WireExecutor(ExecAgent)
	defer h.AttachDispatch(b, []scheme.DispatchRule{
		{On: "Escalate", For: "any", Dispatch: "root-agent"},
	})()

	pub(t, b, "Escalate", "Agent/global/probe")
	waitState(t, h, 1)
	if st := wakes.State(); st.Accepted != 0 {
		t.Fatalf("执行面未登记应零唤醒投影: %+v", st)
	}
	if d := h.Resolve("Escalate", "Agent/global/probe"); d.Exec.Wired {
		t.Fatalf("未登记时决策投影 wired 应 false（诚实投影）: %+v", d)
	}

	// 登记即生效（登记后到达的绑定照投影；幂等登记不重复推进）。
	h.WireExecutor(ExecAgent)
	h.WireExecutor(ExecAgent)
	pub(t, b, "Escalate", "Agent/global/probe")
	waitState(t, h, 2)
	if st := wakes.State(); st.Accepted != 1 {
		t.Fatalf("登记后唤醒应恰 1 条（幂等登记不重复）: %+v", st)
	}

	// 只 WireExecutor 未 BindAgents：零推进不 panic。
	h3 := NewTriggerHub(nil)
	h3.WireExecutor(ExecAgent)
	defer h3.AttachDispatch(b, []scheme.DispatchRule{
		{On: "Escalate", For: "any", Dispatch: "root-agent"},
	})()
	pub(t, b, "Escalate", "Agent/global/probe")
	waitState(t, h3, 1)
}

// TestWakeSinkContract sink 接口契约：WakeLog 满足 WakeSink；hub 挂接面经同包取面断言
// （BindAgents nil=解除挂接幂等）。
func TestWakeSinkContract(t *testing.T) {
	h := NewTriggerHub(nil)
	wakes := NewWakeLog(nil)
	h.BindAgents(wakes)
	if h.agentsSink() != WakeSink(wakes) {
		t.Fatalf("BindAgents 应挂接同 sink")
	}
	h.BindAgents(nil)
	if h.agentsSink() != nil {
		t.Fatalf("BindAgents(nil) 应解除挂接")
	}
}
