package addresscache

import (
	"context"
	"errors"
	"slices"
	"testing"
)

type resolveCall struct {
	addresses  []string
	chain      string
	validators bool
}

type mockDB struct {
	// record calls
	calls []resolveCall

	// behavior controls
	existing map[string]int32
	// errors by attempt index (global)
	errs []error
	// addresses the database leaves out of the result, as if a concurrent
	// writer inserted them mid-statement; each is left out only once
	skipOnce map[string]bool
}

func newMockDB(existing map[string]int32) *mockDB {
	m := &mockDB{existing: map[string]int32{}, skipOnce: map[string]bool{}}
	for k, v := range existing {
		m.existing[k] = v
	}
	return m
}

func (m *mockDB) ResolveAddresses(ctx context.Context, addresses []string, chainName string, validators bool) (map[string]int32, error) {
	m.calls = append(m.calls, resolveCall{slices.Clone(addresses), chainName, validators})

	idx := len(m.calls) - 1
	if idx < len(m.errs) && m.errs[idx] != nil {
		return nil, m.errs[idx]
	}

	res := map[string]int32{}
	for _, a := range addresses {
		if _, ok := m.existing[a]; !ok {
			// assign a synthetic id: len(existing)+1
			m.existing[a] = int32(len(m.existing) + 1)
		}
		if m.skipOnce[a] {
			delete(m.skipOnce, a)
			continue
		}
		res[a] = m.existing[a]
	}
	return res, nil
}

func newCacheForTest(t *testing.T, existing map[string]int32, validators bool, size int) (*AddressCache, *mockDB) {
	t.Helper()
	m := newMockDB(existing)
	c, err := NewAddressCache(m, "chain", validators, size)
	if err != nil {
		t.Fatalf("NewAddressCache: %v", err)
	}
	return c, m
}

func TestNewAddressCache_RejectsInvalidSize(t *testing.T) {
	if _, err := NewAddressCache(newMockDB(nil), "chain", false, 0); err == nil {
		t.Fatal("expected an error for size 0")
	}
}

