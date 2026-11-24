// cored-lint —— docsRoot 存量对象净入 lint（scheme.md 规则 6 checker 应用点②的命令面薄壳）。
// 五闸引擎=internal/repository.LintDocs（与写路径同闸门清单，逐项镜像见 lint.go 头注）；
// 本命令只做装配：--scheme 载入权威注册表 → --docs 净入全量通配 → scheme 镜像漂移闸 → 报告。
// 纪律：纯只读（零写面零 ops DB 零 HTTP——lint 不落库不发事件，判据与 apply 引擎准入同源不旁路）。
//
// 退出码：0=净入全绿（零发现）  1=有发现  2=用法/装配错（scheme 载入失败、docsRoot 不可达等）。
// 输出：逐文件行（文本模式：VIOLATION <path> <code>: <detail>，全绿=仅汇总行；--json=机读报告），
// 尾行汇总 files=<对象文件数> findings=<发现数> mirror=<absent|ok|drift>。
//
// 用法：cored-lint --scheme <core/scheme.yaml> --docs <docsRoot> [--json]
package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

func main() {
	if err := run(); err != nil {
		var uerr usageError
		if errors.As(err, &uerr) {
			fmt.Fprintln(os.Stderr, "cored-lint:", uerr.msg)
			os.Exit(2)
		}
		fmt.Fprintln(os.Stderr, "cored-lint:", err)
		os.Exit(2)
	}
}

type usageError struct{ msg string }

func (e usageError) Error() string { return e.msg }

func run() error {
	schemePath := flag.String("scheme", os.Getenv("CORE_SYSTEM_SCHEME"),
		"scheme 权威声明路径（core/scheme.yaml；缺省读环境变量 CORE_SYSTEM_SCHEME——cored 同名旗标同口径）")
	docsRoot := flag.String("docs", "", "docsRoot 路径（净入扫描根，git 布局=导入导出格式口径）")
	asJSON := flag.Bool("json", false, "机读报告（{files, findings, mirror}）")
	flag.Parse()

	if *schemePath == "" {
		return usageError{"缺少 scheme 路径：用 --scheme <path> 或环境变量 CORE_SYSTEM_SCHEME"}
	}
	if *docsRoot == "" {
		return usageError{"缺少 docsRoot：用 --docs <path>"}
	}
	if fi, err := os.Stat(*docsRoot); err != nil || !fi.IsDir() {
		return usageError{fmt.Sprintf("docsRoot %q 不可达或非目录", *docsRoot)}
	}

	// scheme 载入=启动自检同源（cored 同闸门：载入失败即退出非 0）。
	sch, err := scheme.LoadFile(*schemePath)
	if err != nil {
		return err
	}

	// 净入五闸全量通配。
	findings, files, err := repository.LintDocs(*docsRoot, sch)
	if err != nil {
		return err
	}

	// scheme 镜像漂移闸：docsRoot 内 core/scheme.yaml（消费方镜像）vs 载入权威字节比对。
	authority, err := os.ReadFile(*schemePath)
	if err != nil {
		return fmt.Errorf("scheme 权威重读 %q: %w", *schemePath, err)
	}
	present, drift, err := repository.LintSchemeMirror(*docsRoot, authority)
	if err != nil {
		return err
	}
	mirror := "absent"
	if present {
		mirror = "ok"
		if drift != nil {
			findings = append(findings, *drift)
			mirror = "drift"
		}
	}

	type report struct {
		Files    int                  `json:"files"`
		Findings []repository.Finding `json:"findings"`
		Mirror   string               `json:"mirror"`
	}
	rep := report{Files: files, Findings: findings, Mirror: mirror}
	if *asJSON {
		enc := json.NewEncoder(os.Stdout)
		enc.SetIndent("", "  ")
		if err := enc.Encode(rep); err != nil {
			return fmt.Errorf("报告编码失败: %w", err)
		}
	} else {
		for _, f := range findings {
			fmt.Printf("VIOLATION %s %s: %s\n", f.Path, f.Code, f.Detail)
		}
		fmt.Printf("cored-lint: files=%d findings=%d mirror=%s\n", files, len(findings), mirror)
	}

	if len(findings) > 0 {
		os.Exit(1)
	}
	return nil
}
