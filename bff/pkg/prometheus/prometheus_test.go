package prometheus

import (
	"context"
	"testing"
	"time"

	"github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"strconv"
)

type MockAPI struct {
	mock.Mock
}

func (m *MockAPI) QueryRange(ctx context.Context, query string, r v1.Range, opts ...v1.Option) (model.Value, v1.Warnings, error) {
	args := m.Called(ctx, query, r, opts)
	return args.Get(0).(model.Value), args.Get(1).(v1.Warnings), args.Error(2)
}

func TestExecute(t *testing.T) {
	mockAPI := new(MockAPI)
	repo := repository{Api: mockAPI}

	// All options provided
	var start int64 = 1723438800000
	var end int64 = 1723525200000
	step := "49s"

	expectedRange := v1.Range{
		Start: time.UnixMilli(start),
		End:   time.UnixMilli(end),
		Step:  49 * time.Second,
	}

	mockAPI.On("QueryRange", mock.Anything, "some_query", expectedRange, mock.Anything).Return(model.Matrix{}, v1.Warnings{}, nil)

	options := map[string]string{"start": strconv.FormatInt(start, 10), "end": strconv.FormatInt(end, 10), "step": step}
	_, err := repo.Execute("some_query", options)
	assert.NoError(t, err)
	mockAPI.AssertCalled(t, "QueryRange", mock.Anything, "some_query", expectedRange, mock.Anything)

	// Only start provided
	expectedStart := time.UnixMilli(start)
	expectedStep := 30 * time.Second

	mockAPI.On("QueryRange", mock.Anything, "some_query", mock.MatchedBy(func(r v1.Range) bool {
		return r.Start.Equal(expectedStart) &&
			r.End.After(expectedStart) &&
			r.Step == expectedStep
	}), mock.Anything).Return(model.Matrix{}, v1.Warnings{}, nil)

	options = map[string]string{"start": strconv.FormatInt(start, 10)}
	_, err = repo.Execute("some_query", options)
	assert.NoError(t, err)
	mockAPI.AssertCalled(t, "QueryRange", mock.Anything, "some_query", expectedRange, mock.Anything)
}

func TestResponseDataHandling(t *testing.T) {
	mockAPI := new(MockAPI)
	repo := repository{Api: mockAPI}

	var start int64 = 1723438800000

	t1 := time.UnixMilli(start)
	t2 := t1.Add(30 * time.Second)
	t3 := t2.Add(30 * time.Second)

	mockMatrix := model.Matrix{
		{
			Metric: model.Metric{"name": "test_metric"},
			Values: []model.SamplePair{
				{Timestamp: model.Time(t1.Unix()), Value: 1.0},
				{Timestamp: model.Time(t2.Unix()), Value: 2.0},
				{Timestamp: model.Time(t3.Unix()), Value: 3.0},
			},
		},
	}

	expectedTimeSeries := map[int64]float64{
		t1.Unix(): 1.0,
		t2.Unix(): 2.0,
		t3.Unix(): 3.0,
	}

	mockAPI.On("QueryRange", mock.Anything, "some_query", mock.Anything, mock.Anything).Return(mockMatrix, v1.Warnings{}, nil)

	options := map[string]string{}

	timeSeries, _ := repo.Execute("some_query", options)

	assert.Equal(t, expectedTimeSeries, *timeSeries)
}
