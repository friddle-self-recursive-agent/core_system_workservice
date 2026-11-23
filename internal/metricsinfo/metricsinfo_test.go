// MetricsInfo 数据面首段单测（v0.33）：样本保留环形界/订阅扇出每订阅者独立（多源不争抢）/
// 反压即丢不阻塞生产者/查询投影最小聚合集（声明面外聚合不投影不冒充零值）/窗口标签过滤/
// 自观测采样器（窗口计数即样本、首拍全量对账、metric=events.<type> 命名、查询失败零样本
// 不冒充零值）。设计依据：event.md 规则 13 数据面落点+解耦口径（数据面零信号面依赖）。
package metricsinfo

import (
	"context"
	"testing"
	"time"
)

// TestStoreRecordAndFanout 入库+扇出：每订阅者独立通道收到全量样本（多源不争抢）；空指标名
// 不入库。
func TestStoreRecordAndFanout(t *testing.T) {
	st := NewStore()
	a, b := st.Subscribe(), st.Subscribe()
	st.Record(Sample{Metric: "events.DocApplied", Value: 3, Window: "2s"})
	st.Record(Sample{Metric: "", Value: 9, Window: "2s"}) // 空指标名不入库
	for name, ch := range map[string]<-chan Sample{"a": a, "b": b} {
		select {
		case got := <-ch:
			if got.Metric != "events.DocApplied" || got.Value != 3 || got.Window != "2s" {
				t.Fatalf("订阅 %s 收到错误样本: %+v", name, got)
			}
		case <-time.After(time.Second):
			t.Fatalf("订阅 %s 未收到样本（扇出面断裂）", name)
		}
		select {
		case got := <-ch:
			t.Fatalf("订阅 %s 收到不该存在的空指标样本: %+v", name, got)
		default:
		}
	}
	if got := st.Count("events.DocApplied"); got != 1 {
		t.Fatalf("保留计数=%d 期望 1", got)
	}
	if got := st.Metrics(); len(got) != 1 || got[0] != "events.DocApplied" {
		t.Fatalf("指标清单投影=%v 期望 [events.DocApplied]", got)
	}
}

// TestStoreRingCap 环形保留界：超限旧样本让位新样本（进程内可丢弃库内存有界）。
func TestStoreRingCap(t *testing.T) {
	st := NewStore()
	for i := 0; i <= ringCap; i++ { // ringCap+1 条 → 首条让位
		st.Record(Sample{Metric: "m", Value: float64(i), Window: "w"})
	}
	if got := st.Count("m"); got != ringCap {
		t.Fatalf("保留=%d 期望 %d（环形界）", got, ringCap)
	}
	v, ok := st.Query(Query{Metric: "m", Agg: AggLast})
	if !ok || v != float64(ringCap) {
		t.Fatalf("last=%v ok=%v 期望 %d（旧样本让位）", v, ok, ringCap)
	}
}

// TestStoreFanoutDrop 订阅方不取=反压即丢（生产者不阻塞——解耦口径机械落点），缓冲满后
// 继续入库不卡死。
func TestStoreFanoutDrop(t *testing.T) {
	st := NewStore()
	_ = st.Subscribe() // 订阅后从不取
	done := make(chan struct{})
	go func() {
		defer close(done)
		for i := 0; i < subBuf*3; i++ {
			st.Record(Sample{Metric: "m", Value: float64(i), Window: "w"})
		}
	}()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("反压阻塞了生产者（应满即丢不反压）")
	}
}

