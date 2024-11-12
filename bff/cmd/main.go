package main

import (
	"flag"
	"fmt"
	"local/bff/pkg/metrics"
	"local/bff/pkg/prometheus"
	"local/bff/pkg/server"
	"os"

	"github.com/spf13/viper"
)

type Config struct {
	Port             string                        `json:"port"`
	PrometheusServer string                        `json:"prometheus_server"`
	RoutesConfig     map[string]server.RouteConfig `json:"routes_config"`
	WebappPath       string                        `json:"webapp_path"`
	AccessKey        string                        `json:"access_key"`
	DataPath         string                        `json:"data_path"`
}

func main() {
	fmt.Println("Starting metrics server")
	var webappPath string

	// TODO why is webappPath configured differently from other settings?
	flag.StringVar(&webappPath, "webappPath", "", "Webapp Path")
	flag.Parse()

	if webappPath == "" {
		fmt.Fprintf(os.Stderr, "Error: webappPath is required\n")
		os.Exit(1)
	}

	if err := run(webappPath); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(webappPath string) error {
	viper.SetConfigName("config")
	viper.SetConfigType("json")

	viper.AddConfigPath(".")
	err := viper.ReadInConfig()

	if err != nil {
		// TODO why panic rather than just return error?
		panic(fmt.Errorf("fatal error config file: %s", err))
	}

	rawConfig := make(map[string]interface{})
	if err := viper.Unmarshal(&rawConfig); err != nil {
		return fmt.Errorf("error unmarshaling config file: %s", err)
	}

	var config Config
	config.WebappPath = webappPath
	var ok bool

	// Get access key from environment variable with fallback to config file
	accessKey := os.Getenv("AUTODBA_ACCESS_KEY")
	if accessKey == "" {
		return fmt.Errorf("access key must be set via the AUTODBA_ACCESS_KEY environment variable")
	}
	config.AccessKey = accessKey

	// Get port from environment variable with fallback to config file
	port := os.Getenv("AUTODBA_BFF_PORT")
	if port == "" {
		port, ok = rawConfig["port"].(string)
		if !ok || port == "" {
			return fmt.Errorf("port must be set via the AUTODBA_BFF_PORT environment variable or in the config file")
		}
	}
	config.Port = port

	dataPath := os.Getenv("AUTODBA_DATA_PATH")
	if dataPath == "" {
		dataPath, ok = rawConfig["data_path"].(string)
		if !ok || dataPath == "" {
			return fmt.Errorf("data path must be set via the AUTODBA_DATA_PATH environment variable or in the config file")
		}
	}
	config.DataPath = dataPath

	prometheusURL := os.Getenv("PROMETHEUS_URL")
	if prometheusURL == "" {
		prometheusURL, ok = rawConfig["prometheus_server"].(string)
		if !ok || prometheusURL == "" {
			prometheusURL = "http://localhost:9090"
		}
	}
	config.PrometheusServer = prometheusURL

	if routes, ok := rawConfig["routes_config"].(map[string]interface{}); ok {
		config.RoutesConfig = make(map[string]server.RouteConfig)

		for key, value := range routes {

			var routeConfig server.RouteConfig

			routeMap, _ := value.(map[string]interface{})
			params := routeMap["params"].([]interface{})

			routeConfig.Params = make([]string, len(params))
			for i, param := range params {
				routeConfig.Params[i] = param.(string)
			}

			routeConfig.Options = make(map[string]string)
			options, _ := routeMap["options"].(map[string]interface{})
			for k, v := range options {
				routeConfig.Options[k] = v.(string)
			}

			routeConfig.Metrics = make(map[string]string)
			metrics, _ := routeMap["metrics"].(map[string]interface{})
			for k, v := range metrics {
				routeConfig.Metrics[k] = v.(string)
			}
			config.RoutesConfig[key] = routeConfig
		}
	}

	fmt.Printf("Config:\n%+v\n", config)

	metrics_repo := prometheus.New(config.PrometheusServer)
	metrics_service := metrics.CreateService(metrics_repo)

	server := server.CreateServer(config.RoutesConfig, metrics_service, config.Port, config.WebappPath, config.AccessKey, config.DataPath)

	if err = server.Run(); err != nil {
		return err
	}

	return nil
}
