package main

import (
	"collector-api/internal/api"
	"collector-api/internal/config"
	"collector-api/internal/db"
	"collector-api/internal/storage"
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
)

func main() {
	reprocessFull := flag.Bool("reprocess-full", false, "Reprocess all full snapshots")
	reprocessCompact := flag.Bool("reprocess-compact", false, "Reprocess all compact snapshots")
	flag.Parse()

	// Load the configuration from the global config path
	cfg, err := config.LoadConfigWithDefaultPath()
	if err != nil {
		log.Printf("Failed to load configuration: %v", err)
		os.Exit(-1)
	}

	// Ensure the required storage directories exist
	err = storage.EnsureStorageDirectories(cfg.StorageDir)
	if err != nil {
		log.Printf("Failed to create storage directories: %v", err)
		os.Exit(-1)
	}

	// Initialize the SQLite database
	db.InitDB(cfg.DBPath)

	if err := api.ReprocessSnapshots(cfg, *reprocessFull, *reprocessCompact); err != nil {
		log.Printf("Error reprocessing snapshots: %v", err)
		os.Exit(-1)
	}

	// Setup routes and handlers, passing the configuration
	router := api.SetupRoutes(cfg)

	// Start the HTTP server
	address := fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort)
	log.Printf("Server starting on %s", address)
	if err := http.ListenAndServe(address, router); err != nil {
		log.Fatalf("Error starting server: %v", err)
		os.Exit(-1)
	}
}
