package server

import (
	"encoding/json"
	"fmt"
	"local/bff/pkg/metrics"
	"net/http"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
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
	dbIdentifier    string
	webappPath      string
}

const api_prefix = "/api"
const replace_prefix = "$"

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

func CreateServer(r map[string]RouteConfig, m metrics.Service, port string, dbIdentifier string, webappPath string) Server {
	return server_imp{r, m, port, dbIdentifier, webappPath}
}

func fileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false
		}
		fmt.Printf("Error checking file existence: %v\n", err)
		return false
	}
	return true
}

func (s server_imp) Run() error {
	r := chi.NewRouter()

	r.Use(CORS)

	r.Get("/api/v1/activity", activity_handler(s.metrics_service))
	r.Get("/api/v1/databases", databases_handler(s.metrics_service))
	r.Get("/api/v1/info", info_handler(s.metrics_service))

	r.Route(api_prefix, func(r chi.Router) {
		r.Mount("/", metrics_handler(s.routes_config, s.dbIdentifier, s.metrics_service))
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

func metrics_handler(route_configs map[string]RouteConfig, dbIdentifier string, metrics_service metrics.Service) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Handling metrics request")

		route := strings.TrimPrefix(r.URL.Path, api_prefix)
		fmt.Println("ROUTE: ", route)

		options := make(map[string]string)
		metrics := make(map[string]string)

		if route_config, ok := route_configs[route]; ok {
			fmt.Println("matching route found: extracting params")
			for _, param := range route_config.Params {
				value := r.URL.Query().Get(param)

				if param == "dbidentifier" && value == "" {
					for metric, query := range route_config.Metrics {
						var current_query string
						if metrics[metric] == "" {
							current_query = query
						} else {
							current_query = metrics[metric]
						}

						metrics[metric] = strings.ReplaceAll(current_query, replace_prefix+param, dbIdentifier)

						fmt.Println("metric: ", metric)
						fmt.Println("query: ", query)
						fmt.Println("populated metric: ", metrics[metric])
					}
				} else if value != "" {
					fmt.Println("-----param: ", param)
					fmt.Println("value: ", value)
					fmt.Println("==Populating Queries for : ", param)
					for metric, query := range route_config.Metrics {
						var current_query string
						if metrics[metric] == "" {
							current_query = query
						} else {
							current_query = metrics[metric]
						}

						metrics[metric] = strings.ReplaceAll(current_query, replace_prefix+param, value)

						fmt.Println("metric: ", metric)
						fmt.Println("query: ", query)
						fmt.Println("populated metric: ", metrics[metric])
					}

					fmt.Println("==Populating Options for : ", param)
					for option, input := range route_config.Options {
						var current_input string
						if options[option] == "" {
							current_input = input
						} else {
							current_input = options[option]
						}

						options[option] = strings.ReplaceAll(current_input, replace_prefix+param, value)

						fmt.Println("option: ", option)
						fmt.Println("input: ", input)
						fmt.Println("populated options: ", options[option])
					}
				} else {
					fmt.Println("param: ", param, " not found in request")
					fmt.Println("Route: ", route)
					fmt.Println("Params: ", r.URL.Query())
				}

				fmt.Println("_____________________________________")
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

			fmt.Println("metrics: ", metrics)
			fmt.Println("options: ", options)

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

func activity_handler(metrics_service metrics.Service) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Handling activity request")

		database_list := r.URL.Query().Get("database_list")
		start := r.URL.Query().Get("start")
		end := r.URL.Query().Get("end")
		step := r.URL.Query().Get("step")
		legend := r.URL.Query().Get("legend")
		dim := r.URL.Query().Get("dim")
		filterdim := r.URL.Query().Get("filterdim")
		filterdimselected := r.URL.Query().Get("filterdimselected")
		limit := r.URL.Query().Get("limit")

		requiredParamMap := map[string]string{
			"database_list": database_list,
			"start":         start,
			"end":           end,
			"step":          step,
			"legend":        legend,
			"dim":           dim,
		}

		for paramName, paramValue := range requiredParamMap {
			if paramValue == "" {
				http.Error(w, fmt.Sprintf("Missing param/value: %s", paramName), http.StatusBadRequest)
				return
			}
		}
		parsedStart, err := parseAndValidateInt(start, "start")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		parsedEnd, err := parseAndValidateInt(end, "end")
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}

		if limit != "" {
			_, err = parseAndValidateInt(limit, "limit")
			if err != nil {
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}
		}

		if parsedEnd <= parsedStart {
			http.Error(w, "Parameter 'end' must be greater than 'start'", http.StatusBadRequest)
			return
		}

		promQLInput := PromQLInput{
			DatabaseList:      database_list,
			Start:             start,
			End:               end,
			Step:              step,
			Legend:            legend,
			Dim:               dim,
			FilterDim:         filterdim,
			FilterDimSelected: filterdimselected,
			Limit:             limit,
			Offset:            "", // TODO not in query
		}

		fmt.Println("PromQLInput: ", promQLInput)

		query, err := GenerateActivityCubePromQLQuery(promQLInput)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		fmt.Println("query: ", query)

		options := map[string]string{
			"start": start,
			"end":   end,
			"step":  step,
			"dim":   dim,
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

		currentTime := time.Now().UnixNano() / int64(time.Millisecond)
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
		fmt.Println("Handling database list request")

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
		fmt.Println("Handling info request")

		query := "rds_instance_info"
		options := make(map[string]string)

		results, err := metrics_service.ExecuteRaw(query, options)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		info := make(map[string]string)

		for _, result := range results {
			metric, ok := result["metric"].(map[string]interface{})
			if !ok {
				continue
			}

			for k, v := range metric {
				info[k] = v.(string)
			}
		}

		js, err := json.Marshal(info)
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
