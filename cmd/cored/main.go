// cored Core 常驻进程（v0.3）：Scheme 载入自检 + 只读查询面 + 显式开启的 Apply 写入口。
// 设计依据：core_system_instance/impl/self-recursive-core.md（cored=冷启动入口与调试用途）
// 与 runtime/cored-bridge.md（plugin↔cored 桥：CORED_READY 就绪行 + /v1 查询面）。
// 部署锚点一律经参数/环境注入（--scheme/--docs/--ops-db / CORE_SYSTEM_*），不内置真实路径。
// 写面缺省关闭（缺省只读暴露）；--allow-apply 显式开启后 POST /v1/apply 是唯一写端点，
// 内部收敛 apply.Engine 三步闸门（scheme 准入=kind 注册+spec 结构校验→权限准入→乐观锁落库→DocApplied 事件）。
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/connectors"
	apiserver "github.com/friddle-self-recursive-agent/core_system_workservice/internal/apiserver"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/apply"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/docrepo"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/draft"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventpub"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/flowbind"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/metricsinfo"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/metricsrc"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/outbound"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/patrol"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/proposal"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheduler"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/sessionliveness"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/watchsrc"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/workflow"
	"github.com/friddle-self-recursive-agent/core_system_workservice/secretstore"
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
		"文档库根目录（缺省读 CORE_SYSTEM_DOCS；sqlite 后端=种子导入/导出格式来源（空库一次性导入，非空库不导入）；"+
			"git 后端=兼容权威布局根（需显式 --docs-backend git））")
	docsBackend := flag.String("docs-backend", "sqlite",
		"文档存储后端（git|sqlite；缺省 sqlite=root 默认）。sqlite=docs DB 为权威内容面（--docs-db 必配），"+
			"git 布局降级为导出/公开格式（定时导出+push 远程；--docs 给定时作空库种子导入）；"+
			"git=兼容后端（既有行为面，非 root 部署显式选择）")
	docsDB := flag.String("docs-db", "",
		"docs DB（SQLite）文件路径（--docs-backend sqlite 必配；文档内容权威，不可丢弃——与 ops DB "+
			"可丢弃口径分立，秘密库再分离，见 --secrets-db）")
	addr := flag.String("addr", os.Getenv("CORE_SYSTEM_ADDR"),
		"监听地址（缺省 127.0.0.1:0，即本机回环随机端口）")
	allowApply := flag.Bool("allow-apply", false,
		"显式开启写面（POST /v1/apply）；缺省 false=只读暴露。git 后端需同时配置 --docs；sqlite 后端只需 --docs-db")
	opsDB := flag.String("ops-db", "",
		"ops DB（SQLite）文件路径；为空=进程内存库（DB 非权威可重建，ops-log.md；持久化部署显式传路径）")
	secretsDB := flag.String("secrets-db", "",
		"秘密库（SQLite）文件路径（v0.34 secret 存储落地首段；为空=秘密面关断。凭据真实值只落本库"+
			"（0600 权限独立文件——秘密=权威运行时值非缓存，不随可丢弃的 ops DB 走）；公开仓只见引用，"+
			"任何查询面只出名称零值出（脱敏红线，provider.md 规则 2/6）")
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
	watchSource := flag.Bool("watch-source", false,
		"connector watch 事件桥（v0.30 引擎内部源第三段，event.md 规则 12；缺省 false=关断零事件不改行为面。"+
			"开启时装配层枚举注册表内 Watch 契约驱动，把驱动侧外部变化通道桥接进总线——source 服务侧判定恒 connector，"+
			"watch 事件落 run 同拍即拍（workflow.md 规则 15 v0.30 段，零新开关）")
	metricsSource := flag.Bool("metrics-source", false,
		"metrics 阈值事件桥（v0.31 引擎内部源第四段，event.md 规则 13；缺省 false=关断零事件不改行为面。"+
			"开启时装配层桥接数据面指标样本通道并按声明面阈值绑定做服务侧判定，越限自产 MetricThreshold 进总线——"+
			"source 服务侧判定恒 metrics，metrics 事件落 run 同拍即拍（workflow.md 规则 15，零新开关）；"+
			"阈值声明面=文档库 WorkService.observation.metrics（v0.33 数据面首段随行件；v0.47 起 apply 事件驱动热重绑——声明面落库即生效免重启，零声明=on(0) 等待态 apply 后热激活）")
	metricsWindow := flag.Duration("metrics-window", 30*time.Second,
		"MetricsInfo 数据面采样窗口（v0.33 数据面首段；随 --metrics-source 生效。装配层自观测采样器"+
			"每窗口查询 ops DB event_log 流水做窗口计数入样本库——查询投影即样本；<=0=数据面关断零样本）")
	brakesFlag := flag.Bool("brakes", false,
		"刹车闸（v0.32 刹车首段，workflow.md 规则 17；缺省 false=关断零刹车不改行为面。"+
			"开启时两路生效：受理路径冷却门（同锚同 involved cooldown 窗口内触发不落 run，braked 计数可见；"+
			"跨 involved 不互踩——防震荡语义粒度=同一对象）+推进拍变更上限（每拍每锚 run 相位变更数受生效刹车值约束，"+
			"capped 计数可见；超限本拍零动作下拍续推）。刹车值=spec.brakes 声明收紧（宽松方向钳回强制默认——"+
			"可收紧不可关闭）；文档库未挂载=全锚强制默认）")
	notifyDedupWindow := flag.Duration("notify-dedup-window", 10*time.Minute,
		"通知去重窗口（v0.35 KI-001 通知段——dedup.window 落地；缺省 10m=notification.md 声明面缺省。"+
			"窗口内同 dedupKey 至多一封送达受理，其余 deduped 计数可查询不投影（同因聚合防轰炸）；"+
			"<=0=不去重。判别器：同 dedupKey 窗口内 N 次注入 → 实发 1、deduped=N-1）")
	notifyMinLevel := flag.String("notify-min-level", "",
		"通知等级下限（v0.35 KI-001 通知段——minLevel 落地；缺省空=不过滤。debug|info|warn|error|fatal，"+
			"低于下限的事件零受理零投影 filtered 计数可查询；声明面=Notification on.levels/severity 过滤）")
	proposalAutoApprove := flag.String("proposal-auto-approve", "",
		"Proposal 低危自动批准分流（v0.51 能力 B，concepts/proposal.md 规则 3；缺省空=全人审宁慢勿越。"+
			"priority 逗号清单如 low——命中清单的提案提交即自动批准（引擎自写 fieldManager=proposal:auto，"+
			"approval_log granted 行 judgedBy=引擎主体）+派单照常；未命中=人审（提交自产 ApprovalNeeded——"+
			"event.md 规则 15 source=core）。批准只产派单受理投影，不 enable 任何 workflow 不自动执行代码）")
	sessionWatchInterval := flag.Duration("session-watch-interval", 0,
		"会话活性 watchdog 扫描拍间隔（KI-003，event.md 规则 14 引擎内部源第六段；0=关断零事件零扫描，"+
			"缺省不改行为面。开启需 Scheme 注册 SessionStalled——注册表是引擎内部源唯一开关。停滞判定→自愈固定顺序"+
			"（第一动作=注入固定 prompt「继续」→SessionHealed）→诊断分诊（权限类 ApprovalNeeded 向上提交/人工类按 "+
			"bind 通道上报未配置=零通知零外发/长任务 LongTaskRunning 显式汇报）→连续无效达 --heal-retry-limit 升级 Escalate）")
	stallThreshold := flag.Duration("stall-threshold", sessionliveness.DefaultStallThreshold,
		"会话停滞判定阈值（KI-003；缺省 10m=event.md 规则 14② 声明缺省。updated_at 早于 now-阈值=停滞候选）")
	healRetryLimit := flag.Int("heal-retry-limit", sessionliveness.DefaultHealRetryLimit,
		"同会话连续自愈无效升级阈值（KI-003；缺省 2=event.md 规则 14⑤——达阈值合成 Escalate 停自动处置升级，不静默重试）")
	// SelfHeal 真实接线（v0.56 R-10——agent-session.md §7 自愈第一动作执行面：dsh 宿主侧
	// 装配，cored 只持桥口面+注入词声明；旗标=注入词单点声明，桥回填走插件装配段 env 旁路）。
	selfHealWord := flag.String("self-heal-word", sessionliveness.DefaultHealWord,
		"自愈第一动作注入词（v0.56；缺省「继续」——agent-session.md §7 固定 prompt 单点声明。"+
			"真实注入执行=dsh 宿主 plugin 侧 sessionController.prompt 面（桥回填）；nil 桥=纯观测装配词形不变）")
	var designRepos repoSpecs
	flag.Var(&designRepos, "design-repo",
		"设计文档白名单仓（v0.52 能力 A——console 屏幕 6 直接编辑设计文档 md → Apply → git commit；"+
			"形 name=<仓工作树根>，可重复。注册仓的 *.md 进读面（权威=git HEAD）与写面（白名单→R-PUB-1 lint→"+
			"乐观锁 baseSha 三检查 + 权限准入 + git commit 本地落盘不 push）。缺省不注册=设计文档面 503（面缺席）")
	flag.Parse()

	if *schemePath == "" {
		return errors.New(`缺少 scheme 路径：用 --scheme <path> 或环境变量 CORE_SYSTEM_SCHEME`)
	}
	if *addr == "" {
		*addr = "127.0.0.1:0"
	}
	if *allowApply && *docsRoot == "" && *docsBackend != "sqlite" {
		return errors.New("--allow-apply 需要文档库：git 后端同时给 --docs（布局即权威）；sqlite 后端给 --docs-db 即可")
	}
	if *docsBackend != "git" && *docsBackend != "sqlite" {
		return fmt.Errorf("--docs-backend 只允许 git|sqlite（实得 %q）", *docsBackend)
	}
	if *docsBackend == "sqlite" && *docsDB == "" {
		return errors.New("--docs-backend sqlite 需要 docs DB：同时给 --docs-db（文档内容权威，不可丢内存）")
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

	// 秘密库（v0.34 secret 存储落地首段）：凭据真实值的部署侧权威面（--secrets-db 独立
	// 文件 0600——秘密不随可丢弃的 ops DB 走）。--secrets-db 给定时装配 API 面（列表只出
	// 名称零值出）与引用式解析锚（CORE_SYSTEM_SECRETS_DB 经出站桥 env 传给实现侧命令——
	// 凭据值本身永不跨进程边界）。空=关断：secret: 来源解析缺口可见（引用面照常声明）。
	var secrets *secretstore.Store
	secretsState := "off"
	if *secretsDB != "" {
		secrets, err = secretstore.OpenSQLite(*secretsDB)
		if err != nil {
			return err
		}
		defer secrets.Close()
		secretsState = "on(" + secrets.Path() + ")"
	}
	fmt.Fprintf(os.Stderr, "cored: 秘密库=%s（--secrets-db；on=凭据值部署侧权威面+脱敏 API 面 /v1/secrets 零值出；off=secret: 来源解析缺口可见——引用面照常声明）\n", secretsState)

	// 文档对象读面 + 串行写队列：--docs/--docs-db 给定则打开文档库后端（野对象闸门=scheme 注册表）。
	// --docs-backend 选择权威后端（缺省 sqlite=root 默认）：sqlite=docs DB 权威内容面
	// （--docs 旧布局根作空库种子导入来源，保留 rv；非空库不导入=幂等）；git=兼容后端（显式选择）。
	var store2 apiserver.ObjectSource
	var repo repository.Repository
	switch {
	case *docsBackend == "sqlite":
		repo, err = repository.OpenSQLite(*docsDB, sch, &repository.Options{Ops: store}, *docsRoot)
		if err != nil {
			return err
		}
		defer repo.Close()
		store2 = repo
		fmt.Fprintf(os.Stderr, "cored: 文档库已挂载 db=%s（sqlite 后端，串行写+乐观锁；种子导入来源=%q 空库一次性导入）\n",
			*docsDB, *docsRoot)
	case *docsRoot != "":
		gr, err := repository.NewGit(*docsRoot, sch, &repository.Options{Ops: store})
		if err != nil {
			return err
		}
		repo = gr
		defer repo.Close()
		store2 = repo
		fmt.Fprintf(os.Stderr, "cored: 文档库已挂载 %s（git 后端，串行写+乐观锁）\n", gr.Root())
	}

	// sqlite→git 定时导出器（R-DOC-7 规则 2：git=导出/公开 output/public——周期快照非逐次
	// 账，逐次正文账=doc_history）。env 三键（部署侧配置，R-PUB-2——键名入库、实际值不入）：
	// CORE_SYSTEM_DOCS_EXPORT_INTERVAL 控制开关与间隔（如 30m；未设=不启动，不可解析/非正=
	// 缺口 stderr 可见不启动）；CORE_SYSTEM_DOCS_EXPORT_DIR 必填才启动（目标工作树）；
	// CORE_SYSTEM_DOCS_EXPORT_REMOTE 设了才 push（no force-push，认证走部署侧凭据不入库）。
	// 仅 sqlite 后端装配（git 后端自身已是布局权威，无导出语义）。导出=只读派生任务：每拍
	// 全量快照按对象规范路径写目标树→有变化才 commit→可选 push；失败只告警不回滚 sqlite 写、
	// 不崩 cored（下拍重试自愈）；绝不回导（回导=--docs 显式一次性迁移通道）。
	if _, ok := repo.(*repository.SQLiteRepo); ok {
		if iv := strings.TrimSpace(os.Getenv("CORE_SYSTEM_DOCS_EXPORT_INTERVAL")); iv != "" {
			d, perr := time.ParseDuration(iv)
			switch {
			case perr != nil || d <= 0:
				fmt.Fprintf(os.Stderr, "cored: docs 导出器=off（CORE_SYSTEM_DOCS_EXPORT_INTERVAL=%q 不可解析/非正——缺口可见不启动）\n", iv)
			case os.Getenv("CORE_SYSTEM_DOCS_EXPORT_DIR") == "":
				fmt.Fprintf(os.Stderr, "cored: docs 导出器=off（interval=%s 已设但 CORE_SYSTEM_DOCS_EXPORT_DIR 缺席——必填才启动）\n", d)
			default:
				exp, err := repository.NewExporter(repo, sch, os.Getenv("CORE_SYSTEM_DOCS_EXPORT_DIR"), os.Getenv("CORE_SYSTEM_DOCS_EXPORT_REMOTE"))
				if err != nil {
					return fmt.Errorf("docs 导出器装配: %w", err)
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
		if secrets != nil {
			// 部署装配锚经子进程 env 注入（凭据值本身永不跨进程边界——实现侧经引用式
			// 取值自行解析 secret:<name>，依赖方向声明不变，provider.md 规则 6）。
			bridge.ExtraEnv = []string{"CORE_SYSTEM_SECRETS_DB=" + secrets.Path()}
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
	runOpts.Brakes.Enabled = *brakesFlag // 刹车闸挂接态（缺省关断=零刹车既有行为面零翻转）
	objSpecState := "off（未挂文档库——object 源表不投影）"
	brakesResolverState := "defaults"
	if repo != nil {
		docsSpecs := workflow.DocsSpecs{Lookup: workflowDocLookup(repo)}
		runOpts.Specs = docsSpecs
		runOpts.Brakes.Resolver = docsSpecs                  // spec.brakes 声明收紧（未命中=强制默认兜底）
		runOpts.EnabledRules = workflowEnabledResolver(repo) // 规则开关判定面（v0.58b enabled honoring——E3 余量段收口）
		objSpecState = "on"
		brakesResolverState = "docs"
	}
	runs := workflow.NewRunLog(runOpts)
	// 通知门链接线（v0.35 KI-001 通知段——阶段门+dedup/minLevel 落地）：判定器
	// NotifyPhaseGate（人注意终态族放行 ∧ 其余事件须 involved owner 现势 run 相位终态背书
	// ——Report 序列尾/终点锚；无 run 面=fail-closed 拦）注入送达面；dedup 窗口缺省 10m
	// （notification.md 声明面缺省——同因聚合防轰炸）；minLevel 缺省不过滤。
	// apply 写入（DocApplied 等进度事实）不再触发用户通知——通知只由相位终态事实触发。
	dlOpts.PhaseGate = workflow.NotifyPhaseGate(runs)
	dlOpts.DedupWindow = *notifyDedupWindow
	dlOpts.MinLevel = *notifyMinLevel
	deliveries := workflow.NewDeliveryLog(dlOpts)
	wakes := workflow.NewWakeLog(nil)
	wakeSink := workflow.WakeSink(wakes)
	if addr := strings.TrimSpace(os.Getenv("CORE_SYSTEM_WAKE_ADDR")); addr != "" {
		wakeSink = newWakeNotifier(wakes, addr)
		fmt.Fprintf(os.Stderr, "cored: wake 桥=on(%s)（白名单事件→宿主 /wake/spawn，显式事件拉起自循环）\n", addr)
	} else {
		fmt.Fprintln(os.Stderr, "cored: wake 桥=off（纯投影；CORE_SYSTEM_WAKE_ADDR 缺席）")
	}
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
	triggers.BindAgents(wakeSink)
	triggers.WireExecutor(workflow.ExecWorkflow, workflow.ExecNotification, workflow.ExecAgent) // 机械判别器：resolve execution.wired false→true（三执行面）
	// 事件驱动推进源挂接（v0.29 事件拍——巡检拍与推进拍合一，workflow.md 规则 15 v0.29 段）：
	// **只在推进总闸开启时挂接**（推进总闸旗标 interval>0）——总闸关断=事件拍缺席，事件驱动源
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
	// 会话活性 watchdog（KI-003，event.md 规则 14/14b 引擎内部源第六段；agent-session.md §7）：
	// 装配层定时扫描 ops DB session_heartbeat——超阈值无心跳自产 SessionStalled（判定态词=
	// SuspectedStall；dedupKey=session:<id>）→自愈固定顺序（第一动作=注入固定 prompt「继续」，
	// 恢复记 SessionHealed）→诊断分诊（权限类 ApprovalNeeded 向上提交/人工类 bind 通道上报、
	// 未配置=零通知零外发/长任务 LongTaskRunning 显式汇报）→连续无效达 --heal-retry-limit 合成
	// Escalate 停自动处置。注册表闸门前置（SessionStalled 未注册=unregistered 不启动）；
	// interval=0 关断零事件零扫描（缺省不改行为面）。与注入面同一条主干零旁路。
	heartbeatStore := struct {
		sessionliveness.Registry
	}{store}
	liveness := sessionliveness.New(heartbeatStore, bus, sch, *sessionWatchInterval, *stallThreshold, *healRetryLimit)
	// SelfHeal 真实接线（v0.56 R-10——agent-session.md §7/§8 落点）：桥=dsh 宿主 plugin 侧
	// sessionController.prompt 面，cored 经 env（CORE_SYSTEM_SELFHEAL_ADDR——宿主 loopback HTTP
	// 面）回填。桥缺席（env 未设）=纯观测装配词形不变（healed-self 恢复判定照旧）；桥在=
	// 注入第一动作经宿主会话推进面（sessionController.prompt mode=queue 注入「继续」）。
	// 注入失败=err 返回照实上报绝不静默（下一拍重试第一动作=重试不是跳步）；会话侧回执心跳
	// （source=session phase=inject-continue）非续租——恢复判据=会话侧真实推进心跳。
	if addr := strings.TrimSpace(os.Getenv("CORE_SYSTEM_SELFHEAL_ADDR")); addr != "" {
		liveness.SetSelfHealBridge(newSelfHealHTTPBridge(addr, *selfHealWord))
		fmt.Fprintf(os.Stderr, "cored: 自愈注入桥=on(%s)（CORE_SYSTEM_SELFHEAL_ADDR——第一动作经 dsh 宿主 sessionController.prompt 注入「%s」，失败上报绝不静默；桥缺席=纯观测装配）\n", addr, *selfHealWord)
	}
	liveness.OnEvent(func(typ, session string, err error) {
		if err != nil {
			fmt.Fprintf(os.Stderr, "cored: sessionliveness 会话 %s 合成事件 %s 发布失败（继续上报不静默）: %v\n", session, typ, err)
			return
		}
		fmt.Fprintf(os.Stderr, "cored: sessionliveness 会话 %s 合成事件 %s 已入总线\n", session, typ)
	})
	sessionState := "off"
	switch {
	case *sessionWatchInterval <= 0:
		// 缺省关断
	case !liveness.Enabled():
		sessionState = "unregistered" // interval>0 但 SessionStalled 未注册：闸门前置不启动（缺口可见）
	default:
		sessionState = "on(" + liveness.Interval().String() + ",threshold=" + liveness.Threshold().String() + ")"
		liveCtx, cancelLive := context.WithCancel(context.Background())
		defer cancelLive()
		go liveness.Run(liveCtx)
	}
	fmt.Fprintf(os.Stderr, "cored: 会话活性 watchdog=%s（--session-watch-interval；on=停滞判定 SuspectedStall→注入继续→SessionHealed→诊断分诊→升级 Escalate 全链事件化；unregistered=SessionStalled 未注册——注册表闸门前置；off=关断零事件缺省不改行为面；心跳注册面=POST /v1/sessions/heartbeat 查询面=GET /v1/sessions）\n", sessionState)
	// 通知门链装配锚（v0.35 KI-001 通知段——探针机读判别器）：门恒挂接（修复非选项——
	// apply 写入不触发用户通知）；dedup/minLevel 部署侧可调（<=0/空=关断词形）。
	minLevelWord := *notifyMinLevel
	if minLevelWord == "" {
		minLevelWord = "off"
	}
	fmt.Fprintf(os.Stderr, "cored: 通知门链=on（阶段门=相位终态——人注意终态族 ApprovalNeeded|Escalate|HeartbeatLost|SessionStalled 恒放行（v0.54 KI-003 收编）∧其余事件须 involved owner 现势 run 相位终态背书；apply 写入不触发用户通知；dedup 窗口=%s minLevel=%s；gated|filtered|deduped 三计数=/v1/notifications/deliveries 可查询）\n",
		*notifyDedupWindow, minLevelWord)
	// connector watch 事件桥（v0.30 引擎内部源第三段，event.md 规则 12）：装配层枚举注册表
	// 内 Watch 契约驱动（connectors.Providers——依赖方向不反转：驱动住连接器仓经契约层注册，
	// cored 只消费契约通道），把驱动侧外部变化通道桥接进总线（source 服务侧判定恒 connector；
	// 类型注册闸门前置；level=注册表缺省投影；dedupKey=驱动侧因由声明）。旗标关断（缺省）
	// =零驱动零事件，缺省不改行为面；watch 事件落 run 同拍即拍（workflow.md 规则 15 v0.30
	// 段——拍面驱动源集合扩展，零新开关）。
	watchState := "off"
	if *watchSource {
		if _, err := sch.LookupEventType(watchsrc.EventType); err != nil {
			watchState = "unregistered" // 注册表闸门前置：类型未注册不启动（缺口可见）
		} else {
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
				watchState = "no-driver" // 已开启∧类型已注册∧注册表无 Watch 契约驱动（缺口可见）
			} else {
				watchState = fmt.Sprintf("on(%d)", len(watchSrcs))
				go watchsrc.New(bus, sch, watchSrcs).Run(watchCtx)
			}
		}
	}
	fmt.Fprintf(os.Stderr, "cored: watch 源=%s（--watch-source；on=装配层 watch 桥驱动 Watch 契约通道自产 ExternalChanged source=connector dedupKey=驱动侧因由——event.md 规则 12；no-driver=已开启但注册表无 Watch 契约驱动；unregistered=Scheme 未注册该类型——注册表闸门前置；off=关断零事件缺省不改行为面）\n", watchState)
	// metrics 阈值事件桥（v0.31 引擎内部源第四段，event.md 规则 13）+ MetricsInfo 数据面
	// 首段（v0.33，event.md 规则 13 数据面落点）：数据面=进程内样本库（store/queries 投影，
	// 可丢弃可重建）+自观测采样器（ops DB event_log 流水窗口计数——查询投影即样本）+读面
	// /v1/metrics；信号面=阈值桥（契约不变）——解耦口径：数据面零信号面依赖，信号面不持
	// 数据面，装配层持两者经样本通道适配。阈值声明面=文档库 WorkService.observation.metrics
	// 告警面（v0.33 随行件：判定谓词落声明面——服务侧桥只判定不发明；逐对象 involved=
	// WorkService/<name> 源级绑定锚）。旗标关断（缺省）=零样本源零事件，缺省不改行为面；
	// metrics 事件落 run 同拍即拍（workflow.md 规则 15——拍面驱动源集合扩展，零新开关）。
	metricsState := "off"
	metricsPlaneState := "off"
	var metricsStore *metricsinfo.Store
	var metricsFace apiserver.MetricsFace
	if *metricsSource {
		metricsStore = metricsinfo.NewStore()
		metricsFace = metricsStore
		// 数据面真实生产者首段：装配层自观测采样器（ops DB event_log 流水窗口计数）。
		sampler := metricsinfo.NewSampler(metricsStore, *metricsWindow, store.EventCountsSince)
		if sampler.Enabled() {
			metricsPlaneState = "on(" + sampler.Window().String() + ")"
			metricsCtx, cancelMetricsPlane := context.WithCancel(context.Background())
			defer cancelMetricsPlane()
			go sampler.Run(metricsCtx)
		} else {
			metricsPlaneState = "off（窗口<=0）"
		}
		if _, err := sch.LookupEventType(metricsrc.EventType); err != nil {
			metricsState = "unregistered" // 注册表闸门前置：类型未注册不启动（缺口可见）
		} else {
			// 阈值声明面绑定（v0.33 首段；v0.45 装配段收编 BindWorkservice 一步投影；v0.47 起
			// 装配期枚举只是初值——apply 事件驱动热重绑，声明面落库即生效免重启）：文档库
			// WorkService 对象 observation.metrics 告警面 → 桥阈值绑定；逐对象 involved=
			// WorkService/<scope>/<name>（involvedRef kind/scope/name 三段同构——owner/scope
			// 谓词与 console 事件详情对象链可解析）。零声明/身份不完整=零绑定初值（合法等待
			// 态——apply 热激活，no-source 终态退场）。
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
				return fmt.Errorf("枚举 WorkService 阈值声明面: %w", err)
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
		}
	}
	fmt.Fprintf(os.Stderr, "cored: metrics 阈值源=%s（--metrics-source；on(n)=装配层阈值桥数据面样本通道自产 MetricThreshold source=metrics dedupKey=<metric>:<window>，n=绑定数——event.md 规则 13；on(0)=装配期零绑定合法等待态（v0.47 热重绑——apply 落库即激活，no-source 终态退场）；unregistered=Scheme 未注册该类型——注册表闸门前置；off=关断零事件缺省不改行为面）\n", metricsState)
	fmt.Fprintf(os.Stderr, "cored: metrics 数据面=%s（--metrics-window；on=MetricsInfo store/queries 投影+自观测采样器 event_log 窗口计数——查询投影即样本 /v1/metrics 可查，解耦口径=数据面零信号面依赖；off=关断零样本缺省不改行为面）\n", metricsPlaneState)
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
	brakesState := "off"
	if *brakesFlag {
		brakesState = "on(" + brakesResolverState + ")"
	}
	fmt.Fprintf(os.Stderr, "cored: 刹车闸=%s（--brakes；on=受理路径冷却门（同锚同 involved cooldown 窗口内触发不落 run——braked 计数可见；跨 involved 不互踩）+推进拍变更上限（每拍每锚 run 相位变更数受生效刹车值约束——capped 计数可见，超限本拍零动作下拍续推）——workflow.md 规则 17；刹车值=spec.brakes 声明收紧，宽松方向钳回强制默认（可收紧不可关闭）；off=关断零刹车缺省不改行为面）\n", brakesState)
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
	var dispLog *flowbind.DispatchLog        // 派单受理投影（v0.50 R-8；--allow-apply 装配，nil=面缺席 503）
	var approvalSync *flowbind.SyncApprovals // ask 批准动作面（v0.56 R-10；--allow-apply 装配，nil=面缺席 503）

	// 设计文档面（v0.52 能力 A——cored-bridge.md §2.4）：--design-repo 注册白名单仓（读面=
	// git HEAD 权威），写面引擎随 --allow-apply 装配（白名单→R-PUB-1 lint→乐观锁 baseSha 三检查
	// +权限准入 + git commit 本地落盘不 push）。缺省不注册=面缺席 503（读/写同口径）。
	var designDocs *docrepo.Set
	if len(designRepos) > 0 {
		designDocs, err = docrepo.NewSet(designRepos...)
		if err != nil {
			return err
		}
		fmt.Fprintf(os.Stderr, "cored: 设计文档面=on(%d:%s)（--design-repo；读面权威=git HEAD 零缓存，写面=%s——白名单 .md→R-PUB-1 lint→乐观锁 baseSha 三检查+权限准入 core/design-docs+git commit 本地落盘不 push；草稿=与对象草稿同一缓存 kind=DesignDoc——能力 A，cored-bridge.md §2.4）\n",
			len(designDocs.Names()), strings.Join(designDocs.Names(), ","),
			map[bool]string{true: "on（引擎随 --allow-apply 装配）", false: "off（写路由 405——引擎未装配）"}[*allowApply])
	}
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
		// workflow 执行模型绑定面（v0.49 R-7，workflow-runtime.md §11.1——apply 即实现骨架）：
		// 写面开启时装配（绑定物化=一次普通 apply，同闸门同审计零旁路）。存量枚举=初值
		// （metrics 阈值桥 v0.47 同款范式）+ DocApplied×Workflow 订阅热路径（积压合并一次
		// 枚举装到最新面；失败=保旧面+stderr 告警+下次 apply 重试自愈）。所有面=只写
		// origin=workflow-bind 标注对象，人/agent 实现零盲写。
		// 派单受理投影（v0.50 R-8，§11.2 拉起形态首段）：一次真实物化动作（created|updated）
		// 即一条 apply-workflow 锚派单受理（autoLevel=ask 数据面在册；noop/skipped 不派单）。
		dispLog = flowbind.NewDispatchLog(nil)
		flowMat := &flowbind.Materializer{Repo: repo, Apply: engine, Dispatch: dispLog}
		// 派单 ask 批准动作面（v0.56 R-10——workflow-runtime.md §11.4 第一段）：派单受理同拍
		// 记 pending 批准投影（autoLevel=ask 的派单默认挂批准——ask=Planning 后挂起等批）；
		// 人批/驳回=Apply 判定（fieldManager=workflow:approve|workflow:reject 的 ok 行）→
		// 投影翻转 approved|rejected（审批同步器=apply_log 行集匹配——零新写路径零旁路）。
		// **approved 派单投影 ≠ spawn 已拉起**（红线 a：真实拉起/执行仍 observe）。
		approvalLog := flowbind.NewApprovalLog(nil)
		approvalSync = &flowbind.SyncApprovals{Log: approvalLog}
		dispLog.Approvals = approvalSync
		fmt.Fprintf(os.Stderr, "cored: 派单批准动作面=on（autoLevel=ask 派单默认挂批准投影——ApprovalNeeded→人批=Apply workflow:approve→approved；spawn 接线=agent-session 域 declared-deferred，真实执行仍 observe）\n")
		fbCtx, cancelFlowBind := context.WithCancel(context.Background())
		defer cancelFlowBind()
		n0, fails0 := flowMat.EnsureAll(ctx)
		for _, f := range fails0 {
			fmt.Fprintf(os.Stderr, "cored: workflow 绑定存量物化失败 %s/%s（保旧面）: %v\n", f.Scope, f.Name, f.Err)
		}
		fmt.Fprintf(os.Stderr, "cored: workflow 绑定=on(%d)（--allow-apply；apply 即实现骨架物化——同名/spec.connector 覆写 ConnectorOps+flow.nodes 节点接线，origin=workflow-bind 所有面；实现代码=apply-workflow agent 余量；存量初值=%d 热路径=DocApplied×Workflow 订阅）\n", n0, n0)
		fmt.Fprintf(os.Stderr, "cored: workflow 派单=on（agent=apply-workflow autoLevel=ask——骨架物化即派单受理投影 GET /v1/agents/dispatches；受理≠拉起，真实拉起 deferred——workflow-runtime.md §11.2）\n")
		// 需求提案分流/派单 watcher（v0.51 能力 B，concepts/proposal.md——派单面第二触发源）：
		// DocApplied×Proposal 订阅——submitted∧高危=ApprovalNeeded 自产（source=core，event.md 规则 15）；
		// submitted∧低危命中 --proposal-auto-approve 清单=引擎自写批准（proposal:auto，granted 行随行）；
		// approved=apply-workflow 派单受理（kind=proposal 行——载荷带提案名/目标/验收，eventId 同源）；
		// rejected 不派单。全部写动作走 apply 引擎同一闸门；批准只产受理投影不拉起（红线）。
		autoList := []string{}
		for _, p := range strings.Split(*proposalAutoApprove, ",") {
			if p = strings.TrimSpace(p); p != "" {
				autoList = append(autoList, p)
			}
		}
		propWatcher := &proposal.Watcher{
			Repo: repo, Apply: engine, Bus: bus, Sch: sch, Dispatch: dispLog,
			Approvals: store, AutoApprove: autoList,
		}
		propCtx, cancelProp := context.WithCancel(context.Background())
		defer cancelProp()
		go propWatcher.Start(propCtx, bus,
			func(action string) {
				fmt.Fprintf(os.Stderr, "cored: proposal 分流动作=%s（DocApplied×Proposal→审批分流/派单受理——concepts/proposal.md）\n", action)
			},
			func(err error) {
				fmt.Fprintf(os.Stderr, "cored: proposal 分流失败（保对象面，下次 apply 重试自愈）: %v\n", err)
			})
		triageState := "off（全人审）"
		if len(autoList) > 0 {
			triageState = "on(" + strings.Join(autoList, ",") + ")"
		}
		fmt.Fprintf(os.Stderr, "cored: proposal 分流=%s 派单=on（--proposal-auto-approve；高危提交 ApprovalNeeded 自产 source=core，批准即 apply-workflow 派单受理 kind=proposal——受理≠拉起）\n", triageState)
		go flowMat.Start(fbCtx, bus,
			func(n int) {
				fmt.Fprintf(os.Stderr, "cored: workflow 绑定热路径 n=%d（DocApplied×Workflow→骨架物化/更新——apply 即实现 §11.1）\n", n)
			},
			func(fails []flowbind.Fail) {
				for _, f := range fails {
					fmt.Fprintf(os.Stderr, "cored: workflow 绑定热路径失败 %s/%s（保旧面+下次 apply 重试自愈）: %v\n", f.Scope, f.Name, f.Err)
				}
			})
		// 事件注入面（v0.20 发布面首段，event.md 规则 8）：与 apply 写面同门控（--allow-apply），
		// 发布权限域与 apply 写权限域分立（publish × events 资源域）；注入走与引擎内部事件
		// 同一条总线主干（去重/订阅派发/event_log 归档零旁路）；source 服务侧判定恒 human。
		publisher = &eventpub.Publisher{Sch: sch, Bus: bus, Perm: perm}
		// 设计文档写面引擎（v0.52 能力 A）：写设计文档=permission.md 机读判定（core/design-docs
		// 资源域——scheme.yaml Role 声明词汇 core/*），审计=apply_log 行+DocApplied 事件。
		designDocs.BindEngine(&docrepo.Engine{Bus: bus, Ops: store, Perm: perm, Domain: "core/design-docs"})
		fmt.Fprintf(os.Stderr, "cored: apply 写面已开启（POST /v1/apply；请求级主体判定，匿名拒绝；本地信任仅作引擎 fallback）\n")
		fmt.Fprintf(os.Stderr, "cored: 事件注入面已开启（POST /v1/events；动作=publish × 域=events 与 apply 写面分立；注入通道 source=human——v0.20）\n")
	}

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		return fmt.Errorf("监听 %s: %w", *addr, err)
	}
	// 任务调度器（多任务并发调度首段——concepts/task-scheduler.md，runtime/task-scheduler.md §1-§4）：
	// 装配层 tick goroutine 每拍读全部 Task（跨 scope）→ 取消/租约/Apply 关口冲突链 → 依赖/冲突/
	// 上限（plan 纯函数）→ 写 status=running（repository 系统写 fieldManager=task-scheduler——P2
	// status 单一写者）+ TaskStateChanged（source=core，dedup=task:<id>:<phase>）→ POST 宿主
	// /task/spawn（CORE_SYSTEM_TASK_SPAWN_ADDR 回填；缺席=spawn 关断保持 ready，lastError 可见
	// 退避重试不热循环）。租约回收/取消迁移/applyConflictGate 裁决两路同拍。配置=scheme
	// runtime.taskScheduler 每拍热读（Apply 即时生效）；enabled=false 或未挂文档库=不启动。
	schedCfg := scheduler.ConfigOf(sch)
	schedState := "off（scheme runtime.taskScheduler.enabled=false 或未挂文档库）"
	if schedCfg.Enabled && repo != nil {
		sched := scheduler.New(repo, bus, sch, store, store, scheduler.Options{
			SpawnAddr: strings.TrimSpace(os.Getenv(scheduler.EnvSpawnAddr)),
		})
		schedCtx, cancelSched := context.WithCancel(context.Background())
		defer cancelSched()
		go sched.Run(schedCtx)
		spawnWord := "off（CORE_SYSTEM_TASK_SPAWN_ADDR 缺席——保持 ready，lastError 可见退避重试）"
		if sched.SpawnEnabled() {
			spawnWord = "on(" + strings.TrimSpace(os.Getenv(scheduler.EnvSpawnAddr)) + "/task/spawn)"
		}
		schedState = fmt.Sprintf("on(tick=%s maxConcurrent=%d perService=%d perRepo=%d leaseTtl=%s maxAttempts=%d spawn=%s conflictGate=%s)",
			sched.Interval(), schedCfg.MaxConcurrent, schedCfg.PerService, schedCfg.PerRepo,
			schedCfg.LeaseTTL, schedCfg.MaxAttempts, spawnWord, schedCfg.ApplyConflictGate)
	}
	fmt.Fprintf(os.Stderr, "cored: 任务调度器=%s（runtime.taskScheduler 每拍热读——无关任务并发/相关任务串行；status 写面=task-scheduler；事件=TaskStateChanged source=core；spawn 桥与 wake 桥同族固定载荷不进 shell；applyConflictGate=ask(pause-first|abort-current) 裁决事件 TaskConflictResolved dedup 尾段判别）\n", schedState)
	apiserverOpts := func() []apiserver.Option {
		opts := apiserverOpts(metricsFace, secrets, store, dispLog, approvalSync, designDocs)
		if sessionState != "off" && sessionState != "unregistered" {
			opts = append(opts, apiserver.WithSessionLiveness(liveness))
		}
		// 历史 rv 正文读面（v0.63 rev-content diff 面，§1.40——历史 tab 点行 diff 上一版读半边）
		// 与逐次正文历史清单面（R-DOC-7 规则 4 doc_history 落地面，GET /v1/doc-history）按后端
		// 分立装配：git 后端=rev-content 恒装配（裸库考古，source=git-backend）、doc-history
		// 缺席 503（git 历史=导出快照非逐次账——不冒充）；sqlite 后端=两 face 恒装配
		// （rev-content 数据源=doc_history 逐次账，source=sqlite-doc-history——503 升级为可用）。
		switch src := store2.(type) {
		case *repository.GitRepo:
			opts = append(opts, apiserver.WithRevContent(src))
		case *repository.SQLiteRepo:
			opts = append(opts, apiserver.WithRevContent(src), apiserver.WithDocHistory(src))
		}
		return opts
	}
	srv := &http.Server{Handler: apiserver.New(sch, store2, applier, drafts, triggers, runs, deliveries, wakes, publisher, apiserverOpts()...), ReadHeaderTimeout: 5 * time.Second}
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
