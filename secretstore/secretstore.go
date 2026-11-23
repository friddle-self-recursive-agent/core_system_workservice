// Package secretstore —— 部署侧秘密存储（SQLite 单文件；用户 2026-11-23 指示：secret 配置
// 要求都可以用 secret 存，凭据脱敏默认走 Secret，运行时态不导出）。
//
// 设计依据（声明面）：
//   - core_design/concepts/provider.md 规则 2/6：secret 只允许引用式取值
//     （secret:/env:/vault:/keychain:），明文永不入库；声明面与运行时面分离——
//     公开仓只见引用（{source: "secret:<name>}），真实值只落部署侧本库。
//   - core_design/core/scheme.yaml runtime.storage：opsLog=SQLite（ops DB 可丢弃可重建），
//     **秘密库独立于 ops DB**——秘密=权威运行时值非缓存，丢库=丢凭据，不能跟着
//     可丢弃的 ops DB 走；独立文件 + 0600 权限（最小暴露面）。
//   - 与文档库（repository）分离：docs DB=文档内容权威，本库=凭据值；两库任何
//     查询/导出/备份口径互不携带对方内容（脱敏红线：值不出本包写路径之外的读面）。
//
// 依赖方向：本包零 workservice 内部依赖（叶子包）——连接器仓可安全引用
// （R-ENG-3 依赖方向恒为 连接器仓 → workservice 导出面，本包即导出面）。
package secretstore

import (
	"database/sql"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"time"

	_ "modernc.org/sqlite" // 纯 Go sqlite 驱动（与 ops DB 同驱动；单写者串行）
)

// ErrNotFound 指定名称的秘密不存在。
var ErrNotFound = errors.New("secretstore: 秘密不存在")

// Store 部署侧秘密存储（SQLite；单写者——连接池收敛为 1）。
type Store struct {
	db   *sql.DB
	path string
}

// DriverName 驱动注册名（modernc.org/sqlite；与 opslog 同驱动同进程只注册一次）。
const DriverName = "sqlite"

// OpenSQLite 打开（必要时建表）秘密库。path 为空=进程内存库（仅探针/测试；
// 持久化部署显式传文件路径——秘密=权威运行时值，丢库=丢凭据）。文件以 0600 创建
// （既有文件权限不动——最小暴露面，不越权放宽）。
func OpenSQLite(path string) (*Store, error) {
	dsn := path
	if path == "" {
		path = ":memory:"
		dsn = path
	} else {
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
		if dir := filepath.Dir(path); dir != "" {
			if err := os.MkdirAll(dir, 0o700); err != nil {
				return nil, fmt.Errorf("secretstore: 建目录 %s: %w", dir, err)
			}
		}
		// 以 0600 预创建文件（sqlite 自建默认随 umask；秘密库取最小权限口径）。
		f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY, 0o600)
		if err != nil {
			return nil, fmt.Errorf("secretstore: 预创建 %s: %w", path, err)
		}
		_ = f.Close()
		dsn = "file:" + path + "?_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)"
	}
	db, err := sql.Open(DriverName, dsn)
	if err != nil {
		return nil, fmt.Errorf("secretstore: 打开 %s: %w", path, err)
	}
	db.SetMaxOpenConns(1) // 单写者
	s := &Store{db: db, path: path}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *Store) migrate() error {
	const ddl = `
CREATE TABLE IF NOT EXISTS secret (
	name       TEXT PRIMARY KEY,
	value      TEXT NOT NULL,
	updated_at TEXT NOT NULL
);`
	if _, err := s.db.Exec(ddl); err != nil {
		return fmt.Errorf("secretstore: 建表: %w", err)
	}
	return nil
}

// Path 秘密库文件路径（部署诊断；空=进程内存库）。
func (s *Store) Path() string { return s.path }

// Close 关闭底层连接。
func (s *Store) Close() error { return s.db.Close() }

func stamp() string { return time.Now().UTC().Format(time.RFC3339Nano) }

// Set 落（或更新）一条秘密值（upsert）。name/value 非空校验；空值=拒绝（空值无凭据语义，
// 拒绝可防误清）。
func (s *Store) Set(name, value string) error {
	if name == "" || value == "" {
		return errors.New("secretstore: name 与 value 皆必填（空值拒绝——清空走 Delete）")
	}
	_, err := s.db.Exec(`INSERT INTO secret (name, value, updated_at) VALUES (?, ?, ?)
ON CONFLICT (name) DO UPDATE SET value=excluded.value, updated_at=excluded.updated_at`, name, value, stamp())
	if err != nil {
		return fmt.Errorf("secretstore: 落秘密(%s): %w", name, err)
	}
	return nil
}

