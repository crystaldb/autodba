package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"encoding/json"
	"flag"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/viper"
	"path/filepath"
)

type ContainerConfig struct {
	ProjectDir          string `json:"project_dir"`
	Dockerfile          string `json:"dockerfile"`
	ImageName           string `json:"image_name"`
	ContainerName       string `json:"container_name"`
	PrometheusPort      string `json:"prometheus_port"`
	BffPort             string `json:"bff_port"`
	DefaultMetricPeriod string `json:"default_metric_period"`
	WarmUpTime          string `json:"warm_up_time"`
}

type DbInfo struct {
	Description    string `json:"description"`
	DbConnString   string `json:"db_conn_string"`
	Host           string `json:"host"`
	Name           string `json:"name"`
	Username       string `json:"username"`
	Password       string `json:"password"`
	Port           string `json:"port"`
	AwsRdsInstance string `json:"aws_rds_instance"`
	AwsRegion      string `json:"aws_region"`
	AwsAccessKey   string `json:"aws_access_key"`
	AwsSecret      string `json:"aws_secret"`
}

type DbInfoMap map[string]DbInfo

var (
	cli         *client.Client
	containerID string
)

func readConfig() (*ContainerConfig, error) {
	viper.SetConfigName("container_config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config ContainerConfig
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

func generateCollectorConfig(containerConfig *ContainerConfig, dbInfo DbInfo) (string, error) {
	collectorConfig := fmt.Sprintf(`[server1]
db_host = %s
db_name = %s
db_username = %s
db_password = %s
db_port = %s
aws_db_instance_id = %s
aws_region = %s
aws_access_key_id = %s
aws_secret_access_key = %s
`, dbInfo.Host, dbInfo.Name, dbInfo.Username, dbInfo.Password, dbInfo.Port, dbInfo.AwsRdsInstance, dbInfo.AwsRegion, dbInfo.AwsAccessKey, dbInfo.AwsSecret)

	log.Println("GeneratedCollector config:\n ", collectorConfig)

	configFilePath := filepath.Join(containerConfig.ProjectDir, "autodba-collector.conf")

	absolutePath, err := filepath.Abs(configFilePath)
	if err != nil {
		return "", fmt.Errorf("error getting absolute path: %w", err)
	}

	err = os.WriteFile(absolutePath, []byte(collectorConfig), 0644)
	if err != nil {
		return "", fmt.Errorf("error writing collector config file: %w", err)
	}

	return absolutePath, nil
}

func SetupTestContainer(config *ContainerConfig, dbInfo DbInfo) error {
	ctx := context.Background()
	var err error

	log.Println("Generating collector config...")
	collectorConfigPath, err := generateCollectorConfig(config, dbInfo)
	if err != nil {
		return err
	}
	log.Println("Collector config path : ", collectorConfigPath)

	log.Println("Creating Docker client...")
	cli, err = client.NewClientWithOpts(client.WithVersion("1.41"))
	if err != nil {
		return err
	}

	log.Println("Building Docker image...")
	if err := buildDockerImage(ctx, config); err != nil {
		return err
	}

	log.Println("Stopping and removing existing container...")
	if err := stopAndRemoveContainer(ctx, config.ContainerName); err != nil {
		return err
	}

	log.Println("Preparing mounts and environment variables...")

	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: collectorConfigPath,
			Target: "/usr/local/autodba/share/collector/collector.conf",
		},
	}

	envVars := []string{
		"DB_CONN_STRING=" + dbInfo.DbConnString,
		"AWS_RDS_INSTANCE=" + dbInfo.AwsRdsInstance,
		"AWS_ACCESS_KEY_ID=" + dbInfo.AwsAccessKey,
		"AWS_SECRET_ACCESS_KEY=" + dbInfo.AwsSecret,
		"AWS_REGION=" + dbInfo.AwsRegion,
		"CONFIG_FILE=" + collectorConfigPath,
		"DEFAULT_METRIC_COLLECTION_PERIOD_SECONDS=" + config.DefaultMetricPeriod,
		"WARM_UP_TIME_SECONDS=" + config.WarmUpTime,
	}

	log.Println("Environment variables:")
	for _, envVar := range envVars {
		log.Println(envVar)
	}

	log.Println("Creating and starting the container...")
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: config.ImageName,
		ExposedPorts: map[nat.Port]struct{}{
			"9090/tcp": {},
			"4000/tcp": {},
		},
		Env: envVars,
	}, &container.HostConfig{
		PortBindings: map[nat.Port][]nat.PortBinding{
			"9090/tcp": {{HostPort: config.PrometheusPort}},
			"4000/tcp": {{HostPort: config.BffPort}},
		},
		Mounts: mounts,
	}, nil, nil, config.ContainerName)
	if err != nil {
		return err
	}

	containerID = resp.ID
	log.Printf("Container created with ID: %s\n", containerID)

	if err := cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return err
	}

	log.Println("Waiting for the container to be ready...")

	const maxWaitTime = 5 * time.Minute

	timeout := time.After(maxWaitTime)
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-timeout:
			return fmt.Errorf("timeout waiting for container to be running")
		case <-ticker.C:
			containerJSON, err := cli.ContainerInspect(ctx, containerID)
			if err != nil {
				return err
			}
			log.Printf("Current container status: %s\n", containerJSON.State.Status)
			if containerJSON.State.Status == "running" {
				log.Println("Container is running.")
				log.Println("Container setup completed.")
				return nil
			}
			log.Printf("Current container status: %s. Waiting...\n", containerJSON.State.Status)
		}
	}
}

