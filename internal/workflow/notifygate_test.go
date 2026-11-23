// notifygate_test.go —— 通知阶段门+dedup/minLevel 验收（v0.35 KI-001 通知段）：
// 门链三闸（minLevel 过滤→阶段门→dedup 窗口）判定口径+计数可见+KI-001 §6 判别器
// （未完成相位 apply 通知计数=0；终态=1；同 dedupKey 窗口 N 次注入实发 1）。
// 窗口断言纪律：窗口 ≥1s 判据确定性优先（loop-33 测试纪律——1.2s 余量防抖）。
package workflow

import (
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// gateRec 构造测试用触发留痕。
func gateRec(typ, dedupKey, level, involved string) TriggerRecord {
	return TriggerRecord{
		Seq: 1, EventID: "ev-" + dedupKey, Type: typ, DedupKey: dedupKey,
		Level: level, Involved: involved, ReceivedAt: time.Now(),
	}
}

// gateBinding notification 族绑定。
func gateBinding() ExecutorBinding {
	return ExecutorBinding{Executor: ExecNotification, Target: "mail-chan"}
}

// TestNotifyGateNilOptions 零值可用：无门链选项=既有行为面零翻转（任意事件照常受理——
// 既有 v0.16/v0.17 投影测试全走本口径，此处显式锚定）。
func TestNotifyGateNilOptions(t *testing.T) {
	d := NewDeliveryLog(nil)
	d.Deliver(gateRec("DocApplied", "apply:x:1", "info", "ServiceInstance/systems/s/i"), gateBinding(), OriginRoute)
	st := d.State()
	if st.Accepted != 1 || st.Gated != 0 || st.Filtered != 0 || st.Deduped != 0 {
		t.Fatalf("零值选项应零门链 accepted=1 gated=0 filtered=0 deduped=0: %+v", st)
	}
}

// TestNotifyGatePhaseGate KI-001 §6 判别器（run 轴）：未完成相位的 apply 通知计数=0；
// 终态（序列尾=Report 相位完成/终点锚 failed）=1。门只作用于通知面——gated 计数可见。
func TestNotifyGatePhaseGate(t *testing.T) {
	runs := NewRunLog(nil)
	// owner=s 的现势 run：序列尾 Report（相位终态——推进拍落地形态）。
	tail := RunRecord{Workflow: "dev-delivery", Owner: "s", Status: RunRunning,
		Phases: []string{"Analyze", "Gate", "Report"}, Phase: "Report"}
	// 中途 run：running ∧ phase≠尾 → 非终态。
	mid := RunRecord{Workflow: "dev-delivery", Owner: "s", Status: RunRunning,
		Phases: []string{"Analyze", "Gate", "Report"}, Phase: "Gate"}
	// 挂起 run：闸锚等人（人注意事实由 ApprovalNeeded 事件轴承载）→ 非终态。
	susp := RunRecord{Workflow: "dev-delivery", Owner: "s", Status: RunSuspended,
		Phases: []string{"Analyze", "AwaitingApproval", "Report"}, Phase: "AwaitingApproval"}
	// 终点锚 run：failed=迁移终点语义 → 终态。
	fail := RunRecord{Workflow: "dev-delivery", Owner: "s", Status: RunFailed,
		Phases: []string{"Analyze", "Failed"}, Phase: "Failed"}

	cases := []struct {
		name string
		rec  RunRecord
		want bool
	}{
		{"序列尾终态放行", tail, true},
		{"中途拦", mid, false},
		{"挂起拦", susp, false},
		{"终点锚放行", fail, true},
	}
	for _, tc := range cases {
		d := NewDeliveryLog(&DeliveryOptions{PhaseGate: NotifyPhaseGate(runs)})
		runs.mu.Lock()
		runs.ring = append(runs.ring, tc.rec)
		runs.mu.Unlock()
		d.Deliver(gateRec("DocApplied", "apply:x:1", "info", "ServiceInstance/systems/s/i"), gateBinding(), OriginRoute)
		st := d.State()
		if tc.want {
			if st.Accepted != 1 || st.Gated != 0 {
				t.Fatalf("%s: 应放行 accepted=1 gated=0: %+v", tc.name, st)
			}
		} else {
			if st.Accepted != 0 || st.Gated != 1 || len(st.Deliveries) != 0 {
				t.Fatalf("%s: 应拦 accepted=0 gated=1 零投影: %+v", tc.name, st)
			}
		}
	}
}

// TestNotifyGateTerminalTypes 事件轴：人注意终态族（ApprovalNeeded/Escalate/HeartbeatLost
// ——builtin dispatch 表通知族三型）恒放行，不看 run 面；无 run 面（runs=nil）下其余事件
// 一律拦（fail-closed——无证据不通知）。
func TestNotifyGateTerminalTypes(t *testing.T) {
	gate := NotifyPhaseGate(nil) // 无 run 证据面
	for _, typ := range []string{"ApprovalNeeded", "Escalate", "HeartbeatLost"} {
		if !NotifyTerminalType(typ) {
			t.Fatalf("%s 应属人注意终态族", typ)
		}
		d := NewDeliveryLog(&DeliveryOptions{PhaseGate: gate})
		d.Deliver(gateRec(typ, "k:"+typ, "warn", "Role/global/r"), gateBinding(), OriginRoute)
		if st := d.State(); st.Accepted != 1 {
			t.Fatalf("终态族 %s 无 run 面也应放行: %+v", typ, st)
		}
	}
	// 进度事实：无 run 面 → 拦（apply ≠ 阶段完成）。
	d := NewDeliveryLog(&DeliveryOptions{PhaseGate: gate})
	d.Deliver(gateRec("DocApplied", "apply:x:1", "info", "ServiceInstance/systems/s/i"), gateBinding(), OriginRoute)
	if st := d.State(); st.Accepted != 0 || st.Gated != 1 {
		t.Fatalf("无 run 面进度事实应拦 accepted=0 gated=1: %+v", st)
	}
}

// TestNotifyGateOwnerResolution owner 维度：门按 involved scope 根解析 owner 与 run 落点
// Owner 字段对账（systems/<s> → <s>；跨 owner run 不背书）。
func TestNotifyGateOwnerResolution(t *testing.T) {
	runs := NewRunLog(nil)
	runs.mu.Lock()
	runs.ring = append(runs.ring, RunRecord{Workflow: "dev-delivery", Owner: "s", Status: RunRunning,
		Phases: []string{"Analyze", "Report"}, Phase: "Report"})
	runs.mu.Unlock()
	d := NewDeliveryLog(&DeliveryOptions{PhaseGate: NotifyPhaseGate(runs)})
	// owner=s 终态 run 背书 owner=s 的 apply。
	d.Deliver(gateRec("DocApplied", "apply:s:1", "info", "ServiceInstance/systems/s/i"), gateBinding(), OriginRoute)
	// 其他 owner（global）无现势终态 run → 拦。
	d.Deliver(gateRec("DocApplied", "apply:g:1", "info", "Role/global/r"), gateBinding(), OriginRoute)
	st := d.State()
	if st.Accepted != 1 || st.Gated != 1 {
		t.Fatalf("owner 维度背书应 accepted=1 gated=1: %+v", st)
	}
}

// TestNotifyMinLevel minLevel 等级下限（v0.35 落地）：低于下限零受理零投影 filtered 计数
// 可见；达标照常受理；空下限=不过滤；未知等级词形不冒充达标。
func TestNotifyMinLevel(t *testing.T) {
	d := NewDeliveryLog(&DeliveryOptions{MinLevel: "warn"})
	d.Deliver(gateRec("Escalate", "k:debug", "debug", "Workflow/global/w"), gateBinding(), OriginRoute)
	d.Deliver(gateRec("Escalate", "k:info", "info", "Workflow/global/w"), gateBinding(), OriginRoute)
	d.Deliver(gateRec("ApprovalNeeded", "k:warn", "warn", "Role/global/r"), gateBinding(), OriginRoute)
	d.Deliver(gateRec("Escalate", "k:error", "error", "Workflow/global/w"), gateBinding(), OriginRoute)
	d.Deliver(gateRec("Escalate", "k:word", "loud", "Workflow/global/w"), gateBinding(), OriginRoute)
	st := d.State()
	if st.Filtered != 3 || st.Accepted != 2 {
		t.Fatalf("minLevel=warn 应 filtered=3（debug/info/词外）accepted=2（warn/error）: %+v", st)
	}
	if len(st.Deliveries) != 2 || st.Deliveries[0].Level != "warn" || st.Deliveries[1].Level != "error" {
		t.Fatalf("达标投影应只含 warn/error: %+v", st.Deliveries)
	}
	// 空下限=不过滤（零值口径）。
	d2 := NewDeliveryLog(&DeliveryOptions{})
	d2.Deliver(gateRec("Escalate", "k:debug", "debug", "Workflow/global/w"), gateBinding(), OriginRoute)
	if st2 := d2.State(); st2.Accepted != 1 || st2.Filtered != 0 {
		t.Fatalf("空 minLevel 应零过滤: %+v", st2)
	}
}

// TestNotifyDedupWindow dedup 窗口（v0.35 落地）：窗口内同 dedupKey 至多一封（其余 deduped
// 计数可见不投影）；窗口过期后新事件照常受理；不同 dedupKey 互不聚合；空 dedupKey 永不
// 聚合；DedupWindow<=0=不去重。窗口 ≥1s 判据确定性优先（1.2s 余量防抖）。
func TestNotifyDedupWindow(t *testing.T) {
	d := NewDeliveryLog(&DeliveryOptions{DedupWindow: 1200 * time.Millisecond})
	// 同键三次：第 1 封受理，第 2/3 次 deduped（KI-001 §6 判别器 N=3 → 实发 1、deduped=2）。
	d.Deliver(gateRec("ApprovalNeeded", "approval:r:1", "warn", "Role/global/r"), gateBinding(), OriginRoute)
	d.Deliver(gateRec("ApprovalNeeded", "approval:r:1", "warn", "Role/global/r"), gateBinding(), OriginRoute)
	d.Deliver(gateRec("ApprovalNeeded", "approval:r:1", "warn", "Role/global/r"), gateBinding(), OriginRoute)
	// 异键不聚合；空键永不聚合。
	d.Deliver(gateRec("Escalate", "esc:wf:a", "error", "Workflow/global/w"), gateBinding(), OriginCC)
	d.Deliver(gateRec("Escalate", "", "error", "Workflow/global/w"), gateBinding(), OriginCC)
	d.Deliver(gateRec("Escalate", "", "error", "Workflow/global/w"), gateBinding(), OriginCC)
	st := d.State()
	if st.Accepted != 4 || st.Deduped != 2 || len(st.Deliveries) != 4 {
		t.Fatalf("窗口内应 accepted=4（首封+异键+空键×2）deduped=2: %+v", st)
	}
	if st.Deliveries[0].DedupKey != "approval:r:1" || st.Deliveries[1].DedupKey != "esc:wf:a" {
		t.Fatalf("受理投影 dedupKey 溯源不符: %+v", st.Deliveries)
	}
	// 窗口过期（1.2s + 1.2s 余量）：同键新事件照常受理——窗口语义非永久抑制。
	time.Sleep(2400 * time.Millisecond)
	d.Deliver(gateRec("ApprovalNeeded", "approval:r:1", "warn", "Role/global/r"), gateBinding(), OriginRoute)
	st2 := d.State()
	if st2.Accepted != 5 || st2.Deduped != 2 {
		t.Fatalf("窗口过期后同键应照常受理 accepted=5 deduped=2: %+v", st2)
	}
	// DedupWindow<=0=不去重（零值口径）。
	d3 := NewDeliveryLog(&DeliveryOptions{})
	d3.Deliver(gateRec("ApprovalNeeded", "approval:r:2", "warn", "Role/global/r"), gateBinding(), OriginRoute)
	d3.Deliver(gateRec("ApprovalNeeded", "approval:r:2", "warn", "Role/global/r"), gateBinding(), OriginRoute)
	if st3 := d3.State(); st3.Accepted != 2 || st3.Deduped != 0 {
		t.Fatalf("DedupWindow<=0 应不去重: %+v", st3)
	}
}

// TestNotifyGateOrder 门链顺序：等级过滤先于阶段门（声明面过滤语义从订阅面到聚合面）——
// 未达标事件计 filtered 不计 gated；同键同窗跨闸聚合以放行受理时刻为锚。
func TestNotifyGateOrder(t *testing.T) {
	d := NewDeliveryLog(&DeliveryOptions{PhaseGate: NotifyPhaseGate(nil), MinLevel: "warn"})
	// info 级进度事实：等级先拦（filtered）——不进阶段门判定（gated 不计）。
	d.Deliver(gateRec("DocApplied", "apply:x:1", "info", "ServiceInstance/systems/s/i"), gateBinding(), OriginRoute)
	st := d.State()
	if st.Filtered != 1 || st.Gated != 0 || st.Accepted != 0 {
		t.Fatalf("门链顺序应 filtered=1 gated=0: %+v", st)
	}
}

// TestNotifyGateHubEndToEnd hub 端到端：dispatch 表 route=notification 行 + 双登记接线 +
// 门链注入——apply 事件零送达零投影 gated=1（KI-001 根因 4 的触发面级反证：通知由路由命中
// 而起，门在送达面收口；事件本体照常留痕/run——主干零旁路）。
func TestNotifyGateHubEndToEnd(t *testing.T) {
	bus := eventbus.NewBus(nil)
	d := NewDeliveryLog(&DeliveryOptions{PhaseGate: NotifyPhaseGate(nil)})
	hub := NewTriggerHub(nil)
	hub.BindDeliveries(d)
	hub.WireExecutor(ExecNotification)
	rules := []scheme.DispatchRule{
		{On: "DocApplied", For: "any", Dispatch: "notification/mail-chan"},
	}
	stop := hub.AttachDispatch(bus, rules)
	defer stop()
	// 注入 apply 事件（同管线零旁路）——路由命中 notification 行但阶段门拦。
	bus.Publish(&eventbus.Event{Type: "DocApplied", DedupKey: "apply:x:9", Level: "info",
		Involved: "ServiceInstance/systems/s/i"})
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		if hub.State().Received > 0 {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	st := d.State()
	if hub.State().Received != 1 {
		t.Fatalf("事件应照常留痕 received=1: %+v", hub.State())
	}
	if st.Accepted != 0 || st.Gated != 1 || len(st.Deliveries) != 0 {
		t.Fatalf("apply 事件应门拦 accepted=0 gated=1 零投影: %+v", st)
	}
}
