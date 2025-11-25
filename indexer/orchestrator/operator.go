package orchestrator

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/config"
	dataprocessor "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/data_processor"
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
		runningMode:             runningMode,
		config:                  config,
		chainName:               chainName,
		db:                      db,
		gnoRpcClient:            gnoRpcClient,
		dataProcessor:           dataProcessor,
		queryOperator:           queryOperator,
		isProcessing:            false,
		currentProcessingHeight: 0,
	}
}

func (or *Orchestrator) HistoricProcess(
	fromHeight uint64,
	toHeight uint64) {
	log.Printf("Starting historic process from %d to %d", fromHeight, toHeight)
	startTime := time.Now()

	// Track processing state
	or.isProcessing = true
	or.currentProcessingHeight = fromHeight
	defer func() {
		or.isProcessing = false
		log.Printf("Historic processing completed at height %d", or.currentProcessingHeight)
	}()

	for startHeight := fromHeight; startHeight <= toHeight; {
		chunkEndHeight := min(startHeight+or.config.MaxBlockChunkSize-1, toHeight)

		log.Printf("Processing chunk from %d to %d", startHeight, chunkEndHeight)

		// Update current processing height
		or.currentProcessingHeight = startHeight

		// Process the chunk
		err := or.processChunk(startHeight, chunkEndHeight)
		if err != nil {
			log.Printf("Error processing chunk %d-%d: %v", startHeight, chunkEndHeight, err)

		}

		// Always advance to next chunk, regardless of whether blocks were found
		// Update processing height to end of chunk
		or.currentProcessingHeight = chunkEndHeight
		startHeight = chunkEndHeight + 1
	}

	totalDuration := time.Since(startTime)
	log.Printf("Historic process completed from %d to %d in %v", fromHeight, toHeight, totalDuration)
}

