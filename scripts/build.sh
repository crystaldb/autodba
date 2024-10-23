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

TMP_DIR="/tmp"

for arch in amd64 arm64; do
    # Define paths relative to PARENT_DIR
    PARENT_DIR="${TAR_GZ_DIR}/autodba-${VERSION}-${arch}/autodba-${VERSION}"
    INSTALL_DIR="$PARENT_DIR/bin"
    WEBAPP_DIR="$PARENT_DIR/share/webapp"
    PROMETHEUS_CONFIG_DIR="$PARENT_DIR/config/prometheus"
    AUTODBA_CONFIG_DIR="$PARENT_DIR/config/autodba"
    PROMETHEUS_INSTALL_DIR="$PARENT_DIR/prometheus"
    COLLECTOR_DIR="${PARENT_DIR}/share/collector"
    COLLECTOR_API_SERVER_DIR="${PARENT_DIR}/share/collector_api_server"

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

    # Copy prometheus setup
    echo "Copying prometheus setup files..."
    mkdir -p "${PROMETHEUS_CONFIG_DIR}"
    cp prometheus/prometheus.yml "${PROMETHEUS_CONFIG_DIR}/prometheus.yml"

    # Build collector
    PROTOC_ARCH_SUFFIX="x86_64" # We only build for x86_64, as we're going to run it on x86_64 and use its output at build time
    echo "Building collector..."
    mkdir -p "${COLLECTOR_DIR}"
    git clone --recurse-submodules https://github.com/crystaldb/collector.git "${COLLECTOR_DIR}"
    cd "${COLLECTOR_DIR}"
    wget https://github.com/protocolbuffers/protobuf/releases/download/v3.14.0/protoc-3.14.0-linux-${PROTOC_ARCH_SUFFIX}.zip
    unzip protoc-3.14.0-linux-${PROTOC_ARCH_SUFFIX}.zip -d protoc
    make build
    mv pganalyze-collector collector
    mv pganalyze-collector-helper collector-helper
    mv pganalyze-collector-setup collector-setup
    cd -

    echo "Building collector-api-server..."
    mkdir -p "${COLLECTOR_API_SERVER_DIR}"
    cp -r collector-api/* "${COLLECTOR_API_SERVER_DIR}/"
    cd "${COLLECTOR_API_SERVER_DIR}"
    go build -o collector-api-server ./cmd/server/main.go
    cd -

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
