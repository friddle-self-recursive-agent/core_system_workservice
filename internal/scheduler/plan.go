// plan.go —— 每拍调度判定（纯函数，可表驱动单测）。
//
// runtime/task-scheduler.md §3 算法的机读落点：
//  1. 依赖 DAG：dependsOn 全 done → ready；有未完成依赖 → blocked（reason=依赖）
//  2. 冲突图：touches.repos/paths/objects 相交 ∨ exclusive 同值 → 互斥边
//  3. 上限：global maxConcurrent / perService（按 scopeRef）/ perRepo（按 touches.repos）
//  4. 选 ready 且与现势任务无冲突、未超上限者；按 priority/等待时长排序取并发额
//
// 判据（§8）：无关任务可并发；同 repo 默认串行（perRepo=1）原因可见；依赖 done 自动解锁；
// 失败只阻塞依赖者。候选未入选=保持 ready + 排队原因可见（不静默等待——blocked 与排队
// 分立：依赖未满足才是 blocked）。
package scheduler

import (
	"fmt"
	"sort"
	"strings"
	"time"
)

// TaskView 单个 Task 的调度视图（对象现势 + 排队起算——readySince 由 tick 层进程内观察，
// 重启后从首拍起算：排队时长可见性以进程生命周期为准，v1 不落库）。
type TaskView struct {
	ID            string // "<scope>/<name>"（taskId 契约词形）
	Scope         string
	Name          string
	Phase         string
	Priority      string
	DependsOn     []string
	Touches       Touches
	Service       string // scopeRef 归一（perService 上限键；空=无影响域声明，键独立）
	Attempts      int
	Workspace     Workspace
	Agent         string
	ReadySince    time.Time     // ready 起算（零值=当拍入队——排序兜底当拍，避免零值抢先）
	MaxConcurrent int           // 单任务覆写（只收紧全局上限）
	Timeout       time.Duration // 单任务租约覆写（只收紧）
	RV            int64         // 对象 resourceVersion 现势（Task watchdog 规则12：rv 推进=活性心跳——2026-11-28）
}

// Decision 单任务本拍处置（Phase 空=不写面；Start=本拍拉起；Reason=阻塞/排队原因——人看，
// 落 status.lastError/lastAction 可见）。
type Decision struct {
	Phase  string
	Reason string
	Start  bool
}

// priorityRank priority 词汇 → 排序权重（critical 最先；词外值=normal 同档，不发明次序）。
func priorityRank(p string) int {
	switch p {
	case "critical":
		return 0
	case "high":
		return 1
	case "low":
		return 3
	default:
		return 2 // normal 与未声明
	}
}

// Conflicts 两任务写集是否互斥（concepts/task-scheduler.md §3.3）：repos/objects/exclusive
// 精确相交；paths 按目录前缀包含判交（父子目录=同一写面——保守口径）。
func Conflicts(a, b Touches) bool {
	if intersects(a.Repos, b.Repos) || intersects(a.Objects, b.Objects) || intersects(a.Exclusive, b.Exclusive) {
		return true
	}
	for _, p := range a.Paths {
		for _, q := range b.Paths {
			if pathOverlap(p, q) {
				return true
			}
		}
	}
	return false
}

func intersects(a, b []string) bool {
	for _, x := range a {
		for _, y := range b {
			if x == y {
				return true
			}
		}
	}
	return false
}

// pathOverlap 路径判交：相等或一方为另一方的目录前缀（尾斜杠归一后比较）。
func pathOverlap(a, b string) bool {
	a, b = strings.Trim(a, "/"), strings.Trim(b, "/")
	if a == "" || b == "" {
		return false
	}
	return a == b || strings.HasPrefix(a+"/", b+"/") || strings.HasPrefix(b+"/", a+"/")
}

// LeaseTTLFor 生效租约：单任务 limits.timeout 只收紧不放宽全局 leaseTtl。
func LeaseTTLFor(t TaskView, cfg Config) time.Duration {
	if t.Timeout > 0 && t.Timeout < cfg.LeaseTTL {
		return t.Timeout
	}
	return cfg.LeaseTTL
}

// Recover 租约回收判定：running 超过生效租约无进展 → attempts+1 回收 ready；
// 回收后 attempts 超 maxAttempts → failed（失败隔离，不拖垮无关任务——§3 步骤 6）。
// startedAt 零值=租约未见（重启后首拍观察起算，不判过期）。
func Recover(t TaskView, startedAt time.Time, cfg Config, now time.Time) Decision {
	if startedAt.IsZero() || now.Sub(startedAt) < LeaseTTLFor(t, cfg) {
		return Decision{}
	}
	att := t.Attempts + 1
	if att > cfg.MaxAttempts {
		return Decision{Phase: PhaseFailed,
			Reason: fmt.Sprintf("租约过期回收重试超上限（attempts=%d/%d）", t.Attempts, cfg.MaxAttempts)}
	}
	return Decision{Phase: PhaseReady,
		Reason: fmt.Sprintf("租约过期回收，重新排队（attempts=%d/%d）", att, cfg.MaxAttempts)}
}

