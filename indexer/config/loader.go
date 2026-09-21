package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"go.yaml.in/yaml/v4"
)

var chainNamePattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_-]*$`)

// ValidateChainName checks that a chain name is non-empty and only contains letters, digits, '_' and '-'.
func ValidateChainName(name string) error {
	if strings.TrimSpace(name) == "" {
		return errors.New("chain name is required")
	}
	if !chainNamePattern.MatchString(name) {
		return fmt.Errorf("invalid chain name %q: use only letters, digits, '_' and '-'", name)
	}
	return nil
}

func LoadConfig(path string) (*Config, error) {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
	// all fields except rpc and chain name will throw errors if they are not set
	// so we need a checker for the rpc url
	if config.RpcUrl == "" {
		return nil, errors.New("rpc url is required")
	} else if !strings.HasPrefix(config.RpcUrl, "http://") && !strings.HasPrefix(config.RpcUrl, "https://") {
		return nil, errors.New("rpc url must start with http:// or https://")
	}
	if err := ValidateChainName(config.ChainName); err != nil {
		return nil, err
	}

	if err != nil {
		return nil, err
	}
	return &config, nil
}

func LoadEnvironment(path string) (*Environment, error) {
	possibleEnvFiles := []string{
		".env", // accept only .env for now
	}
	// check if any of the files exist within the current path
	existingFiles := []string{}
	for _, envFile := range possibleEnvFiles {
		formPath := filepath.Join(path, envFile)
		if _, err := os.Stat(formPath); err == nil {
			existingFiles = append(existingFiles, formPath)
		}
	}
	// if there are multiple files, decide which has highest priority
	// 1. production , 2. development, 3. local, 4. default
	// only use the regular .env for now return to this laster
	if len(existingFiles) == 0 {
		fmt.Println("No environment file found. Searching for os environment variables.")
	} else if len(existingFiles) == 1 {
		absPath, err := filepath.Abs(existingFiles[0])
		if err != nil {
			return nil, err
		}
		err = godotenv.Load(absPath)
		if err != nil {
			return nil, fmt.Errorf("error loading .env file: %w", err)
		}
		fmt.Printf("Loaded environment variables from %s\n", absPath)
	}

	environment := Environment{}
	if err := env.Parse(&environment); err != nil {
		return nil, fmt.Errorf("failed to parse environment variables: %w", err)
	}

	return &environment, nil
}
