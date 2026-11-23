// 巡检源单测（v0.22 引擎内部源第二段）：类型注册闸门前置/关断零事件/拍点即因由
// （跨拍不聚合、同拍重放聚合）/source 服务侧判定/level 注册表缺省投影/主循环到点自产。
// 设计依据：concepts/event.md 规则 10 + §7 巡检兜底缓解项。
package patrol

import (
	"context"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

const patrolFixture = `
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
    - {type: ReconcileTick, source: core, defaultLevel: info, dedup: "reconcile:<at>", route: workflows/core-workflow}
`

const patrolFixtureNoTick = `
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

func patrolTestBus(t *testing.T, fixture string) (*eventbus.Bus, *scheme.Scheme, <-chan eventbus.Event, func()) {
	t.Helper()
	sch, err := scheme.Load([]byte(fixture))
	if err != nil {
		t.Fatalf("scheme 载入失败: %v", err)
	}
	bus := eventbus.NewBus(&eventbus.Options{DedupWindow: 300 * time.Millisecond})
	ch, cancel := bus.Subscribe("")
	return bus, sch, ch, cancel
}

// TestPatrolDisabledZeroInterval 关断=零事件（缺省行为面不变；机械判别器关断面）。
func TestPatrolDisabledZeroInterval(t *testing.T) {
	bus, sch, ch, cancel := patrolTestBus(t, patrolFixture)
	defer cancel()
	tk := New(bus, sch, 0)
	if tk.Enabled() {
		t.Fatal("interval=0 应关断（Enabled=false）")
	}
	ctx, stop := context.WithCancel(context.Background())
	stop()
	tk.Run(ctx) // 未开启即返，零事件
	select {
	case ev := <-ch:
		t.Fatalf("关断面不得产事件，收到 %+v", ev)
	case <-time.After(50 * time.Millisecond):
	}
}

// TestPatrolGateUnregistered 注册表闸门前置：类型未注册即不启动（声明先行机械落点）。
func TestPatrolGateUnregistered(t *testing.T) {
	bus, sch, ch, cancel := patrolTestBus(t, patrolFixtureNoTick)
	defer cancel()
	tk := New(bus, sch, time.Millisecond)
	if tk.Enabled() {
		t.Fatal("ReconcileTick 未注册应关断（注册表=引擎内部源唯一开关）")
	}
	ctx, stop := context.WithCancel(context.Background())
	stop()
	tk.Run(ctx)
	select {
	case ev := <-ch:
		t.Fatalf("未注册不得产事件，收到 %+v", ev)
	case <-time.After(50 * time.Millisecond):
	}
}

// TestPatrolRunPublishesPerBeat 主循环到点自产：多拍事件按拍点独立到达，字段=声明投影。
func TestPatrolRunPublishesPerBeat(t *testing.T) {
	bus, sch, ch, cancel := patrolTestBus(t, patrolFixture)
	defer cancel()
	tk := New(bus, sch, 5*time.Millisecond)
	if !tk.Enabled() {
		t.Fatal("已注册∧interval>0 应开启")
	}
	ctx, stop := context.WithCancel(context.Background())
	defer stop()
	go tk.Run(ctx)

	seen := map[string]bool{}
	deadline := time.After(2 * time.Second)
	for len(seen) < 3 {
		select {
		case ev := <-ch:
			if ev.Type != EventType {
				t.Fatalf("类型=%s 期望 %s", ev.Type, EventType)
			}
			if ev.Source != EventSource {
				t.Fatalf("source=%s 期望恒 %s（服务侧判定）", ev.Source, EventSource)
			}
			if ev.Involved != EventInvolved {
				t.Fatalf("involved=%s 期望 %s（对账循环本体）", ev.Involved, EventInvolved)
			}
			if ev.Level != "info" {
				t.Fatalf("level=%s 期望注册表缺省投影 info", ev.Level)
			}
			if len(seen) == 0 && ev.DedupKey[:len("reconcile:")] != "reconcile:" {
				t.Fatalf("dedupKey=%s 期望 reconcile:<at> 拍点前缀", ev.DedupKey)
			}
			seen[ev.DedupKey] = true // 拍点即因由：跨拍 dedupKey 互异=逐拍到达
		case <-deadline:
			t.Fatalf("2s 内仅到 %d 拍（interval=5ms 应远多于 3 拍）", len(seen))
		}
	}
	if len(seen) < 3 {
		t.Fatalf("拍点应互异（跨拍不聚合），实际 %d 个 dedupKey", len(seen))
	}
}

// TestPatrolTickPerBeatDedup 拍点即因由口径：跨拍不聚合（逐拍投递）、同拍重放窗口聚合。
func TestPatrolTickPerBeatDedup(t *testing.T) {
	bus, sch, ch, cancel := patrolTestBus(t, patrolFixture)
	defer cancel()
	tk := New(bus, sch, time.Hour) // Run 不启动；直击 tick 验口径

	t1 := time.Date(2026, 11, 24, 10, 0, 0, 0, time.UTC)
	t2 := t1.Add(time.Minute) // 跨拍：窗口内（300ms 去重窗按墙钟，t1/t2 是注入时刻不同墙钟拍）
	if err := tk.tick(t1); err != nil {
		t.Fatalf("拍1 发布失败: %v", err)
	}
	if err := tk.tick(t2); err != nil {
		t.Fatalf("拍2 发布失败: %v", err)
	}
	select {
	case ev := <-ch:
		if ev.DedupKey == "" {
			t.Fatal("dedupKey 不得为空")
		}
	case <-time.After(time.Second):
		t.Fatal("拍1 未投递")
	}
	select {
	case ev := <-ch:
		if ev.DedupKey == "" {
			t.Fatal("dedupKey 不得为空")
		}
	case <-time.After(time.Second):
		t.Fatal("拍2 未投递（跨拍不得聚合——最迟 N 分钟语义不被去重窗口吞没）")
	}

	// 同拍重放：同 At 二次发布 → 窗口内同因合并（不投递）。
	if err := tk.tick(t2); err != nil {
		t.Fatalf("同拍重放发布失败: %v", err)
	}
	select {
	case ev := <-ch:
		t.Fatalf("同拍重放应窗口聚合，收到 %+v", ev)
	case <-time.After(350 * time.Millisecond): // > 去重窗 300ms
	}
}
