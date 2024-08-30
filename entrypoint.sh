#!/bin/bash

# SPDX-License-Identifier: Apache-2.0

if [ "$DISABLE_DATA_COLLECTION" = false ]; then
  # Start up Prometheus Postgres Exporter
  DATA_SOURCE_NAME="$AUTODBA_TARGET_DB" /usr/lib/prometheus_postgres_exporter/postgres_exporter \
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

  # Start up Prometheus SQL Exporter
  pushd /usr/lib/prometheus_sql_exporter
  ./sql_exporter -config.data-source-name "$AUTODBA_TARGET_DB" &
  popd

  # Start up Prometheus RDS Exporter
  if [[ -n "$AWS_ACCESS_KEY_ID" && -n "$AWS_SECRET_ACCESS_KEY" && -n "$AWS_REGION" ]]; then
    AWS_ACCESS_KEY_ID="$AWS_ACCESS_KEY_ID" \
      AWS_SECRET_ACCESS_KEY="$AWS_SECRET_ACCESS_KEY" \
      /usr/lib/prometheus_rds_exporter/prometheus-rds-exporter -c /usr/lib/prometheus_rds_exporter/prometheus-rds-exporter.yaml --filter-instances "$AWS_RDS_INSTANCE" &
  else
    echo "One or more required AWS environment variables are missing or empty, so not running the RDS Exporter."
  fi
fi

# Start up Prometheus for initialization
/usr/bin/prometheus \
    --config.file=/etc/prometheus/prometheus.yml \
    --storage.tsdb.path=/prometheus \
    --web.console.templates=/etc/prometheus/consoles \
    --web.console.libraries=/etc/prometheus/console_libraries \
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
  mkdir -p /prometheus
  cp -r /home/autodba/restore_backup_uncompressed/home/autodba/backups/prometheus_snapshot_recent/* /prometheus/
  chown -R root:root /prometheus

  echo "Prometheus backup restored."

  # Start up Prometheus for real
  /usr/bin/prometheus \
      --config.file=/etc/prometheus/prometheus.yml \
      --storage.tsdb.path=/prometheus \
      --web.console.templates=/etc/prometheus/consoles \
      --web.console.libraries=/etc/prometheus/console_libraries \
      --web.enable-admin-api &
fi

# Start up bff
/usr/lib/bff/main -dbidentifier="$AWS_RDS_INSTANCE" -webappPath /home/autodba/src/webapp &
BFF_PID=$!

# Wait for a process to exit
wait -n -p EXITED_PID # wait for any job to exit
retcode=$? # store error code so we can propagate it to the container environment

if (( $EXITED_PID == $PROMETHEUS_PID ))
then
  echo "prometheus exited with return code $retcode - killing all jobs"
elif (( $EXITED_PID == $BFF_PID ))
  echo "Bff exited with return code $retcode - killing all jobs"
else
  echo "An unknown background process (with PID=$EXITED_PID) exited with return code $retcode - killing all jobs"
fi

kill $(jobs -p)

wait # wait for all children to exit -- this lets their logs make it out of the container environment
echo done

exit $retcode
