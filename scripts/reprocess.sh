#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

log_info() {
    echo -e "${GREEN}[INFO]${NC} $1"
}

log_warn() {
    echo -e "${YELLOW}[WARN]${NC} $1"
}

# Stop services
log_info "Stopping AutoDBA services..."
systemctl stop autodba autodba-collector

# Set environment variables for reprocessing
export AUTODBA_REPROCESS_FULL_SNAPSHOTS=true
export AUTODBA_REPROCESS_COMPACT_SNAPSHOTS=true

# Start services with reprocessing flags
log_info "Starting services with reprocessing enabled..."
systemctl start autodba
systemctl start autodba-collector

log_info "Reprocessing is running. Check the logs for progress."
