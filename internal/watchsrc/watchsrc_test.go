// watch 桥单测（v0.30 引擎内部源第三段）：类型注册闸门前置/关断零事件/source 服务侧判定
// 恒 connector/level 注册表缺省投影/dedupKey=驱动侧因由声明透传（同因聚合防抖动风暴）/
// 多源独立排干/驱动关频道源退场不殃及他源。设计依据：concepts/event.md 规则 12 + §1 五源
// 之首（Connector watch 外部变化）。
package watchsrc

import (
	"context"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/connectors"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

const watchFixture = `
apiVersion: core/v1
kind: Scheme
metadata: {name: default, scope: builtin}
spec:
  version: core/v1
  kinds:
    - kind: Scheme
      builtin: true
      storage: core/scheme.yaml
      spec: {version: string}
  eventTypes:
    - {type: ExternalChanged, source: connector, defaultLevel: info, dedup: "<obj>:<change>:hash", route: owner-workflow}
`

const watchFixtureNoType = `
apiVersion: core/v1
kind: Scheme
metadata: {name: default, scope: builtin}
spec:
  version: core/v1
  kinds:
    - kind: Scheme
      builtin: true
      storage: core/scheme.yaml
      spec: {version: string}
  eventTypes:
    - {type: DocApplied, source: core, defaultLevel: info, dedup: "apply", route: owner-workflow}
`

// watchTestBus 桥测试装配：scheme 载入+总线（窗口 300ms）+全类型订阅通道。
func watchTestBus(t *testing.T, fixture string) (*eventbus.Bus, *scheme.Scheme, <-chan eventbus.Event) {
	t.Helper()
	sch, err := scheme.Load([]byte(fixture))
	if err != nil {
		t.Fatalf("scheme 载入失败: %v", err)
	}
	bus := eventbus.NewBus(&eventbus.Options{DedupWindow: 300 * time.Millisecond})
	ch, cancel := bus.Subscribe("")
	t.Cleanup(cancel)
	return bus, sch, ch
}

// TestBridgePublishSemantics 桥接语义面：source 服务侧判定恒 connector（驱动载荷不携带
// 来源——connectors.Event 无 source 字段）、level=注册表缺省投影（info）、dedupKey=驱动
// 侧因由声明透传、involved=装配绑定锚、detail 不进总线（规则 2——事件只带指针）。
func TestBridgePublishSemantics(t *testing.T) {
	bus, sch, ch := watchTestBus(t, watchFixture)
	drv := make(chan connectors.Event, 1)
	b := New(bus, sch, []Source{{Name: "onedev-driver", Involved: "WorkService/onedev", Events: drv}})
	if !b.Enabled() || b.Sources() != 1 {
		t.Fatalf("已注册类型+有源应 Enabled: %v", b.Enabled())
	}
	drv <- connectors.Event{DedupKey: "onedev:runner-label:abc123", Detail: "runner 标签变更"}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go b.Run(ctx)
	select {
	case ev := <-ch:
		if ev.Type != EventType || ev.Source != EventSource {
			t.Fatalf("类型/来源应=服务侧判定值: {type=%s source=%s}", ev.Type, ev.Source)
		}
		if ev.Involved != "WorkService/onedev" {
			t.Fatalf("involved 应=装配绑定锚，得 %q", ev.Involved)
		}
		if ev.Level != "info" {
			t.Fatalf("level 应=注册表缺省投影 info，得 %q", ev.Level)
		}
		if ev.DedupKey != "onedev:runner-label:abc123" {
			t.Fatalf("dedupKey 应=驱动侧因由声明透传，得 %q", ev.DedupKey)
		}
	case <-time.After(time.Second):
		t.Fatal("桥接事件未到（排干未生效）")
	}
}

// TestBridgeDedupAggregation 同因聚合：窗口内同 dedupKey 第二条被总线合并（防外部抖动
// 风暴——event.md 规则 1；服务侧透传因由、聚合语义在总线同主干）。
func TestBridgeDedupAggregation(t *testing.T) {
	bus, sch, ch := watchTestBus(t, watchFixture)
	drv := make(chan connectors.Event, 2)
	b := New(bus, sch, []Source{{Name: "s", Involved: "WorkService/x", Events: drv}})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go b.Run(ctx)
	drv <- connectors.Event{DedupKey: "x:label:same"}
	drv <- connectors.Event{DedupKey: "x:label:same"} // 窗口内同因=合并
	deadline := time.After(time.Second)
	for seen := 0; seen < 1; {
		select {
		case <-ch:
			seen++
		case <-deadline:
			t.Fatal("首条事件未到")
		}
	}
	select {
	case ev := <-ch:
		t.Fatalf("窗口内同因第二条应被聚合，却收到 %s", ev.ID)
	case <-time.After(50 * time.Millisecond):
	}
}

// TestBridgeUnregisteredGate 类型注册闸门前置：ExternalChanged 未注册=Enabled=false 且
// Run 即返（关断=零事件——声明先行的机械落点，规则 10 同范式）。
func TestBridgeUnregisteredGate(t *testing.T) {
	bus, sch, ch := watchTestBus(t, watchFixtureNoType)
	drv := make(chan connectors.Event, 1)
	b := New(bus, sch, []Source{{Name: "s", Involved: "WorkService/x", Events: drv}})
	if b.Enabled() {
		t.Fatal("类型未注册应 Enabled=false（注册表闸门前置）")
	}
	done := make(chan struct{})
	go func() { b.Run(context.Background()); close(done) }()
	drv <- connectors.Event{DedupKey: "x:y:z"}
	select {
	case <-done:
	case <-time.After(200 * time.Millisecond):
		t.Fatal("未注册桥应即返（闸门前置不启动）")
	}
	select {
	case ev := <-ch:
		t.Fatalf("未注册不应发布事件，得 %s", ev.ID)
	case <-time.After(50 * time.Millisecond):
	}
}

// TestBridgeNoSources 关断=零源零事件（装配旗标关断的机械判别器关断面）。
func TestBridgeNoSources(t *testing.T) {
	bus, sch, ch := watchTestBus(t, watchFixture)
	b := New(bus, sch, nil)
	if b.Enabled() || b.Sources() != 0 {
		t.Fatal("零源应 Enabled=false（关断=零事件）")
	}
	b.Run(context.Background()) // 即返
	select {
	case ev := <-ch:
		t.Fatalf("零源不应有事件，得 %s", ev.ID)
	case <-time.After(50 * time.Millisecond):
	}
}

// TestBridgeMultiSourceIndependence 多源独立：两源各自排干互不阻塞；驱动侧关频道=该源
// 退场，其余源继续（可重启语义由装配层决定——桥不代管重连）。
func TestBridgeMultiSourceIndependence(t *testing.T) {
	bus, sch, ch := watchTestBus(t, watchFixture)
	a := make(chan connectors.Event)
	c := make(chan connectors.Event, 1)
	b := New(bus, sch, []Source{
		{Name: "gone", Involved: "WorkService/gone", Events: a},
		{Name: "live", Involved: "WorkService/live", Events: c},
	})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go b.Run(ctx)
	close(a) // 第一源退场
	c <- connectors.Event{DedupKey: "live:evt:1"}
	select {
	case ev := <-ch:
		if ev.Involved != "WorkService/live" {
			t.Fatalf("存活源应继续桥接，得 involved=%q", ev.Involved)
		}
	case <-time.After(time.Second):
		t.Fatal("一源退场不应殃及他源（存活源事件未到）")
	}
}

// TestBridgeCtxCancelExit ctx 取消即退场（幂等可重启——与 patrol.Run 同口径）。
func TestBridgeCtxCancelExit(t *testing.T) {
	bus, sch, _ := watchTestBus(t, watchFixture)
	drv := make(chan connectors.Event)
	b := New(bus, sch, []Source{{Name: "s", Involved: "WorkService/x", Events: drv}})
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { b.Run(ctx); close(done) }()
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("ctx 取消后 Run 应退场")
	}
}
