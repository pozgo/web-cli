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

## Authentication

Currently, the API does not require authentication. All endpoints are publicly accessible.

**Note**: In production environments, implement proper authentication and authorization mechanisms.

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
- Command output is encrypted in the database
- Encryption key is stored in `.encryption_key` file (backup this file!)

### Password Handling

- **SSH passwords are NEVER stored in command history** (security feature)
- Sudo passwords are only used for command execution and discarded immediately
- Passwords are transmitted over HTTP - use HTTPS in production

### Authentication

- Current version has no authentication
- **Production deployment requires authentication** (implement JWT, OAuth, etc.)
- Consider IP whitelisting or VPN access

### Best Practices

1. **Use HTTPS** in production environments
2. **Implement authentication** before deploying to production
3. **Backup encryption key** - data cannot be recovered without it
4. **Use SSH key authentication** over passwords when possible
5. **Validate all inputs** on the client side before sending to API
6. **Monitor command history** for suspicious activity
7. **Limit API access** to trusted networks

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

**Last Updated**: November 11, 2025
**API Version**: 1.0.0
