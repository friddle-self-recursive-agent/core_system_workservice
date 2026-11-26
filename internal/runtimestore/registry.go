// registry.go —— 驱动注册表（R-ENG 任务书 §1 registry 件）：Register(name, opener) +
// Open(name, dsn)。内置注册 sqlite（SQLite WAL 单写者）与 memory（进程内切片）两个驱动名；
// 新驱动=部署侧 import 自注册（依赖方向不反转——本包不感知具体驱动内部，同 database/sql
// 驱动注册范式）。Open 未知驱动名=可见错误（缺口不静默吞）。
package runtimestore

import (
	"fmt"
	"sync"
)

// Opener 驱动开启器（dsn=部署侧注入的存储定位串：sqlite=文件路径（空=进程内存库）；
// memory=忽略）。
type Opener func(dsn string) (Store, error)

var (
	driversMu sync.Mutex
	drivers   = map[string]Opener{}
)

// Register 注册驱动（同名覆盖=部署侧可替换实现；init 期调用，幂等安全）。
func Register(name string, opener Opener) {
	driversMu.Lock()
	defer driversMu.Unlock()
	drivers[name] = opener
}

// Open 按驱动名构建 Store（未知驱动=可见错误；驱动内失败=驱动侧错误透传不包装语义）。
func Open(name, dsn string) (Store, error) {
	driversMu.Lock()
	opener, ok := drivers[name]
	driversMu.Unlock()
	if !ok {
		return nil, fmt.Errorf("runtimestore: 未知驱动 %q（内置=sqlite|memory；部署侧驱动须先 Register）", name)
	}
	return opener(dsn)
}

// Drivers 已注册驱动名清单（诊断面用；名字序=注册表 map 遍历序排序后输出，机械可复现）。
func Drivers() []string {
	driversMu.Lock()
	defer driversMu.Unlock()
	out := make([]string, 0, len(drivers))
	for name := range drivers {
		out = append(out, name)
	}
	// 排序保证诊断输出确定性
	for i := 1; i < len(out); i++ {
		for j := i; j > 0 && out[j] < out[j-1]; j-- {
			out[j], out[j-1] = out[j-1], out[j]
		}
	}
	return out
}

func init() {
	// 内置驱动注册（sqlite.go / memory.go 各自 init 同款范式——此处收口亦可，单点声明
	// 便于「内置=sqlite|memory」的词形对账）。
	Register("sqlite", openSQLite)
	Register("memory", openMemory)
}
