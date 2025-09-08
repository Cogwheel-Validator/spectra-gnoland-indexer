package database

import "context"

// FindExistingAccounts finds the existing accounts in the database
//
// Usage:
//
// Used within the account cache package to get query about the existing accounts
// and then we can know which ones to insert
//
// Args:
//
//   - addresses: the addresses to check
//   - chainName: the name of the chain
//
// Returns:
//
//   - map[string]int32: the map of existing addresses and their ids
//   - error: if the query fails
func (t *TimescaleDb) FindExistingAccounts(addresses []string, chainName string, searchValidators bool) (map[string]int32, error) {
	addressesMap := make(map[string]int32)
	// we need to check if the addresses are already in the map
	// so we make this query to the db to get the addresses that are already in the map
	query := ""
	if searchValidators {
		query = `
	SELECT address, id
	FROM gno_validators
	WHERE chain_name = $1
	AND address = ANY($2)
	`
	} else {
		query = `
		SELECT address, id
		FROM gno_addresses
		WHERE chain_name = $1
		AND address = ANY($2)
		`
	}
	rows, err := t.pool.Query(context.Background(), query, chainName, addresses)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var address string
		var id int32
		err := rows.Scan(&address, &id)
		if err != nil {
			return nil, err
		}
		addressesMap[address] = id
	}
	// return the map of existing addresses
	return addressesMap, nil
}

// GetAllAddresses gets all the addresses from the database for a given chain
//
// Usage:
//
// # Only used when the program is initializing to get all the accounts with their ids
//
// Args:
//
//   - chainName: the name of the chain
//
// Returns:
//
//   - map[string]int32: the map of all accounts and their ids
//   - error: if the query fails
func (t *TimescaleDb) GetAllAddresses(chainName string, searchValidators bool) (map[string]int32, error) {
	addressesMap := make(map[string]int32)
	// we need to check if we are searching for validators or accounts
	query := ""
	if searchValidators {
		query += `
		SELECT address, id
		FROM gno_validators
		WHERE chain_name = $1
		`
	} else {
		query += `
		SELECT address, id
		FROM gno_addresses
		WHERE chain_name = $1
		`
	}
	rows, err := t.pool.Query(context.Background(), query, chainName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	for rows.Next() {
		var address string
		var id int32
		err := rows.Scan(&address, &id)
		if err != nil {
			return nil, err
		}
		addressesMap[address] = id
	}
	return addressesMap, nil
}

// CheckCurrentDatabaseName checks the current database name
//
// Usage:
//
// Used to check if the current database is "gnoland"
//
// Returns:
//
//   - string: the name of the current database
//   - error: if the query fails
func (t *TimescaleDb) CheckCurrentDatabaseName() (string, error) {
	query := `
	SELECT current_database()
	`
	row := t.pool.QueryRow(context.Background(), query)
	var currentDb string
	err := row.Scan(&currentDb)
	if err != nil {
		return "", err
	}
	return currentDb, nil
}

// GetLastBlockHeight gets the last block height from the database for a given chain
//
// Usage:
//
// # Used to get the last block height from the database for a given chain
//
// Args:
//   - chainName: the name of the chain
//
// Returns:
//   - uint64: the last block height
//   - error: if the query fails
func (t *TimescaleDb) GetLastBlockHeight(chainName string) (uint64, error) {
	query := `
	SELECT MAX(height)
	FROM blocks
	WHERE chain_name = $1
	`
	row := t.pool.QueryRow(context.Background(), query, chainName)
	var lastBlockHeight uint64
	err := row.Scan(&lastBlockHeight)
	if err != nil {
		return 0, err
	}
	return lastBlockHeight, nil
}
