// watcher.go —— Proposal 分流/派单 watcher（v0.51 能力 B）：DocApplied×Proposal 订阅热路径
// （flowbind.Materializer v0.49 同款范式——装配层 goroutine 承载，非阻塞 sink）。
package proposal

import (
	"context"
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/apply"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/flowbind"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// ApprovalLogger 审批判定账面写半边（opslog.SQLiteStore 满足；auto-approve 判定行落账——
// 账面非权威，落账失败不翻转批准结果）。nil=零落账（缺口可见不静默——stderr 告警）。
type ApprovalLogger interface {
	AppendApproval(l opslog.ApprovalLog) error
}

// Watcher 提案分流/派单 watcher（装配层持有；零值不可用——Repo/Apply/Bus/Sch 必配）。
type Watcher struct {
	Repo        repository.Repository
	Apply       *apply.Engine
	Bus         *eventbus.Bus
	Sch         *scheme.Scheme
	Dispatch    flowbind.DispatchSink // 派单投影落点（v0.50 派单面同 sink 同 face；nil=静默零派单）
	Approvals   ApprovalLogger        // 审批账面写半边（auto-approve granted 行；nil=零落账）
	AutoApprove []string              // 低危自动批准 priority 清单（部署侧分流配置；空=全人审——宁慢勿越）
}

// involvedRef 三段同构拆解（kind/scope/name——与 flowbind 同惯例；scope 四层词汇
// builtin|global|<system>|<system>/<instance> 无内嵌斜杠段，恰三分）。
func involvedRef(involved string) (kind, scope, name string, ok bool) {
	parts := strings.Split(involved, "/")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", false
	}
	return parts[0], parts[1], parts[2], true
}

// Start 订阅热路径（DocApplied×Proposal 精确过滤——其余对象 apply 零无效读）。阻塞调用，
// 装配层 go 例程承载；onLanded(action)=一次分流/派单动作完成（stderr 装配锚）；onErr=单提案
// 处理失败（保继续——单点失败不打死 watcher，下批 apply 重试自愈）。
func (w *Watcher) Start(ctx context.Context, bus *eventbus.Bus, onLanded func(action string), onErr func(err error)) {
	if w == nil || bus == nil {
		return
	}
	ch, cancel := bus.Subscribe("DocApplied")
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		case ev := <-ch:
			kind, scope, name, ok := involvedRef(ev.Involved)
			if !ok || kind != "Proposal" {
				continue // 精确过滤：仅 Proposal 对象 apply 触发（flowbind/阈值桥同款）
			}
			action, err := w.handleLanding(ctx, scope, name, ev.ID)
			if err != nil && onErr != nil {
				onErr(err)
			}
			if action != "" && onLanded != nil {
				onLanded(action)
			}
		}
	}
}

// handleLanding 一次 Proposal 落库的分流判定（objKey=kind/scope/name 三段同构）。
// 返回动作词（approval-needed|auto-approved|dispatched|""=零动作）。
func (w *Watcher) handleLanding(ctx context.Context, scope, name, eventID string) (string, error) {
	if w.Repo == nil || w.Apply == nil || w.Sch == nil {
		return "", nil
	}
	obj, err := w.Repo.Get(ctx, "Proposal", scope, name)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return "", nil // 提案删除竞态（对象已不在册）——零动作
		}
		return "", fmt.Errorf("proposal: 读取落库对象: %w", err)
	}
	spec, ok := Parse(obj.Content)
	if !ok {
		return "", fmt.Errorf("proposal: %s/%s 落库对象不可解析（写路径闸门应已拦——读面防御）", scope, name)
	}
	switch spec.Status {
	case StatusSubmitted:
		if w.autoApproves(spec.Priority) {
			return w.autoApprove(ctx, scope, name, obj, spec)
		}
		w.emitApprovalNeeded(fmt.Sprintf("Proposal/%s/%s", scope, name))
		return "approval-needed", nil
	case StatusApproved:
		w.dispatchApproval(scope, name, spec, obj.ResourceVersion, eventID)
		return "dispatched", nil
	default:
		return "", nil // draft（不入库形）/rejected（判定终态不派单）/landed（产物已回写）——零动作
	}
}

// autoApproves 分流判定（priority ∈ AutoApprove 清单=低危可自动；空清单=全人审——缺省不改行为面）。
func (w *Watcher) autoApproves(priority string) bool {
	if priority == "" {
		return false
	}
	for _, p := range w.AutoApprove {
		if p == priority {
			return true
		}
	}
	return false
}

