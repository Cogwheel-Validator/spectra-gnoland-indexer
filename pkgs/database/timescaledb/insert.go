package timescaledb

import (
	"context"
	"errors"
	"fmt"

	s "github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/schema"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
)

// InsertRows inserts a homogeneous batch of rows into the database using COPY FROM.
// Every element must belong to the same table; the table name and column list are
// read from the first element. CopyRow supplies each row's values in column order
// (kept aligned with TableColumns by TestCopyRowMatchesColumns in pkgs/schema).
//
// A failed COPY is atomic, so nothing from the batch was saved. If the failure is a
// unique violation (a chunk replayed after a partial failure), the rows are retried
// one at a time and the ones that already exist are skipped. Any other error is
// returned as is.
func (t *TimescaleDb) InsertRows(ctx context.Context, rows []s.Insertable) error {
	if len(rows) == 0 {
		return nil
	}

	err := t.copyRows(ctx, rows)
	if !isUniqueViolation(err) {
		return err
	}

	for i := range rows {
		if err := t.copyRows(ctx, rows[i:i+1]); err != nil && !isUniqueViolation(err) {
			return fmt.Errorf("row %d of %d: %w", i+1, len(rows), err)
		}
	}
	return nil
}

func (t *TimescaleDb) copyRows(ctx context.Context, rows []s.Insertable) error {
	pgxSlice := pgx.CopyFromSlice(len(rows), func(i int) ([]any, error) {
		return rows[i].CopyRow(), nil
	})

	_, err := t.pool.CopyFrom(ctx, pgx.Identifier{rows[0].TableName()}, rows[0].TableColumns(), pgxSlice)
	return err
}

// isUniqueViolation reports whether err is a Postgres unique_violation (SQLSTATE 23505).
func isUniqueViolation(err error) bool {
	var pgErr *pgconn.PgError
	return errors.As(err, &pgErr) && pgErr.Code == "23505"
}
