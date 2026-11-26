// Package apiserver cored 查询面：scheme 加载校验结果 + 对象类型清单 get/list
// + 文档对象读面（/v1/objects，读库内真实文档对象而非类型注册表）。
// 设计依据：core_system_instance/runtime/cored-bridge.md（plugin↔cored 桥契约）。
// 缺省只读；写路径唯一入口是 Apply 管线（core_design/concepts/apply.md）——仅当部署方显式开启
// --allow-apply 时暴露 POST /v1/apply，内部收敛到 apply.Engine 三步闸门，不旁路存储。
package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/apply"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/docrepo"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/draft"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventpub"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/flowbind"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/metricsinfo"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/proposal"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/sessionliveness"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
	"github.com/friddle-self-recursive-agent/core_system_workservice/secretstore"
)

// ObjectSource 文档对象读取源（Repository 的读半边；接口化便于测试注入假仓库）。
type ObjectSource interface {
	Get(ctx context.Context, kind, scope, name string) (*repository.Object, error)
	List(ctx context.Context, kind, scope string) ([]*repository.Object, error)
}

// Applier 写路径（apply.Engine 的口面；接口化便于测试注入假引擎）。
type Applier interface {
	Apply(ctx context.Context, o *repository.Object, opt apply.Options) (*apply.Result, error)
	// ApplyDraft 稿→面落库（cored-bridge.md §2.2：POST /v1/apply + fromDraft=true 的引擎侧）。
	ApplyDraft(ctx context.Context, kind, scope, name string, opt apply.Options) (*apply.Result, error)
}

// DraftStaging 草稿缓存口面（draft.Store 满足；接口化便于测试注入假缓存）。
type DraftStaging interface {
	Get(kind, scope, name string) (*draft.Draft, error)
	List() []*draft.Draft
	Put(d *draft.Draft) error
	Delete(kind, scope, name string) error
}

// TriggerFace 触发骨架查询口面（workflow.TriggerHub 满足；接口化便于测试注入假面）。
// dispatch 接线第一段（cored-bridge.md §2.3）：apply 落面→DocApplied 事件→订阅收到留痕；
// v0.13 增 Resolve 派发决策查询（引擎主体第一段：命中行→owner 解析→派发目标投影）；
// v0.14 决策增执行绑定投影（引擎主体第二段：route/cc 目标→执行面 executor/target/wired）。
type TriggerFace interface {
	State() workflow.TriggerState
	Resolve(typ, involved string) workflow.DispatchDecision
}

// RunFace 对账 run 投影查询口面（workflow.RunLog 满足；引擎主体第三段 v0.15——真实拉起
// 第一子段：workflow 触发面接线，run=触发面→workflow 域的执行留痕起点投影）。
type RunFace interface {
	State() workflow.RunState
}

// DeliveryFace 送达投影查询口面（workflow.DeliveryLog 满足；引擎主体第四段 v0.16——真实
// 拉起第二子段：notification 送达面接线，delivery=notification 族绑定→送达域的受理投影；
// v0.17 收尾——出站注入后 status 翻转 sent/failed 与出站留痕随 State 一并出）。
type DeliveryFace interface {
	State() workflow.DeliveryState
}

// WakeFace 唤醒投影查询口面（workflow.WakeLog 满足；引擎主体第五段 v0.18——真实拉起
// 第三子段：agent 触发面接线，wake=`<p>-agent` 族绑定→agent 域的受理投影）。
type WakeFace interface {
	State() workflow.WakeState
}

// DispatchFace 派单受理投影查询口面（flowbind.DispatchLog 满足；v0.50 R-8——workflow-runtime
// §11.2 拉起形态首段：骨架物化即派单，dispatch=骨架→apply-workflow agent 派单接口面的受理
// 投影；受理≠拉起——真实拉起 declared-deferred）。
type DispatchFace interface {
	State() flowbind.DispatchState
}

// PublisherFace 事件发布面口面（eventpub.Publisher 满足；v0.20 多类型事件注入面首段：
// 类型注册校验+发布权限域判定+level 缺省投影+总线发布——注入走同一 trigger 主干，
// concepts/event.md 规则 8）。
type PublisherFace interface {
	Publish(req eventpub.Request, sub *permission.Subject) (*eventbus.Event, error)
}

// EventLogFace 事件归档查询口面（opslog.SQLiteStore 满足；v0.36 console 读面首段——
// event_log 归档行的 HTTP 查询投影 `GET /v1/events`：Welcome 最近重要事件（console-ui.md
// 屏幕 0 数据来源=event_log）与事件流页（E6）的持久源。只读投影：归档行不删不改）。
type EventLogFace interface {
	QueryEvents(q opslog.EventQuery) ([]opslog.EventLog, error)
	// GroupEvents (type,involved) 分组计数（v0.37 console 读面第二段——聚合计数源底层读取；
	// 具体方法不占冻结 Store 签名，QueryEvents 同款口面约定）。
	GroupEvents(q opslog.EventQuery) ([]opslog.EventGroup, int64, bool, error)
}

// ApplyLogFace 应用史查询口面（v0.38 console 读面第三段——文档编辑器历史 tab 的权威数据源；
// opslog.SQLiteStore 满足。具体方法不占冻结 Store 签名，EventLogFace 同款口面约定）。
type ApplyLogFace interface {
	QueryApplies(obj string, limit int) ([]opslog.ApplyLog, error)
}

// ApprovalLogFace 审批判定账面口面（v0.46 审批独立账——ops-log.md §2 v0.46 注记；
// opslog.SQLiteStore 满足。写半边=批准/驳回判定落账（apiserver 写钩子），读半边=
// GET /v1/approvals 查询投影；账面非权威，写路径结果不因落账失败翻转）。
type ApprovalLogFace interface {
	AppendApproval(l opslog.ApprovalLog) error
	QueryApprovals(q opslog.ApprovalQuery) ([]opslog.ApprovalLog, error)
}

// RevContentFace 历史 rv 正文读面口面（v0.63 rev-content diff 面——历史 tab「点行 diff 上一版」
// 余量收口的读半边；git 后端实现=repository.GitRepo）。语义=按对象键与目标 rv 取**该提交时点**
// 的权威正文快照（git 后端权威=仓库文件历史，ops DB 只记 rv 流水——concepts/revision.md §3
// 「git=内容版本权威」）；face 缺席=503 防御面（旧代兜底 405=代际判别器恒成立）。具体方法
// 不占冻结 Repository 签名（QueryApplies 口面同款约定，compile 期断言在 New 尾注记）。
type RevContentFace interface {
	// RevContent 取对象键 (kind, scope, name) 在 rv 提交时点的正文快照。找不到该 rv
	// 的落库提交 → ErrRevNotFound（HTTP 404，与对象缺席分立）；读面零副作用。
	RevContent(ctx context.Context, kind, scope, name string, rv int64) (string, error)
}

// WithRevContent 注入历史 rv 正文读面（v0.63 rev-content diff 面——历史 tab 点行 diff 上一版
// 读半边；git 后端恒装配；sqlite 后端=面缺席 503 诚实降级）。
func WithRevContent(f RevContentFace) Option {
	return func(o *options) { o.revContent = f }
}

// MetricsFace 数据面查询口面（metricsinfo.Store 满足；v0.33 MetricsInfo 数据面首段——
// store/queries 投影读面：指标清单与窗口聚合投影。可选能力：未注入=数据面未装配，
// /v1/metrics 返回空投影——关断口径缺省可见，与 405 代际判别器分立）。
type MetricsFace interface {
	Metrics() []string
	Count(metric string) int
	Query(q metricsinfo.Query) (float64, bool)
}

// SecretFace 部署侧秘密存储口面（secretstore.Store 满足；v0.34 secret 存储落地首段——
// 用户 2026-11-23 指示：凭据脱敏默认走 Secret，运行时态不导出）。脱敏口径：本面一切
// 投影只出名称与更新时刻，**值只进不出**（Set 受理即吞值，任何 GET/列表不回显）。
type SecretFace interface {
	List() ([]secretstore.SecretInfo, error)
	Set(name, value string) error
	Delete(name string) error
}

// Option 查询面构建选项（既有九参构建签名不变——增量面走可变选项装配，测试与装配层
// 既有调用零迁移）。
type Option func(*options)

type options struct {
	metrics         MetricsFace
	secrets         SecretFace
	eventLog        EventLogFace
	applyLog        ApplyLogFace
	approval        ApprovalLogFace
	revContent      RevContentFace
	dispatch        DispatchFace
	inbox           InboxFace
	designDocs      DesignDocs
	sessionLiveness SessionLivenessFace
}

// ProposalsSource 提案枚举源（v0.56 /v1/inbox 组装输入——读面走 ObjectSource.List 现读 +
// proposal.Parse 投影，零新查询；本接口型已并入 ObjectSource 判别路径，不再单列）。

// WithMetrics 注入数据面查询口面（v0.33 MetricsInfo 数据面首段）。
func WithMetrics(m MetricsFace) Option {
	return func(o *options) { o.metrics = m }
}

// WithSecrets 注入部署侧秘密存储口面（v0.34 secret 存储落地首段；--secrets-db 装配）。
func WithSecrets(s SecretFace) Option {
	return func(o *options) { o.secrets = s }
}

// WithEventLog 注入事件归档查询口面（v0.36 console 读面首段 `GET /v1/events`）。
func WithEventLog(f EventLogFace) Option {
	return func(o *options) { o.eventLog = f }
}

// WithApplyLog 注入应用史查询口面（v0.38 console 读面第三段 `GET /v1/history/{kind}/{name}`）。
func WithApplyLog(f ApplyLogFace) Option {
	return func(o *options) { o.applyLog = f }
}

// WithApprovalLog 注入审批判定账面口面（v0.46 审批独立账——写钩子 granted/rejected +
// `GET /v1/approvals` 读面；ops-log.md §2 v0.46 注记）。
func WithApprovalLog(f ApprovalLogFace) Option {
	return func(o *options) { o.approval = f }
}

// WithDispatchFace 注入派单受理投影口面（v0.50 R-8——`GET /v1/agents/dispatches` 读面；
// workflow-runtime.md §11.2 拉起形态首段）。
func WithDispatchFace(f DispatchFace) Option {
	return func(o *options) { o.dispatch = f }
}

// InboxFace 通知中心收件箱同步口面（v0.56 R-10——`GET /v1/inbox` 读面投影组装输入；
// flowbind.SyncApprovals 满足。State=批准投影快照；ApproveIdle=批准投影是否无挂起待决
// （watchdog 投影真实词形——零挂起≠停摆不冒充）；Sync=apply_log 判定行→批准投影翻转
// （workflow:approve|workflow:reject 行原地翻转 approved|rejected））。
type InboxFace interface {
	State() flowbind.ApprovalState
	ApproveIdle() bool
	Sync(rows []opslog.ApplyLog) int
}

