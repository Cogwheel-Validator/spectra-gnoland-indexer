package timescaledb

import (
	"context"

	s "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/schema"
	"github.com/jackc/pgx/v5"
)

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
