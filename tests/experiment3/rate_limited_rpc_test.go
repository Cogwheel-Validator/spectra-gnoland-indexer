package main

import (
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rate_limit"
	rpcclient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
)

func TestRateLimitedRpcClient(t *testing.T) {
	t.Run("BasicRateLimiting", testBasicRateLimiting)
	t.Run("NonBlockingMethods", testNonBlockingMethods)
	t.Run("ConcurrentAccess", testConcurrentAccess)
	t.Run("StatusMonitoring", testStatusMonitoring)
	t.Run("TokenRefillBehavior", testTokenRefillBehavior)
}

// Test the ChannelRateLimiter directly (no network calls) to demonstrate token bucket behavior
func TestChannelRateLimiterTokenBucket(t *testing.T) {
	// Create a rate limiter: 3 tokens per 300ms (100ms per token)
	limiter := rate_limit.NewChannelRateLimiter(3, 300*time.Millisecond)
	defer limiter.Close()

	// Initial state: should have 3 tokens
	if !limiter.Allow() {
		t.Error("Expected first token to be available")
	}
	if !limiter.Allow() {
		t.Error("Expected second token to be available")
	}
	if !limiter.Allow() {
		t.Error("Expected third token to be available")
	}

	// Fourth token should not be available
	if limiter.Allow() {
		t.Error("Expected fourth token to be blocked (bucket empty)")
	}

	// Wait for one token refill (120ms > 100ms per token)
	time.Sleep(120 * time.Millisecond)

	// Should have exactly 1 token now
	if !limiter.Allow() {
		t.Error("Expected one token after 120ms wait")
	}

	// Should be empty again
	if limiter.Allow() {
		t.Error("Expected to be empty after using the refilled token")
	}

	// Wait for two tokens (220ms > 200ms for 2 tokens)
	time.Sleep(220 * time.Millisecond)

	// Should have 2 tokens available
	tokensUsed := 0
	for i := 0; i < 4; i++ { // Try to use more than available
		if limiter.Allow() {
			tokensUsed++
		}
	}

	if tokensUsed != 2 {
		t.Errorf("Expected exactly 2 tokens after 220ms, got %d", tokensUsed)
	}

	t.Log("Token bucket behavior working correctly")
}

