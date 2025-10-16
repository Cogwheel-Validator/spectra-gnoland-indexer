package cmd

import (
	"log"
	"time"

	"github.com/spf13/cobra"
)

var (
	Commit  = "unknown" // Set via ldflags at build time
	Version = "unknown" // Set via ldflags at build time
)

var allowedSslModes = []string{"disable", "require", "verify-ca", "verify-full", "allow", "prefer"}

var RootCmd = &cobra.Command{
	Use:     "indexer",
	Short:   "Spectra Gnoland Indexer",
	Long:    "A blockchain indexer for Gnoland that processes blocks and transactions.",
	Version: Version + " (commit: " + Commit + ")",
}

func init() {
	// Add subcommands to root
	RootCmd.AddCommand(liveCmd)
	RootCmd.AddCommand(historicCmd)
	RootCmd.AddCommand(createDbCmd)
	RootCmd.AddCommand(createUserCmd)

	// Common flags on functional commands such as live and historic commands
	for _, cmd := range []*cobra.Command{createDbCmd, createUserCmd} {
		cmd.Flags().String("db-host", "", "The database host, default is localhost")
		cmd.Flags().Int("db-port", 0, "The database port, default is 5432")
		cmd.Flags().String("db-user", "", "The database user, default is postgres")
		cmd.Flags().String("db-name", "", "The database name, default is postgres")
		cmd.Flags().String("ssl-mode", "", "The SSL mode for the database connection, default is disable")
	}

	// Common flags on functional commands such as live and historic commands
	for _, cmd := range []*cobra.Command{liveCmd, historicCmd} {
		cmd.Flags().StringP("config", "c", "config.yml", "config file path")
		cmd.Flags().IntP("max-req-per-window", "m", 10000000, "max requests per window")
		cmd.Flags().DurationP("rate-limit-window", "r", 1*time.Minute, "rate limit window")
		cmd.Flags().DurationP("timeout", "t", 20*time.Second, "timeout")
		cmd.Flags().BoolP("compress-events", "e", false, "compress events")
	}

	// Add create-user command specific flags
	createUserCmd.Flags().String("privilege", "", "The privilege level for the user (reader or writer)")
	createUserCmd.Flags().String("user", "", "The user name for the user to create")

	// Add create-db command specific flags
	createDbCmd.Flags().String("new-db-name", "", "The database name to create, default is gnoland")
	createDbCmd.Flags().String("chain-name", "", "The chain name for the database type enum, default is gnoland")

	// Live-specific flags
	liveCmd.Flags().BoolP("skip-db-check", "s", false, "skip initial database check")

	// Historic-specific flags
	historicCmd.Flags().Uint64P("from-height", "f", 1, "starting block height")
	historicCmd.Flags().Uint64P("to-height", "o", 1000, "ending block height")

	// Mark required flags for historic mode
	err := historicCmd.MarkFlagRequired("from-height")
	if err != nil {
		log.Fatalf("failed to mark from height as required: %v", err)
	}
	err = historicCmd.MarkFlagRequired("to-height")
	if err != nil {
		log.Fatalf("failed to mark to height as required: %v", err)
	}
}
