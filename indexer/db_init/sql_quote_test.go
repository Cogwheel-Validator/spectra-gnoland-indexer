package dbinit

import (
	"strings"
	"testing"
)

func TestBuildCreateUserSQL(t *testing.T) {
	got, err := buildCreateUserSQL(`bob"; DROP TABLE x;--`, `p'w\'; DROP--`)
	if err != nil {
		t.Fatal(err)
	}
	want := `CREATE USER "bob""; DROP TABLE x;--" WITH PASSWORD E'p''w\\''; DROP--'`
	if got != want {
		t.Errorf("got %s want %s", got, want)
	}
	for _, bad := range []string{"", "  ", "a\x00b"} {
		if _, err := buildCreateUserSQL(bad, "pw"); err == nil {
			t.Errorf("user %q should fail", bad)
		}
	}
	if _, err := buildCreateUserSQL("bob", "a\x00b"); err == nil {
		t.Error("NUL password should fail")
	}
}

func TestBuildPrivilegesSQL(t *testing.T) {
	got, err := buildPrivilegesSQL(`u"x`, "reader", []string{"blocks", `t"; DROP`})
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(got, `GRANT SELECT ON TABLE "blocks" TO "u""x";`) ||
		!strings.Contains(got, `"t""; DROP"`) {
		t.Errorf("unexpected sql: %s", got)
	}
	if _, err := buildPrivilegesSQL("u", "root", nil); err == nil {
		t.Error("invalid privilege should fail")
	}
	if _, err := buildPrivilegesSQL("", "reader", nil); err == nil {
		t.Error("empty user should fail")
	}
}
