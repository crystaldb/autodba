package storage

import (
	"collector-api/internal/config"
	"collector-api/pkg/models"
	"fmt"
	"os"
	"path/filepath"
)

// GetLocalStorageDir returns the directory for storing snapshot files
func GetLocalStorageDir() string {
	cfg, _ := config.LoadConfigWithDefaultPath()
	return cfg.StorageDir
}

// StoreSnapshot stores a regular snapshot to the local directory
func StoreSnapshot(snapshot models.Snapshot) error {
	storageDir := GetLocalStorageDir()

	// Create file path based on snapshot ID
	filePath := filepath.Join(storageDir, fmt.Sprintf("snapshot-%d.json", snapshot.ID))
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Serialize and save the snapshot (pseudo-code)
	// You need to serialize the snapshot as JSON or another format based on your needs
	// err = json.NewEncoder(file).Encode(snapshot)
	if err != nil {
		return err
	}
	return nil
}

// StoreCompactSnapshot stores a compact snapshot to the local directory
func StoreCompactSnapshot(snapshot models.CompactSnapshot) error {
	storageDir := GetLocalStorageDir()

	// Create file path based on compact snapshot ID
	filePath := filepath.Join(storageDir, fmt.Sprintf("compact_snapshot-%d.json", snapshot.ID))
	file, err := os.Create(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	// Serialize and save the compact snapshot (pseudo-code)
	// err = json.NewEncoder(file).Encode(snapshot)
	if err != nil {
		return err
	}
	return nil
}
