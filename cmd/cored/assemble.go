// assemble.go —— cored 装配辅助件（R-ENG-7 行数收口纯搬家：样本通道适配/权限对象投影/
// workflow 声明解析查找/draft ops 持久化适配/旗标收集/环境时长/API 装配选项/自愈注入桥——
// 零语义变化，main.go 拆分伴生件）。
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/connectors"
	apiserver "github.com/friddle-self-recursive-agent/core_system_workservice/internal/apiserver"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/docrepo"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/draft"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/flowbind"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/metricsinfo"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/metricsrc"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/sessionliveness"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/watchsrc"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
	"github.com/friddle-self-recursive-agent/core_system_workservice/secretstore"
)

// metricsSamplePump 数据面→信号面样本通道适配（装配层持两者——metricsinfo 与 metricsrc
// 两包零互依，解耦口径的机械落点；字段直投不改造——窗口标签由样本侧声明服务侧透传同口径）。
func metricsSamplePump(in <-chan metricsinfo.Sample) <-chan metricsrc.Sample {
	out := make(chan metricsrc.Sample)
	go func() {
		for smp := range in { // 反压即丢由数据面扇出侧承担（订阅缓冲满即丢）——此处直投
			out <- metricsrc.Sample{Metric: smp.Metric, Value: smp.Value, Window: smp.Window}
		}
		close(out)
	}()
	return out
}

// permDocs 投影权限对象（Role/Group，global scope——permission.md §1 作用域 Global）。
// permission 包不反向依赖存储层：这里把文档对象投影成它的装载输入。
// workflowDocLookup 对象声明表解析首段（v0.27）装配侧查找：按 workflow 锚名在文档库全
// scope 取 Workflow 对象原文（List 跨 scope 通配；global 同名优先——作用域覆盖语义的确定性
// 投影）。零缓存：每次现读（git 后端权威，apply 落库即生效）。未挂库/未命中=nil。
func workflowDocLookup(repo repository.Repository) func(string) ([]byte, bool) {
	return func(name string) ([]byte, bool) {
		objs := workflowDocObjects(repo, name)
		if len(objs) == 0 {
			return nil, false
		}
		for _, o := range objs {
			if o.Scope == "global" {
				return o.Content, true // global 同名优先（确定性投影）
			}
		}
		return objs[0].Content, true // 无 global=首个对象行（List 序）
	}
}

// workflowEnabledResolver 规则开关判定面（v0.58b enabled honoring——E3 余量段收口）装配侧
// 注入：按 workflow 锚名取文档库对象行集 → workflow.EnabledOf 生效值纯映射（系统层 override
// 优先字典序首个 → global 本体 → 零对象=缺省启用——与 /v1/workflows 聚合面 effectiveObject
// 同 precedence 口径，判定面与聚合面无双源漂移）。零缓存=apply 落库即生效（Specs 同款热装载
// 语义天然跟随）。未命中=ok=false 放行（builtin 锚零对象=注册表缺省启用；缺口可见不静默拦）。
func workflowEnabledResolver(repo repository.Repository) workflow.EnabledResolver {
	return workflow.EnabledFunc(func(name string) (bool, bool) {
		objs := workflowDocObjects(repo, name)
		if len(objs) == 0 {
			return true, false
		}
		rows := make([]workflow.EnabledObject, 0, len(objs))
		for _, o := range objs {
			rows = append(rows, workflow.EnabledObject{Scope: o.Scope, Content: o.Content})
		}
		return workflow.EnabledOf(rows)
	})
}

// workflowDocObjects 按 workflow 锚名在文档库全 scope 取 Workflow 对象行集（enabled.go
// 判定面与既有 workflowDocLookup 共用的装配侧查找单点——List 跨 scope 通配，零缓存现读）。
func workflowDocObjects(repo repository.Repository, name string) []repository.Object {
	if repo == nil || name == "" {
		return nil
	}
	objs, err := repo.List(context.Background(), "Workflow", "")
	if err != nil {
		return nil
	}
	var out []repository.Object
	for _, o := range objs {
		if o.Name == name {
			out = append(out, *o)
		}
	}
	return out
}

