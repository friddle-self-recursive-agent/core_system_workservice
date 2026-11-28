// approvaldiff.go —— 审批判决文档生成器（ops_core_inner_agent——2026-11-28 用户裁定：
// 「审批裁定的时候走一个小内置 agent[调用 llm 即可]整理下 DocumentDiff。生成 Diff 文档。
// 告诉我改了什么」+「继续内置 ops_core_inner_agent 继续 scheme 配置。继续用这个来生成
// DocumentDiff」）。
//
// 职责：run 落人审闸锚（AwaitingApproval 挂起）时——读审批对象现值（基线）与待批稿
// （工作值）→ 调 openai 兼容小模型（Provider low 档）整理人话判决（改了什么/为什么/影响面/
// 风险）→ DocumentDiff 文档落库（workflow-engine writer；同名幂等防重）。
//
// 声明面：scheme docTypes DocumentDiff + Workflow.spec.approval.diffSummary（core_design）；
// 本包=ops_core_inner_agent 的 builtin 实现（装配层挂接 RunOptions.DiffSummarizer）。
package approvaldiff

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/connectors"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/draft"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
	"gopkg.in/yaml.v3"
)

// Generator ops_core_inner_agent 实现（workflow.DiffSummarizer 契约）。
type Generator struct {
	Repo     repository.Repository
	Drafts   *draft.Store
	Secrets  connectors.SecretStore // nil=secret: 引用解析缺口可见（判决降级为缺席注记）
	Provider string                 // 缺省 openai-default（openai 兼容账号包：baseUrl+secret+low 档模型）
	Timeout  time.Duration          // 单次 LLM 调用超时（缺省 30s）
	Logf     func(format string, args ...any)
}

// providerCfg Provider 对象解析投影（config.baseUrl/auth.secret.source/config.models.low）。
type providerCfg struct {
	BaseURL   string
	SecretSrc string
	Model     string
	Mapping   map[string]string
}

// OnApprovalSuspended 契约入口（引擎 goroutine 调用——本方法阻塞自有协程不阻塞引擎）。
// objKey=kind/scope/name（run.Involved 投影）。全程失败只记 stderr 不回灌引擎。
func (g *Generator) OnApprovalSuspended(run workflow.RunRecord, objKey string) {
	g.trace("enter objKey=" + objKey + " run=" + fmt.Sprint(run.Run))
	defer func() {
		if rec := recover(); rec != nil {
			g.trace("panic obj=" + objKey + ": " + fmt.Sprint(rec))
		}
	}()
	if g.Repo == nil || g.Logf == nil || objKey == "" {
		return
	}
	kind, scope, name, ok := splitObjKey(objKey)
	if !ok {
		g.trace("bad-objKey " + objKey)
		return
	}
	ctx := context.Background()
	docName := diffDocNameForRun(kind, name, run.Run)
	// 幂等：同 run 同对象判决已在册=跳过（防重——引擎不重试但部署重启可能重挂；幂等域=
	// run——2026-11-28 用户裁定「还是审批为空」纠偏：旧的不带 run 后缀一对象一名让后续
	// 全部 idempotent-skip，新挂起的待批稿永远读到首份陈旧判决——per-run 各自一份）。
	if _, err := g.Repo.Get(ctx, "DocumentDiff", "global", docName); err == nil {
		g.trace("idempotent-skip " + docName)
		return
	}
	base, working, verdict := g.summarize(ctx, kind, scope, name)
	content := buildDoc(run, objKey, docName, verdict, base, working)
	if _, _, err := g.Repo.Write(ctx, &repository.Object{
		Kind: "DocumentDiff", Scope: "global", Name: docName, Content: []byte(content),
	}, "workflow-engine", "approval-diff:"+objKey); err != nil {
		g.trace("write-fail obj=" + objKey + " err=" + err.Error())
		return
	}
	g.Logf("approvaldiff: 判决文档已落库 %s（obj=%s）", docName, objKey)
}

