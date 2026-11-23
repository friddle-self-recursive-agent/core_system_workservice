// brakes.go —— 刹车首段（v0.32——冷却门与推进拍变更上限，workflow.md 规则 17）。
// 刹车三件套（Brakes，workflow.go）的执行面首段：强制默认+收紧判定+刹车闸两路生效。
// 强制默认=cooldown 2m ∧ maxMutationsPerTick 10（concepts/workflow.md §3 示例值升格为
// builtin 强制默认；avoidRepeatFix 随 DriftReport 回归声明面顺延，本段不落）。
// 收紧判定（可收紧不可关闭的机读落点）：spec.brakes 声明值只允许收紧——cooldown 只增 ∧
// maxMutationsPerTick 只减；宽松方向声明（cooldown 小于默认、变更上限大于默认或非正值）
// 一律钳回强制默认（运行时不即兴不拒绝——钳回=不可关闭的机械落点；声明形态合法性仍由写
// 路径 scheme 校验闸门承担）。
// 刹车闸（装配层可选能力，ResumeSink 同款范式）：未挂接（缺省关断）=零刹车——触发全落
// run、推进不限量，既有行为面零翻转（基线保护）；挂接后两路生效：冷却门（受理路径，同锚
// 同 involved 窗口判——防震荡语义粒度=同一对象，跨 involved 不互踩；cooldown 锚首段=run
// 受理时刻，完成锚随真实执行面接线后切换）与推进拍变更上限（每拍每锚 run 相位变更数受
// 生效刹车值约束，只计实际迁移，超限 run 本拍零动作下拍续推）。
package workflow

import "time"

// DefaultBrakes 刹车三件套强制默认（Workflow 级强制默认：可收紧不可关闭）。
var DefaultBrakes = Brakes{
	Cooldown:         2 * time.Minute, // 成功后最小间隔（防震荡；§3 示例值升格）
	MaxMutationsTick: 10,              // 单轮（推进拍）最大变更数
	AvoidRepeatFix:   true,            // DriftReport 回归面顺延（本段声明占位不落执行面）
}

// BrakeView 生效刹车值投影（声明→留痕面：强制默认钳定后的每锚生效值；run 载荷投影来源）。
type BrakeView struct {
	Cooldown         time.Duration
	MaxMutationsTick int
	Source           string // default=强制默认兜底；spec=对象声明收紧值（经钳定）
}

// 刹车生效值来源词汇（BrakeView.Source）。
const (
	BrakeSourceDefault = "default" // 强制默认（解析未命中/未声明/未注入解析面）
	BrakeSourceSpec    = "spec"    // 对象 spec.brakes 声明收紧值
)

// BrakesResolver 刹车声明解析面（可选能力装配契约）：按锚名解析平台 Workflow 对象的
// spec.brakes 声明值；ok=false=未命中/未声明（强制默认兜底）。DocsSpecs 同源实现
// （spec 投影解析面同款零缓存口径——每次现读现投）。未注入（nil）=全锚强制默认。
type BrakesResolver interface {
	ResolveBrakes(name string) (Brakes, bool)
}

// BrakesFunc 函数适配器（装配层零样板——Lookup 闭包直接成解析面）。
type BrakesFunc func(name string) (Brakes, bool)

// ResolveBrakes 实现 BrakesResolver。
func (f BrakesFunc) ResolveBrakes(name string) (Brakes, bool) { return f(name) }

// BrakeOptions 刹车闸挂接项（RunOptions 成员——装配层可选能力注入点）。缺省 Enabled=false
// =零刹车（既有行为面零翻转——基线保护；关断语义组合声明面先定，workflow.md 规则 17）。
type BrakeOptions struct {
	// Enabled 刹车闸挂接态：false=零刹车（触发全落 run、推进不限量）；true=冷却门+变更上限生效。
	Enabled bool
	// Resolver 刹车声明解析面（nil=全锚强制默认；注入 DocsSpecs 同款解析面=spec.brakes 声明收紧）。
	Resolver BrakesResolver
}

// clampBrakes 收紧判定：宽松方向钳回强制默认（cooldown 只增——小于默认即默认；变更上限
// 只减——大于默认或非正值即默认）。返回生效值；是否携声明值由调用方按解析命中标注来源。
func clampBrakes(b Brakes) BrakeView {
	v := BrakeView{
		Cooldown:         DefaultBrakes.Cooldown,
		MaxMutationsTick: DefaultBrakes.MaxMutationsTick,
		Source:           BrakeSourceDefault,
	}
	if b.Cooldown > v.Cooldown {
		v.Cooldown = b.Cooldown
	}
	if b.MaxMutationsTick > 0 && b.MaxMutationsTick < v.MaxMutationsTick {
		v.MaxMutationsTick = b.MaxMutationsTick
	}
	return v
}

// effectiveBrakes 每锚生效刹车值（表驱动：spec.brakes → 强制默认钳定 → 生效值；解析面
// 未注入/未命中/未声明=强制默认。判定面零即兴——刹车值是声明数据不是运行时状态）。
func effectiveBrakes(r BrakesResolver, name string) BrakeView {
	if r != nil {
		if b, ok := r.ResolveBrakes(name); ok {
			v := clampBrakes(b)
			v.Source = BrakeSourceSpec
			return v
		}
	}
	return BrakeView{
		Cooldown:         DefaultBrakes.Cooldown,
		MaxMutationsTick: DefaultBrakes.MaxMutationsTick,
		Source:           BrakeSourceDefault,
	}
}

// brakeKey 冷却门键（锚+involved——防震荡语义粒度=同一对象；分隔符取不可见控制字符，
// 锚名与 involved 投影不含该字节，键无歧义）。
func brakeKey(wf, involved string) string {
	return wf + "\x1f" + involved
}
