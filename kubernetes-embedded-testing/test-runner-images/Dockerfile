# Build base image with common dependencies
FROM ubuntu:24.04 AS base
ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update && apt-get install -y \
    curl \
    wget \
    gnupg \
    ca-certificates \
    lsb-release \
    bash \
    && rm -rf /var/lib/apt/lists/*

COPY --from=bitnami/kubectl:latest /opt/bitnami/kubectl/bin/kubectl /usr/local/bin/kubectl

RUN curl -fsSL https://raw.githubusercontent.com/metalbear-co/mirrord/main/scripts/install.sh | bash

RUN mkdir -p /workspace /reports
WORKDIR /workspace

RUN kubectl version --client \
    && mirrord --version

# Build Node.js test runner
FROM base AS node-runner
ENV NODE_VERSION=22

RUN curl -fsSL https://deb.nodesource.com/setup_${NODE_VERSION}.x | bash - \
    && apt-get install -y nodejs \
    && rm -rf /var/lib/apt/lists/*

RUN node --version \
    && npm --version

# Build Go test runner
FROM base AS golang-runner
ENV GO_VERSION=1.25.0

RUN wget https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz \
    && tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz \
    && rm go${GO_VERSION}.linux-amd64.tar.gz

ENV PATH=$PATH:/usr/local/go/bin
ENV GOPATH=/go
ENV GOCACHE=/go/cache

RUN mkdir -p /go/src /go/bin /go/cache

RUN go version

# Default to Node.js runner for backward compatibility
FROM node-runner
CMD ["/bin/bash"]
