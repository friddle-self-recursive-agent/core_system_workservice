// trace.go —— ServiceDesc 设计↔实现文档 1:N [metadata 关联] 追踪（R-DOC-9 / BUG-011）。
//
// 双端结构：N 端（ImplDoc/VerifyDoc）frontmatter metadata.trace 声明 implements/verifies
// （+designRev=对齐时的上游 rv）；1 端（设计文档）零声明零写路径——implemented_by/verified_by
// 由本包扫描索引反向聚合派生（「DocIndex 自动反向聚合」读侧实现）。
//
// stale=Suspect Link 派生值（零写路径）：ImplDoc.designRev < 上游设计文档现 rv → stale
// （B11「设计更新自动标下游 stale」——设计 rv 推进即读面派生过期，不落任何状态）。
// designRev 缺席/非数值=无法判定 → 不标 stale（诚实：判不了不冒充）。
package contextidx

import (
	"sort"
	"strconv"
	"strings"
)

// TraceEntry trace 聚合行（implements/implemented_by/verified_by 共用词形；stale/resolved
// 键恒在=代际判别器——false 显式投影不冒充缺席）。
type TraceEntry struct {
	Doc       string   `json:"doc"`                 // 对端文档 kind/scope/name（可解析时）或声明原文
	Sections  []string `json:"sections,omitempty"`  // 章节/条款锚点透传
	DesignRev string   `json:"designRev,omitempty"` // N 端声明的对齐 rv（verifies 侧=targetRev）
	Stale     bool     `json:"stale"`               // 派生：对齐 rv < 上游现 rv
	Resolved  bool     `json:"resolved"`            // 对端是否解析到索引在册对象（未解析=悬空声明如实标注）
}

// TraceReport 单文档双向 trace 聚合（B11 判据词形）。
type TraceReport struct {
	Kind          string       `json:"kind"`
	Scope         string       `json:"scope"`
	Name          string       `json:"name"`
	Implements    []TraceEntry `json:"implements"`     // 我（N 端）声明实现了谁
	ImplementedBy []TraceEntry `json:"implemented_by"` // 谁（N 端）声明实现了我（1 端反向聚合——1 对 N）
	VerifiedBy    []TraceEntry `json:"verified_by"`    // 谁声明验证了我
}

// Trace 单文档双向聚合。文档无 trace 段=空聚合（不 404）；对象不在索引=ErrAbs 由调用面
// 映射 404（对象本身缺席才 404——route 侧判定）。
func (ix *Index) Trace(kind, scope, name string) (TraceReport, bool) {
	d, ok := ix.Get(kind, scope, name)
	if !ok {
		return TraceReport{}, false
	}
	rep := TraceReport{Kind: kind, Scope: scope, Name: name}
	// 正向：我声明的 implements/verifies —— 逐条解析对端、派生 stale。
	for _, ref := range d.Trace.Implements {
		rep.Implements = append(rep.Implements, ix.resolveEntry(ref))
	}
	for _, ref := range d.Trace.Verifies {
		rep.VerifiedBy = append(rep.VerifiedBy, ix.resolveEntry(ref)) // verifies 词形与 implements 同构（对端=上游）
	}
	// 反向：全索引扫描——谁的 implements/verifies 指向我（1 对 N 聚合；O(N) 键比对，零正文扫描）。
	keys := []string{d.Key()}
	if alt := nameAliases(d); len(alt) > 0 {
		keys = append(keys, alt...)
	}
	ix.mu.RLock()
	others := make([]*Doc, 0, len(ix.docs))
	for _, o := range ix.docs {
		if o.Key() == d.Key() {
			continue
		}
		others = append(others, o)
	}
	ix.mu.RUnlock()
	sort.Slice(others, func(i, j int) bool { return others[i].Key() < others[j].Key() })
	for _, o := range others {
		for _, ref := range o.Trace.Implements {
			if !refMatches(ref.Doc, keys) {
				continue
			}
			rep.ImplementedBy = append(rep.ImplementedBy, entryOf(o, ref, ix.rvOf(ref.Doc)))
		}
		for _, ref := range o.Trace.Verifies {
			if !refMatches(ref.Doc, keys) {
				continue
			}
			rep.VerifiedBy = append(rep.VerifiedBy, entryOf(o, ref, ix.rvOf(ref.Doc)))
		}
	}
	return rep, true
}

// deriveStale 投影面 stale 派生：本对象 implements/verifies 任一条已过期 → true。
// 第二返回值=「有 trace 声明可派生」（无声明=零投影不冒充）。
func (ix *Index) deriveStale(d *Doc) (bool, bool) {
	refs := append(append([]TraceRef{}, d.Trace.Implements...), d.Trace.Verifies...)
	if len(refs) == 0 {
		return false, false
	}
	for _, ref := range refs {
		if ix.staleOf(ref, ix.rvOf(ref.Doc)) {
			return true, true
		}
	}
	return false, true
}