// TestStoreQueryProjection 查询投影最小集：count/sum/avg/max/min/last+窗口标签过滤+
// 声明面外聚合不投影不冒充零值+无同键样本=not ok。
func TestStoreQueryProjection(t *testing.T) {
	st := NewStore()
	st.Record(Sample{Metric: "m", Value: 1, Window: "2s"})
	st.Record(Sample{Metric: "m", Value: 5, Window: "2s"})
	st.Record(Sample{Metric: "m", Value: 3, Window: "4s"})
	cases := []struct {
		q    Query
		want float64
	}{
		{Query{Metric: "m", Agg: AggCount}, 3},          // 空窗口=跨窗口全量
		{Query{Metric: "m", Window: "2s", Agg: AggCount}, 2},
		{Query{Metric: "m", Window: "2s", Agg: AggSum}, 6},
		{Query{Metric: "m", Window: "2s", Agg: AggAvg}, 3},
		{Query{Metric: "m", Agg: AggMax}, 5},
		{Query{Metric: "m", Agg: AggMin}, 1},
		{Query{Metric: "m", Agg: AggLast}, 3}, // 保留序尾（4s 窗口样本后入）
	}
	for _, c := range cases {
		got, ok := st.Query(c.q)
		if !ok || got != c.want {
			t.Fatalf("Query %+v = (%v,%v) 期望 %v", c.q, got, ok, c.want)
		}
	}
	if _, ok := st.Query(Query{Metric: "m", Agg: "p99"}); ok {
		t.Fatal("声明面外聚合投影了（应不即兴语义 not ok）")
	}
	if _, ok := st.Query(Query{Metric: "absent", Agg: AggCount}); ok {
		t.Fatal("无同键样本投影了零值（应 not ok 不冒充）")
	}
	if _, ok := st.Query(Query{Metric: "m", Window: "9s", Agg: AggCount}); ok {
		t.Fatal("同键不同窗口样本被投影（应窗口过滤 not ok）")
	}
}

// TestSamplerWindowCounts 自观测采样器：首拍全量对账（since 零值）→ metric=events.<type>
// 窗口计数样本；次拍只计上一窗口增量；窗口标签=采样窗口投影。
func TestSamplerWindowCounts(t *testing.T) {
	st := NewStore()
	total := map[string]int{"DocApplied": 3, "MetricThreshold": 1}
	deltas := []map[string]int{{"DocApplied": 2}, {"DocApplied": 1}}
	calls := 0
	counts := func(since time.Time) (map[string]int, error) {
		if since.IsZero() {
			return total, nil // 首拍全量
		}
		d := map[string]int{}
		if calls-1 < len(deltas) {
			d = deltas[calls-1] // calls 已含首拍
		}
		return d, nil
	}
	sp := NewSampler(st, 2*time.Second, counts)
	if !sp.Enabled() || sp.Window() != 2*time.Second || sp.WindowLabel() != "2s" {
		t.Fatalf("采样器开启面错误: enabled=%v window=%s", sp.Enabled(), sp.WindowLabel())
	}
	now := time.Unix(1700000000, 0)
	sp.tick(now, time.Time{}) // 首拍
	if got := st.Count("events.DocApplied"); got != 1 {
		t.Fatalf("首拍样本数=%d 期望 1", got)
	}
	v, ok := st.Query(Query{Metric: "events.DocApplied", Window: "2s", Agg: AggLast})
	if !ok || v != 3 {
		t.Fatalf("首拍窗口计数=(%v,%v) 期望 3（全量对账）", v, ok)
	}
	calls++
	sp.tick(now.Add(2*time.Second), now) // 次拍：增量 2
	v, ok = st.Query(Query{Metric: "events.DocApplied", Window: "2s", Agg: AggCount})
	if !ok || v != 2 {
		t.Fatalf("两拍样本条数=(%v,%v) 期望 2", v, ok)
	}
	// 阈值桥契约面（同形不同源）：样本字段直投不改造由装配层适配——此处断言数据面形态。
	subs := st.Subscribe()
	calls++ // 第三拍前推进 calls（增量表游标——deltas[calls-1]=deltas[1]={DocApplied: 1}）
	sp.tick(now.Add(4*time.Second), now.Add(2*time.Second)) // 第三拍：增量 1
	select {
	case got := <-subs:
		if got.Metric != "events.DocApplied" || got.Value != 1 || got.Window != "2s" {
			t.Fatalf("扇出样本形态错误: %+v", got)
		}
	case <-time.After(time.Second):
		t.Fatal("扇出未收到第三拍样本")
	}
}

// TestSamplerDisabled 查询失败/关断=零样本不冒充零值（数据面照常活，本拍静默）。
func TestSamplerDisabled(t *testing.T) {
	st := NewStore()
	sp := NewSampler(st, time.Second, func(time.Time) (map[string]int, error) {
		return nil, context.Canceled
	})
	sp.tick(time.Unix(1700000000, 0), time.Time{})
	if got := st.Count("events.DocApplied"); got != 0 {
		t.Fatalf("查询失败落了样本=%d（应零样本不冒充）", got)
	}
	if NewSampler(st, 0, nil).Enabled() {
		t.Fatal("窗口 0+无查询函数应关断")
	}
	var zero *Sampler
	if zero.Enabled() {
		t.Fatal("零值采样器应关断")
	}
}
