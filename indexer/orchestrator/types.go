package orchestrator

import (
	addressCache "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/address_cache"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/config"
	dataProcessor "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/data_processor"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/database"
	queryOperator "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/query"
	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
)

type Orchestrator struct {
	db             *database.TimescaleDb
	rpcClient      *rpcClient.RateLimitedRpcClient
	validatorCache *addressCache.AddressCache
	chainName      string
	dataProcessor  *dataProcessor.DataProcessor
	queryOperator  *queryOperator.QueryOperator
	runningMode    string
	config         *config.Config
}
