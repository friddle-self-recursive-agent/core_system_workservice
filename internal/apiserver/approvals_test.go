// approvals_test.go —— 审批判定账面 HTTP 契约验收（v0.46 审批独立账，ops-log.md §2 v0.46
// 注记 + console-adoption.md §1.17）：批准 granted 钩子（console:approve fromDraft 非 dry-run）、
// noop 幂等重入照落、dry-run 不落、驳回 rejected 钩子（DELETE 弃稿零文档面动作仍落账）、
// 读面投影/过滤、face 缺席 503、双账并存（apply_log 同拍在册）、非批准面 FM 不落账。
package apiserver

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/apply"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/draft"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// approvalServer 带审批账面的测试宿主（真实依赖链：git 后端+引擎+内存草稿缓存+ops DB 双账）。
func approvalServer(t *testing.T) (*httptest.Server, *opslog.SQLiteStore) {
	t.Helper()
	sch, err := scheme.Load([]byte(fixture))
	if err != nil {
		t.Fatalf("载入 fixture: %v", err)
	}
	ops, err := opslog.OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = ops.Close() })
	repo, err := repository.NewGit(t.TempDir(), sch, &repository.Options{Ops: ops})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = repo.Close() })
	drafts := draft.NewStore(nil)
	engine := &apply.Engine{
		Sch: sch, Repo: repo, Bus: eventbus.NewBus(nil), Ops: ops, Drafts: drafts,
		Perm:    fixturePerm(t),
		Subject: permission.Subject{User: "test", Roles: []string{permission.DefaultAdmin}},
	}
	ts := httptest.NewServer(New(sch, repo, engine, drafts, nil, nil, nil, nil, nil,
		WithApplyLog(ops), WithApprovalLog(ops)))
	t.Cleanup(ts.Close)
	return ts, ops
}

func agentDraftBody(name string, baseRV int64) map[string]any {
	return map[string]any{
		"kind": "Agent", "scope": "global", "name": name,
		"content":             "apiVersion: core/v1\nkind: Agent\nmetadata: {name: " + name + ", scope: global}\nspec: {autoLevel: all}\n",
		"fieldManager":        "agent/dev-agent",
		"reasonRef":           "drift/7",
		"baseResourceVersion": baseRV,
	}
}

// approvalsCount 读账行数（decision 过滤可选）。
func approvalsCount(t *testing.T, ts *httptest.Server, query string) int {
	t.Helper()
	code, body := get(t, ts, "/v1/approvals"+query)
	if code != http.StatusOK {
		t.Fatalf("GET /v1/approvals = %d %v", code, body)
	}
	rows, _ := body["approvals"].([]any)
	return len(rows)
}

