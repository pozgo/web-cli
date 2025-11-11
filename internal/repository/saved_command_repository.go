package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pozgo/web-cli/internal/database"
	"github.com/pozgo/web-cli/internal/models"
)

// SavedCommandRepository handles database operations for saved commands
type SavedCommandRepository struct {
	db *database.DB
}

// NewSavedCommandRepository creates a new saved command repository
func NewSavedCommandRepository(db *database.DB) *SavedCommandRepository {
	return &SavedCommandRepository{db: db}
}

// Create creates a new saved command
func (r *SavedCommandRepository) Create(cmd *models.SavedCommandCreate) (*models.SavedCommand, error) {
	// Validate input
	if cmd.Name == "" {
		return nil, fmt.Errorf("name is required")
	}
	if cmd.Command == "" {
		return nil, fmt.Errorf("command is required")
	}

	// Default user to "root" if not specified
	user := cmd.User
	if user == "" {
		user = "root"
	}

	now := time.Now().UTC()

	result, err := r.db.GetConnection().Exec(
		"INSERT INTO saved_commands (name, command, description, user, is_remote, server_id, ssh_key_id, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)",
		cmd.Name,
		cmd.Command,
		cmd.Description,
		user,
		cmd.IsRemote,
		cmd.ServerID,
		cmd.SSHKeyID,
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create saved command: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return &models.SavedCommand{
		ID:          id,
		Name:        cmd.Name,
		Command:     cmd.Command,
		Description: cmd.Description,
		User:        user,
		IsRemote:    cmd.IsRemote,
		ServerID:    cmd.ServerID,
		SSHKeyID:    cmd.SSHKeyID,
		CreatedAt:   now,
		UpdatedAt:   now,
	}, nil
}

// GetByID retrieves a saved command by its ID
func (r *SavedCommandRepository) GetByID(id int64) (*models.SavedCommand, error) {
	var cmd models.SavedCommand

	err := r.db.GetConnection().QueryRow(
		"SELECT id, name, command, description, user, is_remote, server_id, ssh_key_id, created_at, updated_at FROM saved_commands WHERE id = ?",
		id,
	).Scan(&cmd.ID, &cmd.Name, &cmd.Command, &cmd.Description, &cmd.User, &cmd.IsRemote, &cmd.ServerID, &cmd.SSHKeyID, &cmd.CreatedAt, &cmd.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("saved command not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get saved command: %w", err)
	}

	return &cmd, nil
}

// GetAll retrieves all saved commands
func (r *SavedCommandRepository) GetAll() ([]*models.SavedCommand, error) {
	rows, err := r.db.GetConnection().Query(
		"SELECT id, name, command, description, user, is_remote, server_id, ssh_key_id, created_at, updated_at FROM saved_commands ORDER BY name ASC",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query saved commands: %w", err)
	}
	defer rows.Close()

	var commands []*models.SavedCommand
	for rows.Next() {
		var cmd models.SavedCommand

		if err := rows.Scan(&cmd.ID, &cmd.Name, &cmd.Command, &cmd.Description, &cmd.User, &cmd.IsRemote, &cmd.ServerID, &cmd.SSHKeyID, &cmd.CreatedAt, &cmd.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan saved command: %w", err)
		}

		commands = append(commands, &cmd)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating saved commands: %w", err)
	}

	return commands, nil
}

// Update updates an existing saved command
func (r *SavedCommandRepository) Update(id int64, update *models.SavedCommandUpdate) (*models.SavedCommand, error) {
	// Get existing command
	existing, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if update.Name != "" {
		existing.Name = update.Name
	}

	if update.Command != "" {
		existing.Command = update.Command
	}

	if update.Description != "" {
		existing.Description = update.Description
	}

	if update.User != "" {
		existing.User = update.User
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

	existing.UpdatedAt = time.Now().UTC()

	_, err = r.db.GetConnection().Exec(
		"UPDATE saved_commands SET name = ?, command = ?, description = ?, user = ?, is_remote = ?, server_id = ?, ssh_key_id = ?, updated_at = ? WHERE id = ?",
		existing.Name,
		existing.Command,
		existing.Description,
		existing.User,
		existing.IsRemote,
		existing.ServerID,
		existing.SSHKeyID,
		existing.UpdatedAt,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update saved command: %w", err)
	}

	return existing, nil
}

// Delete deletes a saved command by its ID
func (r *SavedCommandRepository) Delete(id int64) error {
	result, err := r.db.GetConnection().Exec("DELETE FROM saved_commands WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete saved command: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("saved command not found")
	}

	return nil
}
