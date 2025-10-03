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

// BlockData represents the actual block data returned in the response body
type BlockData struct {
	Hash      string    `json:"hash" doc:"Block hash (hex-encoded)"`
	Height    uint64    `json:"height" doc:"Block height"`
	Timestamp time.Time `json:"timestamp" doc:"Block timestamp"`
	ChainID   string    `json:"chain_id" doc:"Chain identifier"`
	Txs       []string  `json:"txs" doc:"Transactions (base64 encoded)"`
	TxCount   int       `json:"tx_count" doc:"Number of transactions in the block"`
}
