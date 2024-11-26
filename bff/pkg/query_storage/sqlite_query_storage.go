package query_storage

import (
	"database/sql"

	_ "github.com/mattn/go-sqlite3"
)

type SQLiteQueryStorage struct {
	db *sql.DB
}

func NewSQLiteQueryStorage(dbPath string) (*SQLiteQueryStorage, error) {
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, err
	}

	storage := &SQLiteQueryStorage{db: db}

	return storage, nil
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
