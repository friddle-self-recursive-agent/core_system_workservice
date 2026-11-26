// corectl_tasks.go —— corectl 任务/审批/报告命令族（R-ENG-7 行数收口纯搬家：tasks/tasks
// 详情/submit/cancel/report/inbox/proposals/sessions/version 与 errString/mustJSON 小件——
// 零语义变化，main.go 拆分伴生件）。
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

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