func permDocs(ctx context.Context, repo repository.Repository) ([]permission.ObjectDoc, error) {
	if repo == nil {
		return nil, errors.New("权限对象装载需文档库（--allow-apply 依附 --docs）")
	}
	var out []permission.ObjectDoc
	for _, kind := range []string{"Role", "Group"} {
		objs, err := repo.List(ctx, kind, "global")
		if err != nil {
			return nil, fmt.Errorf("枚举 %s 对象: %w", kind, err)
		}
		for _, o := range objs {
			out = append(out, permission.ObjectDoc{Kind: kind, Name: o.Name, Scope: o.Scope, Content: o.Content})
		}
	}
	return out, nil
}

// draftOpsStore draft.Persistence 的 ops DB 实现（装配层适配器）：draft 与 opslog 两包
// 保持叶子解耦，字段映射只在此处发生（draft.Draft ↔ opslog.DraftRow）。
type draftOpsStore struct{ s *opslog.SQLiteStore }

func (a draftOpsStore) PutDraft(d *draft.Draft) error {
	return a.s.UpsertDraft(opslog.DraftRow{
		Kind: d.Kind, Scope: d.Scope, Name: d.Name, Content: d.Content,
		FieldManager: d.FieldManager, ReasonRef: d.ReasonRef, BaseRV: d.BaseRV,
		CreatedAt: d.CreatedAt, UpdatedAt: d.UpdatedAt,
	})
}

func (a draftOpsStore) DeleteDraft(kind, scope, name string) error {
	return a.s.DeleteDraft(kind, scope, name)
}

func (a draftOpsStore) ListDrafts() ([]*draft.Draft, error) {
	rows, err := a.s.ListDrafts()
	if err != nil {
		return nil, err
	}
	out := make([]*draft.Draft, 0, len(rows))
	for _, r := range rows {
		out = append(out, &draft.Draft{
			Kind: r.Kind, Scope: r.Scope, Name: r.Name, Content: r.Content,
			FieldManager: r.FieldManager, ReasonRef: r.ReasonRef, BaseRV: r.BaseRV,
			CreatedAt: r.CreatedAt, UpdatedAt: r.UpdatedAt,
		})
	}
	return out, nil
}

// repoSpecs 可重复旗标收集（--design-repo name=<仓根>）。
type repoSpecs []string

// envDuration 读环境变量时长（部署锚点经环境注入的旗标缺省面——CORE_SYSTEM_HEARTBEAT_
// INTERVAL 同款口径）：空/未设=缺省值；不可解析=0（启动校验 envDurationValid 另行显式
// 报错——配置错误可见不静默，不在旗标缺省面静默吞）。
func envDuration(key string, def time.Duration) time.Duration {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return def
	}
	if d, err := time.ParseDuration(raw); err == nil {
		return d
	}
	return 0
}

// envDurationValid 环境变量时长词形前置校验（空/未设=合法；不可解析=false——调用方拒绝
// 启动，与 --flag 解析失败同口径的可见失败）。
func envDurationValid(key string) bool {
	raw := strings.TrimSpace(os.Getenv(key))
	if raw == "" {
		return true
	}
	_, err := time.ParseDuration(raw)
	return err == nil
}

func (r *repoSpecs) String() string     { return strings.Join(*r, ",") }
func (r *repoSpecs) Set(v string) error { *r = append(*r, v); return nil }

