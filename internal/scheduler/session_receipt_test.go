// session_receipt_test.go —— 会话终结→done 判定验收（G-08：Task 的 done 判定接真实会话
// 终结信号 SessionEnded；done 时回填 Task.status.agentSession=真实 sessionId）。
// 真 sqlite 仓库 + 真 opslog 心跳账面端到端锚（scheduler_integration_test 同款装配）。
package scheduler

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// newReceiptFixture 真 sqlite 仓库 + 真 ops DB 的 Tick 装配（会话终结判定源挂接）。
func newReceiptFixture(t *testing.T) (context.Context, *repository.SQLiteRepo, *Scheduler, *opslog.SQLiteStore) {
	t.Helper()
	sch, err := scheme.LoadFile(coreSchemePath)
	if err != nil {
		t.Fatalf("装载 scheme: %v", err)
	}
	repo, err := repository.OpenSQLite(t.TempDir()+"/docs.db", sch, nil, "")
	if err != nil {
		t.Fatalf("打开 sqlite: %v", err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	ops, err := opslog.OpenSQLite(filepath.Join(t.TempDir(), "ops.db"))
	if err != nil {
		t.Fatalf("打开 ops DB: %v", err)
	}
	t.Cleanup(func() { _ = ops.Close() })
	bus := eventbus.NewBus(&eventbus.Options{})
	s := New(repo, bus, sch, nil, nil, Options{Sessions: ops})
	ctx := context.Background()
	return ctx, repo, s, ops
}

// writeTaskObject 落一个指定 phase/agentSession 的 Task 对象（身份字段住 frontmatter 顶层
// ——layout.md §3 与既有 taskDocEmptyStatus fixture 同形）。
func writeTaskObject(t *testing.T, ctx context.Context, repo *repository.SQLiteRepo, scope, name, phase, sess string) {
	t.Helper()
	doc := "---\nkind: Task\nscope: " + scope + "\nname: " + name +
		"\nspec:\n  title: 会话终结探针\nstatus:\n  phase: " + phase + "\n"
	if sess != "" {
		doc += "  agentSession: " + sess + "\n"
	}
	doc += "---\n探针正文（会话终结→done 判定）\n"
	if _, _, err := repo.Write(ctx, &repository.Object{
		Kind: "Task", Scope: scope, Name: name, Content: []byte(doc),
	}, "test", ""); err != nil {
		t.Fatalf("写入 Task: %v", err)
	}
}

// phaseOf 读回任务现行 phase 与 agentSession。
func phaseOf(t *testing.T, ctx context.Context, repo *repository.SQLiteRepo, scope, name string) (string, string) {
	t.Helper()
	obj, err := repo.Get(ctx, "Task", scope, name)
	if err != nil {
		t.Fatalf("读回: %v", err)
	}
	_, st, ok := Parse(obj.Content)
	if !ok {
		t.Fatalf("读回不可解析：%s", obj.Content)
	}
	return st.Phase, st.AgentSession
}

// TestSessionEndDrivesDone 精确路：running 任务 agentSession=真实会话锚 → 心跳行置 ended
// → 一拍后 done 且 agentSession 非空（判据锚：done→agentSession 非空）。
func TestSessionEndDrivesDone(t *testing.T) {
	ctx, repo, s, ops := newReceiptFixture(t)
	writeTaskObject(t, ctx, repo, "self-recursive-core", "t-end", PhaseRunning, "dsh-sess-42")
	now := time.Now()
	if err := ops.UpsertSessionHeartbeat(opslog.SessionHeartbeat{
		SessionID: "dsh-sess-42", Ts: now, Phase: "phase-advance", State: opslog.SessionStateLive, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	s.Tick(ctx) // live 行=非终结——零动作（对照面）
	if phase, _ := phaseOf(t, ctx, repo, "self-recursive-core", "t-end"); phase != PhaseRunning {
		t.Fatalf("live 行不得判 done（非 SessionEnded 语义），得 %q", phase)
	}
	if err := ops.UpsertSessionHeartbeat(opslog.SessionHeartbeat{
		SessionID: "dsh-sess-42", Ts: now, Phase: "session-end", State: opslog.SessionStateEnded, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	s.Tick(ctx) // ended 行=真实会话终结 → done
	phase, sess := phaseOf(t, ctx, repo, "self-recursive-core", "t-end")
	if phase != PhaseDone {
		t.Fatalf("会话终结后应 done，得 %q", phase)
	}
	if sess != "dsh-sess-42" {
		t.Fatalf("agentSession 应为真实 sessionId（非空判据），得 %q", sess)
	}
}

// TestSessionEndBackfillsMissingHint 兜底路：spawn hint 缺席（agentSession 空）→ 会话心跳
// 携任务锚的 ended 行 → done + agentSession 由终结行回填（「无真实 sessionId 回填」缺口的收口）。
func TestSessionEndBackfillsMissingHint(t *testing.T) {
	ctx, repo, s, ops := newReceiptFixture(t)
	writeTaskObject(t, ctx, repo, "self-recursive-core", "t-hint", PhaseVerifying, "")
	now := time.Now()
	if err := ops.UpsertSessionHeartbeat(opslog.SessionHeartbeat{
		SessionID: "dsh-sess-77", Ts: now, TaskName: "self-recursive-core/t-hint",
		Phase: "session-end", State: opslog.SessionStateEnded, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	s.Tick(ctx)
	phase, sess := phaseOf(t, ctx, repo, "self-recursive-core", "t-hint")
	if phase != PhaseDone {
		t.Fatalf("兜底路终结行应判 done，得 %q", phase)
	}
	if sess != "dsh-sess-77" {
		t.Fatalf("兜底路应回填行 sessionId，得 %q", sess)
	}
}

// TestStallNotDone 停滞词形不冒充终结：suspected-stall/escalated 行≠done（判定只认
// SessionEnded 语义——反证面）。
func TestStallNotDone(t *testing.T) {
	ctx, repo, s, ops := newReceiptFixture(t)
	writeTaskObject(t, ctx, repo, "self-recursive-core", "t-stall", PhaseRunning, "dsh-sess-9")
	now := time.Now()
	if err := ops.UpsertSessionHeartbeat(opslog.SessionHeartbeat{
		SessionID: "dsh-sess-9", Ts: now, State: opslog.SessionStateSuspectedStall, UpdatedAt: now,
	}); err != nil {
		t.Fatal(err)
	}
	s.Tick(ctx)
	phase, sess := phaseOf(t, ctx, repo, "self-recursive-core", "t-stall")
	if phase != PhaseRunning {
		t.Fatalf("停滞行不得判 done，得 %q", phase)
	}
	if !strings.Contains(phase, "running") || sess == "" {
		t.Fatalf("agentSession 应保持：%q", sess)
	}
}
