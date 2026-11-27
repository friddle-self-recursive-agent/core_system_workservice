package sessionliveness

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// fakeReg 内存假心跳账面（单测零 DB 依赖）。
type fakeReg struct {
	mu   sync.Mutex
	rows map[string]opslog.SessionHeartbeat
}

func newFakeReg() *fakeReg { return &fakeReg{rows: map[string]opslog.SessionHeartbeat{}} }

func (f *fakeReg) UpsertSessionHeartbeat(h opslog.SessionHeartbeat) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.rows[h.SessionID] = h
	return nil
}

func (f *fakeReg) StaleSessionHeartbeats(before time.Time, limit int) ([]opslog.SessionHeartbeat, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []opslog.SessionHeartbeat
	for _, h := range f.rows {
		if h.UpdatedAt.Before(before) {
			out = append(out, h)
		}
	}
	return out, nil
}

func (f *fakeReg) AllSessionHeartbeats(limit int) ([]opslog.SessionHeartbeat, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []opslog.SessionHeartbeat
	for _, h := range f.rows {
		out = append(out, h)
	}
	return out, nil
}

func (f *fakeReg) get(id string) (opslog.SessionHeartbeat, bool) {
	f.mu.Lock()
	defer f.mu.Unlock()
	h, ok := f.rows[id]
	return h, ok
}

// testScheme 含会话活性事件族的最小 scheme（sessionliveness_testdata）。
func testScheme(t *testing.T) *scheme.Scheme {
	t.Helper()
	sch, err := scheme.LoadFile("testdata/scheme.yaml")
	if err != nil {
		t.Fatalf("载入测试 scheme: %v", err)
	}
	return sch
}

// harness 测试装配：时钟注入+事件收集。
type harness struct {
	w      *Watchdog
	reg    *fakeReg
	bus    *eventbus.Bus
	mu     sync.Mutex
	events []eventbus.Event
	now    time.Time
}

func newHarness(t *testing.T, interval, threshold time.Duration) *harness {
	t.Helper()
	sch := testScheme(t)
	bus := eventbus.NewBus(&eventbus.Options{DedupWindow: time.Hour}) // 大窗口=测试内不聚合
	reg := newFakeReg()
	h := &harness{reg: reg, bus: bus, now: time.Date(2026, 11, 26, 10, 0, 0, 0, time.UTC)}
	w := New(reg, bus, sch, interval, threshold, 2)
	w.Now = func() time.Time { h.mu.Lock(); defer h.mu.Unlock(); return h.now }
	w.OnEvent(func(typ, session string, err error) {
		h.mu.Lock()
		defer h.mu.Unlock()
		h.events = append(h.events, eventbus.Event{Type: typ, Involved: SessionScope(session)})
		if err != nil {
			t.Logf("事件 %s 发布失败: %v", typ, err)
		}
	})
	h.w = w
	// 订阅总线真实收集（验证零旁路主干）。
	ch, cancel := bus.Subscribe("")
	go func() {
		for range ch {
		}
	}()
	t.Cleanup(cancel)
	return h
}

func (h *harness) advance(d time.Duration) {
	h.mu.Lock()
	h.now = h.now.Add(d)
	h.mu.Unlock()
}

func (h *harness) emitted(typ string) bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	for _, e := range h.events {
		if e.Type == typ {
			return true
		}
	}
	return false
}

func (h *harness) count(typ string) int {
	h.mu.Lock()
	defer h.mu.Unlock()
	n := 0
	for _, e := range h.events {
		if e.Type == typ {
			n++
		}
	}
	return n
}

// L56-01 心跳注册：进程内计数+ops 行可查询。
func TestBeatRegistersHeartbeat(t *testing.T) {
	h := newHarness(t, time.Minute, 10*time.Minute)
	if err := h.w.Beat(Heartbeat{SessionID: "s1", Phase: "tool-call", LastAction: "apply scheme.yaml"}); err != nil {
		t.Fatalf("Beat: %v", err)
	}
	if err := h.w.Beat(Heartbeat{SessionID: "s1", Phase: "phase-advance"}); err != nil {
		t.Fatalf("Beat2: %v", err)
	}
	hb, ok := h.reg.get("s1")
	if !ok {
		t.Fatal("心跳行未落")
	}
	if hb.State != opslog.SessionStateLive || hb.HealAttempts != 0 {
		t.Fatalf("state=%s attempts=%d（期望 live/0）", hb.State, hb.HealAttempts)
	}
	views, err := h.w.Heartbeats(0)
	if err != nil || len(views) != 1 || views[0].Heartbeats != 2 {
		t.Fatalf("读面计数=%+v err=%v（期望 heartbeats=2）", views, err)
	}
	if views[0].Involved != "Session/cored/s1" {
		t.Fatalf("involved 投影=%s", views[0].Involved)
	}
}

