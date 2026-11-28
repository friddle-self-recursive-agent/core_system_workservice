// approval.go —— 自审批真判执行面（2026-11-28 用户裁定「自审批未接线」修复 / workservice #4）。
//
// 旧行为：spec.approval.mode=agent 与 auto 同词——闸锚盲放行，无判决、无 DocumentDiff、
// 无风险回退。新行为：mode=agent 落 AwaitingApproval 即挂起（suspendWhy=self-approve），
// 锁外派发 ops_core_inner_agent 真判决（按 autoCriteria）；产物 DocumentDiff
// inneragent-approval-run<N> 同名幂等；通过→ApprovalGranted 沿声明表续行，驳回→
// ApprovalRejected 终态；风险超 maxRiskLevel（缺省 low）/判决失败→保持挂起回退人审
// （SuspendWhy 清空+人审判决文档+ApprovalNeeded 上报）。auto 仍盲放行；human 行为不变。
//
// 零翻转口径：RunOptions.ApprovalJudge 未挂接（nil）=旧词形（agent 仍自动放行），
// 装配层显式挂接才生效——与 VerifyGate/ServiceSpawns/InnerAgent 同范式。
package workflow

import (
	"fmt"
	"time"
)

const (
	// SuspendSelfApprove 挂起因由词（RunRecord.SuspendWhy）——自审批判决小 agent 执行窗口。
	SuspendSelfApprove = "self-approve"
	// EdgeApprovalJudge 判决派发轨迹行（history on——挂起同拍留痕，新载荷键=代际判别器）。
	EdgeApprovalJudge = "approval-judge"
	// EdgeApprovalEscalate 超险/判决失败回退人审轨迹行。
	EdgeApprovalEscalate = "approval-escalate"
)

// ApprovalVerdict 自审批判决结果（实现侧=ops_core_inner_agent 直调 Provider 解析 JSON 产出；
// 引擎只消费机器词形，不做即兴解释）。
type ApprovalVerdict struct {
	Granted bool   // 依 autoCriteria 判定是否自动批准
	Risk    string // low|medium|high（不可解析=high 失败收敛）
	Reason  string // 人话理由（落 DocumentDiff verdict）
	DocName string // 判决 DocumentDiff 名（inneragent-approval-run<N>——观测锚）
}

// ApprovalJudge 自审批判决契约（实现侧异步执行——引擎只派发不等待；同名幂等/产物落库
// 归实现侧自管）。RunRecord.ApprovalPolicy 随 run 快照传入，零文档库 IO 依赖。
type ApprovalJudge interface {
	JudgeApproval(run RunRecord, policy ApprovalPolicy, objKey string) (ApprovalVerdict, error)
}

// riskOrdinal 风险序（low<medium<high；未知/空=high——判决词形不可解析时失败收敛回人审）。
func riskOrdinal(risk string) int {
	switch risk {
	case "low":
		return 0
	case "medium":
		return 1
	case "high":
		return 2
	default:
		return 2
	}
}

// riskExceeds 判决风险是否超策略上限（上限缺省 low——scheme approvalGate 强制默认）。
func riskExceeds(risk, max string) bool {
	if max == "" {
		max = "low"
	}
	return riskOrdinal(risk) > riskOrdinal(max)
}

// RiskExceeds 导出判定面（装配层/判决实现侧共用同一风险序口径，零双源）。
func RiskExceeds(risk, max string) bool { return riskExceeds(risk, max) }

// selfApproveMode 该 run 是否走真判决通路（审批面挂接 judge ∧ mode=agent）。
func (r *RunLog) selfApproveMode(mode string) bool {
	return mode == "agent" && r.opts.ApprovalJudge != nil
}

// gateAutoPass 闸锚是否盲放行（auto 恒放行；agent 仅 judge 缺席时保留旧词形零翻转）。
func (r *RunLog) gateAutoPass(mode string) bool {
	return mode == "auto" || (mode == "agent" && r.opts.ApprovalJudge == nil)
}

// beginSelfApproveLocked 闸锚真判决窗口开启（调用方持 r.mu）：挂起细分词形=self-approve、
// 落 approval-judge 轨迹行、登记在途防重；返回派发请求（出锁后 dispatchSpawnReq）。
// 已在途/非闸锚挂起=返回 nil（幂等防御——重启/重复 ApprovalNeeded 不重复调 LLM）。
func (r *RunLog) beginSelfApproveLocked(run *RunRecord) *spawnReq {
	if Phase(run.Phase) != AwaitingApproval || run.Status != RunSuspended {
		return nil
	}
	if r.selfApproveInFlight == nil {
		r.selfApproveInFlight = map[int64]bool{}
	}
	if r.selfApproveInFlight[run.Run] {
		return nil
	}
	r.selfApproveInFlight[run.Run] = true
	run.SuspendWhy = SuspendSelfApprove
	run.History = append(run.History, HistoryRow{From: run.Phase, To: run.Phase, On: EdgeApprovalJudge, At: stampNano(time.Now())})
	return &spawnReq{kind: spawnKindApprovalJudge, run: *run}
}

