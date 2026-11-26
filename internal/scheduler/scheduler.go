// scheduler.go —— TaskScheduler tick 循环（多任务并发调度器主体，薄组件）。
//
// 设计依据：runtime/task-scheduler.md §1-§4 与 concepts/task-scheduler.md。拓扑：
// Task 对象（docs DB）→ 每拍读全部 Task（跨 scope）→ 依赖/冲突/上限（plan.go 纯函数）
// → 写 status（repository 系统写，fieldManager=task-scheduler——P2 status 单一写者）+
// TaskStateChanged 事件（引擎内部源 source=core，dedup=task:<id>:<phase>）→ POST 宿主
// /task/spawn（每任务一个 dsh session；与 wake 桥同族：固定载荷不进 shell、loopback、超时 5s）。
//
// 边界（与 wake 桥同款红线）：
//   - 只在部署侧显式配置 CORE_SYSTEM_TASK_SPAWN_ADDR 时拉起；缺席=spawn 关断保持 ready，
//     lastError 可见 + 退避重试（不热循环刷日志/刷库）。
//   - 调度器只做「选任务 + 限并发 + 拉起 + 状态」；执行/验证/落地复用既有 Agent/Apply/Verify。
//   - spawn 失败不翻转调度面（任务保持 ready 下拍退避重试）；写面乐观锁冲突=下拍重读重试。
//   - 配置=scheme runtime.taskScheduler 每拍热读（Apply 即时生效）；enabled=false 本拍零动作。
package scheduler

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/opslog"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
	"gopkg.in/yaml.v3"
)

// 装配锚常量（与 scheme.yaml runtime.taskScheduler / eventTypes 行同源）。
const (
	// TaskKind 调度对象类型锚（scheme kind: Task）。
	TaskKind = "Task"
	// EventTypeStateChanged Task 状态变迁事件类型（scheme.eventTypes 注册名——注册闸门键）。
	EventTypeStateChanged = "TaskStateChanged"
	// EventTypeCancel 取消请求事件类型（eventLog 扫描键；involved=Task/<scope>/<name>）。
	EventTypeCancel = "TaskCancelRequested"
	// eventTypeAsk Apply 关口冲突待决事件类型（ApprovalNeeded——人注意族，注册闸门前置）。
	eventTypeAsk = "ApprovalNeeded"
	// FieldManager Task.status 系统写面主体（P2：status 只由调度器/执行体写）。
	FieldManager = "task-scheduler"
	// EnvSpawnAddr 宿主 spawn 桥地址 env（127.0.0.1 随机端口；缺席=spawn 关断）。
	EnvSpawnAddr = "CORE_SYSTEM_TASK_SPAWN_ADDR"
	// EnvTickMs 拍间隔 env（毫秒；缺省 5000）。
	EnvTickMs = "CORE_SYSTEM_SCHEDULER_TICK_MS"
	// DefaultTick 缺省拍间隔。
	DefaultTick = 5 * time.Second
	// SpawnTimeout 单次 spawn POST 超时（与 wake 桥同值）。
	SpawnTimeout = 5 * time.Second
	// BackoffBase / BackoffMax spawn 失败退避（指数，封顶——退避重试不热循环）。
	BackoffBase = 30 * time.Second
	BackoffMax  = 5 * time.Minute
)

// CancelSource 取消事件归档读口面（opslog.SQLiteStore 满足；接口化便于测试注入）。
type CancelSource interface {
	QueryEvents(q opslog.EventQuery) ([]opslog.EventLog, error)
}

// Options 装配项（零值可用：tick=缺省、spawn 关断、stderr 注记、time.Now）。
type Options struct {
	Tick      time.Duration // 拍间隔（缺省 DefaultTick；env CORE_SYSTEM_SCHEDULER_TICK_MS 装配层可覆写）
	SpawnAddr string        // 宿主 spawn 桥地址（host:port 或 http://host:port；空=关断）
	Sessions  SessionSource // 会话终结信号读口面（G-08 done 判定——nil=判定关断既有行为面）
	Now       func() time.Time
	Logf      func(format string, args ...any)
}

