// validate_decl.go —— Workflow 对象声明语义准入（2026-11-28 用户令「有没有apply文档的时候。
// validate scheme了吗？没有就实现下」——apply.md 准入闸 ③b）。
//
// 为什么要有：scheme.ValidateDoc 结构准入（字段形状/枚举）过 ≠ 声明可用——**跨字段一致性**
// 无闸时坏声明照落库，run 受理即卡死（2026-11-28 实证：sybran-release-update rv1 phases 缺
// 起点 Idle、run#127 phase=Reporting/status=triggered 永久卡壳）。本闸在写路径拦：
//
//	① phases 声明必填（有 transitions/entryPoints/nodes 而无 phases=拒——无表可查无落点可判）；
//	② transitions from/to ∈ phases（引用未声明相位=拒——查表必空）；
//	③ 节点 phase ∈ phases；unit 词汇 ∈ connectops|agent|service|workflow|gate（session 已退役）；
//	   connectops→connector+op 必填、service→service 必填、workflow→fanout 声明必填；
//	④ entryPoints.node ∈ 节点名集（或 phase ∈ phases）且 events 非空；
//	⑤ triggers.events.types ∈ scheme.eventTypes 注册（未注册类型=触发永不命中——注册表闸门前置）；
//	⑥ approval.mode ∈ auto|agent|human（空=human 缺省合法）。
//
// 解析复用 docsSpecDecl/parsePhaseNode 同一单点（投影与校验同源零双解析漂移）。装配层经
// apply.WithSpecSemantics("Workflow", workflow.ValidateDecl) 注入（注册制——apply 不 import
// workflow 域语义，缺席=既有行为面零翻转）。校验为纯函数零 IO（sch 可 nil——注册校验跳过）。
package workflow

import (
	"fmt"
	"strings"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"gopkg.in/yaml.v3"
)

// declSemantics 语义校验声明形态（与 docsSpecDecl 同源字段子集——校验只读不投影）。
type declSemantics struct {
	Spec struct {
		Phases      []yaml.Node `yaml:"phases"`
		Transitions []struct {
			From string `yaml:"from"`
			On   string `yaml:"on"`
			To   string `yaml:"to"`
		} `yaml:"transitions"`
		EntryPoints []struct {
			Node    string   `yaml:"node"`
			Phase   string   `yaml:"phase"`
			Enabled *bool    `yaml:"enabled"`
			Events  []string `yaml:"events"`
		} `yaml:"entryPoints"`
		Approval struct {
			Mode string `yaml:"mode"`
		} `yaml:"approval"`
		Triggers struct {
			Events struct {
				Types []string `yaml:"types"`
			} `yaml:"events"`
		} `yaml:"triggers"`
	} `yaml:"spec"`
}

// 合法执行单元词汇（§1.3 三型+workflow 词形+gate 控制节点；session 已退役——§1.3 裁决）。
var validUnits = map[string]bool{"connectops": true, "agent": true, "service": true, "workflow": true, "gate": true}

// 合法审批策略词（空=human 缺省）。
var validApprovalModes = map[string]bool{"auto": true, "agent": true, "human": true, "": true}

