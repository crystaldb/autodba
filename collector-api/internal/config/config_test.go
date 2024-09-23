package config_test

import (
	"collector-api/internal/config"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLoadConfig(t *testing.T) {
	// Create a mock config.json for testing
	configFile := `{
		"server_host": "127.0.0.1",
		"server_port": 8080,
		"db_path": "./test.db",
		"storage_dir": "./storage",
		"api_key": "test-api-key",
		"debug": true
	}`

	// Write to a temporary file
	tmpFile, err := os.CreateTemp("", "config.json")
	if err != nil {
		t.Fatal(err)
	}
	defer os.Remove(tmpFile.Name()) // Clean up the file after the test

	// Write the mock config to the temp file
	if _, err := tmpFile.Write([]byte(configFile)); err != nil {
		t.Fatal(err)
	}

	// Load the config from the temp file
	cfg, err := config.LoadConfig(tmpFile.Name())

	// Assert no error occurred and values are correct
	assert.NoError(t, err)
	assert.Equal(t, "127.0.0.1", cfg.ServerHost)
	assert.Equal(t, 8080, cfg.ServerPort)
	assert.Equal(t, "./test.db", cfg.DBPath)
	assert.Equal(t, "./storage", cfg.StorageDir)
	assert.Equal(t, "test-api-key", cfg.APIKey)
	assert.True(t, cfg.Debug)
}
