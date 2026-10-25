package migrations

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/logger"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
)

// lockKey is the advisory lock that keeps two "migrate up" runs apart.
const lockKey int64 = 0x5350454354524101

// Same DDL "setup create-db" generates from schema.SchemaMigration. Only used
// when the table is missing, which happens on databases older than v0.10.1.
const createTableSQL = `CREATE TABLE IF NOT EXISTS schema_migrations (
    version INTEGER NOT NULL,
    description TEXT NOT NULL,
    applied_at TIMESTAMPTZ DEFAULT NOW() NOT NULL,
    success BOOLEAN NOT NULL,
    PRIMARY KEY (version, success)
);`

var (
	// ErrNotVersioned means schema_migrations is missing or empty.
	ErrNotVersioned = errors.New("database has no recorded schema version")
	// ErrPending means the database is missing migrations this binary needs.
	ErrPending = errors.New("database schema is out of date")
	// ErrTooNew means the database has migrations this binary doesn't know.
	ErrTooNew = errors.New("database schema is newer than this indexer")
)

// State is what schema_migrations holds.
type State struct {
	// Applied maps successfully applied versions to when they were applied.
	Applied map[int32]time.Time
	// Failed maps versions with a recorded failed attempt to the last attempt.
	Failed map[int32]time.Time
}

// Current returns the highest applied version, 0 when nothing is applied.
func (s State) Current() int32 {
	var v int32
	for version := range s.Applied {
		v = max(v, version)
	}
	return v
}

// Pending returns the known migrations that are not applied, in order.
func (s State) Pending() []Migration {
	var pending []Migration
	for _, m := range all {
		if _, ok := s.Applied[m.Version]; !ok {
			pending = append(pending, m)
		}
	}
	return pending
}

// Verify reports whether the database matches what this binary expects.
func (s State) Verify() error {
	if len(s.Applied) == 0 {
		return fmt.Errorf("%w: run \"indexer migrate up\" with the database owner account", ErrNotVersioned)
	}
	var unknown []int32
	for version := range s.Applied {
		if _, ok := find(version); !ok {
			unknown = append(unknown, version)
		}
	}
	if len(unknown) > 0 {
		slices.Sort(unknown)
		return fmt.Errorf("%w: database has version(s) %v, this binary knows up to %d, upgrade the indexer",
			ErrTooNew, unknown, Latest())
	}
	if pending := s.Pending(); len(pending) > 0 {
		versions := make([]int32, len(pending))
		for i, m := range pending {
			versions[i] = m.Version
		}
		return fmt.Errorf("%w: database is at version %d, this binary needs %d (missing %v): "+
			"stop the indexer and run \"indexer migrate up\" with the database owner account",
			ErrPending, s.Current(), Latest(), versions)
	}
	return nil
}

// querier is satisfied by *pgxpool.Pool, *pgxpool.Conn and pgx.Tx.
type querier interface {
	Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error)
	Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error)
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
}

// ReadState loads schema_migrations. A missing table reads as an empty state.
func ReadState(ctx context.Context, q querier) (State, error) {
	state := State{Applied: map[int32]time.Time{}, Failed: map[int32]time.Time{}}

	var exists bool
	if err := q.QueryRow(ctx, "SELECT to_regclass('schema_migrations') IS NOT NULL").Scan(&exists); err != nil {
		return state, fmt.Errorf("failed to look up schema_migrations: %w", err)
	}
	if !exists {
		return state, nil
	}

	rows, err := q.Query(ctx, "SELECT version, applied_at, success FROM schema_migrations")
	if err != nil {
		return state, fmt.Errorf("failed to read schema_migrations: %w", err)
	}
	defer rows.Close()
	for rows.Next() {
		var (
			version   int32
			appliedAt time.Time
			success   bool
		)
		if err := rows.Scan(&version, &appliedAt, &success); err != nil {
			return state, fmt.Errorf("failed to read schema_migrations: %w", err)
		}
		if success {
			state.Applied[version] = appliedAt
		} else {
			state.Failed[version] = appliedAt
		}
	}
	return state, rows.Err()
}

// Check is the startup check for "run": it fails unless every known migration
// is applied and none is unknown. Only needs SELECT on schema_migrations.
func Check(ctx context.Context, pool *pgxpool.Pool) error {
	state, err := ReadState(ctx, pool)
	if err != nil {
		return err
	}
	return state.Verify()
}

// RecordAll marks every known migration as applied. "setup create-db" calls it
// because a database it creates already has the latest schema.
func RecordAll(ctx context.Context, pool *pgxpool.Pool) error {
	return record(ctx, pool, all)
}

// Baseline marks every migration up to and including version as applied
// without running it. It is for databases whose schema is known to be at that
// version but that were never versioned.
func Baseline(ctx context.Context, pool *pgxpool.Pool, version int32) error {
	if _, ok := find(version); !ok {
		return fmt.Errorf("unknown migration version %d, latest is %d", version, Latest())
	}
	var upTo []Migration
	for _, m := range all {
		if m.Version <= version {
			upTo = append(upTo, m)
		}
	}
	return record(ctx, pool, upTo)
}

func record(ctx context.Context, q querier, ms []Migration) error {
	if _, err := q.Exec(ctx, createTableSQL); err != nil {
		return fmt.Errorf("failed to create schema_migrations: %w", err)
	}
	for _, m := range ms {
		if _, err := q.Exec(ctx,
			`INSERT INTO schema_migrations (version, description, success) VALUES ($1, $2, TRUE)
			ON CONFLICT (version, success) DO NOTHING`,
			m.Version, m.Description,
		); err != nil {
			return fmt.Errorf("failed to record migration %d: %w", m.Version, err)
		}
	}
	return nil
}

