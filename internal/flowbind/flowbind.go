// Package flowbind —— workflow 执行模型绑定面（v0.49，R-7）。
// 设计依据：core_design concepts/workflow.md §1.3（执行模型：节点=session/agent、默认绑定
// 同名 ConnectorOps、apply 即实现）+ connector.md 规则 6（ConnectorOps=节点流转实现面，
// 未绑定=决策面 execution=unbound 缺口可见）+ instance 仓 runtime/workflow-runtime.md §11.1
// （落地形态二分：骨架+节点接线=cored 确定性机械面；实现代码=apply-workflow agent 面）。
//
// 四件：
//   - Resolve：workflow 对象 → 绑定解析（spec.connector 归一化/缺省同名）+ 生效相位 →
//     flow.nodes 节点骨架投影（name/title/unit/op/connector/provider——unit 声明缺省 session，
//     跨服务=agent 调 agent 由节点声明 unit: agent 表达；声明面外不即兴）。
//   - Skeleton：绑定 → 同名 ConnectorOps 对象内容（确定性：workservice=workflow-engine 宿主
//   - workflow=<锚名> + ops=[flow] + flow.nodes + labels[core/origin]=workflow-bind 所有面标注）。
//   - Materializer：装配期存量枚举初值 + DocApplied×Workflow 订阅热路径（metrics 阈值桥
//     v0.47 同款范式）——物化走 apply.Engine 同一三步闸门（fieldManager=workflow:bind）；
//     幂等=canonical 比对 noop 零提交；spec 变更 → rv+1 diff 可查（apply_log/ history 面）；
//     **所有面**：只写 origin=workflow-bind 标注的对象（自己物化的骨架），人/agent 已写的
//     ConnectorOps 实现零盲写；单对象失败=stderr 告警继续其余（保旧面，不打死写路径）。
//   - Dispatch（dispatch.go）：派单受理投影（v0.50 R-8）——一次真实物化动作（created|updated）
//     即一条对 apply-workflow 锚的派单受理投影（workflow-runtime.md §11.2 拉起形态首段：
//     受理≠拉起，真实拉起 declared-deferred）。
//
// 解耦口径：本包只依赖 apply/repository 既有面，零新写路径零旁路——绑定物化=一次普通
// apply（引擎准入+rv 乐观锁+DocApplied+审计账全量继承）。
package flowbind

import (
	"context"
	"errors"
	"sort"
	"strings"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/apply"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"gopkg.in/yaml.v3"
)

// 所有面标注（metadata.labels[core/origin]）：物化器只写带本标注的对象——人/agent 已写的
// ConnectorOps 实现零盲写（workflow-runtime.md §11.1 所有面语义）。
const OriginLabel = "workflow-bind"

// fieldManager 物化器写面溯源串（apply.md rule 2；workflow: 前缀=非 console 域，决策面可判别）。
const fieldManager = "workflow:bind"

// FlowWorkservice 骨架 ConnectorOps 的宿主 WorkService（docsRoot global/workservices/
// workflow-engine——flow ConnectorOps 实现落点声明，impl ref=workservice 仓 workflow 包）。
const FlowWorkservice = "workflow-engine"

// FlowOp 骨架声明的唯一 op（宿主 WorkService.tools 同名声明——spec.ops ⊆ tools 惯例）。
const FlowOp = "flow"

// Node 节点骨架行（spec.flow.nodes 元素；scheme.yaml ConnectorOps spec.flow 声明同词汇）。
type Node struct {
	Name      string `yaml:"name" json:"name"`
	Title     string `yaml:"title,omitempty" json:"title,omitempty"`
	Unit      string `yaml:"unit" json:"unit"` // session|agent（声明缺省 session——§1.3 执行单元）
	Op        string `yaml:"op,omitempty" json:"op,omitempty"`
	Connector string `yaml:"connector,omitempty" json:"connector,omitempty"`
	Provider  string `yaml:"provider,omitempty" json:"provider,omitempty"`
}

