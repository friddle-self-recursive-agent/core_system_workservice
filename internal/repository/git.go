// git.go —— Repository 的 v1 git 后端：目录即存储（instance 仓 layout.md），
// 串行写队列模拟事务、resourceVersion 文档级乐观锁、幂等重入 no-op、git commit 留痕。
// 设计依据：
//   - concepts/core.md §7：v1 约束 = git 后端 + 单写者串行；Repository 接口留换 DB 的缝；
//   - concepts/revision.md：resourceVersion 每次成功写入自增、不符=Conflict；
//     机器只做整文档文本比对（相同 no-op / 不同提交）；commit message=<fieldManager> <action> <reason-ref>；
//   - concepts/scope.md 四层 scope 物化为目录前缀；core/scheme.yaml 各 kind 的 storage 模板定路径。
//
// 本文件不含部署锚点：文档库根目录经参数注入，rule.md R-PUB-1/2。
package repository

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"gopkg.in/yaml.v3"
)

// 对象存储哨兵错误（查询面按此映射 HTTP 状态码）。
var (
	// ErrNotFound 对象不存在（或声明了 resourceVersion 但对象已不在——并发删除语义）。
	ErrNotFound = errors.New("repository: 对象不存在")
	// ErrNotRegistered kind 未注册（Scheme 闸门：野对象拒绝，concepts/core.md）。
	ErrNotRegistered = errors.New("repository: kind 未注册")
	// ErrScope scope 与该 kind 的 storage 模板不相容（缺系统段/instance 段、或全部分支不可解析）。
	ErrScope = errors.New("repository: scope 与 storage 模板不相容")
	// ErrContent 内容不可解析为文档对象（缺 metadata 区 / 非法 YAML）。
	ErrContent = errors.New("repository: 内容缺 metadata 或不可解析")
	// ErrIdentity 内容与对象键不符（metadata.name/scope 与 (name, scope) 参数矛盾）。
	ErrIdentity = errors.New("repository: 内容身份与对象键不符")
	// ErrAlreadyExists 无版本盲写覆盖已存在的对象（先读 rv 再写）。
	ErrAlreadyExists = errors.New("repository: 对象已存在（需携带读到的 resourceVersion）")
	// ErrNoFieldManager 无主写入拒绝（concepts/revision.md 规则 1）。
	ErrNoFieldManager = errors.New("repository: 写入必须携带 fieldManager")
	// ErrClosed 写队列已关停（进程退场后拒收新写）。
	ErrClosed = errors.New("repository: 写队列已关停")
)

// ConflictError 乐观锁冲突：Expected=写者读到的版本，Current=库内现版本。
type ConflictError struct {
	Kind, Scope, Name string
	Expected, Current int64
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("repository: 乐观锁冲突 %s/%s/%s：读到 %d，现库 %d",
		e.Kind, e.Scope, e.Name, e.Expected, e.Current)
}

// ─────────────────────────────────────────────────────────────────────────────
// 构造与生命周期
// ─────────────────────────────────────────────────────────────────────────────

// Options git 后端可调项（零值可用；测试注入小间隔/假 ops store）。
type Options struct {
	// WatchInterval Watch 轮询 HEAD 的间隔（缺省 500ms；concepts/core.md「轮询 diff 模拟 watch」）。
	WatchInterval time.Duration
	// Ops ops DB 双写钩子（apply_log，concepts/revision.md 双轨审计）。nil=跳过（opslog sqlite 尚未落地）。
	Ops opslog.Store
}

// GitRepo 文档库 git 后端。一个实例对应一个文档库根目录；写路径全部经
// 单 goroutine 串行队列（并发调用安全，落库顺序=入队顺序）。
type GitRepo struct {
	root string
	sch  *scheme.Scheme
	opts Options

	reqs chan *writeReq // 串行写队列
	stop chan struct{}  // 关停信号（Close 触发；队列通道不关闭，消灭 send-on-closed）
	once sync.Once      // 关停幂等
	done chan struct{}  // 处理器退出信号

	gitName  string // commit -c user.name
	gitEmail string // commit -c user.email
}

