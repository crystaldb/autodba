package api

import (
	"bytes"
	"fmt"
	"log"
	"math"
	"net/http"
	"net/url"
	"sort"
	"strconv"

	gogoproto "github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	collector_proto "github.com/pganalyze/collector/output/pganalyze_collector"
	"github.com/prometheus/prometheus/model/value"
	"github.com/prometheus/prometheus/prompb"
)

var (
	prometheusURL = url.URL{
		Scheme: "http",
		Host:   "localhost:9090",
		Path:   "api/v1/write",
	}
)

type prometheusClient struct {
	*http.Client
	endpoint url.URL
}

func (c *prometheusClient) Do(req *http.Request) (*http.Response, error) {
	req.Header.Set("User-Agent", "autodba")
	return c.Client.Do(req)
}

func (c *prometheusClient) RemoteWrite(data []prompb.TimeSeries) (*http.Response, error) {
	payload := &prompb.WriteRequest{
		Timeseries: data,
	}
	b, err := gogoproto.Marshal(payload)
	if err != nil {
		return nil, err
	}

	compressed := snappy.Encode(nil, b)

	req, err := http.NewRequest(
		http.MethodPost,
		c.endpoint.String(),
		bytes.NewBuffer(compressed),
	)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Encoding", "snappy")
	req.Header.Set("Content-Type", "application/x-protobuf")
	req.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")

	return c.Do(req)
}

// Helper function to generate system-level labels
func systemLabels(systemInfo SystemInfo) []prompb.Label {
	return []prompb.Label{
		{Name: "sys_id", Value: systemInfo.SystemID},
		{Name: "sys_id_fallback", Value: systemInfo.SystemIDFallback},
		{Name: "sys_scope", Value: systemInfo.SystemScope},
		{Name: "sys_scope_fallback", Value: systemInfo.SystemScopeFallback},
		{Name: "sys_type", Value: systemInfo.SystemType},
		{Name: "sys_type_fallback", Value: systemInfo.SystemTypeFallback},
	}
}

// Helper function to create a time series
func createTimeSeries(systemInfo SystemInfo, metricName string, additionalLabels []prompb.Label, value float64, timestamp int64) prompb.TimeSeries {
	labels := append(systemLabels(systemInfo), additionalLabels...)
	labels = append(labels, prompb.Label{Name: "__name__", Value: metricName})

	return prompb.TimeSeries{
		Labels: labels,
		Samples: []prompb.Sample{
			{
				Timestamp: timestamp,
				Value:     value,
			},
		},
	}
}

// Helper function to create multiple time series for different metrics with shared labels
func createMultipleTimeSeries(systemInfo SystemInfo, metrics map[string]float64, sharedLabels []prompb.Label, timestamp int64) []prompb.TimeSeries {
	var ts []prompb.TimeSeries
	for metricName, value := range metrics {
		ts = append(ts, createTimeSeries(systemInfo, metricName, sharedLabels, value, timestamp))
	}
	return ts
}

// fullSnapshotMetrics processes a FullSnapshot and generates Prometheus metrics, along with individual time-series tracking
func fullSnapshotMetrics(snapshot *collector_proto.FullSnapshot, systemInfo SystemInfo) ([]prompb.TimeSeries, map[string]map[string]bool) {
	var ts []prompb.TimeSeries
	snapshotTimestamp := snapshot.CollectedAt.AsTime().UnixMilli()

	// Track seen time-series for each type of metric
	seenMetrics := map[string]map[string]bool{
		"cpu":      make(map[string]bool),
		"memory":   make(map[string]bool),
		"db":       make(map[string]bool),
		"query":    make(map[string]bool),
		"relation": make(map[string]bool),
		"index":    make(map[string]bool),
	}

	// Process system-level statistics
	ts = append(ts, processSystemStats(snapshot, systemInfo, snapshotTimestamp, seenMetrics)...)

	// Process database statistics
	ts = append(ts, processDatabaseStats(snapshot, systemInfo, snapshotTimestamp, seenMetrics)...)

	// Process query statistics
	ts = append(ts, processQueryStats(snapshot, systemInfo, snapshotTimestamp, seenMetrics)...)

	// Process relation and index statistics
	ts = append(ts, processRelationAndIndexStats(snapshot, systemInfo, snapshotTimestamp, seenMetrics)...)

	return ts, seenMetrics
}

