// corectl — core_system 命令行（桌面功能的同源投影）。
//
// 设计：core_design/concepts/cli.md（R-UX-8「命令行与桌面同源」）。
// 接口：cored 既有 /v1/* 读写面（cored-bridge.md §2）；不直连 DB、不新增写路径。
// 写命令只走两个既有入口：POST /v1/apply（文档）与 POST /v1/events（事件）。
package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const cliVersion = "0.1.0"

// stringList 可重复 flag（--role/--group）。
type stringList []string

func (s *stringList) String() string     { return strings.Join(*s, ",") }
func (s *stringList) Set(v string) error { *s = append(*s, v); return nil }

// subject 请求级主体（与 cored POST 载荷 subject 字段同形）。
type subject struct {
	User   string   `json:"user,omitempty"`
	Roles  []string `json:"roles,omitempty"`
	Groups []string `json:"groups,omitempty"`
}

type app struct {
	server  string
	subject subject
	jsonOut bool
	out     io.Writer
	err     io.Writer
	hc      *http.Client
}

func main() { os.Exit(run(os.Args[1:], os.Stdout, os.Stderr)) }

func run(args []string, out, errOut io.Writer) int {
	gs := flag.NewFlagSet("corectl", flag.ContinueOnError)
	gs.SetOutput(errOut)
	server := gs.String("server", defaultServer(), "cored 地址（缺省读 CORED_ADDR / ~/.dsh/cored.addr）")
	user := gs.String("subject", envOr("CORECTL_SUBJECT", "console-operator"), "请求级主体 user")
	var roles, groups stringList
	gs.Var(&roles, "role", "主体 role（可重复；缺省 role/global/admin）")
	gs.Var(&groups, "group", "主体 group（可重复）")
	jsonOut := gs.Bool("json", false, "输出紧凑 JSON（缺省人读）")
	timeout := gs.Duration("timeout", 20*time.Second, "单次 HTTP 超时")
	gs.Usage = func() { usage(errOut) }
	if err := gs.Parse(args); err != nil {
		return 2
	}
	rest := gs.Args()
	if len(rest) == 0 {
		usage(errOut)
		return 2
	}
	if strings.TrimSpace(*server) == "" {
		fmt.Fprintln(errOut, "错误：缺 cored 地址（--server / CORED_ADDR / ~/.dsh/cored.addr；地址属部署侧）")
		return 2
	}
	if len(roles) == 0 {
		roles = []string{"role/global/admin"}
	}
	a := &app{
		server:  strings.TrimRight(*server, "/"),
		subject: subject{User: *user, Roles: roles, Groups: groups},
		jsonOut: *jsonOut,
		out:     out,
		err:     errOut,
		hc:      &http.Client{Timeout: *timeout},
	}
	return a.dispatch(rest[0], rest[1:])
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

// defaultServer cored 地址发现：CORED_ADDR → ~/.dsh/cored.addr（插件就绪后写入，R-UX-8）。
func defaultServer() string {
	if v := strings.TrimSpace(os.Getenv("CORED_ADDR")); v != "" {
		return v
	}
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	b, err := os.ReadFile(filepath.Join(home, ".dsh", "cored.addr"))
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(b))
}

func usage(w io.Writer) {
	fmt.Fprint(w, `corectl — core_system 命令行（桌面同源控制面）

用法：corectl [全局参数] <命令> [参数]

全局参数：
  --server <url>     cored 地址（缺省 CORED_ADDR 或 ~/.dsh/cored.addr）
  --subject <user>   请求级主体（缺省 console-operator）
  --role <role>      主体 role，可重复（缺省 role/global/admin）
  --group <group>    主体 group，可重复
  --json             紧凑 JSON 输出
  --timeout <dur>    HTTP 超时（缺省 20s）

读取（对应桌面各屏）：
  report                              关注报告（CLI 摘要）
  status                              控制面健康 + scheme 概览
  get <kind> [name] [--scope s]       对象清单/详情
  cat <kind> <name> [--scope s]       只打印对象正文
  workflows                           工作流列表/现势
  runs                                运行投影
  events [--type T] [--level L] [--limit N] [--since RFC3339]
  approvals [--decision granted|rejected] [--obj k/s/n] [--limit N]
  inbox                               待你处理（四类）
  wakes                               唤醒投影
  dispatches                          派单受理投影
  sessions                            会话心跳/活性
  proposals [name]                    提案列表/详情
  tasks [--phase P] [--scope S]       任务队列简表（status 投影）
  task <scope>/<name>                 任务详情（spec+status 投影+原文）
  find "查询词" [--tags t1,t2] [--k N] 上下文相关性检索（score 0-100）
  trace <kind>/<scope>/<name>         文档 trace 双向聚合（implements/implemented_by/verified_by）

写入（与桌面同一条 Apply/Event/权限链）：
  apply -f <file> [--kind K] [--scope S] [--name N] [--rv N] [--dry-run] [--field-manager M]
  event publish --type T [--involved k/s/n] [--level L] [--dedup K]
  tick [--involved k/s/n] [--dedup K]  发布 ReconcileTick（系统事件触发 core-workflow）
  loop [--dedup K]                     发布 LoopRequested（事件→wake 桥→拉起自循环）
  task submit -f <file>                提交任务（apply kind=Task 同一写链）
  task cancel <scope>/<name>           发 TaskCancelRequested（调度器每拍扫描迁移 canceled）

其他：
  version                              CLI + cored 版本
`)
}