// Binding 绑定解析结果（Resolve 输出）。
type Binding struct {
	Workflow string // workflow 锚名
	Scope    string // workflow 对象 scope（骨架落点 scope）
	Connecto string // 绑定 ConnectorOps 名（spec.connector 归一化/缺省同名）
	Nodes    []Node // 节点流转骨架（生效相位投影）
}

// wfDecl workflow 对象声明投影（宽松解析：读面不放大声明形变——写路径 scheme 校验是形态权威）。
type wfDecl struct {
	Metadata struct {
		Name   string            `yaml:"name"`
		Scope  string            `yaml:"scope"`
		Labels map[string]string `yaml:"labels"`
	} `yaml:"metadata"`
	Spec struct {
		Connector string      `yaml:"connector"`
		Phases    []yaml.Node `yaml:"phases"`
	} `yaml:"spec"`
}

// phaseNode 富相位行/字符串行 → 节点骨架行（字符串行=name、unit 缺省 session；富行=name/title/
// unit/op/connector/provider 逐字段投影——声明面外字段不即兴）。零名行跳过（不可寻址不投影）。
func phaseNode(n yaml.Node) (Node, bool) {
	if n.Kind == yaml.ScalarNode {
		if n.Value == "" {
			return Node{}, false
		}
		return Node{Name: n.Value, Unit: "session"}, true
	}
	if n.Kind == yaml.MappingNode {
		var out Node
		for i := 0; i+1 < len(n.Content); i += 2 {
			k, v := n.Content[i].Value, n.Content[i+1]
			if v.Kind != yaml.ScalarNode {
				continue
			}
			switch k {
			case "name":
				out.Name = v.Value
			case "title":
				out.Title = v.Value
			case "unit":
				out.Unit = v.Value // session|agent（声明词汇；非法值交写路径闸门，投影原样）
			case "op":
				out.Op = v.Value
			case "connector":
				out.Connector = v.Value
			case "provider":
				out.Provider = v.Value
			}
		}
		if out.Name == "" {
			return Node{}, false
		}
		if out.Unit == "" {
			out.Unit = "session" // 执行单元缺省（§1.3：global/服务内=session）
		}
		return out, true
	}
	return Node{}, false
}

// normalizeConnector spec.connector 值归一化：`connector-ops/<name>` 前缀剥离（ref kind 前缀
// 惯例），裸名原样（connector.md 命名约定：缺省=同名）。空值=空（调用方兜同名）。
func normalizeConnector(v string) string {
	if name, ok := strings.CutPrefix(v, "connector-ops/"); ok {
		return name
	}
	return v
}

// Resolve 绑定解析（纯函数）：workflow 对象原文 → 绑定名+节点骨架。不可解析/零相位=ok=false
// （读面不放大声明形变；物化器跳过该对象——写路径校验是形态权威）。
func Resolve(scope, name string, content []byte) (Binding, bool) {
	var d wfDecl
	if err := yaml.Unmarshal(content, &d); err != nil {
		return Binding{}, false
	}
	b := Binding{Workflow: name, Scope: scope}
	if c := normalizeConnector(d.Spec.Connector); c != "" {
		b.Connecto = c
	} else {
		b.Connecto = name // 缺省同名（connector.md 规则 6 命名约定）
	}
	for i := range d.Spec.Phases {
		if n, ok := phaseNode(d.Spec.Phases[i]); ok {
			b.Nodes = append(b.Nodes, n)
		}
	}
	if len(b.Nodes) == 0 {
		return b, false // 零可投影相位=骨架无节点可接线（builtin 注册表锚无文档对象同此形——跳过）
	}
	return b, true
}

// flowDecl 骨架对象声明块（yaml 序出确定性形态）。
type flowDecl struct {
	Nodes []Node `yaml:"nodes"`
}

// connSpec 骨架 spec。
type connSpec struct {
	Workservice string   `yaml:"workservice"`
	Workflow    string   `yaml:"workflow"`
	Ops         []string `yaml:"ops"`
	Flow        flowDecl `yaml:"flow"`
}

// connDoc 骨架对象全文。
type connDoc struct {
	APIVersion string   `yaml:"apiVersion"`
	Kind       string   `yaml:"kind"`
	Metadata   connMeta `yaml:"metadata"`
	Spec       connSpec `yaml:"spec"`
}

