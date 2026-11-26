// backend.go —— 检查器后端（可插拔）。v1 内置两族：
//
//	① cored-lint：本仓净入 lint 引擎（internal/repository.LintDocs——与写路径同闸门清单，
//	   cored-lint 命令面同源）进程内直调，外加 scheme 镜像漂移闸（LintSchemeMirror）；
//	② 命令模板：config 声明固定模板，白名单占位符（仅 {target}）替换后直 exec 走 argv
//	   **不经 shell**——任意字符串（含 target）恒为单一代码参数，无拼接注入面。
//
// 检查器选择：builtin 在册 → 命令模板在册 → unbound（显式失败，见 codeaudit.go resolveChecks）。
package codeaudit

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strings"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// CheckerCoreLint 内置净入 lint 检查器引用名（scheme ruleChecks.checker 可指向此名）。
const CheckerCoreLint = "cored-lint"

// CheckRequest 单条检查执行输入（target=被审目录只读遍历；schemePath=cored-lint 注册表）。
type CheckRequest struct {
	Rule       string
	CheckRef   string
	Severity   string
	Target     string
	SchemePath string
}

// CheckFunc 检查器函数签名：返回 findings 或显式 error（失败语义由调用侧落 error 级
// finding——backend 不静默吞错）。
type CheckFunc func(ctx context.Context, req CheckRequest) ([]Finding, error)

// coreLintCheck 内置后端①：docsRoot 净入五闸（repository.LintDocs，写路径同闸门清单）
// + scheme 镜像漂移闸（LintSchemeMirror）。注册表缺席=显式失败不冒充（cored-lint 命令面
// 同口径：--scheme 必配）。
func coreLintCheck(_ context.Context, req CheckRequest) ([]Finding, error) {
	if req.SchemePath == "" {
		return nil, fmt.Errorf("cored-lint 后端需要 scheme 注册表（params.schemePath 或装配侧 WithSchemePath）——缺席=不判，不冒充零发现")
	}
	sch, err := scheme.LoadFile(req.SchemePath)
	if err != nil {
		return nil, fmt.Errorf("scheme 载入失败 %q: %w", req.SchemePath, err)
	}
	rows, _, err := repository.LintDocs(req.Target, sch)
	if err != nil {
		return nil, fmt.Errorf("净入 lint 遍历 %q: %w", req.Target, err)
	}
	out := make([]Finding, 0, len(rows))
	for _, f := range rows {
		out = append(out, Finding{
			Rule: req.Rule, Severity: req.Severity, File: f.Path, Line: 0,
			Message:  f.Code + ": " + f.Detail,
			CheckRef: req.CheckRef,
		})
	}
	// scheme 镜像漂移闸随行（cored-lint 命令面同清单——双源系漂移零检出缺口同款收口）。
	authority, err := os.ReadFile(req.SchemePath)
	if err == nil {
		if present, drift, derr := repository.LintSchemeMirror(req.Target, authority); derr == nil && present && drift != nil {
			out = append(out, Finding{
				Rule: req.Rule, Severity: req.Severity, File: drift.Path, Line: 0,
				Message:  drift.Code + ": " + drift.Detail,
				CheckRef: req.CheckRef,
			})
		}
	}
	return out, nil
}

// 命令模板占位符白名单：仅 {target}（被审目录路径）。其余 {…} 一律拒绝——固定模板、
// 不拼不可信输入进 shell（本实现不经 shell，白名单是第二道闸）。
var placeholderRe = regexp.MustCompile(`\{[^{}]*\}`)

const placeholderTarget = "{target}"

// buildArgv 模板→argv：词元切分（不支持引号——模板来自部署侧 config 固定声明，词元即参数）；
// 逐词元校验占位符白名单后替换 {target}。白名单外占位符=显式拒绝（注入输入被拒的落点）。
func buildArgv(template, target string) ([]string, error) {
	fields := strings.Fields(template)
	if len(fields) == 0 {
		return nil, fmt.Errorf("命令模板为空")
	}
	argv := make([]string, 0, len(fields))
	for _, f := range fields {
		for _, ph := range placeholderRe.FindAllString(f, -1) {
			if ph != placeholderTarget {
				return nil, fmt.Errorf("命令模板占位符 %s 不在白名单（仅 {target}）——模板拒绝执行", ph)
			}
		}
		argv = append(argv, strings.ReplaceAll(f, placeholderTarget, target))
	}
	return argv, nil
}

// templateCheck 后端②：固定模板命令执行器。输出契约：stdout 空=零发现；非空=JSON
// {"findings":[{file,line,severity,message}]}；非零退出/stderr=显式失败（error 透传）。
func (c *Connector) templateCheck(template string) CheckFunc {
	return func(ctx context.Context, req CheckRequest) ([]Finding, error) {
		argv, err := buildArgv(template, req.Target)
		if err != nil {
			return nil, err
		}
		// 直 exec 走 argv：不经 shell——target 恒为单一代码参数，无拼接注入面。
		cmd := exec.CommandContext(ctx, argv[0], argv[1:]...)
		var stdout, stderr bytes.Buffer
		cmd.Stdout, cmd.Stderr = &stdout, &stderr
		if err := cmd.Run(); err != nil {
			tail := strings.TrimSpace(stderr.String())
			if tail == "" {
				return nil, fmt.Errorf("命令退出异常: %v", err)
			}
			return nil, fmt.Errorf("命令退出异常: %v; stderr: %s", err, tail)
		}
		out := strings.TrimSpace(stdout.String())
		if out == "" {
			return nil, nil
		}
		var parsed struct {
			Findings []struct {
				File     string `json:"file"`
				Line     int    `json:"line"`
				Severity string `json:"severity"`
				Message  string `json:"message"`
			} `json:"findings"`
		}
		if err := json.Unmarshal([]byte(out), &parsed); err != nil {
			return nil, fmt.Errorf("stdout 非 findings JSON（契约={\"findings\":[{file,line,severity,message}]}）: %w", err)
		}
		fs := make([]Finding, 0, len(parsed.Findings))
		for _, f := range parsed.Findings {
			fs = append(fs, Finding{
				Rule: req.Rule, Severity: f.Severity, File: f.File, Line: f.Line,
				Message: f.Message, CheckRef: req.CheckRef,
			})
		}
		return fs, nil
	}
}