func TestAddressSolver_ResolvesAndCaches(t *testing.T) {
	cache, m := newCacheForTest(t, map[string]int32{"a": 10}, false, 10)

	resolved, err := cache.AddressSolver([]string{"a", "b"}, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.GetAddress("a") != 10 {
		t.Fatalf("expected a to keep id 10, got %d", resolved.GetAddress("a"))
	}
	if resolved.GetAddress("b") == 0 {
		t.Fatal("expected b to be resolved")
	}
	if len(m.calls) != 1 {
		t.Fatalf("expected 1 db call, got %d", len(m.calls))
	}
	if !cache.cache.Contains("a") || !cache.cache.Contains("b") {
		t.Fatal("expected a and b to be cached")
	}
}

func TestAddressSolver_CacheHitSkipsDB(t *testing.T) {
	cache, m := newCacheForTest(t, nil, false, 10)

	if _, err := cache.AddressSolver([]string{"a", "b"}, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	resolved, err := cache.AddressSolver([]string{"a", "b"}, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.calls) != 1 {
		t.Fatalf("expected the second call to be served from cache, got %d db calls", len(m.calls))
	}
	if resolved.GetAddress("a") == 0 || resolved.GetAddress("b") == 0 {
		t.Fatal("expected cached ids in the result")
	}
}

func TestAddressSolver_OnlyQueriesMissing(t *testing.T) {
	cache, m := newCacheForTest(t, nil, false, 10)

	if _, err := cache.AddressSolver([]string{"a"}, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if _, err := cache.AddressSolver([]string{"a", "b", "b", ""}, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if got := m.calls[1].addresses; !slices.Equal(got, []string{"b"}) {
		t.Fatalf("expected second query for [b], got %v", got)
	}
}

func TestAddressSolver_EmptyInput(t *testing.T) {
	cache, m := newCacheForTest(t, nil, false, 10)

	resolved, err := cache.AddressSolver(nil, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(resolved) != 0 || len(m.calls) != 0 {
		t.Fatal("expected no db calls and an empty result on empty input")
	}
}

func TestAddressSolver_PassesChainAndValidatorFlag(t *testing.T) {
	cache, m := newCacheForTest(t, nil, true, 10)

	if _, err := cache.AddressSolver([]string{"val1"}, 1); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(m.calls) != 1 || !m.calls[0].validators || m.calls[0].chain != "chain" {
		t.Fatalf("expected a validator query for chain, got %+v", m.calls)
	}
}

func TestAddressSolver_RetriesOnError(t *testing.T) {
	cache, m := newCacheForTest(t, nil, false, 10)
	// fail first attempt, succeed second
	m.errs = []error{errors.New("temp")}

	resolved, err := cache.AddressSolver([]string{"x", "y"}, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.GetAddress("x") == 0 || resolved.GetAddress("y") == 0 {
		t.Fatal("expected x and y to be resolved after retry")
	}
	if len(m.calls) != 2 {
		t.Fatalf("expected 2 attempts, got %d", len(m.calls))
	}
}

func TestAddressSolver_HonorsRetryAttempts(t *testing.T) {
	cache, m := newCacheForTest(t, nil, false, 10)
	m.errs = []error{errors.New("1"), errors.New("2"), errors.New("3")}

	if _, err := cache.AddressSolver([]string{"a"}, 3); err == nil {
		t.Fatal("expected an error after all attempts failed")
	}
	if len(m.calls) != 3 {
		t.Fatalf("expected 3 attempts, got %d", len(m.calls))
	}
	if cache.cache.Contains("a") {
		t.Fatal("expected a to not be cached")
	}
}

func TestAddressSolver_RequeriesAddressesLeftOut(t *testing.T) {
	cache, m := newCacheForTest(t, nil, false, 10)
	m.skipOnce["b"] = true

	resolved, err := cache.AddressSolver([]string{"a", "b"}, 2)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if resolved.GetAddress("b") == 0 {
		t.Fatal("expected b to be resolved on the second query")
	}
	if len(m.calls) != 2 || !slices.Equal(m.calls[1].addresses, []string{"b"}) {
		t.Fatalf("expected a second query for [b], got %+v", m.calls)
	}
}

func TestAddressSolver_PartialResultOnError(t *testing.T) {
	cache, m := newCacheForTest(t, nil, false, 10)
	m.skipOnce["b"] = true
	m.errs = []error{nil, errors.New("down")}

	resolved, err := cache.AddressSolver([]string{"a", "b"}, 2)
	if err == nil {
		t.Fatal("expected an error for the unresolved address")
	}
	if resolved.GetAddress("a") == 0 {
		t.Fatal("expected a to be in the partial result")
	}
	if resolved.GetAddress("b") != 0 {
		t.Fatal("expected b to be missing from the partial result")
	}
}

func TestAddressSolver_ResultSurvivesEviction(t *testing.T) {
	// a batch larger than the cache must still resolve every address
	cache, _ := newCacheForTest(t, nil, false, 2)
	batch := []string{"a", "b", "c", "d", "e"}

	resolved, err := cache.AddressSolver(batch, 1)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	for _, addr := range batch {
		if resolved.GetAddress(addr) == 0 {
			t.Fatalf("expected %s to be resolved", addr)
		}
	}
	if cache.cache.Len() > 2 {
		t.Fatalf("expected the cache to hold at most 2 entries, got %d", cache.cache.Len())
	}
}

func TestAddressSolver_DoesNotOverwriteExistingIDs(t *testing.T) {
	cache, _ := newCacheForTest(t, map[string]int32{"a": 10}, false, 10)

	for range 2 {
		resolved, err := cache.AddressSolver([]string{"a"}, 1)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if got := resolved.GetAddress("a"); got != 10 {
			t.Fatalf("expected a to remain 10, got %d", got)
		}
	}
}
