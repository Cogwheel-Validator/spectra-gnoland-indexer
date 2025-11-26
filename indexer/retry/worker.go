package retry

import (
	"log"
	"time"
)

// RetryResult holds the result of a retry operation
//
// Sort of kinda like union or enum (depending on your programming language)
// but for the result of the retry operation
type RetryResult[T any] struct {
	Value   T
	Error   error
	Success bool
}

// GenericRetryQuery is a generic retry wrapper that can work with any query function
// It handles the retry logic and sends the result to a channel
//
// Parameters:
//   - retryAmount: the number of retry attempts, pulled from the query operator retry options
//   - pause: the number of attempts to pause after failing, pulled from the query operator retry options
//   - pauseTime: the time to pause after failing, pulled from the query operator retry options
//   - exponentialBackoff: the exponential backoff time, pulled from the query operator retry options
//   - fn: the function to retry
//   - args: the arguments to pass to the function
//
// Returns:
//   - <-chan RetryResult[T]: the channel to receive the result
func GenericRetryQuery[T any](
	retryAmount int,
	pause int,
	pauseTime time.Duration,
	exponentialBackoff time.Duration,
	fn func(args ...any) (T, error),
	args ...any,
) <-chan RetryResult[T] {
	// Buffered channel to prevent blocking
	resultChan := make(chan RetryResult[T], 1)

	go func() {
		defer close(resultChan)

		var lastErr error
		var zeroValue T

		for i := range retryAmount {
			result, err := fn(args...)

			if err == nil {
				// if the result is successful, send it to the channel
				resultChan <- RetryResult[T]{
					Value:   result,
					Error:   nil,
					Success: true,
				}
				return
			}

			// if the result is not successful, store the error
			lastErr = err
			log.Printf("Retry attempt %d failed: %v", i+1, err)

			// Don't sleep on the last retry attempt
			if i < retryAmount-1 {
				// Exponential backoff
				backoffDuration := exponentialBackoff * time.Duration(i+1)
				time.Sleep(backoffDuration)

				// Additional pause every pause amount of attempts
				// this is mostly done so the program might have
				// a better chance to recover from the error
				// for example if the RPC is down at that specific time
				// ot there is some internet problem like a short disconnection
				// default for this is 15 seconds backoff, although it can be changed
				// it might slow down the program in the long run
				// but per block chunk this is max of 1 minute
				if (i+1)%pause == 0 {
					time.Sleep(pauseTime)
				}
			}
		}

		// All retries failed
		resultChan <- RetryResult[T]{
			Value:   zeroValue,
			Error:   lastErr,
			Success: false,
		}
	}()

	return resultChan
}

// RetryWithContext is a wrapper that integrates retry logic with concurrent operations
func RetryWithContext[T any](
	retryAmount int,
	pause int,
	pauseTime time.Duration,
	exponentialBackoff time.Duration,
	fn func(args ...any) (T, error),
	onSuccess func(T),
	onFailure func(error),
	args ...any,
) {
	resultChan := GenericRetryQuery(retryAmount, pause, pauseTime, exponentialBackoff, fn, args...)

	go func() {
		result := <-resultChan
		if result.Success {
			onSuccess(result.Value)
		} else {
			if result.Error != nil {
				log.Printf("Error: All retry attempts failed: %v", result.Error)
			}
			onFailure(result.Error)
		}
	}()
}