// apiserverOpts 可选能力装配（九参构建签名零迁移；空=零可选面零行为翻转）。
func apiserverOpts(metricsFace apiserver.MetricsFace, secrets *secretstore.Store, store *opslog.SQLiteStore, dispatch *flowbind.DispatchLog, approvals *flowbind.SyncApprovals, designDocs *docrepo.Set) []apiserver.Option {
	var opts []apiserver.Option
	if metricsFace != nil {
		opts = append(opts, apiserver.WithMetrics(metricsFace))
	}
	if secrets != nil {
		opts = append(opts, apiserver.WithSecrets(secrets))
	}
	// 事件归档查询面（v0.36 console 读面首段 GET /v1/events）：ops DB 恒装配（含内存库
	// ——DB 非权威可重建，归档读面照常成立），恒注入不缺位。
	if store != nil {
		opts = append(opts, apiserver.WithEventLog(store))
	}
	// 应用史查询面（v0.38 console 读面第三段 GET /v1/history）：ops DB 恒装配（同上口径），
	// 恒注入不缺位。
	if store != nil {
		opts = append(opts, apiserver.WithApplyLog(store))
	}
	// 审批判定账面（v0.46 审批独立账：批准/驳回写钩子 + GET /v1/approvals 读面）：ops DB
	// 恒装配（同上口径），恒注入不缺位。
	if store != nil {
		opts = append(opts, apiserver.WithApprovalLog(store))
	}
	// 派单受理投影面（v0.50 R-8，GET /v1/agents/dispatches 读面）：随 flowbind 写面装配
	// （--allow-apply；headless 载体=零物化零派单，face 缺席 503 防御面）。
	if dispatch != nil {
		opts = append(opts, apiserver.WithDispatchFace(dispatch))
	}
	// 通知中心收件箱面（v0.56 R-10，GET /v1/inbox 读面）：随批准投影面装配（--allow-apply；
	// face 缺席 503 防御面）——四类判定分类自读面现读组装，console 零自建状态。
	if approvals != nil {
		opts = append(opts, apiserver.WithInboxFace(approvals))
	}
	// 设计文档面（v0.52 能力 A，/v1/design-docs* 读/史/草稿/apply）：随 --design-repo 装配
	// （写面随 --allow-apply；未注册=面缺席 503）。
	if designDocs != nil {
		opts = append(opts, apiserver.WithDesignDocs(designDocs))
	}
	return opts
}

// selfHealHTTPBridge 自愈注入桥的 HTTP 形（v0.56 R-10——cored watchdog → dsh 宿主 plugin 侧
// 会话推进面）。cored 侧零宿主依赖：桥=loopback POST {addr}/selfheal/bounce 载荷
// {sessionId, word, source:"cored-watchdog"}；宿主 plugin 装配段启 listener（127.0.0.1，
// env 同址回填 CORE_SYSTEM_SELFHEAL_ADDR）——经 sessionController.prompt(mode=queue) 注入
// 固定 prompt「继续」（红线：只对已判停滞会话注入固定词，不新增任何自动任务/自动执行面）。
// 超时=err 返回（watchdog 侧照实上报绝不静默——下一拍重试第一动作）。
type selfHealHTTPBridge struct {
	addr string
	word string
}

func newSelfHealHTTPBridge(addr, word string) *selfHealHTTPBridge {
	return &selfHealHTTPBridge{addr: addr, word: word}
}

func (b *selfHealHTTPBridge) Bounce(sessionID string) error {
	payload, err := json.Marshal(map[string]any{
		"sessionId": sessionID, "word": b.word, "source": "cored-watchdog",
	})
	if err != nil {
		return err
	}
	req, err := http.NewRequest(http.MethodPost, "http://"+b.addr+"/selfheal/bounce", bytes.NewReader(payload))
	if err != nil {
		return err
	}
	req.Header.Set("content-type", "application/json")
	cl := &http.Client{Timeout: 10 * time.Second}
	resp, err := cl.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 8<<10))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("selfheal bridge HTTP %d: %s", resp.StatusCode, strings.TrimSpace(string(body)))
	}
	return nil
}

func (b *selfHealHTTPBridge) InjectWord() string { return b.word }

