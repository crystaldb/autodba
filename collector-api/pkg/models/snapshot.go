package models

type Snapshot struct {
	ID          int64  `json:"id"`
	CollectedAt int64  `json:"collected_at"`
	S3Location  string `json:"local_dir"`
}

type CompactSnapshot struct {
	ID          int64  `json:"id"`
	CollectedAt int64  `json:"collected_at"`
	S3Location  string `json:"local_dir"`
}
