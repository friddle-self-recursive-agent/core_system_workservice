package approvaldiff

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestChatTransient502Backoff code=0 空 reply 的 502 归一瞬态、指数退避 3s→6s，
// 第三次成功产真实判决（BugDoc 502 同族修复回归锚）。
func TestChatTransient502Backoff(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts < 3 {
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprint(w, `{"code":0,"reply":"","error":"ZCode did not stream"}`)
			return
		}
		w.Header().Set("content-type", "application/json")
		fmt.Fprint(w, `{"choices":[{"message":{"content":"  判决正文  "}}]}`)
	}))
	defer srv.Close()
	var sleeps []time.Duration
	g := &Generator{Sleep: func(d time.Duration) { sleeps = append(sleeps, d) }}
	got := g.chat(context.Background(), providerCfg{BaseURL: srv.URL, Model: "probe"}, "probe", "sys", "user")
	if got != "判决正文" {
		t.Fatalf("chat=%q，要第三次成功返回判决正文", got)
	}
	if attempts != 3 {
		t.Fatalf("attempts=%d，要 3", attempts)
	}
	if len(sleeps) != 2 || sleeps[0] != 3*time.Second || sleeps[1] != 6*time.Second {
		t.Fatalf("退避=%v，要 3s→6s 指数退避", sleeps)
	}
}

// TestChatTransientLimitEscalates 重试上限产 Escalate 失败词形（绝不静默编造/退进程）。
func TestChatTransientLimitEscalates(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadGateway)
		fmt.Fprint(w, `{"code":0,"reply":""}`)
	}))
	defer srv.Close()
	g := &Generator{Sleep: func(time.Duration) {}}
	got := g.chat(context.Background(), providerCfg{BaseURL: srv.URL}, "m", "s", "u")
	if attempts != chatAttempts {
		t.Fatalf("attempts=%d，要上限 %d", attempts, chatAttempts)
	}
	if !strings.Contains(got, "Escalate") || !strings.Contains(got, "判决不编造") {
		t.Fatalf("上限后=%q，要 Escalate + 不编造词形", got)
	}
}

// TestChatBodyCodeZeroEmptyIsTransient HTTP 200 但 body=code:0 + 空 reply 同样归一瞬态，
// 不能因为状态码绿就当成功（网关病征的第二种词形）。
func TestChatBodyCodeZeroEmptyIsTransient(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		if attempts == 1 {
			fmt.Fprint(w, `{"code": 0, "reply": ""}`)
			return
		}
		fmt.Fprint(w, `{"choices":[{"message":{"content":"ok"}}]}`)
	}))
	defer srv.Close()
	g := &Generator{Sleep: func(time.Duration) {}}
	if got := g.chat(context.Background(), providerCfg{BaseURL: srv.URL}, "m", "s", "u"); got != "ok" {
		t.Fatalf("chat=%q，要重试后 ok", got)
	}
	if attempts != 2 {
		t.Fatalf("attempts=%d，要 2（body 瞬态也应重试）", attempts)
	}
}

// TestChatNonTransientNoRetry 永久失败（403/400）不重试不升级为 502 误判。
func TestChatNonTransientNoRetry(t *testing.T) {
	attempts := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		w.WriteHeader(http.StatusForbidden)
	}))
	defer srv.Close()
	g := &Generator{Sleep: func(time.Duration) {}}
	got := g.chat(context.Background(), providerCfg{BaseURL: srv.URL}, "m", "s", "u")
	if attempts != 1 {
		t.Fatalf("attempts=%d，非瞬态不重试", attempts)
	}
	if !strings.Contains(got, "响应不可解析") {
		t.Fatalf("got=%q，要不可解析失败词形", got)
	}
}