// trace 部署探针：stderr=socketpair 黑洞（authority cored）——判决链路取证走文件尾随。
func (g *Generator) trace(msg string) {
	if f, err := osOpenAppend("/tmp/approvaldiff.log"); err == nil {
		fmt.Fprintf(f, "%s %s\n", time.Now().Format("15:04:05"), msg)
		f.Close()
	}
	if g.Logf != nil {
		g.Logf("approvaldiff: %s", msg)
	}
}

func osOpenAppend(path string) (*os.File, error) {
	return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
}

// summarize 读双值+调 LLM（任一步失败=诚实缺席词形，不编造）。
func (g *Generator) summarize(ctx context.Context, kind, scope, name string) (base, working, verdict string) {
	base = readBase(g.Repo, ctx, kind, scope, name)
	g.trace("base-read len=" + fmt.Sprint(len(base)))
	working = readWorking(g.Drafts, kind, scope, name)
	g.trace("draft-read len=" + fmt.Sprint(len(working)))
	if working == "" {
		g.trace("no-draft obj=" + kind + "/" + scope + "/" + name)
		return base, working, "（待批稿缺席——本对象可能走对象直写面，判决文档仅留锚）"
	}
	g.trace("llm-call start model-path")
	verdict = g.callLLM(ctx, base, working, kind, name)
	g.trace("llm-call done len=" + fmt.Sprint(len(verdict)))
	return base, working, verdict
}

// complete 通用 LLM 调用（model 已解析真名）——Runner/判决共用内核。
func (g *Generator) complete(ctx context.Context, cfg providerCfg, model, sys, user string) string {
	return g.chat(ctx, cfg, model, sys, user)
}

// ModelOverride taskModelMapping[op] 真网关模型名（tier 词形网关不认——ops_core_inner_agent
// 词形映射在 Provider 对象 taskModelMapping，部署侧 apply 配置）。
func (g *Generator) ModelOverride(op string) (string, bool) {
	if op == "" || g.Repo == nil {
		return "", false
	}
	cfg, err := g.provider()
	if err != nil {
		return "", false
	}
	return cfg.Mapping[op], cfg.Mapping[op] != ""
}

// callLLM 小内置 agent：Provider low 档模型，输入双文档，输出人话判决。
func (g *Generator) callLLM(ctx context.Context, base, working, kind, name string) string {
	cfg, err := g.provider()
	if err != nil {
		return "（Provider 解析失败——" + err.Error() + "）"
	}
	sys := "你是审批辅助小 agent（ops_core_inner_agent）。对比文档基线与待批稿，用中文人话总结：1) 改了什么（逐条、具体）；2) 为什么要改（若可判断）；3) 影响面与风险（有无删改关键约束/权限/验收判据）。不超过 300 字，只依据给出的两份文本，不编造。"
	user := fmt.Sprintf("对象：%s/%s\n\n【基线（现值）】\n%s\n\n【待批稿】\n%s", kind, name, clip(base, 6000), clip(working, 6000))
	return g.chat(ctx, cfg, cfg.Model, sys, user)
}

