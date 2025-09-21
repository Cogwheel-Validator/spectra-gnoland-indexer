package main

import (
	"log"
	"os"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/database"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/integration/synthetic"
	"go.yaml.in/yaml/v4"
)

type TestConfig struct {
	Host                      string        `yaml:"host"`
	Port                      int           `yaml:"port"`
	User                      string        `yaml:"user"`
	Password                  string        `yaml:"password"`
	Dbname                    string        `yaml:"dbname"`
	Sslmode                   string        `yaml:"sslmode"`
	PoolMaxConns              int           `yaml:"pool_max_conns"`
	PoolMinConns              int           `yaml:"pool_min_conns"`
	PoolMaxConnLifetime       time.Duration `yaml:"pool_max_conn_lifetime"`
	PoolMaxConnIdleTime       time.Duration `yaml:"pool_max_conn_idle_time"`
	PoolHealthCheckPeriod     time.Duration `yaml:"pool_health_check_period"`
	PoolMaxConnLifetimeJitter time.Duration `yaml:"pool_max_conn_lifetime_jitter"`
	ChainID                   string        `yaml:"chain_id"`
	MaxHeight                 uint64        `yaml:"max_height"`
	FromHeight                uint64        `yaml:"from_height"`
	ToHeight                  uint64        `yaml:"to_height"`
}

func configLoader() synthetic.SyntheticIntegrationTestConfig {
	yamlFile, err := os.ReadFile("test_config.yml")
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}
	var config TestConfig
	err = yaml.Unmarshal(yamlFile, &config)
	if err != nil {
		log.Fatalf("failed to unmarshal config: %v", err)
	}
	return synthetic.SyntheticIntegrationTestConfig{
		DatabaseConfig: database.DatabasePoolConfig{
			Host:     config.Host,
			Port:     config.Port,
			User:     config.User,
			Password: config.Password,
			Dbname:   config.Dbname,
			Sslmode:  config.Sslmode,
		},
		ChainID:    config.ChainID,
		MaxHeight:  config.MaxHeight,
		FromHeight: config.FromHeight,
		ToHeight:   config.ToHeight,
	}
}

func main() {
	log.Println("Starting synthetic integration test example...")

	testConfig := configLoader()
	log.Println("Test configuration loaded")

	// Run the synthetic integration test
	startTime := time.Now()
	log.Printf(
		`Testing started at %s \n
		Chain ID: %s \n
		Processing blocks %d to %d \n
		Max synthetic height: %d \n`,
		startTime.Format(time.RFC3339),
		testConfig.ChainID,
		testConfig.FromHeight, testConfig.ToHeight,
		testConfig.MaxHeight)

	err := synthetic.RunSyntheticIntegrationTest(&testConfig)
	if err != nil {
		log.Fatalf("Synthetic integration test failed: %v", err)
	}

	duration := time.Since(startTime)

	// if the program reaches this point, the test was successful
	log.Printf("Duration of the test: %v", duration)
	log.Println("Check your test database to see the synthetic data that was processed!")
}
