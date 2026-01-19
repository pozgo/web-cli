package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pozgo/web-cli/internal/database"
	"github.com/pozgo/web-cli/internal/models"
)

// EnvVariableRepository handles database operations for environment variables
type EnvVariableRepository struct {
	db *database.DB
}

// NewEnvVariableRepository creates a new environment variable repository
func NewEnvVariableRepository(db *database.DB) *EnvVariableRepository {
	return &EnvVariableRepository{db: db}
}

// Create creates a new environment variable with encrypted value
func (r *EnvVariableRepository) Create(envVar *models.EnvVariableCreate) (*models.EnvVariable, error) {
	// Validate input
	if envVar.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if envVar.Value == "" {
		return nil, fmt.Errorf("value is required")
	}

	// Default group to "default" if not provided
	group := envVar.Group
	if group == "" {
		group = "default"
	}

	// Encrypt the value
	encryptedValue, err := database.Encrypt(envVar.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt value: %w", err)
	}

	now := time.Now().UTC()

	result, err := r.db.GetConnection().Exec(
		"INSERT INTO env_variables (name, value_encrypted, description, group_name, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		envVar.Name,
		encryptedValue,
		envVar.Description,
		group,
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create environment variable: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return &models.EnvVariable{
		ID:          id,
		Name:        envVar.Name,
		Value:       envVar.Value, // Return unencrypted value
		Description: envVar.Description,
		Group:       group,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// GetByID retrieves an environment variable by its ID
func (r *EnvVariableRepository) GetByID(id int64) (*models.EnvVariable, error) {
	var envVar models.EnvVariable
	var encryptedValue []byte

	err := r.db.GetConnection().QueryRow(
		"SELECT id, name, value_encrypted, description, group_name, created_at, updated_at FROM env_variables WHERE id = ?",
		id,
	).Scan(&envVar.ID, &envVar.Name, &encryptedValue, &envVar.Description, &envVar.Group, &envVar.CreatedAt, &envVar.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("environment variable not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get environment variable: %w", err)
	}

	// Decrypt the value
	decryptedValue, err := database.Decrypt(encryptedValue)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt value: %w", err)
	}
	envVar.Value = decryptedValue

	return &envVar, nil
}

// GetByName retrieves an environment variable by its name
func (r *EnvVariableRepository) GetByName(name string) (*models.EnvVariable, error) {
	var envVar models.EnvVariable
	var encryptedValue []byte

	err := r.db.GetConnection().QueryRow(
		"SELECT id, name, value_encrypted, description, group_name, created_at, updated_at FROM env_variables WHERE name = ?",
		name,
	).Scan(&envVar.ID, &envVar.Name, &encryptedValue, &envVar.Description, &envVar.Group, &envVar.CreatedAt, &envVar.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("environment variable not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get environment variable: %w", err)
	}

	// Decrypt the value
	decryptedValue, err := database.Decrypt(encryptedValue)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt value: %w", err)
	}
	envVar.Value = decryptedValue

	return &envVar, nil
}

// GetAll retrieves all environment variables
func (r *EnvVariableRepository) GetAll() ([]*models.EnvVariable, error) {
	rows, err := r.db.GetConnection().Query(
		"SELECT id, name, value_encrypted, description, group_name, created_at, updated_at FROM env_variables ORDER BY group_name ASC, name ASC",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query environment variables: %w", err)
	}
	defer rows.Close()

	var envVars []*models.EnvVariable
	for rows.Next() {
		var envVar models.EnvVariable
		var encryptedValue []byte

		if err := rows.Scan(&envVar.ID, &envVar.Name, &encryptedValue, &envVar.Description, &envVar.Group, &envVar.CreatedAt, &envVar.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan environment variable: %w", err)
		}

		// Decrypt the value
		decryptedValue, err := database.Decrypt(encryptedValue)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt value: %w", err)
		}
		envVar.Value = decryptedValue

		envVars = append(envVars, &envVar)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating environment variables: %w", err)
	}

	return envVars, nil
}

// GetByGroup retrieves all environment variables in a specific group
func (r *EnvVariableRepository) GetByGroup(group string) ([]*models.EnvVariable, error) {
	rows, err := r.db.GetConnection().Query(
		"SELECT id, name, value_encrypted, description, group_name, created_at, updated_at FROM env_variables WHERE group_name = ? ORDER BY name ASC",
		group,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query environment variables: %w", err)
	}
	defer rows.Close()

	var envVars []*models.EnvVariable
	for rows.Next() {
		var envVar models.EnvVariable
		var encryptedValue []byte

		if err := rows.Scan(&envVar.ID, &envVar.Name, &encryptedValue, &envVar.Description, &envVar.Group, &envVar.CreatedAt, &envVar.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan environment variable: %w", err)
		}

		// Decrypt the value
		decryptedValue, err := database.Decrypt(encryptedValue)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt value: %w", err)
		}
		envVar.Value = decryptedValue

		envVars = append(envVars, &envVar)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating environment variables: %w", err)
	}

	return envVars, nil
}

// GetGroups retrieves all distinct group names
func (r *EnvVariableRepository) GetGroups() ([]string, error) {
	rows, err := r.db.GetConnection().Query(
		"SELECT DISTINCT group_name FROM env_variables ORDER BY group_name ASC",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query groups: %w", err)
	}
	defer rows.Close()

	var groups []string
	for rows.Next() {
		var group string
		if err := rows.Scan(&group); err != nil {
			return nil, fmt.Errorf("failed to scan group: %w", err)
		}
		groups = append(groups, group)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating groups: %w", err)
	}

	return groups, nil
}

// Update updates an existing environment variable
func (r *EnvVariableRepository) Update(id int64, update *models.EnvVariableUpdate) (*models.EnvVariable, error) {
	// Get existing variable
	existing, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if update.Name != "" {
		existing.Name = update.Name
	}

	if update.Value != "" {
		existing.Value = update.Value
	}

	// Description can be cleared, so always update it if the field is present
	if update.Description != "" {
		existing.Description = update.Description
	}

	if update.Group != "" {
		existing.Group = update.Group
	}

	existing.UpdatedAt = time.Now().UTC()

	// Encrypt the value
	encryptedValue, err := database.Encrypt(existing.Value)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt value: %w", err)
	}

	_, err = r.db.GetConnection().Exec(
		"UPDATE env_variables SET name = ?, value_encrypted = ?, description = ?, group_name = ?, updated_at = ? WHERE id = ?",
		existing.Name,
		encryptedValue,
		existing.Description,
		existing.Group,
		existing.UpdatedAt,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update environment variable: %w", err)
	}

	return existing, nil
}

// Delete deletes an environment variable by its ID
func (r *EnvVariableRepository) Delete(id int64) error {
	result, err := r.db.GetConnection().Exec("DELETE FROM env_variables WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete environment variable: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("environment variable not found")
	}

	return nil
}

// GetAllAsMap returns all environment variables as a map for command execution
func (r *EnvVariableRepository) GetAllAsMap() (map[string]string, error) {
	envVars, err := r.GetAll()
	if err != nil {
		return nil, err
	}
	return models.EnvVariablesToMap(envVars), nil
}
