// export.go —— sqlite→git 定时导出器（R-DOC-7 规则 2：git=导出/公开 output/public——
// 周期快照，不是逐次写账本；逐次正文账=doc_history，见 sqlite.go）。设计依据：
// instance 仓 runtime/docs-export.md（导出=读侧派生任务：水位无变化跳过=幂等零提交；
// 失败只告警不回滚 DB 不崩 cored；no force-push；布局=既有 git 导入导出格式不发明第二套）。
//
// 单向 output 红线：导出器只读 docs DB（Repository 读半边）→ 按对象规范路径写目标工作树 →
// 有变化才 git add -A + commit → 显式配置了 remote 才 push；**绝不从工作树回导 DB**
// （回导=SeedFromLayout，仅部署侧 --docs 显式一次性迁移通道）。sidecar（R-DOC-6
// `<文件>.metadatainfo`）随行导出：生成逻辑在 DocIndex/打标面（不在本仓），导出侧只
// 保留既有侧车不发明不篡改，搬文档=搬侧车、删对象=删侧车（同目录同提交）。
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
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"gopkg.in/yaml.v3"
)

// ExportConfig 文档导出独立配置项（存储在 sqlite docs DB 的 docs_export_config 表，
// 初值从 scheme.yaml docTypes[].exportDefault / category=runtime 派生，绝不在代码中硬编码某类型）。
type ExportConfig struct {
	IncludeRuntimeDocs bool     `json:"includeRuntimeDocs"`
	ExcludeDocTypes    []string `json:"excludeDocTypes"`
	ExcludeKinds       []string `json:"excludeKinds"`
	RuntimeDocTypes    []string `json:"runtimeDocTypes"`
	UpdatedAt          string   `json:"updatedAt,omitempty"`
}

// ExportConfigStore 文档导出配置持久化接口（SQLiteRepo 实现；独立配置项读写）。
type ExportConfigStore interface {
	GetExportConfig(ctx context.Context) (ExportConfig, error)
	SetExportConfig(ctx context.Context, cfg ExportConfig) (ExportConfig, error)
}

// DefaultExportConfig 从 scheme 注册表派生默认导出配置（仅读取 scheme.docTypes 声明的
// category=runtime / exportDefault=false，零代码硬编码类型名）。
func DefaultExportConfig(sch *scheme.Scheme) ExportConfig {
	var runtimeTypes []string
	if sch != nil {
		for _, dt := range sch.Spec.DocTypes {
			if dt.IsRuntimeDoc() {
				runtimeTypes = append(runtimeTypes, dt.Name)
			}
		}
	}
	if runtimeTypes == nil {
		runtimeTypes = []string{}
	}
	excl := make([]string, len(runtimeTypes))
	copy(excl, runtimeTypes)
	return ExportConfig{
		IncludeRuntimeDocs: false,
		ExcludeDocTypes:    excl,
		ExcludeKinds:       []string{},
		RuntimeDocTypes:    runtimeTypes,
	}
}

// Exporter sqlite→git 导出器。单实例对应一个目标工作树；cored 装配层经 Run 挂定时拍
// （CORE_SYSTEM_DOCS_EXPORT_INTERVAL 开关/间隔 × CORE_SYSTEM_DOCS_EXPORT_DIR 必填 ×
// CORE_SYSTEM_DOCS_EXPORT_REMOTE 可选 push）。Export 可独立调用（手动触发面）。
type Exporter struct {
	repo   Repository // 只读消费（List 半边；写路径零触碰——导出不参与 Apply/权限/事件链）
	sch    *scheme.Scheme
	dir    string // 目标工作树（output/public；导出根可被 SeedFromLayout 回放）
	remote string // 可选 push 远程（空=本地 commit 止步；no force-push）

	mu sync.Mutex // Export 串行化（Run 定时拍与手动触发并存时防交叠写树）
}

// NewExporter 构建导出器（dir 必填；remote 可选）。不在此建目录/初始化 git（首次 Export 拍
// 惰性就绪——目标树缺失属导出面可自愈态，非装配错误）。
func NewExporter(repo Repository, sch *scheme.Scheme, dir, remote string) (*Exporter, error) {
	if repo == nil {
		return nil, errors.New("repository: 导出器需文档库（sqlite 权威——导出是读侧派生任务）")
	}
	if sch == nil {
		return nil, errors.New("repository: 导出器需 scheme 注册表（对象规范路径解析面）")
	}
	if strings.TrimSpace(dir) == "" {
		return nil, errors.New("repository: 导出器需目标工作树（CORE_SYSTEM_DOCS_EXPORT_DIR 必填才启动）")
	}
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, fmt.Errorf("repository: 导出目录 %q: %w", dir, err)
	}
	return &Exporter{repo: repo, sch: sch, dir: abs, remote: strings.TrimSpace(remote)}, nil
}

