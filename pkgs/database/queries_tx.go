package database

import (
	"context"
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

/*504568-504568

 hypertable_schema |     hypertable_name     | total_chunks | total_hypertable_size |                                    chunk_sizes
-------------------+-------------------------+--------------+-----------------------+-----------------------------------------------------------------------------------
 public            | blocks                  |            3 | 88 MB                 | _hyper_1_1_chunk: 30 MB, _hyper_1_9_chunk: 32 MB, _hyper_1_16_chunk: 26 MB
 public            | validator_block_signing |            3 | 86 MB                 | _hyper_3_2_chunk: 28 MB, _hyper_3_10_chunk: 32 MB, _hyper_3_17_chunk: 26 MB
 public            | vm_msg_call             |            3 | 632 kB                | _hyper_11_6_chunk: 128 kB, _hyper_11_14_chunk: 352 kB, _hyper_11_20_chunk: 152 kB
 public            | transaction_general     |            3 | 608 kB                | _hyper_7_3_chunk: 152 kB, _hyper_7_11_chunk: 288 kB, _hyper_7_18_chunk: 168 kB
 public            | address_tx              |            3 | 432 kB                | _hyper_5_4_chunk: 112 kB, _hyper_5_12_chunk: 224 kB, _hyper_5_19_chunk: 96 kB
 public            | bank_msg_send           |            3 | 328 kB                | _hyper_9_5_chunk: 96 kB, _hyper_9_13_chunk: 144 kB, _hyper_9_21_chunk: 88 kB
 public            | vm_msg_add_package      |            3 | 144 kB                | _hyper_13_8_chunk: 48 kB, _hyper_13_15_chunk: 48 kB, _hyper_13_22_chunk: 48 kB
 public            | vm_msg_run              |            2 | 96 kB                 | _hyper_15_7_chunk: 48 kB, _hyper_15_23_chunk: 48 kB

*/