// NewGit 打开（必要时初始化）文档库。root 即布局根（layout.md §1 顶层布局）；
// 目录不存在则创建；无 .git 则 git init（scheme.yaml runtime.storage：本地文件夹 git=权威）。
func NewGit(root string, sch *scheme.Scheme, opts *Options) (*GitRepo, error) {
	if sch == nil {
		return nil, errors.New("repository: 需 scheme 注册表（野对象闸门）")
	}
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, fmt.Errorf("repository: 根目录 %q: %w", root, err)
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, fmt.Errorf("repository: 创建根目录 %s: %w", abs, err)
	}
	r := &GitRepo{
		root:     abs,
		sch:      sch,
		reqs:     make(chan *writeReq),
		stop:     make(chan struct{}),
		done:     make(chan struct{}),
		gitName:  "cored",
		gitEmail: "cored@localhost",
	}
	if opts != nil {
		r.opts = *opts
	}
	if r.opts.WatchInterval <= 0 {
		r.opts.WatchInterval = 500 * time.Millisecond
	}
	if err := r.ensureGitRepo(); err != nil {
		return nil, err
	}
	go r.processWrites()
	return r, nil
}

// Close 关停写队列（幂等）：排空在途请求后处理器退出；此后新写返回 ErrClosed。
// Watch 频道由其 ctx 自行收尾。
func (r *GitRepo) Close() error {
	r.once.Do(func() { close(r.stop) })
	<-r.done
	return nil
}

// Root 文档库根目录（部署诊断用）。
func (r *GitRepo) Root() string { return r.root }

func (r *GitRepo) ensureGitRepo() error {
	if _, err := r.git(context.Background(), "rev-parse", "--git-dir"); err == nil {
		return nil
	}
	if _, err := r.git(context.Background(), "init", "-q"); err != nil {
		return fmt.Errorf("repository: git init %s: %w", r.root, err)
	}
	return nil
}

