package addresscache

import (
	"context"
	"maps"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/logger"
)

var l = logger.Get()

// NewAddressCache is a constructor for the AddressCache struct
// it will load the addresses from the database into the cache
// it will also load the validator addresses if the loadVal is true
// it will load the regular addresses if the loadVal is false
//
// Parameters:
//   - chainName: the name of the chain
//   - db: the database connection interface
//   - loadVal: whether to load the validator addresses
//
// Returns:
//   - *AddressCache: the AddressCache struct
//
// If something is wrong it will throw a fatal error and close the program
func NewAddressCache(chainName string, db DatabaseForAddresses, loadVal bool) *AddressCache {
	addresses, maxIndex, err := loadAddresses(chainName, loadVal, db)
	if err != nil {
		l.Fatal().Caller().Stack().Err(err).Msg("failed to load addresses")
	}
	return &AddressCache{
		address:      addresses,
		db:           db,
		highestIndex: maxIndex,
	}
}

// addAddresses is a internal method to add addresses to the cache
//
// This method is used to add addresses to the cache
// It will add the addresses to the cache and update the highest index
//
// Parameters:
//   - newAddresses: the new addresses to add to the cache
//
// Returns:
//   - nil
func (a *AddressCache) addAddresses(newAddresses map[string]int32) {
	// add the addresses to the cache
	maps.Copy(a.address, newAddresses)
}

// AddressSolver is a special method that is used to solve the addresses
//
// This method is combination of other methods that serves a purpose of handling the addresses
// If the addresses are not in the cache, it will add them to the cache
// If the address is recorded in the db but not in the cache, it will add it to the cache
// If the address is in the cache, it will skip it
//
// Parameters:
//   - address: the addresses to solve
//   - chainName: the chain name
//   - insertValidators: whether to insert validators
//   - retryAttempts: the number of retry attempts
//   - oneByOne: whether to insert the addresses one by one is allowed(special case)
//
// Returns:
//   - nothing/nil
func (a *AddressCache) AddressSolver(
	address []string,
	chainName string,
	insertValidators bool,
	retryAttempts uint8,
	oneByOne *bool,
) {
	newAddresses := a.findUncached(address)
	if len(newAddresses) == 0 {
		return
	}

	// technically this should be handled by LoadAddresses but let's make one more check
	addressToAdd := a.syncExistingFromDB(newAddresses, chainName, insertValidators)

	a.insertWithRetry(addressToAdd, chainName, insertValidators, retryAttempts, oneByOne)
	a.fetchAndCacheInserted(addressToAdd, chainName, insertValidators)
}

// findUncached returns the subset of addresses not present in the cache.
func (a *AddressCache) findUncached(addresses []string) []string {
	var missing []string
	for _, addr := range addresses {
		if _, ok := a.address[addr]; !ok {
			missing = append(missing, addr)
		}
	}
	return missing
}

// syncExistingFromDB queries the database for addresses that are already recorded
// but missing from the cache, adds them to the cache, and returns whichever
// addresses still need to be inserted.
func (a *AddressCache) syncExistingFromDB(addresses []string, chainName string, insertValidators bool) []string {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	existing, err := a.db.FindExistingAccounts(ctx, addresses, chainName, insertValidators)
	if err != nil {
		return nil
	}
	a.addAddresses(existing)

	return a.findUncached(addresses)
}

// insertWithRetry attempts to insert addresses into the database, retrying up to
// retryAttempts times. On the final attempt, if oneByOne is set, it falls back to
// inserting each address individually so partial progress is preserved.
func (a *AddressCache) insertWithRetry(
	addresses []string,
	chainName string,
	insertValidators bool,
	retryAttempts uint8,
	oneByOne *bool,
) {
	if len(addresses) == 0 {
		return
	}

	for i := range retryAttempts {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		cancel()

		err := a.db.InsertAddresses(ctx, addresses, chainName, insertValidators)
		if err == nil {
			return
		}

		if oneByOne != nil && *oneByOne && i == retryAttempts-1 {
			a.insertOneByOne(addresses, chainName, insertValidators)
		}
	}
}

// insertOneByOne inserts addresses one at a time as a last resort, logging any
// individual failures without aborting the remaining inserts.
func (a *AddressCache) insertOneByOne(addresses []string, chainName string, insertValidators bool) {
	for _, addr := range addresses {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		cancel()

		if err := a.db.InsertAddresses(ctx, []string{addr}, chainName, insertValidators); err != nil {
			l.Error().Caller().Stack().Err(err).Msgf("error inserting address: %s", addr)
		}
	}
}

// fetchAndCacheInserted queries the database for the IDs of newly inserted
// addresses and adds them to the cache.
func (a *AddressCache) fetchAndCacheInserted(addresses []string, chainName string, insertValidators bool) {
	if len(addresses) == 0 {
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	newAddrMap, err := a.db.FindExistingAccounts(ctx, addresses, chainName, insertValidators)
	if err != nil {
		l.Error().Caller().Stack().Err(err).Msg("error finding existing accounts")
		return
	}
	a.addAddresses(newAddrMap)
}

// GetAddress is a method to get the address from the cache
//
// This method is used to get the address from the cache
// If the address is not in the cache, it will return 0
//
// Parameters:
//   - address: the address to get
//
// Returns:
//   - int32: the address id
//   - 0 if the address is not in the cache
func (a *AddressCache) GetAddress(address string) int32 {
	if _, ok := a.address[address]; !ok {
		return 0
	} else {
		return a.address[address]
	}
}

// loadAddresses is int function to load addresses from the database into the cache
//
// This method is called when the program starts and when the cache is empty
// Should only be called once per program start
//
// Parameters:
//   - chainName: the name of the chain
//   - loadVal: whether to load the validator addresses
//   - db: the database connection interface
//
// Returns:
//   - map[string]int32: the map of addresses and their ids
//   - error: if the query fails
func loadAddresses(chainName string, loadVal bool, db DatabaseForAddresses) (map[string]int32, int32, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	addresses, maxIndex, err := db.GetAllAddresses(ctx, chainName, loadVal, nil)
	if err != nil {
		return nil, 0, err
	}
	return addresses, maxIndex, nil
}
