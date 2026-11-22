// Package skill 工具使用的缓存（会话层经验，A38）。
// 设计依据：core_system/concepts/skill.md。缓存不是真相：与 Provider 契约冲突即失效。
package skill

import "time"

// Skill 一份工具使用经验（注入提示+成功调用示例）。
type Skill struct {
	Name      string
	SourceOps string // 来源工具（ConnectorOps 名）
	When      string // 适用条件（任务/场景）
	Prompt    string // 注入提示（派生自 provider.llmPrompt + 实战修正）
	Origin    struct {
		SessionRef string
		CachedAt   time.Time
	}
	CachePolicy struct {
		TTL                        time.Duration
		InvalidateOnProviderChange bool
	}
}

// Store Skill 缓存存取（会话按需加载）。
type Store interface {
	Get(name string) (*Skill, error)
	Put(s *Skill) error
	// Invalidate 工具契约变更/失败率升高/TTL 到期时失效。
	Invalidate(name, reason string) error
}
