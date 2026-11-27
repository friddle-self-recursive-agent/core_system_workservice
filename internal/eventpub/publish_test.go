// 发布面单测（v0.20 注入面首段）：类型注册闸门/来源强制/level 缺省投影与词汇校验/
// 发布权限域分立/同主干投递与窗口聚合。设计依据：concepts/event.md 规则 8。
package eventpub

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/eventbus"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/permission"
	"github.com/friddle-self-recursive-agent/core_system_workservice/internal/scheme"
)

const pubFixture = `
apiVersion: core/v1
kind: Scheme
metadata: {name: default, scope: builtin}
spec:
  version: core/v1
  kinds:
    - kind: Scheme
      builtin: true
      storage: core/scheme.yaml
      spec: {version: string}
  eventTypes:
    - {type: DocApplied, source: core, defaultLevel: info, dedup: "apply 落库必发；engine 自写豁免", route: owner-workflow}
    - {type: Escalate, source: engine, defaultLevel: error, dedup: "<workflow>:<reason>", route: root+notification}
    - {type: HeartbeatLost, source: core, defaultLevel: error, dedup: "<agent>", route: parent-agent}
`

func pubTestBus(t *testing.T) (*eventbus.Bus, *scheme.Scheme, <-chan eventbus.Event, func()) {
	t.Helper()
	sch, err := scheme.Load([]byte(pubFixture))
	if err != nil {
		t.Fatalf("scheme 载入失败: %v", err)
	}
	bus := eventbus.NewBus(&eventbus.Options{DedupWindow: 300 * time.Millisecond})
	ch, cancel := bus.Subscribe("")
	return bus, sch, ch, cancel
}

func adminSub() *permission.Subject {
	return &permission.Subject{User: "pub-admin", Roles: []string{permission.DefaultAdmin}}
}

func roleDoc(name, scope, actions string) permission.ObjectDoc {
	return permission.ObjectDoc{
		Kind: "Role", Name: name, Scope: "global",
		Content: []byte("apiVersion: core/v1\nkind: Role\nmetadata: {name: " + name + ", scope: global}\nspec:\n  permissions:\n  - {scope: " + scope + ", actions: [" + actions + "]}\n"),
	}
}

func TestPublishHumanSourceForced(t *testing.T) {
	bus, sch, ch, cancel := pubTestBus(t)
	defer cancel()
	p := &Publisher{Sch: sch, Bus: bus}
	ev, err := p.Publish(Request{Type: "Escalate", Involved: "Agent/global/esc-1"}, adminSub())
	if err != nil {
		t.Fatalf("Publish() 意外失败: %v", err)
	}
	if ev.Source != SourceHuman {
		t.Errorf("Source = %q, 要恒 %q（来源=服务侧判定非载荷参数）", ev.Source, SourceHuman)
	}
	if ev.Level != "error" {
		t.Errorf("Level = %q, 要注册表 defaultLevel 投影=error（Escalate 权威分诊）", ev.Level)
	}
	if ev.Involved != "Agent/global/esc-1" {
		t.Errorf("Involved = %q, 要透传", ev.Involved)
	}
	got := <-ch
	if got.ID != ev.ID {
		t.Errorf("投递事件 ID=%q 与回执 %q 不同源", got.ID, ev.ID)
	}
	if got.Source != SourceHuman {
		t.Errorf("总线侧 Source=%q, 要 human（订阅方看到的来源与回执一致）", got.Source)
	}
}

func TestPublishLevelDefaultAndVocabulary(t *testing.T) {
	bus, sch, _, cancel := pubTestBus(t)
	defer cancel()
	p := &Publisher{Sch: sch, Bus: bus}
	// 缺省=defaultLevel（HeartbeatLost→error）
	ev, err := p.Publish(Request{Type: "HeartbeatLost"}, adminSub())
	if err != nil || ev.Level != "error" {
		t.Fatalf("缺省 level 投影失败: ev=%+v err=%v", ev, err)
	}
	// 显式 level 透传
	ev, err = p.Publish(Request{Type: "DocApplied", Level: "warn"}, adminSub())
	if err != nil || ev.Level != "warn" {
		t.Fatalf("显式 level 透传失败: ev=%+v err=%v", ev, err)
	}
	// 越词拒绝
	if _, err := p.Publish(Request{Type: "DocApplied", Level: "loud"}, adminSub()); !errors.Is(err, ErrLevel) {
		t.Errorf("越词 level err=%v, 要 ErrLevel", err)
	}
}

func TestPublishTypeGates(t *testing.T) {
	bus, sch, _, cancel := pubTestBus(t)
	defer cancel()
	p := &Publisher{Sch: sch, Bus: bus}
	if _, err := p.Publish(Request{Type: ""}, adminSub()); !errors.Is(err, ErrNoType) {
		t.Errorf("缺 type err=%v, 要 ErrNoType", err)
	}
	if _, err := p.Publish(Request{Type: "Mystery"}, adminSub()); !errors.Is(err, ErrNotRegistered) {
		t.Errorf("未注册类型 err=%v, 要 ErrNotRegistered（类型注册闸门）", err)
	}
}

