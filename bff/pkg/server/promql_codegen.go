package server

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"
)

// PromQLInput represents the inputs required to generate a PromQL query.
type PromQLInput struct {
	DatabaseList      string        `json:"database_list"`
	Start             time.Time     `json:"start"`
	End               time.Time     `json:"end"`
	Step              time.Duration `json:"step"`
	Legend            string        `json:"legend"`
	Dim               string        `json:"dim"`
	FilterDim         string        `json:"filterdim"`
	FilterDimSelected string        `json:"filterdimselected"`
	Limit             int           `json:"limit"`
	LimitLegend       int           `json:"limit_legend"`
	Offset            int           `json:"offset"`
	DbIdentifier      string        `json:"dbidentifier"`
}

// Utility function to validate dimensions
func isValidDimension(dim string) bool {
	validDimensions := map[string]bool{
		"time":             true,
		"datname":          true, // database
		"client_addr":      true, // host
		"application_name": true, // application
		"backend_type":     true, // session_type
		"query_fp":         true, // sql
		"usename":          true, // user
		"wait_event_name":  true, // wait_event
	}
	return validDimensions[dim]
}

const query_fp_label = "query_fp"

func shouldUseRecordingRules(timeRange time.Duration) bool {
	return timeRange >= 6*time.Hour
}

func getRecordingRuleInterval(timeRange time.Duration) string {
	// For now we only have 10m recording rules
	return "10m"
}

func GenerateActivityCubePromQLQuery(input PromQLInput) (string, error) {
	timeRange := input.End.Sub(input.Start)

	if shouldUseRecordingRules(timeRange) {
		// Use recording rules version
		return generateRecordingRuleQuery(input)
	}

	// Existing query generation logic for shorter time ranges
	return GenerateStandardQuery(input)
}

func generateRecordingRuleQuery(input PromQLInput) (string, error) {
	interval := getRecordingRuleInterval(input.End.Sub(input.Start))
	metricName, err := GetRecordingRuleName([]string{input.Dim, input.Legend, input.FilterDim}, interval)
	if err != nil {
		return "", fmt.Errorf("error in getting recording rule name: %w", err)
	}

	return innerGenerateStandardQuery(input, metricName)
}

func GenerateStandardQuery(input PromQLInput) (string, error) {
	return innerGenerateStandardQuery(input, "cc_pg_stat_activity")
}

// Function to generate a PromQL query with sorting and pagination (New AST-based version)
func innerGenerateStandardQuery(input PromQLInput, baseMetric string) (string, error) {
	// Extract and validate parameters
	databaseList := input.DatabaseList
	startTime := input.Start
	endTime := input.End
	dim := input.Dim
	legend := input.Legend
	filterDim := input.FilterDim
	filterDimSelected := input.FilterDimSelected
	limitValue := input.Limit
	offsetValue := input.Offset
	dbIdentifier := input.DbIdentifier

	// Calculate time range in seconds for avg_over_time
	timeRange := endTime.Sub(startTime).Seconds()

	systemType, systemID, systemScope, err := splitDbIdentifier(dbIdentifier)
	if err != nil {
		return "", fmt.Errorf("error in splitting dbIdentifier: %w", err)
	}

	// Construct the base selector
	labels := map[string]string{
		"datname":   escapePromQLLabelValue(databaseList),
		"sys_id":    systemID,
		"sys_scope": systemScope,
		"sys_type":  systemType,
	}

	if filterDim != "" {
		filteredValues := strings.Split(filterDimSelected, ",")
		for i, v := range filteredValues {
			filteredValues[i] = escapePromQLLabelValue(v)
		}
		labels[filterDim] = strings.Join(filteredValues, "|")
	}

	selector := &Selector{
		Metric: baseMetric,
		Labels: labels,
	}

	// Apply label_replace to dim and legend if they are different and dim is not "time"
	var expr Node = selector

	// Construct the aggregation node
	var query Node
	if dim == "time" {
		query = &Aggregation{
			Func: "sum",
			By:   []string{legend},
			Expr: expr,
		}
	} else {
		avgOverTimeWindow := fmt.Sprintf("%ds", int(timeRange))
		step := fmt.Sprintf("%ds", int(input.Step.Seconds()))
		if dim == legend {
			query = genAvgOverDimQuery(dim, expr, avgOverTimeWindow, limitValue, offsetValue, step)
		} else if limitValue == 0 {
			query = &FunctionCall{
				Func: "avg_over_time",
				Args: []Node{
					&Aggregation{
						Func: "sum",
						By:   []string{dim, legend},
						Expr: expr,
					},
				},
				TimeInterval: &LiteralInt{Value: avgOverTimeWindow},
				TimeStep:     &LiteralInt{Value: step},
			}
		} else { // dim != legend and limit != ""
			// Step 1: Generate `limitOffsetAgg` with limit and offset
			limitOffsetAgg := genAvgOverDimQuery(dim, expr, avgOverTimeWindow, limitValue, offsetValue, step)

			// Step 2: Convert all values in `limitOffsetAgg` to `1`
			scaledLimitOffsetAgg := &BinaryExpr{
				Op:    ">",
				LHS:   limitOffsetAgg,
				RHS:   &LiteralInt{Value: "0"},
				On:    []string{},
				Group: "",
				Bool:  true,
			}

			// Step 3: Filter the original data by top `dim` values using `*` with `on(dim)`
			filteredData := &BinaryExpr{
				Op:    "*",
				LHS:   expr,
				RHS:   scaledLimitOffsetAgg,
				On:    []string{dim}, // This ensures that the multiplication happens on the matched `dim` label
				Group: "group_left()",
				Bool:  false,
			}

			// Step 4: Re-aggregate by `legend`
			query = &FunctionCall{
				Func: "avg_over_time",
				Args: []Node{
					&Aggregation{
						Func: "sum",
						By:   []string{dim, legend},
						Expr: filteredData,
					},
				},
				TimeInterval: &LiteralInt{Value: avgOverTimeWindow},
				TimeStep:     &LiteralInt{Value: step},
			}
		}
	}

	if dim != "time" {
		query = &SortDesc{
			Expr: query,
		}
	}

	return query.String(), nil
}

