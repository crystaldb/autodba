package prometheus

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"local/bff/pkg/metrics"
	"strconv"
	"time"
)

type queryApi interface {
	QueryRange(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (model.Value, v1.Warnings, error)
	Query(ctx context.Context, query string, ts time.Time, opts ...v1.Option) (model.Value, v1.Warnings, error)
}

type repository struct {
	Client api.Client
	Api    queryApi
}

func New(promethues_server string) metrics.Repository {

	client, err := api.NewClient(api.Config{
		Address: promethues_server,
	})
	if err != nil {
		panic(err)
	}

	v1api := v1.NewAPI(client)
	return repository{client, v1api}
}

type Matrix []Stream

type Sample struct {
	Timestamp model.Time        `json:"timestamp"`
	Value     model.SampleValue `json:"value"`
}

type Stream struct {
	Metric model.Metric `json:"metric"`
	Values []Sample     `json:"values"`
}

type Metric map[model.LabelName]model.LabelValue

func parseTimeRange(options map[string]string) (*v1.Range, error) {
	var rangeConfig v1.Range

	if start, ok := options["start"]; ok && start != "" {

		millis, err := strconv.ParseInt(start, 10, 64)

		if err != nil {
			fmt.Println("Error parsing timestamp:", err)
			return nil, err
		}

		startTime := time.UnixMilli(millis)
		var endTime time.Time
		var step time.Duration

		if end, ok := options["end"]; ok && end != "" {
			millis, err := strconv.ParseInt(end, 10, 64)
			if err != nil {
				fmt.Println("Error parsing timestamp:", err)
				return nil, err
			}
			endTime = time.UnixMilli(millis)
		} else {
			endTime = time.Now()
		}

		if stepStr, ok := options["step"]; ok && stepStr != "" {
			step, err = time.ParseDuration(stepStr)
			if err != nil {
				fmt.Println("Error parsing step:", err)
				return nil, err
			}
		} else {
			step = (30 * time.Second)
		}

		rangeConfig = v1.Range{
			Start: startTime,
			End:   endTime,
			Step:  step,
		}

	} else {
		rangeConfig = v1.Range{
			Start: time.Now(),
			End:   time.Now(),
			Step:  (30 * time.Second),
		}
	}

	return &rangeConfig, nil
}

func (r repository) Execute(query string, options map[string]string) (*map[int64]map[string]float64, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	rangeConfig, err := parseTimeRange(options)
	if err != nil {
		fmt.Println("Error parsing time range:", err)
		return nil, err
	}

	result, warnings, err := r.Api.QueryRange(ctx, query, *rangeConfig, v1.WithTimeout(5*time.Second))
	if err != nil {
		fmt.Println("Error executing query: ", err)
		return nil, err
	}

	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}

	matrix, ok := result.(model.Matrix)
	if !ok {
		fmt.Println("Result is not a matrix")
		return nil, errors.New("Failed to parse prometheus result. Result is not a matrix")
	}

	timeSeries := make(map[int64]map[string]float64)

	for _, result := range matrix {
		for _, sample := range result.Values {
			if _, exists := timeSeries[int64(sample.Timestamp)]; !exists {
				timeSeries[int64(sample.Timestamp)] = make(map[string]float64)
			}

			var label string
			if len(result.Metric) > 0 {
				label = string(result.Metric["wait_event_name"])
			} else {
				label = "value"
			}
			timeSeries[int64(sample.Timestamp)][label] = float64(sample.Value)
		}
	}

	return &timeSeries, nil
}

