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

usage() {
    echo "Usage: $0 [--system] [--install-dir <path>] [--config <config.conf>]"
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
        echo "Stopping AutoDBA service..."
        systemctl stop autodba
    fi
fi

# Set paths relative to PARENT_DIR
INSTALL_DIR="$PARENT_DIR/bin"
WEBAPP_DIR="$PARENT_DIR/share/webapp"
PROMETHEUS_CONFIG_DIR="$PARENT_DIR/config/prometheus"
AUTODBA_CONFIG_DIR="$PARENT_DIR/config/autodba"
PROMETHEUS_STORAGE_DIR="$PARENT_DIR/prometheus_data"
PROMETHEUS_INSTALL_DIR="$PARENT_DIR/prometheus"

echo "Installing AutoDBA Agent under: $PARENT_DIR"

# Create directories only if PARENT_DIR is not current directory
if [ "$PARENT_DIR" != "$(pwd)" ]; then
    echo "Copying files to installation directory..."
    mkdir -p "${PARENT_DIR}"

    cp -r ./* "$PARENT_DIR"
else
    echo "Using the current directory for installation, no copying needed."
fi

chmod +x "${INSTALL_DIR}/autodba-entrypoint.sh"
chmod +x "${INSTALL_DIR}/prometheus-entrypoint.sh"
chmod +x "${INSTALL_DIR}/collector-api-entrypoint.sh"
chmod +x "${INSTALL_DIR}/bff-entrypoint.sh"

# Systemctl service setup (if needed)
if $SYSTEM_INSTALL && command_exists "systemctl"; then
    if ! id -u autodba >/dev/null 2>&1; then
        echo "Creating 'autodba' user..."

        if command_exists "useradd"; then
            useradd --system --user-group --home-dir /usr/local/autodba --shell /bin/bash autodba
        elif command_exists "adduser"; then
            adduser --system --group --home /usr/local/autodba --shell /bin/bash autodba
        else
            echo "Error: Neither 'useradd' nor 'adduser' found. Please create the user manually."
            exit 1
        fi
    fi
    chown -R autodba:autodba "$PARENT_DIR"
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

[Install]
WantedBy=multi-user.target
EOF

    systemctl daemon-reload
    systemctl enable autodba
    systemctl start autodba
else
    echo "System installation not requested or systemctl is unavailable. Skipping systemd service setup."
    echo "You can run the following command to start the AutoDBA service manually:"
    
    echo "  cd \"${AUTODBA_CONFIG_DIR}\" && PARENT_DIR=\"${PARENT_DIR}\" ${INSTALL_DIR}/autodba-entrypoint.sh"
fi

echo "Installation complete!"
