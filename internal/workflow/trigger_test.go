// trigger_test.go —— 触发骨架验收（dispatch 接线第一段）：
// 收到留痕可查、去重聚合透传、类型过滤、退订排干、环溢出可见、发布×查询并发安全。
package workflow

import (
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
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

// ═══════════ dispatch 第二段（v0.12 订阅面动态化 + for 谓词首段） ═══════════

// pub 发布任意类型事件（connector/external 场景投影；involved=kind/scope/name 同构 objKey）。
func pub(t *testing.T, b *eventbus.Bus, typ, involved string) eventbus.Event {
	t.Helper()
	e := eventbus.Event{Type: typ, Source: "connector", Involved: involved,
		DedupKey: typ + ":" + involved + ":" + time.Now().String(), Level: "info"}
	if err := b.Publish(&e); err != nil {
		t.Fatalf("发布 %s: %v", typ, err)
	}
	return e
}

// waitReceived 等接收计数达预期（订阅 goroutine 异步排干）。
func waitReceived(t *testing.T, h *TriggerHub, want int64) TriggerState {
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

// TestAttachDispatchTableSet 表驱动订阅集合：on 值按表序去重；types=推导集合（判别锚：
// types ≠ 固定 [DocApplied] 而是表行集合）。
func TestAttachDispatchTableSet(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	defer h.AttachDispatch(b, []scheme.DispatchRule{
		{On: "DocApplied", For: "owner:*", Dispatch: "owner-workflow"},
		{On: "ExternalChanged", For: "owner:*", Dispatch: "owner-workflow"},
		{On: "DocApplied", For: "any", Dispatch: "dead-letter"}, // 重复类型去重
	})()

	if got := h.State().Types; len(got) != 2 || got[0] != "DocApplied" || got[1] != "ExternalChanged" {
		t.Fatalf("订阅集合应=[DocApplied ExternalChanged]（表序去重），得 %v", got)
	}
	pub(t, b, "ExternalChanged", "ServiceInstance/developer-system/ci")
	pub(t, b, "DocApplied", "Agent/global/probe")
	st := waitReceived(t, h, 2)
	// 分型计数：留痕按类型各记一条（多类型发布→订阅留痕分型）。
	counts := map[string]int{}
	for _, r := range st.Records {
		counts[r.Type]++
	}
	if counts["DocApplied"] != 1 || counts["ExternalChanged"] != 1 {
		t.Fatalf("分型计数不符: %v", counts)
	}
}

// TestAttachDispatchWildcardCollapse 含通配行 → 坍缩为单条全类型订阅：任意类型都收到一条留痕
// （未注册类型兜底命中），且不因通配+枚举并存而重复计数。
func TestAttachDispatchWildcardCollapse(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	defer h.AttachDispatch(b, []scheme.DispatchRule{
		{On: "DocApplied", For: "owner:*", Dispatch: "owner-workflow"},
		{On: "*", For: "scope:builtin/*", Dispatch: "workflows/builtin-root"},
		{On: "*", For: "any", Dispatch: "dead-letter"},
	})()

	if got := h.State().Types; len(got) != 1 || got[0] != "*" {
		t.Fatalf("含通配应坍缩 types=[*]，得 %v", got)
	}
	// 未注册类型（表内无具体行）→ 通配兜底命中，收到一条（不是三条）。
	pub(t, b, "TemplateFound", "Connector/developer-system/t1")
	st := waitReceived(t, h, 1)
	if st.Records[0].Type != "TemplateFound" {
		t.Fatalf("通配订阅应收到未枚举类型: %+v", st.Records[0])
	}
}

// TestAttachDispatchRowResolution 行解析投影：表序首行胜出，route/for 落留痕（只留痕不派发）；
// scope:builtin/* 命中 builtin scope；未命中任何行 → route=dead-letter 兜底可见；
// owner 字面量行按 owner 解析（v0.13）：global 侧 HeartbeatLost 不命中 agent 行 → 顺延兜底行。
func TestAttachDispatchRowResolution(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	defer h.AttachDispatch(b, []scheme.DispatchRule{
		{On: "DocApplied", For: "owner:*", Dispatch: "owner-workflow"},
		{On: "HeartbeatLost", For: "owner:agent", Dispatch: "parent-agent"}, // owner 解析=global≠agent → 顺延
		{On: "*", For: "scope:builtin/*", Dispatch: "workflows/builtin-root"},
		{On: "*", For: "any", Dispatch: "dead-letter"},
	})()

	pub(t, b, "DocApplied", "Agent/global/probe")                               // 行1
	pub(t, b, "HeartbeatLost", "Agent/global/a1")                               // 行2 owner 解析不等 → 行3 scope global 不命中 → 行4
	pub(t, b, "SessionClosed", "Scheme/builtin/scheme")                         // 行3 scope:builtin/* 命中
	pub(t, b, "MetricThreshold", "ServiceInstance/systems/developer-system/ci") // 行4 兜底
	st := waitReceived(t, h, 4)
	want := []struct{ typ, route, forSel string }{
		{"DocApplied", "owner-workflow", "owner:*"},
		{"HeartbeatLost", "dead-letter", "any"},
		{"SessionClosed", "workflows/builtin-root", "scope:builtin/*"},
		{"MetricThreshold", "dead-letter", "any"},
	}
	for i, w := range want {
		r := st.Records[i]
		if r.Type != w.typ || r.Route != w.route || r.For != w.forSel {
			t.Fatalf("留痕[%d] 行解析不符：want %v 得 {%s %s %s}", i, w, r.Type, r.Route, r.For)
		}
	}
}

// TestAttachDispatchNoMatchDeadLetter 全表未命中（畸形谓词行只行）→ route=dead-letter。
func TestAttachDispatchNoMatchDeadLetter(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	defer h.AttachDispatch(b, []scheme.DispatchRule{
		{On: "DocApplied", For: "bad-predicate", Dispatch: "nowhere"}, // 畸形谓词不命中
	})()

	pub(t, b, "DocApplied", "Agent/global/probe")
	st := waitReceived(t, h, 1)
	if st.Records[0].Route != DeadLetterRoute {
		t.Fatalf("全表未命中应标 dead-letter，得 %q", st.Records[0].Route)
	}
}

// TestAttachDispatchEmptyTable 空表=不订阅：事件收不到留痕、落总线死信（路由不到可见）。
func TestAttachDispatchEmptyTable(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	defer h.AttachDispatch(b, nil)()

	if got := h.State().Types; len(got) != 0 {
		t.Fatalf("空表应零订阅，得 %v", got)
	}
	pub(t, b, "DocApplied", "Agent/global/probe")
	time.Sleep(50 * time.Millisecond)
	if got := h.State().Received; got != 0 {
		t.Fatalf("空表不应有留痕：received=%d", got)
	}
	if len(b.DeadLetters()) == 0 {
		t.Fatal("空表事件应落总线死信（可见）")
	}
}

// TestAttachDispatchUnsubscribeDrain 表驱动订阅的总退订排干（第二段复验第一段语义）。
func TestAttachDispatchUnsubscribeDrain(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	stop := h.AttachDispatch(b, []scheme.DispatchRule{{On: "DocApplied", For: "owner:*", Dispatch: "owner-workflow"}})

	pub(t, b, "DocApplied", "Agent/global/pre")
	waitReceived(t, h, 1)

	stop()
	stop() // 总退订幂等
	pub(t, b, "DocApplied", "Agent/global/post")
	time.Sleep(50 * time.Millisecond)
	if got := h.State().Received; got != 1 {
		t.Fatalf("退订后不应新增留痕：received=%d", got)
	}
	if len(b.DeadLetters()) == 0 {
		t.Fatal("退订后事件应转总线死信")
	}
}

// TestForMatchPredicate 谓词口径单测：any/owner:*、owner 字面量（v0.13 按 owner 解析）、
// scope 前缀式/等值/畸形/未知选择器。
func TestForMatchPredicate(t *testing.T) {
	cases := []struct {
		pred, involved string
		want           bool
	}{
		{"any", "Agent/global/a", true},
		{"owner:*", "Agent/global/a", true},
		{"owner:agent", "Agent/global/a", false}, // owner 解析=global ≠ agent
		{"owner:global", "Agent/global/a", true}, // owner 解析=global（同名=scope 根）
		{"scope:builtin/*", "Scheme/builtin/scheme", true},
		{"scope:builtin/*", "Workflow/builtin/x/y", true}, // 子路径命中
		{"scope:builtin/*", "Agent/global/a", false},      // global 不命中 builtin
		{"scope:global", "Agent/global/a", true},          // 无 /* 等值
		{"scope:global/*", "Agent/global/a", true},        // 前缀式
		{"scope:systems/developer-system/*", "ServiceInstance/systems/developer-system/ci", true},
		{"bad-predicate", "Agent/global/a", false}, // 畸形
		{"kind:Agent", "Agent/global/a", false},    // 未知选择器
		{"any", "", true},
		{"scope:builtin/*", "Scheme", false}, // involved 无 scope 段 → 空 scope 不命中
		{"owner:agent", "Agent", false},      // 无 scope 段 → owner 解析为空，不命中任何字面量
	}
	for _, c := range cases {
		if got := forMatch(c.pred, c.involved); got != c.want {
			t.Errorf("forMatch(%q,%q)=%v want %v", c.pred, c.involved, got, c.want)
		}
	}
}

// ═══════════ 引擎主体第一段（v0.13：owner 解析激活 + 派发决策可查询投影） ═══════════

// TestOwnerOf owner 解析口径（event.md 规则 6）：involved scope 根 systems/<s>→<s>；
// global/builtin 同名（owner=scope 根本身）；无 scope 段=空串。
func TestOwnerOf(t *testing.T) {
	cases := []struct{ involved, want string }{
		{"ServiceInstance/systems/agent/ci", "agent"},
		{"Agent/systems/agent/x", "agent"},
		{"Agent/global/a", "global"},
		{"Scheme/builtin/scheme", "builtin"},
		{"Workflow/builtin/x/y", "builtin"},
		{"ServiceInstance/systems/developer-system/ci", "developer-system"},
		{"Agent", ""}, // 无 scope 段
		{"", ""},
	}
	for _, c := range cases {
		if got := ownerOf(c.involved); got != c.want {
			t.Errorf("ownerOf(%q)=%q want %q", c.involved, got, c.want)
		}
	}
}

// TestForMatchOwnerLiteral owner 字面量谓词激活（v0.13 判别器：owner:agent 行从
// 「恒不命中顺延」（v0.12 首段）变「按 owner 解析命中」）。
func TestForMatchOwnerLiteral(t *testing.T) {
	cases := []struct {
		pred, involved string
		want           bool
	}{
		{"owner:agent", "ServiceInstance/systems/agent/ci", true}, // systems/agent → agent 命中
		{"owner:agent", "Agent/systems/agent/x", true},            // 同上（kind 无关）
		{"owner:agent", "Agent/global/a1", false},                 // global ≠ agent
		{"owner:agent", "Scheme/builtin/scheme", false},           // builtin ≠ agent
		{"owner:global", "Agent/global/a", true},                  // 同名=scope 根
		{"owner:builtin", "Workflow/builtin/x/y", true},           // 同名（子路径仍解析到根）
		{"owner:instance", "SessionInstance/systems/instance/web", true},
		{"owner:agent", "ServiceInstance/systems/developer-system/ci", false}, // 他系统不命中
		{"owner:agent", "", false},                                            // 空 involved 解析为空
		{"owner:*", "Scheme/builtin/scheme", true},                            // 通配恒命中（不变量保持）
	}
	for _, c := range cases {
		if got := forMatch(c.pred, c.involved); got != c.want {
			t.Errorf("forMatch(%q,%q)=%v want %v", c.pred, c.involved, got, c.want)
		}
	}
}

// TestAttachDispatchOwnerLiteralRowHits 表驱动判别器（loop-12 增量）：owner:agent 行按 owner
// 解析命中——HeartbeatLost involved 在 systems/agent 下 → 命中 parent-agent 行（cc 同行投影）；
// 同类型 involved 在 global 下 → 不命中顺延兜底行；owner:instance 行同理。
// （API 无事件注入面——本 fixture 即引擎主体第一段的机械判据，live 侧由 resolve 决策面承载。）
func TestAttachDispatchOwnerLiteralRowHits(t *testing.T) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	defer h.AttachDispatch(b, []scheme.DispatchRule{
		{On: "ExternalChanged", For: "owner:*", Dispatch: "owner-workflow"},
		{On: "HeartbeatLost", For: "owner:agent", Dispatch: "parent-agent", CC: "notification/*"},
		{On: "SessionClosed", For: "owner:instance", Dispatch: "instance-docs"},
		{On: "*", For: "scope:builtin/*", Dispatch: "workflows/builtin-root"},
		{On: "*", For: "any", Dispatch: "dead-letter"},
	})()

	pub(t, b, "HeartbeatLost", "ServiceInstance/systems/agent/ci")     // owner=agent → 行1 命中
	pub(t, b, "HeartbeatLost", "Agent/global/a1")                      // owner=global → 顺延兜底行
	pub(t, b, "SessionClosed", "SessionInstance/systems/instance/web") // owner=instance → 行2 命中
	st := waitReceived(t, h, 3)
	want := []struct{ typ, route, cc, forSel, owner string }{
		{"HeartbeatLost", "parent-agent", "notification/*", "owner:agent", "agent"},
		{"HeartbeatLost", "dead-letter", "", "any", "global"},
		{"SessionClosed", "instance-docs", "", "owner:instance", "instance"},
	}
	for i, w := range want {
		r := st.Records[i]
		if r.Type != w.typ || r.Route != w.route || r.CC != w.cc || r.For != w.forSel || r.Owner != w.owner {
			t.Fatalf("留痕[%d] 行解析不符：want %+v 得 {type:%s route:%s cc:%s for:%s owner:%s}",
				i, w, r.Type, r.Route, r.CC, r.For, r.Owner)
		}
	}
}

// TestResolveDecisionFace 派发决策查询面：纯决策（不注入事件不落留痕不派发）；
// 命中行投影 route/cc/for/owner/row（row 可回指表行）；全表未命中=合成死信 row=NoRow；
// 未挂表=无决策 route 空；决策面零副作用（received 不动）。
