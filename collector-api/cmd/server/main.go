package main

import (
	"collector-api/internal/api"
	"collector-api/internal/config"
	"collector-api/internal/db"
	"fmt"
	"log"
	"net/http"
)

func main() {
	// Load the configuration
	cfg, err := config.LoadConfigWithDefaultPath()
	if err != nil {
		log.Fatalf("Failed to load configuration: %v", err)
	}

	// Initialize the SQLite database
	db.InitDB(cfg.DBPath)

	// Setup routes and handlers, passing the configuration
	router := api.SetupRoutes(cfg)

	// Start the HTTP server
	address := fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort)
	log.Printf("Server starting on %s", address)
	if err := http.ListenAndServe(address, router); err != nil {
		log.Fatalf("Error starting server: %v", err)
	}
}
