#!/bin/bash

# SPDX-Identifier: Apache-2.0

# Set the base directory based on installation
PARENT_DIR="${PARENT_DIR:-/usr/local/autodba}"

# Load environment variables from a JSON config file if provided
if [ -n "$CONFIG_FILE" ]; then
    echo "Loading environment variables from JSON config file: $CONFIG_FILE"
    if [ -f "$CONFIG_FILE" ]; then
        DB_CONN_STRING=$(jq -r '.DB_CONN_STRING' "$CONFIG_FILE")
        AWS_RDS_INSTANCE=$(jq -r '.AWS_RDS_INSTANCE' "$CONFIG_FILE")
        AWS_ACCESS_KEY_ID=$(jq -r '.AWS_ACCESS_KEY_ID' "$CONFIG_FILE")
        AWS_SECRET_ACCESS_KEY=$(jq -r '.AWS_SECRET_ACCESS_KEY' "$CONFIG_FILE")
        AWS_REGION=$(jq -r '.AWS_REGION' "$CONFIG_FILE")
    else
        echo "Error: Config file $CONFIG_FILE does not exist."
        exit 1
    fi
fi

# Ensure required environment variables are set
if [ -z "$DB_CONN_STRING" ]; then
  echo "Error: DB_CONN_STRING environment variable is not set."
  exit 1
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

# Parse the DB connection string
parse_db_conn_string "$DB_CONN_STRING"

# Ensure the directory for the config file exists
COLLECTOR_CONFIG_DIR="$PARENT_DIR/share/collector"
mkdir -p "$COLLECTOR_CONFIG_DIR"

# Create Collector config file from parsed values
COLLECTOR_CONFIG_FILE="$COLLECTOR_CONFIG_DIR/collector.conf"
cat > "$COLLECTOR_CONFIG_FILE" <<EOL
[pganalyze]
api_key = your-secure-api-key
api_base_url = http://localhost:7080

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
EOL

echo "Collector configuration file created at $COLLECTOR_CONFIG_FILE"

function clean_up {
    # Perform program exit housekeeping
    kill $(jobs -p)
    wait # wait for all children to exit -- this lets their logs make it out of the container environment
    exit -1
}

trap clean_up SIGHUP SIGINT SIGTERM

if [ -z "$DISABLE_DATA_COLLECTION" ] || [ "$DISABLE_DATA_COLLECTION" = false ]; then
  # Start up Prometheus Postgres Exporter
  DATA_SOURCE_NAME="$DB_CONN_STRING" "$PARENT_DIR/share/prometheus_exporters/postgres_exporter/postgres_exporter" \
    --exclude-databases="rdsadmin" \
    --collector.database \
    --collector.database_wraparound \
    --collector.locks \
    --collector.long_running_transactions \
    --collector.postmaster \
    --collector.process_idle \
    --collector.replication \
    --collector.replication_slot \
    --collector.stat_activity_autovacuum \
    --collector.stat_bgwriter \
    --collector.stat_database \
    --collector.stat_statements \
    --collector.stat_user_tables \
    --collector.stat_wal_receiver \
    --collector.statio_user_indexes \
    --collector.statio_user_tables \
    --collector.wal &
  POSTGRES_EXPORTER_PID=$!

  # Start up Prometheus SQL Exporter
  pushd "$PARENT_DIR/share/prometheus_exporters/sql_exporter"
  ./sql_exporter -config.data-source-name "$DB_CONN_STRING" &
  SQL_EXPORTER_PID=$!
  popd

  # Start up Prometheus RDS Exporter
  if [[ -n "$AWS_ACCESS_KEY_ID" && -n "$AWS_SECRET_ACCESS_KEY" && -n "$AWS_REGION" ]]; then
    AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
    AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
    "$PARENT_DIR/share/prometheus_exporters/rds_exporter/prometheus-rds-exporter" \
      -c "$PARENT_DIR/share/prometheus_exporters/rds_exporter/prometheus-rds-exporter.yaml" \
      --filter-instances "$AWS_RDS_INSTANCE" &
    RDS_EXPORTER_PID=$!
  else
    echo "AWS environment variables are missing or empty, so not running the RDS Exporter."
  fi

  echo "Starting Collector API Server..."
  pushd "$PARENT_DIR/share/collector_api_server"
  ./collector-api-server &
  COLLECTOR_API_SERVER_PID=$!
  popd
  
  # Start up Collector
  if [ -f "$COLLECTOR_CONFIG_FILE" ]; then
    echo "Starting Collector..."
    $PARENT_DIR/share/collector/collector --config="$COLLECTOR_CONFIG_FILE" &
    COLLECTOR_PID=$!
  else
    echo "Collector configuration file not found, skipping Collector startup."
  fi
fi

# Start up Prometheus for initialization
"$PARENT_DIR/prometheus/prometheus" \
    --config.file="$PARENT_DIR/config/prometheus/prometheus.yml" \
    --storage.tsdb.path="$PARENT_DIR/prometheus_data" \
    --web.console.templates="$PARENT_DIR/config/prometheus/consoles" \
    --web.console.libraries="$PARENT_DIR/config/prometheus/console_libraries" \
    --web.enable-admin-api &
PROMETHEUS_PID=$!

if [ -n "$BACKUP_FILE" ]; then
  echo "Restoring Prometheus backup from $BACKUP_FILE..."

  sleep 1

  # Ensure Prometheus is stopped
  kill $PROMETHEUS_PID
  wait $PROMETHEUS_PID || true

  sleep 1

  # Restore Prometheus snapshot
  mkdir -p "$PARENT_DIR/prometheus"
  cp -r /home/autodba/restore_backup_uncompressed/home/autodba/backups/prometheus_snapshot_recent/* "$PARENT_DIR/prometheus/"
  echo "Prometheus backup restored."

  # Start up Prometheus for real
  "$PARENT_DIR/prometheus/prometheus" \
      --config.file="$PARENT_DIR/config/prometheus/prometheus.yml" \
      --storage.tsdb.path="$PARENT_DIR/prometheus" \
      --web.console.templates="$PARENT_DIR/config/prometheus/consoles" \
      --web.console.libraries="$PARENT_DIR/config/prometheus/console_libraries" \
      --web.enable-admin-api &
fi

# Start up bff
"$PARENT_DIR/bin/autodba-bff" -dbidentifier="$AWS_RDS_INSTANCE" -webappPath "$PARENT_DIR/share/webapp" &
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
elif (( $EXITED_PID == $POSTGRES_EXPORTER_PID ))
then
  echo "Postgres Exporter exited with return code $retcode - killing all jobs"
elif (( $EXITED_PID == $SQL_EXPORTER_PID ))
then
  echo "SQL Exporter exited with return code $retcode - killing all jobs"
elif (( $EXITED_PID == $RDS_EXPORTER_PID ))
then
  echo "RDS Exporter exited with return code $retcode - killing all jobs"
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
