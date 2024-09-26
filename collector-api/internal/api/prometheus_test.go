package api

import (
	"testing"
	"time"

	"github.com/pganalyze/collector/output/pganalyze_collector"
	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestCompactSnapshotMetrics(t *testing.T) {
	now := time.Now()

	systemInfo := map[string]string{
		"system_id":             "test-sys-id",
		"system_id_fallback":    "test-sys-id-fallback",
		"system_scope":          "test-sys-scope",
		"system_scope_fallback": "test-sys-scope-fallback",
		"system_type":           "test-sys-type",
		"system_type_fallback":  "test-sys-type-fallback",
	}

	tests := []struct {
		name     string
		snapshot *pganalyze_collector.CompactSnapshot
		expected []prompb.TimeSeries
	}{
		{
			name: "Single backend with all fields",
			snapshot: &pganalyze_collector.CompactSnapshot{
				Data: &pganalyze_collector.CompactSnapshot_ActivitySnapshot{
					ActivitySnapshot: &pganalyze_collector.CompactActivitySnapshot{
						Backends: []*pganalyze_collector.Backend{
							{
								ApplicationName: "PostgreSQL JDBC Driver",
								BackendType:     "client backend",
								ClientPort:      5432,
								ClientAddr:      "192.168.1.1",
								Pid:             12345,
								QueryText:       "SELECT * FROM users;",
								State:           "active",
								WaitEvent:       "ClientRead",
								WaitEventType:   "Client",
							},
						},
					},
				},
				CollectedAt: timestamppb.New(now),
			},
			expected: []prompb.TimeSeries{
				{
					Labels: []prompb.Label{
						{Name: "__name__", Value: "cc_pg_stat_activity"},
						{Name: "application_name", Value: "PostgreSQL JDBC Driver"},
						{Name: "backend_type", Value: "client backend"},
						{Name: "client_addr", Value: "192.168.1.1"},
						{Name: "client_port", Value: "5432"},
						{Name: "pid", Value: "12345"},
						{Name: "query", Value: "SELECT * FROM users;"},
						{Name: "state", Value: "active"},
						{Name: "sys_id", Value: "test-sys-id"},
						{Name: "sys_id_fallback", Value: "test-sys-id-fallback"},
						{Name: "sys_scope", Value: "test-sys-scope"},
						{Name: "sys_scope_fallback", Value: "test-sys-scope-fallback"},
						{Name: "sys_type", Value: "test-sys-type"},
						{Name: "sys_type_fallback", Value: "test-sys-type-fallback"},
						{Name: "wait_event", Value: "ClientRead"},
						{Name: "wait_event_type", Value: "Client"},
					},
					Samples: []prompb.Sample{
						{
							Timestamp: now.UnixMilli(),
							Value:     1.0,
						},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, _ := compactSnapshotMetrics(tt.snapshot, systemInfo)
			assert.Equal(t, tt.expected, got)
		})
	}
}
