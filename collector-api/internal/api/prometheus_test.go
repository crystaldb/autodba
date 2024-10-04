package api

import (
	"testing"
	"time"

	"github.com/pganalyze/collector/output/pganalyze_collector"
	"github.com/prometheus/prometheus/prompb"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Helper function to create a test SystemInfo
func createTestSystemInfo(id string) SystemInfo {
	return SystemInfo{
		SystemID:            id,
		SystemIDFallback:    id + "-fallback",
		SystemScope:         "test-scope-" + id,
		SystemScopeFallback: "test-scope-fallback-" + id,
		SystemType:          "test-type",
		SystemTypeFallback:  "test-type-fallback",
	}
}

// Helper function to create a test Backend
func createTestBackend(pid int32, appName, state string) *pganalyze_collector.Backend {
	return &pganalyze_collector.Backend{
		Pid:             pid,
		ApplicationName: appName,
		State:           state,
		ClientAddr:      "127.0.0.1",
		ClientPort:      5432,
		QueryText:       "SELECT * FROM test",
		WaitEvent:       "ClientRead",
		WaitEventType:   "Client",
	}
}

func TestCompactSnapshotMetrics(t *testing.T) {
	now := time.Now()
	sysInfo := createTestSystemInfo("test-system-1")

	testCases := []struct {
		name     string
		snapshot *pganalyze_collector.CompactSnapshot
		expected int // Number of expected time series
	}{
		{
			name: "Single backend",
			snapshot: &pganalyze_collector.CompactSnapshot{
				Data: &pganalyze_collector.CompactSnapshot_ActivitySnapshot{
					ActivitySnapshot: &pganalyze_collector.CompactActivitySnapshot{
						Backends: []*pganalyze_collector.Backend{
							createTestBackend(1234, "test-app", "active"),
						},
					},
				},
				CollectedAt: timestamppb.New(now),
			},
			expected: 1,
		},
		{
			name: "Multiple backends",
			snapshot: &pganalyze_collector.CompactSnapshot{
				Data: &pganalyze_collector.CompactSnapshot_ActivitySnapshot{
					ActivitySnapshot: &pganalyze_collector.CompactActivitySnapshot{
						Backends: []*pganalyze_collector.Backend{
							createTestBackend(1234, "test-app-1", "active"),
							createTestBackend(5678, "test-app-2", "idle"),
						},
					},
				},
				CollectedAt: timestamppb.New(now),
			},
			expected: 2,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metrics, seenBackends := compactSnapshotMetrics(tc.snapshot, sysInfo)
			assert.Equal(t, tc.expected, len(metrics), "Unexpected number of time series")
			assert.Equal(t, tc.expected, len(seenBackends), "Unexpected number of seen backends")

			// Check if all metrics have the correct system info labels
			for _, metric := range metrics {
				assertSystemInfoLabels(t, metric.Labels, sysInfo)
			}
		})
	}
}

func TestCreateStaleMarkers(t *testing.T) {
	now := time.Now()

	prevBackends := map[BackendKey]bool{
		{Pid: 1234, ApplicationName: "test-app-1"}: true,
		{Pid: 5678, ApplicationName: "test-app-2"}: true,
	}

	currentBackends := map[BackendKey]bool{
		{Pid: 1234, ApplicationName: "test-app-1"}: true,
		{Pid: 9012, ApplicationName: "test-app-3"}: true,
	}

	staleMarkers := createCompactSnapshotActivityStaleMarkers(prevBackends, currentBackends, SystemInfo{}, now.UnixMilli())

	assert.Equal(t, 1, len(staleMarkers), "Expected one stale marker")
	assert.Equal(t, "5678", getLabelValue(staleMarkers[0].Labels, "pid"), "Unexpected PID in stale marker")
	assert.Equal(t, "test-app-2", getLabelValue(staleMarkers[0].Labels, "application_name"), "Unexpected application name in stale marker")
}

