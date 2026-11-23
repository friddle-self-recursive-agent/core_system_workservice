// objspec.go —— 对象 spec 解析面（v0.27 对象声明表解析首段=表投影；v0.28 相位投影+迁移
// 判定接线，workflows 规则 16 下一段）。
// 规则 14 显式迁移表首段（v0.25）只物化了 builtin 表；平台 Workflow 对象的自声明
//（spec.phases 相位序列+spec.transitions 迁移表，scheme.yaml 声明终态形态）随文档面存在。
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
	"gopkg.in/yaml.v3"
)

// SpecView 对象 spec 投影（v0.28 相位投影段——同一份对象声明的两面一次解析：相位序列
// +声明迁移表）。两个字段同口径：nil/空=未声明=零投影（不冒充空声明）；非空=声明投影
//（顺序保留，原样不解释语义——判定是查表落点的事）。
type SpecView struct {
	Phases      []string        // 声明相位序列投影（spec.phases 原样顺序保留；空=未声明）
	Transitions []TransitionRow // 声明迁移表投影（spec.transitions 原样顺序保留；空=未声明）
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

// docsSpecDecl 文档原文的对象 spec 声明形态（scheme.yaml Workflow spec：
// `phases: []Phase`（相位词汇基座=Agent.status.phase 八相位+workflow 域交付锚 Deploying，
// concepts/workflow.md §1 映射表）+`transitions: ?[]{from: Phase, on: string, to: Phase}`
// ——形态合法性由写路径 scheme 校验闸门承担，此处只做声明→投影的纯映射，不重复校验）。
type docsSpecDecl struct {
	Spec struct {
		Phases      []string `yaml:"phases"`
		Transitions []struct {
			From string `yaml:"from"`
			On   string `yaml:"on"`
			To   string `yaml:"to"`
		} `yaml:"transitions"`
	} `yaml:"spec"`
}

// DocsSpecs 文档面 spec 解析（装配适配）：lookup 由装配层注入（文档库读面按名取 Workflow
// 对象原文），本包只做 spec.phases/spec.transitions 声明 → 投影的映射（引擎不感知 git
// 后端）。零缓存：每次解析现读现投（单次 lookup——相位与表同源同拍）。防御面：lookup
// 未命中/原文不可解析/未声明=零投影（零值 SpecView）。
type DocsSpecs struct {
	Lookup func(name string) (raw []byte, ok bool)
}

// ResolveSpec 实现 SpecResolver：声明→投影纯映射（相位序列与迁移表行 from/on/to 原样，
// 顺序保留）。
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
	if len(doc.Spec.Phases) > 0 {
		view.Phases = doc.Spec.Phases // 声明序列原样（顺序保留——投影面不解释语义）
	}
	if rows := doc.Spec.Transitions; len(rows) > 0 {
		view.Transitions = make([]TransitionRow, len(rows))
		for i, r := range rows {
			view.Transitions[i] = TransitionRow{From: r.From, On: r.On, To: r.To}
		}
	}
	return view
}
