// verifygate.go —— Verify 相位硬 gate（G-09/G-12，R-TYPE-3 rule_type.md 三段链接线——
// rule/rule_type.md 规则 4「Verify 节点必备」与 perf-lint-baseline.md §3.5「connectops-code-audit
// 在 Verify 相位强制执行并作为出报告前置门」的引擎侧落点）。
//
// 设计：gate=workflow 域可选能力（RunOptions.VerifyGate，nil=既有行为面零翻转——能力增量
// 不破既有推进语义）；装配层（cored）挂接真实 gate（connectops-code-audit 连接器，
// connectors/codeaudit 包——注册表 runtime.ruleChecks 三行投影为 audit checks 输入；被审
// 目录=--verify-target 声明，缺席=gate 缺席且 stderr 缺口可见——不静默不冒充）。
// 判定时机=推进拍 run 试图离开 Verifying 相位（Verifying→后继 advance 行命中）：
//   - gate 判定 OK（无 error 级 findings ∧ unbound 空——codeaudit.Result.OK 同口径）→ 放行，
//     判定投影随 run 留痕（VerifyJSON 列写穿透，恢复取最新词形——卡 Verify 证据重启仍在）；
//   - findings 非空（error 级）→ **卡 Verify 不放行**（run 相位/状态原样保留，本拍零迁移，
//     gated 计数可见）+ Escalate 上报（RunOptions.OnVerifyBlocked 装配钩子——cored 发布
//     Escalate 事件进总线，走既有注册类型 root 通道；每拍重审=修复后下拍即过，审计只读幂等）。
//
// 与审批闸锚分立：AwaitingApproval 挂起=等人（外部信号恢复）；卡 Verify=等代码过审
// （checker 重审恢复）——两者都是相位机的挂断面，但恢复信号源不同（审批事件 vs 重审通过），
// 词汇不混用（卡 Verify 不翻 suspended——诚实投影：非审批挂起）。
package workflow

// VerifyVerdict Verify 相位硬 gate 的一次判定投影（R-TYPE-3 三段链回锚：auditId 可回查
// connectops-code-audit findings op；note=卡 Verify 摘要——Escalate 事件与人读面共用词形）。
type VerifyVerdict struct {
	OK            bool   `json:"ok"`                      // 放行判定（无 error findings ∧ unbound 空）
	AuditID       string `json:"auditId,omitempty"`       // 审核结果句柄（findings 回查锚）
	ErrorFindings int    `json:"errorFindings,omitempty"` // error 级 findings 数（非空=卡）
	WarnFindings  int    `json:"warnFindings,omitempty"`  // warn 级 findings 数（不翻转判定——进 Report）
	Unbound       int    `json:"unbound,omitempty"`       // 未绑定数（非空恒卡——unbound=未生效不是跳过）
	Note          string `json:"note,omitempty"`          // 判定摘要（放行="audit ok"；卡=违规摘要）
}

// VerifyGate Verify 相位硬 gate 口面（R-TYPE-3 接线）。实现须承载真实审核执行
// （connectops-code-audit audit op——只读）；判定在推进拍锁外调用（审核执行不持 run 锁——
// 订阅主干/查询面不被 checker 时长阻塞）；返回判定即迁移闸：OK=false=卡 Verify。
type VerifyGate interface {
	Verify(run RunRecord) VerifyVerdict
}

// VerifyGateFunc 函数形态适配器（装配侧闭包注入——cored 用 codeaudit 连接器包一行）。
type VerifyGateFunc func(run RunRecord) VerifyVerdict

// Verify 实现 VerifyGate（函数即 gate）。
func (f VerifyGateFunc) Verify(run RunRecord) VerifyVerdict { return f(run) }

// 编译期断言：函数形态满足 gate 口面。
var _ VerifyGate = VerifyGateFunc(nil)
