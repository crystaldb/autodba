package api

import (
	"bytes"
	"fmt"
	"net/http"
	"net/url"

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

func compactSnapshotMetrics(snapshot *collector_proto.CompactSnapshot) []prompb.TimeSeries {
	var ts []prompb.TimeSeries
	for _, backend := range snapshot.GetActivitySnapshot().GetBackends() {
		backendTS := prompb.TimeSeries{
			Labels: []prompb.Label{
				{
					Name: "__name__", Value: "cc_pg_stat_activity",
				},
				{
					Name:  "application_name",
					Value: backend.GetApplicationName(),
				},
				{
					Name:  "backend_type",
					Value: backend.GetBackendType(),
				},
				{
					Name:  "client_port",
					Value: fmt.Sprintf("%d", backend.GetClientPort()),
				},
				{
					Name:  "client_addr",
					Value: backend.GetClientAddr(),
				},
				{
					Name:  "pid",
					Value: fmt.Sprintf("%d", backend.GetPid()),
				},
				{
					Name:  "query",
					Value: backend.GetQueryText(),
				},
				{
					Name:  "state",
					Value: backend.GetState(),
				},
				{
					Name:  "wait_event",
					Value: backend.GetWaitEvent(),
				},
				{
					Name:  "wait_event_type",
					Value: backend.GetWaitEventType(),
				},
			},
			Samples: []prompb.Sample{
				{
					Timestamp: snapshot.CollectedAt.AsTime().UnixMilli(),
					Value:     1.0,
				},
			},
		}
		ts = append(ts, backendTS)
	}
	return ts
}