type connMeta struct {
	Name   string            `yaml:"name"`
	Scope  string            `yaml:"scope"`
	Labels map[string]string `yaml:"labels"`
}

// Skeleton 绑定 → 同名 ConnectorOps 对象内容（确定性 YAML；rv=0 盲写形——写路径幂等比对去 rv）。
func Skeleton(b Binding) []byte {
	doc := connDoc{
		APIVersion: "core/v1",
		Kind:       "ConnectorOps",
		Metadata: connMeta{
			Name:   b.Connecto,
			Scope:  b.Scope,
			Labels: map[string]string{"core/origin": OriginLabel},
		},
		Spec: connSpec{
			Workservice: FlowWorkservice,
			Workflow:    b.Workflow,
			Ops:         []string{FlowOp},
			Flow:        flowDecl{Nodes: b.Nodes},
		},
	}
	var sb strings.Builder
	sb.WriteString("# workflow 执行模型绑定骨架（cored 绑定物化器维护——workflow-runtime.md §11.1；\n")
	sb.WriteString("# origin=workflow-bind=所有面标注；真实节点实现=apply-workflow agent 面，运行时不做代码即兴）\n")
	enc := yaml.NewEncoder(&sb)
	enc.SetIndent(2)
	_ = enc.Encode(&doc)
	_ = enc.Close()
	return []byte(sb.String())
}

// Materializer 绑定物化器（装配层持有；DocApplied×Workflow 热路径 + Ensure 存量枚举复用）。
// Dispatch=派单投影落点（v0.50 R-8，装配层注入；nil=静默零派单）——一次真实物化动作
// （created|updated）即一条对 apply-workflow 锚的派单受理投影（§11.2 拉起形态首段；
// noop/skipped 不派单——无新实现可做）。
type Materializer struct {
	Repo     repository.Repository
	Apply    *apply.Engine
	Dispatch DispatchSink
}

// involvedRef 三段同构拆解（kind/scope/name——event.md 规则 11 involvedRef 惯例）。
func involvedRef(involved string) (kind, scope, name string, ok bool) {
	parts := strings.Split(involved, "/")
	if len(parts) != 3 || parts[0] == "" || parts[1] == "" || parts[2] == "" {
		return "", "", "", false
	}
	return parts[0], parts[1], parts[2], true
}

// originOf 对象现有 origin 标注（labels.core/origin；解析失败=空）。
func originOf(content []byte) string {
	var d wfDecl
	if err := yaml.Unmarshal(content, &d); err != nil {
		return ""
	}
	return d.Metadata.Labels["core/origin"]
}

// Ensure 单锚物化（幂等）：绑定解析 → 对象缺席=创建 / 在册且 origin=workflow-bind=canonical
// 比对（异=rv 乐观锁更新、同=noop）/ 在册非本物化器所有=零盲写跳过。
// 返回 (action, error)：action ∈ created|updated|noop|skipped。
// created/updated 成功即派单（v0.50 R-8 §11.2）：骨架物化=「该版本实现未落」的机械签名 →
// apply-workflow 锚派单受理投影（rv=落库后版本、eventId=DocApplied 回执——实现侧 diff 基线
// 与来源链可回指）；noop（幂等重入）/skipped（非本物化器所有面）不派单。
func (m *Materializer) Ensure(ctx context.Context, scope, name string, wfContent []byte) (string, error) {
	b, ok := Resolve(scope, name, wfContent)
	if !ok {
		return "skipped", nil
	}
	existing, err := m.Repo.Get(ctx, "ConnectorOps", b.Scope, b.Connecto)
	switch {
	case err == nil:
		if originOf(existing.Content) != OriginLabel {
			return "skipped", nil // 非物化器所有面（人/agent 实现）——零盲写
		}
		desired := Skeleton(b)
		doc := &repository.Object{Kind: "ConnectorOps", Name: b.Connecto, Scope: b.Scope,
			ResourceVersion: existing.ResourceVersion, Content: desired}
		res, err := m.Apply.Apply(ctx, doc, apply.Options{FieldManager: fieldManager,
			ReasonRef: "workflow-bind:" + b.Workflow})
		if err != nil {
			return "", err
		}
		if res.Action == "noop" {
			return "noop", nil
		}
		m.dispatch(res, "updated", b)
		return "updated", nil
	case errors.Is(err, repository.ErrNotFound):
		doc := &repository.Object{Kind: "ConnectorOps", Name: b.Connecto, Scope: b.Scope,
			ResourceVersion: 0, Content: Skeleton(b)}
		res, err := m.Apply.Apply(ctx, doc, apply.Options{FieldManager: fieldManager,
			ReasonRef: "workflow-bind:" + b.Workflow})
		if err != nil {
			return "", err
		}
		m.dispatch(res, "created", b)
		return "created", nil
	default:
		return "", err
	}
}

