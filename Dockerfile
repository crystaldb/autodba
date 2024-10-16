# SPDX-License-Identifier: Apache-2.0

FROM ubuntu:20.04 AS base

RUN useradd --system --user-group --home-dir /home/autodba --shell /bin/bash autodba

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl            \
    jq              \
    procps          \
    wget           \
    software-properties-common

# Install nvm
ENV NVM_DIR /usr/local/nvm
RUN mkdir -p $NVM_DIR \
    && curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.4/install.sh | bash \
    && bash -c "source $NVM_DIR/nvm.sh && nvm install 16.17.0 && nvm use 16.17.0 && nvm alias default 16.17.0"

# Add nvm and node to PATH
ENV PATH $NVM_DIR/versions/node/v16.17.0/bin:$PATH

USER root
RUN mkdir -p /usr/local/autodba/config/autodba
RUN mkdir -p /usr/local/autodba/bin
WORKDIR /usr/local/autodba

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
RUN apt-get update && apt-get install -y --no-install-recommends \
    git             \
    unzip             \
    make \
    libc-dev \
    gcc \
    tar

# Install golang
ENV GOLANG_VERSION="1.22.1"
RUN wget -O go.tgz "https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz" \
    && tar -C /usr/lib -xzf go.tgz \
    && rm go.tgz

# Set golang env vars
ENV PATH="/usr/lib/go/bin:${PATH}" \
    GOROOT="/usr/lib/go"

FROM go_builder as collector_builder
RUN mkdir -p /usr/local/autodba/share/collector && \
    git clone --recurse-submodules https://github.com/crystaldb/collector.git /usr/local/autodba/share/collector && \
    cd /usr/local/autodba/share/collector && \
    wget https://github.com/protocolbuffers/protobuf/releases/download/v3.14.0/protoc-3.14.0-linux-x86_64.zip && unzip protoc-3.14.0-linux-x86_64.zip -d protoc && \
    make build && \
    mv pganalyze-collector collector && \
    mv pganalyze-collector-helper collector-helper && \
    mv pganalyze-collector-setup collector-setup


FROM go_builder as collector_api_server_builder
WORKDIR /usr/local/autodba/share/collector_api_server
COPY collector-api/ /usr/local/autodba/share/collector_api_server/
RUN go build -o collector-api-server ./cmd/server/main.go

FROM go_builder as bff_builder
# Build bff
WORKDIR /home/autodba/bff
COPY bff/go.mod bff/go.sum ./
RUN go mod download
COPY bff/ ./
RUN go build -o main ./cmd/main.go
RUN mkdir -p /usr/local/autodba/bin
RUN cp main /usr/local/autodba/bin/autodba-bff

FROM base AS builder
COPY --from=solid_builder /home/autodba/solid/dist /usr/local/autodba/share/webapp
COPY --from=collector_builder /usr/local/autodba/share/collector /usr/local/autodba/share/collector
COPY --from=collector_api_server_builder  /usr/local/autodba/share/collector_api_server /usr/local/autodba/share/collector_api_server
COPY --from=bff_builder /usr/local/autodba/bin/autodba-bff /usr/local/autodba/bin/autodba-bff
COPY --from=bff_builder /home/autodba/bff/config.json /usr/local/autodba/config/autodba/config.json
COPY entrypoint.sh /usr/local/autodba/bin/autodba-entrypoint.sh

FROM bff_builder as lint
WORKDIR /home/autodba/bff
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2

FROM go_builder as release
WORKDIR /home/autodba
COPY ./ ./
RUN ./scripts/build.sh && \
    mkdir -p release_output && \
    mv build_output/tar.gz/autodba-*.tar.gz release_output/  && \
    rm -rf build_output

FROM go_builder AS test
WORKDIR /home/autodba/bff
COPY bff/go.mod bff/go.sum ./
RUN go mod download
COPY bff/ ./
RUN go test ./pkg/server -v
RUN go test ./pkg/metrics -v
RUN go test ./pkg/prometheus -v

WORKDIR /home/autodba/collector-api
COPY collector-api/go.mod collector-api/go.sum ./
RUN go mod download
COPY collector-api/ ./
RUN go test -v ./...

FROM base AS autodba
USER root

# Install Prometheus
RUN apt-get install -y --no-install-recommends \
    apt-transport-https \
    sqlite3

RUN wget -qO- https://github.com/prometheus/prometheus/releases/download/v2.42.0/prometheus-2.42.0.linux-amd64.tar.gz | tar -xzf - -C /tmp/
RUN mkdir -p /usr/local/autodba/prometheus
RUN cp /tmp/prometheus-2.42.0.linux-amd64/prometheus /usr/local/autodba/prometheus/
RUN cp /tmp/prometheus-2.42.0.linux-amd64/promtool /usr/local/autodba/prometheus/
RUN mkdir -p /usr/local/autodba/config/prometheus
RUN cp -r /tmp/prometheus-2.42.0.linux-amd64/consoles /usr/local/autodba/config/prometheus/
RUN cp -r /tmp/prometheus-2.42.0.linux-amd64/console_libraries /usr/local/autodba/config/prometheus/

# Prometheus port
EXPOSE 9090

# BFF port
EXPOSE 4000

# Copy built files from previous stages
COPY --from=builder /usr/local/autodba/bin /usr/local/autodba/bin
COPY --from=builder /usr/local/autodba/share/webapp /usr/local/autodba/share/webapp
COPY --from=builder /usr/local/autodba/share/collector /usr/local/autodba/share/collector
COPY --from=builder /usr/local/autodba/share/collector_api_server /usr/local/autodba/share/collector_api_server
COPY --from=builder /usr/local/autodba/config/autodba/config.json /usr/local/autodba/config/autodba/config.json

# Monitor setup
COPY monitor/prometheus/prometheus.yml /usr/local/autodba/config/prometheus/prometheus.yml

# Add backup script
COPY scripts/agent/backup.sh /home/autodba/backup.sh
RUN chmod +x /home/autodba/backup.sh
RUN mkdir -p /home/autodba/backups

WORKDIR /usr/local/autodba/config/autodba

# Run the entrypoint script
CMD ["/usr/local/autodba/bin/autodba-entrypoint.sh"]
