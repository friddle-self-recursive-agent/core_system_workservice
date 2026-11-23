// events_read_test.go —— 事件归档查询面 HTTP 验收（v0.36 console 读面首段，cored-bridge.md §2
// `GET /v1/events`）：event_log 归档行的最新优先投影、五参过滤/limit 收敛、face 缺席 503、
// 与注入面 POST 同路径方法分立（非 GET → 405）。
package apiserver

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
)

// evArchiveServer 组装带事件归档查询口面的实例（ops DB=进程内存库；其余面 nil——本面零信号面依赖）。
func evArchiveServer(t *testing.T, store *opslog.SQLiteStore) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(New(testScheme(t), nil, nil, nil, nil, nil, nil, nil, nil, WithEventLog(store)))
	t.Cleanup(ts.Close)
	return ts
}

func evGet(t *testing.T, url string) (int, map[string]any) {
	t.Helper()
	resp, err := http.Get(url)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("载荷不可解析: %v", err)
	}
	return resp.StatusCode, out
}

func TestEventsArchiveEmptyReturnsZeroRows(t *testing.T) {
	store, err := opslog.OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ts := evArchiveServer(t, store)
	code, out := evGet(t, ts.URL+"/v1/events")
	if code != http.StatusOK {
		t.Fatalf("空归档应 200（关断口径可见），得 %d", code)
	}
	if out["count"].(float64) != 0 {
		t.Errorf("count=%v, 要 0", out["count"])
	}
	if rows, ok := out["events"].([]any); !ok || len(rows) != 0 {
		t.Errorf("events 应为空数组非缺席: %v", out["events"])
	}
}

func TestEventsArchiveOrderAndProjection(t *testing.T) {
	store, err := opslog.OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	base := time.Now().UTC().Add(-time.Hour)
	// 三行乱序落库：A(最旧) B(最新) C(中间)——读面要最新优先（ts 降序）。
	rows := []opslog.EventLog{
		{Ts: base, EventID: "evt-a", Type: "DocApplied", Source: "apply", Involved: "Rule/global/r-1", DedupKey: "apply:Rule/global/r-1:1", Severity: "info", RoutedTo: "owner-workflow"},
		{Ts: base.Add(2 * time.Minute), EventID: "evt-b", Type: "Escalate", Source: "human", Involved: "Agent/global/x", DedupKey: "esc:x", Severity: "error", RoutedTo: "root-agent"},
		{Ts: base.Add(time.Minute), EventID: "evt-c", Type: "DocApplied", Source: "apply", Involved: "Rule/global/r-2", DedupKey: "apply:Rule/global/r-2:2", Severity: "warn", RoutedTo: ""},
	}
	for _, r := range rows {
		if err := store.AppendEvent(r); err != nil {
			t.Fatal(err)
		}
	}
	ts := evArchiveServer(t, store)
	code, out := evGet(t, ts.URL+"/v1/events")
	if code != http.StatusOK {
		t.Fatalf("应 200，得 %d", code)
	}
	if out["count"].(float64) != 3 {
		t.Fatalf("count=%v, 要 3", out["count"])
	}
	arr := out["events"].([]any)
	got := make([]string, 0, len(arr))
	for _, it := range arr {
		row := it.(map[string]any)
		got = append(got, row["eventId"].(string))
		// 载荷键=投影契约（eventId/type/source/involved/dedupKey/level/routedTo/ts 全在）。
		for _, k := range []string{"type", "source", "involved", "dedupKey", "level", "routedTo", "ts"} {
			if _, ok := row[k]; !ok {
				t.Errorf("归档行缺键 %s: %v", k, row)
			}
		}
	}
	if got[0] != "evt-b" || got[1] != "evt-c" || got[2] != "evt-a" {
		t.Errorf("要最新优先（b,c,a），得 %v", got)
	}
}