// WithInboxFace 注入通知中心收件箱口面（v0.56；nil=面缺席 503 同款防御面）。
func WithInboxFace(f InboxFace) Option {
	return func(o *options) { o.inbox = f }
}

// WithDesignDocs 注入设计文档面（v0.52 能力 A——`/v1/design-docs*` 读/史/草稿/apply 面；
// docrepo.Set 满足。装配=--design-repo；写面随 --allow-apply 同开关）。
func WithDesignDocs(d DesignDocs) Option {
	return func(o *options) { o.designDocs = d }
}

// compile 期校验：git 后端满足读半边，apply 引擎满足写口面，draft.Store 满足草稿口面，
// workflow.TriggerHub 满足触发骨架口面，eventpub.Publisher 满足发布面口面。
var (
	_ ObjectSource        = (*repository.GitRepo)(nil)
	_ Applier             = (*apply.Engine)(nil)
	_ DraftStaging        = (*draft.Store)(nil)
	_ TriggerFace         = (*workflow.TriggerHub)(nil)
	_ RunFace             = (*workflow.RunLog)(nil)
	_ DeliveryFace        = (*workflow.DeliveryLog)(nil)
	_ WakeFace            = (*workflow.WakeLog)(nil)
	_ PublisherFace       = (*eventpub.Publisher)(nil)
	_ MetricsFace         = (*metricsinfo.Store)(nil)
	_ SecretFace          = (*secretstore.Store)(nil)
	_ EventLogFace        = (*opslog.SQLiteStore)(nil)
	_ ApplyLogFace        = (*opslog.SQLiteStore)(nil)
	_ ApprovalLogFace     = (*opslog.SQLiteStore)(nil)
	_ RevContentFace      = (*repository.GitRepo)(nil)
	_ DesignDocs          = (*docrepo.Set)(nil)
	_ SessionLivenessFace = (*sessionliveness.Watchdog)(nil)
)

