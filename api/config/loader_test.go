package config_test

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Cogwheel-Validator/spectra-gnoland-indexer/api/config"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestLoadConfig_ValidConfig tests loading a complete valid config
func TestLoadConfig_ValidConfig(t *testing.T) {
	tmpDir := t.TempDir()

	configContent := `host: 0.0.0.0
port: 9000
cors_allowed_origins:
- "http://localhost:3000"
- "http://localhost:8000"
cors_allowed_methods:
- "GET"
- "POST"
- "PUT"
cors_allowed_headers:
- "Content-Type"
- "Authorization"
cors_max_age: 7200
chain_name: "gnoland"
`

	configPath := filepath.Join(tmpDir, "test-config.yml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err, "failed to write test config file")

	cfg, err := config.LoadConfig(&config.YamlFileReader{}, configPath)
	require.NoError(t, err, "LoadConfig should not return an error")

	assert.Equal(t, "0.0.0.0", cfg.Host)
	assert.Equal(t, 9000, cfg.Port)
	assert.Equal(t, "gnoland", cfg.ChainName)
	assert.Len(t, cfg.CorsAllowedOrigins, 2)
	assert.Equal(t, "http://localhost:3000", cfg.CorsAllowedOrigins[0])
	assert.Len(t, cfg.CorsAllowedMethods, 3)
	assert.Equal(t, "GET", cfg.CorsAllowedMethods[0])
	assert.Equal(t, 7200, cfg.CorsMaxAge)
}

// TestLoadConfig_DefaultValues tests that empty fields get default values
func TestLoadConfig_DefaultValues(t *testing.T) {
	tmpDir := t.TempDir()

	// Minimal config with only required fields
	configContent := `chain_name: "gnoland"
`

	configPath := filepath.Join(tmpDir, "minimal-config.yml")
	err := os.WriteFile(configPath, []byte(configContent), 0644)
	require.NoError(t, err)

	cfg, err := config.LoadConfig(&config.YamlFileReader{}, configPath)
	require.NoError(t, err)

	// Verify defaults are applied
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 8080, cfg.Port)
}

// TestLoadConfig_FileNotFound tests error handling when config file doesn't exist
func TestLoadConfig_FileNotFound(t *testing.T) {
	_, err := config.LoadConfig(&config.YamlFileReader{}, "/nonexistent/path/to/config.yml")
	assert.Error(t, err, "should return error for missing file")
}

// TestLoadConfig_InvalidYAML tests error handling for malformed YAML
func TestLoadConfig_InvalidYAML(t *testing.T) {
	tmpDir := t.TempDir()

	invalidYAML := `host: 0.0.0.0
port: not_a_number: this_breaks_yaml
chain_name: gnoland
`

	configPath := filepath.Join(tmpDir, "invalid-config.yml")
	err := os.WriteFile(configPath, []byte(invalidYAML), 0644)
	require.NoError(t, err)

	_, err = config.LoadConfig(&config.YamlFileReader{}, configPath)
	assert.Error(t, err, "should return error for invalid YAML")
}

// TestLoadConfig_EmptyFile tests handling of empty config file
func TestLoadConfig_EmptyFile(t *testing.T) {
	tmpDir := t.TempDir()

	configPath := filepath.Join(tmpDir, "empty-config.yml")
	err := os.WriteFile(configPath, []byte(""), 0644)
	require.NoError(t, err)

	cfg, err := config.LoadConfig(&config.YamlFileReader{}, configPath)
	require.NoError(t, err, "should handle empty config gracefully")

	// Should apply defaults
	assert.Equal(t, "localhost", cfg.Host)
	assert.Equal(t, 8080, cfg.Port)
}

// TestLoadEnvironment_NoEnvFile tests that LoadEnvironment handles missing .env gracefully
func TestLoadEnvironment_NoEnvFile(t *testing.T) {
	tmpDir := t.TempDir()

	// Don't create any .env file
	// This should not error, just print a message to stdout
	env, err := config.LoadEnvironment(&config.DefaultEnvFileReader{}, tmpDir)
	require.NoError(t, err, "should not error when .env file doesn't exist")
	assert.NotNil(t, env)
}

