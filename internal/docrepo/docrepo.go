// Package docrepo 设计文档受限文件面（能力 A——2026-11-24 用户裁决「B 之后做 A」，
// 任务书 /home/friddle 外部不入仓：console 屏幕 6 直接编辑设计文档 md → Apply → git commit）。
//
// 路线裁决（任务书两路线择一，取舍入报告 deviations/设计回写）：**受限文件面（路线②）**。
// 设计文档正文的真相本来就在其 git 仓（core_design/concepts/*.md、core_system_instance/runtime/*.md），
// 对象化（路线① kind:DesignDoc 落 docsRoot）会在文档库造第二副本、再靠同步器写回目标仓——
// 单向权威可定义但多一跳同步与漂移判定面；本面零新权威：Apply=写工作树文件 + git add + git commit，
// 读面=git HEAD 内容（git ls-tree / cat-file——**未提交的工作树改动不进读面**，权威=git HEAD 唯一），
// 红线「权威必须唯一」的机械落点：无同步器、无双源、无漂移判定面。
//
// 闸门（任务书写面三检查，同 apply 管线口径）：
//  1. 白名单（--design-repo 注册仓 ∧ *.md 设计文档 ∧ 路径穿越拒绝）——越权路径 403；
//  2. R-PUB-1 lint（core_design lint/personal-info.sh 同清单的 Go 收编：IP/部署绝对路径/个人标识）——
//     未过 400 携违规路径（lint/personal-info.sh 是机械判定面，规则集两处同源同清单）；
//  3. 乐观锁（baseSha=打开时读到的 HEAD blob sha；与现 HEAD 不符 → 409 乐观锁冲突）。
//
// 高危判定（任务书 §5）：写设计文档=permission.md 机读判定（perm.Check(主体, core/design-docs, write)），
// admin 直通=DefaultAdmin 字面量既有语义；审计=apply_log 行 + DocApplied 事件（involved=
// DesignDoc/<repo>/<path>，三段同构惯例——与对象 apply 同一事件主干零旁路）。
//
// 红线（任务书）：不自动 push（commit 到本地仓即止）；不 enable workflow；幂等（同内容重入=noop
// 零提交零事件）；工作树有未提交改动=拒绝 Apply（不吞他人未提交编辑——ErrDirty）。
package docrepo

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
)

// 拒绝哨兵（读面按此映射状态码；与 apply/repository 哨兵词汇对齐）。
var (
	// ErrNoRepo 仓未注册（--design-repo 白名单外）→ 404。
	ErrNoRepo = errors.New("docrepo: 仓未注册（--design-repo 白名单外）")
	// ErrNotFound 文档不在 HEAD（读面权威=git HEAD）→ 404。
	ErrNotFound = errors.New("docrepo: 文档不存在（HEAD 无此文件）")
	// ErrForbidden 路径不在白名单（非注册仓 / 非 .md 设计文档 / 路径穿越）→ 403。
	ErrForbidden = errors.New("docrepo: 路径不在白名单（只读注册仓的 .md 设计文档）")
	// ErrDenied 权限准入拒绝（permission.md 机读判定）→ 403。
	ErrDenied = errors.New("docrepo: 权限准入拒绝")
	// ErrNoFieldManager 无主写入拒绝（apply.md rule 2 同语义）→ 400。
	ErrNoFieldManager = errors.New("docrepo: 写入必须携带 fieldManager（无主写入拒绝）")
	// ErrInvalid 参数不合法（repo/path 键不全）→ 400。
	ErrInvalid = errors.New("docrepo: 参数不合法（需 repo 与 path）")
	// ErrNoEngine 写面引擎未装配（--design-repo 有但 --allow-apply 缺）→ 405。
	ErrNoEngine = errors.New("docrepo: 写面未开启（cored 缺 --allow-apply）")
	// ErrDirty 工作树该文件有未提交改动（不吞他人未提交编辑）→ 409。
	ErrDirty = errors.New("docrepo: 工作树有未提交改动（先让在途编辑落提交后再 Apply）")
	// ErrGit git 操作失败（提交/读取）→ 500。
	ErrGit = errors.New("docrepo: git 操作失败")
)

