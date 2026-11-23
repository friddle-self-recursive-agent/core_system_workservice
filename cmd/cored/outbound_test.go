// outbound_test.go —— cored 装配层出站适配器验收：送达投影→出站信封字段映射+
// 内建渲染经 exec 桥端到端落地（tee 落文件回读断言）。
package main

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/connectors"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/outbound"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
)

// 编译期保险丝：execOutbound 必须满足 workflow.Outbound（装配面契约漂移此处即炸）。
var _ workflow.Outbound = execOutbound{}

// TestExecOutboundEnvelope 适配器映射：workflow 投影 → connectors.OutboundEnvelope
// 全字段同源 + 渲染填充，经 exec 桥（tee）真实落盘验证协议面。
func TestExecOutboundEnvelope(t *testing.T) {
	dir := t.TempDir()
	got := filepath.Join(dir, "envelope.json")
	argv, err := outbound.ParseCommand(`tee ` + got)
	if err != nil {
		t.Fatal(err)
	}
	bridge, err := outbound.NewCommand(argv)
	if err != nil {
		t.Fatal(err)
	}
	ad := execOutbound{cmd: bridge}
	at := time.Date(2026, 11, 24, 3, 4, 5, 0, time.UTC)
	rec := workflow.DeliveryRecord{
		Delivery: 9, Channel: "mail-chan", Origin: workflow.OriginCC, Status: workflow.DeliverAccepted,
		TriggerSeq: 4, EventID: "ev-77", Type: "Escalate", DedupKey: "esc:wf:denied",
		Level: "error", Route: "root-agent", Owner: "global", At: at,
	}
	if err := ad.Send(context.Background(), rec); err != nil {
		t.Fatalf("适配器出站应成功: %v", err)
	}
	raw, err := os.ReadFile(got)
	if err != nil {
		t.Fatal(err)
	}
	var env connectors.OutboundEnvelope
	if err := json.Unmarshal(raw, &env); err != nil {
		t.Fatalf("落盘内容非合法信封: %v（%s）", err, raw)
	}
	if env.Delivery != 9 || env.Channel != "mail-chan" || env.Origin != "cc" ||
		env.TriggerSeq != 4 || env.EventID != "ev-77" || env.Type != "Escalate" ||
		env.DedupKey != "esc:wf:denied" || env.Level != "error" || env.Route != "root-agent" ||
		env.Owner != "global" || env.At != "2026-11-24T03:04:05Z" {
		t.Fatalf("信封溯源字段映射不符: %+v", env)
	}
	if env.Subject == "" || env.Body == "" {
		t.Fatalf("渲染应填充 subject/body: %+v", env)
	}
}
