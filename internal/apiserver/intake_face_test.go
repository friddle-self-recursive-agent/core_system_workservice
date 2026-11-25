package apiserver

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

func TestIntakeFace_ReadOnlyAndWriteEnabled(t *testing.T) {
	sch, err := scheme.Load([]byte(proposalFixtureScheme))
	if err != nil {
		t.Fatalf("scheme.Load: %v", err)
	}

	// 1. 只读装配（applier=nil）：GET /v1/intake/sessions 恒 200，POST /v1/intake/turn 与 /generate 返回 405
	roSrv := New(sch, nil, nil, nil, nil, nil, nil, nil, nil)
	w0 := httptest.NewRecorder()
	roSrv.ServeHTTP(w0, httptest.NewRequest(http.MethodGet, "/v1/intake/sessions", nil))
	if w0.Code != http.StatusOK {
		t.Fatalf("GET /v1/intake/sessions code=%d want 200", w0.Code)
	}
	w1 := httptest.NewRecorder()
	roSrv.ServeHTTP(w1, httptest.NewRequest(http.MethodPost, "/v1/intake/turn", bytes.NewBufferString(`{"kind":"Proposal"}`)))
	if w1.Code != http.StatusMethodNotAllowed {
		t.Fatalf("只读装配 POST /v1/intake/turn code=%d want 405", w1.Code)
	}

	// 2. 写面开启：turn → generate → sessions 审计可查
	ts, _ := proposalServer(t)
	turnBody := `{"kind":"Proposal","scope":"global","message":"标题：对话式录入探针；目标：测试对话生成；验收标准：1. 6轮内全绿；2. 过准入；影响面：前端；优先级：high","subject":{"user":"operator-1","roles":["role/global/admin"]}}`
	resp1, err := http.Post(ts.URL+"/v1/intake/turn", "application/json", bytes.NewBufferString(turnBody))
	if err != nil {
		t.Fatalf("POST /v1/intake/turn: %v", err)
	}
	defer resp1.Body.Close()
	if resp1.StatusCode != http.StatusOK {
		t.Fatalf("POST /v1/intake/turn code=%d", resp1.StatusCode)
	}
	var tr map[string]any
	if err := json.NewDecoder(resp1.Body).Decode(&tr); err != nil {
		t.Fatalf("unmarshal turn: %v", err)
	}
	sid, _ := tr["sessionId"].(string)
	if sid == "" || tr["ready"] != true {
		t.Fatalf("turn 期望 ready=true 且有 sessionId，得: %+v", tr)
	}

	genBody := `{"sessionId":"` + sid + `","kind":"Proposal","scope":"global","name":"prop-intake-1","fields":{},"subject":{"user":"operator-1","roles":["role/global/admin"]}}`
	resp2, err := http.Post(ts.URL+"/v1/intake/generate", "application/json", bytes.NewBufferString(genBody))
	if err != nil {
		t.Fatalf("POST /v1/intake/generate: %v", err)
	}
	defer resp2.Body.Close()
	if resp2.StatusCode != http.StatusOK {
		t.Fatalf("POST /v1/intake/generate code=%d", resp2.StatusCode)
	}
	var gr map[string]any
	if err := json.NewDecoder(resp2.Body).Decode(&gr); err != nil {
		t.Fatalf("unmarshal generate: %v", err)
	}
	if gr["ok"] != true {
		t.Fatalf("generate 期望 ok=true，得: %+v", gr)
	}

	resp3, err := http.Get(ts.URL + "/v1/intake/sessions/" + sid)
	if err != nil {
		t.Fatalf("GET /v1/intake/sessions/%s: %v", sid, err)
	}
	defer resp3.Body.Close()
	if resp3.StatusCode != http.StatusOK {
		t.Fatalf("GET /v1/intake/sessions/%s code=%d", sid, resp3.StatusCode)
	}
}
