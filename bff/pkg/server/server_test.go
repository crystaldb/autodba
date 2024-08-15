package server

import (
	"encoding/json"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"net/http"
	"net/http/httptest"
	"testing"
)

type MockMetricsService struct {
	mock.Mock
}

func (m *MockMetricsService) Execute(metrics map[string]string, options map[string]string) (map[int64]map[string]float64, error) {
	args := m.Called(metrics, options)
	return args.Get(0).(map[int64]map[string]float64), args.Error(1)
}

func TestEndpointsGeneration(t *testing.T) {
	mockMetricsService := new(MockMetricsService)
	mockMetricsService.On("Execute", mock.Anything, mock.Anything).Return(
		map[int64]map[string]float64{
			1620000000000: {"connections": 12.0},
		},
		nil,
	)

	route := "/v1/health"
	routesConfig := map[string]RouteConfig{
		route: {
			Params: []string{"database_id", "start", "end"},
			Options: map[string]string{
				"start": "$start",
				"end":   "$end",
			},
			Metrics: map[string]string{
				"connections": "",
				"cpu":         "",
				"disk_usage":  "",
			},
		},
	}

	handler := metrics_handler(routesConfig, mockMetricsService)

	// TEST configured route exists
	record := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health", nil)
	handler.ServeHTTP(record, req)
	assert.Equal(t, http.StatusOK, record.Code)

	// TEST route not found
	record = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v2/health", nil)
	handler.ServeHTTP(record, req)
	assert.Equal(t, http.StatusNotFound, record.Code)
}

func TestParamsPopulation(t *testing.T) {
	route := "/v1/health"
	routesConfig := map[string]RouteConfig{
		route: {
			Params: []string{"database_id", "start", "end"},
			Options: map[string]string{
				"start": "$start",
				"end":   "$end",
			},
			Metrics: map[string]string{
				"connections": "sum(pg_stat_database_numbackends{datname=~\"$database_id\"})/sum(pg_settings_max_connections)",
			},
		},
	}

	expectedMetrics := map[string]string{
		"connections": "sum(pg_stat_database_numbackends{datname=~\"test_db\"})/sum(pg_settings_max_connections)",
	}

	expectedOptions := map[string]string{
		"start": "0000",
		"end":   "1111",
	}

	mockMetricsService := new(MockMetricsService)

	mockMetricsService.On("Execute", mock.MatchedBy(func(metrics map[string]string) bool {
		assert.Equal(t, expectedMetrics, metrics, "Metrics should be populated with params")
		return true
	}), mock.MatchedBy(func(options map[string]string) bool {
		assert.Equal(t, expectedOptions, options, "Options should be populated with params")
		return true
	})).Return(map[int64]map[string]float64{}, nil)

	handler := metrics_handler(routesConfig, mockMetricsService)

	// TEST params population
	record := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health?database_id=test_db&start=0000&end=1111", nil)
	handler.ServeHTTP(record, req)
	assert.Equal(t, http.StatusOK, record.Code)

	// Arbitrary param order
	record = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/health?end=1111&start=0000&database_id=test_db", nil)
	handler.ServeHTTP(record, req)
	assert.Equal(t, http.StatusOK, record.Code)

}

func TestMissingParams(t *testing.T) {
	route := "/v1/health"
	routesConfig := map[string]RouteConfig{
		route: {
			Params: []string{"database_id", "start", "end"},
			Options: map[string]string{
				"start": "$start",
				"end":   "$end",
			},
			Metrics: map[string]string{
				"connections": "sum(pg_stat_database_numbackends{datname=~\"$database_id\"})/sum(pg_settings_max_connections)",
			},
		},
	}

	mockMetricsService := new(MockMetricsService)

	mockMetricsService.On("Execute", mock.Anything, mock.Anything).Return(map[string]string{"connections": "12"}, nil)
	handler := metrics_handler(routesConfig, mockMetricsService)

	record := httptest.NewRecorder()

	// end time missing, but specified in options, expect bad request
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health?database_id=test_db&start=0000", nil)
	handler.ServeHTTP(record, req)
	assert.Equal(t, http.StatusBadRequest, record.Code)

}

func TestMetricsHandlerJSONFormat(t *testing.T) {

	mockMetricsService := new(MockMetricsService)
	mockMetricsService.On("Execute", mock.Anything, mock.Anything).Return(
		map[int64]map[string]float64{
			1627845600000: map[string]float64{"cpu": 0.75, "disk": 0.65},
			1627845660000: map[string]float64{"cpu": 0.80, "disk": 0.60},
		},
		nil,
	)

	routesConfig := map[string]RouteConfig{
		"/metrics": {
			Params:  []string{},
			Options: map[string]string{},
			Metrics: map[string]string{},
		},
	}

	handler := metrics_handler(routesConfig, mockMetricsService)

	record := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/metrics", nil)
	handler.ServeHTTP(record, req)
	assert.Equal(t, http.StatusOK, record.Code)

	assert.Equal(t, http.StatusOK, record.Code)

	var result []map[string]interface{}
	err := json.Unmarshal(record.Body.Bytes(), &result)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	expectedResult := []map[string]interface{}{
		{
			"time_ms": float64(1627845600000),
			"cpu":     0.75,
			"disk":    0.65,
		},
		{
			"time_ms": float64(1627845660000),
			"cpu":     0.80,
			"disk":    0.60,
		},
	}

	assert.ElementsMatch(t, expectedResult, result)
}
