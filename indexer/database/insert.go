package database

import (
	"context"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/sql_data_types"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
)

// InsertAddresses inserts a slice of addresses into the database
//
// This is a method to insert a slice of addresses into the database
// It should preform better than using INSERT INTO... for a large number of addresses
// because it uses the COPY FROM command
//
// Usage:
//
// # Used inside of the address cache package to insert the addresses to the database
//
// Parameters:
//   - addresses: a slice of addresses to insert
//   - chainName: the name of the chain to insert the addresses to
//   - insertValidators: a boolean to indicate if the addresses are validators or accounts
//
// Returns:
//   - error: an error if the insertion fails
func (t *TimescaleDb) InsertAddresses(addresses []string, chainName string, insertValidators bool) error {
	column_names := []string{"address", "chain_name"}
	var table_name string
	if insertValidators {
		table_name = "gno_validators"
	} else {
		table_name = "gno_addresses"
	}
	// create interface to copy from slice to the db
	pgxSlice := pgx.CopyFromSlice(len(addresses), func(i int) ([]any, error) {
		return []any{addresses[i], chainName}, nil
	})
	// copy the addresses to the db
	_, err := t.pool.CopyFrom(context.Background(), pgx.Identifier{table_name}, column_names, pgxSlice)
	return err
}

// InsertBlocks inserts a slice of blocks into the database using pgx copy function
// it will create the copy from slice to the db and then insert it to the database
//
// Usage:
//
// # Used for inserting a large number of blocks to the database
//
// Args:
//   - blocks: a slice of blocks to insert
//
// Returns:
//   - error: an error if the insertion fails
func (t *TimescaleDb) InsertBlocks(blocks []sql_data_types.Blocks) error {
	// create a copy from slice to the db
	pgxSlice := pgx.CopyFromSlice(len(blocks), func(i int) ([]any, error) {
		return []any{
			blocks[i].Hash,
			blocks[i].Height,
			blocks[i].Timestamp,
			blocks[i].ChainID,
			blocks[i].ProposerAddress,
			makePgxArray(blocks[i].Txs),
			blocks[i].ChainName}, nil
	})

	// mark the columns to be inserted
	columns := blocks[0].TableColumns()

	// insert the data to the db
	_, err := t.pool.CopyFrom(context.Background(), pgx.Identifier{"blocks"}, columns, pgxSlice)
	return err
}

// InsertValidatorBlockSignings inserts a slice of validator block signings into the database using pgx copy function
// it will create the copy from slice to the db and then insert it to the database
//
// Usage:
//
// # Used for inserting a large number of validator block signings to the database
//
// Args:
//   - validatorBlockSigning: a slice of validator block signings to insert
//
// Returns:
//   - error: an error if the insertion fails
func (t *TimescaleDb) InsertValidatorBlockSignings(validatorBlockSigning []sql_data_types.ValidatorBlockSigning) error {
	// create a copy from slice to the db
	pgxSlice := pgx.CopyFromSlice(len(validatorBlockSigning), func(i int) ([]any, error) {
		return []any{
			validatorBlockSigning[i].BlockHeight,
			validatorBlockSigning[i].Timestamp,
			makePgxArray(validatorBlockSigning[i].SignedVals),
			validatorBlockSigning[i].ChainName}, nil
	})

	// mark the columns to be inserted
	columns := validatorBlockSigning[0].TableColumns()

	// insert the data to the db
	_, err := t.pool.CopyFrom(context.Background(), pgx.Identifier{"validator_block_signing"}, columns, pgxSlice)
	return err
}

// InsertTransactionsGeneral inserts a slice of transaction general data into the database using pgx copy function
// it will create the copy from slice to the db and then insert it to the database
//
// Usage:
//
// # Used for inserting a large number of transaction general data to the database
//
// Args:
//   - transactionsGeneral: a slice of transaction general data to insert
//
// Returns:
//   - error: an error if the insertion fails
func (t *TimescaleDb) InsertTransactionsGeneral(transactionsGeneral []sql_data_types.TransactionGeneral) error {
	// create a copy from the slice
	pgxSlice := pgx.CopyFromSlice(len(transactionsGeneral), func(i int) ([]any, error) {
		return []any{
			transactionsGeneral[i].TxHash,
			transactionsGeneral[i].ChainName,
			transactionsGeneral[i].Timestamp,
			makePgxArray(transactionsGeneral[i].MsgTypes),
			makePgxArray(transactionsGeneral[i].TxEvents),
			transactionsGeneral[i].TxEventsCompressed,
			transactionsGeneral[i].CompressionOn,
			transactionsGeneral[i].GasUsed,
			transactionsGeneral[i].GasWanted,
			transactionsGeneral[i].Fee,
		}, nil
	})

	// mark the columns to be inserted
	columns := transactionsGeneral[0].TableColumns()

	// insert the data to the db
	_, err := t.pool.CopyFrom(context.Background(), pgx.Identifier{"transaction_general"}, columns, pgxSlice)
	return err
}

