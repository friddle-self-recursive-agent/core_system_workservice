package intake

import (
	"context"
	"path/filepath"
	"strings"
	"testing"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

func loadRealScheme(t *testing.T) *scheme.Scheme {
	t.Helper()
	p := filepath.Join("..", "..", "..", "core_design", "core", "scheme.yaml")
	sch, err := scheme.LoadFile(p)
	if err != nil {
		t.Fatalf("LoadFile(%s): %v", p, err)
	}
	return sch
}

func TestIntake_RUX1_FreeDescriptionAndAutoFill(t *testing.T) {
	sch := loadRealScheme(t)
	eng := NewEngine(sch, nil, nil, []string{"low"})
	ctx := context.Background()
	subj := &permission.Subject{User: "alice", Roles: []string{"role/global/admin"}}

	// 第 1 轮：建立会话，自动补充 proposer=alice, priority=low, landing.target=doc
	r1, err := eng.Turn(ctx, TurnRequest{
		Kind:    "Proposal",
		Scope:   "global",
		Subject: subj,
	})
	if err != nil {
		t.Fatalf("Turn 1: %v", err)
	}
	if r1.Proposer != "alice" || r1.Fields["proposer"] != "alice" {
		t.Fatalf("proposer 自动补充失败: %+v", r1)
	}
	if r1.Fields["priority"] != "low" || r1.Fields["landing.target"] != "doc" {
		t.Fatalf("缺省自动补充失败: fields=%+v", r1.Fields)
	}
	for _, m := range r1.MissingFields {
		if m == "proposer" || m == "priority" || m == "landing.target" || m == "status" {
			t.Fatalf("已自动推断项不应出现在 missingFields 中重复盘问: %s (missing=%v)", m, r1.MissingFields)
		}
	}
	if !strings.Contains(r1.Reply, "示例") {
		t.Fatalf("提问必须带默认值与示例，得: %s", r1.Reply)
	}

	// 第 2 轮：一段自由描述抽取全部剩余字段
	r2, err := eng.Turn(ctx, TurnRequest{
		SessionID: r1.SessionID,
		Kind:      "Proposal",
		Message:   "标题：编辑器深色模式；目标：夜间巡检降低视觉疲劳；验收标准：1. 主题切换即时生效；2. 对比度满足 WCAG AA；影响面：console 编辑器前端；优先级：high",
		Subject:   subj,
	})
	if err != nil {
		t.Fatalf("Turn 2: %v", err)
	}
	if !r2.Ready || r2.Turns > 6 || len(r2.MissingFields) != 0 {
		t.Fatalf("自由描述后应字段全绿且 ≤6 轮：ready=%v turns=%d missing=%v", r2.Ready, r2.Turns, r2.MissingFields)
	}
	if r2.Fields["priority"] != "high" {
		t.Fatalf("优先级抽取应为 high，得 %v", r2.Fields["priority"])
	}
}

func TestIntake_RUX1_StepByStepWithinSixTurns(t *testing.T) {
	sch := loadRealScheme(t)
	eng := NewEngine(sch, nil, nil, nil)
	ctx := context.Background()

	// 轮 1：直接给标题
	r1, err := eng.Turn(ctx, TurnRequest{Kind: "Proposal", Message: "支持提案对话式录入"})
	if err != nil {
		t.Fatalf("Turn 1: %v", err)
	}
	if r1.AskingField != "summary" {
		t.Fatalf("轮 1 后应问 summary，得 %s", r1.AskingField)
	}
	// 轮 2：回答目标
	r2, err := eng.Turn(ctx, TurnRequest{SessionID: r1.SessionID, Kind: "Proposal", Message: "用对话+字段卡替代旧表单"})
	if err != nil {
		t.Fatalf("Turn 2: %v", err)
	}
	if r2.AskingField != "acceptance" {
		t.Fatalf("轮 2 后应问 acceptance，得 %s", r2.AskingField)
	}
	// 轮 3：回答验收标准
	r3, err := eng.Turn(ctx, TurnRequest{SessionID: r1.SessionID, Kind: "Proposal", Message: "1. 零起点不超过 6 轮全绿\n2. 过 R-PUB lint"})
	if err != nil {
		t.Fatalf("Turn 3: %v", err)
	}
	if r3.AskingField != "impact" {
		t.Fatalf("轮 3 后应问 impact，得 %s", r3.AskingField)
	}
	// 轮 4：跳过 impact 使用默认值
	r4, err := eng.Turn(ctx, TurnRequest{SessionID: r1.SessionID, Kind: "Proposal", SkipCurrent: true})
	if err != nil {
		t.Fatalf("Turn 4: %v", err)
	}
	if !r4.Ready || r4.Turns > 6 {
		t.Fatalf("4 轮（≤6 轮）后应字段全绿：ready=%v turns=%d missing=%v", r4.Ready, r4.Turns, r4.MissingFields)
	}
}

func TestIntake_GenerateAndDryRunFieldAttribution(t *testing.T) {
	sch := loadRealScheme(t)
	eng := NewEngine(sch, nil, nil, nil)
	ctx := context.Background()
	subj := &permission.Subject{User: "console-operator", Roles: []string{"role/global/admin"}}

	// 1. R-PUB-1 IP + 部署路径 + 明文 secret + 非法枚举 综合定位测试
	gBad, err := eng.Generate(ctx, GenerateRequest{
		Kind:  "Proposal",
		Scope: "global",
		Name:  "prop-bad",
		Fields: map[string]any{
			"title":      "",
			"summary":    "部署在 /home/admin/app/ 目录",
			"acceptance": []string{"连接 10.10.1.2 验证"},
			"impact":     "api_key=sk-1234567890abcdef",
			"priority":   "urgent",
		},
		Subject: subj,
	})
	if err != nil {
		t.Fatalf("Generate bad: %v", err)
	}
	if gBad.Ok {
		t.Fatalf("含多项违规时 dry-run 必须为 false")
	}
	errByField := map[string][]string{}
	for _, fe := range gBad.FieldErrors {
		errByField[fe.Field] = append(errByField[fe.Field], fe.Rule)
	}
	for _, expectField := range []string{"title", "summary", "acceptance", "impact", "priority"} {
		if len(errByField[expectField]) == 0 {
			t.Errorf("期望定位到字段 %q 的错误，实际错误表=%+v", expectField, gBad.FieldErrors)
		}
	}

	// 2. 修正后生成通过，且引用式 secret (secret:my-token) 放行
	gOk, err := eng.Generate(ctx, GenerateRequest{
		SessionID: gBad.SessionID,
		Kind:      "Proposal",
		Scope:     "global",
		Name:      "prop-good",
		Fields: map[string]any{
			"title":      "合规提案标题",
			"summary":    "部署在 <部署根目录>/ 下，凭据走 secret:provider-key 引用",
			"acceptance": []string{"零起点 ≤6 轮对话全绿", "生成对象过 spec 准入"},
			"impact":     "console 提案录入屏",
			"priority":   "high",
		},
		Subject: subj,
	})
	if err != nil {
		t.Fatalf("Generate ok: %v", err)
	}
	if !gOk.Ok || len(gOk.FieldErrors) != 0 {
		t.Fatalf("合规提案应 dry-run 全绿：ok=%v errs=%+v", gOk.Ok, gOk.FieldErrors)
	}
	if gOk.GenVersion != 2 {
		t.Fatalf("同一会话第二次生成版本应为 2，得 %d", gOk.GenVersion)
	}
	if gOk.AutoCompleted["proposer"] != "console-operator" || gOk.AutoCompleted["status"] != "submitted" {
		t.Fatalf("系统自动补全缺失：%+v", gOk.AutoCompleted)
	}
}
