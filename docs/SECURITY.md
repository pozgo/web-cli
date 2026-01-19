# Security Guide

This document describes Web CLI's security features and best practices for secure deployment.

## Table of Contents

- [Authentication](#authentication)
- [TLS/HTTPS](#tlshttps)
- [SSH Host Key Verification](#ssh-host-key-verification)
- [Input Validation](#input-validation)
- [Database Encryption](#database-encryption)
- [Encryption Key Management](#encryption-key-management)
- [Password Security](#password-security)
- [SSRF Protection](#ssrf-protection)
- [Path Traversal Protection](#path-traversal-protection)
- [HTTP Security](#http-security)
- [Audit Logging](#audit-logging)
- [Error Information Leakage Prevention](#error-information-leakage-prevention)
- [Security Headers](#security-headers)
- [Production Security Checklist](#production-security-checklist)

---

## Authentication

**Important: Authentication is disabled by default for development convenience.**

### Enable Authentication

```bash
# Enable authentication
export AUTH_ENABLED=true

# Option 1: HTTP Basic Authentication
export AUTH_USERNAME="admin"
export AUTH_PASSWORD="your-secure-password"

# Option 2: API Token (Bearer)
export AUTH_API_TOKEN="your-api-token-here"
```

### Features

- HTTP Basic Authentication support
- Bearer token (API token) support
- Constant-time credential comparison (prevents timing attacks)
- Supports both methods simultaneously (token takes precedence)
- **Startup validation**: Server fails fast if auth is enabled but credentials are missing

### Usage Examples

```bash
# Basic Auth
curl -u admin:password http://localhost:7777/api/health

# Bearer Token
curl -H "Authorization: Bearer your-token" http://localhost:7777/api/health
```

### Testing Authentication

```bash
# Should fail (401 Unauthorized)
curl http://localhost:7777/api/health

# Should succeed with Basic Auth
curl -u admin:password http://localhost:7777/api/health

# Should succeed with Bearer token
curl -H "Authorization: Bearer your-token" http://localhost:7777/api/health
```

---

## TLS/HTTPS

Native TLS support for encrypted connections without requiring a reverse proxy.

### Enable TLS

```bash
# Using command-line flags
./web-cli -tls-cert /path/to/cert.pem -tls-key /path/to/key.pem

# Using environment variables
WEBCLI_TLS_CERT_PATH=/path/to/cert.pem \
WEBCLI_TLS_KEY_PATH=/path/to/key.pem \
./web-cli

# Enforce HTTPS when authentication is enabled
./web-cli -tls-cert cert.pem -tls-key key.pem -require-https
```

### Generate Self-Signed Certificate (Testing)

```bash
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem \
  -days 365 -nodes -subj "/CN=localhost"
```

### Features

- Native Go TLS implementation
- Automatic HTTPS when certificate and key are provided
- Optional HTTPS enforcement (rejects HTTP requests)
- Works with any TLS certificate (self-signed, Let's Encrypt, etc.)

---

## SSH Host Key Verification

Proper host key verification for secure SSH connections.

### Features

- Verifies SSH host keys against `~/.ssh/known_hosts`
- Supports "trust on first use" (TOFU) mode for development
- Detects man-in-the-middle attacks (host key mismatch)
- Automatically saves new trusted host keys
- Thread-safe implementation

### Configuration

- **Strict Mode** (production): Rejects unknown hosts
- **Trust-on-First-Use** (development): Automatically trusts new hosts

---

## Input Validation

All user inputs are validated before processing to prevent injection attacks.

### Validated Inputs

| Input Type | Validation |
|------------|------------|
| IP addresses | IPv4/IPv6 format validation |
| Hostnames | RFC 1123 compliant |
| Port numbers | 1-65535 range |
| SSH private keys | PEM format validation |
| Unix usernames | Alphanumeric, dash, underscore |
| Command names | No null bytes or newlines |
| Vault addresses | SSRF protection |
| Vault secret paths | Path traversal protection |
| Vault secret names | Alphanumeric, dash, underscore, dot |
| Environment variable names | Unix standards |
| Script content | Size limits, no null bytes |

---

## Database Encryption

All sensitive data is encrypted using **AES-256-GCM** (military-grade encryption).

### Encrypted Data

- SSH private keys
- Command history (commands and output)
- Environment variable values
- Vault tokens

### Key Generation

Encryption key is auto-generated on first run and stored with 600 permissions.

---

## Encryption Key Management

### Generate a New Encryption Key

```bash
# macOS/Linux
openssl rand -base64 32

# Or using dd and base64
dd if=/dev/urandom bs=32 count=1 2>/dev/null | base64

# Output example: 7xK9mP2vQ8nL4wR6tY5uE3sA1zD0cF8bG7hJ9kM6nP4=
```

### Use the Generated Key

```bash
# Option 1: Environment variable (recommended for production)
export ENCRYPTION_KEY="7xK9mP2vQ8nL4wR6tY5uE3sA1zD0cF8bG7hJ9kM6nP4="
./web-cli

# Option 2: Save to file (auto-generated on first run)
echo "7xK9mP2vQ8nL4wR6tY5uE3sA1zD0cF8bG7hJ9kM6nP4=" > .encryption_key
chmod 600 .encryption_key
./web-cli
```

### Important

- **Backup your `.encryption_key` file** - data cannot be recovered without it
- For production, use environment variable instead of file
- System entropy is verified before key generation (Linux)

---

## Password Security

- Sudo passwords only used for command execution
- SSH passwords used for authentication fallback
- **Passwords are never stored** in command history
- Passwords cleared from memory after use
- bcrypt password hashing with cost factor 12

---

## SSRF Protection

When using HashiCorp Vault integration, the application includes SSRF (Server-Side Request Forgery) protection.

### Allowed Addresses

- Public IP addresses
- Private IPs (10.x.x.x, 172.16.x.x, 192.168.x.x) for self-hosted Vault
- Localhost (127.0.0.1) for local development
- Valid hostnames

### Blocked Addresses (Attack Vectors)

- Link-local addresses (169.254.x.x) - used by cloud metadata services
- Unspecified addresses (0.0.0.0)
- Cloud metadata hostnames:
  - `metadata.google.internal`
  - `metadata.goog`
  - `metadata`
  - `instance-data`

This allows you to use self-hosted Vault on private networks while preventing attacks against cloud metadata endpoints.

---

## Path Traversal Protection

All Vault secret paths are validated to prevent path traversal attacks.

### Blocked Patterns

- Path traversal sequences (`..`)
- Absolute paths starting with `/`
- Backslashes (`\`)
- URL-encoded characters (`%`)
- Consecutive slashes (`//`)

### Allowed Characters

Only alphanumeric, dash, underscore, dot, and forward slash are allowed.

---

## HTTP Security

### Configurable Timeouts

Protection against slowloris and DoS attacks with customizable values:

| Timeout | Default | Environment Variable | Description |
|---------|---------|---------------------|-------------|
| Read Timeout | 30s | `READ_TIMEOUT` | Time to read request headers |
| Write Timeout | 600s | `WRITE_TIMEOUT` | Time to write response |
| Idle Timeout | 60s | `IDLE_TIMEOUT` | Keep-alive connection timeout |
| Command Timeout | 300s | `COMMAND_TIMEOUT` | Command execution timeout |
| SSH Connect Timeout | 30s | `SSH_CONNECT_TIMEOUT` | SSH connection establishment |
| Vault Timeout | 30s | `VAULT_TIMEOUT` | HashiCorp Vault operations |

### CORS Policy

- Default: localhost only
- Production: Configure via `CORS_ALLOWED_ORIGINS`

```bash
# Single origin
export CORS_ALLOWED_ORIGINS="https://web-cli.example.com"

# Multiple origins (comma-separated)
export CORS_ALLOWED_ORIGINS="https://web-cli.example.com,https://admin.example.com"
```

---

## Audit Logging

Comprehensive audit logging for security compliance and monitoring.

### Enable Audit Logging

```bash
export AUDIT_LOG_PATH=/var/log/web-cli/audit.log
./web-cli
```

### Logged Events

- Command executions (local and remote)
- Script executions
- Terminal sessions (start/end)
- Authentication attempts

### Log Format

JSON Lines (JSONL) - one JSON object per line:

```json
{
  "timestamp": "2025-01-19T10:30:00Z",
  "event_type": "command_execution",
  "outcome": "success",
  "actor": "admin",
  "source_ip": "192.168.1.100",
  "target": "local",
  "command": "df -h",
  "user": "root",
  "exit_code": 0,
  "duration_ms": 150
}
```

### Log Rotation

```bash
# /etc/logrotate.d/web-cli
/var/log/web-cli/audit.log {
    daily
    rotate 30
    compress
    delaycompress
    missingok
    notifempty
    create 640 www-data www-data
}
```

---

## Error Information Leakage Prevention

API error responses are sanitized to prevent information leakage:

- Internal error details (IP addresses, paths, stack traces) are not exposed
- Validation errors are returned with safe, descriptive messages
- Generic "operation failed" messages for unexpected errors
- Sensitive configuration details are never included in responses

---

## Security Headers

The following security headers are set on all responses:

| Header | Value | Purpose |
|--------|-------|---------|
| `X-Frame-Options` | `DENY` | Prevent clickjacking |
| `X-Content-Type-Options` | `nosniff` | Prevent MIME sniffing |
| `X-XSS-Protection` | `1; mode=block` | Enable XSS filter |
| `Referrer-Policy` | `strict-origin-when-cross-origin` | Control referrer information |

---

## Production Security Checklist

### Automatic (Built-in)

- ✅ Authentication validated at startup (fails fast if misconfigured)
- ✅ SSH host key verification enabled
- ✅ Input validation for all user inputs
- ✅ SSRF protection for Vault addresses
- ✅ Path traversal protection for Vault paths
- ✅ Error sanitization (no internal details leaked)
- ✅ Security headers on all responses

### Required Configuration

- [ ] **Enable authentication**: Set `AUTH_ENABLED=true` with credentials
- [ ] **Strong credentials**: Use secure `AUTH_USERNAME` and `AUTH_PASSWORD`
- [ ] **CORS restricted**: Set `CORS_ALLOWED_ORIGINS` to your domain(s)
- [ ] **Encryption key backup**: Backup `.encryption_key` file

### Recommended

- [ ] **HTTPS enabled**: Use native TLS or reverse proxy
- [ ] **Audit logging enabled**: Set `AUDIT_LOG_PATH`
- [ ] **Log rotation configured**: Use logrotate for audit logs
- [ ] **Security scan**: Run `gosec ./...` or similar
- [ ] **Monitor logs**: Check for authentication failures
- [ ] **Firewall configured**: Restrict access to authorized IPs

### Production Environment Variables

```bash
# Authentication (REQUIRED)
AUTH_ENABLED=true
AUTH_USERNAME=admin
AUTH_PASSWORD=$(openssl rand -base64 32)

# CORS Policy (REQUIRED)
CORS_ALLOWED_ORIGINS=https://web-cli.yourdomain.com

# Encryption (REQUIRED)
ENCRYPTION_KEY=$(openssl rand -base64 32)

# TLS/HTTPS (RECOMMENDED)
WEBCLI_TLS_CERT_PATH=/etc/ssl/certs/web-cli.crt
WEBCLI_TLS_KEY_PATH=/etc/ssl/private/web-cli.key
WEBCLI_REQUIRE_HTTPS=true

# Audit Logging (RECOMMENDED)
WEBCLI_AUDIT_LOG_PATH=/var/log/web-cli/audit.log
```

---

For more configuration options, see [CONFIGURATION.md](CONFIGURATION.md).
For deployment instructions, see [DEPLOYMENT.md](DEPLOYMENT.md).