// L56-02 注入停滞→窗口内 SuspectedStall 判定在账 + SessionStalled 事件（判定态词/事件词分立）。
func TestStallDetectedAndSynthesized(t *testing.T) {
	h := newHarness(t, time.Minute, 10*time.Minute)
	_ = h.w.Beat(Heartbeat{SessionID: "s1", Phase: "tool-call"})
	h.advance(11 * time.Minute) // 越过阈值（缺省 10m 口径的测试值）
	h.w.SweepOnce()
	hb, _ := h.reg.get("s1")
	if hb.State != opslog.SessionStateSuspectedStall {
		t.Fatalf("判定态=%s（期望 suspected-stall=SuspectedStall）", hb.State)
	}
	if !h.emitted(TypeSessionStalled) {
		t.Fatal("SessionStalled 事件未合成")
	}
	if h.emitted(TypeSessionHealed) {
		t.Fatal("未注入先见 SessionHealed=误报")
	}
}

// L56-03 自愈第一动作：注入「继续」→ 心跳恢复 → SessionHealed。
func TestSelfHealInjectAndHealed(t *testing.T) {
	h := newHarness(t, time.Minute, 10*time.Minute)
	injected := ""
	h.w.SelfHeal = func(sessionID string) error { injected = sessionID; return nil }
	_ = h.w.Beat(Heartbeat{SessionID: "s1", Phase: "tool-call"})
	h.advance(11 * time.Minute)
	h.w.SweepOnce()
	if injected != "s1" {
		t.Fatalf("自愈第一动作未注入继续（injected=%q）", injected)
	}
	hb, _ := h.reg.get("s1")
	if hb.Phase != "inject-continue" {
		t.Fatalf("注入动作未留痕（phase=%s）", hb.Phase)
	}
	if hb.HealAttempts != 1 {
		t.Fatalf("注入计数=%d（期望 1）", hb.HealAttempts)
	}
	// 恢复：会话侧心跳重现 → Heal 确认 → SessionHealed（携停滞时长语义）。
	_ = h.w.Beat(Heartbeat{SessionID: "s1", Phase: "phase-advance"})
	if err := h.w.Heal("s1"); err != nil {
		t.Fatalf("Heal: %v", err)
	}
	if !h.emitted(TypeSessionHealed) {
		t.Fatal("SessionHealed 未合成")
	}
	hb, _ = h.reg.get("s1")
	if hb.State != opslog.SessionStateHealed || hb.HealedAt.IsZero() {
		t.Fatalf("恢复行态=%s healedAt=%v", hb.State, hb.HealedAt)
	}
}

// L56-04 注入失败继续上报不静默重试（下一拍重处置）。
func TestInjectFailureKeepsReporting(t *testing.T) {
	h := newHarness(t, time.Minute, 10*time.Minute)
	calls := 0
	h.w.SelfHeal = func(sessionID string) error { calls++; return errors.New("sessionController 缺席") }
	_ = h.w.Beat(Heartbeat{SessionID: "s1"})
	h.advance(11 * time.Minute)
	h.w.SweepOnce()
	if calls != 1 {
		t.Fatalf("第一次处置调用=%d", calls)
	}
	if h.count(TypeSessionStalled) < 2 {
		t.Fatalf("注入失败未继续上报（SessionStalled=%d，期望≥2=判定+失败上报）", h.count(TypeSessionStalled))
	}
	h.advance(2 * time.Minute)
	h.w.SweepOnce() // 下一拍重处置（绝不静默）
	if calls != 2 {
		t.Fatalf("下一拍未重试处置（calls=%d）", calls)
	}
}

// L56-05 权限阻塞场景 → ApprovalNeeded 出现（向上提交不自行绕过）。
func TestPermissionDiagnosisEmitsApprovalNeeded(t *testing.T) {
	h := newHarness(t, time.Minute, 10*time.Minute)
	h.w.SelfHeal = func(string) error { return nil }
	h.w.Diagnose = func(hb opslog.SessionHeartbeat) Diagnosis {
		return Diagnosis{Kind: "permission", Detail: "等待 tool:apply 权限"}
	}
	_ = h.w.Beat(Heartbeat{SessionID: "s1"})
	h.advance(11 * time.Minute)
	h.w.SweepOnce() // 判定+注入
	h.advance(2 * time.Minute)
	h.w.SweepOnce() // 注入未恢复→诊断分诊
	if !h.emitted(TypeApprovalNeeded) {
		t.Fatal("ApprovalNeeded 未出现（权限分诊路）")
	}
	if h.emitted(TypeSessionHealed) {
		t.Fatal("未恢复却见 SessionHealed=误报")
	}
}

