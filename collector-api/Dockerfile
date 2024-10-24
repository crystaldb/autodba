FROM ubuntu:20.04

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    wget \
    software-properties-common \
    && rm -rf /var/lib/apt/lists/*

# Install golang
ENV GOLANG_VERSION="1.22.1"
RUN wget -O go.tgz "https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz" \
    && tar -C /usr/lib -xzf go.tgz \
    && rm go.tgz

ENV PATH="/usr/lib/go/bin:${PATH}" \
    GOROOT="/usr/lib/go"

WORKDIR /usr/local/autodba/share/collector_api_server
COPY ./ ./
RUN go mod download
RUN go build -o collector-api-server ./cmd/server/main.go

# Create necessary directories
RUN mkdir -p storage

COPY collector-api-entrypoint.sh /usr/local/autodba/bin/collector-api-entrypoint.sh
RUN chmod +x /usr/local/autodba/bin/collector-api-entrypoint.sh

CMD ["/usr/local/autodba/bin/collector-api-entrypoint.sh"]
