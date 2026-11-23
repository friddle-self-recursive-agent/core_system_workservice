// deliver.go —— notification 送达投影面（引擎主体第四段 v0.16——真实拉起第二子段；
// v0.17 收尾：受理后的真实通道出站挂接——送达状态机 accepted→sent/failed）。
// 事件语义锚（concepts/event.md）：notification/<通道> 族目标（route 主路与 cc 旁路）
// 在执行面=notification 且装配层已登记接线（TriggerHub.WireExecutor(ExecNotification)）时，
// 触发骨架把该绑定推进为一条 DeliveryRecord（一次送达的受理投影），查询面见
// GET /v1/notifications/deliveries。
//
// 解耦边界（诚实投影，与 run.go 同构）：送达记录=送达已受理 + 通道投影（notification 规则 4
// 的送达记录进 ops DB 属账本落地轮；本投影是进程内观测面非账本，账本=event_log）。
// v0.17 起，装配层可经 DeliveryOptions.Outbound 注入**真实出站执行器**（cored-bridge.md
// §2.3：exec 出站桥接实现侧出站入口，依赖方向不反转——本包不感知通道内部）：受理后异步
// 出站，出站完成把该投影就地翻转为 sent/failed（一次尝试，无重试；重试/降级/静默
// declared-deferred）。未注入出站时语义与 v0.16 完全一致（恒 accepted，零 goroutine）。
//
// route 与 cc 双路（与 run 面的差异点，声明如下）：run 投影只消费 route 主路（v0.15 口径），
// 送达投影同时消费 route 主路与 cc 旁路——通知本身即旁路语义（Escalate 的 root 上报为主路、
// notification 送达为旁路；ApprovalNeeded 则以通知为主路），主/旁维度由 Origin 字段留痕，
// 单一 sink 双路收编（不再为旁路单开 sink 面）。非 notification 执行面（workflow/agent/none）
// 不产生送达投影。
package workflow

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// 送达状态词汇（v0.16 受理；v0.17 扩出站结果——status 就地翻转，一投影一终态）。
const (
	DeliverAccepted = "accepted" // 送达已受理（通道投影在位；出站未完成/未注入）
	DeliverSent     = "sent"     // 真实出站成功（Outbound 注入且返回 nil）
	DeliverFailed   = "failed"   // 真实出站失败（Outbound 返回错误/出站队列满/停机前未起跑）
)

// 送达来源词汇（主/旁维度留痕：同一 sink 双路收编的判别锚）。
const (
	OriginRoute = "route" // route 主路绑定（如 ApprovalNeeded → notification/*）
	OriginCC    = "cc"    // cc 旁路绑定（如 Escalate/HeartbeatLost 的旁路通知）
)

// 出站默认值（DeliveryOptions 零值/越界时的兜底）。
const (
	DefaultOutboundTimeout = 60 * time.Second // 单次出站超时
	DefaultOutboundQueue   = 64               // 出站任务队列容量
	DefaultOutboundWorkers = 1                // 出站协程数（1=串行保序）
)

// DeliveryRecord 一条送达投影（notification 族绑定→送达域的受理留痕；观测面非账本）。
// v0.17 出站结果就地写入同一投影（Status 翻转 + Attempts/CompletedAt/Error 留痕）。
type DeliveryRecord struct {
	Delivery    int64     // 送达序号（本 sink 进程内单调递增，1 起）
	Channel     string    // 通道投影（notification/<通道> 的通道段；*=全部已绑通道）
	Origin      string    // 送达来源（OriginRoute/OriginCC——主路/旁路留痕）
	Status      string    // 送达状态（accepted→sent/failed；未注入出站恒 accepted）
	TriggerSeq  int64     // 来源触发留痕 seq（可回指 GET /v1/triggers records）
	EventID     string    // 与发布方回执同源（apply 响应 event.id / event_log.event_id）
	Type        string    // 事件类型
	DedupKey    string    // 与发布方回执同源
	Level       string    // 事件等级（fatal 必接 Notification 规则的机械观察点）
	Route       string    // 命中行 route 原文
	Owner       string    // owner 解析值
	Target      string    // 执行面锚点投影（=ExecutorBinding.Target）
	At          time.Time // 送达受理时刻（收到留痕同刻）
	Attempts    int       // 出站尝试次数（0=未尝试；v0.17 一投影至多一次尝试）
	CompletedAt time.Time // 出站完成时刻（零值=出站未完成）
	Error       string    // 出站失败原因（成功=空）
}

