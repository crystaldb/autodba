package api

import (
	"collector-api/internal/config"
	"net/http"

	"github.com/gorilla/mux"
)

// SetupRoutes defines the API routes and attaches the middleware
func SetupRoutes(cfg *config.Config) *mux.Router {
	router := mux.NewRouter()

	// Attach the logging middleware, passing the debug flag from the config
	router.Use(func(next http.Handler) http.Handler {
		return LoggingMiddleware(next, cfg.Debug)
	})

	// Define the routes
	router.HandleFunc("/v2/snapshots/grant", GrantHandler).Methods("GET")
	router.HandleFunc("/v2/snapshots", SnapshotHandler).Methods("POST")
	router.HandleFunc("/v2/snapshots/compact", CompactSnapshotHandler).Methods("POST")
	router.HandleFunc("/v2/upload", UploadHandler).Methods("POST")

	return router
}
