// runner.go —— 内置轻量 agent 执行单元（2026-11-28 用户裁定「doc-apply 的审批以及部分服务
// 运行起来可以使用内置的 agent 来做润滑剂。即低智商要求的可以使用内置 agent」）：
// executor=inner 的 agent/service 节点由本 Runner 执行——引擎内直调 Provider（节点 modelTier
// 或 taskModelMapping[op] 真名），产物落库，完成信号由引擎 InnerDone 续行。重型通路
// （executor=dsh 缺省）不经此处。
//
// op 化（2026-11-28 批量 fan-out 用户令「judge判断哪些要下发…最后交叉汇总」——通用一句话
// prompt 升级为按 op 语义的真数据面，通用形兜底不变）：
//   - judge-issues-batch：读 issue 批（run 上下文账 op:issue-list 全文，miss=GitHub REST
//     现拉兜底）→ 产出执行计划 JSON → **WorkflowPlan 实体落库**（fanout launcher 的读源，
//     重启可回读）+ run 上下文账回填 → decision 词（worth_fixing/ignore）驱动分叉边；
//   - apply-doc（子链）：读子 run 计划项（launcher ChildStarted 绑定）→ action=defer 直接
//     落稿+decision=defer（「不做也有反馈」的子链半边——回评链解释为什么不做）；fix 项落稿
//     修复要点（developer-system 会话的输入实体）；
//   - merge-results（父链 join 后）：读子 run 终态（Runs 现查）+计划 → 交叉汇总落稿 →
//     decision=release（大规模批完成才发版——默认空词=机械 advance 收束跳过发版）。
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

// Runner 内置轻量 agent 执行器（workflow.InnerAgentRunner 契约）。
type Runner struct {
	G    *Generator            // 复用 Provider 解析+LLM 通路（同账号包/secret/重试）
	Runs *workflow.RunLog      // 子 run 终态现查（merge-results 读面——装配层注入；nil=汇总缺子态）
	Memo *workflow.RunMemo     // run 上下文账（op 结果全文/计划项/结论——装配层注入；nil=只用落库实体）
	Repo repository.Repository // WorkflowPlan/DocumentDiff 落库（优先 G.Repo——nil 时兜底同源）
}

// RunInner 执行一个 inner 节点：按 op 语义完成节点任务（judge/apply-doc/merge-results 各自
// 真数据面；其余 op=通用形）。返回 decision 判断词（空=机械 advance 兜底）。err 非 nil=引擎留挂起。
func (r *Runner) RunInner(run workflow.RunRecord, node workflow.NodeView) (string, error) {
	r.G.trace("inner enter run=" + fmt.Sprint(run.Run) + " node=" + node.Name + " op=" + node.Op)
	if r.G == nil || r.G.Repo == nil {
		return "", fmt.Errorf("inneragent: 未挂接文档库（Repo 缺席）")
	}
	switch node.Op {
	case "judge-issues-batch":
		return r.runJudge(run, node)
	case "apply-doc":
		return r.runApplyDoc(run, node)
	case "merge-results":
		return r.runMerge(run, node)
	}
	r.generic(run, node)
	return "", nil
}