// ConflictError 乐观锁冲突（baseSha 与现 HEAD blob sha 不符）→ 409（载荷携 expected/current——
// 与 apply 管线 repository.ConflictError 同形）。
type ConflictError struct {
	Repo, Path        string
	Expected, Current string // 期望基线 sha / 现库 HEAD sha（空串=HEAD 无此文件）
}

func (e *ConflictError) Error() string {
	return fmt.Sprintf("docrepo: 基线 sha 与仓内 HEAD 不一致（%s/%s：基线 %q ↔ 现 HEAD %q）——重读后重整再落",
		e.Repo, e.Path, e.Expected, e.Current)
}

// LintError R-PUB-1 lint 未过 → 400（载荷携违规清单——任务书判据「lint 失败 → 400 含违规路径」）。
type LintError struct {
	Violations []Violation
}

func (e *LintError) Error() string {
	return fmt.Sprintf("docrepo: R-PUB-1 lint 未过（%d 处违规——公开仓不暴露 IP/部署路径/个人标识）", len(e.Violations))
}

// Violation lint 违规一行。
type Violation struct {
	Line int    `json:"line"` // 1 起行号
	Rule string `json:"rule"` // ip | path | ident（与 lint/personal-info.sh 三清单同源）
	Text string `json:"text"` // 命中文本（截断）
}

// ─────────────────────────────────────────────────────────────────────────────
// R-PUB-1 lint（core_design lint/personal-info.sh 清单的 Go 收编——机械判定面同源）
// ─────────────────────────────────────────────────────────────────────────────

var (
	lintIPRe     = regexp.MustCompile(`([0-9]{1,3}\.){3}[0-9]{1,3}`)
	lintIPAllow  = regexp.MustCompile(`127\.0\.0\.1|0\.0\.0\.0|255\.255\.255\.255|192\.0\.2\.|198\.51\.100\.|203\.0\.113\.`)
	lintPathRe   = regexp.MustCompile(`/root/|/home/[a-zA-Z0-9._-]+/`)
	lintIdentRe  = regexp.MustCompile(`code27\.cn|bigmodel|GLM_CODING|DEEPSEEK_API_KEY|DASHSCOPE_API_KEY|MOONSHOT_API_KEY|glm-[0-9]`)
	lintTruncate = 80
)

// Lint 设计文档正文 R-PUB-1 检查（IP/部署绝对路径/个人标识——lint/personal-info.sh 同清单）。
// 纯函数零副作用；命中=违规清单（行号 1 起），全绿=nil。
func Lint(content []byte) []Violation {
	var out []Violation
	for i, line := range strings.Split(string(content), "\n") {
		hit := func(re *regexp.Regexp, allow *regexp.Regexp, rule string) {
			loc := re.FindStringIndex(line)
			if loc == nil {
				return
			}
			if allow != nil && allow.MatchString(line) {
				return // 整行含白名单词形=合规示例（personal-info.sh 行级过滤同口径）
			}
			text := line[loc[0]:loc[1]]
			if len(text) > lintTruncate {
				text = text[:lintTruncate]
			}
			out = append(out, Violation{Line: i + 1, Rule: rule, Text: text})
		}
		hit(lintIPRe, lintIPAllow, "ip")
		hit(lintPathRe, nil, "path")
		hit(lintIdentRe, nil, "ident")
	}
	return out
}

// ─────────────────────────────────────────────────────────────────────────────
// 仓集合（白名单）
// ─────────────────────────────────────────────────────────────────────────────

// Repo 白名单内一个设计文档仓（name=投影名，root=本地工作树根）。
type Repo struct {
	Name string
	Root string
}

// File 清单一行（瘦身：path+HEAD blob sha+字节数；无正文）。
type File struct {
	Repo  string `json:"repo"`
	Path  string `json:"path"`
	Sha   string `json:"sha"`
	Bytes int    `json:"bytes"`
}

// Commit 最近一次提交投影（任务书读面判据「打开显示正文与最近提交」）。
type Commit struct {
	Sha     string    `json:"sha"`
	Subject string    `json:"subject"`
	At      time.Time `json:"at"`
}

