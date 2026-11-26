// verify_gate.go —— Verify 相位硬 gate 的真实接线（G-09/G-12——R-TYPE-3 rule_type.md
// ruleChecks 三行接进 workflow Verifying 相位）：workflow 推进路径「离开 Verifying」先过
// 本 gate——调 connectops-code-audit 的 audit op（connectors/codeaudit 包）跑真实审核，
// findings 非空（error 级）/unbound → run 卡 Verify 不放行 + Escalate 事件进总线
// （source=core，involved=run 涉及对象链，dedup=verify-gate:<workflow>:<run>——总线窗口
// 去重承担重审风暴抑制）；判定 OK → 放行（判定投影随 run 留痕写穿透）。
//
// 判据锚（G-12 机械探针）：触发 R-PUB-1 违规样例 → Verify 阶段 findings 非空且 run 卡 Verify。
//
// 审核输入面：scheme `runtime.ruleChecks` 注册表行（ruleId/mode/checker/severity——每拍
// 现读热装载：Apply 即时生效）投影为 audit checks（mode=agent/human 行由连接器列 skipped
// 供对账——调用侧路由职责在声明面）；checker 引用按注册名路由：builtin cored-lint 在册；
// lint/*.sh 声明引用=命令模板（本文件装配 WithTemplate，经 lint-json-bridge.sh 桥接为
// codeaudit stdout-JSON 契约——净入脚本 exit-code/文本词形不改，适配在装配侧）。未注册
// 引用=unbound（显式失败不静默——unbound 非 ok 恒卡 Verify，绝不冒充零发现）。
//
// 被审目录=`--verify-target`（只读遍历；缺席=gate 不装配——stderr 缺口可见不静默）。
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/friddle-self-recursive-agent/core_system_workservice/connectors"
	"github.com/friddle-self-recursive-agent/core_system_workservice/connectors/codeaudit"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
)

// lintJSONBridge 净入 lint 脚本 → codeaudit 命令模板契约的桥（cmd/cored 装配伴生件，
// 同目录部署；模板词形 `sh <bridge> <script> {target}`——argv 直传不经 shell 拼接，target
// 恒单参数，codeaudit buildArgv 占位符白名单闸门照常生效）。
const lintJSONBridge = "lint-json-bridge.sh"

// verifyGateConfig Verify gate 装配项。
type verifyGateConfig struct {
	sch       *scheme.Scheme
	target    string // 被审目录（--verify-target）
	schemeDir string // scheme.yaml 目录（lint 脚本相对解析锚：<dir>/../lint/）
}

// newVerifyGate 构建 gate：connectops-code-audit 连接器（注册表自注册实现直取）+ lint
// 脚本命令模板桥接。返回 nil=gate 缺席（target 未声明——缺口由调用方 stderr 可见）。
func newVerifyGate(cfg verifyGateConfig) *verifyGate {
	if strings.TrimSpace(cfg.target) == "" {
		return nil
	}
	conn := codeaudit.New(
		codeaudit.WithSchemePath(cfg.schemePath()),
		codeaudit.WithTemplate("lint/personal-info.sh", lintTemplate(cfg, "personal-info.sh")),
		codeaudit.WithTemplate("lint/metadatainfo.sh", lintTemplate(cfg, "metadatainfo.sh")),
	)
	// V-18：同步覆盖 ops 注册表实例，使 Verify gate 与 /v1/code-audit/findings 共用同一结果 LRU。
	connectors.RegisterOpsConnector(conn)
	return &verifyGate{cfg: cfg, conn: conn}
}

// schemePath 注册表面路径（cored-lint 后端所需）。
func (c *verifyGateConfig) schemePath() string { return c.sch.Source }

// lintTemplate lint 脚本命令模板（脚本=相对 scheme.yaml 目录的 ../lint/<name>；部署侧
// 固定词形——任意字符串不入模板，target 走 {target} 占位符白名单）。
func lintTemplate(cfg verifyGateConfig, name string) string {
	bridge, _ := filepath.Abs(lintJSONBridge)
	script := filepath.Join(cfg.schemeDir, "..", "lint", name)
	if _, err := os.Stat(script); err != nil {
		return "" // 脚本缺席=空模板（不注册——引用落 unbound 显式失败，不冒充）
	}
	return "sh " + bridge + " " + script + " {target}"
}

// verifyGate workflow.VerifyGate 的 connectops-code-audit 适配（审核=只读 op 调用；
// 每拍重审幂等——修复后下拍即过）。
type verifyGate struct {
	cfg  verifyGateConfig
	conn connectors.OpsConnector
}

