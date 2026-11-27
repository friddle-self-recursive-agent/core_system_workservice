// Package workflow 执行态机 + 触发规则（v0.8 吸收 Loop→Workflow，A41；A67 全局去 loop）。
// Workflow = 触发规则→事件整理→事件处理（相位态机）→文档回归；
// 相位/迁移纯声明，预算/成本不在核心循环语义内（budget 是 metrics 关注的东西，暂不做）。
// TaskScheme 管 Executing 相位内的步骤图。
// builtin 六项（scheme.yaml runtime.workflows.builtin 同源）：core-workflow（系统级循环）、
// default-task（任务态机）、simple-shot、watch-only、dev-delivery（研发交付 §1.1）、
// apply-workflow（workflow 落地 agent 任务态机 §1.2）——相位词汇映射表见 concepts/workflow.md §1。
// trigger.go：触发骨架（dispatch 接线第一段 v0.11）——TriggerHub 单订阅者收到事件落可查状态
// （apply→事件→订阅主干第一段），触发引擎主体（dispatch 路由→owner 解析→相位态机→派发→刹车）
// declared-deferred（cored-bridge.md §2.3）。
package workflow

import (
	"sort"
	"time"
)

// Phase 相位。
type Phase string

const (
	Idle             Phase = "Idle"
	Analyzing        Phase = "Analyzing"
	Planning         Phase = "Planning"
	AwaitingApproval Phase = "AwaitingApproval"
	Executing        Phase = "Executing"
	Verifying        Phase = "Verifying"
	Deploying        Phase = "Deploying" // dev-delivery 交付相位（v0.19 相位词汇只加不改：Deploy 步=deploy-system 按 envs 升序发布，workflow.md §1.1）
	Reporting        Phase = "Reporting"
	Failed           Phase = "Failed"
)

// Transition 相位迁移（纯声明：from + on 事件 → to）。
type Transition struct {
	From Phase
	On   string // 事件名（Scheme 注册）
	To   Phase
}

// OnFailed 失败出路：backoff → 超限 Escalate（root / role 持有者双通道）→ Idle。
type OnFailed struct {
	Backoff     time.Duration
	MaxAttempts int
	EscalateTo  string // root（默认）或 role 引用
}

// Spec Workflow 声明。
type Spec struct {
	Name        string
	Phases      []Phase
	Transitions []Transition
	OnFailed    OnFailed
}

// Triggers 触发规则（与事件 Level 联动：minLevel 过滤）。
type Triggers struct {
	Events   []string      // 事件类型列表（路由按类型+owner，v0.11 A47 无 selector）
	MinLevel string        // 事件 Level 过滤（A42：debug/info/warn/error/fatal）
	Resync   time.Duration // 巡检兜底（level-triggered 保险丝）
}

// Brakes 刹车三件套（Workflow 级强制默认：可收紧不可关闭）。
type Brakes struct {
	Cooldown         time.Duration // 成功后最小间隔（防震荡）
	MaxMutationsTick int           // 单轮最大变更数
	AvoidRepeatFix   bool          // DriftReport 记录"改过什么"，避免重复同类修复
}

// FullSpec 完整 Workflow（触发+态机+刹车+回归）。
// 成本观测归 metrics（ConnectorOps 收集→agent 报告），核心循环不带预算机器。
type FullSpec struct {
	Spec
	Triggers Triggers
	Brakes   Brakes
}

// Builtin 默认态机（core_system/concepts/workflow.md §1 相位词汇映射表——六项 builtin
// 的声明相位序列，声明面与投影面同源；相位语义只加不改，workflow.md §4）。
// v0.25 显式迁移表首段：builtin 表随构造物化（序列物化相邻行+审批边表行先例，
// MaterializeTransitions），判定器查表不再读序列——Can 与 run 域迁移判定同源。
func Builtin(name string) *Spec {
	switch name {
	case "simple-shot":
		return builtinSpec(name, []Phase{Idle, Executing, Reporting})
	case "watch-only":
		return builtinSpec(name, []Phase{Idle, Analyzing, Reporting})
	case "bug-lifecycle":
		// §9a 缺陷闭环（2026-11-26 F-04 随行落注册表：scheme builtinSet 七项同源——
		// BugReport→临时 BugDoc→验证→关闭；闸锚=人工确认修复方案）。
		return builtinSpec(name, []Phase{Idle, Analyzing, AwaitingApproval, Executing, Verifying, Reporting})
	case "dev-delivery":
		return builtinSpec(name, []Phase{ // §1.1 步锚映射：Analyze→Gate→Plan→Execute→Verify→Deploy→Report
			Idle, Analyzing, AwaitingApproval, Planning, Executing, Verifying, Deploying, Reporting,
		})
	case "core-system-developer-loop":
		// 2026-11-27 用户裁决：核心内置研发自循环主干（监听日志Event→Agent判断→审批/自动审批→
		// Doc Apply→developer-system→验收→deploy-system→总结归档+回写InSync），支持多节点触发切入
		// （如用户对话修改文档或产出临时 BugDoc 触发 DocApplied 直接从 Planning / doc-apply 节点切入）。
		spec := builtinSpec(name, []Phase{
			Idle, Analyzing, AwaitingApproval, Planning, Executing, Verifying, Deploying, Reporting,
		})
		spec.Transitions = append(spec.Transitions,
			Transition{From: Idle, On: "entry:DocApplied", To: Planning},
			Transition{From: Analyzing, On: "worth_fixing", To: AwaitingApproval},
			Transition{From: Analyzing, On: "ignore", To: Reporting},
			Transition{From: Verifying, On: "retry-develop", To: Executing},
		)
		return spec
	case "apply-workflow":
		return builtinSpec(name, []Phase{ // §1.2 落地 agent 任务态机=default-task 词汇全量复用（不另立词汇）
			Idle, Analyzing, Planning, AwaitingApproval, Executing, Verifying, Reporting,
		})
	case "core-workflow":
		return builtinSpec(name, []Phase{ // 系统级循环（A41；A67 core-loop→core-workflow）
			Idle, Analyzing, Planning, AwaitingApproval, Executing, Verifying, Reporting,
		})
	default: // default-task
		return builtinSpec("default-task", []Phase{
			Idle, Analyzing, Planning, AwaitingApproval, Executing, Verifying, Reporting,
		})
	}
}

