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
	"strings"
	"time"

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

	systemInfo := extractSystemInfo(r)

	// Simulate handling the snapshot submission (replace with actual logic)
	err = handleFullSnapshot(cfg, s3Location, collectedAt, systemInfo)
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

// Global variable to track the previous metrics for each system
var previousMetrics = make(map[SystemInfo]map[string][]prompb.TimeSeries)

const (
	FullSnapshotType    = "full"
	CompactSnapshotType = "compact"
)

func init() {
	// Initialize the previousMetrics map
	previousMetrics = make(map[SystemInfo]map[string][]prompb.TimeSeries)
}

func handleSnapshot(cfg *config.Config, s3Location string, collectedAt int64, systemInfo SystemInfo, processFunc func(*prometheusClient, string, SystemInfo) ([]prompb.TimeSeries, error)) error {
	if cfg.Debug {
		log.Printf("Processing snapshot with s3_location: %s and collected_at: %d", s3Location, collectedAt)
	}

	promClient := prometheusClient{
		Client:   http.DefaultClient,
		endpoint: prometheusURL,
	}

	allMetrics, err := processFunc(&promClient, s3Location, systemInfo)
	if err != nil {
		return fmt.Errorf("process snapshot data: %w", err)
	}

	if err := sendRemoteWrite(promClient, allMetrics); err != nil {
		return fmt.Errorf("send remote write: %w", err)
	}

	if cfg.Debug {
		log.Printf("Snapshot processed successfully!")
	}
	return nil
}

// handleFullSnapshot processes a full snapshot, generates metrics, and sends them to Prometheus
func handleFullSnapshot(cfg *config.Config, s3Location string, collectedAt int64, systemInfo SystemInfo) error {
	return handleSnapshot(cfg, s3Location, collectedAt, systemInfo, processFullSnapshotData)
}

func processFullSnapshotData(promClient *prometheusClient, s3Location string, systemInfo SystemInfo) ([]prompb.TimeSeries, error) {
	pbBytes, err := readAndDecompressSnapshot(s3Location)
	if err != nil {
		return nil, fmt.Errorf("read and decompress snapshot: %w", err)
	}

	var fullSnapshot collector_proto.FullSnapshot
	if err := proto.Unmarshal(pbBytes, &fullSnapshot); err != nil {
		return nil, fmt.Errorf("unmarshal full snapshot: %w", err)
	}

	currentMetrics := fullSnapshotMetrics(&fullSnapshot, systemInfo)

	if previousMetrics[systemInfo] == nil {
		err := initializePreviousMetrics(promClient, systemInfo)
		if err != nil {
			log.Printf("Error in initializing previous metrics: %v", err)
		}
	}

	staleMarkers := createStaleMarkers(previousMetrics[systemInfo][FullSnapshotType], currentMetrics, fullSnapshot.CollectedAt.AsTime().UnixMilli())

	allMetrics := append(currentMetrics, staleMarkers...)

	previousMetrics[systemInfo][FullSnapshotType] = currentMetrics

	return allMetrics, nil
}

