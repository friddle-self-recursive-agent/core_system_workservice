package approvaldiff

import "testing"

// TestDiffDocNameForRun 判决文档 per-run 命名（2026-11-28「审批为空」纠偏：幂等域=run）。
func TestDiffDocNameForRun(t *testing.T) {
	cases := []struct {
		kind, name string
		run        int64
		want       string
	}{
		{"Workflow", "gh-issue-trigger", 83, "docdiff-workflow-gh-issue-trigger-run83"},
		{"Provider", "openai-default", 79, "docdiff-provider-openai-default-run79"},
		{"Workflow", "core-system-developer-loop", 0, "docdiff-workflow-core-system-developer-loop-run0"},
	}
	for _, c := range cases {
		if got := diffDocNameForRun(c.kind, c.name, c.run); got != c.want {
			t.Errorf("diffDocNameForRun(%q,%q,%d) = %q, want %q", c.kind, c.name, c.run, got, c.want)
		}
	}
}

// TestDiffDocNameForRunDistinct 同一对象不同 run 各生成一份（旧一对象一名幂等 skip 根因的回归闸）。
func TestDiffDocNameForRunDistinct(t *testing.T) {
	a := diffDocNameForRun("Workflow", "gh-issue-trigger", 82)
	b := diffDocNameForRun("Workflow", "gh-issue-trigger", 83)
	if a == b {
		t.Fatalf("per-run 名碰撞：%q", a)
	}
}