// builtinSpec 构造 builtin spec：相位序列 + 声明表物化（workflow.md 规则 14——表=序列
// 物化相邻行+审批边表行先例；判定器查表，行为面与序列前进等价）。
func builtinSpec(name string, seq []Phase) *Spec {
	return &Spec{Name: name, Phases: seq, Transitions: MaterializeTransitions(seq)}
}

// EdgeAdvance 推进事件名（声明表物化相邻行的 on 词——序列前进=表的特例，workflow.md 规则 14）。
const EdgeAdvance = "advance"

// MaterializeTransitions builtin 声明表物化（v0.25 显式迁移表首段，workflow.md 规则 14）。
// 纯函数：声明相位序列 → 迁移表行集。
//   - 相邻行：from=seq[i] on=advance to=seq[i+1]（序列前进=表的特例——表行全量物化）；
//   - 审批边表行先例（规则 13 挂起-恢复的表侧落点）：批准边 from=闸锚 on=ApprovalGranted
//     to=闸锚后继（与相邻行同目标、异事件名——恢复判定查表）；驳回边 from=闸锚
//     on=ApprovalRejected to=终点锚 Failed——**表的首处非相邻迁移行**（机械判别器：
//     非相邻迁移按表命中）。
//
// 2026-11-26 F-04 纠偏：驳回边**无条件物化**（不再要求 Failed 在声明序列内——正向序列
// 不再携带侧轨 Failed，`Failed` 是侧轨终态：仅经 ApprovalRejected 边可达；声明序列带不带
// Failed 词形均同表）。声明无闸锚的序列=无对应审批边（无挂起面/驳回零迁移的表侧落点——
// 零动作面在表物化即收口，判定器无需特判）。
func MaterializeTransitions(seq []Phase) []Transition {
	rows := make([]Transition, 0, len(seq)+2)
	for i := 0; i+1 < len(seq); i++ {
		rows = append(rows, Transition{From: seq[i], On: EdgeAdvance, To: seq[i+1]})
	}
	for i, p := range seq {
		if p != AwaitingApproval {
			continue
		}
		if i+1 < len(seq) {
			rows = append(rows, Transition{From: p, On: SignalResume, To: seq[i+1]})
		}
		rows = append(rows, Transition{From: p, On: SignalReject, To: Failed})
	}
	return rows
}

// Can 迁移合法性判定（v0.25 显式迁移表首段起实接：Builtin 物化声明表——查表命中
// from+on → to，未命中=零迁移；与 run 域判定器 stepByTable 同一查表口径，workflow.md 规则 14）。
func (s *Spec) Can(from Phase, on string) (Phase, bool) {
	for _, t := range s.Transitions {
		if t.From == from && t.On == on {
			return t.To, true
		}
	}
	return from, false
}

// builtinSet 内置 workflow 注册表（判定用；与 Builtin 构造的已声明相位集合同源——
// scheme.yaml runtime.workflows.builtin 七项全量对齐（v0.19：dev-delivery/apply-workflow
// 相位词汇映射已落设计，concepts/workflow.md §1 映射表——声明先行、注册表随后，不即兴发明；
// 2026-11-26 F-04 纠偏：声明序列=正向业务步序，Failed 不再列线性尾——侧轨语义随
// MaterializeTransitions 驳回边保留，见 workflow.go §Failed 侧轨注）。
var builtinSet = map[string]bool{
	"core-system-developer-loop": true,
	"core-workflow":              true,
	"default-task":               true,
	"simple-shot":                true,
	"watch-only":                 true,
	"dev-delivery":               true,
	"apply-workflow":             true,
	"bug-lifecycle":              true,
}

// IsBuiltin 名是否命中内置 workflow 注册表（run 锚解析的 SpecBuiltin/SpecObject 判据）。
func IsBuiltin(name string) bool { return builtinSet[name] }

// BuiltinPhaseNames 内置 workflow 的声明相位序列投影（v0.37 console 读面第二段导出面：
// `GET /v1/workflows` 聚合投影的 builtin 相位词汇与 run 投影同源——phaseNames 同一映射，
// 无双源；未命中=空）。
func BuiltinPhaseNames(name string) []string {
	if !IsBuiltin(name) {
		return []string{}
	}
	return phaseNames(Builtin(name).Phases)
}

// BuiltinNames 内置 workflow 注册表名清单（v0.37 console 读面第二段导出面：聚合面行集
// 三源并集的注册表半边——缺任何一源都不吞行；字典序确定性投影）。
func BuiltinNames() []string {
	out := make([]string, 0, len(builtinSet))
	for n := range builtinSet {
		out = append(out, n)
	}
	sort.Strings(out)
	return out
}
