package cmd

import (
	"log"

	mainOperator "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/main_operator"
	mainTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/main_types"
	"github.com/spf13/cobra"
)

var liveCmd = &cobra.Command{
	Use:   "live",
	Short: "Run the indexer in live mode",
	Long: `Runs the spectra indexer, listening to any new blocks and processing them.
	It will check the database for the last processed height and start from there.

	In the events the database is empty, it will start from block height 1. This can be used
	to sync up the database to the latest block height.

	However if you do not need previous data, you can run the live mode with the skip-db-check flag set to true.
	Afterwards you can run live mode normal without the skip-db-check flag.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Running in live mode")

		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			log.Fatalf("failed to get config path: %v", err)
		}

		// Parse RPC flags from root command
		maxRequestsPerWindow, err := RootCmd.Flags().GetInt("max-req-per-window")
		if err != nil {
			log.Fatalf("failed to get max requests per window: %v", err)
		}
		rateLimitWindow, err := RootCmd.Flags().GetDuration("rate-limit-window")
		if err != nil {
			log.Fatalf("failed to get rate limit window: %v", err)
		}
		timeout, err := RootCmd.Flags().GetDuration("timeout")
		if err != nil {
			log.Fatalf("failed to get timeout: %v", err)
		}
		compressEvents, err := RootCmd.Flags().GetBool("compress-events")
		if err != nil {
			log.Fatalf("failed to get compress events: %v", err)
		}

		// Parse live-specific flags
		skipDbCheck, err := cmd.Flags().GetBool("skip-db-check")
		if err != nil {
			log.Fatalf("failed to get skip db check: %v", err)
		}

		// Build flag structs
		rateLimitFlags := mainTypes.RpcFlags{
			RequestsPerWindow: maxRequestsPerWindow,
			TimeWindow:        rateLimitWindow,
			Timeout:           timeout,
		}

		runningFlags := mainTypes.RunningFlags{
			RunningMode:        "live",
			SkipInitialDbCheck: skipDbCheck,
			CompressEvents:     compressEvents,
			FromHeight:         0,
			ToHeight:           0,
		}

		log.Println("Indexer started")
		mainOperator.InitMainOperator(configPath, ".", rateLimitFlags, runningFlags)
	},
}
