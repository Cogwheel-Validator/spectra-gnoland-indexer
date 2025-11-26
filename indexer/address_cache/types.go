package addresscache

import "context"

// A database interface for what AddressCache needs from database
type DatabaseForAddresses interface {
	FindExistingAccounts(ctx context.Context, addresses []string, chainName string, searchValidators bool) (map[string]int32, error)
	InsertAddresses(ctx context.Context, addresses []string, chainName string, insertValidators bool) error
	GetAllAddresses(ctx context.Context, chainName string, searchValidators bool, highestIndex *int32) (map[string]int32, int32, error)
}

// AddressCache is a map of addresses tied to their int32 index in the database
//
// This is used to lower the amount of queries to the database
// Int32 should be sufficient since this should be marked with postgres integer which is 32 bits
// Should be able to store 2^31 addresses which is 2.147.483.647 addresses
type AddressCache struct {
	address      map[string]int32
	db           DatabaseForAddresses
	highestIndex int32
}