// New 构建查询面 HTTP 处理器。端点：
//
//	GET /healthz                        存活探针
//	GET /v1/scheme                      scheme 概览（加载校验结果的投影）
//	GET /v1/scheme/kinds                对象类型清单（list）
//	GET /v1/scheme/kinds/{kind}         单个对象类型详情（get）
//	GET /v1/objects/{kind}?scope=s      文档对象清单（list；scope 缺省=跨 scope 枚举）
//	GET /v1/objects/{kind}/{name}?scope=s 单个文档对象（get；对象键=(kind,scope,name)）
//	GET /v1/events                       事件归档查询面（v0.36 console 读面首段：event_log
//	                                     最新优先投影；limit/type/level/involved/since）
//	GET /v1/history/{kind}/{name}?scope=s 对象应用史投影（v0.38 console 读面第三段：apply_log
//	                                     rv 降序；limit）
//	GET /v1/rev-content/{kind}/{name}?scope=s&rv=N 历史 rv 正文投影（v0.63 rev-content diff 面
//	                                     §1.40：git 权威正文快照读半边；face 缺席 503）
//	GET /v1/approvals                    审批判定账面查询面（v0.46 审批独立账：approval_log
//	                                     rv 降序；decision/obj/limit）
//	GET /v1/drafts                      草稿清单（瘦身；查询面恒开）
//	GET /v1/drafts/{kind}/{name}?scope=s 单条草稿（含正文；无稿 → 404）
//	POST /v1/drafts                     落稿（upsert；写面开关门控）
//	DELETE /v1/drafts/{kind}/{name}?scope=s 弃稿（写面开关门控）
//	GET /v1/triggers                    触发骨架查询面（dispatch 接线：表驱动订阅+留痕；恒开，§2.3）
//	GET /v1/triggers/resolve            派发决策查询面（v0.13 决策投影+v0.14 执行绑定投影；恒开，§2.3）
//	GET /v1/workflows/runs              对账 run 投影面（v0.15 真实拉起第一子段：workflow 触发面接线；恒开，§2.3）
//	GET /v1/notifications/deliveries    送达投影面（v0.16 送达面接线；v0.17 出站状态机 sent/failed；恒开，§2.3）
//	GET /v1/agents/wakes                唤醒投影面（v0.18 引擎主体第五段：agent 触发面接线；恒开，§2.3）
//	GET /v1/agents/dispatches           派单受理投影面（v0.50 R-8：骨架物化即派单——workflow-runtime §11.2；恒开）
//	GET /v1/metrics                     数据面指标清单投影（v0.33 MetricsInfo 数据面首段；空投影=关断口径可见）
//	GET /v1/metrics/query               数据面查询投影（窗口聚合最小集；未命中=404 不冒充零值）
//	GET /v1/secrets                     秘密元数据清单（v0.34；只出名称/更新时刻，值永不回显——脱敏红线）
//	POST /v1/secrets                    秘密落值（--allow-apply 门控；未开启=405；受理即吞值，回执零值）
//	DELETE /v1/secrets/{name}           秘密删除（--allow-apply 门控；幂等）
//	POST /v1/apply                      文档写入口（--allow-apply 显式开启；未开启=405；fromDraft=true 走草稿）
//	POST /v1/events                     事件注入面（v0.20 发布面；--allow-apply 门控；未开启=405；event.md 规则 8）
//
// objects 为 nil（cored 未配置 --docs）时 /v1/objects/* 返回 503；
// applier 为 nil（未开启 --allow-apply）时 POST /v1/apply 与草稿写操作返回 405（缺省只读面不变；
// 草稿 GET 属查询面恒开）。drafts 为 nil（缓存未接线）时 /v1/drafts* 返回 503。
// triggers 为 nil（触发骨架未接线）时 /v1/triggers 与 /v1/triggers/resolve 返回 503。
// runs 为 nil（run 投影面未接线）时 /v1/workflows/runs 返回 503。
// deliveries 为 nil（送达投影面未接线）时 /v1/notifications/deliveries 返回 503。
// wakes 为 nil（唤醒投影面未接线）时 /v1/agents/wakes 返回 503。
// dispatch 为 nil（派单受理投影面未接线——headless 载体/旧装配）时 /v1/agents/dispatches 返回 503。
// publisher 为 nil（写面关断）时 POST /v1/events 返回 405（与 apply 写面同门控同口径）。
// eventLog 为 nil（事件归档查询口面未接线）时 GET /v1/events 返回 503（面注册但能力缺席；
// 装配层随 ops DB 恒注入，此为防御面——旧代判别器仍是 405 兜底）。
// applyLog 为 nil（应用史查询口面未接线）时 GET /v1/history/* 返回 503（同上口径）。
// approval 为 nil（审批判定账面口面未接线）时 GET /v1/approvals 返回 503 且批准/驳回写钩子
// 零落账（账面缺席=无判定行，写路径语义不变——账面非权威，可重建）。
// metrics 未注入（数据面未装配）时 /v1/metrics 返回空投影（关断口径缺省可见——与 405
// 代际判别器分立：本面恒注册，装配与否只见投影丰瘠）。
func New(sch *scheme.Scheme, objects ObjectSource, applier Applier, drafts DraftStaging, triggers TriggerFace, runs RunFace, deliveries DeliveryFace, wakes WakeFace, publisher PublisherFace, opts ...Option) http.Handler {
	o := &options{}
	for _, f := range opts {
		f(o)
	}
	metrics := o.metrics
	secrets := o.secrets
	eventLog := o.eventLog
	applyLog := o.applyLog
	approval := o.approval
	revContent := o.revContent
	dispatch := o.dispatch
	designDocs := o.designDocs
	// logApproval 账面非权威落账：失败只注记 stderr，不翻转写路径结果（ops-log.md §2 v0.46）。
	logApproval := func(l opslog.ApprovalLog) {
		if approval == nil {
			return
		}
		if err := approval.AppendApproval(l); err != nil {
			log.Printf("apiserver: approval_log 落账失败（非权威，写路径不受影响）: %v", err)
		}
	}
	mux := http.NewServeMux()
	mux.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"ok": true})
	})
	mux.HandleFunc("GET /v1/scheme", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, http.StatusOK, sch.Summary())
	})
	mux.HandleFunc("GET /v1/scheme/kinds", func(w http.ResponseWriter, _ *http.Request) {
		kinds := sch.ListKinds()
		writeJSON(w, http.StatusOK, map[string]any{"count": len(kinds), "kinds": kinds})
	})
	mux.HandleFunc("GET /v1/scheme/kinds/{kind}", func(w http.ResponseWriter, r *http.Request) {
		name := r.PathValue("kind")
		k, err := sch.Lookup(name)
		if err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, k)
	})
	mux.HandleFunc("GET /v1/objects/{kind}", func(w http.ResponseWriter, r *http.Request) {
		if objects == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "对象存储未配置（cored 缺 --docs）"})
			return
		}
		kind, scope := r.PathValue("kind"), r.URL.Query().Get("scope")
		if _, err := sch.Lookup(kind); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
			return
		}
		objs, err := objects.List(r.Context(), kind, scope)
		if err != nil {
			writeObjErr(w, err)
			return
		}
		items := make([]map[string]any, 0, len(objs))
		for _, o := range objs {
			items = append(items, objRef(o))
		}
		writeJSON(w, http.StatusOK, map[string]any{"count": len(items), "kind": kind, "scope": scope, "objects": items})
	})
	mux.HandleFunc("GET /v1/objects/{kind}/{name}", func(w http.ResponseWriter, r *http.Request) {
		if objects == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "对象存储未配置（cored 缺 --docs）"})
			return
		}
		kind, name := r.PathValue("kind"), r.PathValue("name")
		scope := r.URL.Query().Get("scope")
		if _, err := sch.Lookup(kind); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
			return
		}
		o, err := objects.Get(r.Context(), kind, scope, name)
		if err != nil {
			writeObjErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"kind": o.Kind, "name": o.Name, "scope": o.Scope,
			"resourceVersion": o.ResourceVersion, "content": string(o.Content),
		})
	})
	// 写入口（缺省关闭）：POST /v1/apply——收敛到 apply 管线三步闸门，不旁路存储。
	// 未开启（applier=nil）时恒 405，缺省只读面语义不变。
	// 主体（v0.9 请求级）：请求体 subject 字段（优先）或 X-Core-Subject 头（同形 JSON）；
	// 两者皆缺=匿名主体照判（无 role 落点 → 403，进程级本地信任不兜底 HTTP 匿名请求）。
	// fromDraft=true：稿→面落库（cored-bridge.md §2.2）——差异以草稿缓存为准，不携带 content。
	mux.HandleFunc("POST /v1/apply", func(w http.ResponseWriter, r *http.Request) {
		if applier == nil {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "apply 写路径未开启（cored 缺 --allow-apply；缺省只读）"})
			return
		}
		var req applyRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<20)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("请求体不可解析: %v", err)})
			return
		}
		opt := apply.Options{
			DryRun: req.DryRun, Force: req.Force,
			FieldManager: req.FieldManager, ReasonRef: req.ReasonRef,
			Subject: resolveSubject(r, &req),
		}
		var res *apply.Result
		if req.FromDraft {
			if req.Content != "" {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "fromDraft=true 不携带 content（差异以草稿缓存为准）"})
				return
			}
			if req.ResourceVersion != 0 {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "fromDraft=true 不携带 resourceVersion（乐观锁期望值以草稿 baseResourceVersion 为准）"})
				return
			}
			if req.Kind == "" || req.Scope == "" || req.Name == "" {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": "对象键不全（需 kind、scope、name）"})
				return
			}
			var err error
			res, err = applier.ApplyDraft(r.Context(), req.Kind, req.Scope, req.Name, opt)
			if err != nil {
				syncDeniedJudgmentRow(req, opt, err, o.inbox) // v0.57：denied 判定行随行同步（观测投影零旁路——账本照落）
				writeApplyErr(w, err)
				return
			}
		} else {
			obj := &repository.Object{
				Kind: req.Kind, Name: req.Name, Scope: req.Scope,
				ResourceVersion: req.ResourceVersion, Content: []byte(req.Content),
			}
			var err error
			res, err = applier.Apply(r.Context(), obj, opt)
			if err != nil {
				syncDeniedJudgmentRow(req, opt, err, o.inbox) // v0.57：denied 判定行随行同步（观测投影零旁路——账本照落）
				writeApplyErr(w, err)
				return
			}
		}
		resp := map[string]any{
			"resource": fmt.Sprintf("%s/%s/%s", req.Kind, req.Scope, req.Name),
			"rv":       res.RV, "action": res.Action, "dryRun": req.DryRun,
		}
		// 审批判定落账（v0.46 审批独立账，ops-log.md §2；v0.51 对称扩展）：console:approve 批准 /
		// console:reject 驳回判定——fromDraft ∧ 非 dry-run ∧ 判定成功即落对应行（dry-run=预检非判定；
		// 非 fromDraft=直接落库非审批屏动作；noop 幂等重入照落，rv=面上现值——判定发生了，零变更可落）。
		decision := ""
		switch req.FieldManager {
		case ApproveFieldManager, flowbind.ApprovalManager: // console:approve | workflow:approve（v0.56 派单批准动作面——§11.5）
			decision = opslog.DecisionGranted
		case RejectFieldManager, flowbind.ApprovalRejecter: // console:reject | workflow:reject
			decision = opslog.DecisionRejected
		}
		// 派单 ask 批准动作面（v0.56 R-10——workflow-runtime.md §11.5 批准投影翻转判定源）：
		// workflow:approve|workflow:reject 溯源串行=派单批准判定（Apply ok/noop 成功事实即翻转
		// 依据——flowbind.SyncApprovals.Sync 消费 apply_log 行集）。判定权威=apply_log 审计行；
		// 本翻转面=观测投影同步（失败不翻转写路径结果，下一拍派单查询面重扫自愈）。
		// v0.58 同拍判定行在账（workflow-runtime.md §11.5 动作面 console 动作化——判定行是
		// 账本事实，动作返回后立即读得到：活体判定窗「批准动作 → inbox 下一次装载反映」等价
		// 缩到「动作返回拍」——历拍重扫自愈路径保持不变作兜底）。判据语义锚不变（记录新增行）。
		if o.inbox != nil && decision != "" && !req.DryRun && res != nil && o.applyLog != nil {
			if rows, err := o.applyLog.QueryApplies(fmt.Sprintf("%s/%s/%s", req.Kind, req.Scope, req.Name), 5); err == nil {
				for _, row := range rows {
					if row.FieldManager == req.FieldManager {
						o.inbox.Sync([]opslog.ApplyLog{row}) // v0.58 判定行在账（含 denied 形账本行零旁路——匹配判定不发明）
						break
					}
				}
			}
		}
		if o.inbox != nil && decision != "" && !req.DryRun && res != nil {
			row := opslog.ApplyLog{
				Ts: time.Now(), FieldManager: req.FieldManager, Subject: func() string {
					if opt.Subject != nil {
						return opt.Subject.ID()
					}
					return "anonymous"
				}(), Obj: fmt.Sprintf("%s/%s/%s", req.Kind, req.Scope, req.Name),
				ResourceVersion: res.RV, Result: func() string {
					if res.Action == "noop" {
						return "noop"
					}
					return "ok"
				}(),
			}
			o.inbox.Sync([]opslog.ApplyLog{row}) // 同步翻转（观测投影——失败仅丢本轮，下一拍重扫自愈）
		}
		if approval != nil && req.FromDraft && !req.DryRun && decision != "" {
			judged := "anonymous"
			if opt.Subject != nil {
				judged = opt.Subject.ID()
			}
			objKey := fmt.Sprintf("%s/%s/%s", req.Kind, req.Scope, req.Name)
			logApproval(opslog.ApprovalLog{
				Ts: time.Now(), ApprovalID: "draft:" + objKey,
				Obj: objKey, Decision: decision,
				JudgedBy: judged, FieldManager: req.FieldManager,
				ResourceVersion: res.RV, ReasonRef: req.ReasonRef,
			})
			// 审批信号族事件随行（v0.51，event.md 规则 11/15——判定是人经 console 做出的，
			// 经发布面同管线 source=human；involved=判定对象——非 Workflow 引用不命中 run 域迁移，
			// dispatch 派发面对审批族无目标=dead-letter 行，零副作用；发布失败不翻转 apply 结果——
			// 事件是通知，判定权威=approval_log 行）。
			if publisher != nil && req.Kind == "Proposal" && opt.Subject != nil {
				typ := "ApprovalGranted"
				if decision == opslog.DecisionRejected {
					typ = "ApprovalRejected"
				}
				if et, err := sch.LookupEventType(typ); err == nil { // 注册表闸门前置
					if _, err := publisher.Publish(eventpub.Request{
						Type: typ, Involved: objKey, DedupKey: "approval:" + objKey, Level: et.DefaultLevel,
					}, opt.Subject); err != nil {
						log.Printf("apiserver: %s 事件发布未遂（非权威，判定不受影响）: %v", typ, err)
					}
				}
			}
		}
		if res.Event != nil {
			resp["event"] = map[string]any{"id": res.Event.ID, "type": res.Event.Type, "dedupKey": res.Event.DedupKey}
		}
		writeJSON(w, http.StatusOK, resp)
	})

	// 事件注入面（v0.20 发布面首段，concepts/event.md 规则 8；写面随 --allow-apply 门控，
	// 未开启恒 405 与 apply 写面同口径）：POST /v1/events——类型注册校验（Scheme.eventTypes，
	// 事件只校验类型注册）→ 主体发布权判定（publish × events 资源域，与 apply 写权限域分立）
	// → level 缺省=注册表 defaultLevel 投影 → 总线发布（与引擎内部事件同一条主干：窗口去重/
	// 订阅派发/event_log 归档零旁路）。注入通道 source 由服务侧判定恒 human，载荷声明的 source
	// 不采纳（来源标注是审计事实不是请求参数）；回执 {id,type,dedupKey,level,source,involved}
	// 与留痕 eventId 同源。旧构建（face 缺席）经兜底 handler 返回 405=代际判别器。
	mux.HandleFunc("POST /v1/events", func(w http.ResponseWriter, r *http.Request) {
		if publisher == nil {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "事件注入面未开启（cored 缺 --allow-apply；缺省只读）"})
			return
		}
		var req eventsRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("请求体不可解析: %v", err)})
			return
		}
		if req.Source != "" && req.Source != eventpub.SourceHuman {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("载荷声明的 source=%q 不采纳：来源标注由服务侧判定（注入通道恒 %s）", req.Source, eventpub.SourceHuman)})
			return
		}
		ev, err := publisher.Publish(eventpub.Request{
			Type: req.Type, Involved: req.Involved, Level: req.Level, DedupKey: req.DedupKey,
		}, subjectFrom(r, req.Subject))
		if err != nil {
			writePublishErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"id": ev.ID, "type": ev.Type, "dedupKey": ev.DedupKey,
			"level": ev.Level, "source": ev.Source, "involved": ev.Involved,
		})
	})

	// 秘密存储面（v0.34 secret 存储落地首段，provider.md 规则 2/6 运行时面；cored
	// --secrets-db 装配）：GET /v1/secrets 列表只出名称与更新时刻（**值永不回显**——脱敏
	// 红线）；POST /v1/secrets 落值（受理即吞值，回执零值）；DELETE /v1/secrets/{name} 幂等
	// 删除。写操作与 apply/events 写面同门控（--allow-apply；未开启恒 405=代际判别器）；
	// face 未注入（--secrets-db 未配置）时全部 503（面注册但能力缺席，关断口径可见）。
	mux.HandleFunc("GET /v1/secrets", func(w http.ResponseWriter, _ *http.Request) {
		if secrets == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "秘密存储未配置（cored 缺 --secrets-db）"})
			return
		}
		rows, err := secrets.List()
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		type row struct {
			Name      string    `json:"name"`
			UpdatedAt time.Time `json:"updatedAt"`
		}
		out := make([]row, 0, len(rows))
		for _, r0 := range rows {
			out = append(out, row{Name: r0.Name, UpdatedAt: r0.UpdatedAt})
		}
		writeJSON(w, http.StatusOK, map[string]any{"count": len(out), "secrets": out})
	})
	mux.HandleFunc("POST /v1/secrets", func(w http.ResponseWriter, r *http.Request) {
		if applier == nil {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "写路径未开启（cored 缺 --allow-apply；缺省只读）"})
			return
		}
		if secrets == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "秘密存储未配置（cored 缺 --secrets-db）"})
			return
		}
		var req struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		}
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 64<<10)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("请求体不可解析: %v", err)})
			return
		}
		if req.Name == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "name 必填"})
			return
		}
		if err := secrets.Set(req.Name, req.Value); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		// 回执零值（受理即吞值——请求体里的 value 不出现在任何响应/留痕）。
		writeJSON(w, http.StatusOK, map[string]any{"name": req.Name, "set": true})
	})
	mux.HandleFunc("DELETE /v1/secrets/{name}", func(w http.ResponseWriter, r *http.Request) {
		if applier == nil {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "写路径未开启（cored 缺 --allow-apply；缺省只读）"})
			return
		}
		if secrets == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "秘密存储未配置（cored 缺 --secrets-db）"})
			return
		}
		name := r.PathValue("name")
		if name == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "name 必填"})
			return
		}
		if err := secrets.Delete(name); err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"name": name, "deleted": true})
	})

	// 草稿缓存面（cored-bridge.md §2.2；apply.md §1 编辑与发布分离）：GET 属查询面恒开；
	// 写操作（落稿/弃稿）与 fromDraft apply 随写面开关（缺省 405）。
	mux.HandleFunc("GET /v1/drafts", func(w http.ResponseWriter, _ *http.Request) {
		if drafts == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "草稿缓存未配置"})
			return
		}
		items := make([]map[string]any, 0)
		for _, d := range drafts.List() {
			items = append(items, draftRef(d))
		}
		writeJSON(w, http.StatusOK, map[string]any{"count": len(items), "drafts": items})
	})
	mux.HandleFunc("GET /v1/drafts/{kind}/{name}", func(w http.ResponseWriter, r *http.Request) {
		if drafts == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "草稿缓存未配置"})
			return
		}
		kind, name := r.PathValue("kind"), r.PathValue("name")
		scope := r.URL.Query().Get("scope")
		d, err := drafts.Get(kind, scope, name)
		if err != nil {
			writeDraftErr(w, err)
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"kind": d.Kind, "name": d.Name, "scope": d.Scope,
			"fieldManager": d.FieldManager, "reasonRef": d.ReasonRef,
			"baseResourceVersion": d.BaseRV,
			"createdAt":           d.CreatedAt.UTC().Format(time.RFC3339Nano),
			"updatedAt":           d.UpdatedAt.UTC().Format(time.RFC3339Nano),
			"content":             string(d.Content),
		})
	})
	mux.HandleFunc("POST /v1/drafts", func(w http.ResponseWriter, r *http.Request) {
		if applier == nil {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "草稿写面未开启（cored 缺 --allow-apply；草稿读面恒开）"})
			return
		}
		if drafts == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "草稿缓存未配置"})
			return
		}
		var req draftRequest
		if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 4<<20)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("请求体不可解析: %v", err)})
			return
		}
		d := &draft.Draft{
			Kind: req.Kind, Scope: req.Scope, Name: req.Name,
			Content: []byte(req.Content), FieldManager: req.FieldManager,
			ReasonRef: req.ReasonRef, BaseRV: req.BaseResourceVersion,
		}
		if err := drafts.Put(d); err != nil {
			writeDraftErr(w, err) // ErrInvalid → 400（键不全/无主/空稿）
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"resource": fmt.Sprintf("%s/%s/%s", req.Kind, req.Scope, req.Name),
			"action":   "staged", "baseResourceVersion": req.BaseResourceVersion,
		})
	})
	mux.HandleFunc("DELETE /v1/drafts/{kind}/{name}", func(w http.ResponseWriter, r *http.Request) {
		if applier == nil {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "草稿写面未开启（cored 缺 --allow-apply；草稿读面恒开）"})
			return
		}
		if drafts == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "草稿缓存未配置"})
			return
		}
		kind, name := r.PathValue("kind"), r.PathValue("name")
		scope := r.URL.Query().Get("scope")
		if err := drafts.Delete(kind, scope, name); err != nil {
			writeDraftErr(w, err)
			return
		}
		// 审批判定落账（v0.46 审批独立账）：驳回=弃稿（零文档面动作）——判定发生了，账面落行
		// 才可审计（弃稿动作此前在 ops DB 零留痕）。judged_by=请求主体头（DELETE 无 body）；
		// 幂等 404（本就无稿）不落账——无判定发生。
		if approval != nil {
			logApproval(opslog.ApprovalLog{
				Ts: time.Now(), ApprovalID: "draft:" + fmt.Sprintf("%s/%s/%s", kind, scope, name),
				Obj: fmt.Sprintf("%s/%s/%s", kind, scope, name), Decision: opslog.DecisionRejected,
				JudgedBy: subjectFrom(r, nil).ID(), FieldManager: "draft:delete",
				ReasonRef: "draft-delete",
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"resource": fmt.Sprintf("%s/%s/%s", kind, scope, name), "action": "discarded",
		})
	})
	// 触发骨架面（cored-bridge.md §2.3；dispatch 接线：第二段 v0.12 订阅面动态化 →
	// 第三段 v0.13 引擎主体第一段 owner 解析激活）：GET 恒开——
	// 「apply 落面→事件→订阅收到留痕」主干第一段的机械可验查询面；
	// types=dispatch 表推导订阅集合（"*"=通配），records 携命中行投影 route/cc/for/owner
	// （全表未命中=dead-letter，死信兜底可见）。留痕是进程内观测面（账本=event_log 归档）；
	// 写面关断不影响本端点（received=0 照常 200）。
	mux.HandleFunc("GET /v1/triggers", func(w http.ResponseWriter, _ *http.Request) {
		if triggers == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "触发骨架未接线"})
			return
		}
		st := triggers.State()
		records := make([]map[string]any, 0, len(st.Records))
		for _, r := range st.Records {
			records = append(records, map[string]any{
				"seq": r.Seq, "eventId": r.EventID, "type": r.Type,
				"dedupKey": r.DedupKey, "source": r.Source, "involved": r.Involved,
				"level": r.Level, "route": r.Route, "cc": r.CC, "for": r.For, "owner": r.Owner,
				"execution":  map[string]any{"route": r.Exec, "cc": r.CCExec},
				"at":         r.At.UTC().Format(time.RFC3339Nano),
				"receivedAt": r.ReceivedAt.UTC().Format(time.RFC3339Nano),
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"count": len(records), "received": st.Received, "dropped": st.Dropped,
			"types": st.Types, "records": records,
		})
	})
	// 派发决策查询面（§2.3 v0.13 引擎主体第一段 → v0.14 第二段增执行投影）：对当前 dispatch
	// 表快照回答「这一事件若到达会路由到哪、由谁执行」——纯决策（零事件注入、零留痕、零派发），
	// 投影 {type, involved, matched, route, cc, for, owner, row, execution}；row=命中表行号（0 起），
	// 全表未命中 row=-1 且 route=dead-letter（合成死信，execution.route.executor=none）。
	// owner:<字面量> 谓词的 live 机械探针落点（无事件注入面也能验证 owner 解析命中）；
	// execution={route:{executor,target,wired}, cc:{…}}=route/cc 目标执行面绑定投影
	// （映射表家族收编，未知目标=unbound 缺口可见；wired=执行面接线态，装配层登记前恒 false）。
	mux.HandleFunc("GET /v1/triggers/resolve", func(w http.ResponseWriter, r *http.Request) {
		if triggers == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "触发骨架未接线"})
			return
		}
		typ := r.URL.Query().Get("type")
		if typ == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "缺 type 查询参数（决策必须有事件类型）"})
			return
		}
		involved := r.URL.Query().Get("involved") // 可缺省（解析为空 owner，决策照常给出）
		d := triggers.Resolve(typ, involved)
		writeJSON(w, http.StatusOK, map[string]any{
			"type": typ, "involved": involved, "matched": d.Matched,
			"route": d.Route, "cc": d.CC, "for": d.For, "owner": d.Owner, "row": d.Row,
			"execution": map[string]any{"route": d.Exec, "cc": d.CCExec},
		})
	})
	// 对账 run 投影面（§2.3 v0.15 引擎主体第三段——真实拉起第一子段：workflow 触发面接线；
	// v0.23 相位推进器首段：受理即启动——引擎单步迁移判定随受理同拍完成，builtin run
	// phase=声明第二相位∧status=running；v0.25 显式迁移表首段：载荷新增 transitions=物化
	// 声明表行投影（新载荷键=代际判别器）；v0.26 异步推进首段：载荷新增 advances=累计异步
	// 推进计数（新载荷键=代际判别器）；v0.28 对象源相位投影与判定接线：object 源 run 相位/
	// 状态随对象自声明 spec 参与同一表驱动判定——未声明 phases=零投影口径不变）。run 出现 ≠
	// 对账执行完毕口径不变。旧构建（face 缺席）经兜底 handler 返回 405=代际判别器
	//（与 loop-13 固化口径一致）。
	mux.HandleFunc("GET /v1/workflows/runs", func(w http.ResponseWriter, _ *http.Request) {
		if runs == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "run 投影面未接线"})
			return
		}
		st := runs.State()
		items := make([]map[string]any, 0, len(st.Runs))
		for _, r := range st.Runs {
			items = append(items, map[string]any{
				"run": r.Run, "workflow": r.Workflow, "specSource": r.SpecSource,
				"phases": r.Phases, "transitions": r.Transitions, "phase": r.Phase, "status": r.Status,
				"triggerSeq": r.TriggerSeq, "eventId": r.EventID, "type": r.Type,
				"dedupKey": r.DedupKey, "route": r.Route, "owner": r.Owner,
				"target": r.Target, "startedAt": r.StartedAt.UTC().Format(time.RFC3339Nano),
				// v0.37 console 读面第二段：来源留痕 involved 投影（run 涉及对象链第一跳——
				// 运行抽屉「涉及文档」取本字段落对象链接，无前端拼凑；新载荷键=代际判别器）。
				"involved": r.Involved,
				// v0.32 刹车首段：生效刹车值投影（声明→留痕面——spec.brakes 钳定后的每锚生效值；
				// 新载荷键=代际判别器）。
				"brakes": map[string]any{
					"cooldown":         r.Brakes.Cooldown.String(),
					"maxMutationsTick": r.Brakes.MaxMutationsTick,
					"source":           r.Brakes.Source,
				},
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"count": len(items), "started": st.Started, "dropped": st.Dropped,
			"advances": st.Advances, // v0.26 异步推进首段：累计推进 run·步计数（新载荷键=代际判别器）
			"braked":   st.Braked,   // v0.32 刹车首段：累计冷却门拦阻触发数（新载荷键=代际判别器）
			"capped":   st.Capped,   // v0.32 刹车首段：累计推进拍上限拦阻 run·步数（新载荷键=代际判别器）
			"disabled": st.Disabled, // v0.58b enabled honoring：累计 enabled=false 锚拦阻触发数（新载荷键=代际判别器）
			"runs":     items,
		})
	})
	// 送达投影面（§2.3 v0.16 引擎主体第四段——真实拉起第二子段：notification 送达面接线；
	// v0.17 收尾：受理后的真实通道出站挂接——status 翻转 accepted→sent/failed + 出站留痕）：
	// 执行面=notification 且装配层已接线的绑定（route 主路与 cc 旁路，origin 维度留痕）推进为
	// 一条送达投影（送达已受理+通道投影；出站注入后异步翻转终态，未注入恒 accepted）。
	// 旧构建（face 缺席）经兜底 handler 返回 405=代际判别器（与 runs 面同口径）；
	// 新代新增载荷键 sent/failed/attempts/completedAt（恒在）与 error（有错才在）——
	// 键的出现/缺席=代际判别器（loop-16 口径）。
	mux.HandleFunc("GET /v1/notifications/deliveries", func(w http.ResponseWriter, _ *http.Request) {
		if deliveries == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "送达投影面未接线"})
			return
		}
		st := deliveries.State()
		items := make([]map[string]any, 0, len(st.Deliveries))
		for _, d := range st.Deliveries {
			completed := ""
			if !d.CompletedAt.IsZero() {
				completed = d.CompletedAt.UTC().Format(time.RFC3339Nano)
			}
			item := map[string]any{
				"delivery": d.Delivery, "channel": d.Channel, "origin": d.Origin,
				"status": d.Status, "triggerSeq": d.TriggerSeq, "eventId": d.EventID,
				"type": d.Type, "dedupKey": d.DedupKey, "level": d.Level,
				"route": d.Route, "owner": d.Owner, "at": d.At.UTC().Format(time.RFC3339Nano),
				"attempts": d.Attempts, "completedAt": completed,
			}
			if d.Error != "" {
				item["error"] = d.Error
			}
			items = append(items, item)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"count": len(items), "accepted": st.Accepted, "sent": st.Sent, "failed": st.Failed,
			"dropped": st.Dropped,
			// v0.35 通知门链三计数（KI-001——新载荷键=代际判别器，键恒在口径）：
			// gated=阶段门拦阻（非相位终态事实）；filtered=minLevel 等级过滤；deduped=窗口去重聚合。
			"gated": st.Gated, "filtered": st.Filtered, "deduped": st.Deduped,
			"deliveries": items,
		})
	})
	// 唤醒投影面（§2.3 v0.18 引擎主体第五段——真实拉起第三子段：agent 触发面接线）：
	// 执行面=agent（`<p>-agent` 族：root-agent/parent-agent）且装配层已接线的绑定（route
	// 主路与 cc 旁路，origin 维度留痕）推进为一条唤醒投影（唤醒已受理+agent 锚投影；
	// 唤醒投影出现 ≠ agent 会话已拉起——诚实投影与 run/送达面同口径）。Escalate→root-agent
	// 行的机械观察点。旧构建（face 缺席）经兜底 handler 返回 405=代际判别器（runs/送达面
	// 同口径）。
	mux.HandleFunc("GET /v1/agents/wakes", func(w http.ResponseWriter, _ *http.Request) {
		if wakes == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "唤醒投影面未接线"})
			return
		}
		st := wakes.State()
		items := make([]map[string]any, 0, len(st.Wakes))
		for _, wk := range st.Wakes {
			items = append(items, map[string]any{
				"wake": wk.Wake, "agent": wk.Agent, "origin": wk.Origin,
				"status": wk.Status, "level": wk.Level, "triggerSeq": wk.TriggerSeq,
				"eventId": wk.EventID, "type": wk.Type, "dedupKey": wk.DedupKey,
				"route": wk.Route, "owner": wk.Owner, "at": wk.At.UTC().Format(time.RFC3339Nano),
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"count": len(items), "accepted": st.Accepted, "dropped": st.Dropped, "wakes": items,
		})
	})
	// 派单受理投影面（v0.50 R-8——workflow-runtime.md §11.2 拉起形态首段；v0.51 扩展=第二触发源
	// Proposal 批准派单）：骨架物化动作（created|updated——「该版本实现未落」的机械签名，kind=workflow）
	// 与 Proposal 批准落库（action=approved，kind=proposal——concepts/proposal.md 规则 2）→
	// apply-workflow 锚派单受理投影；noop/skipped 不派单（无新实现可做）。载荷 {count, accepted,
	// dropped, dispatches}，行含 kind/agent/status/autoLevel/action/锚/rv/eventId（kind=workflow 行含
	// workflow/connector；kind=proposal 行含 proposal/title/summary/acceptance/priority/proposer——
	// 键恒在投影，空值=不适用）。**派单受理 ≠ agent 会话已拉起**（诚实投影与 wake 面同口径：真实拉起=
	// dsh 会话挂载/agent-session 域，declared-deferred）；进程内环（观测面非账本——持久态=骨架对象
	// execution.skeleton=true+apply_log 审计+Proposal 状态机，重启归零=幂等重入不重复派单）。
	// face 缺席=503（防御面，装配层 --allow-apply 随 flowbind/proposal 注入）；旧构建（路由未注册）
	// 经兜底 handler 405=代际判别器。
	mux.HandleFunc("GET /v1/agents/dispatches", func(w http.ResponseWriter, _ *http.Request) {
		if dispatch == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "派单受理投影面未接线（装配缺 flowbind 派单 sink）"})
			return
		}
		st := dispatch.State()
		items := make([]map[string]any, 0, len(st.Records))
		// v0.56 ask 批准动作面（workflow-runtime.md §11.5——拉起投影面第三段）：派单行随行
		// approval 键（state 词=批准投影翻转词——人批/驳回=Apply 判定；审批投影面缺席=零值不冒充）。
		approvalIdx := map[string]string{}
		spawnIdx := map[string]flowbind.SpawnDeclaration{}
		if o.inbox != nil {
			for _, r := range o.inbox.State().Records {
				approvalIdx[r.ApprovalID] = r.State
				if d := flowbind.SpawnOf(r); d.Declared {
					spawnIdx[r.ApprovalID] = d // v0.57 spawn 接线声明投影（approved 行——wired 恒 false）
				}
			}
		}
		for _, d := range st.Records {
			kind := d.Kind
			if kind == "" {
				kind = flowbind.DispatchKindWorkflow // v0.50 形记录投影补 kind（旧 sink 兼容）
			}
			items = append(items, map[string]any{
				"seq": d.Seq, "agent": d.Agent, "status": d.Status, "autoLevel": d.AutoLevel,
				"action": d.Action, "kind": kind,
				"workflow": d.Workflow, "scope": d.Scope, "connector": d.Connector,
				// v0.51 proposal 派单载荷（键恒在=代际判别器；kind=workflow 行空值=不适用）。
				"proposal": d.Proposal, "title": d.Title, "summary": d.Summary,
				"acceptance": d.Acceptance, "priority": d.Priority, "proposer": d.Proposer,
				"rv": d.RV, "eventId": d.EventID,
				// v0.55 节点接线载荷（拉起投影面首件，workflow-runtime.md §11.2 第三段——
				// apply-workflow 拉起输入=骨架节点清单；kind=proposal 行空值=不适用）。
				"nodes": d.Nodes,
				// v0.56 ask 批准动作面（拉起投影面第三段，workflow-runtime.md §11.5）：
				// state=pending|approved|rejected（approved≠spawn 已拉起——真实执行仍 observe）。
				"approval": map[string]any{"state": approvalIdx[matchingKeyOfRecord(d)], "approvalId": matchingKeyOfRecord(d)},
				// v0.57 spawn 接线声明投影（拉起投影面末段，workflow-runtime.md §11.7——零值键
				// =代际判别器）：approved 批准投影行 → {declared:true, wired:false, target, …}
				// 接线声明词形（approved≠拉起——真实拉起=agent-session 域 deferred，红线 a）。
				"spawn": spawnPayload(spawnIdx[matchingKeyOfRecord(d)]),
				"at":    d.At.UTC().Format(time.RFC3339Nano),
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"count": len(items), "accepted": st.Accepted, "dropped": st.Dropped, "dispatches": items,
		})
	})
	// 关注报告（R-UX-6 / concepts/attention.md；2026-11-26）：读侧派生，确定性候选+facts+
	// 证据链；v1 规则版（generatedBy=rules，LLM 叙述/缓存/权限指纹后续段）。数据源=objects+
	// runs+eventLog+inbox；face 缺席=degraded 报告（不 405，不空白）。只读 GET。
	mux.HandleFunc("GET /v1/attention", func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, attentionReport(r.Context(), objects, runs, o))
	})
	// 通知中心收件箱投影面（v0.56 R-10——console-adoption.md §1.25「待你处理收件箱」页面兜底
	// 首段 + notification.md 规则 6「未配置通知通道=仅页面展示」的机读落点）。四类判定分类
	// （机械判定=读面现读自分类——console 不自建持久状态红线 2）：NeedsHuman（归档 NeedsHuman
	// ∧Escalate 行——acceptance.md §9 接管链）；ApprovalNeeded（归档行——审批提升人注意终态族）
	// ∧批准投影挂起行（v0.56 派单 ask 批准动作面——flowbind.SyncApprovals 同一投影）；Proposal
	// 待批（status=submitted 提案——capability B 状态机词汇）；DocReview（草稿缓存行——代改代审
	// 的家，apply.md §6）。载荷 {count, classes, items, watchdog}；行含 class/kind/name/scope/
	// digest/eventId/at/link（link=console 深链路由——规则 6 页面兜底入口恒达操作面）。
	// face 缺席=503（防御面）；旧构建 405=代际判别器。watchdog 投影={approvals, pending,
	// approveIdle}——真实词形（approveIdle=true=批准投影无挂起待决，非停摆）。
	mux.HandleFunc("GET /v1/inbox", func(w http.ResponseWriter, _ *http.Request) {
		if o.inbox == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "通知中心收件箱面未接线（装配缺 flowbind 批准投影 sink）"})
			return
		}
		var evs []opslog.EventLog
		if o.eventLog != nil {
			// 三类人注意事件各拉一窗（type 精确过滤——归档窗 DocApplied 刷屏不挤占人注意行；
			// 半开窗语义=QueryEvents 既有 limit 口径）。
			for _, typ := range []string{"NeedsHuman", "Escalate", "ApprovalNeeded"} {
				if rows, err := o.eventLog.QueryEvents(opslog.EventQuery{Type: typ, Limit: 100}); err == nil {
					evs = append(evs, rows...)
				}
			}
		}
		var draftRows []flowbind.InboxDraftRow
		if drafts != nil {
			for _, d := range drafts.List() {
				draftRows = append(draftRows, flowbind.InboxDraftRow{Kind: d.Kind, Scope: d.Scope,
					Name: d.Name, FieldManager: d.FieldManager, UpdatedAt: d.UpdatedAt.UTC().Format(time.RFC3339Nano)})
			}
		}
		var propRows []flowbind.InboxProposalRow
		if objects != nil {
			if props, err := objects.List(context.Background(), proposalKind, ""); err == nil {
				for _, po := range props {
					if spec, ok := proposal.Parse(po.Content); ok {
						propRows = append(propRows, flowbind.InboxProposalRow{Name: po.Name, Scope: po.Scope,
							Title: spec.Title, Status: spec.Status, UpdatedAt: ""})
					}
				}
			}
		}
		st := o.inbox.State()
		pending := 0
		var pendRows []flowbind.InboxRow
		for _, r := range st.Records {
			if r.State != flowbind.ApprovalPending {
				continue // v0.58：终态行（已判定）零挂起行——回链闭环拍即生效（动作面 console 动作化）
			}
			pending++
			// v0.57 挂起行进行集（console-adoption.md §1.32 L59-UI-02——幽灵计数缺口收口）：
			// ask 批准投影挂起行=ApprovalNeeded 类条目行（行集与计数同源不翻倍——加法投影
			// 摘除，计数=行数）。落点=派单投影所在 wf 详情配置 tab（#/wf/<锚>——批准动作=
			// Apply workflow:approve 判定链路）。
			name := r.Connector
			anchor := r.Workflow
			if r.Kind == flowbind.DispatchKindProposal || r.Proposal != "" {
				name = r.Proposal
				anchor = r.Proposal
			}
			pendRows = append(pendRows, flowbind.InboxRow{
				Class: flowbind.InboxClassApprovalNeeded, Kind: "dispatch", Name: name, Scope: r.Scope,
				Digest:  "派单待批（" + anchor + " · autoLevel=" + r.AutoLevel + "）",
				EventID: r.EventID, At: r.At.UTC().Format(time.RFC3339Nano),
				Link: "#/wf/" + anchor,
			})
		}
		box := flowbind.BuildInbox(evs, draftRows, propRows, nil)
		if len(pendRows) > 0 {
			box.Items = append(box.Items, pendRows...) // 人注意优先类序由 BuildInbox 排序——挂起行同类合并后重排
			flowbind.SortInboxRows(box.Items)
			box.Count = len(box.Items)
			if box.Classes == nil {
				box.Classes = map[string]int{}
			}
			box.Classes[flowbind.InboxClassApprovalNeeded] += len(pendRows)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"count": box.Count, "classes": box.Classes, "items": box.Items,
			"approvals": map[string]any{"count": len(st.Records), "pending": pending,
				"records": approvalsPayload(st.Records)},
			"watchdog": map[string]any{"approvals": len(st.Records), "pending": pending,
				"approveIdle": o.inbox.ApproveIdle()},
		})
	})
	// 数据面查询面（v0.33 MetricsInfo 数据面首段——store/queries 投影读面）：指标清单
	// （样本保留数，排序确定性投影）与窗口聚合查询（queries 最小集投影）。数据面未装配
	// =空投影（关断口径可见）；查询未命中/声明面外聚合=404（不冒充零值）。数据面零信号
	// 面依赖同口径——本面只读样本库，零总线零判定。
	// 会话活性面（KI-003，event.md 规则 14；agent-session.md §7）：
	// POST /v1/sessions/heartbeat=会话侧心跳注册（--allow-apply 门控——写面动作与注入面同开关）；
	// GET /v1/sessions=查询面恒开（判定态=state 词 suspected-stall 判定词 SuspectedStall 的 DB 词形）。
	mux.HandleFunc("POST /v1/sessions/heartbeat", func(w http.ResponseWriter, r *http.Request) {
		if o.sessionLiveness == nil {
			writeJSON(w, http.StatusNotImplemented, map[string]any{"error": "sessionliveness: 会话活性面未装配（--session-watch-interval）"})
			return
		}
		var req sessionliveness.Heartbeat
		if err := json.NewDecoder(io.LimitReader(r.Body, 64<<10)).Decode(&req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "sessionliveness: 心跳载荷解析失败: " + err.Error()})
			return
		}
		// 心跳防伪（v0.56——agent-session.md §7 心跳面唯一口径）：watchdog 自身动作的续租
		// （source=watchdog——注入「继续」后的活性观测心跳）不可经本面自报（来源伪冒=自续租
		// 死循环）；watchdog 侧经 BeatSource 落行，HTTP 面一律拒收（形态校验前置于写面门控）。
		// 判定/注入动作不续租语义（v0.54 既有：状态迁移不续租）随行不变。
		if req.Source == "watchdog" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "sessionliveness: 心跳来源=watchdog 不可经注册面上报（判定面自续租禁入——agent-session.md §7 心跳面唯一口径）"})
			return
		}
		if applier == nil {
			writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": "写面关断（--allow-apply 未开启）——心跳注册面随写面门控"})
			return
		}
		if err := o.sessionLiveness.Beat(req); err != nil {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"ok": true, "sessionId": req.SessionID, "at": time.Now().UTC().Format(time.RFC3339Nano)})
	})
	mux.HandleFunc("GET /v1/sessions", func(w http.ResponseWriter, r *http.Request) {
		if o.sessionLiveness == nil {
			writeJSON(w, http.StatusNotImplemented, map[string]any{"error": "sessionliveness: 会话活性面未装配（--session-watch-interval）"})
			return
		}
		limit := atoiDefault(r.URL.Query().Get("limit"), 200)
		rows, err := o.sessionLiveness.Heartbeats(limit)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"count": len(rows), "sessions": rows,
			"watchdog": map[string]any{
				"enabled":        o.sessionLiveness.Enabled(),
				"interval":       o.sessionLiveness.Interval().String(),
				"stallThreshold": o.sessionLiveness.Threshold().String(),
				"healRetryLimit": o.sessionLiveness.HealLimit(),
			},
		})
	})
	mux.HandleFunc("GET /v1/metrics", func(w http.ResponseWriter, _ *http.Request) {
		items := []map[string]any{}
		if metrics != nil {
			for _, name := range metrics.Metrics() {
				items = append(items, map[string]any{"name": name, "samples": metrics.Count(name)})
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{"metrics": items, "total": len(items)})
	})
	mux.HandleFunc("GET /v1/metrics/query", func(w http.ResponseWriter, r *http.Request) {
		qm := r.URL.Query().Get("metric")
		qw := r.URL.Query().Get("window")
		qa := r.URL.Query().Get("agg")
		if qm == "" || qa == "" {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": "metrics 查询需 metric 与 agg 参数"})
			return
		}
		if metrics == nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "数据面未装配（关断口径）"})
			return
		}
		v, ok := metrics.Query(metricsinfo.Query{Metric: qm, Window: qw, Agg: qa})
		if !ok {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "无投影结果（未知聚合或无同键样本）", "metric": qm, "window": qw, "agg": qa})
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"metric": qm, "window": qw, "agg": qa, "value": v})
	})
	// 事件归档查询面（v0.36 console 读面首段，cored-bridge.md §2 `GET /v1/events`）：
	// event_log 归档行的最新优先查询投影（limit/type/level/involved/since/until 六参最小集；
	// limit≤0=缺省 50，>500 收敛 500——读面防拖库，全量走重建/导出面）。since/until=半开窗
	// [since, until)（until 严格不含端点——E7 向后分页游标：取上一页最旧行 ts 翻页，零重叠）。
	// 只读投影恒开（face 随 ops DB 装配恒在）；旧构建（face 缺席路由未注册）经兜底 handler
	// 返回 405=代际判别器。数据来源=归档（event.md：审计素材）， Welcome 最近重要事件与
	// 事件流页（E6）共用本面——渲染数据无双源红线（console-adoption.md §5 规则 2）。
	mux.HandleFunc("GET /v1/events", func(w http.ResponseWriter, r *http.Request) {
		if eventLog == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "事件归档查询面未接线（装配缺 ops DB 口面）"})
			return
		}
		q := opslog.EventQuery{
			Type:     r.URL.Query().Get("type"),
			Level:    r.URL.Query().Get("level"),
			Involved: r.URL.Query().Get("involved"),
			Limit:    atoiDefault(r.URL.Query().Get("limit"), 0),
		}
		if raw := r.URL.Query().Get("since"); raw != "" {
			t, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("since 不可解析（需 RFC3339）: %v", err)})
				return
			}
			q.Since = t
		}
		if raw := r.URL.Query().Get("until"); raw != "" {
			t, err := time.Parse(time.RFC3339, raw)
			if err != nil {
				writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("until 不可解析（需 RFC3339）: %v", err)})
				return
			}
			q.Until = t
		}
		rows, err := eventLog.QueryEvents(q)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		items := make([]map[string]any, 0, len(rows))
		for _, e := range rows {
			items = append(items, map[string]any{
				"ts": e.Ts.UTC().Format(time.RFC3339Nano), "eventId": e.EventID,
				"type": e.Type, "source": e.Source, "involved": e.Involved,
				"dedupKey": e.DedupKey, "level": e.Severity, "routedTo": e.RoutedTo,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"count": len(items), "events": items})
	})
	// workflow 运行管理聚合面（v0.37 console 读面第二段，E3——cored-bridge.md §2
	// `GET /v1/workflows`）：三源同投（① 声明=builtin 注册表+文档库 Workflow 对象；
	// ② 现势=run 环每锚最新 run 的相位/状态；③ 计数=event_log 归档分组 × dispatch 决议，
	// workflow.RunAnchor 与 run 投影同一锚解析）。只读 GET；旧构建（路由未注册）经兜底
	// handler 405=代际判别器。投影/构造细节见 workflows_face.go。
	mux.HandleFunc("GET /v1/workflows", func(w http.ResponseWriter, r *http.Request) {
		rows, scanned, capped := collectWorkflowRows(r.Context(), objects, runs, triggers, eventLog, applyLog)
		items := make([]map[string]any, 0, len(rows))
		for _, row := range rows {
			items = append(items, map[string]any{
				"name": row.Name, "specSource": row.SpecSource, "phases": row.Phases,
				"objects": workflowObjectsJSON(row.Objects), "effective": row.Effective,
				"latest": row.Latest, "runsTotal": row.RunsTotal, "runs24h": row.Runs24h,
				"execution": row.Execution, // v0.49 R-7 执行模型绑定面（新载荷键=代际判别器）
				// v0.55 R-9 构建生效链判别面（console-adoption.md §1.29——骨架 rv ↔ 绑定 apply 行
				// rv 对账+骨架节点接线投影；新载荷键=代际判别器）。
				"reconcile":        row.Reconcile,
				"executionFlow":    flowNodesJSON(row.Reconcile),
				"executionDigest":  executionDigest(row.Reconcile),
				"executionHistory": workflowHistoryRef(row),
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"count": len(items), "window": "24h",
			"scanned": scanned, "scanCapped": capped, // 计数=有界窗如实标注（不冒充全量承诺）
			"workflows": items,
		})
	})
	// 对象应用史投影面（v0.38 console 读面第三段，E4——cored-bridge.md §2 `GET /v1/history/{kind}/{name}`）：
	// 数据源=ops DB apply_log（ops-log.md §2 一次 Apply 判定的流水；repository 双写 ok/noop/conflict
	// 行+引擎 denied 行）——rev 流水的权威=审计账（git 权威在正文，不在此考古）。obj=kind/scope/name
	// 精确过滤+rv 降序+limit 收敛（≤500 防拖库同口径）。只读 GET；face 缺席=503（防御面，装配层
	// 随 ops DB 恒注入）；旧构建（路由未注册）经兜底 handler 405=代际判别器。
	mux.HandleFunc("GET /v1/history/{kind}/{name}", func(w http.ResponseWriter, r *http.Request) {
		if applyLog == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "应用史查询面未接线（装配缺 ops DB 口面）"})
			return
		}
		kind, name := r.PathValue("kind"), r.PathValue("name")
		scope := r.URL.Query().Get("scope")
		obj := fmt.Sprintf("%s/%s/%s", kind, scope, name)
		limit := atoiDefault(r.URL.Query().Get("limit"), 0)
		rows, err := applyLog.QueryApplies(obj, limit)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		items := make([]map[string]any, 0, len(rows))
		for _, a := range rows {
			items = append(items, map[string]any{
				"ts": a.Ts.UTC().Format(time.RFC3339Nano), "rv": a.ResourceVersion,
				"result": a.Result, "fieldManager": a.FieldManager, "subject": a.Subject,
				"reasonRef": a.ReasonRef,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"count": len(items), "obj": obj, "entries": items})
	})
	// 历史 rv 正文投影面（v0.63 rev-content diff 面，§1.40——历史 tab「点行 diff 上一版」余量
	// 收口的读半边）：`GET /v1/rev-content/{kind}/{name}?scope=s&rv=N` 按对象键+目标 rv 取该
	// 提交时点权威正文快照（数据源=git 后端权威正文——concepts/revision.md §3 git=内容版本权威，
	// ops DB 只记 rv 流水；快照定位=提交消息键段+blob rv 反写行双段对账）。只读 GET 零副作用；
	// face 缺席=503 防御面（sqlite 后端/旧装配——诚实降级非 200 假面）；旧构建 405=代际判别器；
	// 对象键不可解析/缺席 → 400/404 分立；rv 缺席或不可解析 → 400；该 rv 无落库提交 → 404
	// ErrRevNotFound 分立词形（与对象缺席分立——审计不冒充）。
	mux.HandleFunc("GET /v1/rev-content/{kind}/{name}", func(w http.ResponseWriter, r *http.Request) {
		if revContent == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "历史 rv 正文读面未接线（face 缺席——sqlite 后端/旧装配诚实降级）"})
			return
		}
		kind, name := r.PathValue("kind"), r.PathValue("name")
		scope := r.URL.Query().Get("scope")
		rawRV := r.URL.Query().Get("rv")
		rv, err := strconv.ParseInt(rawRV, 10, 64)
		if err != nil || rv <= 0 {
			writeJSON(w, http.StatusBadRequest, map[string]any{"error": fmt.Sprintf("rv 必须为正整数（乐观锁词形），得 %q", rawRV)})
			return
		}
		ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
		defer cancel()
		content, err := revContent.RevContent(ctx, kind, scope, name, rv)
		switch {
		case err == nil:
			writeJSON(w, http.StatusOK, map[string]any{
				"kind": kind, "scope": scope, "name": name, "rv": rv, "content": content,
				"source": "git-backend",
			})
		case errors.Is(err, repository.ErrRevNotFound):
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error(), "obj": fmt.Sprintf("%s/%s/%s", kind, scope, name), "rv": rv})
		case errors.Is(err, repository.ErrNotFound):
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
		case errors.Is(err, repository.ErrScope):
			writeJSON(w, http.StatusBadRequest, map[string]error{"error": err})
		default:
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
		}
	})
	// 审批判定账面查询面（v0.46 审批独立账，ops-log.md §2 v0.46 注记——`GET /v1/approvals`）：
	// 数据源=ops DB approval_log（审批判定独立账：批准 granted/驳回 rejected——apply_log 承
	// 落库流水，本账承"谁在何时对哪条提案做了什么判定"；驳回=弃稿零文档面动作也落行）。
	// decision/obj 精确过滤+rv 降序+limit 收敛（≤500 防拖库同口径）。只读 GET；face 缺席=503
	// （防御面，装配层随 ops DB 恒注入）；旧构建（路由未注册）经兜底 handler 405=代际判别器。
	mux.HandleFunc("GET /v1/approvals", func(w http.ResponseWriter, r *http.Request) {
		if approval == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "审批判定账面查询面未接线（装配缺 ops DB 口面）"})
			return
		}
		qs := opslog.ApprovalQuery{
			Decision: r.URL.Query().Get("decision"),
			Obj:      r.URL.Query().Get("obj"),
			Limit:    atoiDefault(r.URL.Query().Get("limit"), 0),
		}
		rows, err := approval.QueryApprovals(qs)
		if err != nil {
			writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
			return
		}
		items := make([]map[string]any, 0, len(rows))
		for _, a := range rows {
			items = append(items, map[string]any{
				"ts": a.Ts.UTC().Format(time.RFC3339Nano), "approvalId": a.ApprovalID,
				"obj": a.Obj, "decision": a.Decision, "judgedBy": a.JudgedBy,
				"fieldManager": a.FieldManager, "rv": a.ResourceVersion, "reasonRef": a.ReasonRef,
			})
		}
		writeJSON(w, http.StatusOK, map[string]any{"count": len(items), "approvals": items})
	})
	// 需求提案读面（v0.51 能力 B——cored-bridge.md §2 `GET /v1/proposals`；concepts/proposal.md）：
	// 数据源=文档库 Proposal 对象（零缓存现读）；列表=瘦身投影（?status= 闭合集过滤+?scope=
	// 枚举收窄，rv 降序确定性投影）；详情=全 spec 投影+content 原文（scope 缺省=global 优先→其余
	// 首个）。只读 GET 恒开（objects 缺席=503；kind 未注册=404——scheme 闸门语义）；旧构建（路由
	// 未注册）经兜底 handler 405=代际判别器。写路径零新端点（Proposal 走 apply 管线——proposal.md 规则 1）。
	mux.HandleFunc("GET /v1/proposals", func(w http.ResponseWriter, r *http.Request) {
		if objects == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "对象存储未配置（cored 缺 --docs）"})
			return
		}
		if _, err := sch.Lookup(proposalKind); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
			return
		}
		statusFilter, scopeQ := r.URL.Query().Get("status"), r.URL.Query().Get("scope")
		objs, err := objects.List(r.Context(), proposalKind, scopeQ)
		if err != nil {
			writeObjErr(w, err)
			return
		}
		rows := make([]proposalRow, 0, len(objs))
		for _, o := range objs {
			row := proposalRowOf(o.Name, o.Scope, o.ResourceVersion, o.Content)
			if statusFilter != "" && row.Status != statusFilter {
				continue // 状态过滤（闭合集词形——词外值=恒空结果不冒充全量）
			}
			rows = append(rows, row)
		}
		sortProposals(rows)
		writeJSON(w, http.StatusOK, map[string]any{"count": len(rows), "proposals": rows})
	})
	mux.HandleFunc("GET /v1/proposals/{name}", func(w http.ResponseWriter, r *http.Request) {
		if objects == nil {
			writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": "对象存储未配置（cored 缺 --docs）"})
			return
		}
		name, scopeQ := r.PathValue("name"), r.URL.Query().Get("scope")
		if _, err := sch.Lookup(proposalKind); err != nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
			return
		}
		if scopeQ != "" { // 键全给定=直取（404 语义与 objects 面同口径）
			o, err := objects.Get(r.Context(), proposalKind, scopeQ, name)
			if err != nil {
				writeObjErr(w, err)
				return
			}
			writeJSON(w, http.StatusOK, proposalDetail(o))
			return
		}
		objs, err := objects.List(r.Context(), proposalKind, "")
		if err != nil {
			writeObjErr(w, err)
			return
		}
		var pick *repository.Object
		for _, o := range objs {
			if o.Name != name {
				continue
			}
			if o.Scope == "global" {
				pick = o
				break // global 同名优先（确定性投影）
			}
			if pick == nil {
				pick = o
			}
		}
		if pick == nil {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "proposal 未在册（对象键=(kind,scope,name)）"})
			return
		}
		writeJSON(w, http.StatusOK, proposalDetail(pick))
	})
	// 设计文档面（v0.52 能力 A——console 屏幕 6 直接编辑设计文档 md → Apply → git commit）：
	// 读/史/草稿/apply 路由注册（handler 体见 design_docs.go）。面未装配=503；写面随
	// --allow-apply；白名单外 403 / lint 未过 400 携违规 / 基线 sha 不符 409——任务书判据三件。
	registerDesignDocRoutes(mux, designDocs, drafts, applyLog)
	// 对话式录入与三重 dry-run 预检面（v0.53 能力 D + R-UX-1/R-UX-2——cored-bridge.md §2.5）：
	// GET /v1/intake/sessions* 恒开（旧构建 405=代际判别器）；POST /v1/intake/turn 与 /generate 随 --allow-apply 门控。
	registerIntakeRoutes(mux, sch, objects, applier)
	// 其余路径与方法一律 404/405，保持查询面最小。
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		if !strings.HasPrefix(r.URL.Path, "/v1/") && r.URL.Path != "/healthz" {
			writeJSON(w, http.StatusNotFound, map[string]any{"error": "not found"})
			return
		}
		writeJSON(w, http.StatusMethodNotAllowed, map[string]any{"error": fmt.Sprintf("method %s not allowed", r.Method)})
	})
	return mux
}

