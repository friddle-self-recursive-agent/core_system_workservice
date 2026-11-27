// serviceexec.go —— service 节点 session 执行面（2026-11-27 用户裁决「builtin 的 service 的
// agent 可以用 session 来解决。突破下限制」——workflow.md §1.3）。
//
// 语义：run 推进落到 unit=service 节点（developer-system / verify / deploy-system）时，run
// **挂起在该节点**（status=suspended，suspendWhy=service，history 行 on=service-spawn），
// 同拍产出一条 ServiceSpawnRecord 拉起投影——引擎为 service 的 agent（ServiceDesc.instances
// .agentRefs，装配面解析回填）**开一个执行 session**：contextPassing 上一轮上下文作为首轮
// 输入注入（wake 桥/task-spawn 桥同族通路——固定命令、载荷经 stdin、零 shell 拼接）；
// session 心跳终结（部署侧发布 ServiceSessionCompleted 事件）→ run 沿表续行（ServiceSignal），
// 续行 history 行回填真实会话锚。**突破两条既有限制**：①service 节点执行从 declared-deferred
// 变真实通路；②「workflow 直挂 agent 必须 headless」只约束 workflow 直接挂载的 agent 节点，
// service 内部 agent 挂 ServiceDesc.instances.agentRefs，由 service 以 session 驱动（agent.md
// 两类形态边界注记同账）。
//
// 零翻转口径（与 VerifyGate/Brakes 同范式）：RunOptions.ServiceSpawns 缺席（nil）=既有行为面
// 零翻转——service 节点照旧纯查表推进，不挂起不拉起（装配层显式挂接才生效）。
//
// 解耦边界（诚实投影，与 wake.go/dispatch.go 同构）：拉起投影=引擎已受理该节点的 session
// 拉起 + agent 锚投影——**投影出现 ≠ session 已执行完毕**（执行面在 session 域：心跳/终结
// 随 sessionliveness 与部署侧完成事件；本面只证明引擎面的拉起接线在位）。会话锚回执
// （sessionId 优先，缺席=pid:<n> 降级锚——wake 桥 G-08 同词形）经 AttachSession 异步回填。
package workflow

import (
	"strings"
	"sync"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/runtimestore"
)

// service 信号族（completion 信号——scheme.yaml eventTypes 注册后经总线主干到达；hub 按
// 族转发，审批信号族 ResumeSignal 同款「run 域迁移事件不是新对账触发」口径——不另起 run）。
const (
	// SignalServiceDone service 执行 session 完成信号（部署侧 session 退出后经事件注入面
	// 发布；involved=Workflow/<scope>/<name> 寻径，Run 定向可选）。
	SignalServiceDone = "ServiceSessionCompleted"
)

// isServiceSignalType 事件类型是否 service 信号族（hub 转发闸——廉价前置判定）。
func isServiceSignalType(typ string) bool {
	return typ == SignalServiceDone
}

// service 节点挂起/拉起投影状态词。
const (
	// SuspendService 挂起因由词（RunRecord.SuspendWhy——「相位机等待外部信号」的细分投影：
	// service=session 执行中，与审批闸锚挂起区分）。
	SuspendService = "service"
	// SpawnSpawned 拉起已受理（桥回执前/桥缺席的稳态——投影在位不冒充已拉起）。
	SpawnSpawned = "spawned"
	// SpawnCompleted session 完成（部署侧完成信号回填——run 已可续行）。
	SpawnCompleted = "completed"
	// SpawnFailed session 失败退出（run 保持挂起——缺口可见，人不批不续行）。
	SpawnFailed = "failed"
	// EdgeServiceSpawn service 落点迁移行词形（history on——迁移+挂起一体留痕）。
	EdgeServiceSpawn = "service-spawn"
	// EdgeServiceSpawnResume service 完成续行行词形（history on——SessionID=真实会话锚回执）。
	EdgeServiceSpawnResume = "service-resume"
)

// serviceNodeOf 相位 → service 执行单元节点（纯查投影；unit=service 才命中——gate/agent/
// connectops 节点不落本面）。
func serviceNodeOf(nodes []NodeView, phase string) (NodeView, bool) {
	for _, n := range nodes {
		if n.Phase == phase && n.Unit == "service" {
			return n, true
		}
	}
	return NodeView{}, false
}

