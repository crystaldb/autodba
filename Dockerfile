# SPDX-License-Identifier: Apache-2.0

FROM postgres:16.3 AS base

RUN addgroup --system autodba && adduser --system --group autodba --home /home/autodba --shell /bin/bash

# set environment variables
ENV PYTHONDONTWRITEBYTECODE=1
#ENV PYTHONUNBUFFERED 1  # this works around setups where line buffering is disabled; it should not be needed here

RUN apt-get update
RUN apt-get install -y --no-install-recommends \
    curl            \
    jq              \
    nodejs          \
    npm             \
    procps          \
    python3         \
    python3-venv    \
    python3-pip     \
    wget


USER autodba
WORKDIR /home/autodba
#RUN mkdir src elm

# Create a Python virtual environment
RUN python3 -m venv /home/autodba/venv

# Activate virtual environment
ENV PATH="/home/autodba/venv/bin:$PATH"
ENV PGSSLCERT=/tmp/postgresql.crt

# install + cache python dependencies
WORKDIR /home/autodba/src
COPY src/requirements.txt .
RUN pip install --no-cache-dir -r requirements.txt

FROM base AS builder

# install + cache npm dependencies
#WORKDIR /home/autodba/elm
#COPY --chown=autodba:autodba elm/package.json .
#RUN npm install

# install + cache python dependencies
#WORKDIR /home/autodba/elm
#COPY --chown=autodba:autodba elm .
#RUN npm run build  # creates /home/autodba/elm/dist_prod

WORKDIR /home/autodba/src
COPY --chown=autodba:autodba src .
# If we decide to ship compiled python, the command that does it belongs here.

FROM base AS solid_builder
# Build the solid project
WORKDIR /home/autodba/solid
USER root
COPY --chown=autodba:autodba solid/package.json solid/package-lock.json ./
RUN npm install
COPY --chown=autodba:autodba solid ./
RUN npm run build


FROM base as rdsexporter_builder

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

RUN mkdir -p /usr/lib/prometheus_rds_exporter && \
    git clone https://github.com/crystalcld/prometheus-rds-exporter.git /usr/lib/prometheus_rds_exporter && \
    cd /usr/lib/prometheus_rds_exporter && \
    make build

# Build bff
WORKDIR /home/autodba/bff
COPY bff/go.mod bff/go.sum ./
RUN go mod download
COPY bff/ ./
RUN go build -o main ./cmd/main.go
RUN mkdir -p /usr/lib/bff


FROM builder as lint
WORKDIR /home/autodba/src
RUN flake8 --ignore=E501,F401,E302,E305 .

#WORKDIR /home/autodba/elm
#RUN npm run check-format

FROM builder AS test

#WORKDIR /home/autodba/elm
#RUN npm run format # todo...

# TODO: pytest generates a covearge report that we lose.  Write documentation
#       (or a script) that bind mounts the source checkout on top of . and runs
#       `python -m pytest` in the builder container.
WORKDIR /home/autodba/src
RUN POSTGRES_DB=phony_db AUTODBA_TARGET_DB=postgresql://phony_db_user:phony_db_pass@localhost:5432/phony_db python -m pytest

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


COPY --from=builder /home/autodba/src /home/autodba/src
#COPY --from=builder /home/autodba/elm/dist_prod /home/autodba/src/api/static
COPY --from=solid_builder /home/autodba/solid/dist /home/autodba/src/webapp
COPY --from=rdsexporter_builder /usr/lib/prometheus_rds_exporter /usr/lib/prometheus_rds_exporter
COPY --from=rdsexporter_builder /home/autodba/bff/main /usr/lib/bff/main
COPY --from=rdsexporter_builder /home/autodba/bff/config.json /home/autodba/src/config.json

WORKDIR /home/autodba/src

# TODO: Later, when we support external agent DBs (maybe never?) set these using docker args.
ENV FLASK_APP=api/endpoints.py
# may as well hardcode this on until we have a runtime configuration that we're comfortable with.
# Flask makes it clear that our current setup shouldn't be used in production.
ENV FLASK_DEBUG=True
ENV POSTGRES_DB=autodba_db
ENV POSTGRES_USER=autodba_db_user
ENV POSTGRES_PASSWORD=autodba_db_pass
ENV POSTGRES_HOST=localhost
ENV POSTGRES_PORT=5432

RUN mkdir -p /usr/share/grafana/data
COPY monitor/grafana/grafana.ini /etc/grafana/grafana.ini
COPY monitor/grafana/grafana.db.sql /tmp/grafana.db.sql
RUN sqlite3 /usr/share/grafana/data/grafana.db < /tmp/grafana.db.sql && rm /tmp/grafana.db.sql

COPY monitor/grafana/provisioning/dashboards/* /usr/share/grafana/conf/provisioning/dashboards/
# Add dashboards, and, because we rewrite the dashboards in entrypoint.sh, keep "unmodified" versions that we can copy.
COPY monitor/grafana/dashboards/* /var/lib/grafana/dashboards.unmodified/
RUN mkdir /var/lib/grafana/dashboards

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