func objRef(o *repository.Object) map[string]any {
	return map[string]any{"kind": o.Kind, "name": o.Name, "scope": o.Scope, "resourceVersion": o.ResourceVersion}
}

// draftRef 草稿瘦身引用（清单载荷：无正文；bytes 供容量观察）。
func draftRef(d *draft.Draft) map[string]any {
	return map[string]any{
		"kind": d.Kind, "name": d.Name, "scope": d.Scope,
		"fieldManager":        d.FieldManager,
		"baseResourceVersion": d.BaseRV,
		"updatedAt":           d.UpdatedAt.UTC().Format(time.RFC3339Nano),
		"bytes":               len(d.Content),
	}
}

// writeObjErr 仓储哨兵错误 → HTTP 状态码映射。
func writeObjErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, repository.ErrNotRegistered), errors.Is(err, repository.ErrNotFound):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
	case errors.Is(err, repository.ErrScope):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
}

// applyRequest POST /v1/apply 请求体（cored-bridge.md 契约；content=文档原文）。
type applyRequest struct {
	Kind            string          `json:"kind"`
	Scope           string          `json:"scope"`
	Name            string          `json:"name"`
	ResourceVersion int64           `json:"resourceVersion"`
	Content         string          `json:"content"`
	FieldManager    string          `json:"fieldManager"`
	ReasonRef       string          `json:"reasonRef"`
	DryRun          bool            `json:"dryRun"`
	Force           bool            `json:"force"`
	FromDraft       bool            `json:"fromDraft"` // true=稿→面（§2.2）；content/resourceVersion 须缺省
	Subject         *subjectRequest `json:"subject"`   // 请求级主体（v0.9）；缺席回退 X-Core-Subject 头
}

