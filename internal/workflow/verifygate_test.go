// verifygate_test.go —— Verify 相位硬 gate 验收（G-09/G-12，R-TYPE-3 接线）：
// 卡 Verify 不放行（findings 非空）/ gated 计数 / Escalate 上报钩子 / 放行判定留痕 /
// 判定写穿透恢复 / nil gate 既有行为面零翻转 / 竞态复验放弃面。
package workflow

import (
	"strings"
	"testing"
	"time"
)

// gateStub 可编程 gate 假体（按调用序出判定——卡→放行两拍语义锚）。
type gateStub struct {
	verdicts []VerifyVerdict
	calls    int
	saw      []RunRecord
}

func (g *gateStub) Verify(run RunRecord) VerifyVerdict {
	g.calls++
	g.saw = append(g.saw, run)
	if g.calls > len(g.verdicts) {
		return VerifyVerdict{OK: true, AuditID: "audit-ok", Note: "audit ok"}
	}
	return g.verdicts[g.calls-1]
}

// TestVerifyGateBlocksRun 卡 Verify：Verifying 出行先过 gate——findings 非空判定 → run 相位/
// 状态原样保留（不放行）、gated 计数可见、OnVerifyBlocked 上报、判定投影随 run 留痕。
func TestVerifyGateBlocksRun(t *testing.T) {
	runs := NewRunLog(nil)
	seedAdvance(t, runs, "workflows/core-workflow", "core-workflow", Verifying, RunRunning)
	blocked := []RunRecord{}
	gate := &gateStub{verdicts: []VerifyVerdict{{OK: false, AuditID: "audit-x",
		ErrorFindings: 2, Note: "R-PUB-1 违规样例×2"}}}
	runs.opts.VerifyGate = gate
	runs.opts.OnVerifyBlocked = func(run RunRecord, v VerifyVerdict) { blocked = append(blocked, run) }

	if n := runs.Advance(); n != 0 {
		t.Fatalf("卡 Verify 拍应零迁移，得 %d", n)
	}
	r := runAt(runs, 0)
	if r.Phase != string(Verifying) || r.Status != RunRunning {
		t.Fatalf("卡 Verify 应原样保留 {Verifying, running}: {phase=%s, status=%s}", r.Phase, r.Status)
	}
	if r.Verify == nil || r.Verify.OK || r.Verify.ErrorFindings != 2 || r.Verify.AuditID != "audit-x" {
		t.Fatalf("判定投影应随 run 留痕：%+v", r.Verify)
	}
	if st := runs.State(); st.Gated != 1 {
		t.Fatalf("gated 计数应=1，得 %d", st.Gated)
	}
	if len(blocked) != 1 || blocked[0].Run != r.Run {
		t.Fatalf("Escalate 上报应一次且携 run：%v", blocked)
	}
	// 修复后（gate 放行判定）下拍即过——重审通过放行到后继相位。
	if n := runs.Advance(); n != 1 {
		t.Fatalf("放行拍应迁移 1 步，得 %d", n)
	}
	if r := runAt(runs, 0); r.Phase != string(Reporting) || !r.Verify.OK {
		t.Fatalf("放行后应 {Reporting} 且判定 OK：{phase=%s, ok=%v}", r.Phase, r.Verify != nil && r.Verify.OK)
	}
	if len(gate.saw) > 0 && gate.saw[0].Phase != string(Verifying) {
		t.Fatalf("gate 审核输入应为 Verifying 快照：%s", gate.saw[0].Phase)
	}
}

// TestVerifyGateNilPassthrough gate 未挂接=既有行为面零翻转：Verifying 直接过（不写判定投影）。
func TestVerifyGateNilPassthrough(t *testing.T) {
	runs := NewRunLog(nil)
	seedAdvance(t, runs, "workflows/core-workflow", "core-workflow", Verifying, RunRunning)
	if n := runs.Advance(); n != 1 {
		t.Fatalf("无 gate 应直推 1 步，得 %d", n)
	}
	r := runAt(runs, 0)
	if r.Phase != string(Reporting) || r.Verify != nil {
		t.Fatalf("无 gate 不写判定投影（诚实词形）: {phase=%s, verify=%v}", r.Phase, r.Verify)
	}
}

// TestVerifyGateVerdictPersists 判定写穿透恢复：卡 Verify 的判定（VerifyJSON 列）随 Store
// 往返——写入→重开→恢复后 run 仍 {Verifying, 判定在}（卡 Verify 证据重启仍在——G-01 判据面）。
func TestVerifyGateVerdictPersists(t *testing.T) {
	store := newTestStore(t)
	runs := NewRunLog(&RunOptions{Store: store, VerifyGate: VerifyGateFunc(func(run RunRecord) VerifyVerdict {
		return VerifyVerdict{OK: false, AuditID: "audit-p", ErrorFindings: 1, Note: "违规"}
	})})
	runs.StartRun(TriggerRecord{
		Seq: 1, EventID: "ev-g", Type: "DocApplied", DedupKey: "apply:g:1",
		ReceivedAt: time.Now(), Route: "workflows/core-workflow", Owner: "global",
	}, ExecutorBinding{Executor: ExecWorkflow, Target: "core-workflow"})
	runs.mu.Lock()
	runs.ring[len(runs.ring)-1].Phase = string(Verifying) // 直达 Verify 前置态
	runs.mu.Unlock()
	if n := runs.Advance(); n != 0 {
		t.Fatalf("卡 Verify 拍应零迁移，得 %d", n)
	}
	// 重启恢复（同 Store 重开 RunLog——G-01 重启恢复判据的 workflow 侧锚）。
	runs2 := NewRunLog(&RunOptions{Store: store})
	st := runs2.State()
	if len(st.Runs) != 1 {
		t.Fatalf("恢复应得 1 条 run：%v", st.Runs)
	}
	r := st.Runs[0]
	if r.Phase != string(Verifying) || r.Verify == nil || r.Verify.OK || r.Verify.AuditID != "audit-p" {
		t.Fatalf("重启后卡 Verify 证据应仍在：{phase=%s, verify=%+v}", r.Phase, r.Verify)
	}
	if !strings.Contains(r.Verify.Note, "违规") {
		t.Fatalf("判定摘要应往返无损：%+v", r.Verify)
	}
}
