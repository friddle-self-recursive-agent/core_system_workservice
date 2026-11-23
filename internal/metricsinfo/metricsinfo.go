// Package metricsinfo MetricsInfo 数据面首段（v0.33，event.md 规则 13 数据面落点）：进程内
// 样本库（可丢弃可重建——DB 非权威同款哲学，TSDB 后端随 WorkServiceTemplate 安装面顺延）
// + 查询投影（queries 最小聚合集）+ 订阅扇出（装配层把订阅通道绑给信号面阈值桥）。
// 解耦口径（声明先行）：数据面零信号面依赖——本包不 import 总线/注册表/信号面，数据面
// 不知道信号面；信号面不持有数据面——装配层持两者并经样本通道适配（Sample 契约不变，
// 契约面零迁移）。数据面照常产样本与投递可丢（事件可丢语义同款：反压即丢不阻塞生产者——
// 下一拍同因窗口聚合兜底）。
package metricsinfo

import (
	"sort"
	"sync"
)

// Sample 数据面样本记录（数据面自有形态——与信号面契约同形不同源，装配层适配；窗口
// 标签由样本侧=生产者声明，数据面透传不改造）。
type Sample struct {
	Metric string  // 指标名（查询与阈值命中键）
	Value  float64 // 样本值
	Window string  // 聚合窗口标签（生产者声明——dedupKey 因由成分随信号面透传）
}

// ringCap 单指标样本保留上限（进程内可丢弃库的界——内存有界，旧样本让位新样本）。
const ringCap = 1024

// subBuf 订阅通道缺省缓冲（装配层绑定面；满即丢不反压——解耦口径的机械落点）。
const subBuf = 64

// Store 进程内样本库（零值不可用，经 NewStore 构建）：Record 入库+扇出，Query 投影，
// Subscribe 供装配层把样本流绑给信号面。
type Store struct {
	mu    sync.Mutex
	rings map[string][]Sample   // 每指标环形保留（追加+越界裁剪）
	subs  map[int]chan Sample   // 订阅扇出（每订阅者独立通道——多源互不争抢样本）
	next  int
}

// NewStore 构建空样本库。
func NewStore() *Store {
	return &Store{rings: map[string][]Sample{}, subs: map[int]chan Sample{}}
}

// Record 入一条样本（保留+扇出；扇出满即丢不反压——数据面不因信号面停滞）。空指标名
// 不入库（无键样本不可寻址）。
func (s *Store) Record(smp Sample) {
	if s == nil || smp.Metric == "" {
		return
	}
	s.mu.Lock()
	ring := append(s.rings[smp.Metric], smp)
	if len(ring) > ringCap {
		ring = ring[len(ring)-ringCap:]
	}
	s.rings[smp.Metric] = ring
	for _, ch := range s.subs {
		select {
		case ch <- smp:
		default: // 订阅方反压即丢（事件可丢语义同款——窗口聚合兜底）
		}
	}
	s.mu.Unlock()
}

// Subscribe 开一条订阅通道（每订阅者独立扇出——装配层每源一通道，多源不争抢）。
func (s *Store) Subscribe() <-chan Sample {
	ch := make(chan Sample, subBuf)
	s.mu.Lock()
	s.subs[s.next] = ch
	s.next++
	s.mu.Unlock()
	return ch
}

// Metrics 已知指标名（排序输出——投影面确定性）。
func (s *Store) Metrics() []string {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]string, 0, len(s.rings))
	for k := range s.rings {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// Count 单指标保留样本数（读面投影）。
func (s *Store) Count(metric string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.rings[metric])
}

// 聚合最小集（查询投影声明面——声明面外聚合不投影不即兴语义）。
const (
	AggCount = "count" // 样本条数
	AggSum   = "sum"   // 样本值合计
	AggAvg   = "avg"   // 样本值均值
	AggMax   = "max"   // 样本值最大
	AggMin   = "min"   // 样本值最小
	AggLast  = "last"  // 最新样本值（保留序尾——时间即入序）
)

// Query 一条查询投影声明（指标+窗口标签+聚合；窗口空=跨窗口全量）。
type Query struct {
	Metric string
	Window string
	Agg    string
}

// Query 查询投影：保留样本按（指标+窗口）过滤后做聚合。ok=false=未装配/未知聚合/
// 无同键样本（缺口可见，不冒充零值）。
func (s *Store) Query(q Query) (float64, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	ring := s.rings[q.Metric]
	var vals []float64
	for _, smp := range ring {
		if q.Window != "" && smp.Window != q.Window {
			continue
		}
		vals = append(vals, smp.Value)
	}
	if len(vals) == 0 {
		return 0, false
	}
	switch q.Agg {
	case AggCount:
		return float64(len(vals)), true
	case AggSum:
		var t float64
		for _, v := range vals {
			t += v
		}
		return t, true
	case AggAvg:
		var t float64
		for _, v := range vals {
			t += v
		}
		return t / float64(len(vals)), true
	case AggMax:
		m := vals[0]
		for _, v := range vals[1:] {
			if v > m {
				m = v
			}
		}
		return m, true
	case AggMin:
		m := vals[0]
		for _, v := range vals[1:] {
			if v < m {
				m = v
			}
		}
		return m, true
	case AggLast:
		return vals[len(vals)-1], true
	default:
		return 0, false // 声明面外聚合不投影
	}
}
