// outbound_test.go —— exec 出站桥验收（cored-bridge.md §2.3 v0.17）：
// 命令行解析（空白+引号段）、信封 stdin 协议（JSON→子进程）、exit 语义（0=sent/
// 非0=failed 携有界输出尾）、超时打断、内建渲染确定性。
package outbound

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"
	"unicode/utf8"

	"github.com/friddle-self-recursive-agent/core_system_workservice/connectors"
)

func testEnvelope() connectors.OutboundEnvelope {
	return connectors.OutboundEnvelope{
		Delivery: 7, Channel: "mail-chan", Origin: "route", TriggerSeq: 3,
		EventID: "ev-123", Type: "DocApplied", DedupKey: "apply:Role/global/probe:7",
		Level: "info", Route: "notification/mail-chan", Owner: "global",
		At: "2026-11-24T00:00:00Z",
	}
}

// TestParseCommand 命令行解析：空白分隔、引号段内空白保留、空/未闭合=错误。
func TestParseCommand(t *testing.T) {
	cases := []struct {
		name string
		in   string
		want []string
		ok   bool
	}{
		{"裸路径", "/usr/local/bin/send-email", []string{"/usr/local/bin/send-email"}, true},
		{"带参", "send-email --envelope --host smtp.example.com --port 465", []string{"send-email", "--envelope", "--host", "smtp.example.com", "--port", "465"}, true},
		{"双引号段内空白", `sh -c "cat > /tmp/out file.json"`, []string{"sh", "-c", "cat > /tmp/out file.json"}, true},
		{"单引号段", `sh -c 'echo a  b; exit 0'`, []string{"sh", "-c", "echo a  b; exit 0"}, true},
		{"混合", `wrap --msg "hello world" --x 'y z'`, []string{"wrap", "--msg", "hello world", "--x", "y z"}, true},
		{"空串", "   ", nil, false},
		{"引号未闭合", `sh -c "echo oops`, nil, false},
	}
	for _, c := range cases {
		got, err := ParseCommand(c.in)
		if c.ok && (err != nil || !equalStrings(got, c.want)) {
			t.Fatalf("%s: ParseCommand(%q)=%v,%v 期望 %v", c.name, c.in, got, err, c.want)
		}
		if !c.ok && err == nil {
			t.Fatalf("%s: 应报错，实得 %v", c.name, got)
		}
	}
	if _, err := NewCommand(nil); err == nil {
		t.Fatal("NewCommand(nil) 应报错")
	}
	if _, err := NewCommand([]string{"  "}); err == nil {
		t.Fatal("NewCommand(空白) 应报错")
	}
}

func equalStrings(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}

// TestSendEnvelopeStdin 成功路径：信封 JSON 经 stdin 送达子进程，exit 0=sent；
// 子进程侧回读文件断言信封形状（协议面机械验证，含中文正文多字节）。
func TestSendEnvelopeStdin(t *testing.T) {
	dir := t.TempDir()
	got := filepath.Join(dir, "envelope.json")
	argv, err := ParseCommand(`tee ` + got)
	if err != nil {
		t.Fatal(err)
	}
	br, err := NewCommand(argv)
	if err != nil {
		t.Fatal(err)
	}
	env := testEnvelope()
	Render(&env)
	if err := br.Send(context.Background(), env); err != nil {
		t.Fatalf("出站应成功: %v", err)
	}
	raw, err := os.ReadFile(got)
	if err != nil {
		t.Fatal(err)
	}
	var back connectors.OutboundEnvelope
	if err := json.Unmarshal(raw, &back); err != nil {
		t.Fatalf("子进程收到非合法信封 JSON: %v（%s）", err, raw)
	}
	if back != env {
		t.Fatalf("信封往返不一致:\n got %+v\nwant %+v", back, env)
	}
	if !strings.Contains(back.Subject, "DocApplied") || !strings.Contains(back.Body, "apply:Role/global/probe:7") {
		t.Fatalf("渲染应填充 subject/body: %+v", back)
	}
}

// TestSendExitSemantics 失败路径：非 0 exit=错误携退出状态与子进程输出尾（有界）。
func TestSendExitSemantics(t *testing.T) {
	argv, err := ParseCommand(`sh -c 'echo boom-line >&2; exit 3'`)
	if err != nil {
		t.Fatal(err)
	}
	br, _ := NewCommand(argv)
	err = br.Send(context.Background(), testEnvelope())
	if err == nil {
		t.Fatal("非 0 退出应报错")
	}
	msg := err.Error()
	if !strings.Contains(msg, "exit status 3") || !strings.Contains(msg, "boom-line") {
		t.Fatalf("错误应携退出状态与输出尾: %v", msg)
	}
}

// TestSendTimeout 超时打断：ctx 到期子进程被杀，Send 报错不挂死
// （exec 语义：替换 sh 为 sleep——击杀即杀到真进程，孙进程持管道的拖延面由 WaitDelay 兜底）。
func TestSendTimeout(t *testing.T) {
	argv, _ := ParseCommand(`sh -c 'exec sleep 5'`)
	br, _ := NewCommand(argv)
	ctx, cancel := context.WithTimeout(context.Background(), 80*time.Millisecond)
	defer cancel()
	start := time.Now()
	if err := br.Send(ctx, testEnvelope()); err == nil {
		t.Fatal("超时应报错")
	}
	if time.Since(start) > 2*time.Second {
		t.Fatalf("超时未生效（耗时 %s）", time.Since(start))
	}
}

// TestRender 内建渲染：确定性（同投影两次渲染恒等）+ subject/body 键位齐全。
func TestRender(t *testing.T) {
	a, b := testEnvelope(), testEnvelope()
	Render(&a)
	Render(&b)
	if a != b {
		t.Fatalf("渲染应确定性: %+v vs %+v", a, b)
	}
	wantSub := "[cored] DocApplied info mail-chan"
	if a.Subject != wantSub {
		t.Fatalf("subject 应 %q 实得 %q", wantSub, a.Subject)
	}
	if !strings.Contains(a.Body, "origin=route") || !strings.Contains(a.Body, a.EventID) ||
		!strings.Contains(a.Body, a.DedupKey) || !strings.Contains(a.Body, a.Route) {
		t.Fatalf("body 应携溯源键值行: %q", a.Body)
	}
}

// TestTail 有界截断：短串原样、长串取尾、多字节不撕裂（rune 安全）。
func TestTail(t *testing.T) {
	if got := Tail("short", 512); got != "short" {
		t.Fatalf("短串应原样: %q", got)
	}
	long := strings.Repeat("x", 600)
	if got := Tail(long, 512); len(got) != 512 {
		t.Fatalf("长串应截到 512 字节: %d", len(got))
	}
	multi := strings.Repeat("核", 400) // 1200 字节
	got := Tail(multi, 512)
	if !utf8.ValidString(got) {
		t.Fatalf("截断应 rune 安全: %q", got)
	}
	if !strings.HasSuffix(multi, got) {
		t.Fatal("截断应取尾部")
	}
}
