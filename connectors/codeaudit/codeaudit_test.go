// codeaudit_test.go —— connectops-code-audit 单测（R-TYPE-3 判据逐条落码）：
// 检查器选择（builtin/模板/未注册）、unbound 显式错误（绝不静默通过）、findings 形状
// （rule/severity/file/line/message/checkRef 回锚）、命令模板占位符白名单（注入输入被拒）、
// target 校验（绝对路径/目录存在）、agent 规则路由（skipped 对账）与注册表自注册。
package codeaudit

import (
	"bytes"
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/connectors"
)

// invokeAudit 测试helper：构造实现并跑一次 audit op。
func invokeAudit(t *testing.T, c *Connector, params map[string]any) Result {
	t.Helper()
	raw, err := json.Marshal(params)
	if err != nil {
		t.Fatal(err)
	}
	out, err := c.Invoke(context.Background(), OpAudit, raw)
	if err != nil {
		t.Fatalf("audit Invoke 失败: %v", err)
	}
	var res Result
	if err := json.Unmarshal(out, &res); err != nil {
		t.Fatalf("audit 结果不可解析: %v", err)
	}
	return res
}

func invokeFindings(t *testing.T, c *Connector, auditID string) (Result, error) {
	t.Helper()
	raw, _ := json.Marshal(findingsRequest{AuditID: auditID})
	out, err := c.Invoke(context.Background(), OpFindings, raw)
	if err != nil {
		return Result{}, err
	}
	var res Result
	if err := json.Unmarshal(out, &res); err != nil {
		t.Fatal(err)
	}
	return res, nil
}