// subjectRequest 请求级主体（v0.9，permission.md §3 的 HTTP 投影）：
// user=主体标识（user-system 目录侧未接前由调用方断言，审计留痕）；
// roles/groups=主体断言的持有关系——权限只由文档面 Role/Group 对象授予，断言名须能对上装载的对象。
type subjectRequest struct {
	User   string   `json:"user"`
	Roles  []string `json:"roles"`
	Groups []string `json:"groups"`
}

// SubjectHeader 主体头名（与请求体 subject 字段同形 JSON；桥接方不便改体时用头）。
const SubjectHeader = "X-Core-Subject"

// ApproveFieldManager 批准判定的 fieldManager 约定串（E6 审批屏批准=Apply 全流程复用；
// v0.46 审批独立账判定识别键——console 客户端与 cored 判定面共用此约定，双账并存：
// apply_log 承落库流水，approval_log 承审批判定）。
const ApproveFieldManager = "console:approve"

// RejectFieldManager 驳回判定的 fieldManager 约定串（v0.51 能力 B 对称扩展——Proposal 屏
// 驳回=对象态写 rejected 的判定动作（rv=落库后版本，approval_log rejected 行+ApprovalRejected
// 事件）；与弃稿 DELETE（rv=0 零文档面动作）两分——两种「不采纳」的判定都在账）。
const RejectFieldManager = "console:reject"