func TestEventsArchiveFilters(t *testing.T) {
	store, err := opslog.OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	base := time.Now().UTC().Add(-time.Hour)
	for i := 0; i < 6; i++ {
		typ := "DocApplied"
		if i%2 == 1 {
			typ = "Escalate"
		}
		sev := "info"
		if typ == "Escalate" {
			sev = "error"
		}
		if err := store.AppendEvent(opslog.EventLog{
			Ts: base.Add(time.Duration(i) * time.Minute), EventID: fmt.Sprintf("evt-%d", i),
			Type: typ, Source: "apply", Involved: "Rule/global/r", Severity: sev,
		}); err != nil {
			t.Fatal(err)
		}
	}
	ts := evArchiveServer(t, store)
	// type 过滤。
	_, out := evGet(t, ts.URL+"/v1/events?type=Escalate")
	if out["count"].(float64) != 3 {
		t.Errorf("type 过滤 count=%v, 要 3", out["count"])
	}
	// level 过滤。
	_, out = evGet(t, ts.URL+"/v1/events?level=error")
	if out["count"].(float64) != 3 {
		t.Errorf("level 过滤 count=%v, 要 3", out["count"])
	}
	// involved 过滤。
	_, out = evGet(t, ts.URL+"/v1/events?involved=Rule/global/r")
	if out["count"].(float64) != 6 {
		t.Errorf("involved 过滤 count=%v, 要 6", out["count"])
	}
	// since 过滤（含端点：第 4 分钟起=3 行）。
	since := base.Add(3 * time.Minute).Format(time.RFC3339)
	_, out = evGet(t, ts.URL+"/v1/events?since="+since)
	if out["count"].(float64) != 3 {
		t.Errorf("since 过滤 count=%v, 要 3", out["count"])
	}
	// limit 收敛（limit=2 → 最新 2 行）。
	_, out = evGet(t, ts.URL+"/v1/events?limit=2")
	arr := out["events"].([]any)
	if len(arr) != 2 {
		t.Fatalf("limit 收敛 len=%d, 要 2", len(arr))
	}
	if arr[0].(map[string]any)["eventId"] != "evt-5" {
		t.Errorf("limit 取最新，得 %v", arr[0])
	}
	// limit>上限收敛到 500（读面防拖库——不放大行数）。
	_, out = evGet(t, ts.URL+"/v1/events?limit=99999")
	if out["count"].(float64) != 6 {
		t.Errorf("超大 limit 应仅返回现有 6 行，得 %v", out["count"])
	}
}

func TestEventsArchiveBadSinceIs400(t *testing.T) {
	store, err := opslog.OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ts := evArchiveServer(t, store)
	code, out := evGet(t, ts.URL+"/v1/events?since=not-a-time")
	if code != http.StatusBadRequest {
		t.Fatalf("since 不可解析应 400，得 %d", code)
	}
	if _, ok := out["error"]; !ok {
		t.Errorf("错误载荷缺 error: %v", out)
	}
}

func TestEventsArchiveFaceAbsent503(t *testing.T) {
	// face 未注入=503（面注册但能力缺席）；旧代（路由未注册）仍是 405 兜底=代际判别器。
	ts := httptest.NewServer(New(testScheme(t), nil, nil, nil, nil, nil, nil, nil, nil))
	defer ts.Close()
	code, out := evGet(t, ts.URL+"/v1/events")
	if code != http.StatusServiceUnavailable {
		t.Fatalf("face 缺席应 503，得 %d", code)
	}
	if e, _ := out["error"].(string); e == "" {
		t.Errorf("503 载荷缺 error 文本: %v", out)
	}
}

func TestEventsArchivePostStill405MethodSplit(t *testing.T) {
	// 同路径方法分立：POST 注入面照常 405（写面关断），DELETE 等非注册方法=405（mux 口径）。
	store, err := opslog.OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ts := evArchiveServer(t, store)
	code, _ := postEvents(t, ts, authHeader("probe", "global/admin"), `{"type":"Escalate"}`)
	if code != http.StatusMethodNotAllowed {
		t.Fatalf("注入面关断 POST 应 405，得 %d", code)
	}
	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/v1/events", nil)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("DELETE /v1/events 应 405，得 %d", resp.StatusCode)
	}
}
