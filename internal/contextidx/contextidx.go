// Package contextidx 上下文索引数据面（P0）：读侧派生内存 DocIndex + 全局相关性 +
// 按 Tag 分片缓存（TagCtxShard）+ trace 1:N 双向聚合（BUG-002/005/011）。
//
// 设计依据：core_design/concepts/context.md（0–100 全局相关性、按 Tag 分片、
// rebuildThreshold 缺省 75 只对高相关条目失效重建）、rule/rule_doc.md R-DOC-9
// （frontmatter metadata.trace 双向 1:N 追踪）、acceptance/bugs/bug.md B2/B5/B11。
//
// 红线：纯读侧派生——零 DocType、零落库、零写路径（索引与分片缓存删掉可重建，
// 权威=repository 对象面）；索引构建 O(N) 一次 + 单对象 chunk 增量（BUG-005：
// DocApplied 只重切更新对象自身的索引块，其余条目零重算）；find 只用索引字段
// （tags/keywords/scope），不全库扫正文。
package contextidx

import (
	"context"
	"sort"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
)

// ───────────────────────── 索引条目模型 ─────────────────────────

// Tag 打标对（frontmatter tags：[{tag, tag_type}]——R-DOC-6 侧车同词形）。
type Tag struct {
	Tag     string `json:"tag"`
	TagType string `json:"tag_type"`
}

// Key 对象键词形 kind/scope/name（involvedRef 惯例）。
func (t Tag) Key() string {
	if t.TagType == "" {
		return t.Tag
	}
	return t.TagType + ":" + t.Tag
}

// TraceRef trace 声明引用（R-DOC-9：trace.implements/verifies 数组项）。
type TraceRef struct {
	Doc      string   `yaml:"doc" json:"doc"`                                 // 设计/实现文档 URI 或 kind/scope/name
	Sections []string `yaml:"sections,omitempty" json:"sections,omitempty"`   // 章节/条款锚点
	Rev      string   `yaml:"designRev,omitempty" json:"designRev,omitempty"` // 对齐时的上游 rv/sha（verifies 侧=targetRev 同槽）
}

// Trace frontmatter metadata.trace 段（R-DOC-9 双端结构；缺席=零值=无 trace 声明）。
type Trace struct {
	Implements []TraceRef `yaml:"implements,omitempty" json:"implements,omitempty"`
	Verifies   []TraceRef `yaml:"verifies,omitempty" json:"verifies,omitempty"`
	Requires   []string   `yaml:"requires,omitempty" json:"requires,omitempty"` // 声明必须具备下游覆盖的 ImplDoc/VerifyDoc
}

// Doc 索引条目（对象 frontmatter 的投影——只用索引字段，正文不进索引）。
type Doc struct {
	Kind       string    `json:"kind"`
	Scope      string    `json:"scope"`
	Name       string    `json:"name"`
	RV         int64     `json:"resourceVersion,omitempty"`
	Title      string    `json:"title,omitempty"`
	Importance int       `json:"importance,omitempty"`
	Tags       []Tag     `json:"tags,omitempty"`
	Keywords   []string  `json:"-"` // 标题/摘要关键词（切好的索引字段；S_kw 用，不外投）
	Digest     string    `json:"digest,omitempty"`
	Trace      Trace     `json:"-"`
	AppliedAt  time.Time `json:"-"` // 最近一次增量落索引时刻（观测用）
}

// Key 对象键 kind/scope/name。
func (d *Doc) Key() string { return d.Kind + "/" + d.Scope + "/" + d.Name }

// ───────────────────────── 索引 ─────────────────────────

// ObjectSource 文档对象读取源（apiserver.ObjectSource 同形——本包自定义最小口面，
// 不反向依赖 apiserver；repository.GitRepo / SQLiteRepo 均满足）。
type ObjectSource interface {
	Get(ctx context.Context, kind, scope, name string) (*repository.Object, error)
	List(ctx context.Context, kind, scope string) ([]*repository.Object, error)
}

// Index 内存 DocIndex（并发安全；读=快照投影，写=单对象 chunk 增量）。
type Index struct {
	mu      sync.RWMutex
	docs    map[string]*Doc // key=kind/scope/name
	src     ObjectSource
	kinds   func() []string // kind 枚举（scheme.ListKinds 适配；nil=单 kind 注入测试）
	shards  *ShardCache
	version atomic.Int64 // 索引版本（每次增量推进——分片缓存键词形 cacheKey 用）
}

