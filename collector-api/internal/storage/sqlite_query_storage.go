package storage

import (
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
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	storage := &SQLiteQueryStorage{db: db}
	if err := storage.initTables(); err != nil {
		return nil, err
	}

	return storage, nil
}

func (s *SQLiteQueryStorage) initTables() error {
	_, err := s.db.Exec(`
        CREATE TABLE IF NOT EXISTS queries (
            fingerprint TEXT PRIMARY KEY,
            query TEXT
        );
        CREATE TABLE IF NOT EXISTS full_queries (
            fingerprint TEXT PRIMARY KEY,
            full_query TEXT
        );
    `)
	return err
}

func (s *SQLiteQueryStorage) StoreQuery(fingerprint, query, fullQuery string) error {
	tx, err := s.db.Begin()
	if err != nil {
		return err
	}

	_, err = tx.Exec("INSERT OR REPLACE INTO queries (fingerprint, query) VALUES (?, ?)", fingerprint, query)
	if err != nil {
		tx.Rollback()
		return err
	}

	_, err = tx.Exec("INSERT OR REPLACE INTO full_queries (fingerprint, full_query) VALUES (?, ?)", fingerprint, fullQuery)
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
