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
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
)

// wakeNotifier 包装现有 WakeSink：先记投影，再按白名单发宿主唤醒桥；受理回执的会话锚
// 回填 run/唤醒投影（G-08——/v1/loop/current 活体 loop+sessionId 判据的数据源：宿主回执
// sessionId 优先，缺席=宿主进程句柄 pid:<n> 最小非空锚——诚实降级词形，装配注记随 stderr）。
type wakeNotifier struct {
	inner    workflow.WakeSink
	runs     *workflow.RunLog
	wakes    *workflow.WakeLog
	url      string
	hc       *http.Client
	triggers map[string]bool
}

// newWakeNotifier 构建唤醒桥（addr 可为 host:port 或 http://host:port）。
func newWakeNotifier(inner workflow.WakeSink, runs *workflow.RunLog, wakes *workflow.WakeLog, addr string) *wakeNotifier {
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
		runs:     runs,
		wakes:    wakes,
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
		// 会话锚回执（G-08）：宿主回执 {status, pid, sessionId}——sessionId 优先（dsh 会话 id
		// 真锚），缺席=pid:<n>（宿主进程句柄最小非空锚——真实会话 id 由宿主面升级后回填词形）。
		// 回填 run 与唤醒投影（export.go AttachSessionID——写穿透随 Store；未命中只注记不翻转）。
		var ack struct {
			Status    string `json:"status"`
			PID       int64  `json:"pid"`
			SessionID string `json:"sessionId"`
		}
		_ = json.NewDecoder(resp.Body).Decode(&ack)
		sess := strings.TrimSpace(ack.SessionID)
		if sess == "" && ack.PID > 0 {
			sess = fmt.Sprintf("pid:%d", ack.PID)
		}
		if sess != "" {
			if !w.runs.AttachSessionID(rec.EventID, sess) {
				log.Printf("cored: wake 桥会话锚回填未命中 run（event=%s——事件未落 run 或已挤出，缺口可见）", rec.EventID)
			}
			if !w.wakes.AttachSessionID(rec.EventID, sess) {
				log.Printf("cored: wake 桥会话锚回填未命中唤醒投影（event=%s）", rec.EventID)
			}
		}
		log.Printf("cored: wake 桥已受理（type=%s event=%s agent=%s status=%s session=%s）", rec.Type, rec.EventID, b.Target, ack.Status, sess)
	}()
}