// minimalScheme 最小 scheme 文件（cored-lint 后端注册表输入；storage 模板单分支）。
func minimalScheme(t *testing.T) string {
	t.Helper()
	const sch = `apiVersion: core/v1
kind: Scheme
metadata: {name: codeaudit-test, scope: builtin}
spec:
  version: core/v1
  kinds:
    - {kind: Agent, builtin: true, storage: global/agents/<name>/agent.yaml}
`
	path := filepath.Join(t.TempDir(), "scheme.yaml")
	if err := os.WriteFile(path, []byte(sch), 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// strayTarget 造一个含野对象的目标目录（根级 stray.yaml 不落任何 storage 模板=unregistered-path）。
func strayTarget(t *testing.T) string {
	t.Helper()
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "stray.yaml"), []byte("kind: Agent\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	return dir
}

// writeScript 落一个可执行检查器脚本（直 exec 直测命令模板后端——不经 shell 是被测行为本身）。
func writeScript(t *testing.T, dir, body string) string {
	t.Helper()
	path := filepath.Join(dir, "checker.sh")
	if err := os.WriteFile(path, []byte("#!/bin/sh\n"+body), 0o755); err != nil {
		t.Fatal(err)
	}
	return path
}

func TestRegistrySelfRegistration(t *testing.T) {
	c, ok := connectors.LookupOpsConnector(TypeName)
	if !ok {
		t.Fatalf("%s 未自注册进 ops 注册表", TypeName)
	}
	if c.Type() != TypeName || c.Version() == "" {
		t.Fatalf("契约访问器不符: type=%s version=%s", c.Type(), c.Version())
	}
	ops := c.Ops()
	if len(ops) != 2 || ops[0] != OpAudit || ops[1] != OpFindings {
		t.Fatalf("Ops()=%v 要 [audit findings]", ops)
	}
	if err := c.Probe(context.Background()); err != nil {
		t.Fatalf("Probe 应恒可达: %v", err)
	}
	if _, err := c.Invoke(context.Background(), "webhook-send", nil); err == nil {
		t.Fatalf("未声明 op 应显式拒绝")
	}
}

func TestAuditUnboundExplicit(t *testing.T) {
	c := New()
	res := invokeAudit(t, c, map[string]any{
		"target": t.TempDir(),
		"checks": []map[string]any{{"rule": "R-NOPE", "checker": "absent-checker"}},
	})
	if res.OK {
		t.Fatalf("未注册检查器不得通过：ok=%v", res.OK)
	}
	if len(res.Unbound) != 1 {
		t.Fatalf("unbound=%v 要恰一条", res.Unbound)
	}
	u := res.Unbound[0]
	if u.Rule != "R-NOPE" || u.CheckRef != "absent-checker" {
		t.Fatalf("unbound 回锚不符: %+v", u)
	}
	if !strings.Contains(u.Reason, "未注册") {
		t.Fatalf("unbound 原因应显式说明未注册: %s", u.Reason)
	}
	if len(res.Findings) != 0 {
		t.Fatalf("未绑定不应产生 findings: %v", res.Findings)
	}
	if res.Tool.Name != TypeName || res.Tool.Version != Version || res.RanAt == "" || res.AuditID == "" {
		t.Fatalf("结果形状不符: tool=%+v ranAt=%s auditId=%s", res.Tool, res.RanAt, res.AuditID)
	}
}

func TestAuditCoredLintBackend(t *testing.T) {
	schemePath := minimalScheme(t)
	c := New(WithSchemePath(schemePath))

	// 野对象目录：stray.yaml 不落任何 storage 模板 → error finding（阻断 ok）。
	res := invokeAudit(t, c, map[string]any{
		"target": strayTarget(t),
		"checks": []map[string]any{{"rule": "R-DOC-6", "ref": "rule/rule.md#R-DOC-6", "checker": CheckerCoreLint}},
	})
	if res.OK {
		t.Fatalf("存在野对象不得通过")
	}
	if len(res.Findings) != 1 {
		t.Fatalf("findings=%v 要恰一条", res.Findings)
	}
	f := res.Findings[0]
	if f.Rule != "R-DOC-6" || f.CheckRef != CheckerCoreLint {
		t.Fatalf("finding 回锚不符: %+v", f)
	}
	if f.Severity != sevError || f.File != "stray.yaml" {
		t.Fatalf("finding 形状不符: %+v", f)
	}
	if !strings.Contains(f.Message, "unregistered-path") {
		t.Fatalf("finding 消息应携带 lint code: %s", f.Message)
	}

	// 净目录：零发现且零 unbound → ok=true（真实通过，非默认通过）。
	clean := invokeAudit(t, c, map[string]any{
		"target": t.TempDir(),
		"checks": []map[string]any{{"rule": "R-DOC-6", "checker": CheckerCoreLint}},
	})
	if !clean.OK || len(clean.Findings) != 0 || len(clean.Unbound) != 0 {
		t.Fatalf("净目录应通过: ok=%v findings=%v unbound=%v", clean.OK, clean.Findings, clean.Unbound)
	}

	// schemePath 缺席（装配侧也未注入）→ 显式失败不冒充。
	noScheme := New()
	res2 := invokeAudit(t, noScheme, map[string]any{
		"target": t.TempDir(),
		"checks": []map[string]any{{"rule": "R-DOC-6", "checker": CheckerCoreLint}},
	})
	if res2.OK || len(res2.Findings) != 1 || !strings.Contains(res2.Findings[0].Message, "scheme") {
		t.Fatalf("注册表缺席应显式失败: ok=%v findings=%v", res2.OK, res2.Findings)
	}
}

func TestCommandTemplateBackend(t *testing.T) {
	dir := t.TempDir()
	// 检查器脚本：回显收到的 argv1（=被替换后的 target）为 finding message。
	script := writeScript(t, dir, `printf '{"findings":[{"file":"src/a.go","line":7,"severity":"warn","message":"target=%s"}]}\n' "$1"`+"\n")
	c := New(WithTemplate("lint/custom", script+" {target}"))
	res := invokeAudit(t, c, map[string]any{
		"target": t.TempDir(),
		"checks": []map[string]any{{"rule": "R-CUSTOM", "checker": "lint/custom"}},
	})
	if !res.OK {
		t.Fatalf("仅 warn finding 不得翻转 ok: %+v", res)
	}
	if len(res.Findings) != 1 {
		t.Fatalf("findings=%v 要恰一条", res.Findings)
	}
	f := res.Findings[0]
	if f.Rule != "R-CUSTOM" || f.CheckRef != "lint/custom" || f.File != "src/a.go" || f.Line != 7 || f.Severity != sevWarn {
		t.Fatalf("findings 形状不符: %+v", f)
	}
	if !strings.HasPrefix(f.Message, "target=/") {
		t.Fatalf("target 占位符应被替换为被审目录: %s", f.Message)
	}
}

func TestTemplateWhitelistRejectsForeignPlaceholder(t *testing.T) {
	dir := t.TempDir()
	script := writeScript(t, dir, `exit 0`+"\n")
	// 白名单外占位符（{userInput}）→ 模板拒绝执行 → error finding（绝不带病执行）。
	c := New(WithTemplate("lint/evil", script+" {target} {userInput}"))
	res := invokeAudit(t, c, map[string]any{
		"target": t.TempDir(),
		"checks": []map[string]any{{"rule": "R-EVIL", "checker": "lint/evil"}},
	})
	if res.OK {
		t.Fatalf("白名单外占位符不得通过")
	}
	if len(res.Findings) != 1 || !strings.Contains(res.Findings[0].Message, "白名单") {
		t.Fatalf("应显式报白名单拒绝: %+v", res.Findings)
	}
	if res.Findings[0].Severity != sevError {
		t.Fatalf("模板拒绝应 error 级: %+v", res.Findings[0])
	}
}

func TestCommandTargetNoShellInjection(t *testing.T) {
	dir := t.TempDir()
	// 脚本只回显 argv1；不经 shell 意味着 target 里的 ; 与 ` 恒为普通字符。
	script := writeScript(t, dir, `printf '{"findings":[{"file":"x","line":0,"severity":"warn","message":"argv=%s"}]}\n' "$1"`+"\n")
	c := New(WithTemplate("lint/echo", script+" {target}"))
	inject := filepath.Join(t.TempDir(), "a;touch pwn;b `id`;sp ace")
	if err := os.MkdirAll(inject, 0o755); err != nil {
		t.Fatal(err)
	}
	res := invokeAudit(t, c, map[string]any{
		"target": inject,
		"checks": []map[string]any{{"rule": "R-INJECT", "checker": "lint/echo"}},
	})
	if !res.OK || len(res.Findings) != 1 {
		t.Fatalf("注入串应原样透传且不炸: ok=%v findings=%v", res.OK, res.Findings)
	}
	if !strings.Contains(res.Findings[0].Message, "a;touch pwn;b") {
		t.Fatalf("注入串应原样出现在参数里（无 shell 解释）: %s", res.Findings[0].Message)
	}
	// 侧效应验证：pwn 文件不存在（;touch 从未作为命令执行）。
	if _, err := os.Stat(filepath.Join(filepath.Dir(inject), "pwn")); !os.IsNotExist(err) {
		t.Fatalf("注入串不得被执行（shell 语义渗透）")
	}
}

func TestSeverityNormalization(t *testing.T) {
	dir := t.TempDir()
	// 后端输出未知等级 banana → 归一 error（保守显式，不降级）。
	script := writeScript(t, dir, `printf '{"findings":[{"file":"x","line":1,"severity":"banana","message":"m"}]}\n'`+"\n")
	c := New(WithTemplate("lint/banana", script+" {target}"))
	res := invokeAudit(t, c, map[string]any{
		"target": t.TempDir(),
		"checks": []map[string]any{{"rule": "R-BAN", "checker": "lint/banana"}},
	})
	if res.OK || len(res.Findings) != 1 || res.Findings[0].Severity != sevError {
		t.Fatalf("未知等级应归一 error 并阻断: ok=%v findings=%v", res.OK, res.Findings)
	}
}

func TestChecksProjectionAndSkips(t *testing.T) {
	dir := t.TempDir()
	script := writeScript(t, dir, `exit 0`+"\n")
	c := New(WithTemplate("lint/custom", script+" {target}"))
	// checks 缺省 → 读 scheme runtime.ruleChecks 投影；code 行执行，agent/human 行 skipped 对账。
	res := invokeAudit(t, c, map[string]any{
		"target": t.TempDir(),
		"ruleChecks": []map[string]any{
			{"ruleId": "R-CODE-1", "mode": "code", "checker": "lint/custom", "severity": "error"},
			{"ruleId": "R-UX-3", "mode": "agent", "agent": "service-agent"},
			{"ruleId": "R-HUM-1", "mode": "human"},
		},
	})
	if !res.OK {
		t.Fatalf("code 行净通过且其余 skipped 时应 ok: %+v", res)
	}
	if len(res.Skipped) != 2 {
		t.Fatalf("agent/human 行应列 skipped 对账: %+v", res.Skipped)
	}
	for _, s := range res.Skipped {
		if s.Rule != "R-UX-3" && s.Rule != "R-HUM-1" {
			t.Fatalf("skipped 规则锚不符: %+v", s)
		}
		if !strings.Contains(s.Reason, "agent") {
			t.Fatalf("skipped 原因应说明路由（丢给 agent/人审）: %s", s.Reason)
		}
	}
}

func TestEmptyChecksExplicitFail(t *testing.T) {
	c := New()
	// checks/ruleChecks 均缺 → 空审核按失败（绝不静默通过）。
	res := invokeAudit(t, c, map[string]any{"target": t.TempDir()})
	if res.OK {
		t.Fatalf("空审核不得通过")
	}
	if len(res.Unbound) != 1 || !strings.Contains(res.Unbound[0].Reason, "空审核") {
		t.Fatalf("空审核应显式 unbound: %+v", res.Unbound)
	}
}

func TestTargetValidation(t *testing.T) {
	c := New()
	// 缺 target。
	if _, err := c.Invoke(context.Background(), OpAudit, []byte(`{}`)); err == nil {
		t.Fatalf("缺 target 应显式拒绝")
	}
	// 相对路径 target。
	if _, err := c.Invoke(context.Background(), OpAudit, []byte(`{"target":"relative/dir"}`)); err == nil {
		t.Fatalf("相对路径 target 应显式拒绝")
	}
	// 不存在的目录（零发现≠目录缺席）。
	if _, err := c.Invoke(context.Background(), OpAudit, []byte(`{"target":"/nonexistent-codeaudit-dir","checks":[{"rule":"R","checker":"x"}]}`)); err == nil {
		t.Fatalf("不可达 target 应显式拒绝")
	}
}

func TestFindingsRoundtripAndMiss(t *testing.T) {
	c := New()
	res := invokeAudit(t, c, map[string]any{
		"target": t.TempDir(),
		"checks": []map[string]any{{"rule": "R-1", "checker": "absent"}},
	})
	got, err := invokeFindings(t, c, res.AuditID)
	if err != nil {
		t.Fatalf("findings 回查失败: %v", err)
	}
	if got.AuditID != res.AuditID || len(got.Unbound) != len(res.Unbound) {
		t.Fatalf("回查结果与原结果不符: %+v vs %+v", got, res)
	}
	if _, err := invokeFindings(t, c, "audit-absent"); err == nil || !bytes.Contains([]byte(err.Error()), []byte("auditId")) {
		if err == nil {
			t.Fatalf("未知 auditId 应 ErrNoAudit")
		}
	}
	// 缺 auditId。
	if _, err := c.Invoke(context.Background(), OpFindings, []byte(`{}`)); err == nil {
		t.Fatalf("缺 auditId 应显式拒绝")
	}
}

func TestResultShapeJSONAnchors(t *testing.T) {
	// 契约锚：findings 行六字段名与整体结果五键恒在（JSON 形状冻结——调用侧消费面）。
	c := New()
	raw, err := c.Invoke(context.Background(), OpAudit, []byte(`{"target":"`+t.TempDir()+`","checks":[{"rule":"R","checker":"nope"}]}`))
	if err != nil {
		t.Fatal(err)
	}
	var shape map[string]any
	if err := json.Unmarshal(raw, &shape); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"ok", "auditId", "findings", "unbound", "ranAt", "tool"} {
		if _, has := shape[k]; !has {
			t.Fatalf("结果缺键 %s: %v", k, shape)
		}
	}
	row, _ := json.Marshal(shape["unbound"].([]any)[0])
	var u map[string]any
	if err := json.Unmarshal(row, &u); err != nil {
		t.Fatal(err)
	}
	for _, k := range []string{"rule", "reason"} {
		if _, has := u[k]; !has {
			t.Fatalf("unbound 行缺键 %s: %v", k, u)
		}
	}
}
