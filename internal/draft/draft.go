// Package draft 草稿缓存——编辑的家（core_design/concepts/apply.md §1/§6 rule 6）：
// 代改代审的改动先落草稿，不触碰文档面（无 Revision/无事件）；apply 落库那一刻才算数。
// 落点：进程内存为底（单进程单写者）；持久化面可选装配（Persistence）——--ops-db 持久化
// 部署时落稿/弃稿写穿 ops DB draft 表 + 启动恢复（ops-log.md §2），缺省纯内存语义不变。
// 语义锚：草稿可查/可弃，apply 后清空；同对象键覆盖=最新代改稿；baseResourceVersion=
// 落稿时读到的文档面版本（apply-draft 的乐观锁期望值，冲突保留草稿待重整）。
package draft

import (
	"errors"
	"fmt"
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

// Persistence 草稿持久化面（消费方接口）：draft 包不依赖具体存储——由装配方注入实现
// （v1 落点=ops DB draft 表，ops-log.md §2；本包与存储实现保持叶子解耦，映射在装配层）。
// 语义=写穿缓存：Put/Delete 先持久化后落内存（持久化成功才算落稿成功——重启恢复不丢稿）。
type Persistence interface {
	// PutDraft 写穿一条草稿（upsert，同键覆盖；CreatedAt 定格由调用方保证）。
	PutDraft(d *Draft) error
	// DeleteDraft 删一条草稿（键不存在=幂等成功——缓存/存储短暂不一致时写穿不放大错误）。
	DeleteDraft(kind, scope, name string) error
	// ListDrafts 全量草稿（启动恢复唯一入口；含正文与时间戳）。
	ListDrafts() ([]*Draft, error)
}

// Options 草稿缓存装配项。
type Options struct {
	// Persistence 持久化面（nil=纯内存，缺省语义不变：重启即清，草稿可重放）。
	Persistence Persistence
}

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

// Store 草稿缓存（单进程单写者语义由互斥锁保证；cored 生命周期即草稿生命周期——
// 持久化部署除外：写穿面让草稿跨重启存活，恢复走 Restore）。
type Store struct {
	mu sync.Mutex
	m  map[string]*Draft
	p  Persistence // nil=纯内存
}

// NewStore 建草稿缓存。opts 为 nil 或 Options.Persistence 为 nil → 纯内存（缺省语义不变）。
func NewStore(opts *Options) *Store {
	s := &Store{m: make(map[string]*Draft)}
	if opts != nil {
		s.p = opts.Persistence
	}
	return s
}

// Put 落稿（upsert）：同对象键覆盖=最新代改稿。首次落稿定格 CreatedAt，覆盖只动 UpdatedAt。
// 持久化面在位时先写穿后落内存（持久化失败=落稿失败、内存不动——可见即已持久）。
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
	if s.p != nil {
		if err := s.p.PutDraft(d.clone()); err != nil {
			return fmt.Errorf("draft: 写穿落稿: %w", err)
		}
	}
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
	if s.p != nil {
		if err := s.p.DeleteDraft(kind, scope, name); err != nil {
			return fmt.Errorf("draft: 写穿弃稿: %w", err)
		}
	}
	delete(s.m, k)
	return nil
}

// Restore 从持久化面恢复草稿（cored 重启时调用——--ops-db 给定即恢复；整表替换内存，
// 确定性语义：恢复后的内存态=持久化面现态）。返回恢复条数与跳过条数：
// 存储级读错 → err（fail-fast：持久化部署读不开自己的稿=部署坏了）；单行非法（键不全/
// 无主/空稿）→ 跳过并计数（草稿非权威可重放，坏行不拖垮进程；跳过数供探针取证）。
// 无持久化面 → ErrNoStore（纯内存部署没有恢复语义可言）。
func (s *Store) Restore() (restored, skipped int, err error) {
	if s == nil {
		return 0, 0, ErrNoStore
	}
	if s.p == nil {
		return 0, 0, ErrNoStore
	}
	rows, err := s.p.ListDrafts()
	if err != nil {
		return 0, 0, fmt.Errorf("draft: 恢复读取: %w", err)
	}
	m := make(map[string]*Draft, len(rows))
	for _, d := range rows {
		if !d.valid() {
			skipped++
			continue
		}
		d.CreatedAt = d.CreatedAt.UTC()
		d.UpdatedAt = d.UpdatedAt.UTC()
		m[d.key()] = d.clone()
		restored++
	}
	s.mu.Lock()
	s.m = m
	s.mu.Unlock()
	return restored, skipped, nil
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
