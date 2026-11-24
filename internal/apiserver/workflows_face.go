// workflows_face.go —— workflow 运行管理聚合查询面（v0.37 console 读面第二段，E3 增量——
// cored-bridge.md §2 `GET /v1/workflows`）。console-ui.md 屏幕 5 的数据面：workflow 列表
// 聚合投影（运行次数 总/24h、当前相位、涉及对象、生效 enabled），三源同投：
//
//	① 声明源=builtin 注册表 + 文档库 Workflow 对象（跨 scope 枚举——builtin 本体与
//	   override 同名并存；override 语义见 concepts/workflow.md §1：builtin 本体不动，
//	   系统层同名覆盖 origin 标注 override-of-builtin）；
//	② 现势源=run 投影环（每锚最新 run 的 phase/status——「当前相位」；无 run=null 诚实空态）；
//	③ 计数源=event_log 归档分组 × dispatch 决议（每 (type,involved) 组解析一次，
//	   workflow.RunAnchor 与 run 投影同一锚解析函数——聚合无双源）。计数口径=归档行经
//	   派发决议路由到该锚的触发量（含冷却窗内被拦阻的触发——braked 不落 run 环，诚实口径
//	   入桥文档；「run 出现≠执行完毕」投影语义不变；扫描有界上限=opslog.MaxGroupScanRows，
//	   capped 如实标注）。
//
// 旧构建（路由未注册）经兜底 handler 405=代际判别器。
package apiserver

import (
	"context"
	"sort"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/flowbind"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
	"gopkg.in/yaml.v3"
)

// wfObjDecl Workflow 文档对象的声明投影面（scheme.yaml Workflow spec 的只读最小集——
// 形态合法性由写路径 scheme 校验闸门承担，此处只做声明→投影的映射，不重复校验）。
// phases 用 yaml.Node 宽形解析（声明面允许富相位行 {name,title,op,…}——v0.8 吸收 Loop 时
// Workflow.phase 为对象声明；字符串行与富行都投影为展示名）。
type wfObjDecl struct {
	Metadata struct {
		Name   string            `yaml:"name"`
		Scope  string            `yaml:"scope"`
		Labels map[string]string `yaml:"labels"`
	} `yaml:"metadata"`
	Spec struct {
		Enabled   *bool  `yaml:"enabled"`
		Connector string `yaml:"connector"`
		Services  struct {
			Involved []string `yaml:"involved"`
			Mode     string   `yaml:"mode"`
		} `yaml:"services"`
		Dataflow struct {
			Inputs []struct {
				Events []string `yaml:"events"`
				Docs   string   `yaml:"docs"`
			} `yaml:"inputs"`
			Outputs []struct {
				Status    string   `yaml:"status"`
				Notify    []string `yaml:"notify"`
				Artifacts string   `yaml:"artifacts"`
				Docs      string   `yaml:"docs"`
				Reports   []string `yaml:"reports"`
			} `yaml:"outputs"`
		} `yaml:"dataflow"`
		Phases []yaml.Node `yaml:"phases"`
	} `yaml:"spec"`
}

// phaseDisplayName 富相位行/字符串行的展示名投影（富行=name 缺省 title 兜底；字符串行=原样）。
func phaseDisplayName(n yaml.Node) (string, bool) {
	if n.Kind == yaml.ScalarNode {
		return n.Value, n.Value != ""
	}
	if n.Kind == yaml.MappingNode {
		name, title := "", ""
		for i := 0; i+1 < len(n.Content); i += 2 {
			k, v := n.Content[i].Value, n.Content[i+1]
			switch k {
			case "name":
				if v.Kind == yaml.ScalarNode {
					name = v.Value
				}
			case "title":
				if v.Kind == yaml.ScalarNode {
					title = v.Value
				}
			}
		}
		if name != "" {
			return name, true
		}
		if title != "" {
			return title, true
		}
	}
	return "", false
}

// wfObjectRow 单个 Workflow 对象行投影（本体/override 同构——文档与资源同规则，
// console-ui.md 屏幕 2 同构四件套的 kind 徽章之外，行键一致）。
type wfObjectRow struct {
	Scope           string    `json:"scope"`
	ResourceVersion int64     `json:"resourceVersion"`
	Origin          string    `json:"origin"`  // labels[core/origin]；无标注=global 本体 builtin / 系统层 custom
	Enabled         *bool     `json:"enabled"` // spec.enabled（未声明=null——不冒充）
	Involved        []string  `json:"involved"`
	Mode            string    `json:"mode"`
	DataflowInputs  []wfDfIO  `json:"dataflowInputs"`
	DataflowOutputs []wfDfOut `json:"dataflowOutputs"`
	Phases          []string  `json:"phases"`

	raw []byte // 声明原文（execution 绑定解析用——不投影不出口）
}

