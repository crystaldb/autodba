# SPDX-License-Identifier: Apache-2.0

FROM ubuntu:20.04 AS builder

RUN apt-get update && DEBIAN_FRONTEND=noninteractive TZ=Etc/UTC apt-get install -y --no-install-recommends \
    curl \
    wget \
    software-properties-common \
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

FROM builder as release
WORKDIR /home/autodba
COPY ./ ./
RUN ./scripts/build.sh && \
    mkdir -p release_output && \
    mv build_output/tar.gz/autodba-*.tar.gz release_output/  && \
    mv build_output/tar.gz/collector-*.tar.gz release_output/  && \
    rm -rf build_output

FROM builder as lint
COPY bff/ ./
RUN go mod download
RUN curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $(go env GOPATH)/bin v1.55.2

FROM builder AS test
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
