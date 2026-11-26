// attention.go —— GET /v1/attention 关注报告（读侧派生，R-UX-6 / concepts/attention.md）。
//
// v1 = 规则版：确定性候选收集 + facts + 证据链 + generatedBy=rules。LLM 叙述/缓存/权限指纹
// 后续段接入；当前不新增 DocType、不落 git、无写路径。
package apiserver

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/flowbind"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"gopkg.in/yaml.v3"
)

type attentionItem struct {
	CandidateID string            `json:"candidateId"`
	Category    string            `json:"category"` // status|anomaly|proposal
	Severity    string            `json:"severity"` // critical|high|medium
	Title       string            `json:"title"`
	Why         string            `json:"why"`
	Evidence    []map[string]any  `json:"evidence"`
	Action      map[string]string `json:"action"`
}

type attentionClear struct {
	CandidateID string `json:"candidateId"`
	Note        string `json:"note"`
}

var severityRank = map[string]int{"critical": 0, "high": 1, "medium": 2}

// attentionReport 组装规则版报告。所有数字来自服务端聚合（facts），叙述只用事实。
// hidden=default-hidden 类型集（BUG-010，event.md 规则 18——关注报告候选集不混入
// default-hidden 事件；调用方按 scheme 注册表逐请求现取，热刷新跟随；SessionStalled/
// Escalate 等 attention 族生效词形 default-visible 不受影响=异常照常上屏）。
func attentionReport(ctx context.Context, objects ObjectSource, runs RunFace, o *options, hidden map[string]bool) map[string]any {
	now := time.Now().UTC()
	missing := []string{}
	var attention []attentionItem
	var clear []attentionClear
	facts := map[string]any{
		"systems": 0, "instances": 0, "runs": 0, "running": 0,
		"failed": 0, "approvalsPending": 0, "errors": 0,
	}
	if objects == nil {
		missing = append(missing, "objects")
	}
	if runs == nil {
		missing = append(missing, "runs")
	}
	if o.eventLog == nil {
		missing = append(missing, "events")
	}
	if o.inbox == nil {
		missing = append(missing, "inbox")
	}

	// 1) 系统/实例：ServiceDesc 是声明，ServiceInstance 才是现实。没有实例=系统未接入现实。
	instancesByScope := map[string][]string{}
	if objects != nil {
		if insts, err := listMetaOrList(objects, ctx, "ServiceInstance", ""); err == nil {
			for _, in := range insts {
				instancesByScope[in.Scope] = append(instancesByScope[in.Scope], in.Name)
			}
		}
		if descs, err := listMetaOrList(objects, ctx, "ServiceDesc", ""); err == nil {
			facts["systems"] = len(descs)
			for _, d := range descs {
				names := instancesByScope[d.Scope]
				if len(names) == 0 {
					attention = append(attention, attentionItem{
						CandidateID: "c-sys-" + d.Scope + "-no-instance",
						Category:    "status",
						Severity:    "high",
						Title:       d.Name + " 还没有实例",
						Why:         "系统有声明，但没有 ServiceInstance 对象；现在只能看文档，看不到现实。",
						Evidence:    []map[string]any{docEvidence(ctx, objects, "ServiceDesc", d.Scope, d.Name)},
						Action:      map[string]string{"kind": "open", "target": "#/s/" + d.Name},
					})
					continue
				}
				for _, in := range names {
					phase := instancePhase(ctx, objects, d.Scope, in)
					if phase == "" {
						phase = "Unknown"
					}
					if phase != "InSync" {
						att := attentionItem{
							CandidateID: "c-inst-" + d.Scope + "-" + in,
							Category:    "status",
							Severity:    "high",
							Title:       d.Name + " 的实例 " + in + " 不在 InSync",
							Why:         "当前相位 " + phase + "；现实与声明还没对上。",
							Evidence:    []map[string]any{docEvidence(ctx, objects, "ServiceInstance", d.Scope, in)},
							Action:      map[string]string{"kind": "open", "target": "#/s/" + d.Scope + "/i/" + in},
						}
						if phase == "Failed" || phase == "Escalated" {
							att.Severity = "critical"
						}
						attention = append(attention, att)
					} else {
						clear = append(clear, attentionClear{CandidateID: "c-inst-" + d.Scope + "-" + in, Note: d.Name + " 的实例 " + in + " 已追平（InSync）。"})
					}
				}
			}
		}
		if insts, err := listMetaOrList(objects, ctx, "ServiceInstance", ""); err == nil {
			facts["instances"] = len(insts)
		}
	}

	// 2) 运行：失败/挂起=异常；失败计数进 facts。运行中不冒充成功，也不冒充异常。
	if runs != nil {
		st := runs.State()
		facts["runs"] = len(st.Runs)
		for _, r := range st.Runs {
			switch r.Status {
			case "running", "triggered":
				facts["running"] = facts["running"].(int) + 1
			case "failed":
				facts["failed"] = facts["failed"].(int) + 1
				attention = append(attention, attentionItem{
					CandidateID: fmt.Sprintf("c-run-%s-%d", r.Workflow, r.Run),
					Category:    "anomaly",
					Severity:    "high",
					Title:       "工作流 " + r.Workflow + " 的这次运行失败了",
					Why:         "触发源 " + r.Type + "，相位 " + r.Phase + "；需要看运行详情与失败原因。",
					Evidence: []map[string]any{{
						"kind": "run", "id": fmt.Sprintf("%s#%d", r.Workflow, r.Run),
						"ref": fmt.Sprintf("#/wf/%s/r/%d", r.Workflow, r.Run),
						"at":  r.StartedAt.UTC().Format(time.RFC3339),
					}},
					Action: map[string]string{"kind": "open", "target": fmt.Sprintf("#/wf/%s/r/%d", r.Workflow, r.Run)},
				})
			case "suspended":
				attention = append(attention, attentionItem{
					CandidateID: fmt.Sprintf("c-run-%s-%d", r.Workflow, r.Run),
					Category:    "proposal",
					Severity:    "high",
					Title:       "工作流 " + r.Workflow + " 在等人批准",
					Why:         "运行停在审批锚，继续之前需要人判定。",
					Evidence: []map[string]any{{
						"kind": "run", "id": fmt.Sprintf("%s#%d", r.Workflow, r.Run),
						"ref": fmt.Sprintf("#/wf/%s/r/%d", r.Workflow, r.Run),
						"at":  r.StartedAt.UTC().Format(time.RFC3339),
					}},
					Action: map[string]string{"kind": "approve", "target": "#/approvals"},
				})
			}
		}
		if facts["failed"].(int) == 0 {
			clear = append(clear, attentionClear{CandidateID: "c-runs-ok", Note: "本窗口没有失败的运行。"})
		}
	}

	// 3) 最近 error/fatal 事件：逐条给可点证据（最多 5 条）。visibility 过滤（BUG-010，
	// 规则 18）：default-hidden 类型不进候选集（tier=infra 的恢复/节拍事实不淹没注意力）；
	// attention 族（HeartbeatLost/SessionStalled/Escalate）生效词形 default-visible 照常上屏。
	if o.eventLog != nil {
		var evs []opslog.EventLog
		for _, lv := range []string{"error", "fatal"} {
			if rows, err := o.eventLog.QueryEvents(opslog.EventQuery{Level: lv, Limit: 50, HiddenTypes: hidden}); err == nil {
				evs = append(evs, rows...)
			}
		}
		facts["errors"] = len(evs)
		sort.SliceStable(evs, func(i, j int) bool { return evs[i].Ts.After(evs[j].Ts) })
		for i, e := range evs {
			if i >= 5 {
				break
			}
			title := "异常事件：" + e.Type
			why := "等级 " + e.Severity + "，涉及 " + strings.TrimSpace(e.Involved)
			if strings.TrimSpace(e.Involved) == "" {
				why = "等级 " + e.Severity + "，没有关联对象。"
			}
			attention = append(attention, attentionItem{
				CandidateID: "c-event-" + e.EventID,
				Category:    "anomaly",
				Severity:    "high",
				Title:       title,
				Why:         why,
				Evidence: []map[string]any{{
					"kind": "event", "id": e.EventID, "ref": "#/event/" + e.EventID,
					"at": e.Ts.UTC().Format(time.RFC3339),
				}},
				Action: map[string]string{"kind": "open", "target": "#/event/" + e.EventID},
			})
		}
		if len(evs) == 0 {
			clear = append(clear, attentionClear{CandidateID: "c-events-ok", Note: "最近没有 error/fatal 事件。"})
		}
	}

	// 4) 待审批/待处理：只计挂起记录（已决不算）。
	if o.inbox != nil {
		st := o.inbox.State()
		pending := 0
		for _, r := range st.Records {
			if r.State == flowbind.ApprovalPending {
				pending++
			}
		}
		facts["approvalsPending"] = pending
		if pending > 0 {
			attention = append(attention, attentionItem{
				CandidateID: "c-inbox-pending",
				Category:    "proposal",
				Severity:    "high",
				Title:       fmt.Sprintf("%d 项等你处理", pending),
				Why:         "提案/派单/待审需要人判定；不处理会阻塞后续。",
				Evidence:    []map[string]any{{"kind": "inbox", "ref": "#/inbox"}},
				Action:      map[string]string{"kind": "approve", "target": "#/approvals"},
			})
		} else {
			clear = append(clear, attentionClear{CandidateID: "c-inbox-ok", Note: "没有待你处理的项。"})
		}
	}

	// 排序：critical→high→medium；同级保持候选顺序。attention 上限 7（只截 medium 之后）。
	sort.SliceStable(attention, func(i, j int) bool {
		return severityRank[attention[i].Severity] < severityRank[attention[j].Severity]
	})
	if len(attention) > 7 {
		attention = attention[:7]
	}
	if len(clear) > 5 {
		clear = clear[:5]
	}
	summary := "本窗口未发现需要关注的事项。"
	if len(attention) > 0 {
		summary = fmt.Sprintf("本窗口有 %d 条需要关注；%d 项待你处理。", len(attention), facts["approvalsPending"])
	}
	degraded := len(missing) > 0
	return map[string]any{
		"asOf":        now.Format(time.RFC3339),
		"window":      map[string]any{"from": now.Add(-24 * time.Hour).Format(time.RFC3339), "to": now.Format(time.RFC3339)},
		"scope":       "global",
		"generatedBy": "rules",
		"dataQuality": map[string]any{"degraded": degraded, "missing": missing, "stale": false},
		"facts":       facts,
		"summary":     summary,
		"attention":   attention,
		"clear":       clear,
	}
}

func docEvidence(ctx context.Context, objects ObjectSource, kind, scope, name string) map[string]any {
	ev := map[string]any{"kind": "doc", "ref": kind + "/" + scope + "/" + name}
	if objects != nil {
		if obj, err := objects.Get(ctx, kind, scope, name); err == nil {
			sum := sha256.Sum256(obj.Content)
			ev["hash"] = hex.EncodeToString(sum[:])
		}
	}
	return ev
}

func instancePhase(ctx context.Context, objects ObjectSource, scope, name string) string {
	if objects == nil {
		return ""
	}
	obj, err := objects.Get(ctx, "ServiceInstance", scope, name)
	if err != nil {
		return ""
	}
	var doc struct {
		Status struct {
			Phase string `yaml:"phase"`
		} `yaml:"status"`
	}
	if err := yaml.Unmarshal(obj.Content, &doc); err != nil {
		return ""
	}
	return strings.TrimSpace(doc.Status.Phase)
}
