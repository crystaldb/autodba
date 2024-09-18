package config

import (
	"encoding/json"
	"log"
	"os"
)

type Config struct {
	ServerHost string `json:"server_host"`
	ServerPort int    `json:"server_port"`
	DBPath     string `json:"db_path"` // Path to SQLite database file
	StorageDir string `json:"storage_dir"`
	APIKey     string `json:"api_key"`
	Debug      bool   `json:"debug"` // Enable or disable debug logging
}

func LoadConfig() (*Config, error) {
	file, err := os.Open("collector-api-config.json")
	if err != nil {
		log.Fatalf("Could not open config file: %v", err)
		return nil, err
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	config := Config{}
	err = decoder.Decode(&config)
	if err != nil {
		log.Fatalf("Could not decode config JSON: %v", err)
		return nil, err
	}
	return &config, nil
}
