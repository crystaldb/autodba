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
	"os/exec"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
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
		log.Printf("Received full snapshot submission: s3_location=%s, collected_at=%d", s3Location, collectedAt)
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
	task := SnapshotTask{
		S3Location:  s3Location,
		CollectedAt: collectedAt,
		SystemInfo:  systemInfo,
		IsCompact:   false,
	}

	queue := GetQueueInstance()
	if queue.IsLocked() {
		// Queue the task for later processing
		queue.Enqueue(task)
		w.WriteHeader(http.StatusAccepted)
		fmt.Fprint(w, "Full snapshot queued for processing")
		log.Printf("Full snapshot queued for processing: s3_location=%s, collected_at=%d", s3Location, collectedAt)
		return
	}

	// Process immediately if not locked
	err = HandleSnapshots(cfg, []SnapshotTask{task})
	if err != nil {
		if cfg.Debug {
			log.Printf("Error handling full snapshot submission: %v", err)
		}
		http.Error(w, "Failed to handle full snapshot submission", http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusOK)
	fmt.Fprint(w, "Full snapshot successfully processed")
}

// Global variable to track the previous metrics for each system
var previousMetrics = make(map[SystemInfo]map[string][]prompb.TimeSeries)

const (
	FullSnapshotType            = "full"
	CompactActivitySnapshotType = "compact_activity"
	CompactLogSnapshotType      = "compact_log"
	CompactSystemSnapshotType   = "compact_system"
	DefaultSnapshotBatchSize    = 100
)

func init() {
	// Initialize the previousMetrics map
	previousMetrics = make(map[SystemInfo]map[string][]prompb.TimeSeries)
}

