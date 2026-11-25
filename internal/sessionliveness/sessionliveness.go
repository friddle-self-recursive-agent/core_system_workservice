// Package sessionliveness 会话活性面（KI-003，event.md 规则 14/14b；agent-session.md §7）：
// 心跳注册（ops DB session_heartbeat 行+进程内计数）与装配层 watchdog（停滞判定→自愈固定
// 顺序→分诊上报→合成事件进总线）。引擎内部源第六段：生产者=装配层 watchdog（patrol.Ticker
// 同款装配范式——零新 HTTP face 的定时器面；POST /v1/sessions/heartbeat 是会话侧心跳注册面，
// 与判定面分立）。与注入面同一条总线主干零旁路（窗口去重/订阅派发/event_log 归档），差异四处：
// source 服务侧判定恒 core（巡检/阈值桥同口径——引擎自身产生的活性判定是审计事实）；
// 类型注册闸门前置（SessionStalled 未注册即不启动 watchdog——注册表是引擎内部源唯一开关）；
// level=注册表缺省投影（SessionStalled→error、SessionHealed/LongTaskRunning→info）；
// 缺省关断（--session-watch-interval=0 零事件零扫描，缺省不改既有行为面）。
//
// 词汇分立（event.md 规则 14b）：判定态词=SuspectedStall（本包 StateSuspectedStall、查询面
// state=suspected-stall），事件类型词=SessionStalled（进总线的一等事件，注册表 dedup=
// session:<id>）；恢复事件=SessionHealed；长任务显式汇报=LongTaskRunning。判定与处置每一步
// 都落心跳行（状态词）+合成事件（事件词），可查询可回指；处置失败继续上报绝不静默重试
// （发布错误逐条 stderr 告警——不中断 watchdog，下一拍重扫）。
package sessionliveness