// git 在文档库执行 git 子命令（统一注入提交身份，不依赖宿主全局配置）。
func (r *GitRepo) git(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = r.root
	env := append(os.Environ(),
		"GIT_AUTHOR_NAME="+r.gitName, "GIT_AUTHOR_EMAIL="+r.gitEmail,
		"GIT_COMMITTER_NAME="+r.gitName, "GIT_COMMITTER_EMAIL="+r.gitEmail,
	)
	cmd.Env = env
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("git %s: %w（%s）", strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// storage 模板解析（scheme.yaml 各 kind 的 storage 字段 → 具体路径 / 通配模式）
// ─────────────────────────────────────────────────────────────────────────────

// nameSeg 对象名/scope 段的合法字符：layout.md §2 命名规则（小写中划线；容忍点号）。
var nameSeg = regexp.MustCompile(`^[a-z0-9]([a-z0-9.-]*[a-z0-9])?$`)

// scopeDir 把 scope（四层：builtin / global / <system> / <system>/<instance>，concepts/scope.md）
// 物化为目录前缀（layout.md：builtin→core/，global→global/，系统→systems/<s>/，实例→…/instances/<i>/）。
// scope="" 返回空（仅通配枚举语义）。
func scopeDir(scope string) (string, error) {
	switch scope {
	case "":
		return "", nil
	case "builtin":
		return "core", nil
	case "global":
		return "global", nil
	}
	segs := strings.Split(scope, "/")
	if len(segs) > 2 {
		return "", fmt.Errorf("%w: %q（scope 至多 系统/实例 两段）", ErrScope, scope)
	}
	for _, s := range segs {
		if !nameSeg.MatchString(s) {
			return "", fmt.Errorf("%w: 段 %q 不合命名规则（小写中划线）", ErrScope, s)
		}
	}
	dir := "systems/" + segs[0]
	if len(segs) == 2 {
		dir += "/instances/" + segs[1]
	}
	return dir, nil
}

// scopePrefix scope 对应的目录前缀（路径一致性闸门：模板展开结果必须落在本 scope 的目录之下）。
// scope="" 返回空串（跨 scope 通配，前缀不设限）。
func scopePrefix(scope string) (string, error) {
	if scope == "" {
		return "", nil
	}
	dir, err := scopeDir(scope)
	if err != nil {
		return "", err
	}
	return dir + "/", nil
}

// storagePatterns 解析 kind 的 storage 模板为路径模式列表。
//   - concrete（Get/Write）：scope/name 必须给全，返回恰一个具体相对路径；
//   - 通配（List）：未给定的段落展开为 `*`，返回全部可寻址模式。
//
// 模板分支有效性（scheme.yaml 个别条目混有范围词/散文，运行层按此闸门跳过）：
//   - 以 .yaml 结尾，且含 `<name>` 或完全不含占位符（固定单例路径，如 core/scheme.yaml）；
//   - 展开后必须以 scope 前缀开头（global 的对象不落 systems/，反之亦然）；
//   - 具体寻址下展开结果不得残留 `*` 或未识别占位符。
func storagePatterns(k scheme.Kind, scope, name string, concrete bool) ([]string, error) {
	segs := strings.Split(scope, "/")
	prefix, err := scopePrefix(scope)
	if err != nil {
		return nil, err
	}
	var patterns []string
	skipped := 0
	for alt := range strings.SplitSeq(k.Storage, "|") {
		alt = strings.TrimSpace(alt)
		if !validPathTemplate(alt) {
			skipped++
			continue
		}
		p, ok := expandTemplate(alt, segs, name, concrete)
		if !ok || !strings.HasPrefix(p, prefix) || strings.Contains(p, "*") && concrete {
			skipped++
			continue
		}
		patterns = append(patterns, p)
	}
	if len(patterns) == 0 {
		return nil, fmt.Errorf("%w: kind %q storage=%q 对 scope=%q name=%q 无可寻址分支（跳过 %d 个）",
			ErrScope, k.Name, k.Storage, scope, name, skipped)
	}
	sort.Strings(patterns)
	return patterns, nil
}

// validPathTemplate 分支是否为可寻址的路径模板。
func validPathTemplate(alt string) bool {
	if alt == "" || !strings.HasSuffix(alt, ".yaml") {
		return false
	}
	return strings.Contains(alt, "<name>") || !strings.Contains(alt, "<")
}

// expandTemplate 单分支占位符展开；不可满足返回 ok=false（如模板需系统段而 scope=global）。
// `…` 通配 = scope 目录相对 systems/<s>/ 的余部（空则整段收敛），兼容旧式模板。
func expandTemplate(alt string, segs []string, name string, concrete bool) (string, bool) {
	sys, inst := "", ""
	if len(segs) > 0 {
		sys = segs[0]
	}
	if len(segs) > 1 {
		inst = segs[1]
	}
	validSeg := func(v string) bool { return v == "*" || nameSeg.MatchString(v) }
	out := alt

	// <s>：系统段。builtin/global 作用域没有系统段——需要 <s> 的分支对该 scope 不可用。
	if strings.Contains(out, "<s>") {
		switch {
		case scopeIsEmpty(segs):
			out = strings.ReplaceAll(out, "<s>", "*") // List 跨 scope 通配
		case sys == "builtin" || sys == "global":
			return "", false
		case !validSeg(sys):
			return "", false
		default:
			out = strings.ReplaceAll(out, "<s>", sys)
		}
	}

	// <i>：实例段（scope 第二段；缺省时 List 通配、具体寻址不可用）。
	if strings.Contains(out, "<i>") {
		switch {
		case inst == "":
			if concrete {
				return "", false
			}
			out = strings.ReplaceAll(out, "<i>", "*")
		case !validSeg(inst):
			return "", false
		default:
			out = strings.ReplaceAll(out, "<i>", inst)
		}
	}

	// `…`：scope 目录余部（systems/<s>/ 之后的 instances/<i>/ 或空）。
	if strings.Contains(out, "…") {
		rest := ""
		if inst != "" {
			rest = "instances/" + inst + "/"
		}
		out = strings.ReplaceAll(out, "…/", rest)
		out = strings.ReplaceAll(out, "…", rest)
	}

	// <name>：对象名。空名仅在通配枚举下合法。
	if strings.Contains(out, "<name>") {
		if name == "" {
			if concrete {
				return "", false
			}
			name = "*"
		}
		out = strings.ReplaceAll(out, "<name>", name)
	}

	if strings.Contains(out, "<") || strings.Contains(out, ">") {
		return "", false // 仍有未识别占位符
	}
	return filepath.ToSlash(filepath.Clean(out)), true
}

func scopeIsEmpty(segs []string) bool {
	return len(segs) == 0 || (len(segs) == 1 && segs[0] == "")
}

// concretePath 解析出唯一具体路径并做越界防护（scope 四层校验+对象名命名规则）。
func (r *GitRepo) concretePath(k scheme.Kind, scope, name string) (string, error) {
	if scope == "" {
		return "", fmt.Errorf("%w: 写入/读取 %s 需显式 scope", ErrScope, k.Name)
	}
	if _, err := scopeDir(scope); err != nil {
		return "", err
	}
	if !nameSeg.MatchString(name) {
		return "", fmt.Errorf("%w: 对象名 %q 不合命名规则（小写中划线）", ErrScope, name)
	}
	pats, err := storagePatterns(k, scope, name, true)
	if err != nil {
		return "", err
	}
	if len(pats) != 1 {
		return "", fmt.Errorf("%w: kind %q scope=%q 解析出 %d 条路径（要恰一条）", ErrScope, k.Name, scope, len(pats))
	}
	rel := pats[0]
	if filepath.IsAbs(rel) || strings.HasPrefix(rel, "..") {
		return "", fmt.Errorf("%w: 解析路径越界 %q", ErrScope, rel)
	}
	return filepath.Join(r.root, filepath.FromSlash(rel)), nil
}

// ─────────────────────────────────────────────────────────────────────────────
// 文档解析与 resourceVersion 注入（yaml / md+frontmatter 同构）
// ─────────────────────────────────────────────────────────────────────────────

type docMeta struct {
	Name  string
	Scope string
	RV    int64
}

// splitDoc 拆出文档与元数据容器：前导 `---` 围栏=md frontmatter（扁平键，layout.md §3：
// name/docType/scope 直接住顶层），否则整文=yaml（身份字段住 `metadata:` 区，scheme.yaml conventions）。
// 返回文档映射节点（保注释/顺序）、元数据容器节点、正文与形态标记。
func splitDoc(raw []byte) (doc, container *yaml.Node, body []byte, isMD bool, err error) {
	text := string(raw)
	if strings.HasPrefix(text, "---") {
		lines := strings.Split(text, "\n")
		if strings.TrimSpace(lines[0]) != "---" {
			return nil, nil, nil, false, fmt.Errorf("%w: md 首行须为 --- 围栏", ErrContent)
		}
		closeIdx := -1
		for i := 1; i < len(lines); i++ {
			if strings.TrimSpace(lines[i]) == "---" {
				closeIdx = i
				break
			}
		}
		if closeIdx < 0 {
			return nil, nil, nil, false, fmt.Errorf("%w: md frontmatter 未闭合", ErrContent)
		}
		front := strings.Join(lines[1:closeIdx], "\n")
		body = []byte(strings.Join(lines[closeIdx+1:], "\n"))
		var node yaml.Node
		if err := yaml.Unmarshal([]byte(front), &node); err != nil {
			return nil, nil, nil, false, fmt.Errorf("%w: frontmatter 解析失败: %v", ErrContent, err)
		}
		if len(node.Content) == 0 || node.Content[0].Kind != yaml.MappingNode {
			return nil, nil, nil, false, fmt.Errorf("%w: frontmatter 须为映射", ErrContent)
		}
		doc = node.Content[0]
		return doc, doc, body, true, nil // md：容器=frontmatter 根（扁平）
	}
	var node yaml.Node
	if err := yaml.Unmarshal(raw, &node); err != nil {
		return nil, nil, nil, false, fmt.Errorf("%w: yaml 解析失败: %v", ErrContent, err)
	}
	if len(node.Content) == 0 || node.Content[0].Kind != yaml.MappingNode {
		return nil, nil, nil, false, fmt.Errorf("%w: 文档根须为映射", ErrContent)
	}
	doc = node.Content[0]
	container = mapValue(doc, "metadata")
	if container == nil || container.Kind != yaml.MappingNode {
		return nil, nil, nil, false, fmt.Errorf("%w: 缺 metadata 区", ErrContent)
	}
	return doc, container, raw, false, nil
}

// metaOf 从元数据容器提取 name/scope/resourceVersion。
func metaOf(container *yaml.Node) (docMeta, error) {
	var m docMeta
	if container == nil || container.Kind != yaml.MappingNode {
		return m, fmt.Errorf("%w: 缺 metadata 区", ErrContent)
	}
	if n := mapValue(container, "name"); n != nil && n.Kind == yaml.ScalarNode {
		m.Name = n.Value
	}
	if s := mapValue(container, "scope"); s != nil && s.Kind == yaml.ScalarNode {
		m.Scope = s.Value
	}
	if rv := mapValue(container, "resourceVersion"); rv != nil && rv.Kind == yaml.ScalarNode {
		var v int64
		if _, err := fmt.Sscanf(rv.Value, "%d", &v); err == nil {
			m.RV = v
		}
	}
	return m, nil
}

// mapValue 在映射节点里取键的值节点（yaml.v3 映射：Content 交替存放 key/value）。
func mapValue(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}

// setResourceVersion 就地注入/改写元数据容器中的 resourceVersion（保留既有键序与注释）。
func setResourceVersion(container *yaml.Node, rv int64) error {
	if container == nil || container.Kind != yaml.MappingNode {
		return fmt.Errorf("%w: 缺 metadata 区", ErrContent)
	}
	if v := mapValue(container, "resourceVersion"); v != nil {
		v.Value = fmt.Sprintf("%d", rv)
		v.Tag = "!!int"
		v.Style = 0
		return nil
	}
	k := &yaml.Node{Kind: yaml.ScalarNode, Value: "resourceVersion"}
	v := &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: fmt.Sprintf("%d", rv)}
	container.Content = append(container.Content, k, v)
	return nil
}

// encodeDoc 重编码文档：yaml 走节点树（保注释，格式规整为 2 空格缩进）；
// md 重排 frontmatter 围栏 + 原文正文。
func encodeDoc(doc *yaml.Node, body []byte, isMD bool) ([]byte, error) {
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return nil, fmt.Errorf("repository: 文档编码失败: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, err
	}
	out := buf.Bytes()
	if isMD {
		var merged bytes.Buffer
		merged.WriteString("---\n")
		merged.Write(out)
		merged.WriteString("---\n")
		merged.Write(body)
		out = merged.Bytes()
	}
	return out, nil
}

// canonicalWithoutRV 把原文规整为"去掉 resourceVersion"的规范文本（同一编码器产出），
// 幂等判定 = 两份规范文本相等（concepts/revision.md：机器只做整文档文本比对）。
func canonicalWithoutRV(raw []byte) (string, error) {
	doc, container, body, isMD, err := splitDoc(raw)
	if err != nil {
		return "", err
	}
	_ = doc
	removePair(container, "resourceVersion")
	out, err := encodeDoc(doc, body, isMD)
	if err != nil {
		return "", err
	}
	return string(out), nil
}

func removePair(mapping *yaml.Node, key string) {
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			mapping.Content = append(mapping.Content[:i], mapping.Content[i+2:]...)
			return
		}
	}
}

