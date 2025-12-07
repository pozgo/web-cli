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

func TestEnvVariableRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewEnvVariableRepository(db)

	// Test Create
	envVarCreate := &models.EnvVariableCreate{
		Name:        "API_KEY",
		Value:       "super-secret-key-12345",
		Description: "API key for external service",
	}

	created, err := repo.Create(envVarCreate)
	if err != nil {
		t.Fatalf("Failed to create env variable: %v", err)
	}

	if created.ID == 0 {
		t.Error("Created env variable should have non-zero ID")
	}

	if created.Name != envVarCreate.Name {
		t.Errorf("Name mismatch: got %s, want %s", created.Name, envVarCreate.Name)
	}

	if created.Value != envVarCreate.Value {
		t.Errorf("Value mismatch: got %s, want %s", created.Value, envVarCreate.Value)
	}

	// Test GetByID
	retrieved, err := repo.GetByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get env variable: %v", err)
	}

	if retrieved.Value != envVarCreate.Value {
		t.Error("Value should be decrypted and match original")
	}

	// Test GetByName
	retrievedByName, err := repo.GetByName("API_KEY")
	if err != nil {
		t.Fatalf("Failed to get env variable by name: %v", err)
	}

	if retrievedByName.ID != created.ID {
		t.Error("GetByName should return the same variable")
	}

	// Test GetAll
	envVars, err := repo.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all env variables: %v", err)
	}

	if len(envVars) != 1 {
		t.Errorf("Expected 1 env variable, got %d", len(envVars))
	}

	// Test Update
	update := &models.EnvVariableUpdate{
		Value:       "new-secret-key-67890",
		Description: "Updated description",
	}

	updated, err := repo.Update(created.ID, update)
	if err != nil {
		t.Fatalf("Failed to update env variable: %v", err)
	}

	if updated.Value != "new-secret-key-67890" {
		t.Errorf("Value not updated: got %s, want new-secret-key-67890", updated.Value)
	}

	if updated.Description != "Updated description" {
		t.Errorf("Description not updated: got %s", updated.Description)
	}

	// Test GetAllAsMap
	envMap, err := repo.GetAllAsMap()
	if err != nil {
		t.Fatalf("Failed to get env variables as map: %v", err)
	}

	if val, ok := envMap["API_KEY"]; !ok || val != "new-secret-key-67890" {
		t.Errorf("GetAllAsMap should return updated value, got %s", val)
	}

	// Test Delete
	if err := repo.Delete(created.ID); err != nil {
		t.Fatalf("Failed to delete env variable: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(created.ID)
	if err == nil {
		t.Error("Expected error when getting deleted env variable")
	}
}

func TestEnvVariableRepositoryDuplicateName(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewEnvVariableRepository(db)

	// Create first variable
	envVarCreate := &models.EnvVariableCreate{
		Name:  "DUPLICATE_VAR",
		Value: "value1",
	}

	_, err := repo.Create(envVarCreate)
	if err != nil {
		t.Fatalf("Failed to create first env variable: %v", err)
	}

	// Try to create duplicate
	envVarCreate.Value = "value2"
	_, err = repo.Create(envVarCreate)
	if err == nil {
		t.Error("Expected error when creating duplicate env variable name")
	}
}

func TestBashScriptRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewBashScriptRepository(db)

	// Test Create
	scriptCreate := &models.BashScriptCreate{
		Name:        "Test Script",
		Description: "A test backup script",
		Content:     "#!/bin/bash\necho 'Hello World'\nexit 0",
		Filename:    "test-script.sh",
	}

	created, err := repo.Create(scriptCreate)
	if err != nil {
		t.Fatalf("Failed to create bash script: %v", err)
	}

	if created.ID == 0 {
		t.Error("Expected non-zero ID")
	}

	if created.Name != scriptCreate.Name {
		t.Errorf("Name mismatch: got %s, want %s", created.Name, scriptCreate.Name)
	}

	if created.Content != scriptCreate.Content {
		t.Error("Content should be returned unencrypted")
	}

	// Test GetByID
	retrieved, err := repo.GetByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get bash script: %v", err)
	}

	if retrieved.Content != scriptCreate.Content {
		t.Error("Content should be decrypted and match original")
	}

	if retrieved.Filename != scriptCreate.Filename {
		t.Errorf("Filename mismatch: got %s, want %s", retrieved.Filename, scriptCreate.Filename)
	}

	// Test GetByName
	retrievedByName, err := repo.GetByName("Test Script")
	if err != nil {
		t.Fatalf("Failed to get bash script by name: %v", err)
	}

	if retrievedByName.ID != created.ID {
		t.Error("GetByName should return the same script")
	}

	// Test GetAll
	scripts, err := repo.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all bash scripts: %v", err)
	}

	if len(scripts) != 1 {
		t.Errorf("Expected 1 bash script, got %d", len(scripts))
	}

	// Test Update
	update := &models.BashScriptUpdate{
		Content:     "#!/bin/bash\nset -e\necho 'Updated Script'\nexit 0",
		Description: "Updated description",
	}

	updated, err := repo.Update(created.ID, update)
	if err != nil {
		t.Fatalf("Failed to update bash script: %v", err)
	}

	if updated.Content != update.Content {
		t.Error("Content not updated")
	}

	if updated.Description != "Updated description" {
		t.Errorf("Description not updated: got %s", updated.Description)
	}

	// Verify update persisted
	retrieved, err = repo.GetByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get updated bash script: %v", err)
	}

	if retrieved.Content != update.Content {
		t.Error("Updated content not persisted")
	}

	// Test Delete
	if err := repo.Delete(created.ID); err != nil {
		t.Fatalf("Failed to delete bash script: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(created.ID)
	if err == nil {
		t.Error("Expected error when getting deleted bash script")
	}
}

func TestBashScriptRepositoryWithoutOptionalFields(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewBashScriptRepository(db)

	// Test Create without optional fields
	scriptCreate := &models.BashScriptCreate{
		Name:    "Minimal Script",
		Content: "echo hello",
	}

	created, err := repo.Create(scriptCreate)
	if err != nil {
		t.Fatalf("Failed to create bash script: %v", err)
	}

	if created.Description != "" {
		t.Error("Description should be empty")
	}

	if created.Filename != "" {
		t.Error("Filename should be empty")
	}

	// Test GetByID with null optional fields
	retrieved, err := repo.GetByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get bash script: %v", err)
	}

	if retrieved.Description != "" {
		t.Error("Description should be empty after retrieval")
	}

	if retrieved.Filename != "" {
		t.Error("Filename should be empty after retrieval")
	}
}

