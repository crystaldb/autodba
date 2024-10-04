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
var previousMetrics = make(map[SystemInfo]map[string]map[string]bool)

func init() {
	// Initialize the previousMetrics map
	previousMetrics = make(map[SystemInfo]map[string]map[string]bool)
}

// handleFullSnapshot processes a full snapshot, generates metrics, and sends them to Prometheus
func handleFullSnapshot(cfg *config.Config, s3Location string, collectedAt int64, systemInfo SystemInfo) error {
	if cfg.Debug {
		log.Printf("Processing full snapshot with s3_location: %s and collected_at: %d", s3Location, collectedAt)
	}

	// Open the file at the given S3 location
	f, err := os.Open(path.Join(s3Location))
	if err != nil {
		return fmt.Errorf("open file: %w", err)
	}
	defer f.Close()

	// Decompress the zlib compressed snapshot
	decompressor, err := zlib.NewReader(f)
	if err != nil {
		return fmt.Errorf("create zlib reader: %w", err)
	}
	defer decompressor.Close()

	// Read all bytes from the decompressed file
	pbBytes, err := io.ReadAll(decompressor)
	if err != nil {
		return fmt.Errorf("read full snapshot proto: %w", err)
	}

	// Unmarshal the bytes into a FullSnapshot protobuf message
	var fullSnapshot collector_proto.FullSnapshot
	if err := proto.Unmarshal(pbBytes, &fullSnapshot); err != nil {
		return fmt.Errorf("unmarshal full snapshot: %w", err)
	}

	// Initialize a Prometheus client to send the metrics
	promClient := prometheusClient{
		Client:   http.DefaultClient,
		endpoint: prometheusURL,
	}

	// Process the snapshot and extract relevant metrics and current time-series tracking
	metrics, seenMetrics := fullSnapshotMetrics(&fullSnapshot, systemInfo)

	// Generate stale markers for each type of metric that was seen in the previous snapshot but not in this one
	staleMarkers := createFullSnapshotStaleMarkers(previousMetrics[systemInfo], seenMetrics, fullSnapshot.CollectedAt.AsTime().UnixMilli())

	// Combine all metrics and stale markers
	allMetrics := append(metrics, staleMarkers...)

	// Send all the metrics to Prometheus using the Remote Write API
	if err := sendRemoteWrite(promClient, allMetrics); err != nil {
		return fmt.Errorf("send remote write: %w", err)
	}

	// Update the previousMetrics map with the current time-series for future comparisons
	previousMetrics[systemInfo] = seenMetrics

	if cfg.Debug {
		log.Println("Full snapshot processed successfully!")
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

type SystemInfo struct {
	SystemID            string
	SystemIDFallback    string
	SystemScope         string
	SystemScopeFallback string
	SystemType          string
	SystemTypeFallback  string
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

var previousBackends map[SystemInfo]map[BackendKey]bool

func init() {
	previousBackends = make(map[SystemInfo]map[BackendKey]bool)
}

// handleCompactSnapshot processes a compact snapshot, generates metrics and stale markers,
// and sends them to Prometheus
func handleCompactSnapshot(cfg *config.Config, s3Location string, collectedAt int64, systemInfo SystemInfo) error {
	if cfg.Debug {
		log.Printf("Processing compact snapshot with s3_location: %s and collected_at: %d", s3Location, collectedAt)
	}

	f, err := os.Open(path.Join(s3Location))
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

	// Process the snapshot and get metrics and current backends
	metrics, currentBackends := compactSnapshotMetrics(&compactSnapshot, systemInfo)

	// Generate stale markers for time-series (identified by a unique set of labels) no longer present
	staleMarkers := createCompactSnapshotActivityStaleMarkers(previousBackends[systemInfo], currentBackends, systemInfo, compactSnapshot.CollectedAt.AsTime().UnixMilli())

	// Combine active metrics and stale markers
	allMetrics := append(metrics, staleMarkers...)

	// Send all metrics to Prometheus
	if err := sendRemoteWrite(promClient, allMetrics); err != nil {
		return fmt.Errorf("send remote write: %w", err)
	}

	// Update previousBackends for the next snapshot comparison
	previousBackends[systemInfo] = currentBackends

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