// Scheduler 任务调度器（cored 装配层持有；Run 阻塞直到 ctx 取消）。
type Scheduler struct {
	repo     repository.Repository
	bus      *eventbus.Bus
	sch      *scheme.Scheme
	cancels  CancelSource
	applies  ApplyLogSource
	sessions SessionSource // 会话终结信号（G-08 done 判定——nil=关断）

	url  string // spawn 桥 URL（空=关断）
	hc   *http.Client
	tick time.Duration
	now  func() time.Time
	logf func(format string, args ...any)

	mu                   sync.Mutex
	readySince           map[string]time.Time // ready 起算（等待时长排序输入）
	leases               map[string]time.Time // 租约起点（taskId → 本轮拉起时刻）
	backoff              map[string]time.Time // spawn 退避（taskId → 下次允许尝试时刻）
	spawnFails           map[string]int       // spawn 连败计数（退避步进；成功清零）
	cancelPending        map[string]bool      // 已请求取消、等自然结束的任务
	cancelSeen           map[string]bool      // 已处理取消事件 ID（归档游标幂等）
	conflictSeen         map[string]bool      // 已处置 apply 冲突行（ts+obj+fm 复合键——BUG-013）
	conflictResolvedSeen map[string]bool      // 已处置裁决事件 ID
	conflictHold         map[string]bool      // Apply 冲突挂起（等 TaskConflictResolved，不重试）
	conflictPairs        map[string]string    // 挂起对：当前任务 → 在先任务
}

// New 构建调度器（cancels/applies 可为 nil=取消扫描/Apply 关口冲突拦截关断；spawn 地址空=拉起关断保持 ready）。
func New(repo repository.Repository, bus *eventbus.Bus, sch *scheme.Scheme, cancels CancelSource, applies ApplyLogSource, opts Options) *Scheduler {
	tick := opts.Tick
	if tick <= 0 {
		if v := strings.TrimSpace(os.Getenv(EnvTickMs)); v != "" {
			if n, err := strconv.Atoi(v); err == nil && n > 0 {
				tick = time.Duration(n) * time.Millisecond
			}
		}
	}
	if tick <= 0 {
		tick = DefaultTick
	}
	now := opts.Now
	if now == nil {
		now = time.Now
	}
	logf := opts.Logf
	if logf == nil {
		logf = func(format string, args ...any) { fmt.Fprintf(os.Stderr, "cored: scheduler "+format+"\n", args...) }
	}
	url := ""
	if addr := strings.TrimSpace(opts.SpawnAddr); addr != "" {
		if !strings.HasPrefix(addr, "http://") && !strings.HasPrefix(addr, "https://") {
			addr = "http://" + addr
		}
		url = strings.TrimRight(addr, "/") + "/task/spawn"
	}
	return &Scheduler{
		repo: repo, bus: bus, sch: sch, cancels: cancels, applies: applies, sessions: opts.Sessions,
		url: url, hc: &http.Client{Timeout: SpawnTimeout}, tick: tick, now: now, logf: logf,
		readySince: map[string]time.Time{}, leases: map[string]time.Time{},
		backoff: map[string]time.Time{}, spawnFails: map[string]int{},
		cancelPending: map[string]bool{}, cancelSeen: map[string]bool{},
		conflictSeen: map[string]bool{}, conflictResolvedSeen: map[string]bool{},
		conflictHold: map[string]bool{}, conflictPairs: map[string]string{},
	}
}

// SpawnEnabled spawn 桥装配态（stderr 装配锚报告用）。
func (s *Scheduler) SpawnEnabled() bool { return s != nil && s.url != "" }

// Interval 拍间隔（stderr 装配锚报告用）。
func (s *Scheduler) Interval() time.Duration {
	if s == nil {
		return 0
	}
	return s.tick
}

// Run 主循环（ctx 取消即退场；首拍即跑——装配后无需等满一个间隔）。
func (s *Scheduler) Run(ctx context.Context) {
	s.Tick(ctx)
	tk := time.NewTicker(s.tick)
	defer tk.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tk.C:
			s.Tick(ctx)
		}
	}
}

