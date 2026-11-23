// cored Core 常驻进程（v0.3）：Scheme 载入自检 + 只读查询面 + 显式开启的 Apply 写入口。
// 设计依据：core_system_instance/impl/self-recursive-core.md（cored=冷启动入口与调试用途）
// 与 runtime/cored-bridge.md（plugin↔cored 桥：CORED_READY 就绪行 + /v1 查询面）。
// 部署锚点一律经参数/环境注入（--scheme/--docs/--ops-db / CORE_SYSTEM_*），不内置真实路径。
// 写面缺省关闭（缺省只读暴露）；--allow-apply 显式开启后 POST /v1/apply 是唯一写端点，
// 内部收敛 apply.Engine 三步闸门（scheme 准入=kind 注册+spec 结构校验→权限准入→乐观锁落库→DocApplied 事件）。
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	apiserver "github.com/friddle-self-recursive-agent/core_system_workservice/internal/apiserver"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/apply"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/draft"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventpub"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/outbound"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/patrol"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "cored:", err)
		os.Exit(1)
	}
}

func run() error {
	ctx := context.Background()
	schemePath := flag.String("scheme", os.Getenv("CORE_SYSTEM_SCHEME"),
		"core/scheme.yaml 路径（缺省读 CORE_SYSTEM_SCHEME 环境变量）")
	docsRoot := flag.String("docs", os.Getenv("CORE_SYSTEM_DOCS"),
		"文档库根目录（缺省读 CORE_SYSTEM_DOCS；为空则 /v1/objects 返回 503，scheme 面不受影响）")
	addr := flag.String("addr", os.Getenv("CORE_SYSTEM_ADDR"),
		"监听地址（缺省 127.0.0.1:0，即本机回环随机端口）")
	allowApply := flag.Bool("allow-apply", false,
		"显式开启写面（POST /v1/apply）；缺省 false=只读暴露。开启需同时配置 --docs")
	opsDB := flag.String("ops-db", "",
		"ops DB（SQLite）文件路径；为空=进程内存库（DB 非权威可重建，ops-log.md；持久化部署显式传路径）")
	outboundExec := flag.String("outbound-exec", "",
		"真实出站 exec 桥命令行（v0.17 进程内自动投递注入；空白分隔+引号段；空=关，送达止于受理投影。"+
			"实现侧出站命令由部署侧装配注入——依赖方向不反转，cored 不 import 连接器仓）")
	outboundTimeout := flag.Duration("outbound-timeout", 60*time.Second,
		"单次出站超时（--outbound-exec 注入时生效）")
	reconcileInterval := flag.Duration("reconcile-interval", 0,
		"定时巡检事件源间隔（v0.22 引擎内部源第二段，event.md 规则 10；0=关断零事件，缺省不改行为面；"+
			"开启需 Scheme 注册 ReconcileTick——注册表是引擎内部源唯一开关）")
	advanceInterval := flag.Duration("advance-interval", 0,
		"run 相位异步推进间隔（v0.26 相位推进器异步推进首段，workflow.md 规则 15；0=关断，缺省不改行为面——"+
			"受理首步同步推进不受影响。每拍对每现势 run 单步查表推进（stepByTable 复用）：落闸锚挂起、落终点锚终态、序列尾零动作停）")
	flag.Parse()

	if *schemePath == "" {
		return errors.New(`缺少 scheme 路径：用 --scheme <path> 或环境变量 CORE_SYSTEM_SCHEME`)
	}
	if *addr == "" {
		*addr = "127.0.0.1:0"
	}
	if *allowApply && *docsRoot == "" {
		return errors.New("--allow-apply 需要文档库：同时给 --docs（写面依附文档库 git 后端）")
	}

	// 启动自检：scheme 载入校验失败即退出（非 0），桥接方据退出码判定。
	sch, err := scheme.LoadFile(*schemePath)
	if err != nil {
		return err
	}
	sum := sch.Summary()
	fmt.Fprintf(os.Stderr, "cored: scheme %s/%s 载入通过（kinds=%d doctypes=%d eventtypes=%d dispatch=%d）来自 %s\n",
		sum.APIVersion, sum.Version, sum.KindCount, sum.DocTypeCount, sum.EventTypeCount, sum.DispatchCount, sch.Source)

	// ops DB：apply_log/event_log 双轨审计的 DB 侧（可丢弃可重建；git 才是权威）。
	store, err := opslog.OpenSQLite(*opsDB)
	if err != nil {
		return err
	}
	defer store.Close()
	if *opsDB == "" {
		fmt.Fprintf(os.Stderr, "cored: ops DB=进程内存（未配 --ops-db；重启即清，DB 非权威）\n")
	}

	// 文档对象读面 + 串行写队列：--docs 给定则打开 git 后端（野对象闸门=scheme 注册表）。
	var store2 apiserver.ObjectSource
	var repo *repository.GitRepo
	if *docsRoot != "" {
		repo, err = repository.NewGit(*docsRoot, sch, &repository.Options{Ops: store})
		if err != nil {
			return err
		}
		defer repo.Close()
		store2 = repo
		fmt.Fprintf(os.Stderr, "cored: 文档库已挂载 %s（git 后端，串行写+乐观锁）\n", repo.Root())
	}

	// 事件总线 + 写路径引擎（--allow-apply 才暴露；缺省只读）。
	bus := eventbus.NewBus(&eventbus.Options{Ops: store})
	// 触发骨架（dispatch 接线：第二段 v0.12 订阅面动态化 → 第三段 v0.13 引擎主体第一段
	// → 第四段 v0.14 引擎主体第二段：派发执行投影 → 第五段 v0.15 引擎主体第三段
	// 真实拉起第一子段：workflow 触发面接线 → 第六段 v0.16 引擎主体第四段
	// 真实拉起第二子段：notification 送达面接线 → 第四段收尾 v0.17：受理后真实通道出站挂接
	// → 第七段 v0.18 引擎主体第五段：agent 唤醒面接线）：
	// 订阅集合由 scheme.dispatch 表推导（on: 类型集合/通配 *→全类型一条；含通配坍缩），
	// 留痕携命中行投影 route/cc/for/owner（owner 字面量谓词随解析激活：involved scope 根
	// systems/<s>→<s>，global/builtin 同名）+ 执行绑定投影 executor/target/wired，
	// 全表未命中=dead-letter 兜底可见
	// （GET /v1/triggers；eventId/dedupKey 与 apply 回执/event_log 同源）；
	// 派发决策可查询（GET /v1/triggers/resolve：对表快照回答「若到达路由到哪、由谁执行」，
	// 零副作用；execution 投影=route/cc 目标执行面绑定，未知目标=unbound 缺口可见）。
	// 随总线装配、查询面恒开（写面关断时不产生事件，received=0 照常 200）。
	// workflow 触发面接线（v0.15）：run 投影落点挂接 + 执行面登记——route 目标执行面=workflow
	// 的留痕推进为一次对账 run 投影（对账锚→core-workflow；直接引用→按 builtin 注册表解析，
	// 未命中=平台对象引用可见；GET /v1/workflows/runs 查询）。run=触发已受理+声明起点，
	// 相位态机不推进（deferred）。
	// notification 送达面接线（v0.16）：送达投影落点挂接 + 执行面登记——执行面=notification
	// 的绑定（route 主路与 cc 旁路，origin 维度留痕）推进为一条送达投影
	// （GET /v1/notifications/deliveries 查询）。
	// 真实通道出站注入（v0.17）：--outbound-exec 给定时受理后异步出站（exec 桥→实现侧出站
	// 命令，信封协议 connectors.OutboundEnvelope），送达投影 status 翻转 accepted→sent/failed
	// （sent/failed 计数可查；一次尝试，重试/降级/静默 declared-deferred）。
	// agent 唤醒面接线（v0.18）：唤醒投影落点挂接 + 执行面登记——执行面=agent（`<p>-agent`
	// 族：Escalate→root-agent、HeartbeatLost→parent-agent）的绑定（route 主路与 cc 旁路，
	// origin 维度留痕）推进为一条唤醒投影（GET /v1/agents/wakes 查询）。唤醒投影出现 ≠
	// agent 会话已拉起（真实拉起=agent-session 域，declared-deferred）。
	// 相位态机→刹车→冷却→minLevel 仍 declared-deferred（cored-bridge.md §2.3）。
	dlOpts := &workflow.DeliveryOptions{}
	if *outboundExec != "" {
		argv, err := outbound.ParseCommand(*outboundExec)
		if err != nil {
			return fmt.Errorf("--outbound-exec 解析失败: %w", err)
		}
		bridge, err := outbound.NewCommand(argv)
		if err != nil {
			return fmt.Errorf("--outbound-exec 构建失败: %w", err)
		}
		dlOpts.Outbound = execOutbound{cmd: bridge}
		dlOpts.OutboundTimeout = *outboundTimeout
		fmt.Fprintf(os.Stderr, "cored: 真实出站已注入（exec 桥 argv=%d 令牌 超时=%s；送达受理后异步出站 status=accepted→sent/failed）\n",
			len(argv), *outboundTimeout)
	}
	// 对象 spec 解析接线（v0.27 表投影首段+v0.28 相位投影与迁移判定接线，workflows 规则
	// 16）：文档库挂载即注入对象 spec 解析面（object 源 run 相位序列/声明表按文档库 spec
	// 投影，零缓存=apply 落库即生效；相位投影即点亮既有表驱动判定——判定器零改动）。未挂
	// 文档库=能力缺席，既有行为面不变。
	runOpts := &workflow.RunOptions{}
	objSpecState := "off（未挂文档库——object 源表不投影）"
	if repo != nil {
		runOpts.Specs = workflow.DocsSpecs{Lookup: workflowDocLookup(repo)}
		objSpecState = "on"
	}
	runs := workflow.NewRunLog(runOpts)
	deliveries := workflow.NewDeliveryLog(dlOpts)
	wakes := workflow.NewWakeLog(nil)
	defer deliveries.Close() // 先于 stopTriggers 退场（LIFO）——停机期不再有新送达入队
	outboundState := "off（送达止于受理投影）"
	if dlOpts.Outbound != nil {
		outboundState = "on（exec 桥）"
	}
	// run 相位异步推进器（v0.26 相位推进器异步推进首段，workflow.md 规则 15）：装配层定时器
	// 每拍对每现势 run 单步查表推进（stepByTable 复用，判定器不动）——落闸锚挂起、落终点锚
	// 终态、序列尾零动作停；triggered/suspended/failed 零动作防御。interval=0 关断（缺省不改
	// 既有行为面——受理首步同步推进 v0.23 口径恒在；与巡检保险丝同款装配范式）。
	// v0.29 提前构建：事件驱动推进面的挂接（BindEventAdvance）须先于订阅主干 AttachDispatch
	// （同 Bind* 口径——事件开始流入后无接线窗口）；定时拍 goroutine 仍在下方启动。
	advancer := workflow.NewAdvancer(runs, *advanceInterval)
	triggers := workflow.NewTriggerHub(nil)
	triggers.BindRuns(runs)
	triggers.BindDeliveries(deliveries)
	triggers.BindAgents(wakes)
	triggers.WireExecutor(workflow.ExecWorkflow, workflow.ExecNotification, workflow.ExecAgent) // 机械判别器：resolve execution.wired false→true（三执行面）
	// 事件驱动推进源挂接（v0.29 事件拍——巡检拍与推进拍合一，workflow.md 规则 15 v0.29 段）：
	// **只在推进总闸开启时挂接**（--advance-interval>0）——总闸关断=事件拍缺席，事件驱动源
	// 不旁路总闸（推进=一个行为面一个总闸，关断语义单一可预期；巡检-only 部署行为面零翻转）。
	// 事件拍面=引擎内部源事件（source=core）落 run 的同拍全量单步推进（与定时拍同一查表落点）。
	if advancer.Enabled() {
		triggers.BindEventAdvance(runs)
	}
	stopTriggers := triggers.AttachDispatch(bus, sch.Spec.Dispatch)
	defer stopTriggers()
	// 定时巡检源（v0.22 引擎内部源第二段，event.md 规则 10）：装配层定时器到点自产
	// ReconcileTick 进总线（与注入/引擎内部事件同一条主干零旁路——窗口去重/订阅派发/
	// event_log 归档），route=workflows/core-workflow 推进为一次全量对账 run 投影。
	// 注册表闸门前置：未注册即不启动；interval=0 关断零事件（缺省行为面不变）。
	patrolState := "off"
	patrolTicker := patrol.New(bus, sch, *reconcileInterval)
	if patrolTicker.Enabled() {
		patrolState = "on(" + patrolTicker.Interval().String() + ")"
		patrolCtx, cancelPatrol := context.WithCancel(context.Background())
		defer cancelPatrol()
		go patrolTicker.Run(patrolCtx)
	} else if *reconcileInterval > 0 {
		patrolState = "unregistered" // interval>0 但类型未注册：闸门前置不启动（缺口可见）
	}
	fmt.Fprintf(os.Stderr, "cored: 触发骨架已按 dispatch 表接线（行数=%d 订阅集合=%v owner解析=v0.13 决策面=/v1/triggers/resolve 执行投影=v0.14 workflow触发面=已接线 对账run投影=/v1/workflows/runs builtin注册表=6项（相位词汇映射已落设计 v0.19）notification送达面=已接线 送达投影=/v1/notifications/deliveries agent唤醒面=已接线 唤醒投影=/v1/agents/wakes 真实出站=%s 派发执行余量=相位推进器（声明面已完备） 事件发布面=v0.20（注入与引擎内部事件同管线零旁路——event.md 规则 8） ApprovalNeeded 自产=v0.21（人审档拒绝路径引擎内部源——event.md 规则 9） 定时巡检源=v0.22（ReconcileTick 自产——event.md 规则 10））\n",
		len(sch.Spec.Dispatch), triggers.State().Types, outboundState)
	fmt.Fprintf(os.Stderr, "cored: 定时巡检源=%s（--reconcile-interval；on=到点自产 ReconcileTick source=core 拍点即因由 route=workflows/core-workflow；unregistered=Scheme 未注册该类型——注册表闸门前置）\n", patrolState)
	// run 相位异步推进器（v0.26）：构建已提前（事件拍挂接须先于订阅主干），此处只做
	// 定时拍启动与 stderr 装配锚报告。事件拍（v0.29）已随总闸在挂接点生效。
	advanceState := "off"
	if advancer.Enabled() {
		advanceState = "on(" + advancer.Interval().String() + ")"
		advCtx, cancelAdvance := context.WithCancel(context.Background())
		defer cancelAdvance()
		go advancer.Run(advCtx)
	}
	fmt.Fprintf(os.Stderr, "cored: run 推进器=%s（--advance-interval；on=每拍单步查表推进——落闸锚挂起/落终点锚终态/序列尾零动作停；off=关断缺省不改行为面，受理首步同步推进不受影响；事件拍=v0.29 引擎内部源落 run 同拍全量推进（巡检拍与推进拍合一——总闸同闸不旁路，source=core 三重闸））\n", advanceState)
	fmt.Fprintf(os.Stderr, "cored: 对象声明表投影=%s（对象 spec 解析 v0.27 表投影+v0.28 相位投影与迁移判定接线——object 源 run 相位/声明表按文档库 spec 投影，声明了 phases 的对象参与既有表驱动判定；未声明 phases=零投影既有行为面不变）\n", objSpecState)
	// 草稿缓存（apply.md §1 编辑与发布分离）：进程内存为底，读面恒开，写操作随写面开关。
	// --ops-db 给定=持久化部署：落稿/弃稿写穿 ops DB draft 表 + 启动恢复（ops-log.md §2）；
	// 缺省纯内存语义不变（重启即清，草稿可重放）。
	var draftOpts *draft.Options
	if *opsDB != "" {
		draftOpts = &draft.Options{Persistence: draftOpsStore{store}}
	}
	drafts := draft.NewStore(draftOpts)
	if draftOpts != nil {
		restored, skipped, err := drafts.Restore()
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "cored: 草稿恢复 restored=%d skipped=%d（--ops-db 持久化部署，draft 表）\n", restored, skipped)
	}
	var applier apiserver.Applier
	var publisher apiserver.PublisherFace
	if *allowApply {
		// 权限对象装载（v0.9 对象准入）：Role/Group 文档对象是授权面唯一来源——
		// 主体断言的 role 名必须对上 global/roles、global/groups 里的对象才产生权限。
		// 启动装载=权威快照（失败=拒绝启动，授权面宁拒不降级）；其后事件驱动热装载（见下）。
		perm := &permission.System{}
		objs, err := permDocs(ctx, repo)
		if err != nil {
			return err
		}
		if err := perm.LoadObjects(objs); err != nil {
			return fmt.Errorf("权限对象装载失败（授权面宁拒不降级）: %w", err)
		}
		snap := perm.Snapshot()
		fmt.Fprintf(os.Stderr, "cored: 权限对象装载 roles=%d groups=%d gen=%d（global/roles、global/groups）\n",
			len(snap.Roles), len(snap.Groups), snap.Gen)

		// 热装载（v0.10 免重启）：repository.Watch 变更信号（轮询 HEAD）→ 重投影 → 快照整体替换。
		// 失败语义（declared，cored-bridge.md §1）：保旧快照 + stderr 告警 + 定时重试自愈；
		// 不半装载、不打死运行中授权面（投影读与写面串行队列不互斥，暂态撕读/坏对象都不放大为停摆）。
		runCtx, cancelWatch := context.WithCancel(context.Background())
		defer cancelWatch()
		src, err := repo.Watch(runCtx)
		if err != nil {
			return fmt.Errorf("权限热装载订阅文档库变更: %w", err)
		}
		changes := make(chan struct{}, 1)
		go func() {
			defer close(changes)
			for range src { // sha 信号无载荷：连续变更合并，一次重投影装到最新面
				select {
				case changes <- struct{}{}:
				default:
				}
			}
		}()
		ref := &permission.Refresher{
			Sys:  perm,
			Load: func() ([]permission.ObjectDoc, error) { return permDocs(runCtx, repo) },
			Reloaded: func(res permission.RefreshResult) {
				if res.Ok {
					fmt.Fprintf(os.Stderr, "cored: 权限热装载 gen=%d roles=%d groups=%d（文档面变更→快照替换）\n",
						res.Gen, res.Roles, res.Groups)
				} else {
					fmt.Fprintf(os.Stderr, "cored: 权限热装载失败（保旧快照+定时重试自愈）: %v\n", res.Err)
				}
			},
		}
		go func() {
			if err := ref.Run(runCtx, changes); err != nil && !errors.Is(err, context.Canceled) {
				fmt.Fprintf(os.Stderr, "cored: 权限热装载异常退场: %v\n", err)
			}
		}()

		// 主体（v0.9 请求级）：HTTP 写面按请求解析（subject 字段/X-Core-Subject 头，匿名=照判拒绝）；
		// 进程级本地信任退化为引擎缺省 fallback（程序化调用，如 bootstrap/工作流自写）。
		engine := &apply.Engine{
			Sch: sch, Repo: repo, Bus: bus, Ops: store, Drafts: drafts,
			Perm:    perm,
			Subject: permission.Subject{User: "cored:local", Roles: []string{permission.DefaultAdmin}},
		}
		applier = engine
		// 事件注入面（v0.20 发布面首段，event.md 规则 8）：与 apply 写面同门控（--allow-apply），
		// 发布权限域与 apply 写权限域分立（publish × events 资源域）；注入走与引擎内部事件
		// 同一条总线主干（去重/订阅派发/event_log 归档零旁路）；source 服务侧判定恒 human。
		publisher = &eventpub.Publisher{Sch: sch, Bus: bus, Perm: perm}
		fmt.Fprintf(os.Stderr, "cored: apply 写面已开启（POST /v1/apply；请求级主体判定，匿名拒绝；本地信任仅作引擎 fallback）\n")
		fmt.Fprintf(os.Stderr, "cored: 事件注入面已开启（POST /v1/events；动作=publish × 域=events 与 apply 写面分立；注入通道 source=human——v0.20）\n")
	}

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		return fmt.Errorf("监听 %s: %w", *addr, err)
	}
	srv := &http.Server{Handler: apiserver.New(sch, store2, applier, drafts, triggers, runs, deliveries, wakes, publisher), ReadHeaderTimeout: 5 * time.Second}
	serveErr := make(chan error, 1)
	go func() { serveErr <- srv.Serve(ln) }()

	// 就绪行（stdout，机器可读；plugin 桥解析 addr 完成转发查询）。
	// apply=on|off 尾随令牌：写面状态机可读（旧桥解析宽松，不受尾随字段影响）。
	applyState := "off"
	if applier != nil {
		applyState = "on"
	}
	fmt.Printf("CORED_READY addr=%s scheme_version=%s kinds=%d apply=%s\n", ln.Addr(), sum.Version, sum.KindCount, applyState)

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	select {
	case s := <-sig:
		fmt.Fprintf(os.Stderr, "cored: 收到 %s，开始退场\n", s)
	case err := <-serveErr:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("http serve: %w", err)
		}
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		return fmt.Errorf("shutdown: %w", err)
	}
	return nil
}

