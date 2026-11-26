// diff.go —— 聚合快照 → json-patch 保守 diff 器（v0.66 stream 面纯函数件）。
// 设计依据：RFC 6902（JSON Patch 操作词汇）+ RFC 6901（JSON Pointer 路径转义）。
// 口径（任务书 §2）：标准库无 jsonpatch 生成器，本件自写**只产出 add/replace/remove
// 三种操作**的保守 diff 器——不支持 move/copy/test（聚合快照消费方不需要；保守=宁可整段
// replace 也不做细粒度数组下标运算，数组下标 patch 在并发重算下易错位）。纯函数：零共享
// 状态、零 IO，输入=JSON 可解码值树（map[string]any/[]any/标量），输出=操作数组。
//
// 语义：
//   - 两值 DeepEqual 相等 → 零操作；
//   - 双方均为对象（map[string]any）→ 逐键递归：键缺席于新方=remove、缺席于旧方=add、
//     双方在=递归子值；
//   - 其余（数组/标量/类型不一致）→ 父路径整段 replace（数组永不逐下标 diff——保守语义）；
//   - 路径转义 RFC 6901：~ → ~0、/ → ~1。
//
// 应用安全性：数组整段 replace 与对象逐键增删改产生的操作互不重叠（同一路径至多一次
// 操作），任意顺序逐条应用均收敛到新值——消费方无需保序。
package apiserver

import "reflect"

// PatchOp RFC6902 json-patch 操作（本面词汇=op/path/value 三键；remove 无 value）。
type PatchOp struct {
	Op    string `json:"op"`
	Path  string `json:"path"`
	Value any    `json:"value,omitempty"`
}

// patch 操作词（RFC 6902 子集——保守器三词闭集）。
const (
	patchOpAdd     = "add"
	patchOpReplace = "replace"
	patchOpRemove  = "remove"
)

// DiffJSON 旧值树 → 新值树的保守 patch（零差异=nil；调用方空切片判断用 len）。
func DiffJSON(oldVal, newVal any) []PatchOp {
	var ops []PatchOp
	diffValue("", oldVal, newVal, &ops)
	return ops
}

// diffValue 递归（path=父路径 RFC 6901 全路径；ops 就地追加）。
func diffValue(path string, a, b any, ops *[]PatchOp) {
	if reflect.DeepEqual(a, b) {
		return
	}
	am, aok := a.(map[string]any)
	bm, bok := b.(map[string]any)
	if aok && bok {
		diffObject(path, am, bm, ops)
		return
	}
	// 数组/标量/类型不一致：整段 replace（保守——不做数组下标运算）。
	*ops = append(*ops, PatchOp{Op: patchOpReplace, Path: path, Value: b})
}

// diffObject 对象逐键 diff（键序无关——map 语义；输出按旧键序+新键序稳定排列）。
func diffObject(path string, a, b map[string]any, ops *[]PatchOp) {
	// 旧有键：新方缺席=remove；双方在=递归。
	for k, av := range a {
		kp := joinPointer(path, k)
		bv, ok := b[k]
		if !ok {
			*ops = append(*ops, PatchOp{Op: patchOpRemove, Path: kp})
			continue
		}
		diffValue(kp, av, bv, ops)
	}
	// 新增键：add。
	for k, bv := range b {
		if _, ok := a[k]; !ok {
			*ops = append(*ops, PatchOp{Op: patchOpAdd, Path: joinPointer(path, k), Value: bv})
		}
	}
}

// joinPointer RFC 6901 子路径拼接（~ 与 / 转义）。
func joinPointer(parent, key string) string {
	esc := escapePointerToken(key)
	if parent == "" {
		return "/" + esc
	}
	return parent + "/" + esc
}

// escapePointerToken 单 token 转义（RFC 6901：先 ~ 后 /）。
func escapePointerToken(key string) string {
	out := make([]byte, 0, len(key))
	for i := 0; i < len(key); i++ {
		switch key[i] {
		case '~':
			out = append(out, '~', '0')
		case '/':
			out = append(out, '~', '1')
		default:
			out = append(out, key[i])
		}
	}
	return string(out)
}