// generic 通用形（既有行为面——未知 op 的轻量执行：一句话任务+执行结论落稿）。
func (r *Runner) generic(run workflow.RunRecord, node workflow.NodeView) {
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	cfg, err := r.G.provider()
	if err != nil {
		r.G.trace("inner provider-fail " + err.Error())
		return
	}
	model := cfg.Model
	if m, ok := r.G.ModelOverride(node.Op); ok {
		model = m
	}
	sys := "你是 workflow 内置轻量 agent（ops_core_inner_agent——低智商环节的润滑剂）。直接完成节点任务并输出简要执行结论（做了什么/结果/下一步建议），中文 ≤200 字，不编造。"
	user := fmt.Sprintf("run=%d 对象=%s 节点=%s 相位=%s unit=%s op=%s\n请按 op 语义完成本节点轻量任务并输出执行结论。", run.Run, run.Involved, node.Name, node.Phase, node.Unit, node.Op)
	out := r.G.complete(ctx, cfg, model, sys, user)
	docName := fmt.Sprintf("inneragent-%s-run%d", node.Name, run.Run)
	if _, err := r.G.Repo.Get(ctx, "DocumentDiff", "global", docName); err == nil {
		return // 幂等：同节点同 run 产物在册
	}
	if out == "" {
		out = "（LLM 输出缺席——诚实词形不编造）"
	}
	content := "apiVersion: core/v1\nkind: DocumentDiff\nmetadata:\n  name: " + docName +
		"\n  scope: global\n  labels: {core/origin: workflow-engine}\nspec:\n  target: " +
		quoteYAML(run.Involved) + "\n  run: " + fmt.Sprint(run.Run) + "\n  agent: ops_core_inner_agent\n  node: " +
		node.Name + "\n  op: " + node.Op + "\n  verdict: |\n" + indentBlock(out, "    ")
	if _, _, err := r.G.Repo.Write(ctx, &repository.Object{Kind: "DocumentDiff", Scope: "global", Name: docName, Content: []byte(content)}, "workflow-engine", "inner-agent:"+node.Name); err != nil {
		r.G.trace("inner save-fail " + docName + ": " + err.Error())
		return
	}
	r.G.trace("inner saved " + docName)
	r.G.Logf("inneragent: 节点 %s (run %d) 执行完成，产物 %s", node.Name, run.Run, docName)
}

// —— judge-issues-batch：批量合判 → 执行计划实体 ——

// planItem 执行计划项（WorkflowPlan spec.items 词形——launcher 扇出的每项一子 run）。
type planItem struct {
	Key      string `json:"key"`                // 项键（幂等 dedup 锚——issue 号词形）
	Title    string `json:"title,omitempty"`    // 项标题（issue title）
	Action   string `json:"action"`             // fix|defer（fix=下发修复链；defer=不下发——回评链解释为什么不做）
	Reason   string `json:"reason,omitempty"`   // 判断理由（defer 项=为什么不做——反馈词面）
	Risk     string `json:"risk,omitempty"`     // low|medium|high（审批分流判定输入）
	Involved string `json:"involved,omitempty"` // 项溯源锚（Repository/<repo>#<n>——子 run involved）
}

// planDocName 执行计划实体名（run 域幂等锚——plan-run<N>）。
func planDocName(run int64) string { return fmt.Sprintf("plan-run%d", run) }