// ─────────────────────────────────────────────────────────────────────────────
// 读路径：Get / List
// ─────────────────────────────────────────────────────────────────────────────

// Get 读取单个文档对象。身份以文档内 metadata 为权威（Object.Name 优先取文档值）；
// 只读面宽松：文档身份与请求键不一致不拒绝（一致性强校验在写路径）。
func (r *GitRepo) Get(_ context.Context, kind, scope, name string) (*Object, error) {
	k, err := r.sch.Lookup(kind)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotRegistered, err)
	}
	path, err := r.concretePath(k, scope, name)
	if err != nil {
		return nil, err
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return nil, fmt.Errorf("%w: %s/%s/%s", ErrNotFound, kind, scope, name)
		}
		return nil, fmt.Errorf("repository: 读取 %s: %w", path, err)
	}
	_, container, _, _, err := splitDoc(raw)
	if err != nil {
		return nil, err
	}
	m, err := metaOf(container)
	if err != nil {
		return nil, err
	}
	objName := m.Name
	if objName == "" {
		objName = name
	}
	objScope := m.Scope
	if objScope == "" {
		objScope = scope
	}
	return &Object{Kind: kind, Name: objName, Scope: objScope, ResourceVersion: m.RV, Content: raw}, nil
}

// List 枚举对象：scope 给定则限定前缀，scope="" 跨 scope 通配（模板可展开的全部分支）。
// 不可解析/损坏的文件跳过不致命（只读面宽松；写路径才强校验）。
func (r *GitRepo) List(_ context.Context, kind, scope string) ([]*Object, error) {
	k, err := r.sch.Lookup(kind)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", ErrNotRegistered, err)
	}
	pats, err := storagePatterns(k, scope, "", false)
	if err != nil {
		return nil, err
	}
	seen := map[string]bool{}
	var out []*Object
	for _, pat := range pats {
		matches, _ := filepath.Glob(filepath.Join(r.root, filepath.FromSlash(pat)))
		sort.Strings(matches)
		for _, path := range matches {
			if seen[path] {
				continue
			}
			seen[path] = true
			raw, err := os.ReadFile(path)
			if err != nil {
				continue
			}
			doc, container, _, _, err := splitDoc(raw)
			if err != nil {
				continue
			}
			_ = doc
			m, err := metaOf(container)
			if err != nil {
				continue
			}
			objName := m.Name
			if objName == "" {
				objName = strings.TrimSuffix(filepath.Base(path), filepath.Ext(path))
			}
			objScope := m.Scope
			if objScope == "" {
				objScope = scope
			}
			out = append(out, &Object{Kind: kind, Name: objName, Scope: objScope, ResourceVersion: m.RV, Content: raw})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Scope != out[j].Scope {
			return out[i].Scope < out[j].Scope
		}
		return out[i].Name < out[j].Name
	})
	return out, nil
}