// wfDfIO / wfDfOut 数据流声明投影（screen 5 数据流 tab 拼图的节点数据源，节点全可点）。
type wfDfIO struct {
	Events []string `json:"events,omitempty"`
	Docs   string   `json:"docs,omitempty"`
}
type wfDfOut struct {
	Status    string   `json:"status,omitempty"`
	Notify    []string `json:"notify,omitempty"`
	Artifacts string   `json:"artifacts,omitempty"`
	Docs      string   `json:"docs,omitempty"`
	Reports   []string `json:"reports,omitempty"`
}

// wfRow 聚合面单行：一个 workflow 锚的声明∪现势∪计数投影。
type wfRow struct {
	Name       string        `json:"name"`
	SpecSource string        `json:"specSource"` // builtin=注册表命中；object=平台对象引用（run 锚解析同词汇）
	Phases     []string      `json:"phases"`     // 生效相位序列（override 声明优先；未声明 builtin 命中=注册表序列）
	Objects    []wfObjectRow `json:"objects"`    // 本体+override（global 在前，其余 scope 字典序）
	Effective  *wfEffective  `json:"effective"`  // 生效 override（无 override=null——builtin 本体生效）
	Latest     *wfLatest     `json:"latest"`     // 现势 run（环内最新；无 run=null）
	RunsTotal  int64         `json:"runsTotal"`
	Runs24h    int64         `json:"runs24h"`
	Execution  *wfExecution  `json:"execution"` // 执行模型绑定面（v0.49 R-7——workflow-runtime.md §11.1；cored-bridge.md §2 execution 键=代际判别器）
}

// wfExecution 决策面 execution 投影（connector.md 规则 6：未绑定=缺口可见不静默）。
// flowbind.Execution 的聚合面别名（包间零依赖——本面只消费形状，解析逻辑住 flowbind）。
type wfExecution = flowbind.Execution

type wfEffective struct {
	Scope   string `json:"scope"`
	Enabled *bool  `json:"enabled"`
}
type wfLatest struct {
	Run       int64  `json:"run"`
	Phase     string `json:"phase"`
	Status    string `json:"status"`
	StartedAt string `json:"startedAt"`
	EventID   string `json:"eventId"`
}

// originOf 对象 origin 投影：labels.core/origin 优先（override 惯例），无标注按 scope 推导
// （global=builtin 本体文档；系统层=custom——推导值仅展示用，真实标注在 labels）。
func originOf(labels map[string]string, scope string) string {
	if o := labels["core/origin"]; o != "" {
		return o
	}
	if scope == "global" {
		return "builtin"
	}
	return "custom"
}

// parseWfObject 声明投影（宽松解析：字段类型不合=该字段缺席、其余字段照投影——读面不放大
// 声明形变为整面失败；写路径 scheme 校验是形态权威）。
func parseWfObject(raw []byte, rv int64, scope string) wfObjectRow {
	var d wfObjDecl
	row := wfObjectRow{Scope: scope, ResourceVersion: rv, Involved: []string{},
		DataflowInputs: []wfDfIO{}, DataflowOutputs: []wfDfOut{}, Phases: []string{}, raw: raw}
	_ = yaml.Unmarshal(raw, &d) // 错误=局部字段缺席（宽松投影），不致命
	if d.Metadata.Scope != "" {
		row.Scope = d.Metadata.Scope // 正文声明的 scope 与目录前缀一致（存储写路径已校验）
	}
	row.Origin = originOf(d.Metadata.Labels, row.Scope)
	row.Enabled = d.Spec.Enabled
	row.Involved = append(row.Involved, d.Spec.Services.Involved...)
	row.Mode = d.Spec.Services.Mode
	for i := range d.Spec.Phases {
		if name, ok := phaseDisplayName(d.Spec.Phases[i]); ok {
			row.Phases = append(row.Phases, name)
		}
	}
	for _, in := range d.Spec.Dataflow.Inputs {
		row.DataflowInputs = append(row.DataflowInputs, wfDfIO{Events: in.Events, Docs: in.Docs})
	}
	for _, out := range d.Spec.Dataflow.Outputs {
		row.DataflowOutputs = append(row.DataflowOutputs, wfDfOut{
			Status: out.Status, Notify: out.Notify, Artifacts: out.Artifacts, Docs: out.Docs, Reports: out.Reports,
		})
	}
	return row
}

