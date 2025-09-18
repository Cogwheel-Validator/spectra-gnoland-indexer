package main

import (
	"fmt"
	"sync"
	"time"

	rpcclient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
)

func main() {
	fmt.Println("Rate Limited RPC Client Examples")

	// Example 1: Basic usage with blocking calls
	fmt.Println("1. Basic Rate Limited RPC Client:")
	basicExample()

	// Example 2: Non-blocking calls (Try methods)
	fmt.Println("\n2. Non-blocking Rate Limited Calls:")
	nonBlockingExample()

	// Example 3: Concurrent usage demonstration
	fmt.Println("\n3. Concurrent Rate Limited Usage:")
	concurrentExample()

	// Example 4: Rate limiter status monitoring
	fmt.Println("\n4. Rate Limiter Status Monitoring:")
	statusMonitoringExample()
}

func basicExample() {
	// Create a rate-limited client: 5 requests per 3 seconds
	client, err := rpcclient.NewRateLimitedRpcClient(
		"https://gnoland-testnet-rpc.cogwheel.zone",
		nil,           // use default timeout
		5,             // 5 requests
		3*time.Second, // per 3 seconds
	)
	if err != nil {
		fmt.Printf("  Failed to create client: %v\n", err)
		return
	}
	defer client.Close()

	// Make several requests - should be rate limited
	for i := 0; i < 7; i++ {
		start := time.Now()

		err := client.Health()
		elapsed := time.Since(start)

		if err != nil {
			fmt.Printf("  Request %d: FAILED (%v) after %v\n", i+1, err, elapsed.Round(time.Millisecond))
		} else {
			fmt.Printf("  Request %d: SUCCESS after %v\n", i+1, elapsed.Round(time.Millisecond))
		}
	}
}

func nonBlockingExample() {
	// Create a rate-limited client: 3 requests per 2 seconds
	client, err := rpcclient.NewRateLimitedRpcClient(
		"https://gnoland-testnet-rpc.cogwheel.zone",
		nil,
		3,             // 3 requests
		2*time.Second, // per 2 seconds
	)
	if err != nil {
		fmt.Printf("  Failed to create client: %v\n", err)
		return
	}
	defer client.Close()

	// Try making more requests than allowed - non-blocking
	successCount := 0
	rateLimitedCount := 0

	for i := 0; i < 6; i++ {
		err, allowed := client.TryHealth()

		if !allowed {
			rateLimitedCount++
			fmt.Printf("  Request %d: RATE LIMITED (would block)\n", i+1)
		} else if err != nil {
			fmt.Printf("  Request %d: FAILED (%v)\n", i+1, err)
		} else {
			successCount++
			fmt.Printf("  Request %d: SUCCESS (non-blocking)\n", i+1)
		}
	}

	fmt.Printf("  Summary: %d successful, %d rate limited\n", successCount, rateLimitedCount)
}

func concurrentExample() {
	// Create a rate-limited client: 10 requests per 2 seconds
	client, err := rpcclient.NewRateLimitedRpcClient(
		"https://gnoland-testnet-rpc.cogwheel.zone",
		nil,
		10,            // 10 requests
		2*time.Second, // per 2 seconds
	)
	if err != nil {
		fmt.Printf("  Failed to create client: %v\n", err)
		return
	}
	defer client.Close()

	var wg sync.WaitGroup
	const numWorkers = 5
	const requestsPerWorker = 4

	// Start multiple goroutines making concurrent requests
	for worker := 0; worker < numWorkers; worker++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()

			for req := 0; req < requestsPerWorker; req++ {
				start := time.Now()

				err := client.Health()
				elapsed := time.Since(start)

				if err != nil {
					fmt.Printf("  Worker %d, Request %d: FAILED (%v) after %v\n",
						workerID, req+1, err, elapsed.Round(time.Millisecond))
				} else {
					fmt.Printf("  Worker %d, Request %d: SUCCESS after %v\n",
						workerID, req+1, elapsed.Round(time.Millisecond))
				}
			}
		}(worker)
	}

	wg.Wait()
	fmt.Printf("  All %d workers completed (%d total requests)\n",
		numWorkers, numWorkers*requestsPerWorker)
}

func statusMonitoringExample() {
	// Create a rate-limited client: 4 requests per 1 second
	client, err := rpcclient.NewRateLimitedRpcClient(
		"https://gnoland-testnet-rpc.cogwheel.zone",
		nil,
		4,             // 4 requests
		1*time.Second, // per 1 second
	)
	if err != nil {
		fmt.Printf("  Failed to create client: %v\n", err)
		return
	}
	defer client.Close()

	// Monitor status while making requests
	printStatus := func(label string) {
		status := client.GetRateLimiterStatus()
		fmt.Printf("  %s: %d/%d tokens available (Empty: %v, Full: %v)\n",
			label, status.TokensAvailable, status.Capacity, status.IsEmpty, status.IsFull)
	}

	printStatus("Initial state")

	// Use up all tokens with non-blocking calls
	for i := 0; i < 6; i++ {
		_, allowed := client.TryHealth()
		if allowed {
			fmt.Printf("  Request %d: Allowed\n", i+1)
		} else {
			fmt.Printf("  Request %d: Rate limited\n", i+1)
		}
		printStatus(fmt.Sprintf("After request %d", i+1))
	}

	// Wait a bit and check status again
	fmt.Println("  Waiting 500ms for token refill...")
	time.Sleep(500 * time.Millisecond)
	printStatus("After 500ms wait")

	fmt.Println("  Waiting another 600ms for more tokens...")
	time.Sleep(600 * time.Millisecond)
	printStatus("After total 1100ms wait")
}
