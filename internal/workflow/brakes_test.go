// brakes_test.go —— 刹车首段（v0.32，workflows 规则 17）语义单测：冷却门/收紧钳定/推进拍
// 变更上限/刹车闸关断基线保护/生效值投影。时间确定性=TriggerRecord.ReceivedAt 直接构造
// （冷却锚首段=run 受理时刻——rec.ReceivedAt 即受理时刻，免时钟注入）。
package workflow

import (
	"testing"
	"time"
)

// brakeRec 构造触发留痕（involved/receivedAt 直给——冷却门键=锚+involved）。
func brakeRec(involved string, at time.Time) TriggerRecord {
	return TriggerRecord{
		Seq: 1, EventID: "e-" + involved, Type: "DocApplied", DedupKey: "k-" + involved,
		Involved: involved, Level: "warn", At: at, ReceivedAt: at,
		Route: "workflows/core-workflow", Owner: "global",
	}
}

func brakeBinding() ExecutorBinding {
	return ExecutorBinding{Executor: ExecWorkflow, Target: "core-workflow"}
}

// TestBrakeGateOff 刹车闸未挂接（缺省零值）=零刹车基线保护：同锚同 involved 连续两触发全落 run，
// braked=0（既有行为面零翻转——关断语义组合声明面先定）。
func TestBrakeGateOff(t *testing.T) {
	r := NewRunLog(nil) // BrakeOptions 零值=未挂接
	t0 := time.Now()
	r.StartRun(brakeRec("Agent/global/x", t0), brakeBinding())
	r.StartRun(brakeRec("Agent/global/x", t0.Add(time.Second)), brakeBinding())
	st := r.State()
	if st.Braked != 0 {
		t.Fatalf("未挂接 braked=%d，期望 0（零刹车基线保护）", st.Braked)
	}
	if st.Started != 2 {
		t.Fatalf("未挂接 started=%d，期望 2（触发全落 run）", st.Started)
	}
	if len(st.Runs) != 2 {
		t.Fatalf("未挂接 run 数=%d，期望 2", len(st.Runs))
	}
}

// TestCooldownGate 冷却门首段：同锚同 involved cooldown 窗口内第二触发不落 run（braked 可见），
// 窗口外照落；跨 involved 不互踩（防震荡语义粒度=同一对象）。
func TestCooldownGate(t *testing.T) {
	r := NewRunLog(&RunOptions{Brakes: BrakeOptions{Enabled: true}}) // nil resolver=强制默认 2m
	t0 := time.Now()
	r.StartRun(brakeRec("Agent/global/x", t0), brakeBinding())
	r.StartRun(brakeRec("Agent/global/x", t0.Add(30*time.Second)), brakeBinding()) // 窗口内=拦阻
	r.StartRun(brakeRec("Agent/global/y", t0.Add(40*time.Second)), brakeBinding()) // 跨 involved=不互踩
	st := r.State()
	if st.Braked != 1 {
		t.Fatalf("braked=%d，期望 1（窗口内第二触发拦阻）", st.Braked)
	}
	if st.Started != 2 {
		t.Fatalf("started=%d，期望 2（首触发+跨 involved 触发）", st.Started)
	}
	r.StartRun(brakeRec("Agent/global/x", t0.Add(2*time.Minute+time.Second)), brakeBinding()) // 窗口外=照落
	if st2 := r.State(); st2.Started != 3 {
		t.Fatalf("窗口外 started=%d，期望 3（窗口外照落）", st2.Started)
	}
}

// TestCooldownClamp 收紧判定：spec 声明收紧（5m）生效；宽松方向声明（1m）钳回强制默认 2m
// （可收紧不可关闭的机读落点）。
func TestCooldownClamp(t *testing.T) {
	tight := BrakesFunc(func(name string) (Brakes, bool) { return Brakes{Cooldown: 5 * time.Minute}, true })
	r := NewRunLog(&RunOptions{Brakes: BrakeOptions{Enabled: true, Resolver: tight}})
	t0 := time.Now()
	r.StartRun(brakeRec("Agent/global/x", t0), brakeBinding())
	r.StartRun(brakeRec("Agent/global/x", t0.Add(3*time.Minute)), brakeBinding()) // 5m 窗口内=拦阻
	if st := r.State(); st.Braked != 1 {
		t.Fatalf("收紧 5m 窗口内 braked=%d，期望 1", st.Braked)
	}
	loose := BrakesFunc(func(name string) (Brakes, bool) { return Brakes{Cooldown: time.Minute}, true })
	r2 := NewRunLog(&RunOptions{Brakes: BrakeOptions{Enabled: true, Resolver: loose}})
	r2.StartRun(brakeRec("Agent/global/x", t0), brakeBinding())
	r2.StartRun(brakeRec("Agent/global/x", t0.Add(time.Minute)), brakeBinding()) // 声明 1m 钳回 2m=窗口内拦阻
	if st := r2.State(); st.Braked != 1 {
		t.Fatalf("宽松声明钳回后 braked=%d，期望 1（不可关闭——1m 钳回默认 2m）", st.Braked)
	}
}

