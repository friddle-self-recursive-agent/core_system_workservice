// objspec.go —— 对象 spec 解析面（v0.27 对象声明表解析首段=表投影；v0.28 相位投影+迁移
// 判定接线，workflows 规则 16 下一段）。
// 规则 14 显式迁移表首段（v0.25）只物化了 builtin 表；平台 Workflow 对象的自声明
// （spec.phases 相位序列+spec.transitions 迁移表，scheme.yaml 声明终态形态）随文档面存在。
// v0.27 把声明表首段落为**投影**；v0.28 把相位序列一并投影——投影即点亮判定：相位投影后
// 受理首步/审批信号/异步推进三条判定路径对 object 源自然生效（判定器零改动——stepByTable
// 纯查表落点通用，推进器已完全表驱动，接表即得全量推进语义）。
//
// 零投影口径不稀释：未声明 phases 的对象=零相位投影（phases 恒空→status 恒 triggered，
// 三条判定路径零动作防御面原样生效）——相位投影的闸门=声明面，不在引擎侧特判。
//
// 解耦口径：引擎不感知文档库——解析实现由装配层注入（RunOptions.Specs），本包只定义
// 解析面契约与声明→投影的纯映射；未注入（nil）或未命中=零投影（既有行为面不变）。
package workflow

import (
	"time"

	"gopkg.in/yaml.v3"
)

// PostActionView 收尾归档动作声明投影（Reporting 相位通用解释器动作：如临时 BugDoc 标记失效 invalidate /
// 删除 delete / 保留不删除 keep，或回写 ServiceInstance.status.phase=InSync；用户可自由添加/删除条目）。
type PostActionView struct {
	ID            string `json:"id" yaml:"id"`
	Kind          string `json:"kind" yaml:"kind"`
	TargetDocType string `json:"targetDocType,omitempty" yaml:"targetDocType"`
	Action        string `json:"action,omitempty" yaml:"action"` // invalidate|delete|keep
	TargetPhase   string `json:"targetPhase,omitempty" yaml:"targetPhase"`
	Enabled       bool   `json:"enabled" yaml:"enabled"`
	Title         string `json:"title,omitempty" yaml:"title"`
}

// ApprovalPolicy 审批策略声明投影（2026-11-28 用户裁定「自审批真判」——spec.approval
// mode/autoCriteria/maxRiskLevel 随对象 spec 一次解析随行；mode=agent 不再是盲放行，
// 引擎按 autoCriteria 拉 ops_core_inner_agent 真判决，maxRiskLevel 超限回退人审）。
type ApprovalPolicy struct {
	Mode         string `json:"mode,omitempty"`         // human（缺省）| auto（全自动放行）| agent（真判）
	AutoCriteria string `json:"autoCriteria,omitempty"` // agent 模式判决准则（判决小 agent 的 user prompt 输入）
	MaxRiskLevel string `json:"maxRiskLevel,omitempty"` // low|medium|high（缺省 low；超限转人工）
	DiffEnabled  bool   `json:"diffEnabled,omitempty"`  // spec.approvalDiff.enabled（缺省 true）
	DiffAgent    string `json:"diffAgent,omitempty"`    // spec.approvalDiff.agent（缺省 ops_core_inner_agent）
	DiffProvider string `json:"diffProvider,omitempty"` // spec.approvalDiff.provider（缺省 openai-default）
}

// FanoutView 扇出声明投影（2026-11-28 用户裁定第四词形 unit=workflow——节点按
// fanout.from/key 读执行计划，扇出 N 个子 run，父 run join 挂起等全部子 run 终态后
// fanout-complete 迁移到 merge-results 汇聚）。
type FanoutView struct {
	From     string `json:"from,omitempty"`     // 执行计划文档名/字段（batch-judge 产物）
	Key      string `json:"key,omitempty"`      // 计划项数组键（缺省 items）
	Entry    string `json:"entry,omitempty"`    // 子 run 入口事件（缺省 sub:plan-item）
	Workflow string `json:"workflow,omitempty"` // 子 run 目标 workflow（缺省父 workflow）
}