// Tick 每拍：热读配置 → 读全部 Task（跨 scope）→ 取消扫描 → 租约回收 → 依赖/冲突/上限
// → 写面 + 事件 → spawn。任一读面失败=本拍跳过（可丢语义，下拍重试）。
func (s *Scheduler) Tick(ctx context.Context) {
	cfg := ConfigOf(s.sch)
	if !cfg.Enabled {
		return // scheme 热读：Apply 关断=调度器零动作（缺省不改行为面）
	}
	objs, err := s.repo.List(ctx, TaskKind, "")
	if err != nil {
		s.logf("枚举 Task 对象失败（本拍跳过）: %v", err)
		return
	}
	now := s.now()
	views := make([]TaskView, 0, len(objs))
	byObj := map[string]*repository.Object{}
	byView := map[string]TaskView{}
	byPhase := map[string]string{}
	bySession := map[string]string{} // 任务 id → 现行 agentSession（G-08 done 判定匹配键）
	for _, o := range objs {
		spec, st, ok := Parse(o.Content)
		if !ok {
			continue // 非 Task 形（无 status.phase）——读面宽松跳过，不放大声明形变
		}
		id := o.Scope + "/" + o.Name
		v := TaskView{
			ID: id, Scope: o.Scope, Name: o.Name, Phase: st.Phase, Priority: spec.Priority,
			DependsOn: spec.DependsOn, Touches: spec.Touches, Service: spec.ScopeRef,
			Attempts: st.Attempts, Workspace: spec.Workspace, Agent: spec.Agent,
			MaxConcurrent: spec.MaxConcurrent, Timeout: spec.Timeout,
		}
		views = append(views, v)
		byObj[id] = o
		byView[id] = v
		byPhase[id] = st.Phase
		bySession[id] = st.AgentSession
	}
	s.trackReady(byPhase, now)
	s.scanCancels()
	s.applyCancels(ctx, byObj, byPhase)
	s.recoverLeases(ctx, byObj, byView, cfg, now)
	// 会话终结→done 判定（G-08——先于冲突链/dispatch：done 任务本拍即退出参选，
	// agentSession 回填真实 sessionId；nil 源=判定关断既有行为面零翻转）。
	s.observeSessionEnds(ctx, byObj, byPhase, bySession)
	// Apply 关口冲突链（BUG-013）：扫描 409 冲突行 → 实质冲突判定 → 上抛/挂起；裁决事件两路
	// 落面；让位复位。在 Plan 之前——挂起任务不参选不重试（恢复仍由本拍裁决/复位承担）。
	s.scanApplyConflicts(ctx, byView, cfg, byObj)
	s.resolveConflicts(ctx, byObj, byView)
	s.resumeHeld(ctx, byObj, byView)
	s.dispatch(ctx, byObj, views, cfg, now)
}

// trackReady ready 起算维护（等待时长排序输入；非 ready 清除——重启后从首拍起算）。
func (s *Scheduler) trackReady(byPhase map[string]string, now time.Time) {
	s.mu.Lock()
	defer s.mu.Unlock()
	for id, phase := range byPhase {
		if phase == PhaseReady {
			if _, ok := s.readySince[id]; !ok {
				s.readySince[id] = now
			}
			continue
		}
		delete(s.readySince, id)
	}
	for id := range s.readySince {
		if _, ok := byPhase[id]; !ok {
			delete(s.readySince, id) // 对象已不在册（删除/改名）——影子清理
		}
	}
}

