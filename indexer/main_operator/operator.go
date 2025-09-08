package mainoperator

import (
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