// maybeStartDocsExporter sqlite→git 定时导出器装配（R-DOC-7 规则 2：git=导出/公开
// output/public——周期快照非逐次账，逐次正文账=doc_history）。env 三键（部署侧配置，
// R-PUB-2——键名入库、实际值不入）：CORE_SYSTEM_DOCS_EXPORT_INTERVAL 控制开关与间隔
// （如 30m；未设=不启动，不可解析/非正=缺口 stderr 可见不启动）；CORE_SYSTEM_DOCS_EXPORT_DIR
// 必填才启动（目标工作树）；CORE_SYSTEM_DOCS_EXPORT_REMOTE 设了才 push（no force-push，
// 认证走部署侧凭据不入库）。仅 sqlite 后端装配（git 后端自身已是布局权威，无导出语义）。
// 导出=只读派生任务：每拍全量快照按对象规范路径写目标树→有变化才 commit→可选 push；失败只
// 告警不回滚 sqlite 写、不崩 cored（下拍重试自愈）；绝不回导（回导=--docs 显式一次性迁移通道）。
func maybeStartDocsExporter(repo repository.Repository, sch *scheme.Scheme) {
	if _, ok := repo.(*repository.SQLiteRepo); !ok {
		return
	}
	iv := strings.TrimSpace(os.Getenv("CORE_SYSTEM_DOCS_EXPORT_INTERVAL"))
	if iv == "" {
		return
	}
	d, perr := time.ParseDuration(iv)
	switch {
	case perr != nil || d <= 0:
		fmt.Fprintf(os.Stderr, "cored: docs 导出器=off（CORE_SYSTEM_DOCS_EXPORT_INTERVAL=%q 不可解析/非正——缺口可见不启动）\n", iv)
	case os.Getenv("CORE_SYSTEM_DOCS_EXPORT_DIR") == "":
		fmt.Fprintf(os.Stderr, "cored: docs 导出器=off（interval=%s 已设但 CORE_SYSTEM_DOCS_EXPORT_DIR 缺席——必填才启动）\n", d)
	default:
		exp, err := repository.NewExporter(repo, sch, os.Getenv("CORE_SYSTEM_DOCS_EXPORT_DIR"), os.Getenv("CORE_SYSTEM_DOCS_EXPORT_REMOTE"))
		if err != nil {
			fmt.Fprintf(os.Stderr, "cored: docs 导出器装配失败（读面不受影响——导出是只读派生任务）: %v\n", err)
			return
		}
		expCtx, cancelExp := context.WithCancel(context.Background())
		defer cancelExp()
		go exp.Run(expCtx, d)
		pushState := "off（本地 commit 止步）"
		if os.Getenv("CORE_SYSTEM_DOCS_EXPORT_REMOTE") != "" {
			pushState = "on（push " + os.Getenv("CORE_SYSTEM_DOCS_EXPORT_REMOTE") + "）"
		}
		fmt.Fprintf(os.Stderr, "cored: docs 导出器=on(%s → %s push=%s)（sqlite→git 周期快照——git 历史=导出快照非逐次账；失败只告警不回滚不崩；绝不回导）\n",
			d, os.Getenv("CORE_SYSTEM_DOCS_EXPORT_DIR"), pushState)
	}
}

// setupSessionWatchdog 会话活性 watchdog 装配（KI-003，event.md 规则 14/14b 引擎内部源第六段；
// agent-session.md §7）：装配层定时扫描 ops DB session_heartbeat——超阈值无心跳自产
// SessionStalled（判定态词=SuspectedStall；dedupKey=session:<id>）→自愈固定顺序（第一动作=
// 注入固定 prompt「继续」，恢复记 SessionHealed）→诊断分诊（权限类 ApprovalNeeded 向上提交/
// 人工类 bind 通道上报、未配置=零通知零外发/长任务 LongTaskRunning 显式汇报）→连续无效达
// --heal-retry-limit 合成 Escalate 停自动处置。注册表闸门前置（SessionStalled 未注册=
// unregistered 不启动）；interval=0 关断零事件零扫描（缺省不改行为面）。与注入面同一条主干
// 零旁路。SelfHeal 真实接线（v0.56 R-10）：桥=dsh 宿主 plugin 侧 sessionController.prompt 面，
// cored 经 env（CORE_SYSTEM_SELFHEAL_ADDR——宿主 loopback HTTP 面）回填；桥缺席（env 未设）=
// 纯观测装配词形不变；注入失败=err 返回照实上报绝不静默；会话侧回执心跳（source=session
// phase=inject-continue）非续租——恢复判据=会话侧真实推进心跳。返回 watchdog 与装配锚词形。
func setupSessionWatchdog(store *opslog.SQLiteStore, bus *eventbus.Bus, sch *scheme.Scheme,
	interval, threshold time.Duration, healLimit int, selfHealWord string) (*sessionliveness.Watchdog, string) {
	heartbeatStore := struct {
		sessionliveness.Registry
	}{store}
	liveness := sessionliveness.New(heartbeatStore, bus, sch, interval, threshold, healLimit)
	if addr := strings.TrimSpace(os.Getenv("CORE_SYSTEM_SELFHEAL_ADDR")); addr != "" {
		liveness.SetSelfHealBridge(newSelfHealHTTPBridge(addr, selfHealWord))
		fmt.Fprintf(os.Stderr, "cored: 自愈注入桥=on(%s)（CORE_SYSTEM_SELFHEAL_ADDR——第一动作经 dsh 宿主 sessionController.prompt 注入「%s」，失败上报绝不静默；桥缺席=纯观测装配）\n", addr, selfHealWord)
	}
	liveness.OnEvent(func(typ, session string, err error) {
		if err != nil {
			fmt.Fprintf(os.Stderr, "cored: sessionliveness 会话 %s 合成事件 %s 发布失败（继续上报不静默）: %v\n", session, typ, err)
			return
		}
		fmt.Fprintf(os.Stderr, "cored: sessionliveness 会话 %s 合成事件 %s 已入总线\n", session, typ)
	})
	state := "off"
	switch {
	case interval <= 0:
		// 缺省关断
	case !liveness.Enabled():
		state = "unregistered" // interval>0 但 SessionStalled 未注册：闸门前置不启动（缺口可见）
	default:
		state = "on(" + liveness.Interval().String() + ",threshold=" + liveness.Threshold().String() + ")"
		liveCtx, cancelLive := context.WithCancel(context.Background())
		defer cancelLive()
		go liveness.Run(liveCtx)
	}
	fmt.Fprintf(os.Stderr, "cored: 会话活性 watchdog=%s（--session-watch-interval；on=停滞判定 SuspectedStall→注入继续→SessionHealed→诊断分诊→升级 Escalate 全链事件化；unregistered=SessionStalled 未注册——注册表闸门前置；off=关断零事件缺省不改行为面；心跳注册面=POST /v1/sessions/heartbeat 查询面=GET /v1/sessions）\n", state)
	return liveness, state
}