// anchorRunCounts 归档分组 × dispatch 决议 → 每锚触发计数（③ 计数源；since 零值=全量窗）。
// 返回 (每锚计数, 扫描行数, capped)。查询失败=错误上抛（计数面诚实：不出半数字）。
func anchorRunCounts(triggers TriggerFace, eventLog EventLogFace, since time.Time) (map[string]int64, int64, bool, error) {
	if triggers == nil || eventLog == nil {
		return map[string]int64{}, 0, false, nil // face 缺席=计数缺面（诚实缺数字，不 panic）
	}
	groups, scanned, capped, err := eventLog.GroupEvents(opslog.EventQuery{Since: since})
	if err != nil {
		return nil, scanned, capped, err
	}
	out := map[string]int64{}
	for _, g := range groups {
		d := triggers.Resolve(g.Type, g.Involved)
		if d.Exec.Executor != workflow.ExecWorkflow {
			continue
		}
		anchor, _ := workflow.RunAnchor(d.Route)
		if anchor == "" {
			continue // 空 anchor（防御性：非 workflow 家族决议原样可见，不冒充归属）
		}
		out[anchor] += int64(g.Count)
	}
	return out, scanned, capped, nil
}

// collectWorkflowRows 聚合三源投影（纯读取：objects/runs/triggers/eventLog 全 GET 面；
// 系统与对象都在控制面量级——现读现投零缓存，git 后端权威 apply 落库即生效）。
// 返回行集（名字典序）+计数扫描有界口径。
func collectWorkflowRows(ctx context.Context, objects ObjectSource, runs RunFace, triggers TriggerFace, eventLog EventLogFace) (rows []wfRow, scanned int64, capped bool) {
	// ① 声明源：文档库 Workflow 对象（跨 scope 枚举）按名分组；global 本体行在最前。
	objsByName := map[string][]wfObjectRow{}
	if objects != nil {
		if refs, err := objects.List(ctx, "Workflow", ""); err == nil {
			for _, o := range refs {
				objsByName[o.Name] = append(objsByName[o.Name], parseWfObject(o.Content, o.ResourceVersion, o.Scope))
			}
		}
	}
	for name := range objsByName {
		rows := objsByName[name]
		sort.SliceStable(rows, func(i, j int) bool {
			if (rows[i].Scope == "global") != (rows[j].Scope == "global") {
				return rows[i].Scope == "global"
			}
			return rows[i].Scope < rows[j].Scope
		})
		objsByName[name] = rows
	}
	// ② 现势源：run 环内每锚最新 run（Runs 旧→新，后写覆盖）。
	latest := map[string]workflow.RunRecord{}
	if runs != nil {
		for _, r := range runs.State().Runs {
			if r.Workflow == "" {
				continue
			}
			latest[r.Workflow] = r
		}
	}
	// ③ 计数源：全量窗 + 24h 窗两次分组（同一条归档流水，窗口差=24h 量）。
	total, scTotal, capTotal, errTotal := anchorRunCounts(triggers, eventLog, time.Time{})
	recent, sc24, cap24, err24 := anchorRunCounts(triggers, eventLog, time.Now().UTC().Add(-24*time.Hour))
	if errTotal != nil || err24 != nil {
		// 计数面失败不吞行：行集照出（①②源），计数投影=空（不冒充零/半数字——诚实缺面）。
		total, recent = map[string]int64{}, map[string]int64{}
	}
	scanned, capped = scTotal+sc24, capTotal || cap24
	// ④ 执行模型绑定面（v0.49 R-7）：ConnectorOps 跨 scope 枚举一次 → 按名索引 →
	//    逐锚 execution 投影（对象缺席=unbound 缺口可见——connector.md 规则 6）。
	connIdx := flowbind.IndexRows(nil)
	if objects != nil {
		if cobjs, err := objects.List(ctx, "ConnectorOps", ""); err == nil {
			connIdx = flowbind.IndexRows(cobjs)
		} // 枚举失败=execution 缺面（行集照出，不冒充 unbound）
	}
	// 行集=声明对象名 ∪ builtin 注册表 ∪ 现势 run 锚（三源并集——缺任何一源都不吞行）。
	names := map[string]bool{}
	for _, n := range workflow.BuiltinNames() {
		names[n] = true
	}
	for n := range objsByName {
		names[n] = true
	}
	for n := range latest {
		names[n] = true
	}
	rows = make([]wfRow, 0, len(names))
	for name := range names {
		objs := objsByName[name]
		row := wfRow{
			Name:       name,
			SpecSource: specSourceOf(name, objs),
			Phases:     effectivePhases(name, objs),
			Objects:    objs,
			RunsTotal:  total[name],
			Runs24h:    recent[name],
		}
		if eff := effectiveObject(name, objs); eff != nil {
			row.Effective = &wfEffective{Scope: eff.Scope, Enabled: eff.Enabled}
		}
		// execution 投影（v0.49 R-7）：生效对象声明解析（缺= builtin 锚同名兜底——
		// 无文档 builtin 的绑定=same-name ConnectorOps，缺席照实 unbound）。
		if eff := effectiveObject(name, objs); eff != nil {
			row.Execution = flowbind.ExecutionOf(eff.Scope, name, eff.raw, connIdx)
		} else if len(objs) > 0 {
			row.Execution = flowbind.ExecutionOf(objs[0].Scope, name, objs[0].raw, connIdx)
		} else {
			row.Execution = flowbind.ExecutionOf("global", name, nil, connIdx)
		}
		if lr, ok := latest[name]; ok {
			row.Latest = &wfLatest{Run: lr.Run, Phase: lr.Phase, Status: lr.Status,
				StartedAt: lr.StartedAt.UTC().Format(time.RFC3339Nano), EventID: lr.EventID}
		}
		rows = append(rows, row)
	}
	sort.Slice(rows, func(i, j int) bool { return rows[i].Name < rows[j].Name })
	return rows, scanned, capped
}

