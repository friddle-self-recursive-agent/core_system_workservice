// opsrun.go —— connectops 节点真执行面（2026-11-28 用户裁定「git提交这个不是功能。删掉功能。
// 做成workflow的一个节点。即gh-release包括git push/release[scheme调用]」——workflow.md §1.3）。
//
// 语义：run 推进落到 unit=connectops 节点（gh-release / verify / issue-comment …）时，run
// **挂起在该节点**（status=suspended，suspendWhy=connectops，history 行 on=connectops-run），
// 同拍把节点投影交给装配层 ConnectOpsRunner（锁内收集锁外派发——ServiceSpawner 同款非阻塞
// 契约）：实现侧按 node.Connector 解析 ops 连接器（connectors 注册表）→ Invoke(node.Op,
// node.Params+run 溯源派生参数) → 完成回执直接续行（进程内调用 RunLog.OpsSignal——完成源在
// 同进程，不走事件注入面；service 完成源在宿主进程故走事件，两路径各自的诚实边界）。
//
// 这是 connectops 执行单元从 declared-deferred（纯查表机械推进——runs 41-43 九相位 converged
// 零真实执行，#5 空转实证）变真实通路的落点：op 调用即效果（connector.md 规则 11 工具调用域），
// 失败=显式失败留挂起（connectops-error 行+note，缺口可见不静默——service 失败同口径），续行后
// 机械推进沿声明表走 advance 边。
//
// 零翻转口径（与 ServiceSpawns/VerifyGate 同范式）：RunOptions.ConnectOpsRunner 缺席（nil）=
// 既有行为面零翻转——connectops 节点照旧纯查表推进，不挂起不调用（装配层显式挂接才生效）。
package workflow

// connectops 信号/挂起词族（与 serviceexec.go 同构——新载荷键=代际判别器）。
const (
	// SignalOpsDone ops 完成回执词形（进程内直接续行用——非总线事件类型，不经 scheme 注册；
	// 留痕在 run history 行 on=connectops-resume / connectops-error）。
	SignalOpsDone = "ConnectOpsCompleted"
	// SuspendOps 挂起因由词（RunRecord.SuspendWhy——connectops=op 执行中）。
	SuspendOps = "connectops"
	// EdgeOpsRun ops 落点迁移行词形（history on——迁移+挂起一体留痕）。
	EdgeOpsRun = "connectops-run"
	// EdgeOpsResume op 完成续行行词形（history on——Note=结果摘要）。
	EdgeOpsResume = "connectops-resume"
	// EdgeOpsError op 失败行词形（history on——Note=显式错误；run 保持挂起，缺口可见）。
	EdgeOpsError = "connectops-error"
)

// opsNodeOf 相位 → connectops 执行单元节点（纯查投影；unit=connectops 才命中——gate/agent/
// service/workflow 节点不落本面）。同相位多节点时首者胜（确定性）。
func opsNodeOf(nodes []NodeView, phase string) (NodeView, bool) {
	for _, n := range nodes {
		if n.Phase == phase && n.Unit == "connectops" {
			return n, true
		}
	}
	return NodeView{}, false
}

// ConnectOpsRunner run 域 connectops 节点 op 执行面（可选能力——RunOptions.ConnectOpsRunner
// 挂接；nil=零翻转口径）。实现须**非阻塞**（Advance 在订阅路径锁内收集、锁外派发）且
// **不得回调 RunLog 同步方法持锁**（完成续行走 OpsSignal——实现侧 goroutine 内调用）。
type ConnectOpsRunner interface {
	// RunOp 一次 op 执行受理：run 快照 + 落点节点投影（Connector/Op/Params 随节点声明面；
	// run 溯源派生（involved 仓/issue 号）随实现侧声明面解析——引擎不感知对象库）。
	RunOp(run RunRecord, node NodeView)
}
