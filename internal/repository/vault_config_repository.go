package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pozgo/web-cli/internal/database"
	"github.com/pozgo/web-cli/internal/models"
)

// VaultConfigRepository handles Vault configuration database operations
type VaultConfigRepository struct {
	db *sql.DB
}

// NewVaultConfigRepository creates a new VaultConfigRepository
func NewVaultConfigRepository(db *database.DB) *VaultConfigRepository {
	return &VaultConfigRepository{db: db.GetConnection()}
}

// Get retrieves the Vault configuration (there should only be one)
func (r *VaultConfigRepository) Get() (*models.VaultConfig, error) {
	query := `
		SELECT id, address, token_encrypted, namespace, mount_path, enabled, created_at, updated_at
		FROM vault_config
		LIMIT 1
	`

	var cfg models.VaultConfig
	var tokenEncrypted []byte
	var namespace sql.NullString

	err := r.db.QueryRow(query).Scan(
		&cfg.ID,
		&cfg.Address,
		&tokenEncrypted,
		&namespace,
		&cfg.MountPath,
		&cfg.Enabled,
		&cfg.CreatedAt,
		&cfg.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, nil // No configuration exists
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get vault config: %w", err)
	}

	// Decrypt token
	if len(tokenEncrypted) > 0 {
		decrypted, err := database.Decrypt(tokenEncrypted)
		if err != nil {
			return nil, fmt.Errorf("failed to decrypt vault token: %w", err)
		}
		cfg.Token = decrypted
	}

	if namespace.Valid {
		cfg.Namespace = namespace.String
	}

	return &cfg, nil
}

// CreateOrUpdate creates or updates the Vault configuration
func (r *VaultConfigRepository) CreateOrUpdate(create *models.VaultConfigCreate) (*models.VaultConfig, error) {
	// Check if config exists
	existing, err := r.Get()
	if err != nil {
		return nil, err
	}

	// Encrypt token
	tokenEncrypted, err := database.Encrypt(create.Token)
	if err != nil {
		return nil, fmt.Errorf("failed to encrypt vault token: %w", err)
	}

	mountPath := create.MountPath
	if mountPath == "" {
		mountPath = "secret"
	}

	now := time.Now().UTC()

	if existing == nil {
		// Create new config
		query := `
			INSERT INTO vault_config (address, token_encrypted, namespace, mount_path, enabled, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?)
		`

		result, err := r.db.Exec(query,
			create.Address,
			tokenEncrypted,
			nullString(create.Namespace),
			mountPath,
			create.Enabled,
			now,
			now,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to create vault config: %w", err)
		}

		id, _ := result.LastInsertId()
		return &models.VaultConfig{
			ID:        id,
			Address:   create.Address,
			Token:     create.Token,
			Namespace: create.Namespace,
			MountPath: mountPath,
			Enabled:   create.Enabled,
			CreatedAt: now,
			UpdatedAt: now,
		}, nil
	}

	// Update existing config
	query := `
		UPDATE vault_config
		SET address = ?, token_encrypted = ?, namespace = ?, mount_path = ?, enabled = ?, updated_at = ?
		WHERE id = ?
	`

	_, err = r.db.Exec(query,
		create.Address,
		tokenEncrypted,
		nullString(create.Namespace),
		mountPath,
		create.Enabled,
		now,
		existing.ID,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update vault config: %w", err)
	}

	return &models.VaultConfig{
		ID:        existing.ID,
		Address:   create.Address,
		Token:     create.Token,
		Namespace: create.Namespace,
		MountPath: mountPath,
		Enabled:   create.Enabled,
		CreatedAt: existing.CreatedAt,
		UpdatedAt: now,
	}, nil
}

// UpdateToken updates only the Vault token
func (r *VaultConfigRepository) UpdateToken(token string) error {
	tokenEncrypted, err := database.Encrypt(token)
	if err != nil {
		return fmt.Errorf("failed to encrypt vault token: %w", err)
	}

	query := `UPDATE vault_config SET token_encrypted = ?, updated_at = ?`
	_, err = r.db.Exec(query, tokenEncrypted, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to update vault token: %w", err)
	}

	return nil
}

// SetEnabled enables or disables Vault integration
func (r *VaultConfigRepository) SetEnabled(enabled bool) error {
	query := `UPDATE vault_config SET enabled = ?, updated_at = ?`
	_, err := r.db.Exec(query, enabled, time.Now().UTC())
	if err != nil {
		return fmt.Errorf("failed to update vault enabled status: %w", err)
	}

	return nil
}

// Delete removes the Vault configuration
func (r *VaultConfigRepository) Delete() error {
	query := `DELETE FROM vault_config`
	_, err := r.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to delete vault config: %w", err)
	}

	return nil
}

// Note: nullString is defined in server_repository.go
