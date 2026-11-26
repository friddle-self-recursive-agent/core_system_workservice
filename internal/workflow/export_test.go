// export_test.go —— 投影导出面验收（G-08：事件详情↔最近 loop↔session 链——会话锚回填/
// LoopSession/LoopWake accessor/写穿透恢复）与测试 Store 假体锚。
package workflow

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/runtimestore"
)

// newTestStore 测试用 sqlite 运行态投影存储（进程内文件库——写穿透/重启恢复语义锚同口径）。
func newTestStore(t *testing.T) runtimestore.Store {
	t.Helper()
	s, err := runtimestore.Open("sqlite", filepath.Join(t.TempDir(), "rt.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = s.Close() })
	return s
}

// loopRec loop 触发留痕（LoopRequested——自循环拉起源事件；route=对账锚）。
func loopRec(seq int64) TriggerRecord {
	return TriggerRecord{
		Seq: seq, EventID: "ev-loop", Type: LoopType, DedupKey: "loop:1", Source: "apply",
		Level: "info", ReceivedAt: time.Now(), Route: "workflows/core-workflow", Owner: "global",
	}
}

// TestAttachSessionIDLoopSession G-08 主锚：loop run/唤醒落环 → 桥回执回填 SessionID →
// LoopSession/LoopWake accessor 非空投影（/v1/loop/current 读面判据：活体 loop+sessionId）。
func TestAttachSessionIDLoopSession(t *testing.T) {
	runs := NewRunLog(nil)
	wakes := NewWakeLog(nil)
	runs.StartRun(loopRec(1), ExecutorBinding{Executor: ExecWorkflow, Target: "core-workflow"})
	wakes.Wake(loopRec(1), ExecutorBinding{Executor: ExecAgent, Target: "root"}, OriginRoute)

	if r, ok := runs.LoopSession(); !ok || r.SessionID != "" {
		t.Fatalf("回填前 SessionID 应空（不冒充已拉起）：%+v ok=%v", r, ok)
	}
	if w, ok := wakes.LoopWake(); !ok || w.SessionID != "" {
		t.Fatalf("回填前 wake SessionID 应空：%+v", w)
	}
	if !runs.AttachSessionID("ev-loop", "dsh-loop-7") {
		t.Fatal("run 会话回填应命中")
	}
	if !wakes.AttachSessionID("ev-loop", "dsh-loop-7") {
		t.Fatal("wake 会话回填应命中")
	}
	r, ok := runs.LoopSession()
	if !ok || r.SessionID != "dsh-loop-7" || r.Type != LoopType {
		t.Fatalf("LoopSession 应携回填锚：%+v ok=%v", r, ok)
	}
	w, ok := wakes.LoopWake()
	if !ok || w.SessionID != "dsh-loop-7" {
		t.Fatalf("LoopWake 应携回填锚：%+v", w)
	}
	if runs.AttachSessionID("ev-missing", "x") || wakes.AttachSessionID("", "x") {
		t.Fatal("未命中/空锚回填应=false（缺口可见不静默）")
	}
}

// TestLoopSessionRestore G-01+G-08 交叉锚：回填→写穿透→重启恢复（同 Store 重开 run/wake
// 落点）→ accessor 仍得非空会话锚（写入→重启→读面数据仍在判据的会话关联面）。
func TestLoopSessionRestore(t *testing.T) {
	store := newTestStore(t)
	runs := NewRunLog(&RunOptions{Store: store})
	wakes := NewWakeLog(&WakeOptions{Store: store})
	runs.StartRun(loopRec(1), ExecutorBinding{Executor: ExecWorkflow, Target: "core-workflow"})
	wakes.Wake(loopRec(1), ExecutorBinding{Executor: ExecAgent, Target: "root"}, OriginRoute)
	if !runs.AttachSessionID("ev-loop", "dsh-loop-9") || !wakes.AttachSessionID("ev-loop", "dsh-loop-9") {
		t.Fatal("回填应命中")
	}
	runs2 := NewRunLog(&RunOptions{Store: store})
	wakes2 := NewWakeLog(&WakeOptions{Store: store})
	r, ok := runs2.LoopSession()
	if !ok || r.SessionID != "dsh-loop-9" {
		t.Fatalf("重启后 loop 会话锚应仍在：%+v ok=%v", r, ok)
	}
	w, ok := wakes2.LoopWake()
	if !ok || w.SessionID != "dsh-loop-9" {
		t.Fatalf("重启后 wake 会话锚应仍在：%+v ok=%v", w, ok)
	}
}
