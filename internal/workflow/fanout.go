// fanout.go —— 批量合判扇出执行面（2026-11-28 用户裁定第四词形 unit=workflow / workservice #6）。
//
// 语义：run 推进落到 unit=workflow ∧ fanout:{from,key} 节点时挂起（suspendWhy=fanout，
// history on=fanout-spawn），锁外由 FanoutLauncher 读执行计划（batch-judge 产物），
// 为每项建一条子 run（entry:sub:plan-item 锚切入、载荷带 parentRun）；子 run 序号回填父
// run.SubRuns，父 run join 挂起；全部子 run 终态后 fanout-complete 沿声明表迁移到
// merge-results 汇聚相位。run 投影新增 parentRun/subRuns 键（新载荷键=代际判别器）。
//
// 零翻转口径：RunOptions.FanoutLauncher 未挂接（nil）=旧词形（unit=workflow 节点照旧
// 纯查表推进不扇出），与 VerifyGate/ServiceSpawns 同范式。
package workflow

import (
	"fmt"
	"strings"
	"time"
)

const (
	// SuspendFanout 挂起因由词（父 run 等子 run 终态的 join 窗口）。
	SuspendFanout = "fanout"
	// EdgeFanoutSpawn 扇出派发轨迹行（父 run 落点同拍）。
	EdgeFanoutSpawn = "fanout-spawn"
	// EdgeFanoutComplete 全子 run 终态后的汇聚迁移轨迹行（声明边 on=fanout-complete）。
	EdgeFanoutComplete = "fanout-complete"
	// EdgeChildConverged 子 run 终态通知父 run 的轨迹行（子 run 侧留痕）。
	EdgeChildConverged = "child-done"
	// SubRunSpawned / SubRunTerminal 子 run 行状态词形。
	SubRunSpawned  = "spawned"
	SubRunConverged = "converged"
	SubRunFailed    = "failed"
)

// ChildRunSpec 子 run 声明（实现侧从执行计划解析；引擎只消费机器词形，零文档库依赖）。
type ChildRunSpec struct {
	Key      string         // 计划项键（同父 run 内幂等 dedup 锚）
	Workflow string         // 子 run 目标 workflow（空=父 workflow）
	Type     string         // 入口事件（空=sub:plan-item，命中子 spec 的 entry: 迁移行）
	Involved string         // 子 run involved 锚（空=继承父 run）
	Payload  map[string]any // 计划项载荷（issues/scope/risk——透传留痕，不改引擎语义）
}

// FanoutLauncher 扇出执行面（实现侧读执行计划+解析项；非阻塞契约——Advance 锁外派发）。
// 返回空切片=计划无项（父 run 保持 join 挂起，缺口可见不冒充完成；实现侧可发布
// fanout-complete 信号显式放行）。
type FanoutLauncher interface {
	LaunchChildRuns(run RunRecord, node NodeView) []ChildRunSpec
}

// dispatchFanout 扇出派发（锁外 goroutine——文档库 IO 不持 run 锁；panic 不崩主环）。
func (r *RunLog) dispatchFanout(sp spawnReq) {
	if r.opts.FanoutLauncher == nil {
		return
	}
	go func(run RunRecord, node NodeView) {
		defer func() { _ = recover() }()
		items := r.opts.FanoutLauncher.LaunchChildRuns(run, node)
		refs := make([]SubRunRef, 0, len(items))
		for _, item := range items {
			id, ok := r.StartChildRun(run, item)
			if !ok {
				continue
			}
			refs = append(refs, SubRunRef{Run: id, ItemKey: item.Key, Status: SubRunSpawned})
		}
		r.AttachSubRuns(run.Run, refs)
		if r.opts.OnFanoutLaunched != nil {
			r.opts.OnFanoutLaunched(run, refs)
		}
	}(sp.run, sp.node)
}

