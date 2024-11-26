package storage

type QueryStorage interface {
	StoreQuery(fingerprint, query, fullQuery string) error
	GetQuery(fingerprint string) (string, error)
	GetFullQuery(fingerprint string) (string, error)
}
