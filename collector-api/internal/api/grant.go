package api

import (
	"collector-api/internal/auth"
	"collector-api/internal/config"
	"collector-api/internal/storage"
	"collector-api/pkg/models"
	"encoding/json"
	"log"
	"net/http"
	"os"
)

func GrantHandler(w http.ResponseWriter, r *http.Request) {
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
		log.Printf("Authenticated request from %s", r.RemoteAddr)
	}

	selfURL := os.Getenv("COLLECTOR_API_URL")
	if selfURL == "" {
		selfURL = "http://localhost:7080" // fallback to default if not set
	}

	// Populate GrantConfig with dummy data (replace this with actual server config)
	grantConfig := models.GrantConfig{
		ServerID:         "pgServer1",
		ServerURL:        selfURL,
		SentryDsn:        "",
		EnableActivity:   true,
		EnableLogs:       false,
		SchemaTableLimit: 0,
		Features: models.GrantFeatures{
			Logs:                        false,
			StatementResetFrequency:     0,
			StatementTimeoutMs:          0,
			StatementTimeoutMsQueryText: 0,
		},
	}

	// Respond with the Snapshot grant
	grant := models.Grant{
		Valid:    true,
		Config:   grantConfig,
		LocalDir: storage.GetLocalStorageDir(),
		S3URL:    selfURL + "/v2/upload",
		S3Fields: map[string]string{"key": cfg.APIKey},
	}

	// Respond with the grant
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(grant); err != nil {
		if cfg.Debug {
			log.Printf("Error encoding grant response: %v", err)
		}
		http.Error(w, "Failed to encode response", http.StatusInternalServerError)
		return
	}

	if cfg.Debug {
		log.Printf("Grant response successfully sent to %s", r.RemoteAddr)
	}
}
