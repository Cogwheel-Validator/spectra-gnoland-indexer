package database

import (
	"context"
	"log"
)

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

// GetFromToBlocks gets a range of blocks from the database for a given height range and chain name
//
// Usage:
//
// # Used to get a range of blocks from the database for a given height range and chain name
//
// Args:
//   - fromHeight: the starting height of the block
//   - toHeight: the ending height of the block (inclusive)
//   - chainName: the name of the chain
//
// Returns:
//   - []*BlockData: the range of block data
//   - error: if the query fails
func (t *TimescaleDb) GetFromToBlocks(fromHeight uint64, toHeight uint64, chainName string) ([]*BlockData, error) {
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
	rows, err := t.pool.Query(context.Background(), query, fromHeight, toHeight, chainName)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	blocks := make([]*BlockData, 0)
	for rows.Next() {
		var block BlockData
		err := rows.Scan(&block.Hash, &block.Height, &block.Timestamp, &block.ChainID, &block.Txs)
		if err != nil {
			return nil, err
		}
		blocks = append(blocks, &block)
	}
	return blocks, nil
}

func (t *TimescaleDb) GetBankSend(txHash string, chainName string) (*BankSend, error) {
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
	row := t.pool.QueryRow(context.Background(), query, txHash, chainName)
	var bankSend BankSend
	err := row.Scan(&bankSend.TxHash, &bankSend.Timestamp, &bankSend.FromAddress, &bankSend.ToAddress, &bankSend.Amount, &bankSend.Signers)
	if err != nil {
		return nil, err
	}
	return &bankSend, nil
}

func (t *TimescaleDb) GetMsgCall(txHash string, chainName string) (*MsgCall, error) {
	query := `
	SELECT 
	encode(vmc.tx_hash, 'base64') AS tx_hash,
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
	row := t.pool.QueryRow(context.Background(), query, txHash, chainName)
	var msgCall MsgCall
	err := row.Scan(&msgCall.TxHash, &msgCall.Timestamp, &msgCall.Caller, &msgCall.PkgPath, &msgCall.FuncName, &msgCall.Args, &msgCall.Send, &msgCall.MaxDeposit, &msgCall.Signers)
	if err != nil {
		return nil, err
	}
	return &msgCall, nil
}

func (t *TimescaleDb) GetMsgAddPackage(txHash string, chainName string) (*MsgAddPackage, error) {
	query := `
	SELECT 
	encode(vmap.tx_hash, 'base64') AS tx_hash,
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
	row := t.pool.QueryRow(context.Background(), query, txHash, chainName)
	var msgAddPackage MsgAddPackage
	err := row.Scan(&msgAddPackage.TxHash, &msgAddPackage.Timestamp, &msgAddPackage.Creator, &msgAddPackage.PkgPath, &msgAddPackage.PkgName, &msgAddPackage.PkgFileNames, &msgAddPackage.Send, &msgAddPackage.MaxDeposit, &msgAddPackage.Signers)
	if err != nil {
		return nil, err
	}
	return &msgAddPackage, nil
}

func (t *TimescaleDb) GetMsgRun(txHash string, chainName string) (*MsgRun, error) {
	query := `
	SELECT 
	encode(vmr.tx_hash, 'base64') AS tx_hash,
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
	row := t.pool.QueryRow(context.Background(), query, txHash, chainName)
	var msgRun MsgRun
	err := row.Scan(&msgRun.TxHash, &msgRun.Timestamp, &msgRun.Caller, &msgRun.PkgPath, &msgRun.PkgName, &msgRun.PkgFileNames, &msgRun.Send, &msgRun.MaxDeposit, &msgRun.Signers)
	if err != nil {
		log.Println("error getting msg run", err)
		return nil, err
	}
	return &msgRun, nil
}

func (t *TimescaleDb) GetTransaction(txHash string, chainName string) (*Transaction, error) {
	var messageType []string
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
	row := t.pool.QueryRow(context.Background(), query, txHash, chainName)
	var transaction Transaction
	err := row.Scan(
		&transaction.TxHash,
		&transaction.Timestamp,
		&transaction.BlockHeight,
		&transaction.TxEvents,
		&transaction.GasUsed,
		&transaction.GasWanted,
		&transaction.Fee,
		&messageType,
	)
	if err != nil {
		log.Println("error getting transaction", err)
		return nil, err
	}
	return &transaction, nil
}

func (t *TimescaleDb) GetMsgType(txHash string, chainName string) (string, error) {
	query := `
	SELECT msg_types
	FROM transaction_general
	WHERE tx_hash = decode($1, 'base64')
	AND chain_name = $2
	`
	row := t.pool.QueryRow(context.Background(), query, txHash, chainName)
	var msgType []string
	// to future me
	// in the events the transactions can harbor more transaction types this
	// will not work, for now I only have seen with one message type per transaction
	// if this happens at least throw some log warning
	if len(msgType) >= 2 {
		log.Println("warning: transaction has more than one message type", msgType)
		return msgType[0], nil
	}
	err := row.Scan(&msgType)
	if err != nil {
		return "", err
	}
	return msgType[0], nil
}
