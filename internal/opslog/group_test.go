// group_test.go —— event_log 分组计数（v0.37 console 读面第二段，`GET /v1/workflows` 计数源
// 的底层读取）：过滤语义同 QueryEvents、分组键=(type,involved)、排序确定性、扫描有界 capped。
package opslog

import (
	"testing"
	"time"
)

func TestGroupEventsFiltersAndCounts(t *testing.T) {
	store, err := OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	base := time.Now().UTC().Add(-time.Hour)
	rows := []EventLog{
		{Ts: base, EventID: "a", Type: "DocApplied", Source: "apply", Involved: "Rule/s1/x", Severity: "info"},
		{Ts: base.Add(time.Minute), EventID: "b", Type: "DocApplied", Source: "apply", Involved: "Rule/s1/x", Severity: "info"},
		{Ts: base.Add(2 * time.Minute), EventID: "c", Type: "DocApplied", Source: "apply", Involved: "Rule/s1/y", Severity: "warn"},
		{Ts: base.Add(-30 * time.Hour), EventID: "d", Type: "Escalate", Source: "human", Involved: "Agent/global/root", Severity: "error"},
	}
	for _, r := range rows {
		if err := store.AppendEvent(r); err != nil {
			t.Fatal(err)
		}
	}
	groups, scanned, capped, err := store.GroupEvents(EventQuery{})
	if err != nil {
		t.Fatal(err)
	}
	if capped || scanned != int64(len(rows)) {
		t.Fatalf("扫描口径不符: scanned=%d capped=%v", scanned, capped)
	}
	if len(groups) != 3 {
		t.Fatalf("分组数=%d，要 3", len(groups))
	}
	if groups[0].Type != "DocApplied" || groups[0].Involved != "Rule/s1/x" || groups[0].Count != 2 {
		t.Errorf("计数排序首位不符: %+v", groups[0])
	}
	// 窗口过滤（24h 外的 Escalate 不入窗组）。
	g24, _, _, err := store.GroupEvents(EventQuery{Since: base.Add(-time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	total := 0
	for _, g := range g24 {
		if g.Type == "Escalate" {
			t.Errorf("窗口外类型不应入窗组: %+v", g)
		}
		total += g.Count
	}
	if total != 3 {
		t.Errorf("窗组总数=%d，要 3", total)
	}
	// 精确过滤（type+involved）。
	one, _, _, err := store.GroupEvents(EventQuery{Type: "DocApplied", Involved: "Rule/s1/y"})
	if err != nil {
		t.Fatal(err)
	}
	if len(one) != 1 || one[0].Count != 1 {
		t.Errorf("精确过滤不符: %+v", one)
	}
}
