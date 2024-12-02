package server

import (
	"fmt"
	"regexp"
	"sort"
	"strings"
	"time"
)

// RecordingRule represents a single Prometheus recording rule with a name and PromQL expression
type RecordingRule struct {
	Name string // The name of the recording rule
	Expr string // The PromQL expression to be evaluated
}

// ValidDimensions returns all valid dimensions for activity queries
// These dimensions are used to create different aggregation combinations
func ValidDimensions() []string {
	return []string{
		"time",             // timestamp of the measurement
		"datname",          // database name
		"client_addr",      // client host address
		"application_name", // name of the application
		"backend_type",     // type of backend session
		"query",            // SQL query being executed
		"usename",          // database user name
		"wait_event_name",  // name of the wait event if any
	}
}

// TimeScenario defines a time window and recording interval
type TimeScenario struct {
	Duration string // Total time window to consider (e.g., "6h")
	Step     string // Recording interval (e.g., "10m")
}

// Available time scenarios for recording rules
// Currently only using 6h/10m, others are commented out but available if needed
var timeScenarios = []TimeScenario{
	{Duration: "6h", Step: "10m"},
	// {Duration: "12h", Step: "30m"},
	// {Duration: "24h", Step: "30m"},
	// {Duration: "48h", Step: "30m"},
	// {Duration: "168h", Step: "30m"},
	// {Duration: "336h", Step: "60m"},
}

// RecordingRuleGroup represents a group of recording rules with the same interval
type RecordingRuleGroup struct {
	Name     string          // Name of the rule group
	Interval string          // Recording interval for all rules in the group
	Rules    []RecordingRule // List of recording rules in this group
}

// ExtractRecordingRules generates all necessary recording rules for PostgreSQL activity metrics
// It creates two-dimensional aggregations for all valid dimension combinations
func ExtractRecordingRules() []RecordingRuleGroup {
	rulesByInterval := make(map[string][]RecordingRule)
	seenQueries := make(map[string]bool) // Track unique queries to avoid duplicates
	now := time.Now()

	// Generate two-dimension aggregations for each time scenario
	for _, scenario := range timeScenarios {
		fmt.Printf(">>> scenario: %+v\n", scenario)
		// Create interval suffix for rule naming (e.g., "_10m")
		intervalSuffix := "_" + strings.ReplaceAll(scenario.Step, ".", "_")

		// Base parameters for activity queries
		baseParams := ActivityParams{
			DbIdentifier: "a/b/c",
			DatabaseList: "db1", // Dummy value for recording rules
			Start:        "now-" + scenario.Duration,
			End:          "now",
			Step:         scenario.Step,
		}

		// Generate rules for each dimension combination
		for _, dim := range ValidDimensions() {
			for _, legend := range ValidDimensions() {
				if legend == "time" {
					continue // Skip time dimension as legend
				}

				params := baseParams
				params.Dim = dim
				params.Legend = legend

				// Generate and process the PromQL query
				input, err := extractPromQLInput(params, now)
				if err != nil {
					fmt.Printf(">>> error1: %v\n", err)
					continue
				}

				query, err := GenerateStandardQuery(input)
				fmt.Printf(">>> query: %s\n", query)
				if err == nil && query != "" {
					query = cleanupQueryForRecordingRule(query)
					if !seenQueries[query] {
						rulesByInterval[scenario.Step] = append(rulesByInterval[scenario.Step], RecordingRule{
							Name: fmt.Sprintf("cc_pg_stat_activity:sum_by_%s__%s%s", dim, legend, intervalSuffix),
							Expr: query,
						})
						seenQueries[query] = true
					}
				} else {
					fmt.Printf(">>> error2: %v\n", err)
				}
			}
		}
	}

	// Convert map to sorted slice of groups
	var groups []RecordingRuleGroup
	for interval, rules := range rulesByInterval {
		groups = append(groups, RecordingRuleGroup{
			Name:     fmt.Sprintf("activity_cube_recording_rules_%s", strings.ReplaceAll(interval, ".", "_")),
			Interval: interval,
			Rules:    rules,
		})
	}

	// Sort groups by interval for consistent output
	sort.Slice(groups, func(i, j int) bool {
		return groups[i].Interval < groups[j].Interval
	})

	return groups
}

// cleanupQueryForRecordingRule sanitizes a PromQL query for use in recording rules
// It removes specific filters and ensures proper dimension inclusion
func cleanupQueryForRecordingRule(query string) string {
	// Remove sort_desc wrapper if present
	if strings.HasPrefix(query, "sort_desc(") {
		query = strings.TrimPrefix(query, "sort_desc(")
		query = strings.TrimSuffix(query, ")")
	}

	// Remove specific database and system filters while preserving structure
	query = regexp.MustCompile(`\{[^}]*\}`).ReplaceAllStringFunc(query, func(match string) string {
		patterns := []string{
			`datname=~"[^"]*"(,\s*)?`,
			`sys_id=~"[^"]*"(,\s*)?`,
			`sys_scope=~"[^"]*"(,\s*)?`,
			`sys_type=~"[^"]*"(,\s*)?`,
		}

		for _, pattern := range patterns {
			match = regexp.MustCompile(pattern).ReplaceAllString(match, "")
		}
		if match == "{}" {
			return ""
		}
		return match
	})

	// Clean up formatting and ensure proper system dimensions
	query = strings.ReplaceAll(query, "  ", " ")
	if strings.Contains(query, "datname") {
		query = strings.ReplaceAll(query, "sum by(", "sum by(sys_id, sys_scope, sys_type,")
	} else {
		query = strings.ReplaceAll(query, "sum by(", "sum by(datname, sys_id, sys_scope, sys_type,")
	}

	return strings.TrimSpace(query)
}

// GeneratePrometheusConfig generates YAML configuration for Prometheus recording rules
// The output format follows Prometheus recording rules specification
func GeneratePrometheusConfig(groups []RecordingRuleGroup) string {
	var sb strings.Builder
	sb.WriteString("groups:\n")

	for _, group := range groups {
		sb.WriteString(fmt.Sprintf("  - name: %s\n", group.Name))
		sb.WriteString(fmt.Sprintf("    interval: %s\n", group.Interval))
		sb.WriteString("    rules:\n")

		for _, rule := range group.Rules {
			sb.WriteString(fmt.Sprintf("      - record: %s\n", rule.Name))
			sb.WriteString(fmt.Sprintf("        expr: %s\n", rule.Expr))
		}
		sb.WriteString("\n")
	}

	return sb.String()
}
