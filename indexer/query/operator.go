package query

import (
	"log"
	"sync"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/retry"
	rc "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
)

var (
	defaultRetryAmount        = 6
	defaultPause              = 3
	defaultPauseTime          = 15 * time.Second
	defaultExponentialBackoff = 2 * time.Second
)

// NewQueryOperator creates a new query operator
func NewQueryOperator(
	rpcClient RpcClient,
	retryAmount *int,
	pause *int,
	pauseTime *time.Duration,
	exponentialBackoff *time.Duration,
) *QueryOperator {
	if retryAmount == nil || *retryAmount == 0 {
		retryAmount = &defaultRetryAmount
	}
	if pause == nil || *pause == 0 {
		pause = &defaultPause
	}
	if pauseTime == nil || *pauseTime == 0 {
		pauseTime = &defaultPauseTime
	}
	if exponentialBackoff == nil || *exponentialBackoff == 0 {
		exponentialBackoff = &defaultExponentialBackoff
	}
	return &QueryOperator{
		rpcClient:          rpcClient,
		retryAmount:        *retryAmount,
		pause:              *pause,
		pauseTime:          *pauseTime,
		exponentialBackoff: *exponentialBackoff,
	}
}

// A swarm method to get blocks from a to b chain height inclusive
// This is a fan out method that lauches async workers for each block and wait to get the resaults
// The order of the blocks is not guaranteed but it shouldn't matter because at the end of the process
// the indexer should store them all together as one huge slice of blocks, so the order is not important
// the speed is what matters here.
//
// Args:
//   - fromHeight: the start height
//   - toHeight: the end height
//
// Returns:
//   - []*rpcClient.BlockResponse: returns the slice of block responses
//
// The method will not throw an error if the block is missing, not found or there is some query error,
// it will just return nil for the block.
//
// Example:
//
//	var blocks []*rpcClient.BlockResponse
//	blocks = q.GetFromToBlocks(1, 50)
//	for _, block := range blocks {
//		fmt.Println(block.Height)
//	}
//
// The method will not throw an error if the block is missing, not found or there is some query error,
// it will just return nil for the block.
//
// Example:
//
//	var blocks []*rpcClient.BlockResponse
//	blocks = q.GetFromToBlocks(1, 50)
//	for _, block := range blocks {
//		fmt.Println(block.Height)
//	}
func (q *QueryOperator) GetFromToBlocks(fromHeight uint64, toHeight uint64) []*rc.BlockResponse {
	diff := toHeight - fromHeight + 1 // example from 1 to 50 means 50 blocks so +1 is needed because 100-51+1=50
	if diff < 1 {
		return nil
	}

	// Use buffered channel for speed
	blockChan := make(chan *rc.BlockResponse, diff)
	wg := sync.WaitGroup{}
	wg.Add(int(diff))

	// Launch goroutines to get the blocks
	for height := fromHeight; height <= toHeight; height++ {
		go func(height uint64) {
			defer wg.Done()
			block, err := q.rpcClient.GetBlock(height)
			if err != nil {
				// Use retry mechanism with callback pattern
				retry.RetryWithContext(
					q.retryAmount,
					q.pause,
					q.pauseTime,
					q.exponentialBackoff,
					func(args ...any) (*rc.BlockResponse, error) {
						h := args[0].(uint64)
						result, rpcErr := q.rpcClient.GetBlock(h)
						if rpcErr != nil {
							return nil, rpcErr
						}
						return result, nil
					},
					func(result *rc.BlockResponse) {
						blockChan <- result
					},
					func(retryErr error) {
						log.Printf("failed to get block %d after retries: %v", height, retryErr)
						blockChan <- nil
					},
					height,
				)
				return
			}
			blockChan <- block
		}(height)
	}

	// Close channel when all goroutines finish to avoid deadlock
	go func() {
		wg.Wait()
		close(blockChan)
	}()

	// Collect results from the channel
	blocks := make([]*rc.BlockResponse, 0, diff)
	for block := range blockChan {
		blocks = append(blocks, block)
	}

	return blocks
}

// A swarm method to get transactions from a slice of tx hashes
// This is a fan out method that lauches async workers for each tx and wait to get the resaults
// The order of the transactions is not guaranteed but it shouldn't matter because at the end of the process
// the indexer should store them all together as one huge slice of transactions, so the order is not important
// the speed is what matters here.
//
// Args:
//   - txs: a slice of tx hashes
//
// Returns:
//   - []*rpcClient.TxResponse: returns the slice of transaction responses
//
// The method will not throw an error if the transaction is missing, not found or there is some query error,
// it will just return nil for the transaction.
//
// Example:
//
//	var transactions []*rpcClient.TxResponse
//	transactions = q.GetTransactions([]string{"tx_hash_1", "tx_hash_2", "tx_hash_3"})
//	for _, transaction := range transactions {
//		fmt.Println(transaction.Hash)
//	}
func (q *QueryOperator) GetTransactions(txs []string) []*rc.TxResponse {
	nTxs := len(txs)

	if nTxs < 1 {
		return nil
	}

	// Set up the channel and the wait group
	txChan := make(chan *rc.TxResponse, nTxs)
	wg := sync.WaitGroup{}
	wg.Add(nTxs)

	// Launch goroutines to get the transactions
	for _, tx := range txs {
		go func(tx string) {
			defer wg.Done()
			txResponse, err := q.rpcClient.GetTx(tx)
			if err != nil {
				// Use retry mechanism with callback pattern
				retry.RetryWithContext(
					q.retryAmount,
					q.pause,
					q.pauseTime,
					q.exponentialBackoff,
					func(args ...any) (*rc.TxResponse, error) {
						txHash := args[0].(string)
						result, rpcErr := q.rpcClient.GetTx(txHash)
						if rpcErr != nil {
							return nil, rpcErr
						}
						return result, nil
					},
					func(result *rc.TxResponse) {
						txChan <- result
					},
					func(retryErr error) {
						log.Printf("failed to get tx %s after retries: %v", tx, retryErr)
						txChan <- nil
					},
					tx,
				)
				return
			}
			txChan <- txResponse
		}(tx)
	}

	// Close channel when all goroutines finish to avoid deadlock
	go func() {
		wg.Wait()
		close(txChan)
	}()

	// Collect results from the channel
	transactions := make([]*rc.TxResponse, 0, nTxs)
	for tx := range txChan {
		transactions = append(transactions, tx)
	}
	return transactions
}

func (q *QueryOperator) GetLatestBlockHeight() (uint64, error) {
	latestBlockHeight, err := q.rpcClient.GetLatestBlockHeight()
	if err != nil {
		return 0, err
	}
	return latestBlockHeight, nil
}
