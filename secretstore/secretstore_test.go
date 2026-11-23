// secretstore_test.go —— 部署侧秘密存储验收：落/取/删、脱敏列表面（零值出）、
// 文件权限 0600、Resolve 引用式解析（secret:/env: 放行，明文/vault:/keychain: 拒绝）。
package secretstore

import (
	"errors"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestSetGetDelete(t *testing.T) {
	s, err := OpenSQLite(filepath.Join(t.TempDir(), "secrets.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.Set("smtp-auth", "hunter2"); err != nil {
		t.Fatalf("set: %v", err)
	}
	if err := s.Set("smtp-auth", "hunter3"); err != nil { // upsert 覆盖
		t.Fatal(err)
	}
	v, err := s.Secret("smtp-auth")
	if err != nil || v != "hunter3" {
		t.Fatalf("get: %q err=%v", v, err)
	}
	if _, err := s.Secret("absent"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("absent: err=%v", err)
	}
	if err := s.Delete("smtp-auth"); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Secret("smtp-auth"); !errors.Is(err, ErrNotFound) {
		t.Fatalf("deleted: err=%v", err)
	}
	if err := s.Delete("absent"); err != nil { // 幂等
		t.Fatalf("delete absent: %v", err)
	}
	if err := s.Set("empty", ""); err == nil {
		t.Fatal("空值应拒绝")
	}
}

func TestListNeverCarriesValues(t *testing.T) {
	s, err := OpenSQLite(filepath.Join(t.TempDir(), "secrets.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.Set("k1", "topsecret"); err != nil {
		t.Fatal(err)
	}
	rows, err := s.List()
	if err != nil || len(rows) != 1 || rows[0].Name != "k1" {
		t.Fatalf("list: %+v err=%v", rows, err)
	}
	if rows[0].UpdatedAt.IsZero() {
		t.Fatal("updated_at 应有值")
	}
	n, err := s.Count()
	if err != nil || n != 1 {
		t.Fatalf("count: %d err=%v", n, err)
	}
	// 列表面投影类型 SecretInfo 机械断言：无值字段（脱敏红线——列表/查询面零值出）。
	typ := reflect.TypeOf(SecretInfo{})
	for i := 0; i < typ.NumField(); i++ {
		low := strings.ToLower(typ.Field(i).Name)
		if strings.Contains(low, "value") || strings.Contains(low, "secret") {
			t.Fatalf("SecretInfo 携带值字段: %s", typ.Field(i).Name)
		}
	}
}

func TestFilePerms0600(t *testing.T) {
	p := filepath.Join(t.TempDir(), "sub", "secrets.db")
	s, err := OpenSQLite(p)
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	fi, err := os.Stat(p)
	if err != nil {
		t.Fatal(err)
	}
	if perm := fi.Mode().Perm(); perm != 0o600 {
		t.Fatalf("秘密库权限=%o（期望 600）", perm)
	}
}

func TestResolve(t *testing.T) {
	s, err := OpenSQLite(filepath.Join(t.TempDir(), "secrets.db"))
	if err != nil {
		t.Fatal(err)
	}
	defer s.Close()
	if err := s.Set("smtp-auth", "hunter2"); err != nil {
		t.Fatal(err)
	}
	t.Setenv("QQ_TEST_SECRET", "envval")
	cases := []struct {
		src, want string
		wantErr   bool
	}{
		{"secret:smtp-auth", "hunter2", false},
		{"env:QQ_TEST_SECRET", "envval", false},
		{"secret:absent", "", true},
		{"env:QQ_TEST_UNSET", "", true},
		{"vault:secret/k", "", true},
		{"keychain:app", "", true},
		{"plaintext", "", true},
		{"secret:", "", true},
	}
	for _, c := range cases {
		got, err := s.Resolve(c.src)
		if c.wantErr && err == nil {
			t.Errorf("%q: 期望报错，实得 %q", c.src, got)
		}
		if !c.wantErr && (err != nil || got != c.want) {
			t.Errorf("%q: 期望 %q，实得 %q err=%v", c.src, c.want, got, err)
		}
	}
}
