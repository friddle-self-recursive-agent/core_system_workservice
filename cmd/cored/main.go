// cored Core 常驻进程：Apply / Event 分发 / Loop 引擎 / Scheme / Policy。
// 设计依据：core_system/systems/builtin/self-recursive-core + instance 仓 impl/self-recursive-core.md。
package main

import (
	"fmt"
	"os"
)

func main() {
	fmt.Println("cored: core_system control plane (skeleton)")
	// TODO(skeleton): 载入 scheme → 打开 repository(git) + opslog(sqlite)
	//                 → 起 eventbus / loopengine / apply engine / plugin host(dsh)
	os.Exit(2)
}