func buildDockerImage(ctx context.Context, config *ContainerConfig) error {
	log.Println("Preparing build context from Dockerfile...")
	tarball, err := archive.TarWithOptions(config.ProjectDir, &archive.TarOptions{})
	if err != nil {
		return err
	}
	defer tarball.Close()

	log.Println("Building the Docker image...")
	resp, err := cli.ImageBuild(ctx, tarball, types.ImageBuildOptions{
		Tags: []string{config.ImageName},
	})
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	log.Println("Reading build output...")
	if _, err := io.Copy(os.Stdout, resp.Body); err != nil {
		return err
	}

	log.Println("Docker image built successfully.")
	return nil
}

func stopAndRemoveContainer(ctx context.Context, containerName string) error {
	log.Println("Stopping the existing container...")
	if err := cli.ContainerStop(ctx, containerName, container.StopOptions{}); err != nil && !client.IsErrNotFound(err) {
		return err
	}
	log.Println("Removing the existing container...")
	if err := cli.ContainerRemove(ctx, containerName, container.RemoveOptions{}); err != nil && !client.IsErrNotFound(err) {
		return err
	}
	log.Println("Existing container stopped and removed successfully.")
	return nil
}

func TearDownTestContainer() error {
	ctx := context.Background()

	log.Println("Stopping the test container...")
	if err := cli.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
		return err
	}

	log.Println("Removing the test container...")
	if err := cli.ContainerRemove(ctx, containerID, container.RemoveOptions{}); err != nil {
		return err
	}

	log.Println("Test container torn down successfully.")
	return nil
}

func parseCLIArgs() (ContainerConfig, DbInfoMap, error) {
	dbConfigStr := flag.String("dbconfig", "", "JSON string of database configuration map")
	containerConfigStr := flag.String("containerConfig", "", "JSON string of container configuration")
	flag.Parse()

	var config ContainerConfig
	if err := json.Unmarshal([]byte(*containerConfigStr), &config); err != nil {
		return config, DbInfoMap{}, err
	}

	var dbInfoMap DbInfoMap
	if err := json.Unmarshal([]byte(*dbConfigStr), &dbInfoMap); err != nil {
		return config, DbInfoMap{}, err
	}

	return config, dbInfoMap, nil
}

// you can run main to spin up a container without running the test suite
func main() {
	config, dbInfoMap, err := parseCLIArgs()
	if err != nil {
		log.Fatalf("Error parsing CLI args: %v\n", err)
	}

	dbInfo, ok := dbInfoMap["16.3"] // Change "16.3" to any other version if needed
	if !ok {
		log.Fatalf("No DbInfo found for version '16'")
	}

	log.Printf("Container config: %+v\n", config)
	log.Printf("DbInfo :  %+v\n", dbInfo)
	log.Println("RDS Instance: ", dbInfo.AwsRdsInstance)
	log.Println("DB Connection String: ", dbInfo.DbConnString)

	log.Println("Setting up test container...")
	if err := SetupTestContainer(&config, dbInfo); err != nil {
		log.Fatalf("Failed to set up container: %v\n", err)
	}

	// defer func() {
	// 	log.Println("Tearing down test container...")
	// 	if err := TearDownTestContainer(); err != nil {
	// 		log.Printf("Failed to tear down container: %v\n", err)
	// 	}
	// }()
}
