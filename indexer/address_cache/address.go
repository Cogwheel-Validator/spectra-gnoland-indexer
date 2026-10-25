package addresscache

import (
	"context"
	"fmt"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/logger"
	arc "github.com/hashicorp/golang-lru/arc/v2"
)

var l = logger.Get()

const (
	// DefaultAddressCacheSize is the default number of regular addresses kept in memory
	DefaultAddressCacheSize = 100_000
	// DefaultValidatorCacheSize is the default number of validator addresses kept in memory
	DefaultValidatorCacheSize = 1_000
)

// NewAddressCache is a constructor for the AddressCache struct
//
// The cache starts empty and fills up as AddressSolver resolves addresses.
//
// Parameters:
//   - db: the database connection interface
//   - chainName: the name of the chain
//   - validators: whether the cache holds validator addresses
//   - size: the maximum number of addresses kept in memory
//
// Returns:
//   - *AddressCache: the AddressCache struct
//   - error: if the size is not a positive number
func NewAddressCache(db DatabaseForAddresses, chainName string, validators bool, size int) (*AddressCache, error) {
	cache, err := arc.NewARC[string, int32](size)
	if err != nil {
		return nil, fmt.Errorf("failed to create address cache of size %d: %w", size, err)
	}
	return &AddressCache{
		cache:      cache,
		db:         db,
		chainName:  chainName,
		validators: validators,
	}, nil
}

// AddressSolver returns the id of every given address
//
// Addresses in the cache are answered from memory. The rest go to the database in one
// query that inserts the ones it does not have yet and returns the id for all of them,
// those are then added to the cache.
//
// Parameters:
//   - addresses: the addresses to solve
//   - retryAttempts: the number of attempts for the database query
//
// Returns:
//   - Resolved: the id of each address in the batch
//   - error: if some addresses could not be resolved, Resolved still holds the rest
func (a *AddressCache) AddressSolver(addresses []string, retryAttempts uint8) (Resolved, error) {
	resolved := make(Resolved, len(addresses))
	missing := make([]string, 0)
	queued := make(map[string]struct{})
	for _, addr := range addresses {
		// empty addresses are never stored, GetAddress returns 0 for them
		if addr == "" {
			continue
		}
		if _, ok := resolved[addr]; ok {
			continue
		}
		if _, ok := queued[addr]; ok {
			continue
		}
		if id, ok := a.cache.Get(addr); ok {
			resolved[addr] = id
		} else {
			queued[addr] = struct{}{}
			missing = append(missing, addr)
		}
	}
	if len(missing) == 0 {
		return resolved, nil
	}

	err := a.resolveWithRetry(missing, retryAttempts, resolved)
	return resolved, err
}

// resolveWithRetry queries the database for the missing addresses until all of them
// are resolved or the attempts run out. Each attempt only asks for what is still
// missing, so a partial answer is never thrown away.
func (a *AddressCache) resolveWithRetry(missing []string, retryAttempts uint8, resolved Resolved) error {
	attempts := max(int(retryAttempts), 1)
	var lastErr error
	for range attempts {
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		ids, err := a.db.ResolveAddresses(ctx, missing, a.chainName, a.validators)
		cancel()
		if err != nil {
			lastErr = err
			l.Warn("failed to resolve addresses, retrying", "count", len(missing), "validators", a.validators, "error", err)
			continue
		}

		stillMissing := missing[:0]
		for _, addr := range missing {
			if id, ok := ids[addr]; ok {
				a.cache.Add(addr, id)
				resolved[addr] = id
			} else {
				stillMissing = append(stillMissing, addr)
			}
		}
		missing = stillMissing
		if len(missing) == 0 {
			return nil
		}
		lastErr = nil
	}

	if lastErr != nil {
		return fmt.Errorf("failed to resolve %d addresses: %w", len(missing), lastErr)
	}
	return fmt.Errorf("failed to resolve %d addresses: not returned by the database", len(missing))
}
