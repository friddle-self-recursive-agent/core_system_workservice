// scope_test.go —— BUG-010 scope 闸门三路判定机读锚（event.md 规则 19②③：派发/消费双侧）。
package flowbind

import "testing"

// stubScopes 声明解析桩（patterns/own/ok 三态可控）。
type stubScopes struct {
	patterns []string
	own      string
	ok       bool
}

func (s stubScopes) EventScopes(string) ([]string, string, bool) {
	return s.patterns, s.own, s.ok
}

func TestAgentScopeGateJudge(t *testing.T) {
	cases := []struct {
		name    string
		res     AgentEventScopes
		scope   string
		want    bool
		reason  string
		comment string
	}{
		{
			name: "锚不在册只收 global", res: stubScopes{ok: false},
			scope: "global", want: true, comment: "保守缺省放行 global",
		},
		{
			name: "锚不在册拒 systems", res: stubScopes{ok: false},
			scope: "systems/core", want: false, reason: BlockReasonNoAgent,
		},
		{
			name: "在册未声明只收 global", res: stubScopes{ok: true, own: "systems/core"},
			scope: "global", want: true,
		},
		{
			name: "在册未声明收自身 scope", res: stubScopes{ok: true, own: "systems/core"},
			scope: "systems/core", want: true, comment: "own=第二收窄项",
		},
		{
			name: "在册未声明拒他人 scope", res: stubScopes{ok: true, own: "systems/core"},
			scope: "systems/other", want: false, reason: BlockReasonUndeclared,
		},
		{
			name: "已声明子树放行", res: stubScopes{ok: true, patterns: []string{"systems/core/*"}},
			scope: "systems/core/instances/i1", want: true,
		},
		{
			name: "已声明越界拒", res: stubScopes{ok: true, patterns: []string{"systems/core/*"}},
			scope: "systems/other", want: false, reason: BlockReasonOutOfDeclaration,
		},
		{
			name: "通配全收", res: stubScopes{ok: true, patterns: []string{"*"}},
			scope: "builtin", want: true,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			g := &AgentScopeGate{Resolver: c.res}
			got, reason := g.Judge(c.scope, "some-agent")
			if got != c.want {
				t.Fatalf("Judge(%q)=%v 期望 %v（%s）", c.scope, got, c.want, c.comment)
			}
			if !got && reason != c.reason {
				t.Fatalf("拦截因由=%q 期望 %q", reason, c.reason)
			}
		})
	}
}

func TestAgentScopeGateNilResolverDenies(t *testing.T) {
	// 装配缺口=宁拒不降级（默认拒绝+告警回调可见）。
	var blocked *ScopeBlockedBlock
	g := &AgentScopeGate{OnBlocked: func(b ScopeBlockedBlock) { blocked = &b }}
	if g.Allow("global", "a") {
		t.Fatal("nil Resolver 应恒拒（宁拒不降级）")
	}
	if blocked == nil || blocked.Reason != BlockReasonNoAgent {
		t.Fatalf("nil Resolver 拦截应触发 OnBlocked 告警，得 %+v", blocked)
	}
}

func TestAgentScopeGateAllowEmitsBlocked(t *testing.T) {
	// 派发闸拦截必经 OnBlocked（可见不静默——规则 19③ 的机械落点）。
	var calls int
	g := &AgentScopeGate{
		Resolver:  stubScopes{ok: true, patterns: []string{"systems/core/*"}},
		OnBlocked: func(ScopeBlockedBlock) { calls++ },
	}
	if g.Allow("systems/other", "a") {
		t.Fatal("越界应拦")
	}
	if calls != 1 {
		t.Fatalf("OnBlocked 调用数=%d 期望 1（拦截必有痕）", calls)
	}
	if !g.Allow("systems/core", "a") {
		t.Fatal("声明内应放行")
	}
	if calls != 1 {
		t.Fatalf("放行不应产生告警，调用数=%d", calls)
	}
}
