// sampler.go 装配层自观测采样器（v0.33 MetricsInfo 数据面首段真实生产者）：周期查询
// ops DB event_log 流水做窗口计数——查询投影即样本（真实数据面：样本值=上一窗口内归档
// 行数，metric=events.<type>，window=采样窗口标签）。注入面只携一个窄查询函数（opslog
// 零反向依赖——数据面与审计面解耦；装配层在 cmd/cored 完成适配）。与 patrol.Ticker 同
// 范式：缺省关断（窗口<=0 或未注入查询函数=Enabled false 零样本）。
package metricsinfo

import (
	"context"
	"time"
)

// CountsFunc 窗口计数查询函数（装配层注入——opslog.SQLiteStore.EventCountsSince 适配）。
type CountsFunc func(since time.Time) (map[string]int, error)

// Sampler 自观测采样器（cored 装配层持有；零值不可用，经 NewSampler 构建）。
type Sampler struct {
	store  *Store
	window time.Duration
	counts CountsFunc
}

// NewSampler 构建采样器。window<=0 或 counts=nil → 关断（Enabled=false，Run 即返）。
func NewSampler(store *Store, window time.Duration, counts CountsFunc) *Sampler {
	return &Sampler{store: store, window: window, counts: counts}
}

// Enabled 采样器是否开启：窗口>0 ∧ 查询函数已注入 ∧ 库已挂（缺省关断面）。
func (s *Sampler) Enabled() bool {
	return s != nil && s.window > 0 && s.counts != nil && s.store != nil
}

// Window 采样窗口（stderr 装配锚报告用；未开启时返回零值）。
func (s *Sampler) Window() time.Duration {
	if s == nil {
		return 0
	}
	return s.window
}

// Run 采样主循环：每窗口到点自产一批窗口计数样本（ctx 取消即退场，幂等可重启）。
// 未开启（Enabled=false）即返——关断=零样本（机械判别器的关断面）。
func (s *Sampler) Run(ctx context.Context) {
	if !s.Enabled() {
		return
	}
	tk := time.NewTicker(s.window)
	defer tk.Stop()
	var since time.Time // 零值=首拍全量对账（进程起至今——真实数据面的起点口径）
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-tk.C:
			s.tick(now, since) // 查询失败不中断采样（可丢语义——下一窗口重试）
			since = now
		}
	}
}

// WindowLabel 窗口标签口径（采样窗口的机械投影——dedupKey 因由成分二与查询窗口同源）。
func (s *Sampler) WindowLabel() string {
	if s == nil {
		return ""
	}
	return s.window.String()
}

// tick 到点产一批样本：上一窗口（自 since 起）各事件类型归档行数逐类型一条样本
// （metric=events.<type>；首拍 since=零值=全量）。查询失败=本拍零样本（不冒充零值）。
func (s *Sampler) tick(now, since time.Time) {
	counts, err := s.counts(since)
	if err != nil {
		return
	}
	label := s.WindowLabel()
	for typ, n := range counts {
		s.store.Record(Sample{Metric: "events." + typ, Value: float64(n), Window: label})
	}
}
