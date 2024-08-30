#!/bin/bash

# SPDX-License-Identifier: Apache-2.0

set -e

# Get the directory of the currently executing script
SOURCE_DIR=$(dirname "$(readlink -f "$0")")

cd $SOURCE_DIR/..

# Define version
VERSION="0.1.0"

# Define output directories
OUTPUT_DIR="$SOURCE_DIR/../release_output"
TAR_GZ_DIR="${OUTPUT_DIR}/tar.gz"
RPM_DIR="${OUTPUT_DIR}/rpm"
DEB_DIR="${OUTPUT_DIR}/deb"
SRC_DIR="${OUTPUT_DIR}/source"

# Cleanup previous builds
rm -rf "${OUTPUT_DIR}"
mkdir -p "${TAR_GZ_DIR}" "${RPM_DIR}" "${DEB_DIR}" "${SRC_DIR}"

# Build the binary for multiple architectures
echo "Building the project for multiple architectures..."

cd bff

# Build for x86_64
GOARCH=amd64 GOOS=linux go build -o ${OUTPUT_DIR}/autodba-bff-amd64 ./cmd/main.go

# Build for ARM64
GOARCH=arm64 GOOS=linux go build -o ${OUTPUT_DIR}/autodba-bff-arm64 ./cmd/main.go

cd ..

# Build the UI (Solid project)
echo "Building the UI..."
cd solid
npm install
npm run build
cd ..

# Prepare Prometheus and Exporters
echo "Including Prometheus and Exporters..."

# Download and prepare Prometheus and exporters for both architectures
PROM_VERSION="2.43.0"
EXPORTER_VERSION="0.15.0"
SQL_EXPORTER_VERSION="0.14.3"
RDS_EXPORTER_REPO="https://github.com/crystaldb/prometheus-rds-exporter.git"

for arch in amd64 arm64; do
    mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}/prometheus/${arch}"
    mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}/prometheus/${arch}/exporters"

    # Prometheus
    wget -qO- https://github.com/prometheus/prometheus/releases/download/v${PROM_VERSION}/prometheus-${PROM_VERSION}.linux-${arch}.tar.gz | tar -xzf - -C "${TAR_GZ_DIR}/autodba-${VERSION}/prometheus/${arch}" --strip-components=1

    # Postgres Exporter
    wget -qO- https://github.com/prometheus-community/postgres_exporter/releases/download/v${EXPORTER_VERSION}/postgres_exporter-${EXPORTER_VERSION}.linux-${arch}.tar.gz | tar -xzf - -C "${TAR_GZ_DIR}/autodba-${VERSION}/prometheus/${arch}/exporters" --strip-components=1

    # SQL Exporter
    wget -qO- https://github.com/burningalchemist/sql_exporter/releases/download/${SQL_EXPORTER_VERSION}/sql_exporter-${SQL_EXPORTER_VERSION}.linux-${arch}.tar.gz | tar -xzf - -C "${TAR_GZ_DIR}/autodba-${VERSION}/prometheus/${arch}/exporters" --strip-components=1

    # RDS Exporter (Build from source)
    rm -rf "/tmp/prometheus_rds_exporter_$arch"
    git clone "${RDS_EXPORTER_REPO}" "/tmp/prometheus_rds_exporter_$arch"
    cd /tmp/prometheus_rds_exporter_$arch
    GOARCH=${arch} GOOS=linux go build -o "${TAR_GZ_DIR}/autodba-${VERSION}/prometheus/${arch}/exporters/prometheus-rds-exporter"
    cd -
done

# Prepare for tar.gz package
echo "Creating tar.gz package..."
mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}/bin"
mkdir -p "${TAR_GZ_DIR}/autodba-${VERSION}/webapp"
cp -r ${OUTPUT_DIR}/autodba-bff-amd64 "${TAR_GZ_DIR}/autodba-${VERSION}/bin/autodba-bff-amd64"
cp -r ${OUTPUT_DIR}/autodba-bff-arm64 "${TAR_GZ_DIR}/autodba-${VERSION}/bin/autodba-bff-arm64"
cp -r solid/dist/* "${TAR_GZ_DIR}/autodba-${VERSION}/webapp/"
cp entrypoint.sh "${TAR_GZ_DIR}/autodba-${VERSION}/"

tar -czvf "${TAR_GZ_DIR}/autodba-${VERSION}.tar.gz" -C "${TAR_GZ_DIR}" "autodba-${VERSION}"

# Prepare RPM packages
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
        "${TAR_GZ_DIR}/autodba-${VERSION}/prometheus/${ARCH_SUFFIX}"=/usr/local/share/prometheus \
        "${TAR_GZ_DIR}/autodba-${VERSION}/entrypoint.sh"=/usr/local/bin/autodba-entrypoint.sh
done

# Prepare DEB packages
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
        "${TAR_GZ_DIR}/autodba-${VERSION}/prometheus/${arch}"=/usr/local/share/prometheus \
        "${TAR_GZ_DIR}/autodba-${VERSION}/entrypoint.sh"=/usr/local/bin/autodba-entrypoint.sh
done

# Package the source code
echo "Packaging source code..."
git archive --format=tar.gz --output="${SRC_DIR}/autodba-${VERSION}-source.tar.gz" HEAD

echo "Release build complete. Output located in ${OUTPUT_DIR}."
