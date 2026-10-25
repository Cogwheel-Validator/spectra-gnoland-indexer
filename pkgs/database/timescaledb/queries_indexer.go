package timescaledb

import (
	"context"
)

// ResolveAddresses returns the id of every given address, inserting the ones the
// database does not know yet. The resolve_addresses / resolve_validators function
// also records each address as seen on chainName and counts the ones that are new
// to the chain in chain_stats.
func (t *TimescaleDb) ResolveAddresses(
	ctx context.Context,
	addresses []string,
	chainName string,
	validators bool,
) (map[string]int32, error) {
	query := `SELECT address, id FROM resolve_addresses($1, $2)`
	if validators {
		query = `SELECT address, id FROM resolve_validators($1, $2)`
	}
	rows, err := t.pool.Query(ctx, query, chainName, addresses)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	addressesMap := make(map[string]int32, len(addresses))
	for rows.Next() {
		var address string
		var id int32
		if err := rows.Scan(&address, &id); err != nil {
			return nil, err
		}
		addressesMap[address] = id
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return addressesMap, nil
}

// CheckCurrentDatabaseName returns the name of the database the pool is connected to.
func (t *TimescaleDb) CheckCurrentDatabaseName(ctx context.Context) (string, error) {
	query := `
	SELECT current_database()
	`
	row := t.pool.QueryRow(ctx, query)
	var currentDb string
	err := row.Scan(&currentDb)
	if err != nil {
		return "", err
	}
	return currentDb, nil
}

// GetLastBlockHeight returns the highest indexed block height for the given chain.
func (t *TimescaleDb) GetLastBlockHeight(ctx context.Context, chainName string) (uint64, error) {
	query := `
	SELECT
	coalesce(
	    (SELECT
			height
		FROM blocks
		WHERE chain_name = $1
		ORDER BY height DESC
		LIMIT 1
		), 0) AS height
	`
	row := t.pool.QueryRow(ctx, query, chainName)
	var lastBlockHeight uint64
	err := row.Scan(&lastBlockHeight)
	if err != nil {
		return 0, err
	}
	return lastBlockHeight, nil
}
