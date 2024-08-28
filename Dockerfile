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
RUN mkdir -p /usr/lib/bff

FROM base AS autodba

USER root

# Install Prometheus + Grafana
RUN apt-get install -y --no-install-recommends \
    apt-transport-https \
    prometheus \
    software-properties-common \
    sqlite3


RUN npm install -g serve

# Expose port 8080 for HTTP traffic
EXPOSE 8080

# Grafana port
EXPOSE 3000

# Prometheus port
EXPOSE 9090

# Bff port
EXPOSE 4000

# Webapp port
EXPOSE 5000

RUN mkdir -p /etc/apt/keyrings/
RUN wget -q -O - https://apt.grafana.com/gpg.key | gpg --dearmor > /etc/apt/keyrings/grafana.gpg
RUN echo "deb [signed-by=/etc/apt/keyrings/grafana.gpg] https://apt.grafana.com stable main" > /etc/apt/sources.list.d/grafana.list
RUN apt-get update
RUN apt-get install -y grafana

# Install Prometheus exporters
RUN mkdir -p /usr/lib/prometheus_sql_exporter && \
    wget -qO- https://github.com/burningalchemist/sql_exporter/releases/download/0.14.3/sql_exporter-0.14.3.linux-amd64.tar.gz | tar -xzf - -C /usr/lib/prometheus_sql_exporter --strip-components=1
RUN rm /usr/lib/prometheus_sql_exporter/mssql_standard.collector.yml

RUN mkdir -p /usr/lib/prometheus_postgres_exporter && \
    wget -qO- https://github.com/prometheus-community/postgres_exporter/releases/download/v0.15.0/postgres_exporter-0.15.0.linux-amd64.tar.gz | tar -xzf - -C /usr/lib/prometheus_postgres_exporter --strip-components=1


COPY --from=solid_builder /home/autodba/solid/dist /home/autodba/src/webapp
COPY --from=rdsexporter_builder /usr/lib/prometheus_rds_exporter /usr/lib/prometheus_rds_exporter
COPY --from=bff_builder /home/autodba/bff/main /usr/lib/bff/main
COPY --from=bff_builder /home/autodba/bff/config.json /home/autodba/src/config.json
COPY entrypoint.sh /home/autodba/src/entrypoint.sh

WORKDIR /home/autodba/src

COPY monitor/prometheus/sql_exporter/ /usr/lib/prometheus_sql_exporter
COPY monitor/prometheus/rds_exporter/ /usr/lib/prometheus_rds_exporter
COPY monitor/prometheus/prometheus.yml /etc/prometheus/prometheus.yml

# Add backup script
COPY scripts/agent/backup.sh /home/autodba/backup.sh
RUN chmod +x /home/autodba/backup.sh
# Ensure backup directory exists
RUN mkdir -p /home/autodba/backups

# run entrypoint.sh
CMD ["./entrypoint.sh"]
