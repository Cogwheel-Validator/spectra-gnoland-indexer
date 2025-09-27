package query

import (
	"time"

	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
)

// QueryOperator struct to hold the query operator
// holds:
// - the rpc client interface
// - the retry amount
// - the pause after failing for x amount of times
// - the pause time
// - the exponential backoff
type QueryOperator struct {
	rpcClient          RpcClient
	retryAmount        int
	pause              int
	pauseTime          time.Duration
	exponentialBackoff time.Duration
}

// Rate limiter Gnoland RPC client interface
//
// # The interface is used to get the blocks, latest block height, and txs
//
// Methods:
// - GetBlock: to get a block from the rpc client
// - GetLatestBlockHeight: to get the latest block height from the rpc client
// - GetTx: to get a tx from the rpc client
type RpcClient interface {
	GetBlock(height uint64) (*rpcClient.BlockResponse, *rpcClient.RpcHeightError)
	GetLatestBlockHeight() (uint64, *rpcClient.RpcHeightError)
	GetTx(txHash string) (*rpcClient.TxResponse, *rpcClient.RpcStringError)
}
