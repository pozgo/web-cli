# Configuration Guide

Web CLI supports multiple configuration methods with the following priority order:

1. **Command-line flags** (highest priority)
2. **Environment variables**
3. **Configuration file**
4. **Default values** (lowest priority)

## Table of Contents

- [Command-Line Flags](#command-line-flags)
- [Environment Variables](#environment-variables)
- [Configuration File](#configuration-file)
- [Timeout Configuration](#timeout-configuration)
- [Audit Logging](#audit-logging)
- [TLS/HTTPS Configuration](#tlshttps-configuration)
- [CORS Configuration](#cors-configuration)

---

## Command-Line Flags

```bash
./web-cli [options]

Options:
  -port int              Port to listen on (default: 7777)
  -host string           Host to bind to (default: 0.0.0.0)
  -frontend string       Path to frontend build files (default: ./frontend/dist)
  -db string             Path to database file (default: ./data/web-cli.db)
  -encryption-key string Path to encryption key file (default: ./.encryption_key)
  -tls-cert string       Path to TLS certificate file (enables HTTPS)
  -tls-key string        Path to TLS private key file
  -require-https         Require HTTPS when auth is enabled (reject HTTP requests)
```

---

## Environment Variables

All configuration options can be set via environment variables. Both standard and `WEBCLI_`-prefixed variables are supported (prefixed is recommended).

### Server Configuration

| Variable | WEBCLI Prefix | Default | Description |
|----------|---------------|---------|-------------|
| `PORT` | `WEBCLI_PORT` | `7777` | Port to listen on |
| `HOST` | `WEBCLI_HOST` | `0.0.0.0` | Host to bind to |
| `FRONTEND_PATH` | `WEBCLI_FRONTEND_PATH` | `./frontend/dist` | Frontend build files |
| `DATABASE_PATH` | `WEBCLI_DATABASE_PATH` | `./data/web-cli.db` | SQLite database path |
| `ENCRYPTION_KEY_PATH` | `WEBCLI_ENCRYPTION_KEY_PATH` | `./.encryption_key` | Encryption key file |

### Authentication

| Variable | Default | Description |
|----------|---------|-------------|
| `AUTH_ENABLED` | `false` | Enable authentication |
| `AUTH_USERNAME` | (none) | Basic auth username |
| `AUTH_PASSWORD` | (none) | Basic auth password |
| `AUTH_API_TOKEN` | (none) | Bearer token for API access |

### TLS/HTTPS

| Variable | WEBCLI Prefix | Default | Description |
|----------|---------------|---------|-------------|
| `TLS_CERT_PATH` | `WEBCLI_TLS_CERT_PATH` | (none) | TLS certificate file |
| `TLS_KEY_PATH` | `WEBCLI_TLS_KEY_PATH` | (none) | TLS private key file |
| `REQUIRE_HTTPS` | `WEBCLI_REQUIRE_HTTPS` | `false` | Reject HTTP when auth enabled |

### Example Usage

```bash
# Standard environment variables
PORT=8080 HOST=localhost ./web-cli

# WEBCLI-prefixed variables (recommended)
WEBCLI_PORT=8080 \
WEBCLI_HOST=localhost \
WEBCLI_DATABASE_PATH=/var/lib/web-cli/web-cli.db \
./web-cli

# Production example with authentication
AUTH_ENABLED=true \
AUTH_USERNAME=admin \
AUTH_PASSWORD="secure-password" \
WEBCLI_TLS_CERT_PATH=/etc/ssl/certs/web-cli.crt \
WEBCLI_TLS_KEY_PATH=/etc/ssl/private/web-cli.key \
./web-cli
```

---

## Configuration File

Web CLI supports configuration files in YAML, JSON, or TOML format.

### Search Locations

Configuration files are searched in the following order (first found is used):

1. `./config.yaml` (current directory)
2. `./config/config.yaml` (config subdirectory)
3. `/etc/web-cli/config.yaml` (system config)
4. `~/.config/web-cli/config.yaml` (user config)

### Example config.yaml

```yaml
# Server Configuration
port: 7777
host: "0.0.0.0"
frontend_path: "./assets/frontend"
database_path: "./data/web-cli.db"
encryption_key_path: "./.encryption_key"

# TLS/HTTPS Configuration
tls_cert_path: "/etc/ssl/certs/web-cli.crt"
tls_key_path: "/etc/ssl/private/web-cli.key"
require_https: true

# Timeout Configuration (in seconds)
read_timeout: 30
write_timeout: 600
idle_timeout: 60
vault_timeout: 30
command_timeout: 300
ssh_connect_timeout: 30

# Audit Logging (optional)
audit_log_path: "/var/log/web-cli/audit.log"
```

### Example config.json

```json
{
  "port": 7777,
  "host": "0.0.0.0",
  "database_path": "./data/web-cli.db",
  "encryption_key_path": "./.encryption_key",
  "read_timeout": 30,
  "write_timeout": 600,
  "command_timeout": 300
}
```

---

## Timeout Configuration

All timeout values are configurable via environment variables (values in seconds):

| Timeout | Default | Environment Variable | Description |
|---------|---------|---------------------|-------------|
| Read Timeout | 30s | `READ_TIMEOUT` or `WEBCLI_READ_TIMEOUT` | Time to read request headers |
| Write Timeout | 600s (10m) | `WRITE_TIMEOUT` or `WEBCLI_WRITE_TIMEOUT` | Time to write response (high for streaming) |
| Idle Timeout | 60s | `IDLE_TIMEOUT` or `WEBCLI_IDLE_TIMEOUT` | Keep-alive connection timeout |
| Vault Timeout | 30s | `VAULT_TIMEOUT` or `WEBCLI_VAULT_TIMEOUT` | HashiCorp Vault operations |
| Command Timeout | 300s (5m) | `COMMAND_TIMEOUT` or `WEBCLI_COMMAND_TIMEOUT` | Command execution timeout |
| SSH Connect Timeout | 30s | `SSH_CONNECT_TIMEOUT` or `WEBCLI_SSH_CONNECT_TIMEOUT` | SSH connection establishment |

### Example

```bash
# Increase command timeout to 10 minutes
export COMMAND_TIMEOUT=600

# Increase Vault timeout for slow connections
export VAULT_TIMEOUT=60

./web-cli
```

---

## Audit Logging

Enable comprehensive audit logging for security compliance and monitoring.

### Enable Audit Logging

```bash
# Via environment variable
export AUDIT_LOG_PATH=/var/log/web-cli/audit.log
./web-cli

# Or via WEBCLI prefix
export WEBCLI_AUDIT_LOG_PATH=/var/log/web-cli/audit.log
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

Use logrotate to manage log file size:

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

## TLS/HTTPS Configuration

### Enable TLS

```bash
# Using command-line flags
./web-cli -tls-cert /path/to/cert.pem -tls-key /path/to/key.pem

# Using environment variables
WEBCLI_TLS_CERT_PATH=/path/to/cert.pem \
WEBCLI_TLS_KEY_PATH=/path/to/key.pem \
./web-cli

# Enforce HTTPS (reject HTTP requests)
./web-cli -tls-cert cert.pem -tls-key key.pem -require-https
```

### Generate Self-Signed Certificate (Testing)

```bash
openssl req -x509 -newkey rsa:4096 -keyout key.pem -out cert.pem \
  -days 365 -nodes -subj "/CN=localhost"
```

### Features

- Native Go TLS implementation (no reverse proxy required)
- Automatic HTTPS when certificate and key are provided
- Optional HTTPS enforcement (rejects HTTP requests)
- Works with any TLS certificate (self-signed, Let's Encrypt, etc.)

---

## CORS Configuration

Configure Cross-Origin Resource Sharing for frontend access.

### Default Behavior

By default, only localhost origins are allowed:
- `http://localhost:7777`
- `http://127.0.0.1:7777`

### Custom Origins

```bash
# Single origin
export CORS_ALLOWED_ORIGINS="https://web-cli.example.com"

# Multiple origins (comma-separated)
export CORS_ALLOWED_ORIGINS="https://web-cli.example.com,https://admin.example.com"

./web-cli
```

---

## Complete Production Example

```bash
# /etc/systemd/system/web-cli.service environment
AUTH_ENABLED=true
AUTH_USERNAME=admin
AUTH_PASSWORD=your-secure-password
ENCRYPTION_KEY=your-base64-encryption-key
CORS_ALLOWED_ORIGINS=https://web-cli.yourdomain.com
WEBCLI_TLS_CERT_PATH=/etc/ssl/certs/web-cli.crt
WEBCLI_TLS_KEY_PATH=/etc/ssl/private/web-cli.key
WEBCLI_REQUIRE_HTTPS=true
WEBCLI_AUDIT_LOG_PATH=/var/log/web-cli/audit.log
WEBCLI_DATABASE_PATH=/var/lib/web-cli/web-cli.db
```

For more information on security configuration, see [SECURITY.md](SECURITY.md).
For deployment instructions, see [DEPLOYMENT.md](DEPLOYMENT.md).
