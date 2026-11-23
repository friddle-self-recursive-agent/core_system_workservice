// history_read_test.go —— 对象应用史投影面 HTTP 验收（v0.38 console 读面第三段，cored-bridge.md §2
// `GET /v1/history/{kind}/{name}?scope=`）：apply_log 行的 rv 降序投影、obj 精确过滤、limit 收敛、
// face 缺席 503、方法分立（非 GET → 405 兜底）、真实 apply 流水端到端（引擎落账→读面可查）。
package apiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
)

// histServer 组装带应用史查询口面的实例（ops DB=进程内存库；其余面 nil——本面零信号面依赖）。
func histServer(t *testing.T, store *opslog.SQLiteStore) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(New(testScheme(t), nil, nil, nil, nil, nil, nil, nil, nil, WithApplyLog(store)))
	t.Cleanup(ts.Close)
	return ts
}

func histGet(t *testing.T, url string) (int, map[string]any) {
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

func seedApplies(t *testing.T, store *opslog.SQLiteStore) {
	t.Helper()
	base := time.Now().UTC().Add(-time.Hour)
	rows := []opslog.ApplyLog{
		{Ts: base, FieldManager: "agent:probe", Subject: "u=probe", Obj: "Rule/global/r-1", ResourceVersion: 1, Result: "ok", ReasonRef: "bootstrap"},
		{Ts: base.Add(time.Minute), FieldManager: "console:editor", Subject: "u=console-operator", Obj: "Rule/global/r-1", ResourceVersion: 2, Result: "ok", ReasonRef: "console:edit"},
		{Ts: base.Add(2 * time.Minute), FieldManager: "agent:probe", Subject: "u=probe", Obj: "Rule/global/r-2", ResourceVersion: 1, Result: "ok", ReasonRef: "bootstrap"},
		{Ts: base.Add(3 * time.Minute), FieldManager: "console:editor", Subject: "u=console-operator", Obj: "Rule/global/r-1", ResourceVersion: 2, Result: "conflict", ReasonRef: "console:edit"},
		{Ts: base.Add(4 * time.Minute), FieldManager: "agent:probe", Subject: "anonymous", Obj: "Rule/global/r-1", ResourceVersion: 0, Result: "denied", ReasonRef: ""},
	}
	for _, r := range rows {
		if err := store.AppendApply(r); err != nil {
			t.Fatal(err)
		}
	}
}

func TestHistoryEmptyReturnsZeroRows(t *testing.T) {
	store, err := opslog.OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ts := histServer(t, store)
	code, out := histGet(t, ts.URL+"/v1/history/Rule/r-1?scope=global")
	if code != http.StatusOK {
		t.Fatalf("空账应 200（查询面职责=有界窗非全量承诺），得 %d", code)
	}
	if out["count"].(float64) != 0 {
		t.Errorf("count=%v, 要 0", out["count"])
	}
	if rows, ok := out["entries"].([]any); !ok || len(rows) != 0 {
		t.Errorf("entries 应为空数组非缺席: %v", out["entries"])
	}
	if out["obj"] != "Rule/global/r-1" {
		t.Errorf("obj 投影=%v，要 Rule/global/r-1", out["obj"])
	}
}

func TestHistoryRVDescAndProjection(t *testing.T) {
	store, err := opslog.OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	seedApplies(t, store)
	ts := histServer(t, store)
	code, out := histGet(t, ts.URL+"/v1/history/Rule/r-1?scope=global")
	if code != http.StatusOK {
		t.Fatalf("应 200，得 %d", code)
	}
	if out["count"].(float64) != 4 {
		t.Fatalf("r-1 应恰好 4 行（rv1 ok+rv2 ok+rv2 conflict+rv0 denied），count=%v", out["count"])
	}
	rows := out["entries"].([]any)
	// rv 降序：conflict 行（rv=2 现库版本）与 ok 行（rv=2）同 rv 按时刻降序——conflict 更晚在前。
	first := rows[0].(map[string]any)
	if first["result"] != "conflict" || first["rv"].(float64) != 2 {
		t.Errorf("首行应 conflict rv=2（同 rv 时刻降序），得 %v/%v", first["result"], first["rv"])
	}
	second := rows[1].(map[string]any)
	if second["result"] != "ok" || second["fieldManager"] != "console:editor" {
		t.Errorf("次行应 ok console:editor，得 %v/%v", second["result"], second["fieldManager"])
	}
	third := rows[2].(map[string]any)
	if third["rv"].(float64) != 1 || third["reasonRef"] != "bootstrap" {
		t.Errorf("三行应 rv=1 bootstrap，得 %v/%v", third["rv"], third["reasonRef"])
	}
	fourth := rows[3].(map[string]any)
	if fourth["result"] != "denied" || fourth["subject"] != "anonymous" {
		t.Errorf("末行应 denied（rv=0 审计行照投影；subject=anonymous 主体可追溯），得 %v/%v", fourth["result"], fourth["subject"])
	}
}

func TestHistoryObjFilterAndLimitCap(t *testing.T) {
	store, err := opslog.OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	seedApplies(t, store)
	ts := histServer(t, store)
	// limit 收敛：r-1 有 4 行，limit=2 → 2。
	code, out := histGet(t, ts.URL+"/v1/history/Rule/r-1?scope=global&limit=2")
	if code != http.StatusOK {
		t.Fatalf("应 200，得 %d", code)
	}
	if out["count"].(float64) != 2 {
		t.Errorf("limit=2 应截到 2 行，得 %v", out["count"])
	}
	// scope 缺席=对象键另一半（kind/scope/name 全匹配语义）——空 scope 只匹配空 scope 行。
	code, out = histGet(t, ts.URL+"/v1/history/Rule/r-1")
	if code != http.StatusOK || out["count"].(float64) != 0 {
		t.Errorf("缺 scope 应按空 scope 匹配=0 行，得 %d/%v", code, out["count"])
	}
}

func TestHistoryFaceAbsentReturns503(t *testing.T) {
	ts := httptest.NewServer(New(testScheme(t), nil, nil, nil, nil, nil, nil, nil, nil))
	defer ts.Close()
	code, out := histGet(t, ts.URL+"/v1/history/Rule/r-1?scope=global")
	if code != http.StatusServiceUnavailable {
		t.Fatalf("face 缺席应 503（面注册能力缺席防御面），得 %d", code)
	}
	if out["error"] == "" {
		t.Errorf("503 载荷应带 error 诊断: %v", out)
	}
}

func TestHistoryMethodSplit(t *testing.T) {
	store, err := opslog.OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	ts := histServer(t, store)
	resp, err := http.Post(ts.URL+"/v1/history/Rule/r-1?scope=global", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("POST 历史面应 405（只读面方法分立），得 %d", resp.StatusCode)
	}
}