// DeliveryState 查询快照（DeliveryLog.State() 返回值；apiserver
// GET /v1/notifications/deliveries 载荷来源）。
type DeliveryState struct {
	Accepted   int64            // 累计受理（含已被环挤出的）
	Sent       int64            // 累计出站成功（v0.17）
	Failed     int64            // 累计出站失败（v0.17）
	Dropped    int64            // 环溢出丢弃计数（可观测）
	Deliveries []DeliveryRecord // 环内送达投影（旧→新）
}

// Outbound 真实出站执行器（v0.17 注入点；cored-bridge.md §2.3 进程内自动投递的
// 执行面）。解耦：workflow 包不感知通道内部（im/email/webhook、重试、降级）——实现由
// 装配层注入（exec 桥/未来进程内驱动）；Send 在出站协程内同步调用（可阻塞，受
// OutboundTimeout 约束），返回 nil=成功（status→sent），错误=失败（status→failed，
// 错误文本进投影 Error）。实现不得回调 sink 自身（无递归投影）。
type Outbound interface {
	Send(ctx context.Context, rec DeliveryRecord) error
}

// DeliverySink notification 送达面的送达投影落点（引擎主体第四段 v0.16）。
// 解耦：hub 只把「执行面=notification ∧ 已接线」的绑定推进给 sink，不感知通道内部
// （真实投递/重试/降级——v0.17 起真实出站经 Outbound 注入 sink，hub 仍零感知）；
// 扩展性：真实通道落地换实现不动 hub。实现须**非阻塞**（hub 在订阅路径上锁内同步调用）
// 且**不得回调 hub**（进程内投影面）；origin 区分主/旁路。
type DeliverySink interface {
	Deliver(rec TriggerRecord, b ExecutorBinding, origin string)
}

// DeliveryOptions 送达投影环可调项（零值可用=无出站，v0.16 语义）。
type DeliveryOptions struct {
	// RingSize 送达环容量（缺省 256）：溢出丢最旧、dropped 计数可见。
	RingSize int
	// Outbound 真实出站执行器（nil=不注入，恒 accepted；注入后受理即异步出站）。
	Outbound Outbound
	// OutboundTimeout 单次出站超时（<=0 → DefaultOutboundTimeout）。
	OutboundTimeout time.Duration
	// OutboundQueue 出站任务队列容量（<=0 → DefaultOutboundQueue；满=立即判 failed
	// 不阻塞受理路径，错误=出站队列满）。
	OutboundQueue int
	// OutboundWorkers 出站协程数（<=0 → DefaultOutboundWorkers；1=串行保序）。
	OutboundWorkers int
}

// DeliveryLog 送达投影的进程内落点（DeliverySink 缺省实现；并发安全；与 RunLog 同构）。
// v0.17：Outbound 注入后受理即入队异步出站，完成就地把该投影 Status 翻转为 sent/failed
// 并进计数。停机语义（Close）：在途出站等其完成（受超时约束），未起跑任务**留 accepted
// 不重放**（本投影是观测面非账本，重启不重放——与草稿恢复/账本重放面分野）。
type DeliveryLog struct {
	opts DeliveryOptions

	mu       sync.Mutex
	seq      int64
	ring     []DeliveryRecord
	accepted int64
	sent     int64
	failed   int64
	dropped  int64

	outCtx   context.Context     // 出站协程生存期（Close 取消；per-Send 超时独立派生不受其影响）
	cancelFn context.CancelFunc  // outCtx 配对 cancel（Close 触发——协程退出闸）
	outCh    chan DeliveryRecord // 出站任务队列（投影副本；关停后入队无协程消费=留 accepted）
	outWG    sync.WaitGroup
	closeOne sync.Once
}

// NewDeliveryLog 构建送达投影落点（Outbound=nil 时零 goroutine，v0.16 语义不变）。
func NewDeliveryLog(opts *DeliveryOptions) *DeliveryLog {
	d := &DeliveryLog{}
	if opts != nil {
		d.opts = *opts
	}
	if d.opts.RingSize <= 0 {
		d.opts.RingSize = 256
	}
	if d.opts.Outbound != nil {
		if d.opts.OutboundTimeout <= 0 {
			d.opts.OutboundTimeout = DefaultOutboundTimeout
		}
		if d.opts.OutboundQueue <= 0 {
			d.opts.OutboundQueue = DefaultOutboundQueue
		}
		if d.opts.OutboundWorkers <= 0 {
			d.opts.OutboundWorkers = DefaultOutboundWorkers
		}
		d.outCtx, d.cancelFn = context.WithCancel(context.Background())
		d.outCh = make(chan DeliveryRecord, d.opts.OutboundQueue)
		for i := 0; i < d.opts.OutboundWorkers; i++ {
			d.outWG.Add(1)
			go d.outboundLoop()
		}
	}
	return d
}

