// deliver_test.go —— notification 送达投影面验收（引擎主体第四段 v0.16——真实拉起第二子段）：
// 受理投影口径（status=accepted、通道投影、origin 主/旁留痕）+ 接线判定（单登记不推进——
// WireExecutor∧BindDeliveries 双登记才生效）+ route/cc 双路收编 + 环容量/丢最旧/dropped 可见。
package workflow

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// TestDeliverProjection 受理投影口径：未注入出站时状态恒受理词 accepted（v0.17 注入后
// 才异步翻转 sent/failed——送达受理 ≠ 已送达真实通道）；通道投影=绑定锚点；origin 主/旁
// 留痕；溯源字段同源。
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

// ───────────────────────── v0.17 真实通道出站挂接（受理→异步出站→状态翻转） ─────────────────────────

// stubOutbound 测试注入实现（Outbound 契约）：记录收到的投影，可编程失败/阻塞/panic。
type stubOutbound struct {
	mu    sync.Mutex
	got   []DeliveryRecord
	fail  error
	hold  chan struct{} // 非 nil：Send 内阻塞直至放行（队列满/停机语义测试用）
	paniq bool          // true：Send 内 panic（兜底路径测试用）
	delay time.Duration // 每次 Send 前延迟（停机在途测试用）
}

func (s *stubOutbound) Send(_ context.Context, rec DeliveryRecord) error {
	s.mu.Lock()
	s.got = append(s.got, rec)
	s.mu.Unlock()
	if s.delay > 0 {
		time.Sleep(s.delay)
	}
	if s.paniq {
		panic("stub 出站爆炸")
	}
	if s.hold != nil {
		<-s.hold
	}
	return s.fail
}

func (s *stubOutbound) count() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.got)
}

// waitDeliverState 轮询送达投影直至谓词满足（异步出站完成可观察点；超时 2s）。
func waitDeliverState(t *testing.T, d *DeliveryLog, pred func(DeliveryState) bool) DeliveryState {
	t.Helper()
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		st := d.State()
		if pred(st) {
			return st
		}
		time.Sleep(5 * time.Millisecond)
	}
	t.Fatalf("送达投影未在时限内到达期望态: %+v", d.State())
	return DeliveryState{}
}

// TestDeliverOutboundSent 出站成功路径：受理即异步出站，status 就地翻转 accepted→sent，
// Attempts/CompletedAt 留痕，sent 计数进位；注入实现收到的投影=受理投影全字段（信封源头）。
func TestDeliverOutboundSent(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	stub := &stubOutbound{}
	deliver := NewDeliveryLog(&DeliveryOptions{Outbound: stub})
	defer deliver.Close()
	h.BindDeliveries(deliver)
	h.WireExecutor(ExecNotification)
	defer h.AttachDispatch(b, []scheme.DispatchRule{
		{On: "DocApplied", For: "any", Dispatch: "notification/mail-chan"},
	})()

	e := applyEvent(t, b, "Agent/global/probe", 1)
	waitState(t, h, 1)
	st := waitDeliverState(t, deliver, func(s DeliveryState) bool { return s.Sent == 1 })
	if st.Failed != 0 || st.Accepted != 1 {
		t.Fatalf("成功路径计数应 sent=1 failed=0 accepted=1: %+v", st)
	}
	rec := st.Deliveries[0]
	if rec.Status != DeliverSent || rec.Attempts != 1 || rec.CompletedAt.IsZero() || rec.Error != "" {
		t.Fatalf("成功投影应 {sent, attempts=1, completedAt≠零, error=空}: %+v", rec)
	}
	if stub.count() != 1 {
		t.Fatalf("注入实现应恰被调用 1 次: %d", stub.count())
	}
	sent := stub.got[0]
	if sent.Delivery != rec.Delivery || sent.Channel != "mail-chan" || sent.Origin != OriginRoute ||
		sent.EventID != e.ID || sent.Type != e.Type || sent.DedupKey != e.DedupKey || sent.Level != e.Level {
		t.Fatalf("注入实现收到的投影应与受理投影同源: %+v vs %+v", sent, rec)
	}
}

// TestDeliverOutboundFailed 出站失败路径：注入实现返回错误 → status=failed，错误文本进
// 投影（可查），failed 计数进位——诚实投影不吞失败。
func TestDeliverOutboundFailed(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	stub := &stubOutbound{fail: errors.New("smtp 中继不可达")}
	deliver := NewDeliveryLog(&DeliveryOptions{Outbound: stub})
	defer deliver.Close()
	h.BindDeliveries(deliver)
	h.WireExecutor(ExecNotification)
	defer h.AttachDispatch(b, []scheme.DispatchRule{
		{On: "DocApplied", For: "any", Dispatch: "notification/mail-chan"},
	})()

	applyEvent(t, b, "Agent/global/probe", 1)
	waitState(t, h, 1)
	st := waitDeliverState(t, deliver, func(s DeliveryState) bool { return s.Failed == 1 })
	if st.Sent != 0 {
		t.Fatalf("失败路径 sent 应为 0: %+v", st)
	}
	rec := st.Deliveries[0]
	if rec.Status != DeliverFailed || rec.Attempts != 1 || rec.CompletedAt.IsZero() ||
		rec.Error != "smtp 中继不可达" {
		t.Fatalf("失败投影应 {failed, attempts=1, 错误文本留存}: %+v", rec)
	}
}

