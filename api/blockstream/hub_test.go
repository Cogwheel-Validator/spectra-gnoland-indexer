package blockstream_test

import (
	"context"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/blockstream"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakeChain struct {
	mu      sync.Mutex
	tip     uint64
	queries atomic.Int64
}

func (f *fakeChain) advance(n uint64) {
	f.mu.Lock()
	f.tip += n
	f.mu.Unlock()
}

func (f *fakeChain) latest(context.Context) (uint64, error) {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.tip, nil
}

func (f *fakeChain) after(_ context.Context, after uint64, limit int) ([]uint64, error) {
	f.queries.Add(1)
	f.mu.Lock()
	defer f.mu.Unlock()
	var out []uint64
	for h := after + 1; h <= f.tip && len(out) < limit; h++ {
		out = append(out, h)
	}
	return out, nil
}

func newTestHub(f *fakeChain) *blockstream.Hub[uint64] {
	return blockstream.NewHub("test", 5*time.Millisecond, f.after, f.latest, func(h uint64) uint64 { return h })
}

func recv(t *testing.T, ch <-chan uint64) uint64 {
	t.Helper()
	select {
	case v, ok := <-ch:
		require.True(t, ok, "channel closed")
		return v
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for block")
		return 0
	}
}

func TestHub_BroadcastsNewBlocksToAllSubscribers(t *testing.T) {
	t.Parallel()
	f := &fakeChain{tip: 100}
	h := newTestHub(f)

	a, _, unsubA, err := h.Subscribe(context.Background())
	require.NoError(t, err)
	defer unsubA()
	b, _, unsubB, err := h.Subscribe(context.Background())
	require.NoError(t, err)
	defer unsubB()

	f.advance(2)

	assert.Equal(t, uint64(101), recv(t, a))
	assert.Equal(t, uint64(102), recv(t, a))
	assert.Equal(t, uint64(101), recv(t, b))
	assert.Equal(t, uint64(102), recv(t, b))
}

func TestHub_StopsPollingWithoutSubscribers(t *testing.T) {
	t.Parallel()
	f := &fakeChain{tip: 1}
	h := newTestHub(f)

	_, _, unsub, err := h.Subscribe(context.Background())
	require.NoError(t, err)
	time.Sleep(20 * time.Millisecond)
	unsub()
	unsub() // idempotent
	assert.Equal(t, 0, h.SubscriberCount())

	time.Sleep(10 * time.Millisecond)
	before := f.queries.Load()
	time.Sleep(30 * time.Millisecond)
	assert.Equal(t, before, f.queries.Load(), "poller kept running with no subscribers")
}

func TestHub_DropsSlowSubscriber(t *testing.T) {
	t.Parallel()
	f := &fakeChain{tip: 1}
	h := newTestHub(f)

	slow, _, unsub, err := h.Subscribe(context.Background())
	require.NoError(t, err)
	defer unsub()
	time.Sleep(20 * time.Millisecond)
	f.advance(500)

	deadline := time.After(time.Second)
	for {
		select {
		case _, ok := <-slow:
			if !ok {
				return // dropped and closed, as intended
			}
		case <-deadline:
			t.Fatal("slow subscriber was never dropped")
		}
		// Stop draining so the buffer overflows.
		time.Sleep(50 * time.Millisecond)
		if h.SubscriberCount() == 0 {
			return
		}
	}
}

func TestHub_SubscribeReturnsStartHeightAndMissesNothing(t *testing.T) {
	t.Parallel()
	f := &fakeChain{tip: 100}
	h := newTestHub(f)

	ch, start, unsub, err := h.Subscribe(context.Background())
	require.NoError(t, err)
	defer unsub()
	assert.Equal(t, uint64(100), start)

	f.advance(1)
	assert.Equal(t, uint64(101), recv(t, ch))
}

func TestHub_EmptyChainDoesNotSkipFirstBlocks(t *testing.T) {
	t.Parallel()
	f := &fakeChain{}
	h := newTestHub(f)

	ch, start, unsub, err := h.Subscribe(context.Background())
	require.NoError(t, err)
	defer unsub()
	assert.Equal(t, uint64(0), start)

	f.advance(2)
	assert.Equal(t, uint64(1), recv(t, ch))
	assert.Equal(t, uint64(2), recv(t, ch))
}
