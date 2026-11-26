// scope_resolver.go —— Agent 事件 scope 声明解析器（BUG-010 派发闸/消费闸同源提取；R-ENG-7 拆分）。
package main

import (
	"context"
	"strings"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/flowbind"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"gopkg.in/yaml.v3"
)

// agentScopeResolverFor Agent 事件 scope 声明解析（文档库现读投影——BUG-010 派发闸/消费闸
// 共用同一实例零双源）：Agent 锚名 `<p>-agent`/root/parent/<p> 取末段匹配 Agent 对象名；
// 对象 spec.eventScopes=允许的 scopePattern（通配 *），own=对象自身 scope（保守缺省第二收窄项）；
// 解析失败按未声明处理（闸走保守缺省，宁拒不降级）。
func agentScopeResolverFor(repo repository.Repository) flowbind.AgentEventScopes {
	return flowbind.AgentEventScopesFunc(func(agent string) ([]string, string, bool) {
		objs, err := repo.List(context.Background(), "Agent", "")
		if err != nil {
			return nil, "", false
		}
		target := agent
		if i := strings.LastIndex(target, "/"); i >= 0 {
			target = target[i+1:]
		}
		for _, o := range objs {
			if o.Name != target {
				continue
			}
			var doc struct {
				Spec struct {
					EventScopes []string `yaml:"eventScopes"`
				} `yaml:"spec"`
			}
			if err := yaml.Unmarshal(o.Content, &doc); err != nil {
				return nil, o.Scope, true // 文档坏=未声明声明面（闸保守缺省）
			}
			return doc.Spec.EventScopes, o.Scope, true
		}
		return nil, "", false
	})
}