func (a *app) dispatch(cmd string, args []string) int {
	switch cmd {
	case "report":
		return a.cmdReport(args)
	case "status":
		return a.cmdStatus(args)
	case "get":
		return a.cmdGet(args)
	case "cat":
		return a.cmdCat(args)
	case "apply":
		return a.cmdApply(args)
	case "workflows":
		return a.cmdWorkflows(args)
	case "runs":
		return a.cmdRuns(args)
	case "events":
		return a.cmdEvents(args)
	case "event":
		return a.cmdEvent(args)
	case "approvals":
		return a.cmdApprovals(args)
	case "inbox":
		return a.cmdInbox(args)
	case "wakes":
		return a.cmdSimpleGET(args, "/v1/agents/wakes", "wakes")
	case "dispatches":
		return a.cmdSimpleGET(args, "/v1/agents/dispatches", "dispatches")
	case "sessions":
		return a.cmdSessions(args)
	case "proposals":
		return a.cmdProposals(args)
	case "tasks":
		return a.cmdTasks(args)
	case "task":
		return a.cmdTask(args)
	case "find":
		return a.cmdFind(args)
	case "trace":
		return a.cmdTrace(args)
	case "tick":
		return a.cmdTick(args)
	case "loop":
		return a.cmdLoop(args)
	case "version":
		return a.cmdVersion(args)
	case "help", "-h", "--help":
		usage(a.out)
		return 0
	default:
		fmt.Fprintf(a.err, "未知命令：%s\n\n", cmd)
		usage(a.err)
		return 2
	}
}

// ---------- HTTP ----------

func (a *app) request(method, path string, body any) ([]byte, error) {
	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		r = bytes.NewReader(b)
	}
	req, err := http.NewRequest(method, a.server+path, r)
	if err != nil {
		return nil, err
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
		if sub, err := json.Marshal(a.subject); err == nil {
			req.Header.Set("X-Core-Subject", string(sub))
		}
	}
	resp, err := a.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		msg := strings.TrimSpace(string(respBody))
		if msg == "" {
			msg = http.StatusText(resp.StatusCode)
		}
		return nil, fmt.Errorf("HTTP %d: %s", resp.StatusCode, msg)
	}
	return respBody, nil
}

func (a *app) emit(raw []byte) int {
	if a.jsonOut {
		fmt.Fprintln(a.out, string(bytes.TrimSpace(raw)))
		return 0
	}
	var buf bytes.Buffer
	if err := json.Indent(&buf, raw, "", "  "); err != nil {
		fmt.Fprintln(a.out, string(raw))
		return 0
	}
	fmt.Fprintln(a.out, buf.String())
	return 0
}

func (a *app) getJSON(path string, query url.Values, v any) error {
	if len(query) > 0 {
		path += "?" + query.Encode()
	}
	raw, err := a.request(http.MethodGet, path, nil)
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, v)
}

func fail(a *app, err error) int {
	fmt.Fprintf(a.err, "错误：%v\n", err)
	return 1
}

// ---------- 读取命令 ----------

type schemeSummary struct {
	APIVersion     string `json:"apiVersion"`
	Version        string `json:"version"`
	KindCount      int    `json:"kindCount"`
	DocTypeCount   int    `json:"docTypeCount"`
	EventTypeCount int    `json:"eventTypeCount"`
	DispatchCount  int    `json:"dispatchCount"`
}

func (a *app) cmdStatus(args []string) int {
	fs := flag.NewFlagSet("status", flag.ContinueOnError)
	fs.SetOutput(a.err)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	var health map[string]any
	if err := a.getJSON("/healthz", nil, &health); err != nil {
		return fail(a, err)
	}
	var sum schemeSummary
	if err := a.getJSON("/v1/scheme", nil, &sum); err != nil {
		return fail(a, err)
	}
	if a.jsonOut {
		return a.emit(mustJSON(map[string]any{"health": health, "scheme": sum}))
	}
	fmt.Fprintf(a.out, "cored: ok\n")
	fmt.Fprintf(a.out, "scheme: %s %s · kinds=%d docTypes=%d eventTypes=%d dispatch=%d\n",
		sum.APIVersion, sum.Version, sum.KindCount, sum.DocTypeCount, sum.EventTypeCount, sum.DispatchCount)
	return 0
}

