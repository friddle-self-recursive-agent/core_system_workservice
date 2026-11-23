// eventbus_test.go —— 总线最小实现验收：窗口聚合、订阅分发、死信可见、debug 不入库、退订竞态。
package eventbus

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
)

type memOps struct {
	applyRows []opslog.ApplyLog
	eventRows []opslog.EventLog
}

func (m *memOps) AppendApply(l opslog.ApplyLog) error {
	m.applyRows = append(m.applyRows, l)
	return nil
}
func (m *memOps) AppendEvent(l opslog.EventLog) error {
	m.eventRows = append(m.eventRows, l)
	return nil
}
func (m *memOps) Reconcile() (int, error) { return 0, nil }

func TestDedupWindowAggregates(t *testing.T) {
	ops := &memOps{}
	b := NewBus(&Options{Ops: ops, DedupWindow: 200 * time.Millisecond})
	ch, cancel := b.Subscribe("DocApplied")
	defer cancel()

	e := Event{Type: "DocApplied", Source: "apply", Involved: "Agent/global/a", DedupKey: "apply:Agent/global/a:1"}
	if err := b.Publish(&e); err != nil {
		t.Fatal(err)
	}
	if err := b.Publish(&e); err != nil { // 窗口内同因 → 聚合
		t.Fatal(err)
	}
	got := 0
	deadline := time.After(time.Second)
drain:
	for {
		select {
		case <-ch:
			got++
		case <-deadline:
			break drain
		}
	}
	if got != 1 {
		t.Fatalf("窗口内同 DedupKey 应聚合为一次投递，得 %d", got)
	}
	if len(ops.eventRows) != 1 {
		t.Fatalf("聚合事件只归档一次，event_log 行 = %d", len(ops.eventRows))
	}
	// 窗口过后同键再次发布 → 再次投递。
	time.Sleep(250 * time.Millisecond)
	if err := b.Publish(&e); err != nil {
		t.Fatal(err)
	}
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("窗口过期后同键事件应再次投递")
	}
}

func TestSubscribeFilterAndDeadLetter(t *testing.T) {
	b := NewBus(nil)
	ch, cancel := b.Subscribe("MetricThreshold")
	defer cancel()

	// 未订阅类型 → 死信（必须可见，event.md rule 4）。
	if err := b.Publish(&Event{Type: "DocApplied", DedupKey: "k1"}); err != nil {
		t.Fatal(err)
	}
	if len(b.DeadLetters()) != 1 {
		t.Fatalf("路由不到应转死信，得 %d", len(b.DeadLetters()))
	}
	// 订阅类型 → 投递，不死信。
	if err := b.Publish(&Event{Type: "MetricThreshold", DedupKey: "k2"}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-ch:
	case <-time.After(time.Second):
		t.Fatal("订阅类型应送达")
	}
	if len(b.DeadLetters()) != 1 {
		t.Fatalf("死信不应增长，得 %d", len(b.DeadLetters()))
	}
	// 退订后死信。
	cancel()
	if err := b.Publish(&Event{Type: "MetricThreshold", DedupKey: "k3"}); err != nil {
		t.Fatal(err)
	}
	if len(b.DeadLetters()) != 2 {
		t.Fatalf("退订后应转死信，得 %d", len(b.DeadLetters()))
	}
}

func TestDebugNotArchivedAndValidation(t *testing.T) {
	ops := &memOps{}
	b := NewBus(&Options{Ops: ops})
	ch, cancel := b.Subscribe("")
	defer cancel()

	if err := b.Publish(&Event{Type: "ExternalChanged", Level: "debug", DedupKey: "d1"}); err != nil {
		t.Fatal(err)
	}
	select {
	case <-ch: // debug 仍进内存队列
	case <-time.After(time.Second):
		t.Fatal("debug 应照常投递内存订阅方")
	}
	if len(ops.eventRows) != 0 {
		t.Fatalf("debug 默认不入库（rule 1b），event_log 行 = %d", len(ops.eventRows))
	}
	if err := b.Publish(&Event{DedupKey: "d2"}); err == nil {
		t.Fatal("缺 type 应被拒绝（ErrRejected）")
	}
}

// TestSubscribeCancelPublishRace 退订×发布并发互斥：订阅原语的健全性底线——
// 退订关频道与投递并发不得 send-on-closed（历史实现路由快照在锁外投递，存在此竞态窗口；
// 现投递与退订同锁互斥）。触发骨架（workflow.TriggerHub）Attach/Stop 依赖此不变量。
func TestSubscribeCancelPublishRace(t *testing.T) {
	b := NewBus(nil)
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { // 发布者：持续发布独立 dedup 键事件（永不聚合，全量走路由）
		defer wg.Done()
		for i := 0; ; i++ {
			select {
			case <-stop:
				return
			default:
			}
			if err := b.Publish(&Event{Type: "DocApplied", DedupKey: fmt.Sprintf("race-%d", i)}); err != nil {
				t.Errorf("发布被拒: %v", err)
				return
			}
		}
	}()
	for i := 0; i < 200; i++ { // 订阅/退订 200 轮与发布者并发
		_, cancel := b.Subscribe("DocApplied")
		cancel()
		cancel() // 退订幂等
	}
	close(stop)
	wg.Wait()
	if dead := len(b.DeadLetters()); dead == 0 {
		t.Fatal("并发窗口内事件应大部分转死信（无稳定订阅方），死信计数不应为空")
	}
}
