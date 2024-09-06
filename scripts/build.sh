#!/bin/bash

# SPDX-Identifier: Apache-2.0

set -e

# Get the directory of the currently executing script
SOURCE_DIR=$(dirname "$(readlink -f "$0")")

cd $SOURCE_DIR/..

# Define version
VERSION="0.1.0"

# Define output directories
OUTPUT_DIR="$SOURCE_DIR/../build_output"
TAR_GZ_DIR="${OUTPUT_DIR}/tar.gz"
RPM_DIR="${OUTPUT_DIR}/rpm"
DEB_DIR="${OUTPUT_DIR}/deb"
SRC_DIR="${OUTPUT_DIR}/source"

# Cleanup previous builds
rm -rf "${OUTPUT_DIR}"
mkdir -p "${TAR_GZ_DIR}" "${RPM_DIR}" "${DEB_DIR}" "${SRC_DIR}"

# Parse command-line argument for target
TARGET=$1

if [ -z "$TARGET" ]; then
  TARGET="all"
fi

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
    # Create separate directories for each exporter
    mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}/exporters/${arch}/postgres_exporter"
    mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}/exporters/${arch}/sql_exporter"
    mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}/exporters/${arch}/rds_exporter"

    # Postgres Exporter
    wget -qO- https://github.com/prometheus-community/postgres_exporter/releases/download/v${EXPORTER_VERSION}/postgres_exporter-${EXPORTER_VERSION}.linux-${arch}.tar.gz | tar -xzf - -C "${TAR_GZ_DIR}/autodba-${VERSION}/exporters/${arch}/postgres_exporter" --strip-components=1

    # SQL Exporter
    wget -qO- https://github.com/burningalchemist/sql_exporter/releases/download/${SQL_EXPORTER_VERSION}/sql_exporter-${SQL_EXPORTER_VERSION}.linux-${arch}.tar.gz | tar -xzf - -C "${TAR_GZ_DIR}/autodba-${VERSION}/exporters/${arch}/sql_exporter" --strip-components=1
    rm "${TAR_GZ_DIR}/autodba-${VERSION}/exporters/${arch}/sql_exporter/mssql_standard.collector.yml"

    # RDS Exporter (Build from source)
    rm -rf "/tmp/prometheus_rds_exporter"
    git clone "${RDS_EXPORTER_REPO}" "/tmp/prometheus_rds_exporter"
    cd /tmp/prometheus_rds_exporter
    GOARCH=${arch} GOOS=linux go build -o "${TAR_GZ_DIR}/autodba-${VERSION}/exporters/${arch}/rds_exporter/prometheus-rds-exporter"
    cd -
done

# Copy configuration files and monitor setup
echo "Copying configuration and monitor setup files..."
mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}/monitor/prometheus/sql_exporter"
mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}/monitor/prometheus/rds_exporter"
mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}/monitor/prometheus"

