// conflict.go —— Apply 关口冲突拦截（BUG-013 / scheme runtime.taskScheduler.applyConflictGate）。
//
// 词形（concepts/task-scheduler.md 规则 11，2026-11-26 用户裁决）：两个并发任务即便调度期
// 未声明重叠 touches，在落地 Apply 关口也会被乐观锁（409）拦截。第二个任务吃 409 后调度器
// 不盲目重试——按写集判定与在先任务有无实质冲突：
//   - 无实质冲突 → 退避重试（执行体侧重试；调度面只注记可见，不阻塞不静默吞掉）；
//   - 有实质冲突 → 当前任务 blocked（reason=apply-conflict 可见）+ 待决事件上抛
//     （ApprovalNeeded，involved=当前任务，dedup 携冲突对——词形=「暂停第一个任务
//     (pause-first) 还是 放弃当前任务(abort-current)」，落在 blocked 原因与任务文件提示），
//     收到 TaskConflictResolved 裁决事件前不再重试该任务。
//
// 裁决事件契约：eventLog 行 Type=TaskConflictResolved、involved=Task/<scope>/<name>（当前
// 任务）、dedupKey 尾段=abort-current|pause-first（事件载荷无 detail 字段——判别词入 dedup）。
// pause-first=暂停在先任务（blocked 让位），当前任务回 ready 继续执行；在先任务在当前任务
// 终态后自动复位 ready（resumeHeld）。abort-current=放弃当前任务（canceled），在先任务不动。
// 归因契约：执行体 Apply 携 fieldManager=task:<scope>/<name>（任务文件 landing 提示随行）——
// apply_log 冲突行/ok 行据此归因任务；非任务归因的冲突行（人/console apply）不进本流程。
package scheduler

import (
	"context"
	"fmt"
	"strings"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
)

// 裁决/归因词形常量（与 conflict.go 头注契约同源）。
const (
	// EventTypeConflictResolved 冲突裁决事件类型（eventLog 扫描键；involved=Task/<id>）。
	EventTypeConflictResolved = "TaskConflictResolved"
	// FieldManagerPrefix 执行体 Apply 归因词形前缀（fieldManager=task:<scope>/<name>）。
	FieldManagerPrefix = "task:"
	// DecisionAbort / DecisionPause 裁决词（dedupKey 尾段；applyConflictGate 声明同源）。
	DecisionAbort = "abort-current"
	DecisionPause = "pause-first"
)

// ApplyLogSource apply_log 读口面（opslog.SQLiteStore 满足；接口化便于测试注入）。
type ApplyLogSource interface {
	QueryApply(result string) ([]opslog.ApplyLog, error)
	QueryApplies(obj string, limit int) ([]opslog.ApplyLog, error)
}

// taskFromManager 归因词形解析：fieldManager "task:<scope>/<name>" → 任务 ID（非任务归因=空）。
func taskFromManager(fm string) string {
	if t, ok := strings.CutPrefix(fm, FieldManagerPrefix); ok && strings.Contains(t, "/") {
		return t
	}
	return ""
}

// conflictAsk 上抛词形（两个候选动作原文——人看，落 blocked 原因与任务文件提示）。
func conflictAsk(first string) string {
	return fmt.Sprintf("apply-conflict：与在先任务 %s 实质写集冲突——"+
		"暂停第一个任务(pause-first) 还是 放弃当前任务(abort-current)？"+
		"等待 TaskConflictResolved 裁决期间不重试", first)
}

// scanApplyConflicts 扫 apply_log 冲突行（result=conflict ∧ 归因任务 ∧ 执行面现势）→
// 实质冲突判定 → 上抛/注记。归档行去重（ts+obj+fieldManager 复合键）；扫描失败=本拍跳过。
func (s *Scheduler) scanApplyConflicts(ctx context.Context, views map[string]TaskView, cfg Config, byObj map[string]*repository.Object) {
	if s.applies == nil || !strings.HasPrefix(cfg.ApplyConflictGate, "ask(") {
		return // 闸门词形只支持 ask(...)——其他词形=行为面未定义，缺省不拦（缺口可见）
	}
	rows, err := s.applies.QueryApply("conflict")
	if err != nil {
		s.logf("apply 冲突行扫描失败（本拍跳过）: %v", err)
		return
	}
	type hit struct{ second, obj string }
	var hits []hit
	s.mu.Lock()
	for _, r := range rows {
		key := fmt.Sprintf("%d|%s|%s", r.Ts.UnixNano(), r.Obj, r.FieldManager)
		if s.conflictSeen[key] {
			continue
		}
		s.conflictSeen[key] = true
		second := taskFromManager(r.FieldManager)
		if second == "" {
			continue // 人/console apply 的冲突——不归调度器判
		}
		hits = append(hits, hit{second: second, obj: r.Obj})
	}
	s.mu.Unlock()
	for _, h := range hits {
		v, ok := views[h.second]
		if !ok || (v.Phase != PhaseRunning && v.Phase != PhaseVerifying) {
			continue // 任务不在册或已离执行面（裁决/取消路径已接手）——不重复上抛
		}
		s.mu.Lock()
		held := s.conflictHold[h.second]
		s.mu.Unlock()
		if held {
			continue
		}
		first := s.pickInFlight(h.obj, h.second, views)
		if first == "" {
			// 找不到在先任务归因=临时冲突（在先者已终态/非任务写面）——退避重试注记。
			action := "Apply 关口乐观锁冲突（无在先任务归因）——执行体退避重试"
			if err := s.writeStatus(ctx, byObj[h.second], statusPatch{LastAction: &action}, "scheduler:apply-conflict"); err != nil {
				s.logf("任务 %s 冲突注记写面失败（下拍重试）: %v", h.second, err)
			}
			continue
		}
		if !Conflicts(v.Touches, views[first].Touches) {
			// 无实质冲突=退避重试（执行体重试；调度面注记可见，不阻塞）。
			action := fmt.Sprintf("Apply 关口乐观锁冲突（与 %s 无实质写集冲突）——执行体退避重试", first)
			if err := s.writeStatus(ctx, byObj[h.second], statusPatch{LastAction: &action}, "scheduler:apply-conflict"); err != nil {
				s.logf("任务 %s 冲突注记写面失败（下拍重试）: %v", h.second, err)
			}
			continue
		}
		s.raiseConflict(ctx, byObj[h.second], h.second, first, views[first].Phase)
	}
}

