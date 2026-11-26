// eventscope_test.go —— BUG-010 scope 词汇/派生/匹配机读锚（event.md 规则 19）。
package eventscope

import "testing"

func TestValid(t *testing.T) {
	cases := []struct {
		scope string
		want  bool
	}{
		{"builtin", true},
		{"global", true},
		{"systems/core", true},
		{"systems/core/instances/headless", true},
		{"", false},
		{"cored", false},
		{"systems", false},
		{"systems/core/instances", false},
		{"other/core", false},
		{"systems//instances/x", false},
	}
	for _, c := range cases {
		if got := Valid(c.scope); got != c.want {
			t.Errorf("Valid(%q)=%v 期望 %v", c.scope, got, c.want)
		}
	}
}

func TestFromInvolved(t *testing.T) {
	cases := []struct {
		involved string
		want     string
	}{
		{"Agent/global/root-agent", "global"},
		{"Task/systems/core/root-task", "systems/core"},
		{"Agent/systems/core/instances/headless/root-agent", "systems/core/instances/headless"}, // 四层 scope 整段
		{"ServiceInstance/systems/core/instances/headless", "global"},                          // 3 段 scope 不合法=兜底
		{"Session/cored/session-abc", "global"}, // 畸形 scope 段=保守兜底
		{"Agent/h1", "global"},                  // 无 scope 段
		{"bare", "global"},
	}
	for _, c := range cases {
		if got := FromInvolved(c.involved); got != c.want {
			t.Errorf("FromInvolved(%q)=%q 期望 %q", c.involved, got, c.want)
		}
	}
}

func TestMatch(t *testing.T) {
	cases := []struct {
		patterns []string
		scope    string
		want     bool
	}{
		{[]string{"*"}, "systems/anything", true},
		{[]string{"systems/core/*"}, "systems/core", true},              // 子树含根
		{[]string{"systems/core/*"}, "systems/core/instances/x", true},  // 子树成员
		{[]string{"systems/core/*"}, "systems/other", false},            // 邻树不串
		{[]string{"global"}, "global", true},
		{[]string{"global"}, "builtin", false},
		{[]string{"builtin", "global"}, "builtin", true},
		{nil, "global", false}, // 未声明=零命中（保守缺省由闸门定）
		{[]string{}, "global", false},
	}
	for _, c := range cases {
		if got := Match(c.patterns, c.scope); got != c.want {
			t.Errorf("Match(%v, %q)=%v 期望 %v", c.patterns, c.scope, got, c.want)
		}
	}
}
