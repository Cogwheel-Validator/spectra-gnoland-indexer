package query

import (
	"time"
)

// ChannelRateLimiter - Channel-based rate limiter (most Go-idiomatic)
type ChannelRateLimiter struct {
	tokens   chan struct{}
	ticker   *time.Ticker
	capacity int
	done     chan struct{}
}

func NewChannelRateLimiter(requestsAllowed int, timeWindow time.Duration) *ChannelRateLimiter {
	rl := &ChannelRateLimiter{
		tokens:   make(chan struct{}, requestsAllowed),
		capacity: requestsAllowed,
		done:     make(chan struct{}),
	}

	// Fill the bucket initially
	for range requestsAllowed {
		rl.tokens <- struct{}{}
	}

	// Start the refill goroutine
	intervalPerToken := timeWindow / time.Duration(requestsAllowed)
	rl.ticker = time.NewTicker(intervalPerToken)

	go rl.refillTokens()

	return rl
}

func (r *ChannelRateLimiter) refillTokens() {
	for {
		select {
		case <-r.ticker.C:
			select {
			case r.tokens <- struct{}{}:
			default:
				// Bucket is full, skip this refill
			}
		case <-r.done:
			return
		}
	}
}

func (r *ChannelRateLimiter) Allow() bool {
	select {
	case <-r.tokens:
		return true
	default:
		return false
	}
}

func (r *ChannelRateLimiter) Wait() {
	<-r.tokens
}

func (r *ChannelRateLimiter) Close() {
	close(r.done)
	r.ticker.Stop()
}

// GetStatus returns the current status of the rate limiter
func (r *ChannelRateLimiter) GetStatus() ChannelRateLimiterStatus {
	tokensAvailable := len(r.tokens)
	return ChannelRateLimiterStatus{
		TokensAvailable: tokensAvailable,
		Capacity:        r.capacity,
		IsEmpty:         tokensAvailable == 0,
		IsFull:          tokensAvailable == r.capacity,
	}
}

// ChannelRateLimiterStatus provides information about the rate limiter state
type ChannelRateLimiterStatus struct {
	TokensAvailable int  // Number of tokens currently available
	Capacity        int  // Maximum number of tokens
	IsEmpty         bool // Whether the bucket is empty (will block on next request)
	IsFull          bool // Whether the bucket is full
}

// Legacy RateLimitBucket (keeping for backward compatibility)
type RateLimitBucket struct {
	LastRequestTime time.Time
	RequestsCount   int
	RequestsAllowed int
	TimeWindow      time.Duration
}

func NewRateLimitBucket(RequestsAllowed int) *RateLimitBucket {
	return &RateLimitBucket{
		LastRequestTime: time.Now(),
		RequestsAllowed: RequestsAllowed,
		TimeWindow:      1 * time.Minute,
	}
}
