// advance.go —— 相位推进器异步推进首段（v0.26——装配层定时驱动，workflow.md 规则 15）。
// v0.23 受理即启动（同步首步随受理同拍完成）；v0.26 起引擎在受理之后**逐段承接** run 相位
// 推进：装配层定时器（Advancer，patrol.Ticker 同款装配范式——生产者=装配层定时器，零新
// HTTP face，与写面开关分立：只读装配亦可开启）每拍调用 RunLog.Advance()，对环内每个现势
// run 执行**单步迁移判定**——查表落点复用 stepByTable（on=advance；判定器 v0.25 已表驱动，
// 本段只加驱动源不动判定器）。
//
// 落点语义=迁移判定落点的通用规则（v0.24 闸锚挂起第一层判定的推广——StartRun/ResumeSignal
// 同款落点口径）：
//   - 落闸锚（AwaitingApproval）→ 挂起（status=suspended——相位机等待外部审批信号）；
//   - 落终点锚（Failed）→ 终态（status=failed——RunFailed 词汇即迁移终点语义，非引擎失败）；
//   - 序列尾（无 advance 行命中）→ 零动作停（status 保持 running——诚实投影不冒充执行完毕，
//     「run 出现 ≠ 对账执行完毕」口径不变）。
//
// 零动作防御面：相位未投影（未声明 phases 的 object 源——零投影口径不稀释）、triggered、
// suspended（挂起面不前进——审批信号之前相位机等待）、failed（终态不迁移）一律原样保留；
// 每拍每 run 至多一步（逐段承接，不跳步不并发推进同 run）。缺省关断（interval=0）：缺省不改既有行为面，推进按部署显式开启
// （受理首步同步推进不受关断影响——v0.23 口径恒在）。
//
// G-09/G-12 Verify 硬 gate（v0.67）：推进路径的新增闸——run 试图离开 Verifying 相位先过
// VerifyGate（connectops-code-audit 接线，装配层恒挂接）：findings 非空/unbound=卡 Verify
// 不放行 + Escalate 上报；审核执行两阶段（锁外）不阻塞订阅主干。gate 未挂接=既有行为面
// 零翻转（能力增量不破既有推进语义——装配面收口在 cored）。
package workflow

import (
	"context"
	"time"
)

