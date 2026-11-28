// judge.go —— 自审批真判决实现（2026-11-28 用户裁定「自审批未接线」修复 / workservice #4）。
//
// ops_core_inner_agent 在闸锚挂起窗口内直调 Provider（taskModelMapping["approval-judge"]
// 真名优先——网关不认 tier 词形），按 spec.approval.autoCriteria 产出机器判决
// {granted,risk,reason} + 人话判决 DocumentDiff（inneragent-approval-run<N>，同名幂等）。
// 失败/不可解析=风险 high 失败收敛（引擎回退人审，绝不编造批准）。
package approvaldiff

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
)

// JudgeApproval 实现 workflow.ApprovalJudge（编译期断言见文件末）。
func (g *Generator) JudgeApproval(run workflow.RunRecord, policy workflow.ApprovalPolicy, objKey string) (workflow.ApprovalVerdict, error) {
	docName := approvalJudgeDocName(run.Run)
	ctx := context.Background()
	if g == nil || g.Repo == nil {
		return workflow.ApprovalVerdict{}, fmt.Errorf("approvaljudge: 未挂文档库（Repo 缺席）")
	}
	// 同名幂等：重启/重复挂起读回既有判决，不重复调 LLM（幂等域=run）。
	if o, err := g.Repo.Get(ctx, "DocumentDiff", "global", docName); err == nil && o != nil {
		if v, ok := parseVerdictDoc(string(o.Content)); ok {
			v.DocName = docName
			return v, nil
		}
	}
	cfg, err := g.provider()
	if err != nil {
		return workflow.ApprovalVerdict{}, fmt.Errorf("approvaljudge: Provider 解析失败: %w", err)
	}
	model := cfg.Model
	if m, ok := g.ModelOverride("approval-judge"); ok {
		model = m // taskModelMapping 真名优先（ops_core_inner_agent 通用兜底见 provider()）
	}
	criteria := strings.TrimSpace(policy.AutoCriteria)
	if criteria == "" {
		criteria = "仅当变更局限文档或非破坏性代码且测试全绿时自动批准；涉及删库/权限/生产配置拒绝并转人工"
	}
	maxRisk := strings.TrimSpace(policy.MaxRiskLevel)
	if maxRisk == "" {
		maxRisk = "low" // scheme approvalGate 强制默认
	}
	sys := "你是审批判决小 agent（ops_core_inner_agent）。严格按给定准则判断该审批能否自动通过；只依据输入文本，不编造。只输出 JSON 对象，字段 granted(bool)/risk(low|medium|high)/reason(中文≤200字)，不要 Markdown 包裹。"
	user := fmt.Sprintf("对象=%s run=%d 审批准则=%s 允许自动批准的最高风险=%s\n待批内容（现值）：\n%s",
		objKey, run.Run, criteria, maxRisk, clip(readBase(g.Repo, ctx, splitKind(objKey), splitScope(objKey), splitName(objKey)), 6000))
	raw := g.complete(ctx, cfg, model, sys, user)
	v, perr := parseVerdictJSON(raw)
	if perr != nil {
		// 判决不可解析=失败收敛（风险 high 让引擎回退人审，产物仍落库留痕）。
		v = workflow.ApprovalVerdict{Risk: "high", Reason: "（LLM 判决不可解析——" + perr.Error() + "）"}
	}
	v.DocName = docName
	decision := "approve"
	if !v.Granted {
		decision = "reject"
	}
	if workflow.RiskExceeds(v.Risk, maxRisk) {
		decision = "escalate"
	}
	content := approvalVerdictDoc(run, objKey, docName, policy, decision, v)
	if _, _, err := g.Repo.Write(ctx, &repository.Object{Kind: "DocumentDiff", Scope: "global", Name: docName, Content: []byte(content)}, "workflow-engine", "approval-judge:run:"+fmt.Sprint(run.Run)); err != nil {
		return workflow.ApprovalVerdict{}, fmt.Errorf("approvaljudge: 判决文档落库失败: %w", err)
	}
	g.Logf("approvaljudge: run %d 判决=%s risk=%s 文档=%s", run.Run, decision, v.Risk, docName)
	return v, nil
}

