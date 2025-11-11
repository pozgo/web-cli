# Web CLI

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://golang.org)
[![React](https://img.shields.io/badge/React-18-61DAFB?style=flat&logo=react)](https://react.dev)
[![Material-UI](https://img.shields.io/badge/Material--UI-5-007FFF?style=flat&logo=mui)](https://mui.com)
[![License](https://img.shields.io/badge/License-MIT-green.svg)](LICENSE)

A powerful web-based interface for executing commands on local and remote Linux servers. Built with Go and React with Material-UI for a professional, modern user experience.

![Dashboard](docs/images/dashboard.png)
*Main dashboard with feature cards for easy navigation*

## 📋 Table of Contents

- [Features](#-features)
- [Tech Stack](#-tech-stack)
- [Screenshots](#-screenshots)
- [Prerequisites](#-prerequisites)
- [Installation](#-installation)
- [Quick Start](#-quick-start)
- [Usage](#-usage)
- [API Documentation](#-api-documentation)
- [Configuration](#-configuration)
- [Security](#-security)
- [Deployment](#-deployment)
- [Development](#-development)
- [Contributing](#-contributing)
- [License](#-license)

## ✨ Features

### Core Functionality
- **Web-Based Interface**: Professional Material-UI dashboard accessible from any browser
- **Theme Support**: Dark and light mode toggle with preferences saved locally
- **Multi-Platform**: Compiled binaries for Linux x64, macOS Intel, and macOS Apple Silicon

### Command Execution
- **Local Commands**: Execute bash commands on your local server with real-time output
- **Remote Commands**: Connect and execute commands on remote servers via SSH
- **User Selection**: Run commands as different users (current, root, or custom)
- **Sudo Support**: Secure password dialog for root command execution

### Server & SSH Management
- **Admin Panel**: Full management interface with tabbed sections
- **SSH Key Management**: CRUD operations for SSH private keys
- **Server Management**: Manage remote servers with hostname, IP, port, and username
- **Hostname Validation**: Automatic validation according to hostname conventions

### Command Templates & History
- **Command Templates**: Save frequently-used commands for quick re-execution
- **Command Type Management**: Visual indicators (Local/Remote) with type switching
- **Smart Navigation**: Execute button routes to appropriate page based on command type
- **Command History**: Complete execution history with output, exit codes, and timing

### Security
- **Encrypted Database**: SQLite with AES-256-GCM encryption for sensitive data
- **Secure Password Handling**: SSH passwords never stored in command history
- **Automatic Encryption**: All SSH keys and command history encrypted at rest
- **Database Migrations**: Automatic schema versioning and migration system

## 🛠 Tech Stack

### Backend
- **[Go 1.21+](https://golang.org/)**: High-performance backend server
- **[Gorilla Mux](https://github.com/gorilla/mux)**: HTTP router for API endpoints
- **[SQLite](https://www.sqlite.org/)**: Embedded database with migration support
- **[golang.org/x/crypto/ssh](https://pkg.go.dev/golang.org/x/crypto/ssh)**: SSH client implementation
- **AES-256-GCM**: Military-grade encryption for sensitive data

### Frontend
- **[React 18](https://react.dev/)**: Modern UI library
- **[React Router v6](https://reactrouter.com/)**: Client-side routing
- **[Material-UI (MUI) v5](https://mui.com/)**: Professional component library
- **[Vite](https://vitejs.dev/)**: Fast build tool and dev server
- **[Emotion](https://emotion.sh/)**: CSS-in-JS styling

## 📸 Screenshots

### Dashboard
![Dashboard](docs/images/dashboard.png)
*Main dashboard with feature cards for easy navigation*

### Local Command Execution
![Local Commands](docs/images/local-commands.png)
*Execute commands on your local server with real-time output*

### Remote Command Execution
![Remote Commands](docs/images/remote-commands.png)
*Connect to remote servers via SSH and execute commands*

### Admin Panel - SSH Keys
![SSH Keys Management](docs/images/admin-ssh-keys.png)
*Manage SSH private keys for server authentication*

### Admin Panel - Servers
![Server Management](docs/images/admin-servers.png)
*Configure and manage remote servers*

### Saved Commands
![Saved Commands](docs/images/saved-commands.png)
*Save and reuse frequently-used command templates with type indicators*

### Command History
![Command History](docs/images/command-history.png)
*View complete execution history with output and timing*

## 📦 Prerequisites

- **Go**: Version 1.21 or higher ([Download](https://golang.org/dl/))
- **Node.js**: Version 18 or higher ([Download](https://nodejs.org/))
- **npm**: Version 8 or higher (comes with Node.js)

## 🚀 Installation

### Clone the Repository

```bash
git clone https://github.com/pozgo/web-cli.git
cd web-cli
```

### Install Dependencies

#### Backend
```bash
go mod download
```

#### Frontend
```bash
cd frontend
npm install
cd ..
```

## ⚡ Quick Start

### Build and Run

```bash
# Build the application
./build.sh

# Run the server
./web-cli
```

Access the application at `http://localhost:7777`

### Using Management Script

```bash
# Start the server
./manage.sh start

# Stop the server
./manage.sh stop

# Check status
./manage.sh status
```

## 📖 Usage

### Local Command Execution

1. Navigate to **Local Commands** from the dashboard
2. Enter your bash command
3. Select the user to run as (current, root, or custom)
4. Click **Execute Command**
5. View real-time output

### Remote Command Execution

1. First, add SSH keys and servers in the **Admin Panel**
2. Navigate to **Remote Execution**
3. Select a server from the dropdown
4. Choose an SSH key (optional, password fallback available)
5. Enter your command and click **Execute**

### Saved Command Templates

1. Execute any command and check "Save command as template"
2. Access saved commands from the **Saved Commands** page
3. Edit command type (Local/Remote) and details
4. Execute saved commands with one click

## 📚 API Documentation

Web CLI provides a comprehensive RESTful API for programmatic access to all features. Perfect for automation, CI/CD pipelines, and integration with other tools.

### 🚀 Quick Start

**Base URL**: `http://localhost:7777/api`

**Example Request**:
```bash
curl http://localhost:7777/api/health
# Response: {"status":"ok"}
```

### 📋 API Endpoints Overview

| Category | Endpoints | Description |
|----------|-----------|-------------|
| **Health** | 1 endpoint | Server health check |
| **SSH Keys** | 5 endpoints | Manage SSH private keys |
| **Servers** | 5 endpoints | Manage remote servers |
| **Local Users** | 5 endpoints | Manage local user accounts |
| **System Info** | 1 endpoint | Get current system user |
| **Commands** | 1 endpoint | Execute local/remote commands |
| **Saved Commands** | 5 endpoints | Manage command templates |
| **History** | 2 endpoints | View execution history |

**Total: 25 RESTful API endpoints**

### 📖 Complete Documentation

For detailed API documentation including:
- ✅ All 25 endpoints with full descriptions
- ✅ Request/response examples with cURL commands
- ✅ Field descriptions and validation rules
- ✅ Error response formats and status codes
- ✅ Security considerations and best practices
- ✅ Authentication guidance for production
- ✅ Quick reference table for all endpoints

**👉 See [API.md](API.md) for complete API reference**

### 💡 Common Use Cases

**Execute a local command**:
```bash
curl -X POST http://localhost:7777/api/commands/execute \
  -H "Content-Type: application/json" \
  -d '{"command": "df -h", "user": "root"}'
```

**Execute a remote command via SSH**:
```bash
curl -X POST http://localhost:7777/api/commands/execute \
  -H "Content-Type: application/json" \
  -d '{
    "command": "uptime",
    "is_remote": true,
    "server_id": 1,
    "ssh_key_id": 2
  }'
```

**List command history**:
```bash
curl "http://localhost:7777/api/history?limit=10&server=local"
```

## ⚙️ Configuration

### Command-Line Flags

```bash
./web-cli [options]

Options:
  -port int             Port to listen on (default: 7777)
  -host string          Host to bind to (default: 0.0.0.0)
  -frontend string      Path to frontend build files (default: ./frontend/build)
  -db string            Path to database file (default: ./data/web-cli.db)
  -encryption-key string Path to encryption key file (default: ./.encryption_key)
```

### Environment Variables

```bash
PORT=8080 HOST=localhost ./web-cli
```

### Configuration Priority

1. Command-line flags (highest priority)
2. Environment variables
3. Default values (lowest priority)

## 🔒 Security

### Database Encryption

All sensitive data is encrypted using **AES-256-GCM**:
- SSH private keys encrypted at rest
- Command history (commands and output) encrypted
- Encryption key auto-generated on first run
- Key stored in `.encryption_key` with 600 permissions

### Encryption Key Management

**Important:**
- Backup your `.encryption_key` file - data cannot be recovered without it
- For production, use environment variable instead of file

**Generate a new encryption key:**

```bash
# macOS/Linux
openssl rand -base64 32

# Or using dd and base64
dd if=/dev/urandom bs=32 count=1 2>/dev/null | base64

# Output example: 7xK9mP2vQ8nL4wR6tY5uE3sA1zD0cF8bG7hJ9kM6nP4=
```

**Use the generated key:**

```bash
# Option 1: Environment variable (recommended for production)
export ENCRYPTION_KEY="7xK9mP2vQ8nL4wR6tY5uE3sA1zD0cF8bG7hJ9kM6nP4="
./web-cli

# Option 2: Save to file (auto-generated on first run if not exists)
echo "7xK9mP2vQ8nL4wR6tY5uE3sA1zD0cF8bG7hJ9kM6nP4=" > .encryption_key
chmod 600 .encryption_key
./web-cli
```

### Password Security

- Sudo passwords only used for command execution
- SSH passwords used for authentication fallback
- **Passwords are never stored** in command history
- Passwords cleared from memory after use

### Database Migrations

- Automatic schema versioning
- Migrations run on startup
- Safe to restart - migrations only run once
- Current schema version: **11**

## 🌐 Deployment

### Linux Server

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

### systemd Service

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
Environment=ENCRYPTION_KEY=your-key-here

[Install]
WantedBy=multi-user.target
```

Enable and start:

```bash
sudo systemctl enable web-cli
sudo systemctl start web-cli
sudo systemctl status web-cli
```

### Docker (Optional)

> **Note**: Docker support can be added. Create an issue if you need it.

## 💻 Development

### Backend Development

```bash
# Run with hot reload (use air or similar)
go run cmd/web-cli/main.go

# Run tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Format code
go fmt ./...

# Lint code
go vet ./...
```

### Frontend Development

```bash
cd frontend

# Start dev server with hot reload
npm run dev

# Run tests
npm test

# Lint code
npm run lint

# Build for production
npm run build
```

### Development Workflow

1. Start backend:
   ```bash
   go run cmd/web-cli/main.go
   ```

2. Start frontend (new terminal):
   ```bash
   cd frontend && npm run dev
   ```

3. Open `http://localhost:3000` for hot-reload development

### Project Structure

```
web-cli/
├── cmd/web-cli/           # Application entry point
├── internal/              # Private application code
│   ├── config/            # Configuration management
│   ├── database/          # Database, migrations, encryption
│   ├── executor/          # Command execution (local & remote)
│   ├── models/            # Data models
│   ├── repository/        # Data access layer
│   └── server/            # HTTP server and handlers
├── frontend/              # React application
│   ├── src/
│   │   ├── components/    # React components
│   │   ├── theme/         # MUI theme configuration
│   │   └── App.jsx        # Main app with routing
│   └── vite.config.js     # Vite configuration
├── build.sh               # Build script (all platforms)
├── manage.sh              # Server management script
└── go.mod                 # Go dependencies
```

## 🤝 Contributing

Contributions are welcome! Please follow these steps:

1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

### Development Guidelines

- Follow Go best practices and conventions
- Write tests for new features
- Update documentation as needed
- Ensure all tests pass before submitting PR
- Keep commits atomic and well-described

## 📝 License

This project is licensed under the MIT License - see the [LICENSE](LICENSE) file for details.

## 🆘 Support

For issues and questions:
- 📫 Open an issue on [GitHub Issues](https://github.com/pozgo/web-cli/issues)
- 📖 Check the [API Documentation](API.md)
- 🤖 See [CLAUDE.md](CLAUDE.md) for AI development guidance

## 🙏 Acknowledgments

- Built with [Go](https://golang.org/) - Fast, reliable, and efficient
- UI powered by [React](https://react.dev/) and [Material-UI](https://mui.com/)
- Build tool: [Vite](https://vitejs.dev/) - Next generation frontend tooling
- SSH implementation: [golang.org/x/crypto/ssh](https://pkg.go.dev/golang.org/x/crypto/ssh)

**Made with ❤️ by [Pozgo](https://github.com/pozgo)**
