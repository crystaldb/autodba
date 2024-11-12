package db

import (
	"collector-api/pkg/models"
	"database/sql"
	"fmt"
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
		s3_location TEXT,
        system_id TEXT,
        system_scope TEXT,
        system_type TEXT
	);`

	createCompactSnapshotTable := `
	CREATE TABLE IF NOT EXISTS compact_snapshots (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		collected_at INTEGER,
		s3_location TEXT,
        system_id TEXT,
        system_scope TEXT,
        system_type TEXT
	);`

	_, err := db.Exec(createSnapshotTable)
	if err != nil {
		log.Fatalf("Error creating snapshots table: %v", err)
	}

	_, err = db.Exec(createCompactSnapshotTable)
	if err != nil {
		log.Fatalf("Error creating compact_snapshots table: %v", err)
	}

	// Check if we need to add new columns
	if err := addColumnsIfNotExist("snapshots", []string{"system_id", "system_scope", "system_type"}); err != nil {
		log.Fatalf("Error adding columns to snapshots table: %v", err)
	}

	if err := addColumnsIfNotExist("compact_snapshots", []string{"system_id", "system_scope", "system_type"}); err != nil {
		log.Fatalf("Error adding columns to compact_snapshots table: %v", err)
	}
}

func addColumnsIfNotExist(table string, columns []string) error {
	for _, col := range columns {
		var count int
		query := fmt.Sprintf("SELECT COUNT(*) FROM pragma_table_info('%s') WHERE name='%s';", table, col)
		err := db.QueryRow(query).Scan(&count)
		if err != nil {
			return fmt.Errorf("check column existence: %w", err)
		}

		if count == 0 {
			_, err = db.Exec(fmt.Sprintf("ALTER TABLE %s ADD COLUMN %s TEXT;", table, col))
			if err != nil {
				return fmt.Errorf("add column %s: %w", col, err)
			}
		}
	}
	return nil
}

func StoreSnapshotMetadata(snapshot models.Snapshot) error {
	_, err := db.Exec(`
        INSERT INTO snapshots (collected_at, s3_location, system_id, system_scope, system_type)
        VALUES (?, ?, ?, ?, ?)`,
		snapshot.CollectedAt, snapshot.S3Location, snapshot.SystemID, snapshot.SystemScope, snapshot.SystemType)
	return err
}

func StoreCompactSnapshotMetadata(snapshot models.CompactSnapshot) error {
	_, err := db.Exec(`
        INSERT INTO compact_snapshots (collected_at, s3_location, system_id, system_scope, system_type)
        VALUES (?, ?, ?, ?, ?)`,
		snapshot.CollectedAt, snapshot.S3Location, snapshot.SystemID, snapshot.SystemScope, snapshot.SystemType)
	return err
}

func GetAllFullSnapshots() ([]models.Snapshot, error) {
	rows, err := db.Query(`
        SELECT DISTINCT collected_at, s3_location, system_id, system_scope, system_type 
        FROM snapshots 
        ORDER BY collected_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []models.Snapshot
	for rows.Next() {
		var s models.Snapshot
		var systemID, systemScope, systemType sql.NullString
		if err := rows.Scan(
			&s.CollectedAt,
			&s.S3Location,
			&systemID,
			&systemScope,
			&systemType,
		); err != nil {
			return nil, err
		}
		// Convert NullString to string, using empty string if NULL
		s.SystemID = systemID.String
		s.SystemScope = systemScope.String
		s.SystemType = systemType.String
		snapshots = append(snapshots, s)
	}
	return snapshots, rows.Err()
}

func GetAllCompactSnapshots() ([]models.CompactSnapshot, error) {
	rows, err := db.Query(`
        SELECT DISTINCT collected_at, s3_location, system_id, system_scope, system_type 
        FROM compact_snapshots 
        ORDER BY collected_at ASC`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var snapshots []models.CompactSnapshot
	for rows.Next() {
		var s models.CompactSnapshot
		var systemID, systemScope, systemType sql.NullString
		if err := rows.Scan(
			&s.CollectedAt,
			&s.S3Location,
			&systemID,
			&systemScope,
			&systemType,
		); err != nil {
			return nil, err
		}
		// Convert NullString to string, using empty string if NULL
		s.SystemID = systemID.String
		s.SystemScope = systemScope.String
		s.SystemType = systemType.String
		snapshots = append(snapshots, s)
	}
	return snapshots, rows.Err()
}