// Verify 实现 workflow.VerifyGate：scheme ruleChecks 行集 → audit op → 判定投影。
// 注册表缺席/不可解析=unbound 合成行（空审核按失败——绝不静默通过）。
func (g *verifyGate) Verify(run workflow.RunRecord) workflow.VerifyVerdict {
	checks := g.cfg.ruleCheckChecks()
	req := map[string]any{
		"target":     g.cfg.target,
		"checks":     checks,
		"schemePath": g.cfg.schemePath(),
	}
	params, err := json.Marshal(req)
	if err != nil {
		return workflow.VerifyVerdict{OK: false, Note: "audit 参数编码失败: " + err.Error()}
	}
	out, err := g.conn.Invoke(context.Background(), codeaudit.OpAudit, params)
	if err != nil {
		return workflow.VerifyVerdict{OK: false, Note: "connectops-code-audit 调用失败（显式失败不静默）: " + err.Error()}
	}
	var res codeaudit.Result
	if err := json.Unmarshal(out, &res); err != nil {
		return workflow.VerifyVerdict{OK: false, Note: "audit 结果不可解析（显式失败）: " + err.Error()}
	}
	v := workflow.VerifyVerdict{
		OK:      res.OK,
		AuditID: res.AuditID,
		Unbound: len(res.Unbound),
	}
	for _, f := range res.Findings {
		if f.Severity == "error" {
			v.ErrorFindings++
		} else {
			v.WarnFindings++
		}
	}
	if note := skippedNote(res.Skipped); note != "" {
		v.Note = note + "；"
	}
	if v.OK {
		v.Note += fmt.Sprintf("audit ok（auditId=%s）", v.AuditID)
	} else {
		v.Note += fmt.Sprintf("卡 Verify：error findings=%d unbound=%d（auditId=%s——修复后下拍重审即过）",
			v.ErrorFindings, v.Unbound, v.AuditID)
	}
	return v
}

// skippedNote agent/human 规则行摘要（对账可见——非本 gate 职责，列出让声明面可见）。
func skippedNote(sk []codeaudit.Skipped) string {
	if len(sk) == 0 {
		return ""
	}
	names := make([]string, 0, len(sk))
	for _, s := range sk {
		names = append(names, s.Rule)
	}
	return "mode≠code 规则不经 code-audit（调用侧路由）: " + strings.Join(names, ",")
}

// ruleCheckChecks scheme runtime.ruleChecks 注册表 → audit checks 行集（每拍现读——Apply
// 即时生效；行缺 ruleId=跳过（连接器侧会因缺 rule 报错——前置防御）；mode 透传（缺省 code）。
func (c *verifyGateConfig) ruleCheckChecks() []map[string]any {
	rows, _ := c.sch.Spec.Runtime["ruleChecks"].([]any)
	checks := make([]map[string]any, 0, len(rows))
	for _, r := range rows {
		m, ok := r.(map[string]any)
		if !ok {
			continue
		}
		rule, _ := m["ruleId"].(string)
		if rule == "" {
			continue
		}
		row := map[string]any{"rule": rule}
		if v, ok := m["mode"].(string); ok && v != "" {
			row["mode"] = v
		}
		if v, ok := m["checker"].(string); ok && v != "" {
			row["checker"] = v
		}
		if v, ok := m["severity"].(string); ok && v != "" {
			row["severity"] = v
		}
		checks = append(checks, row)
	}
	return checks
}

// verifyEscalatePublisher 卡 Verify → Escalate 事件（既有注册类型 root 通道——G-09/G-12
// 上报面；类型未注册=零事件缺口 stderr 可见；总线 dedup=verify-gate:<workflow>:<run>）。
func verifyEscalatePublisher(bus *eventbus.Bus, sch *scheme.Scheme, logf func(string, ...any)) func(workflow.RunRecord, workflow.VerifyVerdict) {
	return func(run workflow.RunRecord, v workflow.VerifyVerdict) {
		et, err := sch.LookupEventType("Escalate")
		if err != nil {
			logf("Verify gate 卡 Verify 但 Escalate 未注册（升级事件缺席——缺口可见）: %v", err)
			return
		}
		involved := run.Involved
		if involved == "" {
			involved = "Workflow/global/" + run.Workflow
		}
		if err := bus.Publish(&eventbus.Event{
			Type: "Escalate", Source: "core", Involved: involved,
			DedupKey: "verify-gate:" + run.Workflow + ":" + fmt.Sprint(run.Run),
			Level:    et.DefaultLevel,
		}); err != nil {
			logf("Verify gate Escalate 发布失败（继续上报不静默）: %v", err)
			return
		}
		logf("run %s#%d 卡 Verify 已升级（Escalate——%s）", run.Workflow, run.Run, v.Note)
	}
}
