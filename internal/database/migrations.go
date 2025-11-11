package database

import (
	"database/sql"
	"fmt"
	"log"
	"time"
)

// Migration represents a database migration
type Migration struct {
	Version     int
	Description string
	SQL         string
}

// migrations contains all database migrations in order
var migrations = []Migration{
	{
		Version:     1,
		Description: "Create schema_migrations table",
		SQL: `
			CREATE TABLE IF NOT EXISTS schema_migrations (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				version INTEGER NOT NULL UNIQUE,
				description TEXT NOT NULL,
				applied_at DATETIME NOT NULL
			);
		`,
	},
	{
		Version:     2,
		Description: "Create ssh_keys table",
		SQL: `
			CREATE TABLE IF NOT EXISTS ssh_keys (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				public_key_encrypted BLOB NOT NULL,
				created_at DATETIME NOT NULL,
				updated_at DATETIME NOT NULL
			);
			CREATE INDEX IF NOT EXISTS idx_ssh_keys_name ON ssh_keys(name);
		`,
	},
	{
		Version:     3,
		Description: "Create command_history table",
		SQL: `
			CREATE TABLE IF NOT EXISTS command_history (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				command_encrypted BLOB NOT NULL,
				output_encrypted BLOB,
				exit_code INTEGER,
				server TEXT NOT NULL,
				execution_time_ms INTEGER,
				executed_at DATETIME NOT NULL
			);
			CREATE INDEX IF NOT EXISTS idx_command_history_executed_at ON command_history(executed_at DESC);
			CREATE INDEX IF NOT EXISTS idx_command_history_server ON command_history(server);
		`,
	},
	{
		Version:     4,
		Description: "Rename public_key_encrypted to private_key_encrypted in ssh_keys table",
		SQL: `
			ALTER TABLE ssh_keys RENAME COLUMN public_key_encrypted TO private_key_encrypted;
		`,
	},
	{
		Version:     5,
		Description: "Create servers table",
		SQL: `
			CREATE TABLE IF NOT EXISTS servers (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT,
				ip_address TEXT,
				created_at DATETIME NOT NULL,
				updated_at DATETIME NOT NULL,
				CHECK (name IS NOT NULL OR ip_address IS NOT NULL)
			);
			CREATE INDEX IF NOT EXISTS idx_servers_name ON servers(name);
			CREATE INDEX IF NOT EXISTS idx_servers_ip_address ON servers(ip_address);
		`,
	},
	{
		Version:     6,
		Description: "Add port column to servers table",
		SQL: `
			ALTER TABLE servers ADD COLUMN port INTEGER NOT NULL DEFAULT 22;
		`,
	},
	{
		Version:     7,
		Description: "Add user column to command_history table",
		SQL: `
			ALTER TABLE command_history ADD COLUMN user TEXT;
		`,
	},
	{
		Version:     8,
		Description: "Create saved_commands table",
		SQL: `
			CREATE TABLE IF NOT EXISTS saved_commands (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL,
				command TEXT NOT NULL,
				description TEXT,
				user TEXT NOT NULL DEFAULT 'root',
				created_at DATETIME NOT NULL,
				updated_at DATETIME NOT NULL
			);
			CREATE INDEX IF NOT EXISTS idx_saved_commands_name ON saved_commands(name);
		`,
	},
	{
		Version:     9,
		Description: "Add username column to servers table",
		SQL: `
			ALTER TABLE servers ADD COLUMN username TEXT NOT NULL DEFAULT 'root';
		`,
	},
	{
		Version:     10,
		Description: "Create local_users table",
		SQL: `
			CREATE TABLE IF NOT EXISTS local_users (
				id INTEGER PRIMARY KEY AUTOINCREMENT,
				name TEXT NOT NULL UNIQUE,
				created_at DATETIME NOT NULL,
				updated_at DATETIME NOT NULL
			);
			CREATE INDEX IF NOT EXISTS idx_local_users_name ON local_users(name);
		`,
	},
	{
		Version:     11,
		Description: "Add remote command fields to saved_commands table",
		SQL: `
			ALTER TABLE saved_commands ADD COLUMN is_remote INTEGER NOT NULL DEFAULT 0;
			ALTER TABLE saved_commands ADD COLUMN server_id INTEGER;
			ALTER TABLE saved_commands ADD COLUMN ssh_key_id INTEGER;

			CREATE INDEX IF NOT EXISTS idx_saved_commands_server ON saved_commands(server_id);
			CREATE INDEX IF NOT EXISTS idx_saved_commands_ssh_key ON saved_commands(ssh_key_id);
		`,
	},
}

// runMigrations executes all pending migrations
func (db *DB) runMigrations() error {
	// First, ensure the schema_migrations table exists
	if _, err := db.conn.Exec(migrations[0].SQL); err != nil {
		return fmt.Errorf("failed to create schema_migrations table: %w", err)
	}

	// Get current version
	currentVersion := db.getCurrentVersion()

	// Apply pending migrations
	for _, migration := range migrations[1:] { // Skip first migration as it's already applied
		if migration.Version <= currentVersion {
			continue
		}

		log.Printf("Applying migration %d: %s", migration.Version, migration.Description)

		// Start transaction
		tx, err := db.conn.Begin()
		if err != nil {
			return fmt.Errorf("failed to begin transaction: %w", err)
		}

		// Execute migration
		if _, err := tx.Exec(migration.SQL); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to execute migration %d: %w", migration.Version, err)
		}

		// Record migration
		if _, err := tx.Exec(
			"INSERT INTO schema_migrations (version, description, applied_at) VALUES (?, ?, ?)",
			migration.Version,
			migration.Description,
			time.Now().UTC(),
		); err != nil {
			tx.Rollback()
			return fmt.Errorf("failed to record migration %d: %w", migration.Version, err)
		}

		// Commit transaction
		if err := tx.Commit(); err != nil {
			return fmt.Errorf("failed to commit migration %d: %w", migration.Version, err)
		}

		log.Printf("Successfully applied migration %d", migration.Version)
	}

	return nil
}

// getCurrentVersion returns the current schema version
func (db *DB) getCurrentVersion() int {
	var version int
	err := db.conn.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	if err != nil && err != sql.ErrNoRows {
		log.Printf("Warning: failed to get current version: %v", err)
		return 0
	}
	return version
}

// GetVersion returns the current database schema version
func (db *DB) GetVersion() (int, error) {
	var version int
	err := db.conn.QueryRow("SELECT COALESCE(MAX(version), 0) FROM schema_migrations").Scan(&version)
	if err != nil {
		return 0, err
	}
	return version, nil
}
