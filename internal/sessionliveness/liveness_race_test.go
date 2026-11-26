// liveness_race_test.go —— G-04 数据竞争修复锚（design-code-gap-2026-11-26.md G-04；
// 机械判据=`go test -race ./internal/sessionliveness/` 全绿）：
//   - ①Beat/BeatSource 与 sweep 并发：prevStalled 原子判定摘除（healed-self 至多一路）、
//     attempts 裸读改锁内；
//   - ②SetSelfHealBridge 与 sweep 并发：装配回填全程锁内（解锁后写 SelfHealV/SelfHeal 消除）；
//   - ③僵尸心跳条目 GC：EntryTTL 过后进程内五账零残留（存活会话条目保留）；
//   - ④Escalated 续停滞观测每升级幕至多一条（恢复重置后可再报——状态去重非永久静默）。
package sessionliveness

import (
	"sync"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
)

// G-04① 并发 Beat/BeatSource×sweep：-race 判定面（prevStalled 读改原子判定摘除 +
// attempts 锁内读）。收敛断言：无 panic、行态闭合（live/suspected/healed 之一）、
// 恢复事件不越 Beat 上界。
func TestRaceBeatConcurrentWithSweep(t *testing.T) {
	h := newHarness(t, time.Minute, 10*time.Minute)
	if err := h.w.Beat(Heartbeat{SessionID: "s1", Phase: "tool-call"}); err != nil {
		t.Fatal(err)
	}
	h.advance(11 * time.Minute) // 越阈：sweep 侧开始判停滞（prevStalled 写、attempts 读写）
	const beatWorkers = 8
	const beatsPerWorker = 50
	stop := make(chan struct{})
	var workers, sweeper sync.WaitGroup
	sweeper.Add(1)
	go func() { // sweep 循环（handleStale→markStalled/treat 路径）
		defer sweeper.Done()
		for {
			select {
			case <-stop:
				return
			default:
				h.w.SweepOnce()
			}
		}
	}()
	workers.Add(1)
	go func() { // BeatSource 观测续租（attempts 锁内读竞争面）
		defer workers.Done()
		for i := 0; i < 200; i++ {
			_ = h.w.BeatSource("s1", "注入固定 prompt「继续」（自愈第一动作·受理）")
		}
	}()
	for i := 0; i < beatWorkers; i++ {
		workers.Add(1)
		go func() { // 并发 Beat（prevStalled 原子判定摘除 + attempts 恢复清零竞争面）
			defer workers.Done()
			for j := 0; j < beatsPerWorker; j++ {
				_ = h.w.Beat(Heartbeat{SessionID: "s1", Phase: "tool-call"})
			}
		}()
	}
	workers.Wait()
	close(stop)
	sweeper.Wait()
	hb, ok := h.reg.get("s1")
	if !ok {
		t.Fatal("并发压测后心跳行缺席")
	}
	switch hb.State {
	case opslog.SessionStateLive, opslog.SessionStateSuspectedStall, opslog.SessionStateHealed:
	default:
		t.Fatalf("并发压测后行态越闭合集=%s", hb.State)
	}
	heals := h.count(TypeSessionHealed)
	beats := beatWorkers * beatsPerWorker
	if heals > beats { // 每 Beat 至多一恢复事件（含 healed-self 至多一路的朴素上界）
		t.Fatalf("恢复事件数越上界：%d > %d", heals, beats)
	}
}

// G-04② SetSelfHealBridge 与 sweep 并发：装配回填全程锁内（桥运行态回填与判定读取
// 同临界区——解锁后写 SelfHealV/SelfHeal 的竞争消除）。收敛断言：回填落定后查询词形
// 与执行面一致（SelfHealV=桥 ∧ SelfHeal=桥路由）。
func TestRaceSetSelfHealBridgeConcurrentWithSweep(t *testing.T) {
	h := newHarness(t, time.Minute, 10*time.Minute)
	if err := h.w.Beat(Heartbeat{SessionID: "s1", Phase: "tool-call"}); err != nil {
		t.Fatal(err)
	}
	h.advance(11 * time.Minute)
	br := &fakeBridge{word: DefaultHealWord}
	stop := make(chan struct{})
	var workers, sweeper sync.WaitGroup
	sweeper.Add(1)
	go func() { // sweep 循环（markStalled/retryInject 经 selfHealFn 锁内读执行面）
		defer sweeper.Done()
		for {
			select {
			case <-stop:
				return
			default:
				h.w.SweepOnce()
			}
		}
	}()
	workers.Add(1)
	go func() { // 桥装配回填循环（桥↔纯观测交替——回填/摘除与判定并发）
		defer workers.Done()
		for i := 0; i < 200; i++ {
			h.w.SetSelfHealBridge(br)
			if i%2 == 0 {
				h.w.SetSelfHealBridge(nil) // nil=桥摘除（SelfHeal 保持桥路由但 Bounce 报未装配——照实失败上报）
			}
		}
	}()
	workers.Wait()
	close(stop)
	sweeper.Wait()
	h.w.mu.Lock()
	v, fn := h.w.SelfHealV, h.w.SelfHeal
	h.w.mu.Unlock()
	if v != br || fn == nil {
		t.Fatalf("桥回填收敛词形破缺：SelfHealV=%v SelfHeal=%v（期望桥+桥路由）", v, fn != nil)
	}
}

