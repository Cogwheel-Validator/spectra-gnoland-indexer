package timescaledb

import "testing"

func TestBuildCreateDatabaseSQL(t *testing.T) {
	got, err := buildCreateDatabaseSQL(`db"; DROP DATABASE x;--`)
	if err != nil {
		t.Fatal(err)
	}
	if want := `CREATE DATABASE "db""; DROP DATABASE x;--"`; got != want {
		t.Errorf("got %s want %s", got, want)
	}
	if got, _ := buildCreateDatabaseSQL("gnoland-test"); got != `CREATE DATABASE "gnoland-test"` {
		t.Errorf("got %s", got)
	}
	for _, bad := range []string{"", "  ", "a\x00b"} {
		if _, err := buildCreateDatabaseSQL(bad); err == nil {
			t.Errorf("%q should fail", bad)
		}
	}
}
