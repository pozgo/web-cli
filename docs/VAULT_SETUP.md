# HashiCorp Vault Integration Setup Guide

This guide walks you through setting up HashiCorp Vault integration with Web CLI, allowing you to store and retrieve SSH keys, server configurations, environment variables, and scripts from Vault instead of (or in addition to) the local SQLite database.

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Vault Server Setup](#vault-server-setup)
3. [Enable KV Secrets Engine](#enable-kv-secrets-engine)
4. [Create Vault Policy](#create-vault-policy)
5. [Generate Access Token](#generate-access-token)
6. [Configure Web CLI](#configure-web-cli)
7. [Store Secrets in Vault](#store-secrets-in-vault)
8. [Troubleshooting](#troubleshooting)

---

## Prerequisites

- HashiCorp Vault server v1.0 or later
- Vault CLI installed (optional, for setup)
- Web CLI v0.3.0 or later
- Network access from Web CLI to Vault server

---

## Vault Server Setup

### Option 1: Development Server (Testing Only)

For testing purposes, you can start a Vault dev server:

```bash
vault server -dev
```

This creates an in-memory server with:
- Root token displayed in output
- Unsealed state
- KV v2 enabled at `secret/`

**WARNING**: Dev server data is lost on restart. Do not use in production.

### Option 2: Production Server

For production, follow the [official Vault installation guide](https://developer.hashicorp.com/vault/docs/install).

Basic production setup:

```bash
# Initialize Vault (first time only)
vault operator init

# Store the unseal keys and root token securely!

# Unseal Vault (required after restart)
vault operator unseal <key1>
vault operator unseal <key2>
vault operator unseal <key3>

# Login with root token
vault login <root-token>
```

---

## Enable KV Secrets Engine

Web CLI uses the KV v2 secrets engine. If not already enabled:

```bash
# Login to Vault
vault login <token>

# Enable KV v2 at the default path
vault secrets enable -path=secret kv-v2

# Or at a custom path
vault secrets enable -path=web-cli-secrets kv-v2
```

Verify the secrets engine:

```bash
vault secrets list
```

You should see:

```
Path                 Type         Description
----                 ----         -----------
secret/              kv           key/value secret storage
```

---

## Create Vault Policy

Create a policy file `web-cli-policy.hcl`.

**Option A: Using a dedicated `web-cli` KV mount (recommended)**

If you created a separate KV v2 secrets engine mounted at `web-cli/`:

```hcl
# Web CLI Vault Policy
# For dedicated KV v2 mount at "web-cli/"

# SSH Keys
path "web-cli/data/ssh-keys/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "web-cli/metadata/ssh-keys/*" {
  capabilities = ["list", "read", "delete"]
}

# Servers
path "web-cli/data/servers/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "web-cli/metadata/servers/*" {
  capabilities = ["list", "read", "delete"]
}

# Environment Variables
path "web-cli/data/env/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "web-cli/metadata/env/*" {
  capabilities = ["list", "read", "delete"]
}

# Scripts
path "web-cli/data/scripts/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "web-cli/metadata/scripts/*" {
  capabilities = ["list", "read", "delete"]
}

# Token self-lookup (for connection testing)
path "auth/token/lookup-self" {
  capabilities = ["read"]
}
```

**Option B: Using the default `secret` mount**

If you're using the default `secret/` KV v2 secrets engine with `mount_path = "secret"` in web-cli config:

```hcl
# Web CLI Vault Policy
# For default KV v2 mount at "secret/" with mount_path="secret"

# SSH Keys
path "secret/data/ssh-keys/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "secret/metadata/ssh-keys/*" {
  capabilities = ["list", "read", "delete"]
}

# Servers
path "secret/data/servers/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "secret/metadata/servers/*" {
  capabilities = ["list", "read", "delete"]
}

# Environment Variables
path "secret/data/env/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "secret/metadata/env/*" {
  capabilities = ["list", "read", "delete"]
}

# Scripts
path "secret/data/scripts/*" {
  capabilities = ["create", "read", "update", "delete", "list"]
}

path "secret/metadata/scripts/*" {
  capabilities = ["list", "read", "delete"]
}

# Token self-lookup (for connection testing)
path "auth/token/lookup-self" {
  capabilities = ["read"]
}
```

Apply the policy:

```bash
vault policy write web-cli web-cli-policy.hcl
```

### Read-Only Policy (Optional)

If you want Web CLI to only read secrets (not create/update/delete):

```hcl
# Web CLI Read-Only Policy

path "secret/data/web-cli/*" {
  capabilities = ["read", "list"]
}

path "secret/metadata/web-cli/*" {
  capabilities = ["list", "read"]
}

path "auth/token/lookup-self" {
  capabilities = ["read"]
}
```

---

## Generate Access Token

### Option 1: Token with Policy

Create a token with the web-cli policy:

```bash
# Create token with default TTL
vault token create -policy=web-cli

# Create token with custom TTL (e.g., 30 days)
vault token create -policy=web-cli -ttl=720h

# Create token with no expiration (use with caution)
vault token create -policy=web-cli -no-default-policy -orphan -period=0
```

The output includes your token:

```
Key                  Value
---                  -----
token                hvs.CAESIJxxx...
token_accessor       xxx
token_duration       768h
token_renewable      true
token_policies       ["default" "web-cli"]
```

### Option 2: AppRole (Recommended for Production)

For production, consider using AppRole authentication:

```bash
# Enable AppRole
vault auth enable approle

# Create role for web-cli
vault write auth/approle/role/web-cli \
    token_policies="web-cli" \
    token_ttl=1h \
    token_max_ttl=24h \
    secret_id_ttl=0

# Get role ID
vault read auth/approle/role/web-cli/role-id

# Generate secret ID
vault write -f auth/approle/role/web-cli/secret-id

# Login with AppRole to get token
vault write auth/approle/login \
    role_id=<role-id> \
    secret_id=<secret-id>
```

---

## Configure Web CLI

### Via Admin Panel (Recommended)

1. Open Web CLI in your browser
2. Navigate to **Admin Panel**
3. Click the **Vault Integration** tab
4. Fill in the configuration:
   - **Vault Address**: `https://vault.example.com:8200`
   - **Vault Token**: Your token from the previous step
   - **Namespace**: (Optional) For Vault Enterprise
   - **Mount Path**: `secret` (default) or your custom path
5. Toggle **Enable Vault Integration** to ON
6. Click **Save Configuration**
7. Click **Test Connection** to verify

### Via API

```bash
curl -X POST http://localhost:7777/api/vault/config \
  -H "Content-Type: application/json" \
  -u "username:password" \
  -d '{
    "address": "https://vault.example.com:8200",
    "token": "hvs.your-vault-token-here",
    "namespace": "",
    "mount_path": "secret",
    "enabled": true
  }'
```

---

## Store Secrets in Vault

> **Note**: These examples assume a dedicated KV v2 mount at `web-cli/` with `mount_path = "web-cli"` in web-cli config.
> If using the default `secret/` mount with `mount_path = "secret"`, replace `web-cli/` with `secret/` in all commands.

### SSH Keys

Store an SSH key in the default group:

```bash
vault kv put web-cli/ssh-keys/default/my-server-key \
  private_key="$(cat ~/.ssh/id_rsa)" \
  created_at="$(date -u +%Y-%m-%dT%H:%M:%SZ)"
```

Store an SSH key in a custom group (e.g., "production"):

```bash
vault kv put web-cli/ssh-keys/production/prod-key \
  private_key="$(cat ~/.ssh/prod_id_rsa)"
```

Or using JSON:

```bash
vault kv put web-cli/ssh-keys/default/my-server-key - <<EOF
{
  "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n...\n-----END OPENSSH PRIVATE KEY-----",
  "created_at": "2025-01-18T12:00:00Z"
}
EOF
```

Or with a simple custom field name (web-cli auto-detects private keys):

```bash
vault kv put web-cli/ssh-keys/default/my-server-key \
  my-custom-key="$(cat ~/.ssh/id_rsa)"
```

### Servers

Store a server configuration in the default group:

```bash
vault kv put web-cli/servers/default/production-web \
  ip_address="192.168.1.100" \
  port=22 \
  username="deploy"
```

Store a server in a custom group (e.g., "staging"):

```bash
vault kv put web-cli/servers/staging/staging-web \
  ip_address="10.0.0.50" \
  port=22 \
  username="deploy"
```

Or simplified (web-cli accepts `ip`, `host`, or `address` as field names):

```bash
vault kv put web-cli/servers/default/production-web \
  host="192.168.1.100" \
  user="deploy"
```

### Environment Variables

Store an environment variable in the default group:

```bash
vault kv put web-cli/env/default/DATABASE_URL \
  value="postgres://user:pass@db.example.com:5432/myapp" \
  description="Production database connection string"
```

Store an environment variable in a custom group:

```bash
vault kv put web-cli/env/production/API_KEY \
  value="your-api-key-here" \
  description="Production API key"
```

### Scripts

Store a bash script in the default group:

```bash
vault kv put web-cli/scripts/default/deploy-app \
  content="#!/bin/bash\ncd /app && git pull && systemctl restart myapp" \
  description="Deploy application from git" \
  filename="deploy.sh"
```

Store a script in a custom group:

```bash
vault kv put web-cli/scripts/maintenance/backup-db \
  content="#!/bin/bash\npg_dump mydb > /backup/db-\$(date +%Y%m%d).sql" \
  description="Database backup script" \
  filename="backup.sh"
```

---

## Data Structure Reference

> **Note**: Replace `{mount}` with your KV mount path configured in web-cli (e.g., `web-cli` or `secret`)

### Groups/Categories

Web CLI v0.3.0+ supports organizing resources by groups. In Vault, groups are represented as subdirectories:

```
{mount}/data/ssh-keys/{group}/{name}
{mount}/data/servers/{group}/{name}
{mount}/data/env/{group}/{name}
{mount}/data/scripts/{group}/{name}
```

The default group is `default`. Example paths:
- `web-cli/data/ssh-keys/default/my-key`
- `web-cli/data/ssh-keys/production/prod-server-key`
- `web-cli/data/servers/staging/web-server-01`

### SSH Keys
Path: `{mount}/data/ssh-keys/{group}/{name}`

Web-cli supports flexible field names for SSH keys:
- **Recommended**: Use `private_key` field
- **Alternative**: Any field containing "PRIVATE KEY" in its value will be auto-detected

```json
{
  "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n...",
  "created_at": "2025-01-18T12:00:00Z"
}
```

Or simplified (field name can be any identifier):
```json
{
  "my-key-name": "-----BEGIN OPENSSH PRIVATE KEY-----\n..."
}
```

### Servers
Path: `{mount}/data/servers/{group}/{name}`

Web-cli supports flexible field names for servers:
- **IP/Host**: `ip_address`, `ip`, `host`, or `address`
- **Port**: `port` (defaults to 22 if missing)
- **Username**: `username` or `user`

```json
{
  "ip_address": "192.168.1.100",
  "port": 22,
  "username": "admin"
}
```

Or simplified:
```json
{
  "host": "192.168.1.100"
}
```

### Environment Variables
Path: `{mount}/data/env/{group}/{name}`

```json
{
  "value": "secret-value-here",
  "description": "Optional description"
}
```

### Scripts
Path: `{mount}/data/scripts/{group}/{name}`

```json
{
  "content": "#!/bin/bash\necho 'Hello World'",
  "description": "Example script",
  "filename": "hello.sh"
}
```

---

## Viewing Secrets

> **Note**: Replace `web-cli/` with your mount path (e.g., `secret/` if using default mount)

### List All Groups for SSH Keys

```bash
vault kv list web-cli/ssh-keys
```

### List SSH Keys in a Group

```bash
vault kv list web-cli/ssh-keys/default
vault kv list web-cli/ssh-keys/production
```

### Read a Specific Secret

```bash
vault kv get web-cli/ssh-keys/default/my-server-key
```

### Delete a Secret

```bash
vault kv delete web-cli/ssh-keys/default/old-key
```

---

## Troubleshooting

### Connection Failed

**Error**: "vault connection test failed: connection refused"

**Solutions**:
- Verify Vault server is running: `vault status`
- Check the Vault address is correct and accessible
- Ensure firewall allows traffic on Vault port (default: 8200)

### Permission Denied

**Error**: "permission denied" when listing secrets

**Solutions**:
- Verify token has correct policy: `vault token lookup`
- Check policy grants `list` capability on metadata path
- Ensure the mount path in Web CLI matches your Vault configuration

### Token Expired

**Error**: "token is expired"

**Solutions**:
- Generate a new token with the policy
- For long-running deployments, use periodic tokens or AppRole

### Sealed Vault

**Error**: "vault is sealed"

**Solutions**:
- Unseal Vault: `vault operator unseal <key>`
- Configure auto-unseal for production (HSM, cloud KMS, etc.)

### Wrong Mount Path

**Error**: "no secrets at path"

**Solutions**:
- Verify mount path: `vault secrets list`
- Update Web CLI configuration with correct mount path
- Ensure KV v2 is enabled (not KV v1)

### Self-Signed Certificate

**Error**: "x509: certificate signed by unknown authority"

**Solutions**:
- Add CA certificate to system trust store
- Or use `VAULT_SKIP_VERIFY=true` (not recommended for production)

---

## Security Best Practices

1. **Use HTTPS**: Always use TLS in production
2. **Minimal Permissions**: Use the read-only policy if Web CLI doesn't need to write
3. **Token Rotation**: Regularly rotate Vault tokens
4. **Audit Logging**: Enable Vault audit logging for compliance
5. **Network Isolation**: Restrict network access to Vault server
6. **Secret Versioning**: Use KV v2 for secret versioning and soft-delete

---

## Support

For issues with:
- **Web CLI Vault integration**: Open an issue on [GitHub](https://github.com/pozgo/web-cli/issues)
- **HashiCorp Vault**: See [Vault documentation](https://developer.hashicorp.com/vault/docs)
