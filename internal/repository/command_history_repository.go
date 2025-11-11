package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pozgo/web-cli/internal/database"
	"github.com/pozgo/web-cli/internal/models"
)

// CommandHistoryRepository handles database operations for command history
type CommandHistoryRepository struct {
	db *database.DB
}

// NewCommandHistoryRepository creates a new command history repository
func NewCommandHistoryRepository(db *database.DB) *CommandHistoryRepository {
	return &CommandHistoryRepository{db: db}
}

// Create creates a new command history record
func (r *CommandHistoryRepository) Create(history *models.CommandHistoryCreate) (*models.CommandHistory, error) {
	// Encrypt command
	encryptedCommand, err := database.Encrypt(history.Command)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt command: %w", err)
	}

	// Encrypt output if present
	var encryptedOutput []byte
	if history.Output != "" {
		encryptedOutput, err = database.Encrypt(history.Output)
		if err != nil {
			return nil, fmt.Errorf("failed to encrypt output: %w", err)
		}
	}

	now := time.Now().UTC()

	result, err := r.db.GetConnection().Exec(
		"INSERT INTO command_history (command_encrypted, output_encrypted, exit_code, server, user, execution_time_ms, executed_at) VALUES (?, ?, ?, ?, ?, ?, ?)",
		encryptedCommand,
		encryptedOutput,
		history.ExitCode,
		history.Server,
		history.User,
		history.ExecutionTimeMs,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create command history: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return &models.CommandHistory{
		ID:              id,
		Command:         history.Command,
		Output:          history.Output,
		ExitCode:        history.ExitCode,
		Server:          history.Server,
		User:            history.User,
		ExecutionTimeMs: history.ExecutionTimeMs,
		ExecutedAt:      now,
	}, nil
}

// GetByID retrieves a command history record by its ID
func (r *CommandHistoryRepository) GetByID(id int64) (*models.CommandHistory, error) {
	var history models.CommandHistory
	var encryptedCommand []byte
	var encryptedOutput []byte

	var user sql.NullString

	err := r.db.GetConnection().QueryRow(
		"SELECT id, command_encrypted, output_encrypted, exit_code, server, user, execution_time_ms, executed_at FROM command_history WHERE id = ?",
		id,
	).Scan(&history.ID, &encryptedCommand, &encryptedOutput, &history.ExitCode, &history.Server, &user, &history.ExecutionTimeMs, &history.ExecutedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("command history not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get command history: %w", err)
	}

	// Decrypt command
	decryptedCommand, err := database.Decrypt(encryptedCommand)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt command: %w", err)
	}
	history.Command = decryptedCommand

	// Decrypt output if present
	if len(encryptedOutput) > 0 {
		decryptedOutput, err := database.Decrypt(encryptedOutput)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt output: %w", err)
		}
		history.Output = decryptedOutput
	}

	// Set user if present
	if user.Valid {
		history.User = user.String
	}

	return &history, nil
}

// GetAll retrieves all command history records with optional limit
func (r *CommandHistoryRepository) GetAll(limit int) ([]*models.CommandHistory, error) {
	query := "SELECT id, command_encrypted, output_encrypted, exit_code, server, user, execution_time_ms, executed_at FROM command_history ORDER BY executed_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.GetConnection().Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to query command history: %w", err)
	}
	defer rows.Close()

	var histories []*models.CommandHistory
	for rows.Next() {
		var history models.CommandHistory
		var encryptedCommand []byte
		var encryptedOutput []byte
		var user sql.NullString

		if err := rows.Scan(&history.ID, &encryptedCommand, &encryptedOutput, &history.ExitCode, &history.Server, &user, &history.ExecutionTimeMs, &history.ExecutedAt); err != nil {
			return nil, fmt.Errorf("failed to scan command history: %w", err)
		}

		// Decrypt command
		decryptedCommand, err := database.Decrypt(encryptedCommand)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt command: %w", err)
		}
		history.Command = decryptedCommand

		// Decrypt output if present
		if len(encryptedOutput) > 0 {
			decryptedOutput, err := database.Decrypt(encryptedOutput)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt output: %w", err)
			}
			history.Output = decryptedOutput
		}

		// Set user if present
		if user.Valid {
			history.User = user.String
		}

		histories = append(histories, &history)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating command history: %w", err)
	}

	return histories, nil
}

// GetByServer retrieves command history for a specific server
func (r *CommandHistoryRepository) GetByServer(server string, limit int) ([]*models.CommandHistory, error) {
	query := "SELECT id, command_encrypted, output_encrypted, exit_code, server, user, execution_time_ms, executed_at FROM command_history WHERE server = ? ORDER BY executed_at DESC"

	if limit > 0 {
		query += fmt.Sprintf(" LIMIT %d", limit)
	}

	rows, err := r.db.GetConnection().Query(query, server)
	if err != nil {
		return nil, fmt.Errorf("failed to query command history: %w", err)
	}
	defer rows.Close()

	var histories []*models.CommandHistory
	for rows.Next() {
		var history models.CommandHistory
		var encryptedCommand []byte
		var encryptedOutput []byte
		var user sql.NullString

		if err := rows.Scan(&history.ID, &encryptedCommand, &encryptedOutput, &history.ExitCode, &history.Server, &user, &history.ExecutionTimeMs, &history.ExecutedAt); err != nil {
			return nil, fmt.Errorf("failed to scan command history: %w", err)
		}

		// Decrypt command
		decryptedCommand, err := database.Decrypt(encryptedCommand)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt command: %w", err)
		}
		history.Command = decryptedCommand

		// Decrypt output if present
		if len(encryptedOutput) > 0 {
			decryptedOutput, err := database.Decrypt(encryptedOutput)
			if err != nil {
				return nil, fmt.Errorf("failed to decrypt output: %w", err)
			}
			history.Output = decryptedOutput
		}

		// Set user if present
		if user.Valid {
			history.User = user.String
		}

		histories = append(histories, &history)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating command history: %w", err)
	}

	return histories, nil
}

// Delete deletes a command history record by its ID
func (r *CommandHistoryRepository) Delete(id int64) error {
	result, err := r.db.GetConnection().Exec("DELETE FROM command_history WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete command history: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("command history not found")
	}

	return nil
}

// DeleteOlderThan deletes command history records older than the specified time
func (r *CommandHistoryRepository) DeleteOlderThan(before time.Time) (int64, error) {
	result, err := r.db.GetConnection().Exec("DELETE FROM command_history WHERE executed_at < ?", before)
	if err != nil {
		return 0, fmt.Errorf("failed to delete old command history: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return 0, fmt.Errorf("failed to get rows affected: %w", err)
	}

	return rowsAffected, nil
}
