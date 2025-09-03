package addresscache

import (
	"log"
	"maps"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/database"
)

func NewAddressCache(db *database.TimescaleDb) *AddressCache {
	return &AddressCache{
		address:      make(map[string]int32),
		db:           db,
		highestIndex: 0,
	}
}

// AddAddresses is a method to add addresses to the cache
func (a *AddressCache) AddAddresses(newAddresses map[string]int32) {
	// add the addresses to the cache
	maps.Copy(a.address, newAddresses)
}

// LoadAddresses is a method to load addresses from the database into the cache
//
// This method is called when the program starts and when the cache is empty
// Should only be called once per program start
func (a *AddressCache) LoadAddresses(chainName string, searchValidators bool) error {
	addresses, err := a.db.GetAllAddresses(chainName, searchValidators)
	if err != nil {
		return err
	}
	a.address = addresses
	return nil
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
