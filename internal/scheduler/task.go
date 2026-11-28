// task.go —— Task 对象的宽松解析与 status 写面（调度器侧单一真相）。
//
// 设计依据：core_design/core/scheme.yaml kind: Task（spec 人写/上游写，status 只由调度器/
// 执行体写——P2 主体纪律）与 concepts/task-scheduler.md §3.1。解析=读面宽松（字段缺席/形变
// =零值不报错，与 proposal.Parse 同口径——写路径校验是形态权威）；写面=仓库系统写
// （fieldManager=task-scheduler，repository 系统写非 apply 管线——runtime/task-scheduler.md §2）。
// status 改写走 yaml 节点树（保注释/键序，repository.setResourceVersion 同款技法），
// 等值=零改动（幂等：同一原因反复写面自动免写，防热循环刷库）。
package scheduler

import (
	"bytes"
	"fmt"
	"strconv"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// phase 闭合集（scheme.yaml Task.status.phase 声明同源——runtime/task-scheduler.md §2）。
const (
	PhaseQueued    = "queued"
	PhaseReady     = "ready"
	PhaseBlocked   = "blocked"
	PhaseRunning   = "running"
	PhaseVerifying = "verifying"
	PhaseDone      = "done"
	PhaseFailed    = "failed"
	PhaseCanceled  = "canceled"
)

// migrations phase 迁移合法性表（闭合集；terminal 态零出边——终态制）。
// 调度器只写其中子集（queued/ready/blocked/running/failed 侧）；verifying→done/failed
// 由执行体/验证面回写——同一张表判别，主体不同词汇不分家。
var migrations = map[string]map[string]bool{
	PhaseQueued:    {PhaseReady: true, PhaseBlocked: true, PhaseCanceled: true},
	PhaseBlocked:   {PhaseReady: true, PhaseCanceled: true},
	PhaseReady:     {PhaseRunning: true, PhaseBlocked: true, PhaseCanceled: true},
	PhaseRunning:   {PhaseVerifying: true, PhaseReady: true, PhaseFailed: true, PhaseDone: true, PhaseCanceled: true},
	PhaseVerifying: {PhaseDone: true, PhaseFailed: true, PhaseReady: true, PhaseCanceled: true},
	PhaseDone:      {},
	PhaseFailed:    {},
	PhaseCanceled:  {},
}

// CanMigrate phase 迁移合法性（闭合集外=拒绝；未知 from 一律拒绝——野词不冒充合法）。
func CanMigrate(from, to string) bool {
	if from == to {
		return false // 同态=零迁移（写面幂等口径）
	}
	return migrations[from][to]
}

// Touches 写集/独占资源（scheme.yaml Task.spec.touches 声明同源；冲突判定输入）。
type Touches struct {
	Repos     []string
	Paths     []string
	Objects   []string
	Exclusive []string
}

// Workspace 工作区隔离声明（worktree=代码任务独立 worktree/branch；inplace=写集互斥前提下原地）。
type Workspace struct {
	Mode   string
	Base   string
	Branch string
}

// Spec Task.spec 宽松投影（读面不放大声明形变）。
type Spec struct {
	Title          string
	Summary        string
	Source         string
	SourceRef      string
	ScopeRef       string // 影响域归一（perService 上限键）
	DependsOn      []string
	Touches        Touches
	Priority       string
	Agent          string
	OwnerRefs      []string // 归属对象引用（metadata.ownerRefs/spec.ownerRefs——对象页相关任务反查）
	Workspace      Workspace
	PermissionMode string        // 调用 dsh agent 的权限档声明（2026-11-27 用户裁定：显式配置非隐式默认全权；空=宿主缺省 danger-full-access）
	MaxConcurrent  int           // limits.maxConcurrent 单任务覆写（只收紧不放宽）
	Timeout        time.Duration // limits.timeout 单任务覆写（只收紧租约）
}

// Status Task.status 现势投影。
type Status struct {
	Phase        string
	Attempts     int
	AgentSession string
	DependsOn    []string
	LastAction   string
	LastError    string
}

// taskDecl 解析宿主（宽松；refs 容忍字符串/映射两种词形——scheme 声明 ref 但既有文档两栖）。
type taskDecl struct {
	Metadata struct {
		Name      string `yaml:"name"`
		Scope     string `yaml:"scope"`
		OwnerRefs []any  `yaml:"ownerRefs"`
	} `yaml:"metadata"`
	Spec struct {
		Title     string `yaml:"title"`
		Summary   string `yaml:"summary"`
		Source    string `yaml:"source"`
		SourceRef any    `yaml:"sourceRef"`
		ScopeRef  any    `yaml:"scopeRef"`
		OwnerRefs []any  `yaml:"ownerRefs"`
		DependsOn []any  `yaml:"dependsOn"`
		Touches   struct {
			Repos     []any    `yaml:"repos"`
			Paths     []string `yaml:"paths"`
			Objects   []any    `yaml:"objects"`
			Exclusive []string `yaml:"exclusive"`
		} `yaml:"touches"`
		Priority  string `yaml:"priority"`
		Agent     any    `yaml:"agent"`
		Workspace struct {
			Mode   string `yaml:"mode"`
			Base   string `yaml:"base"`
			Branch string `yaml:"branch"`
		} `yaml:"workspace"`
		PermissionMode string `yaml:"permissionMode"`
		Limits         struct {
			MaxConcurrent int    `yaml:"maxConcurrent"`
			Timeout       string `yaml:"timeout"`
		} `yaml:"limits"`
	} `yaml:"spec"`
	Status struct {
		Phase        string `yaml:"phase"`
		Attempts     int    `yaml:"attempts"`
		AgentSession string `yaml:"agentSession"`
		DependsOn    []any  `yaml:"dependsOn"`
		LastAction   string `yaml:"lastAction"`
		LastError    string `yaml:"lastError"`
	} `yaml:"status"`
}

// Parse Task 对象原文 → spec/status 投影（无 status.phase=非 Task 形 → ok=false，调用方跳过）。
func Parse(content []byte) (Spec, Status, bool) {
	var d taskDecl
	if err := yaml.Unmarshal(content, &d); err != nil {
		return Spec{}, Status{}, false
	}
	// 新建 Task 的 status 为空=合法初始态（P2：status 只由调度器写，用户不写 status）——
	// 归一为 queued 参选；空相位不跳过（跳过=新任务永不被调度的死锁，2026-11-26 活体实证）。
	// 非 Task 内容已由 scheme 准入挡住，此处宽松读面不承担 kind 判别。
	phase := strings.TrimSpace(d.Status.Phase)
	if phase == "" {
		phase = PhaseQueued
	}
	spec := Spec{
		Title: d.Spec.Title, Summary: d.Spec.Summary, Source: d.Spec.Source,
		SourceRef: refID(d.Spec.SourceRef), ScopeRef: refID(d.Spec.ScopeRef),
		DependsOn: refIDs(d.Spec.DependsOn), Priority: strings.TrimSpace(d.Spec.Priority),
		Agent:          refID(d.Spec.Agent),
		OwnerRefs:      nonEmpty(append(refIDs(d.Metadata.OwnerRefs), refIDs(d.Spec.OwnerRefs)...)),
		Workspace:      Workspace{Mode: d.Spec.Workspace.Mode, Base: d.Spec.Workspace.Base, Branch: d.Spec.Workspace.Branch},
		PermissionMode: d.Spec.PermissionMode,
		MaxConcurrent:  d.Spec.Limits.MaxConcurrent,
	}
	for _, p := range d.Spec.Touches.Paths {
		if p = strings.TrimSpace(p); p != "" {
			spec.Touches.Paths = append(spec.Touches.Paths, p)
		}
	}
	spec.Touches.Repos = nonEmpty(refIDs(d.Spec.Touches.Repos))
	spec.Touches.Objects = nonEmpty(refIDs(d.Spec.Touches.Objects))
	spec.Touches.Exclusive = nonEmpty(d.Spec.Touches.Exclusive)
	if v, err := time.ParseDuration(strings.TrimSpace(d.Spec.Limits.Timeout)); err == nil {
		spec.Timeout = v
	}
	st := Status{
		Phase: phase, Attempts: d.Status.Attempts, AgentSession: d.Status.AgentSession,
		LastAction: d.Status.LastAction, LastError: d.Status.LastError,
		DependsOn: refIDs(d.Status.DependsOn),
	}
	return spec, st, true
}

// refID 归一对象引用为 "<scope>/<name>"（字符串容忍 task/<scope>/<name> 三段 kind 前缀；
// 映射取 scope/name 两键，scope 缺席=global——与 Task.dependsOn 跨 scope 语义一致）。
func refID(v any) string {
	switch t := v.(type) {
	case string:
		parts := strings.Split(strings.TrimSpace(t), "/")
		if len(parts) == 3 && parts[0] == TaskKind {
			parts = parts[1:]
		}
		return strings.Join(parts, "/")
	case map[string]any:
		name, _ := t["name"].(string)
		if name == "" {
			return ""
		}
		scope, _ := t["scope"].(string)
		if scope == "" {
			scope = "global"
		}
		return scope + "/" + name
	}
	return ""
}

// refIDs 批量归一引用（空值剔除，保序）。
func refIDs(vs []any) []string {
	var out []string
	for _, v := range vs {
		if id := refID(v); id != "" {
			out = append(out, id)
		}
	}
	return out
}

func nonEmpty(vs []string) []string {
	var out []string
	for _, v := range vs {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// statusPatch 一次 status 写面的字段集（nil=字段不动；空串=清空——lastError 语义）。
type statusPatch struct {
	Phase        string
	Attempts     *int
	AgentSession *string
	DependsOn    []string
	LastAction   *string
	LastError    *string
}

// WriteStatus Task.status 机械改写（yaml 节点树保注释/键序；status 区缺席=追加）。
// 返回 changed=false=零改动（值已等——幂等免写，调用方跳过仓库写面）。
func WriteStatus(content []byte, p statusPatch) (out []byte, changed bool, err error) {
	// md 形态（--- 围栏）=Task 对象常态（layout.md §3：身份字段住 frontmatter 顶层）——
	// 只改 frontmatter 内的 status，围栏与正文逐字节保留；丢失围栏=仓库按裸 yaml 解析要求
	// metadata 区而拒绝（2026-11-26 活体实证：重编码整文导致写面恒失败）。
	// 裸 yaml 形态（身份住 metadata 区）走整文重编码（status 在根级、与 metadata 同级）。
	front, body, isMD := splitFrontmatter(string(content))
	var root yaml.Node
	if err := yaml.Unmarshal([]byte(front), &root); err != nil {
		return nil, false, fmt.Errorf("scheduler: Task 原文解析: %w", err)
	}
	if len(root.Content) == 0 || root.Content[0].Kind != yaml.MappingNode {
		return nil, false, fmt.Errorf("scheduler: Task 文档根须为映射")
	}
	doc := root.Content[0]
	st := mapValue(doc, "status")
	if st == nil {
		st = &yaml.Node{Kind: yaml.MappingNode, Tag: "!!map"}
		doc.Content = append(doc.Content,
			&yaml.Node{Kind: yaml.ScalarNode, Value: "status", Tag: "!!str"}, st)
	}
	if st.Kind != yaml.MappingNode {
		return nil, false, fmt.Errorf("scheduler: Task.status 区须为映射")
	}
	if changed = applyPatch(st, p); !changed {
		return nil, false, nil
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return nil, false, fmt.Errorf("scheduler: Task 原文重编码: %w", err)
	}
	if err := enc.Close(); err != nil {
		return nil, false, err
	}
	if isMD {
		return []byte("---\n" + buf.String() + "---\n" + body), true, nil
	}
	return buf.Bytes(), true, nil
}

// splitFrontmatter 拆 md frontmatter 与正文（口径与 repository.splitDoc 同源：首行 ---、
// 第二个独立的 --- 收尾；正文含收尾行之后的全部内容逐字节保留）。非 md=整文作 frontmatter、
// 正文空。
func splitFrontmatter(text string) (front, body string, isMD bool) {
	if !strings.HasPrefix(text, "---") {
		return text, "", false
	}
	lines := strings.Split(text, "\n")
	if strings.TrimSpace(lines[0]) != "---" {
		return text, "", false
	}
	for i := 1; i < len(lines); i++ {
		if strings.TrimSpace(lines[i]) == "---" {
			return strings.Join(lines[1:i], "\n"), strings.Join(lines[i+1:], "\n"), true
		}
	}
	return text, "", false // 未闭合围栏：按整文处理（写面会由仓库侧 ErrContent 拦）
}

// applyPatch 就地落字段（等值=零改动聚合进 changed）。
func applyPatch(st *yaml.Node, p statusPatch) bool {
	changed := false
	if p.Phase != "" {
		changed = setScalar(st, "phase", p.Phase, "!!str") || changed
	}
	if p.Attempts != nil {
		changed = setScalar(st, "attempts", strconv.Itoa(*p.Attempts), "!!int") || changed
	}
	if p.AgentSession != nil {
		changed = setScalar(st, "agentSession", *p.AgentSession, "!!str") || changed
	}
	if p.DependsOn != nil {
		changed = setSequence(st, "dependsOn", p.DependsOn) || changed
	}
	if p.LastAction != nil {
		changed = setScalar(st, "lastAction", *p.LastAction, "!!str") || changed
	}
	if p.LastError != nil {
		changed = setScalar(st, "lastError", *p.LastError, "!!str") || changed
	}
	return changed
}

// setScalar 覆写/追加映射键的标量值（等值=零改动；改写保键序与注释——节点树就地改）。
func setScalar(m *yaml.Node, key, val, tag string) bool {
	if v := mapValue(m, key); v != nil {
		if v.Value == val && v.Tag == tag {
			return false
		}
		v.Value, v.Tag, v.Style = val, tag, 0
		return true
	}
	m.Content = append(m.Content,
		&yaml.Node{Kind: yaml.ScalarNode, Value: key, Tag: "!!str"},
		&yaml.Node{Kind: yaml.ScalarNode, Value: val, Tag: tag})
	return true
}

// setSequence 覆写/追加映射键的序列值（逐项等值比较——依赖投影稳定重写免空翻）。
func setSequence(m *yaml.Node, key string, vals []string) bool {
	if v := mapValue(m, key); v != nil && v.Kind == yaml.SequenceNode {
		if sequenceEqual(v, vals) {
			return false
		}
		v.Content = sequenceNodes(vals)
		return true
	}
	seq := &yaml.Node{Kind: yaml.SequenceNode, Tag: "!!seq", Content: sequenceNodes(vals)}
	m.Content = append(m.Content, &yaml.Node{Kind: yaml.ScalarNode, Value: key, Tag: "!!str"}, seq)
	return true
}

func sequenceEqual(seq *yaml.Node, vals []string) bool {
	if len(seq.Content) != len(vals) {
		return false
	}
	for i, n := range seq.Content {
		if n.Value != vals[i] {
			return false
		}
	}
	return true
}

func sequenceNodes(vals []string) []*yaml.Node {
	out := make([]*yaml.Node, 0, len(vals))
	for _, v := range vals {
		out = append(out, &yaml.Node{Kind: yaml.ScalarNode, Value: v, Tag: "!!str"})
	}
	return out
}

// mapValue 在映射节点里取键的值节点（yaml.v3 映射：Content 交替存放 key/value；
// 与 repository.mapValue 同构——包间不互依，机械复制不引依赖）。
func mapValue(mapping *yaml.Node, key string) *yaml.Node {
	if mapping == nil || mapping.Kind != yaml.MappingNode {
		return nil
	}
	for i := 0; i+1 < len(mapping.Content); i += 2 {
		if mapping.Content[i].Value == key {
			return mapping.Content[i+1]
		}
	}
	return nil
}
