#!/bin/bash

# SPDX-Identifier: Apache-2.0

# Set the base directory based on installation
CONFIG_FILE="${CONFIG_FILE:-./collector.conf}"

if [ ! -f "${CONFIG_FILE}" ]; then
    echo "Error: Config file not found at ${CONFIG_FILE}"
    exit 1
fi

# Check if [autodba] section exists
if ! grep -q "\[autodba\]" "${CONFIG_FILE}"; then
    echo "Error: Required [autodba] section not found in ${CONFIG_FILE}"
    exit 1
fi

# Create a temporary file with the prefix and original content
TEMP_CONFIG=$(mktemp)
{
  sed 's/\[autodba\]/[pganalyze]/' "${CONFIG_FILE}"
} > "$TEMP_CONFIG"

./autodba-collector --config="${TEMP_CONFIG}" --statefile="./state" --verbose