// NodeView 节点声明投影（2026-11-27 service session 裁决「builtin 的 service 的 agent 可以用
// session 来解决」——引擎消费的最小字段集）：unit=service 的节点在 run 推进落点触发
// ServiceSpawn session 拉起面（serviceexec.go）；其余声明字段（refDocs/contextPassing/…）
// 对象 spec 透传不在本投影（消费方=文档面/前端节点表）。
type NodeView struct {
	Name          string         `json:"name"`                    // 节点名（富声明 name；标量相位=相位名同词）
	Phase         string         `json:"phase"`                   // 相位锚（run.Phase 同词域）
	Unit          string         `json:"unit,omitempty"`          // connectops|agent|service|workflow|gate（空=未声明——旧词形透传）
	Executor      string         `json:"executor,omitempty"`      // dsh（缺省，重型通路）| inner（2026-11-28 用户裁定「低智商要求的可以使用内置 agent」——引擎内置轻量 LLM 通路 ops_core_inner_agent，不拉 session 零拉起开销）
	Service       string         `json:"service,omitempty"`       // unit=service：目标系统名（developer-system…）
	Op            string         `json:"op,omitempty"`            // 调用 op（develop|acceptance|release…）
	Connector     string         `json:"connector,omitempty"`     // unit=connectops：ops 连接器名（connectops-gh/connectops-docs…——2026-11-28 opsrun.go 执行面消费，新载荷键=代际判别器）
	Params        map[string]any `json:"params,omitempty"`        // unit=connectops：op 静态参数（声明值——owner/repo/tag 等；动态溯源=实现侧按 run.Involved 派生）
	Role          string         `json:"role,omitempty"`          // permissionRole（P7 权限随文档——随载荷透传给执行通路）
	ContextFields []string       `json:"contextFields,omitempty"` // contextPassing.fields（上一轮上下文声明字段——首轮注入载荷）
	EntryEvents   []string       `json:"entryEvents,omitempty"`   // 入口事件（多入口投影，与 entry: 迁移行同源）
	Fanout        *FanoutView    `json:"fanout,omitempty"`        // unit=workflow 扇出声明（第四词形）
}

// SpecView 对象 spec 投影（v0.28 相位投影段——同一份对象声明的两面一次解析：相位序列
// +声明迁移表；2026-11-27 扩展富节点、多入口 entryPoints、agent/auto 审批、postActions 解释器声明；
// 2026-11-27 service session 裁决再扩 Nodes 节点投影——新载荷键=代际判别器）。
type SpecView struct {
	Phases       []string         // 声明相位序列投影（spec.phases 原样顺序保留；空=未声明）
	Transitions  []TransitionRow  // 声明迁移表投影（spec.transitions 原样顺序保留；空=未声明）
	ApprovalMode string           // 审批策略投影（spec.approval.mode=auto|agent 时闸锚自动通过；空=human 缺省）
	Approval     ApprovalPolicy   // 审批策略全投影（autoCriteria/maxRiskLevel/approvalDiff——自审批真判消费）
	BugDocPolicy string           // 临时 BugDoc 收尾策略投影（invalidate|delete|keep）
	PostActions  []PostActionView // 解释器通用收尾动作清单
	Nodes        []NodeView       // 节点声明投影（富节点 unit/service/op/role——service session 拉起面消费）
}

