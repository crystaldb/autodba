package auth

import (
	"collector-api/internal/config"
	"net/http"
)

func Authenticate(r *http.Request) bool {
	cfg, _ := config.LoadConfigWithDefaultPath()
	apiKey := r.Header.Get("Pganalyze-Api-Key")
	return apiKey == cfg.APIKey
}

func S3Authenticate(key string) bool {
	cfg, _ := config.LoadConfigWithDefaultPath()
	return key == cfg.APIKey
}
