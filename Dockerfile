# SPDX-License-Identifier: Apache-2.0

FROM ubuntu:20.04 AS base

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    wget \
    software-properties-common \
    && rm -rf /var/lib/apt/lists/*

RUN useradd --system --user-group --home-dir /home/autodba --shell /bin/bash autodba

# Install nvm
ENV NVM_DIR /usr/local/nvm
ENV NODE_VERSION 16.17.0
RUN mkdir -p $NVM_DIR \
    && curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.4/install.sh | bash \
    && . $NVM_DIR/nvm.sh \
    && nvm install $NODE_VERSION \
    && nvm alias default $NODE_VERSION \
    && nvm use default

# Add nvm and node to PATH
ENV PATH $NVM_DIR/versions/node/v$NODE_VERSION/bin:$PATH

USER root
RUN mkdir -p /usr/local/autodba/config/autodba /usr/local/autodba/bin
WORKDIR /usr/local/autodba

FROM base AS solid_builder
# Build the solid project
WORKDIR /home/autodba/solid
USER root
COPY --chown=autodba:autodba solid/package.json solid/package-lock.json ./
RUN npm install
COPY --chown=autodba:autodba solid ./
RUN npm run build

FROM base AS go_builder

RUN apt-get update && apt-get install -y --no-install-recommends \
    git \
    unzip \
    make \
    libc-dev \
    gcc \
    tar \
    && rm -rf /var/lib/apt/lists/*

# Install golang
ENV GOLANG_VERSION="1.22.1"
RUN wget -O go.tgz "https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz" \
    && tar -C /usr/lib -xzf go.tgz \
    && rm go.tgz

ENV PATH="/usr/lib/go/bin:${PATH}" \
    GOROOT="/usr/lib/go"

FROM go_builder AS go_builder_with_deps
WORKDIR /app
COPY bff/go.mod bff/go.sum ./bff/
COPY collector-api/go.mod collector-api/go.sum ./collector-api/
RUN cd bff && go mod download
RUN cd collector-api && go mod download

FROM go_builder_with_deps as collector_builder
RUN mkdir -p /usr/local/autodba/share/collector && \
    git clone --recurse-submodules https://github.com/crystaldb/collector.git /usr/local/autodba/share/collector && \
    cd /usr/local/autodba/share/collector && \
    wget https://github.com/protocolbuffers/protobuf/releases/download/v3.14.0/protoc-3.14.0-linux-x86_64.zip && unzip protoc-3.14.0-linux-x86_64.zip -d protoc && \
    make build && \
    mv pganalyze-collector collector && \
    mv pganalyze-collector-helper collector-helper && \
    mv pganalyze-collector-setup collector-setup


FROM go_builder_with_deps as collector_api_server_builder
WORKDIR /usr/local/autodba/share/collector_api_server
COPY collector-api/ ./
RUN go build -o collector-api-server ./cmd/server/main.go

FROM go_builder_with_deps as bff_builder
# Build bff
WORKDIR /home/autodba/bff
COPY bff/ ./
RUN go build -o main ./cmd/main.go
RUN mkdir -p /usr/local/autodba/bin
RUN cp main /usr/local/autodba/bin/autodba-bff

FROM base AS builder
WORKDIR /usr/local/autodba
COPY --from=solid_builder /home/autodba/solid/dist ./share/webapp
COPY --from=collector_builder /usr/local/autodba/share/collector ./share/collector
COPY --from=collector_api_server_builder  /usr/local/autodba/share/collector_api_server ./share/collector_api_server
COPY --from=bff_builder /usr/local/autodba/bin/autodba-bff ./bin/autodba-bff
COPY --from=bff_builder /home/autodba/bff/config.json ./config/autodba/config.json
COPY entrypoint.sh ./bin/autodba-entrypoint.sh

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

# Prometheus port
EXPOSE 9090

# BFF port
EXPOSE 4000

# Install Prometheus and SQLite
RUN apt-get update && apt-get install -y --no-install-recommends \
    apt-transport-https \
    sqlite3 \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /usr/local/autodba

# Install Prometheus
RUN wget -qO- https://github.com/prometheus/prometheus/releases/download/v2.42.0/prometheus-2.42.0.linux-amd64.tar.gz | tar -xzf - -C /tmp/ \
    && mkdir -p ./prometheus ./config/prometheus \
    && cp /tmp/prometheus-2.42.0.linux-amd64/prometheus ./prometheus/ \
    && cp /tmp/prometheus-2.42.0.linux-amd64/promtool ./prometheus/ \
    && cp -r /tmp/prometheus-2.42.0.linux-amd64/consoles ./config/prometheus/ \
    && cp -r /tmp/prometheus-2.42.0.linux-amd64/console_libraries ./config/prometheus/ \
    && rm -rf /tmp/prometheus-2.42.0.linux-amd64

# Prometheus config
COPY prometheus/prometheus.yml ./config/prometheus/prometheus.yml

# Add backup script
COPY scripts/agent/backup.sh /home/autodba/backup.sh
RUN chmod +x /home/autodba/backup.sh
RUN mkdir -p /home/autodba/backups

# Copy built files from previous stages
COPY --from=builder /usr/local/autodba/share/webapp ./share/webapp
COPY --from=builder /usr/local/autodba/config/autodba/config.json ./config/autodba/config.json
COPY --from=builder /usr/local/autodba/share/collector ./share/collector
COPY --from=builder /usr/local/autodba/share/collector_api_server ./share/collector_api_server
COPY --from=builder /usr/local/autodba/bin ./bin

WORKDIR /usr/local/autodba/config/autodba

# Run the entrypoint script
CMD ["/usr/local/autodba/bin/autodba-entrypoint.sh"]
