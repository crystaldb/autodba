# SPDX-License-Identifier: Apache-2.0

FROM debian:bookworm-slim AS base

RUN addgroup --system autodba && adduser --system --group autodba --home /home/autodba --shell /bin/bash

RUN apt-get update
RUN apt-get install -y --no-install-recommends \
    curl            \
    jq              \
    nodejs          \
    npm             \
    procps          \
    wget


USER autodba
RUN mkdir -p /home/autodba/src
WORKDIR /home/autodba/src

FROM base AS solid_builder
# Build the solid project
WORKDIR /home/autodba/solid
USER root
COPY --chown=autodba:autodba solid/package.json solid/package-lock.json ./
RUN npm install
COPY --chown=autodba:autodba solid ./
RUN npm run build


FROM base as go_builder

USER root

RUN apt-get install -y --no-install-recommends \
    git             \
    make

# Install golang
ENV GOLANG_VERSION 1.22.1
RUN wget -O go.tgz "https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz" \
    && tar -C /usr/lib -xzf go.tgz \
    && rm go.tgz

# Set golang env vars
ENV PATH="/usr/lib/go/bin:${PATH}" \
    GOROOT="/usr/lib/go"

FROM go_builder as rdsexporter_builder
RUN mkdir -p /usr/lib/prometheus_rds_exporter && \
    git clone https://github.com/crystaldb/prometheus-rds-exporter.git /usr/lib/prometheus_rds_exporter && \
    cd /usr/lib/prometheus_rds_exporter && \
    make build

FROM go_builder as bff_builder
# Build bff
WORKDIR /home/autodba/bff
COPY bff/go.mod bff/go.sum ./
RUN go mod download
COPY bff/ ./
RUN go build -o main ./cmd/main.go
RUN mkdir -p /usr/local/bin
RUN cp main /usr/local/bin/autodba-bff

FROM base AS builder
COPY --from=solid_builder /home/autodba/solid/dist /usr/local/share/autodba/webapp
COPY --from=rdsexporter_builder /usr/lib/prometheus_rds_exporter /usr/local/share/prometheus_exporters/rds_exporter
COPY --from=bff_builder /usr/local/bin/autodba-bff /usr/local/bin/autodba-bff
COPY --from=bff_builder /home/autodba/bff/config.json /etc/autodba/config.json
COPY entrypoint.sh /usr/local/bin/autodba-entrypoint.sh

FROM bff_builder as lint
WORKDIR /home/autodba/bff
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2
# RUN golangci-lint run -v

FROM bff_builder AS test
WORKDIR /home/autodba/bff
# RUN go test ./pkg/server/server_test.go -v
RUN go test ./pkg/server/promql_codegen_test.go -v
# RUN go test ./pkg/metrics/metrics_test.go -v
# RUN go test ./pkg/prometheus/prometheus_test.go -v

FROM base AS autodba
USER root

# Install Prometheus
RUN apt-get install -y --no-install-recommends \
    apt-transport-https \
    software-properties-common \
    sqlite3

RUN wget -qO- https://github.com/prometheus/prometheus/releases/download/v2.42.0/prometheus-2.42.0.linux-amd64.tar.gz | tar -xzf - -C /tmp/
RUN cp /tmp/prometheus-2.42.0.linux-amd64/prometheus /usr/local/bin/
RUN cp /tmp/prometheus-2.42.0.linux-amd64/promtool /usr/local/bin/
RUN mkdir -p /etc/prometheus /var/lib/prometheus
RUN cp -r /tmp/prometheus-2.42.0.linux-amd64/consoles /etc/prometheus/
RUN cp -r /tmp/prometheus-2.42.0.linux-amd64/console_libraries /etc/prometheus/

# Prometheus port
EXPOSE 9090

# Bff port
EXPOSE 4000

# Install Prometheus Exporters

# SQL Exporter
RUN mkdir -p /usr/local/share/prometheus_exporters/sql_exporter && \
    wget -qO- https://github.com/burningalchemist/sql_exporter/releases/download/0.14.3/sql_exporter-0.14.3.linux-amd64.tar.gz | tar -xzf - -C /usr/local/share/prometheus_exporters/sql_exporter --strip-components=1
RUN rm /usr/local/share/prometheus_exporters/sql_exporter/mssql_standard.collector.yml

# Postgres Exporter
RUN mkdir -p /usr/local/share/prometheus_exporters/postgres_exporter && \
    wget -qO- https://github.com/prometheus-community/postgres_exporter/releases/download/v0.15.0/postgres_exporter-0.15.0.linux-amd64.tar.gz | tar -xzf - -C /usr/local/share/prometheus_exporters/postgres_exporter --strip-components=1

COPY --from=builder /usr/local/bin /usr/local/bin
COPY --from=builder /usr/local/share/autodba/webapp /usr/local/share/autodba/webapp
COPY --from=builder /usr/local/share/prometheus_exporters /usr/local/share/prometheus_exporters
COPY --from=builder /etc/autodba/config.json /etc/autodba/config.json

COPY monitor/prometheus/sql_exporter/ /usr/local/share/prometheus_exporters/sql_exporter
COPY monitor/prometheus/rds_exporter/ /usr/local/share/prometheus_exporters/rds_exporter
COPY monitor/prometheus/prometheus.yml /etc/prometheus/prometheus.yml

# Add backup script
COPY scripts/agent/backup.sh /home/autodba/backup.sh
RUN chmod +x /home/autodba/backup.sh
# Ensure backup directory exists
RUN mkdir -p /home/autodba/backups

WORKDIR /etc/autodba

# run autodba-entrypoint.sh
CMD ["/usr/local/bin/autodba-entrypoint.sh"]
