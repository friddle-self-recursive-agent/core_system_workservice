// Package dsh dsh 通用驱动：exec（命令白名单模板，默认人审）。
// 设计依据：core_system/concepts/connector.md 可行性（通用驱动太危险）。
// 契约归属（2026-11-27 用户裁定）：本包=builtin-connector 的 default-dsh-agent/v1 契约的 builtin 实现
// （systems/builtin/connector/interfaces/default-dsh-agent.yaml）——spawn-agent/exec 的
// permissionMode 是契约显式参数（缺省 danger-full-access、可覆写降档），不做隐式全权。
package dsh