// TestDeliverOutboundPanic 注入实现 panic 兜底：判 failed 不拖垮 sink（后续事件照常受理出站）。
func TestDeliverOutboundPanic(t *testing.T) {
	stub := &stubOutbound{paniq: true}
	deliver := NewDeliveryLog(&DeliveryOptions{Outbound: stub})
	defer deliver.Close()
	deliver.Deliver(TriggerRecord{
		Seq: 1, EventID: "ev1", Type: "Escalate", DedupKey: "k1", Level: "error", ReceivedAt: time.Now(),
	}, ExecutorBinding{Executor: ExecNotification, Target: Wildcard}, OriginCC)
	st := waitDeliverState(t, deliver, func(s DeliveryState) bool { return s.Failed == 1 })
	if st.Deliveries[0].Status != DeliverFailed ||
		!strings.Contains(st.Deliveries[0].Error, "panic") {
		t.Fatalf("panic 应兜底为 failed 且错误携 panic 痕迹: %+v", st.Deliveries[0])
	}
}

// TestDeliverOutboundQueueFull 队列满判 failed 不阻塞受理路径（非阻塞 sink 契约的
// 机械验证）：阻塞注入实现占住出站协程 → 队列容量 1 → 第 3 条受理入队立即翻 failed
// （attempts=0=未尝试），错误=出站队列满。
func TestDeliverOutboundQueueFull(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	hold := make(chan struct{})
	stub := &stubOutbound{hold: hold}
	deliver := NewDeliveryLog(&DeliveryOptions{Outbound: stub, OutboundQueue: 1})
	defer deliver.Close()
	defer close(hold)
	h.BindDeliveries(deliver)
	h.WireExecutor(ExecNotification)
	defer h.AttachDispatch(b, []scheme.DispatchRule{
		{On: "DocApplied", For: "any", Dispatch: "notification/mail-chan"},
	})()

	applyEvent(t, b, "Agent/global/probe", 1) // ① 出站协程取走并在 Send 内阻塞
	waitDeliverState(t, deliver, func(s DeliveryState) bool { return stub.count() == 1 })
	applyEvent(t, b, "Agent/global/probe", 2) // ② 入队（队列占满）
	applyEvent(t, b, "Agent/global/probe", 3) // ③ 队列满 → 立即 failed
	waitState(t, h, 3)
	st := waitDeliverState(t, deliver, func(s DeliveryState) bool { return s.Failed == 1 })
	if st.Accepted != 3 {
		t.Fatalf("受理路径不应被出站阻塞（accepted=3）: %+v", st)
	}
	full := st.Deliveries[2]
	if full.Status != DeliverFailed || full.Attempts != 0 || !strings.Contains(full.Error, "出站队列满") {
		t.Fatalf("第 3 条应 {failed, attempts=0, 错误=出站队列满}: %+v", full)
	}
}

// TestDeliverCloseStopSemantics 停机语义：在途出站等完成（Close 阻塞直至 sent 进位），
// 未起跑任务留 accepted（观测面非账本，重启不重放）；Close 幂等，未注入出站=no-op 不挂死。
func TestDeliverCloseStopSemantics(t *testing.T) {
	stub := &stubOutbound{delay: 30 * time.Millisecond}
	deliver := NewDeliveryLog(&DeliveryOptions{Outbound: stub, OutboundWorkers: 1})
	for i := 1; i <= 3; i++ {
		deliver.Deliver(TriggerRecord{
			Seq: int64(i), EventID: fmt.Sprintf("ev%d", i), Type: "DocApplied",
			DedupKey: fmt.Sprintf("k%d", i), Level: "info", ReceivedAt: time.Now(),
		}, ExecutorBinding{Executor: ExecNotification, Target: "mail-chan"}, OriginRoute)
	}
	waitDeliverState(t, deliver, func(s DeliveryState) bool { return s.Sent >= 2 })
	deliver.Close() // 在途第 3 条等完成（受超时约束）
	st := deliver.State()
	if st.Failed != 0 || st.Sent+st.Failed != 3 {
		t.Fatalf("停机应等在途完成、零 failed（sent+failed=3）: %+v", st)
	}
	if st.Sent != int64(stub.count()) {
		t.Fatalf("sent 计数应与注入实现实际调用一致: %+v vs %d", st, stub.count())
	}
	deliver.Close() // 幂等（不重入、不挂死）

	noop := NewDeliveryLog(nil)
	noop.Close() // 未注入出站=no-op
}

// TestDeliverOutboundEvicted 环挤出后出站完成仍进计数（诚实投影：环是观测窗非账本——
// 已挤出投影无法就地翻转，但 sent 计数不丢）。
func TestDeliverOutboundEvicted(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	stub := &stubOutbound{}
	deliver := NewDeliveryLog(&DeliveryOptions{Outbound: stub, RingSize: 1})
	defer deliver.Close()
	h.BindDeliveries(deliver)
	h.WireExecutor(ExecNotification)
	defer h.AttachDispatch(b, []scheme.DispatchRule{
		{On: "DocApplied", For: "any", Dispatch: "notification/mail-chan"},
	})()

	for i := 1; i <= 3; i++ {
		applyEvent(t, b, "Agent/global/probe", int64(i))
	}
	waitState(t, h, 3)
	st := waitDeliverState(t, deliver, func(s DeliveryState) bool { return s.Sent == 3 })
	if st.Accepted != 3 || len(st.Deliveries) != 1 || st.Dropped != 2 {
		t.Fatalf("环口径 accepted=3 dropped=2 len=1: %+v", st)
	}
	if st.Deliveries[0].Status != DeliverSent || st.Deliveries[0].Delivery != 3 {
		t.Fatalf("环内唯一投影应为最新的 #3 且已翻转 sent: %+v", st.Deliveries[0])
	}
	if int64(stub.count()) != 3 {
		t.Fatalf("挤出投影的出站不应丢（实现实际调用=3）: %d", stub.count())
	}
}
