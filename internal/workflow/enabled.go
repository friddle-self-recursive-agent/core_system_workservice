// enabled.go —— 规则开关判定面（v0.58b 触发路径 enabled honoring，E3 余量段收口——
// console-adoption.md §1.10 L78「触发路径对 enabled 的实时 honoring 余量（触发接线检查
// spec.enabled=下一段）」+ workflow.md §1「关闭即该规则不再触发、不再消耗」的机读落点）。
//
// 语义锚（workflow.md §1，零新语义）：enabled=false 的 workflow 锚=该规则不再触发——
// 触发留痕仍在（hub 订阅与留痕面零改动），run 不落（disabled 计数可见）；已在途 run 的
// 迁移判定不受开关作用（已完成轮次与回归记录保留=收尾同族语义；审批信号 ResumeSignal 与
// 推进器 Advance 零改动）。判定面缺席（未注入/未命中/解析失败）=放行既有行为面零翻转
// （Specs/Brakes 同款可选能力范式——缺省不改既有行为面；写路径 scheme 校验是形态权威，
// 读面不放大解析形变为整面拦阻；enabled 未声明=缺省 true 启用——scheme.yaml `(=true)`）。
//
// 生效对象解析（与 /v1/workflows 聚合面 effectiveObject 同一 precedence 口径——判定面
// 与聚合面无双源漂移）：系统层 override 优先（多系统并存=字典序 scope 首个——确定性投影）
// → global 本体文档 → 零对象=缺省 true。override 语义（workflow.md §1）：系统层同名覆盖
// 生效层定义；builtin 锚零对象=注册表缺省启用。
//
// 解耦口径：引擎不感知文档库——判定面由装配层注入（EnabledResolver），本包只做
// 对象行集→生效值的纯映射（EnabledOf）与受理路径闸判（StartRun）；零缓存=apply 落库
// 即生效（Specs 同款热装载语义天然跟随，无失效窗口）。
package workflow

import "gopkg.in/yaml.v3"

// EnabledObject 规则开关判定输入行（装配层从读面行集投影——宽松键，零宿主依赖；
// 不引 repository 包类型——依赖方向恒装配层→引擎）。
type EnabledObject struct {
	Scope   string // 对象 scope（global=本体文档；systems/<s> 等系统层=override）
	Content []byte // yaml 原文
}

// EnabledResolver 规则开关判定面（v0.58b——可选能力，Specs/Brakes 同款范式）。
// 未注入（nil）=能力缺席=零闸既有行为面零翻转。
type EnabledResolver interface {
	// Enabled 返回 workflow 锚名的生效 enabled 值。ok=false=判定缺席（放行——解析面
	// 未命中不拦阻：判定面不放大解析形变为整面闸）。
	Enabled(name string) (enabled bool, ok bool)
}

// EnabledFunc 函数适配器（装配层零样板——闭包直接成判定面）。
type EnabledFunc func(name string) (bool, bool)

// Enabled 实现 EnabledResolver。
func (f EnabledFunc) Enabled(name string) (bool, bool) { return f(name) }

// docsEnabledDecl 文档原文的 spec.enabled 声明形态（scheme.yaml Workflow spec
// `enabled: "?bool (=true)"`——指针承载三态：nil=未声明=缺省启用；形态合法性由写路径
// scheme 校验闸门承担，此处只做声明→生效值的纯映射）。
type docsEnabledDecl struct {
	Spec struct {
		Enabled *bool `yaml:"enabled"`
	} `yaml:"spec"`
}

// EnabledOf 对象行集 → 生效 enabled 值（纯映射——与聚合面 effectiveObject 同 precedence）。
// 返回 (生效值, 是否有判定)：任一对象行命中即有判定（缺省词形随行）；零对象行=缺省启用
// （builtin 注册表缺省/对象引用未落库——缺省=true，缺省不改既有行为面）。
// precedence：系统层 override 优先（字典序 scope 首个）→ global 本体文档 → 缺省 true。
func EnabledOf(objs []EnabledObject) (bool, bool) {
	var sys *EnabledObject
	var global *EnabledObject
	for i := range objs {
		o := &objs[i]
		if o.Scope == "global" {
			if global == nil {
				global = o
			}
			continue
		}
		if o.Scope == "" {
			continue // 防御性：无 scope 行不参与生效解析（不冒充 override）
		}
		if sys == nil || o.Scope < sys.Scope {
			sys = o // 多系统并存=字典序首个（确定性投影——聚合面同口径）
		}
	}
	if sys != nil {
		return parseEnabled(sys.Content), true
	}
	if global != nil {
		return parseEnabled(global.Content), true
	}
	return true, false // 零对象=缺省启用（缺省不改既有行为面）
}

// parseEnabled 单对象原文 → spec.enabled 生效值（未声明/不可解析=nil→true——缺省启用；
// 声明 false=停用）。
func parseEnabled(raw []byte) bool {
	var doc docsEnabledDecl
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return true // 损坏文档不拦阻（防御性同 ResolveSpec 口径——写路径校验是权威闸门）
	}
	if doc.Spec.Enabled == nil {
		return true // 未声明=缺省启用（scheme.yaml `(=true)` 声明缺省同源）
	}
	return *doc.Spec.Enabled
}
