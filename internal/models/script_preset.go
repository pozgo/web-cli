package models

import "time"

// ScriptPreset represents a saved script execution configuration
// It stores which environment variables to use and optionally remote execution settings
type ScriptPreset struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`        // Display name for the preset
	Description string    `json:"description"` // Optional description
	ScriptID    int64     `json:"script_id"`   // Reference to bash_scripts table
	EnvVarIDs   []int64   `json:"env_var_ids"` // Selected environment variable IDs
	IsRemote    bool      `json:"is_remote"`   // Whether this is for remote execution
	ServerID    *int64    `json:"server_id"`   // Optional server for remote execution
	SSHKeyID    *int64    `json:"ssh_key_id"`  // Optional SSH key for remote execution
	User        string    `json:"user"`        // User to run as (for remote execution)
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ScriptPresetCreate represents the data needed to create a new script preset
type ScriptPresetCreate struct {
	Name        string  `json:"name" validate:"required"`
	Description string  `json:"description,omitempty"`
	ScriptID    int64   `json:"script_id" validate:"required"`
	EnvVarIDs   []int64 `json:"env_var_ids"`
	IsRemote    bool    `json:"is_remote"`
	ServerID    *int64  `json:"server_id,omitempty"`
	SSHKeyID    *int64  `json:"ssh_key_id,omitempty"`
	User        string  `json:"user,omitempty"`
}

// ScriptPresetUpdate represents the data that can be updated for a script preset
type ScriptPresetUpdate struct {
	Name        string  `json:"name,omitempty"`
	Description string  `json:"description,omitempty"`
	ScriptID    *int64  `json:"script_id,omitempty"`
	EnvVarIDs   []int64 `json:"env_var_ids,omitempty"`
	IsRemote    *bool   `json:"is_remote,omitempty"`
	ServerID    *int64  `json:"server_id,omitempty"`
	SSHKeyID    *int64  `json:"ssh_key_id,omitempty"`
	User        string  `json:"user,omitempty"`
}

// ScriptPresetResponse is the API response format
type ScriptPresetResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	ScriptID    int64     `json:"script_id"`
	EnvVarIDs   []int64   `json:"env_var_ids"`
	IsRemote    bool      `json:"is_remote"`
	ServerID    *int64    `json:"server_id"`
	SSHKeyID    *int64    `json:"ssh_key_id"`
	User        string    `json:"user"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ToResponse converts a ScriptPreset to a response
func (p *ScriptPreset) ToResponse() *ScriptPresetResponse {
	envVarIDs := p.EnvVarIDs
	if envVarIDs == nil {
		envVarIDs = []int64{}
	}
	return &ScriptPresetResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		ScriptID:    p.ScriptID,
		EnvVarIDs:   envVarIDs,
		IsRemote:    p.IsRemote,
		ServerID:    p.ServerID,
		SSHKeyID:    p.SSHKeyID,
		User:        p.User,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

// ScriptPresetsToList converts a slice of ScriptPresets to responses
func ScriptPresetsToList(presets []*ScriptPreset) []*ScriptPresetResponse {
	result := make([]*ScriptPresetResponse, len(presets))
	for i, p := range presets {
		result[i] = p.ToResponse()
	}
	return result
}