// raiseConflict 有实质冲突：当前任务 blocked（reason=apply-conflict 词形可见）+ 待决事件上抛
// + 挂起等裁决（不再重试）。
func (s *Scheduler) raiseConflict(ctx context.Context, o *repository.Object, second, first, firstPhase string) {
	reason := conflictAsk(first)
	if err := s.writeStatus(ctx, o, statusPatch{Phase: PhaseBlocked, LastError: &reason}, "scheduler:apply-conflict-block"); err != nil {
		s.logf("任务 %s 冲突挂起写面失败（下拍重试）: %v", second, err)
		return
	}
	s.publishState(second, PhaseBlocked)
	s.publishAsk(second, first) // 待决事件（ApprovalNeeded——人注意族，route notification/*）
	s.mu.Lock()
	s.conflictHold[second] = true
	s.conflictPairs[second] = first
	s.mu.Unlock()
	s.logf("任务 %s 与在先 %s Apply 实质冲突，已挂起待裁决（phase=%s）", second, first, firstPhase)
}

// publishAsk 待决事件（词形已在 blocked 原因与 dedup 对键——事件载荷无文本字段，机制面
// 只出 involved/dedup：ApprovalNeeded involved=当前任务，dedup=task-apply-conflict:<first>:<second>；
// ApprovalNeeded 未注册（旧声明面）=零事件缺口可见，注册闸门与 TaskStateChanged 同范式）。
func (s *Scheduler) publishAsk(second, first string) {
	et, err := s.sch.LookupEventType(eventTypeAsk)
	if err != nil {
		return
	}
	_ = s.bus.Publish(&eventbus.Event{
		Type: eventTypeAsk, Source: "core", Involved: TaskKind + "/" + second,
		DedupKey: "task-apply-conflict:" + first + ":" + second, Level: et.DefaultLevel,
	})
}

// resolveConflicts 扫裁决事件（TaskConflictResolved；dedupKey 尾段=abort-current|pause-first）
// → 两路落面：abort-current=放弃当前任务（canceled）；pause-first=暂停在先任务（blocked 让位）、
// 当前任务回 ready 继续执行。
func (s *Scheduler) resolveConflicts(ctx context.Context, byObj map[string]*repository.Object, byView map[string]TaskView) {
	if s.cancels == nil {
		return
	}
	rows, err := s.cancels.QueryEvents(opslog.EventQuery{Type: EventTypeConflictResolved, Limit: opslog.MaxEventLimit})
	if err != nil {
		s.logf("冲突裁决事件扫描失败（本拍跳过）: %v", err)
		return
	}
	type order struct{ id, decision string }
	var orders []order
	s.mu.Lock()
	for _, r := range rows {
		if s.conflictResolvedSeen[r.EventID] {
			continue
		}
		s.conflictResolvedSeen[r.EventID] = true
		id := strings.TrimPrefix(strings.TrimSpace(r.Involved), TaskKind+"/")
		decision := decisionFromDedup(r.DedupKey)
		if id == "" || decision == "" {
			continue // involved/dedup 不合契约词形——忽略（不冒充全量处置）
		}
		if _, held := s.conflictHold[id]; !held {
			continue // 无挂起对（事件先行于冲突/已裁决）——忽略
		}
		orders = append(orders, order{id: id, decision: decision})
	}
	s.mu.Unlock()
	for _, od := range orders {
		first := s.pairOf(od.id)
		switch od.decision {
		case DecisionAbort: // 放弃当前任务
			action := fmt.Sprintf("已裁决 abort-current：放弃当前任务（与在先 %s 的 Apply 冲突）", first)
			cleared := ""
			if err := s.writeStatus(ctx, byObj[od.id], statusPatch{
				Phase: PhaseCanceled, LastAction: &action, LastError: &cleared,
			}, "scheduler:conflict-abort"); err != nil {
				s.logf("任务 %s abort-current 裁决落面失败（下拍重试）: %v", od.id, err)
				continue
			}
			s.publishState(od.id, PhaseCanceled)
			s.releaseHold(od.id)
			s.logf("任务 %s 已裁决放弃（abort-current，对端 %s）", od.id, first)
		case DecisionPause: // 暂停在先任务，当前任务回 ready 继续执行
			if fo := byObj[first]; fo != nil {
				preason := fmt.Sprintf("已裁决 pause-first：暂停让位（对端 %s 执行中；其终态后自动复位 ready）", od.id)
				if err := s.writeStatus(ctx, fo, statusPatch{Phase: PhaseBlocked, LastError: &preason}, "scheduler:conflict-pause"); err != nil {
					s.logf("在先任务 %s 暂停写面失败（下拍重试）: %v", first, err)
					continue
				}
				s.publishState(first, PhaseBlocked)
			}
			action := fmt.Sprintf("已裁决 pause-first：在先 %s 已暂停让位，继续执行", first)
			cleared := ""
			if err := s.writeStatus(ctx, byObj[od.id], statusPatch{
				Phase: PhaseReady, LastAction: &action, LastError: &cleared,
			}, "scheduler:conflict-resume"); err != nil {
				s.logf("任务 %s 冲突解除写面失败（下拍重试）: %v", od.id, err)
				continue
			}
			s.publishState(od.id, PhaseReady)
			s.mu.Lock()
			delete(s.conflictHold, od.id) // 解除挂起；对子保留（当前任务终态后在先任务自动复位）
			s.mu.Unlock()
			s.logf("任务 %s 已裁决 pause-first（在先 %s 暂停让位）", od.id, first)
		}
	}
}

