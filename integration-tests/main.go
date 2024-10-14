package main

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/archive"
	"github.com/docker/go-connections/nat"
	"github.com/spf13/viper"
)

type Config struct {
	ProjectDir          string `mapstructure:"project_dir"`
	Dockerfile          string `mapstructure:"dockerfile"`
	ImageName           string `mapstructure:"image_name"`
	ContainerName       string `mapstructure:"container_name"`
	PrometheusPort      string `mapstructure:"prometheus_port"`
	BffPort             string `mapstructure:"bff_port"`
	DbConnString        string `mapstructure:"db_conn_string"`
	AwsRdsInstance      string `mapstructure:"aws_rds_instance"`
	DefaultMetricPeriod string `mapstructure:"default_metric_period"`
	WarmUpTime          string `mapstructure:"warm_up_time"`
	ConfigFile          string `mapstructure:"config_file"` // NOTE this path must be absolute
	BackupDir           string `mapstructure:"backup_dir"`
	BackupFile          string `mapstructure:"backup_file"`
}

type DbInfo struct {
	Description    string `json:"description"`
	DbConnString   string `json:"db_conn_string"`
	AwsRdsInstance string `json:"aws_rds_instance"`
}

var (
	cli         *client.Client
	containerID string
)

func readConfig() (*Config, error) {
	viper.SetConfigName("container_config")
	viper.SetConfigType("json")
	viper.AddConfigPath(".")

	if err := viper.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	var config Config
	if err := viper.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	return &config, nil
}

func SetupTestContainer(config *Config, dbInfo DbInfo) error {
	ctx := context.Background()

	log.Println("Creating Docker client...")
	var err error
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
			Source: config.ConfigFile,
			Target: "/usr/local/autodba/share/collector/collector.conf",
		},
	}

	envVars := []string{
		"DEFAULT_METRIC_COLLECTION_PERIOD_SECONDS=" + config.DefaultMetricPeriod,
		"WARM_UP_TIME_SECONDS=" + config.WarmUpTime,
		"DB_CONN_STRING=" + dbInfo.DbConnString,
		"AWS_RDS_INSTANCE=" + dbInfo.AwsRdsInstance,
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

func buildDockerImage(ctx context.Context, config *Config) error {
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

func main() {
	config, err := readConfig()
	if err != nil {
		log.Fatalf("Failed to read config: %v\n", err)
	}

	var testCase = DbInfo{
		Description:    "Version 13",
		DbConnString:   "postgres://postgres:rme49DKjpE4wwx16Bemu@radcliffe-1.c7mrowi2kiu4.us-east-1.rds.amazonaws.com:5432/postgres?sslmode=require",
		AwsRdsInstance: "radcliffe-1",
	}

	log.Println("Setting up test container...")
	if err := SetupTestContainer(config, testCase); err != nil {
		log.Fatalf("Failed to set up container: %v\n", err)
	}

	// defer func() {
	// 	log.Println("Tearing down test container...")
	// 	if err := TearDownTestContainer(); err != nil {
	// 		log.Printf("Failed to tear down container: %v\n", err)
	// 	}
	// }()
}
