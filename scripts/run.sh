#!/bin/bash

# SPDX-License-Identifier: Apache-2.0
set -e -u -o pipefail

# Get the directory of the currently executing script
SOURCE_DIR=$(dirname "$(readlink -f "$0")")

cd $SOURCE_DIR/..

# Initialize variables
IMAGE_NAME="autodba-image"
DOCKERFILE="Dockerfile"
DB_CONN_STRING= # 'postgresql://autodba_db_user:autodba_db_pass@localhost:5432/autodba_db'
AWS_RDS_INSTANCE="" # 'YOURNAME-rds-EXAMPLE'
AWS_REGION=""
AWS_ACCESS_KEY_ID=""
AWS_SECRET_ACCESS_KEY=""
INSTANCE_ID=0  # Default value for instance-id
DEFAULT_METRIC_COLLECTION_PERIOD_SECONDS=5 # Default value for metric collection period (in seconds)
WARM_UP_TIME_SECONDS=60  # Default value for warm-up time (in seconds)
DISABLE_DATA_COLLECTION=false  # Flag to disable data collection
CONFIG_FILE="" # Path to the configuration file
MOUNT_PROMETHEUS_DATA=false # Default to not mounting the prometheus data directory

command_exists() {
    command -v "$1" >/dev/null 2>&1
}

# Function to display usage information
usage() {
    echo "Usage: $0 --db-url <TARGET_DATABASE_URL> [--config <CONFIG_FILE>] [--instance-id <INSTANCE_ID>] [--rds-instance <RDS_INSTANCE_NAME>] [--disable-data-collection] [--continue]"
    echo "Options:"
    echo "--config                    <CONFIG_FILE> path to the configuration file"
    echo "--instance-id               <INSTANCE_ID> if you are running multiple instances of the agent, specify a unique number for each"
    echo "--rds-instance              <RDS_INSTANCE_NAME> collect metrics from an AWS RDS instance"
    echo "--disable-data-collection   to disable the agent collectors from collecting data from the target database"
    exit 1
}

# Parse command-line options
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --config)
            if [[ -n "$2" ]] && [[ ${2:0:1} != "-" ]]; then
                CONFIG_FILE="$2"
                shift 2
            else
                echo "Error: Argument for $1 is missing" >&2
                exit 1
            fi
            ;;
        --db-url)
            if [[ -n "$2" ]] && [[ ${2:0:1} != "-" ]]; then
                DB_CONN_STRING="$2"
                shift 2
            else
                echo "Error: Argument for $1 is missing" >&2
                exit 1
            fi
            ;;
        --instance-id)
            if [[ -n "$2" ]] && [[ ${2:0:1} != "-" ]]; then
                INSTANCE_ID="$2"
                shift 2
            else
                echo "Error: Argument for $1 is missing" >&2
                exit 1
            fi
            ;;
        --instance_id)
          # backwards legacy support
            if [[ -n "$2" ]] && [[ ${2:0:1} != "-" ]]; then
                INSTANCE_ID="$2"
                shift 2
            else
                echo "Error: Argument for $1 is missing" >&2
                exit 1
            fi
            ;;
        --rds-instance)
            if [[ -n "$2" ]] && [[ ${2:0:1} != "-" ]]; then
                AWS_RDS_INSTANCE="$2"
                shift 2
            else
                echo "Error: Argument for $1 is missing" >&2
                exit 1
            fi
            ;;
        --disable-data-collection)
            DISABLE_DATA_COLLECTION=true
            shift
            ;;
        --mount-prometheus-data)
            MOUNT_PROMETHEUS_DATA=true
            shift
            ;;
        *)
            echo "Invalid argument: $1" >&2
            usage
            ;;
    esac
done
# Check if AWS CLI is available
if [[ -n "$AWS_RDS_INSTANCE" ]]; then
    if command_exists "aws"; then
        AWS_ACCESS_KEY_ID=$(aws configure get aws_access_key_id || echo "")
        AWS_SECRET_ACCESS_KEY=$(aws configure get aws_secret_access_key || echo "")
        AWS_REGION=$(aws configure get region || echo "")

        if [[ -z "$AWS_ACCESS_KEY_ID" || -z "$AWS_SECRET_ACCESS_KEY" || -z "$AWS_REGION" ]]; then
            echo "Warning: AWS credentials or region are not configured properly. Proceeding without AWS integration."
        fi
    else
        echo "Warning: AWS CLI is not installed. Proceeding without AWS integration."
    fi
else
    echo "Warning: --rds-instance not specified. Starting without RDS Instance metrics."
fi


