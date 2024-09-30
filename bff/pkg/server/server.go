package server

import (
	"bufio"
	"encoding/json"
	"fmt"
	"local/bff/pkg/metrics"
	"net/http"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-playground/validator/v10"
)

type Server interface {
	Run() error
}

type RouteConfig struct {
	Params  []string          `json:"params"`
	Options map[string]string `json:"options"`
	Metrics map[string]string `json:"metrics"`
}

type server_imp struct {
	routes_config   map[string]RouteConfig
	metrics_service metrics.Service
	port            string
	dbIdentifiers   []string
	webappPath      string
	inputValidator  *validator.Validate
}

type InstanceInfo struct {
	DBIdentifier string `json:"dbIdentifier"`
	SystemID     string `json:"systemId"`
	SystemScope  string `json:"systemScope"`
	SystemType   string `json:"systemType"`
}

type ValidationErrorResponse struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

const api_prefix = "/api"
const replace_prefix = "$"

const (
	SYSTEM_ID_MAX_LENGTH      = 63
	AWS_REGION_MAX_LENGTH     = 50
	CLUSTER_PREFIX_MAX_LENGTH = 11
	AWS_ACCOUNT_ID_LENGTH     = 12
	SYSTEM_SCOPE_MIN_LENGTH   = 13
	SYSTEM_SCOPE_MAX_LENGTH   = 73
)
const awsRegionRegexPattern = `^[a-z]{1,20}(-[a-z0-9]{1,20})*$`

var awsRegionRegex = regexp.MustCompile(awsRegionRegexPattern)

func CORS(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type")

		if r.Method == http.MethodOptions {
			return
		}

		next.ServeHTTP(w, r)
	})
}

func CreateServer(r map[string]RouteConfig, m metrics.Service, port string, dbIdentifiers []string, webappPath string) Server {
	return server_imp{r, m, port, dbIdentifiers, webappPath, CreateValidator()}
}

func CreateValidator() *validator.Validate {

	validate := validator.New(validator.WithRequiredStructEnabled())
	validate.RegisterValidation("duration", ValidateDuration)

	validate.RegisterValidation("dbIdentifier", ValidateDbIdentifier)

	validate.RegisterValidation("databaseList", ValidateDatabaseList)

	validate.RegisterValidation("dim", ValidateDim)

	validate.RegisterValidation("filterDimSelected", ValidateFilterDimSelected)

	validate.RegisterValidation("afterStart", ValidateAfterStart)
	return validate
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		return false
	}
	return true
}

func (s server_imp) Run() error {
	r := chi.NewRouter()

	r.Use(CORS)

	r.Get("/api/v1/activity", activity_handler(s.metrics_service, s.inputValidator))
	r.Get("/api/v1/instance", info_handler(s.metrics_service, s.dbIdentifiers))
	r.Get("/api/v1/instance/{dbIdentifier}/database", databases_handler(s.metrics_service))

	r.Route(api_prefix, func(r chi.Router) {
		r.Mount("/", metrics_handler(s.routes_config, s.dbIdentifiers, s.metrics_service))
	})

	fs := http.FileServer(http.Dir(s.webappPath))

	r.Handle("/*", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		filePath := filepath.Join(s.webappPath, r.URL.Path[1:])
		if fileExists(filePath) {
			fs.ServeHTTP(w, r)
		} else {
			http.ServeFile(w, r, filepath.Join(s.webappPath, "index.html"))
		}
	}))

	return http.ListenAndServe(":"+s.port, r)
}

func ReadDbIdentifiers(configFile string) ([]string, error) {
	file, err := os.Open(configFile)
	if err != nil {
		return nil, fmt.Errorf("error opening config file: %v", err)
	}
	defer file.Close()

	var identifiers []string
	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if !strings.HasPrefix(line, "#") {
			if strings.HasPrefix(line, "aws_db_instance_id") {
				parts := strings.SplitN(line, "=", 2)
				if len(parts) == 2 {
					identifier := strings.TrimSpace(parts[1])
					identifiers = append(identifiers, identifier)
				}
			}
		}
	}

	if err := scanner.Err(); err != nil {
		return nil, fmt.Errorf("error reading config file: %v", err)
	}

	if len(identifiers) == 0 {
		return []string{}, nil
	}

	return identifiers, nil
}

func convertDbIdentifiersToPromQLParam(identifiers []string) string {
	if len(identifiers) == 0 {
		return ""
	}

	quoted := make([]string, len(identifiers))
	for i, id := range identifiers {
		quoted[i] = fmt.Sprintf("%s", id)
	}

	result := strings.Join(quoted, "|")

	// Add parentheses if there are multiple identifiers
	if len(identifiers) > 1 {
		result = "(" + result + ")"
	}

	return result
}