// processSystemStats generates system-level metrics from the FullSnapshot and tracks seen metrics
func processSystemStats(snapshot *collector_proto.FullSnapshot, systemInfo SystemInfo, timestamp int64, seenMetrics map[string]map[string]bool) []prompb.TimeSeries {
	var ts []prompb.TimeSeries

	// Check if System field is nil
	if snapshot.System == nil {
		log.Println("Warning: snapshot.System is nil")
		return ts
	}

	// CPU statistics
	if snapshot.System.CpuStatistics != nil {
		for _, cpuStat := range snapshot.System.CpuStatistics {
			metricKey := fmt.Sprintf("cpu_%d", cpuStat.CpuIdx)
			seenMetrics["cpu"][metricKey] = true

			// Create multiple time-series for CPU usage (user, system, idle, etc.)
			ts = append(ts, createMultipleTimeSeries(systemInfo, map[string]float64{
				"cc_system_cpu_user_percent":   cpuStat.UserPercent,
				"cc_system_cpu_system_percent": cpuStat.SystemPercent,
				"cc_system_cpu_idle_percent":   cpuStat.IdlePercent,
				"cc_system_cpu_nice_percent":   cpuStat.NicePercent,
				"cc_system_cpu_iowait_percent": cpuStat.IowaitPercent,
				"cc_system_cpu_irq_percent":    cpuStat.IrqPercent,
			}, []prompb.Label{
				{Name: "cpu_id", Value: strconv.Itoa(int(cpuStat.CpuIdx))},
			}, timestamp)...)
		}
	} else {
		log.Println("Warning: snapshot.System.CpuStatistics is nil")
	}

	// Memory statistics
	if snapshot.System.MemoryStatistic != nil {
		seenMetrics["memory"]["system_memory"] = true

		// Create multiple time-series for all memory statistics
		ts = append(ts, createMultipleTimeSeries(systemInfo, map[string]float64{
			"cc_system_memory_total_bytes":           float64(snapshot.System.MemoryStatistic.TotalBytes),
			"cc_system_memory_free_bytes":            float64(snapshot.System.MemoryStatistic.FreeBytes),
			"cc_system_memory_cached_bytes":          float64(snapshot.System.MemoryStatistic.CachedBytes),
			"cc_system_memory_buffers_bytes":         float64(snapshot.System.MemoryStatistic.BuffersBytes),
			"cc_system_memory_dirty_bytes":           float64(snapshot.System.MemoryStatistic.DirtyBytes),
			"cc_system_memory_writeback_bytes":       float64(snapshot.System.MemoryStatistic.WritebackBytes),
			"cc_system_memory_slab_bytes":            float64(snapshot.System.MemoryStatistic.SlabBytes),
			"cc_system_memory_mapped_bytes":          float64(snapshot.System.MemoryStatistic.MappedBytes),
			"cc_system_memory_page_tables_bytes":     float64(snapshot.System.MemoryStatistic.PageTablesBytes),
			"cc_system_memory_active_bytes":          float64(snapshot.System.MemoryStatistic.ActiveBytes),
			"cc_system_memory_inactive_bytes":        float64(snapshot.System.MemoryStatistic.InactiveBytes),
			"cc_system_memory_swap_used_bytes":       float64(snapshot.System.MemoryStatistic.SwapUsedBytes),
			"cc_system_memory_swap_total_bytes":      float64(snapshot.System.MemoryStatistic.SwapTotalBytes),
			"cc_system_memory_huge_pages_size_bytes": float64(snapshot.System.MemoryStatistic.HugePagesSizeBytes),
		}, nil, timestamp)...)
	} else {
		log.Println("Warning: snapshot.System.MemoryStatistic is nil")
	}

	// Backend count statistics
	if snapshot.BackendCountStatistics != nil {
		for _, backendStat := range snapshot.BackendCountStatistics {
			// Create labels for backend type and state
			labels := []prompb.Label{
				{Name: "backend_type", Value: backendTypeToString(backendStat.BackendType)},
			}

			if backendStat.HasRoleIdx {
				if snapshot.RoleReferences != nil && len(snapshot.RoleReferences) > int(backendStat.RoleIdx) {
					roleName := snapshot.RoleReferences[backendStat.RoleIdx].Name
					labels = append(labels, prompb.Label{Name: "role", Value: roleName})
				} else {
					log.Println("Warning: snapshot.RoleReferences is nil or index out of range")
				}
			}

			if backendStat.HasDatabaseIdx {
				if snapshot.DatabaseReferences != nil && len(snapshot.DatabaseReferences) > int(backendStat.DatabaseIdx) {
					databaseName := snapshot.DatabaseReferences[backendStat.DatabaseIdx].Name
					labels = append(labels, prompb.Label{Name: "database", Value: databaseName})
				} else {
					log.Println("Warning: snapshot.DatabaseReferences is nil or index out of range")
				}
			}

			if backendStat.State != collector_proto.BackendCountStatistic_UNKNOWN_STATE {
				labels = append(labels, prompb.Label{Name: "state", Value: backendStateToString(backendStat.State)})
			}

			if backendStat.WaitingForLock {
				labels = append(labels, prompb.Label{Name: "waiting_for_lock", Value: "true"})
			}

			// Add to seen metrics
			metricKey := fmt.Sprintf("backend_type_%d_state_%d", backendStat.BackendType, backendStat.State)
			seenMetrics["backend"][metricKey] = true

			// Create a time-series for backend count
			ts = append(ts, createTimeSeries(systemInfo, "cc_backend_count", labels, float64(backendStat.Count), timestamp))
		}
	} else {
		log.Println("Warning: snapshot.BackendCountStatistics is nil")
	}

	// Disk statistics (same as before, processing all fields)
	if snapshot.System.DiskStatistics != nil {
		for _, diskStat := range snapshot.System.DiskStatistics {
			metricKey := fmt.Sprintf("disk_%d", diskStat.DiskIdx)
			seenMetrics["disk"][metricKey] = true

			// Create multiple time-series for disk I/O statistics
			ts = append(ts, createMultipleTimeSeries(systemInfo, map[string]float64{
				"cc_system_disk_read_ops_per_second":    diskStat.ReadOperationsPerSecond,
				"cc_system_disk_write_ops_per_second":   diskStat.WriteOperationsPerSecond,
				"cc_system_disk_read_bytes_per_second":  diskStat.BytesReadPerSecond,
				"cc_system_disk_write_bytes_per_second": diskStat.BytesWrittenPerSecond,
				"cc_system_disk_avg_read_latency":       diskStat.AvgReadLatency,
				"cc_system_disk_avg_write_latency":      diskStat.AvgWriteLatency,
				"cc_system_disk_utilization_percent":    diskStat.UtilizationPercent,
			}, []prompb.Label{
				{Name: "disk_idx", Value: strconv.Itoa(int(diskStat.DiskIdx))},
			}, timestamp)...)
		}
	}

	// Network statistics (same as before, processing all fields)
	if snapshot.System.NetworkStatistics != nil {
		for _, netStat := range snapshot.System.NetworkStatistics {
			metricKey := fmt.Sprintf("network_%d", netStat.NetworkIdx)
			seenMetrics["network"][metricKey] = true

			// Create multiple time-series for network I/O statistics
			ts = append(ts, createMultipleTimeSeries(systemInfo, map[string]float64{
				"cc_system_network_receive_bytes_per_second":  float64(netStat.ReceiveThroughputBytesPerSecond),
				"cc_system_network_transmit_bytes_per_second": float64(netStat.TransmitThroughputBytesPerSecond),
			}, []prompb.Label{
				{Name: "interface_name", Value: snapshot.System.NetworkReferences[netStat.NetworkIdx].InterfaceName},
			}, timestamp)...)
		}
	}

	return ts
}

