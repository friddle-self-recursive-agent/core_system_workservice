// deliver_test.go —— notification 送达投影面验收（引擎主体第四段 v0.16——真实拉起第二子段）：
// 受理投影口径（status=accepted、通道投影、origin 主/旁留痕）+ 接线判定（单登记不推进——
// WireExecutor∧BindDeliveries 双登记才生效）+ route/cc 双路收编 + 环容量/丢最旧/dropped 可见。
package workflow

import (
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// TestDeliverProjection 受理投影口径：状态词不含任何「已投递」语义（真实通道投递 deferred
// ——送达投影出现 ≠ 已送达真实通道）；通道投影=绑定锚点；origin 主/旁留痕；溯源字段同源。
func TestDeliverProjection(t *testing.T) {
	d := NewDeliveryLog(nil)
	d.Deliver(TriggerRecord{
		Seq: 3, EventID: "ev1", Type: "ApprovalNeeded", DedupKey: "apply:x:1",
		Level: "warn", ReceivedAt: time.Now(), Route: "notification/mail-chan", Owner: "global",
	}, ExecutorBinding{Executor: ExecNotification, Target: "mail-chan"}, OriginRoute)
	d.Deliver(TriggerRecord{
		Seq: 4, EventID: "ev2", Type: "Escalate", DedupKey: "esc:wf:denied",
		Level: "error", ReceivedAt: time.Now(), Route: "root-agent", Owner: "global",
	}, ExecutorBinding{Executor: ExecNotification, Target: Wildcard}, OriginCC)
	st := d.State()
	if st.Accepted != 2 || st.Dropped != 0 || len(st.Deliveries) != 2 {
		t.Fatalf("环口径应 accepted=2 dropped=0 len=2: %+v", st)
	}
	route, cc := st.Deliveries[0], st.Deliveries[1]
	if route.Status != DeliverAccepted || cc.Status != DeliverAccepted {
		t.Fatalf("送达状态应恒受理词 accepted: %+v %+v", route, cc)
	}
	if route.Channel != "mail-chan" || route.Origin != OriginRoute {
		t.Fatalf("route 路送达应 {mail-chan, route}: %+v", route)
	}
	if cc.Channel != Wildcard || cc.Origin != OriginCC {
		t.Fatalf("cc 路送达应 {*, cc}: %+v", cc)
	}
	if route.Delivery != 1 || cc.Delivery != 2 {
		t.Fatalf("送达序号应进程内单调 1、2: %+v %+v", route, cc)
	}
	if route.TriggerSeq != 3 || route.EventID != "ev1" || route.Type != "ApprovalNeeded" ||
		route.Level != "warn" || route.Route != "notification/mail-chan" || route.Owner != "global" {
		t.Fatalf("route 路溯源字段不符: %+v", route)
	}
	if cc.TriggerSeq != 4 || cc.Type != "Escalate" || cc.Level != "error" || cc.Route != "root-agent" {
		t.Fatalf("cc 路溯源字段不符（旁路送达仍携命中行原文）: %+v", cc)
	}
}

// TestDeliveryWiring 双路收编 + 双登记判据：装配层同时 WireExecutor(ExecNotification)
// ∧ BindDeliveries 时——route 主路 notification 行 → 送达投影（origin=route，无 run）；
// cc 旁路 notification 行 → 送达投影（origin=cc）+ run 并存（workflow 面同事件双投影）。
func TestDeliveryWiring(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	runs := NewRunLog(nil)
	deliver := NewDeliveryLog(nil)
	h.BindRuns(runs)
	h.BindDeliveries(deliver)
	h.WireExecutor(ExecWorkflow, ExecNotification)
	defer h.AttachDispatch(b, []scheme.DispatchRule{
		{On: "DocApplied", For: "scope:global/*", Dispatch: "notification/mail-chan"},    // 主路通知行
		{On: "DocApplied", For: "any", Dispatch: "owner-workflow", CC: "notification/*"}, // 对账行+旁路通知
	})()

	applyEvent(t, b, "Agent/global/probe", 1)         // scope=global → 行1 主路通知
	applyEvent(t, b, "TaskScheme/systems/probe/x", 2) // scope=systems/probe → 行2 对账+旁路
	waitState(t, h, 2)

	dl := deliver.State()
	if dl.Accepted != 2 || len(dl.Deliveries) != 2 {
		t.Fatalf("两事件应推进 2 条送达投影: %+v", dl)
	}
	if dl.Deliveries[0].Origin != OriginRoute || dl.Deliveries[0].Channel != "mail-chan" ||
		dl.Deliveries[0].TriggerSeq != 1 {
		t.Fatalf("行1 应主路送达 {route, mail-chan, triggerSeq=1}: %+v", dl.Deliveries[0])
	}
	if dl.Deliveries[1].Origin != OriginCC || dl.Deliveries[1].Channel != Wildcard ||
		dl.Deliveries[1].TriggerSeq != 2 {
		t.Fatalf("行2 应旁路送达 {cc, *, triggerSeq=2}: %+v", dl.Deliveries[1])
	}
	// run 并存口径（v0.15 语义不被 v0.16 稀释）：行2 对账 run 1 条，行1 主路通知无 run。
	rs := runs.State()
	if rs.Started != 1 || len(rs.Runs) != 1 || rs.Runs[0].TriggerSeq != 2 {
		t.Fatalf("run 应恰行2 一条（v0.15 口径不变）: %+v", rs)
	}
}

// TestDeliveryUnwired 单登记不推进（接线判定判据）：BindDeliveries 挂接而执行面未登记时
// 零送达投影——与 run 面同口径的机械判别器（单登记不生效，双登记才接线）。
func TestDeliveryUnwired(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	deliver := NewDeliveryLog(nil)
	h.BindDeliveries(deliver) // 只挂 sink，不 WireExecutor(ExecNotification)
	defer h.AttachDispatch(b, []scheme.DispatchRule{
		{On: "DocApplied", For: "any", Dispatch: "notification/mail-chan"},
	})()

	applyEvent(t, b, "Agent/global/probe", 1)
	waitState(t, h, 1)
	if st := deliver.State(); st.Accepted != 0 {
		t.Fatalf("执行面未登记应零送达投影: %+v", st)
	}

	// 登记即生效（登记后到达的绑定照投影；幂等登记不重复推进）。
	h.WireExecutor(ExecNotification)
	h.WireExecutor(ExecNotification)
	applyEvent(t, b, "Agent/global/probe", 2)
	waitState(t, h, 2)
	if st := deliver.State(); st.Accepted != 1 {
		t.Fatalf("登记后送达应恰 1 条（幂等登记不重复）: %+v", st)
	}
}

// TestDeliveryRing 环容量：溢出丢最旧、dropped 计数可见、accepted=dropped+环内恒等式
// （与 run 环同口径）；送达序号跨挤出仍单调（started 语义）。
func TestDeliveryRing(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	deliver := NewDeliveryLog(&DeliveryOptions{RingSize: 2})
	h.BindDeliveries(deliver)
	h.WireExecutor(ExecNotification)
	defer h.AttachDispatch(b, []scheme.DispatchRule{
		{On: "DocApplied", For: "any", Dispatch: "notification/mail-chan"},
	})()

	for i := 1; i <= 4; i++ {
		applyEvent(t, b, "Agent/global/probe", int64(i))
	}
	waitState(t, h, 4)
	st := deliver.State()
	if st.Accepted != 4 || st.Dropped != 2 || len(st.Deliveries) != 2 {
		t.Fatalf("环口径应 accepted=4 dropped=2 len=2: %+v", st)
	}
	if st.Accepted != st.Dropped+int64(len(st.Deliveries)) {
		t.Fatalf("accepted=dropped+环内恒等式不成立: %+v", st)
	}
	if st.Deliveries[0].Delivery != 3 || st.Deliveries[1].Delivery != 4 {
		t.Fatalf("丢最旧后应剩送达 3、4（序号跨挤出单调）: %+v", st.Deliveries)
	}
	if st.Deliveries[1].TriggerSeq != 4 {
		t.Fatalf("送达应携来源留痕 seq=4: %+v", st.Deliveries[1])
	}
}

// TestDeliveryAgentRow 权威表 owner 行形态：HeartbeatLost 行 dispatch=parent-agent ∧
// cc=notification 时——旁路送达 1 条（origin=cc），agent 主路无 run 无送达
// （agent 唤醒面未接线不派发；notification 只消费自身族绑定）。
func TestDeliveryAgentRow(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	runs := NewRunLog(nil)
	deliver := NewDeliveryLog(nil)
	h.BindRuns(runs)
	h.BindDeliveries(deliver)
	h.WireExecutor(ExecWorkflow, ExecNotification)
	defer h.AttachDispatch(b, []scheme.DispatchRule{
		{On: "HeartbeatLost", For: "owner:agent", Dispatch: "parent-agent", CC: "notification/*"},
	})()

	e := eventbus.Event{
		Type: "HeartbeatLost", Source: "core", Involved: "ServiceInstance/systems/agent/ci",
		DedupKey: "hb:ci", Level: "error",
	}
	if err := b.Publish(&e); err != nil {
		t.Fatalf("发布 HeartbeatLost: %v", err)
	}
	waitState(t, h, 1)

	if st := runs.State(); st.Started != 0 {
		t.Fatalf("agent 主路不产生对账 run: %+v", st)
	}
	dl := deliver.State()
	if dl.Accepted != 1 || dl.Deliveries[0].Origin != OriginCC ||
		dl.Deliveries[0].Channel != Wildcard || dl.Deliveries[0].Type != "HeartbeatLost" ||
		dl.Deliveries[0].Level != "error" {
		t.Fatalf("owner 行应恰旁路送达 1 条: %+v", dl)
	}
}