// SpecResolver 对象 spec 解析面（对象声明表解析 v0.27 首段+相位投影 v0.28，workflows
// 规则 16）：按 workflow 锚名解析平台 Workflow 对象的自声明 spec。可选能力（挂起-恢复
// ResumeSink 同款可选能力范式）：装配层未注入=能力缺席，object 源零投影（相位/表皆空
// ——缺口可见口径保持，既有行为面不变）。
type SpecResolver interface {
	// ResolveSpec 返回对象 spec 投影（零缓存：每次现读现投——文档库 git 后端权威，apply
	// 落库即生效热装载跟随）。零值 SpecView=未命中/损坏文档/未声明（零投影，不冒充空声明）。
	ResolveSpec(name string) SpecView
}

// SpecFunc 函数适配器（装配层零样板——Lookup 闭包直接成解析面）。
type SpecFunc func(name string) SpecView

// ResolveSpec 实现 SpecResolver。
func (f SpecFunc) ResolveSpec(name string) SpecView { return f(name) }

// docsSpecDecl 文档原文的对象 spec 声明形态（支持标量相位名或富节点声明 {name, phase, entryEvents, ...}）。
type docsSpecDecl struct {
	Spec struct {
		Phases      []yaml.Node `yaml:"phases"`
		EntryPoints []struct {
			Node    string   `yaml:"node"`
			Phase   string   `yaml:"phase"`
			Enabled *bool    `yaml:"enabled"`
			Events  []string `yaml:"events"`
		} `yaml:"entryPoints"`
		Transitions []struct {
			From string `yaml:"from"`
			On   string `yaml:"on"`
			To   string `yaml:"to"`
		} `yaml:"transitions"`
		Approval struct {
			Mode         string `yaml:"mode"` // auto|agent|human（其余词形=未声明，human 缺省）
			AutoCriteria string `yaml:"autoCriteria"`
			MaxRiskLevel string `yaml:"maxRiskLevel"`
		} `yaml:"approval"`
		ApprovalDiff struct {
			Enabled  *bool  `yaml:"enabled"`
			Agent    string `yaml:"agent"`
			Provider string `yaml:"provider"`
		} `yaml:"approvalDiff"`
		Regression struct {
			BugDocPolicy    string           `yaml:"bugDocPolicy"` // invalidate|delete|keep
			CleanTempBugDoc *bool            `yaml:"cleanTempBugDoc"`
			PostActions     []PostActionView `yaml:"postActions"`
		} `yaml:"regression"`
	} `yaml:"spec"`
}

// parsePhaseNode 解析单条相位节点（标量字符串或富节点 map）→ NodeView（ok=false=不可投影）。
func parsePhaseNode(n yaml.Node) (NodeView, bool) {
	if n.Kind == yaml.ScalarNode {
		if n.Value == "" {
			return NodeView{}, false
		}
		return NodeView{Name: n.Value, Phase: n.Value}, true
	}
	if n.Kind == yaml.MappingNode {
		var m struct {
			Name           string         `yaml:"name"`
			Phase          string         `yaml:"phase"`
			Title          string         `yaml:"title"`
			Unit           string         `yaml:"unit"`
			Service        string         `yaml:"service"`
			Op             string         `yaml:"op"`
			Connector      string         `yaml:"connector"`
			Params         map[string]any `yaml:"params"`
			Executor       string         `yaml:"executor"`
			Role           string         `yaml:"permissionRole"`
			ContextPassing struct {
				PreviousRound bool     `yaml:"previousRound"`
				Fields        []string `yaml:"fields"`
			} `yaml:"contextPassing"`
			EntryEvents []string `yaml:"entryEvents"`
			Fanout      struct {
				From     string `yaml:"from"`
				Key      string `yaml:"key"`
				Entry    string `yaml:"entry"`
				Workflow string `yaml:"workflow"`
			} `yaml:"fanout"`
		}
		if err := n.Decode(&m); err != nil {
			return NodeView{}, false
		}
		p := m.Phase
		if p == "" {
			p = m.Name
		}
		if p == "" {
			p = m.Title
		}
		var fanout *FanoutView
		if m.Fanout.From != "" || m.Fanout.Key != "" || m.Fanout.Entry != "" || m.Fanout.Workflow != "" {
			fanout = &FanoutView{From: m.Fanout.From, Key: m.Fanout.Key, Entry: m.Fanout.Entry, Workflow: m.Fanout.Workflow}
		}
		return NodeView{Name: m.Name, Phase: p, Unit: m.Unit, Service: m.Service, Op: m.Op, Connector: m.Connector,
			Params: m.Params, Role: m.Role, Executor: m.Executor, ContextFields: m.ContextPassing.Fields,
			EntryEvents: m.EntryEvents, Fanout: fanout}, p != ""
	}
	return NodeView{}, false
}

