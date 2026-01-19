package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pozgo/web-cli/internal/database"
	"github.com/pozgo/web-cli/internal/models"
)

// BashScriptRepository handles database operations for bash scripts
type BashScriptRepository struct {
	db *database.DB
}

// NewBashScriptRepository creates a new bash script repository
func NewBashScriptRepository(db *database.DB) *BashScriptRepository {
	return &BashScriptRepository{db: db}
}

// Create creates a new bash script with encrypted content
func (r *BashScriptRepository) Create(script *models.BashScriptCreate) (*models.BashScript, error) {
	// Validate input
	if script.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if script.Content == "" {
		return nil, fmt.Errorf("content is required")
	}

	// Default group to "default" if not provided
	group := script.Group
	if group == "" {
		group = "default"
	}

	// Encrypt the content
	encryptedContent, err := database.Encrypt(script.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt content: %w", err)
	}

	now := time.Now().UTC()

	result, err := r.db.GetConnection().Exec(
		"INSERT INTO bash_scripts (name, description, content_encrypted, filename, group_name, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		script.Name,
		script.Description,
		encryptedContent,
		script.Filename,
		group,
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create bash script: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return &models.BashScript{
		ID:          id,
		Name:        script.Name,
		Description: script.Description,
		Content:     script.Content, // Return unencrypted content
		Filename:    script.Filename,
		Group:       group,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// GetByID retrieves a bash script by its ID
func (r *BashScriptRepository) GetByID(id int64) (*models.BashScript, error) {
	var script models.BashScript
	var encryptedContent []byte
	var description, filename sql.NullString

	err := r.db.GetConnection().QueryRow(
		"SELECT id, name, description, content_encrypted, filename, group_name, created_at, updated_at FROM bash_scripts WHERE id = ?",
		id,
	).Scan(&script.ID, &script.Name, &description, &encryptedContent, &filename, &script.Group, &script.CreatedAt, &script.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("bash script not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get bash script: %w", err)
	}

	// Handle nullable fields
	if description.Valid {
		script.Description = description.String
	}
	if filename.Valid {
		script.Filename = filename.String
	}

	// Decrypt the content
	decryptedContent, err := database.Decrypt(encryptedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt content: %w", err)
	}
	script.Content = decryptedContent

	return &script, nil
}

// GetAll retrieves all bash scripts (without content for listing)
func (r *BashScriptRepository) GetAll() ([]*models.BashScript, error) {
	rows, err := r.db.GetConnection().Query(
		"SELECT id, name, description, content_encrypted, filename, group_name, created_at, updated_at FROM bash_scripts ORDER BY group_name ASC, name ASC",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query bash scripts: %w", err)
	}
	defer rows.Close()

	var scripts []*models.BashScript
	for rows.Next() {
		var script models.BashScript
		var encryptedContent []byte
		var description, filename sql.NullString

		if err := rows.Scan(&script.ID, &script.Name, &description, &encryptedContent, &filename, &script.Group, &script.CreatedAt, &script.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan bash script: %w", err)
		}

		// Handle nullable fields
		if description.Valid {
			script.Description = description.String
		}
		if filename.Valid {
			script.Filename = filename.String
		}

		// Decrypt the content
		decryptedContent, err := database.Decrypt(encryptedContent)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt content: %w", err)
		}
		script.Content = decryptedContent

		scripts = append(scripts, &script)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bash scripts: %w", err)
	}

	return scripts, nil
}

// GetByGroup retrieves all bash scripts in a specific group
func (r *BashScriptRepository) GetByGroup(group string) ([]*models.BashScript, error) {
	rows, err := r.db.GetConnection().Query(
		"SELECT id, name, description, content_encrypted, filename, group_name, created_at, updated_at FROM bash_scripts WHERE group_name = ? ORDER BY name ASC",
		group,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query bash scripts: %w", err)
	}
	defer rows.Close()

	var scripts []*models.BashScript
	for rows.Next() {
		var script models.BashScript
		var encryptedContent []byte
		var description, filename sql.NullString

		if err := rows.Scan(&script.ID, &script.Name, &description, &encryptedContent, &filename, &script.Group, &script.CreatedAt, &script.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan bash script: %w", err)
		}

		// Handle nullable fields
		if description.Valid {
			script.Description = description.String
		}
		if filename.Valid {
			script.Filename = filename.String
		}

		// Decrypt the content
		decryptedContent, err := database.Decrypt(encryptedContent)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt content: %w", err)
		}
		script.Content = decryptedContent

		scripts = append(scripts, &script)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating bash scripts: %w", err)
	}

	return scripts, nil
}

// GetGroups retrieves all distinct group names
func (r *BashScriptRepository) GetGroups() ([]string, error) {
	rows, err := r.db.GetConnection().Query(
		"SELECT DISTINCT group_name FROM bash_scripts ORDER BY group_name ASC",
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

// Update updates an existing bash script
func (r *BashScriptRepository) Update(id int64, update *models.BashScriptUpdate) (*models.BashScript, error) {
	// Get existing script
	existing, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if update.Name != "" {
		existing.Name = update.Name
	}

	if update.Content != "" {
		existing.Content = update.Content
	}

	// Description and Filename can be updated
	if update.Description != "" {
		existing.Description = update.Description
	}

	if update.Filename != "" {
		existing.Filename = update.Filename
	}

	if update.Group != "" {
		existing.Group = update.Group
	}

	existing.UpdatedAt = time.Now().UTC()

	// Encrypt the content
	encryptedContent, err := database.Encrypt(existing.Content)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt content: %w", err)
	}

	_, err = r.db.GetConnection().Exec(
		"UPDATE bash_scripts SET name = ?, description = ?, content_encrypted = ?, filename = ?, group_name = ?, updated_at = ? WHERE id = ?",
		existing.Name,
		existing.Description,
		encryptedContent,
		existing.Filename,
		existing.Group,
		existing.UpdatedAt,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update bash script: %w", err)
	}

	return existing, nil
}

// Delete deletes a bash script by its ID
func (r *BashScriptRepository) Delete(id int64) error {
	result, err := r.db.GetConnection().Exec("DELETE FROM bash_scripts WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete bash script: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("bash script not found")
	}

	return nil
}

// GetByName retrieves a bash script by its name
func (r *BashScriptRepository) GetByName(name string) (*models.BashScript, error) {
	var script models.BashScript
	var encryptedContent []byte
	var description, filename sql.NullString

	err := r.db.GetConnection().QueryRow(
		"SELECT id, name, description, content_encrypted, filename, group_name, created_at, updated_at FROM bash_scripts WHERE name = ?",
		name,
	).Scan(&script.ID, &script.Name, &description, &encryptedContent, &filename, &script.Group, &script.CreatedAt, &script.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("bash script not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get bash script: %w", err)
	}

	// Handle nullable fields
	if description.Valid {
		script.Description = description.String
	}
	if filename.Valid {
		script.Filename = filename.String
	}

	// Decrypt the content
	decryptedContent, err := database.Decrypt(encryptedContent)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt content: %w", err)
	}
	script.Content = decryptedContent

	return &script, nil
}
