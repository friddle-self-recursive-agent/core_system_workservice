// relevance.go —— 全局相关性基元 Relevance(A, B) ∈ [0,100]（BUG-002 打分标度统一）。
//
// 设计依据：core_design/concepts/context.md §1.2——全系统一切分值统一 0–100 标度；
// Relevance(A,B) = clamp( w_tag·S_tag + w_aff·S_aff + w_kw·S_kw + w_scope·S_scope )，
// 缺省权重 0.45/0.20/0.25/0.10（Σ=1）。四因子全部为纯函数（可单测，零外部状态）：
//   - S_tag：tag 集合 Tversky 覆盖（对称化 α=β=0.5，即 Dice 系数）——对偶集大小不膨胀误判；
//   - S_aff：同 tag_type 内 tag 词形亲和（段 Dice；同 tag 自亲和 100）——affinity 矩阵后续段
//     接入时替换词形推导，公式形状不变；
//   - S_kw：标题/摘要关键词重合（小写切分 + 停用词表 + CJK 二元组；不引 jieba——零新依赖红线）；
//   - S_scope：scope 路径前缀同域度（同 scope=100 / global=70 / 前缀同域=85 / 同首段=50 / 无关=0）。
//
// 性能红线：全部 O(|A|+|B|) 集合运算，索引字段（tags/keywords/scope）已预切好，find 不扫正文。
package contextidx

import (
	"math"
	"sort"
	"strings"
	"unicode"
)

// 全局相关性权重（context.md §1.2 缺省值；ContextPolicy.relevanceWeights 后续段可配）。
const (
	WeightTag   = 0.45 // 非对称加权 Tag 覆盖（本实现对称化 Tversky/Dice）
	WeightAff   = 0.20 // Tag 亲和传播
	WeightKw    = 0.25 // 关键词重合
	WeightScope = 0.10 // 作用域先验
)

// Breakdown 相关性四因子分项（一律 0–100，与总分同标度——B2 判据「分值全在 0–100」）。
type Breakdown struct {
	Tag   float64 `json:"tag"`
	Aff   float64 `json:"aff"`
	Kw    float64 `json:"kw"`
	Scope float64 `json:"scope"`
}

// Relevance 任意两实体（文档/查询虚对象/分片虚对象）的全局相关性 ∈ [0,100]。
// 纯函数：读入双方预切好的特征（Tags/Keywords/Scope），零锁零 IO。
func Relevance(a, b *Doc) (float64, Breakdown) {
	if a == nil || b == nil {
		return 0, Breakdown{}
	}
	bd := Breakdown{
		Tag:   scoreTag(a, b),
		Aff:   scoreAff(a, b),
		Kw:    scoreKw(a, b),
		Scope: scoreScope(a.Scope, b.Scope),
	}
	total := WeightTag*bd.Tag + WeightAff*bd.Aff + WeightKw*bd.Kw + WeightScope*bd.Scope
	if total < 0 {
		total = 0
	}
	if total > 100 {
		total = 100
	}
	return total, bd
}

// Round1 统一一位小数词形（score 一律 0–100 整数或一位小数——B2 判据词形）。
func Round1(v float64) float64 {
	return math.Round(v*10) / 10
}

// scoreTag tag 集合 Tversky 覆盖（对称化 α=β=0.5 → Dice）：|A∩B| / (|A∩B| + 0.5|A\B| + 0.5|B\A|)。
// 相同集合=100；不相交=0；对偶侧集合膨胀按半权计（非对称覆盖的对称均值——context.md §1.2①
// 「离散 Cosine 分母膨胀漏判」的替代，对称化后任意 A,B 顺序恒等）。
func scoreTag(a, b *Doc) float64 {
	inter := 0.0
	onlyA, onlyB := 0.0, 0.0
	for _, ta := range a.Tags {
		if containsTag(b.Tags, ta) {
			inter++
		} else {
			onlyA++
		}
	}
	for _, tb := range b.Tags {
		if !containsTag(a.Tags, tb) {
			onlyB++
		}
	}
	denom := inter + 0.5*onlyA + 0.5*onlyB
	if denom <= 0 {
		return 0 // 双方零 tag=tag 维不可判（不作满分——诚实零值）
	}
	return inter / denom * 100
}

// containsTag tag 对匹配（tag 名+tagType 全等；tagType 空=通配任一类型——find 查询词形）。
func containsTag(tags []Tag, t Tag) bool {
	for _, x := range tags {
		if x.Tag != t.Tag {
			continue
		}
		if t.TagType == "" || x.TagType == "" || x.TagType == t.TagType {
			return true
		}
	}
	return false
}

// scoreAff 同 tag_type 亲和传播分：对双方共享的每个 tagType，取该类型内 tag 对亲和的最大值，
// 再对共享类型取均值。tag 对亲和：同 tag=100（自亲和）；异 tag=连字符段 Dice（词形邻近——
// 如 console-ui 与 console 同段 console → 66.7；affinity[t_a][t_b] 矩阵后续段接入即替换）。
// 无共享 tagType=0（跨类型零亲和——不冒充）。
func scoreAff(a, b *Doc) float64 {
	byType := func(tags []Tag) map[string][]string {
		m := map[string][]string{}
		for _, t := range tags {
			m[t.TagType] = append(m[t.TagType], t.Tag)
		}
		return m
	}
	ta, tb := byType(a.Tags), byType(b.Tags)
	types := make([]string, 0, len(ta))
	for typ := range ta {
		if _, ok := tb[typ]; ok {
			types = append(types, typ)
		}
	}
	if len(types) == 0 {
		return 0
	}
	sort.Strings(types) // 确定性
	sum := 0.0
	for _, typ := range types {
		best := 0.0
		for _, x := range ta[typ] {
			for _, y := range tb[typ] {
				if v := tagAffinity(x, y); v > best {
					best = v
				}
			}
		}
		sum += best
	}
	return sum / float64(len(types))
}