// subjectFrom 请求级主体解析（apply/events 写面共用）：subject 字段优先，缺席回退
// X-Core-Subject 头；两者皆缺 → 匿名主体（非 nil 指针=强制请求级判定，进程级本地信任
// 不兜底 HTTP 匿名请求）。
func subjectFrom(r *http.Request, sr *subjectRequest) *permission.Subject {
	if sr != nil {
		return &permission.Subject{User: sr.User, Roles: sr.Roles, Groups: sr.Groups}
	}
	if h := r.Header.Get(SubjectHeader); h != "" {
		var s subjectRequest
		if err := json.Unmarshal([]byte(h), &s); err == nil {
			return &permission.Subject{User: s.User, Roles: s.Roles, Groups: s.Groups}
		}
		// 头不可解析=视为未携带（匿名照判拒绝；不静默降级成本地信任）。
	}
	return &permission.Subject{}
}

// resolveSubject POST /v1/apply 请求级主体解析（subjectFrom 的 apply 侧入口）。
func resolveSubject(r *http.Request, req *applyRequest) *permission.Subject {
	return subjectFrom(r, req.Subject)
}

// eventsRequest POST /v1/events 请求体（v0.20 发布面；source 仅接受 human 值或缺省——
// 服务侧强制判定，其他值=400）。
type eventsRequest struct {
	Type     string          `json:"type"`
	Involved string          `json:"involved"`
	Level    string          `json:"level"`
	DedupKey string          `json:"dedupKey"`
	Source   string          `json:"source"` // 载荷声明不采纳：非 human 值直接拒绝（来源=服务侧判定）
	Subject  *subjectRequest `json:"subject"`
}

