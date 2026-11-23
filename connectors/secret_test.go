// secret_test.go —— 契约层引用式取值解析验收（secret:/env: 放行；nil store 缺口可见；
// 明文/vault:/keychain: 拒绝）。
package connectors

import (
	"errors"
	"os"
	"testing"
)

type stubStore map[string]string

func (s stubStore) Secret(name string) (string, error) {
	v, ok := s[name]
	if !ok {
		return "", errors.New("absent")
	}
	return v, nil
}

func TestResolveSecret(t *testing.T) {
	store := stubStore{"smtp-auth": "hunter2"}
	t.Setenv("CT_TEST_SECRET", "envval")
	getenv := func(k string) string { return os.Getenv(k) }
	cases := []struct {
		src, want string
		wantErr   bool
	}{
		{"secret:smtp-auth", "hunter2", false},
		{"env:CT_TEST_SECRET", "envval", false},
		{"secret:absent", "", true},
		{"env:CT_TEST_UNSET", "", true},
		{"vault:secret/k", "", true},
		{"keychain:app", "", true},
		{"plaintext", "", true},
		{"secret:", "", true},
	}
	for _, c := range cases {
		got, err := ResolveSecret(c.src, store, getenv)
		if c.wantErr && err == nil {
			t.Errorf("%q: 期望报错，实得 %q", c.src, got)
		}
		if !c.wantErr && (err != nil || got != c.want) {
			t.Errorf("%q: 期望 %q，实得 %q err=%v", c.src, c.want, got, err)
		}
	}
	// nil store：secret: 来源=缺口可见报错（部署侧未挂秘密库）。
	if _, err := ResolveSecret("secret:smtp-auth", nil, getenv); err == nil {
		t.Error("nil store 应报错")
	}
	// nil getenv：缺省进程环境。
	if _, err := ResolveSecret("env:PATH", nil, nil); err != nil {
		t.Errorf("nil getenv 缺省进程环境: %v", err)
	}
}
