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
	// seeded and last describe the current run: once seeded, every block with
	// height <= last has already been broadcast and every later one will be.
	// ready is closed when the current run becomes seeded.
	seeded bool
	last   uint64
	ready  chan struct{}
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
func (h *Hub[T]) Subscribe(ctx context.Context) (<-chan T, uint64, func(), error) {
	ch := make(chan T, subscriberBuffer)

	h.mu.Lock()
	h.subs[ch] = struct{}{}
	if h.cancelRun == nil {
		runCtx, cancel := context.WithCancel(context.Background())
		h.cancelRun = cancel
		h.seeded = false
		h.last = 0
		h.ready = make(chan struct{})
		go h.run(runCtx, h.ready)
	}
	ready := h.ready
	h.mu.Unlock()

	var once sync.Once
	unsub := func() {
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

	select {
	case <-ready:
	case <-ctx.Done():
		unsub()
		return nil, 0, nil, ctx.Err()
	}

	// Read under the lock: broadcast advances last under the same lock, so the
	// blocks already delivered to others are exactly those <= start.
	h.mu.Lock()
	start := h.last
	h.mu.Unlock()
	return ch, start, unsub, nil
}

// SubscriberCount reports the number of connected subscribers.
func (h *Hub[T]) SubscriberCount() int {
	h.mu.Lock()
	defer h.mu.Unlock()
	return len(h.subs)
}

func (h *Hub[T]) run(ctx context.Context, ready chan struct{}) {
	ticker := time.NewTicker(h.interval)
	defer ticker.Stop()

	for {
		if h.isSeeded() {
			h.poll(ctx)
		} else {
			h.seed(ctx, ready)
		}
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
		}
	}
}

func (h *Hub[T]) isSeeded() bool {
	h.mu.Lock()
	defer h.mu.Unlock()
	return h.seeded
}

// seed reads the tip as the starting height. A tip of 0 (empty chain) is a
// valid seed, so the first indexed blocks are not skipped.
func (h *Hub[T]) seed(ctx context.Context, ready chan struct{}) {
	tip, err := h.latest(ctx)
	if err != nil {
		if ctx.Err() == nil {
			slog.Error("blockstream: latest height", "chain", h.name, "err", err)
		}
		return
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if ctx.Err() != nil {
		return
	}
	h.seeded = true
	h.last = tip
	close(ready)
}

// poll performs one tick: it broadcasts every block newer than the last
// broadcast height. On error the next tick retries from the same height.
func (h *Hub[T]) poll(ctx context.Context) {
	h.mu.Lock()
	last := h.last
	h.mu.Unlock()

	blocks, err := h.fetch(ctx, last, DefaultBatchLimit)
	if err != nil && ctx.Err() == nil {
		slog.Error("blockstream: fetch", "chain", h.name, "after", last, "err", err)
	}
	for _, b := range blocks {
		h.broadcast(ctx, b)
	}
}

func (h *Hub[T]) broadcast(ctx context.Context, b T) {
	h.mu.Lock()
	defer h.mu.Unlock()
	// A run that was cancelled (last subscriber left) must not deliver into
	// subscribers of a newer run.
	if ctx.Err() != nil {
		return
	}
	if hb := h.heightOf(b); hb > h.last {
		h.last = hb
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