func metrics_handler(route_configs map[string]RouteConfig, dbIdentifiers []string, metrics_service metrics.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		route := strings.TrimPrefix(r.URL.Path, api_prefix)

		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")

		now := time.Now()

		startTime, err := parseTimeParameter(start, now)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		var endTime time.Time
		if end != "" {
			endTime, err = parseTimeParameter(end, now)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		} else {
			endTime = now
		}

		if startTime.After(endTime) {
			http.Error(w, "Parameter 'end' must be greater than 'start'", http.StatusBadRequest)
			return
		}

		options := map[string]string{
			"start": strconv.FormatInt(startTime.UnixMilli(), 10),
			"end":   strconv.FormatInt(endTime.UnixMilli(), 10),
		}
		metrics := make(map[string]string)

		if route_config, ok := route_configs[route]; ok {
			for _, param := range route_config.Params {
				value := r.URL.Query().Get(param)
				if param == "start" || param == "end" {
					continue
				}

				if param == "dbidentifier" && value == "" {
					for metric, query := range route_config.Metrics {
						var current_query string
						if metrics[metric] == "" {
							current_query = query
						} else {
							current_query = metrics[metric]
						}

						metrics[metric] = strings.ReplaceAll(current_query, replace_prefix+param, convertDbIdentifiersToPromQLParam(dbIdentifiers))

					}
				} else if value != "" {
					for metric, query := range route_config.Metrics {
						var current_query string
						if metrics[metric] == "" {
							current_query = query
						} else {
							current_query = metrics[metric]
						}

						metrics[metric] = strings.ReplaceAll(current_query, replace_prefix+param, value)

					}

					for option, input := range route_config.Options {
						var current_input string
						if options[option] == "" {
							current_input = input
						} else {
							current_input = options[option]
						}

						options[option] = strings.ReplaceAll(current_input, replace_prefix+param, value)

					}
				}

			}
			for metric, query := range metrics {
				if strings.Contains(query, replace_prefix) {
					http.Error(w, "Query for Metric: "+metric+" still contains unresolved params: "+query, http.StatusBadRequest)
					return
				}
			}

			for option, input := range options {
				if strings.Contains(input, replace_prefix) {
					http.Error(w, "Option: "+option+" still contains unresolved params: "+input, http.StatusBadRequest)
					return
				}
			}

			results, err := metrics_service.Execute(metrics, options)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			var metrics []map[string]interface{}

			for time, record := range results {
				metric_record := make(map[string]interface{})
				metric_record["time_ms"] = time
				for metric, value := range record {
					metric_record[metric] = value
				}
				metrics = append(metrics, metric_record)
			}

			sort.Slice(metrics, func(i, j int) bool {
				return metrics[i]["time_ms"].(int64) < metrics[j]["time_ms"].(int64)
			})

			js, err := json.Marshal(metrics)
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			currentTime := time.Now().UnixNano() / int64(time.Millisecond)
			wrappedJSON, err := WrapJSON(js, map[string]interface{}{"server_now": currentTime})
			if err != nil {
				http.Error(w, err.Error(), http.StatusInternalServerError)
				return
			}

			w.WriteHeader(http.StatusOK)
			w.Write(wrappedJSON)
		} else {
			http.Error(w, "No matching route found", http.StatusNotFound)
		}
	})
}

type ActivityParams struct {
	DbIdentifier string `validate:"required,dbIdentifier"`
	DatabaseList string `validate:"required,databaseList"`

	Start string `validate:"required"`
	End   string `validate:"required,afterStart"`
	Step  string `validate:"required,duration"`

	Legend            string `validate:"required,dim"`
	Dim               string `validate:"required,dim"`
	FilterDim         string `validate:"omitempty,dim"`
	FilterDimSelected string `validate:"omitempty,filterDimSelected"`

	Limit       string `validate:"omitempty,numeric,gt=0"`
	LimitLegend string `validate:"omitempty,numeric,gt=0"`
}

