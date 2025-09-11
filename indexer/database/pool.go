package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

// NewTimescaleDb is a constructor function that creates a new TimescaleDb instance
//
// Args:
//   - config: the database config, a struct that contains the necessary data from the config file
//
// Returns:
//   - *TimescaleDb: the TimescaleDb instance
//
// The method will not throw an error if the TimescaleDb is not found, it will just return nil
func NewTimescaleDb(config DatabasePoolConfig) *TimescaleDb {
	pool, err := connectToDb(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return &TimescaleDb{
		pool: pool,
	}
}

// connectToDb is a internal function that connects to the database
//
// Args:
//   - config: the database config, a struct that contains the necessary data from the config file
//
// Returns:
//   - *pgxpool.Pool: the database connection pool
//
// The method will not throw an error if the database connection pool is not found, it will just return nil
func connectToDb(config DatabasePoolConfig) (*pgxpool.Pool, error) {
	parseConfig, err := pgxpool.ParseConfig(
		fmt.Sprintf(
			`host=%s port=%d user=%s password=%s 
			dbname=%s sslmode=%s pool_max_conns=%d 
			pool_min_conns=%d pool_max_conn_lifetime=%s 
			pool_max_conn_idle_time=%s pool_health_check_period=%s 
			pool_max_conn_lifetime_jitter=%s`,
			config.Host,
			config.Port,
			config.User,
			config.Password,
			config.Dbname,
			config.Sslmode,
			config.PoolMaxConns,
			config.PoolMinConns,
			config.PoolMaxConnLifetime,
			config.PoolMaxConnIdleTime,
			config.PoolHealthCheckPeriod,
			config.PoolMaxConnLifetimeJitter))
	if err != nil {
		return nil, err
	}
	conn, err := pgxpool.NewWithConfig(context.Background(), parseConfig)
	if err != nil {
		return nil, err
	}
	return conn, nil
}

// create a new database with the given name
//
// Args:
//   - db: the database connection pool
//   - dbname: the name of the database to create
//
// Returns:
//   - error: an error if the creation fails
func CreateDatabase(db *TimescaleDb, dbname string) error {
	_, err := db.pool.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE %s", dbname))
	if err != nil {
		return err
	}
	return nil
}

// Switch to the database with the given name
// this is used to switch to the database after creating it
// most of the time when the postgres server is running, it will be in the "postgres" database
// only to be used initiating command
//
// Args:
//   - db: the database connection pool
//   - config: the database config, a struct that contains the necessary data from the config file
//
// Returns:
//   - error: an error if the switching fails
//
// TODO: the devs could integrate the indexer within already existing timescale db
// so remove hard coded dbname gnoland anywhere else in the project
func SwitchDatabase(db *TimescaleDb, config DatabasePoolConfig, dbname string) error {
	// Close the current connection
	db.pool.Close()

	// Create a new config with the target database name
	newConfig := config
	newConfig.Dbname = dbname

	// Create a new connection to the target database
	newPool, err := connectToDb(newConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to %s database: %w", dbname, err)
	}

	// Replace the old pool with the new one
	db.pool = newPool
	return nil
}

// GetPool returns the database connection pool
//
// Args:
//   - db: the database connection pool
//
// Returns:
//   - *pgxpool.Pool: the database connection pool
func (db *TimescaleDb) GetPool() *pgxpool.Pool {
	return db.pool
}
