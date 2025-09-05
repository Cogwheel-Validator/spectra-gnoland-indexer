package dataprocessor

import (
	addressCache "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/address_cache"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/database"
)

type DataProcessor struct {
	dbPool         *database.TimescaleDb
	addressCache   *addressCache.AddressCache
	validatorCache *addressCache.AddressCache
	chainName      string
}
