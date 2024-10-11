package main

import (
	"context"
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
)

var (
	cli        *client.Client
	containerID string
)

const (
	PROJECT_DIR          = "../"
	DOCKERFILE          = "../Dockerfile"
	IMAGE_NAME          = "autodba:latest"
	CONTAINER_NAME      = "autodba_test"
	PROMETHEUS_PORT     = "9090"
	BFF_PORT            = "4000"
	DB_CONN_STRING      = "postgres://postgres:lTEP7OzeXQr77Ldu@mohammad-dashti-rds-1.cvirkksghnig.us-west-2.rds.amazonaws.com:5432/postgres?sslmode=require"
	AWS_RDS_INSTANCE    = "mohammad-dashti-rds-1"
	DEFAULT_METRIC_PERIOD = "5"
	WARM_UP_TIME        = "60"
	CONFIG_FILE         = "/home/steams/Development/autodba/autodba-collector.conf"
	BACKUP_DIR          = "/home/autodba/ext-backups"
	BACKUP_FILE         = ""
    // -e AWS_ACCESS_KEY_ID="" \
    // -e AWS_SECRET_ACCESS_KEY="" \
    // -e AWS_REGION="" \
    // -e DISABLE_DATA_COLLECTION="false" \
)

func SetupTestContainer() error {
	ctx := context.Background()

	log.Println("Creating Docker client...")
	var err error
	cli, err = client.NewClientWithOpts(client.WithVersion("1.41"))
	if err != nil {
		return err
	}

	log.Println("Building Docker image...")
	if err := buildDockerImage(ctx); err != nil {
		return err
	}

	log.Println("Stopping and removing existing container...")
	if err := stopAndRemoveContainer(ctx); err != nil {
		return err
	}

	log.Println("Preparing mounts and environment variables...")
	mounts := []mount.Mount{
		{
			Type:   mount.TypeBind,
			Source: CONFIG_FILE, // Path to the file on your host
			Target: "/usr/local/autodba/share/collector/collector.conf", // Path inside the container
		},
}

	envVars := []string{
		"DB_CONN_STRING=" + DB_CONN_STRING,
		"AWS_RDS_INSTANCE=" + AWS_RDS_INSTANCE,
		"DEFAULT_METRIC_COLLECTION_PERIOD_SECONDS=" + DEFAULT_METRIC_PERIOD,
		"WARM_UP_TIME_SECONDS=" + WARM_UP_TIME,
	}

	log.Println("Creating and starting the container...")
	resp, err := cli.ContainerCreate(ctx, &container.Config{
		Image: IMAGE_NAME,
		ExposedPorts: map[nat.Port]struct{}{
			"9090/tcp": {},
			"4000/tcp": {},
		},
		Env: envVars,
	}, &container.HostConfig{
		PortBindings: map[nat.Port][]nat.PortBinding{
			"9090/tcp": {{HostPort: PROMETHEUS_PORT}},
			"4000/tcp": {{HostPort: BFF_PORT}},
		},
		Mounts: mounts,
	}, nil, nil, CONTAINER_NAME)
	if err != nil {
		return err
	}

	containerID = resp.ID
	log.Printf("Container created with ID: %s\n", containerID)

	if err := cli.ContainerStart(ctx, containerID, container.StartOptions{}); err != nil {
		return err
	}

	log.Println("Waiting for the container to be ready...")
	time.Sleep(10 * time.Second)

	// Check the container status
	containerJSON, err := cli.ContainerInspect(ctx, containerID)
	if err != nil {
		return err
	}
	if containerJSON.State.Status != "running" {
		log.Printf("Container is not running. Status: %s\n", containerJSON.State.Status)
		return nil
	}

	log.Println("Container setup completed.")
	return nil
}

func stopAndRemoveContainer(ctx context.Context) error {
	log.Println("Stopping the existing container...")
	if err := cli.ContainerStop(ctx, CONTAINER_NAME, container.StopOptions{}); err != nil && !client.IsErrNotFound(err) {
		return err
	}
	log.Println("Removing the existing container...")
	if err := cli.ContainerRemove(ctx, CONTAINER_NAME, container.RemoveOptions{}); err != nil && !client.IsErrNotFound(err) {
		return err
	}
	log.Println("Existing container stopped and removed successfully.")
	return nil
}

func buildDockerImage(ctx context.Context) error {
	log.Println("Preparing build context from Dockerfile...")
	tarball, err := archive.TarWithOptions(PROJECT_DIR, &archive.TarOptions{})
	if err != nil {
		return err
	}
	defer tarball.Close()

	log.Println("Building the Docker image...")
	resp, err := cli.ImageBuild(ctx, tarball, types.ImageBuildOptions{
		Tags: []string{IMAGE_NAME},
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
	log.Println("Setting up test container...")
	if err := SetupTestContainer(); err != nil {
		log.Fatalf("Failed to set up container: %v\n", err)
	}

	// defer func() {
	// 	log.Println("Tearing down test container...")
	// 	if err := TearDownTestContainer(); err != nil {
	// 		log.Printf("Failed to tear down container: %v\n", err)
	// 	}
	// }()
}

