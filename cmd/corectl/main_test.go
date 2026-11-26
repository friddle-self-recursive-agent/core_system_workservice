package main

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func runCLI(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var out, errOut strings.Builder
	code := run(args, &out, &errOut)
	return code, out.String(), errOut.String()
}

func TestTickPublishesReconcileTick(t *testing.T) {
	var got map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost || r.URL.Path != "/v1/events" {
			t.Fatalf("unexpected %s %s", r.Method, r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		w.Header().Set("Content-Type", "application/json")
		io.WriteString(w, `{"id":"evt-1","type":"ReconcileTick","level":"info","source":"human"}`)
	}))
	defer ts.Close()

	code, out, errOut := runCLI(t, "--server", ts.URL, "tick", "--dedup", "cli-test")
	if code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, errOut)
	}
	if got["type"] != "ReconcileTick" || got["dedupKey"] != "cli-test" {
		t.Fatalf("payload=%+v", got)
	}
	sub, _ := got["subject"].(map[string]any)
	if sub["user"] != "console-operator" {
		t.Fatalf("subject=%+v", sub)
	}
	if !strings.Contains(out, "ReconcileTick") {
		t.Fatalf("output=%s", out)
	}
}

func TestApplyParsesFrontmatter(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "doc.yaml")
	content := "---\nname: demo\nscope: global\ndocType: Agent\n---\nkind: Agent\n"
	if err := os.WriteFile(f, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	var got map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/apply" {
			t.Fatalf("path=%s", r.URL.Path)
		}
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		io.WriteString(w, `{"resource":"Agent/global/demo","rv":1,"action":"created"}`)
	}))
	defer ts.Close()

	code, _, errOut := runCLI(t, "--server", ts.URL, "apply", "-f", f, "--dry-run")
	if code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, errOut)
	}
	if got["kind"] != "Agent" || got["scope"] != "global" || got["name"] != "demo" {
		t.Fatalf("header parse=%+v", got)
	}
	if got["dryRun"] != true || got["content"] != content {
		t.Fatalf("body=%+v", got)
	}
}

func TestGetScopeQuery(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/objects/Agent" || r.URL.Query().Get("scope") != "global" {
			t.Fatalf("path=%s query=%s", r.URL.Path, r.URL.RawQuery)
		}
		io.WriteString(w, `{"count":0,"objects":[]}`)
	}))
	defer ts.Close()

	code, out, errOut := runCLI(t, "--server", ts.URL, "get", "Agent", "--scope", "global")
	if code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, errOut)
	}
	if !strings.Contains(out, `"count": 0`) {
		t.Fatalf("output=%s", out)
	}
}

func TestPublishHTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"发布权限拒绝（动作=publish × 域=events）"}`, http.StatusForbidden)
	}))
	defer ts.Close()

	code, _, errOut := runCLI(t, "--server", ts.URL, "event", "publish", "--type", "ReconcileTick")
	if code != 1 {
		t.Fatalf("exit=%d", code)
	}
	if !strings.Contains(errOut, "403") {
		t.Fatalf("stderr=%s", errOut)
	}
}

func TestLoopPublishesLoopRequested(t *testing.T) {
	var got map[string]any
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if err := json.NewDecoder(r.Body).Decode(&got); err != nil {
			t.Fatal(err)
		}
		io.WriteString(w, `{"id":"evt-loop","type":"LoopRequested","level":"info","source":"human"}`)
	}))
	defer ts.Close()

	code, out, errOut := runCLI(t, "--server", ts.URL, "loop", "--dedup", "cli-loop-test")
	if code != 0 {
		t.Fatalf("exit=%d stderr=%s", code, errOut)
	}
	if got["type"] != "LoopRequested" || got["dedupKey"] != "cli-loop-test" {
		t.Fatalf("payload=%+v", got)
	}
	if !strings.Contains(out, "LoopRequested") {
		t.Fatalf("output=%s", out)
	}
}

func TestParseDocHeader(t *testing.T) {
	kind, scope, name := parseDocHeader("---\nname: x\nscope: s\ndocType: Rule\n---\n")
	if kind != "Rule" || scope != "s" || name != "x" {
		t.Fatalf("got %s/%s/%s", kind, scope, name)
	}
}
