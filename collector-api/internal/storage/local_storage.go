package storage

import (
	"collector-api/internal/config"
	"collector-api/pkg/models"
	"encoding/json"
	"fmt"
	"log"
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
	cfg, err := config.LoadConfigWithDefaultPath()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	storageDir := GetLocalStorageDir()

	// Create file path based on snapshot ID
	filePath := filepath.Join(storageDir, fmt.Sprintf("snapshot-%d.json", snapshot.ID))

	if cfg.Debug {
		log.Printf("Storing snapshot: ID=%d, Path=%s", snapshot.ID, filePath)
	}

	file, err := os.Create(filePath)
	if err != nil {
		if cfg.Debug {
			log.Printf("Error creating file for snapshot: %v", err)
		}
		return err
	}
	defer file.Close()

	// Serialize and save the snapshot
	if err := json.NewEncoder(file).Encode(snapshot); err != nil {
		if cfg.Debug {
			log.Printf("Error encoding snapshot: %v", err)
		}
		return err
	}

	if cfg.Debug {
		log.Printf("Snapshot successfully stored at %s", filePath)
	}

	return nil
}

// StoreCompactSnapshot stores a compact snapshot to the local directory
func StoreCompactSnapshot(snapshot models.CompactSnapshot) error {
	cfg, err := config.LoadConfigWithDefaultPath()
	if err != nil {
		return fmt.Errorf("failed to load config: %v", err)
	}

	storageDir := GetLocalStorageDir()

	// Create file path based on compact snapshot ID
	filePath := filepath.Join(storageDir, fmt.Sprintf("compact_snapshot-%d.json", snapshot.ID))

	if cfg.Debug {
		log.Printf("Storing compact snapshot: ID=%d, Path=%s", snapshot.ID, filePath)
	}

	file, err := os.Create(filePath)
	if err != nil {
		if cfg.Debug {
			log.Printf("Error creating file for compact snapshot: %v", err)
		}
		return err
	}
	defer file.Close()

	// Serialize and save the compact snapshot
	if err := json.NewEncoder(file).Encode(snapshot); err != nil {
		if cfg.Debug {
			log.Printf("Error encoding compact snapshot: %v", err)
		}
		return err
	}

	if cfg.Debug {
		log.Printf("Compact snapshot successfully stored at %s", filePath)
	}

	return nil
}

// EnsureStorageDirectories ensures the required subdirectories exist under the base storage directory
func EnsureStorageDirectories(baseDir string) error {
	subdirs := []string{
		baseDir,
	}

	for _, subdir := range subdirs {
		err := os.MkdirAll(subdir, os.ModePerm)
		if err != nil {
			log.Printf("Failed to create directory %s: %v", subdir, err)
			return err
		}
		log.Printf("Created/verified storage directory: %s", subdir)
	}

	return nil
}
