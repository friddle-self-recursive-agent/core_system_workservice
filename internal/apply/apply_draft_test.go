// apply_draft_test.go —— 稿→面落库验收（concepts/apply.md §1 编辑与发布分离 + §6 rule 6）：
// 草稿走同一三步闸门、apply 后清空、冲突保留草稿、noop 清稿不发事件、dry-run 零副作用不清稿、
// 无主/键不全在草稿消费前拒绝、Drafts 未配置 → ErrNoStore。
package apply

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/draft"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/repository"
)

// draftWorld 带草稿缓存的测试世界（其余依赖链同 newWorld）。
func draftWorld(t *testing.T) (*world, *draft.Store) {
	t.Helper()
	w := newWorld(t)
	ds := draft.NewStore(nil)
	w.engine.Drafts = ds
	return w, ds
}

// stage 落稿助手。
func stage(t *testing.T, ds *draft.Store, name string, baseRV int64, extra string) {
	t.Helper()
	err := ds.Put(&draft.Draft{
		Kind: "Agent", Scope: "global", Name: name,
		Content:      agentYAML(name, extra),
		FieldManager: "agent/dev-agent", ReasonRef: "drift/1", BaseRV: baseRV,
	})
	if err != nil {
		t.Fatalf("落稿 %s: %v", name, err)
	}
}

// TestApplyDraftCreateThenClear 草稿 apply = 落库即清稿（apply.md rule 6）：
// BaseRV=0 → create；DocApplied 必发；清稿后 fromDraft 再 apply → ErrNoDraft。
func TestApplyDraftCreateThenClear(t *testing.T) {
	w, ds := draftWorld(t)
	ctx := context.Background()
	events, cancel := w.bus.Subscribe("DocApplied")
	defer cancel()

	stage(t, ds, "core-headless", 0, "")
	res, err := w.engine.ApplyDraft(ctx, "Agent", "global", "core-headless",
		Options{FieldManager: "agent:confirm"})
	if err != nil || res.Action != "create" || res.RV != 1 || res.Event == nil {
		t.Fatalf("草稿创建不符: %+v err=%v", res, err)
	}
	select {
	case evt := <-events:
		if evt.Type != "DocApplied" || evt.Involved != "Agent/global/core-headless" {
			t.Fatalf("DocApplied 载荷不符: %+v", evt)
		}
	case <-time.After(time.Second):
		t.Fatal("草稿 apply 未发 DocApplied")
	}
	// 清稿（rule 6）：缓存空；再走 fromDraft → ErrNoDraft（区别于 noop）。
	if err := ds.Delete("Agent", "global", "core-headless"); !errors.Is(err, draft.ErrNoDraft) {
		t.Fatalf("apply 后应已清稿，Delete 得 %v", err)
	}
	if _, err := w.engine.ApplyDraft(ctx, "Agent", "global", "core-headless",
		Options{FieldManager: "agent:confirm"}); !errors.Is(err, draft.ErrNoDraft) {
		t.Fatalf("空草稿 apply 期望 ErrNoDraft，得 %v", err)
	}
	// 文档面真实落库。
	cur, err := w.repo.Get(ctx, "Agent", "global", "core-headless")
	if err != nil || cur.ResourceVersion != 1 {
		t.Fatalf("文档面未落库: %v rv=%v", err, cur)
	}
}