func HandleSnapshotBatches(cfg *config.Config, tasks []SnapshotTask, batchSize int) error {
	if cfg.Debug {
		log.Printf("Processing %d snapshot tasks in batches of %d", len(tasks), batchSize)
	}

	var errors []error

	for i := 0; i < len(tasks); i += batchSize {
		end := i + batchSize
		if end > len(tasks) {
			end = len(tasks)
		}

		batch := tasks[i:end]
		if cfg.Debug {
			log.Printf("Processing batch %d-%d of %d tasks", i+1, end, len(tasks))
		}

		if err := HandleSnapshots(cfg, batch); err != nil {
			errors = append(errors, fmt.Errorf("process batch %d-%d: %w", i+1, end, err))
			continue // Continue processing other batches
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("encountered %d batch processing errors: %v", len(errors), combineErrors(errors))
	}

	if cfg.Debug {
		log.Printf("Successfully processed all %d tasks in batches", len(tasks))
	}
	return nil
}

func HandleSnapshots(cfg *config.Config, tasks []SnapshotTask) error {
	if cfg.Debug {
		log.Printf("Processing %d snapshot tasks", len(tasks))
	}

	// Group tasks by SystemInfo
	tasksBySystem := make(map[SystemInfo][]SnapshotTask)
	for _, task := range tasks {
		tasksBySystem[task.SystemInfo] = append(tasksBySystem[task.SystemInfo], task)
	}

	// Process each system's tasks concurrently
	var wg sync.WaitGroup
	errorsChan := make(chan error, len(tasksBySystem))
	metricsChan := make(chan []prompb.TimeSeries, len(tasks))

	for systemInfo, systemTasks := range tasksBySystem {
		wg.Add(1)
		go func(sysInfo SystemInfo, tasks []SnapshotTask) {
			defer wg.Done()

			promClient := prometheusClient{
				Client:   http.DefaultClient,
				endpoint: prometheusURL,
			}

			var allQueries []storage.QueryRep

			// Process tasks for this system serially
			for _, task := range tasks {
				if cfg.Debug {
					log.Printf("Processing snapshot for system %s: s3_location=%s, collected_at=%d",
						sysInfo.SystemID, task.S3Location, task.CollectedAt)
				}

				var metrics []prompb.TimeSeries
				var queries []storage.QueryRep
				var err error

				if task.IsCompact {
					metrics, queries, err = processCompactSnapshotData(&promClient, task.S3Location, sysInfo, task.CollectedAt)
				} else {
					metrics, err = processFullSnapshotData(&promClient, task.S3Location, sysInfo, task.CollectedAt)
				}

				if err != nil {
					errorsChan <- fmt.Errorf("process snapshot data for system %s, location %s: %w",
						sysInfo.SystemID, task.S3Location, err)
					return
				}

				allQueries = append(allQueries, queries...)

				metricsChan <- metrics
			}

			// Batch store queries
			if len(allQueries) > 0 {
				if err := storage.QueryStore.StoreBatchQueries(allQueries); err != nil {
					errorsChan <- fmt.Errorf("store batch queries: %w", err)
				}
			}
		}(systemInfo, systemTasks)
	}

	// Close channels when all goroutines complete
	go func() {
		wg.Wait()
		close(errorsChan)
		close(metricsChan)
	}()

	// Collect results
	var allMetrics []prompb.TimeSeries
	var errors []error

	// Estimate capacity for metrics slice
	estimatedMetricsPerTask := 1000
	allMetrics = make([]prompb.TimeSeries, 0, len(tasks)*estimatedMetricsPerTask)

	for metrics := range metricsChan {
		allMetrics = append(allMetrics, metrics...)
	}

	for err := range errorsChan {
		errors = append(errors, err)
	}

	promClient := prometheusClient{
		Client:   http.DefaultClient,
		endpoint: prometheusURL,
	}

	// Send metrics if we have any
	if len(allMetrics) > 0 {
		if err := sendRemoteWrite(promClient, allMetrics); err != nil {
			errors = append(errors, fmt.Errorf("send remote write: %w", err))
		}
	} else if cfg.Debug {
		log.Printf("No metrics to send, skipping remote write")
	}

	if cfg.Debug && len(errors) == 0 {
		log.Printf("All snapshot tasks processed successfully!")
	}

	// Return combined errors if any occurred
	if len(errors) > 0 {
		return fmt.Errorf("encountered %d errors during processing: %v", len(errors), combineErrors(errors))
	}

	return nil
}

// Helper function to combine multiple errors into a single error message
func combineErrors(errors []error) error {
	if len(errors) == 0 {
		return nil
	}

	var combined strings.Builder
	for i, err := range errors {
		if i > 0 {
			combined.WriteString("; ")
		}
		combined.WriteString(err.Error())
	}
	return fmt.Errorf(combined.String())
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

	staleMarkers := createStaleMarkers(previousMetrics[systemInfo][FullSnapshotType], currentMetrics, collectedAt*1000)

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

	task := SnapshotTask{
		S3Location:  s3Location,
		CollectedAt: collectedAt,
		SystemInfo:  systemInfo,
		IsCompact:   true,
	}

	queue := GetQueueInstance()
	if queue.IsLocked() {
		// Queue the task for later processing
		queue.Enqueue(task)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, "Compact snapshot queued for processing")
		log.Printf("Compact snapshot queued for processing: s3_location=%s, collected_at=%d", s3Location, collectedAt)
		return
	}

	// Simulate handling the compact snapshot (you'll replace this with your actual logic)
	err = HandleSnapshots(cfg, []SnapshotTask{task})
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

func processCompactSnapshotData(promClient *prometheusClient, s3Location string, systemInfo SystemInfo, collectedAt int64) ([]prompb.TimeSeries, []storage.QueryRep, error) {
	pbBytes, err := readAndDecompressSnapshot(s3Location)
	if err != nil {
		return nil, nil, fmt.Errorf("read and decompress snapshot: %w", err)
	}

	var compactSnapshot collector_proto.CompactSnapshot
	if err := proto.Unmarshal(pbBytes, &compactSnapshot); err != nil {
		return nil, nil, fmt.Errorf("unmarshal compact snapshot: %w", err)
	}
	// Preallocate slice based on expected size
	backends := compactSnapshot.GetActivitySnapshot().GetBackends()
	queries := make([]storage.QueryRep, 0, len(backends))

	// Batch query processing
	baseRef := compactSnapshot.GetBaseRefs()
	if baseRef != nil {
		queryRefs := baseRef.GetQueryReferences()
		queryInfos := baseRef.GetQueryInformations()

		for _, backend := range backends {
			if !backend.GetHasQueryIdx() {
				continue
			}

			idx := backend.GetQueryIdx()
			if idx >= int32(len(queryRefs)) || idx >= int32(len(queryInfos)) {
				continue
			}

			query := string(queryInfos[idx].GetNormalizedQuery())
			if isQueryEmpty(query) {
				continue
			}

			fp := string(queryRefs[idx].GetFingerprint())
			fingerprint := base64.StdEncoding.EncodeToString([]byte(fp))

			queries = append(queries, storage.QueryRep{
				Fingerprint: fingerprint,
				Query:       query,
				QueryFull:   backend.GetQueryText(),
				CollectedAt: collectedAt,
			})
		}
	}

	currentMetrics := make([]prompb.TimeSeries, 0, len(backends))
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
		return currentMetrics, queries, nil
	}

	prevMetricsForSystem, exists := previousMetrics[systemInfo]
	if !exists {
		if err := initializePreviousMetrics(promClient, systemInfo, snapshotType); err != nil {
			log.Printf("Error in initializing previous metrics: %v", err)
		}
		prevMetricsForSystem = previousMetrics[systemInfo]
	}

	staleMarkers := createStaleMarkers(prevMetricsForSystem[snapshotType], currentMetrics, collectedAt*1000)

	// Preallocate final slice
	allMetrics := make([]prompb.TimeSeries, 0, len(currentMetrics)+len(staleMarkers))
	allMetrics = append(allMetrics, currentMetrics...)
	allMetrics = append(allMetrics, staleMarkers...)

	previousMetrics[systemInfo][snapshotType] = currentMetrics

	return allMetrics, queries, nil
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
	startTime := time.Now()

	queue := GetQueueInstance()
	defer func() {
		queue.Unlock()
		if err := queue.ProcessQueue(cfg); err != nil {
			log.Printf("Error processing queued snapshots: %v", err)
		}
	}()

	var allSnapshots []SnapshotTask

	var earliestTimestamp int64
	var latestTimestamp int64

	// Collect full snapshots if requested
	if reprocessFull {
		fullSnapshots, err := db.GetAllFullSnapshots()
		if err != nil {
			return fmt.Errorf("get full snapshots: %w", err)
		}

		for _, snapshot := range fullSnapshots {
			if earliestTimestamp == 0 || snapshot.CollectedAt < earliestTimestamp {
				earliestTimestamp = snapshot.CollectedAt
			}
			if snapshot.CollectedAt > latestTimestamp {
				latestTimestamp = snapshot.CollectedAt
			}

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

			allSnapshots = append(allSnapshots, SnapshotTask{
				S3Location:  snapshot.S3Location,
				CollectedAt: snapshot.CollectedAt,
				SystemInfo:  systemInfo,
				IsCompact:   false,
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
			if earliestTimestamp == 0 || snapshot.CollectedAt < earliestTimestamp {
				earliestTimestamp = snapshot.CollectedAt
			}
			if snapshot.CollectedAt > latestTimestamp {
				latestTimestamp = snapshot.CollectedAt
			}

			systemInfo := SystemInfo{
				SystemID:    snapshot.SystemID,
				SystemScope: snapshot.SystemScope,
				SystemType:  snapshot.SystemType,
			}

			allSnapshots = append(allSnapshots, SnapshotTask{
				S3Location:  snapshot.S3Location,
				CollectedAt: snapshot.CollectedAt,
				SystemInfo:  systemInfo,
				IsCompact:   true,
			})
		}
	}

	// Sort all snapshots by timestamp
	sort.Slice(allSnapshots, func(i, j int) bool {
		return allSnapshots[i].CollectedAt < allSnapshots[j].CollectedAt
	})

	var targetStart time.Time
	var targetEnd time.Time

	if earliestTimestamp > 0 && latestTimestamp > 0 {
		targetStart = time.Unix(0, earliestTimestamp*int64(time.Second))
		targetEnd = time.Unix(0, latestTimestamp*int64(time.Second))
		// Limit reprocessing to 2 weeks
		twoWeeksAgo := targetEnd.Add(-14 * 24 * time.Hour)
		if targetStart.Before(twoWeeksAgo) {
			if cfg.Debug {
				log.Printf("Limiting reprocessing start time from %v to %v", targetStart, twoWeeksAgo)
			}
			targetStart = twoWeeksAgo
		}
	}

	// Filter snapshots
	filteredSnapshots := make([]SnapshotTask, 0, len(allSnapshots))
	for _, snapshot := range allSnapshots {
		if !time.Unix(0, snapshot.CollectedAt*int64(time.Second)).Before(targetStart) {
			filteredSnapshots = append(filteredSnapshots, snapshot)
		} else if cfg.Debug {
			log.Printf("Filtering out snapshot %s (system_id: %s) before start time %v",
				snapshot.S3Location, snapshot.SystemInfo.SystemID, targetStart)
		}
	}

	if err := HandleSnapshotBatches(cfg, filteredSnapshots, DefaultSnapshotBatchSize); err != nil {
		log.Printf("Error handling snapshot batches: %v", err)
	}

	log.Printf("Successfully handled snapshot batches. Took %v", time.Since(startTime))

	// Evaluate recording rules for the entire time range
	if earliestTimestamp > 0 && latestTimestamp > 0 {

		if err := EvaluateRecordingRules(cfg, targetStart, targetEnd); err != nil {
			return fmt.Errorf("Failed to evaluate recording rules: %v", err)
		}

		log.Printf("Successfully evaluated recording rules. Now, restart the prometheus service with the right configuration to enable.")
	}

	log.Printf("Reprocessing snapshots completed. Took %v", time.Since(startTime))

	return nil
}

func EvaluateRecordingRules(cfg *config.Config, start, end time.Time) error {
	// Create temporary directory for rules evaluation output
	dataDir, err := os.MkdirTemp("", "prometheus-rules-*")
	if err != nil {
		return fmt.Errorf("create temp dir: %w", err)
	}
	defer os.RemoveAll(dataDir)

	// Create blocks from rules
	cmd := exec.Command("../../prometheus/promtool",
		"tsdb", "create-blocks-from", "rules",
		"--start", fmt.Sprintf("%s", start.Format(time.RFC3339)),
		"--end", fmt.Sprintf("%s", end.Format(time.RFC3339)),
		"--url", fmt.Sprintf("http://%s", prometheusURL.Host),
		"--output-dir", dataDir,
		"../../config/prometheus/recording_rules.yml")

	log.Printf("Running promtool command: %v", cmd.Args)

	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("create blocks from rules: %v\nOutput: %s", err, output)
	}

	if cfg.Debug {
		log.Printf("Successfully created blocks. Output: %s", output)
	}

	// Move the generated blocks from the data subdirectory to prometheus data directory
	entries, err := os.ReadDir(dataDir)
	if err != nil {
		return fmt.Errorf("read data directory: %w", err)
	}

	// Replace the existing move block code with:
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		source := path.Join(dataDir, entry.Name())
		dest := path.Join("../../prometheus_data", entry.Name())

		if err := copyDir(source, dest); err != nil {
			return fmt.Errorf("copy block %s: %w", entry.Name(), err)
		}

		// Clean up source directory after successful copy
		if err := os.RemoveAll(source); err != nil {
			log.Printf("Warning: Failed to clean up source directory %s: %v", source, err)
		}

		if cfg.Debug {
			log.Printf("Copied block %s to Prometheus data directory", entry.Name())
		}
	}

	return nil
}

