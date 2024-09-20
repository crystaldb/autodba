package models

// GrantLogs represents a grant for accessing log directories
type GrantLogs struct {
	Valid    bool   `json:"valid"`
	LocalDir string `json:"local_dir"`
}
