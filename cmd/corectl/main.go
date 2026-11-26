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
  delete <kind>/<scope>/<name> [--rv N] --yes   对象删除（写面闸门同 apply；三账留痕）
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
	case "delete":
		return a.cmdDelete(args)
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
	}
	// 主体头随行不限 body（DELETE 无 body 但删除是写面动作——主体判定必需，匿名照判拒绝）。
	if sub, err := json.Marshal(a.subject); err == nil {
		req.Header.Set("X-Core-Subject", string(sub))
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
