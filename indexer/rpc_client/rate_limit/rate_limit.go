package rate_limit

import (
	"time"
)

// ChannelRateLimiter is a channel-based rate limiter
type ChannelRateLimiter struct {
	tokens   chan struct{}
	ticker   *time.Ticker
	capacity int
	done     chan struct{}
}

// NewChannelRateLimiter creates a new ChannelRateLimiter
//
// Args:
//   - requestsAllowed: the number of requests allowed per time window
//   - timeWindow: the time window for rate limiting
//
// Returns:
//   - *ChannelRateLimiter: the rate limiter
//   - error: if the rate limiter fails to create
//
// Example:
//
//	limiter := NewChannelRateLimiter(100, 1*time.Minute)
//	defer limiter.Close()
//
//	limiter.Allow() // returns true if the request is allowed
//	limiter.Wait() // blocks until the request is allowed
//	limiter.Close() // closes the rate limiter
//
//	limiter.GetStatus() // returns the status of the rate limiter
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

// refillTokens refills the tokens in the rate limiter
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

// Allow returns true if the request is allowed
func (r *ChannelRateLimiter) Allow() bool {
	select {
	case <-r.tokens:
		return true
	default:
		return false
	}
}

// Wait blocks until the request is allowed
func (r *ChannelRateLimiter) Wait() {
	<-r.tokens
}

// Close closes the rate limiter
func (r *ChannelRateLimiter) Close() {
	close(r.done)
	r.ticker.Stop()
}

// GetStatus returns the current status of the rate limiter
//
// Returns:
//   - ChannelRateLimiterStatus: the status of the rate limiter
//
// Example:
//
//	status := limiter.GetStatus()
//	fmt.Println(status)
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
//
// Fields:
//   - TokensAvailable: the number of tokens currently available
//   - Capacity: the maximum number of tokens
//   - IsEmpty: whether the bucket is empty (will block on next request)
//   - IsFull: whether the bucket is full
type ChannelRateLimiterStatus struct {
	TokensAvailable int  // Number of tokens currently available
	Capacity        int  // Maximum number of tokens
	IsEmpty         bool // Whether the bucket is empty (will block on next request)
	IsFull          bool // Whether the bucket is full
}
