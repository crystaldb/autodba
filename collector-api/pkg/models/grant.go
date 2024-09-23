package models

// GrantConfig represents configuration settings for logs and snapshots
type GrantConfig struct {
	ServerID  string        `json:"server_id"`
	ServerURL string        `json:"server_url"`
	SentryDsn string        `json:"sentry_dsn"`
	Features  GrantFeatures `json:"features"`

	EnableActivity bool `json:"enable_activity"`
	EnableLogs     bool `json:"enable_logs"`

	SchemaTableLimit int `json:"schema_table_limit"` // Maximum number of tables that can be monitored per server
}

// GrantFeatures represents the features related to logs and statements
type GrantFeatures struct {
	Logs                        bool  `json:"logs"`
	StatementResetFrequency     int   `json:"statement_reset_frequency"`
	StatementTimeoutMs          int32 `json:"statement_timeout_ms"`
	StatementTimeoutMsQueryText int32 `json:"statement_timeout_ms_query_text"`
}

// Grant represents the structure of the grant for accessing snapshots/logs
type Grant struct {
	Valid    bool              `json:"valid"`
	Config   GrantConfig       `json:"config"`
	S3URL    string            `json:"s3_url"`
	S3Fields map[string]string `json:"s3_fields"`
	LocalDir string            `json:"local_dir"`
}
