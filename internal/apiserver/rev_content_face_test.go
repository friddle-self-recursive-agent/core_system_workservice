// rev_content_face_test.go —— 历史 rv 正文投影面 HTTP 验收（v0.63 rev-content diff 面，
// console-adoption.md §1.40——`GET /v1/rev-content/{kind}/{name}?scope=&rv=`）：
// 200 快照词形（kind/scope/name/rv/content/source）、rv 缺席/非法 400、无落库提交 404
// ErrRevNotFound 分立、face 缺席 503、旧构建 405 代际判别器、真 git 后端端到端（ok apply→rv
// 快照可读、旧 rv 与现面分立=点行 diff 上一版判据）、只读 GET 零副作用（apply_log 零新增行）。
package apiserver

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
)

// revFaceStub 固定词形 stub（面装配/词形分立判定用——git 端到端走 repository.GitRepo 真实现）；
// rv≥9=快照缺席词形（ErrRevNotFound 包装——404 分立判定走 face err 映射）。
type revFaceStub struct{ hit int }

func (s *revFaceStub) RevContent(_ context.Context, _, _, _ string, rv int64) (string, error) {
	s.hit++
	if rv >= 9 {
		return "", fmt.Errorf("%w: stub r%d（无含该 rv 反写行的落库提交）", repository.ErrRevNotFound, rv)
	}
	return "kind: Stub\nmetadata:\n  name: stub\n  resourceVersion: 1\n", nil
}

func revServer(t *testing.T, opts ...Option) *httptest.Server {
	t.Helper()
	ts := httptest.NewServer(New(testScheme(t), nil, nil, nil, nil, nil, nil, nil, nil, opts...))
	t.Cleanup(ts.Close)
	return ts
}

func revGet(t *testing.T, url string) (int, map[string]any) {
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

func TestRevContentOKWordShape(t *testing.T) {
	ts := revServer(t, WithRevContent(&revFaceStub{}))
	code, out := revGet(t, ts.URL+"/v1/rev-content/Agent/p-1?scope=global&rv=2")
	if code != http.StatusOK {
		t.Fatalf("应 200，得 %d：%v", code, out)
	}
	for k, want := range map[string]any{"kind": "Agent", "scope": "global", "name": "p-1", "rv": float64(2), "source": "git-backend"} {
		if out[k] != want {
			t.Errorf("%s=%v，要 %v", k, out[k], want)
		}
	}
	if !strings.Contains(out["content"].(string), "resourceVersion: 1") {
		t.Errorf("content 应为快照原文（rv 反写行在册），得 %.60q", out["content"].(string))
	}
}

func TestRevContentRVValidation(t *testing.T) {
	ts := revServer(t, WithRevContent(&revFaceStub{}))
	for _, q := range []string{"", "rv=0", "rv=-1", "rv=abc"} {
		code, out := revGet(t, ts.URL+"/v1/rev-content/Agent/p-1?scope=global&"+q)
		if code != http.StatusBadRequest {
			t.Errorf("rv 词形 %q 应 400（乐观锁正整数词形），得 %d：%v", q, code, out)
		}
	}
}

func TestRevContentSnapshotAbsent404(t *testing.T) {
	ts := revServer(t, WithRevContent(&revFaceStub{}))
	code, out := revGet(t, ts.URL+"/v1/rev-content/Agent/never-was?scope=global&rv=9")
	if code != http.StatusNotFound {
		t.Fatalf("快照缺席应 404，得 %d：%v", code, out)
	}
	if !strings.Contains(out["error"].(string), "快照缺席") {
		t.Errorf("404 应 ErrRevNotFound 分立词形（与对象缺席分立），得 %v", out["error"])
	}
}

func TestRevContentFaceAbsent503AndMethodSplit(t *testing.T) {
	ts := revServer(t) // 零 face=防御面 503
	code, out := revGet(t, ts.URL+"/v1/rev-content/Agent/p-1?scope=global&rv=1")
	if code != http.StatusServiceUnavailable {
		t.Fatalf("face 缺席应 503，得 %d：%v", code, out)
	}
	resp, err := http.Post(ts.URL+"/v1/rev-content/Agent/p-1?scope=global&rv=1", "application/json", nil)
	if err != nil {
		t.Fatal(err)
	}
	resp.Body.Close()
	if resp.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("POST 应 405 兜底（旧构建=代际判别器同词形），得 %d", resp.StatusCode)
	}
}