import (
	"context"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// 事件类型与来源常量（与 scheme.yaml eventTypes 行同源；event.md 规则 14）。
const (
	TypeSessionStalled  = "SessionStalled"  // 事件词（判定态=SuspectedStall 的对外投影）
	TypeSessionHealed   = "SessionHealed"   // 恢复事实
	TypeLongTaskRunning = "LongTaskRunning" // 长任务显式汇报
	TypeApprovalNeeded  = "ApprovalNeeded"  // 权限分诊路（复用既有注册类型，不新增）
	SourceCore          = "core"            // 引擎内部源服务侧判定恒 core
)

// 停滞判定缺省阈值（event.md 规则 14②：缺省 10m——agent-session.md §7 同源声明）。
const DefaultStallThreshold = 10 * time.Minute

// DefaultHealRetryLimit 同会话连续自愈无效升级阈值（event.md 规则 14⑤：缺省 2）。
const DefaultHealRetryLimit = 2

// Heartbeat 一次心跳注册请求（会话推进点上报——相位推进/工具调用完成/产物落盘任一）。
type Heartbeat struct {
	SessionID  string // 必填：会话锚
	Phase      string // 推进点词汇（phase-advance|tool-call|artifact-write|inject…）
	LastAction string // 动作摘要（有界脱敏——不携带凭据/内网信息）
	TaskName   string // 长任务名（空=非长任务声明）
	LongTask   bool   // 声明为长任务（有进展心跳=不判停滞）
}

// SelfHeal 自愈动作执行口（装配层注入——固定顺序第一步的真实执行面）。
// 注入成功后由会话侧自行推进心跳（恢复=心跳重现）；返回错误=处置失败（继续上报不静默）。
type SelfHeal func(sessionID string) error

// Diagnosis 诊断卡因结论（诊断面输出——watchdog 分诊输入）。
type Diagnosis struct {
	Kind           string   // "permission"（权限类）|"human"（人工类）|"longtask"|"none"（无证据=默认人工上报路）
	Detail         string   // 现场摘要（最后工具调用/日志/等待对象/权限面——有界脱敏）
	NotifyChannels []string // bind 用户已配置通道（空=未配置任何通道——不 notify 仅页面展示）
}

// Diagnose 诊断口（装配层注入；nil=无诊断证据面 → 分诊默认人工上报路）。
type Diagnose func(h opslog.SessionHeartbeat) Diagnosis

// Registry 心跳账面口（opslog.SQLiteStore 满足；测试可注假实现）。
type Registry interface {
	UpsertSessionHeartbeat(h opslog.SessionHeartbeat) error
	StaleSessionHeartbeats(before time.Time, limit int) ([]opslog.SessionHeartbeat, error)
	AllSessionHeartbeats(limit int) ([]opslog.SessionHeartbeat, error)
}

// SessionScope 会话锚的 involved 投影（kind/session-id 三段同构惯例；scope 段=cored 会话域词）。
func SessionScope(sessionID string) string { return "Session/cored/" + sessionID }

// Watchdog 装配层 watchdog（cored 持有；零值不可用——Bus/Sch/Registry 必配）。
type Watchdog struct {
	Registry     Registry
	Bus          *eventbus.Bus
	Sch          *scheme.Scheme
	Interval0    time.Duration                        // 扫描拍间隔（--session-watch-interval；0=关断；导出经 Interval()）
	ThresholdV   time.Duration                        // 停滞阈值（--stall-threshold；缺省 DefaultStallThreshold）
	HealLimitV   int                                  // 连续自愈无效升级阈值（--heal-retry-limit；缺省 2）
	SelfHeal     SelfHeal                             // 自愈第一动作执行面（nil=只判定不上报处置——纯观测装配）
	Diagnose     Diagnose                             // 诊断面（nil=分诊默认人工上报路）
	NotifyWindow time.Duration                        // 人工上报去重窗口（缺省=Threshold——同停滞窗至多一封）
	Now          func() time.Time                     // 测试时钟注入（缺省 time.Now）
	onEvent      func(typ, session string, err error) // 装配锚回调（stderr 报告；测试断言）
	mu           sync.Mutex
	counts       map[string]int  // 进程内心跳计数（session → 累计心跳数）
	attempts     map[string]int  // 注入计数（session → 累计注入次数；行 HealAttempts 快照源）
	rounds       map[string]int  // 人工分诊上报轮次（session → 连续无效轮数；达 HealLimit 升级）
	prevStalled  map[string]bool // 曾判停滞集（session → true；无注入面部署的恢复判定锚）
}

// New 构建 watchdog（threshold<=0=缺省 10m；healLimit<=0=缺省 2）。
func New(reg Registry, bus *eventbus.Bus, sch *scheme.Scheme, interval, threshold time.Duration, healLimit int) *Watchdog {
	if threshold <= 0 {
		threshold = DefaultStallThreshold
	}
	if healLimit <= 0 {
		healLimit = DefaultHealRetryLimit
	}
	return &Watchdog{
		Registry: reg, Bus: bus, Sch: sch,
		Interval0: interval, ThresholdV: threshold, HealLimitV: healLimit,
		NotifyWindow: threshold,
		Now:          time.Now,
		counts:       map[string]int{},
		attempts:     map[string]int{},
		rounds:       map[string]int{},
		prevStalled:  map[string]bool{},
	}
}

// OnEvent 装配锚回调（每条合成事件发射后调用；err=nil=已入总线）。
func (w *Watchdog) OnEvent(f func(typ, session string, err error)) {
	if w == nil {
		return
	}
	w.onEvent = f
}

// Enabled watchdog 是否开启：interval>0 ∧ 类型已注册（注册表闸门前置——未注册即不启动）。
func (w *Watchdog) Enabled() bool {
	if w == nil || w.Interval0 <= 0 || w.Bus == nil || w.Sch == nil || w.Registry == nil {
		return false
	}
	_, err := w.Sch.LookupEventType(TypeSessionStalled)
	return err == nil
}

// Interval 扫描拍间隔（查询面 watchdog 投影用；未构建=零值）。
func (w *Watchdog) Interval() time.Duration {
	if w == nil {
		return 0
	}
	return w.Interval0
}

// Threshold 停滞阈值（查询面 watchdog 投影用）。
func (w *Watchdog) Threshold() time.Duration {
	if w == nil {
		return 0
	}
	return w.ThresholdV
}

// HealLimit 连续自愈无效升级阈值（查询面 watchdog 投影用）。
func (w *Watchdog) HealLimit() int {
	if w == nil {
		return 0
	}
	return w.HealLimitV
}

// Beat 注册一次心跳（会话推进点上报）：进程内计数 +1 ∧ ops 行 upsert（state=live 或
// long-task 按声明；heal_attempts=行上现值——心跳不重置自愈计数，只恢复活性）。这是活性
// 租约续租——watchdog 扫描以行的 updated_at 为准。
func (w *Watchdog) Beat(h Heartbeat) error {
	if w == nil || w.Registry == nil {
		return fmt.Errorf("sessionliveness: 心跳账面未装配")
	}
	if h.SessionID == "" {
		return fmt.Errorf("sessionliveness: 心跳缺 session_id（会话锚必填）")
	}
	w.mu.Lock()
	w.counts[h.SessionID]++
	attempts := w.attempts[h.SessionID]
	w.mu.Unlock()
	now := w.now()
	state := opslog.SessionStateLive
	if h.LongTask {
		state = opslog.SessionStateLongTask
	}
	hb := opslog.SessionHeartbeat{
		SessionID: h.SessionID, Ts: now, Phase: clip(h.Phase, 60), LastAction: clip(h.LastAction, 200),
		TaskName: clip(h.TaskName, 120), LongTask: h.LongTask, State: state, UpdatedAt: now,
		HealAttempts: attempts,
	}
	if err := w.Registry.UpsertSessionHeartbeat(hb); err != nil {
		return err
	}
	if attempts > 0 {
		// 注入后的心跳重现=自愈恢复（活性租约续租即恢复确认）——恢复记 SessionHealed
		// （携停滞时长与最后心跳点；连败计数清零）。
		stalled := w.lastStallAt(h.SessionID, now) // 判定瞬刻→本次心跳=停滞时长观测
		w.mu.Lock()
		delete(w.attempts, h.SessionID)
		delete(w.rounds, h.SessionID)
		w.mu.Unlock()
		hb.HealedAt = now
		hb.HealAttempts = 0
		hb.State = opslog.SessionStateHealed
		_ = w.Registry.UpsertSessionHeartbeat(hb) // 行态快照并入恢复戳（state=healed）
		w.emit(w.buildEvent(TypeSessionHealed, hb, stalled), h.SessionID, "healed-by-beat")
	} else if prev, ok := w.prevStalled[h.SessionID]; ok && prev {
		// 无注入执行面部署的恢复判定：曾判停滞（本进程判定过 SuspectedStall）∧注入前
		// 恢复（自行续租）——同记 SessionHealed（恢复=活性事实；判别字段=healAttempts=0，
		// 语义=无自愈动作注入、会话自行恢复——诚实投影不冒充注入恢复）。
		stalled := w.lastStallAt(h.SessionID, now)
		w.mu.Lock()
		delete(w.prevStalled, h.SessionID)
		w.mu.Unlock()
		hb.HealedAt = now
		hb.State = opslog.SessionStateHealed
		_ = w.Registry.UpsertSessionHeartbeat(hb) // 行态快照并入恢复戳（state=healed）
		w.emit(w.buildEvent(TypeSessionHealed, hb, stalled), h.SessionID, "healed-self")
	}
	return nil
}

// lastStallAt 停滞时长观测（判定瞬刻=行进入停滞前最后一次租约时刻；简化=本次心跳-阈值——
// 恢复最晚发生在上一拍之后；观测面用，非权威时长）。
func (w *Watchdog) lastStallAt(sessionID string, now time.Time) time.Duration {
	return w.ThresholdV
}

// Run watchdog 主循环：每 Interval 扫描一次停滞行（ctx 取消即退场，幂等可重启）。
// 未开启（Enabled=false）即返——关断=零事件零扫描（机械判别器的关断面）。
func (w *Watchdog) Run(ctx context.Context) {
	if !w.Enabled() {
		return
	}
	tk := time.NewTicker(w.Interval0)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			w.sweep() // 发布/处置失败不中断 watchdog（可丢语义，event.md §1——下一拍重扫续上报）
		}
	}
}