// G-04③ 僵尸心跳条目 GC：EntryTTL 过后（本测缩窗 1h 复现「陈旧心跳 30m 后消失」同型
// 判据——生产缺省 24h），超期无 beat 会话的进程内五账条目摘除；续租会话条目保留。
func TestStaleEntryGCCollectsZombieMaps(t *testing.T) {
	h := newHarness(t, time.Minute, 10*time.Minute)
	h.w.EntryTTL = time.Hour // 测试缩窗（生产缺省 defaultEntryTTL=24h）
	if err := h.w.Beat(Heartbeat{SessionID: "zombie", Phase: "tool-call"}); err != nil {
		t.Fatal(err)
	}
	if err := h.w.Beat(Heartbeat{SessionID: "live", Phase: "tool-call"}); err != nil {
		t.Fatal(err)
	}
	h.advance(11 * time.Minute)
	h.w.SweepOnce() // 双会话判停滞（prevStalled 双双在账）
	if err := h.w.Beat(Heartbeat{SessionID: "live", Phase: "phase-advance"}); err != nil {
		t.Fatal(err)
	} // live 自行恢复（续租 + prevStalled 摘除）
	h.advance(2 * time.Hour) // zombie 行 updated_at 停在 10:00 → 超 EntryTTL
	if err := h.w.Beat(Heartbeat{SessionID: "live", Phase: "tool-call"}); err != nil {
		t.Fatal(err)
	} // live 续租（保留窗内）
	h.w.SweepOnce() // GC 附带执行
	h.w.mu.Lock()
	defer h.w.mu.Unlock()
	if _, ok := h.w.counts["zombie"]; ok {
		t.Fatal("僵尸会话 counts 条目未清扫")
	}
	if _, ok := h.w.prevStalled["zombie"]; ok {
		t.Fatal("僵尸会话 prevStalled 条目未清扫")
	}
	if _, ok := h.w.escalNoted["zombie"]; ok {
		t.Fatal("僵尸会话 escalNoted 条目未清扫")
	}
	if _, ok := h.w.attempts["zombie"]; ok {
		t.Fatal("僵尸会话 attempts 条目未清扫")
	}
	if h.w.counts["live"] != 3 {
		t.Fatalf("续租会话条目应保留（live counts=%d，期望 3）", h.w.counts["live"])
	}
}

// G-04④ Escalated 续停滞观测状态去重：升级后逐拍 sweep 至多一条 escalated-stall-note
// （不再每轮重发）；恢复→再停滞→再升级 = 新幕，去重态重置后可再报。
func TestEscalatedNoteEmittedOncePerEpisode(t *testing.T) {
	h := newHarness(t, time.Minute, 10*time.Minute)
	h.w.SelfHeal = func(string) error { return nil }
	h.w.Diagnose = func(opslog.SessionHeartbeat) Diagnosis { return Diagnosis{Kind: "human"} }
	if err := h.w.Beat(Heartbeat{SessionID: "s1"}); err != nil {
		t.Fatal(err)
	}
	h.advance(11 * time.Minute)
	h.w.SweepOnce() // 判定+注入#1（attempts=1）
	h.advance(2 * time.Minute)
	h.w.SweepOnce() // 分诊（人工类）rounds=1
	h.advance(2 * time.Minute)
	h.w.SweepOnce() // rounds=2 ≥ HealLimit → Escalate
	if !h.emitted("Escalate") {
		t.Fatal("前置升级未发生")
	}
	h.advance(2 * time.Minute)
	h.w.SweepOnce() // 升级幕第一条续停滞观测
	h.advance(2 * time.Minute)
	h.w.SweepOnce() // 去重：不重发
	h.advance(2 * time.Minute)
	h.w.SweepOnce() // 去重：不重发
	if n := h.count(TypeLongTaskRunning); n != 1 {
		t.Fatalf("Escalated 续停滞观测应每幕至多一条，得 %d", n)
	}
	// 恢复（注入后心跳重现）→ 再停滞 → 再升级：新幕可再报。
	if err := h.w.Beat(Heartbeat{SessionID: "s1", Phase: "phase-advance"}); err != nil {
		t.Fatal(err)
	} // attempts>0 → healed（escalNoted 重置）
	h.advance(11 * time.Minute)
	h.w.SweepOnce() // 再判停滞+注入
	h.advance(2 * time.Minute)
	h.w.SweepOnce() // rounds=1
	h.advance(2 * time.Minute)
	h.w.SweepOnce() // rounds=2 → 再 Escalate（新幕：escalNoted 已清）
	h.advance(2 * time.Minute)
	h.w.SweepOnce() // 第二幕续停滞观测
	if n := h.count(TypeLongTaskRunning); n != 2 {
		t.Fatalf("恢复重置后新升级幕应可再报（累计=%d，期望 2）", n)
	}
	hb, _ := h.reg.get("s1")
	if hb.State != opslog.SessionStateEscalated {
		t.Fatalf("第二幕升级后行态=%s", hb.State)
	}
}
