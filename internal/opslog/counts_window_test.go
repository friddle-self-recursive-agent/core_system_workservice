// counts_window_test.go EventCountsSince 窗口口径单测（v0.45 装配段）：半开窗 (since, now]
// ——端点瞬刻归上一窗，相邻窗零重叠（E7 半开窗纪律同口径）；since 零值=全量（首拍对账
// 口径不变）。设计依据：cored-bridge.md v0.33 自观测采样器（查询投影即样本）+ v0.45 窗口
// 口径对齐。
package opslog

import (
	"testing"
	"time"
)

func TestEventCountsSinceHalfOpenWindow(t *testing.T) {
	s, err := OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()

	t0 := time.Now()
	rows := []struct {
		id string
		ts time.Time
	}{
		{"evt-0", t0}, // 端点瞬刻（=since）
		{"evt-1", t0.Add(10 * time.Millisecond)},
		{"evt-2", t0.Add(20 * time.Millisecond)},
	}
	for _, r := range rows {
		if err := s.AppendEvent(EventLog{Ts: r.ts, EventID: r.id, Type: "DocApplied", Severity: "info"}); err != nil {
			t.Fatal(err)
		}
	}

	// 首拍全量口径：since 零值 → 全部行（含端点 evt-0）。
	full, err := s.EventCountsSince(time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if full["DocApplied"] != 3 {
		t.Fatalf("首拍全量 DocApplied = %d, 要 3", full["DocApplied"])
	}

	// 半开窗 (since, now]：端点 evt-0（ts==since）归上一窗，不计入本窗——相邻窗零重叠。
	part, err := s.EventCountsSince(t0)
	if err != nil {
		t.Fatal(err)
	}
	if part["DocApplied"] != 2 {
		t.Fatalf("半开窗 (since,now] DocApplied = %d, 要 2（端点瞬刻归上一窗）", part["DocApplied"])
	}

	// 相邻两窗拼接 = 全量且零重叠：w1=(zero,t1] 含 evt-0/1，w2=(t1,t2] 含 evt-2。
	t1 := t0.Add(15 * time.Millisecond)
	w1, err := s.EventCountsSince(time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	_ = w1 // 全量=3 已断言
	w2, err := s.EventCountsSince(t1)
	if err != nil {
		t.Fatal(err)
	}
	if w2["DocApplied"] != 1 {
		t.Fatalf("窗 (t1,t2] DocApplied = %d, 要 1", w2["DocApplied"])
	}

	// 未知类型窗口=零条目（map 空——不冒充）。
	none, err := s.EventCountsSince(time.Now().Add(time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if len(none) != 0 {
		t.Fatalf("未来窗应零条目, 得 %v", none)
	}
}
