// Package refresolve 声明式逻辑引用的运行时解析（v0.10，A46）。
// 设计依据：core_system/concepts/ref.md。
// Ref 只声明"要什么"（resolve.system/type/params）；真实地址由 owning system
// 管理并返回（如 onedev 源码地址由 developer_system 管）——以此解耦各系统。
package refresolve

import (
	"context"
	"fmt"
	"time"
)

// Spec Ref.spec.resolve：声明"要什么"。
type Spec struct {
	System string            // resolver：由哪个 system desc 解析
	Type   string            // code | doc | url | resource | service
	Params map[string]string // 如 project: onedev
}

// Resolution 解析结果（含缓存）。
type Resolution struct {
	Value    string    // 真实地址/实例（只存在于这里，不进 Ref.spec）
	By       string    // 解析方（system/instance）
	At       time.Time
	CacheTTL time.Duration
}

// Resolver 由 owning system 实现的解析能力。
type Resolver interface {
	CanResolve(system, typ string) bool
	// Resolve 返回真实地址；解析链深度 ≤2，检环。
	Resolve(ctx context.Context, s Spec, depth int) (*Resolution, error)
}

// Engine 解析引擎：路由到 owning system + 缓存。
type Engine struct {
	Resolvers []Resolver
	cache     map[string]Resolution
}

// Resolve 入口：spec → owning system → 真实地址（缓存期内直接复用）。
func (e *Engine) Resolve(ctx context.Context, s Spec) (*Resolution, error) {
	key := s.System + "/" + s.Type + "/" + fmt.Sprint(s.Params)
	if r, ok := e.cache[key]; ok && time.Since(r.At) < r.CacheTTL {
		return &r, nil
	}
	for _, r := range e.Resolvers {
		if r.CanResolve(s.System, s.Type) {
			out, err := r.Resolve(ctx, s, 1)
			if err != nil {
				return nil, fmt.Errorf("refresolve: %w (resolver unavailable -> Event + 降级旧缓存)", err)
			}
			e.cache[key] = *out
			return out, nil
		}
	}
	return nil, fmt.Errorf("refresolve: no resolver for %s/%s", s.System, s.Type)
}
