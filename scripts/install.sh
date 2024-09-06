#!/bin/bash

# SPDX-Identifier: Apache-2.0

set -e

# Initialize variables
SYSTEM_INSTALL=false
USER_INSTALL_DIR=""
PACKAGE_FILE=""
CONFIG_FILE=""
GENERATE_PACKAGE=false

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
        --package)
            PACKAGE_FILE="$2"
            shift 2
            ;;
        --config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        *)
            echo "Unknown argument: $1"
            exit 1
            ;;
    esac
done

usage() {
    echo "Usage: $0 --package <path/to/package> [--system] [--install-dir <path>] [--config <config.json>]"
    echo ""
    echo "Options:"
    echo "  --package      Path to the package file (.tar.gz, .deb, or .rpm) [REQUIRED]"
    echo "  --system       Install globally under /usr/local/autodba"
    echo "  --install-dir  Specify a custom installation directory. If not specified, $HOME/autodba is used."
    echo "  --config       Path to the AutoDBA config file (optional), or use standard input if not provided."
    exit 1
}

# Ensure that the --package argument is provided
if [ -z "$PACKAGE_FILE" ]; then
    echo "Error: --package argument is required."
    usage
fi

# Set the parent directory
if [ "$SYSTEM_INSTALL" = true ]; then
    PARENT_DIR="/usr/local/autodba"
elif [ -n "$USER_INSTALL_DIR" ]; then
    PARENT_DIR="$USER_INSTALL_DIR"
else
    PARENT_DIR="$HOME/autodba"
fi

# Define paths relative to PARENT_DIR
INSTALL_DIR="$PARENT_DIR/bin"
WEBAPP_DIR="$PARENT_DIR/share/webapp"
EXPORTER_DIR="$PARENT_DIR/share/prometheus_exporters"
PROMETHEUS_CONFIG_DIR="$PARENT_DIR/config/prometheus"
AUTODBA_CONFIG_DIR="$PARENT_DIR/config/autodba"
PROMETHEUS_STORAGE_DIR="$PARENT_DIR/prometheus_data"
PROMETHEUS_INSTALL_DIR="$PARENT_DIR/prometheus"
PROMETHEUS_VERSION="2.42.0"
AUTODBA_CONFIG_FILE="$AUTODBA_CONFIG_DIR/autodba-config.json"

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
echo "Installing AutoDBA under: $PARENT_DIR"

# Function to install Prometheus in a custom path
install_prometheus() {
    if [ -f "$PROMETHEUS_INSTALL_DIR/prometheus" ]; then
        echo "Prometheus is already installed."
    else
        echo "Installing Prometheus in custom path: $PROMETHEUS_INSTALL_DIR..."
        wget -qO- https://github.com/prometheus/prometheus/releases/download/v${PROMETHEUS_VERSION}/prometheus-${PROMETHEUS_VERSION}.linux-${ARCH_SUFFIX}.tar.gz | tar -xzf - -C /tmp/
        mkdir -p "$PROMETHEUS_INSTALL_DIR"
        cp /tmp/prometheus-${PROMETHEUS_VERSION}.linux-${ARCH_SUFFIX}/prometheus "$PROMETHEUS_INSTALL_DIR/"
        cp /tmp/prometheus-${PROMETHEUS_VERSION}.linux-${ARCH_SUFFIX}/promtool "$PROMETHEUS_INSTALL_DIR/"
        mkdir -p ${PROMETHEUS_CONFIG_DIR}
        cp -r /tmp/prometheus-${PROMETHEUS_VERSION}.linux-${ARCH_SUFFIX}/consoles ${PROMETHEUS_CONFIG_DIR}/
        cp -r /tmp/prometheus-${PROMETHEUS_VERSION}.linux-${ARCH_SUFFIX}/console_libraries ${PROMETHEUS_CONFIG_DIR}/
    fi
}

# Function to install from .tar.gz
install_tar_gz() {
    # Ensure required directories exist
    echo "Creating directories..."
    mkdir -p "${INSTALL_DIR}" "${WEBAPP_DIR}" "${EXPORTER_DIR}" "${PROMETHEUS_CONFIG_DIR}" "${AUTODBA_CONFIG_DIR}" "${PROMETHEUS_STORAGE_DIR}"
    echo "Extracting .tar.gz package $PACKAGE_FILE..."
    tar -xzvf "$PACKAGE_FILE" -C /tmp/
    
    cp /tmp/autodba-*/bin/autodba-bff-${ARCH_SUFFIX} "${INSTALL_DIR}/autodba-bff"
    cp -r /tmp/autodba-*/webapp/* "${WEBAPP_DIR}/"

    # Copy each exporter into its own directory
    mkdir -p "${EXPORTER_DIR}/postgres_exporter"
    mkdir -p "${EXPORTER_DIR}/sql_exporter"
    mkdir -p "${EXPORTER_DIR}/rds_exporter"
    
    cp -r /tmp/autodba-*/exporters/${ARCH_SUFFIX}/postgres_exporter/* "${EXPORTER_DIR}/postgres_exporter/"
    cp -r /tmp/autodba-*/exporters/${ARCH_SUFFIX}/sql_exporter/* "${EXPORTER_DIR}/sql_exporter/"
    cp -r /tmp/autodba-*/exporters/${ARCH_SUFFIX}/rds_exporter/* "${EXPORTER_DIR}/rds_exporter/"

    cp /tmp/autodba-*/monitor/prometheus/prometheus.yml "${PROMETHEUS_CONFIG_DIR}/prometheus.yml"
    cp -r /tmp/autodba-*/monitor/prometheus/sql_exporter/* "${EXPORTER_DIR}/sql_exporter/"
    cp -r /tmp/autodba-*/monitor/prometheus/rds_exporter/* "${EXPORTER_DIR}/rds_exporter/"
    cp /tmp/autodba-*/config/config.json "${AUTODBA_CONFIG_DIR}/config.json"
    cp /tmp/autodba-*/entrypoint.sh "${INSTALL_DIR}/autodba-entrypoint.sh"

    chmod +x "${INSTALL_DIR}/autodba-entrypoint.sh"
}

