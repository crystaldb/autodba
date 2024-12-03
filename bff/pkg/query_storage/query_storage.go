package query_storage

type QueryStorage interface {
	GetQuery(fingerprint string) (string, error)
	GetFullQuery(fingerprint string) (string, error)
}
