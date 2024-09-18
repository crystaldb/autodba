package models

type Snapshot struct {
	ID          int64  `json:"id"`
	CollectedAt int64  `json:"collected_at"`
	LocalDir    string `json:"local_dir"`
	Status      string `json:"status"`
}

type CompactSnapshot struct {
	ID          int64  `json:"id"`
	CollectedAt int64  `json:"collected_at"`
	LocalDir    string `json:"local_dir"`
	Type        string `json:"snapshot_type"`
}
