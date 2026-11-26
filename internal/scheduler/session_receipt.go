// session_receipt.go —— Task done 判定的真实会话终结信号接线（G-08 收口——R-UX-3/4
// 可观测链「事件详情↔最近 loop↔session」与「Task done 无真实 sessionId 回填」缺口的
// 调度器侧落点）。
//
// 语义锚（SessionEnded——opslog.SessionStateEnded 会话终结词）：agent 会话（dsh/宿主拉起）
// 终结后，会话侧经既有心跳注册面把行态置 ended——session_heartbeat 行即「真实会话终结」的
// ops 权威投影（心跳=活性租约的同一账面，终结=租约终态，词汇闭合集只加不改）。调度器
// 每拍扫描：running/verifying 任务锚定的会话行出现 ended → **done 判定成立**：
//   - phase 迁移 running|verifying → done（migrations 闭合集内合法迁移——task.go 同表判别）；
//   - **agentSession 回填真实 sessionId**（行 session_id——spawn 回执 hint 缺席时由终结行
//     兜底回填；「done→agentSession 非空」机械判据的落点）；
//   - lastAction 注记终结语义（审计面可见）；TaskStateChanged done 事件同拍发布；
//   - 租约出清（done=离场——租约回收面不再作用）。
//
// 匹配两条路（先精确后兜底，零重复处置）：
//
//	① 精确=Task.status.agentSession == 行 session_id（spawn 回执 hint 已在）；
//	② 兜底=agentSession 空 ∧ 行 task_name == 任务 id（会话侧心跳携任务锚——hint 面缺席
//	  部署的回填路；回填后下拍走①不再重复）。
//
// 边界（与既有调度面同款纪律）：扫描失败=本拍跳过（可丢语义，下拍续扫）；非 running/
// verifying 任务零作用（终态不迁移——闭合集出边表权威）；行缺席/非 ended 词=零动作
// （停滞/恢复词形不是终结——判定只认 SessionEnded 语义，不冒充）。
package scheduler

import (
	"context"
	"strings"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
)

// SessionSource 会话终结信号的读口面（opslog.SQLiteStore 满足；接口化便于测试注入——
// CancelSource 同款口径）。终结词形读经既有 AllSessionHeartbeats 投影，零新存储面。
type SessionSource interface {
	AllSessionHeartbeats(limit int) ([]opslog.SessionHeartbeat, error)
}

// observeSessionEnds 每拍会话终结扫描（Tick 尾段——done 判定与 agentSession 回填单一落点）。
// rows 一次装载逐任务复用（O(会话数) 每拍一次，不逐任务查库）。
func (s *Scheduler) observeSessionEnds(ctx context.Context, byObj map[string]*repository.Object, byPhase map[string]string, bySession map[string]string) {
	if s.sessions == nil {
		return
	}
	rows, err := s.sessions.AllSessionHeartbeats(opslog.MaxEventLimit)
	if err != nil {
		s.logf("会话终结扫描失败（本拍跳过——可丢语义下拍续扫）: %v", err)
		return
	}
	byHeartbeatSession := make(map[string]opslog.SessionHeartbeat, len(rows))
	byHeartbeatTask := make(map[string]opslog.SessionHeartbeat, len(rows))
	for _, h := range rows {
		byHeartbeatSession[h.SessionID] = h
		if tn := strings.TrimSpace(h.TaskName); tn != "" {
			byHeartbeatTask[tn] = h // 同名多会话=尾写胜（最新终结行）
		}
	}
	for id, phase := range byPhase {
		if phase != PhaseRunning && phase != PhaseVerifying {
			continue // 终态不迁移/排队态无会话可判（闭合集出边表权威）
		}
		o, ok := byObj[id]
		if !ok {
			continue
		}
		hb, ok := endedRowFor(id, bySession[id], byHeartbeatSession, byHeartbeatTask)
		if !ok {
			continue // 无终结行（会话在跑/停滞/恢复——非 SessionEnded 语义零动作）
		}
		// done 判定成立：迁移 + agentSession 回填真实 sessionId（非空判据锚）+终结注记。
		sess := hb.SessionID
		action := "会话终结（SessionEnded）→done（真实 sessionId 回填）"
		patch := statusPatch{Phase: PhaseDone, AgentSession: &sess, LastAction: &action}
		cleared := ""
		patch.LastError = &cleared
		if err := s.writeStatus(ctx, o, patch, "scheduler:session-ended"); err != nil {
			s.logf("任务 %s 会话终结 done 写面失败（下拍重试）: %v", id, err)
			continue
		}
		s.publishState(id, PhaseDone)
		s.mu.Lock()
		delete(s.leases, id) // done=离场：租约终止（回收面不再作用）
		s.mu.Unlock()
		s.logf("任务 %s done（会话 %s 终结——agentSession 已回填）", id, sess)
	}
}

// endedRowFor 任务 id 的终结行判定（先精确后兜底；行非 ended 词=零动作——只认 SessionEnded
// 语义，停滞/恢复/升级词形不冒充终结）。
func endedRowFor(taskID, agentSession string, bySession, byTask map[string]opslog.SessionHeartbeat) (opslog.SessionHeartbeat, bool) {
	if sess := strings.TrimSpace(agentSession); sess != "" {
		if h, ok := bySession[sess]; ok && h.State == opslog.SessionStateEnded {
			return h, true // ① 精确路：任务锚定会话的终结行
		}
		return opslog.SessionHeartbeat{}, false // 锚定会话未终结——兜底路不启用（零误判）
	}
	if h, ok := byTask[taskID]; ok && h.State == opslog.SessionStateEnded {
		return h, true // ② 兜底路：hint 缺席——会话心跳携任务锚的终结行（回填源）
	}
	return opslog.SessionHeartbeat{}, false
}