func TestScriptPresetRepository(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// First, create a bash script to reference
	bashScriptRepo := NewBashScriptRepository(db)
	script, err := bashScriptRepo.Create(&models.BashScriptCreate{
		Name:    "Test Script",
		Content: "#!/bin/bash\necho 'Hello'",
	})
	if err != nil {
		t.Fatalf("Failed to create test bash script: %v", err)
	}

	// Create env variables to reference
	envVarRepo := NewEnvVariableRepository(db)
	envVar1, err := envVarRepo.Create(&models.EnvVariableCreate{
		Name:  "VAR1",
		Value: "value1",
	})
	if err != nil {
		t.Fatalf("Failed to create test env variable 1: %v", err)
	}
	envVar2, err := envVarRepo.Create(&models.EnvVariableCreate{
		Name:  "VAR2",
		Value: "value2",
	})
	if err != nil {
		t.Fatalf("Failed to create test env variable 2: %v", err)
	}

	repo := NewScriptPresetRepository(db)

	// Test Create
	presetCreate := &models.ScriptPresetCreate{
		Name:        "My Preset",
		Description: "A test preset",
		ScriptID:    script.ID,
		EnvVarIDs:   []int64{envVar1.ID, envVar2.ID},
		IsRemote:    false,
		User:        "testuser",
	}

	created, err := repo.Create(presetCreate)
	if err != nil {
		t.Fatalf("Failed to create script preset: %v", err)
	}

	if created.ID == 0 {
		t.Error("Created script preset should have non-zero ID")
	}

	if created.Name != presetCreate.Name {
		t.Errorf("Name mismatch: got %s, want %s", created.Name, presetCreate.Name)
	}

	if len(created.EnvVarIDs) != 2 {
		t.Errorf("Expected 2 env var IDs, got %d", len(created.EnvVarIDs))
	}

	// Test GetByID
	retrieved, err := repo.GetByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get script preset: %v", err)
	}

	if retrieved.Name != presetCreate.Name {
		t.Errorf("Name mismatch after retrieval: got %s, want %s", retrieved.Name, presetCreate.Name)
	}

	if len(retrieved.EnvVarIDs) != 2 {
		t.Errorf("Expected 2 env var IDs after retrieval, got %d", len(retrieved.EnvVarIDs))
	}

	if retrieved.ScriptID != script.ID {
		t.Errorf("ScriptID mismatch: got %d, want %d", retrieved.ScriptID, script.ID)
	}

	// Test GetByName
	retrievedByName, err := repo.GetByName("My Preset")
	if err != nil {
		t.Fatalf("Failed to get script preset by name: %v", err)
	}

	if retrievedByName.ID != created.ID {
		t.Error("GetByName should return the same preset")
	}

	// Test GetAll
	presets, err := repo.GetAll()
	if err != nil {
		t.Fatalf("Failed to get all script presets: %v", err)
	}

	if len(presets) != 1 {
		t.Errorf("Expected 1 script preset, got %d", len(presets))
	}

	// Test GetByScriptID
	presetsByScript, err := repo.GetByScriptID(script.ID)
	if err != nil {
		t.Fatalf("Failed to get script presets by script ID: %v", err)
	}

	if len(presetsByScript) != 1 {
		t.Errorf("Expected 1 script preset for script, got %d", len(presetsByScript))
	}

	// Test Update
	newIsRemote := true
	update := &models.ScriptPresetUpdate{
		Name:     "Updated Preset",
		IsRemote: &newIsRemote,
		User:     "newuser",
	}

	updated, err := repo.Update(created.ID, update)
	if err != nil {
		t.Fatalf("Failed to update script preset: %v", err)
	}

	if updated.Name != "Updated Preset" {
		t.Errorf("Name not updated: got %s, want Updated Preset", updated.Name)
	}

	if !updated.IsRemote {
		t.Error("IsRemote should be true after update")
	}

	if updated.User != "newuser" {
		t.Errorf("User not updated: got %s, want newuser", updated.User)
	}

	// Test Delete
	if err := repo.Delete(created.ID); err != nil {
		t.Fatalf("Failed to delete script preset: %v", err)
	}

	// Verify deletion
	_, err = repo.GetByID(created.ID)
	if err == nil {
		t.Error("Expected error when getting deleted script preset")
	}
}

