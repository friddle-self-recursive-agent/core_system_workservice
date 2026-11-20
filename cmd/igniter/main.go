// igniter 点火器：装载 core/ → 启动 cored → 消失。
// 设计依据：instance 仓 loop/bootstrap.md（预算 <1000 行；无配置；冷启动幂等）。
package main

import (
	"fmt"
	"os"
)

func main() {
	coreDir := os.Getenv("CORE_SYSTEM_DOCS") // 默认 core_system 仓检出目录
	if coreDir == "" {
		coreDir = "../core_system"
	}
	fmt.Printf("igniter: 校验并装载 %s/core/ ...\n", coreDir)
	// TODO(skeleton): 1. 校验 core/scheme.yaml 2. 起 cored 3. 退出
	fmt.Println("igniter: not implemented (skeleton)")
	os.Exit(2)
}
