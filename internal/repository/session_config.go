// session_config.go —— 三段式 Session 控制面存储（session.md §5/§8；scheme.yaml
// runtime.context.sessionStages/sessionBackground/sessionMeta 的数据库承载行）。
// GET/PUT /v1/session-config 读写（PUT 随 --allow-apply 门控）；第三段·属性回整
// （apply 触发）的账目落 session_writeback 表——规则版默认自动（importance +5 封顶
// 100 + 会话 tag 并入侧车），LLM 判断回整=declared-deferred，Mode 字段诚实标注。
package repository

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

// sidecarImportanceRe 侧车 importance 行（metadatainfo 侧车词形「importance: 82」）。
var sidecarImportanceRe = regexp.MustCompile(`^(\s*importance:\s*)(\d+)\s*$`)

// SessionStages 三段描述（页面可编辑——「我想控制一些描述」的承载面）。
type SessionStages struct {
	S1 string `json:"s1"` // 第一段·背景装配描述
	S2 string `json:"s2"` // 第二段·预更新执行描述
	S3 string `json:"s3"` // 第三段·属性回整描述
}

// SessionConfig 三段式 Session 控制面（单行存储；缺省=scheme.yaml 声明值硬对齐）。
type SessionConfig struct {
	Stages             SessionStages `json:"stages"`
	BackgroundMaxChars int           `json:"backgroundMaxChars"` // 背景默认最多字数（5000）
	BlockAttribution   bool          `json:"blockAttribution"`   // 背景逐块标识资源/文件（默认 true）
	MetaAutoUpdate     bool          `json:"metaAutoUpdate"`     // 第三段属性回整默认自动（默认 true）
	MetaScope          []string      `json:"metaScope"`          // 回整属性域（ref/tag/importance/metainfo）
	Provider           string        `json:"provider"`           // 会话默认 Provider（openai-default）
	Model              string        `json:"model"`              // 会话默认模型（reasoning-tier）
	UpdatedAt          string        `json:"updatedAt"`
}

// SessionWriteback 第三段回整账行（apply 触发一次一行；界面「自动回整」来源标记的数据源）。
type SessionWriteback struct {
	At              string   `json:"at"`
	Resource        string   `json:"resource"`        // <repo>/<path>
	Subject         string   `json:"subject"`         // 触发主体（apply subject 名）
	SessionID       string   `json:"sessionId,omitempty"`
	Mode            string   `json:"mode"`            // rule-default（LLM 回整 declared-deferred）
	ImportanceDelta int      `json:"importanceDelta"` // 实际侧车 importance 增量（0=侧车缺席/未动）
	TagsAdded       []string `json:"tagsAdded"`       // 并入侧车的会话 tag
	Note            string   `json:"note,omitempty"`
}

// SessionConfigStore 三段控制面存储接口（SQLiteRepo 实现；缺席=GET 返回缺省不冒充持久化）。
type SessionConfigStore interface {
	GetSessionConfig(ctx context.Context) (SessionConfig, error)
	SetSessionConfig(ctx context.Context, cfg SessionConfig) (SessionConfig, error)
	AppendSessionWriteback(ctx context.Context, wb SessionWriteback) error
	LastSessionWritebacks(ctx context.Context, n int) ([]SessionWriteback, error)
}

// DefaultSessionConfig scheme.yaml runtime.context 声明值硬对齐（存储未初始化/读失败时的诚实缺省）。
func DefaultSessionConfig() SessionConfig {
	return SessionConfig{
		Stages: SessionStages{
			S1: "第一段·背景装配：LLM 描述命令 → tag 分析先查 CtxCache → 三路注入（工具使用 / service 描述及 tag / global index 及 tag）→ builtin RAG 查询 → 类似 session 导入描述 → developer_system 定位代码仓库 → 层层整理 → 背景 ≤5k 字逐块溯源",
			S2: "第二段·预更新执行：反复 ask 询问；会话没有命令——能做的永远是文档预更新（draft）；命令都到 workflow 的 agent 节点执行；等待用户 apply",
			S3: "第三段·属性回整：apply 触发——LLM 判断哪些文档 ref/tag/重要程度/metainfo 要改，默认自动更新；内容类 Proposal 仍人审",
		},
		BackgroundMaxChars: 5000,
		BlockAttribution:   true,
		MetaAutoUpdate:     true,
		MetaScope:          []string{"ref", "tag", "importance", "metainfo"},
		Provider:           "openai-default",
		Model:              "reasoning-tier",
		UpdatedAt:          "",
	}
}

