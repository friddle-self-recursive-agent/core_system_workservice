// contextidx_test.go —— 上下文索引数据面验收（BUG-002/005/011——B2/B5/B11 判据；
// 测试锚 L81-01…L81-05，接续仓内最新锚号系）：
//   - Relevance 公式（同文档=100 / 无关≈0 / 单调性 / 值域 0–100 一位小数——B2）；
//   - 分片阈值失效（≥rebuildThreshold 标 stale / <阈值零扰动 / 访问时重建——B5）；
//   - trace 聚合（1 设计 N 实现反向聚合 / designRev 过期 stale / 无 trace 空——B11）；
//   - find TopK 排序（只用索引字段 / 确定性同分字典序 / k 收敛）。
package contextidx

import (
	"context"
	"fmt"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
)

// mkDoc 测试条目构造。
func mkDoc(kind, scope, name string, tags []Tag, title string, imp int) *Doc {
	d := &Doc{Kind: kind, Scope: scope, Name: name, Title: title, Importance: imp, Tags: tags}
	kw := title + " " + name
	for _, t := range tags {
		kw += " " + t.Tag
	}
	d.Keywords = Tokenize(kw)
	return d
}

// ───────────────── Relevance 公式（B2）─────────────────

func TestRelevanceSameDocIs100(t *testing.T) {
	a := mkDoc("Rule", "core_system", "rule-ctx", []Tag{{Tag: "context", TagType: "domain"}, {Tag: "rule", TagType: "doc"}}, "上下文规则 Context rules", 95)
	score, bd := Relevance(a, a)
	if score != 100 {
		t.Fatalf("同文档相关性应=100，得 %v（breakdown=%+v）", score, bd)
	}
	if bd.Tag != 100 || bd.Aff != 100 || bd.Kw != 100 || bd.Scope != 100 {
		t.Fatalf("同文档四因子应全 100，得 %+v", bd)
	}
}

func TestRelevanceUnrelatedNearZero(t *testing.T) {
	a := mkDoc("Rule", "core_system", "ctx-doc", []Tag{{Tag: "context", TagType: "domain"}}, "上下文装配 context", 80)
	b := mkDoc("Provider", "deploy", "git-remote", []Tag{{Tag: "notify", TagType: "domain"}}, "远端通知邮件 mailer", 40)
	score, bd := Relevance(a, b)
	if score > 5 {
		t.Fatalf("无关对象相关性应≈0（≤5），得 %v（breakdown=%+v）", score, bd)
	}
	if score < 0 || score > 100 {
		t.Fatalf("相关性越界 [0,100]: %v", score)
	}
}

func TestRelevanceMonotonicInSharedTags(t *testing.T) {
	// 查询侧固定双 tag；命中侧共享 tag 数 0/1/2 → 相关性严格递增（单调性判据）。
	q := mkDoc("Rule", "core_system", "ctx-q", []Tag{{Tag: "context", TagType: "domain"}, {Tag: "cache", TagType: "mech"}}, "上下文 context cache", 80)
	b0 := mkDoc("Rule", "core_system", "ctx-n0", []Tag{{Tag: "notify", TagType: "domain"}}, "通知 notify mailer", 80)
	b1 := mkDoc("Rule", "core_system", "ctx-n1", []Tag{{Tag: "context", TagType: "domain"}}, "上下文 context", 80)
	b2 := mkDoc("Rule", "core_system", "ctx-n2", []Tag{{Tag: "context", TagType: "domain"}, {Tag: "cache", TagType: "mech"}}, "上下文 context cache", 80)
	s0, _ := Relevance(q, b0)
	s1, _ := Relevance(q, b1)
	s2, _ := Relevance(q, b2)
	if !(s1 > s0) || !(s2 > s1) {
		t.Fatalf("共享 tag 增多应严格升：s0=%v s1=%v s2=%v", s0, s1, s2)
	}
	if s2 < 90 {
		t.Fatalf("全共享（tag/标题/scope 全同）应近满分，得 %v", s2)
	}
}

