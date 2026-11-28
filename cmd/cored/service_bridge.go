// service_bridge.go —— service 节点 session 拉起桥（2026-11-27 用户裁决「builtin 的 service
// 的 agent 可以用 session 来解决。突破下限制」——workflow.md §1.3 / console-ui.md §19.8.9）。
//
// 链路（wake 桥同族——固定载荷零 shell 拼接、loopback、超时 5s）：
//
//	① run 推进落 unit=service 节点 → RunLog.landAfterAdvanceLocked 挂起（suspendWhy=service）
//	   + 拉起请求 → 本桥 SpawnFromRun：解析 service 的 agent（ServiceDesc.instances.agentRefs
//	   文档库现读；未命中=诚实缺席注记不编造）→ 拉起投影落账（ServiceSpawnLog.Dispatch 写穿透）
//	   → 异步 POST 宿主 /service/spawn 桥（CORE_SYSTEM_SERVICE_SPAWN_ADDR）；
//	② 宿主回执 {status, pid, sessionId} → 会话锚回填（AttachSession——sessionId 优先，缺席=
//	   pid:<n> 降级锚，G-08 同词形）；
//	③ 完成面：部署侧 session 退出后经事件注入面发布 ServiceSessionCompleted（scheme 注册类型）
//	   → hub 转发本桥 ServiceSignal：定位该 run 的在途拉起投影翻终态词形（completed；error/fatal
//	   级=failed——失败不续行，run 保持挂起缺口可见）→ 转交 RunLog.ServiceSignal 续行（携带
//	   会话锚——续行 history 行回填真实会话锚）。
//
// 边界：只在 CORE_SYSTEM_SERVICE_SPAWN_ADDR 显式配置时挂接（runOpts.ServiceSpawns 非零才
// 有挂起面——缺席=零翻转口径，service 节点照旧纯查表推进）；HTTP 失败只 stderr 注记不翻转
// 投影行为（投影在账、run 照挂——缺口可见不静默，wake 桥同口径）；出站不执行任何命令，
// 命令由宿主侧配置（task-spawn 桥同红线）。
package main

import (
	"bytes"
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
	"gopkg.in/yaml.v3"
)

// serviceSpawnBridge service 节点 session 拉起桥（workflow.ServiceSpawner +
// workflow.ServiceSignalSink 双能力实现——cored 装配层注入两处：RunOptions.ServiceSpawns /
// TriggerHub.BindServiceSignals）。
type serviceSpawnBridge struct {
	spawns *workflow.ServiceSpawnLog
	runs   *workflow.RunLog      // 完成信号续行面（构造后注入——runs 与 runOpts 循环依赖的解法）
	repo   repository.Repository // ServiceDesc job 文本现读（develop/deploy 作业定义——2026-11-28 用户裁决「developer-system 拉代码 build 接啊」）
	memo   *workflow.RunMemo     // run 上下文账（briefing 实体：planItem/judge 结论——子链会话的输入内容）
	url    string
	hc     *http.Client
	agent  func(service string) (string, bool) // ServiceDesc.instances.agentRefs 解析（装配层注入）
	logf   func(string, ...any)
}

// newServiceSpawnBridge 构建拉起桥（addr 可为 host:port 或 http://host:port；repo/memo 可 nil
// ——nil=briefing 缺席诚实降级，payload 只带锚不带实体）。
func newServiceSpawnBridge(spawns *workflow.ServiceSpawnLog, addr string,
	agentLookup func(string) (string, bool), repo repository.Repository, memo *workflow.RunMemo, logf func(string, ...any)) *serviceSpawnBridge {
	addr = strings.TrimSpace(addr)
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}
	if agentLookup == nil {
		agentLookup = func(string) (string, bool) { return "", false }
	}
	if logf == nil {
		logf = func(string, ...any) {}
	}
	return &serviceSpawnBridge{
		spawns: spawns,
		repo:   repo,
		memo:   memo,
		url:    strings.TrimRight(addr, "/") + "/service/spawn",
		hc:     &http.Client{Timeout: 5 * time.Second},
		agent:  agentLookup,
		logf:   logf,
	}
}

