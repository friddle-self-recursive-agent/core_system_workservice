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
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/draft"
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
	if *allowApply {
		// 权限对象装载（v0.9 对象准入）：Role/Group 文档对象是授权面唯一来源——
		// 主体断言的 role 名必须对上 global/roles、global/groups 里的对象才产生权限。
		// 启动装载=权威快照；事件驱动刷新随 dispatch 接线（declared-deferred）。
		perm := &permission.System{}
		objs, err := permDocs(ctx, repo)
		if err != nil {
			return err
		}
		if err := perm.LoadObjects(objs); err != nil {
			return fmt.Errorf("权限对象装载失败（授权面宁拒不降级）: %w", err)
		}
		fmt.Fprintf(os.Stderr, "cored: 权限对象装载 roles=%d groups=%d（global/roles、global/groups）\n",
			len(perm.Roles), len(perm.Groups))
		// 主体（v0.9 请求级）：HTTP 写面按请求解析（subject 字段/X-Core-Subject 头，匿名=照判拒绝）；
		// 进程级本地信任退化为引擎缺省 fallback（程序化调用，如 bootstrap/工作流自写）。
		engine := &apply.Engine{
			Sch: sch, Repo: repo, Bus: bus, Ops: store, Drafts: drafts,
			Perm:    perm,
			Subject: permission.Subject{User: "cored:local", Roles: []string{permission.DefaultAdmin}},
		}
		applier = engine
		fmt.Fprintf(os.Stderr, "cored: apply 写面已开启（POST /v1/apply；请求级主体判定，匿名拒绝；本地信任仅作引擎 fallback）\n")
	}

	ln, err := net.Listen("tcp", *addr)
	if err != nil {
		return fmt.Errorf("监听 %s: %w", *addr, err)
	}
	srv := &http.Server{Handler: apiserver.New(sch, store2, applier, drafts), ReadHeaderTimeout: 5 * time.Second}
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