// Delete 删一条秘密；键不存在=幂等成功。
func (s *Store) Delete(name string) error {
	if name == "" {
		return errors.New("secretstore: name 必填")
	}
	if _, err := s.db.Exec(`DELETE FROM secret WHERE name = ?`, name); err != nil {
		return fmt.Errorf("secretstore: 删秘密(%s): %w", name, err)
	}
	return nil
}

// Secret 取一条秘密值（connectors.SecretStore 契约的方法形态；值只经本方法出，
// 任何查询/列表面不携带值——脱敏口径见包注释）。
func (s *Store) Secret(name string) (string, error) {
	if name == "" {
		return "", errors.New("secretstore: name 必填")
	}
	var v string
	err := s.db.QueryRow(`SELECT value FROM secret WHERE name = ?`, name).Scan(&v)
	if errors.Is(err, sql.ErrNoRows) {
		return "", fmt.Errorf("%w: %s", ErrNotFound, name)
	}
	if err != nil {
		return "", fmt.Errorf("secretstore: 读秘密(%s): %w", name, err)
	}
	return v, nil
}

// SecretInfo 秘密的元数据行（**不含值**——列表/查询面的唯一投影形态，脱敏红线）。
type SecretInfo struct {
	Name      string
	UpdatedAt time.Time
}

// List 秘密元数据清单（名称+更新时刻，按名排序；永不携带值）。
func (s *Store) List() ([]SecretInfo, error) {
	rows, err := s.db.Query(`SELECT name, updated_at FROM secret ORDER BY name`)
	if err != nil {
		return nil, fmt.Errorf("secretstore: 枚举秘密: %w", err)
	}
	defer rows.Close()
	var out []SecretInfo
	for rows.Next() {
		var name, ts string
		if err := rows.Scan(&name, &ts); err != nil {
			return nil, fmt.Errorf("secretstore: 读行: %w", err)
		}
		t, err := time.Parse(time.RFC3339Nano, ts)
		if err != nil {
			return nil, fmt.Errorf("secretstore: updated_at 解析 %q: %w", ts, err)
		}
		out = append(out, SecretInfo{Name: name, UpdatedAt: t})
	}
	return out, rows.Err()
}

// Count 秘密行数（诊断与验收探针用）。
func (s *Store) Count() (int, error) {
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM secret`).Scan(&n); err != nil {
		return 0, fmt.Errorf("secretstore: 统计: %w", err)
	}
	return n, nil
}

// Resolve 引用式取值解析（provider.md 规则 2 前缀集合的运行时面）：
//   - "secret:<name>" → 本库取值（部署侧 sqlite，权威运行时面）
//   - "env:<VAR>"     → 进程环境变量（既有口径，明文不经本库）
//   - "vault:…" / "keychain:…" → declared-deferred：本实现不支持，显式报错不猜
//
// 其余形态（含明文）=错误（引用式取值闸门——与 repository 写路径同前缀集合）。
func (s *Store) Resolve(source string) (string, error) {
	if name, ok := cutPrefix(source, "secret:"); ok {
		if name == "" {
			return "", fmt.Errorf("secretstore: secret: 来源缺名称（%q）", source)
		}
		return s.Secret(name)
	}
	if v, ok := cutPrefix(source, "env:"); ok {
		if v == "" {
			return "", fmt.Errorf("secretstore: env: 来源缺变量名（%q）", source)
		}
		if got := os.Getenv(v); got != "" {
			return got, nil
		}
		return "", fmt.Errorf("secretstore: 环境变量未设或为空（%s）", source)
	}
	return "", fmt.Errorf("secretstore: 来源不支持（只允许 secret:/env: 引用；vault:/keychain: declared-deferred；实得 %q）", source)
}

func cutPrefix(s, p string) (string, bool) {
	if len(s) >= len(p) && s[:len(p)] == p {
		return s[len(p):], true
	}
	return "", false
}

// 编译期断言：Store 满足连接器契约层的秘密口面（connectors 包契约——本包不 import
// connectors，方向保持 连接器仓 → workservice 单向）。
var _ interface {
	Secret(name string) (string, error)
} = (*Store)(nil)