// Advance 异步推进入口（线程安全——RunLog.mu 已在；每拍一次，环内每 run 至多单步）。
// 返回本拍推进的 run 数（累计值随 RunState.Advances 投影可查——查询面观测点）。
// v0.32 刹车首段（workflows 规则 17）：推进拍变更上限——刹车闸挂接时，每拍对每锚的 run
// 相位变更数受生效刹车值约束（变更单元=run 相位变更留痕；上限只计实际迁移，序列尾零动作
// 不占额）；超限 run 本拍零动作、下拍续推（level-triggered 不丢拍；capped 计数可见）。
// 未挂接=零刹车既有行为面零翻转（基线保护）。
// G-09/G-12 Verify 硬 gate：run 试图**离开 Verifying 相位**（Verifying→后继 advance 行命中）
// 先过 gate（RunOptions.VerifyGate；nil=既有行为面零翻转）。审核执行在锁外（checker 时长
// 不持 run 锁——订阅主干/查询面不被阻塞）：两阶段推进=①锁内扫描即判类（无 gate 路径照旧
// 即时迁移；Verifying 出行候选快照）→②锁外逐候选审核→③锁内复验（同 run∧同相位∧同状态
// 才落迁移——竞态放弃本拍）+判定留痕写穿透。OK=false=卡 Verify（相位/状态原样保留，gated
// 计数可见，判定投影随 run 留痕）+ OnVerifyBlocked 锁外上报（Escalate——cored 装配）。
func (r *RunLog) Advance() int {
	// 阶段①：锁内扫描分类（即时迁移路径照旧；Verifying 出行候选=快照待审；service 落点
	// 拉起请求锁内收集锁外派发——ServiceSpawner 非阻塞契约，serviceexec.go）。
	type gateCand struct {
		run RunRecord // 快照（锁外审核输入）
	}
	var cands []gateCand
	var touched []int     // 阶段①已即时迁移的环内下标（写穿透 upsert）
	var spawnQ []spawnReq // 阶段① service 落点拉起请求（出锁后派发——非阻塞契约）
	advanced := 0
	mutated := map[string]int{} // 本拍每锚已发生迁移数（刹车闸挂接时生效）
	r.mu.Lock()
	for i := range r.ring {
		run := &r.ring[i]
		if len(run.Phases) == 0 || run.Status != RunRunning {
			continue // 相位未投影（未声明 phases 的 object 源）/ 非进行中（triggered/suspended/failed）=零动作
		}
		if r.opts.Brakes.Enabled && run.Brakes.MaxMutationsTick > 0 && mutated[run.Workflow] >= run.Brakes.MaxMutationsTick {
			r.capped++ // 超限本拍零动作（下拍续推——level-triggered 不丢拍）
			continue
		}
		next, ok := stepByTable(run.Transitions, run.Phase, EdgeAdvance)
		if !ok {
			// F-04 正向末步收束（2026-11-26，workflow.md §1：Reporting 完成即 converged，
			// 绝不从 Reporting 经 advance 自动落 Failed）：序列尾（无 advance 行命中）且仍在
			// running 的 run 落 converged——phase 停驻正向末步原样，history 落收束行（零迁移
			// 计数——收束是状态词翻转非相位迁移，不占刹车变更额）。
			if Phase(run.Phase) != AwaitingApproval && (run.Phase != string(Failed)) {
				run.Status = RunConverged
				run.History = append(run.History, HistoryRow{From: run.Phase, To: run.Phase, On: "converged", At: stampNano(time.Now())})
				if r.opts.Store != nil {
					persistAppend("run 收束", r.opts.Store.AppendRun(runToRow(*run)))
				}
			}
			continue
		}
		if r.opts.VerifyGate != nil && Phase(run.Phase) == Verifying {
			// Verify 硬 gate 候选：离开 Verifying 相位先过审（审核锁外——不持 run 锁）。
			cands = append(cands, gateCand{run: *run})
			continue
		}
		if isp := r.innerGateLocked(run); isp != nil {
			spawnQ = append(spawnQ, *isp) // executor=inner 轻量执行单元（2026-11-28 润滑剂裁定）：当前相位停留执行——挂起+锁外派发
			continue
		}
		if sp := r.landAfterAdvanceLocked(run, next); sp != nil {
			spawnQ = append(spawnQ, *sp) // service 执行单元落点：挂起+拉起请求（出锁派发）
		}
		mutated[run.Workflow]++
		advanced++
		touched = append(touched, i)
	}
	// 写穿透（R-ENG §1）：本拍即时迁移过的 run 逐条 upsert（锁内同步与投影同序；失败只注记）。
	if r.opts.Store != nil {
		for _, i := range touched {
			persistAppend("run 推进", r.opts.Store.AppendRun(runToRow(r.ring[i])))
		}
	}
	r.advances += int64(advanced)
	r.mu.Unlock()
	// service 拉起派发（锁外——ServiceSpawner 非阻塞契约：投影在 SpawnFromRun 内自锁落环）；
	// 判决文档生成派发（锁外异步——DiffSummarizer 实现侧自管幂等，引擎不等待不重试）。
	for _, sp := range spawnQ {
		r.dispatchSpawnReq(sp)
	}

	// 阶段②：锁外审核（checker 真实执行——audit 只读幂等；每拍重审=修复后下拍即过）。
	type verdictCand struct {
		run RunRecord
		v   VerifyVerdict
	}
	var verdicts []verdictCand
	if len(cands) > 0 {
		for _, c := range cands {
			verdicts = append(verdicts, verdictCand{run: c.run, v: r.opts.VerifyGate.Verify(c.run)})
		}
	}

	// 阶段③：锁内复验落迁移（同 run∧同相位∧同状态才落——阶段②窗口内被审批信号/受理
	// 改写过的候选放弃本拍，下拍重审重判）+判定留痕；blocked 计数与回调清单出锁后上报。
	var blocked []verdictCand
	gatedPassed := 0
	var spawnQ3 []spawnReq // 阶段③ service 落点拉起请求（出锁后派发——与阶段①同一契约）
	if len(verdicts) > 0 {
		r.mu.Lock()
		for _, vc := range verdicts {
			idx := r.indexOfRunLocked(vc.run.Run)
			if idx < 0 || r.ring[idx].Phase != vc.run.Phase || r.ring[idx].Status != vc.run.Status {
				continue // 竞态放弃本拍（run 已被改写——诚实投影不覆盖他拍迁移）
			}
			run := &r.ring[idx]
			run.Verify = &vc.v // 判定随 run 留痕（放行/卡两词形都留——证据面）
			if !vc.v.OK {
				r.gated++ // 卡 Verify：相位/状态原样保留（本拍零迁移——硬 gate 不放行）
				if r.opts.Store != nil {
					persistAppend("run 卡 Verify 判定", r.opts.Store.AppendRun(runToRow(*run)))
				}
				blocked = append(blocked, vc)
				continue
			}
			if r.opts.Brakes.Enabled && run.Brakes.MaxMutationsTick > 0 && mutated[run.Workflow] >= run.Brakes.MaxMutationsTick {
				r.capped++ // 刹车超限：放行判定留痕、迁移让下拍（level-triggered 不丢拍）
				continue
			}
			next, ok := stepByTable(run.Transitions, run.Phase, EdgeAdvance)
			if !ok {
				continue // 防御：表行在阶段②窗口被改写致 advance 行缺席——零动作
			}
			if sp := r.landAfterAdvanceLocked(run, next); sp != nil {
				spawnQ3 = append(spawnQ3, *sp) // gate 放行落点又是 service 节点（Verifying→Deploying 链）——出锁派发
			}
			mutated[run.Workflow]++
			gatedPassed++
			if r.opts.Store != nil {
				persistAppend("run Verify 放行", r.opts.Store.AppendRun(runToRow(*run)))
			}
		}
		r.advances += int64(gatedPassed)
		r.mu.Unlock()
	}
	// 阶段③ service 拉起派发（锁外——与阶段①同一非阻塞契约）。
	for _, sp := range spawnQ3 {
		r.dispatchSpawnReq(sp)
	}
	// 锁外上报（Escalate 发布不持锁——回调侧可走总线/HTTP 面任意时长）。
	if r.opts.OnVerifyBlocked != nil {
		for _, vc := range blocked {
			r.opts.OnVerifyBlocked(vc.run, vc.v)
		}
	}
	return advanced + gatedPassed
}

