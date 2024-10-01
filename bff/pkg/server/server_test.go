package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"local/bff/pkg/utils"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

type MockMetricsService struct {
	mock.Mock
}

func (m *MockMetricsService) Execute(metrics map[string]string, options map[string]string) (map[int64]map[string]float64, error) {
	args := m.Called(metrics, options)
	return args.Get(0).(map[int64]map[string]float64), args.Error(1)
}

func (m *MockMetricsService) ExecuteRaw(query string, options map[string]string) ([]map[string]interface{}, error) {
	args := m.Called(query, options)
	return args.Get(0).([]map[string]interface{}), args.Error(1)
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
			Params: []string{"start", "end"},
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
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health?start=now&dbidentifier=default_db", nil)
	handler.ServeHTTP(record, req)
	assert.Equal(t, http.StatusOK, record.Code)

	// TEST route not found
	record = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v2/health?start=now&dbidentifier=default_db", nil)
	handler.ServeHTTP(record, req)
	assert.Equal(t, http.StatusNotFound, record.Code)
}

func TestParamsPopulation(t *testing.T) {
	route := "/v1/health"
	routesConfig := map[string]RouteConfig{
		route: {
			Params: []string{"datname", "dbidentifier", "start", "end"},
			Options: map[string]string{
				"start": "$start",
				"end":   "$end",
			},
			Metrics: map[string]string{
				"connections": "sum(pg_stat_database_numbackends{datname=~\"$datname\"})/sum(pg_settings_max_connections)",
			},
		},
	}

	expectedMetrics := map[string]string{
		"connections": "sum(pg_stat_database_numbackends{datname=~\"test_db\"})/sum(pg_settings_max_connections)",
	}

	expectedOptions := map[string]string{
		"start": "0",
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
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health?datname=test_db&start=0000&end=1111&dbidentifier=default_db", nil)
	handler.ServeHTTP(record, req)
	assert.Equal(t, http.StatusOK, record.Code)

	// Arbitrary param order
	record = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/health?end=1111&start=0000&datname=test_db&dbidentifier=default_db", nil)
	handler.ServeHTTP(record, req)
	assert.Equal(t, http.StatusOK, record.Code)

}

func TestMissingParams(t *testing.T) {
	route := "/v1/health"
	routesConfig := map[string]RouteConfig{
		route: {
			Params: []string{"datname", "dbidentifier", "start", "end"},
			Options: map[string]string{
				"start": "$start",
				"end":   "$end",
			},
			Metrics: map[string]string{
				"connections": "sum(pg_stat_database_numbackends{datname=~\"$datname\"})/sum(pg_settings_max_connections)",
			},
		},
	}

	mockMetricsService := new(MockMetricsService)

	mockMetricsService.On("Execute", mock.Anything, mock.Anything).Return(
		map[int64]map[string]float64{
			1620000000000: {"connections": 12.0},
		},
		nil,
	)
	handler := metrics_handler(routesConfig, mockMetricsService)

	record := httptest.NewRecorder()

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health?start=0000&dbidentifier=default_db", nil)
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
		"/v1/metrics": {
			Params: []string{"start", "end"},
			Options: map[string]string{
				"start": "$start",
				"end":   "$end",
			},
			Metrics: map[string]string{
				"cpu":  "",
				"disk": "",
			},
		},
	}

	handler := metrics_handler(routesConfig, mockMetricsService)

	record := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics?start=now&dbidentifier=default_db", nil)
	handler.ServeHTTP(record, req)
	assert.Equal(t, http.StatusOK, record.Code)

	fmt.Println("Response body: ", record.Body.String())
	var result map[string]interface{}
	err := json.Unmarshal(record.Body.Bytes(), &result)
	if err != nil {
		t.Fatalf("Failed to parse response body: %v", err)
	}

	data := result["data"]

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

	assert.ElementsMatch(t, expectedResult, data)
}

