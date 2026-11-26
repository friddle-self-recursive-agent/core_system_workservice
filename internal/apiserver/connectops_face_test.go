// connectops_face_test.go —— POST /v1/connectops/{name}/{op} ops 调试面单测：
// connector 未注册=404 / op 不在声明集=400 / subject 权限门 fail-closed（无效/无权=403 且
// Invoke 不达）/ audit 结果透传=200（findings 形状保持）/ findings 回查未命中=404 /
// method 不符=405（catch-all 代际判别器）/ subject 信封拆封不污 params。
package apiserver

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/connectors"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// connectopsAdminSubject 授权主体（DefaultAdmin 字面量——零值权限系统 fail-closed 下的
// 最低保障通路；Role 对象授权面走装配注入 *permission.System，不在本面单测范围）。
func connectopsAdminSubject() map[string]any {
	return map[string]any{"subject": map[string]any{"user": "auditor-test", "roles": []string{"global/admin"}}}
}

// probeOps 计数假体（G-03 权限门判据机械锚——403 路径 Invoke 计数必须不动）。
type probeOps struct{ calls atomic.Int64 }

func (p *probeOps) Type() string    { return "connectops-perm-probe" }
func (p *probeOps) Version() string { return "test" }
func (p *probeOps) Ops() []string   { return []string{"ping"} }
func (p *probeOps) Invoke(_ context.Context, _ string, _ []byte) ([]byte, error) {
	p.calls.Add(1)
	return []byte(`{"ok":true}`), nil
}
func (p *probeOps) Probe(_ context.Context) error { return nil }

// connectopsFixture 最小 scheme（调试面透传实现侧结果——kind 注册面复用既有 fixture 惯例）。
const connectopsFixture = `apiVersion: core/v1
kind: Scheme
metadata: {name: connectops-test, scope: builtin}
spec:
  version: core/v1
  kinds:
    - {kind: Agent, builtin: true, storage: global/agents/<name>/agent.yaml}
`

func newConnectopsServer(t *testing.T) *httptest.Server {
	t.Helper()
	sch, err := scheme.Load([]byte(connectopsFixture))
	if err != nil {
		t.Fatal(err)
	}
	ts := httptest.NewServer(New(sch, nil, nil, nil, nil, nil, nil, nil, nil))
	t.Cleanup(ts.Close)
	return ts
}

// strayTarget 野对象目标目录（根级 stray.yaml 不落任何 storage 模板=unregistered-path）。
func strayTarget(t *testing.T) (string, string) {
	t.Helper()
	root := t.TempDir()
	schemePath := filepath.Join(root, "scheme.yaml")
	const sch = `apiVersion: core/v1
kind: Scheme
metadata: {name: codeaudit-test, scope: builtin}
spec:
  version: core/v1
  kinds:
    - {kind: Agent, builtin: true, storage: global/agents/<name>/agent.yaml}
`
	if err := os.WriteFile(schemePath, []byte(sch), 0o644); err != nil {
		t.Fatal(err)
	}
	target := filepath.Join(root, "docs")
	if err := os.MkdirAll(target, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(target, "stray.yaml"), []byte("kind: Agent\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return target, schemePath
}

func postConnectops(t *testing.T, ts *httptest.Server, name, op string, body map[string]any) (*http.Response, map[string]any) {
	t.Helper()
	raw, _ := json.Marshal(body)
	resp, err := http.Post(ts.URL+"/v1/connectops/"+name+"/"+op, "application/json", bytes.NewReader(raw))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { resp.Body.Close() })
	var out map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&out); err != nil {
		t.Fatalf("响应不可解析: %v", err)
	}
	return resp, out
}

func TestConnectOpsFaceRouting(t *testing.T) {
	ts := newConnectopsServer(t)

	// connector 未注册=404。
	resp, out := postConnectops(t, ts, "absent-connector", "audit", map[string]any{})
	if resp.StatusCode != http.StatusNotFound || !strings.Contains(out["error"].(string), "未注册") {
		t.Fatalf("未注册 connector 应 404: %d %v", resp.StatusCode, out)
	}

	// op 不在声明集=400。
	resp, out = postConnectops(t, ts, "connectops-code-audit", "webhook-send", map[string]any{})
	if resp.StatusCode != http.StatusBadRequest || !strings.Contains(out["error"].(string), "声明集") {
		t.Fatalf("未声明 op 应 400: %d %v", resp.StatusCode, out)
	}

	// method 不符=405（catch-all 代际判别器）。
	resp2, err := http.Get(ts.URL + "/v1/connectops/connectops-code-audit/audit")
	if err != nil {
		t.Fatal(err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusMethodNotAllowed {
		t.Fatalf("GET 应 405: %d", resp2.StatusCode)
	}
}

func TestConnectOpsAuditFlow(t *testing.T) {
	ts := newConnectopsServer(t)
	target, schemePath := strayTarget(t)

	// audit：subject 信封随体（授权主体过权限门；拆封后不进 params——target 仍可寻址即证）。
	auditBody := func(extra map[string]any) map[string]any {
		body := map[string]any{
			"subject":    map[string]any{"user": "auditor-test", "roles": []string{"global/admin"}},
			"target":     target,
			"checks":     []map[string]any{{"rule": "R-DOC-6", "checker": "cored-lint"}},
			"schemePath": schemePath,
		}
		for k, v := range extra {
			body[k] = v
		}
		return body
	}
	resp, out := postConnectops(t, ts, "connectops-code-audit", "audit", auditBody(nil))
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("audit 应 200: %d %v", resp.StatusCode, out)
	}
	if ok, has := out["ok"].(bool); !has || ok {
		t.Fatalf("存在野对象结果应 ok=false: %v", out)
	}
	findings, has := out["findings"].([]any)
	if !has || len(findings) != 1 {
		t.Fatalf("findings 形状不符: %v", out["findings"])
	}
	row := findings[0].(map[string]any)
	if row["rule"] != "R-DOC-6" || row["file"] != "stray.yaml" || row["severity"] != "error" {
		t.Fatalf("finding 回锚不符: %v", row)
	}
	if _, has := row["checkRef"]; !has {
		t.Fatalf("finding 缺 checkRef: %v", row)
	}
	auditID, _ := out["auditId"].(string)
	if auditID == "" {
		t.Fatalf("结果缺 auditId: %v", out)
	}

	// findings 回查命中=200 且同 id。
	resp, out = postConnectops(t, ts, "connectops-code-audit", "findings", map[string]any{
		"subject": map[string]any{"user": "auditor-test", "roles": []string{"global/admin"}},
		"auditId": auditID,
	})
	if resp.StatusCode != http.StatusOK || out["auditId"] != auditID {
		t.Fatalf("findings 回查应 200 同 id: %d %v", resp.StatusCode, out)
	}

	// 回查未命中=404（ErrNoAudit 与执行失败分立）。
	resp, _ = postConnectops(t, ts, "connectops-code-audit", "findings", map[string]any{
		"subject": map[string]any{"user": "auditor-test", "roles": []string{"global/admin"}},
		"auditId": "audit-absent",
	})
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("未知 auditId 应 404: %d", resp.StatusCode)
	}

	// audit 入参非法（相对路径 target）=502 执行失败面。
	resp, out = postConnectops(t, ts, "connectops-code-audit", "audit", auditBody(map[string]any{"target": "relative/dir"}))
	if resp.StatusCode != http.StatusBadGateway || !strings.Contains(out["error"].(string), "op 执行失败") {
		t.Fatalf("非法 target 应 502 执行失败面: %d %v", resp.StatusCode, out)
	}
}

