#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
set -e -u -o pipefail

# Get the directory of the currently executing script
SOURCE_DIR=$(dirname "$(readlink -f "$0")")

cd $SOURCE_DIR/..

# Initialize variables
INSTANCE_ID=0
CONFIG_FILE=""
KEEP_CONTAINERS=false
USE_COLLECTOR=true
REPROCESS_FULL=false
REPROCESS_COMPACT=false

# Function to display usage information
usage() {
    echo "Usage: $0 [--config <CONFIG_FILE>] [--instance-id <INSTANCE_ID>] [--keep-containers] [--no-collector]"
    echo "Options:"
    echo "--config                      <CONFIG_FILE> path to the configuration file (required unless --no-collector is set)"
    echo "--instance-id                 <INSTANCE_ID> if you are running multiple instances of the agent, specify a unique number for each"
    echo "--reprocess-full-snapshots    reprocess all full snapshots from storage"
    echo "--reprocess-compact-snapshots reprocess all compact snapshots from storage"
    echo "--keep-containers             keep containers running after script exits"
    echo "--no-collector                run without the collector component"
    exit 1
}

# Parse command-line options
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --config)
            CONFIG_FILE="$2"
            shift 2
            ;;
        --instance-id)
            INSTANCE_ID="$2"
            shift 2
            ;;
        --keep-containers)
            KEEP_CONTAINERS=true
            shift
            ;;
        --no-collector)
            USE_COLLECTOR=false
            shift
            ;;
        --reprocess-full-snapshots)
            REPROCESS_FULL=true
            shift
            ;;
        --reprocess-compact-snapshots)
            REPROCESS_COMPACT=true
            shift
            ;;
        *)
            echo "Invalid argument: $1" >&2
            usage
            ;;
    esac
done

# Handle config file
FIXED_CONFIG_FILE="./collector/autodba-collector.conf"
if [[ -z "$CONFIG_FILE" ]]; then
    if [ "$USE_COLLECTOR" = true ]; then
        echo "Error: Config file is required when using collector" >&2
        usage
        exit 1
    fi
    echo "No config file provided, continuing without collector configuration..."
else
    if [[ ! -f "$CONFIG_FILE" ]]; then
        echo "Error: Provided config file $CONFIG_FILE does not exist" >&2
        exit 1
    fi
    cp $CONFIG_FILE $FIXED_CONFIG_FILE
fi

INSTANCE_NAME="autodba-${USER//./_}-${INSTANCE_ID}"

function clean_up {
    # Only stop containers if --keep-containers wasn't specified
    if [ "$KEEP_CONTAINERS" = false ]; then
        # Perform program exit housekeeping
        echo "Stopping all containers: ${INSTANCE_NAME}*"
        docker ps --filter name="${INSTANCE_NAME}*" --filter status=running -aq | xargs docker stop
        echo "Killing child processes"
        kill $(jobs -p)
        wait # wait for all children to exit -- this lets their logs make it out of the container environment
    fi
    exit -1
}

trap clean_up SIGHUP SIGINT SIGTERM

# Set environment variables for docker-compose
export COLLECTOR_API_PORT=$((UID + 7000 + INSTANCE_ID))
export BFF_WEBAPP_PORT=$((UID + 4000 + INSTANCE_ID))
export PROMETHEUS_PORT=$((UID + 6000 + INSTANCE_ID))
export CONFIG_FILE="/usr/local/autodba/config/autodba/collector.conf"

# Prepare docker-compose command
if [ "$USE_COLLECTOR" = true ]; then
    COMPOSE_CMD="docker-compose -p ${INSTANCE_NAME} -f compose.yaml -f compose.collector.yaml"
else
    COMPOSE_CMD="docker-compose -p ${INSTANCE_NAME} -f compose.yaml"
fi

# Stop and remove existing containers
echo "Stopping and removing existing containers..."
$COMPOSE_CMD down

# Set reprocessing environment variables
if [ "$REPROCESS_FULL" = true ]; then
    export AUTODBA_REPROCESS_FULL_SNAPSHOTS=true
else
    export AUTODBA_REPROCESS_FULL_SNAPSHOTS=false
fi

if [ "$REPROCESS_COMPACT" = true ]; then
    export AUTODBA_REPROCESS_COMPACT_SNAPSHOTS=true
else
    export AUTODBA_REPROCESS_COMPACT_SNAPSHOTS=false
fi

# Build and start the containers
echo "Building and starting containers..."
$COMPOSE_CMD up --build -d

echo "=============================================================="
echo ""
echo "Running Docker containers for AutoDBA"
echo ""
echo "           prometheus port: $PROMETHEUS_PORT"
echo "                  bff port: $BFF_WEBAPP_PORT"
echo " collector API server port: $COLLECTOR_API_PORT"
echo ""
echo "=============================================================="

# Tail the logs for convenience
$COMPOSE_CMD logs -f
