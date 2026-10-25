package cli

import (
	"context"
	"fmt"
	"os"
	"strconv"
	"text/tabwriter"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/migrations"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database/timescaledb"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/logger"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/spf13/cobra"
)

var migrateCmd = &cobra.Command{
	Use:   "migrate",
	Short: "Database schema migrations.",
	Long: `Inspect and upgrade the database schema.

Applied migrations are recorded in the schema_migrations table. "run" refuses
to start unless the database is at the version this binary expects (see
--skip-migration-check). Databases created with v0.10.1 have no recorded
version yet; "migrate up" recognizes them and records the baseline first.

Run migrations with the account that owns the tables (e.g. postgres, the one
used for "setup create-db"), with the indexer stopped. Take a backup first.`,
}

func init() {
	migrateCmd.AddCommand(migrateStatusCmd)
	migrateCmd.AddCommand(migrateUpCmd)
	migrateCmd.AddCommand(migrateBaselineCmd)

	for _, cmd := range []*cobra.Command{migrateStatusCmd, migrateUpCmd, migrateBaselineCmd} {
		cmd.Flags().StringP("db-host", "b", "", "The database host, default is localhost")
		cmd.Flags().IntP("db-port", "p", 0, "The database port, default is 5432")
		cmd.Flags().StringP("db-user", "u", "", "The database user, default is postgres")
		cmd.Flags().StringP("db-name", "d", "", "The database name, default is gnoland")
		cmd.Flags().StringP("ssl-mode", "s", "", "The SSL mode for the database connection, default is disable")
		cmd.Flags().String("ssl-rootcert", "", "Path to the CA certificate for server verification (verify-ca/verify-full)")
		cmd.Flags().String("ssl-cert", "", "Path to the client certificate (mutual TLS only)")
		cmd.Flags().String("ssl-key", "", "Path to the client private key (mutual TLS only)")
	}
}

// migrateConnect opens a pool on the database given by the common flags.
func migrateConnect(cmd *cobra.Command) (*pgxpool.Pool, error) {
	params, err := parseCommonFlags(cmd, "gnoland")
	if err != nil {
		return nil, err
	}
	params.password, err = promptPassword()
	if err != nil {
		return nil, err
	}
	return timescaledb.NewTimescaleDbSetup(params.createDatabaseConfig()).GetPool(), nil
}

var migrateStatusCmd = &cobra.Command{
	Use:   "status",
	Short: "Show applied and pending migrations.",
	RunE: func(cmd *cobra.Command, args []string) error {
		pool, err := migrateConnect(cmd)
		if err != nil {
			return err
		}
		defer pool.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		state, err := migrations.ReadState(ctx, pool)
		if err != nil {
			return err
		}

		w := tabwriter.NewWriter(os.Stdout, 0, 0, 2, ' ', 0)
		fmt.Fprintln(w, "VERSION\tRELEASE\tSTATE\tDESCRIPTION")
		for _, m := range migrations.All() {
			status := "pending"
			if at, ok := state.Applied[m.Version]; ok {
				status = "applied " + at.UTC().Format(time.RFC3339)
			} else if at, ok := state.Failed[m.Version]; ok {
				status = "failed " + at.UTC().Format(time.RFC3339)
			}
			fmt.Fprintf(w, "%d\t%s\t%s\t%s\n", m.Version, m.Release, status, m.Description)
		}
		if err := w.Flush(); err != nil {
			return err
		}

		fmt.Printf("\ndatabase version: %d, required: %d\n", state.Current(), migrations.Latest())
		if err := state.Verify(); err != nil {
			fmt.Println(err)
		}
		return nil
	},
}

var migrateUpCmd = &cobra.Command{
	Use:   "up",
	Short: "Apply all pending migrations.",
	Long: `Apply all pending migrations in order. Each one runs in its own transaction
and is recorded in schema_migrations; a failed one is rolled back and recorded
as failed.

Stop the indexer first and run it with the account that owns the tables.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.Get()
		pool, err := migrateConnect(cmd)
		if err != nil {
			return err
		}
		defer pool.Close()

		applied, err := migrations.Up(context.Background(), pool)
		if err != nil {
			return err
		}
		if len(applied) == 0 {
			l.Info("database is up to date", "version", migrations.Latest())
			return nil
		}
		l.Info("database migrated", "applied", len(applied), "version", migrations.Latest())
		return nil
	},
}

var migrateBaselineCmd = &cobra.Command{
	Use:   "baseline [version]",
	Short: "Record a version as applied without running it.",
	Long: `Record every migration up to and including [version] as applied, without
running them.

Only for databases that are known to have that schema but were never
versioned, for example one created by an unreleased build. Recording the wrong
version makes "migrate up" skip changes the database needs.

Useful when developing or when working with a custom build of the indexer.`,
	Args: cobra.ExactArgs(1),
	RunE: func(cmd *cobra.Command, args []string) error {
		l := logger.Get()
		version, err := strconv.ParseInt(args[0], 10, 32)
		if err != nil {
			return fmt.Errorf("invalid version %q: %w", args[0], err)
		}

		pool, err := migrateConnect(cmd)
		if err != nil {
			return err
		}
		defer pool.Close()

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if err := migrations.Baseline(ctx, pool, int32(version)); err != nil {
			return err
		}
		l.Info("recorded baseline", "version", version)
		return nil
	},
}
