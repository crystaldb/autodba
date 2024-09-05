#!/bin/bash

# SPDX-Identifier: Apache-2.0

# Set the base directory based on installation
PARENT_DIR="${PARENT_DIR:-/usr/local/autodba}"

# Load environment variables from a JSON config file if provided
if [ -n "$CONFIG_FILE" ]; then
    echo "Loading environment variables from JSON config file: $CONFIG_FILE"
    if [ -f "$CONFIG_FILE" ]; then
        AUTODBA_TARGET_DB=$(jq -r '.AUTODBA_TARGET_DB' "$CONFIG_FILE")
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
if [ -z "$AUTODBA_TARGET_DB" ]; then
  echo "Error: AUTODBA_TARGET_DB environment variable is not set."
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
  DATA_SOURCE_NAME="$AUTODBA_TARGET_DB" "$PARENT_DIR/share/prometheus_exporters/postgres_exporter/postgres_exporter" \
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
  ./sql_exporter -config.data-source-name "$AUTODBA_TARGET_DB" &
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
wait -n -p EXITED_PID # wait for any job to exit
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
else
  echo "An unknown background process (with PID=$EXITED_PID) exited with return code $retcode - killing all jobs"
fi

kill $(jobs -p)

wait # wait for all children to exit -- this lets their logs make it out of the container environment
echo done

exit $retcode
