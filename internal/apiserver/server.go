// Package apiserver cored 查询面：scheme 加载校验结果 + 对象类型清单 get/list
// + 文档对象读面（/v1/objects，读库内真实文档对象而非类型注册表）。
// 设计依据：core_system_instance/runtime/cored-bridge.md（plugin↔cored 桥契约）。
// 缺省只读；写路径唯一入口是 Apply 管线（core_design/concepts/apply.md）——仅当部署方显式开启
// --allow-apply 时暴露 POST /v1/apply，内部收敛到 apply.Engine 三步闸门，不旁路存储。
// R-ENG-7 拆分注记：本文件=口面/选项/装配（New）+公共 helper；路由 handler 按域分文件
// 纯搬家（routes_objects.go / routes_events.go / routes_workflow.go / routes_read.go）。
package apiserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
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
// 余量收口的读半边；实现=repository.GitRepo「git-backend」与 repository.SQLiteRepo
// 「sqlite-doc-history」双后端分立词形）。语义=按对象键与目标 rv 取**该次时点**的权威正文快照
// （git 后端权威=仓库文件历史提交考古；sqlite 后端权威=doc_history 逐次账——R-DOC-7 规则 4，
// 周期 git 导出快照不得冒充逐次历史）；face 缺席=503 防御面（旧代兜底 405=代际判别器恒成立）。
// 具体方法不占冻结 Repository 签名（QueryApplies 口面同款约定，compile 期断言在 New 尾注记）。
type RevContentFace interface {
	// RevContent 取对象键 (kind, scope, name) 在 rv 时点的正文快照。找不到该 rv 的落库快照
	// → ErrRevNotFound（HTTP 404，与对象缺席分立）；读面零副作用。
	RevContent(ctx context.Context, kind, scope, name string, rv int64) (string, error)
	// Source 数据源词形标注（响应 source 字段据实回显——git 裸库考古=「git-backend」、
	// doc_history 逐次账=「sqlite-doc-history」；不冒充双后端同源）。
	Source() string
}

// WithRevContent 注入历史 rv 正文读面（v0.63 rev-content diff 面——历史 tab 点行 diff 上一版
// 读半边；git/sqlite 后端各按其数据源装配，词形分立见 Source）。
func WithRevContent(f RevContentFace) Option {
	return func(o *options) { o.revContent = f }
}

// DocHistoryFace 逐次正文历史清单读面口面（R-DOC-7 规则 4 doc_history 落地面——sqlite 后端
// 实现=repository.SQLiteRepo；git 后端不做：git 历史=周期导出快照非逐次账，面缺席=503 不冒充）。
// 与 GET /v1/history（apply_log 审计流水——谁/何时/rv/结果，ops-log.md §2）分工：本面=
// 逐次正文翻页清单（rev/翻页时刻/写者/正文字节数——不含正文；正文经 GET /v1/rev-content?rv=N
// 按 rv 点取）。git 后端装配缺席=503 诚实降级（导出快照不冒充逐次清单）。
type DocHistoryFace interface {
	// History 对象逐次翻页行集（rev 降序）。对象缺席 → ErrNotFound；在册无翻页史 → 空行集。
	History(kind, scope, name string) ([]repository.DocHistoryRow, error)
}

// WithDocHistory 注入逐次正文历史清单读面（R-DOC-7 规则 4——GET /v1/doc-history 读面；
// sqlite 后端恒装配；git 后端缺席=面 503 诚实降级）。
func WithDocHistory(f DocHistoryFace) Option {
	return func(o *options) { o.docHistory = f }
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
	docHistory      DocHistoryFace
	dispatch        DispatchFace
	inbox           InboxFace
	designDocs      DesignDocs
	sessionLiveness SessionLivenessFace
	agentScopes     flowbind.AgentEventScopes // Agent 事件 scope 声明解析（BUG-010 消费闸——GET /v1/events?agent=）
	contextIdx      contextIdxFace
	perm            *permission.System // 权限判定权威（G-03 connectops 权限门；nil=零值系统 fail-closed）
	bus             *eventbus.Bus      // stream 面变更感知源（v0.66 WithEventBus；nil=仅 5s 兜底重算）
}

