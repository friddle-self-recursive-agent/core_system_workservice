// ops_bridge.go —— connectops 节点 op 执行桥（2026-11-28 用户裁定「做成workflow的一个节点。
// 即gh-release包括git push/release[scheme调用]」——workflow.md §1.3 opsrun.go 装配半边）。
//
// 链路（service 桥同族——非阻塞契约、锁外派发；完成源在进程内故直接回执不走事件注入面）：
//
//	① run 推进落 unit=connectops 节点 → 引擎挂起（suspendWhy=connectops，on=connectops-run）
//	   → 本桥 RunOp（goroutine）：connectors 注册表按 node.Connector 解析 op 连接器 →
//	   params=node.Params 静态声明 ∪ run.Involved 溯源派生（repo/number——缺席键才补，声明值
//	   优先）→ Invoke(ctx, node.Op, params)；
//	② 回执：成功=OpsSignal(ok, 结果摘要) 续行（on=connectops-resume）；失败=OpsSignal(false,
//	   显式错误) 留挂起（on=connectops-error——缺口可见不静默，service 失败同口径）。
//
// 边界：只在 CORE_SYSTEM_CONNECTOPS_RUN=on 时挂接（runOpts.ConnectOpsRunner 非零才挂起面——
// 缺席=零翻转口径，connectops 节点照旧纯查表推进）；op 调用即效果（connector.md 规则 11），
// 本桥零即兴写动作、params 之外不发明参数；连接器缺席/op 不在声明集=显式失败（诚实缺口）。
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/connectors"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
	"gopkg.in/yaml.v3"
)

// opsRunBridge connectops 节点 op 执行桥（workflow.ConnectOpsRunner 实现——cored 装配层注入
// RunOptions.ConnectOpsRunner）。
type opsRunBridge struct {
	runs *workflow.RunLog // 回执续行面（构造后注入——runOpts 需桥、桥需 runs 的解环同款）
	bus  *eventbus.Bus    // release 完成上报面（GhReleaseEvent 自产——2026-11-28 用户裁定「有Release最新版本[Report的Event]→触发更新」的产源半边；nil=只续行不发布）
	logf func(string, ...any)
}

// opsInvokeTimeout 单 op 调用上限（release-publish=docs 导出+gh 两跳；deploy 侧凭据就绪时
// 常规秒级——上限给足余量，超时=显式失败留挂起）。
const opsInvokeTimeout = 120 * time.Second

// opsNoteMax 回执摘要上限（history Note 词面——超出截断，不冒充全文）。
const opsNoteMax = 220

// RunOp 实现 workflow.ConnectOpsRunner（非阻塞：本函数即 goroutine 体内——advance 派发层
// 已 goroutine 包裹；回执直接调 RunLog.OpsSignal）。
func (b *opsRunBridge) RunOp(run workflow.RunRecord, node workflow.NodeView) {
	logf := b.logf
	if logf == nil {
		logf = func(string, ...any) {}
	}
	c, ok := connectors.LookupOpsConnector(node.Connector)
	if !ok {
		b.done(run, fmt.Sprintf("连接器 %s 未注册（驱动包未导入/装配缺席）", node.Connector), false, "")
		return
	}
	if !connectors.HasOp(c, node.Op) {
		b.done(run, fmt.Sprintf("op %s 不在 %s 声明集", node.Op, node.Connector), false, "")
		return
	}
	// params=node.Params 静态声明 ∪ involved 溯源派生（声明值优先——缺席键才补）。
	params := map[string]any{}
	for k, v := range node.Params {
		params[k] = v
	}
	if repo, num, hit := involvedRepoIssue(run.Involved); hit {
		if _, has := params["repo"]; !has {
			params["repo"] = repo
		}
		if _, has := params["number"]; !has && num > 0 {
			params["number"] = num
		}
	}
	raw, err := json.Marshal(params)
	if err != nil {
		b.done(run, "params 编码失败: "+err.Error(), false, "")
		return
	}
	ctx, cancel := context.WithTimeout(context.Background(), opsInvokeTimeout)
	defer cancel()
	res, err := c.Invoke(ctx, node.Op, raw)
	if err != nil {
		logf("ops 桥 %s/%s 失败（workflow=%s run=%d）: %v", node.Connector, node.Op, run.Workflow, run.Run, err)
		b.done(run, truncateNote("op 执行失败: "+err.Error()), false, "")
		return
	}
	logf("ops 桥 %s/%s 完成（workflow=%s run=%d bytes=%d）", node.Connector, node.Op, run.Workflow, run.Run, len(res))
	if node.Op == "release-publish" {
		b.reportRelease(run, node, params, res)
	}
	b.done(run, truncateNote(string(res)), true, opsDecision(res))
}

// opsDecision op 结果 JSON 的判断词提取（2026-11-28 分叉裁决——实现侧可在 result 声明
// decision 词驱动 workflow 分叉边；缺席/不可解析=空=机械 advance 兜底）。
func opsDecision(res []byte) string {
	var parsed struct {
		Decision string `json:"decision"`
	}
	if err := json.Unmarshal(res, &parsed); err != nil {
		return ""
	}
	return strings.TrimSpace(parsed.Decision)
}

