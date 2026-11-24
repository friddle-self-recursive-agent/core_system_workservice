// lint.go —— docsRoot 存量对象净入 lint（scheme.md 规则 6 checker 应用点②，v0.48）。
// 缺口根因：spec 准入 checker（scheme.ValidateDoc）应用点单一=apply 引擎准入第三步，只挡写入口；
// docsRoot 存量对象零校验放行——读面 globObjects 对不可解析文档静默跳过（存量坏对象在 List 读面
// 不可见），存量形态违规对象读面放行、一编辑即 400 且按原文无法复原（读面宽松/写面严格双严格度劈叉）。
//
// 本面=checker 的存量面应用点：docsRoot 全 yaml 文件系统通配（不经 List——静默丢弃面正是被补的
// 缺口），逐文件五闸与写路径同清单（applyWrite + apply 引擎准入 + 凭据闸门逐项镜像）：
//   ① 路径注册：storage 模板分支命中恰一 kind（零命中=野对象；多命中=歧义）；
//   ② 可解析：splitDoc+metaOf（读面静默丢弃面补检出——globObjects 的 continue 在此显形）；
//   ③ 身份一致：metadata.name/scope vs 模板 <name> 段捕获与路径派生 scope（写面 applyWrite 同闸）+
//      顶层 kind 字段（写面盲区：对象 kind 来自路由而非文档，漂移零检出——lint 补齐）+
//      nameSeg 命名规则（写面 concretePath 同闸）；
//   ④ spec 准入：scheme.ValidateDoc（apply 引擎准入第三步同闸门，cored-bridge.md §2.1）；
//   ⑤ 凭据来源形态：checkCredentialSource（Provider 写面同闸，provider.md 规则 2）。
// scheme 镜像漂移闸（docsRoot core/scheme.yaml 消费方镜像 vs 载入权威字节比对）独立函数：
// LintSchemeMirror——双源系漂移零检出缺口收口。
//
// 纪律：lint 纯只读（零写面零 ops DB 零 HTTP）；判定只机械可验，判不准的不判。

package repository

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"gopkg.in/yaml.v3"
)

// Finding 一条净入 lint 发现（path=docsRoot 相对路径，slash 分隔）。
// code 闭合枚举：unregistered-path | ambiguous-path | unparseable | identity | spec | credential | scheme-drift。
type Finding struct {
	Path   string `json:"path"`
	Code   string `json:"code"`
	Detail string `json:"detail"`
}

func (f Finding) String() string {
	return fmt.Sprintf("%s %s: %s", f.Code, f.Path, f.Detail)
}

// branchMatcher 单个 storage 模板分支的净入匹配器：正则锚定 rel 路径并捕获 <name> 段。
// （与 storagePatterns 的通配展开同源口径——<s>/<i>/`…` 通配、固定单例分支原样、
// <name> 嵌段（doctypes/<name>.yaml / <name> 独段 global/agents/<name>/agent.yaml）皆捕获。）
type branchMatcher struct {
	kindName string
	re       *regexp.Regexp
	hasName  bool
}

func compileBranches(k scheme.Kind) ([]branchMatcher, error) {
	var out []branchMatcher
	for alt := range strings.SplitSeq(k.Storage, "|") {
		alt = strings.TrimSpace(alt)
		if !validPathTemplate(alt) {
			continue // 与 storagePatterns 同口径：不可寻址分支跳过不误报
		}
		var b strings.Builder
		b.WriteString("^")
		hasName := false
		ok := true
		segIdx := 0
		for seg := range strings.SplitSeq(alt, "/") {
			if segIdx > 0 {
				b.WriteString("/")
			}
			segIdx++
			if seg == "" {
				continue
			}
			if strings.Contains(seg, "<name>") {
				prefix, suffix, found := strings.Cut(seg, "<name>")
				if !found || strings.Contains(prefix, "<") || strings.Contains(suffix, "<") {
					ok = false
					break
				}
				b.WriteString(regexp.QuoteMeta(prefix))
				b.WriteString(`([^/]+)`)
				b.WriteString(regexp.QuoteMeta(suffix))
				hasName = true
				continue
			}
			if strings.Contains(seg, "<") || strings.Contains(seg, "…") {
				// <s>/<i>/`…`：跨 scope 通配段（scope 段不捕获——身份闸走路径派生 scope）。
				if strings.Contains(seg, "…") {
					// `…`=scope 目录余部：空 inst 时收敛为空（`…/` 整段收敛），多段通配。
					if seg == "…/" || seg == "…" {
						b.WriteString(`(?:[^/]+(?:/[^/]+)*)?`)
					} else {
						ok = false
						break
					}
					continue
				}
				switch {
				case strings.Contains(seg, "<s>") || strings.Contains(seg, "<i>"):
					if seg != "<s>" && seg != "<i>" && seg != "<s>/<i>" {
						ok = false // 混嵌占位符段不可机读——与 validPathTemplate 同口径跳过
						break
					}
					n := strings.Count(seg, "/") + 1
					for i := 0; i < n; i++ {
						if i > 0 {
							b.WriteString("/")
						}
						b.WriteString(`[^/]+`)
					}
				default:
					ok = false // 未识别占位符
				}
				if !ok {
					break
				}
				continue
			}
			b.WriteString(regexp.QuoteMeta(seg))
		}
		if !ok {
			continue
		}
		b.WriteString("$")
		re, err := regexp.Compile(b.String())
		if err != nil {
			return nil, fmt.Errorf("repository: lint 模板分支 %q 编译失败: %w", alt, err)
		}
		out = append(out, branchMatcher{kindName: k.Name, re: re, hasName: hasName})
	}
	return out, nil
}