type objectList struct {
	Count   int `json:"count"`
	Objects []struct {
		Kind            string `json:"kind"`
		Name            string `json:"name"`
		Scope           string `json:"scope"`
		ResourceVersion int64  `json:"resourceVersion"`
	} `json:"objects"`
}

func (a *app) cmdGet(args []string) int {
	scope, pos, err := parseScopeAndPos(args)
	if err != nil {
		return fail(a, err)
	}
	if len(pos) < 1 {
		fmt.Fprintln(a.err, "用法：corectl get <kind> [name] [--scope s]")
		return 2
	}
	kind := pos[0]
	path := "/v1/objects/" + url.PathEscape(kind)
	if len(pos) > 1 {
		path += "/" + url.PathEscape(pos[1])
	}
	q := url.Values{}
	if scope != "" {
		q.Set("scope", scope)
	}
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	raw, err := a.request(http.MethodGet, path, nil)
	if err != nil {
		return fail(a, err)
	}
	return a.emit(raw)
}

// parseScopeAndPos 解析 `--scope`（允许出现在位置参数之后）+ 位置参数。
func parseScopeAndPos(args []string) (scope string, pos []string, err error) {
	for i := 0; i < len(args); i++ {
		switch {
		case args[i] == "--scope" || args[i] == "-scope":
			if i+1 >= len(args) {
				return "", nil, fmt.Errorf("缺 --scope 值")
			}
			scope = args[i+1]
			i++
		case strings.HasPrefix(args[i], "--scope="):
			scope = strings.TrimPrefix(args[i], "--scope=")
		case strings.HasPrefix(args[i], "-scope="):
			scope = strings.TrimPrefix(args[i], "-scope=")
		case strings.HasPrefix(args[i], "-"):
			return "", nil, fmt.Errorf("未知参数：%s", args[i])
		default:
			pos = append(pos, args[i])
		}
	}
	return scope, pos, nil
}

func (a *app) cmdCat(args []string) int {
	scope, pos, err := parseScopeAndPos(args)
	if err != nil {
		return fail(a, err)
	}
	if len(pos) < 2 {
		fmt.Fprintln(a.err, "用法：corectl cat <kind> <name> [--scope s]")
		return 2
	}
	q := url.Values{}
	if scope != "" {
		q.Set("scope", scope)
	}
	var obj struct {
		Content string `json:"content"`
	}
	path := "/v1/objects/" + url.PathEscape(pos[0]) + "/" + url.PathEscape(pos[1])
	if err := a.getJSON(path, q, &obj); err != nil {
		return fail(a, err)
	}
	fmt.Fprint(a.out, obj.Content)
	if !strings.HasSuffix(obj.Content, "\n") {
		fmt.Fprintln(a.out)
	}
	return 0
}

type workflowsResp struct {
	Count     int `json:"count"`
	Workflows []struct {
		Name    string `json:"name"`
		Runs24h int    `json:"runs24h"`
		Latest  *struct {
			Phase  string `json:"phase"`
			Status string `json:"status"`
		} `json:"latest"`
	} `json:"workflows"`
}

func (a *app) cmdWorkflows(args []string) int {
	fs := flag.NewFlagSet("workflows", flag.ContinueOnError)
	fs.SetOutput(a.err)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	raw, err := a.request(http.MethodGet, "/v1/workflows", nil)
	if err != nil {
		return fail(a, err)
	}
	return a.emit(raw)
}

type runsResp struct {
	Count int `json:"count"`
	Runs  []struct {
		Run    int64  `json:"run"`
		Target string `json:"target"`
		Phase  string `json:"phase"`
		Status string `json:"status"`
	} `json:"runs"`
}

func (a *app) cmdRuns(args []string) int {
	fs := flag.NewFlagSet("runs", flag.ContinueOnError)
	fs.SetOutput(a.err)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	raw, err := a.request(http.MethodGet, "/v1/workflows/runs", nil)
	if err != nil {
		return fail(a, err)
	}
	return a.emit(raw)
}

type eventsResp struct {
	Count  int `json:"count"`
	Events []struct {
		EventID  string `json:"eventId"`
		Type     string `json:"type"`
		Level    string `json:"level"`
		Source   string `json:"source"`
		Involved string `json:"involved"`
		TS       string `json:"ts"`
	} `json:"events"`
}

