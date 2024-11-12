package models

type Snapshot struct {
	ID          int64  `json:"id"`
	CollectedAt int64  `json:"collected_at"`
	S3Location  string `json:"local_dir"`
	SystemID    string `json:"system_id"`
	SystemScope string `json:"system_scope"`
	SystemType  string `json:"system_type"`
}

type CompactSnapshot struct {
	ID          int64  `json:"id"`
	CollectedAt int64  `json:"collected_at"`
	S3Location  string `json:"local_dir"`
	SystemID    string `json:"system_id"`
	SystemScope string `json:"system_scope"`
	SystemType  string `json:"system_type"`
}