// chat HTTP 交互内核（502/503 退避重试×2——BugDoc #3 同源瞬态病）。
func (g *Generator) chat(ctx context.Context, cfg providerCfg, model, sys, user string) string {
	payload := map[string]any{
		"model": model, // taskModelMapping[op] 真名优先（cfg.Model=low 档兜底；旧词形忽略 model 参数已修）
		"messages": []map[string]string{
			{"role": "system", "content": sys},
			{"role": "user", "content": user},
		},
		"temperature": 0.2,
	}
	buf, _ := json.Marshal(payload)
	url := strings.TrimRight(cfg.BaseURL, "/") + "/chat/completions"
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
	if err != nil {
		return "（LLM 请求构造失败——" + err.Error() + "）"
	}
	req.Header.Set("Content-Type", "application/json")
	if cfg.SecretSrc != "" {
		key, kerr := connectors.ResolveSecret(cfg.SecretSrc, g.Secrets, nil)
		if kerr != nil {
			return "（密钥解析失败——" + kerr.Error() + "；引用=" + cfg.SecretSrc + "）"
		}
		req.Header.Set("Authorization", "Bearer "+key)
	}
	timeout := g.Timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	cli := &http.Client{Timeout: timeout}
	// 502/瞬态重试（BugDoc bug-workflow-agent-502-20261127 #3 同源——网关偶发 502 空流
	// 是在册 bug；判决小 agent 侧带退避重试 2 次，重试上限后走失败词形不编造）。
	g.trace("http-do url=" + url + " model=" + cfg.Model)
	var resp *http.Response
	for attempt := 0; attempt < 3; attempt++ {
		if attempt > 0 {
			time.Sleep(time.Duration(attempt) * 3 * time.Second)
			req, _ = http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(buf))
			req.Header.Set("Content-Type", "application/json")
			if cfg.SecretSrc != "" {
				if key, kerr := connectors.ResolveSecret(cfg.SecretSrc, g.Secrets, nil); kerr == nil {
					req.Header.Set("Authorization", "Bearer "+key)
				}
			}
		}
		var doErr error
		resp, doErr = cli.Do(req)
		err = doErr
		g.trace("http-done attempt=" + fmt.Sprint(attempt) + " err=" + fmt.Sprint(err != nil))
		if err == nil && resp.StatusCode != http.StatusBadGateway && resp.StatusCode != http.StatusServiceUnavailable {
			break
		}
		if resp != nil {
			resp.Body.Close()
			resp = nil
		}
	}
	if err != nil || resp == nil {
		return "（LLM 调用失败——" + fmt.Sprint(err) + "）"
	}
	defer resp.Body.Close()
	var out struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil || len(out.Choices) == 0 {
		return "（LLM 响应不可解析——status=" + fmt.Sprint(resp.StatusCode) + "）"
	}
	return strings.TrimSpace(out.Choices[0].Message.Content)
}

// provider 读 Provider 对象（yaml）解析 baseUrl/secret 引用/low 档模型。
func (g *Generator) provider() (providerCfg, error) {
	pn := g.Provider
	if pn == "" {
		pn = "openai-default"
	}
	obj, err := g.Repo.Get(context.Background(), "Provider", "global", pn)
	if err != nil {
		return providerCfg{}, fmt.Errorf("Provider/%s 读取失败: %w", pn, err)
	}
	var d struct {
		Spec struct {
			Auth struct {
				Secret struct {
					Source string `yaml:"source"`
				} `yaml:"secret"`
			} `yaml:"auth"`
			Config struct {
				BaseURL    string `yaml:"baseUrl"`
				ModelTiers struct {
					Low string `yaml:"low"`
				} `yaml:"modelTiers"`
				TaskModelMapping map[string]string `yaml:"taskModelMapping"`
				Models           []string          `yaml:"models"`
			} `yaml:"config"`
		} `yaml:"spec"`
	}
	if err := yaml.Unmarshal(obj.Content, &d); err != nil {
		return providerCfg{}, fmt.Errorf("Provider/%s 解析失败: %w", pn, err)
	}
	if d.Spec.Config.BaseURL == "" {
		return providerCfg{}, fmt.Errorf("Provider/%s 缺 config.baseUrl", pn)
	}
	g.trace("provider-read base=" + d.Spec.Config.BaseURL + " secret=" + d.Spec.Auth.Secret.Source)
	// 模型名解析顺序：taskModelMapping["ops_core_inner_agent"]（真实网关模型名——tier 词形
	// 是全系统语义，网关不认）> modelTiers.low > models[0] > flash-tier 兜底。
	model := ""
	if m, ok := d.Spec.Config.TaskModelMapping["ops_core_inner_agent"]; ok {
		model = m
	}
	if model == "" {
		model = d.Spec.Config.ModelTiers.Low
	}
	if model == "" && len(d.Spec.Config.Models) > 0 {
		model = d.Spec.Config.Models[0] // 退档：models 列表首项=低档
	}
	if model == "" {
		model = "flash-tier"
	}
	return providerCfg{BaseURL: d.Spec.Config.BaseURL, SecretSrc: d.Spec.Auth.Secret.Source, Model: model, Mapping: d.Spec.Config.TaskModelMapping}, nil
}

