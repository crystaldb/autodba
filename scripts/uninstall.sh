#!/bin/bash

# SPDX-Identifier: Apache-2.0

set -e

# Define paths
INSTALL_DIR="/usr/local/bin"
WEBAPP_DIR="/usr/local/share/autodba/webapp"
EXPORTER_DIR="/usr/local/share/prometheus_exporters"
CONFIG_DIR="/etc/autodba"
PROMETHEUS_DIR="/etc/prometheus"
PROMETHEUS_STORAGE_DIR="/prometheus"
SYSTEMD_SERVICE="/etc/systemd/system/autodba.service"

echo "Uninstalling AutoDBA..."

# Stop the service if systemd is used
if [ -f "$SYSTEMD_SERVICE" ]; then
    echo "Stopping and disabling AutoDBA service..."
    sudo systemctl stop autodba
    sudo systemctl disable autodba
    sudo rm -f "$SYSTEMD_SERVICE"
    sudo systemctl daemon-reload
fi

# Remove binaries and scripts
echo "Removing binaries and scripts..."
sudo rm -rf "${INSTALL_DIR}/autodba-bff"
sudo rm -rf "${INSTALL_DIR}/autodba-entrypoint.sh"

# Remove web application files
echo "Removing web application files..."
sudo rm -rf "${WEBAPP_DIR}"

# Remove Prometheus exporters and directories
echo "Removing Prometheus exporters and directories..."
sudo rm -rf "${EXPORTER_DIR}"

# Remove Prometheus configuration and storage
echo "Removing Prometheus configuration and storage..."
sudo rm -rf "${PROMETHEUS_DIR}"
sudo rm -rf "${PROMETHEUS_STORAGE_DIR}"

# Remove AutoDBA configuration files
echo "Removing AutoDBA configuration files..."
sudo rm -rf "${CONFIG_DIR}"

echo "AutoDBA has been successfully uninstalled."

