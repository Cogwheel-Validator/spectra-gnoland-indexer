package addresscache

import (
	"context"

	arc "github.com/hashicorp/golang-lru/arc/v2"
)

// A database interface for what AddressCache needs from database
type DatabaseForAddresses interface {
	// ResolveAddresses returns the id of every address, inserting the missing ones
	// and marking all of them as seen on chainName.
	ResolveAddresses(ctx context.Context, addresses []string, chainName string, validators bool) (map[string]int32, error)
}

// AddressCache keeps the most useful addresses tied to their int32 index in the database
//
// This is used to lower the amount of queries to the database. The cache is an ARC
// (Adaptive Replacement Cache) with a fixed size, so it tracks both how recently and
// how often an address was used, and a burst of one-off addresses does not push out
// the ones that show up all the time.
//
// An address in the cache is known to be in the database and seen on this chain, so
// a cache hit needs no query at all.
//
// Int32 should be sufficient since this should be marked with postgres integer which is 32 bits
// Should be able to store 2^31 addresses which is 2.147.483.647 addresses
type AddressCache struct {
	cache      *arc.ARCCache[string, int32]
	db         DatabaseForAddresses
	chainName  string
	validators bool
}

// Resolved maps the addresses of a single batch to their ids.
//
// It is filled once by AddressSolver and only read afterwards, so it is safe to share
// between goroutines. It does not depend on the cache, the ids stay available even if
// the cache evicts them while the batch is processed.
type Resolved map[string]int32

// GetAddress returns the id of the address, or 0 if it is not part of the batch.
func (r Resolved) GetAddress(address string) int32 {
	return r[address]
}
