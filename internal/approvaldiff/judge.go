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
	"io"
	"net/http"
	"os"
	"strings"
	"time"

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
	subject := g.readSubject(ctx, objKey) // 被审物（对象现值/原始 issue 正文）——判决与人审兜底共用同一份
	user := fmt.Sprintf("对象=%s run=%d 审批准则=%s 允许自动批准的最高风险=%s\n待批内容：\n%s",
		objKey, run.Run, criteria, maxRisk, clip(subject, 6000))
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
	content := approvalVerdictDoc(run, objKey, docName, policy, decision, v, subject)
	if _, _, err := g.Repo.Write(ctx, &repository.Object{Kind: "DocumentDiff", Scope: "global", Name: docName, Content: []byte(content)}, "workflow-engine", "approval-judge:run:"+fmt.Sprint(run.Run)); err != nil {
		return workflow.ApprovalVerdict{}, fmt.Errorf("approvaljudge: 判决文档落库失败: %w", err)
	}
	g.Logf("approvaljudge: run %d 判决=%s risk=%s 文档=%s 被审物=%d字", run.Run, decision, v.Risk, docName, len(subject))
	return v, nil
}

// readSubject 被审物读取（2026-11-28 用户裁定「原始 issue 是什么。我怎么没看到」）：判决与
// 人审兜底共用同一份「被审物」文本——Repository/<name>#<n> 词形（issue 目标）走 GitHub 公共
// REST 拉原始 title+body（零凭据可用；org 从 connectops-gh 对象 config.owner 解析，零硬编码）；
// 其余对象词形走对象库现值（readBase）。两路都空=显式「（被审物不可读）」占位，不冒充空串。
func (g *Generator) readSubject(ctx context.Context, objKey string) string {
	kind, scope, name := splitKind(objKey), splitScope(objKey), splitName(objKey)
	if kind == "Repository" && strings.Contains(name, "#") {
		if s := g.ghIssueSubject(ctx, name); s != "" {
			return s
		}
	}
	if g.Repo != nil {
		if s := readBase(g.Repo, ctx, kind, scope, name); s != "" {
			return s
		}
	}
	return "（被审物不可读——对象库无现值且非可拉取的 issue 词形）"
}

// ghOwner org 前缀解析：ConnectorOps/global/connectops-gh 的 config.owner（行级读，零 yaml
// 依赖——involved 名只带仓不带 org，前缀由此声明面补全）。
func (g *Generator) ghOwner(ctx context.Context) string {
	if g.Repo == nil {
		return ""
	}
	o, err := g.Repo.Get(ctx, "ConnectorOps", "global", "connectops-gh")
	if err != nil || o == nil {
		return ""
	}
	for _, line := range strings.Split(string(o.Content), "\n") {
		s := strings.TrimSpace(line)
		if strings.HasPrefix(s, "owner:") {
			return strings.TrimSpace(strings.TrimPrefix(s, "owner:"))
		}
	}
	return ""
}

// ghIssueSubject GitHub REST 拉 issue 原文（title+labels+body）——name#n 词形专用；仓无 org
// 前缀时以 ghOwner 补全（10s 超时；GH_TOKEN 存在则带认证头防限流）。
func (g *Generator) ghIssueSubject(ctx context.Context, name string) string {
	p := strings.SplitN(name, "#", 2)
	if len(p) != 2 || p[0] == "" || p[1] == "" {
		return ""
	}
	repo := p[0]
	if !strings.Contains(repo, "/") {
		if owner := g.ghOwner(ctx); owner != "" {
			repo = owner + "/" + repo
		}
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues/%s", repo, p[1])
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return ""
	}
	if tok := strings.TrimSpace(os.Getenv("GH_TOKEN")); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	var d struct {
		Title  string `json:"title"`
		Body   string `json:"body"`
		Labels []struct {
			Name string `json:"name"`
		} `json:"labels"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&d); err != nil {
		return ""
	}
	labels := make([]string, 0, len(d.Labels))
	for _, l := range d.Labels {
		labels = append(labels, l.Name)
	}
	return fmt.Sprintf("标题: %s\n标签: %s\n正文:\n%s", d.Title, strings.Join(labels, ","), d.Body)
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

func approvalVerdictDoc(run workflow.RunRecord, objKey, docName string, policy workflow.ApprovalPolicy, decision string, v workflow.ApprovalVerdict, subject string) string {
	return strings.Join([]string{
		"apiVersion: core/v1", "kind: DocumentDiff", "metadata:", "  name: " + docName,
		"  scope: global", "  labels: {core/origin: workflow-engine}", "spec:",
		"  target: " + quoteYAML(objKey), "  run: " + fmt.Sprint(run.Run),
		"  agent: ops_core_inner_agent", "  provider: " + quoteYAML(policy.DiffProvider),
		"  criteria: " + quoteYAML(policy.AutoCriteria), "  maxRiskLevel: " + quoteYAML(policy.MaxRiskLevel),
		"  decision: " + decision, "  risk: " + v.Risk, "  reason: " + quoteYAML(v.Reason),
		"  verdict: |", indentBlock(v.Reason, "    "),
		"  subject: |", indentBlock(clip(subject, 1500), "    "), // 被审物原文（人审兜底「我批准啥」的信息面——2026-11-28 用户裁定）
	}, "\n")
}

// —— objKey 三段投影（splitObjKey 的容错版：缺席段=空串，不阻断判决留痕）——
func splitKind(k string) string  { p := strings.Split(k, "/"); if len(p) > 0 { return p[0] }; return "" }
func splitScope(k string) string { p := strings.Split(k, "/"); if len(p) > 1 { return p[1] }; return "" }
func splitName(k string) string  { p := strings.Split(k, "/"); if len(p) > 2 { return p[2] }; return "" }

var _ workflow.ApprovalJudge = (*Generator)(nil)
