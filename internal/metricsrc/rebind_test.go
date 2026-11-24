// rebind 单测（v0.47，console-adoption.md §1.18 L47-04）：绑定表热重绑语义面——
// Rebind 换表即时生效（旧阈值→新阈值判定切换）/零绑定等待态热激活/DocApplied×WorkService
// 精确过滤（非 WorkService involved 不触发重枚举）/枚举失败保旧表+自愈。设计依据：
// concepts/event.md 规则 13 热重绑注记 + cored-bridge.md v0.47 段。
package metricsrc

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// rebindTestBus 桥测试装配（复用 metricsTestBus——签名显式化便于阅读）。
func rebindTestBus(t *testing.T, fixture string) (*eventbus.Bus, *scheme.Scheme, <-chan eventbus.Event) {
	t.Helper()
	return metricsTestBus(t, fixture)
}

// recvEvent 按类型收事件（全类型订阅通道里 DocApplied 等驱动事件与桥产事件并存——
// 断言只认目标类型，防同 involved 的驱动事件误命中）。
func recvEvent(t *testing.T, ch <-chan eventbus.Event, wantType string) eventbus.Event {
	t.Helper()
	deadline := time.After(2 * time.Second)
	for {
		select {
		case ev := <-ch:
			if ev.Type == wantType {
				return ev
			}
		case <-deadline:
			t.Fatalf("目标类型事件未到: %s", wantType)
		}
	}
}

// assertNoEvent 窗口内零目标类型事件（静默断言——同类误命中防御同上）。
func assertNoEvent(t *testing.T, ch <-chan eventbus.Event, wantType string) {
	t.Helper()
	deadline := time.After(150 * time.Millisecond)
	for {
		select {
		case ev := <-ch:
			if ev.Type == wantType {
				t.Fatalf("不应有 %s 事件，得 %s", wantType, ev.ID)
			}
		case <-deadline:
			return
		}
	}
}

func TestRebindSwapThresholdsImmediately(t *testing.T) {
	bus, sch, ch := rebindTestBus(t, metricsFixture)
	smp := make(chan Sample, 4)
	b := New(bus, sch, smp, []Source{{
		Name: "s", Involved: "WorkService/global/notify",
		Thresholds: []Threshold{{Metric: "m", Op: ">", Limit: 1}},
	}})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go b.Run(ctx)
	// 旧表：m=2 越限（limit 1）→ 事件
	smp <- Sample{Metric: "m", Value: 2, Window: "w1"}
	if ev := recvEvent(t, ch, EventType); ev.Involved != "WorkService/global/notify" {
		t.Fatalf("旧表判定 involved 不符: %q", ev.Involved)
	}
	// 热重绑：阈值收紧到 limit>10 → 同值样本静默（换表即时生效，零重启）
	b.Rebind([]Source{{
		Name: "s", Involved: "WorkService/global/notify",
		Thresholds: []Threshold{{Metric: "m", Op: ">", Limit: 10}},
	}})
	smp <- Sample{Metric: "m", Value: 2, Window: "w2"}
	assertNoEvent(t, ch, EventType)
	// 再热重绑：新增第二绑定（新 involved）→ 越限即发（热激活）
	b.Rebind([]Source{
		{Name: "s", Involved: "WorkService/global/notify", Thresholds: []Threshold{{Metric: "m", Op: ">", Limit: 10}}},
		{Name: "probe", Involved: "WorkService/global/metrics-probe", Thresholds: []Threshold{{Metric: "m", Op: ">=", Limit: 2}}},
	})
	smp <- Sample{Metric: "m", Value: 2, Window: "w3"}
	ev := recvEvent(t, ch, EventType)
	if ev.Involved != "WorkService/global/metrics-probe" {
		t.Fatalf("新绑定判定 involved 不符: %q", ev.Involved)
	}
	if ev.DedupKey != "m:w3" {
		t.Fatalf("dedupKey 应=<metric>:<window>，得 %q", ev.DedupKey)
	}
}