// derivedScope 路径派生 scope（scopeDir 的镜像：目录前缀→scope 两段串）。
//
//	global/ → global；core/ → builtin；systems/<s>/instances/<i>/… → <s>/<i>；
//	systems/<s>/… → <s>；根级（doctypes/ 等 scope 外桶）→ ""（无 scope 段）。
func derivedScope(rel string) string {
	segs := strings.Split(rel, "/")
	switch {
	case segs[0] == "global":
		return "global"
	case segs[0] == "core":
		return "builtin"
	case segs[0] == "systems" && len(segs) >= 4 && segs[2] == "instances":
		return segs[1] + "/" + segs[3]
	case segs[0] == "systems" && len(segs) >= 2:
		return segs[1]
	}
	return ""
}

// LintDocs 净入 lint 引擎：返回全部发现（按 path+code 排序，机械对账锚点）与对象文件计数
// （扩展名 .yaml/.yml 的文件数——含违规者，cmd 汇总「files=」行用）。
// err 仅承载 docsRoot 本身不可用（目录缺失等），不承载单文件问题——单文件问题一律 Finding。
func LintDocs(root string, sch *scheme.Scheme) (findings []Finding, files int, err error) {
	if sch == nil {
		return nil, 0, fmt.Errorf("repository: lint 需要 scheme")
	}
	var matchers []branchMatcher
	for _, k := range sch.ListKinds() {
		bs, berr := compileBranches(k)
		if berr != nil {
			return nil, 0, berr
		}
		matchers = append(matchers, bs...)
	}

	walkErr := filepath.WalkDir(root, func(path string, d fs.DirEntry, ferr error) error {
		if ferr != nil {
			return ferr
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return fs.SkipDir
			}
			return nil
		}
		ext := filepath.Ext(d.Name())
		if ext != ".yaml" && ext != ".yml" {
			return nil // 非对象文件（README/.gitignore 等）不在准入域
		}
		files++
		rel := filepath.ToSlash(strings.TrimPrefix(strings.TrimPrefix(path, root), "/"))

		// ① 路径注册：命中恰一 kind（同 kind 多分支命中=仍单 kind）。
		type hit struct {
			kindName string
			name     string
			hasName  bool
		}
		var hits []hit
		seenKind := map[string]bool{}
		for _, m := range matchers {
			sub := m.re.FindStringSubmatch(rel)
			if sub == nil {
				continue
			}
			name := ""
			if m.hasName {
				name = sub[1]
			}
			hits = append(hits, hit{m.kindName, name, m.hasName})
			seenKind[m.kindName] = true
		}
		if len(hits) == 0 {
			findings = append(findings, Finding{Path: rel, Code: "unregistered-path",
				Detail: "路径不落任何已注册 storage 模板（野对象——写面 ErrNotRegistered/ErrScope 同域）"})
			return nil
		}
		if len(seenKind) > 1 {
			kinds := make([]string, 0, len(seenKind))
			for kn := range seenKind {
				kinds = append(kinds, kn)
			}
			sort.Strings(kinds)
			findings = append(findings, Finding{Path: rel, Code: "ambiguous-path",
				Detail: "命中多个 kind：" + strings.Join(kinds, ", ")})
			return nil
		}
		h := hits[0]
		k, lerr := sch.Lookup(h.kindName)
		if lerr != nil {
			findings = append(findings, Finding{Path: rel, Code: "unregistered-path",
				Detail: fmt.Sprintf("kind %q 未注册: %v", h.kindName, lerr)})
			return nil
		}

		raw, rerr := os.ReadFile(path)
		if rerr != nil {
			findings = append(findings, Finding{Path: rel, Code: "unparseable",
				Detail: fmt.Sprintf("读取失败: %v", rerr)})
			return nil
		}

		// ② 可解析（读面静默丢弃面补检出）。
		doc, container, _, _, perr := splitDoc(raw)
		if perr != nil {
			findings = append(findings, Finding{Path: rel, Code: "unparseable",
				Detail: perr.Error() + "（globObjects 读面静默跳过=List 不可见）"})
			return nil
		}
		m, merr := metaOf(container)
		if merr != nil {
			findings = append(findings, Finding{Path: rel, Code: "unparseable", Detail: merr.Error()})
			return nil
		}

		// ③ 身份一致（写面 applyWrite 同闸 + 写面盲区补齐）。
		if h.hasName {
			if !nameSeg.MatchString(h.name) {
				findings = append(findings, Finding{Path: rel, Code: "identity",
					Detail: fmt.Sprintf("对象名 %q 不合命名规则（小写中划线——concretePath 同闸）", h.name)})
			}
			if m.Name != "" && m.Name != h.name {
				findings = append(findings, Finding{Path: rel, Code: "identity",
					Detail: fmt.Sprintf("metadata.name=%q 与对象名 %q 不符（applyWrite ErrIdentity 同闸）", m.Name, h.name)})
			}
			if m.Name == "" {
				findings = append(findings, Finding{Path: rel, Code: "identity",
					Detail: "metadata.name 缺失（读面回落=盲名对象，身份声明漂移）"})
			}
		}
		derived := derivedScope(rel)
		if m.Scope != "" && m.Scope != derived {
			findings = append(findings, Finding{Path: rel, Code: "identity",
				Detail: fmt.Sprintf("metadata.scope=%q 与物化路径派生 scope %q 不符（L44 迁移缺口同型：身份声明与物化位置漂移）", m.Scope, derived)})
		}
		if kd := mapValue(doc, "kind"); kd != nil && kd.Kind == yaml.ScalarNode && kd.Value != h.kindName {
			findings = append(findings, Finding{Path: rel, Code: "identity",
				Detail: fmt.Sprintf("文档 kind=%q 与路径所属 kind %q 不符（写面盲区：对象 kind 来自路由）", kd.Value, h.kindName)})
		}

		// ④ spec 准入（apply 引擎准入第三步同闸门）。
		if verr := scheme.ValidateDoc(k, raw); verr != nil {
			findings = append(findings, Finding{Path: rel, Code: "spec", Detail: verr.Error()})
		}

		// ⑤ 凭据来源形态（Provider 写面同闸）。
		if cerr := checkCredentialSource(h.kindName, doc); cerr != nil {
			findings = append(findings, Finding{Path: rel, Code: "credential", Detail: cerr.Error()})
		}
		return nil
	})
	if walkErr != nil {
		return nil, files, fmt.Errorf("repository: lint 遍历 %q: %w", root, walkErr)
	}
	sort.Slice(findings, func(i, j int) bool {
		if findings[i].Path != findings[j].Path {
			return findings[i].Path < findings[j].Path
		}
		return findings[i].Code < findings[j].Code
	})
	return findings, files, nil
}

// LintSchemeMirror scheme 镜像漂移闸：docsRoot 内 core/scheme.yaml（消费方镜像，Scheme 单例
// 物化位）与载入权威（--scheme 路径）字节比对。镜像缺席=不判（docsRoot 不必内嵌镜像）；
// 双方在册且不一致=漂移发现。返回 (镜像在册, 发现)。
func LintSchemeMirror(root string, authority []byte) (bool, *Finding, error) {
	raw, err := os.ReadFile(filepath.Join(root, "core", "scheme.yaml"))
	if err != nil {
		return false, nil, nil // 镜像缺席=不判
	}
	if string(raw) == string(authority) {
		return true, nil, nil
	}
	return true, &Finding{Path: "core/scheme.yaml", Code: "scheme-drift",
		Detail: "docsRoot 内 scheme 镜像与载入权威字节不一致（双源系漂移——消费方镜像须与 --scheme 载入源同步）"}, nil
}
