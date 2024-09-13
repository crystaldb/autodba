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

# Function to check if fpm is installed
check_fpm() {
    if ! command -v fpm &> /dev/null; then
        echo "Warning: 'fpm' is not installed. Skipping package creation for .rpm and .deb."
        echo "To install fpm, run: 'sudo apt-get install ruby ruby-dev rubygems build-essential && gem install fpm'"
        return 1
    fi
    return 0
}

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

    mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/bin"
    mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/webapp"
    mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/config"
    
    cp -r ${OUTPUT_DIR}/autodba-bff-${arch} "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/bin/autodba-bff"
    cp -r ${OUTPUT_DIR}/config.json "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/config/config.json"
    cp -r solid/dist/* "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/webapp/"
    cp entrypoint.sh "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/"
done

# Function to create tar.gz package for each architecture
create_tar_gz() {
    for arch in amd64 arm64; do
        echo "Creating tar.gz package for ${arch}..."
        tar -czvf "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}.tar.gz" -C "${TAR_GZ_DIR}" "autodba-${VERSION}-${arch}"
    done
}

# Function to create RPM packages
# Function to create RPM packages
create_rpm() {
    echo "Creating RPM packages..."

    if check_fpm; then
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
                --prefix /usr/local/autodba \
                --package "${RPM_DIR}/" \
                "${TAR_GZ_DIR}/autodba-${VERSION}-${ARCH_SUFFIX}/bin/autodba-bff"=/usr/local/autodba/bin/autodba-bff \
                "${TAR_GZ_DIR}/autodba-${VERSION}-${ARCH_SUFFIX}/webapp"=/usr/local/autodba/share/webapp \
                "${TAR_GZ_DIR}/autodba-${VERSION}-${ARCH_SUFFIX}/exporters/postgres_exporter"=/usr/local/autodba/share/prometheus_exporters/postgres_exporter \
                "${TAR_GZ_DIR}/autodba-${VERSION}-${ARCH_SUFFIX}/exporters/sql_exporter/sql_exporter"=/usr/local/autodba/share/prometheus_exporters/sql_exporter/sql_exporter \
                "${TAR_GZ_DIR}/autodba-${VERSION}-${ARCH_SUFFIX}/exporters/rds_exporter"=/usr/local/autodba/share/prometheus_exporters/rds_exporter \
                "${TAR_GZ_DIR}/autodba-${VERSION}-${ARCH_SUFFIX}/monitor/prometheus/sql_exporter/sql_exporter.yml"=/usr/local/autodba/share/prometheus_exporters/sql_exporter/sql_exporter.yml \
                "${TAR_GZ_DIR}/autodba-${VERSION}-${ARCH_SUFFIX}/monitor/prometheus/rds_exporter"=/usr/local/autodba/share/prometheus_exporters/rds_exporter \
                "${TAR_GZ_DIR}/autodba-${VERSION}-${ARCH_SUFFIX}/monitor/prometheus/prometheus.yml"=/usr/local/autodba/config/prometheus/prometheus.yml \
                "${TAR_GZ_DIR}/autodba-${VERSION}-${ARCH_SUFFIX}/config/config.json"=/usr/local/autodba/config/autodba/config.json \
                "${TAR_GZ_DIR}/autodba-${VERSION}-${ARCH_SUFFIX}/entrypoint.sh"=/usr/local/autodba/bin/autodba-entrypoint.sh
        done
    fi
}

# Function to create DEB packages
create_deb() {
    echo "Creating DEB packages..."

    if check_fpm; then
        for arch in amd64 arm64; do
            fpm -s dir -t deb -n autodba -v "${VERSION}" \
                -a ${arch} \
                --description "AutoDBA is an AI agent for operating PostgreSQL databases." \
                --license "Apache-2.0" \
                --maintainer "CrystalDB Team <info@crystal.cloud>" \
                --url "https://www.crystaldb.cloud/" \
                --prefix /usr/local/autodba \
                --package "${DEB_DIR}/" \
                "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/bin/autodba-bff"=/usr/local/autodba/bin/autodba-bff \
                "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/webapp"=/usr/local/autodba/share/webapp \
                "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/exporters/postgres_exporter"=/usr/local/autodba/share/prometheus_exporters/postgres_exporter \
                "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/exporters/sql_exporter/sql_exporter"=/usr/local/autodba/share/prometheus_exporters/sql_exporter/sql_exporter \
                "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/exporters/rds_exporter"=/usr/local/autodba/share/prometheus_exporters/rds_exporter \
                "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/monitor/prometheus/sql_exporter/sql_exporter.yml"=/usr/local/autodba/share/prometheus_exporters/sql_exporter/sql_exporter.yml \
                "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/monitor/prometheus/rds_exporter"=/usr/local/autodba/share/prometheus_exporters/rds_exporter \
                "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/monitor/prometheus/prometheus.yml"=/usr/local/autodba/config/prometheus/prometheus.yml \
                "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/config/config.json"=/usr/local/autodba/config/autodba/config.json \
                "${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/entrypoint.sh"=/usr/local/autodba/bin/autodba-entrypoint.sh
        done
    fi
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