// Helper function to convert BackendType enum to string
func backendTypeToString(backendType collector_proto.BackendCountStatistic_BackendType) string {
	switch backendType {
	case collector_proto.BackendCountStatistic_CLIENT_BACKEND:
		return "client_backend"
	case collector_proto.BackendCountStatistic_AUTOVACUUM_LAUNCHER:
		return "autovacuum_launcher"
	case collector_proto.BackendCountStatistic_AUTOVACUUM_WORKER:
		return "autovacuum_worker"
	case collector_proto.BackendCountStatistic_BACKGROUND_WRITER:
		return "background_writer"
	case collector_proto.BackendCountStatistic_CHECKPOINTER:
		return "checkpointer"
	case collector_proto.BackendCountStatistic_WALWRITER:
		return "walwriter"
	// Add other cases as necessary
	default:
		return "unknown"
	}
}

// Helper function to convert BackendState enum to string
func backendStateToString(state collector_proto.BackendCountStatistic_BackendState) string {
	switch state {
	case collector_proto.BackendCountStatistic_ACTIVE:
		return "active"
	case collector_proto.BackendCountStatistic_IDLE:
		return "idle"
	case collector_proto.BackendCountStatistic_IDLE_IN_TRANSACTION:
		return "idle_in_transaction"
	case collector_proto.BackendCountStatistic_IDLE_IN_TRANSACTION_ABORTED:
		return "idle_in_transaction_aborted"
	case collector_proto.BackendCountStatistic_FASTPATH_FUNCTION_CALL:
		return "fastpath_function_call"
	case collector_proto.BackendCountStatistic_DISABLED:
		return "disabled"
	// Add other states as necessary
	default:
		return "unknown_state"
	}
}