// ServiceSpawner run 域 service 节点 session 拉起面（可选能力——RunOptions.ServiceSpawns
// 挂接；nil=零翻转口径）。实现须**非阻塞**（Advance 在订阅路径锁内收集、锁外派发）且
// **不得回调 RunLog**（进程内投影面；会话锚回执走 AttachSession 独立入口）。
type ServiceSpawner interface {
	// SpawnFromRun 一次拉起受理：run 快照 + 落点节点投影（agent 解析随实现侧声明面——
	// ServiceDesc.instances.agentRefs 由装配层解析回填，引擎不感知文档库）。
	SpawnFromRun(run RunRecord, node NodeView)
}

// ServiceSignalSink service 完成信号面（hub 可选挂接——BindServiceSignals；装配层桥实现：
// 翻拉起投影终态后转交 RunLog.ServiceSignal 续行并携带会话锚）。非阻塞契约同族（锁内转发）。
type ServiceSignalSink interface {
	ServiceSignal(rec TriggerRecord)
}

// ServiceSpawnRecord 一条 service 节点 session 拉起投影（观测面非账本，账本=event_log 归档
// + run 投影 history 行）。
type ServiceSpawnRecord struct {
	Seq      int64     // 拉起序号（本 sink 进程内单调递增，1 起）
	Run      int64     // 来源 run 序号（可回指 GET /v1/workflows/runs）
	Workflow string    // workflow 锚名
	Node     string    // 落点节点名（spec.phases 富声明 name）
	Phase    string    // 落点相位锚（run.Phase 同词域）
	Service  string    // 目标系统名（developer-system / deploy-system…）
	Op       string    // 调用 op（develop|acceptance|release…）
	Agent    string    // service 的 agent 锚（ServiceDesc.instances.agentRefs 解析；空=未解析——诚实缺席）
	Role     string    // permissionRole 锚（P7 权限随文档；空=未声明）
	Context  []string  // contextPassing.fields（上一轮上下文声明字段——随载荷透传）
	Status   string    // spawned|completed|failed（完成/失败由部署侧信号回填）
	Session  string    // 会话锚回执（dsh sessionId 优先，缺席=pid:<n> 降级锚；空=未回填）
	At       time.Time // 拉起受理时刻
}

// ServiceSpawnState 查询快照（ServiceSpawnLog.State() 返回值）。
type ServiceSpawnState struct {
	Spawned int64                // 累计拉起受理（含已被环挤出的）
	Dropped int64                // 环溢出丢弃计数（可观测）
	Records []ServiceSpawnRecord // 环内拉起投影（旧→新）
}

// ServiceSpawnOptions 拉起投影环可调项（零值可用）。
type ServiceSpawnOptions struct {
	// RingSize 拉起环容量（缺省 256）：溢出丢最旧、dropped 计数可见。
	RingSize int
	// Store 运行态投影存储挂接（R-ENG §1 可选能力——nil=纯内存零翻转）。挂接后写穿透
	// （受理/回填/终态每次变更同步 Append——可演化投影 upsert 同 seq 覆写）+ 启动恢复。
	Store runtimestore.Store
}

// ServiceSpawnLog service 拉起投影的进程内落点（ServiceSpawner 缺省实现；并发安全；
// 与 WakeLog/DispatchLog 同构——无出站机器，出站由装配层包装注入，wake 桥同范式）。
type ServiceSpawnLog struct {
	opts ServiceSpawnOptions

	mu       sync.Mutex
	seq      int64
	ring     []ServiceSpawnRecord
	spawned  int64
	dropped  int64
}

// NewServiceSpawnLog 构建拉起投影落点。挂有 Store 时先做启动恢复（装载历史行续接 seq/
// spawned 与环内最近投影；恢复的是投影视图非拉起动作重放——dispatch.go 同口径）。
func NewServiceSpawnLog(opts *ServiceSpawnOptions) *ServiceSpawnLog {
	s := &ServiceSpawnLog{}
	if opts != nil {
		s.opts = *opts
	}
	if s.opts.RingSize <= 0 {
		s.opts.RingSize = 256
	}
	if s.opts.Store != nil {
		rows := loadServiceSpawnRows(s.opts.Store)
		if len(rows) > 0 {
			recs := make([]ServiceSpawnRecord, len(rows))
			maxSeq := int64(0)
			for i, r := range rows {
				recs[i] = serviceSpawnFromRow(r)
				if r.Seq > maxSeq {
					maxSeq = r.Seq
				}
			}
			s.seq = maxSeq
			s.spawned = int64(len(rows))
			if excess := len(recs) - s.opts.RingSize; excess > 0 {
				recs = recs[excess:]
			}
			s.ring = recs
		}
	}
	return s
}

