// Package draft 草稿缓存——编辑的家（core_design/concepts/apply.md §1/§6 rule 6）：
// 代改代审的改动先落草稿，不触碰文档面（无 Revision/无事件）；apply 落库那一刻才算数。
// 落点：进程内存（v1 骨架——重启即清，草稿可重放；持久化随 ops DB 部署决策另行 declared）。
// 语义锚：草稿可查/可弃，apply 后清空；同对象键覆盖=最新代改稿；baseResourceVersion=
// 落稿时读到的文档面版本（apply-draft 的乐观锁期望值，冲突保留草稿待重整）。
package draft

import (
	"errors"
	"sort"
	"sync"
	"time"
)

var (
	// ErrNoDraft 草稿不存在（读/弃/apply 一个没有落过稿的对象键）。
	ErrNoDraft = errors.New("draft: 草稿不存在")
	// ErrInvalid 草稿不合法：对象键不全 / 无主（fieldManager 空）/ 空内容。
	ErrInvalid = errors.New("draft: 草稿不合法（需 kind、scope、name、fieldManager、非空 content）")
	// ErrNoStore 草稿缓存未配置（引擎/读面未接线）。
	ErrNoStore = errors.New("draft: 草稿缓存未配置")
)

// Draft 一条待确认的代改稿（草稿不是文档面：落库前无 Revision 语义，BaseRV 只是依据版本）。
type Draft struct {
	Kind         string
	Scope        string
	Name         string
	Content      []byte // 文档原文（yaml 或 md+frontmatter），apply 时按存储层语义解析
	FieldManager string // 代改者（落稿溯源；apply 落库时另按写路径 rule 2 校验其自身 fieldManager）
	ReasonRef    string // 指向 DriftReport/Event/工单（审计回链）
	BaseRV       int64  // 落稿所基于的文档面 resourceVersion（0=预期创建）
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// objKey 草稿键（同文档面：kind/scope/name；scope 入键——同名对象可并存多 scope）。
func objKey(kind, scope, name string) string {
	return kind + "/" + scope + "/" + name
}

func (d *Draft) key() string { return objKey(d.Kind, d.Scope, d.Name) }

// valid 键全 + 无主拒绝 + 空稿拒绝（空草稿无落库意义——apply.md rule 1「空草稿=无操作」
// 的反向闸门：落稿面直接拒绝无内容的稿，apply 面的 noop 只来自「稿与面一致」）。
func (d *Draft) valid() bool {
	return d.Kind != "" && d.Scope != "" && d.Name != "" &&
		d.FieldManager != "" && len(d.Content) > 0
}

// clone 深拷贝（返回给读面，防止调用方改动 map 内正文）。
func (d *Draft) clone() *Draft {
	c := *d
	c.Content = append([]byte(nil), d.Content...)
	c.CreatedAt = d.CreatedAt.UTC()
	c.UpdatedAt = d.UpdatedAt.UTC()
	return &c
}

// Store 进程内存草稿缓存（单进程单写者语义由互斥锁保证；cored 生命周期即草稿生命周期）。
type Store struct {
	mu sync.Mutex
	m  map[string]*Draft
}

// NewStore 建空草稿缓存。
func NewStore() *Store {
	return &Store{m: make(map[string]*Draft)}
}

// Put 落稿（upsert）：同对象键覆盖=最新代改稿。首次落稿定格 CreatedAt，覆盖只动 UpdatedAt。
func (s *Store) Put(d *Draft) error {
	if s == nil || d == nil || !d.valid() {
		return ErrInvalid
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	now := time.Now().UTC()
	if old, ok := s.m[d.key()]; ok {
		d.CreatedAt = old.CreatedAt
	} else {
		d.CreatedAt = now
	}
	d.UpdatedAt = now
	s.m[d.key()] = d.clone()
	return nil
}

// Get 读单条草稿（含正文）；无稿 → ErrNoDraft。
func (s *Store) Get(kind, scope, name string) (*Draft, error) {
	if s == nil {
		return nil, ErrNoStore
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	d, ok := s.m[objKey(kind, scope, name)]
	if !ok {
		return nil, ErrNoDraft
	}
	return d.clone(), nil
}

// List 全量草稿（按对象键排序——确定性读面，机械对账可复现）；无正文瘦身由读面层裁剪。
func (s *Store) List() []*Draft {
	if s == nil {
		return nil
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make([]*Draft, 0, len(s.m))
	for _, d := range s.m {
		out = append(out, d.clone())
	}
	sort.Slice(out, func(i, j int) bool { return out[i].key() < out[j].key() })
	return out
}

// Delete 弃稿（人裁不通过=弃；apply 成功后也走本方法清空——apply.md rule 6）。无稿 → ErrNoDraft。
func (s *Store) Delete(kind, scope, name string) error {
	if s == nil {
		return ErrNoStore
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	k := objKey(kind, scope, name)
	if _, ok := s.m[k]; !ok {
		return ErrNoDraft
	}
	delete(s.m, k)
	return nil
}

// Len 草稿数（诊断与对账探针用）。
func (s *Store) Len() int {
	if s == nil {
		return 0
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	return len(s.m)
}
