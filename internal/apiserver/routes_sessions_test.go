// routes_sessions_test.go —— 会话上下文契约面（G-14）HTTP 验收：
//   - GET /v1/sessions/{id}/context：face 缺席 503 防御面；在册会话行投影（state 词形）+
//     pack 缺省词形（takeBudget=30000/degraded=true/tagSigSource=none）+ shards/stats 投影 +
//     writeGate 词形；不在册会话 session=null（200 不冒充）；
//   - PUT：写面门控（未开启=405 与 heartbeat 同口径）/takeBudget 越界 400/tagSig 超限 400
//     （总 >20、每 type >5、空名坏键）/保序去重/乐观锁 409/受理 200 revision 单调；
//   - POST revert：写前快照恢复（含恢复「无覆盖」基线）+ 栈空 404 + 有界栈（≤10 层）。
package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/contextidx"
)

// sessCtxServer 会话上下文面测试服（真 contextidx 索引 + stubLive 活性账 + 写面可选开启）。
func sessCtxServer(t *testing.T, writeOpen bool) *httptest.Server {
	t.Helper()
	ix := contextidx.New(ctxFixture(), func() []string { return []string{"Rule"} })
	if err := ix.Build(context.Background()); err != nil {
		t.Fatalf("索引构建失败: %v", err)
	}
	live := &stubLive{}
	var applier Applier
	if writeOpen {
		applier = &applierFaceStub{}
	}
	ts := httptest.NewServer(New(testScheme(t), nil, applier, nil, nil, nil, nil, nil, nil,
		WithContextIdx(ix), WithSessionLiveness(live)))
	t.Cleanup(ts.Close)
	return ts
}

