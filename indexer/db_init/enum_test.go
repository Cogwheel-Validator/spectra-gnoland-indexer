package dbinit

import "testing"

func TestBuildAddEnumValueSQL(t *testing.T) {
	cases := map[string]string{
		"test-1":  `ALTER TYPE "chain_name" ADD VALUE 'test-1'`,
		"a'b":     `ALTER TYPE "chain_name" ADD VALUE 'a''b'`,
		"a';--":   `ALTER TYPE "chain_name" ADD VALUE 'a'';--'`,
		"a;/*x*/": `ALTER TYPE "chain_name" ADD VALUE 'a;/*x*/'`,
	}
	for in, want := range cases {
		if got := buildAddEnumValueSQL("chain_name", in); got != want {
			t.Errorf("%q: got %s want %s", in, got, want)
		}
	}
}
