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
		FROM gno_accounts
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

// GetAllAccounts gets all the accounts from the database for a given chain
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
func (t *TimescaleDb) GetAllAccounts(chainName string, searchValidators bool) (map[string]int32, error) {
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
		FROM gno_accounts
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
