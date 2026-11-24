// Package proposal —— 需求提案分流/派单面（v0.51，能力 B——core_design concepts/proposal.md）。
// 设计依据：concepts/proposal.md（一等对象：console 提需求→审批→派工）+ event.md 规则 15
// （引擎内部源第五段=Proposal 审批分流 ApprovalNeeded 自产）+ workflow-runtime.md §11.3
// （派单第二来源）+ instance 仓 runtime/console-adoption.md §1.24。
//
// 三件事（零新写路径零旁路——全部写动作走 apply.Engine 同一三步闸门）：
//   - Parse/FlipStatus：Proposal 对象 yaml 的宽松解析与状态翻转（读面不放大声明形变——写路径
//     scheme 校验是形态权威；FlipStatus=spec.status 行的机械翻转，供引擎自写 auto-approve 用）。
//   - Watcher：DocApplied×Proposal 订阅（flowbind.Materializer 热路径同款范式）——
//     status=submitted ∧ 高危（priority 不在 AutoApprove 清单）→ 自产 ApprovalNeeded（source=core，
//     dedupKey=proposal:<obj>，route notification/* 送达+审批屏待审区随行）；
//     status=submitted ∧ 低危命中清单 → 引擎自写批准（fieldManager=proposal:auto，approval_log
//     granted 行 judgedBy=引擎主体）——落库的 DocApplied 再入本 watcher → 派单（单环两拍，无自激：
//     approved 不再写面）；status=approved → 一条对 apply-workflow 锚的派单受理投影（flowbind
//     DispatchSink 复用——同 face GET /v1/agents/dispatches，kind=proposal 行）；rejected/landed/draft
//     零动作（rejected 不派单——proposal.md 规则 2）。
//   - 装配解耦：sink/logger 由装配层注入（nil=静默零派单/零落账——headless 载体/旧装配零影响）；
//     事件发布注册闸门前置（ApprovalNeeded 未注册即不发布——注册表是引擎内部源唯一开关）。
//
// 红线（proposal.md §2）：本包不 enable 任何 workflow、不自动执行代码——批准只产派单**受理投影**，
// 真实拉起=dsh 会话挂载（agent-session 域，declared-deferred）。
package proposal

import (
	"bytes"
	"regexp"

	"gopkg.in/yaml.v3"
)

// 状态闭合集与词汇（scheme.yaml Proposal spec.status 声明同源——proposal.md §4）。
const (
	StatusDraft     = "draft"
	StatusSubmitted = "submitted"
	StatusApproved  = "approved"
	StatusRejected  = "rejected"
	StatusLanded    = "landed"
)

// 引擎自写溯源串（apply.md rule 2；proposal: 前缀=非 console 域，判定面可判别）。
// FieldManagerAuto=低危自动批准（分流命中——部署侧 --proposal-auto-approve）。
const (
	FieldManagerAuto = "proposal:auto"
	ReasonRefAuto    = "proposal:auto-approve"

	approvalSourceCore = "core" // 引擎自身产源标注（event.md 规则 15——与巡检恒 core 同词）
)

// Spec Proposal 对象 spec 的宽松解析投影（读面不放大声明形变）。
type Spec struct {
	Title         string   // 标题
	Summary       string   // 目标
	Acceptance    []string // 验收标准
	Impact        string   // 影响面
	Priority      string   // 优先级（low|normal|high|critical）
	Proposer      string   // 提出人
	Status        string   // 状态闭合集
	LandingTarget string   // 产物指针 target（doc|task）
	LandingRef    string   // 产物指针 ref（可选）
}

// propDecl 解析宿主（宽松：字段缺席/形变=零值，不报错——写路径校验是形态权威）。
type propDecl struct {
	Spec struct {
		Title      string   `yaml:"title"`
		Summary    string   `yaml:"summary"`
		Acceptance []string `yaml:"acceptance"`
		Impact     string   `yaml:"impact"`
		Priority   string   `yaml:"priority"`
		Proposer   string   `yaml:"proposer"`
		Status     string   `yaml:"status"`
		Landing    struct {
			Target string `yaml:"target"`
			Ref    string `yaml:"ref"`
		} `yaml:"landing"`
	} `yaml:"spec"`
}

// Parse Proposal 对象原文 → spec 投影（不可解析/无状态=ok=false，调用方跳过——不放大）。
func Parse(content []byte) (Spec, bool) {
	var d propDecl
	if err := yaml.Unmarshal(content, &d); err != nil {
		return Spec{}, false
	}
	s := Spec{
		Title: d.Spec.Title, Summary: d.Spec.Summary, Acceptance: d.Spec.Acceptance,
		Impact: d.Spec.Impact, Priority: d.Spec.Priority, Proposer: d.Spec.Proposer,
		Status: d.Spec.Status, LandingTarget: d.Spec.Landing.Target, LandingRef: d.Spec.Landing.Ref,
	}
	if s.Status == "" {
		return s, false // 无状态=非 Proposal 形（枚举闭合集必填——声明面权威）
	}
	return s, true
}

// FlipStatus 对象原文的 spec.status 机械翻转（行锚定=缩进+status:+from 精确词——不碰散文/注释；
// 零命中=ok=false：形态漂移交 stderr/审计，不静默冒充成功）。
func FlipStatus(content []byte, from, to string) ([]byte, bool) {
	if from == to || len(content) == 0 {
		return nil, false
	}
	re := regexp.MustCompile(`(?m)^(\s+status:\s*)` + regexp.QuoteMeta(from) + `(\s*)$`)
	out := re.ReplaceAll(content, []byte("${1}"+to+"${2}"))
	if bytes.Equal(out, content) {
		return nil, false
	}
	return out, true
}
