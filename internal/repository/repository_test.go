package repository

import (
	"path/filepath"
	"testing"

	"github.com/pozgo/web-cli/internal/database"
	"github.com/pozgo/web-cli/internal/models"
)

func setupTestDB(t *testing.T) *database.DB {
	tmpDir := t.TempDir()
	dbPath := filepath.Join(tmpDir, "test.db")
	keyPath := filepath.Join(tmpDir, ".encryption_key")

	if err := database.InitializeEncryption(keyPath); err != nil {
		t.Fatalf("Failed to initialize encryption: %v", err)
	}

	db, err := database.New(dbPath)
	if err != nil {
		t.Fatalf("Failed to create database: %v", err)
	}

	return db
}

func TestSSHKeyRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewSSHKeyRepository(db)

	// Test Create
	keyCreate := &models.SSHKeyCreate{
		Name:       "test-key",
		PrivateKey: "-----BEGIN RSA PRIVATE KEY-----\nMIIEpAIBAAKCAQEA...\n-----END RSA PRIVATE KEY-----",
	}

	created, err := repo.Create(keyCreate)
	if err != nil {
		t.Fatalf("Failed to create SSH key: %v", err)
	}

	if created.ID == 0 {
		t.Error("Created SSH key should have non-zero ID")
	}

	if created.Name != keyCreate.Name {
		t.Errorf("Name mismatch: got %s, want %s", created.Name, keyCreate.Name)
	}

	// Test GetByID
	retrieved, err := repo.GetByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get SSH key: %v", err)
	}

	if retrieved.PrivateKey != keyCreate.PrivateKey {
		t.Error("Private key should be decrypted and match original")
	}

	// Test GetAll
	keys, err := repo.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all SSH keys: %v", err)
	}

	if len(keys) != 1 {
		t.Errorf("Expected 1 key, got %d", len(keys))
	}

	// Test Update
	update := &models.SSHKeyUpdate{
		Name: "updated-key",
	}

	updated, err := repo.Update(created.ID, update)
	if err != nil {
		t.Fatalf("Failed to update SSH key: %v", err)
	}

	if updated.Name != "updated-key" {
		t.Errorf("Name not updated: got %s, want updated-key", updated.Name)
	}

	// Test Delete
	if err := repo.Delete(created.ID); err != nil {
		t.Fatalf("Failed to delete SSH key: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(created.ID)
	if err == nil {
		t.Error("Expected error when getting deleted key")
	}
}

func TestCommandHistoryRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewCommandHistoryRepository(db)

	exitCode := 0

	// Test Create
	historyCreate := &models.CommandHistoryCreate{
		Command:         "ls -la",
		Output:          "total 0\ndrwxr-xr-x...",
		ExitCode:        &exitCode,
		Server:          "local",
		ExecutionTimeMs: 150,
	}

	created, err := repo.Create(historyCreate)
	if err != nil {
		t.Fatalf("Failed to create command history: %v", err)
	}

	if created.ID == 0 {
		t.Error("Created command history should have non-zero ID")
	}

	// Test GetByID
	retrieved, err := repo.GetByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get command history: %v", err)
	}

	if retrieved.Command != historyCreate.Command {
		t.Error("Command should be decrypted and match original")
	}

	if retrieved.Output != historyCreate.Output {
		t.Error("Output should be decrypted and match original")
	}

	// Test GetAll
	histories, err := repo.GetAll(10)
	if err != nil {
		t.Fatalf("Failed to get all command histories: %v", err)
	}

	if len(histories) != 1 {
		t.Errorf("Expected 1 history, got %d", len(histories))
	}

	// Test GetByServer
	serverHistories, err := repo.GetByServer("local", 10)
	if err != nil {
		t.Fatalf("Failed to get command histories by server: %v", err)
	}

	if len(serverHistories) != 1 {
		t.Errorf("Expected 1 history for server 'local', got %d", len(serverHistories))
	}

	// Test Delete
	if err := repo.Delete(created.ID); err != nil {
		t.Fatalf("Failed to delete command history: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(created.ID)
	if err == nil {
		t.Error("Expected error when getting deleted history")
	}
}