func (a *app) cmdEvents(args []string) int {
	fs := flag.NewFlagSet("events", flag.ContinueOnError)
	fs.SetOutput(a.err)
	typ := fs.String("type", "", "事件类型")
	level := fs.String("level", "", "等级 debug|info|warn|error|fatal")
	limit := fs.Int("limit", 50, "条数上限（≤500）")
	since := fs.String("since", "", "起始时间 RFC3339")
	showInfra := fs.Bool("show-infra", false, "展开 tier=infra/default-hidden 底层节拍事件（Heartbeat/ReconcileTick 缺省不展示——BUG-010）")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	q := url.Values{}
	if *typ != "" {
		q.Set("type", *typ)
	}
	if *level != "" {
		q.Set("level", *level)
	}
	if *since != "" {
		q.Set("since", *since)
	}
	if *showInfra {
		q.Set("showInfra", "true")
	}
	q.Set("limit", fmt.Sprint(*limit))
	raw, err := a.request(http.MethodGet, "/v1/events?"+q.Encode(), nil)
	if err != nil {
		return fail(a, err)
	}
	return a.emit(raw)
}

func (a *app) cmdApprovals(args []string) int {
	fs := flag.NewFlagSet("approvals", flag.ContinueOnError)
	fs.SetOutput(a.err)
	decision := fs.String("decision", "", "granted|rejected")
	obj := fs.String("obj", "", "kind/scope/name")
	limit := fs.Int("limit", 50, "条数上限（≤500）")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	q := url.Values{}
	if *decision != "" {
		q.Set("decision", *decision)
	}
	if *obj != "" {
		q.Set("obj", *obj)
	}
	q.Set("limit", fmt.Sprint(*limit))
	raw, err := a.request(http.MethodGet, "/v1/approvals?"+q.Encode(), nil)
	if err != nil {
		return fail(a, err)
	}
	return a.emit(raw)
}

func (a *app) cmdInbox(args []string) int {
	fs := flag.NewFlagSet("inbox", flag.ContinueOnError)
	fs.SetOutput(a.err)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	raw, err := a.request(http.MethodGet, "/v1/inbox", nil)
	if err != nil {
		return fail(a, err)
	}
	return a.emit(raw)
}

func (a *app) cmdSessions(args []string) int {
	fs := flag.NewFlagSet("sessions", flag.ContinueOnError)
	fs.SetOutput(a.err)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	raw, err := a.request(http.MethodGet, "/v1/sessions", nil)
	if err != nil {
		return fail(a, err)
	}
	return a.emit(raw)
}

func (a *app) cmdProposals(args []string) int {
	fs := flag.NewFlagSet("proposals", flag.ContinueOnError)
	fs.SetOutput(a.err)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	path := "/v1/proposals"
	if rest := fs.Args(); len(rest) > 0 {
		path += "/" + url.PathEscape(rest[0])
	}
	raw, err := a.request(http.MethodGet, path, nil)
	if err != nil {
		return fail(a, err)
	}
	return a.emit(raw)
}

func (a *app) cmdSimpleGET(args []string, path, name string) int {
	fs := flag.NewFlagSet(name, flag.ContinueOnError)
	fs.SetOutput(a.err)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	raw, err := a.request(http.MethodGet, path, nil)
	if err != nil {
		return fail(a, err)
	}
	return a.emit(raw)
}

// ---------- 任务队列（concepts/task-scheduler.md——corectl tasks/task 同源投影） ----------

// taskCancelEvent / taskConflictResolved 取消与裁决事件类型（corectl 只产事件，判定在 cored
// 调度器——每拍扫 eventLog；词形与 scheduler 包常量同源）。TaskCancelRequested 当前 scheme
// 未注册时 POST /v1/events 返回 404 注册闸门错——缺口可见，注册即通（不改代码）。
const (
	taskCancelEvent      = "TaskCancelRequested"
	taskConflictResolved = "TaskConflictResolved"
)

func (a *app) cmdTasks(args []string) int {
	fs := flag.NewFlagSet("tasks", flag.ContinueOnError)
	fs.SetOutput(a.err)
	phase := fs.String("phase", "", "phase 过滤（queued|ready|blocked|running|verifying|done|failed|canceled）")
	scope := fs.String("scope", "", "scope 收窄")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	q := url.Values{}
	if *phase != "" {
		q.Set("phase", *phase)
	}
	if *scope != "" {
		q.Set("scope", *scope)
	}
	path := "/v1/tasks"
	if len(q) > 0 {
		path += "?" + q.Encode()
	}
	raw, err := a.request(http.MethodGet, path, nil)
	if err != nil {
		return fail(a, err)
	}
	return a.emit(raw)
}

