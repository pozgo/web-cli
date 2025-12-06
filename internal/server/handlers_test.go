package server

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/gorilla/mux"
	"github.com/pozgo/web-cli/internal/database"
	"github.com/pozgo/web-cli/internal/models"
	"github.com/pozgo/web-cli/internal/repository"
)

func setupTestServer(t *testing.T) (*Server, func()) {
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

	server := &Server{
		db: db,
	}

	cleanup := func() {
		db.Close()
	}

	return server, cleanup
}

func TestHandleListBashScripts(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Create a test script
	scriptRepo := repository.NewBashScriptRepository(server.db)
	_, err := scriptRepo.Create(&models.BashScriptCreate{
		Name:    "test-script",
		Content: "#!/bin/bash\necho 'hello'",
	})
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Create request
	req, err := http.NewRequest("GET", "/api/bash-scripts", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Create response recorder
	rr := httptest.NewRecorder()

	// Call handler
	server.handleListBashScripts(rr, req)

	// Check response
	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status: got %v want %v", status, http.StatusOK)
	}

	// Parse response
	var scripts []models.BashScript
	if err := json.NewDecoder(rr.Body).Decode(&scripts); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(scripts) != 1 {
		t.Errorf("Expected 1 script, got %d", len(scripts))
	}

	if scripts[0].Name != "test-script" {
		t.Errorf("Expected script name 'test-script', got '%s'", scripts[0].Name)
	}
}

