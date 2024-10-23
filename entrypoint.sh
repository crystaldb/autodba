#!/bin/bash

# SPDX-Identifier: Apache-2.0

# Set the base directory based on installation
PARENT_DIR="${PARENT_DIR:-/usr/local/autodba}"
CONFIG_FILE="${CONFIG_FILE:-${PARENT_DIR}/share/collector/collector.conf}"

# Check if config file exists
if [ ! -f "${CONFIG_FILE}" ]; then
    echo "Error: Config file not found at ${CONFIG_FILE}"
    exit 1
fi

# Function to extract value from the first server section
extract_value() {
    grep -m1 "^$1" "${CONFIG_FILE}" | cut -d'=' -f2- | tr -d ' '
}

# Extract values for the first server
DB_HOST=$(extract_value "db_host")
DB_NAME=$(extract_value "db_name")
DB_USERNAME=$(extract_value "db_username")
DB_PASSWORD=$(extract_value "db_password")
DB_PORT=$(extract_value "db_port")
AWS_RDS_INSTANCE=$(extract_value "aws_db_instance_id")
AWS_REGION=$(extract_value "aws_region")
AWS_ACCESS_KEY_ID=$(extract_value "aws_access_key_id")
AWS_SECRET_ACCESS_KEY=$(extract_value "aws_secret_access_key")

# Construct the DB_CONN_STRING
DB_CONN_STRING="postgresql://${DB_USERNAME}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}"

# Ensure required environment variables are set
if [ -z "$DB_CONN_STRING" ]; then
  echo "Error: there is no PostgreSQL server connection string defined in the provided config."
  exit 1
fi

function clean_up {
    # Perform program exit housekeeping
    kill $(jobs -p)
    wait # wait for all children to exit -- this lets their logs make it out of the container environment
    exit -1
}

trap clean_up SIGHUP SIGINT SIGTERM

if [ -z "$DISABLE_DATA_COLLECTION" ] || [ "$DISABLE_DATA_COLLECTION" = false ]; then
  echo "Starting Collector API Server..."
  pushd "$PARENT_DIR/share/collector_api_server"
  ./collector-api-server &
  COLLECTOR_API_SERVER_PID=$!
  popd
  
  # Start up Collector
  if [ -f "${CONFIG_FILE}" ]; then
    echo "Starting Collector..."
    # Create a temporary file with the prefix and original content
    TEMP_CONFIG=$(mktemp)
    {
      echo "[pganalyze]"
      echo "api_key = your-secure-api-key"
      echo "api_base_url = http://localhost:7080"
      echo ""
      cat "${CONFIG_FILE}"
    } > "$TEMP_CONFIG"
    $PARENT_DIR/share/collector/collector --config="${TEMP_CONFIG}" --statefile="$PARENT_DIR/share/collector/state" --verbose &
    COLLECTOR_COLLECTOR_PID=$!
  else
    # Check if config file exists
    echo "Error: Collector configuration file not found at ${CONFIG_FILE}"
    exit 1
  fi
fi

# Start up Prometheus for initialization
"$PARENT_DIR/prometheus/prometheus" \
    --config.file="$PARENT_DIR/config/prometheus/prometheus.yml" \
    --enable-feature="remote-write-receiver" \
    --storage.tsdb.path="$PARENT_DIR/prometheus_data" \
    --query.lookback-delta="15m" \
    --web.console.templates="$PARENT_DIR/config/prometheus/consoles" \
    --web.console.libraries="$PARENT_DIR/config/prometheus/console_libraries" \
    --web.enable-admin-api &
PROMETHEUS_PID=$!

# Start up bff
"$PARENT_DIR/bin/autodba-bff" -collectorConfigFile="$CONFIG_FILE" -webappPath "$PARENT_DIR/share/webapp" &
BFF_PID=$!

# Wait for a process to exit
wait -n # wait for any job to exit
EXITED_PID=$!
retcode=$? # store error code so we can propagate it to the container environment

if (( $EXITED_PID == $PROMETHEUS_PID ))
then
  echo "Prometheus exited with return code $retcode - killing all jobs"
elif (( $EXITED_PID == $BFF_PID ))
then
  echo "BFF exited with return code $retcode - killing all jobs"
elif (( $EXITED_PID == $COLLECTOR_PID ))
then
  echo "Collector exited with return code $retcode - killing all jobs"
elif (( $EXITED_PID == $COLLECTOR_API_SERVER_PID ))
then
  echo "Collector API Server exited with return code $retcode - killing all jobs"
else
  echo "An unknown background process (with PID=$EXITED_PID) exited with return code $retcode - killing all jobs"
fi

kill $(jobs -p)

wait # wait for all children to exit -- this lets their logs make it out of the container environment
echo done

exit $retcode