cp -r monitor/prometheus/sql_exporter/* "${TAR_GZ_DIR}/autodba-${VERSION}/monitor/prometheus/sql_exporter/"
cp -r monitor/prometheus/rds_exporter/* "${TAR_GZ_DIR}/autodba-${VERSION}/monitor/prometheus/rds_exporter/"
cp monitor/prometheus/prometheus.yml "${TAR_GZ_DIR}/autodba-${VERSION}/monitor/prometheus/"

# Function to create tar.gz package
create_tar_gz() {
    echo "Creating tar.gz package..."
    mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}/bin"
    mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}/webapp"
    mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}/config"
    cp -r ${OUTPUT_DIR}/autodba-bff-amd64 "${TAR_GZ_DIR}/autodba-${VERSION}/bin/autodba-bff-amd64"
    cp -r ${OUTPUT_DIR}/autodba-bff-arm64 "${TAR_GZ_DIR}/autodba-${VERSION}/bin/autodba-bff-arm64"
    cp -r ${OUTPUT_DIR}/config.json "${TAR_GZ_DIR}/autodba-${VERSION}/config/config.json"
    cp -r solid/dist/* "${TAR_GZ_DIR}/autodba-${VERSION}/webapp/"
    cp entrypoint.sh "${TAR_GZ_DIR}/autodba-${VERSION}/"

    tar -czvf "${TAR_GZ_DIR}/autodba-${VERSION}.tar.gz" -C "${TAR_GZ_DIR}" "autodba-${VERSION}"
}

# Function to create RPM packages
create_rpm() {
    echo "Creating RPM packages..."

    for arch in x86_64 aarch64; do
        if [ "$arch" == "x86_64" ]; then
            ARCH_SUFFIX="amd64"
        else
            ARCH_SUFFIX="arm64"
        fi

        fpm -s dir -t rpm -n autodba -v "${VERSION}" \
            -a ${arch} \
            --description "AutoDBA is an AI agent for operating PostgreSQL databases." \
            --license "Apache-2.0" \
            --maintainer "CrystalDB Team <info@crystal.cloud>" \
            --url "https://www.crystaldb.cloud/" \
            --prefix /usr/local/bin \
            --package "${RPM_DIR}/" \
            "${TAR_GZ_DIR}/autodba-${VERSION}/bin/autodba-bff-${ARCH_SUFFIX}"=/usr/local/bin/autodba-bff \
            "${TAR_GZ_DIR}/autodba-${VERSION}/webapp"=/usr/local/share/autodba/webapp \
            "${TAR_GZ_DIR}/autodba-${VERSION}/exporters/${ARCH_SUFFIX}/postgres_exporter"=/usr/local/share/prometheus_exporters/postgres_exporter \
            "${TAR_GZ_DIR}/autodba-${VERSION}/exporters/${ARCH_SUFFIX}/sql_exporter"=/usr/local/share/prometheus_exporters/sql_exporter \
            "${TAR_GZ_DIR}/autodba-${VERSION}/exporters/${ARCH_SUFFIX}/rds_exporter"=/usr/local/share/prometheus_exporters/rds_exporter \
            "${TAR_GZ_DIR}/autodba-${VERSION}/monitor/prometheus/sql_exporter"=/usr/local/share/prometheus_exporters/sql_exporter \
            "${TAR_GZ_DIR}/autodba-${VERSION}/monitor/prometheus/rds_exporter"=/usr/local/share/prometheus_exporters/rds_exporter \
            "${TAR_GZ_DIR}/autodba-${VERSION}/monitor/prometheus/prometheus.yml"=/etc/prometheus/prometheus.yml \
            "${TAR_GZ_DIR}/autodba-${VERSION}/config/config.json"=/etc/autodba/config.json \
            "${TAR_GZ_DIR}/autodba-${VERSION}/entrypoint.sh"=/usr/local/bin/autodba-entrypoint.sh
    done
}

# Function to create DEB packages
create_deb() {
    echo "Creating DEB packages..."

    for arch in amd64 arm64; do
        fpm -s dir -t deb -n autodba -v "${VERSION}" \
            -a ${arch} \
            --description "AutoDBA is an AI agent for operating PostgreSQL databases." \
            --license "Apache-2.0" \
            --maintainer "CrystalDB Team <info@crystal.cloud>" \
            --url "https://www.crystaldb.cloud/" \
            --prefix /usr/local/bin \
            --package "${DEB_DIR}/" \
            "${TAR_GZ_DIR}/autodba-${VERSION}/bin/autodba-bff-${arch}"=/usr/local/bin/autodba-bff \
            "${TAR_GZ_DIR}/autodba-${VERSION}/webapp"=/usr/local/share/autodba/webapp \
            "${TAR_GZ_DIR}/autodba-${VERSION}/exporters/${arch}/postgres_exporter"=/usr/local/share/prometheus_exporters/postgres_exporter \
            "${TAR_GZ_DIR}/autodba-${VERSION}/exporters/${arch}/sql_exporter"=/usr/local/share/prometheus_exporters/sql_exporter \
            "${TAR_GZ_DIR}/autodba-${VERSION}/exporters/${arch}/rds_exporter"=/usr/local/share/prometheus_exporters/rds_exporter \
            "${TAR_GZ_DIR}/autodba-${VERSION}/monitor/prometheus/sql_exporter"=/usr/local/share/prometheus_exporters/sql_exporter \
            "${TAR_GZ_DIR}/autodba-${VERSION}/monitor/prometheus/rds_exporter"=/usr/local/share/prometheus_exporters/rds_exporter \
            "${TAR_GZ_DIR}/autodba-${VERSION}/monitor/prometheus/prometheus.yml"=/etc/prometheus/prometheus.yml \
            "${TAR_GZ_DIR}/autodba-${VERSION}/config/config.json"=/etc/autodba/config.json \
            "${TAR_GZ_DIR}/autodba-${VERSION}/entrypoint.sh"=/usr/local/bin/autodba-entrypoint.sh
    done
}

# Function to package the source code
package_source() {
    echo "Packaging source code..."
    git archive --format=tar.gz --output="${SRC_DIR}/autodba-${VERSION}-source.tar.gz" HEAD
}

# Print the selected target
echo "Building target: $TARGET"

# Run the appropriate target(s)
case "$TARGET" in
  tar.gz)
    create_tar_gz
    ;;
  rpm)
    create_rpm
    ;;
  deb)
    create_deb
    ;;
  source)
    package_source
    ;;
  all)
    create_tar_gz
    create_rpm
    create_deb
    package_source
    ;;
  *)
    echo "Unsupported target: $TARGET"
    exit 1
    ;;
esac

echo "Release build complete. Output located in ${OUTPUT_DIR}."
