// deliver.go —— notification 送达投影面（引擎主体第四段 v0.16——真实拉起第二子段）。
// 事件语义锚（concepts/event.md）：notification/<通道> 族目标（route 主路与 cc 旁路）
// 在执行面=notification 且装配层已登记接线（TriggerHub.WireExecutor(ExecNotification)）时，
// 触发骨架把该绑定推进为一条 DeliveryRecord（一次送达的受理投影），查询面见
// GET /v1/notifications/deliveries。
//
// 解耦边界（诚实投影，与 run.go 同构）：送达记录=送达已受理 + 通道投影（notification 规则 4
// 的送达记录进 ops DB 属真实通道投递落地轮；本投影是进程内观测面非账本，账本=event_log），
// **不含真实通道投递**（im/email/webhook 经 Provider/ConnectorOps 的出站投递仍属引擎主体
// 余量，cored-bridge.md §2.3）——送达投影出现 ≠ 已送达真实通道，只等于「触发面已把
// notification 族绑定接进送达域」。
//
// route 与 cc 双路（与 run 面的差异点，声明如下）：run 投影只消费 route 主路（v0.15 口径），
// 送达投影同时消费 route 主路与 cc 旁路——通知本身即旁路语义（Escalate 的 root 上报为主路、
// notification 送达为旁路；ApprovalNeeded 则以通知为主路），主/旁维度由 Origin 字段留痕，
// 单一 sink 双路收编（不再为旁路单开 sink 面）。非 notification 执行面（workflow/agent/none）
// 不产生送达投影。
package workflow

import (
	"sync"
	"time"
)

// 送达状态词汇（v0.16 只有受理；真实通道投递落地轮在此扩枚举）。
const (
	DeliverAccepted = "accepted" // 送达已受理（通道投影在位；投递结果 deferred）
)

// 送达来源词汇（主/旁维度留痕：同一 sink 双路收编的判别锚）。
const (
	OriginRoute = "route" // route 主路绑定（如 ApprovalNeeded → notification/*）
	OriginCC    = "cc"    // cc 旁路绑定（如 Escalate/HeartbeatLost 的旁路通知）
)

// DeliveryRecord 一条送达投影（notification 族绑定→送达域的受理留痕；观测面非账本）。
type DeliveryRecord struct {
	Delivery   int64     // 送达序号（本 sink 进程内单调递增，1 起）
	Channel    string    // 通道投影（notification/<通道> 的通道段；*=全部已绑通道）
	Origin     string    // 送达来源（OriginRoute/OriginCC——主路/旁路留痕）
	Status     string    // 送达状态（v0.16 恒 DeliverAccepted）
	TriggerSeq int64     // 来源触发留痕 seq（可回指 GET /v1/triggers records）
	EventID    string    // 与发布方回执同源（apply 响应 event.id / event_log.event_id）
	Type       string    // 事件类型
	DedupKey   string    // 与发布方回执同源
	Level      string    // 事件等级（fatal 必接 Notification 规则的机械观察点）
	Route      string    // 命中行 route 原文
	Owner      string    // owner 解析值
	Target     string    // 执行面锚点投影（=ExecutorBinding.Target）
	At         time.Time // 送达受理时刻（收到留痕同刻）
}

// DeliveryState 查询快照（DeliveryLog.State() 返回值；apiserver
// GET /v1/notifications/deliveries 载荷来源）。
type DeliveryState struct {
	Accepted   int64            // 累计受理（含已被环挤出的）
	Dropped    int64            // 环溢出丢弃计数（可观测）
	Deliveries []DeliveryRecord // 环内送达投影（旧→新）
}

// DeliverySink notification 送达面的送达投影落点（引擎主体第四段 v0.16）。
// 解耦：hub 只把「执行面=notification ∧ 已接线」的绑定推进给 sink，不感知通道内部
// （真实投递/重试/降级）；扩展性：真实通道落地换实现不动 hub。实现须**非阻塞**（hub 在
// 订阅路径上同步调用）且**不得回调 hub**（进程内投影面）；origin 区分主/旁路。
type DeliverySink interface {
	Deliver(rec TriggerRecord, b ExecutorBinding, origin string)
}

// DeliveryOptions 送达投影环可调项（零值可用）。
type DeliveryOptions struct {
	// RingSize 送达环容量（缺省 256）：溢出丢最旧、dropped 计数可见。
	RingSize int
}

// DeliveryLog 送达投影的进程内落点（DeliverySink 缺省实现；并发安全；与 RunLog 同构）。
type DeliveryLog struct {
	opts DeliveryOptions

	mu       sync.Mutex
	seq      int64
	ring     []DeliveryRecord
	accepted int64
	dropped  int64
}

// NewDeliveryLog 构建送达投影落点。
func NewDeliveryLog(opts *DeliveryOptions) *DeliveryLog {
	d := &DeliveryLog{}
	if opts != nil {
		d.opts = *opts
	}
	if d.opts.RingSize <= 0 {
		d.opts.RingSize = 256
	}
	return d
}

// Deliver 记一条送达投影（环满丢最旧，dropped 计数可见）。
// 通道投影取绑定锚点（notification/* → *，全部已绑通道；具体通道名原样）。
func (d *DeliveryLog) Deliver(rec TriggerRecord, b ExecutorBinding, origin string) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.seq++
	d.accepted++
	if excess := len(d.ring) + 1 - d.opts.RingSize; excess > 0 {
		d.ring = d.ring[excess:]
		d.dropped += int64(excess)
	}
	d.ring = append(d.ring, DeliveryRecord{
		Delivery: d.seq, Channel: b.Target, Origin: origin, Status: DeliverAccepted,
		TriggerSeq: rec.Seq, EventID: rec.EventID, Type: rec.Type, DedupKey: rec.DedupKey,
		Level: rec.Level, Route: rec.Route, Owner: rec.Owner, At: rec.ReceivedAt,
	})
}

// State 查询快照（拷贝返回，调用方可安全持有；deliveries 旧→新）。
func (d *DeliveryLog) State() DeliveryState {
	d.mu.Lock()
	defer d.mu.Unlock()
	dl := make([]DeliveryRecord, len(d.ring))
	copy(dl, d.ring)
	return DeliveryState{Accepted: d.accepted, Dropped: d.dropped, Deliveries: dl}
}

// 编译期断言：DeliveryLog 满足 DeliverySink。
var _ DeliverySink = (*DeliveryLog)(nil)
