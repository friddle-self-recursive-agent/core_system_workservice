// proposals_face_test.go —— 需求提案读面 HTTP 契约验收（v0.51 能力 B——cored-bridge.md §2
// `GET /v1/proposals` + console-adoption.md §1.24）：列表/详情/状态过滤、对象缺省（诚实空态）、
// kind 未注册 404、非 GET 405、console:reject 对称扩展（rejected 行+ApprovalRejected 事件）
// 与 console:approve 批准链（granted 行+ApprovalGranted 事件——eventId 同源可回指）。
package apiserver

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/apply"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/draft"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventpub"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// proposalFixtureScheme：Proposal kind + 审批信号族事件（publish 面注册闸门）。
const proposalFixtureScheme = `apiVersion: core/v1
kind: Scheme
metadata: {name: default, scope: builtin}
spec:
  version: core/v1
  kinds:
    - {kind: Scheme, builtin: true, storage: core/scheme.yaml}
    - {kind: Agent, builtin: true, storage: global/agents/<name>/agent.yaml}
    - kind: Proposal
      builtin: true
      storage: "global/proposals/<name>.yaml | systems/<s>/proposals/<name>.yaml"
      spec:
        title: "!string"
        summary: "!string"
        acceptance: "[]string"
        impact: "?string"
        priority: "low|normal|high|critical"
        proposer: "!string"
        status: "!draft|submitted|approved|rejected|landed"
        landing: "?{target: doc|task, ref: ?string}"
  eventTypes:
    - {type: DocApplied, source: core, defaultLevel: info, route: owner-workflow}
    - {type: ApprovalGranted, source: human, defaultLevel: info, dedup: "approval:<obj>", route: dead-letter}
    - {type: ApprovalRejected, source: human, defaultLevel: warn, dedup: "approval:<obj>", route: dead-letter}
`

// proposalPublisherRole 发布域 Role（publish×events——审批信号族事件随行的主体授权；
// 与对象 write 分立——event.md 规则 8 两域不互为准入）。
const proposalPublisherRole = `apiVersion: core/v1
kind: Role
metadata: {name: prop-publisher, scope: global}
spec:
  permissions:
  - {scope: "events", actions: [publish]}
`

// postProp 以 fixture-admin+发布域双 role 主体 POST（审批判定主体=对象 write+publish 两域齐备）。
func postProp(t *testing.T, ts *httptest.Server, path string, payload any) (int, map[string]any) {
	t.Helper()
	return postHeader(t, ts, path, payload, authHeader("agent:test", "role/fixture-admin", "role/prop-publisher"))
}

func proposalServer(t *testing.T) (*httptest.Server, *apply.Engine) {
	t.Helper()
	sch, err := scheme.Load([]byte(proposalFixtureScheme))
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
	bus := eventbus.NewBus(&eventbus.Options{Ops: ops})
	engine := &apply.Engine{
		Sch: sch, Repo: repo, Bus: bus, Ops: ops, Drafts: drafts,
		Perm:    fixturePerm(t),
		Subject: permission.Subject{User: "test", Roles: []string{permission.DefaultAdmin}},
	}
	perm := fixturePerm(t)
	if err := perm.LoadObjects([]permission.ObjectDoc{
		{Kind: "Role", Name: "prop-publisher", Scope: "global", Content: []byte(proposalPublisherRole)},
	}); err != nil {
		t.Fatalf("装载发布域 Role: %v", err)
	}
	publisher := &eventpub.Publisher{Sch: sch, Bus: bus, Perm: perm}
	ts := httptest.NewServer(New(sch, repo, engine, drafts, nil, nil, nil, nil, publisher,
		WithEventLog(ops), WithApplyLog(ops), WithApprovalLog(ops)))
	t.Cleanup(ts.Close)
	return ts, engine
}

func proposalContent(name, status, priority string) string {
	return fmt.Sprintf(`apiVersion: core/v1
kind: Proposal
metadata: {name: %s, scope: global}
spec:
  title: 提案 %s
  summary: 验证需求提交流
  acceptance:
    - 对象在库
    - 派单同源
  impact: probe
  priority: %s
  proposer: console-operator
  status: %s
  landing: {target: doc}
`, name, name, priority, status)
}

