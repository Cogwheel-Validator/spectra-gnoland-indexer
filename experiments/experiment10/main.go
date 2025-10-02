package main

import (
	"context"
	"encoding/base64"
	"fmt"
	"log"
	"math/big"
	"time"

	dataTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

func setupConnection() (*pgxpool.Pool, error) {
	host := "localhost"
	port := 6543
	user := "postgres"
	password := "12345678"
	dbname := "gnoland"
	sslmode := "disable"

	config, err := pgxpool.ParseConfig(
		fmt.Sprintf("host=%s port=%d user=%s password=%s dbname=%s sslmode=%s",
			host, port, user, password, dbname, sslmode),
	)
	if err != nil {
		return nil, err
	}

	config.AfterConnect = func(ctx context.Context, conn *pgx.Conn) error {
		dataTypeNames := dataTypes.CustomTypeNames()

		for _, typeName := range dataTypeNames {
			dataType, err := conn.LoadType(ctx, typeName)
			if err != nil {
				return err
			}
			conn.TypeMap().RegisterType(dataType)
		}

		return nil
	}

	return pgxpool.NewWithConfig(context.Background(), config)
}

func makePgxArray[T any](v []T) pgtype.Array[T] {
	if v == nil {
		return pgtype.Array[T]{Valid: false}
	}

	return pgtype.Array[T]{
		Elements: v,
		Dims:     []pgtype.ArrayDimension{{Length: int32(len(v)), LowerBound: 1}},
		Valid:    true,
	}
}

func main() {
	// let's try to insert real data

	// block
	hash1, err := base64.StdEncoding.DecodeString("FGYVJEJQlox8AMRKnhClsIUa/TrFmWFr9SmNVJDLSIg=")
	if err != nil {
		log.Fatalf("Error decoding the hash1: %s", err)
	}
	hash2, err := base64.StdEncoding.DecodeString("QcG+qinbhad+1OTIch2H6tR/cYnKEEh4MXInmGDv940=")
	if err != nil {
		log.Fatalf("Error decoding the hash1: %s", err)
	}
	timestamp1, err := time.Parse(time.RFC3339Nano, "2025-09-05T07:30:58.618646057Z")
	if err != nil {
		log.Fatalf("Error parsing the timestamp1, %s", err)
	}
	timestamp2, err := time.Parse(time.RFC3339Nano, "2025-09-05T07:30:53.054040872Z")
	if err != nil {
		log.Fatalf("Error parsing the timestamp2, %s", err)
	}
	txHash1, err := base64.StdEncoding.DecodeString("gupsX4SH279MJ2xuosYlRwhHDKv/s9xco7BWnbBPTmE=")
	if err != nil {
		log.Fatalf("Error decoding the txHash1, %s", err)
	}
	txHash2, err := base64.StdEncoding.DecodeString("KClwYV7356OEfzdvahXE3/d+pRUPthPxo8Cy33R8E2o=")
	if err != nil {
		log.Fatalf("Error decoding the txHash2, %s", err)
	}
	blocks := []dataTypes.Blocks{
		{
			Hash:      hash1,
			Height:    970866,
			Timestamp: timestamp1,
			ChainID:   "gnoalnd-testnet-7",
			Txs:       [][]byte{txHash1, txHash2},
			ChainName: "gnoland",
		},
		{
			Hash:      hash2,
			Height:    970865,
			Timestamp: timestamp2,
			ChainID:   "gnoalnd-testnet-7",
			Txs:       nil, // empty
			ChainName: "gnoland",
		},
	}
	transactionGeneral := []dataTypes.TransactionGeneral{
		{
			TxHash:             txHash1,
			ChainName:          "gnoland",
			Timestamp:          timestamp1,
			MsgTypes:           []string{"bank_msg_send"},
			TxEvents:           []dataTypes.Event{}, // null
			TxEventsCompressed: nil,
			CompressionOn:      false,
			GasUsed:            100000,
			GasWanted:          90000,
			Fee: dataTypes.Amount{
				Amount: pgtype.Numeric{Int: big.NewInt(1000000)},
				Denom:  "ugnot",
			},
		},
		{
			TxHash:    txHash2,
			ChainName: "gnoland",
			Timestamp: timestamp2,
			MsgTypes:  []string{"vm_msg_call"},
			TxEvents: []dataTypes.Event{
				{
					// it is fake but just to test it out
					AtType: "tm.Gno.MsgCall",
					Type:   "vm_msg_call",
					Attributes: []dataTypes.Attribute{
						{Key: "caller", Value: "gno1234567890"},
						{Key: "func_name", Value: "test_func"},
					},
					PkgPath: "gno.land/r/tests",
				},
			},
			TxEventsCompressed: nil,
			CompressionOn:      false,
			GasUsed:            100000,
			GasWanted:          90000,
			Fee: dataTypes.Amount{
				Amount: pgtype.Numeric{Int: big.NewInt(1000000)},
				Denom:  "ugnot",
			},
		},
	}
	// treat it like an array because that is how the indexer should do it even for one message
	bankMsgSend := []dataTypes.MsgSend{
		{
			TxHash:      txHash1,
			Timestamp:   timestamp1,
			ChainName:   "gnoland",
			FromAddress: 1,
			ToAddress:   2,
			Amount: []dataTypes.Amount{
				{
					Amount: pgtype.Numeric{Int: big.NewInt(1000000)},
					Denom:  "ugnot",
				},
			},
			Signers: []int32{1, 2},
		},
	}
	vmMsgCall := []dataTypes.MsgCall{
		{
			TxHash:    txHash2,
			Timestamp: timestamp2,
			ChainName: "gnoland",
			Caller:    1,
			PkgPath:   "gno.land/r/tests",
			FuncName:  "test_func",
			Args:      "test_args",
			MaxDeposit: []dataTypes.Amount{
				{
					Amount: pgtype.Numeric{Int: big.NewInt(1000000)},
					Denom:  "ugnot",
				},
			},
			Send: []dataTypes.Amount{
				{
					Amount: pgtype.Numeric{Int: big.NewInt(1000000)},
					Denom:  "ugnot",
				},
			},
			Signers: []int32{1, 2},
		},
	}

	// let's insert the data into the database
	pool, err := setupConnection()
	if err != nil {
		panic(err)
	}
	defer pool.Close()

	ctx := context.Background()

	// insert the blocks
	pgxSlice := pgx.CopyFromSlice(len(blocks), func(i int) ([]any, error) {
		return []any{
			blocks[i].Hash,
			blocks[i].Height,
			blocks[i].Timestamp,
			blocks[i].ChainID,
			makePgxArray(blocks[i].Txs),
			blocks[i].ChainName,
		}, nil
	})
	_, err = pool.CopyFrom(ctx, pgx.Identifier{"blocks"}, []string{"hash", "height", "timestamp", "chain_id", "proposer_address", "txs", "chain_name"}, pgxSlice)
	if err != nil {
		panic(err)
	}

	// insert the transaction general
	pgxSlice = pgx.CopyFromSlice(len(transactionGeneral), func(i int) ([]any, error) {
		return []any{
			transactionGeneral[i].TxHash,
			transactionGeneral[i].ChainName,
			transactionGeneral[i].Timestamp,
			makePgxArray(transactionGeneral[i].MsgTypes),
			transactionGeneral[i].TxEvents,
			transactionGeneral[i].TxEventsCompressed,
			transactionGeneral[i].CompressionOn,
			transactionGeneral[i].GasUsed,
			transactionGeneral[i].GasWanted,
			transactionGeneral[i].Fee,
		}, nil
	})
	_, err = pool.CopyFrom(ctx, pgx.Identifier{"transaction_general"}, []string{
		"tx_hash", "chain_name", "timestamp", "msg_types", "tx_events", "tx_events_compressed", "compression_on", "gas_used", "gas_wanted", "fee"}, pgxSlice)
	if err != nil {
		panic(err)
	}

	// insert the bank msg send
	pgxSlice = pgx.CopyFromSlice(len(bankMsgSend), func(i int) ([]any, error) {
		return []any{bankMsgSend[i].TxHash, bankMsgSend[i].ChainName, bankMsgSend[i].FromAddress, bankMsgSend[i].ToAddress, bankMsgSend[i].Amount, bankMsgSend[i].Timestamp, makePgxArray(bankMsgSend[i].Signers)}, nil
	})
	_, err = pool.CopyFrom(ctx, pgx.Identifier{"bank_msg_send"}, []string{"tx_hash", "chain_name", "from_address", "to_address", "amount", "timestamp", "signers"}, pgxSlice)
	if err != nil {
		panic(err)
	}

	// insert the vm msg call
	pgxSlice = pgx.CopyFromSlice(len(vmMsgCall), func(i int) ([]any, error) {
		return []any{vmMsgCall[i].TxHash, vmMsgCall[i].ChainName, vmMsgCall[i].Caller, vmMsgCall[i].PkgPath, vmMsgCall[i].FuncName, vmMsgCall[i].Args, vmMsgCall[i].Send, vmMsgCall[i].MaxDeposit, vmMsgCall[i].Timestamp, makePgxArray(vmMsgCall[i].Signers)}, nil
	})
	_, err = pool.CopyFrom(ctx, pgx.Identifier{"vm_msg_call"}, []string{"tx_hash", "chain_name", "caller", "pkg_path", "func_name", "args", "send", "max_deposit", "timestamp", "signers"}, pgxSlice)
	if err != nil {
		panic(err)
	}

}
