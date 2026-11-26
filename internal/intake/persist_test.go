// persist_test.go —— 录入会话投影写穿透/启动恢复验收（G-01 第六类投影——intake
// SessionRecord）：sqlite 写穿透往返（Turn 演化 upsert→重启恢复→最新词形）、包级装配锚
// 注入（SetDefaultStore→NewEngine 拾取）、AttachStore 显式通道、nil=纯内存既有行为面。
package intake

import (
	"path/filepath"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/runtimestore"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// testSch 最小 scheme（Turn 需要 kind 注册表——widgets 缺席=零字段照常推进）。
// 复用既有 loadRealScheme fixture（兄弟仓符号链接 worktrees/core_system_workservice/core_design）。
func testSch(t *testing.T) *scheme.Scheme {
	t.Helper()
	return loadRealScheme(t)
}

// TestSessionWriteThroughRestore G-01 锚：Turn 演化 upsert → 重启恢复（同 Store 重开
// Engine）→ 最新词形仍在（写入→重启→GET /v1/intake/sessions 数据仍在的 sink 侧判据）。
func TestSessionWriteThroughRestore(t *testing.T) {
	sch := testSch(t)
	path := filepath.Join(t.TempDir(), "rt.db")
	store, err := runtimestore.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	e := NewEngine(sch, nil, nil, nil)
	e.AttachStore(store)
	res, err := e.Turn(t.Context(), TurnRequest{Kind: "Proposal", Message: "探针会话"})
	if err != nil {
		t.Fatalf("Turn 应可用: %v", err)
	}
	res2, err := e.Turn(t.Context(), TurnRequest{SessionID: res.SessionID, Kind: "Proposal", Message: "第二轮"})
	if err != nil {
		t.Fatal(err)
	}
	if res2.Turns != 2 {
		t.Fatalf("第二轮 turns 应=2: %d", res2.Turns)
	}
	_ = store.Close() // 模拟重启

	store2, err := runtimestore.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer store2.Close()
	e2 := NewEngine(sch, nil, nil, nil)
	e2.AttachStore(store2) // 启动恢复
	got, err := e2.GetSession(res.SessionID)
	if err != nil {
		t.Fatalf("重启后会话应仍在: %v", err)
	}
	if got.Turns != 2 || got.Kind != "Proposal" {
		t.Fatalf("恢复应取最新词形（turns=2）：%+v", got)
	}
	if rows := e2.ListSessions(); len(rows) != 1 {
		t.Fatalf("恢复清单应 1 行：%d", len(rows))
	}
}

// TestSetDefaultStoreInjection 包级装配锚：SetDefaultStore → NewEngine 自动挂接+恢复；
// 收口=关断词形（nil 注入后新建 Engine 纯内存）。
func TestSetDefaultStoreInjection(t *testing.T) {
	sch := testSch(t)
	store, err := runtimestore.Open("sqlite", filepath.Join(t.TempDir(), "rt.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	SetDefaultStore(store)
	defer SetDefaultStore(nil) // 收口：不影响其他测试
	e := NewEngine(sch, nil, nil, nil)
	if _, err := e.Turn(t.Context(), TurnRequest{Kind: "Proposal"}); err != nil {
		t.Fatal(err)
	}
	rows, err := store.ListIntakeSessions(0, 0)
	if err != nil || len(rows) != 1 {
		t.Fatalf("包级注入后写穿透应落 Store：%d %v", len(rows), err)
	}
	// 恢复：同 Store 新建 Engine（包级锚仍在）自动装载。
	e2 := NewEngine(sch, nil, nil, nil)
	if rows := e2.ListSessions(); len(rows) != 1 {
		t.Fatalf("包级锚恢复应 1 行：%d", len(rows))
	}
}

// TestNilStorePureMemory nil Store=既有行为面零翻转（纯内存，零落库）。
func TestNilStorePureMemory(t *testing.T) {
	sch := testSch(t)
	SetDefaultStore(nil)
	e := NewEngine(sch, nil, nil, nil)
	if _, err := e.Turn(t.Context(), TurnRequest{Kind: "Proposal"}); err != nil {
		t.Fatal(err)
	}
	if rows := e.ListSessions(); len(rows) != 1 {
		t.Fatalf("纯内存会话应照常：%d", len(rows))
	}
}