func TestHandleCreateBashScript(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Create request body
	script := models.BashScriptCreate{
		Name:        "new-script",
		Description: "A test script",
		Content:     "#!/bin/bash\necho 'test'",
		Filename:    "test.sh",
	}

	body, _ := json.Marshal(script)
	req, err := http.NewRequest("POST", "/api/bash-scripts", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.handleCreateBashScript(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status: got %v want %v. Body: %s", status, http.StatusCreated, rr.Body.String())
	}

	var created models.BashScript
	if err := json.NewDecoder(rr.Body).Decode(&created); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if created.ID == 0 {
		t.Error("Created script should have non-zero ID")
	}

	if created.Name != "new-script" {
		t.Errorf("Expected name 'new-script', got '%s'", created.Name)
	}
}

func TestHandleGetBashScript(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Create a test script first
	scriptRepo := repository.NewBashScriptRepository(server.db)
	created, err := scriptRepo.Create(&models.BashScriptCreate{
		Name:    "get-script",
		Content: "#!/bin/bash\necho 'get test'",
	})
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Create request with mux vars
	req, err := http.NewRequest("GET", "/api/bash-scripts/1", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Set mux vars
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	rr := httptest.NewRecorder()
	server.handleGetBashScript(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status: got %v want %v. Body: %s", status, http.StatusOK, rr.Body.String())
	}

	var script models.BashScript
	if err := json.NewDecoder(rr.Body).Decode(&script); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if script.Name != created.Name {
		t.Errorf("Expected name '%s', got '%s'", created.Name, script.Name)
	}
}

func TestHandleUpdateBashScript(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Create a test script first
	scriptRepo := repository.NewBashScriptRepository(server.db)
	_, err := scriptRepo.Create(&models.BashScriptCreate{
		Name:    "update-script",
		Content: "#!/bin/bash\necho 'original'",
	})
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Create update request
	update := models.BashScriptUpdate{
		Name:        "updated-script",
		Description: "Updated description",
	}

	body, _ := json.Marshal(update)
	req, err := http.NewRequest("PUT", "/api/bash-scripts/1", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	rr := httptest.NewRecorder()
	server.handleUpdateBashScript(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status: got %v want %v. Body: %s", status, http.StatusOK, rr.Body.String())
	}

	var updated models.BashScript
	if err := json.NewDecoder(rr.Body).Decode(&updated); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if updated.Name != "updated-script" {
		t.Errorf("Expected name 'updated-script', got '%s'", updated.Name)
	}
}

func TestHandleDeleteBashScript(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Create a test script first
	scriptRepo := repository.NewBashScriptRepository(server.db)
	_, err := scriptRepo.Create(&models.BashScriptCreate{
		Name:    "delete-script",
		Content: "#!/bin/bash\necho 'to be deleted'",
	})
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Create delete request
	req, err := http.NewRequest("DELETE", "/api/bash-scripts/1", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	rr := httptest.NewRecorder()
	server.handleDeleteBashScript(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("Handler returned wrong status: got %v want %v. Body: %s", status, http.StatusNoContent, rr.Body.String())
	}

	// Verify deletion
	_, err = scriptRepo.GetByID(1)
	if err == nil {
		t.Error("Expected error when getting deleted script")
	}
}

func TestHandleExecuteScript_ValidationErrors(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	tests := []struct {
		name           string
		payload        models.ScriptExecution
		expectedStatus int
		expectedError  string
	}{
		{
			name:           "missing script id",
			payload:        models.ScriptExecution{},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Script ID is required",
		},
		{
			name: "script not found",
			payload: models.ScriptExecution{
				ScriptID: 999,
			},
			expectedStatus: http.StatusNotFound,
			expectedError:  "Script not found",
		},
		{
			name: "remote without server id",
			payload: models.ScriptExecution{
				ScriptID: 1,
				IsRemote: true,
			},
			expectedStatus: http.StatusBadRequest,
			expectedError:  "Server ID is required for remote execution",
		},
	}

	// Create a test script for the "remote without server id" test
	scriptRepo := repository.NewBashScriptRepository(server.db)
	_, err := scriptRepo.Create(&models.BashScriptCreate{
		Name:    "test-script",
		Content: "#!/bin/bash\necho 'hello'",
	})
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req, err := http.NewRequest("POST", "/api/bash-scripts/execute", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			server.handleExecuteScript(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Handler returned wrong status: got %v want %v. Body: %s",
					status, tt.expectedStatus, rr.Body.String())
			}
		})
	}
}

func TestHandleCreateBashScript_ValidationErrors(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	tests := []struct {
		name           string
		payload        models.BashScriptCreate
		expectedStatus int
	}{
		{
			name: "empty name",
			payload: models.BashScriptCreate{
				Name:    "",
				Content: "#!/bin/bash\necho 'hello'",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "name with special chars",
			payload: models.BashScriptCreate{
				Name:    "test\x00script",
				Content: "#!/bin/bash\necho 'hello'",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "empty content",
			payload: models.BashScriptCreate{
				Name:    "valid-name",
				Content: "",
			},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name: "invalid filename",
			payload: models.BashScriptCreate{
				Name:     "valid-name",
				Content:  "#!/bin/bash\necho 'hello'",
				Filename: "../../../etc/passwd",
			},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req, err := http.NewRequest("POST", "/api/bash-scripts", bytes.NewBuffer(body))
			if err != nil {
				t.Fatalf("Failed to create request: %v", err)
			}
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			server.handleCreateBashScript(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Handler returned wrong status for %s: got %v want %v. Body: %s",
					tt.name, status, tt.expectedStatus, rr.Body.String())
			}
		})
	}
}

// ========== Script Preset Handler Tests ==========

func TestHandleListScriptPresets(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Create a test script first
	scriptRepo := repository.NewBashScriptRepository(server.db)
	script, err := scriptRepo.Create(&models.BashScriptCreate{
		Name:    "test-script",
		Content: "#!/bin/bash\necho 'hello'",
	})
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Create a test preset
	presetRepo := repository.NewScriptPresetRepository(server.db)
	_, err = presetRepo.Create(&models.ScriptPresetCreate{
		Name:      "test-preset",
		ScriptID:  script.ID,
		EnvVarIDs: []int64{},
	})
	if err != nil {
		t.Fatalf("Failed to create test preset: %v", err)
	}

	// Create request
	req, err := http.NewRequest("GET", "/api/script-presets", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	rr := httptest.NewRecorder()
	server.handleListScriptPresets(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status: got %v want %v", status, http.StatusOK)
	}

	var presets []models.ScriptPresetResponse
	if err := json.NewDecoder(rr.Body).Decode(&presets); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(presets) != 1 {
		t.Errorf("Expected 1 preset, got %d", len(presets))
	}

	if presets[0].Name != "test-preset" {
		t.Errorf("Expected preset name 'test-preset', got '%s'", presets[0].Name)
	}
}

func TestHandleCreateScriptPreset(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Create a test script first
	scriptRepo := repository.NewBashScriptRepository(server.db)
	script, err := scriptRepo.Create(&models.BashScriptCreate{
		Name:    "test-script",
		Content: "#!/bin/bash\necho 'hello'",
	})
	if err != nil {
		t.Fatalf("Failed to create test script: %v", err)
	}

	// Create env variables
	envRepo := repository.NewEnvVariableRepository(server.db)
	envVar, err := envRepo.Create(&models.EnvVariableCreate{
		Name:  "TEST_VAR",
		Value: "test_value",
	})
	if err != nil {
		t.Fatalf("Failed to create test env variable: %v", err)
	}

	// Create request body
	preset := models.ScriptPresetCreate{
		Name:        "new-preset",
		Description: "A test preset",
		ScriptID:    script.ID,
		EnvVarIDs:   []int64{envVar.ID},
		IsRemote:    false,
		User:        "testuser",
	}

	body, _ := json.Marshal(preset)
	req, err := http.NewRequest("POST", "/api/script-presets", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.handleCreateScriptPreset(rr, req)

	if status := rr.Code; status != http.StatusCreated {
		t.Errorf("Handler returned wrong status: got %v want %v. Body: %s", status, http.StatusCreated, rr.Body.String())
	}

	var created models.ScriptPresetResponse
	if err := json.NewDecoder(rr.Body).Decode(&created); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if created.Name != "new-preset" {
		t.Errorf("Expected name 'new-preset', got '%s'", created.Name)
	}

	if len(created.EnvVarIDs) != 1 {
		t.Errorf("Expected 1 env var ID, got %d", len(created.EnvVarIDs))
	}
}

func TestHandleGetScriptPreset(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Create a test script and preset
	scriptRepo := repository.NewBashScriptRepository(server.db)
	script, _ := scriptRepo.Create(&models.BashScriptCreate{
		Name:    "test-script",
		Content: "echo hello",
	})

	presetRepo := repository.NewScriptPresetRepository(server.db)
	preset, _ := presetRepo.Create(&models.ScriptPresetCreate{
		Name:     "test-preset",
		ScriptID: script.ID,
	})

	// Create request
	req, err := http.NewRequest("GET", "/api/script-presets/1", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	rr := httptest.NewRecorder()
	server.handleGetScriptPreset(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status: got %v want %v", status, http.StatusOK)
	}

	var retrieved models.ScriptPresetResponse
	if err := json.NewDecoder(rr.Body).Decode(&retrieved); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if retrieved.ID != preset.ID {
		t.Errorf("Expected ID %d, got %d", preset.ID, retrieved.ID)
	}
}

func TestHandleUpdateScriptPreset(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Create a test script and preset
	scriptRepo := repository.NewBashScriptRepository(server.db)
	script, _ := scriptRepo.Create(&models.BashScriptCreate{
		Name:    "test-script",
		Content: "echo hello",
	})

	presetRepo := repository.NewScriptPresetRepository(server.db)
	_, err := presetRepo.Create(&models.ScriptPresetCreate{
		Name:     "test-preset",
		ScriptID: script.ID,
	})
	if err != nil {
		t.Fatalf("Failed to create test preset: %v", err)
	}

	// Create update request
	update := models.ScriptPresetUpdate{
		Name: "updated-preset",
		User: "newuser",
	}
	body, _ := json.Marshal(update)

	req, err := http.NewRequest("PUT", "/api/script-presets/1", bytes.NewBuffer(body))
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req = mux.SetURLVars(req, map[string]string{"id": "1"})
	req.Header.Set("Content-Type", "application/json")

	rr := httptest.NewRecorder()
	server.handleUpdateScriptPreset(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status: got %v want %v. Body: %s", status, http.StatusOK, rr.Body.String())
	}

	var updated models.ScriptPresetResponse
	if err := json.NewDecoder(rr.Body).Decode(&updated); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if updated.Name != "updated-preset" {
		t.Errorf("Expected name 'updated-preset', got '%s'", updated.Name)
	}

	if updated.User != "newuser" {
		t.Errorf("Expected user 'newuser', got '%s'", updated.User)
	}
}

func TestHandleDeleteScriptPreset(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Create a test script and preset
	scriptRepo := repository.NewBashScriptRepository(server.db)
	script, _ := scriptRepo.Create(&models.BashScriptCreate{
		Name:    "test-script",
		Content: "echo hello",
	})

	presetRepo := repository.NewScriptPresetRepository(server.db)
	_, err := presetRepo.Create(&models.ScriptPresetCreate{
		Name:     "delete-preset",
		ScriptID: script.ID,
	})
	if err != nil {
		t.Fatalf("Failed to create test preset: %v", err)
	}

	// Create delete request
	req, err := http.NewRequest("DELETE", "/api/script-presets/1", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	rr := httptest.NewRecorder()
	server.handleDeleteScriptPreset(rr, req)

	if status := rr.Code; status != http.StatusNoContent {
		t.Errorf("Handler returned wrong status: got %v want %v. Body: %s", status, http.StatusNoContent, rr.Body.String())
	}

	// Verify deletion
	_, err = presetRepo.GetByID(1)
	if err == nil {
		t.Error("Expected error when getting deleted preset")
	}
}

func TestHandleGetScriptPresetsByScript(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	// Create a test script
	scriptRepo := repository.NewBashScriptRepository(server.db)
	script, _ := scriptRepo.Create(&models.BashScriptCreate{
		Name:    "test-script",
		Content: "echo hello",
	})

	// Create presets for the script
	presetRepo := repository.NewScriptPresetRepository(server.db)
	_, _ = presetRepo.Create(&models.ScriptPresetCreate{
		Name:     "preset-1",
		ScriptID: script.ID,
	})
	_, _ = presetRepo.Create(&models.ScriptPresetCreate{
		Name:     "preset-2",
		ScriptID: script.ID,
	})

	// Create request
	req, err := http.NewRequest("GET", "/api/bash-scripts/1/presets", nil)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}
	req = mux.SetURLVars(req, map[string]string{"id": "1"})

	rr := httptest.NewRecorder()
	server.handleGetScriptPresetsByScript(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status: got %v want %v", status, http.StatusOK)
	}

	var presets []models.ScriptPresetResponse
	if err := json.NewDecoder(rr.Body).Decode(&presets); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	if len(presets) != 2 {
		t.Errorf("Expected 2 presets, got %d", len(presets))
	}
}

func TestHandleCreateScriptPreset_ValidationErrors(t *testing.T) {
	server, cleanup := setupTestServer(t)
	defer cleanup()

	tests := []struct {
		name           string
		payload        models.ScriptPresetCreate
		expectedStatus int
	}{
		{
			name:           "missing name",
			payload:        models.ScriptPresetCreate{ScriptID: 1},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "missing script id",
			payload:        models.ScriptPresetCreate{Name: "test"},
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "invalid script id",
			payload:        models.ScriptPresetCreate{Name: "test", ScriptID: 999},
			expectedStatus: http.StatusBadRequest,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			body, _ := json.Marshal(tt.payload)
			req, _ := http.NewRequest("POST", "/api/script-presets", bytes.NewBuffer(body))
			req.Header.Set("Content-Type", "application/json")

			rr := httptest.NewRecorder()
			server.handleCreateScriptPreset(rr, req)

			if status := rr.Code; status != tt.expectedStatus {
				t.Errorf("Handler returned wrong status for %s: got %v want %v. Body: %s",
					tt.name, status, tt.expectedStatus, rr.Body.String())
			}
		})
	}
}
