package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"
	"sort"

	gogoproto "github.com/gogo/protobuf/proto"
	"github.com/golang/snappy"
	collector_proto "github.com/pganalyze/collector/output/pganalyze_collector"
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

// BackendKey uniquely identifies a backend session
type BackendKey struct {
	ApplicationName     string
	BackendType         string
	ClientAddr          string
	ClientPort          int32
	Datname             string
	Pid                 int32
	Query               string
	QueryFingerPrint    string
	QueryFull           string
	Role                string
	State               string
	SystemID            string
	SystemIDFallback    string
	SystemScope         string
	SystemScopeFallback string
	SystemType          string
	SystemTypeFallback  string
	WaitEvent           string
	WaitEventType       string
}

// createLabelsForBackend generates Prometheus labels for a given BackendKey
// This function is used for both active backends and stale markers
func createLabelsForBackend(backendKey BackendKey) []prompb.Label {
	labels := []prompb.Label{
		{Name: "__name__", Value: "cc_pg_stat_activity"},
		{Name: "application_name", Value: backendKey.ApplicationName},
		{Name: "backend_type", Value: backendKey.BackendType},
		{Name: "client_addr", Value: backendKey.ClientAddr},
		{Name: "client_port", Value: fmt.Sprintf("%d", backendKey.ClientPort)},
		{Name: "pid", Value: fmt.Sprintf("%d", backendKey.Pid)},
		{Name: "state", Value: backendKey.State},
		{Name: "sys_id", Value: backendKey.SystemID},
		{Name: "sys_id_fallback", Value: backendKey.SystemIDFallback},
		{Name: "sys_scope", Value: backendKey.SystemScope},
		{Name: "sys_scope_fallback", Value: backendKey.SystemScopeFallback},
		{Name: "sys_type", Value: backendKey.SystemType},
		{Name: "sys_type_fallback", Value: backendKey.SystemTypeFallback},
		{Name: "wait_event", Value: backendKey.WaitEvent},
		{Name: "wait_event_type", Value: backendKey.WaitEventType},
	}

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
			ApplicationName:     backend.GetApplicationName(),
			BackendType:         backend.GetBackendType(),
			ClientAddr:          backend.GetClientAddr(),
			ClientPort:          backend.GetClientPort(),
			Pid:                 backend.GetPid(),
			QueryFull:           backend.GetQueryText(),
			State:               backend.GetState(),
			SystemID:            systemInfo.SystemID,
			SystemIDFallback:    systemInfo.SystemIDFallback,
			SystemScope:         systemInfo.SystemScope,
			SystemScopeFallback: systemInfo.SystemScopeFallback,
			SystemType:          systemInfo.SystemType,
			SystemTypeFallback:  systemInfo.SystemTypeFallback,
			WaitEvent:           backend.GetWaitEvent(),
			WaitEventType:       backend.GetWaitEventType(),
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
			Labels: createLabelsForBackend(backendKey),
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