// StartChildRun 建一条子 run（父 run 锚随 TriggerRecord.ParentRun 传递；dedupKey=
// fanout:<parent>:<itemKey> 同父同项幂等——重放/重启不重复建子 run）。返回子 run 序号。
func (r *RunLog) StartChildRun(parent RunRecord, item ChildRunSpec) (int64, bool) {
	wf := strings.TrimSpace(item.Workflow)
	if wf == "" {
		wf = parent.Workflow
	}
	typ := strings.TrimSpace(item.Type)
	if typ == "" {
		typ = "sub:plan-item"
	}
	dedup := fmt.Sprintf("fanout:%d:%s", parent.Run, item.Key)
	r.mu.Lock()
	for i := range r.ring {
		if r.ring[i].ParentRun == parent.Run && r.ring[i].DedupKey == dedup {
			r.mu.Unlock()
			return r.ring[i].Run, true // 幂等命中（重放不重复拉起）
		}
	}
	seq := r.seq + 1 // 预占（StartRun 内部再 seq++）——仅为判定，实际以 StartRun 回执为准
	r.mu.Unlock()
	rec := TriggerRecord{
		Seq: seq, Type: typ, DedupKey: dedup, ReceivedAt: time.Now(),
		Route: "workflows/" + wf, Involved: item.Involved, ParentRun: parent.Run,
	}
	if rec.Involved == "" {
		rec.Involved = parent.Involved
	}
	r.StartRun(rec, ExecutorBinding{Executor: ExecWorkflow, Target: wf})
	st := r.State()
	for i := len(st.Runs) - 1; i >= 0; i-- {
		if st.Runs[i].DedupKey == dedup && st.Runs[i].ParentRun == parent.Run {
			return st.Runs[i].Run, true
		}
	}
	return 0, false
}

// AttachSubRuns 子 run 行集回填父 run（同 itemKey 幂等；空切片=零动作——计划无项不冒充完成）。
func (r *RunLog) AttachSubRuns(parentRun int64, refs []SubRunRef) {
	if len(refs) == 0 {
		return
	}
	r.mu.Lock()
	idx := r.indexOfRunLocked(parentRun)
	if idx < 0 {
		r.mu.Unlock()
		return
	}
	run := &r.ring[idx]
	seen := map[string]bool{}
	for _, x := range run.SubRuns {
		seen[x.ItemKey] = true
	}
	for _, ref := range refs {
		if seen[ref.ItemKey] {
			continue
		}
		seen[ref.ItemKey] = true
		run.SubRuns = append(run.SubRuns, ref)
	}
	if r.opts.Store != nil {
		persistAppend("run 扇出子 run 回填", r.opts.Store.AppendRun(runToRow(*run)))
	}
	r.mu.Unlock()
}

// FanoutSignal 子 run 终态信号（父 run join 判据）：回填 SubRuns[i].Status；全部终态
// （converged/failed）才沿声明表 fanout-complete 边迁移到汇聚相位并翻 running。
// 调用方=子 run 的父通知（notifyParentLocked）或装配层显式信号（TargetRun=父 run）。
func (r *RunLog) FanoutSignal(parentRun, childRun int64, status string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.fanoutSignalLocked(parentRun, childRun, status)
}

// notifyParentLocked 子 run 终态（converged/failed）时通知父 run join（调用方持 r.mu；
// 在 Advance 的收束/终态落点同拍调用——join 不靠外部事件，防止子 run 被环挤出丢信号）。
func (r *RunLog) notifyParentLocked(child *RunRecord) {
	if child.ParentRun == 0 {
		return
	}
	status := SubRunConverged
	if child.Status == RunFailed {
		status = SubRunFailed
	}
	child.History = append(child.History, HistoryRow{From: child.Phase, To: child.Phase, On: EdgeChildConverged, At: stampNano(time.Now())})
	// 复用 FanoutSignal 判定体（此处已持锁——直接内联调用会重入 mutex，故走无锁内部函数）。
	r.fanoutSignalLocked(child.ParentRun, child.Run, status)
}

// fanoutSignalLocked FanoutSignal 的无锁内核（notifyParentLocked 已持 r.mu）。
func (r *RunLog) fanoutSignalLocked(parentRun, childRun int64, status string) bool {
	idx := r.indexOfRunLocked(parentRun)
	if idx < 0 {
		return false
	}
	run := &r.ring[idx]
	if run.Status != RunSuspended || run.SuspendWhy != SuspendFanout {
		return false
	}
	found, allDone := false, len(run.SubRuns) > 0
	for i := range run.SubRuns {
		if run.SubRuns[i].Run == childRun {
			run.SubRuns[i].Status = status
			found = true
		}
		if run.SubRuns[i].Status != SubRunConverged && run.SubRuns[i].Status != SubRunFailed {
			allDone = false
		}
	}
	if !found || !allDone {
		if r.opts.Store != nil && found {
			persistAppend("run 扇出子 run 终态", r.opts.Store.AppendRun(runToRow(*run)))
		}
		return false
	}
	next, ok := stepByTable(run.Transitions, run.Phase, EdgeFanoutComplete)
	if !ok {
		return false
	}
	histAppend(run, run.Phase, next, EdgeFanoutComplete)
	run.Phase, run.Status, run.SuspendWhy = next, RunRunning, ""
	if r.opts.Store != nil {
		persistAppend("run 扇出汇聚", r.opts.Store.AppendRun(runToRow(*run)))
	}
	return true
}