// runJudge 批量合判（父链 Analyzing 相位）：读 issue 批 → LLM 判定下发集 → WorkflowPlan
// 落库 + run 上下文账回填 → decision 词（worth_fixing/ignore）。
func (r *Runner) runJudge(run workflow.RunRecord, node workflow.NodeView) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	// 幂等回读：计划实体在册=重启/重放不重复调 LLM（判决与 items 从实体还原）。
	if o, err := r.planRepo().Get(ctx, "WorkflowPlan", "global", planDocName(run.Run)); err == nil && o != nil {
		if d, sum, items, ok := parsePlanDoc(string(o.Content)); ok {
			r.memoSetPlan(run.Run, d, sum, items)
			r.G.trace("judge idempotent run=" + fmt.Sprint(run.Run) + " items=" + fmt.Sprint(len(items)))
			return d, nil
		}
	}
	issues := r.issueBatch(ctx, run)
	if len(issues) == 0 {
		// 无批可判=显式缺口（不编造计划——诚实词形；留挂起由人/重试接管）。
		return "", fmt.Errorf("judge: issue 批缺席（run 上下文与 GitHub 均无可判项——gh-issue-fetch 未产出或仓库无 open issue）")
	}
	cfg, err := r.G.provider()
	if err != nil {
		return "", fmt.Errorf("judge: Provider 解析失败: %w", err)
	}
	model := cfg.Model
	if m, ok := r.G.ModelOverride(node.Op); ok {
		model = m
	}
	sys := "你是 workflow 批量合判 judge（ops_core_inner_agent）。给定一批 GitHub issue（JSON），判断哪些值得现在下发修复（action=fix）、哪些不做（action=defer 并给出为什么不做的一句话理由——不做也要有反馈）。只输出 JSON 对象：{\"decision\":\"worth_fixing|ignore\",\"summary\":\"一句话汇总\",\"items\":[{\"key\":\"issue号\",\"title\":…,\"action\":\"fix|defer\",\"reason\":…,\"risk\":\"low|medium|high\"}]}，不要 Markdown 包裹，key 用 issue 数字字符串，全部 issue 都要出现在 items 里。"
	user := "run=" + fmt.Sprint(run.Run) + " 仓库=" + run.Involved + "\nissue 批：\n" + clip(string(issues), 12000)
	raw := r.G.complete(ctx, cfg, model, sys, user)
	decision, summary, items, perr := parsePlanJSON(raw)
	if perr != nil || len(items) == 0 {
		// 判定不可解析=失败收敛（留挂起缺口可见——不编造计划不盲目放行）。
		return "", fmt.Errorf("judge: 计划不可解析（%v——原始输出 %d 字已留痕）", perr, len(raw))
	}
	if decision != "ignore" && decision != "worth_fixing" {
		decision = "worth_fixing" // 词形越轨收敛到缺省正路（ignore 需显式声明）
	}
	// involved 补全（issue 号→Repository/<repo>#<n>——子 run involved 锚）。
	repo := involvedRepo(run.Involved)
	for i := range items {
		if items[i].Involved == "" && repo != "" {
			items[i].Involved = "Repository/global/" + repo + "#" + items[i].Key
		}
		if items[i].Risk == "" {
			items[i].Risk = "low"
		}
	}
	content := planDoc(run, node, decision, summary, items)
	if _, _, err := r.planRepo().Write(ctx, &repository.Object{Kind: "WorkflowPlan", Scope: "global", Name: planDocName(run.Run), Content: []byte(content)}, "workflow-engine", "batch-judge:run:"+fmt.Sprint(run.Run)); err != nil {
		return "", fmt.Errorf("judge: 计划实体落库失败: %w", err)
	}
	r.memoSetPlan(run.Run, decision, summary, items)
	r.G.Logf("judge: run %d 计划=%d 项（fix=%d defer=%d）decision=%s 实体=%s", run.Run, len(items), countAction(items, "fix"), countAction(items, "defer"), decision, planDocName(run.Run))
	return decision, nil
}

// issueBatch issue 批内容源（run 上下文账 op:issue-list 全文优先——gh-issue-fetch 节点真产；
// miss=GitHub REST 现拉兜底——重启丢账后 judge 仍可独立成判）。
func (r *Runner) issueBatch(ctx context.Context, run workflow.RunRecord) []byte {
	if raw, ok := r.Memo.Get(run.Run, "op:issue-list"); ok {
		if s, ok := raw.(string); ok && strings.TrimSpace(s) != "" {
			return []byte(s)
		}
	}
	return r.G.ghListIssues(ctx, involvedRepo(run.Involved))
}