// SweepOnce 单次扫描（测试与探针手动驱动入口；生产走 Run 定时拍）。
func (w *Watchdog) SweepOnce() { w.sweep() }

func (w *Watchdog) sweep() {
	now := w.now()
	// 停滞候选=updated_at 早于 now-threshold 的行（含 live 长期未续租与已 suspected 行）。
	stale, err := w.Registry.StaleSessionHeartbeats(now.Add(-w.ThresholdV), 0)
	if err != nil {
		log.Printf("sessionliveness: 停滞扫描失败（下一拍重试）: %v", err)
		return
	}
	// 诊断面只看尚处停滞的行；healed 行 updated_at 已续租自然出池。
	for _, h := range stale {
		w.handleStale(h, now)
	}
}

// handleStale 一行停滞候选的处置（固定顺序状态机：判定→注入继续→恢复/诊断分诊→升级）。
func (w *Watchdog) handleStale(h opslog.SessionHeartbeat, now time.Time) {
	stalledFor := now.Sub(h.UpdatedAt)
	// 长任务：声明为长任务（或有进展心跳）不判停滞——只显式汇报（LongTaskRunning）。
	if h.LongTask {
		w.emitLongTask(h, stalledFor)
		return
	}
	switch h.State {
	case opslog.SessionStateLive:
		// 第一拍判定：live→suspected-stall，合成 SessionStalled（事件词；判定态=SuspectedStall）。
		w.markStalled(h, stalledFor)
	case opslog.SessionStateSuspectedStall:
		// 已判定未恢复：按固定顺序处置。
		w.treat(h, stalledFor, now)
	case opslog.SessionStateHealed:
		// healed 行停滞=再次失活：回 live 口径重新判定（上一轮自愈成功后再度停滞是新因由）。
		w.markStalled(h, stalledFor)
	case opslog.SessionStateEscalated:
		// 已升级：停自动处置——继续上报（Escalate 已在账，不再重复升级；事件留痕不静默）。
		w.emitLongTaskStalledNote(h, stalledFor)
	}
}

