// refresh.go 权限对象事件驱动热装载（免重启，v0.10）：变更信号 → 重投影 → LoadObjects 整体替换快照。
// 落点依据：instance 仓 runtime/cored-bridge.md §1「写面开关」行（授权面刷新语义，declared）。
// 解耦：本包不依赖存储层——变更信号 =「文档面可能已前移」（repository.Watch 或任意事件流适配成
// 空载荷信号），投影由调用方注入（Load），permission 包只管装载与快照替换。
package permission

import (
	"context"
	"time"
)

// RefreshResult 一次热装载的终态（观测面：cored 以此打 stderr 机器行）。
type RefreshResult struct {
	Ok     bool
	Err    error  // Ok=false 时非 nil（投影读错，或装载校验拒绝——坏对象/重名）
	Gen    uint64 // Ok=true 时=新权威快照代次
	Roles  int    // Ok=true 时=快照内 role 数
	Groups int    // Ok=true 时=快照内 group 数
}

// Refresher 热装载循环（Run 阻塞，调用方以 goroutine 承载；ctx 取消即退场）。
//
// 失败语义（declared，cored-bridge.md §1）：**保旧快照 + 告警 + 定时重试自愈**。
// 不半装载（复用 LoadObjects 整体拒绝）、不打死运行中授权面——启动装载仍是整体拒绝启动
// （授权面宁拒不降级），运行时失败只延迟刷新：
//   - 投影读与写面串行写队列不互斥（git 读可能撞上提交窗口的暂态撕读），拒写会让一份坏对象
//     或一次暂态读错放大为全系统写面停摆（自举修复路径本身常是写路径）；
//   - 保旧 = 上一次成功装载的权威快照继续生效（判定留痕一致：Gen 不动）；
//   - 坏对象清除后下一次成功装载自动痊愈；重试定时器保证自愈不依赖后续变更信号。
//
// dispatch/事件订阅接线后，是否升级为拒写闸门由订阅方声明（本包语义不变）。
type Refresher struct {
	Sys *System
	// Load 投影函数：从任意读源取当前 Role/Group 对象集（调用方构造，如 repository→ObjectDoc）。
	// 每次 hot-reload 信号触发一次全量重投影（对象小、频度低，全量最简且无增量漏算）。
	Load func() ([]ObjectDoc, error)
	// Reloaded 每次热装载终态回调（成功/失败都调；nil=静默）。
	Reloaded func(RefreshResult)
	// Retry 失败重试间隔（<=0 取缺省 1s）。
	Retry time.Duration
}

// Run 消费变更信号做热装载直至 ctx 取消。changes 信号无载荷（连续变更合并由信号源负责：
// 每信号触发一次重投影装载，装到的是当下最新面）。返回 ctx.Err()（正常退场）。
func (r *Refresher) Run(ctx context.Context, changes <-chan struct{}) error {
	retry := r.Retry
	if retry <= 0 {
		retry = time.Second
	}
	var timer *time.Timer
	var retryC <-chan time.Time
	disarm := func() {
		if timer != nil {
			timer.Stop()
			timer = nil
		}
		retryC = nil // nil channel：select 分支永久阻塞，等价于摘除
	}
	arm := func() {
		disarm()
		timer = time.NewTimer(retry)
		retryC = timer.C
	}
	defer disarm()
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-changes:
			disarm() // 新变更信号优先于挂起的重试（装一次到最新面即可）
			if !r.refreshOnce() {
				arm()
			}
		case <-retryC:
			disarm()
			if !r.refreshOnce() {
				arm()
			}
		}
	}
}

// refreshOnce 一轮重投影装载。投影读错与装载校验拒绝同途：保旧快照 + 回调告警，返回失败。
func (r *Refresher) refreshOnce() bool {
	res := RefreshResult{}
	docs, err := r.Load()
	if err == nil {
		err = r.Sys.LoadObjects(docs) // 失败=保旧快照（LoadObjects 不半装载）
	}
	if err != nil {
		res.Err = err
	} else {
		snap := r.Sys.Snapshot()
		res.Ok, res.Gen, res.Roles, res.Groups = true, snap.Gen, len(snap.Roles), len(snap.Groups)
	}
	if r.Reloaded != nil {
		r.Reloaded(res)
	}
	return res.Ok
}