// ExportResult 单拍导出投影（stderr 装配锚与测试判据）。
type ExportResult struct {
	Objects   int    // 本次快照对象数
	HighWater int64  // 对象 rv 高水位（提交消息 rv=<high-water> 同源）
	Committed bool   // 本拍是否产生 commit（水位相同=幂等零提交）
	Pushed    bool   // 是否已 push（remote 未配置=恒 false）
	Commit    string // commit 短 sha（未提交=空）
}

// Run 定时导出拍（cored goroutine 装配）：每拍 Export；失败只记 stderr（导出失败不回滚
// sqlite 写、不崩 cored——R-DOC-7 规则 5，下拍重试即自愈）；ctx 取消即退场。
func (e *Exporter) Run(ctx context.Context, interval time.Duration) {
	tk := time.NewTicker(interval)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			res, err := e.Export(ctx)
			if err != nil {
				fmt.Fprintf(os.Stderr, "repository: docs 导出失败（只告警不回滚不崩——R-DOC-7 规则 5，下拍重试）: %v\n", err)
				continue
			}
			if res.Committed {
				fmt.Fprintf(os.Stderr, "repository: docs 导出提交 %s（objects=%d rv=%d pushed=%v；git 历史=导出快照非逐次账）\n",
					res.Commit, res.Objects, res.HighWater, res.Pushed)
			}
		}
	}
}

// exportItem 单对象导出条目（布局相对路径 ↔ 对象一一对应——storage 模板唯一具体分支）。
type exportItem struct {
	rel string
	obj *Object
}

// Export 执行一拍导出：全量对象按 layout 目录即对象规范写目标工作树 → 有变化才
// git add -A + commit（消息 export: <RFC3339> rv=<高水位> objects=<n>）→ remote 已配置才 push。
// 返回 Committed=false + nil err=工作树与快照一致（幂等零提交零推送；变更判定=git status
// 空判——同水位重复导出零提交，同时侧车/外部落件随下拍收编）。
func (e *Exporter) Export(ctx context.Context) (ExportResult, error) {
	e.mu.Lock()
	defer e.mu.Unlock()
	items, err := e.enumerate(ctx)
	if err != nil {
		return ExportResult{}, err
	}
	high := int64(0)
	for _, it := range items {
		if it.obj.ResourceVersion > high {
			high = it.obj.ResourceVersion
		}
	}
	if err := os.MkdirAll(e.dir, 0o755); err != nil {
		return ExportResult{}, fmt.Errorf("repository: 导出目录 %s: %w", e.dir, err)
	}
	if err := e.ensureGit(ctx); err != nil {
		return ExportResult{}, err
	}
	// 写对象文件：文件内容=doc.content 原文逐字节（不重排/不格式化——docs-export.md §3
	// 导出布局=导入导出格式，与 SeedFromLayout 同一解析口径的写半边，可回放）。
	for _, it := range items {
		p := filepath.Join(e.dir, filepath.FromSlash(it.rel))
		if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
			return ExportResult{}, fmt.Errorf("repository: 导出建目录 %s: %w", filepath.Dir(p), err)
		}
		if err := os.WriteFile(p, it.obj.Content, 0o644); err != nil {
			return ExportResult{}, fmt.Errorf("repository: 导出写文件 %s: %w", p, err)
		}
	}
	// 陈旧文件删除 + sidecar 随行（docs-export.md §3：DB 中不存在的对象在导出提交中删除）。
	if err := e.pruneStale(items); err != nil {
		return ExportResult{}, err
	}
	if _, err := e.git(ctx, "add", "-A"); err != nil {
		return ExportResult{}, err
	}
	status, err := e.git(ctx, "status", "--porcelain")
	if err != nil {
		return ExportResult{}, err
	}
	res := ExportResult{Objects: len(items), HighWater: high}
	if strings.TrimSpace(status) == "" {
		// 零变更（写盘内容与索引一致）：零提交零推送（幂等判据——同一水位重复导出=零提交）。
		return res, nil
	}
	msg := fmt.Sprintf("export: %s rv=%d objects=%d", time.Now().UTC().Format(time.RFC3339), high, len(items))
	if _, err := e.git(ctx, "-c", "commit.gpgsign=false", "commit", "-m", msg); err != nil {
		return ExportResult{}, err
	}
	sha, err := e.git(ctx, "rev-parse", "--short", "HEAD")
	if err != nil {
		return ExportResult{}, err
	}
	res.Committed, res.Commit = true, sha
	if e.remote != "" {
		if _, err := e.git(ctx, "push", e.remote, "HEAD"); err != nil {
			// commit 已落、push 未遂：报错回装配层（只告警不崩——下拍重试 push 或人工介入；
			// no force-push——冲突=人工介入不自动覆盖，docs-export.md §4）。
			return res, fmt.Errorf("repository: 导出 push %s: %w", e.remote, err)
		}
		res.Pushed = true
	}
	return res, nil
}

