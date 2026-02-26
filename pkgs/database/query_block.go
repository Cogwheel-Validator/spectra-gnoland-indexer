package database

import "context"

// GetBlock gets a block from the database for a given height and chain name
//
// Usage:
//
// # Used to get a block from the database for a given height and chain name
//
// Parameters:
//   - height: the height of the block
//   - chainName: the name of the chain
//
// Returns:
//   - *BlockData: the block data
//   - error: if the query fails
func (t *TimescaleDb) GetBlock(ctx context.Context, height uint64, chainName string) (*BlockData, error) {
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
	row := t.pool.QueryRow(ctx, query, height, chainName)
	var block BlockData
	err := row.Scan(&block.Hash, &block.Height, &block.Timestamp, &block.ChainID, &block.Txs)
	if err != nil {
		return nil, err
	}
	return &block, nil
}

func (t *TimescaleDb) GetLatestBlock(ctx context.Context, chainName string) (*BlockData, error) {
	query := `
	SELECT encode(hash, 'base64'), 
	height, 
	timestamp, 
	chain_id, 
	(SELECT array_agg(upper(encode(tx, 'base64')))
	FROM unnest(blocks.txs) AS tx 
	) AS txs
	FROM blocks
	WHERE chain_name = $1
	ORDER BY height DESC
	LIMIT 1
	`
	row := t.pool.QueryRow(ctx, query, chainName)
	var block BlockData
	err := row.Scan(&block.Hash, &block.Height, &block.Timestamp, &block.ChainID, &block.Txs)
	if err != nil {
		return nil, err
	}
	return &block, nil
}

// GetLastXBlocks gets the last x blocks from the database for a given chain name
//
// Usage:
//
// # Used to get the last x blocks from the database for a given chain name
//
// Parameters:
//   - chainName: the name of the chain
//   - x: the number of blocks to get
//
// Returns:
//   - []*BlockData: the last x blocks
//   - error: if the query fails
func (t *TimescaleDb) GetLastXBlocks(ctx context.Context, chainName string, x uint64) ([]*BlockData, error) {
	query := `
	SELECT encode(hash, 'base64'), 
	height, 
	timestamp, 
	chain_id, 
	(SELECT array_agg(upper(encode(tx, 'base64')))
	FROM unnest(blocks.txs) AS tx 
	) AS txs
	FROM blocks
	WHERE chain_name = $1
	ORDER BY height DESC
	LIMIT $2
	`
	rows, err := t.pool.Query(ctx, query, chainName, x)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	blocks := make([]*BlockData, 0)
	for rows.Next() {
		block := &BlockData{}
		err := rows.Scan(&block.Hash, &block.Height, &block.Timestamp, &block.ChainID, &block.Txs)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return blocks, nil
}

// GetFromToBlocks gets a range of blocks from the database for a given height range and chain name
//
// Usage:
//
// # Used to get a range of blocks from the database for a given height range and chain name
//
// Parameters:
//   - fromHeight: the starting height of the block
//   - toHeight: the ending height of the block (inclusive)
//   - chainName: the name of the chain
//
// Returns:
//   - []*BlockData: the range of block data
//   - error: if the query fails
func (t *TimescaleDb) GetFromToBlocks(ctx context.Context, fromHeight uint64, toHeight uint64, chainName string) ([]*BlockData, error) {
	query := `
	SELECT encode(hash, 'base64'), 
	height, 
	timestamp, 
	chain_id, 
	(SELECT array_agg(upper(encode(tx, 'base64')))
	FROM unnest(blocks.txs) AS tx 
	) AS txs
	FROM blocks
	WHERE height >= $1 AND height <= $2
	AND chain_name = $3
	`
	rows, err := t.pool.Query(ctx, query, fromHeight, toHeight, chainName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	blocks := make([]*BlockData, 0)
	for rows.Next() {
		block := &BlockData{}
		err := rows.Scan(&block.Hash, &block.Height, &block.Timestamp, &block.ChainID, &block.Txs)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, block)
	}
	return blocks, nil
}