// TestLoadEnvironment_ValidEnvFile tests loading a valid .env file
func TestLoadEnvironment_ValidEnvFile(t *testing.T) {
	tmpDir := t.TempDir()

	envContent := `
	API_DB_HOST=postgres.example.com
	API_DB_PORT=5432
	API_DB_USER=testuser
	API_DB_PASSWORD=testpass
	API_DB_NAME=testdb
	`

	envPath := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envPath, []byte(envContent), 0644)
	require.NoError(t, err)

	// Save and clear any existing env vars to avoid interference
	oldHost := os.Getenv("API_DB_HOST")
	oldPort := os.Getenv("API_DB_PORT")
	oldUser := os.Getenv("API_DB_USER")
	oldPassword := os.Getenv("API_DB_PASSWORD")
	oldName := os.Getenv("API_DB_NAME")

	err = os.Unsetenv("API_DB_HOST")
	require.NoError(t, err)
	err = os.Unsetenv("API_DB_PORT")
	require.NoError(t, err)
	err = os.Unsetenv("API_DB_USER")
	require.NoError(t, err)
	err = os.Unsetenv("API_DB_PASSWORD")
	require.NoError(t, err)
	err = os.Unsetenv("API_DB_NAME")
	require.NoError(t, err)

	t.Cleanup(func() {
		// Restore original env vars
		if oldHost != "" {
			err = os.Setenv("API_DB_HOST", oldHost)
			require.NoError(t, err)
		}
		if oldPort != "" {
			err = os.Setenv("API_DB_PORT", oldPort)
			require.NoError(t, err)
		}
		if oldUser != "" {
			err = os.Setenv("API_DB_USER", oldUser)
			require.NoError(t, err)
		}
		if oldPassword != "" {
			err = os.Setenv("API_DB_PASSWORD", oldPassword)
			require.NoError(t, err)
		}
		if oldName != "" {
			err = os.Setenv("API_DB_NAME", oldName)
			require.NoError(t, err)
		}
	})

	env, err := config.LoadEnvironment(&config.DefaultEnvFileReader{}, tmpDir)
	require.NoError(t, err)

	assert.Equal(t, "postgres.example.com", env.ApiDbHost)
	assert.Equal(t, 5432, env.ApiDbPort)
	assert.Equal(t, "testuser", env.ApiDbUser)
	assert.Equal(t, "testpass", env.ApiDbPassword)
	assert.Equal(t, "testdb", env.ApiDbName)
}

// TestLoadEnvironment_DefaultValues tests that defaults are applied for missing env vars
func TestLoadEnvironment_DefaultValues(t *testing.T) {
	tmpDir := t.TempDir()

	// Empty .env file will use all defaults
	envPath := filepath.Join(tmpDir, ".env")
	err := os.WriteFile(envPath, []byte(""), 0644)
	require.NoError(t, err)

	// Save and clear any existing env vars to avoid interference
	oldHost := os.Getenv("API_DB_HOST")
	oldPort := os.Getenv("API_DB_PORT")
	oldUser := os.Getenv("API_DB_USER")
	oldPassword := os.Getenv("API_DB_PASSWORD")
	oldName := os.Getenv("API_DB_NAME")
	oldSslmode := os.Getenv("API_DB_SSLMODE")

	err = os.Unsetenv("API_DB_HOST")
	require.NoError(t, err)
	err = os.Unsetenv("API_DB_PORT")
	require.NoError(t, err)
	err = os.Unsetenv("API_DB_USER")
	require.NoError(t, err)
	err = os.Unsetenv("API_DB_PASSWORD")
	require.NoError(t, err)
	err = os.Unsetenv("API_DB_NAME")
	require.NoError(t, err)
	err = os.Unsetenv("API_DB_SSLMODE")
	require.NoError(t, err)

	t.Cleanup(func() {
		// Restore original env vars
		if oldHost != "" {
			err = os.Setenv("API_DB_HOST", oldHost)
			require.NoError(t, err)
		}
		if oldPort != "" {
			err = os.Setenv("API_DB_PORT", oldPort)
			require.NoError(t, err)
		}
		if oldUser != "" {
			err = os.Setenv("API_DB_USER", oldUser)
			require.NoError(t, err)
		}
		if oldPassword != "" {
			err = os.Setenv("API_DB_PASSWORD", oldPassword)
			require.NoError(t, err)
		}
		if oldName != "" {
			err = os.Setenv("API_DB_NAME", oldName)
			require.NoError(t, err)
		}
		if oldSslmode != "" {
			err = os.Setenv("API_DB_SSLMODE", oldSslmode)
			require.NoError(t, err)
		}
	})

	env, err := config.LoadEnvironment(&config.DefaultEnvFileReader{}, tmpDir)
	require.NoError(t, err)

	// Verify defaults from types.go
	assert.Equal(t, "localhost", env.ApiDbHost)
	assert.Equal(t, 5432, env.ApiDbPort)
	assert.Equal(t, "postgres", env.ApiDbUser)
	assert.Equal(t, "12345678", env.ApiDbPassword)
	assert.Equal(t, "gnoland", env.ApiDbName)
	assert.Equal(t, "disable", env.ApiDbSslmode)
}
