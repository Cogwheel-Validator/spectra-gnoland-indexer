package accountcache

import (
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/src/database"
)

func NewAccountCache(db *database.TimescaleDb) *AccountCache {
	return &AccountCache{
		address:      make(map[string]int32),
		db:           db,
		highestIndex: 0,
	}
}
