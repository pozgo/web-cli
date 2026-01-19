# Deployment Guide

This document covers various deployment options for Web CLI.

## Table of Contents

- [Docker (Recommended)](#docker-recommended)
- [Docker Compose](#docker-compose)
- [Linux Server (Binary)](#linux-server-binary)
- [systemd Service](#systemd-service)
- [Production Checklist](#production-checklist)

---

## Docker (Recommended)

Web CLI is available as a Docker image for easy deployment.

### Image Details

| Property | Value |
|----------|-------|
| Registry | Docker Hub ([`polinux/web-cli`](https://hub.docker.com/r/polinux/web-cli)) |
| Base Image | Debian Bookworm (slim) |
| Platforms | `linux/amd64`, `linux/arm64` |
| Size | ~100MB compressed |
| Tags | `latest`, `dev`, `main`, version tags (e.g., `v0.2.3`) |

### Quick Start

```bash
docker run -d \
  --name web-cli \
  -p 7777:7777 \
  -v web-cli-data:/data \
  -e AUTH_ENABLED=true \
  -e AUTH_USERNAME=admin \
  -e AUTH_PASSWORD=your-secure-password \
  polinux/web-cli:latest
```

Access at: `http://localhost:7777`

### With TLS/HTTPS

```bash
docker run -d \
  --name web-cli \
  -p 7777:7777 \
  -v web-cli-data:/data \
  -v ./certs:/certs:ro \
  -e WEBCLI_TLS_CERT_PATH=/certs/cert.pem \
  -e WEBCLI_TLS_KEY_PATH=/certs/key.pem \
  -e AUTH_ENABLED=true \
  -e AUTH_USERNAME=admin \
  -e AUTH_PASSWORD=your-secure-password \
  polinux/web-cli:latest
```

### Docker Volumes

| Path | Description |
|------|-------------|
| `/data` | Persistent data (database, encryption key) |
| `/config` | Configuration files (optional) |
| `/certs` | TLS certificates (optional) |

### Docker Environment Variables

| Variable | Default | Description |
|----------|---------|-------------|
| `WEBCLI_PORT` | `7777` | Port to listen on |
| `WEBCLI_HOST` | `0.0.0.0` | Host to bind to |
| `WEBCLI_DATABASE_PATH` | `/data/web-cli.db` | Database file path |
| `WEBCLI_ENCRYPTION_KEY_PATH` | `/data/.encryption_key` | Encryption key path |
| `ENCRYPTION_KEY` | (auto-generated) | Base64 encryption key |
| `AUTH_ENABLED` | `false` | Enable authentication |
| `AUTH_USERNAME` | `admin` | Basic auth username |
| `AUTH_PASSWORD` | (none) | Basic auth password |
| `AUTH_API_TOKEN` | (none) | Bearer token |
| `WEBCLI_TLS_CERT_PATH` | (none) | TLS certificate path |
| `WEBCLI_TLS_KEY_PATH` | (none) | TLS private key path |
| `WEBCLI_REQUIRE_HTTPS` | `false` | Require HTTPS |
| `CORS_ALLOWED_ORIGINS` | (localhost) | Allowed CORS origins |
| `WEBCLI_AUDIT_LOG_PATH` | (none) | Audit log file path |

---

## Docker Compose

### Standalone docker-compose.yml

Create a `docker-compose.yml` file anywhere on your system:

```yaml
services:
  web-cli:
    image: polinux/web-cli:latest
    container_name: web-cli
    restart: unless-stopped
    ports:
      - "7777:7777"
    volumes:
      - web-cli-data:/data
    environment:
      - AUTH_ENABLED=true
      - AUTH_USERNAME=admin
      - AUTH_PASSWORD=changeme123

volumes:
  web-cli-data:
```

### Run

```bash
# Start the container
docker compose up -d

# View logs
docker compose logs -f

# Stop
docker compose down
```

### Production docker-compose.yml

```yaml
services:
  web-cli:
    image: polinux/web-cli:latest
    container_name: web-cli
    restart: unless-stopped
    ports:
      - "7777:7777"
    volumes:
      - web-cli-data:/data
      - ./certs:/certs:ro
    environment:
      # Authentication
      - AUTH_ENABLED=true
      - AUTH_USERNAME=admin
      - AUTH_PASSWORD=${WEB_CLI_PASSWORD}

      # TLS/HTTPS
      - WEBCLI_TLS_CERT_PATH=/certs/cert.pem
      - WEBCLI_TLS_KEY_PATH=/certs/key.pem
      - WEBCLI_REQUIRE_HTTPS=true

      # CORS
      - CORS_ALLOWED_ORIGINS=https://web-cli.yourdomain.com

      # Audit Logging
      - WEBCLI_AUDIT_LOG_PATH=/data/audit.log

volumes:
  web-cli-data:
```

### With .env File

Create `.env` file:

```bash
WEB_CLI_PASSWORD=your-secure-password-here
ENCRYPTION_KEY=your-base64-encryption-key
```

---

## Linux Server (Binary)

### Build and Deploy

```bash
# Build for Linux
./build.sh all

# Copy to server
scp bin/web-cli-linux-x64 user@server:/opt/web-cli/web-cli

# Run on server
ssh user@server
cd /opt/web-cli
./web-cli
```

### Directory Structure

```
/opt/web-cli/
├── web-cli              # Binary
├── data/
│   └── web-cli.db       # Database
├── .encryption_key      # Encryption key
└── config.yaml          # Configuration (optional)
```

### Permissions

```bash
# Create web-cli user
sudo useradd -r -s /bin/false web-cli

# Set ownership
sudo chown -R web-cli:web-cli /opt/web-cli

# Secure encryption key
chmod 600 /opt/web-cli/.encryption_key
```

---

## systemd Service

Create `/etc/systemd/system/web-cli.service`:

```ini
[Unit]
Description=Web CLI Service
After=network.target

[Service]
Type=simple
User=www-data
WorkingDirectory=/opt/web-cli
ExecStart=/opt/web-cli/web-cli
Restart=on-failure
RestartSec=5

# Security Configuration (REQUIRED for production)
Environment=AUTH_ENABLED=true
Environment=AUTH_USERNAME=admin
Environment=AUTH_PASSWORD=your-secure-password-here
Environment=CORS_ALLOWED_ORIGINS=https://web-cli.yourdomain.com
Environment=ENCRYPTION_KEY=your-encryption-key-here

# Audit Logging (recommended)
Environment=AUDIT_LOG_PATH=/var/log/web-cli/audit.log

# TLS/HTTPS (recommended)
# Environment=WEBCLI_TLS_CERT_PATH=/etc/ssl/certs/web-cli.crt
# Environment=WEBCLI_TLS_KEY_PATH=/etc/ssl/private/web-cli.key
# Environment=WEBCLI_REQUIRE_HTTPS=true

# Custom timeouts (optional, values in seconds)
# Environment=COMMAND_TIMEOUT=600
# Environment=VAULT_TIMEOUT=60

[Install]
WantedBy=multi-user.target
```

### Enable and Start

```bash
# Create log directory
sudo mkdir -p /var/log/web-cli
sudo chown www-data:www-data /var/log/web-cli

# Enable and start service
sudo systemctl daemon-reload
sudo systemctl enable web-cli
sudo systemctl start web-cli

# Check status
sudo systemctl status web-cli

# View logs
sudo journalctl -u web-cli -f
```

---

## Production Checklist

### Before Deployment

- [ ] Generate strong password: `openssl rand -base64 32`
- [ ] Generate encryption key: `openssl rand -base64 32`
- [ ] Obtain TLS certificate (Let's Encrypt or similar)
- [ ] Configure firewall rules

### Environment Variables

```bash
# Required
AUTH_ENABLED=true
AUTH_USERNAME=admin
AUTH_PASSWORD=$(openssl rand -base64 32)
ENCRYPTION_KEY=$(openssl rand -base64 32)
CORS_ALLOWED_ORIGINS=https://web-cli.yourdomain.com

# Recommended
WEBCLI_TLS_CERT_PATH=/etc/ssl/certs/web-cli.crt
WEBCLI_TLS_KEY_PATH=/etc/ssl/private/web-cli.key
WEBCLI_REQUIRE_HTTPS=true
WEBCLI_AUDIT_LOG_PATH=/var/log/web-cli/audit.log
```

### Post-Deployment

- [ ] Verify authentication is working
- [ ] Test HTTPS certificate
- [ ] Configure log rotation
- [ ] Set up monitoring/alerting
- [ ] Backup encryption key securely
- [ ] Document admin credentials securely

### Backup Strategy

Critical files to backup:

| File | Importance | Description |
|------|------------|-------------|
| `.encryption_key` | **Critical** | Cannot recover data without it |
| `web-cli.db` | Important | All configuration and history |
| `config.yaml` | Optional | Configuration file |

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

For more configuration options, see [CONFIGURATION.md](CONFIGURATION.md).
For security information, see [SECURITY.md](SECURITY.md).
