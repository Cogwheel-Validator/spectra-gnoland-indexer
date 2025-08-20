package addresscache

import "github.com/Cogwheel-Validator/spectra-gnoland-indexer/src/database"

// AddressCache is a map of addresses tied to their int32 index
//
// This is used to lower the amount of queries to the database
// Int32 should be sufficient since this should be marked with postgres integer which is 32 bits
// Should be able to store 2^31 addresses which is 2.147.483.647 addresses
type AddressCache struct {
	address      map[string]int32
	db           *database.TimescaleDb
	highestIndex int32
}
