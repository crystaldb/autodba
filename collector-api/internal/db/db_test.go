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
		SystemID:    "test-system",
		SystemScope: "test-scope",
		SystemType:  "self_hosted",
	}

	// Store snapshot metadata
	err := db.StoreSnapshotMetadata(snapshot)
	assert.NoError(t, err)

	// Verify the data was stored correctly
	snapshots, err := db.GetAllFullSnapshots()
	assert.NoError(t, err)
	assert.Len(t, snapshots, 1)
	assert.Equal(t, snapshot, snapshots[0])
}

func TestStoreCompactSnapshotMetadata(t *testing.T) {
	// Initialize in-memory SQLite database for testing
	dbPath := ":memory:" // Use in-memory database for testing
	db.InitDB(dbPath)

	// Create a sample compact snapshot
	snapshot := models.CompactSnapshot{
		CollectedAt: 1234567890,
		S3Location:  "/test/dir",
		SystemID:    "test-system",
		SystemScope: "test-scope",
		SystemType:  "self_hosted",
	}

	// Store compact snapshot metadata
	err := db.StoreCompactSnapshotMetadata(snapshot)
	assert.NoError(t, err)

	snapshots, err := db.GetAllCompactSnapshots()
	assert.NoError(t, err)
	assert.Len(t, snapshots, 1)
	assert.Equal(t, snapshot, snapshots[0])
}
