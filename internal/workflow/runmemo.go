// runmemo.go —— run 上下文账（2026-11-28 用户令「developer-system 拉代码 build？接啊 / 代码
// 正常编辑了？接啊」——E3 context 实体传递的引擎半边）。
//
// 为什么要有：节点产出的**实体内容**（ops 结果全文/judge 结论/plan 项）原先只活在两处——
// history Note（≤220 字摘要）与落库文档（DocumentDiff）——下游节点/会话拿不到真内容：
// batch-judge 读不到 gh-issue-fetch 的 issue 列表、developer-system 会话只收到字段名列表
// （contextPassing.fields=[planItem,judge]——run #134 诚实 no-op 的根因）。本账=按 run 键的
// 进程内上下文：上游节点写（ops 桥写 op 结果全文/runner 写 plan 项与结论/fanout launcher
// 写子 run 计划项），下游读者取（runner 组 prompt/service 桥组 briefing）。
//
// 边界（诚实投影）：内存账非持久账——cored 重启即清；持久实体=落库文档（WorkflowPlan/
// DocumentDiff，重启后读者按锚回读，miss=诚实缺席不编造）。写读全带锁；键=词形自由
// （约定键见各写方注释：op:<node>=op 结果全文、planItem=子 run 计划项、planSummary=judge
// 结论、sessionNote=service 会话回执摘要）。
package workflow

import "sync"

// RunMemo run 上下文账（进程内；并发安全；零值可用）。
type RunMemo struct {
	mu   sync.Mutex
	data map[int64]map[string]any
}

// NewRunMemo 构建空账。
func NewRunMemo() *RunMemo { return &RunMemo{data: map[int64]map[string]any{}} }

// Set 写一个上下文键（同 run 同键覆写=最新词形；val 任意 JSON 形——读者自行断言）。
func (m *RunMemo) Set(run int64, key string, val any) {
	if m == nil || run <= 0 || key == "" {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	row, ok := m.data[run]
	if !ok {
		row = map[string]any{}
		m.data[run] = row
	}
	row[key] = val
}

// Get 读一个上下文键（ok=false=缺席——诚实缺席不编造；拷贝语义由调用方约定：
// map/slice 值读者只读不改）。
func (m *RunMemo) Get(run int64, key string) (any, bool) {
	if m == nil || run <= 0 {
		return nil, false
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	row, ok := m.data[run]
	if !ok {
		return nil, false
	}
	v, ok := row[key]
	return v, ok
}

// GetString 读字符串便捷形（缺席/非字符串=空串）。
func (m *RunMemo) GetString(run int64, key string) string {
	v, ok := m.Get(run, key)
	if !ok {
		return ""
	}
	s, _ := v.(string)
	return s
}

// GetMap 读 map 便捷形（缺席/非 map=nil）。
func (m *RunMemo) GetMap(run int64, key string) map[string]any {
	v, ok := m.Get(run, key)
	if !ok {
		return nil
	}
	mp, _ := v.(map[string]any)
	return mp
}

// Drop 清一个 run 的整行（收束后回收——环挤出清账防长驻泄漏；缺失=幂等 noop）。
func (m *RunMemo) Drop(run int64) {
	if m == nil || run <= 0 {
		return
	}
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.data, run)
}
