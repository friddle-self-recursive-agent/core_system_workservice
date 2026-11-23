// envelope.go —— 出站信封契约（R-ENG-3 契约层，v0.17）。
// cored 进程内自动投递（cored-bridge.md §2.3 v0.17）的跨进程协议面：装配层 exec 出站桥
// 把一次送达投影渲染成本信封（JSON，UTF-8）写至出站命令的 stdin——实现侧（连接器仓）
// 出站入口读信封、映射为自己的 op 参数（如 notify 的邮件直达 op）。依赖方向恒为
// **connector 仓 → 本包**（本包只声明信封形状与渲染口径，不含任何通道实现；装配在部署侧）。
//
// secret 口径（provider.md 规则 2/6）：信封永不携带凭据——实现侧凭据一律引用式取值
// （env:/vault:/keychain:），真实值只在部署侧环境变量，经进程 env 透传到出站命令。
package connectors

// OutboundEnvelope 一次真实出站的信封（exec 出站桥 → 实现侧出站命令，stdin JSON）。
// subject/body 由桥内建渲染器生成（模板 ref 声明面 deferred，notification.md §3）；
// 投递溯源字段与送达投影/触发留痕同源（deliver.go），实现侧可回指查询面
// （GET 送达投影端点）。实现侧约定：消费 stdin 全量 JSON；exit 0=出站成功，
// 非 0=失败（桥记 failed，错误摘要有界留存）。
type OutboundEnvelope struct {
	// 投递溯源（与送达投影载荷同源：delivery/channel/origin/triggerSeq/eventId/
	// type/dedupKey/level/route/owner/at）。
	Delivery   int64  `json:"delivery"`   // 送达序号
	Channel    string `json:"channel"`    // 通道投影（notification/* 的通道段；*=全部已绑通道）
	Origin     string `json:"origin"`     // route（主路）| cc（旁路）
	TriggerSeq int64  `json:"triggerSeq"` // 来源触发留痕 seq
	EventID    string `json:"eventId"`    // 与发布方回执同源
	Type       string `json:"type"`       // 事件类型
	DedupKey   string `json:"dedupKey"`   // 与发布方回执同源
	Level      string `json:"level"`      // 事件等级
	Route      string `json:"route"`      // 命中行 route 原文
	Owner      string `json:"owner"`      // owner 解析值
	At         string `json:"at"`         // 送达受理时刻（RFC3339Nano，UTC）
	// 渲染面（桥内建；模板声明面 deferred）。
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
