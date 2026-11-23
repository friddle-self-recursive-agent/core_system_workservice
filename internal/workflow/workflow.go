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

import "time"

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
func Builtin(name string) *Spec {
	switch name {
	case "simple-shot":
		return &Spec{Name: name, Phases: []Phase{Idle, Executing, Reporting}}
	case "watch-only":
		return &Spec{Name: name, Phases: []Phase{Idle, Analyzing, Reporting}}
	case "dev-delivery":
		return &Spec{Name: name, Phases: []Phase{ // §1.1 步锚映射：Analyze→Gate→Plan→Execute→Verify→Deploy→Report
			Idle, Analyzing, AwaitingApproval, Planning, Executing, Verifying, Deploying, Reporting, Failed,
		}}
	case "apply-workflow":
		return &Spec{Name: name, Phases: []Phase{ // §1.2 落地 agent 任务态机=default-task 词汇全量复用（不另立词汇）
			Idle, Analyzing, Planning, AwaitingApproval, Executing, Verifying, Reporting, Failed,
		}}
	case "core-workflow":
		return &Spec{Name: name, Phases: []Phase{ // 系统级循环（A41；A67 core-loop→core-workflow）
			Idle, Analyzing, Planning, AwaitingApproval, Executing, Verifying, Reporting, Failed,
		}}
	default: // default-task
		return &Spec{Name: "default-task", Phases: []Phase{
			Idle, Analyzing, Planning, AwaitingApproval, Executing, Verifying, Reporting, Failed,
		}}
	}
}

// Can 迁移合法性判定。
func (s *Spec) Can(from Phase, on string) (Phase, bool) {
	for _, t := range s.Transitions {
		if t.From == from && t.On == on {
			return t.To, true
		}
	}
	return from, false
}

// builtinSet 内置 workflow 注册表（判定用；与 Builtin 构造的已声明相位集合同源——
// scheme.yaml runtime.workflows.builtin 六项全量对齐（v0.19：dev-delivery/apply-workflow
// 相位词汇映射已落设计，concepts/workflow.md §1 映射表——声明先行、注册表随后，不即兴发明）。
var builtinSet = map[string]bool{
	"core-workflow":  true,
	"default-task":   true,
	"simple-shot":    true,
	"watch-only":     true,
	"dev-delivery":   true,
	"apply-workflow": true,
}

// IsBuiltin 名是否命中内置 workflow 注册表（run 锚解析的 SpecBuiltin/SpecObject 判据）。
func IsBuiltin(name string) bool { return builtinSet[name] }
