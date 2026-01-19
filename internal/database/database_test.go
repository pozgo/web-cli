package database

import (
	"os"
	"path/filepath"
	"testing"
)

func TestEncryptDecrypt(t *testing.T) {
	// Create temporary encryption key
	tmpDir := t.TempDir()
	keyPath := filepath.Join(tmpDir, ".encryption_key")

	if err := InitializeEncryption(keyPath); err != nil {
		t.Fatalf("Failed to initialize encryption: %v", err)
	}

	// Test data
	plaintext := "test-ssh-key-content-here"

	// Encrypt
	ciphertext, err := Encrypt(plaintext)
	if err != nil {
		t.Fatalf("Failed to encrypt: %v", err)
	}

	// Ensure encrypted data is different from plaintext
	if string(ciphertext) == plaintext {
		t.Error("Ciphertext should be different from plaintext")
	}

	// Decrypt
	decrypted, err := Decrypt(ciphertext)
	if err != nil {
		t.Fatalf("Failed to decrypt: %v", err)
	}

	// Verify decrypted matches original
	if decrypted != plaintext {
		t.Errorf("Decrypted text doesn't match original. Got %s, want %s", decrypted, plaintext)
	}
}

func TestDatabaseCreation(t *testing.T) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	keyPath := filepath.Join(tmpDir, ".encryption_key")

	// Initialize encryption
	if err := InitializeEncryption(keyPath); err != nil {
		t.Fatalf("Failed to initialize encryption: %v", err)
	}

	// Create database
	db, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}
	defer db.Close()

	// Verify database file exists
	if _, err := os.Stat(dbPath); os.IsNotExist(err) {
		t.Error("Database file was not created")
	}

	// Verify schema version
	version, err := db.GetVersion()
	if err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}

	if version != 16 {
		t.Errorf("Expected schema version 16, got %d", version)
	}

	// Verify all tables exist
	tables := []string{
		"schema_migrations",
		"ssh_keys",
		"command_history",
		"servers",
		"saved_commands",
		"local_users",
		"env_variables",
		"bash_scripts",
		"vault_config",
	}

	for _, table := range tables {
		var name string
		err := db.conn.QueryRow("SELECT name FROM sqlite_master WHERE type='table' AND name=?", table).Scan(&name)
		if err != nil {
			t.Errorf("Table %s does not exist: %v", table, err)
		}
	}

	// Verify ssh_keys table has correct columns (including migration 4 rename)
	var columnName string
	err = db.conn.QueryRow("SELECT name FROM pragma_table_info('ssh_keys') WHERE name='private_key_encrypted'").Scan(&columnName)
	if err != nil {
		t.Error("ssh_keys table should have private_key_encrypted column after migration 4")
	}

	// Verify servers table has port column (migration 6)
	err = db.conn.QueryRow("SELECT name FROM pragma_table_info('servers') WHERE name='port'").Scan(&columnName)
	if err != nil {
		t.Error("servers table should have port column after migration 6")
	}

	// Verify servers table has username column (migration 9)
	err = db.conn.QueryRow("SELECT name FROM pragma_table_info('servers') WHERE name='username'").Scan(&columnName)
	if err != nil {
		t.Error("servers table should have username column after migration 9")
	}

	// Verify command_history has user column (migration 7)
	err = db.conn.QueryRow("SELECT name FROM pragma_table_info('command_history') WHERE name='user'").Scan(&columnName)
	if err != nil {
		t.Error("command_history table should have user column after migration 7")
	}

	// Verify saved_commands has remote command fields (migration 11)
	remoteFields := []string{"is_remote", "server_id", "ssh_key_id"}
	for _, field := range remoteFields {
		err = db.conn.QueryRow("SELECT name FROM pragma_table_info('saved_commands') WHERE name=?", field).Scan(&columnName)
		if err != nil {
			t.Errorf("saved_commands table should have %s column after migration 11", field)
		}
	}

	// Verify env_variables table has correct columns (migration 12)
	envVarFields := []string{"id", "name", "value_encrypted", "description", "created_at", "updated_at"}
	for _, field := range envVarFields {
		err = db.conn.QueryRow("SELECT name FROM pragma_table_info('env_variables') WHERE name=?", field).Scan(&columnName)
		if err != nil {
			t.Errorf("env_variables table should have %s column after migration 12", field)
		}
	}

	// Verify bash_scripts table has correct columns (migration 13)
	bashScriptFields := []string{"id", "name", "description", "content_encrypted", "filename", "created_at", "updated_at"}
	for _, field := range bashScriptFields {
		err = db.conn.QueryRow("SELECT name FROM pragma_table_info('bash_scripts') WHERE name=?", field).Scan(&columnName)
		if err != nil {
			t.Errorf("bash_scripts table should have %s column after migration 13", field)
		}
	}

	// Verify group_name column exists on resource tables (migration 16)
	groupNameTables := []string{"servers", "ssh_keys", "env_variables", "bash_scripts"}
	for _, table := range groupNameTables {
		err = db.conn.QueryRow("SELECT name FROM pragma_table_info(?) WHERE name='group_name'", table).Scan(&columnName)
		if err != nil {
			t.Errorf("%s table should have group_name column after migration 16", table)
		}
	}
}

func TestDatabasePersistence(t *testing.T) {
	// Create temporary directory for test database
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	keyPath := filepath.Join(tmpDir, ".encryption_key")

	// Initialize encryption
	if err := InitializeEncryption(keyPath); err != nil {
		t.Fatalf("Failed to initialize encryption: %v", err)
	}

	// Create database first time
	db1, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	version1, err := db1.GetVersion()
	if err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}

	db1.Close()

	// Open database second time
	db2, err := New(dbPath)
	if err != nil {
		t.Fatalf("Failed to open existing database: %v", err)
	}
	defer db2.Close()

	version2, err := db2.GetVersion()
	if err != nil {
		t.Fatalf("Failed to get version: %v", err)
	}

	// Versions should match
	if version1 != version2 {
		t.Errorf("Version mismatch: first=%d, second=%d", version1, version2)
	}
}
