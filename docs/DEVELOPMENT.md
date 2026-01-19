# Development Guide

This document covers the development setup and workflow for Web CLI.

## Table of Contents

- [Prerequisites](#prerequisites)
- [Quick Setup](#quick-setup)
- [Development Workflow](#development-workflow)
- [Project Structure](#project-structure)
- [Backend Development](#backend-development)
- [Frontend Development](#frontend-development)
- [Testing](#testing)
- [Building](#building)
- [Contributing](#contributing)

---

## Prerequisites

- **Go**: Version 1.21 or higher ([Download](https://golang.org/dl/))
- **Node.js**: Version 18 or higher ([Download](https://nodejs.org/))
- **npm**: Version 8 or higher (comes with Node.js)

---

## Quick Setup

```bash
# Clone the repository
git clone https://github.com/pozgo/web-cli.git
cd web-cli

# Install Go dependencies
go mod download

# Install frontend dependencies
cd frontend
npm install
cd ..

# Build and run
./build.sh
./web-cli
```

Access at: `http://localhost:7777`

---

## Development Workflow

### Recommended: Two Terminal Setup

**Terminal 1 - Backend:**

```bash
go run cmd/web-cli/main.go
```

**Terminal 2 - Frontend (with hot reload):**

```bash
cd frontend && npm run dev
```

Access frontend dev server at: `http://localhost:3000`
(API calls are proxied to the Go backend on port 7777)

### Using Management Script

```bash
# Start the server
./manage.sh start

# Stop the server
./manage.sh stop

# Check status
./manage.sh status

# Restart
./manage.sh restart
```

---

## Project Structure

```
web-cli/
├── cmd/web-cli/           # Application entry point
│   └── main.go            # Main function
├── internal/              # Private application code
│   ├── audit/             # Audit logging
│   ├── config/            # Configuration management
│   ├── database/          # Database, migrations, encryption
│   ├── executor/          # Command execution (local & remote)
│   │   └── hostkeys.go    # SSH host key verification
│   ├── middleware/        # HTTP middleware (auth, security)
│   ├── models/            # Data models
│   ├── repository/        # Data access layer
│   ├── server/            # HTTP server and handlers
│   ├── terminal/          # Interactive terminal (PTY + WebSocket)
│   ├── validation/        # Input validation
│   └── vault/             # HashiCorp Vault client
├── frontend/              # React application
│   ├── src/
│   │   ├── components/    # React components
│   │   │   ├── Terminal.jsx       # xterm.js terminal
│   │   │   ├── AdminPanel.jsx     # Admin interface
│   │   │   ├── Dashboard.jsx      # Main dashboard
│   │   │   └── ...
│   │   ├── theme/         # MUI theme configuration
│   │   └── App.jsx        # Main app with routing
│   ├── public/            # Static assets
│   ├── package.json       # npm dependencies
│   └── vite.config.js     # Vite configuration
├── docs/                  # Documentation
│   ├── images/            # Screenshots
│   ├── CONFIGURATION.md   # Configuration guide
│   ├── SECURITY.md        # Security guide
│   ├── DEPLOYMENT.md      # Deployment guide
│   └── DEVELOPMENT.md     # This file
├── assets/                # Embedded frontend (production)
├── build.sh               # Build script (all platforms)
├── manage.sh              # Server management script
├── API.md                 # Complete API documentation
├── Dockerfile             # Docker build file
├── docker-compose.yml     # Docker Compose configuration
└── go.mod                 # Go dependencies
```

---

## Backend Development

### Running

```bash
# Run with go run
go run cmd/web-cli/main.go

# Run with custom options
go run cmd/web-cli/main.go -port 8080 -host localhost

# Run with air (hot reload)
air
```

### Testing

```bash
# Run all tests
go test ./...

# Run tests with coverage
go test -cover ./...

# Run tests for a specific package
go test ./internal/validation/...

# Verbose output
go test -v ./...
```

### Linting

```bash
# Format code
go fmt ./...

# Run vet
go vet ./...

# Run golangci-lint (if installed)
golangci-lint run
```

### Generate Swagger Docs

```bash
# Install swag CLI
go install github.com/swaggo/swag/cmd/swag@latest

# Generate documentation
swag init -g cmd/web-cli/main.go -o docs/
```

---

## Frontend Development

### Running

```bash
cd frontend

# Start dev server with hot reload
npm run dev

# Build for production
npm run build

# Preview production build
npm run preview
```

### Testing

```bash
cd frontend

# Run tests
npm test

# Run tests with coverage
npm test -- --coverage
```

### Linting

```bash
cd frontend

# Run ESLint
npm run lint

# Fix auto-fixable issues
npm run lint -- --fix
```

### Dependencies

Key frontend dependencies:

| Package | Purpose |
|---------|---------|
| `react` | UI library |
| `react-router-dom` | Client-side routing |
| `@mui/material` | Material-UI components |
| `@emotion/react` | CSS-in-JS styling |
| `xterm` | Terminal emulator |
| `@monaco-editor/react` | Code editor for YAML/JSON tools |
| `js-yaml` | YAML parsing |

---

## Testing

### Backend Tests

```bash
# All tests
go test ./...

# With coverage report
go test -coverprofile=coverage.out ./...
go tool cover -html=coverage.out

# Specific package
go test ./internal/validation/

# With race detection
go test -race ./...
```

### Frontend Tests

```bash
cd frontend
npm test
```

### Integration Testing

```bash
# Build and run
./build.sh
./web-cli &

# Test API endpoints
curl http://localhost:7777/api/health

# Test with authentication
curl -u admin:password http://localhost:7777/api/health
```

---

## Building

### Development Build

```bash
# Quick build for current platform
go build -o web-cli cmd/web-cli/main.go
```

### Production Build (All Platforms)

```bash
# Build for all platforms
./build.sh all

# Outputs:
# bin/web-cli-linux-x64
# bin/web-cli-darwin-x64
# bin/web-cli-darwin-arm64
```

### Docker Build

```bash
# Build Docker image
docker build -t web-cli .

# Or using docker compose
docker compose build
```

### Build Script Options

```bash
./build.sh          # Build for current platform
./build.sh all      # Build for all platforms
./build.sh linux    # Build for Linux only
./build.sh darwin   # Build for macOS only
```

---

## Contributing

### Getting Started

1. Fork the repository
2. Create a feature branch: `git checkout -b feature/amazing-feature`
3. Make your changes
4. Write/update tests
5. Commit: `git commit -m 'Add some amazing feature'`
6. Push: `git push origin feature/amazing-feature`
7. Open a Pull Request

### Code Style

- **Go**: Follow standard Go formatting (`go fmt`)
- **JavaScript/React**: Follow ESLint rules
- Write tests for new features
- Update documentation as needed

### Commit Messages

Use clear, descriptive commit messages:

```
feat: Add new feature X
fix: Fix bug in component Y
docs: Update README
refactor: Restructure module Z
test: Add tests for feature X
```

### Pull Request Guidelines

- Ensure all tests pass
- Update documentation if needed
- Keep PRs focused (one feature/fix per PR)
- Provide clear description of changes

---

For configuration options, see [CONFIGURATION.md](CONFIGURATION.md).
For security information, see [SECURITY.md](SECURITY.md).
For deployment instructions, see [DEPLOYMENT.md](DEPLOYMENT.md).