// scanCancels 扫 eventLog 中 TaskCancelRequested（involved=Task/<scope>/<name>）→ 入待取消集。
// 归档游标+事件 ID 去重：同一事件只处置一次；扫描失败=本拍跳过（可丢语义，下拍续扫）。
func (s *Scheduler) scanCancels() {
	if s.cancels == nil {
		return
	}
	rows, err := s.cancels.QueryEvents(opslog.EventQuery{Type: EventTypeCancel, Limit: opslog.MaxEventLimit})
	if err != nil {
		s.logf("取消事件扫描失败（本拍跳过）: %v", err)
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, r := range rows {
		if s.cancelSeen[r.EventID] {
			continue
		}
		s.cancelSeen[r.EventID] = true
		id := strings.TrimPrefix(strings.TrimSpace(r.Involved), TaskKind+"/")
		if id == "" || id == r.Involved { // involved 非本域词形——忽略（不冒充全量处置）
			continue
		}
		s.cancelPending[id] = true
	}
}

// applyCancels 处置待取消集：非 running 直接 canceled；running/verifying 先注记等自然结束
// （lastAction 写明，幂等免重复写）；终态=落账出集。
func (s *Scheduler) applyCancels(ctx context.Context, byObj map[string]*repository.Object, byPhase map[string]string) {
	s.mu.Lock()
	pending := make([]string, 0, len(s.cancelPending))
	for id := range s.cancelPending {
		pending = append(pending, id)
	}
	s.mu.Unlock()
	for _, id := range pending {
		o, ok := byObj[id]
		if !ok {
			continue // 任务不在册（未落库/已删）——留集待下拍（事件先行于对象词形）
		}
		phase := byPhase[id]
		switch phase {
		case PhaseDone, PhaseFailed, PhaseCanceled:
			s.mu.Lock()
			delete(s.cancelPending, id) // 已终态：取消请求落账完成，出集
			s.mu.Unlock()
		case PhaseRunning, PhaseVerifying:
			action := "已请求取消：等待自然结束后迁移 canceled（TaskCancelRequested）"
			if err := s.writeStatus(ctx, o, statusPatch{LastAction: &action}, "scheduler:cancel-request"); err != nil {
				s.logf("任务 %s 取消注记写面失败（下拍重试）: %v", id, err)
			}
		default:
			action := "已取消（TaskCancelRequested）"
			cleared := ""
			if err := s.writeStatus(ctx, o, statusPatch{
				Phase: PhaseCanceled, LastAction: &action, LastError: &cleared,
			}, "scheduler:cancel"); err != nil {
				s.logf("任务 %s 取消写面失败（下拍重试）: %v", id, err)
				continue
			}
			s.publishState(id, PhaseCanceled)
			s.releaseHold(id) // 取消优先于 Apply 冲突挂起（挂起对随之解除）
			s.mu.Lock()
			delete(s.cancelPending, id)
			s.mu.Unlock()
		}
	}
}

// recoverLeases 租约回收：running 超过生效租约无进展 → ready（attempts+1）/failed（超上限）。
// 租约起点=进程内本轮拉起时刻（重启后零值=首拍观察起算，不判过期——保守不误杀）。
func (s *Scheduler) recoverLeases(ctx context.Context, byObj map[string]*repository.Object, byView map[string]TaskView, cfg Config, now time.Time) {
	s.mu.Lock()
	started := make(map[string]time.Time, len(s.leases))
	for id, at := range s.leases {
		started[id] = at
	}
	s.mu.Unlock()
	for id, at := range started {
		v, ok := byView[id]
		if !ok {
			s.mu.Lock()
			delete(s.leases, id) // 对象不在册——租约影子清理
			s.mu.Unlock()
			continue
		}
		if v.Phase != PhaseRunning {
			s.mu.Lock()
			delete(s.leases, id) // 已离场（verifying/done/...）——租约终止
			s.mu.Unlock()
			continue
		}
		d := Recover(v, at, cfg, now)
		if d.Phase == "" {
			continue
		}
		o := byObj[id]
		patch := statusPatch{Phase: d.Phase, LastError: &d.Reason}
		if d.Phase == PhaseReady {
			att := v.Attempts + 1
			patch.Attempts = &att
		}
		if err := s.writeStatus(ctx, o, patch, "scheduler:lease-recover"); err != nil {
			s.logf("任务 %s 租约回收写面失败（下拍重试）: %v", id, err)
			continue
		}
		s.publishState(id, d.Phase)
		s.logf("任务 %s %s（%s）", id, d.Phase, d.Reason)
		s.mu.Lock()
		delete(s.leases, id)
		s.mu.Unlock()
	}
}

// dispatch 依赖/冲突/上限判定 → 写面 + 事件 → spawn（写 status=running 先于 POST——§3 步骤 5）。
func (s *Scheduler) dispatch(ctx context.Context, byObj map[string]*repository.Object, views []TaskView, cfg Config, now time.Time) {
	// Apply 冲突挂起集（等裁决不参选——Plan 视图与处置循环双排除）。
	s.mu.Lock()
	held := make(map[string]bool, len(s.conflictHold))
	for id := range s.conflictHold {
		held[id] = true
	}
	s.mu.Unlock()
	planViews := make([]TaskView, 0, len(views))
	for _, v := range views {
		if held[v.ID] {
			continue
		}
		planViews = append(planViews, v)
	}
	// readySince 注入视图（排序输入——进程内观察值）。
	s.mu.Lock()
	for i := range planViews {
		planViews[i].ReadySince = s.readySince[planViews[i].ID]
	}
	s.mu.Unlock()
	for id, d := range Plan(planViews, cfg, now) {
		if held[id] {
			continue // 挂起等裁决——不重试不迁移（resolveConflicts 专面接手）
		}
		o := byObj[id]
		if o == nil {
			continue
		}
		cur := TaskPhase(o.Content)
		// 1) phase 写面（queued/blocked→ready、→blocked；ready 保持=仅排队原因更新）。
		if d.Phase != "" && d.Phase != cur && !CanMigrate(cur, d.Phase) {
			continue // 闭合集外迁移拒绝（对象被并发改写等竞态——下拍重读重判）
		}
		if d.Phase != "" && d.Phase != cur {
			cleared := ""
			patch := statusPatch{Phase: d.Phase, LastError: &cleared, LastAction: strPtr(readyAction(d))}
			if d.Reason != "" {
				patch.LastError = &d.Reason
			}
			if err := s.writeStatus(ctx, o, patch, "scheduler:phase:"+d.Phase); err != nil {
				s.logf("任务 %s 迁移 %s→%s 写面失败（下拍重试）: %v", id, cur, d.Phase, err)
				continue
			}
			s.publishState(id, d.Phase)
			cur = d.Phase
		} else if d.Reason != "" && d.Reason != TaskLastError(o.Content) {
			// 保持原 phase 仅原因更新（排队原因可见不静默——幂等免写由等值比对承担）。
			if err := s.writeStatus(ctx, o, statusPatch{LastError: &d.Reason}, "scheduler:queue-reason"); err != nil {
				s.logf("任务 %s 排队原因写面失败（下拍重试）: %v", id, err)
			}
		}
		// 2) 拉起：ready → 租约 → spawn。
		if d.Start {
			if cur != PhaseReady {
				continue // 迁移写面未落（冲突/失败）——本拍不拉起（状态单一写者不与库内失配）
			}
			if !s.spawnAllowed(id, now) {
				continue // spawn 退避中（不热循环）
			}
			s.spawn(ctx, o, id, now)
		}
	}
}

// spawnAllowed spawn 退避闸（BackoffBase 起指数步进、BackoffMax 封顶——失败重试不刷日志）。
func (s *Scheduler) spawnAllowed(id string, now time.Time) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	if at, ok := s.backoff[id]; ok && now.Before(at) {
		return false
	}
	return true
}

