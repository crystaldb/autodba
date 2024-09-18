package api

import (
	"collector-api/internal/auth"
	"collector-api/internal/storage"
	"collector-api/pkg/models"
	"encoding/json"
	"net/http"
)

func GrantHandler(w http.ResponseWriter, r *http.Request) {
	// Authenticate the request
	if !auth.Authenticate(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	// Generate a local directory grant for Phase 1
	grant := models.Grant{
		Valid:    true,
		LocalDir: storage.GetLocalStorageDir(),
	}

	// Respond with the grant
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(grant)
}
