package mainoperator

import (
	addressCache "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/address_cache"
	dataProcessor "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/data_processor"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/database"
	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
)

// MainOperator is the "brain" of the indexer
//
// It is responsible for delegating the tasks and to
// allocate recources needed for every function or method
type MainOperator struct {
	db             *database.TimescaleDb
	addressCache   *addressCache.AddressCache
	rpcClient      *rpcClient.RateLimitedRpcClient
	dataProcessor  *dataProcessor.DataProcessor
	validatorCache *addressCache.AddressCache
	chainName      string
}