// ghListIssues GitHub REST 拉 open issues（ghOwner 补全 org 前缀——judge.go 同款零硬编码；
// GH_TOKEN 在场则带认证头）。失败=空（调用方显式缺口词形）。
func (r *Generator) ghListIssues(ctx context.Context, repo string) []byte {
	if repo == "" {
		return nil
	}
	if !strings.Contains(repo, "/") {
		if owner := r.ghOwner(ctx); owner != "" {
			repo = owner + "/" + repo
		}
	}
	url := fmt.Sprintf("https://api.github.com/repos/%s/issues?state=open&per_page=50", repo)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil
	}
	if tok := strings.TrimSpace(os.Getenv("GH_TOKEN")); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil
	}
	var list []struct {
		Number      int       `json:"number"`
		Title       string    `json:"title"`
		Body        string    `json:"body"`
		State       string    `json:"state"`
		PullRequest *struct{} `json:"pull_request"` // PR 不入批（issue 域判定）
		Labels      []struct {
			Name string `json:"name"`
		} `json:"labels"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 4<<20)).Decode(&list); err != nil {
		return nil
	}
	type issueRow struct {
		Key    string   `json:"key"`
		Title  string   `json:"title"`
		Labels []string `json:"labels,omitempty"`
		Body   string   `json:"body,omitempty"`
	}
	rows := make([]issueRow, 0, len(list))
	for _, it := range list {
		if it.PullRequest != nil {
			continue
		}
		labels := make([]string, 0, len(it.Labels))
		for _, l := range it.Labels {
			labels = append(labels, l.Name)
		}
		rows = append(rows, issueRow{Key: fmt.Sprint(it.Number), Title: it.Title, Labels: labels, Body: clip(it.Body, 800)})
	}
	out, err := json.Marshal(rows)
	if err != nil {
		return nil
	}
	return out
}

// parsePlanJSON 执行计划 JSON 解析（容忍 ```json 围栏与前后散文——judge.go parseVerdictJSON 同款口径）。
func parsePlanJSON(raw string) (decision, summary string, items []planItem, err error) {
	s := strings.TrimSpace(raw)
	if i := strings.Index(s, "{"); i >= 0 {
		s = s[i:]
	}
	if i := strings.LastIndex(s, "}"); i >= 0 {
		s = s[:i+1]
	}
	var d struct {
		Decision string     `json:"decision"`
		Summary  string     `json:"summary"`
		Items    []planItem `json:"items"`
	}
	if err := json.Unmarshal([]byte(s), &d); err != nil {
		return "", "", nil, err
	}
	for _, it := range d.Items {
		if strings.TrimSpace(it.Key) == "" {
			return "", "", nil, fmt.Errorf("计划项缺 key")
		}
		if it.Action != "fix" && it.Action != "defer" {
			return "", "", nil, fmt.Errorf("项 %s action 词形越轨（%q）", it.Key, it.Action)
		}
		items = append(items, it)
	}
	if len(items) == 0 {
		return "", "", nil, fmt.Errorf("计划零项")
	}
	return strings.TrimSpace(d.Decision), strings.TrimSpace(d.Summary), items, nil
}

// parsePlanDoc 幂等回读（从落库实体还原 decision/summary/items——重放不重复调 LLM）。
func parsePlanDoc(content string) (decision, summary string, items []planItem, ok bool) {
	inItems := false
	for _, line := range strings.Split(content, "\n") {
		s := strings.TrimSpace(line)
		switch {
		case strings.HasPrefix(s, "decision:"):
			decision = strings.TrimSpace(strings.TrimPrefix(s, "decision:"))
		case strings.HasPrefix(s, "summary: "):
			summary = strings.Trim(strings.TrimSpace(strings.TrimPrefix(s, "summary:")), "\"")
		case strings.HasPrefix(s, "items:"):
			inItems = true
		case inItems && strings.HasPrefix(s, "- {") && strings.HasSuffix(s, "}"):
			var it planItem
			if json.Unmarshal([]byte(s[2:]), &it) == nil && it.Key != "" {
				items = append(items, it)
			}
		}
	}
	return decision, summary, items, decision != "" && len(items) > 0
}

// planDoc 执行计划实体词形（WorkflowPlan——scheme kind 注册见 core_design scheme.yaml）。
func planDoc(run workflow.RunRecord, node workflow.NodeView, decision, summary string, items []planItem) string {
	var b strings.Builder
	b.WriteString("apiVersion: core/v1\nkind: WorkflowPlan\nmetadata:\n  name: " + planDocName(run.Run) +
		"\n  scope: global\n  labels: {core/origin: workflow-engine}\nspec:\n  workflow: " + run.Workflow +
		"\n  run: " + fmt.Sprint(run.Run) + "\n  node: " + node.Name +
		"\n  decision: " + decision + "\n  summary: " + quoteYAML(summary) + "\n  items:\n")
	for _, it := range items {
		b.WriteString("    - " + itemJSON(it) + "\n")
	}
	return b.String()
}

// itemJSON 计划项行内 JSON（yaml flow map 词形——parsePlanDoc 回读同源单点）。
func itemJSON(it planItem) string {
	b, _ := json.Marshal(it)
	return string(b)
}

// memoSetPlan run 上下文账回填（plan 项/结论/判断词——下游 launcher 与 merge 的读源）。
func (r *Runner) memoSetPlan(run int64, decision, summary string, items []planItem) {
	if r.Memo == nil {
		return
	}
	r.Memo.Set(run, "planDecision", decision)
	r.Memo.Set(run, "planSummary", summary)
	r.Memo.Set(run, "planItems", items)
}

// —— apply-doc：子链落稿 ——

// runApplyDoc 子链计划项落稿（Planning 相位）：defer 项=直接落稿+decision=defer（沿
// Planning defer 边直达回评——「不做也有反馈」子链半边）；fix 项=修复要点落稿（developer-system
// 会话的输入实体——会话 briefing 读本稿+计划项）。产物幂等锚=inneragent-doc-apply-run<N>。
func (r *Runner) runApplyDoc(run workflow.RunRecord, node workflow.NodeView) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	item := r.childItem(run)
	if item == nil {
		return "", fmt.Errorf("apply-doc: 计划项缺席（子 run 上下文未绑定——launcher 绑定缺口可见）")
	}
	docName := fmt.Sprintf("inneragent-%s-run%d", node.Name, run.Run)
	if item.Action == "defer" {
		// defer 项不调 LLM（判断已由 batch-judge 给出——落稿固化理由即可）。
		content := "apiVersion: core/v1\nkind: DocumentDiff\nmetadata:\n  name: " + docName +
			"\n  scope: global\n  labels: {core/origin: workflow-engine}\nspec:\n  target: " + quoteYAML(run.Involved) +
			"\n  run: " + fmt.Sprint(run.Run) + "\n  node: " + node.Name + "\n  op: " + node.Op +
			"\n  action: defer\n  verdict: |\n" + indentBlock("不下发：项 "+item.Key+" "+item.Title+"——"+item.Reason, "    ")
		if _, _, err := r.planRepo().Write(ctx, &repository.Object{Kind: "DocumentDiff", Scope: "global", Name: docName, Content: []byte(content)}, "workflow-engine", "apply-doc:defer:"+fmt.Sprint(run.Run)); err != nil {
			return "", fmt.Errorf("apply-doc: 落稿失败: %w", err)
		}
		r.G.Logf("apply-doc: run %d 项 %s defer 落稿（为什么不做=%s）", run.Run, item.Key, item.Reason)
		return "defer", nil
	}
	cfg, err := r.G.provider()
	if err != nil {
		return "", fmt.Errorf("apply-doc: Provider 解析失败: %w", err)
	}
	model := cfg.Model
	if m, ok := r.G.ModelOverride(node.Op); ok {
		model = m
	}
	subject := r.G.readSubject(ctx, run.Involved) // issue 原文（title+body——判决/落稿共用同一被审物源）
	sys := "你是 workflow 落稿 agent（ops_core_inner_agent）。按计划项与 issue 原文写修复要点稿（改哪些面/怎么改/怎么验证），中文 ≤300 字，不编造文件名以外的细节。只输出要点稿正文。"
	user := fmt.Sprintf("run=%d issue=%s\n计划项：%s\n\nissue 原文：\n%s", run.Run, run.Involved, itemJSON(*item), clip(subject, 4000))
	out := r.G.complete(ctx, cfg, model, sys, user)
	if out == "" {
		out = "（LLM 输出缺席——诚实词形不编造）"
	}
	content := "apiVersion: core/v1\nkind: DocumentDiff\nmetadata:\n  name: " + docName +
		"\n  scope: global\n  labels: {core/origin: workflow-engine}\nspec:\n  target: " + quoteYAML(run.Involved) +
		"\n  run: " + fmt.Sprint(run.Run) + "\n  node: " + node.Name + "\n  op: " + node.Op +
		"\n  action: fix\n  verdict: |\n" + indentBlock(out, "    ")
	if _, _, err := r.planRepo().Write(ctx, &repository.Object{Kind: "DocumentDiff", Scope: "global", Name: docName, Content: []byte(content)}, "workflow-engine", "apply-doc:"+fmt.Sprint(run.Run)); err != nil {
		return "", fmt.Errorf("apply-doc: 落稿失败: %w", err)
	}
	r.G.Logf("apply-doc: run %d 项 %s fix 落稿（%d 字）", run.Run, item.Key, len(out))
	return "", nil // 机械 advance——下一落点=developer-system 会话
}