// enumerate 全量对象快照：按 scheme 注册表逐 kind 跨 scope 枚举（Repository 读半边）+ storage
// 模板解析为唯一具体布局路径（与写路径 concretePath 同一闸门：kind 注册/scope 四层/命名规则/
// 模板相容）。该 kind 无可寻址分支=跳过（同 SeedFromLayout 读侧宽松口径）；键不可寻址的
// 对象（写路径闸门后的理论不可达态）跳过不致命——导出为 best-effort 读侧派生任务。
func (e *Exporter) enumerate(ctx context.Context) ([]exportItem, error) {
	cfg := DefaultExportConfig(e.sch)
	if store, ok := e.repo.(ExportConfigStore); ok {
		if dbCfg, err := store.GetExportConfig(ctx); err == nil {
			cfg = dbCfg
		}
	}
	exclKinds := make(map[string]bool, len(cfg.ExcludeKinds))
	for _, k := range cfg.ExcludeKinds {
		if s := strings.TrimSpace(k); s != "" {
			exclKinds[s] = true
		}
	}
	exclDocTypes := make(map[string]bool, len(cfg.ExcludeDocTypes))
	if !cfg.IncludeRuntimeDocs {
		for _, dt := range cfg.ExcludeDocTypes {
			if s := strings.TrimSpace(dt); s != "" {
				exclDocTypes[s] = true
			}
		}
	}
	var out []exportItem
	seen := map[string]bool{}
	for _, k := range e.sch.ListKinds() {
		if exclKinds[k.Name] || exclDocTypes[k.Name] {
			continue
		}
		objs, err := e.repo.List(ctx, k.Name, "")
		if err != nil {
			continue // 模板不可展开（无可寻址分支）——跳过不致命
		}
		for _, o := range objs {
			if len(exclDocTypes) > 0 {
				if dt := extractObjectDocType(o.Content); dt != "" && exclDocTypes[dt] {
					continue // 数据库导出配置排除的运行时文档类型（如 BugDoc 等）——跳过不导出
				}
			}
			pats, err := storagePatterns(k, o.Scope, o.Name, true)
			if err != nil || len(pats) != 1 || seen[pats[0]] {
				continue // 键不可寻址/路径冲突（首者胜，确定性）——跳过
			}
			seen[pats[0]] = true
			out = append(out, exportItem{rel: pats[0], obj: o})
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].rel < out[j].rel })
	return out, nil
}

// extractObjectDocType 从对象正文（YAML 或 Markdown frontmatter）提取 docType 声明。
func extractObjectDocType(raw []byte) string {
	body := raw
	trimmed := bytes.TrimSpace(raw)
	if bytes.HasPrefix(trimmed, []byte("---")) {
		rest := trimmed[3:]
		if idx := bytes.Index(rest, []byte("\n---")); idx >= 0 {
			body = rest[:idx]
		}
	}
	var doc struct {
		DocType  string `yaml:"docType"`
		Metadata struct {
			DocType string `yaml:"docType"`
		} `yaml:"metadata"`
		Spec struct {
			DocType string `yaml:"docType"`
		} `yaml:"spec"`
	}
	if err := yaml.Unmarshal(body, &doc); err != nil {
		return ""
	}
	if doc.DocType != "" {
		return doc.DocType
	}
	if doc.Metadata.DocType != "" {
		return doc.Metadata.DocType
	}
	return doc.Spec.DocType
}

