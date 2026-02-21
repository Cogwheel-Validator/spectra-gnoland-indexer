package database

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"strings"
	"time"
)

var defaultLimit = uint64(10)

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

// GetAllBlockSigners gets all of the validators that signed that block + the proposer
//
// Usage:
//
// # Used to get all of the validators that signed that block + the proposer
//
// Parameters:
//   - chainName: the name of the chain
//   - blockHeight: the height of the block
//
// Returns:
//   - *BlockSigners: the block signers
//   - error: if the query fails
func (t *TimescaleDb) GetAllBlockSigners(
	ctx context.Context,
	chainName string,
	blockHeight uint64,
) (*BlockSigners, error) {
	query := `
	SELECT
	vb.block_height,
	gv.address AS proposer,
	array(
		SELECT gv.address 
		FROM unnest(vb.signed_vals) AS signed_val_id
		JOIN gno_validators gv ON gv.id = signed_val_id
	) AS signed_vals
	FROM validator_block_signing vb
	LEFT JOIN gno_validators gv ON vb.proposer = gv.id
	WHERE vb.chain_name = $1
	AND vb.block_height = $2
	`
	row := t.pool.QueryRow(ctx, query, chainName, blockHeight)
	var blockSigners BlockSigners
	err := row.Scan(&blockSigners.BlockHeight, &blockSigners.Proposer, &blockSigners.SignedVals)
	if err != nil {
		return nil, err
	}
	return &blockSigners, nil
}