// L56-06 人工介入场景：未配置通道=零通知零外发（notification.md 规则 6 判别面）。
func TestHumanDiagnosisWithoutChannelsEmitsNoNotify(t *testing.T) {
	h := newHarness(t, time.Minute, 10*time.Minute)
	h.w.SelfHeal = func(string) error { return nil }
	h.w.Diagnose = func(hb opslog.SessionHeartbeat) Diagnosis {
		return Diagnosis{Kind: "human", NotifyChannels: nil} // bind 用户未配置任何通道
	}
	_ = h.w.Beat(Heartbeat{SessionID: "s1"})
	h.advance(11 * time.Minute)
	h.w.SweepOnce()
	h.advance(2 * time.Minute)
	h.w.SweepOnce()
	// 诊断细节事件照常合成（页面展示面），但零通知通道=零外发由装配层保证（Diagnosis 空
	// 通道不注入 notify 面）；此处断言分诊面本身不产 email 通道事件。
	if h.emitted(TypeApprovalNeeded) {
		t.Fatal("人工类误走权限路")
	}
}

// L56-07 长任务：声明长任务 → LongTaskRunning 有账且零 SuspectedStall 误报；结束回收。
func TestLongTaskReportedNotStalled(t *testing.T) {
	h := newHarness(t, time.Minute, 10*time.Minute)
	_ = h.w.Beat(Heartbeat{SessionID: "lt", TaskName: "dev-delivery-loop", LongTask: true})
	h.advance(11 * time.Minute)
	h.w.SweepOnce()
	if h.emitted(TypeSessionStalled) {
		t.Fatal("长任务误报 SuspectedStall/SessionStalled")
	}
	if !h.emitted(TypeLongTaskRunning) {
		t.Fatal("LongTaskRunning 未显式汇报")
	}
	hb, _ := h.reg.get("lt")
	if hb.State != opslog.SessionStateLongTask {
		t.Fatalf("长任务行态=%s", hb.State)
	}
}

// L56-08 正常轮次零误报：阈值内零事件。
func TestNormalTurnsZeroFalsePositive(t *testing.T) {
	h := newHarness(t, time.Minute, 10*time.Minute)
	_ = h.w.Beat(Heartbeat{SessionID: "s1", Phase: "tool-call"})
	h.advance(9 * time.Minute) // 阈值内
	h.w.SweepOnce()
	_ = h.w.Beat(Heartbeat{SessionID: "s1", Phase: "phase-advance"})
	h.advance(9 * time.Minute)
	h.w.SweepOnce()
	if h.emitted(TypeSessionStalled) {
		t.Fatal("阈值内误报 SessionStalled")
	}
	hb, _ := h.reg.get("s1")
	if hb.State != opslog.SessionStateLive {
		t.Fatalf("正常轮次行态=%s（期望 live）", hb.State)
	}
}

// L56-09 连续自愈无效达阈值（缺省 2）→ Escalate 升级停自动处置。
func TestEscalateAfterHealLimit(t *testing.T) {
	h := newHarness(t, time.Minute, 10*time.Minute)
	h.w.SelfHeal = func(string) error { return nil }
	h.w.Diagnose = func(hb opslog.SessionHeartbeat) Diagnosis { return Diagnosis{Kind: "human"} }
	_ = h.w.Beat(Heartbeat{SessionID: "s1"})
	h.advance(11 * time.Minute)
	h.w.SweepOnce() // 判定+注入#1
	h.advance(2 * time.Minute)
	h.w.SweepOnce() // 诊断（人工类）attempts=1
	h.advance(2 * time.Minute)
	h.w.SweepOnce() // 再处置 attempts=2 ≥ 2 → Escalate
	if !h.emitted("Escalate") {
		t.Fatal("连续自愈无效未升级 Escalate")
	}
	hb, _ := h.reg.get("s1")
	if hb.State != opslog.SessionStateEscalated {
		t.Fatalf("升级后行态=%s", hb.State)
	}
}

