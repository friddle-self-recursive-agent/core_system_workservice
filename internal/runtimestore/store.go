// Package runtimestore 可插拔运行态投影存储（R-ENG 任务书 §1——六类进程内投影记录的
// 统一持久化落点：workflow 触发留痕/对账 run/notification 送达/agent 唤醒/flowbind 派单/
// intake 录入会话）。设计依据：
//   - R-ENG-2 幂等建表迁移（DDL=CREATE TABLE IF NOT EXISTS；既有库无损升列——opslog 同款口径）；
//   - Core 单写者（concepts/core.md §7 串行队列——连接池=1，与 opslog.OpenSQLite 同口径）；
//   - 写穿透 + 启动恢复装配语义（消费包 workflow/flowbind/intake 内存照旧工作，每次变更
//     同步 Append；启动时装载历史行续接序号/游标——sqlite 为权威续接，memory 空启动）。
//
// 行结构字段直投自现有内存投影结构（workflow.TriggerRecord/RunRecord/DeliveryRecord/
// WakeRecord、flowbind.DispatchRecord、intake.SessionRecord），**不含语义改造**：时间统一
// 存 RFC3339Nano 文本列（opslog.stamp 同款词形），复杂字段（切片/map/结构体）存 JSON 文本列。
// 本包零业务语义：只追加/列举/关界（Close），不解释行内容、不发明状态机——消费方各投影
// 面仍是权威（Store 是运行态投影的持久化面，不是账本；账本=event_log/apply_log 归档）。
package runtimestore

// TriggerRow 触发留痕行（直投 workflow.TriggerRecord；Exec/CCExec 执行绑定投影=JSON 文本列）。
type TriggerRow struct {
	Seq        int64 // 收到序号（hub 进程内单调递增，1 起）——幂等主键
	EventID    string
	Type       string
	DedupKey   string
	Source     string
	Involved   string
	Level      string
	At         string // RFC3339Nano（事件时间，发布方时钟）
	ReceivedAt string // RFC3339Nano（收到时刻，订阅方时钟）
	Route      string
	CC         string
	ForPred    string // 命中行 for 谓词原文（列名 for_pred——for 为 SQL 保留词）
	Owner      string
	ExecJSON   string // route 目标执行绑定投影 JSON（workflow.ExecutorBinding）
	CCExecJSON string // cc 目标执行绑定投影 JSON
	Scope      string

	Cursor int64 // List 填充的不透明游标（=Seq；消费方透传给下次 List 的 since）
}

// RunRow 对账 run 行（直投 workflow.RunRecord；Phases/Transitions/Brakes=JSON 文本列）。
type RunRow struct {
	Run             int64 // run 序号（sink 进程内单调递增，1 起）——幂等主键
	Workflow        string
	SpecSource      string
	PhasesJSON      string // 声明相位序列 JSON（[]string）
	TransitionsJSON string // 声明迁移表投影 JSON（[]workflow.TransitionRow）
	Phase           string
	Status          string
	BrakesJSON      string // 生效刹车值投影 JSON（workflow.BrakeView）
	TriggerSeq      int64
	EventID         string
	Type            string
	DedupKey        string
	Route           string
	Owner           string
	Target          string
	Involved        string
	StartedAt       string // RFC3339Nano
	SessionID       string // 会话锚（G-08：loop 会话关联——唤醒桥受理回执回填；空=未回填）
	VerifyJSON      string // Verify 相位硬 gate 判定投影 JSON（workflow.VerifyVerdict；空=未审——G-09/G-12）

	Cursor int64 // 不透明游标（=Run）
}

// DeliveryRow 送达投影行（直投 workflow.DeliveryRecord）。
type DeliveryRow struct {
	Delivery    int64 // 送达序号（sink 进程内单调递增，1 起）——幂等主键
	Channel     string
	Origin      string
	Status      string // accepted→sent/failed（出站结果就地翻转——重放以最新行为准）
	TriggerSeq  int64
	EventID     string
	Type        string
	DedupKey    string
	Level       string
	Route       string
	Owner       string
	Target      string
	At          string // RFC3339Nano
	Attempts    int
	CompletedAt string // RFC3339Nano（零值=空串——出站未完成）
	Error       string

	Cursor int64 // 不透明游标（=Delivery）
}