// processDatabaseStats generates metrics for each database in the FullSnapshot and tracks seen metrics
func processDatabaseStats(snapshot *collector_proto.FullSnapshot, systemInfo SystemInfo, timestamp int64, seenMetrics map[string]map[string]bool) []prompb.TimeSeries {
	var ts []prompb.TimeSeries

	for _, dbStat := range snapshot.DatabaseStatictics {
		dbName := snapshot.DatabaseReferences[dbStat.DatabaseIdx].Name
		seenMetrics["db"][dbName] = true

		// Create multiple time-series for transaction commit, rollback, and frozen XID age
		ts = append(ts, createMultipleTimeSeries(systemInfo, map[string]float64{
			"cc_db_xact_commit":    float64(dbStat.XactCommit),
			"cc_db_xact_rollback":  float64(dbStat.XactRollback),
			"cc_db_frozen_xid_age": float64(dbStat.FrozenxidAge),
		}, []prompb.Label{
			{Name: "database", Value: dbName},
		}, timestamp)...)
	}

	return ts
}

// processQueryStats generates metrics for each query in the FullSnapshot and tracks seen metrics
func processQueryStats(snapshot *collector_proto.FullSnapshot, systemInfo SystemInfo, timestamp int64, seenMetrics map[string]map[string]bool) []prompb.TimeSeries {
	var ts []prompb.TimeSeries

	for _, queryStat := range snapshot.QueryStatistics {
		query := snapshot.QueryInformations[queryStat.QueryIdx].GetNormalizedQuery()
		seenMetrics["query"][query] = true

		// Create multiple time-series for query calls and total time
		ts = append(ts, createMultipleTimeSeries(systemInfo, map[string]float64{
			"cc_query_calls":              float64(queryStat.Calls),
			"cc_query_total_time_seconds": queryStat.TotalTime,
		}, []prompb.Label{
			{Name: "query", Value: query},
		}, timestamp)...)
	}

	return ts
}

