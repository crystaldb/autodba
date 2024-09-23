package db

import (
	"collector-api/pkg/models"
	"database/sql"
	"log"
	"os"

	_ "github.com/mattn/go-sqlite3" // SQLite driver
)

var db *sql.DB

func InitDB(dbPath string) {
	var err error

	// Create the SQLite database file if it doesn't exist
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		file, err := os.Create(dbPath)
		if err != nil {
			log.Fatalf("Failed to create database file: %v", err)
		}
		file.Close()
	}

	// Open the SQLite database
	db, err = sql.Open("sqlite3", dbPath)
	if err != nil {
		log.Fatalf("Failed to connect to SQLite database: %v", err)
	}

	// Initialize database schema
	initSchema()
}

func initSchema() {
	// Create the necessary tables for snapshots
	createSnapshotTable := `
	CREATE TABLE IF NOT EXISTS snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		collected_at INTEGER,
		s3_location TEXT
	);`

	createCompactSnapshotTable := `
	CREATE TABLE IF NOT EXISTS compact_snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		collected_at INTEGER,
		s3_location TEXT
	);`

	_, err := db.Exec(createSnapshotTable)
	if err != nil {
		log.Fatalf("Error creating snapshots table: %v", err)
	}

	_, err = db.Exec(createCompactSnapshotTable)
	if err != nil {
		log.Fatalf("Error creating compact_snapshots table: %v", err)
	}
}

func StoreSnapshotMetadata(snapshot models.Snapshot) error {
	// Insert metadata into the SQLite database
	_, err := db.Exec(`
		INSERT INTO snapshots (collected_at, s3_location)
		VALUES (?, ?)`,
		snapshot.CollectedAt, snapshot.S3Location)
	return err
}

func StoreCompactSnapshotMetadata(snapshot models.CompactSnapshot) error {
	// Insert compact snapshot metadata
	_, err := db.Exec(`
		INSERT INTO compact_snapshots (collected_at, s3_location)
		VALUES (?, ?)`,
		snapshot.CollectedAt, snapshot.S3Location)
	return err
}
