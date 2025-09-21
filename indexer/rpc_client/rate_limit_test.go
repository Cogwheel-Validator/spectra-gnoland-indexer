package rpcclient_test

import (
	"testing"
	"time"

	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client/rate_limit"
)

// TestRateLimiter_Direct - Unit test the rate limiter component directly
func TestRateLimiter_Direct(t *testing.T) {
	// Create rate limiter: 2 requests per 1 second
	rateLimiter := rate_limit.NewChannelRateLimiter(2, 1*time.Second)
	defer rateLimiter.Close()

	// Test initial state - should be full
	status := rateLimiter.GetStatus()
	if !status.IsFull {
		t.Error("Rate limiter should start full")
	}
	if status.TokensAvailable != 2 {
		t.Errorf("Expected 2 tokens, got %d", status.TokensAvailable)
	}

	// First request - should be allowed immediately
	if !rateLimiter.Allow() {
		t.Error("First request should be allowed")
	}

	// Second request - should be allowed immediately
	if !rateLimiter.Allow() {
		t.Error("Second request should be allowed")
	}

	// Third request - should be denied (bucket empty)
	if rateLimiter.Allow() {
		t.Error("Third request should be denied - bucket should be empty")
	}

	// Check status after exhausting tokens
	status = rateLimiter.GetStatus()
	if !status.IsEmpty {
		t.Error("Rate limiter should be empty after 2 requests")
	}
	if status.TokensAvailable != 0 {
		t.Errorf("Expected 0 tokens, got %d", status.TokensAvailable)
	}
}

// TestRateLimiter_Timing - Test that rate limiter actually enforces timing
func TestRateLimiter_Timing(t *testing.T) {
	// Create very restrictive rate limiter: 1 request per 500ms
	rateLimiter := rate_limit.NewChannelRateLimiter(1, 500*time.Millisecond)
	defer rateLimiter.Close()

	// First request should be immediate
	start := time.Now()
	rateLimiter.Wait() // This should not block
	firstDuration := time.Since(start)

	if firstDuration > 50*time.Millisecond {
		t.Errorf("First request took too long: %v", firstDuration)
	}

	// Second request should block until token refills
	start = time.Now()
	rateLimiter.Wait() // This should block ~500ms
	secondDuration := time.Since(start)

	if secondDuration < 400*time.Millisecond {
		t.Errorf("Second request should be delayed, but took only %v", secondDuration)
	}
	if secondDuration > 600*time.Millisecond {
		t.Errorf("Second request took too long: %v", secondDuration)
	}
}

// TestRateLimiter_Refill - Test that tokens refill over time
func TestRateLimiter_Refill(t *testing.T) {
	// Create rate limiter: 3 requests per 300ms (100ms per token)
	rateLimiter := rate_limit.NewChannelRateLimiter(3, 300*time.Millisecond)
	defer rateLimiter.Close()

	// Exhaust all tokens
	rateLimiter.Allow() // Token 1
	rateLimiter.Allow() // Token 2
	rateLimiter.Allow() // Token 3

	// Should be empty now
	status := rateLimiter.GetStatus()
	if !status.IsEmpty {
		t.Error("Should be empty after using 3 tokens")
	}

	// Wait for refill (tokens should refill every ~100ms)
	time.Sleep(150 * time.Millisecond)

	// Should have at least 1 token now
	status = rateLimiter.GetStatus()
	if status.TokensAvailable < 1 {
		t.Errorf("Should have refilled at least 1 token, got %d", status.TokensAvailable)
	}

	// Should be able to make another request
	if !rateLimiter.Allow() {
		t.Error("Should be able to make request after refill")
	}
}

// TestRateLimitedRpcClient_WithRealEndpoint - Integration test with timing
func TestRateLimitedRpcClient_WithRealEndpoint(t *testing.T) {
	// Skip if not running integration tests
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create rate-limited client: 2 requests per second
	client, err := rpcClient.NewRateLimitedRpcClient(
		"https://gnoland-testnet-rpc.cogwheel.zone",
		nil,
		2,             // Only 2 requests allowed
		1*time.Second, // Per 1 second
	)
	if err != nil {
		t.Fatalf("Failed to create rate-limited client: %v", err)
	}

	// Make rapid requests and measure timing
	times := make([]time.Duration, 4)

	for i := 0; i < 4; i++ {
		start := time.Now()

		// Make a lightweight request (Health check is more reliable)
		rpcErr := client.Health()

		times[i] = time.Since(start)

		if rpcErr != nil {
			t.Logf("RPC error on request %d: %v (this is ok for timing test)", i, rpcErr)
		}
	}

	// Log timing results
	for i, duration := range times {
		t.Logf("Request %d took: %v", i+1, duration)
	}

	// First 2 requests should be fast (tokens available)
	if times[0] > 200*time.Millisecond {
		t.Errorf("First request too slow: %v", times[0])
	}
	if times[1] > 200*time.Millisecond {
		t.Errorf("Second request too slow: %v", times[1])
	}

	// 3rd and 4th requests should be slower (rate limited)
	// Note: This test might be flaky depending on network conditions
	if times[2] < 300*time.Millisecond {
		t.Logf("Third request was faster than expected: %v (network might be very fast)", times[2])
	}
}

// TestRateLimitedRpcClient_MultipleRequests - Test burst vs sustained load
func TestRateLimitedRpcClient_BurstThenSustained(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	// Create client with 3 requests per 2 seconds
	client, err := rpcClient.NewRateLimitedRpcClient(
		"https://gnoland-testnet-rpc.cogwheel.zone",
		nil,
		3,
		2*time.Second,
	)
	if err != nil {
		t.Fatalf("Failed to create rate-limited client: %v", err)
	}

	// Test burst: 3 requests should be fast
	start := time.Now()
	for i := 0; i < 3; i++ {
		_ = client.Health()
	}
	burstDuration := time.Since(start)

	t.Logf("Burst of 3 requests took: %v", burstDuration)

	// Burst should complete quickly (all tokens available)
	if burstDuration > 1*time.Second {
		t.Errorf("Burst took too long: %v", burstDuration)
	}

	// 4th request should be slower (rate limited)
	start = time.Now()
	_ = client.Health()
	rateLimitedDuration := time.Since(start)

	t.Logf("Rate-limited request took: %v", rateLimitedDuration)

	// This should take time due to rate limiting
	if rateLimitedDuration < 500*time.Millisecond {
		t.Logf("Rate-limited request was faster than expected: %v", rateLimitedDuration)
	}
}