// Plan 每拍调度判定（纯函数）。输入=全部在册 Task 视图 + 生效配置 + 本拍时刻；
// 输出=按 taskId 索引的处置表（只含需写面/可拉起的任务——现势已对的零条目）。
func Plan(tasks []TaskView, cfg Config, now time.Time) map[string]Decision {
	byID := make(map[string]*TaskView, len(tasks))
	for i := range tasks {
		byID[tasks[i].ID] = &tasks[i]
	}
	var active []*TaskView // 现势占额（running/verifying——验证未终态仍持写集）
	for _, t := range byID {
		if t.Phase == PhaseRunning || t.Phase == PhaseVerifying {
			active = append(active, t)
		}
	}

	// 1) 依赖维护：dependsOn 全 done 才可排；未完成 → blocked（reason=依赖现势——人看）。
	depBlocked := map[string]string{}
	for _, t := range byID {
		if terminal(t.Phase) {
			continue
		}
		for _, dep := range t.DependsOn {
			dt, ok := byID[dep]
			if !ok {
				depBlocked[t.ID] = fmt.Sprintf("依赖未完成：%s（在册缺失）", dep)
				break
			}
			if dt.Phase != PhaseDone {
				depBlocked[t.ID] = fmt.Sprintf("依赖未完成：%s:%s", dep, dt.Phase)
				break
			}
		}
	}

	// 2) 候选集：queued/ready/blocked 且依赖已满足（blocked 解锁即回 ready 可参选）。
	var cands []*TaskView
	for _, t := range byID {
		if terminal(t.Phase) || depBlocked[t.ID] != "" {
			continue
		}
		switch t.Phase {
		case PhaseQueued, PhaseReady, PhaseBlocked:
			cands = append(cands, t)
		}
	}
	sort.SliceStable(cands, func(i, j int) bool { // priority 优先、同级按等待时长、末位 ID 兜底（确定性）
		pi, pj := priorityRank(cands[i].Priority), priorityRank(cands[j].Priority)
		if pi != pj {
			return pi < pj
		}
		rsi, rsj := cands[i].ReadySince, cands[j].ReadySince
		if !rsi.Equal(rsj) {
			if rsi.IsZero() {
				return false // 零值（当拍入队）排已知等待者之后
			}
			if rsj.IsZero() {
				return true
			}
			return rsi.Before(rsj)
		}
		return cands[i].ID < cands[j].ID
	})

	decisions := map[string]Decision{}
	selGlobal := 0
	selService := map[string]int{}
	selRepo := map[string]int{}
	selTouches := []*TaskView{}
	for _, c := range cands {
		d := Decision{Phase: PhaseReady} // queued/blocked → ready（解锁即就绪）
		gCap := cfg.MaxConcurrent
		if c.MaxConcurrent > 0 && c.MaxConcurrent < gCap {
			gCap = c.MaxConcurrent // 单任务覆写只收紧
		}
		gUsed := len(active) + selGlobal
		svcUsed := activeService(active, c.Service) + selService[c.Service]
		repoHit := ""
		for _, r := range c.Touches.Repos {
			if activeRepo(active, r)+selRepo[r] >= cfg.PerRepo {
				repoHit = r
				break
			}
		}
		conflictWith := ""
		for _, t := range active {
			if Conflicts(c.Touches, t.Touches) {
				conflictWith = t.ID
				break
			}
		}
		if conflictWith == "" {
			for _, t := range selTouches {
				if Conflicts(c.Touches, t.Touches) {
					conflictWith = t.ID
					break
				}
			}
		}
		switch {
		case gUsed >= gCap:
			d.Reason = fmt.Sprintf("排队：全局并发上限 %d（现势 %d）", gCap, gUsed)
		case svcUsed >= cfg.PerService:
			d.Reason = fmt.Sprintf("排队：同服务并发上限 %d（%s 现势 %d）", cfg.PerService, serviceLabel(c.Service), svcUsed)
		case repoHit != "":
			d.Reason = fmt.Sprintf("排队：同仓串行（perRepo=%d，%s）", cfg.PerRepo, repoHit)
		case conflictWith != "":
			d.Reason = fmt.Sprintf("排队：写集与 %s 互斥（serialize-overlap）", conflictWith)
		default:
			d.Start = true
			selGlobal++
			if c.Service != "" {
				selService[c.Service]++
			}
			for _, r := range c.Touches.Repos {
				selRepo[r]++
			}
			selTouches = append(selTouches, c)
		}
		decisions[c.ID] = d
	}

	// 3) 依赖未满足 → blocked（reason=依赖；已在 blocked 且原因未变=零条目幂等）。
	for _, t := range byID {
		r, ok := depBlocked[t.ID]
		if !ok || terminal(t.Phase) {
			continue
		}
		decisions[t.ID] = Decision{Phase: PhaseBlocked, Reason: r}
	}
	return decisions
}

// terminal 终态（终态制：done/failed/canceled 零出边）。
func terminal(phase string) bool {
	return phase == PhaseDone || phase == PhaseFailed || phase == PhaseCanceled
}

func activeService(active []*TaskView, service string) int {
	n := 0
	for _, t := range active {
		if t.Service == service {
			n++
		}
	}
	return n
}

func activeRepo(active []*TaskView, repo string) int {
	n := 0
	for _, t := range active {
		for _, r := range t.Touches.Repos {
			if r == repo {
				n++
				break
			}
		}
	}
	return n
}

func serviceLabel(service string) string {
	if service == "" {
		return "无影响域"
	}
	return service
}