// ensureExportConfigTable 确保 sqlite docs_export_config 表存在并按 scheme 种子化默认行。
func (r *SQLiteRepo) ensureExportConfigTable(ctx context.Context) error {
	const ddl = `CREATE TABLE IF NOT EXISTS docs_export_config (
		id                   INTEGER PRIMARY KEY CHECK (id = 1),
		include_runtime_docs INTEGER NOT NULL DEFAULT 0,
		exclude_doc_types    TEXT NOT NULL DEFAULT '',
		exclude_kinds        TEXT NOT NULL DEFAULT '',
		updated_at           TEXT NOT NULL
	);`
	if _, err := r.db.ExecContext(ctx, ddl); err != nil {
		return err
	}
	def := DefaultExportConfig(r.sch)
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO docs_export_config (id, include_runtime_docs, exclude_doc_types, exclude_kinds, updated_at) VALUES (1, 0, ?, '', ?)`,
		strings.Join(def.ExcludeDocTypes, ","), now)
	return err
}

// GetExportConfig 读取数据库中的文档导出独立配置。
func (r *SQLiteRepo) GetExportConfig(ctx context.Context) (ExportConfig, error) {
	if err := r.ensureExportConfigTable(ctx); err != nil {
		return DefaultExportConfig(r.sch), err
	}
	var inc int
	var exclDT, exclK, updatedAt string
	err := r.db.QueryRowContext(ctx,
		`SELECT include_runtime_docs, exclude_doc_types, exclude_kinds, updated_at FROM docs_export_config WHERE id = 1`).
		Scan(&inc, &exclDT, &exclK, &updatedAt)
	if err != nil {
		return DefaultExportConfig(r.sch), err
	}
	def := DefaultExportConfig(r.sch)
	return ExportConfig{
		IncludeRuntimeDocs: inc != 0,
		ExcludeDocTypes:    splitCSV(exclDT),
		ExcludeKinds:       splitCSV(exclK),
		RuntimeDocTypes:    def.RuntimeDocTypes,
		UpdatedAt:          updatedAt,
	}, nil
}

// SetExportConfig 更新数据库中的文档导出独立配置。
func (r *SQLiteRepo) SetExportConfig(ctx context.Context, cfg ExportConfig) (ExportConfig, error) {
	if err := r.ensureExportConfigTable(ctx); err != nil {
		return ExportConfig{}, err
	}
	inc := 0
	if cfg.IncludeRuntimeDocs {
		inc = 1
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx,
		`UPDATE docs_export_config SET include_runtime_docs = ?, exclude_doc_types = ?, exclude_kinds = ?, updated_at = ? WHERE id = 1`,
		inc, strings.Join(cfg.ExcludeDocTypes, ","), strings.Join(cfg.ExcludeKinds, ","), now)
	if err != nil {
		return ExportConfig{}, err
	}
	return r.GetExportConfig(ctx)
}

func splitCSV(s string) []string {
	if strings.TrimSpace(s) == "" {
		return []string{}
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

// pruneStale 删除工作树中 DB 快照之外的文件（删除只在导出提交中体现）；`<对象路径>.metadatainfo`
// 侧车随对象路径保留（R-DOC-6/7：同目录同提交，搬文档=搬侧车；生成逻辑在 DocIndex/打标面，
// 导出侧只随行不发明不篡改）；对象已不在册 → 其侧车随行删除（删对象=删侧车，R-DOC-10 同口径）。
func (e *Exporter) pruneStale(items []exportItem) error {
	keep := make(map[string]bool, len(items))
	for _, it := range items {
		keep[it.rel] = true
	}
	return filepath.WalkDir(e.dir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			if d.Name() == ".git" {
				return fs.SkipDir
			}
			return nil
		}
		rel, rerr := filepath.Rel(e.dir, path)
		if rerr != nil {
			return nil // 相对化失败不删（宁留勿误删——导出面不做破坏性猜测）
		}
		rel = filepath.ToSlash(rel)
		if keep[rel] {
			return nil
		}
		if strings.HasSuffix(rel, ".metadatainfo") && keep[strings.TrimSuffix(rel, ".metadatainfo")] {
			return nil // 在册对象的 sidecar 随行保留
		}
		return os.Remove(path)
	})
}

// ensureGit 目标工作树 git 就绪（无 .git 则 init——导出根=output/public 快照仓）。
func (e *Exporter) ensureGit(ctx context.Context) error {
	if _, err := e.git(ctx, "rev-parse", "--git-dir"); err == nil {
		return nil
	}
	if _, err := e.git(ctx, "init", "-q"); err != nil {
		return fmt.Errorf("repository: 导出 git init %s: %w", e.dir, err)
	}
	return nil
}

// git 在导出工作树执行 git 子命令（统一注入提交身份，不依赖宿主全局配置——与 GitRepo.git
// 同范式；认证走部署侧凭据不入库）。
func (e *Exporter) git(ctx context.Context, args ...string) (string, error) {
	cmd := exec.CommandContext(ctx, "git", args...)
	cmd.Dir = e.dir
	cmd.Env = append(os.Environ(),
		"GIT_AUTHOR_NAME=cored", "GIT_AUTHOR_EMAIL=cored@localhost",
		"GIT_COMMITTER_NAME=cored", "GIT_COMMITTER_EMAIL=cored@localhost",
	)
	var out, errOut bytes.Buffer
	cmd.Stdout, cmd.Stderr = &out, &errOut
	if err := cmd.Run(); err != nil {
		return strings.TrimSpace(out.String()), fmt.Errorf("git %s: %w（%s）",
			strings.Join(args, " "), err, strings.TrimSpace(errOut.String()))
	}
	return strings.TrimSpace(out.String()), nil
}
