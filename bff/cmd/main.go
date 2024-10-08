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
}

func main() {
	fmt.Println("Starting metrics server")
	var collectorConfigFile string
	var webappPath string

	flag.StringVar(&collectorConfigFile, "collectorConfigFile", "", "Database identifier")
	flag.StringVar(&webappPath, "webappPath", "", "Webapp Path")
	flag.Parse()

	if collectorConfigFile == "" {
		fmt.Fprintf(os.Stderr, "Error: collectorConfigFile is required\n")
		os.Exit(1)
	}

	if webappPath == "" {
		fmt.Fprintf(os.Stderr, "Error: webappPath is required\n")
		os.Exit(1)
	}

	if err := run(collectorConfigFile, webappPath); err != nil {
		fmt.Fprintf(os.Stderr, "%s\n", err)
		os.Exit(1)
	}
}

func run(collectorConfigFile, webappPath string) error {
	viper.SetConfigName("config")
	viper.SetConfigType("json")

	viper.AddConfigPath(".")
	err := viper.ReadInConfig()

	if err != nil {
		panic(fmt.Errorf("Fatal error config file: %s \n", err))
	}

	rawConfig := make(map[string]interface{})
	if err := viper.Unmarshal(&rawConfig); err != nil {
		return fmt.Errorf("Error unmarshaling config file: %s", err)
	}

	var config Config
	config.WebappPath = webappPath
	config.Port = rawConfig["port"].(string)
	config.PrometheusServer = rawConfig["prometheus_server"].(string)

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

	server := server.CreateServer(config.RoutesConfig, metrics_service, config.Port, config.WebappPath)

	if err = server.Run(); err != nil {
		return err
	}

	return nil
}
