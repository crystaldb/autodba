#!/bin/bash

# SPDX-Identifier: Apache-2.0

set -e

APP_NAME="autodba"
DEB_PACKAGE_PATH=""
RPM_PACKAGE_PATH=""
INSTALL_DIR="/usr/local/bin"
WEBAPP_DIR="/usr/local/share/$APP_NAME/webapp"
PROMETHEUS_DIR="/usr/local/share/$APP_NAME/prometheus"
POSTGRES_EXPORTER_DIR="/usr/local/share/$APP_NAME/postgres_exporter"
SQL_EXPORTER_DIR="/usr/local/share/$APP_NAME/sql_exporter"
RDS_EXPORTER_DIR="/usr/local/share/$APP_NAME/rds_exporter"

# Initialize environment variables with empty values
AUTODBA_TARGET_DB=""
AWS_RDS_INSTANCE=""
AWS_ACCESS_KEY_ID=""
AWS_SECRET_ACCESS_KEY=""
AWS_REGION=""

# Function to install the package
install_package() {
    if [ -n "$DEB_PACKAGE_PATH" ]; then
        echo "Installing DEB package..."
        sudo dpkg -i "$DEB_PACKAGE_PATH"
    elif [ -n "$RPM_PACKAGE_PATH" ]; then
        echo "Installing RPM package..."
        sudo rpm -ivh "$RPM_PACKAGE_PATH"
    else
        echo "No package specified for installation."
        exit 1
    fi
}

# Function to create systemd service files
create_systemd_service_files() {
    # AutoDBA main service
    echo "Creating AutoDBA main service file..."
    sudo tee /etc/systemd/system/$APP_NAME.service > /dev/null <<EOF
[Unit]
Description=AutoDBA Service
After=network.target

[Service]
ExecStart=$INSTALL_DIR/$APP_NAME --db-url $AUTODBA_TARGET_DB --webapp-dir $WEBAPP_DIR
WorkingDirectory=$INSTALL_DIR
Restart=always
Environment="AUTODBA_TARGET_DB=$AUTODBA_TARGET_DB"
Environment="AWS_RDS_INSTANCE=$AWS_RDS_INSTANCE"
Environment="AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID"
Environment="AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY"
Environment="AWS_REGION=$AWS_REGION"

[Install]
WantedBy=multi-user.target
EOF

    # Prometheus service
    echo "Creating Prometheus service file..."
    sudo tee /etc/systemd/system/prometheus.service > /dev/null <<EOF
[Unit]
Description=Prometheus Monitoring System
After=network.target

[Service]
ExecStart=$PROMETHEUS_DIR/prometheus --config.file=$PROMETHEUS_DIR/prometheus.yml --storage.tsdb.path=/var/lib/prometheus
WorkingDirectory=$PROMETHEUS_DIR
Restart=always
User=prometheus

[Install]
WantedBy=multi-user.target
EOF

    # Prometheus Postgres Exporter service
    echo "Creating Prometheus Postgres Exporter service file..."
    sudo tee /etc/systemd/system/postgres_exporter.service > /dev/null <<EOF
[Unit]
Description=Prometheus PostgreSQL Exporter
After=network.target

[Service]
Environment="DATA_SOURCE_NAME=$AUTODBA_TARGET_DB"
ExecStart=$POSTGRES_EXPORTER_DIR/postgres_exporter
WorkingDirectory=$POSTGRES_EXPORTER_DIR
Restart=always
User=prometheus

[Install]
WantedBy=multi-user.target
EOF

    # Prometheus SQL Exporter service
    echo "Creating Prometheus SQL Exporter service file..."
    sudo tee /etc/systemd/system/sql_exporter.service > /dev/null <<EOF
[Unit]
Description=Prometheus SQL Exporter
After=network.target

[Service]
ExecStart=$SQL_EXPORTER_DIR/sql_exporter --config.file=$SQL_EXPORTER_DIR/sql_exporter.yml
WorkingDirectory=$SQL_EXPORTER_DIR
Restart=always
User=prometheus

[Install]
WantedBy=multi-user.target
EOF

    # Prometheus RDS Exporter service
    echo "Creating Prometheus RDS Exporter service file..."
    sudo tee /etc/systemd/system/rds_exporter.service > /dev/null <<EOF
[Unit]
Description=Prometheus RDS Exporter
After=network.target

[Service]
Environment="AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID"
Environment="AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY"
Environment="AWS_REGION=$AWS_REGION"
ExecStart=$RDS_EXPORTER_DIR/prometheus-rds-exporter -c $RDS_EXPORTER_DIR/prometheus-rds-exporter.yaml --filter-instances "$AWS_RDS_INSTANCE"
WorkingDirectory=$RDS_EXPORTER_DIR
Restart=always
User=prometheus

[Install]
WantedBy=multi-user.target
EOF
}