// Up applies every pending migration in order, each in its own transaction.
// It returns the migrations it applied.
//
// A database without any recorded version is recognized as the v0.10.1
// baseline when its address tables still have the chain_name column; any
// other unversioned database needs "migrate baseline" first.
func Up(ctx context.Context, pool *pgxpool.Pool) ([]Migration, error) {
	l := logger.Get()

	conn, err := pool.Acquire(ctx)
	if err != nil {
		return nil, err
	}
	defer conn.Release()

	// Session lock, so it also covers the After statements outside the transactions.
	if _, err := conn.Exec(ctx, "SELECT pg_advisory_lock($1)", lockKey); err != nil {
		return nil, fmt.Errorf("failed to take the migration lock: %w", err)
	}
	defer func() {
		if _, err := conn.Exec(context.WithoutCancel(ctx), "SELECT pg_advisory_unlock($1)", lockKey); err != nil {
			l.Warn("failed to release the migration lock", "error", err)
		}
	}()

	state, err := ReadState(ctx, conn)
	if err != nil {
		return nil, err
	}
	if len(state.Applied) == 0 {
		if err := detectBaseline(ctx, conn); err != nil {
			return nil, err
		}
		l.Info("unversioned database has the v0.10.1 schema, recording it as the baseline")
		baseline, _ := find(BaselineVersion)
		if err := record(ctx, conn, []Migration{baseline}); err != nil {
			return nil, err
		}
		state.Applied[BaselineVersion] = time.Now()
	}
	if err := state.Verify(); err != nil && !errors.Is(err, ErrPending) {
		return nil, err
	}

	var applied []Migration
	for _, m := range state.Pending() {
		l.Info("applying migration", "version", m.Version, "release", m.Release, "description", m.Description)
		start := time.Now()
		if err := apply(ctx, conn, m); err != nil {
			return applied, err
		}
		l.Info("migration applied", "version", m.Version, "took", time.Since(start).Round(time.Millisecond))
		applied = append(applied, m)

		for _, stmt := range m.After {
			l.Info("running post-migration statement", "version", m.Version, "sql", stmt)
			if _, err := conn.Exec(ctx, stmt); err != nil {
				l.Warn("post-migration statement failed, the schema is migrated but you may want to rerun it",
					"version", m.Version, "sql", stmt, "error", err)
			}
		}
	}
	return applied, nil
}

// apply runs one migration and records it in the same transaction. On failure
// the attempt is recorded with success = false.
func apply(ctx context.Context, conn *pgxpool.Conn, m Migration) error {
	err := pgx.BeginFunc(ctx, conn, func(tx pgx.Tx) error {
		if m.SQL != "" {
			if _, err := tx.Exec(ctx, m.SQL); err != nil {
				return err
			}
		}
		_, err := tx.Exec(ctx,
			"INSERT INTO schema_migrations (version, description, success) VALUES ($1, $2, TRUE)",
			m.Version, m.Description,
		)
		return err
	})
	if err == nil {
		return nil
	}

	if _, recErr := conn.Exec(context.WithoutCancel(ctx),
		`INSERT INTO schema_migrations (version, description, success) VALUES ($1, $2, FALSE)
		ON CONFLICT (version, success) DO UPDATE SET applied_at = NOW()`,
		m.Version, m.Description,
	); recErr != nil {
		logger.Get().Warn("failed to record the failed migration attempt", "version", m.Version, "error", recErr)
	}
	// pgconn leaves the hint out of Error(), but migrations put the fix there.
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Hint != "" {
		return fmt.Errorf("migration %d (%s) failed and was rolled back: %w (hint: %s)",
			m.Version, m.Release, err, pgErr.Hint)
	}
	return fmt.Errorf("migration %d (%s) failed and was rolled back: %w", m.Version, m.Release, err)
}

// detectBaseline checks that an unversioned database has the v0.10.1 schema:
// the address tables still have chain_name (dropped by migration 2) and the
// vm_msg_reject_package table added in v0.10.0 exists.
func detectBaseline(ctx context.Context, q querier) error {
	var hasAddresses, hasChainName, hasReject bool
	err := q.QueryRow(ctx, `SELECT
		to_regclass('gno_addresses') IS NOT NULL,
		EXISTS (SELECT 1 FROM information_schema.columns
			WHERE table_schema = current_schema() AND table_name = 'gno_addresses' AND column_name = 'chain_name'),
		to_regclass('vm_msg_reject_package') IS NOT NULL`,
	).Scan(&hasAddresses, &hasChainName, &hasReject)
	if err != nil {
		return fmt.Errorf("failed to inspect the database schema: %w", err)
	}
	switch {
	case !hasAddresses:
		return fmt.Errorf("%w and has no indexer tables: create it with \"indexer setup create-db\"", ErrNotVersioned)
	case !hasChainName:
		return fmt.Errorf("%w and is not a v0.10.1 schema: if you know its version, "+
			"record it with \"indexer migrate baseline <version>\"", ErrNotVersioned)
	case !hasReject:
		return fmt.Errorf("%w and predates v0.10.0, which can't be migrated: "+
			"create a new database with \"indexer setup create-db\" and re-index", ErrNotVersioned)
	}
	return nil
}