// spawn 单任务拉起：写 status=running + 发 TaskStateChanged → 生成任务文件 → POST
// /task/spawn → 租约/回执落面（§3 步骤 5 次序——先占额再拉起）。失败=回退 ready +
// lastError 可见 + 退避（§4 宿主 spawn 桥契约）。
func (s *Scheduler) spawn(ctx context.Context, o *repository.Object, id string, now time.Time) {
	// 1) 先占额：running 写面 + 事件（失败=不拉起，下拍重试——状态单一写者不悬空）。
	action := "拉起中（spawn）"
	cleared := ""
	if err := s.writeStatus(ctx, o, statusPatch{Phase: PhaseRunning, LastAction: &action, LastError: &cleared}, "scheduler:spawn"); err != nil {
		s.logf("任务 %s running 写面失败（本拍不拉起）: %v", id, err)
		return
	}
	s.publishState(id, PhaseRunning)
	// 2) POST 宿主 spawn 桥。
	spec, _, _ := Parse(o.Content)
	taskPath, err := writeTaskFile(id, spec)
	if err != nil {
		s.spawnFail(ctx, o, id, "任务文件生成失败: "+err.Error(), now)
		return
	}
	payload, err := json.Marshal(map[string]any{
		"taskId":   id,
		"taskPath": taskPath,
		"workspace": map[string]any{
			"mode": spec.Workspace.Mode, "path": spec.Workspace.Base, "branch": spec.Workspace.Branch, // path=声明基线（worktree 根由宿主解析）
		},
		"agent": spec.Agent,
	})
	if err != nil {
		s.spawnFail(ctx, o, id, "spawn 载荷编码失败: "+err.Error(), now)
		return
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(payload))
	if err != nil {
		s.spawnFail(ctx, o, id, "spawn 请求构建失败: "+err.Error(), now)
		return
	}
	req.Header.Set("Content-Type", "application/json")
	resp, err := s.hc.Do(req)
	if err != nil {
		s.spawnFail(ctx, o, id, "spawn 桥调用失败: "+err.Error(), now)
		return
	}
	defer resp.Body.Close()
	var ack struct {
		OK          bool   `json:"ok"`
		Status      string `json:"status"`
		SessionHint string `json:"sessionHint"`
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		s.spawnFail(ctx, o, id, fmt.Sprintf("spawn 桥拒绝（HTTP %d）", resp.StatusCode), now)
		return
	}
	_ = json.NewDecoder(resp.Body).Decode(&ack) // 回执宽松（无 body=受理成功语义仍成立）
	// 3) 受理 → 租约起点 + sessionHint 回填 agentSession（§4：会话 ID 执行体心跳/回执）。
	s.mu.Lock()
	s.leases[id] = now
	s.spawnFails[id] = 0
	delete(s.backoff, id)
	s.mu.Unlock()
	if hint := strings.TrimSpace(ack.SessionHint); hint != "" {
		fresh, err := s.repo.Get(ctx, TaskKind, o.Scope, o.Name)
		if err == nil {
			if err := s.writeStatus(ctx, fresh, statusPatch{AgentSession: &hint}, "scheduler:agent-session"); err == nil {
				*o = *fresh // 本拍后续写面续用新 rv
			}
		}
	}
	s.logf("任务 %s 已拉起（taskPath=%s status=%s session=%s）", id, taskPath, ack.Status, ack.SessionHint)
}

