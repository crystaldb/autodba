package api

import (
	"collector-api/internal/auth"
	"collector-api/internal/config"
	"collector-api/internal/db"
	"collector-api/pkg/models"
	"compress/zlib"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"strconv"

	"google.golang.org/protobuf/proto"

	collector_proto "github.com/pganalyze/collector/output/pganalyze_collector"
	"github.com/prometheus/prometheus/prompb"
)

func SnapshotHandler(w http.ResponseWriter, r *http.Request) {
	// Load configuration
	cfg, err := config.LoadConfigWithDefaultPath()
	if err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		if cfg.Debug {
			log.Printf("Error loading config: %v", err)
		}
		return
	}

	// Authenticate the request
	if !auth.Authenticate(r) {
		if cfg.Debug {
			log.Printf("Unauthorized access attempt from %s", r.RemoteAddr)
		}
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if cfg.Debug {
		log.Printf("Authenticated request for snapshot submission from %s", r.RemoteAddr)
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		if cfg.Debug {
			log.Printf("Error parsing form data: %v", err)
		}
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Retrieve form values
	s3Location := r.FormValue("s3_location")
	collectedAtStr := r.FormValue("collected_at")

	// Validate the required fields
	if s3Location == "" || collectedAtStr == "" {
		if cfg.Debug {
			log.Printf("Missing required form values: s3_location or collected_at")
		}
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Parse collectedAt as an int64
	collectedAt, err := strconv.ParseInt(collectedAtStr, 10, 64)
	if err != nil {
		if cfg.Debug {
			log.Printf("Invalid collected_at value: %s", collectedAtStr)
		}
		http.Error(w, "Invalid collected_at value", http.StatusBadRequest)
		return
	}

	if cfg.Debug {
		log.Printf("Received snapshot submission: s3_location=%s, collected_at=%d", s3Location, collectedAt)
	}

	// Store the snapshot metadata
	snapshot := models.Snapshot{
		S3Location:  s3Location,
		CollectedAt: collectedAt,
	}
	err = db.StoreSnapshotMetadata(snapshot)
	if err != nil {
		http.Error(w, "Error storing snapshot", http.StatusInternalServerError)
		return
	}

	// Simulate handling the snapshot submission (replace with actual logic)
	err = handleFullSnapshot(cfg, s3Location, collectedAt)
	if err != nil {
		if cfg.Debug {
			log.Printf("Error handling snapshot submission: %v", err)
		}
		http.Error(w, "Failed to handle snapshot submission", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Snapshot successfully processed")
}

// Dummy function for handling the snapshot submission
func handleFullSnapshot(cfg *config.Config, s3Location string, collectedAt int64) error {
	// Here, implement your actual logic for processing the snapshot
	// e.g., saving to local storage, uploading to S3, etc.
	if cfg.Debug {
		log.Printf("Processing full snapshot submission with s3_location: %s and collected_at: %d", s3Location, collectedAt)
	}
	return nil
}

func CompactSnapshotHandler(w http.ResponseWriter, r *http.Request) {
	// Load configuration
	cfg, err := config.LoadConfigWithDefaultPath()
	if err != nil {
		http.Error(w, "Failed to load config", http.StatusInternalServerError)
		if cfg.Debug {
			log.Printf("Error loading config: %v", err)
		}
		return
	}

	// Authenticate the request
	if !auth.Authenticate(r) {
		if cfg.Debug {
			log.Printf("Unauthorized access attempt from %s", r.RemoteAddr)
		}
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}

	if cfg.Debug {
		log.Printf("Authenticated request for compact snapshot from %s", r.RemoteAddr)
	}

	// Parse form data
	if err := r.ParseForm(); err != nil {
		if cfg.Debug {
			log.Printf("Error parsing form data: %v", err)
		}
		http.Error(w, "Invalid request payload", http.StatusBadRequest)
		return
	}

	// Retrieve form values
	s3Location := r.FormValue("s3_location")
	collectedAtStr := r.FormValue("collected_at")

	if s3Location == "" || collectedAtStr == "" {
		if cfg.Debug {
			log.Printf("Missing required form values: s3_location or collected_at")
		}
		http.Error(w, "Missing required parameters", http.StatusBadRequest)
		return
	}

	// Parse collectedAt as an int64
	collectedAt, err := strconv.ParseInt(collectedAtStr, 10, 64)
	if err != nil {
		if cfg.Debug {
			log.Printf("Invalid collected_at value: %s", collectedAtStr)
		}
		http.Error(w, "Invalid collected_at value", http.StatusBadRequest)
		return
	}

	if cfg.Debug {
		log.Printf("Received compact snapshot: s3_location=%s, collected_at=%d", s3Location, collectedAt)
	}

	// Store the snapshot metadata
	snapshot := models.CompactSnapshot{
		S3Location:  s3Location,
		CollectedAt: collectedAt,
	}
	err = db.StoreCompactSnapshotMetadata(snapshot)
	if err != nil {
		http.Error(w, "Error storing compact snapshot", http.StatusInternalServerError)
		return
	}

	// Simulate handling the compact snapshot (you'll replace this with your actual logic)
	err = handleCompactSnapshot(cfg, s3Location, collectedAt)
	if err != nil {
		if cfg.Debug {
			log.Printf("Error handling compact snapshot: %v", err)
		}
		http.Error(w, "Failed to handle compact snapshot", http.StatusInternalServerError)
		return
	}

	// Respond with success
	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Compact snapshot successfully processed")
}

func handleCompactSnapshot(cfg *config.Config, s3Location string, collectedAt int64) error {
	if cfg.Debug {
		log.Printf("Processing compact snapshot with s3_location: %s and collected_at: %d", s3Location, collectedAt)
	}

	f, err := os.Open(path.Join("/usr/local/autodba/share/collector_api_server/storage", s3Location))
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	decompressor, err := zlib.NewReader(f)
	if err != nil {
		return fmt.Errorf("create zlib reader: %w", err)
	}
	defer decompressor.Close()

	pbBytes, err := io.ReadAll(decompressor)
	if err != nil {
		return fmt.Errorf("read compact snapshot proto: %w", err)
	}

	var compactSnapshot collector_proto.CompactSnapshot
	if err := proto.Unmarshal(pbBytes, &compactSnapshot); err != nil {
		return fmt.Errorf("unmarshal compact snapshot: %w", err)
	}

	promClient := prometheusClient{
		Client:   http.DefaultClient,
		endpoint: prometheusURL,
	}

	promPB := compactSnapshotMetrics(&compactSnapshot)

	// Send original metrics
	if err := sendRemoteWrite(promClient, promPB); err != nil {
		return fmt.Errorf("send remote write: %w", err)
	}

	// // Send Stale Marker after a delay in a goroutine
	// go func() {
	// 	time.Sleep(10 * time.Second)
	// 	stalePromPB := createStaleMarker(promPB)
	// 	if err := sendRemoteWrite(promClient, stalePromPB); err != nil {
	// 		log.Printf("Error sending stale marker: %v", err)
	// 	} else if cfg.Debug {
	// 		log.Println("Stale Marker sent successfully")
	// 	}
	// }()

	if cfg.Debug {
		log.Println("Compact snapshot processed successfully!")
	}
	return nil
}

func sendRemoteWrite(client prometheusClient, promPB []prompb.TimeSeries) error {
	resp, err := client.RemoteWrite(promPB)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode > 299 {
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			return fmt.Errorf("decode response body: %w", err)
		}
		return fmt.Errorf("unexpected status code %d: %s", resp.StatusCode, body)
	}
	return nil
}

// func createStaleMarker(promPB []prompb.TimeSeries) []prompb.TimeSeries {
// 	stalePromPB := make([]prompb.TimeSeries, len(promPB))
// 	for i, ts := range promPB {
// 		stalePromPB[i] = prompb.TimeSeries{
// 			Labels: ts.Labels,
// 			Samples: []prompb.Sample{
// 				{
// 					Timestamp: ts.Samples[0].Timestamp + (10 * time.Second).Milliseconds() - 1, // 9.999 seconds after the original
// 					Value:     math.Float64frombits(value.StaleNaN),
// 				},
// 			},
// 		}
// 	}
// 	return stalePromPB
// }
