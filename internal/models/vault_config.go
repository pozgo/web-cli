package models

import "time"

// VaultConfig represents the HashiCorp Vault configuration
type VaultConfig struct {
	ID        int64     `json:"id"`
	Address   string    `json:"address"`             // Vault server address (e.g., https://vault.example.com:8200)
	Token     string    `json:"token,omitempty"`     // Vault token (decrypted, not included in responses)
	Namespace string    `json:"namespace,omitempty"` // Optional Vault namespace
	MountPath string    `json:"mount_path"`          // KV secrets engine mount path (default: "secret")
	Enabled   bool      `json:"enabled"`             // Whether Vault integration is enabled
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// VaultConfigCreate represents the data needed to create/update Vault configuration
type VaultConfigCreate struct {
	Address   string `json:"address"`
	Token     string `json:"token"`
	Namespace string `json:"namespace,omitempty"`
	MountPath string `json:"mount_path,omitempty"`
	Enabled   bool   `json:"enabled"`
}

// VaultConfigResponse is the API response format (token masked)
type VaultConfigResponse struct {
	ID        int64     `json:"id"`
	Address   string    `json:"address"`
	Namespace string    `json:"namespace,omitempty"`
	MountPath string    `json:"mount_path"`
	Enabled   bool      `json:"enabled"`
	HasToken  bool      `json:"has_token"` // Indicates if a token is configured
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// ToResponse converts VaultConfig to a safe response (without token)
func (v *VaultConfig) ToResponse() VaultConfigResponse {
	return VaultConfigResponse{
		ID:        v.ID,
		Address:   v.Address,
		Namespace: v.Namespace,
		MountPath: v.MountPath,
		Enabled:   v.Enabled,
		HasToken:  v.Token != "",
		CreatedAt: v.CreatedAt,
		UpdatedAt: v.UpdatedAt,
	}
}

// VaultStatus represents the current Vault connection status
type VaultStatus struct {
	Configured  bool   `json:"configured"` // Whether Vault is configured
	Enabled     bool   `json:"enabled"`    // Whether Vault is enabled
	Connected   bool   `json:"connected"`  // Whether connection test passed
	Address     string `json:"address,omitempty"`
	Error       string `json:"error,omitempty"`
	VaultSealed bool   `json:"vault_sealed,omitempty"`
}