// ─────────────────────────────────────────────────────────────────────────────
// 写路径：串行队列 + 乐观锁 + 幂等 + git commit
// ─────────────────────────────────────────────────────────────────────────────

type writeReq struct {
	o      *Object
	fm     string
	reason string
	ctx    context.Context
	resp   chan writeResp
}

type writeResp struct {
	rv     int64
	action string
	err    error
}

// Write 提交写入（并发安全：全部经单 goroutine 串行执行）。
// 语义（concepts/revision.md）：
//   - fm 为空拒绝；kind 未注册拒绝；scope 必填；
//   - 创建：rv=0，成功后 resourceVersion=1；
//   - 更新：携带读到的 rv，不符=ConflictError；
//   - 幂等重入：内容与库内一致（去 rv 比对）→ no-op，返回现 rv，不产生提交；
//   - 已存在但 rv=0 盲写 → ErrAlreadyExists。
func (r *GitRepo) Write(ctx context.Context, o *Object, fieldManager, reasonRef string) (int64, error) {
	if fieldManager == "" {
		return 0, ErrNoFieldManager
	}
	if o == nil || o.Content == nil {
		return 0, fmt.Errorf("%w: 空内容", ErrContent)
	}
	req := &writeReq{
		o: &Object{Kind: o.Kind, Name: o.Name, Scope: o.Scope, ResourceVersion: o.ResourceVersion,
			Content: append([]byte(nil), o.Content...)},
		fm:     fieldManager,
		reason: reasonRef,
		ctx:    ctx,
		resp:   make(chan writeResp, 1),
	}
	select {
	case r.reqs <- req:
	case <-r.stop:
		return 0, ErrClosed
	case <-ctx.Done():
		return 0, ctx.Err()
	}
	select {
	case resp := <-req.resp:
		return resp.rv, resp.err
	case <-ctx.Done():
		return 0, ctx.Err()
	}
}

