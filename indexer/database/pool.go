package database

import (
	"context"
	"fmt"
	"log"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TimescaleDb struct {
	pool *pgxpool.Pool
}

func NewTimescaleDb(config DatabasePoolConfig) *TimescaleDb {
	pool, err := ConnectToDb(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	return &TimescaleDb{
		pool: pool,
	}
}

func ConnectToDb(config DatabasePoolConfig) (*pgxpool.Pool, error) {
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

// create a new database named "gnoland"
func CreateDatabase(db *TimescaleDb, dbname string) error {
	_, err := db.pool.Exec(context.Background(), fmt.Sprintf("CREATE DATABASE %s", dbname))
	if err != nil {
		return err
	}
	return nil
}

// switch to the database named "gnoland"
// this is used to switch to the database after creating it
// most of the time when the postgres server is running, it will be in the "postgres" database
// only to be used initiating command
func SwitchDatabase(db *TimescaleDb, config DatabasePoolConfig) error {
	// Close the current connection
	db.pool.Close()

	// Create a new config with the target database name
	newConfig := config
	newConfig.Dbname = "gnoland"

	// Create a new connection to the target database
	newPool, err := ConnectToDb(newConfig)
	if err != nil {
		return fmt.Errorf("failed to connect to gnoland database: %w", err)
	}

	// Replace the old pool with the new one
	db.pool = newPool
	return nil
}

// GetPool returns the database connection pool
func (db *TimescaleDb) GetPool() *pgxpool.Pool {
	return db.pool
}
