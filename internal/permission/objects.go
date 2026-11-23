// objects.go Role/Group 对象准入装载（concepts/permission.md §3：权限只绑 role，
// group 只绑 role；对象住 global/roles/<name>.yaml、global/groups/<name>.yaml——core/scheme.yaml Role/Group 行）。
// 装载即权威快照：文档面 Role/Group 对象是授权面的唯一来源——主体断言的 role 名必须能对上
// 这里装载的对象才产生权限（断言本身不授权）。装载时机：启动（失败=拒绝启动，宁拒不降级）
// + 事件驱动热装载（refresh.go：变更信号→重投影→整体替换，失败=保旧快照）。
package permission

import (
	"fmt"

	"gopkg.in/yaml.v3"
)

// ObjectDoc 文档面对象投影（storage 无关：permission 包不反向依赖存储层，调用方从任意读源投影）。
type ObjectDoc struct {
	Kind    string // Role | Group（其余 kind 拒绝——授权面宁拒不降级）
	Name    string // 对象名（文档 metadata.name 缺席时的回退键）
	Scope   string
	Content []byte // 文档原文（yaml）
}

// objectDoc yaml 解析目标（对齐 core/scheme.yaml Role/Group spec 声明）。
type objectDoc struct {
	Metadata struct {
		Name  string `yaml:"name"`
		Scope string `yaml:"scope"`
	} `yaml:"metadata"`
	Spec struct {
		On          []string     `yaml:"on"` // Role 对象挂载（A49）
		Permissions []Permission `yaml:"permissions"`
		Roles       []string     `yaml:"roles"` // Group 只绑 role
	} `yaml:"spec"`
}

// LoadObjects 全量装载 Role/Group 对象（替换语义=装载即快照）。任何坏对象整体拒绝（不半装载）：
// 授权面出现解析歧义比静默缺权限更安全（permission.md 风险表：判定灵活→以确定性换可审计）。
// 重名对象拒绝（同一 role 两份=授权歧义）。
func (s *System) LoadObjects(docs []ObjectDoc) error {
	roles := map[string]*Role{}
	groups := map[string]*Group{}
	for _, d := range docs {
		switch d.Kind {
		case "Role", "Group":
		default:
			return fmt.Errorf("permission: 不支持的权限对象 kind %q（Role|Group）", d.Kind)
		}
		var doc objectDoc
		if err := yaml.Unmarshal(d.Content, &doc); err != nil {
			return fmt.Errorf("permission: 解析 %s %s: %w", d.Kind, d.Name, err)
		}
		name := doc.Metadata.Name
		if name == "" {
			name = d.Name
		}
		if name == "" {
			return fmt.Errorf("permission: %s 对象缺 metadata.name", d.Kind)
		}
		switch d.Kind {
		case "Role":
			if _, dup := roles[name]; dup {
				return fmt.Errorf("permission: role %q 重复装载（授权歧义拒绝）", name)
			}
			for i, p := range doc.Spec.Permissions {
				if p.Scope == "" {
					return fmt.Errorf("permission: role %q 第 %d 条权限缺 scope（权限域按 scope，骨架 1）", name, i+1)
				}
				if len(p.Actions) == 0 {
					return fmt.Errorf("permission: role %q 第 %d 条权限缺 actions", name, i+1)
				}
			}
			roles[name] = &Role{Name: name, On: doc.Spec.On, Permissions: doc.Spec.Permissions}
		case "Group":
			if _, dup := groups[name]; dup {
				return fmt.Errorf("permission: group %q 重复装载", name)
			}
			bound := make([]string, 0, len(doc.Spec.Roles))
			for _, r := range doc.Spec.Roles {
				n := normalizeRef(r)
				if n == "" {
					return fmt.Errorf("permission: group %q 含空 role 绑定", name)
				}
				bound = append(bound, n)
			}
			groups[name] = &Group{Name: name, Roles: bound}
		}
	}
	// 替换语义收口：全批校验通过后整体上架（原子指针替换）——读者永远看到某一代完整快照；
	// 任一环节失败不触达此行 = 旧快照原封不动（热装载保旧语义的机制保证，refresh.go）。
	gen := uint64(1)
	if cur := s.snap.Load(); cur != nil {
		gen = cur.Gen + 1 // 权威代次：成功装载一次 +1（判定留痕可追，permission.md 规则 3）
	}
	s.snap.Store(&Snapshot{Roles: roles, Groups: groups, Gen: gen})
	return nil
}
