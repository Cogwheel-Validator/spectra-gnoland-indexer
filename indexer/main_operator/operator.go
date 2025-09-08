package mainoperator

import (
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	addressCache "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/address_cache"
	dataProcessor "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/data_processor"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/database"
	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
)

// This function is not ready to be used
// this is just a placeholder
// every part of the indexer should be initialized within the main operator
func NewMainOperator(
	db *database.TimescaleDb,
	addressCache *addressCache.AddressCache,
	rpcClient *rpcClient.RateLimitedRpcClient,
	validatorCache *addressCache.AddressCache,
	chainName string,
	dataProcessor *dataProcessor.DataProcessor) *MainOperator {
	return &MainOperator{
		db:             db,
		addressCache:   addressCache,
		rpcClient:      rpcClient,
		validatorCache: validatorCache,
		chainName:      chainName,
		dataProcessor:  dataProcessor,
	}
}

// This function processes all data using optimized concurrent execution
//
// Args:
//   - blocks: a slice of blocks
//   - transactions: a map of transactions and timestamps
//   - compressEvents: if true, compress the events
//   - fromHeight: the start height
//   - toHeight: the end height
//
// Returns:
//   - error: if processing fails
//
// The method will not throw an error if the data is not found, it will just return nil
func (mo *MainOperator) processAllConcurrently(
	blocks []*rpcClient.BlockResponse,
	transactions map[*rpcClient.TxResponse]time.Time,
	compressEvents bool,
	fromHeight uint64,
	toHeight uint64) error {

	// Phase 1: Independent concurrent operations
	var wg1 sync.WaitGroup
	var errors []error
	var errorsMutex sync.Mutex

	// Channel to signal when validator addresses are ready
	validatorAddressesDone := make(chan struct{})

	wg1.Add(3)

	// 1. Process validator addresses (populates validator cache)
	go func() {
		defer wg1.Done()
		defer close(validatorAddressesDone) // Signal completion
		log.Printf("Phase 1: Starting ProcessValidatorAddresses")
		mo.dataProcessor.ProcessValidatorAddresses(blocks, fromHeight, toHeight)
		log.Printf("Phase 1: ProcessValidatorAddresses completed")
	}()

	// 2. Process transactions (no dependencies)
	go func() {
		defer wg1.Done()
		log.Printf("Phase 1: Starting ProcessTransactions")
		mo.dataProcessor.ProcessTransactions(transactions, compressEvents, fromHeight, toHeight)
		log.Printf("Phase 1: ProcessTransactions completed")
	}()

	// 3. Process messages (uses separate address cache)
	go func() {
		defer wg1.Done()
		log.Printf("Phase 1: Starting ProcessMessages")
		if err := mo.dataProcessor.ProcessMessages(transactions, fromHeight, toHeight); err != nil {
			errorsMutex.Lock()
			errors = append(errors, fmt.Errorf("ProcessMessages failed: %w", err))
			errorsMutex.Unlock()
		}
		log.Printf("Phase 1: ProcessMessages completed")
	}()

	// Wait for Phase 1 to complete
	wg1.Wait()

	// Check for errors from Phase 1
	if len(errors) > 0 {
		var errorMessages []string
		for _, err := range errors {
			errorMessages = append(errorMessages, err.Error())
		}
		return fmt.Errorf("Phase 1 errors: %s", strings.Join(errorMessages, "; "))
	}

	// Phase 2: Operations that depend on validator addresses
	var wg2 sync.WaitGroup
	wg2.Add(2)

	// Wait for validator addresses to be ready (should already be done)
	<-validatorAddressesDone
	log.Printf("Phase 2: Validator addresses ready, starting dependent operations")

	// 4. Process blocks (needs validator address cache for proposer addresses)
	go func() {
		defer wg2.Done()
		log.Printf("Phase 2: Starting ProcessBlocks")
		mo.dataProcessor.ProcessBlocks(blocks, fromHeight, toHeight)
		log.Printf("Phase 2: ProcessBlocks completed")
	}()

	// 5. Process validator signings (needs validator address cache)
	go func() {
		defer wg2.Done()
		log.Printf("Phase 2: Starting ProcessValidatorSignings")
		mo.dataProcessor.ProcessValidatorSignings(blocks, fromHeight, toHeight)
		log.Printf("Phase 2: ProcessValidatorSignings completed")
	}()

	// Wait for Phase 2 to complete
	wg2.Wait()

	log.Printf("All processing completed successfully from %d to %d", fromHeight, toHeight)
	return nil
}
