// heartbeat.go —— 会话心跳账面（ops-log.md §2 KI-003：session_heartbeat 表族读写——
// R-ENG-7 行数收口纯搬家：SessionHeartbeat 行/状态词闭合集/upsert/停滞扫描/全量投影，
// 零语义变化；opslog.go 拆分伴生件）。
package opslog

import (
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// SessionHeartbeat session_heartbeat 表一行（ops-log.md §2 KI-003 心跳账面）：会话活性租约
// 的 ops 侧权威投影——心跳即 upsert（进程内计数在 sessionliveness 包，DB 行=可查询面）。
// 状态词闭合集：state=live（现势）/suspected-stall（判定态词=SuspectedStall，event.md 规则 14b
// ——判定词≠事件词）/long-task（长任务显式化）/healed（自愈恢复留痕）/escalated（自愈无效升级）/
// ended（会话终结——G-08 SessionEnded 语义 DB 落点：真实会话终结信号，调度器 done 判定源）。
type SessionHeartbeat struct {
	SessionID    string    // 会话锚（involved 投影 Session/<id> 的 id 段）
	Ts           time.Time // 本次推进时刻（=updated_at 同值）
	Phase        string    // 相位/推进点词汇（phase-advance|tool-call|artifact-write|inject 等）
	LastAction   string    // 最后动作摘要（有界；不携带凭据/内网信息——R-PUB-1 事件载荷同红线）
	TaskName     string    // 长任务名（LongTaskRunning 判据字段；空=非长任务）
	LongTask     bool      // 声明为长任务（true=watchdog 不判停滞只显式汇报）
	HealAttempts int       // 累计自愈注入次数（连续无效达阈值=升级）
	HealedAt     time.Time // 最近一次 SessionHealed 时刻（零值=未恢复过）
	State        string    // live|suspected-stall|long-task|healed|escalated
	UpdatedAt    time.Time // 行级更新时刻（watchdog 扫描键=MAX(updated_at)）
}

// 心跳状态词闭合集（sessionliveness/watchdog 与读面共用；词汇=agent-session.md §7）。
const (
	SessionStateLive           = "live"
	SessionStateSuspectedStall = "suspected-stall" // 判定态词=SuspectedStall（下划线形=DB 词形）
	SessionStateLongTask       = "long-task"
	SessionStateHealed         = "healed"
	SessionStateEscalated      = "escalated"
	SessionStateEnded          = "ended" // 会话终结（G-08：真实会话终结信号——调度器 done 判定与 agentSession 回填源）
)

// UpsertSessionHeartbeat 落/刷一行心跳（心跳即写=活性租约续租；SQLite 单写者串行）。
func (s *SQLiteStore) UpsertSessionHeartbeat(h SessionHeartbeat) error {
	if h.SessionID == "" {
		return errors.New("opslog: session_heartbeat.session_id 必填（会话锚）")
	}
	if h.State == "" {
		h.State = SessionStateLive
	}
	if h.Ts.IsZero() {
		h.Ts = time.Now()
	}
	if h.UpdatedAt.IsZero() {
		h.UpdatedAt = h.Ts
	}
	const q = `INSERT INTO session_heartbeat
	(session_id, ts, phase, last_action, task_name, long_task, heal_attempts, healed_at, state, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(session_id) DO UPDATE SET
	ts=excluded.ts, phase=excluded.phase, last_action=excluded.last_action,
	task_name=excluded.task_name, long_task=excluded.long_task,
	heal_attempts=excluded.heal_attempts, healed_at=excluded.healed_at,
	state=excluded.state, updated_at=excluded.updated_at`
	healedCol := "" // 可选时间字段：零值=空串（stamp 的零值回退 now 语义不适用——未恢复过≠恢复于 now）
	if !h.HealedAt.IsZero() {
		healedCol = stamp(h.HealedAt)
	}
	_, err := s.db.Exec(q, h.SessionID, stamp(h.Ts), h.Phase, h.LastAction, h.TaskName,
		boolInt(h.LongTask), h.HealAttempts, healedCol, h.State, stamp(h.UpdatedAt))
	if err != nil {
		return fmt.Errorf("opslog: 落 session_heartbeat(%s): %w", h.SessionID, err)
	}
	return nil
}

// StaleSessionHeartbeats updated_at 早于 before 的心跳行（watchdog 扫描投影；按 updated_at
// 升序=最旧停滞优先；limit 收敛防拖库）。
func (s *SQLiteStore) StaleSessionHeartbeats(before time.Time, limit int) ([]SessionHeartbeat, error) {
	if limit <= 0 {
		limit = DefaultApplyLimit
	}
	if limit > MaxApplyLimit {
		limit = MaxApplyLimit
	}
	rows, err := s.db.Query(`SELECT session_id, ts, phase, last_action, task_name, long_task,
heal_attempts, healed_at, state, updated_at FROM session_heartbeat
WHERE updated_at < ? ORDER BY updated_at ASC LIMIT ?`, stamp(before), limit)
	if err != nil {
		return nil, fmt.Errorf("opslog: 查询停滞心跳行: %w", err)
	}
	defer rows.Close()
	return scanSessionHeartbeats(rows)
}

// AllSessionHeartbeats 心跳行全量投影（GET /v1/sessions 读面底层；updated_at 降序最新优先）。
func (s *SQLiteStore) AllSessionHeartbeats(limit int) ([]SessionHeartbeat, error) {
	if limit <= 0 {
		limit = DefaultApplyLimit
	}
	if limit > MaxApplyLimit {
		limit = MaxApplyLimit
	}
	rows, err := s.db.Query(`SELECT session_id, ts, phase, last_action, task_name, long_task,
heal_attempts, healed_at, state, updated_at FROM session_heartbeat
ORDER BY updated_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, fmt.Errorf("opslog: 查询心跳行: %w", err)
	}
	defer rows.Close()
	return scanSessionHeartbeats(rows)
}

func scanSessionHeartbeats(rows *sql.Rows) ([]SessionHeartbeat, error) {
	var out []SessionHeartbeat
	for rows.Next() {
		var h SessionHeartbeat
		var ts, healedAt, updatedAt string
		var longTask int
		if err := rows.Scan(&h.SessionID, &ts, &h.Phase, &h.LastAction, &h.TaskName, &longTask,
			&h.HealAttempts, &healedAt, &h.State, &updatedAt); err != nil {
			return nil, fmt.Errorf("opslog: 读 session_heartbeat 行: %w", err)
		}
		var err error
		if h.Ts, err = time.Parse(time.RFC3339Nano, ts); err != nil {
			return nil, fmt.Errorf("opslog: session_heartbeat.ts 解析 %q: %w", ts, err)
		}
		if h.HealedAt, err = time.Parse(time.RFC3339Nano, healedAt); err != nil && healedAt != "" {
			return nil, fmt.Errorf("opslog: session_heartbeat.healed_at 解析 %q: %w", healedAt, err)
		}
		if h.UpdatedAt, err = time.Parse(time.RFC3339Nano, updatedAt); err != nil {
			return nil, fmt.Errorf("opslog: session_heartbeat.updated_at 解析 %q: %w", updatedAt, err)
		}
		h.LongTask = longTask != 0
		out = append(out, h)
	}
	return out, rows.Err()
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
