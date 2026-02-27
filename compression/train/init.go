package train

import (
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"go.yaml.in/yaml/v4"
)

// It is identical to the database.DatabasePoolConfig struct.
// However, it is used for the training process and used very rarely.
type TrainingConfig struct {
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	Dbname   string `yaml:"dbname"`
	Sslmode  string `yaml:"sslmode"`

	PoolMaxConns              int           `yaml:"pool_max_conns"`
	PoolMinConns              int           `yaml:"pool_min_conns"`
	PoolMaxConnLifetime       time.Duration `yaml:"pool_max_conn_lifetime"`
	PoolMaxConnIdleTime       time.Duration `yaml:"pool_max_conn_idle_time"`
	PoolHealthCheckPeriod     time.Duration `yaml:"pool_health_check_period"`
	PoolMaxConnLifetimeJitter time.Duration `yaml:"pool_max_conn_lifetime_jitter"`
}

// LoadTrainingConfig loads the training config from the config file
//
// Usage:
//
// # Used to load the training config from the config file
//
// Parameters:
//   - configPath: the path to the config file
//
// Returns:
//   - TrainingConfig: the training config
//   - error: if the training config fails to load
func LoadTrainingConfig(configPath *string) (TrainingConfig, error) {
	var config TrainingConfig
	if configPath == nil {
		return config, fmt.Errorf("config path is required")
	}
	if !strings.HasSuffix(*configPath, ".yml") && !strings.HasSuffix(*configPath, ".yaml") {
		return config, fmt.Errorf("config path must end with .yml")
	}
	yamlFile, err := os.ReadFile(*configPath)
	if err != nil {
		return config, err
	}
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		return config, err
	}
	return config, nil
}

// Init initializes the database connection pool
//
// Usage:
//
// # Used to initialize the database connection pool
//
// Parameters:
//   - config: the configuration for the database connection pool
//
// Returns:
//   - *database.TimescaleDb: the database connection pool
//   - error: if the database connection pool fails to initialize
func InitDatabase(config database.DatabasePoolConfig) *database.TimescaleDb {
	return database.NewTimescaleDb(config)
}
