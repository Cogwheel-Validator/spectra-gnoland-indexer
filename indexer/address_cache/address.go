package addresscache

import (
	"log"
	"maps"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/database"
)

// NewAddressCache is a constructor for the AddressCache struct
// it will load the addresses from the database into the cache
// it will also load the validator addresses if the loadVal is true
// it will load the regular addresses if the loadVal is false
//
// Args:
//   - chainName: the name of the chain
//   - db: the database connection
//   - loadVal: whether to load the validator addresses
//
// Returns:
//   - *AddressCache: the AddressCache struct
//
// If something is wrong it will throw a fatal error and close the program
func NewAddressCache(chainName string, db *database.TimescaleDb, loadVal bool) *AddressCache {
	if loadVal == true {
		// if true load the validator addresses
		addresses, err := loadAddresses(chainName, loadVal, db)
		if err != nil {
			log.Fatalf("failed to load addresses: %v", err)
		}
		return &AddressCache{
			address:      addresses,
			db:           db,
			highestIndex: 0,
		}
	} else {
		// if false load the regular addresses
		addresses, err := loadAddresses(chainName, loadVal, db)
		if err != nil {
			log.Fatalf("failed to load addresses: %v", err)
		}
		return &AddressCache{
			address:      addresses,
			db:           db,
			highestIndex: 0,
		}
	}
}

// AddAddresses is a method to add addresses to the cache
//
// This method is used to add addresses to the cache
// It will add the addresses to the cache and update the highest index
//
// Args:
//   - newAddresses: the new addresses to add to the cache
//
// Returns:
//   - nil
func (a *AddressCache) AddAddresses(newAddresses map[string]int32) {
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
// Given that this logic was ported from the cosmos indexer (written in python)
// but given that the Gnoland as chain it self is in a early stage of development
// and the indexer is not yet fully optimized for the Gnoland chain.
// Some of the logic might be redundant and could be simplified but is kept for making sure
// that the addresses are handled correctly.
//
// TODO:
//   - simplify the logic
//   - optimize the code to be more Go like
//   - add more tests
//   - add more documentation
//   - add more logging
//
// Args:
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
	// first check if the address are in the cache
	var newAddresses []string
	for _, addr := range address {
		if _, ok := a.address[addr]; !ok {
			newAddresses = append(newAddresses, addr)
		}
	}
	if len(newAddresses) == 0 {
		// if there are no new addresses, we can return immediately
		return
	}
	// chech if there is already recorded addresses in the db
	// technically this should be handled by LoadAddresses but let's make one more check
	// probably not needed but just in case
	existingAddresses, err := a.db.FindExistingAccounts(address, chainName, insertValidators)
	if err != nil {
		return
	}

	// if there are existing addresses, we need to add them to the cache
	a.AddAddresses(existingAddresses)

	// one last check to see if there are any addresses that are not in the cache
	addressToAdd := make([]string, 0)
	for _, addr := range newAddresses {
		if _, ok := a.address[addr]; !ok {
			addressToAdd = append(addressToAdd, addr)
		}
	}

	if len(addressToAdd) > 1 {
		for i := range retryAttempts {
			loopErr := a.db.InsertAddresses(addressToAdd, chainName, insertValidators)
			if loopErr != nil {
				// in the events the oneByOne is true the program will try to insert the addresses one by one
				// as a final resort with this some might be inserted but some might not
				if *oneByOne && i == retryAttempts-1 {
					for _, addr := range addressToAdd {
						loopErr := a.db.InsertAddresses([]string{addr}, chainName, insertValidators)
						if loopErr != nil {
							// this is a final resort, so we can log the error for debugging purposes
							log.Println("Error inserting address:", addr, "error:", loopErr)
						}
					}
				}
				continue
			}
			break
		}
	} else if len(addressToAdd) == 1 {
		// if there is only one address to insert, we can do it directly
		loopErr := a.db.InsertAddresses([]string{addressToAdd[0]}, chainName, insertValidators)
		if loopErr != nil {
			return
		}
	}
	// at the end of the function we should add the addresses to the cache
	// we need to make a query of the added addresses to get the ids along with the addresses
	newAddrMap, err := a.db.FindExistingAccounts(addressToAdd, chainName, insertValidators)
	if err != nil {
		log.Println("Error finding existing accounts:", err)
		return
	}
	a.AddAddresses(newAddrMap)
}

// GetAddress is a method to get the address from the cache
//
// This method is used to get the address from the cache
// If the address is not in the cache, it will return 0
//
// Args:
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
// Args:
//   - chainName: the name of the chain
//   - loadVal: whether to load the validator addresses
//   - db: the database connection
//
// Returns:
//   - map[string]int32: the map of addresses and their ids
//   - error: if the query fails
func loadAddresses(chainName string, loadVal bool, db *database.TimescaleDb) (map[string]int32, error) {
	addresses, err := db.GetAllAddresses(chainName, loadVal)
	if err != nil {
		return nil, err
	}
	return addresses, nil
}