// DocsSpecs 文档面 spec 解析（装配适配）：lookup 由装配层注入（文档库读面按名取 Workflow
// 对象原文），本包只做 spec.phases/spec.transitions 声明 → 投影的映射（引擎不感知 git
// 后端）。零缓存：每次解析现读现投（单次 lookup——相位与表同源同拍）。防御面：lookup
// 未命中/原文不可解析/未声明=零投影（零值 SpecView）。
type DocsSpecs struct {
	Lookup func(name string) (raw []byte, ok bool)
}

// ResolveSpec 实现 SpecResolver：声明→投影纯映射（相位序列与迁移表行 from/on/to 原样，
// 顺序保留；同时支持富节点 phases 与多入口 entryPoints 解释器投影）。
func (d DocsSpecs) ResolveSpec(name string) SpecView {
	if d.Lookup == nil || name == "" {
		return SpecView{}
	}
	raw, ok := d.Lookup(name)
	if !ok {
		return SpecView{}
	}
	var doc docsSpecDecl
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return SpecView{} // 损坏文档不投影（防御性——scheme 校验是写路径权威闸门）
	}
	var view SpecView
	switch doc.Spec.Approval.Mode {
	case "auto", "agent":
		view.ApprovalMode = doc.Spec.Approval.Mode
	}
	view.Approval = ApprovalPolicy{
		Mode: doc.Spec.Approval.Mode, AutoCriteria: doc.Spec.Approval.AutoCriteria,
		MaxRiskLevel: doc.Spec.Approval.MaxRiskLevel, DiffEnabled: true,
		DiffAgent: "ops_core_inner_agent", DiffProvider: "openai-default",
	}
	if view.Approval.MaxRiskLevel == "" && view.Approval.Mode == "agent" {
		view.Approval.MaxRiskLevel = "low" // scheme approvalGate 缺省：low 超险回退人审
	}
	if d := doc.Spec.ApprovalDiff; d.Enabled != nil {
		view.Approval.DiffEnabled = *d.Enabled
	}
	if doc.Spec.ApprovalDiff.Agent != "" {
		view.Approval.DiffAgent = doc.Spec.ApprovalDiff.Agent
	}
	if doc.Spec.ApprovalDiff.Provider != "" {
		view.Approval.DiffProvider = doc.Spec.ApprovalDiff.Provider
	}
	nodeToPhase := map[string]string{}
	type entryBinding struct {
		phase  string
		events []string
	}
	var entryBindings []entryBinding
	for i := range doc.Spec.Phases {
		nv, ok := parsePhaseNode(doc.Spec.Phases[i])
		if !ok {
			continue
		}
		view.Phases = append(view.Phases, nv.Phase)
		if nv.Name != "" {
			nodeToPhase[nv.Name] = nv.Phase
		}
		view.Nodes = append(view.Nodes, nv)
		if len(nv.EntryEvents) > 0 {
			entryBindings = append(entryBindings, entryBinding{phase: nv.Phase, events: nv.EntryEvents})
		}
	}
	for _, ep := range doc.Spec.EntryPoints {
		if ep.Enabled != nil && !*ep.Enabled {
			continue
		}
		target := ep.Phase
		if target == "" && ep.Node != "" {
			target = nodeToPhase[ep.Node]
		}
		if target != "" && len(ep.Events) > 0 {
			entryBindings = append(entryBindings, entryBinding{phase: target, events: ep.Events})
		}
	}
	if rows := doc.Spec.Transitions; len(rows) > 0 {
		view.Transitions = make([]TransitionRow, len(rows))
		for i, r := range rows {
			view.Transitions[i] = TransitionRow{From: r.From, On: r.On, To: r.To}
		}
	}
	// 多节点入口（entryPoints / entryEvents）补充为显式 entry:<Event> 迁移行（若尚未显式声明）。
	if len(view.Phases) > 0 && len(entryBindings) > 0 {
		first := view.Phases[0]
		existing := map[string]bool{}
		for _, tr := range view.Transitions {
			existing[tr.From+"|"+tr.On] = true
		}
		for _, eb := range entryBindings {
			if eb.phase == "" || eb.phase == first {
				continue
			}
			for _, ev := range eb.events {
				on := "entry:" + ev
				if !existing[first+"|"+on] {
					view.Transitions = append(view.Transitions, TransitionRow{From: first, On: on, To: eb.phase})
					existing[first+"|"+on] = true
				}
			}
		}
	}
	// 收尾策略与 postActions 投影（解释器通用收尾动作）
	if doc.Spec.Regression.BugDocPolicy != "" {
		view.BugDocPolicy = doc.Spec.Regression.BugDocPolicy
	} else if doc.Spec.Regression.CleanTempBugDoc != nil {
		if *doc.Spec.Regression.CleanTempBugDoc {
			view.BugDocPolicy = "delete"
		} else {
			view.BugDocPolicy = "keep"
		}
	}
	if len(doc.Spec.Regression.PostActions) > 0 {
		view.PostActions = doc.Spec.Regression.PostActions
	}
	return view
}

