# Build stage 1: Frontend
FROM node:20-bookworm-slim AS frontend-builder

WORKDIR /app/frontend

# Copy frontend package files
COPY frontend/package*.json ./

# Install dependencies
RUN npm ci --legacy-peer-deps

# Copy frontend source
COPY frontend/ ./

# Build frontend
RUN npm run build

# Build stage 2: Go binary
FROM golang:1.24-bookworm AS go-builder

# Install build dependencies
RUN apt-get update && apt-get install -y --no-install-recommends \
    git ca-certificates tzdata \
    && rm -rf /var/lib/apt/lists/*

WORKDIR /app

# Copy go mod files first for better caching
COPY go.mod go.sum ./
RUN go mod download

# Copy source code
COPY cmd/ ./cmd/
COPY internal/ ./internal/
COPY assets/ ./assets/

# Copy built frontend from previous stage
COPY --from=frontend-builder /app/frontend/dist ./assets/frontend/

# Build arguments for multi-platform support
ARG TARGETOS=linux
ARG TARGETARCH

# Build the binary for target platform
RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH} go build \
    -ldflags="-s -w -X main.Version=docker" \
    -o /app/web-cli \
    ./cmd/web-cli

# Final stage: Debian-based runtime for proper bash support
FROM debian:bookworm-slim

# Install required packages:
# - ca-certificates: for HTTPS connections
# - tzdata: for timezone support
# - bash: full bash shell for script execution
# - coreutils: standard Unix utilities
# - curl: for health checks and HTTP operations
# - openssh-client: for SSH connections to remote servers
RUN apt-get update && apt-get install -y --no-install-recommends \
    ca-certificates \
    tzdata \
    bash \
    coreutils \
    curl \
    openssh-client \
    && rm -rf /var/lib/apt/lists/* \
    && apt-get clean

# Create non-root user for security
RUN groupadd -g 1000 webcli && \
    useradd -u 1000 -g webcli -s /bin/bash -m webcli

# Create data and config directories
RUN mkdir -p /data /config && \
    chown -R webcli:webcli /data /config

# Create .ssh directory for SSH key operations
RUN mkdir -p /home/webcli/.ssh && \
    chown webcli:webcli /home/webcli/.ssh && \
    chmod 700 /home/webcli/.ssh

WORKDIR /app

# Copy binary from builder
COPY --from=go-builder /app/web-cli /app/web-cli

# Set ownership
RUN chown webcli:webcli /app/web-cli

# Switch to non-root user
USER webcli

# Expose default port
EXPOSE 7777

# Health check using curl
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD curl -sf http://localhost:7777/api/health || exit 1

# Environment variables with defaults
ENV WEBCLI_PORT=7777 \
    WEBCLI_HOST=0.0.0.0 \
    WEBCLI_DATABASE_PATH=/data/web-cli.db \
    WEBCLI_ENCRYPTION_KEY_PATH=/data/.encryption_key \
    SHELL=/bin/bash

# Volume for persistent data
VOLUME ["/data", "/config"]

# Run the application
ENTRYPOINT ["/app/web-cli"]