# Function to install from .deb
install_deb() {
    echo "Installing .deb package $PACKAGE_FILE..."
    dpkg -i "$PACKAGE_FILE"
}

# Function to install from .rpm
install_rpm() {
    echo "Installing .rpm package $PACKAGE_FILE..."
    rpm -i "$PACKAGE_FILE"
}

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Install Prometheus in custom path
install_prometheus

# Determine package type and install accordingly
case "$PACKAGE_FILE" in
    *.tar.gz)
        install_tar_gz
        ;;
    *.deb)
        install_deb
        ;;
    *.rpm)
        install_rpm
        ;;
    *)
        echo "Unsupported package format: $PACKAGE_FILE"
        exit 1
        ;;
esac

# Handle configuration file or fallback to environment variables
if [ -n "$CONFIG_FILE" ]; then
    if [ -f "$CONFIG_FILE" ]; then
        echo "Copying config file to $AUTODBA_CONFIG_FILE"
        mkdir -p "$AUTODBA_CONFIG_DIR"
        cp "$CONFIG_FILE" "$AUTODBA_CONFIG_FILE"
    else
        echo "Error: Config file $CONFIG_FILE does not exist."
        exit 1
    fi
elif [ ! -t 0 ]; then
    echo "Reading config from stdin and saving to $AUTODBA_CONFIG_FILE"

    # Check if jq is installed for JSON validation
    if ! command_exists "jq"; then
        echo "Error: jq is required for JSON validation but is not installed."
        exit 1
    fi

    # Read from stdin and validate it as JSON using jq
    mkdir -p "$AUTODBA_CONFIG_DIR"
    if ! cat > "$AUTODBA_CONFIG_FILE"; then
        echo "Error: Failed to save stdin input to $AUTODBA_CONFIG_FILE"
        exit 1
    fi

    # Validate the JSON file using jq
    if ! jq empty "$AUTODBA_CONFIG_FILE"; then
        echo "Error: Input from stdin is not valid JSON."
        rm -f "$AUTODBA_CONFIG_FILE"
        exit 1
    fi

    echo "Valid JSON config saved at $AUTODBA_CONFIG_FILE"
else
    echo "No config file provided, and no input from stdin detected. Using environment variables."

    # Remove any existing config file
    if [ -f "$AUTODBA_CONFIG_FILE" ]; then
        echo "Removing existing config file: $AUTODBA_CONFIG_FILE"
        rm -f "$AUTODBA_CONFIG_FILE"
    fi

    # Check if required parameters are provided
    if [[ -z "$DB_CONN_STRING" ]]; then
        echo "DB_CONN_STRING environment variable is not set"
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
    mkdir -p "$AUTODBA_CONFIG_DIR"
    cat <<EOF > "$AUTODBA_CONFIG_FILE"
{
    "DB_CONN_STRING": "${DB_CONN_STRING:-""}",
    "AWS_RDS_INSTANCE": "${AWS_RDS_INSTANCE:-""}",
    "AWS_ACCESS_KEY_ID": "${AWS_ACCESS_KEY_ID:-""}",
    "AWS_SECRET_ACCESS_KEY": "${AWS_SECRET_ACCESS_KEY:-""}",
    "AWS_REGION": "${AWS_REGION:-""}"
}
EOF
    echo "Generated config file at $AUTODBA_CONFIG_FILE"
fi

# Systemctl service installation (only if installing as root)
if $SYSTEM_INSTALL && command_exists "systemctl"; then
    echo "Installing systemd service..."
    cat << EOF | tee /etc/systemd/system/autodba.service
[Unit]
Description=AutoDBA Service
After=network.target

[Service]
Type=simple
WorkingDirectory=${AUTODBA_CONFIG_DIR}
ExecStart=${INSTALL_DIR}/autodba-entrypoint.sh
Restart=on-failure
User=$USER
Environment="PARENT_DIR=${PARENT_DIR}"

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable autodba
    systemctl start autodba
else
    echo "System installation not requested or systemctl is unavailable. Skipping systemd service setup."
    echo "You can run the following command to start the AutoDBA service manually:"
    
    echo "  cd \"${AUTODBA_CONFIG_DIR}\" && PARENT_DIR=\"${PARENT_DIR}\" CONFIG_FILE=${AUTODBA_CONFIG_FILE} ${INSTALL_DIR}/autodba-entrypoint.sh"
fi

echo "Installation complete."
