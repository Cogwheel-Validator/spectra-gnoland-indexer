// Package blockstream fans newly indexed blocks out to any number of
// long-lived subscribers (SSE connections) from a single database poller per
// chain.
//
// The poller only runs while at least one subscriber is connected, so an idle
// API costs nothing; with N connected clients it still issues one query per
// tick, not N.
package blockstream

import (
	"context"
	"log/slog"
	"sync"
	"time"
)

// DefaultBatchLimit caps how many blocks a single poll (or a Last-Event-ID
// backfill) returns, so a subscriber reconnecting after a long gap does not
// trigger an unbounded query.
const DefaultBatchLimit = 50

// subscriberBuffer is how many blocks a subscriber may lag behind before it is
// dropped. A dropped subscriber's channel is closed; the SSE client then
// reconnects with Last-Event-ID and backfills what it missed.
const subscriberBuffer = 64

// FetchAfterFunc returns blocks with height > after, ascending, at most limit.
type FetchAfterFunc[T any] func(ctx context.Context, after uint64, limit int) ([]T, error)

// LatestHeightFunc returns the current tip height.
type LatestHeightFunc func(ctx context.Context) (uint64, error)

// Hub polls for new blocks and broadcasts them to subscribers.
type Hub[T any] struct {
	name      string
	interval  time.Duration
	fetch     FetchAfterFunc[T]
	latest    LatestHeightFunc
	heightOf  func(T) uint64
	mu        sync.Mutex
	subs      map[chan T]struct{}
	cancelRun context.CancelFunc
}

// NewHub creates a hub; name is only used in log lines.
func NewHub[T any](
	name string,
	interval time.Duration,
	fetch FetchAfterFunc[T],
	latest LatestHeightFunc,
	heightOf func(T) uint64,
) *Hub[T] {
	return &Hub[T]{
		name:     name,
		interval: interval,
		fetch:    fetch,
		latest:   latest,
		heightOf: heightOf,
		subs:     make(map[chan T]struct{}),
	}
}

// Subscribe registers a new subscriber and starts the poller if it is the
// first one. The returned channel is closed if the subscriber falls too far
// behind; the returned func unsubscribes (safe to call more than once).
func (h *Hub[T]) Subscribe() (<-chan T, func()) {
	ch := make(chan T, subscriberBuffer)

	h.mu.Lock()
	h.subs[ch] = struct{}{}
	if h.cancelRun == nil {
		ctx, cancel := context.WithCancel(context.Background())
		h.cancelRun = cancel
		go h.run(ctx)
	}
	h.mu.Unlock()

	var once sync.Once
	return ch, func() {
		once.Do(func() {
			h.mu.Lock()
			defer h.mu.Unlock()
			if _, ok := h.subs[ch]; ok {
				delete(h.subs, ch)
				close(ch)
			}
			if len(h.subs) == 0 && h.cancelRun != nil {
				h.cancelRun()
				h.cancelRun = nil
			}
		})
	}
}

// SubscriberCount reports the number of connected subscribers.
func (h *Hub[T]) SubscriberCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.subs)
}

func (h *Hub[T]) run(ctx context.Context) {
	var last uint64
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		last = h.poll(ctx, last)
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

// poll performs one tick: it seeds the starting height on the first call,
// afterwards broadcasts every block newer than last. It returns the new last
// height (unchanged on error, so the next tick retries).
func (h *Hub[T]) poll(ctx context.Context, last uint64) uint64 {
	if last == 0 {
		tip, err := h.latest(ctx)
		if err != nil {
			if ctx.Err() == nil {
				slog.Error("blockstream: latest height", "chain", h.name, "err", err)
			}
			return 0
		}
		return tip
	}

	blocks, err := h.fetch(ctx, last, DefaultBatchLimit)
	if err != nil && ctx.Err() == nil {
		slog.Error("blockstream: fetch", "chain", h.name, "after", last, "err", err)
	}
	for _, b := range blocks {
		h.broadcast(ctx, b)
		if hb := h.heightOf(b); hb > last {
			last = hb
		}
	}
	return last
}

func (h *Hub[T]) broadcast(ctx context.Context, b T) {
	h.mu.Lock()
	defer h.mu.Unlock()
	// A run that was cancelled (last subscriber left) must not deliver into
	// subscribers of a newer run.
	if ctx.Err() != nil {
		return
	}
	for ch := range h.subs {
		select {
		case ch <- b:
		default:
			delete(h.subs, ch)
			close(ch)
		}
	}
}