// spawnFail spawn 失败落面：回退 ready + lastError 可见 + 退避步进（指数封顶；
// 退避闸保证失败重试不热循环刷日志/刷库）。
func (s *Scheduler) spawnFail(ctx context.Context, o *repository.Object, id, reason string, now time.Time) {
	s.mu.Lock()
	fails := s.spawnFails[id] + 1
	s.spawnFails[id] = fails
	backoff := BackoffBase << (fails - 1)
	if backoff > BackoffMax || backoff <= 0 {
		backoff = BackoffMax
	}
	s.backoff[id] = now.Add(backoff)
	s.mu.Unlock()
	lastErr := fmt.Sprintf("spawn 失败（第 %d 次，%s 后重试）：%s", fails, backoff, reason)
	action := "拉起失败，保持 ready 退避重试"
	if err := s.writeStatus(ctx, o, statusPatch{Phase: PhaseReady, LastError: &lastErr, LastAction: &action}, "scheduler:spawn-fail"); err != nil {
		s.logf("任务 %s spawn 失败写面失败（下拍重试）: %v", id, err)
	}
	s.publishState(id, PhaseReady) // 回队列（事件与迁移同拍口径）
	s.logf("任务 %s spawn 失败保持 ready（%s）", id, lastErr)
}

// writeStatus Task.status 系统写面（fieldManager=task-scheduler；乐观锁冲突=调用方下拍重读
// 重试；等值=零改动免写——幂等不刷库）。
func (s *Scheduler) writeStatus(ctx context.Context, o *repository.Object, p statusPatch, reasonRef string) error {
	out, changed, err := WriteStatus(o.Content, p)
	if err != nil {
		return err
	}
	if !changed {
		return nil
	}
	next := *o
	next.Content = out
	if _, _, err := s.repo.Write(ctx, &next, FieldManager, reasonRef); err != nil {
		return err
	}
	o.Content = out // 本拍内续用（rv 由仓库推进；再写前需重取——spawn 回执路径已处理）
	return nil
}

