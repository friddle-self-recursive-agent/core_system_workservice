// heartbeat_test.go —— BUG-010 底层心跳源启停/发布机读锚（event.md 规则 18：
// tier=infra + visibility=default-hidden 一等建模）。
package heartbeat

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// testScheme 最小 scheme：Heartbeat 已注册（含 tier/visibility）或按需裁剪。
func testScheme(t *testing.T, withHeartbeat bool) *scheme.Scheme {
	t.Helper()
	evt := ""
	if withHeartbeat {
		evt = "  eventTypes:\n    - {type: Heartbeat, source: \"agent|core\", defaultLevel: debug, tier: infra, visibility: default-hidden, dedup: \"heartbeat:<subject>:<window>\", route: dead-letter}\n"
	}
	// kinds 至少含自注册 Scheme（装载器硬校验）；dispatch 空表=Heartbeat 走死信兜底。
	data := `apiVersion: core/v1
kind: Scheme
metadata: {name: test, scope: global}
spec:
  version: v0-test
  kinds:
    - kind: Scheme
      builtin: true
      storage: core/scheme.yaml
      spec: {version: "!string", kinds: "[]KindReg"}
  docTypes: []
  dispatch: []
` + evt
	sch, err := scheme.Load([]byte(data))
	if err != nil {
		t.Fatalf("装载测试 scheme: %v", err)
	}
	return sch
}

func TestTickerEnabledGate(t *testing.T) {
	bus := eventbus.NewBus(&eventbus.Options{})
	// 未注册事件类型=不启动（注册表闸门前置）。
	if New(bus, testScheme(t, false), time.Second).Enabled() {
		t.Fatal("Heartbeat 未注册时应 Enabled=false")
	}
	// interval<=0=关断（缺省不改行为面）。
	if New(bus, testScheme(t, true), 0).Enabled() {
		t.Fatal("interval=0 时应 Enabled=false（缺省关断）")
	}
	if !New(bus, testScheme(t, true), time.Second).Enabled() {
		t.Fatal("已注册+interval>0 应 Enabled=true")
	}
}

func TestTickerPublishesHeartbeat(t *testing.T) {
	bus := eventbus.NewBus(&eventbus.Options{DedupWindow: time.Millisecond})
	sch := testScheme(t, true)
	tk := New(bus, sch, time.Hour) // 手动 tick，不走 Run 定时
	if err := tk.tick(time.Now()); err != nil {
		t.Fatalf("tick 发布失败: %v", err)
	}
	dl := bus.DeadLetters()
	if len(dl) == 0 {
		t.Fatal("Heartbeat route=dead-letter 应落死信（零分发兜底出口）")
	}
	got := dl[len(dl)-1]
	if got.Type != EventType || got.Source != EventSource {
		t.Fatalf("事件形状 Type/Source=%s/%s 期望 %s/%s", got.Type, got.Source, EventType, EventSource)
	}
	if got.Scope != "global" {
		t.Fatalf("metadata.scope 应为 global（involved 派生回填），得 %q", got.Scope)
	}
	if !strings.HasPrefix(got.DedupKey, "heartbeat:"+Subject+":") {
		t.Fatalf("dedup 词形 %q 期望 heartbeat:<subject>:<window>", got.DedupKey)
	}
	if got.Level != "debug" {
		t.Fatalf("level=注册表缺省投影应 debug（rule 1b 不入归档），得 %q", got.Level)
	}
}

func TestTickerRunStopsOnCancel(t *testing.T) {
	bus := eventbus.NewBus(&eventbus.Options{})
	tk := New(bus, testScheme(t, true), 10*time.Millisecond)
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { tk.Run(ctx); close(done) }()
	time.Sleep(30 * time.Millisecond)
	cancel()
	select {
	case <-done:
	case <-time.After(time.Second):
		t.Fatal("Run 应随 ctx 取消退场")
	}
	if n := len(bus.DeadLetters()); n < 2 {
		t.Fatalf("10ms 拍内 30ms 应至少发 2 拍，死信=%d", n)
	}
}