// buildDoc 组装 DocumentDiff 对象（yaml 对象+verdict 正文；锚=target@rv+run+agent 词形）。
func buildDoc(run workflow.RunRecord, objKey, docName, verdict, base, working string) string {
	changed := "—"
	if base != "" && working != "" && base != working {
		changed = fmt.Sprintf("%d 行 → %d 行（内容有变更）", countLines(base), countLines(working))
	} else if base == "" && working != "" {
		changed = "新对象（基线缺席）"
	}
	return strings.Join([]string{
		"apiVersion: core/v1",
		"kind: DocumentDiff",
		"metadata:",
		"  name: " + docName,
		"  scope: global",
		"  labels: {core/origin: workflow-engine}",
		"spec:",
		"  target: " + quoteYAML(objKey),
		"  run: " + fmt.Sprint(run.Run),
		"  agent: ops_core_inner_agent",
		"  provider: " + quoteYAML(gProvName),
		"  changed: " + quoteYAML(changed),
		"  verdict: |",
	}, "\n") + "\n" + indentBlock(verdict, "    ")
}

// —— 小工具 ——

var gProvName = "openai-default"

func splitObjKey(k string) (kind, scope, name string, ok bool) {
	parts := strings.Split(k, "/")
	if len(parts) != 3 || parts[0] == "" || parts[2] == "" {
		return "", "", "", false
	}
	return parts[0], parts[1], parts[2], true
}

func diffDocName(kind, name string, rv int64) string {
	base := strings.ToLower(strings.NewReplacer("_", "-", "/", "-", ".", "-").Replace(kind + "-" + name))
	if rv > 0 {
		return fmt.Sprintf("docdiff-%s-rv%d", base, rv)
	}
	return "docdiff-" + base
}

// diffDocNameForRun 判决文档名（per-run：docdiff-<kind>-<name>-run<N>——每次人审挂起各自
// 一份判决，幂等域=run；2026-11-28 用户裁定纠偏——旧一对象一名让后续挂起全部 skip，
// 新待批稿永远读到首份陈旧判决）。同 run 内二次挂起（多闸锚对象）复用同份=幂等窗口。
func diffDocNameForRun(kind, name string, run int64) string {
	base := strings.ToLower(strings.NewReplacer("_", "-", "/", "-", ".", "-").Replace(kind + "-" + name))
	return fmt.Sprintf("docdiff-%s-run%d", base, run)
}

// approvalJudgeDocName 自审批真判决 DocumentDiff 名（2026-11-28 #4：inneragent-approval-run<N>
// ——同名幂等；per-run 隔离，与 inneragent-<node>-run<N> 润滑剂产物分域）。
func approvalJudgeDocName(run int64) string {
	return fmt.Sprintf("inneragent-approval-run%d", run)
}

func readBase(repo repository.Repository, ctx context.Context, kind, scope, name string) string {
	o, err := repo.Get(ctx, kind, scope, name)
	if err != nil || o == nil {
		return ""
	}
	return string(o.Content)
}

func readWorking(store *draft.Store, kind, scope, name string) string {
	if store == nil {
		return ""
	}
	d, err := store.Get(kind, scope, name)
	if err != nil || d == nil {
		return ""
	}
	return string(d.Content)
}

func clip(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "\n…（截断）"
}

func countLines(s string) int { return strings.Count(s, "\n") + 1 }

func quoteYAML(s string) string {
	return "\"" + strings.NewReplacer("\\", "\\\\", "\"", "\\\"", "\n", "\\n").Replace(s) + "\""
}

func indentBlock(s, pad string) string {
	var b strings.Builder
	for _, line := range strings.Split(strings.TrimRight(s, "\n"), "\n") {
		if strings.TrimSpace(line) == "" {
			b.WriteString("\n")
			continue
		}
		b.WriteString(pad + line + "\n")
	}
	return b.String()
}