// reportRelease release-publish 成功后的 GhReleaseEvent 自产（2026-11-28 用户裁定产源半边）：
// 仅当 release 跳真发（result.release.tag_name 在场——tag 未声明=跳过不发，无版可报不编造）。
// 事件：source=connector（gh-release 节点经连接器产版），involved=Repository/global/<repo>，
// dedup=gh-release:<repo>:<tag>（同 tag 聚合防重复触发）。发布失败=stderr 注记不回滚续行
// （版已成事实；订阅方丢拍可由 release-list 对账兜底）。
func (b *opsRunBridge) reportRelease(run workflow.RunRecord, node workflow.NodeView, params map[string]any, res []byte) {
	if b.bus == nil {
		return
	}
	var parsed struct {
		Release struct {
			TagName string `json:"tag_name"`
		} `json:"release"`
	}
	if err := json.Unmarshal(res, &parsed); err != nil || parsed.Release.TagName == "" {
		return // 跳过/不可解析=无版可报（诚实缺席——不编造事件）
	}
	repo, _ := params["repo"].(string)
	if repo == "" {
		repo = run.Involved
	}
	evt := eventbus.Event{
		Type:     "GhReleaseEvent",
		Source:   "connector",
		Involved: "Repository/global/" + repo,
		DedupKey: "gh-release:" + repo + ":" + parsed.Release.TagName,
		Level:    "info",
	}
	if err := b.bus.Publish(&evt); err != nil {
		b.logf("ops 桥 GhReleaseEvent 发布失败（release=%s 已成事实，订阅方丢拍走 release-list 对账兜底）: %v", parsed.Release.TagName, err)
	}
}

// done 回执续行（runs 缺席=装配序缺口——留挂起显式注记，不静默不崩溃）。
func (b *opsRunBridge) done(run workflow.RunRecord, note string, ok bool, decision string) {
	if b.runs == nil {
		fmt.Fprintf(os.Stderr, "cored: ops 桥回执缺席 runs（装配序缺口——run=%d 留挂起）: %s\n", run.Run, note)
		return
	}
	b.runs.OpsSignal(run.Run, run.Workflow, note, ok, decision)
}

// involvedRepoIssue run.Involved 溯源派生（Repository/global/<repo>#<n> 词形——issue 域 op
// 的 repo/number 缺省源；其余 involved 词形=不派生不冒充）。
func involvedRepoIssue(involved string) (repo string, number int64, ok bool) {
	name := involved
	if i := strings.LastIndex(name, "/"); i >= 0 {
		name = name[i+1:]
	}
	hi := strings.IndexByte(name, '#')
	if hi <= 0 || hi == len(name)-1 {
		return "", 0, false
	}
	repo = name[:hi]
	n := int64(0)
	for _, ch := range name[hi+1:] {
		if ch < '0' || ch > '9' {
			return "", 0, false // 非纯数字后缀（标签词形）——不派生
		}
		n = n*10 + int64(ch-'0')
	}
	return repo, n, true
}

// truncateNote 回执摘要截断（opsNoteMax 上限——超出截断+省略号注记）。
func truncateNote(s string) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	if len(s) <= opsNoteMax {
		return s
	}
	return s[:opsNoteMax] + "…"
}

// applyConnectorConfigs ConnectorOps 对象声明值注入（registry 可配置契约——spec.config /
// 顶层 config 两形态同源合并，spec 优先；解析失败=跳过该对象 stderr 注记，不崩装配）。
// 实现=类型断言 ApplyConfig（connectops-docs/gh 等可配置连接器）；缺席=不可配置连接器零动作。
func applyConnectorConfigs(repo repository.Repository) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	objs, err := repo.List(ctx, "ConnectorOps", "")
	if err != nil {
		fmt.Fprintf(os.Stderr, "cored: ConnectorOps 声明枚举失败（连接器缺省配置在位——声明注入跳过）: %v\n", err)
		return
	}
	for _, o := range objs {
		if o == nil {
			continue
		}
		var decl struct {
			Config map[string]any `yaml:"config"`
			Spec   struct {
				Config map[string]any `yaml:"config"`
			} `yaml:"spec"`
		}
		if err := yaml.Unmarshal(o.Content, &decl); err != nil {
			fmt.Fprintf(os.Stderr, "cored: ConnectorOps %s/%s 原文不可解析（跳过声明注入）: %v\n", o.Scope, o.Name, err)
			continue
		}
		merged := map[string]any{}
		for k, v := range decl.Config {
			merged[k] = v
		}
		for k, v := range decl.Spec.Config {
			merged[k] = v
		}
		if len(merged) == 0 {
			continue
		}
		c, ok := connectors.LookupOpsConnector(o.Name)
		if !ok {
			continue // 驱动未导入=无实现可配（注册表闸门前置——静默跳过非失败）
		}
		ca, ok := c.(interface{ ApplyConfig([]byte) error })
		if !ok {
			continue // 不可配置连接器（codeaudit 等无状态）——零动作
		}
		raw, err := json.Marshal(merged)
		if err != nil {
			continue
		}
		if err := ca.ApplyConfig(raw); err != nil {
			fmt.Fprintf(os.Stderr, "cored: ConnectorOps %s/%s 配置注入失败（显式缺口——op 执行时照实报错）: %v\n", o.Scope, o.Name, err)
		}
	}
}
