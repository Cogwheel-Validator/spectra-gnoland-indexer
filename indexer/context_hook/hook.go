package contexthook

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// NewSignalHandler is a constructor function that creates a new signal handler with
// cleanup and state dump functions
//
// Args:
//   - cleanup: the cleanup function
//   - stateDump: the state dump function
//
// Returns:
//   - *SignalHandler: the signal handler
//
// The method will not throw an error if the signal handler is not found, it will just return nil
func NewSignalHandler(cleanup func() error, stateDump func() error) *SignalHandler {
	ctx, cancel := context.WithCancel(context.Background())
	return &SignalHandler{
		ctx:       ctx,
		cancel:    cancel,
		cleanup:   cleanup,
		stateDump: stateDump,
	}
}

// Context returns the context that will be cancelled on shutdown signals
//
// Args:
//   - sh: the signal handler
//
// Returns:
//   - context.Context: the context
//
// The method will not throw an error if the context is not found, it will just return nil
func (sh *SignalHandler) Context() context.Context {
	return sh.ctx
}

// StartListening is a method that starts listening for shutdown signals
// It handles:
// - SIGINT (Ctrl+C) and SIGTERM: graceful shutdown with cleanup
// - SIGQUIT: emergency shutdown with state dump
// - SIGKILL cannot be trapped (handled by kernel directly)
//
// The method will not throw an error if the signal is not found, it will just return nil
func (sh *SignalHandler) StartListening() {
	signalChan := make(chan os.Signal, 1)

	// Register signals we can handle
	// Note: SIGKILL cannot be caught, blocked, or ignored
	signal.Notify(signalChan,
		os.Interrupt,    // SIGINT (Ctrl+C)
		syscall.SIGTERM, // SIGTERM (graceful termination)
		syscall.SIGQUIT, // SIGQUIT (emergency dump)
	)

	go func() {
		sig := <-signalChan
		log.Printf("Received signal: %v", sig)

		switch sig {
		case os.Interrupt, syscall.SIGTERM:
			log.Printf("Starting graceful shutdown...")
			sh.gracefulShutdown()
		case syscall.SIGQUIT:
			log.Printf("Emergency shutdown requested, dumping state...")
			sh.emergencyShutdown()
		}
	}()
}

// gracefulShutdown is a private method that performs
// cleanup operations and exits gracefully
//
// Returns:
//   - none
//
// The method will not throw an error if the signal is not found, it will just return nil
func (sh *SignalHandler) gracefulShutdown() {
	// Cancel the context to signal all operations to stop
	sh.cancel()

	// Give operations time to finish gracefully
	shutdownTimeout := 30 * time.Second
	log.Printf("Waiting up to %v for operations to complete...", shutdownTimeout)

	done := make(chan struct{})
	go func() {
		sh.shutdownWg.Wait()
		close(done)
	}()

	select {
	case <-done:
		log.Printf("All operations completed successfully")
	case <-time.After(shutdownTimeout):
		log.Printf("Timeout reached, forcing shutdown")
	}

	// Run cleanup function if provided
	if sh.cleanup != nil {
		log.Printf("Running cleanup operations...")
		if err := sh.cleanup(); err != nil {
			log.Printf("Error during cleanup: %v", err)
		} else {
			log.Printf("Cleanup completed successfully")
		}
	}

	log.Printf("Graceful shutdown complete")
	os.Exit(0)
}

// emergencyShutdown is a private method that dumps state and exits immediately
//
// Returns:
//   - none
//
// The method will not throw an error if the signal is not found, it will just return nil
func (sh *SignalHandler) emergencyShutdown() {
	// Cancel context immediately
	sh.cancel()

	// Dump state if function is provided
	if sh.stateDump != nil {
		log.Printf("Dumping application state...")
		if err := sh.stateDump(); err != nil {
			log.Printf("Error dumping state: %v", err)
		} else {
			log.Printf("State dump completed")
		}
	}

	log.Printf("Emergency shutdown complete")
	os.Exit(1)
}

// RegisterOperation is a method that registers an operation that should be waited for during shutdown
//
// Returns:
//   - none
//
// The method will not throw an error if the signal is not found, it will just return nil
func (sh *SignalHandler) RegisterOperation() {
	sh.shutdownWg.Add(1)
}

// OperationComplete is a method that marks an operation as complete
//
// Returns:
//   - none
//
// The method will not throw an error if the signal is not found, it will just return nil
func (sh *SignalHandler) OperationComplete() {
	sh.shutdownWg.Done()
}