// docsBrakesDecl 文档原文的 spec.brakes 声明形态（scheme.yaml Workflow spec brakes 键——
// cooldown 时长词形+maxMutationsPerTick 计数；avoidRepeatFix 随 DriftReport 回归声明面顺延，
// 本段不解析不投影）。形态合法性由写路径 scheme 校验闸门承担，此处只做声明→收紧值的映射。
type docsBrakesDecl struct {
	Spec struct {
		Brakes struct {
			Cooldown            string `yaml:"cooldown"`
			MaxMutationsPerTick int    `yaml:"maxMutationsPerTick"`
		} `yaml:"brakes"`
	} `yaml:"spec"`
}

// ResolveBrakes 实现 BrakesResolver（v0.32 刹车首段——spec.brakes 声明收紧值的解析面；
// 零缓存口径同 ResolveSpec：每次现读现投）。未挂 Lookup/未命中/损坏文档/未声明=ok=false
// （强制默认兜底——不冒充空声明）。宽松方向钳定在 clampBrakes（生效值侧），解析面只透传
// 声明值原样。
func (d DocsSpecs) ResolveBrakes(name string) (Brakes, bool) {
	if d.Lookup == nil || name == "" {
		return Brakes{}, false
	}
	raw, ok := d.Lookup(name)
	if !ok {
		return Brakes{}, false
	}
	var doc docsBrakesDecl
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return Brakes{}, false // 损坏文档不投影（防御性同 ResolveSpec 口径）
	}
	decl := doc.Spec.Brakes
	if decl.Cooldown == "" && decl.MaxMutationsPerTick == 0 {
		return Brakes{}, false // 未声明 brakes=强制默认兜底（不冒充空声明）
	}
	var b Brakes
	if decl.Cooldown != "" {
		dur, err := time.ParseDuration(decl.Cooldown)
		if err != nil {
			return Brakes{}, false // 词形非法=未声明口径（写路径校验闸门是形态权威）
		}
		b.Cooldown = dur
	}
	b.MaxMutationsTick = decl.MaxMutationsPerTick
	return b, true
}
