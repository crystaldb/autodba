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
	uniqueDimensions := uniqueSortedDimensions(dimensions)

	// Generate the recording rule name based on the number of dimensions
	switch len(uniqueDimensions) {
	case 0:
		return fmt.Sprintf("cc_pg_stat_activity:sum_by_%s", intervalSuffix), nil
	case 1:
		return fmt.Sprintf("cc_pg_stat_activity:sum_by_%s_%s", uniqueDimensions[0], intervalSuffix), nil
	case 2:
		dim1, dim2 := uniqueDimensions[0], uniqueDimensions[1]
		return fmt.Sprintf("cc_pg_stat_activity:sum_by_%s__%s_%s", dim1, dim2, intervalSuffix), nil
	case 3:
		dim1, dim2, dim3 := uniqueDimensions[0], uniqueDimensions[1], uniqueDimensions[2]
		return fmt.Sprintf("cc_pg_stat_activity:sum_by_%s__%s__%s_%s", dim1, dim2, dim3, intervalSuffix), nil
	default:
		return "", fmt.Errorf("one to three dimensions are required, got %d", len(uniqueDimensions))
	}
}

func uniqueSortedDimensions(dimensions []string) []string {
	dimensionSet := make(map[string]bool)
	for _, dim := range dimensions {
		if dim != "datname" && dim != "" {
			dimensionSet[dim] = true
		}
	}

	uniqueDimensions := make([]string, 0, len(dimensionSet))
	for dim := range dimensionSet {
		uniqueDimensions = append(uniqueDimensions, dim)
	}
	// Sort dimensions with "time" first
	sort.Slice(uniqueDimensions, func(i, j int) bool {
		if uniqueDimensions[i] == "time" {
			return true
		}
		if uniqueDimensions[j] == "time" {
			return false
		}
		return uniqueDimensions[i] < uniqueDimensions[j]
	})

	return uniqueDimensions
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
		intervalSuffix := strings.ReplaceAll(scenario.Step, ".", "_")

		// Generate rules for each dimension combination
		for _, dim := range ValidDimensions() {
			for _, legend := range ValidDimensions() {
				if legend == "time" {
					continue // Skip time dimension as legend
				}

				ruleName, query := generateRecordingRuleFormulae(dim, legend, "", intervalSuffix, seenQueries, scenario)
				if query == nil {
					log.Printf("Skipping rule: dim=%s, legend=%s", dim, legend)
					continue
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

				for _, filter := range ValidDimensions() {
					if filter == "time" {
						continue // Skip time dimension as filter
					}

					ruleName, query := generateRecordingRuleFormulae(dim, legend, filter, intervalSuffix, seenQueries, scenario)
					if query == nil {
						log.Printf("Skipping rule: dim=%s, legend=%s, filter=%s", dim, legend, filter)
						continue
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

func generateRecordingRuleFormulae(dim string, legend string, filter string, intervalSuffix string, seenQueries map[string]bool, scenario TimeScenario) (string, Node) {
	dimensions := []string{dim, legend}
	if filter != "" {
		dimensions = append(dimensions, filter)
	}
	ruleName, err := GetRecordingRuleName(dimensions, intervalSuffix)
	if err != nil {
		log.Printf("Error generating recording rule name: %v", err)
		return "", nil
	}

	uniqueDimensions := uniqueSortedDimensions(filterTimeDimension(dimensions))

	if seenQueries[ruleName] {
		log.Printf("Skipping duplicate rule: %s", ruleName)
		return "", nil
	}

	selector := &Selector{
		Metric: "cc_pg_stat_activity",
	}

	var expr Node = selector

	var query Node
	aggregation := &Aggregation{
		Func: "sum",
		By:   append([]string{"sys_id", "sys_scope", "sys_type", "datname"}, uniqueDimensions...),
		Expr: expr,
	}
	if dim == "time" {
		query = aggregation
	} else {
		avgOverTimeWindow := scenario.Duration
		query = &FunctionCall{
			Func: "avg_over_time",
			Args: []Node{
				aggregation,
			},
			TimeInterval: &LiteralInt{Value: avgOverTimeWindow},
			TimeStep:     &LiteralInt{Value: scenario.Step},
		}
	}
	return ruleName, query
}

func filterTimeDimension(dimensions []string) []string {
	filtered := make([]string, 0, len(dimensions))
	for _, dim := range dimensions {
		if dim != "time" {
			filtered = append(filtered, dim)
		}
	}
	return filtered
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