// cmdTask 任务子命令：`task <scope>/<name>` 详情 / `task submit -f <file>` 提交 /
// `task cancel <scope>/<name>` 取消（发 TaskCancelRequested 事件）。
func (a *app) cmdTask(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(a.err, "用法：corectl task <scope>/<name> | task submit -f <file> | task cancel <scope>/<name>")
		return 2
	}
	switch args[0] {
	case "submit":
		return a.cmdTaskSubmit(args[1:])
	case "cancel":
		return a.cmdTaskCancel(args[1:])
	default:
		return a.cmdTaskDetail(args)
	}
}

// cmdTaskDetail 任务详情：GET /v1/tasks/<scope>/<name>（键全给定直取）。
func (a *app) cmdTaskDetail(args []string) int {
	id := args[len(args)-1]
	scope, name, ok := strings.Cut(id, "/")
	if !ok || scope == "" || name == "" {
		fmt.Fprintln(a.err, "错误：任务 ID 词形=<scope>/<name>（如 self-recursive-core/fix-gate-123）")
		return 2
	}
	raw, err := a.request(http.MethodGet, "/v1/tasks/"+url.PathEscape(scope)+"/"+url.PathEscape(name), nil)
	if err != nil {
		return fail(a, err)
	}
	return a.emit(raw)
}

// cmdTaskSubmit 提交任务：POST /v1/apply kind=Task（与桌面/console 同一条 apply 链——零旁路；
// kind 词形强制 Task，frontmatter 声明不符=拒绝）。
func (a *app) cmdTaskSubmit(args []string) int {
	fs := flag.NewFlagSet("task submit", flag.ContinueOnError)
	fs.SetOutput(a.err)
	file := fs.String("f", "", "任务文件（- = stdin）")
	dry := fs.Bool("dry-run", false, "只准入校验，不落库")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *file == "" {
		fmt.Fprintln(a.err, "用法：corectl task submit -f <file>")
		return 2
	}
	var content []byte
	var err error
	if *file == "-" {
		content, err = io.ReadAll(io.LimitReader(os.Stdin, 4<<20))
	} else {
		content, err = os.ReadFile(*file)
	}
	if err != nil {
		return fail(a, err)
	}
	fk, fscope, fname := parseDocHeader(string(content))
	if fk != "" && fk != "Task" {
		return fail(a, fmt.Errorf("frontmatter kind=%q 与任务提交不符（只收 kind: Task）", fk))
	}
	if fscope == "" || fname == "" {
		fmt.Fprintln(a.err, "错误：frontmatter 缺 scope/name（任务文件须带 metadata.scope/metadata.name 或扁平 scope/name）")
		return 2
	}
	body := applyBody{
		Kind: "Task", Scope: fscope, Name: fname,
		Content: string(content), FieldManager: "corectl", DryRun: *dry, Subject: &a.subject,
	}
	raw, err := a.request(http.MethodPost, "/v1/apply", body)
	if err != nil {
		return fail(a, err)
	}
	if a.jsonOut {
		return a.emit(raw)
	}
	fmt.Fprintf(a.out, "已提交任务 %s/%s（kind=Task，apply 同链；调度器下拍选取）\n", fscope, fname)
	fmt.Fprintln(a.out, strings.TrimSpace(string(raw)))
	fmt.Fprintln(a.out, "查看：corectl tasks | corectl task "+fscope+"/"+fname)
	return 0
}

// cmdTaskCancel 取消任务：POST /v1/events type=TaskCancelRequested involved=Task/<scope>/<name>
// （调度器每拍扫描 eventLog → 非 running 直接 canceled / running 注记后自然结束迁移）。
func (a *app) cmdTaskCancel(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(a.err, "用法：corectl task cancel <scope>/<name>")
		return 2
	}
	id := args[len(args)-1]
	scope, name, ok := strings.Cut(id, "/")
	if !ok || scope == "" || name == "" {
		fmt.Fprintln(a.err, "错误：任务 ID 词形=<scope>/<name>")
		return 2
	}
	raw, err := a.request(http.MethodPost, "/v1/events", eventBody{
		Type:     taskCancelEvent,
		Involved: "Task/" + scope + "/" + name,
		DedupKey: fmt.Sprintf("task-cancel:%s/%s:%d", scope, name, time.Now().UnixNano()),
		Subject:  &a.subject,
	})
	if err != nil {
		return fail(a, err)
	}
	if a.jsonOut {
		return a.emit(raw)
	}
	fmt.Fprintf(a.out, "已发布取消请求 %s（involved=Task/%s/%s；调度器每拍扫描迁移 canceled）\n", taskCancelEvent, scope, name)
	fmt.Fprintln(a.out, strings.TrimSpace(string(raw)))
	fmt.Fprintln(a.out, "查看：corectl task "+scope+"/"+name)
	return 0
}

