package cmd

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"slices"
	"syscall"
	"time"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/indexer/config"
	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/pkgs/database"
	"github.com/spf13/cobra"
	"go.yaml.in/yaml/v4"
	"golang.org/x/term"
)

// parseCommonFlags extracts and validates common database flags
func parseCommonFlags(cmd *cobra.Command, defaultDbName string) (*dbParams, error) {
	params := &dbParams{}

	params.host, _ = cmd.Flags().GetString("db-host")
	params.port, _ = cmd.Flags().GetInt("db-port")
	params.user, _ = cmd.Flags().GetString("db-user")
	params.sslMode, _ = cmd.Flags().GetString("ssl-mode")
	params.name, _ = cmd.Flags().GetString("db-name")

	// Apply environment variable fallbacks (for CI/CD)
	if params.host == "" {
		if envHost := os.Getenv("DB_HOST"); envHost != "" {
			params.host = envHost
		}
	}
	if params.port == 0 {
		if envPort := os.Getenv("DB_PORT"); envPort != "" {
			fmt.Sscanf(envPort, "%d", &params.port)
		}
	}
	if params.user == "" {
		if envUser := os.Getenv("DB_USER"); envUser != "" {
			params.user = envUser
		}
	}
	if params.name == "" {
		if envDbName := os.Getenv("DB_NAME"); envDbName != "" {
			params.name = envDbName
		}
	}

	// Apply defaults if still empty
	if params.sslMode == "" {
		params.sslMode = "disable"
	}
	if params.host == "" {
		params.host = "localhost"
	}
	if params.port == 0 {
		params.port = 5432
	}
	if params.user == "" {
		params.user = "postgres"
	}
	if params.name == "" {
		params.name = defaultDbName
	}

	// Validate
	if !slices.Contains(allowedSslModes, params.sslMode) {
		return nil, fmt.Errorf("invalid ssl mode: %s", params.sslMode)
	}
	if params.port > 65535 {
		return nil, fmt.Errorf("invalid port: %d", params.port)
	}

	return params, nil
}

// promptPassword prompts user for password input or reads from environment
func promptPassword() (string, error) {
	// First check if password is provided via environment variable (for CI/CD)
	if envPassword := os.Getenv("DB_PASSWORD"); envPassword != "" {
		return envPassword, nil
	}

	// Interactive mode: prompt user for password
	fmt.Print("Enter the database password: ")
	bytePassword, err := term.ReadPassword(int(syscall.Stdin))
	if err != nil {
		return "", fmt.Errorf("failed to read password: %v", err)
	}
	fmt.Println()
	return string(bytePassword), nil
}

// createDatabaseConfig creates a DatabasePoolConfig from dbParams
func (p *dbParams) createDatabaseConfig() database.DatabasePoolConfig {
	return database.DatabasePoolConfig{
		Host:                      p.host,
		Port:                      p.port,
		User:                      p.user,
		Dbname:                    p.name,
		Password:                  p.password,
		Sslmode:                   p.sslMode,
		PoolMaxConns:              10,
		PoolMinConns:              1,
		PoolMaxConnLifetime:       10 * time.Minute,
		PoolMaxConnIdleTime:       5 * time.Minute,
		PoolHealthCheckPeriod:     1 * time.Minute,
		PoolMaxConnLifetimeJitter: 1 * time.Minute,
	}
}

func createConfig(overwrite bool, fileName string) {
	if fileName == "" {
		fileName = "config.yaml"
	}
	absolutePath, err := filepath.Abs(fileName)
	if err != nil {
		log.Fatalf("failed to get absolute path: %v", err)
	}
	fileName = absolutePath
	// config file
	config := config.Config{
		RpcUrl:                    "http://localhost:26657",
		PoolMaxConns:              50,
		PoolMinConns:              10,
		PoolMaxConnLifetime:       5 * time.Minute,
		PoolMaxConnIdleTime:       5 * time.Minute,
		PoolHealthCheckPeriod:     1 * time.Minute,
		PoolMaxConnLifetimeJitter: 1 * time.Minute,
		LivePooling:               5 * time.Second,
		MaxBlockChunkSize:         50,
		MaxTransactionChunkSize:   100,
		ChainName:                 "gnoland",
		RetryAmount:               &[]int{6}[0],
		Pause:                     &[]int{3}[0],
		PauseTime:                 &[]time.Duration{15 * time.Second}[0],
		ExponentialBackoff:        &[]time.Duration{2 * time.Second}[0],
	}
	// marshal the config to yaml
	yamlFile, err := yaml.Marshal(config)
	if err != nil {
		log.Fatalf("failed to marshal config: %v", err)
	}
	// check if the config file exists
	_, err = os.Stat(fileName)
	if err != nil && !os.IsNotExist(err) {
		// make the file if it doesn't exist
		err = os.WriteFile(fileName, yamlFile, 0644)
		if err != nil {
			log.Fatalf("failed to create config file: %v", err)
		}

	} else if err == nil && overwrite {
		err = os.WriteFile(fileName, yamlFile, 0644)
		if err != nil {
			log.Fatalf("failed to overwrite config file: %v", err)
		}
	} else if err == nil && !overwrite {
		log.Fatalf("config file already exists, use --overwrite to overwrite it")
	} else {
		log.Fatalf("failed to create config file: %v", err)
	}
}
