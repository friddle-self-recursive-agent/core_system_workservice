// runtime_store.go —— 运行态投影存储装配（G-01 收口——R-ENG §1/perf-lint-baseline.md §2.2）：
// `--runtime-store sqlite|memory` 驱动选择；**缺省 sqlite 当 --ops-db 给定**（持久化部署显式
// 传路径才落盘，缺省内存语义不变）；rt_* 表族建在 ops DB 同库（--runtime-db 独立库=非缺省
// 可选路径）；cored 启动时全量装载重建投影（装载在各投影落点构造器内——workflow
// TriggerHub/RunLog/DeliveryLog/WakeLog、flowbind DispatchLog、intake Engine 各自 New 时
// 从 Store 读历史行续接序号/游标）。判据锚：写入→重启→GET 读面数据仍在（sqlite rt_* 可查）。
package main

import (
	"fmt"
	"os"
	"strings"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/runtimestore"
)

// assembleRuntimeStore 运行态投影存储装配：driver=旗标词形（空=缺省判定——sqlite 当
// --ops-db 给定，否则 memory）；dsn=独立库路径（--runtime-db）或 ops DB 同库（缺省）。
// memory 驱动忽略 dsn（进程内切片——测试/无 DB 部署兜底）。未知驱动词形=启动拒绝（缺口
// 可见不静默——驱动注册表 Open 同口径）。
func assembleRuntimeStore(driver, dsn string) (runtimestore.Store, error) {
	store, err := runtimestore.Open(driver, dsn)
	if err != nil {
		return nil, fmt.Errorf("运行态投影存储装配（--runtime-store %s）: %w", driver, err)
	}
	return store, nil
}

// runtimeStoreChoice 旗标缺省判定与 stderr 装配锚词形（机械判别器：sqlite(<path>)=落盘
// 持久化（重启恢复）；memory=进程内既有行为面零翻转）。
func runtimeStoreChoice(flagVal, opsDB, runtimeDB string) (driver, dsn, state string) {
	driver = strings.TrimSpace(flagVal)
	switch driver {
	case "sqlite", "memory":
	case "":
		if strings.TrimSpace(opsDB) != "" {
			driver = "sqlite" // 缺省：持久化部署（--ops-db 给定）落盘
		} else {
			driver = "memory"
		}
	default:
		return "", "", "" // 未知词形——调用方启动拒绝
	}
	dsn = strings.TrimSpace(runtimeDB)
	if dsn == "" {
		dsn = strings.TrimSpace(opsDB) // rt_* 表族建在 ops DB 同库（非缺省独立库=--runtime-db）
	}
	if driver == "memory" {
		state = "memory（进程内既有语义——重启即清；--runtime-store sqlite 显式落盘）"
		return driver, "", state
	}
	if dsn == "" {
		dsn = ":memory:" // sqlite+无路径=进程内存库（等价 memory 的落盘词形缺席）
	}
	state = "sqlite(" + dsn + ")（rt_* 六表写穿透+启动全量装载重建投影——写入→重启→GET 读面数据仍在）"
	return driver, dsn, state
}

// runtimeStoreStateOr 状态词兜底（装配锚 stderr 报告用——未装配=nil 词形诚实投影）。
func runtimeStoreStateOr(s runtimestore.Store, state string) string {
	if s == nil {
		return "off（未装配——投影纯内存）"
	}
	return state
}

// envOr 环境变量读取兜底（runtime store 驱动词形也可经 CORE_SYSTEM_RUNTIME_STORE 注入——
// 部署锚点经参数/环境注入的既有口径；旗标显式值优先）。
func envOr(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