func TestMultipleSystemsHandling(t *testing.T) {
	now := time.Now()
	system1 := createTestSystemInfo("system-1")
	system2 := createTestSystemInfo("system-2")

	previousBackends = make(map[SystemInfo]map[BackendKey]bool)

	testCases := []struct {
		name            string
		snapshot1       *pganalyze_collector.CompactSnapshot
		snapshot2       *pganalyze_collector.CompactSnapshot
		expectedMetrics map[SystemInfo]int
		expectedStale   map[SystemInfo]int
	}{
		{
			name: "Add backends to both systems",
			snapshot1: &pganalyze_collector.CompactSnapshot{
				Data: &pganalyze_collector.CompactSnapshot_ActivitySnapshot{
					ActivitySnapshot: &pganalyze_collector.CompactActivitySnapshot{
						Backends: []*pganalyze_collector.Backend{
							createTestBackend(1001, "app-1", "active"),
							createTestBackend(1002, "app-2", "idle"),
						},
					},
				},
				CollectedAt: timestamppb.New(now),
			},
			snapshot2: &pganalyze_collector.CompactSnapshot{
				Data: &pganalyze_collector.CompactSnapshot_ActivitySnapshot{
					ActivitySnapshot: &pganalyze_collector.CompactActivitySnapshot{
						Backends: []*pganalyze_collector.Backend{
							createTestBackend(2001, "app-3", "active"),
							createTestBackend(2002, "app-4", "idle"),
						},
					},
				},
				CollectedAt: timestamppb.New(now),
			},
			expectedMetrics: map[SystemInfo]int{system1: 2, system2: 2},
			expectedStale:   map[SystemInfo]int{system1: 0, system2: 0},
		},
		{
			name: "Remove a backend from each system",
			snapshot1: &pganalyze_collector.CompactSnapshot{
				Data: &pganalyze_collector.CompactSnapshot_ActivitySnapshot{
					ActivitySnapshot: &pganalyze_collector.CompactActivitySnapshot{
						Backends: []*pganalyze_collector.Backend{
							createTestBackend(1001, "app-1", "active"),
						},
					},
				},
				CollectedAt: timestamppb.New(now),
			},
			snapshot2: &pganalyze_collector.CompactSnapshot{
				Data: &pganalyze_collector.CompactSnapshot_ActivitySnapshot{
					ActivitySnapshot: &pganalyze_collector.CompactActivitySnapshot{
						Backends: []*pganalyze_collector.Backend{
							createTestBackend(2001, "app-3", "active"),
						},
					},
				},
				CollectedAt: timestamppb.New(now),
			},
			expectedMetrics: map[SystemInfo]int{system1: 1, system2: 1},
			expectedStale:   map[SystemInfo]int{system1: 1, system2: 1},
		},
		{
			name: "Add to system 1, remove all from system 2",
			snapshot1: &pganalyze_collector.CompactSnapshot{
				Data: &pganalyze_collector.CompactSnapshot_ActivitySnapshot{
					ActivitySnapshot: &pganalyze_collector.CompactActivitySnapshot{
						Backends: []*pganalyze_collector.Backend{
							createTestBackend(1001, "app-1", "active"),
							createTestBackend(1003, "app-5", "active"),
						},
					},
				},
				CollectedAt: timestamppb.New(now),
			},
			snapshot2: &pganalyze_collector.CompactSnapshot{
				Data: &pganalyze_collector.CompactSnapshot_ActivitySnapshot{
					ActivitySnapshot: &pganalyze_collector.CompactActivitySnapshot{
						Backends: []*pganalyze_collector.Backend{},
					},
				},
				CollectedAt: timestamppb.New(now),
			},
			expectedMetrics: map[SystemInfo]int{system1: 2, system2: 0},
			expectedStale:   map[SystemInfo]int{system1: 0, system2: 1},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			metrics1, currentBackends1 := compactSnapshotMetrics(tc.snapshot1, system1)
			metrics2, currentBackends2 := compactSnapshotMetrics(tc.snapshot2, system2)

			assert.Equal(t, tc.expectedMetrics[system1], len(metrics1), "Unexpected number of metrics for system 1")
			assert.Equal(t, tc.expectedMetrics[system2], len(metrics2), "Unexpected number of metrics for system 2")

			staleMarkers1 := createCompactSnapshotActivityStaleMarkers(previousBackends[system1], currentBackends1, system1, now.UnixMilli())
			staleMarkers2 := createCompactSnapshotActivityStaleMarkers(previousBackends[system2], currentBackends2, system2, now.UnixMilli())

			assert.Equal(t, tc.expectedStale[system1], len(staleMarkers1), "Unexpected number of stale markers for system 1")
			assert.Equal(t, tc.expectedStale[system2], len(staleMarkers2), "Unexpected number of stale markers for system 2")

			for _, marker := range staleMarkers1 {
				assertSystemInfoLabels(t, marker.Labels, system1)
			}
			for _, marker := range staleMarkers2 {
				assertSystemInfoLabels(t, marker.Labels, system2)
			}

			previousBackends[system1] = currentBackends1
			previousBackends[system2] = currentBackends2
		})
	}
}

func assertSystemInfoLabels(t *testing.T, labels []prompb.Label, expectedSysInfo SystemInfo) {
	assert.Equal(t, expectedSysInfo.SystemID, getLabelValue(labels, "sys_id"))
	assert.Equal(t, expectedSysInfo.SystemIDFallback, getLabelValue(labels, "sys_id_fallback"))
	assert.Equal(t, expectedSysInfo.SystemScope, getLabelValue(labels, "sys_scope"))
	assert.Equal(t, expectedSysInfo.SystemScopeFallback, getLabelValue(labels, "sys_scope_fallback"))
	assert.Equal(t, expectedSysInfo.SystemType, getLabelValue(labels, "sys_type"))
	assert.Equal(t, expectedSysInfo.SystemTypeFallback, getLabelValue(labels, "sys_type_fallback"))
}

func getLabelValue(labels []prompb.Label, name string) string {
	for _, label := range labels {
		if label.Name == name {
			return label.Value
		}
	}
	return ""
}
