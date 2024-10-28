#!/bin/bash

# SPDX-Identifier: Apache-2.0

set -e

# Initialize variables
SYSTEM_INSTALL=false
USER_INSTALL_DIR=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --system)
            SYSTEM_INSTALL=true
            PARENT_DIR="/usr/local/autodba-collector"
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
if [ -n "$USER_INSTALL_DIR" ]; then
    PARENT_DIR="$USER_INSTALL_DIR"
elif [ "$SYSTEM_INSTALL" = true ]; then
    PARENT_DIR="/usr/local/autodba-collector"
else
    PARENT_DIR="$(pwd)"
fi

# Define paths
INSTALL_DIR="${PARENT_DIR}/bin"
CONFIG_DIR="${PARENT_DIR}/config"
SYSTEMD_SERVICE="/etc/systemd/system/autodba-collector.service"

echo "Uninstalling AutoDBA Collector from ${PARENT_DIR}..."

# Stop the service if systemd is used
if [ "$SYSTEM_INSTALL" = true ] && [ -f "$SYSTEMD_SERVICE" ]; then
    echo "Stopping and disabling AutoDBA Collector service..."
    sudo systemctl stop autodba-collector
    sudo systemctl disable autodba-collector
    sudo rm -f "$SYSTEMD_SERVICE"
    sudo systemctl daemon-reload
fi

if [ "$PARENT_DIR" != "$(pwd)" ]; then
    # Remove binaries and scripts
    echo "Removing binaries and scripts..."
    [ -d "${INSTALL_DIR}" ] && rm -rf "${INSTALL_DIR}" || true

    # Remove configuration files
    echo "Removing configuration files..."
    [ -d "${CONFIG_DIR}" ] && rm -rf "${CONFIG_DIR}" || true

    # Remove parent directory if empty
    if [ -d "${PARENT_DIR}" ]; then
        echo "Removing parent AutoDBA Collector directory..."
        rmdir --ignore-fail-on-non-empty "${PARENT_DIR}" || true
    fi

    # Clean up system user if it exists
    if [ "$SYSTEM_INSTALL" = true ] && id -u autodba-collector >/dev/null 2>&1; then
        echo "Removing autodba-collector user..."
        userdel autodba-collector
    fi
else
    echo "Not removing the current directory as it is the installation directory."
fi

echo "AutoDBA Collector has been successfully uninstalled."
