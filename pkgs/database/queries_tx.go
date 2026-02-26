package database

import (
	"context"
	"log"
)

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