func TestMissingParameters(t *testing.T) {
	mockService := new(MockMetricsService)
	handler := activity_handler(mockService)

	defaultParams := map[string]string{
		"dbidentifier":  "test",
		"database_list": "postgres",
		"start":         "10",
		"end":           "20",
		"step":          "5000ms",
		"legend":        "wait_event_name",
		"dim":           "time",
	}

	expectedErrors := map[string]string{
		"database_list": "Missing param/value: database_list",
		"start":         "Missing param/value: start",
		"end":           "Missing param/value: end",
		"step":          "Missing param/value: step",
		"legend":        "Missing param/value: legend",
		"dim":           "Missing param/value: dim",
	}

	for paramToRemove, expectedError := range expectedErrors {
		params := make(map[string]string)
		for k, v := range defaultParams {
			params[k] = v
		}
		delete(params, paramToRemove)

		record := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/activity?"+formatQueryParams(params), nil)

		handler.ServeHTTP(record, req)

		assert.Equal(t, http.StatusBadRequest, record.Code)
		assert.Contains(t, record.Body.String(), expectedError)
	}
}

func TestActivityValidationLogic(t *testing.T) {
	mockService := new(MockMetricsService)
	handler := activity_handler(mockService)

	tests := []struct {
		queryParams    map[string]string
		expectedStatus int
		expectedError  string
	}{
		{map[string]string{"start": "10", "end": "5", "limitdim": "10"}, http.StatusBadRequest, "Parameter 'end' must be greater than 'start'"},
		{map[string]string{"start": "10", "end": "20", "limitdim": "0"}, http.StatusBadRequest, "limitdim must be a positive integer"},
		{map[string]string{"start": "10", "end": "20", "limitlegend": "0"}, http.StatusBadRequest, "limitlegend must be a positive integer"},
	}

	defaultParams := map[string]string{"dbidentifier": "test", "database_list": "postgres", "step": "5000ms", "legend": "wait_event_name", "dim": "time", "filterdim": ""}

	mockService.On("ExecuteRaw", mock.Anything, mock.Anything).Return([]map[string]interface{}{}, nil)

	for _, test := range tests {

		record := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/activity?"+formatQueryParams(utils.MergeMaps(test.queryParams, defaultParams)), nil)

		handler.ServeHTTP(record, req)
		assert.Equal(t, test.expectedStatus, record.Code)
		assert.Contains(t, record.Body.String(), test.expectedError)
	}
}
func TestOptions(t *testing.T) {
	mockService := new(MockMetricsService)
	handler := activity_handler(mockService)

	expectedOptions := map[string]string{
		"start": "10",
		"end":   "20",
		"step":  "5000ms",
		"dim":   "time",
	}

	params := map[string]string{
		"dbidentifier":  "test",
		"database_list": "postgres",
		"start":         "10",
		"end":           "20",
		"step":          "5000ms",
		"legend":        "wait_event_name",
		"dim":           "time",
	}

	mockService.On("ExecuteRaw", mock.Anything, expectedOptions).Return([]map[string]interface{}{}, nil)

	record := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/activity?"+formatQueryParams(params), nil)
	handler.ServeHTTP(record, req)

	mockService.AssertExpectations(t)
}
func TestDefaultDBIdentifier(t *testing.T) {
	mockService := new(MockMetricsService)

	routeConfigs := map[string]RouteConfig{
		"/v1/test": {
			Params: []string{"dbidentifier"},
			Metrics: map[string]string{
				"cpu": "sum(rds_cpu_usage_percent_average{dbidentifier=~\"$dbidentifier\"})",
			},
			Options: map[string]string{},
		},
	}

	mockService.On("Execute", mock.Anything, mock.Anything).
		Run(func(args mock.Arguments) {
			metrics := args.Get(0).(map[string]string)
			assert.Equal(t, "sum(rds_cpu_usage_percent_average{dbidentifier=~\"default_db\"})", metrics["cpu"])
		}).
		Return(map[int64]map[string]float64{}, nil)

	// Create a request without the dbidentifier parameter
	record := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/test?start=now&dbidentifier=default_db", nil)
	handler := metrics_handler(routeConfigs, mockService)
	handler.ServeHTTP(record, req)

	mockService.AssertExpectations(t)
}

