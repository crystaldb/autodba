package api

import (
	"collector-api/internal/auth"
	"collector-api/internal/config"
	"collector-api/internal/db"
	"collector-api/internal/storage"
	"collector-api/pkg/models"
	"compress/zlib"
	"encoding/base64"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"sort"
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

	systemInfo := extractSystemInfo(r)

	// Store the snapshot metadata
	snapshot := models.Snapshot{
		S3Location:  s3Location,
		CollectedAt: collectedAt,
		SystemID:    systemInfo.SystemID,
		SystemScope: systemInfo.SystemScope,
		SystemType:  systemInfo.SystemType,
	}
	err = db.StoreSnapshotMetadata(snapshot)
	if err != nil {
		http.Error(w, "Error storing snapshot", http.StatusInternalServerError)
		return
	}

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
	FullSnapshotType            = "full"
	CompactActivitySnapshotType = "compact_activity"
	CompactLogSnapshotType      = "compact_log"
	CompactSystemSnapshotType   = "compact_system"
)

func init() {
	// Initialize the previousMetrics map
	previousMetrics = make(map[SystemInfo]map[string][]prompb.TimeSeries)
}

func handleSnapshot(cfg *config.Config, s3Location string, collectedAt int64, systemInfo SystemInfo, processFunc func(*prometheusClient, string, SystemInfo, int64) ([]prompb.TimeSeries, error)) error {
	if cfg.Debug {
		log.Printf("Processing snapshot with s3_location: %s and collected_at: %d", s3Location, collectedAt)
	}

	promClient := prometheusClient{
		Client:   http.DefaultClient,
		endpoint: prometheusURL,
	}

	allMetrics, err := processFunc(&promClient, s3Location, systemInfo, collectedAt)
	if err != nil {
		return fmt.Errorf("process snapshot data: %w", err)
	}

	if len(allMetrics) == 0 {
		if cfg.Debug {
			log.Printf("No metrics to send, skipping remote write")
		}
		return nil
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

func processFullSnapshotData(promClient *prometheusClient, s3Location string, systemInfo SystemInfo, collectedAt int64) ([]prompb.TimeSeries, error) {
	pbBytes, err := readAndDecompressSnapshot(s3Location)
	if err != nil {
		return nil, fmt.Errorf("read and decompress snapshot: %w", err)
	}

	var fullSnapshot collector_proto.FullSnapshot
	if err := proto.Unmarshal(pbBytes, &fullSnapshot); err != nil {
		return nil, fmt.Errorf("unmarshal full snapshot: %w", err)
	}

	currentMetrics := fullSnapshotMetrics(&fullSnapshot, systemInfo, collectedAt)

	if previousMetrics[systemInfo] == nil {
		err := initializePreviousMetrics(promClient, systemInfo, FullSnapshotType)
		if err != nil {
			log.Printf("Error in initializing previous metrics: %v", err)
		}
	}

	staleMarkers := createStaleMarkers(previousMetrics[systemInfo][FullSnapshotType], currentMetrics, fullSnapshot.CollectedAt.AsTime().UnixMilli())

	allMetrics := append(currentMetrics, staleMarkers...)

	previousMetrics[systemInfo][FullSnapshotType] = currentMetrics

	return allMetrics, nil
}

func initializePreviousMetrics(promClient *prometheusClient, systemInfo SystemInfo, snapshotType string) error {
	previousMetrics[systemInfo] = make(map[string][]prompb.TimeSeries)

	if promClient == nil {
		// Initialize empty map if no client is provided (e.g., during tests)
		return nil
	}

	timeSeriesName := "" // all time-series
	if snapshotType == CompactActivitySnapshotType {
		timeSeriesName = "cc_pg_stat_activity"
	}

	// Query all metrics with the given system info labels
	query := fmt.Sprintf(`%s{sys_id="%s",sys_scope="%s",sys_type="%s"}`,
		timeSeriesName,
		systemInfo.SystemID,
		systemInfo.SystemScope,
		systemInfo.SystemType)

	// Get the latest values
	ts := time.Now()
	result, err := promClient.QueryWithRetry(query, ts, 10)
	if err != nil {
		return fmt.Errorf("query (%s) Prometheus for previous metrics: %w", query, err)
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

		skip := false

		// Convert metric labels to prompb.Label format
		for name, value := range sample.Metric {
			if snapshotType == FullSnapshotType && name == "__name__" && value == "cc_pg_stat_activity" {
				skip = true
				break
			}
			ts.Labels = append(ts.Labels, prompb.Label{
				Name:  string(name),
				Value: string(value),
			})
		}

		if skip {
			continue
		}

		sortLabels(ts.Labels)
		metrics = append(metrics, ts)
	}

	previousMetrics[systemInfo][snapshotType] = metrics

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

	systemInfo := extractSystemInfo(r)

	// Store the snapshot metadata
	snapshot := models.CompactSnapshot{
		S3Location:  s3Location,
		CollectedAt: collectedAt,
		SystemID:    systemInfo.SystemID,
		SystemScope: systemInfo.SystemScope,
		SystemType:  systemInfo.SystemType,
	}
	err = db.StoreCompactSnapshotMetadata(snapshot)
	if err != nil {
		http.Error(w, "Error storing compact snapshot", http.StatusInternalServerError)
		return
	}

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

func processCompactSnapshotData(promClient *prometheusClient, s3Location string, systemInfo SystemInfo, collectedAt int64) ([]prompb.TimeSeries, error) {
	pbBytes, err := readAndDecompressSnapshot(s3Location)
	if err != nil {
		return nil, fmt.Errorf("read and decompress snapshot: %w", err)
	}

	var compactSnapshot collector_proto.CompactSnapshot
	if err := proto.Unmarshal(pbBytes, &compactSnapshot); err != nil {
		return nil, fmt.Errorf("unmarshal compact snapshot: %w", err)
	}
	// store query text by fingerprint in db
	for _, backend := range compactSnapshot.GetActivitySnapshot().GetBackends() {
		baseRef := compactSnapshot.GetBaseRefs()
		if baseRef != nil {
			if backend.GetHasQueryIdx() {
				fp := string(baseRef.GetQueryReferences()[backend.GetQueryIdx()].GetFingerprint())

				fingerprint := base64.StdEncoding.EncodeToString([]byte(fp))
				query := string(baseRef.GetQueryInformations()[backend.GetQueryIdx()].GetNormalizedQuery())
				queryFull := backend.GetQueryText()
				err = storage.QueryStore.StoreQuery(fingerprint, query, queryFull)
				if err != nil {
					return nil, fmt.Errorf("store query: %w", err)
				}
			}
		}

	}

	var currentMetrics []prompb.TimeSeries
	snapshotType := "n/a"

	// Handle different types of snapshot data
	switch data := compactSnapshot.Data.(type) {
	case *collector_proto.CompactSnapshot_ActivitySnapshot:
		currentMetrics = compactSnapshotMetrics(&compactSnapshot, systemInfo, collectedAt)
		snapshotType = CompactActivitySnapshotType
	case *collector_proto.CompactSnapshot_LogSnapshot:
		snapshotType = CompactLogSnapshotType
		log.Printf("Log snapshot processing not yet implemented")
	case *collector_proto.CompactSnapshot_SystemSnapshot:
		snapshotType = CompactSystemSnapshotType
		log.Printf("System snapshot processing not yet implemented")
	case nil:
		log.Printf("Warning: Empty compact snapshot received")
	default:
		log.Printf("Unknown compact snapshot type: %T", data)
	}

	if snapshotType != CompactActivitySnapshotType {
		return currentMetrics, nil
	}

	if previousMetrics[systemInfo] == nil {
		err := initializePreviousMetrics(promClient, systemInfo, snapshotType)
		if err != nil {
			log.Printf("Error in initializing previous metrics: %v", err)
		}
	}

	staleMarkers := createStaleMarkers(previousMetrics[systemInfo][snapshotType], currentMetrics, compactSnapshot.CollectedAt.AsTime().UnixMilli())

	allMetrics := append(currentMetrics, staleMarkers...)

	previousMetrics[systemInfo][snapshotType] = currentMetrics

	return allMetrics, nil
}

func sendRemoteWrite(client prometheusClient, promPB []prompb.TimeSeries) error {
	resp, err := client.RemoteWrite(promPB)
	if err != nil {
		return fmt.Errorf("send remote write request: %w", err)
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

func ReprocessSnapshots(cfg *config.Config, reprocessFull, reprocessCompact bool) error {
	if !reprocessFull && !reprocessCompact {
		return nil
	}

	var allSnapshots []struct {
		timestamp  int64
		s3Location string
		systemInfo SystemInfo
		isCompact  bool
	}

	// Collect full snapshots if requested
	if reprocessFull {
		fullSnapshots, err := db.GetAllFullSnapshots()
		if err != nil {
			return fmt.Errorf("get full snapshots: %w", err)
		}

		for _, snapshot := range fullSnapshots {

			systemInfo := SystemInfo{
				SystemID:    snapshot.SystemID,
				SystemScope: snapshot.SystemScope,
				SystemType:  snapshot.SystemType,
			}
			snapshotSystemInfo, err := extractSystemInfoFromFullSnapshot(snapshot.S3Location)
			if err != nil {
				log.Printf("Error extracting system info from full snapshot %s: %v", snapshot.S3Location, err)
				continue
			}

			if systemInfo != snapshotSystemInfo {
				log.Printf("Warning: System info mismatch for snapshot %s. DB: %+v, Snapshot: %+v",
					snapshot.S3Location, systemInfo, snapshotSystemInfo)
			}

			allSnapshots = append(allSnapshots, struct {
				timestamp  int64
				s3Location string
				systemInfo SystemInfo
				isCompact  bool
			}{
				timestamp:  snapshot.CollectedAt,
				s3Location: snapshot.S3Location,
				systemInfo: systemInfo,
				isCompact:  false,
			})
		}
	}

	// Collect compact snapshots if requested
	if reprocessCompact {
		compactSnapshots, err := db.GetAllCompactSnapshots()
		if err != nil {
			return fmt.Errorf("get compact snapshots: %w", err)
		}

		for _, snapshot := range compactSnapshots {
			systemInfo := SystemInfo{
				SystemID:    snapshot.SystemID,
				SystemScope: snapshot.SystemScope,
				SystemType:  snapshot.SystemType,
			}

			allSnapshots = append(allSnapshots, struct {
				timestamp  int64
				s3Location string
				systemInfo SystemInfo
				isCompact  bool
			}{
				timestamp:  snapshot.CollectedAt,
				s3Location: snapshot.S3Location,
				systemInfo: systemInfo,
				isCompact:  true,
			})
		}
	}

	// Sort all snapshots by timestamp
	sort.Slice(allSnapshots, func(i, j int) bool {
		return allSnapshots[i].timestamp < allSnapshots[j].timestamp
	})

	// Process snapshots in chronological order
	for _, snapshot := range allSnapshots {
		if snapshot.isCompact {
			// if compact snapshot, store query text by fingerprint in db, then process compact snapshot

			log.Printf("Processing compact snapshot: %s (system_id: %s)", snapshot.s3Location, snapshot.systemInfo.SystemID)
			if err := handleCompactSnapshot(cfg, snapshot.s3Location, snapshot.timestamp, snapshot.systemInfo); err != nil {
				log.Printf("Error processing compact snapshot %s (system_id: %s): %v", snapshot.s3Location, snapshot.systemInfo.SystemID, err)
			}
		} else {
			log.Printf("Processing full snapshot: %s (system_id: %s)", snapshot.s3Location, snapshot.systemInfo.SystemID)
			if err := handleFullSnapshot(cfg, snapshot.s3Location, snapshot.timestamp, snapshot.systemInfo); err != nil {
				log.Printf("Error processing full snapshot %s (system_id: %s): %v", snapshot.s3Location, snapshot.systemInfo.SystemID, err)
			}
		}
	}

	return nil
}

var validSystemTypes = []string{
	"self_hosted",
	"amazon_rds",
	"heroku",
	"google_cloudsql",
	"azure_database",
	"crunchy_bridge",
	"aiven",
	"tembo",
}

// New helper functions to extract system info from snapshots
func extractSystemInfoFromFullSnapshot(s3Location string) (SystemInfo, error) {
	pbBytes, err := readAndDecompressSnapshot(s3Location)
	if err != nil {
		return SystemInfo{}, fmt.Errorf("read and decompress snapshot: %w", err)
	}

	var fullSnapshot collector_proto.FullSnapshot
	if err := proto.Unmarshal(pbBytes, &fullSnapshot); err != nil {
		return SystemInfo{}, fmt.Errorf("unmarshal full snapshot: %w", err)
	}

	return SystemInfo{
		SystemID:    fullSnapshot.System.GetSystemId(),
		SystemScope: fullSnapshot.System.GetSystemScope(),
		SystemType:  validSystemTypes[int(fullSnapshot.System.GetSystemInformation().GetType())],
	}, nil
}