// TestApprovalGrantedHookLedger 批准=console:approve fromDraft apply 成功 → granted 行
// （rv=落库 rv、judged_by=请求主体、approval_id=draft:<键>）；apply_log ok 行同拍在册（双账并存）。
func TestApprovalGrantedHookLedger(t *testing.T) {
	ts, ops := approvalServer(t)
	if code, _ := post(t, ts, "/v1/drafts", agentDraftBody("core-headless", 0)); code != http.StatusOK {
		t.Fatal("落稿失败")
	}
	code, body := postAuth(t, ts, "/v1/apply", map[string]any{
		"kind": "Agent", "scope": "global", "name": "core-headless", "fromDraft": true,
		"fieldManager": ApproveFieldManager, "reasonRef": "console:approve",
	})
	if code != http.StatusOK || body["action"] != "create" {
		t.Fatalf("批准 apply = %d %v", code, body)
	}
	// 双账并存：apply_log ok 行 + approval_log granted 行。
	if n, err := ops.CountApply("ok"); err != nil || n != 1 {
		t.Fatalf("apply_log(ok)=%d err=%v, 要 1", n, err)
	}
	if n, err := ops.CountApprovals(opslog.DecisionGranted); err != nil || n != 1 {
		t.Fatalf("approval_log(granted)=%d err=%v, 要 1", n, err)
	}
	code, body = get(t, ts, "/v1/approvals")
	if code != http.StatusOK {
		t.Fatalf("读面 = %d", code)
	}
	rows := body["approvals"].([]any)
	if len(rows) != 1 {
		t.Fatalf("行数=%d, 要 1", len(rows))
	}
	r := rows[0].(map[string]any)
	if r["decision"] != "granted" || r["obj"] != "Agent/global/core-headless" ||
		r["approvalId"] != "draft:Agent/global/core-headless" || r["fieldManager"] != ApproveFieldManager {
		t.Fatalf("granted 行投影: %v", r)
	}
	if r["rv"] != float64(1) {
		t.Fatalf("granted 行 rv=%v, 要 1（落库后版本）", r["rv"])
	}
	if r["judgedBy"] != "agent:test" {
		t.Fatalf("judged_by=%v, 要 agent:test（请求主体 ID）", r["judgedBy"])
	}
	// noop 幂等重入照落：同稿重提（baseRV=1 面上现值）→ action=noop → granted 行 rv=面上现值。
	if code, _ := post(t, ts, "/v1/drafts", agentDraftBody("core-headless", 1)); code != http.StatusOK {
		t.Fatal("落稿失败")
	}
	code, body = postAuth(t, ts, "/v1/apply", map[string]any{
		"kind": "Agent", "scope": "global", "name": "core-headless", "fromDraft": true,
		"fieldManager": ApproveFieldManager, "reasonRef": "console:approve",
	})
	if code != http.StatusOK || body["action"] != "noop" {
		t.Fatalf("noop apply = %d %v", code, body)
	}
	if n, err := ops.CountApprovals(opslog.DecisionGranted); err != nil || n != 2 {
		t.Fatalf("noop 后 granted 行=%d err=%v, 要 2（判定发生了，零变更可落——照落账）", n, err)
	}
	// dry-run 预检不落账：预检非判定。
	if code, _ := post(t, ts, "/v1/drafts", agentDraftBody("dry", 0)); code != http.StatusOK {
		t.Fatal("落稿失败")
	}
	code, _ = postAuth(t, ts, "/v1/apply", map[string]any{
		"kind": "Agent", "scope": "global", "name": "dry", "fromDraft": true, "dryRun": true,
		"fieldManager": ApproveFieldManager,
	})
	if code != http.StatusOK {
		t.Fatalf("dry-run = %d", code)
	}
	if n, err := ops.CountApprovals(opslog.DecisionGranted); err != nil || n != 2 {
		t.Fatalf("dry-run 后 granted 行=%d err=%v, 要 2（dry-run 零判定零落账）", n, err)
	}
	// 非 fromDraft（直接落库）不落账：审批屏判定识别键只在 fromDraft 通路。
	code, _ = postAuth(t, ts, "/v1/apply", map[string]any{
		"kind": "Agent", "scope": "global", "name": "direct", "content": agentDraftBody("direct", 0)["content"],
		"fieldManager": ApproveFieldManager,
	})
	if code != http.StatusOK {
		t.Fatalf("直接落库 = %d", code)
	}
	if n, err := ops.CountApprovals(opslog.DecisionGranted); err != nil || n != 2 {
		t.Fatalf("直接落库后 granted 行=%d err=%v, 要 2（非审批屏动作零判定行）", n, err)
	}
}

// TestApprovalRejectedHookLedger 驳回=DELETE 弃稿成功 → rejected 行（rv=0）；文档面零变化；
// 弃稿后读面无稿；幂等 404 不落账。
func TestApprovalRejectedHookLedger(t *testing.T) {
	ts, ops := approvalServer(t)
	// 先落一个对象（作为面基准），再提案变更并驳回——面内容零变化。
	if code, _ := post(t, ts, "/v1/drafts", agentDraftBody("rej", 0)); code != http.StatusOK {
		t.Fatal("落稿失败")
	}
	if code, _ := postAuth(t, ts, "/v1/apply", map[string]any{
		"kind": "Agent", "scope": "global", "name": "rej", "fromDraft": true, "fieldManager": "agent:confirm",
	}); code != http.StatusOK {
		t.Fatal("基准落库失败")
	}
	if n, err := ops.CountApprovals(""); err != nil || n != 0 {
		t.Fatalf("agent:confirm 落库不应落审批账: %d err=%v", n, err)
	}
	// 变更提案 → 驳回。
	if code, _ := post(t, ts, "/v1/drafts", map[string]any{
		"kind": "Agent", "scope": "global", "name": "rej",
		"content":             "apiVersion: core/v1\nkind: Agent\nmetadata: {name: rej, scope: global}\nspec: {autoLevel: confirm}\n",
		"fieldManager":        "agent/dev-agent",
		"baseResourceVersion": 1,
	}); code != http.StatusOK {
		t.Fatal("变更提案落稿失败")
	}
	req, _ := http.NewRequest(http.MethodDelete, ts.URL+"/v1/drafts/Agent/rej?scope=global", nil)
	req.Header.Set(SubjectHeader, authHeader("human:reviewer"))
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("弃稿 = %d", resp.StatusCode)
	}
	// 弃稿后读面无稿 + 对象面内容零变化（驳回≠反落库）。
	if code, _ := get(t, ts, "/v1/drafts/Agent/rej?scope=global"); code != http.StatusNotFound {
		t.Fatalf("弃稿后应无稿: %d", code)
	}
	code, body := get(t, ts, "/v1/objects/Agent/rej?scope=global")
	if code != http.StatusOK {
		t.Fatalf("对象读面 = %d", code)
	}
	if c, _ := body["content"].(string); c == "" || !contains(c, "autoLevel: all") {
		t.Fatalf("驳回后面对内容应零变化: %q", c)
	}
	// rejected 行在册。
	rows := approvalsRows(t, ts, "")
	if len(rows) != 1 {
		t.Fatalf("行数=%d, 要 1", len(rows))
	}
	r := rows[0]
	if r["decision"] != "rejected" || r["rv"] != float64(0) || r["judgedBy"] != "human:reviewer" {
		t.Fatalf("rejected 行投影: %v", r)
	}
	// 幂等 404（本就无稿）不落账。
	req, _ = http.NewRequest(http.MethodDelete, ts.URL+"/v1/drafts/Agent/ghost?scope=global", nil)
	resp, err = http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("无稿弃稿 = %d, 要 404", resp.StatusCode)
	}
	if n, err := ops.CountApprovals(opslog.DecisionRejected); err != nil || n != 1 {
		t.Fatalf("rejected 行=%d err=%v, 要 1（幂等无判定不落账）", n, err)
	}
}

