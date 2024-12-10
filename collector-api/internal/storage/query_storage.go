package storage

type QueryRep struct {
	Fingerprint string
	Query       string
	QueryFull   string
	CollectedAt int64
}

type QueryStorage interface {
	StoreQuery(fingerprint, query, fullQuery string, collectedAt int64) error
	GetQuery(fingerprint string) (string, error)
	GetFullQuery(fingerprint string) (string, error)
	StoreBatchQueries(queries []QueryRep) error
}
