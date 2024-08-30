#!/bin/bash

# SPDX-License-Identifier: Apache-2.0

set -e

# Define paths
INSTALL_DIR="/usr/local/bin"
WEBAPP_DIR="/usr/local/share/autodba/webapp"
PROM_DIR="/usr/local/share/prometheus"
CONFIG_DIR="/etc/autodba"

# Detect system architecture
ARCH=$(uname -m)
case "$ARCH" in
    x86_64)
        ARCH_SUFFIX="amd64"
        ;;
    aarch64)
        ARCH_SUFFIX="arm64"
        ;;
    *)
        echo "Unsupported architecture: $ARCH"
        exit 1
        ;;
esac

echo "Detected architecture: $ARCH_SUFFIX"

# Function to install from .tar.gz
install_tar_gz() {

    # Ensure required directories exist
    echo "Creating directories..."
    sudo mkdir -p "${INSTALL_DIR}" "${WEBAPP_DIR}" "${PROM_DIR}" "${CONFIG_DIR}"
    echo "Extracting .tar.gz package $1..."
    tar -xzvf "$1" -C /tmp/
    sudo cp /tmp/autodba-*/bin/autodba-bff-${ARCH_SUFFIX} "${INSTALL_DIR}/autodba-bff"
    sudo cp -r /tmp/autodba-*/webapp/* "${WEBAPP_DIR}/"
    sudo cp -r /tmp/autodba-*/prometheus/${ARCH_SUFFIX}/* "${PROM_DIR}/"
    sudo cp /tmp/autodba-*/entrypoint.sh "${INSTALL_DIR}/autodba-entrypoint.sh"
    sudo chmod +x "${INSTALL_DIR}/autodba-entrypoint.sh"
}

# Function to install from .deb
install_deb() {
    echo "Installing .deb package $1..."
    sudo dpkg -i "$1"
}

# Function to install from .rpm
install_rpm() {
    echo "Installing .rpm package $1..."
    sudo rpm -i "$1"
}

# Check if package type is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <package-file>"
    exit 1
fi

# Determine package type and install accordingly
case "$1" in
    *.tar.gz)
        install_tar_gz "$1"
        ;;
    *.deb)
        install_deb "$1"
        ;;
    *.rpm)
        install_rpm "$1"
        ;;
    *)
        echo "Unsupported package format: $1"
        exit 1
        ;;
esac

# Set up the configuration (customize as needed)
echo "Setting up configuration..."
# Here you can add commands to set up your configuration files in CONFIG_DIR

# Optionally customize Prometheus configuration
if [ ! -f "${PROM_DIR}/prometheus.yml" ]; then
    echo "No custom Prometheus configuration found, setting up a default one..."
    cp /tmp/autodba-*/prometheus/prometheus.yml "${PROM_DIR}/prometheus.yml"
fi

# Install systemd service (optional)
if [ "$(which systemctl)" ]; then
    echo "Installing systemd service..."
    cat << EOF | sudo tee /etc/systemd/system/autodba.service
[Unit]
Description=AutoDBA Service
After=network.target

[Service]
Type=simple
ExecStart=${INSTALL_DIR}/autodba-entrypoint.sh
Restart=on-failure
User=root

[Install]
WantedBy=multi-user.target
EOF

    sudo systemctl daemon-reload
    sudo systemctl enable autodba
    sudo systemctl start autodba
fi

# Run the application (if not using systemd)
if [ ! "$(which systemctl)" ]; then
    echo "Starting AutoDBA..."
    "${INSTALL_DIR}/autodba-entrypoint.sh"
fi

echo "Installation complete."
