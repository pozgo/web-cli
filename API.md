# Web CLI API Documentation

Complete API reference for the Web CLI application. All endpoints return JSON responses unless otherwise specified.

## Table of Contents

- [Quick Reference](#quick-reference)
- [Authentication](#authentication)
- [Health Check](#health-check)
- [SSH Keys Management](#ssh-keys-management)
- [Server Management](#server-management)
- [Local Users Management](#local-users-management)
- [System Information](#system-information)
- [Command Execution](#command-execution)
- [Saved Commands Management](#saved-commands-management)
- [Command History](#command-history)
- [Environment Variables Management](#environment-variables-management)
- [Bash Scripts Management](#bash-scripts-management)
- [Script Presets Management](#script-presets-management)
- [Error Responses](#error-responses)
- [Security Considerations](#security-considerations)

## Base URL

```
http://localhost:7777/api
```

Default port is `7777`, configurable via `-port` flag or `PORT` environment variable.

## Quick Reference

| Endpoint | Method | Description |
|----------|--------|-------------|
| `/health` | GET | Server health check |
| `/keys` | GET | List all SSH keys |
| `/keys` | POST | Create SSH key |
| `/keys/{id}` | GET | Get single SSH key |
| `/keys/{id}` | PUT | Update SSH key |
| `/keys/{id}` | DELETE | Delete SSH key |
| `/servers` | GET | List all servers |
| `/servers` | POST | Create server |
| `/servers/{id}` | GET | Get single server |
| `/servers/{id}` | PUT | Update server |
| `/servers/{id}` | DELETE | Delete server |
| `/local-users` | GET | List all local users |
| `/local-users` | POST | Create local user |
| `/local-users/{id}` | GET | Get single local user |
| `/local-users/{id}` | PUT | Update local user |
| `/local-users/{id}` | DELETE | Delete local user |
| `/system/current-user` | GET | Get current system user |
| `/commands/execute` | POST | Execute command (local/remote) |
| `/saved-commands` | GET | List all saved commands |
| `/saved-commands` | POST | Create saved command |
| `/saved-commands/{id}` | GET | Get single saved command |
| `/saved-commands/{id}` | PUT | Update saved command |
| `/saved-commands/{id}` | DELETE | Delete saved command |
| `/history` | GET | List command history |
| `/history/{id}` | GET | Get single history entry |
| `/env-variables` | GET | List all environment variables |
| `/env-variables` | POST | Create environment variable |
| `/env-variables/{id}` | GET | Get single environment variable |
| `/env-variables/{id}` | PUT | Update environment variable |
| `/env-variables/{id}` | DELETE | Delete environment variable |
| `/bash-scripts` | GET | List all bash scripts |
| `/bash-scripts` | POST | Create bash script |
| `/bash-scripts/{id}` | GET | Get single bash script |
| `/bash-scripts/{id}` | PUT | Update bash script |
| `/bash-scripts/{id}` | DELETE | Delete bash script |
| `/bash-scripts/execute` | POST | Execute a bash script |
| `/bash-scripts/{id}/presets` | GET | Get presets for a script |
| `/script-presets` | GET | List all script presets |
| `/script-presets` | POST | Create script preset |
| `/script-presets/{id}` | GET | Get single script preset |
| `/script-presets/{id}` | PUT | Update script preset |
| `/script-presets/{id}` | DELETE | Delete script preset |

## Authentication

Authentication is **disabled by default** for development convenience. In production, enable authentication using environment variables.

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

### Using Authentication

**HTTP Basic Auth:**
```bash
curl -u admin:password http://localhost:7777/api/health
```

**Bearer Token:**
```bash
curl -H "Authorization: Bearer your-token" http://localhost:7777/api/health
```

### Security Features
- Constant-time credential comparison (prevents timing attacks)
- Supports both Basic Auth and Bearer token simultaneously
- Bearer token takes precedence if both are provided
- bcrypt password hashing for stored credentials

**Note**: In production environments, always enable authentication and use HTTPS.

---

## Health Check

### Get Server Health Status

Check if the server is running and responsive.

**Endpoint**: `GET /health`

**Response**: `200 OK`

```json
{
  "status": "ok"
}
```

**Example**:

```bash
curl http://localhost:7777/api/health
```

---

## SSH Keys Management

Manage SSH private keys used for remote server authentication. All keys are encrypted with AES-256-GCM before storage.

### List All SSH Keys

Retrieve all stored SSH keys.

**Endpoint**: `GET /keys`

**Response**: `200 OK`

```json
[
  {
    "id": 1,
    "name": "production-server-key",
    "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtz...\n-----END OPENSSH PRIVATE KEY-----",
    "created_at": "2025-11-10T12:00:00Z",
    "updated_at": "2025-11-10T12:00:00Z"
  },
  {
    "id": 2,
    "name": "dev-server-key",
    "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtz...\n-----END OPENSSH PRIVATE KEY-----",
    "created_at": "2025-11-10T13:00:00Z",
    "updated_at": "2025-11-10T13:00:00Z"
  }
]
```

**Example**:

```bash
curl http://localhost:7777/api/keys
```

---

### Get Single SSH Key

Retrieve a specific SSH key by ID.

**Endpoint**: `GET /keys/{id}`

**Path Parameters**:
- `id` (integer, required): SSH key ID

**Response**: `200 OK`

```json
{
  "id": 1,
  "name": "production-server-key",
  "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtz...\n-----END OPENSSH PRIVATE KEY-----",
  "created_at": "2025-11-10T12:00:00Z",
  "updated_at": "2025-11-10T12:00:00Z"
}
```

**Error Responses**:
- `404 Not Found`: SSH key not found

**Example**:

```bash
curl http://localhost:7777/api/keys/1
```

---

### Create SSH Key

Add a new SSH private key to the system.

**Endpoint**: `POST /keys`

**Request Headers**:
- `Content-Type: application/json`

**Request Body**:

```json
{
  "name": "new-server-key",
  "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtz...\n-----END OPENSSH PRIVATE KEY-----"
}
```

**Fields**:
- `name` (string, required): Descriptive name for the SSH key
- `private_key` (string, required): PEM-encoded SSH private key

**Response**: `201 Created`

```json
{
  "id": 3,
  "name": "new-server-key",
  "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtz...\n-----END OPENSSH PRIVATE KEY-----",
  "created_at": "2025-11-11T10:00:00Z",
  "updated_at": "2025-11-11T10:00:00Z"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid request body or missing required fields

**Example**:

```bash
curl -X POST http://localhost:7777/api/keys \
  -H "Content-Type: application/json" \
  -d '{
    "name": "new-server-key",
    "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n...\n-----END OPENSSH PRIVATE KEY-----"
  }'
```

---

### Update SSH Key

Update an existing SSH key.

**Endpoint**: `PUT /keys/{id}`

**Path Parameters**:
- `id` (integer, required): SSH key ID

**Request Headers**:
- `Content-Type: application/json`

**Request Body**:

```json
{
  "name": "updated-server-key",
  "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtz...\n-----END OPENSSH PRIVATE KEY-----"
}
```

**Fields**:
- `name` (string, required): Updated name for the SSH key
- `private_key` (string, required): Updated PEM-encoded SSH private key

**Response**: `200 OK`

```json
{
  "id": 1,
  "name": "updated-server-key",
  "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\nb3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtz...\n-----END OPENSSH PRIVATE KEY-----",
  "created_at": "2025-11-10T12:00:00Z",
  "updated_at": "2025-11-11T11:00:00Z"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid request body
- `404 Not Found`: SSH key not found

**Example**:

```bash
curl -X PUT http://localhost:7777/api/keys/1 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "updated-server-key",
    "private_key": "-----BEGIN OPENSSH PRIVATE KEY-----\n...\n-----END OPENSSH PRIVATE KEY-----"
  }'
```

---

### Delete SSH Key

Delete an SSH key from the system.

**Endpoint**: `DELETE /keys/{id}`

**Path Parameters**:
- `id` (integer, required): SSH key ID

**Response**: `204 No Content`

**Error Responses**:
- `404 Not Found`: SSH key not found

**Example**:

```bash
curl -X DELETE http://localhost:7777/api/keys/1
```

---

## Server Management

Manage remote servers for SSH connections.

### List All Servers

Retrieve all configured servers.

**Endpoint**: `GET /servers`

**Response**: `200 OK`

```json
[
  {
    "id": 1,
    "name": "production-server",
    "ip_address": "192.168.1.100",
    "port": 22,
    "username": "admin",
    "created_at": "2025-11-10T12:00:00Z",
    "updated_at": "2025-11-10T12:00:00Z"
  },
  {
    "id": 2,
    "name": "dev-server",
    "ip_address": "192.168.1.101",
    "port": 2222,
    "username": "ubuntu",
    "created_at": "2025-11-10T12:05:00Z",
    "updated_at": "2025-11-10T12:05:00Z"
  }
]
```

**Example**:

```bash
curl http://localhost:7777/api/servers
```

---

### Get Single Server

Retrieve a specific server by ID.

**Endpoint**: `GET /servers/{id}`

**Path Parameters**:
- `id` (integer, required): Server ID

**Response**: `200 OK`

```json
{
  "id": 1,
  "name": "production-server",
  "ip_address": "192.168.1.100",
  "port": 22,
  "username": "admin",
  "created_at": "2025-11-10T12:00:00Z",
  "updated_at": "2025-11-10T12:00:00Z"
}
```

**Error Responses**:
- `404 Not Found`: Server not found

**Example**:

```bash
curl http://localhost:7777/api/servers/1
```

---

### Create Server

Add a new server configuration.

**Endpoint**: `POST /servers`

**Request Headers**:
- `Content-Type: application/json`

**Request Body**:

```json
{
  "name": "staging-server",
  "ip_address": "192.168.1.102",
  "port": 22,
  "username": "deploy"
}
```

**Fields**:
- `name` (string, optional): Descriptive server name (must follow hostname conventions if provided)
- `ip_address` (string, optional): Server IP address or hostname
- `port` (integer, optional): SSH port number (default: 22)
- `username` (string, optional): SSH username (default: "root")

**Note**: At least one of `name` or `ip_address` must be provided.

**Response**: `201 Created`

```json
{
  "id": 3,
  "name": "staging-server",
  "ip_address": "192.168.1.102",
  "port": 22,
  "username": "deploy",
  "created_at": "2025-11-11T10:00:00Z",
  "updated_at": "2025-11-11T10:00:00Z"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid request body or validation error

**Example**:

```bash
curl -X POST http://localhost:7777/api/servers \
  -H "Content-Type: application/json" \
  -d '{
    "name": "staging-server",
    "ip_address": "192.168.1.102",
    "port": 22,
    "username": "deploy"
  }'
```

---

### Update Server

Update an existing server configuration.

**Endpoint**: `PUT /servers/{id}`

**Path Parameters**:
- `id` (integer, required): Server ID

**Request Headers**:
- `Content-Type: application/json`

**Request Body**:

```json
{
  "name": "production-server-updated",
  "ip_address": "192.168.1.100",
  "port": 2222,
  "username": "root"
}
```

**Fields**: All fields are optional; only provided fields will be updated.

**Response**: `200 OK`

```json
{
  "id": 1,
  "name": "production-server-updated",
  "ip_address": "192.168.1.100",
  "port": 2222,
  "username": "root",
  "created_at": "2025-11-10T12:00:00Z",
  "updated_at": "2025-11-11T11:00:00Z"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid request body
- `404 Not Found`: Server not found

**Example**:

```bash
curl -X PUT http://localhost:7777/api/servers/1 \
  -H "Content-Type: application/json" \
  -d '{
    "port": 2222
  }'
```

---

### Delete Server

Delete a server configuration.

**Endpoint**: `DELETE /servers/{id}`

**Path Parameters**:
- `id` (integer, required): Server ID

**Response**: `204 No Content`

**Error Responses**:
- `404 Not Found`: Server not found

**Example**:

```bash
curl -X DELETE http://localhost:7777/api/servers/1
```

---

## Local Users Management

Manage local user accounts that can be used for command execution.

### List All Local Users

Retrieve all stored local users.

**Endpoint**: `GET /local-users`

**Response**: `200 OK`

```json
[
  {
    "id": 1,
    "name": "admin",
    "created_at": "2025-11-10T12:00:00Z",
    "updated_at": "2025-11-10T12:00:00Z"
  },
  {
    "id": 2,
    "name": "deploy",
    "created_at": "2025-11-10T13:00:00Z",
    "updated_at": "2025-11-10T13:00:00Z"
  }
]
```

**Example**:

```bash
curl http://localhost:7777/api/local-users
```

---

### Get Single Local User

Retrieve a specific local user by ID.

**Endpoint**: `GET /local-users/{id}`

**Path Parameters**:
- `id` (integer, required): Local user ID

**Response**: `200 OK`

```json
{
  "id": 1,
  "name": "admin",
  "created_at": "2025-11-10T12:00:00Z",
  "updated_at": "2025-11-10T12:00:00Z"
}
```

**Error Responses**:
- `404 Not Found`: Local user not found

**Example**:

```bash
curl http://localhost:7777/api/local-users/1
```

---

### Create Local User

Add a new local user to the system.

**Endpoint**: `POST /local-users`

**Request Headers**:
- `Content-Type: application/json`

**Request Body**:

```json
{
  "name": "jenkins"
}
```

**Fields**:
- `name` (string, required): Username (must be unique)

**Response**: `201 Created`

```json
{
  "id": 3,
  "name": "jenkins",
  "created_at": "2025-11-11T10:00:00Z",
  "updated_at": "2025-11-11T10:00:00Z"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid request body or duplicate username

**Example**:

```bash
curl -X POST http://localhost:7777/api/local-users \
  -H "Content-Type: application/json" \
  -d '{
    "name": "jenkins"
  }'
```

---

### Update Local User

Update an existing local user.

**Endpoint**: `PUT /local-users/{id}`

**Path Parameters**:
- `id` (integer, required): Local user ID

**Request Headers**:
- `Content-Type: application/json`

**Request Body**:

```json
{
  "name": "jenkins-updated"
}
```

**Response**: `200 OK`

```json
{
  "id": 3,
  "name": "jenkins-updated",
  "created_at": "2025-11-11T10:00:00Z",
  "updated_at": "2025-11-11T11:00:00Z"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid request body
- `404 Not Found`: Local user not found

**Example**:

```bash
curl -X PUT http://localhost:7777/api/local-users/3 \
  -H "Content-Type: application/json" \
  -d '{
    "name": "jenkins-updated"
  }'
```

---

### Delete Local User

Delete a local user from the system.

**Endpoint**: `DELETE /local-users/{id}`

**Path Parameters**:
- `id` (integer, required): Local user ID

**Response**: `204 No Content`

**Error Responses**:
- `404 Not Found`: Local user not found

**Example**:

```bash
curl -X DELETE http://localhost:7777/api/local-users/3
```

---

## System Information

Retrieve system information about the server running the application.

### Get Current System User

Retrieve information about the user running the Web CLI application.

**Endpoint**: `GET /system/current-user`

**Response**: `200 OK`

```json
{
  "username": "jdoe",
  "uid": "1000",
  "gid": "1000",
  "name": "John Doe",
  "home_dir": "/home/jdoe"
}
```

**Fields**:
- `username` (string): System username
- `uid` (string): User ID
- `gid` (string): Group ID
- `name` (string): Full name
- `home_dir` (string): Home directory path

**Example**:

```bash
curl http://localhost:7777/api/system/current-user
```

---

## Command Execution

Execute commands locally or on remote servers via SSH.

### Execute Command

Execute a bash command either locally or on a remote server.

**Endpoint**: `POST /commands/execute`

**Request Headers**:
- `Content-Type: application/json`

**Request Body (Local Execution)**:

```json
{
  "command": "ls -la /tmp",
  "user": "root",
  "sudo_password": "your-sudo-password",
  "save_as": "list-tmp"
}
```

**Request Body (Remote Execution)**:

```json
{
  "command": "uptime",
  "user": "root",
  "is_remote": true,
  "server_id": 1,
  "ssh_key_id": 2,
  "ssh_password": "fallback-password-if-needed",
  "save_as": "check-uptime"
}
```

**Fields**:
- `command` (string, required): Bash command to execute
- `user` (string, optional): User to run as (`root`, `current`, or custom username). Default: `"root"`
- `sudo_password` (string, optional): Sudo password for local root execution
- `ssh_password` (string, optional): SSH password for remote execution (fallback if key auth fails). **Never stored in history**
- `is_remote` (boolean, optional): Set to `true` for remote execution. Default: `false`
- `server_id` (integer, optional): Server ID for remote execution (required if `is_remote` is `true`)
- `ssh_key_id` (integer, optional): SSH key ID for remote authentication (optional)
- `save_as` (string, optional): Save command as template with this name

**Response**: `200 OK`

```json
{
  "command": "uptime",
  "output": " 13:46:21 up 5 days,  3:21,  2 users,  load average: 0.52, 0.58, 0.59",
  "exit_code": 0,
  "user": "root",
  "execution_time_ms": 245,
  "executed_at": "2025-11-11T13:46:21Z"
}
```

**Fields**:
- `command` (string): Executed command
- `output` (string): Combined stdout and stderr output
- `exit_code` (integer): Command exit code (0 = success)
- `user` (string): User who executed the command
- `execution_time_ms` (integer): Execution time in milliseconds
- `executed_at` (string): Timestamp of execution (ISO 8601 format)

**Error Responses**:
- `400 Bad Request`: Invalid request body or missing required fields
- `404 Not Found`: Server or SSH key not found (for remote execution)
- `500 Internal Server Error`: Command execution failed

**Security Notes**:
- Commands are automatically saved to history
- **SSH passwords are NEVER stored in history** (security feature)
- Sudo passwords are required for local root execution
- SSH key authentication is preferred over password authentication

**Example (Local)**:

```bash
curl -X POST http://localhost:7777/api/commands/execute \
  -H "Content-Type: application/json" \
  -d '{
    "command": "whoami",
    "user": "current"
  }'
```

**Example (Remote with SSH Key)**:

```bash
curl -X POST http://localhost:7777/api/commands/execute \
  -H "Content-Type: application/json" \
  -d '{
    "command": "df -h",
    "user": "root",
    "is_remote": true,
    "server_id": 1,
    "ssh_key_id": 2
  }'
```

**Example (Remote with Password Fallback)**:

```bash
curl -X POST http://localhost:7777/api/commands/execute \
  -H "Content-Type: application/json" \
  -d '{
    "command": "uptime",
    "user": "admin",
    "is_remote": true,
    "server_id": 1,
    "ssh_password": "secure-password"
  }'
```

---

## Saved Commands Management

Manage reusable command templates for both local and remote execution.

### List All Saved Commands

Retrieve all saved command templates.

**Endpoint**: `GET /saved-commands`

**Response**: `200 OK`

```json
[
  {
    "id": 1,
    "name": "check-disk-space",
    "command": "df -h",
    "description": "Check disk space usage",
    "user": "root",
    "is_remote": false,
    "server_id": null,
    "ssh_key_id": null,
    "created_at": "2025-11-10T12:00:00Z",
    "updated_at": "2025-11-10T12:00:00Z"
  },
  {
    "id": 2,
    "name": "remote-uptime",
    "command": "uptime",
    "description": "Check server uptime",
    "user": "admin",
    "is_remote": true,
    "server_id": 1,
    "ssh_key_id": 2,
    "created_at": "2025-11-10T13:00:00Z",
    "updated_at": "2025-11-10T13:00:00Z"
  }
]
```

**Example**:

```bash
curl http://localhost:7777/api/saved-commands
```

---

### Get Single Saved Command

Retrieve a specific saved command by ID.

**Endpoint**: `GET /saved-commands/{id}`

**Path Parameters**:
- `id` (integer, required): Saved command ID

**Response**: `200 OK`

```json
{
  "id": 1,
  "name": "check-disk-space",
  "command": "df -h",
  "description": "Check disk space usage",
  "user": "root",
  "is_remote": false,
  "server_id": null,
  "ssh_key_id": null,
  "created_at": "2025-11-10T12:00:00Z",
  "updated_at": "2025-11-10T12:00:00Z"
}
```

**Error Responses**:
- `404 Not Found`: Saved command not found

**Example**:

```bash
curl http://localhost:7777/api/saved-commands/1
```

---

### Create Saved Command

Create a new saved command template.

**Endpoint**: `POST /saved-commands`

**Request Headers**:
- `Content-Type: application/json`

**Request Body (Local Command)**:

```json
{
  "name": "list-processes",
  "command": "ps aux",
  "description": "List all running processes",
  "user": "root"
}
```

**Request Body (Remote Command)**:

```json
{
  "name": "remote-memory-check",
  "command": "free -h",
  "description": "Check memory usage on remote server",
  "user": "admin",
  "is_remote": true,
  "server_id": 1,
  "ssh_key_id": 2
}
```

**Fields**:
- `name` (string, required): Descriptive name for the command
- `command` (string, required): Bash command to execute
- `description` (string, optional): Additional description
- `user` (string, optional): User to run as. Default: `"root"`
- `is_remote` (boolean, optional): Whether this is a remote command. Default: `false`
- `server_id` (integer, optional): Server ID for remote commands
- `ssh_key_id` (integer, optional): SSH key ID for remote authentication

**Response**: `201 Created`

```json
{
  "id": 3,
  "name": "list-processes",
  "command": "ps aux",
  "description": "List all running processes",
  "user": "root",
  "is_remote": false,
  "server_id": null,
  "ssh_key_id": null,
  "created_at": "2025-11-11T10:00:00Z",
  "updated_at": "2025-11-11T10:00:00Z"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid request body or missing required fields

**Example**:

```bash
curl -X POST http://localhost:7777/api/saved-commands \
  -H "Content-Type: application/json" \
  -d '{
    "name": "list-processes",
    "command": "ps aux",
    "description": "List all running processes",
    "user": "root"
  }'
```

---

### Update Saved Command

Update an existing saved command template.

**Endpoint**: `PUT /saved-commands/{id}`

**Path Parameters**:
- `id` (integer, required): Saved command ID

**Request Headers**:
- `Content-Type: application/json`

**Request Body**:

```json
{
  "name": "updated-command-name",
  "command": "df -h /",
  "description": "Updated description",
  "user": "admin"
}
```

**Fields**: All fields are optional; only provided fields will be updated.

**Response**: `200 OK`

```json
{
  "id": 1,
  "name": "updated-command-name",
  "command": "df -h /",
  "description": "Updated description",
  "user": "admin",
  "is_remote": false,
  "server_id": null,
  "ssh_key_id": null,
  "created_at": "2025-11-10T12:00:00Z",
  "updated_at": "2025-11-11T11:00:00Z"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid request body
- `404 Not Found`: Saved command not found

**Example**:

```bash
curl -X PUT http://localhost:7777/api/saved-commands/1 \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Updated description"
  }'
```

---

### Delete Saved Command

Delete a saved command template.

**Endpoint**: `DELETE /saved-commands/{id}`

**Path Parameters**:
- `id` (integer, required): Saved command ID

**Response**: `204 No Content`

**Error Responses**:
- `404 Not Found`: Saved command not found

**Example**:

```bash
curl -X DELETE http://localhost:7777/api/saved-commands/1
```

---

## Command History

View execution history for all commands (local and remote).

### List Command History

Retrieve command execution history.

**Endpoint**: `GET /history`

**Query Parameters**:
- `limit` (integer, optional): Maximum number of results to return. Default: 100
- `offset` (integer, optional): Number of results to skip. Default: 0
- `server` (string, optional): Filter by server name (e.g., "local", "production-server")

**Response**: `200 OK`

```json
[
  {
    "id": 1,
    "command": "ls -la /tmp",
    "output": "total 0\ndrwxrwxrwt  10 root  wheel  320 Nov 11 12:00 .\ndrwxr-xr-x  20 root  wheel  640 Nov 11 10:00 ..\n",
    "exit_code": 0,
    "server": "local",
    "user": "root",
    "execution_time_ms": 12,
    "executed_at": "2025-11-11T12:00:00Z"
  },
  {
    "id": 2,
    "command": "uptime",
    "output": " 13:46:21 up 5 days,  3:21,  2 users,  load average: 0.52, 0.58, 0.59",
    "exit_code": 0,
    "server": "production-server",
    "user": "admin",
    "execution_time_ms": 245,
    "executed_at": "2025-11-11T13:46:21Z"
  }
]
```

**Fields**:
- `id` (integer): History entry ID
- `command` (string): Executed command (encrypted in database)
- `output` (string): Command output (encrypted in database)
- `exit_code` (integer): Exit code (0 = success)
- `server` (string): Server name or "local" for local commands
- `user` (string): User who executed the command
- `execution_time_ms` (integer): Execution time in milliseconds
- `executed_at` (string): Timestamp of execution (ISO 8601 format)

**Example**:

```bash
# Get all history
curl http://localhost:7777/api/history

# Get first 10 entries
curl "http://localhost:7777/api/history?limit=10"

# Get local commands only
curl "http://localhost:7777/api/history?server=local"

# Pagination
curl "http://localhost:7777/api/history?limit=20&offset=40"
```

---

### Get Single History Entry

Retrieve a specific history entry by ID.

**Endpoint**: `GET /history/{id}`

**Path Parameters**:
- `id` (integer, required): History entry ID

**Response**: `200 OK`

```json
{
  "id": 1,
  "command": "ls -la /tmp",
  "output": "total 0\ndrwxrwxrwt  10 root  wheel  320 Nov 11 12:00 .\ndrwxr-xr-x  20 root  wheel  640 Nov 11 10:00 ..\n",
  "exit_code": 0,
  "server": "local",
  "user": "root",
  "execution_time_ms": 12,
  "executed_at": "2025-11-11T12:00:00Z"
}
```

**Error Responses**:
- `404 Not Found`: History entry not found

**Example**:

```bash
curl http://localhost:7777/api/history/1
```

---

## Environment Variables Management

Manage encrypted environment variables that can be injected into script executions. All values are encrypted with AES-256-GCM before storage.

### List All Environment Variables

Retrieve all stored environment variables. Values are masked by default for security.

**Endpoint**: `GET /env-variables`

**Query Parameters**:
- `show_values` (boolean, optional): Set to `true` to show actual values. Default: `false`

**Response**: `200 OK`

```json
[
  {
    "id": 1,
    "name": "API_KEY",
    "value": "••••••••",
    "description": "External API key",
    "created_at": "2025-11-10T12:00:00Z",
    "updated_at": "2025-11-10T12:00:00Z"
  },
  {
    "id": 2,
    "name": "DATABASE_URL",
    "value": "••••••••",
    "description": "Database connection string",
    "created_at": "2025-11-10T13:00:00Z",
    "updated_at": "2025-11-10T13:00:00Z"
  }
]
```

**Example**:

```bash
# List with masked values
curl http://localhost:7777/api/env-variables

# List with actual values
curl "http://localhost:7777/api/env-variables?show_values=true"
```

---

### Get Single Environment Variable

Retrieve a specific environment variable by ID.

**Endpoint**: `GET /env-variables/{id}`

**Path Parameters**:
- `id` (integer, required): Environment variable ID

**Query Parameters**:
- `show_value` (boolean, optional): Set to `true` to show actual value. Default: `false`

**Response**: `200 OK`

```json
{
  "id": 1,
  "name": "API_KEY",
  "value": "sk-1234567890abcdef",
  "description": "External API key",
  "created_at": "2025-11-10T12:00:00Z",
  "updated_at": "2025-11-10T12:00:00Z"
}
```

**Error Responses**:
- `404 Not Found`: Environment variable not found

**Example**:

```bash
curl "http://localhost:7777/api/env-variables/1?show_value=true"
```

---

### Create Environment Variable

Add a new encrypted environment variable.

**Endpoint**: `POST /env-variables`

**Request Headers**:
- `Content-Type: application/json`

**Request Body**:

```json
{
  "name": "SECRET_TOKEN",
  "value": "super-secret-value-123",
  "description": "Authentication token for external service"
}
```

**Fields**:
- `name` (string, required): Variable name (e.g., `API_KEY`, `DATABASE_URL`)
- `value` (string, required): Variable value (will be encrypted)
- `description` (string, optional): Description of the variable

**Response**: `201 Created`

```json
{
  "id": 3,
  "name": "SECRET_TOKEN",
  "value": "••••••••",
  "description": "Authentication token for external service",
  "created_at": "2025-11-11T10:00:00Z",
  "updated_at": "2025-11-11T10:00:00Z"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid request body or missing required fields

**Example**:

```bash
curl -X POST http://localhost:7777/api/env-variables \
  -H "Content-Type: application/json" \
  -d '{
    "name": "SECRET_TOKEN",
    "value": "super-secret-value-123",
    "description": "Authentication token"
  }'
```

---

### Update Environment Variable

Update an existing environment variable.

**Endpoint**: `PUT /env-variables/{id}`

**Path Parameters**:
- `id` (integer, required): Environment variable ID

**Request Headers**:
- `Content-Type: application/json`

**Request Body**:

```json
{
  "name": "UPDATED_TOKEN",
  "value": "new-secret-value",
  "description": "Updated description"
}
```

**Fields**: All fields are optional; only provided fields will be updated.

**Response**: `200 OK`

**Error Responses**:
- `400 Bad Request`: Invalid request body
- `404 Not Found`: Environment variable not found

**Example**:

```bash
curl -X PUT http://localhost:7777/api/env-variables/1 \
  -H "Content-Type: application/json" \
  -d '{
    "value": "new-secret-value"
  }'
```

---

### Delete Environment Variable

Delete an environment variable from the system.

**Endpoint**: `DELETE /env-variables/{id}`

**Path Parameters**:
- `id` (integer, required): Environment variable ID

**Response**: `204 No Content`

**Error Responses**:
- `404 Not Found`: Environment variable not found

**Example**:

```bash
curl -X DELETE http://localhost:7777/api/env-variables/1
```

---

## Bash Scripts Management

Manage stored bash scripts that can be executed locally or remotely. Script content is encrypted with AES-256-GCM before storage.

### List All Bash Scripts

Retrieve all stored bash scripts. Script content is not included in list view for performance.

**Endpoint**: `GET /bash-scripts`

**Response**: `200 OK`

```json
[
  {
    "id": 1,
    "name": "deploy-app",
    "description": "Deploy application to production",
    "content": "",
    "filename": "deploy.sh",
    "created_at": "2025-11-10T12:00:00Z",
    "updated_at": "2025-11-10T12:00:00Z"
  },
  {
    "id": 2,
    "name": "backup-database",
    "description": "Backup PostgreSQL database",
    "content": "",
    "filename": "backup.sh",
    "created_at": "2025-11-10T13:00:00Z",
    "updated_at": "2025-11-10T13:00:00Z"
  }
]
```

**Example**:

```bash
curl http://localhost:7777/api/bash-scripts
```

---

### Get Single Bash Script

Retrieve a specific bash script by ID, including full content.

**Endpoint**: `GET /bash-scripts/{id}`

**Path Parameters**:
- `id` (integer, required): Bash script ID

**Response**: `200 OK`

```json
{
  "id": 1,
  "name": "deploy-app",
  "description": "Deploy application to production",
  "content": "#!/bin/bash\nset -e\n\necho \"Deploying...\"\ncd /opt/app\ngit pull origin main\nsystemctl restart app\necho \"Done!\"",
  "filename": "deploy.sh",
  "created_at": "2025-11-10T12:00:00Z",
  "updated_at": "2025-11-10T12:00:00Z"
}
```

**Error Responses**:
- `404 Not Found`: Bash script not found

**Example**:

```bash
curl http://localhost:7777/api/bash-scripts/1
```

---

### Create Bash Script

Add a new bash script to the system.

**Endpoint**: `POST /bash-scripts`

**Request Headers**:
- `Content-Type: application/json`

**Request Body**:

```json
{
  "name": "system-check",
  "description": "Check system health",
  "content": "#!/bin/bash\nset -e\necho \"CPU Usage:\"\ntop -bn1 | head -5\necho \"\\nDisk Usage:\"\ndf -h\necho \"\\nMemory:\"\nfree -h",
  "filename": "system-check.sh"
}
```

**Fields**:
- `name` (string, required): Display name for the script
- `content` (string, required): Bash script content
- `description` (string, optional): Description of what the script does
- `filename` (string, optional): Original filename if uploaded

**Response**: `201 Created`

```json
{
  "id": 3,
  "name": "system-check",
  "description": "Check system health",
  "content": "#!/bin/bash\nset -e\n...",
  "filename": "system-check.sh",
  "created_at": "2025-11-11T10:00:00Z",
  "updated_at": "2025-11-11T10:00:00Z"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid request body or missing required fields

**Example**:

```bash
curl -X POST http://localhost:7777/api/bash-scripts \
  -H "Content-Type: application/json" \
  -d '{
    "name": "system-check",
    "description": "Check system health",
    "content": "#!/bin/bash\necho \"Hello World\""
  }'
```

---

### Update Bash Script

Update an existing bash script.

**Endpoint**: `PUT /bash-scripts/{id}`

**Path Parameters**:
- `id` (integer, required): Bash script ID

**Request Headers**:
- `Content-Type: application/json`

**Request Body**:

```json
{
  "name": "updated-script-name",
  "description": "Updated description",
  "content": "#!/bin/bash\necho \"Updated script\""
}
```

**Fields**: All fields are optional; only provided fields will be updated.

**Response**: `200 OK`

**Error Responses**:
- `400 Bad Request`: Invalid request body
- `404 Not Found`: Bash script not found

**Example**:

```bash
curl -X PUT http://localhost:7777/api/bash-scripts/1 \
  -H "Content-Type: application/json" \
  -d '{
    "description": "Updated description"
  }'
```

---

### Delete Bash Script

Delete a bash script from the system.

**Endpoint**: `DELETE /bash-scripts/{id}`

**Path Parameters**:
- `id` (integer, required): Bash script ID

**Response**: `204 No Content`

**Error Responses**:
- `404 Not Found`: Bash script not found

**Example**:

```bash
curl -X DELETE http://localhost:7777/api/bash-scripts/1
```

---

### Execute Bash Script

Execute a stored bash script locally or on a remote server, optionally injecting environment variables.

**Endpoint**: `POST /bash-scripts/execute`

**Request Headers**:
- `Content-Type: application/json`

**Request Body (Local Execution)**:

```json
{
  "script_id": 1,
  "user": "root",
  "sudo_password": "your-sudo-password",
  "env_var_ids": [1, 2, 3]
}
```

**Request Body (Remote Execution)**:

```json
{
  "script_id": 1,
  "user": "admin",
  "is_remote": true,
  "server_id": 1,
  "ssh_key_id": 2,
  "ssh_password": "fallback-password",
  "env_var_ids": [1, 2]
}
```

**Fields**:
- `script_id` (integer, required): ID of the script to execute
- `user` (string, optional): User to run as. Default: `"root"`
- `sudo_password` (string, optional): Sudo password for local root execution
- `ssh_password` (string, optional): SSH password fallback for remote execution
- `is_remote` (boolean, optional): Set to `true` for remote execution. Default: `false`
- `server_id` (integer, optional): Server ID for remote execution
- `ssh_key_id` (integer, optional): SSH key ID for remote authentication
- `env_var_ids` (array of integers, optional): Environment variable IDs to inject

**Response**: `200 OK`

```json
{
  "script_id": 1,
  "script_name": "deploy-app",
  "output": "Deploying...\nDone!",
  "exit_code": 0,
  "user": "root",
  "server": "local",
  "execution_time_ms": 1234,
  "env_vars_injected": 3
}
```

**Fields**:
- `script_id` (integer): ID of executed script
- `script_name` (string): Name of executed script
- `output` (string): Combined stdout and stderr output
- `exit_code` (integer): Exit code (0 = success)
- `user` (string): User who executed the script
- `server` (string): "local" or server name for remote execution
- `execution_time_ms` (integer): Execution time in milliseconds
- `env_vars_injected` (integer): Number of environment variables injected

**Error Responses**:
- `400 Bad Request`: Invalid request body or missing required fields
- `404 Not Found`: Script, server, or SSH key not found
- `500 Internal Server Error`: Script execution failed

**Example (Local with Env Vars)**:

```bash
curl -X POST http://localhost:7777/api/bash-scripts/execute \
  -H "Content-Type: application/json" \
  -d '{
    "script_id": 1,
    "user": "root",
    "env_var_ids": [1, 2]
  }'
```

**Example (Remote Execution)**:

```bash
curl -X POST http://localhost:7777/api/bash-scripts/execute \
  -H "Content-Type: application/json" \
  -d '{
    "script_id": 1,
    "is_remote": true,
    "server_id": 1,
    "ssh_key_id": 2,
    "env_var_ids": [1]
  }'
```

---

### Get Script Presets by Script

Retrieve all presets associated with a specific script.

**Endpoint**: `GET /bash-scripts/{id}/presets`

**Path Parameters**:
- `id` (integer, required): Bash script ID

**Response**: `200 OK`

```json
[
  {
    "id": 1,
    "name": "Production Deploy",
    "description": "Deploy to production server",
    "script_id": 1,
    "env_var_ids": [1, 2],
    "is_remote": true,
    "server_id": 1,
    "ssh_key_id": 2,
    "user": "deploy",
    "created_at": "2025-11-10T12:00:00Z",
    "updated_at": "2025-11-10T12:00:00Z"
  }
]
```

**Error Responses**:
- `404 Not Found`: Script not found

**Example**:

```bash
curl http://localhost:7777/api/bash-scripts/1/presets
```

---

## Script Presets Management

Manage saved script execution configurations. Presets store which environment variables to inject and optionally remote execution settings.

### List All Script Presets

Retrieve all stored script presets.

**Endpoint**: `GET /script-presets`

**Response**: `200 OK`

```json
[
  {
    "id": 1,
    "name": "Production Deploy",
    "description": "Deploy to production server",
    "script_id": 1,
    "env_var_ids": [1, 2],
    "is_remote": true,
    "server_id": 1,
    "ssh_key_id": 2,
    "user": "deploy",
    "created_at": "2025-11-10T12:00:00Z",
    "updated_at": "2025-11-10T12:00:00Z"
  },
  {
    "id": 2,
    "name": "Local Test",
    "description": "Run script locally for testing",
    "script_id": 1,
    "env_var_ids": [3],
    "is_remote": false,
    "server_id": null,
    "ssh_key_id": null,
    "user": "current",
    "created_at": "2025-11-10T13:00:00Z",
    "updated_at": "2025-11-10T13:00:00Z"
  }
]
```

**Example**:

```bash
curl http://localhost:7777/api/script-presets
```

---

### Get Single Script Preset

Retrieve a specific script preset by ID.

**Endpoint**: `GET /script-presets/{id}`

**Path Parameters**:
- `id` (integer, required): Script preset ID

**Response**: `200 OK`

```json
{
  "id": 1,
  "name": "Production Deploy",
  "description": "Deploy to production server",
  "script_id": 1,
  "env_var_ids": [1, 2],
  "is_remote": true,
  "server_id": 1,
  "ssh_key_id": 2,
  "user": "deploy",
  "created_at": "2025-11-10T12:00:00Z",
  "updated_at": "2025-11-10T12:00:00Z"
}
```

**Error Responses**:
- `404 Not Found`: Script preset not found

**Example**:

```bash
curl http://localhost:7777/api/script-presets/1
```

---

### Create Script Preset

Create a new script execution preset.

**Endpoint**: `POST /script-presets`

**Request Headers**:
- `Content-Type: application/json`

**Request Body (Local Preset)**:

```json
{
  "name": "Local Dev Test",
  "description": "Run locally with dev environment",
  "script_id": 1,
  "env_var_ids": [1, 3],
  "is_remote": false,
  "user": "current"
}
```

**Request Body (Remote Preset)**:

```json
{
  "name": "Staging Deploy",
  "description": "Deploy to staging server",
  "script_id": 1,
  "env_var_ids": [1, 2],
  "is_remote": true,
  "server_id": 2,
  "ssh_key_id": 1,
  "user": "deploy"
}
```

**Fields**:
- `name` (string, required): Display name for the preset
- `script_id` (integer, required): ID of the associated bash script
- `description` (string, optional): Description of the preset
- `env_var_ids` (array of integers, optional): Environment variable IDs to inject
- `is_remote` (boolean, optional): Whether this is for remote execution. Default: `false`
- `server_id` (integer, optional): Server ID for remote execution
- `ssh_key_id` (integer, optional): SSH key ID for remote authentication
- `user` (string, optional): User to run as. Default: `"root"`

**Response**: `201 Created`

```json
{
  "id": 3,
  "name": "Staging Deploy",
  "description": "Deploy to staging server",
  "script_id": 1,
  "env_var_ids": [1, 2],
  "is_remote": true,
  "server_id": 2,
  "ssh_key_id": 1,
  "user": "deploy",
  "created_at": "2025-11-11T10:00:00Z",
  "updated_at": "2025-11-11T10:00:00Z"
}
```

**Error Responses**:
- `400 Bad Request`: Invalid request body or missing required fields

**Example**:

```bash
curl -X POST http://localhost:7777/api/script-presets \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Staging Deploy",
    "script_id": 1,
    "env_var_ids": [1, 2],
    "is_remote": true,
    "server_id": 2,
    "ssh_key_id": 1
  }'
```

---

### Update Script Preset

Update an existing script preset.

**Endpoint**: `PUT /script-presets/{id}`

**Path Parameters**:
- `id` (integer, required): Script preset ID

**Request Headers**:
- `Content-Type: application/json`

**Request Body**:

```json
{
  "name": "Updated Preset Name",
  "env_var_ids": [1, 2, 3],
  "user": "admin"
}
```

**Fields**: All fields are optional; only provided fields will be updated.

**Response**: `200 OK`

**Error Responses**:
- `400 Bad Request`: Invalid request body
- `404 Not Found`: Script preset not found

**Example**:

```bash
curl -X PUT http://localhost:7777/api/script-presets/1 \
  -H "Content-Type: application/json" \
  -d '{
    "env_var_ids": [1, 2, 3]
  }'
```

---

### Delete Script Preset

Delete a script preset from the system.

**Endpoint**: `DELETE /script-presets/{id}`

**Path Parameters**:
- `id` (integer, required): Script preset ID

**Response**: `204 No Content`

**Error Responses**:
- `404 Not Found`: Script preset not found

**Example**:

```bash
curl -X DELETE http://localhost:7777/api/script-presets/1
```

---

## Error Responses

All API endpoints use standard HTTP status codes and return JSON error responses.

### Standard Error Format

```json
{
  "error": "Error message describing what went wrong"
}
```

### Common Status Codes

| Status Code | Meaning | Description |
|------------|---------|-------------|
| `200 OK` | Success | Request succeeded |
| `201 Created` | Created | Resource created successfully |
| `204 No Content` | No Content | Request succeeded, no response body |
| `400 Bad Request` | Client Error | Invalid request body or parameters |
| `404 Not Found` | Not Found | Resource does not exist |
| `500 Internal Server Error` | Server Error | Server encountered an error |

### Example Error Responses

**400 Bad Request**:

```json
{
  "error": "Command is required"
}
```

**404 Not Found**:

```json
{
  "error": "Server not found"
}
```

**500 Internal Server Error**:

```json
{
  "error": "Failed to execute command"
}
```

---

## Security Considerations

### Encryption

- All SSH private keys are encrypted with **AES-256-GCM** before storage
- Environment variable values are encrypted at rest
- Bash script content is encrypted at rest
- Command output is encrypted in the database
- Encryption key is stored in `.encryption_key` file (backup this file!)
- System entropy is verified before generating encryption keys (Linux)

### Password Handling

- **SSH passwords are NEVER stored in command history** (security feature)
- Sudo passwords are only used for command execution and discarded immediately
- Passwords hashed with **bcrypt** (cost factor 12) when stored
- Constant-time comparison prevents timing attacks

### Authentication

- HTTP Basic Authentication supported
- Bearer token (API token) authentication supported
- **Production deployment requires enabling authentication** (`AUTH_ENABLED=true`)
- Supports both methods simultaneously (token takes precedence)

### TLS/HTTPS

- Native TLS support with `-tls-cert` and `-tls-key` flags
- Optional HTTPS enforcement with `-require-https`
- Security headers included (X-Frame-Options, X-Content-Type-Options, etc.)

### SSH Security

- **Host key verification** against `~/.ssh/known_hosts`
- Trust-on-first-use (TOFU) mode for development
- Man-in-the-middle attack detection

### Input Validation

- IP addresses validated (IPv4/IPv6)
- Hostnames validated (RFC 1123 compliant)
- Port numbers validated (1-65535)
- SSH private keys format validated
- Unix usernames validated
- Command names checked for malicious content

### Best Practices

1. **Enable authentication** in production (`AUTH_ENABLED=true`)
2. **Use HTTPS** - either native TLS or reverse proxy
3. **Backup encryption key** - data cannot be recovered without it
4. **Use SSH key authentication** over passwords when possible
5. **Configure CORS** - set `CORS_ALLOWED_ORIGINS` for production
6. **Monitor command history** for suspicious activity
7. **Limit API access** to trusted networks
8. **Set up proper HTTP timeouts** - configured automatically for DoS protection

---

## Rate Limiting

Currently, there is no rate limiting implemented. Consider implementing rate limiting in production to prevent abuse.

---

## Versioning

Current API version: **v1**

The API does not currently use versioning in the URL path. Future versions may use `/api/v2/` prefix.

---

## Support

For API issues or questions:
- GitHub Issues: https://github.com/pozgo/web-cli/issues
- Documentation: README.md and CLAUDE.md

---

## Changelog

### Version 1.1.0 (Security & Features Update)
- **Security Enhancements:**
  - Added bcrypt password hashing (cost factor 12)
  - Added SSH host key verification with TOFU support
  - Added comprehensive input validation
  - Added native TLS/HTTPS support
  - Added HTTPS enforcement option
  - Added security headers middleware
  - Added system entropy verification for key generation
  - Added detailed decryption audit logging
- **Configuration:**
  - Added Viper configuration management
  - Support for config files (YAML, JSON, TOML)
  - Support for WEBCLI_ prefixed environment variables
  - Multiple config file search paths
- **New Features:**
  - Added Environment Variables Management (5 endpoints)
  - Added Bash Scripts Management (7 endpoints)
  - Added Script Presets Management (5 endpoints)
  - Total API endpoints: 41
- **Authentication:**
  - HTTP Basic Auth support
  - Bearer token support
  - Configurable via environment variables

### Version 1.0.0 (Stage 6.1)
- Fixed saved commands filter in LocalCommands component
- Local commands now properly filtered to exclude remote commands
- Remote commands only appear in RemoteCommands dropdown
- Complete isolation between local and remote saved command templates

### Version 1.0.0 (Stage 6)
- Added remote command execution support
- Added `is_remote`, `server_id`, `ssh_key_id` fields to saved commands
- Added SSH password fallback authentication
- Added system information endpoint
- Enhanced command execution with remote capability

### Version 1.0.0 (Stage 5.2.1)
- Added current user endpoint
- Enhanced user selection with local users

### Version 1.0.0 (Stage 5)
- Initial local command execution
- Command history tracking
- Saved commands management

### Version 1.0.0 (Stage 4)
- SSH key management
- Server management
- Local user management
- Admin panel foundation

---

**Last Updated**: December 7, 2025
**API Version**: 1.1.0