// TestApprovalsFaceAbsentUnwired face 缺席=503（防御面；旧代判别器=405 兜底）。
func TestApprovalsFaceAbsentUnwired(t *testing.T) {
	ts, ops := approvalServer(t)
	_ = ts
	// 无 approval 口面宿主（仅 ApplyLog）。
	sch, _ := scheme.Load([]byte(fixture))
	ts2 := httptest.NewServer(New(sch, nil, nil, nil, nil, nil, nil, nil, nil, WithApplyLog(ops)))
	t.Cleanup(ts2.Close)
	code, body := get(t, ts2, "/v1/approvals")
	if code != http.StatusServiceUnavailable {
		t.Fatalf("face 缺席 = %d %v, 要 503", code, body)
	}
}

// TestApprovalsReadProjection decision/obj 过滤与 limit 收敛（读面投影语义）。
func TestApprovalsReadProjection(t *testing.T) {
	ts, ops := approvalServer(t)
	_ = ops
	for i := 0; i < 3; i++ {
		if err := ops.AppendApproval(opslog.ApprovalLog{
			ApprovalID: fmt.Sprintf("draft:Agent/global/a%d", i), Obj: fmt.Sprintf("Agent/global/a%d", i),
			Decision: opslog.DecisionGranted, JudgedBy: "console-operator", ResourceVersion: int64(i + 1),
		}); err != nil {
			t.Fatal(err)
		}
	}
	if err := ops.AppendApproval(opslog.ApprovalLog{
		ApprovalID: "draft:Agent/global/x", Obj: "Agent/global/x", Decision: opslog.DecisionRejected,
	}); err != nil {
		t.Fatal(err)
	}
	if n := approvalsCount(t, ts, "?decision=granted"); n != 3 {
		t.Fatalf("decision 过滤=%d, 要 3", n)
	}
	if n := approvalsCount(t, ts, "?decision=rejected"); n != 1 {
		t.Fatalf("decision 过滤=%d, 要 1", n)
	}
	if n := approvalsCount(t, ts, "?obj=Agent/global/x"); n != 1 {
		t.Fatalf("obj 过滤=%d, 要 1", n)
	}
	if n := approvalsCount(t, ts, "?limit=2"); n != 2 {
		t.Fatalf("limit 收敛=%d, 要 2", n)
	}
	// rv 降序：limit=2 取 rv 最高的两行（a2 rv=3、a1 rv=2）。
	_, body := get(t, ts, "/v1/approvals?limit=2")
	rows := body["approvals"].([]any)
	first := rows[0].(map[string]any)
	if first["rv"] != float64(3) || first["obj"] != "Agent/global/a2" {
		t.Fatalf("rv 降序首行: %v", first)
	}
	// decision 非闭合集 → 500（QueryApprovals 拒绝外溢为读面错误，不冒充空集）。
	if code, _ := get(t, ts, "/v1/approvals?decision=maybe"); code != http.StatusInternalServerError {
		t.Fatalf("闭合集外 decision = %d, 要 500", code)
	}
}

func approvalsRows(t *testing.T, ts *httptest.Server, query string) []map[string]any {
	t.Helper()
	code, body := get(t, ts, "/v1/approvals"+query)
	if code != http.StatusOK {
		t.Fatalf("GET /v1/approvals = %d %v", code, body)
	}
	raw, _ := body["approvals"].([]any)
	out := make([]map[string]any, 0, len(raw))
	for _, r := range raw {
		out = append(out, r.(map[string]any))
	}
	return out
}

func contains(s, sub string) bool {
	return len(sub) == 0 || (len(s) >= len(sub) && indexOf(s, sub) >= 0)
}

func indexOf(s, sub string) int {
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}
