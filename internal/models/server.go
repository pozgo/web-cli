package models

import "time"

// Server represents a remote server configuration stored in the system
// Either Name or IPAddress must be provided (or both can be provided)
type Server struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name,omitempty"`       // Hostname (must follow hostname conventions)
	IPAddress string    `json:"ip_address,omitempty"` // IP address
	Port      int       `json:"port"`                 // SSH port (default: 22)
	Username  string    `json:"username"`             // SSH username for remote connections
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ServerCreate represents the data needed to create a new server
// At least one of Name or IPAddress must be provided
type ServerCreate struct {
	Name      string `json:"name,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
	Port      int    `json:"port"`     // Optional, defaults to 22 if not provided
	Username  string `json:"username"` // SSH username for remote connections
}

// ServerUpdate represents the data that can be updated for a server
type ServerUpdate struct {
	Name      string `json:"name,omitempty"`
	IPAddress string `json:"ip_address,omitempty"`
	Port      int    `json:"port,omitempty"`
	Username  string `json:"username,omitempty"`
}
