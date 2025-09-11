package orchestrator

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/config"
	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
)

const (
	Live     = "live"
	Historic = "historic"
)

func NewOrchestrator(
	runningMode string,
	config *config.Config,
	chainName string,
	db DatabaseHeight,
	gnoRpcClient GnolandRpcClient,
	dataProcessor DataProcessor,
	queryOperator QueryOperator,
) *Orchestrator {
	if runningMode != Live && runningMode != Historic {
		panic("invalid running mode, please choose between live and historic")
	}
	return &Orchestrator{
		runningMode:   runningMode,
		config:        config,
		chainName:     chainName,
		db:            db,
		gnoRpcClient:  gnoRpcClient,
		dataProcessor: dataProcessor,
		queryOperator: queryOperator,
	}
}

func (or *Orchestrator) HistoricProcess(
	fromHeight uint64,
	toHeight uint64) {
	log.Printf("Starting historic process from %d to %d", fromHeight, toHeight)
	startTime := time.Now()

	for startHeight := fromHeight; startHeight <= toHeight; startHeight += or.config.MaxBlockChunkSize {
		chunkEndHeight := min(startHeight+or.config.MaxBlockChunkSize, toHeight)

		chunkStartTime := time.Now()
		log.Printf("Processing chunk from %d to %d", startHeight, chunkEndHeight)

		// Step 1: Get blocks concurrently
		blocks := or.queryOperator.GetFromToBlocks(startHeight, chunkEndHeight)

		if len(blocks) == 0 {
			log.Printf("No valid blocks in chunk %d-%d", startHeight, chunkEndHeight)
			continue
		}

		// Step 2: Collect all transactions from all blocks in this chunk
		allTransactions := or.collectTransactionsFromBlocks(blocks)

		log.Printf("Collected %d transactions from %d blocks in chunk", len(allTransactions), len(blocks))

		// Step 3: Process all data concurrently
		if err := or.processAllConcurrently(blocks, allTransactions, false, startHeight, chunkEndHeight); err != nil {
			log.Printf("Error processing chunk %d-%d: %v", startHeight, chunkEndHeight, err)
			continue
		}

		// Progress logging
		chunkDuration := time.Since(chunkStartTime)
		totalDuration := time.Since(startTime)
		log.Printf("Chunk %d-%d completed in %v, total time: %v",
			startHeight, chunkEndHeight, chunkDuration, totalDuration)
	}

	totalDuration := time.Since(startTime)
	log.Printf("Historic process completed from %d to %d in %v", fromHeight, toHeight, totalDuration)
}

func (or *Orchestrator) LiveProcess(ctx context.Context, skipInitialDbCheck bool) {
	log.Printf("Starting live block processing")

	var lastProcessedHeight uint64
	var err error

	// Initial setup - get starting height
	if !skipInitialDbCheck {
		lastProcessedHeight, err = or.db.GetLastBlockHeight(or.chainName)
		if err != nil {
			log.Printf("Failed to get last block height from database: %v", err)
			log.Printf("Either there are no blocks in the database or the database is not properly configured.")
			log.Printf("Use skipInitialDbCheck=true if this is expected to run from the latest chain height without previous data.")
			return
		}
		log.Printf("Retrieved last processed height from database: %d", lastProcessedHeight)
	} else {
		// Get latest block height from chain
		latestHeight, rpcErr := or.gnoRpcClient.GetLatestBlockHeight()
		if rpcErr != nil {
			log.Printf("Failed to get latest block height from chain: %v", rpcErr)
			return
		}
		lastProcessedHeight = latestHeight
		log.Printf("Starting from latest chain height: %d (skipping database check)", lastProcessedHeight)
	}

	lastProgressTime := time.Now()

	// Main processing loop
	for {
		select {
		case <-ctx.Done():
			log.Printf("Live process interrupted by context cancellation")
			return
		default:
		}

		// Get the latest block height from the chain
		latestHeight, rpcErr := or.gnoRpcClient.GetLatestBlockHeight()
		if rpcErr != nil {
			log.Printf("Error fetching latest block height: %v", rpcErr)
			time.Sleep(time.Duration(or.config.LivePooling) * time.Second)
			continue
		}

		blocksBehind := int64(latestHeight) - int64(lastProcessedHeight)

		// If caught up, wait and continue
		if blocksBehind <= 0 {
			log.Printf("Caught up to height %d. Waiting %d seconds...", latestHeight, or.config.LivePooling)
			time.Sleep(time.Duration(or.config.LivePooling) * time.Second)
			continue
		}

		// Adjust chunk size based on how far behind we are
		currentChunkSize := min(uint64(blocksBehind), or.config.MaxBlockChunkSize)

		chunkStart := lastProcessedHeight + 1
		chunkEnd := min(chunkStart+currentChunkSize-1, latestHeight)

		log.Printf("Processing live chunk %d-%d (behind by %d blocks)", chunkStart, chunkEnd, blocksBehind)

		// Process this chunk
		err = or.processLiveChunk(chunkStart, chunkEnd)
		if err != nil {
			log.Printf("Error processing live chunk %d-%d: %v", chunkStart, chunkEnd, err)
			time.Sleep(time.Duration(or.config.LivePooling) * time.Second)
			continue
		}

		// Update progress
		lastProcessedHeight = chunkEnd
		or.updateProgressMetrics(chunkStart, chunkEnd, blocksBehind, &lastProgressTime)

		// Small delay to prevent overwhelming the API
		time.Sleep(50 * time.Millisecond)
	}
}