func (a *app) cmdReport(args []string) int {
	fs := flag.NewFlagSet("report", flag.ContinueOnError)
	fs.SetOutput(a.err)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	var sum schemeSummary
	var systems objectList
	var wfs workflowsResp
	var runs runsResp
	var inbox struct {
		Count   int            `json:"count"`
		Classes map[string]int `json:"classes"`
	}
	var events eventsResp

	var missing []string
	if err := a.getJSON("/v1/scheme", nil, &sum); err != nil {
		missing = append(missing, "scheme")
	}
	if err := a.getJSON("/v1/objects/ServiceDesc", nil, &systems); err != nil {
		missing = append(missing, "systems")
	}
	if err := a.getJSON("/v1/workflows", nil, &wfs); err != nil {
		missing = append(missing, "workflows")
	}
	if err := a.getJSON("/v1/workflows/runs", nil, &runs); err != nil {
		missing = append(missing, "runs")
	}
	if err := a.getJSON("/v1/inbox", nil, &inbox); err != nil {
		missing = append(missing, "inbox")
	}
	q := url.Values{}
	q.Set("limit", "100")
	if err := a.getJSON("/v1/events", q, &events); err != nil {
		missing = append(missing, "events")
	}

	if a.jsonOut {
		return a.emit(mustJSON(map[string]any{
			"scheme": sum, "systems": systems, "workflows": wfs,
			"runs": runs, "inbox": inbox, "events": events, "missing": missing,
		}))
	}

	running, failed := 0, 0
	for _, r := range runs.Runs {
		switch r.Status {
		case "running", "triggered":
			running++
		case "failed":
			failed++
		}
	}
	fmt.Fprintln(a.out, "关注报告（CLI 摘要）")
	fmt.Fprintf(a.out, "系统：%d 个 ServiceDesc 在册\n", systems.Count)
	fmt.Fprintf(a.out, "工作流：%d 条 · 运行中 %d · 24h 触发 %d · 失败 %d\n",
		wfs.Count, running, totalRuns24h(wfs), failed)
	fmt.Fprintf(a.out, "待处理：%d 项", inbox.Count)
	if len(inbox.Classes) > 0 {
		var parts []string
		for k, v := range inbox.Classes {
			parts = append(parts, fmt.Sprintf("%s %d", k, v))
		}
		fmt.Fprintf(a.out, "（%s）", strings.Join(parts, " / "))
	}
	fmt.Fprintln(a.out)

	fmt.Fprintln(a.out, "异常（最近 100 条事件中的 warn/error/fatal）：")
	n := 0
	for _, e := range events.Events {
		if e.Level != "warn" && e.Level != "error" && e.Level != "fatal" {
			continue
		}
		fmt.Fprintf(a.out, "  [%s] %s · %s · %s\n", e.Level, e.Type, e.Involved, e.TS)
		n++
		if n >= 5 {
			break
		}
	}
	if n == 0 {
		fmt.Fprintln(a.out, "  （未发现 warn/error/fatal）")
	}
	if len(missing) > 0 {
		fmt.Fprintf(a.out, "数据源缺失：%s（摘要降级，未阻塞）\n", strings.Join(missing, ","))
	}
	fmt.Fprintln(a.out, "明细：corectl inbox | corectl events | corectl runs | corectl tick")
	return 0
}

func totalRuns24h(w workflowsResp) int {
	t := 0
	for _, x := range w.Workflows {
		t += x.Runs24h
	}
	return t
}

// ---------- 写入命令 ----------

type applyBody struct {
	Kind            string   `json:"kind"`
	Scope           string   `json:"scope"`
	Name            string   `json:"name"`
	ResourceVersion int64    `json:"resourceVersion,omitempty"`
	Content         string   `json:"content"`
	FieldManager    string   `json:"fieldManager,omitempty"`
	DryRun          bool     `json:"dryRun,omitempty"`
	Subject         *subject `json:"subject,omitempty"`
}