func TestDatabasesHandler(t *testing.T) {
	mockService := new(MockMetricsService)
	mockData := []map[string]interface{}{
		{
			"metric": map[string]interface{}{
				"__name__": "crystal_all_databases",
				"datname":  "postgres",
				"instance": "localhost:9399",
				"job":      "sqlexport",
				"target":   "sqlexport",
			},
			"values": []map[string]interface{}{
				{"timestamp": 1724943882794, "value": 1},
			},
		},
		{
			"metric": map[string]interface{}{
				"__name__": "crystal_all_databases",
				"datname":  "rdsadmin",
				"instance": "localhost:9399",
				"job":      "sqlexport",
				"target":   "sqlexport",
			},
			"values": []map[string]interface{}{
				{"timestamp": 1724943882794, "value": 1},
			},
		},
	}
	mockService.On("ExecuteRaw", "crystal_all_databases", map[string]string{}).Return(mockData, nil)

	handler := databases_handler(mockService)

	record := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instance/DBIDENTIFIER_IGNORED/database", nil)

	handler.ServeHTTP(record, req)
	assert.Equal(t, http.StatusOK, record.Code)

	expectedDbNames := []string{"postgres", "rdsadmin"}
	var dbNames []string
	err := json.Unmarshal(record.Body.Bytes(), &dbNames)
	if err != nil {
		t.Fatalf("Failed to unmarshal response body: %v", err)
	}

	assert.ElementsMatch(t, expectedDbNames, dbNames, "Response body should contain the correct database names")

	mockService.AssertExpectations(t)
}

func TestInfoHandler(t *testing.T) {
	dbIdentifiers := []InstanceInfo{
		{
			DBIdentifier: "test1",
			SystemID:     "test1",
			SystemScope:  "us-west-2",
			SystemType:   "amazon_rds",
		},
		{
			DBIdentifier: "test2",
			SystemID:     "test2",
			SystemScope:  "us-west-2",
			SystemType:   "amazon_rds",
		},
	}
	handler := info_handler_internal(dbIdentifiers)

	record := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instance", nil)

	handler.ServeHTTP(record, req)

	// Assert that the status code is OK
	assert.Equal(t, http.StatusOK, record.Code)

	expectedInstances := []InstanceInfo{
		{
			DBIdentifier: "test1",
			SystemID:     "test1",
			SystemScope:  "us-west-2",
			SystemType:   "amazon_rds",
		},
		{
			DBIdentifier: "test2",
			SystemID:     "test2",
			SystemScope:  "us-west-2",
			SystemType:   "amazon_rds",
		},
	}

	expectedResponse := struct {
		List []InstanceInfo `json:"list"`
	}{
		List: expectedInstances,
	}

	expectedJSON, err := json.Marshal(expectedResponse)
	if err != nil {
		t.Fatalf("Failed to marshal expected response: %v", err)
	}

	assert.JSONEq(t, string(expectedJSON), record.Body.String(), "Response should match the expected JSON")

}

func TestConvertDbIdentifiersToPromQLParam(t *testing.T) {
	testCases := []struct {
		name        string
		identifiers []string
		expected    string
	}{
		{
			name:        "Single identifier",
			identifiers: []string{"instance1"},
			expected:    "instance1",
		},
		{
			name:        "Multiple identifiers",
			identifiers: []string{"instance1", "instance2", "instance3"},
			expected:    `(instance1|instance2|instance3)`,
		},
		{
			name:        "Empty list",
			identifiers: []string{},
			expected:    "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := convertDbIdentifiersToPromQLParam(tc.identifiers)
			if result != tc.expected {
				t.Errorf("Expected %s, got %s", tc.expected, result)
			}
		})
	}
}

func formatQueryParams(params map[string]string) string {
	var buffer bytes.Buffer
	for key, value := range params {
		if buffer.Len() > 0 {
			buffer.WriteString("&")
		}
		buffer.WriteString(key)
		buffer.WriteString("=")
		buffer.WriteString(value)
	}
	return buffer.String()
}