// sessCtxDo 会话上下文面请求（method+body → 状态码+JSON 载荷）。
func sessCtxDo(t *testing.T, ts *httptest.Server, method, path, body string) (int, map[string]any) {
	t.Helper()
	var req *http.Request
	var err error
	if body == "" {
		req, err = http.NewRequest(method, ts.URL+path, nil)
		if err != nil {
			t.Fatal(err)
		}
	} else {
		req, err = http.NewRequest(method, ts.URL+path, strings.NewReader(body))
		if err != nil {
			t.Fatal(err)
		}
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := http.DefaultClient.Do(req)
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

func TestSessionCtxFaceAbsent503(t *testing.T) {
	// contextIdx 未装配：GET=503（与 /v1/context/* 同防御面——不冒充空上下文）。
	ts := httptest.NewServer(New(testScheme(t), nil, nil, nil, nil, nil, nil, nil, nil))
	defer ts.Close()
	code, out := sessCtxDo(t, ts, http.MethodGet, "/v1/sessions/s1/context", "")
	if code != http.StatusServiceUnavailable {
		t.Fatalf("face 缺席=%d（期望 503）", code)
	}
	if s, _ := out["error"].(string); !strings.Contains(s, "contextidx") {
		t.Fatalf("错误词形缺 contextidx 注记: %v", out["error"])
	}
}

func TestSessionCtxReadWordShape(t *testing.T) {
	ts := sessCtxServer(t, true)
	code, out := sessCtxDo(t, ts, http.MethodGet, "/v1/sessions/s1/context", "")
	if code != http.StatusOK {
		t.Fatalf("在册会话=%d（期望 200）", code)
	}
	// 会话行投影（stubLive s1=suspected-stall——state 词形不冒充）。
	sess, _ := out["session"].(map[string]any)
	if sess == nil || sess["state"] != "suspected-stall" || sess["sessionId"] != "s1" {
		t.Fatalf("会话行投影不符: %v", out["session"])
	}
	// pack 缺省词形（无覆盖=缺省预算+无 tagSig+degraded 诚实标注）。
	pack, _ := out["pack"].(map[string]any)
	if pack == nil {
		t.Fatal("pack 缺席")
	}
	if pack["takeBudget"].(float64) != float64(DefaultTakeBudget) || pack["takeBudgetDefault"].(float64) != float64(DefaultTakeBudget) {
		t.Fatalf("takeBudget 缺省词形不符: %v", pack["takeBudget"])
	}
	if pack["tagSigSource"] != "none" || pack["degraded"] != true || pack["tokens"] != nil {
		t.Fatalf("pack 缺省词形不符: tagSigSource=%v degraded=%v tokens=%v", pack["tagSigSource"], pack["degraded"], pack["tokens"])
	}
	// shards/stats：真 contextidx 断言成功（空 tagSig=空分片集；stats 带阈值词形）。
	if shards, _ := out["shards"].([]any); shards == nil || len(shards) != 0 {
		t.Fatalf("shards 缺省应为空集: %v", out["shards"])
	}
	stats, _ := out["shardsStats"].(map[string]any)
	if stats == nil || stats["rebuildThreshold"].(float64) != float64(contextidx.DefaultRebuildThreshold) {
		t.Fatalf("shardsStats 词形不符: %v", out["shardsStats"])
	}
	// writeGate/canRevert/override 缺省词形。
	wg, _ := out["writeGate"].(map[string]any)
	if wg == nil || wg["open"] != true {
		t.Fatalf("writeGate 词形不符: %v", out["writeGate"])
	}
	if out["canRevert"] != false || out["override"] != nil {
		t.Fatalf("覆盖缺省词形不符: canRevert=%v override=%v", out["canRevert"], out["override"])
	}
	// 不在册会话：session=null（200——契约面不因活性账缺席而 404，不冒充在册）。
	code2, out2 := sessCtxDo(t, ts, http.MethodGet, "/v1/sessions/ghost/context", "")
	if code2 != http.StatusOK || out2["session"] != nil {
		t.Fatalf("不在册会话=%d session=%v（期望 200+null）", code2, out2["session"])
	}
}

func TestSessionCtxPutGateAndValidation(t *testing.T) {
	// 写面关断：PUT/revert=405（与 POST /v1/sessions/heartbeat 同口径）。
	tsClosed := sessCtxServer(t, false)
	if code, _ := sessCtxDo(t, tsClosed, http.MethodPut, "/v1/sessions/s1/context", `{"takeBudget":40000}`); code != http.StatusMethodNotAllowed {
		t.Fatalf("关断面 PUT=%d（期望 405）", code)
	}
	if code, _ := sessCtxDo(t, tsClosed, http.MethodPost, "/v1/sessions/s1/context/revert", ""); code != http.StatusMethodNotAllowed {
		t.Fatalf("关断面 revert=%d（期望 405）", code)
	}
	ts := sessCtxServer(t, true)
	// takeBudget 越界=400（ContextPolicy 预算域外不受理）。
	if code, _ := sessCtxDo(t, ts, http.MethodPut, "/v1/sessions/s1/context", `{"takeBudget":100}`); code != http.StatusBadRequest {
		t.Fatalf("预算下越界=%d（期望 400）", code)
	}
	if code, _ := sessCtxDo(t, ts, http.MethodPut, "/v1/sessions/s1/context", `{"takeBudget":999999}`); code != http.StatusBadRequest {
		t.Fatalf("预算上越界=%d（期望 400）", code)
	}
	// tagSig 总量超限=400（>20）。
	many := make([]string, 21)
	for i := range many {
		many[i] = fmt.Sprintf("domain:t%d", i)
	}
	payload, _ := json.Marshal(map[string]any{"takeBudget": 30000, "tagSig": many})
	if code, _ := sessCtxDo(t, ts, http.MethodPut, "/v1/sessions/s1/context", string(payload)); code != http.StatusBadRequest {
		t.Fatalf("tag 总超限=%d（期望 400）", code)
	}
	// 每 tagType 超限=400（>5）。
	sameType := make([]string, 6)
	for i := range sameType {
		sameType[i] = fmt.Sprintf("domain:t%d", i)
	}
	payload, _ = json.Marshal(map[string]any{"takeBudget": 30000, "tagSig": sameType})
	if code, _ := sessCtxDo(t, ts, http.MethodPut, "/v1/sessions/s1/context", string(payload)); code != http.StatusBadRequest {
		t.Fatalf("tagType 超限=%d（期望 400）", code)
	}
	// 空名坏键=400（"domain:"——tag 侧空名不受理）。
	if code, _ := sessCtxDo(t, ts, http.MethodPut, "/v1/sessions/s1/context", `{"takeBudget":30000,"tagSig":["domain:"]}`); code != http.StatusBadRequest {
		t.Fatalf("空名坏键=%d（期望 400）", code)
	}
	// 受理：保序去重+revision=1。
	code, out := sessCtxDo(t, ts, http.MethodPut, "/v1/sessions/s1/context",
		`{"takeBudget":40000,"tagSig":["domain:notify","domain:notify","mech:cache","cache"],"note":"分叉试笔","by":"console:web","baseRevision":0}`)
	if code != http.StatusOK {
		t.Fatalf("受理=%d（期望 200）: %v", code, out)
	}
	ov, _ := out["override"].(map[string]any)
	if ov == nil || ov["revision"].(float64) != 1 || ov["takeBudget"].(float64) != 40000 {
		t.Fatalf("覆盖词形不符: %v", ov)
	}
	if tags, _ := ov["tagSig"].([]any); len(tags) != 3 || tags[0] != "domain:notify" || tags[2] != "cache" {
		t.Fatalf("tagSig 保序去重不符: %v", ov["tagSig"])
	}
	// 读面生效：takeBudget=override 值+tagSigSource=override+chunks 索引候选+shards 逐键投影。
	_, got := sessCtxDo(t, ts, http.MethodGet, "/v1/sessions/s1/context", "")
	pack := got["pack"].(map[string]any)
	if pack["takeBudget"].(float64) != 40000 || pack["tagSigSource"] != "override" {
		t.Fatalf("覆盖未生效: %v", pack)
	}
	if chunks, _ := pack["chunks"].([]any); len(chunks) == 0 {
		t.Fatal("tagSig 生效后 chunks 应有索引候选（fixture 含 notify 打标对象）")
	}
	if shards, _ := got["shards"].([]any); len(shards) != 3 {
		t.Fatalf("shards 应逐 tagKey 投影 3 行: %v", got["shards"])
	}
	if got["override"] == nil || got["canRevert"] != true {
		t.Fatalf("覆盖/可撤销词形不符: %v %v", got["override"], got["canRevert"])
	}
	// 乐观锁：过期 baseRevision=409（不静默重基——editor Apply 同语义）。
	if code, _ := sessCtxDo(t, ts, http.MethodPut, "/v1/sessions/s1/context",
		`{"takeBudget":45000,"baseRevision":99}`); code != http.StatusConflict {
		t.Fatalf("乐观锁冲突=%d（期望 409）", code)
	}
	// 对齐 revision 再写=200，revision 单调推进。
	if code, out := sessCtxDo(t, ts, http.MethodPut, "/v1/sessions/s1/context",
		`{"takeBudget":45000,"baseRevision":1}`); code != http.StatusOK || out["override"].(map[string]any)["revision"].(float64) != 2 {
		t.Fatalf("对齐重写=%d %v（期望 200+revision=2）", code, out)
	}
}

func TestSessionCtxRevertRestoresSnapshot(t *testing.T) {
	ts := sessCtxServer(t, true)
	// 第一次写（快照=无覆盖基线）→ 第二次写（快照=第一次覆盖）。
	if code, _ := sessCtxDo(t, ts, http.MethodPut, "/v1/sessions/s1/context", `{"takeBudget":40000,"tagSig":["domain:notify"]}`); code != http.StatusOK {
		t.Fatal("第一次写失败")
	}
	if code, _ := sessCtxDo(t, ts, http.MethodPut, "/v1/sessions/s1/context", `{"takeBudget":50000,"tagSig":["mech:cache"]}`); code != http.StatusOK {
		t.Fatal("第二次写失败")
	}
	// revert：恢复第一次覆盖（revision 仍单调推进=3）。
	code, out := sessCtxDo(t, ts, http.MethodPost, "/v1/sessions/s1/context/revert", "")
	if code != http.StatusOK {
		t.Fatalf("revert=%d（期望 200）", code)
	}
	ov := out["override"].(map[string]any)
	if ov["takeBudget"].(float64) != 40000 || ov["revision"].(float64) != 3 {
		t.Fatalf("快照恢复词形不符: %v", ov)
	}
	// 再 revert：恢复「无覆盖」基线（override=null，read 面回落缺省预算）。
	code, out = sessCtxDo(t, ts, http.MethodPost, "/v1/sessions/s1/context/revert", "")
	if code != http.StatusOK || out["override"] != nil {
		t.Fatalf("基线恢复=%d override=%v（期望 200+null）", code, out["override"])
	}
	_, got := sessCtxDo(t, ts, http.MethodGet, "/v1/sessions/s1/context", "")
	if got["pack"].(map[string]any)["takeBudget"].(float64) != float64(DefaultTakeBudget) {
		t.Fatal("基线恢复后预算应回落缺省")
	}
	// 栈空：第三次 revert=404（无快照不动作——不冒充撤销成功）。
	if code, _ = sessCtxDo(t, ts, http.MethodPost, "/v1/sessions/s1/context/revert", ""); code != http.StatusNotFound {
		t.Fatalf("栈空 revert=%d（期望 404）", code)
	}
}

func TestSessionCtxUndoDepthBounded(t *testing.T) {
	ts := sessCtxServer(t, true)
	// 连写 12 次（撤销栈有界 ≤10）。
	for i := 0; i < 12; i++ {
		body := `{"takeBudget":` + fmt.Sprintf("%d", MinTakeBudget+i) + `}`
		if code, _ := sessCtxDo(t, ts, http.MethodPut, "/v1/sessions/s1/context", body); code != http.StatusOK {
			t.Fatalf("第 %d 次写失败", i+1)
		}
	}
	// 连撤 10 次皆 200；第 11 次=404（最旧快照已被有界栈丢弃）。
	for i := 0; i < CtxRevertDepth; i++ {
		if code, _ := sessCtxDo(t, ts, http.MethodPost, "/v1/sessions/s1/context/revert", ""); code != http.StatusOK {
			t.Fatalf("第 %d 次撤销失败", i+1)
		}
	}
	if code, _ := sessCtxDo(t, ts, http.MethodPost, "/v1/sessions/s1/context/revert", ""); code != http.StatusNotFound {
		t.Fatalf("越界撤销=%d（期望 404）", code)
	}
	// 有界栈 FIFO 淘汰词形：12 写压栈次序=[nil,ov1..ov10]（满 10 层后 nil/ov1 先被淘汰）→
	// 剩余栈=[ov2..ov11]，10 撤后终态=ov2（takeBudget=MinTakeBudget+1；revision=12 写+10 撤
	// 单调=22）——nil 基线只在不溢栈时可恢复（有界=防拖内存的明示代价）。
	_, got := sessCtxDo(t, ts, http.MethodGet, "/v1/sessions/s1/context", "")
	ov := got["override"].(map[string]any)
	if ov == nil || ov["takeBudget"].(float64) != float64(MinTakeBudget+1) || ov["revision"].(float64) != 22 {
		t.Fatalf("有界撤销终态应=ov2（budget=%v revision=%v）", ov["takeBudget"], ov["revision"])
	}
}
