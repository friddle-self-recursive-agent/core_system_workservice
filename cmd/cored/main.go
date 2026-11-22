// cored Core 常驻进程（v0.3）：Scheme 载入自检 + 只读查询面 + 显式开启的 Apply 写入口。
// 设计依据：core_system_instance/impl/self-recursive-core.md（cored=冷启动入口与调试用途）
// 与 runtime/cored-bridge.md（plugin↔cored 桥：CORED_READY 就绪行 + /v1 查询面）。
// 部署锚点一律经参数/环境注入（--scheme/--docs/--ops-db / CORE_SYSTEM_*），不内置真实路径。
// 写面缺省关闭（缺省只读暴露）；--allow-apply 显式开启后 POST /v1/apply 是唯一写端点，
// 内部收敛 apply.Engine 三步闸门（scheme 准入→权限准入→乐观锁落库→DocApplied 事件）。
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
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "cored:", err)
		os.Exit(1)
	}
}

func run() error {
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
	var applier apiserver.Applier
	if *allowApply {
		// 本地信任面：写面由部署方显式开启，主体在本进程内持内置 admin；
		// 权限系统 v0.9 接请求级主体后收敛（契约锚定 cored-bridge.md §2.1）。
		engine := &apply.Engine{
			Sch: sch, Repo: repo, Bus: bus, Ops: store,
			Perm:    &permission.System{},
			Subject: permission.Subject{User: "cored:local", Roles: []string{permission.DefaultAdmin}},
		}
		applier = engine
		fmt.Fprintf(os.Stderr, "cored: apply 写面已开启（POST /v1/apply；本地信任主体，权限请求级判定待 v0.9 接入）\n")
	}

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		return fmt.Errorf("监听 %s: %w", *addr, err)
	}
	srv := &http.Server{Handler: apiserver.New(sch, store2, applier), ReadHeaderTimeout: 5 * time.Second}
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