// New 构建空索引。src=对象读取源；kinds=注册 kind 枚举（可为 nil——增量路径仍可用）。
// 分片失效阈值经 ThresholdFromEnv 接线（G-07/R-CTX-5：CORE_SYSTEM_CTX_REBUILD_THRESHOLD
// env 可调，缺省 75——生产装配 contextidx.New 单点生效，不再硬编码）。
func New(src ObjectSource, kinds func() []string) *Index {
	ix := &Index{docs: map[string]*Doc{}, src: src, kinds: kinds}
	ix.shards = newShardCache(ix, ThresholdFromEnv())
	return ix
}

// Shards 分片缓存面（TagCtxShard——访问时 stale 重建）。
func (ix *Index) Shards() *ShardCache { return ix.shards }

// Version 索引版本（增量计数——投影观察增量是否发生）。
func (ix *Index) Version() int64 { return ix.version.Load() }

// Build 全量构建（启动一次 O(N)：逐注册 kind 跨 scope 枚举 → 逐对象切 chunk）。
// 单对象解析失败跳过不阻断（坏对象在写路径已被 ErrContent 拦，此处防御读面）。
func (ix *Index) Build(ctx context.Context) error {
	if ix.src == nil {
		return nil
	}
	kinds := []string{}
	if ix.kinds != nil {
		kinds = ix.kinds()
	}
	built := map[string]*Doc{}
	for _, kind := range kinds {
		objs, err := ix.src.List(ctx, kind, "")
		if err != nil {
			continue // 单 kind 枚举失败不阻断全量（读面降级，kind 下次增量自愈）
		}
		for _, o := range objs {
			if d := ParseDoc(o); d != nil {
				built[d.Key()] = d
			}
		}
	}
	ix.mu.Lock()
	defer ix.mu.Unlock()
	ix.docs = built
	ix.version.Add(1)
	return nil
}

// UpdateObject 单对象 chunk 增量（BUG-005 核心落点：DocApplied 只重切更新对象自身的
// 索引块——其余条目零重算、零全量重建）。对象已删（content 不可解析或 Get 缺席）=摘除
// 自身 chunk，并对受影响分片标 stale（G-18 删除面失效——高相关分片下次访问重建）。
// 返回是否落索引。
func (ix *Index) UpdateObject(o *repository.Object) bool {
	if o == nil || o.Kind == "" {
		return false
	}
	d := ParseDoc(o) // 已删/不可解析=nil → 摘除路径
	key := (&Doc{Kind: o.Kind, Scope: o.Scope, Name: o.Name}).Key()
	ix.mu.Lock()
	var removed *Doc
	if d == nil {
		old, ok := ix.docs[key]
		if !ok {
			ix.mu.Unlock()
			return false
		}
		removed = old // 摘除前留档：分片失效按删除前词形打分（G-18）
		delete(ix.docs, key)
	} else {
		ix.docs[d.Key()] = d
	}
	ix.version.Add(1)
	ix.mu.Unlock()
	if removed != nil {
		ix.shards.invalidate(removed, "DocRemoved") // 删除面分片失效（阈值同口径）
	}
	return true
}

// Remove 文档删除失效面（G-18/R-CTX-5 删除即失效）：摘除索引 chunk ∧ 对受影响分片标
// stale——删后 find 立即不命中（Find 读 chunk 集，无需等 rebuild）。仓库侧文档删除路径
// 装配钩子的唯一落点（同步调用一次即可；事件跟随路 Start 的 Get 失败词形同走本实现）。
// 返回是否确有删除（键不在册=false 幂等）。
func (ix *Index) Remove(kind, scope, name string) bool {
	return ix.UpdateObject(&repository.Object{Kind: kind, Scope: scope, Name: name})
}

// Get 按键取条目（快照；返回副本防外改）。
func (ix *Index) Get(kind, scope, name string) (*Doc, bool) {
	key := (&Doc{Kind: kind, Scope: scope, Name: name}).Key()
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	d, ok := ix.docs[key]
	if !ok {
		return nil, false
	}
	cp := *d
	return &cp, true
}