// writePublishErr 发布面错误 → HTTP 状态码映射（cored-bridge.md §2.3 v0.20）。
func writePublishErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, eventpub.ErrNotRegistered):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
	case errors.Is(err, eventpub.ErrDenied):
		writeJSON(w, http.StatusForbidden, map[string]any{"error": err.Error()})
	case errors.Is(err, eventpub.ErrNoType), errors.Is(err, eventpub.ErrLevel),
		errors.Is(err, eventbus.ErrRejected):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
}

// draftRequest POST /v1/drafts 请求体（cored-bridge.md §2.2）。
type draftRequest struct {
	Kind                string `json:"kind"`
	Scope               string `json:"scope"`
	Name                string `json:"name"`
	Content             string `json:"content"`
	FieldManager        string `json:"fieldManager"`
	ReasonRef           string `json:"reasonRef"`
	BaseResourceVersion int64  `json:"baseResourceVersion"` // 落稿所基于的文档面 rv（0=预期创建）
}

// writeApplyErr Apply 管线错误 → HTTP 状态码映射（cored-bridge.md §2.1 错误表）。
func writeApplyErr(w http.ResponseWriter, err error) {
	var conflict *repository.ConflictError
	switch {
	case errors.As(err, &conflict):
		writeJSON(w, http.StatusConflict, map[string]any{
			"error": err.Error(), "expected": conflict.Expected, "current": conflict.Current,
		})
	case errors.Is(err, apply.ErrNotRegistered), errors.Is(err, repository.ErrNotRegistered),
		errors.Is(err, repository.ErrNotFound), errors.Is(err, draft.ErrNoDraft):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
	case errors.Is(err, apply.ErrDenied):
		writeJSON(w, http.StatusForbidden, map[string]any{"error": err.Error()})
	case errors.Is(err, apply.ErrNoFieldManager), errors.Is(err, apply.ErrObject),
		errors.Is(err, apply.ErrSpecInvalid),
		errors.Is(err, repository.ErrScope), errors.Is(err, repository.ErrIdentity),
		errors.Is(err, repository.ErrContent), errors.Is(err, repository.ErrAlreadyExists),
		errors.Is(err, repository.ErrNoFieldManager), errors.Is(err, repository.ErrCredentialSource):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
	case errors.Is(err, repository.ErrClosed):
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
}

// writeDraftErr 草稿面错误 → HTTP 状态码映射（cored-bridge.md §2.2）。
func writeDraftErr(w http.ResponseWriter, err error) {
	switch {
	case errors.Is(err, draft.ErrNoDraft):
		writeJSON(w, http.StatusNotFound, map[string]any{"error": err.Error()})
	case errors.Is(err, draft.ErrInvalid):
		writeJSON(w, http.StatusBadRequest, map[string]any{"error": err.Error()})
	case errors.Is(err, draft.ErrNoStore):
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{"error": err.Error()})
	default:
		writeJSON(w, http.StatusInternalServerError, map[string]any{"error": err.Error()})
	}
}

