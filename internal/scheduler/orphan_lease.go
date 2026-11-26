// orphan_lease.go —— 孤儿租约认领（task-scheduler.md §3 补口）：
// 租约起点=进程内本轮拉起时刻（scheduler.go recoverLeases 注释自认「重启后零值=首拍观察
// 起算」），但 running 视图无需进程内登记也能在册——cored 重启（loop-82/83 实证 14:21
// addr 34643→38307 换代）后进程内 leases 映射清零，而任务 status.phase=running 留在仓库：
// recoverLeases 只扫进程内租约 → 孤儿 running 永不回收也永不重拉（attempts 恒 0、
// lastAction 恒「拉起中（spawn）」）；Plan 又把 running 计入现势占额 → 阻塞全部后续派单。
// 本文件补口：本进程首拍扫 event_log 归档取每任务最近一次 TaskStateChanged running
// 事件（dedup=task:<id>:running 机械键）→ 在册视图仍 running 的认领为租约起点；回收仍走
// recoverLeases 同一条路（租约/退避/attempts 语义零新面，事件可回查——审计事实非凭空推断）。
package scheduler

import (
	"context"
	"strings"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
)

// adoptOrphanLeases 孤儿租约认领：从 event_log 归档恢复每任务最近一次 running 判定时刻。
// 本进程仅首拍扫档一次（adopted 闸——保守不覆盖在途租约；归档读失败=本进程不认领）。
// 归档读失败=本进程不认领（读面失败 Tick 跳过同口径——保守不半途半面；重启再试）。
func (s *Scheduler) adoptOrphanLeases(ctx context.Context, views []TaskView, now time.Time) {
	s.mu.Lock()
	if s.adopted || s.cancels == nil {
		s.mu.Unlock()
		return // 已扫档（本进程一次）；归档读口面未装配=认领关断（既有行为面零翻转）
	}
	s.adopted = true
	s.mu.Unlock()
	rows, err := s.cancels.QueryEvents(opslog.EventQuery{Type: EventTypeStateChanged, Limit: opslog.MaxEventLimit})
	if err != nil {
		s.logf("孤儿租约认领：event_log 读面失败（下拍重试）: %v", err)
		return
	}
	fresh := map[string]time.Time{}
	for _, r := range rows {
		if !strings.HasPrefix(r.DedupKey, "task:") || !strings.HasSuffix(r.DedupKey, ":"+PhaseRunning) {
			continue // 只认领 running 判定（dedup=task:<id>:<phase> 尾段判别——机械键）
		}
		id := strings.TrimSuffix(strings.TrimPrefix(r.DedupKey, "task:"), ":"+PhaseRunning)
		if prev, ok := fresh[id]; ok && !r.Ts.After(prev) {
			continue // 同任务取最新一拍
		}
		fresh[id] = r.Ts
	}
	inflight := 0
	s.mu.Lock()
	for id, at := range fresh {
		v, ok := byViewID(views, id)
		if !ok || v.Phase != PhaseRunning {
			continue // 非 running=无需认领（终态/离场——归档行只作历史）
		}
		s.orphanOrders[id] = at
		if _, leased := s.leases[id]; !leased {
			s.leases[id] = at // 认领为租约起点 → recoverLeases 走既有 TTL 回收路
			inflight++
		}
	}
	s.mu.Unlock()
	if inflight > 0 {
		s.logf("孤儿租约认领：event_log 扫档认领 %d 个在途 running（重启前在途任务进入既有租约回收路——TTL 到点 ready/failed）", inflight)
	}
}

// byViewID 视图查找（views 每拍重建的小切片——线性即可，不引索引面）。
func byViewID(views []TaskView, id string) (TaskView, bool) {
	for _, v := range views {
		if v.ID == id {
			return v, true
		}
	}
	return TaskView{}, false
}