func activity_handler(metrics_service metrics.Service, validate *validator.Validate) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		params := ActivityParams{
			DbIdentifier:      r.URL.Query().Get("dbidentifier"),
			DatabaseList:      r.URL.Query().Get("database_list"),
			Start:             r.URL.Query().Get("start"),
			End:               r.URL.Query().Get("end"),
			Step:              r.URL.Query().Get("step"),
			Legend:            r.URL.Query().Get("legend"),
			Dim:               r.URL.Query().Get("dim"),
			FilterDim:         r.URL.Query().Get("filterdim"),
			FilterDimSelected: r.URL.Query().Get("filterdimselected"),
			Limit:             r.URL.Query().Get("limitdim"),
			LimitLegend:       r.URL.Query().Get("limitlegend"),
		}

		if err := validate.Struct(params); err != nil {
			fmt.Println("Validation failed: ", err)
			http.Error(w, fmt.Sprintf("Validation failed: %v", err), http.StatusBadRequest)
			return
		}

		now := time.Now()

		startTime, err := parseTimeParameter(params.Start, now)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		endTime, err := parseTimeParameter(params.End, now)
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		limitValue := 0
		limitLegendValue := 0
		offsetValue := 0

		if params.Limit != "" {
			limitValue, err = strconv.Atoi(params.Limit)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		if params.LimitLegend != "" {
			limitLegendValue, err = strconv.Atoi(params.LimitLegend)
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		dbListEscaped := strconv.Quote(params.DatabaseList)[1 : len(params.DatabaseList)-1]
		filterDimSelectedEscaped := strconv.Quote(params.FilterDimSelected)[1 : len(params.FilterDimSelected)-1]

		promQLInput := PromQLInput{
			DatabaseList:      dbListEscaped,
			Start:             startTime,
			End:               endTime,
			Legend:            params.Legend,
			Dim:               params.Dim,
			FilterDim:         params.FilterDim,
			FilterDimSelected: filterDimSelectedEscaped,
			Limit:             limitValue,
			LimitLegend:       limitLegendValue,
			Offset:            offsetValue,
		}

		query, err := GenerateActivityCubePromQLQuery(promQLInput)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		options := map[string]string{
			"start": strconv.FormatInt(startTime.UnixMilli(), 10),
			"end":   strconv.FormatInt(endTime.UnixMilli(), 10),
			"step":  params.Step,
			"dim":   params.Dim,
		}

		if params.LimitLegend != "" {
			options["limitlegend"] = params.LimitLegend
			options["legend"] = params.Legend
		}

		results, err := metrics_service.ExecuteRaw(query, options)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		js, err := json.Marshal(results)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		currentTime := now.UnixMilli()
		wrappedJSON, err := WrapJSON(js, map[string]interface{}{"server_now": currentTime})
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.WriteHeader(http.StatusOK)
		w.Write(wrappedJSON)

	})
}

func databases_handler(metrics_service metrics.Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		// dbIdentifier := chi.URLParam(r, "dbIdentifier")

		query := "crystal_all_databases"
		options := make(map[string]string)

		results, err := metrics_service.ExecuteRaw(query, options)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		var dbNames []string

		for _, result := range results {
			metric, ok := result["metric"].(map[string]interface{})
			if !ok {
				continue
			}
			if datname, ok := metric["datname"].(string); ok {
				dbNames = append(dbNames, datname)
			}
		}

		js, err := json.Marshal(dbNames)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(js)

	})
}

func info_handler(metrics_service metrics.Service, dbIdentifiers []string) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

		var instances []InstanceInfo

		dummySystemScope := "us-west-2"
		dummySystemType := "amazon_rds"

		for _, dbIdentifier := range dbIdentifiers {
			instance := InstanceInfo{
				DBIdentifier: dbIdentifier,
				SystemID:     dbIdentifier,
				SystemScope:  dummySystemScope,
				SystemType:   dummySystemType,
			}
			instances = append(instances, instance)
		}

		response := map[string]interface{}{
			"list": instances,
		}

		js, err := json.Marshal(response)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.WriteHeader(http.StatusOK)
		w.Write(js)
	})
}

func parseAndValidateInt(param string, paramName string) (int, error) {
	if param == "" {
		return 0, fmt.Errorf("Missing param/value: %s", paramName)
	}

	value, err := strconv.Atoi(param)
	if err != nil {
		return 0, fmt.Errorf("Invalid %s: %s", paramName, err.Error())
	}

	if value <= 0 {
		return 0, fmt.Errorf("%s must be a positive integer", paramName)
	}

	return value, nil
}

func WrapJSON(data []byte, metadata map[string]interface{}) ([]byte, error) {
	container := map[string]interface{}{
		"data": json.RawMessage(data),
	}

	for key, value := range metadata {
		container[key] = value
	}

	wrappedJSON, err := json.Marshal(container)
	if err != nil {
		return nil, err
	}

	return wrappedJSON, nil
}