// atoiDefault 极小整数查询参解析（空/非法=def；console 读面 limit 用——防拖库收敛在
// opslog.QueryEvents 内，本处不复制上限常量）。
func atoiDefault(raw string, def int) int {
	if raw == "" {
		return def
	}
	n, err := strconv.Atoi(raw)
	if err != nil {
		return def
	}
	return n
}

func writeJSON(w http.ResponseWriter, code int, payload any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(code)
	enc := json.NewEncoder(w)
	enc.SetEscapeHTML(false)
	_ = enc.Encode(payload)
}

// SessionLivenessFace 会话活性面口面（sessionliveness.Watchdog 满足；KI-003——心跳注册面
// POST /v1/sessions/heartbeat 与查询面 GET /v1/sessions；判定面=装配层 watchdog，零新 face）。
type SessionLivenessFace interface {
	Beat(h sessionliveness.Heartbeat) error
	Heartbeats(limit int) ([]sessionliveness.SessionView, error)
	Enabled() bool
	Interval() time.Duration
	Threshold() time.Duration
	HealLimit() int
}

// WithSessionLiveness 注入会话活性面（KI-003：--session-watch-interval 装配；nil=面缺席）。
func WithSessionLiveness(f SessionLivenessFace) Option {
	return func(o *options) { o.sessionLiveness = f }
}

// approvalsPayload 收件箱批准投影行（/v1/inbox approvals.records 载荷——pending/approved/
// rejected 三态词形随行；judgedAt 零值=未判定不投影）。
func approvalsPayload(rs []flowbind.ApprovalRecord) []map[string]any {
	out := make([]map[string]any, 0, len(rs))
	for _, r := range rs {
		out = append(out, map[string]any{
			"seq": r.Seq, "kind": r.Kind, "workflow": r.Workflow, "scope": r.Scope,
			"connector": r.Connector, "proposal": r.Proposal, "state": r.State,
			"autoLevel": r.AutoLevel, "rv": r.RV, "eventId": r.EventID,
			"approvalId": r.ApprovalID, "judgedBy": r.JudgedBy,
			"judgedAt": nilIfZero(r.JudgedAt), "at": r.At.UTC().Format(time.RFC3339Nano),
		})
	}
	return out
}

// nilIfZero 零值时刻不投影（nil=未判定——与 SessionView HealedAt 同口径）。
func nilIfZero(t time.Time) any {
	if t.IsZero() {
		return nil
	}
	return t.UTC().Format(time.RFC3339Nano)
}

// matchingKeyOfRecord 派单行的批准判定键（/v1/agents/dispatches approval.approvalId——
// flowbind.DispatchRecord → matchingKeyOf 同源；空 connector/proposal 行=零值键不匹配任何投影）。
func matchingKeyOfRecord(d flowbind.DispatchRecord) string {
	return flowbind.ApprovalKeyOf(d)
}

// syncDeniedJudgmentRow denied 判定行随行同步（v0.57——批准投影翻转判定的 denied 词形补口：
// apply_log 审计=判定权威（引擎 denied 行照落——apply.go audit），观测投影此前只喂 ok/noop
// 行=denied 判定缺席的投影缺口（判定行重扫词形与账本行集分叉）。本函数只投影账本既有事实，
// 零新写路径零旁路（Sync 匹配判定零翻转——denied 不判定，TestSyncApprovalsFlipByApplyRows
// 锚定）；applyLog face 缺席/非判定串=静默零投影）。
func syncDeniedJudgmentRow(req applyRequest, opt apply.Options, err error, face InboxFace) {
	if face == nil {
		return
	}
	switch req.FieldManager {
	case ApproveFieldManager, flowbind.ApprovalManager, RejectFieldManager, flowbind.ApprovalRejecter:
	default:
		return // 非判定溯源串=零投影（普通 apply 拒绝不参与批准判定面）
	}
	if !errors.Is(err, apply.ErrDenied) && !errors.Is(err, apply.ErrNotRegistered) {
		return // 非 deny 族错误（conflict/传输层）不投——判定行词形=denied 族
	}
	sub := "anonymous"
	if opt.Subject != nil {
		sub = opt.Subject.ID()
	}
	face.Sync([]opslog.ApplyLog{{
		Ts: time.Now(), FieldManager: req.FieldManager, Subject: sub,
		Obj:    fmt.Sprintf("%s/%s/%s", req.Kind, req.Scope, req.Name),
		Result: "denied", PolicyDecisions: err.Error(),
	}})
}

// spawnPayload spawn 接线声明 → 载荷行（/v1/agents/dispatches spawn 键——键恒在=代际判别器；
// 零值声明=接线缺席 `{declared:false, wired:false}` 诚实词形——pending/rejected 行或批准投影
// 面缺席照落零值，不冒充已接线）。
func spawnPayload(d flowbind.SpawnDeclaration) map[string]any {
	return map[string]any{
		"declared": d.Declared, "wired": d.Wired, "target": d.Target,
		"approvalId": d.ApprovalID, "rv": d.RV, "eventId": d.EventID,
	}
}