func genAvgOverDimQuery(dim string, expr Node, avgOverTimeWindow string, limitValue int, offsetValue int, step string) Node {
	avgByDim := &FunctionCall{
		Func: "avg_over_time",
		Args: []Node{
			&Aggregation{
				Func: "sum",
				By:   []string{dim},
				Expr: expr,
			},
		},
		TimeInterval: &LiteralInt{Value: avgOverTimeWindow},
		TimeStep:     &LiteralInt{Value: step},
	}

	var limitOffsetAgg Node
	limitOffsetAgg = avgByDim

	if limitValue > 0 {
		limitOffsetAgg = &Topk{
			Limit: limitValue + offsetValue,
			Expr:  limitOffsetAgg,
		}

		if offsetValue > 0 {
			// Apply `bottomk(limit)` to get the correct subset
			limitOffsetAgg = &Bottomk{
				Limit: limitValue,
				Expr:  limitOffsetAgg,
			}
		}
	}
	return limitOffsetAgg
}

// Utility function to parse label selectors
func parseLabelSelector(selector string) (map[string]string, error) {
	labelMap := make(map[string]string)
	if selector == "" {
		return labelMap, nil
	}

	labels := strings.Split(selector, ",")
	for _, label := range labels {
		label = strings.TrimSpace(label)
		if label == "" {
			continue
		}
		parts := strings.SplitN(label, "=~", 2)
		if len(parts) != 2 {
			return nil, fmt.Errorf("invalid label selector format: len: %d, label: %s, parts: %v", len(labels), label, parts)
		}
		labelKey := parts[0]
		labelValue := strings.Trim(parts[1], `"`)
		labelMap[labelKey] = labelValue
	}
	return labelMap, nil
}

// Utility function to escape PromQL label values
func escapePromQLLabelValue(value string) string {
	replacer := strings.NewReplacer(`\`, `\\`, `"`, `\"`, "\n", "")
	return replacer.Replace(value)
}

// Utility function to parse start and finish times
func parseTimeParameter(param string, now time.Time) (time.Time, error) {
	switch {
	case param == "now":
		return now, nil
	case strings.HasPrefix(param, "now-"):
		duration, err := time.ParseDuration(param[4:])
		if err != nil {
			return time.Time{}, err
		}
		return now.Add(-duration), nil
	default:
		// Assume epoch time as an integer string
		epoch, err := strconv.ParseInt(param, 10, 64)
		if err != nil {
			return time.Time{}, errors.New("invalid time format for parameter: " + param)
		}
		return time.UnixMilli(epoch), nil
	}
}
