// event_advance_test.go —— 事件驱动推进源首段验收（v0.29——巡检拍与推进拍合一，
// workflow.md 规则 15 v0.29 段；v0.30 connector watch 桥接源入拍面；v0.31 metrics 阈值
// 桥接源入拍面）：引擎装配内部源事件（source ∈ beatSources：core 巡检+connector watch 桥）落 run 的同拍全量推进拍
// （与定时拍同一查表落点——判定器零改动）/ 事件拍三重闸（落了 run ∧ source ∈ 拍面集合
// ∧ 装配层挂接）/ 人注与权限自产事件不拍（source 维度分立——注入通道恒 human）/ 总闸
// 关断事件拍缺席（不旁路总闸——关断语义组合声明面先定）/ 多拍确定性（新落 run 参与本拍：
// 受理首步+事件拍两拍步）。
package workflow

import (
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

// pubSrc 指定 source 发布事件（事件拍 source 维度判别的测试注入面；dedupKey 唯一——
// 总线窗口去重不吞测试拍）。
func pubSrc(t *testing.T, b *eventbus.Bus, typ, source, involved string) {
	t.Helper()
	e := eventbus.Event{Type: typ, Source: source, Involved: involved,
		DedupKey: typ + ":" + involved + ":" + time.Now().String(), Level: "info"}
	if err := b.Publish(&e); err != nil {
		t.Fatalf("发布 %s: %v", typ, err)
	}
}

// tickHub 事件拍测试装配：hub（runs+workflow 执行面登记+可选事件拍挂接）+权威表同构
// dispatch 行（ReconcileTick→对账锚）。bindTick=false=总闸关断（装配层不挂接——对照面）。
func tickHub(t *testing.T, bindTick bool) (*eventbus.Bus, *TriggerHub, *RunLog) {
	b := eventbus.NewBus(nil)
	h := NewTriggerHub(nil)
	runs := NewRunLog(nil)
	h.BindRuns(runs)
	h.WireExecutor(ExecWorkflow)
	if bindTick {
		h.BindEventAdvance(runs) // 推进总闸开启才挂接（装配层口径——main.go 同款）
	}
	cancel := h.AttachDispatch(b, []scheme.DispatchRule{
		{On: "ReconcileTick", For: "any", Dispatch: "workflows/core-workflow"},
		{On: "DocApplied", For: "owner:*", Dispatch: "owner-workflow"},
		{On: "ExternalChanged", For: "owner:*", Dispatch: "owner-workflow"},
		{On: "MetricThreshold", For: "owner:*", Dispatch: "owner-workflow"},
	})
	t.Cleanup(cancel) // 退订作用域=整个测试（helper 内 defer 会在返回时提前退订）
	return b, h, runs
}

// TestEventAdvanceTickFires 事件拍生效面：ReconcileTick（source=core）落 run 的同拍全量
// 推进——新落 run 受理首步（Idle→Analyzing）后同拍再进一步（Analyzing→Planning），
// advances 计入（推进留痕不问驱动源——同一条查询键）。
func TestEventAdvanceTickFires(t *testing.T) {
	b, h, runs := tickHub(t, true)
	pubSrc(t, b, "ReconcileTick", SourceEngineInternal, "Workflow/builtin/core-workflow")
	waitReceived(t, h, 1)
	rs := runs.State()
	if len(rs.Runs) != 1 {
		t.Fatalf("巡检事件应落 1 条 run: %+v", rs)
	}
	r := rs.Runs[0]
	if r.Phase != string(Planning) || r.Status != RunRunning {
		t.Fatalf("事件拍同拍应再进一步（受理首步 Analyzing+事件拍 Planning）: {phase=%s, status=%s}", r.Phase, r.Status)
	}
	if rs.Advances != 1 {
		t.Fatalf("事件拍推进应计入 advances（留痕不问驱动源），得 %d", rs.Advances)
	}
}

// TestEventAdvanceSecondTickBatch 多拍确定性：第二拍=现势 run 全量各进一步（新落 run
// 受理首步+事件拍两拍步；上一拍 run 沿表前进落闸锚即挂起）——逐段承接不跳步。
func TestEventAdvanceSecondTickBatch(t *testing.T) {
	b, h, runs := tickHub(t, true)
	pubSrc(t, b, "ReconcileTick", SourceEngineInternal, "Workflow/builtin/core-workflow")
	waitReceived(t, h, 1)
	time.Sleep(5 * time.Millisecond) // 拍间隔离（总线窗口去重不吞第二拍——dedupKey 逐拍互异）
	pubSrc(t, b, "ReconcileTick", SourceEngineInternal, "Workflow/builtin/core-workflow")
	waitReceived(t, h, 2)
	rs := runs.State()
	if len(rs.Runs) != 2 {
		t.Fatalf("两拍应落 2 条 run: %+v", rs)
	}
	// 末条=本拍新落 run（受理首步+事件拍两拍步：Idle→Analyzing→Planning running）。
	last := rs.Runs[1]
	if last.Phase != string(Planning) || last.Status != RunRunning {
		t.Fatalf("新落 run 应 {Planning, running}: {phase=%s, status=%s}", last.Phase, last.Status)
	}
	// 倒数第二条=上一拍 run（上一拍 Planning→本拍 AwaitingApproval 落闸锚挂起——落点语义同定时拍）。
	prev := rs.Runs[0]
	if prev.Phase != string(AwaitingApproval) || prev.Status != RunSuspended {
		t.Fatalf("上一拍 run 应落闸锚挂起 {AwaitingApproval, suspended}: {phase=%s, status=%s}", prev.Phase, prev.Status)
	}
	if rs.Advances != 3 { // 第一拍 1 步 + 第二拍 2 步（新落 run 与上一拍 run 各一）
		t.Fatalf("两拍推进累计应=3，得 %d", rs.Advances)
	}
}

// TestEventAdvanceHumanSourceNoTick source 维度分立：人注事件（source=human）落 run 只走
// 受理首步——不触发事件拍（引擎内部源审计事实是拍面的判据，非类型非路由）。
func TestEventAdvanceHumanSourceNoTick(t *testing.T) {
	b, h, runs := tickHub(t, true)
	pubSrc(t, b, "DocApplied", "human", "Agent/global/probe")
	waitReceived(t, h, 1)
	rs := runs.State()
	if len(rs.Runs) != 1 || rs.Runs[0].Phase != string(Analyzing) || rs.Runs[0].Status != RunRunning {
		t.Fatalf("人注事件落 run 应止于受理首步 {Analyzing, running}: %+v", rs)
	}
	if rs.Advances != 0 {
		t.Fatalf("人注事件不得触发事件拍（advances 应=0），得 %d", rs.Advances)
	}
}

// TestEventAdvanceConnectorWatchTick connector watch 桥接源入拍面（v0.30 驱动源集合
// 扩展，event.md 规则 12）：source=connector（服务侧判定口径）落 run 的同拍全量推进——
// 受理首步+事件拍两拍步（与 core 巡检同拍面；零新开关——拍面随源扩展，总闸口径不变）。
func TestEventAdvanceConnectorWatchTick(t *testing.T) {
	b, h, runs := tickHub(t, true)
	pubSrc(t, b, "ExternalChanged", SourceConnectorWatch, "WorkService/onedev")
	waitReceived(t, h, 1)
	rs := runs.State()
	if len(rs.Runs) != 1 {
		t.Fatalf("watch 事件应落 1 条 run: %+v", rs)
	}
	r := rs.Runs[0]
	if r.Phase != string(Planning) || r.Status != RunRunning {
		t.Fatalf("watch 事件拍同拍应两拍步（受理首步 Analyzing+事件拍 Planning）: {phase=%s, status=%s}", r.Phase, r.Status)
	}
	if rs.Advances != 1 {
		t.Fatalf("watch 事件拍推进应计入 advances，得 %d", rs.Advances)
	}
}

// TestEventAdvanceMetricsTick metrics 阈值桥接源入拍面（v0.31 驱动源集合扩展，
// event.md 规则 13）：source=metrics（服务侧判定口径）落 run 的同拍全量推进——受理首步+
// 事件拍两拍步（与 core 巡检/connector watch 同拍面；零新开关——拍面随源扩展，总闸口径不变）。
func TestEventAdvanceMetricsTick(t *testing.T) {
	b, h, runs := tickHub(t, true)
	pubSrc(t, b, "MetricThreshold", SourceMetricsThreshold, "WorkService/metrics")
	waitReceived(t, h, 1)
	rs := runs.State()
	if len(rs.Runs) != 1 {
		t.Fatalf("metrics 阈值事件应落 1 条 run: %+v", rs)
	}
	r := rs.Runs[0]
	if r.Phase != string(Planning) || r.Status != RunRunning {
		t.Fatalf("metrics 事件拍同拍应两拍步（受理首步 Analyzing+事件拍 Planning）: {phase=%s, status=%s}", r.Phase, r.Status)
	}
	if rs.Advances != 1 {
		t.Fatalf("metrics 事件拍推进应计入 advances，得 %d", rs.Advances)
	}
}

// TestEventAdvancePermissionSourceNoTick 权限自产事件同口径不拍（source=permission——
// 注入通道恒 human 与权限自产均不在拍面驱动源集合，source 维度分立）。
func TestEventAdvancePermissionSourceNoTick(t *testing.T) {
	b, h, runs := tickHub(t, true)
	pubSrc(t, b, "DocApplied", "permission", "Agent/global/probe")
	waitReceived(t, h, 1)
	if rs := runs.State(); rs.Advances != 0 {
		t.Fatalf("权限自产事件不得触发事件拍，得 %d", rs.Advances)
	}
}

// TestEventAdvanceGateOffNoTick 总闸关断事件拍缺席：装配层不挂接（--advance-interval=0）
// 时引擎内部源事件落 run 只走受理首步——事件驱动源不旁路总闸（关断语义组合声明面先定；
// 巡检-only 部署行为面零翻转的机械判据）。
func TestEventAdvanceGateOffNoTick(t *testing.T) {
	b, h, runs := tickHub(t, false)
	pubSrc(t, b, "ReconcileTick", SourceEngineInternal, "Workflow/builtin/core-workflow")
	waitReceived(t, h, 1)
	rs := runs.State()
	if len(rs.Runs) != 1 || rs.Runs[0].Phase != string(Analyzing) || rs.Runs[0].Status != RunRunning {
		t.Fatalf("总闸关断落 run 应止于受理首步 {Analyzing, running}: %+v", rs)
	}
	if rs.Advances != 0 {
		t.Fatalf("总闸关断事件拍应缺席（advances 应=0），得 %d", rs.Advances)
	}
}
