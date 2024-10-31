FROM ubuntu:20.04

RUN apt-get update && DEBIAN_FRONTEND=noninteractive TZ=Etc/UTC apt-get install -y --no-install-recommends \
    curl \
    wget \
    software-properties-common \
    libc-dev \
    gcc \
    && rm -rf /var/lib/apt/lists/*

# Install golang
ENV GOLANG_VERSION="1.22.1"
RUN wget -O go.tgz "https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz" \
    && tar -C /usr/lib -xzf go.tgz \
    && rm go.tgz

ENV PATH="/usr/lib/go/bin:${PATH}" \
    GOROOT="/usr/lib/go"

WORKDIR /usr/local/autodba/share/collector_api_server

# Create necessary directories
RUN mkdir -p storage

# Ensure CGO is enabled
ENV CGO_ENABLED=1
COPY ./go.mod ./go.sum ./
RUN go mod download

COPY collector-api-entrypoint.sh /usr/local/autodba/bin/collector-api-entrypoint.sh
RUN chmod +x /usr/local/autodba/bin/collector-api-entrypoint.sh

COPY ./ ./
RUN go build -o collector-api-server ./cmd/server/main.go

CMD ["/usr/local/autodba/bin/collector-api-entrypoint.sh"]
