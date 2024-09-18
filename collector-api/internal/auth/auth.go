package auth

import (
	"collector-api/internal/config"
	"net/http"
)

func Authenticate(r *http.Request) bool {
	cfg, _ := config.LoadConfig()
	apiKey := r.Header.Get("Pganalyze-Api-Key")
	return apiKey == cfg.APIKey
}
