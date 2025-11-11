package models

import "time"

// LocalUser represents a local system user that can be used for command execution
// These users are stored for easy selection when executing local commands
type LocalUser struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"` // Unix username (must be valid system username)
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// LocalUserCreate represents the data needed to create a new local user entry
type LocalUserCreate struct {
	Name string `json:"name" validate:"required"` // Unix username
}

// LocalUserUpdate represents the data that can be updated for a local user entry
type LocalUserUpdate struct {
	Name string `json:"name,omitempty"` // Unix username
}