// specSourceOf 行的声明来源投影（run 锚解析同词汇：builtin 注册表命中=builtin；其余=object）。
func specSourceOf(name string, objs []wfObjectRow) string {
	if workflow.IsBuiltin(name) {
		return workflow.SpecBuiltin
	}
	if len(objs) > 0 {
		return workflow.SpecObject
	}
	return "" // 防御性（行必来自三源之一；空=两源皆缺的未知名）
}

// effectiveObject 生效 override（系统层同名覆盖；多系统并存=字典序首个——确定性投影）。
// override 语义只在「存在可被覆盖的本体」时成立：builtin 注册表命中（builtin 本体）或
// global 文档对象在场；纯系统层自定义对象（无本体可覆盖）=定义本身，effective=null。
func effectiveObject(name string, objs []wfObjectRow) *wfObjectRow {
	if !workflow.IsBuiltin(name) {
		hasGlobal := false
		for _, o := range objs {
			if o.Scope == "global" {
				hasGlobal = true
				break
			}
		}
		if !hasGlobal {
			return nil
		}
	}
	var eff *wfObjectRow
	for i := range objs {
		if objs[i].Scope == "global" {
			continue
		}
		if eff == nil || objs[i].Scope < eff.Scope {
			eff = &objs[i]
		}
	}
	return eff
}

// effectivePhases 生效相位序列（override 声明优先 → builtin 注册表 → 空数组）。
func effectivePhases(name string, objs []wfObjectRow) []string {
	if eff := effectiveObject(name, objs); eff != nil && len(eff.Phases) > 0 {
		return eff.Phases
	}
	if workflow.IsBuiltin(name) {
		return workflow.BuiltinPhaseNames(name)
	}
	return []string{}
}

// workflowObjectsJSON 对象行 → 载荷 map（零值数组形状归一：[]=声明为空，null=缺席语义；
// enabled=null=spec.enabled 未声明——不冒充）。
func workflowObjectsJSON(objs []wfObjectRow) []map[string]any {
	out := make([]map[string]any, 0, len(objs))
	for _, o := range objs {
		out = append(out, map[string]any{
			"scope": o.Scope, "resourceVersion": o.ResourceVersion, "origin": o.Origin,
			"enabled": o.Enabled, "involved": nonNil(o.Involved), "mode": o.Mode,
			"dataflowInputs": nonNilIO(o.DataflowInputs), "dataflowOutputs": nonNilOut(o.DataflowOutputs),
			"phases": nonNil(o.Phases),
		})
	}
	return out
}

func nonNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

func nonNilIO(s []wfDfIO) []wfDfIO {
	if s == nil {
		return []wfDfIO{}
	}
	return s
}

func nonNilOut(s []wfDfOut) []wfDfOut {
	if s == nil {
		return []wfDfOut{}
	}
	return s
}
