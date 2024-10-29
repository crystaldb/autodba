#!/bin/bash

# SPDX-Identifier: Apache-2.0

# Set the base directory based on installation
PARENT_DIR="${PARENT_DIR:-/usr/local/autodba}"
PROMETHEUS_HOST="${PROMETHEUS_HOST:-localhost:9090}"

# Wait for Prometheus to be ready
echo "collector-api: Waiting for Prometheus to be ready..."
until curl -s ${PROMETHEUS_HOST}/-/ready > /dev/null; do
    sleep 1
done
echo "collector-api: Prometheus is ready."

cd $PARENT_DIR/share/collector_api_server

# Start the Collector API Server
exec ./collector-api-server
