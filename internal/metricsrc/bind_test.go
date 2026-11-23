// bind_test.go 阈值声明面绑定单测（v0.45 装配段）：WorkserviceThresholds 声明→绑定纯映射
// （声明面词形谓词 gt|gte|lt|lte ↔ 桥符号；声明面外词形/无名指标不绑定）+ WorkserviceInvolved
// 三段同构（involvedRef kind/scope/name 惯例——空 scope/空名=空串不冒充旧两段式）+
// BindWorkservice 一步投影（零阈值/身份不完整=ok=false 缺口可见）。设计依据：concepts/event.md
// 规则 13 + cored-bridge.md v0.33/v0.45。
package metricsrc

import "testing"

const bindFixtureDecl = `
apiVersion: core/v1
kind: WorkService
metadata: {name: notify, scope: global}
spec:
  observation:
    metrics:
      - name: events.DocApplied
        query: "event_log 窗口计数（自观测采样器投影）"
        thresholds:
          - {op: gte, limit: 1}
`

func TestBindWorkserviceThreeSegment(t *testing.T) {
	src, ok := BindWorkservice("global", "notify", []byte(bindFixtureDecl))
	if !ok {
		t.Fatal("声明在册∧身份完整 → 应绑定")
	}
	if src.Name != "notify" {
		t.Fatalf("源名 = %q, 要 notify", src.Name)
	}
	// 三段同构：kind/scope/name——owner/scope 谓词与 console 事件详情对象链可解析的唯一形态。
	if src.Involved != "WorkService/global/notify" {
		t.Fatalf("involved = %q, 要 WorkService/global/notify（三段同构）", src.Involved)
	}
	if len(src.Thresholds) != 1 || src.Thresholds[0].Metric != "events.DocApplied" ||
		src.Thresholds[0].Op != ">=" || src.Thresholds[0].Limit != 1 {
		t.Fatalf("阈值绑定不符: %+v", src.Thresholds)
	}
}

func TestBindWorkserviceZeroThresholds(t *testing.T) {
	const noDecl = `
apiVersion: core/v1
kind: WorkService
metadata: {name: bare, scope: global}
spec: {category: tool}
`
	if _, ok := BindWorkservice("global", "bare", []byte(noDecl)); ok {
		t.Fatal("零告警面声明 → 应零绑定（缺口可见不冒充）")
	}
	// 声明面外谓词词形：逐行静默跳过 → 全跳=零绑定。
	const badWord = `
apiVersion: core/v1
kind: WorkService
metadata: {name: w, scope: global}
spec:
  observation:
    metrics:
      - {name: events.X, query: q, thresholds: [{op: within, limit: 5}]}
`
	if _, ok := BindWorkservice("global", "w", []byte(badWord)); ok {
		t.Fatal("声明面外谓词词形 → 应零绑定（不即兴语义）")
	}
}

func TestBindWorkserviceIncompleteIdentity(t *testing.T) {
	// 空 scope（对象身份不完整——loop-44 前旧形态对象）→ 拒绝绑定，不冒充两段式旧形态。
	if _, ok := BindWorkservice("", "notify", []byte(bindFixtureDecl)); ok {
		t.Fatal("空 scope → 应零绑定（身份不完整缺口可见）")
	}
	if got := WorkserviceInvolved("", "notify"); got != "" {
		t.Fatalf("空 scope involved = %q, 要空串", got)
	}
	if got := WorkserviceInvolved("global", ""); got != "" {
		t.Fatalf("空名 involved = %q, 要空串", got)
	}
}

func TestWorkserviceThresholdsMultiPredicate(t *testing.T) {
	const multi = `
apiVersion: core/v1
kind: WorkService
metadata: {name: w, scope: global}
spec:
  observation:
    metrics:
      - {name: events.A, query: q, thresholds: [{op: gt, limit: 10}, {op: lt, limit: 2}]}
      - {name: "", query: q, thresholds: [{op: gte, limit: 1}]}
`
	ths := WorkserviceThresholds([]byte(multi))
	if len(ths) != 2 {
		t.Fatalf("阈值数 = %d, 要 2（无名指标行跳过）", len(ths))
	}
	if ths[0].Op != ">" || ths[0].Limit != 10 || ths[1].Op != "<" || ths[1].Limit != 2 {
		t.Fatalf("词形映射不符: %+v", ths)
	}
}
