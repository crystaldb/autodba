package server

import (
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

func CreateServer(r map[string]RouteConfig, m metrics.Service, port string, webappPath string) Server {
	return server_imp{r, m, port, webappPath, CreateValidator()}
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
	r.Get("/api/v1/instance", info_handler(s.metrics_service))
	r.Get("/api/v1/instance/database", databases_handler(s.metrics_service))

	r.Route(api_prefix, func(r chi.Router) {
		r.Mount("/", metrics_handler(s.routes_config, s.metrics_service))
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

func getActualDbIdentifier(dbIdentifier string) (string, error) {
	_, systemID, _, err := splitDbIdentifier(dbIdentifier)
	return systemID, err
}

func splitDbIdentifier(dbIdentifier string) (string, string, string, error) {
	parts := strings.SplitN(dbIdentifier, "/", 3)

	if len(parts) != 3 {
		return "", "", "", fmt.Errorf("invalid dbidentifier format: %s", dbIdentifier)
	}

	systemType := parts[0]
	systemID := parts[1]
	systemScope := parts[2]

	return systemType, systemID, systemScope, nil
}

func metrics_handler(route_configs map[string]RouteConfig, metrics_service metrics.Service) http.Handler {
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
					http.Error(w, "The 'dbidentifier' parameter is required and cannot be empty.", http.StatusBadRequest)
					return
				} else if value != "" {
					if param == "dbidentifier" {
						value, err = getActualDbIdentifier(value)
						if err != nil {
							http.Error(w, "The 'dbidentifier' is malformatted.", http.StatusBadRequest)
							return
						}
					}
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
			if validationErrors, ok := err.(validator.ValidationErrors); ok {
				for _, validationError := range validationErrors {
					switch validationError.Tag() {
					case "required":
						http.Error(w, fmt.Sprintf("%s is required.", validationError.Field()), http.StatusBadRequest)
						return
					case "afterStart":
						http.Error(w, "End time must be after Start time.", http.StatusBadRequest)
						return
					}
				}
			}
			// Generic error response for other validation failures
			http.Error(w, "Invalid input", http.StatusBadRequest)
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

		stepDuration, err := time.ParseDuration(params.Step)
		if err != nil {
			http.Error(w, "Invalid step duration format.", http.StatusBadRequest)
			return
		}

		totalDuration := endTime.Sub(startTime)
		totalSamples := int(totalDuration / stepDuration)
		if totalSamples > 11000 {
			http.Error(w, "Maximum time samples exceeded. 11000 samples max per query", http.StatusBadRequest)
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

		dbListEscaped := ""
		if params.DatabaseList != "" {
			dbListEscaped = strconv.Quote(params.DatabaseList)
			dbListEscaped = dbListEscaped[1 : len(dbListEscaped)-1]
		}

		filterDimSelectedEscaped := ""
		if params.FilterDimSelected != "" {
			filterDimSelectedEscaped = strconv.Quote(params.FilterDimSelected)
			filterDimSelectedEscaped = filterDimSelectedEscaped[1 : len(filterDimSelectedEscaped)-1]
		}

		dbIdentiferEscaped := ""
		if params.DbIdentifier != "" {
			dbIdentiferEscaped = strconv.Quote(params.DbIdentifier)
			dbIdentiferEscaped = dbIdentiferEscaped[1 : len(dbIdentiferEscaped)-1]
		}

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
			DbIdentifier:      dbIdentiferEscaped,
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

func info_handler(metrics_service metrics.Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query := `count(cc_pg_stat_activity) by (sys_id, sys_scope, sys_scope_fallback,sys_type)`
		now := time.Now().UnixMilli()
		options := map[string]string{
			"start": strconv.FormatInt(now-15*60*1000, 10), // 15 minutes ago
			"end":   strconv.FormatInt(now, 10),            // now
		}

		result, err := metrics_service.ExecuteRaw(query, options)
		if err != nil {
			http.Error(w, "Error querying list of instances.", http.StatusInternalServerError)
			return
		}

		instanceInfoMap := make(map[string]InstanceInfo)

		for _, sample := range result {
			sysID := getValue(sample, "sys_id")
			sysScope := getValue(sample, "sys_scope")
			sysType := getValue(sample, "sys_type")

			dbIdentifier := sysType + "/" + sysID + "/" + sysScope

			if _, exists := instanceInfoMap[dbIdentifier]; !exists {
				instanceInfoMap[dbIdentifier] = InstanceInfo{
					DBIdentifier: dbIdentifier,
					SystemID:     sysID,
					SystemScope:  sysScope,
					SystemType:   sysType,
				}
			}
		}

		var instanceInfos []InstanceInfo
		for _, info := range instanceInfoMap {
			instanceInfos = append(instanceInfos, info)
		}

		info_handler_internal(instanceInfos).ServeHTTP(w, r)
	})
}

func getValue(sample map[string]interface{}, key string) string {
	if mapValue, ok := sample["metric"].(map[string]interface{}); ok {
		if value, ok := mapValue[key].(string); ok && value != "" {
			return value
		}
	}
	return ""
}

func info_handler_internal(instances []InstanceInfo) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

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
var validClusterPrefixes = []string{
	"cluster-ro-",
	"cluster-",
	"", // Allow empty cluster prefix
}

// - A dbIdentifer is defined as <SystemType>/<SystemID>/<SystemScope>
// - SystemID: min 1 char max 63char
// - SystemScope:  <AWS_REGION>/<CLUSTER_PREFIX><AWS_ACCOUNT_ID> where AWS_REGION is between 1 and 50 characters, CLUSTER_PREFIX is between 0 and 11 characters, and AWS_ACCOUNT_ID is 12 characters. Then, the whole SystemScope is betwen 13 and 73 characters.
// - SystemType: One of the following values :
//   - amazon_rds
//   - google_cloudsql
//   - azure_database
//   - heroku
//   - crunchy_bridge
//   - aiven
//   - tembo
//   - self_hosted

func ValidateDbIdentifier(fl validator.FieldLevel) bool {
	dbIdentifier := strings.TrimSpace(fl.Field().String())

	// Remove leading and trailing parentheses if they exist
	if strings.HasPrefix(dbIdentifier, "(") && strings.HasSuffix(dbIdentifier, ")") {
		dbIdentifier = dbIdentifier[1 : len(dbIdentifier)-1]
	}

	// Split by pipe for multiple identifiers
	identifiers := strings.Split(dbIdentifier, "|")

	for _, id := range identifiers {
		id = strings.TrimSpace(id)

		parts := strings.SplitN(id, "/", 3)
		if len(parts) != 3 {
			return false
		}

		systemType := parts[0]
		systemID := parts[1]
		systemScope := parts[2]

		if !isValidSystemType(systemType) {
			return false
		}

		if len(systemID) < 1 || len(systemID) > SYSTEM_ID_MAX_LENGTH {
			return false
		}

		// Split the systemScope into region and optional clusterPrefixAccountID
		scopeParts := strings.SplitN(systemScope, "/", 2)
		systemRegion := scopeParts[0]
		clusterPrefixAccountID := ""
		if len(scopeParts) > 1 {
			clusterPrefixAccountID = scopeParts[1]
		}

		if len(systemRegion) < 1 || len(systemRegion) > AWS_REGION_MAX_LENGTH {
			return false
		}

		if clusterPrefixAccountID != "" {
			if len(clusterPrefixAccountID) < AWS_ACCOUNT_ID_LENGTH {
				return false
			}

			awsAccountID := clusterPrefixAccountID[len(clusterPrefixAccountID)-AWS_ACCOUNT_ID_LENGTH:]
			clusterPrefix := clusterPrefixAccountID[:len(clusterPrefixAccountID)-AWS_ACCOUNT_ID_LENGTH]

			if !isValidClusterPrefix(clusterPrefix) {
				return false
			}

			if len(awsAccountID) != AWS_ACCOUNT_ID_LENGTH {
				return false
			}

			if len(clusterPrefixAccountID) > (CLUSTER_PREFIX_MAX_LENGTH + AWS_ACCOUNT_ID_LENGTH) {
				return false
			}
		}
	}

	return true
}

func isValidSystemType(systemType string) bool {
	for _, validType := range validSystemTypes {
		if systemType == validType {
			return true
		}
	}
	return false
}

func isValidClusterPrefix(clusterPrefix string) bool {
	for _, validPrefix := range validClusterPrefixes {
		if clusterPrefix == validPrefix {
			return true
		}
	}
	return false
}
