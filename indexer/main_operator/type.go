package mainoperator

import (
	addressCache "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/address_cache"
	dataProcessor "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/data_processor"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/database"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/query"
	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
)

// MajorConstructors is a struct that contains the major constructors
// for the indexer
type MajorConstructors struct {
	db             *database.TimescaleDb
	rpcClient      *rpcClient.RateLimitedRpcClient
	validatorCache *addressCache.AddressCache
	addressCache   *addressCache.AddressCache
	dataProcessor  *dataProcessor.DataProcessor
	queryOperator  *query.QueryOperator
}
