// selfheal_bridge_test.go —— v0.56 R-10 自愈注入桥真实接线单测（agent-session.md §7/§8：
// 注入停滞→watchdog 第一动作=sessionController.prompt 注入「继续」（healed-by-beat 词形留痕）
// ∧失败继续上报绝不静默∧纯观测部署（healed-self）回归不破∧判定面自续租禁入）。
package sessionliveness

import (
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
)

// fakeBridge 自愈注入桥假实现（宿主 sessionController.prompt 面的判别替身——
// 本包唯一依赖方向=SelfHealBridge 接口，零宿主 import）。
type fakeBridge struct {
	mu     sync.Mutex
	calls  []string // 收到的注入词序列（按会话调用次序）
	word   string
	failN  int // 前 N 次 Bounce 返回错误（失败上报路径用）
	lastID string
}

func (f *fakeBridge) Bounce(sessionID string) error {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.failN > 0 {
		f.failN--
		return errors.New("宿主 prompt 面暂不可达（判别桩）")
	}
	f.calls = append(f.calls, sessionID)
	f.lastID = sessionID
	return nil
}

func (f *fakeBridge) InjectWord() string { return f.word }

// L58-01 桥接线：注入停滞→第一动作经桥注入固定词「继续」∧注入受理落观测续租行
// （phase 词=inject-continue 等价词形：last_action 前缀注入）∧恢复=会话侧真实推进心跳
// （healed-by-beat 词形——attempts 清零+SessionHealed）∧纯观测词形（无桥）回归不破。
func TestSelfHealBridgeInjectsContinue(t *testing.T) {
	h := newHarness(t, time.Minute, 10*time.Minute)
	br := &fakeBridge{word: DefaultHealWord}
	h.w.SetSelfHealBridge(br)
	if h.w.SelfHealV != br {
		t.Fatal("SetSelfHealBridge 未装配（SelfHealV 词形缺席）")
	}
	if err := h.w.Beat(Heartbeat{SessionID: "s1"}); err != nil {
		t.Fatal(err)
	}
	h.advance(11 * time.Minute)
	h.w.SweepOnce() // 判定+桥注入
	br.mu.Lock()
	calls := len(br.calls)
	br.mu.Unlock()
	if calls != 1 {
		t.Fatalf("桥注入次数=%d（期望 1——第一动作经宿主 sessionController.prompt）", calls)
	}
	// 注入受理=watchdog 侧观测续租落行（不触发 SessionHealed——恢复判据=会话侧真实推进）。
	hb, _ := h.reg.get("s1")
	if hb.LastAction == "" || hb.UpdatedAt.IsZero() {
		t.Fatalf("注入受理观测续租行缺席：%+v", hb)
	}
	if h.emitted(TypeSessionHealed) {
		t.Fatal("注入受理即记 SessionHealed=冒充恢复（恢复判据=会话侧真实推进心跳）")
	}
	// 会话侧真实推进心跳（source=watchdog 不可经 HTTP 面——此处直注 Beat）→ healed-by-beat。
	if err := h.w.Beat(Heartbeat{SessionID: "s1", Phase: "phase-advance", Source: "session"}); err != nil {
		t.Fatal(err)
	}
	if !h.emitted(TypeSessionHealed) {
		t.Fatal("恢复后 SessionHealed 未在账（healed-by-beat 词形缺席）")
	}
	hb, _ = h.reg.get("s1")
	if hb.HealAttempts != 0 {
		t.Fatalf("恢复后注入计数未清零=%d", hb.HealAttempts)
	}
}

// L58-02 失败绝不静默：桥 Bounce 失败→stderr 告警词形（self-heal-inject-failed 处置留痕）
// ∧行态保持 suspected-stall（attempts=0，下一拍重试第一动作=重试不是跳步）。
func TestSelfHealBridgeFailureKeepsReporting(t *testing.T) {
	h := newHarness(t, time.Minute, 10*time.Minute)
	br := &fakeBridge{word: DefaultHealWord, failN: 2}
	h.w.SetSelfHealBridge(br)
	_ = h.w.Beat(Heartbeat{SessionID: "s1"})
	h.advance(11 * time.Minute)
	h.w.SweepOnce() // 注入失败#1
	hb, _ := h.reg.get("s1")
	if hb.State != opslog.SessionStateSuspectedStall {
		t.Fatalf("失败后行态=%s（期望 suspected-stall 照实保持）", hb.State)
	}
	h.advance(2 * time.Minute)
	h.w.SweepOnce() // 重试第一动作（失败#2）
	br.mu.Lock()
	failLeft := br.failN
	br.mu.Unlock()
	if failLeft != 0 {
		t.Fatalf("失败注入未被重试消耗（剩余=%d——下一拍重试第一动作语义破缺）", failLeft)
	}
	h.advance(2 * time.Minute)
	h.w.SweepOnce() // 第三拍注入成功
	br.mu.Lock()
	okCalls := len(br.calls)
	br.mu.Unlock()
	if okCalls != 1 {
		t.Fatalf("成功注入次数=%d（期望 1——失败两拍后第三拍成功）", okCalls)
	}
}

// L58-03 纯观测回归不破：无桥（SetSelfHealBridge 未调用）→ SelfHeal=nil 装配词形不变
// （判定事件在账、零注入、healed-self 恢复判定照旧）。
func TestNoBridgePureObserveUnchanged(t *testing.T) {
	h := newHarness(t, time.Minute, 10*time.Minute)
	_ = h.w.Beat(Heartbeat{SessionID: "s1"})
	h.advance(11 * time.Minute)
	h.w.SweepOnce() // 判定（纯观测——零注入）
	if h.w.SelfHealV != nil {
		t.Fatal("无桥装配 SelfHealV 应为 nil")
	}
	hb, _ := h.reg.get("s1")
	if hb.Phase == "inject-continue" {
		t.Fatal("纯观测装配不得有注入词形落行")
	}
	h.advance(2 * time.Minute)
	_ = h.w.Beat(Heartbeat{SessionID: "s1"}) // 自行恢复
	if !h.emitted(TypeSessionHealed) {
		t.Fatal("纯观测恢复判定破缺（healed-self 词形缺席）")
	}
}

// L58-04 判定面自续租禁入：watchdog 侧动作续租（BeatSource）不触发 SessionHealed 也不清
// attempts——恢复判据唯一=会话侧真实推进心跳（心跳面唯一口径）。
func TestBeatSourceDoesNotHeal(t *testing.T) {
	h := newHarness(t, time.Minute, 10*time.Minute)
	br := &fakeBridge{word: DefaultHealWord}
	h.w.SetSelfHealBridge(br)
	_ = h.w.Beat(Heartbeat{SessionID: "s1"})
	h.advance(11 * time.Minute)
	h.w.SweepOnce() // 判定+注入（受理续租落行）
	if err := h.w.BeatSource("s1", "注入固定 prompt「继续」（自愈第一动作·受理）"); err != nil {
		t.Fatalf("BeatSource 落行失败: %v", err)
	}
	if h.emitted(TypeSessionHealed) {
		t.Fatal("watchdog 观测续租触发 SessionHealed=判定面自续租（禁入）")
	}
	// 未知会话=照实报错（不发明行）。
	if err := h.w.BeatSource("ghost", "x"); err == nil {
		t.Fatal("无行会话 BeatSource 应报错（不冒充锚）")
	}
}
