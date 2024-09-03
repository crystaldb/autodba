package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockRepository struct {
	mock.Mock
}

func (m *MockRepository) Execute(query string, options map[string]string) (*map[int64]map[string]float64, error) {
	args := m.Called(query, options)
	return args.Get(0).(*map[int64]map[string]float64), args.Error(1)
}

func (m *MockRepository) ExecuteRaw(query string, options map[string]string) ([]map[string]interface{}, error) {
	args := m.Called(query, options)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
}

func TestExecute(t *testing.T) {
	mockRepo := new(MockRepository)

	metricsInput := map[string]string{
		"connections": "connections_query",
		"cpu":         "cpu_query",
		"disk":        "disk_query",
	}

	options := map[string]string{}

	mockTimeSeriesConnections := map[int64]map[string]float64{
		1: {"connection_1": 1.5, "connection_2": 1.2, "connection_3": 0.5},
		2: {"connection_1": 2.5, "connection_2": 2.2, "connection_3": 2.5},
		3: {"connection_1": 3.5, "connection_2": 3.2, "connection_3": 3.5},
		4: {"connection_1": 4.5, "connection_2": 4.2, "connection_3": 4.5},
	}
	mockTimeSeriesCPU := map[int64]map[string]float64{
		1: {"cpu": 1.2},
		2: {"any_label": 3.0},
		3: {"cpu": 2.7},
		4: {"any_label": 1.9},
	}
	mockTimeSeriesDisk := map[int64]map[string]float64{
		1: {"any_label": 0.5},
		2: {"disk": 1.2},
		3: {"any_label": 1.9},
		4: {"disk": 2.3},
	}

	aggregateData := map[int64]map[string]float64{
		1: {"connection_1": 1.5, "connection_2": 1.2, "connection_3": 0.5, "cpu": 1.2, "disk": 0.5},
		2: {"connection_1": 2.5, "connection_2": 2.2, "connection_3": 2.5, "cpu": 3.0, "disk": 1.2},
		3: {"connection_1": 3.5, "connection_2": 3.2, "connection_3": 3.5, "cpu": 2.7, "disk": 1.9},
		4: {"connection_1": 4.5, "connection_2": 4.2, "connection_3": 4.5, "cpu": 1.9, "disk": 2.3},
	}

	mockRepo.On("Execute", "connections_query", options).Return(&mockTimeSeriesConnections, nil)
	mockRepo.On("Execute", "cpu_query", options).Return(&mockTimeSeriesCPU, nil)
	mockRepo.On("Execute", "disk_query", options).Return(&mockTimeSeriesDisk, nil)

	service := CreateService(mockRepo)
	result, err := service.Execute(metricsInput, options)

	assert.NoError(t, err)
	assert.Equal(t, aggregateData, result)

	mockRepo.AssertExpectations(t)
}

func TestExecuteSingleSeries(t *testing.T) {
	mockRepo := new(MockRepository)

	metricsInput := map[string]string{
		"connections": "connections_query",
	}

	options := map[string]string{}

	mockTimeSeriesConnections := map[int64]map[string]float64{
		1: {"any_label": 1.5},
		2: {"label2": 2.0},
		3: {"label3": 0.5},
		4: {"label4": 1.2},
	}

	aggregateData := map[int64]map[string]float64{
		1: {"connections": 1.5},
		2: {"connections": 2.0},
		3: {"connections": 0.5},
		4: {"connections": 1.2},
	}

	mockRepo.On("Execute", "connections_query", options).Return(&mockTimeSeriesConnections, nil)

	service := CreateService(mockRepo)
	result, err := service.Execute(metricsInput, options)

	assert.NoError(t, err)
	assert.Equal(t, aggregateData, result)

	mockRepo.AssertExpectations(t)
}

func TestExecuteRaw(t *testing.T) {
	mockRepo := new(MockRepository)
	service := CreateService(mockRepo)

	query := "query"
	options := map[string]string{}

	mockRawData := []map[string]interface{}{
		{
			"metric": map[string]interface{}{
				"__name__": "test_metric",
				"label":    "value",
			},
			"values": []map[string]interface{}{
				{"timestamp": 1234567890, "value": 1},
			},
		},
	}

	mockRepo.On("ExecuteRaw", query, mock.Anything).Return(mockRawData, nil)

	expectedRawData := mockRawData

	result, err := service.ExecuteRaw(query, options)

	assert.NoError(t, err)
	assert.ElementsMatch(t, expectedRawData, result)

	mockRepo.AssertExpectations(t)
}