// processWrites 单写者循环：git 目录无事务，队列即事务边界（concepts/core.md §7）。
// stop 触发后排空在途请求再退出（已入队者不丢）。
func (r *GitRepo) processWrites() {
	defer close(r.done)
	drain := func() (empty bool) {
		select {
		case req := <-r.reqs:
			r.serve(req)
			return false
		default:
			return true
		}
	}
	for {
		select {
		case req := <-r.reqs:
			r.serve(req)
		case <-r.stop:
			for !drain() {
			}
			return
		}
	}
}

func (r *GitRepo) serve(req *writeReq) {
	rv, action, err := r.applyWrite(req)
	select {
	case req.resp <- writeResp{rv: rv, action: action, err: err}:
	case <-req.ctx.Done():
	}
}

func (r *GitRepo) applyWrite(req *writeReq) (int64, string, error) {
	// 出队即事务：git 阶段脱离调用方 ctx（半途 kill 会留 .git/index.lock 拖垮后续写）；
	// 取消语义=排队前与出队时可弃单，落盘开始后不可中断。
	ctx := context.WithoutCancel(req.ctx)
	o := req.o
	if err := req.ctx.Err(); err != nil {
		return 0, "", err
	}
	k, err := r.sch.Lookup(o.Kind)
	if err != nil {
		return 0, "", fmt.Errorf("%w: %v", ErrNotRegistered, err)
	}
	path, err := r.concretePath(k, o.Scope, o.Name)
	if err != nil {
		return 0, "", err
	}
	// 入向文档校验：可解析 + 身份一致。
	inDoc, inContainer, inBody, inMD, err := splitDoc(o.Content)
	if err != nil {
		return 0, "", err
	}
	inMeta, err := metaOf(inContainer)
	if err != nil {
		return 0, "", err
	}
	if inMeta.Name != "" && inMeta.Name != o.Name {
		return 0, "", fmt.Errorf("%w: metadata.name=%q 与对象名 %q 不符", ErrIdentity, inMeta.Name, o.Name)
	}
	if inMeta.Scope != "" && inMeta.Scope != o.Scope {
		return 0, "", fmt.Errorf("%w: metadata.scope=%q 与 scope %q 不符", ErrIdentity, inMeta.Scope, o.Scope)
	}
	inCanon, err := canonicalWithoutRV(o.Content)
	if err != nil {
		return 0, "", err
	}

	action := "update"
	newRV := int64(0)
	storedRaw, readErr := os.ReadFile(path)
	switch {
	case readErr == nil:
		_, storedContainer, _, _, err := splitDoc(storedRaw)
		if err != nil {
			return 0, "", err
		}
		storedMeta, err := metaOf(storedContainer)
		if err != nil {
			return 0, "", err
		}
		storedCanon, err := canonicalWithoutRV(storedRaw)
		if err != nil {
			return 0, "", err
		}
		if inCanon == storedCanon {
			// 幂等重入：内容一致（去 rv 比对）→ no-op，不产生提交。
			r.logApply(o, storedMeta.RV, "noop", req)
			return storedMeta.RV, "noop", nil
		}
		if o.ResourceVersion == 0 {
			return 0, "", fmt.Errorf("%w: %s/%s/%s 现版本 %d", ErrAlreadyExists, o.Kind, o.Scope, o.Name, storedMeta.RV)
		}
		if o.ResourceVersion != storedMeta.RV {
			return 0, "", &ConflictError{Kind: o.Kind, Scope: o.Scope, Name: o.Name,
				Expected: o.ResourceVersion, Current: storedMeta.RV}
		}
		newRV = storedMeta.RV + 1
	case errors.Is(readErr, fs.ErrNotExist):
		if o.ResourceVersion != 0 {
			// 写者读到过它、如今已不在：并发删除，按冲突处理。
			return 0, "", &ConflictError{Kind: o.Kind, Scope: o.Scope, Name: o.Name,
				Expected: o.ResourceVersion, Current: 0}
		}
		action = "create"
		newRV = 1
	default:
		return 0, "", fmt.Errorf("repository: 读取 %s: %w", path, readErr)
	}

	// 注入 resourceVersion 后落盘（原子写：临时文件+rename）。
	if err := setResourceVersion(inContainer, newRV); err != nil {
		return 0, "", err
	}
	out, err := encodeDoc(inDoc, inBody, inMD)
	if err != nil {
		return 0, "", err
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return 0, "", fmt.Errorf("repository: 建目录 %s: %w", filepath.Dir(path), err)
	}
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, out, 0o644); err != nil {
		return 0, "", fmt.Errorf("repository: 写临时文件 %s: %w", tmp, err)
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return 0, "", fmt.Errorf("repository: 落盘 %s: %w", path, err)
	}

	// git 留痕：内容版本权威（concepts/revision.md §3）。
	key := fmt.Sprintf("%s/%s/%s", o.Kind, o.Scope, o.Name)
	msg := fmt.Sprintf("%s %s %s", req.fm, action, key)
	if req.reason != "" {
		msg += " " + req.reason
	}
	if _, err := r.git(ctx, "add", "-A", "--", relToRoot(r.root, path)); err != nil {
		return 0, "", err
	}
	if _, err := r.git(ctx, "-c", "commit.gpgsign=false", "commit", "-m", msg); err != nil {
		// 空提交（内容与暂存一致）不视为失败。
		if !strings.Contains(err.Error(), "nothing to commit") {
			return 0, "", err
		}
	}
	r.logApply(o, newRV, "ok", req)
	return newRV, action, nil
}

