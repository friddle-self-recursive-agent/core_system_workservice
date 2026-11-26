// shard.go —— TagCtxShard 按 Tag 细粒度分片缓存（context.md §3；BUG-005 B5 判据）。
//
// key=tagType:tag（原子触发键），value=该 tag 相关对象集+版本。DocApplied 时只对
// Relevance(更新文档, 分片) ≥ rebuildThreshold（缺省 75，env CORE_SYSTEM_CTX_REBUILD_THRESHOLD
// 可调）的分片标 stale（抑制写放大——弱相关分片零扰动）；stale 分片**下次访问时**重建
// （不在事件路径上逐分片重建——失效与重建解耦，写路径零阻塞）。
package contextidx

import (
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

// DefaultRebuildThreshold rebuildThreshold 缺省 75（context.md ContextPolicy.cache.rebuildThreshold）。
const DefaultRebuildThreshold = 75

// EnvRebuildThreshold 阈值环境变量名（部署侧可调——调整相关度算法分值分布时同步调阈值）。
const EnvRebuildThreshold = "CORE_SYSTEM_CTX_REBUILD_THRESHOLD"

// ThresholdFromEnv 阈值读取（缺省 75；非法值=缺省不冒充自定义）。
func ThresholdFromEnv() int {
	raw := os.Getenv(EnvRebuildThreshold)
	if raw == "" {
		return DefaultRebuildThreshold
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < 0 || n > 100 {
		return DefaultRebuildThreshold
	}
	return n
}

// Shard 分片缓存条目（TagCtxShard 的进程内投影——派生物，可丢弃可重建）。
type Shard struct {
	TagKey  string    `json:"tagKey"`  // tagType:tag
	DocKeys []string  `json:"docKeys"` // 相关对象集（按 Relevance(tag, doc) 降序）
	Version int64     `json:"version"` // 构建时的索引版本
	Stale   bool      `json:"stale"`   // DocApplied 阈值失效标
	Reason  string    `json:"reason,omitempty"`
	BuiltAt time.Time `json:"builtAt"`
}

// ShardCache 分片缓存（并发安全）。
type ShardCache struct {
	mu        sync.Mutex
	shards    map[string]*Shard
	ix        *Index
	threshold int
}

func newShardCache(ix *Index, threshold int) *ShardCache {
	return &ShardCache{shards: map[string]*Shard{}, ix: ix, threshold: threshold}
}

// Threshold 生效阈值（观测面）。
func (c *ShardCache) Threshold() int { return c.threshold }

// Peek 分片快照读（锁内拷贝——观测/测试面；与 Get 分立：不触发重建不改 stale 态）。
func (c *ShardCache) Peek(tagType, tag string) (Shard, bool) {
	c.mu.Lock()
	defer c.mu.Unlock()
	sh, ok := c.shards[tagType+":"+tag]
	if !ok {
		return Shard{}, false
	}
	return *sh, true
}

// Invalidate DocApplied 阈值失效（B5）：只对 Relevance(更新文档, 分片虚对象) ≥ threshold
// 的分片标 stale——分片虚对象=单 tag 文档（tags=[该 tag]、keywords=tag 切词），
// 与全局公式同一标度。低相关分片零扰动（写放大抑制的机械落点）。
func (c *ShardCache) Invalidate(updated *Doc) {
	c.invalidate(updated, "DocApplied")
}

// invalidate 失效内实现（cause=溯源前缀：DocApplied 更新面 / DocRemoved 删除面——G-18）。
func (c *ShardCache) invalidate(updated *Doc, cause string) {
	if updated == nil {
		return
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	for key, sh := range c.shards {
		if sh.Stale {
			continue // 已 stale 不重复标注（Reason 保留首因）
		}
		shardDoc := shardPseudoDoc(key)
		score, _ := Relevance(updated, shardDoc)
		if score >= float64(c.threshold) {
			sh.Stale = true
			sh.Reason = cause + " " + updated.Key() + " relevance=" + strconv.FormatFloat(Round1(score), 'f', -1, 64)
		}
	}
}

// Get 分片取用（访问时重建语义）：命中未 stale=直接返回；缺席或 stale=重建后返回。
// rebuild=索引内该 tag 全量相关对象重捞（O(N) 索引字段打分——分片级一次，事件路径零成本）。
func (c *ShardCache) Get(tagType, tag string) *Shard {
	key := tagType + ":" + tag
	c.mu.Lock()
	defer c.mu.Unlock()
	sh, ok := c.shards[key]
	if ok && !sh.Stale {
		return sh
	}
	// 重建（缺席/stale 同路径——stale 分片下次访问重建）。
	sh = c.rebuildLocked(key, tagType, tag)
	c.shards[key] = sh
	return sh
}

// rebuildLocked 分片重建（持锁）：捞全索引与该 tag 相关（Relevance>0）的对象集，按分降序。
func (c *ShardCache) rebuildLocked(key, tagType, tag string) *Shard {
	shardDoc := shardPseudoDoc(key)
	c.ix.mu.RLock()
	all := make([]*Doc, 0, len(c.ix.docs))
	for _, d := range c.ix.docs {
		all = append(all, d)
	}
	ver := c.ix.version.Load()
	c.ix.mu.RUnlock()
	type scored struct {
		key   string
		score float64
	}
	rows := make([]scored, 0, len(all))
	for _, d := range all {
		s, _ := Relevance(shardDoc, d)
		if s > 0 {
			rows = append(rows, scored{d.Key(), s})
		}
	}
	sort.SliceStable(rows, func(i, j int) bool {
		if rows[i].score != rows[j].score {
			return rows[i].score > rows[j].score
		}
		return rows[i].key < rows[j].key
	})
	keys := make([]string, 0, len(rows))
	for _, r := range rows {
		keys = append(keys, r.key)
	}
	return &Shard{TagKey: key, DocKeys: keys, Version: ver, Stale: false, BuiltAt: time.Now().UTC()}
}

// shardPseudoDoc 分片虚对象（单 tag 文档——全局公式的同标度输入）。
func shardPseudoDoc(tagKey string) *Doc {
	typ, name, found := strings.Cut(tagKey, ":")
	if !found {
		typ, name = "", tagKey
	}
	return &Doc{
		Tags:     []Tag{{Tag: name, TagType: typ}},
		Keywords: Tokenize(name),
	}
}

// Stats 分片缓存观测（count/stale 词形——console ctx 面后续段数据源）。
func (c *ShardCache) Stats() (total, stale int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for _, sh := range c.shards {
		total++
		if sh.Stale {
			stale++
		}
	}
	return total, stale
}