// processRelationAndIndexStats generates relation and index statistics and tracks seen metrics
func processRelationAndIndexStats(snapshot *collector_proto.FullSnapshot, systemInfo SystemInfo, timestamp int64, seenMetrics map[string]map[string]bool) []prompb.TimeSeries {
	var ts []prompb.TimeSeries

	for _, relStat := range snapshot.RelationStatistics {
		relationName := snapshot.RelationReferences[relStat.RelationIdx].RelationName
		seenMetrics["relation"][relationName] = true

		// Create multiple time-series for relation size and sequential scan count
		ts = append(ts, createMultipleTimeSeries(systemInfo, map[string]float64{
			"cc_relation_size_bytes": float64(relStat.SizeBytes),
			"cc_relation_seq_scan":   float64(relStat.SeqScan),
		}, []prompb.Label{
			{Name: "relation", Value: relationName},
		}, timestamp)...)
	}

	for _, idxStat := range snapshot.IndexStatistics {
		indexName := snapshot.IndexReferences[idxStat.IndexIdx].IndexName
		seenMetrics["index"][indexName] = true

		// Create multiple time-series for index size and scan count
		ts = append(ts, createMultipleTimeSeries(systemInfo, map[string]float64{
			"cc_index_size_bytes": float64(idxStat.SizeBytes),
			"cc_index_scan_count": float64(idxStat.IdxScan),
		}, []prompb.Label{
			{Name: "index", Value: indexName},
		}, timestamp)...)
	}

	return ts
}

// createFullSnapshotStaleMarkers generates stale markers for time-series that were present in the previous snapshot but are missing in the current one
func createFullSnapshotStaleMarkers(previousMetrics, currentMetrics map[string]map[string]bool, timestamp int64) []prompb.TimeSeries {
	var staleMarkers []prompb.TimeSeries

	for metricType, previousSeries := range previousMetrics {
		for seriesKey := range previousSeries {
			if !currentMetrics[metricType][seriesKey] {
				// Create a stale marker with NaN value for the missing time-series
				staleMarker := prompb.TimeSeries{
					Labels: []prompb.Label{
						{Name: "__name__", Value: metricType},
						{Name: "series", Value: seriesKey},
					},
					Samples: []prompb.Sample{
						{
							Timestamp: timestamp,
							Value:     math.Float64frombits(value.StaleNaN), // NaN
						},
					},
				}
				staleMarkers = append(staleMarkers, staleMarker)
			}
		}
	}

	return staleMarkers
}

// BackendKey uniquely identifies a backend session
type BackendKey struct {
	ApplicationName  string
	BackendType      string
	ClientAddr       string
	ClientPort       int32
	Datname          string
	Pid              int32
	Query            string
	QueryFingerPrint string
	QueryFull        string
	Role             string
	State            string
	WaitEvent        string
	WaitEventType    string
}

// createLabelsForBackend generates Prometheus labels for a given BackendKey
// This function is used for both active backends and stale markers
func createLabelsForBackend(backendKey BackendKey, systemInfo SystemInfo) []prompb.Label {
	labels := append(systemLabels(systemInfo), []prompb.Label{
		{Name: "__name__", Value: "cc_pg_stat_activity"},
		{Name: "application_name", Value: backendKey.ApplicationName},
		{Name: "backend_type", Value: backendKey.BackendType},
		{Name: "client_addr", Value: backendKey.ClientAddr},
		{Name: "client_port", Value: fmt.Sprintf("%d", backendKey.ClientPort)},
		{Name: "pid", Value: fmt.Sprintf("%d", backendKey.Pid)},
		{Name: "state", Value: backendKey.State},
		{Name: "wait_event", Value: backendKey.WaitEvent},
		{Name: "wait_event_type", Value: backendKey.WaitEventType},
	}...)

	if backendKey.Query != "" {
		labels = append(labels, prompb.Label{Name: "query", Value: backendKey.Query})
		labels = append(labels, prompb.Label{Name: "query_fp", Value: backendKey.QueryFingerPrint})
		labels = append(labels, prompb.Label{Name: "query_full", Value: backendKey.QueryFull})
	} else {
		labels = append(labels, prompb.Label{Name: "query", Value: backendKey.QueryFull})
	}

	if backendKey.Role != "" {
		labels = append(labels, prompb.Label{Name: "usename", Value: backendKey.Role})
	}
	if backendKey.Datname != "" {
		labels = append(labels, prompb.Label{Name: "datname", Value: backendKey.Datname})
	}

	// Sort labels by name
	sort.Slice(labels, func(i, j int) bool {
		return labels[i].Name < labels[j].Name
	})

	return labels
}

