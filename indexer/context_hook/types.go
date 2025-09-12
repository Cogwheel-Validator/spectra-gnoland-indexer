package contexthook

import (
	"context"
	"sync"
)

// SignalHandler manages signal handling and graceful shutdown
type SignalHandler struct {
	ctx        context.Context
	cancel     context.CancelFunc
	cleanup    func() error
	stateDump  func() error
	shutdownWg sync.WaitGroup
}