func copyDir(src, dst string) error {
	// Create the destination directory
	if err := os.MkdirAll(dst, 0755); err != nil {
		return fmt.Errorf("create destination directory: %w", err)
	}

	// Read source directory
	entries, err := os.ReadDir(src)
	if err != nil {
		return fmt.Errorf("read source directory: %w", err)
	}

	for _, entry := range entries {
		srcPath := path.Join(src, entry.Name())
		dstPath := path.Join(dst, entry.Name())

		if entry.IsDir() {
			if err := copyDir(srcPath, dstPath); err != nil {
				return fmt.Errorf("copy directory %s: %w", entry.Name(), err)
			}
		} else {
			// Copy file
			if err := copyFile(srcPath, dstPath); err != nil {
				return fmt.Errorf("copy file %s: %w", entry.Name(), err)
			}
		}
	}

	return nil
}

func copyFile(src, dst string) error {
	sourceFile, err := os.Open(src)
	if err != nil {
		return fmt.Errorf("open source file: %w", err)
	}
	defer sourceFile.Close()

	destFile, err := os.Create(dst)
	if err != nil {
		return fmt.Errorf("create destination file: %w", err)
	}
	defer destFile.Close()

	if _, err := io.Copy(destFile, sourceFile); err != nil {
		return fmt.Errorf("copy file contents: %w", err)
	}

	// Copy file permissions
	sourceInfo, err := os.Stat(src)
	if err != nil {
		return fmt.Errorf("get source file info: %w", err)
	}

	return os.Chmod(dst, sourceInfo.Mode())
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
