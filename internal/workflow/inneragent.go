// inneragent.go —— 内置轻量 agent 执行单元（2026-11-28 用户裁定「doc-apply 的审批以及部分
// 服务运行起来可以使用内置的 agent 来做润滑剂。即低智商要求的可以使用内置 agent」）。
//
// 词形：节点 phases 行 executor: inner（unit=agent|service）——引擎内直调 Provider 轻量
// LLM（ops_core_inner_agent 词形，不拉 session 零拉起开销）：挂起（suspendWhy=inner-agent）
// →锁外派发 RunInner（实现=装配层 approvaldiff.Runner：op 语义+产物落库）→完成调 InnerDone
// 续行（on=inner-resume 轨迹行，SessionID 空=轻量通路无会话锚）。重型通路（executor 缺省/
// dsh）=既有行为面零翻转。
package workflow

import (
	"fmt"
	"os"
	"time"
)

// EdgeInnerSpawn / EdgeInnerResume 轨迹行词形（新载荷键=代际判别器）。
const (
	EdgeInnerSpawn  = "inner-spawn"
	EdgeInnerResume = "inner-resume"
)

// SuspendInner 挂起因由词（RunRecord.SuspendWhy——内置轻量 agent 执行窗口）。
const SuspendInner = "inner-agent"

// innerGateLocked 停留门：当前相位=executor=inner 的 agent/service 节点且未执行过
// （History 无 inner-spawn 行）→挂起（suspendWhy=inner-agent）+派发请求。执行完成
// InnerDone 翻 running，下拍本门放行（History 标记防重）沿表 advance 离开。
// 调用方持 r.mu；InnerAgent 缺席=零翻转（返回 nil 既有行为面不变）。
func (r *RunLog) innerGateLocked(run *RunRecord) *spawnReq {
	if r.opts.InnerAgent == nil {
		trace("gate-skip nil-inner-agent run=" + fmt.Sprint(run.Run))
		return nil
	}
	node, ok := nodeOfPhase(run.Nodes, run.Phase)
	if !ok {
		trace("gate-skip no-node run=" + fmt.Sprint(run.Run) + " phase=" + run.Phase + " nodes=" + fmt.Sprint(len(run.Nodes)))
		return nil
	}
	if node.Executor != "inner" {
		trace("gate-skip exec=" + node.Executor + " unit=" + node.Unit + " phase=" + run.Phase)
		return nil
	}
	if node.Unit != "agent" && node.Unit != "service" {
		trace("gate-skip unit=" + node.Unit)
		return nil
	}
	for _, h := range run.History {
		if h.On == EdgeInnerSpawn && h.To == run.Phase {
			return nil // 已执行过（InnerDone 续行回本相位）——放行 advance 离开
		}
	}
	at := stampNano(time.Now())
	run.History = append(run.History, HistoryRow{From: run.Phase, To: run.Phase, On: EdgeInnerSpawn, At: at})
	run.Status = RunSuspended
	run.SuspendWhy = SuspendInner
	if r.opts.Store != nil {
		persistAppend("run 内置 agent 挂起", r.opts.Store.AppendRun(runToRow(*run)))
	}
	return &spawnReq{kind: spawnKindInner, run: *run, node: node}
}

// nodeOfPhase 相位 → 节点投影（landInnerLocked 辅助——纯查投影）。
func nodeOfPhase(nodes []NodeView, phase string) (NodeView, bool) {
	for _, n := range nodes {
		if n.Phase == phase {
			return n, true
		}
	}
	return NodeView{}, false
}

// InnerDone 内置轻量 agent 完成信号的 run 域续行（匹配口径=ServiceSignal 同款但只作用于
// SuspendWhy=inner-agent 的现势 run；TargetRun>0=定向覆盖）。命中=沿轨迹落续行行
// （on=inner-resume）、status 翻回 running（SuspendWhy 清空），下拍 Advance 沿表续行。
func (r *RunLog) InnerDone(runID int64, note string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	for i := len(r.ring) - 1; i >= 0; i-- {
		run := &r.ring[i]
		if run.Run != runID {
			continue
		}
		if run.Status != RunSuspended || run.SuspendWhy != SuspendInner {
			return // 非内置 agent 挂起态（已续行/被别的窗口接管）=幂等 noop
		}
		at := stampNano(time.Now())
		run.History = append(run.History, HistoryRow{From: run.Phase, To: run.Phase, On: EdgeInnerResume, At: at, SessionID: note})
		run.Status = RunRunning
		run.SuspendWhy = ""
		if r.opts.Store != nil {
			persistAppend("run 内置 agent 续行", r.opts.Store.AppendRun(runToRow(*run)))
		}
		return
	}
}

// trace 部署探针（stderr 黑洞——文件尾随取证；R-ENG-8 tmp 口径）。
func trace(msg string) {
	if f, err := os.OpenFile("/tmp/approvaldiff.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644); err == nil {
		fmt.Fprintf(f, "%s %s\n", time.Now().Format("15:04:05"), msg)
		f.Close()
	}
}