// dispatch 记一条派单受理投影（物化成功后调用；sink 缺席=静默）。eventID 取 DocApplied
// 回执（create/update 非 nil；防御式零值容忍——回执缺席不阻派单）。
func (m *Materializer) dispatch(res *apply.Result, action string, b Binding) {
	if m.Dispatch == nil || res == nil {
		return
	}
	var eventID string
	if res.Event != nil {
		eventID = res.Event.ID
	}
	m.Dispatch.Dispatch(DispatchRecord{
		Agent: DispatchAgent, Status: DispatchAccepted, AutoLevel: DispatchAutoLevel,
		Action: action, Kind: DispatchKindWorkflow, Workflow: b.Workflow, Scope: b.Scope,
		Connector: b.Connecto, RV: res.RV, EventID: eventID, At: time.Now(),
	})
}

// Fail 单锚物化失败（stderr 归装配层——包零日志依赖，metricsrc 钩子同款解耦）。
type Fail struct {
	Scope string
	Name  string
	Err   error
}

// EnsureAll 存量枚举初值（装配期一次）：全部 Workflow 对象逐锚 Ensure。返回 n=物化动作数
// （created+updated；noop/skipped 不计）+ fails=逐锚失败清单（保旧面语义——单锚失败不打死其余）。
func (m *Materializer) EnsureAll(ctx context.Context) (int, []Fail) {
	return m.ensure(ctx)
}

// ensure 枚举主体（装配初值与热路径同一逻辑——装到最新面幂等）。
func (m *Materializer) ensure(ctx context.Context) (int, []Fail) {
	if m == nil || m.Repo == nil || m.Apply == nil {
		return 0, nil
	}
	objs, err := m.Repo.List(ctx, "Workflow", "")
	if err != nil {
		return 0, []Fail{{Err: err}}
	}
	n := 0
	var fails []Fail
	for _, o := range objs {
		act, err := m.Ensure(ctx, o.Scope, o.Name, o.Content)
		if err != nil {
			fails = append(fails, Fail{Scope: o.Scope, Name: o.Name, Err: err})
			continue
		}
		if act == "created" || act == "updated" {
			n++
		}
	}
	return n, fails
}

// Start 订阅热路径（DocApplied×Workflow 精确过滤——非 Workflow 对象 apply 零无效枚举）。
// 连续变更积压排干合并一次全量枚举（装到最新面）；阻塞调用，装配层 go 例程承载。
// onApplied(n)=热路径物化动作数（stderr 装配锚）；onErr=枚举/单锚失败（保旧面+下次 apply 重试自愈）。
func (m *Materializer) Start(ctx context.Context, bus *eventbus.Bus, onApplied func(n int), onErr func(fails []Fail)) {
	if m == nil || bus == nil {
		return
	}
	ch, cancel := bus.Subscribe("DocApplied")
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		case ev := <-ch:
			kind, _, _, ok := involvedRef(ev.Involved)
			if !ok || kind != "Workflow" {
				continue // 精确过滤：仅 Workflow 对象 apply 触发（WorkService 热重绑同款）
			}
			// 积压排干合并：连续 apply 一次全量枚举装到最新面（O(文档数) 幂等 noop 便宜）。
		drain:
			for {
				select {
				case <-ch:
				default:
					break drain
				}
			}
			n, fails := m.ensure(ctx)
			if len(fails) > 0 && onErr != nil {
				onErr(fails)
			}
			if n > 0 && onApplied != nil {
				onApplied(n)
			}
		}
	}
}

