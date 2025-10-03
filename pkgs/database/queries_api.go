package database

import "context"

// GetBlock gets a block from the database for a given height and chain name
//
// Usage:
//
// # Used to get a block from the database for a given height and chain name
//
// Args:
//   - height: the height of the block
//   - chainName: the name of the chain
//
// Returns:
//   - *BlockData: the block data
//   - error: if the query fails
func (t *TimescaleDb) GetBlock(height uint64, chainName string) (*BlockData, error) {
	query := `
	SELECT encode(hash, 'base64'), 
	height, 
	timestamp, 
	chain_id, 
	(SELECT array_agg(upper(encode(tx, 'base64')))
	FROM unnest(blocks.txs) AS tx 
	) AS txs
	FROM blocks
	WHERE height = $1
	AND chain_name = $2
	`
	row := t.pool.QueryRow(context.Background(), query, height, chainName)
	var block BlockData
	err := row.Scan(&block.Hash, &block.Height, &block.Timestamp, &block.ChainID, &block.Txs)
	if err != nil {
		return nil, err
	}
	return &block, nil
}

// GetAllIntAddresses gets all the addresses from the database for a given chain name.
// Unlike the address cache for the indexer, the difference here is that map here is of int32 showing
// the string address. It should be the same as GetAllAddresses but with the key being the int32 address.
//
// Usage:
//
// # Used to get all the addresses from the database for a given chain name
//
// Args:
//   - chainName: the name of the chain
//   - searchValidators: whether to search for validators or accounts
//   - highestIndex: the highest index of the addresses already recorded or it could be a 0
//
// Returns:
//   - map[int32]string: the map of all addresses and their ids
//   - error: if the query fails
func (t *TimescaleDb) GetAllIntAddresses(
	chainName string,
	searchValidators bool,
	highestIndex *int32) (map[int32]string, int32, error) {
	addressesMap := make(map[int32]string)
	var maxIndex int32 = 0
	if highestIndex != nil {
		maxIndex = *highestIndex
	}
	query := ""
	if searchValidators {
		query += `
		SELECT address, id
		FROM gno_validators
		WHERE chain_name = $1
		AND id > $2
		`
	} else {
		query += `
		SELECT address, id
		FROM gno_addresses
		WHERE chain_name = $1
		AND id > $2
		`
	}
	rows, err := t.pool.Query(context.Background(), query, chainName, highestIndex)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	for rows.Next() {
		var address string
		var id int32
		err := rows.Scan(&address, &id)
		if err != nil {
			return nil, 0, err
		}
		addressesMap[id] = address
		if id > maxIndex {
			maxIndex = id
		}
	}
	return addressesMap, maxIndex, nil
}
