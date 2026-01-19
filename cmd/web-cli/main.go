package main

import (
	"log"

	"github.com/pozgo/web-cli/assets"
	"github.com/pozgo/web-cli/internal/audit"
	"github.com/pozgo/web-cli/internal/config"
	"github.com/pozgo/web-cli/internal/database"
	"github.com/pozgo/web-cli/internal/server"

	_ "github.com/pozgo/web-cli/docs" // Swagger docs
)

// @title Web CLI API
// @version 1.1.0
// @description Web-based CLI tool for executing shell commands locally and remotely with SSH key management, script storage, and interactive terminal sessions.
// @termsOfService http://swagger.io/terms/

// @contact.name Web CLI Support
// @contact.url https://github.com/pozgo/web-cli
// @contact.email linux@ozgo.info

// @license.name MIT
// @license.url https://opensource.org/licenses/MIT

// @host localhost:7777
// @BasePath /api

// @securityDefinitions.basic BasicAuth
// @description Basic authentication using username and password

// @securityDefinitions.apikey BearerAuth
// @in header
// @name Authorization
// @description Bearer token authentication (format: "Bearer {token}")

// @tag.name SSH Keys
// @tag.description SSH private key management for remote connections

// @tag.name Servers
// @tag.description Remote server configuration management

// @tag.name Commands
// @tag.description Command execution and management

// @tag.name Saved Commands
// @tag.description Saved command templates for reuse

// @tag.name Command History
// @tag.description Command execution history

// @tag.name Local Users
// @tag.description Local system users for command execution

// @tag.name Environment Variables
// @tag.description Encrypted environment variable storage

// @tag.name Bash Scripts
// @tag.description Bash script storage and execution

// @tag.name Script Presets
// @tag.description Script execution configuration presets

// @tag.name Terminal
// @tag.description Interactive terminal WebSocket sessions

// @tag.name System
// @tag.description System information endpoints

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize encryption
	log.Println("Initializing encryption...")
	if err := database.InitializeEncryption(cfg.EncryptionKeyPath); err != nil {
		log.Fatalf("Failed to initialize encryption: %v", err)
	}

	// Initialize database
	log.Printf("Initializing database at %s...", cfg.DatabasePath)
	db, err := database.New(cfg.DatabasePath)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer db.Close()

	// Get database version
	version, err := db.GetVersion()
	if err != nil {
		log.Printf("Warning: failed to get database version: %v", err)
	} else {
		log.Printf("Database schema version: %d", version)
	}

	// Initialize audit logging
	if cfg.AuditLogPath != "" {
		auditLogger, err := audit.Initialize(cfg.AuditLogPath)
		if err != nil {
			log.Printf("Warning: Failed to initialize audit logging: %v", err)
		} else {
			log.Printf("Audit logging enabled: %s", cfg.AuditLogPath)
			defer auditLogger.Close()
		}
	} else {
		log.Println("Audit logging is disabled (set AUDIT_LOG_PATH to enable)")
	}

	// Set embedded frontend
	server.EmbeddedFrontend = assets.FrontendFS

	// Create and start server
	srv, err := server.New(cfg, db)
	if err != nil {
		log.Fatalf("Failed to initialize server: %v", err)
	}

	log.Fatal(srv.Start())
}