// markStalled 判定停滞：行态→suspected-stall + 合成 SessionStalled（每步留事件）。
func (w *Watchdog) markStalled(h opslog.SessionHeartbeat, stalledFor time.Duration) {
	w.mu.Lock()
	w.prevStalled[h.SessionID] = true
	w.mu.Unlock()
	w.updateState(h.SessionID, opslog.SessionStateSuspectedStall, nil)
	evt := w.buildEvent(TypeSessionStalled, h, stalledFor)
	w.emit(evt, h.SessionID, "suspected-stall（注入继续前判定）")
	// 自愈第一动作：注入固定 prompt「继续」（保留上下文，不重启）——每步留事件；
	// 注入动作本身落心跳行（phase=inject，事件回执不落 DB——DB 行=租约面非事件账本）。
	if w.SelfHeal != nil {
		if err := w.SelfHeal(h.SessionID); err != nil {
			// 处置失败继续上报，绝不静默重试：失败上报事件 + stderr 告警；行态保持 suspected
			// （HealAttempts=0，下一拍 treat 重试第一动作——注入未成功留痕=重试，不是跳步）。
			log.Printf("sessionliveness: 会话 %s 注入「继续」失败（下一拍重试上报）: %v", h.SessionID, err)
			w.emit(w.buildDetailEvent(TypeSessionStalled, h, stalledFor,
				"自愈注入失败（继续上报）"), h.SessionID, "self-heal-inject-failed")
			return
		}
		w.recordInject(h.SessionID)
		w.updateState(h.SessionID, opslog.SessionStateSuspectedStall, func(hb *opslog.SessionHeartbeat) {
			hb.Phase = "inject-continue"
			hb.LastAction = clip("注入固定 prompt「继续」（自愈第一动作）", 200)
		})
	}
}