# Function to start all services
start_services() {
    echo "Starting all services..."
    sudo systemctl daemon-reload
    sudo systemctl enable $APP_NAME prometheus postgres_exporter sql_exporter rds_exporter
    sudo systemctl start $APP_NAME prometheus postgres_exporter sql_exporter rds_exporter
}

# Function to display usage information
usage() {
    echo "Usage: $0 --deb <path/to/autodba.deb> --rpm <path/to/autodba.rpm> --db-url <TARGET_DATABASE_URL> --rds-instance <AWS_RDS_INSTANCE> --aws-access-key <AWS_ACCESS_KEY_ID> --aws-secret-key <AWS_SECRET_ACCESS_KEY> --aws-region <AWS_REGION>"
    exit 1
}

# Parse command-line arguments
while [[ "$#" -gt 0 ]]; do
    case $1 in
        --deb)
            if [[ -n "$2" ]] && [[ ${2:0:1} != "-" ]]; then
                DEB_PACKAGE_PATH="$2"
                shift 2
            else
                echo "Error: Argument for $1 is missing" >&2
                usage
            fi
            ;;
        --rpm)
            if [[ -n "$2" ]] && [[ ${2:0:1} != "-" ]]; then
                RPM_PACKAGE_PATH="$2"
                shift 2
            else
                echo "Error: Argument for $1 is missing" >&2
                usage
            fi
            ;;
        --db-url)
            if [[ -n "$2" ]] && [[ ${2:0:1} != "-" ]]; then
                AUTODBA_TARGET_DB="$2"
                shift 2
            else
                echo "Error: Argument for $1 is missing" >&2
                usage
            fi
            ;;
        --rds-instance)
            if [[ -n "$2" ]] && [[ ${2:0:1} != "-" ]]; then
                AWS_RDS_INSTANCE="$2"
                shift 2
            else
                echo "Error: Argument for $1 is missing" >&2
                usage
            fi
            ;;
        --aws-access-key)
            if [[ -n "$2" ]] && [[ ${2:0:1} != "-" ]]; then
                AWS_ACCESS_KEY_ID="$2"
                shift 2
            else
                echo "Error: Argument for $1 is missing" >&2
                usage
            fi
            ;;
        --aws-secret-key)
            if [[ -n "$2" ]] && [[ ${2:0:1} != "-" ]]; then
                AWS_SECRET_ACCESS_KEY="$2"
                shift 2
            else
                echo "Error: Argument for $1 is missing" >&2
                usage
            fi
            ;;
        --aws-region)
            if [[ -n "$2" ]] && [[ ${2:0:1} != "-" ]]; then
                AWS_REGION="$2"
                shift 2
            else
                echo "Error: Argument for $1 is missing" >&2
                usage
            fi
            ;;
        *)
            echo "Invalid argument: $1" >&2
            usage
            ;;
    esac
done

# Validate required arguments
if [[ -z "$DEB_PACKAGE_PATH" && -z "$RPM_PACKAGE_PATH" ]]; then
    echo "Error: No package specified."
    usage
fi

if [[ -z "$AUTODBA_TARGET_DB" ]]; then
    echo "Error: --db-url is required."
    usage
fi

# Execute functions
install_package
create_systemd_service_files
start_services

echo "Installation completed successfully."
echo "Use 'sudo systemctl status $APP_NAME' to check the service status."
