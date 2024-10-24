#!/bin/bash

# SPDX-Identifier: Apache-2.0

# Set the base directory based on installation
PARENT_DIR="${PARENT_DIR:-/usr/local/autodba}"
CONFIG_FILE="${CONFIG_FILE:-${PARENT_DIR}/share/collector/collector.conf}"

if [ ! -f "${CONFIG_FILE}" ]; then
    echo "Error: Config file not found at ${CONFIG_FILE}"
    exit 1
fi

# Create a temporary file with the prefix and original content
TEMP_CONFIG=$(mktemp)
{
  echo "[pganalyze]"
  echo "api_key = your-secure-api-key"
  echo "api_base_url = http://collector-api:7080"
  echo ""
  cat "${CONFIG_FILE}"
} > "$TEMP_CONFIG"

$PARENT_DIR/share/collector/collector --config="${TEMP_CONFIG}" --statefile="$PARENT_DIR/share/collector/state" --verbose
