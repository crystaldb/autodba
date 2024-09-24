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
						{Name: "client_port", Value: "5432"},
						{Name: "client_addr", Value: "192.168.1.1"},
						{Name: "pid", Value: "12345"},
						{Name: "query", Value: "SELECT * FROM users;"},
						{Name: "state", Value: "active"},
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
		{
			name: "Multiple backends",
			snapshot: &pganalyze_collector.CompactSnapshot{
				Data: &pganalyze_collector.CompactSnapshot_ActivitySnapshot{
					ActivitySnapshot: &pganalyze_collector.CompactActivitySnapshot{
						Backends: []*pganalyze_collector.Backend{
							{
								ApplicationName: "pgbench",
								BackendType:     "client backend",
								ClientPort:      5433,
								ClientAddr:      "192.168.1.2",
								Pid:             54321,
								QueryText:       "UPDATE users SET name='Alice';",
								State:           "idle",
								WaitEvent:       "ClientRead",
								WaitEventType:   "Client",
							},
							{
								ApplicationName: "pgbench",
								BackendType:     "client backend",
								ClientPort:      5434,
								ClientAddr:      "192.168.1.3",
								Pid:             54322,
								QueryText:       "SELECT name FROM users;",
								State:           "active",
								WaitEvent:       "transactionid",
								WaitEventType:   "Lock",
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
						{Name: "application_name", Value: "pgbench"},
						{Name: "backend_type", Value: "client backend"},
						{Name: "client_port", Value: "5433"},
						{Name: "client_addr", Value: "192.168.1.2"},
						{Name: "pid", Value: "54321"},
						{Name: "query", Value: "UPDATE users SET name='Alice';"},
						{Name: "state", Value: "idle"},
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
				{
					Labels: []prompb.Label{
						{Name: "__name__", Value: "cc_pg_stat_activity"},
						{Name: "application_name", Value: "pgbench"},
						{Name: "backend_type", Value: "client backend"},
						{Name: "client_port", Value: "5434"},
						{Name: "client_addr", Value: "192.168.1.3"},
						{Name: "pid", Value: "54322"},
						{Name: "query", Value: "SELECT name FROM users;"},
						{Name: "state", Value: "active"},
						{Name: "wait_event", Value: "transactionid"},
						{Name: "wait_event_type", Value: "Lock"},
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
			result := compactSnapshotMetrics(tt.snapshot)

			for i := range tt.expected {
				// Compare the overall length of the result
				assert.Equal(t, len(tt.expected), len(result), "Number of time series mismatch")

				// Compare the number of samples
				assert.Equal(t, len(tt.expected[i].Samples), len(result[i].Samples), "Number of samples mismatch")

				// Compare sample values
				assert.Equal(t, tt.expected[i].Samples[0].Value, result[i].Samples[0].Value, "Sample value mismatch")

				// Compare the timestamps
				if tt.expected[i].Samples[0].Timestamp != tt.snapshot.CollectedAt.AsTime().UnixMilli() {
					t.Errorf("timestamp mismatch: expected %d, got %d", tt.expected[i].Samples[0].Timestamp, tt.snapshot.CollectedAt.AsTime().UnixMilli())
				}

				// Compare the number of labels
				assert.Equal(t, len(tt.expected[i].Labels), len(result[i].Labels), "Number of labels mismatch")

				// Compare each label
				for j, expectedLabel := range tt.expected[i].Labels {
					assert.Equal(t, expectedLabel.Name, result[i].Labels[j].Name, "Label name mismatch")
					assert.Equal(t, expectedLabel.Value, result[i].Labels[j].Value, "Label value mismatch")
				}
			}
		})
	}
}
