// Package scheme 载入并校验类型注册表（core_system/core/scheme.yaml）。
// 设计依据：core_system/concepts/scheme.md。启动自检：野对象拒绝。
package scheme

import "fmt"

// Kind 注册表中的一个对象类型。
type Kind struct {
	Name    string `yaml:"kind"`
	Builtin bool   `yaml:"builtin"`
	Storage string `yaml:"storage"` // 存储路径模板（instance 仓 layout.md）
}

// EventType 事件类型注册项。
type EventType struct {
	Type   string `yaml:"type"`
	Source string `yaml:"source"`
	Route  string `yaml:"route"` // 缺省路由
}

// Scheme 类型注册表（含事件路由表"dispatch"）。
type Scheme struct {
	Version    string     `yaml:"version"`
	Kinds      []Kind     `yaml:"kinds"`
	DocTypes   []string   `yaml:"docTypes"`
	EventTypes []EventType `yaml:"eventTypes"`
}

// Lookup 查找对象类型；未注册返回错误（Scheme 闸门）。
func (s *Scheme) Lookup(kind string) (Kind, error) {
	for _, k := range s.Kinds {
		if k.Name == kind {
			return k, nil
		}
	}
	return Kind{}, fmt.Errorf("scheme: kind %q not registered", kind)
}

// Load 从 core/scheme.yaml 载入（TODO: yaml 解析与 dispatch 表）。
func Load(_ []byte) (*Scheme, error) {
	return nil, fmt.Errorf("scheme: not implemented (skeleton)")
}