# Function to extract parts from the DB connection string
function parse_db_conn_string() {
  local conn_string="$1"
  
  db_host=$(echo "$conn_string" | sed -E 's/.*@([^:]+).*/\1/')  # Extract host
  db_name=$(echo "$conn_string" | sed -E 's/.*\/([^?]+).*/\1/')  # Correct extraction of dbname
  db_username=$(echo "$conn_string" | sed -E 's/.*\/\/([^:]+):.*/\1/')  # Extract username
  db_password=$(echo "$conn_string" | sed -E 's/.*\/\/[^:]+:([^@]+)@.*/\1/')  # Extract password
  db_port=$(echo "$conn_string" | sed -E 's/.*:(543[0-9]{1}).*/\1/')  # Extract port
  
  echo "Parsed connection string:"
  echo "  DB Host: $db_host"
  echo "  DB Name: $db_name"
  echo "  DB Username: $db_username"
  echo "  DB Password: (hidden)"
  echo "  DB Port: $db_port"
}

# Function to create config file
create_config_file() {
    local config_file="$1"

    # Parse the DB connection string
    parse_db_conn_string "$DB_CONN_STRING"

    cat > "$config_file" <<EOF
[server1]
db_host = $db_host
db_name = $db_name
db_username = $db_username
db_password = $db_password
db_port = $db_port
aws_db_instance_id = $AWS_RDS_INSTANCE
aws_region = $AWS_REGION
aws_access_key_id = $AWS_ACCESS_KEY_ID
aws_secret_access_key = $AWS_SECRET_ACCESS_KEY
EOF
}

# Handle config file
FIXED_CONFIG_FILE="./collector/autodba-collector.conf"
if [[ -z "$CONFIG_FILE" ]]; then
    echo "No config file provided. Creating one based on provided parameters..."

    # Check if required parameters are provided
    if [[ -z "$DB_CONN_STRING" ]]; then
        echo "Missing required parameters"
        usage
    fi
    create_config_file "$FIXED_CONFIG_FILE"
else
    if [[ ! -f "$CONFIG_FILE" ]]; then
        echo "Error: Provided config file $CONFIG_FILE does not exist" >&2
        exit 1
    fi

    cp $CONFIG_FILE $FIXED_CONFIG_FILE
fi

# Adjust port numbers based on INSTANCE_ID
PROMETHEUS_PORT=$((UID + 6000 + INSTANCE_ID))
BFF_PORT=$((UID + 4000 + INSTANCE_ID))
COLLECTOR_API_SERVER_PORT=$((UID + 7080 + INSTANCE_ID))
CONTAINER_NAME="autodba-$USER-$INSTANCE_ID"

# Build Docker image
echo "Running tests + lint ..."
# docker build -f "$DOCKERFILE" . --target lint
# docker build -f "$DOCKERFILE" . --target test

# Build Docker image
echo "Building Docker image using Dockerfile: $DOCKERFILE ..."
docker build -t "$IMAGE_NAME" -f "$DOCKERFILE" .

echo "Stopping and removing existing container '$CONTAINER_NAME'..."
docker stop "$CONTAINER_NAME" > /dev/null 2>/dev/null || true
docker rm "$CONTAINER_NAME" > /dev/null 2>/dev/null || true

# Prepare the Prometheus data volume mount
PROMETHEUS_DATA_MOUNT=""
if [ "$MOUNT_PROMETHEUS_DATA" = true ]; then
    PROMETHEUS_DATA_MOUNT="-v $SOURCE_DIR/../prometheus_data:/usr/local/autodba/prometheus_data"
fi

# Run the container
echo "=============================================================="
echo ""
echo "Running Docker container: $CONTAINER_NAME"
echo ""
echo "           prometheus port: $PROMETHEUS_PORT"
echo "                  bff port: $BFF_PORT"
echo " collector API server port: $COLLECTOR_API_SERVER_PORT"
echo ""
echo "=============================================================="

docker run --name "$CONTAINER_NAME" \
    -p "$PROMETHEUS_PORT":9090 \
    -p "$BFF_PORT":4000 \
    -p "$COLLECTOR_API_SERVER_PORT":7080 \
    -e DB_CONN_STRING="$DB_CONN_STRING" \
    -e AWS_RDS_INSTANCE="$AWS_RDS_INSTANCE" \
    -e DEFAULT_METRIC_COLLECTION_PERIOD_SECONDS=$DEFAULT_METRIC_COLLECTION_PERIOD_SECONDS \
    -e WARM_UP_TIME_SECONDS="$WARM_UP_TIME_SECONDS" \
    -e AWS_ACCESS_KEY_ID="${AWS_ACCESS_KEY_ID:-""}" \
    -e AWS_SECRET_ACCESS_KEY="${AWS_SECRET_ACCESS_KEY:-""}" \
    -e AWS_REGION="${AWS_REGION:-""}" \
    -e DISABLE_DATA_COLLECTION="$DISABLE_DATA_COLLECTION" \
    -e CONFIG_FILE="/usr/local/autodba/config/autodba/collector.conf" \
    --mount type=bind,source="$FIXED_CONFIG_FILE",target="/usr/local/autodba/config/autodba/collector.conf",readonly \
    $PROMETHEUS_DATA_MOUNT \
    "$IMAGE_NAME"
    # -v "$VOLUME_NAME":/var/lib/postgresql/data \
    # --env-file "$ENV_FILE" \
