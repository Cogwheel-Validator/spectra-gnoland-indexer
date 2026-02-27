package main

import (
	"log"
	"os"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/compression/train"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "train",
	Short: "Train",
	Long:  "Train tool for Zstd dictionary training the Spectra Gnoland Indexer",
	Run: func(cmd *cobra.Command, args []string) {
		log.Println("Training Zstd dictionary")
		configPath, err := cmd.Flags().GetString("config")
		if err != nil {
			log.Fatalf("failed to get config path: %v", err)
		}
		amount := cmd.Flags().Uint64("amount", 1000, "the amount of events to collect")
		chainName := cmd.Flags().String("chain-name", "gnoland", "the name of the chain")
		dictPath := cmd.Flags().String("dict-path", "events.zstd.bin", "the path to the zstd dictionary")

		loadedConfig, err := train.LoadTrainingConfig(&configPath)
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}
		dbConfig := database.DatabasePoolConfig{
			Host:                      loadedConfig.Host,
			Port:                      loadedConfig.Port,
			User:                      loadedConfig.User,
			Password:                  loadedConfig.Password,
			Dbname:                    loadedConfig.Dbname,
			Sslmode:                   loadedConfig.Sslmode,
			PoolMaxConns:              loadedConfig.PoolMaxConns,
			PoolMinConns:              loadedConfig.PoolMinConns,
			PoolMaxConnLifetime:       loadedConfig.PoolMaxConnLifetime,
			PoolMaxConnIdleTime:       loadedConfig.PoolMaxConnIdleTime,
			PoolHealthCheckPeriod:     loadedConfig.PoolHealthCheckPeriod,
			PoolMaxConnLifetimeJitter: loadedConfig.PoolMaxConnLifetimeJitter,
		}
		db := train.InitDatabase(dbConfig)

		events, err := train.CollectEvents(db, *chainName, *amount)
		if err != nil {
			log.Fatalf("failed to collect events: %v", err)
		}
		dict := train.BuildZstdDict(events)
		err = os.WriteFile(*dictPath, dict, 0644)
		if err != nil {
			log.Fatalf("failed to write zstd dictionary: %v", err)
		}
		log.Println("Zstd dictionary built")
		log.Println("Zstd dictionary written to: ", *dictPath)
	},
}

func init() {
	rootCmd.Flags().StringP("config", "c", "", "the path to the config file")
	rootCmd.Flags().Uint64P("amount", "a", 1000, "the amount of events to collect")
	rootCmd.Flags().StringP("chain-name", "n", "gnoland", "the name of the chain")
	rootCmd.Flags().StringP("dict-path", "d", "events.zstd.bin", "the path to the zstd dictionary")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("failed to execute command: %v", err)
	}
}
