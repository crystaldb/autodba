#!/bin/bash

# SPDX-Identifier: Apache-2.0

# Set the base directory based on installation
PARENT_DIR="${PARENT_DIR:-/usr/local/autodba}"
CONFIG_FILE="${CONFIG_FILE:-${PARENT_DIR}/share/collector/collector.conf}"
WEBAPP_PATH="${PARENT_DIR}/share/webapp"

# Wait for Prometheus to be ready
echo "Waiting for Prometheus to be ready..."
until curl -s http://prometheus:9090/-/ready > /dev/null; do
    sleep 1
done
echo "Prometheus is ready."

exec ${PARENT_DIR}/bin/autodba-bff -collectorConfigFile="${CONFIG_FILE}" -webappPath "${WEBAPP_PATH}"