// ensureSessionConfigTables 建 session_config（单行）与 session_writeback（账目）两表。
func (r *SQLiteRepo) ensureSessionConfigTables(ctx context.Context) error {
	const ddlCfg = `CREATE TABLE IF NOT EXISTS session_config (
		id                    INTEGER PRIMARY KEY CHECK (id = 1),
		stage_s1              TEXT NOT NULL DEFAULT '',
		stage_s2              TEXT NOT NULL DEFAULT '',
		stage_s3              TEXT NOT NULL DEFAULT '',
		background_max_chars  INTEGER NOT NULL DEFAULT 5000,
		block_attribution     INTEGER NOT NULL DEFAULT 1,
		meta_auto_update      INTEGER NOT NULL DEFAULT 1,
		meta_scope            TEXT NOT NULL DEFAULT '',
		provider              TEXT NOT NULL DEFAULT 'openai-default',
		model                 TEXT NOT NULL DEFAULT 'reasoning-tier',
		updated_at            TEXT NOT NULL DEFAULT ''
	);`
	const ddlWb = `CREATE TABLE IF NOT EXISTS session_writeback (
		id               INTEGER PRIMARY KEY AUTOINCREMENT,
		at               TEXT NOT NULL,
		resource         TEXT NOT NULL,
		subject          TEXT NOT NULL DEFAULT '',
		session_id       TEXT NOT NULL DEFAULT '',
		mode             TEXT NOT NULL DEFAULT 'rule-default',
		importance_delta INTEGER NOT NULL DEFAULT 0,
		tags_added       TEXT NOT NULL DEFAULT '',
		note             TEXT NOT NULL DEFAULT ''
	);`
	if _, err := r.db.ExecContext(ctx, ddlCfg); err != nil {
		return err
	}
	if _, err := r.db.ExecContext(ctx, ddlWb); err != nil {
		return err
	}
	def := DefaultSessionConfig()
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx,
		`INSERT OR IGNORE INTO session_config (id, stage_s1, stage_s2, stage_s3, background_max_chars, block_attribution, meta_auto_update, meta_scope, provider, model, updated_at)
		 VALUES (1, ?, ?, ?, ?, 1, 1, ?, ?, ?, ?)`,
		def.Stages.S1, def.Stages.S2, def.Stages.S3, def.BackgroundMaxChars,
		strings.Join(def.MetaScope, ","), def.Provider, def.Model, now)
	return err
}

// GetSessionConfig 读三段控制面（表/行缺席=缺省值+error 随行——调用方决定是否降级词形）。
func (r *SQLiteRepo) GetSessionConfig(ctx context.Context) (SessionConfig, error) {
	if err := r.ensureSessionConfigTables(ctx); err != nil {
		return DefaultSessionConfig(), err
	}
	var s1, s2, s3, scope, provider, model, updatedAt string
	var maxChars, blockAttr, autoUpdate int
	err := r.db.QueryRowContext(ctx,
		`SELECT stage_s1, stage_s2, stage_s3, background_max_chars, block_attribution, meta_auto_update, meta_scope, provider, model, updated_at FROM session_config WHERE id = 1`).
		Scan(&s1, &s2, &s3, &maxChars, &blockAttr, &autoUpdate, &scope, &provider, &model, &updatedAt)
	if err != nil {
		return DefaultSessionConfig(), err
	}
	def := DefaultSessionConfig()
	cfg := SessionConfig{
		Stages:             SessionStages{S1: s1, S2: s2, S3: s3},
		BackgroundMaxChars: maxChars,
		BlockAttribution:   blockAttr != 0,
		MetaAutoUpdate:     autoUpdate != 0,
		MetaScope:          splitCSV(scope),
		Provider:           provider,
		Model:              model,
		UpdatedAt:          updatedAt,
	}
	if len(cfg.MetaScope) == 0 {
		cfg.MetaScope = def.MetaScope
	}
	return cfg, nil
}

// SetSessionConfig 全量覆盖三段控制面（PUT 语义；返回存储拍现行值）。
func (r *SQLiteRepo) SetSessionConfig(ctx context.Context, cfg SessionConfig) (SessionConfig, error) {
	if err := r.ensureSessionConfigTables(ctx); err != nil {
		return SessionConfig{}, err
	}
	if cfg.BackgroundMaxChars <= 0 {
		cfg.BackgroundMaxChars = DefaultSessionConfig().BackgroundMaxChars
	}
	if len(cfg.MetaScope) == 0 {
		cfg.MetaScope = DefaultSessionConfig().MetaScope
	}
	now := time.Now().UTC().Format(time.RFC3339)
	_, err := r.db.ExecContext(ctx,
		`UPDATE session_config SET stage_s1=?, stage_s2=?, stage_s3=?, background_max_chars=?, block_attribution=?, meta_auto_update=?, meta_scope=?, provider=?, model=?, updated_at=? WHERE id = 1`,
		cfg.Stages.S1, cfg.Stages.S2, cfg.Stages.S3, cfg.BackgroundMaxChars,
		boolInt(cfg.BlockAttribution), boolInt(cfg.MetaAutoUpdate),
		strings.Join(cfg.MetaScope, ","), cfg.Provider, cfg.Model, now)
	if err != nil {
		return SessionConfig{}, err
	}
	return r.GetSessionConfig(ctx)
}

