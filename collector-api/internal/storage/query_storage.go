package storage

type QueryStorage interface {
	StoreQuery(fingerprint, query, fullQuery string, collectedAt int64) error
	GetQuery(fingerprint string) (string, error)
	GetFullQuery(fingerprint string) (string, error)
}