// childItem 子 run 计划项读取（run 上下文账 planItem——fanout launcher ChildStarted 绑定）。
func (r *Runner) childItem(run workflow.RunRecord) *planItem {
	if r.Memo == nil {
		return nil
	}
	if m := r.Memo.GetMap(run.Run, "planItem"); m != nil {
		b, err := json.Marshal(m)
		if err == nil {
			var it planItem
			if json.Unmarshal(b, &it) == nil && it.Key != "" {
				return &it
			}
		}
	}
	return nil
}

// —— merge-results：交叉汇总 ——

// runMerge 父链 join 后交叉汇总（Reporting 相位）：子 run 终态现查（Runs.State 按 ParentRun
// 过滤）+计划对照 → 汇总落稿 → decision=release（大规模批完成才发版）或空词（默认跳过发版
// 机械收束）。
func (r *Runner) runMerge(run workflow.RunRecord, node workflow.NodeView) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	docName := fmt.Sprintf("inneragent-%s-run%d", node.Name, run.Run)
	// 幂等回读（汇总稿在册=不重复调 LLM；decision 从稿面还原）。
	if o, err := r.planRepo().Get(ctx, "DocumentDiff", "global", docName); err == nil && o != nil {
		if d := docDecision(string(o.Content)); d != "" {
			return d, nil
		}
	}
	_, summary, items, ok := func() (string, string, []planItem, bool) {
		if o, err := r.planRepo().Get(ctx, "WorkflowPlan", "global", planDocName(run.Run)); err == nil && o != nil {
			return parsePlanDoc(string(o.Content))
		}
		return "", "", nil, false
	}()
	if !ok {
		if m := r.Memo.GetMap(run.Run, "planItems"); m != nil {
			if b, err := json.Marshal(m); err == nil {
				var its []planItem
				if json.Unmarshal(b, &its) == nil && len(its) > 0 {
					items = its
				}
			}
		}
	}
	children := r.childOutcomes(run.Run)
	cfg, err := r.G.provider()
	if err != nil {
		return "", fmt.Errorf("merge: Provider 解析失败: %w", err)
	}
	model := cfg.Model
	if m, ok := r.G.ModelOverride(node.Op); ok {
		model = m
	}
	sys := "你是 workflow 交叉汇总 agent（ops_core_inner_agent）。给定执行计划与各子链终态，写交叉汇总：计划多少项/下发多少/完成多少/失败多少/defer 多少（各是什么）/整体是否构成一次大规模功能修复完成（是→decision=release，构成发版条件；否→decision 留空）。只输出 JSON：{\"decision\":\"release|\",\"summary\":\"中文汇总 ≤300 字\"}。"
	user := fmt.Sprintf("run=%d 计划=%d 项\n%s\n子链终态：\n%s", run.Run, len(items), itemsDigest(items), children)
	raw := r.G.complete(ctx, cfg, model, sys, user)
	decision, msum, perr := parseMergeJSON(raw)
	if perr != nil {
		decision, msum = "", "（汇总不可解析——"+perr.Error()+"）"
	}
	if summary != "" && msum != "" && len(items) > 0 {
		msum = summary + "；" + msum
	}
	content := "apiVersion: core/v1\nkind: DocumentDiff\nmetadata:\n  name: " + docName +
		"\n  scope: global\n  labels: {core/origin: workflow-engine}\nspec:\n  target: " + quoteYAML(run.Involved) +
		"\n  run: " + fmt.Sprint(run.Run) + "\n  node: " + node.Name + "\n  op: " + node.Op +
		"\n  decision: " + decision + "\n  verdict: |\n" + indentBlock(msum, "    ") +
		"\n  children: |\n" + indentBlock(children, "    ")
	if _, _, err := r.planRepo().Write(ctx, &repository.Object{Kind: "DocumentDiff", Scope: "global", Name: docName, Content: []byte(content)}, "workflow-engine", "merge:"+fmt.Sprint(run.Run)); err != nil {
		return "", fmt.Errorf("merge: 汇总落稿失败: %w", err)
	}
	if r.Memo != nil {
		r.Memo.Set(run.Run, "mergeSummary", msum)
	}
	r.G.Logf("merge: run %d 汇总完成 decision=%q（%d 子链）", run.Run, decision, strings.Count(children, "\n"))
	return decision, nil
}