// Doc 单文档读面（content=HEAD 正文——权威唯一，工作树未提交改动不进读面）。
type Doc struct {
	Repo       string  `json:"repo"`
	Path       string  `json:"path"`
	Sha        string  `json:"sha"`
	Bytes      int     `json:"bytes"`
	Content    []byte  `json:"-"`
	ContentStr string  `json:"content"`
	LastCommit *Commit `json:"lastCommit"`
}

// Set 设计文档仓集合（白名单权威：仅本集合注册仓的 .md 可读可写）。
// Apply 走单互斥串行（git 工作树/索引操作非并发安全——写路径串行化与 repository 串行队列同口径）。
type Set struct {
	mu     sync.Mutex
	repos  []Repo
	deps   *Engine
	gitLog bool // 最近提交投影是否可查（缺省开；测试可关）
}

// Engine 写面依赖（--allow-apply 装配注入；nil=写面未开启 405）。
// 依赖方向：docrepo 不 import 存储/权限实现——经窄口面注入（flowbind/watchsrc 同款范式）。
type Engine struct {
	Bus  EventSink // DocApplied 发布（nil=不发布不报错——可丢语义 event.md §1）
	Ops  AuditSink // apply_log 审计行（nil=不落账不推荐——ops-log.md）
	Perm PermFace  // 权限判定（permission.md 机读——写设计文档=高危）
	// Domain 写权限域（permission.md 动作域词汇：core/* 资源域，scheme.yaml Role spec 声明随行）。
	Domain string
}

// EventSink 事件发布窄口面（eventbus.Bus 满足）。
type EventSink interface {
	Publish(e *eventbus.Event) error
}

// AuditSink 审计落账窄口面（opslog.SQLiteStore 满足）。
type AuditSink interface {
	AppendApply(l opslog.ApplyLog) error
}

// PermFace 权限判定窄口面（permission.System 满足）。
type PermFace interface {
	Check(sub permission.Subject, scope, action string, by permission.JudgedBy) error
}

// NewSet 解析 --design-repo 规格清单（形 "name=path"，可重复）。
// 仓根必须存在且是 git 仓（rev-parse 通过）——否则装配失败拒绝启动（写面依附真实仓）。
func NewSet(specs ...string) (*Set, error) {
	s := &Set{gitLog: true}
	seen := map[string]bool{}
	for _, spec := range specs {
		spec = strings.TrimSpace(spec)
		if spec == "" {
			continue
		}
		name, root, ok := strings.Cut(spec, "=")
		name, root = strings.TrimSpace(name), strings.TrimSpace(root)
		if !ok || name == "" || root == "" {
			return nil, fmt.Errorf("docrepo: --design-repo 规格不合法（需 name=path）: %q", spec)
		}
		if name == "draft" || name == "history" || name == "apply" {
			return nil, fmt.Errorf("docrepo: 仓名 %q 为保留路由段", name)
		}
		if seen[name] {
			return nil, fmt.Errorf("docrepo: 仓名重复 %q", name)
		}
		seen[name] = true
		abs, err := filepath.Abs(root)
		if err != nil {
			return nil, fmt.Errorf("docrepo: 仓根 %q: %w", root, err)
		}
		if st, err := os.Stat(abs); err != nil || !st.IsDir() {
			return nil, fmt.Errorf("docrepo: 仓根不存在或非目录: %s", abs)
		}
		if _, err := r0git(abs, "rev-parse", "--git-dir"); err != nil {
			return nil, fmt.Errorf("docrepo: 仓根不是 git 仓: %s: %w", abs, err)
		}
		s.repos = append(s.repos, Repo{Name: name, Root: abs})
	}
	sort.Slice(s.repos, func(i, j int) bool { return s.repos[i].Name < s.repos[j].Name })
	return s, nil
}

// BindEngine 装配层注入写面引擎（--allow-apply 装配后调用；nil=写面未开启）。
func (s *Set) BindEngine(e *Engine) { s.mu.Lock(); s.deps = e; s.mu.Unlock() }

// WriteEnabled 写面就绪态（引擎已装配=--design-repo ∧ --allow-apply；读面无关本态）。
func (s *Set) WriteEnabled() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.deps != nil
}

