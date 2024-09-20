package api

import (
	"collector-api/internal/auth"
	"collector-api/internal/config"
	"collector-api/internal/storage"
	"collector-api/pkg/models"
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
)

func GrantLogsHandler(w http.ResponseWriter, r *http.Request) {
	// Load configuration
	cfg, err := config.LoadConfigWithDefaultPath()
	if err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		if cfg.Debug {
			log.Printf("Error loading config: %v", err)
		}
		return
	}

	// Authenticate the request
	if !auth.Authenticate(r) {
		if cfg.Debug {
			log.Printf("Unauthorized access attempt from %s", r.RemoteAddr)
		}
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if cfg.Debug {
		log.Printf("Authenticated request for logs grant from %s", r.RemoteAddr)
	}

	// Define the logs subdirectory dynamically
	logsDir := filepath.Join(storage.GetLocalStorageDir(), "logs")

	if cfg.Debug {
		log.Printf("Logs directory resolved to: %s", logsDir)
	}

	// Respond with the Logs storage subdirectory
	grantLogs := models.GrantLogs{
		Valid:    true,
		LocalDir: logsDir, // Use the logs subdirectory
	}

	// Respond with the logs grant
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(grantLogs); err != nil {
		if cfg.Debug {
			log.Printf("Error encoding grant logs response: %v", err)
		}
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	if cfg.Debug {
		log.Printf("Logs grant response successfully sent to %s", r.RemoteAddr)
	}
}
