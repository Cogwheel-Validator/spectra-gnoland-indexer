// Package migrations versions the indexer database schema.
//
// Every schema change after v0.10.1 ships as a numbered migration. Applied
// versions are recorded in the schema_migrations table: "setup create-db"
// records all of them because it already creates the latest schema, "migrate
// up" applies the missing ones to an existing database, and "run" refuses to
// start when the database is not at the version the binary expects.
//
// Migrations are frozen once released. They are plain SQL instead of being
// generated from pkgs/schema, so later changes to the structs can't change what
// an old migration does.
//
// Also this will only work on the indexer made exactly with version `v0.10.x`.
// Anything older will probably break since this was only created and tested
// for v0.10.1.
package migrations

import (
	"embed"
	"fmt"
)

//go:embed sql/*.sql
var sqlFiles embed.FS

// Migration is a single schema change.
type Migration struct {
	// Version is the sequential migration number, stored in schema_migrations.
	Version int32
	// Release is the indexer release that introduced the schema.
	Release string
	// Description is a human-readable summary, stored in schema_migrations.
	Description string
	// SQL runs in one transaction together with the schema_migrations insert.
	// Empty for the baseline, which only marks an existing schema.
	SQL string
	// After holds statements that can't run inside a transaction (refreshing a
	// continuous aggregate, for example). They run one by one after the commit
	// and are best effort: they only speed things up, the schema is already
	// correct without them.
	After []string
}

// BaselineVersion is the schema created by v0.10.1, the first release that is
// versioned. Databases created by v0.10.1 have an empty schema_migrations
// table and are recognized by "migrate up".
const BaselineVersion int32 = 1

// Package history of all migrations.
var all = []Migration{
	{
		Version:     BaselineVersion,
		Release:     "v0.10.1",
		Description: "baseline: schema created by v0.10.1",
	},
	{
		Version: 2,
		Release: "v1.0.0",
		Description: "shared address tables with chains column, chain_stats, " +
			"resolve functions, tx_counter failed count and gas",
		SQL:   mustRead("sql/0002_v1_0_0.sql"),
		After: []string{"CALL refresh_continuous_aggregate('tx_counter', NULL, NOW())"},
	},
}

func mustRead(name string) string {
	b, err := sqlFiles.ReadFile(name)
	if err != nil {
		panic(fmt.Sprintf("migrations: %v", err))
	}
	return string(b)
}

// All returns every known migration in version order.
func All() []Migration {
	return all
}

// Latest returns the version this binary expects the database to be at.
func Latest() int32 {
	return all[len(all)-1].Version
}

// find returns the migration with the given version.
func find(version int32) (Migration, bool) {
	for _, m := range all {
		if m.Version == version {
			return m, true
		}
	}
	return Migration{}, false
}
