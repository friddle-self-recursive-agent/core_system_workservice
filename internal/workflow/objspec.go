// objspec.go —— 对象 spec 声明表解析首段（v0.27——object 源 run 表投影，workflows 规则 16）。
// 规则 14 显式迁移表首段（v0.25）只物化了 builtin 表；平台 Workflow 对象的自声明表
// （spec.transitions，scheme.yaml 声明终态形态先行）随文档面存在，run 侧不投影=空数组。
// 本段把「对象自声明表解析」的首段落为**投影**：object 源 run 的 transitions 从恒空数组
// 变为对象 spec 声明表行投影（声明→留痕面，顺序保留）。投影≠判定：相位仍不投影（phases
// 恒空）→ 受理首步/审批信号/异步推进三条判定路径的既有「相位不投影=零动作」防御面原样
// 生效，object 源行为面零翻转；相位投影与迁移判定接线=下一段 declared-deferred。
//
// 解耦口径：引擎不感知文档库——解析实现由装配层注入（RunOptions.Specs），本包只定义
// 解析面契约与声明→投影的纯映射；未注入（nil）或未命中=零投影（既有行为面不变）。
package workflow

import (
	"gopkg.in/yaml.v3"
)

// SpecResolver 对象 spec 解析面（对象声明表解析首段 v0.27）：按 workflow 锚名解析平台
// Workflow 对象的声明迁移表。可选能力（挂起-恢复 ResumeSink 同款可选能力范式）：装配层
// 未注入=能力缺席，object 源表不投影（空数组——缺口可见口径保持）。
type SpecResolver interface {
	// ResolveSpec 返回声明表投影行（顺序保留）；nil/空=未命中或未声明（零投影，不冒充
	// 空表声明）。
	ResolveSpec(name string) []TransitionRow
}

// SpecFunc 函数适配器（装配层零样板——Lookup 闭包直接成解析面）。
type SpecFunc func(name string) []TransitionRow

// ResolveSpec 实现 SpecResolver。
func (f SpecFunc) ResolveSpec(name string) []TransitionRow { return f(name) }

// docsTransitions 文档原文的 spec.transitions 声明形态（scheme.yaml Workflow spec
// `transitions: ?[]{from: Phase, on: string, to: Phase}`——形态合法性由写路径 scheme
// 校验闸门承担，此处只做声明→投影的纯映射，不重复校验）。
type docsTransitions struct {
	Spec struct {
		Transitions []struct {
			From string `yaml:"from"`
			On   string `yaml:"on"`
			To   string `yaml:"to"`
		} `yaml:"transitions"`
	} `yaml:"spec"`
}

// DocsSpecs 文档面 spec 解析（装配适配）：lookup 由装配层注入（文档库读面按名取 Workflow
// 对象原文），本包只做 spec.transitions 声明 → 投影行的映射（引擎不感知 git 后端）。
// 零缓存：每次解析现读现投——文档库 git 后端权威，apply 落库即生效（热装载语义天然跟随，
// 无失效窗口）。防御面：lookup 未命中/原文不可解析/无声明=零投影（nil）。
type DocsSpecs struct {
	Lookup func(name string) (raw []byte, ok bool)
}

// ResolveSpec 实现 SpecResolver：声明表行投影（from/on/to 原样，顺序保留）。
func (d DocsSpecs) ResolveSpec(name string) []TransitionRow {
	if d.Lookup == nil || name == "" {
		return nil
	}
	raw, ok := d.Lookup(name)
	if !ok {
		return nil
	}
	var doc docsTransitions
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return nil // 损坏文档不投影（防御性——scheme 校验是写路径权威闸门）
	}
	rows := doc.Spec.Transitions
	if len(rows) == 0 {
		return nil
	}
	out := make([]TransitionRow, len(rows))
	for i, r := range rows {
		out[i] = TransitionRow{From: r.From, On: r.On, To: r.To}
	}
	return out
}