// TestMaxMutationsTickCap 推进拍变更上限：每拍每锚变更数受生效刹车值约束，超限 run 本拍零
// 动作（capped 可见）、下拍续推（level-triggered 不丢拍）；序列尾零动作不占额。
func TestMaxMutationsTickCap(t *testing.T) {
	tight := BrakesFunc(func(name string) (Brakes, bool) { return Brakes{MaxMutationsTick: 2}, true })
	r := NewRunLog(&RunOptions{Brakes: BrakeOptions{Enabled: true, Resolver: tight}})
	t0 := time.Now()
	walkRec := func(inv string) TriggerRecord { // 短序列 watch-only（受理首步 Analyzing→一拍至 Reporting→序列尾）
		rec := brakeRec(inv, t0)
		rec.Route = "workflows/watch-only"
		return rec
	}
	for _, inv := range []string{"Agent/global/a", "Agent/global/b", "Agent/global/c"} {
		r.StartRun(walkRec(inv), brakeBinding()) // 三 run 同锚异 involved（冷却门不互踩），受理首步至 Analyzing
	}
	if n := r.Advance(); n != 2 {
		t.Fatalf("首拍推进数=%d，期望 2（上限 2）", n)
	}
	st := r.State()
	if st.Capped != 1 {
		t.Fatalf("capped=%d，期望 1（超额 run 本拍零动作）", st.Capped)
	}
	if st.Runs[2].Phase != string(Analyzing) {
		t.Fatalf("超额 run 相位=%s，期望 Analyzing（本拍零动作）", st.Runs[2].Phase)
	}
	if n := r.Advance(); n != 1 {
		t.Fatalf("次拍推进数=%d，期望 1（下拍续推——不丢拍）", n)
	}
	if st := r.State(); st.Runs[2].Phase != string(Reporting) {
		t.Fatalf("次拍后超额 run 相位=%s，期望 Reporting（续推到位）", st.Runs[2].Phase)
	}
	if n := r.Advance(); n != 0 {
		t.Fatalf("三拍推进数=%d，期望 0（全序列尾——零动作不占额）", n)
	}
	if st := r.State(); st.Capped != 1 {
		t.Fatalf("序列尾 capped=%d，期望 1（零动作不占额不计数）", st.Capped)
	}
}

// TestMaxMutationsClampLoose 宽松方向变更上限声明（50）钳回强制默认 10：13 个同锚 run 单拍
// 至多 10 变更，超额 3 计 capped（钳定生效的机械落点）。
func TestMaxMutationsClampLoose(t *testing.T) {
	loose := BrakesFunc(func(name string) (Brakes, bool) { return Brakes{MaxMutationsTick: 50}, true })
	r := NewRunLog(&RunOptions{Brakes: BrakeOptions{Enabled: true, Resolver: loose}})
	t0 := time.Now()
	for i := 0; i < 13; i++ {
		r.StartRun(brakeRec(string(rune('a'+i))+"/global/x", t0), brakeBinding())
	}
	if n := r.Advance(); n != 10 {
		t.Fatalf("首拍推进数=%d，期望 10（50 钳回默认 10）", n)
	}
	if st := r.State(); st.Capped != 3 {
		t.Fatalf("capped=%d，期望 3（超额 3 计数）", st.Capped)
	}
}

