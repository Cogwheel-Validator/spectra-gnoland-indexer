package cli

import (
	"time"

	mainOperator "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/main_operator"
	mainTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/main_types"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/logger"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:   "run",
	Short: "Run the indexer",
	Long:  `Run the indexer in either live or historic mode.`,
}

// parseCommonRunFlags reads the persistent flags shared by the run subcommands
// (live and historic) and assembles the RPC rate-limit configuration.
func parseCommonRunFlags(cmd *cobra.Command) (configPath string, rpcFlags mainTypes.RpcFlags, compressEvents bool, err error) {
	l := logger.Get()

	configPath, err = cmd.Flags().GetString("config")
	if err != nil {
		l.Error("failed to get config path", "error", err)
		return "", mainTypes.RpcFlags{}, false, err
	}
	maxRequestsPerWindow, err := cmd.Flags().GetInt("max-req-per-window")
	if err != nil {
		l.Error("failed to get max requests per window", "error", err)
		return "", mainTypes.RpcFlags{}, false, err
	}
	rateLimitWindow, err := cmd.Flags().GetDuration("rate-limit-window")
	if err != nil {
		l.Error("failed to get rate limit window", "error", err)
		return "", mainTypes.RpcFlags{}, false, err
	}
	timeout, err := cmd.Flags().GetDuration("timeout")
	if err != nil {
		l.Error("failed to get timeout", "error", err)
		return "", mainTypes.RpcFlags{}, false, err
	}
	compressEvents, err = cmd.Flags().GetBool("compress-events")
	if err != nil {
		l.Error("failed to get compress events", "error", err)
		return "", mainTypes.RpcFlags{}, false, err
	}

	rpcFlags = mainTypes.RpcFlags{
		RequestsPerWindow: maxRequestsPerWindow,
		TimeWindow:        rateLimitWindow,
		Timeout:           timeout,
	}
	return configPath, rpcFlags, compressEvents, nil
}

// launchIndexer logs startup state and starts the main operator for either run mode.
func launchIndexer(
	cmd *cobra.Command, configPath string, rpcFlags mainTypes.RpcFlags, runningFlags mainTypes.RunningFlags,
) error {
	l := logger.Get()
	skipMigrationCheck, err := cmd.Flags().GetBool("skip-migration-check")
	if err != nil {
		l.Error("failed to get skip migration check", "error", err)
		return err
	}
	runningFlags.SkipMigrationCheck = skipMigrationCheck

	l.Info("indexer started")
	if runningFlags.CompressEvents {
		l.Warn("compress events is enabled, this is experimental and it might slow down the data processing speed")
	}
	mainOperator.InitMainOperator(configPath, ".", rpcFlags, runningFlags)
	return nil
}

func init() {
	// Add subcommands
	runCmd.AddCommand(liveCmd)
	runCmd.AddCommand(historicCmd)

	// Persistent flags that apply to all run subcommands (live and historic)
	runCmd.PersistentFlags().StringP("config", "c", "config.yml", "config file path")
	runCmd.PersistentFlags().IntP("max-req-per-window", "m", 10000000, "max requests per window")
	runCmd.PersistentFlags().DurationP("rate-limit-window", "r", 1*time.Minute, "rate limit window")
	runCmd.PersistentFlags().DurationP("timeout", "t", 20*time.Second, "timeout")
	runCmd.PersistentFlags().BoolP("compress-events", "e", false, "compress events")
	runCmd.PersistentFlags().Bool("skip-migration-check", false,
		"start even if the database schema version doesn't match this binary (for forks with their own schema)")
}
