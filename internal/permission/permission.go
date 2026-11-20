// Package permission 权限系统（v0.9，接替已删除的 Policy，A44/A45）。
// 设计依据：core_system/concepts/permission.md。
// 骨架：权限只绑 role；user→role、user→group→role；默认 global/admin 全权限；
// agent 判定方灵活（agent/role/parent-agent）；跨系统由被请求方（B 侧）判定。
package permission

import "fmt"

// Permission 一条权限（scope 域 + 动作）。
type Permission struct {
	Scope   string   // "global"、"systems/<name>/*"（system 开头各自权限域）
	Actions []string // read/write/approve/call...
}

// Role 角色（权限只出现在这里）。
type Role struct {
	Name        string
	Permissions []Permission
}

// Group 组（只绑 role，不绑权限）。
type Group struct {
	Name  string
	Roles []string
}

// Subject 判定主体。
type Subject struct {
	User   string
	Groups []string
	Roles  []string
}

// JudgedBy 判定方（灵活不限死，A45）。
type JudgedBy int

const (
	ByAgent  JudgedBy = iota // agent 自治（auto 级白名单内）
	ByRole                   // role 权限表查表
	ByParent                 // 上级 agent
)

// System 权限系统。
type System struct {
	Roles  map[string]*Role
	Groups map[string]*Group
}

// DefaultAdmin 内置全权限角色（不可删，A45）。
const DefaultAdmin = "global/admin"

// Check 本系统内动作判定。判定方由调用方声明（灵活）；跨系统请求走 RequestCross。
func (s *System) Check(sub Subject, scope, action string, by JudgedBy) error {
	roles := append(append([]string{}, sub.Roles...), s.groupsRoles(sub)...)
	for _, r := range roles {
		if r == DefaultAdmin {
			return nil // 默认全权限
		}
		if role, ok := s.Roles[r]; ok && role.Holds(scope, action) {
			return nil
		}
	}
	return fmt.Errorf("permission: denied %s on %s (judged by %d)", action, scope, by)
}

// RequestCross 跨系统：A 的 agent 请求 B 的 agent——由 B 侧判定（A 无自判权，A45）。
// B 侧判定留痕（SessionLog/审计）。
func (s *System) RequestCross(bSubject Subject, bScope, action string, by JudgedBy) error {
	return s.Check(bSubject, bScope, action, by)
}

func (s *System) groupsRoles(sub Subject) []string {
	var out []string
	for _, g := range sub.Groups {
		if grp, ok := s.Groups[g]; ok {
			out = append(out, grp.Roles...)
		}
	}
	return out
}

// Holds role 是否持有某 scope 动作。
func (r *Role) Holds(scope, action string) bool {
	for _, p := range r.Permissions {
		if p.Scope != scope {
			continue
		}
		for _, a := range p.Actions {
			if a == action || a == "*" {
				return true
			}
		}
	}
	return false
}

// HardRules 内置硬规则（原 Policy 硬规则迁入，代码级不可配置放宽，A44）。
// deny delete builtinRoot / deny disable scheme / deny remove DefaultAdmin.
func HardRules(obj, action string) error {
	if obj == "agent/builtinRoot" && (action == "delete" || action == "disable") {
		return fmt.Errorf("permission: hard rule denies %s on %s", action, obj)
	}
	if obj == "scheme/default" && action == "disable" {
		return fmt.Errorf("permission: hard rule denies %s on %s", action, obj)
	}
	if obj == DefaultAdmin && action == "delete" {
		return fmt.Errorf("permission: hard rule denies deleting %s", obj)
	}
	return nil
}
