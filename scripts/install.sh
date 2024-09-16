#!/bin/bash

# SPDX-Identifier: Apache-2.0

set -e

# Initialize variables
SYSTEM_INSTALL=false
USER_INSTALL_DIR=""
CONFIG_FILE=""

# Parse arguments
while [[ $# -gt 0 ]]; do
    case "$1" in
        --system)
            SYSTEM_INSTALL=true
            shift
            ;;
        --install-dir)
            USER_INSTALL_DIR="$2"
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
    echo "Usage: $0 [--system] [--install-dir <path>] [--config <config.json>]"
    echo ""
    echo "Options:"
    echo "  --system       Install globally under /usr/local/autodba"
    echo "  --install-dir  Specify a custom installation directory. If not specified, $HOME/autodba is used."
    echo "  --config       Path to the AutoDBA config file (optional), or use standard input if not provided."
    exit 1
}

# Set the parent directory
if [ -n "$USER_INSTALL_DIR" ]; then
    PARENT_DIR="$USER_INSTALL_DIR"
elif [ "$SYSTEM_INSTALL" = true ]; then
    PARENT_DIR="/usr/local/autodba"
else
    PARENT_DIR="$(pwd)"
fi

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Stop the service if it's already running
if $SYSTEM_INSTALL && command_exists "systemctl"; then
    if systemctl is-active --quiet autodba; then
        echo "Stopping running AutoDBA service..."
        systemctl stop autodba
    fi
fi

# Set paths relative to PARENT_DIR
INSTALL_DIR="$PARENT_DIR/bin"
WEBAPP_DIR="$PARENT_DIR/share/webapp"
EXPORTER_DIR="$PARENT_DIR/share/prometheus_exporters"
PROMETHEUS_CONFIG_DIR="$PARENT_DIR/config/prometheus"
AUTODBA_CONFIG_DIR="$PARENT_DIR/config/autodba"
PROMETHEUS_STORAGE_DIR="$PARENT_DIR/prometheus_data"
PROMETHEUS_INSTALL_DIR="$PARENT_DIR/prometheus"
AUTODBA_CONFIG_FILE="$AUTODBA_CONFIG_DIR/autodba-config.json"

echo "Installing AutoDBA under: $PARENT_DIR"

# Create directories only if PARENT_DIR is not current directory
if [ "$PARENT_DIR" != "$(pwd)" ]; then
    echo "Copying files to installation directory..."
    mkdir -p "${PARENT_DIR}"

    cp -r ./* "$PARENT_DIR"
    chmod +x "${INSTALL_DIR}/autodba-entrypoint.sh"
else
    echo "Using the current directory for installation, no copying needed."
fi

# Handle configuration file
if [ -n "$CONFIG_FILE" ]; then
    if [ -f "$CONFIG_FILE" ]; then
        cp "$CONFIG_FILE" "${AUTODBA_CONFIG_FILE}"
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

    # Read and validate the JSON file using jq
    mkdir -p "$AUTODBA_CONFIG_DIR"
    if ! cat > "$AUTODBA_CONFIG_FILE"; then
        echo "Error: Failed to save stdin input to $AUTODBA_CONFIG_FILE"
        exit 1
    fi

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
    cat <<EOF > "${AUTODBA_CONFIG_FILE}"
{
    "DB_CONN_STRING": "${DB_CONN_STRING:-""}",
    "AWS_RDS_INSTANCE": "${AWS_RDS_INSTANCE:-""}",
    "AWS_ACCESS_KEY_ID": "${AWS_ACCESS_KEY_ID:-""}",
    "AWS_SECRET_ACCESS_KEY": "${AWS_SECRET_ACCESS_KEY:-""}",
    "AWS_REGION": "${AWS_REGION:-""}"
}
EOF
    echo "Generated config file at ${AUTODBA_CONFIG_FILE}"
fi

# Systemctl service setup (if needed)
if $SYSTEM_INSTALL && command_exists "systemctl"; then
    echo "Setting up systemd service..."
    cat << EOF | tee /etc/systemd/system/autodba.service
[Unit]
Description=AutoDBA Service
After=network.target

[Service]
Type=simple
WorkingDirectory=${AUTODBA_CONFIG_DIR}
ExecStart=${INSTALL_DIR}/autodba-entrypoint.sh
Restart=on-failure
User=autodba
Group=autodba
Environment="PARENT_DIR=${PARENT_DIR}"
Environment="CONFIG_FILE=${AUTODBA_CONFIG_DIR}/autodba-config.json"

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

echo "Installation complete!"
