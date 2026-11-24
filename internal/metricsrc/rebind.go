// rebind.go apply 事件驱动热重绑（v0.47，event.md 规则 13 注记；console-adoption.md §1.18
// R-4）：阈值声明面的生效面收口——装配期一次枚举（v0.33/v0.45 段）只是初值，此后绑定表随
// apply 落库热替换，免重启。触发面=DocApplied（与 apply 落库同一条总线主干零旁路——落库
// 必发，apply.md rule 1），involved kind=WorkService 精确过滤（阈值声明面只在 WorkService
// 对象 observation.metrics——其他对象 apply 零无效重枚举）。重绑目标=最新面：连续变更积压
// 排干后一次重枚举（合并语义——不逐事件重做）；枚举失败=保旧表+告警回调+下次 apply 重试
// 自愈（权限热装载 v0.10 同款 declared 语义：不半装载、不打死运行中判定面）。
// 依赖方向：本包（信号面）不 import 存储层——全量重枚举由装配层注入闭包（repo 归装配层）。
package metricsrc

import (
	"context"
	"strings"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
)

// workServicePrefix DocApplied involved 的 WorkService kind 前缀（重绑触发过滤——
// involvedRef kind/scope/name 三段同构惯例的首段）。
const workServicePrefix = "WorkService/"

// RebindHooks 热重绑回调（装配层 stderr 装配锚报告用；字段 nil=静默——零值结构体安全）。
type RebindHooks struct {
	Applied func(n int)     // 重绑成功（n=新表绑定数；零=合法等待态）
	Failed  func(err error) // 重绑失败（保旧表——下次 apply 重试自愈）
}

// StartRebind 启动 apply 事件驱动热重绑主循环（装配层 fire-and-forget；ctx 取消即退场，
// 总线订阅随退订函数释放）。enum 为装配层注入的全量重枚举闭包（每次现读文档库——git 后端
// 权威，零缓存口径与 workflowDocLookup 同款）。
func StartRebind(ctx context.Context, bus *eventbus.Bus, b *Bridge, enum func() ([]Source, error), hooks RebindHooks) {
	if bus == nil || b == nil || enum == nil {
		return
	}
	ch, cancel := bus.Subscribe("DocApplied")
	go func() {
		defer cancel()
		for {
			select {
			case <-ctx.Done():
				return
			case ev, ok := <-ch:
				if !ok {
					return
				}
				if !strings.HasPrefix(ev.Involved, workServicePrefix) {
					continue // 非 WorkService 对象 apply 不动阈值绑定表（精确触发）
				}
				drainPending(ch) // 合并积压：连续变更一次重绑到最新面
				srcs, err := enum()
				if err != nil {
					if hooks.Failed != nil {
						hooks.Failed(err)
					}
					continue // 保旧表（下次 apply 重试自愈——不半装载）
				}
				b.Rebind(srcs)
				if hooks.Applied != nil {
					hooks.Applied(len(srcs))
				}
			}
		}
	}()
}

// drainPending 非阻塞排干积压事件（重绑目标=最新面，逐事件重做无意义；频道关=静默返回
// ——退场路径由主循环的 ok 判定承担）。
func drainPending(ch <-chan eventbus.Event) {
	for {
		select {
		case _, ok := <-ch:
			if !ok {
				return
			}
		default:
			return
		}
	}
}
