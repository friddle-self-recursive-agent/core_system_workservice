package approvaldiff

import (
	"strings"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
)

func TestApprovalJudgeDocName(t *testing.T) {
	if got := approvalJudgeDocName(112); got != "inneragent-approval-run112" {
		t.Fatalf("自审批判决文档名不符: %s", got)
	}
	if approvalJudgeDocName(1) == approvalJudgeDocName(2) {
		t.Fatal("不同 run 判决文档名不得碰撞")
	}
}

func TestParseVerdictJSON(t *testing.T) {
	cases := []struct {
		raw       string
		granted   bool
		risk      string
		wantError bool
	}{
		{`{"granted":true,"risk":"low","reason":"测试全绿"}`, true, "low", false},
		{"```json\n{\"granted\":false,\"risk\":\"high\",\"reason\":\"删库\"}\n```", false, "high", false},
		{`{"granted":true,"risk":"???","reason":"x"}`, true, "high", false}, // 词形非法=失败收敛
		{`{"granted":true,"reason":"缺 risk"}`, true, "high", false},
		{`前后散文`, false, "", true},
	}
	for _, c := range cases {
		v, err := parseVerdictJSON(c.raw)
		if (err != nil) != c.wantError {
			t.Errorf("parseVerdictJSON(%q) err=%v wantError=%v", c.raw, err, c.wantError)
			continue
		}
		if err == nil && (v.Granted != c.granted || v.Risk != c.risk) {
			t.Errorf("parseVerdictJSON(%q) = %+v", c.raw, v)
		}
	}
}

func TestParseVerdictDocIdempotentRead(t *testing.T) {
	v := workflow.ApprovalVerdict{Granted: false, Risk: "medium", Reason: "涉及权限"}
	doc := approvalVerdictDoc(workflow.RunRecord{Run: 42}, "Workflow/global/x", approvalJudgeDocName(42),
		workflow.ApprovalPolicy{AutoCriteria: "准则", MaxRiskLevel: "low"}, "escalate", v)
	got, ok := parseVerdictDoc(doc)
	if !ok || got.Granted || got.Risk != "medium" || !strings.Contains(got.Reason, "涉及权限") {
		t.Fatalf("幂等回读不符: ok=%v v=%+v\ndoc=%s", ok, got, doc)
	}
}