// permDocs 投影权限对象（Role/Group，global scope——permission.md §1 作用域 Global）。
// permission 包不反向依赖存储层：这里把文档对象投影成它的装载输入。
// workflowDocLookup 对象声明表解析首段（v0.27）装配侧查找：按 workflow 锚名在文档库全
// scope 取 Workflow 对象原文（List 跨 scope 通配；global 同名优先——作用域覆盖语义的确定性
// 投影）。零缓存：每次现读（git 后端权威，apply 落库即生效）。未挂库/未命中=nil。
func workflowDocLookup(repo *repository.GitRepo) func(string) ([]byte, bool) {
	return func(name string) ([]byte, bool) {
		if repo == nil || name == "" {
			return nil, false
		}
		objs, err := repo.List(context.Background(), "Workflow", "")
		if err != nil {
			return nil, false
		}
		var pick *repository.Object
		for _, o := range objs {
			if o.Name != name {
				continue
			}
			if o.Scope == "global" {
				return o.Content, true // global 同名优先（确定性投影）
			}
			if pick == nil {
				pick = o
			}
		}
		if pick == nil {
			return nil, false
		}
		return pick.Content, true
	}
}

func permDocs(ctx context.Context, repo *repository.GitRepo) ([]permission.ObjectDoc, error) {
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
