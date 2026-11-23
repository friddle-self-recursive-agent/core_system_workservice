// Package permission 权限系统（v0.9，接替已删除的 Policy，A44/A45）。
// 设计依据：core_system/concepts/permission.md。
// 骨架：权限只绑 role；user→role、user→group→role；默认 global/admin 全权限；
// agent 判定方灵活（agent/role/parent-agent）；跨系统由被请求方（B 侧）判定。
package permission

import (
	"fmt"
	"strings"
	"sync/atomic"
)

// Permission 一条权限（scope 域 + 动作）。
type Permission struct {
	Scope   string   // "global"、"systems/<name>/*"（system 开头各自权限域）
	Actions []string // read/write/approve/call...
}

// Role 角色（权限只出现在这里；A49：原 Bind 的对象挂载并入为 On）。
type Role struct {
	Name        string
	On          []string // 对象挂载（原 Bind.role/on，A49）
	Permissions []Permission
}

// HoldsOn 对象是否挂载了本 role。
func (r *Role) HoldsOn(obj string) bool {
	for _, o := range r.On {
		if o == obj {
			return true
		}
	}
	return false
}

// Group 组（只绑 role，不绑权限）。
type Group struct {
	Name  string
	Roles []string
}

// Subject 判定主体（v0.9 请求级主体：HTTP 写面按请求解析携带；进程级本地信任=引擎缺省 fallback）。
// User 是主体标识（user-system 目录侧未接前由调用方断言，审计留痕用）；
// Roles/Groups 是主体断言的持有关系——权限本身只由文档面 Role/Group 对象授予（objects.go 装载）。
type Subject struct {
	User   string
	Groups []string
	Roles  []string
}

// ID 主体标识摘要（审计列 subject 用）：user 优先；无 user 记断言的 roles/groups；全空=anonymous。
func (s Subject) ID() string {
	if s.User != "" {
		return s.User
	}
	parts := append(append([]string{}, s.Roles...), s.Groups...)
	if len(parts) == 0 {
		return "anonymous"
	}
	return strings.Join(parts, ",")
}

// normalizeRef 引用规整：文档面惯例带 kind 前缀（role/devtools-admin、group/devtools-team），
// 注册表键为裸名——判定前统一剥前缀（DefaultAdmin 全字面匹配，不受影响）。
func normalizeRef(ref string) string {
	if s, ok := strings.CutPrefix(ref, "role/"); ok {
		return s
	}
	if s, ok := strings.CutPrefix(ref, "group/"); ok {
		return s
	}
	return ref
}

// JudgedBy 判定方（灵活不限死，A45）。
type JudgedBy int

const (
	ByAgent  JudgedBy = iota // agent 自治（auto 级白名单内）
	ByRole                   // role 权限表查表
	ByParent                 // 上级 agent
)

// String 判定方名（审计留痕可读：谁判的——permission.md 规则 3）。
func (j JudgedBy) String() string {
	switch j {
	case ByAgent:
		return "agent"
	case ByRole:
		return "role"
	case ByParent:
		return "parent-agent"
	}
	return fmt.Sprintf("judged-by(%d)", int(j))
}

// Snapshot 权威快照（装载即快照，objects.go）：不可变整体替换——判定高频读/装载低频写，
// 读方无锁共享（热装载原子换指针，判定永远对某一代完整快照进行，不判半装载态）。
// Gen 权威代次：成功装载一次 +1（0=从未装载）——判定依据可追（permission.md 规则 3 留痕）。
type Snapshot struct {
	Roles  map[string]*Role
	Groups map[string]*Group
	Gen    uint64
}

// System 权限系统。零值=空授权面（nil 快照）：除内置 admin 字面量外全拒（fail-closed）。
type System struct {
	snap atomic.Pointer[Snapshot]
}

// Snapshot 当前权威快照；未装载=nil（调用方按空授权面处理）。
func (s *System) Snapshot() *Snapshot { return s.snap.Load() }

// DefaultAdmin 内置全权限角色（不可删，A45）。
const DefaultAdmin = "global/admin"

