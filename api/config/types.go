package config

import "time"

type ApiConfig struct {
	// Basic connection info
	Host string `yaml:"host"`
	Port int    `yaml:"port"`
	// CORS config
	CorsAllowedOrigins []string `yaml:"cors_allowed_origins"`
	CorsAllowedMethods []string `yaml:"cors_allowed_methods"`
	CorsAllowedHeaders []string `yaml:"cors_allowed_headers"`
	CorsMaxAge         int      `yaml:"cors_max_age"`
	ChainName          string   `yaml:"chain_name"`
}

type ApiEnv struct {
	ApiDbHost                      string        `env:"API_DB_HOST" envDefault:"localhost"`
	ApiDbPort                      int           `env:"API_DB_PORT" envDefault:"5432"`
	ApiDbUser                      string        `env:"API_DB_USER" envDefault:"postgres"`
	ApiDbPassword                  string        `env:"API_DB_PASSWORD" envDefault:"12345678"`
	ApiDbName                      string        `env:"API_DB_NAME" envDefault:"gnoland"`
	ApiDbSslmode                   string        `env:"API_DB_SSLMODE" envDefault:"disable"`
	ApiDbPoolMaxConns              int           `env:"API_DB_POOL_MAX_CONNS" envDefault:"50"`
	ApiDbPoolMinConns              int           `env:"API_DB_POOL_MIN_CONNS" envDefault:"10"`
	ApiDbPoolMaxConnLifetime       time.Duration `env:"API_DB_POOL_MAX_CONN_LIFETIME" envDefault:"10m"`
	ApiDbPoolMaxConnIdleTime       time.Duration `env:"API_DB_POOL_MAX_CONN_IDLE_TIME" envDefault:"5m"`
	ApiDbPoolHealthCheckPeriod     time.Duration `env:"API_DB_POOL_HEALTH_CHECK_PERIOD" envDefault:"1m"`
	ApiDbPoolMaxConnLifetimeJitter time.Duration `env:"API_DB_POOL_MAX_CONN_LIFETIME_JITTER" envDefault:"1m"`
}
