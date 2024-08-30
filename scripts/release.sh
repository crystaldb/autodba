#!/bin/bash

# SPDX-Identifier: Apache-2.0

set -e

APP_NAME="autodba"
VERSION="1.0.0"
BUILD_DIR="/tmp/$APP_NAME-build"
DIST_DIR="$(pwd)/dist"
BINARY_PATH="$BUILD_DIR/$APP_NAME"
SOLID_DIR="/path/to/your/application/solid" # Adjust this path to the actual location of your solid directory

# URLs for the Prometheus and exporters
PROMETHEUS_URL="https://github.com/prometheus/prometheus/releases/download/v2.37.1/prometheus-2.37.1.linux-amd64.tar.gz"
POSTGRES_EXPORTER_URL="https://github.com/prometheus-community/postgres_exporter/releases/download/v0.15.0/postgres_exporter-0.15.0.linux-amd64.tar.gz"
SQL_EXPORTER_URL="https://github.com/burningalchemist/sql_exporter/releases/download/0.14.3/sql_exporter-0.14.3.linux-amd64.tar.gz"
RDS_EXPORTER_REPO="https://github.com/crystaldb/prometheus-rds-exporter.git"

# Create directories
mkdir -p $BUILD_DIR
mkdir -p $DIST_DIR

# Cleanup any previous build artifacts
rm -rf $BUILD_DIR/*
rm -rf $DIST_DIR/*

# Build the Go application
echo "Building Go application..."
cd /path/to/your/application/bff  # Replace with the correct path to your Go project
go mod download
go build -o $BINARY_PATH ./cmd/main.go

# Build the frontend (SolidJS)
echo "Building frontend..."
cd $SOLID_DIR
npm install
npm run build

# Copy the built frontend assets to the build directory
mkdir -p $BUILD_DIR/webapp
cp -r $SOLID_DIR/dist/* $BUILD_DIR/webapp/

# Download and extract Prometheus
echo "Downloading Prometheus..."
mkdir -p $BUILD_DIR/prometheus
wget -qO- $PROMETHEUS_URL | tar -xzf - -C $BUILD_DIR/prometheus --strip-components=1

# Download and extract Prometheus Postgres Exporter
echo "Downloading Prometheus Postgres Exporter..."
mkdir -p $BUILD_DIR/postgres_exporter
wget -qO- $POSTGRES_EXPORTER_URL | tar -xzf - -C $BUILD_DIR/postgres_exporter --strip-components=1

# Download and extract Prometheus SQL Exporter
echo "Downloading Prometheus SQL Exporter..."
mkdir -p $BUILD_DIR/sql_exporter
wget -qO- $SQL_EXPORTER_URL | tar -xzf - -C $BUILD_DIR/sql_exporter --strip-components=1

# Clone and build Prometheus RDS Exporter
echo "Cloning Prometheus RDS Exporter..."
mkdir -p $BUILD_DIR/rds_exporter
git clone $RDS_EXPORTER_REPO $BUILD_DIR/rds_exporter
cd $BUILD_DIR/rds_exporter
make build

# Create DEB package
echo "Creating DEB package..."
mkdir -p $BUILD_DIR/DEBIAN
cat <<EOF > $BUILD_DIR/DEBIAN/control
Package: $APP_NAME
Version: $VERSION
Section: base
Priority: optional
Architecture: amd64
Maintainer: Your Name <your.email@example.com>
Description: AutoDBA is an AI-powered PostgreSQL management agent.
EOF

# Prepare directories for DEB package
mkdir -p $BUILD_DIR/usr/local/bin
mkdir -p $BUILD_DIR/usr/local/share/$APP_NAME/webapp
mkdir -p $BUILD_DIR/usr/local/share/$APP_NAME/prometheus
mkdir -p $BUILD_DIR/usr/local/share/$APP_NAME/postgres_exporter
mkdir -p $BUILD_DIR/usr/local/share/$APP_NAME/sql_exporter
mkdir -p $BUILD_DIR/usr/local/share/$APP_NAME/rds_exporter

# Copy binaries and webapp
cp $BINARY_PATH $BUILD_DIR/usr/local/bin/$APP_NAME
cp -r $BUILD_DIR/webapp/* $BUILD_DIR/usr/local/share/$APP_NAME/webapp/
cp -r $BUILD_DIR/prometheus/* $BUILD_DIR/usr/local/share/$APP_NAME/prometheus/
cp -r $BUILD_DIR/postgres_exporter/* $BUILD_DIR/usr/local/share/$APP_NAME/postgres_exporter/
cp -r $BUILD_DIR/sql_exporter/* $BUILD_DIR/usr/local/share/$APP_NAME/sql_exporter/
cp -r $BUILD_DIR/rds_exporter/bin/prometheus-rds-exporter $BUILD_DIR/usr/local/share/$APP_NAME/rds_exporter/

dpkg-deb --build $BUILD_DIR $DIST_DIR/${APP_NAME}_${VERSION}_amd64.deb

# Create RPM package
echo "Creating RPM package..."
mkdir -p $BUILD_DIR/RPMS
cat <<EOF > $BUILD_DIR/$APP_NAME.spec
Name: $APP_NAME
Version: $VERSION
Release: 1%{?dist}
Summary: AutoDBA is an AI-powered PostgreSQL management agent.

License: Apache-2.0
URL: http://example.com
Packager: Your Name <your.email@example.com>
Requires: curl, wget, jq, nodejs, npm, git, sqlite
BuildArch: x86_64

%description
AutoDBA is an AI-powered PostgreSQL management agent.

%prep
%setup -q

%build

%install
rm -rf %{buildroot}
mkdir -p %{buildroot}/usr/local/bin
mkdir -p %{buildroot}/usr/local/share/$APP_NAME/webapp
mkdir -p %{buildroot}/usr/local/share/$APP_NAME/prometheus
mkdir -p %{buildroot}/usr/local/share/$APP_NAME/postgres_exporter
mkdir -p %{buildroot}/usr/local/share/$APP_NAME/sql_exporter
mkdir -p %{buildroot}/usr/local/share/$APP_NAME/rds_exporter

cp $BINARY_PATH %{buildroot}/usr/local/bin/$APP_NAME
cp -r $BUILD_DIR/webapp/* %{buildroot}/usr/local/share/$APP_NAME/webapp/
cp -r $BUILD_DIR/prometheus/* %{buildroot}/usr/local/share/$APP_NAME/prometheus/
cp -r $BUILD_DIR/postgres_exporter/* %{buildroot}/usr/local/share/$APP_NAME/postgres_exporter/
cp -r $BUILD_DIR/sql_exporter/* %{buildroot}/usr/local/share/$APP_NAME/sql_exporter/
cp -r $BUILD_DIR/rds_exporter/bin/prometheus-rds-exporter %{buildroot}/usr/local/share/$APP_NAME/rds_exporter/

%files
/usr/local/bin/$APP_NAME
/usr/local/share/$APP_NAME/webapp/*
/usr/local/share/$APP_NAME/prometheus/*
/usr/local/share/$APP_NAME/postgres_exporter/*
/usr/local/share/$APP_NAME/sql_exporter/*
/usr/local/share/$APP_NAME/rds_exporter/*

%post
echo "Installation complete."

%preun
echo "Preparing to uninstall."

%postun
echo "Uninstallation complete."

%changelog
* $(date +'%a %b %d %Y') Your Name <your.email@example.com> - $VERSION-1
- Initial RPM release.
EOF

rpmbuild -bb --buildroot=$BUILD_DIR/$APP_NAME $BUILD_DIR/$APP_NAME.spec
cp ~/rpmbuild/RPMS/x86_64/${APP_NAME}-${VERSION}-1.el7.x86_64.rpm $DIST_DIR/

echo "Release packages created in $DIST_DIR:"
ls -lh $DIST_DIR

# Cleanup build directory
rm -rf $BUILD_DIR
