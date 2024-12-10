#!/bin/bash

set -e

# Get the current directory
CURRENT_DIR=$(pwd)

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
VERSION="0.6.0"
ARCH="amd64"
CONFIG_PATH="${CURRENT_DIR}/autodba.conf"

# Function to print colored messages
log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

show_usage() {
    cat << EOF
Usage: $0 [OPTIONS]
Options:
    --config PATH          Path to autodba.conf file (default: ./autodba.conf)
    --version VERSION      AutoDBA version (default: 0.6.0)
    --help                Show this help message
EOF
    exit 1
}

# Parse command line arguments
while [ "$#" -gt 0 ]; do
    case "$1" in
        --config=*)
            CONFIG_PATH="${1#*=}"
            ;;
        --version=*)
            VERSION="${1#*=}"
            ;;
        --help)
            show_usage
            ;;
        *)
            log_error "Unknown parameter: $1"
            show_usage
            ;;
    esac
    shift
done

# Check prerequisites
if [ "$EUID" -ne 0 ]; then 
    log_error "Please run as root (sudo)"
    exit 1
fi

if [ ! -f "$CONFIG_PATH" ]; then
    log_error "Config file not found at: $CONFIG_PATH"
    exit 1
fi

# Install AutoDBA Agent
log_info "Installing AutoDBA Agent..."
cd ${CURRENT_DIR}
wget "https://github.com/crystaldb/autodba/releases/latest/download/autodba-${VERSION}-${ARCH}.tar.gz"
tar -xzvf "autodba-${VERSION}-${ARCH}.tar.gz"
cd "autodba-${VERSION}"
./install.sh --system

# Verify AutoDBA Agent installation
if ! systemctl is-active --quiet autodba; then
    log_error "AutoDBA Agent installation failed"
    exit 1
fi
log_info "AutoDBA Agent installed successfully"

# Install AutoDBA Collector
log_info "Installing AutoDBA Collector..."
cd ${CURRENT_DIR}
wget "https://github.com/crystaldb/autodba/releases/latest/download/collector-${VERSION}-${ARCH}.tar.gz"
tar -xzvf "collector-${VERSION}-${ARCH}.tar.gz"
cd "collector-${VERSION}"
./install.sh --config "$CONFIG_PATH" --system

# Verify AutoDBA Collector installation
if ! systemctl is-active --quiet autodba-collector; then
    log_error "AutoDBA Collector installation failed"
    exit 1
fi
log_info "AutoDBA Collector installed successfully"

# Clean up downloaded files
cd ${CURRENT_DIR}
rm -rf "autodba-${VERSION}" "collector-${VERSION}" "autodba-${VERSION}-${ARCH}.tar.gz" "collector-${VERSION}-${ARCH}.tar.gz"

log_info "Installation completed successfully!"
log_info "You can access the AutoDBA web portal at http://localhost:4000"
log_warn "Please make sure to secure your installation and keep your credentials safe"