// tagAffinity 单对 tag 亲和 ∈ [0,100]：全等=100；否则连字符/下划线/点段 Dice。
func tagAffinity(x, y string) float64 {
	if x == y {
		return 100
	}
	sx, sy := splitTagSegments(x), splitTagSegments(y)
	inter := 0
	for _, s := range sx {
		if strings.EqualFold(s, "") {
			continue
		}
		for _, t := range sy {
			if strings.EqualFold(s, t) {
				inter++
				break
			}
		}
	}
	den := len(sx) + len(sy)
	if den == 0 {
		return 0
	}
	return float64(2*inter) / float64(den) * 100
}

// splitTagSegments tag 名切段（-/_/. 分隔，小写）。
func splitTagSegments(tag string) []string {
	return strings.FieldsFunc(strings.ToLower(tag), func(r rune) bool {
		return r == '-' || r == '_' || r == '.'
	})
}

// scoreKw 关键词重合分：预切关键词集合的 Dice（2|A∩B|/(|A|+|B|)）×100。
// 切词在索引构建期完成（keywords 字段），本函数零重复切分——find 性能红线。
func scoreKw(a, b *Doc) float64 {
	if len(a.Keywords) == 0 || len(b.Keywords) == 0 {
		return 0
	}
	set := make(map[string]struct{}, len(b.Keywords))
	for _, k := range b.Keywords {
		set[k] = struct{}{}
	}
	inter := 0
	seen := map[string]struct{}{}
	for _, k := range a.Keywords {
		if _, dup := seen[k]; dup {
			continue
		}
		seen[k] = struct{}{}
		if _, ok := set[k]; ok {
			inter++
		}
	}
	den := len(seen) + len(set)
	if den == 0 {
		return 0
	}
	return float64(2*inter) / float64(den) * 100
}

// scoreScope 作用域先验分（context.md §1.2④ 词形）：同 scope=100；任一侧 global=70；
// 路径前缀同域（段边界对齐）=85；同首段=50；无关=0。
func scoreScope(a, b string) float64 {
	a, b = strings.TrimSpace(a), strings.TrimSpace(b)
	if a == b {
		return 100
	}
	if a == "global" || b == "global" {
		return 70
	}
	if isScopePrefix(a, b) || isScopePrefix(b, a) {
		return 85
	}
	sa, sb := strings.Split(a, "/"), strings.Split(b, "/")
	if sa[0] == sb[0] {
		return 50
	}
	return 0
}

// isScopePrefix a 是否 b 的段边界前缀（"core" ⊑ "core/mvp"；"co" ⊑ "core" 不成立）。
func isScopePrefix(a, b string) bool {
	if a == "" || b == "" {
		return false
	}
	if !strings.HasPrefix(b, a) {
		return false
	}
	return len(b) == len(a) || b[len(a)] == '/'
}

// stopwords 停用词表（小写；英文功能词 + 常见中文单字虚词——CJK 已切二元组，单字虚词兜底）。
var stopwords = map[string]struct{}{}

func init() {
	en := []string{
		"a", "an", "the", "and", "or", "of", "to", "in", "on", "for", "with", "by",
		"is", "are", "be", "as", "at", "it", "its", "this", "that", "from", "not",
	}
	zh := []string{"的", "了", "在", "是", "与", "和", "或", "及", "对", "为", "之", "等", "中"}
	for _, w := range append(en, zh...) {
		stopwords[w] = struct{}{}
	}
}

// Tokenize 关键词切分（索引构建期一次；零外部依赖——手写切分）：ASCII 词小写化；
// 连续 CJK 段切二元组（单字保留）；停用词剔除。标题/摘要/name 共用本入口。
func Tokenize(text string) []string {
	var out []string
	var word strings.Builder
	var cjk strings.Builder
	flushWord := func() {
		w := strings.ToLower(word.String())
		word.Reset()
		if w != "" {
			if _, stop := stopwords[w]; !stop {
				out = append(out, w)
			}
		}
	}
	flushCJK := func() {
		runes := []rune(cjk.String())
		cjk.Reset()
		for i, r := range runes {
			if _, stop := stopwords[string(r)]; stop {
				continue
			}
			if i+1 < len(runes) {
				if _, stop := stopwords[string(runes[i+1])]; !stop {
					out = append(out, string(r)+string(runes[i+1]))
				}
			} else {
				out = append(out, string(r)) // 单字 CJK 尾字保留
			}
		}
	}
	for _, r := range text {
		switch {
		case unicode.Is(unicode.Han, r):
			flushWord()
			cjk.WriteRune(r)
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			flushCJK()
			word.WriteRune(r)
		default:
			flushWord()
			flushCJK()
		}
	}
	flushWord()
	flushCJK()
	return out
}
