# Multi-stage Dockerfile for Portfolio Backend
# Uses Oracle Linux 9 slim with native Oracle Instant Client packages for CGO / godror support.

# --- Build Stage ---
FROM oraclelinux:9-slim AS builder

# Prepare yum vars and install Oracle Instant Client 19c SDK and GCC.
# Go is installed from the official tarball (checksum-pinned) because the
# distro `golang` package lags behind the go.mod required toolchain.
RUN mkdir -p /etc/yum/vars && \
    microdnf install -y oracle-instantclient-release-el9 && \
    microdnf install -y \
        oracle-instantclient19.19-basic \
        oracle-instantclient19.19-devel \
        curl \
        gcc \
        git && \
    microdnf clean all && \
    curl -fsSLo /tmp/go.tgz https://go.dev/dl/go1.26.6.linux-amd64.tar.gz && \
    echo "708effb774be8237570d0add163225abbdfaf4fca28b2611df167beba4feef89  /tmp/go.tgz" | sha256sum -c - && \
    rm -rf /usr/local/go && \
    tar -C /usr/local -xzf /tmp/go.tgz && \
    rm /tmp/go.tgz

ENV PATH="/usr/local/go/bin:${PATH}"

WORKDIR /build

# Cache Go modules layer
COPY go.mod go.sum ./
RUN go mod download

# Copy source code and build production binary
COPY . .
RUN CGO_ENABLED=1 GOOS=linux go build \
    -ldflags="-s -w -X main.version=1.0.0" \
    -o /build/bin/api \
    ./cmd/api

# --- Runtime Stage ---
FROM oraclelinux:9-slim AS runner

# Install Oracle Instant Client 19c runtime, CA certs, timezone data, and curl for healthcheck
RUN mkdir -p /etc/yum/vars && \
    microdnf install -y \
        oracle-instantclient-release-el9 \
        shadow-utils \
        tzdata \
        ca-certificates \
        curl && \
    microdnf install -y \
        oracle-instantclient19.19-basic && \
    microdnf clean all && \
    useradd -u 10001 -m -s /bin/sh appuser

WORKDIR /app
RUN mkdir -p /app/logs /app/uploads && chown -R appuser:appuser /app

# Copy binary and required static docs
COPY --from=builder --chown=appuser:appuser /build/bin/api /app/api
COPY --chown=appuser:appuser docs/openapi.yaml /app/docs/openapi.yaml

# Run as non-privileged user
USER appuser

# Expose HTTP port
EXPOSE 8080

# Configure container healthcheck
HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD curl -f http://localhost:8080/api/v1/health || exit 1

ENTRYPOINT ["/app/api"]