// compactSnapshotMetrics processes a compact snapshot and returns time series for each backend
// It also returns a map of seen backends for stale marker generation
func compactSnapshotMetrics(snapshot *collector_proto.CompactSnapshot, systemInfo SystemInfo) ([]prompb.TimeSeries, map[BackendKey]bool) {
	var ts []prompb.TimeSeries
	snapshotTimestamp := snapshot.CollectedAt.AsTime().UnixMilli()
	seenBackends := make(map[BackendKey]bool)
	baseRef := snapshot.GetBaseRefs()

	for _, backend := range snapshot.GetActivitySnapshot().GetBackends() {
		// Create a unique BackendKey for each backend
		backendKey := BackendKey{
			ApplicationName: backend.GetApplicationName(),
			BackendType:     backend.GetBackendType(),
			ClientAddr:      backend.GetClientAddr(),
			ClientPort:      backend.GetClientPort(),
			Pid:             backend.GetPid(),
			QueryFull:       backend.GetQueryText(),
			State:           backend.GetState(),
			WaitEvent:       backend.GetWaitEvent(),
			WaitEventType:   backend.GetWaitEventType(),
		}

		if baseRef != nil {
			if backend.GetHasRoleIdx() {
				backendKey.Role = baseRef.GetRoleReferences()[backend.GetRoleIdx()].GetName()
			}
			if backend.GetHasDatabaseIdx() {
				backendKey.Datname = baseRef.GetDatabaseReferences()[backend.GetDatabaseIdx()].GetName()
			}
			if backend.GetHasQueryIdx() {
				backendKey.Query = string(baseRef.GetQueryInformations()[backend.GetQueryIdx()].GetNormalizedQuery())
				backendKey.QueryFingerPrint = string(baseRef.GetQueryReferences()[backend.GetQueryIdx()].GetFingerprint())
			}
		}
		// Create a time series for the backend
		backendTS := prompb.TimeSeries{
			Labels: createLabelsForBackend(backendKey, systemInfo),
			Samples: []prompb.Sample{
				{
					Timestamp: snapshotTimestamp,
					Value:     1.0,
				},
			},
		}
		ts = append(ts, backendTS)
		seenBackends[backendKey] = true
	}

	return ts, seenBackends
}

// createCompactSnapshotActivityStaleMarkers generates stale markers for backends that were present in the previous snapshot
// but are missing in the current one
func createCompactSnapshotActivityStaleMarkers(previousBackends, currentBackends map[BackendKey]bool, systemInfo SystemInfo, timestamp int64) []prompb.TimeSeries {
	var staleMarkers []prompb.TimeSeries

	for backendKey := range previousBackends {
		if !currentBackends[backendKey] {
			// Create a stale marker with NaN value for backends no longer present
			staleMarker := prompb.TimeSeries{
				Labels: createLabelsForBackend(backendKey, systemInfo),
				Samples: []prompb.Sample{
					{
						Timestamp: timestamp,
						Value:     math.Float64frombits(value.StaleNaN), // NaN
					},
				},
			}
			staleMarkers = append(staleMarkers, staleMarker)
		}
	}

	return staleMarkers
}