// GetBankSend gets the bank send message for a given transaction hash
//
// Usage:
//
// # Used to get the bank send message for a given transaction hash
//
// Parameters:
//   - txHash: the hash of the transaction
//   - chainName: the name of the chain
//
// Returns:
//   - []*BankSend: the bank send messages
//   - error: if the query fails
func (t *TimescaleDb) GetBankSend(
	ctx context.Context,
	txHash string,
	chainName string,
) ([]*BankSend, error) {
	query := `
	SELECT 
    encode(bms.tx_hash, 'base64') AS tx_hash,
    bms.timestamp,
    gn_from.address AS from_address,
    gn_to.address AS to_address,
    bms.amount,
    array(
        SELECT gn.address 
        FROM unnest(bms.signers) AS signer_id
        JOIN gno_addresses gn ON gn.id = signer_id
    ) AS signers
	FROM bank_msg_send bms
	LEFT JOIN gno_addresses gn_from ON bms.from_address = gn_from.id
	LEFT JOIN gno_addresses gn_to ON bms.to_address = gn_to.id
	WHERE bms.tx_hash = decode($1, 'base64')
	AND bms.chain_name = $2
	`
	rows, err := t.pool.Query(ctx, query, txHash, chainName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	bankSends := make([]*BankSend, 0)
	for rows.Next() {
		bankSend := &BankSend{}
		err := rows.Scan(
			&bankSend.TxHash,
			&bankSend.Timestamp,
			&bankSend.FromAddress,
			&bankSend.ToAddress,
			&bankSend.Amount,
			&bankSend.Signers,
		)
		if err != nil {
			return nil, err
		}
		bankSends = append(bankSends, bankSend)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return bankSends, nil
}

// GetMsgCall gets the msg call message for a given transaction hash
//
// Usage:
//
// # Used to get the msg call message for a given transaction hash
//
// Parameters:
//   - txHash: the hash of the transaction
//   - chainName: the name of the chain
//
// Returns:
//   - []*MsgCall: the msg call messages
//   - error: if the query fails
func (t *TimescaleDb) GetMsgCall(
	ctx context.Context,
	txHash string,
	chainName string,
) ([]*MsgCall, error) {
	query := `
	SELECT 
	encode(vmc.tx_hash, 'base64') AS tx_hash,
	vmc.message_counter,
	vmc.timestamp,
	gn.address AS caller,
	vmc.pkg_path,
	vmc.func_name,
	vmc.args,
	vmc.send,
	vmc.max_deposit,
	array(
		SELECT gn.address 
		FROM unnest(vmc.signers) AS signer_id
		JOIN gno_addresses gn ON gn.id = signer_id
	) AS signers
	FROM vm_msg_call vmc
	LEFT JOIN gno_addresses gn ON vmc.caller = gn.id
	WHERE vmc.tx_hash = decode($1, 'base64')
	AND vmc.chain_name = $2
	`
	rows, err := t.pool.Query(ctx, query, txHash, chainName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	msgCalls := make([]*MsgCall, 0)
	for rows.Next() {
		msgCall := &MsgCall{}
		err := rows.Scan(
			&msgCall.TxHash,
			&msgCall.MessageCounter,
			&msgCall.Timestamp,
			&msgCall.Caller,
			&msgCall.PkgPath,
			&msgCall.FuncName,
			&msgCall.Args,
			&msgCall.Send,
			&msgCall.MaxDeposit,
			&msgCall.Signers,
		)
		if err != nil {
			return nil, err
		}
		msgCalls = append(msgCalls, msgCall)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return msgCalls, nil
}

// GetMsgAddPackage gets the msg add package message for a given transaction hash
//
// Usage:
//
// # Used to get the msg add package message for a given transaction hash
//
// Parameters:
//   - txHash: the hash of the transaction
//   - chainName: the name of the chain
//
// Returns:
//   - []*MsgAddPackage: the msg add package messages
//   - error: if the query fails
func (t *TimescaleDb) GetMsgAddPackage(
	ctx context.Context,
	txHash string,
	chainName string,
) ([]*MsgAddPackage, error) {
	query := `
	SELECT 
	encode(vmap.tx_hash, 'base64') AS tx_hash,
	vmap.message_counter,
	vmap.timestamp,
	gn.address AS creator,
	vmap.pkg_path,
	vmap.pkg_name,
	vmap.pkg_file_names,
	vmap.send,
	vmap.max_deposit,
	array(
		SELECT gn.address 
		FROM unnest(vmap.signers) AS signer_id
		JOIN gno_addresses gn ON gn.id = signer_id
	) AS signers
	FROM vm_msg_add_package vmap
	LEFT JOIN gno_addresses gn ON vmap.creator = gn.id
	WHERE vmap.tx_hash = decode($1, 'base64')
	AND vmap.chain_name = $2
	`
	rows, err := t.pool.Query(ctx, query, txHash, chainName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	msgAddPackages := make([]*MsgAddPackage, 0)
	for rows.Next() {
		msgAddPackage := &MsgAddPackage{}
		err := rows.Scan(
			&msgAddPackage.TxHash,
			&msgAddPackage.MessageCounter,
			&msgAddPackage.Timestamp,
			&msgAddPackage.Creator,
			&msgAddPackage.PkgPath,
			&msgAddPackage.PkgName,
			&msgAddPackage.PkgFileNames,
			&msgAddPackage.Send,
			&msgAddPackage.MaxDeposit,
			&msgAddPackage.Signers,
		)
		if err != nil {
			return nil, err
		}
		msgAddPackages = append(msgAddPackages, msgAddPackage)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return msgAddPackages, nil
}

// GetMsgRun gets the msg run message for a given transaction hash
//
// Usage:
//
// # Used to get the msg run message for a given transaction hash
//
// Parameters:
//   - txHash: the hash of the transaction
//   - chainName: the name of the chain
//
// Returns:
//   - []*MsgRun: the msg run messages
//   - error: if the query fails
func (t *TimescaleDb) GetMsgRun(
	ctx context.Context,
	txHash string,
	chainName string,
) ([]*MsgRun, error) {
	query := `
	SELECT 
	encode(vmr.tx_hash, 'base64') AS tx_hash,
	vmr.message_counter,
	vmr.timestamp,
	gn.address AS caller,
	vmr.pkg_path,
	vmr.pkg_name,
	vmr.pkg_file_names,
	vmr.send,
	vmr.max_deposit,
	array(
		SELECT gn.address 
		FROM unnest(vmr.signers) AS signer_id
		JOIN gno_addresses gn ON gn.id = signer_id
	) AS signers
	FROM vm_msg_run vmr
	LEFT JOIN gno_addresses gn ON vmr.caller = gn.id
	WHERE vmr.tx_hash = decode($1, 'base64')
	AND vmr.chain_name = $2
	`
	rows, err := t.pool.Query(ctx, query, txHash, chainName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	msgRuns := make([]*MsgRun, 0)
	for rows.Next() {
		msgRun := &MsgRun{}
		err := rows.Scan(
			&msgRun.TxHash,
			&msgRun.MessageCounter,
			&msgRun.Timestamp,
			&msgRun.Caller,
			&msgRun.PkgPath,
			&msgRun.PkgName,
			&msgRun.PkgFileNames,
			&msgRun.Send,
			&msgRun.MaxDeposit,
			&msgRun.Signers,
		)
		if err != nil {
			return nil, err
		}
		msgRuns = append(msgRuns, msgRun)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return msgRuns, nil
}

// GetTransaction gets the transaction for a given transaction hash
//
// Usage:
//
// # Used to get the transaction for a given transaction hash
//
// Parameters:
//   - txHash: the hash of the transaction
//   - chainName: the name of the chain
//
// Returns:
//   - *Transaction: the transaction
//   - error: if the query fails
func (t *TimescaleDb) GetTransaction(ctx context.Context, txHash string, chainName string) (*Transaction, error) {
	query := `
	SELECT 
	encode(tx.tx_hash, 'base64') AS tx_hash,
	tx.timestamp,
	tx.block_height,
	tx.tx_events,
	tx.gas_used,
	tx.gas_wanted,
	tx.fee,
	tx.msg_types
	FROM transaction_general tx
	WHERE tx.tx_hash = decode($1, 'base64')
	AND tx.chain_name = $2
	`
	row := t.pool.QueryRow(ctx, query, txHash, chainName)
	var transaction Transaction
	err := row.Scan(
		&transaction.TxHash,
		&transaction.Timestamp,
		&transaction.BlockHeight,
		&transaction.TxEvents,
		&transaction.GasUsed,
		&transaction.GasWanted,
		&transaction.Fee,
		&transaction.MsgTypes,
	)
	if err != nil {
		log.Println("error getting transaction", err)
		return nil, err
	}
	return &transaction, nil
}

// GetLastXTransactions gets the last x transactions from the database for a given chain name
//
// Usage:
//
// # Used to get the last x transactions from the database for a given chain name
//
// Parameters:
//   - chainName: the name of the chain
//   - x: the number of transactions to get
func (t *TimescaleDb) GetLastXTransactions(ctx context.Context, chainName string, x uint64) ([]*Transaction, error) {
	query := `
	SELECT
	encode(tx.tx_hash, 'base64') AS tx_hash,
	tx.timestamp,
	tx.block_height,
	tx.tx_events,
	tx.gas_used,
	tx.gas_wanted,
	tx.fee,
	tx.msg_types
	FROM transaction_general tx
	WHERE tx.chain_name = $1
	ORDER BY tx.timestamp DESC
	LIMIT $2
	`
	rows, err := t.pool.Query(ctx, query, chainName, x)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	transactions := make([]*Transaction, 0)
	for rows.Next() {
		transaction := &Transaction{}
		err := rows.Scan(&transaction.TxHash, &transaction.Timestamp, &transaction.BlockHeight, &transaction.TxEvents, &transaction.GasUsed, &transaction.GasWanted, &transaction.Fee, &transaction.MsgTypes)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, transaction)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetMsgTypes gets the message type for a given transaction hash
//
// Usage:
//
// # Used to get the message type for a given transaction hash
//
// Parameters:
//   - txHash: the hash of the transaction
//   - chainName: the name of the chain
//
// Returns:
//   - []string: the message types
//   - error: if the query fails
func (t *TimescaleDb) GetMsgTypes(ctx context.Context, txHash string, chainName string) ([]string, error) {
	query := `
	SELECT msg_types
	FROM transaction_general
	WHERE tx_hash = decode($1, 'base64')
	AND chain_name = $2
	`
	row := t.pool.QueryRow(ctx, query, txHash, chainName)
	var msgTypes []string
	err := row.Scan(&msgTypes)
	if err != nil {
		return nil, err
	}
	return msgTypes, nil
}

// GetAddressTxs gets the transactions for a given address for a certain time period
//
// Usage:
//
// # Used to get the transactions for a given address for a certain time period
//
// Parameters:
//   - address: the address
//   - chainName: the name of the chain
//   - fromTimestamp: the starting timestamp
//   - toTimestamp: the ending timestamp
//
// Returns:
//   - []*AddressTx: the transactions
//   - error: if the query fails
func (t *TimescaleDb) GetAddressTxs(
	ctx context.Context,
	address string,
	chainName string,
	fromTimestamp *time.Time,
	toTimestamp *time.Time,
	limit *uint64,
	page *uint64,
	cursor *string,
) (*[]AddressTx, string, uint64, error) {
	hasTsRange := fromTimestamp != nil && toTimestamp != nil
	noTsRange := fromTimestamp == nil && toTimestamp == nil

	var mode string
	switch {
	case hasTsRange:
		mode = "timestamp"
	case noTsRange && page == nil:
		mode = "cursor"
	case noTsRange && cursor == nil:
		mode = "limit_page"
	default:
		return nil, "", 0, fmt.Errorf("invalid query parameters")
	}

	accountId, err := t.getAccountId(ctx, address, chainName)
	if err != nil {
		return nil, "", 0, fmt.Errorf("error getting account id: %w", err)
	}

	txCount, err := t.getTxsCount(ctx, accountId, chainName)
	if err != nil {
		return nil, "", 0, fmt.Errorf("error getting tx count: %w", err)
	}

	var addressTxs *[]AddressTx
	var nextCursor string

	switch mode {
	case "timestamp":
		addressTxs, err = t.getAddressTxsTimestampQuery(
			ctx, accountId, chainName, *fromTimestamp, *toTimestamp, limit,
		)
		if err != nil {
			return nil, "", 0, err
		}
	case "cursor":
		addressTxs, nextCursor, err = t.getAddressTxsCursorQuery(
			ctx, accountId, chainName, cursor, limit, txCount,
		)
		if err != nil {
			return nil, "", 0, err
		}
	case "limit_page":
		addressTxs, err = t.getAddressTxsLimitPageQuery(
			ctx, accountId, chainName, limit, *page,
		)
		if err != nil {
			return nil, "", 0, err
		}
	}

	return addressTxs, nextCursor, txCount, nil
}

func (t *TimescaleDb) getAddressTxsTimestampQuery(
	ctx context.Context,
	accountId int32,
	chainName string,
	fromTimestamp time.Time,
	toTimestamp time.Time,
	limit *uint64,
) (*[]AddressTx, error) {
	if limit == nil {
		limit = &defaultLimit
	}
	var args []any

	query := `
		SELECT
		encode(tx.tx_hash, 'base64') AS tx_hash,
		tx.timestamp,
		tx.msg_types
		FROM address_tx tx
		WHERE tx.address = $1
		AND tx.chain_name = $2
		AND tx.timestamp >= $3
		AND tx.timestamp <= $4
		ORDER BY tx.timestamp DESC
		LIMIT $5
		`
	args = append(args, accountId, chainName, fromTimestamp, toTimestamp, *limit)

	addressTxs, err := t.execAccQuery(ctx, query, args)
	if err != nil {
		return nil, err
	}
	return addressTxs, nil
}

func (t *TimescaleDb) getAddressTxsCursorQuery(
	ctx context.Context,
	accountId int32,
	chainName string,
	cursor *string,
	limit *uint64,
	txCount uint64,
) (*[]AddressTx, string, error) {
	if limit == nil {
		limit = &defaultLimit
	}
	// Fetch limit+1 to detect if there are more rows; only return limit to the caller.
	fetchLimit := *limit + 1
	var query string
	var args []any

	if cursor == nil {
		query = `
		SELECT
		encode(tx.tx_hash, 'base64') AS tx_hash,
		tx.timestamp,
		tx.msg_types
		FROM address_tx tx
		WHERE tx.address = $1
		AND tx.chain_name = $2
		ORDER BY tx.timestamp DESC
		LIMIT $3
		`
		args = append(args, accountId, chainName, fetchLimit)
	} else {
		timestamp, txHash, err := unmarshalCursorParam(*cursor)
		if err != nil {
			return nil, "", err
		}
		decodedTxHash, err := base64.URLEncoding.Strict().DecodeString(txHash)
		if err != nil {
			return nil, "", fmt.Errorf("error decoding tx hash: %w", err)
		}
		query = `
		SELECT
		encode(tx.tx_hash, 'base64') AS tx_hash,
		tx.timestamp,
		tx.msg_types
		FROM address_tx tx
		WHERE tx.address = $1
		AND tx.chain_name = $2
		AND (tx.timestamp, tx.tx_hash) < ($3::timestamptz, $4)
		ORDER BY tx.timestamp DESC, tx.tx_hash DESC
		LIMIT $5
		`
		args = append(args, accountId, chainName, timestamp, decodedTxHash, fetchLimit)
	}

	addressTxs, err := t.execAccQuery(ctx, query, args)
	if err != nil {
		return nil, "", err
	}
	// If we got more than limit, there is a next page: return only the first limit and set nextCursor.
	if len(*addressTxs) > int(*limit) {
		page := (*addressTxs)[:int(*limit)]
		lastAddressTx := page[len(page)-1]
		nextCursor := makeCursorParam(lastAddressTx.Timestamp, lastAddressTx.Hash)
		return &page, nextCursor, nil
	}
	return addressTxs, "", nil
}

func (t *TimescaleDb) getAddressTxsLimitPageQuery(
	ctx context.Context,
	accountId int32,
	chainName string,
	limit *uint64,
	page uint64,
) (*[]AddressTx, error) {
	if limit == nil {
		limit = &defaultLimit
	}

	var query string
	var args []any

	offset := page * *limit

	query = `
	SELECT
	encode(tx.tx_hash, 'base64') AS tx_hash,
	tx.timestamp,
	tx.msg_types
	FROM address_tx tx
	WHERE tx.address = $1
	AND tx.chain_name = $2
	ORDER BY tx.timestamp DESC
	LIMIT $4 OFFSET $5
	`
	args = append(args, accountId, chainName, *limit, offset)

	addressTxs, err := t.execAccQuery(ctx, query, args)
	if err != nil {
		return nil, err
	}
	return addressTxs, nil
}

func (t *TimescaleDb) getTxsCount(
	ctx context.Context,
	accountId int32,
	chainName string,
) (uint64, error) {
	query := `
	SELECT COUNT(*) FROM address_tx WHERE address = $1 AND chain_name = $2
	`
	row := t.pool.QueryRow(ctx, query, accountId, chainName)
	var count uint64
	err := row.Scan(&count)
	if err != nil {
		return 0, err
	}
	return count, nil
}

func makeCursorParam(
	timestamp time.Time,
	txHash string,
) string {
	txHashBytes, err := base64.StdEncoding.DecodeString(txHash)
	if err != nil {
		// TODO: log error
		return "error decoding tx hash"
	}
	timestamp = timestamp.Round(time.Second)
	base64Url := base64.URLEncoding.Strict().EncodeToString(txHashBytes)
	return timestamp.Format(time.RFC3339) + "|" + base64Url
}

func unmarshalCursorParam(
	cursor string,
) (time.Time, string, error) {
	parts := strings.Split(cursor, "|")
	if len(parts) != 2 {
		return time.Time{}, "", fmt.Errorf("invalid cursor")
	}
	timestamp, err := time.Parse(time.RFC3339, parts[0])
	if err != nil {
		return time.Time{}, "", err
	}
	return timestamp, parts[1], nil
}

func (t *TimescaleDb) execAccQuery(
	ctx context.Context,
	query string,
	args []any,
) (*[]AddressTx, error) {

	addressTxs := make([]AddressTx, 0)
	rows, err := t.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var addressTx AddressTx
		err := rows.Scan(&addressTx.Hash, &addressTx.Timestamp, &addressTx.MsgTypes)
		if err != nil {
			return nil, err
		}
		addressTxs = append(addressTxs, addressTx)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return &addressTxs, nil
}

func (t *TimescaleDb) getAccountId(
	ctx context.Context,
	address string,
	chainName string,
) (int32, error) {
	query := `
	SELECT id FROM gno_addresses WHERE address = $1 AND chain_name = $2
	`
	row := t.pool.QueryRow(ctx, query, address, chainName)
	var id int32
	err := row.Scan(&id)
	if err != nil {
		return 0, err
	}
	return id, nil
}
