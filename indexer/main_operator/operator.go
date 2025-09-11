package mainoperator

import (
	"context"
	"log"
	"time"

	addressCache "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/address_cache"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/config"
	dp "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/data_processor"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/database"
	mainTypes "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/main_types"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/orchestrator"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/query"
	rpcClient "github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/rpc_client"
)

// This function is not ready to be used
// this is just a placeholder
// every part of the indexer should be initialized within the main operator
func InitMainOperator(
	configPath string,
	envPath string,
	rpcFlags mainTypes.RpcFlags,
	runningFlags mainTypes.RunningFlags) {
	// load config
	conf, err := config.LoadConfig(configPath)
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	// load environment
	env, err := config.LoadEnvironment(envPath)
	if err != nil {
		log.Fatalf("failed to load environment: %v", err)
	}

	// get the chain name
	chainName := &conf.ChainName

	mc := initializeMajorConstructors(conf, env, *chainName, rpcFlags)

	// initialize the orchestrator
	orch := orchestrator.NewOrchestrator(
		runningFlags.RunningMode, conf, *chainName, mc.db, mc.gnoRpcClient, mc.dataProcessor, mc.queryOperator,
	)

	// let the orchestrator do it's thing
	if runningFlags.RunningMode == "live" {
		orch.LiveProcess(context.Background(), runningFlags.SkipInitialDbCheck)
	} else if runningFlags.RunningMode == "historic" {
		if runningFlags.FromHeight == 0 || runningFlags.ToHeight == 0 {
			log.Fatalf("from height and to height are required for historic mode")
		} else if runningFlags.FromHeight > runningFlags.ToHeight {
			log.Fatalf("from height must be less than to height")
		}
		orch.HistoricProcess(runningFlags.FromHeight, runningFlags.ToHeight)
	} else {
		log.Fatalf("invalid running mode, please choose between live and historic")
	}
}

// initializeDatabase is a private function to initialize the database
// it is used to initialize the database for the main operator
//
// Args:
//   - conf: the config
//   - env: the environment
//
// Returns:
//   - the database
//   - error if any
func initializeDatabase(conf *config.Config, env *config.Environment) *database.TimescaleDb {
	// check if the config has any null
	// if rpc is null throw an error and exit
	if conf.RpcUrl == "" {
		log.Fatalf("rpc url is required")
	}
	// if pool max connections is 0 or nil set a default of 100
	if conf.PoolMaxConns == 0 {
		conf.PoolMaxConns = 100
	}
	// set to a default of 10 if not set
	if conf.PoolMinConns == 0 {
		conf.PoolMinConns = 10
	}
	// set to a default of 10 minutes if not set
	if conf.PoolMaxConnLifetime == 0 {
		conf.PoolMaxConnLifetime = 10 * time.Minute
	}
	// set to a default of 5 minutes if not set
	if conf.PoolMaxConnIdleTime == 0 {
		conf.PoolMaxConnIdleTime = 5 * time.Minute
	}
	// set to a default of 1 minute if not set
	if conf.PoolHealthCheckPeriod == 0 {
		conf.PoolHealthCheckPeriod = 1 * time.Minute
	}
	// set to a default of 1 minute if not set
	if conf.PoolMaxConnLifetimeJitter == 0 {
		conf.PoolMaxConnLifetimeJitter = 1 * time.Minute
	}

	// pull config and env data to load init the database pool

	dbConfig := database.DatabasePoolConfig{
		Host:                      env.Host,
		Port:                      env.Port,
		User:                      env.User,
		Password:                  env.Password,
		Dbname:                    env.Dbname,
		Sslmode:                   env.Sslmode,
		PoolMaxConns:              conf.PoolMaxConns,
		PoolMinConns:              conf.PoolMinConns,
		PoolMaxConnLifetime:       conf.PoolMaxConnLifetime,
		PoolMaxConnIdleTime:       conf.PoolMaxConnIdleTime,
		PoolHealthCheckPeriod:     conf.PoolHealthCheckPeriod,
		PoolMaxConnLifetimeJitter: conf.PoolMaxConnLifetimeJitter,
	}

	// no need to return error since it will throw a fatal error and exit the program
	db := database.NewTimescaleDb(dbConfig)
	return db
}

func initializeMajorConstructors(
	conf *config.Config,
	env *config.Environment,
	chainName string,
	rpcFlags mainTypes.RpcFlags) *MajorConstructors {
	// initialize the rpc client
	// check the flags first
	// this is yet to be implemented but for now just set it and later fix anything
	if rpcFlags.RequestsPerWindow == 0 {
		// realistically this could be ignored
		// if this really is the case set it to 10 million since
		// this should indicate that no rate limiting is needed
		rpcFlags.RequestsPerWindow = 10000000
	}
	if rpcFlags.TimeWindow == 0 {
		// set it to a default of 1 minute
		rpcFlags.TimeWindow = 1 * time.Minute
	} else if rpcFlags.TimeWindow <= 0 {
		log.Fatalf("time window must be greater than 0")
	}

	// init all of the major constructors

	// initialize the database
	db := initializeDatabase(conf, env)

	// initialize the rpc client
	gnoRpcClient, err := rpcClient.NewRateLimitedRpcClient(
		conf.RpcUrl, nil, rpcFlags.RequestsPerWindow, rpcFlags.TimeWindow,
	)
	if err != nil {
		log.Fatalf("failed to initialize rpc client: %v", err)
	}

	// initialize the validator cache
	validatorCache := addressCache.NewAddressCache(chainName, db, true)

	// initialize the address cache
	addressCache := addressCache.NewAddressCache(chainName, db, false)

	// initialize the data processor
	dataProcessor := dp.NewDataProcessor(db, addressCache, validatorCache, chainName)

	// initialize the query operator
	queryOperator := query.NewQueryOperator(gnoRpcClient)

	return &MajorConstructors{
		db:             db,
		gnoRpcClient:   gnoRpcClient,
		validatorCache: validatorCache,
		addressCache:   addressCache,
		dataProcessor:  dataProcessor,
		queryOperator:  queryOperator,
	}
}