// ApprovalVerdictDone 判决回执入口（实现侧锁外调用）：命中 self-approve 挂起窗口才落迁移。
// 通过→ApprovalGranted 声明边续行；驳回→ApprovalRejected 声明边终态；风险超限/判决缺席
// →回退人审（SuspendWhy 清空、ApprovalNeeded 上报、人审判决文档派发）。返回是否消费。
func (r *RunLog) ApprovalVerdictDone(runID int64, v ApprovalVerdict) bool {
	r.mu.Lock()
	idx := r.indexOfRunLocked(runID)
	if idx < 0 {
		r.mu.Unlock()
		return false
	}
	run := &r.ring[idx]
	if run.Status != RunSuspended || run.SuspendWhy != SuspendSelfApprove || Phase(run.Phase) != AwaitingApproval {
		r.mu.Unlock()
		return false // 已被人批/驳回/重启=幂等 noop
	}
	delete(r.selfApproveInFlight, runID)
	var fallback *spawnReq
	if riskExceeds(v.Risk, run.ApprovalPolicy.MaxRiskLevel) {
		// 超险/不可解析：保持挂起转人工（审批提升通路不变），生成人审 DocumentDiff。
		run.SuspendWhy = "" // 人审挂起（缺省词形）
		run.History = append(run.History, HistoryRow{From: run.Phase, To: run.Phase, On: EdgeApprovalEscalate, At: stampNano(time.Now())})
		if r.opts.DiffSummarizer != nil {
			fallback = &spawnReq{kind: spawnKindDiffDoc, run: *run}
		}
	} else if v.Granted {
		if next, ok := stepByTable(run.Transitions, run.Phase, SignalResume); ok {
			histAppend(run, run.Phase, next, SignalResume)
			run.Phase = next
			run.Status = RunRunning
			if Phase(next) == AwaitingApproval {
				if r.selfApproveMode(run.ApprovalMode) {
					if sp := r.beginSelfApproveLocked(run); sp != nil {
						fallback = sp // 连续闸锚：下一段再判
					}
				} else {
					run.Status = RunSuspended
				}
			}
		} else {
			// 声明表缺批准边=零动作防御（不即兴越迁），保持挂起回退人审。
			run.SuspendWhy = ""
			fallback = &spawnReq{kind: spawnKindDiffDoc, run: *run}
		}
	} else {
		if to, ok := stepByTable(run.Transitions, run.Phase, SignalReject); ok {
			histAppend(run, run.Phase, to, SignalReject)
			run.Phase, run.Status = to, RunFailed
		} else {
			run.SuspendWhy = "" // 缺驳回边=保持挂起转人工（不冒充终态）
			fallback = &spawnReq{kind: spawnKindDiffDoc, run: *run}
		}
	}
	if r.opts.Store != nil {
		persistAppend("run 自审批判决", r.opts.Store.AppendRun(runToRow(*run)))
	}
	r.selfApproved++
	r.mu.Unlock()
	if fallback != nil {
		r.dispatchSpawnReq(*fallback)
	}
	return true
}

// dispatchApprovalJudge 自审批判决派发（锁外 goroutine——LLM 阻塞不持 run 锁；panic 屏蔽
// 防实现侧崩主环）。判决完成回执 ApprovalVerdictDone；失败/异常=保持挂起并回退人审，
// 由 JudgeSuspendedApprovals 或人介入重试，不编造批准。
func (r *RunLog) dispatchApprovalJudge(sp spawnReq) {
	if r.opts.ApprovalJudge == nil {
		return
	}
	go func(run RunRecord) {
		defer func() { _ = recover() }()
		v, err := r.opts.ApprovalJudge.JudgeApproval(run, run.ApprovalPolicy, run.Involved)
		if err != nil || v.Risk == "" {
			// 失败收敛：风险词形缺席=high（riskExceeds 未知词按 high），显式赋值走回退人审。
			v = ApprovalVerdict{Risk: "high", Reason: fmt.Sprintf("（判决失败——%v）", err), DocName: v.DocName}
		}
		if r.opts.OnApprovalVerdict != nil {
			r.opts.OnApprovalVerdict(run, v) // 审计事件发布（锁外——装配层走 eventpub/bus）
		}
		r.ApprovalVerdictDone(run.Run, v)
	}(sp.run)
}

// JudgeSuspendedApprovals 存量挂起 run 批量判决清账（issue #4 的 85 个挂起 run 通路）：
// 扫环内 agent 模式 ∧ 闸锚挂起（含旧 self-approve 在途重放），按上限逐条派发判决。
// 返回派发数/跳过数。仅派发入口，判定/迁移仍走 ApprovalVerdictDone（同一落点无双源）。
func (r *RunLog) JudgeSuspendedApprovals(limit int) (dispatched, skipped int) {
	r.mu.Lock()
	var q []spawnReq
	for i := len(r.ring) - 1; i >= 0 && (limit <= 0 || len(q) < limit); i-- {
		run := &r.ring[i]
		if run.Status != RunSuspended || Phase(run.Phase) != AwaitingApproval || run.ApprovalMode != "agent" {
			continue
		}
		if !r.selfApproveMode(run.ApprovalMode) {
			skipped++
			continue
		}
		if sp := r.beginSelfApproveLocked(run); sp != nil {
			if r.opts.Store != nil {
				persistAppend("run 批量判决派发", r.opts.Store.AppendRun(runToRow(*run)))
			}
			q = append(q, *sp)
		} else {
			skipped++
		}
	}
	r.mu.Unlock()
	for _, sp := range q {
		r.dispatchApprovalJudge(sp)
	}
	return len(q), skipped
}