// TestApplyDraftNoopClearsNoEvent 稿与面一致 → noop：零提交零事件，但稿被确认消费（清稿）。
func TestApplyDraftNoopClearsNoEvent(t *testing.T) {
	w, ds := draftWorld(t)
	ctx := context.Background()
	fm := Options{FieldManager: "agent:bootstrap"}

	if _, err := w.engine.Apply(ctx, &repository.Object{Kind: "Agent", Name: "a", Scope: "global",
		Content: agentYAML("a", "")}, fm); err != nil {
		t.Fatal(err)
	}
	stage(t, ds, "a", 1, "") // 与现面同内容、同 BaseRV
	res, err := w.engine.ApplyDraft(ctx, "Agent", "global", "a",
		Options{FieldManager: "agent:confirm"})
	if err != nil || res.Action != "noop" || res.Event != nil || res.RV != 1 {
		t.Fatalf("同稿应 noop 无事件: %+v err=%v", res, err)
	}
	if err := ds.Delete("Agent", "global", "a"); !errors.Is(err, draft.ErrNoDraft) {
		t.Fatalf("noop 后也应清稿，Delete 得 %v", err)
	}
	if c := w.commits(t); c != 1 {
		t.Fatalf("noop 不得产生提交: 提交=%d", c)
	}
	if n, _ := w.ops.CountEvent("DocApplied"); n != 1 {
		t.Fatalf("noop 不得发事件: event_log=%d", n)
	}
}

// TestApplyDraftConflictKeepsDraft 乐观锁语义（apply.md §3「后写者重读」）：
// 落稿后文档面前移 → Conflict 且草稿保留；按现版本重整草稿后再 apply → 成功 rv+1。
func TestApplyDraftConflictKeepsDraft(t *testing.T) {
	w, ds := draftWorld(t)
	ctx := context.Background()
	fm := Options{FieldManager: "agent:bootstrap"}

	if _, err := w.engine.Apply(ctx, &repository.Object{Kind: "Agent", Name: "a", Scope: "global",
		Content: agentYAML("a", "")}, fm); err != nil {
		t.Fatal(err)
	}
	stage(t, ds, "a", 1, ",\n    gen: 2") // 基于 rv=1 的稿
	if _, err := w.engine.Apply(ctx, &repository.Object{Kind: "Agent", Name: "a", Scope: "global",
		ResourceVersion: 1, Content: agentYAML("a", ",\n    other: writer")}, fm); err != nil {
		t.Fatal(err)
	} // 文档面前移至 rv=2

	_, err := w.engine.ApplyDraft(ctx, "Agent", "global", "a", Options{FieldManager: "agent:confirm"})
	var ce *repository.ConflictError
	if !errors.As(err, &ce) || ce.Current != 2 || ce.Expected != 1 {
		t.Fatalf("陈旧稿应 Conflict(expected=1,current=2)，得 %v", err)
	}
	// 草稿保留（不消费）：可重整后再落。
	got, gerr := ds.Get("Agent", "global", "a")
	if gerr != nil {
		t.Fatalf("冲突后草稿应保留: %v", gerr)
	}
	got.BaseRV = 2 // 后写者重读后重整（LLM 语义 diff 的人裁落点，v1 直接换基线）
	if err := ds.Put(got); err != nil {
		t.Fatalf("重整落稿: %v", err)
	}
	res, err := w.engine.ApplyDraft(ctx, "Agent", "global", "a", Options{FieldManager: "agent:confirm"})
	if err != nil || res.Action != "update" || res.RV != 3 || res.Event == nil {
		t.Fatalf("重整后应 update rv=3: %+v err=%v", res, err)
	}
	if err := ds.Delete("Agent", "global", "a"); !errors.Is(err, draft.ErrNoDraft) {
		t.Fatalf("成功后应清稿，Delete 得 %v", err)
	}
}