// ValidateDecl Workflow 对象声明语义校验（纯函数；sch=nil 时跳过事件类型注册校验）。
// 返回 nil=过；非 nil=聚合错误（项数+逐项人读消息——写路径 400 直达作者）。
func ValidateDecl(raw []byte, sch *scheme.Scheme) error {
	var doc declSemantics
	if err := yaml.Unmarshal(raw, &doc); err != nil {
		return fmt.Errorf("workflow 声明不可解析: %w", err)
	}
	var errs []string
	add := func(format string, a ...any) {
		errs = append(errs, fmt.Sprintf(format, a...))
	}

	// ① 相位序列投影（标量=相位名；富节点=phase 锚）+节点集（parsePhaseNode 单点）。
	phaseSet := map[string]bool{}
	var nodeNames []string
	for i := range doc.Spec.Phases {
		if nv, ok := parsePhaseNode(doc.Spec.Phases[i]); ok {
			phaseSet[nv.Phase] = true
			if nv.Name != "" {
				nodeNames = append(nodeNames, nv.Name)
			}
		}
	}
	hasDecls := len(doc.Spec.Transitions) > 0 || len(doc.Spec.EntryPoints) > 0 || len(phaseSet) > 0
	if hasDecls && len(phaseSet) == 0 {
		add("phases 未声明（有 transitions/entryPoints 而无相位序列——受理无落点可判，run 必卡死）")
	}

	// ② 节点语义（unit 词汇×声明完备性×相位归属）。
	nodeSet := map[string]bool{}
	for _, n := range nodeNames {
		nodeSet[n] = true
	}
	for i := range doc.Spec.Phases {
		nv, ok := parsePhaseNode(doc.Spec.Phases[i])
		if !ok || nv.Unit == "" {
			continue
		}
		if nv.Name == "" {
			add("富节点缺 name（相位 %s——执行单元节点必须具名）", nv.Phase)
			continue
		}
		if !phaseSet[nv.Phase] {
			add("节点 %s 相位 %s 未在 phases 声明（落点不可达——run 永不执行该节点）", nv.Name, nv.Phase)
		}
		if !validUnits[nv.Unit] {
			add("节点 %s unit=%q 不在词汇 [connectops|agent|service|workflow|gate]（session 已退役）", nv.Name, nv.Unit)
			continue
		}
		switch nv.Unit {
		case "connectops":
			if nv.Connector == "" || nv.Op == "" {
				add("节点 %s unit=connectops 缺 connector/op 声明（执行面按声明查注册表——缺席即显式失败空转）", nv.Name)
			}
		case "service":
			if nv.Service == "" {
				add("节点 %s unit=service 缺 service 声明（拉起面按 service 锚解析）", nv.Name)
			}
		case "workflow":
			if nv.Fanout == nil {
				add("节点 %s unit=workflow 缺 fanout 声明（扇出计划键 from/key——缺即纯查表空转）", nv.Name)
			}
		}
	}

	// ③ 迁移表：from/to ∈ phases、on 非空。
	for i, t := range doc.Spec.Transitions {
		if t.From == "" || t.On == "" || t.To == "" {
			add("transitions[%d] 缺 from/on/to（%q→%q on %q——查表行三段必填）", i, t.From, t.To, t.On)
			continue
		}
		if len(phaseSet) > 0 {
			if !phaseSet[t.From] {
				add("transitions[%d] from=%q 未在 phases 声明（查表必空）", i, t.From)
			}
			// to 侧豁免 "Failed"：F-04 终点锚（驳回边合法目标——run.Phase=Failed 终态词形，
			// 引擎 ResumeSignal/收束判定专治；活体教训 2026-11-28 语义闸首拍即拦了存量驳回边）。
			if !phaseSet[t.To] && t.To != "Failed" {
				add("transitions[%d] to=%q 未在 phases 声明（落点不可达；Failed 终点锚豁免）", i, t.To)
			}
		}
	}

	// ④ 入口：node 可解析（节点名或相位）+events 非空。
	for i, ep := range doc.Spec.EntryPoints {
		if ep.Node == "" {
			add("entryPoints[%d] 缺 node", i)
			continue
		}
		if !nodeSet[ep.Node] && !phaseSet[ep.Node] && (ep.Phase == "" || !phaseSet[ep.Phase]) {
			add("entryPoints[%d] node=%q 不在节点名集/相位集（入口不可达）", i, ep.Node)
		}
		if len(ep.Events) == 0 {
			add("entryPoints[%d] 缺 events（入口事件词形——无事件则永不受理）", i)
		}
	}

	// ⑤ 触发事件类型注册（sch=nil 跳过——纯声明校验场景）。
	if sch != nil {
		for _, typ := range doc.Spec.Triggers.Events.Types {
			if _, err := sch.LookupEventType(typ); err != nil {
				add("triggers.events.types %q 未在 scheme.eventTypes 注册（触发永不命中——先注册事件词）", typ)
			}
		}
	}

	// ⑥ 审批策略词。
	if m := doc.Spec.Approval.Mode; !validApprovalModes[m] {
		add("approval.mode=%q 不在词汇 [auto|agent|human]（空=human 缺省）", m)
	}

	if len(errs) == 0 {
		return nil
	}
	return fmt.Errorf("workflow 声明未过语义准入（%d 项）：%s", len(errs), strings.Join(errs, "；"))
}
