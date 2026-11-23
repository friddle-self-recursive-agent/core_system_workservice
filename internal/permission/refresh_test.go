package permission

// 热装载测试（v0.10，refresh.go）：替换语义 / 失败保旧快照 + 重试自愈 / 零值 fail-closed /
// 装载×判定并发安全（-race）。变更信号以内存 channel 注入（信号源无关性——repository.Watch
// 只是生产环境的一个信号源适配）。

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

const roleWriterHot = `kind: Role
metadata: {name: hot-writer}
spec:
  permissions:
  - {scope: "global/*", actions: [write]}
`

const roleOtherHot = `kind: Role
metadata: {name: hot-other}
spec:
  permissions:
  - {scope: "ops/probe", actions: [call]}
`

// eventually 条件轮询等待（热装载是异步的：信号→装载→上架之间存在调度窗口）。
func eventually(t *testing.T, d time.Duration, cond func() bool) {
	t.Helper()
	ddl := time.Now().Add(d)
	for time.Now().Before(ddl) {
		if cond() {
			return
		}
		time.Sleep(2 * time.Millisecond)
	}
	t.Fatal("条件超时未满足")
}

// runRefresher 启动热装载循环并注册退场清理。
func runRefresher(t *testing.T, r *Refresher, changes chan struct{}) context.CancelFunc {
	t.Helper()
	ctx, cancel := context.WithCancel(context.Background())
	done := make(chan struct{})
	go func() { _ = r.Run(ctx, changes); close(done) }()
	t.Cleanup(func() {
		cancel()
		select {
		case <-done:
		case <-time.After(2 * time.Second):
			t.Error("Refresher.Run 未按 ctx 退场")
		}
	})
	return cancel
}

// TestRefreshReplacesSnapshot 热装载=整体替换：信号一次，快照换代（Gen+1），
// 旧授权随替换消失、新授权生效——装载即权威快照的可重入性。
func TestRefreshReplacesSnapshot(t *testing.T) {
	s := &System{}
	if err := s.LoadObjects([]ObjectDoc{{Kind: "Role", Name: "hot-writer", Scope: "global", Content: []byte(roleWriterHot)}}); err != nil {
		t.Fatal(err)
	}
	var results []RefreshResult
	var mu sync.Mutex
	r := &Refresher{
		Sys: s,
		Load: func() ([]ObjectDoc, error) {
			return []ObjectDoc{{Kind: "Role", Name: "hot-other", Scope: "global", Content: []byte(roleOtherHot)}}, nil
		},
		Reloaded: func(res RefreshResult) {
			mu.Lock()
			results = append(results, res)
			mu.Unlock()
		},
	}
	changes := make(chan struct{}, 4)
	runRefresher(t, r, changes)
	changes <- struct{}{}
	eventually(t, time.Second, func() bool {
		snap := s.Snapshot()
		return snap != nil && snap.Gen == 2 && len(snap.Roles) == 1
	})
	if err := s.Check(Subject{User: "u", Roles: []string{"role/hot-writer"}}, "global", "write", ByRole); err == nil {
		t.Fatal("被替换掉的旧 role 应随换代失去授权")
	}
	if err := s.Check(Subject{User: "u", Roles: []string{"role/hot-other"}}, "ops/probe", "call", ByRole); err != nil {
		t.Fatalf("新装载的 role 应生效: %v", err)
	}
	mu.Lock()
	defer mu.Unlock()
	if len(results) != 1 || !results[0].Ok || results[0].Gen != 2 || results[0].Roles != 1 {
		t.Fatalf("Reloaded 终态不符: %+v", results)
	}
}

// TestRefreshFailureKeepsSnapshot 热装载失败（坏对象）=保旧快照 + 告警回调：
// 授权面继续按上一代权威快照判定（不半装载、不停摆）。
func TestRefreshFailureKeepsSnapshot(t *testing.T) {
	s := &System{}
	if err := s.LoadObjects([]ObjectDoc{{Kind: "Role", Name: "hot-writer", Scope: "global", Content: []byte(roleWriterHot)}}); err != nil {
		t.Fatal(err)
	}
	bad := []ObjectDoc{{Kind: "Role", Name: "broken", Scope: "global", Content: []byte("kind: Role\nmetadata: {name: broken}\nspec:\n  permissions:\n  - {scope: global}\n")}} // 缺 actions
	var failures atomic.Int64
	r := &Refresher{
		Sys:  s,
		Load: func() ([]ObjectDoc, error) { return bad, nil },
		Reloaded: func(res RefreshResult) {
			if !res.Ok {
				failures.Add(1)
			}
		},
	}
	changes := make(chan struct{}, 4)
	runRefresher(t, r, changes)
	changes <- struct{}{}
	eventually(t, time.Second, func() bool { return failures.Load() > 0 })
	snap := s.Snapshot()
	if snap == nil || snap.Gen != 1 || len(snap.Roles) != 1 {
		t.Fatalf("失败装载不得换代：gen=%v roles=%v", snap, snap)
	}
	if err := s.Check(Subject{User: "u", Roles: []string{"role/hot-writer"}}, "global", "write", ByRole); err != nil {
		t.Fatalf("保旧快照：旧授权应继续生效: %v", err)
	}
}