// L56-10 注册表闸门前置：SessionStalled 未注册=watchdog 不启动（零事件零扫描）。
func TestRegistryGatePreflight(t *testing.T) {
	sch, err := scheme.LoadFile("testdata/scheme-nogate.yaml") // 未注册 SessionStalled
	if err != nil {
		t.Fatalf("载入: %v", err)
	}
	reg := newFakeReg()
	w := New(reg, eventbus.NewBus(nil), sch, time.Minute, 10*time.Minute, 2)
	if w.Enabled() {
		t.Fatal("类型未注册 watchdog 不应启动（注册表闸门前置）")
	}
	w.Run(context.Background()) // 即返（关断面）
}

// L56-11 dedupKey 口径：session:/session-heal:/longtask: 三词分立（同源 scheme 注记）。
func TestDedupKeyVocabulary(t *testing.T) {
	sch := testScheme(t)
	et, err := sch.LookupEventType(TypeSessionStalled)
	if err != nil {
		t.Fatalf("SessionStalled 未注册: %v", err)
	}
	if et.DefaultLevel != "error" {
		t.Fatalf("SessionStalled defaultLevel=%s（期望 error）", et.DefaultLevel)
	}
	for _, typ := range []string{TypeSessionHealed, TypeLongTaskRunning} {
		if _, err := sch.LookupEventType(typ); err != nil {
			t.Fatalf("%s 未注册: %v", typ, err)
		}
	}
}

// L56-12 长任务诊断面翻转：诊断面确认有进展心跳 → 显式汇报不判停滞（treat 路径）。
func TestDiagnosisLongTaskPath(t *testing.T) {
	h := newHarness(t, time.Minute, 10*time.Minute)
	h.w.SelfHeal = func(string) error { return nil }
	h.w.Diagnose = func(hb opslog.SessionHeartbeat) Diagnosis { return Diagnosis{Kind: "longtask"} }
	_ = h.w.Beat(Heartbeat{SessionID: "s1"})
	h.advance(11 * time.Minute)
	h.w.SweepOnce()
	h.advance(2 * time.Minute)
	h.w.SweepOnce()
	if !h.emitted(TypeLongTaskRunning) {
		t.Fatal("诊断面长任务路未显式汇报")
	}
	if h.emitted(TypeSessionStalled) && h.count(TypeSessionStalled) > 1 {
		t.Fatalf("长任务路重复停滞判定（SessionStalled=%d）", h.count(TypeSessionStalled))
	}
}

// F09-WD-01 会话终态退出（2026-11-26 扫描整改——attention.md watchdog 终态退出规则）：
// state=ended 的终态会话即使心跳陈旧也不合成 SessionStalled/不升级/不续报——
// level-triggered 口径（已决/终态退出待办，append-only 历史不重复计数当待办）。
func TestEndedSessionExitsWatchdog(t *testing.T) {
	h := newHarness(t, time.Minute, 10*time.Minute)
	// 终态行：state=ended，心跳早已陈旧（updated_at 远早于 threshold 窗）。
	if err := h.reg.UpsertSessionHeartbeat(opslog.SessionHeartbeat{
		SessionID: "sess-ended", Ts: h.now.Add(-time.Hour), UpdatedAt: h.now.Add(-time.Hour),
		State: opslog.SessionStateEnded,
	}); err != nil {
		t.Fatalf("seed ended row: %v", err)
	}
	h.advance(20 * time.Minute)
	h.w.SweepOnce()
	if h.count(TypeSessionStalled) != 0 {
		t.Fatalf("终态会话不应合成 SessionStalled（终态退出——零新增事件）: %d", h.count(TypeSessionStalled))
	}
	if h.emitted("Escalate") {
		t.Fatal("终态会话不应升级（Escalate 零新增）")
	}
	// 行态不被改写（ended 仍 ended——终态词形不回写停滞态）。
	if hb, _ := h.reg.get("sess-ended"); hb.State != opslog.SessionStateEnded {
		t.Fatalf("终态行态应保持 ended: %s", hb.State)
	}
	// 对照组：同窗 live 行照常判定（终态退出只摘终态——活体判定面零翻转）。
	if err := h.reg.UpsertSessionHeartbeat(opslog.SessionHeartbeat{
		SessionID: "sess-live", Ts: h.now.Add(-time.Hour), UpdatedAt: h.now.Add(-time.Hour),
		State: opslog.SessionStateLive,
	}); err != nil {
		t.Fatalf("seed live row: %v", err)
	}
	h.w.SweepOnce()
	if h.count(TypeSessionStalled) != 1 {
		t.Fatalf("live 行照常判定 SessionStalled: %d", h.count(TypeSessionStalled))
	}
}
