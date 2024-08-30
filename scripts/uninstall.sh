#!/bin/bash

# SPDX-License-Identifier: Apache-2.0

set -e

# Define paths
INSTALL_DIR="/usr/local/bin"
WEBAPP_DIR="/usr/local/share/autodba/webapp"
PROM_DIR="/usr/local/share/prometheus"
CONFIG_DIR="/etc/autodba"
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

# Remove Prometheus and exporters
echo "Removing Prometheus and exporters..."
sudo rm -rf "${PROM_DIR}"

# Remove configuration files
echo "Removing configuration files..."
sudo rm -rf "${CONFIG_DIR}"

echo "AutoDBA has been successfully uninstalled."
