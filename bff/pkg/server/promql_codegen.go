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
	DatabaseList      string `json:"database_list"`
	Start             string `json:"start"`
	End               string `json:"end"`
	Step              string `json:"step"`
	Legend            string `json:"legend"`
	Dim               string `json:"dim"`
	FilterDim         string `json:"filterdim"`
	FilterDimSelected string `json:"filterdimselected"`
	Limit             string `json:"limit"`
	LimitLegend       string `json:"limit_legend"`
	Offset            string `json:"offset"`
}

// Utility function to validate dimensions
func isValidDimension(dim string) bool {
	validDimensions := map[string]bool{
		"time":             true,
		"datname":          true, // database
		"client_addr":      true, // host
		"application_name": true, // application
		"backend_type":     true, // session_type
		"query":            true, // sql
		"usename":          true, // user
		"wait_event_name":  true, // wait_event
	}
	return validDimensions[dim]
}

// Function to generate a PromQL query with sorting and pagination (New AST-based version)
func GenerateActivityCubePromQLQuery(input PromQLInput) (string, error) {
	// Get the current time at the start of the function
	now := time.Now()

	// Extract and validate parameters
	databaseList := input.DatabaseList
	start := input.Start
	finish := input.End
	dim := input.Dim
	legend := input.Legend
	filterDim := input.FilterDim
	filterDimSelected := input.FilterDimSelected
	limit := input.Limit
	offset := input.Offset

	// Parse start and finish times
	startTime, err := parseTimeParameter(start, now)
	if err != nil {
		return "", err
	}

	var endTime time.Time
	if finish != "" {
		endTime, err = parseTimeParameter(finish, now)
		if err != nil {
			return "", err
		}
	} else {
		endTime = now
	}

	// Calculate time range in seconds for avg_over_time
	timeRange := endTime.Sub(startTime).Seconds()

	// Check required parameters
	if databaseList == "" || start == "" || dim == "" || legend == "" {
		return "", errors.New("missing required query parameters")
	}

	// Validate dimensions against whitelist
	if !isValidDimension(dim) || !isValidDimension(legend) || (filterDim != "" && !isValidDimension(filterDim)) {
		return "", errors.New("invalid dimension parameter")
	}

	// Construct the base selector
	labels := map[string]string{
		"datname": escapePromQLLabelValue(databaseList),
	}

	if dim == "query" {
		// filter out internal queries
		labels["query"] = "^.*"
	}

	if filterDim != "" {
		if filterDim == "wait_event_name" {
			// Handle wait_event_name special case
			filterExpr := handleWaitEventNameFilter(filterDimSelected)
			labelMap, err := parseLabelSelector(filterExpr)
			if err != nil {
				return "", fmt.Errorf("error parsing label selector: %w", err)
			}
			for k, v := range labelMap {
				labels[k] = v
			}
		} else {
			filteredValues := strings.Split(filterDimSelected, ",")
			for i, v := range filteredValues {
				if v == "/*pg internal*/" {
					// Special case: filter where the column value is empty
					filteredValues[i] = ""
				} else {
					filteredValues[i] = escapePromQLLabelValue(v)
				}
			}
			labels[filterDim] = strings.Join(filteredValues, "|")
		}
	}

	selector := &Selector{
		Metric: "cc_pg_stat_activity",
		Labels: labels,
	}

	// Apply label_replace to dim and legend if they are different and dim is not "time"
	var expr Node = selector
	if dim != "time" {
		expr = applyLabelReplaceAST(expr, dim)
	}
	if legend != dim {
		expr = applyLabelReplaceAST(expr, legend)
	}

	limitValue := 0
	offsetValue := 0
	if limit != "" {
		limitValue, err = strconv.Atoi(limit)
		if err != nil {
			return "", errors.New("invalid limit value")
		}
	}

	if offset != "" {
		offsetValue, err = strconv.Atoi(offset)
		if err != nil {
			return "", errors.New("invalid offset value")
		}
	}

	// Construct the aggregation node
	var query Node
	if dim == "time" {
		query = &Aggregation{
			Func: "count",
			By:   []string{legend},
			Expr: expr,
		}
	} else {
		avgOverTimeWindow := fmt.Sprintf("%ds", int(timeRange))
		if dim == legend {
			query = genAvgOverDimQuery(dim, expr, avgOverTimeWindow, limit, limitValue, offset, offsetValue)
		} else if limit == "" {
			query = &FunctionCall{
				Func: "avg_over_time",
				Args: []Node{
					&Aggregation{
						Func: "count",
						By:   []string{dim, legend},
						Expr: expr,
					},
				},
				TimeRange: &LiteralInt{Value: avgOverTimeWindow}, // Add the time range window
			}
		} else { // dim != legend and limit != ""
			// Step 1: Generate `limitOffsetAgg` with limit and offset
			limitOffsetAgg := genAvgOverDimQuery(dim, expr, avgOverTimeWindow, limit, limitValue, offset, offsetValue)

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
						Func: "count",
						By:   []string{dim, legend},
						Expr: filteredData,
					},
				},
				TimeRange: &LiteralInt{Value: avgOverTimeWindow}, // Add the time range window
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

func genAvgOverDimQuery(dim string, expr Node, avgOverTimeWindow string, limit string, limitValue int, offset string, offsetValue int) Node {
	avgByDim := &FunctionCall{
		Func: "avg_over_time",
		Args: []Node{
			&Aggregation{
				Func: "count",
				By:   []string{dim},
				Expr: expr,
			},
		},
		TimeRange: &LiteralInt{Value: avgOverTimeWindow},
	}

	var limitOffsetAgg Node
	limitOffsetAgg = avgByDim

	if limit != "" && limitValue > 0 {
		limitOffsetAgg = &Topk{
			Limit: limitValue + offsetValue,
			Expr:  limitOffsetAgg,
		}

		if offset != "" && offsetValue > 0 {
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

// Function to apply label_replace in an AST
func applyLabelReplaceAST(expr Node, label string) Node {
	if label == "wait_event_name" {
		return &LabelReplace{
			Expr: &FunctionCall{
				Func: "label_join",
				Args: []Node{
					expr,
					&LiteralString{Value: "wait_event_name"},
					&LiteralString{Value: ":"},
					&LiteralString{Value: "wait_event_type"},
					&LiteralString{Value: "wait_event"},
				},
			},
			DstLabel:    "wait_event_name",
			Replacement: "CPU",
			SrcLabel:    "wait_event_name",
			Regex:       ":",
		}
	}
	return &LabelReplace{
		Expr:        expr,
		DstLabel:    label,
		Replacement: "/*pg internal*/",
		SrcLabel:    label,
		Regex:       "",
	}
}

// Function to handle special filtering for wait_event_name
func handleWaitEventNameFilter(filterDimSelected string) string {
	filteredValues := strings.Split(filterDimSelected, ",")
	var waitEventTypes []string
	var waitEvents []string
	var conditions []string

	for _, v := range filteredValues {
		v = escapePromQLLabelValue(v)
		if v == "CPU" || v == "/*pg internal*/" || v == "" {
			// Special case for CPU: filter where wait_event is empty
			waitEventTypes = append(waitEventTypes, "")
			waitEvents = append(waitEvents, "")
		} else {
			// Handle wait_event_type:wait_event pair
			parts := strings.Split(v, ":")
			if len(parts) == 2 {
				waitEventTypes = append(waitEventTypes, parts[0])
				waitEvents = append(waitEvents, parts[1])
			}
		}
	}

	// Create condition strings if we have wait_event_types or wait_events
	if len(waitEventTypes) > 0 {
		conditions = append(conditions, fmt.Sprintf(`wait_event_type=~"%s"`, strings.Join(waitEventTypes, "|")))
	}
	if len(waitEvents) > 0 {
		conditions = append(conditions, fmt.Sprintf(`wait_event=~"%s"`, strings.Join(waitEvents, "|")))
	}

	// Combine all conditions with a logical AND (join them with commas)
	return strings.Join(conditions, ",")
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
