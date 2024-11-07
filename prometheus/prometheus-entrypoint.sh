#!/bin/bash

# SPDX-Identifier: Apache-2.0

# Set the base directory based on installation
PARENT_DIR="${PARENT_DIR:-/usr/local/autodba}"

# Start up Prometheus for initialization
"$PARENT_DIR/prometheus/prometheus" \
    --config.file="$PARENT_DIR/config/prometheus/prometheus.yml" \
    --storage.tsdb.path="$PARENT_DIR/prometheus_data" \
    --query.lookback-delta="15m" \
    --web.console.templates="$PARENT_DIR/config/prometheus/consoles" \
    --web.console.libraries="$PARENT_DIR/config/prometheus/console_libraries" \
    --web.enable-remote-write-receiver \
    --web.enable-admin-api