// Len 索引条目数。
func (ix *Index) Len() int {
	ix.mu.RLock()
	defer ix.mu.RUnlock()
	return len(ix.docs)
}

// Start DocApplied 订阅热路径（flowbind.Materializer 同款范式——事件→单对象 Get→
// chunk 增量，零全量枚举）。阻塞调用，装配层 go 例程承载。
func (ix *Index) Start(ctx context.Context, bus *eventbus.Bus) {
	if ix == nil || bus == nil {
		return
	}
	ch, cancel := bus.Subscribe("DocApplied")
	defer cancel()
	for {
		select {
		case <-ctx.Done():
			return
		case ev := <-ch:
			kind, scope, name, ok := splitInvolved(ev.Involved)
			if !ok {
				continue
			}
			if ix.src == nil {
				continue
			}
			// 单对象读回 + 自身 chunk 重切（读失败=对象已删——删除词形与坏对象同路：
			// 摘除 chunk + 分片失效，G-18 删后 find 立即不命中）。
			o, err := ix.src.Get(ctx, kind, scope, name)
			if err != nil {
				ix.Remove(kind, scope, name)
				continue
			}
			ix.UpdateObject(o)
			// 分片失效（B5）：只对 Relevance(更新文档, 分片) ≥ rebuildThreshold 的分片标 stale，
			// 下次访问重建——抑制写放大（不逐分片立刻重建）。
			if d, ok := ix.Get(kind, scope, name); ok {
				ix.shards.Invalidate(d)
			}
		}
	}
}

// splitInvolved 三段拆解（kind/scope/name——scope 允许多段，尾段=name）。
func splitInvolved(involved string) (kind, scope, name string, ok bool) {
	parts := strings.Split(involved, "/")
	if len(parts) < 3 || parts[0] == "" || parts[len(parts)-1] == "" {
		return "", "", "", false
	}
	kind = parts[0]
	name = parts[len(parts)-1]
	scope = strings.Join(parts[1:len(parts)-1], "/")
	if scope == "" {
		return "", "", "", false
	}
	return kind, scope, name, true
}

// ───────────────────────── find / 投影 ─────────────────────────

// FindQuery find 输入（query-doc 虚构输入：q=关键词词形、tags=tagType:tag 或 tag）。
type FindQuery struct {
	Q    string
	Tags []string
	K    int
}

// Hit find 单命中行。
type Hit struct {
	Doc       Doc
	Score     float64   `json:"score"`     // 0–100（一位小数）
	Relevance Breakdown `json:"relevance"` // 四因子分项
	Why       string    `json:"why"`       // 确定性归因（tag/关键词/scope 命中词形）
}

// Find TopK 检索：query-doc（q/tags 虚构）× 全索引 Relevance 排序（只用索引字段，
// O(N) 次纯函数打分零正文扫描）。k≤0=缺省 10，>50 收敛 50（防拖库同口径）。
func (ix *Index) Find(q FindQuery) []Hit {
	if q.K <= 0 {
		q.K = 10
	}
	if q.K > 50 {
		q.K = 50
	}
	qd := &Doc{Title: q.Q, Tags: parseTagList(q.Tags), Keywords: Tokenize(q.Q + " " + strings.Join(q.Tags, " "))}
	ix.mu.RLock()
	all := make([]*Doc, 0, len(ix.docs))
	for _, d := range ix.docs {
		all = append(all, d)
	}
	ix.mu.RUnlock()

	hits := make([]Hit, 0, len(all))
	for _, d := range all {
		score, bd := Relevance(qd, d)
		if score <= 0 {
			continue // 零相关不入榜（context.md ② RankScore θ_min 同旨——毫不相关不混入）
		}
		hits = append(hits, Hit{Doc: *d, Score: Round1(score), Relevance: bd, Why: whyOf(qd, d, bd)})
	}
	sort.SliceStable(hits, func(i, j int) bool {
		if hits[i].Score != hits[j].Score {
			return hits[i].Score > hits[j].Score
		}
		return hits[i].Doc.Key() < hits[j].Doc.Key() // 同分按键字典序（确定性投影）
	})
	if len(hits) > q.K {
		hits = hits[:q.K]
	}
	return hits
}

