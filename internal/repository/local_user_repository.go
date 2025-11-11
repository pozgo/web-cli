package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pozgo/web-cli/internal/database"
	"github.com/pozgo/web-cli/internal/models"
)

// LocalUserRepository handles database operations for local users
type LocalUserRepository struct {
	db *database.DB
}

// NewLocalUserRepository creates a new local user repository
func NewLocalUserRepository(db *database.DB) *LocalUserRepository {
	return &LocalUserRepository{db: db}
}

// Create creates a new local user in the database
func (r *LocalUserRepository) Create(user *models.LocalUserCreate) (*models.LocalUser, error) {
	// Validate that name is provided
	if user.Name == "" {
		return nil, fmt.Errorf("name is required")
	}

	now := time.Now().UTC()

	result, err := r.db.GetConnection().Exec(
		"INSERT INTO local_users (name, created_at, updated_at) VALUES (?, ?, ?)",
		user.Name,
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create local user: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return &models.LocalUser{
		ID:        id,
		Name:      user.Name,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// GetByID retrieves a local user by its ID
func (r *LocalUserRepository) GetByID(id int64) (*models.LocalUser, error) {
	var user models.LocalUser

	err := r.db.GetConnection().QueryRow(
		"SELECT id, name, created_at, updated_at FROM local_users WHERE id = ?",
		id,
	).Scan(&user.ID, &user.Name, &user.CreatedAt, &user.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("local user not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get local user: %w", err)
	}

	return &user, nil
}

// GetAll retrieves all local users
func (r *LocalUserRepository) GetAll() ([]*models.LocalUser, error) {
	rows, err := r.db.GetConnection().Query(
		"SELECT id, name, created_at, updated_at FROM local_users ORDER BY name ASC",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query local users: %w", err)
	}
	defer rows.Close()

	var users []*models.LocalUser
	for rows.Next() {
		var user models.LocalUser

		if err := rows.Scan(&user.ID, &user.Name, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan local user: %w", err)
		}

		users = append(users, &user)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating local users: %w", err)
	}

	return users, nil
}

// Update updates an existing local user
func (r *LocalUserRepository) Update(id int64, update *models.LocalUserUpdate) (*models.LocalUser, error) {
	// Get existing user
	existing, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if update.Name != "" {
		existing.Name = update.Name
	}

	// Validate that name is not empty after update
	if existing.Name == "" {
		return nil, fmt.Errorf("name cannot be empty")
	}

	existing.UpdatedAt = time.Now().UTC()

	_, err = r.db.GetConnection().Exec(
		"UPDATE local_users SET name = ?, updated_at = ? WHERE id = ?",
		existing.Name,
		existing.UpdatedAt,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update local user: %w", err)
	}

	return existing, nil
}

// Delete deletes a local user by its ID
func (r *LocalUserRepository) Delete(id int64) error {
	result, err := r.db.GetConnection().Exec("DELETE FROM local_users WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete local user: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("local user not found")
	}

	return nil
}