func (or *Orchestrator) LiveProcess(ctx context.Context, skipInitialDbCheck bool) {
	log.Printf("Starting live block processing")

	var lastProcessedHeight uint64
	var err error

	// Track our current processing state for potential cleanup
	or.isProcessing = true
	or.currentProcessingHeight = 0
	defer func() {
		or.isProcessing = false
		log.Printf("Live processing stopped at height %d", or.currentProcessingHeight)
	}()

	// Initial setup - get starting height
	if !skipInitialDbCheck {
		lastProcessedHeight, err = or.db.GetLastBlockHeight(or.chainName)
		if err != nil {
			log.Printf("Failed to get last block height from database: %v", err)
			log.Printf("Either there are no blocks in the database or the database is not properly configured.")
			log.Printf("Use skipInitialDbCheck=true if this is expected to run from the latest chain height without previous data.")
			log.Printf("Starting from height 1")
			lastProcessedHeight = 0
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

	or.currentProcessingHeight = lastProcessedHeight
	lastProgressTime := time.Now()

	// Main processing loop
	for {
		select {
		case <-ctx.Done():
			log.Printf("Live process interrupted by context cancellation")
			or.saveProcessingState(lastProcessedHeight, "live_interrupted")
			return
		default:
		}

		// Get the latest block height from the chain
		latestHeight, rpcErr := or.gnoRpcClient.GetLatestBlockHeight()
		if rpcErr != nil {
			log.Printf("Error fetching latest block height: %v", rpcErr)
			time.Sleep(or.config.LivePooling)
			continue
		}

		blocksBehind := latestHeight - lastProcessedHeight

		// If caught up, wait and continue
		if blocksBehind <= 0 {
			log.Printf("Caught up to height %d. Waiting %d seconds...", latestHeight, or.config.LivePooling/time.Second)
			time.Sleep(or.config.LivePooling)
			continue
		}

		// Adjust chunk size based on how far behind we are
		currentChunkSize := min(blocksBehind, or.config.MaxBlockChunkSize)

		chunkStart := lastProcessedHeight + 1
		chunkEnd := min(chunkStart+currentChunkSize-1, latestHeight)

		log.Printf("Processing live chunk %d-%d (behind by %d blocks)", chunkStart, chunkEnd, blocksBehind)

		// Update current processing height
		or.currentProcessingHeight = chunkStart

		// Process this chunk
		err = or.processChunk(chunkStart, chunkEnd)
		if err != nil {
			log.Printf("Error processing live chunk %d-%d: %v", chunkStart, chunkEnd, err)
			time.Sleep(or.config.LivePooling)
			continue
		}

		// Update progress
		lastProcessedHeight = chunkEnd
		or.currentProcessingHeight = chunkEnd
		or.updateProgressMetrics(chunkStart, chunkEnd, blocksBehind, &lastProgressTime)

		// Small delay to prevent overwhelming the API
		time.Sleep(50 * time.Millisecond)
	}
}

// processChunk processes a single chunk of blocks for live processing
func (or *Orchestrator) processChunk(chunkStart, chunkEnd uint64) error {
	chunkStartTime := time.Now()

	// Step 1: Get blocks concurrently
	var wg sync.WaitGroup
	wg.Add(2)

	var blocks []*rpcClient.BlockResponse
	var commits []*rpcClient.CommitResponse

	go func() {
		defer wg.Done()
		blocks = or.queryOperator.GetFromToBlocks(chunkStart, chunkEnd)
	}()
	go func() {
		defer wg.Done()
		commits = or.queryOperator.GetFromToCommits(chunkStart, chunkEnd)
	}()

	wg.Wait()

	if len(blocks) == 0 && len(commits) == 0 {
		log.Printf("No valid blocks in live chunk %d-%d", chunkStart, chunkEnd)
		return nil
	}

	// Step 2: Collect all transactions from all blocks in this chunk
	allTransactions := or.collectTransactionsFromBlocks(blocks)

	log.Printf("Collected %d transactions from %d blocks in live chunk", len(allTransactions), len(blocks))

	// Step 3: Process all data concurrently
	if err := or.processAll(blocks, commits, allTransactions, false, chunkStart, chunkEnd); err != nil {
		return fmt.Errorf("failed to process live chunk %d-%d: %w", chunkStart, chunkEnd, err)
	}

	chunkDuration := time.Since(chunkStartTime)
	log.Printf("Chunk %d-%d completed in %v", chunkStart, chunkEnd, chunkDuration)

	return nil
}

// updateProgressMetrics updates and logs progress metrics for live processing
func (or *Orchestrator) updateProgressMetrics(
	chunkStart, chunkEnd, blocksBehind uint64,
	lastProgressTime *time.Time,
) {
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

/*
	collectTransactionsFromBlocks extracts all transactions from blocks and queries them concurrently

Parameters:
  - blocks: a slice of blocks

Returns:
  - a slice of transactions

The method will not throw an error if the transactions are not found, it will just return an empty slice.
*/
func (or *Orchestrator) collectTransactionsFromBlocks(blocks []*rpcClient.BlockResponse) []dataprocessor.TrasnactionsData {
	// Collect all transaction hashes from all blocks
	var allTxHashes []string
	blockTxData := make([]struct {
		txHash      string
		blockHeight uint64
		timestamp   time.Time
	}, 0)

	for _, block := range blocks {
		if block == nil {
			continue
		}

		// this should not be nil but just in case
		// also we need to collect all of the tx hashes, decode the base64 to raw bytes then to sha256
		// and then sha256 to base64
		txHashes := block.GetTxHashes()
		if txHashes == nil {
			continue
		}
		for _, txHash := range txHashes {
			txHashBytes, err := base64.StdEncoding.DecodeString(txHash)
			if err != nil {
				continue
			}
			txHashSha256 := sha256.Sum256(txHashBytes)
			txHashFinal := base64.StdEncoding.EncodeToString(txHashSha256[:])
			blockHeight, err := block.GetHeight()
			if err != nil {
				log.Printf("Failed to get block height: %v", err)
				continue
			}
			allTxHashes = append(allTxHashes, txHashFinal)
			blockTxData = append(blockTxData, struct {
				txHash      string
				blockHeight uint64
				timestamp   time.Time
			}{
				txHash:      txHashFinal,
				blockHeight: blockHeight,
				timestamp:   block.GetTimestamp(),
			})
		}
	}

	if len(allTxHashes) == 0 {
		return make([]dataprocessor.TrasnactionsData, 0)
	}

	log.Printf("Fetching %d transactions concurrently", len(allTxHashes))

	// Query all transactions concurrently
	transactions := or.queryOperator.GetTransactions(allTxHashes)

	// Build map of transactions with their timestamps
	txData := make([]dataprocessor.TrasnactionsData, 0)
	for _, tx := range transactions {
		if tx != nil {
			for _, blockTx := range blockTxData {
				if blockTx.txHash == tx.GetHash() {
					txData = append(txData, dataprocessor.TrasnactionsData{
						Response:    tx,
						Timestamp:   blockTx.timestamp,
						BlockHeight: blockTx.blockHeight,
					})
				}
			}
		}
	}

	log.Printf("Successfully collected %d valid transactions", len(txData))
	return txData
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
func (or *Orchestrator) processAll(
	blocks []*rpcClient.BlockResponse,
	commits []*rpcClient.CommitResponse,
	transactions []dataprocessor.TrasnactionsData,
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
		return fmt.Errorf("phase 1 errors: %s", strings.Join(errorMessages, "; "))
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
		or.dataProcessor.ProcessValidatorSignings(commits, fromHeight, toHeight)
		log.Printf("Phase 2: ProcessValidatorSignings completed")
	}()

	// Wait for Phase 2 to complete
	wg2.Wait()

	log.Printf("All processing completed successfully from %d to %d", fromHeight, toHeight)
	return nil
}

// saveProcessingState is a private method that saves
// the current processing state to a file
//
// Args:
//   - height: the height of the processing state
//   - reason: the reason for the processing state
//
// Returns:
//   - none
//
// The method will not throw an error if the processing state is not found, it will just return nil
func (or *Orchestrator) saveProcessingState(height uint64, reason string) {
	state := ProcessingState{
		ChainName:               or.chainName,
		RunningMode:             or.runningMode,
		IsProcessing:            or.isProcessing,
		CurrentProcessingHeight: height,
		// it would be hard to read timestamp from the height since the program might not have the
		// data so use current time
		Timestamp: time.Now(),
		Reason:    reason,
	}

	// Create state directory if it doesn't exist
	stateDir := "state_dumps"
	if err := os.MkdirAll(stateDir, 0755); err != nil {
		log.Printf("Failed to create state directory: %v", err)
		return
	}

	// Create filename with timestamp
	filename := fmt.Sprintf("processing_state_%s_%d.json",
		time.Now().Format("20060102_150405"), height)
	filepath := filepath.Join(stateDir, filename)

	// Marshal to JSON
	data, err := json.MarshalIndent(state, "", "  ")
	if err != nil {
		log.Printf("Failed to marshal processing state: %v", err)
		return
	}

	// Write to file
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		log.Printf("Failed to write processing state: %v", err)
		return
	}

	log.Printf("Processing state saved to %s", filepath)
}

// Cleanup performs cleanup operations for the orchestrator
//
// Returns:
//   - error: if cleanup fails
func (or *Orchestrator) Cleanup() error {
	log.Printf("Starting orchestrator cleanup...")

	// Save current state before cleanup
	or.saveProcessingState(or.currentProcessingHeight, "cleanup_requested")
	log.Printf("Orchestrator cleanup completed - state saved successfully")

	return nil
}

// DumpState creates an emergency state dump with current processing information
func (or *Orchestrator) DumpState() error {
	log.Printf("Creating emergency state dump...")

	// Save processing state
	or.saveProcessingState(or.currentProcessingHeight, "emergency_dump")

	// Create additional diagnostic information
	diagnostics := map[string]interface{}{
		"chain_name":                or.chainName,
		"running_mode":              or.runningMode,
		"is_processing":             or.isProcessing,
		"current_processing_height": or.currentProcessingHeight,
		"config": map[string]interface{}{
			"max_block_chunk_size": or.config.MaxBlockChunkSize,
			"live_pooling":         or.config.LivePooling,
			"rpc_url":              or.config.RpcUrl,
		},
		// it would be hard to read timestamp from the height since the program might not have the data so use
		// current time
		"timestamp":   time.Now(),
		"dump_reason": "emergency_shutdown",
	}

	// Create diagnostics directory if it doesn't exist
	diagDir := "diagnostics"
	if err := os.MkdirAll(diagDir, 0755); err != nil {
		return fmt.Errorf("failed to create diagnostics directory: %w", err)
	}

	// Create filename with timestamp
	filename := fmt.Sprintf("emergency_dump_%s.json", time.Now().Format("20060102_150405"))
	filepath := filepath.Join(diagDir, filename)

	// Marshal to JSON
	data, err := json.MarshalIndent(diagnostics, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal diagnostics: %w", err)
	}

	// Write to file
	if err := os.WriteFile(filepath, data, 0644); err != nil {
		return fmt.Errorf("failed to write diagnostics: %w", err)
	}

	log.Printf("Emergency state dump saved to %s", filepath)
	return nil
}
