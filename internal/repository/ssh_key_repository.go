package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pozgo/web-cli/internal/database"
	"github.com/pozgo/web-cli/internal/models"
)

// SSHKeyRepository handles database operations for SSH keys
type SSHKeyRepository struct {
	db *database.DB
}

// NewSSHKeyRepository creates a new SSH key repository
func NewSSHKeyRepository(db *database.DB) *SSHKeyRepository {
	return &SSHKeyRepository{db: db}
}

// Create creates a new SSH key in the database
func (r *SSHKeyRepository) Create(key *models.SSHKeyCreate) (*models.SSHKey, error) {
	// Encrypt the private key
	encryptedKey, err := database.Encrypt(key.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}

	now := time.Now().UTC()

	result, err := r.db.GetConnection().Exec(
		"INSERT INTO ssh_keys (name, private_key_encrypted, created_at, updated_at) VALUES (?, ?, ?, ?)",
		key.Name,
		encryptedKey,
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create SSH key: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return &models.SSHKey{
		ID:         id,
		Name:       key.Name,
		PrivateKey: key.PrivateKey,
		CreatedAt:  now,
		UpdatedAt:  now,
	}, nil
}

// GetByID retrieves an SSH key by its ID
func (r *SSHKeyRepository) GetByID(id int64) (*models.SSHKey, error) {
	var key models.SSHKey
	var encryptedKey []byte

	err := r.db.GetConnection().QueryRow(
		"SELECT id, name, private_key_encrypted, created_at, updated_at FROM ssh_keys WHERE id = ?",
		id,
	).Scan(&key.ID, &key.Name, &encryptedKey, &key.CreatedAt, &key.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("SSH key not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get SSH key: %w", err)
	}

	// Decrypt the private key
	decryptedKey, err := database.Decrypt(encryptedKey)
	if err != nil {
		return nil, fmt.Errorf("failed to decrypt private key: %w", err)
	}

	key.PrivateKey = decryptedKey
	return &key, nil
}

// GetAll retrieves all SSH keys
func (r *SSHKeyRepository) GetAll() ([]*models.SSHKey, error) {
	rows, err := r.db.GetConnection().Query(
		"SELECT id, name, private_key_encrypted, created_at, updated_at FROM ssh_keys ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query SSH keys: %w", err)
	}
	defer rows.Close()

	var keys []*models.SSHKey
	for rows.Next() {
		var key models.SSHKey
		var encryptedKey []byte

		if err := rows.Scan(&key.ID, &key.Name, &encryptedKey, &key.CreatedAt, &key.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan SSH key: %w", err)
		}

		// Decrypt the private key
		decryptedKey, err := database.Decrypt(encryptedKey)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt private key: %w", err)
		}

		key.PrivateKey = decryptedKey
		keys = append(keys, &key)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating SSH keys: %w", err)
	}

	return keys, nil
}

// Update updates an existing SSH key
func (r *SSHKeyRepository) Update(id int64, update *models.SSHKeyUpdate) (*models.SSHKey, error) {
	// Get existing key
	existing, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if update.Name != "" {
		existing.Name = update.Name
	}

	if update.PrivateKey != "" {
		existing.PrivateKey = update.PrivateKey
	}

	// Encrypt the private key
	encryptedKey, err := database.Encrypt(existing.PrivateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt private key: %w", err)
	}

	existing.UpdatedAt = time.Now().UTC()

	_, err = r.db.GetConnection().Exec(
		"UPDATE ssh_keys SET name = ?, private_key_encrypted = ?, updated_at = ? WHERE id = ?",
		existing.Name,
		encryptedKey,
		existing.UpdatedAt,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update SSH key: %w", err)
	}

	return existing, nil
}

// Delete deletes an SSH key by its ID
func (r *SSHKeyRepository) Delete(id int64) error {
	result, err := r.db.GetConnection().Exec("DELETE FROM ssh_keys WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete SSH key: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("SSH key not found")
	}

	return nil
}
