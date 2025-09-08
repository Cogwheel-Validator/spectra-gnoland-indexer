package config

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/caarlos0/env/v6"
	"github.com/joho/godotenv"
	"go.yaml.in/yaml/v4"
)

func LoadConfig(path string) (*Config, error) {
	yamlFile, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var config Config
	err = yaml.Unmarshal(yamlFile, &config)
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
