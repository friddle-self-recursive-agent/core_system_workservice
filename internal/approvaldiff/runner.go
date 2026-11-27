// runner.go —— 内置轻量 agent 执行单元（2026-11-28 用户裁定「doc-apply 的审批以及部分服务
// 运行起来可以使用内置的 agent 来做润滑剂。即低智商要求的可以使用内置 agent」）：
// executor=inner 的 agent/service 节点由本 Runner 执行——引擎内直调 Provider（节点 modelTier
// 或 taskModelMapping[op] 真名），产物落 DocumentDiff（执行记录/判决/摘要——低智商环节的
// 润滑剂输出），完成信号由引擎 InnerDone 续行。重型通路（executor=dsh 缺省）不经此处。
package approvaldiff

import (
	"context"
	"fmt"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
)

// Runner 内置轻量 agent 执行器（workflow.InnerAgentRunner 契约）。
type Runner struct {
	G *Generator // 复用 Provider 解析+LLM 通路（同账号包/secret/重试）
}

// RunInner 执行一个 inner 节点：LLM 整理节点任务（op+run 上下文）→产物 DocumentDiff 落库
// （name=inneragent-<node>-run<run>）。err 非 nil=引擎留挂起。
func (r *Runner) RunInner(run workflow.RunRecord, node workflow.NodeView) error {
	r.G.trace("inner enter run=" + fmt.Sprint(run.Run) + " node=" + node.Name + " op=" + node.Op)
	if r.G == nil || r.G.Repo == nil {
		return fmt.Errorf("inneragent: 未挂接文档库（Repo 缺席）")
	}
	ctx := context.Background()
	cfg, err := r.G.provider()
	if err != nil {
		return fmt.Errorf("inneragent: Provider 解析失败: %w", err)
	}
	model := cfg.Model
	if m, ok := r.G.ModelOverride(node.Op); ok {
		model = m // taskModelMapping[op] 真名优先（网关不认 tier 词形）
	}
	sys := "你是 workflow 内置轻量 agent（ops_core_inner_agent——低智商环节的润滑剂）。直接完成节点任务并输出简要执行结论（做了什么/结果/下一步建议），中文 ≤200 字，不编造。"
	user := fmt.Sprintf("run=%d 对象=%s 节点=%s 相位=%s unit=%s op=%s\n请按 op 语义完成本节点轻量任务并输出执行结论。", run.Run, run.Involved, node.Name, node.Phase, node.Unit, node.Op)
	r.G.trace("inner llm-call model=" + model)
	out := r.G.complete(ctx, cfg, model, sys, user)
	r.G.trace("inner llm-done len=" + fmt.Sprint(len(out)))
	docName := fmt.Sprintf("inneragent-%s-run%d", node.Name, run.Run)
	if _, err := r.G.Repo.Get(ctx, "DocumentDiff", "global", docName); err == nil {
		return nil // 幂等：同节点同 run 产物在册
	}
	if out == "" {
		out = "（LLM 输出缺席——诚实词形不编造）"
	}
	content := "apiVersion: core/v1\nkind: DocumentDiff\nmetadata:\n  name: " + docName +
		"\n  scope: global\n  labels: {core/origin: workflow-engine}\nspec:\n  target: " +
		quoteYAML(run.Involved) + "\n  run: " + fmt.Sprint(run.Run) + "\n  agent: ops_core_inner_agent\n  node: " +
		node.Name + "\n  op: " + node.Op + "\n  verdict: |\n" + indentBlock(out, "    ")
	if _, _, err := r.G.Repo.Write(ctx, &repository.Object{Kind: "DocumentDiff", Scope: "global", Name: docName, Content: []byte(content)}, "workflow-engine", "inner-agent:"+node.Name); err != nil {
		return fmt.Errorf("inneragent: 产物落库失败: %w", err)
	}
	r.G.trace("inner saved " + docName)
	r.G.Logf("inneragent: 节点 %s (run %d) 执行完成，产物 %s", node.Name, run.Run, docName)
	return nil
}