func TestScriptPresetRepositoryRemoteExecution(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	// Create dependencies
	bashScriptRepo := NewBashScriptRepository(db)
	script, err := bashScriptRepo.Create(&models.BashScriptCreate{
		Name:    "Remote Script",
		Content: "#!/bin/bash\necho 'Remote'",
	})
	if err != nil {
		t.Fatalf("Failed to create test bash script: %v", err)
	}

	serverRepo := NewServerRepository(db)
	server, err := serverRepo.Create(&models.ServerCreate{
		Name:      "Test Server",
		IPAddress: "192.168.1.100",
		Port:      22,
		Username:  "admin",
	})
	if err != nil {
		t.Fatalf("Failed to create test server: %v", err)
	}

	sshKeyRepo := NewSSHKeyRepository(db)
	sshKey, err := sshKeyRepo.Create(&models.SSHKeyCreate{
		Name:       "Test Key",
		PrivateKey: "-----BEGIN RSA PRIVATE KEY-----\ntest\n-----END RSA PRIVATE KEY-----",
	})
	if err != nil {
		t.Fatalf("Failed to create test SSH key: %v", err)
	}

	repo := NewScriptPresetRepository(db)

	// Test Create with remote settings
	presetCreate := &models.ScriptPresetCreate{
		Name:        "Remote Preset",
		Description: "A preset for remote execution",
		ScriptID:    script.ID,
		EnvVarIDs:   []int64{},
		IsRemote:    true,
		ServerID:    &server.ID,
		SSHKeyID:    &sshKey.ID,
		User:        "root",
	}

	created, err := repo.Create(presetCreate)
	if err != nil {
		t.Fatalf("Failed to create remote script preset: %v", err)
	}

	if !created.IsRemote {
		t.Error("IsRemote should be true")
	}

	if created.ServerID == nil || *created.ServerID != server.ID {
		t.Errorf("ServerID mismatch: got %v, want %d", created.ServerID, server.ID)
	}

	if created.SSHKeyID == nil || *created.SSHKeyID != sshKey.ID {
		t.Errorf("SSHKeyID mismatch: got %v, want %d", created.SSHKeyID, sshKey.ID)
	}

	// Test GetByID with nullable foreign keys
	retrieved, err := repo.GetByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get remote script preset: %v", err)
	}

	if retrieved.ServerID == nil || *retrieved.ServerID != server.ID {
		t.Error("ServerID should be preserved after retrieval")
	}

	if retrieved.SSHKeyID == nil || *retrieved.SSHKeyID != sshKey.ID {
		t.Error("SSHKeyID should be preserved after retrieval")
	}
}

func TestScriptPresetRepositoryEmptyEnvVars(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	bashScriptRepo := NewBashScriptRepository(db)
	script, err := bashScriptRepo.Create(&models.BashScriptCreate{
		Name:    "Test Script",
		Content: "echo test",
	})
	if err != nil {
		t.Fatalf("Failed to create test bash script: %v", err)
	}

	repo := NewScriptPresetRepository(db)

	// Test Create with nil env_var_ids
	presetCreate := &models.ScriptPresetCreate{
		Name:      "Empty Vars Preset",
		ScriptID:  script.ID,
		EnvVarIDs: nil,
	}

	created, err := repo.Create(presetCreate)
	if err != nil {
		t.Fatalf("Failed to create script preset with nil env vars: %v", err)
	}

	// Test GetByID - should return empty slice, not nil
	retrieved, err := repo.GetByID(created.ID)
	if err != nil {
		t.Fatalf("Failed to get script preset: %v", err)
	}

	if retrieved.EnvVarIDs == nil {
		t.Error("EnvVarIDs should be empty slice, not nil")
	}

	if len(retrieved.EnvVarIDs) != 0 {
		t.Errorf("Expected 0 env var IDs, got %d", len(retrieved.EnvVarIDs))
	}
}

func TestScriptPresetRepositoryValidation(t *testing.T) {
	db := setupTestDB(t)
	defer db.Close()

	repo := NewScriptPresetRepository(db)

	// Test Create without name
	_, err := repo.Create(&models.ScriptPresetCreate{
		Name:     "",
		ScriptID: 1,
	})
	if err == nil {
		t.Error("Expected error when creating preset without name")
	}

	// Test Create without script_id
	_, err = repo.Create(&models.ScriptPresetCreate{
		Name:     "Test",
		ScriptID: 0,
	})
	if err == nil {
		t.Error("Expected error when creating preset without script_id")
	}
}
