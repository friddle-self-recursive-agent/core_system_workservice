// secrets_test.go —— 秘密存储面 HTTP 契约验收（v0.34 secret 存储落地首段）：
// face 未注入=503（能力缺席可见）、写面缺省 405、落值受理即吞值（回执零值）、
// 列表零值出（脱敏红线）、删除幂等。
package apiserver

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/apply"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"github.com/friddle-self-recursive-agent/core_system_workservice/secretstore"
)

// secretsServer 带写面与秘密库的测试宿主（真实依赖链：git 后端 + 引擎 + 秘密库）。
func secretsServer(t *testing.T) *httptest.Server {
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
	st, err := secretstore.OpenSQLite("")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	engine := &apply.Engine{
		Sch: sch, Repo: repo, Bus: eventbus.NewBus(nil), Ops: ops,
		Perm:    fixturePerm(t),
		Subject: permission.Subject{User: "test", Roles: []string{permission.DefaultAdmin}},
	}
	ts := httptest.NewServer(New(sch, repo, engine, nil, nil, nil, nil, nil, nil, WithSecrets(st)))
	t.Cleanup(ts.Close)
	return ts
}

// secretsFaceOffServer 无秘密库无写面的裸宿主（关断口径）。
func secretsFaceOffServer(t *testing.T) *httptest.Server {
	t.Helper()
	sch, err := scheme.Load([]byte(fixture))
	if err != nil {
		t.Fatalf("载入 fixture: %v", err)
	}
	ts := httptest.NewServer(New(sch, nil, nil, nil, nil, nil, nil, nil, nil))
	t.Cleanup(ts.Close)
	return ts
}

func TestSecretsFaceOff(t *testing.T) {
	ts := secretsFaceOffServer(t)
	if code, _ := get(t, ts, "/v1/secrets"); code != http.StatusServiceUnavailable {
		t.Fatalf("face 未注入: 期望 503，实得 %d", code)
	}
	if code, _ := post(t, ts, "/v1/secrets", map[string]any{"name": "k", "value": "v"}); code != http.StatusMethodNotAllowed {
		t.Fatalf("写面未开启: 期望 405，实得 %d", code)
	}
}

func TestSecretsLifecycle(t *testing.T) {
	ts := secretsServer(t)
	// 落值：受理即吞值（回执零值）。
	code, body := post(t, ts, "/v1/secrets", map[string]any{"name": "smtp-auth", "value": "topsecret"})
	if code != http.StatusOK {
		t.Fatalf("落值: %d %v", code, body)
	}
	if raw, _ := json.Marshal(body); strings.Contains(string(raw), "topsecret") {
		t.Fatalf("回执携带值（脱敏红线）: %s", raw)
	}
	// 列表零值出。
	code, body = get(t, ts, "/v1/secrets")
	if code != http.StatusOK {
		t.Fatalf("列表: %d", code)
	}
	if raw, _ := json.Marshal(body); strings.Contains(string(raw), "topsecret") {
		t.Fatalf("列表脱敏失败: %s", raw)
	}
	if body["count"] != float64(1) {
		t.Fatalf("列表 count: %v", body)
	}
	// 删除幂等。
	code, body = deleteReq(t, ts, "/v1/secrets/smtp-auth")
	if code != http.StatusOK || body["deleted"] != true {
		t.Fatalf("删除: %d %v", code, body)
	}
	if _, body = get(t, ts, "/v1/secrets"); body["count"] != float64(0) {
		t.Fatalf("删除后列表: %v", body)
	}
	if code, _ = deleteReq(t, ts, "/v1/secrets/absent"); code != http.StatusOK {
		t.Fatalf("幂等删除: %d", code)
	}
	// 空值拒绝（400）。
	if code, _ = post(t, ts, "/v1/secrets", map[string]any{"name": "k", "value": ""}); code != http.StatusBadRequest {
		t.Fatalf("空值: %d", code)
	}
	// 缺 name（400）。
	if code, _ = post(t, ts, "/v1/secrets", map[string]any{"value": "v"}); code != http.StatusBadRequest {
		t.Fatalf("缺 name: %d", code)
	}
}

// deleteReq DELETE 并解析响应（秘密面/草稿面删除共用形态）。
func deleteReq(t *testing.T, ts *httptest.Server, path string) (int, map[string]any) {
	t.Helper()
	req, err := http.NewRequest(http.MethodDelete, ts.URL+path, nil)
	if err != nil {
		t.Fatalf("new request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("DELETE %s: %v", path, err)
	}
	defer resp.Body.Close()
	var body map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&body); err != nil {
		t.Fatalf("DELETE %s 响应非 JSON: %v", path, err)
	}
	return resp.StatusCode, body
}