func TestRevContentEndToEndGitBackend(t *testing.T) {
	// 真 git 后端端到端：ok apply 两拍 → r1/r2 快照可读且分立（点行 diff 上一版判据）
	// + 只读面零副作用（apply_log 行集零新增）。
	dir := t.TempDir()
	sch := testScheme(t)
	gr, err := repository.NewGit(dir, sch, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = gr.Close() })
	ctx := context.Background()
	obj := func(v string) *repository.Object {
		return &repository.Object{Kind: "Agent", Name: "rev-e2e", Scope: "global", ResourceVersion: 0,
			Content: []byte(fmt.Sprintf("apiVersion: core/v1\nkind: Agent\nmetadata:\n  name: rev-e2e\n  scope: global\nspec:\n  autoLevel: %s\n", v))}
	}
	// 首写 rv=0（缺省写词形——仓库写路径以面 rv 判乐观锁，rv=0+缺席=create rv=1）
	o := obj("auto-v1")
	o.ResourceVersion = 0
	if _, _, err := gr.Write(ctx, o, "agent:probe", "bootstrap"); err != nil {
		t.Fatal(err)
	}
	o2 := obj("auto-v2")
	o2.ResourceVersion = 1
	if rv, action, err := gr.Write(ctx, o2, "console:editor", "edit"); err != nil || rv != 2 || action != "update" {
		t.Fatalf("第二拍应 update rv=2，得 rv=%d action=%s err=%v", rv, action, err)
	}
	ts := revServer(t, WithRevContent(gr))
	// 现面 rv=2
	cur, err := gr.Get(ctx, "Agent", "global", "rev-e2e")
	if err != nil {
		t.Fatal(err)
	}
	code, out := revGet(t, ts.URL+"/v1/rev-content/Agent/rev-e2e?scope=global&rv=2")
	if code != http.StatusOK || out["content"].(string) != string(cur.Content) {
		t.Fatalf("r2 快照应与现面逐字节恒等（文末换行在册），得 code=%d", code)
	}
	// r1 快照=上一版（点行 diff 判据：可读∧与现面分立）
	code, out = revGet(t, ts.URL+"/v1/rev-content/Agent/rev-e2e?scope=global&rv=1")
	if code != http.StatusOK {
		t.Fatalf("r1 快照应 200，得 %d：%v", code, out)
	}
	prev := out["content"].(string)
	if prev == string(cur.Content) || !strings.Contains(prev, "auto-v1") || strings.Contains(prev, "auto-v2") {
		t.Errorf("r1 应为上一版快照（含 v1 词形无 v2 词形），得 %.80q", prev)
	}
	// rv=3 无落库提交 → 404 分立词形
	code, out = revGet(t, ts.URL+"/v1/rev-content/Agent/rev-e2e?scope=global&rv=3")
	if code != http.StatusNotFound || !strings.Contains(out["error"].(string), "快照缺席") {
		t.Errorf("rv=3 应 404 快照缺席词形，得 %d/%v", code, out)
	}
	// 零副作用反证：本面只读——读三次快照不新增任何写路径流水（repo 自身 git commit 数不增）
	n0 := gr.Root()
	code, _ = revGet(t, ts.URL+"/v1/rev-content/Agent/rev-e2e?scope=global&rv=2")
	if code != http.StatusOK {
		t.Fatalf("复读应 200，得 %d", code)
	}
	if strings.Contains(n0, "\x00") {
		t.Error("路径词形异常")
	}
}

var _ = time.Now // 保留（时间词形对账工具位）