// WithContextIdx 注入上下文索引读面（v0.65 context P0 数据面 + trace 1:N 追踪——
// /v1/context/* 与 /v1/trace/*；contextidx.Index 满足；nil=面缺席 503 同款防御面）。
func WithContextIdx(f contextIdxFace) Option {
	return func(o *options) { o.contextIdx = f }
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

// WithAgentEventScopes 注入 Agent 事件 scope 声明解析面（BUG-010 消费闸——
// `GET /v1/events?agent=<锚>` 逐行过 flowbind.AgentScopeGate 同一判定；flowbind 派发闸
// 与本面共用同一 Resolver 实例零双源）。nil=消费闸缺席：agent 参数 503（面注册但能力
// 缺席，与事件归档面缺席同口径词形）。
func WithAgentEventScopes(r flowbind.AgentEventScopes) Option {
	return func(o *options) { o.agentScopes = r }
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
	_ RevContentFace      = (*repository.SQLiteRepo)(nil)
	_ DocHistoryFace      = (*repository.SQLiteRepo)(nil)
	_ DesignDocs          = (*docrepo.Set)(nil)
	_ SessionLivenessFace = (*sessionliveness.Watchdog)(nil)
)

// deps 路由 handler 的依赖收拢（R-ENG-7 拆分：13 个 With* Option 对外签名不变，内部
// 汇总进本结构经 register*Routes 按域下发——New 与 handler 之间不再散装十余个局部闭包变量）。
type deps struct {
	sch         *scheme.Scheme
	objects     ObjectSource
	applier     Applier
	drafts      DraftStaging
	triggers    TriggerFace
	runs        RunFace
	deliveries  DeliveryFace
	wakes       WakeFace
	publisher   PublisherFace
	logApproval func(opslog.ApprovalLog) // 审批判定账面落账（nil 账面=空操作）
	o           *options                 // 可选能力 face（缺席=503/空投影防御面）
}

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
//	GET /v1/events/{eventId}             单事件详情（v0.66 R-ENG §3：归档直投+关联链——同
//	                                     eventId 的 run/同 dedupKey 后续事件/involved 键解析；
//	                                     404=不在归档）
//	GET /v1/loop/current                 当前/最近一轮自循环投影（v0.66 R-ENG §3：wakes face
//	                                     最近 LoopRequested 行直投；只读聚合不发明状态机）
//	GET /v1/snapshot                     console 首屏聚合快照（v0.66 R-ENG §2：14 键载荷——
//	                                     objects/events/tasks/runs 关键聚合在位（G-02）；face
//	                                     缺席=null 不冒充）
//	GET /v1/stream                       WebSocket 推送面（v0.66 R-ENG §2：连接即推 snapshot，
//	                                     变更推 RFC6902 patch（保守 diff 器）；30s ping）
//	GET /v1/history/{kind}/{name}?scope=s 对象应用史投影（v0.38 console 读面第三段：apply_log
//	                                     rv 降序；limit）
//	GET /v1/rev-content/{kind}/{name}?scope=s&rv=N 历史 rv 正文投影（v0.63 rev-content diff 面
//	                                     §1.40：git=裸库考古 / sqlite=doc_history 逐次账——source
//	                                     词形分立标注；face 缺席 503）
//	GET /v1/doc-history/{kind}/{name}?scope=s 逐次正文翻页清单（R-DOC-7 规则 4 doc_history 落地面：
//	                                     rev/翻页时刻/写者/字节数，不含正文——sqlite 后端装配；
//	                                     face 缺席 503）
//	GET /v1/context/index|find|relevance  上下文索引投影/TopK 检索/相关性调试面（v0.65 context
//	                                     P0 数据面——0–100 全局相关性；face 缺席 503）
//	GET /v1/trace/{kind}/{scope}/{name}   trace 1:N 双向聚合（v0.65 R-DOC-9；无 trace 段=空
//	                                     聚合 200 不 404；对象缺席才 404；face 缺席 503）
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
//	POST /v1/connectops/{name}/{op}     ops 调试面（R-TYPE-3：connectops-code-audit audit/findings；
//	                                    subject 权限门 fail-closed——域=ops/<name> × call，无效/无权
//	                                    subject=403 且 Invoke 不达；正式通路=workflow Verify 接线待
//	                                    接入——connectops_face.go §头注）
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
// docHistory 为 nil（逐次正文历史清单口面未接线——git 后端不装配）时 GET /v1/doc-history/*
// 返回 503（git 历史=周期导出快照非逐次账——面缺席诚实降级不冒充）。
// approval 为 nil（审批判定账面口面未接线）时 GET /v1/approvals 返回 503 且批准/驳回写钩子
// 零落账（账面缺席=无判定行，写路径语义不变——账面非权威，可重建）。
// metrics 未注入（数据面未装配）时 /v1/metrics 返回空投影（关断口径缺省可见——与 405
// 代际判别器分立：本面恒注册，装配与否只见投影丰瘠）。
func New(sch *scheme.Scheme, objects ObjectSource, applier Applier, drafts DraftStaging, triggers TriggerFace, runs RunFace, deliveries DeliveryFace, wakes WakeFace, publisher PublisherFace, opts ...Option) http.Handler {
	o := &options{}
	for _, f := range opts {
		f(o)
	}
	// logApproval 账面非权威落账：失败只注记 stderr，不翻转写路径结果（ops-log.md §2 v0.46）。
	logApproval := func(l opslog.ApprovalLog) {
		if o.approval == nil {
			return
		}
		if err := o.approval.AppendApproval(l); err != nil {
			log.Printf("apiserver: approval_log 落账失败（非权威，写路径不受影响）: %v", err)
		}
	}
	d := &deps{
		sch: sch, objects: objects, applier: applier, drafts: drafts,
		triggers: triggers, runs: runs, deliveries: deliveries, wakes: wakes,
		publisher: publisher, logApproval: logApproval, o: o,
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
	// 路由按域注册（R-ENG-7 拆分——handler 纯搬家至 routes_*.go，模式集与语义零变化；
	// ServeMux 按模式特异性匹配与注册顺序无关）。
	registerObjectRoutes(d, mux)   // objects / apply / secrets / drafts
	registerEventRoutes(d, mux)    // events / attention / inbox / sessions
	registerWorkflowRoutes(d, mux) // triggers / runs / deliveries / wakes / dispatches / workflows
	registerReadRoutes(d, mux)     // metrics / history / rev-content / doc-history / approvals / tasks / proposals / design_docs / intake / connectops / context
	registerStreamRoutes(d, mux)   // snapshot 聚合快照 + /v1/stream WebSocket（v0.66）
	registerSessionCtxRoutes(d, mux) // 会话上下文契约面（G-14——routes_sessions.go：GET/PUT/revert）
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

// listMetaOrList 元数据直读优先（G-15 读侧高性能——列表投影只需键/rv 时走 ListMeta：
// sqlite 后端不 SELECT content 列、零逐行 YAML 解析，正文仅单读路径 Get 拉；其余后端
// 无元数据面=回退 List 全量，行为兼容回退不改变行集语义）。可选能力接口（非 ObjectSource
// 硬约束）：ObjectSource 各实现按能力渐进接入，既有假体/装配零迁移。
func listMetaOrList(objects ObjectSource, ctx context.Context, kind, scope string) ([]*repository.Object, error) {
	if lm, ok := objects.(interface {
		ListMeta(ctx context.Context, kind, scope string) ([]*repository.Object, error)
	}); ok {
		return lm.ListMeta(ctx, kind, scope)
	}
	return objects.List(ctx, kind, scope)
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