// SpawnFromRun 实现 workflow.ServiceSpawner（非阻塞：投影同步落账，桥异步发——wake 桥同款）。
func (b *serviceSpawnBridge) SpawnFromRun(run workflow.RunRecord, node workflow.NodeView) {
	agent, ok := b.agent(node.Service)
	if !ok {
		b.logf("service 桥 agent 解析缺席（service=%s——ServiceDesc.instances.agentRefs 未命中，投影照落诚实缺席）", node.Service)
	}
	rec := b.spawns.Dispatch(workflow.ServiceSpawnRecord{
		Run: run.Run, Workflow: run.Workflow, Node: node.Name, Phase: node.Phase,
		Service: node.Service, Op: node.Op, Agent: agent, Role: node.Role,
		Context: node.ContextFields, Status: workflow.SpawnSpawned, At: time.Now(),
	})
	if rec.Seq == 0 {
		return // 单飞命中（在途拉起已在账——不重复出站）
	}
	payload := map[string]any{
		"spawnSeq": rec.Seq, "workflow": rec.Workflow, "run": rec.Run,
		"node": rec.Node, "phase": rec.Phase, "service": rec.Service, "op": rec.Op,
		"agent": rec.Agent, "role": rec.Role, "context": rec.Context,
		"involved": run.Involved, // 溯源锚随载荷（2026-11-28 sybran-release-update——派发会话须知道「什么到了货」：GhReleaseEvent involved=Repository/<repo>）
	}
	// 作业定义随载荷（2026-11-28 用户裁决「都做到 developer/deploy-system 的 agent」：
	// develop/deploy 作业文本=ServiceDesc spec.ops[].job 声明面——宿主会话 prompt 组装源）。
	if job, ok := serviceJobLookup(b.repo, rec.Service, rec.Op); ok {
		payload["job"] = job
	}
	// briefing 实体随载荷（2026-11-28 E3 context 实体传递——contextPassing 字段名的真内容）：
	// 子 run=计划项+父 judge 结论；issue 词形 involved 附原文（title+body——会话知道要修什么）。
	if brief := b.briefing(run); brief != nil {
		payload["briefing"] = brief
	}
	go func() {
		body, err := json.Marshal(payload)
		if err != nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, b.url, bytes.NewReader(body))
		if err != nil {
			b.logf("service 桥请求构建失败: %v", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := b.hc.Do(req)
		if err != nil {
			b.logf("service 桥调用失败（workflow=%s run=%d node=%s）: %v", rec.Workflow, rec.Run, rec.Node, err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			b.logf("service 桥拒绝（workflow=%s run=%d node=%s status=%d）", rec.Workflow, rec.Run, rec.Node, resp.StatusCode)
			return
		}
		// 会话锚回执（G-08 同词形）：宿主回执 {status, pid, sessionId}——sessionId 优先，
		// 缺席=pid:<n>（宿主进程句柄最小非空锚——真实会话 id 由宿主面升级后回填词形）。
		var ack struct {
			Status    string `json:"status"`
			PID       int64  `json:"pid"`
			SessionID string `json:"sessionId"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&ack)
		sess := strings.TrimSpace(ack.SessionID)
		if sess == "" && ack.PID > 0 {
			sess = fmt.Sprintf("pid:%d", ack.PID)
		}
		if sess != "" {
			b.spawns.AttachSession(rec.Seq, sess)
		}
		b.logf("service 桥已受理（workflow=%s run=%d node=%s service=%s op=%s agent=%s status=%s session=%s）",
			rec.Workflow, rec.Run, rec.Node, rec.Service, rec.Op, rec.Agent, ack.Status, sess)
	}()
}

// ServiceSignal 实现 workflow.ServiceSignalSink（完成信号面——hub 锁内转发，非阻塞契约）：
// 定位该锚最新 service 挂起 run → 在途拉起投影翻终态词形 → completed 才转交 RunLog 续行
// （failed=run 保持挂起缺口可见）；重复信号/无匹配 run=noop 幂等。
func (b *serviceSpawnBridge) ServiceSignal(rec workflow.TriggerRecord) {
	if b.runs == nil {
		return // 装配序防御（runs 注入前事件不到达——订阅主干晚于装配）
	}
	name, ok := workflow.InvolvedWorkflowRef(rec.Involved)
	if !ok {
		return
	}
	st := b.runs.State()
	runN, node := int64(0), ""
	for i := len(st.Runs) - 1; i >= 0; i-- {
		r := st.Runs[i]
		if rec.TargetRun > 0 && r.Run != rec.TargetRun {
			continue
		}
		if r.Workflow != name || r.Status != workflow.RunSuspended || r.SuspendWhy != workflow.SuspendService {
			continue
		}
		runN = r.Run
		// 节点名按落点相位反查节点投影（spawn 记录键=node.Name 非相位词——
		// Executing≠developer-system；活体教训 2026-11-27：相位词直查 MarkCompleted 落空 noop）。
		node = r.Phase
		for _, nv := range r.Nodes {
			if nv.Phase == r.Phase && nv.Unit == "service" {
				node = nv.Name
				break
			}
		}
		break
	}
	if runN == 0 {
		b.logf("service 桥完成信号无匹配 run（workflow=%s involved=%s——非 service 挂起态或已续行，noop）", name, rec.Involved)
		return
	}
	status := workflow.SpawnCompleted
	if rec.Level == "error" || rec.Level == "fatal" {
		status = workflow.SpawnFailed // 失败退出：run 保持挂起（缺口可见，不冒充可续行）
	}
	spawnRec, sess, ok := b.spawns.MarkCompleted(runN, node, status)
	if !ok {
		b.logf("service 桥完成信号无在途拉起（workflow=%s run=%d node=%s——重复信号或投影已挤出，noop）", name, runN, node)
		return
	}
	_ = spawnRec // 拉起投影终态词形已在账（写穿透随 MarkCompleted）——续行锚=sess
	if status == workflow.SpawnFailed {
		b.logf("service 桥 session 失败退出（workflow=%s run=%d node=%s session=%s——run 保持挂起，缺口可见）", name, runN, node, sess)
		return
	}
	b.runs.ServiceSignal(rec, sess) // 续行（携带会话锚——续行 history 行回填真实会话锚）
}

// serviceJobLookup ServiceDesc spec.ops[].job 作业文本解析（2026-11-28 用户裁决「都做到
// developer/deploy-system 的 agent」——作业定义在 ServiceDesc 层不在 workflow 节点层；节点只
// 声明 service+op）。未挂库/未命中/未声明 job=ok=false（诚实缺席——会话退化为通用词形）。
func serviceJobLookup(repo repository.Repository, service, op string) (string, bool) {
	if repo == nil || service == "" || op == "" {
		return "", false
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	objs, err := repo.List(ctx, "ServiceDesc", "")
	if err != nil {
		return "", false
	}
	var raw []byte
	for _, o := range objs {
		if o.Name == service {
			raw = o.Content
			if o.Scope == "global" {
				break // global 同名优先（serviceAgentLookup 同款确定性投影）
			}
		}
	}
	if raw == nil {
		return "", false
	}
	var doc struct {
		Spec struct {
			Ops []struct {
				Op  string `yaml:"op"`
				Job string `yaml:"job"`
			} `yaml:"ops"`
		} `yaml:"spec"`
	}
	if yaml.Unmarshal(raw, &doc) != nil {
		return "", false
	}
	for _, o := range doc.Spec.Ops {
		if o.Op == op && strings.TrimSpace(o.Job) != "" {
			return o.Job, true
		}
	}
	return "", false
}

// briefing 会话输入实体组装（E3 context 实体传递——contextPassing 声明字段的真内容）：
// planItem=子 run 计划项（fanout 桥 ChildStarted 绑定）；judge=父 run 合判结论；
// issue=involved issue 词形的原文（title+body——会话知道要修什么，不编造）。全空=nil
// （payload 不带 briefing 键——诚实缺席）。
func (b *serviceSpawnBridge) briefing(run workflow.RunRecord) map[string]any {
	brief := map[string]any{}
	if b.memo != nil {
		if m := b.memo.GetMap(run.Run, "planItem"); m != nil {
			brief["planItem"] = m
		}
		if run.ParentRun > 0 {
			if s := b.memo.GetString(run.ParentRun, "planSummary"); s != "" {
				brief["judge"] = s
			}
		}
		if s := b.memo.GetString(run.Run, "mergeSummary"); s != "" {
			brief["merge"] = s
		}
	}
	if repo, num, hit := involvedRepoIssue(run.Involved); hit && num > 0 {
		if subj := ghIssueText(b.repo, repo, num); subj != "" {
			brief["issue"] = subj
		}
	}
	if len(brief) == 0 {
		return nil
	}
	return brief
}

// ghIssueText GitHub REST 拉 issue 原文（title+body——approvaldiff.judge.go readSubject 的
// cored 侧同源形；org 前缀由 ConnectorOps/global/connectops-gh config.owner 补全，零硬编码；
// 10s 超时；GH_TOKEN 在场带认证头。失败=空串诚实缺席）。
func ghIssueText(src repository.Repository, repo string, number int64) string {
	if repo == "" || number <= 0 {
		return ""
	}
	if !strings.Contains(repo, "/") {
		if owner := ghOwnerOf(src); owner != "" {
			repo = owner + "/" + repo
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet,
		fmt.Sprintf("https://api.github.com/repos/%s/issues/%d", repo, number), nil)
	if err != nil {
		return ""
	}
	if tok := strings.TrimSpace(os.Getenv("GH_TOKEN")); tok != "" {
		req.Header.Set("Authorization", "Bearer "+tok)
	}
	resp, err := (&http.Client{}).Do(req)
	if err != nil {
		return ""
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return ""
	}
	var d struct {
		Title string `json:"title"`
		Body  string `json:"body"`
	}
	if err := json.NewDecoder(io.LimitReader(resp.Body, 1<<20)).Decode(&d); err != nil {
		return ""
	}
	return fmt.Sprintf("#%d %s\n%s", number, d.Title, d.Body)
}

// ghOwnerOf org 前缀解析（ConnectorOps/global/connectops-gh 的 config.owner——judge.go
// ghOwner 同款行级读，零 yaml 依赖零硬编码；未挂库=空串）。
func ghOwnerOf(src repository.Repository) string {
	if src == nil {
		return ""
	}
	o, err := src.Get(context.Background(), "ConnectorOps", "global", "connectops-gh")
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

// serviceAgentLookup ServiceDesc.instances.agentRefs 解析（装配侧注入——文档库现读零缓存，
// workflowDocLookup 同款口径；首个 instance 的首个 agentRef=service 的执行 agent 投影）。
// 未挂库/未命中=ok=false（诚实缺席——投影照落 agent 空，不编造）。
func serviceAgentLookup(repo repository.Repository) func(string) (string, bool) {
	if repo == nil {
		return func(string) (string, bool) { return "", false }
	}
	return func(service string) (string, bool) {
		if service == "" {
			return "", false
		}
		objs, err := repo.List(context.Background(), "ServiceDesc", "")
		if err != nil {
			return "", false
		}
		var raw []byte
		for _, o := range objs {
			if o.Name == service {
				raw = o.Content
				if o.Scope == "global" {
					break // global 同名优先（workflowDocLookup 同款确定性投影）
				}
			}
		}
		if raw == nil {
			return "", false
		}
		var doc struct {
			Instances []struct {
				AgentRefs []string `yaml:"agentRefs"`
			} `yaml:"instances"`
		}
		if yaml.Unmarshal(raw, &doc) != nil {
			return "", false
		}
		for _, inst := range doc.Instances {
			for _, a := range inst.AgentRefs {
				if a = strings.TrimSpace(a); a != "" {
					return a, true
				}
			}
		}
		return "", false
	}
}
