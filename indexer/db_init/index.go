package dbinit

import (
	"context"
	"fmt"
	"strings"
)

// IndexDef describes a single CREATE INDEX statement.
//
// This is intentionally hand-written rather than derived via reflection:
// indexes can be unique, partial, per-chunk, or ordered in ways that don't
// map onto a single struct field, and column order matters for which
// queries the index can serve. Declare one IndexDef per index you actually
// need, next to the table it belongs to.
//
// Rules to follow when declaring indexes:
//
// 1. If the table is regular PostgreSQL table, use of PerChunk is not supported.
// 2. If it is a hypertable do not use Unique or Concurrently, you have to use explicitly
// the WITH (timescaledb.transaction_per_chunk) option, which PerChunk should do.
type IndexDef struct {
	// Name is the index name. Required, must be unique per database.
	Name string
	// Table is the table (or hypertable) the index is created on.
	Table string
	// Columns are the indexed columns/expressions in order, e.g.
	// []string{"chain_name", "block_height DESC"}.
	Columns []string
	// Unique marks the index as UNIQUE.
	Unique bool
	// Concurrently builds the index with CREATE INDEX CONCURRENTLY, which
	// avoids blocking writes but cannot run inside a transaction block and
	// cannot be batched with other statements in the same Exec call.
	// Use this when adding an index to a table that already has live data
	// (e.g. retrofitting in production). Not needed on a freshly created,
	// empty table.
	Concurrently bool
	// PerChunk adds WITH (timescaledb.transaction_per_chunk), which is
	// required when creating an index on a hypertable that already has
	// chunks, since a single CREATE INDEX would otherwise try to take a
	// lock across every chunk in one transaction.
	PerChunk bool
	// Where adds a WHERE clause for a partial index. Optional.
	Where string
}

// GenerateCreateIndexSQL builds the CREATE INDEX statement for an IndexDef.
func GenerateCreateIndexSQL(idx IndexDef) (string, error) {
	if err := validateIndexDef(idx); err != nil {
		return "", err
	}
	var sb strings.Builder

	sb.WriteString("CREATE ")
	if idx.Unique {
		sb.WriteString("UNIQUE ")
	}
	sb.WriteString("INDEX ")
	if idx.Concurrently {
		sb.WriteString("CONCURRENTLY ")
	}
	sb.WriteString("IF NOT EXISTS ")
	sb.WriteString(idx.Name)
	sb.WriteString(" ON ")
	sb.WriteString(idx.Table)
	sb.WriteString(" (")
	sb.WriteString(strings.Join(idx.Columns, ", "))
	sb.WriteString(")")

	if idx.PerChunk {
		sb.WriteString(" WITH (timescaledb.transaction_per_chunk)")
	}

	if idx.Where != "" {
		sb.WriteString(" WHERE ")
		sb.WriteString(idx.Where)
	}

	sb.WriteString(";")

	return sb.String(), nil
}

func validateIndexDef(idx IndexDef) error {
	if idx.PerChunk && idx.Concurrently {
		return fmt.Errorf("PerChunk and Concurrently cannot be set at the same time")
	}

	// According to the TigerData docs this is not supported for some reason.
	if idx.PerChunk && idx.Unique {
		return fmt.Errorf("PerChunk and Unique cannot be set at the same time")
	}
	return nil
}

// CreateIndex executes the CREATE INDEX statement for the given IndexDef.
//
// If idx.Concurrently is set, this must be the only statement in its Exec
// call (it is, since Exec here only ever runs one statement) and must not
// be called from inside an explicit transaction.
func (db *DBInitializer) CreateIndex(idx IndexDef) error {
	sql, err := GenerateCreateIndexSQL(idx)
	if err != nil {
		return fmt.Errorf("failed to generate create index SQL: %w", err)
	}

	_, err = db.pool.Exec(context.Background(), sql)
	if err != nil {
		return fmt.Errorf("failed to create index %s on %s: %w", idx.Name, idx.Table, err)
	}

	l.Info().Msgf("Successfully created index: %s", idx.Name)
	return nil
}
