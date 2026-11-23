// outbound.go —— cored 装配层出站适配器（v0.17 进程内自动投递注入）。
// 字段映射只在此处发生（与 draftOpsStore 同惯例）：workflow（送达投影）与 outbound
// （exec 出站桥）两包叶子解耦——投影 → 出站信封（connectors.OutboundEnvelope）→ 桥。
// 实现（exec 桥指向的命令）住连接器仓/部署侧，cored 不 import 连接器仓——依赖方向不反转。
package main

import (
	"context"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/connectors"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/outbound"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
)

// execOutbound workflow.Outbound 的 exec 桥实现：送达投影渲染为出站信封后交给命令
// （exit 0=成功→sent；非 0/超时=失败→failed，错误摘要有界留存）。
type execOutbound struct{ cmd *outbound.Command }

// Send 一次真实出站（在送达投出站协程内被调用，可阻塞——受 OutboundTimeout 约束）。
func (e execOutbound) Send(ctx context.Context, rec workflow.DeliveryRecord) error {
	env := connectors.OutboundEnvelope{
		Delivery:   rec.Delivery,
		Channel:    rec.Channel,
		Origin:     rec.Origin,
		TriggerSeq: rec.TriggerSeq,
		EventID:    rec.EventID,
		Type:       rec.Type,
		DedupKey:   rec.DedupKey,
		Level:      rec.Level,
		Route:      rec.Route,
		Owner:      rec.Owner,
		At:         rec.At.UTC().Format(time.RFC3339Nano),
	}
	outbound.Render(&env) // 内建渲染（模板 ref 声明面 deferred）
	return e.cmd.Send(ctx, env)
}
