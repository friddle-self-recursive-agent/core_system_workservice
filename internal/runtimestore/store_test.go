// store_test.go —— 运行态投影存储验收（R-ENG §1）：sqlite 往返（写入→重开→恢复→序号
// 续接）、五类流水幂等重放、会话 upsert 演化、memory 幂等、并发 Append 安全（-race）、
// 注册表 Open（内置驱动名/未知驱动词形）。
package runtimestore

import (
	"path/filepath"
	"sync"
	"testing"
)

// TestSQLiteRoundTripRestoreSeq sqlite 往返：写入→关闭→重开→列举恢复→同 id 重放 noop→
// 新 id 追加（序号续接判据）。
func TestSQLiteRoundTripRestoreSeq(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime.db")
	s, err := Open("sqlite", path)
	if err != nil {
		t.Fatalf("打开 sqlite: %v", err)
	}
	mustAppendAll(t, s, 1, 2, 3)
	if err := s.Close(); err != nil {
		t.Fatalf("关闭: %v", err)
	}

	// 重开（同一路径——权威续接口径）。
	s2, err := Open("sqlite", path)
	if err != nil {
		t.Fatalf("重开 sqlite: %v", err)
	}
	defer s2.Close()
	trs, err := s2.ListTriggers(0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if len(trs) != 3 || trs[0].Seq != 1 || trs[2].Seq != 3 {
		t.Fatalf("恢复应得 3 行升序（seq 1..3），得 %v", trs)
	}
	if trs[0].EventID != "evt-1" || trs[2].Type != "DocApplied" {
		t.Fatalf("行字段应往返无损：%v", trs)
	}
	runs, err := s2.ListRuns(0, 0)
	if err != nil || len(runs) != 3 {
		t.Fatalf("run 行恢复应 3 行：%v %v", runs, err)
	}
	if runs[0].PhasesJSON == "" {
		t.Fatalf("JSON 列应往返无损：%v", runs[0])
	}

	// 幂等重放：同 id 再追加 → 行数不变。
	mustAppendAll(t, s2, 1, 2, 3)
	trs, _ = s2.ListTriggers(0, 0)
	if len(trs) != 3 {
		t.Fatalf("同 id 重放应 noop（仍 3 行），得 %d", len(trs))
	}

	// 序号续接：新 seq 4 追加成功且游标分页可用。
	mustAppendAll(t, s2, 4)
	trs, _ = s2.ListTriggers(0, 0)
	if len(trs) != 4 || trs[3].Seq != 4 {
		t.Fatalf("seq 4 应追加成功，得 %v", trs)
	}
	page, err := s2.ListTriggers(2, trs[1].Cursor)
	if err != nil || len(page) != 2 || page[0].Seq != 3 {
		t.Fatalf("游标分页应从 seq 3 起两行：%v %v", page, err)
	}
}

// TestMemoryIdempotentAndUpsert memory 驱动：五类流水幂等 + 会话 upsert 演化（恢复取
// 最新词形）。
func TestMemoryIdempotentAndUpsert(t *testing.T) {
	s, err := Open("memory", "")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	mustAppendAll(t, s, 1, 2)
	mustAppendAll(t, s, 2) // 重放
	trs, _ := s.ListTriggers(0, 0)
	if len(trs) != 2 {
		t.Fatalf("memory 幂等重放应 noop（2 行），得 %d", len(trs))
	}
	// 会话 upsert：同 id 两轮推进 → 一行最新词形（id 取独立锚——不与 mustAppendAll 撞行）。
	sess := IntakeSessionRow{SessionID: "sess-upsert", Kind: "Proposal", Turns: 1, FieldsJSON: "{}", SourcesJSON: "{}"}
	if err := s.AppendIntakeSession(sess); err != nil {
		t.Fatal(err)
	}
	sess.Turns = 3
	sess.Ready = true
	if err := s.AppendIntakeSession(sess); err != nil {
		t.Fatal(err)
	}
	rows, err := s.ListIntakeSessions(0, 0)
	if err != nil || len(rows) != 3 { // sess-1/sess-2（mustAppendAll）+ sess-upsert 单行
		t.Fatalf("会话应 upsert 不增行：%v %v", rows, err)
	}
	// 找出 upsert 锚行验证最新词形（游标=首轮插入序）。
	var up *IntakeSessionRow
	for i := range rows {
		if rows[i].SessionID == "sess-upsert" {
			up = &rows[i]
		}
	}
	if up == nil || up.Turns != 3 || !up.Ready {
		t.Fatalf("恢复应取最新词形：%v", rows)
	}
	// 游标分页（插入序）——从 sess-1 游标之后翻一页。
	var s1 *IntakeSessionRow
	for i := range rows {
		if rows[i].SessionID == "sess-1" {
			s1 = &rows[i]
		}
	}
	page, _ := s.ListIntakeSessions(1, s1.Cursor)
	if len(page) != 1 || page[0].SessionID != "sess-2" {
		t.Fatalf("会话游标分页应得 sess-2：%v", page)
	}
}

// TestConcurrentAppendSafety 并发 Append 安全（-race 下无数据竞争；全部行在册）。
func TestConcurrentAppendSafety(t *testing.T) {
	for _, driver := range []string{"memory", "sqlite"} {
		var s Store
		var err error
		if driver == "sqlite" {
			s, err = Open("sqlite", filepath.Join(t.TempDir(), "rt.db"))
		} else {
			s, err = Open("memory", "")
		}
		if err != nil {
			t.Fatal(err)
		}
		const n = 32
		var wg sync.WaitGroup
		for i := 1; i <= n; i++ {
			wg.Add(1)
			go func(seq int64) {
				defer wg.Done()
				_ = s.AppendTrigger(TriggerRow{Seq: seq, EventID: "evt", Type: "DocApplied"})
				_ = s.AppendWake(WakeRow{Wake: seq})
			}(int64(i))
		}
		wg.Wait()
		trs, err := s.ListTriggers(0, 0)
		if err != nil || len(trs) != n {
			t.Fatalf("%s: 并发追加应全量在册 %d 行：%d %v", driver, n, len(trs), err)
		}
		wks, err := s.ListWakes(0, 0)
		if err != nil || len(wks) != n {
			t.Fatalf("%s: 并发唤醒追加应全量在册：%d %v", driver, len(wks), err)
		}
		_ = s.Close()
	}
}

// TestRegistryOpen 注册表：内置驱动名可开；未知驱动=可见错误；sqlite 空路径=进程内存库；
// 重复 Register 幂等（同名覆盖）。
func TestRegistryOpen(t *testing.T) {
	s, err := Open("sqlite", "") // 空 DSN=进程内存库
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AppendTrigger(TriggerRow{Seq: 1, Type: "DocApplied"}); err != nil {
		t.Fatal(err)
	}
	rows, _ := s.ListTriggers(0, 0)
	if len(rows) != 1 {
		t.Fatalf("进程内存 sqlite 库应可用：%v", rows)
	}
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := Open("bogus", ""); err == nil {
		t.Fatal("未知驱动应可见错误")
	}
	Register("memory", openMemory) // 幂等重注册（同名覆盖=合法）
	if _, err := Open("memory", "ignored-dsn"); err != nil {
		t.Fatal(err)
	}
	names := Drivers()
	if len(names) < 2 {
		t.Fatalf("内置驱动应至少 sqlite|memory 两名：%v", names)
	}
}

// TestSQLiteSessionAndVerifyColumns G-08/G-09 列往返：run 行 session_id/verify_json 与
// wake 行 session_id 写入→重开→恢复无损（升列幂等——重复打开不炸）；唤醒会话锚回填
// （AppendWakeSession 就地覆写，受理词形 IGNORE 幂等不覆盖回填）。
func TestSQLiteSessionAndVerifyColumns(t *testing.T) {
	path := filepath.Join(t.TempDir(), "runtime.db")
	s, err := Open("sqlite", path)
	if err != nil {
		t.Fatal(err)
	}
	if err := s.AppendRun(RunRow{Run: 1, Workflow: "core-workflow", SessionID: "loop-sess-1",
		VerifyJSON: `{"ok":false,"auditId":"audit-x","errorFindings":2,"note":"R-PUB-1 违规"}`}); err != nil {
		t.Fatal(err)
	}
	if err := s.AppendWake(WakeRow{Wake: 1, Agent: "root", SessionID: "loop-sess-1"}); err != nil {
		t.Fatal(err)
	}
	// 幂等重开（升列 duplicate column 容忍路径）。
	if err := s.Close(); err != nil {
		t.Fatal(err)
	}
	s2, err := Open("sqlite", path)
	if err != nil {
		t.Fatalf("升列后重开应无损: %v", err)
	}
	defer s2.Close()
	runs, err := s2.ListRuns(0, 0)
	if err != nil || len(runs) != 1 {
		t.Fatalf("run 行恢复：%v %v", runs, err)
	}
	if runs[0].SessionID != "loop-sess-1" || runs[0].VerifyJSON == "" {
		t.Fatalf("run session/verify 列应往返无损：%v", runs[0])
	}
	// 唤醒会话锚回填：受理词形（无 session）先落，回填后就地覆写。
	if err := s2.AppendWake(WakeRow{Wake: 2, Agent: "root"}); err != nil {
		t.Fatal(err)
	}
	if err := s2.AppendWakeSession(2, "loop-sess-2"); err != nil {
		t.Fatal(err)
	}
	if err := s2.AppendWake(WakeRow{Wake: 2, Agent: "root"}); err != nil { // 同 wake 重放=IGNORE noop，不得覆盖回填
		t.Fatal(err)
	}
	wks, err := s2.ListWakes(0, 0)
	if err != nil || len(wks) != 2 {
		t.Fatalf("唤醒行恢复：%v %v", wks, err)
	}
	if wks[0].SessionID != "loop-sess-1" || wks[1].SessionID != "loop-sess-2" {
		t.Fatalf("唤醒 session_id 应往返/回填无损：%v", wks)
	}
	// memory 驱动同语义（AppendWakeSession 回填 + AppendWake 重放不覆盖）。
	ms, err := Open("memory", "")
	if err != nil {
		t.Fatal(err)
	}
	defer ms.Close()
	_ = ms.AppendWake(WakeRow{Wake: 7, Agent: "root"})
	_ = ms.AppendWakeSession(7, "sess-mem")
	_ = ms.AppendWake(WakeRow{Wake: 7, Agent: "root"})
	mw, _ := ms.ListWakes(0, 0)
	if len(mw) != 1 || mw[0].SessionID != "sess-mem" {
		t.Fatalf("memory 回填应就地覆写且重放不覆盖：%v", mw)
	}
}

// mustAppendAll 六类行各追加一条（seq n——五类流水同号不同 kind；会话 id=sess-<n>）。
func mustAppendAll(t *testing.T, s Store, seqs ...int64) {
	t.Helper()
	for _, n := range seqs {
		if err := s.AppendTrigger(TriggerRow{Seq: n, EventID: "evt-" + itoa(n), Type: "DocApplied",
			DedupKey: "apply:x:1", Source: "apply", Involved: "Agent/global/x", Level: "info",
			At: "2026-11-26T00:00:00Z", ReceivedAt: "2026-11-26T00:00:00.000000001Z",
			ExecJSON: `{"executor":"workflow","target":"probe","wired":true}`}); err != nil {
			t.Fatal(err)
		}
		if err := s.AppendRun(RunRow{Run: n, Workflow: "core-workflow", SpecSource: "builtin",
			PhasesJSON: `["Idle","Analyzing"]`, TransitionsJSON: `[]`, Phase: "Analyzing", Status: "running",
			BrakesJSON: `{"cooldown":"0s"}`, TriggerSeq: n, EventID: "evt-" + itoa(n),
			StartedAt: "2026-11-26T00:00:00Z"}); err != nil {
			t.Fatal(err)
		}
		if err := s.AppendDelivery(DeliveryRow{Delivery: n, Channel: "email", Origin: "route",
			Status: "accepted", At: "2026-11-26T00:00:00Z"}); err != nil {
			t.Fatal(err)
		}
		if err := s.AppendWake(WakeRow{Wake: n, Agent: "root", Origin: "route", Status: "accepted",
			At: "2026-11-26T00:00:00Z"}); err != nil {
			t.Fatal(err)
		}
		if err := s.AppendDispatch(DispatchRow{Seq: n, Agent: "apply-workflow", Status: "accepted",
			AutoLevel: "ask", Action: "created", Kind: "workflow",
			AcceptanceJSON: `[]`, NodesJSON: `[]`, At: "2026-11-26T00:00:00Z"}); err != nil {
			t.Fatal(err)
		}
		if err := s.AppendIntakeSession(IntakeSessionRow{SessionID: "sess-" + itoa(n), Kind: "Proposal",
			Turns: int(n), FieldsJSON: "{}", SourcesJSON: "{}", CreatedAt: "2026-11-26T00:00:00Z",
			UpdatedAt: "2026-11-26T00:00:00Z"}); err != nil {
			t.Fatal(err)
		}
	}
}

// itoa 极小整数转串（测试锚固定词形用）。
func itoa(n int64) string {
	if n == 0 {
		return "0"
	}
	var b []byte
	for n > 0 {
		b = append([]byte{byte('0' + n%10)}, b...)
		n /= 10
	}
	return string(b)
}
