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
func (r *RunLog) Advance() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	advanced := 0
	mutated := map[string]int{} // 本拍每锚已发生迁移数（刹车闸挂接时生效）
	var touched []int           // 本拍迁移过的环内下标（写穿透 upsert——R-ENG §1）
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
			continue // 序列尾（无 advance 行命中）=零动作停（不冒充执行完毕；不占变更上限额）
		}
		run.Phase = next
		switch Phase(next) {
		case AwaitingApproval: // 落闸锚=挂起（迁移判定落点即判定——规则 13 第一层通用化）
			run.Status = RunSuspended
		case Failed: // 落终点锚=终态（迁移终点语义——非引擎失败语义）
			run.Status = RunFailed
		}
		mutated[run.Workflow]++
		advanced++
		touched = append(touched, i)
	}
	// 写穿透（R-ENG §1）：本拍迁移过的 run 逐条 upsert（锁内同步与投影同序；失败只注记）。
	if r.opts.Store != nil {
		for _, i := range touched {
			persistAppend("run 推进", r.opts.Store.AppendRun(runToRow(r.ring[i])))
		}
	}
	r.advances += int64(advanced)
	return advanced
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