// processLiveChunk processes a single chunk of blocks for live processing
func (or *Orchestrator) processLiveChunk(chunkStart, chunkEnd uint64) error {
	chunkStartTime := time.Now()

	// Step 1: Get blocks concurrently
	blocks := or.queryOperator.GetFromToBlocks(chunkStart, chunkEnd)

	if len(blocks) == 0 {
		log.Printf("No valid blocks in live chunk %d-%d", chunkStart, chunkEnd)
		return nil
	}

	// Step 2: Collect all transactions from all blocks in this chunk
	allTransactions := or.collectTransactionsFromBlocks(blocks)

	log.Printf("Collected %d transactions from %d blocks in live chunk", len(allTransactions), len(blocks))

	// Step 3: Process all data concurrently
	if err := or.processAllConcurrently(blocks, allTransactions, false, chunkStart, chunkEnd); err != nil {
		return fmt.Errorf("failed to process live chunk %d-%d: %w", chunkStart, chunkEnd, err)
	}

	chunkDuration := time.Since(chunkStartTime)
	log.Printf("Live chunk %d-%d completed in %v", chunkStart, chunkEnd, chunkDuration)

	return nil
}

// updateProgressMetrics updates and logs progress metrics for live processing
func (or *Orchestrator) updateProgressMetrics(chunkStart, chunkEnd uint64, blocksBehind int64, lastProgressTime *time.Time) {
	now := time.Now()
	timeSinceLastProgress := now.Sub(*lastProgressTime)

	// Log progress every 30 seconds or significant milestones
	if timeSinceLastProgress >= 30*time.Second || blocksBehind <= 10 {
		blocksProcessed := chunkEnd - chunkStart + 1
		log.Printf("Live progress: processed %d blocks (%d-%d), %d blocks behind, last update: %v ago",
			blocksProcessed, chunkStart, chunkEnd, blocksBehind, timeSinceLastProgress.Round(time.Second))
		*lastProgressTime = now
	}
}

// collectTransactionsFromBlocks extracts all transactions from blocks and queries them concurrently
// This mimics the Python _process_historical_block_chunk behavior
func (or *Orchestrator) collectTransactionsFromBlocks(blocks []*rpcClient.BlockResponse) map[*rpcClient.TxResponse]time.Time {
	// Collect all transaction hashes from all blocks
	var allTxHashes []string
	blockTimestamps := make(map[string]time.Time) // txHash -> block timestamp

	for _, block := range blocks {
		if block == nil || block.Result.Block.Data.Txs == nil {
			continue
		}

		for _, txHash := range *block.Result.Block.Data.Txs {
			allTxHashes = append(allTxHashes, txHash)
			blockTimestamps[txHash] = block.Result.Block.Header.Time
		}
	}

	if len(allTxHashes) == 0 {
		return make(map[*rpcClient.TxResponse]time.Time)
	}

	log.Printf("Fetching %d transactions concurrently", len(allTxHashes))

	// Query all transactions concurrently
	transactions := or.queryOperator.GetTransactions(allTxHashes)

	// Build map of transactions with their timestamps
	txMap := make(map[*rpcClient.TxResponse]time.Time)
	for _, tx := range transactions {
		if tx != nil {
			if timestamp, exists := blockTimestamps[tx.Result.Hash]; exists {
				txMap[tx] = timestamp
			}
		}
	}

	log.Printf("Successfully collected %d valid transactions", len(txMap))
	return txMap
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
func (or *Orchestrator) processAllConcurrently(
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
		or.dataProcessor.ProcessValidatorAddresses(blocks, fromHeight, toHeight)
		log.Printf("Phase 1: ProcessValidatorAddresses completed")
	}()

	// 2. Process transactions (no dependencies)
	go func() {
		defer wg1.Done()
		log.Printf("Phase 1: Starting ProcessTransactions")
		or.dataProcessor.ProcessTransactions(transactions, compressEvents, fromHeight, toHeight)
		log.Printf("Phase 1: ProcessTransactions completed")
	}()

	// 3. Process messages (uses separate address cache)
	go func() {
		defer wg1.Done()
		log.Printf("Phase 1: Starting ProcessMessages")
		if err := or.dataProcessor.ProcessMessages(transactions, fromHeight, toHeight); err != nil {
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
		or.dataProcessor.ProcessBlocks(blocks, fromHeight, toHeight)
		log.Printf("Phase 2: ProcessBlocks completed")
	}()

	// 5. Process validator signings (needs validator address cache)
	go func() {
		defer wg2.Done()
		log.Printf("Phase 2: Starting ProcessValidatorSignings")
		or.dataProcessor.ProcessValidatorSignings(blocks, fromHeight, toHeight)
		log.Printf("Phase 2: ProcessValidatorSignings completed")
	}()

	// Wait for Phase 2 to complete
	wg2.Wait()

	log.Printf("All processing completed successfully from %d to %d", fromHeight, toHeight)
	return nil
}