func applyProposal(t *testing.T, engine *apply.Engine, name, status, priority string) *apply.Result {
	t.Helper()
	res, err := engine.Apply(context.Background(), &repository.Object{
		Kind: "Proposal", Name: name, Scope: "global", ResourceVersion: 0,
		Content: []byte(proposalContent(name, status, priority)),
	}, apply.Options{FieldManager: "test:propose"})
	if err != nil {
		t.Fatalf("apply Proposal/%s: %v", name, err)
	}
	return res
}

// TestProposalsFaceListDetailFilter 列表=瘦身行（rv 降序）；?status= 过滤；详情=全 spec+原文。
func TestProposalsFaceListDetailFilter(t *testing.T) {
	ts, engine := proposalServer(t)
	// 空态诚实。
	code, body := get(t, ts, "/v1/proposals")
	if code != http.StatusOK || body["count"] != float64(0) {
		t.Fatalf("空列表: %d %v", code, body)
	}
	// 两条提案（不同状态）——创建顺序 a（approved）后 b（submitted）。
	applyProposal(t, engine, "prop-a", "approved", "high")
	resB := applyProposal(t, engine, "prop-b", "submitted", "low")
	// 列表（同 rv=1 → 名升序兜底）：prop-a 在前——确定性投影锚。
	code, body = get(t, ts, "/v1/proposals")
	if code != http.StatusOK || body["count"] != float64(2) {
		t.Fatalf("列表: %d %v", code, body)
	}
	rows := body["proposals"].([]any)
	if rows[0].(map[string]any)["name"] != "prop-a" {
		t.Fatalf("同 rv 名升序期望 prop-a 在前: %v", body)
	}
	// 状态过滤（prop-b 仍 submitted）。
	code, body = get(t, ts, "/v1/proposals?status=approved")
	if code != http.StatusOK || body["count"] != float64(1) {
		t.Fatalf("status 过滤: %d %v", code, body)
	}
	row := body["proposals"].([]any)[0].(map[string]any)
	if row["name"] != "prop-a" || row["title"] != "提案 prop-a" || row["proposer"] != "console-operator" {
		t.Fatalf("过滤行投影: %v", row)
	}
	// rv 降序确定性投影：prop-b 更新到 rv=2 → 排最前。
	if _, err := engine.Apply(context.Background(), &repository.Object{
		Kind: "Proposal", Name: "prop-b", Scope: "global", ResourceVersion: resB.RV,
		Content: []byte(proposalContent("prop-b", "approved", "low")),
	}, apply.Options{FieldManager: "test:propose"}); err != nil {
		t.Fatalf("prop-b 更新: %v", err)
	}
	code, body = get(t, ts, "/v1/proposals")
	if rows[0] = body["proposals"].([]any)[0]; rows[0].(map[string]any)["name"] != "prop-b" {
		t.Fatalf("rv 降序期望 prop-b(rv=2) 在前: %v", body)
	}
	// 详情（scope 缺省 global 命中）。
	code, body = get(t, ts, "/v1/proposals/prop-a")
	if code != http.StatusOK {
		t.Fatalf("详情: %d %v", code, body)
	}
	spec := body["spec"].(map[string]any)
	if spec["status"] != "approved" || spec["priority"] != "high" || spec["impact"] != "probe" {
		t.Fatalf("详情 spec 投影: %v", spec)
	}
	acc := spec["acceptance"].([]any)
	if len(acc) != 2 || acc[1] != "派单同源" {
		t.Fatalf("验收标准投影: %v", acc)
	}
	if body["landing"].(map[string]any)["target"] != "doc" {
		t.Fatalf("产物指针投影: %v", body["landing"])
	}
	// 缺席=404（不冒充空对象）。
	if code, _ := get(t, ts, "/v1/proposals/absent"); code != http.StatusNotFound {
		t.Fatalf("缺席详情应 404，得 %d", code)
	}
	// 非 GET=405（查询面只读）。
	resp, err := http.Post(ts.URL+"/v1/proposals", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("POST /v1/proposals 应 405，得 %d", resp.StatusCode)
	}
}