// startWatchSource connector watch 事件桥装配（v0.30 引擎内部源第三段，event.md 规则 12）：
// 装配层枚举注册表内 Watch 契约驱动（connectors.Providers——依赖方向不反转：驱动住连接器仓
// 经契约层注册，cored 只消费契约通道），把驱动侧外部变化通道桥接进总线（source 服务侧判定恒
// connector；类型注册闸门前置；level=注册表缺省投影；dedupKey=驱动侧因由声明）。旗标关断
// （缺省）=零驱动零事件，缺省不改行为面；watch 事件落 run 同拍即拍（workflow.md 规则 15
// v0.30 段——拍面驱动源集合扩展，零新开关）。返回装配锚词形。
func startWatchSource(flagOn bool, bus *eventbus.Bus, sch *scheme.Scheme) string {
	state := "off"
	if !flagOn {
		return state
	}
	if _, err := sch.LookupEventType(watchsrc.EventType); err != nil {
		return "unregistered" // 注册表闸门前置：类型未注册不启动（缺口可见）
	}
	watchCtx, cancelWatchSrc := context.WithCancel(context.Background())
	defer cancelWatchSrc()
	var watchSrcs []watchsrc.Source
	for _, p := range connectors.Providers() { // Type 排序快照——装配顺序确定性
		if !p.Contract().Watch {
			continue
		}
		ch, err := p.Watch(watchCtx)
		if err != nil {
			fmt.Fprintf(os.Stderr, "cored: watch 驱动 %s 订阅失败（跳过该源）: %v\n", p.Type(), err)
			continue
		}
		watchSrcs = append(watchSrcs, watchsrc.Source{Name: p.Type(), Involved: "WorkService/" + p.Type(), Events: ch})
	}
	if len(watchSrcs) == 0 {
		return "no-driver" // 已开启∧类型已注册∧注册表无 Watch 契约驱动（缺口可见）
	}
	go watchsrc.New(bus, sch, watchSrcs).Run(watchCtx)
	return fmt.Sprintf("on(%d)", len(watchSrcs))
}

