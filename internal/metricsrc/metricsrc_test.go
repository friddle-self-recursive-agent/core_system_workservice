// metrics 阈值桥单测（v0.31 引擎内部源第四段）：类型注册闸门前置/关断零事件/source 服务
// 侧判定恒 metrics/level 注册表缺省投影（warn）/dedupKey=<metric>:<window> 因由口径/
// 谓词判定服务侧（声明序首个同键阈值）/声明面外谓词不即兴语义。设计依据：
// concepts/event.md §5 + 规则 13（MetricsInfo 阈值判断后发 MetricThreshold——数据面到
// 信号面的桥）。v0.47：桥级单样本流+可替换绑定表（Rebind 热重绑——多源独立排干测试随
// per-source 通道契约退场，重绑语义测试见 rebind_test.go）。
package metricsrc

import (
	"context"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

const metricsFixture = `
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
    - {type: MetricThreshold, source: metrics, defaultLevel: warn, dedup: "<metric>:<window>", route: owner-workflow}
`

const metricsFixtureNoType = `
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

// metricsTestBus 桥测试装配：scheme 载入+总线（窗口 300ms）+全类型订阅通道。
func metricsTestBus(t *testing.T, fixture string) (*eventbus.Bus, *scheme.Scheme, <-chan eventbus.Event) {
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

// TestBridgePublishSemantics 阈值越限桥接语义面：source 服务侧判定恒 metrics（数据面样本
// 不携带来源）、level=注册表缺省投影（warn）、dedupKey=<metric>:<window> 因由口径、
// involved=装配绑定锚。
func TestBridgePublishSemantics(t *testing.T) {
	bus, sch, ch := metricsTestBus(t, metricsFixture)
	smp := make(chan Sample, 1)
	b := New(bus, sch, smp, []Source{{
		Name: "metrics-projection", Involved: "WorkService/metrics",
		Thresholds: []Threshold{{Metric: "apply.failure_rate", Op: ">", Limit: 0.1}},
	}})
	if !b.Enabled() || b.Sources() != 1 {
		t.Fatalf("已注册类型+样本流已挂应 Enabled: %v", b.Enabled())
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go b.Run(ctx)
	smp <- Sample{Metric: "apply.failure_rate", Value: 0.42, Window: "5m"}
	select {
	case ev := <-ch:
		if ev.Type != EventType || ev.Source != EventSource {
			t.Fatalf("类型/来源应=服务侧判定值: {type=%s source=%s}", ev.Type, ev.Source)
		}
		if ev.Involved != "WorkService/metrics" {
			t.Fatalf("involved 应=装配绑定锚，得 %q", ev.Involved)
		}
		if ev.Level != "warn" {
			t.Fatalf("level 应=注册表缺省投影 warn，得 %q", ev.Level)
		}
		if ev.DedupKey != "apply.failure_rate:5m" {
			t.Fatalf("dedupKey 应=<metric>:<window> 因由口径，得 %q", ev.DedupKey)
		}
	case <-time.After(time.Second):
		t.Fatal("阈值越限事件未到（判定未生效）")
	}
}

// TestBridgeNoBreachNoEvent 未越限静默：样本在阈值内=数据面照常产样本零事件（阈值判断
// 后才发事件——§5 数据面到信号面单向桥，不是样本面镜像）。
func TestBridgeNoBreachNoEvent(t *testing.T) {
	bus, sch, ch := metricsTestBus(t, metricsFixture)
	smp := make(chan Sample, 2)
	b := New(bus, sch, smp, []Source{{
		Name: "s", Involved: "WorkService/metrics",
		Thresholds: []Threshold{{Metric: "m", Op: ">", Limit: 1}},
	}})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go b.Run(ctx)
	smp <- Sample{Metric: "m", Value: 0.5, Window: "5m"} // 阈值内
	smp <- Sample{Metric: "m", Value: 1.0, Window: "5m"} // 边界（> 严格小于等于）
	select {
	case ev := <-ch:
		t.Fatalf("未越限样本不应发事件，得 %s", ev.ID)
	case <-time.After(100 * time.Millisecond):
	}
}

// TestBridgeDedupAggregation 同因聚合：窗口内同 <metric>:<window> 第二条被总线合并
// （防越限抖动风暴——event.md 规则 1；聚合语义在总线同主干）。
func TestBridgeDedupAggregation(t *testing.T) {
	bus, sch, ch := metricsTestBus(t, metricsFixture)
	smp := make(chan Sample, 2)
	b := New(bus, sch, smp, []Source{{
		Name: "s", Involved: "WorkService/metrics",
		Thresholds: []Threshold{{Metric: "m", Op: ">=", Limit: 1}},
	}})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go b.Run(ctx)
	smp <- Sample{Metric: "m", Value: 2, Window: "5m"}
	smp <- Sample{Metric: "m", Value: 3, Window: "5m"} // 窗口内同因=合并
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

// TestBridgeUnregisteredGate 类型注册闸门前置：MetricThreshold 未注册=Enabled=false 且
// Run 即返（关断=零事件——声明先行的机械落点，规则 10/12 同范式）。
func TestBridgeUnregisteredGate(t *testing.T) {
	bus, sch, ch := metricsTestBus(t, metricsFixtureNoType)
	smp := make(chan Sample, 1)
	b := New(bus, sch, smp, []Source{{
		Name: "s", Involved: "WorkService/metrics",
		Thresholds: []Threshold{{Metric: "m", Op: ">", Limit: 0}},
	}})
	if b.Enabled() {
		t.Fatal("类型未注册应 Enabled=false（注册表闸门前置）")
	}
	done := make(chan struct{})
	go func() { b.Run(context.Background()); close(done) }()
	smp <- Sample{Metric: "m", Value: 1, Window: "5m"}
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

// TestBridgeNoStream 样本流未挂=关断（Enabled=false——桥无流即无事件源；装配旗标关断的
// 机械判别器关断面）。
func TestBridgeNoStream(t *testing.T) {
	bus, sch, ch := metricsTestBus(t, metricsFixture)
	b := New(bus, sch, nil, nil)
	if b.Enabled() || b.Sources() != 0 {
		t.Fatal("样本流未挂应 Enabled=false（关断=零事件）")
	}
	b.Run(context.Background()) // 即返
	select {
	case ev := <-ch:
		t.Fatalf("无流不应有事件，得 %s", ev.ID)
	case <-time.After(50 * time.Millisecond):
	}
}

// TestBridgeZeroBindingsWaitingState 零绑定=合法等待态（v0.47 热重绑——装配期零声明不再
// 是不启动的终态）：桥挂流运行、样本静默消费零事件；Sources()=0。
func TestBridgeZeroBindingsWaitingState(t *testing.T) {
	bus, sch, ch := metricsTestBus(t, metricsFixture)
	smp := make(chan Sample, 1)
	b := New(bus, sch, smp, nil)
	if !b.Enabled() {
		t.Fatal("零绑定但流已挂+类型已注册 → 应 Enabled=true（等待态非关断——v0.47）")
	}
	if b.Sources() != 0 {
		t.Fatalf("零绑定 Sources()=0，得 %d", b.Sources())
	}
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go b.Run(ctx)
	smp <- Sample{Metric: "m", Value: 1, Window: "5m"} // 零绑定=样本静默消费
	select {
	case ev := <-ch:
		t.Fatalf("零绑定不应发布事件，得 %s", ev.ID)
	case <-time.After(100 * time.Millisecond):
	}
}

// TestBridgeFirstThresholdWins 声明序首个同键阈值参与判定（表序首个同 dispatch 口径）：
// 同指标双阈值时第二阈值不重复判定（同 dedupKey 亦会聚合——判定面单值化避免歧义）。
func TestBridgeFirstThresholdWins(t *testing.T) {
	bus, sch, ch := metricsTestBus(t, metricsFixture)
	smp := make(chan Sample, 1)
	b := New(bus, sch, smp, []Source{{
		Name: "s", Involved: "WorkService/metrics",
		Thresholds: []Threshold{
			{Metric: "m", Op: ">", Limit: 10},  // 首个：0.5 不越限——静默
			{Metric: "m", Op: ">", Limit: 0.1}, // 次个：不参与判定
		},
	}})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go b.Run(ctx)
	smp <- Sample{Metric: "m", Value: 0.5, Window: "5m"}
	select {
	case ev := <-ch:
		t.Fatalf("判定应止于声明序首个同键阈值（静默），得 %s", ev.ID)
	case <-time.After(100 * time.Millisecond):
	}
}

// TestBridgeUnknownOpNoPublish 声明面外谓词不判定不发布不即兴语义（契约面最小集外的
// 谓词随声明面扩展顺延，不静默发明比较语义）。
func TestBridgeUnknownOpNoPublish(t *testing.T) {
	bus, sch, ch := metricsTestBus(t, metricsFixture)
	smp := make(chan Sample, 1)
	b := New(bus, sch, smp, []Source{{
		Name: "s", Involved: "WorkService/metrics",
		Thresholds: []Threshold{{Metric: "m", Op: "==", Limit: 1}},
	}})
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	go b.Run(ctx)
	smp <- Sample{Metric: "m", Value: 1, Window: "5m"}
	select {
	case ev := <-ch:
		t.Fatalf("声明面外谓词不应发布事件，得 %s", ev.ID)
	case <-time.After(100 * time.Millisecond):
	}
}

// TestBridgeCtxCancelExit ctx 取消即退场（幂等可重启——与 patrol/watchsrc.Run 同口径）。
func TestBridgeCtxCancelExit(t *testing.T) {
	bus, sch, _ := metricsTestBus(t, metricsFixture)
	smp := make(chan Sample)
	b := New(bus, sch, smp, []Source{{
		Name: "s", Involved: "WorkService/metrics",
		Thresholds: []Threshold{{Metric: "m", Op: ">", Limit: 0}},
	}})
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
