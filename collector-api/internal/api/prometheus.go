package api

import (
	"bytes"
	"fmt"
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

	// CPU statistics
	if snapshot.System.CpuStatistics != nil {
		for _, cpuStat := range snapshot.System.CpuStatistics {
			metricKey := fmt.Sprintf("cpu_%d", cpuStat.CpuIdx)
			seenMetrics["cpu"][metricKey] = true

			ts = append(ts, createTimeSeries(systemInfo, "cc_system_cpu_usage", []prompb.Label{
				{Name: "cpu_id", Value: strconv.Itoa(int(cpuStat.CpuIdx))},
			}, cpuStat.UserPercent, timestamp))
		}
	}

	// Memory statistics
	if snapshot.System.MemoryStatistic != nil {
		seenMetrics["memory"]["system_memory"] = true

		ts = append(ts, createTimeSeries(systemInfo, "cc_system_memory_usage", nil, float64(snapshot.System.MemoryStatistic.TotalBytes), timestamp))
	}

	return ts
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
