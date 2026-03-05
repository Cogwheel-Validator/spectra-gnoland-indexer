package database

import (
	"context"
	"encoding/base64"
	"log"

	dictloader "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/dict_loader"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/events_proto"
	"github.com/cosmos/gogoproto/proto"
	"github.com/klauspost/compress/zstd"
)

var dictBytes = dictloader.LoadDict()
var zstdDict = zstd.WithDecoderDicts(dictBytes)
var zstdReader, _ = zstd.NewReader(nil, zstdDict)

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
	tx.tx_events_compressed,
	tx.compression_on,
	tx.gas_used,
	tx.gas_wanted,
	tx.fee,
	tx.msg_types
	FROM transaction_general tx
	WHERE tx.tx_hash = decode($1, 'base64')
	AND tx.chain_name = $2
	`
	row := t.pool.QueryRow(ctx, query, txHash, chainName)
	var transaction FullTxData
	err := row.Scan(
		&transaction.TxHash,
		&transaction.Timestamp,
		&transaction.BlockHeight,
		&transaction.TxEvents,
		&transaction.TxEventsCompressed,
		&transaction.CompressionOn,
		&transaction.GasUsed,
		&transaction.GasWanted,
		&transaction.Fee,
		&transaction.MsgTypes,
	)
	if err != nil {
		log.Println("error getting transaction", err)
		return nil, err
	}
	tx, err := transaction.ToTransaction(decodeEvents)
	if err != nil {
		return nil, err
	}
	return tx, nil
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
	tx.tx_events_compressed,
	tx.compression_on,
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
		transaction := &FullTxData{}
		err := rows.Scan(
			&transaction.TxHash,
			&transaction.Timestamp,
			&transaction.BlockHeight,
			&transaction.TxEvents,
			&transaction.TxEventsCompressed,
			&transaction.CompressionOn,
			&transaction.GasUsed,
			&transaction.GasWanted,
			&transaction.Fee,
			&transaction.MsgTypes)
		if err != nil {
			return nil, err
		}
		tx, err := transaction.ToTransaction(decodeEvents)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return transactions, nil
}

// GetTransactionsByOffset gets the transactions by offset for a given chain name
//
// Usage:
//
// # Used to get the transactions by offset for a given chain name
//
// Parameters:
//   - chainName: the name of the chain
//   - limit: the limit of the transactions to get
//   - offset: the offset of the transactions to get
func (t *TimescaleDb) GetTransactionsByOffset(
	ctx context.Context,
	chainName string,
	limit uint64,
	offset uint64,
) ([]*Transaction, error) {
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
	LIMIT $2 OFFSET $3
	`
	rows, err := t.pool.Query(ctx, query, chainName, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	transactions := make([]*Transaction, 0)
	for rows.Next() {
		transaction := &FullTxData{}
		err := rows.Scan(&transaction.TxHash, &transaction.Timestamp, &transaction.BlockHeight, &transaction.TxEvents, &transaction.GasUsed, &transaction.GasWanted, &transaction.Fee, &transaction.MsgTypes)
		if err != nil {
			return nil, err
		}
		tx, err := transaction.ToTransaction(decodeEvents)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return transactions, nil
}

func (t *TimescaleDb) GetTransactionsByCursor(
	ctx context.Context,
	chainName string,
	cursor string,
	limit uint64,
) ([]*Transaction, error) {
	var query string
	// if txHash and timestamp are nil make same query as GetLastXTransactions
	if cursor == "" {
		return t.GetLastXTransactions(ctx, chainName, limit)
	}
	timestamp, txHash, err := unmarshalCursorParam(cursor)
	if err != nil {
		return nil, err
	}

	decodedTxHash, err := base64.URLEncoding.Strict().DecodeString(txHash)
	if err != nil {
		return nil, err
	}
	query = `
	SELECT
	encode(tx.tx_hash, 'base64') AS tx_hash,
	tx.timestamp,
	tx.block_height,
	tx.tx_events,
	tx.tx_events_compressed,
	tx.compression_on,
	tx.gas_used,
	tx.gas_wanted,
	tx.fee,
	tx.msg_types
	FROM transaction_general tx
	WHERE tx.chain_name = $1
	AND (tx.timestamp, tx.tx_hash) < ($2, $3)
	ORDER BY tx.timestamp DESC, tx.tx_hash DESC
	LIMIT $4
	`
	args := []any{chainName, timestamp, decodedTxHash, limit}
	rows, err := t.pool.Query(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	transactions := make([]*Transaction, 0)
	for rows.Next() {
		transaction := &FullTxData{}
		err := rows.Scan(
			&transaction.TxHash,
			&transaction.Timestamp,
			&transaction.BlockHeight,
			&transaction.TxEvents,
			&transaction.TxEventsCompressed,
			&transaction.CompressionOn,
			&transaction.GasUsed,
			&transaction.GasWanted,
			&transaction.Fee,
			&transaction.MsgTypes,
		)
		if err != nil {
			return nil, err
		}
		tx, err := transaction.ToTransaction(decodeEvents)
		if err != nil {
			return nil, err
		}
		transactions = append(transactions, tx)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return transactions, nil
}

func decompressEvents(txEvents []byte) ([]byte, error) {
	decompressed, err := zstdReader.DecodeAll(txEvents, nil)
	if err != nil {
		return nil, err
	}
	return decompressed, nil
}

func protoUnmarshal(rawData []byte) (*events_proto.TxEvents, error) {
	txEvents := &events_proto.TxEvents{}
	err := proto.Unmarshal(rawData, txEvents)
	if err != nil {
		return nil, err
	}
	return txEvents, nil
}

func decodeEvents(txEvents []byte) (*[]Event, error) {
	if len(txEvents) == 0 {
		return &[]Event{}, nil
	}
	decompressed, err := decompressEvents(txEvents)
	if err != nil {
		return nil, err
	}
	txEventsProto, err := protoUnmarshal(decompressed)
	if err != nil {
		return nil, err
	}
	events := make([]Event, 0, len(txEventsProto.Events))
	for _, event := range txEventsProto.Events {
		attributes := make([]Attribute, 0, len(event.Attributes))
		for _, attribute := range event.Attributes {
			attributes = append(attributes, Attribute{
				Key:   attribute.Key,
				Value: attribute.GetStringValue(),
			})
		}
		events = append(events, Event{
			AtType:     event.AtType,
			Type:       event.Type,
			Attributes: attributes,
			PkgPath:    *event.PkgPath,
		})
	}
	return &events, nil
}
