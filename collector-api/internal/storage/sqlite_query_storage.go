package storage

import (
	"collector-api/internal/db"
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

var QueryStore QueryStorage

type SQLiteQueryStorage struct {
	db *sql.DB
}

func InitQueryStorage(dbPath string) error {
	storage, err := NewSQLiteQueryStorage(dbPath)
	if err != nil {
		return err
	}

	QueryStore = storage
	return nil
}

func NewSQLiteQueryStorage(dbPath string) (*SQLiteQueryStorage, error) {
	// Initialize the SQLite database
	database, err := db.InitDB(dbPath)
	if err != nil {
		return nil, err
	}

	storage := &SQLiteQueryStorage{db: database}
	if err := storage.initTables(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *SQLiteQueryStorage) initTables() error {
	_, err := s.db.Exec(`
        CREATE TABLE IF NOT EXISTS queries (
            fingerprint TEXT PRIMARY KEY,
            query TEXT,
            last_update INTEGER
        );
        CREATE TABLE IF NOT EXISTS full_queries (
            fingerprint TEXT PRIMARY KEY,
            full_query TEXT,
            last_update INTEGER
        );
    `)
	return err
}

func (s *SQLiteQueryStorage) StoreQuery(fingerprint, query, fullQuery string, collectedAt int64) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT OR REPLACE INTO queries (fingerprint, query, last_update) VALUES (?, ?, ?)", fingerprint, query, collectedAt)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("INSERT OR REPLACE INTO full_queries (fingerprint, full_query, last_update) VALUES (?, ?, ?)", fingerprint, fullQuery, collectedAt)
	if err != nil {
		tx.Rollback()
		return err
	}

	return tx.Commit()
}

func (s *SQLiteQueryStorage) GetQuery(fingerprint string) (string, error) {
	var query string
	err := s.db.QueryRow("SELECT query FROM queries WHERE fingerprint = ?", fingerprint).Scan(&query)
	return query, err
}

func (s *SQLiteQueryStorage) GetFullQuery(fingerprint string) (string, error) {
	var fullQuery string
	err := s.db.QueryRow("SELECT full_query FROM full_queries WHERE fingerprint = ?", fingerprint).Scan(&fullQuery)
	return fullQuery, err
}