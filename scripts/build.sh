#!/bin/bash

# SPDX-Identifier: Apache-2.0

set -e

# Get the directory of the currently executing script
SOURCE_DIR=$(dirname "$(readlink -f "$0")")

cd $SOURCE_DIR/..

# Get the latest git tag as the version
VERSION=$(git describe --tags --abbrev=0 2>/dev/null)
if [ -z "$VERSION" ]; then
  echo "Warning: No git tags found. Please create a tag before building. Using default version v0.1.0."
  VERSION="v0.1.0"
fi
# Remove the leading 'v' if it exists
VERSION=${VERSION#v}

# Define Prometheus version
PROMETHEUS_VERSION="2.42.0"

# Define output directories
OUTPUT_DIR="$SOURCE_DIR/../build_output"
TAR_GZ_DIR="${OUTPUT_DIR}/tar.gz"

# Cleanup previous builds
rm -rf "${OUTPUT_DIR}"
mkdir -p "${TAR_GZ_DIR}"

# Build the binary for multiple architectures
echo "Building the project for multiple architectures..."

cd bff

# Build for x86_64
GOARCH=amd64 GOOS=linux go build -o ${OUTPUT_DIR}/autodba-bff-amd64 ./cmd/main.go

# Build for ARM64
GOARCH=arm64 GOOS=linux go build -o ${OUTPUT_DIR}/autodba-bff-arm64 ./cmd/main.go

# Copy the config.json
cp config.json ${OUTPUT_DIR}/config.json

cd ..

# Build the UI (Solid project)
echo "Building the UI..."
cd solid
npm install
npm run build
cd ..

# Prepare Prometheus exporters
echo "Including Prometheus exporters..."

EXPORTER_VERSION="0.15.0"
SQL_EXPORTER_VERSION="0.14.3"
RDS_EXPORTER_REPO="https://github.com/crystaldb/prometheus-rds-exporter.git"

for arch in amd64 arm64; do
    echo "Downloading Prometheus tarball for ${arch}..."
    PROMETHEUS_DIR="${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/prometheus"
    mkdir -p "${PROMETHEUS_DIR}"
    wget -qO "${PROMETHEUS_DIR}/prometheus-${PROMETHEUS_VERSION}.linux-${arch}.tar.gz" "https://github.com/prometheus/prometheus/releases/download/v${PROMETHEUS_VERSION}/prometheus-${PROMETHEUS_VERSION}.linux-${arch}.tar.gz"

    # Create separate directories for each exporter
    mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/exporters/postgres_exporter"
    mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/exporters/sql_exporter"
    mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/exporters/rds_exporter"

    # Postgres Exporter
    wget -qO- https://github.com/prometheus-community/postgres_exporter/releases/download/v${EXPORTER_VERSION}/postgres_exporter-${EXPORTER_VERSION}.linux-${arch}.tar.gz | tar -xzf - -C "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/exporters/postgres_exporter" --strip-components=1

    # SQL Exporter
    wget -qO- https://github.com/burningalchemist/sql_exporter/releases/download/${SQL_EXPORTER_VERSION}/sql_exporter-${SQL_EXPORTER_VERSION}.linux-${arch}.tar.gz | tar -xzf - -C "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/exporters/sql_exporter" --strip-components=1
    rm "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/exporters/sql_exporter/mssql_standard.collector.yml"

    # RDS Exporter (Build from source)
    rm -rf "/tmp/prometheus_rds_exporter"
    git clone "${RDS_EXPORTER_REPO}" "/tmp/prometheus_rds_exporter"
    cd /tmp/prometheus_rds_exporter
    GOARCH=${arch} GOOS=linux go build -o "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/exporters/rds_exporter/prometheus-rds-exporter"
    cd -

    # Copy configuration files and monitor setup
    echo "Copying configuration and monitor setup files..."
    mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/monitor/prometheus/sql_exporter"
    mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/monitor/prometheus/rds_exporter"
    mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/monitor/prometheus"

    cp -r monitor/prometheus/sql_exporter/* "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/monitor/prometheus/sql_exporter/"
    cp -r monitor/prometheus/rds_exporter/* "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/monitor/prometheus/rds_exporter/"
    cp monitor/prometheus/prometheus.yml "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/monitor/prometheus/"

    # Prepare directories for install
    mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/bin"
    mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/webapp"
    mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/config"
    
    cp -r ${OUTPUT_DIR}/autodba-bff-${arch} "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/bin/autodba-bff"
    cp -r ${OUTPUT_DIR}/config.json "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/config/config.json"
    cp -r solid/dist/* "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/webapp/"
    cp entrypoint.sh "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/"
    
    # Copy the install.sh script into the root of the tarball
    cp install.sh "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/"
done

# Function to create tar.gz package for each architecture
create_tar_gz() {
    for arch in amd64 arm64; do
        echo "Creating tar.gz package for ${arch}..."
        tar -czvf "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}.tar.gz" -C "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}" .
    done
}

# Call the function to create the tar.gz
create_tar_gz