// autoApprove 低危自动批准（引擎自写——subject=引擎缺省 fallback cored:local；走同一三步闸门
// 零旁路）。落库的 DocApplied 再入 watcher → dispatchApproval（单环两拍，approved 不再写面=无自激）。
func (w *Watcher) autoApprove(ctx context.Context, scope, name string, obj *repository.Object, spec Spec) (string, error) {
	updated, ok := FlipStatus(obj.Content, StatusSubmitted, StatusApproved)
	if !ok {
		return "", fmt.Errorf("proposal: %s/%s 状态行翻转失败（spec.status=submitted 未命中——形态漂移）", scope, name)
	}
	res, err := w.Apply.Apply(ctx, &repository.Object{
		Kind: "Proposal", Name: name, Scope: scope,
		ResourceVersion: obj.ResourceVersion, Content: updated,
	}, apply.Options{FieldManager: FieldManagerAuto, ReasonRef: ReasonRefAuto + ":" + spec.Priority})
	if err != nil {
		return "", fmt.Errorf("proposal: %s/%s 自动批准落库失败（保 submitted 面，下次 apply 重试自愈）: %w", scope, name, err)
	}
	if res.Action == "noop" {
		return "", nil // 幂等重入（approved 已在面）——零动作零噪音
	}
	// 判定落账（approval_log granted 行：judgedBy=引擎主体、fieldManager=proposal:auto——
	// 账面非权威：落账失败只注记告警，批准结果不变）。
	if w.Approvals != nil {
		if err := w.Approvals.AppendApproval(opslog.ApprovalLog{
			Ts: time.Now(), ApprovalID: "proposal:" + scope + "/" + name,
			Obj: fmt.Sprintf("Proposal/%s/%s", scope, name), Decision: opslog.DecisionGranted,
			JudgedBy: "cored:local", FieldManager: FieldManagerAuto,
			ResourceVersion: res.RV, ReasonRef: ReasonRefAuto + ":" + spec.Priority,
		}); err != nil {
			log.Printf("proposal: auto-approve approval_log 落账失败（非权威，批准不受影响）: %v", err)
		}
	}
	return "auto-approved", nil
}

// emitApprovalNeeded 高危提交的自产事件（event.md 规则 15）：source=core、dedupKey=proposal:<obj>、
// level=注册表缺省投影；类型注册闸门前置（未注册即不发布）。**分流判定是权威、事件是通知**——
// 发布失败不改变「待人审」事实（审批屏待审区另有草稿/对象行兜底可见）。
func (w *Watcher) emitApprovalNeeded(objKey string) {
	if w.Bus == nil || w.Sch == nil {
		return
	}
	et, err := w.Sch.LookupEventType("ApprovalNeeded")
	if err != nil {
		return // 未注册不发布（注册表闸门，声明先行的机械落点）
	}
	evt := eventbus.Event{
		Type: "ApprovalNeeded", Source: approvalSourceCore, Involved: objKey,
		DedupKey: "proposal:" + objKey,
		Level:    et.DefaultLevel, // 注册表缺省投影（ApprovalNeeded→warn）；空值由总线归一 info
	}
	_ = w.Bus.Publish(&evt) // 可丢语义（event.md §1）；审批事实在对象状态机，不在事件
}

// dispatchApproval 批准派单（proposal.md 规则 2）：status=approved 落库即一条对 apply-workflow 锚的
// 派单受理投影（同 face GET /v1/agents/dispatches，kind=proposal 行——载荷带提案名/目标/验收/优先级/
// 提出人，执行体输入随行；eventId=本落库 DocApplied 回执——来源链可回指）。sink 缺席=静默。
func (w *Watcher) dispatchApproval(scope, name string, spec Spec, rv int64, eventID string) {
	if w.Dispatch == nil {
		return
	}
	w.Dispatch.Dispatch(flowbind.DispatchRecord{
		Agent: flowbind.DispatchAgent, Status: flowbind.DispatchAccepted, AutoLevel: flowbind.DispatchAutoLevel,
		Action: "approved", Kind: flowbind.DispatchKindProposal,
		Scope: scope,
		Proposal: name, Title: spec.Title, Summary: spec.Summary,
		Acceptance: spec.Acceptance, Priority: spec.Priority, Proposer: spec.Proposer,
		RV: rv, EventID: eventID, At: time.Now(),
	})
}