// AppendSessionWriteback 追加第三段回整账行。
func (r *SQLiteRepo) AppendSessionWriteback(ctx context.Context, wb SessionWriteback) error {
	if err := r.ensureSessionConfigTables(ctx); err != nil {
		return err
	}
	if wb.At == "" {
		wb.At = time.Now().UTC().Format(time.RFC3339)
	}
	if wb.Mode == "" {
		wb.Mode = "rule-default"
	}
	_, err := r.db.ExecContext(ctx,
		`INSERT INTO session_writeback (at, resource, subject, session_id, mode, importance_delta, tags_added, note) VALUES (?, ?, ?, ?, ?, ?, ?, ?)`,
		wb.At, wb.Resource, wb.Subject, wb.SessionID, wb.Mode, wb.ImportanceDelta,
		strings.Join(wb.TagsAdded, ","), wb.Note)
	return err
}

// LastSessionWritebacks 最近 n 条回整账（新→旧）。
func (r *SQLiteRepo) LastSessionWritebacks(ctx context.Context, n int) ([]SessionWriteback, error) {
	if err := r.ensureSessionConfigTables(ctx); err != nil {
		return nil, err
	}
	if n <= 0 {
		n = 10
	}
	rows, err := r.db.QueryContext(ctx,
		`SELECT at, resource, subject, session_id, mode, importance_delta, tags_added, note FROM session_writeback ORDER BY id DESC LIMIT ?`, n)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := make([]SessionWriteback, 0, n)
	for rows.Next() {
		var wb SessionWriteback
		var tags string
		if err := rows.Scan(&wb.At, &wb.Resource, &wb.Subject, &wb.SessionID, &wb.Mode, &wb.ImportanceDelta, &tags, &wb.Note); err != nil {
			return nil, err
		}
		wb.TagsAdded = splitCSV(tags)
		out = append(out, wb)
	}
	return out, rows.Err()
}

// BumpSidecarImportance 第三段规则版侧车手术（纯函数）：importance 行 +delta 封顶 100。
// 返回 (新内容, 实际增量, found)；侧车无 importance 行=found=false 原样返回。
func BumpSidecarImportance(content string, delta int) (string, int, bool) {
	lines := strings.Split(content, "\n")
	for i, ln := range lines {
		if !sidecarImportanceRe.MatchString(ln) {
			continue
		}
		m := sidecarImportanceRe.FindStringSubmatch(ln)
		var cur int
		_, _ = fmt.Sscanf(m[2], "%d", &cur)
		next := cur + delta
		if next > 100 {
			next = 100
		}
		if next < 0 {
			next = 0
		}
		lines[i] = m[1] + fmt.Sprintf("%d", next)
		return strings.Join(lines, "\n"), next - cur, true
	}
	return content, 0, false
}

// MergeSidecarTags 第三段规则版侧车手术（纯函数）：会话 tag 并入 tags: 块（tag_type: session，
// 已存在即跳过；无 tags: 块则在 importance 行前插入新块；两缺=文末追加）。返回 (新内容, 实际并入)。
func MergeSidecarTags(content string, tags []string) (string, []string) {
	var add []string
	for _, t := range tags {
		key := strings.TrimSpace(t)
		if key == "" || strings.Contains(content, "{tag: "+key+",") || strings.Contains(content, "{tag: "+key+"}") {
			continue
		}
		add = append(add, key)
	}
	if len(add) == 0 {
		return content, nil
	}
	sort.Strings(add)
	lines := strings.Split(content, "\n")
	var entries []string
	for _, t := range add {
		entries = append(entries, "  - {tag: "+t+", tag_type: session}")
	}
	if has, last := sidecarTagsTail(lines); has {
		out := append([]string{}, lines[:last+1]...)
		out = append(out, entries...)
		out = append(out, lines[last+1:]...)
		return strings.Join(out, "\n"), add
	}
	insIdx := -1
	for i, ln := range lines {
		if strings.HasPrefix(ln, "importance:") {
			insIdx = i
			break
		}
	}
	if insIdx >= 0 {
		out := append([]string{}, lines[:insIdx]...)
		out = append(out, "tags:")
		out = append(out, entries...)
		out = append(out, lines[insIdx:]...)
		return strings.Join(out, "\n"), add
	}
	out := append([]string{}, lines...)
	out = append(out, "tags:")
	out = append(out, entries...)
	return strings.Join(out, "\n"), add
}

// sidecarTagsTail 定位 tags: 块最后一个条目行（缩进 "  -" 词形）；无块=has=false。
func sidecarTagsTail(lines []string) (bool, int) {
	has := false
	last := -1
	for i, ln := range lines {
		if strings.HasPrefix(ln, "tags:") {
			has = true
			last = i
			continue
		}
		if has && strings.HasPrefix(ln, "  - ") {
			last = i
		}
	}
	return has, last
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