// publishState TaskStateChanged（引擎内部源 source=core，dedup=task:<id>:<phase>——
// 注册闸门前置：未注册不发布，与巡检/会话活性同范式）。
func (s *Scheduler) publishState(id, phase string) {
	et, err := s.sch.LookupEventType(EventTypeStateChanged)
	if err != nil {
		return // 类型未注册（旧声明面）——零事件，缺口可见
	}
	_ = s.bus.Publish(&eventbus.Event{
		Type: EventTypeStateChanged, Source: "core", Involved: TaskKind + "/" + id,
		DedupKey: "task:" + id + ":" + phase, Level: et.DefaultLevel,
	})
}

// writeTaskFile 生成任务文件（Task.spec 摘要 + touches + workspace + 验证要求 + acceptance
// 提示）落部署侧临时目录（os.TempDir()；不进公开仓——runtime/task-scheduler.md §4）。
func writeTaskFile(id string, spec Spec) (string, error) {
	doc := map[string]any{
		"taskId":    id,
		"title":     spec.Title,
		"summary":   spec.Summary, // 验收=summary 声明（acceptance 内嵌口径——Task.spec 无独立验收字段）
		"scopeRef":  spec.ScopeRef,
		"dependsOn": spec.DependsOn,
		"touches": map[string]any{
			"repos": spec.Touches.Repos, "paths": spec.Touches.Paths,
			"objects": spec.Touches.Objects, "exclusive": spec.Touches.Exclusive,
		},
		"workspace": map[string]any{
			"mode": spec.Workspace.Mode, "base": spec.Workspace.Base, "branch": spec.Workspace.Branch,
		},
		"agent":    spec.Agent,
		"priority": spec.Priority,
		"verify": "验证要求（R-TYPE-3）：完成前必须过 Verify——code 规则→connectops-code-audit；" +
			"agent 规则→规则原文+diff 丢给 agent（verdict+证据）；human 规则→审批。验证失败不得 done。",
		"landing": "产物落地唯一路径=Apply/合并队列；Apply 须携 fieldManager=task:" + id +
			"（apply_log 冲突行归因键——BUG-013 Apply 关口冲突拦截）；完成后回写 Task.status（lastAction 写清结果）。",
	}
	var buf bytes.Buffer
	enc := yaml.NewEncoder(&buf)
	enc.SetIndent(2)
	if err := enc.Encode(doc); err != nil {
		return "", err
	}
	if err := enc.Close(); err != nil {
		return "", err
	}
	f, err := os.CreateTemp("", "cored-task-*.yaml")
	if err != nil {
		return "", err
	}
	if _, err := f.Write(buf.Bytes()); err != nil {
		f.Close()
		os.Remove(f.Name())
		return "", err
	}
	return f.Name(), f.Close()
}

// TaskPhase 对象原文的现行 phase（宽松；不可解析=空串）。
func TaskPhase(content []byte) string {
	_, st, ok := Parse(content)
	if !ok {
		return ""
	}
	return st.Phase
}

// TaskLastError 对象原文的现行 lastError（排队原因等值比对输入）。
func TaskLastError(content []byte) string {
	_, st, ok := Parse(content)
	if !ok {
		return ""
	}
	return st.LastError
}

func readyAction(d Decision) string {
	if d.Start {
		return "就绪（依赖/冲突/上限通过，待拉起）"
	}
	if d.Reason != "" {
		return "排队中（原因见 lastError）"
	}
	return "就绪"
}

func strPtr(s string) *string { return &s }
