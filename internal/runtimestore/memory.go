// memory.go —— 运行态投影存储进程内切片实现（R-ENG 任务书 §1 memory 件：测试/无 DB
// 部署兜底）。与 sqlite 驱动同一 Store 口面：追加幂等（五类流水=同 id 重放 noop；会话=
// 同 id upsert 覆写）；列举=游标升序分页。零持久化（进程生命周期即存储——空启动口径，
// 消费方恢复面拿到零行=行为与未挂 Store 一致）。
package runtimestore

import (
	"sort"
	"sync"
)

// MemoryStore 进程内切片存储（并发安全；dsn 忽略）。
type MemoryStore struct {
	mu        sync.Mutex
	triggers  []TriggerRow
	runs      []RunRow
	deliver   []DeliveryRow
	wakes     []WakeRow
	dispatch  []DispatchRow
	sessions  map[string]IntakeSessionRow
	sessOrder []int64 // 会话插入序（rowid 模拟——游标分页语义与 sqlite 对齐）
	sessSeq   int64
}

// openMemory 构建进程内存储（registry 内置驱动入口；dsn 忽略）。
func openMemory(_ string) (Store, error) {
	return &MemoryStore{sessions: map[string]IntakeSessionRow{}}, nil
}

// Close 关界（进程内实现=no-op）。
func (m *MemoryStore) Close() error { return nil }

// AppendTrigger 追加触发留痕（同 seq 重放 noop）。
func (m *MemoryStore) AppendTrigger(r TriggerRow) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, x := range m.triggers {
		if x.Seq == r.Seq {
			return nil
		}
	}
	m.triggers = append(m.triggers, r)
	return nil
}

// AppendRun 追加对账 run（同 run 重放 noop）。
func (m *MemoryStore) AppendRun(r RunRow) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, x := range m.runs {
		if x.Run == r.Run {
			return nil
		}
	}
	m.runs = append(m.runs, r)
	return nil
}

// AppendDelivery 追加送达投影（同 delivery 重放 noop）。
func (m *MemoryStore) AppendDelivery(r DeliveryRow) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, x := range m.deliver {
		if x.Delivery == r.Delivery {
			return nil
		}
	}
	m.deliver = append(m.deliver, r)
	return nil
}

// AppendWake 追加唤醒投影（同 wake 重放 noop；session_id 非空=演化回填词形——
// 受理行已在账时就地覆写会话锚，与 sqlite AppendWakeSession 同语义合并面）。
func (m *MemoryStore) AppendWake(r WakeRow) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.wakes {
		if m.wakes[i].Wake == r.Wake {
			if r.SessionID != "" {
				m.wakes[i].SessionID = r.SessionID // 演化回填（受理词形不变，只补会话锚）
			}
			return nil
		}
	}
	m.wakes = append(m.wakes, r)
	return nil
}

// AppendWakeSession 回填唤醒会话锚（同 wake 行就地覆写 session_id；行缺席=noop——与
// sqlite 驱动同口径，受理未落库时回填无锚可挂）。
func (m *MemoryStore) AppendWakeSession(wake int64, sessionID string) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for i := range m.wakes {
		if m.wakes[i].Wake == wake {
			m.wakes[i].SessionID = sessionID
			return nil
		}
	}
	return nil
}

// AppendDispatch 追加派单投影（同 seq 重放 noop）。
func (m *MemoryStore) AppendDispatch(r DispatchRow) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	for _, x := range m.dispatch {
		if x.Seq == r.Seq {
			return nil
		}
	}
	m.dispatch = append(m.dispatch, r)
	return nil
}

// AppendIntakeSession 追加/更新会话（同 id upsert——覆写最新投影；插入序游标保持首轮值）。
func (m *MemoryStore) AppendIntakeSession(r IntakeSessionRow) error {
	m.mu.Lock()
	defer m.mu.Unlock()
	if old, ok := m.sessions[r.SessionID]; ok {
		r.Cursor = old.Cursor // 游标=首轮插入序（演化态覆写不换行位）
	} else {
		m.sessSeq++
		r.Cursor = m.sessSeq
		m.sessOrder = append(m.sessOrder, m.sessSeq)
	}
	m.sessions[r.SessionID] = r
	return nil
}

// ListTriggers 触发行集（seq 升序；since=游标；limit≤0=缺省 500）。
func (m *MemoryStore) ListTriggers(limit int, since int64) ([]TriggerRow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rows := append([]TriggerRow{}, m.triggers...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].Seq < rows[j].Seq })
	out := []TriggerRow{}
	for _, r := range rows {
		if r.Seq <= since {
			continue
		}
		r.Cursor = r.Seq
		out = append(out, r)
		if len(out) >= listLimit(limit) {
			break
		}
	}
	return out, nil
}

// ListRuns run 行集（run 升序）。
func (m *MemoryStore) ListRuns(limit int, since int64) ([]RunRow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rows := append([]RunRow{}, m.runs...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].Run < rows[j].Run })
	out := []RunRow{}
	for _, r := range rows {
		if r.Run <= since {
			continue
		}
		r.Cursor = r.Run
		out = append(out, r)
		if len(out) >= listLimit(limit) {
			break
		}
	}
	return out, nil
}

// ListDeliveries 送达行集（delivery 升序）。
func (m *MemoryStore) ListDeliveries(limit int, since int64) ([]DeliveryRow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rows := append([]DeliveryRow{}, m.deliver...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].Delivery < rows[j].Delivery })
	out := []DeliveryRow{}
	for _, r := range rows {
		if r.Delivery <= since {
			continue
		}
		r.Cursor = r.Delivery
		out = append(out, r)
		if len(out) >= listLimit(limit) {
			break
		}
	}
	return out, nil
}

// ListWakes 唤醒行集（wake 升序）。
func (m *MemoryStore) ListWakes(limit int, since int64) ([]WakeRow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rows := append([]WakeRow{}, m.wakes...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].Wake < rows[j].Wake })
	out := []WakeRow{}
	for _, r := range rows {
		if r.Wake <= since {
			continue
		}
		r.Cursor = r.Wake
		out = append(out, r)
		if len(out) >= listLimit(limit) {
			break
		}
	}
	return out, nil
}

// ListDispatches 派单行集（seq 升序）。
func (m *MemoryStore) ListDispatches(limit int, since int64) ([]DispatchRow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	rows := append([]DispatchRow{}, m.dispatch...)
	sort.Slice(rows, func(i, j int) bool { return rows[i].Seq < rows[j].Seq })
	out := []DispatchRow{}
	for _, r := range rows {
		if r.Seq <= since {
			continue
		}
		r.Cursor = r.Seq
		out = append(out, r)
		if len(out) >= listLimit(limit) {
			break
		}
	}
	return out, nil
}

// ListIntakeSessions 会话行集（插入序游标升序——与 sqlite rowid 语义对齐）。
func (m *MemoryStore) ListIntakeSessions(limit int, since int64) ([]IntakeSessionRow, error) {
	m.mu.Lock()
	defer m.mu.Unlock()
	out := []IntakeSessionRow{}
	order := append([]int64{}, m.sessOrder...)
	sort.Slice(order, func(i, j int) bool { return order[i] < order[j] })
	byCursor := map[int64]IntakeSessionRow{}
	for _, s := range m.sessions {
		byCursor[s.Cursor] = s
	}
	for _, cur := range order {
		if cur <= since {
			continue
		}
		out = append(out, byCursor[cur])
		if len(out) >= listLimit(limit) {
			break
		}
	}
	return out, nil
}
