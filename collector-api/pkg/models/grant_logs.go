package models

// GrantLogs represents a grant for accessing log directories
// GrantLogsEncryptionKey represents encryption details for logs
type GrantLogsEncryptionKey struct {
	CiphertextBlob string `json:"ciphertext_blob"`
	KeyId          string `json:"key_id"`
	Plaintext      string `json:"plaintext"`
}

// GrantS3 represents S3-related details for logs and snapshots
type GrantS3 struct {
	S3URL    string            `json:"s3_url"`
	S3Fields map[string]string `json:"s3_fields"`
}

// GrantLogs represents the structure of the grant for log access
type GrantLogs struct {
	Valid         bool                   `json:"valid"`
	Config        GrantConfig            `json:"config"`
	Logdata       GrantS3                `json:"logdata"`
	Snapshot      GrantS3                `json:"snapshot"`
	EncryptionKey GrantLogsEncryptionKey `json:"encryption_key"`
}
