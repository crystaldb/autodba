#!/bin/bash

# Default value for SUFFIX
SUFFIX=$(date +%Y%m%d%H%M%S)

# Parse command line arguments for SUFFIX
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --suffix) SUFFIX="$2"; shift ;;
        -h|--help) 
            echo "Usage: $0 [--suffix <suffix>]"
            echo "  --suffix <suffix>  Specify the suffix for backup files (default: current timestamp)"
            exit 0
            ;;
        *) 
            echo "Unknown parameter passed: $1"
            echo "Usage: $0 [--suffix <suffix>]"
            exit 1 
            ;;
    esac
    shift
done

BACKUP_DIR="/home/autodba/backups"

# Ensure backup directory exists
mkdir -p $BACKUP_DIR

# Backup PostgreSQL
PG_BACKUP_FILE="$BACKUP_DIR/postgresql_backup_${SUFFIX}.sql"
pg_dump -U $POSTGRES_USER -h $POSTGRES_HOST -p $POSTGRES_PORT $POSTGRES_DB > $PG_BACKUP_FILE

# Create Prometheus snapshot
SNAPSHOT_NAME=$(curl -XPOST http://localhost:9090/api/v1/admin/tsdb/snapshot | jq -r '.data.name')
if [ -z "$SNAPSHOT_NAME" ]; then
    echo "Failed to create Prometheus snapshot"
    exit 1
fi

# Backup Prometheus snapshot
PROMETHEUS_SNAPSHOT_DIR="/prometheus/snapshots/$SNAPSHOT_NAME"
PROMETHEUS_BACKUP_DIR="$BACKUP_DIR/prometheus_snapshot_${SUFFIX}"
mkdir -p $PROMETHEUS_BACKUP_DIR
cp -r $PROMETHEUS_SNAPSHOT_DIR/* $PROMETHEUS_BACKUP_DIR

echo "Backups completed:"
echo "PostgreSQL: $PG_BACKUP_FILE"
echo "Prometheus: $PROMETHEUS_BACKUP_DIR"