// treat 已判定未恢复行（固定顺序状态机的续步）：注入未成功留痕=重试第一动作
// （绝不静默——处置失败继续上报）；注入已留痕仍未恢复=诊断分诊→升级判定。
func (w *Watchdog) treat(h opslog.SessionHeartbeat, stalledFor time.Duration, now time.Time) {
	if h.HealAttempts == 0 {
		// 注入动作未留痕（上一拍注入失败）——重试第一动作，不是静默跳步。
		w.retryInject(h, stalledFor)
		return
	}
	// 恢复由 Beat 完成（行 updated_at 续租出池）——走到这里=注入后仍未恢复，诊断卡因。
	d := Diagnosis{Kind: "human"}
	if w.Diagnose != nil {
		d = w.Diagnose(h)
	}
	switch d.Kind {
	case "longtask":
		// 长任务（诊断面确认有进展心跳）：显式汇报不判停滞（结束回收=恢复出池）。
		w.emitLongTask(h, stalledFor)
		return
	case "permission":
		// 权限类：向上提交 ApprovalNeeded（走审批链+通知，不自行绕过）。
		w.emitPermission(h, stalledFor, d)
		return
	default:
		// 人工类：bind 用户通道上报（已配置通道走既有 notification 面——SessionStalled
		// cc 旁路已随判定事件送达；此处补发诊断细节事件）+ 未配置=页面展示零外发。
		w.emitHuman(h, stalledFor, d)
		// 连续自愈无效达阈值 → Escalate 停自动处置升级（每步留事件，不静默重试）。
		w.mu.Lock()
		w.rounds[h.SessionID]++
		rounds := w.rounds[h.SessionID]
		w.mu.Unlock()
		if rounds >= w.HealLimitV {
			w.escalate(h, stalledFor, rounds)
		}
	}
}

// retryInject 注入失败后的重试（处置失败继续上报——事件与 stderr 双面可见，绝不静默跳步）。
func (w *Watchdog) retryInject(h opslog.SessionHeartbeat, stalledFor time.Duration) {
	if w.SelfHeal == nil {
		return // 纯观测装配（无自愈执行面）——判定事件已在账，不冒充处置
	}
	if err := w.SelfHeal(h.SessionID); err != nil {
		log.Printf("sessionliveness: 会话 %s 注入「继续」失败（继续上报，下一拍重试）: %v", h.SessionID, err)
		w.emit(w.buildDetailEvent(TypeSessionStalled, h, stalledFor,
			"自愈注入失败（继续上报）"), h.SessionID, "self-heal-inject-failed")
		return
	}
	w.recordInject(h.SessionID)
	w.updateState(h.SessionID, opslog.SessionStateSuspectedStall, func(hb *opslog.SessionHeartbeat) {
		hb.Phase = "inject-continue"
		hb.LastAction = clip("注入固定 prompt「继续」（自愈第一动作）", 200)
	})
}

// escalate 升级：合成 Escalate（既有注册类型，走 root+notification 既有权威行）+行态 escalated。
func (w *Watchdog) escalate(h opslog.SessionHeartbeat, stalledFor time.Duration, attempts int) {
	et, err := w.Sch.LookupEventType("Escalate")
	if err != nil {
		log.Printf("sessionliveness: Escalate 未注册（升级事件缺席——缺口可见）: %v", err)
		return
	}
	evt := &eventbus.Event{
		Type: "Escalate", Source: SourceCore,
		Involved: SessionScope(h.SessionID),
		DedupKey: "session-escalate:" + h.SessionID,
		Level:    et.DefaultLevel,
	}
	w.emit(evt, h.SessionID, fmt.Sprintf("escalated（自愈无效×%d 停自动处置）", attempts))
	w.updateState(h.SessionID, opslog.SessionStateEscalated, nil)
}

// emitLongTask 长任务显式汇报（任务名/已运行时长/当前阶段/决策点；有进展=不判停滞；结束回收）。
func (w *Watchdog) emitLongTask(h opslog.SessionHeartbeat, running time.Duration) {
	evt := w.buildDetailEvent(TypeLongTaskRunning, h, running,
		fmt.Sprintf("长任务 %s 运行中 %.0f 分钟（有进展心跳=不判停滞；结束回收）", h.TaskName, running.Minutes()))
	w.emit(evt, h.SessionID, "long-task-report")
}

// emitLongTaskStalledNote escalated 行的续停滞观测（不静默——继续上报；不再重复升级）。
func (w *Watchdog) emitLongTaskStalledNote(h opslog.SessionHeartbeat, stalledFor time.Duration) {
	evt := w.buildDetailEvent(TypeLongTaskRunning, h, stalledFor,
		fmt.Sprintf("会话停滞 %.0f 分钟（已升级停自动处置——人工接管）", stalledFor.Minutes()))
	w.emit(evt, h.SessionID, "escalated-stall-note")
}