// Names 注册仓名清单（装配锚投影）。
func (s *Set) Names() []string {
	out := make([]string, 0, len(s.repos))
	for _, r := range s.repos {
		out = append(out, r.Name)
	}
	return out
}

func (s *Set) repo(name string) (*Repo, bool) {
	for i := range s.repos {
		if s.repos[i].Name == name {
			return &s.repos[i], true
		}
	}
	return nil, false
}

// ObjectKey 对象键投影（事件 involved / 审计 obj 三段同构惯例：DesignDoc/<repo>/<path>）。
func ObjectKey(repo, path string) string { return "DesignDoc/" + repo + "/" + path }

// ─────────────────────────────────────────────────────────────────────────────
// 白名单校验（读/写共用同一闸）
// ─────────────────────────────────────────────────────────────────────────────

// validatePath 白名单闸：注册仓 ∧ 相对干净路径 ∧ .md 设计文档 ∧ 不落 .git/。
// 越权路径=ErrForbidden（任务书判据「非白名单路径 → 403」）。
func validatePath(repo, path string) error {
	if repo == "" || path == "" {
		return ErrInvalid
	}
	if strings.ContainsAny(path, "\\\x00") || strings.Contains(path, "\n") {
		return fmt.Errorf("%w: 路径含非法字符 %q", ErrForbidden, path)
	}
	if strings.HasPrefix(path, "/") || filepath.IsAbs(path) {
		return fmt.Errorf("%w: 绝对路径 %q（只收仓内相对路径）", ErrForbidden, path)
	}
	clean := filepath.Clean(path)
	if clean == "." || clean == ".." || strings.HasPrefix(clean, "../") || strings.Contains(clean, "/../") {
		return fmt.Errorf("%w: 路径穿越 %q", ErrForbidden, path)
	}
	if filepath.Clean(path) != path {
		return fmt.Errorf("%w: 路径未规范化 %q（用 %q）", ErrForbidden, path, clean)
	}
	for _, seg := range strings.Split(clean, "/") {
		if seg == ".git" {
			return fmt.Errorf("%w: .git 目录不可经本面读写", ErrForbidden)
		}
	}
	if !strings.HasSuffix(clean, ".md") {
		return fmt.Errorf("%w: 只收 .md 设计文档（实得 %q）", ErrForbidden, path)
	}
	return nil
}

// ─────────────────────────────────────────────────────────────────────────────
// 读面（权威=git HEAD）
// ─────────────────────────────────────────────────────────────────────────────

// List 全部注册仓的 .md 文件清单（git ls-tree -r HEAD——权威=HEAD；瘦身心智：无正文）。
func (s *Set) List() ([]File, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []File
	for _, r := range s.repos {
		tree, err := r0git(r.Root, "ls-tree", "-r", "-z", "--full-tree", "HEAD")
		if err != nil {
			continue // 空仓（无 HEAD）=零文件合法
		}
		for _, rec := range strings.Split(tree, "\x00") {
			if rec == "" {
				continue
			}
			meta, p, ok := strings.Cut(rec, "\t")
			if !ok || !strings.HasSuffix(p, ".md") {
				continue
			}
			fields := strings.Fields(meta) // mode type sha
			if len(fields) < 3 {
				continue
			}
			out = append(out, File{Repo: r.Name, Path: p, Sha: fields[2]})
		}
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].Repo != out[j].Repo {
			return out[i].Repo < out[j].Repo
		}
		return out[i].Path < out[j].Path
	})
	for i := range out {
		out[i].Bytes = -1 // 清单不读内容（防拖库同口径——字节数按需在 Read 补）
	}
	return out, nil
}

// Read 单文档：HEAD 正文 + blob sha + 最近提交（任务书读面判据 1）。
func (s *Set) Read(repoName, path string) (*Doc, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.repo(repoName)
	if !ok {
		return nil, ErrNoRepo
	}
	if err := validatePath(repoName, path); err != nil {
		return nil, err
	}
	return s.read(r, path)
}

