# Build stage 1: Frontend
FROM node:20-alpine AS frontend-builder

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
FROM golang:1.24-alpine AS go-builder

# Install build dependencies
RUN apk add --no-cache git ca-certificates tzdata

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

# Build the binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build \
    -ldflags="-s -w -X main.Version=docker" \
    -o /app/web-cli \
    ./cmd/web-cli

# Final stage: Minimal runtime image
FROM alpine:3.20

# Install ca-certificates for HTTPS and tzdata for timezones
RUN apk add --no-cache ca-certificates tzdata

# Create non-root user for security
RUN addgroup -g 1000 webcli && \
    adduser -u 1000 -G webcli -s /bin/sh -D webcli

# Create data directory
RUN mkdir -p /data /config && \
    chown -R webcli:webcli /data /config

WORKDIR /app

# Copy binary from builder
COPY --from=go-builder /app/web-cli /app/web-cli

# Set ownership
RUN chown webcli:webcli /app/web-cli

# Switch to non-root user
USER webcli

# Expose default port
EXPOSE 7777

# Health check
HEALTHCHECK --interval=30s --timeout=5s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:7777/api/health || exit 1

# Environment variables with defaults
ENV WEBCLI_PORT=7777 \
    WEBCLI_HOST=0.0.0.0 \
    WEBCLI_DATABASE_PATH=/data/web-cli.db \
    WEBCLI_ENCRYPTION_KEY_PATH=/data/.encryption_key

# Volume for persistent data
VOLUME ["/data", "/config"]

# Run the application
ENTRYPOINT ["/app/web-cli"]
