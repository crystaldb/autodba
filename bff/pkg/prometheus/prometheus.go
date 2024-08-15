package prometheus

import (
	"context"
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

func (r repository) Execute(query string, options map[string]string) (*map[int64]float64, error) {

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	var range_config v1.Range

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

		range_config = v1.Range{
			Start: startTime,
			End:   endTime,
			Step:  step,
		}

	} else {
		range_config = v1.Range{
			Start: time.Now(),
			End:   time.Now(),
			Step:  (30 * time.Second),
		}
	}

	result, warnings, err := r.Api.QueryRange(ctx, query, range_config, v1.WithTimeout(5*time.Second))
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

	timeSeries := make(map[int64]float64)

	for _, result := range matrix {
		for _, sample := range result.Values {
			timeSeries[int64(sample.Timestamp)] = float64(sample.Value)
		}
	}

	return &timeSeries, nil
}
