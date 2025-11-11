package models

import "time"

// SSHKey represents an SSH private key stored in the system
type SSHKey struct {
	ID         int64     `json:"id"`
	Name       string    `json:"name"`
	PrivateKey string    `json:"private_key"` // Decrypted value
	CreatedAt  time.Time `json:"created_at"`
	UpdatedAt  time.Time `json:"updated_at"`
}

// SSHKeyCreate represents the data needed to create a new SSH key
type SSHKeyCreate struct {
	Name       string `json:"name" validate:"required"`
	PrivateKey string `json:"private_key" validate:"required"`
}

// SSHKeyUpdate represents the data that can be updated for an SSH key
type SSHKeyUpdate struct {
	Name       string `json:"name,omitempty"`
	PrivateKey string `json:"private_key,omitempty"`
}