// InsertMsgSend inserts a slice of MsgSend messages into the database
//
// Args:
//   - messages: a slice of MsgSend messages to insert
//
// Returns:
//   - error: an error if the insertion fails
func (t *TimescaleDb) InsertMsgSend(messages []sql_data_types.MsgSend) error {
	pgxSlice := pgx.CopyFromSlice(len(messages), func(i int) ([]any, error) {
		return []any{
			messages[i].TxHash,
			messages[i].ChainName,
			messages[i].FromAddress,
			messages[i].ToAddress,
			makePgxArray(messages[i].Amount),
			makePgxArray(messages[i].Signers),
			messages[i].Timestamp,
		}, nil
	})

	columns := messages[0].TableColumns()
	_, err := t.pool.CopyFrom(context.Background(), pgx.Identifier{"bank_msg_send"}, columns, pgxSlice)
	return err
}

// InsertMsgCall inserts a slice of MsgCall messages into the database
//
// Args:
//   - messages: a slice of MsgCall messages to insert
//
// Returns:
//   - error: an error if the insertion fails
func (t *TimescaleDb) InsertMsgCall(messages []sql_data_types.MsgCall) error {
	pgxSlice := pgx.CopyFromSlice(len(messages), func(i int) ([]any, error) {
		return []any{
			messages[i].TxHash,
			messages[i].Timestamp,
			messages[i].ChainName,
			messages[i].Caller,
			messages[i].PkgPath,
			messages[i].FuncName,
			messages[i].Args,
			makePgxArray(messages[i].Send),
			makePgxArray(messages[i].MaxDeposit),
			makePgxArray(messages[i].Signers),
		}, nil
	})

	columns := messages[0].TableColumns()
	_, err := t.pool.CopyFrom(context.Background(), pgx.Identifier{"vm_msg_call"}, columns, pgxSlice)
	return err
}

// InsertMsgAddPackage inserts a slice of MsgAddPackage messages into the database
//
// Args:
//   - messages: a slice of MsgAddPackage messages to insert
//
// Returns:
//   - error: an error if the insertion fails
func (t *TimescaleDb) InsertMsgAddPackage(messages []sql_data_types.MsgAddPackage) error {
	pgxSlice := pgx.CopyFromSlice(len(messages), func(i int) ([]any, error) {
		return []any{
			messages[i].TxHash,
			messages[i].Timestamp,
			messages[i].ChainName,
			messages[i].Creator,
			messages[i].PkgPath,
			messages[i].PkgName,
			makePgxArray(messages[i].PkgFileNames),
			makePgxArray(messages[i].Send),
			makePgxArray(messages[i].MaxDeposit),
			makePgxArray(messages[i].Signers),
		}, nil
	})

	columns := messages[0].TableColumns()
	_, err := t.pool.CopyFrom(context.Background(), pgx.Identifier{"vm_msg_add_package"}, columns, pgxSlice)
	return err
}

// InsertMsgRun inserts a slice of MsgRun messages into the database
//
// Args:
//   - messages: a slice of MsgRun messages to insert
//
// Returns:
//   - error: an error if the insertion fails
func (t *TimescaleDb) InsertMsgRun(messages []sql_data_types.MsgRun) error {
	pgxSlice := pgx.CopyFromSlice(len(messages), func(i int) ([]any, error) {
		return []any{
			messages[i].TxHash,
			messages[i].Timestamp,
			messages[i].ChainName,
			messages[i].Caller,
			messages[i].PkgPath,
			messages[i].PkgName,
			makePgxArray(messages[i].PkgFileNames),
			makePgxArray(messages[i].Send),
			makePgxArray(messages[i].MaxDeposit),
			makePgxArray(messages[i].Signers),
		}, nil
	})

	columns := messages[0].TableColumns()
	_, err := t.pool.CopyFrom(context.Background(), pgx.Identifier{"vm_msg_run"}, columns, pgxSlice)
	return err
}

// makePgxArray is a helper generic function to create a pgx array from a slice
//
// In theory it should be similar to pq.Array i think, it should be used for the some composite types and
// bytearrays but to be sure it should be usable on any type that is supposed to be inserted into the database as
// an array
//
// Args:
//   - v: a slice of any type
//
// Returns:
//   - pgtype.Array[T]: a pgx array
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