// ReadSidecar 侧车只读（R-DOC-6 `<md路径>.metadatainfo`——2026-11-28 用户裁决 metadatainfo
// 上屏：console 元数据 header 组件数据源）。白名单面只收 .md，侧车走本口面：
// 仅 `*.md.metadatainfo` 单一后缀放行（仓注册+路径穿越+.git 拒绝同 validatePath 口径），
// 零写路径（read 只读 HEAD）。
func (s *Set) ReadSidecar(repoName, path string) (*Doc, error) {
	if !strings.HasSuffix(path, ".md.metadatainfo") {
		return nil, fmt.Errorf("%w: 侧车口面只收 <md>.metadatainfo（实得 %q）", ErrForbidden, path)
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.repo(repoName)
	if !ok {
		return nil, ErrNoRepo
	}
	if err := validateSidecarPath(repoName, path); err != nil {
		return nil, err
	}
	return s.read(r, path)
}

// validateSidecarPath 侧车路径校验（<md>.metadatainfo：剥离后缀后必须过 .md 白名单校验）。
func validateSidecarPath(repoName, path string) error {
	base := strings.TrimSuffix(path, ".metadatainfo")
	if base == path {
		return fmt.Errorf("%w: 侧车路径缺 .metadatainfo 后缀 %q", ErrForbidden, path)
	}
	return validatePath(repoName, base+".md")
}

func (s *Set) read(r *Repo, path string) (*Doc, error) {
	sha, err := r0git(r.Root, "rev-parse", "HEAD:"+path)
	if err != nil {
		return nil, fmt.Errorf("%w: %s/%s（%v）", ErrNotFound, r.Name, path, errTail(err))
	}
	out, err := r0gitBytes(r.Root, "cat-file", "blob", sha)
	if err != nil {
		return nil, fmt.Errorf("%w: 读 HEAD 内容失败 %s/%s: %v", ErrGit, r.Name, path, errTail(err))
	}
	d := &Doc{Repo: r.Name, Path: path, Sha: sha, Bytes: len(out), Content: out, ContentStr: string(out)}
	if s.gitLog {
		log, err := r0git(r.Root, "log", "-1", "--format=%H%x1f%s%x1f%aI", "--", path)
		if err == nil && log != "" {
			parts := strings.SplitN(log, "\x1f", 3)
			if len(parts) == 3 {
				at, _ := time.Parse(time.RFC3339, strings.TrimSpace(parts[2]))
				d.LastCommit = &Commit{Sha: parts[0], Subject: parts[1], At: at}
			}
		}
	}
	return d, nil
}

// headFile HEAD 侧文件状态（sha+内容；不存在=空 sha+nil 内容）。
func (s *Set) headFile(r *Repo, path string) (string, []byte, error) {
	sha, err := r0git(r.Root, "rev-parse", "HEAD:"+path)
	if err != nil {
		return "", nil, nil // HEAD 无此文件（将创建）
	}
	content, err := r0gitBytes(r.Root, "cat-file", "blob", sha)
	if err != nil {
		return "", nil, err
	}
	return sha, content, nil
}

// dirty 工作树该文件是否有未提交改动（ porcelain 对比 HEAD——不吞他人未提交编辑）。
func (s *Set) dirty(r *Repo, path string) (bool, error) {
	status, err := r0git(r.Root, "status", "--porcelain", "-z", "--", path)
	if err != nil {
		return false, err
	}
	return strings.TrimSpace(status) != "", nil
}

// ─────────────────────────────────────────────────────────────────────────────
// Apply（dry-run 三检查 → git commit）
// ─────────────────────────────────────────────────────────────────────────────

// ApplyRequest 一次设计文档 Apply 的输入（apiserver 投影；fromDraft 的取稿/清稿在装配层）。
type ApplyRequest struct {
	Repo         string
	Path         string
	BaseSha      string              // 打开时读到的 HEAD blob sha（""=预期创建）
	Content      []byte              // 新正文（fromDraft 时由装配层从草稿取）
	FieldManager string              // 必填（apply.md rule 2 同语义）
	ReasonRef    string              // 审计回链（console:apply 等）
	DryRun       bool                // 零副作用预检（与真执行同闸门）
	Subject      *permission.Subject // 请求级主体（nil=ErrNoEngine 侧兜底拒绝——HTTP 匿名不享受本地信任）
}

// Result 一次 Apply 的结果（commit 仅真实落库时在册；noop/dry-run=nil）。
type Result struct {
	Action string  // create | update | noop | dry-run
	Sha    string  // 落库后 HEAD blob sha（noop/dry-run=现 HEAD sha；将创建的 dry-run=""）
	Commit *Commit // 新提交（message 自动生成）
	Event  *eventbus.Event
}

// Apply 设计文档写路径三检查（白名单→lint→乐观锁）→ git commit。幂等：同内容重入=noop
// （零提交零事件零审计行之外不落 ok 行——落 noop 行照旧，审计流水完整性优先）。
// 红线：不 push（commit 到本地仓即止）；不 enable workflow（事件主干零旁路、下游过滤面各管各的）。
func (s *Set) Apply(ctx context.Context, req ApplyRequest) (*Result, error) {
	if req.FieldManager == "" {
		return nil, ErrNoFieldManager
	}
	if req.Repo == "" || req.Path == "" {
		return nil, ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	r, ok := s.repo(req.Repo)
	if !ok {
		return nil, ErrNoRepo
	}
	if err := validatePath(req.Repo, req.Path); err != nil {
		return nil, err
	}
	dep := s.deps
	if dep == nil {
		return nil, ErrNoEngine
	}
	if req.Subject == nil {
		req.Subject = &permission.Subject{} // 匿名主体照判（不享受本地信任——cored-bridge §2.1 同口径）
	}
	// 权限准入（permission.md 机读判定——写设计文档=高危，admin 直通=DefaultAdmin 字面量既有语义）。
	domain := dep.Domain
	if domain == "" {
		domain = "core/design-docs"
	}
	if err := dep.Perm.Check(*req.Subject, domain, "write", permission.ByRole); err != nil {
		s.audit(dep, req, 0, "denied", err.Error(), req.Subject)
		return nil, fmt.Errorf("%w: %v", ErrDenied, err)
	}
	// 乐观锁基准：现 HEAD blob sha（""=HEAD 无此文件）。
	curSha, curContent, err := s.headFile(r, req.Path)
	if err != nil {
		return nil, fmt.Errorf("%w: 读 HEAD 基线失败: %v", ErrGit, errTail(err))
	}
	// 工作树未提交改动=拒绝（不吞他人未提交编辑——ErrDirty）。
	if d, err := s.dirty(r, req.Path); err == nil && d {
		return nil, ErrDirty
	}
	// R-PUB-1 lint（白名单后、落库前——坏内容不落公开仓）。
	if vs := Lint(req.Content); len(vs) > 0 {
		return nil, &LintError{Violations: vs}
	}
	if req.BaseSha != curSha {
		return nil, &ConflictError{Repo: req.Repo, Path: req.Path, Expected: req.BaseSha, Current: curSha}
	}
	// dry-run：零副作用（不提交不落 ok 行不发事件），返回与真执行一致的准入结论+现版本。
	if req.DryRun {
		return &Result{Action: "dry-run", Sha: curSha}, nil
	}
	// 幂等：同内容重入=noop（零提交零事件；noop 审计行照落——判定发生了）。
	if curContent != nil && bytes.Equal(curContent, req.Content) {
		s.audit(dep, req, 0, "noop", "", req.Subject)
		return &Result{Action: "noop", Sha: curSha}, nil
	}
	action := "update"
	if curContent == nil {
		action = "create"
	}
	verb := "更新"
	if action == "create" {
		verb = "新增"
	}
	msg := fmt.Sprintf("docs(%s): %s设计文档 %s（console 屏幕 6 在线编辑——能力 A；fieldManager=%s）",
		req.Repo, verb, req.Path, req.FieldManager)
	if err := os.WriteFile(filepath.Join(r.Root, filepath.FromSlash(req.Path)), req.Content, 0o644); err != nil {
		return nil, fmt.Errorf("%w: 写工作树文件失败: %v", ErrGit, errTail(err))
	}
	if _, err := r0gitCtx(ctx, r.Root, "add", "--", req.Path); err != nil {
		return nil, fmt.Errorf("%w: git add 失败: %v", ErrGit, errTail(err))
	}
	if _, err := r0gitCtx(ctx, r.Root, "commit", "-m", msg, "--", req.Path); err != nil {
		return nil, fmt.Errorf("%w: git commit 失败（工作树文件已更新）: %v", ErrGit, errTail(err))
	}
	newSha, err := r0gitCtx(ctx, r.Root, "rev-parse", "HEAD:"+req.Path)
	if err != nil {
		return nil, fmt.Errorf("%w: 落库后读新 blob sha 失败: %v", ErrGit, errTail(err))
	}
	commit := &Commit{Sha: shortSha(newSha)}
	if log, err := r0gitCtx(ctx, r.Root, "log", "-1", "--format=%H%x1f%s%x1f%aI", "--", req.Path); err == nil {
		if parts := strings.SplitN(log, "\x1f", 3); len(parts) == 3 {
			at, _ := time.Parse(time.RFC3339, strings.TrimSpace(parts[2]))
			commit = &Commit{Sha: parts[0], Subject: parts[1], At: at}
		}
	}
	s.audit(dep, req, 0, "ok", "", req.Subject)
	res := &Result{Action: action, Sha: newSha, Commit: commit}
	// DocApplied 必发（apply.md rule 1 同语义；事件主干零旁路——三段 involved=DesignDoc/<repo>/<path>）。
	if dep.Bus != nil {
		evt := eventbus.Event{
			Type: "DocApplied", Source: "apply", Involved: ObjectKey(req.Repo, req.Path),
			DedupKey: fmt.Sprintf("apply:%s:%s", ObjectKey(req.Repo, req.Path), shortSha(newSha)),
			Level:    "info",
		}
		if err := dep.Bus.Publish(&evt); err != nil {
			return nil, fmt.Errorf("docrepo: DocApplied 发布失败（提交已落 %s）: %w", commit.Sha, err)
		}
		res.Event = &evt
	}
	return res, nil
}

// audit 审计行（ok/noop/conflict/denied——denied 行携主体标识，permission.md 规则 3 判定留痕）。
// 审计失败不改变写路径结果（流水非权威）。
func (s *Set) audit(dep *Engine, req ApplyRequest, rv int64, result, decision string, sub *permission.Subject) {
	if dep == nil || dep.Ops == nil {
		return
	}
	id := "anonymous"
	if sub != nil {
		id = sub.ID()
	}
	_ = dep.Ops.AppendApply(opslog.ApplyLog{
		Ts: time.Now(), FieldManager: req.FieldManager, Subject: id,
		Obj: ObjectKey(req.Repo, req.Path), ResourceVersion: rv,
		Result: result, PolicyDecisions: decision, ReasonRef: req.ReasonRef,
	})
}

// ─────────────────────────────────────────────────────────────────────────────
// 草稿（kind=DesignDoc 与对象草稿同一缓存同一持久化面——apiserver 装配层桥接）
// ─────────────────────────────────────────────────────────────────────────────

// DraftKind 设计文档草稿的 kind 段（draft.Store 键=(kind,scope,name)——scope=仓、name=路径）。
const DraftKind = "DesignDoc"

// ─────────────────────────────────────────────────────────────────────────────
// git 执行（各仓自带提交身份=仓/全局 git config——cored 不注入作者，提交身份随仓）
// ─────────────────────────────────────────────────────────────────────────────

func r0git(root string, args ...string) (string, error) {
	return r0gitCtx(context.Background(), root, args...)
}

// r0gitBytes 内容字节读取（cat-file blob 用——不做 TrimSpace，正文逐字节保真）。
func r0gitBytes(root string, args ...string) ([]byte, error) {
	cmd := exec.Command("git", args...)
	cmd.Dir = root
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("git %s: %w（%s）", strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}
	return out.Bytes(), nil
}

func r0gitCtx(ctx context.Context, root string, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = root
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		return out.String(), fmt.Errorf("git %s: %w（%s）", strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}
	return strings.TrimSpace(out.String()), nil
}

func errTail(err error) string {
	s := err.Error()
	if len(s) > 160 {
		return s[:160]
	}
	return s
}

func shortSha(sha string) string {
	if len(sha) > 12 {
		return sha[:12]
	}
	return sha
}