// childOutcomes 子 run 终态现查（Runs.State 按 ParentRun 过滤——run 序号/状态/相位/末行注记）。
func (r *Runner) childOutcomes(parent int64) string {
	if r.Runs == nil {
		return "（子 run 现查缺席——Runs 未注入）"
	}
	st := r.Runs.State()
	var b strings.Builder
	for i := len(st.Runs) - 1; i >= 0; i-- {
		c := st.Runs[i]
		if c.ParentRun != parent {
			continue
		}
		note := ""
		for j := len(c.History) - 1; j >= 0; j-- {
			if c.History[j].Note != "" {
				note = c.History[j].Note
				break
			}
		}
		fmt.Fprintf(&b, "run%d status=%s phase=%s involved=%s note=%s\n", c.Run, c.Status, c.Phase, c.Involved, clip(note, 160))
	}
	if b.Len() == 0 {
		return "（无子 run——计划零项或扇出缺席）"
	}
	return b.String()
}

// parseMergeJSON 汇总 JSON 解析（decision 允许空词=默认跳过发版）。
func parseMergeJSON(raw string) (decision, summary string, err error) {
	s := strings.TrimSpace(raw)
	if i := strings.Index(s, "{"); i >= 0 {
		s = s[i:]
	}
	if i := strings.LastIndex(s, "}"); i >= 0 {
		s = s[:i+1]
	}
	var d struct {
		Decision string `json:"decision"`
		Summary  string `json:"summary"`
	}
	if err := json.Unmarshal([]byte(s), &d); err != nil {
		return "", "", err
	}
	d.Decision = strings.TrimSpace(d.Decision)
	if d.Decision != "" && d.Decision != "release" {
		d.Decision = "" // 词形越轨收敛到缺省（不发版）——release 需显式声明
	}
	return d.Decision, strings.TrimSpace(d.Summary), nil
}

