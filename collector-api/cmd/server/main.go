package main

import (
	"collector-api/internal/api"
	"collector-api/internal/config"
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

	err = storage.InitQueryStorage(cfg.DBPath)
	if err != nil {
		log.Printf("Failed to initialize query storage: %v", err)
		os.Exit(-1)
	}

	// Create error channel for goroutines
	errChan := make(chan error, 2)

	// Start reprocessing in a goroutine if needed
	if *reprocessFull || *reprocessCompact {
		queue := api.GetQueueInstance()
		// Lock the queue before creating the goroutine to avoid race conditions
		queue.Lock()
		go func() {
			if err := api.ReprocessSnapshots(cfg, *reprocessFull, *reprocessCompact); err != nil {
				errChan <- fmt.Errorf("reprocessing snapshots: %w", err)
				return
			}
			errChan <- nil
		}()
	}

	// Start HTTP server in a goroutine
	go func() {
		router := api.SetupRoutes(cfg)
		address := fmt.Sprintf("%s:%d", cfg.ServerHost, cfg.ServerPort)
		log.Printf("Server starting on %s", address)
		if err := http.ListenAndServe(address, router); err != nil {
			errChan <- fmt.Errorf("HTTP server: %w", err)
			return
		}
	}()

	// Wait for errors from either goroutine
	for err := range errChan {
		if err != nil {
			log.Printf("Error: %v", err)
			os.Exit(-1)
		}
	}
}