// resolveEntry 正向声明条目解析（对端解析+stale 派生）。
func (ix *Index) resolveEntry(ref TraceRef) TraceEntry {
	e := TraceEntry{Doc: ref.Doc, Sections: ref.Sections, DesignRev: ref.Rev}
	if target, ok := ix.lookupRef(ref.Doc); ok {
		e.Resolved = true
		e.Stale = ix.staleOf(ref, target.RV)
	}
	return e
}

// entryOf 反向聚合条目（对端=声明方 o；上游 rv 已解析）。
func entryOf(o *Doc, ref TraceRef, upRV int64) TraceEntry {
	return TraceEntry{
		Doc:       o.Key(),
		Sections:  ref.Sections,
		DesignRev: ref.Rev,
		Stale:     staleRev(ref.Rev, upRV),
		Resolved:  true,
	}
}

// staleOf stale 派生核心：designRev（数值）< 上游现 rv → true。
func (ix *Index) staleOf(ref TraceRef, upRV int64) bool {
	return staleRev(ref.Rev, upRV)
}

// staleRev 数值比对（designRev/targetRev 可解析为数值且上游 rv>0 才判；缺席/非数值/上游
// rv 不可得=判不了 → false 不冒充）。
func staleRev(rev string, upRV int64) bool {
	if rev == "" || upRV <= 0 {
		return false
	}
	n, err := strconv.ParseInt(strings.TrimSpace(rev), 10, 64)
	if err != nil {
		return false // sha 词形/非数值=机械判不了（诚实降级，不猜）
	}
	return n < upRV
}

// rvOf 按声明原文解析上游现 rv（可解析到索引对象才返回 >0）。
func (ix *Index) rvOf(refDoc string) int64 {
	if d, ok := ix.lookupRef(refDoc); ok {
		return d.RV
	}
	return 0
}

// lookupRef 声明原文 → 索引对象：kind/scope/name 精确键；否则尾段名匹配（路径词形
// "systems/x/desc.md" → 名 desc；索引键序首个命中——确定性）。
func (ix *Index) lookupRef(ref string) (*Doc, bool) {
	if ref == "" {
		return nil, false
	}
	if d, ok := ix.Get(splitKindScopeName(ref)); ok {
		return d, true
	}
	segs := strings.Split(strings.TrimSuffix(ref, ".md"), "/")
	cand := segs[len(segs)-1]
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	keys := make([]string, 0, len(ix.docs))
	for k := range ix.docs {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		d := ix.docs[k]
		if d.Name == cand {
			return d, true
		}
	}
	return nil, false
}

// splitKindScopeName "Kind/Scope/Name" 三段 → Get 入参（段数≠3 返回零值词形=不命中）。
func splitKindScopeName(ref string) (string, string, string) {
	parts := strings.Split(ref, "/")
	if len(parts) < 3 {
		return "\x00", "", "" // 非三段=不可能键（查不到）
	}
	return parts[0], strings.Join(parts[1:len(parts)-1], "/"), parts[len(parts)-1]
}

// refMatches 声明原文命中集合（精确键或尾段名别名）。
func refMatches(ref string, keys []string) bool {
	if ref == "" {
		return false
	}
	for _, k := range keys {
		if ref == k {
			return true
		}
	}
	segs := strings.Split(strings.TrimSuffix(ref, ".md"), "/")
	cand := segs[len(segs)-1]
	for _, k := range keys {
		parts := strings.Split(k, "/")
		if parts[len(parts)-1] == cand {
			return true
		}
	}
	return false
}

// nameAliases 对象的别名键（同名跨 scope 的悬空声明兜底——如按裸 name 声明的旧词形）。
func nameAliases(d *Doc) []string {
	return []string{d.Name}
}

// RelevanceReport /v1/context/relevance 调试面载荷（B2 判据落点）。
type RelevanceReport struct {
	A         string    `json:"a"`
	B         string    `json:"b"`
	Score     float64   `json:"score"`
	Breakdown Breakdown `json:"breakdown"`
}

// RelevanceByKey 两在册对象按键求全局相关性（a/b=kind/scope/name）。
func (ix *Index) RelevanceByKey(a, b string) (RelevanceReport, error) {
	ka, kb, na := splitKindScopeName(a)
	da, ok := ix.Get(ka, kb, na)
	if !ok {
		return RelevanceReport{}, errAbsent{a}
	}
	ka, kb, na = splitKindScopeName(b)
	db, ok := ix.Get(ka, kb, na)
	if !ok {
		return RelevanceReport{}, errAbsent{b}
	}
	score, bd := Relevance(da, db)
	return RelevanceReport{A: da.Key(), B: db.Key(), Score: Round1(score), Breakdown: bd}, nil
}

// errAbsent 对象不在索引（route 映射 404）。
type errAbsent struct{ key string }

func (e errAbsent) Error() string { return "对象不在索引: " + e.key }