// startMetricsPlane metrics 阈值事件桥（v0.31 引擎内部源第四段，event.md 规则 13）+
// MetricsInfo 数据面首段（v0.33，event.md 规则 13 数据面落点）装配：数据面=进程内样本库
// （store/queries 投影，可丢弃可重建）+自观测采样器（ops DB event_log 流水窗口计数——查询
// 投影即样本）+读面 /v1/metrics；信号面=阈值桥（契约不变）——解耦口径：数据面零信号面依赖，
// 信号面不持数据面，装配层持两者经样本通道适配。阈值声明面=文档库
// WorkService.observation.metrics 告警面（v0.33 随行件：判定谓词落声明面——服务侧桥只判定
// 不发明；逐对象 involved=WorkService/<name> 源级绑定锚）。旗标关断（缺省）=零样本源零事件，
// 缺省不改行为面；metrics 事件落 run 同拍即拍（workflow.md 规则 15——拍面驱动源集合扩展，
// 零新开关）。返回（信号面锚词形/数据面锚词形/样本库/读面）。
func startMetricsPlane(flagOn bool, window time.Duration, bus *eventbus.Bus, sch *scheme.Scheme,
	repo repository.Repository, store *opslog.SQLiteStore) (string, string, *metricsinfo.Store, apiserver.MetricsFace, error) {
	metricsState := "off"
	metricsPlaneState := "off"
	var metricsStore *metricsinfo.Store
	var metricsFace apiserver.MetricsFace
	if !flagOn {
		return metricsState, metricsPlaneState, metricsStore, metricsFace, nil
	}
	metricsStore = metricsinfo.NewStore()
	metricsFace = metricsStore
	// 数据面真实生产者首段：装配层自观测采样器（ops DB event_log 流水窗口计数）。
	sampler := metricsinfo.NewSampler(metricsStore, window, store.EventCountsSince)
	if sampler.Enabled() {
		metricsPlaneState = "on(" + sampler.Window().String() + ")"
		metricsCtx, cancelMetricsPlane := context.WithCancel(context.Background())
		defer cancelMetricsPlane()
		go sampler.Run(metricsCtx)
	} else {
		metricsPlaneState = "off（窗口<=0）"
	}
	if _, err := sch.LookupEventType(metricsrc.EventType); err != nil {
		// 注册表闸门前置：类型未注册不启动（缺口可见）
		return "unregistered", metricsPlaneState, metricsStore, metricsFace, nil
	}
	// 阈值声明面绑定（v0.33 首段；v0.45 装配段收编 BindWorkservice 一步投影；v0.47 起
	// 装配期枚举只是初值——apply 事件驱动热重绑，声明面落库即生效免重启）：文档库 WorkService
	// 对象 observation.metrics 告警面 → 桥阈值绑定；逐对象 involved=WorkService/<scope>/<name>
	//（involvedRef kind/scope/name 三段同构——owner/scope 谓词与 console 事件详情对象链可解析）。
	// 零声明/身份不完整=零绑定初值（合法等待态——apply 热激活，no-source 终态退场）。
	enumMetrics := func() ([]metricsrc.Source, error) {
		if repo == nil {
			return nil, nil
		}
		wss, err := repo.List(context.Background(), "WorkService", "")
		if err != nil {
			return nil, err
		}
		var srcs []metricsrc.Source
		for _, o := range wss {
			if src, ok := metricsrc.BindWorkservice(o.Scope, o.Name, o.Content); ok {
				srcs = append(srcs, src)
			}
		}
		return srcs, nil // 零告警面声明对象零绑定（等待态口径：判定静默，apply 热激活）
	}
	metricsSrcs, err := enumMetrics()
	if err != nil {
		return "", "", nil, nil, fmt.Errorf("枚举 WorkService 阈值声明面: %w", err)
	}
	metricsState = fmt.Sprintf("on(%d)", len(metricsSrcs))
	metricsCtx, cancelMetricsSrc := context.WithCancel(context.Background())
	defer cancelMetricsSrc()
	// 桥级单样本流（v0.47：绑定表与样本流解耦——重绑只换声明不碰流，零通道抖动）；
	// 零绑定初值也照常挂流运行（等待态可热激活）。
	metricsBridge := metricsrc.New(bus, sch, metricsSamplePump(metricsStore.Subscribe()), metricsSrcs)
	go metricsBridge.Run(metricsCtx)
	metricsrc.StartRebind(metricsCtx, bus, metricsBridge, enumMetrics, metricsrc.RebindHooks{
		Applied: func(n int) {
			fmt.Fprintf(os.Stderr, "cored: metrics 阈值热重绑 n=%d（DocApplied×WorkService→绑定表替换——v0.47 声明面落库即生效免重启）\n", n)
		},
		Failed: func(err error) {
			fmt.Fprintf(os.Stderr, "cored: metrics 阈值热重绑失败（保旧表+下次 apply 重试自愈）: %v\n", err)
		},
	})
	return metricsState, metricsPlaneState, metricsStore, metricsFace, nil
}
