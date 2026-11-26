// config.go —— runtime.taskScheduler 配置解析（scheme 声明面热读，Apply 即时生效）。
//
// 设计依据：core_design/core/scheme.yaml runtime.taskScheduler（enabled/maxConcurrent/
// perService/perRepo/strategy/conflictPolicy/leaseTtl/maxAttempts/spawn）与
// runtime/task-scheduler.md §3「配置来自 scheme.runtime.taskScheduler，Apply 即时生效
// （调度器热读）」。每拍现读零缓存（Scheme.Replace 原子热刷新后下拍即生效——与对象声明表
// 投影同口径）；strategy/conflictPolicy 当前唯一实现即 independent-first /
// serialize-overlap（声明值不参与判定，只透传 stderr 装配锚——未来策略面扩展点）。
package scheduler

import (
	"strconv"
	"strings"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// Config 调度器生效配置（scheme 声明 + 缺省兜底——宽松方向缺键回落声明缺省，不发明数值）。
type Config struct {
	Enabled           bool
	MaxConcurrent     int           // 全局并发上限
	PerService        int           // 同服务并发上限（键=scopeRef 归一）
	PerRepo           int           // 同仓并发上限（键=touches.repos；缺省 1=串行）
	LeaseTTL          time.Duration // 任务租约（防僵尸）
	MaxAttempts       int           // 失败重试上限（超限=failed）
	Spawn             string        // 宿主 spawn 桥声明锚（host:/task/spawn；真实地址走 env）
	ApplyConflictGate string        // Apply 关口冲突拦截词形（ask(pause-first|abort-current)——BUG-013）
}

// ConfigDefaults scheme 声明缺省（runtime.taskScheduler 注释同源；缺键=本值不发明）。
func ConfigDefaults() Config {
	return Config{
		MaxConcurrent:     4,
		PerService:        2,
		PerRepo:           1,
		LeaseTTL:          30 * time.Minute,
		MaxAttempts:       2,
		Spawn:             "host:/task/spawn",
		ApplyConflictGate: "ask(pause-first|abort-current)",
	}
}

// ConfigOf 从 scheme 热读生效配置（nil/缺区=全缺省且 enabled=false——关断是缺省行为面）。
func ConfigOf(s *scheme.Scheme) Config {
	c := ConfigDefaults()
	if s == nil {
		c.Enabled = false
		return c
	}
	rt, _ := s.Spec.Runtime["taskScheduler"].(map[string]any)
	if rt == nil {
		return c
	}
	c.Enabled = asBool(rt["enabled"], false)
	c.MaxConcurrent = asInt(rt["maxConcurrent"], c.MaxConcurrent)
	c.PerService = asInt(rt["perService"], c.PerService)
	c.PerRepo = asInt(rt["perRepo"], c.PerRepo)
	c.LeaseTTL = asDuration(rt["leaseTtl"], c.LeaseTTL)
	c.MaxAttempts = asInt(rt["maxAttempts"], c.MaxAttempts)
	c.Spawn = asString(rt["spawn"], c.Spawn)
	c.ApplyConflictGate = asString(rt["applyConflictGate"], c.ApplyConflictGate)
	// 上限声明只收紧语义的正数兜底（0/负=声明残缺，回落缺省而非关断——保守并发不静默放开）。
	if c.MaxConcurrent <= 0 {
		c.MaxConcurrent = ConfigDefaults().MaxConcurrent
	}
	if c.PerService <= 0 {
		c.PerService = ConfigDefaults().PerService
	}
	if c.PerRepo <= 0 {
		c.PerRepo = ConfigDefaults().PerRepo
	}
	if c.LeaseTTL <= 0 {
		c.LeaseTTL = ConfigDefaults().LeaseTTL
	}
	if c.MaxAttempts <= 0 {
		c.MaxAttempts = ConfigDefaults().MaxAttempts
	}
	return c
}

func asBool(v any, def bool) bool {
	switch t := v.(type) {
	case bool:
		return t
	case string:
		b, err := strconv.ParseBool(strings.TrimSpace(t))
		if err == nil {
			return b
		}
	}
	return def
}

func asInt(v any, def int) int {
	switch t := v.(type) {
	case int:
		return t
	case int64:
		return int(t)
	case float64:
		return int(t)
	case string:
		n, err := strconv.Atoi(strings.TrimSpace(t))
		if err == nil {
			return n
		}
	}
	return def
}

func asString(v any, def string) string {
	if s, ok := v.(string); ok && strings.TrimSpace(s) != "" {
		return strings.TrimSpace(s)
	}
	return def
}

// asDuration 声明词形=Go duration 字面量（30m/45m——scheme 注释同源）；非词形回落缺省。
func asDuration(v any, def time.Duration) time.Duration {
	if s, ok := v.(string); ok {
		if d, err := time.ParseDuration(strings.TrimSpace(s)); err == nil {
			return d
		}
	}
	if d, ok := v.(time.Duration); ok {
		return d
	}
	return def
}
