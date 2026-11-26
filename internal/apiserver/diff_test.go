// diff_test.go —— 保守 diff 器单测（v0.66 stream 面：增/改/删/嵌套 map/数组 replace
// 保守语义/路径转义/根标量/零差异——任务书 §2 判据）。
package apiserver

import (
	"encoding/json"
	"reflect"
	"testing"
)

// applyOps 逐条应用 patch（测试内最小 RFC6902 应用器——只验本面三词闭集语义）。
func applyOps(t *testing.T, doc map[string]any, ops []PatchOp) map[string]any {
	t.Helper()
	for _, op := range ops {
		// 只有一级嵌套以下的路径在用例中出现；深路径用 pointer 手工走两级。
		switch op.Op {
		case patchOpRemove:
			removePointer(t, doc, op.Path)
		case patchOpAdd, patchOpReplace:
			setPointer(t, doc, op.Path, op.Value)
		default:
			t.Fatalf("未知操作词 %q（保守器闭集=add/replace/remove）", op.Op)
		}
	}
	return doc
}

// setPointer 按 RFC6901 路径写值（测试用；只支持两级以内）。
func setPointer(t *testing.T, doc map[string]any, path string, val any) {
	t.Helper()
	tok, rest := splitPointer(path)
	if rest == "" {
		doc[tok] = val
		return
	}
	sub, ok := doc[tok].(map[string]any)
	if !ok {
		sub = map[string]any{}
		doc[tok] = sub
	}
	setPointer(t, sub, rest, val)
}

// removePointer 按 RFC6901 路径删键（测试用；只支持两级以内）。
func removePointer(t *testing.T, doc map[string]any, path string) {
	t.Helper()
	tok, rest := splitPointer(path)
	if rest == "" {
		delete(doc, tok)
		return
	}
	sub, ok := doc[tok].(map[string]any)
	if !ok {
		t.Fatalf("remove 路径 %q 父级缺席", path)
	}
	removePointer(t, sub, rest)
}

// splitPointer 拆出首段与余段（路径以 / 起）。
func splitPointer(path string) (tok, rest string) {
	p := path[1:] // 去 /
	if i := indexByte(p, '/'); i >= 0 {
		return unescapeToken(p[:i]), p[i:]
	}
	return unescapeToken(p), ""
}

func indexByte(s string, b byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == b {
			return i
		}
	}
	return -1
}

func unescapeToken(tok string) string {
	out := make([]byte, 0, len(tok))
	for i := 0; i < len(tok); i++ {
		if tok[i] == '~' && i+1 < len(tok) {
			i++
			switch tok[i] {
			case '0':
				out = append(out, '~')
			case '1':
				out = append(out, '/')
			}
			continue
		}
		out = append(out, tok[i])
	}
	return string(out)
}

func TestDiffJSONNoChange(t *testing.T) {
	a := map[string]any{"x": 1, "list": []any{"a", "b"}, "nested": map[string]any{"k": "v"}}
	if ops := DiffJSON(a, a); len(ops) != 0 {
		t.Fatalf("相等值应零操作，得 %v", ops)
	}
	// JSON 往返等价（两侧同走一次解码——解码域内类型一致 float64；快照构造器两次现建
	// 同形同型，diff 器以此为准）也判等。
	raw, _ := json.Marshal(a)
	var roundA, roundB map[string]any
	if err := json.Unmarshal(raw, &roundA); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &roundB); err != nil {
		t.Fatal(err)
	}
	if ops := DiffJSON(roundA, roundB); len(ops) != 0 {
		t.Fatalf("JSON 往返等价应零操作，得 %v", ops)
	}
}

func TestDiffJSONAddRemoveReplace(t *testing.T) {
	oldDoc := map[string]any{"keep": "1", "gone": "x", "chg": "old"}
	newDoc := map[string]any{"keep": "1", "chg": "new", "born": "y"}
	ops := DiffJSON(oldDoc, newDoc)
	got := applyOps(t, map[string]any{"keep": "1", "gone": "x", "chg": "old"}, ops)
	if !reflect.DeepEqual(got, newDoc) {
		t.Fatalf("应用 patch 后应收敛新值：%v vs %v（ops=%v）", got, newDoc, ops)
	}
	// 词形闭集校验：remove/add/replace 各一。
	kinds := map[string]int{}
	for _, op := range ops {
		kinds[op.Op]++
	}
	if kinds[patchOpRemove] != 1 || kinds[patchOpAdd] != 1 || kinds[patchOpReplace] != 1 {
		t.Fatalf("应 remove/add/replace 各一：%v", ops)
	}
}

func TestDiffJSONNestedMapRecurses(t *testing.T) {
	oldDoc := map[string]any{"obj": map[string]any{"a": 1, "b": 2}}
	newDoc := map[string]any{"obj": map[string]any{"a": 1, "b": 3, "c": 4}}
	ops := DiffJSON(oldDoc, newDoc)
	// 嵌套对象应递归出细粒度操作（/obj/b replace + /obj/c add），而非整段 /obj replace。
	if len(ops) != 2 {
		t.Fatalf("嵌套对象应递归细粒度（2 操作），得 %v", ops)
	}
	got := applyOps(t, oldDoc, ops)
	if !reflect.DeepEqual(got, newDoc) {
		t.Fatalf("应用后应收敛：%v vs %v", got, newDoc)
	}
}

func TestDiffJSONArrayReplacedWholesale(t *testing.T) {
	oldDoc := map[string]any{"list": []any{"a", "b", "c"}}
	newDoc := map[string]any{"list": []any{"a", "x"}}
	ops := DiffJSON(oldDoc, newDoc)
	if len(ops) != 1 || ops[0].Op != patchOpReplace || ops[0].Path != "/list" {
		t.Fatalf("数组应整段 replace（保守语义，不逐下标），得 %v", ops)
	}
}

func TestDiffJSONTypeChangeReplaces(t *testing.T) {
	oldDoc := map[string]any{"v": map[string]any{"x": 1}}
	newDoc := map[string]any{"v": "scalar-now"}
	ops := DiffJSON(oldDoc, newDoc)
	if len(ops) != 1 || ops[0].Op != patchOpReplace || ops[0].Path != "/v" {
		t.Fatalf("类型不一致应父路径整段 replace，得 %v", ops)
	}
}

func TestDiffJSONPointerEscaping(t *testing.T) {
	oldDoc := map[string]any{"a/b": 1, "c~d": 1}
	newDoc := map[string]any{"a/b": 2, "c~d": 2}
	ops := DiffJSON(oldDoc, newDoc)
	paths := map[string]bool{}
	for _, op := range ops {
		paths[op.Path] = true
	}
	if !paths["/a~1b"] || !paths["/c~0d"] {
		t.Fatalf("RFC6901 转义应为 /a~1b 与 /c~0d，得 %v", ops)
	}
	got := applyOps(t, oldDoc, ops)
	if !reflect.DeepEqual(got, newDoc) {
		t.Fatalf("转义路径应用后应收敛：%v vs %v", got, newDoc)
	}
}

func TestDiffJSONRootScalarReplace(t *testing.T) {
	ops := DiffJSON("old", "new")
	if len(ops) != 1 || ops[0].Op != patchOpReplace || ops[0].Path != "" {
		t.Fatalf("根标量应整文档 replace（空路径），得 %v", ops)
	}
}
