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

  # Check if collector_api_server exists before starting it. This acts as a feature-flag.
  if [ -d "$PARENT_DIR/share/collector_api_server" ]; then
    echo "Starting Collector API Server..."
    pushd "$PARENT_DIR/share/collector_api_server"
    ./collector-api-server &
    COLLECTOR_API_SERVER_PID=$!
    popd
  else
    echo "Warning: Skipping Collector API Server. Directory does not exist: $PARENT_DIR/share/collector_api_server"
  fi
  
  # Check if collector exists before starting it. This acts as a feature-flag.
  if [ -d "$PARENT_DIR/share/collector" ]; then
    # Start up Collector
    if [ -f "${CONFIG_FILE}" ]; then
      echo "Starting Collector..."
      $PARENT_DIR/share/collector/collector --config="${CONFIG_FILE}" --statefile="$PARENT_DIR/share/collector/state" --verbose &
      COLLECTOR_COLLECTOR_PID=$!
    else
      # Check if config file exists
      echo "Error: Collector configuration file not found at ${CONFIG_FILE}"
      exit 1
    fi
  else
    echo "Warning: Skipping Collector. Directory does not exist: $PARENT_DIR/share/collector"
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
