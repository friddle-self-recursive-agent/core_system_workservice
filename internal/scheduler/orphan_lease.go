// orphan_lease_scope.go —— 认领归档域收窄（loop-87 活体实证缺口）：
// adoptOrphanLeases 用 CancelSource.QueryEvents 读「本进程 ops 库」归档，但部署现势=
// authority（--allow-apply）与 executor（headless）双 cored 各持独立 ops 库、共享同一
// docs 任务库——调度判定的 running 事件（dedup=task:<id>:running）常落在 executor 的
// ops 归档（本机实证：authority core-system.db 无 17:35 running 行、core-headless.db
// 有）；authority 认领不到该 running 起点→ 拉起后首拍即按「租约起点=now」TTL 到点判
// failed（拉起 4.5 分钟 < leaseTtl 30m 仍被回收——loop-87 18:00:39 拉起/18:05:11 失败
// 实证）。收窄：认领只看「归档域内」的 running——跨域 running（本库无该 running 行）
// 不认领不判据（保守等待人工/部署侧处置），消除跨库失明导致的误回收。
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
// loop-87 收窄：认领域=「本库归档可见的 running」——在册 running 但本库无 running 判定
// 的事件视为跨域 running（另一实例判定的），不认领租约起点（不进 TTL 回收路=不误判 failed）。
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
	inflight, crossDomain := 0, 0
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
	for _, v := range views {
		if v.Phase != PhaseRunning {
			continue
		}
		if _, ok := fresh[v.ID]; !ok {
			crossDomain++ // loop-87：在册 running 本库无 running 判定=跨域 running（另一实例判定面）——不认领
		}
	}
	s.mu.Unlock()
	if inflight > 0 {
		s.logf("孤儿租约认领：event_log 扫档认领 %d 个在途 running（重启前在途任务进入既有租约回收路——TTL 到点 ready/failed）", inflight)
	}
	if crossDomain > 0 {
		s.logf("孤儿租约认领：%d 个在途 running 本库归档无 running 判定（跨域=另一实例判定面）——不认领不回收，保守等待", crossDomain)
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