func TestRelevanceSymmetricAndRounded(t *testing.T) {
	a := mkDoc("Rule", "core_system", "doc-a", []Tag{{Tag: "context", TagType: "domain"}, {Tag: "rule", TagType: "doc"}}, "context 上下文 rules", 80)
	b := mkDoc("Rule", "core_system", "doc-b", []Tag{{Tag: "context", TagType: "domain"}}, "context 上下文", 60)
	ab, bda := Relevance(a, b)
	ba, bdb := Relevance(b, a)
	if ab != ba {
		t.Fatalf("对称化 Tversky 应交换恒等：Relevance(a,b)=%v Relevance(b,a)=%v", ab, ba)
	}
	for _, bd := range []Breakdown{bda, bdb} {
		for _, v := range []float64{bd.Tag, bd.Aff, bd.Kw, bd.Scope} {
			if v < 0 || v > 100 {
				t.Fatalf("分项越界 [0,100]: %+v", bd)
			}
		}
	}
}

func TestTokenizeCJKAndStopwords(t *testing.T) {
	toks := Tokenize("上下文装配 Context of the")
	joined := fmt.Sprint(toks)
	if !contains(joined, "上下") || !contains(joined, "下文") {
		t.Fatalf("CJK 应切二元组：%v", toks)
	}
	if contains(joined, "the") || contains(joined, "of") {
		t.Fatalf("英文停用词应剔除：%v", toks)
	}
	if contains(joined, "context") == false {
		t.Fatalf("ASCII 词应保留小写：%v", toks)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (s == sub || len(sub) == 0 || index(s, sub) >= 0)
}

func index(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// ───────────────── 索引与增量（BUG-005）─────────────────

// fakeSrc 测试对象源（List/Get 内存映射）。
type fakeSrc struct{ objs []*repository.Object }

func (f *fakeSrc) List(_ context.Context, kind, scope string) ([]*repository.Object, error) {
	var out []*repository.Object
	for _, o := range f.objs {
		if o.Kind != kind {
			continue
		}
		if scope != "" && o.Scope != scope {
			continue
		}
		out = append(out, o)
	}
	return out, nil
}

func (f *fakeSrc) Get(_ context.Context, kind, scope, name string) (*repository.Object, error) {
	for _, o := range f.objs {
		if o.Kind == kind && o.Scope == scope && o.Name == name {
			return o, nil
		}
	}
	return nil, repository.ErrNotFound
}

func mdDoc(kind, scope, name string, fm string) *repository.Object {
	return &repository.Object{Kind: kind, Scope: scope, Name: name, ResourceVersion: 1,
		Content: []byte("---\n" + fm + "\n---\n# " + name + "\n正文\n")}
}

func TestBuildOFullScanOnceAndIncrement(t *testing.T) {
	src := &fakeSrc{objs: []*repository.Object{
		mdDoc("Rule", "core_system", "doc-a", "name: doc-a\ntags: [{tag: context, tag_type: domain}]\nimportance: 80"),
		mdDoc("Rule", "core_system", "doc-b", "name: doc-b\ntags: [{tag: cache, tag_type: mech}]\nimportance: 60"),
	}}
	ix := New(src, func() []string { return []string{"Rule"} })
	if err := ix.Build(context.Background()); err != nil {
		t.Fatalf("全量构建失败: %v", err)
	}
	if ix.Len() != 2 {
		t.Fatalf("索引应 2 条，得 %d", ix.Len())
	}
	v0 := ix.Version()
	// 单对象增量：只动自身 chunk（BUG-005——非全量重建，版本 +1、条目替换）。
	ok := ix.UpdateObject(mdDoc("Rule", "core_system", "doc-a", "name: doc-a\ntags: [{tag: context, tag_type: domain}, {tag: shard, tag_type: mech}]\nimportance: 90"))
	if !ok {
		t.Fatal("增量应落索引")
	}
	if ix.Version() != v0+1 {
		t.Fatalf("增量应推进版本一次：v0=%d now=%d", v0, ix.Version())
	}
	d, ok := ix.Get("Rule", "core_system", "doc-a")
	if !ok || d.Importance != 90 || len(d.Tags) != 2 {
		t.Fatalf("增量词形未生效：%+v ok=%v", d, ok)
	}
	// 删除词形：不可解析=摘除自身 chunk（其余条目不动）。
	gone := ix.UpdateObject(&repository.Object{Kind: "Rule", Scope: "core_system", Name: "doc-b", Content: []byte("")})
	if !gone || ix.Len() != 1 {
		t.Fatalf("坏对象应摘除自身 chunk：gone=%v len=%d", gone, ix.Len())
	}
}

// ───────────────── 分片阈值失效（B5）─────────────────

func TestShardThresholdInvalidation(t *testing.T) {
	src := &fakeSrc{objs: []*repository.Object{
		mdDoc("Rule", "core_system", "cache-doc", "name: cache-doc\ntags: [{tag: cache, tag_type: mech}]\nimportance: 80"),
		mdDoc("Rule", "core_system", "mixed-doc", "name: mixed-doc\ntags: [{tag: cache, tag_type: mech}, {tag: workflow, tag_type: domain}, {tag: notify, tag_type: domain}]\nimportance: 60"),
		mdDoc("Rule", "deploy", "other-doc", "name: other-doc\ntags: [{tag: git, tag_type: domain}]\nimportance: 50"),
	}}
	ix := New(src, func() []string { return []string{"Rule"} })
	_ = ix.Build(context.Background())
	c := ix.Shards()
	if c.Threshold() != DefaultRebuildThreshold {
		t.Fatalf("缺省阈值应 75，得 %d", c.Threshold())
	}
	// 预热：cache 分片建好未 stale。
	sh := c.Get("mech", "cache")
	if sh.Stale || len(sh.DocKeys) == 0 {
		t.Fatalf("分片初建词形异常：%+v", sh)
	}
	// 高相关更新（单 tag 文档，Relevance=100 ≥ 75）→ 标 stale（B5 前半）。
	ix.UpdateObject(mdDoc("Rule", "core_system", "cache-doc", "name: cache-doc\ntags: [{tag: cache, tag_type: mech}]\nimportance: 85"))
	d, _ := ix.Get("Rule", "core_system", "cache-doc")
	c.Invalidate(d)
	sh2 := c.shards["mech:cache"]
	if !sh2.Stale {
		t.Fatalf("高相关分片应标 stale（Relevance≥75）：%+v", sh2)
	}
	if sh2.Reason == "" {
		t.Fatal("stale 应携失效因由（DocApplied 溯源）")
	}
	// stale 分片下次访问重建（失效与重建解耦）。
	sh3 := c.Get("mech", "cache")
	if sh3.Stale {
		t.Fatal("访问后应重建去 stale")
	}
	if len(sh3.DocKeys) == 0 {
		t.Fatal("重建后成员集应为非空")
	}
	// 弱相关更新（三 tag 文档含 cache，对称 Dice<75）→ 零扰动（B5 后半：抑制写放大）。
	ix.UpdateObject(mdDoc("Rule", "core_system", "mixed-doc", "name: mixed-doc\ntags: [{tag: cache, tag_type: mech}, {tag: workflow, tag_type: domain}, {tag: notify, tag_type: domain}]"))
	dm, _ := ix.Get("Rule", "core_system", "mixed-doc")
	sh4 := c.Get("mech", "cache") // 先确保新鲜
	if sh4.Stale {
		t.Fatal("前置应新鲜")
	}
	c.Invalidate(dm)
	if sh := c.shards["mech:cache"]; sh.Stale {
		t.Fatalf("弱相关分片不应被失效（Relevance<75）：%s", sh.Reason)
	}
	// 无关更新 → 零扰动。
	do, _ := ix.Get("Rule", "deploy", "other-doc")
	c.Invalidate(do)
	if sh := c.shards["mech:cache"]; sh.Stale {
		t.Fatal("无关文档不应触发失效")
	}
}

func TestThresholdFromEnv(t *testing.T) {
	t.Setenv("CORE_SYSTEM_CTX_REBUILD_THRESHOLD", "60")
	if v := ThresholdFromEnv(); v != 60 {
		t.Fatalf("env 阈值应生效 60，得 %d", v)
	}
	t.Setenv("CORE_SYSTEM_CTX_REBUILD_THRESHOLD", "bogus")
	if v := ThresholdFromEnv(); v != DefaultRebuildThreshold {
		t.Fatalf("非法 env 应回缺省 75，得 %d", v)
	}
	t.Setenv("CORE_SYSTEM_CTX_REBUILD_THRESHOLD", "")
	if v := ThresholdFromEnv(); v != DefaultRebuildThreshold {
		t.Fatalf("缺省应 75，得 %d", v)
	}
}

// ───────────────── trace 聚合（B11 / R-DOC-9）─────────────────

func designAndImpls() *fakeSrc {
	// 1 设计（rv=3）+ N 实现（两篇 implements 指向设计；一篇 designRev 过期）+ 一篇验证文档。
	return &fakeSrc{objs: []*repository.Object{
		{Kind: "Rule", Scope: "core_system", Name: "design-doc", ResourceVersion: 3, Content: []byte("---\nname: design-doc\nimportance: 90\ntrace:\n  requires: [impl-doc-a]\n---\n# design-doc\n设计正文\n")},
		{Kind: "Rule", Scope: "core_system", Name: "impl-doc-a", ResourceVersion: 1, Content: []byte("---\nname: impl-doc-a\ntags: [{tag: impl, tag_type: lifecycle}]\ntrace:\n  implements:\n    - {doc: Rule/core_system/design-doc, sections: [\"§3\"], designRev: \"3\"}\n---\n# impl-doc-a\n实现正文\n")},
		{Kind: "Rule", Scope: "core_system", Name: "impl-doc-b", ResourceVersion: 1, Content: []byte("---\nname: impl-doc-b\ntrace:\n  implements:\n    - {doc: Rule/core_system/design-doc, designRev: \"2\"}\n---\n# impl-doc-b\n实现正文 B\n")},
		{Kind: "Rule", Scope: "core_system", Name: "verify-doc", ResourceVersion: 1, Content: []byte("---\nname: verify-doc\ntrace:\n  verifies:\n    - {doc: Rule/core_system/design-doc, targetRev: \"3\"}\n---\n# verify-doc\n验证正文\n")},
	}}
}

func TestTraceOneDesignNImplReverseAggregation(t *testing.T) {
	src := designAndImpls()
	ix := New(src, func() []string { return []string{"Rule"} })
	_ = ix.Build(context.Background())
	rep, ok := ix.Trace("Rule", "core_system", "design-doc")
	if !ok {
		t.Fatal("设计文档应在索引")
	}
	// 1 端零声明（除 requires）：implements 空。
	if len(rep.Implements) != 0 {
		t.Fatalf("设计文档 implements 应空，得 %+v", rep.Implements)
	}
	// N 端反向聚合：implemented_by=2（1 对 N）。
	if len(rep.ImplementedBy) != 2 {
		t.Fatalf("implemented_by 应聚合 2 篇实现，得 %+v", rep.ImplementedBy)
	}
	if len(rep.VerifiedBy) != 1 {
		t.Fatalf("verified_by 应聚合 1 篇验证，得 %+v", rep.VerifiedBy)
	}
	// 反向行词形：对端文档键 + sections 透传。
	foundA := false
	for _, e := range rep.ImplementedBy {
		if e.Doc == "Rule/core_system/impl-doc-a" {
			foundA = true
			if len(e.Sections) != 1 || e.Sections[0] != "§3" {
				t.Fatalf("sections 应透传： %+v", e)
			}
		}
	}
	if !foundA {
		t.Fatalf("implemented_by 缺 impl-doc-a：%+v", rep.ImplementedBy)
	}
}

func TestTraceStaleDerivation(t *testing.T) {
	src := designAndImpls()
	ix := New(src, func() []string { return []string{"Rule"} })
	_ = ix.Build(context.Background())
	// 设计 rv=3：impl-a 声明 designRev=3 → synced；impl-b designRev=2 → stale（B11 判据）。
	rep, _ := ix.Trace("Rule", "core_system", "design-doc")
	staleMap := map[string]bool{}
	for _, e := range rep.ImplementedBy {
		staleMap[e.Doc] = e.Stale
	}
	if staleMap["Rule/core_system/impl-doc-a"] {
		t.Fatal("designRev=3（对齐现 rv）不应 stale")
	}
	if !staleMap["Rule/core_system/impl-doc-b"] {
		t.Fatal("designRev=2 < 现 rv=3 应 stale（设计更新自动标下游过期）")
	}
	// 正向：impl-a 查自己的 implements——resolved 且不 stale。
	fwd, ok := ix.Trace("Rule", "core_system", "impl-doc-a")
	if !ok || len(fwd.Implements) != 1 {
		t.Fatalf("实现文档 implements 应 1 条：ok=%v %+v", ok, fwd.Implements)
	}
	if !fwd.Implements[0].Resolved || fwd.Implements[0].Stale {
		t.Fatalf("对齐声明应 resolved 且非 stale： %+v", fwd.Implements[0])
	}
	// 投影面 stale 派生：design-doc（requires 声明族）有可派生 trace；过期实现令其 stale。
	design, _ := ix.Get("Rule", "core_system", "design-doc")
	if _, any := ix.deriveStale(design); any {
		t.Fatal("无 implements/verifies 声明=无派生面（requires 不参与 stale）")
	}
	implb, _ := ix.Get("Rule", "core_system", "impl-doc-b")
	if st, any := ix.deriveStale(implb); !any || !st {
		t.Fatalf("过期实现应派生 stale：st=%v any=%v", st, any)
	}
}

func TestTraceNoTraceEmptyAggregation(t *testing.T) {
	src := &fakeSrc{objs: []*repository.Object{
		mdDoc("Rule", "core_system", "plain-doc", "name: plain-doc\nimportance: 50"),
	}}
	ix := New(src, func() []string { return []string{"Rule"} })
	_ = ix.Build(context.Background())
	rep, ok := ix.Trace("Rule", "core_system", "plain-doc")
	if !ok {
		t.Fatal("在册对象应返回聚合")
	}
	if len(rep.Implements) != 0 || len(rep.ImplementedBy) != 0 || len(rep.VerifiedBy) != 0 {
		t.Fatalf("无 trace 段=空聚合不 404，得 %+v", rep)
	}
	// 对象缺席=false（route 映射 404——与空聚合分立）。
	if _, ok := ix.Trace("Rule", "nowhere", "ghost"); ok {
		t.Fatal("缺席对象不应返回聚合")
	}
}

// ───────────────── find / relevance / 投影 ─────────────────

func TestFindTopKOrdering(t *testing.T) {
	src := &fakeSrc{objs: []*repository.Object{
		mdDoc("Rule", "core_system", "ctx-a", "name: ctx-a\ntags: [{tag: context, tag_type: domain}, {tag: cache, tag_type: mech}]\nimportance: 90"),
		mdDoc("Rule", "core_system", "ctx-b", "name: ctx-b\ntags: [{tag: context, tag_type: domain}]\nimportance: 60"),
		mdDoc("Rule", "deploy", "git-c", "name: git-c\ntags: [{tag: notify, tag_type: domain}]\nimportance: 99"),
	}}
	ix := New(src, func() []string { return []string{"Rule"} })
	_ = ix.Build(context.Background())
	hits := ix.Find(FindQuery{Q: "上下文 context", Tags: []string{"domain:context"}, K: 2})
	if len(hits) > 2 {
		t.Fatalf("k=2 应收敛，得 %d", len(hits))
	}
	if len(hits) != 2 {
		t.Fatalf("tag+关键词双命中应 2 条，得 %+v", hits)
	}
	if hits[0].Score < hits[1].Score {
		t.Fatalf("应按相关性降序：%v < %v", hits[0].Score, hits[1].Score)
	}
	for _, h := range hits {
		if h.Score <= 0 || h.Score > 100 {
			t.Fatalf("score 越界（0–100 一位小数）: %v", h.Score)
		}
		if h.Why == "" {
			t.Fatal("命中应携归因 why")
		}
	}
	// 无关对象不入榜（零相关剔除——context.md θ_min 同旨）。
	for _, h := range hits {
		if h.Doc.Name == "git-c" {
			t.Fatalf("零相关对象不应入榜：%+v", hits)
		}
	}
	// k 缺省 10、上限 50 由路由侧收敛——此处锚 k≤0 缺省不炸。
	if h := ix.Find(FindQuery{Q: "context"}); len(h) == 0 {
		t.Fatal("缺省 k 应可用")
	}
}

func TestRelevanceByKeyRouteShape(t *testing.T) {
	src := designAndImpls()
	ix := New(src, func() []string { return []string{"Rule"} })
	_ = ix.Build(context.Background())
	rep, err := ix.RelevanceByKey("Rule/core_system/design-doc", "Rule/core_system/impl-doc-a")
	if err != nil {
		t.Fatalf("按键相关性失败: %v", err)
	}
	if rep.A != "Rule/core_system/design-doc" || rep.B != "Rule/core_system/impl-doc-a" {
		t.Fatalf("词形应回显 a/b：%+v", rep)
	}
	if rep.Score <= 0 || rep.Score > 100 {
		t.Fatalf("score 越界: %v", rep.Score)
	}
	if _, err := ix.RelevanceByKey("Rule/nowhere/ghost", "Rule/core_system/design-doc"); err == nil {
		t.Fatal("缺席对象应报错（route 映射 404）")
	}
}

func TestProjectScopeAndTagTypeFilter(t *testing.T) {
	src := &fakeSrc{objs: []*repository.Object{
		mdDoc("Rule", "core_system", "doc-a", "name: doc-a\ntags: [{tag: context, tag_type: domain}]\nimportance: 80"),
		mdDoc("Rule", "core_system/mvp", "doc-b", "name: doc-b\ntags: [{tag: cache, tag_type: mech}]\nimportance: 60"),
		mdDoc("Rule", "deploy", "doc-c", "name: doc-c\ntags: [{tag: notify, tag_type: domain}]\nimportance: 40"),
	}}
	ix := New(src, func() []string { return []string{"Rule"} })
	_ = ix.Build(context.Background())
	all := ix.Project("", "")
	if len(all) != 3 {
		t.Fatalf("全量投影应 3 条，得 %d", len(all))
	}
	if v := ix.Project("core_system", ""); len(v) != 2 {
		t.Fatalf("scope 前缀过滤应 2 条，得 %d", len(v))
	}
	if v := ix.Project("", "mech"); len(v) != 1 || v[0].Name != "doc-b" {
		t.Fatalf("tagType 过滤词形异常：%+v", v)
	}
	for _, v := range all {
		if len(v.Tags) == 0 {
			t.Fatalf("投影行应带 tag 键词形：%+v", v)
		}
	}
}

func TestParseDocTraceBothShapes(t *testing.T) {
	// yaml 词形：metadata.trace 嵌套。
	y := &repository.Object{Kind: "Rule", Scope: "s", Name: "y-doc", ResourceVersion: 7, Content: []byte(
		"name: y-doc\nmetadata:\n  name: y-doc\n  trace:\n    implements:\n      - {doc: Rule/s/up, designRev: \"1\"}\ntags: [{tag: mech, tag_type: domain}]\nimportance: 70\n")}
	d := ParseDoc(y)
	if d == nil || d.RV != 7 || len(d.Trace.Implements) != 1 || d.Trace.Implements[0].Rev != "1" {
		t.Fatalf("yaml metadata.trace 词形解析失败：%+v", d)
	}
	if d.Importance != 70 || len(d.Tags) != 1 {
		t.Fatalf("tags/importance 解析失败：%+v", d)
	}
	// md 词形：frontmatter 扁平 trace（导出词形）。
	m := mdDoc("Rule", "s", "m-doc", "name: m-doc\ntrace:\n  verifies:\n    - {doc: Rule/s/y-doc}\ntitle: 手写标题")
	d2 := ParseDoc(m)
	if d2 == nil || len(d2.Trace.Verifies) != 1 || d2.Trace.Verifies[0].Doc != "Rule/s/y-doc" {
		t.Fatalf("md 扁平 trace 词形解析失败：%+v", d2)
	}
	if d2.Title != "手写标题" {
		t.Fatalf("title 应取 frontmatter：%+v", d2)
	}
}
