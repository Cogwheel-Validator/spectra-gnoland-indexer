package addresscache

import (
	"log"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/src/database"
)

func NewAddressCache(db *database.TimescaleDb) *AddressCache {
	return &AddressCache{
		address:      make(map[string]int32),
		db:           db,
		highestIndex: 0,
	}
}

func (t *AddressCache) LoadAddresses(chainName string, searchValidators bool) error {
	addresses, err := t.db.GetAllAddresses(chainName, searchValidators)
	if err != nil {
		return err
	}
	t.address = addresses
	return nil
}

func (t *AddressCache) AddressSolver(
	address []string,
	chainName string,
	insertValidators bool,
	retryAttempts uint8,
	oneByOne *bool,
) (int32, error) {
	// first check if the address are in the cache
	var newAddresses []string
	for _, addr := range address {
		if _, ok := t.address[addr]; !ok {
			newAddresses = append(newAddresses, addr)
		}
	}
	if len(newAddresses) > 1 {
		for i := uint8(0); i < retryAttempts; i++ {
			loopErr := t.db.InsertAddresses(newAddresses, chainName, insertValidators)
			if loopErr != nil {
				// in the events the oneByOne is true the program will try to insert the addresses one by one
				// as a final resort with this some might be inserted but some might not
				if *oneByOne && i == retryAttempts-1 {
					for _, addr := range newAddresses {
						loopErr := t.db.InsertAnAddress(addr, chainName, insertValidators)
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
	} else if len(newAddresses) == 1 {
		// if there is only one address to insert, we can do it directly
		loopErr := t.db.InsertAnAddress(newAddresses[0], chainName, insertValidators)
		if loopErr != nil {
			return 0, loopErr
		}
	}

	return 0, nil
}