// emitPermission 权限分诊路：ApprovalNeeded（复用既有注册类型——走审批链+通知）。
func (w *Watchdog) emitPermission(h opslog.SessionHeartbeat, stalledFor time.Duration, d Diagnosis) {
	et, err := w.Sch.LookupEventType(TypeApprovalNeeded)
	if err != nil {
		log.Printf("sessionliveness: ApprovalNeeded 未注册（权限分诊事件缺席）: %v", err)
		return
	}
	evt := &eventbus.Event{
		Type: TypeApprovalNeeded, Source: SourceCore,
		Involved: SessionScope(h.SessionID),
		DedupKey: "session-permission:" + h.SessionID,
		Level:    et.DefaultLevel,
	}
	w.emit(evt, h.SessionID, "diagnosis=permission（向上提交，不自行绕过）")
}

// emitHuman 人工分诊路：bind 用户已配置通道经 SessionStalled cc 旁路送达（判定事件已投）；
// 此处落诊断细节（未配置任何通道=不 notify 仅页面展示——零外发，notification.md 规则 6）。
func (w *Watchdog) emitHuman(h opslog.SessionHeartbeat, stalledFor time.Duration, d Diagnosis) {
	note := "diagnosis=human（bind 通道上报）"
	if len(d.NotifyChannels) == 0 {
		note = "diagnosis=human（未配置通道=不 notify 仅页面展示，零外发）"
	}
	evt := w.buildDetailEvent(TypeSessionStalled, h, stalledFor, note+"；"+d.Detail)
	w.emit(evt, h.SessionID, note)
}

// buildEvent 合成停滞事件（event.md 规则 14：involved=session 锚、dedupKey=session:<id>、
// level=注册表缺省投影）。
func (w *Watchdog) buildEvent(typ string, h opslog.SessionHeartbeat, stalledFor time.Duration) *eventbus.Event {
	return w.buildDetailEvent(typ, h, stalledFor, "")
}

// buildDetailEvent 带现场摘要的合成事件（dedupKey 按类型分立：session:/session-heal:/longtask:）。
func (w *Watchdog) buildDetailEvent(typ string, h opslog.SessionHeartbeat, at time.Duration, detail string) *eventbus.Event {
	et, err := w.Sch.LookupEventType(typ)
	if err != nil {
		return nil // 闸门前置：未注册不发布（Enabled 已挡，防御性再断）
	}
	dk := "session:" + h.SessionID
	switch typ {
	case TypeSessionHealed:
		dk = "session-heal:" + h.SessionID
	case TypeLongTaskRunning:
		dk = "longtask:" + h.SessionID
	}
	_ = detail // detail 不进总线（event.md 规则 2：事件只带指针）——现场在 ops 行可查询
	return &eventbus.Event{
		Type: typ, Source: SourceCore, Involved: SessionScope(h.SessionID),
		DedupKey: dk, Level: et.DefaultLevel,
	}
}

// emit 发布合成事件（可丢语义；发布错误经装配锚上报不静默）。
func (w *Watchdog) emit(e *eventbus.Event, session, note string) {
	if e == nil {
		return
	}
	err := w.Bus.Publish(e)
	w.onEventp(e.Type, session, err, note)
}

func (w *Watchdog) onEventp(typ, session string, err error, note string) {
	if w.onEvent != nil {
		w.onEvent(typ, session, err)
	}
	_ = note
}

