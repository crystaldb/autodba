#!/bin/bash

# SPDX-Identifier: Apache-2.0

set -x -e

# Define paths
INSTALL_DIR="/usr/local/bin"
WEBAPP_DIR="/usr/local/share/autodba/webapp"
EXPORTER_DIR="/usr/local/share/prometheus_exporters"
CONFIG_DIR="/etc/prometheus"
AUTODBA_CONFIG_DIR="/etc/autodba"
PROMETHEUS_STORAGE_DIR="/prometheus"
PROMETHEUS_VERSION="2.42.0"

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

# Function to install Prometheus if not already installed
install_prometheus() {
    if command -v /usr/local/bin/prometheus >/dev/null 2>&1; then
        echo "Prometheus is already installed."
    else
        echo "Prometheus is not installed. Installing Prometheus..."
        wget -qO- https://github.com/prometheus/prometheus/releases/download/v${PROMETHEUS_VERSION}/prometheus-${PROMETHEUS_VERSION}.linux-${ARCH_SUFFIX}.tar.gz | tar -xzf - -C /tmp/
        sudo cp /tmp/prometheus-${PROMETHEUS_VERSION}.linux-${ARCH_SUFFIX}/prometheus /usr/local/bin/
        sudo cp /tmp/prometheus-${PROMETHEUS_VERSION}.linux-${ARCH_SUFFIX}/promtool /usr/local/bin/
        sudo mkdir -p /etc/prometheus /var/lib/prometheus
        sudo cp -r /tmp/prometheus-${PROMETHEUS_VERSION}.linux-${ARCH_SUFFIX}/consoles /etc/prometheus/
        sudo cp -r /tmp/prometheus-${PROMETHEUS_VERSION}.linux-${ARCH_SUFFIX}/console_libraries /etc/prometheus/
    fi
}

# Function to install from .tar.gz
install_tar_gz() {
    # Ensure required directories exist
    echo "Creating directories..."
    sudo mkdir -p "${INSTALL_DIR}" "${WEBAPP_DIR}" "${EXPORTER_DIR}" "${CONFIG_DIR}" "${AUTODBA_CONFIG_DIR}" "${PROMETHEUS_STORAGE_DIR}"
    sudo chown -R prometheus:prometheus "${PROMETHEUS_STORAGE_DIR}"
    echo "Extracting .tar.gz package $1..."
    tar -xzvf "$1" -C /tmp/
    sudo cp /tmp/autodba-*/bin/autodba-bff-${ARCH_SUFFIX} "${INSTALL_DIR}/autodba-bff"
    sudo cp -r /tmp/autodba-*/webapp/* "${WEBAPP_DIR}/"

    # Copy each exporter into its own directory
    sudo mkdir -p "${EXPORTER_DIR}/postgres_exporter"
    sudo mkdir -p "${EXPORTER_DIR}/sql_exporter"
    sudo mkdir -p "${EXPORTER_DIR}/rds_exporter"
    
    sudo cp -r /tmp/autodba-*/exporters/${ARCH_SUFFIX}/postgres_exporter/* "${EXPORTER_DIR}/postgres_exporter/"
    sudo cp -r /tmp/autodba-*/exporters/${ARCH_SUFFIX}/sql_exporter/* "${EXPORTER_DIR}/sql_exporter/"
    sudo cp -r /tmp/autodba-*/exporters/${ARCH_SUFFIX}/rds_exporter/* "${EXPORTER_DIR}/rds_exporter/"

    sudo cp /tmp/autodba-*/monitor/prometheus/prometheus.yml "${CONFIG_DIR}/prometheus.yml"
    sudo cp -r /tmp/autodba-*/monitor/prometheus/sql_exporter/* "${EXPORTER_DIR}/sql_exporter/"
    sudo cp -r /tmp/autodba-*/monitor/prometheus/rds_exporter/* "${EXPORTER_DIR}/rds_exporter/"
    sudo chown -R prometheus:prometheus ${CONFIG_DIR} /var/lib/prometheus
    sudo cp /tmp/autodba-*/config/config.json "${AUTODBA_CONFIG_DIR}/config.json"
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

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Check if package type is provided
if [ -z "$1" ]; then
    echo "Usage: $0 <package-file>"
    exit 1
fi

# Install Prometheus if not already installed
install_prometheus

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

# Set up the configuration
echo "Setting up configuration..."

# Check if required parameters are provided
if [[ -z "$AUTODBA_TARGET_DB" ]]; then
    echo "AUTODBA_TARGET_DB environment variable is not set"
    usage
fi

if [[ -z "$AWS_RDS_INSTANCE" ]]; then
    echo "Warning: AWS_RDS_INSTANCE environment variable is not set"
fi

if [[ -n "$AWS_RDS_INSTANCE" ]]; then
  if ! command_exists "aws"; then
    echo "Warning: AWS CLI is not installed. Please install AWS CLI to fetch AWS credentials."
  else
    # Fetch AWS Access Key and AWS Secret Key
    AWS_ACCESS_KEY_ID=$(aws configure get aws_access_key_id || echo "")
    AWS_SECRET_ACCESS_KEY=$(aws configure get aws_secret_access_key || echo "")
    AWS_REGION=$(aws configure get region || echo "")

    if [[ -z "$AWS_ACCESS_KEY_ID" || -z "$AWS_SECRET_ACCESS_KEY" || -z "$AWS_REGION" ]]; then
        echo "Warning: AWS credentials or region are not configured properly. Proceeding without AWS integration."
    fi
  fi
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
WorkingDirectory=${AUTODBA_CONFIG_DIR}
ExecStart=${INSTALL_DIR}/autodba-entrypoint.sh
Restart=on-failure
User=root
Environment="AUTODBA_TARGET_DB=${AUTODBA_TARGET_DB}"
Environment="AWS_RDS_INSTANCE=${AWS_RDS_INSTANCE}"
Environment="AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID}"
Environment="AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY}"
Environment="AWS_REGION=${AWS_REGION}"

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
    cd "${AUTODBA_CONFIG_DIR}"
    sudo AUTODBA_TARGET_DB=${AUTODBA_TARGET_DB} AWS_RDS_INSTANCE=${AWS_RDS_INSTANCE} AWS_ACCESS_KEY_ID=${AWS_ACCESS_KEY_ID} AWS_SECRET_ACCESS_KEY=${AWS_SECRET_ACCESS_KEY} AWS_REGION=${AWS_REGION} ${INSTALL_DIR}/autodba-entrypoint.sh
fi

echo "Installation complete."
