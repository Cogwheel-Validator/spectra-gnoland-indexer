package migrations

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestRegistry(t *testing.T) {
	if all[0].Version != BaselineVersion || all[0].SQL != "" {
		t.Fatalf("first migration must be the empty baseline, got %+v", all[0])
	}
	for i, m := range all {
		if m.Version != int32(i+1) {
			t.Errorf("migration %d has version %d, versions must be sequential from 1", i, m.Version)
		}
		if m.Release == "" || m.Description == "" {
			t.Errorf("migration %d needs a release and a description", m.Version)
		}
		if i > 0 && strings.TrimSpace(m.SQL) == "" {
			t.Errorf("migration %d has no SQL", m.Version)
		}
		// The SQL runs inside the runner's transaction.
		for _, kw := range []string{"BEGIN;", "COMMIT;"} {
			if strings.Contains(strings.ToUpper(m.SQL), kw) {
				t.Errorf("migration %d must not manage its own transaction (%s)", m.Version, kw)
			}
		}
	}
	if Latest() != all[len(all)-1].Version {
		t.Errorf("Latest() = %d", Latest())
	}
}

func state(versions ...int32) State {
	s := State{Applied: map[int32]time.Time{}, Failed: map[int32]time.Time{}}
	for _, v := range versions {
		s.Applied[v] = time.Now()
	}
	return s
}

func TestVerify(t *testing.T) {
	everything := make([]int32, 0, len(all))
	for _, m := range all {
		everything = append(everything, m.Version)
	}

	tests := []struct {
		name  string
		state State
		want  error
	}{
		{"empty", state(), ErrNotVersioned},
		{"baseline only", state(BaselineVersion), ErrPending},
		{"up to date", state(everything...), nil},
		{"gap", state(everything[1:]...), ErrPending},
		{"newer", state(append(everything, Latest()+1)...), ErrTooNew},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.state.Verify()
			if !errors.Is(err, tt.want) || (tt.want == nil) != (err == nil) {
				t.Errorf("Verify() = %v, want %v", err, tt.want)
			}
		})
	}
}

func TestPendingAndCurrent(t *testing.T) {
	s := state(BaselineVersion)
	s.Failed[2] = time.Now()
	if s.Current() != BaselineVersion {
		t.Errorf("Current() = %d", s.Current())
	}
	pending := s.Pending()
	if len(pending) != len(all)-1 || pending[0].Version != 2 {
		t.Errorf("Pending() = %+v", pending)
	}
}