// applyAdvanceLocked 单步迁移落环（调用方持 r.mu——阶段①/③共用同一落点语义：落闸锚挂起/
// 落侧轨终点锚 Failed 终态；其余相位 running 续行）。
// F-04（2026-11-26 状态机纠偏）：正向序列止于 Reporting（声明面已摘除线性尾 Failed）——
// 落正向末步本拍仍 running（迁移判定=相位步进）；收束=下一拍序列尾判定（Advance 序列尾
// 分支落 converged——一步一迁移口径不变，Reporting 完成即正常收束）。
func (r *RunLog) applyAdvanceLocked(run *RunRecord, next string) {
	at := stampNano(time.Now())
	run.History = append(run.History, HistoryRow{From: run.Phase, To: next, On: EdgeAdvance, At: at})
	run.Phase = next
	switch Phase(next) {
	case AwaitingApproval: // 落闸锚=挂起（迁移判定落点即判定——规则 13 第一层通用化）
		if !isAutoPassApproval(run.ApprovalMode) { // auto/agent=闸锚自动通过（spec.approval.mode——Agent 节点自动判断级别；直续行不挂起）
			run.Status = RunSuspended
		}
	case Failed: // 落侧轨终点锚=终态（仅声明仍携带 Failed 词形的旧对象 spec 可达——非引擎失败语义）
		run.Status = RunFailed
	}
}

// spawnReq 落点派发请求（锁内收集、出锁派发——ServiceSpawner/DiffSummarizer 非阻塞契约
// 的中转词形）。kind 区分落点类型：service=拉起执行单元；diffdoc=审批判决文档生成
// （2026-11-28 用户裁定 ops_core_inner_agent）。
const (
	spawnKindService = "service"
	spawnKindDiffDoc = "diffdoc"
	spawnKindInner   = "inner"
)

type spawnReq struct {
	kind string    // service（缺省词形兼容）| diffdoc
	run  RunRecord // 派发时点 run 快照（含 Nodes/Target/Involved——拉起/判决投影溯源面）
	node NodeView  // 落点节点投影（unit=service；diffdoc 时零值）
}

// landAfterAdvanceLocked 单步迁移落点（含 service 执行单元挂起面——2026-11-27 service
// session 裁决；含审批判决文档触发面——2026-11-28 ops_core_inner_agent 裁决）：落点=service
// 节点且 ServiceSpawner 已挂接 → 挂起（suspendWhy=service）+ 迁移行 on=service-spawn +
// 返回拉起请求；落点=人审闸锚且实际挂起（auto/agent 直续不触发）且 DiffSummarizer 已挂接
// → 返回判决文档生成请求；其余落点=applyAdvanceLocked 既有语义（返回 nil）。调用方持 r.mu。
// 两个能力面缺席（nil）=零翻转：既有行为面不变。
func (r *RunLog) landAfterAdvanceLocked(run *RunRecord, next string) *spawnReq {
	if r.opts.ServiceSpawns != nil {
		if node, ok := serviceNodeOf(run.Nodes, next); ok {
			at := stampNano(time.Now())
			run.History = append(run.History, HistoryRow{From: run.Phase, To: next, On: EdgeServiceSpawn, At: at})
			run.Phase = next
			run.Status = RunSuspended
			run.SuspendWhy = SuspendService
			return &spawnReq{kind: spawnKindService, run: *run, node: node}
		}
	}
	r.applyAdvanceLocked(run, next)
	// 审批判决文档触发窗口：本拍刚落人审挂起（挂起 run 后续拍被 Status!=RunRunning 跳过，
	// 本函数是唯一命中点）；auto/agent 直续（Status 仍 running）不触发。
	if r.opts.DiffSummarizer != nil && Phase(run.Phase) == AwaitingApproval && run.Status == RunSuspended {
		return &spawnReq{kind: spawnKindDiffDoc, run: *run}
	}
	return nil
}