// docDecision 稿面 decision 还原（merge 幂等回读）。
func docDecision(content string) string {
	for _, line := range strings.Split(content, "\n") {
		s := strings.TrimSpace(line)
		if strings.HasPrefix(s, "decision:") {
			return strings.TrimSpace(strings.TrimPrefix(s, "decision:"))
		}
	}
	return ""
}

// itemsDigest 计划项摘要（merge prompt 词面——一行一项）。
func itemsDigest(items []planItem) string {
	var b strings.Builder
	for _, it := range items {
		fmt.Fprintf(&b, "#%s %s action=%s risk=%s %s\n", it.Key, it.Title, it.Action, it.Risk, it.Reason)
	}
	return b.String()
}

// countAction 计划项动作计数。
func countAction(items []planItem, action string) int {
	n := 0
	for _, it := range items {
		if it.Action == action {
			n++
		}
	}
	return n
}

// involvedRepo involved 投影 → 仓名（Repository/global/<repo>[#<n>] 词形——judge/列表的仓锚）。
func involvedRepo(involved string) string {
	name := involved
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	if i := strings.IndexByte(name, '#'); i > 0 {
		name = name[:i]
	}
	return strings.TrimSpace(name)
}

// planRepo 落库面（Repo 字段兜底 G.Repo——同一文档库双注入口的收敛单点）。
func (r *Runner) planRepo() repository.Repository {
	if r.Repo != nil {
		return r.Repo
	}
	return r.G.Repo
}
