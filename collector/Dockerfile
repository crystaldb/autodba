FROM ubuntu:20.04

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

RUN mkdir -p /usr/local/autodba/share/collector && \
    git clone --recurse-submodules https://github.com/crystaldb/collector.git /usr/local/autodba/share/collector && \
    cd /usr/local/autodba/share/collector && \
    git checkout 636bc0996ecc881acc0942859f8a28684dadfc7e && \
    wget https://github.com/protocolbuffers/protobuf/releases/download/v3.14.0/protoc-3.14.0-linux-x86_64.zip && unzip protoc-3.14.0-linux-x86_64.zip -d protoc && \
    make build && \
    mv pganalyze-collector collector && \
    mv pganalyze-collector-helper collector-helper && \
    mv pganalyze-collector-setup collector-setup

COPY collector-entrypoint.sh /usr/local/autodba/bin/collector-entrypoint.sh
RUN chmod +x /usr/local/autodba/bin/collector-entrypoint.sh

WORKDIR /usr/local/autodba/share/collector

CMD ["/usr/local/autodba/bin/collector-entrypoint.sh"]