// TestApplyDraftGatesKeepDraft 草稿不旁路闸门：野 kind → ErrNotRegistered（denied 行），
// 无主/键不全在消费前拒绝；所有拒绝路径草稿保留、零提交。
func TestApplyDraftGatesKeepDraft(t *testing.T) {
	w, ds := draftWorld(t)
	ctx := context.Background()

	wild := &draft.Draft{Kind: "Unicorn", Scope: "global", Name: "rainbow",
		Content: []byte("kind: Unicorn\n"), FieldManager: "agent:dev", BaseRV: 0}
	if err := ds.Put(wild); err != nil {
		t.Fatalf("落稿（落稿面不做 scheme 校验——闸门在写路径）: %v", err)
	}
	_, err := w.engine.ApplyDraft(ctx, "Unicorn", "global", "rainbow", Options{FieldManager: "agent:confirm"})
	if !errors.Is(err, ErrNotRegistered) {
		t.Fatalf("野 kind 稿 apply 应被拒，得 %v", err)
	}
	if _, gerr := ds.Get("Unicorn", "global", "rainbow"); gerr != nil {
		t.Fatalf("拒绝路径草稿应保留: %v", gerr)
	}
	if n, _ := w.ops.CountApply("denied"); n != 1 {
		t.Fatalf("denied 审计行 = %d, 要 1", n)
	}

	// 无主确认拒绝（写路径 rule 2 同源）；键不全拒绝。
	stage(t, ds, "a", 0, "")
	if _, err := w.engine.ApplyDraft(ctx, "Agent", "global", "a", Options{}); !errors.Is(err, ErrNoFieldManager) {
		t.Fatalf("无主确认应拒，得 %v", err)
	}
	if _, err := w.engine.ApplyDraft(ctx, "Agent", "", "a", Options{FieldManager: "fm"}); !errors.Is(err, ErrObject) {
		t.Fatalf("键不全应拒（ErrObject），得 %v", err)
	}
	if _, gerr := ds.Get("Agent", "global", "a"); gerr != nil {
		t.Fatalf("拒绝路径草稿应保留: %v", gerr)
	}
	if w.commits(t) != 0 {
		t.Fatal("被拒草稿 apply 不得产生提交")
	}
}

// TestApplyDraftDryRunZeroSideEffect 草稿 dry-run：准入结论一致、零副作用、不清稿。
func TestApplyDraftDryRunZeroSideEffect(t *testing.T) {
	w, ds := draftWorld(t)
	ctx := context.Background()

	stage(t, ds, "a", 0, ",\n    gen: 2")
	res, err := w.engine.ApplyDraft(ctx, "Agent", "global", "a",
		Options{FieldManager: "agent:confirm", DryRun: true})
	if err != nil || res.Action != "dry-run" || res.RV != 0 || res.Event != nil {
		t.Fatalf("将创建的草稿 dry-run 不符: %+v err=%v", res, err)
	}
	if _, gerr := ds.Get("Agent", "global", "a"); gerr != nil {
		t.Fatalf("dry-run 不得清稿: %v", gerr)
	}
	if w.commits(t) != 0 {
		t.Fatal("dry-run 不得产生提交")
	}
	if n, _ := w.ops.CountApply("ok"); n != 0 {
		t.Fatalf("dry-run 不得落 ok 行: %d", n)
	}

	// 真执行后：dry-run 返回现版本（准入结论与真执行一致）。
	if _, err := w.engine.ApplyDraft(ctx, "Agent", "global", "a", Options{FieldManager: "agent:confirm"}); err != nil {
		t.Fatal(err)
	}
	stage(t, ds, "a", 1, ",\n    gen: 3")
	res2, err := w.engine.ApplyDraft(ctx, "Agent", "global", "a",
		Options{FieldManager: "agent:confirm", DryRun: true})
	if err != nil || res2.Action != "dry-run" || res2.RV != 1 {
		t.Fatalf("现面 dry-run 应返回现版本 rv=1: %+v err=%v", res2, err)
	}
}

// TestApplyDraftStoreRequired Drafts 未配置 → ErrNoStore（只读部署不受影响）。
func TestApplyDraftStoreRequired(t *testing.T) {
	w := newWorld(t)
	_, err := w.engine.ApplyDraft(context.Background(), "Agent", "global", "a", Options{FieldManager: "fm"})
	if !errors.Is(err, draft.ErrNoStore) {
		t.Fatalf("无草稿缓存应 ErrNoStore，得 %v", err)
	}
}
