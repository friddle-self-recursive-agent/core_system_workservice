// Package outbound —— cored 进程内自动投递的 exec 出站桥（cored-bridge.md §2.3 v0.17）。
//
// 职责：把一次送达投影渲染为出站信封（connectors.OutboundEnvelope，JSON），经子进程
// stdin 交给**实现侧出站命令**（部署侧装配注入，如连接器仓的邮件直达入口）——exit 0=
// 出站成功，非 0=失败。依赖方向不反转：本包（workservice）不 import 任何连接器仓，
// 实现侧命令由部署侧装配以 argv 注入（cmd/cored --outbound-exec）；secret 引用式取值
// （env:*）在实现侧进程内解析，cored 旗标/信封/日志均不经手凭据（子进程 env 全量透传）。
//
// 与 OpsConnector 契约的分野：OpsConnector（connectors 包）是进程内五方法契约（连接器仓
// 以 Go 包挂靠）；本包是**跨进程**装配面——同一实现侧出站能力，两种挂接方式任选，声明面
// （WorkService.tools/connector）不变。
package outbound

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
	"unicode"

	"github.com/friddle-self-recursive-agent/core_system_workservice/connectors"
)

// MaxErrTail 出站失败时留存到错误文本里的子进程输出尾部上限（字节）——
// 有界诊断：错误进送达投影可查，但不无限吞输出。
const MaxErrTail = 512

// DefaultWaitDelay 超时击杀后等待孙进程释放管道的上限（击杀只及直接子进程；
// 无此上界，持管道的孙进程可拖住 Run 至其自然退出）。
const DefaultWaitDelay = 3 * time.Second

// Command exec 出站桥（并发安全：无共享可变状态，一次 Send 一个子进程会话）。
type Command struct {
	argv      []string
	WaitDelay time.Duration // 超时击杀后管道释放等待上界（0 → DefaultWaitDelay）
}

// NewCommand 构建出站桥（argv 至少一个令牌；argv[0]=可执行）。
func NewCommand(argv []string) (*Command, error) {
	if len(argv) == 0 || strings.TrimSpace(argv[0]) == "" {
		return nil, fmt.Errorf("outbound: argv 为空（--outbound-exec 需至少给出可执行）")
	}
	cp := make([]string, len(argv))
	copy(cp, argv)
	return &Command{argv: cp}, nil
}

// ParseCommand 解析出站命令行：空白分隔令牌，支持单/双引号段（段内空白保留、无转义——
// 模板面保持最小；复杂引号需求走包装脚本）。空串/仅引号外空白=错误。
func ParseCommand(s string) ([]string, error) {
	var toks []string
	var cur strings.Builder
	inTok, quote := false, rune(0)
	flush := func() {
		if inTok {
			toks = append(toks, cur.String())
			cur.Reset()
			inTok = false
		}
	}
	for _, r := range s {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote = r
			inTok = true
		case unicode.IsSpace(r):
			flush()
		default:
			cur.WriteRune(r)
			inTok = true
		}
	}
	if quote != 0 {
		return nil, fmt.Errorf("outbound: 命令行引号未闭合（%q）", s)
	}
	flush()
	if len(toks) == 0 {
		return nil, fmt.Errorf("outbound: 命令行为空")
	}
	return toks, nil
}

// Send 执行一次真实出站：信封 JSON → 子进程 stdin；exit 0=成功。
// 子进程 stdout/stderr 合并捕获（仅失败时取尾部进错误文本——成功输出不落投影）。
// env 全量透传（secret 由实现侧自行引用式解析，provider.md 规则 6）。
func (c *Command) Send(ctx context.Context, env connectors.OutboundEnvelope) error {
	payload, err := json.Marshal(env)
	if err != nil {
		return fmt.Errorf("outbound: 信封编码失败: %w", err)
	}
	cmd := exec.CommandContext(ctx, c.argv[0], c.argv[1:]...)
	cmd.Env = Environ() // 全量透传（显式声明依赖方向：secret 只在实现侧进程内解析）
	waitDelay := c.WaitDelay
	if waitDelay <= 0 {
		waitDelay = DefaultWaitDelay
	}
	cmd.WaitDelay = waitDelay
	cmd.Stdin = bytes.NewReader(payload)
	var tail bytes.Buffer
	cmd.Stdout = &tail
	cmd.Stderr = &tail
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("outbound exec %s: %v: 输出尾=%q", c.argv[0], err, Tail(tail.String(), MaxErrTail))
	}
	return nil
}

// Tail 取 s 尾部至多 n 字节（rune 安全截断——诊断文本不撕裂多字节字符）。
func Tail(s string, n int) string {
	if len(s) <= n {
		return s
	}
	r := []rune(s)
	if len(r) > n {
		r = r[len(r)-n:]
	}
	return string(r)
}

// Render 内建渲染（模板 ref 声明面 deferred，notification.md §3）：subject=
// [cored] <type> <level> <channel>；body=投递溯源键值行（查询面可回指——机械可验，
// 非人类文案工程）。幂等确定性：同投影渲染结果恒相同。
func Render(env *connectors.OutboundEnvelope) {
	env.Subject = fmt.Sprintf("[cored] %s %s %s",
		strings.TrimSpace(env.Type), strings.TrimSpace(env.Level), env.Channel)
	var b strings.Builder
	fmt.Fprintf(&b, "cored 进程内自动投递（受理后异步出站）\n")
	fmt.Fprintf(&b, "事件类型: %s\n", env.Type)
	fmt.Fprintf(&b, "等级: %s\n", env.Level)
	fmt.Fprintf(&b, "通道: %s（origin=%s）\n", env.Channel, env.Origin)
	fmt.Fprintf(&b, "事件: %s\n", env.EventID)
	fmt.Fprintf(&b, "去重键: %s\n", env.DedupKey)
	fmt.Fprintf(&b, "路由: %s（owner=%s）\n", env.Route, env.Owner)
	fmt.Fprintf(&b, "送达序号: %d（触发留痕 seq=%d）\n", env.Delivery, env.TriggerSeq)
	fmt.Fprintf(&b, "受理时刻: %s\n", env.At)
	env.Body = b.String()
}

// Environ 缺省子进程环境（全量透传；独立函数=测试注入点语义锚）。
func Environ() []string { return os.Environ() }