func (r *GitRepo) logApply(o *Object, rv int64, result string, req *writeReq) {
	if r.opts.Ops == nil {
		return
	}
	_ = r.opts.Ops.AppendApply(opslog.ApplyLog{
		Ts: time.Now(), FieldManager: req.fm, Obj: fmt.Sprintf("%s/%s/%s", o.Kind, o.Scope, o.Name),
		ResourceVersion: rv, Result: result, ReasonRef: req.reason,
	})
}

func relToRoot(root, path string) string {
	rel, err := filepath.Rel(root, path)
	if err != nil {
		return path
	}
	return filepath.ToSlash(rel)
}

// ─────────────────────────────────────────────────────────────────────────────
// Watch：轮询 HEAD 变更（v1 兜底；git hook/换 DB 后端时再升级）
// ─────────────────────────────────────────────────────────────────────────────

// Watch 订阅库变更：先发当前 HEAD（无提交则不发），其后每次 HEAD 前移发新 sha。
// ctx 取消即关停（频道关闭）。git 暂态错误（如 commit 进行中）静默跳过、下个周期重试。
func (r *GitRepo) Watch(ctx context.Context) (<-chan string, error) {
	ch := make(chan string, 8)
	go func() {
		defer close(ch)
		last := ""
		if head, err := r.git(ctx, "rev-parse", "HEAD"); err == nil {
			last = head
			sendChange(ctx, ch, last)
		}
		tk := time.NewTicker(r.opts.WatchInterval)
		defer tk.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-tk.C:
				head, err := r.git(ctx, "rev-parse", "HEAD")
				if err != nil || head == last {
					continue
				}
				last = head
				sendChange(ctx, ch, last)
			}
		}
	}()
	return ch, nil
}

func sendChange(ctx context.Context, ch chan string, v string) {
	select {
	case ch <- v:
	case <-ctx.Done():
	default: // 慢消费者：丢旧保新（watch 是通知不是事件总线，eventbus 另有去重）
		select {
		case <-ch:
		default:
		}
		select {
		case ch <- v:
		case <-ctx.Done():
		}
	}
}
