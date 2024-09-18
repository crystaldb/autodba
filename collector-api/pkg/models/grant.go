package models

type Grant struct {
	Valid    bool   `json:"valid"`
	LocalDir string `json:"local_dir"`
	S3URL    string `json:"s3_url,omitempty"`
}
