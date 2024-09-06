#!/bin/bash

# SPDX-Identifier: Apache-2.0

set -e -x

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Initialize variables
SYSTEM_INSTALL=false
USER_INSTALL_DIR=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --system)
            SYSTEM_INSTALL=true
            PARENT_DIR="/usr/local/autodba"
            shift
            ;;
        --install-dir)
            USER_INSTALL_DIR="$2"
            shift 2
            ;;
        *)
            echo "Unknown argument: $1"
            exit 1
            ;;
    esac
done

# Set the parent directory
if [ "$SYSTEM_INSTALL" = true ]; then
    PARENT_DIR="/usr/local/autodba"
elif [ -n "$USER_INSTALL_DIR" ]; then
    PARENT_DIR="$USER_INSTALL_DIR"
else
    PARENT_DIR="$HOME/autodba"
fi

# Define paths under the parent autodba directory
INSTALL_DIR="${PARENT_DIR}/bin"
WEBAPP_DIR="${PARENT_DIR}/share/webapp"
EXPORTER_DIR="${PARENT_DIR}/share/prometheus_exporters"
CONFIG_DIR="${PARENT_DIR}/config"
PROMETHEUS_CONFIG_DIR="${CONFIG_DIR}/prometheus"
PROMETHEUS_DIR="${PARENT_DIR}/prometheus"
PROMETHEUS_STORAGE_DIR="${PARENT_DIR}/prometheus_data"
SYSTEMD_SERVICE="/etc/systemd/system/autodba.service"

echo "Uninstalling AutoDBA from ${PARENT_DIR}..."

# Stop the service if systemd is used
if [ $SYSTEM_INSTALL && -f "$SYSTEMD_SERVICE"]; then
    echo "Stopping and disabling AutoDBA service..."
    sudo systemctl stop autodba
    sudo systemctl disable autodba
    sudo rm -f "$SYSTEMD_SERVICE"
    sudo systemctl daemon-reload
fi

# Remove binaries and scripts
echo "Removing binaries and scripts..."
rm -rf "${INSTALL_DIR}/bin"

# Remove web application files
echo "Removing web application files..."
rm -rf "${WEBAPP_DIR}"

# Remove Prometheus exporters and directories
echo "Removing Prometheus exporters and directories..."
rm -rf "${EXPORTER_DIR}"

# Remove Prometheus configuration and storage
echo "Removing Prometheus configuration and storage..."
rm -rf "${PROMETHEUS_DIR}"
rm -rf "${PROMETHEUS_CONFIG_DIR}"
rm -rf "${PROMETHEUS_STORAGE_DIR}"

# Remove AutoDBA configuration files
echo "Removing AutoDBA configuration files..."
rm -rf "${CONFIG_DIR}"

# Remove parent directory if empty
if [ -d "${PARENT_DIR}" ]; then
  echo "Removing parent AutoDBA directory..."
  rm -rf "${PARENT_DIR}"
fi

echo "Removing tmp directories..."
rm -rf /tmp/autodba-0.1.0
rm -rf /tmp/prometheus_rds_exporter
rm -rf /tmp/prometheus-2.42.0.linux-amd64

echo "AutoDBA has been successfully uninstalled."
