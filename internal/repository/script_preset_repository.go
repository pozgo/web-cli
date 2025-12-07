package repository

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	"github.com/pozgo/web-cli/internal/database"
	"github.com/pozgo/web-cli/internal/models"
)

// ScriptPresetRepository handles database operations for script presets
type ScriptPresetRepository struct {
	db *database.DB
}

// NewScriptPresetRepository creates a new script preset repository
func NewScriptPresetRepository(db *database.DB) *ScriptPresetRepository {
	return &ScriptPresetRepository{db: db}
}

// Create creates a new script preset
func (r *ScriptPresetRepository) Create(preset *models.ScriptPresetCreate) (*models.ScriptPreset, error) {
	// Validate input
	if preset.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if preset.ScriptID == 0 {
		return nil, fmt.Errorf("script_id is required")
	}

	// Serialize env_var_ids to JSON
	envVarIDsJSON, err := json.Marshal(preset.EnvVarIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize env_var_ids: %w", err)
	}

	now := time.Now().UTC()

	result, err := r.db.GetConnection().Exec(
		`INSERT INTO script_presets 
		(name, description, script_id, env_var_ids, is_remote, server_id, ssh_key_id, user, created_at, updated_at) 
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		preset.Name,
		preset.Description,
		preset.ScriptID,
		string(envVarIDsJSON),
		boolToInt(preset.IsRemote),
		preset.ServerID,
		preset.SSHKeyID,
		preset.User,
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create script preset: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return &models.ScriptPreset{
		ID:          id,
		Name:        preset.Name,
		Description: preset.Description,
		ScriptID:    preset.ScriptID,
		EnvVarIDs:   preset.EnvVarIDs,
		IsRemote:    preset.IsRemote,
		ServerID:    preset.ServerID,
		SSHKeyID:    preset.SSHKeyID,
		User:        preset.User,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// GetByID retrieves a script preset by its ID
func (r *ScriptPresetRepository) GetByID(id int64) (*models.ScriptPreset, error) {
	var preset models.ScriptPreset
	var description, envVarIDsJSON, user sql.NullString
	var serverID, sshKeyID sql.NullInt64
	var isRemote int

	err := r.db.GetConnection().QueryRow(
		`SELECT id, name, description, script_id, env_var_ids, is_remote, server_id, ssh_key_id, user, created_at, updated_at 
		FROM script_presets WHERE id = ?`,
		id,
	).Scan(&preset.ID, &preset.Name, &description, &preset.ScriptID, &envVarIDsJSON, &isRemote, &serverID, &sshKeyID, &user, &preset.CreatedAt, &preset.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("script preset not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get script preset: %w", err)
	}

	// Handle nullable fields
	if description.Valid {
		preset.Description = description.String
	}
	if user.Valid {
		preset.User = user.String
	}
	if serverID.Valid {
		preset.ServerID = &serverID.Int64
	}
	if sshKeyID.Valid {
		preset.SSHKeyID = &sshKeyID.Int64
	}

	preset.IsRemote = isRemote != 0

	// Parse env_var_ids JSON
	if envVarIDsJSON.Valid && envVarIDsJSON.String != "" && envVarIDsJSON.String != "null" {
		if err := json.Unmarshal([]byte(envVarIDsJSON.String), &preset.EnvVarIDs); err != nil {
			return nil, fmt.Errorf("failed to parse env_var_ids: %w", err)
		}
	}
	// Ensure empty slice instead of nil
	if preset.EnvVarIDs == nil {
		preset.EnvVarIDs = []int64{}
	}

	return &preset, nil
}

// GetAll retrieves all script presets
func (r *ScriptPresetRepository) GetAll() ([]*models.ScriptPreset, error) {
	rows, err := r.db.GetConnection().Query(
		`SELECT id, name, description, script_id, env_var_ids, is_remote, server_id, ssh_key_id, user, created_at, updated_at 
		FROM script_presets ORDER BY name ASC`,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query script presets: %w", err)
	}
	defer rows.Close()

	var presets []*models.ScriptPreset
	for rows.Next() {
		preset, err := r.scanPreset(rows)
		if err != nil {
			return nil, err
		}
		presets = append(presets, preset)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating script presets: %w", err)
	}

	return presets, nil
}

// GetByScriptID retrieves all presets for a specific script
func (r *ScriptPresetRepository) GetByScriptID(scriptID int64) ([]*models.ScriptPreset, error) {
	rows, err := r.db.GetConnection().Query(
		`SELECT id, name, description, script_id, env_var_ids, is_remote, server_id, ssh_key_id, user, created_at, updated_at 
		FROM script_presets WHERE script_id = ? ORDER BY name ASC`,
		scriptID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query script presets: %w", err)
	}
	defer rows.Close()

	var presets []*models.ScriptPreset
	for rows.Next() {
		preset, err := r.scanPreset(rows)
		if err != nil {
			return nil, err
		}
		presets = append(presets, preset)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating script presets: %w", err)
	}

	return presets, nil
}

// Update updates an existing script preset
func (r *ScriptPresetRepository) Update(id int64, update *models.ScriptPresetUpdate) (*models.ScriptPreset, error) {
	// Get existing preset
	existing, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if update.Name != "" {
		existing.Name = update.Name
	}
	if update.Description != "" {
		existing.Description = update.Description
	}
	if update.ScriptID != nil {
		existing.ScriptID = *update.ScriptID
	}
	if update.EnvVarIDs != nil {
		existing.EnvVarIDs = update.EnvVarIDs
	}
	if update.IsRemote != nil {
		existing.IsRemote = *update.IsRemote
	}
	if update.ServerID != nil {
		existing.ServerID = update.ServerID
	}
	if update.SSHKeyID != nil {
		existing.SSHKeyID = update.SSHKeyID
	}
	if update.User != "" {
		existing.User = update.User
	}

	existing.UpdatedAt = time.Now().UTC()

	// Serialize env_var_ids to JSON
	envVarIDsJSON, err := json.Marshal(existing.EnvVarIDs)
	if err != nil {
		return nil, fmt.Errorf("failed to serialize env_var_ids: %w", err)
	}

	_, err = r.db.GetConnection().Exec(
		`UPDATE script_presets 
		SET name = ?, description = ?, script_id = ?, env_var_ids = ?, is_remote = ?, server_id = ?, ssh_key_id = ?, user = ?, updated_at = ? 
		WHERE id = ?`,
		existing.Name,
		existing.Description,
		existing.ScriptID,
		string(envVarIDsJSON),
		boolToInt(existing.IsRemote),
		existing.ServerID,
		existing.SSHKeyID,
		existing.User,
		existing.UpdatedAt,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update script preset: %w", err)
	}

	return existing, nil
}

// Delete deletes a script preset by its ID
func (r *ScriptPresetRepository) Delete(id int64) error {
	result, err := r.db.GetConnection().Exec("DELETE FROM script_presets WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete script preset: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("script preset not found")
	}

	return nil
}

// GetByName retrieves a script preset by its name
func (r *ScriptPresetRepository) GetByName(name string) (*models.ScriptPreset, error) {
	var preset models.ScriptPreset
	var description, envVarIDsJSON, user sql.NullString
	var serverID, sshKeyID sql.NullInt64
	var isRemote int

	err := r.db.GetConnection().QueryRow(
		`SELECT id, name, description, script_id, env_var_ids, is_remote, server_id, ssh_key_id, user, created_at, updated_at 
		FROM script_presets WHERE name = ?`,
		name,
	).Scan(&preset.ID, &preset.Name, &description, &preset.ScriptID, &envVarIDsJSON, &isRemote, &serverID, &sshKeyID, &user, &preset.CreatedAt, &preset.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("script preset not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get script preset: %w", err)
	}

	// Handle nullable fields
	if description.Valid {
		preset.Description = description.String
	}
	if user.Valid {
		preset.User = user.String
	}
	if serverID.Valid {
		preset.ServerID = &serverID.Int64
	}
	if sshKeyID.Valid {
		preset.SSHKeyID = &sshKeyID.Int64
	}

	preset.IsRemote = isRemote != 0

	// Parse env_var_ids JSON
	if envVarIDsJSON.Valid && envVarIDsJSON.String != "" && envVarIDsJSON.String != "null" {
		if err := json.Unmarshal([]byte(envVarIDsJSON.String), &preset.EnvVarIDs); err != nil {
			return nil, fmt.Errorf("failed to parse env_var_ids: %w", err)
		}
	}
	// Ensure empty slice instead of nil
	if preset.EnvVarIDs == nil {
		preset.EnvVarIDs = []int64{}
	}

	return &preset, nil
}

// scanPreset scans a row into a ScriptPreset
func (r *ScriptPresetRepository) scanPreset(rows *sql.Rows) (*models.ScriptPreset, error) {
	var preset models.ScriptPreset
	var description, envVarIDsJSON, user sql.NullString
	var serverID, sshKeyID sql.NullInt64
	var isRemote int

	if err := rows.Scan(&preset.ID, &preset.Name, &description, &preset.ScriptID, &envVarIDsJSON, &isRemote, &serverID, &sshKeyID, &user, &preset.CreatedAt, &preset.UpdatedAt); err != nil {
		return nil, fmt.Errorf("failed to scan script preset: %w", err)
	}

	// Handle nullable fields
	if description.Valid {
		preset.Description = description.String
	}
	if user.Valid {
		preset.User = user.String
	}
	if serverID.Valid {
		preset.ServerID = &serverID.Int64
	}
	if sshKeyID.Valid {
		preset.SSHKeyID = &sshKeyID.Int64
	}

	preset.IsRemote = isRemote != 0

	// Parse env_var_ids JSON
	if envVarIDsJSON.Valid && envVarIDsJSON.String != "" && envVarIDsJSON.String != "null" {
		if err := json.Unmarshal([]byte(envVarIDsJSON.String), &preset.EnvVarIDs); err != nil {
			return nil, fmt.Errorf("failed to parse env_var_ids: %w", err)
		}
	}
	// Ensure empty slice instead of nil
	if preset.EnvVarIDs == nil {
		preset.EnvVarIDs = []int64{}
	}

	return &preset, nil
}

// boolToInt converts a boolean to an integer (0 or 1)
func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