// Heal 看门狗侧恢复确认（诊断面/会话侧确认恢复后调用）：行态→healed + 合成 SessionHealed。
// 停滞时长=恢复时刻-行 updated_at（携停滞时长与最后心跳点——event.md 规则 14③）。
func (w *Watchdog) Heal(sessionID string) error {
	rows, err := w.Registry.AllSessionHeartbeats(0)
	if err != nil {
		return err
	}
	for _, h := range rows {
		if h.SessionID != sessionID {
			continue
		}
		now := w.now()
		stalledFor := now.Sub(h.UpdatedAt)
		w.updateState(sessionID, opslog.SessionStateHealed, func(hb *opslog.SessionHeartbeat) {
			hb.HealedAt = now
		})
		w.emit(w.buildDetailEvent(TypeSessionHealed, h, stalledFor, ""), sessionID, "healed")
		w.mu.Lock()
		delete(w.attempts, sessionID) // 成功自愈=连续无效计数清零
		delete(w.rounds, sessionID)
		delete(w.prevStalled, sessionID)
		w.mu.Unlock()
		return nil
	}
	return fmt.Errorf("sessionliveness: 会话 %s 无心跳行（无法确认恢复）", sessionID)
}

// Heartbeats 读面投影（GET /v1/sessions 底层；进程内计数并入——计数=观测面非权威）。
func (w *Watchdog) Heartbeats(limit int) ([]SessionView, error) {
	rows, err := w.Registry.AllSessionHeartbeats(limit)
	if err != nil {
		return nil, err
	}
	w.mu.Lock()
	defer w.mu.Unlock()
	out := make([]SessionView, 0, len(rows))
	for _, h := range rows {
		var healedAt *time.Time
		if !h.HealedAt.IsZero() {
			t := h.HealedAt
			healedAt = &t
		}
		out = append(out, SessionView{
			SessionID: h.SessionID, State: h.State, Phase: h.Phase, LastAction: h.LastAction,
			TaskName: h.TaskName, LongTask: h.LongTask, HealAttempts: h.HealAttempts,
			LastBeat: h.UpdatedAt, HealedAt: healedAt,
			Heartbeats: w.counts[h.SessionID],
			Involved:   SessionScope(h.SessionID),
		})
	}
	return out, nil
}

// SessionView 查询面投影行（state 词=DB 词形：suspected-stall 判定态词=SuspectedStall）。
type SessionView struct {
	SessionID    string     `json:"sessionId"`
	State        string     `json:"state"`
	Phase        string     `json:"phase"`
	LastAction   string     `json:"lastAction"`
	TaskName     string     `json:"taskName,omitempty"`
	LongTask     bool       `json:"longTask"`
	HealAttempts int        `json:"healAttempts"`
	Heartbeats   int        `json:"heartbeats"`
	Involved     string     `json:"involved"`
	LastBeat     time.Time  `json:"lastBeat"`
	HealedAt     *time.Time `json:"healedAt,omitempty"` // nil=未恢复过（零值时间不投影）
}

// updateState 行态更新（状态机落点；healed_at 随行）。
func (w *Watchdog) updateState(sessionID, state string, mutate func(*opslog.SessionHeartbeat)) {
	rows, err := w.Registry.AllSessionHeartbeats(0)
	if err != nil {
		log.Printf("sessionliveness: 行态更新读取失败（下一拍重试）: %v", err)
		return
	}
	for _, h := range rows {
		if h.SessionID != sessionID {
			continue
		}
		h.State = state
		// 状态迁移不续租（租约=活性推进；watchdog 自身的状态翻转不算心跳——否则停滞行
		// 永远逃出扫描池）。Ts/UpdatedAt 保持原值=停滞时长判定锚。
		w.mu.Lock()
		h.HealAttempts = w.attempts[sessionID] // 行快照并入注入计数（treat 分诊依据=行上现值）
		w.mu.Unlock()
		if mutate != nil {
			mutate(&h)
		}
		if err := w.Registry.UpsertSessionHeartbeat(h); err != nil {
			log.Printf("sessionliveness: 行态更新落盘失败（下一拍重试）: %v", err)
		}
		return
	}
}

// recordInject 注入计数（连续自愈无效判定输入）。
func (w *Watchdog) recordInject(sessionID string) {
	w.mu.Lock()
	w.attempts[sessionID]++
	w.mu.Unlock()
}

// attempts 计数为进程内状态（判定以内存为准）；行 HealAttempts=落行时刻快照（观测面）。

func (w *Watchdog) now() time.Time {
	if w.Now == nil {
		return time.Now()
	}
	return w.Now()
}

// clip 有界化（动作摘要/任务名/相位词——防载荷膨胀，event.md 规则 2 同口径）。
func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n]
}
