// design_docs_test.go —— 设计文档面 HTTP 契约验收（v0.52 能力 A——cored-bridge.md §2.4）：
// 清单/单文档读面（HEAD 权威+最近提交）、白名单外 403、HEAD 缺席 404、apply dry-run 零副作用、
// 真落库（git commit 在仓+DocApplied 事件+apply_log ok 行）、陈旧基线 409、lint 未过 400 携违规、
// 无主写入 400、fromDraft 取稿/清稿、草稿 upsert/弃稿、未注册仓 404、写面关断 405。
package apiserver

import (
	"net/http"
	"net/http/httptest"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/docrepo"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/draft"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

const designDocsScheme = `apiVersion: core/v1
kind: Scheme
metadata: {name: default, scope: builtin}
spec:
  version: core/v1
  kinds:
    - {kind: Scheme, builtin: true, storage: core/scheme.yaml}
  eventTypes:
    - {type: DocApplied, source: core, defaultLevel: info, route: owner-workflow}
`

// designDocsRole 写设计文档域 Role（core/* 资源域——scheme.yaml Role 声明词汇 v0.52 扩展）。
const designDocsRole = `apiVersion: core/v1
kind: Role
metadata: {name: design-editor, scope: global}
spec:
  permissions:
  - {scope: "core/*", actions: [read, write]}
`

// designDocRepo 建临时 git 设计文档仓（本地提交身份注入——不依赖宿主全局配置）。
func designDocRepo(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	git := func(args ...string) {
		cmd := exec.Command("git", args...)
		cmd.Dir = root
		env := append(os.Environ(),
			"GIT_AUTHOR_NAME=t", "GIT_AUTHOR_EMAIL=t@t",
			"GIT_COMMITTER_NAME=t", "GIT_COMMITTER_EMAIL=t@t")
		cmd.Env = env
		if out, err := cmd.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v（%s）", args, err, out)
		}
	}
	git("init", "-q")
	git("config", "user.name", "t")
	git("config", "user.email", "t@t")
	for p, c := range files {
		full := filepath.Join(root, filepath.FromSlash(p))
		if err := os.MkdirAll(filepath.Dir(full), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(full, []byte(c), 0o644); err != nil {
			t.Fatal(err)
		}
		git("add", "--", p)
	}
	git("commit", "-q", "-m", "seed 概念文档")
	return root
}

func designDocsServer(t *testing.T) (*httptest.Server, *docrepo.Set, *draft.Store) {
	return designDocsServerVariant(t, true)
}

func designDocsServerNoEngine(t *testing.T) *httptest.Server {
	ts, _, _ := designDocsServerVariant(t, false)
	return ts
}

func designDocsServerVariant(t *testing.T, withEngine bool) (*httptest.Server, *docrepo.Set, *draft.Store) {
	t.Helper()
	sch, err := scheme.Load([]byte(designDocsScheme))
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
	dr := designDocRepo(t, map[string]string{"concepts/proposal.md": "# 提案\n正文。\n"})
	set, err := docrepo.NewSet("core_design=" + dr)
	if err != nil {
		t.Fatal(err)
	}
	drafts := draft.NewStore(nil)
	bus := eventbus.NewBus(&eventbus.Options{Ops: ops})
	perm := fixturePerm(t)
	if err := perm.LoadObjects([]permission.ObjectDoc{
		{Kind: "Role", Name: "design-editor", Scope: "global", Content: []byte(designDocsRole)},
	}); err != nil {
		t.Fatalf("装载设计文档域 Role: %v", err)
	}
	if withEngine {
		set.BindEngine(&docrepo.Engine{Bus: bus, Ops: ops, Perm: perm, Domain: "core/design-docs"})
	}
	ts := httptest.NewServer(New(sch, repo, nil, drafts, nil, nil, nil, nil, nil,
		WithEventLog(ops), WithApplyLog(ops), WithDesignDocs(set)))
	t.Cleanup(ts.Close)
	return ts, set, drafts
}

// postDesignDoc 以 design-editor 主体 POST（role 断言须对上装载 Role——core/* 域）。
func postDesignDoc(t *testing.T, ts *httptest.Server, path string, payload any) (int, map[string]any) {
	t.Helper()
	return postHeader(t, ts, path, payload, authHeader("agent:design-probe", "role/design-editor"))
}

func TestDesignDocsReadFace(t *testing.T) {
	ts, _, _ := designDocsServer(t)
	code, body := get(t, ts, "/v1/design-docs")
	if code != http.StatusOK {
		t.Fatalf("清单: %d %v", code, body)
	}
	if body["count"] != float64(1) {
		t.Fatalf("清单计数: %v", body)
	}
	repos := body["repos"].([]any)
	if len(repos) != 1 || repos[0].(map[string]any)["name"] != "core_design" {
		t.Fatalf("仓分组投影: %v", repos)
	}
	// 单文档读面：HEAD 正文+blob sha+最近提交（任务书判据 1——打开显示正文与最近提交）。
	code, body = get(t, ts, "/v1/design-docs/core_design/concepts/proposal.md")
	if code != http.StatusOK {
		t.Fatalf("单文档: %d %v", code, body)
	}
	if body["content"] != "# 提案\n正文。\n" || body["sha"] == "" {
		t.Fatalf("读面投影: %v", body)
	}
	lc := body["lastCommit"].(map[string]any)
	if lc["subject"] != "seed 概念文档" || lc["at"] == "" {
		t.Fatalf("最近提交投影: %v", lc)
	}
	// 白名单外（非 .md）→ 403。
	code, _ = get(t, ts, "/v1/design-docs/core_design/core/scheme.yaml")
	if code != http.StatusForbidden {
		t.Fatalf("非白名单路径应 403（实得 %d）", code)
	}
	// 路径穿越：HTTP 层先被 ServeMux 规范化（404）——白名单闸（403）在 docrepo.validatePath 单测锁定。
	code, _ = get(t, ts, "/v1/design-docs/core_design/../etc/passwd.md")
	if code != http.StatusNotFound && code != http.StatusForbidden {
		t.Fatalf("路径穿越不应达读面（实得 %d）", code)
	}
	// HEAD 缺席 → 404；未注册仓 → 404。
	if code, _ = get(t, ts, "/v1/design-docs/core_design/concepts/absent.md"); code != http.StatusNotFound {
		t.Fatalf("HEAD 缺席应 404（实得 %d）", code)
	}
	if code, _ = get(t, ts, "/v1/design-docs/other-repo/a.md"); code != http.StatusNotFound {
		t.Fatalf("未注册仓应 404（实得 %d）", code)
	}
}

func TestDesignDocsApplyFlow(t *testing.T) {
	ts, _, drafts := designDocsServer(t)
	// ① dry-run（新文件）：零提交零事件零审计行，回执 sha=空（HEAD 无此文件）。
	code, body := postDesignDoc(t, ts, "/v1/design-docs/apply", map[string]any{
		"repo": "core_design", "path": "concepts/probe.md", "baseSha": "",
		"content": "# 探针文档\n", "fieldManager": "console:editor", "dryRun": true,
	})
	if code != http.StatusOK || body["action"] != "dry-run" {
		t.Fatalf("dry-run: %d %v", code, body)
	}
	// ② 真落库：git commit 在仓 + DocApplied + ok 行。
	code, body = postDesignDoc(t, ts, "/v1/design-docs/apply", map[string]any{
		"repo": "core_design", "path": "concepts/probe.md", "baseSha": "",
		"content": "# 探针文档\n", "fieldManager": "console:editor", "reasonRef": "console:apply",
	})
	if code != http.StatusOK || body["action"] != "create" {
		t.Fatalf("create: %d %v", code, body)
	}
	commit := body["commit"].(map[string]any)
	if commit["sha"] == "" || commit["subject"] == "" {
		t.Fatalf("commit 投影: %v", commit)
	}
	evt := body["event"].(map[string]any)
	if evt["type"] != "DocApplied" || evt["dedupKey"] == "" {
		t.Fatalf("DocApplied 同源: %v", evt)
	}
	code, body = get(t, ts, "/v1/design-docs/history/core_design/concepts/probe.md")
	if code != http.StatusOK || body["count"] != float64(1) {
		t.Fatalf("史面: %d %v", code, body)
	}
	row := body["entries"].([]any)[0].(map[string]any)
	if row["result"] != "ok" || row["fieldManager"] != "console:editor" {
		t.Fatalf("史面行: %v", row)
	}
	// ③ 落稿→fromDraft apply→稿清空（rule 6）。
	code, body = postDesignDoc(t, ts, "/v1/design-docs/draft", map[string]any{
		"repo": "core_design", "path": "concepts/probe.md", "content": "# 探针文档\n改\n",
		"fieldManager": "console:editor:designdoc",
	})
	if code != http.StatusOK || body["action"] != "staged" {
		t.Fatalf("落稿: %d %v", code, body)
	}
	sha := ""
	if code, body = get(t, ts, "/v1/design-docs/core_design/concepts/probe.md"); code == http.StatusOK {
		sha = body["sha"].(string)
	}
	code, body = postDesignDoc(t, ts, "/v1/design-docs/apply", map[string]any{
		"repo": "core_design", "path": "concepts/probe.md", "baseSha": sha,
		"fromDraft": true, "fieldManager": "console:editor:designdoc",
	})
	if code != http.StatusOK || body["action"] != "update" {
		t.Fatalf("fromDraft apply: %d %v", code, body)
	}
	if _, err := drafts.Get("DesignDoc", "core_design", "concepts/probe.md"); err == nil {
		t.Fatalf("apply 后稿未清（rule 6）")
	}
	// ④ 陈旧基线 → 409（expected/current 在载荷）。
	code, body = postDesignDoc(t, ts, "/v1/design-docs/apply", map[string]any{
		"repo": "core_design", "path": "concepts/probe.md", "baseSha": "deadbeef",
		"content": "# 陈旧\n", "fieldManager": "console:editor",
	})
	if code != http.StatusConflict || body["expected"] != "deadbeef" || body["current"] == "" {
		t.Fatalf("乐观锁 409: %d %v", code, body)
	}
	// ⑤ lint 未过 → 400 携违规路径/行号。
	code, body = postDesignDoc(t, ts, "/v1/design-docs/apply", map[string]any{
		"repo": "core_design", "path": "concepts/probe.md", "baseSha": body["current"].(string),
		"content": "# 探针\n去 10.9.8.7 排查\n", "fieldManager": "console:editor",
	})
	if code != http.StatusBadRequest {
		t.Fatalf("lint 应 400（实得 %d）", code)
	}
	vs := body["violations"].([]any)
	if len(vs) != 1 || vs[0].(map[string]any)["rule"] != "ip" || vs[0].(map[string]any)["line"] != float64(2) {
		t.Fatalf("lint 违规投影: %v", body)
	}
	// ⑥ 越权路径/未注册仓/无主写入。
	code, _ = postDesignDoc(t, ts, "/v1/design-docs/apply", map[string]any{
		"repo": "core_design", "path": "core/scheme.yaml", "content": "x\n",
		"fieldManager": "f",
	})
	if code != http.StatusForbidden {
		t.Fatalf("非白名单应 403（实得 %d）", code)
	}
	code, _ = postDesignDoc(t, ts, "/v1/design-docs/apply", map[string]any{
		"repo": "nope", "path": "a.md", "content": "x\n", "fieldManager": "f",
	})
	if code != http.StatusNotFound {
		t.Fatalf("未注册仓应 404（实得 %d）", code)
	}
	code, _ = postDesignDoc(t, ts, "/v1/design-docs/apply", map[string]any{
		"repo": "core_design", "path": "b.md", "content": "x\n",
	})
	if code != http.StatusBadRequest {
		t.Fatalf("无主写入应 400（实得 %d）", code)
	}
	// ⑦ 幂等 noop：同内容重入零提交（action=noop，零 event 键）。
	code, cur := get(t, ts, "/v1/design-docs/core_design/concepts/probe.md")
	code, body = postDesignDoc(t, ts, "/v1/design-docs/apply", map[string]any{
		"repo": "core_design", "path": "concepts/probe.md", "baseSha": cur["sha"].(string),
		"content": cur["content"].(string), "fieldManager": "console:editor",
	})
	if code != http.StatusOK || body["action"] != "noop" {
		t.Fatalf("noop: %d %v", code, body)
	}
	if _, has := body["event"]; has {
		t.Fatalf("noop 不应发事件: %v", body)
	}
	// ⑧ 匿名主体 → 403（denied 行在账——permission.md 规则 3 判定留痕）。
	code, _ = postHeader(t, ts, "/v1/design-docs/apply", map[string]any{
		"repo": "core_design", "path": "b.md", "content": "x\n", "fieldManager": "f",
	}, authHeader("anon"))
	if code != http.StatusForbidden {
		t.Fatalf("匿名主体应 403（实得 %d）", code)
	}
}

func TestDesignDocsWriteGate(t *testing.T) {
	// 写面关断（引擎未装配）→ 草稿写/apply 405；读面照常 200（装配即可读）。
	ts := designDocsServerNoEngine(t)
	if code, _ := get(t, ts, "/v1/design-docs"); code != http.StatusOK {
		t.Fatalf("读面应恒开: %d", code)
	}
	code, body := postDesignDoc(t, ts, "/v1/design-docs/draft", map[string]any{
		"repo": "core_design", "path": "a.md", "content": "x\n", "fieldManager": "f",
	})
	if code != http.StatusMethodNotAllowed {
		t.Fatalf("写面关断应 405（实得 %d %v）", code, body)
	}
	code, body = postDesignDoc(t, ts, "/v1/design-docs/apply", map[string]any{
		"repo": "core_design", "path": "a.md", "content": "x\n", "fieldManager": "f",
	})
	if code != http.StatusMethodNotAllowed {
		t.Fatalf("apply 写面关断应 405（实得 %d %v）", code, body)
	}
}

func TestDesignDocsDraftDiscard(t *testing.T) {
	ts, _, _ := designDocsServer(t)
	if code, _ := postDesignDoc(t, ts, "/v1/design-docs/draft", map[string]any{
		"repo": "core_design", "path": "a.md", "content": "x\n", "fieldManager": "f",
	}); code != http.StatusOK {
		t.Fatalf("落稿前置失败: %d", code)
	}
	req, err := http.NewRequest(http.MethodDelete, ts.URL+"/v1/design-docs/draft/core_design/a.md", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("弃稿应 200（实得 %d）", resp.StatusCode)
	}
	// 幂等 404。
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusNotFound {
		t.Fatalf("弃稿幂等应 404（实得 %d）", resp2.StatusCode)
	}
}
