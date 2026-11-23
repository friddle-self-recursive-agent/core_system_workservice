// notifygate.go —— 通知阶段门（KI-001 P0 通知段修复，v0.35）：
// 通知只由**相位终态事实**触发，apply 写入等进度事实不触发用户通知。
//
// 根因口径（runtime/known-issues.md KI-001）：通知触发此前只看事件类型+等级+路由命中，
// 不看相位是否完成——apply 成功（对象写入）被当作可通知的进度事件。判别器（KI-001 §6）：
// 未完成相位的 apply 通知计数=0；终态=1；同 dedupKey 窗口内 N 次注入实发 1（dedup 面，
// deliver.go）。
//
// 门的两轴（判定器纯函数，注入点=DeliveryOptions.PhaseGate——送达面零 RunLog 耦合，
// 装配层经 NotifyPhaseGate 接线）：
//  1. **人注意终态族**（事件轴）：ApprovalNeeded（审批提升=相位机停在闸锚等人）、
//     Escalate（升级=相位处置权移交）、HeartbeatLost（心跳丢失=agent 现势终态故障）——
//     三型是 builtin dispatch 表通知族权威行（scheme.yaml dispatch），天然相位终态/人注意
//     事实，恒放行（builtin 本体不动——闸不稀释既有通知路由语义）；
//  2. **run 相位终态背书**（run 轴）：其余事件（DocApplied 等进度事实）放行须 involved
//     owner 的现势 run 相位已达终态——序列尾（Report 相位完成=推进拍落地）或终点锚
//     （failed=RunFailed 迁移终点语义）。无 run/相位未投影/中途/挂起=无终态证据，一律拦
//     （fail-closed：无证据不通知——apply ≠ 阶段完成）。
//
// 拦阻不丢弃语义：被门事件只是不进用户通知面（送达投影零受理），事件本体照常留痕/归档/
// 推进 run（触发主干零旁路）；拦阻计数 gated 可查询（送达载荷新键=代际判别器）。
package workflow

// NotifyGate 通知阶段门谓词（true=放行进送达面）。实现须**非阻塞**（hub 在订阅路径上锁内
// 同步调用）且**不得回调 hub/delivery sink**（纯判定）。
type NotifyGate = func(rec TriggerRecord) bool

// notifyTerminalTypes 人注意终态族（事件轴白名单——builtin dispatch 表通知族三型，
// 判定依据见包注释）。词汇与 scheme.yaml eventTypes/dispatch 权威行同源。
var notifyTerminalTypes = map[string]bool{
	"ApprovalNeeded": true, // 审批提升：相位机停在闸锚等人（run 域挂起面同源词汇）
	"Escalate":       true, // 升级：处置权移交（onEscalate 策略面同源词汇）
	"HeartbeatLost":  true, // 心跳丢失：agent 现势终态故障（error 级人注意事实）
}

// NotifyTerminalType 事件类型是否人注意终态族（判定器事件轴的机读投影）。
func NotifyTerminalType(typ string) bool { return notifyTerminalTypes[typ] }

// PhaseTerminal run 相位终态判定（run 轴判别器）：
//   - status=failed：终点锚（RunFailed 即迁移终点语义，run.go 落点口径）——终态；
//   - status=running ∧ phase=声明序列尾：推进拍走满声明序列（Report 相位完成）——终态；
//   - 其余（triggered 无相位投影/suspended 闸锚挂起/running 中途）：无终态证据——非终态
//     （挂起面等人事实已由 ApprovalNeeded 事件轴承载，run 轴不重复放行）。
func (rec RunRecord) PhaseTerminal() bool {
	if rec.Status == RunFailed {
		return true
	}
	if rec.Status == RunRunning && len(rec.Phases) > 0 && rec.Phase != "" &&
		rec.Phase == rec.Phases[len(rec.Phases)-1] {
		return true
	}
	return false
}

// TerminalRunFor owner 的现势 run（环内该 owner 最新一条）与其相位终态判定。
// 环内无该 owner 的 run → ok=false（无 run 证据面）。扫描旧→新取末条命中=现势 run
// （与 ResumeSignal「锚最新 run=现势 run」同一口径，维度换成 owner）。
// 须在锁内调用（RunLog.mu）——导出包装 TerminalRunOf 加锁后调用本方法。
func (r *RunLog) terminalRunForLocked(owner string) (RunRecord, bool) {
	for i := len(r.ring) - 1; i >= 0; i-- {
		if r.ring[i].Owner == owner {
			return r.ring[i], true
		}
	}
	return RunRecord{}, false
}

// TerminalRunOf owner 的现势 run 相位终态判定（并发安全查询包装）。
// owner 解析口径=ownerOf（involved scope 根 systems/<s>→<s>）——与 run 落点留痕的 Owner
// 字段同源（rec.Owner=命中行谓词的解析值）。
func (r *RunLog) TerminalRunOf(owner string) (RunRecord, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.terminalRunForLocked(owner)
}

// NotifyPhaseGate 装配层缺省阶段门（两轴判定器，见包注释）：人注意终态族放行；其余事件查
// involved owner 现势 run 相位终态背书。runs=nil=无 run 证据面 → 非终态族一律拦
// （fail-closed——只读装配/无 run 面部署下进度事实零通知，与 KI-001 根因 4 同口径）。
// 返回值直接注入 DeliveryOptions.PhaseGate；nil 参数返回的谓词对终态族仍放行
// （人注意事实不因 run 面缺席而静默——HeartbeatLost 在无 run 面部署同样要送到人）。
func NotifyPhaseGate(runs *RunLog) NotifyGate {
	return func(rec TriggerRecord) bool {
		if notifyTerminalTypes[rec.Type] {
			return true
		}
		if runs == nil {
			return false
		}
		rr, ok := runs.TerminalRunOf(ownerOf(rec.Involved))
		return ok && rr.PhaseTerminal()
	}
}

// 事件等级词汇与序（event.md levels：debug<info<warn<error<fatal）。
var levelRank = map[string]int{
	"debug": 1, "info": 2, "warn": 3, "error": 4, "fatal": 5,
}

// LevelAtLeast 事件等级是否达到下限（minLevel 空=不过滤恒真；未知等级=不冒充达标
// ——词外面零放行，与 event.md 等级词汇表同源）。
func LevelAtLeast(level, minLevel string) bool {
	if minLevel == "" {
		return true
	}
	lv, ok1 := levelRank[level]
	mv, ok2 := levelRank[minLevel]
	return ok1 && ok2 && lv >= mv
}
