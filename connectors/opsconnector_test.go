package connectors

import (
	"context"
	"sort"
	"testing"
)

// fakeOps 契约测试桩。
type fakeOps struct{ ops []string }

func (f *fakeOps) Type() string    { return "fake-ops" }
func (f *fakeOps) Version() string { return "v0.0.0-test" }
func (f *fakeOps) Ops() []string   { return f.ops }
func (f *fakeOps) Invoke(_ context.Context, _ string, _ []byte) ([]byte, error) {
	return nil, nil
}
func (f *fakeOps) Probe(_ context.Context) error { return nil }

func TestOpsConnectorRegistry(t *testing.T) {
	before := len(OpsConnectorTypes())
	a := &fakeOps{ops: []string{"op-a", "op-b"}}
	b := &fakeOps{ops: []string{"op-c"}}

	RegisterOpsConnector(a)
	got, ok := LookupOpsConnector("fake-ops")
	if !ok || got != OpsConnector(a) {
		t.Fatalf("LookupOpsConnector(fake-ops)= %v,%v 要 a,true", got, ok)
	}
	if got.Type() != "fake-ops" || got.Version() != "v0.0.0-test" {
		t.Fatalf("契约访问器不符: type=%s version=%s", got.Type(), got.Version())
	}
	// 同型重注册=后到覆盖（幂等安全，部署侧重配语义）。
	RegisterOpsConnector(b)
	got2, _ := LookupOpsConnector("fake-ops")
	if got2 != OpsConnector(b) {
		t.Fatalf("重注册应后到覆盖")
	}
	// 未注册类型不命中。
	if _, ok := LookupOpsConnector("absent-ops"); ok {
		t.Fatalf("absent-ops 不应命中")
	}
	// 类型清单含新注册且字典序。
	ts := OpsConnectorTypes()
	if len(ts) != before+1 {
		t.Fatalf("OpsConnectorTypes 长度=%d 要 %d", len(ts), before+1)
	}
	if !sort.StringsAreSorted(ts) {
		t.Fatalf("类型清单应字典序: %v", ts)
	}
}

func TestOpsConnectorHasOp(t *testing.T) {
	c := &fakeOps{ops: []string{"notify", "send-email"}}
	for _, op := range []string{"notify", "send-email"} {
		if !HasOp(c, op) {
			t.Fatalf("HasOp(%s)=false 要 true", op)
		}
	}
	if HasOp(c, "webhook-send") {
		t.Fatalf("未声明 op 应不命中")
	}
	if HasOp(c, "") {
		t.Fatalf("空 op 应不命中")
	}
}