// WakeRow agent 唤醒投影行（直投 workflow.WakeRecord）。
type WakeRow struct {
	Wake       int64 // 唤醒序号（sink 进程内单调递增，1 起）——幂等主键
	Agent      string
	Origin     string
	Status     string
	Level      string
	TriggerSeq int64
	EventID    string
	Type       string
	DedupKey   string
	Route      string
	Owner      string
	At         string // RFC3339Nano
	SessionID  string // 会话锚（G-08：loop 会话关联——唤醒桥受理回执回填；空=未回填）

	Cursor int64 // 不透明游标（=Wake）
}

// DispatchRow 派单受理投影行（直投 flowbind.DispatchRecord；Acceptance/Nodes=JSON 文本列）。
type DispatchRow struct {
	Seq            int64 // 派单序号（sink 进程内单调递增，1 起）——幂等主键
	Agent          string
	Status         string
	AutoLevel      string
	Action         string
	Kind           string
	Workflow       string
	Scope          string
	Connector      string
	Proposal       string
	Title          string
	Summary        string
	AcceptanceJSON string // 验收标准 JSON（[]string）
	Priority       string
	Proposer       string
	RV             int64
	EventID        string
	NodesJSON      string // 节点清单 JSON（[]flowbind.Node）
	At             string // RFC3339Nano

	Cursor int64 // 不透明游标（=Seq）
}

// IntakeSessionRow 录入会话投影行（直投 intake.SessionRecord；Fields/Sources=JSON 文本列）。
// 与五类流水行不同：会话是**演化态**（同 id 多轮推进）——Append=upsert（同 id 覆写最新投影），
// 重放恢复取最新词形。
type IntakeSessionRow struct {
	SessionID     string // 幂等主键（upsert）
	Kind          string
	Scope         string
	Name          string
	Proposer      string
	Turns         int
	GenVersion    int
	AskingField   string
	Ready         bool
	LastDryRunOk  bool
	FieldErrorCnt int
	ReasonRef     string
	FieldsJSON    string // 字段快照 JSON（map[string]any）
	SourcesJSON   string // 字段来源 JSON（map[string]string）
	CreatedAt     string // RFC3339Nano
	UpdatedAt     string // RFC3339Nano

	Cursor int64 // 不透明游标（存储侧行序——消费方透传不解释）
}

// Store 运行态投影存储口面（六类投影的追加/列举/关界）。实现须并发安全；追加幂等
// （五类流水=同 (kind,id) 重放 noop；会话=同 id upsert 覆写）；List 游标=上次返回行集
// 末行的 Cursor（0=从头），limit≤0=缺省 500（读面防拖库同口径）。驱动经 Register 注册、
// Open 构造（内置 sqlite 与 memory）。
type Store interface {
	AppendTrigger(r TriggerRow) error
	AppendRun(r RunRow) error
	AppendDelivery(r DeliveryRow) error
	AppendWake(r WakeRow) error
	AppendWakeSession(wake int64, sessionID string) error // G-08：唤醒会话锚回填（演化态就地覆写）
	AppendDispatch(r DispatchRow) error
	AppendIntakeSession(r IntakeSessionRow) error

	ListTriggers(limit int, since int64) ([]TriggerRow, error)
	ListRuns(limit int, since int64) ([]RunRow, error)
	ListDeliveries(limit int, since int64) ([]DeliveryRow, error)
	ListWakes(limit int, since int64) ([]WakeRow, error)
	ListDispatches(limit int, since int64) ([]DispatchRow, error)
	ListIntakeSessions(limit int, since int64) ([]IntakeSessionRow, error)

	Close() error
}

// listLimit List 行数归一（≤0=缺省 500——读面防拖库同口径，恢复面分页循环消费）。
func listLimit(limit int) int {
	if limit <= 0 {
		return 500
	}
	return limit
}
