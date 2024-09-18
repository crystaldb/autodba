package api

import (
	"collector-api/internal/auth"
	"collector-api/internal/db"
	"collector-api/internal/storage"
	"collector-api/pkg/models"
	"encoding/json"
	"net/http"
)

func SubmitSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.Authenticate(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var snapshot models.Snapshot
	err := json.NewDecoder(r.Body).Decode(&snapshot)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Store the snapshot metadata
	err = db.StoreSnapshotMetadata(snapshot)
	if err != nil {
		http.Error(w, "Error storing snapshot", http.StatusInternalServerError)
		return
	}

	// Move the snapshot to local storage
	err = storage.StoreSnapshot(snapshot)
	if err != nil {
		http.Error(w, "Error storing snapshot data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func SubmitCompactSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	if !auth.Authenticate(r) {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	var snapshot models.CompactSnapshot
	err := json.NewDecoder(r.Body).Decode(&snapshot)
	if err != nil {
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Store the compact snapshot metadata
	err = db.StoreCompactSnapshotMetadata(snapshot)
	if err != nil {
		http.Error(w, "Error storing compact snapshot", http.StatusInternalServerError)
		return
	}

	// Move the compact snapshot to local storage
	err = storage.StoreCompactSnapshot(snapshot)
	if err != nil {
		http.Error(w, "Error storing compact snapshot data", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
}
