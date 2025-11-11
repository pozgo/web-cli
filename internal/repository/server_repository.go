package repository

import (
	"database/sql"
	"fmt"
	"time"

	"github.com/pozgo/web-cli/internal/database"
	"github.com/pozgo/web-cli/internal/models"
)

// ServerRepository handles database operations for servers
type ServerRepository struct {
	db *database.DB
}

// NewServerRepository creates a new server repository
func NewServerRepository(db *database.DB) *ServerRepository {
	return &ServerRepository{db: db}
}

// Create creates a new server in the database
func (r *ServerRepository) Create(server *models.ServerCreate) (*models.Server, error) {
	// Validate that at least one field is provided
	if server.Name == "" && server.IPAddress == "" {
		return nil, fmt.Errorf("at least one of name or ip_address must be provided")
	}

	// Default port to 22 if not provided or invalid
	port := server.Port
	if port <= 0 {
		port = 22
	}

	// Default username to root if not provided
	username := server.Username
	if username == "" {
		username = "root"
	}

	now := time.Now().UTC()

	result, err := r.db.GetConnection().Exec(
		"INSERT INTO servers (name, ip_address, port, username, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?)",
		nullString(server.Name),
		nullString(server.IPAddress),
		port,
		username,
		now,
		now,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create server: %w", err)
	}

	id, err := result.LastInsertId()
	if err != nil {
		return nil, fmt.Errorf("failed to get last insert ID: %w", err)
	}

	return &models.Server{
		ID:        id,
		Name:      server.Name,
		IPAddress: server.IPAddress,
		Port:      port,
		Username:  username,
		CreatedAt: now,
		UpdatedAt: now,
	}, nil
}

// GetByID retrieves a server by its ID
func (r *ServerRepository) GetByID(id int64) (*models.Server, error) {
	var server models.Server
	var name, ipAddress sql.NullString

	err := r.db.GetConnection().QueryRow(
		"SELECT id, name, ip_address, port, username, created_at, updated_at FROM servers WHERE id = ?",
		id,
	).Scan(&server.ID, &name, &ipAddress, &server.Port, &server.Username, &server.CreatedAt, &server.UpdatedAt)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("server not found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get server: %w", err)
	}

	server.Name = name.String
	server.IPAddress = ipAddress.String

	return &server, nil
}

// GetAll retrieves all servers
func (r *ServerRepository) GetAll() ([]*models.Server, error) {
	rows, err := r.db.GetConnection().Query(
		"SELECT id, name, ip_address, port, username, created_at, updated_at FROM servers ORDER BY created_at DESC",
	)
	if err != nil {
		return nil, fmt.Errorf("failed to query servers: %w", err)
	}
	defer rows.Close()

	var servers []*models.Server
	for rows.Next() {
		var server models.Server
		var name, ipAddress sql.NullString

		if err := rows.Scan(&server.ID, &name, &ipAddress, &server.Port, &server.Username, &server.CreatedAt, &server.UpdatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan server: %w", err)
		}

		server.Name = name.String
		server.IPAddress = ipAddress.String
		servers = append(servers, &server)
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating servers: %w", err)
	}

	return servers, nil
}

// Update updates an existing server
func (r *ServerRepository) Update(id int64, update *models.ServerUpdate) (*models.Server, error) {
	// Get existing server
	existing, err := r.GetByID(id)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if update.Name != "" {
		existing.Name = update.Name
	}

	if update.IPAddress != "" {
		existing.IPAddress = update.IPAddress
	}

	if update.Port > 0 {
		existing.Port = update.Port
	}

	if update.Username != "" {
		existing.Username = update.Username
	}

	// Validate that at least one field is set after update
	if existing.Name == "" && existing.IPAddress == "" {
		return nil, fmt.Errorf("at least one of name or ip_address must be provided")
	}

	// Ensure port is valid (default to 22 if somehow invalid)
	if existing.Port <= 0 {
		existing.Port = 22
	}

	// Ensure username is not empty (default to root if somehow empty)
	if existing.Username == "" {
		existing.Username = "root"
	}

	existing.UpdatedAt = time.Now().UTC()

	_, err = r.db.GetConnection().Exec(
		"UPDATE servers SET name = ?, ip_address = ?, port = ?, username = ?, updated_at = ? WHERE id = ?",
		nullString(existing.Name),
		nullString(existing.IPAddress),
		existing.Port,
		existing.Username,
		existing.UpdatedAt,
		id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update server: %w", err)
	}

	return existing, nil
}

// Delete deletes a server by its ID
func (r *ServerRepository) Delete(id int64) error {
	result, err := r.db.GetConnection().Exec("DELETE FROM servers WHERE id = ?", id)
	if err != nil {
		return fmt.Errorf("failed to delete server: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("server not found")
	}

	return nil
}

// nullString converts an empty string to sql.NullString
func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{Valid: false}
	}
	return sql.NullString{String: s, Valid: true}
}
