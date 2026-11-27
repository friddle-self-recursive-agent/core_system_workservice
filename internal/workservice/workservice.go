// Package workservice 六面体（v0.13，A53）：外部系统的完整管理描述。
// 设计依据：core_system/concepts/workservice.md。
// 六面：connector（含原 ConnectorProvider）/ observation / data / flows / runtime / tools。
// 层级判例：System=deploy_system → 实例=业务域开发实例 → WorkService=jenkins/onedev/bug-management。
package workservice

// ConnectorFacet 连接面（原 ConnectorProvider 并入，A53）。
type ConnectorFacet struct {
	Types          []string       // 类型标注（scm/ci…）
	Impl           Impl           // tool | image | source(+Ref)
	Download       string         // image:// git:// pkg://（必填）
	Scheme         map[string]any // 使用方式 schema
	Usage          string         // 人读
	LLMPrompt      string         // LLM 注入提示（必填）
	NameResolution string         // 无前缀=默认org；带前缀=github.com/<org>/<name>（A38）
}

// Impl 实现绑定三形态。
type Impl struct {
	Kind string // tool | image | source
	Ref  string // source 形态：Ref 指代码地址
}

// Spec 六面体。
type Spec struct {
	Category    string         // developer | storage | tool | business…
	Connector   ConnectorFacet // 连接面
	Observation struct {       // 观测面（意图；数据在 MetricsInfo）
		Metrics []struct{ Name, Query string }
		LogsRef string
	}
	Data struct { // 数据面（Ref 解析权威之一）
		Entities []struct{ Name, Store, Access string }
	}
	Flows []struct { // 流程面
		Name   string
		Stages []string
	}
	Runtime struct { // 运行状态观测面
		Probes []struct {
			HTTPGet string
			Version string
			Period  string
		}
	}
	Tools []Tool // 工具与调用描述面
}

// Tool 工具条目：cleanup 成对声明（A51，代替 Finalizer）。
type Tool struct {
	Op         string
	Params     map[string]string
	Idempotent bool
	Undo       string // full | partial | none
	Cleanup    string // 成对清理 op（有外部创建动作必填，Lint 强制）
}
