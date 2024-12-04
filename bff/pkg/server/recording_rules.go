package server

import (
	"fmt"
	"log"
	"sort"
	"strings"
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

// GetRecordingRuleName generates a consistent recording rule name by sorting dimensions
// and applying special rules for 'time' and 'datname' dimensions
func GetRecordingRuleName(dimensions []string, intervalSuffix string) (string, error) {
	// Validate input
	if len(dimensions) != 2 {
		return "", fmt.Errorf("exactly two dimensions are required, got %d", len(dimensions))
	}

	// Handle time dimension - it should always be first if present
	if dimensions[1] == "time" {
		dimensions[0], dimensions[1] = dimensions[1], dimensions[0]
	}

	// If one of the dimensions is time, don't sort them (keep time first)
	if dimensions[0] != "time" {
		sort.Strings(dimensions)
	}

	// If datname is present, remove it from the name
	dim1, dim2 := dimensions[0], dimensions[1]
	if dim1 == "datname" {
		dim1 = dim2
		dim2 = ""
	} else if dim2 == "datname" {
		dim2 = ""
	}

	if dim1 == dim2 {
		dim2 = ""
	}

	// If we removed datname and have an empty second dimension, just use the first
	if dim2 == "" {
		return fmt.Sprintf("cc_pg_stat_activity:sum_by_%s%s",
			dim1,
			intervalSuffix), nil
	}

	// Otherwise use both dimensions
	return fmt.Sprintf("cc_pg_stat_activity:sum_by_%s__%s%s",
		dim1,
		dim2,
		intervalSuffix), nil
}

// ExtractRecordingRules generates all necessary recording rules for PostgreSQL activity metrics
// It creates two-dimensional aggregations for all valid dimension combinations
// ExtractRecordingRules generates all necessary recording rules for PostgreSQL activity metrics
// It creates two-dimensional aggregations for all valid dimension combinations
func ExtractRecordingRules() []RecordingRuleGroup {
	rulesByInterval := make(map[string][]RecordingRule)
	seenQueries := make(map[string]bool)

	// Generate rules for each time scenario
	for _, scenario := range timeScenarios {
		// Create interval suffix for rule naming (e.g., "_10m")
		intervalSuffix := "_" + strings.ReplaceAll(scenario.Step, ".", "_")

		// Generate rules for each dimension combination
		for _, dim := range ValidDimensions() {
			for _, legend := range ValidDimensions() {
				if legend == "time" {
					continue // Skip time dimension as legend
				}

				ruleName, err := GetRecordingRuleName([]string{dim, legend}, intervalSuffix)
				if err != nil {
					log.Printf("Error generating recording rule name: %v", err)
					continue // Skip this combination if there's an error
				}

				if seenQueries[ruleName] {
					log.Printf("Skipping duplicate rule: %s", ruleName)
					continue
				}

				// Create base selector
				selector := &Selector{
					Metric: "cc_pg_stat_activity",
				}

				// Apply label_replace to dim and legend if they are different and dim is not "time"
				var expr Node = selector

				// Construct the aggregation node
				var query Node
				if dim == "time" {
					aggregation := &Aggregation{
						Func: "sum",
						By:   []string{"sys_id", "sys_scope", "sys_type", "datname"},
						Expr: expr,
					}
					// Add legend to aggregation "by" clause if not already included
					if !contains(aggregation.By, legend) {
						if legend != "datname" {
							aggregation.By = append(aggregation.By, legend)
						}
					}
					query = aggregation
				} else {
					avgOverTimeWindow := scenario.Duration
					if dim == legend {
						aggregation := &Aggregation{
							Func: "sum",
							By:   []string{"sys_id", "sys_scope", "sys_type", "datname"},
							Expr: expr,
						}
						// Add legend to aggregation "by" clause if not already included
						if !contains(aggregation.By, dim) {
							if dim != "datname" {
								aggregation.By = append(aggregation.By, dim)
							}
						}
						query = &FunctionCall{
							Func: "avg_over_time",
							Args: []Node{
								aggregation,
							},
							TimeInterval: &LiteralInt{Value: avgOverTimeWindow},
							TimeStep:     &LiteralInt{Value: scenario.Step},
						}
					} else {
						aggregation := &Aggregation{
							Func: "sum",
							By:   []string{"sys_id", "sys_scope", "sys_type", "datname"},
							Expr: expr,
						}
						// Add dim to aggregation "by" clause if not already included
						if !contains(aggregation.By, dim) {
							if dim != "datname" {
								aggregation.By = append(aggregation.By, dim)
							}
						}

						// Add legend to aggregation "by" clause if not already included
						if !contains(aggregation.By, legend) {
							if legend != "datname" {
								aggregation.By = append(aggregation.By, legend)
							}
						}

						query = &FunctionCall{
							Func: "avg_over_time",
							Args: []Node{
								aggregation,
							},
							TimeInterval: &LiteralInt{Value: avgOverTimeWindow},
							TimeStep:     &LiteralInt{Value: scenario.Step},
						}
					}
				}

				// Generate query string
				queryStr := query.String()

				// if !seenQueries[queryStr] {
				rulesByInterval[scenario.Step] = append(rulesByInterval[scenario.Step], RecordingRule{
					Name: ruleName,
					Expr: queryStr,
				})
				seenQueries[ruleName] = true
				// 	seenQueries[queryStr] = true
				// }
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

// Helper function to check if a slice contains a string
func contains(slice []string, str string) bool {
	for _, v := range slice {
		if v == str {
			return true
		}
	}
	return false
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