// TestConnectOpsPermissionFailClosed G-03 权限门（fail-closed）：无效/无权 subject=403 且
// Invoke 不达（计数假体机械锚）；授权主体（DefaultAdmin 字面量）过门=200 且 Invoke 恰一次；
// 不可解析 X-Core-Subject 头=视为未携带（匿名照拒，不静默降级本地信任）。
func TestConnectOpsPermissionFailClosed(t *testing.T) {
	ts := newConnectopsServer(t)
	probe := &probeOps{}
	connectors.RegisterOpsConnector(probe) // 全局注册表（同型覆盖=幂等；仅本测试锚用）

	// 无效 subject（匿名——无 subject 字段无头）→ 403，Invoke 不达。
	resp, out := postConnectops(t, ts, "connectops-perm-probe", "ping", map[string]any{})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("匿名 subject 应 403 fail-closed，得 %d: %v", resp.StatusCode, out)
	}
	if !strings.Contains(out["error"].(string), "fail-closed") {
		t.Fatalf("403 错误词应标 fail-closed: %v", out)
	}
	if probe.calls.Load() != 0 {
		t.Fatalf("403 路径 Invoke 不得被调用，计得 %d", probe.calls.Load())
	}

	// 无权 subject（断言无授权 role）→ 403，Invoke 仍不达。
	resp, _ = postConnectops(t, ts, "connectops-perm-probe", "ping", map[string]any{
		"subject": map[string]any{"user": "dev-one", "roles": []string{"role/devtools"}},
	})
	if resp.StatusCode != http.StatusForbidden {
		t.Fatalf("无权 subject 应 403，得 %d", resp.StatusCode)
	}
	if probe.calls.Load() != 0 {
		t.Fatalf("无权路径 Invoke 不得被调用，计得 %d", probe.calls.Load())
	}

	// 授权主体过门 → 200 透传，Invoke 恰一次。
	resp, out = postConnectops(t, ts, "connectops-perm-probe", "ping", connectopsAdminSubject())
	if resp.StatusCode != http.StatusOK || out["ok"] != true {
		t.Fatalf("授权主体应 200 透传: %d %v", resp.StatusCode, out)
	}
	if probe.calls.Load() != 1 {
		t.Fatalf("过门后 Invoke 应恰一次，计得 %d", probe.calls.Load())
	}

	// X-Core-Subject 头不可解析=匿名照拒（403）；可解析授权头=过门。
	req, _ := http.NewRequest(http.MethodPost, ts.URL+"/v1/connectops/connectops-perm-probe/ping", bytes.NewReader([]byte(`{}`)))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set(SubjectHeader, "{broken")
	resp2, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	resp2.Body.Close()
	if resp2.StatusCode != http.StatusForbidden {
		t.Fatalf("坏 subject 头应 403，得 %d", resp2.StatusCode)
	}
	if probe.calls.Load() != 1 {
		t.Fatalf("坏头路径 Invoke 不得被调用，计得 %d", probe.calls.Load())
	}
	req2, _ := http.NewRequest(http.MethodPost, ts.URL+"/v1/connectops/connectops-perm-probe/ping", bytes.NewReader([]byte(`{}`)))
	req2.Header.Set("Content-Type", "application/json")
	req2.Header.Set(SubjectHeader, `{"user":"header-admin","roles":["global/admin"]}`)
	resp3, err := http.DefaultClient.Do(req2)
	if err != nil {
		t.Fatal(err)
	}
	resp3.Body.Close()
	if resp3.StatusCode != http.StatusOK {
		t.Fatalf("授权 subject 头应 200，得 %d", resp3.StatusCode)
	}
	if probe.calls.Load() != 2 {
		t.Fatalf("头通路过门后 Invoke 应累计两次，计得 %d", probe.calls.Load())
	}
}