func TestPublishPermissionDomainSeparate(t *testing.T) {
	bus, sch, ch, cancel := pubTestBus(t)
	defer cancel()
	sys := &permission.System{}
	// 授权面一次装载两 role：ev-publisher 只持 publish×events（无对象 write）、
	// obj-writer 只持对象 write（无 publish）——两域分立的两面各取一个主体。
	if err := sys.LoadObjects([]permission.ObjectDoc{
		roleDoc("ev-publisher", "events", "publish"),
		roleDoc("obj-writer", "global", "read, write"),
	}); err != nil {
		t.Fatalf("权限对象装载失败: %v", err)
	}
	p := &Publisher{Sch: sch, Bus: bus, Perm: sys}
	pub := permission.Subject{User: "human-1", Roles: []string{"ev-publisher"}}
	if _, err := p.Publish(Request{Type: "Escalate"}, &pub); err != nil {
		t.Fatalf("持 publish 主体注入意外拒绝: %v", err)
	}
	<-ch // 同主干投递确认
	// 分立面①：同一主体对象写被拒（publish 不产生 write——apply 域判定直接查证）
	if err := sys.Check(pub, "global", "write", permission.ByRole); err == nil {
		t.Errorf("publish×events 主体竟通过对象 write 判定（两域未分立）")
	}
	// 分立面②：仅持对象 write 的主体注入被拒（write 不产生 publish）
	writer := permission.Subject{User: "human-2", Roles: []string{"obj-writer"}}
	_, err := p.Publish(Request{Type: "Escalate"}, &writer)
	if !errors.Is(err, ErrDenied) {
		t.Errorf("仅对象 write 主体注入 err=%v, 要 ErrDenied（write×scope ≠ publish×events）", err)
	}
	// 匿名拒绝（无 role 落点 fail-closed）
	if _, err := p.Publish(Request{Type: "Escalate"}, &permission.Subject{User: "anon"}); !errors.Is(err, ErrDenied) {
		t.Errorf("匿名注入 err=%v, 要 ErrDenied", err)
	}
	// 错误文本携动作×域判定摘要（permission.md 规则 3 留痕口径）
	if _, err := p.Publish(Request{Type: "Escalate"}, &writer); err == nil || !strings.Contains(err.Error(), "publish on events") {
		t.Errorf("拒绝文本缺动作×域判定摘要: %v", err)
	}
}

func TestPublishDedupWindowAggregates(t *testing.T) {
	bus, sch, ch, cancel := pubTestBus(t)
	defer cancel()
	p := &Publisher{Sch: sch, Bus: bus}
	key := "wf-x:quota"
	ev1, err := p.Publish(Request{Type: "Escalate", DedupKey: key}, adminSub())
	if err != nil {
		t.Fatalf("首条注入失败: %v", err)
	}
	if _, err := p.Publish(Request{Type: "Escalate", DedupKey: key}, adminSub()); err != nil {
		t.Fatalf("窗口内重入失败: %v", err)
	}
	// 窗口内同因第二条不投递不归档：频道只应有第一条（非阻塞读验证）
	select {
	case got := <-ch:
		if got.DedupKey != key || got.ID != ev1.ID {
			t.Errorf("首条 DedupKey=%q ID=%q", got.DedupKey, got.ID)
		}
	default:
		t.Fatalf("首条未投递")
	}
	select {
	case extra := <-ch:
		t.Errorf("窗口内重入竟投递第二条: %+v", extra)
	default:
		// 聚合正确：只投第一条
	}
}

func TestPublishNilPermFailClosed(t *testing.T) {
	bus, sch, _, cancel := pubTestBus(t)
	defer cancel()
	p := &Publisher{Sch: sch, Bus: bus} // Perm=nil → 空授权面（非 admin 主体全拒）
	if _, err := p.Publish(Request{Type: "Escalate"}, &permission.Subject{User: "no-one"}); !errors.Is(err, ErrDenied) {
		t.Errorf("nil Perm 注入 err=%v, 要 ErrDenied（宁拒不降级）", err)
	}
}

func TestPublishCarriesRunTarget(t *testing.T) {
	bus, sch, ch, cancel := pubTestBus(t)
	defer cancel()
	p := &Publisher{Sch: sch, Bus: bus}
	ev, err := p.Publish(Request{Type: "Escalate", Involved: "Workflow/global/w", Run: 11}, adminSub())
	if err != nil {
		t.Fatalf("发布失败: %v", err)
	}
	if ev.Run != 11 {
		t.Fatalf("回执应携带 run target=11，得 %d", ev.Run)
	}
	select {
	case got := <-ch:
		if got.Run != 11 {
			t.Fatalf("总线事件应携带 run target=11，得 %d", got.Run)
		}
	case <-time.After(time.Second):
		t.Fatal("订阅者未收到事件")
	}
}
