package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/config"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/handlers"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/keystore"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/ratelimit"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/valkey"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/danielgtaylor/huma/v2"
	"github.com/danielgtaylor/huma/v2/adapters/humachi"
	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"github.com/spf13/cobra"
)

func runServe(cmd *cobra.Command, args []string) {
	var configPath string
	var err error
	var certFilePath string
	var keyFilePath string

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

	certFilePath, err = cmd.Flags().GetString("cert-file")
	if err != nil {
		log.Fatalf("failed to get cert file path: %v", err)
	}
	keyFilePath, err = cmd.Flags().GetString("key-file")
	if err != nil {
		log.Fatalf("failed to get key file path: %v", err)
	}

	router := chi.NewMux()
	router.Use(middleware.Logger)
	router.Use(middleware.Recoverer)
	router.Use(middleware.CleanPath)
	router.Use(middleware.Compress(5, "application/json", "application/problem+json"))
	router.Use(middleware.Heartbeat("/"))

	corsOptions := cors.Options{
		AllowedOrigins: conf.CorsAllowedOrigins,
		AllowedMethods: conf.CorsAllowedMethods,
		AllowedHeaders: conf.CorsAllowedHeaders,
		MaxAge:         conf.CorsMaxAge,
	}
	if len(corsOptions.AllowedOrigins) == 0 {
		corsOptions.AllowedOrigins = []string{"*"}
	}
	if len(corsOptions.AllowedMethods) == 0 {
		corsOptions.AllowedMethods = []string{"GET"}
	}
	if len(corsOptions.AllowedHeaders) == 0 {
		corsOptions.AllowedHeaders = []string{"Origin", "Content-Type", "Accept", "X-API-Key"}
	}
	if corsOptions.MaxAge == 0 {
		corsOptions.MaxAge = 600
	}
	router.Use(cors.Handler(corsOptions))

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

	// Initialize key store and rate limiter
	ks := keystore.NewKeyStore(db)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	if err := ks.Refresh(ctx); err != nil {
		log.Printf("warning: initial key store refresh failed: %v", err)
	}
	cancel()

	refreshInterval := conf.KeyRefreshInterval
	if refreshInterval == 0 {
		refreshInterval = 5 * time.Minute
	}
	ks.StartPeriodicRefresh(refreshInterval)

	valkeyEnv, err := config.LoadValkeyEnvironment(&config.DefaultEnvFileReader{}, ".")
	if err != nil {
		log.Fatalf("failed to load valkey environment: %v", err)
	}

	valkeyClient, err := valkey.NewValkeyClient(valkeyEnv.Host, valkeyEnv.Port)
	if err != nil {
		log.Fatalf("failed to create valkey client: %v", err)
	}

	ipRPM := conf.IpRpmLimit
	if ipRPM == 0 {
		ipRPM = 30
	}

	rl := ratelimit.NewRateLimiter(valkeyClient, ks, ipRPM, 1*time.Minute)
	router.Use(rl.Middleware)

	humaConfig := huma.DefaultConfig("Spectra Gnoland Indexer API", Version)
	humaConfig.Components.SecuritySchemes = map[string]*huma.SecurityScheme{
		"apiKey": {
			Type: "apiKey",
			Name: "X-API-Key",
			In:   "header",
			Description: `API key for authenticated access with a higher rate limit. Pass your key in the X-API-Key header.
			You can query the APi without the API key but the stricter rate limit will be applied.`,
		},
	}
	// Apply the scheme globally but make it optional (empty map = anonymous access allowed).
	humaConfig.Security = []map[string][]string{
		{"apiKey": {}},
		{},
	}

	api := humachi.New(router, humaConfig)

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

	blocksHandler := handlers.NewBlocksHandler(db, conf.ChainName)
	transactionsHandler := handlers.NewTransactionsHandler(db, conf.ChainName)
	addressHandler := handlers.NewAddressHandler(db, conf.ChainName)

	router.Get("/favicon.ico", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/x-icon")
		_, err := w.Write(favicon)
		if err != nil {
			log.Printf("failed to write favicon: %v", err)
			http.Error(w, "Internal Server Error", http.StatusInternalServerError)
			return
		}
	})

	huma.Get(api, "/blocks/{height}", blocksHandler.GetBlock,
		func(op *huma.Operation) {
			op.Summary = "Get Block Height"
			op.Description = "Retrieve block data by its height"
		})
	huma.Get(api, "/blocks/{from_height}/{to_height}", blocksHandler.GetFromToBlocks,
		func(op *huma.Operation) {
			op.Summary = "Get From To Blocks"
			op.Description = `Retrieve blocks data by its height range. 
			From height must be less than to height and the difference must be less than 100.
			The response will contain the blocks data in the range.
			`
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

	huma.Get(
		api, "/transactions/{tx_hash}", transactionsHandler.GetTransactionBasic,
		func(op *huma.Operation) {
			op.Summary = "Get Transaction Basic"
			op.Description = "Retrieve basic transaction data by its hash"
		})
	huma.Get(
		api,
		"/transactions/{tx_hash}/messages",
		transactionsHandler.GetTransactionMessage,
		func(op *huma.Operation) {
			op.Summary = "Get All Transaction Messages"
			op.Description = "Retrieve all messages contained within a transaction by its hash"
		})
	huma.Get(api, "/transactions", transactionsHandler.GetTransactionsByCursor,
		func(op *huma.Operation) {
			op.Summary = "Get Transactions"
			op.Description = `Retrieve transactions by setting the limit and using cursor.
			To fetch multiple transaction you can use this endpoint. Without cursor you will
			fetch latest data. However if you need to acquire older data you can use cursor.
			The cursor is a string in the form of timestamp|tx_hash(base64url encoded).
			The timestamp is the timestamp of the transaction and the tx_hash is the hash of the transaction.
			The tx_hash is base64url encoded to be able to query safely via API.
			`
		})

	huma.Get(api, "/address/{address}/txs", addressHandler.GetAddressTxs,
		func(op *huma.Operation) {
			op.Summary = "Get Address Transactions"
			op.Description = `Retrieve all transactions for a given address.
			There are 3 ways to query the transactions:
			
			1. by timestamp range
			2. by cursor
			3. by limit and page

			For the timestamp range, you can specify the from and to timestamps.
			For the cursor, just make the first query without any parameters besides the address. 
			The query will contain the data alongside the next cursor that can be used as a query to get the data needed.
			For the limit and page, you can specify the limit and page to get the next set of transactions.
			`
		})

	huma.Post(api, "/convert/base64-to-base64url", handlers.ConvertFromBase64toBase64Url,
		func(op *huma.Operation) {
			op.Summary = "Convert Base64 to Base64Url"
			op.Description = "Convert a base64 encoded tx hash to a base64url encoded tx hash"
		})
	huma.Post(api, "/convert/base64url-to-base64", handlers.ConvertFromBase64UrlToBase64,
		func(op *huma.Operation) {
			op.Summary = "Convert Base64Url to Base64"
			op.Description = "Convert a base64url encoded tx hash to a base64 encoded tx hash"
		})

	addr := fmt.Sprintf("%s:%d", conf.Host, conf.Port)

	if certFilePath != "" && keyFilePath != "" {
		log.Printf("Starting server on %s with HTTPS", addr)
		err = http.ListenAndServeTLS(addr, certFilePath, keyFilePath, router)
		if err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	} else {
		log.Printf("Starting server on %s with HTTP", addr)
		err = http.ListenAndServe(addr, router)
		if err != nil {
			log.Fatalf("failed to start server: %v", err)
		}
	}
}
