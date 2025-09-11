package config

import (
	"time"
)

type Environment struct {
	Host    string `env:"DB_HOST" envDefault:"localhost"`
	Port    int    `env:"DB_PORT" envDefault:"5432"`
	User    string `env:"DB_USER" envDefault:"postgres"`
	Sslmode string `env:"DB_SSLMODE" envDefault:"disable"`
	// do not use password default unless for development or testing!!!
	Password string `env:"DB_PASSWORD" envDefault:"12345678"`
	Dbname   string `env:"DB_NAME" envDefault:"gnoland"`
}

type Config struct {
	RpcUrl                    string        `yaml:"rpc"`
	PoolMaxConns              int           `yaml:"pool_max_conns"`
	PoolMinConns              int           `yaml:"pool_min_conns"`
	PoolMaxConnLifetime       time.Duration `yaml:"pool_max_conn_lifetime"`
	PoolMaxConnIdleTime       time.Duration `yaml:"pool_max_conn_idle_time"`
	PoolHealthCheckPeriod     time.Duration `yaml:"pool_health_check_period"`
	PoolMaxConnLifetimeJitter time.Duration `yaml:"pool_max_conn_lifetime_jitter"`
	LivePooling               uint64        `yaml:"live_pooling"`
	MaxBlockChunkSize         uint64        `yaml:"max_block_chunk_size"`
	MaxTransactionChunkSize   uint64        `yaml:"max_transaction_chunk_size"`
	ChainName                 string        `yaml:"chain_name"`
}