// Deliver 记一条送达投影（环满丢最旧，dropped 计数可见）。
// 通道投影取绑定锚点（notification/* → *，全部已绑通道；具体通道名原样）。
// 已注入出站时受理即入队（非阻塞：hub 在订阅路径上锁内同步调用本方法）；队列满=立即判
// failed（错误=出站队列满），受理路径永不阻塞。hub 锁内调用本方法：只取 d.mu、无出站
// 同步等待——出站全部异步。
func (d *DeliveryLog) Deliver(rec TriggerRecord, b ExecutorBinding, origin string) {
	d.mu.Lock()
	d.seq++
	d.accepted++
	if excess := len(d.ring) + 1 - d.opts.RingSize; excess > 0 {
		d.ring = d.ring[excess:]
		d.dropped += int64(excess)
	}
	proj := DeliveryRecord{
		Delivery: d.seq, Channel: b.Target, Origin: origin, Status: DeliverAccepted,
		TriggerSeq: rec.Seq, EventID: rec.EventID, Type: rec.Type, DedupKey: rec.DedupKey,
		Level: rec.Level, Route: rec.Route, Owner: rec.Owner, At: rec.ReceivedAt,
	}
	d.ring = append(d.ring, proj)
	d.mu.Unlock()

	if d.opts.Outbound == nil {
		return
	}
	select {
	case d.outCh <- proj:
	default:
		d.finish(proj.Delivery, fmt.Errorf("出站队列满（容量=%d，受理路径不阻塞）", d.opts.OutboundQueue))
	}
}

// outboundLoop 出站协程：逐条消费队列；Close 取消生存期后退出（队列残留任务留 accepted）。
func (d *DeliveryLog) outboundLoop() {
	defer d.outWG.Done()
	for {
		select {
		case <-d.outCtx.Done():
			return
		case proj := <-d.outCh:
			d.dispatch(proj)
		}
	}
}

// dispatch 执行一次真实出站（per-Send 超时独立派生：Close 不打断在途——在途等完成）。
// attempts 在起跑时进位（队列满/未起跑=0 次）；panic 兜底=判 failed（注入实现缺陷不拖垮 sink）。
func (d *DeliveryLog) dispatch(proj DeliveryRecord) {
	defer func() {
		if r := recover(); r != nil {
			d.finish(proj.Delivery, fmt.Errorf("出站实现 panic: %v", r))
		}
	}()
	d.attempt(proj.Delivery)
	ctx, cancel := context.WithTimeout(context.Background(), d.opts.OutboundTimeout)
	defer cancel()
	d.finish(proj.Delivery, d.opts.Outbound.Send(ctx, proj))
}

// attempt 出站起跑进位（attempts=真实尝试次数；投影已被环挤出=无处可记，只留结果计数）。
func (d *DeliveryLog) attempt(seq int64) {
	d.mu.Lock()
	defer d.mu.Unlock()
	for i := len(d.ring) - 1; i >= 0; i-- {
		if d.ring[i].Delivery == seq {
			d.ring[i].Attempts++
			return
		}
	}
}

// finish 出站结果落投影（就地把同序号投影 Status 翻转 + CompletedAt/Error；attempts 已在
// dispatch 起跑时进位；投影已被环挤出=只进计数——诚实投影：环是观测窗非账本）。
func (d *DeliveryLog) finish(seq int64, err error) {
	status, text := DeliverSent, ""
	if err != nil {
		status, text = DeliverFailed, err.Error()
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	for i := len(d.ring) - 1; i >= 0; i-- {
		if d.ring[i].Delivery == seq {
			d.ring[i].Status = status
			d.ring[i].CompletedAt = time.Now()
			d.ring[i].Error = text
			break
		}
	}
	if status == DeliverSent {
		d.sent++
	} else {
		d.failed++
	}
}

// State 查询快照（拷贝返回，调用方可安全持有；deliveries 旧→新）。
func (d *DeliveryLog) State() DeliveryState {
	d.mu.Lock()
	defer d.mu.Unlock()
	dl := make([]DeliveryRecord, len(d.ring))
	copy(dl, d.ring)
	return DeliveryState{Accepted: d.accepted, Sent: d.sent, Failed: d.failed,
		Dropped: d.dropped, Deliveries: dl}
}

// Close 停机（幂等）：取消出站协程生存期并等在途任务完成（受 per-Send 超时约束——
// per-Send 超时独立派生自 Background，cancel 不打断在途）。队列内未起跑任务留 accepted
// （见类型注释停机语义）。未注入出站=no-op。
func (d *DeliveryLog) Close() {
	d.closeOne.Do(func() {
		if d.opts.Outbound == nil {
			return
		}
		d.cancelFn()
		d.outWG.Wait()
	})
}

// 编译期断言：DeliveryLog 满足 DeliverySink。
var _ DeliverySink = (*DeliveryLog)(nil)
