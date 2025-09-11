package orchestrator

import (
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/config"
	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
)

// Define interfaces where we USE them (consumer-side interfaces)
type DataProcessor interface {
	ProcessValidatorAddresses(blocks []*rpcClient.BlockResponse, fromHeight uint64, toHeight uint64)
	ProcessBlocks(blocks []*rpcClient.BlockResponse, fromHeight uint64, toHeight uint64)
	ProcessTransactions(transactions map[*rpcClient.TxResponse]time.Time, compressEvents bool, fromHeight uint64, toHeight uint64)
	ProcessMessages(transactions map[*rpcClient.TxResponse]time.Time, fromHeight uint64, toHeight uint64) error
	ProcessValidatorSignings(blocks []*rpcClient.BlockResponse, fromHeight uint64, toHeight uint64)
}

type QueryOperator interface {
	GetFromToBlocks(fromHeight uint64, toHeight uint64) []*rpcClient.BlockResponse
	GetTransactions(txs []string) []*rpcClient.TxResponse
	GetLatestBlockHeight() (uint64, error)
}

// Only needed for one opetaion
// Part of the timescaledb interface
type DatabaseHeight interface {
	GetLastBlockHeight(chainName string) (uint64, error)
}

// Only needed for one opetaion
// Part of the rpc client interface
type GnolandRpcClient interface {
	GetLatestBlockHeight() (uint64, *rpcClient.RpcHeightError)
}

// Orchestrator struct to hold the orchestrator
// holds:
// - the database height interface
// - the gno rpc client interface
// - the chain name
// - the data processor interface
// - the query operator interface
// - the running mode
// - the config
type Orchestrator struct {
	db            DatabaseHeight
	gnoRpcClient  GnolandRpcClient
	chainName     string
	dataProcessor DataProcessor
	queryOperator QueryOperator
	runningMode   string
	config        *config.Config
}
