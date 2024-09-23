package db_test

import (
	"collector-api/internal/db"
	"collector-api/pkg/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStoreSnapshotMetadata(t *testing.T) {
	// Initialize in-memory SQLite database for testing
	dbPath := ":memory:" // Use in-memory database for testing
	db.InitDB(dbPath)

	// Create a sample snapshot
	snapshot := models.Snapshot{
		CollectedAt: 1234567890,
		S3Location:  "/test/dir",
	}

	// Store snapshot metadata
	err := db.StoreSnapshotMetadata(snapshot)
	assert.NoError(t, err)

	// Check that snapshot was inserted (you can extend this with SELECT queries)
}

func TestStoreCompactSnapshotMetadata(t *testing.T) {
	// Initialize in-memory SQLite database for testing
	dbPath := ":memory:" // Use in-memory database for testing
	db.InitDB(dbPath)

	// Create a sample compact snapshot
	snapshot := models.CompactSnapshot{
		CollectedAt: 1234567890,
		S3Location:  "/test/dir",
	}

	// Store compact snapshot metadata
	err := db.StoreCompactSnapshotMetadata(snapshot)
	assert.NoError(t, err)

	// Check that compact snapshot was inserted (you can extend this with SELECT queries)
}