// SpawnFromRun 记一条拉起投影（ServiceSpawner 缺省实现——agent 留空=未解析诚实缺席；
// ServiceDesc.instances.agentRefs 解析由装配层包装注入，引擎侧不感知文档库）。
func (s *ServiceSpawnLog) SpawnFromRun(run RunRecord, node NodeView) {
	s.Dispatch(ServiceSpawnRecord{
		Run: run.Run, Workflow: run.Workflow, Node: node.Name, Phase: node.Phase,
		Service: node.Service, Op: node.Op, Role: node.Role, Context: node.ContextFields,
		Status: SpawnSpawned, At: time.Now(),
	})
}

// Dispatch 记一条拉起投影（环满丢最旧，dropped 计数可见；装配层桥在 SpawnFromRun 基础上
// 先解析 agent 再走本入口落投影——flowbind DispatchLog 同构）。返回落账后的存储词形
// （Seq=受理序号——桥异步回执会话锚回填的定位锚；单飞命中=返回零值记录 ok 语义由 Seq=0
// 判别）。同 run 同节点重复受理=noop 单飞（Advance 挂起面天然幂等——挂起 run 不再推进，
// 不会重复落点；此防御兜底人为重放）。
func (s *ServiceSpawnLog) Dispatch(rec ServiceSpawnRecord) ServiceSpawnRecord {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.ring {
		if s.ring[i].Run == rec.Run && s.ring[i].Node == rec.Node && s.ring[i].Status == SpawnSpawned {
			return ServiceSpawnRecord{Seq: 0} // 已有在途拉起（单飞幂等——不重复受理）
		}
	}
	s.seq++
	s.spawned++
	if excess := len(s.ring) + 1 - s.opts.RingSize; excess > 0 {
		s.ring = s.ring[excess:]
		s.dropped += int64(excess)
	}
	rec.Seq = s.seq
	if rec.At.IsZero() {
		rec.At = time.Now()
	}
	if rec.Status == "" {
		rec.Status = SpawnSpawned
	}
	s.ring = append(s.ring, rec)
	// 写穿透（R-ENG §1）：受理词形同步 Append（可演化投影 upsert——同 seq 覆写）。
	if s.opts.Store != nil {
		persistAppend("service 拉起受理", s.opts.Store.AppendServiceSpawn(serviceSpawnToRow(s.ring[len(s.ring)-1])))
	}
	return s.ring[len(s.ring)-1]
}

// AttachSession 会话锚回执回填（装配层桥异步调用——dsh sessionId 优先，缺席=pid:<n> 降级锚；
// 未命中只注记不翻转——wake 桥 G-08 同口径）。
func (s *ServiceSpawnLog) AttachSession(seq int64, sess string) bool {
	sess = strings.TrimSpace(sess)
	if sess == "" {
		return false
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := range s.ring {
		if s.ring[i].Seq == seq {
			s.ring[i].Session = sess
			if s.opts.Store != nil {
				persistAppend("service 拉起会话回填", s.opts.Store.AppendServiceSpawn(serviceSpawnToRow(s.ring[i])))
			}
			return true
		}
	}
	return false
}

// MarkCompleted 完成信号回填（装配层桥在 ServiceSessionCompleted 到达时调用——按 run+节点
// 定位在途拉起并翻终态词形）。返回被翻词形的记录拷贝与会话锚（桥转交 RunLog.ServiceSignal
// 作续行 history 的会话锚）；未命中（已被环挤出/重复信号）=false，重复完成信号幂等 noop。
func (s *ServiceSpawnLog) MarkCompleted(run int64, node, status string) (ServiceSpawnRecord, string, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for i := len(s.ring) - 1; i >= 0; i-- {
		if s.ring[i].Run == run && s.ring[i].Node == node && s.ring[i].Status == SpawnSpawned {
			s.ring[i].Status = status
			if s.opts.Store != nil {
				persistAppend("service 拉起终态", s.opts.Store.AppendServiceSpawn(serviceSpawnToRow(s.ring[i])))
			}
			return s.ring[i], s.ring[i].Session, true
		}
	}
	return ServiceSpawnRecord{}, "", false
}

// State 查询快照（拷贝返回，调用方可安全持有；records 旧→新）。
func (s *ServiceSpawnLog) State() ServiceSpawnState {
	s.mu.Lock()
	defer s.mu.Unlock()
	rs := make([]ServiceSpawnRecord, len(s.ring))
	copy(rs, s.ring)
	return ServiceSpawnState{Spawned: s.spawned, Dropped: s.dropped, Records: rs}
}

// 编译期断言：ServiceSpawnLog 满足 ServiceSpawner。
var _ ServiceSpawner = (*ServiceSpawnLog)(nil)
