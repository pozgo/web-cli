#!/bin/bash
set -euo pipefail

# healthcheck.sh - Docker/container health check script for web-cli
#
# Determines the correct scheme (http/https) based on TLS configuration
# and checks the /api/health endpoint. The health endpoint does not
# require authentication.
#
# Environment variables used:
#   WEBCLI_PORT           - Port the server listens on (default: 7777)
#   WEBCLI_TLS_CERT_PATH  - TLS certificate path (if set, uses https)
#   WEBCLI_TLS_KEY_PATH   - TLS key path (if set, uses https)
#   TLS_CERT_PATH         - Alternative TLS cert path env var
#   TLS_KEY_PATH          - Alternative TLS key path env var

readonly PORT="${WEBCLI_PORT:-7777}"

# Determine scheme based on TLS configuration
# Check both WEBCLI_TLS_* and TLS_* env vars (server accepts both)
CERT_PATH="${WEBCLI_TLS_CERT_PATH:-${TLS_CERT_PATH:-}}"
KEY_PATH="${WEBCLI_TLS_KEY_PATH:-${TLS_KEY_PATH:-}}"

if [[ -n "${CERT_PATH}" && -n "${KEY_PATH}" ]]; then
    SCHEME="https"
else
    SCHEME="http"
fi

readonly URL="${SCHEME}://localhost:${PORT}/api/health"

# Use -k to accept self-signed certificates in TLS mode
if [[ "${SCHEME}" == "https" ]]; then
    exec curl -sfk "${URL}"
else
    exec curl -sf "${URL}"
fi