// TestStartRebindFilterAndCoalesce 过滤+合并：非 WorkService involved 不触发重枚举；
// WorkService involved 触发一次全量重枚举（积压事件合并——枚举计数=1 非 N）。
func TestStartRebindFilterAndCoalesce(t *testing.T) {
	bus, sch, ch := rebindTestBus(t, metricsFixture)
	smp := make(chan Sample, 1)
	b := New(bus, sch, smp, nil)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	calls := make(chan int, 8)
	enum := func() ([]Source, error) {
		calls <- 1
		return []Source{{Name: "s", Involved: "WorkService/global/notify", Thresholds: []Threshold{{Metric: "m", Op: ">", Limit: 0}}}}, nil
	}
	applied := make(chan int, 8)
	StartRebind(ctx, bus, b, enum, RebindHooks{Applied: func(n int) { applied <- n }})
	time.Sleep(50 * time.Millisecond) // 订阅挂接窗
	// 非 WorkService 对象 apply：不触发重枚举
	bus.Publish(&eventbus.Event{Type: "DocApplied", Involved: "Workflow/global/dev-delivery"})
	bus.Publish(&eventbus.Event{Type: "MetricThreshold", Involved: "WorkService/global/notify", Source: EventSource})
	time.Sleep(120 * time.Millisecond)
	select {
	case <-calls:
		t.Fatal("非 WorkService involved 不应触发重枚举")
	case <-time.After(50 * time.Millisecond):
	}
	// WorkService apply：触发重枚举+绑定表替换
	bus.Publish(&eventbus.Event{Type: "DocApplied", Involved: "WorkService/global/notify"})
	select {
	case n := <-applied:
		if n != 1 {
			t.Fatalf("重绑后绑定数=1，得 %d", n)
		}
	case <-time.After(time.Second):
		t.Fatal("WorkService apply 未触发热重绑（Applied 回调未到）")
	}
	select {
	case <-calls:
	default:
		t.Fatal("枚举闭包未被调用")
	}
	// 绑定表已换：样本越限即发（零绑定等待态被 apply 热激活的机械面）
	smp <- Sample{Metric: "m", Value: 1, Window: "w1"}
	if ev := recvEvent(t, ch, EventType); ev.Involved != "WorkService/global/notify" {
		t.Fatalf("热激活后 involved 不符: %q", ev.Involved)
	}
}

// TestStartRebindFailureKeepsOldTable 枚举失败=保旧表（Failed 回调+旧表继续判定——
// 不半装载不打死运行中判定面，权限热装载 declared 同款）。
func TestStartRebindFailureKeepsOldTable(t *testing.T) {
	bus, sch, ch := rebindTestBus(t, metricsFixture)
	smp := make(chan Sample, 2)
	b := New(bus, sch, smp, []Source{{
		Name: "s", Involved: "WorkService/global/notify",
		Thresholds: []Threshold{{Metric: "m", Op: ">", Limit: 1}},
	}})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go b.Run(ctx) // 判定主循环在位——枚举失败只影响重绑，不打死运行中判定面
	boom := errors.New("枚举失败（模拟文档库故障）")
	fail := true
	enum := func() ([]Source, error) {
		if fail {
			return nil, boom
		}
		return []Source{{Name: "s", Involved: "WorkService/global/new", Thresholds: []Threshold{{Metric: "x", Op: ">", Limit: 0}}}}, nil
	}
	failed := make(chan error, 8)
	StartRebind(ctx, bus, b, enum, RebindHooks{Failed: func(err error) { failed <- err }})
	time.Sleep(50 * time.Millisecond)
	bus.Publish(&eventbus.Event{Type: "DocApplied", Involved: "WorkService/global/notify"})
	select {
	case err := <-failed:
		if !errors.Is(err, boom) {
			t.Fatalf("Failed 回调错误不符: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("枚举失败未回调 Failed")
	}
	// 旧表原样：越限判定继续（involved 不变——保旧表非清表）
	smp <- Sample{Metric: "m", Value: 2, Window: "w1"}
	if ev := recvEvent(t, ch, EventType); ev.Involved != "WorkService/global/notify" {
		t.Fatalf("失败后应保旧表继续判定，得 involved=%q", ev.Involved)
	}
	// 自愈：下次 apply（枚举恢复）→ 新表生效
	fail = false
	bus.Publish(&eventbus.Event{Type: "DocApplied", Involved: "WorkService/global/notify"})
	time.Sleep(150 * time.Millisecond)
	smp <- Sample{Metric: "x", Value: 1, Window: "w2"}
	if ev := recvEvent(t, ch, EventType); ev.Involved != "WorkService/global/new" {
		t.Fatalf("重试自愈后应换新表，得 involved=%q", ev.Involved)
	}
}

// TestStartRebindNilGuards nil 防御：nil bus/bridge/enum 即返不启动（零 panic 面）。
func TestStartRebindNilGuards(t *testing.T) {
	StartRebind(context.Background(), nil, &Bridge{}, func() ([]Source, error) { return nil, nil }, RebindHooks{})
	StartRebind(context.Background(), &eventbus.Bus{}, nil, func() ([]Source, error) { return nil, nil }, RebindHooks{})
	bus, sch, _ := rebindTestBus(t, metricsFixture)
	StartRebind(context.Background(), bus, New(bus, sch, nil, nil), nil, RebindHooks{})
}
