// execution.go —— 构建生效链判别投影（v0.55，R-9——console-adoption.md §1.29）。
// 设计依据：instance 仓 runtime/console-adoption.md §4 落地顺序表 R-9 项（workflow 运行管理屏
// 「构建产物 digest/构建生效链」判别面——/v1/workflows 聚合面 execution 投影（state=
// bound|unbound+connector+skeleton）已落（v0.49），余量=「run 相位推进产生的效果与 build 产物
// 对账」的判别面：对账=execution.skeleton 翻转 + apply_log 对象 rv + 运行抽屉涉及文档对象链
// 三者联动可查）+ workflow-runtime.md §11.2 第三段（拉起投影面——派单受理行携节点接线载荷，
// apply-workflow 拉起输入面首件）。
//
// 判别面三问（机读，缺一即缺口词形可见）：
//  1. 骨架产物 digest=骨架对象落库 rv（skeletonRv——构建产物的版本指纹；absent=0）；
//  2. 生效对象 rv=ConnectorOps 面上现值（聚合面 objects 行 resourceVersion 同源）；
//  3. 链态=骨架 rv 与最近一次绑定器 apply 行（apply_log fieldManager=workflow:bind ok 行）
//     的 rv 对账——一致（consistent）/ 人写演进后未再物化（divergent——骨架翻转为实现后
//     物化器零盲写，链照实分裂可见）/ 对象缺席（absent）。
//
// 解耦口径：本件纯投影零写路径——apply_log 行是审计账（ops-log.md §2），读面聚合现读
// 现投零缓存；digest 不发明新词形（rv=权威版本指纹，git 提交数 ≙ ok 行数对账不变量在册）。
package flowbind

import (
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"gopkg.in/yaml.v3"
)

// 绑定器 apply 溯源词汇（Ensure 写面 reasonRef 前缀——apply_log 行回指锚）。
const BindReasonPrefix = "workflow-bind:"

// 生效链态词（chain 闭合集——缺口词形可见不静默）。
const (
	// ChainConsistent 骨架 rv=最近绑定 apply 行 rv（创建/更新/noop 幂等重入同窗一致）。
	ChainConsistent = "consistent"
	// ChainDivergent 面上 rv>绑定 apply 行 rv——对象在绑定器之外演进过（人/agent 实现
	// 翻转或外部编辑），物化器零盲写=链照实分裂可见（实现侧 diff 基线=apply 行 rv）。
	ChainDivergent = "divergent"
	// ChainAbsent 绑定对象缺席（unbound 同窗词形——骨架产物不存在，digest=0）。
	ChainAbsent = "absent"
)

// Reconcile 构建生效链判别投影（/v1/workflows 行集 reconcile 键——v0.55 新载荷键=代际判别器）。
type Reconcile struct {
	// SkeletonRV 骨架产物 digest=绑定对象落库 rv（absent=0——无产物无指纹）。
	SkeletonRV int64 `json:"skeletonRv"`
	// Bound/Skeleton 与 Execution 投影同源复读（行内自洽——消费方零跨键拼装）。
	Bound    bool `json:"bound"`
	Skeleton bool `json:"skeleton"`
	// LastBindApply 最近一次绑定器 apply 判定（ok=落库|noop=幂等重入零差异；缺席=空——
	// 装配期物化与热路径同走 apply 引擎，正常部署必有行；空=审计账缺行，如实投影）。
	LastBindApply string `json:"lastBindApply"`
	// LastBindRV 最近绑定器 apply 行的 rv（ok 行=落库后版本；noop 行=面上现值——幂等重入
	// 照落现值；缺席=0）。实现侧 diff 基线（R-8 派单行 rv 同词汇）。
	LastBindRV int64 `json:"lastBindRv"`
	// LastBindAt 最近绑定器 apply 行时刻（RFC3339Nano；缺席=空串）。
	LastBindAt string `json:"lastBindAt"`
	// Chain 生效链态（consistent|divergent|absent 闭合集——缺口词形可见）。
	Chain string `json:"chain"`
	// Nodes 骨架节点接线投影（spec.flow.nodes 同源单点真相——派单行 nodes 同一投影；
	// absent=空）。聚合面 executionFlow 键与派单拉起输入面共用。
	Nodes []Node `json:"nodes,omitempty"`
}

// NodesOf 骨架对象 spec.flow.nodes 投影（宽松解析——声明面外不即兴；解析失败=零节点）。
// 派单行与聚合面共用同一投影（单一真相零硬编码——apply-workflow 拉起输入面=骨架节点清单）。
func NodesOf(content []byte) []Node {
	var d struct {
		Spec struct {
			Flow flowDecl `yaml:"flow"`
		} `yaml:"spec"`
	}
	if err := yaml.Unmarshal(content, &d); err != nil {
		return nil
	}
	return d.Spec.Flow.Nodes
}

// ReconcileOf 生效链判别（纯函数）：execution 投影+绑定对象（bound 时非 nil）+该锚的
// apply_log 行集（QueryApplies 全量过滤由调用方裁剪）→ 判别投影。行集扫描口径：绑定器行
// =fieldManager=fieldManagerApply ∧ reasonRef 前缀=BindReasonPrefix；行序任意（调用方 rv
// 降序/落库序均可——本函数取「rv 最大者」为最近行，与 QueryApplies 排序口径解耦）。
func ReconcileOf(e *Execution, conn *repository.Object, applies []opslog.ApplyLog) *Reconcile {
	rc := &Reconcile{}
	if e != nil {
		rc.Bound = e.State == ExecBound
		rc.Skeleton = e.Skeleton
	}
	if conn == nil {
		rc.Chain = ChainAbsent // 对象缺席=无产物无指纹（unbound 同窗缺口可见）
		return rc
	}
	rc.SkeletonRV = conn.ResourceVersion
	rc.Nodes = NodesOf(conn.Content)
	// 最近绑定器行：apply_log 按 obj 精确过滤的行集内，fieldManager=绑定器 ∧ reasonRef
	// 前缀=workflow-bind: 的行；rv 最大者为最近（ok=落库后版本单调，noop=面上现值）。
	for i := range applies {
		a := applies[i]
		if a.FieldManager != fieldManager || !hasPrefix(a.ReasonRef, BindReasonPrefix) {
			continue
		}
		if a.ResourceVersion > rc.LastBindRV || rc.LastBindApply == "" {
			rc.LastBindRV = a.ResourceVersion
			rc.LastBindApply = a.Result
			rc.LastBindAt = a.Ts.UTC().Format(rfc3339Nano)
		}
	}
	switch {
	case rc.LastBindApply == "":
		rc.Chain = ChainDivergent // 面上有对象而绑定器审计行缺席=链缺对账锚（如实分裂可见）
	case rc.LastBindRV == conn.ResourceVersion:
		rc.Chain = ChainConsistent // ok 行 rv=落库后版本 / noop 行 rv=面上现值——同窗一致
	default:
		rc.Chain = ChainDivergent // 面上 rv 前移而绑定行未跟=绑定器之外演进过（人/agent 实现）
	}
	return rc
}

// hasPrefix strings.HasPrefix 的局部别名（零额外依赖口味——保持包内 import 面最小）。
func hasPrefix(s, p string) bool {
	return len(s) >= len(p) && s[:len(p)] == p
}

// rfc3339Nano RFC3339Nano 词形（投影格式单点）。
const rfc3339Nano = "2006-01-02T15:04:05.999999999Z07:00"
