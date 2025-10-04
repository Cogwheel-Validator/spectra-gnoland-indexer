package main

import (
	"log"
	"time"

	mainOperator "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/main_operator"
	mainTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/main_types"
	"github.com/spf13/cobra"
)

var (
	Commit  = "unknown" // Set via ldflags at build time
	Version = "unknown" // Set via ldflags at build time
)

var rootCmd = &cobra.Command{
	Use:     "indexer",
	Short:   "Spectra Gnoland Indexer",
	Long:    "A blockchain indexer for Gnoland that processes blocks and transactions.",
	Version: Version + " (commit: " + Commit + ")",
}

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
		maxRequestsPerWindow, err := rootCmd.Flags().GetInt("max-req-per-window")
		if err != nil {
			log.Fatalf("failed to get max requests per window: %v", err)
		}
		rateLimitWindow, err := rootCmd.Flags().GetDuration("rate-limit-window")
		if err != nil {
			log.Fatalf("failed to get rate limit window: %v", err)
		}
		timeout, err := rootCmd.Flags().GetDuration("timeout")
		if err != nil {
			log.Fatalf("failed to get timeout: %v", err)
		}
		compressEvents, err := rootCmd.Flags().GetBool("compress-events")
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

var historicCmd = &cobra.Command{
	Use:   "historic",
	Short: "Run the indexer in historic mode",
	Long: `Runs the spectra indexer in historic mode, processing blocks from a given height to a given height.
	The historic mode takes in starting height point and a finishing height. It should be used to 
	sync up the database to the latest block height. 

	It can also be useful if you want to index blockchain partially and work with data for any kind of testing
	or partial scan of the chain where you want to index from a certain height to a certain height.
	`,
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Running in historic mode")

		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			log.Fatalf("failed to get config path: %v", err)
		}

		// Parse RPC flags from root command
		maxRequestsPerWindow, err := rootCmd.Flags().GetInt("max-req-per-window")
		if err != nil {
			log.Fatalf("failed to get max requests per window: %v", err)
		}
		rateLimitWindow, err := rootCmd.Flags().GetDuration("rate-limit-window")
		if err != nil {
			log.Fatalf("failed to get rate limit window: %v", err)
		}
		timeout, err := rootCmd.Flags().GetDuration("timeout")
		if err != nil {
			log.Fatalf("failed to get timeout: %v", err)
		}
		compressEvents, err := rootCmd.Flags().GetBool("compress-events")
		if err != nil {
			log.Fatalf("failed to get compress events: %v", err)
		}

		// Parse historic-specific flags
		fromHeight, err := cmd.Flags().GetUint64("from-height")
		if err != nil {
			log.Fatalf("failed to get from height: %v", err)
		}
		toHeight, err := cmd.Flags().GetUint64("to-height")
		if err != nil {
			log.Fatalf("failed to get to height: %v", err)
		}

		// Build flag structs
		rateLimitFlags := mainTypes.RpcFlags{
			RequestsPerWindow: maxRequestsPerWindow,
			TimeWindow:        rateLimitWindow,
			Timeout:           timeout,
		}

		runningFlags := mainTypes.RunningFlags{
			RunningMode:        "historic",
			SkipInitialDbCheck: false, // Not applicable for historic mode
			CompressEvents:     compressEvents,
			FromHeight:         fromHeight,
			ToHeight:           toHeight,
		}

		log.Println("Indexer started")
		mainOperator.InitMainOperator(configPath, ".", rateLimitFlags, runningFlags)
	},
}

func init() {
	// Add subcommands to root
	rootCmd.AddCommand(liveCmd)
	rootCmd.AddCommand(historicCmd)

	// Common flags on root command
	rootCmd.PersistentFlags().StringP("config", "c", "config.yml", "config file path")
	rootCmd.PersistentFlags().IntP("max-req-per-window", "m", 10000000, "max requests per window")
	rootCmd.PersistentFlags().DurationP("rate-limit-window", "r", 1*time.Minute, "rate limit window")
	rootCmd.PersistentFlags().DurationP("timeout", "t", 20*time.Second, "timeout")
	rootCmd.PersistentFlags().BoolP("compress-events", "e", false, "compress events")

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

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("failed to execute command: %v", err)
	}
}
