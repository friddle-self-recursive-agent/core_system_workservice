// trigger_test.go —— 触发骨架验收（dispatch 接线第一段）：
// 收到留痕可查、去重聚合透传、类型过滤、退订排干、环溢出可见、发布×查询并发安全。
package workflow

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
)

// applyEvent 模拟 apply.Engine.finish 的 DocApplied 发布（同源字段投影）。
func applyEvent(t *testing.T, b *eventbus.Bus, objKey string, rv int64) eventbus.Event {
	t.Helper()
	e := eventbus.Event{
		Type: "DocApplied", Source: "apply", Involved: objKey,
		DedupKey: fmt.Sprintf("apply:%s:%d", objKey, rv), Level: "info",
	}
	if err := b.Publish(&e); err != nil {
		t.Fatalf("发布 DocApplied: %v", err)
	}
	return e
}

func waitState(t *testing.T, h *TriggerHub, want int64) TriggerState {
	t.Helper()
	deadline := time.After(time.Second)
	for {
		st := h.State()
		if st.Received == want {
			return st
		}
		select {
		case <-deadline:
			t.Fatalf("收到数不达预期：want=%d got=%d", want, st.Received)
		case <-time.After(time.Millisecond):
		}
	}
}

func TestTriggerHubRecordsTrace(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil) // 零值容量 256
	stop := h.Attach(b, "DocApplied")
	defer stop()

	e := applyEvent(t, b, "Agent/global/probe", 1)
	st := waitState(t, h, 1)
	if len(st.Records) != 1 {
		t.Fatalf("留痕应 1 条，得 %d", len(st.Records))
	}
	r := st.Records[0]
	if r.Seq != 1 || r.EventID != e.ID || r.DedupKey != e.DedupKey ||
		r.Type != "DocApplied" || r.Involved != "Agent/global/probe" || r.Source != "apply" {
		t.Fatalf("留痕字段与发布回执不同源: %+v vs %+v", r, e)
	}
	if r.ReceivedAt.Before(r.At) {
		t.Fatalf("receivedAt(%v) 不应早于事件 at(%v)", r.ReceivedAt, r.At)
	}
	if len(st.Types) != 1 || st.Types[0] != "DocApplied" {
		t.Fatalf("订阅类型集合应=[DocApplied]，得 %v", st.Types)
	}
}

func TestTriggerHubDedupPassthrough(t *testing.T) {
	b := eventbus.NewBus(&eventbus.Options{DedupWindow: 200 * time.Millisecond})
	h := NewTriggerHub(nil)
	stop := h.Attach(b, "DocApplied")
	defer stop()

	e := applyEvent(t, b, "Agent/global/probe", 2)
	applyEvent(t, b, "Agent/global/probe", 2) // 窗口内同因 → 总线聚合，订阅方只收一次
	st := waitState(t, h, 1)
	if st.Records[0].DedupKey != e.DedupKey {
		t.Fatalf("聚合后留痕 dedupKey 不符: %+v", st.Records[0])
	}
	time.Sleep(250 * time.Millisecond)
	applyEvent(t, b, "Agent/global/probe", 2) // 窗口过期 → 再次投递
	st = waitState(t, h, 2)
	if st.Records[1].Seq != 2 {
		t.Fatalf("窗口过期后应新增 seq=2 留痕: %+v", st.Records[1])
	}
}

func TestTriggerHubTypeFilter(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	stop := h.Attach(b, "DocApplied")
	defer stop()

	if err := b.Publish(&eventbus.Event{Type: "ExternalChanged", DedupKey: "x1"}); err != nil {
		t.Fatal(err)
	}
	st := waitState(t, h, 0)
	if len(st.Records) != 0 {
		t.Fatalf("未订阅类型不应产生留痕: %+v", st.Records)
	}
	applyEvent(t, b, "Agent/global/probe", 3)
	waitState(t, h, 1)
}

