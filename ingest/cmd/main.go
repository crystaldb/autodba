package main

import (
	"bytes"
	"fmt"
	"net/http"
	"time"

	"github.com/golang/protobuf/proto"
	"github.com/golang/snappy"
	remote "github.com/prometheus/prometheus/prompb"
)

func sendData(url string, series []remote.TimeSeries) error {
	req := &remote.WriteRequest{
		Timeseries: series,
	}

	data, err := proto.Marshal(req)
	if err != nil {
		return fmt.Errorf("failed to marshal data: %w", err)
	}

	compressedData := snappy.Encode(nil, data)

	httpReq, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(compressedData))
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	httpReq.Header.Set("Content-Encoding", "snappy")
	httpReq.Header.Set("Content-Type", "application/x-protobuf")
	httpReq.Header.Set("User-Agent", "your-application-name/version")
	httpReq.Header.Set("X-Prometheus-Remote-Write-Version", "0.1.0")

	resp, err := http.DefaultClient.Do(httpReq)
	if err != nil {
		return fmt.Errorf("failed to send request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("failed to write data: %s", resp.Status)
	}

	return nil
}

func main() {
	historicalData := []remote.TimeSeries{
		{
			Labels: []remote.Label{
				{Name: "__name__", Value: "crystal_all_databases"},
				{Name: "datname", Value: "newdatabase"},
				{Name: "instance", Value: "localhost:9399"},
				{Name: "job", Value: "sqlexport"},
				{Name: "target", Value: "sqlexport"},
			},
			Samples: []remote.Sample{
				{
					Value:     1.0,
					Timestamp: time.Now().UnixNano() / int64(time.Millisecond),
				},
			},
		},
	}

	if err := sendData("http://localhost:7001/api/v1/write", historicalData); err != nil {
		panic(err)
	}

	fmt.Println("Data sent successfully!")
}