// whyOf 确定性归因（人读 why——命中来源显式，不冒充）。
func whyOf(q, d *Doc, bd Breakdown) string {
	var parts []string
	shared := sharedTags(q, d)
	if len(shared) > 0 {
		keys := make([]string, 0, len(shared))
		for _, t := range shared {
			keys = append(keys, t.Key())
		}
		parts = append(parts, "tag 命中 "+strings.Join(keys, ","))
	}
	if bd.Aff >= 30 {
		parts = append(parts, "同类型 tag 亲和")
	}
	if bd.Kw >= 20 {
		parts = append(parts, "标题/关键词重合")
	}
	if bd.Scope >= 50 {
		parts = append(parts, "scope 同域")
	}
	if len(parts) == 0 {
		return "弱相关（低分入榜）"
	}
	return strings.Join(parts, "；")
}

// sharedTags 双方共享 tag（query 通配类型按名匹配）。
func sharedTags(a, b *Doc) []Tag {
	var out []Tag
	for _, t := range a.Tags {
		if containsTag(b.Tags, t) && !containsTag(out, t) {
			if bt := matchTag(b.Tags, t); bt != (Tag{}) {
				out = append(out, bt)
			} else {
				out = append(out, t)
			}
		}
	}
	return out
}

// matchTag 取 b 侧实际命中的 tag（词形归 b——投影真实 tagType）。
func matchTag(tags []Tag, t Tag) Tag {
	for _, x := range tags {
		if x.Tag == t.Tag {
			return x
		}
	}
	return Tag{}
}

// DocView /v1/context/index 投影行。
type DocView struct {
	Kind       string   `json:"kind"`
	Scope      string   `json:"scope"`
	Name       string   `json:"name"`
	Title      string   `json:"title,omitempty"`
	Importance int      `json:"importance,omitempty"`
	Tags       []string `json:"tags"`
	Stale      bool     `json:"stale,omitempty"` // trace 派生（下游设计 rv 推进→上游实现 stale——R-DOC-9）
}

// Project 索引投影（scope 前缀过滤 + tagType 过滤；scope 空=全量）。
func (ix *Index) Project(scope, tagType string) []DocView {
	ix.mu.RLock()
	all := make([]*Doc, 0, len(ix.docs))
	for _, d := range ix.docs {
		all = append(all, d)
	}
	ix.mu.RUnlock()
	sort.Slice(all, func(i, j int) bool { return all[i].Key() < all[j].Key() })
	out := make([]DocView, 0, len(all))
	for _, d := range all {
		if scope != "" && !scopeMatches(d.Scope, scope) {
			continue
		}
		view := DocView{
			Kind: d.Kind, Scope: d.Scope, Name: d.Name, Title: d.Title, Importance: d.Importance,
			Tags: tagKeys(d.Tags),
		}
		if tagType != "" && !hasTagType(d.Tags, tagType) {
			continue
		}
		// stale 派生（零写路径）：本对象 implements/verifies 指向的上游现 rv 推进过 designRev → stale。
		if st, any := ix.deriveStale(d); any {
			view.Stale = st
		}
		out = append(out, view)
	}
	return out
}

// scopeMatches scope 过滤：前缀同域（段边界；空=全量）。
func scopeMatches(docScope, want string) bool {
	if want == "" {
		return true
	}
	return docScope == want || isScopePrefix(want, docScope)
}

func tagKeys(tags []Tag) []string {
	out := make([]string, 0, len(tags))
	for _, t := range tags {
		out = append(out, t.Key())
	}
	sort.Strings(out)
	return out
}

func hasTagType(tags []Tag, typ string) bool {
	for _, t := range tags {
		if t.TagType == typ {
			return true
		}
	}
	return false
}

// parseTagList "tagType:tag" / "tag" 逗号清单 → Tag 对（tagType 缺省=通配空）。
func parseTagList(raw []string) []Tag {
	out := make([]Tag, 0, len(raw))
	for _, item := range raw {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if typ, tag, found := strings.Cut(item, ":"); found && typ != "" && tag != "" {
			out = append(out, Tag{Tag: tag, TagType: typ})
		} else {
			out = append(out, Tag{Tag: item})
		}
	}
	return out
}
