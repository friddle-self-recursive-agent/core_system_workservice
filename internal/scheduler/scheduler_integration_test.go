// scheduler_integration_test.go —— Tick 端到端锚（真 sqlite 仓库 + 真 scheme）：
// 新建 Task（status 空）→ 一拍后 status.phase 落 ready 且 TaskStateChanged 已发；
// 2026-11-26 活体缺口的回归锚（空 status 曾被 Parse 跳过，任务永不调度）。
package scheduler

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// coreSchemePath 测试装载 core/scheme.yaml（CORE_SYSTEM_SCHEME 环境变量优先——与 cored
// 启动旗标同源词形；缺省=Mac 侧开发机布局。非该布局机器设 CORE_SYSTEM_SCHEME 即绿）。
func coreSchemePath() string {
	if v := strings.TrimSpace(os.Getenv("CORE_SYSTEM_SCHEME")); v != "" {
		return v
	}
	return "/Users/friddle/Project/self-recursive-agent/core_design/core/scheme.yaml"
}

func TestTickMigratesFreshTaskToReady(t *testing.T) {
	sch, err := scheme.LoadFile(coreSchemePath())
	if err != nil {
		t.Fatalf("装载 scheme: %v", err)
	}
	repo, err := repository.OpenSQLite(t.TempDir()+"/docs.db", sch, nil, "")
	if err != nil {
		t.Fatalf("打开 sqlite: %v", err)
	}
	defer repo.Close()

	ctx := context.Background()
	content := []byte(taskDocEmptyStatus)
	if _, _, err := repo.Write(ctx, &repository.Object{
		Kind: "Task", Scope: "self-recursive-core", Name: "t1", Content: content,
	}, "test", ""); err != nil {
		t.Fatalf("写入 Task: %v", err)
	}

	bus := eventbus.NewBus(&eventbus.Options{})
	got, cancel := bus.Subscribe("TaskStateChanged")
	defer cancel()

	// spawn 关断（SpawnAddr 空）——只验相位迁移与事件，不拉起会话。
	s := New(repo, bus, sch, nil, nil, Options{})
	s.Tick(ctx)

	obj, err := repo.Get(ctx, "Task", "self-recursive-core", "t1")
	if err != nil {
		t.Fatalf("读回: %v", err)
	}
	_, st, ok := Parse(obj.Content)
	if !ok {
		t.Fatalf("读回不可解析：%s", obj.Content)
	}
	if st.Phase != PhaseReady {
		t.Fatalf("一拍后相位应为 ready（spawn 关断时保持 ready 等桥），得 %q（原文：%s）", st.Phase, obj.Content)
	}
	if obj.ResourceVersion <= 1 {
		t.Fatalf("写面未落库：rv=%d 应 >1", obj.ResourceVersion)
	}
	select {
	case ev := <-got:
		if ev.Type != "TaskStateChanged" {
			t.Fatalf("事件类型=%s", ev.Type)
		}
	case <-time.After(time.Second):
		t.Fatal("应发 TaskStateChanged（状态迁移与事件同拍）")
	}
}
