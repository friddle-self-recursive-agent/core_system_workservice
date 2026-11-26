// wake_bridge.go —— 唤醒桥（事件→自循环拉起）。
//
// 设计：core_design/concepts/cli.md / rule_ux.md R-UX-8（命令行提交事件触发系统自循环）。
// 链路：事件总线 → dispatch → agent 族绑定 → WakeNotifier（本文件）→ 宿主 plugin
// /wake/spawn → 部署侧 loop 命令拉起 core-headless 会话。
//
// 边界：
//   - 只在部署侧显式配置 CORE_SYSTEM_WAKE_ADDR 时启用；缺省=纯投影（既有行为面零翻转）。
//   - 只转发白名单事件类型（缺省 LoopRequested）且 agent 锚以 root 开头；其余唤醒零出站。
//   - 出站只 POST 固定载荷到宿主 loopback，**不执行任何命令**；命令由宿主侧配置。
//   - 异步发送（事件订阅路径不被 HTTP 阻塞）；失败只 stderr 注记，不翻转事件受理。
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
)

// wakeNotifier 包装现有 WakeSink：先记投影，再按白名单发宿主唤醒桥。
type wakeNotifier struct {
	inner    workflow.WakeSink
	url      string
	hc       *http.Client
	triggers map[string]bool
}

// newWakeNotifier 构建唤醒桥（addr 可为 host:port 或 http://host:port）。
func newWakeNotifier(inner workflow.WakeSink, addr string) *wakeNotifier {
	addr = strings.TrimSpace(addr)
	if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
		addr = "http://" + addr
	}
	triggers := map[string]bool{}
	for _, t := range strings.Split(os.Getenv("CORE_SYSTEM_WAKE_TYPES"), ",") {
		if t = strings.TrimSpace(t); t != "" {
			triggers[t] = true
		}
	}
	if len(triggers) == 0 {
		triggers["LoopRequested"] = true
	}
	return &wakeNotifier{
		inner:    inner,
		url:      strings.TrimRight(addr, "/") + "/wake/spawn",
		hc:       &http.Client{Timeout: 5 * time.Second},
		triggers: triggers,
	}
}

// Wake 实现 workflow.WakeSink（非阻塞：投影同步落，桥异步发）。
func (w *wakeNotifier) Wake(rec workflow.TriggerRecord, b workflow.ExecutorBinding, origin string) {
	w.inner.Wake(rec, b, origin)
	if !w.triggers[rec.Type] || !strings.HasPrefix(b.Target, "root") {
		return
	}
	payload := map[string]any{
		"agent":    b.Target,
		"origin":   origin,
		"type":     rec.Type,
		"eventId":  rec.EventID,
		"dedupKey": rec.DedupKey,
		"level":    rec.Level,
	}
	go func() {
		body, err := json.Marshal(payload)
		if err != nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		req, err := http.NewRequestWithContext(ctx, http.MethodPost, w.url, bytes.NewReader(body))
		if err != nil {
			log.Printf("cored: wake 桥请求构建失败: %v", err)
			return
		}
		req.Header.Set("Content-Type", "application/json")
		resp, err := w.hc.Do(req)
		if err != nil {
			log.Printf("cored: wake 桥调用失败（type=%s event=%s）: %v", rec.Type, rec.EventID, err)
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode >= 300 {
			log.Printf("cored: wake 桥拒绝（type=%s event=%s status=%d）", rec.Type, rec.EventID, resp.StatusCode)
			return
		}
		log.Printf("cored: wake 桥已受理（type=%s event=%s agent=%s）", rec.Type, rec.EventID, b.Target)
	}()
}
