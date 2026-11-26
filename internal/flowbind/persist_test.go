// persist_test.go —— 派单投影写穿透/启动恢复验收（G-01 收口——R-ENG §1 消费侧）：
// sqlite 写穿透往返（受理→重启恢复→seq/accepted 续接）、投影字段无损、幂等重放、
// memory 空启动口径。
package flowbind

import (
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/runtimestore"
)

// TestDispatchStoreRoundTripRestore 写穿透→重启恢复：受理落 Store → 同 Store 重开 DispatchLog
// → 环内投影恢复（seq/accepted 续接；载荷字段无损）→ 新派单 seq 续接（判据：写入→重启→
// GET /v1/agents/dispatches 数据仍在的 sink 侧锚）。
func TestDispatchStoreRoundTripRestore(t *testing.T) {
	path := t.TempDir() + "/rt.db"
	store, err := runtimestore.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	d := NewDispatchLog(&DispatchOptions{Store: store})
	d.Dispatch(DispatchRecord{Agent: DispatchAgent, Status: DispatchAccepted, AutoLevel: DispatchAutoLevel,
		Action: "created", Kind: DispatchKindWorkflow, Workflow: "probe-wf", Scope: "global",
		Connector: "probe-conn", Acceptance: []string{"a1"}, Nodes: []Node{{Name: "n1", Unit: "session"}},
		RV: 3, EventID: "ev-d1", At: time.Now()})
	_ = store.Close() // 模拟重启（写穿透已同步落库）

	store2, err := runtimestore.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	defer store2.Close()
	d2 := NewDispatchLog(&DispatchOptions{Store: store2})
	st := d2.State()
	if st.Accepted != 1 || len(st.Records) != 1 {
		t.Fatalf("恢复应得 1 行（accepted 续接）：%+v", st)
	}
	r := st.Records[0]
	if r.Seq != 1 || r.Workflow != "probe-wf" || r.Connector != "probe-conn" || r.RV != 3 ||
		len(r.Acceptance) != 1 || len(r.Nodes) != 1 || r.Nodes[0].Name != "n1" {
		t.Fatalf("恢复行字段应无损：%+v", r)
	}
	// seq 续接：新派单 seq=2（不与恢复行撞号）。
	d2.Dispatch(DispatchRecord{Agent: DispatchAgent, Status: DispatchAccepted, Action: "updated",
		Kind: DispatchKindWorkflow, Workflow: "probe-wf", At: time.Now()})
	st = d2.State()
	if len(st.Records) != 2 || st.Records[1].Seq != 2 {
		t.Fatalf("seq 应续接（新行 seq=2）：%+v", st.Records)
	}
	// 幂等重放：同 Store 再重开（行数不翻倍）。
	_ = runtimestoreOpenClose(t, path)
	d3 := NewDispatchLog(&DispatchOptions{Store: store2})
	if st := d3.State(); st.Accepted != 2 || len(st.Records) != 2 {
		t.Fatalf("重开应幂等恢复 2 行：%+v", st)
	}
}

// runtimestoreOpenClose 开-关一轮（升列幂等重开路径锚）。
func runtimestoreOpenClose(t *testing.T, path string) error {
	t.Helper()
	s, err := runtimestore.Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	return s.Close()
}

// TestDispatchMemoryNoStore memory/nil Store=既有行为面零翻转（空启动零行）。
func TestDispatchMemoryNoStore(t *testing.T) {
	d := NewDispatchLog(nil)
	if st := d.State(); st.Accepted != 0 || len(st.Records) != 0 {
		t.Fatalf("nil Store 应空启动：%+v", st)
	}
	d.Dispatch(DispatchRecord{Agent: DispatchAgent, Action: "created", Kind: DispatchKindWorkflow})
	if st := d.State(); st.Accepted != 1 {
		t.Fatalf("纯内存受理应照常：%+v", st)
	}
}
