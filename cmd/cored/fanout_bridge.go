// fanout_bridge.go —— 批量 fan-out 执行桥（2026-11-28 用户令「issuse是一批的。judge判断哪些
// 要下发。然后启动多个文档apply…agent继续下发agent任务。走多链条。最后交叉汇总」——
// fanout.go 第四词形的装配半边）。
//
// 链路：run 落 unit=workflow ∧ fanout 节点 → 引擎挂起（suspendWhy=fanout）→ 本桥
// LaunchChildRuns：读执行计划（batch-judge 产物——run 上下文账 planItems 优先，
// miss=WorkflowPlan 实体 plan-run<N> 回读——重启丢账仍可扇出）→ 每项一子 run
// （entry:sub:plan-item 锚切入子链 Planning；involved=项溯源锚 Repository/<repo>#<n>）；
// 子 run 建成即 ChildStarted 绑定计划项到子 run 上下文账（planItem 键——子链 apply-doc/
// developer-system 会话的输入实体）。计划零项=FanoutZero 显式放行（父 run 不挂死）。
//
// 边界：本桥只扇出不执行（子链执行=inner runner/service 桥/connectops 桥各自执行面）；
// fix 与 defer 项都扇出——defer 项在子链内走 defer 判断边直达回评（不做也有反馈），
// 扇出面不裁项（裁项=judge 的权，本桥只读计划）。
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
)

// fanItem 执行计划项（WorkflowPlan spec.items 词形——与 approvaldiff.planItem 同源 JSON 键，
// 本桥独立词形零跨包依赖）。
type fanItem struct {
	Key      string `json:"key"`
	Title    string `json:"title,omitempty"`
	Action   string `json:"action,omitempty"`
	Reason   string `json:"reason,omitempty"`
	Risk     string `json:"risk,omitempty"`
	Involved string `json:"involved,omitempty"`
}

// fanoutBridge 扇出执行桥（workflow.FanoutLauncher + workflow.ChildBindLauncher 双能力）。
type fanoutBridge struct {
	runs *workflow.RunLog      // FanoutZero 放行面（构造后注入——解环同款）
	repo repository.Repository // WorkflowPlan 实体回读（重启丢账后的读源）
	memo *workflow.RunMemo     // run 上下文账（planItems 现读 + 子 run planItem 绑定写）
	logf func(string, ...any)
}

// LaunchChildRuns 实现 workflow.FanoutLauncher（非阻塞契约——引擎锁外派发 goroutine 调用）。
func (b *fanoutBridge) LaunchChildRuns(run workflow.RunRecord, node workflow.NodeView) []workflow.ChildRunSpec {
	logf := b.logf
	if logf == nil {
		logf = func(string, ...any) {}
	}
	items := b.planItems(run)
	if len(items) == 0 {
		logf("fanout 桥计划零项（run=%d——judge 未产出可扇出计划，父 run FanoutZero 放行）", run.Run)
		return nil
	}
	workflowName := ""
	entry := "sub:plan-item"
	if node.Fanout != nil {
		if node.Fanout.Workflow != "" {
			workflowName = node.Fanout.Workflow
		}
		if node.Fanout.Entry != "" {
			entry = node.Fanout.Entry
		}
	}
	specs := make([]workflow.ChildRunSpec, 0, len(items))
	for _, it := range items {
		specs = append(specs, workflow.ChildRunSpec{
			Key:      it.Key,
			Workflow: workflowName, // 空=父 workflow（fanout.go StartChildRun 缺省口径）
			Type:     entry,
			Involved: it.Involved, // 空=继承父 run involved
			Payload:  itemPayload(it),
		})
	}
	logf("fanout 桥扇出 %d 项（run=%d workflow=%s entry=%s fix=%d defer=%d）", len(specs), run.Run, workflowName, entry, countFan(items, "fix"), countFan(items, "defer"))
	return specs
}

// ChildStarted 实现 workflow.ChildBindLauncher：计划项→子 run 上下文绑定（planItem 键——
// apply-doc/developer briefing 的读源；子 run 序号此时已定，绑定零竞态）。
func (b *fanoutBridge) ChildStarted(parentRun, childRun int64, item workflow.ChildRunSpec) {
	if b.memo == nil {
		return
	}
	m := item.Payload
	if m == nil {
		m = map[string]any{}
	}
	m["key"] = item.Key // 键随绑定补全（payload 缺 key 时子链仍可判）
	b.memo.Set(childRun, "planItem", m)
	b.memo.Set(childRun, "parentRun", parentRun)
}

// planItems 执行计划读取（run 上下文账 planItems 优先——judge 同进程刚产出；miss=
// WorkflowPlan 实体 plan-run<N> 回读——重启丢账后 launcher 独立成扇）。
func (b *fanoutBridge) planItems(run workflow.RunRecord) []fanItem {
	if b.memo != nil {
		if v, ok := b.memo.Get(run.Run, "planItems"); ok {
			if items := decodeItems(v); len(items) > 0 {
				return items
			}
		}
	}
	if b.repo == nil {
		return nil
	}
	o, err := b.repo.Get(context.Background(), "WorkflowPlan", "global", fmt.Sprintf("plan-run%d", run.Run))
	if err != nil || o == nil {
		return nil
	}
	return parsePlanItems(string(o.Content))
}

// decodeItems 上下文账值 → 计划项（JSON 往返——approvaldiff.planItem 词形独立化）。
func decodeItems(v any) []fanItem {
	b, err := json.Marshal(v)
	if err != nil {
		return nil
	}
	var items []fanItem
	if json.Unmarshal(b, &items) != nil {
		return nil
	}
	return items
}

// parsePlanItems 计划实体原文 → 计划项（items 段 "- {json}" 行——approvaldiff.planDoc 写形回读）。
func parsePlanItems(content string) []fanItem {
	var items []fanItem
	inItems := false
	for _, line := range strings.Split(content, "\n") {
		s := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(s, "items:"):
			inItems = true
		case inItems && strings.HasPrefix(s, "- {") && strings.HasSuffix(s, "}"):
			var it fanItem
			if json.Unmarshal([]byte(s[2:]), &it) == nil && it.Key != "" {
				items = append(items, it)
			}
		}
	}
	return items
}

// itemPayload 计划项 → 子 run 载荷（透传留痕——map 词形供上下文绑定与查询面）。
func itemPayload(it fanItem) map[string]any {
	b, err := json.Marshal(it)
	if err != nil {
		return map[string]any{"key": it.Key}
	}
	m := map[string]any{}
	if json.Unmarshal(b, &m) != nil {
		return map[string]any{"key": it.Key}
	}
	return m
}

// countFan 动作计数（日志词面）。
func countFan(items []fanItem, action string) int {
	n := 0
	for _, it := range items {
		if it.Action == action {
			n++
		}
	}
	return n
}
