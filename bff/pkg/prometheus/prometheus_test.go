package prometheus

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockQueryAPI struct {
	mock.Mock
}

func (m *MockQueryAPI) QueryRange(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (model.Value, v1.Warnings, error) {
	args := m.Called(ctx, query, r)
	return args.Get(0).(model.Value), v1.Warnings{}, args.Error(2)
}

func (m *MockQueryAPI) Query(ctx context.Context, query string, ts time.Time, opts ...v1.Option) (model.Value, v1.Warnings, error) {
	args := m.Called(ctx, query, ts)
	return args.Get(0).(model.Value), v1.Warnings{}, args.Error(2)
}
func TestParseTimeRange(t *testing.T) {

	// Valid start and end
	options1 := map[string]string{"start": "1638316800000", "end": "1638317400000", "step": "1m"}
	result1, err1 := parseTimeRange(options1)
	assert.NoError(t, err1)
	assert.Equal(t, int64(1638316800000), result1.Start.UnixMilli())
	assert.Equal(t, int64(1638317400000), result1.End.UnixMilli())
	assert.Equal(t, 1*time.Minute, result1.Step)

	now := time.Now().UnixMilli()

	// Invalid Start
	options2 := map[string]string{"start": "invalid", "step": "1m"}
	_, err2 := parseTimeRange(options2)
	assert.Error(t, err2)

	// No options
	options3 := map[string]string{}
	result3, err3 := parseTimeRange(options3)
	assert.NoError(t, err3)
	assert.GreaterOrEqual(t, result3.Start.UnixMilli(), now)
	assert.GreaterOrEqual(t, result3.End.UnixMilli(), now)
	assert.Equal(t, 30*time.Second, result3.Step)
}

func TestExecute(t *testing.T) {
	mockAPI := new(MockQueryAPI)
	repo := repository{Api: mockAPI}

	query := "test_query"
	options := map[string]string{
		"start": "1638316800000",
		"end":   "1638317400000",
		"step":  "1m",
	}

	startTime := time.UnixMilli(1638316800000)
	endTime := time.UnixMilli(1638317400000)
	rangeConfig := v1.Range{
		Start: startTime,
		End:   endTime,
		Step:  1 * time.Minute,
	}

	matrix := model.Matrix{
		{
			Metric: model.Metric{"wait_event_name": "test_event"},
			Values: []model.SamplePair{
				{
					Timestamp: model.Time(startTime.UnixMilli()),
					Value:     model.SampleValue(1.0),
				},
			},
		},
	}

	expected := map[int64]map[string]float64{
		startTime.UnixMilli(): {"test_event": 1.0},
	}

	mockAPI.On("QueryRange", mock.Anything, query, rangeConfig).Return(matrix, v1.Warnings{}, nil)

	result, err := repo.Execute(query, options)
	assert.NoError(t, err)
	assert.Equal(t, expected, *result)
}

func TestExecuteRaw(t *testing.T) {
	mockAPI := new(MockQueryAPI)
	repo := repository{Api: mockAPI}

	query := "test_query"
	options := map[string]string{
		"start": "1638316800000",
		"end":   "1638317400000",
		"dim":   "time",
	}

	startTime := time.UnixMilli(1638316800000)
	endTime := time.UnixMilli(1638317400000)
	rangeConfig := v1.Range{
		Start: startTime,
		End:   endTime,
		Step:  30 * time.Second,
	}

	matrix := model.Matrix{
		{
			Metric: model.Metric{"label": "value"},
			Values: []model.SamplePair{
				{
					Timestamp: model.Time(startTime.UnixMilli()),
					Value:     model.SampleValue(1.0),
				},
			},
		},
	}

	expected := []map[string]interface{}{
		{
			"metric": map[string]interface{}{"label": "value"},
			"values": []map[string]interface{}{
				{
					"timestamp": startTime.UnixMilli(),
					"value":     1.0,
				},
			},
		},
	}

	mockAPI.On("QueryRange", mock.Anything, query, rangeConfig).Return(matrix, nil, nil)

	result, err := repo.ExecuteRaw(query, options)
	assert.NoError(t, err)
	assert.Equal(t, expected, result)
}
func TestProcessMatrix(t *testing.T) {
	tests := []struct {
		name              string
		matrix            model.Matrix
		isTimeSeriesQuery bool
		legend            string
		limitLegend       int
		expected          []map[string]interface{}
	}{
		{
			name:              "Empty Matrix",
			matrix:            model.Matrix{},
			isTimeSeriesQuery: true,
			legend:            "legend",
			limitLegend:       5,
			expected:          []map[string]interface{}{},
		},
		{
			name: "Limit Legend",
			matrix: model.Matrix{
				{
					Metric: model.Metric{"legend": "value1"},
					Values: []model.SamplePair{
						{
							Timestamp: model.Time(1),
							Value:     model.SampleValue(1.0),
						},
						{
							Timestamp: model.Time(2),
							Value:     model.SampleValue(2.0),
						},
					},
				},
				{
					Metric: model.Metric{"legend": "value2"},
					Values: []model.SamplePair{
						{
							Timestamp: model.Time(1),
							Value:     model.SampleValue(22.0),
						},
						{
							Timestamp: model.Time(2),
							Value:     model.SampleValue(11.0),
						},
					},
				},
				{
					Metric: model.Metric{"legend": "value3"},
					Values: []model.SamplePair{
						{
							Timestamp: model.Time(1),
							Value:     model.SampleValue(3.0),
						},
						{
							Timestamp: model.Time(2),
							Value:     model.SampleValue(3.0),
						},
					},
				},
				{
					Metric: model.Metric{"legend": "value4"},
					Values: []model.SamplePair{
						{
							Timestamp: model.Time(1),
							Value:     model.SampleValue(10.0),
						},
						{
							Timestamp: model.Time(2),
							Value:     model.SampleValue(13.0),
						},
					},
				},
			},
			isTimeSeriesQuery: true,
			legend:            "legend",
			limitLegend:       2,
			expected: []map[string]interface{}{
				{
					"metric": map[string]string{
						"legend": "value2",
					},
					"values": []map[string]interface{}{
						{"timestamp": int64(1), "value": 22.0},
						{"timestamp": int64(2), "value": 11.0},
					},
				},
				{
					"metric": map[string]interface{}{
						"legend": "other",
					},
					"values": []map[string]interface{}{
						{"timestamp": int64(1), "value": 14.0},
						{"timestamp": int64(2), "value": 18.0},
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := processMatrix(tt.matrix, tt.isTimeSeriesQuery, tt.legend, tt.limitLegend)
			if err != nil {
				t.Fatalf("processMatrix() error = %v", err)
			}

			assert.Equal(t, tt.expected, result)
		})
	}
}
