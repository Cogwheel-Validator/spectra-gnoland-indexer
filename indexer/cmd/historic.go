package cmd

import (
	"log"

	mainOperator "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/main_operator"
	mainTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/main_types"
	"github.com/spf13/cobra"
)

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