// parseVerdictJSON 严格 JSON 判决解析（容忍 ```json 围栏与前后散文：取首个 { 到末个 }）。
func parseVerdictJSON(raw string) (workflow.ApprovalVerdict, error) {
	s := strings.TrimSpace(raw)
	if i := strings.Index(s, "{"); i >= 0 {
		s = s[i:]
	}
	if i := strings.LastIndex(s, "}"); i >= 0 {
		s = s[:i+1]
	}
	var d struct {
		Granted *bool  `json:"granted"`
		Risk    string `json:"risk"`
		Reason  string `json:"reason"`
	}
	if err := json.Unmarshal([]byte(s), &d); err != nil {
		return workflow.ApprovalVerdict{}, err
	}
	if d.Granted == nil {
		return workflow.ApprovalVerdict{}, fmt.Errorf("判决缺 granted 字段")
	}
	risk := strings.ToLower(strings.TrimSpace(d.Risk))
	if risk != "low" && risk != "medium" && risk != "high" {
		risk = "high" // 词形非法=失败收敛
	}
	return workflow.ApprovalVerdict{Granted: *d.Granted, Risk: risk, Reason: strings.TrimSpace(d.Reason)}, nil
}

// parseVerdictDoc 幂等回读（从自身落库的 DocumentDiff 还原机器判决词形）。
func parseVerdictDoc(content string) (workflow.ApprovalVerdict, bool) {
	var v workflow.ApprovalVerdict
	found := false
	for _, line := range strings.Split(content, "\n") {
		s := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(s, "decision:"):
			d := strings.TrimSpace(strings.TrimPrefix(s, "decision:"))
			if d == "approve" || d == "reject" || d == "escalate" {
				v.Granted = d == "approve"
				found = true
			}
		case strings.HasPrefix(s, "risk:"):
			v.Risk = strings.TrimSpace(strings.TrimPrefix(s, "risk:"))
		case strings.HasPrefix(s, "reason:"):
			v.Reason = strings.Trim(strings.TrimSpace(strings.TrimPrefix(s, "reason:")), "\"")
		}
	}
	return v, found && v.Risk != ""
}

func approvalVerdictDoc(run workflow.RunRecord, objKey, docName string, policy workflow.ApprovalPolicy, decision string, v workflow.ApprovalVerdict) string {
	return strings.Join([]string{
		"apiVersion: core/v1", "kind: DocumentDiff", "metadata:", "  name: " + docName,
		"  scope: global", "  labels: {core/origin: workflow-engine}", "spec:",
		"  target: " + quoteYAML(objKey), "  run: " + fmt.Sprint(run.Run),
		"  agent: ops_core_inner_agent", "  provider: " + quoteYAML(policy.DiffProvider),
		"  criteria: " + quoteYAML(policy.AutoCriteria), "  maxRiskLevel: " + quoteYAML(policy.MaxRiskLevel),
		"  decision: " + decision, "  risk: " + v.Risk, "  reason: " + quoteYAML(v.Reason),
		"  verdict: |", indentBlock(v.Reason, "    "),
	}, "\n")
}

// —— objKey 三段投影（splitObjKey 的容错版：缺席段=空串，不阻断判决留痕）——
func splitKind(k string) string  { p := strings.Split(k, "/"); if len(p) > 0 { return p[0] }; return "" }
func splitScope(k string) string { p := strings.Split(k, "/"); if len(p) > 1 { return p[1] }; return "" }
func splitName(k string) string  { p := strings.Split(k, "/"); if len(p) > 2 { return p[2] }; return "" }

var _ workflow.ApprovalJudge = (*Generator)(nil)