// TestProposalRejectHookAndEvents console:reject 对称扩展：fromDraft 驳回 → approval_log
// rejected 行（rv=落库后版本）∧ ApprovalRejected 事件归档（involved=Proposal 锚）；
// console:approve → granted 行 ∧ ApprovalGranted 事件（判定与事件同拍可回指）。
func TestProposalRejectHookAndEvents(t *testing.T) {
	ts, engine := proposalServer(t)
	res := applyProposal(t, engine, "prop-r", "submitted", "high")
	// 草稿=状态写 rejected（console 驳回动作的稿形态）。
	if code, _ := post(t, ts, "/v1/drafts", map[string]any{
		"kind": "Proposal", "scope": "global", "name": "prop-r",
		"content":             proposalContent("prop-r", "rejected", "high"),
		"fieldManager":        "console:reject",
		"baseResourceVersion": res.RV,
	}); code != http.StatusOK {
		t.Fatal("驳回稿落稿失败")
	}
	code, body := postProp(t, ts, "/v1/apply", map[string]any{
		"kind": "Proposal", "scope": "global", "name": "prop-r", "fromDraft": true,
		"fieldManager": RejectFieldManager, "reasonRef": "console:reject",
	})
	if code != http.StatusOK {
		t.Fatalf("驳回 apply: %d %v", code, body)
	}
	if code, body := get(t, ts, "/v1/approvals?decision=rejected"); code != http.StatusOK {
		t.Fatalf("rejected 账面: %d", code)
	} else {
		rows := body["approvals"].([]any)
		if len(rows) != 1 {
			t.Fatalf("rejected 行应 1: %v", body)
		}
		row := rows[0].(map[string]any)
		if row["fieldManager"] != RejectFieldManager || row["rv"] != float64(res.RV+1) {
			t.Fatalf("rejected 行投影: %v", row)
		}
	}
	// 事件随行：ApprovalRejected 归档（source=human，involved=Proposal 锚）。
	if code, body := get(t, ts, "/v1/events?type=ApprovalRejected"); code != http.StatusOK {
		t.Fatalf("ApprovalRejected 归档: %d", code)
	} else {
		evs := body["events"].([]any)
		if len(evs) != 1 {
			t.Fatalf("ApprovalRejected 事件应 1: %v", body)
		}
		ev := evs[0].(map[string]any)
		if ev["source"] != "human" || ev["involved"] != "Proposal/global/prop-r" {
			t.Fatalf("事件投影: %v", ev)
		}
	}
	// 批准链同构验证：另一提案 console:approve → granted 行 + ApprovalGranted 事件。
	res2 := applyProposal(t, engine, "prop-g", "submitted", "low")
	if code, _ := post(t, ts, "/v1/drafts", map[string]any{
		"kind": "Proposal", "scope": "global", "name": "prop-g",
		"content":             proposalContent("prop-g", "approved", "low"),
		"fieldManager":        "console:approve",
		"baseResourceVersion": res2.RV,
	}); code != http.StatusOK {
		t.Fatal("批准稿落稿失败")
	}
	if code, _ := postProp(t, ts, "/v1/apply", map[string]any{
		"kind": "Proposal", "scope": "global", "name": "prop-g", "fromDraft": true,
		"fieldManager": ApproveFieldManager, "reasonRef": "console:approve",
	}); code != http.StatusOK {
		t.Fatal("批准 apply 失败")
	}
	if code, body := get(t, ts, "/v1/approvals?decision=granted&obj=Proposal/global/prop-g"); code != http.StatusOK {
		t.Fatalf("granted 账面: %d", code)
	} else if rows := body["approvals"].([]any); len(rows) != 1 {
		t.Fatalf("granted 行应 1: %v", body)
	}
	if code, body := get(t, ts, "/v1/events?type=ApprovalGranted"); code != http.StatusOK {
		t.Fatalf("ApprovalGranted 归档: %d", code)
	} else if evs := body["events"].([]any); len(evs) != 1 {
		t.Fatalf("ApprovalGranted 事件应 1: %v", body)
	}
}

// TestProposalsKindUnregistered kind 未注册（旧声明面）→ 404（scheme 闸门语义与 objects 面一致）。
func TestProposalsKindUnregistered(t *testing.T) {
	sch, err := scheme.Load([]byte(fixture)) // fixture scheme 无 Proposal kind
	if err != nil {
		t.Fatal(err)
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
	ts := httptest.NewServer(New(sch, repo, nil, nil, nil, nil, nil, nil, nil))
	defer ts.Close()
	if code, _ := get(t, ts, "/v1/proposals"); code != http.StatusNotFound {
		t.Fatalf("kind 未注册应 404，得 %d", code)
	}
}
