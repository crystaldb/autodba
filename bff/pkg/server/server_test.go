package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"strconv"
	"time"

	"github.com/go-playground/validator/v10"
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
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health?start=now&dbidentifier=amazon_rds/default_db/us-west-2/cvirkksghnig", nil)
	handler.ServeHTTP(record, req)
	assert.Equal(t, http.StatusOK, record.Code)

	// TEST route not found
	record = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v2/health?start=now&dbidentifier=amazon_rds/default_db/us-west-2/cvirkksghnig", nil)
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
	req := httptest.NewRequest(http.MethodGet, "/api/v1/health?datname=test_db&start=0000&end=1111&dbidentifier=amazon_rds/default_db/us-west-2/cvirkksghnig", nil)
	handler.ServeHTTP(record, req)
	assert.Equal(t, http.StatusOK, record.Code)

	// Arbitrary param order
	record = httptest.NewRecorder()
	req = httptest.NewRequest(http.MethodGet, "/api/v1/health?end=1111&start=0000&datname=test_db&dbidentifier=amazon_rds/default_db/us-west-2/cvirkksghnig", nil)
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

	req := httptest.NewRequest(http.MethodGet, "/api/v1/health?start=0000&dbidentifier=amazon_rds/default_db/us-west-2/cvirkksghnig", nil)
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
	req := httptest.NewRequest(http.MethodGet, "/api/v1/metrics?start=now&dbidentifier=amazon_rds/default_db/us-west-2/cvirkksghnig", nil)
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

	validate := CreateValidator()
	handler := activity_handler(mockService, validate)

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
		"database_list": "DatabaseList",
		"start":         "Start",
		"end":           "End",
		"step":          "Step",
		"legend":        "Legend",
		"dim":           "Dim",
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

	validate := CreateValidator()
	handler := activity_handler(mockService, validate)

	tests := []struct {
		name           string
		queryParams    map[string]string
		expectedStatus int
	}{
		{
			name: "Valid parameters",
			queryParams: map[string]string{
				"dbidentifier":  "amazon_rds/mohammad-dashti-rds-1/us-west-2/cvirkksghnig",
				"database_list": "postgres",
				"start":         "now-1h",
				"end":           "now",
				"step":          "1m",
				"legend":        "wait_event_name",
				"dim":           "time",
			},
			expectedStatus: http.StatusOK,
		},
		{
			name: "Missing dbidentifier",
			queryParams: map[string]string{
				"database_list": "postgres",
				"start":         "now-1h",
				"end":           "now",
				"step":          "1m",
				"legend":        "wait_event_name",
				"dim":           "time",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing database_list",
			queryParams: map[string]string{
				"dbidentifier": "amazon_rds/mohammad-dashti-rds-1/us-west-2/cvirkksghnig",
				"start":        "now-1h",
				"end":          "now",
				"step":         "1m",
				"legend":       "wait_event_name",
				"dim":          "time",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing start",
			queryParams: map[string]string{
				"dbidentifier":  "amazon_rds/mohammad-dashti-rds-1/us-west-2/cvirkksghnig",
				"database_list": "postgres",
				"end":           "now",
				"step":          "1m",
				"legend":        "wait_event_name",
				"dim":           "time",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing end",
			queryParams: map[string]string{
				"dbidentifier":  "amazon_rds/mohammad-dashti-rds-1/us-west-2/cvirkksghnig",
				"database_list": "postgres",
				"start":         "now",
				"step":          "1m",
				"legend":        "wait_event_name",
				"dim":           "time",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing legend",
			queryParams: map[string]string{
				"dbidentifier":  "amazon_rds/mohammad-dashti-rds-1/us-west-2/cvirkksghnig",
				"database_list": "postgres",
				"start":         "now-1h",
				"end":           "now",
				"step":          "1m",
				"dim":           "time",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Missing Dim",
			queryParams: map[string]string{
				"dbidentifier":  "amazon_rds/mohammad-dashti-rds-1/us-west-2/cvirkksghnig",
				"database_list": "postgres",
				"start":         "now-1h",
				"end":           "now",
				"step":          "1m",
				"legend":        "wait_event_name",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Bad Dim",
			queryParams: map[string]string{
				"dbidentifier":  "amazon_rds/mohammad-dashti-rds-1/us-west-2/cvirkksghnig",
				"database_list": "postgres",
				"start":         "now-1h",
				"end":           "now",
				"step":          "1m",
				"legend":        "wait_event_name",
				"dim":           "invalid",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "Invalid time range",
			queryParams: map[string]string{
				"start":         "now",
				"end":           "now-1h", // end before start
				"dbidentifier":  "amazon_rds/mohammad-dashti-rds-1/us-west-2/cvirkksghnig",
				"database_list": "postgres",
				"step":          "1m",
				"legend":        "wait_event_name",
				"dim":           "time",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	mockService.On("ExecuteRaw", mock.Anything, mock.Anything).Return([]map[string]interface{}{}, nil)

	for _, test := range tests {

		record := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/api/v1/activity?"+formatQueryParams(test.queryParams), nil)

		handler.ServeHTTP(record, req)
		assert.Equal(t, test.expectedStatus, record.Code)
	}
}
func TestOptions(t *testing.T) {
	mockService := new(MockMetricsService)

	validate := CreateValidator()
	handler := activity_handler(mockService, validate)

	expectedOptions := map[string]string{
		"start": "10",
		"end":   "20",
		"step":  "5000ms",
		"dim":   "time",
	}

	params := map[string]string{
		"dbidentifier":  "amazon_rds/mohammad-dashti-rds-1/us-west-2/cvirkksghnig",
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
			assert.Equal(t, "sum(rds_cpu_usage_percent_average{dbidentifier=~\"amazon_rds/default_db/us-west-2/cvirkksghnig\"})", metrics["cpu"])
		}).
		Return(map[int64]map[string]float64{}, nil)

	// Create a request without the dbidentifier parameter
	record := httptest.NewRecorder()
	req := httptest.NewRequest("GET", "/api/v1/test?start=now&dbidentifier=amazon_rds/default_db/us-west-2/cvirkksghnig", nil)
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
	req := httptest.NewRequest(http.MethodGet, "/api/v1/instance/database?dbidentifier=DBIDENTIFIER_IGNORED", nil)

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
			DBIdentifier: "amazon_rds/test1/us-west-2",
			SystemID:     "test1",
			SystemScope:  "us-west-2",
			SystemType:   "amazon_rds",
		},
		{
			DBIdentifier: "amazon_rds/test2/us-west-2",
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
			DBIdentifier: "amazon_rds/test1/us-west-2",
			SystemID:     "test1",
			SystemScope:  "us-west-2",
			SystemType:   "amazon_rds",
		},
		{
			DBIdentifier: "amazon_rds/test2/us-west-2",
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

func TestValidateAfterStart(t *testing.T) {
	validate := validator.New()
	validate.RegisterValidation("afterStart", ValidateAfterStart)

	tests := []struct {
		start string
		end   string
		valid bool
	}{
		{"now", "now", true},
		{"now", "now-1m", false},
		{"now-1m", "now", true},
		{"now-1m", "now-1m", true},
		{"invalid", "now", false},
		{"now", "invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.start+"_"+tt.end, func(t *testing.T) {
			input := struct {
				Start string `validate:"required"`
				End   string `validate:"required,afterStart"`
			}{
				Start: tt.start,
				End:   tt.end,
			}
			err := validate.Struct(input)
			if (err == nil) != tt.valid {
				t.Errorf("expected valid: %v, got error: %v", tt.valid, err)
			}
		})
	}
}

func TestValidateFilterDimSelected(t *testing.T) {
	validate := validator.New()
	validate.RegisterValidation("filterDimSelected", ValidateFilterDimSelected)

	tests := []struct {
		filterDim         string
		filterDimSelected string
		valid             bool
	}{
		{"Time", "1627891200000", true},
		{"Time", "invalid", false},
		{"Other", "someString", true},
		{"Other", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.filterDim+"_"+tt.filterDimSelected, func(t *testing.T) {
			input := struct {
				FilterDim         string `validate:"required"`
				FilterDimSelected string `validate:"omitempty,filterDimSelected"`
			}{
				FilterDim:         tt.filterDim,
				FilterDimSelected: tt.filterDimSelected,
			}
			err := validate.Struct(input)
			if (err == nil) != tt.valid {
				t.Errorf("expected valid: %v, got error: %v", tt.valid, err)
			}
		})
	}
}

func TestValidateDatabaseList(t *testing.T) {
	validate := validator.New()
	validate.RegisterValidation("databaseList", ValidateDatabaseList)

	tests := []struct {
		dbList string
		valid  bool
	}{
		{"(db1|db2|db3)", true},
		{"db1", true},
		{"db1|", false},
		{"db1|db2", false},
		{"db1|invalid@db", false},
	}

	for _, tt := range tests {
		t.Run(tt.dbList, func(t *testing.T) {
			input := struct {
				DatabaseList string `validate:"required,databaseList"`
			}{
				DatabaseList: tt.dbList,
			}
			err := validate.Struct(input)
			if (err == nil) != tt.valid {
				t.Errorf("expected valid: %v, got error: %v", tt.valid, err)
			}
		})
	}
}

func TestValidateDbIdentifier(t *testing.T) {
	validate := validator.New()
	validate.RegisterValidation("dbIdentifier", ValidateDbIdentifier)

	tests := []struct {
		dbIdentifier string
		valid        bool
	}{
		// Valid case
		{"amazon_rds/mohammad-dashti-rds-1/us-west-2/cvirkksghnig", true},

		// Valid, no awsaccountid
		{"amazon_rds/mohammad-dashti-rds-1/us-west-2", true},

		// Multiple ids
		{"(amazon_rds/mohammad-dashti-rds-1/us-west-2/cvirkksghnig|amazon_rds/mohammad-dashti-rds-1/us-west-2/cvirkksghnig)", true},

		// One invalid id
		{"(amazon_rds/mohammad-dashti-rds-1/us-west-2/cvirkksghnig|amazon_rds/mohammad-dashti-rds-1/us-west-2/cvirkksghnig|invalid)", false},

		// Invalid list format
		{"amazon_rds/mohammad-dashti-rds-1/us-west-2/cvirkksghnig,amazon_rds/mohammad-dashti-rds-1/us-west-2/cvirkksghnig", false},

		// Invalid SystemID (too long)
		{"amazon_rds/aVeryLongSystemIDThatExceedsTheMaximumAllowedLengthForSystemIDThatIs63Chars/us-east-1/cvirkksghnig", false},

		// Invalid AWS_REGION (too long)
		{"amazon_rds/mohammad-dashti-rds-1/us-east-1-verylongregionname-whichisnotvalidbecauseitisverylong/cvirkksghnig", false},

		// Invalid CLUSTER_PREFIX (too long)
		{"amazon_rds/mohammad-dashti-rds-1/us-west-1/verylongclusterprefix123456789012/cvirkksghnig", false},
		// Invalid AWS_ACCOUNT_ID (too short)
		{"amazon_rds/mohammad-dashti-rds-1/us-west-1/cvirkksghnig123/cvirkksghnig", false},
		// Invalid AWS_ACCOUNT_ID (too long)
		{"amazon_rds/mohammad-dashti-rds-1/us-west-1/cvirkksghnig12345/cvirkksghnig", false},

		// Invalid SystemType (not valid)
		{"amazon_rds/mohammad-dashti-rds-1/us-west-1/cvirkksghnig/invalid_system_type", false},
		{"amazon_rds/mohammad-dashti-rds-1/us-west-1/cluster-ro-cvirkksghnig/extra", false}, // Extra part
	}

	for _, tt := range tests {
		t.Run(tt.dbIdentifier, func(t *testing.T) {
			input := struct {
				DbIdentifier string `validate:"required,dbIdentifier"`
			}{
				DbIdentifier: tt.dbIdentifier,
			}
			err := validate.Struct(input)
			if (err == nil) != tt.valid {
				t.Errorf("expected valid: %v, got error: %v", tt.valid, err)
			}
		})
	}
}

func TestValidateDuration(t *testing.T) {
	validate := validator.New()
	validate.RegisterValidation("duration", ValidateDuration)

	tests := []struct {
		duration string
		valid    bool
	}{
		{"10s", true},
		{"1m30s", true},
		{"invalid-duration", false},
	}

	for _, tt := range tests {
		t.Run(tt.duration, func(t *testing.T) {
			input := struct {
				Duration string `validate:"required,duration"`
			}{
				Duration: tt.duration,
			}
			err := validate.Struct(input)
			if (err == nil) != tt.valid {
				t.Errorf("expected valid: %v, got error: %v", tt.valid, err)
			}
		})
	}
}

func TestValidateDim(t *testing.T) {
	validate := validator.New()
	validate.RegisterValidation("dim", ValidateDim)

	tests := []struct {
		dimension string
		valid     bool
	}{
		{"time", true},
		{"datname", true},
		{"invalidDim", false},
	}

	for _, tt := range tests {
		t.Run(tt.dimension, func(t *testing.T) {
			input := struct {
				Dim string `validate:"required,dim"`
			}{
				Dim: tt.dimension,
			}
			err := validate.Struct(input)
			if (err == nil) != tt.valid {
				t.Errorf("expected valid: %v, got error: %v", tt.valid, err)
			}
		})
	}
}
func TestExtractPromQLInput(t *testing.T) {
	now := time.Now()
	startTime := now.Add(-10 * time.Minute)

	tests := []struct {
		name        string
		params      ActivityParams
		expected    PromQLInput
		expectError bool
	}{
		{
			name: "Valid Parameters with Limits",
			params: ActivityParams{
				DbIdentifier:      "test_db",
				DatabaseList:      "db1",
				Start:             strconv.FormatInt(startTime.UnixMilli(), 10),
				End:               strconv.FormatInt(now.UnixMilli(), 10),
				Step:              "10s",
				Legend:            "Test Legend",
				Dim:               "dim",
				FilterDim:         "filter",
				FilterDimSelected: "selected",
				Limit:             "100",
				LimitLegend:       "0",
			},
			expected: PromQLInput{
				DatabaseList:      "db1",
				Start:             startTime,
				End:               now,
				Legend:            "Test Legend",
				Dim:               "dim",
				FilterDim:         "filter",
				FilterDimSelected: "selected",
				Limit:             100,
				LimitLegend:       0,
				Offset:            0,
				DbIdentifier:      "test_db",
			},
			expectError: false,
		},
		{
			name: "Valid Parameters with Fixed Unix Timestamps",
			params: ActivityParams{
				DbIdentifier: "test_db",
				DatabaseList: "db1",
				Start:        strconv.FormatInt(startTime.UnixMilli(), 10),
				End:          strconv.FormatInt(now.UnixMilli(), 10),
				Step:         "10s",
				Limit:        "100",
				LimitLegend:  "0",
			},
			expected: PromQLInput{
				DatabaseList:      "db1",
				Start:             startTime,
				End:               now,
				Legend:            "",
				Dim:               "",
				FilterDim:         "",
				FilterDimSelected: "",
				Limit:             100,
				LimitLegend:       0,
				Offset:            0,
				DbIdentifier:      "test_db",
			},
			expectError: false,
		},
		{
			name: "Invalid Start Time",
			params: ActivityParams{
				DbIdentifier: "test_db",
				DatabaseList: "db1",
				Start:        "invalid-time",
				End:          strconv.FormatInt(now.UnixMilli(), 10),
				Step:         "10s",
				Limit:        "100",
				LimitLegend:  "0",
			},
			expected:    PromQLInput{},
			expectError: true,
		},
		{
			name: "Invalid Limit",
			params: ActivityParams{
				DbIdentifier: "test_db",
				DatabaseList: "db1",
				Start:        strconv.FormatInt(now.UnixMilli(), 10),
				End:          strconv.FormatInt(now.UnixMilli(), 10),
				Step:         "10s",
				Limit:        "invalid-limit",
				LimitLegend:  "0",
			},
			expected:    PromQLInput{},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := extractPromQLInput(tt.params, now)

			if (err != nil) != tt.expectError {
				t.Errorf("expected error: %v, got: %v", tt.expectError, err)
				return
			}

			if result.DatabaseList != tt.expected.DatabaseList {
				t.Errorf("DatabaseList: expected %s, got %s", tt.expected.DatabaseList, result.DatabaseList)
			}
			if result.Start.UnixMilli() != tt.expected.Start.UnixMilli() {
				t.Errorf("Start: expected %v, got %v", tt.expected.Start, result.Start)
			}
			if result.End.UnixMilli() != tt.expected.End.UnixMilli() {
				t.Errorf("End: expected %v, got %v", tt.expected.End, result.End)
			}
			if result.Legend != tt.expected.Legend {
				t.Errorf("Legend: expected %s, got %s", tt.expected.Legend, result.Legend)
			}
			if result.Dim != tt.expected.Dim {
				t.Errorf("Dim: expected %s, got %s", tt.expected.Dim, result.Dim)
			}
			if result.FilterDim != tt.expected.FilterDim {
				t.Errorf("FilterDim: expected %s, got %s", tt.expected.FilterDim, result.FilterDim)
			}
			if result.FilterDimSelected != tt.expected.FilterDimSelected {
				t.Errorf("FilterDimSelected: expected %s, got %s", tt.expected.FilterDimSelected, result.FilterDimSelected)
			}
			if result.Limit != tt.expected.Limit {
				t.Errorf("Limit: expected %d, got %d", tt.expected.Limit, result.Limit)
			}
			if result.LimitLegend != tt.expected.LimitLegend {
				t.Errorf("LimitLegend: expected %d, got %d", tt.expected.LimitLegend, result.LimitLegend)
			}
			if result.Offset != tt.expected.Offset {
				t.Errorf("Offset: expected %d, got %d", tt.expected.Offset, result.Offset)
			}
			if result.DbIdentifier != tt.expected.DbIdentifier {
				t.Errorf("DbIdentifier: expected %s, got %s", tt.expected.DbIdentifier, result.DbIdentifier)
			}
		})
	}
}