// TestBrakesProjection 生效刹车值投影（声明→留痕面）：解析命中=spec 来源携声明收紧值；
// 未命中=强制默认来源；投影不依赖刹车闸挂接态（声明真值恒投影）。
func TestBrakesProjection(t *testing.T) {
	tight := BrakesFunc(func(name string) (Brakes, bool) {
		if name == "core-workflow" {
			return Brakes{Cooldown: 5 * time.Minute, MaxMutationsTick: 2}, true
		}
		return Brakes{}, false
	})
	r := NewRunLog(&RunOptions{Brakes: BrakeOptions{Enabled: true, Resolver: tight}}) // 挂接态投影
	t0 := time.Now()
	r.StartRun(brakeRec("Agent/global/x", t0), brakeBinding()) // 锚=core-workflow（命中声明）
	other := brakeRec("Agent/global/y", t0)
	other.Route = "workflows/watch-only"
	r.StartRun(other, ExecutorBinding{Executor: ExecWorkflow, Target: "watch-only"}) // 未命中锚
	off := NewRunLog(nil)                                                            // 未挂接态投影（声明真值恒投影）
	off.StartRun(brakeRec("Agent/global/z", t0), brakeBinding())
	st := r.State()
	if st.Runs[0].Brakes.Cooldown != 5*time.Minute || st.Runs[0].Brakes.MaxMutationsTick != 2 || st.Runs[0].Brakes.Source != BrakeSourceSpec {
		t.Fatalf("命中锚投影=%+v，期望 5m/2/spec", st.Runs[0].Brakes)
	}
	if st.Runs[1].Brakes.Source != BrakeSourceDefault || st.Runs[1].Brakes.MaxMutationsTick != DefaultBrakes.MaxMutationsTick {
		t.Fatalf("未命中锚投影=%+v，期望 default/10", st.Runs[1].Brakes)
	}
	so := off.State()
	if so.Runs[0].Brakes.Source != BrakeSourceDefault || so.Runs[0].Brakes.Cooldown != DefaultBrakes.Cooldown {
		t.Fatalf("未挂接投影=%+v，期望 default/2m（声明真值恒投影）", so.Runs[0].Brakes)
	}
}

// TestDocsSpecsResolveBrakes 文档面 brakes 解析：声明收紧值透传原样；未声明/损坏=未命中；
// cooldown 词形非法=未命中（写路径校验闸门是形态权威——解析面防御性兜底）。
func TestDocsSpecsResolveBrakes(t *testing.T) {
	docs := map[string][]byte{
		"declared": []byte("spec:\n  brakes:\n    cooldown: 5m\n    maxMutationsPerTick: 2\n"),
		"partial":  []byte("spec:\n  brakes:\n    maxMutationsPerTick: 3\n"),
		"plain":    []byte("spec: {}\n"),
		"broken":   []byte("spec:\n  brakes:\n    cooldown: notaduration\n"),
		"badyaml":  []byte("spec: [oops\n"),
	}
	d := DocsSpecs{Lookup: func(name string) ([]byte, bool) { b, ok := docs[name]; return b, ok }}
	if b, ok := d.ResolveBrakes("declared"); !ok || b.Cooldown != 5*time.Minute || b.MaxMutationsTick != 2 {
		t.Fatalf("declared=%+v ok=%v，期望 5m/2 true", b, ok)
	}
	if b, ok := d.ResolveBrakes("partial"); !ok || b.Cooldown != 0 || b.MaxMutationsTick != 3 {
		t.Fatalf("partial=%+v ok=%v，期望 0/3 true", b, ok)
	}
	if _, ok := d.ResolveBrakes("plain"); ok {
		t.Fatal("未声明 brakes 应未命中（强制默认兜底）")
	}
	if _, ok := d.ResolveBrakes("broken"); ok {
		t.Fatal("cooldown 词形非法应未命中（防御性兜底）")
	}
	if _, ok := d.ResolveBrakes("badyaml"); ok {
		t.Fatal("损坏文档应未命中（防御性同 ResolveSpec 口径）")
	}
	if _, ok := d.ResolveBrakes("ghost"); ok {
		t.Fatal("未命中锚应未命中")
	}
	if _, ok := d.ResolveBrakes(""); ok {
		t.Fatal("空锚应未命中")
	}
}

// TestCooldownGateEventAdvance 事件拍与定时拍共用同一推进入口——上限对事件拍同闸（同一
// Advance 落点，零旁路）。
func TestCooldownGateEventAdvance(t *testing.T) {
	tight := BrakesFunc(func(name string) (Brakes, bool) { return Brakes{MaxMutationsTick: 1}, true })
	r := NewRunLog(&RunOptions{Brakes: BrakeOptions{Enabled: true, Resolver: tight}})
	var sink EventAdvanceSink = r
	t0 := time.Now()
	walkRec := func(inv string) TriggerRecord { // 短序列 watch-only（一拍一步确定性）
		rec := brakeRec(inv, t0)
		rec.Route = "workflows/watch-only"
		return rec
	}
	r.StartRun(walkRec("Agent/global/a"), brakeBinding())
	r.StartRun(walkRec("Agent/global/b"), brakeBinding())
	if n := sink.Advance(); n != 1 {
		t.Fatalf("事件拍经能力接口推进数=%d，期望 1（上限 1 同闸）", n)
	}
	if st := r.State(); st.Capped != 1 {
		t.Fatalf("capped=%d，期望 1", st.Capped)
	}
}
