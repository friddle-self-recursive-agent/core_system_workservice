// bind.go 阈值声明面绑定解析（v0.33 MetricsInfo 数据面首段随行件——判定谓词落声明面的
// 机读落点，event.md 规则 13/§5）：WorkService 对象 observation.metrics 的告警面声明
// （scheme.yaml thresholds 行——workservice.md §observation 告警面）→ 信号面桥阈值绑定。
// 声明面词形谓词（gt|gte|lt|lte——scheme 枚举机读词形）↔ 桥符号谓词（"/>=\"/<\"/<=",
// metricsrc.Threshold 契约不变）的映射只在此处发生；声明面外词形不绑定不即兴语义（该行
// 静默跳过——判定面拒绝同口径，未越限样本静默同范式）。形态合法性归写路径 scheme 校验
// 闸门，此处只做声明→绑定的纯映射。
package metricsrc

import "gopkg.in/yaml.v3"

// opWords 声明面词形谓词 ↔ 桥符号谓词（最小集同 judge——声明面外词形不映射）。
var opWords = map[string]string{
	"gt":  ">",
	"gte": ">=",
	"lt":  "<",
	"lte": "<=",
}

// wsDecl WorkService 对象原文的观测面声明形态（scheme.yaml WorkService.spec.observation：
// metrics 行 {name, query, thresholds}——未知字段放行口径，只取所需）。
type wsDecl struct {
	Spec struct {
		Observation struct {
			Metrics []struct {
				Name       string `yaml:"name"`
				Query      string `yaml:"query"`
				Thresholds []struct {
					Op    string  `yaml:"op"`
					Limit float64 `yaml:"limit"`
				} `yaml:"thresholds"`
			} `yaml:"metrics"`
		} `yaml:"observation"`
	} `yaml:"spec"`
}

// WorkserviceThresholds 从一份 WorkService 对象原文解析告警面声明 → 桥阈值绑定。
// 未声明/损坏原文/谓词词形声明面外=零绑定（不冒充空判定面——绑定面缺口可见口径）。
func WorkserviceThresholds(raw []byte) []Threshold {
	var d wsDecl
	if err := yaml.Unmarshal(raw, &d); err != nil {
		return nil
	}
	var out []Threshold
	for _, m := range d.Spec.Observation.Metrics {
		for _, th := range m.Thresholds {
			op, ok := opWords[th.Op]
			if !ok || m.Name == "" {
				continue // 声明面外谓词/无名指标不绑定（判定面拒绝同口径）
			}
			out = append(out, Threshold{Metric: m.Name, Op: op, Limit: th.Limit})
		}
	}
	return out
}

// WorkserviceInvolved 对象 involvedRef 投影（v0.45 装配段：involvedRef kind/scope/name
// 三段同构惯例——event.md 规则 11 首段同款；两段式不可被 owner/scope 谓词解析，console
// 事件详情对象链同口径死链）。scope 空=对象身份不完整，返回空串（绑定面拒绝，缺口可见
// ——不冒充两段式旧形态）。
func WorkserviceInvolved(scope, name string) string {
	if scope == "" || name == "" {
		return ""
	}
	return "WorkService/" + scope + "/" + name
}

// BindWorkservice 单对象阈值桥绑定（装配层枚举→源绑定的完整投影，v0.45 装配段收编）：
// 阈值解析（WorkserviceThresholds）+involved 三段同构（WorkserviceInvolved）一步到位。
// 零阈值声明/空 scope/空名=ok=false（零绑定缺口可见——不冒充空判定面不即兴身份）。
func BindWorkservice(scope, name string, raw []byte) (Source, bool) {
	ths := WorkserviceThresholds(raw)
	if len(ths) == 0 {
		return Source{}, false
	}
	inv := WorkserviceInvolved(scope, name)
	if inv == "" {
		return Source{}, false
	}
	return Source{Name: name, Involved: inv, Thresholds: ths}, true
}
