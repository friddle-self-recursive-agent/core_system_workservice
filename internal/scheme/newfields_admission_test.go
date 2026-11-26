// newfields_admission_test.go —— 锚定 519fbd2 新 scheme 字段过准入（BUG-012/013/014/010 机读面）：
// rollback 四类/trustLevel/trustGate/applyConflictGate/eventScopes 声明进 scheme 后，
// 携这些字段的文档必须通过 ValidateDoc（未知字段放行+新枚举不误伤）。
// 文档形状注意：md 文档 spec 必须在 frontmatter 内（body 不参与准入解析——docSpecNode 只取围栏内）。
package scheme

import (
	"os"
	"strings"
	"testing"
)

// loadCoreScheme 读本机 core_design scheme.yaml（绝对路径注入，避免测试依赖仓库相对布局）。
func loadCoreScheme(t *testing.T) *Scheme {
	t.Helper()
	// CORE_SYSTEM_SCHEME 环境变量优先（与 cored 启动旗标同源词形；缺省=Mac 侧开发机布局）。
	path := strings.TrimSpace(os.Getenv("CORE_SYSTEM_SCHEME"))
	if path == "" {
		path = "/Users/friddle/Project/self-recursive-agent/core_design/core/scheme.yaml"
	}
	sch, err := LoadFile(path)
	if err != nil {
		t.Fatalf("装载 scheme：%v", err)
	}
	return sch
}

func TestNewSchemeFieldsAdmission(t *testing.T) {
	sch := loadCoreScheme(t)

	cases := []struct {
		name string
		kind string
		doc  string
	}{
		{
			name: "WorkService.tools.rollback 四类回滚",
			kind: "WorkService",
			doc: `---
kind: WorkService
scope: probe-s
name: ws-rollback
spec:
  tools:
    - op: deploy
      params: [env]
      idempotent: false
      rollback:
        policy: ask
        type: agent
        description: 按 runbook 回滚发布
---
# ws
`,
		},
		{
			name: "ConnectorOps.trustLevel external",
			kind: "ConnectorOps",
			doc: `---
kind: ConnectorOps
scope: probe-s
name: ops-trust
spec:
  workservice: ws-rollback
  trustLevel: external
  ops: [deploy]
---
# ops
`,
		},
		{
			name: "Workflow.trustGate+onFailed.rollbackPolicy",
			kind: "Workflow",
			doc: `---
kind: Workflow
scope: probe-s
name: wf-trust
spec:
  phases: [{id: p1}]
  onFailed:
    backoff: 1m
    maxAttempts: 2
    rollbackPolicy: by-step
    escalate: {to: root}
  trustGate:
    enabled: true
    mode: agent
    onReject: ask
---
# wf
`,
		},
		{
			name: "Agent.eventScopes 声明",
			kind: "Agent",
			doc: `---
kind: Agent
scope: probe-s
name: agent-scoped
spec:
  eventScopes: ["probe-s/*", "global"]
---
# agent
`,
		},
		{
			name: "TaskScheme.steps.rollback+onFailure.rollbackPolicy",
			kind: "TaskScheme",
			doc: `---
kind: TaskScheme
scope: probe-s
name: ts-rollback
spec:
  aspect: deploy
  steps:
    - id: s1
      connector: ops-trust
      op: deploy
      args: {}
      rollback:
        policy: optional
        type: human
        description: 人按文档回滚
  onFailure:
    retry: 1
    backoff: 30s
    rollbackPolicy: auto
    then: escalate
---
# ts
`,
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			k, err := sch.Lookup(tc.kind)
			if err != nil {
				t.Fatalf("scheme 未注册 kind=%s（先确认 core_design scheme 已含 519fbd2 字段）：%v", tc.kind, err)
			}
			if err := ValidateDoc(k, []byte(tc.doc)); err != nil {
				t.Fatalf("新字段文档被准入拒绝（BUG-012/013/014 回潮）：%v", err)
			}
		})
	}
}
