package metrics

import (
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"testing"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Execute(query string, options map[string]string) (*map[int64]float64, error) {
	args := m.Called(query, options)
	return args.Get(0).(*map[int64]float64), args.Error(1)
}

func TestExecute(t *testing.T) {
	mockRepo := new(MockRepository)

	metrics := map[string]string{
		"connections": "connections_query",
		"cpu":         "cpu_query",
		"disk":        "disk_query",
	}

	options := map[string]string{}

	mockTimeSeriesConnections := map[int64]float64{1: 1.5, 2: 2.0, 3: 0.5, 4: 1.2}
	mockTimeSeriesCPU := map[int64]float64{1: 1.2, 2: 3, 3: 2.7, 4: 1.9}
	mockTimeSeriesDisk := map[int64]float64{1: 0.5, 2: 1.2, 3: 1.9, 4: 2.3}

	mockRepo.On("Execute", "connections_query", options).Return(&mockTimeSeriesConnections, nil)
	mockRepo.On("Execute", "cpu_query", options).Return(&mockTimeSeriesCPU, nil)
	mockRepo.On("Execute", "disk_query", options).Return(&mockTimeSeriesDisk, nil)

	service := CreateService(mockRepo)

	results, _ := service.Execute(metrics, options)

	// assert that the results contains 1 value for each time sampled, 1, 2, 3, 4
	assert.Len(t, results, 4)

	// Assert that for each time samples, a record is created with the correct values for each metric
	for timestamp, record := range mockTimeSeriesConnections {
		assert.Contains(t, results, timestamp, "Expected timestamp %d in results", timestamp)
		assert.Equal(t, record, results[timestamp]["connections"], "Incorrect value for connections at timestamp %d", timestamp)
	}

	for timestamp, record := range mockTimeSeriesCPU {
		assert.Contains(t, results, timestamp, "Expected timestamp %d in results", timestamp)
		assert.Equal(t, record, results[timestamp]["cpu"], "Incorrect value for cpu at timestamp %d", timestamp)
	}

	for timestamp, record := range mockTimeSeriesDisk {
		assert.Contains(t, results, timestamp, "Expected timestamp %d in results", timestamp)
		assert.Equal(t, record, results[timestamp]["disk"], "Incorrect value for disk at timestamp %d", timestamp)
	}

}