func initializePreviousMetrics(promClient *prometheusClient, systemInfo SystemInfo) error {
	if promClient == nil {
		// Initialize empty map if no client is provided (e.g., during tests)
		previousMetrics[systemInfo] = make(map[string][]prompb.TimeSeries)
		return nil
	}

	// Query for both full and compact snapshot metrics
	for _, snapshotType := range []string{FullSnapshotType, CompactSnapshotType} {
		// Query all metrics with the given system info labels
		query := fmt.Sprintf(`{sys_id="%s",sys_scope="%s",sys_type="%s"}`,
			systemInfo.SystemID,
			systemInfo.SystemScope,
			systemInfo.SystemType)

		// Get the latest values within the last 15 minutes
		result, err := promClient.Query(query, time.Now().Add(-15*time.Minute))
		if err != nil {
			log.Printf("Error querying Prometheus for previous metrics: %v", err)
			continue
		}

		var metrics []prompb.TimeSeries
		for _, sample := range result {
			ts := prompb.TimeSeries{
				Labels: make([]prompb.Label, 0, len(sample.Metric)),
				Samples: []prompb.Sample{
					{
						Value:     float64(sample.Value),
						Timestamp: sample.Timestamp.UnixMilli(),
					},
				},
			}

			// Convert metric labels to prompb.Label format
			for name, value := range sample.Metric {
				ts.Labels = append(ts.Labels, prompb.Label{
					Name:  string(name),
					Value: string(value),
				})
			}

			metrics = append(metrics, ts)
		}

		if previousMetrics[systemInfo] == nil {
			previousMetrics[systemInfo] = make(map[string][]prompb.TimeSeries)
		}
		previousMetrics[systemInfo][snapshotType] = metrics

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

	systemInfo := extractSystemInfo(r)

	// Simulate handling the compact snapshot (you'll replace this with your actual logic)
	err = handleCompactSnapshot(cfg, s3Location, collectedAt, systemInfo)
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

// extractSystemInfo extracts system information from request headers
func extractSystemInfo(r *http.Request) SystemInfo {
	systemInfo := make(map[string]string)
	headers := []string{
		"Pganalyze-System-Id",
		"Pganalyze-System-Id-Fallback",
		"Pganalyze-System-Scope",
		"Pganalyze-System-Scope-Fallback",
		"Pganalyze-System-Type",
		"Pganalyze-System-Type-Fallback",
	}

	for _, header := range headers {
		value := r.Header.Get(header)
		if value != "" {
			key := strings.ToLower(strings.TrimPrefix(header, "Pganalyze-"))
			key = strings.ReplaceAll(key, "-", "_")
			systemInfo[key] = value
		}
	}

	sysInfo := SystemInfo{
		SystemID:            systemInfo["system_id"],
		SystemIDFallback:    systemInfo["system_id_fallback"],
		SystemScope:         systemInfo["system_scope"],
		SystemScopeFallback: systemInfo["system_scope_fallback"],
		SystemType:          systemInfo["system_type"],
		SystemTypeFallback:  systemInfo["system_type_fallback"],
	}

	return sysInfo
}

func readAndDecompressSnapshot(s3Location string) ([]byte, error) {
	// Open the file at the given S3 location
	f, err := os.Open(path.Join(s3Location))
	if err != nil {
		return nil, fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	// Decompress the zlib compressed snapshot
	decompressor, err := zlib.NewReader(f)
	if err != nil {
		return nil, fmt.Errorf("create zlib reader: %w", err)
	}
	defer decompressor.Close()

	// Read all bytes from the decompressed file
	pbBytes, err := io.ReadAll(decompressor)
	if err != nil {
		return nil, fmt.Errorf("read compact/full snapshot proto: %w", err)
	}

	return pbBytes, nil
}

// handleCompactSnapshot processes a compact snapshot, generates metrics and stale markers,
// and sends them to Prometheus
func handleCompactSnapshot(cfg *config.Config, s3Location string, collectedAt int64, systemInfo SystemInfo) error {
	return handleSnapshot(cfg, s3Location, collectedAt, systemInfo, processCompactSnapshotData)
}

func processCompactSnapshotData(promClient *prometheusClient, s3Location string, systemInfo SystemInfo) ([]prompb.TimeSeries, error) {
	pbBytes, err := readAndDecompressSnapshot(s3Location)
	if err != nil {
		return nil, fmt.Errorf("read and decompress snapshot: %w", err)
	}

	var compactSnapshot collector_proto.CompactSnapshot
	if err := proto.Unmarshal(pbBytes, &compactSnapshot); err != nil {
		return nil, fmt.Errorf("unmarshal compact snapshot: %w", err)
	}

	currentMetrics := compactSnapshotMetrics(&compactSnapshot, systemInfo)

	if previousMetrics[systemInfo] == nil {
		err := initializePreviousMetrics(promClient, systemInfo)
		if err != nil {
			log.Printf("Error in initializing previous metrics: %v", err)
		}
	}

	staleMarkers := createStaleMarkers(previousMetrics[systemInfo][CompactSnapshotType], currentMetrics, compactSnapshot.CollectedAt.AsTime().UnixMilli())

	allMetrics := append(currentMetrics, staleMarkers...)

	previousMetrics[systemInfo][CompactSnapshotType] = currentMetrics

	return allMetrics, nil
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
