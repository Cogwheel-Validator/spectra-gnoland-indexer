package main

import (
	_ "embed"
	"fmt"
	"log"
	"net/http"
	"strings"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/config"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/handlers"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/spf13/cobra"
	// "github.com/go-chi/httprate"
)

//go:embed public/favicon.ico
var favicon []byte

var (
	Commit  = "unknown" // Set via ldflags at build time
	Version = "unknown" // Set via ldflags at build time
)

var rootCmd = &cobra.Command{
	Use:     "api",
	Short:   "Spectra Gnoland Indexer API",
	Long:    "API for the Spectra Gnoland Indexer",
	Version: Version + " (commit: " + Commit + ")",
	Run: func(cmd *cobra.Command, args []string) {
		var configPath string
		var err error
		var certFilePath string
		var keyFilePath string

		// check all of the flags

		// check config file
		configPath, err = cmd.Flags().GetString("config")
		if err != nil {
			log.Fatalf("failed to get config path: %v", err)
		}
		conf, err := config.LoadConfig(&config.YamlFileReader{}, configPath)
		if err != nil {
			log.Fatalf("failed to load config: %v", err)
		}

		env, err := config.LoadEnvironment(&config.DefaultEnvFileReader{}, ".")
		if err != nil {
			log.Fatalf("failed to load environment: %v", err)
		}

		// check cert file and key file
		certFilePath, err = cmd.Flags().GetString("cert-file")
		if err != nil {
			log.Fatalf("failed to get cert file path: %v", err)
		}
		keyFilePath, err = cmd.Flags().GetString("key-file")
		if err != nil {
			log.Fatalf("failed to get key file path: %v", err)
		}

		// Setup router and middleware
		router := chi.NewMux()
		router.Use(middleware.Logger)
		router.Use(middleware.Recoverer)
		router.Use(middleware.CleanPath)
		router.Use(middleware.Compress(5, "application/json", "application/problem+json"))
		// heartbeat route
		router.Use(middleware.Heartbeat("/"))

		// Configure CORS from config file
		corsOptions := cors.Options{
			AllowedOrigins: conf.CorsAllowedOrigins,
			AllowedMethods: conf.CorsAllowedMethods,
			AllowedHeaders: conf.CorsAllowedHeaders,
			MaxAge:         conf.CorsMaxAge,
		}
		// Set defaults if not provided
		if len(corsOptions.AllowedOrigins) == 0 {
			corsOptions.AllowedOrigins = []string{"*"}
		}
		if len(corsOptions.AllowedMethods) == 0 {
			corsOptions.AllowedMethods = []string{"GET"}
		}
		if len(corsOptions.AllowedHeaders) == 0 {
			corsOptions.AllowedHeaders = []string{"Origin", "Content-Type", "Accept"}
		}
		if corsOptions.MaxAge == 0 {
			corsOptions.MaxAge = 600
		}
		router.Use(cors.Handler(corsOptions))

		/* rate limiting middleware
		I think it would be easier to implement the rate limit to some reverse proxy like nginx, apache, etc.
		but I will leave it here for now
		if you do decide you want more control over the rate limiting, to use the chi rate limiting middleware
		just remove the comment lines and build it with the chi rate limiting middleware
		and adjust the rate limit and time window as you see fit
		TODO: add proper documentation for this and maybe add it as a option in the config file?
		router.Use(httprate.LimitByIP(100, 1*time.Minute))
		*/

		api := humachi.New(router, huma.DefaultConfig("Spectra Gnoland Indexer API", Version))

		openApi := api.OpenAPI()
		openApi.Info = &huma.Info{
			Title:       "Spectra Gnoland Indexer API",
			Version:     strings.TrimPrefix(Version, "v"),
			Description: "API for the Spectra Gnoland Indexer",
			Contact: &huma.Contact{
				Name:  "Cogwheel Validator",
				URL:   "https://cogwheel.zone",
				Email: "info@cogwheel.zone",
			},
			License: &huma.License{
				Name: "Apache 2.0",
				URL:  "https://github.com/Cogwheel-Validator/spectra-gnoland-indexer?tab=Apache-2.0-1-ov-file#readme",
			},
		}
		openApi.ExternalDocs = &huma.ExternalDocs{
			URL: "https://github.com/Cogwheel-Validator/spectra-gnoland-indexer/tree/main/docs",
		}

		// Initialize database connection from environment variables
		db := database.NewTimescaleDb(database.DatabasePoolConfig{
			Host:                      env.ApiDbHost,
			Port:                      env.ApiDbPort,
			User:                      env.ApiDbUser,
			Password:                  env.ApiDbPassword,
			Dbname:                    env.ApiDbName,
			Sslmode:                   env.ApiDbSslmode,
			PoolMaxConns:              env.ApiDbPoolMaxConns,
			PoolMinConns:              env.ApiDbPoolMinConns,
			PoolMaxConnLifetime:       env.ApiDbPoolMaxConnLifetime,
			PoolMaxConnIdleTime:       env.ApiDbPoolMaxConnIdleTime,
			PoolHealthCheckPeriod:     env.ApiDbPoolHealthCheckPeriod,
			PoolMaxConnLifetimeJitter: env.ApiDbPoolMaxConnLifetimeJitter,
		})

		// Initialize handlers with dependencies
		blocksHandler := handlers.NewBlocksHandler(db, conf.ChainName)
		transactionsHandler := handlers.NewTransactionsHandler(db, conf.ChainName)
		addressHandler := handlers.NewAddressHandler(db, conf.ChainName)

		// favicon route
		router.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "image/x-icon")
			w.Write(favicon)
		})

		// Register Block API routes
		huma.Get(api, "/block/{height}", blocksHandler.GetBlock,
			func(op *huma.Operation) {
				op.Summary = "Get Block Height"
				op.Description = "Retrieve block data by its height"
			})
		huma.Get(api, "/blocks/{from_height}/{to_height}", blocksHandler.GetFromToBlocks,
			func(op *huma.Operation) {
				op.Summary = "Get From To Blocks"
				op.Description = "Retrieve blocks data by its height range"
			})
		huma.Get(api, "/blocks/{block_height}/signers", blocksHandler.GetAllBlockSigners,
			func(op *huma.Operation) {
				op.Summary = "Get All Block Signers"
				op.Description = "Retrieve all validators that signed a block by its height"
			})
		huma.Get(api, "/blocks/latest", blocksHandler.GetLatestBlock,
			func(op *huma.Operation) {
				op.Summary = "Get Latest Block"
				op.Description = "Retrieve the latest block data"
			})
		huma.Get(api, "/blocks", blocksHandler.GetLastXBlocks,
			func(op *huma.Operation) {
				op.Summary = "Get Last X Blocks"
				op.Description = "Retrieve the last X blocks data"
			})

		// Register Transaction API routes
		huma.Get(
			api, "/transaction/{tx_hash}", transactionsHandler.GetTransactionBasic,
			func(op *huma.Operation) {
				op.Summary = "Get Transaction Basic"
				op.Description = "Retrieve basic transaction data by its hash"
			})
		huma.Get(
			api,
			"/transaction/{tx_hash}/message",
			transactionsHandler.GetTransactionMessage,
			func(op *huma.Operation) {
				op.Summary = "Get All Transaction Messages"
				op.Description = "Retrieve all messages contained within a transaction by its hash"
			})
		huma.Get(api, "/transactions", transactionsHandler.GetLastXTransactions,
			func(op *huma.Operation) {
				op.Summary = "Get Last X Transactions"
				op.Description = "Retrieve the last X transactions data"
			})

		// Register Address API routes
		huma.Get(api, "/address/{address}/txs", addressHandler.GetAddressTxs,
			func(op *huma.Operation) {
				op.Summary = "Get Address Transactions"
				op.Description = "Retrieve all transactions for a given address for a certain time period"
			})

		// Start server using config values
		addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)
		log.Printf("Starting server on %s", addr)

		// if cert file and key file are provided, use https
		if certFilePath != "" && keyFilePath != "" {
			err = http.ListenAndServeTLS(addr, certFilePath, keyFilePath, router)
			log.Printf("Starting server on %s with HTTPS", addr)
			if err != nil {
				log.Fatalf("failed to start server: %v", err)
			}
		} else {
			err = http.ListenAndServe(addr, router)
			log.Printf("Starting server on %s with HTTP", addr)
			if err != nil {
				log.Fatalf("failed to start server: %v", err)
			}
		}

	},
}

func init() {
	rootCmd.PersistentFlags().StringP("config", "c", "config-api.yml", "config file path")
	rootCmd.PersistentFlags().StringP("cert-file", "t", "", "certificate file path")
	rootCmd.PersistentFlags().StringP("key-file", "k", "", "key file path")
}

func main() {
	if err := rootCmd.Execute(); err != nil {
		log.Fatalf("failed to execute command: %v", err)
	}
}