// resumeHeld 让位复位：冲突对子中当前任务已终态/不在册 → 在先任务（blocked 让位中）自动回 ready。
func (s *Scheduler) resumeHeld(ctx context.Context, byObj map[string]*repository.Object, byView map[string]TaskView) {
	s.mu.Lock()
	pairs := make(map[string]string, len(s.conflictPairs))
	for second, first := range s.conflictPairs {
		pairs[second] = first
	}
	s.mu.Unlock()
	for second, first := range pairs {
		if sv, ok := byView[second]; ok && !terminal(sv.Phase) {
			continue // 当前任务仍在执行/排队——让位继续
		}
		fv, ok := byView[first]
		if !ok {
			s.mu.Lock()
			delete(s.conflictPairs, second)
			s.mu.Unlock()
			continue
		}
		if fv.Phase == PhaseBlocked {
			action := fmt.Sprintf("复位 ready（冲突对端 %s 已终态，让位解除）", second)
			cleared := ""
			if err := s.writeStatus(ctx, byObj[first], statusPatch{
				Phase: PhaseReady, LastAction: &action, LastError: &cleared,
			}, "scheduler:conflict-resume-held"); err != nil {
				s.logf("在先任务 %s 复位写面失败（下拍重试）: %v", first, err)
				continue
			}
			s.publishState(first, PhaseReady)
			s.logf("在先任务 %s 已复位 ready（对端 %s 终态）", first, second)
		}
		s.mu.Lock()
		delete(s.conflictPairs, second)
		s.mu.Unlock()
	}
}

// pickInFlight 归因在先任务：obj 的 ok 行里最近一条任务归因写面（fieldManager=task:<id>），
// 且该任务现势 running/verifying（在先执行中）；找不到=空（临时冲突，退避不挂起）。
func (s *Scheduler) pickInFlight(obj, second string, views map[string]TaskView) string {
	rows, err := s.applies.QueryApplies(obj, 200)
	if err != nil {
		return ""
	}
	for _, r := range rows { // 最新优先（rv 降序——QueryApplies 排序契约）
		if r.Result != "ok" {
			continue
		}
		fid := taskFromManager(r.FieldManager)
		if fid == "" || fid == second {
			continue
		}
		if fv, ok := views[fid]; ok && (fv.Phase == PhaseRunning || fv.Phase == PhaseVerifying) {
			return fid
		}
	}
	return ""
}

// releaseHold 挂起解除（取消/终态路径共用——取消优先于冲突挂起）。
func (s *Scheduler) releaseHold(id string) {
	s.mu.Lock()
	delete(s.conflictHold, id)
	delete(s.conflictPairs, id)
	s.mu.Unlock()
}

// pairOf 挂起对的对端任务 ID（无=空串）。
func (s *Scheduler) pairOf(id string) string {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.conflictPairs[id]
}

// decisionFromDedup 裁决词解析（dedupKey 尾段=abort-current|pause-first；不合词形=空）。
func decisionFromDedup(dedup string) string {
	parts := strings.Split(strings.TrimSpace(dedup), ":")
	if len(parts) == 0 {
		return ""
	}
	last := parts[len(parts)-1]
	if last == DecisionAbort || last == DecisionPause {
		return last
	}
	return ""
}