// TestRefreshRetryHealsWithoutSignal 失败后定时重试自愈：投影转好，无新变更信号也痊愈
// （坏对象滞留场景的自愈路径——重试不依赖下一次文档面变更）。
func TestRefreshRetryHealsWithoutSignal(t *testing.T) {
	s := &System{}
	if err := s.LoadObjects([]ObjectDoc{{Kind: "Role", Name: "hot-writer", Scope: "global", Content: []byte(roleWriterHot)}}); err != nil {
		t.Fatal(err)
	}
	var bad atomic.Bool
	bad.Store(true)
	r := &Refresher{
		Sys: s,
		Load: func() ([]ObjectDoc, error) {
			if bad.Load() {
				return []ObjectDoc{{Kind: "Role", Name: "broken", Content: []byte("spec: [")}}, nil // yaml 不可解析
			}
			return []ObjectDoc{{Kind: "Role", Name: "hot-other", Scope: "global", Content: []byte(roleOtherHot)}}, nil
		},
		Retry: 10 * time.Millisecond,
	}
	changes := make(chan struct{}, 4)
	runRefresher(t, r, changes)
	changes <- struct{}{}
	eventually(t, time.Second, func() bool {
		snap := s.Snapshot()
		return snap != nil && len(snap.Roles) == 1 && snap.Roles["hot-writer"] != nil // 失败批次=保旧
	})
	bad.Store(false) // 投影转好；不发新信号
	eventually(t, 2*time.Second, func() bool {
		snap := s.Snapshot()
		return snap != nil && snap.Gen == 2 && snap.Roles["hot-other"] != nil
	})
}

// TestZeroValueSystemFailClosed 零值系统=空授权面：非 admin 主体全拒（nil 快照 fail-closed），
// 内置 admin 字面量仍放行（本地信任 fallback 最低保障，bootstrap/工作流自写不因此失能）。
func TestZeroValueSystemFailClosed(t *testing.T) {
	s := &System{}
	if s.Snapshot() != nil {
		t.Fatal("零值系统应为 nil 快照")
	}
	if err := s.Check(Subject{User: "u", Roles: []string{"role/anything"}}, "global", "write", ByRole); err == nil {
		t.Fatal("nil 快照应拒绝非 admin 主体")
	}
	if err := s.Check(Subject{User: "cored:local", Roles: []string{DefaultAdmin}}, "任意域", "write", ByRole); err != nil {
		t.Fatalf("内置 admin 字面量应放行: %v", err)
	}
}

// TestConcurrentLoadAndCheck 装载×判定并发安全（-race 守护）：热装载整体替换与高频判定并行，
// 判定永远落在某一代完整快照上（原子指针语义），无数据竞争、无半装载态可见。
func TestConcurrentLoadAndCheck(t *testing.T) {
	s := &System{}
	if err := s.LoadObjects([]ObjectDoc{{Kind: "Role", Name: "hot-writer", Scope: "global", Content: []byte(roleWriterHot)}}); err != nil {
		t.Fatal(err)
	}
	docs := func(name string, content string) []ObjectDoc {
		return []ObjectDoc{{Kind: "Role", Name: name, Scope: "global", Content: []byte(content)}}
	}
	stop := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)
	go func() { // 装载方：两代快照交替替换
		defer wg.Done()
		next := roleOtherHot
		for i := 0; i < 200; i++ {
			_ = s.LoadObjects(docs("hot-other", next))
			if next == roleOtherHot {
				next = roleWriterHot
			} else {
				next = roleOtherHot
			}
		}
		close(stop)
	}()
	for { // 判定方：与替换并行高频判定（结论两可——name 随代次在换——只验证不 panic/不撕裂）
		select {
		case <-stop:
			wg.Wait()
			if snap := s.Snapshot(); snap == nil || snap.Gen == 0 || len(snap.Roles) != 1 {
				t.Fatalf("终态快照应完整: %v", snap)
			}
			return
		default:
			_ = s.Check(Subject{User: "u", Roles: []string{"role/hot-writer", "role/hot-other"}}, "global", "write", ByRole)
		}
	}
}
