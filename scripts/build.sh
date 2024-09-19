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
TMP_DIR="/tmp"

for arch in amd64 arm64; do
    # Define paths relative to PARENT_DIR
    PARENT_DIR="${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/autodba-${VERSION}"
    INSTALL_DIR="$PARENT_DIR/bin"
    WEBAPP_DIR="$PARENT_DIR/share/webapp"
    EXPORTER_DIR="$PARENT_DIR/share/prometheus_exporters"
    PROMETHEUS_CONFIG_DIR="$PARENT_DIR/config/prometheus"
    AUTODBA_CONFIG_DIR="$PARENT_DIR/config/autodba"
    PROMETHEUS_INSTALL_DIR="$PARENT_DIR/prometheus"

    echo "Downloading Prometheus tarball for ${arch}..."
    # Prepare clean
    rm -rf $TMP_DIR/prometheus-*
    mkdir -p "${PROMETHEUS_INSTALL_DIR}"
    wget -qO- https://github.com/prometheus/prometheus/releases/download/v${PROMETHEUS_VERSION}/prometheus-${PROMETHEUS_VERSION}.linux-${arch}.tar.gz | tar -xzf - -C $TMP_DIR/
    cp $TMP_DIR/prometheus-${PROMETHEUS_VERSION}.linux-${arch}/prometheus "${PROMETHEUS_INSTALL_DIR}/"
    cp $TMP_DIR/prometheus-${PROMETHEUS_VERSION}.linux-${arch}/promtool "${PROMETHEUS_INSTALL_DIR}/"
    mkdir -p ${PROMETHEUS_CONFIG_DIR}
    cp -r $TMP_DIR/prometheus-${PROMETHEUS_VERSION}.linux-${arch}/consoles ${PROMETHEUS_CONFIG_DIR}/
    cp -r $TMP_DIR/prometheus-${PROMETHEUS_VERSION}.linux-${arch}/console_libraries ${PROMETHEUS_CONFIG_DIR}/
    # Cleanup
    rm -rf $TMP_DIR/prometheus-*

    # Create separate directories for each exporter
    mkdir -p "${EXPORTER_DIR}/postgres_exporter"
    mkdir -p "${EXPORTER_DIR}/sql_exporter"
    mkdir -p "${EXPORTER_DIR}/rds_exporter"

    # Postgres Exporter
    wget -qO- https://github.com/prometheus-community/postgres_exporter/releases/download/v${EXPORTER_VERSION}/postgres_exporter-${EXPORTER_VERSION}.linux-${arch}.tar.gz | tar -xzf - -C "${EXPORTER_DIR}/postgres_exporter" --strip-components=1

    # SQL Exporter
    wget -qO- https://github.com/burningalchemist/sql_exporter/releases/download/${SQL_EXPORTER_VERSION}/sql_exporter-${SQL_EXPORTER_VERSION}.linux-${arch}.tar.gz | tar -xzf - -C "${EXPORTER_DIR}/sql_exporter" --strip-components=1
    rm "${EXPORTER_DIR}/sql_exporter/mssql_standard.collector.yml"

    # RDS Exporter (Build from source)
    # Prepare clean
    rm -rf "/tmp/prometheus_rds_exporter"
    git clone "${RDS_EXPORTER_REPO}" "/tmp/prometheus_rds_exporter"
    cd /tmp/prometheus_rds_exporter
    GOARCH=${arch} GOOS=linux go build -o "${EXPORTER_DIR}/rds_exporter/prometheus-rds-exporter"
    # Cleanup
    rm -rf "/tmp/prometheus_rds_exporter"
    cd -

    # Copy configuration files and monitor setup
    echo "Copying configuration and monitor setup files..."
    mkdir -p "${EXPORTER_DIR}/sql_exporter"
    mkdir -p "${EXPORTER_DIR}/rds_exporter"
    mkdir -p "${PROMETHEUS_CONFIG_DIR}"

    cp -r monitor/prometheus/sql_exporter/* "${EXPORTER_DIR}/sql_exporter/"
    cp -r monitor/prometheus/rds_exporter/* "${EXPORTER_DIR}/rds_exporter/"
    cp monitor/prometheus/prometheus.yml "${PROMETHEUS_CONFIG_DIR}/prometheus.yml"

    # Prepare directories for install
    mkdir -p "${INSTALL_DIR}"
    mkdir -p "${WEBAPP_DIR}"
    mkdir -p "${AUTODBA_CONFIG_DIR}"
    
    cp -r ${OUTPUT_DIR}/autodba-bff-${arch} "${INSTALL_DIR}/autodba-bff"
    cp -r ${OUTPUT_DIR}/config.json "${AUTODBA_CONFIG_DIR}/config.json"
    cp -r solid/dist/* "${WEBAPP_DIR}"
    cp entrypoint.sh "${INSTALL_DIR}/autodba-entrypoint.sh"
    chmod +x "${INSTALL_DIR}/autodba-entrypoint.sh"
    
    # Copy the `install.sh` and `uninstall.sh` scripts into the root of the tarball
    cp scripts/install.sh "${PARENT_DIR}/"
    cp scripts/uninstall.sh "${PARENT_DIR}/"
    cp scripts/Makefile "${PARENT_DIR}/"
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

echo "Release build complete. Output located in ${OUTPUT_DIR}."