// ───────────────────────── 决策面 execution 投影（聚合面消费） ─────────────────────────

// Execution 聚合面 execution 投影（/v1/workflows 行集键——cored-bridge.md §2 v0.49）。
type Execution struct {
	State          string `json:"state"`          // bound|unbound（connector.md 规则 6：缺口不静默）
	Connector      string `json:"connector"`      // 绑定 ConnectorOps 名（spec.connector 覆写/缺省同名）
	ConnectorScope string `json:"connectorScope"` // 对象 scope（bound 时；unbound=空）
	Skeleton       bool   `json:"skeleton"`       // true=对象仍为物化器骨架（origin=workflow-bind）——实现代码未落
}

// ExecUnbound 决策面 unbound 词（与 workflow.ExecUnbound 同词汇表——缺口可见）。
const ExecUnbound = "unbound"

// ExecBound 决策面 bound 词。
const ExecBound = "bound"

// connMetaMinimal ConnectorOps 对象最小投影（判定用；宽松解析）。
type connMetaMinimal struct {
	Metadata struct {
		Labels map[string]string `yaml:"labels"`
	} `yaml:"metadata"`
	Spec struct {
		Workservice string `yaml:"workservice"`
		Workflow    string `yaml:"workflow"`
	} `yaml:"spec"`
}

// connRows ConnectorOps 对象按名分组索引（聚合面一次枚举逐锚复用）。
type connRows map[string][]*repository.Object

// indexConnRows 跨 scope 枚举 → 按名分组（scope 字典序确定性——解析取序恒定）。
// IndexRows 跨 scope 枚举 → 按名分组索引（scope 字典序确定性——解析取序恒定）。聚合面一次枚举逐锚复用。
func IndexRows(objs []*repository.Object) connRows {
	out := connRows{}
	for _, o := range objs {
		out[o.Name] = append(out[o.Name], o)
	}
	for n := range out {
		rows := out[n]
		sort.SliceStable(rows, func(i, j int) bool { return rows[i].Scope < rows[j].Scope })
		out[n] = rows
	}
	return out
}

// executionOf 单锚 execution 投影：绑定名（spec.connector 覆写/缺省同名）→ 对象查找
// （workflow scope 优先 → global → 其余 scope 首个——确定性投影）；缺席=unbound。
// 绑定名解析失败（不可解析声明）=同名兜底仍可判——对象缺席照实 unbound。
// ExecutionOf 单锚 execution 投影（聚合面消费）：绑定名（spec.connector 覆写/缺省同名）→ 对象查找（workflow scope 优先 → global → 其余 scope 首个——确定性投影）；缺席=unbound。
// 绑定名解析失败（不可解析声明）=同名兜底仍可判——对象缺席照实 unbound。
func ExecutionOf(wfScope, wfName string, wfContent []byte, rows connRows) *Execution {
	conn := wfName
	if b, ok := Resolve(wfScope, wfName, wfContent); ok {
		conn = b.Connecto
	}
	for _, cand := range []string{wfScope, "global"} {
		if o := lookup(rows, conn, cand); o != nil {
			return execRow(conn, o)
		}
	}
	if rs := rows[conn]; len(rs) > 0 {
		return execRow(conn, rs[0])
	}
	return &Execution{State: ExecUnbound, Connector: conn}
}

func lookup(rows connRows, name, scope string) *repository.Object {
	for _, o := range rows[name] {
		if o.Scope == scope {
			return o
		}
	}
	return nil
}

func execRow(conn string, o *repository.Object) *Execution {
	e := &Execution{State: ExecBound, Connector: conn, ConnectorScope: o.Scope}
	var d connMetaMinimal
	if err := yaml.Unmarshal(o.Content, &d); err == nil {
		e.Skeleton = d.Metadata.Labels["core/origin"] == OriginLabel
	}
	return e
}
