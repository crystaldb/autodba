FROM ubuntu:20.04 AS base

RUN apt-get update && apt-get install -y --no-install-recommends \
    curl \
    wget \
    software-properties-common \
    && rm -rf /var/lib/apt/lists/*

RUN useradd --system --user-group --home-dir /home/autodba --shell /bin/bash autodba

# Install nvm and Node.js
ENV NVM_DIR /usr/local/nvm
ENV NODE_VERSION 16.17.0
RUN mkdir -p $NVM_DIR \
    && curl -o- https://raw.githubusercontent.com/nvm-sh/nvm/v0.39.4/install.sh | bash \
    && . $NVM_DIR/nvm.sh \
    && nvm install $NODE_VERSION \
    && nvm alias default $NODE_VERSION \
    && nvm use default

ENV PATH $NVM_DIR/versions/node/v$NODE_VERSION/bin:$PATH

FROM base AS solid_builder
# Build web-app
WORKDIR /home/autodba/solid
COPY solid/package.json solid/package-lock.json ./
RUN npm install
COPY solid ./
RUN npm run build

FROM base AS go_builder
# Install golang
ENV GOLANG_VERSION="1.22.1"
RUN wget -O go.tgz "https://golang.org/dl/go${GOLANG_VERSION}.linux-amd64.tar.gz" \
    && tar -C /usr/lib -xzf go.tgz \
    && rm go.tgz
ENV PATH="/usr/lib/go/bin:${PATH}" \
    GOROOT="/usr/lib/go"

FROM go_builder as bff_builder
# Build BFF
WORKDIR /home/autodba/bff
RUN mkdir -p /usr/local/autodba/bin
COPY ./go.mod ./go.sum ./
RUN go mod download
COPY ./ ./
RUN go build -o main ./cmd/main.go
RUN cp main /usr/local/autodba/bin/autodba-bff

FROM base as bff_webapp
WORKDIR /usr/local/autodba

COPY bff-entrypoint.sh ./bin/bff-entrypoint.sh
RUN chmod +x ./bin/bff-entrypoint.sh

# Copy built files from previous stages
COPY --from=solid_builder /home/autodba/solid/dist ./share/webapp
COPY --from=bff_builder /home/autodba/bff/config.json ./config/autodba/config.json
COPY --from=bff_builder /usr/local/autodba/bin ./bin

WORKDIR /usr/local/autodba/config/autodba
CMD ["/usr/local/autodba/bin/bff-entrypoint.sh"]
