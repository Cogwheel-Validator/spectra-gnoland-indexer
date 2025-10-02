package database

import (
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

// TimescaleDb is the database connection pool
type TimescaleDb struct {
	pool *pgxpool.Pool
}

// DatabasePoolConfig is the configuration for the database pool.
type DatabasePoolConfig struct {
	// Basic connection info
	User     string
	Password string
	Host     string
	Port     int
	Dbname   string
	Sslmode  string

	// Pool config
	PoolMaxConns              int
	PoolMinConns              int
	PoolMaxConnLifetime       time.Duration
	PoolMaxConnIdleTime       time.Duration
	PoolHealthCheckPeriod     time.Duration
	PoolMaxConnLifetimeJitter time.Duration
}
