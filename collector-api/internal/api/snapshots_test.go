package api

import (
	"testing"

	collector_proto "github.com/pganalyze/collector/output/pganalyze_collector"
	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestProcessFullSnapshotData(t *testing.T) {
	testCases := []struct {
		name     string
		filename string
		systemID string
	}{
		{
			name:     "RDS Full Snapshot",
			filename: "test_data/full-snapshot-rds-1.binpb",
			systemID: "rds-instance-100",
		},
		{
			name:     "Aurora Full Snapshot",
			filename: "test_data/full-snapshot-aurora-1.binpb",
			systemID: "aurora-instance-100",
		},
		{
			name:     "RDS Full Snapshot",
			filename: "test_data/full-snapshot-rds-2.binpb",
			systemID: "rds-instance-100",
		},
		{
			name:     "Aurora Full Snapshot",
			filename: "test_data/full-snapshot-aurora-2.binpb",
			systemID: "aurora-instance-100",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a mock SystemInfo
			systemInfo := SystemInfo{
				SystemID:    tc.systemID,
				SystemScope: "test-scope",
				SystemType:  "amazon_rds",
			}

			// Call processFullSnapshotData
			allMetrics, err := processFullSnapshotData(tc.filename, systemInfo)
			assert.NoError(t, err)

			// for _, metric := range allMetrics {
			// 	t.Logf("Metric: %s", metric.String())
			// }

			// Verify that we got some metrics
			assert.NotEmpty(t, allMetrics)

			pbBytes, err := readAndDecompressSnapshot(tc.filename)
			assert.NoError(t, err)

			// Unmarshal the snapshot to verify specific metrics
			var fullSnapshot collector_proto.FullSnapshot
			err = proto.Unmarshal(pbBytes, &fullSnapshot)
			assert.NoError(t, err)

			// jsonData, err := protojson.Marshal(&fullSnapshot)
			// if err != nil {
			// 	t.Errorf("Failed to marshal fullSnapshot to JSON: %v", err)
			// } else {
			// 	t.Logf("FullSnapshot JSON: %s", jsonData)
			// }

			// Verify system metrics
			if fullSnapshot.System != nil {
				if fullSnapshot.System.CpuStatistics != nil {
					assert.True(t, containsMetric(allMetrics, "cc_system_cpu_user_percent"))
				}

				if fullSnapshot.System.MemoryStatistic != nil {
					assert.True(t, containsMetric(allMetrics, "cc_system_memory_total_bytes"))
				}
			}

			// Verify database metrics
			if len(fullSnapshot.DatabaseStatictics) > 0 {
				assert.True(t, containsMetric(allMetrics, "cc_db_xact_commit"))
			}

			// Verify query metrics
			if len(fullSnapshot.QueryStatistics) > 0 {
				assert.True(t, containsMetric(allMetrics, "cc_query_calls"))
			}

			// Verify relation metrics
			if len(fullSnapshot.RelationStatistics) > 0 {
				assert.True(t, containsMetric(allMetrics, "cc_relation_size_bytes"))
			}

			// Verify index metrics
			if len(fullSnapshot.IndexStatistics) > 0 {
				assert.True(t, containsMetric(allMetrics, "cc_index_size_bytes"))
			}
		})
	}
}

func containsMetric(metrics []prompb.TimeSeries, metricName string) bool {
	for _, metric := range metrics {
		for _, label := range metric.Labels {
			if label.Name == "__name__" && label.Value == metricName {
				return true
			}
		}
	}
	return false
}