func TestTriggerHubUnsubscribeDrains(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	stop := h.Attach(b, "DocApplied")

	applyEvent(t, b, "Agent/global/pre", 1)
	waitState(t, h, 1)

	stop()
	stop() // 总退订幂等
	applyEvent(t, b, "Agent/global/post", 2)
	time.Sleep(50 * time.Millisecond) // 给错误投递留暴露窗口
	st := h.State()
	if st.Received != 1 {
		t.Fatalf("退订后不应新增收到：received=%d", st.Received)
	}
	if len(b.DeadLetters()) == 0 {
		t.Fatal("退订后事件应转死信（可见）")
	}
}

func TestTriggerHubRingOverflow(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(&TriggerOptions{RingSize: 2})
	stop := h.Attach(b, "DocApplied")
	defer stop()

	for i := 1; i <= 3; i++ {
		applyEvent(t, b, fmt.Sprintf("Agent/global/o%d", i), int64(i))
	}
	st := waitState(t, h, 3)
	if len(st.Records) != 2 || st.Dropped != 1 {
		t.Fatalf("环容量 2 应留 2 丢 1：records=%d dropped=%d", len(st.Records), st.Dropped)
	}
	if st.Records[0].Involved != "Agent/global/o2" || st.Records[1].Involved != "Agent/global/o3" {
		t.Fatalf("应丢最旧留 o2/o3: %+v", st.Records)
	}
	if st.Received != 3 {
		t.Fatalf("累计收到含挤出应=3，得 %d", st.Received)
	}
}

// TestTriggerHubAttachDedup 同类型重复 Attach 去重（单订阅者不变量：一类型一条总线订阅）。
func TestTriggerHubAttachDedup(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	stop1 := h.Attach(b, "DocApplied")
	stop2 := h.Attach(b, "DocApplied")
	defer stop2()

	applyEvent(t, b, "Agent/global/probe", 1)
	st := waitState(t, h, 1)
	if len(st.Types) != 1 {
		t.Fatalf("同类型重复 Attach 应去重：types=%v", st.Types)
	}
	stop1() // 两个句柄退订同一条订阅——总线幂等
	applyEvent(t, b, "Agent/global/probe", 2)
	time.Sleep(50 * time.Millisecond)
	if got := h.State().Received; got != 1 {
		t.Fatalf("退订后不应再收到：received=%d", got)
	}
}

// TestTriggerHubPublishStateRace 发布×查询并发（-race 下的数据竞争与一致性底线）。
// 爆发期总线订阅缓冲（16）满 → 事件按可丢语义转死信（event.md P3 level-triggered，
// 系统正确性不依赖事件送达）——断言的是恒等式与可见性，不是「全量必达」。
func TestTriggerHubPublishStateRace(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(&TriggerOptions{RingSize: 32})
	stop := h.Attach(b, "DocApplied")
	defer stop()

	const n = 64
	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			applyEvent(t, b, fmt.Sprintf("Agent/global/r%d", i), int64(i+1))
		}
	}()
	wg.Add(1)
	go func() {
		defer wg.Done()
		for i := 0; i < n; i++ {
			_ = h.State() // 并发读不撕裂
		}
	}()
	wg.Wait()
	// 发布完毕后等接收计数稳定（排干缓冲），再做不变量断言。
	var st TriggerState
	prev := int64(-1)
	deadline := time.After(2 * time.Second)
	for {
		st = h.State()
		if st.Received == prev {
			break
		}
		prev = st.Received
		select {
		case <-deadline:
			t.Fatal("接收计数未在时限内稳定")
		case <-time.After(20 * time.Millisecond):
		}
	}
	if st.Received == 0 || st.Received > n {
		t.Fatalf("收到数应落在 (0,%d]：received=%d", n, st.Received)
	}
	if st.Dropped != st.Received-int64(len(st.Records)) {
		t.Fatalf("恒等式 received=records+dropped 不成立：received=%d records=%d dropped=%d",
			st.Received, len(st.Records), st.Dropped)
	}
	if got := len(b.DeadLetters()); got == 0 && st.Received < n {
		t.Fatalf("未送达事件应全部可见于死信：received=%d dead=%d", st.Received, got)
	}
}