func (a *app) cmdApply(args []string) int {
	fs := flag.NewFlagSet("apply", flag.ContinueOnError)
	fs.SetOutput(a.err)
	file := fs.String("f", "", "文档文件（- = stdin）")
	kind := fs.String("kind", "", "对象 kind（缺省读 frontmatter）")
	scope := fs.String("scope", "", "对象 scope（缺省读 frontmatter）")
	name := fs.String("name", "", "对象 name（缺省读 frontmatter）")
	rv := fs.Int64("rv", 0, "乐观锁 resourceVersion")
	dry := fs.Bool("dry-run", false, "只准入校验，不落库")
	fm := fs.String("field-manager", "corectl", "fieldManager（审计）")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *file == "" {
		fmt.Fprintln(a.err, "用法：corectl apply -f <file> [--kind K --scope S --name N] [--dry-run]")
		return 2
	}
	var content []byte
	var err error
	if *file == "-" {
		content, err = io.ReadAll(io.LimitReader(os.Stdin, 4<<20))
	} else {
		content, err = os.ReadFile(*file)
	}
	if err != nil {
		return fail(a, err)
	}
	fk, fsco, fn := parseDocHeader(string(content))
	if *kind == "" {
		*kind = fk
	}
	if *scope == "" {
		*scope = fsco
	}
	if *name == "" {
		*name = fn
	}
	if *kind == "" || *scope == "" || *name == "" {
		fmt.Fprintln(a.err, "错误：kind/scope/name 不全（frontmatter 未提供时显式传 --kind/--scope/--name）")
		return 2
	}
	body := applyBody{
		Kind: *kind, Scope: *scope, Name: *name, ResourceVersion: *rv,
		Content: string(content), FieldManager: *fm, DryRun: *dry, Subject: &a.subject,
	}
	raw, err := a.request(http.MethodPost, "/v1/apply", body)
	if err != nil {
		return fail(a, err)
	}
	return a.emit(raw)
}

// parseDocHeader 读 YAML frontmatter 的 kind/docType/scope/name（够用即可，不引 YAML 依赖）。
func parseDocHeader(content string) (kind, scope, name string) {
	in := false
	for _, line := range strings.Split(content, "\n") {
		t := strings.TrimSpace(line)
		if t == "---" {
			if !in {
				in = true
				continue
			}
			break
		}
		if !in {
			continue
		}
		k, v, ok := strings.Cut(t, ":")
		if !ok {
			continue
		}
		v = strings.Trim(strings.TrimSpace(v), `"'`)
		switch strings.TrimSpace(k) {
		case "kind":
			kind = v
		case "docType":
			if kind == "" {
				kind = v
			}
		case "scope":
			scope = v
		case "name":
			name = v
		}
	}
	return
}

type eventBody struct {
	Type     string   `json:"type"`
	Involved string   `json:"involved,omitempty"`
	Level    string   `json:"level,omitempty"`
	DedupKey string   `json:"dedupKey,omitempty"`
	Scope    string   `json:"scope,omitempty"` // metadata.scope 四层词显式声明（BUG-010；空=服务端按 involved 派生）
	Subject  *subject `json:"subject,omitempty"`
}

func (a *app) cmdEvent(args []string) int {
	if len(args) == 0 {
		fmt.Fprintln(a.err, "用法：corectl event publish --type T [--involved k/s/n] [--level L] [--dedup K]")
		return 2
	}
	switch args[0] {
	case "publish":
		return a.cmdEventPublish(args[1:])
	case "list":
		return a.cmdEvents(args[1:])
	default:
		fmt.Fprintf(a.err, "未知 event 子命令：%s（publish|list）\n", args[0])
		return 2
	}
}

func (a *app) cmdEventPublish(args []string) int {
	fs := flag.NewFlagSet("event publish", flag.ContinueOnError)
	fs.SetOutput(a.err)
	typ := fs.String("type", "", "事件类型（Scheme.eventTypes 注册名）")
	involved := fs.String("involved", "", "involvedRef kind/scope/name")
	level := fs.String("level", "", "level（缺省=注册表 defaultLevel）")
	dedup := fs.String("dedup", "", "去重键（缺省=事件 id）")
	scope := fs.String("scope", "", "metadata.scope 四层词（builtin|global|systems/<s>|systems/<s>/instances/<i>；缺省按 involved 派生）")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *typ == "" {
		fmt.Fprintln(a.err, "错误：缺 --type")
		return 2
	}
	raw, err := a.request(http.MethodPost, "/v1/events", eventBody{
		Type: *typ, Involved: *involved, Level: *level, DedupKey: *dedup, Scope: *scope, Subject: &a.subject,
	})
	if err != nil {
		return fail(a, err)
	}
	return a.emit(raw)
}

// cmdFind 上下文相关性检索（/v1/context/find——B2 判据的 CLI 面：score 恒 0-100）。
func (a *app) cmdFind(args []string) int {
	fs := flag.NewFlagSet("find", flag.ContinueOnError)
	fs.SetOutput(a.err)
	tags := fs.String("tags", "", "Tag 过滤（逗号分隔，形如 doc:acceptance 或裸 tag）")
	k := fs.Int("k", 10, "TopK（≤50）")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	q := url.Values{}
	if rest := fs.Args(); len(rest) > 0 {
		q.Set("q", strings.Join(rest, " "))
	}
	if *tags != "" {
		q.Set("tags", *tags)
	}
	q.Set("k", fmt.Sprint(*k))
	path := "/v1/context/find"
	if enc := q.Encode(); enc != "" {
		path += "?" + enc
	}
	raw, err := a.request(http.MethodGet, path, nil)
	if err != nil {
		return fail(a, err)
	}
	return a.emit(raw)
}