func (r repository) ExecuteRaw(query string, options map[string]string) ([]map[string]interface{}, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	rangeConfig, err := parseTimeRange(options)
	if err != nil {
		fmt.Println("Error parsing time range:", err)
		return nil, err
	}

	isTimeSeriesQuery := true
	if value, ok := options["dim"]; ok && value != "time" {
		isTimeSeriesQuery = false
	}

	var result model.Value
	var warnings v1.Warnings

	if isTimeSeriesQuery {
		result, warnings, err = r.Api.QueryRange(ctx, query, *rangeConfig, v1.WithTimeout(60*time.Second))
	} else {
		result, warnings, err = r.Api.Query(ctx, query, rangeConfig.End, v1.WithTimeout(60*time.Second))
	}
	if err != nil {
		fmt.Println("Error executing query: ", err)
		return nil, err
	}

	if len(warnings) > 0 {
		fmt.Printf("Warnings: %v\n", warnings)
	}
	switch r := result.(type) {
	case model.Vector:
		fmt.Printf("Vector: %v\n", r)
		jsonVector, err := processVector(r)
		if err != nil {
			return nil, err
		}

		jsonData, err := json.MarshalIndent(jsonVector, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			return nil, err
		}

		fmt.Println("Vector result (pretty-printed JSON):")
		fmt.Println(string(jsonData))
		return jsonVector, nil
	case model.Matrix:
		fmt.Printf("Matrix: %v\n", r)
		jsonMatrix, err := processMatrix(r, isTimeSeriesQuery)
		if err != nil {
			return nil, err
		}

		jsonData, err := json.MarshalIndent(jsonMatrix, "", "  ")
		if err != nil {
			fmt.Println("Error marshaling JSON:", err)
			return nil, err
		}

		fmt.Println("Query result (pretty-printed JSON):")
		fmt.Println(string(jsonData))

		return jsonMatrix, nil
	default:
		fmt.Println("Result is of unknown type")
		return nil, errors.New("Failed to parse Prometheus result. Result is of unknown type")
	}
}
func processVector(vector model.Vector) ([]map[string]interface{}, error) {
	if len(vector) == 0 {
		return []map[string]interface{}{}, nil
	}

	var jsonVector []map[string]interface{}
	for _, sample := range vector {
		metricMap := make(map[string]interface{})
		for k, v := range sample.Metric {
			metricMap[string(k)] = string(v)
		}

		jsonSample := map[string]interface{}{
			"metric": metricMap,
			"values": []map[string]interface{}{
				{
					"timestamp": int64(sample.Timestamp),
					"value":     float64(sample.Value),
				},
			},
		}
		jsonVector = append(jsonVector, jsonSample)
	}

	return jsonVector, nil
}

func processMatrix(matrix model.Matrix, isTimeSeriesQuery bool) ([]map[string]interface{}, error) {
	if len(matrix) == 0 {
		return []map[string]interface{}{}, nil
	}

	var jsonMatrix []map[string]interface{}
	for _, stream := range matrix {
		metricMap := make(map[string]interface{})
		for k, v := range stream.Metric {
			metricMap[string(k)] = string(v)
		}

		values := make([]map[string]interface{}, len(stream.Values))
		for i, sample := range stream.Values {
			values[i] = map[string]interface{}{
				"timestamp": int64(sample.Timestamp),
				"value":     float64(sample.Value),
			}

			if !isTimeSeriesQuery {
				break
			}
		}

		jsonStream := map[string]interface{}{
			"metric": metricMap,
			"values": values,
		}
		jsonMatrix = append(jsonMatrix, jsonStream)
	}

	return jsonMatrix, nil
}

// for _, stream := range matrix {
// 			metricMap := make(map[string]interface{})
// 			for k, v := range stream.Metric {
// 				metricMap[string(k)] = string(v)
// 			}

// 			var lenght int
// 			if isTimeSeriesQuery {
// 				lenght = len(stream.Values)
// 			} else {
// 				lenght = 1
// 			}

// 			values := make([]map[string]interface{}, lenght)
// 			for i, sample := range stream.Values {
// 				values[i] = map[string]interface{}{
// 					"timestamp": int64(sample.Timestamp),
// 					"value":     float64(sample.Value),
// 				}

// 				if !isTimeSeriesQuery {
// 					break
// 				}
// 			}

// 			jsonStream := map[string]interface{}{
// 				"metric": metricMap,
// 				"values": values,
// 			}
// 			jsonMatrix = append(jsonMatrix, jsonStream)
// 		}
