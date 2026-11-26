// connectops_face_test.go —— POST /v1/connectops/{name}/{op} ops 调试面单测：
// connector 未注册=404 / op 不在声明集=400 / audit 结果透传=200（findings 形状保持）/
// findings 回查未命中=404 / method 不符=405（catch-all 代际判别器）/ subject 信封拆封不污 params。
package apiserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

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

	// audit：subject 信封随体（拆封后不进 params——target 仍可寻址即证）。
	resp, out := postConnectops(t, ts, "connectops-code-audit", "audit", map[string]any{
		"subject":    map[string]any{"user": "auditor-test"},
		"target":     target,
		"checks":     []map[string]any{{"rule": "R-DOC-6", "checker": "cored-lint"}},
		"schemePath": schemePath,
	})
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
	resp, out = postConnectops(t, ts, "connectops-code-audit", "findings", map[string]any{"auditId": auditID})
	if resp.StatusCode != http.StatusOK || out["auditId"] != auditID {
		t.Fatalf("findings 回查应 200 同 id: %d %v", resp.StatusCode, out)
	}

	// 回查未命中=404（ErrNoAudit 与执行失败分立）。
	resp, _ = postConnectops(t, ts, "connectops-code-audit", "findings", map[string]any{"auditId": "audit-absent"})
	if resp.StatusCode != http.StatusNotFound {
		t.Fatalf("未知 auditId 应 404: %d", resp.StatusCode)
	}

	// audit 入参非法（相对路径 target）=502 执行失败面。
	resp, out = postConnectops(t, ts, "connectops-code-audit", "audit", map[string]any{"target": "relative/dir"})
	if resp.StatusCode != http.StatusBadGateway || !strings.Contains(out["error"].(string), "op 执行失败") {
		t.Fatalf("非法 target 应 502 执行失败面: %d %v", resp.StatusCode, out)
	}
}