// Check 本系统内动作判定。判定方由调用方声明（灵活）；跨系统请求走 RequestCross。
// 快照读取点：原子取当前代——热装载（LoadObjects/热刷新）只整体替换，不锁判定路径。
func (s *System) Check(sub Subject, scope, action string, by JudgedBy) error {
	for _, r := range sub.Roles {
		if normalizeRef(r) == DefaultAdmin {
			return nil // 默认全权限（字面量，不依赖装载快照——本地信任 fallback 最低保障）
		}
	}
	if snap := s.snap.Load(); snap != nil {
		for _, r := range sub.Roles {
			if role, ok := snap.Roles[normalizeRef(r)]; ok && role.Holds(scope, action) {
				return nil
			}
		}
		for _, name := range groupsRoles(snap, sub) {
			if role, ok := snap.Roles[name]; ok && role.Holds(scope, action) {
				return nil
			}
		}
	}
	return fmt.Errorf("permission: denied %s on %s (subject=%s judged by %s)", action, scope, sub.ID(), by)
}

// RequestCross 跨系统：A 的 agent 请求 B 的 agent——由 B 侧判定（A 无自判权，A45）。
// B 侧判定留痕（SessionLog/审计）。
func (s *System) RequestCross(bSubject Subject, bScope, action string, by JudgedBy) error {
	return s.Check(bSubject, bScope, action, by)
}

func groupsRoles(snap *Snapshot, sub Subject) []string {
	var out []string
	for _, g := range sub.Groups {
		if grp, ok := snap.Groups[normalizeRef(g)]; ok {
			out = append(out, grp.Roles...)
		}
	}
	return out
}

// Holds role 是否持有某动作域的某动作。
// 动作域两种形（permission.md §1/§3 + scheme.yaml Role spec 域词汇 global|systems/<n>/*）：
//   - 对象域：apply 等对象写路径传对象 scope 值（"<n>" 系统层 / "<n>/<i>" 实例层 / global），
//     按 layout.md「路径=scope」升为路径形再匹配（"<n>"→systems/<n>）；
//   - 资源域：调用方原样传域串（如 ops/onedev-ci 的 ConnectorOps call 域），精确比对。
//
// "X/*" 模式命中基域及其子域（含基域本身）。
func (r *Role) Holds(domain, action string) bool {
	for _, p := range r.Permissions {
		if !scopeMatch(p.Scope, domain) {
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

// scopeMatch 权限域模式 × 动作域：精确相等，或 "X/*" 通配（对象域经 domainPath 升路径后比对）。
func scopeMatch(pattern, domain string) bool {
	if pattern == domain {
		return true
	}
	base, ok := strings.CutSuffix(pattern, "/*")
	if !ok {
		return false
	}
	path := domainPath(domain)
	return path == base || strings.HasPrefix(path, base+"/")
}

// domainPath 动作域 → 路径形（layout.md「路径=scope」）：对象 scope 值升为目录路径；
// 已是路径形（systems/、core/ 前缀）、资源域串与保留值原样返回。
func domainPath(domain string) string {
	switch domain {
	case "", "global", "builtin":
		return domain
	}
	if strings.HasPrefix(domain, "systems/") || strings.HasPrefix(domain, "core/") {
		return domain
	}
	if i := strings.IndexByte(domain, '/'); i >= 0 {
		// 实例层 "<system>/<instance>" → systems/<system>/instances/<instance>
		return "systems/" + domain[:i] + "/instances/" + domain[i+1:]
	}
	// 系统层 "<system>" → systems/<system>
	return "systems/" + domain
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

// HumanReviewKind 人审档对象类（首段，permission.md §4 机读判定 + event.md 规则 9）：
// 授权面装载类（Role|Group——LoadObjects 收编的 kind 即授权面本体）。
// 锚=规则 5/P7「改 role/绑定=改权限=高敏变更（人审）」：对授权面自身的写被拒=试图变更
// 权限结构=人审档，写路径引擎据此自产 ApprovalNeeded。hard rules 拒绝≠人审档（永不可批，
// 不产事件）；dry-run 零副作用不产事件。对象类清单可扩（User/Provider/builtin scope 随
// 声明面扩展入表），首段只收 Role|Group 与装载面同源。
func HumanReviewKind(kind string) bool {
	return kind == "Role" || kind == "Group"
}