// cmdTrace 文档 trace 双向聚合（/v1/trace——R-DOC-9 / B11：implements/implemented_by/verified_by）。
func (a *app) cmdTrace(args []string) int {
	fs := flag.NewFlagSet("trace", flag.ContinueOnError)
	fs.SetOutput(a.err)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	rest := fs.Args()
	if len(rest) != 1 {
		fmt.Fprintln(a.err, "用法：corectl trace <kind>/<scope>/<name>")
		return 2
	}
	parts := strings.SplitN(rest[0], "/", 3)
	if len(parts) != 3 {
		fmt.Fprintln(a.err, "错误：键须为 kind/scope/name 三段（scope 可含多段时用 kind = 前缀、末段为 name 亦可）")
		return 2
	}
	raw, err := a.request(http.MethodGet, "/v1/trace/"+parts[0]+"/"+parts[1]+"/"+parts[2], nil)
	if err != nil {
		return fail(a, err)
	}
	return a.emit(raw)
}

func (a *app) cmdTick(args []string) int {
	fs := flag.NewFlagSet("tick", flag.ContinueOnError)
	fs.SetOutput(a.err)
	involved := fs.String("involved", "", "involvedRef kind/scope/name")
	dedup := fs.String("dedup", "", "去重键（缺省 cli-tick:<unixnano>）")
	level := fs.String("level", "info", "level")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *dedup == "" {
		*dedup = fmt.Sprintf("cli-tick:%d", time.Now().UnixNano())
	}
	raw, err := a.request(http.MethodPost, "/v1/events", eventBody{
		Type: "ReconcileTick", Involved: *involved, Level: *level, DedupKey: *dedup, Subject: &a.subject,
	})
	if err != nil {
		return fail(a, err)
	}
	if a.jsonOut {
		return a.emit(raw)
	}
	fmt.Fprintln(a.out, "已发布系统事件 ReconcileTick（触发 core-workflow 对账 run）")
	fmt.Fprintln(a.out, strings.TrimSpace(string(raw)))
	fmt.Fprintln(a.out, "查看：corectl runs / corectl events --type ReconcileTick")
	return 0
}

func (a *app) cmdLoop(args []string) int {
	fs := flag.NewFlagSet("loop", flag.ContinueOnError)
	fs.SetOutput(a.err)
	dedup := fs.String("dedup", "", "去重键（缺省 cli-loop:<unixnano>）")
	level := fs.String("level", "info", "level")
	if err := fs.Parse(args); err != nil {
		return 2
	}
	if *dedup == "" {
		*dedup = fmt.Sprintf("cli-loop:%d", time.Now().UnixNano())
	}
	raw, err := a.request(http.MethodPost, "/v1/events", eventBody{
		Type: "LoopRequested", Level: *level, DedupKey: *dedup, Subject: &a.subject,
	})
	if err != nil {
		return fail(a, err)
	}
	if a.jsonOut {
		return a.emit(raw)
	}
	fmt.Fprintln(a.out, "已发布系统事件 LoopRequested（事件→wake 桥→宿主拉起自循环）")
	fmt.Fprintln(a.out, strings.TrimSpace(string(raw)))
	fmt.Fprintln(a.out, "查看：corectl wakes | corectl events --type LoopRequested")
	return 0
}

func (a *app) cmdVersion(args []string) int {
	fs := flag.NewFlagSet("version", flag.ContinueOnError)
	fs.SetOutput(a.err)
	if err := fs.Parse(args); err != nil {
		return 2
	}
	var sum schemeSummary
	err := a.getJSON("/v1/scheme", nil, &sum)
	if a.jsonOut {
		return a.emit(mustJSON(map[string]any{"cli": cliVersion, "scheme": sum, "error": errString(err)}))
	}
	fmt.Fprintf(a.out, "corectl %s\n", cliVersion)
	if err != nil {
		fmt.Fprintf(a.out, "cored: 不可达（%v）\n", err)
		return 0
	}
	fmt.Fprintf(a.out, "cored scheme: %s %s · kinds=%d\n", sum.APIVersion, sum.Version, sum.KindCount)
	return 0
}

func errString(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

func mustJSON(v any) []byte {
	b, err := json.Marshal(v)
	if err != nil {
		return []byte(`{"error":"marshal"}`)
	}
	return b
}
