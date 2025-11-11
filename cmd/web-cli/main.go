package main

import (
	"log"

	"github.com/pozgo/web-cli/assets"
	"github.com/pozgo/web-cli/internal/config"
	"github.com/pozgo/web-cli/internal/database"
	"github.com/pozgo/web-cli/internal/server"
)

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

	// Set embedded frontend
	server.EmbeddedFrontend = assets.FrontendFS

	// Create and start server
	srv := server.New(cfg, db)

	log.Fatal(srv.Start())
}