func isAutoPassApproval(mode string) bool {
	return mode == "auto" || mode == "agent"
}

// indexOfRunLocked 按 run 序号定位环内下标（-1=已被环挤出/不在册）。
func (r *RunLog) indexOfRunLocked(id int64) int {
	for i := range r.ring {
		if r.ring[i].Run == id {
			return i
		}
	}
	return -1
}

// EventAdvanceSink 事件驱动推进面（v0.29 事件驱动推进源首段——巡检拍与推进拍合一，
// workflow.md 规则 15 v0.29 段）——RunSink 的**可选能力接口**：实现侧声明承接「引擎内部源
// 事件落 run 同拍」的全量推进拍（与定时拍同一查表落点/同一落点语义——判定器零改动，只加
// 驱动源）。hub 按能力断言转发（类型断言失败=不转发——能力增量不破既有 sink 契约，
// ResumeSink 同款范式）；**挂接时机由装配层控制**（推进总闸 `--advance-interval` 开启才
// 挂接——事件驱动源不旁路总闸，关断语义组合声明面先定）；转发前提=该事件**落了 run** ∧
// `source ∈ beatSources`（引擎装配的内部源审计事实——event.md 规则 10/规则 12 服务侧
// 判定口径：v0.29 首段生产者=ReconcileTick source=core，v0.30 connector watch 桥接源
// ExternalChanged source=connector；metrics 阈值等其余引擎内部源随各自声明面顺延自动纳入）。
type EventAdvanceSink interface {
	Advance() int
}

// 编译期断言：RunLog 满足 EventAdvanceSink（事件拍面在位——v0.29）。
var _ EventAdvanceSink = (*RunLog)(nil)

// Advancer 装配层定时驱动器（v0.26 异步推进首段）——patrol.Ticker 同款装配范式：
// 生产者=装配层定时器（触发面非请求端点）；cored 装配层持有，每拍调推进入口。
// 缺省关断（interval=0 零推进）：缺省不改既有行为面，按部署显式开启。
type Advancer struct {
	runs     *RunLog
	interval time.Duration
}

// NewAdvancer 构建推进器。interval<=0=关断（Enabled=false，Run 即返）。
func NewAdvancer(runs *RunLog, interval time.Duration) *Advancer {
	return &Advancer{runs: runs, interval: interval}
}

// Enabled 推进器是否开启：interval>0 ∧ runs 已挂接（缺省 0=关断——缺省不改既有行为面）。
func (a *Advancer) Enabled() bool {
	return a != nil && a.interval > 0 && a.runs != nil
}

// Interval 推进间隔（stderr 装配锚报告用；未开启时返回零值）。
func (a *Advancer) Interval() time.Duration {
	if a == nil {
		return 0
	}
	return a.interval
}

// Run 推进主循环：每 interval 到点调用一次推进入口（ctx 取消即退场，幂等可重启）。
// 未开启（Enabled=false）即返——关断=零推进（机械判别器的关断面）。
func (a *Advancer) Run(ctx context.Context) {
	if !a.Enabled() {
		return
	}
	tk := time.NewTicker(a.interval)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			a.runs.Advance() // 查表纯判定无失败路径；零迁移拍=零动作
		}
	}
}

// dispatchSpawnReq 落点请求出锁分派（service=拉起执行单元；diffdoc=审批判决文档生成——
// goroutine 异步，引擎不等待；实现侧幂等/防重自管）。kind 空词形=既有 service 兼容。
func (r *RunLog) dispatchSpawnReq(sp spawnReq) {
	switch sp.kind {
	case spawnKindDiffDoc:
		if r.opts.DiffSummarizer == nil {
			return
		}
		go func(run RunRecord, objKey string) {
			// panic 屏蔽：判决文档是旁路面——引擎主环零阻断（实现侧自行记日志）。
			_ = recover()
			r.opts.DiffSummarizer.OnApprovalSuspended(run, objKey)
		}(sp.run, sp.run.Involved)
	case spawnKindInner:
		if r.opts.InnerAgent == nil {
			return
		}
		go func(run RunRecord, node NodeView) {
			defer func() { _ = recover() }() // panic 不崩引擎（实现侧 trace 自记）
			if err := r.opts.InnerAgent.RunInner(run, node); err != nil {
				return // 失败=留挂起（run.status.lastError 由实现侧写；续行靠重试/人介入）
			}
			r.InnerDone(run.Run, "inner:"+node.Name)
		}(sp.run, sp.node)
	default:
		if r.opts.ServiceSpawns != nil {
			r.opts.ServiceSpawns.SpawnFromRun(sp.run, sp.node)
		}
	}
}