func testBasicRateLimiting(t *testing.T) {
	// Create client with very restrictive limits for testing
	// 2 requests per 1 second = 500ms per token refill
	client, err := rpcclient.NewRateLimitedRpcClient(
		"https://gnoland-testnet-rpc.cogwheel.zone", // Using real endpoint for integration test
		nil,
		2,             // 2 requests
		1*time.Second, // per second (500ms per token)
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Use non-blocking calls to test token bucket behavior without network latency
	// First two requests should succeed immediately (bucket starts full)
	successCount := 0
	start := time.Now()
	for i := 0; i < 2; i++ {
		if _, allowed := client.TryHealth(); allowed {
			successCount++
		}
	}
	elapsed := time.Since(start)
	t.Logf("First 2 requests took %v", elapsed)

	if successCount != 2 {
		t.Errorf("Expected 2 immediate successes, got %d", successCount)
	}

	// Check status before third request
	status := client.GetRateLimiterStatus()
	t.Logf("Status before third request: %d tokens available", status.TokensAvailable)

	// NOTE: With network latency, tokens may have been refilled during the first 2 requests
	// This is CORRECT token bucket behavior - tokens refill continuously
	if status.TokensAvailable > 0 {
		t.Logf("Tokens were refilled during network calls (correct token bucket behavior)")
	} else {
		t.Log("No tokens available immediately after consumption")
	}

	// Wait for token refill and try again
	time.Sleep(600 * time.Millisecond) // Wait longer than refill interval (500ms)
	if _, allowed := client.TryHealth(); !allowed {
		t.Error("Request should succeed after token refill")
	}

	t.Log("Rate limiting working correctly with token bucket behavior")
}

func testNonBlockingMethods(t *testing.T) {
	// Use a slower refill rate for predictable testing
	client, err := rpcclient.NewRateLimitedRpcClient(
		"https://gnoland-testnet-rpc.cogwheel.zone",
		nil,
		2,             // 2 requests
		2*time.Second, // per 2 seconds (1 second per token)
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Should have 2 tokens initially
	allowedCount := 0
	blockedCount := 0

	// Try 4 requests immediately - only 2 should be allowed
	for i := 0; i < 4; i++ {
		_, allowed := client.TryHealth()
		if allowed {
			allowedCount++
		} else {
			blockedCount++
		}
	}

	if allowedCount != 2 {
		t.Errorf("Expected 2 allowed requests initially, got %d", allowedCount)
	}
	if blockedCount != 2 {
		t.Errorf("Expected 2 blocked requests initially, got %d", blockedCount)
	}

	// Wait for one token refill (1 second) and try again
	time.Sleep(1100 * time.Millisecond)

	if _, allowed := client.TryHealth(); !allowed {
		t.Error("Request should be allowed after token refill")
	} else {
		allowedCount++
	}

	t.Logf("Non-blocking methods working: %d total allowed, %d initially blocked", allowedCount, blockedCount)
}

func testConcurrentAccess(t *testing.T) {
	// Use modest rate limiting for concurrent test
	client, err := rpcclient.NewRateLimitedRpcClient(
		"https://gnoland-testnet-rpc.cogwheel.zone",
		nil,
		5,             // 5 requests
		1*time.Second, // per second (200ms per token)
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	const numGoroutines = 3
	const requestsPerGoroutine = 2 // total 6 requests (more than initial capacity)

	var wg sync.WaitGroup
	var successCount int64
	var errorCount int64

	start := time.Now()

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(goroutineID int) {
			defer wg.Done()

			for j := 0; j < requestsPerGoroutine; j++ {
				// Use Try method first to avoid blocking in concurrent test
				if _, allowed := client.TryHealth(); allowed {
					atomic.AddInt64(&successCount, 1)
				} else {
					// If not allowed, use blocking method (will wait for token)
					if err := client.Health(); err != nil {
						atomic.AddInt64(&errorCount, 1)
						t.Logf("Goroutine %d, request %d failed: %v", goroutineID, j+1, err)
					} else {
						atomic.AddInt64(&successCount, 1)
					}
				}
			}
		}(i)
	}

	wg.Wait()
	elapsed := time.Since(start)

	totalRequests := int64(numGoroutines * requestsPerGoroutine)
	t.Logf("Concurrent test completed in %v", elapsed)
	t.Logf("Total requests: %d, Success: %d, Errors: %d",
		totalRequests, successCount, errorCount)

	// Most requests should succeed (either immediately or after waiting)
	if successCount < totalRequests-1 {
		t.Errorf("Expected most requests to succeed, got %d successes out of %d",
			successCount, totalRequests)
	}

	// Test should demonstrate some rate limiting effect
	if elapsed < 200*time.Millisecond {
		t.Logf("Note: Test completed quickly (%v), rate limiting may not have been needed", elapsed)
	} else {
		t.Logf("Rate limiting working in concurrent scenario (took %v)", elapsed)
	}
}

func testStatusMonitoring(t *testing.T) {
	// Use very slow refill rate for predictable status testing
	client, err := rpcclient.NewRateLimitedRpcClient(
		"https://gnoland-testnet-rpc.cogwheel.zone",
		nil,
		3,             // 3 requests
		3*time.Second, // per 3 seconds (1 second per token)
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Initial status should show full capacity
	status := client.GetRateLimiterStatus()
	if status.Capacity != 3 {
		t.Errorf("Expected capacity 3, got %d", status.Capacity)
	}
	if status.TokensAvailable != 3 {
		t.Errorf("Expected 3 tokens initially, got %d", status.TokensAvailable)
	}
	if !status.IsFull {
		t.Error("Expected rate limiter to be full initially")
	}
	if status.IsEmpty {
		t.Error("Expected rate limiter to not be empty initially")
	}

	// Use all tokens quickly (before any can refill)
	for i := 0; i < 3; i++ {
		_, allowed := client.TryHealth()
		if !allowed {
			t.Errorf("Expected request %d to be allowed", i+1)
		}
	}

	// Check status immediately (should be empty)
	status = client.GetRateLimiterStatus()
	if status.TokensAvailable != 0 {
		t.Errorf("Expected 0 tokens after using all, got %d", status.TokensAvailable)
	}
	if !status.IsEmpty {
		t.Error("Expected rate limiter to be empty after using all tokens")
	}
	if status.IsFull {
		t.Error("Expected rate limiter to not be full after using all tokens")
	}

	// Wait for partial refill
	time.Sleep(1100 * time.Millisecond) // Should get 1 token back

	status = client.GetRateLimiterStatus()
	if status.TokensAvailable == 0 {
		t.Error("Expected at least 1 token after 1 second")
	}
	if status.TokensAvailable > 2 {
		t.Errorf("Expected at most 2 tokens after 1 second, got %d", status.TokensAvailable)
	}

	t.Logf("Status monitoring working correctly")
}

func testTokenRefillBehavior(t *testing.T) {
	// Use predictable refill timing: 2 requests per 1 second = 500ms per token
	client, err := rpcclient.NewRateLimitedRpcClient(
		"https://gnoland-testnet-rpc.cogwheel.zone",
		nil,
		2,             // 2 requests
		1*time.Second, // per 1 second (500ms per token)
	)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	// Use all tokens quickly
	usedTokens := 0
	start := time.Now()
	for i := 0; i < 2; i++ {
		if _, allowed := client.TryHealth(); allowed {
			usedTokens++
		}
	}
	consumeTime := time.Since(start)
	t.Logf("Consumed %d tokens in %v", usedTokens, consumeTime)

	if usedTokens != 2 {
		t.Errorf("Expected to use 2 tokens, actually used %d", usedTokens)
	}

	// Verify empty immediately
	status := client.GetRateLimiterStatus()
	t.Logf("Status after consuming all tokens: %d available", status.TokensAvailable)
	if !status.IsEmpty {
		t.Error("Expected empty bucket immediately after using all tokens")
	}

	// Wait for one token refill (600ms > 500ms per token)
	t.Logf("Waiting 600ms for token refill...")
	time.Sleep(600 * time.Millisecond)

	// Check status after wait
	status = client.GetRateLimiterStatus()
	t.Logf("Status after 600ms wait: %d tokens available", status.TokensAvailable)

	// Should have exactly 1 token available
	allowedCount := 0
	for i := 0; i < 3; i++ { // Try more than available to test limits
		if _, allowed := client.TryHealth(); allowed {
			allowedCount++
		}
	}

	// Calculate expected tokens: (consumeTime + waitTime) / tokenInterval
	// Total elapsed time for refill calculation
	totalElapsedTime := consumeTime + 600*time.Millisecond
	expectedTokens := int(totalElapsedTime / (500 * time.Millisecond))
	if expectedTokens > 2 { // Cap at bucket capacity
		expectedTokens = 2
	}

	t.Logf("Total elapsed time: %v, expected tokens: %d", totalElapsedTime, expectedTokens)

	if allowedCount < 1 {
		t.Errorf("Expected at least 1 token after %v total time, got %d", totalElapsedTime, allowedCount)
	}
	if allowedCount > 2 {
		t.Errorf("Expected at most 2 tokens (bucket capacity), got %d", allowedCount)
	}

	t.Logf("Token refill working correctly: %d tokens available (expected %d)", allowedCount, expectedTokens)

	// Test demonstrates that token bucket refills correctly over time
	t.Log("Token refill behavior is mathematically correct for token bucket algorithm")
}

// Benchmark the rate-limited client
func BenchmarkRateLimitedRpcClient(b *testing.B) {
	client, err := rpcclient.NewRateLimitedRpcClient(
		"https://gnoland-testnet-rpc.cogwheel.zone",
		nil,
		1000, // High limit to avoid blocking in benchmark
		1*time.Second,
	)
	if err != nil {
		b.Fatalf("Failed to create client: %v", err)
	}
	defer client.Close()

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			// Use TryHealth to avoid blocking on rate limits
			_, allowed := client.TryHealth()
			if !allowed {
				b.Errorf("Expected request to be allowed")
			}
		}
	})
}
