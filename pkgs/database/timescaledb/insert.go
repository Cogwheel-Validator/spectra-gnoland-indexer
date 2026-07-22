package timescaledb

import (
	"context"

	s "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/schema"
	"github.com/jackc/pgx/v5"
)

// InsertAddresses inserts a slice of addresses into the database using COPY FROM.
func (t *TimescaleDb) InsertAddresses(
	ctx context.Context,
	addresses []string,
	chainName string,
	insertValidators bool,
) error {
	column_names := []string{"address", "chain_name"}
	var table_name string
	if insertValidators {
		table_name = "gno_validators"
	} else {
		table_name = "gno_addresses"
	}
	pgxSlice := pgx.CopyFromSlice(len(addresses), func(i int) ([]any, error) {
		return []any{addresses[i], chainName}, nil
	})
	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{table_name}, column_names, pgxSlice)
	return err
}

// InsertRows inserts a homogeneous batch of rows into the database using COPY FROM.
// Every element must belong to the same table; the table name and column list are
// read from the first element. CopyRow supplies each row's values in column order
// (kept aligned with TableColumns by TestCopyRowMatchesColumns in pkgs/schema).
func (t *TimescaleDb) InsertRows(ctx context.Context, rows []s.Insertable) error {
	if len(rows) == 0 {
		return nil
	}

	pgxSlice := pgx.CopyFromSlice(len(rows), func(i int) ([]any, error) {
		return rows[i].CopyRow(), nil
	})

	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{rows[0].TableName()}, rows[0].TableColumns(), pgxSlice)
	return err
}