func ValidateAfterStart(fl validator.FieldLevel) bool {
	start := fl.Parent().FieldByName("Start").String()
	end := fl.Field().String()

	now := time.Now()

	startTime, err := parseTimeParameter(start, now)
	if err != nil {
		return false
	}

	endTime, err := parseTimeParameter(end, now)
	if err != nil {
		return false
	}

	return endTime.After(startTime) || endTime.Equal(startTime)
}

func ValidateFilterDimSelected(fl validator.FieldLevel) bool {
	filterDim := fl.Parent().FieldByName("FilterDim").String()
	filterDimSelected := fl.Field().String()

	if filterDim == "Time" {
		if _, err := strconv.ParseInt(filterDimSelected, 10, 64); err == nil {
			return true
		}
		return false
	}

	// If filterDim is not "Time", accept any arbitrary string
	return len(filterDimSelected) > 0
}

func ValidateDatabaseList(fl validator.FieldLevel) bool {
	regex := `^(?:[a-zA-Z0-9_-]+|\([a-zA-Z0-9_-]+(\|[a-zA-Z0-9_-]+)*\))$`
	matched, _ := regexp.MatchString(regex, fl.Field().String())
	return matched

}

func ValidateDuration(fl validator.FieldLevel) bool {
	_, err := time.ParseDuration(fl.Field().String())
	return err == nil // Return true if parsing was successful
}

func ValidateDim(fl validator.FieldLevel) bool {
	value := fl.Field().String()
	return isValidDimension(value)
}

var validSystemTypes = []string{
	"amazon_rds",
	"google_cloudsql",
	"azure_database",
	"heroku",
	"crunchy_bridge",
	"aiven",
	"tembo",
	"self_hosted",
}

// A dbIdentifer is defined as <SystemID>-<SystemScope>-<SystemType>
// SystemID: min 1 char maxx 63char
// SystemScope:  <AWS_REGION>/<CLUSTER_PREFIX><AWS_ACCOUNT_ID> where AWS_REGION is between 1 and 50 characters, CLUSTER_PREFIX is between 0 and 11 characters, and AWS_ACCOUNT_ID is 12 characters. Then, the whole SystemScope is betwen 13 and 73 characters.
// SystemType: One of the following values :
// amazon_rds
// google_cloudsql
// azure_database
// heroku
// crunchy_bridge
// aiven
// tembo
// self_hosted

func ValidateDbIdentifier(fl validator.FieldLevel) bool {
	dbIdentifier := fl.Field().String()

	// TODO remove
	return true

	// Split the dbIdentifier into components using "+" as the separator
	parts := strings.Split(dbIdentifier, "+")
	if len(parts) != 3 {
		return false
	}

	systemID := parts[0]
	systemScope := parts[1]
	systemType := parts[2]

	// Validate SystemID
	if len(systemID) < 1 || len(systemID) > SYSTEM_ID_MAX_LENGTH {
		return false
	}

	// Validate SystemScope
	scopeParts := strings.Split(systemScope, "/")
	if len(scopeParts) != 2 {
		return false
	}

	awsRegion := scopeParts[0]
	clusterPrefixAccountID := scopeParts[1]

	// Validate AWS_REGION using regex
	if !awsRegionRegex.MatchString(awsRegion) {
		return false
	}

	// Validate length of clusterPrefixAccountID
	if len(clusterPrefixAccountID) < AWS_ACCOUNT_ID_LENGTH || len(clusterPrefixAccountID) > (AWS_ACCOUNT_ID_LENGTH+CLUSTER_PREFIX_MAX_LENGTH) {
		return false // Ensure total length is within limits
	}

	// Extract CLUSTER_PREFIX and AWS_ACCOUNT_ID
	clusterPrefix := clusterPrefixAccountID[:len(clusterPrefixAccountID)-AWS_ACCOUNT_ID_LENGTH]
	awsAccountID := clusterPrefixAccountID[len(clusterPrefixAccountID)-AWS_ACCOUNT_ID_LENGTH:]

	// Validate lengths of clusterPrefix and awsAccountID
	if len(clusterPrefix) > CLUSTER_PREFIX_MAX_LENGTH || len(awsAccountID) != AWS_ACCOUNT_ID_LENGTH {
		return false
	}

	// Validate SystemScope length
	totalScopeLength := len(awsRegion) + len(clusterPrefixAccountID) + 1 // +1 for '/'
	if totalScopeLength < SYSTEM_SCOPE_MIN_LENGTH || totalScopeLength > SYSTEM_SCOPE_MAX_LENGTH {
		return false
	}

	// Validate SystemType
	isValidType := false
	for _, validType := range validSystemTypes {
		if systemType == validType {
			isValidType = true
			break
		}
	}

	return isValidType
}
